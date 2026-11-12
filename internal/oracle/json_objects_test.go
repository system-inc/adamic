package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// JSON.stringify of an object reference (#d9ep7de) prints the allocation, through json_objects.c's provider:
// its own names in Node's order and each slot by the representation byte its last write left, never by the
// declared type. Each fixture is held to Node on native sanitized and released, the JavaScript backend, wasm32
// and the counted build. Mutants: walking the shape's names in declaration order instead of
// adamic_object_ordered reds view_order (its index key comes first on Node); reading every null pointer as null
// reds nullish (its undefined fields print null); and reading a null pointer in a String slot as null reds
// nullish too (a string | undefined a variable wrote, as in { name: maybeName }, prints "name":null).
// Function fields are omitted and function array entries print null; nonfinite fields print null; decoded
// allocations print their fields even through a class field (#pmjwn79). Treating a closure as an empty
// object in json_stringify.c instead of a function reds functions.
// Dropping the nullable write branch reds null_variable/0; writing the null tag for both cases reds
// undefined_variable/0. Dropping the logical assignment byte update reds logical_and_null/0.
// Leaving the decoder's first slot without its representation byte in json_decode.c reds decoded/0,
// decoded/1 and decoded/4 (an implementation word).
var jsonObjectsFixtures = []string{"literal", "class", "nested", "optional", "view_order", "hidden", "nullish", "functions", "nonfinite", "decoded", "null_variable", "undefined_variable", "logical_and_null", "logical_coalesce_undefined"}

func TestJSONObjectsAgreeWithNode(t *testing.T) {
	t.Parallel()
	for _, name := range jsonObjectsFixtures {
		for backend := 0; backend < 5; backend++ {
			t.Run(name+"/"+strconv.Itoa(backend), func(t *testing.T) {
				t.Parallel()
				fixtureAgreesOnBackend(t, "internal/oracle/testdata/json_objects_"+name+".a", backend)
			})
		}
	}
}

// What JSON.stringify of an object reference can reach and cannot print as Node does
// stops before lowering: toJSON members, library values, a class itself, a weak
// link and a graph type. Nullable reference writes carry their own semantic tags.
// Asking only the leaf class's own members for toJSON lowers to_json_inherited.a.
func TestJSONObjectsRefused(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/json_objects_refused"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".a") {
			continue
		}
		programs++
		path := directory + "/" + name
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			absolute, err := filepath.Abs(filepath.Join(repository, path))
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowered(t, absolute)
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) {
				t.Fatalf("want NotYet, got %v", err)
			}
			expected, err := os.ReadFile(strings.TrimSuffix(absolute, ".a") + ".refused")
			if err != nil {
				t.Fatal(err)
			}
			got := strings.ReplaceAll(notYet.Error(), strings.TrimSuffix(absolute, path), "")
			if got != strings.TrimSpace(string(expected)) {
				t.Fatalf("diagnostic: got %q, want %q", got, strings.TrimSpace(string(expected)))
			}
		})
	}
	if programs != 9 {
		t.Fatalf("%d programs in %s, want 9: a program added or removed needs this count changed with it", programs, directory)
	}
}
