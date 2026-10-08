// Command probe records stricter-option diagnostics without permitting emission.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/system-inc/adamic/internal/load"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: probe <tsconfig.json>")
		os.Exit(2)
	}
	report, err := load.AuditProjectOptions(context.Background(), os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
