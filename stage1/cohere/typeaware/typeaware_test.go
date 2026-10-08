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
	"github.com/system-inc/adamic/internal/childguard"
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
}

func (h *harness) run(name string, command *exec.Cmd) result {
	h.t.Helper()
	h.next++
	if command.Dir == "" {
		command.Dir = h.repository
	}
	stem := filepath.Join(h.directory, fmt.Sprintf("%03d-%s", h.next, name))
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
	guard := typeawareRunGuard
	if name == "go" || strings.Contains(name, "build") || command.Path == "go" || filepath.Base(command.Path) == "go" || filepath.Base(command.Path) == "clang" || filepath.Base(command.Path) == "adamic" {
		guard = childguard.Options{FirstOutput: 10 * time.Minute, Stall: 2 * time.Minute, Ceiling: 30 * time.Minute}
	}
	runError := childguard.Run(command, guard)
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

// Shared builds and emissions finish before parallel, child-local readers run.
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
	optimized := ""
	if os.Getenv("ADAMIC_TYPEAWARE_BENCH") == "1" {
		optimized = h.build(stage0, "native", entry, normal, false)
	}
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

	// Parent owns every immutable build and emission until its parallel children finish.
	adapter, runner := h.nodeRuntime()
	client := h.bridgeClient("checker-client", normal)
	nodeEntry, emitted := h.javascriptEntry("main", entry, adapter)
	truth := h.must("generated-go", exec.Command(oracle, config, manifest))
	if !bytes.Contains(truth.stdout, []byte("is false instead.")) || !bytes.Contains(truth.stdout, []byte("is string instead.")) {
		t.Fatal("nonempty union/string controls missing")
	}
	var generatedFindings, generatedQueries int
	fmt.Sscanf(summary(truth.stdout), "findings %d queries %d", &generatedFindings, &generatedQueries)
	if generatedFindings < 20 || generatedQueries < 35 {
		t.Fatalf("generated coverage too small: %s", summary(truth.stdout))
	}
	t.Logf("generated coverage: %d files; %s", len(paths), summary(truth.stdout))

	probe := h.write("cost-probe.ts", "-1;\n")
	probeManifest := h.write("probe.manifest", probe+"\n")
	releasedEntry := filepath.Join(repository, "stage1/cohere/typeaware/testdata/released.ts")
	released := h.build(stage0, "released", releasedEntry, normal, false)
	releasedNode, releasedJS := h.javascriptEntry("released", releasedEntry, adapter)
	costEntry := filepath.Join(repository, "stage1/cohere/typeaware/testdata/query_cost.ts")
	cost := h.build(stage0, "query-cost", costEntry, normal, false)
	costText, err := os.ReadFile(costEntry)
	if err != nil {
		t.Fatal(err)
	}
	badSource := h.write("bad-kind.ts", strings.Replace(string(costText), "'PrefixUnaryExpression'", "'Identifier'", 1))
	bad := h.build(stage0, "bad-kind", badSource, normal, false)
	badNode, badJS := h.javascriptEntry("bad-kind", badSource, adapter)
	partsOnly := h.write("parts-unlinked.ts", "import { tsgoTypeParts } from 'adamic'; console.log(tsgoTypeParts(1, 'source.ts', 0, 1, 'Identifier'));\n")

	text, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/unary_minus.ts"))
	if err != nil {
		t.Fatal(err)
	}
	anchor := "this.parser.node(node.children[0] ?? panic('minus without operand'))"
	if strings.Count(string(text), anchor) != 1 {
		t.Fatal("wrong-node anchor changed")
	}
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
	wrongNode, wrongJS := h.javascriptEntry("wrong", mutantEntry, adapter)
	// Only this semantic mutant is the package's sanitized native canary.
	wrong := h.build(stage0, "wrong-node", mutantEntry, sanitized, true)

	type mutation struct {
		name, client, node, emitted string
		args                        []string
		message                     string
		native                      string
		want                        []byte
	}
	mutations := []mutation{{name: "wrong-node", client: client, node: wrongNode, emitted: wrongJS, args: []string{config, manifest}, native: wrong, want: truth.stdout}}
	staleOverlay := h.overlay("stale", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains released roots.")
	staleArchive := h.archive("stale-checker", staleOverlay, false)
	mutations = append(mutations, mutation{name: "stale", client: h.bridgeClient("stale-client", staleArchive), node: releasedNode, emitted: releasedJS, args: []string{config, probe}})
	for _, change := range []struct{ name, value, message string }{
		{"empty-frame", "", "checker returned no type parts"},
		{"missing-header", "x", "invalid checker type frame"},
		{"bad-length", "1\n-1\n", "invalid checker type length or flags"},
	} {
		overlay := h.overlay(change.name, "bridge/tsgo/archive/main.go", "*parts = buffer(answer)", "_ = answer; *parts = buffer("+strconv.Quote(change.value)+")")
		archive := h.archive(change.name+"-checker", overlay, false)
		mutations = append(mutations, mutation{name: change.name, client: h.bridgeClient(change.name+"-client", archive), node: nodeEntry, emitted: emitted, args: []string{config, manifest}, message: change.message})
	}
	ignoredKind := h.overlay("ignored-kind", "bridge/tsgo/checker/program.go", `strings.TrimPrefix(candidate.Kind.String(), "Kind") == kind`, `kind != ""`)
	ignoredArchive := h.archive("ignored-kind-checker", ignoredKind, false)
	mutations = append(mutations, mutation{name: "ignored-kind", client: h.bridgeClient("ignored-kind-client", ignoredArchive), node: badNode, emitted: badJS, args: []string{config, probe, "2"}})

	type corpusCase struct {
		name, config, manifest string
		want                   []byte
	}
	cases := []corpusCase{{"generated", config, manifest, truth.stdout}}
	if corpus := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"); corpus != "" {
		compilerPaths := corpusfiles.Upstream(t, corpus, compilerCommit, []string{"src/compiler"}, []string{"*.ts"})
		sort.Strings(compilerPaths)
		compilerManifest := h.write("compiler.manifest", strings.Join(compilerPaths, "\n")+"\n")
		compilerConfig := filepath.Join(corpus, "src/compiler/tsconfig.json")
		want := h.must("compiler-go", exec.Command(oracle, compilerConfig, compilerManifest))
		t.Logf("compiler corpus: %d files, original compiler tsconfig; %s", len(compilerPaths), summary(want.stdout))
		cases = append(cases, corpusCase{"compiler", compilerConfig, compilerManifest, want.stdout})
	} else {
		t.Log("compiler corpus skipped: set ADAMIC_TYPESCRIPT_SOURCE (#xq2ecw6)")
	}
	t.Run("cases", func(t *testing.T) {
		t.Parallel()
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				t.Parallel()
				child := h.child(t)
				for _, backend := range []string{"native", "Node", "emitted JavaScript"} {
					var got result
					if backend == "native" {
						got = child.must("native", exec.Command(binary, c.config, c.manifest))
					} else {
						source := nodeEntry
						if backend == "emitted JavaScript" {
							source = emitted
						}
						got = child.javascriptRun(backend, runner, source, client, c.config, c.manifest)
					}
					if got.err != nil || len(got.stderr) != 0 || !bytes.Equal(got.stdout, c.want) {
						t.Fatalf("%s mismatch: %v\n%s\nfirst differing byte %d", backend, got.err, got.stderr, firstDifference(got.stdout, c.want))
					}
					t.Logf("%s: %d identical finding bytes; %s", backend, len(got.stdout), summary(got.stdout))
				}
			})
		}
	})
	t.Run("mutants", func(t *testing.T) {
		t.Parallel()
		for _, m := range mutations {
			t.Run(m.name, func(t *testing.T) {
				t.Parallel()
				child := h.child(t)
				var nodeOutput []byte
				for _, backend := range []string{"Node", "emitted JavaScript", "sanitized native"} {
					if backend == "sanitized native" && m.native == "" {
						continue
					}
					var got result
					if backend == "sanitized native" {
						got = child.run("native", exec.Command(m.native, m.args...))
					} else {
						source := m.node
						if backend == "emitted JavaScript" {
							source = m.emitted
						}
						got = child.javascriptRun(backend, runner, source, m.client, m.args...)
					}
					if m.message != "" {
						code, ok := got.err.(*exec.ExitError)
						if !ok || code.ExitCode() != 70 || !bytes.Contains(got.stderr, []byte(m.message)) {
							t.Fatalf("%s mutant escaped on %s: %v %s", m.name, backend, got.err, got.stderr)
						}
					} else if got.err != nil || len(got.stderr) != 0 || (m.want != nil && bytes.Equal(got.stdout, m.want)) {
						t.Fatalf("%s mutant survived on %s: %v %s", m.name, backend, got.err, got.stderr)
					}
					if backend == "Node" {
						nodeOutput = got.stdout
					} else if !bytes.Equal(nodeOutput, got.stdout) {
						t.Fatalf("%s differs from mutated Node", backend)
					}
					t.Logf("%s mutant caught on %s; %s", m.name, backend, summary(got.stdout))
				}
			})
		}
	})
	t.Run("modes", func(t *testing.T) {
		t.Parallel()
		for _, mode := range []struct{ name, entry string }{{"unlinked", entry}, {"parts-unlinked", partsOnly}} {
			t.Run(mode.name, func(t *testing.T) {
				t.Parallel()
				child := h.child(t)
				refusal := child.run(mode.name, exec.Command(stage0, "build", mode.entry, "-o", filepath.Join(child.directory, mode.name)))
				if refusal.err == nil || !bytes.Contains(refusal.stderr, []byte("unlinked typescript-go library call")) {
					t.Fatal("unlinked type-parts call was accepted")
				}
			})
		}
		for _, mode := range []struct {
			name, binary, node, emitted, message string
			args                                 []string
		}{
			{"released", released, releasedNode, releasedJS, "invalid or released checker handle", []string{config, probe}},
			{"bad-kind", bad, badNode, badJS, "no exact Identifier node", []string{config, probe, "2"}},
		} {
			t.Run(mode.name, func(t *testing.T) {
				t.Parallel()
				child := h.child(t)
				for _, backend := range []string{"native", "Node", "emitted JavaScript"} {
					var got result
					if backend == "native" {
						got = child.run("native", exec.Command(mode.binary, mode.args...))
					} else {
						source := mode.node
						if backend == "emitted JavaScript" {
							source = mode.emitted
						}
						got = child.javascriptRun(backend, runner, source, client, mode.args...)
					}
					code, ok := got.err.(*exec.ExitError)
					if !ok || code.ExitCode() != 70 || !bytes.Contains(got.stderr, []byte(mode.message)) {
						t.Fatalf("%s escaped on %s: %v %s", mode.name, backend, got.err, got.stderr)
					}
				}
			})
		}
		rounds := 1
		if os.Getenv("ADAMIC_TYPEAWARE_BENCH") == "1" {
			rounds = 3
		}
		for round := 1; round <= rounds; round++ {
			t.Run(fmt.Sprintf("query-cost-%d", round), func(t *testing.T) {
				t.Parallel()
				child := h.child(t)
				cmd := exec.Command(cost, config, probe, "10000")
				cmd.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
				nativeCost := child.must("native", cmd)
				direct := child.must("go", exec.Command(oracle, config, probeManifest, "--query-cost", "10000"))
				if !bytes.Equal(nativeCost.stdout, direct.stdout) {
					t.Fatal("query-cost result differs from direct Go")
				}
				t.Logf("query cost round %d native: %s; direct Go: %s; same output=%s", round, strings.TrimSpace(string(nativeCost.stderr)), strings.TrimSpace(string(direct.stderr)), strings.TrimSpace(string(direct.stdout)))
			})
		}
	})
	if os.Getenv("ADAMIC_TYPEAWARE_BENCH") == "1" && len(cases) > 1 {
		t.Run("bench", func(t *testing.T) {
			t.Parallel()
			for _, c := range cases {
				for round := 1; round <= 3; round++ {
					t.Run(fmt.Sprintf("%s-%d", c.name, round), func(t *testing.T) {
						t.Parallel()
						child := h.child(t)
						binaries := []string{oracle, optimized}
						if round%2 == 0 {
							binaries[0], binaries[1] = binaries[1], binaries[0]
						}
						var expected []byte
						for _, binary := range binaries {
							cmd := exec.Command(binary, c.config, c.manifest, "--count")
							cmd.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
							got := child.must(filepath.Base(binary), cmd)
							if expected != nil && !bytes.Equal(expected, got.stdout) {
								t.Fatal("timed finding counts differ")
							}
							expected = got.stdout
							var findings, queries int
							if _, err := fmt.Sscanf(string(got.stdout), "findings %d queries %d", &findings, &queries); err != nil {
								t.Fatal(err)
							}
							t.Logf("%s round %d %s: %.6fs %.2f findings/s findings=%d queries=%d; %s", c.name, round, filepath.Base(binary), got.elapsed.Seconds(), float64(findings)/got.elapsed.Seconds(), findings, queries, strings.TrimSpace(string(got.stderr)))
						}
					})
				}
			}
		})
	}
}

func TestPinnedTypeFlags(t *testing.T) {
	t.Parallel()
	actual := checker.TypeFlagsAny | checker.TypeFlagsNever | checker.TypeFlagsNumberLike | checker.TypeFlagsBigIntLike
	if actual != 334017 {
		t.Fatalf("update Adamic's mask and rerun cohere oracle: flags=%d", actual)
	}
}
