package auth_test

import (
	"os"
	"strconv"
	"time"

	"github.com/danielpenchev98/UShare/web-server/internal/auth"
	myerr "github.com/danielpenchev98/UShare/web-server/internal/error"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Auth module", func() {

	const (
		secretKey     = "SECRET"
		secretVal     = "secret"
		issuerKey     = "ISSUER"
		issuerVal     = "issuer"
		expirationKey = "EXPIRATION"
		expirationVal = 24
	)

	BeforeEach(func() {
		os.Clearenv()
	})

	Context("NewJwtCreatorImpl", func() {
		When("Creating new Jwt creator", func() {
			Context("and secret env variable is missing", func() {
				It("returns error", func() {
					_, err := auth.NewJwtCreatorImpl()
					Expect(err).To(HaveOccurred())
					_, ok := err.(*myerr.ServerError)
					Expect(ok).To(Equal(true))
				})
			})

			Context("and secret env variable exists", func() {
				BeforeEach(func() {
					os.Setenv(secretKey, secretVal)
				})
				AfterEach(func() {
					os.Unsetenv(secretKey)
				})

				Context("and issuer env variable is missing", func() {
					It("returns error", func() {
						_, err := auth.NewJwtCreatorImpl()
						Expect(err).To(HaveOccurred())
						_, ok := err.(*myerr.ServerError)
						Expect(ok).To(Equal(true))
					})
				})

				Context("and issuer env variable exists", func() {
					BeforeEach(func() {
						os.Setenv(issuerKey, issuerVal)
					})

					AfterEach(func() {
						os.Unsetenv(issuerKey)
					})

					Context("and expiration env variable is missing", func() {
						It("returns error", func() {
							_, err := auth.NewJwtCreatorImpl()
							Expect(err).To(HaveOccurred())
							_, ok := err.(*myerr.ServerError)
							Expect(ok).To(Equal(true))
						})
					})

					Context("and expiration env vairable exitst", func() {
						Context("and expiration variable is in illegal format", func() {
							BeforeEach(func() {
								os.Setenv(expirationKey, "wrong-format")
							})

							AfterEach(func() {
								os.Unsetenv(expirationKey)
							})

							It("returns error", func() {
								_, err := auth.NewJwtCreatorImpl()
								Expect(err).To(HaveOccurred())
								_, ok := err.(*myerr.ServerError)
								Expect(ok).To(Equal(true))
							})
						})

						Context("and expiration variable is in legal format", func() {
							BeforeEach(func() {
								os.Setenv(expirationKey, strconv.Itoa(expirationVal))
							})

							AfterEach(func() {
								os.Unsetenv(expirationKey)
							})

							It("succeeds", func() {
								actualResult, err := auth.NewJwtCreatorImpl()
								Expect(err).NotTo(HaveOccurred())
								expectedResult := &auth.JwtCreatorImpl{
									Secret:          secretVal,
									Issuer:          issuerVal,
									ExpirationHours: expirationVal,
								}
								Expect(actualResult).To(Equal(expectedResult))
							})
						})
					})
				})
			})
		})
	})
	Context("JwtCreator", func() {
		var jwtCreator auth.JwtCreator
		BeforeEach(func() {
			jwtCreator = &auth.JwtCreatorImpl{
				Secret:          secretVal,
				Issuer:          issuerVal,
				ExpirationHours: expirationVal,
			}
		})

		Context("GenerateToken", func() {
			When("creating a token", func() {
				It("succeeds", func() {
					_, err := jwtCreator.GenerateToken(1)
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		Context("ValidateToken", func() {
			var token string
			const userID = 1

			When("validating legal token", func() {
				BeforeEach(func() {
					token, _ = jwtCreator.GenerateToken(userID)
				})

				It("classifies the token as legal", func() {
					claims, err := jwtCreator.ValidateToken(token)
					Expect(err).NotTo(HaveOccurred())
					Expect(claims.UserID).To(Equal(uint(userID)))
					Expect(claims.Issuer).To(Equal(issuerVal))
				})
			})

			When("validating expired token", func() {

				BeforeEach(func() {
					jwtCreator = &auth.JwtCreatorImpl{
						Secret:          secretVal,
						Issuer:          issuerVal,
						ExpirationHours: 0,
					}

					token, _ = jwtCreator.GenerateToken(userID)
				})

				It("returns error", func() {
					time.Sleep(1 * time.Second)
					_, err := jwtCreator.ValidateToken(token)
					Expect(err).To(HaveOccurred())
					_, ok := err.(*myerr.ClientError)
					Expect(ok).To(Equal(true))
				})
			})

			When("validating wrong type of token", func() {
				It("returns error", func() {
					_, err := jwtCreator.ValidateToken("wrong-type")
					Expect(err).To(HaveOccurred())
					_, ok := err.(*myerr.ClientError)
					Expect(ok).To(Equal(true))
				})
			})
		})
	})

})
