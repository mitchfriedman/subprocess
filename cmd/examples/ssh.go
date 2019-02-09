package main

import (
	"expect/subprocess"
	"flag"
	"fmt"
	"regexp"
	"time"
)

var destination = flag.String("destination", "", "address of ssh server")

func main() {
	flag.Parse()
	fmt.Printf("sshing into: %s", *destination)
	child, err := subprocess.NewSubProcess("ssh", *destination)
	if err != nil {
		fmt.Println("error creating child: ", err)
		return
	}

	defer child.Close()
	err = child.Start()
	if err != nil {
		fmt.Println("error starting child: ", err)
		return
	}

	found, err := child.ExpectWithTimeout(regexp.MustCompile("bash.*"), time.Second*20)
	if err != nil {
		fmt.Println("error expecting: ", err)
		return
	}

	if !found {
		fmt.Println("did not see expected")
		return
	}

	fmt.Println("successfully failed to connect!")
}
