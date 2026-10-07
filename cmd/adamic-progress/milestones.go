package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	plan "github.com/system-inc/adamic/documentation/progress"
)

type condition struct {
	Kind     string `json:"kind"`
	Track    string `json:"track,omitempty"`
	Name     string `json:"name,omitempty"`
	Ref      string `json:"ref,omitempty"`
	File     string `json:"file,omitempty"`
	Pointer  string `json:"pointer,omitempty"`
	Operator string `json:"operator,omitempty"`
	Value    any    `json:"value,omitempty"`
	Prefix   string `json:"prefix,omitempty"`
	Suffix   string `json:"suffix,omitempty"`
}
type milestone struct {
	Owner       string         `json:"owner,omitempty"`
	Source      string         `json:"source,omitempty"`
	Observation map[string]any `json:"observation,omitempty"`
	ID          string         `json:"id"`
	Track       string         `json:"track"`
	Claim       string         `json:"claim"`
	Deadline    time.Time      `json:"deadline"`
	Measurement string         `json:"measurement"`
	Conditions  []condition    `json:"conditions"`
}
type conditionResult struct {
	Known   bool   `json:"known"`
	Done    bool   `json:"done"`
	OnTrack bool   `json:"on_track"`
	Note    string `json:"note"`
}
type milestoneResult struct {
	milestone
	Status      string            `json:"status"`
	Color       string            `json:"color"`
	SecondsLeft float64           `json:"seconds_left"`
	Overdue     bool              `json:"overdue"`
	Evidence    []conditionResult `json:"evidence"`
}

func (r repository) forRef(ref string) (repository, bool, error) {
	if ref == "main" || ref == "" {
		return r, true, nil
	}
	b, e := r.git("rev-parse", "--verify", ref+"^{commit}")
	if e != nil {
		return r, false, nil
	}
	r.ref = strings.TrimSpace(string(b))
	b, e = r.git("ls-tree", "-r", "--name-only", r.ref)
	if e != nil {
		return r, false, e
	}
	r.paths = strings.Fields(string(b))
	return r, true, nil
}
func metricNamed(d dashboard, track, name string) (metric, bool) {
	for _, t := range d.Tracks {
		if t.Name != track {
			continue
		}
		for _, m := range t.Measures {
			if m.Name == name {
				return m, true
			}
		}
	}
	return metric{}, false
}
func threshold(m metric, target float64, deadline, now time.Time) conditionResult {
	result := conditionResult{Known: m.Known, Note: fmt.Sprintf("%s: %s/%g; %s", m.Name, value(func() *float64 {
		if !m.Known {
			return nil
		}
		v := m.Done
		return &v
	}()), target, m.Source)}
	if !m.Known {
		result.Note += "; not measurable yet"
		return result
	}
	result.Done = m.Done >= target
	if result.Done {
		result.OnTrack = true
		return result
	}
	hours := deadline.Sub(now).Hours()
	if hours > 0 && m.Progress != nil && m.Progress.Actual != nil {
		result.OnTrack = *m.Progress.Actual >= (target-m.Done)/hours
	}
	return result
}

type backendObservation struct {
	Stdout *string `json:"stdout"`
	Stderr *string `json:"stderr"`
	Exit   *int    `json:"exit"`
}

func equalBackend(a, b *backendObservation) bool {
	return a != nil && b != nil && a.Stdout != nil && b.Stdout != nil && a.Stderr != nil && b.Stderr != nil && a.Exit != nil && b.Exit != nil && *a.Stdout == *b.Stdout && *a.Stderr == *b.Stderr && *a.Exit == *b.Exit
}

// Fixture identities from codex/stage3-fixtures-host 1037217.
const hostFixtureNames = `01_readFile_utf8.a
02_readFile_utf16le.a
03_readFile_utf16be.a
04_readFile_missing.a
05_writeFile.a
06_fileExists.a
07_directoryExists.a
08_getDirectories.a
09_realpath.a
10_getModifiedTime.a
11_setModifiedTime.a
12_deleteFile.a
13_createDirectory.a
14_getCurrentDirectory.a
15_getExecutingFilePath.a
16_getEnvironmentVariable.a
17_write.a
18_exit_0.a
19_exit_1.a
20_exit_2.a
21_createHash.a
22_createHash_fallback.a
23_newLine.a
24_useCaseSensitiveFileNames.a
25_readDirectory.a`

func hostMeasure(r repository) (metric, error) {
	const file = "stage3/fixtures/host/status.json"
	b, ok, e := r.blob(file)
	if e != nil {
		return metric{}, e
	}
	if !ok {
		return unknown("host fixtures both backends", 25, r.ref+":"+file, "not measurable yet"), nil
	}
	var rows []struct {
		File       string              `json:"file"`
		Node       *backendObservation `json:"node"`
		Native     *backendObservation `json:"native"`
		JavaScript *backendObservation `json:"javascript"`
		Stage0     struct {
			Outcome string `json:"outcome"`
		} `json:"stage0"`
	}
	if e = json.Unmarshal(b, &rows); e != nil {
		return metric{}, fmt.Errorf("%s: %w", file, e)
	}
	present := map[string]bool{}
	sources := 0
	for _, p := range r.paths {
		if strings.HasPrefix(p, "stage3/fixtures/host/") && strings.HasSuffix(p, ".a") {
			present[strings.TrimPrefix(p, "stage3/fixtures/host/")] = true
			sources++
		}
	}
	for _, name := range strings.Fields(hostFixtureNames) {
		if !present[name] {
			return unknown("host fixtures both backends", 25, r.ref+":"+file, "not measurable yet: missing pinned fixture "+name), nil
		}
	}
	seen := map[string]bool{}
	green, unrecorded := 0, 0
	for _, row := range rows {
		if seen[row.File] || !present[row.File] {
			return metric{}, fmt.Errorf("%s duplicate or absent fixture %s", file, row.File)
		}
		seen[row.File] = true
		if row.Node == nil || row.Node.Stdout == nil || row.Node.Stderr == nil || row.Node.Exit == nil {
			return metric{}, fmt.Errorf("%s missing Node observation %s", file, row.File)
		}
		switch row.Stage0.Outcome {
		case "Checker", "Refused", "NotYet":
			continue
		case "Compiles":
			if !completeBackend(row.Native) || !completeBackend(row.JavaScript) {
				unrecorded++
				continue
			}
			if equalBackend(row.Native, row.Node) && equalBackend(row.JavaScript, row.Node) {
				green++
			}
		default:
			return metric{}, fmt.Errorf("%s unknown outcome for %s", file, row.File)
		}
	}
	if sources != 25 || len(rows) != 25 || unrecorded > 0 {
		return unknown("host fixtures both backends", 25, r.ref+":"+file, fmt.Sprintf("not measurable yet: %d confirmed green; %d/%d records, %d missing backend results", green, len(rows), sources, unrecorded)), nil
	}
	return measured("host fixtures both backends", float64(green), 25, r.ref+":"+file), nil
}
func recordValue(r repository, c condition) (any, bool, error) {
	var data any
	ok, e := r.decode(c.File, &data)
	if e != nil || !ok {
		return nil, ok, e
	}
	value := data
	for _, part := range strings.Split(c.Pointer, ".") {
		object, ok := value.(map[string]any)
		if !ok {
			return nil, false, nil
		}
		value, ok = object[part]
		if !ok {
			return nil, false, nil
		}
	}
	return value, true, nil
}
func compareRecord(actual any, operator string, expected any) (bool, error) {
	switch operator {
	case "equals":
		a, e := json.Marshal(actual)
		if e != nil {
			return false, e
		}
		b, e := json.Marshal(expected)
		return string(a) == string(b), e
	case "at_least", "less_than":
		a, ok := actual.(float64)
		b, valid := expected.(float64)
		if !ok || !valid || math.IsNaN(a) {
			return false, errors.New("numeric milestone record required")
		}
		if operator == "at_least" {
			return a >= b, nil
		}
		return a < b, nil
	case "nonempty":
		list, ok := actual.([]any)
		if !ok {
			return false, errors.New("milestone flags must be an array")
		}
		if len(list) == 0 {
			return false, nil
		}
		for _, flag := range list {
			text, ok := flag.(string)
			if !ok || strings.TrimSpace(text) == "" {
				return false, errors.New("milestone flags must have names")
			}
		}
		return true, nil
	}
	return false, fmt.Errorf("unknown milestone operator %s", operator)
}
func checkCondition(r repository, d dashboard, goal milestone, c condition) (conditionResult, error) {
	absent := conditionResult{Note: "not measurable yet"}
	switch c.Kind {
	case "metric":
		m, ok := metricNamed(d, c.Track, c.Name)
		if !ok {
			return absent, nil
		}
		target, ok := c.Value.(float64)
		if !ok {
			return absent, errors.New("numeric milestone target required")
		}
		return threshold(m, target, goal.Deadline, d.Time), nil
	case "host":
		scope, ok, e := r.forRef(c.Ref)
		if e != nil || !ok {
			return absent, e
		}
		m, e := hostMeasure(scope)
		if e != nil {
			return absent, e
		}
		target, ok := c.Value.(float64)
		if !ok {
			return absent, errors.New("host target required")
		}
		if c.Ref == "main" {
			if recorded, found := metricNamed(d, "Stage 3", m.Name); found {
				m.Progress = recorded.Progress
			}
		}
		return threshold(m, target, goal.Deadline, d.Time), nil
	case "record":
		v, ok, e := recordValue(r, c)
		if e != nil || !ok {
			return absent, e
		}
		done, e := compareRecord(v, c.Operator, c.Value)
		return conditionResult{Known: true, Done: done, OnTrack: done, Note: c.File + "#" + c.Pointer}, e
	case "path", "path_prefix":
		found := false
		for _, p := range r.paths {
			if (c.Kind == "path" && p == c.File) || (c.Kind == "path_prefix" && strings.HasPrefix(p, c.Prefix) && strings.HasSuffix(p, c.Suffix)) {
				found = true
				break
			}
		}
		return conditionResult{Known: true, Done: found, OnTrack: found, Note: "main file presence: " + c.File + c.Prefix + c.Suffix}, nil
	case "inventory_accounted":
		m, ok := metricNamed(d, "Stage 1", "lint rules")
		if !ok || !m.Known {
			return absent, nil
		}
		var data struct {
			Refused map[string]string `json:"refused_rules"`
		}
		ok, e := r.decode("stage1/progress.json", &data)
		if e != nil {
			return absent, e
		}
		if !ok {
			return absent, nil
		}
		inventory := map[string]bool{}
		for _, name := range m.InventoryNames {
			inventory[name] = true
		}
		accounted := map[string]bool{}
		for _, name := range m.Registered {
			accounted[name] = true
		}
		for name, reason := range data.Refused {
			if !inventory[name] || strings.TrimSpace(reason) == "" {
				return absent, fmt.Errorf("invalid refused inventory rule %s", name)
			}
			accounted[name] = true
		}
		return conditionResult{Known: true, Done: len(accounted) == int(m.Total), OnTrack: len(accounted) == int(m.Total), Note: fmt.Sprintf("%d/%g rules ported or explicitly refused", len(accounted), m.Total)}, nil
	case "host_loader":
		var record struct {
			Commit   string   `json:"hook_commit"`
			Branches []string `json:"branches"`
		}
		ok, e := r.decode("documentation/progress/host-loader.json", &record)
		if e != nil || !ok {
			return absent, e
		}
		if record.Commit == "" || len(record.Branches) != 4 {
			return absent, errors.New("host-loader record needs hook_commit and four branches")
		}
		seen := map[string]bool{}
		done := true
		for _, ref := range record.Branches {
			if seen[ref] {
				return absent, errors.New("duplicate host loader branch")
			}
			seen[ref] = true
			_, ok, e := r.forRef(ref)
			if e != nil || !ok {
				return absent, e
			}
			if _, e = r.git("merge-base", "--is-ancestor", record.Commit, ref); e != nil {
				done = false
			}
		}
		return conditionResult{Known: true, Done: done, OnTrack: done, Note: "hook ancestry in all four named host branches"}, nil
	case "parse_instructions":
		return checkParseInstructions(r, c)
	case "backlog_hourly":
		return checkBacklog(r, goal.Deadline, d.Time)
	}
	return absent, fmt.Errorf("unknown milestone kind %s", c.Kind)
}
func checkBacklog(r repository, start, now time.Time) (conditionResult, error) {
	if now.Before(start) {
		return conditionResult{Known: true, OnTrack: true, Note: "hourly obligation starts at deadline"}, nil
	}
	b, ok, e := r.blob("documentation/velocity/patch-backlog.csv")
	if e != nil || !ok {
		return conditionResult{Note: "not measurable yet: no archived hourly patch backlog"}, e
	}
	rows, e := csv.NewReader(strings.NewReader(string(b))).ReadAll()
	if e != nil {
		return conditionResult{}, e
	}
	if len(rows) == 0 || len(rows[0]) != 2 || rows[0][0] != "timestamp" || rows[0][1] != "count" {
		return conditionResult{}, errors.New("backlog.csv requires timestamp,count")
	}
	counts := map[int]int{}
	for _, row := range rows[1:] {
		stamp, e := time.Parse(time.RFC3339, row[0])
		if e != nil {
			return conditionResult{}, e
		}
		count, e := strconv.Atoi(row[1])
		if e != nil || count < 0 {
			return conditionResult{}, errors.New("invalid backlog count")
		}
		delta := stamp.Sub(start)
		if delta < 0 || stamp.After(now) {
			continue
		}
		if delta%time.Hour != 0 {
			return conditionResult{}, errors.New("backlog record must lie on an obligation hour")
		}
		h := int(delta / time.Hour)
		if _, exists := counts[h]; exists {
			return conditionResult{}, errors.New("duplicate backlog hour")
		}
		counts[h] = count
	}
	for h := 0; h <= int(now.Sub(start)/time.Hour); h++ {
		n, ok := counts[h]
		if !ok {
			return conditionResult{Note: "not measurable yet: missing backlog hour " + strconv.Itoa(h)}, nil
		}
		if h > 0 && n >= counts[h-1] {
			return conditionResult{Known: true, Note: "backlog did not fall at hour " + strconv.Itoa(h)}, nil
		}
	}
	return conditionResult{Known: true, OnTrack: true, Note: "backlog fell at every elapsed obligation hour; ongoing"}, nil
}
func milestones(r repository, d dashboard) ([]milestoneResult, error) {
	var goals []milestone
	b, ok, e := r.blob("documentation/progress/milestones.json")
	if e != nil {
		return nil, e
	}
	if !ok {
		b = plan.Milestones
	}
	if e = json.Unmarshal(b, &goals); e != nil {
		return nil, e
	}
	results := make([]milestoneResult, 0, len(goals))
	seen := map[string]bool{}
	for _, goal := range goals {
		_, offset := goal.Deadline.Zone()
		if seen[goal.ID] || goal.ID == "" || goal.Deadline.IsZero() || offset != -6*3600 || len(goal.Conditions) == 0 {
			return nil, errors.New("invalid micro-deadline definition")
		}
		seen[goal.ID] = true
		result := milestoneResult{milestone: goal, SecondsLeft: goal.Deadline.Sub(d.Time).Seconds(), Status: "done", Color: "green"}
		onTrack := true
		done := true
		for _, c := range goal.Conditions {
			observed, e := checkCondition(r, d, goal, c)
			if e != nil {
				return nil, fmt.Errorf("milestone %s: %w", goal.ID, e)
			}
			result.Evidence = append(result.Evidence, observed)
			done = done && observed.Done
			onTrack = onTrack && observed.OnTrack
		}
		if !done {
			result.Status = "red"
			result.Color = "red"
			if onTrack {
				result.Status = "on track"
				result.Color = "green"
			}
		}
		result.Overdue = result.SecondsLeft < 0 && result.Status == "red"
		results = append(results, result)
	}
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].Overdue != results[j].Overdue {
			return results[i].Overdue
		}
		return results[i].Deadline.Before(results[j].Deadline)
	})
	return results, nil
}
func renderMilestones(w io.Writer, goals []milestoneResult) {
	fmt.Fprintln(w, "\nMicro-deadlines (overdue red items first)")
	for _, g := range goals {
		marker := "🟢"
		if g.Color == "red" {
			marker = "🔴"
		}
		left := fmt.Sprintf("%.1fh left", g.SecondsLeft/3600)
		if g.Overdue {
			left = fmt.Sprintf("overdue %.1fh", -g.SecondsLeft/3600)
		}
		fmt.Fprintf(w, "%s %-8s %s %s | %s | %s\n", marker, g.Status, g.Track, g.Deadline.In(mdt).Format("Jan 2 15:04 MST"), left, g.Claim)
		for _, e := range g.Evidence {
			if !e.Known {
				fmt.Fprintf(w, "  %s\n", e.Note)
				break
			}
		}
	}
}

func completeBackend(o *backendObservation) bool {
	return o != nil && o.Stdout != nil && o.Stderr != nil && o.Exit != nil
}

// Parse-only instruction counts must cover the same complete batch on both sides.
func checkParseInstructions(r repository, c condition) (conditionResult, error) {
	absent := conditionResult{Note: "not measurable yet: batch8's 77 files require parse-only instruction counts on both sides"}
	var record struct {
		Parse *struct {
			Driver      string   `json:"driver"`
			Files       int      `json:"files"`
			Unit        string   `json:"unit"`
			NativeScope string   `json:"native_scope"`
			GoScope     string   `json:"go_scope"`
			Native      *float64 `json:"native_instructions"`
			Go          *float64 `json:"go_parse_instructions"`
		} `json:"parse_batch8"`
	}
	ok, err := r.decode("stage1/progress.json", &record)
	if err != nil || !ok {
		return absent, err
	}
	p := record.Parse
	if p == nil || p.Driver != "batch8" || p.Files != 77 || p.Unit != "instructions" || p.NativeScope != "parse_alone" || p.GoScope != "parse_alone" || p.Native == nil || p.Go == nil {
		return absent, nil
	}
	for _, count := range []float64{*p.Native, *p.Go} {
		if count <= 0 || math.IsInf(count, 0) || math.IsNaN(count) || math.Trunc(count) != count {
			return absent, errors.New("parse instruction counts must be positive integers")
		}
	}
	limit, ok := c.Value.(float64)
	if !ok || limit <= 0 || math.IsInf(limit, 0) || math.IsNaN(limit) {
		return absent, errors.New("invalid parse instruction ratio target")
	}
	ratio := *p.Native / *p.Go
	done := ratio <= limit
	return conditionResult{Known: true, Done: done, OnTrack: done, Note: fmt.Sprintf("batch8 77 files: native %.0f / Go parse-alone %.0f instructions = %.3fx; need <=%gx", *p.Native, *p.Go, ratio, limit)}, nil
}
