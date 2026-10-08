package typeaware

import (
	"bytes"
	"encoding/json"
	"fmt"
	goast "go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/corpusfiles"
)

const compilerCommit = "050880ce59e30b356b686bd3144efe24f875ebc8"

type result struct {
	stdout, stderr []byte
	elapsed        time.Duration
	err            error
}
type harness struct {
	t                     *testing.T
	repository, directory string
	next                  int
	sixBuilds             bool
	setupStarted          time.Time
	sixFiles              []string
	sixVersions           []string
	sixProducts           map[string]sixBuildInputs
}

func (h *harness) run(name string, command *exec.Cmd) result {
	h.t.Helper()
	h.next++
	if command.Dir == "" {
		command.Dir = h.repository
	}
	stem := filepath.Join(h.directory, fmt.Sprintf("%03d-%s", h.next, strings.ReplaceAll(name, "/", "-")))
	out, err := os.Create(stem + ".stdout")
	if err != nil {
		h.t.Fatal(err)
	}
	defer out.Close()
	report, err := os.Create(stem + ".stderr")
	if err != nil {
		h.t.Fatal(err)
	}
	defer report.Close()
	command.Stdout = out
	command.Stderr = report
	started := time.Now()
	runError := command.Run()
	elapsed := time.Since(started)

	stdout, err := os.ReadFile(out.Name())
	if err != nil {
		h.t.Fatal(err)
	}
	stderr, err := os.ReadFile(report.Name())
	if err != nil {
		h.t.Fatal(err)
	}
	return result{stdout, stderr, elapsed, runError}
}
func (h *harness) must(name string, command *exec.Cmd) result {
	h.t.Helper()
	r := h.run(name, command)
	if r.err != nil {
		h.t.Fatalf("%s: %v\n%s\n%s", name, r.err, r.stdout, r.stderr)
	}
	return r
}
func (h *harness) write(name, text string) string {
	h.t.Helper()
	path := filepath.Join(h.directory, name)
	if err := os.WriteFile(path, []byte(text), 0644); err != nil {
		h.t.Fatal(err)
	}
	return path
}
func (h *harness) overlay(name, path, from, to string) string {
	h.t.Helper()
	original := filepath.Join(h.repository, path)
	data, err := os.ReadFile(original)
	if err != nil {
		h.t.Fatal(err)
	}
	if strings.Count(string(data), from) != 1 {
		h.t.Fatalf("nonunique mutant %s", name)
	}
	side := h.write(name+filepath.Ext(path), strings.Replace(string(data), from, to, 1))
	data, err = json.Marshal(map[string]any{"Replace": map[string]string{original: side}})
	if err != nil {
		h.t.Fatal(err)
	}
	return h.write(name+".json", string(data))
}
func (h *harness) archive(name, overlay string, sanitize bool) string {
	h.t.Helper()
	path := filepath.Join(h.directory, name+".a")
	args := []string{"build", "-buildmode=c-archive", "-o", path}
	if overlay != "" {
		args = append(args, "-overlay", overlay)
	}
	args = append(args, "./bridge/tsgo/archive")
	cmd := exec.Command("go", args...)
	if sanitize {
		cmd.Env = append(os.Environ(), "CC=clang", "CGO_CFLAGS=-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all")
	}
	if h.sixBuilds {
		return h.sixBuildProduct(name, cmd)
	}
	h.must(name, cmd)
	return path
}
func (h *harness) build(stage0, name, entry, archive string, sanitize bool) string {
	h.t.Helper()
	path := filepath.Join(h.directory, name)
	args := []string{"build", entry, "-o", path, "--tsgo", archive}
	if sanitize {
		args = append(args, "--sanitize")
	}
	if h.sixBuilds {
		return h.sixBuildProduct(name, exec.Command(stage0, args...))
	}
	h.must(name, exec.Command(stage0, args...))
	return path
}
func firstDifference(a, b []byte) int {
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return min(len(a), len(b))
}
func (h *harness) compare(name, oracle, binary, config, manifest string) result {
	h.t.Helper()
	want := h.must(name+"-go", exec.Command(oracle, config, manifest))
	got := h.must(name+"-native", exec.Command(binary, config, manifest))
	if len(got.stderr) != 0 {
		h.t.Fatalf("sanitizer stderr: %s", got.stderr)
	}
	if !bytes.Equal(got.stdout, want.stdout) {
		i := firstDifference(got.stdout, want.stdout)
		h.t.Fatalf("%s mismatch byte %d: native %q Go %q", name, i, got.stdout[max(0, i-50):min(len(got.stdout), i+250)], want.stdout[max(0, i-50):min(len(want.stdout), i+250)])
	}
	h.t.Logf("%s: %d identical finding bytes; %s", name, len(want.stdout), summary(want.stdout))
	return want
}
func summary(data []byte) string {
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	return lines[len(lines)-1]
}

// Not parallel: native builds, sanitizer subprocesses and timings share a machine.
// The optional external corpus is never fetched by a test.
func TestTypeAwareAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if artifacts := os.Getenv("ADAMIC_TYPEAWARE_ARTIFACTS"); artifacts != "" {
		directory, err = filepath.Abs(artifacts)
		if err != nil {
			t.Fatal(err)
		}
		if err = os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
	}
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	normal := h.archive("checker", "", false)
	sanitized := h.archive("checker-asan", "", true)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/main.ts")
	binary := h.build(stage0, "native-asan", entry, sanitized, true)
	optimized := h.build(stage0, "native", entry, normal, false)
	oracle := filepath.Join(directory, "oracle")
	virtual := filepath.Join(repository, "cohere/adamic_typeaware_oracle.go")
	data, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: filepath.Join(repository, "stage1/cohere/typeaware/testdata/oracle.go")}})
	if err != nil {
		t.Fatal(err)
	}
	overlay := h.write("oracle-overlay.json", string(data))
	cmd := exec.Command("go", "build", "-overlay", overlay, "-o", oracle, virtual)
	cmd.Dir = filepath.Join(repository, "cohere")
	h.must("oracle-build", cmd)
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")

	// Extract the production rule's table sources without duplicating its verdicts.
	sourceFile := filepath.Join(repository, "cohere/internal/lint/rules/typescript/no_unsafe_unary_minus_test.go")
	tree, err := parser.ParseFile(token.NewFileSet(), sourceFile, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	unique := map[string]bool{}
	goast.Inspect(tree, func(n goast.Node) bool {
		literal, ok := n.(*goast.CompositeLit)
		if !ok || len(literal.Elts) < 2 {
			return true
		}
		value, ok := literal.Elts[1].(*goast.BasicLit)
		if !ok || value.Kind != token.STRING {
			return true
		}
		text, err := strconv.Unquote(value.Value)
		if err == nil && strings.Contains(text, ";") {
			unique[text] = true
		}
		return true
	})
	// Whole-expression and compound operands, imports, Unicode, CRLF, constraint
	// resolution, narrowed unions, templates and repeated unsafe sites.
	unique["declare const text: string; -(text); (-text); - -text; -text.length; -({value:text}).value; -(() => text)();"] = true
	unique["/* 世界 🌍 */\r\nconst é = '🌍'; -é; -`世界`;\r\n(a: boolean) => -a;"] = true
	unique["<T extends number | string>(x:T) => -x; <T extends any>(x:T) => -x; type N = -1 | -2;"] = true
	unique["function f(x: number | string) { if(typeof x === 'number') { return -x; } return -x; }"] = true
	unique["declare const n: { value: string | number }; -n.value; -(n.value); -[1,'x'][1];"] = true
	keys := make([]string, 0, len(unique))
	for source := range unique {
		keys = append(keys, source)
	}
	sort.Strings(keys)
	var paths []string
	for i, text := range keys {
		// Treat every fixture as a module so declarations do not contaminate each other.
		paths = append(paths, h.write(fmt.Sprintf("case-%03d.ts", i), text+"\nexport {};\n"))
	}
	helper := h.write("helper.ts", "export const text: string = 'x';\n")
	paths = append(paths, h.write("import.ts", "import { text } from './helper.js'; -text;\n"), helper)
	manifest := h.write("generated.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("generated", oracle, binary, config, manifest)
	if !bytes.Contains(truth.stdout, []byte("is false instead.")) || !bytes.Contains(truth.stdout, []byte("is string instead.")) {
		t.Fatal("nonempty union/string controls missing")
	}
	var generatedFindings, generatedQueries int
	fmt.Sscanf(summary(truth.stdout), "findings %d queries %d", &generatedFindings, &generatedQueries)
	if generatedFindings < 20 || generatedQueries < 35 {
		t.Fatalf("generated coverage too small: %s", summary(truth.stdout))
	}
	t.Logf("generated coverage: %d files", len(paths))

	// No bridge calls can reach an ordinary backend without explicit linkage.
	refusal := h.run("unlinked", exec.Command(stage0, "build", entry, "-o", filepath.Join(directory, "unlinked")))
	if refusal.err == nil || !bytes.Contains(refusal.stderr, []byte("unlinked typescript-go library call")) {
		t.Fatal("unlinked type-parts call was accepted")
	}

	partsOnly := h.write("parts-unlinked.ts", "import { tsgoTypeParts } from 'adamic'; console.log(tsgoTypeParts(1, 'source.ts', 0, 1, 'Identifier'));\n")
	refusal = h.run("parts-unlinked", exec.Command(stage0, "build", partsOnly, "-o", filepath.Join(directory, "parts-unlinked")))
	if refusal.err == nil || !bytes.Contains(refusal.stderr, []byte("unlinked typescript-go library call")) {
		t.Fatal("unlinked type-parts call was accepted")
	}

	// Asking the unary expression's type rather than its operand still compiles
	// and finishes; the independent production-rule findings must disagree.
	original := filepath.Join(repository, "stage1/cohere/typeaware/unary_minus.ts")
	text, err := os.ReadFile(original)
	if err != nil {
		t.Fatal(err)
	}
	anchor := "this.parser.node(node.children[0] ?? panic('minus without operand'))"
	if strings.Count(string(text), anchor) != 1 {
		t.Fatal("wrong-node anchor changed")
	}
	// Preserve imports by making both mutant sources next to one another in scratch
	// and replacing their relative dependencies with absolute source paths.
	mutant := strings.Replace(string(text), anchor, "this.parser.node(index)", 1)
	mutant = strings.ReplaceAll(mutant, "../../typescript", filepath.Join(repository, "stage1/typescript"))
	mutant = strings.ReplaceAll(mutant, "../lint", filepath.Join(repository, "stage1/cohere/lint"))
	h.write("wrong-node.ts", mutant)
	mainText, err := os.ReadFile(entry)
	if err != nil {
		t.Fatal(err)
	}
	mainSource := strings.ReplaceAll(string(mainText), "../../typescript", filepath.Join(repository, "stage1/typescript"))
	mainSource = strings.Replace(mainSource, "./unary_minus.ts", "./wrong-node.ts", 1)
	mutantEntry := h.write("wrong-main.ts", mainSource)
	wrong := h.build(stage0, "wrong-node", mutantEntry, normal, true)
	observed := h.must("wrong-node-run", exec.Command(wrong, config, manifest))
	if len(observed.stderr) != 0 || bytes.Equal(observed.stdout, truth.stdout) {
		t.Fatal("wrong-node mutant was not caught by findings alone")
	}
	t.Logf("wrong-node type mutant: cohere byte oracle catches byte %d; %s", firstDifference(observed.stdout, truth.stdout), summary(observed.stdout))

	probe := h.write("cost-probe.ts", "-1;\n")
	releasedEntry := filepath.Join(repository, "stage1/cohere/typeaware/testdata/released.ts")
	released := h.build(stage0, "released", releasedEntry, normal, false)
	stale := h.run("released-run", exec.Command(released, config, probe))
	if code, ok := stale.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(stale.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released query escaped: %v %s", stale.err, stale.stderr)
	}
	t.Log("released program queried: native panic 70 with invalid or released checker handle")
	staleOverlay := h.overlay("stale", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains released roots.")
	staleArchive := h.archive("stale-checker", staleOverlay, false)
	staleBinary := h.build(stage0, "stale-native", releasedEntry, staleArchive, false)
	stale = h.must("stale-mutant-run", exec.Command(staleBinary, config, probe))
	if len(stale.stderr) != 0 {
		t.Fatal("stale mutant did not finish cleanly")
	}
	t.Log("released registry deletion mutant: stale-query expectation catches exit 0 instead of 70")

	// Bad frame mutants must fail explicitly before any finding can be invented.
	for _, change := range []struct{ name, value, message string }{
		{"empty-frame", "", "checker returned no type parts"},
		{"missing-header", "x", "invalid checker type frame"},
		{"bad-length", "1\n-1\n", "invalid checker type length or flags"},
	} {
		quoted := strconv.Quote(change.value)
		overlay := h.overlay(change.name, "bridge/tsgo/archive/main.go", "*parts = buffer(answer)", "_ = answer; *parts = buffer("+quoted+")")
		archive := h.archive(change.name+"-checker", overlay, false)
		malformed := h.build(stage0, change.name+"-native", entry, archive, false)
		got := h.run(change.name+"-run", exec.Command(malformed, config, manifest))
		if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || !bytes.Contains(got.stderr, []byte(change.message)) {
			t.Fatalf("%s escaped: %v %s", change.name, got.err, got.stderr)
		}
		t.Logf("%s mutant: panic 70, %s", change.name, change.message)
	}

	// A kind/span mismatch cannot silently select a neighboring or enclosing node.
	costEntry := filepath.Join(repository, "stage1/cohere/typeaware/testdata/query_cost.ts")
	cost := h.build(stage0, "query-cost", costEntry, normal, false)
	costText, err := os.ReadFile(costEntry)
	if err != nil {
		t.Fatal(err)
	}
	badSource := h.write("bad-kind.ts", strings.Replace(string(costText), "'PrefixUnaryExpression'", "'Identifier'", 1))
	bad := h.build(stage0, "bad-kind", badSource, normal, false)
	got := h.run("bad-kind-run", exec.Command(bad, config, probe, "2"))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || !bytes.Contains(got.stderr, []byte("no exact Identifier node")) {
		t.Fatalf("exact lookup mismatch escaped: %v %s", got.err, got.stderr)
	}
	ignoredKind := h.overlay("ignored-kind", "bridge/tsgo/checker/program.go", `strings.TrimPrefix(candidate.Kind.String(), "Kind") == kind`, `kind != ""`)
	ignoredArchive := h.archive("ignored-kind-checker", ignoredKind, false)
	ignoredBinary := h.build(stage0, "ignored-kind-native", badSource, ignoredArchive, false)
	ignored := h.must("ignored-kind-run", exec.Command(ignoredBinary, config, probe, "2"))
	if len(ignored.stderr) != 0 {
		t.Fatal("kind-guard mutant did not finish cleanly")
	}
	t.Log("exact kind guard mutant: mismatch-refusal expectation catches exit 0")

	probeManifest := h.write("probe.manifest", probe+"\n")
	for round := 1; round <= 3; round++ {
		if round > 1 && os.Getenv("ADAMIC_TYPEAWARE_BENCH") != "1" {
			break
		}
		cmd := exec.Command(cost, config, probe, "10000")
		cmd.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
		nativeCost := h.must(fmt.Sprintf("query-cost-native-%d", round), cmd)
		direct := h.must(fmt.Sprintf("query-cost-go-%d", round), exec.Command(oracle, config, probeManifest, "--query-cost", "10000"))
		if !bytes.Equal(nativeCost.stdout, direct.stdout) {
			t.Fatal("query-cost result differs from direct Go")
		}
		t.Logf("query cost round %d native: %s; direct Go: %s; same output=%s", round, strings.TrimSpace(string(nativeCost.stderr)), strings.TrimSpace(string(direct.stderr)), strings.TrimSpace(string(direct.stdout)))
	}

	corpus := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if corpus == "" {
		t.Log("compiler corpus skipped: set ADAMIC_TYPESCRIPT_SOURCE (#xq2ecw6)")
		return
	}
	paths = corpusfiles.Upstream(t, corpus, compilerCommit, []string{"src/compiler"}, []string{"*.ts"})
	sort.Strings(paths)
	compilerManifest := h.write("compiler.manifest", strings.Join(paths, "\n")+"\n")
	compilerConfig := filepath.Join(corpus, "src/compiler/tsconfig.json")
	t.Logf("compiler corpus: %d files, original compiler tsconfig", len(paths))
	h.compare("compiler", oracle, binary, compilerConfig, compilerManifest)
	if os.Getenv("ADAMIC_TYPEAWARE_BENCH") == "1" {
		for _, corpus := range []struct{ name, config, manifest string }{{"compiler", compilerConfig, compilerManifest}, {"generated", config, manifest}} {
			for round := 1; round <= 3; round++ {
				// Alternate order to avoid always assigning one implementation the warm cache.
				binaries := []string{oracle, optimized}
				if round%2 == 0 {
					binaries[0], binaries[1] = binaries[1], binaries[0]
				}
				var expected []byte
				for _, binary := range binaries {
					command := exec.Command(binary, corpus.config, corpus.manifest, "--count")
					command.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
					r := h.must(fmt.Sprintf("bench-%s-%d-%s", corpus.name, round, filepath.Base(binary)), command)
					if expected != nil && !bytes.Equal(expected, r.stdout) {
						t.Fatal("timed finding counts differ")
					}
					expected = r.stdout
					var findings, queries int
					if _, err := fmt.Sscanf(string(r.stdout), "findings %d queries %d", &findings, &queries); err != nil {
						t.Fatal(err)
					}
					t.Logf("%s round %d %s: %.6fs %.2f findings/s findings=%d queries=%d; %s", corpus.name, round, filepath.Base(binary), r.elapsed.Seconds(), float64(findings)/r.elapsed.Seconds(), findings, queries, strings.TrimSpace(string(r.stderr)))
				}
			}
		}
	}
}

func TestPinnedTypeFlags(t *testing.T) {
	t.Parallel()
	actual := checker.TypeFlagsAny | checker.TypeFlagsNever | checker.TypeFlagsNumberLike | checker.TypeFlagsBigIntLike
	if actual != 334017 {
		t.Fatalf("update Adamic's mask and rerun cohere oracle: flags=%d", actual)
	}
}
