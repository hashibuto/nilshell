package main

import (
	"fmt"

	ns "github.com/hashibuto/nilshell"
)

func completer(beforeCursor, afterCursor, full string) *ns.Suggestions {
	return nil
}

func main() {
	r := ns.NewReader(ns.ReaderConfig{
		Debug:   true,
		LogFile: "/tmp/log.txt",
		ProcessFunction: func(s string) error {
			fmt.Println("got command")
			return nil
		},
	})
	r.ReadLoop()
}
