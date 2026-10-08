package regexp_test

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"hash"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
	"unicode/utf16"

	"github.com/system-inc/adamic/internal/native"
	regex "github.com/system-inc/adamic/internal/regexp"
)

const cohereData = "testdata/cohere"

type coherePattern struct {
	ID           int               `json:"id"`
	File         string            `json:"file"`
	Line         int               `json:"line"`
	Column       int               `json:"column"`
	Kind         string            `json:"kind"`
	Pattern      string            `json:"pattern"`
	PatternUnits []uint16          `json:"patternUnits"`
	Flags        string            `json:"flags"`
	Fixture      bool              `json:"fixture"`
	Inputs       []json.RawMessage `json:"inputs"`
}
type cohereInventory struct {
	Version          int               `json:"version"`
	TypeScript       string            `json:"typescript"`
	Files            []json.RawMessage `json:"files"`
	Patterns         []coherePattern   `json:"patterns"`
	Dynamic          []json.RawMessage `json:"dynamic"`
	ParseDiagnostics []json.RawMessage `json:"parseDiagnostics"`
	Revisions        map[string]string `json:"revisions"`
}
type cohereResult struct {
	Index       *int                `json:"index"`
	Captures    [][]int             `json:"captures"`
	Values      [][]uint16          `json:"values"`
	Groups      map[string][]int    `json:"groups"`
	GroupValues map[string][]uint16 `json:"groupValues"`
	LastIndex   uint64              `json:"lastIndex"`
}
type cohereCase struct {
	Pattern   int               `json:"pattern"`
	Input     []uint16          `json:"input"`
	Sources   []json.RawMessage `json:"sources"`
	Series    int               `json:"series"`
	Step      int               `json:"step"`
	LastIndex uint64            `json:"lastIndex"`
	Expected  cohereResult      `json:"expected"`
}
type cohereObservation struct {
	Version         int    `json:"version"`
	Node            string `json:"node"`
	Seed            uint32 `json:"seed"`
	InventorySHA256 string `json:"inventorySHA256"`
	Patterns        []struct {
		ID               int    `json:"id"`
		NodeError        string `json:"nodeError"`
		NodePatternError string `json:"nodePatternError"`
		SourceError      string `json:"sourceError"`
		GeneratorError   string `json:"generatorError"`
		InputCount       int    `json:"inputCount"`
		CaseCount        int    `json:"caseCount"`
	} `json:"patterns"`
	Cases []cohereCase `json:"cases"`
}
type cohereIssue struct {
	Pattern   int           `json:"pattern"`
	Engine    string        `json:"engine"`
	Kind      string        `json:"kind"`
	Error     string        `json:"error,omitempty"`
	FirstCase int           `json:"firstCase"`
	Count     int           `json:"count"`
	Digest    string        `json:"digest,omitempty"`
	Want      *cohereResult `json:"want,omitempty"`
	Got       *cohereResult `json:"got,omitempty"`
}

func cohereRead(t *testing.T) (cohereInventory, cohereObservation) {
	t.Helper()
	var inventory cohereInventory
	data, err := os.ReadFile(filepath.Join(cohereData, "inventory.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &inventory); err != nil {
		t.Fatal(err)
	}
	var observations cohereObservation
	file, err := os.Open(filepath.Join(cohereData, "observations.json.gz"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	compressed, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer compressed.Close()
	if err := json.NewDecoder(compressed).Decode(&observations); err != nil {
		t.Fatal(err)
	}
	if inventory.Version != 1 || observations.Version != 1 || len(inventory.Patterns) == 0 || len(inventory.Patterns) != len(observations.Patterns) {
		t.Fatal("invalid or empty cohere corpus")
	}
	if fmt.Sprintf("%x", sha256.Sum256(data)) != observations.InventorySHA256 {
		t.Fatal("inventory changed without regenerating Node observations")
	}
	counts := make([]int, len(inventory.Patterns))
	for i, p := range inventory.Patterns {
		if p.ID != i || observations.Patterns[i].ID != i || p.File == "" || p.Line < 1 || p.PatternUnits == nil {
			t.Fatalf("invalid inventory row %d", i)
		}
	}
	for i, c := range observations.Cases {
		if c.Pattern < 0 || c.Pattern >= len(counts) {
			t.Fatalf("invalid case %d", i)
		}
		counts[c.Pattern]++
		if c.Expected.Captures != nil && (c.Expected.Index == nil || len(c.Expected.Captures) != len(c.Expected.Values)) {
			t.Fatalf("incomplete Node observation %d", i)
		}
		if c.Step > 0 && (i == 0 || observations.Cases[i-1].Series != c.Series || observations.Cases[i-1].Step+1 != c.Step || observations.Cases[i-1].Pattern != c.Pattern || !reflect.DeepEqual(observations.Cases[i-1].Input, c.Input)) {
			t.Fatalf("broken exec sequence %d", i)
		}
	}
	for i, count := range counts {
		if count != observations.Patterns[i].CaseCount || count == 0 && observations.Patterns[i].NodeError == "" {
			t.Fatalf("uncovered pattern %d", i)
		}
	}
	return inventory, observations
}

// Keep named dictionaries separate: the named-only mutant leaves every other
// field equal, and must be killed here rather than by capture-array comparison.
func cohereDifference(got, want cohereResult) string {
	if !reflect.DeepEqual(got.Groups, want.Groups) || !reflect.DeepEqual(got.GroupValues, want.GroupValues) {
		return "named groups"
	}
	if !reflect.DeepEqual(got.Index, want.Index) {
		return "match index"
	}
	if !reflect.DeepEqual(got.Captures, want.Captures) {
		return "capture indices"
	}
	if !reflect.DeepEqual(got.Values, want.Values) {
		return "capture values"
	}
	if got.LastIndex != want.LastIndex {
		return "lastIndex"
	}
	return ""
}
func cohereGoResult(r *regex.RegExp, input []uint16) (cohereResult, error) {
	match, err := r.Exec(input)
	got := cohereResult{LastIndex: r.LastIndex}
	if err != nil || match == nil {
		return got, err
	}
	got.Captures = make([][]int, len(match.Captures))
	got.Values = make([][]uint16, len(match.Captures))
	for i, c := range match.Captures {
		if c.Start >= 0 {
			got.Captures[i] = []int{c.Start, c.End}
			got.Values[i] = append([]uint16{}, input[c.Start:c.End]...)
		}
	}
	index := match.Captures[0].Start
	got.Index = &index
	if len(match.Groups) > 0 {
		got.Groups = map[string][]int{}
		got.GroupValues = map[string][]uint16{}
		for name, c := range match.Groups {
			got.Groups[name] = nil
			got.GroupValues[name] = nil
			if c.Start >= 0 {
				got.Groups[name] = []int{c.Start, c.End}
				got.GroupValues[name] = append([]uint16{}, input[c.Start:c.End]...)
			}
		}
	}
	return got, nil
}

type cohereIssues struct {
	rows    map[string]*cohereIssue
	digests map[string]hash.Hash
}

func (issues *cohereIssues) refusal(pattern int, engine, kind, message string) {
	key := fmt.Sprintf("%06d/%s/%s", pattern, engine, kind)
	issues.rows[key] = &cohereIssue{Pattern: pattern, Engine: engine, Kind: kind, Error: message, FirstCase: -1, Count: 1}
}
func (issues *cohereIssues) disagreement(pattern, index int, engine, kind string, got, want cohereResult) {
	key := fmt.Sprintf("%06d/%s/disagreement", pattern, engine)
	row := issues.rows[key]
	if row == nil {
		row = &cohereIssue{Pattern: pattern, Engine: engine, Kind: kind, FirstCase: index, Got: &got, Want: &want}
		issues.rows[key] = row
		issues.digests[key] = sha256.New()
	}
	row.Count++
	// Include Node too: changing a recorded expectation cannot hide inside a known failure.
	encoded, _ := json.Marshal(struct {
		Case      int
		Kind      string
		Got, Want cohereResult
	}{index, kind, got, want})
	issues.digests[key].Write(encoded)
}
func (issues *cohereIssues) sorted() []cohereIssue {
	keys := make([]string, 0, len(issues.rows))
	for key := range issues.rows {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	rows := make([]cohereIssue, 0, len(keys))
	for _, key := range keys {
		row := *issues.rows[key]
		if digest := issues.digests[key]; digest != nil {
			row.Digest = fmt.Sprintf("%x", digest.Sum(nil))
		}
		rows = append(rows, row)
	}
	return rows
}

// Not parallel: the explicit recording mode writes a single report, and the
// native harness compiles the entire corpus once rather than spawning per case.
func TestCoherePatterns(t *testing.T) {
	inventory, observations := cohereRead(t)
	issues := cohereIssues{rows: map[string]*cohereIssue{}, digests: map[string]hash.Hash{}}
	programs := make([]*regex.Program, len(inventory.Patterns))
	declarations := make([]string, len(programs))
	for i, p := range inventory.Patterns {
		// Also check the exact source flags before adding the observation-only d flag.
		program, err := regex.CompileUTF16(p.PatternUnits, p.Flags)
		if err != nil {
			issues.refusal(i, "compiler", "refusal", err.Error())
			continue
		}
		if observations.Patterns[i].NodeError != "" {
			kind := "wrong acceptance"
			if observations.Patterns[i].NodePatternError == "" {
				kind = "source-only rejection"
			}
			issues.refusal(i, "compiler", kind, observations.Patterns[i].NodeError)
			continue
		}
		if !strings.Contains(p.Flags, "d") {
			program, err = regex.CompileUTF16(p.PatternUnits, p.Flags+"d")
		}
		if err != nil {
			issues.refusal(i, "compiler", "indices refusal", err.Error())
			continue
		}
		programs[i] = program
		declarations[i], err = program.NativeDeclarations(fmt.Sprintf("cohere_regex_%d", i))
		if err != nil {
			issues.refusal(i, "native", "serialization refusal", err.Error())
		}
	}
	var runner *regex.RegExp
	goCompared := 0
	for i, c := range observations.Cases {
		if programs[c.Pattern] == nil {
			continue
		}
		if c.Step == 0 {
			runner = programs[c.Pattern].New()
		}
		if c.Step == 0 {
			runner.LastIndex = c.LastIndex
		}
		runner.StepLimit = 10_000_000
		got, err := cohereGoResult(runner, c.Input)
		if err != nil {
			t.Fatalf("Go pattern %d case %d: %v; computation did not finish", c.Pattern, i, err)
		}
		goCompared++
		if difference := cohereDifference(got, c.Expected); difference != "" {
			issues.disagreement(c.Pattern, i, "Go", difference, got, c.Expected)
		}
		if len(got.Captures) > 0 && got.Captures[0][0] == got.Captures[0][1] && strings.ContainsAny(inventory.Patterns[c.Pattern].Flags, "gy") {
			runner.LastIndex = cohereAdvance(c.Input, runner.LastIndex, inventory.Patterns[c.Pattern].Flags)
		}
	}
	nativeCompared := cohereNative(t, inventory, observations, declarations, &issues)
	rows := issues.sorted()
	if os.Getenv("ADAMIC_COHERE_RECORD") == "1" {
		encoded, err := json.MarshalIndent(rows, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(cohereData, "issues.json"), append(encoded, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
		cohereReport(t, inventory, observations, rows, goCompared, nativeCompared)
	} else {
		encoded, err := os.ReadFile(filepath.Join(cohereData, "issues.json"))
		if err != nil {
			t.Fatal(err)
		}
		var want []cohereIssue
		if err = json.Unmarshal(encoded, &want); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(rows, want) {
			for i := 0; i < max(len(rows), len(want)); i++ {
				if i >= len(rows) || i >= len(want) || !reflect.DeepEqual(rows[i], want[i]) {
					var gotRow, wantRow any
					if i < len(rows) {
						gotRow = rows[i]
					}
					if i < len(want) {
						wantRow = want[i]
					}
					a, _ := json.Marshal(gotRow)
					b, _ := json.Marshal(wantRow)
					t.Fatalf("cohere failure inventory changed at row %d\nobserved %s\nrecorded %s", i, a, b)
				}
			}
		}
	}
	refused, nodeRejected, named := 0, 0, 0
	for i, p := range programs {
		if p == nil {
			refused++
		}
		if observations.Patterns[i].NodeError != "" {
			nodeRejected++
		}
	}
	for _, c := range observations.Cases {
		if c.Expected.Groups != nil {
			named++
		}
	}
	t.Logf("%d files; %d patterns; %d unresolved constructors; %d Node rejected; %d not executable by Adamic; %d Go comparisons; %d native comparisons; %d named-group executions; %d pinned issue rows", len(inventory.Files), len(programs), len(inventory.Dynamic), nodeRejected, refused, goCompared, nativeCompared, named, len(rows))
}
func cohereAdvance(input []uint16, index uint64, flags string) uint64 {
	if strings.ContainsAny(flags, "uv") && index+1 < uint64(len(input)) && input[index] >= 0xd800 && input[index] <= 0xdbff && input[index+1] >= 0xdc00 && input[index+1] <= 0xdfff {
		return index + 2
	}
	return index + 1
}

func cohereNative(t *testing.T, inventory cohereInventory, observations cohereObservation, declarations []string, issues *cohereIssues) int {
	t.Helper()
	var source, rows, units strings.Builder
	source.WriteString("#include \"adamic.h\"\n#include <stdio.h>\n#include <stdlib.h>\n")
	for _, declaration := range declarations {
		source.WriteString(declaration)
	}
	offset, count := 0, 0
	for i, c := range observations.Cases {
		if declarations[c.Pattern] == "" {
			continue
		}
		fmt.Fprintf(&rows, "{&cohere_regex_%d,%d,%d,%d,%d,UINT64_C(%d),%d},\n", c.Pattern, i, offset, len(c.Input), c.Series, c.LastIndex, c.Step)
		for _, unit := range c.Input {
			fmt.Fprintf(&units, "%d,", unit)
		}
		offset += len(c.Input)
		count++
	}
	if count == 0 {
		t.Fatal("no native comparisons")
	}
	fmt.Fprintf(&source, "static const uint16_t input_units[]={%s0};\n", units.String())
	source.WriteString("typedef struct {const adamic_regex_program *program;size_t index,offset,length,series;uint64_t last;size_t step;} probe;\nstatic const probe probes[]={\n")
	source.WriteString(rows.String())
	source.WriteString("};\n")
	harness, err := os.ReadFile(filepath.Join(cohereData, "native.c"))
	if err != nil {
		t.Fatal(err)
	}
	source.Write(harness)
	binary := filepath.Join(t.TempDir(), "cohere-regex")
	t.Logf("native: %d executions, %d C bytes", count, source.Len())
	if err := native.Build(source.String(), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	compared := 0
	for _, mode := range []string{"bounded", "unlimited", "vm"} {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		command := exec.CommandContext(ctx, binary, mode)
		command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1")
		output, err := command.Output()
		cancel()
		if err != nil {
			if exit, ok := err.(*exec.ExitError); ok {
				t.Fatalf("native %s: %v\n%s", mode, err, exit.Stderr)
			}
			t.Fatal(err)
		}
		scanner := bufio.NewScanner(bytes.NewReader(output))
		scanner.Buffer(make([]byte, 4096), 1<<20)
		seen := make([]bool, len(observations.Cases))
		observed := 0
		for scanner.Scan() {
			var row struct {
				Case   int          `json:"case"`
				Result cohereResult `json:"result"`
			}
			if err := json.Unmarshal(scanner.Bytes(), &row); err != nil {
				t.Fatalf("native JSON: %v\n%s", err, scanner.Bytes())
			}
			if row.Case < 0 || row.Case >= len(seen) || seen[row.Case] || declarations[observations.Cases[row.Case].Pattern] == "" {
				t.Fatalf("invalid native row %d", row.Case)
			}
			seen[row.Case] = true
			observed++
			c := observations.Cases[row.Case]
			if difference := cohereDifference(row.Result, c.Expected); difference != "" {
				issues.disagreement(c.Pattern, row.Case, "native "+mode, difference, row.Result, c.Expected)
			}
		}
		if err := scanner.Err(); err != nil {
			t.Fatal(err)
		}
		if observed != count {
			t.Fatalf("native %s observed %d of %d", mode, observed, count)
		}
		compared += observed
		t.Logf("native %s: %d comparisons, sanitizers and leak check passed", mode, observed)
	}
	return compared
}

func TestCohereNamedGroupGuard(t *testing.T) {
	_, observations := cohereRead(t)
	for i, c := range observations.Cases {
		if len(c.Expected.Groups) == 0 {
			continue
		}
		encoded, _ := json.Marshal(c.Expected)
		var mutant cohereResult
		json.Unmarshal(encoded, &mutant)
		for name, capture := range mutant.Groups {
			if capture == nil {
				mutant.Groups[name] = []int{0, 0}
			} else {
				mutant.Groups[name] = []int{capture[0], capture[1] + 1}
			}
			break
		}
		if cohereDifference(mutant, c.Expected) != "named groups" {
			t.Fatalf("named-only mutant survived at case %d", i)
		}
		t.Logf("named-only sentinel from corpus pattern %d case %d catches omission of the named dictionary comparison", c.Pattern, i)
		return
	}
	t.Fatal("no recorded named-group match to hold the comparison")
}

func cohereFeature(p coherePattern, issue cohereIssue, nodeError string) string {
	if issue.Kind == "source-only rejection" {
		return "Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect."
	}
	if nodeError != "" {
		return "Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established."
	}
	if strings.Contains(p.Pattern, "9223372036854775808,9223372036854775807") {
		return "Node-compatible quantifier bound conversion near 2^63. Observation: Node accepts the bounds; Adamic rejects their exact integer order. Inference: V8 rounds these bounds to the same double."
	}
	if strings.Contains(p.Pattern, `[\0\1\01\123\08\8]`) {
		return "Annex B legacy octal escapes inside character classes (\\123 must match U+0053), alongside decimal escapes and backreferences. The witness contains U+0053; inspect class escape decoding first."
	}
	if strings.Contains(issue.Error, "property") || strings.Contains(p.Pattern, "\\p{") || strings.Contains(p.Pattern, "\\P{") {
		return "Unicode property data or Unicode-set semantics; inspect the recorded refusal/result."
	}
	if strings.Contains(p.Pattern, "(?<") {
		return "Named captures, lookbehind, or duplicate-name semantics; the observed field identifies the difference."
	}
	if issue.Kind == "lastIndex" {
		return "Stateful g/y lastIndex, empty matches, and UTF-16 AdvanceStringIndex."
	}
	if issue.Kind == "capture indices" || issue.Kind == "capture values" {
		return "Capture participation and offsets under backtracking or Unicode matching."
	}
	return "ECMAScript pattern parsing or matching semantics; the exact refusal or first counterexample is recorded below."
}
func cohereReport(t *testing.T, inventory cohereInventory, observations cohereObservation, issues []cohereIssue, goCount, nativeCount int) {
	t.Helper()
	var report strings.Builder
	report.WriteString("# Cohere regex patterns held to Node\n\n")
	fmt.Fprintf(&report, "Recorded from Adamic `%s`, cohere `%s`, and cohere's TypeScript `%s`. TypeScript API %s parsed %d files. There are %d static occurrences (%d literals, %d constant constructors), %d unresolved constructors, and %d parser diagnostics in intentionally malformed and other source fixtures.\n\n", inventory.Revisions["adamic"], inventory.Revisions["cohere"], inventory.Revisions["typescript"], inventory.TypeScript, len(inventory.Files), len(inventory.Patterns), cohereKindCount(inventory, "literal"), cohereKindCount(inventory, "constructor"), len(inventory.Dynamic), len(inventory.ParseDiagnostics))
	rejected, inputs, calls, gaps := 0, 0, 0, 0
	for _, p := range observations.Patterns {
		if p.NodeError != "" {
			rejected++
		}
		inputs += p.InputCount
		if p.GeneratorError != "" {
			gaps++
		}
	}
	for _, p := range inventory.Patterns {
		calls += len(p.Inputs)
	}
	fmt.Fprintf(&report, "%s is the external oracle. %d patterns are rejected by Node; %d input strings produce %d recorded exec calls. There are %d source/trace input records and %d generator parser gaps. Adamic completed %d Go comparisons and %d native comparisons (bounded, unlimited automatic dispatch, and forced VM), with ASan, UBSan, and LeakSanitizer on native runs.\n\n", observations.Node, rejected, inputs, len(observations.Cases), calls, gaps, goCount, nativeCount)
	report.WriteString("Extraction, fixture tracing, generation, regeneration commands, limitations, and mutants are described in [the corpus README](../internal/regexp/testdata/cohere/README.md). CI reads the Node observations without invoking Node. Pattern strings, flags, exact UTF-16 units, all occurrence locations, all input provenance, and every observation are in the inventory and compressed testdata. The test compares all cases and pins every observed refusal or disagreement exactly in `issues.json`; it never exempts an entire pattern. Updating this report requires the explicit `ADAMIC_COHERE_RECORD=1` mode.\n\n")
	report.WriteString("## Node-valid gaps for the next unit\n\n")
	for _, issue := range issues {
		if observations.Patterns[issue.Pattern].NodeError != "" || strings.HasPrefix(issue.Engine, "native ") {
			continue
		}
		p := inventory.Patterns[issue.Pattern]
		fmt.Fprintf(&report, "- Pattern %d at `%s:%d`: %s. %s\n", p.ID, p.File, p.Line, issue.Kind, cohereFeature(p, issue, ""))
	}
	fixtureCount := 0
	for _, p := range inventory.Patterns {
		if p.Fixture {
			fixtureCount++
		}
	}
	sourceOnly := 0
	for _, p := range observations.Patterns {
		if p.SourceError != "" && p.NodePatternError == "" {
			sourceOnly++
		}
	}
	fmt.Fprintf(&report, "\nScope: %d implementation/script occurrences and %d fixture/test occurrences. Of Node's %d rejections, %d are invalid literal source tokens whose recovered body/flags would be valid constructors. They do not establish regex compiler defects.\n\n", len(inventory.Patterns)-fixtureCount, fixtureCount, rejected, sourceOnly)
	report.WriteString("## Unresolved constructors\n\nThese are observations of expressions the constant evaluator could not reduce, not proof that all are inherently dynamic.\n\n")
	for _, entry := range inventory.Dynamic {
		var d struct {
			File               string
			Line               int
			Expression, Reason string
		}
		if err := json.Unmarshal(entry, &d); err != nil {
			t.Fatal(err)
		}
		fmt.Fprintf(&report, "- `%s:%d`: ``%s`` (%s).\n", d.File, d.Line, d.Expression, d.Reason)
	}
	report.WriteString("\n## Refusals and disagreements\n\nEach entry names an observed failure. The feature descriptions are triage inferences, not fixes. Agreed rejection of invalid fixtures is listed because the unit asks for every refusal. Wrong acceptance of a Node-invalid pattern is a compiler gap. Source-only rejections mean an invalid literal token was recovered by TypeScript, but its separated body/flags are valid; those are not regex compiler defects.\n\n")
	if len(issues) == 0 {
		report.WriteString("None observed.\n")
	}
	for _, issue := range issues {
		p := inventory.Patterns[issue.Pattern]
		fmt.Fprintf(&report, "### Pattern %d, %s, %s\n\nLocation: `%s:%d:%d`.\n\n```text\npattern: %q\nflags: %q\n```\n\n", p.ID, issue.Engine, issue.Kind, p.File, p.Line, p.Column, p.Pattern, p.Flags)
		if issue.Error != "" {
			fmt.Fprintf(&report, "Refusal/diagnostic: %s\n\n", issue.Error)
		}
		if issue.FirstCase >= 0 {
			c := observations.Cases[issue.FirstCase]
			want, _ := json.Marshal(issue.Want)
			got, _ := json.Marshal(issue.Got)
			fmt.Fprintf(&report, "First disagreement: case %d, UTF-16 input `%v` (%q), initial lastIndex %d, sequence %d step %d. %d disagreements for this pattern and engine.\n\n```json\nNode: %s\nAdamic: %s\n```\n\n", issue.FirstCase, c.Input, string(utf16.Decode(c.Input)), c.LastIndex, c.Series, c.Step, issue.Count, want, got)
		}
		if observations.Patterns[p.ID].NodeError != "" {
			fmt.Fprintf(&report, "Node refusal: %s\n\n", observations.Patterns[p.ID].NodeError)
		}
		fmt.Fprintf(&report, "Feature needed (inference): %s\n\n", cohereFeature(p, issue, observations.Patterns[p.ID].NodeError))
	}
	if err := os.WriteFile("../../docs/regexp-cohere-patterns.md", []byte(strings.TrimRight(report.String(), "\n")+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
}
func cohereKindCount(inventory cohereInventory, kind string) int {
	count := 0
	for _, p := range inventory.Patterns {
		if p.Kind == kind {
			count++
		}
	}
	return count
}

// This flag-free CI test also guards the expectation comparator itself with
// real recorded captures, rather than a made-up result mirroring its code.
func TestCohereCaptureGuard(t *testing.T) {
	_, observations := cohereRead(t)
	for i, c := range observations.Cases {
		if len(c.Expected.Captures) < 2 || c.Expected.Captures[1] == nil {
			continue
		}
		encoded, _ := json.Marshal(c.Expected)
		var mutant cohereResult
		json.Unmarshal(encoded, &mutant)
		mutant.Captures[1][1]++
		if cohereDifference(mutant, c.Expected) != "capture indices" {
			t.Fatalf("capture-only mutant survived at case %d", i)
		}
		t.Logf("recorded capture mutant caught at pattern %d case %d", c.Pattern, i)
		return
	}
	t.Fatal("no recorded participating capture")
}

func TestCohereObservationGuards(t *testing.T) {
	_, observations := cohereRead(t)
	checks := []struct {
		name, kind string
		eligible   func(cohereResult) bool
		mutate     func(*cohereResult)
	}{
		{"index", "match index", func(r cohereResult) bool { return r.Index != nil }, func(r *cohereResult) { *r.Index++ }},
		{"values", "capture values", func(r cohereResult) bool { return len(r.Values) > 0 && len(r.Values[0]) > 0 }, func(r *cohereResult) { r.Values[0][0]++ }},
		{"named values", "named groups", func(r cohereResult) bool { return len(r.GroupValues) > 0 }, func(r *cohereResult) {
			for name := range r.GroupValues {
				r.GroupValues[name] = append(r.GroupValues[name], 0)
				break
			}
		}},
		{"lastIndex", "lastIndex", func(r cohereResult) bool { return true }, func(r *cohereResult) { r.LastIndex++ }},
	}
	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			for i, c := range observations.Cases {
				if !check.eligible(c.Expected) {
					continue
				}
				if cohereDifference(c.Expected, c.Expected) != "" {
					t.Fatal("control must agree")
				}
				encoded, err := json.Marshal(c.Expected)
				if err != nil {
					t.Fatal(err)
				}
				var mutant cohereResult
				if err = json.Unmarshal(encoded, &mutant); err != nil {
					t.Fatal(err)
				}
				check.mutate(&mutant)
				if cohereDifference(mutant, c.Expected) != check.kind {
					t.Fatalf("%s-only mutant survived at case %d", check.name, i)
				}
				t.Logf("%s-only mutant caught using corpus pattern %d case %d", check.name, c.Pattern, i)
				return
			}
			t.Fatalf("no real corpus result for %s", check.name)
		})
	}
}
