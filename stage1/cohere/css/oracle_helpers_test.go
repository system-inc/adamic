package css

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/toolchain"
)

// cssParserCommand bounds each owned child, including its compiler descendants.
func cssParserCommand(t *testing.T, name string, arguments ...string) *exec.Cmd {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	return cssParserCommandContext(t, ctx, cancel, name, arguments...)
}

// Shared setup has no deadline; cancellation still kills the process group.
func cssParserSetupCommand(t *testing.T, name string, arguments ...string) *exec.Cmd {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	return cssParserCommandContext(t, ctx, cancel, name, arguments...)
}

// cssParserGoCommand is cssParserSetupCommand for go as toolchain resolves it, run in directory under toolchain.Env,
// the environment whose compilers the Go oracle's key names.
func cssParserGoCommand(t *testing.T, directory string, arguments ...string) *exec.Cmd {
	t.Helper()
	resolved := toolchain.CommandIn(directory, "go", arguments...)
	if resolved.Err != nil {
		return resolved
	}
	command := cssParserSetupCommand(t, resolved.Path, arguments...)
	command.Dir, command.Env = resolved.Dir, resolved.Env
	return command
}

// cssCompilers are the C and C++ compilers go builds a Go oracle with, and the C compiler's report.
type cssCompilers struct{ c, cPlusPlus, report string }

// cssGoCompilers reads the compilers toolchain.Env gives go: the caller's CC and CXX, or clang and clang++ when it
// names neither.
func cssGoCompilers() cssCompilers {
	var compilers cssCompilers
	for _, setting := range toolchain.Env() {
		switch name, value, _ := strings.Cut(setting, "="); name {
		case "CC":
			compilers.c = value
		case "CXX":
			compilers.cPlusPlus = value
		}
	}
	compilers.report = "no C compiler"
	if fields := strings.Fields(compilers.c); len(fields) > 0 {
		compilers.report = buildcache.Tool(fields[0], "--version")
	}
	return compilers
}

func cssParserCommandContext(t *testing.T, ctx context.Context, cancel context.CancelFunc, name string, arguments ...string) *exec.Cmd {
	t.Helper()
	command := exec.CommandContext(ctx, name, arguments...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		if command.Process == nil {
			return os.ErrProcessDone
		}
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	command.WaitDelay = time.Second
	t.Cleanup(func() {
		cancel()
		if command.Process != nil && command.ProcessState == nil {
			_ = command.Cancel()
		}
	})
	return command
}

// Setup retains process-group cancellation without a deadline of its own.
func compositionSetupCommand(t *testing.T, name string, arguments ...string) *exec.Cmd {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	return compositionCommandContext(ctx, name, arguments...)
}

// compositionGoCommand is compositionSetupCommand for go as toolchain resolves it,
// run in directory under toolchain.Env, the environment whose compilers the
// oracle's key names.
func compositionGoCommand(t *testing.T, directory string, arguments ...string) *exec.Cmd {
	t.Helper()
	resolved := toolchain.CommandIn(directory, "go", arguments...)
	if resolved.Err != nil {
		return resolved
	}
	command := compositionSetupCommand(t, resolved.Path, arguments...)
	command.Dir, command.Env = resolved.Dir, resolved.Env
	return command
}

func compositionCommandContext(ctx context.Context, name string, arguments ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, arguments...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return os.ErrProcessDone
		}
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	cmd.WaitDelay = time.Second
	return cmd
}

// The overlay source and the complete Go module trees address each oracle by
// content. Corpus enumeration stays live: only its binary is cached. The build
// runs go under toolchain.Env (compositionGoCommand), so the key names the C and
// C++ compilers that environment gives go, and the C compiler by its report.
func compositionOracle(t *testing.T, name, sidePath, overlayTarget, packagePath string) string {
	t.Helper()
	repo, _ := filepath.Abs(repository)
	side, _ := filepath.Abs(sidePath)
	sideRelative, err := filepath.Rel(repo, side)
	if err != nil {
		t.Fatal(err)
	}
	inputs := compositionOracleInputs(name, sideRelative, overlayTarget, packagePath, cssGoCompilers())
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
		command := compositionGoCommand(t, filepath.Join(repo, "cohere"), "test", "-c", "-overlay="+overlayPath, "-o", filepath.Join(dir, "oracle"), packagePath)
		output, err := command.CombinedOutput()
		if err != nil {
			return fmt.Errorf("Go oracle: %w\n%s", err, output)
		}
		return nil
	})
	return filepath.Join(directory, "oracle")
}

func compositionOracleInputs(name, sideRelative, overlayTarget, packagePath string, compilers cssCompilers) buildcache.Inputs {
	return buildcache.Inputs{
		Name:      name,
		Files:     []string{"cohere", "go.mod", "go.work", sideRelative},
		Flags:     []string{"go test -c", packagePath, "overlay=" + overlayTarget, "GOFLAGS=" + os.Getenv("GOFLAGS"), "GOEXPERIMENT=" + os.Getenv("GOEXPERIMENT"), "CGO_ENABLED=" + os.Getenv("CGO_ENABLED"), "GOOS=" + os.Getenv("GOOS"), "GOARCH=" + os.Getenv("GOARCH"), "CC=" + compilers.c, "CXX=" + compilers.cPlusPlus},
		Toolchain: []string{buildcache.Tool("go", "version"), runtime.GOOS, runtime.GOARCH, compilers.report},
	}
}

const repository = "../../.."

const testThePortParsesAsGoCohereDoesShards = 388

// Loaded formatter output gaps reached 66.19 seconds; four minutes gives over 3x headroom.
const childStall = 4 * time.Minute

type mutant struct{ name, file, from, to string }

var mutants = []mutant{
	{"custom properties lose their block values", "parser.ts", "const custom = start.value.startsWith('--');", "const custom = false;"},
	{"comments always disappear from values", "parser.ts", "else {\n                        value += current.value;\n                    }", "else {\n                        value += '';\n                    }"},
	{"closing braces no longer include the closing byte", "parser.ts", "this.finish(node, token.start);\n            this.current = node.parent;", "this.finish(node, token.start, 0);\n            this.current = node.parent;"},
}

type cssParserCorpus struct{ inputs, answers, keys []string }

func readCSSParserCorpus(t *testing.T, path, answers string) cssParserCorpus {
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
	corpus := cssParserCorpus{inputs: inputs}
	manifest, err := os.ReadFile(filepath.Join(filepath.Dir(path), "ids.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(manifest, &corpus.keys); err != nil {
		t.Fatal(err)
	}
	if len(corpus.keys) != len(inputs) {
		t.Fatal("stable IDs differ from live enumeration")
	}
	for i := range inputs {
		if lines[i*2] != fmt.Sprintf("case %d", i) {
			t.Fatalf("missing or repeated unsplit case ID at %d: %q", i, lines[i*2])
		}
		corpus.answers = append(corpus.answers, lines[i*2+1])
	}
	return corpus
}

// Fixed headroom: live repository files use repository-relative path, file-local
// case index and mode. Pinned cohere text/generated keys cannot shift when a
// repository file is added. Check kind is part of the stable key.
type cssParserShard struct {
	name, kind       string
	lo, hi, mutation int
	cases            []int
	groups           []cssParserShard
}

func cssParserCaseKey(kind string, mutation int, key string) string {
	return fmt.Sprintf("%s/%d/%s", kind, mutation, key)
}

func cssParserShards(keys []string) []cssParserShard {
	units := make([]cssParserShard, testThePortParsesAsGoCohereDoesShards)
	for i := range units {
		units[i].name = fmt.Sprintf("shard-%03d", i)
	}
	for variant := -1; variant < len(mutants); variant++ {
		kind := "agreement"
		if variant >= 0 {
			kind = "mutant"
		}
		buckets := make([][]int, len(units))
		for i, key := range keys {
			hash := sha256.Sum256([]byte(cssParserCaseKey(kind, variant, key)))
			bucket := int(binary.BigEndian.Uint64(hash[:8]) % uint64(len(units)))
			buckets[bucket] = append(buckets[bucket], i)
		}
		for i, cases := range buckets {
			if len(cases) > 0 {
				units[i].groups = append(units[i].groups, cssParserShard{name: units[i].name, kind: kind, mutation: variant, cases: cases})
			}
		}
	}
	return units
}

func writeCSSParserSlice(t *testing.T, corpus cssParserCorpus, indices []int) (string, string) {
	t.Helper()
	path := filepath.Join(cssParserTopTempDir(t), "cases.txt")
	var inputs, answers strings.Builder
	for local, i := range indices {
		fmt.Fprintf(&inputs, "%s\n", corpus.inputs[i])
		fmt.Fprintf(&answers, "case %d\n%s\n", local, corpus.answers[i])
	}
	if err := os.WriteFile(path, []byte(inputs.String()), 0644); err != nil {
		t.Fatal(err)
	}
	return path, answers.String()
}

// Every non-Go product uses the shared cache, including temporary mutants.
// Go overlay recipes also use buildcache.Product; products are immutable.
type cssParserInputs struct {
	Name         string
	Files, Flags []string
	Toolchain    string
}

func cssParserProduct(t *testing.T, inputs cssParserInputs, build func(dir string) error) string {
	t.Helper()
	keyed := cssParserCacheInputs(t, inputs)
	return buildcache.Product(t, keyed, build)
}

// Go overlay builds use the same content-addressed cache as the other products. The build runs go under
// toolchain.Env (cssParserGoCommand), so its key names the C and C++ compilers that environment gives go, and the C
// compiler by its report: nothing the oracle links uses cgo today, so no compiler changes its bytes, but one the key
// didn't name could start to unseen.
func cssParserGoProduct(t *testing.T, name string, build func(dir string) error) string {
	t.Helper()
	return oracleProduct(t, cssParserGoProductInputs(name, cssGoCompilers()), build)
}

func cssParserGoProductInputs(name string, compilers cssCompilers) buildcache.Inputs {
	return buildcache.Inputs{
		Name:      "css " + name,
		Files:     []string{"cohere", "go.mod", "go.work", "stage1/cohere/css/testdata", "stage1/cohere/css/parser_shards_test.go"},
		Flags:     []string{"go test -c", "./internal/format/css/postcss", "overlay=internal/format/css/postcss/adamic_port_side_test.go", name, "GOFLAGS=" + os.Getenv("GOFLAGS"), "GOEXPERIMENT=" + os.Getenv("GOEXPERIMENT"), "CGO_ENABLED=" + os.Getenv("CGO_ENABLED"), "GOOS=" + os.Getenv("GOOS"), "GOARCH=" + os.Getenv("GOARCH"), "CC=" + compilers.c, "CXX=" + compilers.cPlusPlus},
		Toolchain: []string{buildcache.Tool("go", "version"), runtime.GOOS, runtime.GOARCH, compilers.report},
	}
}

// Temporary generated inputs are keyed by content and logical filename. Their
// directory names never enter a key; repository inputs stay repository-relative.
func cssParserCacheInputs(t *testing.T, inputs cssParserInputs) buildcache.Inputs {
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
		command := cssParserGoCommand(t, cohere, "test", "-c", "-overlay="+path, "-o", filepath.Join(dir, "oracle"), "./internal/format/css/postcss")
		if output, err := childguard.CombinedOutput(command, childguard.Options{Stall: childStall}); err != nil {
			return fmt.Errorf("Go oracle: %w\n%s", err, output)
		}
		return nil
	})
	return filepath.Join(dir, "oracle")
}

func cssParserOracleIdentity(t *testing.T, path string) string {
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

func cachedCSSParserOracleOutputs(t *testing.T, oracle string) (string, string) {
	t.Helper()
	repo, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	side, err := filepath.Abs("testdata/cohere_side_test.go")
	if err != nil {
		t.Fatal(err)
	}
	inputs := cssParserInputs{Name: "Go parser oracle answers", Files: []string{side, filepath.Join(repo, "cohere/internal/format/css")}, Flags: []string{cssParserOracleIdentity(t, oracle)}, Toolchain: buildcache.Tool("go", "version")}
	roots := []string{repo}
	if fixtures := os.Getenv("ADAMIC_CSS_FIXTURES"); fixtures != "" {
		roots = append(roots, fixtures)
	}
	inputs.Flags = append(inputs.Flags, fmt.Sprintf("corpus-roots=%d", len(roots)))
	for index, root := range roots {
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				if entry.Name() == ".git" || entry.Name() == "node_modules" {
					return filepath.SkipDir
				}
				return nil
			}
			ext := strings.ToLower(filepath.Ext(path))
			if ext != ".css" && ext != ".scss" && ext != ".less" {
				return nil
			}
			if index == 0 {
				inputs.Files = append(inputs.Files, path)
				return nil
			}
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			relative, err := filepath.Rel(root, path)
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
		request := map[string]string{"cases": filepath.Join(dir, "cases.txt"), "answers": filepath.Join(dir, "answers.txt"), "ids": filepath.Join(dir, "ids.json"), "repository": repo, "fixtures": os.Getenv("ADAMIC_CSS_FIXTURES")}
		data, err := json.Marshal(request)
		if err != nil {
			return err
		}
		requestPath := filepath.Join(cssParserTopTempDir(t), "request.json")
		if err := os.WriteFile(requestPath, data, 0644); err != nil {
			return err
		}
		cmd := cssParserSetupCommand(t, oracle, "-test.timeout=0", "-test.v", "-test.count=1", "-test.run=^TestAdamicPortCases$")
		cmd.Dir = filepath.Join(repo, "cohere/internal/format/css/postcss")
		cmd.Env = append(os.Environ(), "ADAMIC_PORT_REQUEST="+requestPath)
		output, err := childguard.CombinedOutput(cmd, childguard.Options{Stall: childStall})
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

func cachedCSSParserPostCSSAnswers(t *testing.T, script, library, path, identity string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	dir := oracleParserProduct(t, cssParserInputs{Name: "PostCSS oracle answers", Files: []string{script}, Flags: []string{identity, fmt.Sprintf("input-sha256=%x", sha256.Sum256(data))}, Toolchain: buildcache.Tool("node", "--version")}, func(dir string) error {
		input := filepath.Join(dir, "input.txt")
		if err := os.WriteFile(input, data, 0644); err != nil {
			return err
		}
		cmd := cssParserCommand(t, "node", script, library, input)
		var stdout, stderr bytes.Buffer
		cmd.Stdout, cmd.Stderr = &stdout, &stderr
		if err := childguard.Run(cmd, childguard.Options{Stall: childStall}); err != nil || stderr.Len() != 0 {
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
	identity := cssParserOracleIdentity(t, library)
	prepareCSSParserTopCorpus(t)
	corpus := cssParserTopCorpus(t)
	ordinal, built := 0, 0
	for _, root := range cssParserShards(corpus.keys) {
		for _, unit := range root.groups {
			if unit.kind != "agreement" {
				continue
			}
			if ordinal%cssParserPostCSSAnswersProducts == part {
				cases, _ := writeCSSParserSlice(t, corpus, unit.cases)
				cachedCSSParserPostCSSAnswers(t, script, library, cases, identity)
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

var cssParserTopRootOnce sync.Once

var cssParserTopRoot string

var cssParserTopRootError error

func cssParserTopTempDir(t *testing.T) string {
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
	corpus cssParserCorpus
	units  []cssParserShard
}

func prepareCSSParserTopCorpus(t *testing.T) {
	t.Helper()
	cssParserTopCorpusSetup.once.Do(func() {
		oracle := cssParserOracle(t)
		cases, answers := cachedCSSParserOracleOutputs(t, oracle)
		cssParserTopCorpusSetup.corpus = readCSSParserCorpus(t, cases, answers)
		cssParserTopCorpusSetup.units = cssParserShards(cssParserTopCorpusSetup.corpus.keys)
	})
}

func cssParserTopCorpus(t *testing.T) cssParserCorpus {
	t.Helper()
	prepareCSSParserTopCorpus(t)

	if len(cssParserTopCorpusSetup.corpus.keys) == 0 {
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

func oracleParserProduct(t *testing.T, inputs cssParserInputs, build func(string) error) string {
	t.Helper()
	return oracleProduct(t, cssParserCacheInputs(t, inputs), build)
}
