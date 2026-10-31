package lower

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

func TestArrayPredicatePreservesDeclaredElementContract(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`function guard(value: string | number[]): value is number[] { return Array.isArray(value); } console.log(String(guard([1, 2]))); console.log(String(guard('x')));`,
		`function guard(value: unknown): value is readonly unknown[] { return Array.isArray(value); } console.log(String(guard([1, 2]))); console.log(String(guard('x')));`,
		`function count(value: unknown): number { if (Array.isArray(value)) return value.length; return -1; } console.log(String(count([1, 2]))); console.log(String(count('x')));`,
		`const value: unknown = [1, 2]; if (Array.isArray(value)) console.log(String(value.length));`,
		`function guard(value: string | readonly number[]): value is readonly number[] { if (Array.isArray(value)) { return true; } return false; } console.log(String(guard([1, 2]))); console.log(String(guard('x')));`,
	} {
		lowersAndAgreesWithNode(t, source)
	}
}

func TestArrayPredicateDoesNotMisclassifyNativeTuples(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`const tuple: [number, string] = [1, 'x']; console.log(String(Array.isArray(tuple)));`,
		`function test(value: {}): boolean { return Array.isArray(value); }`,
	} {
		_, err := lowerSource(t, source)
		var notYet *NotYet
		if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "tuple or an erased") {
			t.Fatalf("%s: want an explicit representation refusal, got %v", source, err)
		}
	}
}

// The builtin is a member whose every declaration is the library's or the prelude's. A program's own
// ArrayConstructor merge adds a declaration of its own to isArray, whose predicate the engine never
// checks, so that member is no builtin. Lower refuses the merge first, as a global augmentation, and
// this pins the recognizer behind it, on the checked program. Mutant: arrayPredicateMember taking any
// declaration, which recognizes the merged member as the builtin.
func TestArrayPredicateIgnoresAProgramsOwnMerge(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		source  string
		builtin bool
	}{
		{"const value: unknown = ['a'];\nconst result = Array.isArray(value);\n", true},
		{"const value: unknown = ['a'];\nconst result = Array.isArray(value);\ndeclare global { interface ArrayConstructor { isArray(arg: unknown): arg is string[]; } }\n", false},
	} {
		path := writeAgreementSource(t, "main.a", probe.source)
		program, err := load.Load([]string{path})
		if err != nil {
			t.Fatal(err)
		}
		file := program.Files()[0]
		checked, release := program.Checker(context.Background(), file)
		call := file.Statements.Nodes[1].AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes[0].Initializer()
		builtin := (&lowering{checker: checked}).arrayPredicateMember(call.AsCallExpression().Expression)
		release()
		if builtin != probe.builtin {
			t.Errorf("%q: builtin %v, want %v", probe.source, builtin, probe.builtin)
		}
		if _, err := Lower(context.Background(), program); probe.builtin == (err != nil) {
			t.Errorf("%q: Lower: %v", probe.source, err)
		}
	}
}

func TestUnknownArrayPredicateRefusesUnrepresentedObservations(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`function test(value: unknown): unknown { if (Array.isArray(value)) return value[0]; return undefined; }`,
		`function test(value: unknown): boolean { return Array.isArray(value); } const pair: [number, string] = [1, 'x']; console.log(String(test(pair)));`,
	} {
		_, err := lowerSource(t, source)
		var notYet *NotYet
		if !errors.As(err, &notYet) {
			t.Fatalf("%s: want named representation refusal, got %v", source, err)
		}
	}
}

func TestArrayPredicateCoexistsWithUnknownReflection(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`function test(value: unknown): string { return typeof value; } console.log(test(1)); console.log(test('x'));`,
		`function test(value: unknown): boolean { return value === null; } console.log(String(test(null))); console.log(String(test(1)));`,
		`function test(value: unknown): unknown { return value; } console.log(String(test(7) === 7));`,
		`function errorCode(error: unknown): string | undefined { return typeof error === 'object' && error !== null && 'code' in error && typeof error.code === 'string' ? error.code : undefined; } console.log(errorCode({ code: 'ok' }) ?? 'missing'); console.log(errorCode(null) ?? 'missing');`,
		`function isArray(value: unknown): value is readonly unknown[] { return Array.isArray(value); } function length(value: unknown): number { if (isArray(value)) return value.length; return -1; } console.log(String(length([1, 2]))); console.log(String(length('x')));`,
	} {
		lowersAndAgreesWithNode(t, source)
	}
}
