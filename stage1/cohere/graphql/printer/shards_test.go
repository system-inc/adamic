package printer

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

// This local adapter prepares each product once per invocation until
// internal/buildcache is available on the base. No package cache is written.
type printerBuildInputs struct {
	Name         string
	Files, Flags []string
	Toolchain    string
}

func printerBuild(t *testing.T, inputs printerBuildInputs, build func(dir string) error) string {
	t.Helper()
	dir := t.TempDir()
	start := time.Now()
	if err := build(dir); err != nil {
		t.Fatalf("build %s: %v", inputs.Name, err)
	}
	t.Logf("build %s cold wall %.3fs (toolchain %s, %d files, flags %q)", inputs.Name, time.Since(start).Seconds(), inputs.Toolchain, len(inputs.Files), inputs.Flags)
	return dir
}

func printerInputFiles(t *testing.T, roots ...string) []string {
	t.Helper()
	seen := map[string]bool{}
	for _, root := range roots {
		absolute, err := filepath.Abs(root)
		if err != nil {
			t.Fatal(err)
		}
		err = filepath.WalkDir(absolute, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.Name() == ".git" || entry.Name() == "node_modules" {
				if entry.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if entry.Type().IsRegular() {
				seen[path] = true
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	files := make([]string, 0, len(seen))
	for path := range seen {
		files = append(files, path)
	}
	sort.Strings(files)
	return files
}

func printerOracle(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	side, _ := filepath.Abs("testdata/cohere_side_test.go")
	generator, _ := filepath.Abs("../testdata/cohere_side_test.go")
	cohere := filepath.Join(root, "cohere")
	return printerBuild(t, printerBuildInputs{
		Name:  "Go GraphQL printer oracle",
		Files: append([]string{side, generator}, printerInputFiles(t, cohere, root+"/go.mod", root+"/go.work")...),
		Flags: []string{"go test -c", "overlay: adamic_printer_test.go, adamic_generator_test.go"}, Toolchain: runtime.Version(),
	}, func(dir string) error {
		overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{
			cohere + "/internal/format/graphql/adamic_printer_test.go":   side,
			cohere + "/internal/format/graphql/adamic_generator_test.go": generator,
		}})
		if err != nil {
			return err
		}
		path := dir + "/overlay.json"
		if err := os.WriteFile(path, overlay, 0644); err != nil {
			return err
		}
		command := bounded(t, "go", "test", "-c", "-o="+dir+"/oracle", "-overlay="+path, "./internal/format/graphql")
		command.Dir = cohere
		output, err := combinedOutput(command)
		if err != nil {
			return fmt.Errorf("%w: %s", err, output)
		}
		return nil
	}) + "/oracle"
}

type printerProducts struct{ source, backend, sanitized, release string }

func preparePrinterProducts(t *testing.T, path string) printerProducts {
	t.Helper()
	var source string
	loweredProduct := printerBuild(t, printerBuildInputs{
		Name: "lowered GraphQL printer", Files: printerInputFiles(t, repository, filepath.Dir(filepath.Dir(filepath.Dir(path)))), Toolchain: runtime.Version(),
	}, func(dir string) error {
		program := lowered(t, path)
		source = native.C(program)
		if err := os.WriteFile(dir+"/port.c", []byte(source), 0644); err != nil {
			return err
		}
		return os.WriteFile(dir+"/program.mjs", []byte(javascript.JavaScript(program)), 0644)
	})
	clang := execute(t, nil, "clang", "--version")
	if clang.exitCode != 0 {
		t.Fatalf("clang version: %s", clang.stderr)
	}
	inputs := printerInputFiles(t, filepath.Join(repository, "internal/native"))
	sanitized := printerBuild(t, printerBuildInputs{Name: "sanitized GraphQL printer", Files: append([]string{loweredProduct + "/port.c"}, inputs...), Flags: native.Flags(native.Options{Sanitize: true}), Toolchain: string(clang.stdout)}, func(dir string) error {
		return native.Build(source, dir+"/port", native.Options{Sanitize: true})
	}) + "/port"
	release := ""
	if runtime.GOOS == "darwin" {
		release = printerBuild(t, printerBuildInputs{Name: "release GraphQL printer", Files: append([]string{loweredProduct + "/port.c"}, inputs...), Flags: native.Flags(native.Options{}), Toolchain: string(clang.stdout)}, func(dir string) error {
			return native.Build(source, dir+"/port", native.Options{})
		}) + "/port"
	}
	return printerProducts{source: path, backend: loweredProduct + "/program.mjs", sanitized: sanitized, release: release}
}

type printerCase struct{ id, input, want string }
type printerShard struct {
	mode, path string
	cases      []printerCase
}

func enumeratePrinter(t *testing.T, mode, path, want string) []printerCase {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(string(data), "\n") || !strings.HasSuffix(want, "\n") {
		t.Fatal("corpus protocol lacks final newline")
	}
	inputs := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	answers := strings.Split(strings.TrimSuffix(want, "\n"), "\n")
	if len(inputs) != len(answers) || len(inputs) == 0 {
		t.Fatalf("%s enumeration: %d inputs, %d answers", mode, len(inputs), len(answers))
	}
	result := make([]printerCase, len(inputs))
	for i := range inputs {
		if !strings.HasPrefix(inputs[i], ">") {
			t.Fatalf("%s case %d lacks input marker", mode, i)
		}
		result[i] = printerCase{id: fmt.Sprintf("%s/case-%06d", mode, i), input: inputs[i], want: answers[i]}
	}
	return result
}

// Validate both count and ID set, and the exact original input/answer pairing.
func printerShardUnion(whole []printerCase, shards []printerShard) error {
	expected := make(map[string]printerCase, len(whole))
	for _, item := range whole {
		if _, exists := expected[item.id]; exists {
			return fmt.Errorf("duplicate unsplit id %s", item.id)
		}
		expected[item.id] = item
	}
	seen := map[string]bool{}
	count := 0
	for number, shard := range shards {
		for _, item := range shard.cases {
			original, exists := expected[item.id]
			if !exists || seen[item.id] || item != original {
				return fmt.Errorf("shard-%03d: missing, changed or repeated id %s", number, item.id)
			}
			seen[item.id] = true
			count++
		}
	}
	if count != len(whole) || len(seen) != len(expected) {
		return fmt.Errorf("shard union %d ids, unsplit %d", count, len(whole))
	}
	return nil
}

func printerShardSelection(value string, count int) ([]bool, error) {
	index, boxes := 0, 1
	if value != "" {
		parts := strings.Split(value, "/")
		if len(parts) != 2 {
			return nil, fmt.Errorf("ADAMIC_TEST_SHARD=%q: want i/n", value)
		}
		var err error
		index, err = strconv.Atoi(parts[0])
		if err != nil {
			return nil, err
		}
		boxes, err = strconv.Atoi(parts[1])
		if err != nil || boxes <= 0 || index < 0 || index >= boxes {
			return nil, fmt.Errorf("ADAMIC_TEST_SHARD=%q: require 0 <= i < n", value)
		}
	}
	selected := make([]bool, count)
	for i := range selected {
		selected[i] = i%boxes == index
	}
	return selected, nil
}

func printerShardDisagreement(number int, shard printerShard, result run) error {
	if result.exitCode != 0 || len(result.stderr) != 0 {
		return fmt.Errorf("shard-%03d: exit %d, %s", number, result.exitCode, result.stderr)
	}
	var want strings.Builder
	for _, item := range shard.cases {
		want.WriteString(item.want)
		want.WriteByte('\n')
	}
	if string(result.stdout) == want.String() {
		return nil
	}
	rows := strings.Split(string(result.stdout), "\n")
	for i, item := range shard.cases {
		if i >= len(rows) || rows[i] != item.want {
			return fmt.Errorf("shard-%03d %s: %s", number, item.id, firstDifference(string(result.stdout), want.String()))
		}
	}
	return fmt.Errorf("shard-%03d: extra output or changed final newline", number)
}

const testPrinterAsGoCohereShards = 4

// The four fixed option modes own every case in their mode. ADAMIC_TEST_SHARD=i/n
// selects shard indices modulo n equal to i; unset runs all. The gate can instead
// select TestPrinterAsGoCohere/shard-NNN directly. Builds are shared setup inputs.
func TestPrinterAsGoCohere(t *testing.T) {
	oracle := printerOracle(t)
	path := printerDirectory(t, "", "", "")
	products := preparePrinterProducts(t, path)
	var whole []printerCase
	var shards []printerShard
	for _, mode := range []string{"defaults", "narrow", "tight", "tabs"} {
		cases, want := printerCases(t, mode, oracle)
		enumeration := enumeratePrinter(t, mode, cases, want)
		whole = append(whole, enumeration...)
		shards = append(shards, printerShard{mode: mode, path: cases, cases: enumeration})
	}
	if len(shards) != testPrinterAsGoCohereShards {
		t.Fatalf("enumerated %d shards, declared %d", len(shards), testPrinterAsGoCohereShards)
	}
	if err := printerShardUnion(whole, shards); err != nil {
		t.Fatal(err)
	}
	t.Logf("union: %d unique mode/case ids across %d shards", len(whole), len(shards))
	selected, err := printerShardSelection(os.Getenv("ADAMIC_TEST_SHARD"), len(shards))
	if err != nil {
		t.Fatal(err)
	}
	for number, shard := range shards {
		if !selected[number] {
			continue
		}
		t.Run(fmt.Sprintf("shard-%03d", number), func(t *testing.T) {
			t.Parallel()
			start := time.Now()
			t.Cleanup(func() {
				if elapsed := time.Since(start); elapsed > 30*time.Second {
					t.Errorf("invalid test unit: %.3fs exceeds 30s", elapsed.Seconds())
				}
			})
			t.Logf("mode %s, cases 0..%d", shard.mode, len(shard.cases)-1)
			check := func(name string, result run) {
				if err := printerShardDisagreement(number, shard, result); err != nil {
					t.Errorf("%s: %v", name, err)
				}
			}
			args := []string{"--cases", shard.path, shard.mode}
			check("Node", onNode(t, products.source, args...))
			check("native", execute(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, products.sanitized, args...))
			check("JS backend", onNode(t, products.backend, args...))
			switch runtime.GOOS {
			case "linux":
				check("leaks", execute(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, products.sanitized, args...))
			case "darwin":
				report := execute(t, nil, "leaks", append([]string{"--atExit", "--", products.release}, args...)...)
				if report.exitCode != 0 {
					t.Errorf("leaks: exit %d stdout %s stderr %s", report.exitCode, report.stdout, report.stderr)
				}
			default:
				t.Fatalf("no leak check for %s", runtime.GOOS)
			}
		})
	}
}

func TestPrinterShardUnionRejectsMissingAndRepeated(t *testing.T) {
	whole := []printerCase{{id: "defaults/case-000000", input: ">query{a}", want: "ok"}, {id: "tabs/case-000000", input: ">query{b}", want: "ok"}}
	shards := []printerShard{{cases: whole[:1]}, {cases: whole[1:]}}
	if err := printerShardUnion(whole, shards); err != nil {
		t.Fatal(err)
	}
	for _, invalid := range [][]printerShard{{{cases: whole[:1]}}, {{cases: []printerCase{whole[0], whole[0]}}}, {{cases: []printerCase{{id: whole[0].id, input: "changed", want: "ok"}, whole[1]}}}} {
		if err := printerShardUnion(whole, invalid); err == nil {
			t.Fatal("invalid union accepted")
		}
	}
}

func TestPrinterShardSelection(t *testing.T) {
	for _, boxes := range []int{1, 2, 7, 100} {
		seen := make([]int, testPrinterAsGoCohereShards)
		for box := 0; box < boxes; box++ {
			selected, err := printerShardSelection(fmt.Sprintf("%d/%d", box, boxes), len(seen))
			if err != nil {
				t.Fatal(err)
			}
			for i, runs := range selected {
				if runs {
					seen[i]++
				}
			}
		}
		for i, count := range seen {
			if count != 1 {
				t.Fatalf("shard-%03d assigned %d times", i, count)
			}
		}
	}
	for _, invalid := range []string{"1", "x/2", "0/0", "-1/2", "2/2", "0/x"} {
		if _, err := printerShardSelection(invalid, 4); err == nil {
			t.Fatalf("invalid selector accepted: %s", invalid)
		}
	}
	selected, err := printerShardSelection("", 4)
	if err != nil {
		t.Fatal(err)
	}
	for i, runs := range selected {
		if !runs {
			t.Fatalf("unset selector omitted shard-%03d", i)
		}
	}
}

func TestPrinterShardPlantedDisagreement(t *testing.T) {
	oracle := printerOracle(t)
	path, _ := filepath.Abs("main.ts")
	var shards []printerShard
	for _, mode := range []string{"defaults", "tabs"} {
		cases, want := printerCases(t, mode, oracle)
		enumeration := enumeratePrinter(t, mode, cases, want)
		// Use the first two enumerated oracle cases as a real Node control.
		sample := append([]printerCase(nil), enumeration[:2]...)
		filename := filepath.Join(t.TempDir(), "cases.txt")
		if err := os.WriteFile(filename, []byte(sample[0].input+"\n"+sample[1].input+"\n"), 0644); err != nil {
			t.Fatal(err)
		}
		shards = append(shards, printerShard{mode: mode, path: filename, cases: sample})
	}
	results := make([]run, len(shards))
	for i, shard := range shards {
		results[i] = onNode(t, path, "--cases", shard.path, shard.mode)
		if err := printerShardDisagreement(i, shard, results[i]); err != nil {
			t.Fatal(err)
		}
	}
	shards[0].cases[1].want += "planted disagreement"
	caught := 0
	for i, shard := range shards {
		if err := printerShardDisagreement(i, shard, results[i]); err != nil {
			caught++
			if i != 0 || !strings.Contains(err.Error(), "shard-000 defaults/case-000001") {
				t.Fatalf("wrong owner: %v", err)
			}
			t.Logf("planted disagreement caught: %v", err)
		}
	}
	if caught != 1 {
		t.Fatalf("planted disagreement caught by %d shards, want exactly one", caught)
	}
}
