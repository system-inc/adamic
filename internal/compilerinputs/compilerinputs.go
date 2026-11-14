// Package compilerinputs derives the repository inputs linked by each compiler
// root set and holds their generated projections. A port's compiler is already
// linked into its test binary, so a product key must name these sources explicitly.
// Runtime callers receive static copies and never run go list.
//
// Lowering names tsc's linked package directories rather than its baselines and
// fourslash tests (#4ycgm2n). Markdownblocks retains its older whole-tsc key as an
// explicit coarsening. The exact-output guard keeps every projection generated.
package compilerinputs

import (
	"fmt"
	"slices"
	"strings"
)

//go:generate go run ./generate -root ../..

type inputSet struct {
	roots  []string
	inputs []string
}

// Lowering is the compiler's repository paths, relative to the repository root, as a key's Files names them: a fresh
// copy each call, so a caller appending its own files changes no one else's.
func Lowering() []string {
	return slices.Clone(lowering)
}

// Markdownblocks keeps its existing coarse compiler key. The generated projection
// includes tsc whole and internal whole, rather than changing those product recipes.
func Markdownblocks() []string { return slices.Clone(markdownblocks) }

// Admission is the lowering companion's inputs, including cmd/adamic's adapter site.
func Admission() []string { return slices.Clone(admission) }

// AdmissionText is the generator's inputs.txt representation of the committed projection.
func AdmissionText() []byte { return admissionText(Admission()) }

func admissionText(inputs []string) []byte { return []byte(strings.Join(inputs, "\n") + "\n") }

// InputsFor returns the committed dependency set for roots, not a new go list run.
// Consumer guards judge their projections against this snapshot; the exact-output
// guard alone compares the snapshot with the live compiler links.
func InputsFor(roots []string) ([]string, error) {
	roots = orderedSet(roots)
	for _, set := range generatedInputs {
		if slices.Equal(roots, set.roots) {
			return slices.Clone(set.inputs), nil
		}
	}
	return nil, fmt.Errorf("no generated compiler inputs for roots %v", roots)
}
