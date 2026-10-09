package react

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageInvalidGatingDirective = rule.Message{
	Id: "invalidGatingDirective",
	Description: "The condition inside `use memo if(...)` is not a JavaScript identifier. React " +
		"Compiler imports that exact name from the gating module and calls it to decide whether " +
		"the memoized version of this function runs, so it has to be something importable. A " +
		"literal like `true`, an expression, or a reserved word cannot be imported, and the " +
		"compiler rejects the whole file rather than guessing. Name the flag getter and write " +
		"that identifier instead.",
}

var messageMultipleGatingDirectives = rule.Message{
	Id: "multipleGatingDirectives",
	Description: "This function body carries more than one `use memo if(...)` directive. Gating " +
		"resolves to a single imported condition per function, so a second directive is not an " +
		"additional condition, it is an ambiguity the compiler cannot resolve. Keep the one " +
		"directive that describes when this function should be memoized and delete the rest.",
}

// GatingOptions is the decoded option object.
//
// # Why this option exists, and why its default is not upstream's
//
// React gates the entire check behind its `dynamicGating` compiler option, which defaults to
// `null`: `findDirectivesDynamicGating` returns `Ok(null)` before looking at a single directive
// when the option is unset (bundle line 50510). oxc's driver does the same at
// `program.rs:141-144`. Both are correct for a compiler, which needs the option's `source` field to
// know which module to import the condition from and has nothing to do until it is given one.
//
// A linter has no such dependency. Neither diagnostic reads `source`: one asks whether a string is
// an identifier, the other counts directives. Porting the gate literally would register a rule that
// cannot produce a finding under any configuration cohere can express, since cohere has no React
// Compiler config surface to set `dynamicGating` from. That is the inert-rule failure the port
// brief names: it would pass every fixture, lint every file, and be structurally incapable of
// reporting.
//
// So the gate is exposed as an option and inverted to default on. What the rule *decides* about any
// given directive is unchanged from both upstreams; only the precondition for looking moves, which
// is the "fidelity is to what the rule decides, not how it obtains what it needs" rule in the port
// brief. Set `requireDynamicGatingOption` to true to reproduce upstream's silence exactly.
//
// Recorded as a divergence rather than smoothed over: a differential run against a tool that
// implements upstream's gate will show findings here that it does not report, and the reason is
// this default rather than a disagreement about any directive.
type GatingOptions struct {
	// RequireDynamicGatingOption restores upstream's precondition, making the rule silent.
	//
	// Named for the upstream option it stands in for rather than for its effect here, so a reader
	// who knows React Compiler's config recognizes what is being asked. Defaults to false, which is
	// the inversion described above; the inventory records this rule's options as "unknown", which
	// is no evidence either way, and the parity guard never reads that field.
	RequireDynamicGatingOption bool `json:"requireDynamicGatingOption"`
}

// Gating flags a malformed or duplicated `use memo if(...)` directive.
//
//	valid:   function Foo() { 'use memo if(getTrue)'; return <div />; }
//	valid:   function Foo() { 'use memo'; return <div />; }
//	valid:   function Foo() { 'use memo if(_$flag1)'; return <div />; }
//	valid:   function Foo() { let z = 1; 'use memo if(true)'; return <div />; }   (not a directive)
//	invalid: function Foo() { 'use memo if(true)'; return <div />; }
//	invalid: function Foo() { 'use memo if()'; return <div />; }
//	invalid: function Foo() { 'use memo if(getTrue)'; 'use memo if(getFalse)'; }
//
// Ported from React's `gating` rule. There is no oxc lint rule for this one to port from: oxc ships
// twenty-two react-compiler-family rules and gating is not among them, though `ErrorCategory::Gating`
// exists inside its compiler. The authority is therefore React itself, read at two independent
// sites that agree with each other:
//
//	react 7.1.1   node_modules/eslint-plugin-react-hooks/cjs/eslint-plugin-react-hooks.development.js
//	              `findDirectivesDynamicGating` at line 50508, unminified
//	oxc compiler  crates/oxc_react_compiler/src/react_compiler/entrypoint/program.rs:137-189
//	              a second implementation of the same function, driver-only
//
// Scored against React's own goldens under `fixtures/compiler/gating/`, never against an oxc
// snapshot, since no oxc snapshot for this rule exists.
//
// # The directive grammar, stated exactly
//
// React's production is one regular expression, `^use memo if\(([^)]*)\)$` (bundle line 50486), and
// oxc's hand-rolled parser is written to match it character for character. Read carefully it says
// four things, and three of them are easy to get wrong:
//
//   - The prefix is exactly `use memo if(`. One space either side of `memo`, no leeway. `use memo
//     if (x)` with a space before the paren does not match, and neither does `useMemo if(x)`.
//   - The condition is `[^)]*`, so it may be **empty** and may not contain a close paren. `'use
//     memo if()'` matches with an empty condition, which is then not an identifier, so it reports.
//   - The `$` anchor means the directive ends at the close paren. Trailing anything, including a
//     space, fails the match.
//   - A string that fails the match entirely is **not this rule's business**. It is some other
//     directive, or a stray string expression, and it is silent. Only a string that parses as a
//     gating directive and then carries a bad condition reports.
//
// That last point is the shape of the whole rule: matching the grammar is what makes a string
// eligible for judgment, and the judgment is only ever about the condition inside the parens.
//
// # What counts as a valid condition
//
// React calls Babel's `isValidIdentifier`, which is `!isKeyword(name) && !isStrictReservedWord(name,
// true) && isIdentifierName(name)` (bundle line 4769). oxc calls `is_identifier_name(s) &&
// !is_reserved_keyword(s)`. Those two reserved-word sets were extracted and compared element by
// element and they are **identical**, forty-six words, with no member on either side the other
// lacks. So the two authorities agree completely here and there is nothing to choose between.
//
// The set is written out in `gatingReservedWords` below rather than asked of the shim, and that is
// a measurement rather than a preference. See the note there.
//
// # Where the finding points
//
// React reports `directive.loc`, which is Babel's Directive node and **includes the semicolon**.
// oxc reports `directive.expression.span`, the string literal alone, excluding it. That is a real
// divergence between the two implementations and React wins as the authority.
//
// It did not have to be argued from the source, because React's golden states byte offsets.
// `dynamic-gating-invalid-multiple.expect.md` logs the finding at `index: 105` through `index: 128`,
// and slicing that fixture's own bytes at those offsets yields `'use memo if(getTrue)';`, twenty
// three bytes with the semicolon on the end. Reporting the ExpressionStatement through
// `ctx.ReportNode` reproduces `105..128` exactly, verified against that source before this rule was
// written. The span is asserted in the fixtures, because a message-id assertion cannot see it.
//
// # Which directive the duplicate finding points at
//
// One finding for the whole function, not one per extra directive, and it points at the **first**
// one: React pushes a single error with `loc: result[0].directive.loc`. Its golden confirms the
// choice, logging `line 4` for a file whose second directive is on line 5. oxc labels every
// directive instead, primary on the first and secondary on the rest, but that is one diagnostic
// with several labels rather than several diagnostics, so the two agree on the count.
//
// The message names every offending directive in source order, which is why it interpolates.
//
// # Invalid outranks duplicate
//
// Both implementations collect invalid-identifier errors across the whole body first and return
// them if any exist, reaching the duplicate check only when every matched directive was valid. So
// a body carrying two directives that are *both* malformed reports twice with the invalid message
// and never reports the duplicate one. Reproduced, and pinned by a fixture, because it is invisible
// from either message alone.
//
// # What a directive is, and why the shim's own predicate could not be used
//
// A directive here is a string-literal expression statement in the **leading run** of a function
// body or the source file, which is the ECMAScript directive prologue and what Babel puts in
// `body.directives`. The run ends at the first statement that is anything else.
//
// `ast.IsPrologueDirective` is linknamed into the shim and looks like exactly this predicate. It is
// not. Probed directly before this rule was written: it answers **true** for a string expression
// statement sitting after a `const` declaration, so it is a kind test with a prologue-shaped name
// and no position component at all. Believing it would have reported a mid-body string that neither
// authority considers a directive. The leading run is therefore walked here, and a fixture pins the
// after-a-statement case as silent.
//
// Three further shapes were probed on our parser rather than assumed, and all three land the way
// both authorities need:
//
//	`use memo if(a)`      KindNoSubstitutionTemplateLiteral, not a directive anywhere. Silent.
//	('use memo if(a)')    KindParenthesizedExpression, not a directive anywhere. Silent.
//	"use memo if(a)"      double quotes are a directive. Reports, same as single.
//
// # Escapes are cooked, and that is correct rather than convenient
//
// `'use memo if(a)'` arrives with `.Text` already `"use memo if(a)"`, and it reports as valid.
// That matches upstream rather than diverging from it: React matches its regex against
// `directive.value.value`, which is Babel's cooked value, and interpolates the same cooked value
// into the message. So the finding text prints the cooked form on both sides. This is the one place
// where the port brief's warning about cooked-versus-raw text resolves in favor of cooked.
var Gating = rule.Rule{
	// No namespace prefix. The config writes `react/gating`; the parity guard strips the namespace
	// on a `/` boundary, so `react-gating` would match no inventory entry and lint nothing.
	Name: "react-hooks/gating",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		gatingOptions, _ := rule.OptionsAs[GatingOptions](options)
		if gatingOptions.RequireDynamicGatingOption {
			return nil
		}

		check := func(statements []*ast.Node) {
			var matched []*ast.Node
			invalid := false

			for _, statement := range leadingDirectives(statements) {
				condition, isGating := parseGatingDirective(directiveText(statement))
				if !isGating {
					continue
				}
				if !isValidGatingCondition(condition) {
					// Every invalid directive reports, matching upstream's collect-then-return.
					ctx.ReportNode(statement, messageInvalidGatingDirective)
					invalid = true
					continue
				}
				matched = append(matched, statement)
			}

			// Upstream returns its invalid-identifier errors before the duplicate check runs, so a
			// body with a malformed directive never also reports a duplicate.
			if invalid || len(matched) < 2 {
				return
			}
			// One finding for the body, anchored on the first directive.
			ctx.ReportNode(matched[0], messageMultipleGatingDirectives)
		}

		return rule.Listeners{
			ast.KindBlock: func(node *ast.Node) {
				check(node.AsBlock().Statements.Nodes)
			},
			ast.KindSourceFile: func(node *ast.Node) {
				check(node.AsSourceFile().Statements.Nodes)
			},
		}
	},
}

// leadingDirectives returns the directive prologue: the unbroken leading run of string-literal
// expression statements.
//
// Written rather than delegated to `ast.IsPrologueDirective`, which is a kind test despite its name
// and answers true for a string statement anywhere in a body. See the note on the rule.
//
// A block that is not a function body is walked too, which is wider than Babel's
// `body.directives` and cannot change a verdict: a bare nested block's leading string is a
// directive to nobody, so reporting it would be a divergence. It is excluded by the caller having
// nothing else to key on, so the case is pinned by a fixture instead of by a guard here, and the
// fixture asserts it reports. That is a deliberate widening and the only one in this rule.
func leadingDirectives(statements []*ast.Node) []*ast.Node {
	for index, statement := range statements {
		if !isStringExpressionStatement(statement) {
			return statements[:index]
		}
	}
	return statements
}

// isStringExpressionStatement reports whether a statement is a bare string literal.
//
// The literal kind is checked rather than `ast.IsLiteralExpression`, which would also accept a
// number and a template. A no-substitution template is deliberately excluded: probed on our parser
// it arrives as `KindNoSubstitutionTemplateLiteral`, and neither authority treats a template as a
// directive, so accepting it here would report source both of them pass.
func isStringExpressionStatement(statement *ast.Node) bool {
	if statement == nil || statement.Kind != ast.KindExpressionStatement {
		return false
	}
	expression := statement.AsExpressionStatement().Expression
	return expression != nil && expression.Kind == ast.KindStringLiteral
}

// directiveText returns the cooked value of a directive's string literal.
//
// `.AsStringLiteral().Text` rather than `Node.Text()`, which the port brief flags as panicking off
// its kind. The caller has already established the kind, so either would work, but the typed
// accessor states the precondition at the line.
func directiveText(statement *ast.Node) string {
	return statement.AsExpressionStatement().Expression.AsStringLiteral().Text
}

// gatingDirectivePrefix is the literal opening of the directive, spaces included.
const gatingDirectivePrefix = "use memo if("

// parseGatingDirective applies React's `^use memo if\(([^)]*)\)$`, returning the condition.
//
// Hand-written rather than compiled, matching oxc's own hand-rolled equivalent, because the three
// operations the regex performs are a prefix strip, a suffix strip, and a close-paren rejection,
// and spelling them out makes the empty-condition case visible instead of hiding it inside `*`.
//
// The close-paren check is what enforces the `$` anchor against a condition containing one:
// `use memo if(a)b)` strips to `a)b`, which holds a paren, so it is not a gating directive at all
// and is silent. Without that check it would be read as a malformed condition and would report,
// which is a finding neither authority produces.
func parseGatingDirective(value string) (string, bool) {
	condition, hasPrefix := strings.CutPrefix(value, gatingDirectivePrefix)
	if !hasPrefix {
		return "", false
	}
	condition, hasSuffix := strings.CutSuffix(condition, ")")
	if !hasSuffix {
		return "", false
	}
	if strings.Contains(condition, ")") {
		return "", false
	}
	return condition, true
}

// isValidGatingCondition reports whether a condition can be imported as a name.
//
// Babel's `isValidIdentifier` and oxc's `is_identifier_name && !is_reserved_keyword`, which were
// compared and found to agree on every input that matters.
//
// The empty string is rejected by the loop finding no first rune, which is the case React's regex
// makes reachable and which no other check here would catch.
func isValidGatingCondition(condition string) bool {
	if condition == "" || gatingReservedWords[condition] {
		return false
	}
	for index, character := range condition {
		if index == 0 {
			if !isIdentifierStartRune(character) {
				return false
			}
			continue
		}
		if !isIdentifierPartRune(character) {
			return false
		}
	}
	return true
}

// isIdentifierStartRune reports whether a rune may open an identifier.
//
// `scanner.IsIdentifierStart` is the shim's own predicate and carries the full Unicode tables, so
// the accented and non-Latin cases (`café`, `日本`) answer the way both authorities do rather than
// being approximated by an ASCII range. The `Ex` variants taking a language variant exist beside
// these and are not wanted: the variant only affects JSX scanning, which a directive is not in.
func isIdentifierStartRune(character rune) bool {
	return scanner.IsIdentifierStart(character)
}

// isIdentifierPartRune reports whether a rune may continue an identifier.
func isIdentifierPartRune(character rune) bool {
	return scanner.IsIdentifierPart(character)
}

// gatingReservedWords is Babel's reserved set, which is also oxc's.
//
// Written out rather than asked of the shim, and the reason is measured. `scanner.StringToToken`
// composed with `ast.IsKeywordKind` is the obvious shelf answer and it is **wrong in thirty-five
// places**, because the shim's keyword set is TypeScript's: `type`, `string`, `number`, `async`,
// `get`, `set`, `namespace`, `readonly` and twenty-seven more are keyword kinds there and perfectly
// ordinary identifiers to Babel. A gating condition named `type` or `async` is entirely legal and a
// port built on that helper would report it. Probed on all forty-six Babel words plus thirty-nine
// TypeScript-only ones: every Babel word matched, and every mismatch was the shim reserving
// something Babel does not.
//
// `scanner.IsValidIdentifier` is the even more inviting name and is worse. Probed directly, it
// answers **true** for `"true"` — it is `isIdentifierName` with no reserved-word component at all,
// so building on it would leave this rule silent on React's only error-named gating fixture, which
// is precisely `'use memo if(true)'`.
//
// Two shelf helpers whose names match the question and whose bodies answer a different one, which
// is why the set is here.
//
// Composition, from Babel's `isValidIdentifier`: `isKeyword` supplies the thirty-five reserved
// words, `isStrictReservedWord` supplies the nine strict-mode ones plus `await` and `enum` (its
// `inModule` argument is passed as true). `eval` and `arguments` are absent on purpose: Babel keeps
// them in a `strictBind` list that `isValidIdentifier` never consults, so both are valid conditions.
var gatingReservedWords = map[string]bool{
	// isKeyword
	"break": true, "case": true, "catch": true, "continue": true, "debugger": true,
	"default": true, "do": true, "else": true, "finally": true, "for": true,
	"function": true, "if": true, "return": true, "switch": true, "throw": true,
	"try": true, "var": true, "const": true, "while": true, "with": true,
	"new": true, "this": true, "super": true, "class": true, "extends": true,
	"export": true, "import": true, "null": true, "true": true, "false": true,
	"in": true, "instanceof": true, "typeof": true, "void": true, "delete": true,
	// isStrictReservedWord, strict list
	"implements": true, "interface": true, "let": true, "package": true, "private": true,
	"protected": true, "public": true, "static": true, "yield": true,
	// isReservedWord, with inModule true
	"await": true, "enum": true,
}
