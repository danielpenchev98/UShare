package common

import "time"

//BasicResponse - simple response to a client request
type BasicResponse struct {
	Status int `json:"status"`
}

/*ErrorResponse is sent to the client of the REST API when
there is an error with request or the server*/
type ErrorResponse struct {
	ErrorCode int    `json:"errorcode"` //status code of the request - 4xx or 5xx
	ErrorMsg  string `json:"message"`   //desription of the error
}

//RegistrationResponse is returned to the client when the registration was successfull
//it contains the statius of hist request and the jwt token
type RegistrationResponse struct {
	Status   int    `json:"status"`
	JWTToken string `json:"jwt_token"`
}

//LoginResponse - when the login is succesfull a JWT is sent to the user
type LoginResponse struct {
	Status int    `json:"status"`
	Token  string `json:"token"`
}

//GroupInfo - response payload, containing only the most important details about a group
type GroupInfo struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	OwnerID uint   `josn:"owner_id"`
}

//UserInfo - response payload, containing only the most important details about a user
type UserInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

//FileInfoResponse - response of a request for fetching information about file
type FileInfoResponse struct {
	ID         uint      `json:"file_id"`
	Name       string    `json:"file_name"`
	UploadedAt time.Time `json:"uploaded_at"`
	OwnerID    uint      `json:"owner_id"`
}
