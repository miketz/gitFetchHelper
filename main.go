package main

import (
	"fmt"
	"log"
	"os/exec"
)

func main() {
	cmd := exec.Command("git", "branch")
	stdout, err := cmd.Output()
	if err != nil {
		log.Fatalf("error: %v", err.Error())
	}
	fmt.Println(string(stdout))
}