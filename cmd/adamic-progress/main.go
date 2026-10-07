// Command adamic-progress reads landed evidence and git, never a gate.
package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"math"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/system-inc/adamic/internal/meterdata"
	"github.com/system-inc/adamic/internal/stage1progress"
)

const help = `Usage: go run ./cmd/adamic-progress [--json] [--history]
Reads origin/main (the last fetched main), never fetches, builds a compiler, or runs a gate.
Each Git read has its own 30-second bound; failed reads are cached during the report.
A missing track, history point or milestone observation leaves the other sections available.
Sources and accounting:
  Clock: current UTC and fixed MDT (UTC-6). Day 1 starts October 4 2026 23:49 MDT.
  Stage 3 deadline: October 9 23:49 MDT. Stage 1 and Apple: October 10 23:49 MDT.
  Stage 3: latest timestamp_utc in stage3/meter/runs/*/report.json on main,
    Paired runs use main/report.json for main; area/report.json is labeled separately.
    checker_own_file counts source-local diagnostic success, not whole-program success;
    area totals use their recorded source count and never contribute to main overall.
    runner schema from codex/stage3-fixtures-runner. Per-file checker/lowering
    booleans are counted separately, out of the plan's 78 (generated sources included).
    adamic-meter's files_reaching_lowering is checker success, not lowering success.
    Its shared report schema is used for legacy recorded meter files; it cannot
    establish lowering success. No main run means "no meter run on main yet".
  Patch size: stage3/patch-set.md Total row: files, added and removed lines;
    temporary adaptation rows included. Size has no percentage target and is
    excluded from the overall bar. Missing patch record is unknown.
  Stage 3 milestones: stage3/progress.json milestones object with scanner_native,
    parser_native, no_emit_match, real_programs_match booleans. Missing is unknown;
    a driver source alone cannot prove native execution or exact diagnostics.
  Stage 1 lines: shared adamic-stage1-progress inventory, same main snapshot,
    reviewed coverage mappings and pinned cohere non-test physical Go lines.
  Lint: distinct production rule variables in []rule.Rule registrations in
    stage1/cohere/{lint,typeaware}/testdata/oracle*.go on main, matched by family
    and variable to inventory.json rules. Inventory comes from main first, then
    origin/codex/lint-inventory. These registrations record port oracle scope,
    not the presence of upstream Go rules or the number of finding strings.
  Formatters: stage1/progress.json formatters object: json, css, graphql,
    markdown, yaml, javascript, typescript booleans recording complete byte parity.
    Slice parser/printer parity does not establish a whole formatter. Missing unknown.
  Speed: main's stage1/cohere/lint/performance/*-measurements.json best.native
    and best.Go seconds, latest by git commit time; Go/native is relative speed.
    Syntax bar is min(Go/native,1); equality is not faster. Whole-cohere
    speed is a 0/1 milestone for native_seconds < go_seconds, with the timings shown. This is syntax-lint corpus
    speed only, so it does not establish whole cohere speed. Unknown whole speed
    from stage1/progress.json native_seconds/go_seconds remains in the overall.
  Apple: example existence on main, examples/apple/{window,fetch,counter,
    bindings,app}.a, one milestone each. Existence records delivery, not execution.
  Overall: equal-weight mean of goal fractions, with unknowns contributing zero
    to a lower bound; number of known measures shown. No patch-size weighting.
  History: 25 hourly observations ending now. For each moment, select the main
    first-parent commit whose committer timestamp was current, and git show the
    same sources above. Git has no historical push journal; commit timestamps
    approximate landing times. Meter runs additionally respect their recorded
    timestamp cutoff. Unknown evidence stays unknown, never fabricated zero.
    Show now, 1h, 6h, 24h, and a min/max-scaled Unicode sparkline per measure.
    --history prints all hourly values; --json includes observations and rates.
  Pace and ETA: actual is change over the trailing six hours, or the oldest
    known hourly point within that window. Needed is remaining / hours until
    deadline. Green means actual >= needed; red means below; unknown is explicit.
    Track ETA extrapolates its overall lower-bound percentage at that rate;
    it is an inference, not certification of unrecorded parity. Nonpositive or
    insufficient rates say "not enough history yet". Micro-deadlines use their
    own remaining quantity and deadline against the same measured rate.
  Micro-deadlines: documentation/progress/milestones.json on main, falling back
    to the command's embedded plan before integration. Each entry documents its
    exact record keys, threshold, source and ISO MDT deadline. Overdue red first.
    Additional stage3/progress.json records: cycle_census_ruled, merge_tree_live,
    nested_functions.{passed,total,scanner_frame}, scanner_tokens_all_compiler_match,
    parser_trees_match, records_table_main, records_lowering_main. Adaptation 20
    requires its adaptation20*/adapt.cjs file on main.
    stage1/progress.json: readonly_walk.{driver,retains_per_node},
    lint_release.{release,flags}, refused_rules with explicit nonempty reasons.
    parse_batch8.{driver,files,unit,native_scope,go_scope,native_instructions,
    go_parse_instructions}: batch8, 77 files, instructions, both parse_alone.
    Native/Go parse-only instruction ratio <=1.5 by Oct 8 noon and <=1 by Oct 9
    noon (runtime, #93z4yv7). Go whole-run counts never certify these goals.
    Apple progress.json: generated_bindings.{appkit,foundation,rename_table},
    real_app.{list,state,typed_fetch,async,crossing_cost_table_with_swift}.
  Host: stage3/fixtures/host/status.json recorded results for all 25 .a fixtures
    from codex/stage3-fixtures-host 1037217. A green fixture requires recorded
    Node stdout/stderr/exit matching both native and JavaScript observations.
    Checker-blocked fixtures count zero; incomplete backend evidence is "not
    measurable yet". Never execute a fixture or gate. Final scope is main;
    interim checkpoints explicitly read origin/area/library if fetched.
    documentation/progress/host-loader.json names hook_commit and four branch
    refs; git ancestry proves their node:* loader merge checkpoint.
  Slowest: largest projected lateness when all tracks have ETAs; otherwise lowest
    overall lower bound, with lateness explicitly unknown and remaining percent.
  Velocity: git rev-list main total; committer timestamps in last 24h/hour,
    distinct commit hashes reachable from fetched origin refs excluding main/HEAD
    and commits already on main (identity count, retained for comparison).
    Patch backlog: git log --no-merges -p | git patch-id --stable; deduplicate
    patch IDs across origin branches, then exclude every patch ID on main.
    Rebases/cherry-picks with identical patches count once. Merge/empty commits
    have no standalone patch; whitespace is ignored. Squashes or conflict edits
    may change patch identity. Both labelled counts remain visible.
    Hash at most 16 commits per git log/patch-id batch. Cache by commit SHA in
    git rev-parse --git-path adamic-progress/patch-ids-v1.json (atomic checkpoint
    after each batch). Cache version fixes Myers, stable, binary, no-renames
    settings. A second run hashes only uncached commits; refs are reevaluated.
    Backlog has its own 60-second budget, independent of other reads. Failure
    or timeout leaves the count unknown/null and prints why; report still succeeds.
    --json patch_backlog_stats exposes hashes, hits, batches and wall seconds.
    documentation/velocity/landings.csv: timestamp, timestamp_utc or pushed_at_utc column,
    optional landings or commits_landed count (default 1), summed into UTC hour buckets.
    Absence explicitly reported. Velocity has no goal percentage.
    Historical remote backlog cannot be reconstructed from current refs:
    documentation/velocity/backlog.csv (timestamp,count) records commit counts;
    documentation/velocity/patch-backlog.csv (timestamp,count) records patch counts;
    without it past backlog is unknown. The hourly-falling checkpoint requires
    an observation each elapsed hour and a strictly decreasing count.
Malformed evidence marks the affected section unavailable with its error; the rest
of the report still prints. Unknown values are never silently converted to zero.
`

var mdt = time.FixedZone("MDT", -6*3600)
var beginning = time.Date(2026, 10, 4, 23, 49, 0, 0, mdt)

type metric struct {
	Progress       *trend   `json:"progress,omitempty"`
	Name           string   `json:"name"`
	Done           float64  `json:"done"`
	Total          float64  `json:"total"`
	Known          bool     `json:"known"`
	Percent        *float64 `json:"percent"`
	Source         string   `json:"source"`
	Note           string   `json:"note,omitempty"`
	Registered     []string `json:"registered_rules,omitempty"`
	InventoryNames []string `json:"-"`
}

func measured(name string, done, total float64, source string) metric {
	m := metric{Name: name, Done: done, Total: total, Known: true, Source: source}
	if total > 0 {
		p := 100 * math.Min(1, math.Max(0, done/total))
		m.Percent = &p
	}
	return m
}
func unknown(name string, total float64, source, note string) metric {
	return metric{Name: name, Total: total, Source: source, Note: note}
}

type track struct {
	Name          string     `json:"name"`
	Deadline      time.Time  `json:"deadline"`
	SecondsLeft   float64    `json:"seconds_left"`
	Measures      []metric   `json:"measures"`
	Overall       metric     `json:"overall"`
	ETA           *time.Time `json:"eta"`
	HistorySource string     `json:"history_source"`
	ETAScope      string     `json:"eta_scope"`
}
type velocity struct {
	Pending      []string       `json:"-"`
	Measures     []metric       `json:"measures"`
	Total        int            `json:"commits_on_main"`
	Day          int            `json:"landed_24h"`
	Hour         int            `json:"landed_1h"`
	Backlog      int            `json:"distinct_backlog"`
	PatchBacklog *int           `json:"distinct_patch_backlog"`
	PatchError   string         `json:"patch_backlog_error,omitempty"`
	PatchStats   *backlogResult `json:"patch_backlog_stats,omitempty"`
	Error        string         `json:"error,omitempty"`
	Landings     map[string]int `json:"landings_per_hour"`
	Note         string         `json:"note,omitempty"`
}
type dashboard struct {
	SectionErrors []string          `json:"section_errors,omitempty"`
	Milestones    []milestoneResult `json:"milestones"`
	Time          time.Time         `json:"time"`
	Day           int               `json:"creation_day"`
	Main          string            `json:"main_commit"`
	Tracks        []track           `json:"tracks"`
	Velocity      velocity          `json:"velocity"`
	Slowest       string            `json:"slowest"`
}
type repository struct {
	backlogContext context.Context
	backlogBudget  time.Duration
	gitExecutable  string
	at             time.Time
	cache          map[string][]byte
	readErrors     map[string]error
	readBudget     time.Duration
	root, ref      string
	ctx            context.Context
	paths          []string
}

func (r repository) gitCommand(args ...string) *exec.Cmd {
	executable := r.gitExecutable
	if executable == "" {
		executable = "git"
	}
	command := exec.CommandContext(r.ctx, executable, append([]string{"-C", r.root}, args...)...)
	command.WaitDelay = time.Second
	return command
}
func (r repository) git(args ...string) ([]byte, error) {
	key := r.root + "\x00" + strings.Join(args, "\x00")
	if err := r.readErrors[key]; err != nil {
		return nil, err
	}
	if b, ok := r.cache[key]; ok {
		return b, nil
	}
	budget := r.readBudget
	if budget <= 0 {
		budget = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(r.ctx, budget)
	defer cancel()
	r.ctx = ctx
	c := r.gitCommand(args...)
	b, e := c.Output()
	if e != nil {
		err := fmt.Errorf("git %s: %w", strings.Join(args, " "), e)
		if r.readErrors != nil {
			r.readErrors[key] = err
		}
		return nil, err
	}
	if r.cache != nil {
		r.cache[key] = b
	}
	return b, nil
}
func (r repository) blob(name string) ([]byte, bool, error) {
	present := false
	for _, p := range r.paths {
		if p == name {
			present = true
			break
		}
	}
	if !present {
		return nil, false, nil
	}
	b, e := r.git("show", r.ref+":"+name)
	return b, true, e
}
func (r repository) decode(name string, target any) (bool, error) {
	b, ok, e := r.blob(name)
	if e != nil || !ok {
		return ok, e
	}
	if e = json.Unmarshal(b, target); e != nil {
		return true, fmt.Errorf("%s: %w", name, e)
	}
	return true, nil
}
func bar(percent *float64) string {
	if percent == nil {
		return "[????????????????????]"
	}
	n := int(math.Floor(math.Max(0, math.Min(100, *percent)) / 5))
	return "[" + strings.Repeat("█", n) + strings.Repeat("░", 20-n) + "]"
}
func finish(t *track, r repository, now time.Time) error {
	sum := 0.0
	known := 0
	goals := 0
	for _, m := range t.Measures {
		if m.Name == "patch size" || m.Name == "syntax-lint native/Go speed" || strings.HasPrefix(m.Name, "area ") || m.Name == "compiler own-file checker" || m.Name == "meter record error" {
			continue
		}
		goals++
		if m.Known && m.Percent != nil {
			known++
			sum += *m.Percent / 100
		}
	}
	t.Overall = measured("overall (lower bound)", sum, float64(goals), "equal-weight goal fractions")
	t.Overall.Note = fmt.Sprintf("%d/%d measures known", known, goals)
	t.SecondsLeft = t.Deadline.Sub(now).Seconds()
	t.HistorySource = "main first-parent snapshots and recorded meter runs"
	t.ETAScope = "actual rate in overall lower-bound goal fractions"
	return nil
}
func stage3(r repository) (track, error) {
	t := track{Name: "Stage 3", Deadline: beginning.Add(5 * 24 * time.Hour), ETAScope: "stage3 meter lowering coverage proxy (not exact diagnostic completion)"}
	type run struct {
		Timestamp string      `json:"timestamp_utc"`
		Files     []meterFile `json:"files"`
		Totals    struct {
			Own      *int `json:"checker_own_file"`
			Source   int  `json:"source_files"`
			Checker  int  `json:"checker"`
			Lowering int  `json:"lowering"`
		} `json:"totals"`
		meterdata.Report
	}
	var latest, area *run
	areaPath := ""
	var areaTime time.Time
	var warnings []string
	latestPath := ""
	var latestTime time.Time
	for _, p := range r.paths {
		if !strings.HasPrefix(p, "stage3/meter/runs/") || !strings.HasSuffix(p, "/report.json") {
			continue
		}
		isArea := strings.Contains(p, "/area/")
		parent := strings.TrimSuffix(p, "/report.json")
		if !isArea && !strings.HasSuffix(parent, "/main") {
			paired := false
			for _, candidate := range r.paths {
				if candidate == parent+"/main/report.json" {
					paired = true
					break
				}
			}
			if paired {
				continue
			}
		}
		var v run
		_, e := r.decode(p, &v)
		if e != nil {
			warnings = append(warnings, e.Error())
			continue
		}
		if v.Files == nil && (v.FilesExamined <= 0 || v.FilesExamined > 78 || v.FilesReachingLowering < 0 || v.FilesReachingLowering > v.FilesExamined) {
			warnings = append(warnings, p+": missing or invalid legacy meter observations")
			continue
		}
		if v.Files != nil {
			if e := validateMeter(v.Files, v.Totals.Source, v.Totals.Checker, v.Totals.Lowering); e != nil {
				warnings = append(warnings, fmt.Sprintf("%s: %v", p, e))
				continue
			}
		}
		if v.Totals.Own != nil {
			n := 0
			for _, f := range v.Files {
				if f.Source && strings.HasPrefix(f.File, "src/compiler/") && f.Own != nil && *f.Own {
					n++
				}
			}
			if n != *v.Totals.Own {
				warnings = append(warnings, fmt.Sprintf("%s: inconsistent own-file total: recorded=%d per-file=%d", p, *v.Totals.Own, n))
				continue
			}
		}
		stamp, e := time.Parse("20060102T150405Z", v.Timestamp)
		if e != nil {
			warnings = append(warnings, fmt.Sprintf("%s timestamp: %v", p, e))
			continue
		}

		if !r.at.IsZero() && stamp.After(r.at) {
			continue
		}
		if isArea {
			if stamp.After(areaTime) {
				area = &v
				areaPath = p
				areaTime = stamp
			}
			continue
		}
		if stamp.After(latestTime) {
			latest = &v
			latestPath = p
			latestTime = stamp
		}
	}
	if latest == nil {
		for _, name := range []string{"compiler checker", "compiler lowering"} {
			t.Measures = append(t.Measures, unknown(name, 78, "stage3/meter/runs/*/report.json", "no meter run on main yet"))
		}
	} else {
		if latest.Files == nil {
			t.Measures = append(t.Measures, measured("compiler checker", float64(latest.FilesReachingLowering), 78, latestPath), unknown("compiler lowering", 78, latestPath, "legacy meter records checker success only"))
		} else {
			checked, lowered, sources := latest.Totals.Checker, latest.Totals.Lowering, latest.Totals.Source
			for _, v := range []struct {
				name string
				n    int
			}{{"compiler checker", checked}, {"compiler lowering", lowered}} {
				m := measured(v.name, float64(v.n), 78, latestPath)
				m.Note = fmt.Sprintf("recorded %s; %d source files examined", latest.Timestamp, sources)
				t.Measures = append(t.Measures, m)
			}
		}
	}
	if latest != nil && latest.Totals.Own != nil {
		t.Measures = append(t.Measures, measured("compiler own-file checker", float64(*latest.Totals.Own), 78, latestPath))
	}
	if area != nil {
		for _, item := range []struct {
			name string
			n    int
		}{{"area compiler checker", area.Totals.Checker}, {"area compiler lowering", area.Totals.Lowering}} {
			t.Measures = append(t.Measures, measured(item.name, float64(item.n), float64(area.Totals.Source), areaPath))
		}
		if area.Totals.Own != nil {
			t.Measures = append(t.Measures, measured("area compiler own-file checker", float64(*area.Totals.Own), float64(area.Totals.Source), areaPath))
		}
	}
	for _, warning := range warnings {
		t.Measures = append(t.Measures, unknown("meter record error", 0, "record validation", warning))
	}
	b, ok, e := r.blob("stage3/patch-set.md")
	if e != nil {
		return t, e
	}
	p := unknown("patch size", 0, "stage3/patch-set.md", "no patch record on main yet")
	if ok {
		match := regexp.MustCompile(`(?m)^\|\s*\*\*Total\*\*\s*\|\s*(\d+)\s*\|\s*(\d+)\s*\|\s*(\d+)\s*\|`).FindSubmatch(b)
		if match == nil {
			return t, errors.New("stage3/patch-set.md missing Total row")
		}
		files, _ := strconv.Atoi(string(match[1]))
		added, _ := strconv.Atoi(string(match[2]))
		removed, _ := strconv.Atoi(string(match[3]))
		p = measured("patch size", float64(added+removed), 0, "stage3/patch-set.md")
		p.Note = fmt.Sprintf("%d files, +%d/-%d lines; temporary entries included; no size target", files, added, removed)
	}
	t.Measures = append(t.Measures, p)
	var status struct {
		Milestones map[string]bool `json:"milestones"`
	}
	_, e = r.decode("stage3/progress.json", &status)
	if e != nil {
		return t, e
	}
	for _, key := range []string{"scanner_native", "parser_native", "no_emit_match", "real_programs_match"} {
		v, ok := status.Milestones[key]
		m := unknown(key, 1, "stage3/progress.json#milestones."+key, "no native/diagnostic milestone record on main yet")
		if ok {
			m = measured(key, boolNumber(v), 1, m.Source)
		}
		t.Measures = append(t.Measures, m)
	}
	host, e := hostMeasure(r)
	if e != nil {
		return t, e
	}
	t.Measures = append(t.Measures, host)
	return t, nil
}
func boolNumber(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

// Port oracle registrations are the reviewed scope, not upstream registry presence.
func lint(r repository) (metric, error) {
	const path = "stage1/cohere/lint/inventory/inventory.json"
	source := r.ref + ":" + path
	b, ok, e := r.blob(path)
	if e != nil {
		return metric{}, e
	}
	if !ok {
		b, e = r.git("show", "origin/codex/lint-inventory:"+path)
		if e != nil {
			return unknown("lint rules", 0, source, "no inventory on main or origin/codex/lint-inventory"), nil
		}
		source = "origin/codex/lint-inventory:" + path
	}
	var inventory struct {
		Rules []struct{ Name, Family, Variable string }
	}
	if e = json.Unmarshal(b, &inventory); e != nil {
		return metric{}, fmt.Errorf("%s: %w", source, e)
	}
	if len(inventory.Rules) == 0 {
		return metric{}, errors.New("empty lint inventory")
	}
	registered := map[string]bool{}
	for _, name := range r.paths {
		if !(strings.HasPrefix(name, "stage1/cohere/lint/testdata/oracle") || strings.HasPrefix(name, "stage1/cohere/typeaware/testdata/oracle")) || !strings.HasSuffix(name, ".go") {
			continue
		}
		b, _, e = r.blob(name)
		if e != nil {
			return metric{}, e
		}
		f, e := parser.ParseFile(token.NewFileSet(), name, b, 0)
		if e != nil {
			return metric{}, e
		}
		families := map[string]string{}
		for _, imp := range f.Imports {
			path, _ := strconv.Unquote(imp.Path.Value)
			if !strings.Contains(path, "/internal/lint/rules/") {
				continue
			}
			parts := strings.Split(path, "/")
			family := parts[len(parts)-1]
			alias := family
			if imp.Name != nil {
				alias = imp.Name.Name
			}
			families[alias] = family
		}
		ast.Inspect(f, func(n ast.Node) bool {
			literal, ok := n.(*ast.CompositeLit)
			if !ok {
				return true
			}
			array, ok := literal.Type.(*ast.ArrayType)
			if !ok {
				return true
			}
			selector, ok := array.Elt.(*ast.SelectorExpr)
			if !ok || selector.Sel.Name != "Rule" {
				return true
			}
			for _, element := range literal.Elts {
				s, ok := element.(*ast.SelectorExpr)
				if !ok {
					continue
				}
				id, ok := s.X.(*ast.Ident)
				if ok && families[id.Name] != "" {
					registered[families[id.Name]+"/"+s.Sel.Name] = true
				}
			}
			return true
		})
	}
	names := map[string]bool{}
	for _, rule := range inventory.Rules {
		if names[rule.Name] {
			return metric{}, fmt.Errorf("duplicate inventory rule %s", rule.Name)
		}
		names[rule.Name] = true
	}
	count := 0
	for _, rule := range inventory.Rules {
		if registered[rule.Family+"/"+rule.Variable] {
			count++
		}
	}
	m := measured("lint rules", float64(count), float64(len(inventory.Rules)), source+" + main port oracle registrations")
	m.Note = "registered port scope, including partial contracts; not full configuration parity"
	for _, rule := range inventory.Rules {
		m.InventoryNames = append(m.InventoryNames, rule.Name)
		if registered[rule.Family+"/"+rule.Variable] {
			m.Registered = append(m.Registered, rule.Name)
		}
	}
	return m, nil
}
func stage1(r repository, lines func(string, string) (*stage1progress.Report, error)) (track, error) {
	t := track{Name: "Stage 1", Deadline: beginning.Add(6 * 24 * time.Hour)}
	report, e := lines(r.root, r.ref)
	if e != nil {
		return t, e
	}
	if report == nil {
		t.Measures = append(t.Measures, unknown("ported Go lines", 0, "adamic-stage1-progress shared inventory at "+r.ref, "historical inventory unavailable: slice GAPS.md not recorded yet"))
	} else {
		t.Measures = append(t.Measures, measured("ported Go lines", float64(report.Ported), float64(report.Total), "adamic-stage1-progress shared inventory at "+r.ref))
	}
	rules, e := lint(r)
	if e != nil {
		return t, e
	}
	t.Measures = append(t.Measures, rules)
	var status struct {
		Formatters map[string]bool `json:"formatters"`
		Native     *float64        `json:"native_seconds"`
		Go         *float64        `json:"go_seconds"`
	}
	_, e = r.decode("stage1/progress.json", &status)
	if e != nil {
		return t, e
	}
	for _, key := range []string{"json", "css", "graphql", "markdown", "yaml", "javascript", "typescript"} {
		m := unknown(key+" formatter byte-identical", 1, "stage1/progress.json#formatters."+key, "no complete formatter parity record on main yet")
		if v, ok := status.Formatters[key]; ok {
			m = measured(m.Name, boolNumber(v), 1, m.Source)
		}
		t.Measures = append(t.Measures, m)
	}
	m := unknown("whole cohere native speed", 1, "stage1/progress.json native_seconds/go_seconds", "no whole-cohere speed record on main yet")
	if status.Native != nil && status.Go != nil {
		if *status.Native <= 0 || *status.Go <= 0 {
			return t, errors.New("invalid whole-cohere timings")
		}
		m = measured(m.Name, boolNumber(*status.Native < *status.Go), 1, m.Source)
		m.Note = fmt.Sprintf("native %.3fs vs Go %.3fs, %.3fx Go speed", *status.Native, *status.Go, *status.Go / *status.Native)
		if *status.Native >= *status.Go {
			m.Note = "native has not beaten Go"
		}
	}
	t.Measures = append(t.Measures, m)
	// Keep scoped speed observations outside the overall whole-cohere goals.
	latestPath := ""
	var lastTime time.Time
	for _, p := range r.paths {
		if !strings.HasPrefix(p, "stage1/cohere/lint/performance/") || !strings.HasSuffix(p, "-measurements.json") {
			continue
		}
		stamp, e := r.git("log", "-1", "--format=%cI", r.ref, "--", p)
		if e != nil {
			return t, e
		}
		when, e := time.Parse(time.RFC3339, strings.TrimSpace(string(stamp)))
		if e != nil {
			return t, e
		}
		if when.After(lastTime) || (when.Equal(lastTime) && p > latestPath) {
			lastTime = when
			latestPath = p
		}
	}
	if latestPath != "" {
		var v struct {
			Best map[string]struct {
				Seconds float64 `json:"seconds"`
			} `json:"best"`
		}
		_, e = r.decode(latestPath, &v)
		if e != nil {
			return t, e
		}
		native, goTime := v.Best["native"].Seconds, v.Best["Go"].Seconds
		if native <= 0 || goTime <= 0 {
			return t, fmt.Errorf("%s invalid best timings", latestPath)
		}
		m := measured("syntax-lint native/Go speed", goTime/native, 1, latestPath)
		m.Note = fmt.Sprintf("native %.3fs vs Go %.3fs, %.2fx Go speed; scoped observation excluded from overall", native, goTime, goTime/native)
		m.Total = 0
		t.Measures = append(t.Measures, m)
	}

	return t, nil
}
func apple(r repository) (track, error) {
	t := track{Name: "Apple", Deadline: beginning.Add(6 * 24 * time.Hour)}
	for _, v := range []struct{ name, path string }{{"a window", "window"}, {"network fetch with typed decoding", "fetch"}, {"SwiftUI counter", "counter"}, {"generated bindings", "bindings"}, {"real app with async", "app"}} {
		p := "examples/apple/" + v.path + ".a"
		present := false
		for _, name := range r.paths {
			if name == p {
				present = true
				break
			}
		}
		t.Measures = append(t.Measures, measured(v.name, boolNumber(present), 1, r.ref+":"+p))
	}

	return t, nil
}
func readLandings(b []byte, now time.Time) (map[string]int, error) {
	rows, e := csv.NewReader(strings.NewReader(string(b))).ReadAll()
	if e != nil {
		return nil, e
	}
	result := map[string]int{}
	if len(rows) == 0 {
		return nil, errors.New("empty landings.csv")
	}
	timeCol, countCol := -1, -1
	for i, col := range rows[0] {
		switch col {
		case "timestamp", "timestamp_utc", "pushed_at_utc":
			timeCol = i
		case "landings", "commits_landed":
			countCol = i
		}
	}
	if timeCol < 0 {
		return nil, errors.New("landings.csv has no recognized time column (timestamp, timestamp_utc, pushed_at_utc)")
	}
	for _, row := range rows[1:] {
		stamp, e := time.Parse(time.RFC3339, row[timeCol])
		if e != nil {
			return nil, e
		}
		n := 1
		if countCol >= 0 {
			n, e = strconv.Atoi(row[countCol])
			if e != nil || n < 0 {
				return nil, errors.New("invalid landings count")
			}
		}
		if stamp.After(now) {
			continue
		}
		key := stamp.UTC().Truncate(time.Hour).Format(time.RFC3339)
		result[key] += n
	}
	return result, nil
}
func speed(r repository, now time.Time) (velocity, error) { return measureVelocity(r, now, true) }
func measureVelocity(r repository, now time.Time, patches bool) (velocity, error) {
	var v velocity
	count, e := r.git("rev-list", "--count", r.ref)
	if e != nil {
		return v, e
	}
	v.Total, e = strconv.Atoi(strings.TrimSpace(string(count)))
	if e != nil {
		return v, e
	}
	log, e := r.git("log", "--format=%ct", r.ref)
	if e != nil {
		return v, e
	}
	for _, line := range strings.Fields(string(log)) {
		seconds, e := strconv.ParseInt(line, 10, 64)
		if e != nil {
			return v, e
		}
		stamp := time.Unix(seconds, 0)
		if stamp.After(now) {
			continue
		}
		if !stamp.Before(now.Add(-24 * time.Hour)) {
			v.Day++
		}
		if !stamp.Before(now.Add(-time.Hour)) {
			v.Hour++
		}
	}
	refs, e := r.git("for-each-ref", "--format=%(refname)", "refs/remotes/origin")
	if e != nil {
		return v, e
	}
	var pending []string
	for _, ref := range strings.Fields(string(refs)) {
		if ref != "refs/remotes/origin/HEAD" && ref != "refs/remotes/origin/main" {
			pending = append(pending, ref)
		}
	}
	if len(pending) > 0 {
		args := append([]string{"rev-list", "--count"}, pending...)
		args = append(args, "--not", r.ref)
		b, e := r.git(args...)
		if e != nil {
			return v, e
		}
		v.Backlog, e = strconv.Atoi(strings.TrimSpace(string(b)))
		if e != nil {
			return v, e
		}
	}
	v.Pending = pending
	if patches {
		result, err := r.cachedPatchBacklog(pending)
		v.PatchStats = &result
		if err != nil {
			v.PatchError = err.Error()
		} else {
			v.PatchBacklog = &result.Count
		}
	}
	b, ok, e := r.blob("documentation/velocity/landings.csv")
	if e != nil {
		return v, e
	}
	if !ok {
		v.Note = "no landings.csv on main yet"
	} else {
		v.Landings, e = readLandings(b, now)
	}
	return v, e
}
func slowest(tracks []track, now time.Time) string {
	all := true
	for _, t := range tracks {
		if t.ETA == nil {
			all = false
		}
	}
	worst := tracks[0]
	for _, t := range tracks[1:] {
		if all {
			if t.ETA.Sub(t.Deadline) > worst.ETA.Sub(worst.Deadline) {
				worst = t
			}
		} else if *t.Overall.Percent < *worst.Overall.Percent {
			worst = t
		}
	}
	if !all {
		remaining := 100 - *worst.Overall.Percent
		return fmt.Sprintf("Slowest: %s (lowest recorded lower bound); %.1f%% remains; behind deadline: unknown (not enough history yet)", worst.Name, remaining)
	}
	late := worst.ETA.Sub(worst.Deadline)
	if late > 0 {
		return fmt.Sprintf("Slowest: %s; projected %.1fh behind deadline", worst.Name, late.Hours())
	}
	return fmt.Sprintf("Slowest: %s; 0h behind deadline (projected %.1fh early)", worst.Name, -late.Hours())
}
func collect(r repository, now time.Time, lines func(string, string) (*stage1progress.Report, error)) (dashboard, error) {
	if r.cache == nil {
		r.cache = map[string][]byte{}
	}
	if r.readErrors == nil {
		r.readErrors = map[string]error{}
	}
	r.at = now
	d := dashboard{Time: now.UTC(), Day: int(math.Floor(now.Sub(beginning).Hours()/24)) + 1}
	if d.Day < 1 {
		d.Day = 0
	}
	commit, e := r.git("rev-parse", r.ref+"^{commit}")
	if e != nil {
		return d, e
	}
	d.Main = strings.TrimSpace(string(commit))
	r.ref = d.Main
	paths, e := r.git("ls-tree", "-r", "--name-only", r.ref)
	if e != nil {
		return d, e
	}
	r.paths = strings.Fields(string(paths))
	s3, e := stage3(r)
	if e != nil {
		d.SectionErrors = append(d.SectionErrors, "Stage 3: "+e.Error())
		s3 = missingTrack("Stage 3", beginning.Add(5*24*time.Hour), e)
	}
	for _, m := range s3.Measures {
		if m.Name == "meter record error" {
			d.SectionErrors = append(d.SectionErrors, "Stage 3: "+m.Note)
		}
	}
	s1, e := stage1(r, lines)
	if e != nil {
		d.SectionErrors = append(d.SectionErrors, "Stage 1: "+e.Error())
		s1 = missingTrack("Stage 1", beginning.Add(6*24*time.Hour), e)
	}
	a, e := apple(r)
	if e != nil {
		d.SectionErrors = append(d.SectionErrors, "Apple: "+e.Error())
		a = missingTrack("Apple", beginning.Add(6*24*time.Hour), e)
	}
	d.Tracks = []track{s3, s1, a}
	for i := range d.Tracks {
		if e = finish(&d.Tracks[i], r, now); e != nil {
			d.SectionErrors = append(d.SectionErrors, d.Tracks[i].Name+": "+e.Error())
		}
	}
	d.Velocity, e = measureVelocity(r, now, false)
	if e != nil {
		d.Velocity.Error = e.Error()
		d.SectionErrors = append(d.SectionErrors, "velocity: "+e.Error())
	}
	type patchResult struct {
		result backlogResult
		err    error
	}
	patches := make(chan patchResult, 1)
	patchRepo := r
	patchRepo.cache = nil
	patchRepo.readErrors = nil
	if r.backlogContext != nil {
		patchRepo.ctx = r.backlogContext
	}
	pending := append([]string{}, d.Velocity.Pending...)
	metadataError := d.Velocity.Error
	go func() {
		if metadataError != "" {
			patches <- patchResult{err: fmt.Errorf("backlog metadata unavailable: %s", metadataError)}
			return
		}
		result, err := patchRepo.cachedPatchBacklog(pending)
		patches <- patchResult{result, err}
	}()
	if e = reconstruct(r, &d, lines); e != nil {
		d.SectionErrors = append(d.SectionErrors, "history: "+e.Error())
		initializeMissingHistory(&d)
	}
	d.Milestones, e = milestones(r, d)
	if e != nil {
		d.SectionErrors = append(d.SectionErrors, "milestones: "+e.Error())
	}
	patch := <-patches
	d.Velocity.PatchStats = &patch.result
	if patch.err != nil {
		d.Velocity.PatchError = patch.err.Error()
	} else {
		d.Velocity.PatchBacklog = &patch.result.Count
	}
	m := &d.Velocity.Measures[4]
	progress := m.Progress
	*m = velocityMetrics(d.Velocity, now)[4]
	points := progress.Timeline
	points[24] = point(*m, now, d.Main)
	m.Progress = calculate(points, time.Time{}, now)
	d.Slowest = slowest(d.Tracks, now)
	return d, nil
}
func missingTrack(name string, deadline time.Time, err error) track {
	return track{Name: name, Deadline: deadline, Measures: []metric{unknown("track measurements", 1, name, "section unavailable: "+err.Error())}}
}
func patchCount(count *int) string {
	if count == nil {
		return "unknown"
	}
	return strconv.Itoa(*count)
}
func printMetric(w io.Writer, m metric) {
	percent := "unknown"
	raw := fmt.Sprintf("%.0f/%.0f", m.Done, m.Total)
	if strings.HasPrefix(m.Name, "overall") {
		raw = fmt.Sprintf("%.3f/%.0f goal equivalents", m.Done, m.Total)
	}
	if m.Percent != nil {
		percent = fmt.Sprintf("%.1f%%", *m.Percent)
		if m.Name == "syntax-lint native/Go speed" {
			raw = fmt.Sprintf("%.3fx Go speed", m.Done)
		}
	} else if !m.Known {
		raw = fmt.Sprintf("?/%.0f", m.Total)
		if m.Total == 0 {
			raw = "?/unknown target"
		}
	} else {
		raw = fmt.Sprintf("%.0f changed lines", m.Done)
	}
	fmt.Fprintf(w, "  %-38s %s %7s %s", m.Name, bar(m.Percent), percent, raw)
	if m.Note != "" {
		fmt.Fprintf(w, "; %s", m.Note)
	}
	fmt.Fprintln(w)
}
func render(w io.Writer, d dashboard) { renderReport(w, d, false) }
func renderReport(w io.Writer, d dashboard, history bool) {
	fmt.Fprintf(w, "Adamic progress | %s | day %d of creation | main %.12s\n", d.Time.In(mdt).Format("2006-01-02 15:04:05 MST"), d.Day, d.Main)
	var upcoming []milestoneResult
	var overdue []milestoneResult
	for _, g := range d.Milestones {
		if g.Overdue {
			overdue = append(overdue, g)
		} else {
			upcoming = append(upcoming, g)
		}
	}
	if len(overdue) > 0 {
		renderMilestones(w, overdue)
	}
	for _, t := range d.Tracks {
		fmt.Fprintf(w, "\n%s | deadline %s | %.1fh left\n", t.Name, t.Deadline.In(mdt).Format("Jan 2 15:04 MST"), t.SecondsLeft/3600)
		printMetric(w, t.Overall)
		printTrend(w, t.Overall, t.Deadline)
		for _, m := range t.Measures {
			printMetric(w, m)
			printTrend(w, m, t.Deadline)
		}
		eta := "not enough history yet"
		if t.ETA != nil {
			eta = t.ETA.In(mdt).Format("Jan 2 15:04 MST")
		}
		fmt.Fprintf(w, "  ETA: %s; %s\n", eta, t.ETAScope)
	}
	v := d.Velocity
	if v.Error != "" {
		fmt.Fprintln(w, "\nVelocity missing:", v.Error)
	} else {
		fmt.Fprintf(w, "\nVelocity | %d commits on main | %d landed/24h | %d landed/1h | %d distinct commits waiting (identity count) | %s distinct patches waiting (patch-id count)\n", v.Total, v.Day, v.Hour, v.Backlog, patchCount(v.PatchBacklog))
	}
	if v.PatchError != "" {
		fmt.Fprintln(w, "Backlog missing:", v.PatchError)
	}
	for _, err := range d.SectionErrors {
		fmt.Fprintln(w, "Section missing:", err)
	}
	if v.Note != "" {
		fmt.Fprintln(w, "Landings/hour:", v.Note)
	} else {
		var keys []string
		for k := range v.Landings {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Fprintf(w, "Landings/hour %s: %d\n", k, v.Landings[k])
		}
	}
	for _, m := range v.Measures {
		printTrend(w, m, time.Time{})
	}
	renderMilestones(w, upcoming)
	if history {
		renderHistory(w, d)
	}
	fmt.Fprintln(w, d.Slowest)
}
func run(args []string, out, errOut io.Writer) int {
	flags := flag.NewFlagSet("adamic-progress", flag.ContinueOnError)
	flags.SetOutput(errOut)
	flags.Usage = func() { fmt.Fprint(errOut, help) }
	asJSON := flags.Bool("json", false, "write JSON")
	history := flags.Bool("history", false, "print per-hour measurements for the last 24 hours")
	if e := flags.Parse(args); e != nil {
		if errors.Is(e, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 {
		flags.Usage()
		return 2
	}
	rootCtx, rootCancel := context.WithCancel(context.Background())
	defer rootCancel()
	ctx := rootCtx
	d, e := collect(repository{root: ".", ref: "origin/main", ctx: ctx, backlogContext: rootCtx}, time.Now(), stage1progress.NewMeasurer(ctx))
	if e != nil {
		fmt.Fprintln(errOut, "adamic-progress:", e)
		return 1
	}
	if *asJSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		if e = encoder.Encode(d); e != nil {
			fmt.Fprintln(errOut, e)
			return 1
		}
	} else {
		renderReport(out, d, *history)
	}
	return 0
}
func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

type meterFile struct {
	File     string `json:"file"`
	Source   bool   `json:"source"`
	Checker  bool   `json:"checker"`
	Lowering bool   `json:"lowering"`
	Own      *bool  `json:"checker_own_file,omitempty"`
}

func validateMeter(files []meterFile, source, checker, lowering int) error {
	if files == nil {
		return errors.New("meter record missing per-file observations")
	}
	seen := map[string]bool{}
	sources, checked, lowered := 0, 0, 0
	for _, f := range files {
		if f.File == "" || seen[f.File] {
			return errors.New("missing or duplicate meter file")
		}
		seen[f.File] = true
		if f.Lowering && !f.Checker {
			return errors.New("lowering without checker")
		}
		if f.Source && strings.HasPrefix(f.File, "src/compiler/") {
			sources++
			if f.Checker {
				checked++
			}
			if f.Lowering {
				lowered++
			}
		} else if f.Checker || f.Lowering {
			return errors.New("non-compiler source credited")
		}
	}
	if sources != source || checked != checker || lowered != lowering {
		return fmt.Errorf("inconsistent compiler totals: recorded source/checker/lowering=%d/%d/%d; per-file=%d/%d/%d", source, checker, lowering, sources, checked, lowered)
	}
	return nil
}
