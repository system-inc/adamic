package oracle

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckedViewDictionaryPropagation(t *testing.T) {
	for _, mode := range []string{"helper", "generic", "callback", "field"} {
		for _, suffix := range []string{"good", "wrong"} {
			t.Run(mode+"-"+suffix, func(t *testing.T) {
				program, path := interfaceFixture(t, "dictionaries/source/flow-"+mode+"-"+suffix)
				truth := onNode(t, path)
				expected := "name\n"
				if suffix == "wrong" {
					expected = "42\n"
				}
				if truth.exitCode != 0 || string(truth.stdout) != expected {
					t.Fatalf("source Node: %#v", truth)
				}
				sanitized, binary := nativelyUncached(t, program)
				for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if suffix == "good" {
						if difference := disagreement(truth, got); difference != "" {
							t.Fatal(difference)
						}
					} else if got.exitCode != 70 || viewReadDiagnosticMismatch("adamic: panic: field read failed: entry.name is not a string; expected string, found number\n", got.stderr, program) {
						t.Fatalf("transitive named check: %#v", got)
					}
				}
				if suffix == "good" {
					if report := leaks(t, program, binary); report != "" {
						t.Fatal(report)
					}
				}
			})
		}
	}
}

func TestCheckedViewDictionaryArrays(t *testing.T) {
	for _, test := range []struct{ name, stdout, field, expected, found string }{
		{"paths-producer-good", "name\n", "", "", ""},
		{"paths-good", "name\n", "", "", ""},
		{"paths-missing", "missing\n", "", "", ""},
		{"paths-container-wrong", "number\n", "view.paths", "MapLike<string[]> | undefined", "number"},
		{"paths-wrong", "number\n", "table['item']", "string[] | undefined", "number"},
		{"paths-element-wrong", "42\n", "values[0]", "string", "number"},
		{"paths-key-missing", "undefined\n", "", "", ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			program, path := interfaceFixture(t, "dictionaries/source/"+test.name)
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != test.stdout {
				t.Fatalf("source Node: %#v", truth)
			}
			sanitized, binary := nativelyUncached(t, program)
			for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if test.field == "" {
					if difference := disagreement(truth, got); difference != "" {
						t.Fatal(difference)
					}
				} else {
					want := fmt.Sprintf("adamic: panic: field read failed: %s; expected %s, found %s\n", test.field, test.expected, test.found)
					if test.name == "paths-element-wrong" {
						want = fmt.Sprintf("adamic: panic: element read failed: %s expected %s, found %s\n", test.field, test.expected, test.found)
					}
					if test.name == "paths-container-wrong" {
						want = fmt.Sprintf("adamic: panic: field read failed: %s matches no member of %s; expected %s, found %s\n", test.field, test.expected, test.expected, test.found)
					}
					if got.exitCode != 70 || viewReadDiagnosticMismatch(want, got.stderr, program) {
						t.Fatalf("pinned array dictionary check: %#v; want %q", got, want)
					}
					t.Logf("pinned exit 70: %s", got.stderr)
				}
			}
			if test.field == "" {
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
		})
	}
}

func TestCheckedViewDictionaryRoot(t *testing.T) {
	for _, name := range []string{"good", "wrong"} {
		t.Run(name, func(t *testing.T) {
			program, path := interfaceFixture(t, "dictionaries/source/root-"+name)
			truth := onNode(t, path)
			expected := "name\nundefined\n"
			if name == "wrong" {
				expected = "42\n"
			}
			if truth.exitCode != 0 || string(truth.stdout) != expected {
				t.Fatalf("source Node: %#v", truth)
			}
			sanitized, binary := nativelyUncached(t, program)
			for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if name == "good" {
					if difference := disagreement(truth, got); difference != "" {
						t.Fatal(difference)
					}
				} else if got.exitCode != 70 || viewReadDiagnosticMismatch("adamic: panic: field read failed: view['item']; expected string | undefined, found number\n", got.stderr, program) {
					t.Fatalf("root dictionary check: %#v", got)
				}
			}
			if name == "good" {
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
		})
	}
}

func TestCheckedViewDictionaryWriteGuard(t *testing.T) {
	path := filepath.Join(repository, "stage3/interface-downcasts/dictionaries/source/write-guard.a")
	path, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "42\n" {
		t.Fatalf("source Node: %#v", truth)
	}
	_, err = lowered(t, path)
	if err == nil || !strings.Contains(err.Error(), "dictionary write without a producer storage certificate") {
		t.Fatalf("uncertified write was admitted: %v", err)
	}
}

func TestCheckedViewDictionaryCompilerPaths(t *testing.T) {
	for _, test := range []struct{ name, stdout, field, expected, found string }{
		{"compiler-paths-good", "name\n", "", "", ""},
		{"compiler-paths-missing", "missing\n", "", "", ""},
		{"compiler-paths-container-wrong", "number\n", "view.paths", "MapLike<string[]> | undefined", "number"},
		{"compiler-paths-wrong", "number\n", "table['item']", "string[] | undefined", "number"},
		{"compiler-paths-element-wrong", "42\n", "values[0]", "string", "number"},
		{"compiler-paths-key-missing", "undefined\n", "", "", ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			program, path := interfaceFixture(t, "dictionaries/source/"+test.name)
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != test.stdout {
				t.Fatalf("source Node: %#v", truth)
			}
			sanitized, binary := nativelyUncached(t, program)
			for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if test.field == "" {
					if difference := disagreement(truth, got); difference != "" {
						t.Fatal(difference)
					}
				} else {
					want := fmt.Sprintf("adamic: panic: field read failed: %s; expected %s, found %s\n", test.field, test.expected, test.found)
					if test.name == "compiler-paths-element-wrong" {
						want = fmt.Sprintf("adamic: panic: element read failed: %s expected %s, found %s\n", test.field, test.expected, test.found)
					}

					if got.exitCode != 70 || viewReadDiagnosticMismatch(want, got.stderr, program) {
						t.Fatalf("pinned array dictionary check: %#v; want %q", got, want)
					}
					t.Logf("pinned exit 70: %s", got.stderr)
				}
			}
			if test.field == "" {
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
		})
	}
}

func TestCheckedViewDictionaryOptionalPaths(t *testing.T) {
	for _, test := range []struct{ name, stdout, field, expected, found string }{
		{"optional-paths-once", "missing\nname\n2\n", "", "", ""},
		{"optional-paths-receiver-missing", "missing\n", "", "", ""},
		{"optional-paths-good", "name\n", "", "", ""},
		{"optional-paths-missing", "missing\n", "", "", ""},
		{"optional-paths-container-wrong", "number\n", "view?.paths", "MapLike<string[]> | undefined", "number"},
		{"optional-paths-wrong", "number\n", "table['item']", "string[] | undefined", "number"},
		{"optional-paths-element-wrong", "42\n", "values[0]", "string", "number"},
		{"optional-paths-key-missing", "undefined\n", "", "", ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			program, path := interfaceFixture(t, "dictionaries/source/"+test.name)
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != test.stdout {
				t.Fatalf("source Node: %#v", truth)
			}
			sanitized, binary := nativelyUncached(t, program)
			for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if test.field == "" {
					if difference := disagreement(truth, got); difference != "" {
						t.Fatal(difference)
					}
				} else {
					want := fmt.Sprintf("adamic: panic: field read failed: %s; expected %s, found %s\n", test.field, test.expected, test.found)
					if test.name == "optional-paths-element-wrong" {
						want = fmt.Sprintf("adamic: panic: element read failed: %s expected %s, found %s\n", test.field, test.expected, test.found)
					}

					if got.exitCode != 70 || viewReadDiagnosticMismatch(want, got.stderr, program) {
						t.Fatalf("pinned array dictionary check: %#v; want %q", got, want)
					}
					t.Logf("pinned exit 70: %s", got.stderr)
				}
			}
			if test.field == "" {
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
		})
	}
}
