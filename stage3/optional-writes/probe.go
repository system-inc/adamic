// Probe distinguishes scheduled contracts from successfully lowered checks.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"os"
)

func main() {
	result := struct {
		Sites     []load.OptionSite `json:"sites"`
		Scheduled []load.OptionSite `json:"scheduled"`
		Checked   []string          `json:"checked"`
		Errors    []string          `json:"errors"`
	}{Sites: []load.OptionSite{}, Scheduled: []load.OptionSite{}, Checked: []string{}, Errors: []string{}}
	program, err := load.Load(os.Args[1:])
	if err != nil {
		var rejected *load.CheckError
		if errors.As(err, &rejected) {
			result.Sites = rejected.OptionSites
			result.Scheduled = rejected.ScheduledOptionSites
			result.Errors = rejected.Diagnostics
		} else {
			result.Errors = append(result.Errors, err.Error())
		}
	} else {
		_, err = lower.Lower(context.Background(), program)
		if err != nil {
			result.Errors = append(result.Errors, err.Error())
		} else {
			result.Checked = program.ExplainedOptionalChecks()
		}
	}
	if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
