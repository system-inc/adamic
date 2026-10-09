package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func TestStep20IntrinsicIteratorWrites(t *testing.T) {
	t.Parallel()
	for _, receiver := range []string{"[1]", "new String('x')", "new Map<string, number>()", "new Set<number>()", "new Uint8Array(1)", "Array.prototype", "String.prototype", "Map.prototype", "Set.prototype", "Uint8Array.prototype"} {
		t.Run(receiver, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "write.a")
			source := "(" + receiver + ")[Symbol.iterator] = () => { throw new Error('replacement'); };\nconsole.log('written');\n"
			if err := os.WriteFile(path, []byte(source), 0644); err != nil {
				t.Fatal(err)
			}
			node := onNode(t, path)
			if node.exitCode != 0 || len(node.stderr) != 0 || string(node.stdout) != "written\n" {
				t.Fatalf("Node: %+v", node)
			}
			_, err := lowered(t, path)
			var refused *lower.Refused
			if !errors.As(err, &refused) || refused.What != "writing Symbol.iterator on a built-in or its prototype (adamic/intrinsic-iterator)" {
				t.Fatalf("want ruled intrinsic iterator refusal, got %v", err)
			}
		})
	}
	for _, name := range []string{"intrinsic_iterator_write.a", "intrinsic_iterator_alias.a", "intrinsic_iterator_prototype.a"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := lowered(t, filepath.Join(repository, "stage3/fixtures/iteration", name))
			var refused *lower.Refused
			if !errors.As(err, &refused) || refused.What != "writing Symbol.iterator on a built-in or its prototype (adamic/intrinsic-iterator)" {
				t.Fatalf("want ruled intrinsic iterator refusal, got %v", err)
			}
		})
	}
}
