package lower

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	_ "github.com/system-inc/adamic/internal/ccjobserver"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/native/agreementproducts"
	"github.com/system-inc/adamic/internal/native/runtimeproducts"
	"github.com/system-inc/adamic/internal/toolchain"
)

type agreementTesting interface {
	Helper()
	TempDir() string
	Fatal(...any)
	Fatalf(string, ...any)
}

type nodeObservation struct {
	stdout, stderr []byte
	code           int
}

// lowersAndAgreesWithNode holds acceptance to behavior. Any additional IR
// assertion must name in a comment the property behavior cannot observe.
func lowersAndAgreesWithNode(t *testing.T, source string) *ir.Program {
	t.Helper()
	return lowersAndAgreesWithNodeExit(t, source, 0)
}

func lowersAndAgreesWithNodeExit(t *testing.T, source string, exit int) *ir.Program {
	t.Helper()
	return agreeSource(t, source, exit, nil)
}

// lowersAndAgreesWithNodeNative adds an unsanitized native observation to the
// source and JavaScript agreement. Each opting-in row names what only native shows,
// and runs a program from agreementproducts' programs by its name.
func lowersAndAgreesWithNodeNative(t *testing.T, name string) *ir.Program {
	t.Helper()
	nativeAgreement(t, name)
	return agreeSourceNative(t, name)
}

// nativeAgreement fails t unless name is a program agreementproducts holds, the
// programs every native agreement row builds, each named for the row that runs it.
// TestProduct_LowerAgreementC lowers each one, and TestProduct_LowerAgreementNative
// builds each one's binary ahead under the key its row's runAgreementNative fetches,
// so a runner that reads products and never builds them (ADAMIC_BUILD_CACHE=read)
// finds every binary. A program not there has no product.
func nativeAgreement(t agreementTesting, name string) {
	t.Helper()
	if !slices.Contains(agreementproducts.Programs(), name) {
		t.Fatalf("native agreement program %q is not in agreementproducts' programs, so no product builds it", name)
	}
}

// TestProduct_LowerAgreementC builds the lower agreement C (agreementproducts.C): every
// program's source and the C stage 0 lowers it to, planted as its row plants it. Each
// native is handed its C there, a product, never a file its test wrote.
func TestProduct_LowerAgreementC(t *testing.T) {
	t.Parallel()
	agreementproducts.C(t)
}

// TestProduct_LowerAgreementNative builds every program's native from the lower
// agreement C, under the key its row's runAgreementNative fetches.
func TestProduct_LowerAgreementNative(t *testing.T) {
	t.Parallel()
	product := agreementC(t)
	for _, name := range agreementproducts.Programs() {
		buildAgreementNative(t, product, name)
	}
}

// Handed the archive native.Build links, native.BuildLean writes Build's bytes for every
// program in agreementproducts' programs, wherever its output is and whatever jobserver it's
// handed, and handed the release archive runtimeproducts holds, the same bytes again: the
// lean product is the binary the rows checked when Build made it (#egjbcre). Each program's C
// is the lower agreement C's, which is first the C this test lowers in process from the same
// source, planted as its row plants it, byte for byte: stage 0's `adamic c` is the lowering
// the rows ran before. It's built six ways and compared byte for byte: by Build; by BuildLean
// into lean under no jobserver; into another directory, named apart and longer, so its build
// directory beside the binary is too; into lean again under a jobserver of one token and one
// of eight; and with the product's archive. Each binary is named native, as the product names
// it: on macOS the linker's ad-hoc signature names the binary by its file's name, so binaries
// named apart differ there and nowhere else.
func TestNativeAgreementLeanBuildIsBuildsBytes(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_NATIVE_SPLIT") == "1" {
		t.Skip("ADAMIC_NATIVE_SPLIT=1 splits native.Build, and a lean build is one translation unit")
	}
	options := native.Options{}
	compiler, err := toolchain.Path("clang")
	if err != nil {
		t.Fatal(err)
	}
	jobservers := map[int]*native.Jobserver{0: nil, 1: agreementJobserver(t, 1), 8: agreementJobserver(t, 8)}
	product := agreementC(t)
	for _, name := range agreementproducts.Programs() {
		source := filepath.Join(product, name, "main.c")
		held, err := os.ReadFile(source)
		if err != nil {
			t.Fatal(err)
		}
		code := string(held)
		program, err := os.ReadFile(filepath.Join(product, name, "main.a"))
		if err != nil {
			t.Fatal(err)
		}
		lowered, err := agreementproducts.Plant(name, native.C(lowerAgreementEntry(t, writeAgreementSource(t, string(program)))))
		if err != nil {
			t.Fatal(err)
		}
		if lowered != code {
			t.Errorf("%s: the lower agreement C holds %d bytes of C, and this test lowers %d other bytes from its source", name, len(code), len(lowered))
			continue
		}
		directory := t.TempDir()
		for _, output := range []string{"built", "lean", "another-output-directory", "product"} {
			if err := os.Mkdir(filepath.Join(directory, output), 0o755); err != nil {
				t.Fatal(err)
			}
		}
		if err := native.Build(code, filepath.Join(directory, "built", "native"), options); err != nil {
			t.Fatalf("%s: native.Build: %v", name, err)
		}
		built, err := os.ReadFile(filepath.Join(directory, "built", "native"))
		if err != nil {
			t.Fatal(err)
		}
		library, err := native.RuntimeLibraryForSource("", code, options)
		if err != nil {
			t.Fatal(err)
		}
		archive := runtimeproducts.Archive(t, "release", native.Features(code), agreementArchiveOwners, agreementArchiveGroup)
		same := true
		for _, way := range []struct {
			name, output, archive string
			tokens                int
		}{
			{"lean", "lean", filepath.Dir(library), 0},
			{"at another output directory", "another-output-directory", filepath.Dir(library), 0},
			{"under one token", "lean", filepath.Dir(library), 1},
			{"under eight tokens", "lean", filepath.Dir(library), 8},
			{"with the product's archive", "product", archive, 0},
		} {
			output := filepath.Join(directory, way.output, "native")
			if err := native.BuildLean(compiler, source, way.archive, output, options, jobservers[way.tokens]); err != nil {
				t.Fatalf("%s: native.BuildLean %s: %v", name, way.name, err)
			}
			binary, err := os.ReadFile(output)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(binary, built) {
				same = false
				t.Errorf("%s: native.BuildLean's binary %s (%d bytes) differs from native.Build's (%d bytes)", name, way.name, len(binary), len(built))
			}
			if _, err := os.Stat(filepath.Join(directory, way.output, "build")); !errors.Is(err, os.ErrNotExist) {
				t.Errorf("%s: native.BuildLean %s left its build directory behind: %v", name, way.name, err)
			}
		}
		if same {
			t.Logf("%s, features %v: %d bytes, the same six ways", name, native.Features(code), len(built))
		}
	}
}

// agreementJobserver is a compile jobserver holding tokens, a FIFO filled as Loom fills one.
func agreementJobserver(t *testing.T, tokens int) *native.Jobserver {
	t.Helper()
	path := filepath.Join(t.TempDir(), "jobserver")
	if err := syscall.Mkfifo(path, 0o600); err != nil {
		t.Fatal(err)
	}
	// Opened for reading and writing, as native opens it, so filling it doesn't wait for a reader.
	filler, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { filler.Close() })
	if _, err := filler.Write([]byte(strings.Repeat("+", tokens))); err != nil {
		t.Fatal(err)
	}
	jobserver, err := native.OpenJobserver(path)
	if err != nil {
		t.Fatal(err)
	}
	return jobserver
}

// agreeSourceNative runs the program name from the lower agreement C, its source
// lowered here to JavaScript and agreed with Node, and its native, from the C the
// product holds, agreed with Node too. A planted native answer lives only in that C
// (agreementproducts.Plant), so it proves this leg independently of the JavaScript
// comparison.
func agreeSourceNative(t testing.TB, name string) *ir.Program {
	t.Helper()
	product := agreementC(t)
	source, err := os.ReadFile(filepath.Join(product, name, "main.a"))
	if err != nil {
		t.Fatal(err)
	}
	path := writeAgreementSource(t, string(source))
	program := agreeEntry(t, path, lowerAgreementEntry(t, path), 0, nil, nil)
	want := runAgreementNode(t, path)
	got := runAgreementNative(t, product, name)
	compareNativeAgreement(t, got, want)
	return program
}

// agreementArchiveOwners are the product tests that build the release archives an agreement
// native links, one a group of runtimeproducts' Sets, and agreementArchiveGroup the one group
// whose test TestProduct_LowerAgreementNative waits for in cloud/fast-gate/product-order.json:
// every program in agreementproducts' programs turns on no feature, the first set. It finds the
// archive its C needs by name, and a program needing another group's fails, naming the file.
// agreementCOwner is the product test that builds the lower agreement C, which
// TestProduct_LowerAgreementNative waits for there too.
const (
	agreementArchiveOwners = "internal/oracle TestProduct_NativeRuntimesRelease"
	agreementArchiveGroup  = 0
	agreementCOwner        = "internal/lower TestProduct_LowerAgreementC"
)

// agreementC is the lower agreement C's directory, found first, as the release archive
// is: built ahead by its owner under the trace, and outside it built here when it's missed.
func agreementC(t testing.TB) string {
	t.Helper()
	return agreementproducts.FindC(t, agreementCOwner)
}

// buildAgreementNative is the one recipe for a native agreement binary: the product
// test builds it ahead and runAgreementNative fetches it, so the key can't drift
// between them. It returns the product's directory, which holds the binary as native.
// The binary is a lean product (agreementproducts) of what it's handed: name's C in the
// lower agreement C, product, and the release archive of the runtime for the C's
// features, each found first as a product of its own. A hit runs the binary without
// invoking clang.
func buildAgreementNative(t testing.TB, product, name string) string {
	t.Helper()
	source := filepath.Join(product, name, "main.c")
	code, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	archive := runtimeproducts.Archive(t, "release", native.Features(string(code)), agreementArchiveOwners, agreementArchiveGroup)
	return agreementproducts.Product(t, source, archive, os.Getenv("ADAMIC_NATIVE_SPLIT"))
}

func runAgreementNative(t testing.TB, product, name string) nodeObservation {
	t.Helper()
	binary := filepath.Join(buildAgreementNative(t, product, name), "native")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	exit := 0
	if ctx.Err() != nil {
		t.Fatalf("Native %s: %v", binary, ctx.Err())
	}
	if err != nil {
		failure, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("Native %s: %v", binary, err)
		}
		exit = failure.ExitCode()
	}
	return nodeObservation{stdout.Bytes(), stderr.Bytes(), exit}
}

func compareNativeAgreement(t agreementTesting, got, want nodeObservation) {
	t.Helper()
	if !bytes.Equal(got.stdout, want.stdout) {
		t.Fatalf("Native backend stdout = %q, source Node = %q", got.stdout, want.stdout)
	}
	if got.code != want.code {
		t.Fatalf("Native backend exit = %d, source Node = %d; stderr %q", got.code, want.code, got.stderr)
	}
}

// The per-call mutation seam lets proving tests plant bad lowered output without
// changing global state, so acceptance rows remain safe under t.Parallel.
func agreeSource(t agreementTesting, source string, exit int, mutate func(*ir.Program)) *ir.Program {
	t.Helper()
	return agreeEntry(t, writeAgreementSource(t, source), nil, exit, nil, mutate)
}

func writeAgreementSource(t agreementTesting, source string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "main.a")
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func lowerAgreementEntry(t agreementTesting, path string) *ir.Program {
	t.Helper()
	checked, err := load.Load([]string{path})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	program, err := Lower(context.Background(), checked)
	if err != nil {
		t.Fatalf("Lower refused acceptance row: %v", err)
	}
	return program
}

// checkedFailure is only for existing soundness probes whose inserted check
// deliberately rejects a lying TypeScript predicate that Node trusts.
func agreeEntry(t agreementTesting, path string, program *ir.Program, exit int, checkedFailure *nodeObservation, mutate func(*ir.Program)) *ir.Program {
	t.Helper()
	if program == nil {
		program = lowerAgreementEntry(t, path)
	}
	if mutate != nil {
		mutate(program)
	}
	generated := filepath.Join(t.TempDir(), "generated.mjs")
	if err := os.WriteFile(generated, []byte(javascript.JavaScript(program)), 0o644); err != nil {
		t.Fatal(err)
	}
	want := runAgreementNode(t, path)
	if len(want.stdout) == 0 {
		t.Fatal("empty-answer probe: source Node stdout is empty; print what the row computes")
	}
	if want.code != exit {
		t.Fatalf("source Node exit = %d, want %d; stderr %q", want.code, exit, want.stderr)
	}
	got := runAgreementNode(t, generated)
	if checkedFailure != nil {
		want = *checkedFailure
	}
	compareAgreement(t, got, want)
	// Preserve the existing predicate probes' complete checked-failure contract.
	if checkedFailure != nil && !bytes.Equal(got.stderr, want.stderr) {
		t.Fatalf("checked predicate stderr = %q, want %q", got.stderr, want.stderr)
	}
	return program
}

func runAgreementNode(t agreementTesting, path string) nodeObservation {
	t.Helper()
	runner, err := filepath.Abs("../../oracle/node.mjs")
	if err != nil {
		t.Fatal(err)
	}
	// The node toolchain.Path finds, under toolchain's environment, as every other site of lower's runs it.
	node, err := toolchain.Path("node")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, node, "--disable-warning=ExperimentalWarning", runner, path)
	command.Env = toolchain.Env()
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err = command.Run()
	code := 0
	if ctx.Err() != nil {
		t.Fatalf("Node %s: %v", path, ctx.Err())
	}
	if err != nil {
		failure, ok := err.(*exec.ExitError)
		if !ok {
			t.Fatalf("Node %s: %v", path, err)
		}
		code = failure.ExitCode()
	}
	return nodeObservation{stdout.Bytes(), stderr.Bytes(), code}
}

func compareAgreement(t agreementTesting, got, want nodeObservation) {
	t.Helper()
	if !bytes.Equal(got.stdout, want.stdout) {
		t.Fatalf("JavaScript backend stdout = %q, source Node = %q", got.stdout, want.stdout)
	}
	if got.code != want.code {
		t.Fatalf("JavaScript backend exit = %d, source Node = %d", got.code, want.code)
	}
	if got.code != 0 {
		firstLine := func(value []byte) string { return strings.SplitN(string(value), "\n", 2)[0] }
		if firstLine(got.stderr) != firstLine(want.stderr) {
			t.Fatalf("JavaScript backend stderr first line = %q, source Node = %q", firstLine(got.stderr), firstLine(want.stderr))
		}
	} else if len(got.stderr) != 0 || len(want.stderr) != 0 {
		t.Fatalf("unexpected stderr: JavaScript backend %q, source Node %q", got.stderr, want.stderr)
	}
}

type agreementRecorder struct {
	*testing.T
	message string
}

type agreementRecordedFailure struct{}

func (r *agreementRecorder) Fatal(values ...any) {
	r.message = fmt.Sprint(values...)
	panic(agreementRecordedFailure{})
}
func (r *agreementRecorder) Fatalf(format string, values ...any) {
	r.Fatal(fmt.Sprintf(format, values...))
}
func recordAgreementFailure(t *testing.T, source string, mutate func(*ir.Program)) string {
	t.Helper()
	recorder := &agreementRecorder{T: t}
	func() {
		defer func() {
			if value := recover(); value != nil {
				if _, ok := value.(agreementRecordedFailure); !ok {
					panic(value)
				}
			}
		}()
		agreeSource(recorder, source, 0, mutate)
	}()
	return recorder.message
}

func TestAgreementRejectsWrongLoweredOutput(t *testing.T) {
	t.Parallel()
	message := recordAgreementFailure(t, "console.log('computed');", func(program *ir.Program) { program.Strings[0] = "wrong" })
	if !strings.Contains(message, "JavaScript backend stdout") {
		t.Fatalf("planted wrong output was not caught by stdout comparison: %q", message)
	}
}

func TestAgreementRejectsEmptyAnswer(t *testing.T) {
	t.Parallel()
	message := recordAgreementFailure(t, "const computed = 1 + 2;", nil)
	if !strings.Contains(message, "print what the row computes") {
		t.Fatalf("silent row was not refused by empty-answer probe: %q", message)
	}
}

func TestAgreementAcceptsComputedAnswer(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, "console.log(`${1 + 2}`);")
}

func TestAgreementAcceptsExpectedPanic(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNodeExit(t, "import { panic } from 'adamic'; console.log('before panic'); panic('witness');", 70)
}

func TestAgreementEnumConstantWitness(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, "enum Answer { Value = 7 } console.log(`${Answer.Value}`);")
}

func TestAgreementInheritedFieldWitness(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, "class Base { readonly #base = 7; read(): number { return this.#base; } } class Child extends Base { readonly own = 11; } const child = new Child(); console.log(`${child.read()}:${child.own}:${Object.keys(child).join(',')}`);")
}

func TestAgreementParameterPropertyWitness(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, "class Box { constructor(readonly value: number) {} } console.log(`${new Box(7).value}`);")
}

func TestAgreementStaticInitializerWitness(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, staticSideEffectSource)
}

func TestAgreementParseIntRadixWitness(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, "console.log(['10', '10', '10'].map(Number.parseInt).join(','));")
}

// recordNativeAgreementFailure runs a program from agreementproducts' programs by its
// name. A name that isn't one fails t itself, not the recorder.
func recordNativeAgreementFailure(t *testing.T, name string) string {
	t.Helper()
	nativeAgreement(t, name)
	recorder := &agreementRecorder{T: t}
	func() {
		defer func() {
			if value := recover(); value != nil {
				if _, ok := value.(agreementRecordedFailure); !ok {
					panic(value)
				}
			}
		}()
		agreeSourceNative(recorder, name)
	}()
	return recorder.message
}

func TestNativeAgreementRejectsWrongOutput(t *testing.T) {
	t.Parallel()
	// Only native shows the planted C string; JavaScript still prints computed.
	message := recordNativeAgreementFailure(t, "PlantedWrongOutput")
	if !strings.Contains(message, "Native backend stdout") {
		t.Fatalf("planted native output was not caught by stdout comparison: %q", message)
	}
}

func TestNativeAgreementRejectsWrongExit(t *testing.T) {
	t.Parallel()
	// Only native shows the planted exit status after printing the correct answer.
	message := recordNativeAgreementFailure(t, "PlantedWrongExit")
	if !strings.Contains(message, "Native backend exit = 7, source Node = 0") {
		t.Fatalf("planted native exit was not caught by exit comparison: %q", message)
	}
}

// TestNativeAgreementRowsNameRegisteredPrograms reads this package's tests, building
// nothing: every native agreement row names a program in agreementproducts' programs by
// a literal, only the registry's own plumbing reaches the native build, and every
// program the registry holds is run by some row, so the product test builds exactly
// what the rows fetch.
func TestNativeAgreementRowsNameRegisteredPrograms(t *testing.T) {
	t.Parallel()
	// The functions below the registry, and the only functions that may call each.
	reaches := map[string]map[string]bool{
		"agreeSourceNative":    {"lowersAndAgreesWithNodeNative": true, "recordNativeAgreementFailure": true},
		"runAgreementNative":   {"agreeSourceNative": true},
		"buildAgreementNative": {"runAgreementNative": true, "TestProduct_LowerAgreementNative": true},
	}
	rows := map[string]bool{"lowersAndAgreesWithNodeNative": true, "recordNativeAgreementFailure": true}
	files, err := filepath.Glob("*_test.go")
	if err != nil {
		t.Fatal(err)
	}
	fileSet := token.NewFileSet()
	used := map[string]bool{}
	for _, name := range files {
		file, err := parser.ParseFile(fileSet, name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok {
				// Outside a function a name can only be held (var f = runAgreementNative), and a call through what holds
				// it is one no check here follows.
				ast.Inspect(declaration, func(node ast.Node) bool {
					if identifier, ok := node.(*ast.Ident); ok && (rows[identifier.Name] || reaches[identifier.Name] != nil) {
						t.Errorf("%s: %s is named outside a function, where no row check reaches it", fileSet.Position(identifier.Pos()), identifier.Name)
					}
					return true
				})
				continue
			}
			if function.Body == nil {
				continue
			}
			called := map[*ast.Ident]bool{}
			ast.Inspect(function.Body, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				callee, ok := call.Fun.(*ast.Ident)
				if !ok || !rows[callee.Name] {
					return true
				}
				called[callee] = true
				program := ""
				if len(call.Args) == 2 {
					if literal, ok := call.Args[1].(*ast.BasicLit); ok && literal.Kind == token.STRING {
						program, _ = strconv.Unquote(literal.Value)
					}
				}
				if !slices.Contains(agreementproducts.Programs(), program) {
					t.Errorf("%s: %s must name a program in agreementproducts' programs by a literal, so TestProduct_LowerAgreementNative builds it", fileSet.Position(call.Pos()), callee.Name)
				}
				used[program] = true
				return true
			})
			ast.Inspect(function.Body, func(node ast.Node) bool {
				identifier, ok := node.(*ast.Ident)
				if !ok {
					return true
				}
				if rows[identifier.Name] && !called[identifier] {
					t.Errorf("%s: %s is used other than called with a program's name", fileSet.Position(identifier.Pos()), identifier.Name)
				}
				if callers, guarded := reaches[identifier.Name]; guarded && !callers[function.Name.Name] {
					t.Errorf("%s: %s reaches %s, which builds a program no product builds; name one in agreementproducts' programs instead", fileSet.Position(identifier.Pos()), function.Name.Name, identifier.Name)
				}
				return true
			})
		}
	}
	for _, name := range agreementproducts.Programs() {
		if !used[name] {
			t.Errorf("agreementproducts' programs hold %s, which no row runs, so its product is built for nothing", name)
		}
	}
}
