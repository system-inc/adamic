package oracle

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestCheckedViewFlagDowncast(test *testing.T) {
	test.Parallel()
	test.Run("tsc source", func(test *testing.T) {
		test.Parallel()
		path, absolutePathError := filepath.Abs(
			filepath.Join(repository, "stage3/fixtures/assertions/13_flag_downcast.ts"),
		)
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		program, loweredError := lowered(test, path)
		if loweredError != nil {
			test.Fatal(loweredError)
		}
		want := onNode(test, path)
		if difference := disagreement(run{stdout: []byte("8\n0\n")}, want); difference != "" {
			test.Fatal("source Node: " + difference)
		}
		sanitized, binary := nativelyUncached(test, program)
		for backend, got := range map[string]run{
			"native sanitized": sanitized,
			"native release":   releasedUncached(test, program),
			"JavaScript":       onJavaScriptBackend(test, program),
		} {
			if difference := disagreement(want, got); difference != "" {
				test.Errorf("%s: %s", backend, difference)
			}
		}
		if report := leaks(test, program, binary); report != "" {
			test.Fatal(report)
		}
	})
	test.Run("wrong scalar and check removal", func(test *testing.T) {
		test.Parallel()
		program, path := interfaceFixture(test, "lane1/flag-downcast-wrong")
		if difference := disagreement(run{stdout: []byte("true\n")}, onNode(test, path)); difference != "" {
			test.Fatal("source Node: " + difference)
		}
		want := run{exitCode: 70, stderr: []byte(
			"adamic: panic: field read failed: (symbol as TransientSymbol).links.checkFlags is not a number; expected number, found boolean\n",
		)}
		sanitized, _ := nativelyUncached(test, program)
		for backend, got := range map[string]run{
			"native sanitized": sanitized,
			"native release":   releasedUncached(test, program),
			"JavaScript":       onJavaScriptBackend(test, program),
		} {
			if difference := disagreement(want, got); difference != "" {
				test.Fatalf("%s check: %s; got %#v", backend, difference, got)
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
			test.Fatal("field-check mutant changed no read")
		}
		for backend, got := range map[string]run{
			"native release": releasedUncached(test, program),
			"JavaScript":     onJavaScriptBackend(test, program),
		} {
			if strings.Contains(string(got.stderr), "Sanitizer") {
				test.Fatal("semantic pin must catch the mutant")
			}
			if disagreement(want, got) == "" {
				test.Fatalf("%s field-check mutant survived", backend)
			}
			test.Logf(
				"%s field-check mutant caught: exit %d stdout %q stderr %q",
				backend,
				got.exitCode,
				got.stdout,
				got.stderr,
			)
		}
	})
}
