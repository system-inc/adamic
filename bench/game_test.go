package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"
)

// gameConfiguration is bench/game/programs.json, read and checked the way -game reads it.
func gameConfiguration(t *testing.T) gameFile {
	t.Helper()
	file, err := loadGame(filepath.Join("game", "programs.json"))
	if err != nil {
		t.Fatal(err)
	}
	return file
}

// The configuration parses, holds together, and names only files that are there: every entry is a
// vendored file, every vendored file exists, every Adamic source and input generator is a bench/
// program, and nothing in bench/game is unaccounted for, neither vendored nor one of its own files.
func TestTheGameConfigurationNamesOnlyVendoredFilesThatExist(t *testing.T) {
	t.Parallel()
	file := gameConfiguration(t)
	vendored := map[string]bool{}
	for _, listed := range file.Vendored {
		vendored[listed.Path] = true
		if _, err := os.Stat(filepath.Join("game", filepath.FromSlash(listed.Path))); err != nil {
			t.Errorf("programs.json vendors %s, which isn't there: %v", listed.Path, err)
		}
	}
	for _, benchmark := range file.Programs {
		if _, err := benchSource(".", benchmark.Adamic); err != nil {
			t.Errorf("%s: %v", benchmark.Name, err)
		}
		if benchmark.Input != nil {
			if _, err := benchSource(".", benchmark.Input.Generator); err != nil {
				t.Errorf("%s's input: %v", benchmark.Name, err)
			}
		}
		if !vendored[benchmark.Name+"/LICENSE"] {
			t.Errorf("%s's LICENSE isn't vendored beside its entries", benchmark.Name)
		}
		for _, entry := range benchmark.Entries {
			if !vendored[entry.File] {
				t.Errorf("%s names %s, which isn't vendored", benchmark.Name, entry.File)
			}
		}
	}
	if !vendored["LICENSE"] {
		t.Error("the game's root LICENSE isn't vendored")
	}
	ours := []string{"README.md", "programs.json", "go.mod", "go.sum"}
	err := filepath.WalkDir("game", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative := filepath.ToSlash(strings.TrimPrefix(path, "game"+string(filepath.Separator)))
		if entry.IsDir() {
			if relative == "rust-deps" {
				return filepath.SkipDir
			}
			return nil
		}
		if !vendored[relative] && !slices.Contains(ours, relative) {
			t.Errorf("bench/game/%s is neither vendored in programs.json nor one of bench/game's own files", relative)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// Every vendored file is the archive's bytes: its sha256 is the one programs.json records, read from
// the game's archive when the file was vendored, so an edit, a line-ending conversion or a stray
// formatter reds here.
func TestVendoredFilesAreTheArchivesBytes(t *testing.T) {
	t.Parallel()
	file := gameConfiguration(t)
	if len(file.Vendored) == 0 {
		t.Fatal("programs.json vendors nothing")
	}
	for _, listed := range file.Vendored {
		contents, err := os.ReadFile(filepath.Join("game", filepath.FromSlash(listed.Path)))
		if err != nil {
			t.Error(err)
			continue
		}
		sum := sha256.Sum256(contents)
		if got := hex.EncodeToString(sum[:]); got != listed.Sha256 {
			t.Errorf("bench/game/%s has sha256 %s; the archive's is %s", listed.Path, got, listed.Sha256)
		}
	}
}

// Each entry is built as the game built it: a Rust entry's build line carries the flags the runner
// passes, and every crate it passes with --extern is one the runner passes too; a Go entry's builds
// its own file. Every pinned crate is pinned once, the same in Cargo.toml as in programs.json, and is
// one some entry links against.
func TestEveryEntryIsBuiltAsTheGameBuiltIt(t *testing.T) {
	t.Parallel()
	file := gameConfiguration(t)
	flags := strings.Join(file.Rust.Flags, " ")
	extern := regexp.MustCompile(`--extern ([a-z0-9_]+)=`)
	used := map[string]bool{}
	for _, benchmark := range file.Programs {
		for _, entry := range benchmark.Entries {
			base := entry.File[strings.LastIndex(entry.File, "/")+1:]
			switch entry.Language {
			case goLanguage:
				if !strings.Contains(entry.BuildLine, "go1.23.1") || !strings.HasSuffix(entry.BuildLine, " "+base) {
					t.Errorf("%s's build line %q isn't Go 1.23.1's go build of it", entry.File, entry.BuildLine)
				}
			case rustLanguage:
				if !strings.Contains(entry.BuildLine, "rust-"+file.Rust.Version+"/bin/rustc "+flags+" ") {
					t.Errorf("%s's build line %q doesn't pass %s with rustc %s", entry.File, entry.BuildLine, flags, file.Rust.Version)
				}
				for _, match := range extern.FindAllStringSubmatch(entry.BuildLine, -1) {
					if !slices.ContainsFunc(entry.Crates, func(crate string) bool { return strings.ReplaceAll(crate, "-", "_") == match[1] }) {
						t.Errorf("%s's build line passes --extern %s, which its crates don't", entry.File, match[1])
					}
				}
			}
			for _, crate := range entry.Crates {
				used[crate] = true
			}
		}
	}
	manifest, err := os.ReadFile(filepath.Join("game", "rust-deps", "Cargo.toml"))
	if err != nil {
		t.Fatal(err)
	}
	pinned := map[string]string{}
	for _, match := range regexp.MustCompile(`(?m)^([a-z0-9_-]+) = "=([0-9.]+)"$`).FindAllStringSubmatch(string(manifest), -1) {
		if _, twice := pinned[match[1]]; twice {
			t.Errorf("Cargo.toml pins %s twice", match[1])
		}
		pinned[match[1]] = match[2]
	}
	for crate, version := range file.Crates {
		if pinned[crate] != version {
			t.Errorf("programs.json pins %s at %s, and Cargo.toml at %q", crate, version, pinned[crate])
		}
		if !used[crate] {
			t.Errorf("programs.json pins %s, which no entry links against", crate)
		}
	}
	for crate := range pinned {
		if _, listed := file.Crates[crate]; !listed {
			t.Errorf("Cargo.toml pins %s, which programs.json doesn't", crate)
		}
	}
}

// timed is a finished run of one duration, every figure from it.
func timed(duration time.Duration) measurement {
	return measurement{elapsed: duration, user: duration, system: duration / 10, peak: 4 << 20, stdout: "42\n"}
}

// reportOf is the report printGameReport writes for two programs: fasta, which Adamic beats on the
// Go entry and trails on the Rust one, and mandelbrot, whose Adamic build stopped.
func reportOf(t *testing.T) string {
	t.Helper()
	fasta := gameProgram{Name: "fasta", Adamic: "holdout/fasta", Size: 25000000, GameSize: 25000000, Entries: []gameEntry{
		{File: "fasta/fasta.go-2.go", Language: goLanguage, Bases: []string{fastestBasis}, Threads: severalThreads, GameElapsed: 1.273, GameCpu: 4.073},
		{File: "fasta/fasta.rust-7.rust", Language: rustLanguage, Bases: []string{fastestBasis}, Threads: severalThreads, GameElapsed: 0.778, GameCpu: 1.545},
	}}
	mandelbrot := gameProgram{Name: "mandelbrot", Adamic: "holdout/mandelbrot", Size: 16000, GameSize: 16000, Entries: []gameEntry{
		{File: "mandelbrot/mandelbrot.go-4.go", Language: goLanguage, Bases: []string{fastestBasis}, Threads: severalThreads, GameElapsed: 3.774, GameCpu: 15.009},
	}}
	results := []gameResult{
		{benchmark: fasta, cells: []gameCell{
			{name: adamicLanguage, best: timed(2 * time.Second)},
			{name: nodeLanguage, best: timed(5 * time.Second)},
			{name: "go-2", best: timed(3 * time.Second)},
			{name: "rust-7", best: timed(time.Second)},
		}},
		{benchmark: mandelbrot, cells: []gameCell{
			{name: adamicLanguage, best: measurement{failure: "adamic build stopped: exit status 1 NotYet: process.stdout.write of a Uint8Array"}},
			{name: nodeLanguage, best: timed(9 * time.Second)},
			{name: "go-4", best: timed(time.Second)},
		}},
	}
	file := gameFile{
		Source:  gameSource{GameMachine: "the game's i5-3330"},
		Absent:  []gameAbsence{{Name: "sort", Why: "not a Benchmarks Game program"}},
		Refused: []gameAbsence{{Name: "pidigits", Why: "no bigint yet"}},
	}
	var out bytes.Buffer
	printGameReport(&out, file, results)
	return out.String()
}

// The table puts a loss beside a win: under the basis it stands on, the entry Adamic trails is listed
// as a loss and the one it beats as a win, both before the next section. A row whose Adamic build
// stopped is a gap, named with the compiler's line, and so are the programs with no entry and the
// one Adamic refuses.
func TestTheGameTablePutsALossBesideAWinAndNamesAGap(t *testing.T) {
	t.Parallel()
	report := reportOf(t)
	losses := strings.Index(report, "losses on the fastest basis")
	loss := strings.Index(report, "- fasta: Adamic 2.000 s against rust-7's 1.000 s (several threads), 2.00x")
	wins := strings.Index(report, "wins on the fastest basis")
	win := strings.Index(report, "- fasta: Adamic 2.000 s against go-2's 3.000 s (several threads), 0.67x")
	headline := strings.Index(report, "headline, one thread against one thread only")
	if losses < 0 || !(losses < loss && loss < wins && wins < win && win < headline) {
		t.Fatalf("the loss and the win aren't listed beside each other under their basis (losses at %d, loss %d, wins %d, win %d, headline %d):\n%s", losses, loss, wins, win, headline, report)
	}
	gaps := strings.Index(report, "gaps, where a program or a runner has no timing")
	for _, want := range []string{
		"- mandelbrot adamic: adamic build stopped: exit status 1 NotYet: process.stdout.write of a Uint8Array",
		"- sort: no Go or Rust entry: not a Benchmarks Game program",
		"- pidigits: refused by Adamic, never timed: no bigint yet",
		"- mandelbrot: the game has no Rust entry on the fastest basis",
	} {
		if at := strings.Index(report, want); gaps < 0 || at < gaps {
			t.Errorf("the gaps don't list %q:\n%s", want, report)
		}
	}
	if !strings.Contains(report, "| mandelbrot (n = 16,000) | gap, below |") {
		t.Errorf("mandelbrot's Adamic cell doesn't point to its gap:\n%s", report)
	}
	if strings.Contains(report, "mandelbrot: Adamic") {
		t.Errorf("mandelbrot, whose Adamic build stopped, is listed as a loss or a win:\n%s", report)
	}
}

// The headline counts one thread against one thread only. Adamic beats an entry that runs several
// threads here, and trails one that runs one: the headline names the one-thread entry and never the
// several-thread one, though both stand on the fastest basis. And the configuration refuses an entry
// that runs several threads on the one thread basis, so no table can headline one there either.
func TestAOneThreadRowIsNeverHeadlinedAgainstSeveralThreads(t *testing.T) {
	t.Parallel()
	nbody := gameProgram{Name: "nbody", Adamic: "nbody", Size: 1000000, GameSize: 50000000, Entries: []gameEntry{
		{File: "nbody/nbody.go-3.go", Language: goLanguage, Bases: []string{fastestBasis, oneThreadBasis}, Threads: oneThread, GameElapsed: 6.385, GameCpu: 6.4},
		{File: "nbody/nbody.rust-9.rust", Language: rustLanguage, Bases: []string{fastestBasis}, Threads: severalThreads, GameElapsed: 2.192, GameCpu: 2.189},
	}}
	results := []gameResult{{benchmark: nbody, cells: []gameCell{
		{name: adamicLanguage, best: timed(2 * time.Second)},
		{name: nodeLanguage, best: timed(5 * time.Second)},
		{name: "go-3", best: timed(time.Second)},
		{name: "rust-9", best: timed(4 * time.Second)},
	}}}
	headline := strings.Join(gameHeadline(results), "\n")
	if strings.Contains(headline, "rust-9") || !strings.Contains(headline, "against Rust's one-thread entries: faster on 0 of 0") {
		t.Errorf("the headline counted rust-9, which runs several threads:\n%s", headline)
	}
	if !strings.Contains(headline, "against Go's one-thread entries: faster on 0 of 1 (none), slower on 1 (nbody go-3 2.00x)") {
		t.Errorf("the headline didn't count go-3, which runs one thread, as a loss:\n%s", headline)
	}
	file := gameFile{
		Go:       gameGo{Version: "go1.23.1"},
		Rust:     gameRust{Version: "1.84.1", Flags: []string{"-C", "opt-level=3"}},
		Vendored: []vendoredFile{{Path: "nbody/nbody.rust-9.rust", Sha256: strings.Repeat("0", 64)}},
		Programs: []gameProgram{{Name: "nbody", Adamic: "nbody", Size: 1000000, GameSize: 50000000, Entries: []gameEntry{
			{File: "nbody/nbody.rust-9.rust", Language: rustLanguage, Bases: []string{oneThreadBasis}, Threads: severalThreads, BuildLine: "rustc", GameElapsed: 2.192, GameCpu: 2.189},
		}}},
	}
	if err := file.validate(); err == nil || !strings.Contains(err.Error(), "one thread basis") {
		t.Errorf("an entry running several threads on the one thread basis read %v; want it refused", err)
	}
	file.Programs[0].Entries[0].Threads = oneThread
	if err := file.validate(); err != nil {
		t.Errorf("the same entry running one thread was refused: %v", err)
	}
}

// Every sample runs each runner once. Over 2n samples of three or more runners, each follows two
// different runners, so none always follows the same one, and each runs in every position twice. The
// mutant is the plain rotation: every runner follows only the one before it.
func TestTheGameOrderVariesWhoRunsBefore(t *testing.T) {
	t.Parallel()
	for count := 3; count <= 7; count++ {
		everyone := make([]int, count)
		positions := make([][]int, count)
		predecessors := make([]map[int]bool, count)
		for runner := range count {
			everyone[runner] = runner
			positions[runner] = make([]int, count)
			predecessors[runner] = map[int]bool{}
		}
		for sample := range 2 * count {
			order := gameOrder(sample, count)
			if !slices.Equal(slices.Sorted(slices.Values(order)), everyone) {
				t.Fatalf("sample %d of %d runners runs %v, not each runner once", sample, count, order)
			}
			for position, runner := range order {
				positions[runner][position]++
				if position > 0 {
					predecessors[runner][order[position-1]] = true
				}
			}
		}
		for runner := range count {
			if len(predecessors[runner]) < 2 {
				t.Fatalf("over %d samples of %d runners, runner %d followed only %v", 2*count, count, runner, predecessors[runner])
			}
			for position, times := range positions[runner] {
				if times != 2 {
					t.Fatalf("over %d samples of %d runners, runner %d ran in position %d %d times; want 2", 2*count, count, runner, position, times)
				}
			}
		}
	}
}

// A crate's rlib is told from another version's by the .d file cargo writes beside it, and a version
// that isn't there is refused with how many candidates there were.
func TestCrateLibraryPicksThePinnedVersionByItsDependencyFile(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	write := func(name, contents string) {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(contents), 0644); err != nil {
			t.Fatal(err)
		}
	}
	write("libfutures-aaaa.rlib", "")
	write("futures-aaaa.d", "/target/release/deps/futures-aaaa.d: /registry/src/index.crates.io-1/futures-0.1.31/src/lib.rs\n")
	write("libfutures-bbbb.rlib", "")
	write("futures-bbbb.d", "/target/release/deps/futures-bbbb.d: /registry/src/index.crates.io-1/futures-0.3.31/src/lib.rs\n")
	write("libfutures_util-cccc.rlib", "")
	write("futures_util-cccc.d", "/target/release/deps/futures_util-cccc.d: /registry/src/index.crates.io-1/futures-util-0.1.31/src/lib.rs\n")
	write("libtokio_threadpool-dddd.rlib", "")
	write("tokio_threadpool-dddd.d", "/target/release/deps/tokio_threadpool-dddd.d: /registry/src/index.crates.io-1/tokio-threadpool-0.1.18/src/lib.rs\n")
	if got, err := crateLibrary(directory, "futures", "0.1.31"); err != nil || got != filepath.Join(directory, "libfutures-aaaa.rlib") {
		t.Errorf("futures 0.1.31 read %q, %v; want libfutures-aaaa.rlib", got, err)
	}
	if got, err := crateLibrary(directory, "tokio-threadpool", "0.1.18"); err != nil || got != filepath.Join(directory, "libtokio_threadpool-dddd.rlib") {
		t.Errorf("tokio-threadpool 0.1.18 read %q, %v; want libtokio_threadpool-dddd.rlib", got, err)
	}
	if got, err := crateLibrary(directory, "futures", "0.2.0"); err == nil || !strings.Contains(err.Error(), "0 of the 2") {
		t.Errorf("futures 0.2.0, which isn't there, read %q, %v; want it refused, naming 0 of the 2 candidates", got, err)
	}
}

// perf stat's report gives a count, summed over a hybrid machine's two kinds of core, or why it has
// none.
func TestPerfInstructionsReadsTheCountOrWhyNot(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		report string
		count  int64
		why    string
	}{
		{"# started on Thu Nov 19 2026\n\n123456789,,instructions:u,1000000,100.00,,\n", 123456789, ""},
		{"100,,cpu_core/instructions:u/,10,100.00,,\n<not counted>,,cpu_atom/instructions:u/,0,0.00,,\n", 100, ""},
		{"<not supported>,,instructions:u,0,100.00,,\n", 0, "instructions:u read <not supported>"},
		{"# started on Thu Nov 19 2026\n", 0, "perf wrote no instructions line"},
	} {
		count, why := perfInstructions(test.report)
		if count != test.count || why != test.why {
			t.Errorf("%q read %d, %q; want %d, %q", test.report, count, why, test.count, test.why)
		}
	}
}

// The entries that link PCRE are exactly those whose source imports go-pcre or that link pcre2-sys,
// so -pcre reaches both regexredux entries and nothing else. With -pcre, a Go build of one looks in the
// prefix's pkgconfig, headers and libraries, keeping cgo's -O2 -g, and its runs load the prefix's
// libraries; without it, nothing is added and the system's are used.
func TestPcreReachesTheEntriesThatLinkIt(t *testing.T) {
	t.Parallel()
	file := gameConfiguration(t)
	linking := 0
	for _, benchmark := range file.Programs {
		for _, entry := range benchmark.Entries {
			source, err := os.ReadFile(filepath.Join("game", filepath.FromSlash(entry.File)))
			if err != nil {
				t.Fatal(err)
			}
			links := strings.Contains(string(source), `"github.com/GRbit/go-pcre"`) || slices.Contains(entry.Crates, "pcre2-sys")
			if entry.Pcre != links {
				t.Errorf("%s says pcre %v, and its source and crates say %v", entry.File, entry.Pcre, links)
			}
			if entry.Pcre {
				linking++
			}
		}
	}
	if linking != 2 {
		t.Errorf("%d entries link PCRE; want regexredux's two", linking)
	}
	request := gameRequest{pcre: "/srv/loom-tools/pcre"}
	if got, want := request.pcreBuild(), []string{
		"PKG_CONFIG_LIBDIR=/srv/loom-tools/pcre/usr/lib/x86_64-linux-gnu/pkgconfig",
		"PKG_CONFIG_SYSROOT_DIR=/srv/loom-tools/pcre",
		"CGO_CFLAGS=-O2 -g -I/srv/loom-tools/pcre/usr/include",
		"CGO_LDFLAGS=-O2 -g -L/srv/loom-tools/pcre/usr/lib/x86_64-linux-gnu",
	}; !slices.Equal(got, want) {
		t.Errorf("-pcre's build settings are %q; want %q", got, want)
	}
	if got := request.pcreRun(); !slices.Equal(got, []string{"LD_LIBRARY_PATH=/srv/loom-tools/pcre/usr/lib/x86_64-linux-gnu"}) {
		t.Errorf("-pcre's run settings are %q", got)
	}
	if none := (gameRequest{}); none.pcreBuild() != nil || none.pcreRun() != nil || none.pcreLibraries() != "" {
		t.Errorf("with no -pcre, the settings are %q and %q; want none", none.pcreBuild(), none.pcreRun())
	}
}

// An entry's name is the game's: its file's name without the program's and the language's extension.
func TestAnEntrysNameIsTheGames(t *testing.T) {
	t.Parallel()
	for file, want := range map[string]string{
		"fasta/fasta.go-2.go":          "go-2",
		"fasta/fasta.go":               "go",
		"revcomp/revcomp.rust":         "rust",
		"nbody/nbody.rust-9.rust":      "rust-9",
		"spectralnorm/spectralnorm.go": "go",
	} {
		language := goLanguage
		if strings.Contains(file, ".rust") {
			language = rustLanguage
		}
		if got := entryName(gameEntry{File: file, Language: language}); got != want {
			t.Errorf("%s is named %q; want %q", file, got, want)
		}
	}
}

// A stopped build's gap line is the line that says why, never go's progress line above it: J-wme5jp2's
// regexredux go-5 gap read only "go: downloading github.com/GRbit/go-pcre v1.0.0".
func TestAStoppedBuildsGapLineIsItsError(t *testing.T) {
	for _, each := range []struct{ output, want string }{
		{"go: downloading github.com/GRbit/go-pcre v1.0.0\nregexredux.go-5.go:9:2: no required module provides package\n", "regexredux.go-5.go:9:2: no required module provides package"},
		{"# command-line-arguments\n./x.go:3:1: syntax error\n", "./x.go:3:1: syntax error"},
		{"warning: unused variable\nerror[E0425]: cannot find value\n", "error[E0425]: cannot find value"},
	} {
		if got := stopLine(each.output); got != each.want {
			t.Errorf("stopLine(%q) = %q, want %q", each.output, got, each.want)
		}
	}
}
