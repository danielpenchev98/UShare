package auth

import (
	"os"
	"strconv"
	"time"

	myerr "github.com/danielpenchev98/UShare/web-server/internal/error"
	jwt "github.com/dgrijalva/jwt-go"
)

const (
	secretKey     = "SECRET"
	issuerKey     = "ISSUER"
	expirationKey = "EXPIRATION"
)

//go:generate mockgen --source=auth.go --destination auth_mocks/auth.go --package auth_mocks

//JwtCreator - a wrapper of jwt library
type JwtCreator interface {
	GenerateToken(uint) (string, error)
	ValidateToken(string) (*JwtClaim, error)
}

//JwtCreatorImpl - implementation of JwtCreator
type JwtCreatorImpl struct {
	Secret          string
	Issuer          string
	ExpirationHours int64
}

//NewJwtCreatorImpl - creates an instance of JwtCreatorImpl
func NewJwtCreatorImpl() (*JwtCreatorImpl, error) {
	secret := os.Getenv(secretKey)
	if len(secret) == 0 {
		return nil, myerr.NewServerError("Missing value for \"secret\" jwt config")
	}

	issuer := os.Getenv(issuerKey)
	if len(issuer) == 0 {
		return nil, myerr.NewServerError("Missing value for \"issuer\" jwt config")
	}

	expirationStr := os.Getenv(expirationKey)
	if len(expirationStr) == 0 {
		return nil, myerr.NewServerError("Missing value for \"expirationHours\" jwt config")
	}
	expirationHours, err := strconv.ParseInt(expirationStr, 10, 64)
	if err != nil {
		return nil, myerr.NewServerErrorWrap(err, "Wrong typeof value for \"expiration\" jwt config")
	}

	return &JwtCreatorImpl{
		Secret:          secret,
		Issuer:          issuer,
		ExpirationHours: int64(expirationHours),
	}, nil
}

// JwtClaim adds email as a claim to the token
type JwtClaim struct {
	UserID uint
	jwt.StandardClaims
}

//GenerateToken - generates a token, encrypting the userID in it
//returns the token and error
func (j *JwtCreatorImpl) GenerateToken(userID uint) (string, error) {
	claims := &JwtClaim{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(j.ExpirationHours)).Unix(),
			Issuer:    j.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.Secret))
}

//ValidateToken - validates a given JWT token
//returns the encrypted data in the token and error if the token is invalid
func (j *JwtCreatorImpl) ValidateToken(signedToken string) (*JwtClaim, error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&JwtClaim{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(j.Secret), nil
		},
	)

	if err != nil {
		return nil, myerr.NewClientError("Invalid Token. Please login again.")
	}

	claims, _ := token.Claims.(*JwtClaim)
	return claims, nil
}
