package main

import (
	"fmt"
	"os"

	"github.com/codebuddy/redis-tui/cmd"
)

func main() {
	if err := cmd.Execute(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "redis-tui:", err)
		os.Exit(1)
	}
}
