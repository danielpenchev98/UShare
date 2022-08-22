package validator

import (
	"regexp"

	myerr "github.com/danielpenchev98/UShare/web-server/internal/error"
)

//go:generate mockgen --source=validator.go --destination validator_mocks/validator.go --package validator_mocks

//Validator - interface, declaring all needed validation methods
type Validator interface {
	ValidateUsername(username string) error
	ValidatePassword(password string) error
}

//BasicValidator is implementation of Validator interface with basic functionality
type BasicValidator struct {
	usernameRules []rule
	passwordRules []rule
}

type rule struct {
	regex    string
	errorMsg string
}

//NewBasicValidator returns an implementaion of the Validator interface
func NewBasicValidator() *BasicValidator {
	return &BasicValidator{
		usernameRules: getBasicUsernameRules(),
		passwordRules: getBasicPasswordRules(),
	}
}

//ValidateUsername validates username
//returns error if the validation fails
func (v *BasicValidator) ValidateUsername(username string) error {
	return checkRules(v.usernameRules, username)
}

//ValidatePassword validates passwords
//returns error if the validation fails
func (v *BasicValidator) ValidatePassword(password string) error {
	return checkRules(v.passwordRules, password)
}

func checkRules(rules []rule, target string) error {
	for _, rule := range rules {
		matched, _ := regexp.Match(rule.regex, []byte(target))
		if !matched {
			return myerr.NewClientError(rule.errorMsg)
		}
	}
	return nil
}

func getBasicUsernameRules() []rule {
	return []rule{
		rule{regex: "^.{8,20}$", errorMsg: "Username should be between 8 and 20 symbols"},
		rule{regex: "^[a-zA-Z].*", errorMsg: "Username should always begin only with a letter"},
		rule{regex: "^[-_0-9a-zA-Z]+$", errorMsg: "Username cannot contain special symbols except \"-\" and \"_\""},
	}
}

func getBasicPasswordRules() []rule {
	return []rule{
		rule{regex: "^.{10,}$", errorMsg: "Password should be greater than 9 symbols"},
		rule{regex: ".*[0-9].*", errorMsg: "Password should contain atleast one number"},
		rule{regex: ".*[^-_0-9a-zA-Z].*", errorMsg: "Password should contain atleast one special char"},
	}
}
