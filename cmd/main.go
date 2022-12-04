package main

import (
	ns "github.com/hashibuto/nilshell"
)

func main() {
	shell := ns.NewShell(
		"» ",
		func(beforeCursor, afterCursor string) []*ns.AutoComplete {
			return nil
		},
		func(ns *ns.NilShell, cmd string) {},
	)
	shell.ReadUntilTerm()
}
