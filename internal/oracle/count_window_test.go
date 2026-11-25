package oracle

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/native_runtime"
)

// A loop's body must not allocate inside its period, and the counted build's one tally at exit can't say
// whether one does. adamic_count_snapshot (count.h) can: these splice two snapshots around the one call of
// a program's measured, as TestStringLengthJoinMemory splices its read before main returns, and print the
// difference, the allocations, frees and growths that call made (#42q54q5). It's a measurement, never a
// proof; the bounded marker (#td8v3ve) is the proof. Every number pinned here is exact, derived from the
// runtime beside it, and each program runs as it's built, untouched, agreeing with Node, and spliced, still
// printing what Node prints and letting go of everything.
//
// Mutant: count.h's ADAMIC_COUNT_GROWTH defined as ((void)0) under ADAMIC_COUNT, so no growth is ever
// counted. "push" and "map" go red, each reading 0 growths; "objects" and "arithmetic", whose growths are
// 0 either way, stay green (measure/42q54q5/proof.sh plants it).

// countWindowUnpinned is a column a case doesn't pin: its derivation would lean on more than the runtime.
const countWindowUnpinned = -1

var countWindowCases = []struct {
	name, source                string
	allocations, frees, growths int
}{
	// Each pass makes one Point, one adamic_object_new and so one adamic_allocate (object.c), and a literal
	// is reused in place only from a spread (reuse.go), so five passes are five allocations. An object's
	// slots are inline, with no buffer to grow. Its frees aren't pinned: when each is let go is lowering's.
	{"objects", `interface Point { x: number; y: number }
let sum = 0;
function measured(): void {
  for (let index = 0; index < 5; index++) {
    const point: Point = { x: index, y: index * 2 };
    sum = sum + point.x + point.y;
  }
}
measured();
console.log(String(sum));
`, 5, countWindowUnpinned, 0},
	// Numbers held in globals, a number loop counter, and arithmetic: nothing is made, let go of or grown.
	{"arithmetic", `let total = 0;
const factor = 3;
function measured(): void {
  for (let index = 0; index < 1000; index++) {
    total = total + index * factor - (index % 7);
  }
}
measured();
console.log(String(total));
`, 0, 0, 0},
	// [] is adamic_array_make(0, ...), with no storage (emit_expressions.go's ArrayLiteral). push grows a
	// full array to 4 from nothing and doubles it after (array.c): the first push's realloc is the buffer's
	// first allocation, not a growth, and the fifth (4 to 8), the ninth (8 to 16) and the seventeenth (16
	// to 32) are, so twenty pushes are 3 growths.
	{"push", `let pushed = 0;
function measured(): void {
  const values: number[] = [];
  for (let index = 0; index < 20; index++) {
    values.push(index * 2);
  }
  pushed = values.length;
}
measured();
console.log(String(pushed));
`, countWindowUnpinned, countWindowUnpinned, 3},
	// new Map() is adamic_map_new, with no table (map.c). The first set rebuilds into 8 entries from
	// nothing, the table's first allocation and not a growth. A set into a full table whose entries are all
	// live doubles it (adamic_map_set: count * 2 > capacity), one growth per rebuild: the ninth set (8 to
	// 16) and the seventeenth (16 to 32), so seventeen keys are 2 growths.
	{"map", `let size = 0;
function measured(): void {
  const squares = new Map<number, number>();
  for (let index = 0; index < 17; index++) {
    squares.set(index, index * index);
  }
  size = squares.size;
}
measured();
console.log(String(size));
`, countWindowUnpinned, countWindowUnpinned, 2},
}

var countWindowLine = regexp.MustCompile(`(?m)^count window: allocations (\d+) frees (\d+) growths (\d+)$`)

func TestCountWindow(t *testing.T) {
	t.Parallel()
	for _, c := range countWindowCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "main.a")
			if err := os.WriteFile(path, []byte(c.source), 0600); err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			got, _ := natively(t, program)
			if difference := disagreement(truth, got); difference != "" {
				t.Fatalf("native: %s; Node %q; native %q; stderr %q", difference, truth.stdout, got.stdout, got.stderr)
			}
			code := native.C(program)
			symbol := emittedFunctionSymbol(t, code, "measured")
			// The one call is a statement of its own, the line's first: the definition and its prototype
			// start with their return type.
			call := regexp.MustCompile(`(?m)^[ \t]*((?:\(void\))?` + regexp.QuoteMeta(symbol) + `\([^;\n]*\);)`)
			calls := call.FindAllStringSubmatchIndex(code, -1)
			if len(calls) != 1 {
				named := regexp.MustCompile(`(?m)^.*` + regexp.QuoteMeta(symbol) + `.*$`).FindAllString(code, -1)
				t.Fatalf("want one call of %s, found %d; the lines naming it:\n%s", symbol, len(calls), strings.Join(named, "\n"))
			}
			start, end := calls[0][2], calls[0][3]
			// Braced, so the snapshots are a block's own wherever the call stands.
			code = "#include <stdio.h>\n" + code[:start] +
				"{ adamic_count_window count_window_before = adamic_count_snapshot(); " + code[start:end] +
				" adamic_count_window count_window_after = adamic_count_snapshot();" +
				" fprintf(stderr, \"count window: allocations %zu frees %zu growths %zu\\n\"," +
				" count_window_after.allocations - count_window_before.allocations," +
				" count_window_after.frees - count_window_before.frees," +
				" count_window_after.growths - count_window_before.growths); }" + code[end:]
			binary := filepath.Join(t.TempDir(), "counted")
			if err := native.Build(code, binary, native_runtime.Options{Count: true}); err != nil {
				t.Fatal(err)
			}
			ran := execute(t, binary)
			if ran.exitCode != truth.exitCode || string(ran.stdout) != string(truth.stdout) {
				t.Fatalf("the spliced build disagrees with Node: exit %d, stdout %q, stderr %q", ran.exitCode, ran.stdout, ran.stderr)
			}
			if report := leakcheck.Unbalanced(leakRun(ran)); report != "" {
				t.Fatal(report)
			}
			windows := countWindowLine.FindAllSubmatch(ran.stderr, -1)
			if len(windows) != 1 {
				t.Fatalf("want one count window, measured is called once; stderr %q", ran.stderr)
			}
			for index, column := range []struct {
				name string
				want int
			}{{"allocations", c.allocations}, {"frees", c.frees}, {"growths", c.growths}} {
				measured, err := strconv.Atoi(string(windows[0][index+1]))
				if err != nil {
					t.Fatalf("%s %q: %v", column.name, windows[0][index+1], err)
				}
				t.Logf("%s %d", column.name, measured)
				if column.want != countWindowUnpinned && measured != column.want {
					t.Errorf("%s across the call: %d, want %d", column.name, measured, column.want)
				}
			}
		})
	}
}
