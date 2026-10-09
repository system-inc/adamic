package json

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/gatesample"
)

// The seat measured 888s for port agreement and 226s for upstream parity.
// ceil(888/30)=30 and ceil(226/30)=8; pin 32 and 8 respectively.
// These budget the corpus work, not compiler startup or mandatory fixed checks.
func sampledCorpusCases(t *testing.T, stride int) []textCase {
	t.Helper()
	cases := corpusCases(t)
	var names []string
	for _, item := range cases {
		if !strings.HasPrefix(item.Name, "generated/") {
			names = append(names, item.Name)
		}
	}
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	// This recorded upstream witness is a fixed check in both modes. Keeping it
	// also leaves the exact nine-discrepancy report and its comparisons untouched.
	selection, err := gatesample.Select(root, names, stride, "stage1/cohere/json/gaps/numeric-separators.json")
	if err != nil {
		t.Fatalf("%s: %v", t.Name(), err)
	}
	if !selection.Sample {
		return cases
	}
	t.Log(selection.Log(t.Name()))
	included := map[string]bool{}
	for _, name := range selection.Paths {
		included[name] = true
	}
	var result []textCase
	for _, item := range cases {
		if strings.HasPrefix(item.Name, "generated/") || included[item.Name] {
			result = append(result, item)
		}
	}
	return result
}
