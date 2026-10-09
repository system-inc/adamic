package nexus

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// correctnessNoImportCycleLoadTimeReadRun runs the rule on subject with every file of the fixture in
// the program.
func correctnessNoImportCycleLoadTimeReadRun(t *testing.T, files map[string]string, subject string) rule_testing.Result {
	t.Helper()
	return rule_testing.RunTypedFiles(t, CorrectnessNoImportCycleLoadTimeRead, files, subject)
}

// correctnessNoImportCycleLoadTimeReadExpect asserts the findings, in source order, by the text each
// points at.
func correctnessNoImportCycleLoadTimeReadExpect(t *testing.T, result rule_testing.Result, want ...string) {
	t.Helper()
	// The shared assertion by message id, which proves the rule can fire; the checks below hold
	// each finding's exact text.
	if len(want) > 0 {
		wantIds := make([]string, len(want))
		for index := range wantIds {
			wantIds[index] = correctnessNoImportCycleLoadTimeReadId
		}
		rule_testing.ExpectFindings(t, result, wantIds...)
	}
	diagnostics := append(result.Diagnostics[:0:0], result.Diagnostics...)
	sort.SliceStable(diagnostics, func(left int, right int) bool {
		return diagnostics[left].Range.Pos() < diagnostics[right].Range.Pos()
	})
	text := result.SourceFile.Text()
	var got []string
	for _, diagnostic := range diagnostics {
		if diagnostic.Message.Id != correctnessNoImportCycleLoadTimeReadId {
			t.Fatalf("unexpected message id %q", diagnostic.Message.Id)
		}
		if len(diagnostic.Fixes) != 0 || len(diagnostic.Suggestions) != 0 {
			t.Fatalf("expected no fix and no suggestion, got %d and %d", len(diagnostic.Fixes), len(diagnostic.Suggestions))
		}
		got = append(got, strings.TrimSpace(text[diagnostic.Range.Pos():diagnostic.Range.End()]))
	}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("findings %q, want %q", got, want)
	}
}

const (
	correctnessNoImportCycleLoadTimeReadIdentifiersFile = "/repository/nexus/source/protocols/base/errors/BaseErrorIdentifiers.ts"
	correctnessNoImportCycleLoadTimeReadBaseErrorFile   = "/repository/nexus/source/protocols/base/errors/BaseError.ts"
	correctnessNoImportCycleLoadTimeReadRateLimiterFile = "/repository/nexus/source/modules/rate-limiter/RateLimiterModuleErrors.ts"
)

// correctnessNoImportCycleLoadTimeReadBaseErrors models the three-file cycle in Nexus's
// `source/protocols/base/errors/` in ahra on 2026-10-03, trimmed to what makes it a cycle: the
// identifiers file spreads every module's declarations at top level, the rate limiter's declarations
// import the identifier keys back for `isRateLimitErrorData`, and `BaseError.ts` imports them too for
// `isBaseErrorData`. `fixed` is the fix the doc comment names: the rate limiter's wire-shape guard
// moves to a file of its own, so the declarations file imports nothing back.
func correctnessNoImportCycleLoadTimeReadBaseErrors(fixed bool) map[string]string {
	rateLimitGuard := []string{
		"export interface RateLimitErrorDataInterface extends BaseErrorDataInterface {",
		"    readonly identifier: 'RateLimitExceeded';",
		"}",
		"export function isRateLimitErrorData(error: unknown): error is RateLimitErrorDataInterface {",
		"    if(!isBaseErrorData(error)) {",
		"        return false;",
		"    }",
		"    return (error as RateLimitErrorDataInterface).identifier === BaseErrorIdentifierKeys.RateLimitExceeded;",
		"}",
	}
	rateLimitGuardImports := []string{
		"import type { BaseErrorDataInterface } from '../../protocols/base/errors/BaseError';",
		"import { isBaseErrorData } from '../../protocols/base/errors/BaseError';",
		"import { BaseErrorIdentifierKeys } from '../../protocols/base/errors/BaseErrorIdentifiers';",
	}
	rateLimiterDeclarations := []string{
		"import type { BaseErrorDeclarationsType } from '../../protocols/base/errors/BaseErrorDeclaration';",
		"import { typeOnly } from '../../types/ObjectTypes';",
		"export const RateLimiterModuleErrors = {",
		"    RateLimitExceeded: { statusCode: 429, data: typeOnly<{ limit: number; remaining: number }>() },",
		"    RateLimiterNotConfigured: { statusCode: 500 },",
		"    RateLimitCheckFailed: {},",
		"} as const satisfies BaseErrorDeclarationsType;",
	}
	files := map[string]string{
		correctnessNoImportCycleLoadTimeReadIdentifiersFile: strings.Join([]string{
			"import { AccountModuleErrors } from '../../../modules/account/AccountModuleErrors';",
			"import { RateLimiterModuleErrors } from '../../../modules/rate-limiter/RateLimiterModuleErrors';",
			"import type { BaseErrorDeclarationsType } from './BaseErrorDeclaration';",
			"import { BaseErrors } from './BaseErrors';",
			"export const BaseErrorIdentifiers = {",
			"    ...BaseErrors,",
			"    ...AccountModuleErrors,",
			"    ...RateLimiterModuleErrors,",
			"} as const satisfies BaseErrorDeclarationsType;",
			"export type BaseErrorIdentifierType = keyof typeof BaseErrorIdentifiers;",
			"export const BaseErrorIdentifierKeys = Object.fromEntries(",
			"    Object.keys(BaseErrorIdentifiers).map(function(key) {",
			"        return [key, key];",
			"    }),",
			") as { [Key in BaseErrorIdentifierType]: Key };",
		}, "\n"),
		correctnessNoImportCycleLoadTimeReadBaseErrorFile: strings.Join([]string{
			"import { BaseErrorIdentifierKeys } from './BaseErrorIdentifiers';",
			"import type { BaseErrorIdentifierType } from './BaseErrorIdentifiers';",
			"export interface BaseErrorDataInterface {",
			"    readonly identifier: BaseErrorIdentifierType;",
			"}",
			"export function isBaseErrorData(error: unknown): error is BaseErrorDataInterface {",
			"    return typeof error === 'object' && error !== null && 'identifier' in error &&",
			"        typeof error.identifier === 'string' && error.identifier in BaseErrorIdentifierKeys;",
			"}",
			"export class BaseError extends Error {}",
		}, "\n"),
		"/repository/nexus/source/protocols/base/errors/BaseErrors.ts": strings.Join([]string{
			"import type { BaseErrorDeclarationsType } from './BaseErrorDeclaration';",
			"export const BaseErrors = {",
			"    InternalServerError: { statusCode: 500 },",
			"} as const satisfies BaseErrorDeclarationsType;",
		}, "\n"),
		"/repository/nexus/source/protocols/base/errors/BaseErrorDeclaration.ts": strings.Join([]string{
			"export interface BaseErrorDeclarationInterface {",
			"    readonly statusCode?: number;",
			"    readonly data?: unknown;",
			"}",
			"export type BaseErrorDeclarationsType = Readonly<Record<string, BaseErrorDeclarationInterface>>;",
		}, "\n"),
		"/repository/nexus/source/types/ObjectTypes.ts": strings.Join([]string{
			"export function typeOnly<Shape>(): Shape {",
			"    return undefined as Shape;",
			"}",
		}, "\n"),
		"/repository/nexus/source/modules/account/AccountModuleErrors.ts": strings.Join([]string{
			"import type { BaseErrorDeclarationsType } from '../../protocols/base/errors/BaseErrorDeclaration';",
			"export const AccountModuleErrors = {",
			"    AccountNotFound: { statusCode: 404 },",
			"} as const satisfies BaseErrorDeclarationsType;",
		}, "\n"),
	}
	if fixed {
		files[correctnessNoImportCycleLoadTimeReadRateLimiterFile] = strings.Join(rateLimiterDeclarations, "\n")
		files["/repository/nexus/source/modules/rate-limiter/RateLimitErrorData.ts"] = strings.Join(append(rateLimitGuardImports, rateLimitGuard...), "\n")
	} else {
		files[correctnessNoImportCycleLoadTimeReadRateLimiterFile] = strings.Join(append(append(rateLimitGuardImports, rateLimiterDeclarations...), rateLimitGuard...), "\n")
	}
	return files
}

// The real site: the spread of the one module in the cycle is reported, the two outside it are not.
func TestCorrectnessNoImportCycleLoadTimeReadFiresOnBaseErrorIdentifiers(t *testing.T) {
	t.Parallel()

	result := correctnessNoImportCycleLoadTimeReadRun(t, correctnessNoImportCycleLoadTimeReadBaseErrors(false), correctnessNoImportCycleLoadTimeReadIdentifiersFile)
	correctnessNoImportCycleLoadTimeReadExpect(t, result, "RateLimiterModuleErrors")
	if !strings.Contains(result.Diagnostics[0].Message.Description, "BaseErrorIdentifiers.ts → RateLimiterModuleErrors.ts → BaseErrorIdentifiers.ts") {
		t.Fatalf("the message should name the cycle, got %q", result.Diagnostics[0].Message.Description)
	}
}

// The other two members of the cycle only read the identifier keys inside functions.
func TestCorrectnessNoImportCycleLoadTimeReadStaysSilentOnTheRestOfTheBaseErrorCycle(t *testing.T) {
	t.Parallel()

	for _, subject := range []string{correctnessNoImportCycleLoadTimeReadRateLimiterFile, correctnessNoImportCycleLoadTimeReadBaseErrorFile} {
		rule_testing.ExpectClean(t, correctnessNoImportCycleLoadTimeReadRun(t, correctnessNoImportCycleLoadTimeReadBaseErrors(false), subject))
	}
}

// After the fix the guard lives beside the declarations rather than in them, nothing imports back
// into the identifiers file from the rate limiter's declarations, and the spread is silent.
func TestCorrectnessNoImportCycleLoadTimeReadStaysSilentOnFixedBaseErrorIdentifiers(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, correctnessNoImportCycleLoadTimeReadRun(t, correctnessNoImportCycleLoadTimeReadBaseErrors(true), correctnessNoImportCycleLoadTimeReadIdentifiersFile))
}

const (
	correctnessNoImportCycleLoadTimeReadSubjectFile = "/repository/source/A.ts"
	correctnessNoImportCycleLoadTimeReadOtherFile   = "/repository/source/B.ts"
)

// correctnessNoImportCycleLoadTimeReadOther is B.ts: every kind of binding, plus a back import of A
// read inside a function, which is a runtime edge and so closes the cycle.
var correctnessNoImportCycleLoadTimeReadOther = []string{
	"import { fromA } from './A';",
	"export const Limit = 10;",
	"export let counter = 0;",
	"export var legacy = 1;",
	"export const { destructured } = { destructured: 1 };",
	"export class Base {}",
	"export enum Color { Red }",
	"export const enum Flag { On = 1 }",
	"export declare const ambient: number;",
	"export function helper(): number { return 1; }",
	"export const make = function<Value>(value: Value): Value { return value; };",
	"export const tag = function(value: unknown, context: ClassDecoratorContext): void {};",
	"export default class DefaultThing {}",
	"export function readA(): number { return fromA; }",
}

// correctnessNoImportCycleLoadTimeReadSubject is A.ts: its own export B reads, the names the cases
// call, then the case.
func correctnessNoImportCycleLoadTimeReadSubject(lines ...string) string {
	return strings.Join(append([]string{
		"export const fromA = 1;",
		"export type FromAType = number;",
		"declare function use(value: unknown): void;",
	}, lines...), "\n")
}

// correctnessNoImportCycleLoadTimeReadPair runs the rule on A.ts against B.ts.
func correctnessNoImportCycleLoadTimeReadPair(t *testing.T, subject []string, other []string) rule_testing.Result {
	t.Helper()
	return correctnessNoImportCycleLoadTimeReadRun(t, map[string]string{
		correctnessNoImportCycleLoadTimeReadSubjectFile: correctnessNoImportCycleLoadTimeReadSubject(subject...),
		correctnessNoImportCycleLoadTimeReadOtherFile:   strings.Join(other, "\n"),
	}, correctnessNoImportCycleLoadTimeReadSubjectFile)
}

func TestCorrectnessNoImportCycleLoadTimeReadFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines []string
		want  []string
	}{
		{"a const in an initializer", []string{"import { Limit } from './B';", "export const doubled = Limit * 2;"}, []string{"Limit"}},
		{"a class extended", []string{"import { Base } from './B';", "export class Derived extends Base {}"}, []string{"Base"}},
		{"a namespace member", []string{"import * as b from './B';", "export const doubled = b.Limit * 2;"}, []string{"b.Limit"}},
		{"typeof, which throws in the dead zone", []string{"import { Limit } from './B';", "export const kind = typeof Limit;"}, []string{"Limit"}},
		{"a shorthand property", []string{"import { Limit } from './B';", "export const bag = { Limit };"}, []string{"Limit"}},
		{"export default of an expression", []string{"import { Limit } from './B';", "export default Limit;"}, []string{"Limit"}},
		{"a static field", []string{"import { Limit } from './B';", "export class Holder { static limit = Limit; }"}, []string{"Limit"}},
		{"a static block", []string{"import { Limit } from './B';", "export class Holder { static { use(Limit); } }"}, []string{"Limit"}},
		{"a class decorator", []string{"import { tag } from './B';", "@tag", "export class Tagged {}"}, []string{"tag"}},
		{"a condition at top level", []string{"import { Limit } from './B';", "if(Limit > 1) {", "    use(1);", "}"}, []string{"Limit"}},
		{"let, var, a destructured const and an enum", []string{"import { counter, legacy, destructured, Color } from './B';", "use(counter);", "use(legacy);", "use(destructured);", "use(Color.Red);"}, []string{"counter", "legacy", "destructured", "Color"}},
		{"a default import of a class", []string{"import DefaultThing from './B';", "use(DefaultThing);"}, []string{"DefaultThing"}},
		{"a renamed import", []string{"import { Limit as Ceiling } from './B';", "use(Ceiling);"}, []string{"Ceiling"}},
		{"an instantiation expression", []string{"import { make } from './B';", "export const makeNumber = make<number>;"}, []string{"make"}},
		{"two reads", []string{"import { Limit } from './B';", "use(Limit);", "export const a = [Limit];"}, []string{"Limit", "Limit"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			correctnessNoImportCycleLoadTimeReadExpect(t, correctnessNoImportCycleLoadTimeReadPair(t, testCase.lines, correctnessNoImportCycleLoadTimeReadOther), testCase.want...)
		})
	}
}

func TestCorrectnessNoImportCycleLoadTimeReadStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines []string
	}{
		{"a function body", []string{"import { Limit } from './B';", "export function doubled(): number { return Limit * 2; }"}},
		{"an arrow function", []string{"import { Limit } from './B';", "export const doubled = () => Limit * 2;"}},
		{"a function expression", []string{"import { Limit } from './B';", "export const doubled = function() { return Limit * 2; };"}},
		{"methods, accessors, a constructor and an instance field", []string{
			"import { Limit } from './B';",
			"export class Holder {",
			"    limit = Limit;",
			"    get value(): number { return Limit; }",
			"    constructor() { use(Limit); }",
			"    read(): number { return Limit; }",
			"}",
		}},
		{"an object method", []string{"import { Limit } from './B';", "export const holder = { read() { return Limit; } };"}},
		{"a parameter default", []string{"import { Limit } from './B';", "export function read(value = Limit): number { return value; }"}},
		{"a hoisted function called at load", []string{"import { helper } from './B';", "export const one = helper();"}},
		{"type positions", []string{"import { Limit, Base } from './B';", "export const typed: typeof Limit = 10;", "export const cast = 10 as typeof Limit;", "export type Instance = Base;"}},
		{"an export that forwards the binding", []string{"import { Limit } from './B';", "export { Limit };"}},
		{"an ambient declaration", []string{"import { ambient } from './B';", "use(ambient);"}},
		{"a const enum", []string{"import { Flag, Limit } from './B';", "use(Flag.On);", "export function read(): number { return Limit; }"}},
		{"a local shadowing the import", []string{"import { Limit } from './B';", "{", "    const Limit = 3;", "    use(Limit);", "}", "export function read(): number { return Limit; }"}},
		{"the namespace object itself", []string{"import * as b from './B';", "use(b);"}},
		{"a module outside the cycle", []string{"import { Outside } from './C';", "import { Limit } from './B';", "use(Outside);", "export function read(): number { return Limit; }"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := correctnessNoImportCycleLoadTimeReadRun(t, map[string]string{
				correctnessNoImportCycleLoadTimeReadSubjectFile: correctnessNoImportCycleLoadTimeReadSubject(testCase.lines...),
				correctnessNoImportCycleLoadTimeReadOtherFile:   strings.Join(correctnessNoImportCycleLoadTimeReadOther, "\n"),
				"/repository/source/C.ts":                       "export const Outside = 1;",
			}, correctnessNoImportCycleLoadTimeReadSubjectFile)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// correctnessNoImportCycleLoadTimeReadBackEdge is a B.ts whose only tie back to A.ts is the given
// lines, so whether A's top-level read is reported is decided by whether those lines execute.
func correctnessNoImportCycleLoadTimeReadBackEdge(lines ...string) []string {
	// Last, so a directive among the lines stays the file's first statement.
	return append(append([]string{}, lines...), "export const Limit = 10;")
}

// The back edge executes, so A.ts's read of Limit at load is in a cycle.
func TestCorrectnessNoImportCycleLoadTimeReadFiresOnEveryRuntimeBackEdge(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines []string
	}{
		{"a side-effect import", []string{"import './A';"}},
		{"export star", []string{"export * from './A';"}},
		{"export star as a namespace", []string{"export * as a from './A';"}},
		{"a named re-export of a value", []string{"export { fromA } from './A';"}},
		{"a value read in a function", []string{"import { fromA } from './A';", "export function read(): number { return fromA; }"}},
		{"a value read through a shorthand property", []string{"import { fromA } from './A';", "export function bag() { return { fromA }; }"}},
		{"a namespace import read", []string{"import * as a from './A';", "export function read(): number { return a.fromA; }"}},
		{"a mixed import whose value binding is read", []string{"import { type FromAType, fromA } from './A';", "export function read(): FromAType { return fromA; }"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := correctnessNoImportCycleLoadTimeReadPair(t, []string{"import { Limit } from './B';", "export const doubled = Limit * 2;"}, correctnessNoImportCycleLoadTimeReadBackEdge(testCase.lines...))
			correctnessNoImportCycleLoadTimeReadExpect(t, result, "Limit")
		})
	}
}

// The back edge is erased or deferred, so there is no runtime cycle and the read is safe.
func TestCorrectnessNoImportCycleLoadTimeReadStaysSilentWithoutARuntimeBackEdge(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines []string
	}{
		{"import type", []string{"import type { FromAType } from './A';", "export const typed: FromAType = 1;"}},
		{"an inline type specifier", []string{"import { type FromAType } from './A';", "export const typed: FromAType = 1;"}},
		{"export type", []string{"export type { FromAType } from './A';"}},
		{"a named re-export of a type", []string{"export { FromAType } from './A';"}},
		{"a value import used only in types, which TypeScript elides", []string{"import { fromA } from './A';", "export type Copy = typeof fromA;"}},
		{"a value import whose only use is a shadowing parameter", []string{"import { fromA } from './A';", "export function read(fromA: number): number { return fromA; }"}},
		{"a dynamic import", []string{"export async function later(): Promise<unknown> { return import('./A'); }"}},
		{"a server actions module", []string{"'use server';", "import { fromA } from './A';", "export async function read(): Promise<number> { return fromA; }"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := correctnessNoImportCycleLoadTimeReadPair(t, []string{"import { Limit } from './B';", "export const doubled = Limit * 2;"}, correctnessNoImportCycleLoadTimeReadBackEdge(testCase.lines...))
			rule_testing.ExpectClean(t, result)
		})
	}
}

// A read through a barrel is a read of the declaring file's binding, and the barrel's re-export is a
// runtime edge, so A → Barrel → B → A is the cycle.
func TestCorrectnessNoImportCycleLoadTimeReadFollowsABarrel(t *testing.T) {
	t.Parallel()

	files := map[string]string{
		correctnessNoImportCycleLoadTimeReadSubjectFile: correctnessNoImportCycleLoadTimeReadSubject("import { Limit } from './Barrel';", "export const doubled = Limit * 2;"),
		"/repository/source/Barrel.ts":                  "export { Limit } from './B';",
		correctnessNoImportCycleLoadTimeReadOtherFile:   strings.Join(correctnessNoImportCycleLoadTimeReadBackEdge("import './A';"), "\n"),
	}
	result := correctnessNoImportCycleLoadTimeReadRun(t, files, correctnessNoImportCycleLoadTimeReadSubjectFile)
	correctnessNoImportCycleLoadTimeReadExpect(t, result, "Limit")
	if !strings.Contains(result.Diagnostics[0].Message.Description, "A.ts → Barrel.ts → B.ts → A.ts") {
		t.Fatalf("the message should name the cycle through the barrel, got %q", result.Diagnostics[0].Message.Description)
	}
}

// A barrel in the cycle whose binding is declared outside it: the barrel is mid-load, but the binding
// lives in a file that cannot be, because it imports nothing back.
func TestCorrectnessNoImportCycleLoadTimeReadStaysSilentWhenTheDeclarationIsOutsideTheCycle(t *testing.T) {
	t.Parallel()

	files := map[string]string{
		correctnessNoImportCycleLoadTimeReadSubjectFile: correctnessNoImportCycleLoadTimeReadSubject("import { Outside } from './Barrel';", "export const doubled = Outside * 2;"),
		"/repository/source/Barrel.ts":                  "export { Outside } from './C';\nimport './A';",
		"/repository/source/C.ts":                       "export const Outside = 1;",
	}
	rule_testing.ExpectClean(t, correctnessNoImportCycleLoadTimeReadRun(t, files, correctnessNoImportCycleLoadTimeReadSubjectFile))
}

// A re-export of a name another module exported with `export type` is elided like any type, so it is
// no edge even though the name lands on a value: B.ts reaches Mid.ts only through it, and the cycle
// A → B → Mid → A does not run.
func TestCorrectnessNoImportCycleLoadTimeReadStaysSilentThroughATypeOnlyReExport(t *testing.T) {
	t.Parallel()

	files := map[string]string{
		correctnessNoImportCycleLoadTimeReadSubjectFile: correctnessNoImportCycleLoadTimeReadSubject("import { Limit } from './B';", "export const doubled = Limit * 2;"),
		correctnessNoImportCycleLoadTimeReadOtherFile:   "export { fromA } from './Mid';\nexport const Limit = 10;",
		"/repository/source/Mid.ts":                     "import './A';\nexport type { fromA } from './A';",
	}
	rule_testing.ExpectClean(t, correctnessNoImportCycleLoadTimeReadRun(t, files, correctnessNoImportCycleLoadTimeReadSubjectFile))
}

// correctnessNoImportCycleLoadTimeReadJsxPair runs the rule on a JSX pair modelled on the component
// cycles in ahra (`TaskRowExpandable.tsx` and `TaskRowExpandedChildren.tsx` render each other).
// `element` is A.tsx's top-level use of B.tsx's component; `component` is how B.tsx declares it.
func correctnessNoImportCycleLoadTimeReadJsxPair(t *testing.T, element string, component string) rule_testing.Result {
	t.Helper()
	return rule_testing.RunTypedFilesWithSetup(t, CorrectnessNoImportCycleLoadTimeRead, map[string]string{
		"/repository/source/A.tsx": strings.Join([]string{
			"import { Child } from './B';",
			"declare global { namespace JSX { interface IntrinsicElements { div: unknown } } }",
			"export function Parent(): unknown { return <div><Child /></div>; }",
			element,
		}, "\n"),
		"/repository/source/B.tsx": strings.Join([]string{
			"import { Parent } from './A';",
			component,
			"export function Nested(): unknown { return <Parent />; }",
		}, "\n"),
	}, "/repository/source/A.tsx", func(directory string) {
		config := `{"compilerOptions": {"strict": true, "target": "ES2022", "lib": ["ES2022"], "jsx": "preserve", "moduleDetection": "force", "types": []}, "include": ["**/*.tsx"]}`
		if err := os.WriteFile(filepath.Join(directory, "tsconfig.json"), []byte(config), 0o644); err != nil {
			t.Fatalf("writing the JSX tsconfig: %v", err)
		}
	})
}

// Components that render each other inside function bodies are the common, harmless cycle.
func TestCorrectnessNoImportCycleLoadTimeReadStaysSilentOnComponentsRenderingEachOther(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, correctnessNoImportCycleLoadTimeReadJsxPair(t, "", "export const Child = function(): unknown { return null; };"))
}

// An element built at load reads its component then. The closing tag is not a second read.
func TestCorrectnessNoImportCycleLoadTimeReadFiresOnAnElementBuiltAtLoad(t *testing.T) {
	t.Parallel()

	result := correctnessNoImportCycleLoadTimeReadJsxPair(t, "export const placeholder = <Child></Child>;", "export const Child = function(): unknown { return null; };")
	correctnessNoImportCycleLoadTimeReadExpect(t, result, "Child")
}

// The same element is safe when the component is a function declaration, which is hoisted.
func TestCorrectnessNoImportCycleLoadTimeReadStaysSilentOnAnElementOfAHoistedComponent(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, correctnessNoImportCycleLoadTimeReadJsxPair(t, "export const placeholder = <Child></Child>;", "export function Child(): unknown { return null; }"))
}
