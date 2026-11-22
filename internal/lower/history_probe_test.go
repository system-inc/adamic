//go:build historyprobe

// The history probe is a measurement, not a pin, so it builds only under its tag: go test -tags
// historyprobe -run '^TestHistoryProbe$' -v ./internal/lower/. It asks whether lowering one program
// twice gives the checker a different history (#77ncr00): a Go map's iteration order can change which
// type relations lowering asks the checker, and refuseAnyRelation's member-typed-any stop answers by
// what earlier walks asked it. Never landed.

package lower

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

// historyProbeRuns is how many times each program is lowered, each from a fresh load.
const historyProbeRuns = 5

// historyProbeRelations are the checker's relation caches the probe counts, by the name each row
// prints and the Checker field that holds it (relation_overflow.go's relationTables).
var historyProbeRelations = []struct{ label, field string }{
	{"assignable", "assignableRelation"},
	{"comparable", "comparableRelation"},
	{"identity", "identityRelation"},
	{"subtype", "subtypeRelation"},
	{"strictSubtype", "strictSubtypeRelation"},
}

// historyProbeT1 and historyProbeT7 are TestUnionKeepingPromiseLikeMemberAnyStop's t1 and t7, copied
// as they stand there.
const historyProbeT1 = `import type { Weak } from 'adamic';
interface Box { readonly label: string }
type Strong = { readonly kind: 'v'; readonly v: Box } | { readonly kind: 'w'; readonly w: number };
type Weakly = { readonly kind: 'v'; readonly v: Weak<Box> } | { readonly kind: 'w'; readonly w: number };
function show(x: Weakly): string { return x.kind === 'v' ? x.v?.label ?? 'none' : 'w'; }
async function produce(): Promise<Strong> { return { kind: 'v', v: { label: 'label' } }; }
function hold(): Promise<Strong> | number { return produce(); }
async function main(): Promise<void> {
  const held: Promise<Strong> | number = hold();
  const viewed: PromiseLike<Weakly> | number = held;
  if (typeof viewed !== 'number') {
    const x = await viewed;
    console.log(show(x));
  }
}
await main();
`

const historyProbeT7 = `import type { Weak } from 'adamic';
interface Box { readonly label: string }
type Strong = { readonly kind: 'v'; readonly v: Box } | { readonly kind: 'w'; readonly w: number };
type Weakly = { readonly kind: 'v'; readonly v: Weak<Box> } | { readonly kind: 'w'; readonly w: number };
function show(x: Weakly): string { return x.kind === 'v' ? x.v?.label ?? 'none' : 'w'; }
async function produce(): Promise<Strong> { return { kind: 'v', v: { label: 'label' } }; }
function hold(): Promise<Strong> { return produce(); }
async function main(): Promise<void> {
  const held: Promise<Strong> = hold();
  const viewed: PromiseLike<Weakly> = held;
  const x = await viewed;
  console.log(show(x));
}
await main();
`

// historyProbeCount is the size of one relation cache, read by name the way RelationOverflowMarked
// reads its marks, or "?" when the field isn't where it's expected. Len reads an unexported map
// without Interface, so the reflection never needs an exported door in internal/load.
func historyProbeCount(typeChecker *checker.Checker, field string) string {
	relation := reflect.ValueOf(typeChecker).Elem().FieldByName(field)
	if !relation.IsValid() || relation.Kind() != reflect.Pointer {
		return "?"
	}
	if relation.IsNil() {
		return "0"
	}
	results := relation.Elem().FieldByName("results")
	if !results.IsValid() || results.Kind() != reflect.Map {
		return "?"
	}
	return fmt.Sprint(results.Len())
}

// historyProbeRow lowers the program at path from a fresh load, a new Program and checker, and is its
// row's stop and counts: the stop from the file's own name on, or LOWERS.
func historyProbeRow(path string) (string, string) {
	counts := make([]string, 0, len(historyProbeRelations))
	program, err := load.Load([]string{path})
	if err != nil {
		for _, relation := range historyProbeRelations {
			counts = append(counts, relation.label+"=?")
		}
		return "LOAD " + strings.ReplaceAll(err.Error(), "\n", " | "), strings.Join(counts, " ")
	}
	ctx := context.Background()
	_, lowered := Lower(ctx, program)
	stop := "LOWERS"
	if lowered != nil {
		stop = lowered.Error()
		if index := strings.Index(stop, filepath.Base(path)+":"); index >= 0 {
			stop = stop[index:]
		}
		stop = strings.ReplaceAll(stop, "\n", " | ")
	}
	// Stage 0 checks single-threaded, so the entry's checker is the one Lower used and released.
	typeChecker, release := program.Checker(ctx, program.Files()[0])
	for _, relation := range historyProbeRelations {
		counts = append(counts, relation.label+"="+historyProbeCount(typeChecker, relation.field))
	}
	release()
	return stop, strings.Join(counts, " ")
}

// Not parallel: one program lowered at a time, read as one log.
// TestHistoryProbe lowers t1 and t7, as .a and as .ts, and the oracle's no_escape.ts, each five times
// from a fresh checker, and prints a ROW per lowering and a DISTINCT per program: how many different
// (stop, counts) tuples its runs gave.
func TestHistoryProbe(t *testing.T) {
	directory := t.TempDir()
	type probe struct{ name, path string }
	probes := []probe{}
	for _, extension := range []string{".a", ".ts"} {
		for _, source := range []struct{ name, text string }{{"t1", historyProbeT1}, {"t7", historyProbeT7}} {
			// One directory a program, every file named main as lowerSourceAs names it, so each stop
			// reads main.a: or main.ts: and every run of one program loads the same path.
			programDirectory := filepath.Join(directory, source.name+extension)
			if err := os.Mkdir(programDirectory, 0o755); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(programDirectory, "main"+extension)
			if err := os.WriteFile(path, []byte(source.text), 0o644); err != nil {
				t.Fatal(err)
			}
			probes = append(probes, probe{source.name + extension, path})
		}
	}
	// no_escape.ts is the largest .ts program under 400 lines in internal/oracle/testdata (381 lines,
	// 43 shapes), one the oracle lowers (predicate_false_flow_test.go), with no imports.
	noEscape, err := filepath.Abs("../oracle/testdata/predicate_false_flow/no_escape.ts")
	if err != nil {
		t.Fatal(err)
	}
	probes = append(probes, probe{"no_escape.ts", noEscape})

	for _, entry := range probes {
		distinct := map[string]bool{}
		for run := 1; run <= historyProbeRuns; run++ {
			stop, counts := historyProbeRow(entry.path)
			distinct[stop+"\x00"+counts] = true
			fmt.Printf("ROW %s run=%d stop=%s %s\n", entry.name, run, stop, counts)
		}
		fmt.Printf("DISTINCT %s %d\n", entry.name, len(distinct))
	}
}
