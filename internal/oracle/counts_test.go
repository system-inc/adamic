package oracle

import (
	"bytes"
	"context"
	"crypto/sha256"
	"flag"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/native_runtime"
	"github.com/system-inc/adamic/internal/oracle/countproducts"
	"github.com/system-inc/adamic/internal/toolchain"
)

// updateCounts rewrites the committed per-fixture records with what was measured: go test ./internal/oracle -run TestCounts -args -update-counts
var updateCounts = flag.Bool("update-counts", false, "rewrite per-fixture count records with the counts this run measured")

// countsDirectory mirrors repository fixture paths, with one committed record per fixture.
const countsDirectory = "testdata/counts"

// Extra counters register source witnesses that need an explicit frontend adapter, each a row of the table.
var additionalFixtureCounts []countCase

// A countCase is one runtime count record: its key, the first column, and prepare, which readies the row: it returns what
// the row runs (the emitted C, and how an input fixture is run), which the row's digest is made of (countDigest), and
// measure, which builds that C counted, runs it and returns the row's counts. A row countproducts measures is readied for
// its shard's product as well (countRows).
type countCase struct {
	key     string
	prepare func(*testing.T) (runs string, measure func() string)
}

// countsSampleSeed seeds the fixed tenth of rows every run measures whatever their digests say (countSampled). Changing it
// picks another tenth.
const countsSampleSeed = "counts-sample-v1"

// countSampled is whether the row keyed key is in the fixed sample: the first byte of the sha256 of the seed and the key
// under 26 of 256, the same rows on every run and every machine, about 1 in 10.
func countSampled(key string) bool {
	return cacheKey(countsSampleSeed, key)[:2] < "1a"
}

// countsShared is what every row's digest reads besides the row's own: the harness (countproducts.Harness), the
// runtime's sources and the counted build's flags, none of them spelled by the machine.
var countsShared = sync.OnceValues(func() (string, error) {
	runtime, err := native_runtime.RuntimeSourceDigest()
	if err != nil {
		return "", err
	}
	return cacheKey(countproducts.Harness, runtime, strings.Join(native_runtime.Flags(native_runtime.Options{Count: true}), " ")), nil
})

// countPortable is emitted C as every checkout spells it, for a row's digest: the TypeScript frontend names its source
// file by its absolute path in what a failed non-null assertion says (non_null_boolean.ts:3:17), and a scratch copy's
// path is a temporary directory, so the source file's path becomes <source> and the repository's becomes <repository>.
func countPortable(t *testing.T, source string, absolute string) string {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
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
		err := filepath.WalkDir(filepath.Join(repository, "internal", "oracle", "testdata", name), func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			relative, err := filepath.Rel(repository, path)
			if err != nil {
				return err
			}
			info, err := os.Lstat(path)
			if err != nil {
				return err
			}
			switch {
			case info.Mode()&os.ModeSymlink != 0:
				target, err := os.Readlink(path)
				if err != nil {
					return err
				}
				parts = append(parts, "link "+filepath.ToSlash(relative)+" "+target)
			case info.IsDir():
				parts = append(parts, "directory "+filepath.ToSlash(relative))
			default:
				contents, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				parts = append(parts, fmt.Sprintf("file %s executable %v %x", filepath.ToSlash(relative), info.Mode()&0o111 != 0, sha256.Sum256(contents)))
			}
			return nil
		})
		if err != nil {
			return "", err
		}
	}
	return cacheKey(parts...), nil
})

// countDigest is the digest of the row keyed key that runs runs, as its last column holds it.
func countDigest(t *testing.T, key string, runs string) string {
	t.Helper()
	shared, err := countsShared()
	if err != nil {
		t.Fatal(err)
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
func counted(t *testing.T, path string, input bool, arguments []string, unreadable bool, writes bool) (string, func() string) {
	t.Helper()
	absolute, err := filepath.Abs(filepath.Join(repository, path))
	if err != nil {
		t.Fatal(err)
	}
	// Name sources before lowering: a diagnostic string's encoded length also
	// carries its path, so replacing paths in the emitted C is too late.
	loaded, err := load.LoadDisplayed([]string{absolute}, func(name string) string {
		if name == absolute {
			// The extension selects the checked TypeScript assertion frontend.
			return "<source>" + filepath.Ext(name)
		}
		return countPortable(t, name, "")
	})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	program, err := lower.Lower(context.Background(), loaded)
	if err != nil {
		t.Fatalf("Lower: %v", err)
	}
	source := native.C(program)
	runs := fmt.Sprintf("input %v arguments %q unreadable %v writes %v\n%s", input, arguments, unreadable, writes, countPortable(t, source, absolute))
	if input {
		inputs, err := countInputs()
		if err != nil {
			t.Fatal(err)
		}
		runs = "inputs " + inputs + "\n" + runs
	}
	return runs, countedMeasure(t, path, source, input, arguments, unreadable, writes)
}

// countRows are the rows countproducts measures, by key, each as its prepare readied it (countedMeasure,
// interfaceCastCounted, and the environment and realpath counters): a shard's product reads a row here once its prepare
// has run (countMeasurements). A row whose prepare readies none is measured here, in its shard's kept product.
var countRows sync.Map

// countedRow is the row countproducts measures for source, counted's C of the fixture at path, run as counted runs it.
// An input fixture runs from its own directory in the checkout, so a rows file holds that directory from the
// repository's root, and the digest of the trees it reads (countInputs), which key its shard.
func countedRow(t *testing.T, path string, source string, input bool, arguments []string, unreadable bool, writes bool) countproducts.Row {
	t.Helper()
	if !input {
		return countproducts.Row{Key: path, C: source, Run: countproducts.Plain, ParallelMap: emitsParallelMap(source)}
	}
	inputs, err := countInputs()
	if err != nil {
		t.Fatal(err)
	}
	return countproducts.Row{
		Key: path, C: source, Run: countproducts.Input, Directory: filepath.ToSlash(filepath.Dir(path)), Arguments: arguments,
		Unreadable: unreadable, Writes: writes, FSFile: strings.HasPrefix(path, "internal/oracle/testdata/node_fs_file_"), Inputs: inputs,
		// It must finish, as TestInputAgreesWithNode wants it to on Node.
		Checked: true,
	}
}

// countedMeasure readies counted's row for its shard's product (countRows), and returns measure, which measures it in
// this process (countedRun).
func countedMeasure(t *testing.T, path string, source string, input bool, arguments []string, unreadable bool, writes bool) func() string {
	t.Helper()
	countRows.Store(path, countedRow(t, path, source, input, arguments, unreadable, writes))
	return func() string { return countedRun(t, path, source, input, arguments, unreadable, writes) }
}

// countedRun measures counted's row in this process as its shard's product does (countMeasuredHere), the row kept in
// the gate cache by everything it runs, but under -update-counts, which measures afresh.
func countedRun(t *testing.T, path string, source string, input bool, arguments []string, unreadable bool, writes bool) string {
	t.Helper()
	row := countedRow(t, path, source, input, arguments, unreadable, writes)
	if *updateCounts {
		return countMeasuredHere(t, row)
	}
	given := identity(t)
	// countproducts: what the cache holds is the row's line as countproducts.Measure gives it, where it was the run's
	// output, so its keys are apart from those.
	context := cacheKey(given.context, "countproducts", path, fmt.Sprint(input, unreadable, writes), cacheKey(arguments...))
	if input {
		context = cacheKey(context, inputIdentity(t))
	}
	key := nativeResultKey(source, given.libraries[2], given.nodeVersion, context)
	return cachedResult(t, given.cache, nativeResults, key, func() string { return countMeasuredHere(t, row) })
}

// countMeasuredHere measures row in this test's process, as its shard's product measures it (countproducts.Measure):
// with the clang and the counted runtime archive a counted native.Build here would take, the archive from
// runtimeproducts' supplier (native_runtime.RuntimeLibraryForSource).
func countMeasuredHere(t *testing.T, row countproducts.Row) string {
	t.Helper()
	compiler, err := toolchain.Path(native_runtime.CompilerName(native_runtime.Options{Count: true}))
	if err != nil {
		t.Fatal(err)
	}
	library, err := native_runtime.RuntimeLibraryForSource("", row.C, native_runtime.Options{Count: true})
	if err != nil {
		t.Fatal(err)
	}
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	measured, err := countproducts.Measure(row, compiler, root, filepath.Dir(library))
	if err != nil {
		t.Fatal(err)
	}
	return measured
}

// countCases is every committed runtime count key, in its existing shard order: the lowered fixtures, the input fixtures, the
// node:fs file fixtures, the interface casts, then each extra counter, as registered.
func countCases() []countCase {
	var cases []countCase
	for _, fixture := range fixtures {
		if fixture.lowers && !uncounted[fixture.path] && !refusedAdamicNonNullFixture(fixture.path) {
			cases = append(cases, countCase{fixture.path, func(t *testing.T) (string, func() string) { return counted(t, fixture.path, false, nil, false, false) }})
		}
	}
	for _, fixture := range inputFixtures {
		if uncounted[fixture.path] {
			continue
		}
		cases = append(cases, countCase{fixture.path, func(t *testing.T) (string, func() string) {
			return counted(t, fixture.path, true, fixture.arguments, fixture.unreadable, fixture.writes)
		}})
	}
	for _, fixture := range fsFileFixtures {
		path := "internal/oracle/testdata/node_fs_file_" + fixture + ".a"
		cases = append(cases, countCase{path, func(t *testing.T) (string, func() string) { return counted(t, path, true, nil, false, false) }})
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

// countMeasurements is shard's rows, each measured from exactly what it runs, by key. A runner reads them and never
// builds them; TestProduct_CountMeasurements builds them ahead. Each row's prepare lowers it here, as it did; the rows
// it readies for countproducts (countRows) are written as a rows file (countRowsFile) and measured by countproducts'
// product, keyed by that file and its own code and never by this package's (#a4frknq). A row that readies none is
// measured here, in the shard's kept product, keyed as every row once was (countKeptMeasurements): today the program
// region rows, built with a program region's archives and held to a balanced count, their parse rows' C a product of
// its own, and the function and merged-field ruling rows, whose C is lowered from a copy in a test's temporary
// directory by its path there (rulingProgramAt), so a rows file holding it would be other bytes on every run.
func countMeasurements(t *testing.T, shard int) map[string]string {
	t.Helper()
	var rows []countproducts.Row
	var kept []keptCount
	for _, each := range countCases() {
		if familyShard(each.key, countProductShards) != shard {
			continue
		}
		runs, measure := each.prepare(t)
		if row, readied := countRows.Load(each.key); readied {
			rows = append(rows, row.(countproducts.Row))
		} else {
			kept = append(kept, keptCount{each.key, runs, measure})
		}
	}
	measured := map[string]string{}
	if len(rows) > 0 {
		root, err := filepath.Abs(repository)
		if err != nil {
			t.Fatal(err)
		}
		encoded, sets, err := countproducts.Rows(rows, root)
		if err != nil {
			t.Fatal(err)
		}
		file := countRowsFile(t, shard, encoded)
		archives := make([]string, len(sets))
		for index, set := range sets {
			archives[index] = countArchive(t, set)
		}
		maps.Copy(measured, countMeasurementsIn(t, countproducts.Product(t, file, root, archives)))
	}
	if len(kept) > 0 {
		maps.Copy(measured, countKeptMeasurements(t, shard, kept))
	}
	return measured
}

// countRowsFile is shard's rows file, encoded, as a product of its own, which countproducts' product reads by its bytes:
// a file a test wrote outside every product is a read of temp its window didn't make (agreementproducts' C, #egjbcre).
// It's keyed by the bytes it holds, so building it only writes them.
func countRowsFile(t *testing.T, shard int, encoded []byte) string {
	t.Helper()
	inputs := buildcache.Inputs{
		Name:  "count-measurement rows",
		Flags: []string{fmt.Sprintf("shard %d of %d", shard, countProductShards), fmt.Sprintf("rows %x", sha256.Sum256(encoded))},
	}
	directory := buildcache.Product(t, inputs, func(directory string) error {
		return os.WriteFile(filepath.Join(directory, "rows.json"), encoded, 0o644)
	})
	return filepath.Join(directory, "rows.json")
}

// countArchive is the directory of the counted runtime archive a program turning on the features of set links, the one
// a counted native.Build of it here would link, from runtimeproducts' supplier (native_runtime.RuntimeLibraryForSource).
func countArchive(t *testing.T, set []string) string {
	t.Helper()
	var source strings.Builder
	for _, feature := range set {
		fmt.Fprintf(&source, "#define %s 1\n", feature)
	}
	library, err := native_runtime.RuntimeLibraryForSource("", source.String(), native_runtime.Options{Count: true})
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Dir(library)
}

// A keptCount is a row its shard measures here, in its kept product (countKeptMeasurements): its key, what it runs, and
// its measure, as its prepare readied them.
type keptCount struct {
	key     string
	runs    string
	measure func() string
}

// countKeptMeasurements is the rows kept, of shard's, measured here, each by its own measure, in a product keyed as
// count-measurements was: by each row's key and digest (countDigest), and the clang that builds it by its report. A
// traced build's read set holds only what its build reads, and the C is emitted before it, so the digest carries what
// the row runs into the key.
func countKeptMeasurements(t *testing.T, shard int, kept []keptCount) map[string]string {
	t.Helper()
	compiler, err := toolchain.Path("clang")
	if err != nil {
		t.Fatal(err)
	}
	inputs := buildcache.Inputs{
		Name:      "count-measurements kept",
		Flags:     []string{fmt.Sprintf("shard %d of %d", shard, countProductShards)},
		Toolchain: []string{buildcache.CompilerReport(compiler)},
	}
	for _, each := range kept {
		inputs.Flags = append(inputs.Flags, each.key+" "+countDigest(t, each.key, each.runs))
	}
	return countMeasurementsIn(t, buildcache.Product(t, inputs, func(directory string) error {
		var rows strings.Builder
		for _, each := range kept {
			fmt.Fprintf(&rows, "%s\t%s\n", each.key, each.measure())
		}
		return os.WriteFile(filepath.Join(directory, "rows.tsv"), []byte(rows.String()), 0o644)
	}))
}

// countMeasurementsIn is the rows a product's rows.tsv holds, by key: a line each, its key, a tab and its row.
func countMeasurementsIn(t *testing.T, directory string) map[string]string {
	t.Helper()
	contents, err := os.ReadFile(filepath.Join(directory, "rows.tsv"))
	if err != nil {
		t.Fatal(err)
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
func countMeasured(t *testing.T, key string) string {
	t.Helper()
	return countShard(t, familyShard(key, countProductShards))[key]
}

// countShard is shard's measurements, had once in a process by whichever test asks first: its product test or one of
// its rows. Keying a product lowers every row the shard holds, so a run that read a shard in its product test and
// again in its first row lowered each row twice (+9% on the oracle unit, Golem, #frgdx02).
func countShard(t *testing.T, index int) map[string]string {
	t.Helper()
	shard := &countShardMeasurements[index]
	shard.once.Do(func() { shard.measured = countMeasurements(t, index) })
	return shard.measured
}

// countMeasurementProducts builds shard's product ahead, in the tree build, and in a run finds it for the shard's rows.
//
// Its tests are sequential, never t.Parallel(): each settles its shard's countproducts product, a lean build keyed by
// its own window, and go runs no parallel top-level test's body until every sequential one has finished, so no other
// test reads inside the window to put its reads in a shard's key (internal/lower's files, as it lowers, among them). No
// oracle test leaves a goroutine reading past its end either, a watcher or a server: every one a test starts is done
// before it returns (moves_test.go's two emitters, each received; output_test.go's prompt reader, received;
// cache_test.go's workers, waited for). Inside its window a shard's rows run four at a time, every one finished before
// its build returns (countproducts.BuildCountMeasurements). Keying a shard still lowers each of its rows first, so a
// run of them all lowers one shard at a time.
func countMeasurementProducts(t *testing.T, shard int) {
	t.Helper()
	if os.Getenv("ADAMIC_STAGE3_FIXTURE") != "" {
		t.Skip("stage3/fixtures' oracle hook counts nothing")
	}
	if countShard(t, shard) == nil {
		t.Fatalf("count-measurements shard %d couldn't be had: the test that first asked for it says why", shard)
	}
}

// Not parallel: its shard's lean window keys every read the process makes while it's open (countMeasurementProducts).
func TestProduct_CountMeasurements00(t *testing.T) { countMeasurementProducts(t, 0) }

// Not parallel: its shard's lean window keys every read the process makes while it's open (countMeasurementProducts).
func TestProduct_CountMeasurements01(t *testing.T) { countMeasurementProducts(t, 1) }

// Not parallel: its shard's lean window keys every read the process makes while it's open (countMeasurementProducts).
func TestProduct_CountMeasurements02(t *testing.T) { countMeasurementProducts(t, 2) }

// Not parallel: its shard's lean window keys every read the process makes while it's open (countMeasurementProducts).
func TestProduct_CountMeasurements03(t *testing.T) { countMeasurementProducts(t, 3) }

// Not parallel: its shard's lean window keys every read the process makes while it's open (countMeasurementProducts).
func TestProduct_CountMeasurements04(t *testing.T) { countMeasurementProducts(t, 4) }

// Not parallel: its shard's lean window keys every read the process makes while it's open (countMeasurementProducts).
func TestProduct_CountMeasurements05(t *testing.T) { countMeasurementProducts(t, 5) }

// Not parallel: its shard's lean window keys every read the process makes while it's open (countMeasurementProducts).
func TestProduct_CountMeasurements06(t *testing.T) { countMeasurementProducts(t, 6) }

// Not parallel: its shard's lean window keys every read the process makes while it's open (countMeasurementProducts).
func TestProduct_CountMeasurements07(t *testing.T) { countMeasurementProducts(t, 7) }

// Not parallel: its shard's lean window keys every read the process makes while it's open (countMeasurementProducts).
func TestProduct_CountMeasurements08(t *testing.T) { countMeasurementProducts(t, 8) }

// Not parallel: its shard's lean window keys every read the process makes while it's open (countMeasurementProducts).
func TestProduct_CountMeasurements09(t *testing.T) { countMeasurementProducts(t, 9) }

// Not parallel: its shard's lean window keys every read the process makes while it's open (countMeasurementProducts).
func TestProduct_CountMeasurements10(t *testing.T) { countMeasurementProducts(t, 10) }

// Not parallel: its shard's lean window keys every read the process makes while it's open (countMeasurementProducts).
func TestProduct_CountMeasurements11(t *testing.T) { countMeasurementProducts(t, 11) }

// Not parallel: its shard's lean window keys every read the process makes while it's open (countMeasurementProducts).
func TestProduct_CountMeasurements12(t *testing.T) { countMeasurementProducts(t, 12) }

// Not parallel: its shard's lean window keys every read the process makes while it's open (countMeasurementProducts).
func TestProduct_CountMeasurements13(t *testing.T) { countMeasurementProducts(t, 13) }

// Not parallel: its shard's lean window keys every read the process makes while it's open (countMeasurementProducts).
func TestProduct_CountMeasurements14(t *testing.T) { countMeasurementProducts(t, 14) }

// Not parallel: its shard's lean window keys every read the process makes while it's open (countMeasurementProducts).
func TestProduct_CountMeasurements15(t *testing.T) { countMeasurementProducts(t, 15) }

// A shard's rows file is the same bytes however many times its rows are readied, and in whatever order: one row of each
// way countproducts runs one, readied twice in opposite orders, encodes to one file, so a shard's key moves only with
// what its rows run. Mutants: countproducts.Rows writing rows in the order it's handed them, or ranging over a map of
// them, rather than sorted by key; a row's C naming its test's temporary directory, as a ruling row's does
// (rulingProgramAt).
func TestACountRowsFileIsTheSameBytesTwice(t *testing.T) {
	t.Parallel()
	// A plain row, input rows with arguments that aren't valid UTF-8 and with a directory to write in, a node:fs file
	// fixture, an environment row and a realpath row, and a checked predicate row and a statements_small TypeScript row,
	// whose C is lowered from a copy in a test's temporary directory under its logical name, so each is readied from
	// another copy each time.
	keys := []string{
		"dedication/dedication.a", "internal/oracle/testdata/arguments.a", "internal/oracle/testdata/walk.a",
		"internal/oracle/testdata/node_fs_file_close.a", "internal/oracle/testdata/process_env/reads.a", "internal/oracle/testdata/host_realpath/paths.a",
		"internal/lower/testdata/predicates_checked/true.a (.ts checked mode)", "internal/oracle/testdata/statements_small_typescript/nonnull.a",
	}
	cases := map[string]countCase{}
	for _, each := range countCases() {
		cases[each.key] = each
	}
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	readied := func(order []string) []byte {
		t.Helper()
		var rows []countproducts.Row
		for _, key := range order {
			each, found := cases[key]
			if !found {
				t.Fatalf("%s is no count row", key)
			}
			each.prepare(t)
			row, held := countRows.Load(key)
			if !held {
				t.Fatalf("%s readied no countproducts row", key)
			}
			rows = append(rows, row.(countproducts.Row))
		}
		encoded, _, err := countproducts.Rows(rows, root)
		if err != nil {
			t.Fatal(err)
		}
		return encoded
	}
	reversed := slices.Clone(keys)
	slices.Reverse(reversed)
	first, second := readied(keys), readied(reversed)
	if !bytes.Equal(first, second) {
		at := 0
		for at < min(len(first), len(second)) && first[at] == second[at] {
			at++
		}
		t.Errorf("one shard's rows, readied twice, encoded to %d and %d bytes, apart from byte %d", len(first), len(second), at)
	}
}

// TestCountsAreRecorded is a family (families_test.go): TestCountsAreRecorded_000 and on each hold the rows their
// shard owns to their committed files, so a change fails until that fixture's record is updated. A row's
// counts come from its measurement product (countMeasured), which the tree build measured from exactly what the row runs
// now; a sampled row (countSampled) and the planted row are measured here as well. Under -update-counts they measure
// every row, and TestMain writes each measured record.
func runCountsShard(t *testing.T, shard int) {
	t.Helper()
	planted, probing := familyProbe("TestCountsAreRecorded")
	cases := countCases()
	keys := countKeys()
	indices := familyCases(t, "TestCountsAreRecorded", shard, keys)
	if !*updateCounts {
		countShardsAhead(t, shard, keys, indices)
	}
	for _, index := range indices {
		each := cases[index]
		t.Run(each.key, func(t *testing.T) {
			t.Parallel()
			if *updateCounts {
				_, measure := each.prepare(t)
				measuredCounts.Store(each.key, measure())
				return
			}
			recorded, err := recordedCounts()
			if err != nil {
				t.Fatalf("%v: record the counts with -update-counts", err)
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
			measured := countMeasured(t, each.key)
			if !probing {
				regenerateCountRecord(t, "runtime", each.key, measured)
			}
			if measured == "" {
				t.Fatalf("%s holds no measurement of %s", fmt.Sprintf("count-measurements shard %d", familyShard(each.key, countProductShards)), each.key)
			}
			if measured != want {
				t.Errorf("the counts changed and %s wasn't updated with them (go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts):\nrecorded: %s\nmeasured: %s",
					countFile("runtime", each.key), want, measured)
			}
			if countSampled(each.key) || probing && each.key == planted {
				_, measure := each.prepare(t)
				if row := measure(); row != want {
					t.Errorf("the counts changed and %s wasn't updated with them (go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts):\nrecorded: %s\nmeasured here: %s",
						countFile("runtime", each.key), want, row)
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
func countShardsAhead(t *testing.T, shard int, keys []string, indices []int) {
	t.Helper()
	read := map[int]bool{}
	for _, index := range indices {
		read[familyShard(keys[index], countProductShards)] = true
	}
	for offset := range countProductShards {
		if product := (shard + offset) % countProductShards; read[product] {
			countShard(t, product)
		}
	}
}

// The predicate records retain their own schema and independent measurements.
func TestCountsAreRecordedTable(t *testing.T) {
	t.Parallel()
	if *updateCounts {
		for key, row := range predicateCountsTable(t) {
			measuredPredicateCounts.Store(key, row)
		}
		return
	}
	keys, err := predicateCountKeys()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := readCountRows(countsDirectory, "predicates", keys); err != nil {
		t.Fatal(err)
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
			row, err := readCountRow(countsDirectory, "runtime", key)
			if err != nil {
				return err
			}
			rows[key] = row
		}
	}
	if measured == 0 {
		return fmt.Errorf("-update-counts measured no runtime record: select the shards with -run TestCountsAreRecorded, unanchored")
	}
	keys, err := predicateCountKeys()
	if err != nil {
		return err
	}
	predicates := map[string]string{}
	for _, key := range keys {
		if row, ok := measuredPredicateCounts.Load(key); ok {
			predicates[key] = row.(string)
		} else {
			row, err := readCountRow(countsDirectory, "predicates", key)
			if err != nil {
				return err
			}
			predicates[key] = row
		}
	}
	return writeCountFiles(countsDirectory, rows, predicates)
}
