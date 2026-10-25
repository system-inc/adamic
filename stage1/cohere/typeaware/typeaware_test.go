package typeaware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	goast "go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/corpusfiles"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/stage1/cohere/executable"
)

// The upstream commit fixes the compiler population; this is a lost-corpus
// guard, not the live case total used by the shard union.
const compilerCommit = "050880ce59e30b356b686bd3144efe24f875ebc8"
const compilerFilesAtPin = 77

func pinnedCompilerFiles(t *testing.T, corpus string) []string {
	t.Helper()
	paths := corpusfiles.Upstream(t, corpus, compilerCommit, []string{"src/compiler"}, []string{"*.ts"})
	if len(paths) != compilerFilesAtPin {
		t.Fatalf("compiler corpus at %s: %d files, want %d", compilerCommit, len(paths), compilerFilesAtPin)
	}
	return paths
}

// Fixture extraction is fixed by cohere's submodule commit, plus the explicit
// generated controls in these tests. It never enumerates repository files.
func pinnedFixtureFiles(t *testing.T, repository string, names []string) {
	t.Helper()
	roots := make([]string, len(names))
	for i, name := range names {
		roots[i] = "internal/lint/rules/typescript/" + name + "_test.go"
	}
	corpusfiles.Upstream(t, filepath.Join(repository, "cohere"), corpusfiles.CohereCommit, roots, []string{"*.go"})
}

type result struct {
	stdout, stderr []byte
	elapsed        time.Duration
	err            error
}
type harness struct {
	t                     *testing.T
	repository, directory string
	next                  int
	parallel              bool
	sixBuilds             bool
	sharedSetup           bool
	setupStarted          time.Time
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
	if h.parallel {
		environment := command.Env
		if environment == nil {
			environment = os.Environ()
		}
		command.Env = append(environment, "GOMAXPROCS=1")
	}
	started := time.Now()
	var runError error
	if h.sharedSetup {
		runError = typeAwareRunBuildCommand(command)
	} else {
		runError = typeAwareRunCommand(command)
	}
	elapsed := time.Since(started)
	h.t.Logf("phase command %s %.6fs", name, elapsed.Seconds())
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

// archive is the unchanged checker archive, plain (checker.a) or sanitized (checker-asan.a), a buildcache.GoBuildIn
// product as checkerMutantArchive's mutants are. Its key is the archive's own graph, which reaches no compiler code, so
// a compiler edit that leaves that graph alone finds it built; as an ordinary product of this package it was keyed by the
// test binary's whole recipe, which links the compiler, and rebuilt on every compiler edit (#gr5y0yf). Every reader asks
// it by this one request: this package's checks, the volume guard's and the volume agreement's. A check that plants a
// change in it reads checkerMutantArchive.
func (h *harness) archive(sanitize bool) string {
	h.t.Helper()
	if sanitize {
		return buildcache.GoBuildIn(h.t, "", "checker-asan.a", "./bridge/tsgo/archive", []string{"-buildmode=c-archive"}, checkerSanitized...)
	}
	return buildcache.GoBuildIn(h.t, "", "checker.a", "./bridge/tsgo/archive", []string{"-buildmode=c-archive"})
}

// unchangedArchive says whether archive is the unchanged checker archive, plain or sanitized (h.archive), rather than a
// mutant's, whose name is its own.
func unchangedArchive(archive string) bool {
	return filepath.Base(archive) == "checker.a" || filepath.Base(archive) == "checker-asan.a"
}

// lowerToC is entry's emitted C, by the same load, lowering and TSGo emission pipeline as stage zero.
func (h *harness) lowerToC(entry, name string) (string, error) {
	return h.lowerOverlayToC(entry, nil, name)
}

// lowerOverlayToC is lowerToC over entry with the overlay's files replaced by its source, read in place
// (load.LoadOverlay; a nil overlay loads the files as they are).
func (h *harness) lowerOverlayToC(entry string, overlay map[string]string, name string) (string, error) {
	started := time.Now()
	loaded, err := load.LoadOverlay([]string{entry}, overlay)
	h.t.Logf("phase load %s %.6fs", name, time.Since(started).Seconds())
	if err != nil {
		return "", err
	}
	loaded.EnableTSGo()
	started = time.Now()
	program, err := lower.Lower(context.Background(), loaded)
	h.t.Logf("phase lowering %s %.6fs", name, time.Since(started).Seconds())
	if err != nil {
		return "", err
	}
	started = time.Now()
	source, err := native.TSGoC(program)
	h.t.Logf("phase emission %s %.6fs", name, time.Since(started).Seconds())
	return source, err
}

func (h *harness) lowered(entry, name string) string {
	h.t.Helper()
	build := func(_ string) (string, error) { return h.lowerToC(entry, name) }
	var source string
	var err error
	if strings.HasPrefix(entry, h.repository+string(os.PathSeparator)) {
		key, keyErr := typeAwareLoweredKey(h.repository, entry)
		if keyErr != nil {
			h.t.Fatal(keyErr)
		}
		source, err = cachedTypeAwareProduct(key, build)
	} else {
		// Generated mutants belong to their check; scratch paths cannot be gate products.
		source, err = build("")
	}
	if err != nil {
		h.t.Fatal(err)
	}
	return source
}

// loweredMutant is lowered for a mutant entry written outside the repository, its emitted C a product under key. The
// key names no path and needs none: every file the entry reads is one typeAwareSharedInputs names, or one this
// package's tests write from those by an edit their own source holds, and emission names no path. Lowering a six
// mutant took 16 to 20 s of its shard's own time under load (Pupa, load 13), the native already a product.
func (h *harness) loweredMutant(entry, name, key string) string {
	h.t.Helper()
	source, err := cachedTypeAwareProduct(key, func(_ string) (string, error) { return h.lowerToC(entry, name) })
	if err != nil {
		h.t.Fatal(err)
	}
	return source
}

// build uses the same load, lowering and TSGo emission pipeline as stage zero.
// Only the completed C string is shared, because emission mutates its IR.
func (h *harness) build(_ string, name, entry, archive string, sanitize bool) string {
	h.t.Helper()
	path := filepath.Join(h.directory, name)
	source := h.lowered(entry, name)
	var err error
	compile := func(output string) error {
		started := time.Now()
		err := typeAwareNativeBuild(source, output, archive, sanitize)
		h.t.Logf("phase clang %s %.6fs", name, time.Since(started).Seconds())
		return err
	}
	// Only unchanged repository programs can be reused by another test. A mutant
	// runs once and stays in its child directory, avoiding a retained binary copy.
	if !strings.HasPrefix(entry, h.repository+string(os.PathSeparator)) || !unchangedArchive(archive) {
		if err := compile(path); err != nil {
			h.t.Fatal(err)
		}
		return path
	}
	key, err := typeAwareNativeKey("binary", source, archive, sanitize)
	if err != nil {
		h.t.Fatal(err)
	}
	binary, err := cachedTypeAwareProduct(key, func(directory string) (string, error) {
		binary := filepath.Join(directory, "native")
		return binary, compile(binary)
	})
	if err != nil {
		h.t.Fatal(err)
	}
	// Callers may remove their binaries. Give them copies, preserving the product.
	data, err := os.ReadFile(binary)
	if err != nil {
		h.t.Fatal(err)
	}
	if err := executable.Write(path, data); err != nil {
		h.t.Fatal(err)
	}

	return path
}

// mutantNative is a mutant's native as a product, for a mutant whose compile would otherwise fill its shard's own time
// (wrong-node's ASan link took 22.0 s of TestTypeAwareAgreementAndMutants_003's 24.4 s, and facts-length's 25.6 s of
// TestSixRuleAgreementAndMutants_028's 26.1 s, alone on one core). Its key is typeAwareNativeKey's, as h.build's
// cached natives' is: the emitted C's hash (emission names no path, so the same mutant lowered from two directories is
// the same C) and the archive's. Its TestProduct_ makes it ahead. Every six mutant's native is one (sixMutants, and
// declaration-source's); any other mutant still compiles inside its own check, as h.build does.
func (h *harness) mutantNative(name, entry, archive string, sanitize bool) string {
	h.t.Helper()
	return h.mutantNativeOf(name, h.lowered(entry, name), archive, sanitize)
}

// mutantNativeOf is mutantNative for a mutant whose C is already emitted (loweredMutant).
func (h *harness) mutantNativeOf(name, source, archive string, sanitize bool) string {
	h.t.Helper()
	key, err := typeAwareNativeKey("mutant native "+name, source, archive, sanitize)
	if err != nil {
		h.t.Fatal(err)
	}
	binary, err := cachedTypeAwareProduct(key, func(directory string) (string, error) {
		binary := filepath.Join(directory, "native")
		started := time.Now()
		err := typeAwareNativeBuild(source, binary, archive, sanitize)
		h.t.Logf("phase clang %s %.6fs", name, time.Since(started).Seconds())
		return binary, err
	})
	if err != nil {
		h.t.Fatal(err)
	}
	return binary
}

// typeAwareLoweredKey keys an entry's emitted C by the entry's path in the repository, never where the checkout sits:
// a key carries no machine path, or a runner's checkout reads none of the builder's products.
func typeAwareLoweredKey(repository, entry string) (string, error) {
	relative, err := filepath.Rel(repository, entry)
	if err != nil {
		return "", err
	}
	return "C " + filepath.ToSlash(relative), nil
}

// typeAwareNativeKey keys a native by kind, the emitted C's hash, the archive's bytes' hash and sanitize. The archive
// is a product whose path names the cache directory, and so the machine, it sits in: by its bytes, the same archive
// keys the same native wherever it is cached (tsgo's archive rebuilds byte for byte, Builder's Linux proof, Oct 12).
func typeAwareNativeKey(kind, source, archive string, sanitize bool) (string, error) {
	digest, err := typeAwareArchiveDigest(archive)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s %x archive=%s sanitize=%t", kind, sha256.Sum256([]byte(source)), digest, sanitize), nil
}

// typeAwareArchiveDigests are the archives' SHA256s, read once per process each: a product never changes in place.
var typeAwareArchiveDigests sync.Map

func typeAwareArchiveDigest(archive string) (string, error) {
	if digest, ok := typeAwareArchiveDigests.Load(archive); ok {
		return digest.(string), nil
	}
	data, err := os.ReadFile(archive)
	if err != nil {
		return "", err
	}
	digest := fmt.Sprintf("%x", sha256.Sum256(data))
	typeAwareArchiveDigests.Store(archive, digest)
	return digest, nil
}

// The keys of a native and of an emitted C carry no machine path: one product keys alike from two checkouts and two
// cache directories, by name key and by declared files. Mutants: the archive's path back in the native's key; the
// entry's absolute path back in the C's.
//
// A declared key is the name key's inputs plus the declared files' contents, so equal name keys over equal Files, in
// one repository, are equal declared keys. The test compares those, never buildcache.Key itself: Key hashed every file
// under typeAwareSharedInputs' Files (internal, stage1/typescript, cohere/TypeScript/tsc) four times: 69.5 s on
// typeaware#2/7 for chg_rcdxaf56 against a two-day p90 of 43.8 s (Oct 19), a red on a change that never touched it.
func TestTypeAwareKeysCarryNoMachinePath(t *testing.T) {
	t.Parallel()
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	var natives, lowered []string
	for _, machine := range []string{t.TempDir(), t.TempDir()} {
		// Where the archive product sits under this machine's ADAMIC_BUILD_CACHE_DIR, and where its checkout is.
		archive := filepath.Join(machine, "cache", "local", "d91f8adc3a72", "checker.a")
		if err := os.MkdirAll(filepath.Dir(archive), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(archive, []byte("the same archive bytes on both machines"), 0644); err != nil {
			t.Fatal(err)
		}
		native, err := typeAwareNativeKey("binary", "int main(void) { return 0; }\n", archive, true)
		if err != nil {
			t.Fatal(err)
		}
		checkout := filepath.Join(machine, "checkout")
		c, err := typeAwareLoweredKey(checkout, filepath.Join(checkout, "stage1/cohere/typeaware/main.ts"))
		if err != nil {
			t.Fatal(err)
		}
		natives, lowered = append(natives, native), append(lowered, c)
	}
	for _, keys := range []struct {
		name string
		keys []string
	}{{"native", natives}, {"emitted C", lowered}} {
		first, second := typeAwareSharedInputs(keys.keys[0]), typeAwareSharedInputs(keys.keys[1])
		if a, b := buildcache.NameKey(repository, first), buildcache.NameKey(repository, second); a != b {
			t.Fatalf("%s: name key %s on one machine, %s on the other (%q, %q)", keys.name, a[:12], b[:12], keys.keys[0], keys.keys[1])
		}
		if !slices.Equal(first.Files, second.Files) {
			t.Fatalf("%s: declared files %q on one machine, %q on the other", keys.name, first.Files, second.Files)
		}
		t.Logf("%s: one name key (%s) and the same declared files from two machines", keys.name, buildcache.NameKey(repository, first)[:12])
	}
}

// typeAwareWrongNodeNative is the wrong-node mutant's native over archive: main.ts asking the unary expression's type
// rather than its operand's. It still compiles and finishes; the independent production-rule findings must disagree.
// The mutant is main.ts as the repository has it, with unary_minus.ts overlaid in place (load.LoadOverlay), so every
// import keeps its relative path. Written to a scratch directory with its imports made absolute, its C named the
// checkout, so typeAwareNativeKey moved with every checkout and Loom's key check refused each tree (#nwczaa5). The C is
// a product keyed by name, as a six mutant's is (loweredMutant), and refuses to be one if it names the checkout.
func typeAwareWrongNodeNative(h *harness, archive string, sanitize bool) string {
	h.t.Helper()
	source, err := cachedTypeAwareProduct("C typeaware mutant wrong-node", func(_ string) (string, error) {
		mutant, err := typeAwareWrongNodeSource(h.repository)
		if err != nil {
			return "", err
		}
		unary := filepath.Join(h.repository, "stage1/cohere/typeaware/unary_minus.ts")
		source, err := h.lowerOverlayToC(filepath.Join(h.repository, "stage1/cohere/typeaware/main.ts"), map[string]string{unary: mutant}, "wrong-node")
		if err != nil {
			return "", err
		}
		return source, typeAwareCNamesNoCheckout(source, h.repository)
	})
	if err != nil {
		h.t.Fatal(err)
	}
	return h.mutantNativeOf("wrong-node", source, archive, sanitize)
}

// typeAwareWrongNodeSource is unary_minus.ts with the wrong-node mutation: the type asked of the whole expression.
func typeAwareWrongNodeSource(repository string) (string, error) {
	text, err := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/unary_minus.ts"))
	if err != nil {
		return "", err
	}
	anchor := "this.parser.node(node.children[0] ?? panic('minus without operand'))"
	if strings.Count(string(text), anchor) != 1 {
		return "", fmt.Errorf("wrong-node anchor changed")
	}
	return strings.Replace(string(text), anchor, "this.parser.node(index)", 1), nil
}

// typeAwareCNamesNoCheckout refuses emitted C that names the checkout it was lowered in: a product keyed by its bytes
// would then key differently in every checkout.
func typeAwareCNamesNoCheckout(source, repository string) error {
	if strings.Contains(source, filepath.ToSlash(repository)) || strings.Contains(source, repository) {
		return fmt.Errorf("emitted C names the checkout %s, so its key would move with the checkout", repository)
	}
	return nil
}

// The guard on wrong-node's C can fail: C naming the checkout is refused, C naming only relative paths isn't.
func TestTypeAwareCNamesNoCheckoutCanFail(t *testing.T) {
	t.Parallel()
	repository := filepath.Join(t.TempDir(), "adamic")
	named := "#include \"" + filepath.ToSlash(filepath.Join(repository, "stage1/typescript/scanner.ts")) + "\"\n"
	if typeAwareCNamesNoCheckout(named, repository) == nil {
		t.Fatal("C naming the checkout passed")
	}
	if err := typeAwareCNamesNoCheckout("/* ../../typescript/scanner.ts */\n", repository); err != nil {
		t.Fatal(err)
	}
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

// The optional external corpus is never fetched by a test.
// Products are prepared once. ADAMIC_TEST_SHARD=i/n selects zero-based parallel
// check shards; unset runs every complete corpus and mutant check.
const testTypeAwareAgreementAndMutantsShards = 21

func prepareTypeAwareAgreementAndMutants(t *testing.T) *typeAwarePlan {
	started := time.Now()
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := filepath.Join(productDirectory, "typeaware-plan")
	if err := os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	if artifacts := os.Getenv("ADAMIC_TYPEAWARE_ARTIFACTS"); artifacts != "" {
		directory, err = filepath.Abs(artifacts)
		if err != nil {
			t.Fatal(err)
		}
		if err = os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
	}
	h := &harness{t: t, repository: repository, directory: directory, sixBuilds: true, setupStarted: started}
	traceGroup(t)
	stage0 := "" // Native builds use the in-process compiler; only refusal shards need stage zero.
	normal := func(h *harness) string { return typeAwareArchive(h, "checker", false) }
	sanitized := func(h *harness) string { return typeAwareArchive(h, "checker-asan", true) }
	entry := filepath.Join(repository, "stage1/cohere/typeaware/main.ts")
	binary := func(h *harness) string { return typeAwareBuild(h, stage0, "native-asan", entry, sanitized(h), true) }
	optimized := func(h *harness) string { return typeAwareBuild(h, stage0, "native", entry, normal(h), false) }
	// Common products are ready before any case deadline starts.
	typeAwareStage0(h)
	readyBinary, readyOptimized := binary(h), optimized(h)
	binary = func(*harness) string { return readyBinary }
	optimized = func(*harness) string { return readyOptimized }
	readyNormal, readySanitized := normal(h), sanitized(h)
	normal = func(*harness) string { return readyNormal }
	sanitized = func(*harness) string { return readySanitized }
	oracle := typeAwareOracle(h, "typeaware")
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	var generatedPathsForUnion []string
	var shards []sixShard
	var expected []string
	rounds := 1
	if os.Getenv("ADAMIC_TYPEAWARE_BENCH") == "1" {
		rounds = 3
	}
	finish := func(compiler []string, benchmark bool) *typeAwarePlan {
		unsplit := typeAwareUnsplitIDs(generatedPathsForUnion, compiler, rounds, benchmark)
		if err := sixUnion(unsplit, []sixShard{{name: "original-work", ids: expected}}); err != nil {
			t.Fatal(err)
		}
		return newTypeAwarePlan(t, h, unsplit, shards, testTypeAwareAgreementAndMutantsShards)
	}
	add := func(name string, ids []string, run func(*harness)) {
		shards = append(shards, sixShard{name: name, ids: ids, run: run})
	}

	// Extract the production rule's table sources without duplicating its verdicts.
	pinnedFixtureFiles(t, repository, []string{"no_unsafe_unary_minus"})
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
	truth := h.must("generated-go", exec.Command(oracle, config, manifest))
	generatedPaths := append([]string(nil), paths...)
	generatedPathsForUnion = generatedPaths
	expected = append(expected, sixIDs("agreement/generated", generatedPaths)...)
	add("agreement/generated", sixIDs("agreement/generated", generatedPaths), func(h *harness) {
		t := h.t
		truth := h.compare("generated", oracle, binary(h), config, manifest)
		if err := sixFindingUnion(truth.stdout, generatedPaths); err != nil {
			t.Fatal(err)
		}
		if !bytes.Contains(truth.stdout, []byte("is false instead.")) || !bytes.Contains(truth.stdout, []byte("is string instead.")) {
			t.Fatal("nonempty union/string controls missing")
		}
		var generatedFindings, generatedQueries int
		fmt.Sscanf(summary(truth.stdout), "findings %d queries %d", &generatedFindings, &generatedQueries)
		if generatedFindings < 20 || generatedQueries < 35 {
			t.Fatalf("generated coverage too small: %s", summary(truth.stdout))
		}
		t.Logf("generated coverage: %d files", len(generatedPaths))

	})
	// No bridge calls can reach an ordinary backend without explicit linkage.
	expected = append(expected, "refusal/unlinked")
	add("refusal/unlinked", []string{"refusal/unlinked"}, func(h *harness) {
		stage0 := typeAwareStage0(h)
		t := h.t
		refusal := h.run("unlinked", exec.Command(stage0, "build", entry, "-o", filepath.Join(directory, "unlinked")))
		if refusal.err == nil || !bytes.Contains(refusal.stderr, []byte("unlinked typescript-go library call")) {
			t.Fatal("unlinked type-parts call was accepted")
		}
	})

	partsOnly := h.write("parts-unlinked.ts", "import { tsgoTypeParts } from 'adamic'; console.log(tsgoTypeParts(1, 'source.ts', 0, 1, 'Identifier'));\n")
	expected = append(expected, "refusal/parts-unlinked")
	add("refusal/parts-unlinked", []string{"refusal/parts-unlinked"}, func(h *harness) {
		stage0 := typeAwareStage0(h)
		t := h.t
		refusal := h.run("parts-unlinked", exec.Command(stage0, "build", partsOnly, "-o", filepath.Join(directory, "parts-unlinked")))
		if refusal.err == nil || !bytes.Contains(refusal.stderr, []byte("unlinked typescript-go library call")) {
			t.Fatal("unlinked type-parts call was accepted")
		}
	})

	// Asking the unary expression's type rather than its operand still compiles
	// and finishes; the independent production-rule findings must disagree.
	expected = append(expected, sixIDs("mutant/wrong-node", generatedPaths)...)
	add("mutant/wrong-node", sixIDs("mutant/wrong-node", generatedPaths), func(h *harness) {
		t := h.t
		wrong := typeAwareSetup(h, "wrong-node", func() string { return typeAwareWrongNodeNative(h, normal(h), true) })
		observed := h.must("wrong-node-run", exec.Command(wrong, config, manifest))
		if len(observed.stderr) != 0 || bytes.Equal(observed.stdout, truth.stdout) {
			t.Fatal("wrong-node mutant was not caught by findings alone")
		}
		t.Logf("wrong-node type mutant: cohere byte oracle catches byte %d; %s", firstDifference(observed.stdout, truth.stdout), summary(observed.stdout))
	})

	probe := h.write("cost-probe.ts", "-1;\n")
	releasedEntry := filepath.Join(repository, "stage1/cohere/typeaware/testdata/released.ts")
	expected = append(expected, []string{"refusal/released"}...)
	add("refusal/released", []string{"refusal/released"}, func(h *harness) {
		t := h.t
		released := typeAwareBuild(h, stage0, "released", releasedEntry, normal(h), false)
		stale := h.run("released-run", exec.Command(released, config, probe))
		if code, ok := stale.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(stale.stderr) != "adamic: panic: invalid or released checker handle\n" {
			t.Fatalf("released query escaped: %v %s", stale.err, stale.stderr)
		}
		t.Log("released program queried: native panic 70 with invalid or released checker handle")
	})
	expected = append(expected, []string{"mutant/stale"}...)
	add("mutant/stale", []string{"mutant/stale"}, func(h *harness) {
		t := h.t
		staleArchive := typeAwareMutantArchive(h, "retained-handle")
		staleBinary := typeAwareBuild(h, stage0, "stale-native", releasedEntry, staleArchive, false)
		stale := h.must("stale-mutant-run", exec.Command(staleBinary, config, probe))
		if len(stale.stderr) != 0 {
			t.Fatal("stale mutant did not finish cleanly")
		}
		t.Log("released registry deletion mutant: stale-query expectation catches exit 0 instead of 70")
	})

	// Bad frame mutants must fail explicitly before any finding can be invented.
	// Each frame's value is planted by its checker mutant (checker_mutants_test.go).
	for _, change := range []struct{ name, message string }{
		{"empty-frame", "checker returned no type parts"},
		{"missing-header", "invalid checker type frame"},
		{"bad-length", "invalid checker type length or flags"},
	} {
		expected = append(expected, sixIDs("mutant/"+change.name, generatedPaths)...)
		add("mutant/"+change.name, sixIDs("mutant/"+change.name, generatedPaths), func(h *harness) {
			t := h.t
			archive := typeAwareMutantArchive(h, change.name)
			malformed := typeAwareBuild(h, stage0, change.name+"-native", entry, archive, false)
			got := h.run(change.name+"-run", exec.Command(malformed, config, manifest))
			if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || !bytes.Contains(got.stderr, []byte(change.message)) {
				t.Fatalf("%s escaped: %v %s", change.name, got.err, got.stderr)
			}
			t.Logf("%s mutant: panic 70, %s", change.name, change.message)
		})

	}

	// A kind/span mismatch cannot silently select a neighboring or enclosing node.
	costEntry := filepath.Join(repository, "stage1/cohere/typeaware/testdata/query_cost.ts")
	cost := func(h *harness) string { return typeAwareBuild(h, stage0, "query-cost", costEntry, normal(h), false) }
	readyCost := cost(h)
	cost = func(*harness) string { return readyCost }
	costText, err := os.ReadFile(costEntry)
	if err != nil {
		t.Fatal(err)
	}
	badSource := h.write("bad-kind.ts", strings.Replace(string(costText), "'PrefixUnaryExpression'", "'Identifier'", 1))
	expected = append(expected, []string{"refusal/bad-kind"}...)
	add("refusal/bad-kind", []string{"refusal/bad-kind"}, func(h *harness) {
		t := h.t
		bad := typeAwareBuild(h, stage0, "bad-kind", badSource, normal(h), false)
		got := h.run("bad-kind-run", exec.Command(bad, config, probe, "2"))
		if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || !bytes.Contains(got.stderr, []byte("no exact Identifier node")) {
			t.Fatalf("exact lookup mismatch escaped: %v %s", got.err, got.stderr)
		}
	})
	expected = append(expected, []string{"mutant/ignored-kind"}...)
	add("mutant/ignored-kind", []string{"mutant/ignored-kind"}, func(h *harness) {
		t := h.t
		ignoredArchive := typeAwareMutantArchive(h, "ignored-kind")
		ignoredBinary := typeAwareBuild(h, stage0, "ignored-kind-native", badSource, ignoredArchive, false)
		ignored := h.must("ignored-kind-run", exec.Command(ignoredBinary, config, probe, "2"))
		if len(ignored.stderr) != 0 {
			t.Fatal("kind-guard mutant did not finish cleanly")
		}
		t.Log("exact kind guard mutant: mismatch-refusal expectation catches exit 0")
	})

	probeManifest := h.write("probe.manifest", probe+"\n")
	for round := 1; round <= 3; round++ {

		id := fmt.Sprintf("cost/query-%d", round)
		ids := []string{id}
		if round > rounds {
			ids = nil
		}
		expected = append(expected, ids...)
		add(id, ids, func(h *harness) {
			if round > rounds {
				h.t.Skip("set ADAMIC_TYPEAWARE_BENCH=1")
			}
			t := h.t
			cmd := exec.Command(cost(h), config, probe, "10000")
			cmd.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
			nativeCost := h.must(fmt.Sprintf("query-cost-native-%d", round), cmd)
			direct := h.must(fmt.Sprintf("query-cost-go-%d", round), exec.Command(oracle, config, probeManifest, "--query-cost", "10000"))
			if !bytes.Equal(nativeCost.stdout, direct.stdout) {
				t.Fatal("query-cost result differs from direct Go")
			}
			t.Logf("query cost round %d native: %s; direct Go: %s; same output=%s", round, strings.TrimSpace(string(nativeCost.stderr)), strings.TrimSpace(string(direct.stderr)), strings.TrimSpace(string(direct.stdout)))
		})

	}

	corpus := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if corpus == "" {
		t.Log("compiler corpus skipped: set ADAMIC_TYPESCRIPT_SOURCE (#xq2ecw6)")
	}
	paths = nil
	if corpus != "" {
		paths = pinnedCompilerFiles(t, corpus)
	}
	sort.Strings(paths)
	compilerManifest := h.write("compiler.manifest", strings.Join(paths, "\n")+"\n")
	compilerConfig := filepath.Join(corpus, "src/compiler/tsconfig.json")
	t.Logf("compiler corpus: %d files, original compiler tsconfig", len(paths))
	expected = append(expected, sixIDs("agreement/compiler", paths)...)
	add("agreement/compiler", sixIDs("agreement/compiler", paths), func(h *harness) {
		if len(paths) == 0 {
			h.t.Skip("set ADAMIC_TYPESCRIPT_SOURCE")
		}
		truth := h.compare("compiler", oracle, binary(h), compilerConfig, compilerManifest)
		if err := sixFindingUnion(truth.stdout, paths); err != nil {
			h.t.Fatal(err)
		}
	})
	{
		for _, corpus := range []struct{ name, config, manifest string }{{"compiler", compilerConfig, compilerManifest}, {"generated", config, manifest}} {
			for round := 1; round <= 3; round++ {
				files := generatedPaths
				if corpus.name == "compiler" {
					files = paths
				}
				name := fmt.Sprintf("bench/%s-%d", corpus.name, round)
				ids := sixIDs(name, files)
				if rounds != 3 || len(paths) == 0 {
					ids = nil
				}
				expected = append(expected, ids...)
				add(name, ids, func(h *harness) {
					if rounds != 3 || len(paths) == 0 {
						h.t.Skip("set ADAMIC_TYPESCRIPT_SOURCE and ADAMIC_TYPEAWARE_BENCH=1")
					}
					t := h.t
					// Alternate order to avoid always assigning one implementation the warm cache.
					binaries := []string{oracle, optimized(h)}
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
				})

			}
		}
	}
	return finish(paths, rounds == 3 && len(paths) != 0)
}

func TestPinnedTypeFlags(t *testing.T) {
	t.Parallel()
	actual := checker.TypeFlagsAny | checker.TypeFlagsNever | checker.TypeFlagsNumberLike | checker.TypeFlagsBigIntLike
	if actual != 334017 {
		t.Fatalf("update Adamic's mask and rerun cohere oracle: flags=%d", actual)
	}
}

func typeAwareUnsplitIDs(generated, compiler []string, rounds int, benchmark bool) []string {
	ids := []string{"refusal/unlinked", "refusal/parts-unlinked", "refusal/released", "mutant/stale", "refusal/bad-kind", "mutant/ignored-kind"}
	for _, name := range []string{"agreement/generated", "mutant/wrong-node", "mutant/empty-frame", "mutant/missing-header", "mutant/bad-length"} {
		ids = append(ids, sixIDs(name, generated)...)
	}
	for round := 1; round <= rounds; round++ {
		ids = append(ids, fmt.Sprintf("cost/query-%d", round))
	}
	ids = append(ids, sixIDs("agreement/compiler", compiler)...)
	if benchmark {
		for round := 1; round <= 3; round++ {
			ids = append(ids, sixIDs(fmt.Sprintf("bench/compiler-%d", round), compiler)...)
			ids = append(ids, sixIDs(fmt.Sprintf("bench/generated-%d", round), generated)...)
		}
	}
	return ids
}
