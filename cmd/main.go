package main

import (
	"strings"

	ns "github.com/hashibuto/nilshell"
)

func main() {
	shell := ns.NewShell(
		"» ",
		func(beforeCursor, afterCursor string) []*ns.AutoComplete {
			if strings.HasPrefix("help", beforeCursor) {
				return []*ns.AutoComplete{
					{
						Name: "help",
					},
				}
			}

			return nil
		},
		func(ns *ns.NilShell, cmd string) {

		},
	)
	shell.ReadUntilTerm()
}
