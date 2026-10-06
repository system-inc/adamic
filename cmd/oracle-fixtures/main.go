// oracle-fixtures lists discovered programs or renders their recorded counts.
package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/system-inc/adamic/internal/oracle/fixturedata"
)

func main() {
	repository := flag.String("repository", ".", "repository root")
	list := flag.Bool("list", false, "list discovered paths and options as JSON")
	counts := flag.Bool("counts", false, "render the recorded counts table")
	flag.Parse()
	if *list == *counts {
		fmt.Fprintln(os.Stderr, "choose exactly one of -list or -counts")
		os.Exit(2)
	}
	fixtures, err := fixturedata.Discover(*repository)
	if err != nil {
		fail(err)
	}
	if *counts {
		table, err := fixturedata.RenderCounts(fixtures)
		if err != nil {
			fail(err)
		}
		fmt.Print(table)
		return
	}
	type entry struct {
		Path         string   `json:"path"`
		Lowers       bool     `json:"lowers"`
		Checked      bool     `json:"checked"`
		Input        bool     `json:"input"`
		ArgumentsHex []string `json:"argumentsHex,omitempty"`
		Unreadable   bool     `json:"unreadable,omitempty"`
		Writes       bool     `json:"writes,omitempty"`
		Uncounted    bool     `json:"uncounted,omitempty"`
	}
	entries := make([]entry, 0, len(fixtures))
	for _, f := range fixtures {
		e := entry{Path: f.Path, Lowers: f.Lowers, Checked: f.Checked, Input: f.Input, Unreadable: f.Unreadable, Writes: f.Writes, Uncounted: f.Uncounted}
		for _, argument := range f.Arguments {
			e.ArgumentsHex = append(e.ArgumentsHex, hex.EncodeToString([]byte(argument)))
		}
		entries = append(entries, e)
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(entries); err != nil {
		fail(err)
	}
}

func fail(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
