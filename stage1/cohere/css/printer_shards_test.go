package css

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Each declaration and leaf calls the same recipe through a process-local Once.
// Cached source trees live with their product, never in a declaring test's TempDir.
type cssExecutable struct{ binary, javascript, source string }

var cssOracleOnce [2]sync.Once
var cssOracles [2]string
var cssExecutableOnce [5]sync.Once
var cssExecutables [5]cssExecutable

func cssOracleProduct(t *testing.T, printer bool) string {
	t.Helper()
	index := 0
	if printer {
		index = 1
	}
	cssOracleOnce[index].Do(func() {
		repo, err := filepath.Abs(repository)
		if err != nil {
			t.Fatal(err)
		}
		side, pkg, target := "cohere_side_test.go", "postcss", "adamic_port_side_test.go"
		if printer {
			side, pkg, target = "print_side_test.go", "", "adamic_print_side_test.go"
		}
		sidePath := filepath.Join(repo, "stage1/cohere/css/testdata", side)
		inputs := buildcache.Inputs{
			Name:      "css-printer-oracle-" + strconv.Itoa(index),
			Files:     []string{"stage1/cohere/css/testdata/" + side, "cohere/internal", "cohere/TypeScript/tsc", "cohere/TypeScript-shim", "cohere/go.mod", "cohere/go.sum", "cohere/TypeScript/tsc/go.mod", "cohere/TypeScript/tsc/go.sum"},
			Flags:     []string{"test", "-c", "./internal/format/css/" + pkg},
			Toolchain: []string{buildcache.Tool("go", "version"), runtime.GOOS, runtime.GOARCH},
		}
		directory := buildcache.Product(t, inputs, func(dir string) error {
			overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{filepath.Join(repo, "cohere/internal/format/css", pkg, target): sidePath}})
			if err != nil {
				return err
			}
			overlayPath := filepath.Join(dir, "overlay.json")
			if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
				return err
			}
			// A product carries no deadline of its own: a cold Go build of cohere's css package can pass 60 s
			// on a compiler-changing candidate, and the build phase's 10-minute ceiling bounds it (rule 10).
			cmd := exec.CommandContext(t.Context(), "go", "test", "-c", "-overlay="+overlayPath, "-o", filepath.Join(dir, "oracle"), "./internal/format/css/"+pkg)
			cmd.Dir = filepath.Join(repo, "cohere")
			output, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("Go oracle build: %w\n%s", err, output)
			}
			return nil
		})
		cssOracles[index] = filepath.Join(directory, "oracle")
	})
	// A build that failed inside the Once leaves the path empty for every later caller in this process.
	// Say so, rather than running an empty command.
	if cssOracles[index] == "" {
		t.Fatalf("css oracle product %d unavailable: its build failed earlier in this process (see that test's log)", index)
	}
	return cssOracles[index]
}

// variant -1 is the regular sanitized printer; 0..2 are its mutants;
// variant 3 is the unsanitized printer used by macOS's leaks tool.
func cssPrinterProduct(t *testing.T, variant int) cssExecutable {
	t.Helper()
	index := variant + 1
	cssExecutableOnce[index].Do(func() {
		sanitize := variant != 3
		inputs := buildcache.Inputs{
			Name:      "css-printer-executable-" + strconv.Itoa(variant),
			Files:     []string{"stage1/cohere/css", "stage1/cohere/selector", "stage1/cohere/values", "stage1/cohere/mediaquery", "stage1/cohere/cssstrings", "stage1/cohere/cssnumbers", "internal", "bridge/tsgo", "cohere/policy", "cohere/internal", "cohere/rule_runner", "cohere/go.mod", "cohere/go.sum", "cohere/TypeScript/tsc", "cohere/TypeScript-shim", "go.mod", "go.work"},
			Flags:     append(native.Flags(native.Options{Sanitize: sanitize}), "variant="+strconv.Itoa(variant), "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT")),
			Toolchain: []string{runtime.Version(), buildcache.Tool("clang", "--version"), runtime.GOOS, runtime.GOARCH},
		}
		directory := buildcache.Product(t, inputs, func(dir string) error {
			var mutation *mutant
			if variant >= 0 && variant < len(printerMutants) {
				mutation = &printerMutants[variant]
			}
			original := portDirectory(t, mutation)
			if err := moveTree(filepath.Dir(original), filepath.Join(dir, "sources")); err != nil {
				return err
			}
			source := filepath.Join(dir, "sources/css/print_main.ts")
			checked, err := load.Load([]string{source})
			if err != nil {
				return err
			}
			program, err := lower.Lower(t.Context(), checked)
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
		cssExecutables[index] = cssExecutable{filepath.Join(directory, "native"), filepath.Join(directory, "main.mjs"), filepath.Join(directory, "sources/css/print_main.ts")}
	})
	if cssExecutables[index].binary == "" {
		t.Fatalf("css printer product %d unavailable: its build failed earlier in this process (see that test's log)", variant)
	}
	return cssExecutables[index]
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

// Ownership depends only on the case bytes, not corpus order or other cases.
func cssPrinterOwner(line string, count int) int {
	sum := sha256.Sum256([]byte(line))
	return int(binary.LittleEndian.Uint64(sum[:8]) % uint64(count))
}

func cssPartition(lines []string, count int) [][]cssCase {
	shards := make([][]cssCase, count)
	for id, line := range lines {
		owner := cssPrinterOwner(line, count)
		shards[owner] = append(shards[owner], cssCase{id, line})
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
	if count%8 != 0 || count == 0 {
		return fmt.Errorf("invalid mode shard count %d", count)
	}
	for index, unit := range plan {
		group := index / (count / 8)
		if unit.mutant != group/2-1 || unit.mode != []string{"default", "narrow"}[group%2] {
			return fmt.Errorf("incorrect mode/check group in shard %d", index)
		}
		for _, c := range unit.cases {
			if c.id/len(lines) != group || cssPrinterOwner(c.line, count/8) != index%(count/8) {
				return fmt.Errorf("incorrect case owner in shard %d", index)
			}
		}
	}
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
	t.Parallel()
	lines := cssPrinterCorpusLines(t)
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
	if len(caught) != 1 || caught[0] != fmt.Sprintf("shard-%03d", cssPrinterOwner(lines[0], testCSSPrinterAgreesWithGoShards/8)) {
		t.Fatalf("planted mode/case %d caught by %v, want exactly its hash owner", planted, caught)
	}
	t.Logf("planted disagreement mode/case %d caught exactly by %s", planted, caught[0])
	broken := cssModePlan(lines, testCSSPrinterAgreesWithGoShards)
	owner := cssPrinterOwner(lines[0], testCSSPrinterAgreesWithGoShards/8)
	broken[owner].cases = append(broken[owner].cases, broken[owner].cases[0])
	if cssModeUnion(lines, broken, testCSSPrinterAgreesWithGoShards) == nil {
		t.Fatal("repeated case accepted")
	}
	broken = cssModePlan(lines, testCSSPrinterAgreesWithGoShards)
	broken[owner].cases = broken[owner].cases[1:]
	if cssModeUnion(lines, broken, testCSSPrinterAgreesWithGoShards) == nil {
		t.Fatal("missing case accepted")
	}
}

// moveTree moves a directory into a product. The port's copy lives under t.TempDir() and the product
// under the build cache, which can be different filesystems on a pool instance, where os.Rename fails
// with "invalid cross-device link". There it copies the tree instead, files, directories and symbolic
// links alike.
func moveTree(source, destination string) error {
	if err := os.Rename(source, destination); err == nil {
		return nil
	}
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)
		info, err := entry.Info()
		if err != nil {
			return err
		}
		switch {
		case entry.IsDir():
			return os.MkdirAll(target, info.Mode().Perm())
		case entry.Type()&fs.ModeSymlink != 0:
			link, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(link, target)
		default:
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			return os.WriteFile(target, data, info.Mode().Perm())
		}
	})
}
