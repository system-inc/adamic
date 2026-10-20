package lower

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	_ "github.com/system-inc/adamic/internal/ccjobserver"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/native"
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
// and runs a program from nativeAgreementPrograms by its name.
func lowersAndAgreesWithNodeNative(t *testing.T, name string) *ir.Program {
	t.Helper()
	entry := nativeAgreement(t, name)
	return agreeSourceNative(t, entry.source, entry.planted(t))
}

// nativeAgreementProgram is a program a native agreement row builds: its source and,
// for a row that plants a wrong native answer, the change it makes to the C, which
// fails rather than plants when the C has no site for it.
type nativeAgreementProgram struct {
	source string
	mutate func(code string) (string, error)
}

// nativeAgreementPrograms are every program the native agreement rows build, named
// for the row that runs it. A row reaches its program by name, and
// TestProduct_LowerAgreementNative builds each one ahead under the key the row
// fetches, so a runner that reads products and never builds them
// (ADAMIC_BUILD_CACHE=read) finds every binary. A program not here has no product.
var nativeAgreementPrograms = map[string]nativeAgreementProgram{
	"ArrayNarrowingOptionalNumber":  {source: "const items: (number | undefined)[] = [1, 2];\nif (items.every(value => value !== undefined)) console.log('total ' + items.reduce((a, b) => a + b, 0));\n"},
	"ArrayNarrowingOptionalString":  {source: "const items: (string | undefined)[] = ['a', 'b'];\nif (items.every(value => value !== undefined)) console.log(items.join(''));\n"},
	"ArrayNarrowingSomeNegated":     {source: "function isNumber(value: string | number): value is number { return typeof value === 'number'; }\nconst items: (string | number)[] = ['a', 'b', 1];\nitems.pop();\nif (!items.some(isNumber)) {\n let text = '';\n for (const value of items) if (typeof value === 'string') text += value;\n console.log(text);\n}\n"},
	"ArrayNarrowingFindLastIndex":   {source: "function isNumber(value: string | number): value is number { return typeof value === 'number'; }\nconst items: (string | number)[] = ['x', 2, 4];\nconsole.log('index ' + items.findLastIndex(isNumber));\n"},
	"ArrayNarrowingPreservedUnion":  {source: "function isScalar(value: string | number | boolean): value is string | number { return typeof value !== 'boolean'; }\nconst items: (string | number | boolean)[] = [1, 'b'];\nif (items.every(isScalar)) {\n for (const value of items) console.log(typeof value === 'number' ? 'number ' + value : value);\n}\n"},
	"ArrayNarrowingBooleanCallback": {source: "const items: (string | number)[] = [1, 2, 'discard'];\nitems.pop();\nconsole.log(String(items.every((value): boolean => typeof value === 'number')));\n"},
	"FX5MixedLiteral":               {source: "const items: (string | number)[] = [1, 'a'];\nconsole.log(`${typeof items[0]} ${items[0]}`);\nconsole.log(`${typeof items[1]} ${items[1]}`);\n"},
	"RegExpUnicodeClassSource":      {source: "const regex = new RegExp('[[/]/', 'u'); console.log(`${regex.source}|${regex.flags}|${regex.test('//')}`);"},
	"RegExpSlashClassSource":        {source: "const regex = new RegExp('[/]/', 'u'); console.log(`${regex.source}|${regex.flags}|${regex.test('//')}`);"},
	"RegExpEscapedSlashSource":      {source: "const regex = new RegExp('\\\\/', 'u'); console.log(`${regex.source}|${regex.flags}|${regex.test('/')}`);"},
	"TypedArrayFromArray": {source: `const values = new Uint8Array([1, 2]);
console.log(values.length.toString());
for (const value of values) { console.log(value.toString()); }
`},
	"TypedArrayFromEmptyArray": {source: `const values = new Uint8Array([]);
console.log(values.length.toString());
for (const value of values) { console.log(value.toString()); }
`},
	"PlantedWrongOutput": {source: "console.log('computed');", mutate: func(code string) (string, error) {
		return strings.ReplaceAll(code, "computed", "wrong"), nil
	}},
	"PlantedWrongExit": {source: "console.log('computed');", mutate: func(code string) (string, error) {
		// main returns the host's exit status (process.exitCode, else 0); plant a fixed 7 there.
		const exit = "return adamic_host_exit_status();"
		index := strings.LastIndex(code, exit)
		if index < 0 {
			return "", errors.New("native main has no successful exit to mutate")
		}
		return code[:index] + strings.Replace(code[index:], exit, "return 7;", 1), nil
	}},
}

func nativeAgreement(t agreementTesting, name string) nativeAgreementProgram {
	t.Helper()
	entry, ok := nativeAgreementPrograms[name]
	if !ok {
		t.Fatalf("native agreement program %q is not in nativeAgreementPrograms, so no product builds it", name)
	}
	return entry
}

// planted is the entry's change to C as agreeSourceNative's seam takes it, failing t
// when the C has no site for it; nil for an entry that plants nothing.
func (entry nativeAgreementProgram) planted(t agreementTesting) func(string) string {
	if entry.mutate == nil {
		return nil
	}
	return func(code string) string {
		mutated, err := entry.mutate(code)
		if err != nil {
			t.Fatal(err)
		}
		return mutated
	}
}

// TestProduct_LowerAgreementNative builds every program in nativeAgreementPrograms,
// lowered and planted as its row lowers and plants it, under the key its row's
// runAgreementNative fetches.
func TestProduct_LowerAgreementNative(t *testing.T) {
	t.Parallel()
	for _, entry := range nativeAgreementPrograms {
		program := lowerAgreementEntry(t, writeAgreementSource(t, entry.source))
		if _, err := buildAgreementNative(agreementNativeCode(program, entry.planted(t))); err != nil {
			t.Fatalf("Native build: %v", err)
		}
	}
}

// The mutation seam changes only C, after JavaScript agreement, so planted native
// answers prove this leg independently of the JavaScript comparison.
func agreeSourceNative(t agreementTesting, source string, mutate func(string) string) *ir.Program {
	t.Helper()
	path := writeAgreementSource(t, source)
	program := agreeEntry(t, path, lowerAgreementEntry(t, path), 0, nil, nil)
	want := runAgreementNode(t, path)
	got := runAgreementNative(t, agreementNativeCode(program, mutate))
	compareNativeAgreement(t, got, want)
	return program
}

// agreementNativeCode is the C a native agreement row builds: the program's, with
// the row's planted change.
func agreementNativeCode(program *ir.Program, mutate func(string) string) string {
	code := native.C(program)
	if mutate != nil {
		code = mutate(code)
	}
	return code
}

// agreementNativeBuilds lets one native agreement product build at a time. Each build's
// traced window holds the whole process's reads, so parallel rows building at once each
// recorded the others' reads as their own: a ring of 14 products whose sets read each
// other, which Loom refuses to record and rebuilds in every tree (#kfjqsnh). The rows
// still check in parallel; only the builds they wait on are serial.
var agreementNativeBuilds sync.Mutex

// buildAgreementNative is the one recipe for a native agreement binary: the product
// test builds it ahead and runAgreementNative fetches it, so the key can't drift
// between them. It returns the product's directory, which holds the binary as native.
func buildAgreementNative(code string) (string, error) {
	agreementNativeBuilds.Lock()
	defer agreementNativeBuilds.Unlock()
	options := native.Options{}
	flags := append(native.Flags(options), fmt.Sprintf("C=%x", sha256.Sum256([]byte(code))),
		"ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"))
	// The shared build-product cache hashes C contents, runtime and build code,
	// native flags and toolchain. A hit runs the binary without invoking clang.
	return buildcache.Get(buildcache.Inputs{
		// v2: the sets recorded beside another build's window name that build's reads (#kfjqsnh).
		Name:      "lower agreement native v2",
		Files:     []string{"internal/native"},
		Flags:     flags,
		Toolchain: []string{runtime.GOOS, runtime.GOARCH, buildcache.Tool("clang", "--version")},
	}, func(directory string) error {
		return native.Build(code, filepath.Join(directory, "native"), options)
	})
}

func runAgreementNative(t agreementTesting, code string) nodeObservation {
	t.Helper()
	directory, err := buildAgreementNative(code)
	if err != nil {
		t.Fatalf("Native build: %v", err)
	}
	binary := filepath.Join(directory, "native")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err = command.Run()
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

// recordNativeAgreementFailure runs a program from nativeAgreementPrograms by its
// name. Its planted change fails t itself, not the recorder, when the C has no site.
func recordNativeAgreementFailure(t *testing.T, name string) string {
	t.Helper()
	entry := nativeAgreement(t, name)
	recorder := &agreementRecorder{T: t}
	func() {
		defer func() {
			if value := recover(); value != nil {
				if _, ok := value.(agreementRecordedFailure); !ok {
					panic(value)
				}
			}
		}()
		agreeSourceNative(recorder, entry.source, entry.planted(t))
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
// nothing: every native agreement row names a program in nativeAgreementPrograms by
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
				if _, registered := nativeAgreementPrograms[program]; !registered {
					t.Errorf("%s: %s must name a program in nativeAgreementPrograms by a literal, so TestProduct_LowerAgreementNative builds it", fileSet.Position(call.Pos()), callee.Name)
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
					t.Errorf("%s: %s reaches %s, which builds a program no product builds; name one in nativeAgreementPrograms instead", fileSet.Position(identifier.Pos()), function.Name.Name, identifier.Name)
				}
				return true
			})
		}
	}
	for name := range nativeAgreementPrograms {
		if !used[name] {
			t.Errorf("nativeAgreementPrograms holds %s, which no row runs, so its product is built for nothing", name)
		}
	}
}
