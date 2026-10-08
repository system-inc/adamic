package lint

import (
	"strings"
	"testing"
)

// Built through a Go overlay so this rule owns every changed line.
func TestReactNoArrowFunctionLifecycleSelected(t *testing.T) {
	t.Parallel()
	oracle := goOracle(t)
	var rows []string
	for _, row := range upstream(t) {
		fields := strings.Split(row, "\t")
		if len(fields) > 1 && fields[1] == "react/no-arrow-function-lifecycle" {
			rows = append(rows, row)
		}
	}
	if len(rows) != 102 {
		t.Fatalf("upstream capture count drift: %d, want 102", len(rows))
	}
	t.Logf("react/no-arrow-function-lifecycle: %d unique captured upstream cases", len(rows))
	for _, d := range prepareRegistry(t, ".") {
		if d.Name != "react/no-arrow-function-lifecycle" {
			continue
		}
		for _, row := range ownedWitnessRows(t, ".", d) {
			count := execute(t, "", oracle, "--manifest", manifest(t, []string{row}), "--count")
			if string(count.output) == "0\n" {
				t.Fatal("witness reports no findings", row)
			}
			rows = append(rows, row)
			rows = append(rows, strings.SplitN(row, "\t", 2)[0]+"\tall")
		}
	}
	compare(t, oracle, buildPort(t, ".", true), ".", manifest(t, rows))
}

func TestReactNoArrowFunctionLifecycleNativeMutant(t *testing.T) {
	t.Parallel()
	oracle := goOracle(t)
	var rows []string
	for _, d := range prepareRegistry(t, ".") {
		if d.Name == "react/no-arrow-function-lifecycle" {
			rows = ownedWitnessRows(t, ".", d)
		}
	}
	path := manifest(t, rows)
	want := execute(t, "", oracle, "--manifest", path).output
	directory := mutant(t, "name === 'getDerivedStateFromProps'", "name === 'neverALifecycleMethod'", "rules/react-no-arrow-function-lifecycle/rule.a")
	binary := buildMutantPort(t, directory)
	got := execute(t, "", binary, "--manifest", path).output
	diff := difference(got, want)
	if diff == "" {
		t.Fatal("static lifecycle omission survived on sanitized native")
	}
	t.Logf("react_lifecycle_static_list_omitted caught on sanitized native: %s", diff)
}
