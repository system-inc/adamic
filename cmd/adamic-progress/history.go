package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/system-inc/adamic/internal/stage1progress"
)

type observation struct {
	Time   time.Time `json:"time"`
	Commit string    `json:"main_commit"`
	Value  *float64  `json:"value"`
	Total  *float64  `json:"target"`
	Source string    `json:"source,omitempty"`
}
type trend struct {
	Now       *float64      `json:"now"`
	Hour      *float64      `json:"one_hour_ago"`
	SixHours  *float64      `json:"six_hours_ago"`
	Day       *float64      `json:"24_hours_ago"`
	Timeline  []observation `json:"hourly"`
	Sparkline string        `json:"sparkline"`
	Actual    *float64      `json:"actual_per_hour"`
	Needed    *float64      `json:"needed_per_hour"`
	Window    float64       `json:"rate_window_hours"`
	Status    string        `json:"status"`
	Note      string        `json:"note,omitempty"`
}
type landing struct {
	Commit string
	Time   time.Time
}

// First-parent snapshots date a merge by its main committer date, not the
// branch's author date. Git does not retain the time of a remote push.
func mainAt(log []landing, moment time.Time) string {
	for _, commit := range log {
		if !commit.Time.After(moment) {
			return commit.Commit
		}
	}
	return ""
}
func point(m metric, stamp time.Time, commit string) observation {
	p := observation{Time: stamp, Commit: commit, Source: m.Source}
	if m.Known {
		v := m.Done
		p.Value = &v
	}
	if m.Total > 0 {
		v := m.Total
		p.Total = &v
	}
	if strings.HasPrefix(m.Name, "overall") && m.Percent != nil {
		v := *m.Percent
		p.Value = &v
		target := 100.0
		p.Total = &target
	}
	if m.Name == "syntax-lint native/Go speed" {
		target := 1.0
		p.Total = &target
	}
	return p
}
func sparkline(points []observation) string {
	lo, hi := math.Inf(1), math.Inf(-1)
	for _, p := range points {
		if p.Value != nil {
			lo = math.Min(lo, *p.Value)
			hi = math.Max(hi, *p.Value)
		}
	}
	glyphs := []rune("▁▂▃▄▅▆▇█")
	var out strings.Builder
	for _, p := range points {
		if p.Value == nil {
			out.WriteRune('?')
			continue
		}
		n := 0
		if hi > lo {
			n = int(math.Round(7 * (*p.Value - lo) / (hi - lo)))
		}
		out.WriteRune(glyphs[n])
	}
	return out.String()
}
func calculate(points []observation, deadline, now time.Time) *trend {
	t := &trend{Timeline: points, Status: "unknown"}
	if len(points) != 25 {
		t.Note = "24-hour reconstruction unavailable"
		return t
	}
	t.Now = points[24].Value
	t.Hour = points[23].Value
	t.SixHours = points[18].Value
	t.Day = points[0].Value
	t.Sparkline = sparkline(points)
	// Use a trailing six-hour rate. If the measure was not recorded six hours
	// ago, use its oldest known hourly observation within that window.
	if t.Now != nil {
		for i := 18; i < 24; i++ {
			if points[i].Value == nil {
				continue
			}
			hours := now.Sub(points[i].Time).Hours()
			if hours > 0 {
				rate := (*t.Now - *points[i].Value) / hours
				t.Actual = &rate
				t.Window = hours
				break
			}
		}
	}
	current := points[24]
	if current.Value != nil && current.Total != nil {
		remaining := math.Max(0, *current.Total-*current.Value)
		hours := deadline.Sub(now).Hours()
		needed := 0.0
		if remaining > 0 {
			if hours > 0 {
				needed = remaining / hours
			} else {
				t.Note = "deadline passed with work remaining"
			}
		}
		if remaining == 0 || hours > 0 {
			t.Needed = &needed
		}
	}
	if t.Actual != nil && t.Needed != nil {
		t.Status = "red"
		if *t.Actual >= *t.Needed {
			t.Status = "green"
		}
	}
	if current.Total == nil {
		t.Note = "no deadline target for this observation"
	}
	return t
}
func snapshot(r repository, commit string, moment time.Time, lines func(string, string) (*stage1progress.Report, error)) ([]track, error) {
	r.ref = commit
	r.at = moment
	b, e := r.git("ls-tree", "-r", "--name-only", commit)
	if e != nil {
		return nil, e
	}
	r.paths = strings.Fields(string(b))
	s3, e := stage3(r)
	if e != nil {
		return nil, e
	}
	s1, e := stage1(r, lines)
	if e != nil {
		return nil, e
	}
	a, e := apple(r)
	if e != nil {
		return nil, e
	}
	tracks := []track{s3, s1, a}
	for i := range tracks {
		if e = finish(&tracks[i], r, moment); e != nil {
			return nil, e
		}
	}
	return tracks, nil
}
func reconstruct(r repository, d *dashboard, lines func(string, string) (*stage1progress.Report, error)) error {
	b, e := r.git("log", "--first-parent", "--format=%H %ct", d.Main)
	if e != nil {
		return e
	}
	var log []landing
	for _, row := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		parts := strings.Fields(row)
		if len(parts) != 2 {
			continue
		}
		seconds, e := strconv.ParseInt(parts[1], 10, 64)
		if e != nil {
			return e
		}
		log = append(log, landing{parts[0], time.Unix(seconds, 0)})
	}
	history := make([][]track, 25)
	velocityHistory := make([][]metric, 25)
	commits := make([]string, 25)
	for i := 0; i < 25; i++ {
		moment := d.Time.Add(time.Duration(i-24) * time.Hour)
		commit := mainAt(log, moment)
		commits[i] = commit
		if i == 24 {
			history[i] = d.Tracks
			velocityHistory[i] = velocityMetrics(d.Velocity, d.Time)
			commits[i] = d.Main
			continue
		}
		if commit == "" {
			continue
		}
		// Files are reused by the read cache, but a meter's timestamp cutoff is
		// evaluated for each moment, even when main's commit did not change.
		tracks, e := snapshot(r, commit, moment, lines)
		if e != nil {
			return fmt.Errorf("history %s main %.12s: %w", moment.Format(time.RFC3339), commit, e)
		}
		history[i] = tracks
		scoped := r
		scoped.ref = commit
		paths, e := r.git("ls-tree", "-r", "--name-only", commit)
		if e != nil {
			return e
		}
		scoped.paths = strings.Fields(string(paths))
		vm, e := historicalVelocity(scoped, moment)
		if e != nil {
			return e
		}
		velocityHistory[i] = vm
	}
	for ti := range d.Tracks {
		target := &d.Tracks[ti]
		for mi := -1; mi < len(target.Measures); mi++ {
			current := &target.Overall
			if mi >= 0 {
				current = &target.Measures[mi]
			}
			points := make([]observation, 25)
			for i := 0; i < 25; i++ {
				moment := d.Time.Add(time.Duration(i-24) * time.Hour)
				points[i] = observation{Time: moment, Commit: commits[i]}
				if history[i] == nil {
					continue
				}
				past := history[i][ti].Overall
				if mi >= 0 {
					found := false
					for _, m := range history[i][ti].Measures {
						if m.Name == current.Name {
							past = m
							found = true
							break
						}
					}
					if !found {
						continue
					}
				}
				points[i] = point(past, moment, commits[i])
			}
			current.Progress = calculate(points, target.Deadline, d.Time)
		}
		target.ETA = nil
		target.ETAScope = "extrapolated six-hour overall lower-bound rate from main history"
		p := target.Overall.Progress
		if p.Now != nil && *p.Now >= 100 {
			eta := d.Time
			target.ETA = &eta
		} else if p.Now != nil && p.Actual != nil && *p.Actual > 0 {
			hours := (100 - *p.Now) / *p.Actual
			if hours < float64(math.MaxInt64)/float64(time.Hour) {
				eta := d.Time.Add(time.Duration(hours * float64(time.Hour)))
				target.ETA = &eta
			}
		}
	}
	d.Velocity.Measures = velocityMetrics(d.Velocity, d.Time)
	for mi := range d.Velocity.Measures {
		m := &d.Velocity.Measures[mi]
		points := make([]observation, 25)
		for i := 0; i < 25; i++ {
			stamp := d.Time.Add(time.Duration(i-24) * time.Hour)
			points[i] = observation{Time: stamp, Commit: commits[i]}
			for _, past := range velocityHistory[i] {
				if past.Name == m.Name {
					points[i] = point(past, stamp, commits[i])
					break
				}
			}
		}
		m.Progress = calculate(points, time.Time{}, d.Time)
	}
	return nil
}
func value(v *float64) string {
	if v == nil {
		return "?"
	}
	return strconv.FormatFloat(*v, 'f', 3, 64)
}
func printTrend(w io.Writer, m metric, deadline time.Time) {
	p := m.Progress
	if p == nil {
		return
	}
	marker := "⚪"
	if p.Status == "green" {
		marker = "🟢"
	}
	if p.Status == "red" {
		marker = "🔴"
	}
	unit := ""
	if strings.HasPrefix(m.Name, "overall") {
		unit = "%"
	} else if m.Name == "syntax-lint native/Go speed" {
		unit = "x"
	}
	fmt.Fprintf(w, "  %s %s: doing %s%s/h, need %s%s/h", marker, m.Name, value(p.Actual), unit, value(p.Needed), unit)
	if !deadline.IsZero() {
		fmt.Fprintf(w, " to land by %s", deadline.In(mdt).Format("Jan 2 15:04 MST"))
	}
	if p.Actual != nil {
		fmt.Fprintf(w, " (last %.0fh)", p.Window)
	}
	if p.Note != "" {
		fmt.Fprintf(w, "; %s", p.Note)
	}
	fmt.Fprintln(w)
	fmt.Fprintf(w, "    now %s | 1h ago %s | 6h ago %s | 24h ago %s | %s\n", value(p.Now), value(p.Hour), value(p.SixHours), value(p.Day), p.Sparkline)
}
func renderHistory(w io.Writer, d dashboard) {
	fmt.Fprintln(w, "\nPer-hour measurements (oldest to newest; ? means no recorded observation)")
	for _, t := range d.Tracks {
		fmt.Fprintln(w, "\n"+t.Name)
		fmt.Fprint(w, "MDT time | main commit | overall %")
		for _, m := range t.Measures {
			fmt.Fprintf(w, " | %s", m.Name)
		}
		fmt.Fprintln(w)
		for i, p := range t.Overall.Progress.Timeline {
			fmt.Fprintf(w, "%s | %.12s | %s", p.Time.In(mdt).Format("Jan 2 15:04"), p.Commit, value(p.Value))
			for _, m := range t.Measures {
				fmt.Fprintf(w, " | %s", value(m.Progress.Timeline[i].Value))
			}
			fmt.Fprintln(w)
		}
	}
	fmt.Fprintln(w, "\nVelocity (past remote backlog requires archived refs/counts)")
	if len(d.Velocity.Measures) > 0 {
		fmt.Fprint(w, "MDT time")
		for _, m := range d.Velocity.Measures {
			fmt.Fprintf(w, " | %s", m.Name)
		}
		fmt.Fprintln(w)
		for i, p := range d.Velocity.Measures[0].Progress.Timeline {
			fmt.Fprint(w, p.Time.In(mdt).Format("Jan 2 15:04"))
			for _, m := range d.Velocity.Measures {
				fmt.Fprintf(w, " | %s", value(m.Progress.Timeline[i].Value))
			}
			fmt.Fprintln(w)
		}
	}

}

func velocityMetrics(v velocity, moment time.Time) []metric {
	result := []metric{measured("commits on main", float64(v.Total), 0, "git rev-list main"), measured("commits landed in last 24h", float64(v.Day), 0, "main commit timestamps"), measured("commits landed in last hour", float64(v.Hour), 0, "main commit timestamps"), measured("distinct remote backlog", float64(v.Backlog), 0, "fetched origin commit union minus main")}
	landings := unknown("recorded landings in current UTC hour", 0, "documentation/velocity/landings.csv", "no landings.csv on main yet")
	if v.Landings != nil {
		key := moment.UTC().Truncate(time.Hour).Format(time.RFC3339)
		landings = measured(landings.Name, float64(v.Landings[key]), 0, landings.Source)
	}
	return append(result, landings)
}
func historicalVelocity(r repository, moment time.Time) ([]metric, error) {
	count, e := r.git("rev-list", "--count", r.ref)
	if e != nil {
		return nil, e
	}
	total, e := strconv.Atoi(strings.TrimSpace(string(count)))
	if e != nil {
		return nil, e
	}
	b, e := r.git("log", "--format=%ct", r.ref)
	if e != nil {
		return nil, e
	}
	v := velocity{Total: total}
	for _, row := range strings.Fields(string(b)) {
		seconds, e := strconv.ParseInt(row, 10, 64)
		if e != nil {
			return nil, e
		}
		stamp := time.Unix(seconds, 0)
		if stamp.After(moment) {
			continue
		}
		if !stamp.Before(moment.Add(-24 * time.Hour)) {
			v.Day++
		}
		if !stamp.Before(moment.Add(-time.Hour)) {
			v.Hour++
		}
	}
	landingBytes, hasLandings, err := r.blob("documentation/velocity/landings.csv")
	if err != nil {
		return nil, err
	}
	if hasLandings {
		v.Landings, err = readLandings(landingBytes, moment)
		if err != nil {
			return nil, err
		}
	}
	metrics := velocityMetrics(v, moment)
	metrics[3] = unknown("distinct remote backlog", 0, "documentation/velocity/backlog.csv", "past remote refs not recorded; not measurable yet")
	b, ok, e := r.blob("documentation/velocity/backlog.csv")
	if e != nil {
		return nil, e
	}
	if ok {
		rows, e := csv.NewReader(strings.NewReader(string(b))).ReadAll()
		if e != nil {
			return nil, e
		}
		if len(rows) == 0 || len(rows[0]) != 2 || rows[0][0] != "timestamp" || rows[0][1] != "count" {
			return nil, errors.New("backlog.csv requires timestamp,count")
		}
		var latest time.Time
		for _, row := range rows[1:] {
			stamp, e := time.Parse(time.RFC3339, row[0])
			if e != nil {
				return nil, e
			}
			n, e := strconv.Atoi(row[1])
			if e != nil || n < 0 {
				return nil, errors.New("invalid backlog count")
			}
			if !stamp.After(moment) && stamp.After(latest) {
				latest = stamp
				metrics[3] = measured("distinct remote backlog", float64(n), 0, "documentation/velocity/backlog.csv")
			}
		}
	}
	return metrics, nil
}
