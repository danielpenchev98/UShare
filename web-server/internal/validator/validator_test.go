package validator_test

import (
	. "github.com/danielpenchev98/UShare/web-server/internal/validator"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Validator", func() {

	var validator Validator

	BeforeEach(func() {
		validator = NewBasicValidator()
	})

	Describe("username validation", func() {
		When("username is invalid", func() {
			Context("username less than 7 symbols", func() {
				It("returns error", func() {
					err := validator.ValidateUsername("example")
					Expect(err).To(HaveOccurred())
				})
			})

			Context("username greather than 20 symbols", func() {
				It("returns error", func() {
					err := validator.ValidateUsername("somerandomlygeneratedusername")
					Expect(err).To(HaveOccurred())
				})
			})

			Context("username begins with non char symbol", func() {
				It("returns error", func() {
					err := validator.ValidateUsername("9someusername")
					Expect(err).To(HaveOccurred())
				})
			})

			Context("username should contains a special character(except _ and -)", func() {
				It("returns error", func() {
					err := validator.ValidateUsername("nonumberusername!")
					Expect(err).To(HaveOccurred())
				})
			})
		})
		When("username is valid", func() {
			It("succeeds", func() {
				err := validator.ValidateUsername("valid-username")
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
	Describe("password validation", func() {
		When("password is invalid", func() {
			Context("password less than 10 symbols", func() {
				It("returns error", func() {
					err := validator.ValidatePassword("1random~")
					Expect(err).To(HaveOccurred())
				})
			})

			Context("password doesnt contain special symbol", func() {
				It("returns error", func() {
					err := validator.ValidatePassword("1someusername")
					Expect(err).To(HaveOccurred())
				})
			})

			Context("password doesnt contain digit symbol", func() {
				It("returns error", func() {
					err := validator.ValidatePassword("password~")
					Expect(err).To(HaveOccurred())
				})
			})
		})
		When("password is valid", func() {
			When("password is valid", func() {
				It("succeeds", func() {
					err := validator.ValidatePassword("1validpassword~")
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})
})
