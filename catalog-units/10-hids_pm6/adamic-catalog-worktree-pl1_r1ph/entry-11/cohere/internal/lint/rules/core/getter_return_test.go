package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// getterReturnFile is where the fixtures live: a TypeScript file in a real program, since the rule
// asks the checker whether a descriptor's `Object` is the global, and since it checks TypeScript files
// as ESLint does under the house config (#mnmx9s4). It ran against a `.js` name while the rule
// declined TypeScript, and `TestGetterReturnChecksTypeScript` below pins that it no longer does.
const getterReturnFile = "Getter.ts"

// The corpus is oxc's, copied rather than rewritten.
//
// Every case below is verbatim from `oxc/crates/oxc_linter/src/rules/eslint/getter_return.rs`. The
// extractor reports two Tester blocks: block 1 carries 47 pass and 38 fail and is snapshotted, and
// block 2 carries the single TypeScript pass case and is not. Cases extracted came to 48 pass
// because block 2's case joins them. The snapshot records 38 diagnostics from those 38 fail inputs,
// so one finding per input is measured here rather than assumed.
//
// Copied because a fixture a porter invents encodes the same belief as the port, and the case that
// catches a bug is the one nobody would think to write. Two of these earned that outright: the
// try/catch pair and the loop pair, both of which a plausible first implementation gets backwards.
func TestGetterReturnFires(t *testing.T) {
	t.Parallel()

	// ESLint's id: `expected` for a getter with no return of its own, or at a bare `return;`, and
	// `expectedAlways` at the head of one that returns on some paths and falls off the end on another
	cases := []struct {
		name       string
		sourceText string
		id         string
	}{
		{"an empty object getter", "var foo = { get bar() {} };", "expected"},
		{"an empty getter with a newline in its head", "var foo = { get\n bar () {} };", "expected"},
		{"an if with no else", "var foo = { get bar(){if(baz) {return true;}} };", "expectedAlways"},
		{"a return inside a nested function only", "var foo = { get bar() { ~function () {return true;}} };", "expected"},
		{"a bare return with allowImplicit off", "var foo = { get bar() { return; } };", "expected"},
		{"an empty class getter", "class foo { get bar(){} }", "expected"},
		{"a static getter with a newline in its head", "var foo = class {\n  static get\nbar(){} }", "expected"},
		{"a class getter whose if has no else", "class foo { get bar(){ if (baz) { return true; }}}", "expectedAlways"},
		{"a class getter returning only from a nested function", "class foo { get bar(){ ~function () { return true; }()}}", "expected"},
		{"an empty defineProperty getter", "Object.defineProperty(foo, 'bar', { get: function (){}});", "expected"},
		{"an empty named defineProperty getter", "Object.defineProperty(foo, 'bar', { get: function getfoo (){}});", "expected"},
		{"an empty defineProperty shorthand getter", "Object.defineProperty(foo, 'bar', { get(){} });", "expected"},
		{"an empty defineProperty arrow getter", "Object.defineProperty(foo, 'bar', { get: () => {}});", "expected"},
		{"a defineProperty getter whose if has no else", "Object.defineProperty(foo, \"bar\", { get: function (){if(bar) {return true;}}});", "expectedAlways"},
		{"a defineProperty getter returning only from a nested function", "Object.defineProperty(foo, \"bar\", { get: function (){ ~function () { return true; }()}});", "expected"},
		{"an empty Reflect.defineProperty getter", "Reflect.defineProperty(foo, 'bar', { get: function (){}});", "expected"},
		{"an empty Object.create getter", "Object.create(foo, { bar: { get: function() {} } })", "expected"},
		{"an empty Object.create shorthand getter", "Object.create(foo, { bar: { get() {} } })", "expected"},
		{"an empty Object.create arrow getter", "Object.create(foo, { bar: { get: () => {} } })", "expected"},
		{"an optional-chained defineProperty", "Object?.defineProperty(foo, 'bar', { get: function (){} });", "expected"},
		{"a parenthesized optional-chained defineProperty", "(Object?.defineProperty)(foo, 'bar', { get: function (){} });", "expected"},
		// A quoted `get` key, written here. Every descriptor case upstream ships spells the key as
		// the bare identifier `get`, so the arm of the key reader that handles a string literal was
		// never exercised and a mutant blanking it survived the whole corpus. `{ 'get': fn }` is the
		// same property as `{ get: fn }` and has to be read the same way.
		{"a descriptor with a quoted get key", "Object.defineProperty(foo, 'bar', { 'get': function (){} });", "expected"},
		// Two cases written here rather than imported, and they are what keeps the upward paren walk
		// alive. Every parenthesized case upstream ships parenthesizes the CALLEE, and that is the
		// downward walk's job. A mutant neutering the upward walk survived the entire corpus.
		//
		// The upward walk exists for a parenthesized DESCRIPTOR, where the parentheses sit between
		// the object literal and the call argument list, so the parent chain gains a level and the
		// descriptor stops being found without it.
		{"a parenthesized descriptor argument", "Object.defineProperty(foo, 'bar', ({ get: function (){} }));", "expected"},
		{"a parenthesized descriptor map argument", "Object.create(foo, ({ bar: { get: function (){} } }));", "expected"},
		{"a try whose catch does not return", "var foo = { get bar() { try { return a(); } catch {} } };", "expectedAlways"},
		{"a try whose catch and finally do not return", "var foo = { get bar() { try { return a(); } catch {  } finally {  } } };", "expectedAlways"},
		// A labeled block and a `with` block, both written here. Upstream ships neither, so mutants
		// making the rule stop unwrapping them survived the whole corpus. These are the fail halves:
		// a label or a `with` wrapping a block that does NOT exit is still a path through.
		{"a labeled block that does not return", "var foo = { get bar() { outer: { qux(); } } };", "expected"},
		// Three switch cases written here rather than imported, all three earned by surviving
		// mutants. Upstream's whole corpus contains exactly ONE switch, and it is the friendliest
		// possible shape: a default is present and every single clause returns. So three separate
		// discriminations in this branch were never exercised by it.
		//
		// This one kills the mutant that dropped the every-clause-exits gate. A clause that falls
		// out of the switch is a path through the getter, and the presence of a default says
		// nothing about it.
		{"a switch with a default whose other clause does not return", "var foo = { get bar(){ switch (baz) { case 1: qux(); break; default: return d; } } };", "expectedAlways"},
		// And this one kills the mutant that looked for a CaseClause instead of a DefaultClause when
		// deciding whether a default exists. Under that mutant a switch made only of case clauses
		// reads as having a default, every clause returns, and the getter is credited even though a
		// scrutinee matching neither case falls straight out. Upstream's single switch carries both
		// kinds, so it cannot tell the two predicates apart.
		//
		// The switch has to be the getter's ONLY statement. A first attempt at this case wrote it
		// as a clean fixture with a trailing `return c;` after the switch, which passes under the
		// mutant and under the original alike, because the block exits on the trailing return and
		// the switch's own verdict is never consulted. The mutant survived that fixture and the
		// re-score is the only reason it was noticed.
		{"a switch with no default at all", "var foo = { get bar(){ switch (baz) { case 1: return a; case 2: return b; } } };", "expectedAlways"},
		// And this one kills the mutant that let a TRAILING empty clause pass. `case 1:` with no
		// statements and nothing after it falls straight out of the switch. An empty clause in the
		// middle is a deliberate fallthrough and is fine, which is why the check is positional
		// rather than a blanket refusal of empty clauses.
		{"a switch whose default is empty and last", "var foo = { get bar(){ switch (baz) { case 1: return a; default: } } };", "expectedAlways"},
		{
			"a for loop whose only return is inside the body",
			"\n        var foo = {\n            get bar() {\n                for (let i = 0; i<10; i++) {\n                    return i;\n                }\n            }\n        }",
			"expectedAlways",
		},
		// Written here rather than imported, because upstream has no case like it and a mutation
		// sweep proved the gap. Its whole corpus contains exactly one `if` carrying an `else`, and
		// both arms of it return, so a rule crediting an `if` when only ONE arm exits passes every
		// upstream fixture. Two mutants survived on exactly that: dropping the then-arm conjunct,
		// and dropping the else-arm conjunct. These two cases are the distinguishing inputs, one
		// for each conjunct, and each one alone kills only its own mutant.
		{"an if whose else returns but whose then does not", "var foo = { get bar(){ if (baz) { qux(); } else { return false; } } };", "expectedAlways"},
		{"an if whose then returns but whose else does not", "var foo = { get bar(){ if (baz) { return true; } else { qux(); } } };", "expectedAlways"},
		{
			"a while loop whose only return is inside the body",
			"\n        var foo = {\n            get bar() {\n                let i = 0;\n                while (i < 10) {\n                    return i;\n                }\n            }\n        }",
			"expectedAlways",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunTyped(t, GetterReturn, getterReturnFile, testCase.sourceText), testCase.id)
		})
	}
}

// The fail cases upstream runs with `allowImplicit: true`.
//
// These are the sharp half of the option. Turning the option on does not turn the rule off: a
// getter with no return at all still fires, and so does one whose only return sits behind an if.
// A port that read the option as a global mute would pass every fixture in the block above and
// fail every one of these.
func TestGetterReturnFiresWithAllowImplicit(t *testing.T) {
	t.Parallel()

	// ESLint's id: `expected` for a getter with no return of its own, or at a bare `return;`, and
	// `expectedAlways` at the head of one that returns on some paths and falls off the end on another
	cases := []struct {
		name       string
		sourceText string
		id         string
	}{
		{"an empty object getter", "var foo = { get bar() {} };", "expected"},
		{"an if with no else returning implicitly", "var foo = { get bar() {if (baz) {return;}} };", "expectedAlways"},
		{"an empty class getter", "class foo { get bar(){} }", "expected"},
		{"a class getter whose if has no else", "class foo { get bar(){if (baz) {return true;} } }", "expectedAlways"},
		{"an empty defineProperties getter", "Object.defineProperties(foo, { bar: { get: function () {}} });", "expected"},
		{"a defineProperties getter whose if has no else", "Object.defineProperties(foo, { bar: { get: function (){if(bar) {return true;}}}});", "expectedAlways"},
		{"a defineProperties getter returning only from a nested function", "Object.defineProperties(foo, { bar: { get: function () {~function () { return true; }()}} });", "expected"},
		{"an empty defineProperty getter", "Object.defineProperty(foo, \"bar\", { get: function (){}});", "expected"},
		{"an empty Object.create getter", "Object.create(foo, { bar: { get: function (){} } });", "expected"},
		{"an empty Reflect.defineProperty getter", "Reflect.defineProperty(foo, \"bar\", { get: function (){}});", "expected"},
		{"an optional-chained defineProperty", "Object?.defineProperty(foo, 'bar', { get: function (){} });", "expected"},
		{"a parenthesized optional-chained defineProperty", "(Object?.defineProperty)(foo, 'bar', { get: function (){} });", "expected"},
		{"a parenthesized optional-chained Object.create", "(Object?.create)(foo, { bar: { get: function (){} } });", "expected"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunTypedWithOptions(t, GetterReturn, getterReturnFile, testCase.sourceText,
					GetterReturnOptions{AllowImplicit: true}), testCase.id)
		})
	}
}

// The clean cases, and they are where the whole discrimination lives.
//
// Three groups of them do real work. The `var foo = { get: function () {} }` family says a property
// literally named `get` is not a getter unless something treats the object as a descriptor. The
// `foo.defineProperty(...)` family says the right method name on the wrong object is not the
// builtin. And the control-flow group at the end is the reason this rule needed an analysis at all
// rather than a search for the token `return`.
func TestGetterReturnStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"an object getter returning a value", "var foo = { get bar(){return true;} };"},
		{"a class getter returning a value", "class foo { get bar(){return true;} }"},
		{"a class getter whose if has an else and both return", "class foo { get bar(){if(baz){return true;} else {return false;} } }"},
		{"a method literally named get", "class foo { get(){return true;} }"},
		{"a defineProperty getter returning a value", "Object.defineProperty(foo, \"bar\", { get: function () {return true;}});"},
		{"a defineProperty getter with a nested function and its own return", "Object.defineProperty(foo, \"bar\", { get: function () { ~function (){ return true; }();return true;}});"},
		{"a defineProperty setter", "Object.defineProperty(foo, \"bar\", { set: function () {}});"},
		{"a defineProperty arrow setter", "Object.defineProperty(foo, \"bar\", { set: () => {}});"},
		{"a defineProperties getter returning a value", "Object.defineProperties(foo, { bar: { get: function () {return true;}} });"},
		{"a defineProperties getter with a nested function and its own return", "Object.defineProperties(foo, { bar: { get: function () { ~function (){ return true; }(); return true;}} });"},
		{"a defineProperties setter", "Object.defineProperties(foo, { bar: { set: function () {}} });"},
		{"a Reflect.defineProperty getter returning a value", "Reflect.defineProperty(foo, \"bar\", { get: function () {return true;}});"},
		{"a Reflect.defineProperty getter with a nested function and its own return", "Reflect.defineProperty(foo, \"bar\", { get: function () { ~function (){ return true; }();return true;}});"},
		{"an Object.create shorthand getter returning a value", "Object.create(foo, { bar: { get() {return true;} } });"},
		{"an Object.create getter returning a value", "Object.create(foo, { bar: { get: function () {return true;} } });"},
		{"an Object.create arrow getter returning a value", "Object.create(foo, { bar: { get: () => {return true;} } });"},
		{"a variable named get holding an empty function", "var get = function(){};"},
		{"a variable named get holding a returning function", "var get = function(){ return true; };"},
		{"an ordinary empty method", "var foo = { bar(){} };"},
		{"an ordinary returning method", "var foo = { bar(){ return true; } };"},
		{"an ordinary empty function property", "var foo = { bar: function(){} };"},
		{"an ordinary function property returning bare", "var foo = { bar: function(){return;} };"},
		{"an ordinary function property returning a value", "var foo = { bar: function(){return true;} };"},
		{"a plain object with a get property", "var foo = { get: function () {} }"},
		{"a plain object with a get arrow property", "var foo = { get: () => {}};"},
		{"a class field named get", "class C { get; foo() {} }"},
		{"defineProperty on the wrong object", "foo.defineProperty(null, { get() {} });"},
		{"defineProperties on the wrong object", "foo.defineProperties(null, { bar: { get() {} } });"},
		{"create on the wrong object", "foo.create(null, { bar: { get() {} } });"},
		{"a getter that throws", "var foo = { get willThrowSoValid() { throw MyException() } };"},
		{
			"an arrow getter with an expression body",
			"const originalClearTimeout = targetWindow.clearTimeout;\n        Object.defineProperty(targetWindow, 'vscodeOriginalClearTimeout', { get: () => originalClearTimeout });\n        ",
		},
		{
			"a getter whose return follows a logical-or subscript",
			"\n        var foo = {\n                get bar() {\n                        let name = ([] || [])[1];\n                        return name;\n                },\n        };\n        ",
		},
		{"a try that returns with an empty finally", "var foo = { get bar() { try { return a(); } finally {  } } };"},
		// Also written here rather than imported. Upstream's only try-with-finally case has an EMPTY
		// finally, so the branch crediting a finally that exits on its own never fires on the
		// corpus: the try block carries that case by itself. A mutant deleting the finally-wins
		// branch entirely survived every upstream fixture. This is the distinguishing input, a
		// finally that returns while the try does not, and it is correct to stay silent because a
		// finally runs on every path out of the try and the catch both.
		{"a try whose finally returns while the try does not", "var foo = { get bar() { try { a(); } finally { return b(); } } };"},
		{"a try whose finally returns while the catch does not", "var foo = { get bar() { try { a(); } catch { c(); } finally { return b(); } } };"},
		// The pass half of the quoted-key case, plus its negative. A quoted `set` is still not a
		// getter, so the key reader has to return the actual text rather than merely succeed.
		{"a descriptor with a quoted get key that returns", "Object.defineProperty(foo, 'bar', { 'get': function (){ return 1; } });"},
		{"a descriptor with a quoted set key", "Object.defineProperty(foo, 'bar', { 'set': function (){} });"},
		// The pass halves of the labeled and `with` mutants above. A label wraps a statement without
		// changing whether it exits, and neither does `with`, so a return inside either one counts.
		// It takes both halves to kill those mutants: the fail case alone cannot tell "never
		// unwraps" from "unwraps correctly", because both report on a body that does not exit.
		{"a labeled block that returns", "var foo = { get bar() { outer: { return 1; } } };"},
		{"a labeled if whose arms both return", "var foo = { get bar() { lbl: if (a) { return 1; } else { return 2; } } };"},
		{"a with block that returns", "var foo = { get bar() { with (o) { return 1; } } };"},
		// The arrow with an expression body, as a DESCRIPTOR getter rather than the object-literal
		// spelling upstream ships. Upstream's one case of this shape is clean under every mutant
		// tried, because two gates subsume each other there: `bodyOf` declines a non-block arrow
		// body, and `check` declines a non-block body again. Written at the descriptor position and
		// paired with the empty-block arrow in the Fires block, the pair pins that an expression
		// body is credited while an empty block is not.
		{"a descriptor arrow getter with an expression body", "Object.defineProperty(foo, 'bar', { get: () => 1 });"},
		// The counterpart to the trailing-empty-clause fail case above: an empty clause in the
		// MIDDLE is a deliberate fallthrough, `case A:` reaching `case B`'s return, and it must not
		// be read as a path out.
		{"a switch with a fallthrough empty clause in the middle", "var foo = { get bar(){ switch (baz) { case 1: case 2: return a; default: return d; } } };"},
		{
			"a switch with a default where every clause returns",
			"\n        var foo = {\n            get bar() {\n                switch (baz) {\n                    case VS_LIGHT_THEME: return a;\n                    case VS_HC_THEME: return b;\n                    case VS_HC_LIGHT_THEME: return c;\n                    default: return d;\n                }\n            }\n        };\n        ",
		},
		{
			"a class getter with a loop and a trailing return",
			"\n        export default class ProgressBar extends Component {\n            get steps() {\n                const steps = [];\n\n                for (let i = 0; i < this.maxStepsNumber; i++) {\n                    steps.push({\n                        stepnum: i + 1,\n                    });\n                }\n\n                return steps;\n            }\n        }",
		},
		{
			"a for loop returning inside plus a trailing return",
			"\n        var foo = {\n            get bar() {\n                for (let i = 0; i<10; i++) {\n                    if (i === 5) {\n                        return i;\n                    }\n                }\n                return 0;\n            }\n        }",
		},
		{
			"a while loop returning inside plus a trailing return",
			"\n        var foo = {\n            get bar() {\n                let i = 0;\n                while (i < 10) {\n                    if (i === 5) {\n                        return i;\n                    }\n                    i++;\n                }\n                return 0;\n            }\n        }",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, GetterReturn, getterReturnFile, testCase.sourceText))
		})
	}
}

// The pass cases upstream runs with `allowImplicit: true`.
//
// The option's actual job: a bare `return;` now satisfies the rule, including one behind an if that
// is followed by a real return.
func TestGetterReturnStaysSilentWithAllowImplicit(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"a bare return in an object getter", "var foo = { get bar() {return;} };"},
		{"a value return in an object getter", "var foo = { get bar(){return true;} };"},
		{"a bare return behind an if followed by a real return", "var foo = { get bar(){if(bar) {return;} return true;} };"},
		{"a value return in a class getter", "class foo { get bar(){return true;} }"},
		{"a bare return in a class getter", "class foo { get bar(){return;} }"},
		{"a value return in a defineProperty getter", "Object.defineProperty(foo, \"bar\", { get: function () {return true;}});"},
		{"a bare return in a defineProperty getter", "Object.defineProperty(foo, \"bar\", { get: function (){return;}});"},
		{"a value return in a defineProperties getter", "Object.defineProperties(foo, { bar: { get: function () {return true;}} });"},
		{"a bare return in a defineProperties getter", "Object.defineProperties(foo, { bar: { get: function () {return;}} });"},
		{"a value return in a Reflect.defineProperty getter", "Reflect.defineProperty(foo, \"bar\", { get: function () {return true;}});"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunTypedWithOptions(t, GetterReturn, getterReturnFile, testCase.sourceText,
					GetterReturnOptions{AllowImplicit: true}))
		})
	}
}

// A TypeScript file is checked, as ESLint checks it where the house config turns the rule on in .ts
// files (#mnmx9s4). oxc's second Tester block pinned the opposite, a getter annotated
// `boolean | undefined` with a path that returns nothing as clean, by declining every TypeScript file.
// ESLint reports it, since the rule reads no types, and ESLint is the authority here (#jjfa7qb).
func TestGetterReturnChecksTypeScript(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, GetterReturn, getterReturnFile,
		"var foo = {\n            get bar(): boolean | undefined {\n                if (Math.random() > 0.5) {\n                    return true;\n                }\n            }\n        };"),
		"expectedAlways")
}

// A descriptor call whose `Object` is a local is not the platform method, so its `get` is not a
// getter, as no-setter-return reads it.
func TestGetterReturnAsksWhetherObjectIsTheGlobal(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, rule_testing.RunTyped(t, GetterReturn, getterReturnFile,
		"declare let Object: { defineProperty(target: object, key: string, value: object): void };\nObject.defineProperty({}, 'bar', { get: function () {} });"))
	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, GetterReturn, getterReturnFile,
		"Object.defineProperty({}, 'bar', { get: function () {} });"), "expected")
}

// The span, which the message-id fixtures above cannot see.
//
// `ExpectFindings` asserts which message fired and how many times, and nothing about where. A rule
// pointing at the whole getter body, or at the enclosing object, or one token off, passes every
// assertion above. Upstream anchors on the function's head via `getFunctionHeadLoc`, and this pins
// that we do too, by slicing the source with the finding's own range.
//
// Six cases rather than one, because the head is exactly what varies. ESLint's `getFunctionHeadLoc`
// runs from the property to the opening paren of the parameters: an object getter's head is `get bar`,
// a descriptor function's is `get: function ` (the property is where it starts), a shorthand's is
// `get`, and an arrow's is `get: `. A range built from the wrong node would still look right on one
// of them.
//
// The port this rule followed ran the head to the body's start, `get bar()` and `function ()`, so
// every one of ESLint's 35 corpus rows read as a different span until the head followed ESLint's
// (#jjfa7qb). The newline cases keep the whitespace before the paren, as ESLint's range does.
func TestGetterReturnPointsAtTheGetterHead(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		want       string
	}{
		{
			"an object getter",
			"var foo = { get bar() {} };",
			"get bar",
		},
		{
			"a descriptor getter",
			"Object.defineProperty(foo, 'bar', { get: function (){}});",
			"get: function ",
		},
		{
			"a descriptor shorthand getter",
			"Object.defineProperty(foo, 'bar', { get(){} });",
			"get",
		},
		{
			"a descriptor arrow getter",
			"Object.defineProperty(foo, 'bar', { get: () => {}});",
			"get: ",
		},
		// Two of upstream's fail cases put a newline inside the head, and they are the reason the
		// range is built from the declaration's first token to the body's start rather than from
		// anything simpler. A range guessed as "the name plus a fixed number of characters", or one
		// stopping at the first line break, gets both of these wrong while getting every
		// single-line case above right.
		{
			"an object getter with a newline in its head",
			"var foo = { get\n bar () {} };",
			"get\n bar ",
		},
		{
			"a static class getter with a newline in its head",
			"var foo = class {\n  static get\nbar(){} }",
			"static get\nbar",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, GetterReturn, getterReturnFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
			}
			reported := testCase.sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if reported != testCase.want {
				t.Errorf("the finding points at %q, wanted %q", reported, testCase.want)
			}
		})
	}
}
