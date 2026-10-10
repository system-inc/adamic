package formatfiles

import (
	"os"
	"regexp"
	"strconv"
	"testing"
)

// Keep every full-corpus mutant independently discoverable by the gate.
func TestFormatfilesMutantsUnion(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("formatfiles_mutants_test.go")
	if err != nil {
		t.Fatal(err)
	}
	matches := regexp.MustCompile(`(?m)^func TestFormatfilesMutant_([0-9]+)\(t \*testing.T\) \{\s*t\.Parallel\(\)\s*formatfilesCheckMutant\(t, ([0-9]+)\)\s*\}`).FindAllSubmatch(source, -1)
	seen := make(map[int]bool)
	for _, match := range matches {
		index, err := strconv.Atoi(string(match[1]))
		target, targetErr := strconv.Atoi(string(match[2]))
		if err != nil || targetErr != nil || target != index || index >= len(mutants) || seen[index] {
			t.Fatalf("invalid or repeated mutant index %s", match[1])
		}
		seen[index] = true
	}
	if len(seen) != len(mutants) {
		t.Fatalf("enumerated %d, want %d mutants", len(seen), len(mutants))
	}
}

func TestFormatfilesMutant_000(t *testing.T) {
	t.Parallel()
	formatfilesCheckMutant(t, 0)
}
func TestFormatfilesMutant_001(t *testing.T) {
	t.Parallel()
	formatfilesCheckMutant(t, 1)
}
func TestFormatfilesMutant_002(t *testing.T) {
	t.Parallel()
	formatfilesCheckMutant(t, 2)
}
func TestFormatfilesMutant_003(t *testing.T) {
	t.Parallel()
	formatfilesCheckMutant(t, 3)
}
func TestFormatfilesMutant_004(t *testing.T) {
	t.Parallel()
	formatfilesCheckMutant(t, 4)
}
func TestFormatfilesMutant_005(t *testing.T) {
	t.Parallel()
	formatfilesCheckMutant(t, 5)
}
func TestFormatfilesMutant_006(t *testing.T) {
	t.Parallel()
	formatfilesCheckMutant(t, 6)
}
func TestFormatfilesMutant_007(t *testing.T) {
	t.Parallel()
	formatfilesCheckMutant(t, 7)
}
func TestFormatfilesMutant_008(t *testing.T) {
	t.Parallel()
	formatfilesCheckMutant(t, 8)
}
func TestFormatfilesMutant_009(t *testing.T) {
	t.Parallel()
	formatfilesCheckMutant(t, 9)
}
func TestFormatfilesMutant_010(t *testing.T) {
	t.Parallel()
	formatfilesCheckMutant(t, 10)
}
func TestFormatfilesMutant_011(t *testing.T) {
	t.Parallel()
	formatfilesCheckMutant(t, 11)
}
func TestFormatfilesMutant_012(t *testing.T) {
	t.Parallel()
	formatfilesCheckMutant(t, 12)
}
func TestFormatfilesMutant_013(t *testing.T) {
	t.Parallel()
	formatfilesCheckMutant(t, 13)
}
func TestFormatfilesMutant_014(t *testing.T) {
	t.Parallel()
	formatfilesCheckMutant(t, 14)
}
