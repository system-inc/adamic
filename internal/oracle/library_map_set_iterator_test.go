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

func TestLibraryMapSetIteratorCopiesRefused(test *testing.T) {
	test.Parallel()
	for _, collection := range []string{"map", "set"} {
		for _, part := range []string{"keys", "values", "entries"} {
			for _, copy := range []string{"spread", "assign"} {
				name := collection + "_" + part + "_" + copy
				test.Run(name, func(test *testing.T) {
					test.Parallel()
					path, err := filepath.Abs(filepath.Join(
						repository,
						"internal/oracle/testdata/library_map_set_iterator_refuse/library_map_set_iterator_"+name+".a",
					))
					if err != nil {
						test.Fatal(err)
					}
					// Source Node is the independent oracle: inherited next is absent from the copy.
					node := onNode(test, path)
					if node.exitCode == 0 || !strings.Contains(string(node.stderr), "copy.next is not a function") {
						test.Fatalf("Node must reject copy.next: exit %d, stderr %s", node.exitCode, node.stderr)
					}
					_, err = lowered(test, path)
					var refusal *lower.NotYet
					if !errors.As(err, &refusal) ||
						!strings.Contains(err.Error(), "next is inherited, not an own enumerable field") {
						test.Fatalf("want inherited-next refusal, got %v", err)
					}
				})
			}
		}
	}
}
