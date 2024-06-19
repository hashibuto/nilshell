package main

import (
	"fmt"
	"strings"

	ns "github.com/hashibuto/nilshell"
)

var ac []*ns.AutoComplete = []*ns.AutoComplete{
	{
		Value:   "helpur",
		Display: "helpur (-h / --help)",
	},
	{
		Value:   "helpinhand",
		Display: "helpinhand (-g / --helper) awesome helpster",
	},
	{
		Value:   "helping",
		Display: "helping (-g / --helper)  helping of fiber a day",
	},
}

func main() {
	shell := ns.NewShell(
		"\033[33m » \033[0m",
		func(beforeCursor, afterCursor string, full string) []*ns.AutoComplete {
			newAc := []*ns.AutoComplete{}
			for _, acItem := range ac {
				if strings.HasPrefix(acItem.Value, beforeCursor) {
					newAc = append(newAc, acItem)
				}
			}

			return newAc
		},
		func(ns *ns.NilShell, cmd string) {
			fmt.Println("Executed a command")
		},
	)
	shell.ReadUntilTerm()
}
