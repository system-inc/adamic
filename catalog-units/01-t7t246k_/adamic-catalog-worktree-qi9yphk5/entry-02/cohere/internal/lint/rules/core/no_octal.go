package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoOctal = rule.Message{
	Id: "noOctal",
	Description: "This number is written in the legacy octal form, a bare leading zero, so `0755` " +
		"is 493 rather than seven hundred fifty five. The zero is easy to miss and the value the " +
		"program runs with is not the one the digits appear to spell. The form is also a syntax " +
		"error in strict mode and in every module, so the same literal is silently reinterpreted " +
		"in one file and rejected outright in another. Write `0o755` when octal is meant, which " +
		"says so, or drop the leading zero when it is not.",
}

// NoOctal flags a number written in the legacy octal form: a leading `0` followed by a digit.
//
//	valid:   var a = 0;
//	valid:   var a = 0.5;
//	valid:   var a = 0x1F;
//	valid:   var a = 0o755;
//	valid:   var a = 0b11;
//	invalid: var a = 0755;
//	invalid: var a = 07;
//	invalid: var a = 08;
//
// # This rule has no oxc implementation, so ESLint is the port target
//
// The whole `oxc_linter` crate has no `no_octal.rs` and no `NoOctal`, confirmed by grepping for both
// with `no-with` and `NoWith` run as controls through the same commands. So there is no inline Rust
// corpus and no snapshot diagnostic count to import, and ESLint does not ship its `tests/` directory
// either. Every fixture beside this file is one the port invented, which the test file says plainly
// because the next reader should know which kind of artifact they are holding.
//
// ESLint's entire implementation is one predicate: a literal whose value is a number and whose *raw*
// text matches `/^0\d/u`. No options, `schema: []`, one message id, no fix and no suggestion.
//
// # `08` and `09` are not octal and upstream reports them anyway
//
// This is the one place a careful porter diverges by being careful. `8` and `9` are not octal
// digits, so `08` and `0195` are NonOctalDecimalIntegerLiterals that evaluate in base ten, and
// reasoning from the rule's *name* says to leave them alone. Upstream reports them, measured by
// running its rule rather than by reading its regular expression, because they are the same banned
// legacy grammar production: both are a syntax error in strict mode, both are the leading zero that
// changes meaning between one file and the next. So the predicate here is the spelling, `0` then a
// decimal digit, and not whether the remaining digits happen to be octal-legal.
//
// `no-loss-of-precision`, in this same package, has to make the opposite call and its
// `rawNumericLiteralIsBaseTen` does check that every digit is octal, because there the question is
// what value the literal actually denotes. Two rules reading the same characters for different
// reasons, which is why neither shares the other's helper.
//
// # The raw span, because our parser hands back a cooked value
//
// `node.Text()` is normalized: `0755` arrives as `"493"`, `0x1F` as `"31"`, `0e5` as `"0"`. The
// leading zero this rule exists to find has already been erased by the time the value is available,
// so reading the parsed text would make the rule structurally silent on every input it is for, and
// every clean fixture would still pass. Measured with a probe rather than assumed, and pinned by
// `TestNoOctalReadsRawTextRatherThanTheCookedValue`, whose pairs cook to identical strings on both
// sides of the judgment: `0755` and `0o755` are both `"493"`, and one of them is upstream's
// recommended fix.
//
// So the raw text comes out of the source file at the literal's own token range, which is also what
// makes the reported span the literal alone rather than the statement around it.
//
// # Where our parser and ESLint's disagree, and what was chosen
//
// `01.5`, `0777.5`, `0755n` and `0_7_5_5` are syntax errors in real JavaScript. ESLint's parser
// rejects the file, so its rule never runs and upstream has no opinion about them. TypeScript's
// parser recovers from all four, so this port has to decide, and it reports: the subject is present
// in the source however the parser recovered from it. `01.5` recovers as two literals, `01` and
// `.5`, so it reports once, on the `01`. `0755n` loses its suffix and arrives as a plain
// `KindNumericLiteral`, so it is not exempt the way a real BigInt would be, which is worth stating
// because the exemption `no-loss-of-precision` gets for free from `KindBigIntLiteral` does not
// transfer here.
//
// # The compiler already rejects every one of these
//
// Worth knowing before enabling it: `tsc` raises TS1121 for a legacy octal and TS1489 for a decimal
// with a leading zero. These are grammar errors, not check errors, so they fire in `.ts` and `.js`
// alike, with `strict` on or off, with `checkJs` off, and they survive both `@ts-nocheck` and
// `@ts-ignore`. There is no configuration in which the compiler accepts this syntax. The rule is
// therefore redundant with the type phase rather than inert, and it is kept for the reason a linter
// keeps any such rule: it reports in the lint phase, which runs without the compiler and gives a
// message about what to write instead rather than a grammar error.
var NoOctal = rule.Rule{
	Name: "no-octal",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			// `KindNumericLiteral` alone is upstream's `typeof node.value === "number"`. A string
			// holding the same digits is a different kind and a comment is not a node, so both are
			// declined by the listener rather than by a check inside it.
			ast.KindNumericLiteral: func(node *ast.Node) {
				literalRange := rule.TokenRange(ctx.SourceFile, node)
				raw := ctx.SourceFile.Text()[literalRange.Pos():literalRange.End()]
				if rawNumericLiteralIsLegacyOctal(raw) {
					ctx.ReportNode(node, messageNoOctal)
				}
			},
		}
	},
}

// rawNumericLiteralIsLegacyOctal answers upstream's `/^0\d/u` against a literal's raw source text.
//
// Two characters decide it and the second one is the whole rule. A leading `0` alone is just zero,
// and a `0` followed by `.`, `e`, `x`, `o`, `b` or nothing at all is a fraction, an exponent, or one
// of the three modern prefixes. Only a `0` followed immediately by a decimal digit is the legacy
// production.
//
// Written as an explicit digit test rather than as a regular expression because the predicate is two
// characters wide and a compiled pattern per call would cost more than the rule does. The prefix
// letters need no arm of their own: they are not decimal digits, so `0x1F` and `0o755` fall out of
// the same comparison that lets `0.5` through, in either letter case.
func rawNumericLiteralIsLegacyOctal(raw string) bool {
	if len(raw) < 2 || raw[0] != '0' {
		return false
	}
	return raw[1] >= '0' && raw[1] <= '9'
}
