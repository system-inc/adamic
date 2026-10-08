package oracle

import (
	"os"
	"path/filepath"
	"testing"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{path: "internal/oracle/testdata/never_reached.a", lowers: true, checked: true})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{path: "internal/oracle/testdata/never_call.a", lowers: true})
}

// A call invalidates the checker's narrowing. Node proves this branch is reached;
// Adamic must stop at the never initializer before printing continued.
func TestNeverReached(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/never_reached.a"))
	if err != nil {
		t.Fatal(err)
	}
	source := onNode(t, path)
	if source.exitCode != 0 || string(source.stdout) != "continued\n" || len(source.stderr) != 0 {
		t.Fatalf("Node: %#v", source)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := run{exitCode: 70, stderr: []byte("adamic: panic: unreachable expression value at never_reached.a:5:31\n")}
	native, _ := nativelyUncached(t, program)
	for _, got := range []run{native, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Errorf("%s: exit %d, stdout %q, stderr %q", difference, got.exitCode, got.stdout, got.stderr)
		}
	}
}

// Property and indexed reads can also retain a narrowing across a mutating call.
func TestNeverReadKinds(t *testing.T) {
	t.Parallel()
	for _, test := range []struct{ name, source, expression string }{
		{"property", "const holder: { value: boolean } = { value: false };\nfunction change(): void { holder.value = true; }\nif (holder.value === false) {\n    change();\n    if (holder.value !== false) {\n        const impossible: never = holder.value;\n        console.log('continued');\n    }\n}\n", "holder.value"},
		{"element", "const values: boolean[] = [false];\nfunction change(): void { values[0] = true; }\nif (values[0] === false) {\n    change();\n    if (values[0] !== false) {\n        const impossible: never = values[0];\n        console.log('continued');\n    }\n}\n", "values[0]"},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), test.name+".a")
			if err := os.WriteFile(path, []byte(test.source), 0600); err != nil {
				t.Fatal(err)
			}
			source := onNode(t, path)
			if source.exitCode != 0 || string(source.stdout) != "continued\n" || len(source.stderr) != 0 {
				t.Fatalf("Node: %#v", source)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			want := run{exitCode: 70, stderr: []byte("adamic: panic: unreachable expression " + test.expression + " at " + test.name + ".a:6:35\n")}
			native, _ := nativelyUncached(t, program)
			for _, got := range []run{native, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Errorf("%s: exit %d, stdout %q, stderr %q", difference, got.exitCode, got.stdout, got.stderr)
				}
			}
		})
	}
}

func TestNeverCondition(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "condition.a")
	source := "let flag: boolean = false;\nfunction change(): void { flag = true; }\nchange();\nif (flag) {\n    if (flag) { console.log('continued'); }\n}\n"
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	observed := onNode(t, path)
	if observed.exitCode != 0 || string(observed.stdout) != "continued\n" || len(observed.stderr) != 0 {
		t.Fatalf("Node: %#v", observed)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := run{exitCode: 70, stderr: []byte("adamic: panic: unreachable expression flag at condition.a:5:9\n")}
	actual, _ := nativelyUncached(t, program)
	for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Errorf("%s: exit %d, stdout %q, stderr %q", difference, got.exitCode, got.stdout, got.stderr)
		}
	}
}

func TestNeverFieldConditions(t *testing.T) {
	t.Parallel()
	for _, test := range []struct{ name, declaration, write, read string }{
		{"property_condition", "const holder: { value: boolean } = { value: false };", "holder.value = true", "holder.value"},
		{"element_condition", "const values: boolean[] = [false];", "values[0] = true", "values[0]"},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), test.name+".a")
			source := test.declaration + "\nfunction change(): void { " + test.write + "; }\nif (" + test.read + " === false) {\n    change();\n    if (" + test.read + " !== false) {\n        if (" + test.read + ") { console.log('continued'); }\n    }\n}\n"
			if err := os.WriteFile(path, []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
			observed := onNode(t, path)
			if observed.exitCode != 0 || string(observed.stdout) != "continued\n" || len(observed.stderr) != 0 {
				t.Fatalf("Node: %#v", observed)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			want := run{exitCode: 70, stderr: []byte("adamic: panic: unreachable expression " + test.read + " at " + test.name + ".a:6:13\n")}
			actual, _ := nativelyUncached(t, program)
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Errorf("%s: exit %d, stdout %q, stderr %q", difference, got.exitCode, got.stdout, got.stderr)
				}
			}
		})
	}
}
