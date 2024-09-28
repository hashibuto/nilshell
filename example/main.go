package main

import (
	"fmt"
	"strings"

	ns "github.com/hashibuto/nilshell"
	"github.com/hashibuto/nimble"
)

var suggestions = []*ns.Suggestion{
	{
		Value:   "carrot",
		Display: "carrot (is an orange vegetable)",
	},
	{
		Value:   "cucumber",
		Display: "cucumber (green and refreshing)",
	},
	{
		Value:   "zucchini",
		Display: "zuccini (a kind of squash)",
	},
	{
		Value:   "tomato",
		Display: "tomato (great on salad)",
	},
	{
		Value:   "pommodori",
		Display: "pommodori (a kind of tomato)",
	},
	{
		Value:   "pepper",
		Display: "pepper (green or red)",
	},
	{
		Value:   "paprika",
		Display: "paprika",
	},
	{
		Value:   "tom-and-jerry",
		Display: "tom-and-jerry (cat and mouse)",
	},
	{
		Value:   "zoo",
		Display: "zoo",
	},
	{
		Value:   "papa",
		Display: "papa (father)",
	},
	{
		Value:   "cuckoo-clock",
		Display: "cuckoo-clock (a clock with bird sound)",
	},
	{
		Value:   "cartographer",
		Display: "cartographer (one who does mapping?)",
	},
}

func completer(beforeCursor, afterCursor, full string) *ns.Suggestions {
	x := strings.ToLower(beforeCursor)
	suggs := nimble.Filter[*ns.Suggestion](func(index int, v *ns.Suggestion) bool {
		return strings.HasPrefix(strings.ToLower(v.Value), x)
	}, suggestions...)

	return &ns.Suggestions{
		Total: len(suggs),
		Items: suggs,
	}
}

func main() {
	r := ns.NewReader(ns.ReaderConfig{
		Debug:   true,
		LogFile: "/tmp/log.txt",
		ProcessFunction: func(s string) error {
			fmt.Println("got command")
			return nil
		},
		CompletionFunction: completer,
	})
	r.ReadLoop()
}
