// Package oracles owns the compiler-free CSS oracle products and shared test inputs.
// Relative paths resolve from stage1/cohere/css: the parent tests already run there,
// and this package's TestMain changes its working directory to that parent directory.
package oracles

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/toolchain"
)

// CompilerConfiguration records the C and C++ compilers used by Go and the C compiler's report.
type CompilerConfiguration struct{ CCompiler, CPlusPlusCompiler, Report string }

// GoCompilerConfiguration reads the compilers toolchain.Env gives go: the caller's CC and CXX, or clang and clang++ when it
// names neither.
func GoCompilerConfiguration() CompilerConfiguration {
	var compilers CompilerConfiguration
	for _, setting := range toolchain.Env() {
		switch name, value, _ := strings.Cut(setting, "="); name {
		case "CC":
			compilers.CCompiler = value
		case "CXX":
			compilers.CPlusPlusCompiler = value
		}
	}
	compilers.Report = "no C compiler"
	if fields := strings.Fields(compilers.CCompiler); len(fields) > 0 {
		compilers.Report = buildcache.Tool(fields[0], "--version")
	}
	return compilers
}

// The overlay source and the complete Go module trees address each oracle by
// content. Corpus enumeration stays live: only its binary is cached. The build
// runs go under toolchain.Env, so the key names the C and
// C++ compilers that environment gives go, and the C compiler by its report.
// CompositionOracle builds or fetches a Go oracle with the supplied overlay.
func CompositionOracle(t *testing.T, name, sidePath, overlayTarget, packagePath string) string {
	t.Helper()
	repo, _ := filepath.Abs(repository)
	side, _ := filepath.Abs(sidePath)
	sideRelative, err := filepath.Rel(repo, side)
	if err != nil {
		t.Fatal(err)
	}
	inputs := CompositionOracleInputs(name, sideRelative, overlayTarget, packagePath, GoCompilerConfiguration())
	directory := oracleProduct(t, inputs, func(dir string) error {
		overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{filepath.Join(repo, "cohere", overlayTarget): side}})
		if err != nil {
			return err
		}
		overlayPath := filepath.Join(dir, "overlay.json")
		if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
			return err
		}
		// The overlay names this checkout's paths: it builds the product and is never part of it (#tqrqx60).
		defer os.Remove(overlayPath)
		command := toolchain.CommandContextIn(t.Context(), filepath.Join(repo, "cohere"), "go", append(append([]string{"test"}, buildcache.ReproducibleGoFlags()...), "-c", "-overlay="+overlayPath, "-o", filepath.Join(dir, "oracle"), packagePath)...)
		command.WaitDelay = time.Second
		output, err := childguard.CombinedOutput(command, childguard.Options{Unbounded: true})
		if err != nil {
			return fmt.Errorf("Go oracle: %w\n%s", err, output)
		}
		return nil
	})
	return filepath.Join(directory, "oracle")
}

// CompositionOracleInputs describes a composition oracle without building it.
func CompositionOracleInputs(name, sideRelative, overlayTarget, packagePath string, compilers CompilerConfiguration) buildcache.Inputs {
	return buildcache.Inputs{
		Name:      name,
		Files:     []string{"cohere", "go.mod", "go.work", sideRelative},
		Flags:     []string{"go test -c " + strings.Join(buildcache.ReproducibleGoFlags(), " "), packagePath, "overlay=" + overlayTarget, "GOFLAGS=" + os.Getenv("GOFLAGS"), "GOEXPERIMENT=" + os.Getenv("GOEXPERIMENT"), "CGO_ENABLED=" + os.Getenv("CGO_ENABLED"), "GOOS=" + os.Getenv("GOOS"), "GOARCH=" + os.Getenv("GOARCH"), "CC=" + compilers.CCompiler, "CXX=" + compilers.CPlusPlusCompiler},
		Toolchain: []string{buildcache.Tool("go", "version"), runtime.GOOS, runtime.GOARCH, compilers.Report},
	}
}

const repository = "../../.."

// ParserShardCount is the fixed number of CSS parser agreement shards.
const ParserShardCount = 388

// Loaded formatter output gaps reached 66.19 seconds; four minutes gives over 3x headroom.
// ChildStallLimit bounds the time a CSS child may stop making progress.
const ChildStallLimit = 4 * time.Minute

// ParserCorpus holds the live inputs, Go answers, and stable case keys.
type ParserCorpus struct{ Inputs, Answers, CaseKeys []string }

func readCSSParserCorpus(t *testing.T, path, answers string) ParserCorpus {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	inputs := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	lines := strings.Split(strings.TrimSuffix(answers, "\n"), "\n")
	if len(lines) != len(inputs)*2 {
		t.Fatalf("oracle has %d lines for %d cases", len(lines), len(inputs))
	}
	corpus := ParserCorpus{Inputs: inputs}
	manifest, err := os.ReadFile(filepath.Join(filepath.Dir(path), "ids.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(manifest, &corpus.CaseKeys); err != nil {
		t.Fatal(err)
	}
	if len(corpus.CaseKeys) != len(inputs) {
		t.Fatal("stable IDs differ from live enumeration")
	}
	for i := range inputs {
		if lines[i*2] != fmt.Sprintf("case %d", i) {
			t.Fatalf("missing or repeated unsplit case ID at %d: %q", i, lines[i*2])
		}
		corpus.Answers = append(corpus.Answers, lines[i*2+1])
	}
	return corpus
}

// Fixed headroom: live repository files use repository-relative path, file-local
// case index and mode. Pinned cohere text/generated keys cannot shift when a
// repository file is added. Check kind is part of the stable key.
// ParserShard describes a shard or one of its agreement or mutant groups.
type ParserShard struct {
	Name, Kind               string
	LowerBound, hi, Mutation int
	Cases                    []int
	Groups                   []ParserShard
}

// ParserCaseKey identifies one check independently of corpus enumeration order.
func ParserCaseKey(kind string, mutation int, key string) string {
	return fmt.Sprintf("%s/%d/%s", kind, mutation, key)
}

// cssParserBucket is the shard that owns a check: a hash of its kind, mutation and stable key, so adding a file
// never moves another check.
func cssParserBucket(kind string, mutation int, key string) int {
	hash := sha256.Sum256([]byte(ParserCaseKey(kind, mutation, key)))
	return int(binary.BigEndian.Uint64(hash[:8]) % ParserShardCount)
}

// ParserMutantBucket is the one shard where a mutant's group runs: the shard that owns its first witness. Over cases
// that aren't witnesses a mutant could fail only by crashing, which is the mutant's behavior, not the port's; on Oct 12
// those runs took 702 of the family's active seconds (#testaudit). It is -1 when the mutant has no witness in the corpus.
func ParserMutantBucket(keys []string, mutation int, witnesses []int) int {
	if len(witnesses) == 0 || witnesses[0] < 0 || witnesses[0] >= len(keys) {
		return -1
	}
	return cssParserBucket("mutant", mutation, keys[witnesses[0]])
}

// ParserAgreementShards gives every shard its agreement group, every case by its hash owner.
func ParserAgreementShards(keys []string) []ParserShard {
	units := make([]ParserShard, ParserShardCount)
	for i := range units {
		units[i].Name = fmt.Sprintf("shard-%03d", i)
	}
	buckets := make([][]int, len(units))
	for i, key := range keys {
		bucket := cssParserBucket("agreement", -1, key)
		buckets[bucket] = append(buckets[bucket], i)
	}
	for i, cases := range buckets {
		if len(cases) > 0 {
			units[i].Groups = append(units[i].Groups, ParserShard{Name: units[i].Name, Kind: "agreement", Mutation: -1, Cases: cases})
		}
	}
	return units
}

// ParserWithMutants adds each mutant's group, its witnesses (witnesses[mutation], from cssParserWitnesses), to the
// shard that owns its first witness. It copies the plan rather than changing the one it's given.
func ParserWithMutants(units []ParserShard, keys []string, witnesses [][]int) []ParserShard {
	plan := make([]ParserShard, len(units))
	for i, unit := range units {
		plan[i] = unit
		plan[i].Groups = append([]ParserShard{}, unit.Groups...)
	}
	for mutation, ids := range witnesses {
		if bucket := ParserMutantBucket(keys, mutation, ids); bucket >= 0 {
			plan[bucket].Groups = append(plan[bucket].Groups, ParserShard{Name: plan[bucket].Name, Kind: "mutant", Mutation: mutation, Cases: ids})
		}
	}
	return plan
}

// WriteParserSlice writes selected cases and returns their renumbered Go answers.
func WriteParserSlice(t *testing.T, corpus ParserCorpus, indices []int) (string, string) {
	t.Helper()
	path := filepath.Join(TemporaryDirectory(t), "cases.txt")
	var inputs, answers strings.Builder
	for local, i := range indices {
		fmt.Fprintf(&inputs, "%s\n", corpus.Inputs[i])
		fmt.Fprintf(&answers, "case %d\n%s\n", local, corpus.Answers[i])
	}
	if err := os.WriteFile(path, []byte(inputs.String()), 0644); err != nil {
		t.Fatal(err)
	}
	return path, answers.String()
}

// Every non-Go product uses the shared cache, including temporary mutants.
// Go overlay recipes also use buildcache.Product; products are immutable.
// ParserProductInputs describes the existing inputs of a CSS compiler-side product.
type ParserProductInputs struct {
	Name         string
	Files, Flags []string
	Toolchain    string
}

// ParserProduct fetches or builds a CSS product using its existing cache inputs.
func ParserProduct(t *testing.T, inputs ParserProductInputs, build func(dir string) error) string {
	t.Helper()
	keyed := cssParserCacheInputs(t, inputs)
	return buildcache.Product(t, keyed, build)
}

// Go overlay builds use the same content-addressed cache as the other products. The build runs go under
// toolchain.Env, so its key names the C and C++ compilers that environment gives go, and the C
// compiler by its report: nothing the oracle links uses cgo today, so no compiler changes its bytes, but one the key
// didn't name could start to unseen.
func cssParserGoProduct(t *testing.T, name string, build func(dir string) error) string {
	t.Helper()
	return oracleProduct(t, ParserOracleInputs(name, GoCompilerConfiguration()), build)
}

// ParserOracleInputs describes the Go parser oracle without building it.
func ParserOracleInputs(name string, compilers CompilerConfiguration) buildcache.Inputs {
	return buildcache.Inputs{
		Name:      "css " + name,
		Files:     []string{"cohere", "go.mod", "go.work", "stage1/cohere/css/testdata", "stage1/cohere/css/parser_shards_test.go"},
		Flags:     []string{"go test -c " + strings.Join(buildcache.ReproducibleGoFlags(), " "), "./internal/format/css/postcss", "overlay=internal/format/css/postcss/adamic_port_side_test.go", name, "GOFLAGS=" + os.Getenv("GOFLAGS"), "GOEXPERIMENT=" + os.Getenv("GOEXPERIMENT"), "CGO_ENABLED=" + os.Getenv("CGO_ENABLED"), "GOOS=" + os.Getenv("GOOS"), "GOARCH=" + os.Getenv("GOARCH"), "CC=" + compilers.CCompiler, "CXX=" + compilers.CPlusPlusCompiler},
		Toolchain: []string{buildcache.Tool("go", "version"), runtime.GOOS, runtime.GOARCH, compilers.Report},
	}
}

// Temporary generated inputs are keyed by content and logical filename. Their
// directory names never enter a key; repository inputs stay repository-relative.
func cssParserCacheInputs(t *testing.T, inputs ParserProductInputs) buildcache.Inputs {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	keyed := buildcache.Inputs{Name: "css " + inputs.Name, Flags: append([]string{}, inputs.Flags...), Toolchain: []string{inputs.Toolchain, buildcache.Tool("go", "version"), runtime.GOOS + "/" + runtime.GOARCH}}
	keyed.Files = append(keyed.Files, "stage1/cohere/css/parser_shards_test.go", "stage1/cohere/css/css_test.go")
	keyed.Flags = append(keyed.Flags, "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"))

	for _, name := range inputs.Files {
		absolute, err := filepath.Abs(name)
		if err != nil {
			t.Fatal(err)
		}
		relative, err := filepath.Rel(root, absolute)
		if err != nil {
			t.Fatal(err)
		}
		if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			content, err := os.ReadFile(name)
			if err != nil {
				t.Fatal(err)
			}
			logical := filepath.Base(name)
			if filepath.Ext(name) == ".ts" {
				logical = filepath.Base(filepath.Dir(name)) + "/" + logical
			}
			keyed.Flags = append(keyed.Flags, fmt.Sprintf("generated-input=%s:%x", logical, sha256.Sum256(content)))
		} else {
			keyed.Files = append(keyed.Files, filepath.ToSlash(relative))
		}
	}
	// A lowering build runs the compiler linked into this test binary. Its key names the compiler by its sources (the
	// internal/ and cohere/TypeScript directories, go.mod, the Go release) rather than by this binary's bytes, which
	// differ wherever the test binary was built, so Workshop's product and a runner's ask are one key (#nm31pcn).

	return keyed
}

func cssParserOracle(t *testing.T) string {
	t.Helper()
	repo, _ := filepath.Abs(repository)
	cohere := filepath.Join(repo, "cohere")
	side, _ := filepath.Abs("testdata/cohere_side_test.go")
	dir := cssParserGoProduct(t, "Go parser oracle", func(dir string) error {
		overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{filepath.Join(cohere, "internal/format/css/postcss/adamic_port_side_test.go"): side}})
		if err != nil {
			return err
		}
		path := filepath.Join(dir, "overlay.json")
		if err := os.WriteFile(path, overlay, 0644); err != nil {
			return err
		}
		// The overlay names this checkout's paths: it builds the product and is never part of it (#tqrqx60).
		defer os.Remove(path)
		command := toolchain.CommandContextIn(t.Context(), cohere, "go", append(append([]string{"test"}, buildcache.ReproducibleGoFlags()...), "-c", "-overlay="+path, "-o", filepath.Join(dir, "oracle"), "./internal/format/css/postcss")...)
		command.WaitDelay = time.Second
		if output, err := childguard.CombinedOutput(command, childguard.Options{Stall: ChildStallLimit}); err != nil {
			return fmt.Errorf("Go oracle: %w\n%s", err, output)
		}
		return nil
	})
	return filepath.Join(dir, "oracle")
}

// ContentIdentity hashes the paths and content of a file tree in stable order.
func ContentIdentity(t *testing.T, path string) string {
	t.Helper()
	hash := sha256.New()
	err := filepath.WalkDir(path, func(name string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		content, err := os.ReadFile(name)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(path, name)
		if err != nil {
			return err
		}
		fmt.Fprintf(hash, "%s\x00%x\n", filepath.ToSlash(relative), sha256.Sum256(content))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("oracle-content-sha256=%x", hash.Sum(nil))
}

// CachedParserOutputs returns the parser corpus file and its cached Go answers.
func CachedParserOutputs(t *testing.T) (string, string) {
	t.Helper()
	oracle := cssParserOracle(t)
	repo, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	side, err := filepath.Abs("testdata/cohere_side_test.go")
	if err != nil {
		t.Fatal(err)
	}
	inputs := ParserProductInputs{Name: "Go parser oracle answers", Files: []string{side, filepath.Join(repo, "cohere/internal/format/css")}, Flags: []string{ContentIdentity(t, oracle)}, Toolchain: buildcache.Tool("go", "version")}
	// The repository's corpus files come from its corpus roots alone (cssCorpusRoots), never a walk of the whole
	// repository, whose listings would enter every corpus-reading product's read set (#t128y5x).
	inputs.Flags = append(inputs.Flags, "corpus-roots="+strings.Join(cssCorpusRoots, ","))
	if err := WalkCorpusRoots(repo, func(path string) error { inputs.Files = append(inputs.Files, path); return nil }); err != nil {
		t.Fatal(err)
	}
	if fixtures := os.Getenv("ADAMIC_CSS_FIXTURES"); fixtures != "" {
		err := filepath.WalkDir(fixtures, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				if entry.Name() == ".git" || entry.Name() == "node_modules" {
					return filepath.SkipDir
				}
				return nil
			}
			if !IsCorpusFile(path) {
				return nil
			}
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			relative, err := filepath.Rel(fixtures, path)
			if err != nil {
				return err
			}
			inputs.Flags = append(inputs.Flags, fmt.Sprintf("fixture=%s:%x", filepath.ToSlash(relative), sha256.Sum256(content)))
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	dir := oracleParserProduct(t, inputs, func(dir string) error {
		request := map[string]any{"cases": filepath.Join(dir, "cases.txt"), "answers": filepath.Join(dir, "answers.txt"), "ids": filepath.Join(dir, "ids.json"), "repository": repo, "roots": cssCorpusRoots, "fixtures": os.Getenv("ADAMIC_CSS_FIXTURES")}
		data, err := json.Marshal(request)
		if err != nil {
			return err
		}
		requestPath := filepath.Join(TemporaryDirectory(t), "request.json")
		if err := os.WriteFile(requestPath, data, 0644); err != nil {
			return err
		}
		cmd := toolchain.CommandContextIn(t.Context(), "", oracle, "-test.timeout=0", "-test.v", "-test.count=1", "-test.run=^TestAdamicPortCases$")
		cmd.WaitDelay = time.Second
		cmd.Dir = filepath.Join(repo, "cohere/internal/format/css/postcss")
		cmd.Env = append(os.Environ(), "ADAMIC_PORT_REQUEST="+requestPath)
		output, err := childguard.CombinedOutput(cmd, childguard.Options{Stall: ChildStallLimit})
		if err != nil {
			return fmt.Errorf("Go oracle: %w\n%s", err, output)
		}
		t.Logf("Go oracle: %s", output)
		return nil
	})
	cases := filepath.Join(dir, "cases.txt")
	answers, err := os.ReadFile(filepath.Join(dir, "answers.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if keep := os.Getenv("ADAMIC_CSS_KEEP_RAW"); keep != "" {
		if err := os.WriteFile(keep, answers, 0644); err != nil {
			t.Fatal(err)
		}
	}
	if keep := os.Getenv("ADAMIC_CSS_KEEP"); keep != "" {
		content, err := os.ReadFile(cases)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(keep, content, 0644); err != nil {
			t.Fatal(err)
		}
	}
	return cases, string(answers)
}

// CachedPostCSSAnswers returns the cached PostCSS answers for the supplied cases.
func CachedPostCSSAnswers(t *testing.T, script, library, path, identity string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	dir := oracleParserProduct(t, ParserProductInputs{Name: "PostCSS oracle answers", Files: []string{script}, Flags: []string{identity, fmt.Sprintf("input-sha256=%x", sha256.Sum256(data))}, Toolchain: buildcache.Tool("node", "--version")}, func(dir string) error {
		input := filepath.Join(dir, "input.txt")
		if err := os.WriteFile(input, data, 0644); err != nil {
			return err
		}
		cmd := toolchain.Command("node", script, library, input)
		cmd.Env = os.Environ()
		cmd.WaitDelay = time.Second
		var stdout, stderr bytes.Buffer
		cmd.Stdout, cmd.Stderr = &stdout, &stderr
		if err := childguard.Run(cmd, childguard.Options{Stall: ChildStallLimit, Ceiling: 90 * time.Second}); err != nil || stderr.Len() != 0 {
			return fmt.Errorf("PostCSS oracle: %v; %s", err, &stderr)
		}
		return os.WriteFile(filepath.Join(dir, "answers.txt"), stdout.Bytes(), 0644)
	})
	answer, err := os.ReadFile(filepath.Join(dir, "answers.txt"))
	if err != nil {
		t.Fatal(err)
	}
	return string(answer)
}

// cssParserPostCSSAnswersProducts is how many product tests share the agreement units' PostCSS answers, one node run
// per unit (388 units on Oct 10), so each product test stays well inside a product's 600 s.
const cssParserPostCSSAnswersProducts = 8

// cssParserPostCSSAnswers builds the PostCSS answers that every agreement unit whose ordinal is part modulo
// cssParserPostCSSAnswersProducts reads, each keyed as that unit asks for it: its own cases, the pinned library
// (ADAMIC_CSS_LIBRARY) and testdata/library.mjs. Without the library the units skip, and so does this.
func cssParserPostCSSAnswers(t *testing.T, part int) {
	t.Helper()
	library := os.Getenv("ADAMIC_CSS_LIBRARY")
	if library == "" {
		t.Skip("set ADAMIC_CSS_LIBRARY to the pinned npm scratch directory")
	}
	script, err := filepath.Abs("testdata/library.mjs")
	if err != nil {
		t.Fatal(err)
	}
	identity := ContentIdentity(t, library)
	prepareCSSParserTopCorpus(t)
	corpus := LoadParserCorpus(t)
	ordinal, built := 0, 0
	for _, root := range ParserAgreementShards(corpus.CaseKeys) {
		for _, unit := range root.Groups {
			if unit.Kind != "agreement" {
				continue
			}
			if ordinal%cssParserPostCSSAnswersProducts == part {
				cases, _ := WriteParserSlice(t, corpus, unit.Cases)
				CachedPostCSSAnswers(t, script, library, cases, identity)
				built++
			}
			ordinal++
		}
	}
	if built == 0 {
		t.Fatalf("part %d of %d holds no agreement unit, of %d", part, cssParserPostCSSAnswersProducts, ordinal)
	}
	t.Logf("PostCSS answers for %d of %d agreement units", built, ordinal)
}

var cssOracleOnce [2]sync.Once

var cssOracles [2]string

// ComparisonOracle returns the Go parser (printer=false) or printer (printer=true) comparison binary.
func ComparisonOracle(t *testing.T, printer bool) string {
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
		directory := oracleProduct(t, inputs, func(dir string) error {
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
			cmd := toolchain.CommandContextIn(t.Context(), filepath.Join(repo, "cohere"), "go", append(append([]string{"test"}, buildcache.ReproducibleGoFlags()...), "-c", "-overlay="+overlayPath, "-o", filepath.Join(dir, "oracle"), "./internal/format/css/"+pkg)...)
			// This printer build historically inherited the caller environment.
			cmd.Env = nil
			output, err := childguard.CombinedOutput(cmd, childguard.Options{Unbounded: true})
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

var cssParserTopRootOnce sync.Once

var cssParserTopRoot string

var cssParserTopRootError error

// TemporaryDirectory creates shared test inputs that outlive an individual parallel test.
func TemporaryDirectory(t *testing.T) string {
	t.Helper()
	cssParserTopRootOnce.Do(func() { cssParserTopRoot, cssParserTopRootError = os.MkdirTemp("", "adamic-css-top-") })
	if cssParserTopRootError != nil {
		t.Fatal(cssParserTopRootError)
	}
	dir, err := os.MkdirTemp(cssParserTopRoot, "input-")
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

var cssParserTopCorpusSetup struct {
	once   sync.Once
	corpus ParserCorpus
	units  []ParserShard
}

func prepareCSSParserTopCorpus(t *testing.T) {
	t.Helper()
	cssParserTopCorpusSetup.once.Do(func() {
		cases, answers := CachedParserOutputs(t)
		cssParserTopCorpusSetup.corpus = readCSSParserCorpus(t, cases, answers)
		cssParserTopCorpusSetup.units = ParserAgreementShards(cssParserTopCorpusSetup.corpus.CaseKeys)
	})
}

// LoadParserCorpus loads the shared corpus once and checks that setup completed.
func LoadParserCorpus(t *testing.T) ParserCorpus {
	t.Helper()
	prepareCSSParserTopCorpus(t)

	if len(cssParserTopCorpusSetup.corpus.CaseKeys) == 0 {
		t.Fatal("shared corpus setup did not complete")
	}
	return cssParserTopCorpusSetup.corpus
}

// Only the compiler-free oracles test binary may settle oracle products. Ordinary
// untraced developer runs may still build them locally. A traced consumer must
// fetch the producer's recorded read set instead of attaching its own recipe.
var oracleProducer bool

func oracleProduct(t *testing.T, inputs buildcache.Inputs, build func(string) error) string {
	t.Helper()
	// A fresh name key prevents recordUnion from retaining an old compiler-linked
	// recipe that still evaluates unchanged on the migration tree.
	inputs.Name += " oracle-owner-v1"
	return buildcache.Product(t, inputs, func(directory string) error {
		if buildcache.Tracing() && !oracleProducer {
			return fmt.Errorf("oracle %s is missing: build stage1/cohere/css/oracles TestProduct_ units first", inputs.Name)
		}
		return build(directory)
	})
}

func oracleParserProduct(t *testing.T, inputs ParserProductInputs, build func(string) error) string {
	t.Helper()
	return oracleProduct(t, cssParserCacheInputs(t, inputs), build)
}

// cssCorpusRoots are the directories, relative to the repository, that css's corpus takes its CSS, SCSS and Less files
// from. The Go oracle and the corpus product's key walk only these, never the whole repository: a walk of the
// repository lists its root and every directory under it, and since a product's read set takes the reads its process
// made before it, a new file anywhere in the repository moved every corpus-reading product, all 388 PostCSS answers
// among them (#t128y5x). Sorted, so walking them in order visits the files in the order a whole-repository walk did.
// TestCSSCorpusRootsHoldEveryCorpusFile keeps them complete.
var cssCorpusRoots = []string{"cohere/internal/format/css/testdata", "cohere/internal/lint/rules/tailwind", "stage1/cohere/css/testdata"}

// IsCorpusFile reports whether a file name has one of the corpus's extensions.
func IsCorpusFile(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".css", ".scss", ".less":
		return true
	}
	return false
}

// WalkCorpusRoots walks every corpus root under the repository in order, as the Go oracle's side tests do, skipping
// .git and node_modules, and calls visit with each corpus file's absolute path.
func WalkCorpusRoots(repo string, visit func(path string) error) error {
	for _, root := range cssCorpusRoots {
		err := filepath.WalkDir(filepath.Join(repo, filepath.FromSlash(root)), func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				if entry.Name() == ".git" || entry.Name() == "node_modules" {
					return filepath.SkipDir
				}
				return nil
			}
			if !IsCorpusFile(path) {
				return nil
			}
			return visit(path)
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// EnableProductProduction grants this process ownership of the oracle products.
// Only the compiler-free oracles test binary calls it, from TestMain before tests run.
func EnableProductProduction() {
	oracleProducer = true
}

// CleanupTemporaryDirectories removes the shared inputs after either test binary finishes.
func CleanupTemporaryDirectories() error {
	if cssParserTopRoot == "" {
		return nil
	}
	return os.RemoveAll(cssParserTopRoot)
}

// CorpusRoots returns a copy of the repository-relative directories enumerated by the oracle.
func CorpusRoots() []string {
	return append([]string{}, cssCorpusRoots...)
}
