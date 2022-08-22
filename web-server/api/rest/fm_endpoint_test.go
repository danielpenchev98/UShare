package rest_test

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"

	"github.com/danielpenchev98/UShare/web-server/api/rest"
	"github.com/danielpenchev98/UShare/web-server/internal/db/dao/dao_mocks"
	"github.com/danielpenchev98/UShare/web-server/internal/db/models"
	myerr "github.com/danielpenchev98/UShare/web-server/internal/error"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func setupRouterFmEndpoint(fmRest rest.FileManagementEndpoint, userID uint) *gin.Engine {
	r := gin.Default()

	protected := r.Group("/protected").Use(func(c *gin.Context) {
		c.Set("userID", userID)
		c.Next()
	})
	{
		protected.POST("/group/file/upload", fmRest.UploadFile)
		protected.GET("/group/file/download", fmRest.DownloadFile)
		protected.DELETE("/group/file/delete", fmRest.DeleteFile)
	}
	return r
}

func createFormFile(filePath string) (*bytes.Buffer, string) {
	file, _ := os.Open(filePath)
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", filepath.Base(file.Name()))
	io.Copy(part, file)
	writer.Close()

	return body, writer.FormDataContentType()
}

var _ = Describe("UamEndpoint", func() {
	var (
		router   *gin.Engine
		recorder *httptest.ResponseRecorder
		fmDAO    *dao_mocks.MockFmDAO
		uamDAO   *dao_mocks.MockUamDAO
		req      *http.Request
	)

	const (
		username  = "username"
		password  = "password"
		userID    = 1
		groupName = "groupName"
		groupID   = 2
		groupsDir = "."
		fileName  = "test"
		fileID    = 3
	)

	var (
		inputFilePath  string
		outputFilePath string
	)

	BeforeEach(func() {
		controller := gomock.NewController(GinkgoT())
		uamDAO = dao_mocks.NewMockUamDAO(controller)
		fmDAO = dao_mocks.NewMockFmDAO(controller)
		fmRest := rest.NewFileManagementEndpointImpl(uamDAO, fmDAO, groupsDir)

		router = setupRouterFmEndpoint(fmRest, userID)
		recorder = httptest.NewRecorder()

		inputFilePath = path.Join(groupsDir, groupName, fileName)
		outputFilePath = path.Join(groupsDir, groupName, fmt.Sprint(fileID))
	})

	Context("UploadFile", func() {
		BeforeEach(func() {
			os.Mkdir(path.Join(groupsDir, groupName), 0777)
			os.Create(inputFilePath)

		})

		AfterEach(func() {
			os.RemoveAll(path.Join(groupsDir, groupName))
		})

		When("upload request is sent and authentication passes", func() {

			Context("and 'file' key not used for the file attachment", func() {
				BeforeEach(func() {
					uamDAO.EXPECT().
						GetGroup(gomock.Any()).
						Times(0)

					uamDAO.EXPECT().
						MemberExists(gomock.Any(), gomock.Any()).
						Times(0)

					fmDAO.EXPECT().
						AddFileInfo(gomock.Any(), gomock.Any(), gomock.Any()).
						Times(0)

					req, _ = http.NewRequest("POST", "/protected/group/file/upload", nil)
					req.Header.Set("Authorization", "Bearer sometoken")
				})

				It("returns bad request error response", func() {
					router.ServeHTTP(recorder, req)
					assertErrorResponse(recorder, http.StatusBadRequest, "Problem with the file")
				})
			})

			Context("and file has been attached to the request", func() {
				var (
					form        *bytes.Buffer
					contentType string
				)

				BeforeEach(func() {
					form, contentType = createFormFile(inputFilePath)
				})

				Context("and group_name is not specified as query param", func() {
					BeforeEach(func() {
						uamDAO.EXPECT().
							GetGroup(gomock.Any()).
							Times(0)

						uamDAO.EXPECT().
							MemberExists(gomock.Any(), gomock.Any()).
							Times(0)

						fmDAO.EXPECT().
							AddFileInfo(gomock.Any(), gomock.Any(), gomock.Any()).
							Times(0)

						req, _ = http.NewRequest("POST", "/protected/group/file/upload", form)
						req.Header.Add("Content-Type", contentType)
						req.Header.Set("Authorization", "Bearer sometoken")
					})

					It("returns bad request error response", func() {
						router.ServeHTTP(recorder, req)
						assertErrorResponse(recorder, http.StatusBadRequest, "Groupname isnt specified")
					})
				})

				Context("and group_name is specified as query param", func() {
					BeforeEach(func() {
						req, _ = http.NewRequest("POST", fmt.Sprintf("/protected/group/file/upload?group_name=%s", groupName), form)
						req.Header.Add("Content-Type", contentType)
						req.Header.Set("Authorization", "Bearer sometoken")
					})

					Context("and query to get group fails fue to system failure", func() {
						BeforeEach(func() {
							uamDAO.EXPECT().
								GetGroup(groupName).
								Return(models.Group{}, myerr.NewServerError("test-error"))

							uamDAO.EXPECT().
								MemberExists(gomock.Any(), gomock.Any()).
								Times(0)

							fmDAO.EXPECT().
								AddFileInfo(gomock.Any(), gomock.Any(), gomock.Any()).
								Times(0)
						})

						It("returns internal server error response", func() {
							router.ServeHTTP(recorder, req)
							assertErrorResponse(recorder, http.StatusInternalServerError, "Problem with the server")
						})
					})

					Context("and query to get group succeeds", func() {
						Context("and group not found", func() {
							BeforeEach(func() {
								uamDAO.EXPECT().
									GetGroup(groupName).
									Return(models.Group{}, myerr.NewClientError("test-error"))

								uamDAO.EXPECT().
									MemberExists(gomock.Any(), gomock.Any()).
									Times(0)

								fmDAO.EXPECT().
									AddFileInfo(gomock.Any(), gomock.Any(), gomock.Any()).
									Times(0)
							})

							It("returns bad request response", func() {
								router.ServeHTTP(recorder, req)
								assertErrorResponse(recorder, http.StatusBadRequest, "test-error")
							})
						})

						Context("and group is found", func() {
							var group models.Group
							BeforeEach(func() {
								group = models.Group{
									Name: group.Name,
									ID:   groupID,
								}
							})

							Context("and query to check if membership exists fails", func() {
								BeforeEach(func() {
									gomock.InOrder(
										uamDAO.EXPECT().
											GetGroup(groupName).
											Return(group, nil),

										uamDAO.EXPECT().
											MemberExists(uint(userID), uint(groupID)).
											Return(false, myerr.NewServerError("test-error")),
									)

									fmDAO.EXPECT().
										AddFileInfo(gomock.Any(), gomock.Any(), gomock.Any()).
										Times(0)
								})

								It("returns internal server error", func() {
									router.ServeHTTP(recorder, req)
									assertErrorResponse(recorder, http.StatusInternalServerError, "Problem with the server")
								})
							})

							Context("and query to check if membership exists succeeds", func() {
								Context("and membership doesnt exist", func() {
									BeforeEach(func() {
										gomock.InOrder(
											uamDAO.EXPECT().
												GetGroup(groupName).
												Return(group, nil),

											uamDAO.EXPECT().
												MemberExists(uint(userID), uint(groupID)).
												Return(false, nil),
										)

										fmDAO.EXPECT().
											AddFileInfo(gomock.Any(), gomock.Any(), gomock.Any()).
											Times(0)
									})

									It("returns bad request error response", func() {
										router.ServeHTTP(recorder, req)
										assertErrorResponse(recorder, http.StatusBadRequest, "Invalid user input")
									})
								})

								Context("and membership exists", func() {
									Context("and query for adding file info fails", func() {
										BeforeEach(func() {
											gomock.InOrder(
												uamDAO.EXPECT().
													GetGroup(groupName).
													Return(group, nil),

												uamDAO.EXPECT().
													MemberExists(uint(userID), uint(groupID)).
													Return(true, nil),

												fmDAO.EXPECT().
													AddFileInfo(uint(userID), fileName, groupName).
													Return(uint(fileID), myerr.NewServerError("test-error")),
											)

										})

										It("returns internal server error", func() {
											router.ServeHTTP(recorder, req)
											assertErrorResponse(recorder, http.StatusInternalServerError, "Problem with the server")
										})
									})

									Context("and file is uploaded successfully", func() {
										BeforeEach(func() {
											gomock.InOrder(
												uamDAO.EXPECT().
													GetGroup(groupName).
													Return(group, nil),

												uamDAO.EXPECT().
													MemberExists(uint(userID), uint(groupID)).
													Return(true, nil),

												fmDAO.EXPECT().
													AddFileInfo(uint(userID), fileName, groupName).
													Return(uint(fileID), nil),
											)

										})

										It("succeeds", func() {
											router.ServeHTTP(recorder, req)
											Expect(recorder.Code).To(Equal(http.StatusCreated))
											_, err := os.Stat(outputFilePath)
											Expect(err).To(BeNil())
										})
									})
								})
							})
						})

					})
				})
			})

		})
	})
})
