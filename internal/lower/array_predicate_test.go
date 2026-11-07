package lower

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

func TestArrayPredicatePreservesDeclaredElementContract(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		`function guard(value: string | number[]): value is number[] { return Array.isArray(value); } console.log(String(guard([1, 2]))); console.log(String(guard('x')));`,
		`function guard(value: unknown): value is readonly unknown[] { return Array.isArray(value); } console.log(String(guard([1, 2]))); console.log(String(guard('x')));`,
		`function count(value: unknown): number { if (Array.isArray(value)) return value.length; return -1; } console.log(String(count([1, 2]))); console.log(String(count('x')));`,
		`const value: unknown = [1, 2]; if (Array.isArray(value)) console.log(String(value.length));`,
		`function guard(value: string | readonly number[]): value is readonly number[] { if (Array.isArray(value)) { return true; } return false; } console.log(String(guard([1, 2]))); console.log(String(guard('x')));`,
	} {
		lowersAndAgreesWithNode(test, source)
	}
}

func TestArrayPredicateDoesNotMisclassifyNativeTuples(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		`const tuple: [number, string] = [1, 'x']; console.log(String(Array.isArray(tuple)));`,
		`function test(value: {}): boolean { return Array.isArray(value); }`,
	} {
		_, lowerSourceError := lowerSource(test, source)
		var notYet *NotYet
		if !errors.As(lowerSourceError, &notYet) || !strings.Contains(lowerSourceError.Error(), "tuple or an erased") {
			test.Fatalf("%s: want an explicit representation refusal, got %v", source, lowerSourceError)
		}
	}
}

// The builtin is a member whose every declaration is the library's or the prelude's. A program's own
// ArrayConstructor merge adds a declaration of its own to isArray, whose predicate the engine never
// checks, so that member is no builtin. Lower refuses the merge first, as a global augmentation, and
// this pins the recognizer behind it, on the checked program. Mutant: arrayPredicateMember taking any
// declaration, which recognizes the merged member as the builtin.
func TestArrayPredicateIgnoresAProgramsOwnMerge(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct {
		source  string
		builtin bool
	}{
		{"const value: unknown = ['a'];\nconst result = Array.isArray(value);\n", true},
		{
			"const value: unknown = ['a'];\nconst result = Array.isArray(value);\ndeclare global { interface ArrayConstructor { isArray(arg: unknown): arg is string[]; } }\n",
			false,
		},
	} {
		path := writeAgreementSource(test, "main.a", probe.source)
		program, loadError := load.Load([]string{path})
		if loadError != nil {
			test.Fatal(loadError)
		}
		file := program.Files()[0]
		checked, release := program.Checker(context.Background(), file)
		call := file.Statements.Nodes[1].AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes[0].Initializer()
		builtin := (&lowering{checker: checked}).arrayPredicateMember(call.AsCallExpression().Expression)
		release()
		if builtin != probe.builtin {
			test.Errorf("%q: builtin %v, want %v", probe.source, builtin, probe.builtin)
		}
		_, lowerError := Lower(context.Background(), program)
		var refused *Refused
		if probe.builtin && lowerError != nil ||
			!probe.builtin && (!errors.As(lowerError, &refused) || refused.Error() != path+
				":3:1: Adamic 0.1 refuses a global augmentation; define the values in a module and import them; use Adamic-generated platform bindings for host values") {
			test.Errorf("%q: Lower: %v", probe.source, lowerError)
		}
	}
}

func TestUnknownArrayPredicateRefusesUnrepresentedObservations(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		`function test(value: unknown): unknown { if (Array.isArray(value)) return value[0]; return undefined; }`,
		`function test(value: unknown): boolean { return Array.isArray(value); } const pair: [number, string] = [1, 'x']; console.log(String(test(pair)));`,
	} {
		_, lowerSourceError := lowerSource(test, source)
		var notYet *NotYet
		if !errors.As(lowerSourceError, &notYet) {
			test.Fatalf("%s: want named representation refusal, got %v", source, lowerSourceError)
		}
	}
}

func TestArrayPredicateCoexistsWithUnknownReflection(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		`function test(value: unknown): string { return typeof value; } console.log(test(1)); console.log(test('x'));`,
		`function test(value: unknown): boolean { return value === null; } console.log(String(test(null))); console.log(String(test(1)));`,
		`function test(value: unknown): unknown { return value; } console.log(String(test(7) === 7));`,
		`function errorCode(error: unknown): string | undefined { return typeof error === 'object' && error !== null && 'code' in error && typeof error.code === 'string' ? error.code : undefined; } console.log(errorCode({ code: 'ok' }) ?? 'missing'); console.log(errorCode(null) ?? 'missing');`,
		`function isArray(value: unknown): value is readonly unknown[] { return Array.isArray(value); } function length(value: unknown): number { if (isArray(value)) return value.length; return -1; } console.log(String(length([1, 2]))); console.log(String(length('x')));`,
	} {
		lowersAndAgreesWithNode(test, source)
	}
}

// Since C1 the checker narrows Array.isArray and a readonly unknown[] wrapper itself, so lowering
// reads the checker's narrowed type, with no recovery from the declared type of its own. These are
// the reads that recovery used to answer: a negated test read as a scalar, a for-of over the array
// member, a wrapper's element read with ??, an unknown, and a true branch with no array member at
// all, where the checker says Box & readonly unknown[] and the recovery said never. Mutant: the
// narrowed local read through its declared type in place of the checker's (locals.go:277), which
// stops the for-of over the array member.
func TestArrayPredicateReadsTheCheckersNarrowing(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		"function scalar(value: string | number | number[]): string { if (!Array.isArray(value)) { return `scalar ${value}`; } return `array ${value.length}`; }\nconsole.log(scalar('x')); console.log(scalar(2)); console.log(scalar([1, 2]));",
		"function total(value: string | number[]): number { let sum = 0; if (Array.isArray(value)) { for (const item of value) { sum += item; } } return sum; }\nconsole.log(String(total([1, 2, 3]))); console.log(String(total('x')));",
		"function isList(value: unknown): value is readonly unknown[] { return Array.isArray(value); }\nfunction first(value: string | number[]): string { if (isList(value)) { return `list ${value[0] ?? -1}`; } return `other ${value}`; }\nconsole.log(first([7])); console.log(first([])); console.log(first('q'));",
		"function size(value: unknown): string { if (Array.isArray(value)) { return `size ${value.length}`; } return 'none'; }\nconsole.log(size([1])); console.log(size(3));",
		"class Box { n = 1; }\nfunction never(value: Box): string { if (Array.isArray(value)) { return `array ${value.n}`; } return `box ${value.n}`; }\nconsole.log(never(new Box()));",
		"function text(value: string): string { if (Array.isArray(value)) { return `array ${value.length}`; } return `string ${value}`; }\nconsole.log(text('x'));",
	} {
		lowersAndAgreesWithNode(test, source)
	}
}
