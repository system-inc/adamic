// Command probe records stricter-option diagnostics without permitting emission.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/system-inc/adamic/internal/load"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: probe <tsconfig.json> | --production <source files...>")
		os.Exit(2)
	}
	var report *load.ProjectOptionReport
	var err error
	if os.Args[1] == "--production" {
		_, err = load.Load(os.Args[2:])
		var rejected *load.CheckError
		if errors.As(err, &rejected) {
			siteMessages := make(map[string]bool)
			for _, site := range rejected.OptionSites {
				siteMessages[site.Message] = true
			}
			report = &load.ProjectOptionReport{ProjectErrors: []string{}, Sites: rejected.OptionSites}
			for _, message := range rejected.Diagnostics {
				if !siteMessages[message] {
					report.ProjectErrors = append(report.ProjectErrors, message)
				}
			}
			err = nil
		} else if err == nil {
			report = &load.ProjectOptionReport{ProjectErrors: []string{}, Sites: []load.OptionSite{}}
		}
	} else if len(os.Args) == 2 {
		report, err = load.AuditProjectOptions(context.Background(), os.Args[1])
	} else {
		err = fmt.Errorf("audit accepts one tsconfig path")
	}
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
