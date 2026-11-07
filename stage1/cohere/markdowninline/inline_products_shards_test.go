package markdowninline

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/system-inc/adamic/stage1/cohere/markdowninline/oracles"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	_ "github.com/system-inc/adamic/internal/ccjobserver"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/compilerinputs"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Census protocol: literal count, uniform zero-padded shard names. Full corpus
// enumeration must match this constant; sampling keeps the same shard slots.
const testMarkdownInlineShards = 2048
const inlineModes = "wefnspctrukvhijlboq"

// Setup commands have no test-side deadline. The gate bounds each whole unit.
// Cancellation still reaches compiler descendants on command completion. The
// command's context comes with it, for childguard.Capture to end the reads by.
func inlineCommand(name string, arguments ...string) (*exec.Cmd, context.Context, context.CancelFunc) {
	cancelContext, cancel := context.WithCancel(context.Background())
	return inlineCommandContext(cancelContext, name, arguments...), cancelContext, cancel
}
func inlineCommandContext(callerContext context.Context, name string, arguments ...string) *exec.Cmd {
	command := exec.CommandContext(callerContext, name, arguments...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		if command.Process == nil {
			return os.ErrProcessDone
		}
		killError := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if errors.Is(killError, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return killError
	}
	return command
}
func inlineEnvironment(names ...string) []string {
	var flags []string
	for _, name := range names {
		flags = append(flags, name+"="+os.Getenv(name))
	}
	return flags
}

type inlineProductSlot struct {
	once      sync.Once
	directory string
}

var inlineProducts sync.Map

func inlineBuild(test *testing.T, in buildcache.Inputs, build func(string) error) string {
	test.Helper()
	value, _ := inlineProducts.LoadOrStore(in.Name, &inlineProductSlot{})
	slot := value.(*inlineProductSlot)
	slot.once.Do(func() { slot.directory = buildcache.Product(test, in, build) })
	if slot.directory == "" {
		test.Fatalf("product %s failed", in.Name)
	}
	return slot.directory
}

// inlineOverlayBuild is the Go side, cohere with testdata/bridge.go and the driver testdata/go_driver.go its overlay
// lays over it, a product (buildcache.GoBuildIn) built ahead. It returns the go-printer binary. The overlay names this
// checkout's paths, so it sits in a temporary directory and is never part of the product (#tqrqx60).
func inlineOverlayBuild(test *testing.T) string { return oracles.Printer(test) }
func inlineWrite(dir, name string, data []byte) error {
	return os.WriteFile(filepath.Join(dir, name), data, 0644)
}
func buildInlineProgram(main, dir string) error {
	p, loadError := load.Load([]string{main})
	if loadError != nil {
		return loadError
	}
	program, lowerError := lower.Lower(context.Background(), p)
	if lowerError != nil {
		return lowerError
	}
	if inlineWriteError := inlineWrite(dir, "program.c", []byte(native.C(program))); inlineWriteError != nil {
		return inlineWriteError
	}
	return inlineWrite(dir, "program.mjs", []byte(javascript.JavaScript(program)))
}
func buildInlineLowered(dir string) error {
	main, absolutePathError := filepath.Abs("main.ts")
	if absolutePathError != nil {
		return absolutePathError
	}
	return buildInlineProgram(main, dir)
}

type inlineSource string

func (source inlineSource) buildNative(dir string) error    { return source.buildBinary(dir, false) }
func (source inlineSource) buildSanitized(dir string) error { return source.buildBinary(dir, true) }
func (source inlineSource) buildBinary(dir string, sanitize bool) error {
	data, readError := os.ReadFile(filepath.Join(string(source), "program.c"))
	if readError != nil {
		return readError
	}
	return native.Build(string(data), filepath.Join(dir, "port"), native.Options{Sanitize: sanitize})
}
func (source inlineSource) buildNode(dir string) error {
	script, absolutePathError := filepath.Abs("testdata/build_node.mjs")
	if absolutePathError != nil {
		return absolutePathError
	}
	command, ctx, cancel := inlineCommand("node", "--disable-warning=ExperimentalWarning", script, string(source), dir)
	defer cancel()
	// Read through childguard.Capture for execute's reason: a WaitDelay could fail a build that had finished. The
	// build script prints nothing unless it fails, so only childguard's 30-minute first-output window could have
	// fired, and the 600 s Loom gives a product unit ends a hung build sooner.
	stdout, stderr, captureError := childguard.Capture(ctx, command)
	if captureError != nil {
		return fmt.Errorf("%w: %s%s", captureError, stdout, stderr)
	}
	return nil
}

type inlineMutation struct{ name, file, from, to string }

func (mutation inlineMutation) buildLowered(dir string) error {
	for _, name := range []string{"main.ts", "inline.ts", "classes.ts"} {
		data, readError := os.ReadFile(name)
		if readError != nil {
			return readError
		}
		if name == mutation.file {
			if strings.Count(string(data), mutation.from) != 1 {
				return fmt.Errorf("%s must change one site", mutation.name)
			}
			data = []byte(strings.Replace(string(data), mutation.from, mutation.to, 1))
		}
		if inlineWriteError := inlineWrite(dir, name, data); inlineWriteError != nil {
			return inlineWriteError
		}
	}
	return buildInlineProgram(filepath.Join(dir, "main.ts"), dir)
}

// Lowering runs in-process, so a lowered product's key names the compiler linked into this test binary by its sources
// (compilerinputs.Lowering), with the port's three .ts files, the lowering recipe here, and the mutations
// inline_build_products_test.go plants. The overlay Go oracle is a buildcache.GoBuildIn product of its own.
func inlineLoweredInputs(index int) buildcache.Inputs {
	name := inlineProductName(index) + " lowered"
	if index < 0 {
		name = "markdowninline lowered program"
	}
	in := buildcache.Inputs{
		Name:      name,
		Files:     compilerinputs.Lowering(),
		Toolchain: []string{runtime.Version(), buildcache.Tool("go", "version"), buildcache.Tool("clang", "--version")},
	}
	for _, file := range []string{
		"main.ts",
		"inline.ts",
		"classes.ts",
		"inline_products_shards_test.go",
		"inline_build_products_test.go",
	} {
		in.Files = append(in.Files, "stage1/cohere/markdowninline/"+file)
	}
	in.Flags = append(in.Flags, buildcache.Tool(
		"go",
		"env",
		"GOOS",
		"GOARCH",
		"CGO_ENABLED",
		"GOFLAGS",
		"CC",
		"CGO_CFLAGS",
		"CGO_CPPFLAGS",
		"CGO_CXXFLAGS",
		"CGO_LDFLAGS",
		"GOTOOLCHAIN",
		"GOEXPERIMENT",
	))
	return in
}

// A dependent product is keyed by bytes read from its prerequisite, never by
// the machine-specific cache path. This also covers each mutation's exact input.
func inlineProductSources(test *testing.T, in buildcache.Inputs, source string, names ...string) buildcache.Inputs {
	test.Helper()
	for _, name := range names {
		data, readError := os.ReadFile(filepath.Join(source, name))
		if readError != nil {
			test.Fatal(readError)
		}
		in.Flags = append(in.Flags, fmt.Sprintf("source %s sha256 %x", name, sha256.Sum256(data)))
	}
	return in
}

// A native product keys the compiler that builds it (compilerinputs.Lowering, which holds internal/native), the recipe
// here, and its C by content.
func inlineNativeInputs(test *testing.T, name, source string, sanitize bool) buildcache.Inputs {
	in := buildcache.Inputs{
		Name:      name,
		Files:     append(compilerinputs.Lowering(), "stage1/cohere/markdowninline/inline_products_shards_test.go"),
		Flags:     native.Flags(native.Options{Sanitize: sanitize}),
		Toolchain: []string{buildcache.Tool("clang", "--version"), runtime.GOOS, runtime.GOARCH},
	}
	in.Flags = append(in.Flags, inlineEnvironment(
		"LIBRARY_PATH",
		"CPATH",
		"C_INCLUDE_PATH",
		"SOURCE_DATE_EPOCH",
		"SDKROOT",
		"MACOSX_DEPLOYMENT_TARGET",
		"LD_LIBRARY_PATH",
		"DYLD_LIBRARY_PATH",
		"ADAMIC_NATIVE_SPLIT",
	)...)
	return inlineProductSources(test, in, source, "program.c")
}
func inlineNodeInputs(test *testing.T, name, source string) buildcache.Inputs {
	in := buildcache.Inputs{
		Name: name,
		Files: []string{
			"stage1/cohere/markdowninline/testdata/build_node.mjs",
			"stage1/cohere/markdowninline/inline_products_shards_test.go",
		},
		Flags:     []string{"--disable-warning=ExperimentalWarning", "stripTypeScriptTypes mode=transform"},
		Toolchain: []string{buildcache.Tool("node", "--version")},
	}
	in.Flags = append(in.Flags, inlineEnvironment("NODE_OPTIONS")...)
	return inlineProductSources(test, in, source, "main.ts", "inline.ts", "classes.ts")
}

type inlineShard struct {
	content string
	ids     []int
}

// Keys use repository-relative paths or an index within the fixed generated set.
// Neither a corpus insertion nor a checkout's absolute path moves existing cases.
func inlineShardIndex(key string, mode byte) int {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s\x00%c", key, mode)))
	return int(binary.BigEndian.Uint64(sum[:8]) % testMarkdownInlineShards)
}
func enumerateInlineShards(names []string, texts []string, files int) []inlineShard {
	shards := make([]inlineShard, testMarkdownInlineShards)
	for i := range shards {
		shards[i].content = fmt.Sprintf("hash bucket %d/%d", i, testMarkdownInlineShards)
	}
	for text := range texts {
		key := fmt.Sprintf("generated:%d", text-files)
		if text < files {
			key = "corpus:" + names[text]
		}
		for mode := range len(inlineModes) {
			shard := inlineShardIndex(key, inlineModes[mode])
			shards[shard].ids = append(shards[shard].ids, text*len(inlineModes)+mode)
		}
	}
	return shards
}
func inlineUnion(shards []inlineShard, expected map[int]bool) error {
	seen := map[int]bool{}
	count := 0
	for i, s := range shards {
		for _, id := range s.ids {
			count++
			if !expected[id] {
				return fmt.Errorf("shard-%04d foreign case %d", i, id)
			}
			if seen[id] {
				return fmt.Errorf("shard-%04d repeated case %d", i, id)
			}
			seen[id] = true
		}
	}
	if count != len(expected) || len(seen) != len(expected) {
		return fmt.Errorf("union has %d cases/%d ids, want %d", count, len(seen), len(expected))
	}
	for id := range expected {
		if !seen[id] {
			return fmt.Errorf("missing case %d", id)
		}
	}
	return nil
}
func inlineBox() (int, int, error) {
	value := os.Getenv("ADAMIC_TEST_SHARD")
	if value == "" {
		return 0, 1, nil
	}
	parts := strings.Split(value, "/")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("ADAMIC_TEST_SHARD must be zero-based i/n")
	}
	i, parseError := strconv.Atoi(parts[0])
	n, f := strconv.Atoi(parts[1])
	if parseError != nil || f != nil || n < 1 || i < 0 || i >= n {
		return 0, 0, fmt.Errorf("invalid ADAMIC_TEST_SHARD=%q", value)
	}
	return i, n, nil
}

var inlineEncoder = strings.NewReplacer(`\`, `\\`, "\n", `\n`, "\r", `\r`, "\t", `\t`)

func inlineInput(texts []string, ids []int) []byte {
	var input strings.Builder
	for _, id := range ids {
		input.WriteByte(inlineModes[id%len(inlineModes)])
		input.WriteString(inlineEncoder.Replace(texts[id/len(inlineModes)]))
		input.WriteByte('\n')
	}
	return []byte(input.String())
}
func inlineDifference(name string, ids []int, got, want []byte) error {
	if bytes.Equal(got, want) {
		return nil
	}
	offset := 0
	for offset < len(got) && offset < len(want) && got[offset] == want[offset] {
		offset++
	}
	row := bytes.Count(want[:min(offset, len(want))], []byte("\n"))
	id := -1
	if row < len(ids) {
		id = ids[row]
	}
	return fmt.Errorf(
		"%s disagreement: case %d mode %c byte %d lengths %d/%d",
		name,
		id,
		inlineModes[max(0, id)%len(inlineModes)],
		offset,
		len(got),
		len(want),
	)
}
func inlineNativeRun(test *testing.T, sanitized, fast string, args ...string) run {
	test.Helper()
	switch runtime.GOOS {
	case "linux":
		return execute(test, []string{"ASAN_OPTIONS=detect_leaks=1"}, sanitized, args...)
	case "darwin":
		r := execute(test, nil, sanitized, args...)
		check := execute(test, nil, "leaks", append([]string{"--atExit", "--", fast}, args...)...)
		if check.exitCode != 0 {
			test.Fatalf("leaks: %s %s", check.stdout, check.stderr)
		}
		return r
	default:
		test.Fatalf("no leak check for %s", runtime.GOOS)
		return run{}
	}
}
func prepareInlineShards(
	test *testing.T,
	paths, texts []string,
	files, generatedEnd int,
	selected map[string]bool,
) func(*testing.T, int) {
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	names := make([]string, files)
	for i, path := range paths {
		name, relativePathError := filepath.Rel(root, path)
		if relativePathError != nil {
			test.Fatal(relativePathError)
		}
		names[i] = filepath.ToSlash(name)
	}
	shards := enumerateInlineShards(names, texts, files)
	full := make(map[int]bool, len(texts)*len(inlineModes))
	for id := range len(texts) * len(inlineModes) {
		full[id] = true
	}
	if inlineUnionError := inlineUnion(shards, full); inlineUnionError != nil {
		test.Fatal(inlineUnionError)
	}
	test.Logf("full counted union: %d cases, each id exactly once", len(full))
	if len(shards) != testMarkdownInlineShards {
		test.Fatalf("enumerated %d shards; census constant is %d", len(shards), testMarkdownInlineShards)
	}
	expected := map[int]bool{}
	for text := range texts {
		if text >= files || selected[paths[text]] {
			for mode := range len(inlineModes) {
				expected[text*len(inlineModes)+mode] = true
			}
		}
	}
	for i := range shards {
		var ids []int
		for _, id := range shards[i].ids {
			if expected[id] {
				ids = append(ids, id)
			}
		}
		shards[i].ids = ids
	}
	if inlineUnionError := inlineUnion(shards, expected); inlineUnionError != nil {
		test.Fatal(inlineUnionError)
	}
	test.Logf("counted union: %d cases, each id exactly once; %d shards", len(expected), len(shards))
	box, boxes, inlineBoxError := inlineBox()
	if inlineBoxError != nil {
		test.Fatal(inlineBoxError)
	}
	goBinary := inlineOverlayBuild(test)
	nodeDirectory := inlineNodeProduct(test)
	lowerDirectory := inlineLoweredProduct(test, -1)
	nativeDirectory := inlineNativeProduct(test, -1, false)
	sanitizedDirectory := inlineNativeProduct(test, -1, true)
	fast, sanitized := filepath.Join(nativeDirectory, "port"), filepath.Join(sanitizedDirectory, "port")
	mutantNativeDirectory := inlineNativeProduct(test, 0, true)
	mutantFast := fast
	if runtime.GOOS == "darwin" {
		mutantFast = filepath.Join(inlineNativeProduct(test, 0, false), "port")
	}
	target := -1
	for text := range texts {
		if texts[text] == "😀😀😀😀" {
			if target != -1 {
				test.Fatal("planted text repeated")
			}
			target = text * len(inlineModes)
		}
	}
	if target < 0 {
		test.Fatal("planted case absent")
	}
	targetOwners := 0
	for _, s := range shards {
		for _, id := range s.ids {
			if id == target {
				targetOwners++
			}
		}
	}
	if targetOwners != 1 {
		test.Fatalf("planted case has %d owners", targetOwners)
	}
	// Build the original mutants once too; their small generated witness corpus
	// avoids sending every large repository text through known-wrong programs.
	type mutantProduct struct{ name, binary, fast, node string }
	var mutants []mutantProduct
	for i, m := range inlineMutations[1:] {
		index := i + 1
		bin := inlineNativeProduct(test, index, true)
		node := inlineMutationNodeProduct(test, index)
		mf := fast
		if runtime.GOOS == "darwin" {
			mf = filepath.Join(inlineNativeProduct(test, index, false), "port")
		}
		mutants = append(
			mutants,
			mutantProduct{m.name, filepath.Join(bin, "port"), mf, filepath.Join(node, "main.mjs")},
		)
	}
	library := os.Getenv("ADAMIC_MARKDOWNINLINE_LIBRARY")
	if library == "" {
		test.Log("external library not checked: set ADAMIC_MARKDOWNINLINE_LIBRARY")
	}
	// Each top-level test runs only its own stable bucket.
	return func(test *testing.T, ordinal int) {
		if ordinal%boxes != box {
			test.Skip("ADAMIC_TEST_SHARD selects another bucket")
		}
		shard := shards[ordinal]
		dir := test.TempDir()
		test.Logf("%s: %d cases", shard.content, len(shard.ids))
		if len(shard.ids) == 0 && ordinal != 0 {
			return
		}
		cases := filepath.Join(dir, "cases.txt")
		write(test, cases, inlineInput(texts, shard.ids))
		// Each side's run is timed, and the times logged, so a shard that runs long names the child it waited on
		// (#z9pt5mn).
		var times []string
		timed := func(side string, child func() run) run {
			started := time.Now()
			result := child()
			times = append(times, fmt.Sprintf("%s %.2fs", side, time.Since(started).Seconds()))
			return result
		}
		want := timed("Go", func() run { return execute(test, nil, goBinary, cases) })
		clean(test, "Go", want)
		compare := func(side string, r run) {
			clean(test, side, r)
			if inlineDifferenceError := inlineDifference(
				test.Name()+" "+side,
				shard.ids,
				r.stdout,
				want.stdout,
			); inlineDifferenceError != nil {
				test.Fatal(inlineDifferenceError)
			}
		}
		compare(
			"native",
			timed("native", func() run { return inlineNativeRun(test, sanitized, fast, "--batch", cases) }),
		)
		compare(
			"Node",
			timed(
				"Node",
				func() run { return onNode(test, filepath.Join(nodeDirectory, "main.mjs"), "--batch", cases) },
			),
		)
		compare("JavaScript backend", timed(
			"backend",
			func() run { return onNode(test, filepath.Join(lowerDirectory, "program.mjs"), "--batch", cases) },
		))
		test.Logf("children: %s", strings.Join(times, ", "))
		if library != "" {
			compare("Prettier", execute(test, nil, "node", "testdata/library.mjs", library, cases))
		}
		owns := false
		for _, id := range shard.ids {
			if id == target {
				owns = true
			}
		}
		if owns {
			r := inlineNativeRun(test, filepath.Join(mutantNativeDirectory, "port"), mutantFast, "--batch", cases)
			clean(test, "planted mutant", r)
			inlineDifferenceError := inlineDifference(test.Name(), shard.ids, r.stdout, want.stdout)
			if inlineDifferenceError == nil {
				test.Fatal("planted mutant survived")
			}
			a, b := bytes.Split(r.stdout, []byte("\n")), bytes.Split(want.stdout, []byte("\n"))
			differences := 0
			if len(a) != len(b) {
				test.Fatal("planted mutant changed row count")
			}
			for row := range b {
				if !bytes.Equal(a[row], b[row]) {
					differences++
					if row >= len(shard.ids) || shard.ids[row] != target {
						test.Fatalf("mutant changed foreign case %d", row)
					}
				}
			}
			if differences != 1 {
				test.Fatalf("planted mutant changed %d cases", differences)
			}
			test.Logf("planted mutant caught exactly once: %v", inlineDifferenceError)
		}
		for _, id := range shard.ids {
			if id%len(inlineModes) != 0 || id/len(inlineModes) >= files {
				continue
			}
			text := texts[id/len(inlineModes)]
			raw := filepath.Join(dir, "raw.md")
			write(test, raw, []byte(text))
			single := filepath.Join(dir, "single.txt")
			write(test, single, inlineInput(texts, []int{id}))
			answer := execute(test, nil, goBinary, single)
			clean(test, "Go raw", answer)
			decoded := strings.NewReplacer(`\n`, "\n", `\r`, "\r", `\t`, "\t", `\\`, `\`).Replace(
				strings.TrimSuffix(string(answer.stdout), "\n"),
			)
			r := inlineNativeRun(test, sanitized, fast, raw, "w")
			clean(test, "raw", r)
			equal(test, "raw", r.stdout, []byte(decoded))
		}
		if ordinal == 0 {
			for _, text := range []string{"", "a", "a\n", "a\r\n\n", "😀*x*", "\\_"} {
				raw := filepath.Join(dir, "raw-extra.md")
				write(test, raw, []byte(text))
				single := filepath.Join(dir, "raw-extra.txt")
				write(test, single, []byte("w"+inlineEncoder.Replace(text)+"\n"))
				answer := execute(test, nil, goBinary, single)
				clean(test, "Go raw witness", answer)
				decoded := strings.NewReplacer(`\n`, "\n", `\r`, "\r", `\t`, "\t", `\\`, `\`).Replace(
					strings.TrimSuffix(string(answer.stdout), "\n"),
				)
				r := inlineNativeRun(test, sanitized, fast, raw, "w")
				clean(test, "raw witness", r)
				equal(test, "raw witness", r.stdout, []byte(decoded))
			}
		}

		// Original mutants use a single small full generated witness set in shard 0.
		if ordinal == 0 {
			var ids []int
			for text := files; text < generatedEnd; text++ {
				if len(texts[text]) <= 128 {
					for mode := range len(inlineModes) {
						ids = append(ids, text*len(inlineModes)+mode)
					}
				}
			}
			probe := filepath.Join(dir, "mutants.txt")
			write(test, probe, inlineInput(texts, ids))
			answer := execute(test, nil, goBinary, probe)
			clean(test, "Go mutant witnesses", answer)
			for _, m := range mutants {
				for _, r := range []run{
					inlineNativeRun(test, m.binary, m.fast, "--batch", probe),
					onNode(test, m.node, "--batch", probe),
				} {
					clean(test, m.name, r)
					if inlineDifferenceError := inlineDifference(
						test.Name()+" "+m.name,
						ids,
						r.stdout,
						answer.stdout,
					); inlineDifferenceError == nil {
						test.Fatalf("%s survived", m.name)
					} else {
						test.Log(inlineDifferenceError)
					}
				}
			}
		}
	}
}

func TestMarkdownInlineShardUnion(test *testing.T) {
	test.Parallel()
	expected := map[int]bool{0: true, 1: true, 2: true}
	if inlineUnionError := inlineUnion(
		[]inlineShard{{ids: []int{0, 1}}, {ids: []int{2}}},
		expected,
	); inlineUnionError != nil {
		test.Fatal(inlineUnionError)
	}
	for _, shards := range [][]inlineShard{{{ids: []int{0, 1}}}, {{ids: []int{0, 1, 2, 2}}}, {{ids: []int{0, 1, 3}}}} {
		if inlineUnion(shards, expected) == nil {
			test.Fatal("invalid union accepted")
		}
	}
}

// Not parallel: mutates the process-wide ADAMIC_TEST_SHARD environment.
func TestMarkdownInlineShardSelector(test *testing.T) {
	for _, value := range []string{"", "0/1", "2/3"} {
		test.Setenv("ADAMIC_TEST_SHARD", value)
		if _, _, inlineBoxError := inlineBox(); inlineBoxError != nil {
			test.Fatal(inlineBoxError)
		}
	}
	for _, value := range []string{"1", "0/0", "-1/3", "3/3", "x/2"} {
		test.Setenv("ADAMIC_TEST_SHARD", value)
		if _, _, inlineBoxError := inlineBox(); inlineBoxError == nil {
			test.Fatalf("accepted %q", value)
		}
	}
}

func TestMarkdownInlineProductSourceInputs(test *testing.T) {
	test.Parallel()
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	first, second := test.TempDir(), test.TempDir()
	write(test, filepath.Join(first, "program.c"), []byte("original"))
	write(test, filepath.Join(second, "program.c"), []byte("original"))
	key := func(dir string) string {
		in := inlineProductSources(test, buildcache.Inputs{Name: "markdowninline prerequisite"}, dir, "program.c")
		value, keyError := buildcache.Key(root, in)
		if keyError != nil {
			test.Fatal(keyError)
		}
		return value
	}
	original := key(first)
	if original != key(second) {
		test.Fatal("cache location changed a content key")
	}
	write(test, filepath.Join(second, "program.c"), []byte("mutated"))
	if original == key(second) {
		test.Fatal("changed prerequisite reused a product key")
	}
}

// An edit to a package the compiler links must move every lowered and native product's key, and an edit to tsc's
// baselines, which nothing links, must keep it.
func TestMarkdownInlineCompilerProductsKeyWhatTheCompilerLinks(test *testing.T) {
	test.Parallel()
	// A native product keys its C by content, so a stand-in program.c keys it without lowering anything.
	source := test.TempDir()
	write(test, filepath.Join(source, "program.c"), nil)
	var products []buildcache.Inputs
	for index := -1; index < len(inlineMutations); index++ {
		products = append(
			products,
			inlineLoweredInputs(index),
			inlineNativeInputs(test, inlineProductName(index)+" native", source, false),
			inlineNativeInputs(test, inlineProductName(index)+" sanitized", source, true),
		)
	}
	covers := func(entry, path string) bool { return entry == path || strings.HasPrefix(path, entry+"/") }
	for _, product := range products {
		linked, baselines := false, false
		for _, entry := range product.Files {
			linked = linked || covers(entry, "cohere/TypeScript/tsc/internal/checker")
			baselines = baselines || covers(entry, "cohere/TypeScript/tsc/testdata")
		}
		if !linked {
			test.Errorf(
				"%s: no Files entry covers cohere/TypeScript/tsc/internal/checker, a package the compiler links",
				product.Name,
			)
		}
		if baselines {
			test.Errorf(
				"%s: a Files entry covers cohere/TypeScript/tsc/testdata, baselines nothing links",
				product.Name,
			)
		}
	}
	test.Logf("%d lowered and native keys name what the compiler links and leave tsc's baselines out", len(products))
}

func TestMarkdownInlineShardAssignmentStable(test *testing.T) {
	test.Parallel()
	before := enumerateInlineShards([]string{"a.md", "z.md"}, []string{"a", "z", "generated"}, 2)
	after := enumerateInlineShards([]string{"a.md", "new.md", "z.md"}, []string{"a", "new", "z", "generated"}, 3)
	locate := func(shards []inlineShard, id int) int {
		for i, shard := range shards {
			for _, found := range shard.ids {
				if found == id {
					return i
				}
			}
		}
		test.Fatalf("missing case %d", id)
		return -1
	}
	if len(before) != testMarkdownInlineShards || len(after) != testMarkdownInlineShards {
		test.Fatal("unstable shard count")
	}
	for mode := range len(inlineModes) {
		for _, pair := range [][2]int{{0, 0}, {1, 2}, {2, 3}} {
			if locate(before, pair[0]*len(inlineModes)+mode) != locate(after, pair[1]*len(inlineModes)+mode) {
				test.Fatalf("insertion moved text %d mode %c", pair[0], inlineModes[mode])
			}
		}
	}
}

func TestMarkdownInlineUnion(test *testing.T) {
	test.Parallel()
	data := collectInlineCorpus(test)
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	names := make([]string, data.files)
	for i, path := range data.paths {
		name, relativePathError := filepath.Rel(root, path)
		if relativePathError != nil {
			test.Fatal(relativePathError)
		}
		names[i] = filepath.ToSlash(name)
	}
	shards := enumerateInlineShards(names, data.texts, data.files)
	if len(shards) != testMarkdownInlineShards {
		test.Fatal("shard count differs from census constant")
	}
	expected := make(map[int]bool, len(data.texts)*len(inlineModes))
	for id := range len(data.texts) * len(inlineModes) {
		expected[id] = true
	}
	if inlineUnionError := inlineUnion(shards, expected); inlineUnionError != nil {
		test.Fatal(inlineUnionError)
	}
	test.Logf("live union: %d cases, each id exactly once; %d shards", len(expected), len(shards))
}
