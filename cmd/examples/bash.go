package main

import (
	"expect/subprocess"
	"fmt"
)

func main() {
	child, _ := subprocess.NewSubProcess("bash")
	if err := child.Start(); err != nil {
		fmt.Println("could not start: ", err)
	}
	defer child.Close()

	fmt.Println("starting interact...")
	if err := child.Interact(); err != nil {
		fmt.Println("failde interacting")
	}
}
