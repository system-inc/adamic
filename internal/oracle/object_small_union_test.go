package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"path/filepath"
	"testing"
)

func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/object_small_union_stale.a", true, true})
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/object_small_union.a", true, false})
}

func TestObjectSmallUnionFieldCheck(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ path, output string }{

		{"object_small_union_stale.a", "after toText: wordswordswords1\n"},
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
				if function.Name == "narrowed_union_field" {
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
			for name, got := range map[string]run{"native": mutant, "JavaScript": onJavaScriptBackend(t, program)} {
				if disagreement(want, got) == "" {
					t.Fatalf("%s unchecked field member mutant survived", name)
				}
				t.Logf("%s unchecked field member mutant caught: exit %d, stdout %q, stderr %q", name, got.exitCode, got.stdout, got.stderr)
			}
		})
	}
}
