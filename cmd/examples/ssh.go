package main

import (
	"flag"
	"fmt"
	"github.com/mitchfriedman/subprocess/subprocess"
	"regexp"
	"time"
)

var destination = flag.String("destination", "", "address of ssh server")

func main() {
	child, err := subprocess.NewSubProcess("ssh", "user@host")
	if err != nil {
		fmt.Printf("could not create subprocess %v\n", err)
		return
	}
	if err := child.Start(); err != nil {
		fmt.Printf("could not start subprocess %v\n", err)
		return
	}
	defer child.Close()

	line := regexp.MustCompile("Are you sure you want to continue connecting.*")
	found, err := child.ExpectWithTimeout(line, 5*time.Second)
	if err != nil || !found {
		fmt.Printf("failed to ssh: %v\n", err)
	}
}
