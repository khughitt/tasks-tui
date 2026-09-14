// Command tasks-tui is a keyboard-driven terminal front end for the tasks tracker.
package main

import (
	"fmt"
	"os"
)

const version = "0.1.0"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println("tasks-tui", version)
		return
	}
	fmt.Fprintln(os.Stderr, "tasks-tui: not yet wired")
	os.Exit(1)
}
