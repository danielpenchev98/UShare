package dao

import (
	"errors"
	"fmt"
	"log"

	"github.com/danielpenchev98/UShare/web-server/internal/db/models"
	myerr "github.com/danielpenchev98/UShare/web-server/internal/error"
	"gorm.io/gorm"
)

//go:generate mockgen --source=uam_dao.go --destination dao_mocks/uam_dao.go --package dao_mocks

//UamDAO - interface for working with the Database in regards to the User Access Management
type UamDAO interface {
	Migrate() error
	CreateUser(string, string) error
	GetUser(string) (models.User, error)
	DeleteUser(uint) error
	CreateGroup(uint, string) error
	AddUserToGroup(uint, string, string) error
	RemoveUserFromGroup(uint, string, string) error
	MemberExists(uint, uint) (bool, error)
	DeactivateGroup(uint, string) error
	GetGroup(string) (models.Group, error)
	GetDeactivatedGroupNames() ([]string, error)
	EraseDeactivatedGroups([]string) error
	GetAllGroups() ([]models.Group, error)
	GetAllUsers() ([]models.User, error)
	GetAllUsersInGroup(uint, string) ([]models.User, error)
}

//UamDAOImpl - implementation of UamDAO
type UamDAOImpl struct {
	dbConn *gorm.DB
}

//NewUamDAOImpl - function for creation an instance of UamDAOImpl
func NewUamDAOImpl(dbConn *gorm.DB) *UamDAOImpl {
	return &UamDAOImpl{dbConn: dbConn}
}

//Migrate - function which updates the models(table structure) in db
func (i *UamDAOImpl) Migrate() error {
	return i.dbConn.AutoMigrate(models.User{}, models.Group{}, models.Membership{})
}

//CreateUser - creates a new user in the database, given username and password (encrypted)
func (i *UamDAOImpl) CreateUser(username string, password string) error {
	return i.dbConn.Transaction(func(tx *gorm.DB) error {
		var count int64
		result := tx.Table("users").Where("username = ?", username).Count(&count)

		if result.Error != nil {
			return myerr.NewServerErrorWrap(result.Error, "Problem with the lookup of users")
		} else if count > 0 {
			return myerr.NewClientError("A user with the same username exists")
		}

		user := models.User{
			Username: username,
			Password: password,
		}

		log.Printf("Creating user with username [%s]", username)
		if result := tx.Create(&user); result.Error != nil {
			return myerr.NewServerErrorWrap(result.Error, "Problem with the creation of new user")
		}
		log.Printf("User with username [%s] created", username)

		return nil
	})
}

//DeleteUser - deletes user given an id of the user
func (i *UamDAOImpl) DeleteUser(userID uint) error {
	var count int64
	result := i.dbConn.Table("users").Where("id = ?", userID).Count(&count)

	if result.Error != nil {
		return myerr.NewServerErrorWrap(result.Error, "Problem with the lookup if user exists")
	} else if count == 0 {
		return myerr.NewItemNotFoundError("User with that id does not exist")
	}

	log.Printf("Deleting user with id [%d]\n", userID)
	if result = i.dbConn.Delete(&models.User{}, userID); result.Error != nil {
		return myerr.NewServerErrorWrap(result.Error, "Problem with the deletion of the user from db")
	}
	log.Printf("User with id [%d] is deleted\n", userID)

	return nil

}

//GetUser - fetches information about an existing user
func (i *UamDAOImpl) GetUser(username string) (models.User, error) {
	return getUserWithConn(i.dbConn, username)
}

//CreateGroup - creates a new group for sharing files
func (i *UamDAOImpl) CreateGroup(userID uint, groupName string) error {
	return i.dbConn.Transaction(func(tx *gorm.DB) error {
		var count int64

		result := tx.Table("groups").Where("name = ?", groupName).Count(&count)
		if result.Error != nil {
			return myerr.NewServerErrorWrap(result.Error, "Problem with the lookup of groups")
		} else if count > 0 {
			return myerr.NewClientError("A group with the same name exists")
		}

		group := models.Group{
			Name:    groupName,
			OwnerID: userID,
		}

		log.Printf("Creating group [%s] with owner [%d]\n", groupName, userID)
		if result := tx.Create(&group); result.Error != nil {
			return myerr.NewServerErrorWrap(result.Error, "Problem with the creation of group [%s] in db")
		}
		log.Printf("Group with name [%s] and owner [%d] created\n", groupName, userID)

		membership := models.Membership{
			UserID:  userID,
			GroupID: group.ID,
		}

		//its usedless to check if the membership already exists, because basically the group is created in this transaction
		log.Printf("Creating membership of user [%d] for group [%d]\n", userID, group.ID)
		if result := tx.Create(&membership); result.Error != nil {
			return myerr.NewServerErrorWrap(result.Error, "Problem with the creation of membership in db")
		}
		log.Printf("Membership of user [%d] for group [%d] is created\n", userID, group.ID)

		return nil
	})
}

//GetGroup - gets information about the group
func (i *UamDAOImpl) GetGroup(groupName string) (models.Group, error) {
	return getGroupWithConn(i.dbConn, groupName)
}

//AddUserToGroup - adds a new member to a specified group
func (i *UamDAOImpl) AddUserToGroup(ownerID uint, username string, groupName string) error {
	return i.dbConn.Transaction(func(tx *gorm.DB) error {
		var (
			count int64
			group models.Group
			user  models.User
			err   error
		)

		group, err = getGroupWithConn(tx, groupName)
		if err != nil {
			return err
		} else if group.OwnerID != ownerID {
			return myerr.NewClientError("Only the group owner can add members to the group")
		} else if !group.Active {
			return myerr.NewClientError("The group is currently being deleted")
		}

		user, err = getUserWithConn(tx, username)
		if err != nil {
			return err
		}

		result := tx.Table("memberships").
			Where("group_id = ?", group.ID).
			Where("user_id = ?", user.ID).
			Count(&count)

		if result.Error != nil {
			return myerr.NewServerErrorWrap(result.Error, "Problem with the lookup of membership in db")
		} else if count != 0 {
			return myerr.NewClientError("The user is already a member of the group")
		}

		membership := models.Membership{
			GroupID: group.ID,
			UserID:  user.ID,
		}

		log.Printf("Creating membership for user with id [%d] in group with id [%d]", membership.UserID, membership.GroupID)
		if result := tx.Create(&membership); result.Error != nil {
			return myerr.NewServerErrorWrap(result.Error, "Problem with the creation of new membership in db")
		}
		log.Printf("Membership for user with id [%d] in group id [%d] created", membership.UserID, membership.GroupID)

		return nil
	})
}

//DeactivateGroup - deletes all memberships and changes the status of the group to non active
func (i *UamDAOImpl) DeactivateGroup(currUserID uint, groupName string) error {
	return i.dbConn.Transaction(func(tx *gorm.DB) error {
		group, err := getGroupWithConn(tx, groupName)
		if err != nil {
			return err
		} else if group.OwnerID != currUserID {
			return myerr.NewClientError("Only the group owner can delete the group")
		} else if !group.Active {
			return myerr.NewClientError("The group is currently being deleted")
		}

		log.Printf("Revolking membership for users in group [%s]", groupName)
		result := tx.Table("memberships").
			Where("group_id = ?", group.ID).Delete(&models.Membership{})
		if result.Error != nil {
			return myerr.NewServerErrorWrap(result.Error, "Problem with deletion of memberships in db")
		}
		log.Printf("Revolked membership for users in group [%s]", groupName)

		log.Printf("Change status of group [%s] to non active\n", groupName)
		if result = tx.Model(&group).Update("active", false); result.Error != nil {
			return myerr.NewServerErrorWrap(result.Error, "Problem with deletion of the group in db")
		}
		log.Printf("Status of group [%s] is set to non active\n", groupName)
		return nil
	})
}

//RemoveUserFromGroup - removes a membership of a user to a specific group
func (i *UamDAOImpl) RemoveUserFromGroup(currUserID uint, username string, groupName string) error {
	return i.dbConn.Transaction(func(tx *gorm.DB) error {
		var (
			group models.Group
			user  models.User
			err   error
		)

		group, err = getGroupWithConn(tx, groupName)
		if err != nil {
			return err
		} else if !group.Active {
			return myerr.NewClientError("The group is currently being deleted")
		}

		user, err = getUserWithConn(tx, username)
		if err != nil {
			return err
		}

		if group.OwnerID != currUserID && user.ID != currUserID {
			return myerr.NewClientError("Only the owner of the group can revoke membership of other members")
		} else if group.OwnerID == currUserID && user.ID == currUserID {
			return myerr.NewClientError("The owner cannot remove its own membership. Yet to be added this functionality")
		}

		log.Printf("Revolking membership for user with id [%d] in group with id [%d]", user.ID, group.ID)
		result := tx.Where("user_id = ?", user.ID).
			Where("group_id = ?", group.ID).
			Delete(&models.Membership{})

		if result.Error != nil {
			return myerr.NewServerErrorWrap(result.Error, "Problem with the creation of new membership in db")
		} else if result.RowsAffected == 0 {
			return myerr.NewClientError("Membership not found")
		}
		log.Printf("Membership for user with id [%d] in group id [%d] is revoked", user.ID, group.ID)

		return nil
	})
}

//MemberExists - check if membership exists for a particular group
func (i *UamDAOImpl) MemberExists(userID uint, groupID uint) (bool, error) {
	var count int64
	result := i.dbConn.Table("memberships").
		Where("user_id = ?", userID).
		Where("group_id = ?", groupID).
		Count(&count)

	if result.Error != nil {
		return false, myerr.NewServerErrorWrap(result.Error, "Problem with check existance of membership")
	}
	return count != 0, nil
}

//GetDeactivatedGroupNames - retrieves names of all deactivated groups, which are still not deleted
func (i *UamDAOImpl) GetDeactivatedGroupNames() ([]string, error) {
	var groupNames []string
	result := i.dbConn.Table("groups").
		Where("active", false).Pluck("name", &groupNames)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return make([]string, 0), nil
	} else if result.Error != nil {
		return nil, myerr.NewServerErrorWrap(result.Error, "Problem with finding all groups, whose resources shuld be deleted")
	}
	return groupNames, nil
}

//EraseDeactivatedGroups - deletes pernamently the groups, instead of just deactivate them
func (i *UamDAOImpl) EraseDeactivatedGroups(groupNames []string) error {
	return i.dbConn.Transaction(func(tx *gorm.DB) error {
		for _, groupName := range groupNames {
			result := tx.Unscoped().Where("name = ?", groupName).Delete(&models.Group{})
			if result.Error != nil {
				return myerr.NewServerErrorWrap(result.Error, "Couldnt delete the inactive groups")
			} else if result.RowsAffected == 0 {
				fmt.Println("Warning. Tried to delete already deleted group")
			}
		}
		return nil
	})
}

//GetAllGroups - retrieves all groups
func (i *UamDAOImpl) GetAllGroups() ([]models.Group, error) {
	var groups []models.Group
	result := i.dbConn.Find(&groups)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return make([]models.Group, 0), nil
	} else if result.Error != nil {
		return nil, myerr.NewServerErrorWrap(result.Error, "Problem with fetching all groups")
	}
	return groups, nil
}

//GetAllUsers - retrieves all users
func (i *UamDAOImpl) GetAllUsers() ([]models.User, error) {
	var users []models.User
	result := i.dbConn.Find(&users)
	if result.Error != nil {
		return nil, myerr.NewServerErrorWrap(result.Error, "Problem with fetching all users")
	}
	return users, nil
}

//GetAllUsersInGroup - retrieves all users in a group
func (i *UamDAOImpl) GetAllUsersInGroup(userID uint, groupName string) ([]models.User, error) {
	var users []models.User
	err := i.dbConn.Transaction(func(tx *gorm.DB) error {
		group, errGet := getGroupWithConn(tx, groupName)
		if _, ok := errGet.(*myerr.ItemNotFoundError); ok || !group.Active {
			return myerr.NewClientError("Invalid group")
		} else if errGet != nil {
			return errGet
		}

		var count int64
		result := tx.Table("memberships").
			Where("group_id = ?", group.ID).
			Where("user_id = ?", userID).
			Count(&count)

		if result.Error != nil {
			return myerr.NewServerErrorWrap(result.Error, "Problem with the lookup of membership in db")
		} else if count == 0 {
			return myerr.NewClientError("The user is not a member of the group")
		}

		log.Printf("Group id %d\n", group.ID)

		result = tx.Table("users").Joins("inner join memberships on users.id = memberships.user_id").
			Where("memberships.group_id = ?", group.ID).Find(&users)

		if result.Error != nil {
			return myerr.NewServerErrorWrap(result.Error, "Problem with the lookup of users in db")
		}

		return nil
	})
	return users, err
}

func getUserWithConn(dbConn *gorm.DB, username string) (models.User, error) {
	var user models.User

	result := dbConn.Table("users").
		Where("username = ?", username).
		Find(&user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return user, myerr.NewItemNotFoundError("User does not exist")
	} else if result.Error != nil {
		return user, myerr.NewServerErrorWrap(result.Error, "Problem with the lookup if user exists")
	}

	return user, nil
}

func getGroupWithConn(dbConn *gorm.DB, groupName string) (models.Group, error) {
	var group models.Group

	result := dbConn.Table("groups").
		Where("name = ?", groupName).
		Find(&group)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return group, myerr.NewItemNotFoundError(fmt.Sprintf("Group [%s] does not exist", groupName))
	} else if result.Error != nil {
		return group, myerr.NewServerErrorWrap(result.Error, "Problem with the lookup if group exists")
	}

	return group, nil
}
