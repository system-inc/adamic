package css

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"
)

type cssPrinterInputSet struct {
	cases, printerOracle string
	product              cssExecutable
	leakBinary           string
	mutated              []cssExecutable
	shards               []cssModeShard
	witnessShards        map[int]bool
	repo                 string
}

var cssPrinterInputs cssPrinterInputSet

// Build products are inputs to the timed units. TestMain materializes them once
// before M.Run starts leaf clocks, retaining them until all units finish. No
// package-local cache is used. Shared caching of sanitized native.Build products
// still requires reproducible sanitizer source paths in the native builder.
// Not parallel: product input lifetimes are shared by all printer units.
func TestMain(m *testing.M) { os.Exit(cssPrinterMain(m)) }
func cssPrinterMain(m *testing.M) (exit int) {
	if code, handled := cssPrinterNativeBuildMode(); handled {
		return code
	}
	needed := true
	pattern := ""
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "-test.list=") && strings.TrimPrefix(arg, "-test.list=") != "" {
			needed = false
		}
		if strings.HasPrefix(arg, "-test.run=") {
			pattern = strings.TrimPrefix(arg, "-test.run=")
		}
	}
	if needed {
		match, err := regexp.Compile(pattern)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		needed = false
		for shard := 0; shard < testCSSPrinterAgreesWithGoShards; shard++ {
			if match.MatchString(fmt.Sprintf("TestCSSPrinterAgreesWithGo_%03d", shard)) {
				needed = true
				break
			}
		}
	}
	if !needed {
		return m.Run()
	}
	directory, err := os.MkdirTemp("", "adamic-css-printer-inputs-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	builder := &cssPrinterBuilder{directory: directory}
	defer func() {
		for i := len(builder.cleanups) - 1; i >= 0; i-- {
			builder.cleanups[i]()
		}
		_ = os.RemoveAll(directory)
		if problem := recover(); problem != nil {
			fmt.Fprintln(os.Stderr, problem)
			exit = 1
		}
	}()
	started, cpu := time.Now(), cssCPU()
	cssPrinterInputs = cssPreparePrinterInputs(builder)
	fmt.Printf("printer input preparation including products: %s\n", time.Since(started))
	code := m.Run()
	fmt.Printf("printer total CPU including products: %.3fs\n", cssCPU()-cpu)
	return code
}

type cssPrinterBuilder struct {
	directory string
	cleanups  []func()
}

func (*cssPrinterBuilder) Helper() {}
func (b *cssPrinterBuilder) TempDir() string {
	dir, err := os.MkdirTemp(b.directory, "input-")
	if err != nil {
		b.Fatal(err)
	}
	return dir
}
func (*cssPrinterBuilder) Fatal(values ...any)                 { panic(fmt.Sprint(values...)) }
func (*cssPrinterBuilder) Fatalf(format string, values ...any) { panic(fmt.Sprintf(format, values...)) }
func (*cssPrinterBuilder) Logf(format string, values ...any)   { fmt.Printf(format+"\n", values...) }
func (b *cssPrinterBuilder) Cleanup(cleanup func())            { b.cleanups = append(b.cleanups, cleanup) }
func cssPreparePrinterInputs(t cssTesting) cssPrinterInputSet {
	rawOracle := cssOracleProduct(t, false)
	printerOracle := cssOracleProduct(t, true)
	cases := cssPrinterCases(t, rawOracle)
	directory := cssPrinterPortDirectory(t, nil)
	product := cssPrinterProduct(t, filepath.Join(directory, "print_main.ts"), "sanitized printer and lowered backends", true)
	leakBinary := product.binary
	if runtime.GOOS == "darwin" {
		leakBinary = cssPrinterProduct(t, product.source, "leak printer", false).binary
	}
	mutated := make([]cssExecutable, len(printerMutants))
	for index, mutation := range printerMutants {
		dir := cssPrinterPortDirectory(t, &mutation)
		mutated[index] = cssPrinterProduct(t, filepath.Join(dir, "print_main.ts"), "mutant "+mutation.name, true)
	}
	shards := cssPrinterModeShards(t, cases, testCSSPrinterAgreesWithGoShards)
	caseCount := 0
	for _, unit := range shards {
		if unit.mutant == -1 && unit.mode == "default" {
			caseCount += len(unit.cases)
		}
	}
	witnessShards := map[int]bool{}
	for shard, unit := range shards {
		if unit.mutant < 0 {
			continue
		}
		witness := []int{0, 2, 12}[unit.mutant]
		for _, c := range unit.cases {
			if c.id%caseCount == witness {
				witnessShards[shard] = true
			}
		}
	}
	if len(witnessShards) != 6 {
		t.Fatalf("found %d mutant witness units, want 6", len(witnessShards))
	}
	repo, _ := filepath.Abs(repository)
	return cssPrinterInputSet{cases, printerOracle, product, leakBinary, mutated, shards, witnessShards, repo}
}

// Printer-owned copy preserves source input lifetime independently of a leaf T.
func cssPrinterPortDirectory(t cssTesting, applied *mutant) string {
	directory := t.TempDir()
	for _, slice := range []string{"css", "selector", "values", "mediaquery", "cssstrings", "cssnumbers"} {
		entries, err := os.ReadDir(filepath.Join("..", slice))
		if err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(directory, slice)
		if err := os.MkdirAll(target, 0755); err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".ts") {
				continue
			}
			data, err := os.ReadFile(filepath.Join("..", slice, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			source := string(data)
			if applied != nil && slice == "css" && entry.Name() == applied.file {
				if strings.Count(source, applied.from) != 1 {
					t.Fatalf("mutant %s must match exactly once", applied.name)
				}
				source = strings.Replace(source, applied.from, applied.to, 1)
			}
			if err := os.WriteFile(filepath.Join(target, entry.Name()), []byte(source), 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
	return filepath.Join(directory, "css")
}
func TestCSSPrinterAgreesWithGoUnion(t *testing.T) {
	t.Parallel()
	table, err := parser.ParseFile(token.NewFileSet(), "printer_units_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	for shard := 0; shard < testCSSPrinterAgreesWithGoShards; shard++ {
		name := fmt.Sprintf("TestCSSPrinterAgreesWithGo_%03d", shard)
		if table.Scope.Lookup(name) == nil {
			t.Fatalf("missing top-level unit %s", name)
		}
	}
	enumerated := 0
	for name := range table.Scope.Objects {
		if strings.HasPrefix(name, "TestCSSPrinterAgreesWithGo_") {
			enumerated++
		}
	}
	if enumerated != testCSSPrinterAgreesWithGoShards {
		t.Fatalf("enumerated %d top-level units, want %d", enumerated, testCSSPrinterAgreesWithGoShards)
	}
	cases := cssPrinterInputs.cases
	if cases == "" {
		cases = cssPrinterCases(t, cssOracleProduct(t, false))
	}
	_ = cssPrinterModeShards(t, cases, testCSSPrinterAgreesWithGoShards)
}
