package commands

import "github.com/jedib0t/go-pretty/v6/table"

//Help - shows information about all available commands
func Help() {
	commands := []table.Row{
		{"register", "register a new user", "-usr=<username>(Required) and -pass=<password>(Required)"},
		{"login", "login as a registered user", "-usr=<username>(Required) and -pass=<password>(Required)"},
		{"show-all-users", "show all existing users", "None"},
		{"create-group", "create a new group", "-grp=<group_name>(Required)"},
		{"delete-group", "delete group", "-grp=<group_name>(Required)"},
		{"show-all-groups", "show all existing groups", "None"},
		{"add-member", "add a new member to a group", "-usr=<username>(Required) and -grp=<group_name>(Required)"},
		{"remove-member", "revoke membership", "-usr=<username>(Required) and -grp=<group_name>(Required)"},
		{"show-all-members", "show all members of a group", "-grp=<group_name>(Required)"},
		{"upload-file", "upload a file to a group", "-grp=<group_name>(Required) and -filepath=<path_to_file>(Required)"},
		{"download-file", "download a file from a group", "-grp=<group_name>(Required), -fileid=<id_of_file>(Required) and -target=<output_file_path>(Required)"},
		{"delete-file", "delete file from a group", "-grp=<group_name>(Required) and -fileid=<id_of_file>(Required)"},
		{"show-all-files", "show all files from a group", "-grp=<group_name>(Required)"},
		{"help", "show all available commands", "None"},
	}

	columnNames := table.Row{"Command", "Description", "Flags"}
	PrintTable(columnNames, commands)
}
