package oracle

import (
	"os"
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

// Array.isArray distinguishes this narrowed array from the map arm before casting.
func TestNarrowedUnionObjectTagPassing(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/reland_refused/narrowed_union_object_tag.a"))
	if err != nil {
		t.Fatal(err)
	}
	checkNarrowedArrayTag(t, path, run{stdout: []byte("1\n")}, run{stdout: []byte("1\n")})
}

func TestNarrowedUnionObjectTagMisfit(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "narrowed_union_object_tag_misfit.a")
	source := `function first(): number[] | Map<string, number> { return [1]; }
let shared: number[] | Map<string, number> = first();
function change(): void { shared = new Map<string, number>(); }
shared = [1];
change();
console.log("" + shared.length);
`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	checkNarrowedArrayTag(t, path, run{stdout: []byte("undefined\n")}, run{exitCode: 70, stderr: []byte("adamic: panic: union member where the checker narrowed it away: a call since the narrowing put it back\n")})
}

func checkNarrowedArrayTag(t *testing.T, path string, nodeWant, want run) {
	t.Helper()
	node := onNode(t, path)
	if diff := disagreement(nodeWant, node); diff != "" {
		t.Fatalf("Node: %s: %#v", diff, node)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	native, _ := nativelyUncached(t, program)
	for backend, got := range map[string]run{"native-sanitized": native, "native-release": releasedUncached(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
		if diff := disagreement(want, got); diff != "" {
			t.Errorf("%s: %s: exit %d stdout %q stderr %q", backend, diff, got.exitCode, got.stdout, got.stderr)
		}
		t.Logf("%s: exit %d stdout %q stderr %q", backend, got.exitCode, got.stdout, got.stderr)
	}
}
