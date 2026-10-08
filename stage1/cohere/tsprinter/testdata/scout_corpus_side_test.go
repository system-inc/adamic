package native

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/format/estree"
	"github.com/system-inc/cohere/internal/format/formatoptions"
	"os"
	"path/filepath"
	"testing"
)

func TestScoutFormatterCorpus(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile(os.Getenv("ADAMIC_SCOUT_CORPUS"))
	if err != nil {
		t.Fatal(err)
	}
	var request struct {
		Files []struct {
			Repo, Pin, Path, SHA256, Outcome string
			Answer                           *string
		}
		Checkouts map[string]string
	}
	if err = json.Unmarshal(data, &request); err != nil {
		t.Fatal(err)
	}
	if len(request.Files) == 0 {
		t.Fatal("empty corpus")
	}
	output, err := os.Create(os.Getenv("ADAMIC_SCOUT_CORPUS_ANSWERS"))
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	compressed := gzip.NewWriter(output)
	defer compressed.Close()
	encoder := json.NewEncoder(compressed)
	for _, item := range request.Files {
		path := filepath.Join(request.Checkouts[item.Repo+"@"+item.Pin], item.Path)
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		digest := sha256.Sum256(source)
		if hex.EncodeToString(digest[:]) != item.SHA256 {
			t.Fatalf("source changed: %s", path)
		}
		kinds := map[string]int{}
		formatted, formatErr := func() (text string, failure error) {
			defer func() {
				if value := recover(); value != nil {
					failure = fmt.Errorf("Go panic: %v", value)
				}
			}()
			file := estree.ParseSourceFile(path, string(source))
			var visit func(*ast.Node) bool
			visit = func(node *ast.Node) bool { kinds[node.Kind.String()]++; node.ForEachChild(visit); return false }
			visit(file.AsNode())
			return (Formatter{Options: formatoptions.Default()}).Format(path, string(source))
		}()
		result := map[string]any{"repo": item.Repo, "pin": item.Pin, "path": item.Path, "kinds": kinds}
		if formatErr != nil {
			result["error"] = formatErr.Error()
		} else {
			digest := sha256.Sum256([]byte(formatted))
			result["sha256"] = hex.EncodeToString(digest[:])
			if item.Answer != nil && *item.Answer != formatted {
				result["mismatch"] = true
			}
		}
		if err = encoder.Encode(result); err != nil {
			t.Fatal(err)
		}
	}
}
