package main

import (
	"fmt"
	"log"
	"os/exec"
)

func main() {
	cmd := exec.Command("ls", "-la")
	stdout, err := cmd.Output()
	if err != nil {
		log.Fatalf("failed")
	}
	fmt.Println(string(stdout))
}
