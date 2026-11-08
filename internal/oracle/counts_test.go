package oracle

import (
	"context"
	"crypto/sha256"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/toolchain"
)

// updateCounts rewrites the committed per-fixture records with what was measured: go test ./internal/oracle -run
// TestCounts -args -update-counts
var updateCounts = flag.Bool(
	"update-counts",
	false,
	"rewrite per-fixture count records with the counts this run measured",
)

// countsDirectory mirrors repository fixture paths, with one committed record per fixture.
const countsDirectory = "testdata/counts"

// Extra counters register source witnesses that need an explicit frontend adapter, each a row of the table.
var additionalFixtureCounts []countCase

// A countCase is one runtime count record: its key, the first column, and prepare, which readies the row: it returns
// what the row runs (the emitted C, and how an input fixture is run), which the row's digest is made of (countDigest),
// and measure, which builds that C counted, runs it and returns the row's counts.
type countCase struct {
	key     string
	prepare func(*testing.T) (runs string, measure func() string)
}

// countsSampleSeed seeds the fixed tenth of rows every run measures whatever their digests say (countSampled). Changing
// it picks another tenth.
const countsSampleSeed = "counts-sample-v1"

// countSampled is whether the row keyed key is in the fixed sample: the first byte of the sha256 of the seed and the
// key under 26 of 256, the same rows on every run and every machine, about 1 in 10.
func countSampled(key string) bool {
	return cacheKey(countsSampleSeed, key)[:2] < "1a"
}

// countsHarness is how a row is built and run beyond what it runs (counted, interfaceCastCounted, pinnedStack): bump it
// when that changes, and every row's digest moves with it.
const countsHarness = "counts-harness-v1"

// countsShared is what every row's digest reads besides the row's own: the harness, the runtime's sources and the
// counted build's flags, none of them spelled by the machine.
var countsShared = sync.OnceValues(func() (string, error) {
	runtime, runtimeSourceDigestError := native.RuntimeSourceDigest()
	if runtimeSourceDigestError != nil {
		return "", runtimeSourceDigestError
	}
	return cacheKey(countsHarness, runtime, strings.Join(native.Flags(native.Options{Count: true}), " ")), nil
})

// countPortable is emitted C as every checkout spells it, for a row's digest: the TypeScript frontend names its source
// file by its absolute path in what a failed non-null assertion says (non_null_boolean.ts:3:17), and a scratch copy's
// path is a temporary directory, so the source file's path becomes <source> and the repository's becomes <repository>.
func countPortable(test *testing.T, source string, absolute string) string {
	test.Helper()
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	if absolute != "" {
		source = strings.ReplaceAll(source, absolute, "<source>")
	}
	return strings.ReplaceAll(source, root, "<repository>")
}

// countInputs is the trees an input fixture reads as it runs (inputIdentity's, testdata/reading and testdata/walking)
// as every checkout holds them, for an input row's digest: each entry by its path from the repository, what it is (a
// directory, a file, executable or not, or a link and its target) and a file by the sha256 of what it holds. Nothing
// else of its mode: a checkout's umask spells the rest, and git keeps only the executable bit.
var countInputs = sync.OnceValues(func() (string, error) {
	parts := []string{"count-inputs-v1"}
	for _, name := range []string{"reading", "walking"} {
		walkError := filepath.WalkDir(filepath.Join(
			repository,
			"internal",
			"oracle",
			"testdata",
			name,
		), func(path string, entry os.DirEntry, pathError error) error {
			if pathError != nil {
				return pathError
			}
			relative, relativePathError := filepath.Rel(repository, path)
			if relativePathError != nil {
				return relativePathError
			}
			info, statError := os.Lstat(path)
			if statError != nil {
				return statError
			}
			switch {
			case info.Mode()&os.ModeSymlink != 0:
				target, readLinkError := os.Readlink(path)
				if readLinkError != nil {
					return readLinkError
				}
				parts = append(parts, "link "+filepath.ToSlash(relative)+" "+target)
			case info.IsDir():
				parts = append(parts, "directory "+filepath.ToSlash(relative))
			default:
				contents, readError := os.ReadFile(path)
				if readError != nil {
					return readError
				}
				parts = append(parts, fmt.Sprintf(
					"file %s executable %v %x",
					filepath.ToSlash(relative),
					info.Mode()&0o111 != 0,
					sha256.Sum256(contents),
				))
			}
			return nil
		})
		if walkError != nil {
			return "", walkError
		}
	}
	return cacheKey(parts...), nil
})

// countDigest is the digest of the row keyed key that runs runs, as its last column holds it.
func countDigest(test *testing.T, key string, runs string) string {
	test.Helper()
	shared, countsSharedError := countsShared()
	if countsSharedError != nil {
		test.Fatal(countsSharedError)
	}
	return cacheKey(shared, key, runs)[:16]
}

// countsLine is the line a counted build writes last to stderr (runtime/count.c).
var countsLine = leakcheck.CountsLine

// pinnedStack is a command run with an 8 MiB stack, whatever the stack of whoever runs the test: a
// fixture that recurses until the stack runs out is counted at the depth it reached, which moves with
// the stack's size (ulimit -s), and the table has to be the same on every machine. The shell's ulimit
// is POSIX's, on Linux and macOS alike; where the hard limit won't allow 8 MiB, it fails out loud.
func pinnedStack(name string, arguments ...string) (string, []string) {
	return "/bin/sh", append([]string{"-c", `ulimit -s 8192 && exec "$0" "$@"`, name}, arguments...)
}

// uncounted is the fixtures the table leaves out, each for the reason the table's header gives: a
// count that measures something other than counting.
var uncounted = map[string]bool{
	// It lists the shared testdata directory, so its count moves with any entry another run leaves there (#33h46w1).
	"internal/oracle/testdata/node_fs_directory_empty_path.a": true,
	"internal/oracle/testdata/stack_overflow.a":               true,
	// Its counted build is clang over its very large C, 70.85 s on a 4-CPU share, over a unit's 60 s; the moves tests
	// build it as products and hold it to Node and the leak check.
	moveObjectsPath: true,
}

// counted lowers a fixture and returns what its row runs, the emitted C and how it is run, and measure, which builds
// that C counted, runs it, and returns its row of the table. An input fixture runs as TestInputAgreesWithNode runs it:
// from its own directory, with its arguments, a directory of its own to write in when it writes, and as nobody when the
// test is root. It must finish, as it must there, so its row counts the whole program and not where it stopped.
func counted(
	test *testing.T,
	path string,
	input bool,
	arguments []string,
	unreadable bool,
	writes bool,
) (string, func() string) {
	test.Helper()
	absolute, absolutePathError := filepath.Abs(filepath.Join(repository, path))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	// Name sources before lowering: a diagnostic string's encoded length also
	// carries its path, so replacing paths in the emitted C is too late.
	loaded, loadDisplayedError := load.LoadDisplayed([]string{absolute}, func(name string) string {
		if name == absolute {
			// The extension selects the checked TypeScript assertion frontend.
			return "<source>" + filepath.Ext(name)
		}
		return countPortable(test, name, "")
	})
	if loadDisplayedError != nil {
		test.Fatalf("Load: %v", loadDisplayedError)
	}
	program, lowerError := lower.Lower(context.Background(), loaded)
	if lowerError != nil {
		test.Fatalf("Lower: %v", lowerError)
	}
	source := native.C(program)
	runs := fmt.Sprintf(
		"input %v arguments %q unreadable %v writes %v\n%s",
		input,
		arguments,
		unreadable,
		writes,
		countPortable(test, source, absolute),
	)
	if input {
		inputs, countInputsError := countInputs()
		if countInputsError != nil {
			test.Fatal(countInputsError)
		}
		runs = "inputs " + inputs + "\n" + runs
	}
	return runs, func() string { return countedRun(test, path, absolute, source, input, arguments, unreadable, writes) }
}

// countedRun builds counted's source counted, runs it, and returns its row.
func countedRun(
	test *testing.T,
	path string,
	absolute string,
	source string,
	input bool,
	arguments []string,
	unreadable bool,
	writes bool,
) string {
	test.Helper()
	given := identity(test)
	context := cacheKey(given.context, path, fmt.Sprint(input, unreadable, writes), cacheKey(arguments...))
	if input {
		context = cacheKey(context, inputIdentity(test))
	}
	key := nativeResultKey(source, given.libraries[2], given.nodeVersion, context)
	executeCounted := func() recordedRun {
		var result run
		if input {
			shared := sharedDirectory(test)
			how := inputRun{directory: filepath.Dir(absolute), arguments: arguments}
			if os.Geteuid() == 0 {
				how.credential = &syscall.Credential{Uid: 65534, Gid: 65534}
			}
			if unreadable {
				path := filepath.Join(shared, "unreadable.txt")
				if writeError := os.WriteFile(path, []byte("secret\n"), 0o000); writeError != nil {
					test.Fatal(writeError)
				}
				how.arguments = append(append([]string{}, how.arguments...), path)
			}
			if writes {
				how.arguments = append([]string{writable(test, shared, "counted")}, how.arguments...)
			}
			if strings.HasPrefix(path, "internal/oracle/testdata/node_fs_file_") {
				how = fileSystemFilePrepare(test, shared, "counted")
			}
			binary := filepath.Join(shared, "program")
			if buildError := native.Build(source, binary, native.Options{Count: true}); buildError != nil {
				test.Fatal(buildError)
			}
			name, pinned := pinnedStack(binary, how.arguments...)
			result = executeInput(test, how, nil, name, pinned...)
			if result.exitCode != 0 {
				test.Fatalf(
					"want an input fixture's counted run to finish, as TestInputAgreesWithNode wants it to on Node: exit %d, stderr %q",
					result.exitCode,
					result.stderr,
				)
			}
		} else {
			binary := filepath.Join(test.TempDir(), "program")
			if buildError := native.Build(source, binary, native.Options{Count: true}); buildError != nil {
				test.Fatal(buildError)
			}
			name, pinned := pinnedStack(binary)
			if emitsParallelMap(source) {
				// One worker keeps peak liveness reproducible across machines.
				result = executeParallel(test, "1", false, name, pinned...)
			} else {
				result = execute(test, name, pinned...)
			}
		}
		return record(result)
	}
	var result run
	if *updateCounts {
		result = executeCounted().run()
	} else {
		result = cachedResult(test, given.cache, nativeResults, key, executeCounted).run()
	}
	match := countsLine.FindSubmatch(result.stderr)
	if match == nil {
		test.Fatalf("no counts at the end of stderr (exit %d): %q", result.exitCode, result.stderr)
	}
	return fmt.Sprintf(
		"| %s | %s | %s | %s | %s | %s | %s |",
		path,
		match[1],
		match[2],
		match[3],
		match[4],
		match[5],
		match[6],
	)
}

// countCases is every committed runtime count key, in its existing shard order: the lowered fixtures, the input
// fixtures, the node:fs file fixtures, the interface casts, then each extra counter, as registered.
func countCases() []countCase {
	var cases []countCase
	for _, fixture := range fixtures {
		if fixture.lowers && !uncounted[fixture.path] && !refusedAdamicNonNullFixture(fixture.path) {
			cases = append(cases, countCase{fixture.path, func(test *testing.T) (string, func() string) {
				return counted(test, fixture.path, false, nil, false, false)
			}})
		}
	}
	for _, fixture := range inputFixtures {
		if uncounted[fixture.path] {
			continue
		}
		cases = append(cases, countCase{fixture.path, func(test *testing.T) (string, func() string) {
			return counted(test, fixture.path, true, fixture.arguments, fixture.unreadable, fixture.writes)
		}})
	}
	for _, fixture := range fileSystemFileFixtures {
		path := "internal/oracle/testdata/node_fs_file_" + fixture + ".a"
		cases = append(cases, countCase{
			path,
			func(test *testing.T) (string, func() string) { return counted(test, path, true, nil, false, false) },
		})
	}
	cases = append(cases, interfaceCastCountCases()...)
	return append(cases, additionalFixtureCounts...)
}

// countKeys is every row's key, the key TestCountsAreRecorded's family (families_test.go) owns it by.
func countKeys() []string {
	cases := countCases()
	keys := make([]string, len(cases))
	for index, each := range cases {
		keys[index] = each.key
	}
	return keys
}

// recordedCounts reads independent committed values, never the measurements compared with them.
var recordedCounts = sync.OnceValues(func() (map[string]string, error) {
	return readCountRows(countsDirectory, "runtime", countKeys())
})

// measuredCounts is the rows this run measured under -update-counts, for TestMain to write to their individual files.
var measuredCounts sync.Map

// measuredPredicateCounts holds each predicate fixture measured under -update-counts.
var measuredPredicateCounts sync.Map

// countProductShards is how many products hold the rows' measurements, each the rows familyShard gives it: a runner
// fetches each one, so not one per row, and a change rebuilds only the shards whose rows' digests it moved.
const countProductShards = 16

// countMeasurementInputs keys the product holding shard's measurements: every row it holds by its key and digest
// (countDigest), and the clang that builds it by its report. A traced build's read set holds only what its build reads,
// and the C is emitted before it, so the digest carries what the row runs into the key.
func countMeasurementInputs(test *testing.T, shard int) (buildcache.Inputs, []countCase, []func() string) {
	test.Helper()
	compiler, pathError := toolchain.Path("clang")
	if pathError != nil {
		test.Fatal(pathError)
	}
	inputs := buildcache.Inputs{
		Name:      "count-measurements",
		Flags:     []string{fmt.Sprintf("shard %d of %d", shard, countProductShards)},
		Toolchain: []string{buildcache.Tool(compiler, "--version")},
	}
	var cases []countCase
	var measures []func() string
	for _, each := range countCases() {
		if familyShard(each.key, countProductShards) != shard {
			continue
		}
		runs, measure := each.prepare(test)
		inputs.Flags = append(inputs.Flags, each.key+" "+countDigest(test, each.key, runs))
		cases = append(cases, each)
		measures = append(measures, measure)
	}
	return inputs, cases, measures
}

// countMeasurements is shard's product: each of its rows measured from exactly what it runs, by key. A runner reads it
// and never builds it; TestProduct_CountMeasurements builds it ahead.
func countMeasurements(test *testing.T, shard int) map[string]string {
	test.Helper()
	inputs, cases, measures := countMeasurementInputs(test, shard)
	directory := buildcache.Product(test, inputs, func(directory string) error {
		var rows strings.Builder
		for index, each := range cases {
			fmt.Fprintf(&rows, "%s\t%s\n", each.key, measures[index]())
		}
		return os.WriteFile(filepath.Join(directory, "rows.tsv"), []byte(rows.String()), 0o644)
	})
	contents, readError := os.ReadFile(filepath.Join(directory, "rows.tsv"))
	if readError != nil {
		test.Fatal(readError)
	}
	measured := map[string]string{}
	for _, line := range strings.Split(strings.TrimSuffix(string(contents), "\n"), "\n") {
		if key, row, found := strings.Cut(line, "\t"); found {
			measured[key] = row
		}
	}
	return measured
}

// countShardMeasurements is each shard's measurements, read once in a process by the first row that needs them.
var countShardMeasurements [countProductShards]struct {
	once     sync.Once
	measured map[string]string
}

// countMeasured is the row keyed key as its shard's product measured it, or "" when the product couldn't be had (the
// test that first asked for it says why).
func countMeasured(test *testing.T, key string) string {
	test.Helper()
	return countShard(test, familyShard(key, countProductShards))[key]
}

// countShard is shard's measurements, had once in a process by whichever test asks first: its product test or one of
// its rows. Keying a product lowers every row the shard holds, so a run that read a shard in its product test and
// again in its first row lowered each row twice (+9% on the oracle unit, Golem, #frgdx02).
func countShard(test *testing.T, index int) map[string]string {
	test.Helper()
	shard := &countShardMeasurements[index]
	shard.once.Do(func() { shard.measured = countMeasurements(test, index) })
	return shard.measured
}

// countMeasurementProducts builds shard's product ahead, in the tree build, and in a run finds it for the shard's rows.
func countMeasurementProducts(test *testing.T, shard int) {
	test.Helper()
	if os.Getenv("ADAMIC_STAGE3_FIXTURE") != "" {
		test.Skip("stage3/fixtures' oracle hook counts nothing")
	}
	if countShard(test, shard) == nil {
		test.Fatalf("count-measurements shard %d couldn't be had: the test that first asked for it says why", shard)
	}
}

func TestProduct_CountMeasurements00(test *testing.T) {
	test.Parallel()
	countMeasurementProducts(test, 0)
}
func TestProduct_CountMeasurements01(test *testing.T) {
	test.Parallel()
	countMeasurementProducts(test, 1)
}
func TestProduct_CountMeasurements02(test *testing.T) {
	test.Parallel()
	countMeasurementProducts(test, 2)
}
func TestProduct_CountMeasurements03(test *testing.T) {
	test.Parallel()
	countMeasurementProducts(test, 3)
}
func TestProduct_CountMeasurements04(test *testing.T) {
	test.Parallel()
	countMeasurementProducts(test, 4)
}
func TestProduct_CountMeasurements05(test *testing.T) {
	test.Parallel()
	countMeasurementProducts(test, 5)
}
func TestProduct_CountMeasurements06(test *testing.T) {
	test.Parallel()
	countMeasurementProducts(test, 6)
}
func TestProduct_CountMeasurements07(test *testing.T) {
	test.Parallel()
	countMeasurementProducts(test, 7)
}
func TestProduct_CountMeasurements08(test *testing.T) {
	test.Parallel()
	countMeasurementProducts(test, 8)
}
func TestProduct_CountMeasurements09(test *testing.T) {
	test.Parallel()
	countMeasurementProducts(test, 9)
}
func TestProduct_CountMeasurements10(test *testing.T) {
	test.Parallel()
	countMeasurementProducts(test, 10)
}
func TestProduct_CountMeasurements11(test *testing.T) {
	test.Parallel()
	countMeasurementProducts(test, 11)
}
func TestProduct_CountMeasurements12(test *testing.T) {
	test.Parallel()
	countMeasurementProducts(test, 12)
}
func TestProduct_CountMeasurements13(test *testing.T) {
	test.Parallel()
	countMeasurementProducts(test, 13)
}
func TestProduct_CountMeasurements14(test *testing.T) {
	test.Parallel()
	countMeasurementProducts(test, 14)
}
func TestProduct_CountMeasurements15(test *testing.T) {
	test.Parallel()
	countMeasurementProducts(test, 15)
}

// TestCountsAreRecorded is a family (families_test.go): TestCountsAreRecorded_000 and on each hold the rows their
// shard owns to their committed files, so a change fails until that fixture's record is updated. A row's
// counts come from its measurement product (countMeasured), which the tree build measured from exactly what the row
// runs now; a sampled row (countSampled) and the planted row are measured here as well. Under -update-counts they
// measure every row, and TestMain writes each measured record.
func runCountsShard(test *testing.T, shard int) {
	test.Helper()
	planted, probing := familyProbe("TestCountsAreRecorded")
	cases := countCases()
	keys := countKeys()
	indices := familyCases(test, "TestCountsAreRecorded", shard, keys)
	if !*updateCounts {
		countShardsAhead(test, shard, keys, indices)
	}
	for _, index := range indices {
		each := cases[index]
		test.Run(each.key, func(test *testing.T) {
			test.Parallel()
			if *updateCounts {
				_, measure := each.prepare(test)
				measuredCounts.Store(each.key, measure())
				return
			}
			recorded, recordedCountsError := recordedCounts()
			if recordedCountsError != nil {
				test.Fatalf("%v: record the counts with -update-counts", recordedCountsError)
			}
			want := recorded[each.key]
			if probing && each.key == planted {
				// TestFamiliesCatchAPlantedCase's mutant: the row as recorded, one allocation more.
				if fields := strings.Split(want, " | "); len(fields) == 7 {
					allocations, _ := strconv.Atoi(fields[1])
					fields[1] = strconv.Itoa(allocations + 1)
					want = strings.Join(fields, " | ")
				}
			}
			measured := countMeasured(test, each.key)
			if !probing {
				regenerateCountRecord(test, "runtime", each.key, measured)
			}
			if measured == "" {
				test.Fatalf(
					"%s holds no measurement of %s",
					fmt.Sprintf("count-measurements shard %d", familyShard(each.key, countProductShards)),
					each.key,
				)
			}
			if measured != want {
				test.Errorf(
					"the counts changed and %s wasn't updated with them (go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts):\nrecorded: %s\nmeasured: %s",
					countFile("runtime", each.key),
					want,
					measured,
				)
			}
			if countSampled(each.key) || probing && each.key == planted {
				_, measure := each.prepare(test)
				if row := measure(); row != want {
					test.Errorf(
						"the counts changed and %s wasn't updated with them (go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts):\nrecorded: %s\nmeasured here: %s",
						countFile("runtime", each.key),
						want,
						row,
					)
				}
			}
		})
	}
}

// countShardsAhead has, before a shard's rows start, the measurements of every product shard they read, starting from
// the one its own number names and going round. The first to ask keys a product shard by lowering every row it holds,
// one after another, and a row that asked first held one of the package's parallel slots while the rows behind it held
// the rest, waiting: oracle#1/11, which holds all 32 shards and none of the product tests, keyed its 16 products almost
// one at a time on one core, 73 s on Golem where its tests' own work is a fraction (#s1g24yh). Asked here, by the
// shards' own tests, eight shards key eight products at once, and a row finds its measurement had.
func countShardsAhead(test *testing.T, shard int, keys []string, indices []int) {
	test.Helper()
	read := map[int]bool{}
	for _, index := range indices {
		read[familyShard(keys[index], countProductShards)] = true
	}
	for offset := range countProductShards {
		if product := (shard + offset) % countProductShards; read[product] {
			countShard(test, product)
		}
	}
}

// The predicate records retain their own schema and independent measurements.
func TestCountsAreRecordedTable(test *testing.T) {
	test.Parallel()
	if *updateCounts {
		for key, row := range predicateCountsTable(test) {
			measuredPredicateCounts.Store(key, row)
		}
		return
	}
	keys, predicateCountKeysError := predicateCountKeys()
	if predicateCountKeysError != nil {
		test.Fatal(predicateCountKeysError)
	}
	if _, readCountRowsError := readCountRows(countsDirectory, "predicates", keys); readCountRowsError != nil {
		test.Fatal(readCountRowsError)
	}
}

// writeCounts preserves unmeasured records and writes measured rows through the portable writer.
// A selector naming no runtime measurement still fails rather than silently doing nothing.
func writeCounts() error {
	rows := map[string]string{}
	measured := 0
	for _, key := range countKeys() {
		if row, ok := measuredCounts.Load(key); ok {
			rows[key] = row.(string)
			measured++
		} else {
			row, readCountRowError := readCountRow(countsDirectory, "runtime", key)
			if readCountRowError != nil {
				return readCountRowError
			}
			rows[key] = row
		}
	}
	if measured == 0 {
		return fmt.Errorf(
			"-update-counts measured no runtime record: select the shards with -run TestCountsAreRecorded, unanchored",
		)
	}
	keys, predicateCountKeysError := predicateCountKeys()
	if predicateCountKeysError != nil {
		return predicateCountKeysError
	}
	predicates := map[string]string{}
	for _, key := range keys {
		if row, ok := measuredPredicateCounts.Load(key); ok {
			predicates[key] = row.(string)
		} else {
			row, readCountRowError := readCountRow(countsDirectory, "predicates", key)
			if readCountRowError != nil {
				return readCountRowError
			}
			predicates[key] = row
		}
	}
	return writeCountFiles(countsDirectory, rows, predicates)
}
