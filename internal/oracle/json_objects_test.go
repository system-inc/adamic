package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// JSON.stringify of an object reference (#d9ep7de) prints the allocation, through json_objects.c's provider:
// its own names in Node's order and each slot by the representation byte its last write left, never by the
// declared type. Each fixture is held to Node on native sanitized and released, the JavaScript backend, wasm32
// and the counted build. Mutants: walking the shape's names in declaration order instead of
// adamic_object_ordered reds view_order (its index key comes first on Node); reading every null pointer as null
// reds nullish (its undefined fields print null).
var jsonObjectsFixtures = []string{"literal", "class", "nested", "optional", "view_order", "hidden", "nullish"}

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

// What JSON.stringify of an object reference can reach and can't print as Node does stops before it lowers,
// each road in its own words, pinned whole in its .refused file: a toJSON method, one a base class declares, an
// own toJSON spelled as a computed key or written through a view the allocation can be seen as, a Date and a
// Map, a class itself, a Weak link and a graph type. Mutant: jsonReach asking only the leaf class's own members
// for toJSON (its declaration's symbol table) instead of GetPropertyOfType lowers to_json_inherited.a.
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

// What only the allocation can say stops at run time, natively, in the words its .stopped file pins, after
// printing what Node printed before it; the JavaScript backend agrees with Node. null_variable.a's field holds a
// null pointer a variable wrote, whose byte says Object, so native can't tell null from undefined there.
func TestJSONObjectsStopAtRunTime(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"null_variable"} {
		for _, target := range []string{"sanitized", "released", "wasm32", "javascript"} {
			t.Run(name+"/"+target, func(t *testing.T) {
				t.Parallel()
				path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/json_objects_stopped", name+".a"))
				if err != nil {
					t.Fatal(err)
				}
				program, err := lowered(t, path)
				if err != nil {
					t.Fatal(err)
				}
				want := onNodeWith(t, inputRun{}, path)
				if target == "javascript" {
					script := filepath.Join(t.TempDir(), "program.mjs")
					if err := os.WriteFile(script, []byte(javascript.JavaScript(program)), 0600); err != nil {
						t.Fatal(err)
					}
					if difference := disagreement(want, onNodeWith(t, inputRun{}, script)); difference != "" {
						t.Fatal(difference)
					}
					return
				}
				stopped, err := os.ReadFile(strings.TrimSuffix(path, ".a") + ".stopped")
				if err != nil {
					t.Fatal(err)
				}
				options := native.Options{Sanitize: target == "sanitized"}
				if target == "wasm32" {
					if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
						t.Skip("set ADAMIC_ORACLE_WASI=1")
					}
					options.Target = "wasm32-wasi"
				}
				binary := filepath.Join(t.TempDir(), "program")
				if err := native.Build(native.C(program), binary, options); err != nil {
					t.Fatal(err)
				}
				var got run
				if target == "wasm32" {
					got = execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/wasi.mjs"), binary)
				} else {
					got = executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary)
				}
				message := strings.TrimSpace(string(stopped))
				if got.exitCode != 70 || !strings.Contains(string(got.stderr), message) || !strings.HasPrefix(string(want.stdout), string(got.stdout)) || len(got.stdout) == 0 {
					t.Fatalf("want exit 70 with %q after part of Node's %q: exit %d, stdout %q, stderr %q", message, want.stdout, got.exitCode, got.stdout, got.stderr)
				}
			})
		}
	}
}
