package tsprinter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// The executable and every ordered input are inputs to corpus generation.
// Only reference answers are cached; every tested printer still runs each case.
func tsPrinterOracleOutputs(t *testing.T, root string, files []string, gaps, family, oracle string) string {
	t.Helper()
	test, variable := "TestAdamicExpressionCorpus", "ADAMIC_TS_EXPRESSION_REQUEST"
	if family == "statements" {
		test, variable = "TestAdamicStatementCorpus", "ADAMIC_TS_STATEMENT_REQUEST"
	}
	external := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if external != "" {
		var err error
		external, err = filepath.Abs(external)
		if err != nil {
			t.Fatal(err)
		}
	}
	flags := []string{"test-run=" + test, "metadata=canonical-v1"}
	for _, file := range files {
		flags = append(flags, "input-id="+tsPrinterOutputLabel(file, root, external, true))
	}
	inputs := append([]string{oracle, gaps, root + "/stage1/cohere/tsprinter/expression_products_test.go", root + "/stage1/cohere/tsprinter/expressions_test.go"}, files...)
	product := expressionBuild(t, printerBuildInputs{Name: "TS " + family + " oracle outputs", Files: inputs, Flags: flags, Toolchain: runtime.Version()}, func(directory string) error {
		request, err := json.Marshal(map[string]any{"Files": files, "Directory": directory, "Gaps": gaps})
		if err != nil {
			return err
		}
		// Requests name this builder's paths and must never become shared outputs.
		requestPath := filepath.Join(t.TempDir(), "request.json")
		if err := os.WriteFile(requestPath, request, 0644); err != nil {
			return err
		}
		command := bounded(t, oracle, "-test.v", "-test.count=1", "-test.run=^"+test+"$", "-test.timeout=0")
		command.Dir = root + "/cohere"
		command.Env = append(os.Environ(), variable+"="+requestPath)
		output, err := combinedOutput(command)
		if err != nil {
			return fmt.Errorf("Go %s corpus: %w\n%s", family, err, output)
		}
		t.Log(string(output))
		return tsPrinterRewriteMetadata(directory, root, external, true)
	})
	// Shard preparation and caller-requested audits use a private writable copy.
	directory := t.TempDir()
	names := []string{"cases.txt", "answers.txt", "cases.json", "coverage.json"}
	if family == "expressions" {
		names = append(names, "gaps.txt", "gap-answers.txt", "gaps.json")
	}
	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(product, name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, name), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := tsPrinterRewriteMetadata(directory, root, external, false); err != nil {
		t.Fatal(err)
	}
	return directory
}

func tsPrinterOutputLabel(label, root, external string, canonical bool) string {
	for _, pair := range [][2]string{{root, "repo"}, {external, "typescript"}} {
		if pair[0] == "" {
			continue
		}
		from, to := pair[0]+"/", pair[1]+"/"
		if !canonical {
			from, to = to, from
		}
		if strings.HasPrefix(label, from) {
			return to + strings.TrimPrefix(label, from)
		}
	}
	return label
}

// Rewrite path metadata alone. Source and Want may contain the same strings,
// and changing those would silently change the independent oracle's answers.
func tsPrinterRewriteMetadata(directory, root, external string, canonical bool) error {
	path := filepath.Join(directory, "cases.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var cases []printerCase
	if err := json.Unmarshal(data, &cases); err != nil {
		return err
	}
	for i := range cases {
		cases[i].Label = tsPrinterOutputLabel(cases[i].Label, root, external, canonical)
	}
	data, err = json.Marshal(cases)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}
	path = filepath.Join(directory, "coverage.json")
	data, err = os.ReadFile(path)
	if err != nil {
		return err
	}
	var report map[string]json.RawMessage
	if err := json.Unmarshal(data, &report); err != nil {
		return err
	}
	if raw, exists := report["file_refusals"]; exists {
		var refusals map[string]string
		if err := json.Unmarshal(raw, &refusals); err != nil {
			return err
		}
		rewritten := make(map[string]string, len(refusals))
		var replacements []string
		for _, pair := range [][2]string{{root, "repo"}, {external, "typescript"}} {
			if pair[0] == "" {
				continue
			}
			from, to := pair[0]+"/", pair[1]+"/"
			if !canonical {
				from, to = to, from
			}
			replacements = append(replacements, from, to)
		}
		replace := strings.NewReplacer(replacements...)
		for label, diagnostic := range refusals {
			rewritten[tsPrinterOutputLabel(label, root, external, canonical)] = replace.Replace(diagnostic)
		}
		report["file_refusals"], err = json.Marshal(rewritten)
		if err != nil {
			return err
		}
	}
	data, err = json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func TestTSPrinterCachedMetadataRelocation(t *testing.T) {
	first, second, product, copy := t.TempDir(), t.TempDir(), t.TempDir(), t.TempDir()
	original := printerCase{Label: first + "/case.ts:3:expression", Source: "'repo/keep.ts'", Want: "typescript/keep.ts"}
	data, _ := json.Marshal([]printerCase{original})
	if err := os.WriteFile(product+"/cases.json", data, 0644); err != nil {
		t.Fatal(err)
	}
	report, _ := json.Marshal(map[string]any{"file_refusals": map[string]string{first + "/case.ts": first + "/case.ts: refused"}, "cases": 1})
	if err := os.WriteFile(product+"/coverage.json", report, 0644); err != nil {
		t.Fatal(err)
	}
	if err := tsPrinterRewriteMetadata(product, first, "", true); err != nil {
		t.Fatal(err)
	}
	canonical, err := os.ReadFile(product + "/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"cases.json", "coverage.json"} {
		data, err := os.ReadFile(product + "/" + name)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(copy+"/"+name, data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := tsPrinterRewriteMetadata(copy, second, "", false); err != nil {
		t.Fatal(err)
	}
	data, _ = os.ReadFile(copy + "/cases.json")
	var relocated []printerCase
	if err := json.Unmarshal(data, &relocated); err != nil {
		t.Fatal(err)
	}
	want := original
	want.Label = second + "/case.ts:3:expression"
	if len(relocated) != 1 || relocated[0] != want {
		t.Fatalf("relocation changed an oracle case: %+v", relocated)
	}
	unchanged, _ := os.ReadFile(product + "/cases.json")
	if !bytes.Equal(unchanged, canonical) {
		t.Fatal("materialization changed shared product")
	}
	if err := tsPrinterRewriteMetadata(copy, second, "", true); err != nil {
		t.Fatal(err)
	}
	roundtrip, _ := os.ReadFile(copy + "/cases.json")
	if !bytes.Equal(roundtrip, canonical) {
		t.Fatal("metadata differs after relocating and canonicalizing")
	}
}
