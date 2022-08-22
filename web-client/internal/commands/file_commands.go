package commands

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/danielpenchev98/UShare/web-client/internal/endpoints"
	"github.com/danielpenchev98/UShare/web-client/internal/restclient"
	"github.com/jedib0t/go-pretty/v6/table"
)

//FileUploadResponse - used to extract the id of the file, which was uploaded on the server
type FileUploadResponse struct {
	FileID uint `json:"file_id"`
}

//FileRequest - used to request file information
type FileRequest struct {
	GroupPayload
	FileID uint `json:"file_id"`
}

//FileInfo - contains all the information about a file
type FileInfo struct {
	ID         uint      `json:"file_id"`
	Name       string    `json:"file_name"`
	OwnerID    uint      `json:"owner_id"`
	UploadedAt time.Time `json:"uploaded_at"`
}

//FilesInfoResponse - response, containing information about multiple files
type FilesInfoResponse struct {
	Status    int        `json:"status"`
	FilesInfo []FileInfo `json:"files"`
}

//UploadFile - command for uploading a file to the server
func UploadFile(hostURL, token string) {
	uploadFileCommand := flag.NewFlagSet("upload-file", flag.ExitOnError)
	filePath := uploadFileCommand.String("filepath", "", "Path to the file")
	groupName := uploadFileCommand.String("grp", "", "Name of the group, in which the file will be uploaded")

	uploadFileCommand.Parse(os.Args[2:])
	if *groupName == "" || *filePath == "" {
		uploadFileCommand.PrintDefaults()
		return
	}

	successBody := FileUploadResponse{}

	restClient := restclient.NewRestClientImpl(token)
	url := fmt.Sprintf("%s%s?group_name=%s", hostURL, endpoints.UploadFileAPIEndpoint, *groupName)
	err := restClient.UploadFile(url, *filePath, &successBody)

	if err != nil {
		fmt.Printf("Problem with the file upload request. %s\n", err.Error())
		return
	}

	fmt.Printf("File was successfully uploaded in group %s.\n The id of the file is %d\n", *groupName, successBody.FileID)
}

//DownloadFile - command for downloading a file from the server
func DownloadFile(hostURL, token string) {
	downloadFileCommand := flag.NewFlagSet("download-file", flag.ExitOnError)
	fileID := downloadFileCommand.Int("fileid", -1, "File id")
	groupName := downloadFileCommand.String("grp", "", "Name of the group, owning the file")
	targetPath := downloadFileCommand.String("target", "", "Target destination of file")

	downloadFileCommand.Parse(os.Args[2:])

	if *fileID == -1 || *groupName == "" || *targetPath == "" {
		downloadFileCommand.PrintDefaults()
		return
	}

	restClient := restclient.NewRestClientImpl(token)
	url := fmt.Sprintf("%s%s?group_name=%s&file_id=%d", hostURL, endpoints.DownloadFileAPIEndpoint, *groupName, *fileID)
	err := restClient.DownloadFile(url, *targetPath)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("File was successfully download.")
}

//DeleteFile - command for deletion of file on the server
func DeleteFile(hostURL, token string) {
	deleteFileCommand := flag.NewFlagSet("delete-file", flag.ExitOnError)
	fileID := deleteFileCommand.Int("fileid", -1, "File id")
	groupName := deleteFileCommand.String("grp", "", "Name of the group")

	deleteFileCommand.Parse(os.Args[2:])

	if *fileID == -1 || *groupName == "" {
		deleteFileCommand.PrintDefaults()
		return
	}

	reqBody := FileRequest{
		FileID: uint(*fileID),
	}
	reqBody.GroupName = *groupName

	restClient := restclient.NewRestClientImpl(token)
	url := hostURL + endpoints.DeleteFileAPIEndpoint
	err := restClient.Delete(url, &reqBody, nil)

	if err != nil {
		fmt.Printf("Problem with the file deletion request. %s\n", err.Error())
		return
	}

	fmt.Println("File was successfully deleted")
}

//ShowAllFilesInGroup - command for fetching information about all files uploaded for a specific group
func ShowAllFilesInGroup(hostURL, token string) {
	getAllFilesCommand := flag.NewFlagSet("show-all-files", flag.ExitOnError)
	groupName := getAllFilesCommand.String("grp", "", "Name of the group")

	getAllFilesCommand.Parse(os.Args[2:])

	if *groupName == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	successBody := FilesInfoResponse{}
	restClient := restclient.NewRestClientImpl(token)
	url := fmt.Sprintf("%s%s?group_name=%s", hostURL, endpoints.GetAllFilesAPIEndpoint, *groupName)
	err := restClient.Get(url, &successBody)

	if err != nil {
		fmt.Printf("Problem with the retrieval of group files. %s\n", err.Error())
		return
	}

	tableRows := make([]table.Row, len(successBody.FilesInfo))
	for _, fileInfo := range successBody.FilesInfo {
		tableRows = append(tableRows, table.Row{fileInfo.ID, fileInfo.Name, fileInfo.UploadedAt, fileInfo.OwnerID})
	}
	PrintTable(table.Row{"ID", "Name", "UploadedAt", "OwnerID"}, tableRows)
}
