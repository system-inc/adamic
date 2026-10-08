package oracle

import (
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// Register here to keep this unit out of the shared oracle_test.go.
func init() {
	for _, path := range []string{"e4eec87_f2_union_narrow_call.a", "e4eec87_f2b_union_narrow_number.a", "narrowed_union_valid.a"} {
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
func TestNarrowedUnionMemberCheck(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ path, output string }{
		{"e4eec87_f2_union_narrow_call.a", "after toNumber: undefined\n"},
		{"e4eec87_f2b_union_narrow_number.a", "after toText: wordswordswords1\n"},
	} {
		t.Run(probe.path, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", probe.path))
			if err != nil {
				t.Fatal(err)
			}
			if node := onNode(t, path); disagreement(run{stdout: []byte(probe.output)}, node) != "" {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q", node.exitCode, node.stdout, node.stderr)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			want := run{exitCode: 70, stderr: []byte("adamic: panic: union member where the checker narrowed it away: a call since the narrowing put it back\n")}
			native, _ := natively(t, program)
			for name, got := range map[string]run{"native": native, "release": released(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatalf("%s: %s: exit %d, stdout %q, stderr %q", name, difference, got.exitCode, got.stdout, got.stderr)
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
				t.Fatal("mutant removed no check")
			}
			mutant, _ := natively(t, program)
			if disagreement(want, mutant) == "" {
				t.Fatal("unchecked member mutant survived")
			}
			t.Logf("unchecked member mutant caught: exit %d, stdout %q, stderr %q", mutant.exitCode, mutant.stdout, mutant.stderr)
		})
	}
}

// The runtime kind check now distinguishes the two layouts.
func TestNarrowedUnionObjectTagLowers(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/reland_refused/narrowed_union_object_tag.a"))
	if err != nil {
		t.Fatal(err)
	}
	if node := onNode(t, path); disagreement(run{stdout: []byte("1\n")}, node) != "" {
		t.Fatalf("source Node: %#v", node)
	}
	if _, err := lowered(t, path); err != nil {
		t.Fatal(err)
	}
}
