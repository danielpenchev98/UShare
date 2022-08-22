# Web-Client

This client is specifically created for communication with the REST API of the `web-server`. It abstracts all
the needed technical details and makes the user experience much smoother.

## Functionalities
The supported operations are:
* User Login/Registration/Deletion
* Group creation/deletion
* Add member to a specific group/Remove member from a specific group
* Upload/Download/Delete files

## Configurations
The CLI uses `github.com/go-resty/resty` for the request executions and `github.com/jedib0t/go-pretty` for
representing some of the information in form of a table.
Before using the client, one must also install `go`(preferably version `1.5.*`) and explicitly set the environment variable `HOST_URL`, to specify the host url 
of the server. For instance: `http://localhost:8080`.

## Installation
```bash
# Clone repo
git clone https://github.com/danielpenchev98/FMI-Golang.git

# Go to web-server subproject
cd FMI-Golang/UShare/web-client

# Build web-server - several dependencies will be downloaded
go build ./...

# Go to cmd dir, containg the server startup file
cd cmd

# Start the server
go run client.go
```

## Commands
Every single command begins with executing `client.go` in the `cmd` package

### Help
```bash
go run client.go help
```
Result: Shows all available commands with description and their flags

### Registration
```bash
go run client.go register -usr=<username> -pass=<password>
```
Result: A new user is registered in the system

### Login
```bash
go run client.go login -usr=<username> -pass=<password>
```
Result: The user is logged in the system. A JWToken is issued to the user 
and he has to create an env variable named `JWT` with the value of the token

### Show users
```bash
go run client.go show-all-users
```
Result: A table, containing information about all users is displayed. The information contains the `id` of the user and its `username`
### Create group
```bash
go run client.go create-group -grp=<group_name>
```
Result: A new group with the specified name is created. And the only member of that group is you, the owner

### Delete group
```bash
go run client.go delete-group -grp=<group_name>
```
Result: If the user, executing this command, is the owner, then the whole group is deleted (the files and memberships also)
Transition of ownership is yet to be implemented

### Show groups
```bash
go run client.go show-all-groups
```
Result: A table, containing information about all groups is displayed. The information contains the `name` of the group,
the `id` of the group and the `id` of the owner(User)

### Add member
```bash
go run client.go add-member -grp=<group_name> -usr=<username>
```
Result: An existing user is added to the group. He can now upload files ot it.

### Remove member
```bash
go run client.go remove-member -grp=<group_name> -usr=<username>
```
Result: An existing member in this group is removed from it. His files arent removed from the group

### Show members
```bash
go run client.go show-all-members -grp=<group_name>
```
Result: Information is shown about every member of the group. This information includes the user `id` and its `username`

### Upload file
```bash
go run client.go upload-file -grp=<group_name> -filepath=<full_file_path>
```
Result: The file is uploaded on the server and only members of the group can see its existence. The id of the file is shown in the output.

### Delete file
```bash
go run client.go delete-file -grp=<group_name> -fileid=<full_id>
```
Result: The file is removed from the group. Only the group owner and the owner of the file can remove it. Only with the `file_id` one can delete it because of multiple files with the same name.

### Download file
```bash
go run client.go delete-file -grp=<group_name> -fileid=<full_id> -target=<target_file_path>
```
Result: The file is downloaded from the server. `target_file_path` should be also a full path in the filesystem.

### Show files
```bash
go run client.go show-all-file -grp=<group_name>
```
Result: Information about all files for a particular group is deiplayed. This information contains the file `id`, `name`, `UploadedAt` timestamp and the `owner_id`



