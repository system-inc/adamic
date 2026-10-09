package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageDivRegex = rule.Message{
	Id: "unexpected",
	Description: "This regular expression begins with an equals sign, so its first two characters " +
		"are the divide-assign operator spelled backwards, and a reader scanning the line has to " +
		"work out which one it is. Writing the equals sign as a one-member character class " +
		"matches exactly the same input and can only be read one way.",
}

// NoDivRegex flags a regular expression literal whose pattern begins with an equals sign.
//
//	valid:   /foo/ig
//	valid:   /\=foo/         // the equals sign is escaped, so the first character is a backslash
//	valid:   /[=]foo/        // already the repair
//	valid:   '=foo'          // a string, not a regular expression
//	invalid: /=foo/          // fixed to /[=]foo/
//	invalid: /=/
//	invalid: /=foo/g
//
// # The test is positional, and that is the whole rule
//
// Upstream takes the literal's first token and asks whether `token.value[1] === "="`. Index 1 is
// the character right after the opening slash, so this is a question about the second byte of the
// token text and nothing else. It is not a question about what the pattern matches: `/(=)/` and
// `/[=]foo/` both match an equals sign and both are clean, while `/==/` reports; all three
// measured against the installed ESLint build.
//
// So no regular expression parse is involved and none should be. A port that walked the pattern
// and asked whether the first element can match `=` would report `/[=]foo/`, which is the repair
// this rule proposes, and upstream's own clean case.
//
// The escape case falls out of the same positional reading rather than needing its own branch:
// in `/\=foo/` the character at index 1 is a backslash, so the test fails. That is upstream's
// second clean case.
//
// # The repair rewrites one byte and is a fix rather than a suggestion
//
// Upstream replaces the range `[start+1, start+2)` with `[=]`, which turns `/=foo/` into
// `/[=]foo/`. A one-member character class matches exactly the character it names, so the pattern
// accepts the same input before and after and there is no second valid answer to pick between.
// That is what makes it a fix the engine may apply unattended rather than a suggestion needing an
// author, and it is the same reasoning `no-regex-spaces` records for its own repair.
//
// The span of the FINDING is the whole literal including its flags, which upstream's corpus states
// as columns 29 through 35 for `/=foo/`. The span of the FIX is the single equals sign. Those
// differ deliberately, and asserting only the first would not notice a repair that rewrote the
// wrong byte, which is why the tests assert the rewritten source rather than the fix text.
//
// # Only a literal, never a constructor
//
// Upstream reads the first TOKEN of the node and requires its type to be `RegularExpression`, so
// `new RegExp('=foo')` never qualifies: its first token is `new`. Measured, that input is clean.
// Here the listener key is the regular expression kind, which asks the same question at the parser
// instead. `no-regex-spaces` handles both spellings and this one does not, and the difference is
// upstream's rather than an omission.
var NoDivRegex = rule.Rule{
	Name: "no-div-regex",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindRegularExpressionLiteral: func(node *ast.Node) {
				literalRange := rule.TokenRange(ctx.SourceFile, node)
				// Sliced from the range rather than read off the node, though for this kind the
				// two are the same string. A mutation swapping in `node.Text()` survived the whole
				// fixture set, and probing eight shapes including ones with leading block and line
				// comments found them identical every time, so that survivor is equivalent rather
				// than a blind spot. The range form is kept because the fix offset below has to
				// come from the range in any case, and reading both from one place means an
				// offset and the text it indexes cannot drift apart.
				text := ctx.SourceFile.Text()[literalRange.Pos():literalRange.End()]

				// Index 1 is the character after the opening slash.
				//
				// The length test is not defensive padding; it prevents a panic on input the
				// parser really does produce. The grammar has no literal shorter than two
				// characters, but error recovery does: probing malformed source found six shapes
				// yielding a one-byte literal holding just the opening slash, among them
				// `var a = /;` and a file truncated mid-pattern. Indexing byte 1 of one of those
				// takes the run down, and no findings fixture can see a panic, which is why a
				// mutation deleting this survived until a test was written that would crash
				// without it.
				if len(text) < 2 || text[1] != '=' {
					return
				}

				// The equals sign alone, which is what the repair replaces. The finding below
				// still points at the whole literal, matching upstream.
				equalsRange := core.NewTextRange(literalRange.Pos()+1, literalRange.Pos()+2)

				ctx.ReportNodeWithFixes(node, messageDivRegex,
					rule.ReplaceRange(equalsRange, "[=]"))
			},
		}
	},
}
