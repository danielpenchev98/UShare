package commands

import (
	"flag"
	"fmt"
	"os"

	"github.com/danielpenchev98/UShare/web-client/internal/endpoints"
	"github.com/danielpenchev98/UShare/web-client/internal/restclient"
	"github.com/jedib0t/go-pretty/v6/table"
)

//LoginResponse - response, containing the jw token
type LoginResponse struct {
	Status int    `json:"status"`
	Token  string `json:"token"`
}

//CredentialsPayload - information used for the login and registration of user
type CredentialsPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

//UserInfo - contains information about a user
type UserInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

//UsersInfoResponse - response containing information about multiple users
type UsersInfoResponse struct {
	Status    uint       `json:"status"`
	UsersInfo []UserInfo `json:"users"`
}

//RegisterUser - command for registration of user
func RegisterUser(hostURL string) {
	registrationCommand := flag.NewFlagSet("register", flag.ExitOnError)

	username := registrationCommand.String("usr", "", "username")
	password := registrationCommand.String("pass", "", "password")

	registrationCommand.Parse(os.Args[2:])

	if *username == "" && *password == "" {
		registrationCommand.PrintDefaults()
		return
	}

	rqBody := CredentialsPayload{
		Username: *username,
		Password: *password,
	}

	restClient := restclient.NewRestClientImpl("")
	url := hostURL + endpoints.RegisterAPIEndpoint
	err := restClient.Post(url, &rqBody, nil)

	if err != nil {
		fmt.Printf("Problem with the registration request. %s\n", err.Error())
		return
	}

	fmt.Println("User successfully created")
}

//Login - command for login of user
func Login(hostURL string) {
	loginCommand := flag.NewFlagSet("login", flag.ExitOnError)

	username := loginCommand.String("usr", "", "username")
	password := loginCommand.String("pass", "", "password")

	loginCommand.Parse(os.Args[2:])

	if *username == "" || *password == "" {
		loginCommand.PrintDefaults()
		return
	}

	rqBody := CredentialsPayload{
		Username: *username,
		Password: *password,
	}

	successBody := LoginResponse{}

	restClient := restclient.NewRestClientImpl("")
	url := hostURL + endpoints.LoginAPIEndpoint
	err := restClient.Post(url, &rqBody, &successBody)

	if err != nil {
		fmt.Printf("Problem with the login request. %s\n", err.Error())
		return
	}

	fmt.Println("Login is successful")
	fmt.Printf("Please set the env variable 'JWT' with the following value:\n%s\n", successBody.Token)
}

//ShowAllUsers - command for showing information about all users
func ShowAllUsers(hostURL string, token string) {
	successBody := UsersInfoResponse{}

	restClient := restclient.NewRestClientImpl(token)
	url := hostURL + endpoints.GetAllUsersAPIEndpoint
	err := restClient.Get(url, &successBody)

	if err != nil {
		fmt.Printf("Problem with the group creation request. %s\n", err.Error())
		return
	}

	tableRows := make([]table.Row, len(successBody.UsersInfo))
	for _, userInfo := range successBody.UsersInfo {
		tableRows = append(tableRows, table.Row{userInfo.ID, userInfo.Username})
	}
	PrintTable(table.Row{"ID", "Username"}, tableRows)
}

//ShowAllMembers - command for showing information about all members of a group
func ShowAllMembers(hostURL, token string) {
	getAllMembers := flag.NewFlagSet("show-all-members", flag.ExitOnError)
	groupName := getAllMembers.String("grp", "", "Name of the group")

	getAllMembers.Parse(os.Args[2:])

	if *groupName == "" {
		getAllMembers.PrintDefaults()
		os.Exit(1)
	}

	successBody := UsersInfoResponse{}
	restClient := restclient.NewRestClientImpl(token)
	url := fmt.Sprintf("%s%s?group_name=%s", hostURL, endpoints.GetAllMembersAPIEndpoint, *groupName)
	err := restClient.Get(url, &successBody)

	if err != nil {
		fmt.Printf("Problem with the group creation request. %s\n", err.Error())
		return
	}

	tableRows := make([]table.Row, len(successBody.UsersInfo))
	for _, userInfo := range successBody.UsersInfo {
		tableRows = append(tableRows, table.Row{userInfo.ID, userInfo.Username})
	}
	PrintTable(table.Row{"ID", "Username"}, tableRows)
}
