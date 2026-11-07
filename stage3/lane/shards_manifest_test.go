package lane

import (
	"context"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestShardsMatchManifest(test *testing.T) {
	test.Parallel()
	root := repositoryRoot(test)
	timeoutContext, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	command := boundedCommand(timeoutContext, root, "python3", "-B", "stage3/lane/generate_shards.py", "--check")
	if output, combinedOutputError := command.CombinedOutput(); combinedOutputError != nil {
		test.Fatalf("manifest generation: %v: %s", combinedOutputError, output)
	}
	data, readError := os.ReadFile(filepath.Join(root, "stage3/lane/shards.json"))
	if readError != nil {
		test.Fatal(readError)
	}
	var manifest struct {
		Units []unit `json:"units"`
	}
	if decodeError := json.Unmarshal(data, &manifest); decodeError != nil {
		test.Fatal(decodeError)
	}
	var expected, actual []string
	for _, row := range manifest.Units {
		parts := strings.Split(row.Name, "-")
		for index, part := range parts {
			parts[index] = strings.ToUpper(part[:1]) + part[1:]
		}
		expected = append(expected, "TestShard_"+strings.Join(parts, ""))
	}
	files, globError := filepath.Glob(filepath.Join(root, "stage3/lane/*_test.go"))
	if globError != nil {
		test.Fatal(globError)
	}
	for _, path := range files {
		source, parseFileError := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if parseFileError != nil {
			test.Fatal(parseFileError)
		}
		for _, declaration := range source.Decls {
			if function, ok := declaration.(*ast.FuncDecl); ok &&
				function.Recv == nil &&
				strings.HasPrefix(function.Name.Name, "TestShard_") {
				actual = append(actual, function.Name.Name)
			}
		}
	}
	sort.Strings(expected)
	sort.Strings(actual)
	if !reflect.DeepEqual(expected, actual) {
		test.Fatalf("manifest/test names differ: want %v, got %v", expected, actual)
	}
}
