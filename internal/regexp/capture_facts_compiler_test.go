package regexp

import (
	"encoding/json"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	tsparser "github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// This fixed checker contract is read by the test, not a measurement dump.
// Parse every pinned compiler source again so omissions and new literals fail.
// Not parallel: writes testdata/capture-tsc.json when ADAMIC_UPDATE_CAPTURE_TABLE=1.
func TestCaptureFactsCompilerTable(t *testing.T) {
	data, err := os.ReadFile("testdata/capture-tsc.json")
	if err != nil {
		t.Fatal(err)
	}
	var table []captureLiteral
	if err = json.Unmarshal(data, &table); err != nil {
		t.Fatal(err)
	}
	// Reads: cohere
	root := "../../cohere/TypeScript/tsc/testdata/fixtures/compiler"
	var actual []captureLiteral
	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".ts") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		text := string(data)
		absolute, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		file := tsparser.ParseSourceFile(ast.SourceFileParseOptions{FileName: tspath.RootedFilePathFromAbsolute(absolute)}, text, core.ScriptKindTS)
		var visit func(*ast.Node) bool
		visit = func(n *ast.Node) bool {
			if n.Kind == ast.KindRegularExpressionLiteral {
				literal := n.Text()
				end := strings.LastIndex(literal, "/")
				facts, err := CaptureFacts(literal[1:end], literal[end+1:])
				if err != nil {
					t.Errorf("%s: %v", path, err)
				}
				pos := strings.Index(text[n.Pos():], literal) + n.Pos()
				actual = append(actual, captureLiteral{File: strings.TrimPrefix(filepath.ToSlash(path), "../../"), Line: strings.Count(text[:pos], "\n") + 1, Pattern: literal[1:end], Flags: literal[end+1:], Facts: facts})
			}
			n.ForEachChild(visit)
			return false
		}
		visit(file.AsNode())
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if os.Getenv("ADAMIC_UPDATE_CAPTURE_TABLE") == "1" {
		data, err := json.MarshalIndent(actual, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile("testdata/capture-tsc.json", append(data, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
		return
	}
	if !reflect.DeepEqual(actual, table) {
		t.Fatal("compiler literal facts changed; regenerate and review capture-tsc.json")
	}
	if len(actual) != 88 {
		t.Fatalf("literal coverage: %d", len(actual))
	}
	t.Logf("checked %d compiler literal sites", len(actual))
}
