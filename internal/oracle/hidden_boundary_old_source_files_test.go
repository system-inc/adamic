package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/hidden_boundary_old_source_files.a", true, false})
}

// The short closure witness already lowers on the area base. It must still read
// the enclosing cell, rather than a fabricated empty array. The opt-in mutant
// uses the same Node comparison and compiles normally in both backends.
func TestHiddenBoundaryOldSourceFilesCapture(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/hidden_boundary_old_source_files.a"))
	if err != nil {
		t.Fatal(err)
	}
	expected := onNode(t, path)
	if expected.exitCode != 0 || string(expected.stdout) != "7\n" || len(expected.stderr) != 0 {
		t.Fatalf("source Node: %+v", expected)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	if os.Getenv("ADAMIC_MUTANT_HIDDEN_15_CAPTURE") == "1" {
		changed := 0
		mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), func(value ir.Expression) ir.Expression {
			read, ok := value.(ir.Read)
			if ok && program.Locals[read.Local].Name == "oldSourceFiles" && program.Locals[read.Local].Captured {
				changed++
				return ir.ArrayLiteral{Element: ir.Number}
			}
			return value
		})
		if changed != 1 {
			t.Fatalf("want one captured array read, changed %d", changed)
		}
	}
	actual, binary := nativelyUncached(t, program)
	if report := leaksUncached(t, program, binary); report != "" {
		t.Fatal(report)
	}
	for name, got := range map[string]run{"native": actual, "javascript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 0 || len(got.stderr) != 0 {
			t.Errorf("%s did not finish cleanly: %+v", name, got)
			continue
		}
		if difference := disagreement(expected, got); difference != "" {
			t.Errorf("%s: %s: Node %q, compiled %q", name, difference, expected.stdout, got.stdout)
		}
	}
}

// This preceding initializer is the actual dependency behind the census's
// rolled-back oldSourceFiles binding. Retain the full source and pinned stop;
// treating a failed declaration as an initialized empty array would be unsound.
func TestHiddenBoundaryOldSourceFilesStaticsStop(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/refusals/hidden_boundary_old_source_files_statics.a"))
	if err != nil {
		t.Fatal(err)
	}
	expected := onNode(t, path)
	if expected.exitCode != 0 || string(expected.stdout) != "1:7\n" || len(expected.stderr) != 0 {
		t.Fatalf("source Node: %+v", expected)
	}
	_, err = lowered(t, path)
	var stopped *lower.NotYet
	if !errors.As(err, &stopped) || !strings.Contains(err.Error(), "a method call through a structural signature in a program with statics") {
		t.Fatalf("want the preceding structural/statics stop, got %v", err)
	}
}
