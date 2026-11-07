package markdownblocks

import (
	"bytes"
	"context"
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
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/stage1/cohere/markdownblocks/oracles"
)

const testMdastMalformedEventsShards = 3

// The error modes are a pinned corpus, one case per contiguous shard.
func malformedEventsEnumeration(test *testing.T) [][]string {
	test.Helper()
	cases := []string{"unclosed", "not-open", "mismatch"}
	shards := make([][]string, testMdastMalformedEventsShards)
	seen := make(map[string]int)
	for index, name := range cases {
		shard := index * testMdastMalformedEventsShards / len(cases)
		shards[shard] = append(shards[shard], name)
	}
	for _, names := range shards {
		if len(names) == 0 {
			test.Fatal("empty malformed-events shard")
		}
		for _, name := range names {
			seen[name]++
		}
	}
	for _, name := range cases {
		if seen[name] != 1 {
			test.Fatalf("case %s occurs %d times", name, seen[name])
		}
	}
	if len(seen) != len(cases) {
		test.Fatal("malformed-events shard union differs from enumeration")
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
		test.Fatalf("planted disagreement belongs to %d shards", caught)
	}
	test.Logf("shard union: %d cases, each exactly once; %d shards", len(seen), len(shards))
	return shards
}

func malformedEventsExpected(name string, truth []byte) []byte {
	expected := []byte("adamic: panic: " + string(truth))
	if os.Getenv("ADAMIC_MALFORMED_EVENTS_PLANT_FAILURE") == "1" && name == "not-open" {
		expected = append(expected, []byte("planted disagreement\n")...)
	}
	return expected
}

func malformedEventsSetupTimer(test *testing.T) func() {
	started := time.Now()
	return func() { test.Logf("TestMdastMalformedEvents (setup): %.3fs", time.Since(started).Seconds()) }
}

// Cache the lowered backend products together, so checking/lowering and code
// generation happen once before any shard starts. No IR serialization is needed.
func malformedEventsProducts(test *testing.T, ctx context.Context, main string) (string, string) {
	return malformedEventsBuildProduct(test, ctx, main, false)
}

func malformedEventsBuildProduct(test *testing.T, ctx context.Context, main string, lowerOnly bool) (string, string) {
	test.Helper()
	inputs := buildcache.Inputs{
		Name: "markdownblocks-malformed-events-lowered",
		// What lowering reads (whitespaceLayoutBuildFiles, which TestLoweringInputsCoverCompilerDependencies keeps
		// complete), not the whole package and cohere; the native build below keeps the same Files.
		Files:     whitespaceLayoutBuildFiles(),
		Toolchain: []string{runtime.Version()},
	}
	loweredDir := buildcache.Product(test, inputs, func(dir string) error {
		program, loweredResultError := loweredResult(main)
		if loweredResultError != nil {
			return loweredResultError
		}
		if writeError := os.WriteFile(
			filepath.Join(dir, "main.c"),
			[]byte(native.C(program)),
			0644,
		); writeError != nil {
			return writeError
		}
		return os.WriteFile(filepath.Join(dir, "program.mjs"), []byte(javascript.JavaScript(program)), 0644)
	})
	if lowerOnly {
		return "", filepath.Join(loweredDir, "program.mjs")
	}
	options := native.Options{Sanitize: true}
	inputs.Name = "markdownblocks-malformed-events-native"
	inputs.Flags = append(native.Flags(options), "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"))
	for _, name := range []string{
		"ADAMIC_CLOSURE_CONVENTION",
		"ADAMIC_CANONICAL_CLOSURES",
		"ADAMIC_CLOSURE_RECEIVERS",
		"ADAMIC_REGEXP_REPLACE_CALLBACK",
		"ADAMIC_NODE_HOST",
	} {
		inputs.Flags = append(inputs.Flags, name+"="+os.Getenv(name))
	}
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
	nativeDir := buildcache.Product(test, inputs, func(dir string) error {
		source, readError := os.ReadFile(filepath.Join(loweredDir, "main.c"))
		if readError != nil {
			return readError
		}
		return malformedEventsNativeBuild(ctx, string(source), filepath.Join(dir, "port"), options)
	})
	return filepath.Join(nativeDir, "port"), filepath.Join(loweredDir, "program.mjs")
}

// Each top-level shard is visible to go test -list and the test-only gate.
func TestMdastMalformedEvents_000(test *testing.T) {
	test.Parallel()
	malformedEventsRun(test)
}
func TestMdastMalformedEvents_001(test *testing.T) {
	test.Parallel()
	malformedEventsRun(test)
}
func TestMdastMalformedEvents_002(test *testing.T) {
	test.Parallel()
	malformedEventsRun(test)
}
func TestMdastMalformedEventsUnion(test *testing.T) {
	test.Parallel()
	shards := malformedEventsEnumeration(test)
	if len(shards) != testMdastMalformedEventsShards {
		test.Fatal("shard count")
	}
	// Keep the declaration list checked against the pinned enumeration.
	declarations := []string{
		"TestMdastMalformedEvents_000",
		"TestMdastMalformedEvents_001",
		"TestMdastMalformedEvents_002",
	}
	if len(declarations) != len(shards) {
		test.Fatal("top-level shard declaration count")
	}
}

func malformedEventsSelectedCases(test *testing.T) []string {
	shards := malformedEventsEnumeration(test)
	for index, cases := range shards {
		if test.Name() == fmt.Sprintf("TestMdastMalformedEvents_%03d", index) {
			return cases
		}
	}
	test.Fatalf("unknown malformed-events shard %s", test.Name())
	return nil
}

type malformedEventsProductSet struct{ goBinary, main, fork, nativeBinary, javascriptPath string }

var malformedEventsOnce sync.Once
var malformedEventsShared malformedEventsProductSet

func malformedEventsRun(test *testing.T) {
	products := malformedEventsSharedSetup(test)
	configureMarkdownMemory(test)
	markdownMemory.acquire(1)
	test.Cleanup(func() { markdownMemory.release(1) })
	ctx, finish := malformedEventsDeadline(test, test.Name())
	defer finish()
	goBinary, main, fork := products.goBinary, products.main, products.fork
	nativeBinary, javascriptPath := products.nativeBinary, products.javascriptPath
	for _, name := range malformedEventsSelectedCases(test) {
		truth := malformedEventsExecute(test, ctx, nil, goBinary, "--error", name)
		clean(test, "Go error as value", truth)
		original := malformedEventsExecute(test, ctx, nil, "node", "testdata/mdast_library.mjs", fork, "--error", name)
		clean(test, "fork error as value", original)
		forkMessages := map[string]string{
			"unclosed": "Cannot close document, a token (`paragraph`, 1:1-1:1) is still open\n",
			"not-open": "Cannot close `paragraph` (1:1-1:1): it’s not open\n",
			"mismatch": "Cannot close `strong` (1:1-1:1): a different token (`paragraph`, 1:1-1:1) is open\n",
		}
		equal(test, "retained fork diagnostic", original.stdout, []byte(forkMessages[name]))

		cases := "gaps/event_" + name + ".txt"

		answer := malformedEventsExecute(test, ctx, []string{"ASAN_OPTIONS=detect_leaks=0"}, nativeBinary, cases)
		for _, side := range []run{
			answer,
			malformedEventsNode(test, ctx, main, cases),
			malformedEventsNode(test, ctx, javascriptPath, cases),
		} {
			if side.exitCode != 70 {
				test.Fatalf("%s expected error exit70 got%d", name, side.exitCode)
			}
			equal(test, "empty pre-error stdout", side.stdout, nil)
			equal(test, "normalized actual error", side.stderr, malformedEventsExpected(name, truth.stdout))
		}
		scratch := test.TempDir()
		if createDirectoryError := os.Mkdir(filepath.Join(scratch, "testdata"), 0755); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		from, to := "Cannot close document, a token", "Cannot close document, token"
		if name == "not-open" {
			from, to = "it’s not open", "it’s already closed"
		}
		if name == "mismatch" {
			from, to = "a different token", "another token"
		}
		for _, file := range []string{
			"tokenizerEvents.ts",
			"tokenArena.ts",
			"tokenSource.ts",
			"mdastNode.ts",
			"inputChunks.ts",
			"codec.ts",
			"mdastArena.ts",
			"mdastCompile.ts",
			"parseFrontMatter.ts",
			"identifier.ts",
			"identifierCaseKeys.ts",
			"identifierCaseValues.ts",
			"decodeString.ts",
			"upperEntityNames.ts",
			"upperEntityValues.ts",
			"lowerEntityNames.ts",
			"lowerEntityValues.ts",
			"testdata/mdast_probe.ts",
		} {
			data, readError := os.ReadFile(file)
			if readError != nil {
				test.Fatal(readError)
			}
			if file == "mdastCompile.ts" {
				if strings.Count(string(data), from) != 1 {
					test.Fatal("error mutation anchor")
				}
				data = []byte(strings.Replace(string(data), from, to, 1))
			}
			write(test, filepath.Join(scratch, file), data)
		}
		mutant := malformedEventsNode(test, ctx, filepath.Join(scratch, "testdata/mdast_probe.ts"), cases)
		if mutant.exitCode != 70 {
			test.Fatalf("error mutant did not reach expected error: %d", mutant.exitCode)
		}
		equal(test, "error mutant stdout", mutant.stdout, nil)
		if bytes.Equal(mutant.stderr, []byte("adamic: panic: "+string(truth.stdout))) {
			test.Fatal("error mutant survived")
		}
		test.Logf("stderr-only message mutant caught; Go=%q fork=%q", truth.stdout, original.stdout)
		test.Logf("%s", truth.stdout)
	}
}

// Preparation (including the cache lock) completes before a shard's case clock
// starts. The named setup test exercises the same preparation without a deadline;
// filtered shard runs can also fetch it without charging setup to their cases.
func malformedEventsSharedSetup(test *testing.T) malformedEventsProductSet {
	test.Helper()
	malformedEventsOnce.Do(func() {
		ctx, finish := markdownLayoutSetupContext(test.Context())
		defer finish()
		defer malformedEventsSetupTimer(test)()
		configureMarkdownMemory(test)
		markdownMemory.acquire(1)
		defer markdownMemory.release(1)
		root, absolutePathError := filepath.Abs(repository)
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		cohere := filepath.Join(root, "cohere")
		goDir := malformedEventsGoProduct(test, ctx, root)
		goBinary := filepath.Join(goDir, "go-errors")
		fork := os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK")
		if fork == "" {
			fork = filepath.Join(cohere, "internal/format/prettier/bundles")
		}
		main, absolutePathError := filepath.Abs("testdata/mdast_probe.ts")
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		nativeBinary, javascriptPath := malformedEventsProducts(test, ctx, main)
		malformedEventsShared = malformedEventsProductSet{goBinary, main, fork, nativeBinary, javascriptPath}
	})

	if malformedEventsShared.goBinary == "" {
		test.Fatal("malformed-events setup did not complete")
	}
	return malformedEventsShared
}

// One 90-second hard deadline per shard, after shared setup.
func malformedEventsDeadline(test *testing.T, name string) (context.Context, func()) {
	test.Helper()
	started := time.Now()
	ctx, cancel := context.WithTimeout(test.Context(), 90*time.Second)
	timer := time.AfterFunc(90*time.Second, func() {
		cancel()
		panic("cooked: " + name + " exceeded its 90s deadline")
	})
	return ctx, func() {
		timer.Stop()
		cancel()
		test.Logf("%s case/unit time: %.3fs", name, time.Since(started).Seconds())
	}
}

func malformedEventsCommand(ctx context.Context, name string, arguments ...string) *exec.Cmd {
	command := exec.CommandContext(ctx, name, arguments...)
	command.Cancel = func() error {
		killError := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if errors.Is(killError, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return killError
	}
	return command
}

// malformedEventsExecute reads a run's output through childguard.Capture, to its end with no timer deciding when:
// exec's own copying under a one-second WaitDelay reported exec.ErrWaitDelay whenever the reads hadn't reached EOF a
// second after the child exited, so a run that had finished could fail its shard under load (#n3gk22h's shape).
func malformedEventsExecute(
	test *testing.T,
	ctx context.Context,
	environment []string,
	name string,
	arguments ...string,
) run {
	test.Helper()
	command := malformedEventsCommand(ctx, name, arguments...)
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	stdout, stderr, captureError := childguard.Capture(ctx, command)
	if ctx.Err() != nil {
		test.Fatalf("cooked: %s exceeded 90s: %v", test.Name(), ctx.Err())
	}
	var exitError *exec.ExitError
	if captureError != nil && !errors.As(captureError, &exitError) {
		test.Fatalf("running %s: %v", name, captureError)
	}
	return run{stdout, stderr, command.ProcessState.ExitCode()}
}

func malformedEventsNode(test *testing.T, ctx context.Context, path string, arguments ...string) run {
	test.Helper()
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle", "node.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	return malformedEventsExecute(
		test,
		ctx,
		nil,
		"node",
		append([]string{"--disable-warning=ExperimentalWarning", runner, path}, arguments...)...,
	)
}

func malformedEventsNativeBuild(ctx context.Context, source, output string, options native.Options) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if writeError := os.WriteFile(
		filepath.Join(filepath.Dir(output), "main.c"),
		[]byte(source),
		0644,
	); writeError != nil {
		return writeError
	}
	// native.Build compiles in a directory of its own and gives clang no absolute path, so the binary names no build
	// directory: given <build directory>/main.c, the sanitized port kept it in its globals' metadata, random per build
	// (landable-17c's tree, Oct 11: markdownblocks-malformed-events-native, refs/action/ddf7d5cae0d2).
	if buildError := native.Build(source, output, options); buildError != nil {
		return fmt.Errorf("native build: %w", buildError)
	}
	return nil
}

func malformedEventsGoProduct(test *testing.T, ctx context.Context, root string) string {
	test.Helper()
	return filepath.Dir(oracles.MalformedEventsGoOracle(test))
}
