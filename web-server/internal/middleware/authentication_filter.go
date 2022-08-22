package middleware

import (
	"net/http"
	"strings"

	"github.com/danielpenchev98/UShare/web-server/api/common"
	"github.com/danielpenchev98/UShare/web-server/internal/auth"
	"github.com/gin-gonic/gin"
)

//AuthzFilter - middleware for filtering unauthorized requests
type AuthzFilter interface {
	Authz(c *gin.Context)
}

//AuthzFilterImpl - implementation of AuthorizationFilter
type AuthzFilterImpl struct {
	jwtCreator auth.JwtCreator
}

//NewAuthzFilterImpl - creates a new instance of AuthzFilterImpl
func NewAuthzFilterImpl(creator auth.JwtCreator) *AuthzFilterImpl {
	return &AuthzFilterImpl{
		jwtCreator: creator,
	}
}

//Authz - creating handlers for filtering unauthorized requests
func (f *AuthzFilterImpl) Authz(c *gin.Context) {
	clientToken := c.Request.Header.Get("Authorization")
	if clientToken == "" {
		c.JSON(http.StatusForbidden, common.ErrorResponse{
			ErrorCode: http.StatusForbidden,
			ErrorMsg:  "No Authorization header provided",
		})
		c.Abort() //stop the propagation of the request to the next handler
		return
	}

	extractedToken := strings.Split(clientToken, "Bearer ")
	if len(extractedToken) == 2 {
		clientToken = strings.TrimSpace(extractedToken[1])
	} else {
		c.JSON(http.StatusBadRequest, common.ErrorResponse{
			ErrorCode: http.StatusBadRequest,
			ErrorMsg:  "Incorrect Format of Authorization Token",
		})
		c.Abort()
		return
	}

	claims, err := f.jwtCreator.ValidateToken(clientToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, common.ErrorResponse{
			ErrorCode: http.StatusUnauthorized,
			ErrorMsg:  "Invalid Authorization token",
		})
		c.Abort()
		return
	}

	c.Set("userID", claims.UserID)
	c.Next()
}
