package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

func TestAdaptUndefinedReturnsPreservesNodeAndDisk(t *testing.T) {
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
			if diagnosticCount(before, "7030") != 4 {
				t.Fatalf("want four return findings: %v", before)
			}
			overlay, rewrites, err := adaptations([]string{path}, before)
			if err != nil {
				t.Fatal(err)
			}
			if len(rewrites) != 1 || rewrites[0].FunctionsChanged != 4 || rewrites[0].Removed != 4 {
				t.Fatalf("rewrites=%#v", rewrites)
			}
			adapted := overlay[path]
			if !strings.Contains(adapted, "return void 0 /* bare */;") {
				t.Fatalf("bare return not explicit: %s", adapted)
			}
			_, after := load.LoadOverlay([]string{path}, overlay)
			if after != nil {
				t.Fatal(after)
			}
			originalOutput := nodeOptionalSource(t, source)
			if output := nodeOptionalSource(t, adapted); !bytes.Equal(output, originalOutput) {
				t.Fatalf("adapted Node output=%q, original=%q", output, originalOutput)
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
	if diagnosticCount(before, "7030") == 0 {
		t.Fatalf("no return finding: %v", before)
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
	if err := os.WriteFile(root, []byte(`import {f} from "./external.ts"; console.log(f(false));`), 0644); err != nil {
		t.Fatal(err)
	}
	_, before := load.Load([]string{root})
	if diagnosticCount(before, "7030") != 1 {
		t.Fatalf("no external finding: %v", before)
	}
	overlay, changed, err := returnAdaptations([]string{root}, nil, before)
	if err != nil || changed != 0 || len(overlay) != 0 {
		t.Fatalf("external changed: %d %v %v", changed, overlay, err)
	}
}
