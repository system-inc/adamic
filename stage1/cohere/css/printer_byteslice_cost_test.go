package css

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/stage1/cohere/css/oracles"
)

// Doubling these matched non-ASCII prefixes measures the same sanitized binary,
// interleaved under the same load. On four CPUs, alone and beside three CPU
// burners, clean scaling measured 2.10–2.13x and prefix rebuilding 3.05–3.12x.
// Linear parsing and startup dilute the mutant's quadratic 4x printer cost;
// 2.6x separates the measured ranges with headroom on both sides.
const cssPrinterPreflightScalingBound = 2.6
const cssPrinterPreflightSmallBytes = 6 * 1024

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

var cssPrinterCostMutantOnce sync.Once
var cssPrinterCostMutantBinary string

func cssPrinterCostMutant(t *testing.T) string {
	t.Helper()
	cssPrinterCostMutantOnce.Do(func() {
		options := native.Options{Sanitize: true}
		inputs := buildcache.Inputs{
			Name:      "css-printer-prefix-cost-mutant",
			Files:     cssLoweringFiles,
			Flags:     append(native.Flags(options), cssPrinterPrefixLookup, cssPrinterPrefixRebuild, "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT")),
			Toolchain: []string{runtime.Version(), buildcache.Tool("clang", "--version"), runtime.GOOS, runtime.GOARCH},
		}
		directory := buildcache.Product(t, inputs, func(dir string) error {
			copied := portDirectory(t, &mutant{"byte offsets rebuild prefixes", "byte_view.ts", cssPrinterPrefixLookup, cssPrinterPrefixRebuild})
			if err := moveTree(filepath.Dir(copied), filepath.Join(dir, "sources")); err != nil {
				return err
			}
			source := filepath.Join(dir, "sources/css/print_main.ts")
			checked, err := load.Load([]string{source})
			if err != nil {
				return err
			}
			program, err := lower.Lower(t.Context(), checked)
			if err != nil {
				return err
			}
			return native.Build(native.C(program), filepath.Join(dir, "native"), options)
		})
		cssPrinterCostMutantBinary = filepath.Join(directory, "native")
	})
	if cssPrinterCostMutantBinary == "" {
		t.Fatal("CSS cost mutant preparation failed")
	}
	return cssPrinterCostMutantBinary
}

func TestProduct_CSSPrinterPrefixCostMutant(t *testing.T) {
	t.Parallel()
	cssPrinterCostMutant(t)
}

func cssPrinterPreflightCases(t *testing.T) [2]string {
	t.Helper()
	source, err := os.ReadFile(filepath.Join(repository, "cohere/internal/lint/rules/tailwind/vendored/testdata/tailwindcss/index.css"))
	if err != nil {
		t.Fatal(err)
	}
	escape := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t")
	escaped := escape.Replace(string(source))
	if len(escaped) != 30736 || !strings.HasPrefix(escaped, "@layer theme, base, components, utilities;") || !strings.Contains(escaped, "color-mix(…)") {
		t.Fatal("the cost witness must be the 30,736-byte Tailwind preflight with the ellipsis")
	}
	// Keep the live ellipsis comment in both prefixes, so neither ByteView can
	// select its ASCII path. Both cuts stay in @theme's declaration list.
	ellipsis := strings.Index(string(source), "color-mix(…)")
	commentStart := strings.LastIndex(string(source[:ellipsis]), "/*")
	commentTail := strings.Index(string(source[ellipsis:]), "*/")
	if commentStart < 0 || commentTail < 0 {
		t.Fatal("missing ellipsis comment")
	}
	commentEnd := ellipsis + commentTail + 2
	comment := string(source[commentStart:commentEnd]) + "\n"
	const closing = "\n  }\n}\n"
	var paths [2]string
	for index, size := range []int{cssPrinterPreflightSmallBytes, 2 * cssPrinterPreflightSmallBytes} {
		cut := strings.LastIndex(string(source[:size-len(comment)-len(closing)]), ";") + 1
		prefix := string(source[:cut])
		if cut == 0 || strings.Count(prefix, "{") != 2 || strings.Contains(prefix, "}") || strings.Contains(prefix, "/*") {
			t.Fatal("the preflight prefix must end at a declaration inside @layer theme / @theme")
		}
		text := prefix + closing + comment
		// Whitespace makes the source byte lengths exactly 1:2; at most one partial
		// declaration is padded, without changing the printer's document shape.
		text += strings.Repeat(" ", size-len(text))
		paths[index] = filepath.Join(t.TempDir(), "preflight.txt")
		// Preserve both syntax modes from the original printer corpus.
		if err := os.WriteFile(paths[index], []byte(">C"+escape.Replace(text)+"\n>S"+escape.Replace(text)+"\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return paths
}

func cssPrinterPreflightCost(t *testing.T, binary, oracle string) error {
	t.Helper()
	paths := cssPrinterPreflightCases(t)
	// Go remains the byte-for-byte oracle for each input; its speed is not the
	// denominator of a sanitized cost comparison.
	var expected [2]string
	for index, path := range paths {
		expected[index] = cssPrinterAnswers(t, oracle, path, "default")
	}
	best := [2]time.Duration{time.Duration(math.MaxInt64), time.Duration(math.MaxInt64)}
	// Interleave sizes and keep the minimum of three samples of each, reducing
	// scheduler noise without a machine-dependent absolute time floor.
	for sample := 0; sample < 3; sample++ {
		for index, path := range paths {
			started := time.Now()
			if err := cssPrinterSanitizedAgreement(t, binary, expected[index], path, "output", "once", "default"); err != nil {
				return err
			}
			elapsed := time.Since(started)
			if elapsed < best[index] {
				best[index] = elapsed
			}
			t.Logf("preflight scaling sample %d: bytes=%d sanitized=%s", sample+1, (index+1)*cssPrinterPreflightSmallBytes, elapsed)
		}
	}
	ratio := float64(best[1]) / float64(best[0])
	t.Logf("preflight scaling: small=%s large=%s ratio=%.3fx bound=%.2fx", best[0], best[1], ratio, cssPrinterPreflightScalingBound)
	if ratio > cssPrinterPreflightScalingBound {
		return fmt.Errorf("CSS preflight scaling %.3fx exceeds scaling bound %.2fx (small %s, large %s)", ratio, cssPrinterPreflightScalingBound, best[0], best[1])
	}
	return nil
}

// Ordinary top-level tests are enumerated by adamic-gate's go test -list census;
// no opt-in environment or optional fixture root is required for this guard.
func TestCSSPrinterPreflightCost(t *testing.T) {
	t.Parallel()
	oracle := oracles.ComparisonOracle(t, true)
	binary := cssPrinterProduct(t, -1).binary
	// A scratch probe can run the exact guard with the restored slow path.
	if os.Getenv("ADAMIC_CSS_PREFIX_COST_MUTANT") == "1" {
		binary = cssPrinterCostMutant(t)
	}
	if err := cssPrinterPreflightCost(t, binary, oracle); err != nil {
		t.Fatal(err)
	}
}

func TestCSSPrinterPreflightCostMutant(t *testing.T) {
	t.Parallel()
	oracle := oracles.ComparisonOracle(t, true)
	binary := cssPrinterCostMutant(t)
	err := cssPrinterPreflightCost(t, binary, oracle)
	if err == nil {
		t.Fatal("prefix-rebuild cost mutant survived the scaling guard")
	}
	if !strings.Contains(err.Error(), "exceeds scaling bound") {
		t.Fatal(err)
	}
	t.Logf("cost mutant caught by the same guard: %v", err)
}
