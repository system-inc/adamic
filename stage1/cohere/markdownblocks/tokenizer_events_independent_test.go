package markdownblocks

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

const testTokenizerEventsShards = 512

// The top-level TestTokenizerEvents_NNN units run the fixed shard grid.
// ADAMIC_TEST_SHARD=i/n (zero based) optionally selects case and mutant units
// by their deterministic ordinal modulo n. Build products are shared for this run.
// A fixed 512-way hash of path/generated-name plus its per-name case index
// leaves headroom without moving existing cases when repository files grow.
// Shards 000–002 additionally retain the three full-corpus mutant checks.
func tokenizerEventTopLevelUnit(test *testing.T, unit int) {
	setup := tokenizerEventReady(test)
	configureMarkdownMemory(test)
	markdownMemory.acquire(4)
	defer markdownMemory.release(4)
	ctx, finish := malformedEventsDeadline(test, test.Name())
	defer finish()
	inputs, names, files := setup.inputs, setup.names, setup.files
	mutants := []struct{ name, from, to string }{
		{
			"event rollback",
			"while(this.events.length > info.from) this.events.pop();",
			"while(this.events.length > info.from + 2) this.events.pop();",
		},
		{"virtual serialization", "if(!expandTabs && atTab) continue;", "if(expandTabs && atTab) continue;"},
		{"restored construct", "this.construct = info.construct;", "this.construct = 8;"},
	}
	shards := setup.shards
	if len(shards) != testTokenizerEventsShards {
		test.Fatalf("enumerated %d shards, declared %d", len(shards), testTokenizerEventsShards)
	}
	selected := tokenizerEventSelection(test)
	fullNames, fullWant, answers := names, setup.fullWant, setup.answers
	fullCases := ""
	if unit < 3 {
		fullCases = filepath.Join(test.TempDir(), "full-cases.txt")
		write(test, fullCases, setup.fullBatch)
	}
	fork := setup.fork
	main, backend, goBinary := setup.products.main, setup.products.backend, setup.products.goBinary
	binary, fast := setup.products.sanitized, setup.products.release
	slots := setup.slots
	runMutant := func(test *testing.T, mutant int) {
		m := mutants[mutant]
		test.Logf("mutant: %s; all %d case ids", m.name, len(inputs))

		cases, want, names := fullCases, fullWant, fullNames
		scratch := test.TempDir()
		if createDirectoryError := os.Mkdir(filepath.Join(scratch, "testdata"), 0755); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		for _, file := range []string{
			"inputChunks.ts",
			"tokenizerEvents.ts",
			"tokenArena.ts",
			"codec.ts",
			"testdata/events_probe.ts",
		} {
			data, readError := os.ReadFile(file)
			if readError != nil {
				test.Fatal(readError)
			}
			if file == "tokenizerEvents.ts" {
				if strings.Count(string(data), m.from) != 1 {
					test.Fatal("mutation anchor")
				}
				data = []byte(strings.Replace(string(data), m.from, m.to, 1))
			}
			write(test, filepath.Join(scratch, file), data)
		}
		slots <- struct{}{}
		defer func() { <-slots }()
		result := malformedEventsNode(test, ctx, filepath.Join(scratch, "testdata/events_probe.ts"), cases)
		clean(test, m.name, result)
		if bytes.Equal(result.stdout, want.stdout) {
			test.Fatal("survived")
		}
		observed := strings.Split(string(result.stdout), "\n")
		for index, line := range strings.Split(string(want.stdout), "\n") {
			if index >= len(observed) {
				test.Logf("caught by missing event result%d", index)
				break
			}
			if observed[index] != line {
				test.Logf("caught by %s at event byte%d", names[index], firstDifference(observed[index], line))
				break
			}
		}
	}
	for ordinal, rows := range shards {
		if ordinal != unit || !selected(ordinal) {
			continue
		}
		func(test *testing.T) {
			if len(rows) == 0 {
				if ordinal < len(mutants) {
					runMutant(test, ordinal)
				}
				return
			}
			part := make([][]uint16, len(rows))
			for index, row := range rows {
				part[index] = inputs[row]
			}
			cases := filepath.Join(test.TempDir(), "cases.txt")
			write(test, cases, numericBatch(part))
			var expected bytes.Buffer
			for _, row := range rows {
				expected.Write(answers[row])
				expected.WriteByte('\n')
			}
			want := run{stdout: expected.Bytes()}
			compare := func(side string, result run) {
				clean(test, side, result)
				if tokenizerEventDifferenceError := tokenizerEventDifference(
					test.Name(),
					side,
					want.stdout,
					result.stdout,
					rows,
				); tokenizerEventDifferenceError != nil {
					test.Fatal(tokenizerEventDifferenceError)
				}
			}
			runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle", "node.mjs"))
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
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
				checks = append(
					checks,
					check{"leaks", binary, []string{cases}, []string{"ASAN_OPTIONS=detect_leaks=1"}, 1},
				)
			} else {
				report := malformedEventsExecute(test, ctx, nil, "leaks", "--atExit", "--", fast, cases)
				if report.exitCode != 0 {
					test.Fatalf("leaks: exit %d\n%s", report.exitCode, report.stdout)
				}
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
			for index, side := range checks {
				tasks[index] = startFixtureTask(&workers, func() (checked, error) {
					slots <- struct{}{}
					defer func() { <-slots }()
					started := time.Now()
					var results []run
					for repeat := 0; repeat < side.repeats; repeat++ {
						result, pilotExecuteResultError := pilotExecuteResult(
							ctx,
							side.environment,
							side.command,
							side.args...,
						)
						if pilotExecuteResultError != nil {
							return checked{}, pilotExecuteResultError
						}
						results = append(results, result)
					}
					return checked{results, time.Since(started)}, nil
				})
			}
			for index, side := range checks {
				result := tasks[index].await(test)
				for _, run := range result.runs {
					compare(side.name, run)
				}
				if side.repeats == 3 {
					test.Logf(
						"%s %.1f texts/s; three runs, startup and identical numeric UTF-16/event transport included",
						side.name,
						float64(3*len(part))/result.elapsed.Seconds(),
					)
				}
			}
			test.Logf("union cases %d; stable keys hashed to shard %03d", len(rows), ordinal)
			if ordinal < len(mutants) {
				runMutant(test, ordinal)
			}
		}(test)
	}

	test.Logf(
		"%d cases, %d case shards, 3 full-corpus mutants in shards 000–002; %d physical documents; complete UTF-16 and generated corpus",
		len(inputs),
		len(shards),
		files,
	)
}
