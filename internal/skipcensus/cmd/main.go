// skipcensus scans source or checks a go test -json log without caching results.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/system-inc/adamic/internal/skipcensus"
)

func main() {
	root := flag.String("root", ".", "repository root")
	audit := flag.Bool("audit", false, "audit source declarations without checking a log")
	scan := flag.Bool("scan", false, "print AST census instead of checking a log")
	flag.Parse()
	rows, err := skipcensus.Scan(*root)
	if err != nil {
		fatal(err)
	}
	if *audit {
		fmt.Printf("census: %d annotated sites\n", len(rows))
		return
	}
	if *scan {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(rows); err != nil {
			fatal(err)
		}
		return
	}
	if flag.NArg() != 1 {
		fatal(fmt.Errorf("usage: skipcensus [-root directory] log.jsonl"))
	}
	log, err := os.Open(flag.Arg(0))
	if err != nil {
		fatal(err)
	}
	defer log.Close()
	if err := skipcensus.CheckLog(log, os.Stdout, rows); err != nil {
		fatal(err)
	}
}
func fatal(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
