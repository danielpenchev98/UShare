package main

import (
	"fmt"
	"os"

	"github.com/danielpenchev98/UShare/web-client/internal/commands"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Print("No arguments were supplied")
		os.Exit(1)
	}

	command := os.Args[1]
	if command == "help" {
		commands.Help()
		return
	}

	commandsWithHostURL(command)
}

func commandsWithHostURL(command string) {
	hostURL := os.Getenv("HOST_URL")
	if hostURL == "" {
		fmt.Print("HOST_URL hasnt been set")
		return
	}

	switch command {
	case "login":
		commands.Login(hostURL)
	case "register":
		commands.RegisterUser(hostURL)
	default:
		commandsWithAuth(command, hostURL)
	}
}

func commandsWithAuth(command, hostURL string) {
	token := os.Getenv("JWT")
	if hostURL == "" {
		fmt.Print("JWT hasnt been set. Please set this environment variable. To get its value, please login again")
		return
	}

	switch command {
	case "create-group":
		commands.CreateGroup(hostURL, token)
	case "delete-group":
		commands.DeleteGroup(hostURL, token)
	case "add-member":
		commands.AddMember(hostURL, token)
	case "remove-member":
		commands.RemoveMember(hostURL, token)
	case "upload-file":
		commands.UploadFile(hostURL, token)
	case "download-file":
		commands.DownloadFile(hostURL, token)
	case "delete-file":
		commands.DeleteFile(hostURL, token)
	case "show-all-files":
		commands.ShowAllFilesInGroup(hostURL, token)
	case "show-all-groups":
		commands.ShowAllGroups(hostURL, token)
	case "show-all-users":
		commands.ShowAllUsers(hostURL, token)
	case "show-all-members":
		commands.ShowAllMembers(hostURL, token)
	default:
		fmt.Printf("Invalid command [%s]\n", command)
		commands.Help()
	}
}
