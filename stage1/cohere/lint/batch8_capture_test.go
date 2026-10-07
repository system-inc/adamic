package lint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func batch8Upstream(t *testing.T) []string {
	root, err := filepath.Abs(filepath.Join(repository, "cohere"))
	if err != nil {
		t.Fatal(err)
	}
	harness := filepath.Join(root, "internal/lint/testing/rule_testing.go")
	data, err := os.ReadFile(harness)
	if err != nil {
		t.Fatal(err)
	}
	original := "return Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}"
	replacement := "result := Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}\n RecordAssertedCase(t, result)\n return result"
	if strings.Count(string(data), original) != 1 {
		t.Fatal("capture overlay anchor changed")
	}
	directory := t.TempDir()
	side := filepath.Join(directory, "rule_testing.go")
	if err := os.WriteFile(side, []byte(strings.Replace(string(data), original, replacement, 1)), 0644); err != nil {
		t.Fatal(err)
	}
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{harness: side}})
	overlayPath := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	capture := filepath.Join(directory, "capture")
	execute(t, root, "env", "COHERE_DOCS_CAPTURE="+capture, "go", "test", "-overlay="+overlayPath, "./internal/lint/rules/core", "./internal/lint/rules/react", "-run", "Test(NoOctalEscape|NoUnexpectedMultiline|NoUnusedPrivateClassMembers|NoUselessConstructor|PreferTemplate|ForwardRefUsesRef|JsxNoCommentTextnodes|NoFindDOMNode|NoIsMounted|NoRedundantShouldComponentUpdate)", "-count=1", "-timeout=10m")
	files, err := filepath.Glob(filepath.Join(capture, "*.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	type record struct {
		Rule, File, Source, Outcome, FixedSource string
		Options                                  struct {
			Mode, Null      string
			AllowEmptyCatch bool
		}
	}
	unique := map[string]record{}
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, line := range bytes.Split(data, []byte("\n")) {
			if len(line) == 0 {
				continue
			}
			var row record
			if err := json.Unmarshal(line, &row); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains("|no-octal-escape|no-unexpected-multiline|no-unused-private-class-members|no-useless-constructor|prefer-template|react/forward-ref-uses-ref|react/jsx-no-comment-textnodes|react/no-find-dom-node|react/no-is-mounted|react/no-redundant-should-component-update|", "|"+row.Rule+"|") {
				continue
			}
			key := fmt.Sprintf("%s\t%s\t%+v\t%s", row.Rule, row.File, row.Options, row.Source)
			unique[key] = row
		}
	}
	var keys []string
	for key := range unique {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var rows []string
	for i, key := range keys {
		row := unique[key]
		path := filepath.Join(directory, fmt.Sprintf("case-%03d%s", i, filepath.Ext(row.File)))
		if err := os.WriteFile(path, []byte(row.Source), 0644); err != nil {
			t.Fatal(err)
		}
		rows = append(rows, fmt.Sprintf("%s\t%s\t%s\t%s\t%t", path, row.Rule, row.Options.Mode, row.Options.Null, row.Options.AllowEmptyCatch))
	}
	if len(rows) != 714 {
		t.Fatalf("capture unexpectedly small: %d cases", len(rows))
	}
	t.Logf("batch 8 cohere cases: %d unique source/rule/file/options combinations", len(rows))
	data, err = json.MarshalIndent(unique, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "records.json"), data, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "manifest.txt"), []byte(strings.Join(rows, "\n")+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	return rows
}

// Not parallel: a capture invokes the upstream tests and records their asserted inputs.
func TestBatch8Capture(t *testing.T) { batch8Upstream(t) }
