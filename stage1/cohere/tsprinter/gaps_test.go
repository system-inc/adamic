package tsprinter

import (
	"path/filepath"
	"testing"
)

func TestCompilerGaps(t *testing.T) {
	t.Parallel()
	t.Run("defaultSort", func(t *testing.T) {
		t.Parallel()
		path, err := filepath.Abs("gaps/defaultSort.ts")
		if err != nil {
			t.Fatal(err)
		}
		expected := onNode(t, path)
		if expected.exitCode != 0 || len(expected.stderr) != 0 || string(expected.stdout) != "im\n" {
			t.Fatalf("Node: %+v", expected)
		}
		program := lowered(t, path)
		actual, binary := natively(t, program)
		for _, side := range []run{actual, onJavaScriptBackend(t, program)} {
			if side.exitCode != 0 || len(side.stderr) != 0 || string(side.stdout) != string(expected.stdout) {
				t.Fatalf("closed sort gap differs from Node: %+v", side)
			}
		}
		if report := leaks(t, program, binary); report != "" {
			t.Fatal(report)
		}
	})
}

// prefixUpdateValue.ts lowers on compiler/area-stack (views slice 1, Oct 8): a numeric
// PrefixUnaryExpression whose value is read. Held to Node on native, the JavaScript backend and
// the leak check. The port's separate decrement statement still stands; retiring it is cohere's.
// Not parallel: native.Build writes the shared user cache directory adamic/runtime (and adamic/units when split builds are enabled).
func TestClosedPrefixUpdateValueGap(t *testing.T) {
	path, err := filepath.Abs("gaps/prefixUpdateValue.ts")
	if err != nil {
		t.Fatal(err)
	}
	program := lowered(t, path)
	native, binary := natively(t, program)
	for _, result := range []run{onNode(t, path), native, onJavaScriptBackend(t, program)} {
		if result.exitCode != 0 || len(result.stderr) != 0 || string(result.stdout) != "2\n" {
			t.Fatalf("prefix update value: %+v", result)
		}
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
}

// Not parallel: native.Build writes the shared user cache directory adamic/runtime (and adamic/units when split builds are enabled).
func TestNumberConstructor(t *testing.T) {
	path, err := filepath.Abs("testdata/numberConstructor.ts")
	if err != nil {
		t.Fatal(err)
	}
	program := lowered(t, path)
	native, binary := natively(t, program)
	for _, result := range []run{onNode(t, path), native, onJavaScriptBackend(t, program)} {
		if result.exitCode != 0 || len(result.stderr) != 0 || string(result.stdout) != "17\n" {
			t.Fatalf("Number constructor: %+v", result)
		}
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
}
