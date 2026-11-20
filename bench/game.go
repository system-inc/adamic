// Game measures Adamic against the Computer Language Benchmarks Game's own Go and Rust entries,
// vendored unchanged in bench/game, all four languages on one box.
//
//	go run ./bench -game -go <Go 1.23.1's go> -rustc <rustc 1.84.1> -crates <rust-deps' target/x86_64-unknown-linux-gnu/release/deps> [-pcre <prefix>] [-samples 15] [-only fasta,nbody] [-seed n]
//
// bench/game/programs.json names each program's runners: Adamic native, built by this checkout's
// adamic from the holdout port or the everyday source, Node running the same source, and the game's
// entries, each built as the game built it, with the game's toolchains, which are taken as flags and
// never guessed: Go 1.23.1 with GOAMD64=v2 and a bare go build, and rustc 1.84.1 with the game's
// flags, linked with --extern against the rlibs bench/game/rust-deps builds. The two regexredux
// entries link PCRE: with -pcre, the prefix libpcre and libpcre2 are unpacked under, their builds get
// its headers and libraries and their runs LD_LIBRARY_PATH at them; without it, the system's. Every
// build is a named step, bounded as compare.go bounds its steps; one that stops is a gap with the
// compiler's line.
//
// Every runner's output is checked against Node's byte for byte, at the row's size, before any time
// counts, and again on every timed run: a runner whose output isn't Node's is a gap, never a timing.
// A program that reads fasta's output gets the file compare.go's makeInput makes, as the first
// argument for Adamic and Node, as the holdout ports read it, and on standard input for Go and Rust,
// with 0 as their argument, as the game runs them.
//
// Each sample runs the four floors, a program that only prints in each language, then each
// program's runners, in an order rotated one place every two samples and reversed on odd samples,
// so no runner always follows the same one. Every native binary gets fresh pages before each sample
// and every run an ADAMIC_BENCH_PAD of random length, as in compare.go. Each runner's wall, user and
// system time and peak resident set are each its own best of -samples runs, and a best within its
// language's floor is refused, as run.go refuses one. When perf can count, instructions:u come from
// perf stat runs made after the timed ones, each runner's median of three: perf's own start and
// resident set would land in a timed run's wall time and peak.
//
// The report is one table per basis: the game's fastest entry, its fastest entry that runs one
// thread, and, where the fastest needs an instruction set beyond the baseline, its fastest portable
// one. Every row says the basis it stands on and whether it's like for like, and the headline counts
// only one thread against one thread: Adamic's programs here run one thread, so an entry that runs
// several is reported and never headlined. The game's own elapsed time appears as a labeled sanity
// column, from its i5-3330 at its n, never beside ours as a comparison. Losses are listed beside the
// wins, then every gap. The rows are reported and never tuned against, as the holdout's rule says.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

// A gameRequest is what -game was asked.
type gameRequest struct {
	goBinary string // the go that builds the Go entries, as an absolute path
	rustc    string // the rustc that builds the Rust entries, as an absolute path
	crates   string // the directory of rlibs the Rust entries link against, as an absolute path
	// pcre is the prefix libpcre and libpcre2 are unpacked under (dpkg -x), as an absolute path, or
	// empty, when the regexredux entries build and run against the system's.
	pcre    string
	samples int
	only    string
	timeout time.Duration
	seed    uint64 // ADAMIC_BENCH_PAD's lengths are drawn from it
}

// gameFile is bench/game/programs.json.
type gameFile struct {
	Source gameSource `json:"source"`
	Go     gameGo     `json:"go"`
	Rust   gameRust   `json:"rust"`
	// Crates is each crate a Rust entry links against, by its package name, to the exact version
	// bench/game/rust-deps/Cargo.toml pins it at.
	Crates   map[string]string `json:"crates"`
	Programs []gameProgram     `json:"programs"`
	// Absent is every program bench/ measures that the game has no entry for, and Refused every game
	// program Adamic refuses; both are listed among the gaps and never timed.
	Absent  []gameAbsence `json:"absent"`
	Refused []gameAbsence `json:"refused"`
	// Vendored is every file in bench/game that came from the game's archive, with its sha256 there.
	Vendored []vendoredFile `json:"vendored"`
}

// gameSource is where the vendored files and the game's figures came from.
type gameSource struct {
	Archive      string `json:"archive"`
	Blob         string `json:"blob"`
	Sha256       string `json:"sha256"`
	Measurements string `json:"measurements"`
	Repository   string `json:"repository"`
	Commit       string `json:"commit"`
	GameMachine  string `json:"gameMachine"`
}

// gameGo is how the game built its Go entries: go build with this environment, at this version.
type gameGo struct {
	Version     string   `json:"version"`
	Environment []string `json:"environment"`
}

// gameRust is how the game built its Rust entries: rustc with these flags, at this version.
type gameRust struct {
	Version string   `json:"version"`
	Flags   []string `json:"flags"`
}

// A gameProgram is one of the game's programs as measured here.
type gameProgram struct {
	Name string `json:"name"`
	// Adamic is the source Adamic and Node run, a bench/ program named without its extension.
	Adamic string `json:"adamic"`
	// Size is n, every Go and Rust entry's first argument; 0 for a program that reads Input on
	// standard input, as the game passes it.
	Size  int           `json:"size"`
	Input *programInput `json:"input"`
	// GameSize is the n of the game's own figures, and Ours, when set, why ours differs from it.
	GameSize int         `json:"gameSize"`
	Ours     string      `json:"ours"`
	Entries  []gameEntry `json:"entries"`
}

// A gameEntry is one of the game's programs in Go or Rust, vendored unchanged.
type gameEntry struct {
	File     string `json:"file"` // under bench/game, as the archive names it
	Language string `json:"language"`
	// Bases is every basis the entry stands on for its language: fastest, one thread, portable.
	Bases []string `json:"bases"`
	// Threads is one or several: how many the entry keeps busy, read from the game's CPU time over
	// its elapsed time.
	Threads        string `json:"threads"`
	InstructionSet string `json:"instructionSet"` // what it needs beyond the baseline, when anything
	// Pcre says the entry links PCRE, Go's through go-pcre and Rust's through pcre2-sys, so -pcre's
	// headers and libraries are handed to its build and its runs.
	Pcre bool `json:"pcre"`
	// Crates is every crate a Rust entry is linked against with --extern, by package name: those its
	// build line passes with --extern, and those its source names that the game found through -L.
	Crates    []string `json:"crates"`
	BuildLine string   `json:"buildLine"` // the game's, as its program page prints it
	// GameElapsed and GameCpu are the game's own figures, in seconds at its n, for the sanity column.
	GameElapsed float64 `json:"gameElapsed"`
	GameCpu     float64 `json:"gameCpu"`
}

// A gameAbsence is a program with no row, and why.
type gameAbsence struct {
	Name string `json:"name"`
	Why  string `json:"why"`
}

// A vendoredFile is a file under bench/game and its sha256 in the game's archive.
type vendoredFile struct {
	Path   string `json:"path"`
	Sha256 string `json:"sha256"`
}

// The bases an entry stands on. The fastest is the game's fastest entry for its language, mostly
// several threads; one thread is its fastest entry that runs one; portable is its fastest entry that
// needs no instruction set beyond the baseline, listed only where the fastest needs one.
const (
	fastestBasis   = "fastest"
	oneThreadBasis = "one thread"
	portableBasis  = "portable"
)

var gameBases = [...]string{fastestBasis, oneThreadBasis, portableBasis}

// The four languages a row runs in; each has its own floor.
const (
	adamicLanguage = "adamic"
	nodeLanguage   = "node"
	goLanguage     = "go"
	rustLanguage   = "rust"
)

var entryLanguages = [...]string{goLanguage, rustLanguage}

// How many threads an entry keeps busy.
const (
	oneThread      = "one"
	severalThreads = "several"
)

// A result's cells: Adamic, Node, then each entry in programs.json's order.
const (
	gameAdamic     = 0
	gameNode       = 1
	gameFirstEntry = 2
)

// instructionRounds is how many perf runs each runner gets once the timed runs are done; the median
// is reported.
const instructionRounds = 3

// A gameRunner is one way a row's program runs: Adamic native, Node, or one of the game's entries.
type gameRunner struct {
	name     string // adamic, node, or the entry's name in the game's words: go-2, rust-7
	language string // whose floor judges its runs
	command  []string
	input    string // a file on standard input, how the game hands a Go or Rust entry its input
	binary   string // the native binary, given fresh pages before each sample; none for Node
	// environment is what every run of it adds to the process's: LD_LIBRARY_PATH at -pcre's libraries
	// for an entry that links PCRE.
	environment []string
	failure     string // why it has no timing: its build stopped, its output isn't Node's, a run failed
	runs        []measurement
	// instructions holds each perf run's instructions:u, and instructionsFailure why one had none.
	instructions        []int64
	instructionsFailure string
}

// A gameRow is one program and its runners, Adamic, Node, then each entry in programs.json's order,
// with Node's output, which every other runner's is checked against.
type gameRow struct {
	benchmark gameProgram
	runners   []*gameRunner
	answer    string
}

// A gameResult is one program's measured cells, as the report reads them: Adamic, Node, then each
// entry in programs.json's order.
type gameResult struct {
	benchmark gameProgram
	cells     []gameCell
}

// A gameCell is one runner's figures, each its own best of the runs and judged against its floor, or
// the failure that makes it a gap, with its median instructions:u or why there are none.
type gameCell struct {
	name         string
	best         measurement
	instructions string
}

func measureGame(request gameRequest) {
	if request.samples < 1 {
		fail(fmt.Errorf("-samples must be at least 1, not %d", request.samples))
	}
	if request.goBinary == "" || request.rustc == "" {
		fail(errors.New("-game builds the game's entries with the game's toolchains and takes their paths, -go <Go 1.23.1's go> and -rustc <rustc 1.84.1>; neither is guessed"))
	}
	// The paths are made absolute before benchDirectory moves to the repository's root, so a relative
	// one still names what it named where -game was started.
	var err error
	if request.goBinary, err = absoluteTool(request.goBinary, "go"); err != nil {
		fail(err)
	}
	if request.rustc, err = absoluteTool(request.rustc, "rustc"); err != nil {
		fail(err)
	}
	if request.crates, err = absoluteDirectory(request.crates, "crates"); err != nil {
		fail(err)
	}
	if request.pcre, err = absoluteDirectory(request.pcre, "pcre"); err != nil {
		fail(err)
	}
	directory, err := benchDirectory()
	if err != nil {
		fail(err)
	}
	root := filepath.Dir(directory)
	file, err := loadGame(filepath.Join(directory, "game", "programs.json"))
	if err != nil {
		fail(err)
	}
	benchmarks, err := selectGamePrograms(file, request.only)
	if err != nil {
		fail(err)
	}
	for _, benchmark := range benchmarks {
		for _, entry := range benchmark.Entries {
			if len(entry.Crates) > 0 && request.crates == "" {
				fail(fmt.Errorf("%s's %s links against crates, so -game takes -crates <directory>: the rlibs `cargo build --release --target x86_64-unknown-linux-gnu` leaves in bench/game/rust-deps/target/x86_64-unknown-linux-gnu/release/deps, built with -rustc's toolchain", benchmark.Name, entryName(entry)))
			}
		}
	}
	goVersion, err := toolchainVersion(request.goBinary, "version", file.Go.Version)
	if err != nil {
		fail(err)
	}
	rustVersion, err := toolchainVersion(request.rustc, "--version", file.Rust.Version)
	if err != nil {
		fail(err)
	}
	work, err := os.MkdirTemp("", "adamic-game-")
	if err != nil {
		fail(err)
	}
	defer os.RemoveAll(work)
	// Every runner, Go and Rust's as much as Adamic's, is measured through the reaper, so each peak is its
	// own and never bench's, which here holds the inputs (J-wme5jp2 read 1321.4 MB for every runner).
	if reaper, err = buildReaper(directory, work); err != nil {
		fail(err)
	}

	describeMachine([]runtimeKind{{name: adamicLanguage}, {name: nodeLanguage}, {name: goLanguage}, {name: rustLanguage}})
	if commit, err := exec.Command("git", "rev-parse", "HEAD").Output(); err == nil {
		fmt.Printf("adamic: %s", commit)
	}
	fmt.Printf("go: %s, %s\n", request.goBinary, goVersion)
	fmt.Printf("rustc: %s, %s\n", request.rustc, rustVersion)
	if request.pcre != "" {
		fmt.Printf("pcre: %s, its headers and libraries handed to the entries that link PCRE, which run with LD_LIBRARY_PATH=%s\n", request.pcre, request.pcreLibraries())
	} else {
		fmt.Println("pcre: the system's, with no -pcre")
	}
	fmt.Printf("game: %s (blob %s, sha256 %s) at %s commit %s\n", file.Source.Archive, file.Source.Blob, file.Source.Sha256, file.Source.Repository, file.Source.Commit)
	fmt.Printf("programs: %d; samples: %d\n", len(benchmarks), request.samples)
	fmt.Printf("seed: %d, which draws every run's ADAMIC_BENCH_PAD length; -seed %d draws the same again\n", request.seed, request.seed)
	fmt.Printf("load before: %s\n", loadAverage())
	fmt.Printf("processes over 5%% CPU at start: %s\n\n", busyProcesses())

	adamic := filepath.Join(work, "adamic")
	build := exec.Command("go", "build", "-buildvcs=false", "-o", adamic, "./cmd/adamic")
	build.Dir = root
	if output, err := step("building adamic at this checkout", build); err != nil {
		fail(fmt.Errorf("building adamic: %v\n%s", err, output))
	}
	builder := gameBuilder{request: request, file: file, directory: directory, root: root, work: work, adamic: adamic}
	floors, err := builder.floors()
	if err != nil {
		fail(err)
	}
	rows := []*gameRow{}
	for _, benchmark := range benchmarks {
		row, err := builder.row(benchmark)
		if err != nil {
			fail(err)
		}
		rows = append(rows, row)
	}
	// Each input is made once, by the generator built with this checkout's adamic, before any run.
	inputs := map[string]string{}
	for _, row := range rows {
		if row.benchmark.Input == nil {
			continue
		}
		key := fmt.Sprintf("%s-%d", strings.ReplaceAll(row.benchmark.Input.Generator, "/", "_"), row.benchmark.Input.Size)
		path, made := inputs[key]
		if !made {
			path, err = makeInput(directory, work, adamic, *row.benchmark.Input, key)
			if err != nil {
				for _, runner := range row.runners {
					if runner.failure == "" {
						runner.failure = "it has no input: " + err.Error()
					}
				}
				continue
			}
			inputs[key] = path
			fmt.Printf("input %s: %s at size %d\n", path, row.benchmark.Input.Generator, row.benchmark.Input.Size)
		}
		row.attach(path)
	}
	for _, row := range rows {
		row.check(request.timeout)
	}

	random := rand.New(rand.NewPCG(request.seed, 0))
	for sample := 0; sample < request.samples; sample++ {
		for _, floor := range floors {
			if floor.binary != "" {
				if err := freshPages(floor.binary); err != nil {
					fail(err)
				}
			}
		}
		for _, row := range rows {
			for _, runner := range row.ready() {
				if runner.binary != "" {
					if err := freshPages(runner.binary); err != nil {
						fail(err)
					}
				}
			}
		}
		for _, position := range gameOrder(sample, len(floors)) {
			floor := floors[position]
			result := runFrom(floor.command, "", request.timeout, padded(random, nil)...)
			floor.runs = append(floor.runs, result)
			fmt.Fprintf(os.Stderr, "sample %d floor %s: %s\n", sample+1, floor.language, short(result))
		}
		for _, row := range rows {
			ready := row.ready()
			for _, position := range gameOrder(sample, len(ready)) {
				runner := ready[position]
				result := runFrom(runner.command, runner.input, request.timeout, padded(random, runner.environment)...)
				fmt.Fprintf(os.Stderr, "sample %d %s %s: %s\n", sample+1, row.benchmark.Name, runner.name, short(result))
				switch {
				case result.failure != "":
					runner.failure = fmt.Sprintf("in sample %d it %s", sample+1, result.failure)
				case result.stdout != row.answer:
					runner.failure = fmt.Sprintf("in sample %d its output wasn't Node's: %s", sample+1, outputDifference(result.stdout, row.answer))
				default:
					runner.runs = append(runner.runs, result)
				}
			}
		}
	}

	instructionsNote := countEveryRunner(work, floors, rows, request.timeout)

	resolution := clockResolution()
	floorSlowest := map[string]measurement{}
	floorCells := []string{}
	for _, floor := range floors {
		slowest := slowestOf(floor.runs)
		if slowest.failure != "" {
			fail(fmt.Errorf("the %s floor, a program that only prints, did not run, so no %s runner can be judged: %s", floor.language, floor.language, slowest.failure))
		}
		floorSlowest[floor.language] = slowest
		floorCells = append(floorCells, fmt.Sprintf("%s %s to %s", floor.language, seconds(bestOf(floor.runs).elapsed), seconds(slowest.elapsed)))
	}
	results := []gameResult{}
	for _, row := range rows {
		result := gameResult{benchmark: row.benchmark}
		for _, runner := range row.runners {
			cell := gameCell{name: runner.name, instructions: instructionsText(runner)}
			if runner.failure != "" {
				cell.best = measurement{failure: runner.failure}
			} else {
				cell.best = judge(bestOfEach(runner.runs), floorSlowest[runner.language], resolution)
			}
			result.cells = append(result.cells, cell)
		}
		results = append(results, result)
	}

	fmt.Printf("wall clock resolution: %s; floor, a program that only prints, fastest to slowest of every run: %s\n", resolution, strings.Join(floorCells, ", "))
	fmt.Printf("%d interleaved samples: each runs the four floors, then each program's runners, in an order rotated one place every two samples and reversed on odd samples, so no runner always follows the same one. Every native binary gets fresh pages before each sample, and every run an ADAMIC_BENCH_PAD of random length. Each figure is its own best of the runs: wall clock, user and system time across every thread, and peak resident set. Adamic over an entry is Adamic's best wall time over the entry's, below 1.00x where Adamic is faster. Every runner's output was checked against Node's byte for byte before any time counted, and on every run. An entry that runs several threads uses every CPU this box has, where the game's had four.\n", request.samples)
	fmt.Printf("instructions: %s\n\n", instructionsNote)
	printGameReport(os.Stdout, file, results)
	fmt.Printf("\nload after: %s\n", loadAverage())
}

// absoluteTool is a toolchain flag's path, found the way a shell finds it and made absolute.
func absoluteTool(name, flagName string) (string, error) {
	path, err := exec.LookPath(name)
	if err != nil {
		return "", fmt.Errorf("-%s %s: %w", flagName, name, err)
	}
	return filepath.Abs(path)
}

// absoluteDirectory is a directory flag's path made absolute, or empty when the flag wasn't given.
func absoluteDirectory(path, flagName string) (string, error) {
	if path == "" {
		return "", nil
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	information, err := os.Stat(absolute)
	if err != nil {
		return "", fmt.Errorf("-%s: %w", flagName, err)
	}
	if !information.IsDir() {
		return "", fmt.Errorf("-%s %s isn't a directory", flagName, absolute)
	}
	return absolute, nil
}

// pcreLibraries is the directory -pcre's libraries sit in, as dpkg -x unpacks them, or empty with no
// -pcre.
func (request gameRequest) pcreLibraries() string {
	if request.pcre == "" {
		return ""
	}
	return filepath.Join(request.pcre, "usr", "lib", "x86_64-linux-gnu")
}

// pcreBuild is what a Go build of an entry that links PCRE adds to its environment: pkg-config reads
// -pcre's pkgconfig directory in place of its default search (PKG_CONFIG_LIBDIR, so no host .pc file is
// ever read), with -pcre as its sysroot, since the .pc files dpkg -x unpacks still say /usr; and cgo's
// compiler and linker look in its headers and libraries outright too. cgo's own -O2 -g stay, which
// setting CGO_CFLAGS or CGO_LDFLAGS would otherwise drop. The pkg-config on PATH is the caller's (J-q175xd8
// found Golem had none). Nothing without -pcre.
func (request gameRequest) pcreBuild() []string {
	if request.pcre == "" {
		return nil
	}
	return []string{
		"PKG_CONFIG_LIBDIR=" + filepath.Join(request.pcreLibraries(), "pkgconfig"),
		"PKG_CONFIG_SYSROOT_DIR=" + request.pcre,
		"CGO_CFLAGS=-O2 -g -I" + filepath.Join(request.pcre, "usr", "include"),
		"CGO_LDFLAGS=-O2 -g -L" + request.pcreLibraries(),
	}
}

// pcreRun is what every run of an entry that links PCRE adds to its environment, so the dynamic
// loader finds -pcre's libraries: LD_LIBRARY_PATH, chosen over an rpath so each binary is linked as
// the game linked it. Nothing without -pcre.
func (request gameRequest) pcreRun() []string {
	if request.pcre == "" {
		return nil
	}
	return []string{"LD_LIBRARY_PATH=" + request.pcreLibraries()}
}

// toolchainVersion is a toolchain's first line of its version, and whether it's the game's. It's asked
// from the temporary directory with GOTOOLCHAIN=local, as the entries are built: from the repository,
// whose go.mod names a newer Go, -go's go version would switch to that one and report it.
func toolchainVersion(binary, versionFlag, want string) (string, error) {
	command := exec.Command(binary, versionFlag)
	command.Dir = os.TempDir()
	command.Env = append(environment(), "GOTOOLCHAIN=local")
	output, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s %s: %v %s", binary, versionFlag, err, firstLine(string(output)))
	}
	line := firstLine(string(output))
	if !strings.Contains(line+" ", " "+want+" ") {
		return line + ", not the game's " + want, nil
	}
	return line + ", the game's", nil
}

// loadGame reads programs.json and checks it holds together, so a mistake in it is named before
// anything is built.
func loadGame(path string) (gameFile, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return gameFile{}, err
	}
	var file gameFile
	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&file); err != nil {
		return gameFile{}, fmt.Errorf("%s: %w", path, err)
	}
	if err := file.validate(); err != nil {
		return gameFile{}, fmt.Errorf("%s: %w", path, err)
	}
	return file, nil
}

// validate checks what the runner and the report rely on: every entry is a vendored file of its own
// program's, in Go or Rust, on at most one basis per language, and an entry on the one thread basis
// runs one thread, so the headline can never count one that runs several.
func (file gameFile) validate() error {
	if file.Go.Version == "" || file.Rust.Version == "" || len(file.Rust.Flags) == 0 {
		return errors.New("go and rust must name the game's versions, and rust its flags")
	}
	listed := map[string]bool{}
	for _, vendored := range file.Vendored {
		if listed[vendored.Path] || len(vendored.Sha256) != 64 {
			return fmt.Errorf("vendored %s is listed twice or has no sha256", vendored.Path)
		}
		listed[vendored.Path] = true
	}
	names := map[string]bool{}
	for _, benchmark := range file.Programs {
		if benchmark.Name == "" || names[benchmark.Name] {
			return fmt.Errorf("program %q has no name or is listed twice", benchmark.Name)
		}
		names[benchmark.Name] = true
		if benchmark.Adamic == "" {
			return fmt.Errorf("%s names no source for Adamic and Node", benchmark.Name)
		}
		if (benchmark.Input == nil) == (benchmark.Size == 0) {
			return fmt.Errorf("%s needs exactly one of a size and an input", benchmark.Name)
		}
		if benchmark.Size < 0 || benchmark.GameSize < 1 {
			return fmt.Errorf("%s: its size and the game's must be positive", benchmark.Name)
		}
		taken := map[string]string{}
		for _, entry := range benchmark.Entries {
			if !listed[entry.File] || !strings.HasPrefix(entry.File, benchmark.Name+"/") {
				return fmt.Errorf("%s names %s, which isn't one of its vendored files", benchmark.Name, entry.File)
			}
			if entry.Language != goLanguage && entry.Language != rustLanguage {
				return fmt.Errorf("%s is in %q; an entry is in go or rust", entry.File, entry.Language)
			}
			if entry.Threads != oneThread && entry.Threads != severalThreads {
				return fmt.Errorf("%s runs %q threads; an entry runs one or several", entry.File, entry.Threads)
			}
			if len(entry.Bases) == 0 {
				return fmt.Errorf("%s stands on no basis", entry.File)
			}
			for _, basis := range entry.Bases {
				if !slices.Contains(gameBases[:], basis) {
					return fmt.Errorf("%s stands on %q; the bases are %s", entry.File, basis, strings.Join(gameBases[:], ", "))
				}
				key := entry.Language + " " + basis
				if previous, isTaken := taken[key]; isTaken {
					return fmt.Errorf("%s and %s are both %s's %s entry on the %s basis", previous, entry.File, benchmark.Name, entry.Language, basis)
				}
				taken[key] = entry.File
				if basis == oneThreadBasis && entry.Threads != oneThread {
					return fmt.Errorf("%s stands on the one thread basis and runs %s threads", entry.File, entry.Threads)
				}
			}
			if entry.Language == goLanguage && len(entry.Crates) > 0 {
				return fmt.Errorf("%s is Go and names crates", entry.File)
			}
			for _, crate := range entry.Crates {
				if _, pinned := file.Crates[crate]; !pinned {
					return fmt.Errorf("%s links against %s, which crates doesn't pin", entry.File, crate)
				}
			}
			if entry.BuildLine == "" || entry.GameElapsed <= 0 || entry.GameCpu <= 0 {
				return fmt.Errorf("%s needs the game's build line and its figures", entry.File)
			}
		}
	}
	for _, absence := range append(slices.Clone(file.Absent), file.Refused...) {
		if names[absence.Name] || absence.Why == "" {
			return fmt.Errorf("%s is listed as absent or refused with no why, or has a row as well", absence.Name)
		}
	}
	return nil
}

// selectGamePrograms is the programs -only names, or every one.
func selectGamePrograms(file gameFile, only string) ([]gameProgram, error) {
	if only == "" {
		return file.Programs, nil
	}
	selected := []gameProgram{}
	for _, name := range strings.Split(only, ",") {
		if name == "" {
			continue
		}
		index := slices.IndexFunc(file.Programs, func(benchmark gameProgram) bool { return benchmark.Name == name })
		if index >= 0 {
			selected = append(selected, file.Programs[index])
			continue
		}
		for _, absence := range append(slices.Clone(file.Absent), file.Refused...) {
			if absence.Name == name {
				return nil, fmt.Errorf("-only %s: %s has no row here: %s", name, name, absence.Why)
			}
		}
		known := []string{}
		for _, benchmark := range file.Programs {
			known = append(known, benchmark.Name)
		}
		return nil, fmt.Errorf("-only %s: no program %s in bench/game/programs.json; there are %s", name, name, strings.Join(known, ", "))
	}
	return selected, nil
}

// entryName is an entry's name in the game's words, its file's name without the program's and the
// language's extension: go-2 for fasta.go-2.go, go for fasta.go, rust for revcomp.rust.
func entryName(entry gameEntry) string {
	base := entry.File[strings.LastIndex(entry.File, "/")+1:]
	_, name, _ := strings.Cut(base, ".")
	if trimmed, found := strings.CutSuffix(name, "."+entry.Language); found {
		return trimmed
	}
	return name
}

// benchSource is a bench/ program's source, exactly one of <name>.ts and <name>.a.
func benchSource(directory, name string) (string, error) {
	sources := []string{}
	for _, extension := range []string{".ts", ".a"} {
		source := filepath.Join(directory, filepath.FromSlash(name)+extension)
		if _, err := os.Stat(source); err == nil {
			sources = append(sources, source)
		}
	}
	if len(sources) != 1 {
		return "", fmt.Errorf("%s must be exactly one of bench/%s.ts and bench/%s.a; found %d", name, name, name, len(sources))
	}
	return sources[0], nil
}

// A gameBuilder builds every runner's binary in the work directory, each a named step.
type gameBuilder struct {
	request   gameRequest
	file      gameFile
	directory string // bench/
	root      string
	work      string
	adamic    string
}

// floors builds the program that only prints in each language, the four runners' floors.
func (builder gameBuilder) floors() ([]*gameRunner, error) {
	floorDirectory := filepath.Join(builder.work, "floor")
	if err := os.MkdirAll(floorDirectory, 0755); err != nil {
		return nil, err
	}
	sources := [][2]string{
		{"floor.ts", "console.log(\"0\");\n"},
		{"floor.go", "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"0\")\n}\n"},
		{"floor.rs", "fn main() {\n    println!(\"0\");\n}\n"},
	}
	for _, source := range sources {
		if err := os.WriteFile(filepath.Join(floorDirectory, source[0]), []byte(source[1]), 0644); err != nil {
			return nil, err
		}
	}
	adamicFloor := filepath.Join(floorDirectory, "adamic-floor")
	if output, err := step("building the floor with adamic", exec.Command(builder.adamic, "build", filepath.Join(floorDirectory, "floor.ts"), "-o", adamicFloor)); err != nil {
		return nil, fmt.Errorf("adamic build of the floor: %v %s", err, firstLine(string(output)))
	}
	goFloor := filepath.Join(floorDirectory, "go-floor")
	if err := builder.buildGo("the floor", floorDirectory, "floor.go", goFloor, false); err != nil {
		return nil, err
	}
	rustFloor := filepath.Join(floorDirectory, "rust-floor")
	if err := builder.buildRust("the floor", floorDirectory, "floor.rs", rustFloor, nil, false); err != nil {
		return nil, err
	}
	return []*gameRunner{
		{name: "floor", language: adamicLanguage, command: []string{adamicFloor}, binary: adamicFloor},
		{name: "floor", language: nodeLanguage, command: nodeCommand(program{name: "floor", source: filepath.Join(floorDirectory, "floor.ts")}, builder.root)},
		{name: "floor", language: goLanguage, command: []string{goFloor}, binary: goFloor},
		{name: "floor", language: rustLanguage, command: []string{rustFloor}, binary: rustFloor},
	}, nil
}

// row builds one program's runners. A build that stops makes its runner a gap, with the compiler's
// line; only the work directory failing stops the measurement.
func (builder gameBuilder) row(benchmark gameProgram) (*gameRow, error) {
	binaries := filepath.Join(builder.work, "bin")
	if err := os.MkdirAll(binaries, 0755); err != nil {
		return nil, err
	}
	row := &gameRow{benchmark: benchmark}
	native := &gameRunner{name: adamicLanguage, language: adamicLanguage}
	node := &gameRunner{name: nodeLanguage, language: nodeLanguage}
	source, err := benchSource(builder.directory, benchmark.Adamic)
	if err != nil {
		native.failure, node.failure = err.Error(), err.Error()
	} else {
		native.binary = filepath.Join(binaries, benchmark.Name+".adamic")
		native.command = []string{native.binary}
		if output, err := step("building "+benchmark.Name+" with adamic", exec.Command(builder.adamic, "build", source, "-o", native.binary)); err != nil {
			native.failure = fmt.Sprintf("adamic build stopped: %v %s", err, firstLine(string(output)))
		}
		contents, err := os.ReadFile(source)
		if err != nil {
			return nil, err
		}
		node.command = nodeCommand(program{name: benchmark.Name, source: source, adamic: importsAdamic(string(contents))}, builder.root)
	}
	row.runners = append(row.runners, native, node)
	for _, entry := range benchmark.Entries {
		name := entryName(entry)
		runner := &gameRunner{name: name, language: entry.Language, binary: filepath.Join(binaries, benchmark.Name+"."+name)}
		runner.command = []string{runner.binary, strconv.Itoa(benchmark.Size)}
		what := benchmark.Name + "'s " + name
		if entry.Pcre {
			runner.environment = builder.request.pcreRun()
		}
		var stopped error
		switch entry.Language {
		case goLanguage:
			stopped = builder.buildGo(what, filepath.Join(builder.directory, "game"), filepath.FromSlash(entry.File), runner.binary, entry.Pcre)
		case rustLanguage:
			buildDirectory, err := builder.rustSource(benchmark.Name, name, entry.File)
			if err != nil {
				return nil, err
			}
			stopped = builder.buildRust(what, buildDirectory, benchmark.Name+".rs", runner.binary, entry.Crates, entry.Pcre)
		}
		if stopped != nil {
			runner.failure = stopped.Error()
		}
		row.runners = append(row.runners, runner)
	}
	return row, nil
}

// rustSource copies a Rust entry to <program>.rs in a directory of its own, as the game copies it:
// rustc names the crate after its file, and fasta.rust-7 isn't a crate's name.
func (builder gameBuilder) rustSource(benchmarkName, name, file string) (string, error) {
	buildDirectory := filepath.Join(builder.work, "rust", benchmarkName+"-"+name)
	if err := os.MkdirAll(buildDirectory, 0755); err != nil {
		return "", err
	}
	contents, err := os.ReadFile(filepath.Join(builder.directory, "game", filepath.FromSlash(file)))
	if err != nil {
		return "", err
	}
	return buildDirectory, os.WriteFile(filepath.Join(buildDirectory, benchmarkName+".rs"), contents, 0644)
}

// buildGo builds a Go source as the game did, a bare go build with GOAMD64=v2, holding the toolchain
// to -go's own (GOTOOLCHAIN=local) and keeping the repository's workspace and any GOFLAGS out of it.
// A source that links PCRE gets -pcre's headers and libraries (pcreBuild).
func (builder gameBuilder) buildGo(what, directory, source, binary string, pcre bool) error {
	command := exec.Command(builder.request.goBinary, "build", "-o", binary, source)
	command.Dir = directory
	command.Env = append(append(environment(), builder.file.Go.Environment...), "GOTOOLCHAIN=local", "GOWORK=off", "GOFLAGS=")
	if pcre {
		command.Env = append(command.Env, builder.request.pcreBuild()...)
	}
	if output, err := step("building "+what+" with "+builder.request.goBinary, command); err != nil {
		printStopped(what, output)
		return fmt.Errorf("go build stopped: %v %s", err, stopLine(string(output)))
	}
	return nil
}

// buildRust builds a Rust source as the game did, with its flags, -L at the crates' directory, and
// --extern for each crate at its pinned version. A source that links PCRE gets -L native= at -pcre's
// libraries too, where the linker finds the libpcre2 pcre2-sys names when it was built against them.
func (builder gameBuilder) buildRust(what, directory, source, binary string, crates []string, pcre bool) error {
	arguments := slices.Clone(builder.file.Rust.Flags)
	if builder.request.crates != "" {
		arguments = append(arguments, "-L", builder.request.crates)
	}
	if pcre && builder.request.pcre != "" {
		arguments = append(arguments, "-L", "native="+builder.request.pcreLibraries())
	}
	for _, crate := range crates {
		library, err := crateLibrary(builder.request.crates, crate, builder.file.Crates[crate])
		if err != nil {
			return err
		}
		arguments = append(arguments, "--extern", strings.ReplaceAll(crate, "-", "_")+"="+library)
	}
	arguments = append(arguments, source, "-o", binary)
	command := exec.Command(builder.request.rustc, arguments...)
	command.Dir = directory
	if output, err := step("building "+what+" with "+builder.request.rustc, command); err != nil {
		printStopped(what, output)
		return fmt.Errorf("rustc stopped: %v %s", err, stopLine(string(output)))
	}
	return nil
}

// crateLibrary is the rlib cargo built for a crate at its pinned version in the crates' directory. A
// crate's rlib is lib<name>-<hash>.rlib, and two versions of one crate are two such files, so they're
// told apart by the dependency file cargo writes beside each, <name>-<hash>.d, which lists the crate's
// sources under the registry's <package>-<version>/ directory.
func crateLibrary(directory, crate, version string) (string, error) {
	library := strings.ReplaceAll(crate, "-", "_")
	candidates, err := filepath.Glob(filepath.Join(directory, "lib"+library+"-*.rlib"))
	if err != nil {
		return "", err
	}
	marker := "/" + crate + "-" + version + "/"
	found := []string{}
	for _, candidate := range candidates {
		hash := strings.TrimSuffix(strings.TrimPrefix(filepath.Base(candidate), "lib"+library+"-"), ".rlib")
		dependencies, err := os.ReadFile(filepath.Join(directory, library+"-"+hash+".d"))
		if err != nil {
			continue
		}
		if strings.Contains(string(dependencies), marker) {
			found = append(found, candidate)
		}
	}
	if len(found) != 1 {
		return "", fmt.Errorf("%s %s: %d of the %d lib%s-*.rlib in %s are that version by the .d file beside each; want exactly one, as cargo build --release in bench/game/rust-deps leaves them", crate, version, len(found), len(candidates), library, directory)
	}
	return found[0], nil
}

// printStopped prints a stopped build's whole output to the run's log, where the gap below the tables
// quotes one line of it: a toolchain's first line is often a progress line ("go: downloading ...") and
// the error that stopped it comes after (J-wme5jp2, regexredux go-5).
func printStopped(what string, output []byte) {
	fmt.Printf("bench: %s stopped, its whole output:\n%s\n", what, strings.TrimRight(string(output), "\n"))
}

// stopLine is the line of a build's output that says why it stopped: the first that begins with
// "error" when there is one, since rustc prints its warnings first, and otherwise the first that isn't
// a "#" heading, which go build prints above its errors, or go's "go: downloading" progress line.
func stopLine(output string) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "error") {
			return line
		}
	}
	for _, line := range lines {
		if line != "" && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "go: downloading ") {
			return line
		}
	}
	return ""
}

// attach gives a row's runners its input file: as the first argument to Adamic and Node, as the
// holdout ports read it, and on standard input to Go and Rust, whose argument is already 0, as the
// game runs them.
func (row *gameRow) attach(input string) {
	for _, runner := range row.runners {
		switch runner.language {
		case adamicLanguage, nodeLanguage:
			runner.command = append(runner.command, input)
		default:
			runner.input = input
		}
	}
}

// check runs Node once for the row's answer and every other runner once against it, untimed, before
// any sample. A runner whose output isn't Node's byte for byte is a gap, and with no answer from Node
// every runner is, since none can be checked.
func (row *gameRow) check(timeout time.Duration) {
	node := row.runners[gameNode]
	if node.failure == "" {
		fmt.Fprintf(os.Stderr, "bench: checking %s's outputs against Node's\n", row.benchmark.Name)
		reference := runFrom(node.command, node.input, timeout, node.environment...)
		if reference.failure != "" {
			node.failure = "checking its output, it " + reference.failure
		}
		row.answer = reference.stdout
	}
	if node.failure != "" {
		for index, runner := range row.runners {
			if index != gameNode && runner.failure == "" {
				runner.failure = "Node has no answer to check it against: " + node.failure
			}
		}
		return
	}
	for index, runner := range row.runners {
		if index == gameNode || runner.failure != "" {
			continue
		}
		result := runFrom(runner.command, runner.input, timeout, runner.environment...)
		switch {
		case result.failure != "":
			runner.failure = "checking its output, it " + result.failure
		case result.stdout != row.answer:
			runner.failure = "its output isn't Node's: " + outputDifference(result.stdout, row.answer)
		}
	}
}

// ready is the row's runners that still have no failure.
func (row *gameRow) ready() []*gameRunner {
	ready := []*gameRunner{}
	for _, runner := range row.runners {
		if runner.failure == "" {
			ready = append(ready, runner)
		}
	}
	return ready
}

// gameOrder is the order a sample runs count runners in: rotated one place every two samples, and
// reversed on odd samples. A rotation alone keeps who runs just before whom, so a runner would always
// follow the same other (the rotation compare.go gave up, M3, J-d1emr4h); reversed, each follows the
// one it led instead, so of three or more runners each follows two different ones, and over 2n
// samples each runs in every position twice.
func gameOrder(sample, count int) []int {
	order := make([]int, count)
	if count == 0 {
		return order
	}
	rotation := (sample / 2) % count
	for position := range order {
		order[position] = (rotation + position) % count
	}
	if sample%2 == 1 {
		slices.Reverse(order)
	}
	return order
}

// outputDifference says where an output first parts from Node's, by line. An output past
// outputKeptWhole is kept as its first line and a digest, so two large ones part at the digest's line.
func outputDifference(output, answer string) string {
	outputLines, answerLines := strings.Split(output, "\n"), strings.Split(answer, "\n")
	for index := 0; index < max(len(outputLines), len(answerLines)); index++ {
		outputLine, answerLine := "nothing", "nothing"
		if index < len(outputLines) {
			outputLine = clipped(outputLines[index])
		}
		if index < len(answerLines) {
			answerLine = clipped(answerLines[index])
		}
		if outputLine != answerLine {
			return fmt.Sprintf("line %d is %s where Node's is %s", index+1, outputLine, answerLine)
		}
	}
	return "the two differ"
}

// clipped is a line quoted, and cut to its first 100 bytes when it's longer.
func clipped(line string) string {
	if len(line) > 100 {
		return strconv.Quote(line[:100]) + "..."
	}
	return strconv.Quote(line)
}

// bestOfEach is each figure's own best over a runner's finished runs: the least wall, user and system
// time and the least peak, each possibly from a different run.
func bestOfEach(runs []measurement) measurement {
	if len(runs) == 0 {
		return measurement{failure: "not run"}
	}
	best := runs[0]
	for _, turn := range runs[1:] {
		best.elapsed = min(best.elapsed, turn.elapsed)
		best.user = min(best.user, turn.user)
		best.system = min(best.system, turn.system)
		best.peak = min(best.peak, turn.peak)
	}
	return best
}

// countEveryRunner runs every runner instructionRounds times under perf stat -e instructions:u, after
// the timed runs and in the same rotated order, and says how the counts were made, or why there are
// none. perf wraps only these runs: its own start and resident set would land in a timed run's wall
// time and peak.
func countEveryRunner(work string, floors []*gameRunner, rows []*gameRow, timeout time.Duration) string {
	perf, err := exec.LookPath("perf")
	if err != nil {
		return "none counted: perf isn't on this machine's PATH"
	}
	if _, why := countInstructions(perf, work, floors[0].command, "", nil, timeout); why != "" {
		return "none counted: perf stat -e instructions:u can't count here: " + why
	}
	for round := 0; round < instructionRounds; round++ {
		for _, row := range rows {
			ready := row.ready()
			for _, position := range gameOrder(round, len(ready)) {
				runner := ready[position]
				count, why := countInstructions(perf, work, runner.command, runner.input, runner.environment, timeout)
				if why != "" {
					runner.instructionsFailure = why
					continue
				}
				runner.instructions = append(runner.instructions, count)
			}
		}
	}
	return fmt.Sprintf("instructions:u is each runner's median of %d runs under perf stat -e instructions:u, made after the timed runs, untimed", instructionRounds)
}

// countInstructions runs a command under perf stat and reads its instructions:u, or why there are none.
func countInstructions(perf, work string, command []string, input string, settings []string, timeout time.Duration) (int64, string) {
	report, err := os.CreateTemp(work, "perf-")
	if err != nil {
		return 0, err.Error()
	}
	report.Close()
	defer os.Remove(report.Name())
	counted := append([]string{perf, "stat", "-e", "instructions:u", "-x", ",", "-o", report.Name(), "--"}, command...)
	if result := runFrom(counted, input, timeout, settings...); result.failure != "" {
		return 0, "under perf it " + result.failure
	}
	contents, err := os.ReadFile(report.Name())
	if err != nil {
		return 0, err.Error()
	}
	return perfInstructions(string(contents))
}

// perfInstructions reads perf stat's comma-separated report: the count is the first field of each
// line whose event names instructions, summed, since a machine with two kinds of core reports one
// line for each. A count perf couldn't make reads as <not supported> or <not counted>.
func perfInstructions(report string) (int64, string) {
	total, counted, reasons := int64(0), false, []string{}
	for _, line := range strings.Split(report, "\n") {
		fields := strings.Split(line, ",")
		if strings.HasPrefix(line, "#") || len(fields) < 3 || !strings.Contains(fields[2], "instructions") {
			continue
		}
		count, err := strconv.ParseInt(strings.TrimSpace(fields[0]), 10, 64)
		if err != nil {
			reasons = append(reasons, fields[2]+" read "+strings.TrimSpace(fields[0]))
			continue
		}
		total += count
		counted = true
	}
	if counted {
		return total, ""
	}
	if len(reasons) > 0 {
		return 0, strings.Join(reasons, "; ")
	}
	return 0, "perf wrote no instructions line"
}

// instructionsText is a runner's median instructions:u, or why it has none, or nothing when none were
// counted at all, which the header says.
func instructionsText(runner *gameRunner) string {
	if len(runner.instructions) > 0 {
		sorted := slices.Clone(runner.instructions)
		slices.Sort(sorted)
		return thousands(sorted[len(sorted)/2]) + " instructions:u"
	}
	if runner.instructionsFailure != "" {
		return "no instructions:u, " + runner.instructionsFailure
	}
	return ""
}

// printGameReport writes one table per basis, each with its losses beside its wins, then the
// headline, one thread against one thread only, then every gap.
func printGameReport(out io.Writer, file gameFile, results []gameResult) {
	for _, basis := range gameBases {
		rows := []gameResult{}
		for _, result := range results {
			if entryOn(result.benchmark, goLanguage, basis) >= 0 || entryOn(result.benchmark, rustLanguage, basis) >= 0 {
				rows = append(rows, result)
			}
		}
		if len(rows) == 0 {
			continue
		}
		fmt.Fprintf(out, "the %s basis:\n\n", basis)
		fmt.Fprintf(out, "| program | Adamic | Node | Go | Rust | Adamic / Go | Adamic / Rust | Adamic / Node | basis | %s, elapsed at its n: a sanity check, never a comparison |\n", file.Source.GameMachine)
		fmt.Fprintln(out, "|---|---|---|---|---|---|---|---|---|---|")
		for _, result := range rows {
			goIndex := entryOn(result.benchmark, goLanguage, basis)
			rustIndex := entryOn(result.benchmark, rustLanguage, basis)
			cells := []string{
				fmt.Sprintf("%s (%s)", result.benchmark.Name, gameSizeText(result.benchmark)),
				cellText(result.cells[gameAdamic]),
				cellText(result.cells[gameNode]),
				entryCellText(result, goIndex),
				entryCellText(result, rustIndex),
				entryRatio(result, goIndex),
				entryRatio(result, rustIndex),
				ratio(result.cells[gameAdamic].best, result.cells[gameNode].best),
				basisText(basis, result.benchmark, goIndex, rustIndex),
				sanityText(result.benchmark, goIndex, rustIndex),
			}
			fmt.Fprintln(out, "| "+strings.Join(cells, " | ")+" |")
		}
		losses, wins := lossesAndWins(rows, basis)
		fmt.Fprintf(out, "\nlosses on the %s basis, where Adamic's best wall time is above the entry's:\n", basis)
		printList(out, losses)
		fmt.Fprintf(out, "wins on the %s basis, where it's below:\n", basis)
		printList(out, wins)
		fmt.Fprintln(out)
	}
	fmt.Fprintln(out, "headline, one thread against one thread only: Adamic's programs here run one thread, so an entry that runs several is never counted, whatever basis it stands on:")
	printList(out, gameHeadline(results))
	fmt.Fprintln(out, "\ngaps, where a program or a runner has no timing:")
	printList(out, gameGaps(file, results))
}

// entryOn is the index of a program's entry in a language on a basis, or -1 when the game has none.
func entryOn(benchmark gameProgram, language, basis string) int {
	return slices.IndexFunc(benchmark.Entries, func(entry gameEntry) bool {
		return entry.Language == language && slices.Contains(entry.Bases, basis)
	})
}

// gameSizeText is a row's size as the runners get it, marked ours where the game's differs.
func gameSizeText(benchmark gameProgram) string {
	size := "n = " + thousands(int64(benchmark.Size))
	if benchmark.Input != nil {
		size = fmt.Sprintf("%s's output for %s", benchmark.Input.Generator, thousands(int64(benchmark.Input.Size)))
	}
	if benchmark.Ours != "" {
		size += ", ours"
	}
	return size
}

// cellText is a runner's figures, or a pointer to its gap below.
func cellText(cell gameCell) string {
	if cell.best.failure != "" {
		return "gap, below"
	}
	text := fmt.Sprintf("%s, user %s, sys %s, %s", seconds(cell.best.elapsed), seconds(cell.best.user), seconds(cell.best.system), megabytes(cell.best.peak))
	if cell.instructions != "" {
		text += ", " + cell.instructions
	}
	return text
}

// entryCellText is an entry's cell, named and with its threads, or why there's none on the basis.
func entryCellText(result gameResult, index int) string {
	if index < 0 {
		return "no entry on this basis"
	}
	entry := result.benchmark.Entries[index]
	return fmt.Sprintf("%s, %s: %s", entryName(entry), threadsText(entry), cellText(result.cells[gameFirstEntry+index]))
}

// entryRatio is Adamic's best wall time over an entry's.
func entryRatio(result gameResult, index int) string {
	if index < 0 {
		return ""
	}
	return ratio(result.cells[gameAdamic].best, result.cells[gameFirstEntry+index].best)
}

// threadsText is how many threads an entry runs, and the instruction set it needs, if any.
func threadsText(entry gameEntry) string {
	text := "one thread"
	if entry.Threads == severalThreads {
		text = "several threads"
	}
	if entry.InstructionSet != "" {
		text += ", needs " + entry.InstructionSet
	}
	return text
}

// basisText says the basis a row stands on and whether it's like for like: Adamic's program runs one
// thread, so a row with an entry that runs several isn't.
func basisText(basis string, benchmark gameProgram, goIndex, rustIndex int) string {
	several := []string{}
	for _, index := range []int{goIndex, rustIndex} {
		if index >= 0 && benchmark.Entries[index].Threads == severalThreads {
			several = append(several, entryName(benchmark.Entries[index]))
		}
	}
	if len(several) == 0 {
		return basis + ", like for like: one thread each"
	}
	verb := "runs"
	if len(several) > 1 {
		verb = "run"
	}
	return fmt.Sprintf("%s, not like for like: %s %s several threads, Adamic one", basis, strings.Join(several, " and "), verb)
}

// sanityText is the game's own elapsed time for the row's entries, at the game's n, labeled by the
// column; ours is never set beside it.
func sanityText(benchmark gameProgram, goIndex, rustIndex int) string {
	figures := []string{}
	for _, index := range []int{goIndex, rustIndex} {
		if index >= 0 {
			figures = append(figures, fmt.Sprintf("%s %.3f s", entryName(benchmark.Entries[index]), benchmark.Entries[index].GameElapsed))
		}
	}
	if len(figures) == 0 {
		return ""
	}
	text := strings.Join(figures, ", ") + " at n = " + thousands(int64(benchmark.GameSize))
	if benchmark.Ours != "" {
		text += "; ours is " + benchmark.Ours
	}
	return text
}

// lossesAndWins lists, for a basis, every entry Adamic's best wall time is above and every one it's
// below, so a win is never read without the losses beside it.
func lossesAndWins(rows []gameResult, basis string) ([]string, []string) {
	losses, wins := []string{}, []string{}
	for _, result := range rows {
		adamic := result.cells[gameAdamic].best
		for _, language := range entryLanguages {
			index := entryOn(result.benchmark, language, basis)
			if index < 0 {
				continue
			}
			other := result.cells[gameFirstEntry+index].best
			if adamic.failure != "" || other.failure != "" {
				continue
			}
			entry := result.benchmark.Entries[index]
			line := fmt.Sprintf("- %s: Adamic %s against %s's %s (%s), %s", result.benchmark.Name, seconds(adamic.elapsed), entryName(entry), seconds(other.elapsed), threadsText(entry), ratio(adamic, other))
			if adamic.elapsed >= other.elapsed {
				losses = append(losses, line)
			} else {
				wins = append(wins, line)
			}
		}
	}
	return losses, wins
}

// gameHeadline counts, for each language, the entries that run one thread that Adamic is faster and
// slower than. An entry that runs several threads is never counted, on any basis: Adamic's program
// runs one, so beating or trailing it says nothing like for like.
func gameHeadline(results []gameResult) []string {
	lines := []string{}
	for _, language := range entryLanguages {
		faster, slower := []string{}, []string{}
		for _, result := range results {
			adamic := result.cells[gameAdamic].best
			for index, entry := range result.benchmark.Entries {
				if entry.Language != language || entry.Threads != oneThread {
					continue
				}
				other := result.cells[gameFirstEntry+index].best
				if adamic.failure != "" || other.failure != "" {
					continue
				}
				named := fmt.Sprintf("%s %s %s", result.benchmark.Name, entryName(entry), ratio(adamic, other))
				if adamic.elapsed < other.elapsed {
					faster = append(faster, named)
				} else {
					slower = append(slower, named)
				}
			}
		}
		lines = append(lines, fmt.Sprintf("- against %s's one-thread entries: faster on %d of %d (%s), slower on %d (%s)", languageName(language), len(faster), len(faster)+len(slower), joinedOrNone(faster), len(slower), joinedOrNone(slower)))
	}
	return lines
}

// gameGaps is every program and runner without a timing, and why: the programs with no Go or Rust
// entry, the ones Adamic refuses, the bases the game has no entry on, and each runner's failure.
func gameGaps(file gameFile, results []gameResult) []string {
	gaps := []string{}
	for _, absence := range file.Absent {
		gaps = append(gaps, fmt.Sprintf("- %s: no Go or Rust entry: %s", absence.Name, absence.Why))
	}
	for _, refusal := range file.Refused {
		gaps = append(gaps, fmt.Sprintf("- %s: refused by Adamic, never timed: %s", refusal.Name, refusal.Why))
	}
	for _, basis := range []string{fastestBasis, oneThreadBasis} {
		for _, result := range results {
			for _, language := range entryLanguages {
				if entryOn(result.benchmark, language, basis) < 0 {
					gaps = append(gaps, fmt.Sprintf("- %s: the game has no %s entry on the %s basis", result.benchmark.Name, languageName(language), basis))
				}
			}
		}
	}
	for _, result := range results {
		for _, cell := range result.cells {
			if cell.best.failure != "" {
				gaps = append(gaps, fmt.Sprintf("- %s %s: %s", result.benchmark.Name, cell.name, cell.best.failure))
			}
		}
	}
	return gaps
}

func languageName(language string) string {
	if language == goLanguage {
		return "Go"
	}
	return "Rust"
}

func joinedOrNone(names []string) string {
	if len(names) == 0 {
		return "none"
	}
	return strings.Join(names, ", ")
}
