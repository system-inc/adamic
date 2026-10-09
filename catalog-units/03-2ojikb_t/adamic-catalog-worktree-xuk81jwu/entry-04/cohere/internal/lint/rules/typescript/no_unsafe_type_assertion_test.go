package typescript

import (
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const noUnsafeTypeAssertionFile = "/repository/source/Asserting.ts"

func noUnsafeTypeAssertionCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// TestNoUnsafeTypeAssertionStaysSilent is upstream's five passing cases verbatim, plus shapes
// upstream does not write whose silence was measured against the installed 8.67.0 build.
//
// Every case runs in its OWN program, which for this rule is load-bearing rather than tidy. The
// first thing it asks is whether two types are REFERENCE identical, and the checker interns one
// type per shape per program. Several of upstream's cases each declare their own `T`; batched into
// one program those merge, identity starts answering true, and cases that really report come back
// clean. RunTyped is one file per call, so the fixtures are safe by construction and the oracle
// used to measure them had to be built that way on purpose.
func TestNoUnsafeTypeAssertionStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []string{
		// Upstream's valid list, byte for byte.
		"\ntype Obj = { foo: string };\nfunction func<T extends Obj>(a: T) {\n  const b = a as T;\n}\n      ",
		"\nfunction parameterExtendsOtherParameter<T extends string | number, V extends T>(\n  x: T,\n  y: V,\n) {\n  y as T;\n}\n      ",
		"\nfunction parameterExtendsUnconstrainedParameter<T, V extends T>(x: T, y: V) {\n  y as T;\n}\n      ",
		"\nfunction unconstrainedToUnknown<T>(x: T) {\n  x as unknown;\n}\n      ",
		"\nfunction stringToWider<T extends string>(x: T) {\n  x as number | string; // allowed\n}\n      ",

		// Measured silent on the installed build. Each pins a step of the decision sequence that
		// upstream's corpus does not reach, and each was chosen by naming the input that separates
		// a surviving mutant from the original rather than by guessing.

		// The object literal widening. `{ foo: 'hi', bar: 1 }` is NOT assignable to `{ foo: string }`
		// while it carries its excess-property information, so without the widening this reports
		// `unsafeTypeAssertion`. Measured: removing the widening makes exactly this case report,
		// and upstream is silent on it. Its excess-property-free sibling below is silent either
		// way, so the pair is what isolates the widening rather than the assertion.
		"const o = { foo: 'hi', bar: 1 } as { foo: string };\n",
		"const p = { foo: 'hi' } as { foo: string };\n",

		// The identity early return, which is EQUIVALENT to falling through and is kept anyway.
		// Twelve identity shapes were probed against a mutant that disables it, including
		// `any as any`, `Set<any> as Set<any>`, `T as T`, `never as never` and an error type
		// asserted to itself, and none of them changed verdict. The argument is one sentence: when
		// the two types are the same object, IsUnsafeAssignment answers false in both directions
		// and isTypeAssignableTo answers true by reflexivity, so both versions reach silence by
		// different routes. The check stays because it is upstream's and because it saves two
		// checker calls on a rule that runs on every assertion in the tree, but no fixture can
		// prove it and this comment is the record instead.
		"declare const x: any;\nconst y = x as any;\n",
		"declare const x: string;\nconst y = x as string;\n",
	}
	for index, sourceText := range cases {
		t.Run(noUnsafeTypeAssertionCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoUnsafeTypeAssertion,
				noUnsafeTypeAssertionFile, sourceText))
		})
	}
}

// TestNoUnsafeTypeAssertionFires is upstream's seventeen reporting cases verbatim, with the message
// id, the rendered message text, and the span of every finding.
//
// The rendered text is asserted by equality rather than by a substring test, because four of the
// five messages interpolate a type name through the checker and one of them interpolates it TWICE.
// A substring predicate cannot see a wrong name in either slot, and the type name is the only part
// of this rule's output that a reader uses to find the site.
func TestNoUnsafeTypeAssertionFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText   string
		wantFindings []struct {
			id      string
			message string
			span    string
		}
	}{
		{
			sourceText: "\ntype Obj = { foo: string };\nfunction func<T extends Obj>() {\n  const myObj = { foo: 'hi' } as T;\n}\n        ",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "unsafeTypeAssertionAssignableToConstraint",
					message: "Unsafe type assertion: the original type is assignable to the constraint of type 'T', but 'T' could be instantiated with a different subtype of its constraint.",
					span:    "{ foo: 'hi' } as T",
				},
			},
		},
		{
			sourceText: "\ntype Obj = { foo: string };\nfunction func<T extends Obj>() {\n  const o: Obj = { foo: 'hi' };\n  const myObj = o as T;\n}\n        ",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "unsafeTypeAssertionAssignableToConstraint",
					message: "Unsafe type assertion: the original type is assignable to the constraint of type 'T', but 'T' could be instantiated with a different subtype of its constraint.",
					span:    "o as T",
				},
			},
		},
		{
			sourceText: "\nexport function myfunc<CustomObjectT extends string>(\n  input: number,\n): CustomObjectT {\n  const newCustomObject = input as CustomObjectT;\n  return newCustomObject;\n}\n        ",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "unsafeTypeAssertion",
					message: "Unsafe type assertion: type 'CustomObjectT' is more narrow than the original type.",
					span:    "input as CustomObjectT",
				},
			},
		},
		{
			sourceText: "\nfunction unknownConstraint<T extends unknown>(x: T, y: string) {\n  y as T; // banned; generic arbitrary subtype\n}\n        ",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "unsafeTypeAssertionAssignableToConstraint",
					message: "Unsafe type assertion: the original type is assignable to the constraint of type 'T', but 'T' could be instantiated with a different subtype of its constraint.",
					span:    "y as T",
				},
			},
		},
		{
			sourceText: "\nfunction unconstrained<T>(x: T, y: string) {\n  y as T;\n}\n        ",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "unsafeToUnconstrainedTypeAssertion",
					message: "Unsafe type assertion: 'T' could be instantiated with an arbitrary type which could be unrelated to the original type.",
					span:    "y as T",
				},
			},
		},
		{
			sourceText: "\n// constraint of any functions like constraint of `unknown`\n// (even the TS error message has this verbiage)\nfunction anyConstraint<T extends any>(x: T, y: string) {\n  y as T; // banned; generic arbitrary subtype\n}\n        ",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "unsafeTypeAssertionAssignableToConstraint",
					message: "Unsafe type assertion: the original type is assignable to the constraint of type 'T', but 'T' could be instantiated with a different subtype of its constraint.",
					span:    "y as T",
				},
			},
		},
		{
			sourceText: "\nfunction constraintWiderThanUncastType<T extends string | number>(\n  x: T,\n  y: string,\n) {\n  y as T; // banned; assignable to constraint\n}\n        ",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "unsafeTypeAssertionAssignableToConstraint",
					message: "Unsafe type assertion: the original type is assignable to the constraint of type 'T', but 'T' could be instantiated with a different subtype of its constraint.",
					span:    "y as T",
				},
			},
		},
		{
			sourceText: "\nfunction constraintEqualUncastType<T extends string>(x: T, y: string) {\n  y as T; // banned; assignable to constraint\n}\n        ",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "unsafeTypeAssertionAssignableToConstraint",
					message: "Unsafe type assertion: the original type is assignable to the constraint of type 'T', but 'T' could be instantiated with a different subtype of its constraint.",
					span:    "y as T",
				},
			},
		},
		{
			sourceText: "\nfunction constraintNarrowerThanUncastType<T extends string>(\n  x: T,\n  y: string | number,\n) {\n  y as T; // banned; *not* assignable to constraint\n}\n        ",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "unsafeTypeAssertion",
					message: "Unsafe type assertion: type 'T' is more narrow than the original type.",
					span:    "y as T",
				},
			},
		},
		{
			sourceText: "\nfunction assertFromAny<T extends string | number>(x: T, y: any) {\n  y as T; // banned; just an `any` complaint. Not a generic subtype.\n}\n        ",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "unsafeOfAnyTypeAssertion",
					message: "Unsafe assertion from `any` detected: consider using type guards or a safer assertion.",
					span:    "y as T",
				},
			},
		},
		{
			sourceText: "\nfunction parameterExtendsOtherParameter<T extends string | number, V extends T>(\n  x: T,\n  y: V,\n) {\n  x as V; // banned; assignable to constraint\n}\n        ",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "unsafeTypeAssertionAssignableToConstraint",
					message: "Unsafe type assertion: the original type is assignable to the constraint of type 'V', but 'V' could be instantiated with a different subtype of its constraint.",
					span:    "x as V",
				},
			},
		},
		{
			sourceText: "\nfunction parameterExtendsUnconstrainedParameter<T, V extends T>(x: T, y: V) {\n  x as V; // banned; unconstrained arbitrary type\n}\n        ",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "unsafeToUnconstrainedTypeAssertion",
					message: "Unsafe type assertion: 'V' could be instantiated with an arbitrary type which could be unrelated to the original type.",
					span:    "x as V",
				},
			},
		},
		{
			sourceText: "\nfunction twoUnconstrained<T, V>(x: T, y: V) {\n  y as T;\n}\n        ",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "unsafeToUnconstrainedTypeAssertion",
					message: "Unsafe type assertion: 'T' could be instantiated with an arbitrary type which could be unrelated to the original type.",
					span:    "y as T",
				},
			},
		},
		{
			sourceText: "\nfunction toNarrower<T>(x: T, y: string) {\n  x as string; // banned; ordinary 'string' narrower than 'T'.\n}\n        ",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "unsafeTypeAssertion",
					message: "Unsafe type assertion: type 'string' is more narrow than the original type.",
					span:    "x as string",
				},
			},
		},
		{
			sourceText: "\nfunction unconstrainedToAny<T>(x: T) {\n  x as any;\n}\n        ",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "unsafeToAnyTypeAssertion",
					message: "Unsafe assertion to `any` detected: consider using a more specific type to ensure safety.",
					span:    "x as any",
				},
			},
		},
		{
			sourceText: "\nfunction stringToAny<T extends string>(x: T) {\n  x as any;\n}\n        ",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "unsafeToAnyTypeAssertion",
					message: "Unsafe assertion to `any` detected: consider using a more specific type to ensure safety.",
					span:    "x as any",
				},
			},
		},
		{
			sourceText: "\nfunction stringToNarrower<T extends string>(x: T) {\n  x as 'a' | 'b';\n}\n        ",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "unsafeTypeAssertion",
					message: "Unsafe type assertion: type '\"a\" | \"b\"' is more narrow than the original type.",
					span:    "x as 'a' | 'b'",
				},
			},
		},

		// The `unknown` to `any` special case, tested ahead of everything else because it is the one
		// direction the assignability test below would call safe: `unknown` really is assignable
		// to `any`. Removing that branch takes this case silent, measured, and upstream reports it.
		// Nothing in upstream's corpus writes the shape.
		{
			sourceText: "declare const u: unknown;\nconst v = u as any;\n",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "unsafeToAnyTypeAssertion",
					message: "Unsafe assertion to `any` detected: consider using a more specific type to ensure safety.",
					span:    "u as any",
				},
			},
		},

		// The ANGLE BRACKET spelling of an assertion, which upstream's corpus never writes: every one of
		// its twenty two cases uses `as`. Both spellings mean the same thing and upstream matches both
		// node types, so a port registering only the `as` listener passes the entire imported corpus
		// while being silent on half the syntax. Measured reporting on the installed build.
		{
			sourceText: "declare const x: string | number;\nconst y = <string>x;\n",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "unsafeTypeAssertion",
					message: "Unsafe type assertion: type 'string' is more narrow than the original type.",
					span:    "<string>x",
				},
			},
		},
		// The same spelling reaching a different message, so the second listener is pinned on more than
		// one arm of the decision sequence.
		{
			sourceText: "declare const u: unknown;\nconst v = <any>u;\n",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "unsafeToAnyTypeAssertion",
					message: "Unsafe assertion to `any` detected: consider using a more specific type to ensure safety.",
					span:    "<any>u",
				},
			},
		},
		// An ERROR type, which renders as error typed rather than `any`. Upstream draws that
		// distinction deliberately: the type is any-shaped only because something else failed to
		// compile, and calling it `any` would send the reader looking for an annotation that is not
		// there. The corpus writes no unresolved name anywhere, so a port dropping the distinction
		// passes every imported case while telling readers the wrong thing about a whole class of site.
		{
			sourceText: "declare const x: NotDeclared;\nconst y = x as string;\n",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "unsafeOfAnyTypeAssertion",
					message: "Unsafe assertion from error typed detected: consider using type guards or a safer assertion.",
					span:    "x as string",
				},
			},
		},
		// An object type that is NOT an object literal, which is what keeps the widening narrow. Widening
		// every object type rather than only literals would change what this case asks the checker.
		// The rendered type carries a trailing semicolon the source does not, because the name comes
		// from the checker rather than from the source text. That is why the message is asserted by
		// equality rather than by a substring of what was written.
		{
			sourceText: "declare const d: Date;\nconst e = d as { getTime(): number; extra: string };\n",
			wantFindings: []struct {
				id      string
				message string
				span    string
			}{
				{
					id:      "unsafeTypeAssertion",
					message: "Unsafe type assertion: type '{ getTime(): number; extra: string; }' is more narrow than the original type.",
					span:    "d as { getTime(): number; extra: string }",
				},
			},
		},
	}
	for index, testCase := range cases {
		t.Run(noUnsafeTypeAssertionCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnsafeTypeAssertion, noUnsafeTypeAssertionFile,
				testCase.sourceText)

			wantIds := make([]string, 0, len(testCase.wantFindings))
			for _, want := range testCase.wantFindings {
				wantIds = append(wantIds, want.id)
			}
			rule_testing.ExpectFindings(t, result, wantIds...)

			// RunTyped writes the fixture trimmed, so the span slice is against that text rather
			// than against the Go literal above.
			onDisk := strings.TrimSpace(testCase.sourceText) + "\n"
			for findingIndex, want := range testCase.wantFindings {
				diagnostic := result.Diagnostics[findingIndex]
				if diagnostic.Message.Description != want.message {
					t.Fatalf("finding %d message: expected %q, got %q", findingIndex, want.message,
						diagnostic.Message.Description)
				}
				gotSpan := onDisk[diagnostic.Range.Pos():diagnostic.Range.End()]
				if gotSpan != want.span {
					t.Fatalf("finding %d span: expected %q, got %q", findingIndex, want.span, gotSpan)
				}
			}
		})
	}
}

// TestNoUnsafeTypeAssertionRequiresTheTypedHarness pins that this rule cannot work without a checker
// and declines rather than crashing when handed none.
//
// Unreachable from every other test here, because RunTyped always supplies a live checker, so a
// mutant neutralizing the guard survives the whole fixture set. The plain harness is the only
// instrument that sees it, and the risk it covers is not theoretical for this rule: the body
// dereferences the checker four times on its first path.
func TestNoUnsafeTypeAssertionRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !NoUnsafeTypeAssertion.NeedsTypeChecker {
		t.Fatal("the rule must declare NeedsTypeChecker: every step of its judgment is a type question")
	}

	sourceText := "declare const x: string | number;\nconst y = x as string;\n"

	rule_testing.ExpectClean(t, rule_testing.Run(t, NoUnsafeTypeAssertion,
		noUnsafeTypeAssertionFile, sourceText))

	// The control, so the silence above is the guard declining rather than the rule being unable to
	// see this shape at all.
	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoUnsafeTypeAssertion,
		noUnsafeTypeAssertionFile, sourceText), "unsafeTypeAssertion")
}
