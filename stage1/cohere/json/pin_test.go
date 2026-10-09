package json

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

type caseIdentity struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}
type groupPin struct {
	Count  int    `json:"count"`
	SHA256 string `json:"sha256"`
}
type corpusPin struct {
	Count  int                 `json:"count"`
	SHA256 string              `json:"sha256"`
	Groups map[string]groupPin `json:"groups"`
	Cases  []caseIdentity      `json:"cases"`
}

func caseGroup(path string) string {
	if strings.HasPrefix(path, "generated/") {
		return "generated"
	}
	if strings.HasPrefix(path, "cohere/TypeScript/") {
		return "TypeScript"
	}
	if strings.HasPrefix(path, "cohere/") {
		return "cohere"
	}
	return "provisioned"
}

// The digest is SHA256 of sorted path<TAB>SHA256(text)<LF> identities.
func pinForCases(cases []textCase) corpusPin {
	pin := corpusPin{Count: len(cases), Groups: map[string]groupPin{}}
	for _, item := range cases {
		pin.Cases = append(pin.Cases, caseIdentity{item.Name, fmt.Sprintf("%x", sha256.Sum256([]byte(item.Text)))})
	}
	sort.Slice(pin.Cases, func(i, j int) bool {
		if pin.Cases[i].Path == pin.Cases[j].Path {
			return pin.Cases[i].SHA256 < pin.Cases[j].SHA256
		}
		return pin.Cases[i].Path < pin.Cases[j].Path
	})
	var all strings.Builder
	groups := map[string]*strings.Builder{}
	for _, item := range pin.Cases {
		line := item.Path + "\t" + item.SHA256 + "\n"
		all.WriteString(line)
		group := caseGroup(item.Path)
		if groups[group] == nil {
			groups[group] = &strings.Builder{}
		}
		groups[group].WriteString(line)
		record := pin.Groups[group]
		record.Count++
		pin.Groups[group] = record
	}
	pin.SHA256 = fmt.Sprintf("%x", sha256.Sum256([]byte(all.String())))
	for group, lines := range groups {
		record := pin.Groups[group]
		record.SHA256 = fmt.Sprintf("%x", sha256.Sum256([]byte(lines.String())))
		pin.Groups[group] = record
	}
	return pin
}

func corpusPinError(want, got corpusPin) error {
	// Recompute the committed manifest too: a corrupted count/digest cannot pass.
	summarize := func(pin corpusPin) corpusPin {
		var lines strings.Builder
		for _, item := range pin.Cases {
			lines.WriteString(item.Path + "\t" + item.SHA256 + "\n")
		}
		pin.SHA256 = fmt.Sprintf("%x", sha256.Sum256([]byte(lines.String())))
		return pin
	}
	if summarize(want).SHA256 != want.SHA256 || want.Count != len(want.Cases) {
		return fmt.Errorf("corpus pin manifest/count/digest is inconsistent")
	}
	var differences []string
	for _, group := range []string{"provisioned", "TypeScript", "cohere", "generated"} {
		a, b := want.Groups[group], got.Groups[group]
		if a != b {
			differences = append(differences, fmt.Sprintf("%s: got %d, pinned %d (identity SHA256 got %s, pinned %s)", group, b.Count, a.Count, b.SHA256, a.SHA256))
		}
	}
	if want.Count == got.Count && want.SHA256 == got.SHA256 && len(differences) == 0 {
		return nil
	}
	expected, observed := map[string]string{}, map[string]string{}
	for _, item := range want.Cases {
		expected[item.Path] = item.SHA256
	}
	for _, item := range got.Cases {
		observed[item.Path] = item.SHA256
	}
	for _, item := range want.Cases {
		hash, exists := observed[item.Path]
		if !exists {
			differences = append(differences, "missing "+item.Path)
		} else if hash != item.SHA256 {
			differences = append(differences, "changed text "+item.Path)
		}
	}
	for _, item := range got.Cases {
		if _, exists := expected[item.Path]; !exists {
			differences = append(differences, "extra "+item.Path)
		}
	}
	return fmt.Errorf("corpus pin: got %d, pinned %d; %s", got.Count, want.Count, strings.Join(differences, "; "))
}

func verifyCorpusPin(t *testing.T, cases []textCase) {
	t.Helper()
	encoded, err := os.ReadFile("testdata/corpus.pin")
	if err != nil {
		t.Fatalf("%s corpus pin: %v", t.Name(), err)
	}
	var expected corpusPin
	if err := json.Unmarshal(encoded, &expected); err != nil {
		t.Fatalf("%s corpus pin: %v", t.Name(), err)
	}
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	actual, count, err := validateCorpus(root, cases, expected)
	t.Logf("corpus repository: %d tracked JSON files; checked against Git HEAD", count)
	if err != nil {
		t.Fatalf("%s: %v", t.Name(), err)
	}
	for _, group := range []string{"provisioned", "TypeScript", "cohere", "generated"} {
		t.Logf("corpus pin group %s: %d, SHA256 %s", group, actual.Groups[group].Count, actual.Groups[group].SHA256)
	}
}

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
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = validateCorpus(root, short, full)
	if err == nil || !strings.Contains(err.Error(), "missing stage3/api/node_modules/typescript/package.json") || !strings.Contains(err.Error(), "provisioned: got 0, pinned 18") {
		t.Fatalf("short-box pin mutant survived: %v", err)
	}

	t.Logf("caught the 18-input provisioning mutant: %v", err)
}

const testCorpusPinBeforeSamplingShards = 1

// ADAMIC_TEST_SHARD=i/n selects ordinal modulo n; unset runs every shard.
// The sampling probe uses a subprocess so its process-wide environment is isolated.
func TestCorpusPinBeforeSampling(t *testing.T) {
	if os.Getenv("ADAMIC_JSON_PIN_PROBE") == "1" {
		jsonCorpusPinSamplingProbe(t)
		return
	}
	t.Parallel()
	cases := corpusCases(t)
	shards := []nativeChunk{{start: 0, end: len(cases)}}
	if len(shards) != testCorpusPinBeforeSamplingShards {
		t.Fatal("corpus-pin shard count changed")
	}
	if err := jsonPortUnion(cases, shards); err != nil {
		t.Fatal(err)
	}
	index, count, err := jsonPortSelection(os.Getenv("ADAMIC_TEST_SHARD"))
	if err != nil {
		t.Fatal(err)
	}
	for ordinal, shard := range shards {
		if ordinal%count != index {
			continue
		}
		t.Run(fmt.Sprintf("shard-%03d", ordinal), func(t *testing.T) {
			t.Parallel()
			binary, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			result := execute(t, []string{"ADAMIC_JSON_PIN_PROBE=1"}, binary, "-test.run=^TestCorpusPinBeforeSampling$", "-test.v", "-test.timeout=30s")
			if result.exitCode != 0 || len(result.stderr) != 0 {
				t.Fatalf("%s sampling probe: exit %d stderr %s output %s", t.Name(), result.exitCode, result.stderr, result.stdout)
			}
			t.Logf("case range [%d:%d]; union %d cases; isolated probe:\n%s", shard.start, shard.end, len(cases), result.stdout)
		})
	}
}

func jsonCorpusPinSamplingProbe(t *testing.T) {
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
