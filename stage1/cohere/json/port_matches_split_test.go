package json

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"hash/fnv"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Fixed buckets retain their identities when the repository corpus grows.
const testPortMatchesGoCohereSplitShards = 128

func portMatchesBucket(name string) int {
	h := fnv.New64a()
	_, _ = h.Write([]byte(name))
	return int(h.Sum64() % testPortMatchesGoCohereSplitShards)
}

func portMatchesPartition(t *testing.T) [][]textCase {
	t.Helper()
	cases := sampledCorpusCases(t, 32)
	shards := make([][]textCase, testPortMatchesGoCohereSplitShards)
	seen := make(map[string]textCase, len(cases))
	for _, item := range cases {
		if _, exists := seen[item.Name]; exists {
			t.Fatalf("duplicate case %q", item.Name)
		}
		seen[item.Name] = item
		n := portMatchesBucket(item.Name)
		shards[n] = append(shards[n], item)
	}
	count := 0
	for n, items := range shards {
		for _, item := range items {
			want, exists := seen[item.Name]
			if !exists || want != item || portMatchesBucket(item.Name) != n {
				t.Fatalf("shard %d invalid case %q", n, item.Name)
			}
			delete(seen, item.Name)
			count++
		}
	}
	if count != len(cases) || len(seen) != 0 {
		t.Fatalf("union %d of %d", count, len(cases))
	}
	// Check the actual top-level enumeration, including wrapper ordinals.
	file, err := parser.ParseFile(token.NewFileSet(), "port_matches_split_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	declared := make(map[string]bool)
	for _, d := range file.Decls {
		if f, ok := d.(*ast.FuncDecl); ok {
			declared[f.Name.Name] = true
		}
	}
	for n := 0; n < testPortMatchesGoCohereSplitShards; n++ {
		name := fmt.Sprintf("TestPortMatchesGoCohere_%03d", n)
		if !declared[name] {
			t.Fatalf("missing %s", name)
		}
		delete(declared, name)
	}
	for name := range declared {
		if strings.HasPrefix(name, "TestPortMatchesGoCohere_") {
			t.Fatalf("unexpected shard %s", name)
		}
	}
	t.Logf("shard union: %d cases exactly once across %d shards", count, len(shards))
	return shards
}

var portMatchesShared struct {
	ready                                     bool
	shards                                    [][]textCase
	oracle, entry, script, release, sanitized string
}

// Not parallel: prepares shared immutable products before any parallel leaf resumes.
func TestPortMatchesGoCohereSplit_Setup(t *testing.T) {
	deadline := portMatchesDeadline(t.Name())
	defer deadline.Stop()

	started := time.Now()
	defer func() { t.Logf("TestPortMatchesGoCohere (setup): %.3fs", time.Since(started).Seconds()) }()
	portMatchesShared.shards = portMatchesPartition(t)
	goTools, err := portMatchesJsonGoToolchain()
	if err != nil {
		t.Fatal(err)
	}
	clangTools, err := portMatchesJsonClangToolchain()
	if err != nil {
		t.Fatal(err)
	}
	oracleDir := buildcache.Product(t, buildcache.Inputs{
		Name:      "json split Go oracle",
		Files:     []string{"go.mod", "go.work", "cohere", "stage1/cohere/json/testdata/cohere_driver.go", "stage1/cohere/json/port_matches_split_test.go"},
		Flags:     append([]string{"go build -overlay=<product>/overlay.json -o=<product>/go-cohere ./command/formatter_comparison"}, portMatchesJsonBuildEnvironment("PATH", "GOFLAGS", "GOTOOLCHAIN")...),
		Toolchain: goTools,
	}, portMatchesBuildJSONGoOracle)

	portMatchesShared.oracle = filepath.Join(oracleDir, "go-cohere")
	loweredDir := buildcache.Product(t, portMatchesJsonLoweredPortInputs(goTools), portMatchesBuildJSONLoweredPort)
	cBytes, err := os.ReadFile(filepath.Join(loweredDir, "main.c"))
	if err != nil {
		t.Fatal(err)
	}
	c := string(cBytes)
	portMatchesShared.entry = filepath.Join(loweredDir, "main.ts")
	portMatchesShared.script = filepath.Join(loweredDir, "program.mjs")
	portMatchesShared.release = filepath.Join(buildcache.Product(t, portMatchesJsonNativePortInputs(c, false, clangTools), portMatchesBuildJSONReleasePort(c)), "port")
	portMatchesShared.sanitized = filepath.Join(buildcache.Product(t, portMatchesJsonNativePortInputs(c, true, clangTools), portMatchesBuildJSONSanitizedPort(c)), "port")
	// The four-digit port/upstream wrappers use the same immutable products.
	// Complete their once guards here; no leaf is allowed to become a builder.
	jsonTopCorpus(t)
	jsonTopState.goOnce.Do(func() { jsonTopState.goTools = goTools })
	jsonTopState.clangOnce.Do(func() { jsonTopState.clangTools = clangTools })
	jsonTopState.oracle.Do(func() {
		jsonTopState.oracleDir = oracleDir
		jsonTopState.oracleCached = true
	})
	for name, directory := range map[string]string{
		"lowered":   loweredDir,
		"release":   filepath.Dir(portMatchesShared.release),
		"sanitized": filepath.Dir(portMatchesShared.sanitized),
	} {
		product := &jsonTopBuild{dir: directory}
		product.once.Do(func() {})
		jsonTopState.products.Store(name, product)
	}
	portMatchesShared.ready = true
}

func portMatchesRun(t *testing.T, ordinal int) {
	t.Helper()
	if !portMatchesShared.ready {
		t.Fatal("shared setup was not selected")
	}
	deadline := portMatchesDeadline(t.Name())
	defer deadline.Stop()
	items := portMatchesShared.shards[ordinal]
	if len(items) == 0 {
		t.Log("empty hash bucket")
		return
	}
	started := time.Now()
	input, _ := protocol(items, make([]answer, len(items)))
	path := filepath.Join(t.TempDir(), "cases.txt")
	if err := os.WriteFile(path, []byte(input), 0644); err != nil {
		t.Fatal(err)
	}
	reference := execute(t, nil, portMatchesShared.oracle, "--cases", path)
	if reference.exitCode != 0 || len(reference.stderr) != 0 {
		t.Fatalf("Go oracle exit %d: %s", reference.exitCode, reference.stderr)
	}
	expected := string(reference.stdout)
	if len(strings.Split(strings.TrimSuffix(expected, "\n"), "\n")) != len(items) {
		t.Fatalf("Go oracle did not answer %d cases", len(items))
	}
	compare(t, "Node", onNode(t, portMatchesShared.entry, "--cases", path), expected, items)
	release := execute(t, nil, portMatchesShared.release, "--cases", path)
	compare(t, "release", release, expected, items)
	// Bound sanitizer inputs exactly as the original test does.
	answers := portMatchesAnswers(t, expected, len(items))
	chunks := nativeChunks(t, items, answers)
	sanitized := runNativeChunks(t, portMatchesShared.sanitized, []string{"ASAN_OPTIONS=detect_leaks=0"}, chunks, release, items, "native ASan/UBSan")
	compare(t, "native ASan/UBSan", sanitized, expected, items)
	if runtime.GOOS == "linux" {
		leaked := runNativeChunks(t, portMatchesShared.sanitized, []string{"ASAN_OPTIONS=detect_leaks=1"}, chunks, release, items, "LeakSanitizer")
		compare(t, "LeakSanitizer", leaked, expected, items)
	} else {
		for _, chunk := range chunks {
			report := execute(t, nil, "leaks", "--atExit", "--", portMatchesShared.release, "--cases", chunk.path)
			if report.exitCode != 0 {
				t.Fatalf("leaks: %s", report.stdout)
			}
		}
	}
	compare(t, "JavaScript backend", onNode(t, portMatchesShared.script, "--cases", path), expected, items)
	if os.Getenv("ADAMIC_JSON_GUARD_CALIBRATE") == "1" {
		calibrateNativeCases(t, portMatchesShared.sanitized, items, answers)
	}
	if os.Getenv("ADAMIC_JSON_BENCH") == "1" {
		for round := 0; round < 3; round++ {
			compare(t, "release timed", execute(t, nil, portMatchesShared.release, "--cases", path), expected, items)
			compare(t, "Node timed", onNode(t, portMatchesShared.entry, "--cases", path), expected, items)
			compare(t, "Go timed", execute(t, nil, portMatchesShared.oracle, "--cases", path), expected, items)
		}
	}
	t.Logf("shard %03d: %d cases; execution %.3fs", ordinal, len(items), time.Since(started).Seconds())
}

func portMatchesAnswers(t *testing.T, protocol string, count int) []answer {
	t.Helper()
	lines := strings.Split(strings.TrimSuffix(protocol, "\n"), "\n")
	if len(lines) != count {
		t.Fatal("oracle answer count")
	}
	answers := make([]answer, count)
	unescape := strings.NewReplacer("\\\\", "\\", "\\n", "\n", "\\r", "\r", "\\t", "\t")
	for i, line := range lines {
		kind, value, ok := strings.Cut(line, "\t")
		if !ok {
			t.Fatal("oracle protocol")
		}
		switch kind {
		case "ok":
			answers[i].Output = unescape.Replace(value)
		case "error":
			answers[i].Error = unescape.Replace(value)
		default:
			t.Fatal("oracle status")
		}
	}
	return answers
}

// A single changed oracle answer must be caught by exactly its owning shard.
func TestPortMatchesGoCohereSplitUnion(t *testing.T) {
	t.Parallel()
	shards := portMatchesPartition(t)
	var planted string
	for _, items := range shards {
		if len(items) > 0 {
			planted = items[0].Name
			break
		}
	}
	if planted == "" {
		t.Fatal("empty corpus")
	}
	caught := 0
	for n, items := range shards {
		answers := make([]answer, len(items))
		_, expected := protocol(items, answers)
		for i, item := range items {
			if item.Name == planted {
				answers[i].Output = "planted disagreement"
			}
		}
		_, actual := protocol(items, answers)
		if comparisonError("planted", run{stdout: []byte(actual)}, expected, items) != nil {
			caught++
			if n != portMatchesBucket(planted) {
				t.Fatal("wrong shard caught planted failure")
			}
			t.Logf("planted %q caught by TestPortMatchesGoCohere_%03d", planted, n)
		}
	}
	if caught != 1 {
		t.Fatalf("planted failure caught by %d shards", caught)
	}
}

// Build inputs use the shared cache API; no package-local disk cache.
type portMatchesJsonBuildInputs = buildcache.Inputs

func portMatchesJsonBuildEnvironment(names ...string) []string {
	var flags []string
	for _, name := range names {
		value, set := os.LookupEnv(name)
		flags = append(flags, fmt.Sprintf("env:%s:set=%t:value=%s", name, set, value))
	}
	return flags
}

func portMatchesJsonGoToolchain() ([]string, error) {
	command := exec.Command("go", "version")
	directory, err := filepath.Abs(filepath.Join(repository, "cohere"))
	if err != nil {
		return nil, err
	}
	command.Dir = directory
	version, err := command.Output()
	if err != nil {
		return nil, err
	}
	configuration := exec.Command("go", "env", "-json")
	configuration.Dir = directory
	settings, err := configuration.Output()
	if err != nil {
		return nil, err
	}
	// Go generates a fresh scratch directory even for go env. It is mapped to
	// /tmp/go-build in compiler output and cannot affect a product, so normalize
	// only that generated prefix-map argument; retain every real compiler flag.
	var resolved map[string]string
	if err := json.Unmarshal(settings, &resolved); err != nil {
		return nil, err
	}
	compilerFlags := strings.Fields(resolved["GOGCCFLAGS"])
	for i, flag := range compilerFlags {
		if strings.HasPrefix(flag, "-ffile-prefix-map=") && strings.HasSuffix(flag, "=/tmp/go-build") {
			compilerFlags[i] = "-ffile-prefix-map=<Go scratch>=/tmp/go-build"
		}
	}
	resolved["GOGCCFLAGS"] = strings.Join(compilerFlags, " ")
	settings, err = json.Marshal(resolved)
	if err != nil {
		return nil, err
	}
	// Include resolved Go settings as well as explicit environment overrides.
	return []string{"runtime.Version=" + runtime.Version(), string(version), "go env=" + string(settings), "GOOS=" + runtime.GOOS, "GOARCH=" + runtime.GOARCH}, nil
}

func portMatchesJsonClangToolchain() ([]string, error) {
	clang, err := exec.LookPath("clang")
	if err != nil {
		return nil, err
	}
	version, err := exec.Command(clang, "--version").Output()
	if err != nil {
		return nil, err
	}
	archiver := filepath.Join(filepath.Dir(clang), "llvm-ar")
	if info, err := os.Stat(archiver); err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0111 == 0 {
		archiver, err = exec.LookPath("ar")
		if err != nil {
			return nil, err
		}
	}
	arVersion, err := exec.Command(archiver, "--version").Output()
	if err != nil {
		return nil, err
	}
	return []string{clang, string(version), archiver, string(arVersion), "GOOS=" + runtime.GOOS, "GOARCH=" + runtime.GOARCH}, nil
}

// portMatchesBuildJSONGoOracle writes the driver and overlay only into its product directory.
// Go's usual module/action caches remain managed by the Go toolchain.
func portMatchesBuildJSONGoOracle(dir string) error {
	cohere, err := filepath.Abs(filepath.Join(repository, "cohere"))
	if err != nil {
		return err
	}
	source, err := filepath.Abs("testdata/cohere_driver.go")
	if err != nil {
		return err
	}
	overlay := filepath.Join(dir, "overlay.json")
	encoded, err := json.Marshal(map[string]any{"Replace": map[string]string{filepath.Join(cohere, "command/formatter_comparison/main.go"): source}})
	if err != nil {
		return err
	}
	if err := os.WriteFile(overlay, encoded, 0644); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "build", "-overlay="+overlay, "-o", filepath.Join(dir, "go-cohere"), "./command/formatter_comparison")
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if err == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return err
	}
	command.WaitDelay = time.Second
	command.Dir = cohere
	if output, err := childguard.CombinedOutput(command, jsonGuard); err != nil {
		return fmt.Errorf("Go driver: %w\n%s", err, output)
	}
	return nil
}

func portMatchesJsonLoweredPortInputs(toolchain []string) portMatchesJsonBuildInputs {
	files := []string{"go.mod", "go.work", "internal", "cohere/TypeScript", "cohere/TypeScript-shim", "cohere/rule_runner", "cohere/static_single_assignment", "cohere/mutation_aliasing", "stage1/cohere/json/port_matches_split_test.go"}
	for _, name := range portFiles {
		files = append(files, "stage1/cohere/json/"+name)
	}
	return portMatchesJsonBuildInputs{Name: "json lowered port and backend sources", Files: files, Flags: append([]string{"load.Load(main.ts)", "lower.Lower", "native.C", "javascript.JavaScript"}, portMatchesJsonBuildEnvironment("PATH", "GOFLAGS", "GOTOOLCHAIN", "GOOS", "GOARCH", "GOAMD64", "GOARM64", "CGO_ENABLED", "GOEXPERIMENT", "GOENV")...), Toolchain: toolchain}
}

// The immutable lowering product contains the copied TypeScript and both emitted
// backends. Persisting emitted sources avoids serializing IR interface values;
// fetching this product will avoid loading, checking, lowering and emission.
func portMatchesBuildJSONLoweredPort(dir string) error {
	for _, name := range portFiles {
		contents, err := os.ReadFile(name)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, name), contents, 0644); err != nil {
			return err
		}
	}
	program, err := load.Load([]string{filepath.Join(dir, "main.ts")})
	if err != nil {
		return fmt.Errorf("Load: %w", err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		return fmt.Errorf("Lower: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.c"), []byte(native.C(lowered)), 0644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "program.mjs"), []byte(javascript.JavaScript(lowered)), 0644)
}

func portMatchesJsonNativePortInputs(source string, sanitize bool, toolchain []string) portMatchesJsonBuildInputs {
	options := native.Options{Sanitize: sanitize}
	name := "json native release port"
	if sanitize {
		name = "json native ASan UBSan port"
	}
	flags := append([]string(nil), native.Flags(options)...)
	// Generated C is a captured input, not a repository file. Its digest includes
	// emitted feature defines; internal/ also covers runtime sources/link policy.
	flags = append(flags, fmt.Sprintf("generated-C-sha256=%x", sha256.Sum256([]byte(source))), fmt.Sprintf("options=%+v", options), "-I=<runtime product>", "-o=<dir>/port", "-lm")
	flags = append(flags, portMatchesJsonBuildEnvironment("PATH", "ADAMIC_NATIVE_SPLIT", "ADAMIC_NATIVE_JOBS", "ADAMIC_GATE_UNCACHED", "XDG_CACHE_HOME", "HOME", "TMPDIR", "CPATH", "C_INCLUDE_PATH", "LIBRARY_PATH", "SDKROOT", "MACOSX_DEPLOYMENT_TARGET")...)
	return portMatchesJsonBuildInputs{Name: name, Files: []string{"internal", "stage1/cohere/json/port_matches_split_test.go"}, Flags: flags, Toolchain: toolchain}
}

// Each factory returns the exact func(dir string) error accepted by buildcache.
// Native.Build retains its existing compiler scratch/runtime action caches.
func portMatchesBuildJSONReleasePort(source string) func(string) error {
	return func(dir string) error { return native.Build(source, filepath.Join(dir, "port"), native.Options{}) }
}
func portMatchesBuildJSONSanitizedPort(source string) func(string) error {
	return func(dir string) error {
		return native.Build(source, filepath.Join(dir, "port"), native.Options{Sanitize: true})
	}
}

func TestPortMatchesGoCohere_000(t *testing.T) { t.Parallel(); portMatchesRun(t, 0) }
func TestPortMatchesGoCohere_001(t *testing.T) { t.Parallel(); portMatchesRun(t, 1) }
func TestPortMatchesGoCohere_002(t *testing.T) { t.Parallel(); portMatchesRun(t, 2) }
func TestPortMatchesGoCohere_003(t *testing.T) { t.Parallel(); portMatchesRun(t, 3) }
func TestPortMatchesGoCohere_004(t *testing.T) { t.Parallel(); portMatchesRun(t, 4) }
func TestPortMatchesGoCohere_005(t *testing.T) { t.Parallel(); portMatchesRun(t, 5) }
func TestPortMatchesGoCohere_006(t *testing.T) { t.Parallel(); portMatchesRun(t, 6) }
func TestPortMatchesGoCohere_007(t *testing.T) { t.Parallel(); portMatchesRun(t, 7) }
func TestPortMatchesGoCohere_008(t *testing.T) { t.Parallel(); portMatchesRun(t, 8) }
func TestPortMatchesGoCohere_009(t *testing.T) { t.Parallel(); portMatchesRun(t, 9) }
func TestPortMatchesGoCohere_010(t *testing.T) { t.Parallel(); portMatchesRun(t, 10) }
func TestPortMatchesGoCohere_011(t *testing.T) { t.Parallel(); portMatchesRun(t, 11) }
func TestPortMatchesGoCohere_012(t *testing.T) { t.Parallel(); portMatchesRun(t, 12) }
func TestPortMatchesGoCohere_013(t *testing.T) { t.Parallel(); portMatchesRun(t, 13) }
func TestPortMatchesGoCohere_014(t *testing.T) { t.Parallel(); portMatchesRun(t, 14) }
func TestPortMatchesGoCohere_015(t *testing.T) { t.Parallel(); portMatchesRun(t, 15) }
func TestPortMatchesGoCohere_016(t *testing.T) { t.Parallel(); portMatchesRun(t, 16) }
func TestPortMatchesGoCohere_017(t *testing.T) { t.Parallel(); portMatchesRun(t, 17) }
func TestPortMatchesGoCohere_018(t *testing.T) { t.Parallel(); portMatchesRun(t, 18) }
func TestPortMatchesGoCohere_019(t *testing.T) { t.Parallel(); portMatchesRun(t, 19) }
func TestPortMatchesGoCohere_020(t *testing.T) { t.Parallel(); portMatchesRun(t, 20) }
func TestPortMatchesGoCohere_021(t *testing.T) { t.Parallel(); portMatchesRun(t, 21) }
func TestPortMatchesGoCohere_022(t *testing.T) { t.Parallel(); portMatchesRun(t, 22) }
func TestPortMatchesGoCohere_023(t *testing.T) { t.Parallel(); portMatchesRun(t, 23) }
func TestPortMatchesGoCohere_024(t *testing.T) { t.Parallel(); portMatchesRun(t, 24) }
func TestPortMatchesGoCohere_025(t *testing.T) { t.Parallel(); portMatchesRun(t, 25) }
func TestPortMatchesGoCohere_026(t *testing.T) { t.Parallel(); portMatchesRun(t, 26) }
func TestPortMatchesGoCohere_027(t *testing.T) { t.Parallel(); portMatchesRun(t, 27) }
func TestPortMatchesGoCohere_028(t *testing.T) { t.Parallel(); portMatchesRun(t, 28) }
func TestPortMatchesGoCohere_029(t *testing.T) { t.Parallel(); portMatchesRun(t, 29) }
func TestPortMatchesGoCohere_030(t *testing.T) { t.Parallel(); portMatchesRun(t, 30) }
func TestPortMatchesGoCohere_031(t *testing.T) { t.Parallel(); portMatchesRun(t, 31) }
func TestPortMatchesGoCohere_032(t *testing.T) { t.Parallel(); portMatchesRun(t, 32) }
func TestPortMatchesGoCohere_033(t *testing.T) { t.Parallel(); portMatchesRun(t, 33) }
func TestPortMatchesGoCohere_034(t *testing.T) { t.Parallel(); portMatchesRun(t, 34) }
func TestPortMatchesGoCohere_035(t *testing.T) { t.Parallel(); portMatchesRun(t, 35) }
func TestPortMatchesGoCohere_036(t *testing.T) { t.Parallel(); portMatchesRun(t, 36) }
func TestPortMatchesGoCohere_037(t *testing.T) { t.Parallel(); portMatchesRun(t, 37) }
func TestPortMatchesGoCohere_038(t *testing.T) { t.Parallel(); portMatchesRun(t, 38) }
func TestPortMatchesGoCohere_039(t *testing.T) { t.Parallel(); portMatchesRun(t, 39) }
func TestPortMatchesGoCohere_040(t *testing.T) { t.Parallel(); portMatchesRun(t, 40) }
func TestPortMatchesGoCohere_041(t *testing.T) { t.Parallel(); portMatchesRun(t, 41) }
func TestPortMatchesGoCohere_042(t *testing.T) { t.Parallel(); portMatchesRun(t, 42) }
func TestPortMatchesGoCohere_043(t *testing.T) { t.Parallel(); portMatchesRun(t, 43) }
func TestPortMatchesGoCohere_044(t *testing.T) { t.Parallel(); portMatchesRun(t, 44) }
func TestPortMatchesGoCohere_045(t *testing.T) { t.Parallel(); portMatchesRun(t, 45) }
func TestPortMatchesGoCohere_046(t *testing.T) { t.Parallel(); portMatchesRun(t, 46) }
func TestPortMatchesGoCohere_047(t *testing.T) { t.Parallel(); portMatchesRun(t, 47) }
func TestPortMatchesGoCohere_048(t *testing.T) { t.Parallel(); portMatchesRun(t, 48) }
func TestPortMatchesGoCohere_049(t *testing.T) { t.Parallel(); portMatchesRun(t, 49) }
func TestPortMatchesGoCohere_050(t *testing.T) { t.Parallel(); portMatchesRun(t, 50) }
func TestPortMatchesGoCohere_051(t *testing.T) { t.Parallel(); portMatchesRun(t, 51) }
func TestPortMatchesGoCohere_052(t *testing.T) { t.Parallel(); portMatchesRun(t, 52) }
func TestPortMatchesGoCohere_053(t *testing.T) { t.Parallel(); portMatchesRun(t, 53) }
func TestPortMatchesGoCohere_054(t *testing.T) { t.Parallel(); portMatchesRun(t, 54) }
func TestPortMatchesGoCohere_055(t *testing.T) { t.Parallel(); portMatchesRun(t, 55) }
func TestPortMatchesGoCohere_056(t *testing.T) { t.Parallel(); portMatchesRun(t, 56) }
func TestPortMatchesGoCohere_057(t *testing.T) { t.Parallel(); portMatchesRun(t, 57) }
func TestPortMatchesGoCohere_058(t *testing.T) { t.Parallel(); portMatchesRun(t, 58) }
func TestPortMatchesGoCohere_059(t *testing.T) { t.Parallel(); portMatchesRun(t, 59) }
func TestPortMatchesGoCohere_060(t *testing.T) { t.Parallel(); portMatchesRun(t, 60) }
func TestPortMatchesGoCohere_061(t *testing.T) { t.Parallel(); portMatchesRun(t, 61) }
func TestPortMatchesGoCohere_062(t *testing.T) { t.Parallel(); portMatchesRun(t, 62) }
func TestPortMatchesGoCohere_063(t *testing.T) { t.Parallel(); portMatchesRun(t, 63) }
func TestPortMatchesGoCohere_064(t *testing.T) { t.Parallel(); portMatchesRun(t, 64) }
func TestPortMatchesGoCohere_065(t *testing.T) { t.Parallel(); portMatchesRun(t, 65) }
func TestPortMatchesGoCohere_066(t *testing.T) { t.Parallel(); portMatchesRun(t, 66) }
func TestPortMatchesGoCohere_067(t *testing.T) { t.Parallel(); portMatchesRun(t, 67) }
func TestPortMatchesGoCohere_068(t *testing.T) { t.Parallel(); portMatchesRun(t, 68) }
func TestPortMatchesGoCohere_069(t *testing.T) { t.Parallel(); portMatchesRun(t, 69) }
func TestPortMatchesGoCohere_070(t *testing.T) { t.Parallel(); portMatchesRun(t, 70) }
func TestPortMatchesGoCohere_071(t *testing.T) { t.Parallel(); portMatchesRun(t, 71) }
func TestPortMatchesGoCohere_072(t *testing.T) { t.Parallel(); portMatchesRun(t, 72) }
func TestPortMatchesGoCohere_073(t *testing.T) { t.Parallel(); portMatchesRun(t, 73) }
func TestPortMatchesGoCohere_074(t *testing.T) { t.Parallel(); portMatchesRun(t, 74) }
func TestPortMatchesGoCohere_075(t *testing.T) { t.Parallel(); portMatchesRun(t, 75) }
func TestPortMatchesGoCohere_076(t *testing.T) { t.Parallel(); portMatchesRun(t, 76) }
func TestPortMatchesGoCohere_077(t *testing.T) { t.Parallel(); portMatchesRun(t, 77) }
func TestPortMatchesGoCohere_078(t *testing.T) { t.Parallel(); portMatchesRun(t, 78) }
func TestPortMatchesGoCohere_079(t *testing.T) { t.Parallel(); portMatchesRun(t, 79) }
func TestPortMatchesGoCohere_080(t *testing.T) { t.Parallel(); portMatchesRun(t, 80) }
func TestPortMatchesGoCohere_081(t *testing.T) { t.Parallel(); portMatchesRun(t, 81) }
func TestPortMatchesGoCohere_082(t *testing.T) { t.Parallel(); portMatchesRun(t, 82) }
func TestPortMatchesGoCohere_083(t *testing.T) { t.Parallel(); portMatchesRun(t, 83) }
func TestPortMatchesGoCohere_084(t *testing.T) { t.Parallel(); portMatchesRun(t, 84) }
func TestPortMatchesGoCohere_085(t *testing.T) { t.Parallel(); portMatchesRun(t, 85) }
func TestPortMatchesGoCohere_086(t *testing.T) { t.Parallel(); portMatchesRun(t, 86) }
func TestPortMatchesGoCohere_087(t *testing.T) { t.Parallel(); portMatchesRun(t, 87) }
func TestPortMatchesGoCohere_088(t *testing.T) { t.Parallel(); portMatchesRun(t, 88) }
func TestPortMatchesGoCohere_089(t *testing.T) { t.Parallel(); portMatchesRun(t, 89) }
func TestPortMatchesGoCohere_090(t *testing.T) { t.Parallel(); portMatchesRun(t, 90) }
func TestPortMatchesGoCohere_091(t *testing.T) { t.Parallel(); portMatchesRun(t, 91) }
func TestPortMatchesGoCohere_092(t *testing.T) { t.Parallel(); portMatchesRun(t, 92) }
func TestPortMatchesGoCohere_093(t *testing.T) { t.Parallel(); portMatchesRun(t, 93) }
func TestPortMatchesGoCohere_094(t *testing.T) { t.Parallel(); portMatchesRun(t, 94) }
func TestPortMatchesGoCohere_095(t *testing.T) { t.Parallel(); portMatchesRun(t, 95) }
func TestPortMatchesGoCohere_096(t *testing.T) { t.Parallel(); portMatchesRun(t, 96) }
func TestPortMatchesGoCohere_097(t *testing.T) { t.Parallel(); portMatchesRun(t, 97) }
func TestPortMatchesGoCohere_098(t *testing.T) { t.Parallel(); portMatchesRun(t, 98) }
func TestPortMatchesGoCohere_099(t *testing.T) { t.Parallel(); portMatchesRun(t, 99) }
func TestPortMatchesGoCohere_100(t *testing.T) { t.Parallel(); portMatchesRun(t, 100) }
func TestPortMatchesGoCohere_101(t *testing.T) { t.Parallel(); portMatchesRun(t, 101) }
func TestPortMatchesGoCohere_102(t *testing.T) { t.Parallel(); portMatchesRun(t, 102) }
func TestPortMatchesGoCohere_103(t *testing.T) { t.Parallel(); portMatchesRun(t, 103) }
func TestPortMatchesGoCohere_104(t *testing.T) { t.Parallel(); portMatchesRun(t, 104) }
func TestPortMatchesGoCohere_105(t *testing.T) { t.Parallel(); portMatchesRun(t, 105) }
func TestPortMatchesGoCohere_106(t *testing.T) { t.Parallel(); portMatchesRun(t, 106) }
func TestPortMatchesGoCohere_107(t *testing.T) { t.Parallel(); portMatchesRun(t, 107) }
func TestPortMatchesGoCohere_108(t *testing.T) { t.Parallel(); portMatchesRun(t, 108) }
func TestPortMatchesGoCohere_109(t *testing.T) { t.Parallel(); portMatchesRun(t, 109) }
func TestPortMatchesGoCohere_110(t *testing.T) { t.Parallel(); portMatchesRun(t, 110) }
func TestPortMatchesGoCohere_111(t *testing.T) { t.Parallel(); portMatchesRun(t, 111) }
func TestPortMatchesGoCohere_112(t *testing.T) { t.Parallel(); portMatchesRun(t, 112) }
func TestPortMatchesGoCohere_113(t *testing.T) { t.Parallel(); portMatchesRun(t, 113) }
func TestPortMatchesGoCohere_114(t *testing.T) { t.Parallel(); portMatchesRun(t, 114) }
func TestPortMatchesGoCohere_115(t *testing.T) { t.Parallel(); portMatchesRun(t, 115) }
func TestPortMatchesGoCohere_116(t *testing.T) { t.Parallel(); portMatchesRun(t, 116) }
func TestPortMatchesGoCohere_117(t *testing.T) { t.Parallel(); portMatchesRun(t, 117) }
func TestPortMatchesGoCohere_118(t *testing.T) { t.Parallel(); portMatchesRun(t, 118) }
func TestPortMatchesGoCohere_119(t *testing.T) { t.Parallel(); portMatchesRun(t, 119) }
func TestPortMatchesGoCohere_120(t *testing.T) { t.Parallel(); portMatchesRun(t, 120) }
func TestPortMatchesGoCohere_121(t *testing.T) { t.Parallel(); portMatchesRun(t, 121) }
func TestPortMatchesGoCohere_122(t *testing.T) { t.Parallel(); portMatchesRun(t, 122) }
func TestPortMatchesGoCohere_123(t *testing.T) { t.Parallel(); portMatchesRun(t, 123) }
func TestPortMatchesGoCohere_124(t *testing.T) { t.Parallel(); portMatchesRun(t, 124) }
func TestPortMatchesGoCohere_125(t *testing.T) { t.Parallel(); portMatchesRun(t, 125) }
func TestPortMatchesGoCohere_126(t *testing.T) { t.Parallel(); portMatchesRun(t, 126) }
func TestPortMatchesGoCohere_127(t *testing.T) { t.Parallel(); portMatchesRun(t, 127) }

// Each independently reported unit owns a full 90-second budget.
func portMatchesDeadline(name string) *time.Timer {
	return time.AfterFunc(90*time.Second, func() {
		fmt.Fprintf(os.Stderr, "%s exceeded its 90s deadline\n", name)
		os.Exit(124)
	})
}
