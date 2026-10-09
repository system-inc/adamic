// adamic-bench validates checker outputs before publishing measurements.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const instrument = "Go monotonic time.Now + Linux wait4 ru_maxrss (KiB)"

type program struct {
	Name      string `json:"name"`
	Directory string `json:"directory"`
	EmitDir   string `json:"emit_dir"`
}
type contestant struct {
	Name           string   `json:"name"`
	Command        []string `json:"command"`
	VersionCommand []string `json:"version_command"`
	Pending        bool     `json:"pending"`
}
type manifest struct {
	Programs    []program    `json:"programs"`
	Contestants []contestant `json:"contestants"`
}
type machine struct {
	CPU      string `json:"cpu_model"`
	Governor string `json:"governor"`
	Cores    int    `json:"cores"`
	OS       string `json:"os"`
	Go       string `json:"go_version"`
}
type record struct {
	Schema       int               `json:"schema"`
	Instrument   string            `json:"instrument"`
	ManifestHash string            `json:"manifest_sha256"`
	Program      string            `json:"program"`
	InputHash    string            `json:"input_sha256"`
	Contestant   string            `json:"contestant"`
	Command      []string          `json:"command"`
	Version      string            `json:"version"`
	Pending      bool              `json:"pending"`
	Phase        string            `json:"phase"`
	Iteration    int               `json:"iteration"`
	Status       string            `json:"status"`
	ExitCode     int               `json:"exit_code"`
	Stdout       string            `json:"stdout"`
	Stderr       string            `json:"stderr"`
	Emitted      map[string]string `json:"emitted"`
	Wall         *float64          `json:"wall_ms"`
	RSS          *float64          `json:"peak_rss_kib"`
	Started      string            `json:"started_utc"`
	LoadBefore   float64           `json:"load_before"`
	LoadAfter    float64           `json:"load_after"`
	Machine      machine           `json:"machine"`
	Error        string            `json:"error"`
}
type options struct {
	warmup, runs int
	maxLoad      float64
	timeout      time.Duration
}
type refused struct{ reason string }

func (r refused) Error() string { return r.reason }
func hash(b []byte) string      { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }
func load() (float64, error) {
	b, e := os.ReadFile("/proc/loadavg")
	if e != nil {
		return 0, e
	}
	f := strings.Fields(string(b))
	if len(f) == 0 {
		return 0, errors.New("empty load")
	}
	return strconv.ParseFloat(f[0], 64)
}
func quiet(max float64, get func() (float64, error)) (float64, error) {
	n, e := get()
	if e != nil || math.IsNaN(n) || math.IsInf(n, 0) || n < 0 {
		return n, refused{"load unavailable"}
	}
	if n > max {
		return n, refused{fmt.Sprintf("load %.2f exceeds threshold %.2f", n, max)}
	}
	return n, nil
}
func host() machine {
	m := machine{CPU: "unavailable", Governor: "unavailable", Cores: runtime.NumCPU(), OS: runtime.GOOS, Go: runtime.Version()}
	b, _ := os.ReadFile("/proc/cpuinfo")
	for _, l := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(l, "model name") {
			m.CPU = strings.TrimSpace(strings.SplitN(l, ":", 2)[1])
			break
		}
	}
	files, _ := filepath.Glob("/sys/devices/system/cpu/cpu[0-9]*/cpufreq/scaling_governor")
	gs := map[string]bool{}
	for _, f := range files {
		b, e := os.ReadFile(f)
		if e == nil {
			gs[strings.TrimSpace(string(b))] = true
		}
	}
	if len(gs) > 0 {
		var names []string
		for g := range gs {
			names = append(names, g)
		}
		sort.Strings(names)
		m.Governor = strings.Join(names, ",")
	}
	return m
}

// snapshot rejects links and special files, and fingerprints paths as well as bytes.
func snapshot(root string) (map[string]string, error) {
	out := map[string]string{}
	e := filepath.WalkDir(root, func(p string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() {
			return nil
		}
		info, e := d.Info()
		if e != nil {
			return e
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("nonregular input/output: %s", p)
		}
		b, e := os.ReadFile(p)
		if e != nil {
			return e
		}
		r, e := filepath.Rel(root, p)
		if e != nil {
			return e
		}
		out[filepath.ToSlash(r)] = hash(b)
		return nil
	})
	return out, e
}
func fingerprint(s map[string]string) string { b, _ := json.Marshal(s); return hash(b) }
func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(p string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		r, e := filepath.Rel(src, p)
		if e != nil {
			return e
		}
		target := filepath.Join(dst, r)
		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		i, e := d.Info()
		if e != nil {
			return e
		}
		if !i.Mode().IsRegular() {
			return fmt.Errorf("nonregular input %s", p)
		}
		b, e := os.ReadFile(p)
		if e != nil {
			return e
		}
		return os.WriteFile(target, b, i.Mode().Perm())
	})
}

// Kill the process group on timeout, including compiler children.
func command(argv []string, dir string, timeout time.Duration) (string, string, int, float64, float64, error) {
	c := exec.Command(argv[0], argv[1:]...)
	c.Dir = dir
	c.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	var out, errout bytes.Buffer
	c.Stdout = &out
	c.Stderr = &errout
	start := time.Now()
	if e := c.Start(); e != nil {
		return "", "", -1, 0, 0, e
	}
	done := make(chan error, 1)
	go func() { done <- c.Wait() }()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	var e error
	select {
	case e = <-done:
	case <-timer.C:
		syscall.Kill(-c.Process.Pid, syscall.SIGKILL)
		<-done
		return out.String(), errout.String(), -1, 0, 0, errors.New("timeout: process group killed")
	}
	wall := float64(time.Since(start).Nanoseconds()) / 1e6
	if e != nil {
		var ex *exec.ExitError
		if !errors.As(e, &ex) {
			return out.String(), errout.String(), -1, 0, 0, e
		}
	}
	if c.ProcessState.ExitCode() < 0 {
		return out.String(), errout.String(), -1, 0, 0, errors.New("command terminated by signal")
	}
	u, ok := c.ProcessState.SysUsage().(*syscall.Rusage)
	if !ok || u.Maxrss <= 0 {
		return out.String(), errout.String(), c.ProcessState.ExitCode(), 0, 0, errors.New("RSS unavailable")
	}
	return out.String(), errout.String(), c.ProcessState.ExitCode(), wall, float64(u.Maxrss), nil
}
func invoke(p program, c contestant, o options, get func() (float64, error)) (record, error) {
	r := record{Schema: 1, Instrument: instrument, Program: p.Name, Contestant: c.Name, Command: c.Command, Pending: c.Pending, Status: "error", ExitCode: -1, Machine: host(), Emitted: map[string]string{}}
	n, e := quiet(o.maxLoad, get)
	if e != nil {
		return r, e
	}
	r.LoadBefore = n
	dir, e := os.MkdirTemp("", "adamic-bench-")
	if e != nil {
		return r, e
	}
	defer os.RemoveAll(dir)
	if e = copyTree(p.Directory, dir); e != nil {
		return r, e
	}
	before, e := snapshot(dir)
	if e != nil {
		return r, e
	}
	r.InputHash = fingerprint(before)
	r.Started = time.Now().UTC().Format(time.RFC3339Nano)
	stdout, stderr, code, wall, rss, execErr := command(c.Command, dir, o.timeout)
	r.Stdout = strings.ReplaceAll(stdout, dir, "<program>")
	r.Stderr = strings.ReplaceAll(stderr, dir, "<program>")
	r.ExitCode = code
	r.LoadAfter, e = get()
	if e != nil {
		r.LoadAfter = 0
		return r, e
	}
	if math.IsNaN(r.LoadAfter) || math.IsInf(r.LoadAfter, 0) || r.LoadAfter < 0 {
		r.LoadAfter = 0
		return r, errors.New("invalid load after")
	}
	if execErr != nil {
		return r, execErr
	}
	all, e := snapshot(dir)
	if e != nil {
		return r, e
	}
	prefix := filepath.ToSlash(p.EmitDir) + "/"
	for path, h := range all {
		if strings.HasPrefix(path, prefix) {
			r.Emitted[strings.TrimPrefix(path, prefix)] = h
			delete(all, path)
		}
	}
	if !reflect.DeepEqual(before, all) {
		return r, errors.New("command changed files outside emit_dir")
	}
	r.Wall = &wall
	r.RSS = &rss
	r.Status = "identical"
	if c.Pending {
		r.Status = "pending"
	}
	return r, nil
}
func equal(a, b record) bool {
	return a.InputHash == b.InputHash && a.ExitCode == b.ExitCode && a.Stdout == b.Stdout && a.Stderr == b.Stderr && reflect.DeepEqual(a.Emitted, b.Emitted)
}
func validate(records []record, oracle record) bool {
	bad := false
	failed := false
	for _, r := range records {
		if r.Status == "error" {
			failed = true
		}
		if r.Status != "error" && !equal(r, oracle) {
			bad = true
		}
	}
	if bad || failed {
		for i := range records {
			records[i].Wall = nil
			records[i].RSS = nil
			if bad {
				records[i].Status = "differs"
			} else {
				records[i].Status = "error"
			}
		}
	}
	return !bad && !failed
}
func quantile(v []float64, p float64) float64 {
	s := append([]float64(nil), v...)
	sort.Float64s(s)
	x := float64(len(s)-1) * p
	i := int(x)
	if i == len(s)-1 {
		return s[i]
	}
	return s[i] + (s[i+1]-s[i])*(x-float64(i))
}
func triple(v []float64) string {
	return fmt.Sprintf("%.2f / %.2f / %.2f", quantile(v, .5), quantile(v, .1), quantile(v, .9))
}
func valid(m manifest) error {
	if len(m.Programs) == 0 || len(m.Contestants) == 0 {
		return errors.New("programs and contestants required")
	}
	names := map[string]bool{}
	for _, p := range m.Programs {
		clean := filepath.Clean(p.EmitDir)
		if p.Name == "" || names[p.Name] || p.Directory == "" || clean == "." || filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
			return errors.New("invalid program declaration")
		}
		names[p.Name] = true
		if _, e := os.Lstat(filepath.Join(p.Directory, clean)); !os.IsNotExist(e) {
			return errors.New("emit_dir must be absent from input")
		}
		if _, e := snapshot(p.Directory); e != nil {
			return e
		}
	}
	names = map[string]bool{}
	for i, c := range m.Contestants {
		if c.Name == "" || names[c.Name] || len(c.Command) == 0 || c.Command[0] == "" || len(c.VersionCommand) == 0 || c.VersionCommand[0] == "" || (c.Name == "adamic-native" && !c.Pending) || (i == 0 && c.Pending) {
			return errors.New("invalid contestant declaration (native slot must be pending; oracle must be real)")
		}
		names[c.Name] = true
	}
	return nil
}
func benchmark(m manifest, o options, mhash string, jsonout, table io.Writer, get func() (float64, error)) (int, error) {
	if _, e := quiet(o.maxLoad, get); e != nil {
		return 2, e
	}
	versions := map[string]string{}
	for _, c := range m.Contestants {
		if _, e := quiet(o.maxLoad, get); e != nil {
			return 2, e
		}
		a, b, code, _, _, e := command(c.VersionCommand, "", o.timeout)
		if e != nil || code != 0 {
			return 1, fmt.Errorf("version command %s failed: %v (exit %d)", c.Name, e, code)
		}
		versions[c.Name] = strings.TrimSpace(a + b)
	}
	fmt.Fprintln(table, "PROGRAM  CONTESTANT  STATUS  WALL ms median / p10 / p90  RSS KiB median / p10 / p90  VS ORACLE")
	enc := json.NewEncoder(jsonout)
	exit := 0
	for _, p := range m.Programs {
		groups := make([][]record, len(m.Contestants))
		var oracle record
		for round := 0; round < o.warmup+o.runs; round++ {
			for k := 0; k < len(m.Contestants); k++ {
				i := (k + round) % len(m.Contestants)
				c := m.Contestants[i]
				r, e := invoke(p, c, o, get)
				if e != nil {
					var refusal refused
					if errors.As(e, &refusal) {
						for _, partial := range groups {
							for _, completed := range partial {
								completed.Status = "error"
								completed.Error = "incomplete program: " + e.Error()
								completed.Wall, completed.RSS = nil, nil
								if writeErr := enc.Encode(completed); writeErr != nil {
									return 1, writeErr
								}
							}
						}
						return 2, e
					}
					r.Error = e.Error()
					r.Status = "error"
				}
				r.ManifestHash = mhash
				r.Version = versions[c.Name]
				r.Iteration = round
				r.Phase = "warmup"
				if round >= o.warmup {
					r.Phase = "timed"
					r.Iteration = round - o.warmup
				}
				groups[i] = append(groups[i], r)
				if round == 0 && i == 0 {
					oracle = r
				}
			}
		}
		oracleOK := validate(groups[0], oracle)
		base := 0.0
		if oracleOK {
			var v []float64
			for _, r := range groups[0] {
				if r.Phase == "timed" {
					v = append(v, *r.Wall)
				}
			}
			base = quantile(v, .5)
		}
		for i, rs := range groups {
			ok := validate(rs, oracle)
			if !oracleOK {
				ok = false
				for j := range rs {
					rs[j].Status = "error"
					rs[j].Error = "oracle unstable or failed"
					rs[j].Wall = nil
					rs[j].RSS = nil
				}
			}
			for _, r := range rs {
				if e := enc.Encode(r); e != nil {
					return 1, e
				}
			}
			if !ok {
				exit = 1
				fmt.Fprintf(table, "%s  %s  %s  —  —  —\n", p.Name, m.Contestants[i].Name, rs[0].Status)
				continue
			}
			var wall, rss []float64
			for _, r := range rs {
				if r.Phase == "timed" {
					wall = append(wall, *r.Wall)
					rss = append(rss, *r.RSS)
				}
			}
			relation := "tie"
			median := quantile(wall, .5)
			if median < base {
				relation = "win"
			} else if median > base {
				relation = "loss"
			}
			if i == 0 {
				relation = "oracle"
			}
			if m.Contestants[i].Pending {
				relation += " (stand-in; pending)"
			}
			fmt.Fprintf(table, "%s  %s  %s  %s  %s  %s\n", p.Name, m.Contestants[i].Name, rs[0].Status, triple(wall), triple(rss), relation)
		}
	}
	return exit, nil
}
func run(args []string, out, errout io.Writer) int {
	f := flag.NewFlagSet("adamic-bench", flag.ContinueOnError)
	f.SetOutput(errout)
	path := f.String("manifest", "", "manifest JSON")
	dest := f.String("json", "", "JSONL output file")
	o := options{}
	f.IntVar(&o.warmup, "warmup", 1, "warmup repetitions")
	f.IntVar(&o.runs, "runs", 5, "timed repetitions")
	f.Float64Var(&o.maxLoad, "max-load", .5, "absolute one-minute load limit")
	f.DurationVar(&o.timeout, "timeout", 60*time.Second, "per-command deadline")
	if e := f.Parse(args); e != nil {
		return 2
	}
	fail := func(code int, e error) int { fmt.Fprintln(errout, "adamic-bench:", e); return code }
	if runtime.GOOS != "linux" || *path == "" || *dest == "" || f.NArg() != 0 || o.warmup < 0 || o.runs < 1 || o.timeout <= 0 || math.IsNaN(o.maxLoad) || math.IsInf(o.maxLoad, 0) || o.maxLoad < 0 {
		return fail(2, errors.New("invalid flags; Linux, --manifest and --json required"))
	}
	b, e := os.ReadFile(*path)
	if e != nil {
		return fail(2, e)
	}
	var m manifest
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if e = d.Decode(&m); e != nil {
		return fail(2, e)
	}
	var extra any
	if e = d.Decode(&extra); e != io.EOF {
		return fail(2, errors.New("trailing manifest data"))
	}
	for i := range m.Programs {
		p := &m.Programs[i]
		if !filepath.IsAbs(p.Directory) {
			p.Directory = filepath.Join(filepath.Dir(*path), p.Directory)
		}
		p.EmitDir = filepath.Clean(p.EmitDir)
	}
	if e = valid(m); e != nil {
		return fail(2, e)
	}
	if _, e = quiet(o.maxLoad, load); e != nil {
		return fail(2, e)
	}
	// O_EXCL prevents accidentally overwriting inputs or previous evidence.
	file, e := os.OpenFile(*dest, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if e != nil {
		return fail(2, e)
	}
	code, e := benchmark(m, o, hash(b), file, out, load)
	closeErr := file.Close()
	if e != nil {
		return fail(code, e)
	}
	if closeErr != nil {
		return fail(1, closeErr)
	}
	return code
}
func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }
