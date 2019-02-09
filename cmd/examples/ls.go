package main

import (
	"expect/subprocess"
	"fmt"
	"regexp"
	"time"
)

func main() {
	child, _ := subprocess.NewSubProcess("ls", "-l")
	defer child.Close()

	child.Start()

	expressions := []*regexp.Regexp{
		regexp.MustCompile(".mod"),
	}

	index, err := child.ExpectExpressionsWithTimeout(expressions, 5*time.Second)
	if err != nil {
		fmt.Println("err expecting: ", err)
	}

	if index < 0 {
		return
	}

	fmt.Println("done")
}
