package native

// The first link's rows 2 and 3 (#sp53hvd, #bcbp175 C), a measurement carried only in proof commits and never merged.
// It reads a program's emitted C, splits it with the build's own splitter (splitC), compiles each unit as a split build
// compiles it (the unit's flags, unitCompileArguments, ClangIn, the toolchain's environment), and records each clang's
// wall time and peak memory (the child's maximum resident set) and, where clang refuses a unit, its first error. With
// the whole set compiled on one box it links as buildUnitsWithLibrary links, and reports the link's undefined symbols,
// each with how many references name it, or the binary's size. Every row goes to units.jsonl and a table on stdout.
//
//	ADAMIC_FIRSTLINK_C        the emitted C (required; the test skips without it)
//	ADAMIC_FIRSTLINK_OUT      the directory the report goes to (required)
//	ADAMIC_FIRSTLINK_SHARD    i/n: compile only the units whose index is i modulo n (default every unit)
//	ADAMIC_FIRSTLINK_JOBS     clangs at once (default the box's CPUs)
//	ADAMIC_FIRSTLINK_LINK     1 links, only with every unit compiled
//	ADAMIC_FIRSTLINK_RELEASE  1 builds as a shipped release (ThinLTO, lld)
//	ADAMIC_FIRSTLINK_PLANT    unit: unit 1 refused by an #error; link: a function calling a symbol nothing defines

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/native_runtime"
	"github.com/system-inc/adamic/internal/toolchain"
)

type firstLinkUnit struct {
	Index       int     `json:"index"`
	Name        string  `json:"name"`
	Lines       int     `json:"lines"`
	Bytes       int     `json:"bytes"`
	WallSeconds float64 `json:"wallSeconds"`
	PeakKiB     int64   `json:"peakKiB"`
	Refused     bool    `json:"refused"`
	FirstError  string  `json:"firstError,omitempty"`
	object      string
}

var firstLinkUndefined = []*regexp.Regexp{
	regexp.MustCompile("undefined reference to `([^']+)'"),
	regexp.MustCompile(`undefined symbol: (\S+)`),
}

func TestFirstLinkMeasure(t *testing.T) {
	path, out := os.Getenv("ADAMIC_FIRSTLINK_C"), os.Getenv("ADAMIC_FIRSTLINK_OUT")
	if path == "" {
		t.Skip("ADAMIC_FIRSTLINK_C names no emitted C: a first-link measurement only")
	}
	if out == "" {
		t.Fatal("ADAMIC_FIRSTLINK_OUT names no report directory")
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	source := string(contents)
	options := native_runtime.Options{Split: true, Release: os.Getenv("ADAMIC_FIRSTLINK_RELEASE") == "1"}
	started := time.Now()
	header, units, err := splitC(source)
	if err != nil {
		t.Fatalf("FAIL split: %v", err)
	}
	fmt.Printf("SPLIT %d lines, %d bytes of C, %d units, units.h %d bytes, in %.2f s\n", strings.Count(source, "\n"), len(source), len(units), len(header), time.Since(started).Seconds())
	switch os.Getenv("ADAMIC_FIRSTLINK_PLANT") {
	case "unit":
		if len(units) < 2 {
			t.Fatal("PLANT unit needs two units")
		}
		units[1].source = "#error adamic_firstlink_planted\n" + units[1].source
	case "link":
		units[len(units)-1].source += "\nvoid adamic_firstlink_planted_missing(void);\nvoid adamic_firstlink_planted_use(void) { adamic_firstlink_planted_missing(); }\n"
	}
	shard, shards := 0, 1
	if value := os.Getenv("ADAMIC_FIRSTLINK_SHARD"); value != "" {
		parts := strings.SplitN(value, "/", 2)
		if len(parts) != 2 {
			t.Fatalf("ADAMIC_FIRSTLINK_SHARD %q isn't i/n", value)
		}
		if shard, err = strconv.Atoi(parts[0]); err != nil {
			t.Fatal(err)
		}
		if shards, err = strconv.Atoi(parts[1]); err != nil || shards < 1 || shard < 0 || shard >= shards {
			t.Fatalf("ADAMIC_FIRSTLINK_SHARD %q isn't i/n with 0 <= i < n", value)
		}
	}
	jobs := runtime.NumCPU()
	if value := os.Getenv("ADAMIC_FIRSTLINK_JOBS"); value != "" {
		if jobs, err = strconv.Atoi(value); err != nil || jobs < 1 {
			t.Fatalf("ADAMIC_FIRSTLINK_JOBS %q", value)
		}
	}
	compiler, err := toolchain.Path("clang")
	if err != nil {
		t.Fatal(err)
	}
	runtimeFiles, err := native_runtime.ReadRuntime(native_runtime.Sources, "c")
	if err != nil {
		t.Fatal(err)
	}
	common := []native_runtime.RuntimeFile{{Name: "units.h", Contents: []byte(header)}}
	for _, file := range runtimeFiles {
		if strings.HasSuffix(file.Name, ".h") {
			common = append(common, file)
		}
	}
	flags := append(native_runtime.SourceFlags(source, options), programDebugFlags(options)...)
	fmt.Printf("FLAGS %s\n", strings.Join(flags, " "))
	directory, err := os.MkdirTemp("", "adamic-firstlink-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(directory)
	var selected []*firstLinkUnit
	for index, unit := range units {
		if index%shards == shard {
			selected = append(selected, &firstLinkUnit{Index: index, Name: unit.name, Lines: strings.Count(unit.source, "\n"), Bytes: len(unit.source)})
		}
	}
	work := make(chan *firstLinkUnit)
	var workers sync.WaitGroup
	compileStarted := time.Now()
	for range min(jobs, max(len(selected), 1)) {
		workers.Go(func() {
			for row := range work {
				firstLinkCompile(row, units[row.Index], common, flags, compiler, directory)
			}
		})
	}
	for _, row := range selected {
		work <- row
	}
	close(work)
	workers.Wait()
	compileWall := time.Since(compileStarted).Seconds()
	report, err := os.Create(filepath.Join(out, "units.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	encoder := json.NewEncoder(report)
	refused, total, peak := 0, 0.0, int64(0)
	for _, row := range selected {
		if err := encoder.Encode(row); err != nil {
			t.Fatal(err)
		}
		total += row.WallSeconds
		peak = max(peak, row.PeakKiB)
		if row.Refused {
			refused++
		}
	}
	report.Close()
	fmt.Printf("COMPILED shard %d/%d: %d units, %d refused, clang %.1f s summed, %.1f s wall at %d jobs, peak %d KiB\n", shard, shards, len(selected), refused, total, compileWall, jobs, peak)
	byWall := slices.Clone(selected)
	slices.SortFunc(byWall, func(a, b *firstLinkUnit) int { return int((b.WallSeconds - a.WallSeconds) * 1e6) })
	fmt.Println("| unit | lines | bytes | clang s | peak KiB |")
	fmt.Println("|---|---:|---:|---:|---:|")
	for _, row := range byWall[:min(20, len(byWall))] {
		fmt.Printf("| %s | %d | %d | %.2f | %d |\n", row.Name, row.Lines, row.Bytes, row.WallSeconds, row.PeakKiB)
	}
	for _, row := range selected {
		if row.Refused {
			fmt.Printf("REFUSED %s: %s\n", row.Name, row.FirstError)
		}
	}
	if os.Getenv("ADAMIC_FIRSTLINK_LINK") != "1" {
		return
	}
	if shards != 1 || refused != 0 {
		fmt.Printf("LINK not run: %d of %d units compiled here, %d refused\n", len(selected), len(units), refused)
		return
	}
	firstLinkLink(t, source, options, compiler, directory, selected, out)
}

// firstLinkCompile compiles one unit as compileUnitObject does, timing clang and reading its peak resident set.
func firstLinkCompile(row *firstLinkUnit, unit compilationUnit, common []native_runtime.RuntimeFile, flags []string, compiler, directory string) {
	temporary, err := os.MkdirTemp(directory, ".unit-")
	if err != nil {
		row.Refused, row.FirstError = true, err.Error()
		return
	}
	for _, file := range append([]native_runtime.RuntimeFile{{Name: unit.name, Contents: []byte(unit.source)}}, common...) {
		if err := os.WriteFile(filepath.Join(temporary, file.Name), file.Contents, 0o644); err != nil {
			row.Refused, row.FirstError = true, err.Error()
			return
		}
	}
	command := native_runtime.ClangIn(temporary, compiler, append(slices.Clone(flags), unitCompileArguments(unit.name, temporary, flags, false)...)...)
	working := []string{}
	for _, setting := range command.Env {
		if strings.HasPrefix(setting, "PWD=") {
			working = append(working, setting)
		}
	}
	command.Env = append(toolchain.Env(), working...)
	var output bytes.Buffer
	command.Stdout, command.Stderr = &output, &output
	started := time.Now()
	err = command.Run()
	row.WallSeconds = time.Since(started).Seconds()
	if command.ProcessState != nil {
		if usage, ok := command.ProcessState.SysUsage().(*syscall.Rusage); ok {
			row.PeakKiB = usage.Maxrss
		}
	}
	if err != nil {
		row.Refused = true
		row.FirstError = firstLinkFirstError(output.String(), err)
		return
	}
	row.object = filepath.Join(temporary, "unit.o")
}

// firstLinkFirstError is clang's first error line, or the run's own failure when it printed none.
func firstLinkFirstError(output string, err error) string {
	scanner := bufio.NewScanner(strings.NewReader(output))
	scanner.Buffer(make([]byte, 1<<20), 1<<20)
	for scanner.Scan() {
		if line := scanner.Text(); strings.Contains(line, "error:") {
			return line
		}
	}
	return err.Error()
}

// firstLinkLink links every unit as buildUnitsWithLibrary does, and reports the undefined symbols or the binary.
func firstLinkLink(t *testing.T, source string, options native_runtime.Options, compiler, directory string, rows []*firstLinkUnit, out string) {
	runtimeOptions := options
	runtimeOptions.Split, runtimeOptions.Jobs = false, 0
	started := time.Now()
	library, err := native_runtime.RuntimeLibraryForSource("", source, runtimeOptions)
	if err != nil {
		fmt.Printf("LINK FAIL runtime archive: %v\n", err)
		return
	}
	fmt.Printf("RUNTIME archive %s in %.1f s\n", filepath.Base(library), time.Since(started).Seconds())
	output := filepath.Join(out, "firstlink.bin")
	arguments := append(LinkFlags(options), "-o", output)
	if slices.Contains(arguments, "-fuse-ld=lld") {
		arguments = append(arguments, "-Wl,--error-limit=0")
	}
	for index, row := range rows {
		name := fmt.Sprintf("unit-%d.o", index)
		if err := os.Symlink(row.object, filepath.Join(directory, name)); err != nil {
			t.Fatal(err)
		}
		arguments = append(arguments, name)
	}
	if err := os.Symlink(filepath.Dir(library), filepath.Join(directory, "runtime")); err != nil {
		t.Fatal(err)
	}
	arguments = append(arguments, native_runtime.RuntimeLinkFlags(filepath.Join("runtime", filepath.Base(library)))...)
	compilerLink, err := regexpCompilerLinkFlags(source, options, directory)
	if err != nil {
		t.Fatal(err)
	}
	arguments = append(arguments, compilerLink...)
	arguments = append(arguments, objectPrefixFlags()...)
	arguments = append(arguments, native_runtime.MathLinkFlags(options)...)
	link := native_runtime.ClangIn(directory, compiler, arguments...)
	var linked bytes.Buffer
	link.Stdout, link.Stderr = &linked, &linked
	started = time.Now()
	err = link.Run()
	wall := time.Since(started).Seconds()
	var peak int64
	if link.ProcessState != nil {
		if usage, ok := link.ProcessState.SysUsage().(*syscall.Rusage); ok {
			peak = usage.Maxrss
		}
	}
	if err := os.WriteFile(filepath.Join(out, "link.log"), linked.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	if err == nil {
		info, statErr := os.Stat(output)
		if statErr != nil {
			t.Fatal(statErr)
		}
		stripped := output + ".stripped"
		strip := exec.Command("strip", "-o", stripped, output)
		size := int64(-1)
		if strip.Run() == nil {
			if strippedInfo, err := os.Stat(stripped); err == nil {
				size = strippedInfo.Size()
			}
		}
		fmt.Printf("LINK OK in %.1f s, peak %d KiB: %d bytes, %d stripped\n", wall, peak, info.Size(), size)
		return
	}
	references := map[string]int{}
	for _, line := range strings.Split(linked.String(), "\n") {
		for _, pattern := range firstLinkUndefined {
			if match := pattern.FindStringSubmatch(line); match != nil {
				references[match[1]]++
			}
		}
	}
	symbols := make([]string, 0, len(references))
	for symbol := range references {
		symbols = append(symbols, symbol)
	}
	slices.Sort(symbols)
	fmt.Printf("LINK FAIL in %.1f s, peak %d KiB: %d undefined symbols, %d references\n", wall, peak, len(symbols), func() (sum int) {
		for _, count := range references {
			sum += count
		}
		return
	}())
	for _, symbol := range symbols {
		fmt.Printf("UNDEFINED %s %d\n", symbol, references[symbol])
	}
	if len(symbols) == 0 {
		fmt.Println("LINK FAIL with no undefined symbol named; the log's head:")
		lines := strings.Split(linked.String(), "\n")
		fmt.Println(strings.Join(lines[:min(20, len(lines))], "\n"))
	}
}
