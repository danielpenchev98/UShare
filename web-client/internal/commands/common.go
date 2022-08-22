package commands

import (
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
)

//GroupPayload - used as a payload of group requests
type GroupPayload struct {
	GroupName string `json:"group_name"`
}

//PrintTable - used for pritty printing records of information
func PrintTable(columNames table.Row, records []table.Row) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(columNames)
	t.AppendRows(records)
	t.Render()
}
