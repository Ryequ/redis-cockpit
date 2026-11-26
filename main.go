package main

import (
	"fmt"
	"os"

	"github.com/Ryequ/redis-cockpit/cmd"
)

func main() {
	if err := cmd.Execute(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "redis-cockpit:", err)
		os.Exit(1)
	}
}
