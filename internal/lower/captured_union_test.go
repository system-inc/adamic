package lower

import (
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

// A captured boolean | undefined lowers to a cell holding adamic_value's maybe_boolean byte (#fds729j):
// the local keeps its MaybeBoolean type, and every closure that reads or writes it carries the cell.
// closures counts the function values in the fixture that name it. uninitialized declares two locals
// named seen, one in each function, and each must be carried by its closure.
func TestCapturedMaybeBooleanLowersToOneCell(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		fixture, name string
		closures      int
	}{
		{"uninitialized", "seen", 1},
		{"written", "seen", 2},
		// run reads the cell itself, as the function that declares it; write is the only closure.
		{"read_after", "seen", 1},
		{"optional_parameter", "recursive", 1},
		{"narrowed", "seen", 3},
		{"async", "seen", 1},
		{"async", "given", 1},
		{"reporter", "pretty", 2},
	} {
		for _, extension := range []string{".ts", ".a"} {
			t.Run(probe.fixture+"_"+probe.name+extension, func(t *testing.T) {
				t.Parallel()
				source, err := os.ReadFile(filepath.Join("..", "oracle", "testdata", "captured_maybe_boolean", probe.fixture+".ts"))
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
					if local.Type != ir.MaybeBoolean || !local.Captured {
						t.Fatalf("%s: Type %v Captured %t, want a captured boolean | undefined", local.Name, local.Type, local.Captured)
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
