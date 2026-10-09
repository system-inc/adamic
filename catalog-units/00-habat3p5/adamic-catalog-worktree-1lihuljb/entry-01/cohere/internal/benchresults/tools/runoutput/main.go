// Command runoutput prints what one measured run said about itself, read from its `--json` output
// (benchresults.ParseRunOutput), as quiet_machine.sh's columns: engine seconds, findings and cache use,
// tab-separated. quiet_machine.sh builds it once and runs it after each measured run, outside the timing:
//
//	runoutput <log>
//
// It exits 1 when the log holds no summary it can read, so the benchmark stops rather than recording a
// column it could not fill.
package main

import (
	"fmt"
	"os"

	"github.com/system-inc/cohere/internal/benchresults"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: runoutput <log>")
		os.Exit(2)
	}
	output, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "runoutput: %v\n", err)
		os.Exit(1)
	}
	run, err := benchresults.ParseRunOutput(output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "runoutput: %s: %v\n", os.Args[1], err)
		os.Exit(1)
	}
	fmt.Printf("%.3f\t%d\t%s\n", run.EngineSeconds, run.Findings, run.Cache)
}
