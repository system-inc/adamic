package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestMicroDeadlineStatusesAndOverdueFirst(t *testing.T) {
	t.Parallel()
	r, write, git := fixture(t)
	write("stage3/progress.json", `{"milestones":{"cycle_census_ruled":true}}`)
	write("documentation/progress/milestones.json", `[
 {"id":"onpace","track":"Stage 3","claim":"checker 40","deadline":"2026-10-08T07:00:00-06:00","measurement":"meter","conditions":[{"kind":"metric","track":"Stage 3","name":"compiler checker","value":40}]},
 {"id":"done","track":"Stage 3","claim":"ruled","deadline":"2026-10-07T02:00:00-06:00","measurement":"record","conditions":[{"kind":"record","file":"stage3/progress.json","pointer":"milestones.cycle_census_ruled","operator":"equals","value":true}]},
 {"id":"overdue","track":"Stage 3","claim":"78 lowering","deadline":"2026-10-07T06:00:00-06:00","measurement":"meter","conditions":[{"kind":"metric","track":"Stage 3","name":"compiler lowering","value":78}]}
 ]`)
	git("add", ".")
	git("commit", "-qm", "plan")
	r.ref = "HEAD"
	r.paths = []string{"stage3/progress.json", "documentation/progress/milestones.json"}
	now := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)
	checker := measured("compiler checker", 30, 78, "meter")
	checker.Progress = &trend{Actual: pointer(10)}
	lowering := measured("compiler lowering", 0, 78, "meter")
	d := dashboard{Time: now, Tracks: []track{{Name: "Stage 3", Measures: []metric{checker, lowering}}}}
	goals, e := milestones(r, d)
	if e != nil {
		t.Fatal(e)
	}
	if goals[0].ID != "overdue" || !goals[0].Overdue || goals[0].Status != "red" {
		t.Fatal("overdue not first", goals)
	}
	statuses := map[string]string{}
	for _, g := range goals {
		statuses[g.ID] = g.Status
	}
	if statuses["onpace"] != "on track" || statuses["done"] != "done" {
		t.Fatal(statuses)
	}
}
func hostFixtureRecords(t *testing.T, write func(string, string), green int, missingJS bool) {
	t.Helper()
	var rows []map[string]any
	for index, name := range strings.Fields(hostFixtureNames) {
		n := index + 1
		write("stage3/fixtures/host/"+name, "fixture")
		observation := map[string]any{"stdout": "true\n", "stderr": "", "exit": 0}
		row := map[string]any{"file": name, "node": observation, "stage0": map[string]any{"outcome": "Checker"}}
		if n <= green {
			row["stage0"] = map[string]any{"outcome": "Compiles"}
			row["native"] = observation
			if !missingJS {
				row["javascript"] = observation
			}
		}
		rows = append(rows, row)
	}
	write("stage3/fixtures/host/status.json", string(mustJSON(t, rows)))
}
func snapshotPaths(t *testing.T, r repository, git func(...string) string) repository {
	t.Helper()
	git("add", ".")
	git("commit", "-qm", "evidence")
	r.ref = "HEAD"
	b, e := r.git("ls-tree", "-r", "--name-only", "HEAD")
	if e != nil {
		t.Fatal(e)
	}
	r.paths = strings.Fields(string(b))
	return r
}
func TestHostBothBackendsAndMissingResults(t *testing.T) {
	t.Parallel()
	r, write, git := fixture(t)
	m, e := hostMeasure(r)
	if e != nil || m.Known || m.Note != "not measurable yet" {
		t.Fatal("missing host invented", m, e)
	}
	hostFixtureRecords(t, write, 12, false)
	r = snapshotPaths(t, r, git)
	m, e = hostMeasure(r)
	if e != nil || !m.Known || m.Done != 12 || m.Total != 25 {
		t.Fatal("host count", m, e)
	}
	hostFixtureRecords(t, write, 25, true)
	r = snapshotPaths(t, r, git)
	m, e = hostMeasure(r)
	if e != nil || m.Known || !strings.Contains(m.Note, "not measurable yet") {
		t.Fatal("missing JS credited", m, e)
	}
	hostFixtureRecords(t, write, 25, false)
	r = snapshotPaths(t, r, git)
	m, e = hostMeasure(r)
	if e != nil || m.Done != 25 {
		t.Fatal("full host", m, e)
	}
	// Native output differs by one byte; Node and JS still agree exactly.
	b, _, e := r.blob("stage3/fixtures/host/status.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []map[string]any
	if e = json.Unmarshal(b, &rows); e != nil {
		t.Fatal(e)
	}
	rows[0]["native"].(map[string]any)["stdout"] = "false\n"
	write("stage3/fixtures/host/status.json", string(mustJSON(t, rows)))
	r = snapshotPaths(t, r, git)
	m, e = hostMeasure(r)
	if e != nil || m.Done != 24 {
		t.Fatal("native mismatch credited", m, e)
	}
	rows[0]["native"].(map[string]any)["stdout"] = rows[0]["node"].(map[string]any)["stdout"].(string)
	rows[1]["javascript"].(map[string]any)["stdout"] = "wrong JS\n"
	write("stage3/fixtures/host/status.json", string(mustJSON(t, rows)))
	r = snapshotPaths(t, r, git)
	m, e = hostMeasure(r)
	if e != nil || m.Done != 24 {
		t.Fatal("JavaScript mismatch credited", m, e)
	}

}
func TestHostCheckpointScopeAndLoaderAncestry(t *testing.T) {
	t.Parallel()
	r, write, git := fixture(t)
	hostFixtureRecords(t, write, 25, false)
	r = snapshotPaths(t, r, git)
	goal := milestone{Deadline: time.Now().Add(time.Hour)}
	d := dashboard{Time: time.Now()}
	result, e := checkCondition(r, d, goal, condition{Kind: "host", Ref: "origin/area/library", Value: float64(12)})
	if e != nil || result.Known {
		t.Fatal("wrong branch host credited", result, e)
	}
	git("update-ref", "refs/remotes/origin/area/library", "HEAD")
	result, e = checkCondition(r, d, goal, condition{Kind: "host", Ref: "origin/area/library", Value: float64(25)})
	if e != nil || !result.Done {
		t.Fatal("library host not measured", result, e)
	}
	hook := git("rev-parse", "HEAD")
	var branches []string
	for n := 0; n < 4; n++ {
		ref := fmt.Sprintf("origin/host%d", n)
		git("update-ref", "refs/remotes/"+ref, "HEAD")
		branches = append(branches, ref)
	}
	write("documentation/progress/host-loader.json", string(mustJSON(t, map[string]any{"hook_commit": hook, "branches": branches})))
	r = snapshotPaths(t, r, git)
	result, e = checkCondition(r, d, goal, condition{Kind: "host_loader"})
	if e != nil || !result.Done {
		t.Fatal("four loader branches", result, e)
	}
	// Move one host ref to a commit that cannot contain the recorded hook.
	git("update-ref", "refs/remotes/origin/host3", hook)
	newHook := git("rev-parse", "HEAD")
	write("documentation/progress/host-loader.json", string(mustJSON(t, map[string]any{"hook_commit": newHook, "branches": branches})))
	r = snapshotPaths(t, r, git)
	result, e = checkCondition(r, d, goal, condition{Kind: "host_loader"})
	if e != nil || result.Done {
		t.Fatal("hook absent from branches credited", result, e)
	}
}
func TestBacklogMustFallEveryElapsedHour(t *testing.T) {
	t.Parallel()
	r, write, git := fixture(t)
	start := time.Date(2026, 10, 7, 6, 0, 0, 0, mdt)
	now := start.Add(2 * time.Hour)
	write("documentation/velocity/patch-backlog.csv", "timestamp,count\n2026-10-07T06:00:00-06:00,20\n2026-10-07T07:00:00-06:00,18\n2026-10-07T08:00:00-06:00,17\n")
	r = snapshotPaths(t, r, git)
	observed, e := checkBacklog(r, start, now)
	if e != nil || !observed.OnTrack || observed.Done {
		t.Fatal("ongoing hourly backlog", observed, e)
	}
	write("documentation/velocity/patch-backlog.csv", "timestamp,count\n2026-10-07T06:00:00-06:00,20\n2026-10-07T07:00:00-06:00,20\n2026-10-07T08:00:00-06:00,17\n")
	r = snapshotPaths(t, r, git)
	observed, e = checkBacklog(r, start, now)
	if e != nil || observed.OnTrack || !observed.Known {
		t.Fatal("flat backlog hidden", observed, e)
	}
	write("documentation/velocity/patch-backlog.csv", "timestamp,count\n2026-10-07T06:00:00-06:00,20\n2026-10-07T08:00:00-06:00,17\n")
	r = snapshotPaths(t, r, git)
	observed, e = checkBacklog(r, start, now)
	if e != nil || observed.Known || !strings.Contains(observed.Note, "missing backlog hour") {
		t.Fatal("missing hour credited", observed, e)
	}
}
func TestRefusedInventoryRequiresReasons(t *testing.T) {
	t.Parallel()
	r, write, git := fixture(t)
	write("stage1/progress.json", `{"refused_rules":{"two":"unsupported checker API"}}`)
	r = snapshotPaths(t, r, git)
	m := measured("lint rules", 1, 2, "fixture")
	m.Registered = []string{"one"}
	m.InventoryNames = []string{"one", "two"}
	d := dashboard{Tracks: []track{{Name: "Stage 1", Measures: []metric{m}}}}
	result, e := checkCondition(r, d, milestone{}, condition{Kind: "inventory_accounted"})
	if e != nil || !result.Done {
		t.Fatal(result, e)
	}
	write("stage1/progress.json", `{"refused_rules":{"two":""}}`)
	r = snapshotPaths(t, r, git)
	if _, e = checkCondition(r, d, milestone{}, condition{Kind: "inventory_accounted"}); e == nil {
		t.Fatal("empty refusal reason counted")
	}
}

func TestBatch8ParseInstructionMilestones(t *testing.T) {
	t.Parallel()
	r, write, git := fixture(t)
	condition := condition{Kind: "parse_instructions", Value: 1.5}
	observe := func(native, goCount float64, scope string, files int) conditionResult {
		t.Helper()
		write("stage1/progress.json", string(mustJSON(t, map[string]any{"parse_batch8": map[string]any{"driver": "batch8", "files": files, "unit": "instructions", "native_scope": "parse_alone", "go_scope": scope, "native_instructions": native, "go_parse_instructions": goCount}})))
		r = snapshotPaths(t, r, git)
		result, err := checkParseInstructions(r, condition)
		if err != nil {
			t.Fatal(err)
		}
		return result
	}
	if result := observe(8970000000, 2960000000, "whole_run", 77); result.Known || result.Done {
		t.Fatal("whole Go run credited", result)
	}
	if result := observe(150, 100, "parse_alone", 77); !result.Done {
		t.Fatal("inclusive 1.5x boundary", result)
	}
	if result := observe(151, 100, "parse_alone", 77); result.Done || !result.Known {
		t.Fatal("over 1.5x credited", result)
	}
	if result := observe(100, 100, "parse_alone", 76); result.Known {
		t.Fatal("incomplete batch credited", result)
	}
	condition.Value = float64(1)
	if result := observe(100, 100, "parse_alone", 77); !result.Done {
		t.Fatal("Go equality rejected", result)
	}
	if result := observe(101, 100, "parse_alone", 77); result.Done {
		t.Fatal("slower than Go credited", result)
	}
	write("stage1/progress.json", `{"parse_batch8":{"driver":"batch8","files":77,"unit":"instructions","native_scope":"parse_alone","go_scope":"parse_alone","native_instructions":100,"go_parse_instructions":0}}`)
	r = snapshotPaths(t, r, git)
	if _, err := checkParseInstructions(r, condition); err == nil {
		t.Fatal("zero denominator accepted")
	}
	observe(100, 100, "parse_alone", 77)
	// The shipped plan carries the owner, source, context, and exact MDT deadlines.
	goals, err := milestones(r, dashboard{Time: time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatal(err)
	}
	found := 0
	for _, goal := range goals {
		if !strings.HasPrefix(goal.ID, "batch8-parse-") {
			continue
		}
		found++
		want := "2026-10-08T12:00:00-06:00"
		if goal.ID == "batch8-parse-go" {
			want = "2026-10-09T12:00:00-06:00"
		}
		if goal.Owner != "runtime" || goal.Source != "#93z4yv7" || goal.Deadline.Format(time.RFC3339) != want || goal.Observation["go_whole_run_instructions"] != float64(2960000000) {
			t.Fatal("plan metadata lost", goal)
		}
	}
	if found != 2 {
		t.Fatal("missing parse milestones", found)
	}
}
