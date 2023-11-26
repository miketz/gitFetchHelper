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
	if len(stdout) == 0 {
		fmt.Printf("no output\n")
		return
	}
	fmt.Println(string(stdout))
}
