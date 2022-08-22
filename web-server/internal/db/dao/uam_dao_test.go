package dao

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"regexp"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/danielpenchev98/UShare/web-server/internal/db/models"
	myerr "github.com/danielpenchev98/UShare/web-server/internal/error"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Any struct{}

// Match satisfies sqlmock.Argument interface
func (a Any) Match(v driver.Value) bool {
	return true
}

var _ = Describe("UamDAO", func() {
	var (
		uamDao UamDAO
		mock   sqlmock.Sqlmock
	)

	const (
		username  = "username"
		password  = "password"
		groupName = "test-group"
		userID    = 1
		groupID   = 2
	)

	BeforeEach(func() {
		var (
			db  *sql.DB
			err error
		)

		db, mock, err = sqlmock.New()
		Expect(err).NotTo(HaveOccurred())

		gdb, err := gorm.Open(postgres.New(postgres.Config{
			Conn: db,
		}), &gorm.Config{})
		Expect(err).NotTo(HaveOccurred())

		uamDao = NewUamDAOImpl(gdb)
	})

	AfterEach(func() {
		err := mock.ExpectationsWereMet()
		Expect(err).ShouldNot(HaveOccurred())
	})

	Context("CreateUser", func() {
		When("request if user exists fails", func() {
			BeforeEach(func() {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) FROM "users"`)).
					WithArgs(username).
					WillReturnError(fmt.Errorf("some error"))
				mock.ExpectRollback()

			})

			It("propagates error", func() {
				err := uamDao.CreateUser(username, password)
				Expect(err).To(HaveOccurred())
				_, ok := err.(*myerr.ServerError)
				Expect(ok).To(Equal(true))
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})
		})

		When("request if user exists is succesful", func() {
			Context("and user already exists", func() {
				BeforeEach(func() {
					rows := sqlmock.NewRows([]string{"count"}).AddRow(2)

					mock.ExpectBegin()
					mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) FROM "users"`)).
						WithArgs(username).
						WillReturnRows(rows)
					mock.ExpectRollback()
				})

				It("propagates error", func() {
					err := uamDao.CreateUser(username, password)
					Expect(err).To(HaveOccurred())
					_, ok := err.(*myerr.ClientError)
					Expect(ok).To(Equal(true))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			Context("and user doesnt exist", func() {
				var rows *sqlmock.Rows
				BeforeEach(func() {
					rows = sqlmock.NewRows([]string{"count"}).AddRow(0)
				})

				Context("and creation query is successful", func() {
					BeforeEach(func() {
						mock.ExpectBegin()
						mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) FROM "users"`)).
							WithArgs(username).
							WillReturnRows(rows)
						mock.ExpectQuery("INSERT INTO \"users\"").
							WithArgs(Any{}, Any{}, username, password). // driver.NamedValue - {Name: Ordinal:1 Value:2020-12-28 01:22:59.344298 +0200 EET}"
							WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
						mock.ExpectCommit()
					})

					It("succeeds", func() {
						err := uamDao.CreateUser(username, password)
						Expect(err).NotTo(HaveOccurred())
						Expect(mock.ExpectationsWereMet()).To(BeNil())
					})
				})

				Context("and creation query fails", func() {
					BeforeEach(func() {
						mock.ExpectBegin()
						mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) FROM "users"`)).
							WithArgs(username).
							WillReturnRows(rows)
						mock.ExpectQuery("INSERT INTO \"users\"").
							WithArgs(Any{}, Any{}, username, password). // driver.NamedValue - {Name: Ordinal:1 Value:2020-12-28 01:22:59.344298 +0200 EET}"
							WillReturnError(fmt.Errorf("some error"))
						mock.ExpectRollback()
					})

					It("propagates error", func() {
						err := uamDao.CreateUser(username, password)
						Expect(err).To(HaveOccurred())
						_, ok := err.(*myerr.ServerError)
						Expect(ok).To(Equal(true))
						Expect(mock.ExpectationsWereMet()).To(BeNil())
					})
				})
			})
		})

	})

	Context("Delete user", func() {
		When("request if user exists fails", func() {
			BeforeEach(func() {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) FROM "users"`)).
					WithArgs(userID).
					WillReturnError(fmt.Errorf("some error"))
			})

			It("propagates error", func() {
				err := uamDao.DeleteUser(uint(userID))
				Expect(err).To(HaveOccurred())
				_, ok := err.(*myerr.ServerError)
				Expect(ok).To(Equal(true))
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})
		})

		When("request if user exists is succesful", func() {
			Context("and user does not exist", func() {
				BeforeEach(func() {
					rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
					mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) FROM "users"`)).
						WithArgs(userID).
						WillReturnRows(rows)

				})

				It("propagates error", func() {
					err := uamDao.DeleteUser(uint(userID))
					Expect(err).To(HaveOccurred())
					_, ok := err.(*myerr.ItemNotFoundError)
					Expect(ok).To(Equal(true))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			Context("and user exists", func() {
				var existCountRows *sqlmock.Rows
				BeforeEach(func() {
					existCountRows = sqlmock.NewRows([]string{"count"}).AddRow(1)
				})

				Context("and deletion query fails", func() {
					BeforeEach(func() {
						mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) FROM "users"`)).
							WithArgs(userID).
							WillReturnRows(existCountRows)
						mock.ExpectExec("DELETE FROM \"users\"").
							WithArgs(userID). // driver.NamedValue - {Name: Ordinal:1 Value:2020-12-28 01:22:59.344298 +0200 EET}"
							WillReturnError(fmt.Errorf("some error"))
					})

					It("propagates error", func() {
						err := uamDao.DeleteUser(userID)
						Expect(err).To(HaveOccurred())
						_, ok := err.(*myerr.ServerError)
						Expect(ok).To(Equal(true))
						Expect(mock.ExpectationsWereMet()).To(BeNil())
					})
				})

				Context("and deletion query is successful", func() {
					BeforeEach(func() {
						mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) FROM "users"`)).
							WithArgs(userID).
							WillReturnRows(existCountRows)
						mock.ExpectExec("DELETE FROM \"users\"").
							WithArgs(userID).
							WillReturnResult(sqlmock.NewResult(0, 1))
					})

					It("succeeds", func() {
						err := uamDao.DeleteUser(uint(userID))
						Expect(err).NotTo(HaveOccurred())
						Expect(mock.ExpectationsWereMet()).To(BeNil())
					})
				})
			})
		})
	})

	Context("GetUser", func() {
		When("request to get userID fails", func() {
			BeforeEach(func() {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).
					WithArgs(username).
					WillReturnError(fmt.Errorf("some error"))
			})

			It("propagates error", func() {
				_, err := uamDao.GetUser(username)
				Expect(err).To(HaveOccurred())
				_, ok := err.(*myerr.ServerError)
				Expect(ok).To(Equal(true))
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})
		})

		When("request to get userID is succesful", func() {
			Context("and user does not exist", func() {
				BeforeEach(func() {
					mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).
						WithArgs(username).
						WillReturnError(gorm.ErrRecordNotFound)
				})

				It("propagates error", func() {
					_, err := uamDao.GetUser(username)
					Expect(err).To(HaveOccurred())
					_, ok := err.(*myerr.ItemNotFoundError)
					Expect(ok).To(Equal(true))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			Context("and user exists", func() {
				var mockTime time.Time

				BeforeEach(func() {
					mockTime = time.Now()
					rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "username", "password"}).AddRow(1, mockTime, mockTime, username, password)
					mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).
						WithArgs(username).
						WillReturnRows(rows)
				})

				It("succeds", func() {
					user, err := uamDao.GetUser(username)
					Expect(err).NotTo(HaveOccurred())

					Expect(user.ID).To(Equal(uint(1)))
					Expect(user.Username).To(Equal(username))
					Expect(user.Password).To(Equal(password))
					Expect(user.CreatedAt).To(Equal(mockTime))
					Expect(user.UpdatedAt).To(Equal(mockTime))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})
		})
	})

	Context("CreateGroup", func() {
		When("request if group exists fails", func() {
			BeforeEach(func() {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) FROM "groups"`)).
					WithArgs(groupName).
					WillReturnError(fmt.Errorf("some error"))
				mock.ExpectRollback()
			})

			It("propagates error", func() {
				err := uamDao.CreateGroup(uint(userID), groupName)
				Expect(err).To(HaveOccurred())
				_, ok := err.(*myerr.ServerError)
				Expect(ok).To(Equal(true))
				Expect(mock.ExpectationsWereMet()).To(BeNil())
			})
		})

		When("request if group exists is succesful", func() {
			Context("and user already exists", func() {
				BeforeEach(func() {
					rows := sqlmock.NewRows([]string{"count"}).AddRow(1)

					mock.ExpectBegin()
					mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) FROM "groups"`)).
						WithArgs(groupName).
						WillReturnRows(rows)
					mock.ExpectRollback()
				})

				It("propagates error", func() {
					err := uamDao.CreateGroup(uint(userID), groupName)
					Expect(err).To(HaveOccurred())
					_, ok := err.(*myerr.ClientError)
					Expect(ok).To(Equal(true))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			Context("and group doesnt exist", func() {
				var zeroCountRows *sqlmock.Rows

				BeforeEach(func() {
					zeroCountRows = sqlmock.NewRows([]string{"count"}).AddRow(0)
				})

				Context("and group creation query fails", func() {
					BeforeEach(func() {
						mock.ExpectBegin()
						mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) FROM "groups"`)).
							WithArgs(groupName).
							WillReturnRows(zeroCountRows)
						mock.ExpectQuery("INSERT INTO \"groups\"").
							WithArgs(Any{}, Any{}, groupName, userID, true). // driver.NamedValue - {Name: Ordinal:1 Value:2020-12-28 01:22:59.344298 +0200 EET}"
							WillReturnError(fmt.Errorf("some error"))
						mock.ExpectRollback()
					})

					It("propagates error", func() {
						err := uamDao.CreateGroup(uint(userID), groupName)
						Expect(err).To(HaveOccurred())
						_, ok := err.(*myerr.ServerError)
						Expect(ok).To(Equal(true))
						Expect(mock.ExpectationsWereMet()).To(BeNil())
					})
				})

				Context("and group creation query is successful", func() {
					var creationRows *sqlmock.Rows
					var group models.Group

					BeforeEach(func() {
						group = models.Group{
							Name:    groupName,
							OwnerID: uint(userID),
						}
						group.ID = uint(userID)
						creationRows = sqlmock.NewRows([]string{"id"}).AddRow(group.ID)
					})

					Context("and membership creation query fails", func() {
						BeforeEach(func() {
							mock.ExpectBegin()
							mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) FROM "groups"`)).
								WithArgs(groupName).
								WillReturnRows(zeroCountRows)
							mock.ExpectQuery("INSERT INTO \"groups\"").
								WithArgs(Any{}, Any{}, groupName, userID, true). // driver.NamedValue - {Name: Ordinal:1 Value:2020-12-28 01:22:59.344298 +0200 EET}"
								WillReturnRows(creationRows)
							mock.ExpectQuery("INSERT INTO \"memberships\"").
								WithArgs(Any{}, Any{}, group.ID, group.OwnerID). // driver.NamedValue - {Name: Ordinal:1 Value:2020-12-28 01:22:59.344298 +0200 EET}"
								WillReturnError(fmt.Errorf("some error"))
							mock.ExpectRollback()
						})

						It("propagates error", func() {
							err := uamDao.CreateGroup(uint(userID), groupName)
							Expect(err).To(HaveOccurred())
							_, ok := err.(*myerr.ServerError)
							Expect(ok).To(Equal(true))
							Expect(mock.ExpectationsWereMet()).To(BeNil())
						})
					})

					Context("and membership creation query is successful", func() {
						BeforeEach(func() {
							mock.ExpectBegin()
							mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) FROM "groups"`)).
								WithArgs(groupName).
								WillReturnRows(zeroCountRows)
							mock.ExpectQuery("INSERT INTO \"groups\"").
								WithArgs(Any{}, Any{}, groupName, userID, true). // driver.NamedValue - {Name: Ordinal:1 Value:2020-12-28 01:22:59.344298 +0200 EET}"
								WillReturnRows(creationRows)
							mock.ExpectQuery("INSERT INTO \"memberships\"").
								WithArgs(Any{}, Any{}, group.ID, group.OwnerID). // driver.NamedValue - {Name: Ordinal:1 Value:2020-12-28 01:22:59.344298 +0200 EET}"
								WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
							mock.ExpectCommit()
						})

						It("succeeds", func() {
							err := uamDao.CreateGroup(uint(userID), groupName)
							Expect(err).NotTo(HaveOccurred())
							Expect(mock.ExpectationsWereMet()).To(BeNil())
						})
					})
				})
			})
		})
	})

	Context("AddUserToGroup", func() {
		When("get group request fails", func() {
			Context("and there is a problem with the database", func() {
				BeforeEach(func() {
					mock.ExpectBegin()
					mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "groups"`)).
						WithArgs(groupName).
						WillReturnError(fmt.Errorf("some error"))
					mock.ExpectRollback()
				})

				It("propagates error", func() {
					err := uamDao.AddUserToGroup(uint(userID), username, groupName)
					Expect(err).To(HaveOccurred())
					_, ok := err.(*myerr.ServerError)
					Expect(ok).To(Equal(true))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			Context("and group doesnt exist", func() {
				BeforeEach(func() {
					mock.ExpectBegin()
					mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "groups"`)).
						WithArgs(groupName).
						WillReturnError(gorm.ErrRecordNotFound)
					mock.ExpectRollback()
				})

				It("propagates error", func() {
					err := uamDao.AddUserToGroup(uint(userID), username, groupName)
					Expect(err).To(HaveOccurred())
					_, ok := err.(*myerr.ItemNotFoundError)
					Expect(ok).To(Equal(true))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})
		})

		When("and get group request is sucessfull", func() {
			Context("and group is deactivated", func() {
				BeforeEach(func() {
					rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "owner_id", "active"}).
						AddRow(1, time.Now(), time.Now(), groupName, userID, false)
					mock.ExpectBegin()
					mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "groups"`)).
						WithArgs(groupName).
						WillReturnRows(rows)
					mock.ExpectRollback()
				})

				It("propagates error", func() {
					err := uamDao.AddUserToGroup(uint(userID), username, groupName)
					Expect(err).To(HaveOccurred())
					_, ok := err.(*myerr.ClientError)
					Expect(ok).To(Equal(true))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			Context("and group is active", func() {
				Context("and you arent the owner of the group", func() {
					BeforeEach(func() {
						rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "owner_id", "active"}).
							AddRow(1, time.Now(), time.Now(), groupName, userID+1, true)
						mock.ExpectBegin()
						mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "groups"`)).
							WithArgs(groupName).
							WillReturnRows(rows)
						mock.ExpectRollback()
					})

					It("propagates error", func() {
						err := uamDao.AddUserToGroup(uint(userID), username, groupName)
						Expect(err).To(HaveOccurred())
						_, ok := err.(*myerr.ClientError)
						Expect(ok).To(Equal(true))
						Expect(mock.ExpectationsWereMet()).To(BeNil())
					})
				})

				Context("and you are the owner of the group", func() {
					var groupRow *sqlmock.Rows
					BeforeEach(func() {
						groupRow = sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "owner_id", "active"}).
							AddRow(groupID, time.Now(), time.Now(), groupName, userID, true)
					})

					Context("and user get request fails", func() {
						Context("problem with the database", func() {
							BeforeEach(func() {
								mock.ExpectBegin()
								mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "groups"`)).
									WithArgs(groupName).
									WillReturnRows(groupRow)
								mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).
									WithArgs(username).
									WillReturnError(fmt.Errorf("some error"))
								mock.ExpectRollback()
							})

							It("propagates error", func() {
								err := uamDao.AddUserToGroup(uint(userID), username, groupName)
								Expect(err).To(HaveOccurred())
								_, ok := err.(*myerr.ServerError)
								Expect(ok).To(Equal(true))
								Expect(mock.ExpectationsWereMet()).To(BeNil())
							})
						})

						Context("and user does not exist", func() {
							BeforeEach(func() {
								mock.ExpectBegin()
								mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "groups"`)).
									WithArgs(groupName).
									WillReturnRows(groupRow)
								mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).
									WithArgs(username).
									WillReturnError(gorm.ErrRecordNotFound)
								mock.ExpectRollback()
							})

							It("propagates error", func() {
								err := uamDao.AddUserToGroup(uint(userID), username, groupName)
								Expect(err).To(HaveOccurred())
								_, ok := err.(*myerr.ItemNotFoundError)
								Expect(ok).To(Equal(true))
								Expect(mock.ExpectationsWereMet()).To(BeNil())
							})
						})

					})

					Context("and user request is successfull", func() {
						var userRows *sqlmock.Rows
						BeforeEach(func() {
							userRows = sqlmock.NewRows([]string{"id", "created_at", "updated_at", "username", "password"}).
								AddRow(userID, time.Now(), time.Now(), username, password)
						})

						Context("and request if membership exists fails", func() {
							Context("and problem with lookup in the db for the membership", func() {
								BeforeEach(func() {
									mock.ExpectBegin()
									mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "groups"`)).
										WithArgs(groupName).
										WillReturnRows(groupRow)
									mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).
										WithArgs(username).
										WillReturnRows(userRows)
									mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) FROM "memberships"`)).
										WithArgs(groupID, userID).
										WillReturnError(fmt.Errorf("some error"))
									mock.ExpectRollback()
								})

								It("propagates error", func() {
									err := uamDao.AddUserToGroup(uint(userID), username, groupName)
									Expect(err).To(HaveOccurred())
									_, ok := err.(*myerr.ServerError)
									Expect(ok).To(Equal(true))
									Expect(mock.ExpectationsWereMet()).To(BeNil())
								})
							})
							Context("and memberships found", func() {
								BeforeEach(func() {
									mock.ExpectBegin()
									mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "groups"`)).
										WithArgs(groupName).
										WillReturnRows(groupRow)
									mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).
										WithArgs(username).
										WillReturnRows(userRows)
									mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) FROM "memberships"`)).
										WithArgs(groupID, userID).
										WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
									mock.ExpectRollback()
								})

								It("propagates error", func() {
									err := uamDao.AddUserToGroup(uint(userID), username, groupName)
									Expect(err).To(HaveOccurred())
									_, ok := err.(*myerr.ClientError)
									Expect(ok).To(Equal(true))
									Expect(mock.ExpectationsWereMet()).To(BeNil())
								})
							})
						})

						Context("and request if membership exists is successful", func() {
							Context("and creation of membership fails", func() {
								BeforeEach(func() {
									mock.ExpectBegin()
									mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "groups"`)).
										WithArgs(groupName).
										WillReturnRows(groupRow)
									mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).
										WithArgs(username).
										WillReturnRows(userRows)
									mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) FROM "memberships"`)).
										WithArgs(groupID, userID).
										WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
									mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "memberships"`)).
										WithArgs(Any{}, Any{}, groupID, userID).
										WillReturnError(fmt.Errorf("some error"))
									mock.ExpectRollback()
								})

								It("propagates error", func() {
									err := uamDao.AddUserToGroup(uint(userID), username, groupName)
									Expect(err).To(HaveOccurred())
									_, ok := err.(*myerr.ServerError)
									Expect(ok).To(Equal(true))
									Expect(mock.ExpectationsWereMet()).To(BeNil())
								})
							})
							Context("and creation succeeds", func() {
								BeforeEach(func() {
									mock.ExpectBegin()
									mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "groups"`)).
										WithArgs(groupName).
										WillReturnRows(groupRow)
									mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).
										WithArgs(username).
										WillReturnRows(userRows)
									mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) FROM "memberships"`)).
										WithArgs(groupID, userID).
										WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
									mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "memberships"`)).
										WithArgs(Any{}, Any{}, groupID, userID).
										WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
									mock.ExpectCommit()
								})

								It("returns no error", func() {
									err := uamDao.AddUserToGroup(uint(userID), username, groupName)
									Expect(err).NotTo(HaveOccurred())
									Expect(mock.ExpectationsWereMet()).To(BeNil())
								})
							})
						})
					})
				})
			})

		})

	})

	Context("RemoveUserFromGroup", func() {

		When("get group request fails", func() {
			Context("problem with the database", func() {
				BeforeEach(func() {
					mock.ExpectBegin()
					mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "groups"`)).
						WithArgs(groupName).
						WillReturnError(fmt.Errorf("some error"))
					mock.ExpectRollback()
				})

				It("propagates error", func() {
					err := uamDao.RemoveUserFromGroup(uint(userID), username, groupName)
					Expect(err).To(HaveOccurred())
					_, ok := err.(*myerr.ServerError)
					Expect(ok).To(Equal(true))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			Context("and group doesnt exist", func() {
				BeforeEach(func() {
					mock.ExpectBegin()
					mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "groups"`)).
						WithArgs(groupName).
						WillReturnError(gorm.ErrRecordNotFound)
					mock.ExpectRollback()
				})

				It("propagates error", func() {
					err := uamDao.RemoveUserFromGroup(uint(userID), username, groupName)
					Expect(err).To(HaveOccurred())
					_, ok := err.(*myerr.ItemNotFoundError)
					Expect(ok).To(Equal(true))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})
		})

		When("get group request succeeds", func() {
			Context("and group is deactivated", func() {
				BeforeEach(func() {
					deactivatedGroupRows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "owner_id", "active"}).
						AddRow(groupID, time.Now(), time.Now(), groupName, userID, false)
					mock.ExpectBegin()
					mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "groups"`)).
						WithArgs(groupName).
						WillReturnRows(deactivatedGroupRows)
					mock.ExpectRollback()
				})

				It("propagates error", func() {
					err := uamDao.RemoveUserFromGroup(uint(userID), username, groupName)
					Expect(err).To(HaveOccurred())
					_, ok := err.(*myerr.ClientError)
					Expect(ok).To(Equal(true))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			Context("and group is active", func() {
				var groupRow *sqlmock.Rows
				BeforeEach(func() {
					groupRow = sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "owner_id", "active"}).
						AddRow(groupID, time.Now(), time.Now(), groupName, userID, true)
				})

				Context("and get user request fails", func() {
					Context("problem with the database", func() {
						BeforeEach(func() {
							mock.ExpectBegin()
							mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "groups"`)).
								WithArgs(groupName).
								WillReturnRows(groupRow)
							mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).
								WithArgs(username).
								WillReturnError(fmt.Errorf("some error"))
							mock.ExpectRollback()
						})

						It("propagates error", func() {
							err := uamDao.RemoveUserFromGroup(uint(userID), username, groupName)
							Expect(err).To(HaveOccurred())
							_, ok := err.(*myerr.ServerError)
							Expect(ok).To(Equal(true))
							Expect(mock.ExpectationsWereMet()).To(BeNil())
						})
					})

					Context("user does not exist", func() {
						BeforeEach(func() {
							mock.ExpectBegin()
							mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "groups"`)).
								WithArgs(groupName).
								WillReturnRows(groupRow)
							mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).
								WithArgs(username).
								WillReturnError(gorm.ErrRecordNotFound)
							mock.ExpectRollback()
						})

						It("propagates error", func() {
							err := uamDao.RemoveUserFromGroup(uint(userID), username, groupName)
							Expect(err).To(HaveOccurred())
							_, ok := err.(*myerr.ItemNotFoundError)
							Expect(ok).To(Equal(true))
							Expect(mock.ExpectationsWereMet()).To(BeNil())
						})
					})

				})

				Context("and get user request succeeds", func() {
					var (
						userRows     *sqlmock.Rows
						targetUserID uint
					)

					BeforeEach(func() {
						targetUserID = userID + 1
						userRows = sqlmock.NewRows([]string{"id", "created_at", "updated_at", "username", "password"}).
							AddRow(targetUserID, time.Now(), time.Now(), username, password)
					})

					Context("and you arent the owner of the group or the targeted user", func() {
						BeforeEach(func() {
							mock.ExpectBegin()
							mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "groups"`)).
								WithArgs(groupName).
								WillReturnRows(groupRow)
							mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).
								WithArgs(username).
								WillReturnRows(userRows)
							mock.ExpectRollback()
						})

						It("propagates error", func() {
							err := uamDao.RemoveUserFromGroup(uint(userID+2), username, groupName)
							Expect(err).To(HaveOccurred())
							_, ok := err.(*myerr.ClientError)
							Expect(ok).To(Equal(true))
							Expect(mock.ExpectationsWereMet()).To(BeNil())
						})
					})

					Context("and you are the owner of the group and target of the deletion", func() {

						BeforeEach(func() {
							rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "username", "password"}).
								AddRow(userID, time.Now(), time.Now(), username, password)

							mock.ExpectBegin()
							mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "groups"`)).
								WithArgs(groupName).
								WillReturnRows(groupRow)
							mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).
								WithArgs(username).
								WillReturnRows(rows)
							mock.ExpectRollback()
						})

						It("propagates error", func() {
							err := uamDao.RemoveUserFromGroup(uint(userID), username, groupName)
							Expect(err).To(HaveOccurred())
							_, ok := err.(*myerr.ClientError)
							Expect(ok).To(Equal(true))
							Expect(mock.ExpectationsWereMet()).To(BeNil())
						})
					})

					Context("and you are the owner(not the targeted user) or the targeted users", func() {
						Context("and delete membership request fails", func() {
							BeforeEach(func() {
								mock.ExpectBegin()
								mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "groups"`)).
									WithArgs(groupName).
									WillReturnRows(groupRow)
								mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).
									WithArgs(username).
									WillReturnRows(userRows)
								mock.ExpectExec("DELETE FROM \"memberships\"").
									WithArgs(targetUserID, groupID).
									WillReturnError(fmt.Errorf("some error"))
								mock.ExpectRollback()
							})

							It("propagates error", func() {
								err := uamDao.RemoveUserFromGroup(uint(userID), username, groupName)
								Expect(err).To(HaveOccurred())
								_, ok := err.(*myerr.ServerError)
								Expect(ok).To(Equal(true))
								Expect(mock.ExpectationsWereMet()).To(BeNil())
							})
						})
						Context("and delete membership requests succeeds", func() {
							Context("and membership did not exist", func() {
								BeforeEach(func() {
									mock.ExpectBegin()
									mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "groups"`)).
										WithArgs(groupName).
										WillReturnRows(groupRow)
									mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).
										WithArgs(username).
										WillReturnRows(userRows)
									mock.ExpectExec("DELETE FROM \"memberships\"").
										WithArgs(targetUserID, groupID).
										WillReturnResult(sqlmock.NewResult(0, 0))
									mock.ExpectRollback()
								})

								It("propagates error", func() {
									err := uamDao.RemoveUserFromGroup(uint(userID), username, groupName)
									Expect(err).To(HaveOccurred())
									_, ok := err.(*myerr.ClientError)
									Expect(ok).To(Equal(true))
									Expect(mock.ExpectationsWereMet()).To(BeNil())
								})
							})

							Context("and existing membership successfully deleted", func() {
								BeforeEach(func() {
									mock.ExpectBegin()
									mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "groups"`)).
										WithArgs(groupName).
										WillReturnRows(groupRow)
									mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).
										WithArgs(username).
										WillReturnRows(userRows)
									mock.ExpectExec("DELETE FROM \"memberships\"").
										WithArgs(targetUserID, groupID).
										WillReturnResult(sqlmock.NewResult(0, 1))
									mock.ExpectCommit()
								})

								It("propagates error", func() {
									err := uamDao.RemoveUserFromGroup(uint(userID), username, groupName)
									Expect(err).NotTo(HaveOccurred())
									Expect(mock.ExpectationsWereMet()).To(BeNil())
								})
							})
						})
					})
				})
			})

		})

	})

	Context("DeactivateGroup", func() {
		When("get group request fails", func() {
			Context("problem with the database", func() {
				BeforeEach(func() {
					mock.ExpectBegin()
					mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "groups"`)).
						WithArgs(groupName).
						WillReturnError(fmt.Errorf("some error"))
					mock.ExpectRollback()
				})

				It("propagates error", func() {
					err := uamDao.DeactivateGroup(uint(userID), groupName)
					Expect(err).To(HaveOccurred())
					_, ok := err.(*myerr.ServerError)
					Expect(ok).To(Equal(true))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			Context("and group doesnt exist", func() {
				BeforeEach(func() {
					mock.ExpectBegin()
					mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "groups"`)).
						WithArgs(groupName).
						WillReturnError(gorm.ErrRecordNotFound)
					mock.ExpectRollback()
				})

				It("propagates error", func() {
					err := uamDao.DeactivateGroup(uint(userID), groupName)
					Expect(err).To(HaveOccurred())
					_, ok := err.(*myerr.ItemNotFoundError)
					Expect(ok).To(Equal(true))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})
		})

		When("get group request succeeds", func() {
			Context("and group is deactivated", func() {
				BeforeEach(func() {
					deactivatedGroupRow := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "owner_id", "active"}).
						AddRow(groupID, time.Now(), time.Now(), groupName, userID, false)
					mock.ExpectBegin()
					mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "groups"`)).
						WithArgs(groupName).
						WillReturnRows(deactivatedGroupRow)
					mock.ExpectRollback()
				})

				It("propagates error", func() {
					err := uamDao.DeactivateGroup(uint(userID), groupName)
					Expect(err).To(HaveOccurred())
					_, ok := err.(*myerr.ClientError)
					Expect(ok).To(Equal(true))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			Context("and group is active", func() {
				var groupRow *sqlmock.Rows
				BeforeEach(func() {
					groupRow = sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "owner_id", "active"}).
						AddRow(groupID, time.Now(), time.Now(), groupName, userID, true)
				})

				Context("and you are not the owner of the targeted group", func() {
					BeforeEach(func() {
						mock.ExpectBegin()
						mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "groups"`)).
							WithArgs(groupName).
							WillReturnRows(groupRow)
						mock.ExpectRollback()
					})

					It("propagates error", func() {
						err := uamDao.DeactivateGroup(uint(userID+2), groupName)
						Expect(err).To(HaveOccurred())
						_, ok := err.(*myerr.ClientError)
						Expect(ok).To(Equal(true))
						Expect(mock.ExpectationsWereMet()).To(BeNil())
					})
				})

				Context("and you are the owner of the targeted group", func() {

					Context("and request to revoke memberships fail", func() {
						BeforeEach(func() {
							mock.ExpectBegin()
							mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "groups"`)).
								WithArgs(groupName).
								WillReturnRows(groupRow)
							mock.ExpectExec("DELETE FROM \"memberships\"").
								WithArgs(groupID).
								WillReturnError(fmt.Errorf("some error"))
							mock.ExpectRollback()
						})
						It("propagates error", func() {
							err := uamDao.DeactivateGroup(uint(userID), groupName)
							Expect(err).To(HaveOccurred())
							_, ok := err.(*myerr.ServerError)
							Expect(ok).To(Equal(true))
							Expect(mock.ExpectationsWereMet()).To(BeNil())
						})
					})

					Context("and request to revoke memberships succeeds", func() {
						Context("and request to delete group fails", func() {
							BeforeEach(func() {
								mock.ExpectBegin()
								mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "groups"`)).
									WithArgs(groupName).
									WillReturnRows(groupRow)
								mock.ExpectExec("DELETE FROM \"memberships\"").
									WithArgs(groupID).
									WillReturnResult(sqlmock.NewResult(0, 1))
								mock.ExpectExec("UPDATE \"groups\"").
									WithArgs(false, Any{}, groupID).
									WillReturnError(fmt.Errorf("some error"))
								mock.ExpectRollback()
							})
							It("propagates error", func() {
								err := uamDao.DeactivateGroup(uint(userID), groupName)
								Expect(err).To(HaveOccurred())
								_, ok := err.(*myerr.ServerError)
								Expect(ok).To(Equal(true))
								Expect(mock.ExpectationsWereMet()).To(BeNil())
							})
						})

						Context("and request to delete group succeeds", func() {
							BeforeEach(func() {
								mock.ExpectBegin()
								mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "groups"`)).
									WithArgs(groupName).
									WillReturnRows(groupRow)
								mock.ExpectExec("DELETE FROM \"memberships\"").
									WithArgs(groupID).
									WillReturnResult(sqlmock.NewResult(0, 1))
								mock.ExpectExec("UPDATE \"groups\"").
									WithArgs(false, Any{}, groupID).
									WillReturnResult(sqlmock.NewResult(0, 1))
								mock.ExpectCommit()
							})
							It("propagates error", func() {
								err := uamDao.DeactivateGroup(uint(userID), groupName)
								Expect(err).NotTo(HaveOccurred())
							})
						})
					})

				})
			})

		})
	})

	Context("MemberExists", func() {
		When("request to check count of memberships with given user id and group name", func() {
			Context("and request fails", func() {
				Context("because of non-client problem", func() {
					BeforeEach(func() {
						mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) FROM "memberships"`)).
							WithArgs(uint(userID), uint(groupID)).
							WillReturnError(fmt.Errorf("some error"))
					})

					It("propagates error", func() {
						_, err := uamDao.MemberExists(uint(userID), uint(groupID))
						Expect(err).To(HaveOccurred())
						_, ok := err.(*myerr.ServerError)
						Expect(ok).To(Equal(true))
						Expect(mock.ExpectationsWereMet()).To(BeNil())
					})
				})
			})

			Context("and request succeeds", func() {
				Context("and memberships were not found", func() {
					BeforeEach(func() {
						mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) FROM "memberships"`)).
							WithArgs(uint(userID), uint(groupID)).
							WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
					})

					It("succeeds", func() {
						result, err := uamDao.MemberExists(uint(userID), uint(groupID))
						Expect(err).NotTo(HaveOccurred())
						Expect(result).To(BeFalse())
						Expect(mock.ExpectationsWereMet()).To(BeNil())
					})
				})

				Context("and membership was found", func() {
					BeforeEach(func() {
						mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(1) FROM "memberships"`)).
							WithArgs(uint(userID), uint(groupID)).
							WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
					})

					It("succeeds", func() {
						result, err := uamDao.MemberExists(uint(userID), uint(groupID))
						Expect(err).NotTo(HaveOccurred())
						Expect(result).To(BeTrue())
						Expect(mock.ExpectationsWereMet()).To(BeNil())
					})
				})

			})
		})
	})

	Context("GetGroup", func() {
		When("request to get group info is sent", func() {
			Context("and problem with the db occurrs", func() {
				BeforeEach(func() {
					mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "groups"`)).
						WithArgs(groupName).
						WillReturnError(fmt.Errorf("some error"))
				})

				It("propagates error", func() {
					_, err := uamDao.GetGroup(groupName)
					Expect(err).To(HaveOccurred())
					_, ok := err.(*myerr.ServerError)
					Expect(ok).To(Equal(true))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			Context("and query to db succeeds", func() {
				Context("and no group is found", func() {
					BeforeEach(func() {
						mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "groups"`)).
							WithArgs(groupName).
							WillReturnError(gorm.ErrRecordNotFound)
					})

					It("propagates error", func() {
						_, err := uamDao.GetGroup(groupName)
						Expect(err).To(HaveOccurred())
						_, ok := err.(*myerr.ItemNotFoundError)
						Expect(ok).To(Equal(true))
						Expect(mock.ExpectationsWereMet()).To(BeNil())
					})
				})

				Context("and the group was found", func() {
					BeforeEach(func() {
						mockTime := time.Now()
						rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "owner_id", "active"}).AddRow(1, mockTime, mockTime, groupName, userID, true)
						mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "groups"`)).
							WithArgs(groupName).
							WillReturnRows(rows)
					})

					It("succeds", func() {
						group, err := uamDao.GetGroup(groupName)
						Expect(err).ToNot(HaveOccurred())
						Expect(group.Name).To(Equal(groupName))
						Expect(group.OwnerID).To(Equal(uint(userID)))
						Expect(mock.ExpectationsWereMet()).To(BeNil())
					})
				})
			})
		})
	})

	Context("GetDeactivatedGroupNames", func() {
		When("request for all deactivated group names is sent", func() {
			Context("and the query to the db fails", func() {
				BeforeEach(func() {
					mock.ExpectQuery(regexp.QuoteMeta(`SELECT "name" FROM "groups"`)).
						WillReturnError(fmt.Errorf("some error"))
				})

				It("propagates error", func() {
					_, err := uamDao.GetDeactivatedGroupNames()
					Expect(err).To(HaveOccurred())
					_, ok := err.(*myerr.ServerError)
					Expect(ok).To(Equal(true))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			Context("and query to the db succeeds", func() {
				Context("and no deactivated groups are found", func() {
					BeforeEach(func() {
						mock.ExpectQuery(regexp.QuoteMeta(`SELECT "name" FROM "groups"`)).
							WillReturnRows(sqlmock.NewRows([]string{}))
					})

					It("propagates error", func() {
						groups, err := uamDao.GetDeactivatedGroupNames()
						Expect(err).ToNot(HaveOccurred())
						Expect(groups).To(BeEmpty())
						Expect(mock.ExpectationsWereMet()).To(BeNil())
					})
				})

				Context("and deactivated groups are found", func() {
					BeforeEach(func() {
						mock.ExpectQuery(regexp.QuoteMeta(`SELECT "name" FROM "groups"`)).
							WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow(groupName))
					})

					It("propagates error", func() {
						groups, err := uamDao.GetDeactivatedGroupNames()
						Expect(err).ToNot(HaveOccurred())
						Expect(len(groups)).To(Equal(1))
						Expect(groups[0]).To(Equal(groupName))
						Expect(mock.ExpectationsWereMet()).To(BeNil())
					})
				})
			})
		})
	})

	Context("EraseDeactivatedGroups", func() {
		When("request to delete all deactivated groups is sent", func() {
			Context("and deletion query fails", func() {
				BeforeEach(func() {
					mock.ExpectBegin()
					mock.ExpectExec("DELETE FROM \"groups\"").
						WithArgs(groupName).
						WillReturnError(fmt.Errorf("some error"))
					mock.ExpectRollback()
				})

				It("propagates error", func() {
					err := uamDao.EraseDeactivatedGroups([]string{groupName})
					Expect(err).To(HaveOccurred())
					_, ok := err.(*myerr.ServerError)
					Expect(ok).To(Equal(true))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			Context("and deletion query succeeds", func() {
				BeforeEach(func() {
					mock.ExpectBegin()
					mock.ExpectExec("DELETE FROM \"groups\"").
						WithArgs(groupName).
						WillReturnResult(sqlmock.NewResult(0, 1))
					mock.ExpectCommit()
				})

				It("propagates error", func() {
					err := uamDao.EraseDeactivatedGroups([]string{groupName})
					Expect(err).ToNot(HaveOccurred())
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})
		})
	})

	Context("GetAllGroups", func() {
		When("request to get all existing groups is sent", func() {
			Context("and query to db fails", func() {
				BeforeEach(func() {
					mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "groups"`)).
						WillReturnError(fmt.Errorf("some error"))
				})

				It("propagates error", func() {
					_, err := uamDao.GetAllGroups()
					Expect(err).To(HaveOccurred())
					_, ok := err.(*myerr.ServerError)
					Expect(ok).To(Equal(true))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			Context("and query to db succeeds", func() {
				Context("and no group is found", func() {
					BeforeEach(func() {
						mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "groups"`)).
							WillReturnRows(sqlmock.NewRows([]string{}))
					})

					It("propagates error", func() {
						groups, err := uamDao.GetAllGroups()
						Expect(err).ToNot(HaveOccurred())
						Expect(groups).To(BeEmpty())
						Expect(mock.ExpectationsWereMet()).To(BeNil())
					})
				})

				Context("and the group was found", func() {
					BeforeEach(func() {
						mockTime := time.Now()
						rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "owner_id", "active"}).AddRow(1, mockTime, mockTime, groupName, userID, true)
						mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "groups"`)).
							WillReturnRows(rows)
					})

					It("succeds", func() {
						groups, err := uamDao.GetAllGroups()
						Expect(err).ToNot(HaveOccurred())
						Expect(len(groups)).To(Equal(1))
						Expect(groups[0].Name).To(Equal(groupName))
						Expect(groups[0].OwnerID).To(Equal(uint(userID)))
						Expect(mock.ExpectationsWereMet()).To(BeNil())
					})
				})
			})
		})
	})

	Context("GetAllUsers", func() {
		When("request to get all existing users is sent", func() {
			Context("and query to db fails", func() {
				BeforeEach(func() {
					mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).
						WillReturnError(fmt.Errorf("some error"))
				})

				It("propagates error", func() {
					_, err := uamDao.GetAllUsers()
					Expect(err).To(HaveOccurred())
					_, ok := err.(*myerr.ServerError)
					Expect(ok).To(Equal(true))
					Expect(mock.ExpectationsWereMet()).To(BeNil())
				})
			})

			Context("and query to db succeeds", func() {
				Context("and no users is found", func() {
					BeforeEach(func() {
						mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).
							WillReturnRows(sqlmock.NewRows([]string{}))
					})

					It("propagates error", func() {
						groups, err := uamDao.GetAllUsers()
						Expect(err).ToNot(HaveOccurred())
						Expect(groups).To(BeEmpty())
						Expect(mock.ExpectationsWereMet()).To(BeNil())
					})
				})

				Context("and users were found", func() {
					BeforeEach(func() {
						mockTime := time.Now()
						rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "username", "password"}).AddRow(userID, mockTime, mockTime, username, password)
						mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users"`)).
							WillReturnRows(rows)
					})

					It("succeeds", func() {
						users, err := uamDao.GetAllUsers()
						Expect(err).ToNot(HaveOccurred())
						Expect(len(users)).To(Equal(1))
						Expect(users[0].ID).To(Equal(uint(userID)))
						Expect(users[0].Username).To(Equal(username))
						Expect(users[0].Password).To(Equal(password))
						Expect(mock.ExpectationsWereMet()).To(BeNil())
					})
				})
			})
		})
	})
})
