package lint

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
)

const testDecodedOptionsAndMutantShards = 6

// The fixture and its six assertions are pinned, rather than a growing corpus.
func decodedOptionsCases() []string {
	return []string{"Node", "emitted JavaScript", "native", "Go count", "mutant Node", "mutant emitted JavaScript"}
}

func decodedOptionsSlice(shard int) []int {
	var indices []int
	for index := range decodedOptionsCases() {
		if index%testDecodedOptionsAndMutantShards == shard {
			indices = append(indices, index)
		}
	}
	return indices
}

func TestDecodedOptionsAndMutant_000(t *testing.T) { t.Parallel(); decodedOptionsShard(t, 0) }
func TestDecodedOptionsAndMutant_001(t *testing.T) { t.Parallel(); decodedOptionsShard(t, 1) }
func TestDecodedOptionsAndMutant_002(t *testing.T) { t.Parallel(); decodedOptionsShard(t, 2) }
func TestDecodedOptionsAndMutant_003(t *testing.T) { t.Parallel(); decodedOptionsShard(t, 3) }
func TestDecodedOptionsAndMutant_004(t *testing.T) { t.Parallel(); decodedOptionsShard(t, 4) }
func TestDecodedOptionsAndMutant_005(t *testing.T) { t.Parallel(); decodedOptionsShard(t, 5) }

func TestDecodedOptionsAndMutantUnion(t *testing.T) {
	t.Parallel()
	// Read actual top-level registrations: a missing, repeated, or misnumbered
	// wrapper must fail even if the slice helper itself still covers every case.
	file, err := parser.ParseFile(token.NewFileSet(), "bv61ffb_decoded_options_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	registrations := map[int]int{}
	const prefix = "TestDecodedOptionsAndMutant_"
	for _, decl := range file.Decls {
		function, ok := decl.(*ast.FuncDecl)
		if !ok || !strings.HasPrefix(function.Name.Name, prefix) {
			continue
		}
		suffix := strings.TrimPrefix(function.Name.Name, prefix)
		if suffix == "Setup" {
			continue
		}
		shard, err := strconv.Atoi(suffix)
		if err != nil || shard < 0 || shard >= testDecodedOptionsAndMutantShards || suffix != fmt.Sprintf("%03d", shard) {
			t.Fatalf("invalid shard name %s", function.Name.Name)
		}
		registrations[shard]++
		calls := 0
		ast.Inspect(function.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			name, ok := call.Fun.(*ast.Ident)
			if !ok || name.Name != "decodedOptionsShard" {
				return true
			}
			calls++
			if len(call.Args) != 2 {
				t.Fatalf("bad wrapper %s", function.Name.Name)
			}
			number, ok := call.Args[1].(*ast.BasicLit)
			if !ok || number.Value != strconv.Itoa(shard) {
				t.Fatalf("wrapper %s runs wrong slice", function.Name.Name)
			}
			return true
		})
		if calls != 1 {
			t.Fatalf("wrapper %s calls runner %d times", function.Name.Name, calls)
		}
	}
	if len(registrations) != testDecodedOptionsAndMutantShards {
		t.Fatalf("registered %d shards, want %d", len(registrations), testDecodedOptionsAndMutantShards)
	}
	seen := make([]int, len(decodedOptionsCases()))
	for shard := range testDecodedOptionsAndMutantShards {
		if registrations[shard] != 1 {
			t.Fatalf("shard %03d registered %d times", shard, registrations[shard])
		}
		for _, index := range decodedOptionsSlice(shard) {
			seen[index]++
		}
	}
	for index, count := range seen {
		if count != 1 {
			t.Fatalf("case %s covered %d times", decodedOptionsCases()[index], count)
		}
	}
	t.Logf("union: %d cases, each exactly once across %d top-level shards", len(seen), len(registrations))
}

// Both the live shards and the planted-failure test use this exact evaluator.
func decodedOptionsCheck(index int, got, want []byte) error {
	name := decodedOptionsCases()[index]
	if index == 3 {
		if !bytes.Equal(got, []byte("0\n")) {
			return fmt.Errorf("JSON catch option did not override the legacy default: %q", got)
		}
	} else if index >= 4 {
		if bytes.Equal(got, want) {
			return fmt.Errorf("ignored decoded-option mutant survived on %s", name)
		}
	} else if diff := difference(got, want); diff != "" {
		return fmt.Errorf("%s: %s", name, diff)
	}
	return nil
}

func TestDecodedOptionsAndMutantPlantedFailure(t *testing.T) {
	t.Parallel()
	planted := 1
	var caught []int
	for shard := range testDecodedOptionsAndMutantShards {
		failures := 0
		for _, index := range decodedOptionsSlice(shard) {
			want := []byte("case catch\n0 findings\n")
			got := append([]byte(nil), want...)
			if index == 3 {
				got = []byte("0\n")
			}
			if index >= 4 {
				got = []byte("case catch\n1 finding\n")
			}
			if err := decodedOptionsCheck(index, got, want); err != nil {
				t.Fatalf("unplanted case %s: %v", decodedOptionsCases()[index], err)
			}
			if index == planted {
				got = []byte("case catch\nplanted disagreement\n")
			}
			if decodedOptionsCheck(index, got, want) != nil {
				failures++
			}
		}
		if failures > 0 {
			caught = append(caught, shard)
		}
	}
	if len(caught) != 1 || caught[0] != planted%testDecodedOptionsAndMutantShards {
		t.Fatalf("planted disagreement caught by shards %v", caught)
	}
	t.Logf("planted case %s caught by exactly shard-%03d", decodedOptionsCases()[planted], caught[0])
}

func decodedOptionsInputs(t *testing.T, name string) buildcache.Inputs {
	t.Helper()
	prepareRegistry(t, ".")
	files := []string{"go.mod", "go.work", "stage1/cohere/lint/bv61ffb_decoded_options_test.go", "stage1/cohere/lint/shared_test.go", "stage1/cohere/lint/registry", "oracle/adamic.mjs", "cohere/TypeScript/tsc", "cohere/TypeScript-shim"}
	for _, file := range portFiles(t) {
		files = append(files, filepath.ToSlash(filepath.Join("stage1/cohere/lint", file)))
	}
	// Hash implementation sources, excluding unrelated tests and evidence trees.
	err := filepath.WalkDir(filepath.Join(repository, "internal"), func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && (entry.Name() == "testdata" || entry.Name() == "performance") {
			return filepath.SkipDir
		}
		extension := filepath.Ext(path)
		if !entry.IsDir() && (extension == ".go" || extension == ".c" || extension == ".h") && !strings.HasSuffix(path, "_test.go") {
			relative, err := filepath.Rel(repository, path)
			if err != nil {
				return err
			}
			files = append(files, filepath.ToSlash(relative))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(files)
	return buildcache.Inputs{Name: name, Files: files, Flags: []string{"Sanitize=true", "Target=native", "Split=true", "Jobs=4", "ADAMIC_NATIVE_SPLIT=" + os.Getenv("ADAMIC_NATIVE_SPLIT"), "GOOS=" + runtime.GOOS, "GOARCH=" + runtime.GOARCH, "source-root=" + packageDirectory, "GOFLAGS=" + os.Getenv("GOFLAGS"), "GOEXPERIMENT=" + os.Getenv("GOEXPERIMENT"), "CGO_ENABLED=" + os.Getenv("CGO_ENABLED")}, Toolchain: []string{runtime.Version(), buildcache.Tool("clang", "--version")}}
}

// A nonempty setup path marks the child that owns the setup test. The parent
// runs it before m.Run, reads its immutable results, then starts leaf timers.
const decodedOptionsSetupPath = "ADAMIC_DECODED_OPTIONS_SETUP_PATH"

type decodedOptionsProducts struct {
	Directory, Changed, Module, ChangedModule, Binary, Oracle, Manifest string
	Want                                                                []byte
}

var decodedOptionsReady *decodedOptionsProducts

func decodedOptionsCommand(ctx context.Context, directory, name string, args ...string) *exec.Cmd {
	command := exec.CommandContext(ctx, name, args...)
	command.Dir = directory
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if err == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return err
	}
	command.WaitDelay = time.Second
	return command
}

func decodedOptionsBeforeTests() error {
	// TestMain is called before testing parses flags; list-only invocations must
	// remain build-free, and unrelated filtered tests must not fetch our products.
	flag.Parse()
	if os.Getenv(decodedOptionsSetupPath) != "" || flag.Lookup("test.list").Value.String() != "" {
		return nil
	}
	filter, err := regexp.Compile(flag.Lookup("test.run").Value.String())
	if err != nil {
		return err
	}
	selected := filter.MatchString("TestDecodedOptionsAndMutant_Setup")
	for shard := range testDecodedOptionsAndMutantShards {
		selected = selected || filter.MatchString(fmt.Sprintf("TestDecodedOptionsAndMutant_%03d", shard))
	}
	if !selected {
		return nil
	}
	return decodedOptionsInitialize()
}

func decodedOptionsInitialize() error {
	path := filepath.Join(sharedDirectory, "decoded-options-setup.json")
	deadline := time.Now().Add(90 * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()
	command := decodedOptionsCommand(ctx, "", os.Args[0], "-test.run=^TestDecodedOptionsAndMutant_Setup$", "-test.timeout=0", "-test.v")
	command.Env = append(os.Environ(), decodedOptionsSetupPath+"="+path, "ADAMIC_DECODED_OPTIONS_SETUP_DEADLINE="+strconv.FormatInt(deadline.UnixNano(), 10))
	var output bytes.Buffer
	command.Stdout, command.Stderr = &output, &output
	err := command.Run()
	// Prefix child test markers so test2json never reports duplicate setup tests.
	for _, line := range strings.Split(strings.TrimSuffix(output.String(), "\n"), "\n") {
		fmt.Printf("decoded-options setup: %s\n", line)
	}
	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("cooked: TestDecodedOptionsAndMutant_Setup exceeded 90s: %w", ctx.Err())
		}
		return fmt.Errorf("decoded-option setup: %w", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var ready decodedOptionsProducts
	if err := json.Unmarshal(data, &ready); err != nil {
		return err
	}
	decodedOptionsReady = &ready
	return nil
}

func decodedOptionsSetupTest(t *testing.T) {
	t.Helper()
	path := os.Getenv(decodedOptionsSetupPath)
	if path == "" {
		if decodedOptionsReady == nil {
			if err := decodedOptionsInitialize(); err != nil {
				t.Fatal(err)
			}
		}
		return
	}
	started := time.Now()
	deadline, err := strconv.ParseInt(os.Getenv("ADAMIC_DECODED_OPTIONS_SETUP_DEADLINE"), 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	// Cancel nested compiler groups before the outer setup group is killed.
	ctx, cancel := context.WithDeadline(context.Background(), time.Unix(0, deadline).Add(-250*time.Millisecond))
	defer cancel()
	// Both original and mutant source copies live in the content-addressed
	// product, rather than in a setup test's TempDir.
	lowered := decodedOptionsSetupLowered(t, false)
	changed := decodedOptionsSetupLowered(t, true)
	inputs := decodedOptionsInputs(t, "lint-decoded-options-native-setup-v2")
	inputs.Flags = append(inputs.Flags, "registry=no-empty")
	nativeProduct := buildcache.Product(t, inputs, func(output string) error {
		source, err := os.ReadFile(filepath.Join(lowered, "lint.c"))
		if err != nil {
			return err
		}
		return native.Build(string(source), filepath.Join(output, "scanner"), native.Options{Sanitize: true, Split: true, Jobs: 4})
	})
	inputs.Name = "lint-decoded-options-go-oracle-setup-v2"
	inputs.Files = append(inputs.Files, "cohere/go.mod", "cohere/internal", "stage1/cohere/lint/testdata/oracle.go")
	inputs.Flags = append(inputs.Flags, "Go oracle=full live registry", "GOTOOLCHAIN="+os.Getenv("GOTOOLCHAIN"))
	oracleProduct := buildcache.Product(t, inputs, func(output string) error {
		_, err := decodedOptionsGoOracle(ctx, output)
		return err
	})
	// The fixture and manifest belong to the parent TestMain's directory. They
	// survive the setup child and are cleaned up by the parent after all leaves.
	fixture := filepath.Join(filepath.Dir(path), "catch.ts")
	if err := os.WriteFile(fixture, []byte("try { work(); } catch(e) {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(filepath.Dir(path), "decoded-options-manifest.txt")
	row := fixture + "\tno-empty\t\t\tfalse\t{\"AllowEmptyCatch\":true}\n"
	if err := os.WriteFile(manifestPath, []byte(row), 0644); err != nil {
		t.Fatal(err)
	}
	ready := decodedOptionsProducts{
		Directory: packageDirectory, Changed: filepath.Join(changed, "port"),
		Module: filepath.Join(lowered, "lint.mjs"), ChangedModule: filepath.Join(changed, "lint.mjs"),
		Binary: filepath.Join(nativeProduct, "scanner"), Oracle: filepath.Join(oracleProduct, "oracle"), Manifest: manifestPath,
	}
	ready.Want = decodedOptionsExecute(t, ctx, ready.Oracle, "--manifest", manifestPath)
	data, err := json.Marshal(ready)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	t.Logf("TestDecodedOptionsAndMutant_Setup: %s; all shared state ready before shard deadlines", time.Since(started))
}

// This is the same overlay and full live-rule oracle as goOracleIn, with
// a context deadline and process-group cancellation instead of a lazy shared build.
func decodedOptionsGoOracle(ctx context.Context, directory string) (string, error) {
	root, err := filepath.Abs(filepath.Join(repository, "cohere"))
	if err != nil {
		return "", err
	}
	descriptors, err := registry.Generate(packageDirectory)
	if err != nil {
		return "", err
	}
	replacements := map[string]string{}
	var virtualFiles []string
	add := func(name, source string) {
		virtual := filepath.Join(root, "adamic_lint_"+name+".go")
		replacements[virtual] = source
		virtualFiles = append(virtualFiles, virtual)
	}
	add("oracle", filepath.Join(packageDirectory, "testdata/oracle.go"))
	add("registry", filepath.Join(packageDirectory, ".generated/registry.go"))
	for _, descriptor := range descriptors {
		add(strings.ReplaceAll(descriptor.Slug, "-", "_"), filepath.Join(packageDirectory, "rules", descriptor.Slug, "oracle.go"))
	}
	overlay, err := json.Marshal(map[string]any{"Replace": replacements})
	if err != nil {
		return "", err
	}
	overlayPath := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
		return "", err
	}
	binary := filepath.Join(directory, "oracle")
	args := append([]string{"build", "-overlay=" + overlayPath, "-o", binary}, virtualFiles...)
	command := decodedOptionsCommand(ctx, root, "go", args...)
	output, err := command.CombinedOutput()
	if ctx.Err() != nil {
		return "", fmt.Errorf("cooked: oracle setup: %w", ctx.Err())
	}
	if err != nil {
		return "", fmt.Errorf("Go oracle: %w\n%s", err, output)
	}
	if len(commandDiagnostics("go", output)) != 0 {
		return "", fmt.Errorf("Go oracle diagnostics: %s", output)
	}
	return binary, nil
}

func decodedOptionsSetupLowered(t *testing.T, mutated bool) string {
	t.Helper()
	name := "lint-decoded-options-lowered-setup-v2"
	if mutated {
		name += "-ignored-allowemptycatch"
	}
	inputs := decodedOptionsInputs(t, name)
	inputs.Flags = append(inputs.Flags, "registry=no-empty")
	return buildcache.Product(t, inputs, func(output string) error {
		from, to := "", ""
		if mutated {
			from, to = "'allowemptycatch'", "'ignored-allowemptycatch'"
		}
		directory := copyPort(t, filepath.Join(output, "port"), from, to, "main.ts")
		// Keep the original main/options decoder and no-empty implementation. Only
		// unused factories are omitted; Node and the Go oracle retain the full registry.
		entries, err := os.ReadDir(filepath.Join(directory, "rules"))
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if entry.IsDir() && entry.Name() != "no-empty" {
				if err := os.RemoveAll(filepath.Join(directory, "rules", entry.Name())); err != nil {
					return err
				}
			}
		}
		prepareRegistry(t, directory)
		program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
		if err != nil {
			return err
		}
		lowered, err := lower.Lower(context.Background(), program)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(output, "lint.c"), []byte(native.C(lowered)), 0644); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(output, "lint.mjs"), []byte(javascript.JavaScript(lowered)), 0644)
	})
}

func decodedOptionsExecute(t *testing.T, ctx context.Context, name string, args ...string) []byte {
	t.Helper()
	command := decodedOptionsCommand(ctx, "", name, args...)
	output, err := os.CreateTemp(t.TempDir(), "stdout-")
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	var stderr bytes.Buffer
	command.Stdout, command.Stderr = output, &stderr
	err = command.Run()
	if ctx.Err() != nil {
		t.Fatalf("cooked: shard exceeded 90s: %v", ctx.Err())
	}
	if err != nil || len(commandDiagnostics(name, stderr.Bytes())) != 0 {
		t.Fatalf("%s %v: %v\n%s", name, args, err, &stderr)
	}
	data, err := os.ReadFile(output.Name())
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func decodedOptionsShard(t *testing.T, shard int) {
	t.Helper()
	s := decodedOptionsReady
	if s == nil {
		t.Fatal("shared setup must finish before a shard starts; no lazy builds are allowed")
	}
	// No build, registry generation, oracle preparation, or lock acquisition is
	// allowed below this boundary. Each leaf owns only its case deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	started := time.Now()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	for _, index := range decodedOptionsSlice(shard) {
		var got []byte
		switch index {
		case 0:
			got = decodedOptionsExecute(t, ctx, "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(s.Directory, "main.ts"), "--manifest", s.Manifest)
		case 1:
			got = decodedOptionsExecute(t, ctx, "node", "--disable-warning=ExperimentalWarning", runner, s.Module, "--manifest", s.Manifest)
		case 2:
			got = decodedOptionsExecute(t, ctx, s.Binary, "--manifest", s.Manifest)
		case 3:
			got = decodedOptionsExecute(t, ctx, s.Oracle, "--manifest", s.Manifest, "--count")
		case 4:
			got = decodedOptionsExecute(t, ctx, "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(s.Changed, "main.ts"), "--manifest", s.Manifest)
		case 5:
			got = decodedOptionsExecute(t, ctx, "node", "--disable-warning=ExperimentalWarning", runner, s.ChangedModule, "--manifest", s.Manifest)
		default:
			t.Fatalf("unhandled case %d", index)
		}
		if err := decodedOptionsCheck(index, got, s.Want); err != nil {
			t.Fatal(err)
		}
		if index >= 4 {
			t.Logf("ignored decoded-option mutant caught on %s: %s", decodedOptionsCases()[index], difference(got, s.Want))
		}
	}
	elapsed := time.Since(started)
	t.Logf("shard-%03d cases only: %.6fs; cooked=%t", shard, elapsed.Seconds(), elapsed >= 60*time.Second)
	if elapsed >= 60*time.Second {
		t.Fatal("cooked: shard exceeded 60s budget")
	}
}

// Setup runs in a separately bounded child before the parent starts m.Run.
func TestDecodedOptionsAndMutant_Setup(t *testing.T) {
	t.Parallel()
	decodedOptionsSetupTest(t)
}
