package lower

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

func TestPrimitiveIdentityCheckerProof(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		declaration string
		primitive   bool
	}{
		{"type Target = number;", true}, {"type Target = string;", true}, {"type Target = boolean;", true}, {"type Target = bigint;", true}, {"type Target = undefined;", true}, {"type Target = null;", true},
		{"type Target = number|string|boolean|bigint|undefined|null;", true},
		{"type Target = number|object;", false}, {"type Target = string|(()=>void);", false},
		{"type Target = object;", false}, {"type Target = {};", false}, {"type Target = unknown;", false}, {"type Target = any;", false},
	} {
		path := filepath.Join(t.TempDir(), "primitive.a")
		if err := os.WriteFile(path, []byte(test.declaration), 0600); err != nil {
			t.Fatal(err)
		}
		program, err := load.Load([]string{path})
		if err != nil {
			t.Fatal(err)
		}
		file := program.Files()[0]
		checked, release := program.Checker(context.Background(), file)
		l := &lowering{program: program, checker: checked}
		got := l.primitiveIdentityType(checked.GetTypeAtLocation(file.Statements.Nodes[0].Name()))
		release()
		if got != test.primitive {
			t.Fatalf("%s primitive=%t, want %t", test.declaration, got, test.primitive)
		}
	}
}
