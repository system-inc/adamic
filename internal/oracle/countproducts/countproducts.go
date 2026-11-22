// Package countproducts measures internal/oracle's counts table: each row's emitted C built counted and run as its row
// runs, a shard of the table at a time, in a buildcache product internal/oracle's TestProduct_CountMeasurements tests
// settle and every row of TestCountsAreRecorded reads.
//
// Measured in oracle's own product, a shard was keyed by its window and oracle.test's whole recipe, the lowering with
// it (buildcache's leanCallers, settle.go's recipeFor), so every compiler landing rebuilt and ran again all 16 shards,
// about 1,700 core-seconds a tree, where 8 of 9 landings move no row's C (#a4frknq). So oracle keys and lowers each row
// as it did, and writes a shard's rows, each row's C and how it runs, as a product of its own (Rows); the measuring is
// here, a LeanProductOf build of that rows file, clang, the checkout and the counted runtime archives
// (BuildCountMeasurements). It's keyed by its own window and this package's import closure, which holds native, ir and
// flow, the code that turns C into the counted binary, and nothing of lower, javascript or oracle
// (TestTheMeasuringImportsNothingThatLowers).
package countproducts

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"syscall"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/ccjobserver"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/native_runtime"
	"github.com/system-inc/adamic/internal/toolchain"
)

// Harness is how a row is built and run beyond what it runs (Measure, pinnedStack): bump it when that changes, and every
// shard's key moves, with every row's digest internal/oracle makes of it (countDigest). v2: built lean (native.BuildLean)
// with the archive native.Build would link, here, where v1 built in oracle through native.Build. v3: every Plain and
// Input row runs with runEnvironment alone, where v2 inherited the process's environment, which no key names.
const Harness = "counts-harness-v3"

// A Run is how a row's counted binary runs.
type Run string

const (
	// Plain runs the binary alone, with an 8 MiB stack (pinnedStack), and one worker when its C calls the pool.
	Plain Run = "Plain"
	// Input runs it as internal/oracle's TestInputAgreesWithNode runs an input fixture: from its own directory, with its
	// arguments, a directory of its own to write in when it writes, and as nobody when the process is root.
	Input Run = "Input"
	// Environment runs it with exactly its row's variables beside the sanitizers' settings, nothing inherited.
	Environment Run = "Environment"
	// Realpath runs it in a directory laid out for node:fs's realpath (realpathLayout), with only the sanitizers'
	// settings.
	Realpath Run = "Realpath"
)

// A Row is one row of the counts table as its shard's product measures it: its key, the C it builds counted, and
// everything its run needs.
type Row struct {
	// Key is the row's key, the first column of its rows.tsv line and the path its counts line names.
	Key string
	// C is the row's emitted C, built with the counted runtime archive of the features it turns on.
	C   string
	Run Run
	// ParallelMap is a Plain row whose C calls the pool, run with one worker, which keeps peak liveness the same on
	// every machine.
	ParallelMap bool
	// Directory is an Input row's working directory, from the repository's root, and Arguments its arguments.
	Directory string
	Arguments []string
	// Unreadable, Writes and FSFile are an Input row's other doors: the path of a file it may not read as its last
	// argument, a directory of its own to write in as its first, and a node:fs file fixture's prepared directory in
	// place of both (fsFilePrepare).
	Unreadable, Writes, FSFile bool
	// Inputs is the digest of the trees an Input row reads as it runs, there only so they key its shard.
	Inputs string
	// Variables are an Environment row's environment.
	Variables []string
	// Checked has the run end with Exit, or fails the row.
	Checked bool
	Exit    int
}

// rowsVersion names how a rows file is written: changing it moves every shard's key with its bytes.
const rowsVersion = "count-rows-v1"

// repositoryMark stands in a rows file for the repository's root wherever a row's C names it, so a shard's rows are the
// same bytes in every checkout, as countPortable keeps a row's digest; a build puts back the root it's handed. No C a
// row emits holds a NUL, so the mark is never a C's own text (Rows refuses one that holds it).
const repositoryMark = "\x00repository\x00"

// A rowsFile is a shard's rows file: the feature sets its rows turn on, in the order a build is handed their archives,
// and its rows, sorted by key.
type rowsFile struct {
	Version string
	Sets    [][]string
	Rows    []fileRow
}

// A fileRow is a Row as its rows file holds it. What may hold any bytes is held as bytes, which encoding/json writes in
// base64: a string it writes coerced to valid UTF-8, and arguments.a's arguments aren't valid UTF-8, on purpose.
type fileRow struct {
	Key         string
	C           []byte
	Run         Run
	ParallelMap bool
	Directory   string
	Arguments   [][]byte
	Unreadable  bool
	Writes      bool
	FSFile      bool
	Inputs      string
	Variables   [][]byte
	Checked     bool
	Exit        int
}

// Rows is the rows file of a shard holding rows, from the checkout at repository, an absolute path, and the feature
// sets whose counted archives a build of it is handed, in order (Product). Its bytes are the same whatever order rows
// come in and wherever the checkout is: rows sorted by key, the sets by their names, and the checkout's root in a C as
// repositoryMark.
func Rows(rows []Row, repository string) ([]byte, [][]string, error) {
	if !filepath.IsAbs(repository) {
		return nil, nil, fmt.Errorf("countproducts: the repository %q isn't an absolute path", repository)
	}
	sorted := slices.Clone(rows)
	slices.SortFunc(sorted, func(first, second Row) int { return strings.Compare(first.Key, second.Key) })
	file := rowsFile{Version: rowsVersion}
	for index, row := range sorted {
		if index > 0 && sorted[index-1].Key == row.Key {
			return nil, nil, fmt.Errorf("countproducts: two rows keyed %s", row.Key)
		}
		if strings.Contains(row.C, repositoryMark) {
			return nil, nil, fmt.Errorf("countproducts: %s's C holds a NUL, as the rows file's mark for the repository does", row.Key)
		}
		features := native_runtime.Features(row.C)
		if !slices.ContainsFunc(file.Sets, func(set []string) bool { return slices.Equal(set, features) }) {
			file.Sets = append(file.Sets, features)
		}
		file.Rows = append(file.Rows, fileRow{
			Key: row.Key, C: []byte(strings.ReplaceAll(row.C, repository, repositoryMark)), Run: row.Run, ParallelMap: row.ParallelMap,
			Directory: row.Directory, Arguments: byteStrings(row.Arguments), Unreadable: row.Unreadable, Writes: row.Writes,
			FSFile: row.FSFile, Inputs: row.Inputs, Variables: byteStrings(row.Variables), Checked: row.Checked, Exit: row.Exit,
		})
	}
	slices.SortFunc(file.Sets, func(first, second []string) int {
		return strings.Compare(strings.Join(first, " "), strings.Join(second, " "))
	})
	encoded, err := json.Marshal(file)
	if err != nil {
		return nil, nil, err
	}
	return append(encoded, '\n'), file.Sets, nil
}

// readRows is the rows and the feature sets a rows file holds, each C with repository, the checkout it's read in, put
// back in place of repositoryMark.
func readRows(contents []byte, repository string) ([]Row, [][]string, error) {
	var file rowsFile
	if err := json.Unmarshal(contents, &file); err != nil {
		return nil, nil, fmt.Errorf("countproducts: a rows file: %w", err)
	}
	if file.Version != rowsVersion {
		return nil, nil, fmt.Errorf("countproducts: a rows file written as %q, read as %q", file.Version, rowsVersion)
	}
	rows := make([]Row, len(file.Rows))
	for index, row := range file.Rows {
		rows[index] = Row{
			Key: row.Key, C: strings.ReplaceAll(string(row.C), repositoryMark, repository), Run: row.Run, ParallelMap: row.ParallelMap,
			Directory: row.Directory, Arguments: textStrings(row.Arguments), Unreadable: row.Unreadable, Writes: row.Writes,
			FSFile: row.FSFile, Inputs: row.Inputs, Variables: textStrings(row.Variables), Checked: row.Checked, Exit: row.Exit,
		}
	}
	return rows, file.Sets, nil
}

// byteStrings are values as bytes, nil for none.
func byteStrings(values []string) [][]byte {
	if values == nil {
		return nil
	}
	held := make([][]byte, len(values))
	for index, value := range values {
		held[index] = []byte(value)
	}
	return held
}

// textStrings are held as strings, nil for none, byteStrings undone.
func textStrings(held [][]byte) []string {
	if held == nil {
		return nil
	}
	values := make([]string, len(held))
	for index, value := range held {
		values[index] = string(value)
	}
	return values
}

// Product is the directory holding rows.tsv, the measured row of every row the rows file at rows holds, a line each,
// "key\t| key | allocations | frees | retains | releases | peak | regions |", in the order the file holds them.
// repository is the checkout an Input row runs in, an absolute path, and archives the directories of the counted runtime
// archives of the file's feature sets, in the file's order (Rows), each as native_runtime.RuntimeRequest.Build fills one.
//
// clang is found here, outside the build's window, as a counted native.Build finds it, keyed by its report under the
// counted build's triple (buildcache.Tool, which names what it is and never where) and handed by its absolute path, as
// agreementproducts hands it: a handed path never keys, so the key is the same on every machine running the same clang.
// The rows file, the checkout and the archives are handed alike: the window values what it reads there by content, a
// product by its bytes.
func Product(t testing.TB, rows string, repository string, archives []string) string {
	t.Helper()
	// A rows file a test wrote outside every product is a read under the temporary directory, which a traced window can
	// only refuse as temp not made by this run (agreementproducts' C, #egjbcre): refused here first, on every machine.
	if !strings.HasPrefix(buildcache.Relative(rows), "<build cache>/") {
		t.Fatalf("countproducts: the rows file %s isn't a product's: hand a shard's rows as internal/oracle's countRowsFile writes them", rows)
	}
	contents, err := os.ReadFile(rows)
	if err != nil {
		t.Fatal(err)
	}
	_, sets, err := readRows(contents, repository)
	if err != nil {
		t.Fatalf("countproducts: %s: %v", rows, err)
	}
	if len(archives) != len(sets) {
		t.Fatalf("countproducts: %s turns on %d feature sets %v, and %d archives are handed", rows, len(sets), sets, len(archives))
	}
	compiler, err := toolchain.Path(native_runtime.CompilerName(options()))
	if err != nil {
		t.Fatal(err)
	}
	keyed, err := keyingOf(contents, compiler)
	if err != nil {
		t.Fatal(err)
	}
	paths := append([]string{rows, compiler, repository}, archives...)
	return buildcache.LeanProductOf(t, buildcache.Handed{Inputs: keyed.inputs(), Paths: paths}, BuildCountMeasurements)
}

// A keying is what a shard's key names besides the code that measures it, each apart, so a test can move one alone and
// key the shard without building it (buildcache.Key).
type keying struct {
	// rows is the sha256 of the shard's rows file: every row's key and C, and how it runs.
	rows string
	// harness is Harness.
	harness string
	// runtime is the runtime's sources (native_runtime.RuntimeSourceDigest), which every counted archive is built from.
	runtime string
	// flags are the counted build's flags (native_runtime.Flags), which name its target.
	flags []string
	// compiler is clang's report under that target.
	compiler string
	// platform is the build's goos and goarch, as native's builds read them (toolchain.Platform of the environment).
	platform []string
	// temporary is the temporary directory a counted run is handed (runEnvironment's TMPDIR): its spelling and its
	// directory's mode, which node:os's tmpdir and a stat of it read, or that it's unset.
	temporary string
}

// keyingOf is the keying of the shard whose rows file holds rows, measured with compiler, an absolute path, on this
// machine.
func keyingOf(rows []byte, compiler string) (keying, error) {
	resolved, err := native_runtime.ResolveTarget(options())
	if err != nil {
		return keying{}, err
	}
	sources, err := native_runtime.RuntimeSourceDigest()
	if err != nil {
		return keying{}, err
	}
	goos, goarch := toolchain.Platform(os.Environ())
	temporary, err := temporaryOf()
	if err != nil {
		return keying{}, err
	}
	return keying{
		rows:      fmt.Sprintf("%x", sha256.Sum256(rows)),
		harness:   Harness,
		runtime:   sources,
		flags:     native_runtime.Flags(options()),
		compiler:  buildcache.Tool(compiler, "--target="+resolved.Target, "--version"),
		platform:  []string{goos, goarch},
		temporary: temporary,
	}, nil
}

// temporaryOf is the temporary directory a counted run is handed, as a shard's key names it: TMPDIR's spelling and its
// directory's mode, sticky bit included, or that it's unset (runEnvironment).
func temporaryOf() (string, error) {
	directory := os.Getenv("TMPDIR")
	if directory == "" {
		return "TMPDIR unset", nil
	}
	info, err := os.Stat(directory)
	if err != nil {
		return "", fmt.Errorf("countproducts: TMPDIR %s: %w", directory, err)
	}
	return fmt.Sprintf("TMPDIR=%s %v", directory, info.Mode()&(os.ModePerm|os.ModeSticky|os.ModeDir)), nil
}

// inputs key a shard's product: what keyed names, and as files native's that native.BuildLean reaches (nativeFiles) and
// this package's own (declaredFiles), which an untraced build keys by content and a traced one by its recipe as well.
func (keyed keying) inputs() buildcache.Inputs {
	return buildcache.Inputs{
		// v2: measured lean, here, from a rows file internal/oracle keys (#a4frknq); v1 was oracle's own product.
		Name:      "count-measurements v2",
		Files:     append(nativeFiles(), declaredFiles()...),
		Flags:     append([]string{"rows " + keyed.rows, "harness " + keyed.harness, "runtime " + keyed.runtime, "temporary " + keyed.temporary}, keyed.flags...),
		Toolchain: append(slices.Clone(keyed.platform), keyed.compiler),
	}
}

// nativeFiles are internal/native's and internal/native_runtime's files defining everything native.BuildLean reaches,
// agreementproducts.NativeFiles, which runtimeproducts' census holds BuildLean to. They're spelled here rather than
// imported, since a count's recipe keys every package this package imports, and agreementproducts' programs move with
// internal/lower's agreement rows; TestTheNativeFilesAreAgreementProductsOwn holds the two lists to one.
func nativeFiles() []string {
	return []string{"internal/native/lean_build.go", "internal/native/native.go", "internal/native/regexp_compiler_library.go", "internal/native_runtime/jobserver.go", "internal/native_runtime/library.go", "internal/native_runtime/options.go", "internal/native_runtime/target.go"}
}

// declaredFiles are this package's files, which its build reaches and its key reads as files; its tests are apart, so
// an edit of one moves no shard (TestTheCountBuildReachesOnlyItsDeclaredFiles holds the build to these).
func declaredFiles() []string {
	return []string{"internal/oracle/countproducts/countproducts.go", "internal/oracle/countproducts/run.go"}
}

// countWorkers is how many rows a shard's build measures at once: the CPUs a product's unit is given, as runtimeproducts'
// Build compiles four units at once. Each clang holds a token of the process's jobserver as well.
const countWorkers = 4

// BuildCountMeasurements measures every row of the rows file it's handed into directory's rows.tsv (Product). It's
// handed the rows file, clang, the checkout and the counted archives, in that order, and is a top-level function of
// what it's handed alone (buildcache.LeanProductOf), but for one ambient handle: the process's compile jobserver
// (ccjobserver.Process), as agreementproducts' BuildAgreementNative takes it, which decides when clang runs and never
// what it writes. Rows run countWorkers at a time, and every worker is done before it returns, so nothing it starts
// reads past its window.
func BuildCountMeasurements(handed buildcache.Handed, directory string) error {
	if len(handed.Paths) < 3 {
		return fmt.Errorf("countproducts: handed %q, want the rows file, clang, the repository and an archive for each feature set", handed.Paths)
	}
	contents, err := os.ReadFile(handed.Paths[0])
	if err != nil {
		return err
	}
	rows, sets, err := readRows(contents, handed.Paths[2])
	if err != nil {
		return err
	}
	archives := handed.Paths[3:]
	if len(archives) != len(sets) {
		return fmt.Errorf("countproducts: the rows file turns on %d feature sets %v, and %d archives are handed", len(sets), sets, len(archives))
	}
	lines := make([]string, len(rows))
	failures := make([]error, len(rows))
	next := make(chan int)
	var workers sync.WaitGroup
	for range countWorkers {
		workers.Go(func() {
			for index := range next {
				archive, err := archiveFor(sets, archives, rows[index].C)
				if err == nil {
					lines[index], err = Measure(rows[index], handed.Paths[1], handed.Paths[2], archive)
				}
				failures[index] = err
			}
		})
	}
	for index := range rows {
		next <- index
	}
	close(next)
	workers.Wait()
	if err := errors.Join(failures...); err != nil {
		return err
	}
	var table strings.Builder
	for index, row := range rows {
		fmt.Fprintf(&table, "%s\t%s\n", row.Key, lines[index])
	}
	return os.WriteFile(filepath.Join(directory, "rows.tsv"), []byte(table.String()), 0o644)
}

// archiveFor is the archive of archives, in the order of sets, built for the features code turns on.
func archiveFor(sets [][]string, archives []string, code string) (string, error) {
	features := native_runtime.Features(code)
	index := slices.IndexFunc(sets, func(set []string) bool { return slices.Equal(set, features) })
	if index < 0 {
		return "", fmt.Errorf("countproducts: the rows file holds no archive for the features %v", features)
	}
	return archives[index], nil
}

// Measure builds row's C counted with compiler, an absolute path, linked with the counted runtime archive in archive's
// directory, runs it as its row runs, an Input row in the checkout at repository, and returns its row of the table:
// "| key | allocations | frees | retains | releases | peak | regions |". A shard's build measures each of its rows so,
// and internal/oracle's tests measure one row so in their own process.
func Measure(row Row, compiler string, repository string, archive string) (string, error) {
	switch row.Run {
	case Plain, Input, Environment, Realpath:
	default:
		return "", fmt.Errorf("countproducts: %s runs as %q, which no row runs as", row.Key, row.Run)
	}
	scratch, err := os.MkdirTemp("", "adamic-count-")
	if err != nil {
		return "", err
	}
	result, err := measured(row, compiler, repository, archive, scratch)
	if removed := os.RemoveAll(scratch); removed != nil && err == nil {
		err = fmt.Errorf("its scratch directory stays behind: %w", removed)
	}
	if err != nil {
		return "", fmt.Errorf("countproducts: %s: %w", row.Key, err)
	}
	if row.Checked && result.exitCode != row.Exit {
		return "", fmt.Errorf("countproducts: %s: want its counted run to exit %d, as it does on Node: exit %d, stderr %q", row.Key, row.Exit, result.exitCode, result.stderr)
	}
	match := leakcheck.CountsLine.FindSubmatch(result.stderr)
	if match == nil {
		return "", fmt.Errorf("countproducts: %s: no counts at the end of stderr (exit %d): %q", row.Key, result.exitCode, result.stderr)
	}
	return fmt.Sprintf("| %s | %s | %s | %s | %s | %s | %s |", row.Key, match[1], match[2], match[3], match[4], match[5], match[6]), nil
}

// measured builds row's C, written into scratch, a directory of its own, and runs it as its row runs.
func measured(row Row, compiler string, repository string, archive string, scratch string) (run, error) {
	source := filepath.Join(scratch, "main.c")
	if err := os.WriteFile(source, []byte(row.C), 0o644); err != nil {
		return run{}, err
	}
	if row.Run == Input {
		return measuredInput(row, compiler, repository, archive, source)
	}
	binary := filepath.Join(scratch, "program")
	if err := buildCounted(compiler, source, archive, binary); err != nil {
		return run{}, err
	}
	switch row.Run {
	case Environment:
		return environmentRun(row.Variables, binary)
	case Realpath:
		directory := filepath.Join(scratch, "realpath")
		if err := realpathLayout(directory); err != nil {
			return run{}, err
		}
		return realpathRun(directory, binary)
	}
	name, pinned := pinnedStack(binary)
	if row.ParallelMap {
		// One worker keeps peak liveness reproducible across machines.
		return executeParallel("1", name, pinned...)
	}
	return execute(name, pinned...)
}

// measuredInput builds an Input row's C at source into a directory shared with whoever it runs as, and runs it there as
// its row runs. It must finish, as it must on Node, so its row counts the whole program and not where it stopped
// (Checked, which its row sets).
func measuredInput(row Row, compiler string, repository string, archive string, source string) (result run, err error) {
	shared, err := sharedDirectory()
	if err != nil {
		return run{}, err
	}
	defer func() {
		if removed := os.RemoveAll(shared); removed != nil && err == nil {
			err = fmt.Errorf("the shared directory stays behind: %w", removed)
		}
	}()
	how := inputRun{directory: filepath.Join(repository, filepath.FromSlash(row.Directory)), arguments: row.Arguments}
	if os.Geteuid() == 0 {
		how.credential = &syscall.Credential{Uid: 65534, Gid: 65534}
	}
	if row.Unreadable {
		path := filepath.Join(shared, "unreadable.txt")
		if err := os.WriteFile(path, []byte("secret\n"), 0o000); err != nil {
			return run{}, err
		}
		how.arguments = append(append([]string{}, how.arguments...), path)
	}
	if row.Writes {
		written, err := writable(shared, "counted")
		if err != nil {
			return run{}, err
		}
		how.arguments = append([]string{written}, how.arguments...)
	}
	if row.FSFile {
		var prepared inputRun
		var restore func() error
		prepared, restore, err = fsFilePrepare(shared, "counted")
		if err != nil {
			return run{}, err
		}
		// Deferred after the shared directory's removal, so it runs first: locked and closed keep their files from
		// RemoveAll until they get their permissions back.
		defer func() {
			if restored := restore(); restored != nil && err == nil {
				err = restored
			}
		}()
		how = prepared
	}
	binary := filepath.Join(shared, "program")
	if err := buildCounted(compiler, source, archive, binary); err != nil {
		return run{}, err
	}
	name, pinned := pinnedStack(binary, how.arguments...)
	return executeInput(how, name, pinned...)
}

// buildCounted compiles the C at source counted with compiler and links it with the archive in archive's directory into
// output, as a counted native.Build would with the archive it would link (native.BuildLean, which reads nothing of the
// process but its platform).
func buildCounted(compiler string, source string, archive string, output string) error {
	jobserver, err := ccjobserver.Process()
	if err != nil {
		return err
	}
	return native.BuildLean(compiler, source, archive, output, options(), jobserver)
}

// options are what every row is built with: counted, for the build's own native triple.
func options() native_runtime.Options {
	return native_runtime.Options{Count: true}
}
