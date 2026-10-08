package main

import (
	"fmt"
	"os"

	"github.com/system-inc/adamic/stage3/scouts/step24/dumpdiff"
)

func main() { os.Exit(run()) }
func run() int {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: dumpdiff reference.dump actual.dump")
		return 2
	}
	left, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	defer left.Close()
	right, err := os.Open(os.Args[2])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	defer right.Close()
	difference, err := dumpdiff.Compare(left, right)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	if difference == nil {
		return 0
	}
	fmt.Printf("line %d\nreference %s: %q\nactual    %s: %q\n", difference.Line, difference.ReferencePath, difference.Reference, difference.ActualPath, difference.Actual)
	return 1
}
