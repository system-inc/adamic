package main

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"os/exec"

	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/stage1progress"
)

func pointer(n float64) *float64 { return &n }
func TestSixHourRateNeededAndSparkline(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)
	points := make([]observation, 25)
	for i := range points {
		points[i] = observation{Time: now.Add(time.Duration(i-24) * time.Hour), Value: pointer(float64(i)), Total: pointer(36)}
	}
	p := calculate(points, now.Add(6*time.Hour), now)
	if *p.Actual != 1 || *p.Needed != 2 || p.Status != "red" || *p.Hour != 23 || *p.SixHours != 18 || *p.Day != 0 || p.Window != 6 {
		t.Fatalf("rate/history %#v", p)
	}
	if p.Sparkline != "▁▁▂▂▂▂▃▃▃▄▄▄▅▅▅▅▆▆▆▇▇▇▇██" {
		t.Fatal("sparkline inverted or incorrectly scaled", p.Sparkline)
	}
	p = calculate(points, now.Add(24*time.Hour), now)
	if p.Status != "green" || *p.Needed != .5 {
		t.Fatal("green pace", p)
	}
	points[18].Value = nil
	p = calculate(points, now.Add(24*time.Hour), now)
	if p.Window != 5 || *p.Actual != 1 {
		t.Fatal("missing observation invented", p)
	}
	for i := range points {
		points[i].Value = nil
	}
	p = calculate(points, now.Add(time.Hour), now)
	if p.Actual != nil || p.Needed != nil || p.Sparkline != strings.Repeat("?", 25) {
		t.Fatal("unknown treated as zero", p)
	}
}
func TestFlatRegressingAndOverdueRate(t *testing.T) {
	t.Parallel()
	now := time.Now()
	points := make([]observation, 25)
	for i := range points {
		points[i] = observation{Time: now.Add(time.Duration(i-24) * time.Hour), Value: pointer(5), Total: pointer(10)}
	}
	p := calculate(points, now.Add(time.Hour), now)
	if *p.Actual != 0 || p.Status != "red" {
		t.Fatal("flat forecast", p)
	}
	points[24].Value = pointer(2)
	p = calculate(points, now.Add(time.Hour), now)
	if *p.Actual != -0.5 || p.Status != "red" {
		t.Fatal("regression hidden", p)
	}
	p = calculate(points, now.Add(-time.Hour), now)
	if p.Needed != nil || !strings.Contains(p.Note, "deadline passed") {
		t.Fatal("overdue required pace invented", p)
	}
}
func TestHourlyGitReconstruction(t *testing.T) {
	t.Parallel()
	r, write, git := fixture(t)
	now := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)
	write("stage1/cohere/lint/inventory/inventory.json", `{"rules":[{"name":"one","family":"core","variable":"One"},{"name":"two","family":"core","variable":"Two"},{"name":"three","family":"core","variable":"Three"},{"name":"four","family":"core","variable":"Four"},{"name":"five","family":"core","variable":"Five"},{"name":"six","family":"core","variable":"Six"},{"name":"seven","family":"core","variable":"Seven"},{"name":"eight","family":"core","variable":"Eight"}]}`)
	commit := func(age time.Duration, coverage int, registrations string) {
		t.Helper()
		write("coverage.json", string(mustJSON(t, coverage)))
		write("stage1/cohere/lint/testdata/oracle.go", `package main; import rules "github.com/system-inc/cohere/internal/lint/rules/core"; var subjects=[]rule.Rule{`+registrations+`}`)
		git("add", ".")
		c := exec.Command("git", "-C", r.root, "commit", "-qm", "recorded snapshot")
		stamp := now.Add(-age).Format(time.RFC3339)
		c.Env = append(os.Environ(), "GIT_AUTHOR_DATE="+stamp, "GIT_COMMITTER_DATE="+stamp)
		if b, e := c.CombinedOutput(); e != nil {
			t.Fatalf("commit %v %s", e, b)
		}
	}
	commit(25*time.Hour, 10, "rules.One")
	write("examples/apple/window.a", "window")
	commit(7*time.Hour, 20, "rules.One,rules.Two")
	write("examples/apple/counter.a", "counter")
	commit(2*time.Hour, 30, "rules.One,rules.Two,rules.Three")
	commit(time.Hour/2, 50, "rules.One,rules.Two,rules.Three,rules.Four")
	git("update-ref", "refs/remotes/origin/main", "HEAD")
	lines := func(root, ref string) (*stage1progress.Report, error) {
		b, e := exec.Command("git", "-C", root, "show", ref+":coverage.json").Output()
		if e != nil {
			return nil, e
		}
		var n int
		e = json.Unmarshal(b, &n)
		return &stage1progress.Report{Total: 100, Ported: n}, e
	}
	d, e := collect(r, now, lines)
	if e != nil {
		t.Fatal(e)
	}
	p := d.Tracks[1].Measures[0].Progress
	if *p.Now != 50 || *p.Hour != 30 || *p.SixHours != 20 || *p.Day != 10 || *p.Actual != 5 || p.Status != "green" {
		t.Fatalf("main reconstruction %#v", p)
	}
	rules := d.Tracks[1].Measures[1].Progress
	if *rules.Now != 4 || *rules.Hour != 3 || *rules.SixHours != 2 || *rules.Day != 1 || math.Abs(*rules.Actual-1.0/3) > 1e-9 {
		t.Fatalf("registered rule history %#v", rules)
	}
	// A branch author date must not make its unlanded commit appear on main.
	write("coverage.json", "99")
	commit(10*time.Minute, 99, "rules.One")
	git("update-ref", "refs/remotes/origin/pending", "HEAD")
	d, e = collect(r, now, lines)
	if e != nil {
		t.Fatal(e)
	}
	if *d.Tracks[1].Measures[0].Progress.Now != 50 {
		t.Fatal("pending commit credited")
	}
	var out bytes.Buffer
	renderReport(&out, d, true)
	if !strings.Contains(out.String(), "Per-hour measurements") || !strings.Contains(out.String(), "doing 5.000/h") || !strings.HasSuffix(out.String(), d.Slowest+"\n") {
		t.Fatal("history/rate presentation", out.String())
	}
}
func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, e := json.Marshal(v)
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func TestMainCommitAtBoundaries(t *testing.T) {
	t.Parallel()
	now := time.Now()
	log := []landing{{"new", now}, {"old", now.Add(-time.Hour)}}
	if mainAt(log, now) != "new" || mainAt(log, now.Add(-time.Second)) != "old" || mainAt(log, now.Add(-2*time.Hour)) != "" {
		t.Fatal("main time boundary")
	}
}
func TestRecordedMeterTimeCutoff(t *testing.T) {
	t.Parallel()
	r, write, git := fixture(t)
	first := "stage3/meter/runs/first/report.json"
	future := "stage3/meter/runs/future/report.json"
	write(first, `{"timestamp_utc":"20261008T080000Z","files":[{"file":"src/compiler/scanner.ts","source":true,"checker":false,"lowering":false}],"totals":{"source_files":1,"checker":0,"lowering":0}}`)
	write(future, `{"timestamp_utc":"20261008T100000Z","files":[{"file":"src/compiler/scanner.ts","source":true,"checker":true,"lowering":true}],"totals":{"source_files":1,"checker":1,"lowering":1}}`)
	git("add", ".")
	git("commit", "-qm", "meter")
	r.ref = "HEAD"
	r.paths = []string{first, future}
	r.at = time.Date(2026, 10, 8, 9, 0, 0, 0, time.UTC)
	track, e := stage3(r)
	if e != nil {
		t.Fatal(e)
	}
	if track.Measures[0].Done != 0 || track.Measures[1].Done != 0 {
		t.Fatal("future run read into past")
	}
}

func TestRecordedLandingHistoryCutoff(t *testing.T) {
	t.Parallel()
	r, write, git := fixture(t)
	write("documentation/velocity/landings.csv", "timestamp,landings\n2026-10-08T10:15:00Z,2\n2026-10-08T10:45:00Z,3\n2026-10-08T11:15:00Z,7\n")
	r = snapshotPaths(t, r, git)
	moment := time.Date(2026, 10, 8, 10, 30, 0, 0, time.UTC)
	values, err := historicalVelocity(r, moment)
	if err != nil {
		t.Fatal(err)
	}
	last := values[len(values)-1]
	if !last.Known || last.Done != 2 {
		t.Fatal("future landings credited", last)
	}
	values, err = historicalVelocity(r, moment.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	last = values[len(values)-1]
	if !last.Known || last.Done != 7 {
		t.Fatal("wrong hour", last)
	}
}
