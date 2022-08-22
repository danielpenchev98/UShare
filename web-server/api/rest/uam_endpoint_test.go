package rest_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"

	"github.com/danielpenchev98/UShare/web-server/api/common"
	"github.com/danielpenchev98/UShare/web-server/api/rest"
	"github.com/danielpenchev98/UShare/web-server/internal/auth/auth_mocks"
	"github.com/danielpenchev98/UShare/web-server/internal/db/dao/dao_mocks"
	"github.com/danielpenchev98/UShare/web-server/internal/db/models"
	myerr "github.com/danielpenchev98/UShare/web-server/internal/error"
	"github.com/danielpenchev98/UShare/web-server/internal/validator/validator_mocks"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/crypto/bcrypt"
)

func setupRouter(uamRest rest.UamEndpoint, userID uint) *gin.Engine {
	r := gin.Default()

	public := r.Group("/public")
	{
		public.POST("/user/registration", uamRest.CreateUser)
		public.POST("/user/login", uamRest.Login)
	}
	protected := r.Group("/protected").Use(func(c *gin.Context) {
		c.Set("userID", userID)
		c.Next()
	})
	{
		protected.DELETE("/user/deletion", uamRest.DeleteUser)
		protected.DELETE("/group/deletion", uamRest.DeleteGroup)
		protected.POST("/group/creation", uamRest.CreateGroup)
		protected.POST("/group/membership/revocation", uamRest.RevokeMembership)
		protected.POST("/group/membership/invitation", uamRest.AddMember)
	}
	return r
}

func assertErrorResponse(recorder *httptest.ResponseRecorder, expStatusCode int, expMessage string) {
	Expect(recorder.Code).To(Equal(expStatusCode))
	body := common.ErrorResponse{}
	json.Unmarshal([]byte(recorder.Body.String()), &body)
	Expect(body.ErrorCode).To(Equal(expStatusCode))
	Expect(body.ErrorMsg).To(ContainSubstring(expMessage))
}

var _ = Describe("UamEndpoint", func() {
	var (
		router     *gin.Engine
		recorder   *httptest.ResponseRecorder
		jwtCreator *auth_mocks.MockJwtCreator
		uamDAO     *dao_mocks.MockUamDAO
		validator  *validator_mocks.MockValidator
		req        *http.Request
	)

	const (
		username  = "username"
		password  = "password"
		userID    = 1
		groupName = "groupName"
		groupsDir = "."
	)

	BeforeEach(func() {
		controller := gomock.NewController(GinkgoT())
		uamDAO = dao_mocks.NewMockUamDAO(controller)
		jwtCreator = auth_mocks.NewMockJwtCreator(controller)
		validator = validator_mocks.NewMockValidator(controller)
		uamRest := rest.NewUamEndPointImpl(uamDAO, jwtCreator, validator, groupsDir)

		router = setupRouter(uamRest, userID)
		recorder = httptest.NewRecorder()
	})

	Context("CreateUser", func() {
		When("creation request is sent", func() {
			var reqBody *common.RequestWithCredentials

			BeforeEach(func() {
				reqBody = &common.RequestWithCredentials{
					Username: username,
					Password: password,
				}
			})

			Context("with non-json body", func() {

				BeforeEach(func() {
					validator.EXPECT().
						ValidateUsername(username).
						Times(0)

					validator.EXPECT().
						ValidatePassword(password).
						Times(0)

					uamDAO.EXPECT().
						CreateUser(gomock.Any(), gomock.Any()).
						Times(0)

					req, _ = http.NewRequest("POST", "/public/user/registration", strings.NewReader("test"))
				})

				It("returns bad request", func() {
					router.ServeHTTP(recorder, req)
					assertErrorResponse(recorder, http.StatusBadRequest, "Invalid json body")
				})
			})

			Context("with json body", func() {

				BeforeEach(func() {
					jsonBody, _ := json.Marshal(*reqBody)
					req, _ = http.NewRequest("POST", "/public/user/registration", bytes.NewBuffer(jsonBody))
					req.Header.Set("Content-Type", "application/json")
				})

				Context("with invalid format of username", func() {
					BeforeEach(func() {
						validator.EXPECT().
							ValidateUsername(gomock.Any()).
							Return(myerr.NewClientError("test-error"))

						uamDAO.EXPECT().
							CreateUser(gomock.Any(), gomock.Any()).
							Times(0)
					})

					It("returns bad request", func() {
						router.ServeHTTP(recorder, req)
						assertErrorResponse(recorder, http.StatusBadRequest, "test-error")
					})
				})

				Context("with valid formatted username", func() {
					Context("and invalid formatted password", func() {
						BeforeEach(func() {
							gomock.InOrder(
								validator.EXPECT().
									ValidateUsername(username).
									Return(nil),

								validator.EXPECT().
									ValidatePassword(password).
									Return(myerr.NewClientError("test-error")),
							)

							uamDAO.EXPECT().
								CreateUser(gomock.Any(), gomock.Any()).
								Times(0)
						})

						It("returns bad request", func() {
							router.ServeHTTP(recorder, req)
							assertErrorResponse(recorder, http.StatusBadRequest, "test-error")
						})
					})

					Context("and valid formatted password", func() {
						Context("and CreateUser request fails", func() {
							BeforeEach(func() {
								gomock.InOrder(
									validator.EXPECT().
										ValidateUsername(username).
										Return(nil),

									validator.EXPECT().
										ValidatePassword(password).
										Return(nil),

									uamDAO.EXPECT().
										CreateUser(username, gomock.Any()). // Any() because bscrypt cannot be mocked easily -> Could be wrapped
										Return(myerr.NewServerError("test-error")),
								)
							})

							It("returns bad request", func() {
								router.ServeHTTP(recorder, req)
								assertErrorResponse(recorder, http.StatusInternalServerError, "Problem with the server, please try again later")
							})
						})

						Context("and CreateUser request succeeds", func() {
							Context("and user already exists", func() {
								BeforeEach(func() {
									gomock.InOrder(
										validator.EXPECT().
											ValidateUsername(username).
											Return(nil),

										validator.EXPECT().
											ValidatePassword(password).
											Return(nil),

										uamDAO.EXPECT().
											CreateUser(username, gomock.Any()). // Any() because bscrypt cannot be mocked easily -> Could be wrapped
											Return(myerr.NewClientError("test-error")),
									)
								})

								It("returns bad request", func() {
									router.ServeHTTP(recorder, req)
									assertErrorResponse(recorder, http.StatusBadRequest, "test-error")
								})
							})

							Context("and user doesnt already exist", func() {
								BeforeEach(func() {
									gomock.InOrder(
										validator.EXPECT().
											ValidateUsername(username).
											Return(nil),

										validator.EXPECT().
											ValidatePassword(password).
											Return(nil),

										uamDAO.EXPECT().
											CreateUser(username, gomock.Any()). // Any() because bscrypt cannot be mocked easily -> Could be wrapped
											Return(nil),
									)
								})

								It("returns successfull response", func() {
									router.ServeHTTP(recorder, req)

									Expect(recorder.Code).To(Equal(http.StatusCreated))
									body := common.BasicResponse{}
									json.Unmarshal([]byte(recorder.Body.String()), &body)
									Expect(body.Status).To(Equal(http.StatusCreated))
								})
							})
						})
					})
				})
			})
		})
	})

	Context("Login", func() {
		When("login request is sent", func() {
			var reqBody *common.RequestWithCredentials

			const (
				username = "username"
				password = "password"
			)

			BeforeEach(func() {
				reqBody = &common.RequestWithCredentials{
					Username: username,
					Password: password,
				}
			})

			Context("with non-json body", func() {
				BeforeEach(func() {
					uamDAO.EXPECT().
						GetUser(gomock.Any()).
						Times(0)

					jwtCreator.EXPECT().
						GenerateToken(gomock.Any()).
						Times(0)

					req, _ = http.NewRequest("POST", "/public/user/login", strings.NewReader("test"))
				})

				It("returns bad request", func() {
					router.ServeHTTP(recorder, req)
					assertErrorResponse(recorder, http.StatusBadRequest, "Invalid json body")
				})
			})

			Context("with json body", func() {
				BeforeEach(func() {
					jsonBody, _ := json.Marshal(*reqBody)
					req, _ = http.NewRequest("POST", "/public/user/login", bytes.NewBuffer(jsonBody))
					req.Header.Set("Content-Type", "application/json")
				})

				Context("and request to check if user exist fail", func() {
					BeforeEach(func() {
						uamDAO.EXPECT().
							GetUser(username).
							Return(models.User{}, myerr.NewServerError("test-error"))

						jwtCreator.EXPECT().
							GenerateToken(gomock.Any()).
							Times(0)
					})

					It("returns internal server error response", func() {
						router.ServeHTTP(recorder, req)
						assertErrorResponse(recorder, http.StatusInternalServerError, "Problem with the server, please try again later")
					})
				})

				Context("and request to check if user exist succeeds", func() {
					Context("and user doesnt exist", func() {
						Context("username doesn't exist", func() {
							BeforeEach(func() {
								uamDAO.EXPECT().
									GetUser(username).
									Return(models.User{}, myerr.NewItemNotFoundError("test-error"))

								jwtCreator.EXPECT().
									GenerateToken(gomock.Any()).
									Times(0)
							})

							It("returns bad request error response", func() {
								router.ServeHTTP(recorder, req)
								assertErrorResponse(recorder, http.StatusBadRequest, "Invalid credentials")
							})
						})
						Context("passwords doesnt match", func() {
							BeforeEach(func() {
								encryptedPass, _ := bcrypt.GenerateFromPassword([]byte("different-password"), bcrypt.DefaultCost)

								uamDAO.EXPECT().
									GetUser(username).
									Return(models.User{Username: username, Password: string(encryptedPass)}, nil)

								jwtCreator.EXPECT().
									GenerateToken(gomock.Any()).
									Times(0)
							})

							It("returns bad request error response", func() {
								router.ServeHTTP(recorder, req)
								assertErrorResponse(recorder, http.StatusBadRequest, "Invalid credentials")
							})
						})

					})

					Context("and user exist", func() {
						var user models.User

						BeforeEach(func() {
							encryptedPass, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

							user = models.User{
								Username: username,
								Password: string(encryptedPass),
							}
							user.ID = 1
						})

						Context("and token generation fails", func() {
							BeforeEach(func() {

								gomock.InOrder(
									uamDAO.EXPECT().
										GetUser(user.Username).
										Return(user, nil),

									jwtCreator.EXPECT().
										GenerateToken(user.ID).
										Return("", errors.New("test-error")),
								)
							})

							It("returns internal server error response", func() {
								router.ServeHTTP(recorder, req)
								assertErrorResponse(recorder, http.StatusInternalServerError, "Problem with the server, please try again later")
							})
						})

						Context("and token generation succeeds", func() {
							const token = "token"

							BeforeEach(func() {
								gomock.InOrder(
									uamDAO.EXPECT().
										GetUser(user.Username).
										Return(user, nil),

									jwtCreator.EXPECT().
										GenerateToken(user.ID).
										Return(token, nil),
								)
							})

							It("returns successful response", func() {
								router.ServeHTTP(recorder, req)

								Expect(recorder.Code).To(Equal(http.StatusCreated))
								body := common.LoginResponse{}
								json.Unmarshal([]byte(recorder.Body.String()), &body)
								Expect(body.Status).To(Equal(http.StatusCreated))
								Expect(body.Token).To(Equal(token))
							})
						})
					})
				})
			})
		})
	})

	Context("DeleteUser", func() {
		When("login request is sent and authentication passes", func() {

			BeforeEach(func() {
				req, _ = http.NewRequest("DELETE", "/protected/user/deletion", nil)
				req.Header.Set("Authorization", "Bearer sometoken")
			})

			Context("operation of deleting user from db fails", func() {
				BeforeEach(func() {
					uamDAO.EXPECT().
						DeleteUser(uint(userID)).
						Return(myerr.NewServerError("test-error"))
				})

				It("returns internal server error response", func() {
					router.ServeHTTP(recorder, req)
					assertErrorResponse(recorder, http.StatusInternalServerError, "Problem with the server, please try again later")
				})
			})

			Context("operation of deleting user from db succeeds", func() {
				Context("and user doesnt exist", func() {
					BeforeEach(func() {
						uamDAO.EXPECT().
							DeleteUser(uint(userID)).
							Return(myerr.NewItemNotFoundError("test-error"))
					})

					It("returns internal server error response", func() {
						router.ServeHTTP(recorder, req)
						assertErrorResponse(recorder, http.StatusInternalServerError, "Problem with the server, please try again later")
					})
				})

				Context("and user exists", func() {
					BeforeEach(func() {
						uamDAO.EXPECT().
							DeleteUser(uint(userID)).
							Return(nil)
					})

					It("returns success response", func() {
						router.ServeHTTP(recorder, req)

						Expect(recorder.Code).To(Equal(http.StatusOK))
						body := common.BasicResponse{}
						json.Unmarshal([]byte(recorder.Body.String()), &body)
						Expect(body.Status).To(Equal(http.StatusOK))
					})
				})

			})
		})
	})

	Context("CreateGroup", func() {
		When("login request is sent and authentication passes", func() {
			Context("with non-json body", func() {

				BeforeEach(func() {
					validator.EXPECT().
						ValidateUsername(username).
						Times(0)

					uamDAO.EXPECT().
						CreateGroup(gomock.Any(), gomock.Any()).
						Times(0)

					req, _ = http.NewRequest("POST", "/protected/group/creation", strings.NewReader("test"))
				})

				It("returns bad request", func() {
					router.ServeHTTP(recorder, req)
					assertErrorResponse(recorder, http.StatusBadRequest, "Invalid json body")
				})
			})

			Context("with json body", func() {
				var rqBody common.GroupPayload

				BeforeEach(func() {
					rqBody = common.GroupPayload{GroupName: groupName}
					jsonBody, _ := json.Marshal(&rqBody)
					req, _ = http.NewRequest("POST", "/protected/group/creation", bytes.NewBuffer(jsonBody))
					req.Header.Set("Authorization", "Bearer sometoken")
				})

				Context("and group name fails the validation", func() {
					BeforeEach(func() {
						validator.EXPECT().
							ValidateUsername(rqBody.GroupName).
							Return(myerr.NewClientError("test-error"))

						uamDAO.EXPECT().
							CreateGroup(gomock.Any(), gomock.Any()).
							Times(0)
					})

					It("return bad request", func() {
						router.ServeHTTP(recorder, req)
						assertErrorResponse(recorder, http.StatusBadRequest, "test-error")
					})

				})

				Context("and group name passes the validation", func() {
					AfterEach(func() {
						os.RemoveAll(path.Join(groupsDir, groupName))
					})

					Context("and operation of creation group from db fails", func() {
						Context("because connection to db fails", func() {
							BeforeEach(func() {
								gomock.InOrder(
									validator.EXPECT().
										ValidateUsername(rqBody.GroupName).
										Return(nil),
									uamDAO.EXPECT().
										CreateGroup(uint(userID), rqBody.GroupName).
										Return(myerr.NewServerError("test-error")),
								)
							})

							It("returns internal server error response", func() {
								router.ServeHTTP(recorder, req)
								assertErrorResponse(recorder, http.StatusInternalServerError, "Problem with the server, please try again later")
							})
						})

						Context("because group already exists", func() {
							BeforeEach(func() {
								gomock.InOrder(
									validator.EXPECT().
										ValidateUsername(rqBody.GroupName).
										Return(nil),
									uamDAO.EXPECT().
										CreateGroup(uint(userID), rqBody.GroupName).
										Return(myerr.NewClientError("test-error")),
								)
							})

							It("returns bad request response", func() {
								router.ServeHTTP(recorder, req)
								assertErrorResponse(recorder, http.StatusBadRequest, "test-error")
							})
						})

					})

					Context("and operation of creating group from db suceeeds", func() {
						BeforeEach(func() {
							gomock.InOrder(
								validator.EXPECT().
									ValidateUsername(rqBody.GroupName).
									Return(nil),
								uamDAO.EXPECT().
									CreateGroup(uint(userID), rqBody.GroupName).
									Return(nil),
							)
						})

						It("returns Created", func() {
							router.ServeHTTP(recorder, req)

							Expect(recorder.Code).To(Equal(http.StatusCreated))
							body := common.BasicResponse{}
							json.Unmarshal([]byte(recorder.Body.String()), &body)
							Expect(body.Status).To(Equal(http.StatusCreated))
						})
					})
				})
			})
		})
	})

	Context("AddMember", func() {
		When("request a user to be added to group is sent and authentication passes", func() {
			Context("with non-json body", func() {

				BeforeEach(func() {
					uamDAO.EXPECT().
						AddUserToGroup(uint(userID), username, groupName).
						Times(0)

					req, _ = http.NewRequest("POST", "/protected/group/membership/invitation", strings.NewReader("test"))
				})

				It("returns bad request", func() {
					router.ServeHTTP(recorder, req)
					assertErrorResponse(recorder, http.StatusBadRequest, "Invalid json body")
				})
			})

			Context("with json body", func() {
				var rqBody common.GroupMembershipPayload

				BeforeEach(func() {
					rqBody = common.GroupMembershipPayload{Username: username}
					rqBody.GroupName = groupName
					jsonBody, _ := json.Marshal(&rqBody)
					req, _ = http.NewRequest("POST", "/protected/group/membership/invitation", bytes.NewBuffer(jsonBody))
					req.Header.Set("Authorization", "Bearer sometoken")
				})

				Context("and membership creation fails", func() {
					Context("and request fails due to problem with the server", func() {
						BeforeEach(func() {
							uamDAO.EXPECT().
								AddUserToGroup(uint(userID), username, groupName).
								Return(myerr.NewServerError("some-error"))
						})

						It("returns internal server error", func() {
							router.ServeHTTP(recorder, req)
							assertErrorResponse(recorder, http.StatusInternalServerError, "Problem with the server, please try again later")
						})
					})

					Context("and username or group doesnt exist", func() {
						BeforeEach(func() {
							uamDAO.EXPECT().
								AddUserToGroup(uint(userID), username, groupName).
								Return(myerr.NewClientError("some-error"))
						})

						It("returns bad request", func() {
							router.ServeHTTP(recorder, req)
							assertErrorResponse(recorder, http.StatusBadRequest, "some-error")
						})
					})
				})

				Context("and membership creation succeeds", func() {
					BeforeEach(func() {
						uamDAO.EXPECT().
							AddUserToGroup(uint(userID), username, groupName).
							Return(nil)
					})

					It("returns created", func() {
						router.ServeHTTP(recorder, req)

						Expect(recorder.Code).To(Equal(http.StatusCreated))
						body := common.BasicResponse{}
						json.Unmarshal([]byte(recorder.Body.String()), &body)
						Expect(body.Status).To(Equal(http.StatusCreated))
					})
				})
			})
		})
	})

	Context("RevokeMembership", func() {
		When("request a user to be added to group is sent and authentication passes", func() {
			Context("with non-json body", func() {

				BeforeEach(func() {
					uamDAO.EXPECT().
						RemoveUserFromGroup(uint(userID), username, groupName).
						Times(0)

					req, _ = http.NewRequest("POST", "/protected/group/membership/revocation", strings.NewReader("test"))
				})

				It("returns bad request", func() {
					router.ServeHTTP(recorder, req)
					assertErrorResponse(recorder, http.StatusBadRequest, "Invalid json body")
				})
			})

			Context("with json body", func() {
				var rqBody common.GroupMembershipPayload

				BeforeEach(func() {
					rqBody = common.GroupMembershipPayload{Username: username}
					rqBody.GroupName = groupName
					jsonBody, _ := json.Marshal(&rqBody)
					req, _ = http.NewRequest("POST", "/protected/group/membership/revocation", bytes.NewBuffer(jsonBody))
					req.Header.Set("Authorization", "Bearer sometoken")
				})

				Context("and membership deletion fails", func() {
					Context("and request fails due to problem with the server", func() {
						BeforeEach(func() {
							uamDAO.EXPECT().
								RemoveUserFromGroup(uint(userID), username, groupName).
								Return(myerr.NewServerError("some-error"))
						})

						It("returns internal server error", func() {
							router.ServeHTTP(recorder, req)
							assertErrorResponse(recorder, http.StatusInternalServerError, "Problem with the server, please try again later")
						})
					})

					Context("and username or group doesnt exist", func() {
						BeforeEach(func() {
							uamDAO.EXPECT().
								RemoveUserFromGroup(uint(userID), username, groupName).
								Return(myerr.NewClientError("some-error"))
						})

						It("returns bad request", func() {
							router.ServeHTTP(recorder, req)
							assertErrorResponse(recorder, http.StatusBadRequest, "some-error")
						})
					})
				})

				Context("and membership deletion succeeds", func() {
					BeforeEach(func() {
						uamDAO.EXPECT().
							RemoveUserFromGroup(uint(userID), username, groupName).
							Return(nil)
					})

					It("returns created", func() {
						router.ServeHTTP(recorder, req)

						Expect(recorder.Code).To(Equal(http.StatusOK))
						body := common.BasicResponse{}
						json.Unmarshal([]byte(recorder.Body.String()), &body)
						Expect(body.Status).To(Equal(http.StatusOK))
					})
				})
			})
		})
	})

	Context("DeleteGroup", func() {
		When("request a user to be added to group is sent and authentication passes", func() {
			Context("with non-json body", func() {

				BeforeEach(func() {
					uamDAO.EXPECT().
						DeactivateGroup(uint(userID), groupName).
						Times(0)

					req, _ = http.NewRequest("DELETE", "/protected/group/deletion", strings.NewReader("test"))
				})

				It("returns bad request", func() {
					router.ServeHTTP(recorder, req)
					assertErrorResponse(recorder, http.StatusBadRequest, "Invalid json body")
				})
			})

			Context("with json body", func() {
				var rqBody common.GroupPayload

				BeforeEach(func() {
					rqBody = common.GroupPayload{GroupName: groupName}
					jsonBody, _ := json.Marshal(&rqBody)
					req, _ = http.NewRequest("DELETE", "/protected/group/deletion", bytes.NewBuffer(jsonBody))
					req.Header.Set("Authorization", "Bearer sometoken")
				})

				Context("and membership deletion fails", func() {
					Context("and request fails due to problem with the server", func() {
						BeforeEach(func() {
							uamDAO.EXPECT().
								DeactivateGroup(uint(userID), groupName).
								Return(myerr.NewServerError("some-error"))
						})

						It("returns internal server error", func() {
							router.ServeHTTP(recorder, req)
							assertErrorResponse(recorder, http.StatusInternalServerError, "Problem with the server, please try again later")
						})
					})

					Context("and username or group doesnt exist or user doesnt have required permissions", func() {
						BeforeEach(func() {
							uamDAO.EXPECT().
								DeactivateGroup(uint(userID), groupName).
								Return(myerr.NewClientError("some-error"))
						})

						It("returns bad request", func() {
							router.ServeHTTP(recorder, req)
							assertErrorResponse(recorder, http.StatusBadRequest, "some-error")
						})
					})
				})

				Context("and membership deletion succeeds", func() {
					BeforeEach(func() {
						uamDAO.EXPECT().
							DeactivateGroup(uint(userID), groupName).
							Return(nil)
					})

					It("returns created", func() {
						router.ServeHTTP(recorder, req)

						Expect(recorder.Code).To(Equal(http.StatusOK))
						body := common.BasicResponse{}
						json.Unmarshal([]byte(recorder.Body.String()), &body)
						Expect(body.Status).To(Equal(http.StatusOK))
					})
				})
			})
		})
	})
})
