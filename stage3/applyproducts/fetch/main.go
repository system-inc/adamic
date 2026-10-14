package main

import (
	"fmt"
	"github.com/system-inc/adamic/stage3/applyproducts"
	"os"
)

func main() {
	directory, err := applyproducts.Get()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(directory)
}
