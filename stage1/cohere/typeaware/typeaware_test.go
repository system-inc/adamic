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

func pinnedCompilerFiles(test *testing.T, corpus string) []string {
	test.Helper()
	paths := corpusfiles.Upstream(test, corpus, compilerCommit, []string{"src/compiler"}, []string{"*.ts"})
	if len(paths) != compilerFilesAtPin {
		test.Fatalf("compiler corpus at %s: %d files, want %d", compilerCommit, len(paths), compilerFilesAtPin)
	}
	return paths
}

// Fixture extraction is fixed by cohere's submodule commit, plus the explicit
// generated controls in these tests. It never enumerates repository files.
func pinnedFixtureFiles(test *testing.T, repository string, names []string) {
	test.Helper()
	roots := make([]string, len(names))
	for i, name := range names {
		roots[i] = "internal/lint/rules/typescript/" + name + "_test.go"
	}
	corpusfiles.Upstream(test, filepath.Join(repository, "cohere"), corpusfiles.CohereCommit, roots, []string{"*.go"})
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
	out, createError := os.Create(stem + ".stdout")
	if createError != nil {
		h.t.Fatal(createError)
	}
	defer out.Close()
	report, createError := os.Create(stem + ".stderr")
	if createError != nil {
		h.t.Fatal(createError)
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
	stdout, readError := os.ReadFile(out.Name())
	if readError != nil {
		h.t.Fatal(readError)
	}
	stderr, readError := os.ReadFile(report.Name())
	if readError != nil {
		h.t.Fatal(readError)
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
	if writeError := os.WriteFile(path, []byte(text), 0644); writeError != nil {
		h.t.Fatal(writeError)
	}
	return path
}

// archive is the unchanged checker archive, plain (checker.a) or sanitized (checker-asan.a), a buildcache.GoBuildIn
// product as checkerMutantArchive's mutants are. Its key is the archive's own graph, which reaches no compiler code, so
// a compiler edit that leaves that graph alone finds it built; as an ordinary product of this package it was keyed by
// the test binary's whole recipe, which links the compiler, and rebuilt on every compiler edit (#gr5y0yf). Every reader
// asks it by this one request: this package's checks, the volume guard's and the volume agreement's. A check that
// plants a change in it reads checkerMutantArchive.
func (h *harness) archive(sanitize bool) string {
	h.t.Helper()
	if sanitize {
		return buildcache.GoBuildIn(
			h.t,
			"",
			"checker-asan.a",
			"./bridge/tsgo/archive",
			[]string{"-buildmode=c-archive"},
			checkerSanitized...,
		)
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
	loaded, loadOverlayError := load.LoadOverlay([]string{entry}, overlay)
	h.t.Logf("phase load %s %.6fs", name, time.Since(started).Seconds())
	if loadOverlayError != nil {
		return "", loadOverlayError
	}
	loaded.EnableTSGo()
	started = time.Now()
	program, lowerError := lower.Lower(context.Background(), loaded)
	h.t.Logf("phase lowering %s %.6fs", name, time.Since(started).Seconds())
	if lowerError != nil {
		return "", lowerError
	}
	started = time.Now()
	source, tsGoCError := native.TSGoC(program)
	h.t.Logf("phase emission %s %.6fs", name, time.Since(started).Seconds())
	return source, tsGoCError
}

func (h *harness) lowered(entry, name string) string {
	h.t.Helper()
	build := func(_ string) (string, error) { return h.lowerToC(entry, name) }
	var source string
	var err error
	if strings.HasPrefix(entry, h.repository+string(os.PathSeparator)) {
		key, keyError := typeAwareLoweredKey(h.repository, entry)
		if keyError != nil {
			h.t.Fatal(keyError)
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
	source, cachedTypeAwareProductError := cachedTypeAwareProduct(
		key,
		func(_ string) (string, error) { return h.lowerToC(entry, name) },
	)
	if cachedTypeAwareProductError != nil {
		h.t.Fatal(cachedTypeAwareProductError)
	}
	return source
}

// typeAwareBadKindSource is testdata/query_cost.ts asking for an Identifier where it asks for a PrefixUnaryExpression,
// so the exact-kind lookup refuses.
func typeAwareBadKindSource(repository string) (string, error) {
	text, readError := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/testdata/query_cost.ts"))
	if readError != nil {
		return "", readError
	}
	if !strings.Contains(string(text), "'PrefixUnaryExpression'") {
		return "", fmt.Errorf("bad-kind anchor changed")
	}
	return strings.Replace(string(text), "'PrefixUnaryExpression'", "'Identifier'", 1), nil
}

// typeAwareBadKindNative is the bad-kind program's native over archive as a product. The program is query_cost.ts with
// typeAwareBadKindSource overlaid in place (load.LoadOverlay), as wrong-node's is, so its C names no checkout, and that
// C is a product keyed by name that refuses to be one if it does (#nwczaa5). The bad-kind refusal and the ignored-kind
// mutant fetch what their TestProduct_ built ahead instead of lowering and linking inside their shards.
func typeAwareBadKindNative(h *harness, name, archive string) string {
	h.t.Helper()
	return typeAwareSetup(h, name, func() string {
		source, cachedTypeAwareProductError := cachedTypeAwareProduct("C typeaware mutant bad-kind", func(
			_ string,
		) (string, error) {
			mutant, typeAwareBadKindSourceError := typeAwareBadKindSource(h.repository)
			if typeAwareBadKindSourceError != nil {
				return "", typeAwareBadKindSourceError
			}
			entry := filepath.Join(h.repository, "stage1/cohere/typeaware/testdata/query_cost.ts")
			source, lowerOverlayToCError := h.lowerOverlayToC(entry, map[string]string{entry: mutant}, "bad-kind")
			if lowerOverlayToCError != nil {
				return "", lowerOverlayToCError
			}
			return source, typeAwareCNamesNoCheckout(source, h.repository)
		})
		if cachedTypeAwareProductError != nil {
			h.t.Fatal(cachedTypeAwareProductError)
		}
		return h.mutantNativeOf(name, source, archive, false)
	})
}

// build uses the same load, lowering and TSGo emission pipeline as stage zero.
// Only the completed C string is shared, because emission mutates its IR.
func (h *harness) build(_ string, name, entry, archive string, sanitize bool) string {
	h.t.Helper()
	path := filepath.Join(h.directory, name)
	source := h.lowered(entry, name)
	var typeAwareNativeKeyError error
	compile := func(output string) error {
		started := time.Now()
		typeAwareNativeBuildError := typeAwareNativeBuild(source, output, archive, sanitize)
		h.t.Logf("phase clang %s %.6fs", name, time.Since(started).Seconds())
		return typeAwareNativeBuildError
	}
	// Only unchanged repository programs can be reused by another test. A mutant
	// runs once and stays in its child directory, avoiding a retained binary copy.
	if !strings.HasPrefix(entry, h.repository+string(os.PathSeparator)) || !unchangedArchive(archive) {
		if compileError := compile(path); compileError != nil {
			h.t.Fatal(compileError)
		}
		return path
	}
	key, typeAwareNativeKeyError := typeAwareNativeKey("binary", source, archive, sanitize)
	if typeAwareNativeKeyError != nil {
		h.t.Fatal(typeAwareNativeKeyError)
	}
	binary, cachedTypeAwareProductError := cachedTypeAwareProduct(key, func(directory string) (string, error) {
		binary := filepath.Join(directory, "native")
		return binary, compile(binary)
	})
	if cachedTypeAwareProductError != nil {
		h.t.Fatal(cachedTypeAwareProductError)
	}
	// Callers may remove their binaries. Give them copies, preserving the product.
	data, readError := os.ReadFile(binary)
	if readError != nil {
		h.t.Fatal(readError)
	}
	if writeError := executable.Write(path, data); writeError != nil {
		h.t.Fatal(writeError)
	}

	return path
}

// mutantNative is a mutant's native as a product, for a mutant whose compile would otherwise fill its shard's own time
// (wrong-node's ASan link took 22.0 s of TestTypeAwareAgreementAndMutants_003's 24.4 s, and facts-length's 25.6 s of
// TestSixRuleAgreementAndMutants_028's 26.1 s, alone on one core). Its key is typeAwareNativeKey's, as h.build's
// cached natives' is: the emitted C's hash (emission names no path, so the same mutant lowered from two directories is
// the same C) and the archive's. Its TestProduct_ makes it ahead. Every six mutant's native is one (sixMutants, and
// declaration-source's), and so is every TypeAwareAgreement mutant's: wrong-node, stale, the three bad frames and
// ignored-kind (the frames' links took 17 to 37 s of _006 to _008 under load, Eve and Adam, Oct 23).
func (h *harness) mutantNative(name, entry, archive string, sanitize bool) string {
	h.t.Helper()
	return h.mutantNativeOf(name, h.lowered(entry, name), archive, sanitize)
}

// mutantNativeOf is mutantNative for a mutant whose C is already emitted (loweredMutant).
func (h *harness) mutantNativeOf(name, source, archive string, sanitize bool) string {
	h.t.Helper()
	key, typeAwareNativeKeyError := typeAwareNativeKey("mutant native "+name, source, archive, sanitize)
	if typeAwareNativeKeyError != nil {
		h.t.Fatal(typeAwareNativeKeyError)
	}
	binary, cachedTypeAwareProductError := cachedTypeAwareProduct(key, func(directory string) (string, error) {
		binary := filepath.Join(directory, "native")
		started := time.Now()
		typeAwareNativeBuildError := typeAwareNativeBuild(source, binary, archive, sanitize)
		h.t.Logf("phase clang %s %.6fs", name, time.Since(started).Seconds())
		return binary, typeAwareNativeBuildError
	})
	if cachedTypeAwareProductError != nil {
		h.t.Fatal(cachedTypeAwareProductError)
	}
	return binary
}

// typeAwareLoweredKey keys an entry's emitted C by the entry's path in the repository, never where the checkout sits:
// a key carries no machine path, or a runner's checkout reads none of the builder's products.
func typeAwareLoweredKey(repository, entry string) (string, error) {
	relative, relativePathError := filepath.Rel(repository, entry)
	if relativePathError != nil {
		return "", relativePathError
	}
	return "C " + filepath.ToSlash(relative), nil
}

// typeAwareNativeKey keys a native by kind, the emitted C's hash, the archive's bytes' hash and sanitize. The archive
// is a product whose path names the cache directory, and so the machine, it sits in: by its bytes, the same archive
// keys the same native wherever it is cached (tsgo's archive rebuilds byte for byte, Builder's Linux proof, Oct 12).
func typeAwareNativeKey(kind, source, archive string, sanitize bool) (string, error) {
	digest, typeAwareArchiveDigestError := typeAwareArchiveDigest(archive)
	if typeAwareArchiveDigestError != nil {
		return "", typeAwareArchiveDigestError
	}
	return fmt.Sprintf("%s %x archive=%s sanitize=%t", kind, sha256.Sum256([]byte(source)), digest, sanitize), nil
}

// typeAwareArchiveDigests are the archives' SHA256s, read once per process each: a product never changes in place.
var typeAwareArchiveDigests sync.Map

func typeAwareArchiveDigest(archive string) (string, error) {
	if digest, ok := typeAwareArchiveDigests.Load(archive); ok {
		return digest.(string), nil
	}
	data, readError := os.ReadFile(archive)
	if readError != nil {
		return "", readError
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
func TestTypeAwareKeysCarryNoMachinePath(test *testing.T) {
	test.Parallel()
	repository, absolutePathError := filepath.Abs("../../..")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	var natives, lowered []string
	for _, machine := range []string{test.TempDir(), test.TempDir()} {
		// Where the archive product sits under this machine's ADAMIC_BUILD_CACHE_DIR, and where its checkout is.
		archive := filepath.Join(machine, "cache", "local", "d91f8adc3a72", "checker.a")
		if createDirectoryError := os.MkdirAll(filepath.Dir(archive), 0755); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		if writeError := os.WriteFile(
			archive,
			[]byte("the same archive bytes on both machines"),
			0644,
		); writeError != nil {
			test.Fatal(writeError)
		}
		native, typeAwareNativeKeyError := typeAwareNativeKey("binary", "int main(void) { return 0; }\n", archive, true)
		if typeAwareNativeKeyError != nil {
			test.Fatal(typeAwareNativeKeyError)
		}
		checkout := filepath.Join(machine, "checkout")
		c, typeAwareLoweredKeyError := typeAwareLoweredKey(
			checkout,
			filepath.Join(checkout, "stage1/cohere/typeaware/main.ts"),
		)
		if typeAwareLoweredKeyError != nil {
			test.Fatal(typeAwareLoweredKeyError)
		}
		natives, lowered = append(natives, native), append(lowered, c)
	}
	for _, keys := range []struct {
		name string
		keys []string
	}{{"native", natives}, {"emitted C", lowered}} {
		first, second := typeAwareSharedInputs(keys.keys[0]), typeAwareSharedInputs(keys.keys[1])
		if a, b := buildcache.NameKey(repository, first), buildcache.NameKey(repository, second); a != b {
			test.Fatalf(
				"%s: name key %s on one machine, %s on the other (%q, %q)",
				keys.name,
				a[:12],
				b[:12],
				keys.keys[0],
				keys.keys[1],
			)
		}
		if !slices.Equal(first.Files, second.Files) {
			test.Fatalf("%s: declared files %q on one machine, %q on the other", keys.name, first.Files, second.Files)
		}
		test.Logf(
			"%s: one name key (%s) and the same declared files from two machines",
			keys.name,
			buildcache.NameKey(repository, first)[:12],
		)
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
	source, cachedTypeAwareProductError := cachedTypeAwareProduct("C typeaware mutant wrong-node", func(
		_ string,
	) (string, error) {
		mutant, typeAwareWrongNodeSourceError := typeAwareWrongNodeSource(h.repository)
		if typeAwareWrongNodeSourceError != nil {
			return "", typeAwareWrongNodeSourceError
		}
		unary := filepath.Join(h.repository, "stage1/cohere/typeaware/unary_minus.ts")
		source, lowerOverlayToCError := h.lowerOverlayToC(
			filepath.Join(h.repository, "stage1/cohere/typeaware/main.ts"),
			map[string]string{unary: mutant},
			"wrong-node",
		)
		if lowerOverlayToCError != nil {
			return "", lowerOverlayToCError
		}
		return source, typeAwareCNamesNoCheckout(source, h.repository)
	})
	if cachedTypeAwareProductError != nil {
		h.t.Fatal(cachedTypeAwareProductError)
	}
	return h.mutantNativeOf("wrong-node", source, archive, sanitize)
}

// typeAwareWrongNodeSource is unary_minus.ts with the wrong-node mutation: the type asked of the whole expression.
func typeAwareWrongNodeSource(repository string) (string, error) {
	text, readError := os.ReadFile(filepath.Join(repository, "stage1/cohere/typeaware/unary_minus.ts"))
	if readError != nil {
		return "", readError
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
func TestTypeAwareCNamesNoCheckoutCanFail(test *testing.T) {
	test.Parallel()
	repository := filepath.Join(test.TempDir(), "adamic")
	named := "#include \"" + filepath.ToSlash(filepath.Join(repository, "stage1/typescript/scanner.ts")) + "\"\n"
	if typeAwareCNamesNoCheckout(named, repository) == nil {
		test.Fatal("C naming the checkout passed")
	}
	if typeAwareCNamesNoCheckoutError := typeAwareCNamesNoCheckout(
		"/* ../../typescript/scanner.ts */\n",
		repository,
	); typeAwareCNamesNoCheckoutError != nil {
		test.Fatal(typeAwareCNamesNoCheckoutError)
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
func (h *harness) compare(name, oracle, binary, configuration, manifest string) result {
	h.t.Helper()
	want := h.must(name+"-go", exec.Command(oracle, configuration, manifest))
	got := h.must(name+"-native", exec.Command(binary, configuration, manifest))
	if len(got.stderr) != 0 {
		h.t.Fatalf("sanitizer stderr: %s", got.stderr)
	}
	if !bytes.Equal(got.stdout, want.stdout) {
		i := firstDifference(got.stdout, want.stdout)
		h.t.Fatalf(
			"%s mismatch byte %d: native %q Go %q",
			name,
			i,
			got.stdout[max(0, i-50):min(len(got.stdout), i+250)],
			want.stdout[max(0, i-50):min(len(want.stdout), i+250)],
		)
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

func prepareTypeAwareAgreementAndMutants(test *testing.T) *typeAwarePlan {
	started := time.Now()
	repository, absolutePathError := filepath.Abs("../../..")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	directory := filepath.Join(productDirectory, "typeaware-plan")
	if createDirectoryError := os.MkdirAll(directory, 0755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	if artifacts := os.Getenv("ADAMIC_TYPEAWARE_ARTIFACTS"); artifacts != "" {
		directory, absolutePathError = filepath.Abs(artifacts)
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		if createDirectoryError := os.MkdirAll(directory, 0755); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
	}
	h := &harness{t: test, repository: repository, directory: directory, sixBuilds: true, setupStarted: started}
	traceGroup(test)
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
	configuration := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	var generatedPathsForUnion []string
	var shards []sixShard
	var expected []string
	rounds := 1
	if os.Getenv("ADAMIC_TYPEAWARE_BENCH") == "1" {
		rounds = 3
	}
	finish := func(compiler []string, benchmark bool) *typeAwarePlan {
		unsplit := typeAwareUnsplitIDs(generatedPathsForUnion, compiler, rounds, benchmark)
		if sixUnionError := sixUnion(
			unsplit,
			[]sixShard{{name: "original-work", ids: expected}},
		); sixUnionError != nil {
			test.Fatal(sixUnionError)
		}
		return newTypeAwarePlan(test, h, unsplit, shards, testTypeAwareAgreementAndMutantsShards)
	}
	add := func(name string, ids []string, run func(*harness)) {
		shards = append(shards, sixShard{name: name, ids: ids, run: run})
	}

	// Extract the production rule's table sources without duplicating its verdicts.
	pinnedFixtureFiles(test, repository, []string{"no_unsafe_unary_minus"})
	sourceFile := filepath.Join(repository, "cohere/internal/lint/rules/typescript/no_unsafe_unary_minus_test.go")
	tree, parseFileError := parser.ParseFile(token.NewFileSet(), sourceFile, nil, 0)
	if parseFileError != nil {
		test.Fatal(parseFileError)
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
		text, unquoteError := strconv.Unquote(value.Value)
		if unquoteError == nil && strings.Contains(text, ";") {
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
	truth := h.must("generated-go", exec.Command(oracle, configuration, manifest))
	generatedPaths := append([]string(nil), paths...)
	generatedPathsForUnion = generatedPaths
	expected = append(expected, sixIDs("agreement/generated", generatedPaths)...)
	add("agreement/generated", sixIDs("agreement/generated", generatedPaths), func(h *harness) {
		test := h.t
		truth := h.compare("generated", oracle, binary(h), configuration, manifest)
		if sixFindingUnionError := sixFindingUnion(truth.stdout, generatedPaths); sixFindingUnionError != nil {
			test.Fatal(sixFindingUnionError)
		}
		if !bytes.Contains(truth.stdout, []byte("is false instead.")) ||
			!bytes.Contains(truth.stdout, []byte("is string instead.")) {
			test.Fatal("nonempty union/string controls missing")
		}
		var generatedFindings, generatedQueries int
		fmt.Sscanf(summary(truth.stdout), "findings %d queries %d", &generatedFindings, &generatedQueries)
		if generatedFindings < 20 || generatedQueries < 35 {
			test.Fatalf("generated coverage too small: %s", summary(truth.stdout))
		}
		test.Logf("generated coverage: %d files", len(generatedPaths))

	})
	// No bridge calls can reach an ordinary backend without explicit linkage.
	expected = append(expected, "refusal/unlinked")
	add("refusal/unlinked", []string{"refusal/unlinked"}, func(h *harness) {
		stage0 := typeAwareStage0(h)
		test := h.t
		refusal := h.run("unlinked", exec.Command(stage0, "build", entry, "-o", filepath.Join(directory, "unlinked")))
		if refusal.err == nil || !bytes.Contains(refusal.stderr, []byte("unlinked typescript-go library call")) {
			test.Fatal("unlinked type-parts call was accepted")
		}
	})

	partsOnly := h.write(
		"parts-unlinked.ts",
		"import { tsgoTypeParts } from 'adamic'; console.log(tsgoTypeParts(1, 'source.ts', 0, 1, 'Identifier'));\n",
	)
	expected = append(expected, "refusal/parts-unlinked")
	add("refusal/parts-unlinked", []string{"refusal/parts-unlinked"}, func(h *harness) {
		stage0 := typeAwareStage0(h)
		test := h.t
		refusal := h.run(
			"parts-unlinked",
			exec.Command(stage0, "build", partsOnly, "-o", filepath.Join(directory, "parts-unlinked")),
		)
		if refusal.err == nil || !bytes.Contains(refusal.stderr, []byte("unlinked typescript-go library call")) {
			test.Fatal("unlinked type-parts call was accepted")
		}
	})

	// Asking the unary expression's type rather than its operand still compiles
	// and finishes; the independent production-rule findings must disagree.
	expected = append(expected, sixIDs("mutant/wrong-node", generatedPaths)...)
	add("mutant/wrong-node", sixIDs("mutant/wrong-node", generatedPaths), func(h *harness) {
		test := h.t
		wrong := typeAwareSetup(h, "wrong-node", func() string { return typeAwareWrongNodeNative(h, normal(h), true) })
		observed := h.must("wrong-node-run", exec.Command(wrong, configuration, manifest))
		if len(observed.stderr) != 0 || bytes.Equal(observed.stdout, truth.stdout) {
			test.Fatal("wrong-node mutant was not caught by findings alone")
		}
		test.Logf(
			"wrong-node type mutant: cohere byte oracle catches byte %d; %s",
			firstDifference(observed.stdout, truth.stdout),
			summary(observed.stdout),
		)
	})

	probe := h.write("cost-probe.ts", "-1;\n")
	releasedEntry := filepath.Join(repository, "stage1/cohere/typeaware/testdata/released.ts")
	expected = append(expected, []string{"refusal/released"}...)
	add("refusal/released", []string{"refusal/released"}, func(h *harness) {
		test := h.t
		released := typeAwareBuild(h, stage0, "released", releasedEntry, normal(h), false)
		stale := h.run("released-run", exec.Command(released, configuration, probe))
		if code, ok := stale.err.(*exec.ExitError); !ok ||
			code.ExitCode() != 70 ||
			string(stale.stderr) != "adamic: panic: invalid or released checker handle\n" {
			test.Fatalf("released query escaped: %v %s", stale.err, stale.stderr)
		}
		test.Log("released program queried: native panic 70 with invalid or released checker handle")
	})
	expected = append(expected, []string{"mutant/stale"}...)
	add("mutant/stale", []string{"mutant/stale"}, func(h *harness) {
		test := h.t
		staleBinary := typeAwareMutantNative(
			h,
			"stale-native",
			releasedEntry,
			typeAwareMutantArchive(h, "retained-handle"),
			false,
		)
		stale := h.must("stale-mutant-run", exec.Command(staleBinary, configuration, probe))
		if len(stale.stderr) != 0 {
			test.Fatal("stale mutant did not finish cleanly")
		}
		test.Log("released registry deletion mutant: stale-query expectation catches exit 0 instead of 70")
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
			test := h.t
			malformed := typeAwareMutantNative(
				h,
				change.name+"-native",
				entry,
				typeAwareMutantArchive(h, change.name),
				false,
			)
			got := h.run(change.name+"-run", exec.Command(malformed, configuration, manifest))
			if code, ok := got.err.(*exec.ExitError); !ok ||
				code.ExitCode() != 70 ||
				!bytes.Contains(got.stderr, []byte(change.message)) {
				test.Fatalf("%s escaped: %v %s", change.name, got.err, got.stderr)
			}
			test.Logf("%s mutant: panic 70, %s", change.name, change.message)
		})

	}

	// A kind/span mismatch cannot silently select a neighboring or enclosing node.
	costEntry := filepath.Join(repository, "stage1/cohere/typeaware/testdata/query_cost.ts")
	cost := func(h *harness) string { return typeAwareBuild(h, stage0, "query-cost", costEntry, normal(h), false) }
	readyCost := cost(h)
	cost = func(*harness) string { return readyCost }
	expected = append(expected, []string{"refusal/bad-kind"}...)
	add("refusal/bad-kind", []string{"refusal/bad-kind"}, func(h *harness) {
		test := h.t
		bad := typeAwareBadKindNative(h, "bad-kind", normal(h))
		got := h.run("bad-kind-run", exec.Command(bad, configuration, probe, "2"))
		if code, ok := got.err.(*exec.ExitError); !ok ||
			code.ExitCode() != 70 ||
			!bytes.Contains(got.stderr, []byte("no exact Identifier node")) {
			test.Fatalf("exact lookup mismatch escaped: %v %s", got.err, got.stderr)
		}
	})
	expected = append(expected, []string{"mutant/ignored-kind"}...)
	add("mutant/ignored-kind", []string{"mutant/ignored-kind"}, func(h *harness) {
		test := h.t
		ignoredBinary := typeAwareBadKindNative(h, "ignored-kind-native", typeAwareMutantArchive(h, "ignored-kind"))
		ignored := h.must("ignored-kind-run", exec.Command(ignoredBinary, configuration, probe, "2"))
		if len(ignored.stderr) != 0 {
			test.Fatal("kind-guard mutant did not finish cleanly")
		}
		test.Log("exact kind guard mutant: mismatch-refusal expectation catches exit 0")
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
			test := h.t
			command := exec.Command(cost(h), configuration, probe, "10000")
			command.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
			nativeCost := h.must(fmt.Sprintf("query-cost-native-%d", round), command)
			direct := h.must(
				fmt.Sprintf("query-cost-go-%d", round),
				exec.Command(oracle, configuration, probeManifest, "--query-cost", "10000"),
			)
			if !bytes.Equal(nativeCost.stdout, direct.stdout) {
				test.Fatal("query-cost result differs from direct Go")
			}
			test.Logf(
				"query cost round %d native: %s; direct Go: %s; same output=%s",
				round,
				strings.TrimSpace(string(nativeCost.stderr)),
				strings.TrimSpace(string(direct.stderr)),
				strings.TrimSpace(string(direct.stdout)),
			)
		})

	}

	corpus := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if corpus == "" {
		test.Log("compiler corpus skipped: set ADAMIC_TYPESCRIPT_SOURCE (#xq2ecw6)")
	}
	paths = nil
	if corpus != "" {
		paths = pinnedCompilerFiles(test, corpus)
	}
	sort.Strings(paths)
	compilerManifest := h.write("compiler.manifest", strings.Join(paths, "\n")+"\n")
	compilerConfiguration := filepath.Join(corpus, "src/compiler/tsconfig.json")
	test.Logf("compiler corpus: %d files, original compiler tsconfig", len(paths))
	expected = append(expected, sixIDs("agreement/compiler", paths)...)
	add("agreement/compiler", sixIDs("agreement/compiler", paths), func(h *harness) {
		if len(paths) == 0 {
			h.t.Skip("set ADAMIC_TYPESCRIPT_SOURCE")
		}
		truth := h.compare("compiler", oracle, binary(h), compilerConfiguration, compilerManifest)
		if sixFindingUnionError := sixFindingUnion(truth.stdout, paths); sixFindingUnionError != nil {
			h.t.Fatal(sixFindingUnionError)
		}
	})
	{
		for _, corpus := range []struct{ name, config, manifest string }{
			{"compiler", compilerConfiguration, compilerManifest},
			{"generated", configuration, manifest},
		} {
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
					test := h.t
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
							test.Fatal("timed finding counts differ")
						}
						expected = r.stdout
						var findings, queries int
						if _, scanError := fmt.Sscanf(
							string(r.stdout),
							"findings %d queries %d",
							&findings,
							&queries,
						); scanError != nil {
							test.Fatal(scanError)
						}
						test.Logf(
							"%s round %d %s: %.6fs %.2f findings/s findings=%d queries=%d; %s",
							corpus.name,
							round,
							filepath.Base(binary),
							r.elapsed.Seconds(),
							float64(findings)/r.elapsed.Seconds(),
							findings,
							queries,
							strings.TrimSpace(string(r.stderr)),
						)
					}
				})

			}
		}
	}
	return finish(paths, rounds == 3 && len(paths) != 0)
}

func TestPinnedTypeFlags(test *testing.T) {
	test.Parallel()
	actual := checker.TypeFlagsAny | checker.TypeFlagsNever | checker.TypeFlagsNumberLike | checker.TypeFlagsBigIntLike
	if actual != 334017 {
		test.Fatalf("update Adamic's mask and rerun cohere oracle: flags=%d", actual)
	}
}

func typeAwareUnsplitIDs(generated, compiler []string, rounds int, benchmark bool) []string {
	ids := []string{
		"refusal/unlinked",
		"refusal/parts-unlinked",
		"refusal/released",
		"mutant/stale",
		"refusal/bad-kind",
		"mutant/ignored-kind",
	}
	for _, name := range []string{
		"agreement/generated",
		"mutant/wrong-node",
		"mutant/empty-frame",
		"mutant/missing-header",
		"mutant/bad-length",
	} {
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
