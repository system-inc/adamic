package oracle

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestCheckedViewFlagDowncast(t *testing.T) {
	t.Run("tsc source", func(t *testing.T) {
		path, err := filepath.Abs(filepath.Join(repository, "stage3/fixtures/assertions/13_flag_downcast.a"))
		if err != nil {
			t.Fatal(err)
		}
		program, err := lowered(t, path)
		if err != nil {
			t.Fatal(err)
		}
		want := onNode(t, path)
		if difference := disagreement(run{stdout: []byte("8\n0\n")}, want); difference != "" {
			t.Fatal("source Node: " + difference)
		}
		sanitized, binary := nativelyUncached(t, program)
		for backend, got := range map[string]run{"native sanitized": sanitized, "native release": releasedUncached(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
			if difference := disagreement(want, got); difference != "" {
				t.Errorf("%s: %s", backend, difference)
			}
		}
		if report := leaks(t, program, binary); report != "" {
			t.Fatal(report)
		}
	})
	t.Run("wrong scalar and check removal", func(t *testing.T) {
		program, path := interfaceFixture(t, "lane1/flag-downcast-wrong")
		if difference := disagreement(run{stdout: []byte("true\n")}, onNode(t, path)); difference != "" {
			t.Fatal("source Node: " + difference)
		}
		want := run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: (symbol as TransientSymbol).links.checkFlags is not a number; expected number, found boolean\n")}
		sanitized, _ := nativelyUncached(t, program)
		for backend, got := range map[string]run{"native sanitized": sanitized, "native release": releasedUncached(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
			if difference := disagreement(want, got); difference != "" {
				t.Fatalf("%s check: %s; got %#v", backend, difference, got)
			}
		}
		changed := false
		mutate := func(value ir.Expression) ir.Expression {
			if property, ok := value.(ir.Property); ok && property.Name == "checkFlags" && property.View != "" {
				changed = true
				property.View = ""
				return property
			}
			return value
		}
		mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), mutate)
		mutateStringExpressions(reflect.ValueOf(&program.Functions).Elem(), mutate)
		if !changed {
			t.Fatal("field-check mutant changed no read")
		}
		for backend, got := range map[string]run{"native release": releasedUncached(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
			if strings.Contains(string(got.stderr), "Sanitizer") {
				t.Fatal("semantic pin must catch the mutant")
			}
			if disagreement(want, got) == "" {
				t.Fatalf("%s field-check mutant survived", backend)
			}
			t.Logf("%s field-check mutant caught: exit %d stdout %q stderr %q", backend, got.exitCode, got.stdout, got.stderr)
		}
	})
}
