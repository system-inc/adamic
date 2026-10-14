package css

import (
	"cmp"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
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
			Flags:     []string{"go test -c " + strings.Join(buildcache.ReproducibleGoFlags(), " "), "./internal/format/css/" + pkg},
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
			// The overlay names this checkout's paths: it builds the product and is never part of it (#tqrqx60).
			defer os.Remove(overlayPath)
			// A product carries no deadline of its own: a cold Go build of cohere's css package can pass 60 s
			// on a compiler-changing candidate, and the build phase's 10-minute ceiling bounds it (rule 10).
			cmd := exec.CommandContext(t.Context(), "go", append(append([]string{"test"}, buildcache.ReproducibleGoFlags()...), "-c", "-overlay="+overlayPath, "-o", filepath.Join(dir, "oracle"), "./internal/format/css/"+pkg)...)
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

func TestProduct_CSSPrinterParserOracle(t *testing.T)        { t.Parallel(); cssOracleProduct(t, false) }
func TestProduct_CSSPrinterOracle(t *testing.T)              { t.Parallel(); cssOracleProduct(t, true) }
func TestProduct_CSSPrinterSanitizedAndLowered(t *testing.T) { t.Parallel(); cssPrinterProduct(t, -1) }
func TestProduct_CSSPrinterSemicolonMutant(t *testing.T)     { t.Parallel(); cssPrinterProduct(t, 0) }
func TestProduct_CSSPrinterIndentMutant(t *testing.T)        { t.Parallel(); cssPrinterProduct(t, 1) }
func TestProduct_CSSPrinterWidthMutant(t *testing.T)         { t.Parallel(); cssPrinterProduct(t, 2) }
func TestProduct_CSSPrinterDarwinLeaks(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "darwin" {
		t.Skip("macOS leaks requires an unsanitized executable")
	}
	cssPrinterProduct(t, 3)
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
	cssOracleRun(t, oracle, "TestAdamicPortCases", map[string]any{"Cases": cases, "Answers": answers, "Repository": repo, "Roots": cssCorpusRoots, "Fixtures": os.Getenv("ADAMIC_CSS_FIXTURES")})
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

// cssModePlan is the printer's shards: first the unmutated printer, each mode by cssPrinterPartitions content-hash
// partitions, then each mutant in each mode over exactly its witnesses, witnesses[mutant][mode] (cssPrinterWitnesses).
// A case's id names its check group ((mutant+1)*2 + mode) times the corpus size plus its line, so no two shards share
// an id.
func cssModePlan(lines []string, witnesses [][][]int) []cssModeShard {
	plan := cssPrinterUnmutatedPlan(lines)
	for variant := range witnesses {
		for modeIndex := range cssPrinterModes {
			plan = append(plan, cssPrinterMutantShard(lines, variant, modeIndex, witnesses[variant][modeIndex]))
		}
	}
	return plan
}

// cssPrinterUnmutatedPlan is the plan's first 128 shards, which read no witnesses.
func cssPrinterUnmutatedPlan(lines []string) []cssModeShard {
	var plan []cssModeShard
	for modeIndex, mode := range cssPrinterModes {
		for _, part := range cssPartition(lines, cssPrinterPartitions) {
			unit := cssModeShard{mode: mode, mutant: -1}
			for _, c := range part {
				unit.cases = append(unit.cases, cssCase{modeIndex*len(lines) + c.id, c.line})
			}
			plan = append(plan, unit)
		}
	}
	return plan
}

// cssPrinterWitnessesRun is how many of a mutant's witnesses its shard runs. The witnesses gather the corpus's big
// stylesheets, where a mutant differs: run over all of them the six shards' own work came to about 137 s a run on the
// Mac, one shard 43 s, where 64 apiece keeps each near a second. The product still holds and checks every witness
// (Loom's ruling B on #tpjban7).
const cssPrinterWitnessesRun = 64

// cssPrinterChosenWitnesses is the rule that picks them: the smallest witnesses by bytes, ties by case id, never by
// hand, back in corpus order. Ids outside the corpus are left for cssModeUnion to refuse.
func cssPrinterChosenWitnesses(lines []string, ids []int) []int {
	var chosen []int
	for _, id := range ids {
		if id >= 0 && id < len(lines) {
			chosen = append(chosen, id)
		}
	}
	slices.SortFunc(chosen, func(a, b int) int { return cmp.Or(cmp.Compare(len(lines[a]), len(lines[b])), cmp.Compare(a, b)) })
	chosen = chosen[:min(len(chosen), cssPrinterWitnessesRun)]
	slices.Sort(chosen)
	return chosen
}

func cssPrinterMutantShard(lines []string, variant, modeIndex int, ids []int) cssModeShard {
	unit := cssModeShard{mode: cssPrinterModes[modeIndex], mutant: variant}
	group := (variant+1)*2 + modeIndex
	for _, id := range cssPrinterChosenWitnesses(lines, ids) {
		unit.cases = append(unit.cases, cssCase{group*len(lines) + id, lines[id]})
	}
	return unit
}

// cssModeUnion holds the plan exact: every case in both modes exactly once across the unmutated shards, each by its
// hash owner, and each mutant shard in each mode exactly its chosen witnesses, from a witness set that isn't empty
// and lies inside the corpus.
func cssModeUnion(lines []string, plan []cssModeShard, witnesses [][][]int) error {
	if len(witnesses) != len(printerMutants) {
		return fmt.Errorf("witnesses for %d printer mutants, want %d", len(witnesses), len(printerMutants))
	}
	if len(plan) != testCSSPrinterAgreesWithGoShards {
		return fmt.Errorf("enumerated %d printer shards, declared %d", len(plan), testCSSPrinterAgreesWithGoShards)
	}
	var units [][]cssCase
	for index, unit := range plan[:2*cssPrinterPartitions] {
		group := index / cssPrinterPartitions
		if unit.mutant != -1 || unit.mode != cssPrinterModes[group] {
			return fmt.Errorf("incorrect mode/check group in shard %d", index)
		}
		for _, c := range unit.cases {
			if c.id/len(lines) != group || cssPrinterOwner(c.line, cssPrinterPartitions) != index%cssPrinterPartitions {
				return fmt.Errorf("incorrect case owner in shard %d", index)
			}
		}
		units = append(units, unit.cases)
	}
	if err := cssUnion(append(append([]string{}, lines...), lines...), units, 2*cssPrinterPartitions); err != nil {
		return err
	}
	for offset, unit := range plan[2*cssPrinterPartitions:] {
		index := 2*cssPrinterPartitions + offset
		variant, modeIndex := offset/2, offset%2
		if unit.mutant != variant || unit.mode != cssPrinterModes[modeIndex] {
			return fmt.Errorf("incorrect mutant/mode in shard %d", index)
		}
		ids := witnesses[variant][modeIndex]
		if err := cssPrinterWitnessError(printerMutants[variant].name, unit.mode, ids); err != nil {
			return err
		}
		for _, id := range ids {
			if id < 0 || id >= len(lines) {
				return fmt.Errorf("mutant %d's %s witness %d is not in the corpus of %d", variant, unit.mode, id, len(lines))
			}
		}
		ids = cssPrinterChosenWitnesses(lines, ids)
		group := (variant+1)*2 + modeIndex
		seen := map[int]bool{}
		for _, c := range unit.cases {
			id := c.id - group*len(lines)
			if c.id < 0 || c.id/len(lines) != group || seen[id] || lines[id] != c.line {
				return fmt.Errorf("incorrect, repeated or foreign case %d in mutant shard %d", c.id, index)
			}
			seen[id] = true
		}
		for _, id := range ids {
			if !seen[id] {
				return fmt.Errorf("mutant shard %d is missing chosen witness case %d", index, id)
			}
		}
		if len(seen) != len(ids) {
			return fmt.Errorf("mutant shard %d holds %d cases, its chosen witnesses are %d", index, len(seen), len(ids))
		}
	}
	return nil
}

// cssPrinterAllWitnesses reads every mutant's witness product.
func cssPrinterAllWitnesses(t *testing.T) [][][]int {
	t.Helper()
	var witnesses [][][]int
	for variant := range printerMutants {
		witnesses = append(witnesses, cssPrinterWitnesses(t, variant))
	}
	return witnesses
}

func TestCSSPrinterShardingCatchesDisagreement(t *testing.T) {
	t.Parallel()
	lines := cssPrinterCorpusLines(t)
	witnesses := cssPrinterAllWitnesses(t)
	plan := cssModePlan(lines, witnesses)
	if err := cssModeUnion(lines, plan, witnesses); err != nil {
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
	if len(caught) != 1 || caught[0] != fmt.Sprintf("shard-%03d", cssPrinterOwner(lines[0], cssPrinterPartitions)) {
		t.Fatalf("planted mode/case %d caught by %v, want exactly its hash owner", planted, caught)
	}
	t.Logf("planted disagreement mode/case %d caught exactly by %s", planted, caught[0])
	owner := cssPrinterOwner(lines[0], cssPrinterPartitions)
	mutantShard := 2 * cssPrinterPartitions
	foreign := 0
	for slices.Contains(witnesses[0][0], foreign) {
		foreign++
	}
	for _, broken := range []struct {
		name  string
		shard int
		edit  func([]cssCase) []cssCase
	}{
		{"repeated case", owner, func(cases []cssCase) []cssCase { return append(cases, cases[0]) }},
		{"missing case", owner, func(cases []cssCase) []cssCase { return cases[1:] }},
		{"repeated mutant case", mutantShard, func(cases []cssCase) []cssCase { return append(cases, cases[0]) }},
		{"mutant shard missing a witness", mutantShard, func(cases []cssCase) []cssCase { return cases[1:] }},
		{"mutant shard with a case that isn't a witness", mutantShard, func(cases []cssCase) []cssCase {
			return append(cases, cssCase{2*len(lines) + foreign, lines[foreign]})
		}},
	} {
		plan := cssModePlan(lines, witnesses)
		plan[broken.shard].cases = broken.edit(plan[broken.shard].cases)
		if cssModeUnion(lines, plan, witnesses) == nil {
			t.Fatalf("%s accepted", broken.name)
		}
	}
	empty := [][][]int{witnesses[0], witnesses[1], {{}, witnesses[2][1]}}
	if err := cssModeUnion(lines, cssModePlan(lines, empty), empty); err == nil || !strings.Contains(err.Error(), "differs from Go nowhere") {
		t.Fatalf("an empty witness set read as %v", err)
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
