package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/stage1progress"
)

func TestBarsAgainstLiteralOracle(t *testing.T) {
	t.Parallel()
	for _, v := range []struct {
		p    float64
		want string
	}{{0, "[░░░░░░░░░░░░░░░░░░░░]"}, {25, "[█████░░░░░░░░░░░░░░░]"}, {75, "[███████████████░░░░░]"}, {100, "[████████████████████]"}} {
		if got := bar(&v.p); got != v.want {
			t.Fatalf("bar %.0f = %s, want %s", v.p, got, v.want)
		}
	}
	if bar(nil) != "[????????????????????]" {
		t.Fatal("unknown rendered as observed")
	}
}
func fixture(t *testing.T) (repository, func(string, string), func(...string) string) {
	t.Helper()
	root := t.TempDir()
	git := func(args ...string) string {
		t.Helper()
		c := exec.Command("git", append([]string{"-C", root}, args...)...)
		b, e := c.CombinedOutput()
		if e != nil {
			t.Fatalf("git %v: %v %s", args, e, b)
		}
		return strings.TrimSpace(string(b))
	}
	git("init", "-q")
	git("config", "user.name", "fixture")
	git("config", "user.email", "fixture@example.invalid")
	write := func(name, text string) {
		t.Helper()
		p := filepath.Join(root, name)
		if e := os.MkdirAll(filepath.Dir(p), 0755); e != nil {
			t.Fatal(e)
		}
		if e := os.WriteFile(p, []byte(text), 0644); e != nil {
			t.Fatal(e)
		}
	}
	return repository{root: root, ref: "origin/main", ctx: context.Background()}, write, git
}
func lines(string, string) (*stage1progress.Report, error) {
	return &stage1progress.Report{Total: 100, Ported: 25}, nil
}
func TestGitSnapshotMissingRecordsAndDistinctBacklog(t *testing.T) {
	t.Parallel()
	r, write, git := fixture(t)
	write("examples/apple/window.a", "window")
	write("stage1/cohere/lint/inventory/inventory.json", `{"rules":[{"name":"one","family":"core","variable":"One"},{"name":"two","family":"core","variable":"Two"}]}`)
	write("stage1/cohere/lint/testdata/oracle.go", `package main; import rules "github.com/system-inc/cohere/internal/lint/rules/core"; var subjects = []rule.Rule{rules.One,rules.One}`)
	git("add", ".")
	git("commit", "-qm", "main")
	git("update-ref", "refs/remotes/origin/main", "HEAD")
	write("worker", "pending")
	git("add", "worker")
	git("commit", "-qm", "worker")
	git("update-ref", "refs/remotes/origin/one", "HEAD")
	git("update-ref", "refs/remotes/origin/two", "HEAD")
	// Neither a dirty example nor a worker branch can earn main credit.
	write("examples/apple/counter.a", "dirty")
	now := time.Now().Add(time.Second)
	d, e := collect(r, now, lines)
	if e != nil {
		t.Fatal(e)
	}
	if d.Velocity.Total != 1 || d.Velocity.Backlog != 1 || d.Velocity.Day != 1 || d.Velocity.Hour != 1 {
		t.Fatalf("velocity %#v", d.Velocity)
	}
	if d.Tracks[0].Measures[0].Known || d.Tracks[0].Measures[0].Note != "no meter run on main yet" {
		t.Fatal("invented meter")
	}
	if d.Tracks[1].Measures[1].Done != 1 || d.Tracks[1].Measures[1].Total != 2 {
		t.Fatalf("lint %v", d.Tracks[1].Measures[1])
	}
	if *d.Tracks[2].Overall.Percent != 20 {
		t.Fatal("Apple snapshot/overall", d.Tracks[2].Overall)
	}
	if d.Tracks[0].Deadline.In(mdt).Format("Jan 2 15:04 MST") != "Oct 9 23:49 MDT" || d.Tracks[1].Deadline.In(mdt).Format("Jan 2 15:04 MST") != "Oct 10 23:49 MDT" {
		t.Fatal("deadlines shifted")
	}
	var out bytes.Buffer
	render(&out, d)
	if !strings.Contains(out.String(), "no landings.csv on main yet") || !strings.HasSuffix(out.String(), d.Slowest+"\n") {
		t.Fatal(out.String())
	}
	b, e := json.Marshal(d)
	if e != nil || !json.Valid(b) {
		t.Fatalf("json %v", e)
	}
}
func TestRecordedMeterAndTrackHistories(t *testing.T) {
	t.Parallel()
	r, write, git := fixture(t)
	write("stage3/meter/runs/first/report.json", `{"timestamp_utc":"20261008T080000Z","files":[{"file":"src/compiler/scanner.ts","source":true,"checker":false,"lowering":false}],"totals":{"source_files":1,"checker":0,"lowering":0}}`)
	write("stage3/meter/runs/latest/report.json", `{"timestamp_utc":"20261008T100000Z","files":[{"file":"src/compiler/scanner.ts","source":true,"checker":true,"lowering":false}],"totals":{"source_files":1,"checker":1,"lowering":0}}`)
	write("stage3/patch-set.md", "| Adaptation | Files | Lines added | Lines removed |\n| temporary | 1 | 20 | 10 |\n| **Total** | 1 | 20 | 10 |\n")
	for _, name := range []string{"stage3", "stage1", "apple"} {
		write("documentation/progress/"+name+".json", `[{"time":"2026-10-08T08:00:00Z","fraction":0.25},{"time":"2026-10-08T10:00:00Z","fraction":0.5}]`)
	}
	write("documentation/velocity/landings.csv", "timestamp_utc,landings\n2026-10-08T08:01:00Z,2\n2026-10-08T08:59:00Z,3\n")
	git("add", ".")
	git("commit", "-qm", "records")
	git("update-ref", "refs/remotes/origin/main", "HEAD")
	d, e := collect(r, time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC), lines)
	if e != nil {
		t.Fatal(e)
	}
	if d.Tracks[0].Measures[0].Done != 1 || d.Tracks[0].Measures[1].Done != 0 || d.Tracks[0].Measures[0].Total != 78 {
		t.Fatal("checker/lowering conflated", d.Tracks[0])
	}
	if d.Tracks[0].Measures[2].Done != 30 || d.Tracks[0].Measures[2].Percent != nil {
		t.Fatal("patch total/goal invented")
	}
	if d.Tracks[0].ETA == nil || d.Tracks[1].ETA != nil || d.Tracks[2].ETA != nil {
		t.Fatal("rate-derived ETA ignored or flat rate forecast", d.Tracks)
	}
	if d.Tracks[0].Measures[0].Progress.Hour == nil || *d.Tracks[0].Measures[0].Progress.Hour != 1 {
		t.Fatal("meter history cutoff ignored")
	}
	if d.Day != 4 || d.Velocity.Landings["2026-10-08T08:00:00Z"] != 5 {
		t.Fatalf("clock/landings %v", d)
	}
}
func TestMeterRefusesInconsistentRecord(t *testing.T) {
	t.Parallel()
	r, write, git := fixture(t)
	p := "stage3/meter/runs/x/report.json"
	write(p, `{"timestamp_utc":"20261008T100000Z","files":[{"file":"src/compiler/parser.ts","source":true,"checker":false,"lowering":true}],"totals":{"source_files":1,"checker":0,"lowering":1}}`)
	git("add", ".")
	git("commit", "-qm", "invalid")
	r.ref = "HEAD"
	r.paths = []string{p}
	track, e := stage3(r)
	found := false
	for _, m := range track.Measures {
		if strings.Contains(m.Note, "lowering without checker") {
			found = true
		}
	}
	if e != nil || track.Measures[0].Known || !found {
		t.Fatal("invalid meter observation hidden or credited", track, e)
	}
}
func TestHelpExplainsSources(t *testing.T) {
	t.Parallel()
	var out, err bytes.Buffer
	if run([]string{"--help"}, &out, &err) != 0 {
		t.Fatal(err.String())
	}
	for _, source := range []string{"stage3/meter", "patch-set.md", "inventory.json", "adamic-stage1-progress", "examples/apple", "landings.csv", "documentation/progress"} {
		if !strings.Contains(err.String(), source) {
			t.Fatal("unexplained source", source)
		}
	}
}

func TestLegacyMeterSchemaCannotClaimLowering(t *testing.T) {
	t.Parallel()
	r, write, git := fixture(t)
	p := "stage3/meter/runs/legacy/report.json"
	write(p, `{"timestamp_utc":"20261008T100000Z","root":"src/compiler","files_examined":78,"files_reaching_lowering":15,"reasons":[]}`)
	git("add", ".")
	git("commit", "-qm", "legacy")
	r.ref = "HEAD"
	r.paths = []string{p}
	track, e := stage3(r)
	if e != nil {
		t.Fatal(e)
	}
	if track.Measures[0].Done != 15 || track.Measures[1].Known {
		t.Fatal("legacy lowering invented", track)
	}
}
func TestForecastLatenessAndUnknown(t *testing.T) {
	t.Parallel()
	now := time.Now()
	p := 10.0
	q := 20.0
	a := now.Add(2 * time.Hour)
	b := now.Add(4 * time.Hour)
	tracks := []track{{Name: "Stage 3", Deadline: now, ETA: &a, Overall: metric{Percent: &p}}, {Name: "Stage 1", Deadline: now, ETA: &b, Overall: metric{Percent: &q}}}
	if !strings.Contains(slowest(tracks, now), "Stage 1; projected 4.0h behind") {
		t.Fatal(slowest(tracks, now))
	}
	tracks[1].ETA = nil
	if !strings.Contains(slowest(tracks, now), "behind deadline: unknown") {
		t.Fatal("unknown ETA suppressed")
	}
}
func TestMalformedLandingsAndBoundaries(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)
	for _, data := range []string{"", "when,count\nx,1\n", "timestamp,landings\n2026-10-08T08:00:00Z,-1\n", "timestamp\ninvalid\n"} {
		if _, e := readLandings([]byte(data), now); e == nil {
			t.Fatal("invalid CSV accepted", data)
		}
	}
	result, e := readLandings([]byte("timestamp\n2026-10-08T11:59:00Z\n2026-10-08T12:01:00Z\n"), now)
	if e != nil || len(result) != 1 || result["2026-10-08T11:00:00Z"] != 1 {
		t.Fatal(result, e)
	}
}

func TestRecordedParityAndSpeedScope(t *testing.T) {
	t.Parallel()
	r, write, git := fixture(t)
	write("stage3/progress.json", `{"milestones":{"scanner_native":true,"parser_native":false,"no_emit_match":false,"real_programs_match":false}}`)
	write("stage1/progress.json", `{"formatters":{"json":true,"css":false},"native_seconds":2,"go_seconds":2}`)
	write("stage1/cohere/lint/performance/1-measurements.json", `{"best":{"native":{"seconds":4},"Go":{"seconds":2}}}`)
	git("add", ".")
	git("commit", "-qm", "parity and speed")
	r.ref = "HEAD"
	tree, e := r.git("ls-tree", "-r", "--name-only", "HEAD")
	if e != nil {
		t.Fatal(e)
	}
	r.paths = strings.Fields(string(tree))
	s3, e := stage3(r)
	if e != nil {
		t.Fatal(e)
	}
	if s3.Measures[3].Done != 1 || s3.Measures[4].Done != 0 {
		t.Fatal("milestone evidence lost", s3)
	}
	s1, e := stage1(r, lines)
	if e != nil {
		t.Fatal(e)
	}
	if s1.Measures[2].Done != 1 || s1.Measures[9].Done != 0 {
		t.Fatal("parity or equal speed wrong", s1)
	}
	if e = finish(&s1, r, time.Now().Add(time.Second)); e != nil {
		t.Fatal(e)
	}
	if s1.Overall.Total != 10 || s1.Overall.Done != 1.25 {
		t.Fatal("scoped timing or missing inventory changed overall goals", s1.Overall)
	}
	last := s1.Measures[len(s1.Measures)-1]
	if last.Done != 0.5 || *last.Percent != 50 {
		t.Fatal("speed ratio inverted", last)
	}
}
