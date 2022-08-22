package rest

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/danielpenchev98/UShare/web-server/api/common"
	"github.com/danielpenchev98/UShare/web-server/internal/db/dao"
	"github.com/danielpenchev98/UShare/web-server/internal/db/models"
	myerr "github.com/danielpenchev98/UShare/web-server/internal/error"
	"github.com/gin-gonic/gin"
)

//FileManagementEndpoint - used as interface of rest endpoint for the management of files
type FileManagementEndpoint interface {
	UploadFile(*gin.Context)
	DownloadFile(*gin.Context)
	DeleteFile(*gin.Context)
	RetrieveAllFilesInfo(c *gin.Context)
}

//FileManagementEndpointImpl - implementation of FileManagementEndpoint interface
type FileManagementEndpointImpl struct {
	UamDAO    dao.UamDAO
	groupsDir string
	FmDAO     dao.FmDAO
}

//NewFileManagementEndpointImpl - instance creation of FileManagementEndpointImpl
func NewFileManagementEndpointImpl(uam dao.UamDAO, fm dao.FmDAO, groupsDir string) *FileManagementEndpointImpl {
	return &FileManagementEndpointImpl{
		UamDAO:    uam,
		FmDAO:     fm,
		groupsDir: groupsDir,
	}
}

//UploadFile - handler for the upload of files from a user of specific group
//returns 500, if there is a problem with the server
//returns 400, if the user input is invalid
//returns 201, if the file is uploaded
func (i *FileManagementEndpointImpl) UploadFile(c *gin.Context) {
	var (
		userID uint
		err    error
	)

	if userID, err = common.GetIDFromContext(c); err != nil {
		common.SendErrorResponse(c, err)
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		common.SendErrorResponse(c, myerr.NewClientError("Problem with the file"))
		return
	}

	var group models.Group
	groupName := c.Query("group_name")
	if groupName == "" {
		common.SendErrorResponse(c, myerr.NewClientError("Groupname isnt specified"))
		return
	}

	group, err = i.UamDAO.GetGroup(groupName)
	switch err.(type) {
	case nil:
		break
	case (*myerr.ItemNotFoundError):
		common.SendErrorResponse(c, myerr.NewClientError("Invalid group"))
		return
	default:
		common.SendErrorResponse(c, err)
		return
	}

	if exists, err := i.UamDAO.MemberExists(userID, group.ID); err != nil {
		common.SendErrorResponse(c, err)
		return
	} else if exists != true {
		common.SendErrorResponse(c, myerr.NewClientError("Invalid user input"))
		return
	}

	fileID, err := i.FmDAO.AddFileInfo(userID, file.Filename, groupName)
	if err != nil {
		common.SendErrorResponse(c, err)
		return
	}

	dst := fmt.Sprintf("%s/%s/%d", i.groupsDir, groupName, fileID)
	if err = c.SaveUploadedFile(file, dst); err != nil {
		i.FmDAO.RemoveFileInfo(userID, fileID, groupName)
		common.SendErrorResponse(c, myerr.NewServerError(fmt.Sprintf("Couldnt save the file in the group dir [%s]", groupName)))
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  http.StatusCreated,
		"file_id": fileID,
	})
}

//DownloadFile - downloads a file given group
//returns 500, if an error occurs due to system failure
//returns 400 - if the user doesnt have enough permissions
//returns 200 + the downloaded file if the users has the permissions
func (i *FileManagementEndpointImpl) DownloadFile(c *gin.Context) {
	var (
		userID uint
		err    error
	)

	if userID, err = common.GetIDFromContext(c); err != nil {
		common.SendErrorResponse(c, err)
		return
	}

	fileIDString := c.Query("file_id")
	if fileIDString == "" {
		common.SendErrorResponse(c, myerr.NewClientError("File id isnt specified"))
		return
	}
	fileID, err := strconv.ParseUint(fileIDString, 0, 32)
	if err != nil {
		common.SendErrorResponse(c, myerr.NewClientError("Unvalid format of file id"))
		return
	}
	groupName := c.Query("group_name")
	if groupName == "" {
		common.SendErrorResponse(c, myerr.NewClientError("Groupname isnt specified"))
		return
	}

	group, err := i.UamDAO.GetGroup(groupName)
	if exists, err := i.UamDAO.MemberExists(userID, group.ID); err != nil {
		common.SendErrorResponse(c, err)
		return
	} else if exists != true {
		common.SendErrorResponse(c, myerr.NewClientError("Invalid user input"))
		return
	}

	fileInfo, err := i.FmDAO.GetFileInfo(userID, uint(fileID), groupName)
	if err != nil {
		common.SendErrorResponse(c, err)
	}

	c.Writer.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileInfo.Name))
	filePath := fmt.Sprintf("%s/%s/%d", i.groupsDir, groupName, fileInfo.ID)
	c.File(filePath)
}

//DeleteFile - deletes a file from the system
//returns 500, if an error occurs due to system failure
//returns 400, if the user doesnt have enough permissions
//returns 200, if the file is succesfully deleted
func (i *FileManagementEndpointImpl) DeleteFile(c *gin.Context) {
	var (
		userID uint
		err    error
	)

	if userID, err = common.GetIDFromContext(c); err != nil {
		common.SendErrorResponse(c, err)
		return
	}

	var rq common.FileRequestPayload
	if err := c.ShouldBindJSON(&rq); err != nil {
		common.SendErrorResponse(c, myerr.NewClientError("Invalid json body"))
		return
	}

	group, err := i.UamDAO.GetGroup(rq.GroupName)
	if exists, err := i.UamDAO.MemberExists(userID, group.ID); err != nil {
		common.SendErrorResponse(c, err)
		return
	} else if exists != true {
		common.SendErrorResponse(c, myerr.NewClientError("Invalid user input"))
		return
	}

	if err := i.FmDAO.RemoveFileInfo(userID, rq.FileID, rq.GroupName); err != nil {
		common.SendErrorResponse(c, err)
		return
	}

	path := fmt.Sprintf("%s/%s/%d", i.groupsDir, rq.GroupName, rq.FileID)
	os.Remove(path)

	c.JSON(http.StatusOK, common.BasicResponse{
		Status: http.StatusOK,
	})
}

//RetrieveAllFilesInfo - retrieves info about all files owned by a particular group
//returns 500, if error occurrs due to system failure
//returns 400, if the user doesnt have enough permissions
//returns 200 + info about files
func (i *FileManagementEndpointImpl) RetrieveAllFilesInfo(c *gin.Context) {
	var (
		userID uint
		err    error
	)

	if userID, err = common.GetIDFromContext(c); err != nil {
		common.SendErrorResponse(c, err)
		return
	}

	groupName := c.Query("group_name")
	if groupName == "" {
		common.SendErrorResponse(c, myerr.NewClientError("Groupname isnt specified"))
		return
	}

	fileInfos, err := i.FmDAO.GetAllFilesInfo(userID, groupName)
	if _, ok := err.(*myerr.ClientError); ok {
		common.SendErrorResponse(c, myerr.NewClientErrorWrap(err, "Problem with file retrieval"))
		return
	} else if err != nil {
		common.SendErrorResponse(c, err)
		return
	}

	fileResponses := make([]common.FileInfoResponse, 0, len(fileInfos))
	for _, fileInfo := range fileInfos {
		fileResponses = append(fileResponses, common.FileInfoResponse{
			ID:         fileInfo.ID,
			Name:       fileInfo.Name,
			UploadedAt: fileInfo.CreatedAt,
			OwnerID:    fileInfo.OwnerID,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status": http.StatusOK,
		"files":  fileResponses,
	})
}
