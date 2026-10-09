package estree

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/edit"
)

// parseFromCorpusVariable names more roots for the proof below, separated by the path list separator: a real
// tree, such as ahra, read the way the fixtures are. Unset, the proof reads the repository's fixtures alone.
const parseFromCorpusVariable = "COHERE_PARSE_FROM_CORPUS"

// The formatter takes the fix engine's guard's tree rather than parsing every TypeScript file a second time
// (#dk2502g). This proves the two are one tree, on every TypeScript file of the fixtures and of any tree the
// variable names: the guard's tree, captured from the transform the engine really calls, converts to an ESTree
// equal, property for property and comment for comment, to the one a fresh parse converts to, and it is the
// tree ParseTypeScriptFrom takes rather than one it refuses and parses again. The mutant is a tree of other
// bytes: converted, it must differ, so the comparison could have failed, and ParseTypeScriptFrom must refuse it.
func TestTheGuardsTreeIsTheTreeTheFormatterWouldParse(t *testing.T) {
	t.Parallel()
	roots := []string{moduleRoot(t)}
	if extra := os.Getenv(parseFromCorpusVariable); extra != "" {
		roots = append(roots, filepath.SplitList(extra)...)
	}
	files := typeScriptFiles(roots)
	if len(files) < 50 {
		t.Fatalf("the corpus holds %d files, too few for agreement to mean anything", len(files))
	}

	compared, hashbangs, unparseable, mutantsCaught := 0, 0, 0, 0
	for _, fileName := range files {
		contents, err := os.ReadFile(fileName)
		if err != nil {
			t.Fatal(err)
		}
		text := string(contents)
		if strings.HasPrefix(text, "#!") {
			hashbangs++
			continue
		}
		if parses, _ := edit.Parses(fileName, text); !parses {
			// The engine refuses a file that does not parse before the transform is reached.
			unparseable++
			continue
		}

		var guards *ast.SourceFile
		capture := func(_ string, given string, parsed *ast.SourceFile) (string, error) {
			guards = parsed
			return given, nil
		}
		noProposals := func(string, string) ([]edit.Proposal, error) { return nil, nil }
		if _, err := edit.FixAndTransformText(fileName, text, noProposals, capture, 1); err != nil {
			t.Fatalf("%s: %v", fileName, err)
		}
		if guards == nil {
			t.Errorf("%s: the engine handed the transform no tree", fileName)
			continue
		}
		if !treeIsFor(fileName, text, guards) {
			t.Errorf("%s: ParseTypeScriptFrom refuses the guard's tree, so the file is still parsed twice", fileName)
			continue
		}

		fresh := convertedDump(t, fileName, ParseSourceFile(fileName, text))
		if taken := convertedDump(t, fileName, guards); taken != fresh {
			t.Errorf("%s: the guard's tree converts to another ESTree than a fresh parse does", fileName)
		}
		compared++

		// The mutant: the tree of the text with one more space in front, every offset one past the file's own.
		mutant := ParseSourceFile(fileName, " "+text)
		if treeIsFor(fileName, text, mutant) {
			t.Errorf("%s: ParseTypeScriptFrom would take a tree of other bytes", fileName)
		}
		if convertedDump(t, fileName, mutant) != fresh {
			mutantsCaught++
		}
	}
	t.Logf("%d TypeScript files compared, %d with a hashbang left to their own parse, %d that do not parse, %d mutants caught",
		compared, hashbangs, unparseable, mutantsCaught)
	if compared < 50 {
		t.Fatalf("only %d files were compared", compared)
	}
	if mutantsCaught != compared {
		t.Errorf("a tree of other bytes converted to the same ESTree on %d of %d files, so the comparison could not fail there",
			compared-mutantsCaught, compared)
	}
}

// convertedDump is the ESTree a tree converts to, with its comments, as JSON, postprocessed against the tree's
// own text. A file that fails to convert dumps its error, so two trees that fail alike compare equal and one that
// fails alone does not.
func convertedDump(t *testing.T, fileName string, sourceFile *ast.SourceFile) string {
	t.Helper()
	text := sourceFile.Text()
	program, comments, _, err := convertTypeScript(sourceFile, text, nil)
	if err != nil {
		return "error: " + err.Error()
	}
	offsets := make([]int, len(text)+1)
	for index := range offsets {
		offsets[index] = index
	}
	dumped := make([]any, len(comments))
	for index, comment := range comments {
		dumped[index] = map[string]any{"type": comment.Type(), "range": comment.Range, "value": comment.String("value")}
	}
	encoded, err := json.Marshal(map[string]any{"ast": dumpValue(program, offsets), "comments": dumped})
	if err != nil {
		t.Fatalf("%s: %v", fileName, err)
	}
	return string(encoded)
}

// moduleRoot is the module's root, where the fixtures are.
func moduleRoot(t *testing.T) string {
	t.Helper()
	directory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(directory, "go.mod")); err == nil {
			return directory
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			t.Fatal("no go.mod above the test's directory")
		}
		directory = parent
	}
}

// typeScriptFiles is every .ts, .tsx, .mts and .cts file under roots, outside dependency, build and cache
// directories and the TypeScript checkout.
func typeScriptFiles(roots []string) []string {
	extensions := map[string]bool{".ts": true, ".tsx": true, ".mts": true, ".cts": true}
	skipped := map[string]bool{"node_modules": true, ".git": true, ".cache": true, ".next": true, "dist": true, "TypeScript": true}
	var files []string
	for _, root := range roots {
		filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if entry.IsDir() {
				if path != root && skipped[entry.Name()] {
					return filepath.SkipDir
				}
				return nil
			}
			if extensions[strings.ToLower(filepath.Ext(path))] && entry.Type()&fs.ModeSymlink == 0 {
				files = append(files, path)
			}
			return nil
		})
	}
	return files
}
