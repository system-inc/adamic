package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// The loop counter sweep: for loops at every edge of what lower/counters.go keeps in an integer,
// counting up and down, by steps of one and more, from starts and to bounds computed by arithmetic,
// against Node. Each loop runs at most six passes (huge bounds break out), and prints how many it ran,
// the sum of its counter and of an array read by it, the sign of 1 / counter on its first pass (an
// integer has no -0) and its last value, so a counter that ran past 2^53, where JavaScript's + step
// rounds, or that lost a fraction or a sign, says so. Every loop is also checked to have been
// kept in an integer exactly when it should: the ones that fall back to a double must, and the ones
// that qualify mustn't miss out, so the sweep proves what it's sweeping.
// loopCounterCaseCount is audited by the gate enumerator and checked against the generator.
const loopCounterCaseCount = 1193

type loopCounterCase struct {
	name     string
	source   string
	expected map[string]bool
}

// loopCounterCases preserves the original sweep order and cN identifiers.
func loopCounterCases() []loopCounterCase {
	two32, two53 := 4294967296.0, 9007199254740992.0
	// Every start and bound with the range it can take, worked out by hand rather than by the pass's
	// own arithmetic, and whether it's one the pass may keep (whole within 2^53 at both ends and, for a
	// start, never -0).
	type value struct {
		source    string
		low, high float64
		whole     bool
	}
	starts := []value{
		{"0", 0, 0, true}, {"-0", 0, 0, false}, {"2", 2, 2, true}, {"-2", -2, -2, true}, {"0.5", 0, 0, false},
		{"9007199254740989", two53 - 3, two53 - 3, true}, {"-9007199254740989", 3 - two53, 3 - two53, true},
		{"9007199254740992", two53, two53, true},
		{"-9007199254740992", -two53, -two53, true}, {"9007199254740994", 0, 0, false}, {"2 ** 3", 0, 0, false},
		{"values.length - 1", -1, two32 - 1, true}, {"wholeConstant + 1", 1, two32 + 1, true},
		{"0 - values.length", -two32, 0, true}, {"-empty.length", 0, 0, false}, {"0 * -1", 0, 0, false},
		{"values.length * 2", 0, 0, false}, {"text.length / 2", 0, 0, false}, {"changing", 0, 0, false},
	}
	bounds := []value{
		{"values.length", 0, two32, true}, {"text.length", 0, two32, true}, {"sizes.size", 0, two32, true},
		{"wholeConstant", 0, two32, true}, {"5", 5, 5, true}, {"0", 0, 0, true}, {"-0", 0, 0, true}, {"-3", -3, -3, true},
		{"9007199254740991", two53 - 1, two53 - 1, true}, {"9007199254740992", two53, two53, true},
		{"-9007199254740992", -two53, -two53, true}, {"9007199254740994", 0, 0, false},
		{"2.5", 0, 0, false}, {"-0.5", 0, 0, false}, {"NaN", 0, 0, false}, {"Infinity", 0, 0, false},
		{"-Infinity", 0, 0, false}, {"fractionConstant", 0, 0, false}, {"changing", 0, 0, false}, {"2 ** 2", 0, 0, false},
		{"values.length - 1", -1, two32 - 1, true}, {"values.length + 2", 2, two32 + 2, true},
		{"values.length * 2", 0, 2 * two32, true}, {"-values.length", -two32, 0, true},
		{"wholeConstant - 1", -1, two32 - 1, true}, {"derived", -3, 2*two32 - 3, true},
		{"9007199254740991 - values.length", two53 - 1 - two32, two53 - 1, true},
		{"values.length - 9007199254740991", 1 - two53, two32 + 1 - two53, true},
		{"values.length + 9007199254740990", 0, 0, false}, {"text.length / 2", 0, 0, false},
		{"values.length % 3", 0, 0, false},
		{"9007199254740000 + values.length", 0, 0, false}, {"-9007199254740000 - values.length", 0, 0, false},
		{"values.length * -2097152", -two53, 0, true}, {"values.length * 2097153", 0, 0, false},
		{"values.length * values.length", 0, 0, false},
	}
	// The updates, each with its step (0 for one the pass never takes) and its comparison.
	type update struct {
		comparison, source string
		step               float64
	}
	directions := []update{{"<", "NAME++", 1}, {"<=", "NAME += 1", 1}, {">", "NAME--", -1}, {">=", "NAME -= 1", -1}}
	steps := []update{
		{"<", "NAME += 2", 2}, {"<=", "NAME = NAME + 3", 3}, {">", "NAME -= 2", -2}, {">=", "NAME = NAME - 3", -3},
		{"<", "NAME += 0", 0}, {"<", "NAME += 0.5", 0}, {"<", "NAME += 1.5", 0}, {">", "NAME -= 1.5", 0}, {">", "NAME -= -1", 0}, {"<", "NAME += 2 ** 1", 0},
		{"<", "NAME = 1 + NAME", 0}, {">", "NAME++", 1}, {"<", "NAME--", -1}, {">=", "NAME += 2", 2},
	}
	const prelude = (`const values = [10, 20, 30, 40, 50];
const empty: number[] = [];
const text = 'héllo';
const sizes = new Map<string, number>([['a', 1], ['b', 2], ['c', 3]]);
const wholeConstant = values.length;
const derived = values.length * 2 - 3;
const fractionConstant = 4.5;
let changing = 4;
changing = 6;
function report(name: string, seen: number, sum: number, sign: number, last: number): void {
	console.log(` + "`${name} ${seen} ${sum} ${sign} ${last}`" + `);
}
`)
	var result []loopCounterCase
	loop := func(name string, header string, extra string) {
		var source strings.Builder
		source.WriteString(prelude)
		fmt.Fprintf(&source, `function %s(): void {
	let seen = 0;
	let sum = 0;
	let sign = 0;
	let last = 0;
	for (%s) {
		if (seen === 0) {
			sign = 1 / %s;
		}
%s		last = %s;
		sum += %s + (values[%s] ?? 0.25);
		seen += 1;
		if (seen >= 6) {
			break;
		}
	}
	report('%s', seen, sum, sign, last);
}
%s();
`, name, header, name, extra, name, name, name, name, name)
		result = append(result, loopCounterCase{name, source.String(), map[string]bool{}})
	}
	// kept is the rule the pass follows, said again: the furthest a counter can go, one step past the
	// last value its bound lets through, counting toward the bound, is within 2^53. It's worked out in
	// integers, where 2^53 + 1 is still 2^53 + 1.
	kept := func(start value, bound value, each update) bool {
		if !start.whole || !bound.whole || each.step == 0 {
			return false
		}
		low, high, step := int64(bound.low), int64(bound.high), int64(each.step)
		var furthest int64
		switch {
		case step > 0 && each.comparison == "<":
			furthest = high - 1 + step
		case step > 0 && each.comparison == "<=":
			furthest = high + step
		case step < 0 && each.comparison == ">":
			furthest = low + 1 + step
		case step < 0 && each.comparison == ">=":
			furthest = low + step
		default:
			return false
		}
		return furthest <= 1<<53 && furthest >= -(1<<53)
	}
	cases := 0
	add := func(start value, bound value, each update) {
		cases++
		name := fmt.Sprintf("c%d", cases)
		header := fmt.Sprintf("let %s = %s; %s %s %s; %s", name, start.source, name, each.comparison, bound.source,
			strings.ReplaceAll(each.source, "NAME", name))
		loop(name, header, "")
		result[len(result)-1].expected[name] = kept(start, bound, each)
	}
	// Whether a loop qualifies turns on its start alone, and on its bound and step together, so the
	// sweep crosses each with the edges of the others rather than everything with everything, which
	// would build thousands of loops under the sanitizers: every start, counting up and down, against
	// two bounds; every bound against the starts that reach its edges; and every step against the
	// starts and bounds at its.
	edgeStarts := []value{starts[0], starts[5], starts[6], starts[11]}
	for _, start := range starts {
		for _, bound := range []value{bounds[0], bounds[7]} {
			for _, each := range directions {
				add(start, bound, each)
			}
		}
	}
	for _, start := range edgeStarts {
		for _, bound := range bounds {
			for _, each := range directions {
				add(start, bound, each)
			}
		}
	}
	edgeBounds := []value{bounds[0], bounds[8], bounds[9], bounds[10], bounds[14], bounds[20], bounds[26], bounds[27]}
	for _, each := range steps {
		for _, start := range edgeStarts {
			for _, bound := range edgeBounds {
				add(start, bound, each)
			}
		}
	}
	// The loops that must keep a double for what their bodies do.
	specials := []struct {
		header, extra string
		integer       bool
	}{
		{"let NAME = 0; NAME < 5; ++NAME", "", true},
		{"let NAME = 9; NAME > 0; --NAME", "", true},
		{"let NAME = 0; NAME < 9; NAME++", "\t\tif (NAME === 1) {\n\t\t\tNAME += 2;\n\t\t}\n", false},
		{"let NAME = 0; NAME < 5; NAME++", "\t\tconst read = (): number => NAME;\n\t\tsum += read();\n", false},
		{"let NAME = 9; NAME >= 0; NAME -= 2", "\t\tconst read = (): number => NAME;\n\t\tsum += read();\n", false},
	}
	for _, special := range specials {
		cases++
		name := fmt.Sprintf("c%d", cases)
		loop(name, strings.ReplaceAll(special.header, "NAME", name), strings.ReplaceAll(special.extra, "NAME", name))
		result[len(result)-1].expected[name] = special.integer
	}

	// Loops in a kept counter's body that read it, for their start or their bound, in the range it
	// holds there; -I is -0 when I is 0, and a product's sign is refused the same way.
	nested := []struct {
		outer, inner         string
		outerKept, innerKept bool
	}{
		{"let I = 0; I < values.length; I++", "let J = I + 1; J < values.length; J++", true, true},
		{"let I = 0; I < values.length; I++", "let J = I; J >= 0; J--", true, true},
		{"let I = 0; I < values.length; I++", "let J = -I; J < 3; J++", true, false},
		{"let I = 0; I < values.length; I++", "let J = I * 2; J < 9; J++", true, false},
		{"let I = 0; I < values.length; I++", "let J = 0; J < I; J++", true, true},
		{"let I = 0; I < values.length; I++", "let J = 0; J <= I * 3 - 1; J += 2", true, true},
		{"let I = 9007199254740989; I < 9007199254740992; I++", "let J = I; J <= I + 1; J++", true, false},
		{"let I = 9007199254740989; I < 9007199254740992; I++", "let J = I; J < I + 1; J++", true, true},
		{"let I = 9007199254740989; I <= 9007199254740991; I += 1", "let J = I + 1; J < 9007199254740992; J++", true, true},
		{"let I = 0; I < 3; I += 0.5", "let J = I + 1; J < 5; J++", false, false},
		{"let I = values.length; I > 0; I--", "let J = I - 1; J >= 0; J -= 2", true, true},
		{"let I = 5; I < 3; I++", "let J = I; J < 9; J++", true, false},
	}
	for _, each := range nested {
		cases++
		name := fmt.Sprintf("c%d", cases)
		inner := name + "j"
		header := strings.ReplaceAll(each.inner, "J", inner)
		header = strings.ReplaceAll(header, "I", name)
		extra := fmt.Sprintf("\t\tfor (%s) {\n\t\t\tsign += 1 / %s;\n\t\t\tsum += %s + (values[%s] ?? 0.25);\n"+
			"\t\t\tseen += 1;\n\t\t\tif (seen >= 6) {\n\t\t\t\tbreak;\n\t\t\t}\n\t\t}\n", header, inner, inner, inner)
		loop(name, strings.ReplaceAll(each.outer, "I", name), extra)
		result[len(result)-1].expected[name] = each.outerKept
		result[len(result)-1].expected[inner] = each.innerKept
	}

	return result
}

// A batch keeps compilation shared while every execution and assertion belongs to one case.
// sync.Once builds each batch exactly once; different batches build in parallel.
type loopCounterBatch struct {
	once                                   sync.Once
	source, code, path, sanitized, release string
	oracle                                 run
	lines                                  map[string][]byte
	ready                                  bool
}

func (batch *loopCounterBatch) build(t *testing.T) {
	t.Helper()
	batch.once.Do(func() {
		if err := os.WriteFile(batch.path, []byte(batch.source), 0o644); err != nil {
			t.Fatal(err)
		}
		program, err := lowered(t, batch.path)
		if err != nil {
			t.Fatalf("Lower: %v", err)
		}
		batch.code = native.C(program)
		if err := native.Build(batch.code, batch.sanitized, native.Options{Sanitize: true}); err != nil {
			t.Fatal(err)
		}
		if err := native.Build(batch.code, batch.release, native.Options{}); err != nil {
			t.Fatal(err)
		}
		// The old oracle also ran the sweep together. These functions only read shared globals.
		// Observe each batch once on Node, then give each leaf its own expected output line.
		batch.oracle = onNode(t, batch.path)
		batch.lines = map[string][]byte{}
		for _, line := range strings.Split(string(batch.oracle.stdout), "\n") {
			name, _, ok := strings.Cut(line, " ")
			if ok {
				batch.lines[name] = []byte(line + "\n")
			}
		}
		batch.ready = true
	})
	if !batch.ready {
		t.Fatal("shared loop batch build failed")
	}
}

func TestLoopCountersAgreeWithNode(t *testing.T) {
	t.Parallel()
	identity(t)
	cases := loopCounterCases()
	if len(cases) != loopCounterCaseCount {
		t.Fatalf("case count %d, want %d", len(cases), loopCounterCaseCount)
	}
	const batchSize = 64
	prelude := cases[0].source[:strings.Index(cases[0].source, "function c1():")]
	batches := make([]*loopCounterBatch, (len(cases)+batchSize-1)/batchSize)
	for index := range batches {
		directory := t.TempDir()
		batch := &loopCounterBatch{path: filepath.Join(directory, "batch.ts"), sanitized: filepath.Join(directory, "sanitized"), release: filepath.Join(directory, "release")}
		var source strings.Builder
		source.WriteString("import { programArguments } from 'adamic';\n")
		source.WriteString(prelude)
		source.WriteString("const selected = programArguments()[0];\n")
		end := min((index+1)*batchSize, len(cases))
		for _, each := range cases[index*batchSize : end] {
			declaration := strings.TrimSuffix(strings.TrimPrefix(each.source, prelude), each.name+"();\n")
			source.WriteString(declaration)
			fmt.Fprintf(&source, "if (selected === undefined || selected === '%s') { %s(); }\n", each.name, each.name)
		}
		batch.source = source.String()
		batches[index] = batch
	}
	for index, each := range cases {
		batch := batches[index/batchSize]
		t.Run(each.name, func(t *testing.T) {
			t.Parallel()
			batch.build(t)
			for name, integer := range each.expected {
				kept := regexp.MustCompile(`int64_t adamic_local_\d+_` + name + ` =`).MatchString(batch.code)
				if kept != integer {
					t.Errorf("%s: kept in an integer %t, want %t", name, kept, integer)
				}
			}
			stdout, ok := batch.lines[each.name]
			if !ok {
				t.Fatalf("Node omitted %s: exit %d, stderr %q", each.name, batch.oracle.exitCode, batch.oracle.stderr)
			}
			oracle := run{stdout: stdout, stderr: batch.oracle.stderr, exitCode: batch.oracle.exitCode}
			var environment []string
			if runtime.GOOS == "linux" {
				environment = []string{"ASAN_OPTIONS=detect_leaks=0"}
			}
			sanitized := executeWith(t, environment, batch.sanitized, each.name)
			release := execute(t, batch.release, each.name)
			if difference := disagreement(oracle, sanitized); difference != "" {
				t.Errorf("sanitized: %s\n%s", difference, lineDifferences(oracle.stdout, sanitized.stdout))
			}
			if difference := disagreement(oracle, release); difference != "" {
				t.Errorf("release: %s\n%s", difference, lineDifferences(oracle.stdout, release.stdout))
			}
		})
	}
}

// lineDifferences is the first few lines where two outputs differ.
func lineDifferences(want []byte, got []byte) string {
	wantLines, gotLines := strings.Split(string(want), "\n"), strings.Split(string(got), "\n")
	var report strings.Builder
	shown := 0
	for index := 0; index < len(wantLines) && index < len(gotLines) && shown < 6; index++ {
		if wantLines[index] != gotLines[index] {
			fmt.Fprintf(&report, "Node   %s\nnative %s\n", wantLines[index], gotLines[index])
			shown++
		}
	}
	return report.String()
}
