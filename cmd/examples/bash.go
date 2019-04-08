package main

import (
	"fmt"
	"github.com/mitchfriedman/subprocess/subprocess"
)

func main() {
	child, _ := subprocess.NewSubProcess("bash")
	if err := child.Start(); err != nil {
		fmt.Printf("could not start: ", err)
	}
	defer child.Close()
	child.Interact()
}
