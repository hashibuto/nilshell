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
	{
		Value:   "dog",
		Display: "dog",
	},
	{
		Value:   "doggy",
		Display: "doggy",
	},
	{
		Value:   "doggo",
		Display: "doggo",
	},
	{
		Value:   "big1",
		Display: "big column is a big column, let's see how much we can fit into it to make it work",
	},
	{
		Value:   "big2",
		Display: "big column is a big column, let's see how much we can fit into it to make it work",
	},
	{
		Value:   "big3",
		Display: "big column is a big column, let's see how much we can fit into it to make it work",
	},
	{
		Value:   "big4",
		Display: "big column is a big column, let's see how much we can fit into it to make it work",
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
