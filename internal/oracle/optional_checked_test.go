package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/lower"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOptionalCheckedReads(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join(repository, "internal/lower/testdata/optional_widening/widening*.a"))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			path, err := filepath.Abs(path)
			if err != nil {
				t.Fatal(err)
			}
			observation, err := os.ReadFile(strings.TrimSuffix(path, ".a") + ".expected")
			if err != nil {
				t.Fatal(err)
			}
			if difference := disagreement(run{stdout: observation}, onNode(t, path)); difference != "" {
				t.Fatal("Node: " + difference)
			}
			program, err := lowered(t, path)
			if strings.HasSuffix(path, "class_expression.a") {
				var pending *lower.NotYet
				if !errors.As(err, &pending) || !strings.Contains(err.Error(), "a ClassExpression yet") {
					t.Fatalf("want explicit class expression boundary, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			expression := "wider.y"
			if strings.Contains(path, "_array") || strings.Contains(path, "_parameter") {
				expression = "point.y"
			}
			want := run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: " + expression + " is not a number | undefined; expected number | undefined, found boolean\n")}
			actual, _ := nativelyUncached(t, program)
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := viewReadDisagreement(want, got, program); difference != "" {
					t.Fatalf("%s; got %#v", difference, got)
				}
			}
		})
	}
}

func TestOptionalCheckedReadGood(t *testing.T) {
	for _, probe := range []struct{ name, source, output string }{
		{"absent", "const source: { x: number } = { x: 1 }; const wider: { x: number; y?: number } = source; console.log(`${wider.y === undefined}`);", "true\n"},
		{"present", "const actual = { x: 1, y: 7 }; const source: { x: number } = actual; const wider: { x: number; y?: number } = source; console.log(`${wider.y ?? 0}`);", "7\n"},
		{"undefined", "const actual: { x: number; y: number | undefined } = { x: 1, y: undefined }; const source: { x: number } = actual; const wider: { x: number; y?: number } = source; console.log(`${wider.y === undefined}`);", "true\n"},
		{"boolean", "const actual = { x: 1, y: false }; const source: { x: number } = actual; const wider: { x: number; y?: boolean } = source; console.log(`${wider.y ?? true}`);", "false\n"},
		{"boolean_absent", "const source: { x: number } = { x: 1 }; const wider: { x: number; y?: boolean } = source; console.log(`${wider.y === undefined}`);", "true\n"},
		{"object", "const actual = { x: 1, y: { ready: true } }; const source: { x: number } = actual; const wider: { x: number; y?: { ready: boolean } } = source; console.log(`${wider.y?.ready ?? false}`);", "true\n"},
		{"string", "const actual = { x: 1, y: 'ok' }; const source: { x: number } = actual; const wider: { x: number; y?: string } = source; console.log(wider.y ?? 'absent');", "ok\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), probe.name+".a")
			if err := os.WriteFile(path, []byte(probe.source), 0600); err != nil {
				t.Fatal(err)
			}
			want := run{stdout: []byte(probe.output)}
			if difference := disagreement(want, onNode(t, path)); difference != "" {
				t.Fatal("Node: " + difference)
			}
			program, err := lowered(t, path)
			if strings.HasSuffix(path, "class_expression.a") {
				var pending *lower.NotYet
				if !errors.As(err, &pending) || !strings.Contains(err.Error(), "a ClassExpression yet") {
					t.Fatalf("want explicit class expression boundary, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			actual, _ := nativelyUncached(t, program)
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := viewReadDisagreement(want, got, program); difference != "" {
					t.Fatal(difference)
				}
			}
		})
	}
}

func TestOptionalClassReads(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join(repository, "internal/lower/testdata/optional_widening/class_*.a"))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		if strings.HasSuffix(path, "class_field.a") {
			continue
		}
		t.Run(filepath.Base(path), func(t *testing.T) {
			path, err := filepath.Abs(path)
			if err != nil {
				t.Fatal(err)
			}
			if difference := disagreement(run{stdout: []byte("string\n")}, onNode(t, path)); difference != "" {
				t.Fatal("Node: " + difference)
			}
			program, err := lowered(t, path)
			if strings.HasSuffix(path, "class_expression.a") {
				var pending *lower.NotYet
				if !errors.As(err, &pending) || !strings.Contains(err.Error(), "a ClassExpression yet") {
					t.Fatalf("want explicit class expression boundary, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			want := run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: wider.y is not a number | undefined; expected number | undefined, found string\n")}
			actual, _ := nativelyUncached(t, program)
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := viewReadDisagreement(want, got, program); difference != "" {
					t.Fatal(difference)
				}
			}
		})
	}
}

func TestOptionalNullAndLiteralReads(t *testing.T) {
	for _, probe := range []struct{ name, source, node, diagnostic string }{
		{"null", "const raw={x:1,y:null}; const hidden:{x:number}=raw; const wider:{x:number;y?:number}=hidden; console.log(typeof wider.y);", "object\n", "field read failed: wider.y is not a number | undefined; expected number | undefined, found null"},
		{"literal", "const raw={x:1,y:1}; const hidden:{x:number}=raw; const wider:{x:number;y?:2}=hidden; console.log(`${wider.y}`);", "1\n", "field read failed: wider.y expected 2 | undefined, found number 1"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), probe.name+".a")
			if err := os.WriteFile(path, []byte(probe.source), 0600); err != nil {
				t.Fatal(err)
			}
			if difference := disagreement(run{stdout: []byte(probe.node)}, onNode(t, path)); difference != "" {
				t.Fatal("Node: " + difference)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			want := run{exitCode: 70, stderr: []byte("adamic: panic: " + probe.diagnostic + "\n")}
			actual, _ := nativelyUncached(t, program)
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := viewReadDisagreement(want, got, program); difference != "" {
					t.Fatalf("%s; got %#v", difference, got)
				}
			}
		})
	}
}
