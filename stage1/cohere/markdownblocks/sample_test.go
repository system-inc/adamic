package markdownblocks

import (
	"fmt"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/gatesample"
	"github.com/system-inc/adamic/stage1/cohere/markdownblocks/oracles"
)

// The slowest cited seat sweep was 241s; ceil(241/30)=9.
// All nine layout tests share this stride and one baseline. Numeric controls,
// scalar/sequence checks and the full mutant streams are retained in full.
const markdownCorpusStride = 9

func selectMarkdownFiles(test *testing.T, root string, inputs []auditInput, files, stride int) gatesample.Selection {
	test.Helper()
	names := make([]string, files)
	for index := range names {
		names[index] = inputs[index].Name
	}
	selection, selectError := gatesample.Select(root, names, stride)
	if selectError != nil {
		test.Fatalf("%s: %v", test.Name(), selectError)
	}
	var controls []int
	for index, input := range inputs[files:] {
		if !input.Corpus {
			controls = append(controls, index)
		}
	}
	return selection.Generated(len(inputs)-files, controls...)
}

func selectedMarkdownInputs(inputs []auditInput, files int, selection gatesample.Selection) ([]auditInput, int) {
	if !selection.Sample {
		return inputs, files
	}
	included := map[string]bool{}
	for _, name := range selection.Paths {
		included[name] = true
	}
	var selected []auditInput
	for _, input := range inputs[:files] {
		if included[input.Name] {
			selected = append(selected, input)
		}
	}
	checkedFiles := len(selected)
	for index, input := range inputs[files:] {
		if included[gatesample.GeneratedKey(index)] {
			selected = append(selected, input)
		}
	}
	return selected, checkedFiles
}

func sampledNumericInputs(
	inputs [][]uint16,
	names []string,
	corpus []auditInput,
	selection gatesample.Selection,
) ([][]uint16, []string, []int) {
	included := map[string]bool{}
	for _, name := range selection.Paths {
		included[name] = true
	}
	files := selectionPhysicalCount(corpus)
	var selected [][]uint16
	var selectedNames []string
	var rows []int
	for index, name := range names {
		// Preserve fixed numeric probes, including the explicit smoke controls.
		keep := !selection.Sample || index == 0 || index > len(corpus)
		if index > 0 && index <= len(corpus) {
			if index-1 < files {
				keep = keep || included[name]
			} else {
				keep = keep || included[gatesample.GeneratedKey(index-1-files)]
			}
		}
		if keep {
			selected = append(selected, inputs[index])
			selectedNames = append(selectedNames, name)
			rows = append(rows, index)
		}
	}
	return selected, selectedNames, rows
}

func numericBatch(inputs [][]uint16) []byte { return oracles.NumericBatch(inputs) }

func selectedNumericAnswers(test *testing.T, answer run, rows []int, total int) run {
	test.Helper()
	lines := strings.Split(strings.TrimSuffix(string(answer.stdout), "\n"), "\n")
	if len(lines) != total {
		test.Fatalf("full numeric oracle returned %d/%d rows", len(lines), total)
	}
	selected := make([]string, len(rows))
	for index, row := range rows {
		selected[index] = lines[row]
	}
	answer.stdout = []byte(strings.Join(selected, "\n") + "\n")
	return answer
}

// Not parallel: shared markdownMemory configuration; existing helper controls parallel execution.
func TestSampleRetainsFixedMarkdownInputs(test *testing.T) {
	parallelMarkdown(test)
	inputs := []auditInput{{Name: "a.md"}, {Name: "b.md"}, {Name: "generated/list"}, {Name: "generated/quote"}}
	selection := gatesample.Selection{
		Sample: true,
		Paths:  []string{"b.md", gatesample.GeneratedKey(0), gatesample.GeneratedKey(1)},
	}
	selected, files := selectedMarkdownInputs(inputs, 2, selection)
	if files != 1 ||
		len(selected) != 3 ||
		selected[0].Name != "b.md" ||
		selected[1].Name != "generated/list" ||
		selected[2].Name != "generated/quote" {
		test.Fatalf("lost fixed inputs: %v (%d files)", selected, files)
	}
	full, files := selectedMarkdownInputs(inputs, 2, gatesample.Selection{})
	if len(full) != len(inputs) || files != 2 {
		test.Fatal("unset changed corpus counts")
	}
	numeric := [][]uint16{{}, {1}, {2}, {3}, {4}}
	names := []string{"empty", "a.md", "b.md", "generated/list", "unit/0"}
	kept, labels, rows := sampledNumericInputs(numeric, names, inputs[:2], selection)
	if len(kept) != 4 || strings.Join(labels, ",") != "empty,b.md,generated/list,unit/0" {
		test.Fatalf("lost numeric controls: %v", labels)
	}
	got := selectedNumericAnswers(test, run{stdout: []byte("empty\na\nb\nlist\nunit\n")}, rows, 5)
	if string(got.stdout) != "empty\nb\nlist\nunit\n" {
		test.Fatalf("oracle rows: %q", got.stdout)
	}
}

func selectionPhysicalCount(inputs []auditInput) int {
	for index, input := range inputs {
		if strings.HasPrefix(input.Name, "generated/") {
			return index
		}
	}
	return len(inputs)
}

// Not parallel: the landing switch is process-wide.
func TestGeneratedLayoutSelection(test *testing.T) {
	test.Setenv("ADAMIC_GATE_SAMPLE", strings.Repeat("0", 40))
	test.Setenv("ADAMIC_GATE_CHANGED", "")
	corpus := []auditInput{
		{Name: "generated/repeated", Corpus: true},
		{Name: "generated/control"},
		{Name: "generated/repeated", Corpus: true},
	}
	selection := selectMarkdownFiles(test, test.TempDir(), corpus, 0, 3)
	selected, files := selectedMarkdownInputs(corpus, 0, selection)
	if files != 0 || len(selected) != 2 || selected[1].Name != "generated/control" {
		test.Fatalf("indexed corpus/control: %v", selected)
	}
	numeric := [][]uint16{{}, {1}, {2}, {3}, {4}}
	names := []string{"empty", corpus[0].Name, corpus[1].Name, corpus[2].Name, "unit/0"}
	_, _, rows := sampledNumericInputs(numeric, names, corpus, selection)
	if fmt.Sprint(rows) != "[0 1 2 4]" {
		test.Fatalf("duplicate label changed index selection: %v", rows)
	}
}
