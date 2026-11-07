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
func copyRules(test *testing.T) string {
	test.Helper()
	root := test.TempDir()
	// Reads: stage1/cohere/lint/rules/*
	walkError := filepath.WalkDir("../rules", func(path string, entry os.DirEntry, pathError error) error {
		if pathError != nil {
			return pathError
		}
		relative, relativePathError := filepath.Rel("..", path)
		if relativePathError != nil {
			return relativePathError
		}
		target := filepath.Join(root, relative)
		if entry.IsDir() {
			if parts := strings.Split(filepath.ToSlash(relative), "/"); len(parts) == 3 && parts[2] != "testdata" {
				return filepath.SkipDir
			}
			return os.MkdirAll(target, 0755)
		}
		data, readError := os.ReadFile(path)
		if readError != nil {
			return readError
		}
		return os.WriteFile(target, data, 0644)
	})
	if walkError != nil {
		test.Fatal(walkError)
	}
	return root
}

func TestDescriptorRejections(test *testing.T) {
	test.Parallel()
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
		test.Run(change.name, func(test *testing.T) {
			test.Parallel()
			root := copyRules(test)
			path := filepath.Join(root, "rules/no-debugger/rule.json")
			data, readError := os.ReadFile(path)
			if readError != nil {
				test.Fatal(readError)
			}
			var d map[string]any
			if decodeError := json.Unmarshal(data, &d); decodeError != nil {
				test.Fatal(decodeError)
			}
			change.mutate(d)
			data, encodeError := json.Marshal(d)
			if encodeError != nil {
				test.Fatal(encodeError)
			}
			if writeError := os.WriteFile(path, data, 0644); writeError != nil {
				test.Fatal(writeError)
			}
			_, generateError := Generate(root)
			if generateError == nil || !strings.Contains(generateError.Error(), change.want) {
				test.Fatalf("mutant survived or wrong failure: %v", generateError)
			}
			test.Logf("mutant rejected: %v", generateError)
		})
	}
}

func TestDuplicateOracleAdapter(test *testing.T) {
	test.Parallel()
	root := copyRules(test)
	path := filepath.Join(root, "rules/no-empty/oracle.go")
	data, readError := os.ReadFile(path)
	if readError != nil {
		test.Fatal(readError)
	}
	if writeError := os.WriteFile(
		path,
		bytes.ReplaceAll(data, []byte("oracleNoEmpty"), []byte("oracleNoDebugger")),
		0644,
	); writeError != nil {
		test.Fatal(writeError)
	}
	path = filepath.Join(root, "rules/no-empty/rule.json")
	data, readError = os.ReadFile(path)
	if readError != nil {
		test.Fatal(readError)
	}
	if writeError := os.WriteFile(
		path,
		bytes.ReplaceAll(data, []byte("oracleNoEmpty"), []byte("oracleNoDebugger")),
		0644,
	); writeError != nil {
		test.Fatal(writeError)
	}
	_, generateError := Generate(root)
	if generateError == nil || !strings.Contains(generateError.Error(), "duplicate oracle adapter") {
		test.Fatalf("duplicate adapter mutant survived: %v", generateError)
	}
	test.Logf("mutant rejected: %v", generateError)
}

func TestAdamicRuleModule(test *testing.T) {
	test.Parallel()
	root := copyRules(test)
	directory := filepath.Join(root, "rules/no-debugger")
	if renameError := os.Rename(
		filepath.Join(directory, "rule.ts"),
		filepath.Join(directory, "rule.a"),
	); renameError != nil {
		test.Fatal(renameError)
	}
	descriptors, generateError := Generate(root)
	if generateError != nil {
		test.Fatal(generateError)
	}
	ts, _ := Render(descriptors)
	if !bytes.Contains(ts, []byte("no-debugger/rule.a")) {
		test.Fatal(".a entry not registered")
	}
	if writeError := os.WriteFile(
		filepath.Join(directory, "rule.ts"),
		[]byte("stale rename"),
		0644,
	); writeError != nil {
		test.Fatal(writeError)
	}
	if _, generateError := Generate(
		root,
	); generateError == nil || !strings.Contains(generateError.Error(), "ambiguous") {
		test.Fatalf("ambiguous module survived: %v", generateError)
	}
	test.Log("stale .ts alongside .a rejected")
}

// Every rule has a witness, and a rule whose witnesses are gone is refused by name. The check was Discover's, which
// every lint test process ran; it is this test's now, and lint-registry's (#s57yq6y).
func TestEveryRuleHasAWitness(test *testing.T) {
	test.Parallel()
	// Reads: stage1/cohere/lint/rules/*
	descriptors, discoverError := Discover("..")
	if discoverError != nil {
		test.Fatal(discoverError)
	}
	// Reads: stage1/cohere/lint/rules/*
	if checkWitnessesError := CheckWitnesses("..", descriptors); checkWitnessesError != nil {
		test.Fatal(checkWitnessesError)
	}
	root := copyRules(test)
	if removeError := os.RemoveAll(filepath.Join(root, "rules/no-debugger/testdata")); removeError != nil {
		test.Fatal(removeError)
	}
	descriptors, discoverError = Discover(root)
	if discoverError != nil {
		test.Fatal(discoverError)
	}
	checkWitnessesError := CheckWitnesses(root, descriptors)
	if checkWitnessesError == nil ||
		!strings.Contains(checkWitnessesError.Error(), "no-debugger/rule.json: missing witness") {
		test.Fatalf("a rule without witnesses survived: %v", checkWitnessesError)
	}
	test.Logf("%d rules with witnesses; a rule without one refused: %v", len(descriptors), checkWitnessesError)
}
