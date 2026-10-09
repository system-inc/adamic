// Command record writes one bench/quiet_machine.sh run to bench/results as a validated record
// (internal/benchresults). quiet_machine.sh --record runs it after the copy's tree is checked, from the
// module root:
//
//	go run ./internal/benchresults/tools/record -runs <runs.tsv> -engine <engine> -project <name> \
//	  -project-commit <sha> -idle-floor <percent> -settle <seconds> -rounds <n> -modes <Cold,Warm,...> \
//	  -band <percent> [-stopped-early] [-edit <path>]
//
// The engine's commit and whether its tree was modified come from the binary's own build info, not from
// its --version prose, so a binary that names no commit is refused rather than recorded as clean.
package main

import (
	"bufio"
	"debug/buildinfo"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/system-inc/cohere/internal/benchresults"
)

func main() {
	runsPath := flag.String("runs", "", "the run's runs.tsv")
	engine := flag.String("engine", "", "the pinned engine binary every run used")
	project := flag.String("project", "", "the project's name")
	projectCommit := flag.String("project-commit", "", "the commit the copy was extracted at")
	idleFloor := flag.Float64("idle-floor", 0, "the share of all cores, in percent, idle before and after a quiet run")
	settle := flag.Float64("settle", 0, "the settle seconds")
	rounds := flag.Int("rounds", 0, "the rounds taken")
	modes := flag.String("modes", "", "the measured modes, comma-separated: Cold, Warm, Edit, Idle")
	band := flag.Float64("band", 0, "the early stop's agreement band, in percent")
	stoppedEarly := flag.Bool("stopped-early", false, "the run stopped early because its quiet runs agreed")
	edit := flag.String("edit", "", "the file the Edit runs changed, when Edit was measured")
	out := flag.String("out", benchresults.Directory, "the directory to write the record to")
	flag.Parse()

	measured := []benchresults.Mode{}
	for _, mode := range strings.Split(*modes, ",") {
		measured = append(measured, benchresults.Mode(mode))
	}
	path, err := record(*runsPath, *engine, *project, *projectCommit, *idleFloor, *settle, *rounds, measured, *band, *stoppedEarly, *edit, *out)
	if err != nil {
		fmt.Fprintf(os.Stderr, "record: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("recorded %s\n", path)
}

func record(runsPath, engine, project, projectCommit string, idleFloor, settle float64, rounds int, measured []benchresults.Mode,
	band float64, stoppedEarly bool, edit, out string) (string, error) {
	cohere, err := engineProvenance(engine)
	if err != nil {
		return "", err
	}
	machine, err := thisMachine()
	if err != nil {
		return "", err
	}
	runs, err := readRuns(runsPath)
	if err != nil {
		return "", err
	}
	result := benchresults.Record{
		Schema:        benchresults.SchemaVersion,
		RecordedAt:    time.Now().UTC().Format(time.RFC3339),
		Cohere:        cohere,
		Project:       benchresults.Project{Name: project, Commit: projectCommit},
		Machine:       machine,
		IdleFloor:     idleFloor,
		SettleSeconds: settle,
		Rounds:        rounds,
		MeasuredModes: measured,
		Band:          band,
		StoppedEarly:  stoppedEarly,
		Edit:          edit,
		Runs:          runs,
		Modes:         benchresults.Summarize(measured, runs),
	}
	encoded, err := benchresults.Encode(result)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(out, benchresults.FileName(result))
	if _, err := os.Stat(path); err == nil {
		return "", fmt.Errorf("%s exists; a second run of this commit on this machine today needs a record of its own", path)
	}
	return path, os.WriteFile(path, encoded, 0o644)
}

// engineProvenance is the engine's commit, its first --version line, and whether it was built from a
// modified tree, as the engine states them (internal/release/packaging's Provenance). The launcher builds
// from a committed snapshot with VCS stamping off and stamps the commit itself, so the commit comes from
// the `commit:` line, lengthened to the full commit by this repository, and the tree was modified when the
// `source:` line is printed or the binary's own vcs.modified stamp says so.
func engineProvenance(engine string) (benchresults.Cohere, error) {
	output, err := exec.Command(engine, "--version").Output()
	if err != nil {
		return benchresults.Cohere{}, fmt.Errorf("%s --version: %w", engine, err)
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	cohere := benchresults.Cohere{Version: strings.TrimSpace(lines[0])}
	short := ""
	for _, line := range lines[1:] {
		key, value, found := strings.Cut(strings.TrimSpace(line), ":")
		if !found {
			continue
		}
		switch key {
		case "commit":
			short = strings.TrimSpace(value)
		case "source":
			cohere.Dirty = true
		}
	}
	if short == "" {
		return cohere, fmt.Errorf("%s names no commit, so nothing reproduces its numbers", engine)
	}
	full, err := exec.Command("git", "rev-parse", "--verify", "--quiet", short+"^{commit}").Output()
	if err != nil {
		return cohere, fmt.Errorf("%s was built from %s, which this repository does not have", engine, short)
	}
	cohere.Commit = strings.TrimSpace(string(full))
	if info, err := buildinfo.ReadFile(engine); err == nil {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.modified" && setting.Value == "true" {
				cohere.Dirty = true
			}
		}
	}
	return cohere, nil
}

// thisMachine is the hardware, read from sysctl and sw_vers. The host's name is never read.
func thisMachine() (benchresults.Machine, error) {
	read := func(name string, arguments ...string) (string, error) {
		output, err := exec.Command(name, arguments...).Output()
		if err != nil {
			return "", fmt.Errorf("%s %s: %w", name, strings.Join(arguments, " "), err)
		}
		return strings.TrimSpace(string(output)), nil
	}
	machine := benchresults.Machine{}
	var err error
	if machine.Model, err = read("sysctl", "-n", "hw.model"); err != nil {
		return machine, err
	}
	if machine.Processor, err = read("sysctl", "-n", "machdep.cpu.brand_string"); err != nil {
		return machine, err
	}
	cores, err := read("sysctl", "-n", "hw.ncpu")
	if err != nil {
		return machine, err
	}
	if machine.Cores, err = strconv.Atoi(cores); err != nil {
		return machine, err
	}
	memory, err := read("sysctl", "-n", "hw.memsize")
	if err != nil {
		return machine, err
	}
	if machine.MemoryBytes, err = strconv.ParseInt(memory, 10, 64); err != nil {
		return machine, err
	}
	version, err := read("sw_vers", "-productVersion")
	if err != nil {
		return machine, err
	}
	machine.OS = "macOS " + version
	return machine, nil
}

// readRuns reads quiet_machine.sh's runs table: mode, round, verdict_s, settled_s, engine_s, findings,
// cache, load_before, load_after, idle_before, idle_after, quiet, exit, primes, tab-separated, with a header
// line the script prints apart.
func readRuns(path string) ([]benchresults.Run, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	modes := map[string]benchresults.Mode{"cold": benchresults.ModeCold, "warm": benchresults.ModeWarm, "edit": benchresults.ModeEdit,
		"idle": benchresults.ModeIdle}
	runs := []benchresults.Run{}
	scanner := bufio.NewScanner(file)
	for line := 1; scanner.Scan(); line++ {
		fields := strings.Split(scanner.Text(), "\t")
		if len(fields) != 14 {
			return nil, fmt.Errorf("%s:%d: %d fields, want 14", path, line, len(fields))
		}
		mode, found := modes[fields[0]]
		if !found {
			return nil, fmt.Errorf("%s:%d: no mode %q", path, line, fields[0])
		}
		if fields[11] != "quiet" && fields[11] != "loaded" {
			return nil, fmt.Errorf("%s:%d: quiet is %q, want quiet or loaded", path, line, fields[11])
		}
		numbers := map[int]float64{}
		for _, index := range []int{1, 2, 3, 4, 5, 7, 8, 9, 10, 12, 13} {
			if numbers[index], err = strconv.ParseFloat(fields[index], 64); err != nil {
				return nil, fmt.Errorf("%s:%d: field %d: %w", path, line, index+1, err)
			}
		}
		runs = append(runs, benchresults.Run{
			Mode: mode, Round: int(numbers[1]), VerdictSeconds: numbers[2], SettledSeconds: numbers[3],
			EngineSeconds: numbers[4], LoadBefore: numbers[7], LoadAfter: numbers[8], IdleBefore: numbers[9], IdleAfter: numbers[10],
			Quiet: fields[11] == "quiet", Findings: int(numbers[5]), Cache: fields[6], Exit: int(numbers[12]), Primes: int(numbers[13]),
		})
	}
	return runs, scanner.Err()
}
