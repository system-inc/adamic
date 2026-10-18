package json

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/json/oracles"
)

type caseIdentity = oracles.CaseIdentity
type groupPin = oracles.GroupPin
type corpusPin = oracles.CorpusPin

func caseGroup(path string) string { return oracles.CaseGroup(path) }

// The digest is SHA256 of sorted path<TAB>SHA256(text)<LF> identities.
func pinForCases(cases []textCase) corpusPin { return oracles.PinForCases(oracleCases(cases)) }

func corpusPinError(want, got corpusPin) error { return oracles.CorpusPinError(want, got) }

func TestCorpusPinDiagnostics(t *testing.T) {
	t.Parallel()
	cases := []textCase{{"a.json", "{}"}, {"cohere/b.json", "[]"}, {"cohere/TypeScript/c.json", "0"}, {"generated/0/probe.json", "null"}}
	pin := pinForCases(cases)
	if err := corpusPinError(pin, pinForCases([]textCase{cases[3], cases[0], cases[2], cases[1]})); err != nil {
		t.Fatal("order changed identity", err)
	}
	for _, mutant := range []struct {
		name    string
		cases   []textCase
		message string
	}{
		{"missing", cases[1:], "missing a.json"},
		{"extra", append(append([]textCase(nil), cases...), textCase{"extra.json", "true"}), "extra extra.json"},
		{"one byte", []textCase{{"a.json", "{ }"}, cases[1], cases[2], cases[3]}, "changed text a.json"},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			err := corpusPinError(pin, pinForCases(mutant.cases))
			if err == nil || !strings.Contains(err.Error(), mutant.message) || !strings.Contains(err.Error(), "provisioned:") {
				t.Fatalf("pin mutant survived: %v", err)
			}
			t.Logf("caught: %v", err)
		})
	}
}

func TestCorpusPinRejectsMissingProvisionedInputs(t *testing.T) {
	t.Parallel()
	cases := corpusCases(t)
	encoded, err := os.ReadFile("testdata/corpus.pin")
	if err != nil {
		t.Fatal(err)
	}
	var full corpusPin
	if err := json.Unmarshal(encoded, &full); err != nil {
		t.Fatal(err)
	}
	var short []textCase
	for _, item := range cases {
		if !strings.HasPrefix(item.Name, "stage3/api/node_modules/") {
			short = append(short, item)
		}
	}
	if len(cases)-len(short) != 18 {
		t.Fatal("provisioned input witness changed")
	}
	_, _, err = validatePinnedCorpus(t, short, full)
	if err == nil || !strings.Contains(err.Error(), "missing stage3/api/node_modules/typescript/package.json") || !strings.Contains(err.Error(), "provisioned: got 0, pinned 18") {
		t.Fatalf("short-box pin mutant survived: %v", err)
	}

	t.Logf("caught the 18-input provisioning mutant: %v", err)
}

// Not parallel: this proves the process-wide sample switch cannot bypass the pin.
func TestCorpusPinBeforeSampling(t *testing.T) {
	t.Setenv("ADAMIC_GATE_SAMPLE", "00000008"+strings.Repeat("0", 32))
	changed := t.TempDir() + "/changed.txt"
	named := "stage3/api/node_modules/typescript/package.json"
	if err := os.WriteFile(changed, []byte(named+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ADAMIC_GATE_CHANGED", changed)
	total := len(corpusCases(t))
	cases := sampledCorpusCases(t, 32)
	found, generated := false, 0
	for _, item := range cases {
		if item.Name == named {
			found = true
		}
		if caseGroup(item.Name) == "generated" {
			generated++
		}
	}
	if !found || generated != 34 || len(cases) >= total {
		t.Fatalf("sample changed/controls/count: %v %d %d", found, generated, len(cases))
	}
}
