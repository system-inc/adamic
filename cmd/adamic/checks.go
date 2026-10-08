package main

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

func renderWithChecks(backend, path string) int {
	program, code := compile(path)
	if program == nil {
		return code
	}
	if native.UsesTSGo(program) {
		if backend == "js" {
			fmt.Fprintln(os.Stderr, "adamic: tsgo is an external native checker library; JavaScript is not supported")
		} else {
			fmt.Fprintln(os.Stderr, "adamic: tsgo requires a native build with --tsgo <archive>")
		}
		return 1
	}
	if len(program.NonNullChecks.Sites) == 0 || len(program.PredicateChecks.Sites) != 0 {
		explainPredicateChecks(os.Stderr, program)
	}
	explainNonNullChecks(os.Stderr, program)
	if backend == "js" {
		fmt.Print(javascript.JavaScript(program))
	} else {
		fmt.Print(native.C(program))
	}
	return 0
}

func explainPredicateChecks(output io.Writer, program *ir.Program) {
	sites := append([]ir.PredicateCallCheck(nil), program.PredicateChecks.Sites...)
	sort.SliceStable(sites, func(i, j int) bool { return sites[i].Where < sites[j].Where })
	for _, site := range sites {
		if site.Overload == 0 {
			fmt.Fprintf(output, "%s: predicate %s\n", relative(site.Where), site.Function)
		} else {
			fmt.Fprintf(output, "%s: predicate overload %d of %s\n", relative(site.Where), site.Overload, site.Function)
		}
		for _, direction := range site.Directions {
			status := direction.Status
			if status == "unobservable" {
				status += " (proven)"
			}
			fmt.Fprintf(output, "  %s: %s; %s\n", direction.Direction, status, direction.Reason)
		}
	}
	counts := program.PredicateChecks
	fmt.Fprintf(output, "adamic: predicate checks: proven %d checked %d unobservable %d\n", counts.Proven, counts.Checked, counts.Unobservable)
}

func explainNonNullChecks(output io.Writer, program *ir.Program) {
	if len(program.NonNullChecks.Sites) != 0 {
		sites := append([]ir.NonNullCheck(nil), program.NonNullChecks.Sites...)
		sort.SliceStable(sites, func(i, j int) bool { return sites[i].Where < sites[j].Where })
		for _, site := range sites {
			status := "checked"
			if site.Proven {
				status = "proven"
			}
			fmt.Fprintf(output, "%s: non-null assertion %s: %s\n", relative(site.Where), site.Expression, status)
		}
		fmt.Fprintf(output, "adamic: non-null checks: proven %d checked %d\n", program.NonNullChecks.Proven, program.NonNullChecks.Checked)
	}
	if len(program.NonNullChecks.Sites) != 0 {
		fmt.Fprintf(output, "adamic: checks: proven %d checked %d unobservable 0\n", program.NonNullChecks.Proven, program.NonNullChecks.Checked)
	}
}
