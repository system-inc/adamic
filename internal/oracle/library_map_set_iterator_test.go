package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// Register this slice without changing the compiler-owned oracle dispatch file.
func init() {
	for _, name := range []string{"number_hash", "exhausted"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path: "internal/oracle/testdata/library_map_set_iterator_" + name + ".a", lowers: true})
	}
}

func TestLibraryMapSetIteratorCopiesRefused(t *testing.T) {
	t.Parallel()
	for _, collection := range []string{"map", "set"} {
		for _, part := range []string{"keys", "values", "entries"} {
			for _, copy := range []string{"spread", "assign"} {
				name := collection + "_" + part + "_" + copy
				t.Run(name, func(t *testing.T) {
					t.Parallel()
					path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_map_set_iterator_refuse/library_map_set_iterator_"+name+".a"))
					if err != nil {
						t.Fatal(err)
					}
					// Source Node is the independent oracle: inherited next is absent from the copy.
					node := onNode(t, path)
					if node.exitCode == 0 || !strings.Contains(string(node.stderr), "copy.next is not a function") {
						t.Fatalf("Node must reject copy.next: exit %d, stderr %s", node.exitCode, node.stderr)
					}
					_, err = lowered(t, path)
					var refusal *lower.NotYet
					if !errors.As(err, &refusal) || !strings.Contains(err.Error(), "next is inherited, not an own enumerable field") {
						t.Fatalf("want inherited-next refusal, got %v", err)
					}
				})
			}
		}
	}
}
