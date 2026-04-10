package app

import (
	"fmt"
	"strings"

	"github.com/example/setupx/internal/ui"
)

func RunInteractive() error {
	for {
		fmt.Println("\nsetupx interactive")
		fmt.Println("  1) import repo")
		fmt.Println("  2) list profiles")
		fmt.Println("  3) use profile")
		fmt.Println("  4) quit")
		choice, err := ui.Prompt("Select [1-4]: ")
		if err != nil {
			return err
		}
		switch strings.TrimSpace(choice) {
		case "1":
			url, err := ui.Prompt("Git URL: ")
			if err != nil {
				return err
			}
			if err := ImportRepo(url); err != nil {
				fmt.Printf("Import failed: %v\n", err)
			}
		case "2":
			if err := ListProfiles(); err != nil {
				fmt.Printf("List failed: %v\n", err)
			}
		case "3":
			id, err := ui.Prompt("Profile ID: ")
			if err != nil {
				return err
			}
			if err := UseProfile(id, "", false); err != nil {
				fmt.Printf("Apply failed: %v\n", err)
			}
		case "4", "q", "quit", "exit":
			return nil
		default:
			fmt.Println("Invalid choice")
		}
	}
}
