// Package benchresults is the recorded form of a bench/quiet_machine.sh run: one JSON file per run in
// bench/results, committed, which the website's benchmarks page reads (#rwgs6j2, for #7tmkt32).
//
// A quiet number is the only number the benchmark reports, so the record is shaped so that one cannot be
// claimed without being earned. A run's Quiet is the machine's measured idleness around it against the floor,
// nothing else. A mode carries a Quiet summary only when at least one of its runs was quiet, and the summary
// is what those runs give. Validate holds all three, and both the writer and every reader call it, so a
// record that disagrees with its own runs is refused wherever it is opened.
//
// # Schema 2 judges quiet by idleness, and schema 1 by load
//
// Schema 1 called a run quiet when the one-minute load average was under a ceiling before and after it. That
// average mostly measures cohere's own last run, decaying for 30 to 60 seconds, so every measured run waited
// up to a minute for it and the benchmark spent most of its wall time waiting (#bdqqfa1). Schema 2 samples
// the share of all cores idle over the seconds just before a run and just after it, and calls the run quiet
// when both are at or over the floor. It also records which modes were measured, since a run may measure
// cold alone, and whether it stopped early because its quiet runs already agreed. A schema 1 record is read
// and validated by its own rules, so what was committed under it still says what it said.
package benchresults

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// SchemaVersion is the version of the record's shape that is written. A reader accepts it and
// loadSchemaVersion, and refuses any other.
const SchemaVersion = 2

// loadSchemaVersion is the shape that judged quiet by the one-minute load against a ceiling.
const loadSchemaVersion = 1

// Directory is where records are committed, relative to the module root.
const Directory = "bench/results"

// Mode is what a measured run does: Cold runs with --no-cache, Warm replays an unchanged tree, Edit runs
// after a one-line comment is added to one function body, and Idle runs with --no-cache after the machine
// has sat idle, the case a developer who pauses between saves mostly meets.
type Mode string

const (
	ModeCold Mode = "Cold"
	ModeWarm Mode = "Warm"
	ModeEdit Mode = "Edit"
	ModeIdle Mode = "Idle"
)

// Modes are every mode a record can measure, in the order the report prints them.
var Modes = []Mode{ModeCold, ModeWarm, ModeEdit, ModeIdle}

// loadSchemaModes are the modes a schema 1 record always measured.
var loadSchemaModes = []Mode{ModeCold, ModeWarm, ModeEdit}

// Record is one benchmark run, recorded.
type Record struct {
	Schema     int     `json:"schema"`
	RecordedAt string  `json:"recordedAt"`
	Cohere     Cohere  `json:"cohere"`
	Project    Project `json:"project"`
	Machine    Machine `json:"machine"`

	// LoadCeiling is, in a schema 1 record, the highest one-minute load a quiet run may start or end at.
	LoadCeiling float64 `json:"loadCeiling,omitempty"`
	// IdleFloor is, in a schema 2 record, the share of all cores, in percent, that must be idle over the
	// seconds before a run and after it for the run to be quiet.
	IdleFloor float64 `json:"idleFloor,omitempty"`
	// SettleSeconds is the longest each measured run waited for the machine to be quiet before it ran anyway.
	SettleSeconds float64 `json:"settleSeconds"`
	// Rounds is how many runs of each measured mode were taken.
	Rounds int `json:"rounds"`
	// MeasuredModes is, in a schema 2 record, the modes the run measured, each Rounds times.
	MeasuredModes []Mode `json:"measuredModes,omitempty"`
	// Band is, in a schema 2 record, the agreement that let the run stop early: every measured mode's quiet
	// runs within this percent of their median, worst to best. Zero means the run never stops early.
	Band float64 `json:"band,omitempty"`
	// StoppedEarly is true when the run stopped before its last round because its quiet runs agreed.
	StoppedEarly bool `json:"stoppedEarly,omitempty"`
	// Edit is the file the Edit runs changed, relative to the project, when Edit was measured.
	Edit string `json:"edit,omitempty"`

	Runs  []Run                `json:"runs"`
	Modes map[Mode]ModeSummary `json:"modes"`
}

// Cohere is the engine every run used.
type Cohere struct {
	// Commit is the full commit the engine was built from.
	Commit string `json:"commit"`
	// Version is the first line of the engine's --version.
	Version string `json:"version"`
	// Dirty is true for an engine built from a tree with uncommitted changes, which no commit reproduces.
	Dirty bool `json:"dirty"`
}

// Project is the project measured, as the copy pinned it.
type Project struct {
	Name   string `json:"name"`
	Commit string `json:"commit"`
}

// Machine is the hardware the runs were measured on. It names no host and no person.
type Machine struct {
	Model       string `json:"model"`
	Processor   string `json:"processor"`
	Cores       int    `json:"cores"`
	MemoryBytes int64  `json:"memoryBytes"`
	OS          string `json:"os"`
}

// Run is one measured run.
type Run struct {
	Mode  Mode `json:"mode"`
	Round int  `json:"round"`
	// VerdictSeconds is what a developer waits for: until the engine has its exit code.
	VerdictSeconds float64 `json:"verdictSeconds"`
	// SettledSeconds is until every process the run started has exited, its cache write included.
	SettledSeconds float64 `json:"settledSeconds"`
	// EngineSeconds is the engine's own wall clock, from its --json summary.
	EngineSeconds float64 `json:"engineSeconds"`
	// LoadBefore and LoadAfter are the one-minute load around the run. In a schema 1 record they decide
	// Quiet; in schema 2 they are kept as what they are, a slow average, beside the idleness that decides.
	LoadBefore float64 `json:"loadBefore"`
	LoadAfter  float64 `json:"loadAfter"`
	// IdleBefore and IdleAfter are, in a schema 2 record, the share of all cores idle, in percent, over the
	// seconds just before the run and just after it.
	IdleBefore float64 `json:"idleBefore,omitempty"`
	IdleAfter  float64 `json:"idleAfter,omitempty"`
	// Quiet is true when the run was measured quiet: by the record's schema, its idleness at or over the floor
	// on both sides, or in schema 1 its load at or under the ceiling on both sides.
	Quiet bool `json:"quiet"`
	// Findings and Cache are what the run reported about itself (RunOutput).
	Findings int    `json:"findings"`
	Cache    string `json:"cache"`
	Exit     int    `json:"exit"`
	// Primes is how many unmeasured runs the round needed before one replayed whole, from 1 to 3, the same
	// on each of the round's runs. Two is expected after the first round, since the edit run's cache entry
	// holds the edited bytes and the next run rechecks the restored file and its importers. A schema 2 round
	// that measures neither Warm nor Edit reads no cache and primes nothing, so it records 0.
	Primes int `json:"primes"`
}

// ModeSummary is one mode's verdict seconds.
type ModeSummary struct {
	// Quiet is the quiet runs' best, median and worst, and is absent when no run of the mode was quiet.
	Quiet *QuietSummary `json:"quiet,omitempty"`
	// LoadedVerdictSeconds are the loaded runs' verdict seconds, sorted: shown, never reported as the number.
	LoadedVerdictSeconds []float64 `json:"loadedVerdictSeconds"`
}

// QuietSummary is a mode's number: taken from quiet runs only.
type QuietSummary struct {
	Count  int     `json:"count"`
	Best   float64 `json:"best"`
	Median float64 `json:"median"`
	Worst  float64 `json:"worst"`
}

// Summarize is the summaries runs give for each of modes, the same way quiet_machine.sh prints them: the
// median of an even count is the lower middle.
func Summarize(modes []Mode, runs []Run) map[Mode]ModeSummary {
	summaries := map[Mode]ModeSummary{}
	for _, mode := range modes {
		quiet, loaded := []float64{}, []float64{}
		for _, run := range runs {
			if run.Mode != mode {
				continue
			}
			if run.Quiet {
				quiet = append(quiet, run.VerdictSeconds)
			} else {
				loaded = append(loaded, run.VerdictSeconds)
			}
		}
		sort.Float64s(quiet)
		sort.Float64s(loaded)
		summary := ModeSummary{LoadedVerdictSeconds: loaded}
		if len(quiet) > 0 {
			summary.Quiet = &QuietSummary{Count: len(quiet), Best: quiet[0], Median: quiet[(len(quiet)-1)/2], Worst: quiet[len(quiet)-1]}
		}
		summaries[mode] = summary
	}
	return summaries
}

// MeasuredModesOf is the modes a record measured: those it names, or under schema 1 the three it always did.
func MeasuredModesOf(record Record) []Mode {
	if record.Schema == loadSchemaVersion {
		return loadSchemaModes
	}
	return record.MeasuredModes
}

// Validate refuses a record that is not the honest account of its own runs: a run called quiet whose
// idleness was under the floor (in schema 1, whose loads were over the ceiling), or the reverse; a quiet
// summary that is not what the quiet runs give, including one on a mode with no quiet run; a measured mode
// with a run count other than the rounds, or a run of a mode it did not measure; a run whose cache use its
// mode cannot have, which means the run was not what its mode says, or the instrument read nothing; an early
// stop its quiet runs do not justify. A dirty engine is refused too, since no commit reproduces its numbers.
func Validate(record Record) error {
	if record.Schema != SchemaVersion && record.Schema != loadSchemaVersion {
		return fmt.Errorf("schema %d, want %d or %d", record.Schema, SchemaVersion, loadSchemaVersion)
	}
	if record.Cohere.Dirty {
		return fmt.Errorf("cohere %s was built from a modified tree, which no commit reproduces", record.Cohere.Commit)
	}
	if len(record.Cohere.Commit) != 40 || len(record.Project.Commit) != 40 {
		return fmt.Errorf("cohere commit %q and project commit %q must each be a full commit", record.Cohere.Commit, record.Project.Commit)
	}
	byLoad := record.Schema == loadSchemaVersion
	if record.Rounds < 1 {
		return fmt.Errorf("%d rounds measure nothing", record.Rounds)
	}
	if byLoad && record.LoadCeiling <= 0 {
		return fmt.Errorf("a load ceiling of %g measures nothing", record.LoadCeiling)
	}
	if !byLoad && (record.IdleFloor <= 0 || record.IdleFloor > 100) {
		return fmt.Errorf("an idle floor of %g%% is not a share of the cores", record.IdleFloor)
	}
	measured := MeasuredModesOf(record)
	if err := knownModes(measured); err != nil {
		return err
	}
	if measures(measured, ModeEdit) != (record.Edit != "") {
		return fmt.Errorf("the record names edit file %q, and Edit measured is %v", record.Edit, measures(measured, ModeEdit))
	}
	if !byLoad && record.Band < 0 {
		return fmt.Errorf("a band of %g%%", record.Band)
	}
	// A round reads the cache only to measure Warm or Edit; a round that measures neither primes nothing.
	primesFloor, primesCeiling := 1, 3
	if !byLoad && !measures(measured, ModeWarm) && !measures(measured, ModeEdit) {
		primesFloor, primesCeiling = 0, 0
	}
	counts := map[Mode]int{}
	primesByRound := map[int]int{}
	for _, run := range record.Runs {
		counts[run.Mode]++
		if byLoad {
			quiet := run.LoadBefore <= record.LoadCeiling && run.LoadAfter <= record.LoadCeiling
			if run.Quiet != quiet {
				return fmt.Errorf("%s round %d is marked quiet %v at loads %g and %g against a ceiling of %g", run.Mode, run.Round,
					run.Quiet, run.LoadBefore, run.LoadAfter, record.LoadCeiling)
			}
		} else {
			quiet := run.IdleBefore >= record.IdleFloor && run.IdleAfter >= record.IdleFloor
			if run.Quiet != quiet {
				return fmt.Errorf("%s round %d is marked quiet %v at %g%% and %g%% idle against a floor of %g%%", run.Mode, run.Round,
					run.Quiet, run.IdleBefore, run.IdleAfter, record.IdleFloor)
			}
		}
		if run.VerdictSeconds <= 0 || run.SettledSeconds < run.VerdictSeconds {
			return fmt.Errorf("%s round %d settled at %gs before or without its verdict at %gs", run.Mode, run.Round,
				run.SettledSeconds, run.VerdictSeconds)
		}
		if run.EngineSeconds <= 0 {
			return fmt.Errorf("%s round %d has no engine seconds, so its output was not read", run.Mode, run.Round)
		}
		if run.Primes < primesFloor || run.Primes > primesCeiling {
			return fmt.Errorf("%s round %d needed %d primes, and a round measuring %v runs %d to %d", run.Mode, run.Round, run.Primes,
				measured, primesFloor, primesCeiling)
		}
		if primes, seen := primesByRound[run.Round]; seen && primes != run.Primes {
			return fmt.Errorf("round %d's runs disagree on its primes, %d and %d", run.Round, primes, run.Primes)
		}
		primesByRound[run.Round] = run.Primes
		if err := cacheFitsMode(run); err != nil {
			return fmt.Errorf("%s round %d: %w", run.Mode, run.Round, err)
		}
	}
	for _, mode := range measured {
		if counts[mode] != record.Rounds {
			return fmt.Errorf("%d %s runs for %d rounds", counts[mode], mode, record.Rounds)
		}
	}
	if len(counts) != len(measured) {
		return fmt.Errorf("runs of a mode other than %v", measured)
	}
	if record.StoppedEarly {
		if byLoad || record.Band <= 0 {
			return fmt.Errorf("the record stopped early with no band to agree within")
		}
		if err := agree(Summarize(measured, record.Runs), measured, record.Band); err != nil {
			return fmt.Errorf("the record stopped early, but %w", err)
		}
	}
	want, err := json.Marshal(Summarize(measured, record.Runs))
	if err != nil {
		return err
	}
	got, err := json.Marshal(record.Modes)
	if err != nil {
		return err
	}
	if string(got) != string(want) {
		return fmt.Errorf("the mode summaries are not what the runs give:\n  recorded %s\n  the runs %s", got, want)
	}
	return nil
}

// Agreement is how many quiet runs a mode needs before its runs can be said to agree.
const Agreement = 3

// agree refuses an early stop the quiet runs do not justify: every measured mode needs at least Agreement
// quiet runs, from best to worst within band percent of their median.
func agree(summaries map[Mode]ModeSummary, measured []Mode, band float64) error {
	for _, mode := range measured {
		quiet := summaries[mode].Quiet
		if quiet == nil || quiet.Count < Agreement {
			return fmt.Errorf("%s has fewer than %d quiet runs", mode, Agreement)
		}
		if quiet.Worst-quiet.Best > quiet.Median*band/100 {
			return fmt.Errorf("%s's quiet runs spread %gs to %gs, more than %g%% of %gs", mode, quiet.Best, quiet.Worst, band, quiet.Median)
		}
	}
	return nil
}

// knownModes refuses a measured list that is empty, names a mode no benchmark measures, or names one twice.
func knownModes(measured []Mode) error {
	if len(measured) == 0 {
		return fmt.Errorf("no measured modes")
	}
	seen := map[Mode]bool{}
	for _, mode := range measured {
		if !measures(Modes, mode) || seen[mode] {
			return fmt.Errorf("measured modes %v: %q is unknown or repeated", measured, mode)
		}
		seen[mode] = true
	}
	return nil
}

func measures(measured []Mode, mode Mode) bool {
	for _, each := range measured {
		if each == mode {
			return true
		}
	}
	return false
}

// cacheFitsMode refuses a cache use the run's mode cannot have: Cold and Idle run with --no-cache, so their
// cache is off; Warm replays an unchanged tree after a prime that replayed, so it is a whole replay; Edit changed one
// file, so some files and not all of them were taken from the cache.
func cacheFitsMode(run Run) error {
	replayed, inScope := 0, 0
	partial := false
	if _, err := fmt.Sscanf(run.Cache, "files %d/%d", &replayed, &inScope); err == nil {
		partial = replayed > 0 && replayed < inScope
	}
	switch {
	case (run.Mode == ModeCold || run.Mode == ModeIdle) && run.Cache == "off":
	case run.Mode == ModeWarm && run.Cache == "replay":
	case run.Mode == ModeEdit && partial:
	default:
		return fmt.Errorf("its cache read %q, which a %s run cannot have", run.Cache, run.Mode)
	}
	return nil
}

// Encode is a record as it is committed: validated, indented, with a final newline.
func Encode(record Record) ([]byte, error) {
	if err := Validate(record); err != nil {
		return nil, err
	}
	encoded, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(encoded, '\n'), nil
}

// FileName is the record's file name: its date, the engine's commit and the machine's model.
func FileName(record Record) string {
	model := strings.Map(func(character rune) rune {
		switch {
		case character >= 'a' && character <= 'z', character >= '0' && character <= '9':
			return character
		case character >= 'A' && character <= 'Z':
			return character + 'a' - 'A'
		}
		return '-'
	}, record.Machine.Model)
	return fmt.Sprintf("%s-%s-%s.json", record.RecordedAt[:10], record.Cohere.Commit[:12], model)
}

// Read reads every record in directory, sorted by file name, and refuses the first that does not
// validate. A directory that does not exist holds no records.
func Read(directory string) ([]Record, error) {
	paths, err := filepath.Glob(filepath.Join(directory, "*.json"))
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	records := []Record{}
	for _, path := range paths {
		contents, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var record Record
		decoder := json.NewDecoder(strings.NewReader(string(contents)))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&record); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		if err := Validate(record); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		if name := FileName(record); filepath.Base(path) != name {
			return nil, fmt.Errorf("%s: a record of this run is named %s", path, name)
		}
		records = append(records, record)
	}
	return records, nil
}
