package css

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/stage1/cohere/css/oracles"
)

// Widening from 6 to 24 KiB compounds the linear/quadratic scaling gap. The guard counts the printer's allocations
// instead of timing it: a counted build (native.Options.Count) reports them, and one input gives one count on any box
// under any load. Timed, the same clean change read 4.04x quiet and 6.18x on a loaded Eve against a 6x bound (Loom,
// Oct 21). Counted, the printer allocated 512,884 then 2,025,835 values (3.95x) and the prefix-rebuild mutant 1,214,890
// then 14,536,375 (11.97x), every count the same on Pupa and Lunatic, and beside 24 CPU burners on Lunatic's 8 CPUs at
// a load of 31. Allocations part them widest, since rebuilding a prefix allocates for each character it walks; its
// releases scaled 7.55x and its retains not at all. Their geometric middle is 6.88x, and a 7x bound leaves each side
// about 1.7x.
const cssPrinterPreflightScalingBound = 7.0
const cssPrinterPreflightSmallBytes = 6 * 1024
const cssPrinterPreflightLargeBytes = 24 * 1024

const cssPrinterPrefixLookup = "return this.ascii ? byte : (this.characters[byte] ?? 0) + (this.partial[byte] ?? 0);"
const cssPrinterPrefixRebuild = `if(this.ascii) return byte;
        let cursor = 0;
        let result = '';
        for(const character of this.text) {
            if(cursor >= byte) break;
            const next = cursor + utf8Length(character);
            if(next <= byte) result += character;
            else result += replacement(byte - cursor);
            cursor = next;
        }
        return result.length;`

// cssPrinterCostMutants are the counted printers the guard measures: the printer as it is, and the printer with the
// byte-offset slow path restored.
var cssPrinterCostMutants = []*mutant{
	nil,
	{"byte offsets rebuild prefixes", "byte_view.ts", cssPrinterPrefixLookup, cssPrinterPrefixRebuild},
}

var cssPrinterCountedOnce [2]sync.Once
var cssPrinterCountedNatives [2]string

// cssPrinterCounted is one of cssPrinterCostMutants' counted native, built ahead as a product. Its key is what lowering
// reads, as cssPrinterProduct's is, with the build's flags and, for the mutant, its edit.
func cssPrinterCounted(test *testing.T, index int) string {
	test.Helper()
	cssPrinterCountedOnce[index].Do(func() {
		options := native.Options{Count: true}
		name := "css-printer-counted"
		flags := append(native.Flags(options), "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"))
		if applied := cssPrinterCostMutants[index]; applied != nil {
			name = "css-printer-prefix-cost-mutant-counted"
			flags = append(flags, applied.from, applied.to)
		}
		inputs := buildcache.Inputs{
			Name:      name,
			Files:     cssLoweringFiles,
			Flags:     flags,
			Toolchain: []string{runtime.Version(), buildcache.Tool("clang", "--version"), runtime.GOOS, runtime.GOARCH},
		}
		directory := buildcache.Product(test, inputs, func(dir string) error {
			copied := portDirectory(test, cssPrinterCostMutants[index])
			if moveTreeError := moveTree(filepath.Dir(copied), filepath.Join(dir, "sources")); moveTreeError != nil {
				return moveTreeError
			}
			source := filepath.Join(dir, "sources/css/print_main.ts")
			checked, loadError := load.Load([]string{source})
			if loadError != nil {
				return loadError
			}
			program, lowerError := lower.Lower(test.Context(), checked)
			if lowerError != nil {
				return lowerError
			}
			return native.Build(native.C(program), filepath.Join(dir, "native"), options)
		})
		cssPrinterCountedNatives[index] = filepath.Join(directory, "native")
	})
	if cssPrinterCountedNatives[index] == "" {
		test.Fatalf(
			"counted css printer %d unavailable: its build failed earlier in this process (see that test's log)",
			index,
		)
	}
	return cssPrinterCountedNatives[index]
}

func TestProduct_CssPrinterCounted(test *testing.T) {
	test.Parallel()
	cssPrinterCounted(test, 0)
}

func TestProduct_CssPrinterPrefixCostMutant(test *testing.T) {
	test.Parallel()
	cssPrinterCounted(test, 1)
}

func cssPrinterPreflightCases(test *testing.T) [2]string {
	test.Helper()
	source, readError := os.ReadFile(
		filepath.Join(repository, "cohere/internal/lint/rules/tailwind/vendored/testdata/tailwindcss/index.css"),
	)
	if readError != nil {
		test.Fatal(readError)
	}
	escape := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t")
	escaped := escape.Replace(string(source))
	if len(escaped) != 30736 ||
		!strings.HasPrefix(escaped, "@layer theme, base, components, utilities;") ||
		!strings.Contains(escaped, "color-mix(…)") {
		test.Fatal("the cost witness must be the 30,736-byte Tailwind preflight with the ellipsis")
	}
	// Keep the live ellipsis comment in both prefixes, so neither ByteView can
	// select its ASCII path. Cuts end at complete declarations or rules.
	ellipsis := strings.Index(string(source), "color-mix(…)")
	commentStart := strings.LastIndex(string(source[:ellipsis]), "/*")
	commentTail := strings.Index(string(source[ellipsis:]), "*/")
	if commentStart < 0 || commentTail < 0 {
		test.Fatal("missing ellipsis comment")
	}
	commentEnd := ellipsis + commentTail + 2
	comment := string(source[commentStart:commentEnd]) + "\n"

	var paths [2]string
	for index, size := range []int{cssPrinterPreflightSmallBytes, cssPrinterPreflightLargeBytes} {
		closing := "\n  }\n}\n"
		boundary := ";"
		depth := 2
		if size > 18*1024 {
			closing, boundary, depth = "\n}\n", "}", 1
		}
		cut := strings.LastIndex(string(source[:size-len(comment)-len(closing)]), boundary) + 1
		prefix := string(source[:cut])
		if cut == 0 ||
			strings.Count(prefix, "{")-strings.Count(prefix, "}") != depth ||
			strings.LastIndex(prefix, "/*") > strings.LastIndex(prefix, "*/") {
			test.Fatal("the preflight prefix must end outside a comment at the expected block depth")
		}
		text := prefix + closing + comment
		// Pad to the exact requested byte length after the last complete
		// declaration or rule, preserving the printer's document shape.
		text += strings.Repeat(" ", size-len(text))
		paths[index] = filepath.Join(test.TempDir(), "preflight.txt")
		// Preserve both syntax modes from the original printer corpus.
		if writeError := os.WriteFile(
			paths[index],
			[]byte(">C"+escape.Replace(text)+"\n>S"+escape.Replace(text)+"\n"),
			0644,
		); writeError != nil {
			test.Fatal(writeError)
		}
	}
	return paths
}

// cssPrinterCountedAllocations runs a counted printer on one input, holds its output to Go's answer, and returns how
// many values it allocated. A counted build writes its counts to stderr as one line,
// "adamic: counts: allocations 3 frees 3 retains 7 releases 10 peak 2 regions 0", and anything else there is an error.
func cssPrinterCountedAllocations(test *testing.T, binary, expected, path string) (uint64, error) {
	test.Helper()
	result := execute(test, nil, binary, path, "output", "once", "default")
	report, found := strings.CutPrefix(string(result.stderr), "adamic: counts: ")
	fields := strings.Fields(report)
	if result.exitCode != 0 ||
		!found ||
		strings.Count(report, "\n") != 1 ||
		!strings.HasSuffix(report, "\n") ||
		len(fields) < 2 ||
		fields[0] != "allocations" {
		return 0, fmt.Errorf("exit %d stderr %s", result.exitCode, result.stderr)
	}
	if difference := firstDifference(string(result.stdout), expected); difference != "" {
		return 0, fmt.Errorf("output differs from Go: %s", difference)
	}
	test.Logf("counts: %s", strings.TrimSpace(report))
	return strconv.ParseUint(fields[1], 10, 64)
}

func cssPrinterPreflightCost(test *testing.T, binary, oracle string) error {
	test.Helper()
	paths := cssPrinterPreflightCases(test)
	var allocations [2]uint64
	for index, path := range paths {
		// Go remains the byte-for-byte oracle for each input.
		expected := cssPrinterAnswers(test, oracle, path, "default")
		// The guard rests on a count being exact, so each input runs twice and a count that moved is an error.
		for repeat := 0; repeat < 2; repeat++ {
			counted, cssPrinterCountedAllocationsError := cssPrinterCountedAllocations(test, binary, expected, path)
			if cssPrinterCountedAllocationsError != nil {
				return cssPrinterCountedAllocationsError
			}
			if repeat == 1 && counted != allocations[index] {
				return fmt.Errorf(
					"the counted printer's allocations on one input moved between runs: %d then %d",
					allocations[index],
					counted,
				)
			}
			allocations[index] = counted
		}
	}
	ratio := float64(allocations[1]) / float64(allocations[0])
	test.Logf(
		"preflight scaling: allocations small=%d large=%d ratio=%.3fx bound=%.2fx",
		allocations[0],
		allocations[1],
		ratio,
		cssPrinterPreflightScalingBound,
	)
	if ratio > cssPrinterPreflightScalingBound {
		return fmt.Errorf(
			"CSS preflight scaling %.3fx exceeds scaling bound %.2fx (allocations %d then %d)",
			ratio,
			cssPrinterPreflightScalingBound,
			allocations[0],
			allocations[1],
		)
	}
	return nil
}

// Ordinary top-level tests are enumerated by adamic-gate's go test -list census;
// no opt-in environment or optional fixture root is required for this guard.
func TestCssPrinterPreflightCost(test *testing.T) {
	test.Parallel()
	oracle := oracles.ComparisonOracle(test, true)
	if cssPrinterPreflightCostError := cssPrinterPreflightCost(
		test,
		cssPrinterCounted(test, 0),
		oracle,
	); cssPrinterPreflightCostError != nil {
		test.Fatal(cssPrinterPreflightCostError)
	}
}

func TestCssPrinterPreflightCostMutant(test *testing.T) {
	test.Parallel()
	oracle := oracles.ComparisonOracle(test, true)
	cssPrinterPreflightCostError := cssPrinterPreflightCost(test, cssPrinterCounted(test, 1), oracle)
	if cssPrinterPreflightCostError == nil {
		test.Fatal("prefix-rebuild cost mutant survived the scaling guard")
	}
	if !strings.Contains(cssPrinterPreflightCostError.Error(), "exceeds scaling bound") {
		test.Fatal(cssPrinterPreflightCostError)
	}
	test.Logf("cost mutant caught by the same guard: %v", cssPrinterPreflightCostError)
}
