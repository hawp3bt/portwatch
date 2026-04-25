package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/user/portwatch/internal/tags"
)

const defaultTagsPath = ".portwatch_tags.json"

// runTags handles the "tags" sub-command.
//
//	portwatch tags list
//	portwatch tags set <port> <label> [description]
//	portwatch tags remove <port>
func runTags(args []string) {
	fs := flag.NewFlagSet("tags", flag.ExitOnError)
	tagsFile := fs.String("file", defaultTagsPath, "path to tags JSON file")
	_ = fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: portwatch tags <list|set|remove> [args]")
		os.Exit(1)
	}

	reg, err := tags.New(*tagsFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tags: %v\n", err)
		os.Exit(1)
	}

	switch fs.Arg(0) {
	case "list":
		runTagsList(reg)
	case "set":
		runTagsSet(reg, fs.Args()[1:])
	case "remove":
		runTagsRemove(reg, fs.Args()[1:])
	default:
		fmt.Fprintf(os.Stderr, "tags: unknown sub-command %q\n", fs.Arg(0))
		os.Exit(1)
	}
}

func runTagsList(reg *tags.Registry) {
	list := reg.List()
	if len(list) == 0 {
		fmt.Println("no tags defined")
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PORT\tLABEL\tDESCRIPTION")
	for _, t := range list {
		fmt.Fprintf(w, "%d\t%s\t%s\n", t.Port, t.Label, t.Description)
	}
	w.Flush()
}

func runTagsSet(reg *tags.Registry, args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: portwatch tags set <port> <label> [description]")
		os.Exit(1)
	}
	port, err := strconv.Atoi(args[0])
	if err != nil || port < 1 || port > 65535 {
		fmt.Fprintf(os.Stderr, "tags set: invalid port %q\n", args[0])
		os.Exit(1)
	}
	desc := ""
	if len(args) >= 3 {
		desc = args[2]
	}
	if err := reg.Set(tags.Tag{Port: port, Label: args[1], Description: desc}); err != nil {
		fmt.Fprintf(os.Stderr, "tags set: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("tagged port %d as %q\n", port, args[1])
}

func runTagsRemove(reg *tags.Registry, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: portwatch tags remove <port>")
		os.Exit(1)
	}
	port, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "tags remove: invalid port %q\n", args[0])
		os.Exit(1)
	}
	if err := reg.Remove(port); err != nil {
		fmt.Fprintf(os.Stderr, "tags remove: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("removed tag for port %d\n", port)
}
