package main

import (
	"fmt"
	"io"

	"github.com/system-inc/adamic/internal/ir"
)

func explainChecks(program *ir.Program, output io.Writer) {
	checks := ir.InsertedChecks(program)
	counts := make(map[string]int)
	for _, check := range checks {
		counts[check.Kind]++
		fmt.Fprintf(output, "%s: checked %s\n", relative(check.Where), check.Kind)
	}
	fmt.Fprintf(output, "checked: indexed-presence=%d catch-error=%d json-stringify-defined=%d optional-write=%d caught-type=%d\ntrusted: 0\n", counts["indexed-presence"], counts["catch-error"], counts["json-stringify-defined"], counts["optional-write"], counts["caught-type"])
}
