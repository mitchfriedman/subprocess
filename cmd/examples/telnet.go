package main

import (
	"expect/subprocess"
	"fmt"
	"regexp"
	"time"
)

func main() {
	child, err := subprocess.NewSubProcess("telnet")
	if err != nil {
		fmt.Println("error creating child: ", err)
		return
	}

	defer child.Close()
	child.Start()

	found, err := child.ExpectWithTimeout(regexp.MustCompile("telnet>"), time.Second*20)
	if err != nil {
		fmt.Println("error expecting: ", err)
		return
	}

	if !found {
		fmt.Println("did not see expected")
		return
	}

	/*
		err = child.SendLine("open localhost")
		if err != nil {
			fmt.Println("error sending: ", err)
			return
		}

		found, err = child.Expect(regexp.MustCompile("telnet: Unable to connect to remote host"))
		if err != nil {
			fmt.Println("error expecting: ", err)
			return
		}

		if !found {
			fmt.Println("did not see expected")
			return
		}*/

	fmt.Println("successfully failed to connect!")
}
