package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// The loop counter sweep: for loops at every edge of what lower/counters.go keeps in an integer,
// against Node. Each loop runs at most six passes (huge bounds break out), and prints how many it ran,
// the sum of its counter and of an array read by it, the sign of 1 / counter on its first pass (an
// integer has no -0) and its last value, so a counter that ran past 2^53, where JavaScript's + 1 stops
// changing it, or that lost a fraction or a sign, says so. Every loop is also checked to have been
// kept in an integer exactly when it should: the ones that fall back to a double must, and the ones
// that qualify mustn't miss out, so the sweep proves what it's sweeping.
func TestLoopCountersAgreeWithNode(t *testing.T) {
	t.Parallel()
	two53 := 9007199254740992.0
	starts := []struct {
		source string
		whole  bool
	}{
		{"0", true}, {"-0", false}, {"2", true}, {"-2", true}, {"0.5", false},
		{"9007199254740989", true}, {"9007199254740992", true}, {"-9007199254740992", true},
		{"9007199254740994", false}, {"2 ** 3", false},
	}
	bounds := []struct {
		source string
		// limit is the most the bound can be and still be a whole bound: 0 for one that's never one.
		value float64
		whole bool
	}{
		{"values.length", 0, true}, {"text.length", 0, true}, {"sizes.size", 0, true}, {"wholeConstant", 0, true},
		{"5", 5, true}, {"0", 0, true}, {"-3", -3, true}, {"9007199254740991", two53 - 1, true},
		{"9007199254740992", two53, true}, {"9007199254740994", 0, false}, {"2.5", 0, false}, {"-0.5", 0, false},
		{"NaN", 0, false}, {"Infinity", 0, false}, {"-Infinity", 0, false},
		{"fractionConstant", 0, false}, {"changing", 0, false}, {"2 ** 2", 0, false},
	}
	var source strings.Builder
	source.WriteString(`const values = [10, 20, 30, 40, 50];
const text = 'héllo';
const sizes = new Map<string, number>([['a', 1], ['b', 2], ['c', 3]]);
const wholeConstant = values.length;
const fractionConstant = 4.5;
let changing = 4;
changing = 6;
`)
	expected := map[string]bool{}
	loop := func(name string, header string, extra string) {
		fmt.Fprintf(&source, `function %s(): string {
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
	return %s;
}
console.log(%s());
`, name, header, name, extra, name, name, name, fmt.Sprintf("`%s ${seen} ${sum} ${sign} ${last}`", name), name)
	}
	cases := 0
	for _, start := range starts {
		for _, bound := range bounds {
			for _, comparison := range []string{"<", "<="} {
				cases++
				name := fmt.Sprintf("c%d", cases)
				loop(name, fmt.Sprintf("let %s = %s; %s %s %s; %s++", name, start.source, name, comparison, bound.source, name), "")
				whole := bound.whole
				if whole && bound.value != 0 || bound.source == "0" {
					// A constant: no larger than 2^53 for <, and smaller for <=.
					limit := two53
					if comparison == "<=" {
						limit = two53 - 1
					}
					whole = bound.value <= limit && bound.value >= -limit
				}
				expected[name] = start.whole && whole
			}
		}
	}
	// The other updates that count, and the loops that must keep a double for what they do.
	specials := []struct {
		header, extra string
		integer       bool
	}{
		{"let NAME = 0; NAME < 5; ++NAME", "", true},
		{"let NAME = 0; NAME < 5; NAME += 1", "", true},
		{"let NAME = 0; NAME < 5; NAME = NAME + 1", "", true},
		{"let NAME = 0; NAME < 9; NAME += 2", "", false},
		{"let NAME = 0; NAME < 9; NAME++", "\t\tif (NAME === 1) {\n\t\t\tNAME += 2;\n\t\t}\n", false},
		{"let NAME = 0; NAME < 5; NAME++", "\t\tconst read = (): number => NAME;\n\t\tsum += read();\n", false},
	}
	for _, special := range specials {
		cases++
		name := fmt.Sprintf("c%d", cases)
		loop(name, strings.ReplaceAll(special.header, "NAME", name), strings.ReplaceAll(special.extra, "NAME", name))
		expected[name] = special.integer
	}

	path := filepath.Join(t.TempDir(), "counters.ts")
	if err := os.WriteFile(path, []byte(source.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatalf("Lower: %v", err)
	}
	code := native.C(program)
	integers := 0
	for name, integer := range expected {
		kept := regexp.MustCompile(`int64_t adamic_local_\d+_` + name + ` =`).MatchString(code)
		if kept != integer {
			t.Errorf("%s: kept in an integer %t, want %t", name, kept, integer)
		}
		if integer {
			integers++
		}
	}
	oracle := onNode(t, path)
	sanitized, _ := natively(t, program)
	release := released(t, program)
	if difference := disagreement(oracle, sanitized); difference != "" {
		t.Errorf("sanitized: %s\n%s", difference, lineDifferences(oracle.stdout, sanitized.stdout))
	}
	if difference := disagreement(oracle, release); difference != "" {
		t.Errorf("release: %s\n%s", difference, lineDifferences(oracle.stdout, release.stdout))
	}
	t.Logf("%d loops, %d kept in an integer", cases, integers)
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
