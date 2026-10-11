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
	"sync"
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
const testPortMatchesGoCohereSplitShards = 512

func portMatchesBucket(name string) int {
	h := fnv.New64a()
	_, _ = h.Write([]byte(name))
	return int(h.Sum64() % testPortMatchesGoCohereSplitShards)
}

func portMatchesPartition(t *testing.T) [][]textCase {
	t.Helper()
	allCases := sampledCorpusCases(t, 32)
	cases, ordinaryShards, dedicated := jsonPortClassOrderPartition(allCases)
	if err := jsonPortClassOrderUnion(allCases, cases, ordinaryShards, dedicated); err != nil {
		t.Fatal(err)
	}
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
	declared := make(map[string]bool)
	for _, path := range []string{"port_matches_prepared_split_test.go", "grain30_json_test.go"} {
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range file.Decls {
			if f, ok := d.(*ast.FuncDecl); ok {
				declared[f.Name.Name] = true
			}
		}
	}
	for n := 0; n < testPortMatchesGoCohereSplitShards; n++ {
		for _, name := range jsonGrain30Names(n) {
			if !declared[name] {
				t.Fatalf("missing %s", name)
			}
			delete(declared, name)
		}
	}
	for name := range declared {
		if strings.HasPrefix(name, "TestPortMatchesGoCohere_") {
			t.Fatalf("unexpected shard %s", name)
		}
	}
	t.Logf("shard union: %d cases exactly once across %d parent buckets", count, len(shards))
	return shards
}

var portMatchesShared struct {
	once                                      sync.Once
	ready                                     bool
	shards                                    [][]textCase
	oracle, entry, script, release, sanitized string
}

func TestPortMatchesGoCohereSplit_Setup(t *testing.T) {
	t.Parallel()
	portMatchesPrepare(t)
}

// Shared builds have no test deadline; the process timeout covers preparation.
func portMatchesPrepare(t *testing.T) {
	t.Helper()
	portMatchesShared.once.Do(func() { portMatchesPrepareProducts(t) })
	if !portMatchesShared.ready {
		t.Fatal("shared preparation failed")
	}
}

func portMatchesPrepareProducts(t *testing.T) {
	started := time.Now()
	defer func() { t.Logf("TestPortMatchesGoCohere (setup): %.3fs", time.Since(started).Seconds()) }()
	portMatchesShared.shards = portMatchesPartition(t)
	oracleDir := portMatchesProduct(t, "oracle")
	loweredDir := portMatchesProduct(t, "lowered")
	portMatchesShared.oracle = filepath.Join(oracleDir, "go-cohere")
	portMatchesShared.entry = filepath.Join(loweredDir, "main.ts")
	portMatchesShared.script = filepath.Join(loweredDir, "program.mjs")
	portMatchesShared.release = filepath.Join(portMatchesProduct(t, "release"), "port")
	portMatchesShared.sanitized = filepath.Join(portMatchesProduct(t, "sanitized"), "port")
	goTools, err := portMatchesJsonGoToolchain()
	if err != nil {
		t.Fatal(err)
	}
	clangTools, err := portMatchesJsonClangToolchain()
	if err != nil {
		t.Fatal(err)
	}
	// The four-digit port/upstream wrappers use the same immutable products.
	// Publish the same products for either independently selected shard family.
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
	portMatchesPrepare(t)
	portMatchesRunItems(t, ordinal, portMatchesShared.shards[ordinal], false, true)
}

func portMatchesRunItems(t *testing.T, ordinal int, items []textCase, concurrent, leak bool) {
	t.Helper()
	deadline := portMatchesDeadline(t.Name())
	defer deadline.Stop()
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
	var reference, node, release, backend run
	if concurrent {
		sides := jsonGrain30Sides(t, path)
		reference, node, release, backend = sides[0], sides[1], sides[2], sides[3]
	} else {
		reference = execute(t, nil, portMatchesShared.oracle, "--cases", path)
	}
	if reference.exitCode != 0 || len(reference.stderr) != 0 {
		t.Fatalf("Go oracle exit %d: %s", reference.exitCode, reference.stderr)
	}
	expected := string(reference.stdout)
	if len(strings.Split(strings.TrimSuffix(expected, "\n"), "\n")) != len(items) {
		t.Fatalf("Go oracle did not answer %d cases", len(items))
	}
	if !concurrent {
		node = onNode(t, portMatchesShared.entry, "--cases", path)
		release = execute(t, nil, portMatchesShared.release, "--cases", path)
	}
	compare(t, "Node", node, expected, items)
	compare(t, "release", release, expected, items)
	// Bound sanitizer inputs exactly as the original test does.
	answers := portMatchesAnswers(t, expected, len(items))
	chunks := nativeChunks(t, items, answers)
	sanitized := runNativeChunks(t, portMatchesShared.sanitized, []string{"ASAN_OPTIONS=detect_leaks=0"}, chunks, release, items, "native ASan/UBSan")
	compare(t, "native ASan/UBSan", sanitized, expected, items)
	if leak && runtime.GOOS == "linux" {
		leaked := runNativeChunks(t, portMatchesShared.sanitized, []string{"ASAN_OPTIONS=detect_leaks=1"}, chunks, release, items, "LeakSanitizer")
		compare(t, "LeakSanitizer", leaked, expected, items)
	} else if leak {
		for _, chunk := range chunks {
			report := execute(t, nil, "leaks", "--atExit", "--", portMatchesShared.release, "--cases", chunk.path)
			if report.exitCode != 0 {
				t.Fatalf("leaks: %s", report.stdout)
			}
		}
	}
	if !concurrent {
		backend = onNode(t, portMatchesShared.script, "--cases", path)
	}
	compare(t, "JavaScript backend", backend, expected, items)
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
	// Where go keeps things and the policy that picked its release say nothing about a product (the release is go
	// version above), and differ between Workshop and a runner (#t37sw0f).
	for _, name := range []string{"GOROOT", "GOTOOLDIR", "GOPATH", "GOENV", "GOCACHE", "GOMODCACHE", "GOTMPDIR", "GOBIN", "GOCACHEPROG", "GOPROXY", "GOSUMDB", "GOPRIVATE", "GONOPROXY", "GONOSUMDB", "GOINSECURE", "GOAUTH", "GOVCS", "GOTOOLCHAIN", "GOTELEMETRY", "GOTELEMETRYDIR"} {
		delete(resolved, name)
	}
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
	if _, err := exec.Command(clang, "--version").Output(); err != nil {
		return nil, err
	}
	archiver := filepath.Join(filepath.Dir(clang), "llvm-ar")
	if info, err := os.Stat(archiver); err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0111 == 0 {
		archiver, err = exec.LookPath("ar")
		if err != nil {
			return nil, err
		}
	}
	if _, err := exec.Command(archiver, "--version").Output(); err != nil {
		return nil, err
	}
	// Each tool by its report, which names what it is and not where this machine keeps it (#t37sw0f).
	return []string{buildcache.Tool(clang, "--version"), buildcache.Tool(archiver, "--version"), "GOOS=" + runtime.GOOS, "GOARCH=" + runtime.GOARCH}, nil
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
	// The overlay names this checkout's paths: it builds the product and is never part of it (#tqrqx60).
	defer os.Remove(overlay)
	ctx := context.Background()
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

func portMatchesJsonLoweredPortInputs() portMatchesJsonBuildInputs {
	files := []string{"go.mod", "go.work", "internal", "cohere/TypeScript", "cohere/TypeScript-shim", "cohere/rule_runner", "cohere/static_single_assignment", "cohere/mutation_aliasing", "stage1/cohere/json/port_matches_prepared_split_test.go"}
	for _, name := range portFiles {
		files = append(files, "stage1/cohere/json/"+name)
	}
	// It lowers in this process and runs clang, never go: its compiler is keyed by its sources (Files) and the Go
	// release, not by go env or GOFLAGS, which differ between the tree builder and a runner (#nm31pcn).
	return portMatchesJsonBuildInputs{Name: "json lowered port and backend sources", Files: files, Flags: append([]string{"load.Load(main.ts)", "lower.Lower", "native.C", "javascript.JavaScript"}, portMatchesJsonBuildEnvironment("GOOS", "GOARCH", "GOAMD64", "GOARM64", "CGO_ENABLED", "GOEXPERIMENT")...), Toolchain: []string{runtime.Version()}}
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
	flags = append(flags, portMatchesJsonBuildEnvironment("ADAMIC_NATIVE_SPLIT", "CPATH", "C_INCLUDE_PATH", "LIBRARY_PATH", "SDKROOT", "MACOSX_DEPLOYMENT_TARGET")...)
	return portMatchesJsonBuildInputs{Name: name, Files: []string{"internal", "stage1/cohere/json/port_matches_prepared_split_test.go"}, Flags: flags, Toolchain: toolchain}
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
func TestPortMatchesGoCohere_128(t *testing.T) { t.Parallel(); portMatchesRun(t, 128) }
func TestPortMatchesGoCohere_129(t *testing.T) { t.Parallel(); portMatchesRun(t, 129) }
func TestPortMatchesGoCohere_130(t *testing.T) { t.Parallel(); portMatchesRun(t, 130) }
func TestPortMatchesGoCohere_131(t *testing.T) { t.Parallel(); portMatchesRun(t, 131) }
func TestPortMatchesGoCohere_132(t *testing.T) { t.Parallel(); portMatchesRun(t, 132) }
func TestPortMatchesGoCohere_133(t *testing.T) { t.Parallel(); portMatchesRun(t, 133) }
func TestPortMatchesGoCohere_134(t *testing.T) { t.Parallel(); portMatchesRun(t, 134) }
func TestPortMatchesGoCohere_135(t *testing.T) { t.Parallel(); portMatchesRun(t, 135) }
func TestPortMatchesGoCohere_136(t *testing.T) { t.Parallel(); portMatchesRun(t, 136) }
func TestPortMatchesGoCohere_137(t *testing.T) { t.Parallel(); portMatchesRun(t, 137) }
func TestPortMatchesGoCohere_138(t *testing.T) { t.Parallel(); portMatchesRun(t, 138) }
func TestPortMatchesGoCohere_139(t *testing.T) { t.Parallel(); portMatchesRun(t, 139) }
func TestPortMatchesGoCohere_140(t *testing.T) { t.Parallel(); portMatchesRun(t, 140) }
func TestPortMatchesGoCohere_141(t *testing.T) { t.Parallel(); portMatchesRun(t, 141) }
func TestPortMatchesGoCohere_142(t *testing.T) { t.Parallel(); portMatchesRun(t, 142) }
func TestPortMatchesGoCohere_143(t *testing.T) { t.Parallel(); portMatchesRun(t, 143) }
func TestPortMatchesGoCohere_144(t *testing.T) { t.Parallel(); portMatchesRun(t, 144) }
func TestPortMatchesGoCohere_145(t *testing.T) { t.Parallel(); portMatchesRun(t, 145) }
func TestPortMatchesGoCohere_146(t *testing.T) { t.Parallel(); portMatchesRun(t, 146) }
func TestPortMatchesGoCohere_147(t *testing.T) { t.Parallel(); portMatchesRun(t, 147) }
func TestPortMatchesGoCohere_148(t *testing.T) { t.Parallel(); portMatchesRun(t, 148) }
func TestPortMatchesGoCohere_149(t *testing.T) { t.Parallel(); portMatchesRun(t, 149) }
func TestPortMatchesGoCohere_150(t *testing.T) { t.Parallel(); portMatchesRun(t, 150) }
func TestPortMatchesGoCohere_151(t *testing.T) { t.Parallel(); portMatchesRun(t, 151) }
func TestPortMatchesGoCohere_152(t *testing.T) { t.Parallel(); portMatchesRun(t, 152) }
func TestPortMatchesGoCohere_153(t *testing.T) { t.Parallel(); portMatchesRun(t, 153) }
func TestPortMatchesGoCohere_154(t *testing.T) { t.Parallel(); portMatchesRun(t, 154) }
func TestPortMatchesGoCohere_155(t *testing.T) { t.Parallel(); portMatchesRun(t, 155) }
func TestPortMatchesGoCohere_156(t *testing.T) { t.Parallel(); portMatchesRun(t, 156) }
func TestPortMatchesGoCohere_157(t *testing.T) { t.Parallel(); portMatchesRun(t, 157) }
func TestPortMatchesGoCohere_158(t *testing.T) { t.Parallel(); portMatchesRun(t, 158) }
func TestPortMatchesGoCohere_159(t *testing.T) { t.Parallel(); portMatchesRun(t, 159) }
func TestPortMatchesGoCohere_160(t *testing.T) { t.Parallel(); portMatchesRun(t, 160) }
func TestPortMatchesGoCohere_161(t *testing.T) { t.Parallel(); portMatchesRun(t, 161) }
func TestPortMatchesGoCohere_162(t *testing.T) { t.Parallel(); portMatchesRun(t, 162) }
func TestPortMatchesGoCohere_163(t *testing.T) { t.Parallel(); portMatchesRun(t, 163) }
func TestPortMatchesGoCohere_164(t *testing.T) { t.Parallel(); portMatchesRun(t, 164) }
func TestPortMatchesGoCohere_165(t *testing.T) { t.Parallel(); portMatchesRun(t, 165) }
func TestPortMatchesGoCohere_166(t *testing.T) { t.Parallel(); portMatchesRun(t, 166) }
func TestPortMatchesGoCohere_167(t *testing.T) { t.Parallel(); portMatchesRun(t, 167) }
func TestPortMatchesGoCohere_168(t *testing.T) { t.Parallel(); portMatchesRun(t, 168) }
func TestPortMatchesGoCohere_169(t *testing.T) { t.Parallel(); portMatchesRun(t, 169) }
func TestPortMatchesGoCohere_170(t *testing.T) { t.Parallel(); portMatchesRun(t, 170) }
func TestPortMatchesGoCohere_171(t *testing.T) { t.Parallel(); portMatchesRun(t, 171) }
func TestPortMatchesGoCohere_172(t *testing.T) { t.Parallel(); portMatchesRun(t, 172) }
func TestPortMatchesGoCohere_173(t *testing.T) { t.Parallel(); portMatchesRun(t, 173) }
func TestPortMatchesGoCohere_174(t *testing.T) { t.Parallel(); portMatchesRun(t, 174) }
func TestPortMatchesGoCohere_175(t *testing.T) { t.Parallel(); portMatchesRun(t, 175) }
func TestPortMatchesGoCohere_176(t *testing.T) { t.Parallel(); portMatchesRun(t, 176) }
func TestPortMatchesGoCohere_177(t *testing.T) { t.Parallel(); portMatchesRun(t, 177) }
func TestPortMatchesGoCohere_178(t *testing.T) { t.Parallel(); portMatchesRun(t, 178) }
func TestPortMatchesGoCohere_179(t *testing.T) { t.Parallel(); portMatchesRun(t, 179) }
func TestPortMatchesGoCohere_180(t *testing.T) { t.Parallel(); portMatchesRun(t, 180) }
func TestPortMatchesGoCohere_181(t *testing.T) { t.Parallel(); portMatchesRun(t, 181) }
func TestPortMatchesGoCohere_182(t *testing.T) { t.Parallel(); portMatchesRun(t, 182) }
func TestPortMatchesGoCohere_183(t *testing.T) { t.Parallel(); portMatchesRun(t, 183) }
func TestPortMatchesGoCohere_184(t *testing.T) { t.Parallel(); portMatchesRun(t, 184) }
func TestPortMatchesGoCohere_185(t *testing.T) { t.Parallel(); portMatchesRun(t, 185) }
func TestPortMatchesGoCohere_186(t *testing.T) { t.Parallel(); portMatchesRun(t, 186) }
func TestPortMatchesGoCohere_187(t *testing.T) { t.Parallel(); portMatchesRun(t, 187) }
func TestPortMatchesGoCohere_188(t *testing.T) { t.Parallel(); portMatchesRun(t, 188) }
func TestPortMatchesGoCohere_189(t *testing.T) { t.Parallel(); portMatchesRun(t, 189) }
func TestPortMatchesGoCohere_190(t *testing.T) { t.Parallel(); portMatchesRun(t, 190) }
func TestPortMatchesGoCohere_191(t *testing.T) { t.Parallel(); portMatchesRun(t, 191) }
func TestPortMatchesGoCohere_192(t *testing.T) { t.Parallel(); portMatchesRun(t, 192) }
func TestPortMatchesGoCohere_193(t *testing.T) { t.Parallel(); portMatchesRun(t, 193) }
func TestPortMatchesGoCohere_194(t *testing.T) { t.Parallel(); portMatchesRun(t, 194) }
func TestPortMatchesGoCohere_195(t *testing.T) { t.Parallel(); portMatchesRun(t, 195) }
func TestPortMatchesGoCohere_196(t *testing.T) { t.Parallel(); portMatchesRun(t, 196) }
func TestPortMatchesGoCohere_197(t *testing.T) { t.Parallel(); portMatchesRun(t, 197) }
func TestPortMatchesGoCohere_198(t *testing.T) { t.Parallel(); portMatchesRun(t, 198) }
func TestPortMatchesGoCohere_199(t *testing.T) { t.Parallel(); portMatchesRun(t, 199) }
func TestPortMatchesGoCohere_200(t *testing.T) { t.Parallel(); portMatchesRun(t, 200) }
func TestPortMatchesGoCohere_201(t *testing.T) { t.Parallel(); portMatchesRun(t, 201) }
func TestPortMatchesGoCohere_202(t *testing.T) { t.Parallel(); portMatchesRun(t, 202) }
func TestPortMatchesGoCohere_203(t *testing.T) { t.Parallel(); portMatchesRun(t, 203) }
func TestPortMatchesGoCohere_204(t *testing.T) { t.Parallel(); portMatchesRun(t, 204) }
func TestPortMatchesGoCohere_205(t *testing.T) { t.Parallel(); portMatchesRun(t, 205) }
func TestPortMatchesGoCohere_206(t *testing.T) { t.Parallel(); portMatchesRun(t, 206) }
func TestPortMatchesGoCohere_207(t *testing.T) { t.Parallel(); portMatchesRun(t, 207) }
func TestPortMatchesGoCohere_208(t *testing.T) { t.Parallel(); portMatchesRun(t, 208) }
func TestPortMatchesGoCohere_209(t *testing.T) { t.Parallel(); portMatchesRun(t, 209) }
func TestPortMatchesGoCohere_210(t *testing.T) { t.Parallel(); portMatchesRun(t, 210) }
func TestPortMatchesGoCohere_211(t *testing.T) { t.Parallel(); portMatchesRun(t, 211) }
func TestPortMatchesGoCohere_212(t *testing.T) { t.Parallel(); portMatchesRun(t, 212) }
func TestPortMatchesGoCohere_213(t *testing.T) { t.Parallel(); portMatchesRun(t, 213) }
func TestPortMatchesGoCohere_214(t *testing.T) { t.Parallel(); portMatchesRun(t, 214) }
func TestPortMatchesGoCohere_215(t *testing.T) { t.Parallel(); portMatchesRun(t, 215) }
func TestPortMatchesGoCohere_216(t *testing.T) { t.Parallel(); portMatchesRun(t, 216) }
func TestPortMatchesGoCohere_217(t *testing.T) { t.Parallel(); portMatchesRun(t, 217) }
func TestPortMatchesGoCohere_218(t *testing.T) { t.Parallel(); portMatchesRun(t, 218) }
func TestPortMatchesGoCohere_219(t *testing.T) { t.Parallel(); portMatchesRun(t, 219) }
func TestPortMatchesGoCohere_220(t *testing.T) { t.Parallel(); portMatchesRun(t, 220) }
func TestPortMatchesGoCohere_221(t *testing.T) { t.Parallel(); portMatchesRun(t, 221) }
func TestPortMatchesGoCohere_222(t *testing.T) { t.Parallel(); portMatchesRun(t, 222) }
func TestPortMatchesGoCohere_223(t *testing.T) { t.Parallel(); portMatchesRun(t, 223) }
func TestPortMatchesGoCohere_224(t *testing.T) { t.Parallel(); portMatchesRun(t, 224) }
func TestPortMatchesGoCohere_225(t *testing.T) { t.Parallel(); portMatchesRun(t, 225) }
func TestPortMatchesGoCohere_226(t *testing.T) { t.Parallel(); portMatchesRun(t, 226) }
func TestPortMatchesGoCohere_227(t *testing.T) { t.Parallel(); portMatchesRun(t, 227) }
func TestPortMatchesGoCohere_228(t *testing.T) { t.Parallel(); portMatchesRun(t, 228) }
func TestPortMatchesGoCohere_229(t *testing.T) { t.Parallel(); portMatchesRun(t, 229) }
func TestPortMatchesGoCohere_230(t *testing.T) { t.Parallel(); portMatchesRun(t, 230) }
func TestPortMatchesGoCohere_231(t *testing.T) { t.Parallel(); portMatchesRun(t, 231) }
func TestPortMatchesGoCohere_232(t *testing.T) { t.Parallel(); portMatchesRun(t, 232) }
func TestPortMatchesGoCohere_233(t *testing.T) { t.Parallel(); portMatchesRun(t, 233) }
func TestPortMatchesGoCohere_234(t *testing.T) { t.Parallel(); portMatchesRun(t, 234) }
func TestPortMatchesGoCohere_235(t *testing.T) { t.Parallel(); portMatchesRun(t, 235) }
func TestPortMatchesGoCohere_236(t *testing.T) { t.Parallel(); portMatchesRun(t, 236) }
func TestPortMatchesGoCohere_237(t *testing.T) { t.Parallel(); portMatchesRun(t, 237) }
func TestPortMatchesGoCohere_238(t *testing.T) { t.Parallel(); portMatchesRun(t, 238) }
func TestPortMatchesGoCohere_239(t *testing.T) { t.Parallel(); portMatchesRun(t, 239) }
func TestPortMatchesGoCohere_240(t *testing.T) { t.Parallel(); portMatchesRun(t, 240) }
func TestPortMatchesGoCohere_241(t *testing.T) { t.Parallel(); portMatchesRun(t, 241) }
func TestPortMatchesGoCohere_242(t *testing.T) { t.Parallel(); portMatchesRun(t, 242) }
func TestPortMatchesGoCohere_243(t *testing.T) { t.Parallel(); portMatchesRun(t, 243) }
func TestPortMatchesGoCohere_244(t *testing.T) { t.Parallel(); portMatchesRun(t, 244) }
func TestPortMatchesGoCohere_245(t *testing.T) { t.Parallel(); portMatchesRun(t, 245) }
func TestPortMatchesGoCohere_246(t *testing.T) { t.Parallel(); portMatchesRun(t, 246) }
func TestPortMatchesGoCohere_247(t *testing.T) { t.Parallel(); portMatchesRun(t, 247) }
func TestPortMatchesGoCohere_248(t *testing.T) { t.Parallel(); portMatchesRun(t, 248) }
func TestPortMatchesGoCohere_249(t *testing.T) { t.Parallel(); portMatchesRun(t, 249) }
func TestPortMatchesGoCohere_250(t *testing.T) { t.Parallel(); portMatchesRun(t, 250) }
func TestPortMatchesGoCohere_251(t *testing.T) { t.Parallel(); portMatchesRun(t, 251) }
func TestPortMatchesGoCohere_252(t *testing.T) { t.Parallel(); portMatchesRun(t, 252) }
func TestPortMatchesGoCohere_253(t *testing.T) { t.Parallel(); portMatchesRun(t, 253) }
func TestPortMatchesGoCohere_254(t *testing.T) { t.Parallel(); portMatchesRun(t, 254) }
func TestPortMatchesGoCohere_255(t *testing.T) { t.Parallel(); portMatchesRun(t, 255) }
func TestPortMatchesGoCohere_256(t *testing.T) { t.Parallel(); portMatchesRun(t, 256) }
func TestPortMatchesGoCohere_257(t *testing.T) { t.Parallel(); portMatchesRun(t, 257) }
func TestPortMatchesGoCohere_258(t *testing.T) { t.Parallel(); portMatchesRun(t, 258) }
func TestPortMatchesGoCohere_259(t *testing.T) { t.Parallel(); portMatchesRun(t, 259) }
func TestPortMatchesGoCohere_260(t *testing.T) { t.Parallel(); portMatchesRun(t, 260) }
func TestPortMatchesGoCohere_261(t *testing.T) { t.Parallel(); portMatchesRun(t, 261) }
func TestPortMatchesGoCohere_262(t *testing.T) { t.Parallel(); portMatchesRun(t, 262) }
func TestPortMatchesGoCohere_263(t *testing.T) { t.Parallel(); portMatchesRun(t, 263) }
func TestPortMatchesGoCohere_264(t *testing.T) { t.Parallel(); portMatchesRun(t, 264) }
func TestPortMatchesGoCohere_265(t *testing.T) { t.Parallel(); portMatchesRun(t, 265) }
func TestPortMatchesGoCohere_266(t *testing.T) { t.Parallel(); portMatchesRun(t, 266) }
func TestPortMatchesGoCohere_267(t *testing.T) { t.Parallel(); portMatchesRun(t, 267) }
func TestPortMatchesGoCohere_268(t *testing.T) { t.Parallel(); portMatchesRun(t, 268) }
func TestPortMatchesGoCohere_269(t *testing.T) { t.Parallel(); portMatchesRun(t, 269) }
func TestPortMatchesGoCohere_270(t *testing.T) { t.Parallel(); portMatchesRun(t, 270) }
func TestPortMatchesGoCohere_271(t *testing.T) { t.Parallel(); portMatchesRun(t, 271) }
func TestPortMatchesGoCohere_272(t *testing.T) { t.Parallel(); portMatchesRun(t, 272) }
func TestPortMatchesGoCohere_273(t *testing.T) { t.Parallel(); portMatchesRun(t, 273) }
func TestPortMatchesGoCohere_274(t *testing.T) { t.Parallel(); portMatchesRun(t, 274) }
func TestPortMatchesGoCohere_275(t *testing.T) { t.Parallel(); portMatchesRun(t, 275) }
func TestPortMatchesGoCohere_276(t *testing.T) { t.Parallel(); portMatchesRun(t, 276) }
func TestPortMatchesGoCohere_277(t *testing.T) { t.Parallel(); portMatchesRun(t, 277) }
func TestPortMatchesGoCohere_278(t *testing.T) { t.Parallel(); portMatchesRun(t, 278) }
func TestPortMatchesGoCohere_279(t *testing.T) { t.Parallel(); portMatchesRun(t, 279) }
func TestPortMatchesGoCohere_280(t *testing.T) { t.Parallel(); portMatchesRun(t, 280) }
func TestPortMatchesGoCohere_281(t *testing.T) { t.Parallel(); portMatchesRun(t, 281) }
func TestPortMatchesGoCohere_282(t *testing.T) { t.Parallel(); portMatchesRun(t, 282) }
func TestPortMatchesGoCohere_283(t *testing.T) { t.Parallel(); portMatchesRun(t, 283) }
func TestPortMatchesGoCohere_284(t *testing.T) { t.Parallel(); portMatchesRun(t, 284) }
func TestPortMatchesGoCohere_285(t *testing.T) { t.Parallel(); portMatchesRun(t, 285) }
func TestPortMatchesGoCohere_286(t *testing.T) { t.Parallel(); portMatchesRun(t, 286) }
func TestPortMatchesGoCohere_287(t *testing.T) { t.Parallel(); portMatchesRun(t, 287) }
func TestPortMatchesGoCohere_288(t *testing.T) { t.Parallel(); portMatchesRun(t, 288) }
func TestPortMatchesGoCohere_289(t *testing.T) { t.Parallel(); portMatchesRun(t, 289) }
func TestPortMatchesGoCohere_290(t *testing.T) { t.Parallel(); portMatchesRun(t, 290) }
func TestPortMatchesGoCohere_291(t *testing.T) { t.Parallel(); portMatchesRun(t, 291) }
func TestPortMatchesGoCohere_292(t *testing.T) { t.Parallel(); portMatchesRun(t, 292) }
func TestPortMatchesGoCohere_293(t *testing.T) { t.Parallel(); portMatchesRun(t, 293) }
func TestPortMatchesGoCohere_294(t *testing.T) { t.Parallel(); portMatchesRun(t, 294) }
func TestPortMatchesGoCohere_295(t *testing.T) { t.Parallel(); portMatchesRun(t, 295) }
func TestPortMatchesGoCohere_296(t *testing.T) { t.Parallel(); portMatchesRun(t, 296) }
func TestPortMatchesGoCohere_297(t *testing.T) { t.Parallel(); portMatchesRun(t, 297) }
func TestPortMatchesGoCohere_298(t *testing.T) { t.Parallel(); portMatchesRun(t, 298) }
func TestPortMatchesGoCohere_299(t *testing.T) { t.Parallel(); portMatchesRun(t, 299) }
func TestPortMatchesGoCohere_300(t *testing.T) { t.Parallel(); portMatchesRun(t, 300) }
func TestPortMatchesGoCohere_301(t *testing.T) { t.Parallel(); portMatchesRun(t, 301) }
func TestPortMatchesGoCohere_302(t *testing.T) { t.Parallel(); portMatchesRun(t, 302) }
func TestPortMatchesGoCohere_303(t *testing.T) { t.Parallel(); portMatchesRun(t, 303) }
func TestPortMatchesGoCohere_304(t *testing.T) { t.Parallel(); portMatchesRun(t, 304) }
func TestPortMatchesGoCohere_305(t *testing.T) { t.Parallel(); portMatchesRun(t, 305) }
func TestPortMatchesGoCohere_306(t *testing.T) { t.Parallel(); portMatchesRun(t, 306) }
func TestPortMatchesGoCohere_307(t *testing.T) { t.Parallel(); portMatchesRun(t, 307) }
func TestPortMatchesGoCohere_308(t *testing.T) { t.Parallel(); portMatchesRun(t, 308) }
func TestPortMatchesGoCohere_309(t *testing.T) { t.Parallel(); portMatchesRun(t, 309) }
func TestPortMatchesGoCohere_310(t *testing.T) { t.Parallel(); portMatchesRun(t, 310) }
func TestPortMatchesGoCohere_311(t *testing.T) { t.Parallel(); portMatchesRun(t, 311) }
func TestPortMatchesGoCohere_312(t *testing.T) { t.Parallel(); portMatchesRun(t, 312) }
func TestPortMatchesGoCohere_313(t *testing.T) { t.Parallel(); portMatchesRun(t, 313) }
func TestPortMatchesGoCohere_314(t *testing.T) { t.Parallel(); portMatchesRun(t, 314) }
func TestPortMatchesGoCohere_315(t *testing.T) { t.Parallel(); portMatchesRun(t, 315) }
func TestPortMatchesGoCohere_316(t *testing.T) { t.Parallel(); portMatchesRun(t, 316) }
func TestPortMatchesGoCohere_317(t *testing.T) { t.Parallel(); portMatchesRun(t, 317) }
func TestPortMatchesGoCohere_318(t *testing.T) { t.Parallel(); portMatchesRun(t, 318) }
func TestPortMatchesGoCohere_319(t *testing.T) { t.Parallel(); portMatchesRun(t, 319) }
func TestPortMatchesGoCohere_320(t *testing.T) { t.Parallel(); portMatchesRun(t, 320) }
func TestPortMatchesGoCohere_321(t *testing.T) { t.Parallel(); portMatchesRun(t, 321) }
func TestPortMatchesGoCohere_322(t *testing.T) { t.Parallel(); portMatchesRun(t, 322) }
func TestPortMatchesGoCohere_323(t *testing.T) { t.Parallel(); portMatchesRun(t, 323) }
func TestPortMatchesGoCohere_324(t *testing.T) { t.Parallel(); portMatchesRun(t, 324) }
func TestPortMatchesGoCohere_325(t *testing.T) { t.Parallel(); portMatchesRun(t, 325) }
func TestPortMatchesGoCohere_326(t *testing.T) { t.Parallel(); portMatchesRun(t, 326) }
func TestPortMatchesGoCohere_327(t *testing.T) { t.Parallel(); portMatchesRun(t, 327) }
func TestPortMatchesGoCohere_328(t *testing.T) { t.Parallel(); portMatchesRun(t, 328) }
func TestPortMatchesGoCohere_329(t *testing.T) { t.Parallel(); portMatchesRun(t, 329) }
func TestPortMatchesGoCohere_330(t *testing.T) { t.Parallel(); portMatchesRun(t, 330) }
func TestPortMatchesGoCohere_331(t *testing.T) { t.Parallel(); portMatchesRun(t, 331) }
func TestPortMatchesGoCohere_332(t *testing.T) { t.Parallel(); portMatchesRun(t, 332) }
func TestPortMatchesGoCohere_333(t *testing.T) { t.Parallel(); portMatchesRun(t, 333) }
func TestPortMatchesGoCohere_334(t *testing.T) { t.Parallel(); portMatchesRun(t, 334) }
func TestPortMatchesGoCohere_335(t *testing.T) { t.Parallel(); portMatchesRun(t, 335) }
func TestPortMatchesGoCohere_336(t *testing.T) { t.Parallel(); portMatchesRun(t, 336) }
func TestPortMatchesGoCohere_337(t *testing.T) { t.Parallel(); portMatchesRun(t, 337) }
func TestPortMatchesGoCohere_338(t *testing.T) { t.Parallel(); portMatchesRun(t, 338) }
func TestPortMatchesGoCohere_339(t *testing.T) { t.Parallel(); portMatchesRun(t, 339) }
func TestPortMatchesGoCohere_340(t *testing.T) { t.Parallel(); portMatchesRun(t, 340) }
func TestPortMatchesGoCohere_341(t *testing.T) { t.Parallel(); portMatchesRun(t, 341) }
func TestPortMatchesGoCohere_342(t *testing.T) { t.Parallel(); portMatchesRun(t, 342) }
func TestPortMatchesGoCohere_343(t *testing.T) { t.Parallel(); portMatchesRun(t, 343) }
func TestPortMatchesGoCohere_344(t *testing.T) { t.Parallel(); portMatchesRun(t, 344) }
func TestPortMatchesGoCohere_345(t *testing.T) { t.Parallel(); portMatchesRun(t, 345) }
func TestPortMatchesGoCohere_346(t *testing.T) { t.Parallel(); portMatchesRun(t, 346) }
func TestPortMatchesGoCohere_347(t *testing.T) { t.Parallel(); portMatchesRun(t, 347) }
func TestPortMatchesGoCohere_348(t *testing.T) { t.Parallel(); portMatchesRun(t, 348) }
func TestPortMatchesGoCohere_349(t *testing.T) { t.Parallel(); portMatchesRun(t, 349) }
func TestPortMatchesGoCohere_350(t *testing.T) { t.Parallel(); portMatchesRun(t, 350) }
func TestPortMatchesGoCohere_351(t *testing.T) { t.Parallel(); portMatchesRun(t, 351) }
func TestPortMatchesGoCohere_352(t *testing.T) { t.Parallel(); portMatchesRun(t, 352) }
func TestPortMatchesGoCohere_353(t *testing.T) { t.Parallel(); portMatchesRun(t, 353) }
func TestPortMatchesGoCohere_354(t *testing.T) { t.Parallel(); portMatchesRun(t, 354) }
func TestPortMatchesGoCohere_355(t *testing.T) { t.Parallel(); portMatchesRun(t, 355) }
func TestPortMatchesGoCohere_356(t *testing.T) { t.Parallel(); portMatchesRun(t, 356) }
func TestPortMatchesGoCohere_357(t *testing.T) { t.Parallel(); portMatchesRun(t, 357) }
func TestPortMatchesGoCohere_358(t *testing.T) { t.Parallel(); portMatchesRun(t, 358) }
func TestPortMatchesGoCohere_359(t *testing.T) { t.Parallel(); portMatchesRun(t, 359) }
func TestPortMatchesGoCohere_360(t *testing.T) { t.Parallel(); portMatchesRun(t, 360) }
func TestPortMatchesGoCohere_361(t *testing.T) { t.Parallel(); portMatchesRun(t, 361) }
func TestPortMatchesGoCohere_362(t *testing.T) { t.Parallel(); portMatchesRun(t, 362) }
func TestPortMatchesGoCohere_363(t *testing.T) { t.Parallel(); portMatchesRun(t, 363) }
func TestPortMatchesGoCohere_364(t *testing.T) { t.Parallel(); portMatchesRun(t, 364) }
func TestPortMatchesGoCohere_365(t *testing.T) { t.Parallel(); portMatchesRun(t, 365) }
func TestPortMatchesGoCohere_366(t *testing.T) { t.Parallel(); portMatchesRun(t, 366) }
func TestPortMatchesGoCohere_367(t *testing.T) { t.Parallel(); portMatchesRun(t, 367) }
func TestPortMatchesGoCohere_368(t *testing.T) { t.Parallel(); portMatchesRun(t, 368) }
func TestPortMatchesGoCohere_369(t *testing.T) { t.Parallel(); portMatchesRun(t, 369) }
func TestPortMatchesGoCohere_370(t *testing.T) { t.Parallel(); portMatchesRun(t, 370) }
func TestPortMatchesGoCohere_371(t *testing.T) { t.Parallel(); portMatchesRun(t, 371) }
func TestPortMatchesGoCohere_372(t *testing.T) { t.Parallel(); portMatchesRun(t, 372) }
func TestPortMatchesGoCohere_373(t *testing.T) { t.Parallel(); portMatchesRun(t, 373) }
func TestPortMatchesGoCohere_374(t *testing.T) { t.Parallel(); portMatchesRun(t, 374) }
func TestPortMatchesGoCohere_375(t *testing.T) { t.Parallel(); portMatchesRun(t, 375) }
func TestPortMatchesGoCohere_376(t *testing.T) { t.Parallel(); portMatchesRun(t, 376) }
func TestPortMatchesGoCohere_377(t *testing.T) { t.Parallel(); portMatchesRun(t, 377) }
func TestPortMatchesGoCohere_378(t *testing.T) { t.Parallel(); portMatchesRun(t, 378) }
func TestPortMatchesGoCohere_379(t *testing.T) { t.Parallel(); portMatchesRun(t, 379) }
func TestPortMatchesGoCohere_380(t *testing.T) { t.Parallel(); portMatchesRun(t, 380) }
func TestPortMatchesGoCohere_381(t *testing.T) { t.Parallel(); portMatchesRun(t, 381) }
func TestPortMatchesGoCohere_382(t *testing.T) { t.Parallel(); portMatchesRun(t, 382) }
func TestPortMatchesGoCohere_383(t *testing.T) { t.Parallel(); portMatchesRun(t, 383) }
func TestPortMatchesGoCohere_384(t *testing.T) { t.Parallel(); portMatchesRun(t, 384) }
func TestPortMatchesGoCohere_385(t *testing.T) { t.Parallel(); portMatchesRun(t, 385) }
func TestPortMatchesGoCohere_386(t *testing.T) { t.Parallel(); portMatchesRun(t, 386) }
func TestPortMatchesGoCohere_387(t *testing.T) { t.Parallel(); portMatchesRun(t, 387) }
func TestPortMatchesGoCohere_388(t *testing.T) { t.Parallel(); portMatchesRun(t, 388) }
func TestPortMatchesGoCohere_389(t *testing.T) { t.Parallel(); portMatchesRun(t, 389) }
func TestPortMatchesGoCohere_390(t *testing.T) { t.Parallel(); portMatchesRun(t, 390) }
func TestPortMatchesGoCohere_391(t *testing.T) { t.Parallel(); portMatchesRun(t, 391) }
func TestPortMatchesGoCohere_392(t *testing.T) { t.Parallel(); portMatchesRun(t, 392) }
func TestPortMatchesGoCohere_393(t *testing.T) { t.Parallel(); portMatchesRun(t, 393) }
func TestPortMatchesGoCohere_394(t *testing.T) { t.Parallel(); portMatchesRun(t, 394) }
func TestPortMatchesGoCohere_395(t *testing.T) { t.Parallel(); portMatchesRun(t, 395) }
func TestPortMatchesGoCohere_396(t *testing.T) { t.Parallel(); portMatchesRun(t, 396) }
func TestPortMatchesGoCohere_397(t *testing.T) { t.Parallel(); portMatchesRun(t, 397) }
func TestPortMatchesGoCohere_398(t *testing.T) { t.Parallel(); portMatchesRun(t, 398) }
func TestPortMatchesGoCohere_399(t *testing.T) { t.Parallel(); portMatchesRun(t, 399) }
func TestPortMatchesGoCohere_400(t *testing.T) { t.Parallel(); portMatchesRun(t, 400) }
func TestPortMatchesGoCohere_401(t *testing.T) { t.Parallel(); portMatchesRun(t, 401) }
func TestPortMatchesGoCohere_402(t *testing.T) { t.Parallel(); portMatchesRun(t, 402) }
func TestPortMatchesGoCohere_403(t *testing.T) { t.Parallel(); portMatchesRun(t, 403) }
func TestPortMatchesGoCohere_404(t *testing.T) { t.Parallel(); portMatchesRun(t, 404) }
func TestPortMatchesGoCohere_405(t *testing.T) { t.Parallel(); portMatchesRun(t, 405) }
func TestPortMatchesGoCohere_406(t *testing.T) { t.Parallel(); portMatchesRun(t, 406) }
func TestPortMatchesGoCohere_407(t *testing.T) { t.Parallel(); portMatchesRun(t, 407) }
func TestPortMatchesGoCohere_408(t *testing.T) { t.Parallel(); portMatchesRun(t, 408) }
func TestPortMatchesGoCohere_409(t *testing.T) { t.Parallel(); portMatchesRun(t, 409) }
func TestPortMatchesGoCohere_410(t *testing.T) { t.Parallel(); portMatchesRun(t, 410) }
func TestPortMatchesGoCohere_411(t *testing.T) { t.Parallel(); portMatchesRun(t, 411) }
func TestPortMatchesGoCohere_412(t *testing.T) { t.Parallel(); portMatchesRun(t, 412) }
func TestPortMatchesGoCohere_413(t *testing.T) { t.Parallel(); portMatchesRun(t, 413) }
func TestPortMatchesGoCohere_414(t *testing.T) { t.Parallel(); portMatchesRun(t, 414) }
func TestPortMatchesGoCohere_415(t *testing.T) { t.Parallel(); portMatchesRun(t, 415) }
func TestPortMatchesGoCohere_416(t *testing.T) { t.Parallel(); portMatchesRun(t, 416) }
func TestPortMatchesGoCohere_417(t *testing.T) { t.Parallel(); portMatchesRun(t, 417) }
func TestPortMatchesGoCohere_418(t *testing.T) { t.Parallel(); portMatchesRun(t, 418) }
func TestPortMatchesGoCohere_419(t *testing.T) { t.Parallel(); portMatchesRun(t, 419) }
func TestPortMatchesGoCohere_420(t *testing.T) { t.Parallel(); portMatchesRun(t, 420) }
func TestPortMatchesGoCohere_421(t *testing.T) { t.Parallel(); portMatchesRun(t, 421) }
func TestPortMatchesGoCohere_422(t *testing.T) { t.Parallel(); portMatchesRun(t, 422) }
func TestPortMatchesGoCohere_423(t *testing.T) { t.Parallel(); portMatchesRun(t, 423) }
func TestPortMatchesGoCohere_424(t *testing.T) { t.Parallel(); portMatchesRun(t, 424) }
func TestPortMatchesGoCohere_425(t *testing.T) { t.Parallel(); portMatchesRun(t, 425) }
func TestPortMatchesGoCohere_426(t *testing.T) { t.Parallel(); portMatchesRun(t, 426) }
func TestPortMatchesGoCohere_427(t *testing.T) { t.Parallel(); portMatchesRun(t, 427) }
func TestPortMatchesGoCohere_428(t *testing.T) { t.Parallel(); portMatchesRun(t, 428) }
func TestPortMatchesGoCohere_429(t *testing.T) { t.Parallel(); portMatchesRun(t, 429) }
func TestPortMatchesGoCohere_430(t *testing.T) { t.Parallel(); portMatchesRun(t, 430) }
func TestPortMatchesGoCohere_431(t *testing.T) { t.Parallel(); portMatchesRun(t, 431) }
func TestPortMatchesGoCohere_432(t *testing.T) { t.Parallel(); portMatchesRun(t, 432) }
func TestPortMatchesGoCohere_433(t *testing.T) { t.Parallel(); portMatchesRun(t, 433) }
func TestPortMatchesGoCohere_434(t *testing.T) { t.Parallel(); portMatchesRun(t, 434) }
func TestPortMatchesGoCohere_435(t *testing.T) { t.Parallel(); portMatchesRun(t, 435) }
func TestPortMatchesGoCohere_436(t *testing.T) { t.Parallel(); portMatchesRun(t, 436) }
func TestPortMatchesGoCohere_437(t *testing.T) { t.Parallel(); portMatchesRun(t, 437) }
func TestPortMatchesGoCohere_438(t *testing.T) { t.Parallel(); portMatchesRun(t, 438) }
func TestPortMatchesGoCohere_439(t *testing.T) { t.Parallel(); portMatchesRun(t, 439) }
func TestPortMatchesGoCohere_440(t *testing.T) { t.Parallel(); portMatchesRun(t, 440) }
func TestPortMatchesGoCohere_441(t *testing.T) { t.Parallel(); portMatchesRun(t, 441) }
func TestPortMatchesGoCohere_442(t *testing.T) { t.Parallel(); portMatchesRun(t, 442) }
func TestPortMatchesGoCohere_443(t *testing.T) { t.Parallel(); portMatchesRun(t, 443) }
func TestPortMatchesGoCohere_444(t *testing.T) { t.Parallel(); portMatchesRun(t, 444) }
func TestPortMatchesGoCohere_445(t *testing.T) { t.Parallel(); portMatchesRun(t, 445) }
func TestPortMatchesGoCohere_446(t *testing.T) { t.Parallel(); portMatchesRun(t, 446) }
func TestPortMatchesGoCohere_447(t *testing.T) { t.Parallel(); portMatchesRun(t, 447) }
func TestPortMatchesGoCohere_448(t *testing.T) { t.Parallel(); portMatchesRun(t, 448) }
func TestPortMatchesGoCohere_449(t *testing.T) { t.Parallel(); portMatchesRun(t, 449) }
func TestPortMatchesGoCohere_450(t *testing.T) { t.Parallel(); portMatchesRun(t, 450) }
func TestPortMatchesGoCohere_451(t *testing.T) { t.Parallel(); portMatchesRun(t, 451) }
func TestPortMatchesGoCohere_452(t *testing.T) { t.Parallel(); portMatchesRun(t, 452) }
func TestPortMatchesGoCohere_453(t *testing.T) { t.Parallel(); portMatchesRun(t, 453) }
func TestPortMatchesGoCohere_454(t *testing.T) { t.Parallel(); portMatchesRun(t, 454) }
func TestPortMatchesGoCohere_455(t *testing.T) { t.Parallel(); portMatchesRun(t, 455) }
func TestPortMatchesGoCohere_456(t *testing.T) { t.Parallel(); portMatchesRun(t, 456) }
func TestPortMatchesGoCohere_457(t *testing.T) { t.Parallel(); portMatchesRun(t, 457) }
func TestPortMatchesGoCohere_458(t *testing.T) { t.Parallel(); portMatchesRun(t, 458) }
func TestPortMatchesGoCohere_459(t *testing.T) { t.Parallel(); portMatchesRun(t, 459) }
func TestPortMatchesGoCohere_460(t *testing.T) { t.Parallel(); portMatchesRun(t, 460) }
func TestPortMatchesGoCohere_461(t *testing.T) { t.Parallel(); portMatchesRun(t, 461) }
func TestPortMatchesGoCohere_462(t *testing.T) { t.Parallel(); portMatchesRun(t, 462) }
func TestPortMatchesGoCohere_463(t *testing.T) { t.Parallel(); portMatchesRun(t, 463) }
func TestPortMatchesGoCohere_464(t *testing.T) { t.Parallel(); portMatchesRun(t, 464) }
func TestPortMatchesGoCohere_465(t *testing.T) { t.Parallel(); portMatchesRun(t, 465) }
func TestPortMatchesGoCohere_466(t *testing.T) { t.Parallel(); portMatchesRun(t, 466) }
func TestPortMatchesGoCohere_467(t *testing.T) { t.Parallel(); portMatchesRun(t, 467) }
func TestPortMatchesGoCohere_468(t *testing.T) { t.Parallel(); portMatchesRun(t, 468) }
func TestPortMatchesGoCohere_469(t *testing.T) { t.Parallel(); portMatchesRun(t, 469) }
func TestPortMatchesGoCohere_470(t *testing.T) { t.Parallel(); portMatchesRun(t, 470) }
func TestPortMatchesGoCohere_471(t *testing.T) { t.Parallel(); portMatchesRun(t, 471) }
func TestPortMatchesGoCohere_472(t *testing.T) { t.Parallel(); portMatchesRun(t, 472) }
func TestPortMatchesGoCohere_473(t *testing.T) { t.Parallel(); portMatchesRun(t, 473) }
func TestPortMatchesGoCohere_474(t *testing.T) { t.Parallel(); portMatchesRun(t, 474) }
func TestPortMatchesGoCohere_475(t *testing.T) { t.Parallel(); portMatchesRun(t, 475) }
func TestPortMatchesGoCohere_476(t *testing.T) { t.Parallel(); portMatchesRun(t, 476) }
func TestPortMatchesGoCohere_477(t *testing.T) { t.Parallel(); portMatchesRun(t, 477) }
func TestPortMatchesGoCohere_478(t *testing.T) { t.Parallel(); portMatchesRun(t, 478) }
func TestPortMatchesGoCohere_479(t *testing.T) { t.Parallel(); portMatchesRun(t, 479) }
func TestPortMatchesGoCohere_480(t *testing.T) { t.Parallel(); portMatchesRun(t, 480) }
func TestPortMatchesGoCohere_481(t *testing.T) { t.Parallel(); portMatchesRun(t, 481) }
func TestPortMatchesGoCohere_482(t *testing.T) { t.Parallel(); portMatchesRun(t, 482) }
func TestPortMatchesGoCohere_483(t *testing.T) { t.Parallel(); portMatchesRun(t, 483) }
func TestPortMatchesGoCohere_484(t *testing.T) { t.Parallel(); portMatchesRun(t, 484) }
func TestPortMatchesGoCohere_485(t *testing.T) { t.Parallel(); portMatchesRun(t, 485) }
func TestPortMatchesGoCohere_486(t *testing.T) { t.Parallel(); portMatchesRun(t, 486) }
func TestPortMatchesGoCohere_487(t *testing.T) { t.Parallel(); portMatchesRun(t, 487) }
func TestPortMatchesGoCohere_488(t *testing.T) { t.Parallel(); portMatchesRun(t, 488) }
func TestPortMatchesGoCohere_489(t *testing.T) { t.Parallel(); portMatchesRun(t, 489) }
func TestPortMatchesGoCohere_490(t *testing.T) { t.Parallel(); portMatchesRun(t, 490) }
func TestPortMatchesGoCohere_491(t *testing.T) { t.Parallel(); portMatchesRun(t, 491) }
func TestPortMatchesGoCohere_492(t *testing.T) { t.Parallel(); portMatchesRun(t, 492) }
func TestPortMatchesGoCohere_493(t *testing.T) { t.Parallel(); portMatchesRun(t, 493) }
func TestPortMatchesGoCohere_494(t *testing.T) { t.Parallel(); portMatchesRun(t, 494) }
func TestPortMatchesGoCohere_495(t *testing.T) { t.Parallel(); portMatchesRun(t, 495) }
func TestPortMatchesGoCohere_496(t *testing.T) { t.Parallel(); portMatchesRun(t, 496) }
func TestPortMatchesGoCohere_497(t *testing.T) { t.Parallel(); portMatchesRun(t, 497) }
func TestPortMatchesGoCohere_498(t *testing.T) { t.Parallel(); portMatchesRun(t, 498) }
func TestPortMatchesGoCohere_499(t *testing.T) { t.Parallel(); portMatchesRun(t, 499) }
func TestPortMatchesGoCohere_500(t *testing.T) { t.Parallel(); portMatchesRun(t, 500) }
func TestPortMatchesGoCohere_501(t *testing.T) { t.Parallel(); portMatchesRun(t, 501) }
func TestPortMatchesGoCohere_502(t *testing.T) { t.Parallel(); portMatchesRun(t, 502) }
func TestPortMatchesGoCohere_503(t *testing.T) { t.Parallel(); portMatchesRun(t, 503) }
func TestPortMatchesGoCohere_504(t *testing.T) { t.Parallel(); portMatchesRun(t, 504) }
func TestPortMatchesGoCohere_505(t *testing.T) { t.Parallel(); portMatchesRun(t, 505) }
func TestPortMatchesGoCohere_506(t *testing.T) { t.Parallel(); portMatchesRun(t, 506) }
func TestPortMatchesGoCohere_507(t *testing.T) { t.Parallel(); portMatchesRun(t, 507) }
func TestPortMatchesGoCohere_508(t *testing.T) { t.Parallel(); portMatchesRun(t, 508) }
func TestPortMatchesGoCohere_509(t *testing.T) { t.Parallel(); portMatchesRun(t, 509) }
func TestPortMatchesGoCohere_510(t *testing.T) { t.Parallel(); portMatchesRun(t, 510) }
func TestPortMatchesGoCohere_511(t *testing.T) { t.Parallel(); portMatchesRun(t, 511) }

// Each independently reported unit owns a full 90-second budget.
func portMatchesDeadline(name string) *time.Timer {
	return time.AfterFunc(90*time.Second, func() {
		fmt.Fprintf(os.Stderr, "%s exceeded its 90s deadline\n", name)
		os.Exit(124)
	})
}

// The build phase and isolated shards use these exact recipes and keys.
var portMatchesProducts sync.Map

func portMatchesProduct(t *testing.T, name string) string {
	t.Helper()
	value, _ := portMatchesProducts.LoadOrStore(name, &jsonTopBuild{})
	product := value.(*jsonTopBuild)
	product.once.Do(func() {
		goTools, err := portMatchesJsonGoToolchain()
		if err != nil {
			t.Fatal(err)
		}
		switch name {
		case "oracle":
			product.dir = buildcache.Product(t, buildcache.Inputs{
				Name:      "json split Go oracle",
				Files:     []string{"go.mod", "go.work", "cohere", "stage1/cohere/json/testdata/cohere_driver.go", "stage1/cohere/json/port_matches_prepared_split_test.go"},
				Flags:     append([]string{"go build -overlay=<product>/overlay.json -o=<product>/go-cohere ./command/formatter_comparison"}, portMatchesJsonBuildEnvironment("GOFLAGS")...),
				Toolchain: goTools,
			}, portMatchesBuildJSONGoOracle)
		case "lowered":
			product.dir = buildcache.Product(t, portMatchesJsonLoweredPortInputs(), portMatchesBuildJSONLoweredPort)
		case "release", "sanitized":
			lowered := portMatchesProduct(t, "lowered")
			source, err := os.ReadFile(filepath.Join(lowered, "main.c"))
			if err != nil {
				t.Fatal(err)
			}
			clangTools, err := portMatchesJsonClangToolchain()
			if err != nil {
				t.Fatal(err)
			}
			build := portMatchesBuildJSONReleasePort(string(source))
			if name == "sanitized" {
				build = portMatchesBuildJSONSanitizedPort(string(source))
			}
			product.dir = buildcache.Product(t, portMatchesJsonNativePortInputs(string(source), name == "sanitized", clangTools), build)
		default:
			t.Fatalf("unknown product %q", name)
		}
	})
	if product.dir == "" {
		t.Fatalf("product %s preparation failed", name)
	}
	return product.dir
}

func TestProduct_JSONGoOracle(t *testing.T)        { t.Parallel(); portMatchesProduct(t, "oracle") }
func TestProduct_JSONLoweredPort(t *testing.T)     { t.Parallel(); portMatchesProduct(t, "lowered") }
func TestProduct_JSONNativeRelease(t *testing.T)   { t.Parallel(); portMatchesProduct(t, "release") }
func TestProduct_JSONNativeSanitized(t *testing.T) { t.Parallel(); portMatchesProduct(t, "sanitized") }
