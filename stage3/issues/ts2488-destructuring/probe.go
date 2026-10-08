// This probe asks Adamic's loader for diagnostics without lowering the program.
package main

import (
	"fmt"
	"github.com/system-inc/adamic/internal/load"
	"os"
)

func main() {
	_, err := load.Load(os.Args[1:])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
