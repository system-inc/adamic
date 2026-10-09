package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Not parallel: fixture tests set a process-wide environment for child compilers.
// The test binary is a deterministic compiler stand-in, with planted differences.
func TestCompilerHelper(t *testing.T) {
	if os.Getenv("BENCH_HELPER") != "1" {
		return
	}
	if len(os.Args) > 2 && os.Args[len(os.Args)-1] == "version" {
		os.Stdout.WriteString("fixture 1\n")
		os.Exit(0)
	}
	mode := os.Args[len(os.Args)-1]
	if mode == "slow" {
		time.Sleep(50 * time.Millisecond)
	}
	if err := os.Mkdir("out", 0755); err != nil {
		os.Exit(9)
	}
	content := "same"
	if mode == "emit-diff" {
		content = "different"
	}
	if err := os.WriteFile("out/index.js", []byte(content), 0644); err != nil {
		os.Exit(9)
	}
	if mode == "diagnostic-diff" {
		os.Stdout.WriteString("planted diagnostic\n")
	}
	os.Exit(0)
}
func fixture(t *testing.T, mode string, pending bool) (manifest, options) {
	t.Helper()
	t.Setenv("BENCH_HELPER", "1")
	dir := t.TempDir()
	if e := os.WriteFile(filepath.Join(dir, "index.ts"), []byte("const x = 1"), 0644); e != nil {
		t.Fatal(e)
	}
	bin, e := os.Executable()
	if e != nil {
		t.Fatal(e)
	}
	cmd := func(m string) []string { return []string{bin, "-test.run=^TestCompilerHelper$", "--", m} }
	return manifest{Programs: []program{{Name: "small", Directory: dir, EmitDir: "out"}}, Contestants: []contestant{{Name: "oracle", Command: cmd("same"), VersionCommand: cmd("version")}, {Name: "candidate", Command: cmd(mode), VersionCommand: cmd("version"), Pending: pending}}}, options{warmup: 1, runs: 2, maxLoad: .5, timeout: 5 * time.Second}
}
func TestDifferenceWithholdsEveryTime(t *testing.T) {
	for _, mode := range []string{"emit-diff", "diagnostic-diff"} {
		t.Run(mode, func(t *testing.T) {
			m, o := fixture(t, mode, false)
			var records, table bytes.Buffer
			code, e := benchmark(m, o, "hash", &records, &table, func() (float64, error) { return .01, nil })
			if e != nil || code != 1 {
				t.Fatalf("code %d error %v", code, e)
			}
			dec := json.NewDecoder(&records)
			count := 0
			for dec.More() {
				var r record
				if e := dec.Decode(&r); e != nil {
					t.Fatal(e)
				}
				if r.Contestant == "candidate" {
					count++
					if r.Status != "differs" || r.Wall != nil || r.RSS != nil {
						t.Fatalf("time escaped: %+v", r)
					}
				}
			}
			if count != 3 || !strings.Contains(table.String(), "candidate  differs  —  —") {
				t.Fatal(table.String())
			}
		})
	}
}
func TestHighLoadRefusesBeforeCommands(t *testing.T) {
	m, o := fixture(t, "same", false)
	m.Contestants[0].VersionCommand = []string{"/does/not/exist"}
	var records, table bytes.Buffer
	code, e := benchmark(m, o, "hash", &records, &table, func() (float64, error) { return .51, nil })
	if code != 2 || e == nil || !strings.Contains(e.Error(), "exceeds") || records.Len() != 0 || table.Len() != 0 {
		t.Fatalf("%d %v %s", code, e, records.String())
	}
}
func TestPendingNeverPass(t *testing.T) {
	m, o := fixture(t, "same", true)
	var records, table bytes.Buffer
	code, e := benchmark(m, o, "hash", &records, &table, func() (float64, error) { return .01, nil })
	if e != nil || code != 0 {
		t.Fatalf("%d %v", code, e)
	}
	if !strings.Contains(table.String(), "(stand-in; pending)") || strings.Contains(table.String(), "pass") {
		t.Fatal(table.String())
	}
}
func TestLoadRisesBeforeInvocation(t *testing.T) {
	m, o := fixture(t, "same", false)
	n := 0
	var records, table bytes.Buffer
	code, e := benchmark(m, o, "hash", &records, &table, func() (float64, error) {
		n++
		if n >= 4 {
			return 1, nil
		}
		return 0, nil
	})
	if code != 2 || e == nil || records.Len() != 0 {
		t.Fatalf("%d %v", code, e)
	}
}
func TestQuantiles(t *testing.T) {
	v := []float64{30, 10, 20}
	for p, want := range map[float64]float64{.1: 12, .5: 20, .9: 28} {
		if got := quantile(v, p); got != want {
			t.Fatalf("got %g want %g", got, want)
		}
	}
}
func TestTimeoutKillsGroup(t *testing.T) {
	_, _, _, _, _, e := command([]string{"sh", "-c", "sleep 30 & wait"}, "", 30*time.Millisecond)
	if e == nil || !strings.Contains(e.Error(), "timeout") {
		t.Fatal(e)
	}
}
func TestRejectUnmarkedNative(t *testing.T) {
	m, _ := fixture(t, "same", false)
	m.Contestants[1].Name = "adamic-native"
	if valid(m) == nil {
		t.Fatal("accepted unmarked native stand-in")
	}
}

func TestLossIsNamed(t *testing.T) {
	m, o := fixture(t, "slow", false)
	var records, table bytes.Buffer
	code, e := benchmark(m, o, "hash", &records, &table, func() (float64, error) { return .01, nil })
	if e != nil || code != 0 || !strings.Contains(table.String(), "loss") {
		t.Fatalf("code=%d err=%v table=%s", code, e, table.String())
	}
}
func TestLateDifferenceRemovesEarlierTimes(t *testing.T) {
	wall, rss := 1.0, 2.0
	oracle := record{Status: "identical", Emitted: map[string]string{}, Wall: &wall, RSS: &rss}
	records := []record{oracle, oracle}
	records[1].Stdout = "late mutant"
	if validate(records, oracle) {
		t.Fatal("accepted mutant")
	}
	for _, r := range records {
		if r.Wall != nil || r.RSS != nil || r.Status != "differs" {
			t.Fatal("earlier time escaped")
		}
	}
}
