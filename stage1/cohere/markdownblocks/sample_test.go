package markdownblocks

import (
	"fmt"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/gatesample"
)

// The slowest cited seat sweep was 241s; ceil(241/30)=9.
// All nine layout tests share this stride and one baseline. Numeric controls,
// generated documents and the full mutant streams are not file samples.
const markdownCorpusStride = 9

func selectMarkdownFiles(t *testing.T, root string, inputs []auditInput, files, stride int) gatesample.Selection {
	t.Helper()
	names := make([]string, files)
	for i := range names {
		names[i] = inputs[i].Name
	}
	selection, err := gatesample.Select(root, names, stride)
	if err != nil {
		t.Fatalf("%s: %v", t.Name(), err)
	}
	return selection
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
	return append(selected, inputs[files:]...), len(selected)
}

func sampledNumericInputs(inputs [][]uint16, names []string, physical []auditInput, selection gatesample.Selection) ([][]uint16, []string, []int) {
	included := map[string]bool{}
	for _, name := range selection.Paths {
		included[name] = true
	}
	physicalNames := map[string]bool{}
	for _, input := range physical {
		physicalNames[input.Name] = true
	}
	var selected [][]uint16
	var selectedNames []string
	var rows []int
	for i, name := range names {
		// Preserve fixed numeric probes, including the explicit smoke controls.
		if !selection.Sample || !physicalNames[name] || included[name] {
			selected = append(selected, inputs[i])
			selectedNames = append(selectedNames, name)
			rows = append(rows, i)
		}
	}
	return selected, selectedNames, rows
}

func numericBatch(inputs [][]uint16) []byte {
	var batch strings.Builder
	for _, units := range inputs {
		if len(units) == 0 {
			batch.WriteString("-")
		} else {
			for i, unit := range units {
				if i > 0 {
					batch.WriteByte(',')
				}
				fmt.Fprint(&batch, unit)
			}
		}
		batch.WriteByte('\n')
	}
	return []byte(batch.String())
}

func selectedNumericAnswers(t *testing.T, answer run, rows []int, total int) run {
	t.Helper()
	lines := strings.Split(strings.TrimSuffix(string(answer.stdout), "\n"), "\n")
	if len(lines) != total {
		t.Fatalf("full numeric oracle returned %d/%d rows", len(lines), total)
	}
	selected := make([]string, len(rows))
	for i, row := range rows {
		selected[i] = lines[row]
	}
	answer.stdout = []byte(strings.Join(selected, "\n") + "\n")
	return answer
}

func TestSampleRetainsFixedMarkdownInputs(t *testing.T) {
	parallelMarkdown(t)
	inputs := []auditInput{{Name: "a.md"}, {Name: "b.md"}, {Name: "generated/list"}, {Name: "generated/quote"}}
	selection := gatesample.Selection{Sample: true, Paths: []string{"b.md"}}
	selected, files := selectedMarkdownInputs(inputs, 2, selection)
	if files != 1 || len(selected) != 3 || selected[0].Name != "b.md" || selected[1].Name != "generated/list" || selected[2].Name != "generated/quote" {
		t.Fatalf("lost fixed inputs: %v (%d files)", selected, files)
	}
	full, files := selectedMarkdownInputs(inputs, 2, gatesample.Selection{})
	if len(full) != len(inputs) || files != 2 {
		t.Fatal("unset changed corpus counts")
	}
	numeric := [][]uint16{{}, {1}, {2}, {3}, {4}}
	names := []string{"empty", "a.md", "b.md", "generated/list", "unit/0"}
	kept, labels, rows := sampledNumericInputs(numeric, names, inputs[:2], selection)
	if len(kept) != 4 || strings.Join(labels, ",") != "empty,b.md,generated/list,unit/0" {
		t.Fatalf("lost numeric controls: %v", labels)
	}
	got := selectedNumericAnswers(t, run{stdout: []byte("empty\na\nb\nlist\nunit\n")}, rows, 5)
	if string(got.stdout) != "empty\nb\nlist\nunit\n" {
		t.Fatalf("oracle rows: %q", got.stdout)
	}
}
