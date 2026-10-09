package markdownblocks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode/utf16"
)

const testTokenizerEventsShards = 512

// TestTokenizerEvents runs every shard when ADAMIC_TEST_SHARD is unset.
// ADAMIC_TEST_SHARD=i/n (zero based) selects parallel case and mutant units
// by their deterministic ordinal modulo n. Build products are shared for this run.
// A fixed 512-way hash of path/generated-name plus its per-name case index
// leaves headroom without moving existing cases when repository files grow.
// Shards 000–002 additionally retain the three full-corpus mutant checks.
func TestTokenizerEvents(t *testing.T) {
	setupStarted := time.Now()
	defer func() { t.Logf("setup before shards: %.3fs", time.Since(setupStarted).Seconds()) }()
	// Keep the group in the serial phase. A parallel parent retaining its
	// memory reservation while its children wait can block other admissions.
	configureMarkdownMemory(t)
	markdownMemory.acquire(4)
	t.Cleanup(func() { markdownMemory.release(4) })
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	corpus, files := blockCorpus(t, root, "whitespace")
	// Upstream selection verifies the two fixed checkout pins against Git.
	// The live repository portion must independently remain non-empty.
	repositoryFiles, cohereFiles, checkerFiles := 0, 0, 0
	for _, item := range corpus {
		switch {
		case strings.HasPrefix(item.Name, "cohere/TypeScript/"):
			checkerFiles++
		case strings.HasPrefix(item.Name, "cohere/"):
			cohereFiles++
		case !strings.HasPrefix(item.Name, "generated/"):
			repositoryFiles++
		}
	}
	if repositoryFiles == 0 {
		t.Fatal("repository Markdown corpus is empty")
	}
	// Only pinned upstream counts are fixed: cohere 7945d102 and
	// TypeScript d92d9bfe, using auditCorpus's explicitly selected roots.
	if census() && (cohereFiles != 786 || checkerFiles != 67) {
		t.Fatalf("pinned upstream file counts: cohere %d/786, TypeScript %d/67", cohereFiles, checkerFiles)
	}
	inputs := [][]uint16{{}}
	names := []string{"empty"}
	for _, item := range corpus {
		inputs = append(inputs, utf16.Encode([]rune(item.Text)))
		names = append(names, item.Name)
	}
	for unit := 0; unit <= 65535; unit++ {
		inputs = append(inputs, []uint16{uint16(unit)})
		names = append(names, fmt.Sprintf("unit/%d", unit))
	}
	alphabet := []uint16{0, 9, 10, 13, 65279, 65}
	for length := 1; length <= 5; length++ {
		count := 1
		for i := 0; i < length; i++ {
			count *= len(alphabet)
		}
		for n := 0; n < count; n++ {
			units := make([]uint16, length)
			number := n
			for i := range units {
				units[i] = alphabet[number%len(alphabet)]
				number /= len(alphabet)
			}
			inputs = append(inputs, units)
			names = append(names, fmt.Sprintf("sequence/%d/%d", length, n))
		}
	}
	for _, prefix := range []string{"", "a", "ab", "abc", "abcd", "abcde", "中", "😀", "a😀"} {
		for _, tail := range []string{"\t", "\t\t", "\r\n", "\rX\n", "\rX", "\x00\t", "\ufeff\t"} {
			inputs = append(inputs, utf16.Encode([]rune(prefix+tail)))
			names = append(names, fmt.Sprintf("columns/%q/%q", prefix, tail))
		}
	}
	mutants := []struct{ name, from, to string }{
		{"event rollback", "while(this.events.length > info.from) this.events.pop();", "while(this.events.length > info.from + 2) this.events.pop();"},
		{"virtual serialization", "if(!expandTabs && atTab) continue;", "if(expandTabs && atTab) continue;"},
		{"restored construct", "this.construct = info.construct;", "this.construct = 8;"},
	}
	keys := tokenizerEventKeys(names)
	shards := tokenizerEventShards(keys)
	if len(shards) != testTokenizerEventsShards {
		t.Fatalf("enumerated %d shards, declared %d", len(shards), testTokenizerEventsShards)
	}
	validateTokenizerEventUnion(t, len(inputs), shards)
	selected := tokenizerEventSelection(t)
	dir := t.TempDir()
	cases := filepath.Join(dir, "cases.txt")
	write(t, cases, numericBatch(inputs))
	fullCases, fullNames := cases, names
	cohere := filepath.Join(root, "cohere")
	mainPath := filepath.Join(cohere, "cmd/adamic_chunks/main.go")
	driver, err := filepath.Abs("testdata/events_go.go")
	if err != nil {
		t.Fatal(err)
	}
	bridge, err := filepath.Abs("testdata/events_bridge.go")
	if err != nil {
		t.Fatal(err)
	}
	transportBridge, err := filepath.Abs("testdata/events_transport.go")
	if err != nil {
		t.Fatal(err)
	}
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{mainPath: driver, filepath.Join(cohere, "internal/format/markdown/micromark/adamic_chunks.go"): bridge, filepath.Join(cohere, "internal/format/markdown/micromark/adamic_event_transport.go"): transportBridge}})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(dir, "overlay.json")
	write(t, overlayPath, overlay)
	goBinary := filepath.Join(dir, "go-chunks")
	// Inputs: the pinned cohere tree and these three overlay files; Go's
	// toolchain and the build flags below. No persistent package cache.
	tokenizerEventProduct(t, tokenizerEventBuildInputs{
		Name: "Go oracle", Files: []string{cohere, driver, bridge, transportBridge},
		Flags: []string{"build", "-overlay=overlay.json"}, Toolchain: runtime.Version(),
	}, dir, func(dir string) error {
		build := bounded(t, "go", "build", "-overlay="+overlayPath, "-o", filepath.Join(dir, "go-chunks"), mainPath)
		build.Dir = cohere
		output, err := combinedOutput(build)
		if err != nil {
			return fmt.Errorf("Go events %w %s", err, output)
		}
		return nil
	})
	fullWant := execute(t, nil, goBinary, fullCases)
	clean(t, "full Go mutant oracle", fullWant)
	answers := bytes.Split(bytes.TrimSuffix(fullWant.stdout, []byte("\n")), []byte("\n"))
	if len(answers) != len(inputs) {
		t.Fatalf("Go oracle rows %d, want %d", len(answers), len(inputs))
	}
	fork := os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK")
	if fork == "" {
		fork = filepath.Join(cohere, "internal/format/prettier/bundles")
	}
	installed, err := os.ReadFile(filepath.Join(fork, "plugins/markdown.js"))
	if err != nil {
		t.Fatal(err)
	}
	pinned, err := os.ReadFile(filepath.Join(cohere, "internal/format/prettier/bundles/plugins/markdown.js"))
	if err != nil {
		t.Fatal(err)
	}
	equal(t, "pinned bundle", installed, pinned)
	main, err := filepath.Abs("testdata/events_probe.ts")
	if err != nil {
		t.Fatal(err)
	}
	// Inputs: main and its transitive TypeScript imports, pinned checker,
	// lowering implementation and toolchain. Emit both backends once.
	var program *ir.Program
	var source string
	backend := filepath.Join(dir, "backend.mjs")
	tokenizerEventProduct(t, tokenizerEventBuildInputs{
		Name: "lowered program and backends", Files: []string{main, "codec.ts", "inputChunks.ts", "tokenizerEvents.ts", "tokenArena.ts", filepath.Join(root, "internal/load"), filepath.Join(root, "internal/lower"), filepath.Join(root, "internal/native"), filepath.Join(root, "internal/javascript"), filepath.Join(cohere, "TypeScript")},
		Toolchain: runtime.Version(),
	}, dir, func(dir string) error {
		var err error
		program, err = loweredResult(main)
		if err != nil {
			return err
		}
		source = native.C(program)
		if err := os.WriteFile(filepath.Join(dir, "program.c"), []byte(source), 0644); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "backend.mjs"), []byte(javascript.JavaScript(program)), 0644)
	})
	clangVersion := execute(t, nil, "clang", "--version")
	clean(t, "clang version", clangVersion)
	toolchain := strings.SplitN(string(clangVersion.stdout), "\n", 2)[0]
	binary, fast := filepath.Join(dir, "sanitized"), filepath.Join(dir, "release")
	// Inputs: emitted C, embedded runtime, native.Flags for each mode,
	// and the configured clang toolchain. Compile once before any shards.
	tokenizerEventProduct(t, tokenizerEventBuildInputs{
		Name: "sanitized native", Files: []string{filepath.Join(dir, "program.c"), filepath.Join(root, "internal/native/runtime")},
		Flags: native.Flags(native.Options{Sanitize: true}), Toolchain: toolchain,
	}, dir, func(dir string) error {
		return native.Build(source, filepath.Join(dir, "sanitized"), native.Options{Sanitize: true})
	})
	tokenizerEventProduct(t, tokenizerEventBuildInputs{
		Name: "release native", Files: []string{filepath.Join(dir, "program.c"), filepath.Join(root, "internal/native/runtime")},
		Flags: native.Flags(native.Options{}), Toolchain: toolchain,
	}, dir, func(dir string) error {
		return native.Build(source, filepath.Join(dir, "release"), native.Options{})
	})
	// Share four process slots across the whole group, including mutants.
	// Per-shard pools oversubscribe a four-CPU instance when shards overlap.
	slots := make(chan struct{}, 4)
	runMutant := func(t *testing.T, mutant int) {
		m := mutants[mutant]
		t.Logf("mutant: %s; all %d case ids", m.name, len(inputs))

		cases, want, names := fullCases, fullWant, fullNames
		scratch := t.TempDir()
		if err := os.Mkdir(filepath.Join(scratch, "testdata"), 0755); err != nil {
			t.Fatal(err)
		}
		for _, file := range []string{"inputChunks.ts", "tokenizerEvents.ts", "tokenArena.ts", "codec.ts", "testdata/events_probe.ts"} {
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			if file == "tokenizerEvents.ts" {
				if strings.Count(string(data), m.from) != 1 {
					t.Fatal("mutation anchor")
				}
				data = []byte(strings.Replace(string(data), m.from, m.to, 1))
			}
			write(t, filepath.Join(scratch, file), data)
		}
		slots <- struct{}{}
		defer func() { <-slots }()
		result := onNode(t, filepath.Join(scratch, "testdata/events_probe.ts"), cases)
		clean(t, m.name, result)
		if bytes.Equal(result.stdout, want.stdout) {
			t.Fatal("survived")
		}
		observed := strings.Split(string(result.stdout), "\n")
		for i, line := range strings.Split(string(want.stdout), "\n") {
			if i >= len(observed) {
				t.Logf("caught by missing event result%d", i)
				break
			}
			if observed[i] != line {
				t.Logf("caught by %s at event byte%d", names[i], firstDifference(observed[i], line))
				break
			}
		}
	}
	for ordinal, rows := range shards {
		if !selected(ordinal) {
			continue
		}
		t.Run(fmt.Sprintf("shard-%03d", ordinal), func(t *testing.T) {
			t.Parallel()
			if len(rows) == 0 {
				if ordinal < len(mutants) {
					runMutant(t, ordinal)
				}
				return
			}
			part := make([][]uint16, len(rows))
			for i, row := range rows {
				part[i] = inputs[row]
			}
			cases := filepath.Join(t.TempDir(), "cases.txt")
			write(t, cases, numericBatch(part))
			var expected bytes.Buffer
			for _, row := range rows {
				expected.Write(answers[row])
				expected.WriteByte('\n')
			}
			want := run{stdout: expected.Bytes()}
			compare := func(side string, result run) {
				clean(t, side, result)
				if err := tokenizerEventDifference(t.Name(), side, want.stdout, result.stdout, rows); err != nil {
					t.Fatal(err)
				}
			}
			runner, err := filepath.Abs(filepath.Join(repository, "oracle", "node.mjs"))
			if err != nil {
				t.Fatal(err)
			}
			type check struct {
				name, command     string
				args, environment []string
				repeats           int
			}
			checks := []check{
				{"actual original tokenizer", "node", []string{"testdata/events_library.mjs", fork, cases}, nil, 1},
				{"native", binary, []string{cases}, []string{"ASAN_OPTIONS=detect_leaks=0"}, 1},
				{"source Node", "node", []string{"--disable-warning=ExperimentalWarning", runner, main, cases}, nil, 1},
				{"backend", "node", []string{"--disable-warning=ExperimentalWarning", runner, backend, cases}, nil, 1},
				{"actual Go", goBinary, []string{cases}, nil, 3},
				{"native release", fast, []string{cases}, nil, 3},
				{"actual original Node", "node", []string{"testdata/events_library.mjs", fork, cases}, nil, 3},
			}
			if runtime.GOOS == "linux" {
				checks = append(checks, check{"leaks", binary, []string{cases}, []string{"ASAN_OPTIONS=detect_leaks=1"}, 1})
			} else if report := leaks(t, program, binary, cases); report != "" {
				t.Fatal(report)
			}
			// Independent sides share read-only products and case transport. Four
			// shared process slots bound the group on a four-CPU instance.
			var workers sync.WaitGroup
			defer workers.Wait()
			type checked struct {
				runs    []run
				elapsed time.Duration
			}
			tasks := make([]*fixtureTask[checked], len(checks))
			for i, side := range checks {
				tasks[i] = startFixtureTask(&workers, func() (checked, error) {
					slots <- struct{}{}
					defer func() { <-slots }()
					started := time.Now()
					var results []run
					for repeat := 0; repeat < side.repeats; repeat++ {
						result, err := executeResult(t, side.environment, side.command, side.args...)
						if err != nil {
							return checked{}, err
						}
						results = append(results, result)
					}
					return checked{results, time.Since(started)}, nil
				})
			}
			for i, side := range checks {
				result := tasks[i].await(t)
				for _, run := range result.runs {
					compare(side.name, run)
				}
				if side.repeats == 3 {
					t.Logf("%s %.1f texts/s; three runs, startup and identical numeric UTF-16/event transport included", side.name, float64(3*len(part))/result.elapsed.Seconds())
				}
			}
			t.Logf("union cases %d; stable keys hashed to shard %03d", len(rows), ordinal)
			if ordinal < len(mutants) {
				runMutant(t, ordinal)
			}
		})
	}

	t.Logf("%d cases, %d case shards, 3 full-corpus mutants in shards 000–002; %d physical documents; complete UTF-16 and generated corpus", len(inputs), len(shards), files)
}
