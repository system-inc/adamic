package scanner

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
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
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

const scannerShardCount = 16

var scannerMutants = []struct{ name, file, from, to string }{
	{"source scanned as Identifier", "tokens.ts", `['source', 'SourceKeyword']`, `['source', 'Identifier']`},
	{"punctuator != scanned as ==", "tokens.ts", `['!=', 'ExclamationEqualsToken']`, `['!=', 'EqualsEqualsToken']`},
	{"invalid decimal separator accepted", "scanner.ts", "this.error(previous ? 6189 : 6188, this.pos, 1);", "if (base !== 10) { this.error(previous ? 6189 : 6188, this.pos, 1); } else { this.flags &= ~16384; }"},
	{"regex rescan skipped", "main.ts", "scanner.rescanSlash();", "scanner.code();"},
}

// Identity is content-addressed by buildcache; sync.Once only prevents repeat fetches in a process.
type scannerPrepared struct {
	once      sync.Once
	directory string
}

var scannerProducts sync.Map

func scannerFetch(t *testing.T, inputs buildcache.Inputs, recipe func(string) error) string {
	t.Helper()
	identity, err := json.Marshal(inputs)
	if err != nil {
		t.Fatal(err)
	}
	value, _ := scannerProducts.LoadOrStore(string(identity), &scannerPrepared{})
	prepared := value.(*scannerPrepared)
	prepared.once.Do(func() { prepared.directory = buildcache.Product(t, inputs, recipe) })
	if prepared.directory == "" {
		t.Fatal("scanner product preparation failed")
	}
	return prepared.directory
}

func scannerOracleProduct(t *testing.T) string {
	t.Helper()
	inputs := buildcache.Inputs{Name: "scanner-go-oracle", Files: []string{"cohere/TypeScript/tsc", "go.mod", "go.work", "stage1/typescript/scanner/testdata/oracle.go"}, Flags: []string{"go build", "-overlay", "GOTOOLCHAIN=" + os.Getenv("GOTOOLCHAIN"), "GOFLAGS=" + os.Getenv("GOFLAGS"), "GOWORK=" + os.Getenv("GOWORK"), "GOEXPERIMENT=" + os.Getenv("GOEXPERIMENT"), "CGO_ENABLED=" + os.Getenv("CGO_ENABLED"), "GOOS=" + runtime.GOOS, "GOARCH=" + runtime.GOARCH}, Toolchain: []string{buildcache.Tool("go", "version")}}
	return scannerFetch(t, inputs, func(directory string) error {
		root, err := filepath.Abs(filepath.Join(repository, "cohere/TypeScript/tsc"))
		if err != nil {
			return err
		}
		side, err := filepath.Abs("testdata/oracle.go")
		if err != nil {
			return err
		}
		virtual := filepath.Join(root, "adamic_scanner_oracle.go")
		overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: side}})
		if err != nil {
			return err
		}
		path := filepath.Join(directory, "overlay.json")
		if err := os.WriteFile(path, overlay, 0644); err != nil {
			return err
		}
		command := exec.CommandContext(context.Background(), "go", "build", "-overlay="+path, "-o", filepath.Join(directory, "oracle"), virtual)
		command.Dir = root
		output, err := command.CombinedOutput()
		if err != nil {
			return fmt.Errorf("oracle: %w\n%s", err, output)
		}
		return nil
	})
}

func scannerPortProduct(t *testing.T, source string, sanitize bool) string {
	t.Helper()
	options := native.Options{Sanitize: sanitize}
	flags := append(native.Flags(options), "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"), "ADAMIC_NATIVE_JOBS="+os.Getenv("ADAMIC_NATIVE_JOBS"), "ADAMIC_GATE_UNCACHED="+os.Getenv("ADAMIC_GATE_UNCACHED"))
	contents := make(map[string][]byte)
	for _, name := range portFiles {
		data, err := os.ReadFile(filepath.Join(source, name))
		if err != nil {
			t.Fatal(err)
		}
		contents[name] = data
		flags = append(flags, name+"="+string(data))
	}
	inputs := buildcache.Inputs{Name: "scanner-native", Files: []string{"internal", "cohere", "go.mod", "go.work"}, Flags: flags, Toolchain: []string{runtime.Version(), buildcache.Tool("clang", "--version")}}
	return scannerFetch(t, inputs, func(directory string) error {
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

func scannerNative(t *testing.T, mutant int, sanitize bool) string {
	t.Helper()
	source := "."
	if mutant >= 0 {
		m := scannerMutants[mutant]
		source = copyPort(t, m.file, m.from, m.to)
	}
	return scannerPortProduct(t, source, sanitize)
}

func scannerProductTime(t *testing.T, build func() string) {
	t.Helper()
	started := time.Now()
	build()
	t.Logf("product cold/fetch time %.3fs", time.Since(started).Seconds())
}

func TestProduct_ScannerOracle(t *testing.T) {
	t.Parallel()
	scannerProductTime(t, func() string { return scannerOracleProduct(t) })
}
func TestProduct_ScannerNative(t *testing.T) {
	t.Parallel()
	scannerProductTime(t, func() string { return scannerNative(t, -1, true) })
}
func TestProduct_ScannerRelease(t *testing.T) {
	t.Parallel()
	scannerProductTime(t, func() string { return scannerNative(t, -1, false) })
}
func TestProduct_ScannerSourceIdentifier(t *testing.T) {
	t.Parallel()
	scannerProductTime(t, func() string { return scannerNative(t, 0, true) })
}
func TestProduct_ScannerPunctuator(t *testing.T) {
	t.Parallel()
	scannerProductTime(t, func() string { return scannerNative(t, 1, true) })
}
func TestProduct_ScannerDecimalSeparator(t *testing.T) {
	t.Parallel()
	scannerProductTime(t, func() string { return scannerNative(t, 2, true) })
}
func TestProduct_ScannerRegexRescan(t *testing.T) {
	t.Parallel()
	scannerProductTime(t, func() string { return scannerNative(t, 3, true) })
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

func scannerManifestLines(t *testing.T, path string) []string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
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

func scannerRun(t *testing.T, ctx context.Context, name string, args ...string) []byte {
	t.Helper()
	command := exec.CommandContext(ctx, name, args...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error { return syscall.Kill(-command.Process.Pid, syscall.SIGKILL) }
	command.WaitDelay = 3 * time.Second
	var stderr bytes.Buffer
	command.Stderr = &stderr
	output, err := command.Output()
	if err != nil || stderr.Len() != 0 {
		t.Fatalf("%s: %v\n%s", name, err, &stderr)
	}
	return output
}

func scannerNode(t *testing.T, ctx context.Context, directory, manifest string) []byte {
	t.Helper()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	return scannerRun(t, ctx, "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.ts"), "--manifest", manifest)
}

func scannerShard(t *testing.T, shard int) {
	t.Helper()
	setup := time.Now()
	oracle := goOracle(t)
	base := scannerNative(t, -1, true)
	products := make(map[int]string)
	for i, m := range scannerMutants {
		if scannerOwner("mutant:"+m.name) == shard {
			products[i] = scannerNative(t, i, true)
		}
	}
	asked := askedCorpus(t)
	lines := scannerSelect(scannerManifestLines(t, asked.all), shard)
	manifest := filepath.Join(t.TempDir(), "shard.txt")
	if err := os.WriteFile(manifest, []byte(strings.Join(lines, "\n")+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	setupTime := time.Since(setup)
	started := time.Now()
	// Setup carries no test-side deadline. Only this shard's comparisons are budgeted.
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	want := scannerRun(t, ctx, oracle, "--manifest", manifest)
	got := scannerRun(t, ctx, filepath.Join(base, "scanner"), "--manifest", manifest)
	if diff := difference(got, want); diff != "" {
		t.Errorf("native: %s", diff)
	}
	got = scannerNode(t, ctx, base, manifest)
	if diff := difference(got, want); diff != "" {
		t.Errorf("Node: %s", diff)
	}
	// The opt-in planted answer corruption is owned by precisely the shard containing this case.
	if os.Getenv("ADAMIC_SCANNER_PLANT_FAILURE") == "1" {
		whole := scannerManifestLines(t, asked.all)
		planted := scannerCaseIdentity(whole[0])
		for _, line := range lines {
			if scannerCaseIdentity(line) == planted {
				if difference(append([]byte("planted failure\n"), got...), want) != "" {
					t.Error("caught planted answer corruption")
				}
			}
		}
	}
	if len(products) > 0 {
		edgeWant := scannerRun(t, ctx, oracle, "--manifest", asked.edges)
		for i, product := range products {
			for _, side := range []struct {
				name   string
				output []byte
			}{
				{"Node", scannerNode(t, ctx, product, asked.edges)},
				{"native", scannerRun(t, ctx, filepath.Join(product, "scanner"), "--manifest", asked.edges)},
			} {
				if diff := difference(side.output, edgeWant); diff == "" {
					t.Errorf("%s %s: mutant survives comparison", scannerMutants[i].name, side.name)
				} else {
					t.Logf("%s %s caught: %s", scannerMutants[i].name, side.name, diff)
				}
			}
		}
	}
	t.Logf("shard %03d: %d cases, %d mutants; setup %.3fs / own work %.3fs", shard, len(lines), len(products), setupTime.Seconds(), time.Since(started).Seconds())
}

func TestScannerShardCoverage(t *testing.T) {
	t.Parallel()
	asked := askedCorpus(t)
	whole := scannerManifestLines(t, asked.all)
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
		t.Logf("shard %03d owns %d cases", shard, len(selected))
	}
	if len(whole) != asked.files+asked.generated {
		t.Fatalf("whole count %d != %d files + %d generated", len(whole), asked.files, asked.generated)
	}
	if len(counts) != len(whole) {
		t.Fatalf("union %d != whole %d", len(counts), len(whole))
	}
	for _, line := range whole {
		if counts[scannerCaseIdentity(line)] != 1 {
			t.Errorf("case ownership %q: %d", line, counts[scannerCaseIdentity(line)])
		}
	}
	for i, count := range mutantCounts {
		if count != 1 {
			t.Errorf("mutant %s owned %d times", scannerMutants[i].name, count)
		}
	}
	if caught != 1 {
		t.Fatalf("planted failure caught by %d shards", caught)
	}
	t.Logf("counted union: %d cases and %d mutants exactly once; planted corruption caught once", len(whole), len(scannerMutants))
}

func TestScannerAgreesWithTypescriptGo_000(t *testing.T) { t.Parallel(); scannerShard(t, 0) }

func TestScannerAgreesWithTypescriptGo_001(t *testing.T) { t.Parallel(); scannerShard(t, 1) }

func TestScannerAgreesWithTypescriptGo_002(t *testing.T) { t.Parallel(); scannerShard(t, 2) }

func TestScannerAgreesWithTypescriptGo_003(t *testing.T) { t.Parallel(); scannerShard(t, 3) }

func TestScannerAgreesWithTypescriptGo_004(t *testing.T) { t.Parallel(); scannerShard(t, 4) }

func TestScannerAgreesWithTypescriptGo_005(t *testing.T) { t.Parallel(); scannerShard(t, 5) }

func TestScannerAgreesWithTypescriptGo_006(t *testing.T) { t.Parallel(); scannerShard(t, 6) }

func TestScannerAgreesWithTypescriptGo_007(t *testing.T) { t.Parallel(); scannerShard(t, 7) }

func TestScannerAgreesWithTypescriptGo_008(t *testing.T) { t.Parallel(); scannerShard(t, 8) }

func TestScannerAgreesWithTypescriptGo_009(t *testing.T) { t.Parallel(); scannerShard(t, 9) }

func TestScannerAgreesWithTypescriptGo_010(t *testing.T) { t.Parallel(); scannerShard(t, 10) }

func TestScannerAgreesWithTypescriptGo_011(t *testing.T) { t.Parallel(); scannerShard(t, 11) }

func TestScannerAgreesWithTypescriptGo_012(t *testing.T) { t.Parallel(); scannerShard(t, 12) }

func TestScannerAgreesWithTypescriptGo_013(t *testing.T) { t.Parallel(); scannerShard(t, 13) }

func TestScannerAgreesWithTypescriptGo_014(t *testing.T) { t.Parallel(); scannerShard(t, 14) }

func TestScannerAgreesWithTypescriptGo_015(t *testing.T) { t.Parallel(); scannerShard(t, 15) }
