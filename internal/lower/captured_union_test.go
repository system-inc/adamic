package lower

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// A captured let of differently held members lowers to one cell holding the Union, the same payload
// an uncaptured one has and a placeholder one already had: no placeholder, and every closure that
// reads or writes it carries that one cell (#n7sx9gz). The oracle's TestCapturedUnionAgrees runs them.
// closures counts the function values in the fixture that reach the cell, each of which must carry
// it. The declaring function's own reads and writes don't count, since it holds the cell itself.
func TestCapturedUnionLowersToOneCell(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		fixture, name string
		closures      int
	}{
		{"class", "active", 5},
		{"object", "active", 5},
		{"by_value", "active", 3},
		{"released", "active", 2},
		// assign writes the cell and onPropertySet, the callback it's nested in, carries it there.
		{"nested", "extended", 2},
		{"alternating", "active", 2},
		{"graph", "active", 2},
		// report is the only closure: run writes the cell itself, as the function that declares it,
		// and each Session holds report, not a closure of its own (measured 1 on 54397e8f's proof).
		{"cycle", "active", 1},
	} {
		for _, extension := range []string{".ts", ".a"} {
			t.Run(probe.fixture+extension, func(t *testing.T) {
				t.Parallel()
				source, err := os.ReadFile(filepath.Join("..", "oracle", "testdata", "captured_union", probe.fixture+".ts"))
				if err != nil {
					t.Fatal(err)
				}
				program, err := lowerSourceAs(t, string(source), extension)
				if err != nil {
					t.Fatal(err)
				}
				found := false
				for index, local := range program.Locals {
					if local.Name != probe.name {
						continue
					}
					found = true
					if local.Type != ir.Union || !local.Captured || local.Placeholder != "" {
						t.Fatalf("%s: Type %v Captured %t Placeholder %q, want a captured Union that isn't a placeholder", local.Name, local.Type, local.Captured, local.Placeholder)
					}
					t.Logf("%s: graph cell %t", local.Name, local.GraphCell)
					if (probe.fixture == "graph" || probe.fixture == "cycle") && !local.GraphCell {
						t.Errorf("%s: not a graph cell, though it's on a cycle", local.Name)
					}
					closures := 0
					for _, function := range program.Functions {
						if slices.Contains(function.Environment, index) {
							closures++
						}
					}
					t.Logf("%s: %d closures carry its cell", local.Name, closures)
					if closures < probe.closures {
						t.Errorf("%s: %d closures carry its cell, want the %d that read or write it", local.Name, closures, probe.closures)
					}
				}
				if !found {
					t.Fatalf("no local named %s", probe.name)
				}
			})
		}
	}
}

// boolean | undefined has no word of its own in a cell yet, so a captured one still stops, with
// main's words.
func TestCapturedMaybeBooleanStillStops(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, "function run(): boolean {\n\tlet seen: boolean | undefined;\n\tconst mark = (): void => {\n\t\tseen = true;\n\t};\n\tmark();\n\treturn seen ?? false;\n}\nconsole.log(`${run()}`);\n")
	// Where carries the temporary directory the source was written to, so only its end is main's.
	var notYet *NotYet
	if !errors.As(err, &notYet) || filepath.Base(notYet.Where) != "main.a:4:3" || notYet.What != "a boolean | undefined variable a function value captures" {
		t.Fatalf("got %v, want main's stop at main.a:4:3", err)
	}
}
