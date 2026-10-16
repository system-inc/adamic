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

func TestShardsMatchManifest(t *testing.T) {
	t.Parallel()
	root := repositoryRoot(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	command := boundedCommand(ctx, root, "python3", "-B", "stage3/lane/generate_shards.py", "--check")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("manifest generation: %v: %s", err, output)
	}
	data, err := os.ReadFile(filepath.Join(root, "stage3/lane/shards.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Units []unit `json:"units"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	var expected, actual []string
	for _, row := range manifest.Units {
		parts := strings.Split(row.Name, "-")
		for i, part := range parts {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
		expected = append(expected, "TestShard_"+strings.Join(parts, ""))
	}
	files, err := filepath.Glob(filepath.Join(root, "stage3/lane/*_test.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range files {
		source, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, declaration := range source.Decls {
			if function, ok := declaration.(*ast.FuncDecl); ok && function.Recv == nil && strings.HasPrefix(function.Name.Name, "TestShard_") {
				actual = append(actual, function.Name.Name)
			}
		}
	}
	sort.Strings(expected)
	sort.Strings(actual)
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("manifest/test names differ: want %v, got %v", expected, actual)
	}
}
