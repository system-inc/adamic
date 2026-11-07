package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// Register here to keep this unit out of the shared oracle_test.go.
func init() {
	for _, path := range []string{
		"e4eec87_f2_union_narrow_call.a",
		"e4eec87_f2b_union_narrow_number.a",
		"narrowed_union_valid.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{
			"internal/oracle/testdata/" + path, true, path != "narrowed_union_valid.a",
		})
	}
}

// Source Node continues with the new member. Adamic must reject that member before the cast,
// just as it rejects undefined restored after a narrowing.
func TestNarrowedUnionMemberCheck(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ path, output string }{
		{"e4eec87_f2_union_narrow_call.a", "after toNumber: undefined\n"},
		{"e4eec87_f2b_union_narrow_number.a", "after toText: wordswordswords1\n"},
	} {
		test.Run(probe.path, func(test *testing.T) {
			test.Parallel()
			path, absolutePathError := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", probe.path))
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			if node := onNode(test, path); disagreement(run{stdout: []byte(probe.output)}, node) != "" {
				test.Fatalf("source Node: exit %d, stdout %q, stderr %q", node.exitCode, node.stdout, node.stderr)
			}
			program, loweredError := lowered(test, path)
			if loweredError != nil {
				test.Fatal(loweredError)
			}
			want := run{exitCode: 70, stderr: []byte(
				"adamic: panic: union member where the checker narrowed it away: a call since the narrowing put it back\n",
			)}
			native, _ := natively(test, program)
			for name, got := range map[string]run{
				"native":     native,
				"release":    released(test, program),
				"JavaScript": onJavaScriptBackend(test, program),
			} {
				if difference := disagreement(want, got); difference != "" {
					test.Fatalf(
						"%s: %s: exit %d, stdout %q, stderr %q",
						name,
						difference,
						got.exitCode,
						got.stdout,
						got.stderr,
					)
				}
			}
			// Restore the unchecked cast, without changing its input or making invalid C.
			removed := 0
			for index := range program.Functions {
				function := &program.Functions[index]
				if function.Name == "narrowed_union_member" {
					if _, check := function.Body[0].(ir.If); check {
						function.Body = function.Body[1:]
						removed++
					}
				}
			}
			if removed == 0 {
				test.Fatal("mutant removed no check")
			}
			mutant, _ := natively(test, program)
			if disagreement(want, mutant) == "" {
				test.Fatal("unchecked member mutant survived")
			}
			test.Logf(
				"unchecked member mutant caught: exit %d, stdout %q, stderr %q",
				mutant.exitCode,
				mutant.stdout,
				mutant.stderr,
			)
		})
	}
}

// Ambiguous object tags need a stronger IR predicate before they can be cast safely.
func TestNarrowedUnionObjectTagRefusal(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/reland_refused/narrowed_union_object_tag.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	if node := onNode(test, path); disagreement(run{stdout: []byte("1\n")}, node) != "" {
		test.Fatalf("source Node: %#v", node)
	}
	_, loweredError := lowered(test, path)
	if loweredError == nil ||
		!strings.Contains(loweredError.Error(), "narrowed_union_object_tag.a") ||
		!strings.Contains(
			loweredError.Error(),
			"stage 0 can't lower a narrowed union member whose object tag cannot be checked with typeof; keep differently held object kinds in separately typed variables yet",
		) {
		test.Fatalf("want refusal naming the path and fix, got %v", loweredError)
	}
}
