package main

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/system-inc/cohere/internal/edit"
	"github.com/system-inc/cohere/internal/release/packaging"
)

// swiftContractVersion is the version of swift/Contract.md this front door speaks.
const swiftContractVersion = 3

// versionReport is what `cohere --version` prints about cohere itself: its provenance, then the Swift
// contract it speaks, on every platform and every build, since it is true of each. The release reads
// that line from the cohere it stages and refuses a cohere-swift that will not speak the same version,
// so the number comes from the binary that ships rather than from a copy that could drift.
func versionReport() string {
	return fmt.Sprintf("%s\n  swift contract: %d", release.Current(), swiftContractVersion)
}

// swiftMode is which question a Swift run answers, and so which records it may carry.
//
// The listing modes end without a summary by design: `--version` prints one provenance record and
// `--rules` prints rule records, as the contract says. A run in the checking mode that ends without a
// summary is a crash. The mode is what lets one absence mean two different things without the front
// door guessing which.
type swiftMode string

const (
	swiftModeCheck        swiftMode = "Check"
	swiftModeVersion      swiftMode = "Version"
	swiftModeRules        swiftMode = "Rules"
	swiftModeRulesEnabled swiftMode = "RulesEnabled"
)

// The records, as swift/Contract.md defines them. Fields the contract does not list are ignored by
// leaving them out of these types, which is what the contract asks of the front door.

type swiftProvenanceRecord struct {
	Contract           *int   `json:"contract"`
	Engine             string `json:"engine"`
	Version            string `json:"version"`
	Commit             string `json:"commit"`
	SourceTreeModified bool   `json:"sourceTreeModified"`
	Toolchain          string `json:"toolchain"`
	SwiftSyntax        string `json:"swiftSyntax"`
	SwiftFormat        string `json:"swiftFormat"`
}

type swiftProjectRecord struct {
	Root                string `json:"root"`
	Package             string `json:"package"`
	ElapsedMilliseconds int    `json:"elapsedMilliseconds"`
	FilesInPackage      int    `json:"filesInPackage"`
	FilesOurs           int    `json:"filesOurs"`
	FilesInScope        int    `json:"filesInScope"`
	ScopeDescription    string `json:"scopeDescription"`
	Excluded            []struct {
		File   string `json:"file"`
		Reason string `json:"reason"`
	} `json:"excluded"`
}

// swiftUnreadableRecord is a file the engine could not read, so nothing in it was checked. Contract 2
// made it a record: before, the engine said only on stderr that it had fallen short, and the front door
// could believe it but not say which files.
type swiftUnreadableRecord struct {
	File  string `json:"file"`
	Error string `json:"error"`
}

type swiftFindingRecord struct {
	Source    string `json:"source"`
	File      string `json:"file"`
	Line      int    `json:"line"`
	Column    int    `json:"column"`
	Severity  string `json:"severity"`
	Rule      string `json:"rule"`
	MessageId string `json:"messageId"`
	Message   string `json:"message"`
}

// swiftUnusedRecord is one item of the --unused report. It carries a finding's place and words but is
// not a finding: the report is not a gate, so it is never counted and never moves the exit code, as the
// TypeScript unused report never does.
type swiftUnusedRecord struct {
	File      string `json:"file"`
	Line      int    `json:"line"`
	Column    int    `json:"column"`
	Rule      string `json:"rule"`
	MessageId string `json:"messageId"`
	Message   string `json:"message"`
	Subject   string `json:"subject"`
}

// swiftUnusedCoverageRecord is what one rule of the --unused report looked at, so a report that checked
// nothing never reads like one that found nothing.
type swiftUnusedCoverageRecord struct {
	Rule                string         `json:"rule"`
	FilesChecked        int            `json:"filesChecked"`
	FilesNotChecked     map[string]int `json:"filesNotChecked"`
	Checked             int            `json:"checked"`
	Skipped             map[string]int `json:"skipped"`
	Found               *int           `json:"found"`
	ElapsedMilliseconds int            `json:"elapsedMilliseconds"`
}

type swiftFixRecord struct {
	FilesConsidered     int            `json:"filesConsidered"`
	FilesRewritten      int            `json:"filesRewritten"`
	FixesApplied        int            `json:"fixesApplied"`
	FixesRefused        int            `json:"fixesRefused"`
	RefusalsByReason    map[string]int `json:"refusalsByReason"`
	FilesReformatted    int            `json:"filesReformatted"`
	FilesNotFormatted   int            `json:"filesNotFormatted"`
	NotFormattedReasons map[string]int `json:"notFormattedReasons"`
	FormatScope         string         `json:"formatScope"`
	// ChangedFiles is every file the phase wrote, once each. Additive within contract 3: an engine from
	// before it sends none, which is "not listed", never "no files changed", so it is a pointer.
	ChangedFiles *[]swiftChangedFile `json:"changedFiles"`
}

// swiftChangedFile is one file the fix phase wrote: its fixes by the rule that made them, and whether the
// formatter changed it.
type swiftChangedFile struct {
	File      string         `json:"file"`
	FixedBy   map[string]int `json:"fixedBy"`
	Formatted bool           `json:"formatted"`
}

type swiftTypesRecord struct {
	Diagnostics         int      `json:"diagnostics"`
	Files               int      `json:"files"`
	ElapsedMilliseconds int      `json:"elapsedMilliseconds"`
	FilesWithoutRecord  []string `json:"filesWithoutRecord"`
}

type swiftLintRecord struct {
	Findings             int            `json:"findings"`
	RulesRun             int            `json:"rulesRun"`
	FilesWalked          int            `json:"filesWalked"`
	NodesVisited         int            `json:"nodesVisited"`
	ElapsedMilliseconds  int            `json:"elapsedMilliseconds"`
	ReusedFrom           string         `json:"reusedFrom"`
	RulesSilent          []string       `json:"rulesSilent"`
	RulesWatchedAndQuiet int            `json:"rulesWatchedAndQuiet"`
	RulesScopedOff       map[string]int `json:"rulesScopedOff"`
	RulesNotConfigured   []string       `json:"rulesNotConfigured"`
	ConfigNote           string         `json:"configNote"`
	Crashes              []struct {
		File  string `json:"file"`
		Error string `json:"error"`
	} `json:"crashes"`
}

type swiftPhaseRecord struct {
	Name                string `json:"name"`
	Outcome             string `json:"outcome"`
	ElapsedMilliseconds int    `json:"elapsedMilliseconds"`
	Detail              string `json:"detail"`
}

type swiftRuleRecord struct {
	Name     string `json:"name"`
	Severity string `json:"severity"`
}

type swiftSummaryRecord struct {
	Findings *int  `json:"findings"`
	Complete *bool `json:"complete"`
	ExitCode *int  `json:"exitCode"`
}

// swiftOutcomes maps the contract's camelCase outcomes onto the ones phases.go prints.
var swiftOutcomes = map[string]phaseOutcome{
	"ran":        outcomeRan,
	"skipped":    outcomeSkipped,
	"notReached": outcomeNotReached,
	"reused":     outcomeReused,
}

// swiftRun renders one engine run's records as they arrive, through the printers a TypeScript run
// uses, and refuses any stream that breaks the contract.
//
// Every refusal is an error rather than a skipped line, for the reason the contract gives: a skipped
// record is a finding nobody sees. And every count the stream states twice is compared, because two
// numbers that should agree and do not are a defect in one of them, and printing either would be
// printing a number nobody can vouch for.
type swiftRun struct {
	out  io.Writer
	mode swiftMode

	// details is `--coverage`: name every rule once under its coverage fact rather than only counting.
	// A front-door flag, never forwarded: the engine's records already carry what it names.
	details bool

	// engineSourceCommit is the commit whose Swift sources the engine is guaranteed to be built from,
	// or empty when nothing vouches for it. Set only when the engine came through the launcher's
	// committed path, whose cache key is the object ids of that commit's Swift inputs. The engine itself
	// says `commit dev` and must: one cached engine serves every commit with identical Swift inputs, so a
	// commit baked into it would name whichever commit built it first, true of the build and false of
	// the run. The front door knows the run.
	engineSourceCommit string

	// requested is which phases the caller asked for, by the flags the TypeScript run reads the same way
	// (swiftRequestedPhases). An engine that reports one of them skipped as not requested is refused: it
	// would turn a phase that was asked for and never ran into a run that looks complete. Nil, as in a test
	// that renders a stream with no command line, checks nothing.
	requested map[phaseName]bool

	provenance *swiftProvenanceRecord
	project    *swiftProjectRecord
	summary    *swiftSummaryRecord
	report     *pipelineReport
	rules      []swiftRuleRecord

	// findings is every finding record received; findingsInPhase only those since the last phase
	// record, which is how the types and lint records' own counts are checked.
	findings        int
	findingsInPhase int

	// nextPhase indexes phaseOrder: phases must arrive in pipeline order, each exactly once.
	nextPhase int

	// Gaps the records name outside the phase outcomes: files the compiler left no record for, and
	// files a rule crashed on. Each is already printed as a note; these count them for the summary.
	filesWithoutRecord int
	crashes            int
	unreadable         []swiftUnreadableRecord

	excludedPrinted bool

	// The --unused report: items held until their rule's coverage record says how many there are, so the
	// section prints with its count first, in the TypeScript report's shape. unusedCoverage counts the
	// coverage records, which an unused phase that ran must have.
	unusedHeadingPrinted bool
	unusedPending        []swiftUnusedRecord
	unusedCoverage       int
}

func newSwiftRun(out io.Writer, mode swiftMode, rootNote string, processStart time.Time) *swiftRun {
	return &swiftRun{
		out:  out,
		mode: mode,
		report: &pipelineReport{
			processStart:          processStart,
			rootNote:              rootNote,
			graphLabel:            "package",
			graphNotBuiltSentence: "no package was described and no phase ran",
		},
	}
}

// accept renders one line of the engine's stdout.
func (r *swiftRun) accept(line []byte) error {
	var envelope struct {
		Kind string `json:"kind"`
	}
	if err := json.Unmarshal(line, &envelope); err != nil {
		return fmt.Errorf("a line that is not a record: %q (%v)", truncateForError(line), err)
	}

	if r.summary != nil {
		return fmt.Errorf("a %q record after the summary, which is always last", envelope.Kind)
	}
	if r.provenance == nil && envelope.Kind != "provenance" {
		return fmt.Errorf("the first record was %q, and provenance is always first", envelope.Kind)
	}

	switch envelope.Kind {
	case "provenance":
		return decodeRecord(line, &swiftProvenanceRecord{}, r.acceptProvenance)
	case "project":
		return decodeRecord(line, &swiftProjectRecord{}, r.acceptProject)
	case "unreadable":
		return decodeRecord(line, &swiftUnreadableRecord{}, r.acceptUnreadable)
	case "finding":
		return decodeRecord(line, &swiftFindingRecord{}, r.acceptFinding)
	case "fix":
		return decodeRecord(line, &swiftFixRecord{}, r.acceptFix)
	case "types":
		return decodeRecord(line, &swiftTypesRecord{}, r.acceptTypes)
	case "lint":
		return decodeRecord(line, &swiftLintRecord{}, r.acceptLint)
	case "phase":
		return decodeRecord(line, &swiftPhaseRecord{}, r.acceptPhase)
	case "rule":
		return decodeRecord(line, &swiftRuleRecord{}, r.acceptRule)
	case "unused":
		return decodeRecord(line, &swiftUnusedRecord{}, r.acceptUnused)
	case "unusedCoverage":
		return decodeRecord(line, &swiftUnusedCoverageRecord{}, r.acceptUnusedCoverage)
	case "summary":
		return decodeRecord(line, &swiftSummaryRecord{}, r.acceptSummary)
	default:
		// An unknown kind is refused rather than ignored: it is a record this front door cannot render,
		// so whatever it carried would reach nobody.
		return fmt.Errorf("a record of unknown kind %q", envelope.Kind)
	}
}

// decodeRecord decodes a line into its record type and hands it on.
func decodeRecord[Record any](line []byte, record *Record, accept func(*Record) error) error {
	if err := json.Unmarshal(line, record); err != nil {
		return fmt.Errorf("a record that does not decode: %q (%v)", truncateForError(line), err)
	}
	return accept(record)
}

func (r *swiftRun) acceptProvenance(record *swiftProvenanceRecord) error {
	if r.provenance != nil {
		return fmt.Errorf("a second provenance record")
	}
	// Checked here as well as by the engine against `--contract`, because the flag and the record are
	// two independent ways for the versions to disagree, and the contract asks for both.
	if record.Contract == nil || *record.Contract != swiftContractVersion {
		spoken := "none"
		if record.Contract != nil {
			spoken = fmt.Sprint(*record.Contract)
		}
		return fmt.Errorf("the engine speaks contract %s and this front door speaks contract %d", spoken, swiftContractVersion)
	}
	r.provenance = record
	r.report.engineSourceTreeModified = record.SourceTreeModified
	return nil
}

func (r *swiftRun) acceptProject(record *swiftProjectRecord) error {
	if r.mode != swiftModeCheck {
		return fmt.Errorf("a project record in a %s run", r.mode)
	}
	if r.project != nil {
		return fmt.Errorf("a second project record")
	}
	r.project = record

	elapsed := time.Duration(record.ElapsedMilliseconds) * time.Millisecond
	line := fmt.Sprintf("package described in %s — %d Swift files in the package, %d of them ours",
		round(elapsed), record.FilesInPackage, record.FilesOurs)
	if record.FilesInScope != record.FilesOurs || record.ScopeDescription != "" {
		line += fmt.Sprintf(", %d in scope (%s)", record.FilesInScope, record.ScopeDescription)
	}
	fmt.Fprintln(accountOutput(r.out), line)
	activeSummary.FilesInScope, activeSummary.FilesChecked = record.FilesInScope, record.FilesInScope
	activeSummary.Gaps.NothingToCheck = record.FilesInScope == 0

	r.report.graph = elapsed
	r.report.filesInScope = record.FilesInScope
	r.report.filesInProgram = record.FilesOurs
	return nil
}

// acceptUnreadable holds a file nothing checked, to be named with the excluded files and counted as a
// gap. Its place is fixed, after the project and before the first phase, so the files a run could not
// read are all known before any phase reports on the files it could.
func (r *swiftRun) acceptUnreadable(record *swiftUnreadableRecord) error {
	if r.mode != swiftModeCheck {
		return fmt.Errorf("an unreadable record in a %s run", r.mode)
	}
	if r.project == nil {
		return fmt.Errorf("an unreadable record before the project record")
	}
	if r.nextPhase > 0 {
		return fmt.Errorf("an unreadable record after phase %s, and contract 2 places them before the first phase", phaseOrder[r.nextPhase-1])
	}
	if record.File == "" {
		return fmt.Errorf("an unreadable record with no file")
	}
	r.unreadable = append(r.unreadable, *record)
	return nil
}

func (r *swiftRun) acceptFinding(record *swiftFindingRecord) error {
	if r.mode != swiftModeCheck {
		return fmt.Errorf("a finding record in a %s run", r.mode)
	}
	if record.File == "" || record.Line < 1 || record.Column < 1 {
		return fmt.Errorf("a finding with no position (file %q, line %d, column %d)", record.File, record.Line, record.Column)
	}
	r.findings++
	r.findingsInPhase++
	// A compiler finding is a type error, as the TypeScript run counts its own; the rest are findings.
	// The contract spells the severities error and warning; the view spells a warning warn, as the settings do.
	severity := "error"
	if record.Severity == "warning" {
		severity = "warn"
	}
	rule, messageID := record.Rule, record.MessageId
	if record.Source == "compiler" {
		activeSummary.TypeErrors++
		if rule != "" {
			rule = "#" + rule
		} else {
			rule = "swiftc"
		}
	} else {
		activeSummary.Findings++
	}
	printFinding(r.out,
		runFinding{Path: record.File, Line: record.Line, Column: record.Column, Severity: severity, Rule: rule, MessageID: messageID, Message: singleLineDescription(record.Message)},
		swiftFindingLine(record)+"\n")
	return nil
}

// swiftFindingLine is one finding as one line, in the shape the TypeScript printers use.
//
// A compiler finding reads like tsc's `file:line:col - error TS2322: message`, with the warning group
// where the code goes, as swiftc writes it. A rule or format finding reads like a TypeScript rule
// finding, `[rule/messageId]` at the end. Newlines in the message are collapsed, for the reason
// printRuleDiagnostic gives.
func swiftFindingLine(record *swiftFindingRecord) string {
	position := fmt.Sprintf("%s:%d:%d", record.File, record.Line, record.Column)
	message := singleLineDescription(record.Message)
	if record.Source == "compiler" {
		line := fmt.Sprintf("%s - %s: %s", position, record.Severity, message)
		if record.Rule != "" {
			line += fmt.Sprintf(" [#%s]", record.Rule)
		}
		return line
	}
	return fmt.Sprintf("%s - %s [%s/%s]", position, message, record.Rule, record.MessageId)
}

// unusedPlace refuses an --unused report record anywhere but between the lint phase record and the
// unused one, where the contract puts the report.
func (r *swiftRun) unusedPlace(kind string) error {
	if r.mode != swiftModeCheck {
		return fmt.Errorf("an %s record in a %s run", kind, r.mode)
	}
	if r.nextPhase >= len(phaseOrder) || phaseOrder[r.nextPhase] != phaseUnused {
		return fmt.Errorf("an %s record outside the unused phase, which comes after lint", kind)
	}
	return nil
}

func (r *swiftRun) acceptUnused(record *swiftUnusedRecord) error {
	if err := r.unusedPlace("unused"); err != nil {
		return err
	}
	if record.File == "" || record.Line < 1 || record.Column < 1 || record.Rule == "" || record.Subject == "" {
		return fmt.Errorf("an unused record missing its file, position, rule or subject (file %q, line %d, column %d, rule %q)",
			record.File, record.Line, record.Column, record.Rule)
	}
	r.unusedPending = append(r.unusedPending, *record)
	return nil
}

// acceptUnusedCoverage prints one rule's section of the report: its items, counted and sorted, then what
// it looked at and what it could not.
func (r *swiftRun) acceptUnusedCoverage(record *swiftUnusedCoverageRecord) error {
	if err := r.unusedPlace("unusedCoverage"); err != nil {
		return err
	}
	if record.Rule == "" || record.Found == nil {
		return fmt.Errorf("an unusedCoverage record missing its rule or found count")
	}
	items := []swiftUnusedRecord{}
	for _, pending := range r.unusedPending {
		if pending.Rule == record.Rule {
			items = append(items, pending)
		}
	}
	if len(items) != len(r.unusedPending) {
		return fmt.Errorf("unused records of another rule arrived before the coverage record for %s", record.Rule)
	}
	if *record.Found != len(items) {
		return fmt.Errorf("the unusedCoverage record for %s counts %d found and %d unused records came with it", record.Rule, *record.Found, len(items))
	}
	r.unusedPending = nil
	r.unusedCoverage++

	if !r.unusedHeadingPrinted {
		r.unusedHeadingPrinted = true
		fmt.Fprintf(r.out, "\nunused: a report, not a gate — nothing here fails a build\n")
	}
	// File, line, then column, so two items on one line print in the same order on every run.
	sort.SliceStable(items, func(first, second int) bool {
		if items[first].File != items[second].File {
			return items[first].File < items[second].File
		}
		if items[first].Line != items[second].Line {
			return items[first].Line < items[second].Line
		}
		return items[first].Column < items[second].Column
	})
	if len(items) > 0 {
		fmt.Fprintf(r.out, "\n  %s — %d\n", swiftUnusedSection(record.Rule), len(items))
		for _, item := range items {
			fmt.Fprintf(r.out, "    %s:%d:%d — %s\n", item.File, item.Line, item.Column, item.Subject)
		}
	}
	fmt.Fprintf(r.out, "  looked at %d files and %d %s (%s)\n", record.FilesChecked, record.Checked, swiftUnusedNoun(record.Rule), record.Rule)
	for _, reason := range sortedKeys(record.FilesNotChecked) {
		fmt.Fprintf(r.out, "  not checked for unused %s: %d files (%s)\n", swiftUnusedNoun(record.Rule), record.FilesNotChecked[reason], reason)
	}
	for _, reason := range sortedKeys(record.Skipped) {
		fmt.Fprintf(r.out, "  never reported: %d %s (%s)\n", record.Skipped[reason], swiftUnusedNoun(record.Rule), reason)
	}
	return nil
}

// swiftUnusedWords holds each known rule's heading and the noun for what it judges, in the TypeScript
// report's words. A rule not listed here prints under its own id and counts "items", so a new rule never
// reads under another rule's heading.
var swiftUnusedWords = map[string]struct{ section, noun string }{
	"cohere-swift/correctness-no-unused-import":      {"imports nothing uses", "imports"},
	"cohere-swift/correctness-no-unused-declaration": {"declarations nothing uses", "declarations"},
}

// swiftUnusedSection is the heading of one rule's items.
func swiftUnusedSection(rule string) string {
	if words, known := swiftUnusedWords[rule]; known {
		return words.section
	}
	return rule
}

// swiftUnusedNoun names what a rule judges, for its coverage line.
func swiftUnusedNoun(rule string) string {
	if words, known := swiftUnusedWords[rule]; known {
		return words.noun
	}
	return "items"
}

func sortedKeys(counts map[string]int) []string {
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (r *swiftRun) acceptFix(record *swiftFixRecord) error {
	if r.mode != swiftModeCheck {
		return fmt.Errorf("a fix record in a %s run", r.mode)
	}
	// The engine's numbers fed to the summary the TypeScript run prints, so the line and its rules
	// (the refusal breakdown beside the refusal count, formatting counted apart from fixes) are the
	// ones edit.Summary already keeps.
	summary := edit.Summary{
		FilesConsidered:       record.FilesConsidered,
		FilesChanged:          record.FilesRewritten,
		FixesApplied:          record.FixesApplied,
		FixesRefused:          record.FixesRefused,
		RefusalsByReason:      record.RefusalsByReason,
		FilesTransformed:      record.FilesReformatted,
		FilesTransformSkipped: record.FilesNotFormatted,
		TransformSkipReasons:  record.NotFormattedReasons,
	}
	fmt.Fprintln(accountOutput(r.out), summary.String())
	fmt.Fprintf(accountOutput(r.out), "format scope: %s\n", record.FormatScope)

	// An engine from before the list sends none, and the run then states the counts alone: the footer
	// leaves out a cohered count it was not given, rather than print a zero it does not know.
	if record.ChangedFiles == nil {
		return nil
	}
	changed, err := r.changedFiles(*record.ChangedFiles, record.FilesRewritten)
	if err != nil {
		return err
	}
	activeSummary.Changed = changed
	printChangedFiles(r.out, changed)
	return nil
}

// changedFiles turns the fix record's list into the summary's, each path relative to the package root as a
// TypeScript run's are. It refuses a list whose length differs from the count the record states, a file
// named twice, and a file that neither a fix nor the formatter changed, since "N cohered" must be the list.
func (r *swiftRun) changedFiles(files []swiftChangedFile, rewritten int) ([]changedFile, error) {
	if len(files) != rewritten {
		return nil, fmt.Errorf("the fix record counts %d files rewritten and lists %d", rewritten, len(files))
	}
	root := ""
	if r.project != nil {
		root = r.project.Root
	}
	seen := map[string]bool{}
	changed := make([]changedFile, 0, len(files))
	for _, file := range files {
		if seen[file.File] {
			return nil, fmt.Errorf("the fix record lists %s twice", file.File)
		}
		seen[file.File] = true
		if len(file.FixedBy) == 0 && !file.Formatted {
			return nil, fmt.Errorf("the fix record lists %s as rewritten by no fix and not by the formatter", file.File)
		}
		path := file.File
		if relative, err := filepath.Rel(root, file.File); root != "" && err == nil && !strings.HasPrefix(relative, "..") {
			path = relative
		}
		changed = append(changed, changedFile{
			Path:      path,
			Fixed:     len(file.FixedBy) > 0,
			Formatted: file.Formatted,
			FixedBy:   file.FixedBy,
		})
	}
	return changed, nil
}

func (r *swiftRun) acceptTypes(record *swiftTypesRecord) error {
	if r.mode != swiftModeCheck {
		return fmt.Errorf("a types record in a %s run", r.mode)
	}
	if record.Diagnostics != r.findingsInPhase {
		return fmt.Errorf("the types record counts %d diagnostics and %d finding records came with it", record.Diagnostics, r.findingsInPhase)
	}
	elapsed := time.Duration(record.ElapsedMilliseconds) * time.Millisecond
	fmt.Fprintf(accountOutput(r.out), "types: %d diagnostics over %d files in %s\n", record.Diagnostics, record.Files, round(elapsed))

	// The compiler checks the whole module whatever the scope, so a scoped run's types line covers more
	// than the scope does. Said, so the larger number does not read as a scope that leaked.
	if r.project != nil && record.Files != r.project.FilesInScope {
		fmt.Fprintf(accountOutput(r.out), "  types: covered the whole package, not only the %d files in scope\n", r.project.FilesInScope)
	}
	for _, file := range record.FilesWithoutRecord {
		fmt.Fprintf(accountOutput(r.out), "  types: no compiler record for %s, so its diagnostics are unknown\n", file)
	}
	r.filesWithoutRecord += len(record.FilesWithoutRecord)
	return nil
}

func (r *swiftRun) acceptLint(record *swiftLintRecord) error {
	if r.mode != swiftModeCheck {
		return fmt.Errorf("a lint record in a %s run", r.mode)
	}
	if record.Findings != r.findingsInPhase {
		return fmt.Errorf("the lint record counts %d findings and %d finding records came with it", record.Findings, r.findingsInPhase)
	}

	walkCost := fmt.Sprintf("in %s", round(time.Duration(record.ElapsedMilliseconds)*time.Millisecond))
	if record.ReusedFrom != "" {
		walkCost = fmt.Sprintf("walked by the %s phase (nothing was rewritten, so its findings still hold)", record.ReusedFrom)
	}
	fmt.Fprintf(accountOutput(r.out), "lint: %d findings — %d rules over %d files, %d nodes visited, %s\n",
		record.Findings, record.RulesRun, record.FilesWalked, record.NodesVisited, walkCost)
	activeSummary.Rules = record.RulesRun
	activeSummary.FilesInScope, activeSummary.FilesChecked = record.FilesWalked, record.FilesWalked
	activeSummary.Nodes = record.NodesVisited
	activeSummary.Gaps.CrashedFiles += len(record.Crashes)

	// The same coverage block a TypeScript run prints, from the same writer, so one fact reads the same
	// whichever engine states it: one counted line whose terms add up, crashes in full, and each rule
	// named once behind --coverage.
	summary, err := classifySwiftCoverage(record)
	if err != nil {
		return err
	}
	writeCoverage(accountOutput(r.out), summary, r.details)
	r.crashes += len(record.Crashes)

	if record.ConfigNote != "" {
		fmt.Fprintf(accountOutput(r.out), "  %s\n", record.ConfigNote)
	}
	r.writeExcluded()
	return nil
}

// classifySwiftCoverage puts every rule the lint record accounts for into exactly one category.
//
// The record names its silent rules and only counts the rest: rulesRun is every rule, of which
// rulesWatchedAndQuiet listened and reported nothing and rulesSilent listened to no file, so what
// remains found something. A silent rule is described by why it was silent when the record says (not
// configured, or off by the config), and otherwise by the one sentence the record can back: the engine
// reports one silent set where the TypeScript walk can tell "offered nothing" from "declined
// everything", so it claims neither.
//
// A record whose parts add up to more than its whole is refused, for the reason every count stated
// twice is compared: the counted line would print arithmetic nobody can vouch for.
func classifySwiftCoverage(record *swiftLintRecord) (coverageSummary, error) {
	notConfigured := make(map[string]bool, len(record.RulesNotConfigured))
	for _, name := range record.RulesNotConfigured {
		notConfigured[name] = true
	}

	silent := append([]string(nil), record.RulesSilent...)
	sort.Strings(silent)
	isSilent := make(map[string]bool, len(silent))
	summary := coverageSummary{Unnamed: map[coverageCategory]int{}}
	for _, name := range silent {
		if isSilent[name] {
			return coverageSummary{}, fmt.Errorf("the lint record names rule %s as silent twice", name)
		}
		isSilent[name] = true
		entry := ruleCoverageEntry{Name: name}
		switch {
		case notConfigured[name]:
			entry.Category = coverageNotConfigured
		case record.RulesScopedOff[name] > 0:
			entry.Category = coverageOffByTheConfig
			entry.Details = []string{fmt.Sprintf("%d files", record.RulesScopedOff[name])}
		default:
			entry.Category = coverageListenedToNothing
		}
		summary.Entries = append(summary.Entries, entry)
	}

	found := record.RulesRun - record.RulesWatchedAndQuiet - len(silent)
	if record.RulesWatchedAndQuiet < 0 || found < 0 {
		return coverageSummary{}, fmt.Errorf("the lint record counts %d rules run, and %d watched and quiet plus %d silent is more than that",
			record.RulesRun, record.RulesWatchedAndQuiet, len(silent))
	}
	summary.Unnamed[coverageRanAndFoundNothing] = record.RulesWatchedAndQuiet
	summary.Unnamed[coverageFoundSomething] = found

	// A rule the record places among the counted ones and also names as off, or unconfigured, for some
	// files ran on the rest. Its category is the counted one; the partial fact is said once, by name.
	for name, files := range record.RulesScopedOff {
		if !isSilent[name] {
			summary.PartialFacts = append(summary.PartialFacts, fmt.Sprintf("%s is off by the config for %d files and ran on the rest", name, files))
		}
	}
	for name := range notConfigured {
		if !isSilent[name] {
			summary.PartialFacts = append(summary.PartialFacts, fmt.Sprintf("%s is not in the config for some files and ran on the rest", name))
		}
	}
	sort.Strings(summary.PartialFacts)

	for _, crash := range record.Crashes {
		summary.Crashes = append(summary.Crashes, coverageCrash{File: crash.File, Cause: crash.Error})
	}
	summary.sortEntries()
	return summary, nil
}

// excludedNamedLimit is how many excluded files of one reason are named before they are counted.
const excludedNamedLimit = 5

// writeExcluded says which files the package description left out, once, under the lint line when
// there is one and above the phase line when there is not: a file nobody checked is a coverage fact
// whether or not lint got to run.
//
// Grouped by reason, and counted past a handful. The first run on ahraos-macos printed 107 lines of
// `vendored under Vendor/` between the lint line and the verdict, which buries the two lines a reader
// must act on under a hundred that say one thing. A few files are named, because one file left out
// for a surprising reason is the case worth seeing; a hundred for the same reason are one fact.
func (r *swiftRun) writeExcluded() {
	if r.excludedPrinted || r.project == nil {
		return
	}
	r.excludedPrinted = true

	filesByReason := map[string][]string{}
	reasons := []string{}
	add := func(file string, reason string) {
		if _, seen := filesByReason[reason]; !seen {
			reasons = append(reasons, reason)
		}
		filesByReason[reason] = append(filesByReason[reason], file)
	}
	for _, excluded := range r.project.Excluded {
		add(excluded.File, excluded.Reason)
	}
	// A file the engine could not read is printed the way an excluded file is, because to a reader it is
	// the same fact: a file nothing checked, and why.
	for _, unreadable := range r.unreadable {
		add(unreadable.File, "could not be read: "+unreadable.Error)
	}
	sort.Strings(reasons)
	for _, reason := range reasons {
		files := filesByReason[reason]
		if len(files) > excludedNamedLimit {
			fmt.Fprintf(accountOutput(r.out), "  not checked: %d files (%s)\n", len(files), reason)
			continue
		}
		for _, file := range files {
			fmt.Fprintf(accountOutput(r.out), "  not checked: %s (%s)\n", file, reason)
		}
	}
}

func (r *swiftRun) acceptPhase(record *swiftPhaseRecord) error {
	if r.mode != swiftModeCheck {
		return fmt.Errorf("a phase record in a %s run", r.mode)
	}
	if r.nextPhase >= len(phaseOrder) || string(phaseOrder[r.nextPhase]) != record.Name {
		expected := "nothing more"
		if r.nextPhase < len(phaseOrder) {
			expected = string(phaseOrder[r.nextPhase])
		}
		return fmt.Errorf("a phase record for %q where the pipeline order expects %s", record.Name, expected)
	}
	outcome, known := swiftOutcomes[record.Outcome]
	if !known {
		return fmt.Errorf("phase %s has outcome %q, which is not one of ran, skipped, notReached, reused", record.Name, record.Outcome)
	}
	if phaseOrder[r.nextPhase] == phaseUnused {
		if len(r.unusedPending) > 0 {
			return fmt.Errorf("%d unused records arrived with no unusedCoverage record to count them", len(r.unusedPending))
		}
		if record.Outcome == "ran" && r.unusedCoverage == 0 {
			return fmt.Errorf("the unused phase ran and no unusedCoverage record said what it looked at")
		}
	}
	name := phaseOrder[r.nextPhase]
	if r.requested[name] && skippedAsNotRequested(phaseRecord{Outcome: outcome, Detail: record.Detail}) {
		return fmt.Errorf("the %s phase was asked for, and the engine skipped it as %q", name, record.Detail)
	}
	r.nextPhase++
	r.findingsInPhase = 0
	r.report.record(name, outcome, time.Duration(record.ElapsedMilliseconds)*time.Millisecond, record.Detail)
	return nil
}

// swiftRequestedPhases is which phases a command line asks for, read as the TypeScript run reads it: no
// phase flag asks for fix, types and lint; `--fix`, `--types` or `--lint` asks for only the ones named;
// and any `--unused` flag asks for unused as well. Takes the flag reader as a parameter so a test can
// state its flags, as swiftEngineArguments does.
func swiftRequestedPhases(given map[string]bool, value func(string) string) map[phaseName]bool {
	named := func(flag string) bool { return given[flag] && value(flag) == "true" }
	anyPhaseNamed := named("fix") || named("types") || named("lint")
	return map[phaseName]bool{
		phaseFix:    named("fix") || !anyPhaseNamed,
		phaseTypes:  named("types") || !anyPhaseNamed,
		phaseLint:   named("lint") || !anyPhaseNamed,
		phaseUnused: named("unused") || named("unused-all") || named("unused-deep"),
	}
}

func (r *swiftRun) acceptRule(record *swiftRuleRecord) error {
	if r.mode != swiftModeRules && r.mode != swiftModeRulesEnabled {
		return fmt.Errorf("a rule record in a %s run", r.mode)
	}
	r.rules = append(r.rules, *record)
	return nil
}

func (r *swiftRun) acceptSummary(record *swiftSummaryRecord) error {
	if r.mode != swiftModeCheck {
		return fmt.Errorf("a summary record in a %s run", r.mode)
	}
	if record.Findings == nil || record.Complete == nil || record.ExitCode == nil {
		return fmt.Errorf("a summary missing findings, complete or exitCode")
	}
	if r.nextPhase != len(phaseOrder) {
		// The reason markRemainingNotReached exists: a phase absent from the report reads the same as
		// one the reporter forgot.
		return fmt.Errorf("a summary with no record for phase %s", phaseOrder[r.nextPhase])
	}
	if *record.Findings != r.findings {
		return fmt.Errorf("the summary counts %d findings and %d finding records arrived", *record.Findings, r.findings)
	}

	// The two views of completeness are compared in one direction only. Where the records show a gap
	// and the summary claims none, the engine is printing green over work it did not do, which is the
	// failure this tool exists to stop, and the run is refused. The other direction is allowed: an engine
	// that says it fell short is believed, even where no record shows the gap.
	namedGaps := r.filesWithoutRecord > 0 || r.crashes > 0 || len(r.unreadable) > 0
	// Judged by the phases the caller asked for: `complete` means every requested phase ran, so `--lint`
	// with fix and types skipped as not requested is complete. Any other skip is still a gap.
	gapInRecords := !r.report.checkedEveryRequestedPhase() || namedGaps
	if gapInRecords && *record.Complete {
		return fmt.Errorf("the summary calls the run complete, and its own records show what it did not check")
	}
	if !*record.Complete && r.report.checkedEveryRequestedPhase() {
		if namedGaps {
			r.report.incompleteBeyondPhases = "the notes above name the files nothing checked"
			activeSummary.Gaps.Unread = swiftUncheckedFiles(r.filesWithoutRecord, len(r.unreadable))
		} else {
			r.report.incompleteBeyondPhases = "the Swift engine reported that it fell short without a record saying where"
			activeSummary.Gaps.Unread = "the Swift engine fell short without saying where"
		}
	}

	expectedExit := 0
	if r.findings > 0 || !*record.Complete {
		expectedExit = 1
	}
	if *record.ExitCode != expectedExit {
		return fmt.Errorf("the summary says exit %d, and %d findings in a run complete=%t means exit %d",
			*record.ExitCode, r.findings, *record.Complete, expectedExit)
	}

	r.summary = record
	r.writeExcluded()
	if r.project == nil {
		r.report.graphNotBuilt = true
	}
	writeRunEnd(r.report, r.out)
	return nil
}

// finish decides the run's exit code once the engine has exited, from what it wrote and how it ended.
//
// Every path that is not a validated summary, or a listing mode's expected ending, is a failure that
// says nothing was checked. There is no path through here where a run that did not finish prints
// green.
func (r *swiftRun) finish(engineExit int, ended string) (int, error) {
	if engineExit == 2 {
		// The engine's own sentence is already on stderr, passed through untouched.
		return 1, fmt.Errorf("the Swift engine did not run, so nothing was checked")
	}

	switch r.mode {
	case swiftModeVersion:
		if engineExit != 0 || r.provenance == nil {
			return 1, fmt.Errorf("the Swift engine %s without stating its version", ended)
		}
		fmt.Fprintln(r.out, versionReport())
		fmt.Fprintf(r.out, "swift engine: %s %s (commit %s, %s, swift-syntax %s, swift-format %s)\n",
			r.provenance.Engine, r.provenance.Version, r.provenance.Commit,
			r.provenance.Toolchain, r.provenance.SwiftSyntax, r.provenance.SwiftFormat)
		if r.provenance.SourceTreeModified {
			fmt.Fprintln(r.out, "  the Swift engine was built from a modified tree, so no commit reproduces it")
		}
		if r.engineSourceCommit != "" {
			fmt.Fprintf(r.out, "  the Swift engine was built from Swift sources identical to commit %s, as the launcher's cache key requires\n",
				release.ShortCommit(r.engineSourceCommit))
		}
		return 0, nil

	case swiftModeRules, swiftModeRulesEnabled:
		if engineExit != 0 || r.provenance == nil {
			return 1, fmt.Errorf("the Swift engine %s before listing its rules", ended)
		}
		sort.Slice(r.rules, func(first, second int) bool { return r.rules[first].Name < r.rules[second].Name })
		for _, listed := range r.rules {
			if r.mode == swiftModeRulesEnabled {
				fmt.Fprintf(r.out, "%s\t%s\n", listed.Name, listed.Severity)
			} else {
				fmt.Fprintln(r.out, listed.Name)
			}
		}
		return 0, nil
	}

	if r.summary == nil {
		r.writeUnfinished(fmt.Sprintf("the Swift engine %s before finishing", ended))
		return 1, fmt.Errorf("the Swift engine %s without a summary, so nothing was checked", ended)
	}
	if engineExit != *r.summary.ExitCode {
		return 1, fmt.Errorf("the Swift engine %s, and its summary said it would exit %d", ended, *r.summary.ExitCode)
	}
	return *r.summary.ExitCode, nil
}

// writeUnfinished prints the phase line for a run that stopped early, every phase the engine never
// recorded marked not reached, so the last line a reader sees is never a clean bill of health.
func (r *swiftRun) writeUnfinished(reason string) {
	if r.mode != swiftModeCheck || r.summary != nil {
		return
	}
	for ; r.nextPhase < len(phaseOrder); r.nextPhase++ {
		r.report.record(phaseOrder[r.nextPhase], outcomeNotReached, 0, reason)
	}
	if r.project == nil {
		r.report.graphNotBuilt = true
	}
	writeRunEnd(r.report, r.out)
}

// truncateForError keeps a refused line short enough to read in an error.
func truncateForError(line []byte) string {
	const limit = 200
	if len(line) <= limit {
		return string(line)
	}
	return string(line[:limit]) + "…"
}

// swiftUncheckedFiles says, for the footer, how many files nothing checked and why, since the notes naming
// them are --verbose's. A crashed file is counted apart, as the footer's own marker. Empty when the only
// gaps were crashes.
func swiftUncheckedFiles(withoutRecord, unreadable int) string {
	total := withoutRecord + unreadable
	if total == 0 {
		return ""
	}
	var kinds []string
	if withoutRecord > 0 {
		kinds = append(kinds, fmt.Sprintf("%d with no compiler record", withoutRecord))
	}
	if unreadable > 0 {
		kinds = append(kinds, fmt.Sprintf("%d unreadable", unreadable))
	}
	return fmt.Sprintf("%s nothing checked: %s", counted(total, "file", "files"), strings.Join(kinds, ", "))
}
