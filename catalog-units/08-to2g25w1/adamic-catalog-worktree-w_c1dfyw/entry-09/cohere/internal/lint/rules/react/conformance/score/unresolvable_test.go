package react_conformance_score

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rules/react/conformance"
)

// TestUnresolvableTypesIsEarnedNotAssumed probes the excuse rather than trusting it.
//
// `VerdictUnresolvableTypes` says a type-aware rule read `any` because the fixture never imports
// react. That is a claim about a MECHANISM, and the danger the category carries is that it is also
// what a simply-broken rule looks like. So this supplies the missing types — a local `react.d.ts`
// declaring the hooks the fixture calls — and requires the rule's answer to CHANGE.
//
// A fixture that stays silent with types present was never excused by their absence, and its
// verdict is an excuse rather than a measurement.
func TestUnresolvableTypesIsEarnedNotAssumed(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("builds a TypeScript program per fixture")
	}
	fixtures, err := react_conformance.Load(fixtureRoot)
	if err != nil {
		t.Fatalf("loading the corpus: %v", err)
	}

	subject := Rules["set-state-in-render"]
	moved, stayedSilent := 0, 0
	for _, fixture := range SelectFixtures(fixtures, "set-state-in-render") {
		result, analyzeErr := Analyze(subject, fixture, t.TempDir())
		verdict := react_conformance.Classify(fixture, result, analyzeErr)
		if verdict.Verdict != react_conformance.VerdictUnresolvableTypes {
			continue
		}

		// Same fixture, with an import of a local react declaration prepended so the checker can
		// resolve `useState` to a real signature.
		// The import is prepended, and the fixture's own first line is kept. An earlier version of
		// this probe stripped a hardcoded pragma line with TrimPrefix, which matched only some of
		// these fixtures and silently left the rest unchanged; the module specifier was also './react'
		// rather than 'react', so the `declare module 'react'` stub could not resolve. Both bugs
		// produced the same symptom — every fixture silent — which read as the rule being broken.
		typed := fixture
		typed.Source = "import {useState} from 'react';\n" + fixture.Source
		directory := t.TempDir()
		if err := writeReactStub(directory); err != nil {
			t.Fatalf("writing the stub: %v", err)
		}
		withTypes, _ := Analyze(subject, typed, directory)

		if len(withTypes.Errors) > 0 {
			moved++
		} else {
			stayedSilent++
			t.Logf("  %s stayed silent even with react types present", fixture.Name)
		}
	}

	t.Logf("of the unresolvable-types fixtures: %d start reporting once types resolve, %d stay silent", moved, stayedSilent)
	if moved == 0 {
		t.Error("no fixture changed answer when types were supplied, so `unresolvable types` is not the mechanism it claims")
	}
}

func writeReactStub(directory string) error {
	// Copied in SHAPE from `set_state_in_render_test.go:reactStateDeclarations`, not invented.
	//
	// The first version of this stub declared the setter as a plain function type and every fixture
	// stayed silent, which read exactly like the rule being broken. It was the stub. This rule keys
	// on the setter type's ALIAS symbol, so `Dispatch` has to be a TYPE ALIAS and `RefObject` an
	// INTERFACE, matching `@types/react`. A stub that made both aliases, or neither, tests a checker
	// behaviour that does not exist and produces a confident wrong measurement.
	const stub = `declare module 'react' {
  export type SetStateAction<S> = S | ((prev: S) => S);
  export type Dispatch<A> = (value: A) => void;
  export interface RefObject<T> { current: T }
  export function useState<S>(initial?: S): [S, Dispatch<SetStateAction<S>>];
  export function useRef<T>(initial: T): RefObject<T>;
  export function useMemo<T>(compute: () => T, deps: unknown[]): T;
  export function useEffect(effect: () => void, deps?: unknown[]): void;
}`
	return os.WriteFile(filepath.Join(directory, "react.d.ts"), []byte(stub), 0o644)
}
