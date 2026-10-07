package stage1progress

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

func Main() {
	root := flag.String("repo", ".", "Adamic repository with initialized cohere submodule")
	ref := flag.String("ref", "HEAD", "Git snapshot to inventory")
	branches := flag.Bool("pending", false, "inventory origin/main and every unintegrated fetched stage 1/scanner branch")
	asJSON := flag.Bool("json", false, "write complete machine-readable reports")
	flag.Parse()
	refs := []string{*ref}
	var integrated []string
	if *branches {
		pendingRefs, already, err := pending(*root)
		if err != nil {
			fail(err)
		}
		integrated = already
		refs = append([]string{"origin/main"}, pendingRefs...)
	}
	var reports []*report
	cache := map[string]*inventory{}
	for _, ref := range refs {
		report, err := measure(*root, ref, cache)
		if err != nil {
			fail(err)
		}
		reports = append(reports, report)
	}
	if *branches {
		combined, err := union(reports)
		if err != nil {
			fail(err)
		}
		reports = append(reports, combined)
	}
	if *asJSON {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(struct {
			Reports    []*report `json:"reports"`
			Integrated []string  `json:"already_integrated,omitempty"`
		}{reports, integrated}); err != nil {
			fail(err)
		}
		return
	}
	for _, report := range reports {
		fmt.Printf("\n%s %s | cohere %s\n", report.Ref, report.Commit, report.Cohere)
		writer := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(writer, "GO LINES\tPORTED\tSTATUS\tPACKAGE\tEVIDENCE")
		for _, row := range report.Packages {
			fmt.Fprintf(writer, "%d\t%d\t%s\t%s", row.GoLines, row.PortedLines, row.Status, row.Package)
			if len(row.Evidence) != 0 {
				fmt.Fprintf(writer, "\t%s", strings.Join(row.Evidence, ", "))
			}
			fmt.Fprintln(writer)
		}
		writer.Flush()
		fmt.Printf("TOTAL %d/%d Go lines = %.4f%% conservative credited coverage; %d complete, %d partial, %d not started packages\n", report.Ported, report.Total, report.Percent, report.Complete, report.Partial, report.NotStarted)
		for _, slice := range report.Slices {
			fmt.Printf("SCOPE %s credited=%t: %s | GAPS: %s\n", slice.Directory, slice.Credited, slice.Note, slice.GapSummary)
		}
		for _, dependency := range report.External {
			fmt.Println("EXTERNAL", dependency)
		}
	}
	if len(integrated) > 0 {
		fmt.Println("\nAlready integrated:", strings.Join(integrated, ", "))
	}
}
func fail(err error) { fmt.Fprintln(os.Stderr, "adamic-stage1-progress:", err); os.Exit(1) }
