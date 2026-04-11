package app

import (
	"fmt"

	"github.com/example/agent-equip/internal/store"
)

func ListProfiles() error {
	r, err := store.LoadRegistry()
	if err != nil {
		return err
	}
	if len(r.Profiles) == 0 {
		fmt.Println("No profiles imported yet.")
		return nil
	}
	for _, p := range r.Profiles {
		fmt.Printf("- %s [%s] %s\n", p.ID, p.Tool, p.Name)
		if p.Description != "" {
			fmt.Printf("    %s\n", p.Description)
		}
	}
	return nil
}
