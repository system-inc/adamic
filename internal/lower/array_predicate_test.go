package lower

import (
	"context"
	"errors"
	"fmt"
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
		_, err = Lower(context.Background(), program)
		var refused *Refused
		if probe.builtin && err != nil || !probe.builtin && (!errors.As(err, &refused) || refused.Error() != path+":3:1: Adamic 0.1 refuses a global augmentation; define the values in a module and import them; use Adamic-generated platform bindings for host values") {
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

// Since C1 the checker narrows Array.isArray and a readonly unknown[] wrapper itself, so lowering
// reads the checker's narrowed type, with no recovery from the declared type of its own. These are
// the reads that recovery used to answer: a negated test read as a scalar, a for-of over the array
// member, a wrapper's element read with ??, an unknown, and a true branch with no array member at
// all, where the checker says Box & readonly unknown[] and the recovery said never. Mutant: the
// narrowed local read through its declared type in place of the checker's (locals.go:277), which
// stops the for-of over the array member.
func TestArrayPredicateReadsTheCheckersNarrowing(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"function scalar(value: string | number | number[]): string { if (!Array.isArray(value)) { return `scalar ${value}`; } return `array ${value.length}`; }\nconsole.log(scalar('x')); console.log(scalar(2)); console.log(scalar([1, 2]));",
		"function total(value: string | number[]): number { let sum = 0; if (Array.isArray(value)) { for (const item of value) { sum += item; } } return sum; }\nconsole.log(String(total([1, 2, 3]))); console.log(String(total('x')));",
		"function isList(value: unknown): value is readonly unknown[] { return Array.isArray(value); }\nfunction first(value: string | number[]): string { if (isList(value)) { return `list ${value[0] ?? -1}`; } return `other ${value}`; }\nconsole.log(first([7])); console.log(first([])); console.log(first('q'));",
		"function size(value: unknown): string { if (Array.isArray(value)) { return `size ${value.length}`; } return 'none'; }\nconsole.log(size([1])); console.log(size(3));",
		"class Box { n = 1; }\nfunction never(value: Box): string { if (Array.isArray(value)) { return `array ${value.n}`; } return `box ${value.n}`; }\nconsole.log(never(new Box()));",
		"function text(value: string): string { if (Array.isArray(value)) { return `array ${value.length}`; } return `string ${value}`; }\nconsole.log(text('x'));",
	} {
		lowersAndAgreesWithNode(t, source)
	}
}

// An interface arm declaring a required property no array can have is in Array.isArray's domain (Adamic's ruling
// on #s8gy89r): Tagged's kind and name are on no type stored as an array, so isArray decides `Tagged | readonly
// string[]` both ways, as Node does. Its false branch reads value.kind and value.name through the checked
// narrowed-union read, which excludes the array member with !Array.isArray because Tagged is in the domain
// (locals.go); with that read's old class-only test, it stops there as "a narrowed union member whose object tag
// cannot be checked with typeof".
func TestArrayPredicateDecidesAnArrayExcludingInterfaceArm(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, `interface Tagged { kind: 't'; name: string }
function describe(value: Tagged | readonly string[]): string {
	if (Array.isArray(value)) return `+"`array ${value.length} ${value.join('+')}`"+`;
	return `+"`tagged ${value.kind} ${value.name}`"+`;
}
function isList(value: Tagged | readonly string[]): boolean { return Array.isArray(value); }
console.log(describe({ kind: 't', name: 'x' }));
console.log(describe(['a', 'b']));
console.log(describe([]));
console.log(String(isList({ kind: 't', name: 'y' })));
console.log(String(isList(['z'])));
console.log(String(isList([])));
`)
}

// An interface arm whose every required property an array can have stays NotYet at isArray (Language's pins on
// #s8gy89r), each beside Tagged and handed the array Node calls one, so Node's answer is true where Adamic's
// arm would have said false: Pair's are numeric-named and length, a tuple's; StringAt's at is lib's Array's;
// TextRange's pos and end are NodeArray's, a program interface extending ReadonlyArray, and an Object.assign's
// onto an array, an intersection no declaration names; Named's label is Labeled's, which extends ReadonlyArray;
// Iterates' Symbol.iterator is lib Array's, matched by the checker's own name for the symbol member (Language's
// pins on #9z2ksv7). Raw's raw is TemplateStringsArray's, which the library walk holds for the day that value
// lowers; today the widening refusal meets the value first, where the tag hands its TemplateStringsArray to
// isList as Raw | Tagged, by its kind, words, line and column.
// Mutant: arrayStoredPropertyNames left holding lib's Array and ReadonlyArray members alone, without the closed
// world. The two TextRange rows and the Named row go red here, each lowering past isArray; symbol-iterator stays
// green, its name being lib Array's, and so does template-strings, which stops before isArray.
func TestArrayPredicateInterfaceArmAnArrayCanSatisfyStaysNotYet(t *testing.T) {
	t.Parallel()
	const tagged = "interface Tagged { kind: 't'; name: string }\n"
	for _, row := range []struct {
		name, source, node string
		// where and what are the first stop's, a NotYet at isArray's argument when where is empty; fix is set
		// for a Refused, whose fix it is.
		where, what, fix string
	}{
		{"pair", "interface Pair { 0: string; 1: string; length: 2 }\nfunction isList(value: Tagged | Pair): boolean { return Array.isArray(value); }\nconst pair: [string, string] = ['a', 'b'];\nconsole.log(String(isList(pair)));\n", "true\n", "", "", ""},
		{"string-at", "interface StringAt { at(index: number): string | undefined }\nfunction isList(value: Tagged | StringAt): boolean { return Array.isArray(value); }\nconst list: string[] = ['a'];\nconsole.log(String(isList(list)));\n", "true\n", "", "", ""},
		{"text-range-node-array", "interface TextRange { pos: number; end: number }\ninterface NodeArray<T> extends ReadonlyArray<T>, TextRange { readonly hasTrailingComma?: boolean }\nfunction isList(value: Tagged | TextRange): boolean { return Array.isArray(value); }\nconsole.log(String(isList({ kind: 't', name: 'x' })));\n", "false\n", "", "", ""},
		{"text-range-assign", "interface TextRange { pos: number; end: number }\nfunction isList(value: Tagged | TextRange): boolean { return Array.isArray(value); }\nconsole.log(String(isList(Object.assign([1], { pos: 0, end: 1 }))));\n", "true\n", "", "", ""},
		{"read-only-array-extension", "interface Labeled extends ReadonlyArray<string> { readonly label: string }\ninterface Named { label: string }\nfunction isList(value: Tagged | Named): boolean { return Array.isArray(value); }\nconsole.log(String(isList({ kind: 't', name: 'x' })));\n", "false\n", "", "", ""},
		{"symbol-iterator", "interface Iterates { [Symbol.iterator](): Iterator<string> }\nfunction isList(value: Tagged | Iterates): boolean { return Array.isArray(value); }\nconst list: string[] = ['a'];\nconsole.log(String(isList(list)));\n", "true\n", "", "", ""},
		{"template-strings", "interface Raw { raw: readonly string[] }\nfunction tag(parts: TemplateStringsArray): string { return String(isList(parts)); }\nfunction isList(value: Tagged | Raw): boolean { return Array.isArray(value); }\nconsole.log(tag`ab`);\n", "true\n", ":3:74", "a value of type TemplateStringsArray seen as Raw | Tagged, whose readonly field raw becomes writable: a readonly field may hold something narrower than readonly string[], which a write of readonly string[] would replace", "keep raw readonly in the type it's seen as, or copy the value ({ ...value }) (adamic/invariant-mutable)"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			source := tagged + row.source
			if node := runAgreementNode(t, writeAgreementSource(t, "main.a", source)); string(node.stdout) != row.node || node.code != 0 {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q; want %q", node.code, node.stdout, node.stderr, row.node)
			}
			_, err := lowerSource(t, source)
			if row.fix != "" {
				var refused *Refused
				if !errors.As(err, &refused) || !strings.HasSuffix(refused.Where, row.where) || refused.What != row.what || refused.Fix != row.fix {
					t.Fatalf("want Refused at %s, %q; %q, got %v", row.where, row.what, row.fix, err)
				}
				return
			}
			var notYet *NotYet
			// The argument isArray is handed, by line and column, so a stop elsewhere (Object.assign's) isn't this one.
			where, what := row.where, row.what
			if where == "" {
				what = "Array.isArray on a tuple or an erased object/any/unknown view"
				for index, line := range strings.Split(source, "\n") {
					if column := strings.Index(line, "Array.isArray("); column >= 0 {
						where = fmt.Sprintf(":%d:%d", index+1, column+len("Array.isArray(")+1)
					}
				}
			}
			if !errors.As(err, &notYet) || !strings.HasSuffix(notYet.Where, where) || notYet.What != what {
				t.Fatalf("want NotYet at %s, %q, got %v", where, what, err)
			}
		})
	}
}

// TextRange beside readonly string[] stays NotYet at the first stop its false branch's read would meet: the
// isArray call's argument, since pos and end are NodeArray's, so the narrowed read never adds a test of its own.
func TestArrayPredicateNotYetInterfaceArmStopsAtTheCall(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `interface TextRange { pos: number; end: number }
interface NodeArray<T> extends ReadonlyArray<T>, TextRange { readonly hasTrailingComma?: boolean }
function describe(value: TextRange | readonly string[]): string {
	if (Array.isArray(value)) return 'array';
	return `+"`range ${value.pos}`"+`;
}
console.log(describe({ pos: 0, end: 1 }));
`)
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.HasSuffix(notYet.Where, ":4:20") || notYet.What != "Array.isArray on a tuple or an erased object/any/unknown view" {
		t.Fatalf("want NotYet at isArray's argument (:4:20), got %v", err)
	}
}

// Lower's probe on #9z2ksv7: a checked view checks at a read, never at its cast, so a tuple, held as an object, could
// sit behind an arm Array.isArray's domain admits with nothing read, and the brand would call it no array where Node
// calls it one. Each road from a tuple to Named, written as .ts, where a cast to an interface can be a checked view,
// and measured on main and at the cut (J-kbp1hvv, J-k8ssmqr). Node prints array for each. Every road but the overlap
// view stops where the tuple would enter it, by its own refusal; the overlap view (`[number, number] | Other` as
// Named, overlapViewCast) lowered and printed named natively, and now stops at the isArray call by
// checkArrayArmViews. Each row goes red if its road opens, so the pass and the domain are judged again.
// Mutant: checkArrayArmViews returning nil at once. overlap-view lowers and goes red.
func TestArrayPredicateArmViewRoadsStop(t *testing.T) {
	t.Parallel()
	const named = "interface Named { readonly name: string }\nconst pair: [number, number] = [1, 2];\n"
	const print = "console.log(Array.isArray(x) ? \"array\" : \"named\");\n"
	const uncheckable = "Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)"
	for _, row := range []struct {
		name, source, stop string
	}{
		{"unknown-hop", named + "const hidden: unknown = pair;\nconst x: Named | string[] = hidden as Named;\n" + print, ":4:29: " + uncheckable},
		{"json-parse", named + "const x: Named | string[] = JSON.parse(\"[1,2]\") as Named;\n" + print, ":3:29: " + uncheckable},
		{"object-hop", named + "const o: object = pair;\nconst x: Named | string[] = o as Named;\n" + print, ":4:29: " + uncheckable},
		{"union-member-cast", named + "function hold(): Named | [number, number] { return pair; }\nconst u = hold();\nconst x: Named | string[] = u as Named;\n" + print, ":5:29: " + uncheckable},
		{"structural-view", named + "interface Lengthy { readonly length: number; readonly name?: string }\ninterface NamedLength { readonly name: string; readonly length: number }\nconst s: Lengthy = pair;\nconst x: NamedLength | string[] = s as NamedLength;\n" + print, ":5:20: Adamic 0.1 refuses optional property name in Lengthy absent from structural source [number, number], which can hide fields; declare name on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"},
		{"empty-object-view", named + "const e: {} = pair;\nconst x: Named | string[] = e as Named;\n" + print, ":4:29: " + uncheckable},
		{"function-return", named + "function view(value: unknown): Named { return value as Named; }\nconst x: Named | string[] = view(pair);\n" + print, ":3:47: " + uncheckable},
		{"overlap-view", named + "interface Other { readonly name: string; readonly kind: 'o' }\nfunction hold(): [number, number] | Other { return pair; }\nconst u = hold();\nconst x: Named | string[] = u as Named;\n" + print, ""},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			path := writeAgreementSource(t, "main.ts", row.source)
			if node := runAgreementNode(t, path); string(node.stdout) != "array\n" || node.code != 0 {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q; want array", node.code, node.stdout, node.stderr)
			}
			program, err := load.Load([]string{path})
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			_, err = Lower(context.Background(), program)
			if row.stop != "" {
				var refused *Refused
				if !errors.As(err, &refused) || err.Error() != path+row.stop {
					t.Fatalf("want %q, got %v", row.stop, err)
				}
				return
			}
			var notYet *NotYet
			if !errors.As(err, &notYet) || notYet.Where != path+":7:27" || notYet.What != "Array.isArray over Named, which the checked view of Named over Other | [number, number] at "+path+":6:29 can hold an array behind, unread" {
				t.Fatalf("want checkArrayArmViews' NotYet at :7:27, got %v", err)
			}
		})
	}
}
