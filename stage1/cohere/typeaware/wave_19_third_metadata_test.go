package typeaware

import (
	"encoding/json"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"
)

func TestWave19ThirdNumericMetadata(t *testing.T) {
	root := os.Getenv("ADAMIC_WAVE19_THIRD_METADATA")
	if root == "" {
		root = "wave_19_third"
	}
	rows := []struct {
		name  string
		kinds []int
	}{{"symbol-description", []int{int(ast.KindCallExpression)}}, {"valid-typeof", []int{int(ast.KindBinaryExpression)}}, {"require-await", []int{int(ast.KindFunctionDeclaration), int(ast.KindFunctionExpression), int(ast.KindArrowFunction), int(ast.KindMethodDeclaration), int(ast.KindGetAccessor), int(ast.KindSetAccessor), int(ast.KindConstructor)}}}
	pattern := regexp.MustCompile(`readonly listenerKinds: readonly number\[\]=([\[0-9,]+\]);`)
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			data, e := os.ReadFile(filepath.Join(root, row.name, "rule.json"))
			if e != nil {
				t.Fatal(e)
			}
			var metadata struct{ Kinds []int }
			if e = json.Unmarshal(data, &metadata); e != nil {
				t.Fatal(e)
			}
			if !reflect.DeepEqual(metadata.Kinds, row.kinds) {
				t.Fatalf("numeric metadata differs: %v want %v", metadata.Kinds, row.kinds)
			}
			data, e = os.ReadFile(filepath.Join(root, row.name, "rule.a"))
			if e != nil {
				t.Fatal(e)
			}
			found := pattern.FindSubmatch(data)
			if len(found) != 2 {
				t.Fatal("missing numeric listener declaration")
			}
			var actual []int
			if e = json.Unmarshal(found[1], &actual); e != nil {
				t.Fatal(e)
			}
			if !reflect.DeepEqual(actual, row.kinds) {
				t.Fatalf("listener differs: %v want %v", actual, row.kinds)
			}
		})
	}
}
