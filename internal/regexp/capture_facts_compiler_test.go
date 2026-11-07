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
func TestCaptureFactsCompilerTable(test *testing.T) {
	data, readError := os.ReadFile("testdata/capture-tsc.json")
	if readError != nil {
		test.Fatal(readError)
	}
	var table []captureLiteral
	if decodeError := json.Unmarshal(data, &table); decodeError != nil {
		test.Fatal(decodeError)
	}
	// Reads: cohere
	root := "../../cohere/TypeScript/tsc/testdata/fixtures/compiler"
	var actual []captureLiteral
	walkError := filepath.WalkDir(root, func(path string, entry os.DirEntry, pathError error) error {
		if pathError != nil {
			return pathError
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".ts") {
			return nil
		}
		data, readError := os.ReadFile(path)
		if readError != nil {
			return readError
		}
		text := string(data)
		absolute, absolutePathError := filepath.Abs(path)
		if absolutePathError != nil {
			return absolutePathError
		}
		file := tsparser.ParseSourceFile(
			ast.SourceFileParseOptions{FileName: tspath.RootedFilePathFromAbsolute(absolute)},
			text,
			core.ScriptKindTS,
		)
		var visit func(*ast.Node) bool
		visit = func(node *ast.Node) bool {
			if node.Kind == ast.KindRegularExpressionLiteral {
				literal := node.Text()
				end := strings.LastIndex(literal, "/")
				facts, captureFactsError := CaptureFacts(literal[1:end], literal[end+1:])
				if captureFactsError != nil {
					test.Errorf("%s: %v", path, captureFactsError)
				}
				pos := strings.Index(text[node.Pos():], literal) + node.Pos()
				actual = append(actual, captureLiteral{
					File:    strings.TrimPrefix(filepath.ToSlash(path), "../../"),
					Line:    strings.Count(text[:pos], "\n") + 1,
					Pattern: literal[1:end],
					Flags:   literal[end+1:],
					Facts:   facts,
				})
			}
			node.ForEachChild(visit)
			return false
		}
		visit(file.AsNode())
		return nil
	})
	if walkError != nil {
		test.Fatal(walkError)
	}
	if os.Getenv("ADAMIC_UPDATE_CAPTURE_TABLE") == "1" {
		data, encodeError := json.MarshalIndent(actual, "", "  ")
		if encodeError != nil {
			test.Fatal(encodeError)
		}
		if writeError := os.WriteFile("testdata/capture-tsc.json", append(data, '\n'), 0644); writeError != nil {
			test.Fatal(writeError)
		}
		return
	}
	if !reflect.DeepEqual(actual, table) {
		test.Fatal("compiler literal facts changed; regenerate and review capture-tsc.json")
	}
	if len(actual) != 88 {
		test.Fatalf("literal coverage: %d", len(actual))
	}
	test.Logf("checked %d compiler literal sites", len(actual))
}
