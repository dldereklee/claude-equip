package main

import (
	"fmt"
	"os"

	"github.com/example/agent-equip/internal/equip"
)

func main() {
	if err := equip.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
