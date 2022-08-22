package rest

import (
	"net/http"
	"github.com/gin-gonic/gin"
)


//CheckHealth is used for checking the health of the server. Returns 200 if the server is responsive
func CheckHealth(c *gin.Context) {
	c.Status(http.StatusOK)
}