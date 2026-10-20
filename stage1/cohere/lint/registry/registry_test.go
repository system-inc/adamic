package registry

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// copyRules copies what Generate reads: each rule's own files and its testdata/, where Discover finds its witnesses,
// never another directory below a rule (evidence/), whose churn would move these tests' read sets (#s57yq6y).
func copyRules(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	// Reads: stage1/cohere/lint/rules/*
	err := filepath.WalkDir("../rules", func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel("..", path)
		if err != nil {
			return err
		}
		target := filepath.Join(root, relative)
		if entry.IsDir() {
			if parts := strings.Split(filepath.ToSlash(relative), "/"); len(parts) == 3 && parts[2] != "testdata" {
				return filepath.SkipDir
			}
			return os.MkdirAll(target, 0755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0644)
	})
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func TestDescriptorRejections(t *testing.T) {
	t.Parallel()
	for _, change := range []struct {
		name   string
		mutate func(map[string]any)
		want   string
	}{
		{"duplicate public name", func(d map[string]any) { d["name"] = "no-empty" }, "duplicate rule name"},
		{"unknown field", func(d map[string]any) { d["vist"] = "visit" }, "unknown field"},
		{"missing named export", func(d map[string]any) { d["visit"] = "missing" }, "missing named hook"},
		{"bad kind", func(d map[string]any) { d["kinds"] = []string{"NotARealNode"} }, "invalid or duplicate kind"},
		{"missing oracle export", func(d map[string]any) { d["oracle"] = "missing" }, "missing oracle adapter exports"},
		{"no listener", func(d map[string]any) { d["kinds"] = []string{} }, "visit and kinds required"},
		{"missing factory", func(d map[string]any) { d["factory"] = "missing" }, "missing named factory or class"},
		{"missing class", func(d map[string]any) { d["class"] = "Missing" }, "missing named factory or class"},
		{"missing finish hook", func(d map[string]any) { d["finish"] = "missing" }, "missing named hook"},
		{"unsafe public name", func(d map[string]any) { d["name"] = "bad\\name" }, "invalid or duplicate rule name"},
	} {
		t.Run(change.name, func(t *testing.T) {
			t.Parallel()
			root := copyRules(t)
			path := filepath.Join(root, "rules/no-debugger/rule.json")
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var d map[string]any
			if err := json.Unmarshal(data, &d); err != nil {
				t.Fatal(err)
			}
			change.mutate(d)
			data, err = json.Marshal(d)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, data, 0644); err != nil {
				t.Fatal(err)
			}
			_, err = Generate(root)
			if err == nil || !strings.Contains(err.Error(), change.want) {
				t.Fatalf("mutant survived or wrong failure: %v", err)
			}
			t.Logf("mutant rejected: %v", err)
		})
	}
}

func TestDuplicateOracleAdapter(t *testing.T) {
	t.Parallel()
	root := copyRules(t)
	path := filepath.Join(root, "rules/no-empty/oracle.go")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, bytes.ReplaceAll(data, []byte("oracleNoEmpty"), []byte("oracleNoDebugger")), 0644); err != nil {
		t.Fatal(err)
	}
	path = filepath.Join(root, "rules/no-empty/rule.json")
	data, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, bytes.ReplaceAll(data, []byte("oracleNoEmpty"), []byte("oracleNoDebugger")), 0644); err != nil {
		t.Fatal(err)
	}
	_, err = Generate(root)
	if err == nil || !strings.Contains(err.Error(), "duplicate oracle adapter") {
		t.Fatalf("duplicate adapter mutant survived: %v", err)
	}
	t.Logf("mutant rejected: %v", err)
}

func TestAdamicRuleModule(t *testing.T) {
	t.Parallel()
	root := copyRules(t)
	directory := filepath.Join(root, "rules/no-debugger")
	if err := os.Rename(filepath.Join(directory, "rule.ts"), filepath.Join(directory, "rule.a")); err != nil {
		t.Fatal(err)
	}
	descriptors, err := Generate(root)
	if err != nil {
		t.Fatal(err)
	}
	ts, _ := Render(descriptors)
	if !bytes.Contains(ts, []byte("no-debugger/rule.a")) {
		t.Fatal(".a entry not registered")
	}
	if err := os.WriteFile(filepath.Join(directory, "rule.ts"), []byte("stale rename"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Generate(root); err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("ambiguous module survived: %v", err)
	}
	t.Log("stale .ts alongside .a rejected")
}

// Every rule has a witness, and a rule whose witnesses are gone is refused by name. The check was Discover's, which
// every lint test process ran; it is this test's now, and lint-registry's (#s57yq6y).
func TestEveryRuleHasAWitness(t *testing.T) {
	t.Parallel()
	// Reads: stage1/cohere/lint/rules/*
	descriptors, err := Discover("..")
	if err != nil {
		t.Fatal(err)
	}
	// Reads: stage1/cohere/lint/rules/*
	if err := CheckWitnesses("..", descriptors); err != nil {
		t.Fatal(err)
	}
	root := copyRules(t)
	if err := os.RemoveAll(filepath.Join(root, "rules/no-debugger/testdata")); err != nil {
		t.Fatal(err)
	}
	descriptors, err = Discover(root)
	if err != nil {
		t.Fatal(err)
	}
	err = CheckWitnesses(root, descriptors)
	if err == nil || !strings.Contains(err.Error(), "no-debugger/rule.json: missing witness") {
		t.Fatalf("a rule without witnesses survived: %v", err)
	}
	t.Logf("%d rules with witnesses; a rule without one refused: %v", len(descriptors), err)
}
