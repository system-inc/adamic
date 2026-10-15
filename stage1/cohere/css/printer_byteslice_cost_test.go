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

// Enough room for the sanitized allocator, but below the prefix-rebuild cost.
// Both times are measured in this unit on the identical live, pinned fixture.
const cssPrinterPreflightCostRatio = 48.0

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

func cssPrinterPreflightCases(t *testing.T) string {
	t.Helper()
	source, err := os.ReadFile(filepath.Join(repository, "cohere/internal/lint/rules/tailwind/vendored/testdata/tailwindcss/index.css"))
	if err != nil {
		t.Fatal(err)
	}
	escaped := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t").Replace(string(source))
	if len(escaped) != 30736 || !strings.HasPrefix(escaped, "@layer theme, base, components, utilities;") || !strings.Contains(escaped, "color-mix(…)") {
		t.Fatal("the cost witness must be the 30,736-byte Tailwind preflight with the ellipsis")
	}
	path := filepath.Join(t.TempDir(), "preflight.txt")
	// Preserve both occurrences in the printer corpus: CSS and SCSS.
	if err := os.WriteFile(path, []byte(">C"+escaped+"\n>S"+escaped+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func cssPrinterPreflightCost(t *testing.T, binary, oracle string) error {
	t.Helper()
	path := cssPrinterPreflightCases(t)
	goBest, portBest := time.Duration(math.MaxInt64), time.Duration(math.MaxInt64)
	// Minimum of three alternating samples reduces scheduler noise without an
	// absolute time floor that could let a fast-machine cost regression pass.
	for sample := 0; sample < 3; sample++ {
		started := time.Now()
		expected := cssPrinterAnswers(t, oracle, path, "default")
		goElapsed := time.Since(started)
		started = time.Now()
		if err := cssPrinterSanitizedAgreement(t, binary, expected, path, "output", "once", "default"); err != nil {
			return err
		}
		portElapsed := time.Since(started)
		if goElapsed < goBest {
			goBest = goElapsed
		}
		if portElapsed < portBest {
			portBest = portElapsed
		}
		t.Logf("preflight cost sample %d: Go=%s sanitized=%s", sample+1, goElapsed, portElapsed)
	}
	ratio := float64(portBest) / float64(goBest)
	t.Logf("preflight cost: Go=%s sanitized=%s ratio=%.2fx bound=%.0fx", goBest, portBest, ratio, cssPrinterPreflightCostRatio)
	if ratio > cssPrinterPreflightCostRatio {
		return fmt.Errorf("CSS preflight port cost %.2fx exceeds Go-relative bound %.0fx (Go %s, sanitized %s)", ratio, cssPrinterPreflightCostRatio, goBest, portBest)
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
		t.Fatal("prefix-rebuild cost mutant survived the Go-relative guard")
	}
	if !strings.Contains(err.Error(), "exceeds Go-relative bound") {
		t.Fatal(err)
	}
	t.Logf("cost mutant caught by the same guard: %v", err)
}
