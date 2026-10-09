package tsprinter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type printerCase struct{ Label, Source, Want string }

func tscCorpus(t *testing.T, root string, files []string, family string, statements bool, oracle string) ([]printerCase, string) {
	t.Helper()
	gaps, _ := filepath.Abs("testdata/notyet.json")
	directory := tsPrinterOracleOutputs(t, root, files, gaps, family, oracle)
	data, err := os.ReadFile(filepath.Join(directory, "cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	var selected, cases []printerCase
	if err := json.Unmarshal(data, &selected); err != nil {
		t.Fatal(err)
	}
	escape := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t")
	var batch strings.Builder
	for _, item := range selected {
		// This focused gate holds repository fragments, not generated cases.
		if strings.HasPrefix(item.Label, root+"/stage3/drivers/tsc/corpus/") {
			cases = append(cases, item)
			batch.WriteString(">" + escape.Replace(item.Source) + "\n")
		}
	}
	path := filepath.Join(directory, "tsc-cases.txt")
	if err := os.WriteFile(path, []byte(batch.String()), 0644); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(cases)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "tsc-cases.json"), encoded, 0644); err != nil {
		t.Fatal(err)
	}

	return cases, directory
}
