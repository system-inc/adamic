package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

func TestImplicitReturnsNeedNoAdaptation(t *testing.T) {
	t.Parallel()
	const source = `type Result<T> = T | undefined;
function lookup<T>(value: T, mode: number, undefined: number): Result<T> {
 try {
  if (mode === 1) return value;
  if (mode === 2) return /* bare */;
  console.log("fallthrough");
 } finally { console.log("finally " + mode); if (mode === 3) return value; }
 // trailing comment
}
const marker = "😀"; const arrow = (mode: number): number | undefined => {
 function inner(mode: number): number | undefined { if (mode === 1) return 7; }
 if (mode === 1) return inner(mode);
 console.log("arrow");
};
class Box { method(mode: number): number | undefined { if (mode === 1) return 9; } }
console.log("value " + lookup(5, 0, 42)); console.log("value " + lookup(5, 1, 42)); console.log("value " + lookup(5, 2, 42)); console.log("value " + lookup(5, 3, 42));
console.log("value " + arrow(0)); console.log("value " + arrow(1)); console.log("value " + new Box().method(0));
`
	for _, extension := range []string{".ts", ".a", ".a.ts"} {
		t.Run(extension, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "main"+extension)
			if err := os.WriteFile(path, []byte(source), 0644); err != nil {
				t.Fatal(err)
			}
			_, before := load.Load([]string{path})
			if before != nil {
				t.Fatalf("implicit undefined returns should compile without TS7030: %v", before)
			}
			overlay, rewrites, err := adaptations([]string{path}, before)
			if err != nil {
				t.Fatal(err)
			}
			if len(rewrites) != 0 || len(overlay) != 0 {
				t.Fatalf("rewrites=%#v", rewrites)
			}
			_, after := load.LoadOverlay([]string{path}, overlay)
			if after != nil {
				t.Fatal(after)
			}
			const want = "fallthrough\nfinally 0\nvalue undefined\nfinally 1\nvalue 5\nfinally 2\nvalue undefined\nfallthrough\nfinally 3\nvalue 5\narrow\nvalue undefined\nvalue 7\nvalue undefined\n"
			if output := nodeOptionalSource(t, source); string(output) != want {
				t.Fatalf("implicit return Node output=%q, want=%q", output, want)
			}
			disk, err := os.ReadFile(path)
			if err != nil || string(disk) != source {
				t.Fatal("source disk changed", err)
			}
			_, changed, err := returnAdaptations([]string{path}, overlay, after)
			if err != nil || changed != 0 {
				t.Fatalf("idempotence: %d %v", changed, err)
			}
		})
	}
}

func TestReturnAdaptationLeavesUnprovenContracts(t *testing.T) {
	t.Parallel()
	const source = `function inferred(flag: boolean) { if (flag) return 1; }
function required(flag: boolean): number { if (flag) return 1; }
function uncertain<T>(flag: boolean, value: T): T { if (flag) return value; }
function opaque(flag: boolean): unknown { if (flag) return 1; }
function anything(flag: boolean): any { if (flag) return 1; }
async function asynchronous(flag: boolean): Promise<number | undefined> { if (flag) return 1; }
function* generator(flag: boolean): Generator<number, number | undefined> { if (flag) return 1; }
`
	path := filepath.Join(t.TempDir(), "main.ts")
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	_, before := load.Load([]string{path})
	if diagnosticCount(before, "7030") != 0 || diagnosticCount(before, "2366") != 2 {
		t.Fatalf("want only the two required-return findings, without TS7030: %v", before)
	}
	overlay, changed, err := returnAdaptations([]string{path}, nil, before)
	if err != nil || changed != 0 || len(overlay) != 0 {
		t.Fatalf("unproven contract changed: %d %v %v", changed, overlay, err)
	}
}

func TestReturnAdaptationOwnsOnlyRoots(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	root := filepath.Join(directory, "main.ts")
	external := filepath.Join(directory, "external.ts")
	if err := os.WriteFile(external, []byte(`export function f(flag: boolean): number | undefined { if (flag) return 1; }`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(root, []byte(`import {f} from "./external.ts"; console.log('' + f(false));`), 0644); err != nil {
		t.Fatal(err)
	}
	_, before := load.Load([]string{root})
	if before != nil {
		t.Fatalf("external implicit return should compile without TS7030: %v", before)
	}
	overlay, changed, err := returnAdaptations([]string{root}, nil, before)
	if err != nil || changed != 0 || len(overlay) != 0 {
		t.Fatalf("external changed: %d %v %v", changed, overlay, err)
	}
}
