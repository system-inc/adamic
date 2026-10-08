package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"path/filepath"
	"testing"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/narrowed_array_undefined.a", true, false})
}
func TestNarrowedArrayStaleMemberChecked(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/lower/testdata/narrowed_array_stale.a"))
	if err != nil {
		t.Fatal(err)
	}
	if got := onNode(t, path); disagreement(run{stdout: []byte("14\n")}, got) != "" {
		t.Fatalf("Node: %+v", got)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := run{exitCode: 70, stderr: []byte("adamic: panic: union member where the checker narrowed it away: a call since the narrowing put it back\n")}
	native, _ := natively(t, program)
	for name, got := range map[string]run{"native": native, "JavaScript": onJavaScriptBackend(t, program)} {
		if diff := disagreement(want, got); diff != "" {
			t.Fatalf("%s: %s", name, diff)
		}
	}
	removed := 0
	for index := range program.Functions {
		function := &program.Functions[index]
		if function.Name == "narrowed_union_member" {
			if _, ok := function.Body[0].(ir.If); ok {
				function.Body = function.Body[1:]
				removed++
			}
		}
	}
	if removed == 0 {
		t.Fatal("mutant removed no check")
	}
	native, _ = natively(t, program)
	for name, got := range map[string]run{"native": native, "JavaScript": onJavaScriptBackend(t, program)} {
		if disagreement(want, got) == "" {
			t.Fatalf("%s unchecked-array mutant survived", name)
		}
		t.Logf("%s unchecked-array mutant caught: exit %d stdout %q stderr %q", name, got.exitCode, got.stdout, got.stderr)
	}
}
