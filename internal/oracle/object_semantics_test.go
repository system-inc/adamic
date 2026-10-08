package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"path/filepath"
	"reflect"
	"testing"
)

func init() {
	for _, path := range []string{"cached_own", "ordinary_primitive", "catchable_errors"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{
			"internal/oracle/testdata/object_semantics/" + path + ".a", true, false,
		})
	}
}

func TestObjectCachedIntrinsic(t *testing.T) {
	t.Parallel()
	objectSemanticsNode(t, "cached_own")
}

func TestObjectOrdinaryPrimitive(t *testing.T) {
	t.Parallel()
	objectSemanticsNode(t, "ordinary_primitive")
}

func objectSemanticsNode(t *testing.T, name string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/object_semantics/"+name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	t.Logf("Node: stdout=%q stderr=%q exit=%d", want.stdout, want.stderr, want.exitCode)
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	got, binary := nativelyUncached(t, program)
	if diff := disagreement(want, got); diff != "" {
		t.Errorf("native: %s, stdout=%q stderr=%q exit=%d", diff, got.stdout, got.stderr, got.exitCode)
	}
	if got.exitCode == 0 {
		if report := leaks(t, program, binary); report != "" {
			t.Error(report)
		}
	}
	if diff := disagreement(want, releasedUncached(t, program)); diff != "" {
		t.Errorf("release: %s", diff)
	}
	if diff := disagreement(want, onJavaScriptBackend(t, program)); diff != "" {
		t.Errorf("JavaScript: %s", diff)
	}
}

// A borrowed intrinsic must ignore the receiver's own hasOwnProperty. This
// mutant gives every query the false result of that overriding field instead.
func TestObjectCachedIntrinsicMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/object_semantics/cached_own.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := 0
	mutate := func(value ir.Expression) ir.Expression {
		if _, ok := value.(ir.HasOwn); ok {
			changed++
			return ir.BooleanConstant{Value: false}
		}
		return value
	}
	for index := range program.Functions {
		mutateStringExpressions(reflect.ValueOf(&program.Functions[index].Body).Elem(), mutate)
	}
	mutateStringExpressions(reflect.ValueOf(&program.Main).Elem(), mutate)
	if changed == 0 {
		t.Fatal("mutant did not reach a borrowed own-key query")
	}
	want := onNode(t, path)
	got, binary := nativelyUncached(t, program)
	if got.exitCode != 0 || len(got.stderr) != 0 {
		t.Fatalf("mutant must finish: %+v", got)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
	if difference := disagreement(want, got); difference != "stdout differs" {
		t.Fatalf("native mutant escaped: %s", difference)
	}
	if difference := disagreement(want, onJavaScriptBackend(t, program)); difference != "stdout differs" {
		t.Fatalf("JavaScript mutant escaped: %s", difference)
	}
	t.Logf("Node caught %d queries dispatched to the receiver override; native exited zero without leaks", changed)
}

// Exhausting OrdinaryToPrimitive must throw. The mutant silently substitutes
// the ordinary object tag, which Node catches even though native is leak-clean.
func TestObjectOrdinaryPrimitiveMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/object_semantics/ordinary_primitive.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	tag := len(program.Strings)
	program.Strings = append(program.Strings, "[object Object]")
	changed := 0
	for index := range program.Functions {
		function := &program.Functions[index]
		if function.Name != "ordinary_primitive" {
			continue
		}
		for at, statement := range function.Body {
			if _, throws := statement.(ir.Throw); throws {
				function.Body[at] = ir.Return{Value: ir.StringConstant{Index: tag}}
				changed++
			}
		}
	}
	if changed == 0 {
		t.Fatal("no conversion-exhaustion throw mutated")
	}
	want := onNode(t, path)
	got, binary := nativelyUncached(t, program)
	if got.exitCode != 0 || len(got.stderr) != 0 {
		t.Fatalf("mutant must finish: %+v", got)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
	if difference := disagreement(want, got); difference != "stdout differs" {
		t.Fatalf("native mutant escaped: %s", difference)
	}
	if difference := disagreement(want, onJavaScriptBackend(t, program)); difference != "stdout differs" {
		t.Fatalf("JavaScript mutant escaped: %s", difference)
	}
	t.Logf("Node caught %d conversion-exhaustion throws erased; native exited zero without leaks", changed)
}

func TestObjectAccessorReadinessRemainsTerminal(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "stage3/fixtures/objects/15_accessor_absence.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := onNode(t, path)
	if source.exitCode != 0 {
		t.Fatalf("source Node: %+v", source)
	}
	got, _ := nativelyUncached(t, program)
	if got.exitCode != 70 || string(got.stderr) != "adamic: panic: read before assignment: variable 'secondAccessor' in secondAccessor\n" {
		t.Fatalf("readiness check changed: %+v", got)
	}
	if difference := disagreement(got, onJavaScriptBackend(t, program)); difference != "" {
		t.Fatalf("JavaScript: %s", difference)
	}
	if difference := disagreement(got, releasedUncached(t, program)); difference != "" {
		t.Fatalf("release: %s", difference)
	}
}

func TestObjectCatchableErrors(t *testing.T) {
	t.Parallel()
	objectSemanticsNode(t, "catchable_errors")
}

// Wrong exception identity is a clean miscompile; only source Node catches it.
func TestObjectCatchableErrorNameMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/object_semantics/catchable_errors.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	index := len(program.Strings)
	program.Strings = append(program.Strings, "Error")
	changed := 0
	mutate := func(value ir.Expression) ir.Expression {
		if error, ok := value.(ir.MakeError); ok && error.Name != nil {
			error.Name = ir.StringConstant{Index: index}
			changed++
			return error
		}
		return value
	}
	for at := range program.Functions {
		mutateStringExpressions(reflect.ValueOf(&program.Functions[at].Body).Elem(), mutate)
	}
	if changed == 0 {
		t.Fatal("no checked library error mutated")
	}
	want := onNode(t, path)
	got, binary := nativelyUncached(t, program)
	if got.exitCode != 0 || len(got.stderr) != 0 {
		t.Fatalf("mutant must finish: %+v", got)
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
	if difference := disagreement(want, got); difference != "stdout differs" {
		t.Fatalf("native mutant escaped: %s", difference)
	}
	if difference := disagreement(want, onJavaScriptBackend(t, program)); difference != "stdout differs" {
		t.Fatalf("JavaScript mutant escaped: %s", difference)
	}
	t.Logf("Node caught %d incorrectly named exceptions; native exited zero without leaks", changed)
}
