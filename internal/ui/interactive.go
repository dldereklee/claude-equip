package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func Prompt(msg string) (string, error) {
	fmt.Print(msg)
	r := bufio.NewReader(os.Stdin)
	v, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(v), nil
}

func ChooseScope() (string, error) {
	fmt.Println("Apply config to:")
	fmt.Println("  1) local")
	fmt.Println("  2) global")
	v, err := Prompt("Select [1-2]: ")
	if err != nil {
		return "", err
	}
	if v == "2" || strings.EqualFold(v, "global") {
		return "global", nil
	}
	return "local", nil
}
