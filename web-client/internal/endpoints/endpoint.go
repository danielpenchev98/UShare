package endpoints

const (
	//apiVersionPath - version of the api endpoint
	apiVersionPath = "/v1"
	//publicAPIPath - publicly accessible api path
	publicAPIPath = apiVersionPath + "/public"
	//protectedAPIPath - protected api path
	protectedAPIPath = apiVersionPath + "/protected"
	//LoginAPIEndpoint - api endpoint for user login
	LoginAPIEndpoint = publicAPIPath + "/user/login"
	//RegisterAPIEndpoint - api endpoint for user registration
	RegisterAPIEndpoint = publicAPIPath + "/user/registration"
	//CreateGroupAPIEndpoint - api endpoint for group creation
	CreateGroupAPIEndpoint = protectedAPIPath + "/group/creation"
	//DeleteGroupAPIEndpoint - api endpoint for group deletion
	DeleteGroupAPIEndpoint = protectedAPIPath + "/group/deletion"
	//AddMemberAPIEndpoint - api endpoint for adding an user to a group
	AddMemberAPIEndpoint = protectedAPIPath + "/group/invitation"
	//RemoveMemberAPIEndpoint - api endpoint for removing an user from a group
	RemoveMemberAPIEndpoint = protectedAPIPath + "/group/membership/revocation"
	//UploadFileAPIEndpoint - api endpoint for uploading a file for a specific group
	UploadFileAPIEndpoint = protectedAPIPath + "/group/file/upload"
	//DownloadFileAPIEndpoint - api endpoint for downloading a file from a specific group
	DownloadFileAPIEndpoint = protectedAPIPath + "/group/file/download"
	//DeleteFileAPIEndpoint - api endpoint for deleting file, given a group
	DeleteFileAPIEndpoint = protectedAPIPath + "/group/file/deletion"
	//GetAllFilesAPIEndpoint - api endpoint for fetching all files, uploaded for a specific group
	GetAllFilesAPIEndpoint = protectedAPIPath + "/group/files"
	//GetAllGroupsAPIEndpoint - api endpoint for fetching all existing groups
	GetAllGroupsAPIEndpoint = protectedAPIPath + "/groups"
	//GetAllUsersAPIEndpoint - api endpoint for fetching all users
	GetAllUsersAPIEndpoint = protectedAPIPath + "/users"
	//GetAllMembersAPIEndpoint - api endpoint for fetching all members of a group
	GetAllMembersAPIEndpoint = protectedAPIPath + "/group/users"
)
