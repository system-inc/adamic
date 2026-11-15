// Command scan-all is built with the measurement overlay; the ordinary binary
// refuses to run. The real driver lives in ../main.go.txt.
// Compiler consumer marker: github.com/system-inc/adamic/internal/lower.
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "scan-all requires its generated no-output overlay; see stage3/census/scan-all/README.md")
	os.Exit(2)
}
