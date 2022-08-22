package rest

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/danielpenchev98/UShare/web-server/api/common"
	"github.com/danielpenchev98/UShare/web-server/internal/auth"
	"github.com/danielpenchev98/UShare/web-server/internal/db/dao"
	myerr "github.com/danielpenchev98/UShare/web-server/internal/error"
	val "github.com/danielpenchev98/UShare/web-server/internal/validator"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

//UamEndpoint - rest endpoint for configuration of the user access management
type UamEndpoint interface {
	CreateUser(*gin.Context)
	DeleteUser(*gin.Context)
	Login(*gin.Context)

	CreateGroup(*gin.Context)
	AddMember(*gin.Context)
	RevokeMembership(*gin.Context)
	DeleteGroup(*gin.Context)
}

//UamEndpointImpl - implementation of UamEndpoint
type UamEndpointImpl struct {
	uamDAO     dao.UamDAO
	jwtCreator auth.JwtCreator
	validator  val.Validator
	groupsDir  string
}

//NewUamEndPointImpl - function for creation an instance of UamEndpointImpl
func NewUamEndPointImpl(uamDAO dao.UamDAO, creator auth.JwtCreator, validator val.Validator, groupsDir string) *UamEndpointImpl {
	return &UamEndpointImpl{
		uamDAO:     uamDAO,
		jwtCreator: creator,
		validator:  validator,
		groupsDir:  groupsDir,
	}
}

//CreateUser - handler for user creation request
//returns 500, if error occurrs due to system failure
//returns 400 if the user input was invalid
//returns 201 if the user was successfully created
func (i *UamEndpointImpl) CreateUser(c *gin.Context) {
	var rq common.RequestWithCredentials

	//Decide what exactly to return as response -> custom message + 400 or?
	if err := c.ShouldBindJSON(&rq); err != nil {
		log.Println("Cannot unmarshall this shit")
		common.SendErrorResponse(c, myerr.NewClientError("Invalid json body"))
		return
	}

	if err := validateRegistration(i.validator, rq); err != nil {
		log.Println(err)
		common.SendErrorResponse(c, err)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(rq.Password), bcrypt.DefaultCost)
	if err != nil {
		err = myerr.NewServerErrorWrap(err, "Problem encryption of password during the registration.")
		common.SendErrorResponse(c, err)
		return
	}

	log.Printf("Creating the user with username %s and password %s\n", rq.Username, rq.Password)

	err = i.uamDAO.CreateUser(rq.Username, string(hashedPassword))
	if _, ok := err.(*myerr.ClientError); ok {
		common.SendErrorResponse(c, err)
		return
	} else if err != nil {
		err = myerr.NewServerErrorWrap(err, "Problem crearing the user in the db.")
		common.SendErrorResponse(c, err)
		return
	}

	c.JSON(http.StatusCreated, common.BasicResponse{
		Status: http.StatusCreated,
	})
}

//DeleteUser - handler for user deletion request
//returns 500, if error occurrs due to system failure
//returns 400 if the user input was invalid
//returns 200 if the user was successfully deleted
func (i *UamEndpointImpl) DeleteUser(c *gin.Context) {
	userID, err := common.GetIDFromContext(c)
	if err != nil {
		common.SendErrorResponse(c, err)
	}

	if err = i.uamDAO.DeleteUser(uint(userID)); err != nil {
		err = myerr.NewServerErrorWrap(err, "Problem with deletion of user.")
		common.SendErrorResponse(c, myerr.NewServerErrorWrap(err, "Problem with deleting user"))
		return
	}

	c.JSON(http.StatusOK, common.BasicResponse{
		Status: http.StatusOK,
	})
}

//Login - handler for user login request
//returns 500, if error occurrs due to system failure
//returns 400 if the user input was invalid
//returns 201 if the login was successfull
func (i *UamEndpointImpl) Login(c *gin.Context) {
	var request common.RequestWithCredentials
	if err := c.ShouldBindJSON(&request); err != nil {
		common.SendErrorResponse(c, myerr.NewClientError("Invalid json body"))
		return
	}

	user, err := i.uamDAO.GetUser(request.Username)
	if err != nil {
		if _, ok := err.(*myerr.ItemNotFoundError); ok {
			common.SendErrorResponse(c, myerr.NewClientError("Invalid credentials"))
		} else {
			err = myerr.NewServerErrorWrap(err, "Problem with Login.")
			common.SendErrorResponse(c, err)
		}
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password))
	if err != nil {
		common.SendErrorResponse(c, myerr.NewClientError("Invalid credentials"))
		return
	}

	signedToken, err := i.jwtCreator.GenerateToken(user.ID)
	if err != nil {
		err = myerr.NewServerErrorWrap(err, "Problem with generating Jwt token in the login logic.")
		common.SendErrorResponse(c, err)
		return
	}

	c.JSON(http.StatusCreated, common.LoginResponse{
		Status: http.StatusCreated,
		Token:  signedToken,
	})
}

//CreateGroup - handler for group creation request
//returns 500, if error occurrs due to system failure
//returns 400 if the user input was invalid
//returns 201 if the group was successfully created
func (i *UamEndpointImpl) CreateGroup(c *gin.Context) {
	userID, err := common.GetIDFromContext(c)
	if err != nil {
		common.SendErrorResponse(c, err)
	}

	var rq common.GroupPayload
	if err := c.ShouldBindJSON(&rq); err != nil {
		common.SendErrorResponse(c, myerr.NewClientError("Invalid json body"))
		return
	}

	//TODO should rename this function - maybe? or create specific function for the group name
	if err = i.validator.ValidateUsername(rq.GroupName); err != nil {
		err = myerr.NewClientErrorWrap(err, "Problem with the group name")
		common.SendErrorResponse(c, err)
		return
	}

	groupDir := path.Join(i.groupsDir, rq.GroupName)
	fmt.Println(groupDir)
	if _, err := os.Stat(groupDir); !os.IsNotExist(err) {
		common.SendErrorResponse(c, myerr.NewClientError("Problem with creation of group. Reason: Group already exists"))
		return
	} else if err = os.Mkdir(groupDir, 0755); err != nil {
		common.SendErrorResponse(c, myerr.NewServerErrorWrap(err, "Problem with creation of directory"))
		return
	}

	err = i.uamDAO.CreateGroup(userID, rq.GroupName)
	if _, ok := err.(*myerr.ClientError); ok {
		common.SendErrorResponse(c, err)
		return
	} else if err != nil {
		os.RemoveAll(groupDir)
		common.SendErrorResponse(c, myerr.NewServerErrorWrap(err, "Problem with creation of group."))
		return
	}

	c.JSON(http.StatusCreated, common.BasicResponse{
		Status: http.StatusCreated,
	})
}

//AddMember - handler for membership creation request
//returns 500, if error occurrs due to system failure
//returns 400 if the user input was invalid
//returns 201 if the user was successfully added to the group
func (i *UamEndpointImpl) AddMember(c *gin.Context) {
	userID, err := common.GetIDFromContext(c)
	if err != nil {
		common.SendErrorResponse(c, err)
	}

	var rq common.GroupMembershipPayload
	if err := c.ShouldBindJSON(&rq); err != nil {
		common.SendErrorResponse(c, myerr.NewClientError("Invalid json body"))
		return
	}

	err = i.uamDAO.AddUserToGroup(userID, rq.Username, rq.GroupName)
	if _, ok := err.(*myerr.ClientError); ok {
		common.SendErrorResponse(c, err)
		return
	} else if err != nil {
		err = myerr.NewServerErrorWrap(err, "Problem with creation of group.")
		common.SendErrorResponse(c, err)
		return
	}

	c.JSON(http.StatusCreated, common.BasicResponse{
		Status: http.StatusCreated,
	})
}

//RevokeMembership - handler for membership deletion request
//returns 500, if error occurrs due to system failure
//returns 400 if the user input was invalid
//returns 200 if the user was successfully removed from the group
func (i *UamEndpointImpl) RevokeMembership(c *gin.Context) {
	userID, err := common.GetIDFromContext(c)
	if err != nil {
		common.SendErrorResponse(c, err)
		return
	}

	var rq common.GroupMembershipPayload
	if err := c.ShouldBindJSON(&rq); err != nil {
		common.SendErrorResponse(c, myerr.NewClientError("Invalid json body"))
		return
	}

	err = i.uamDAO.RemoveUserFromGroup(userID, rq.Username, rq.GroupName)

	if err != nil {
		if _, ok := err.(*myerr.ServerError); ok {
			err = myerr.NewServerErrorWrap(err, "Couldnt remove membership")
		}
		common.SendErrorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, common.BasicResponse{
		Status: http.StatusOK,
	})
}

//DeleteGroup - handler for group deletion request
//returns 500, if error occurrs due to system failure
//returns 400 if the user input was invalid
//returns 20- if the group was successfully deleted
func (i *UamEndpointImpl) DeleteGroup(c *gin.Context) {
	userID, err := common.GetIDFromContext(c)
	if err != nil {
		common.SendErrorResponse(c, err)
		return
	}

	var rq common.GroupPayload
	if err = c.ShouldBindJSON(&rq); err != nil {
		common.SendErrorResponse(c, myerr.NewClientError("Invalid json body"))
		return
	}

	err = i.uamDAO.DeactivateGroup(userID, rq.GroupName)
	if _, ok := err.(*myerr.ClientError); ok {
		common.SendErrorResponse(c, err)
		return
	} else if err != nil {
		err = myerr.NewServerErrorWrap(err, "Problem with deletion of group.")
		common.SendErrorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, common.BasicResponse{
		Status: http.StatusOK,
	})
}

//GetAllGroupsInfo - handler for fetching info about every active group
//returns 500, if error occurrs due to system failure
//returns 400 if the user input was invalid
//returns 200 otherwise
func (i *UamEndpointImpl) GetAllGroupsInfo(c *gin.Context) {
	groups, err := i.uamDAO.GetAllGroups()
	if _, ok := err.(*myerr.ServerError); ok {
		err = myerr.NewServerErrorWrap(err, "Problem with fetching all groups.")
		common.SendErrorResponse(c, err)
		return
	}

	groupsInfo := make([]common.GroupInfo, 0, len(groups))
	for _, group := range groups {
		groupsInfo = append(groupsInfo, common.GroupInfo{
			ID:      group.ID,
			Name:    group.Name,
			OwnerID: group.OwnerID,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"groups": groupsInfo,
	})
}

//GetAllUsersInfo - handler for fetching info about every user
//returns 500, if error occurrs due to system failure
//returns 400 if the user input was invalid
//returns 200 otherwise
func (i *UamEndpointImpl) GetAllUsersInfo(c *gin.Context) {
	users, err := i.uamDAO.GetAllUsers()
	if _, ok := err.(*myerr.ServerError); ok {
		err = myerr.NewServerErrorWrap(err, "Problem with fetching all users.")
		common.SendErrorResponse(c, err)
		return
	}

	usersInfo := make([]common.UserInfo, 0, len(users))
	for _, user := range users {
		usersInfo = append(usersInfo, common.UserInfo{
			ID:       user.ID,
			Username: user.Username,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"users":  usersInfo,
	})
}

//GetAllUsersInGroup - handler for fetching info about every user in a specific group
//returns 500, if error occurrs due to system failure
//returns 400 if the user input was invalid
//returns 200 otherwise
func (i *UamEndpointImpl) GetAllUsersInGroup(c *gin.Context) {
	userID, err := common.GetIDFromContext(c)
	if err != nil {
		common.SendErrorResponse(c, err)
		return
	}

	groupName := c.Query("group_name")
	if groupName == "" {
		common.SendErrorResponse(c, myerr.NewClientError("Groupname isnt specified"))
		return
	}

	users, err := i.uamDAO.GetAllUsersInGroup(userID, groupName)
	if _, ok := err.(*myerr.ClientError); ok {
		err = myerr.NewClientErrorWrap(err, "Cannot retrieve the group users")
		common.SendErrorResponse(c, err)
		return
	} else if err != nil {
		err = myerr.NewServerErrorWrap(err, "Problem with fetching all users in particular group.")
		common.SendErrorResponse(c, err)
		return
	}

	usersInfo := make([]common.UserInfo, 0, len(users))
	for _, user := range users {
		usersInfo = append(usersInfo, common.UserInfo{
			ID:       user.ID,
			Username: user.Username,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"users":  usersInfo,
	})
}

func validateRegistration(validator val.Validator, rq common.RequestWithCredentials) error {
	if err := validator.ValidateUsername(rq.Username); err != nil {
		return myerr.NewClientErrorWrap(err, "Problem with the username")
	}

	if err := validator.ValidatePassword(rq.Password); err != nil {
		return myerr.NewClientErrorWrap(err, "Problem with the password")
	}
	return nil
}
