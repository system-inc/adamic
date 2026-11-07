package oracle

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// Checked predicate results the direction walk proves read (#6fas8b6, Lower's ruling of Oct 29), in
// the three shapes tsc writes, each its own program with predicates that tell the truth: argument,
// getDeclaredExpandoInitializer's call answer passed as an argument under &&; inferred_call,
// isEffectiveModuleDeclaration's inferred predicate from || over checked calls, called where its false
// side leaves the type whole; inferred_some, isScopeMarker's, passed to some's (value: T) => boolean.
// body_call holds a checked call in the body of a function the checker infers a predicate for, outside
// its returned expression, which is classified as any call is. Each lowers and agrees with Node on
// sanitized native and the JavaScript backend. Mutants: the argument's return removed stops argument
// as "read by a CallExpression inside a condition"; the inferred rule taken for the whole body stops
// body_call at isLeaf's call.
func TestPredicateDirectionUsesLower(test *testing.T) {
	test.Parallel()
	// #80cs7rz: inferred_call is NotYet now: symbol.valueDeclaration.name reads through ir.Defined (the read the
	// `=== undefined` test narrowed), which the allocation graph doesn't follow, so its reach is unknown; the per-read
	// check wins it back. Pinned in TestViewWidenedPriceAccepted.
	for _, name := range []string{"argument", "inferred_some", "body_call"} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			path, absolutePathError := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/predicate_direction_uses", name+".ts"),
			)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			truth := onNode(test, path)
			if truth.exitCode != 0 || len(truth.stderr) != 0 || len(truth.stdout) == 0 {
				test.Fatalf("Node: %+v", truth)
			}
			program, loweredError := lowered(test, path)
			if loweredError != nil {
				test.Fatal(loweredError)
			}
			if program.PredicateChecks.Checked == 0 {
				test.Fatalf("no checked predicate call: %+v", program.PredicateChecks)
			}
			sanitized, _ := nativelyUncached(test, program)
			for _, got := range []run{sanitized, onJavaScriptBackend(test, program)} {
				if difference := disagreement(truth, got); difference != "" {
					test.Fatal(difference)
				}
			}
		})
	}
}

// inferred_some with isExportAssignment lying on true, saying yes to a V: isScopeMarker's true is
// its calls' true, so the lie stops at isExportAssignment's own true check, with the located message,
// on both backends, where Node prints true.
func TestPredicateDirectionUsesInferredLieStops(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/predicate_direction_uses/inferred_lie.ts"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	if difference := disagreement(run{stdout: []byte("true\n")}, onNode(test, path)); difference != "" {
		test.Fatal("source Node: " + difference)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	where := ""
	for _, site := range program.PredicateChecks.Sites {
		if site.Function == "isExportAssignment" {
			where = site.Where
		}
	}
	if where == "" {
		test.Fatalf("no check at isExportAssignment's call: %+v", program.PredicateChecks.Sites)
	}
	want := run{stderr: []byte(fmt.Sprintf(
		"adamic: panic: predicate isExportAssignment at %s true branch: source Statement, target ExportAssignment\n",
		where,
	)), exitCode: 70}
	sanitized, _ := nativelyUncached(test, program)
	for _, got := range []run{sanitized, onJavaScriptBackend(test, program)} {
		if difference := disagreement(want, got); difference != "" {
			test.Fatal(difference)
		}
	}
}

// Inferred predicates whose false side the walk can't clear stay NotYet at the checked call, each
// naming why: not, a ! in the returned expression; false_read, a caller reading n.c on the false side
// of a direct call; predicate_parameter, the function passed to a parameter typed typeof isAOrB, whose
// calls narrow by its claim; alias, a call outside the returned expression initializing the const it
// returns. filter, the function passed to filter's value is S, never reaches the call: the contract
// pass refuses it first, naming the argument. Each runs on Node, and its whole first diagnostic is
// pinned in its .stopped file. Mutants: ! admitted lowers not; the false-reading call admitted lowers
// false_read; a predicate-typed parameter admitted lowers predicate_parameter; the alias let through
// lowers alias.
func TestPredicateDirectionUsesStayNotYet(test *testing.T) {
	test.Parallel()
	for _, name := range []string{"not", "false_read", "predicate_parameter", "alias", "filter"} {
		path := "internal/oracle/testdata/predicate_direction_uses/stopped/" + name + ".ts"
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			absolute, absolutePathError := filepath.Abs(filepath.Join(repository, path))
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			truth := onNode(test, absolute)
			if truth.exitCode != 0 || len(truth.stderr) != 0 || len(truth.stdout) == 0 {
				test.Fatalf("Node: %+v", truth)
			}
			_, stopped := lowered(test, absolute)
			var notYet *lower.NotYet
			var refused *lower.Refused
			if !errors.As(stopped, &notYet) && (name != "filter" || !errors.As(stopped, &refused)) {
				test.Fatalf("want a NotYet, got %v", stopped)
			}
			expected, readError := os.ReadFile(strings.TrimSuffix(absolute, ".ts") + ".stopped")
			if readError != nil {
				test.Fatal(readError)
			}
			got := strings.ReplaceAll(stopped.Error(), strings.TrimSuffix(absolute, path), "")
			if got != strings.TrimSpace(string(expected)) {
				test.Fatalf("diagnostic: got %q, want %q", got, strings.TrimSpace(string(expected)))
			}
		})
	}
}
