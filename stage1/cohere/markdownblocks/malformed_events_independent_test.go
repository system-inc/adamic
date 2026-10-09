package markdownblocks

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

const testMdastMalformedEventsShards = 3

// The error modes are a pinned corpus, one case per contiguous shard.
func malformedEventsEnumeration(t *testing.T) [][]string {
	t.Helper()
	cases := []string{"unclosed", "not-open", "mismatch"}
	shards := make([][]string, testMdastMalformedEventsShards)
	seen := make(map[string]int)
	for index, name := range cases {
		shard := index * testMdastMalformedEventsShards / len(cases)
		shards[shard] = append(shards[shard], name)
	}
	for _, names := range shards {
		if len(names) == 0 {
			t.Fatal("empty malformed-events shard")
		}
		for _, name := range names {
			seen[name]++
		}
	}
	for _, name := range cases {
		if seen[name] != 1 {
			t.Fatalf("case %s occurs %d times", name, seen[name])
		}
	}
	if len(seen) != len(cases) {
		t.Fatal("malformed-events shard union differs from enumeration")
	}
	// A planted disagreement must route to exactly one shard. The opt-in
	// actual run changes that case's expected bytes and exercises the oracle check.
	caught := 0
	for _, names := range shards {
		for _, name := range names {
			if name == "not-open" {
				caught++
			}
		}
	}
	if caught != 1 {
		t.Fatalf("planted disagreement belongs to %d shards", caught)
	}
	t.Logf("shard union: %d cases, each exactly once; %d shards", len(seen), len(shards))
	return shards
}

func malformedEventsExpected(name string, truth []byte) []byte {
	expected := []byte("adamic: panic: " + string(truth))
	if os.Getenv("ADAMIC_MALFORMED_EVENTS_PLANT_FAILURE") == "1" && name == "not-open" {
		expected = append(expected, []byte("planted disagreement\n")...)
	}
	return expected
}

func malformedEventsSetupTimer(t *testing.T) func() {
	started := time.Now()
	return func() { t.Logf("TestMdastMalformedEvents (setup): %.3fs", time.Since(started).Seconds()) }
}

// Cache the lowered backend products together, so checking/lowering and code
// generation happen once before any shard starts. No IR serialization is needed.
func malformedEventsProducts(t *testing.T, ctx context.Context, main string) (string, string) {
	return malformedEventsBuildProduct(t, ctx, main, false)
}

func malformedEventsBuildProduct(t *testing.T, ctx context.Context, main string, lowerOnly bool) (string, string) {
	t.Helper()
	inputs := buildcache.Inputs{
		Name:      "markdownblocks-malformed-events-lowered",
		Files:     []string{"stage1/cohere/markdownblocks", "internal", "go.mod", "cohere"},
		Toolchain: []string{runtime.Version()},
	}
	loweredDir := buildcache.Product(t, inputs, func(dir string) error {
		program, err := loweredResult(main)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, "main.c"), []byte(native.C(program)), 0644); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "program.mjs"), []byte(javascript.JavaScript(program)), 0644)
	})
	if lowerOnly {
		return "", filepath.Join(loweredDir, "program.mjs")
	}
	options := native.Options{Sanitize: true}
	inputs.Name = "markdownblocks-malformed-events-native"
	inputs.Flags = append(native.Flags(options), "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"))
	for _, name := range []string{"ADAMIC_CLOSURE_CONVENTION", "ADAMIC_CANONICAL_CLOSURES", "ADAMIC_CLOSURE_RECEIVERS", "ADAMIC_REGEXP_REPLACE_CALLBACK", "ADAMIC_NODE_HOST", "ADAMIC_NATIVE_JOBS", "ADAMIC_GATE_UNCACHED"} {
		inputs.Flags = append(inputs.Flags, name+"="+os.Getenv(name))
	}
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
	nativeDir := buildcache.Product(t, inputs, func(dir string) error {
		source, err := os.ReadFile(filepath.Join(loweredDir, "main.c"))
		if err != nil {
			return err
		}
		return malformedEventsNativeBuild(ctx, string(source), filepath.Join(dir, "port"), options)
	})
	return filepath.Join(nativeDir, "port"), filepath.Join(loweredDir, "program.mjs")
}

// Each top-level shard is visible to go test -list and the test-only gate.
func TestMdastMalformedEvents_000(t *testing.T) {
	t.Parallel()
	malformedEventsRun(t)
}
func TestMdastMalformedEvents_001(t *testing.T) {
	t.Parallel()
	malformedEventsRun(t)
}
func TestMdastMalformedEvents_002(t *testing.T) {
	t.Parallel()
	malformedEventsRun(t)
}
func TestMdastMalformedEventsUnion(t *testing.T) {
	t.Parallel()
	shards := malformedEventsEnumeration(t)
	if len(shards) != testMdastMalformedEventsShards {
		t.Fatal("shard count")
	}
	// Keep the declaration list checked against the pinned enumeration.
	declarations := []string{"TestMdastMalformedEvents_000", "TestMdastMalformedEvents_001", "TestMdastMalformedEvents_002"}
	if len(declarations) != len(shards) {
		t.Fatal("top-level shard declaration count")
	}
}

func malformedEventsSelectedCases(t *testing.T) []string {
	shards := malformedEventsEnumeration(t)
	for index, cases := range shards {
		if t.Name() == fmt.Sprintf("TestMdastMalformedEvents_%03d", index) {
			return cases
		}
	}
	t.Fatalf("unknown malformed-events shard %s", t.Name())
	return nil
}

type malformedEventsProductSet struct{ goBinary, main, fork, nativeBinary, javascriptPath string }

var malformedEventsOnce sync.Once
var malformedEventsShared malformedEventsProductSet

func malformedEventsRun(t *testing.T) {
	products := malformedEventsSharedSetup(t)
	configureMarkdownMemory(t)
	markdownMemory.acquire(1)
	t.Cleanup(func() { markdownMemory.release(1) })
	ctx, finish := malformedEventsDeadline(t, t.Name())
	defer finish()
	goBinary, main, fork := products.goBinary, products.main, products.fork
	nativeBinary, javascriptPath := products.nativeBinary, products.javascriptPath
	for _, name := range malformedEventsSelectedCases(t) {
		truth := malformedEventsExecute(t, ctx, nil, goBinary, "--error", name)
		clean(t, "Go error as value", truth)
		original := malformedEventsExecute(t, ctx, nil, "node", "testdata/mdast_library.mjs", fork, "--error", name)
		clean(t, "fork error as value", original)
		forkMessages := map[string]string{
			"unclosed": "Cannot close document, a token (`paragraph`, 1:1-1:1) is still open\n",
			"not-open": "Cannot close `paragraph` (1:1-1:1): it’s not open\n",
			"mismatch": "Cannot close `strong` (1:1-1:1): a different token (`paragraph`, 1:1-1:1) is open\n",
		}
		equal(t, "retained fork diagnostic", original.stdout, []byte(forkMessages[name]))

		cases := "gaps/event_" + name + ".txt"

		answer := malformedEventsExecute(t, ctx, []string{"ASAN_OPTIONS=detect_leaks=0"}, nativeBinary, cases)
		for _, side := range []run{answer, malformedEventsNode(t, ctx, main, cases), malformedEventsNode(t, ctx, javascriptPath, cases)} {
			if side.exitCode != 70 {
				t.Fatalf("%s expected error exit70 got%d", name, side.exitCode)
			}
			equal(t, "empty pre-error stdout", side.stdout, nil)
			equal(t, "normalized actual error", side.stderr, malformedEventsExpected(name, truth.stdout))
		}
		scratch := t.TempDir()
		if e := os.Mkdir(filepath.Join(scratch, "testdata"), 0755); e != nil {
			t.Fatal(e)
		}
		from, to := "Cannot close document, a token", "Cannot close document, token"
		if name == "not-open" {
			from, to = "it’s not open", "it’s already closed"
		}
		if name == "mismatch" {
			from, to = "a different token", "another token"
		}
		for _, file := range []string{"tokenizerEvents.ts", "tokenArena.ts", "tokenSource.ts", "mdastNode.ts", "inputChunks.ts", "codec.ts", "mdastArena.ts", "mdastCompile.ts", "parseFrontMatter.ts", "identifier.ts", "identifierCaseKeys.ts", "identifierCaseValues.ts", "decodeString.ts", "upperEntityNames.ts", "upperEntityValues.ts", "lowerEntityNames.ts", "lowerEntityValues.ts", "testdata/mdast_probe.ts"} {
			data, e := os.ReadFile(file)
			if e != nil {
				t.Fatal(e)
			}
			if file == "mdastCompile.ts" {
				if strings.Count(string(data), from) != 1 {
					t.Fatal("error mutation anchor")
				}
				data = []byte(strings.Replace(string(data), from, to, 1))
			}
			write(t, filepath.Join(scratch, file), data)
		}
		mutant := malformedEventsNode(t, ctx, filepath.Join(scratch, "testdata/mdast_probe.ts"), cases)
		if mutant.exitCode != 70 {
			t.Fatalf("error mutant did not reach expected error: %d", mutant.exitCode)
		}
		equal(t, "error mutant stdout", mutant.stdout, nil)
		if bytes.Equal(mutant.stderr, []byte("adamic: panic: "+string(truth.stdout))) {
			t.Fatal("error mutant survived")
		}
		t.Logf("stderr-only message mutant caught; Go=%q fork=%q", truth.stdout, original.stdout)
		t.Logf("%s", truth.stdout)
	}
}

// Preparation (including the cache lock) completes before a shard's case clock
// starts. The named setup test exercises the same preparation without a deadline;
// filtered shard runs can also fetch it without charging setup to their cases.
func malformedEventsSharedSetup(t *testing.T) malformedEventsProductSet {
	t.Helper()
	malformedEventsOnce.Do(func() {
		ctx, finish := markdownLayoutSetupContext(t.Context())
		defer finish()
		defer malformedEventsSetupTimer(t)()
		configureMarkdownMemory(t)
		markdownMemory.acquire(1)
		defer markdownMemory.release(1)
		root, e := filepath.Abs(repository)
		if e != nil {
			t.Fatal(e)
		}
		cohere := filepath.Join(root, "cohere")
		goDir := malformedEventsGoProduct(t, ctx, root)
		goBinary := filepath.Join(goDir, "go-errors")
		fork := os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK")
		if fork == "" {
			fork = filepath.Join(cohere, "internal/format/prettier/bundles")
		}
		main, e := filepath.Abs("testdata/mdast_probe.ts")
		if e != nil {
			t.Fatal(e)
		}
		nativeBinary, javascriptPath := malformedEventsProducts(t, ctx, main)
		malformedEventsShared = malformedEventsProductSet{goBinary, main, fork, nativeBinary, javascriptPath}
	})

	if malformedEventsShared.goBinary == "" {
		t.Fatal("malformed-events setup did not complete")
	}
	return malformedEventsShared
}

// One 90-second hard deadline per shard, after shared setup.
func malformedEventsDeadline(t *testing.T, name string) (context.Context, func()) {
	t.Helper()
	started := time.Now()
	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	timer := time.AfterFunc(90*time.Second, func() {
		cancel()
		panic("cooked: " + name + " exceeded its 90s deadline")
	})
	return ctx, func() {
		timer.Stop()
		cancel()
		t.Logf("%s case/unit time: %.3fs", name, time.Since(started).Seconds())
	}
}

func malformedEventsCommand(ctx context.Context, name string, arguments ...string) *exec.Cmd {
	command := exec.CommandContext(ctx, name, arguments...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	command.WaitDelay = time.Second
	return command
}

func malformedEventsExecute(t *testing.T, ctx context.Context, environment []string, name string, arguments ...string) run {
	t.Helper()
	command := malformedEventsCommand(ctx, name, arguments...)
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	if ctx.Err() != nil {
		t.Fatalf("cooked: %s exceeded 90s: %v", t.Name(), ctx.Err())
	}
	var exitError *exec.ExitError
	if err != nil && !errors.As(err, &exitError) {
		t.Fatalf("running %s: %v", name, err)
	}
	return run{stdout.Bytes(), stderr.Bytes(), command.ProcessState.ExitCode()}
}

func malformedEventsNode(t *testing.T, ctx context.Context, path string, arguments ...string) run {
	t.Helper()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle", "node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	return malformedEventsExecute(t, ctx, nil, "node", append([]string{"--disable-warning=ExperimentalWarning", runner, path}, arguments...)...)
}

func malformedEventsNativeBuild(ctx context.Context, source, output string, options native.Options) error {
	library, err := native.RuntimeLibraryForSource("", source, options)
	if err != nil {
		return err
	}
	sourcePath := filepath.Join(filepath.Dir(output), "main.c")
	if err := os.WriteFile(sourcePath, []byte(source), 0644); err != nil {
		return err
	}
	flags := append(native.Flags(options), "-I", filepath.Dir(library), "-o", output, sourcePath)
	flags = append(flags, native.RuntimeLinkFlags(library)...)
	flags = append(flags, "-lm")
	if result, err := malformedEventsCommand(ctx, "clang", flags...).CombinedOutput(); err != nil {
		return fmt.Errorf("native build: %w\n%s", err, result)
	}
	return nil
}

func malformedEventsGoProduct(t *testing.T, ctx context.Context, root string) string {
	t.Helper()
	cohere := filepath.Join(root, "cohere")
	return buildcache.Product(t, buildcache.Inputs{
		Name:      "markdownblocks-malformed-events-go-errors",
		Files:     []string{"cohere", "stage1/cohere/markdownblocks/testdata/mdast_go.go", "stage1/cohere/markdownblocks/testdata/mdast_bridge.go", "stage1/cohere/markdownblocks/testdata/events_transport.go"},
		Flags:     []string{"go build", "overlay: mdast_go.go, mdast_bridge.go, events_transport.go", "GOFLAGS=" + os.Getenv("GOFLAGS"), "GOTOOLCHAIN=" + os.Getenv("GOTOOLCHAIN"), "GOOS=" + os.Getenv("GOOS"), "GOARCH=" + os.Getenv("GOARCH"), "CGO_ENABLED=" + os.Getenv("CGO_ENABLED"), "GOAMD64=" + os.Getenv("GOAMD64")},
		Toolchain: []string{buildcache.Tool("go", "version")},
	}, func(dir string) error {
		mainPath := filepath.Join(cohere, "cmd/adamic_mdast_errors/main.go")
		replace := map[string]string{}
		for _, p := range []struct{ target, source string }{{mainPath, "testdata/mdast_go.go"}, {filepath.Join(cohere, "internal/format/markdown/mdast/adamic_mdast.go"), "testdata/mdast_bridge.go"}, {filepath.Join(cohere, "internal/format/markdown/micromark/adamic_events.go"), "testdata/events_transport.go"}} {
			source, err := filepath.Abs(p.source)
			if err != nil {
				return err
			}
			replace[p.target] = source
		}
		overlay, err := json.Marshal(map[string]any{"Replace": replace})
		if err != nil {
			return err
		}
		overlayPath := filepath.Join(dir, "overlay.json")
		if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
			return err
		}
		command := malformedEventsCommand(ctx, "go", "build", "-overlay="+overlayPath, "-o", filepath.Join(dir, "go-errors"), mainPath)
		command.Dir = cohere
		if output, err := command.CombinedOutput(); err != nil {
			return fmt.Errorf("Go errors: %w\n%s", err, output)
		}
		return nil
	})
}
