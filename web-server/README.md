# Web-Server

## Description
The main idea of this server is to allow people to share freely their files with everyone.
In order to keep the anonymity of the users and the interformation they share, the following limitations are imposed:
* Users are grouped in `groups`. One user can belong to many groups and one group can have many members
* File are uploaded, given a specific `group`. Only the members of the `group` can access/view the `group` files
* The only identification of the user is his `username` (also his `id`)
Also there are limitations in terms of implementation:
* Only the `owner` of the `group` and the `owner` of the file can delete it from the group
* When the `owner` deletes the group or deletes his account, there is no transition of ownership (yet). Instead all group recources are deleted (files, memberships, etc)
* The group resources aren't deleted immediately. Instead, when the group is request to be deleted, the group swithces to `deactivated` state. And after a particular time period the rosources are erased. After this operation succeeds, the name of the `group` is available for usage.

## Configuration
The server uses the following external dependencies, which should be installed:
### Production
* `go` - preferably versions above `1.5.0`
* `github.com/dgrijalva/jwt-go` - used for validation/creation of JWTokens
* `github.com/gin-gonic/gin` - used for the implementation of the REST API
* `github.com/pkg/errors` - used for easier creation of errors
* `github.com/robfig/cron/v3` - used for the async job for deletion of group resources
* `golang.org/x/crypto` - used for encryption of user information
* `gorm.io/gorm` - used for mapping models (go structs) to sql tables
* `gorm.io/driver/postgres` - used for the communication with the `postgres` database
### Testing
* `github.com/DATA-DOG/go-sqlmock` - used for testing the request, sent to the database
* `github.com/golang/mock` - used for mocking external dependencies
* `github.com/onsi/ginkgo` - used as the main testing framework
* `github.com/onsi/gomega` - used for assertions

The following environment variables must be set:
### Server configuration
* `HOST` - env variable, containing the host name, on which the server will be running
* `PORT` - env variable, containing the port number, which the server will run on
### DB configuration
* `DB_NAME` - env variable, containing the name of the database
* `DB_USER` - env variable, containing the db username
* `DB_PASS` - env variable, containing the db password
* `DB_PORT` - env variable, containing the port on which the db server is running on
* `DB_HOST` - env variable, containing the domain of the db server
### Auth configuration
* `SECRET` - env variable, containing a value, used for the encryption/decryption of the token
* `ISSUER` - env variable, containing the name of authority, issuing the token
* `EXPIRATION` - env variable, containing the expiration time of the issued tokens (in hours)

## Installation
```bash
# Clone repo
git clone https://github.com/danielpenchev98/FMI-Golang.git

# Go to web-server subproject
cd FMI-Golang/UShare/web-server

# Build web-server - several dependencies will be downloaded
go build ./...

# Go to cmd dir, containg the server startup file
cd cmd

# Start the server
go run server.go
```

## Running tests
```bash
# Execute it in web-server directory
ginkgo ./...
```

## API endpoints
There are 2 types of endpoints - `public`, which can be access freely, and `protected`, which additionaly require `JWToken` in the `Auth Header` 
Also every server response sends `JSON object` with the `status code` of the request. This detail will be skipped in the table below.

|api endpoint | payload | usage | result |
|--|--|--|--|
|`POST /v1/public/user/registration` | `JSON object` containing username and password | User registration |-|
|`POST /v1/public/user/login`|`JSON object` containing username and password|User login|`JWToken`|
|`GET /v1/protected/users`|-|Fetch information about all users|Information records about users|
|`POST /v1/protected/group/creation`|`JSON object` containing the `group name` |New group with the specified name is created|-|
|`DELETE /v1/protected/group/deletion`|`JSON object` containing the `group name`|The group with the specified name is deleted|-|
|`POST /v1/protected/group/invitation`|`JSON object` containing the `group name` and the user's `username` |Membership created|-|
|`DELETE /v1/protected/group/membership/revocation`|`JSON object` containing the `group name` and the member's `username`|Membership revoked|-|
|`GET /v1/protected/group/users`| `QueryParameter` containing the `group name` |Fetch information about all members of a group | Information records about the members|
|`GET /v1/protected/groups`|-|Fetch information about all groups|Information records about the members|
|`POST /v1/protected/group/file/upload`|`Form-data` containing a file and `QueryParameter` containg the `group name`|File Upload|ID of the file(`file_id`)|
|`GET /v1/protected/group/file/download`|`QueryParameters` containing the `group name` and the `file_id`|File Download|File|
|`DELETE /v1/protected/group/file/deletion`|`JSON object` containing the `group name` and the `file_id`|File deletion|-|
|`GET /v1/protected/group/files`|`QueryParameter` containing the `group name`|Fetch information about all files for a given group|Information records about the files|

## AWS deployment
For more information please refer to [aws-doc.pdf](/web-server/docs/aws-doc.pdf) (*The document is written currently in Bulgarian*)
