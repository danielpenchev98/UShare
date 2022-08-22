package common

//RequestWithCredentials - request representation for login
type RequestWithCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

//GroupPayload - request payload, containing the group name
type GroupPayload struct {
	GroupName string `json:"group_name"`
}

//GroupMembershipPayload - request payload, containing the group name and username
type GroupMembershipPayload struct {
	GroupPayload
	Username string `json:"username"`
}

//FileRequestPayload - request payload, containing the group name and the file id, owned by that group
type FileRequestPayload struct {
	GroupPayload
	FileID uint `json:"file_id"`
}
