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
func TestCapturedUnionLowersToOneCell(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ fixture, name string }{
		{"class", "active"},
		{"object", "active"},
		{"by_value", "active"},
		{"released", "active"},
		{"nested", "extended"},
		{"alternating", "active"},
		{"graph", "active"},
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
					if probe.fixture == "graph" && !local.GraphCell {
						t.Errorf("%s: not a graph cell, though the methods that capture it are on a cycle", local.Name)
					}
					closures := 0
					for _, function := range program.Functions {
						if slices.Contains(function.Environment, index) {
							closures++
						}
					}
					if closures < 2 {
						t.Errorf("%s: %d closures carry its cell, want every one that reads or writes it", local.Name, closures)
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
	var notYet *NotYet
	if !errors.As(err, &notYet) || err.Error() != "main.a:4:3: stage 0 can't lower a boolean | undefined variable a function value captures yet" {
		t.Fatalf("got %v, want main's stop at 4:3", err)
	}
}
