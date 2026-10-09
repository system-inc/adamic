package printer

import (
	"crypto/sha256"
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

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

// Inputs describe non-Go read-only shared products; the Go oracle keys its own (printerOracle).
type printerBuildInputs struct {
	Name         string
	Files, Flags []string
	Toolchain    string
}

func printerBuild(t *testing.T, inputs printerBuildInputs, build func(dir string) error) string {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	key := buildcache.Inputs{Name: inputs.Name, Flags: append([]string{}, inputs.Flags...), Toolchain: []string{inputs.Toolchain, runtime.Version(), runtime.GOOS, runtime.GOARCH}}
	for _, file := range inputs.Files {
		relative, err := filepath.Rel(root, file)
		if err != nil {
			t.Fatal(err)
		}
		if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			// Generated inputs live in a read-only product outside the repository.
			// Its bytes, rather than the machine-specific cache path, identify it.
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			key.Flags = append(key.Flags, fmt.Sprintf("generated:%s:%x", filepath.Base(file), sha256.Sum256(data)))
		} else {
			key.Files = append(key.Files, filepath.ToSlash(relative))
		}
	}
	key.Flags = append(key.Flags, "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"))
	return buildcache.Product(t, key, build)
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

func printerLoweredProduct(t *testing.T, path string) string {
	t.Helper()
	return printerBuild(t, printerBuildInputs{
		Name: "lowered GraphQL printer", Files: printerInputFiles(t, repository, filepath.Dir(filepath.Dir(filepath.Dir(path)))), Toolchain: runtime.Version(),
	}, func(dir string) error {
		start := time.Now()
		program := lowered(t, path)
		t.Logf("lowering wall %.3fs", time.Since(start).Seconds())
		start = time.Now()
		source := native.C(program)
		t.Logf("C emission wall %.3fs", time.Since(start).Seconds())
		if err := os.WriteFile(dir+"/port.c", []byte(source), 0644); err != nil {
			return err
		}
		return os.WriteFile(dir+"/program.mjs", []byte(javascript.JavaScript(program)), 0644)
	})
}

func printerCompiledProduct(t *testing.T, loweredProduct string, options native.Options) string {
	t.Helper()
	data, err := os.ReadFile(loweredProduct + "/port.c")
	if err != nil {
		t.Fatal(err)
	}
	clang := buildcache.Tool("clang", "--version")
	name := "release GraphQL printer"
	if options.Sanitize {
		name = "sanitized GraphQL printer"
	}
	inputs := printerInputFiles(t, filepath.Join(repository, "internal/native"))
	return printerBuild(t, printerBuildInputs{Name: name, Files: append([]string{loweredProduct + "/port.c"}, inputs...), Flags: native.Flags(options), Toolchain: clang}, func(dir string) error {
		start := time.Now()
		err := native.Build(string(data), dir+"/port", options)
		t.Logf("clang wall %.3fs", time.Since(start).Seconds())
		return err
	}) + "/port"
}

type printerProducts struct{ source, backend, sanitized, release string }

func preparePrinterProducts(t *testing.T, path string) printerProducts {
	t.Helper()
	loweredProduct := printerLoweredProduct(t, path)
	sanitized := printerCompiledProduct(t, loweredProduct, native.Options{Sanitize: true})
	release := ""
	if runtime.GOOS == "darwin" {
		release = printerCompiledProduct(t, loweredProduct, native.Options{})
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

func TestPrinterShardUnionRejectsMissingAndRepeated(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
