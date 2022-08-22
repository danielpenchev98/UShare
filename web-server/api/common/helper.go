package common

import (
	"fmt"
	"log"
	"net/http"

	myerr "github.com/danielpenchev98/UShare/web-server/internal/error"
	"github.com/gin-gonic/gin"
)

//GetIDFromContext - extracts id from the context
func GetIDFromContext(c *gin.Context) (uint, error) {
	id, ok := c.Get("userID")
	if !ok {
		log.Println("Problem retieval of userID from context.")
		return 0, myerr.NewServerError("Cannot retrieve the user id")
	}

	var userID uint
	if userID, ok = id.(uint); !ok {
		return 0, myerr.NewClientError("Invalid user ID")
	}
	return userID, nil
}

//SendErrorResponse - generic method for sending error response to the user
func SendErrorResponse(c *gin.Context, err error) {
	errorCode, errorMsg := getErrorResponseArguments(err)
	c.JSON(errorCode, ErrorResponse{
		ErrorCode: errorCode,
		ErrorMsg:  errorMsg,
	})
}

func getErrorResponseArguments(err error) (errorCode int, errorMsg string) {
	switch err.(type) {
	case *myerr.ClientError:
		errorCode = http.StatusBadRequest
		errorMsg = fmt.Sprintf("Invalid request. Reason :%s", err.Error())
	case *myerr.ItemNotFoundError:
		errorCode = http.StatusNotFound
		errorMsg = err.Error()
	default:
		log.Println(err)
		errorCode = http.StatusInternalServerError
		errorMsg = fmt.Sprintf("Problem with the server, please try again later")
	}
	return
}
