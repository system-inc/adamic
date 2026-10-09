package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

func TestAdaptOptionalPropertiesPreservesRuntimeAndDisk(t *testing.T) {
	t.Parallel()
	const name = "testdata/optional/main.ts"
	source, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	_, before := load.Load([]string{name})
	if optionalDiagnosticCount(before) != 5 {
		t.Fatalf("want five exact-optional diagnostics, got %v", before)
	}
	overlay, rewrites, err := adaptations([]string{name}, before)
	if err != nil {
		t.Fatal(err)
	}
	absolute, _ := filepath.Abs(name)
	adapted := overlay[absolute]
	if len(rewrites) != 1 || rewrites[0].Rewrite != optionalRewrite || rewrites[0].Removed != 5 || rewrites[0].DeclarationsChanged != 3 {
		t.Fatalf("adaptations = %#v", rewrites)
	}
	if _, err := load.LoadOverlay([]string{name}, overlay); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(adapted, "callback?: (() => string) | undefined") {
		t.Fatalf("function union precedence not preserved: %s", adapted)
	}
	want := []byte("p\ntrue true p,callback\ntrue true\ntrue true\n")
	for label, text := range map[string]string{"original": string(source), "adapted": adapted} {
		actual := nodeOptionalSource(t, text)
		if !bytes.Equal(actual, want) {
			t.Fatalf("%s runtime output = %q, want %q", label, actual, want)
		}
	}
	after, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(source, after) {
		t.Fatal("adaptation changed source on disk")
	}
	_, afterCheck := load.LoadOverlay([]string{name}, overlay)
	next, changed, err := optionalAdaptations([]string{name}, overlay, afterCheck)
	if err != nil || changed != 0 || next[absolute] != adapted {
		t.Fatalf("not idempotent: changed=%d, err=%v", changed, err)
	}
	measured, err := measure("testdata/optional", true)
	if err != nil {
		t.Fatal(err)
	}
	if measured.FilesReachingLowering != 1 {
		t.Fatalf("lowering entries = %d, reasons %#v", measured.FilesReachingLowering, measured.Reasons)
	}
}

func nodeOptionalSource(t *testing.T, source string) []byte {
	t.Helper()
	command := exec.Command("node", "--disable-warning=ExperimentalWarning", "--input-type=module", "-e", `import {stripTypeScriptTypes} from 'node:module'; let source = ''; for await (const part of process.stdin) source += part; await import('data:text/javascript;base64,' + Buffer.from(stripTypeScriptTypes(source)).toString('base64'));`)
	command.Stdin = strings.NewReader(source)
	out, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("Node source oracle: %v\n%s", err, out)
	}
	return out
}

func TestOptionalAdaptationLeavesUnsafeContractsRejected(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name, source string
		code         string
	}{
		{"required", "interface S { p: string }; const s: S = {p: 'x'}; s.p = undefined;", "2322"},
		{"wrong value", "interface S { p?: string }; const s: S = {}; s.p = 42;", "2322"},
		{"indexed read", "interface S { p?: string }; const s: S = {}; const a: string[] = []; s.p = a[0];", "2412"},
		{"live method", "class S { f?(): string { return 'x'; } }; const s = new S(); s.f = undefined;", "2412"},
		{"generic", "interface S { p?: string }; function f<T extends S>(s: T, p: string | undefined): void { s.p = p; }", "2412"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "main.ts")
			if err := os.WriteFile(path, []byte(test.source), 0644); err != nil {
				t.Fatal(err)
			}
			_, before := load.Load([]string{path})
			if diagnosticCount(before, test.code) == 0 {
				t.Fatalf("missing initial diagnostic %s: %v", test.code, before)
			}
			overlay, rewrites, err := adaptations([]string{path}, before)
			if err != nil {
				t.Fatal(err)
			}
			if len(rewrites) != 0 || len(overlay) != 0 {
				t.Fatalf("unsafe contract adapted: %#v %#v", rewrites, overlay)
			}
			_, after := load.LoadOverlay([]string{path}, overlay)
			if diagnosticCount(after, test.code) == 0 {
				t.Fatalf("lost diagnostic %s: %v", test.code, after)
			}
		})
	}
}

func TestOptionalAdaptationRechecksPresenceNarrowing(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "main.ts")
	source := "interface S { p?: string }; const s: S = {}; s.p = undefined; function f(s: S): number { if ('p' in s) return s.p.length; return 0; }"
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	_, before := load.Load([]string{path})
	overlay, rewrites, err := adaptations([]string{path}, before)
	if err != nil {
		t.Fatal(err)
	}
	if len(rewrites) != 1 || rewrites[0].Removed != 1 {
		t.Fatalf("adaptations = %#v", rewrites)
	}
	_, after := load.LoadOverlay([]string{path}, overlay)
	if diagnosticCount(after, "18048") != 1 {
		t.Fatalf("presence consumer was not rechecked: %v", after)
	}
}

func TestOptionalAdaptationContextualDeclarations(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"interface S { p?: string }; function f(p: string | undefined): S { return {p}; }",
		"interface S { p?: string }; const s: S = {p: undefined}; void s;",
		"interface S {outer?: {inner?: string}}; const s: S={}; s.outer=undefined; const x: NonNullable<S['outer']>={};x.inner=undefined;",
		"interface S { p?: string }; function f(a: string[]): S { return {p: a.length ? a[0] : undefined}; }",
		"interface S { f?(p: number): string }; const s: S = {}; s.f = undefined;",
		"interface S { f?<T>(p: T): T }; const s: S = {}; s.f = undefined;",
		"interface S { p?: string }; function f(s: S): void { void s; }; f({p: undefined});",
		"interface A {p: string | undefined}; interface B {p?: string}; function f(a: A): B {return a;}",
		"interface A {kind: 'a';p?: string}; interface B {kind:'b';p?: string}; function f(p: string | undefined): A | B {return {kind:'a',p};}",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "main.ts")
			if err := os.WriteFile(path, []byte(source), 0644); err != nil {
				t.Fatal(err)
			}
			_, before := load.Load([]string{path})
			if optionalDiagnosticCount(before) == 0 {
				t.Fatalf("initial diagnostic: %v", before)
			}
			overlay, rewrites, err := adaptations([]string{path}, before)
			if err != nil {
				t.Fatal(err)
			}
			if len(rewrites) != 1 || rewrites[0].Removed != optionalDiagnosticCount(before) {
				t.Fatalf("adaptations = %#v", rewrites)
			}
			if _, err := load.LoadOverlay([]string{path}, overlay); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestOptionalAdaptationOwnsOnlyNamedRoots(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	external := filepath.Join(dir, "external.ts")
	main := filepath.Join(dir, "main.ts")
	if err := os.WriteFile(external, []byte("export interface S {p?: string}"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(main, []byte("import type {S} from './external.ts'; const s:S={};s.p=undefined;"), 0644); err != nil {
		t.Fatal(err)
	}
	_, before := load.Load([]string{main})
	overlay, rewrites, err := adaptations([]string{main}, before)
	if err != nil {
		t.Fatal(err)
	}
	if len(overlay) != 0 || len(rewrites) != 0 {
		t.Fatalf("rewrote an external declaration: %#v %#v", overlay, rewrites)
	}
	_, after := load.LoadOverlay([]string{main}, overlay)
	if diagnosticCount(after, "2412") != 1 {
		t.Fatalf("external contract not preserved: %v", after)
	}
}

func TestOptionalAdaptationPreservesLiveMethodPlacement(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "main.ts")
	source := "class Action { optional?(): string { return 'x'; } }; const action = new Action(); console.log(`${Object.hasOwn(action, 'optional')} ${Object.hasOwn(Action.prototype, 'optional')}`); action.optional = undefined; interface Settings { value?: string }; const settings: Settings = {}; settings.value = undefined;"
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	_, before := load.Load([]string{path})
	if diagnosticCount(before, "2412") != 2 {
		t.Fatalf("initial diagnostic: %v", before)
	}
	overlay, rewrites, err := adaptations([]string{path}, before)
	if err != nil {
		t.Fatal(err)
	}
	adapted, changed := overlay[path]
	if !changed {
		adapted = source
	}
	// Check Node first: a method-to-field mutant can typecheck and preserve calls while
	// changing own-property and prototype placement. The runtime oracle must catch it.
	actual := nodeOptionalSource(t, adapted)
	if !bytes.Equal(actual, []byte("false true\n")) {
		t.Fatalf("adapted live method runtime output = %q, want false true", actual)
	}
	want := strings.Replace(source, "value?: string", "value?: (string) | undefined", 1)
	if adapted != want || len(rewrites) != 1 || rewrites[0].Rewrite != optionalRewrite || rewrites[0].Removed != 1 || rewrites[0].DeclarationsChanged != 1 {
		t.Fatalf("only the optional property may change, not the live method: %#v %q", rewrites, adapted)
	}
	_, after := load.LoadOverlay([]string{path}, overlay)
	if diagnosticCount(after, "2412") != 1 {
		t.Fatalf("the live-method assignment must remain rejected: %v", after)
	}
}

func TestOptionalAdaptationUTF16AndAdamicExtension(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".ts", ".a", ".a.ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "main"+extension)
			source := "const glyph = '😀'; interface S {p?: string}; const s: S = {}; s.p = undefined; void glyph;"
			if err := os.WriteFile(path, []byte(source), 0644); err != nil {
				t.Fatal(err)
			}
			_, before := load.Load([]string{path})
			overlay, rewrites, err := adaptations([]string{path}, before)
			if err != nil {
				t.Fatal(err)
			}
			if len(rewrites) != 1 || rewrites[0].Removed != 1 {
				t.Fatalf("adaptations = %#v; original=%v", rewrites, before)
			}
			if _, err := load.LoadOverlay([]string{path}, overlay); err != nil {
				t.Fatal(err)
			}
		})
	}
}
