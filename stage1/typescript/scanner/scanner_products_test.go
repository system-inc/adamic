package scanner

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/stage1/typescript/scanner/oracles"
)

const scannerShardCount = 16

var scannerMutants = []struct{ name, file, from, to string }{
	{"source scanned as Identifier", "tokens.ts", `['source', 'SourceKeyword']`, `['source', 'Identifier']`},
	{"punctuator != scanned as ==", "tokens.ts", `['!=', 'ExclamationEqualsToken']`, `['!=', 'EqualsEqualsToken']`},
	{
		"invalid decimal separator accepted",
		"scanner.ts",
		"this.error(previous ? 6189 : 6188, this.pos, 1);",
		"if (base !== 10) { this.error(previous ? 6189 : 6188, this.pos, 1); } else { this.flags &= ~16384; }",
	},
	{"regex rescan skipped", "main.ts", "scanner.rescanSlash();", "scanner.code();"},
}

// Identity is content-addressed by buildcache; sync.Once only prevents repeat fetches in a process.
type scannerPrepared struct {
	once      sync.Once
	directory string
}

var scannerProducts sync.Map

func scannerFetch(test *testing.T, inputs buildcache.Inputs, recipe func(string) error) string {
	test.Helper()
	identity, err := json.Marshal(inputs)
	if err != nil {
		test.Fatal(err)
	}
	value, _ := scannerProducts.LoadOrStore(string(identity), &scannerPrepared{})
	prepared := value.(*scannerPrepared)
	prepared.once.Do(func() { prepared.directory = buildcache.Product(test, inputs, recipe) })
	if prepared.directory == "" {
		test.Fatal("scanner product preparation failed")
	}
	return prepared.directory
}

func scannerOracleProduct(test *testing.T) string {
	test.Helper()
	return oracles.ScannerOracleProduct(test)
}

func scannerPortProduct(test *testing.T, source string, sanitize bool) string {
	test.Helper()
	options := native.Options{Sanitize: sanitize}
	flags := append(native.Flags(options), "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"))
	contents := make(map[string][]byte)
	for _, name := range portFiles {
		data, err := os.ReadFile(filepath.Join(source, name))
		if err != nil {
			test.Fatal(err)
		}
		contents[name] = data
		flags = append(flags, name+"="+string(data))
	}
	inputs := buildcache.Inputs{
		Name:      "scanner-native",
		Files:     []string{"internal", "cohere", "go.mod", "go.work"},
		Flags:     flags,
		Toolchain: []string{runtime.Version(), buildcache.Tool("clang", "--version")},
	}
	return scannerFetch(test, inputs, func(directory string) error {
		for _, name := range portFiles {
			if err := os.WriteFile(filepath.Join(directory, name), contents[name], 0644); err != nil {
				return err
			}
		}
		program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
		if err != nil {
			return err
		}
		lowered, err := lower.Lower(context.Background(), program)
		if err != nil {
			return err
		}
		return native.Build(native.C(lowered), filepath.Join(directory, "scanner"), options)
	})
}

func scannerNative(test *testing.T, mutant int, sanitize bool) string {
	test.Helper()
	source := "."
	if mutant >= 0 {
		m := scannerMutants[mutant]
		source = copyPort(test, m.file, m.from, m.to)
	}
	return scannerPortProduct(test, source, sanitize)
}

func scannerProductTime(test *testing.T, build func() string) {
	test.Helper()
	started := time.Now()
	build()
	test.Logf("product cold/fetch time %.3fs", time.Since(started).Seconds())
}

func TestProduct_ScannerFindsOracle(test *testing.T) {
	test.Parallel()
	scannerProductTime(test, func() string { return scannerOracleProduct(test) })
}
func TestProduct_ScannerNative(test *testing.T) {
	test.Parallel()
	scannerProductTime(test, func() string { return scannerNative(test, -1, true) })
}
func TestProduct_ScannerRelease(test *testing.T) {
	test.Parallel()
	scannerProductTime(test, func() string { return scannerNative(test, -1, false) })
}
func TestProduct_ScannerSourceIdentifier(test *testing.T) {
	test.Parallel()
	scannerProductTime(test, func() string { return scannerNative(test, 0, true) })
}
func TestProduct_ScannerPunctuator(test *testing.T) {
	test.Parallel()
	scannerProductTime(test, func() string { return scannerNative(test, 1, true) })
}
func TestProduct_ScannerDecimalSeparator(test *testing.T) {
	test.Parallel()
	scannerProductTime(test, func() string { return scannerNative(test, 2, true) })
}
func TestProduct_ScannerRegexRescan(test *testing.T) {
	test.Parallel()
	scannerProductTime(test, func() string { return scannerNative(test, 3, true) })
}

func scannerOwner(identity string) int {
	return int(sha256.Sum256([]byte(identity))[0]) % scannerShardCount
}

// Generated names are ordinal and repository/upstream names relative: no temporary directory enters ownership.
func scannerCaseIdentity(line string) string {
	fields := strings.SplitN(line, "\t", 2)
	path := filepath.ToSlash(fields[1])
	for _, marker := range []string{"/src/compiler/", "/stage1/"} {
		if at := strings.LastIndex(path, marker); at >= 0 {
			return fields[0] + "\t" + path[at+1:]
		}
	}
	return fields[0] + "\t" + filepath.Base(path)
}

func scannerManifestLines(test *testing.T, path string) []string {
	test.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		test.Fatal(err)
	}
	return strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
}

func scannerSelect(lines []string, shard int) []string {
	var selected []string
	for _, line := range lines {
		if scannerOwner(scannerCaseIdentity(line)) == shard {
			selected = append(selected, line)
		}
	}
	return selected
}

// scannerRun reads a run's output through childguard.Capture, to its end with no timer deciding when: exec's own
// copying under a three-second WaitDelay reported exec.ErrWaitDelay whenever the reads hadn't reached EOF three seconds
// after the child exited, so a Go oracle, native or Node run that had finished could fail its shard under load
// (#n3gk22h's shape). The shard's 45 s deadline ends a hung run.
func scannerRun(test *testing.T, ctx context.Context, name string, args ...string) []byte {
	test.Helper()
	command := exec.CommandContext(ctx, name, args...)
	command.Cancel = func() error { return syscall.Kill(-command.Process.Pid, syscall.SIGKILL) }
	output, stderr, err := childguard.Capture(ctx, command)
	if err != nil || len(stderr) != 0 {
		test.Fatalf("%s: %v\n%s", name, err, stderr)
	}
	return output
}

func scannerNode(test *testing.T, ctx context.Context, directory, manifest string) []byte {
	test.Helper()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		test.Fatal(err)
	}
	return scannerRun(
		test,
		ctx,
		"node",
		"--disable-warning=ExperimentalWarning",
		runner,
		filepath.Join(directory, "main.ts"),
		"--manifest",
		manifest,
	)
}

func scannerShard(test *testing.T, shard int) {
	test.Helper()
	setup := time.Now()
	oracle := goOracle(test)
	base := scannerNative(test, -1, true)
	products := make(map[int]string)
	for i, m := range scannerMutants {
		if scannerOwner("mutant:"+m.name) == shard {
			products[i] = scannerNative(test, i, true)
		}
	}
	asked := askedCorpus(test)
	lines := scannerSelect(scannerManifestLines(test, asked.all), shard)
	manifest := filepath.Join(test.TempDir(), "shard.txt")
	if err := os.WriteFile(manifest, []byte(strings.Join(lines, "\n")+"\n"), 0644); err != nil {
		test.Fatal(err)
	}
	setupTime := time.Since(setup)
	started := time.Now()
	// Setup carries no test-side deadline. Only this shard's comparisons are budgeted.
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	want := scannerRun(test, ctx, oracle, "--manifest", manifest)
	got := scannerRun(test, ctx, filepath.Join(base, "scanner"), "--manifest", manifest)
	if diff := difference(got, want); diff != "" {
		test.Errorf("native: %s", diff)
	}
	got = scannerNode(test, ctx, base, manifest)
	if diff := difference(got, want); diff != "" {
		test.Errorf("Node: %s", diff)
	}
	// The opt-in planted answer corruption is owned by precisely the shard containing this case.
	if os.Getenv("ADAMIC_SCANNER_PLANT_FAILURE") == "1" {
		whole := scannerManifestLines(test, asked.all)
		planted := scannerCaseIdentity(whole[0])
		for _, line := range lines {
			if scannerCaseIdentity(line) == planted {
				if difference(append([]byte("planted failure\n"), got...), want) != "" {
					test.Error("caught planted answer corruption")
				}
			}
		}
	}
	if len(products) > 0 {
		edgeWant := scannerRun(test, ctx, oracle, "--manifest", asked.edges)
		for i, product := range products {
			for _, side := range []struct {
				name   string
				output []byte
			}{
				{"Node", scannerNode(test, ctx, product, asked.edges)},
				{"native", scannerRun(test, ctx, filepath.Join(product, "scanner"), "--manifest", asked.edges)},
			} {
				if diff := difference(side.output, edgeWant); diff == "" {
					test.Errorf("%s %s: mutant survives comparison", scannerMutants[i].name, side.name)
				} else {
					test.Logf("%s %s caught: %s", scannerMutants[i].name, side.name, diff)
				}
			}
		}
	}
	test.Logf(
		"shard %03d: %d cases, %d mutants; setup %.3fs / own work %.3fs",
		shard,
		len(lines),
		len(products),
		setupTime.Seconds(),
		time.Since(started).Seconds(),
	)
}

func TestScannerShardCoverage(test *testing.T) {
	test.Parallel()
	asked := askedCorpus(test)
	whole := scannerManifestLines(test, asked.all)
	counts := make(map[string]int)
	mutantCounts := make([]int, len(scannerMutants))
	planted := scannerCaseIdentity(whole[0])
	caught := 0
	for shard := 0; shard < scannerShardCount; shard++ {
		selected := scannerSelect(whole, shard)
		for _, line := range selected {
			counts[scannerCaseIdentity(line)]++
			if scannerCaseIdentity(line) == planted && difference([]byte("planted\n"), []byte("original\n")) != "" {
				caught++
			}
		}
		for i, m := range scannerMutants {
			if scannerOwner("mutant:"+m.name) == shard {
				mutantCounts[i]++
			}
		}
		test.Logf("shard %03d owns %d cases", shard, len(selected))
	}
	if len(whole) != asked.files+asked.generated {
		test.Fatalf("whole count %d != %d files + %d generated", len(whole), asked.files, asked.generated)
	}
	if len(counts) != len(whole) {
		test.Fatalf("union %d != whole %d", len(counts), len(whole))
	}
	for _, line := range whole {
		if counts[scannerCaseIdentity(line)] != 1 {
			test.Errorf("case ownership %q: %d", line, counts[scannerCaseIdentity(line)])
		}
	}
	for i, count := range mutantCounts {
		if count != 1 {
			test.Errorf("mutant %s owned %d times", scannerMutants[i].name, count)
		}
	}
	if caught != 1 {
		test.Fatalf("planted failure caught by %d shards", caught)
	}
	test.Logf(
		"counted union: %d cases and %d mutants exactly once; planted corruption caught once",
		len(whole),
		len(scannerMutants),
	)
}

func TestScannerAgreesWithTypescriptGo_000(test *testing.T) { test.Parallel(); scannerShard(test, 0) }

func TestScannerAgreesWithTypescriptGo_001(test *testing.T) { test.Parallel(); scannerShard(test, 1) }

func TestScannerAgreesWithTypescriptGo_002(test *testing.T) { test.Parallel(); scannerShard(test, 2) }

func TestScannerAgreesWithTypescriptGo_003(test *testing.T) { test.Parallel(); scannerShard(test, 3) }

func TestScannerAgreesWithTypescriptGo_004(test *testing.T) { test.Parallel(); scannerShard(test, 4) }

func TestScannerAgreesWithTypescriptGo_005(test *testing.T) { test.Parallel(); scannerShard(test, 5) }

func TestScannerAgreesWithTypescriptGo_006(test *testing.T) { test.Parallel(); scannerShard(test, 6) }

func TestScannerAgreesWithTypescriptGo_007(test *testing.T) { test.Parallel(); scannerShard(test, 7) }

func TestScannerAgreesWithTypescriptGo_008(test *testing.T) { test.Parallel(); scannerShard(test, 8) }

func TestScannerAgreesWithTypescriptGo_009(test *testing.T) { test.Parallel(); scannerShard(test, 9) }

func TestScannerAgreesWithTypescriptGo_010(test *testing.T) { test.Parallel(); scannerShard(test, 10) }

func TestScannerAgreesWithTypescriptGo_011(test *testing.T) { test.Parallel(); scannerShard(test, 11) }

func TestScannerAgreesWithTypescriptGo_012(test *testing.T) { test.Parallel(); scannerShard(test, 12) }

func TestScannerAgreesWithTypescriptGo_013(test *testing.T) { test.Parallel(); scannerShard(test, 13) }

func TestScannerAgreesWithTypescriptGo_014(test *testing.T) { test.Parallel(); scannerShard(test, 14) }

func TestScannerAgreesWithTypescriptGo_015(test *testing.T) { test.Parallel(); scannerShard(test, 15) }
