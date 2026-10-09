package css

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// These inputs sit beside each build callback for migration to internal/buildcache.
// That package is absent on this base: build once per test run, then share products.
// No package-local cache is used.
type cssBuildInputs struct {
	Name         string
	Files, Flags []string
	Toolchain    string
}

func cssCPU() float64 {
	var self, children syscall.Rusage
	_ = syscall.Getrusage(syscall.RUSAGE_SELF, &self)
	_ = syscall.Getrusage(syscall.RUSAGE_CHILDREN, &children)
	return float64(self.Utime.Sec+self.Stime.Sec+children.Utime.Sec+children.Stime.Sec) + float64(self.Utime.Usec+self.Stime.Usec+children.Utime.Usec+children.Stime.Usec)/1e6
}

func cssMeasurement(t *testing.T) func() {
	t.Helper()
	started, cpu := time.Now(), cssCPU()
	t.Cleanup(func() { t.Logf("total CPU (products and logic): %.3fs", cssCPU()-cpu) })
	return func() { t.Logf("setup including products: %s", time.Since(started)) }
}

func cssProduct(t *testing.T, inputs cssBuildInputs, build func(string) error) string {
	t.Helper()
	directory := t.TempDir()
	started, cpu := time.Now(), cssCPU()
	if err := build(directory); err != nil {
		t.Fatal(err)
	}
	t.Logf("build %s cold wall=%s CPU=%.3fs; toolchain=%s flags=%v inputs=%d", inputs.Name, time.Since(started), cssCPU()-cpu, inputs.Toolchain, inputs.Flags, len(inputs.Files))
	return directory
}

func cssSourceInputs(t *testing.T, directory string) []string {
	t.Helper()
	var files []string
	for _, slice := range []string{"css", "selector", "values", "mediaquery", "cssstrings", "cssnumbers"} {
		entries, err := os.ReadDir(filepath.Join(filepath.Dir(directory), slice))
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".ts") {
				files = append(files, filepath.Join(filepath.Dir(directory), slice, entry.Name()))
			}
		}
	}
	return files
}

type cssExecutable struct{ binary, javascript, source string }

func cssPrinterProduct(t *testing.T, path, name string, sanitize bool) cssExecutable {
	t.Helper()
	inputs := cssBuildInputs{Name: name, Files: cssSourceInputs(t, filepath.Dir(path)), Flags: native.Flags(native.Options{Sanitize: sanitize}), Toolchain: runtime.Version() + "; clang"}
	directory := cssProduct(t, inputs, func(dir string) error {
		checked, err := load.Load([]string{path})
		if err != nil {
			return err
		}
		program, err := lower.Lower(context.Background(), checked)
		if err != nil {
			return err
		}
		code := native.C(program)
		if err := os.WriteFile(filepath.Join(dir, "main.c"), []byte(code), 0644); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, "main.mjs"), []byte(javascript.JavaScript(program)), 0644); err != nil {
			return err
		}
		return native.Build(code, filepath.Join(dir, "native"), native.Options{Sanitize: sanitize})
	})
	return cssExecutable{filepath.Join(directory, "native"), filepath.Join(directory, "main.mjs"), path}
}

func cssOracleProduct(t *testing.T, printer bool) string {
	t.Helper()
	repo, _ := filepath.Abs(repository)
	side, pkg, test := "cohere_side_test.go", "postcss", "TestAdamicPortCases"
	target := "adamic_port_side_test.go"
	if printer {
		side, pkg, test, target = "print_side_test.go", "", "TestAdamicPrinterCases", "adamic_print_side_test.go"
	}
	sidePath, _ := filepath.Abs(filepath.Join("testdata", side))
	packagePath := filepath.Join(repo, "cohere", "internal", "format", "css", pkg)
	inputs := cssBuildInputs{Name: "Go " + test, Files: []string{sidePath, filepath.Join(repo, "cohere", "go.mod"), filepath.Join(repo, "cohere", "go.sum")}, Flags: []string{"test", "-c", "./internal/format/css/" + pkg}, Toolchain: runtime.Version()}
	directory := cssProduct(t, inputs, func(dir string) error {
		overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{filepath.Join(packagePath, target): sidePath}})
		if err != nil {
			return err
		}
		overlayPath := filepath.Join(dir, "overlay.json")
		if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
			return err
		}
		cmd := bounded(t, "go", "test", "-c", "-overlay="+overlayPath, "-o", filepath.Join(dir, "oracle"), "./internal/format/css/"+pkg)
		cmd.Dir = filepath.Join(repo, "cohere")
		output, err := childguard.CombinedOutput(cmd, childguard.Options{Stall: childStall})
		if err != nil {
			return fmt.Errorf("Go oracle build: %w\n%s", err, output)
		}
		return nil
	})
	return filepath.Join(directory, "oracle")
}

func cssOracleRun(t *testing.T, binary, test string, request any) {
	t.Helper()
	data, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "request.json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	cmd := bounded(t, binary, "-test.run=^"+test+"$", "-test.timeout=0", "-test.v")
	repo, _ := filepath.Abs(repository)
	cmd.Dir = filepath.Join(repo, "cohere", "internal", "format", "css")
	if test == "TestAdamicPortCases" {
		cmd.Dir = filepath.Join(cmd.Dir, "postcss")
	}
	cmd.Env = append(os.Environ(), "ADAMIC_PORT_REQUEST="+path)
	output, err := childguard.CombinedOutput(cmd, childguard.Options{Stall: childStall})
	if err != nil {
		t.Fatalf("oracle: %v %s", err, output)
	}
	t.Logf("%s", output)
}

func cssPrinterCases(t *testing.T, oracle string) string {
	t.Helper()
	directory := t.TempDir()
	cases := filepath.Join(directory, "cases.txt")
	answers := filepath.Join(directory, "answers.txt")
	repo, _ := filepath.Abs(repository)
	cssOracleRun(t, oracle, "TestAdamicPortCases", map[string]string{"Cases": cases, "Answers": answers, "Repository": repo, "Fixtures": os.Getenv("ADAMIC_CSS_FIXTURES")})
	return cases
}

func cssPrinterAnswers(t *testing.T, oracle, cases, mode string) string {
	t.Helper()
	answers := filepath.Join(t.TempDir(), "answers.txt")
	cssOracleRun(t, oracle, "TestAdamicPrinterCases", map[string]string{"Cases": cases, "Answers": answers, "Mode": mode})
	data, err := os.ReadFile(answers)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

type cssCase struct {
	id   int
	line string
}

func cssPartition(lines []string, count int) [][]cssCase {
	shards := make([][]cssCase, count)
	// Keep the existing mutant witness IDs in their named units. Assign the
	// remaining cases largest first to the least loaded shard. Stable ID ties
	// make the partition independent of map iteration and execution order.
	weights := make([]int, count)
	var pending []int
	for id, line := range lines {
		if id < 13 {
			shard := id % count
			shards[shard] = append(shards[shard], cssCase{id, line})
			weights[shard] += len(line)
		} else {
			pending = append(pending, id)
		}
	}
	sort.Slice(pending, func(i, j int) bool {
		a, b := pending[i], pending[j]
		if len(lines[a]) != len(lines[b]) {
			return len(lines[a]) > len(lines[b])
		}
		return a < b
	})
	for _, id := range pending {
		shard := 0
		for candidate := 1; candidate < count; candidate++ {
			if weights[candidate] < weights[shard] || (weights[candidate] == weights[shard] && len(shards[candidate]) < len(shards[shard])) {
				shard = candidate
			}
		}
		shards[shard] = append(shards[shard], cssCase{id, lines[id]})
		weights[shard] += len(lines[id])
	}
	for _, shard := range shards {
		sort.Slice(shard, func(i, j int) bool { return shard[i].id < shard[j].id })
	}
	return shards
}

func cssUnion(lines []string, shards [][]cssCase, count int) error {
	if len(shards) != count {
		return fmt.Errorf("enumerated %d shards, want %d", len(shards), count)
	}
	seen := make([]bool, len(lines))
	total := 0
	for _, shard := range shards {
		for _, c := range shard {
			if c.id < 0 || c.id >= len(lines) || seen[c.id] {
				return fmt.Errorf("missing or repeated case id %d", c.id)
			}
			if c.line != lines[c.id] {
				return fmt.Errorf("case id %d differs from unsplit enumeration", c.id)
			}
			seen[c.id] = true
			total++
		}
	}
	if total != len(lines) {
		return fmt.Errorf("union has %d cases, want %d", total, len(lines))
	}
	for id, found := range seen {
		if !found {
			return fmt.Errorf("missing case id %d", id)
		}
	}
	return nil
}

func cssShards(t *testing.T, cases string, count int) [][]cssCase {
	t.Helper()
	data, err := os.ReadFile(cases)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	for id, line := range lines {
		if len(line) < 2 || line[0] != '>' || (line[1] != 'C' && line[1] != 'S') {
			t.Fatalf("invalid corpus case %d", id)
		}
	}
	shards := cssPartition(lines, count)
	if err := cssUnion(lines, shards, count); err != nil {
		t.Fatal(err)
	}
	t.Logf("union: %d unique case ids, %d shards, unsplit SHA256=%x", len(lines), count, sha256.Sum256(data))
	return shards
}

func cssSelected(t *testing.T, count int) (int, int) {
	t.Helper()
	selection := os.Getenv("ADAMIC_TEST_SHARD")
	if selection == "" {
		return 0, count
	}
	parts := strings.Split(selection, "/")
	if len(parts) != 2 {
		t.Fatal("ADAMIC_TEST_SHARD must be i/n")
	}
	index, e1 := strconv.Atoi(parts[0])
	n, e2 := strconv.Atoi(parts[1])
	if e1 != nil || e2 != nil || n != count || index < 0 || index >= count {
		t.Fatalf("invalid ADAMIC_TEST_SHARD=%q; require 0 <= i < n=%d", selection, count)
	}
	return index, index + 1
}

func cssShardFile(t *testing.T, cases []cssCase) string {
	t.Helper()
	var b strings.Builder
	for _, c := range cases {
		b.WriteString(c.line)
		b.WriteByte('\n')
	}
	path := filepath.Join(t.TempDir(), "cases.txt")
	if err := os.WriteFile(path, []byte(b.String()), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func cssAgreement(result run, expected string) error {
	if result.exitCode != 0 || len(result.stderr) != 0 {
		return fmt.Errorf("exit %d stderr %s", result.exitCode, result.stderr)
	}
	if difference := firstDifference(string(result.stdout), expected); difference != "" {
		return fmt.Errorf("%s", difference)
	}
	return nil
}

type cssModeShard struct {
	mode   string
	mutant int
	cases  []cssCase
}

func cssModePlan(lines []string, count int) []cssModeShard {
	parts := cssPartition(lines, count/8)
	var plan []cssModeShard
	for variant := -1; variant < len(printerMutants); variant++ {
		for modeIndex, mode := range []string{"default", "narrow"} {
			for _, part := range parts {
				unit := cssModeShard{mode: mode, mutant: variant}
				for _, c := range part {
					unit.cases = append(unit.cases, cssCase{((variant+1)*2+modeIndex)*len(lines) + c.id, c.line})
				}
				plan = append(plan, unit)
			}
		}
	}
	return plan
}

func cssModeUnion(lines []string, plan []cssModeShard, count int) error {
	var units [][]cssCase
	for _, unit := range plan {
		units = append(units, unit.cases)
	}
	var enumeration []string
	for range 8 {
		enumeration = append(enumeration, lines...)
	}
	return cssUnion(enumeration, units, count)
}

func cssPrinterModeShards(t *testing.T, path string, count int) []cssModeShard {
	t.Helper()
	parts := cssShards(t, path, count/8)
	total := 0
	for _, part := range parts {
		total += len(part)
	}
	lines := make([]string, total)
	for _, part := range parts {
		for _, c := range part {
			lines[c.id] = c.line
		}
	}
	plan := cssModePlan(lines, count)
	if err := cssModeUnion(lines, plan, count); err != nil {
		t.Fatal(err)
	}
	t.Logf("check union: %d unique check-group/mode/case ids, %d case ids per group and mode, %d shards", total*8, total, count)
	return plan
}

func TestCSSPrinterShardingCatchesDisagreement(t *testing.T) {
	lines := make([]string, 256)
	for id := range lines {
		lines[id] = fmt.Sprintf(">Ccase-%d", id)
	}
	plan := cssModePlan(lines, testCSSPrinterAgreesWithGoShards)
	if err := cssModeUnion(lines, plan, testCSSPrinterAgreesWithGoShards); err != nil {
		t.Fatal(err)
	}
	planted := 0
	var caught []string
	for id, unit := range plan {
		var want, got strings.Builder
		for _, c := range unit.cases {
			fmt.Fprintf(&want, "case %d\n%s\n", c.id, c.line)
			line := c.line
			if c.id == planted {
				line = "planted disagreement"
			}
			fmt.Fprintf(&got, "case %d\n%s\n", c.id, line)
		}
		if cssAgreement(run{stdout: []byte(got.String())}, want.String()) != nil {
			caught = append(caught, fmt.Sprintf("shard-%03d", id))
		}
	}
	if len(caught) != 1 || caught[0] != "shard-000" {
		t.Fatalf("planted mode/case %d caught by %v, want exactly shard-000", planted, caught)
	}
	t.Logf("planted disagreement mode/case %d caught exactly by %s", planted, caught[0])
	broken := cssModePlan(lines, testCSSPrinterAgreesWithGoShards)
	broken[0].cases = append(broken[0].cases, broken[1].cases[0])
	if cssModeUnion(lines, broken, testCSSPrinterAgreesWithGoShards) == nil {
		t.Fatal("repeated case accepted")
	}
	broken = cssModePlan(lines, testCSSPrinterAgreesWithGoShards)
	broken[0].cases = broken[0].cases[1:]
	if cssModeUnion(lines, broken, testCSSPrinterAgreesWithGoShards) == nil {
		t.Fatal("missing case accepted")
	}
}
