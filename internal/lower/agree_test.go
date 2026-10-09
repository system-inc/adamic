package lower

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/native"
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
// source and JavaScript agreement. Each opting-in row names what only native shows.
func lowersAndAgreesWithNodeNative(t *testing.T, source string) *ir.Program {
	t.Helper()
	return agreeSourceNative(t, source, nil)
}

// The mutation seam changes only C, after JavaScript agreement, so planted native
// answers prove this leg independently of the JavaScript comparison.
func agreeSourceNative(t agreementTesting, source string, mutate func(string) string) *ir.Program {
	t.Helper()
	path := filepath.Join(t.TempDir(), "main.a")
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	program := agreeEntry(t, path, nil, 0, nil, nil)
	want := runAgreementNode(t, path)
	code := native.C(program)
	if mutate != nil {
		code = mutate(code)
	}
	got := runAgreementNative(t, code)
	compareNativeAgreement(t, got, want)
	return program
}

func runAgreementNative(t agreementTesting, code string) nodeObservation {
	t.Helper()
	options := native.Options{}
	flags := append(native.Flags(options), fmt.Sprintf("C=%x", sha256.Sum256([]byte(code))),
		"ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"))
	// The shared build-product cache hashes C contents, runtime and build code,
	// native flags and toolchain. A hit runs the binary without invoking clang.
	directory, err := buildcache.Get(buildcache.Inputs{
		Name:      "lower agreement native",
		Files:     []string{"internal/native"},
		Flags:     flags,
		Toolchain: []string{runtime.GOOS, runtime.GOARCH, buildcache.Tool("clang", "--version")},
	}, func(directory string) error {
		return native.Build(code, filepath.Join(directory, "native"), options)
	})
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
	path := filepath.Join(t.TempDir(), "main.a")
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	return agreeEntry(t, path, nil, exit, nil, mutate)
}

// checkedFailure is only for existing soundness probes whose inserted check
// deliberately rejects a lying TypeScript predicate that Node trusts.
func agreeEntry(t agreementTesting, path string, program *ir.Program, exit int, checkedFailure *nodeObservation, mutate func(*ir.Program)) *ir.Program {
	t.Helper()
	if program == nil {
		checked, err := load.Load([]string{path})
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		program, err = Lower(context.Background(), checked)
		if err != nil {
			t.Fatalf("Lower refused acceptance row: %v", err)
		}
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "node", "--disable-warning=ExperimentalWarning", runner, path)
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
	// Step 21 holds uncaught errors to stdout and exit 1; engine-specific
	// stderr rendering is outside that contract. Explicit panic diagnostics stay pinned.
	if got.code != 0 && got.code != 1 {
		firstLine := func(value []byte) string { return strings.SplitN(string(value), "\n", 2)[0] }
		if firstLine(got.stderr) != firstLine(want.stderr) {
			t.Fatalf("JavaScript backend stderr first line = %q, source Node = %q", firstLine(got.stderr), firstLine(want.stderr))
		}
	} else if got.code == 0 && (len(got.stderr) != 0 || len(want.stderr) != 0) {
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

func recordNativeAgreementFailure(t *testing.T, source string, mutate func(string) string) string {
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
		agreeSourceNative(recorder, source, mutate)
	}()
	return recorder.message
}

func TestNativeAgreementRejectsWrongOutput(t *testing.T) {
	t.Parallel()
	// Only native shows the planted C string; JavaScript still prints computed.
	message := recordNativeAgreementFailure(t, "console.log('computed');", func(code string) string {
		return strings.ReplaceAll(code, "computed", "wrong")
	})
	if !strings.Contains(message, "Native backend stdout") {
		t.Fatalf("planted native output was not caught by stdout comparison: %q", message)
	}
}

func TestNativeAgreementRejectsWrongExit(t *testing.T) {
	t.Parallel()
	// Only native shows the planted exit status after printing the correct answer.
	message := recordNativeAgreementFailure(t, "console.log('computed');", func(code string) string {
		index := strings.LastIndex(code, "return 0;")
		if index < 0 {
			t.Fatal("native main has no successful exit to mutate")
		}
		return code[:index] + strings.Replace(code[index:], "return 0;", "return 7;", 1)
	})
	if !strings.Contains(message, "Native backend exit = 7, source Node = 0") {
		t.Fatalf("planted native exit was not caught by exit comparison: %q", message)
	}
}
