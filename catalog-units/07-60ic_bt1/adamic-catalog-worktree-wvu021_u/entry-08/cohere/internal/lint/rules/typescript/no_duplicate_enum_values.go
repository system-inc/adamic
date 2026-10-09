package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoDuplicateEnumValues = rule.Message{
	Id: "noDuplicateEnumValues",
	Description: "Two members of this enum are initialized to the same value. TypeScript permits " +
		"it, but a reader expects members of one enum to name distinct things, and a reverse " +
		"lookup by value can only return one of them, so the later member silently wins. The " +
		"usual cause is a copied line whose value was never changed. Give each member its own " +
		"value.",
}

// NoDuplicateEnumValues flags an enum whose members are initialized to the same literal twice.
//
//	valid:   enum E { A, B }
//	valid:   enum E { A = 1, B = 2 }
//	valid:   enum E { A = 1, B = '1' }
//	valid:   enum E { A = -0, B = +0 }
//	valid:   enum E { A = NaN, B = NaN }
//	valid:   enum E { A = 'A', B = 'B', C = 2, D = 1 + 1 }
//	invalid: enum E { A = 1, B = 1 }
//	invalid: enum E { A = 0x10, B = 16 }
//	invalid: enum E { A = 'A', B = 'A' }
//
// Ported from oxc's `typescript/no_duplicate_enum_values`, which is what the gate runs. Read against
// `@typescript-eslint`'s own `no-duplicate-enum-values.js` as a second opinion; the two disagree in
// three places and oxc is followed in all three, with the disagreements named below because they are
// each a case where the more careful implementation is the one this port does not have.
//
// `meta.schema` is `[]` upstream and the inventory says `options: "no"`. This rule reads none, and
// there is no second tester block in oxc's file to suggest a hidden option surface.
//
// # Only two initializer shapes participate, and everything else is invisible
//
// oxc matches `Expression::NumericLiteral` and `Expression::StringLiteral` and falls through on
// every other variant. That is narrower than it sounds and it is the whole of the rule's reach:
//
//   - A unary sign is not part of a literal. `-1` parses as a prefix unary applied to `1`, so
//     `enum E { A = -1, B = -1 }` is silent, and upstream ships that exact input **commented out**
//     in its fail vector rather than as a passing case, which is upstream saying it knows and has
//     not fixed it. Six more commented-out fail cases sit beside it, all of them signs or coercions.
//     Reproduced as silence rather than improved on, and the corpus pins the observable half:
//     `A = -0, B = +0` and `A = -+-0, B = +-+0` are both shipped passing cases.
//   - A template literal is a different node even with no substitution. `A = 'A', B = `+"`A`"+` is
//     silent, and upstream marks that with a literal `TODO: Fix the following cases where there is a
//     raw template literal` above two commented-out fail cases. `@typescript-eslint` **does** resolve
//     a substitution-free template, so this is a real behavioral difference between the two
//     upstreams and oxc's answer is the one that ships here.
//   - A BigInt is `KindBigIntLiteral`, a separate kind, so `A = 1n, B = 1n` is silent in both tools.
//     Confirmed against the release binary rather than inferred from the kind name.
//   - A parenthesized literal is a `KindParenthesizedExpression` wrapping the literal, and oxc's
//     match sees the wrapper rather than the literal. `enum E { A = 1, B = (1) }` is **silent**
//     upstream, measured on the release oxlint binary alongside a firing control, and no parenthesis
//     skip is applied here for that reason. This is the case the corpus is silent about and where a
//     free-looking `SkipParentheses` would have shipped a divergence.
//   - Anything computed is invisible: `D = 1 + 1`, `D = foo()`, `B = NaN`, `B = Infinity`. `NaN` and
//     `Infinity` are identifiers rather than literals, which is why upstream can ship
//     `A = NaN, B = NaN` as a *passing* case even though the rule would otherwise call two
//     identical values duplicates.
//
// # A literal's parsed text is the comparison key, and our parser hands it over
//
// oxc compares numbers as `f64` and strings as their cooked value. Our parser has already done both:
// a `NumericLiteral`'s `Text` field holds the canonical decimal rendering of the double it parsed to
// (`0x10` arrives as `16`, `1_0` as `10`, `1.0` as `1`, `1e21` as `1e+21`, `1e400` as `Infinity`),
// and a `StringLiteral`'s `Text` holds the cooked value with escapes already decoded. Probed on all
// of those spellings before building on it.
//
// Comparing those strings is exactly equivalent to oxc comparing the underlying values, because the
// rendering is a bijection: two literals share a `Text` if and only if they parsed to the same
// double. The one place that could go wrong is the negative zero oxc's `==` treats as equal to
// positive zero, and it cannot arise, because a sign is never inside the literal.
//
// # Where the finding points, and the asymmetry upstream has between numbers and strings
//
// oxc emits one diagnostic carrying two labels, the earlier initializer and the later one, and the
// earlier is the primary position the diagnostic prints. Our harness carries one range per finding,
// so the earlier initializer is what this reports. Measured on the release binary:
// `enum E {\n  A = 1,\n  B = 1,\n}` prints `2:7`, which is `A`'s initializer rather than `B`'s.
//
// The part that is not guessable is what "the earlier initializer" means when a value appears three
// times, and **upstream answers it differently for numbers than for strings**. The number table is a
// vector that is only appended to when the value is new, so the recorded position never moves and a
// third copy is reported against the *first*. The string table is a map written on every hit, so the
// recorded position advances and a third copy is reported against the *second*. Measured on the
// release binary over `A B C` triples: the numeric enum reports twice, both at `2:7`; the string
// enum reports at `7:7` and then `8:7`.
//
// That difference is invisible in the imported corpus, whose longest fail case has one repeat of each
// value, so both readings agree on all four fail inputs. It is reproduced here deliberately, with
// fixtures that upstream does not have, because it is a difference in where a finding points rather
// than in whether one appears, and a fixture asserting only message ids cannot see it.
//
// # Once per extra copy, never pairwise
//
// Three copies of one value produce two findings and four produce three, rather than three and six.
// Measured on the release binary before writing a line, because a duplicate-detection rule has three
// plausible counting readings and the corpus separates none of them.
//
// # What the member name resolves to
//
// The help text names the offending member, and oxc reads it through `id.static_name()`, which
// resolves a computed key holding a literal. `ast.TryGetTextOfPropertyName` does the same and never
// panics on a `ComputedPropertyName` the way `Node.Text()` would. Probed against the release binary:
// `enum E { "A" = 1, ["B"] = 1 }` reports and its help says `Give B a unique value`, unquoted, which
// is what this reproduces. The member name lives in the message description rather than in a format
// verb, because `rule.Message` has no interpolation layer, so the identifying detail a reader needs
// is the span rather than the text.
var NoDuplicateEnumValues = rule.Rule{
	// No namespace prefix. The config writes `typescript/no-duplicate-enum-values` and the parity
	// guard strips the namespace on a `/` boundary.
	Name: "@typescript-eslint/no-duplicate-enum-values",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindEnumDeclaration: func(node *ast.Node) {
				declaration := node.AsEnumDeclaration()
				if declaration == nil || declaration.Members == nil {
					return
				}

				// Two tables rather than one, because a string and a number that render the same
				// are not duplicates: `enum E { A = 1, B = '1' }` is a shipped passing case and a
				// single table keyed by text would report it.
				//
				// They also differ in how they are written, which is upstream's behavior rather
				// than an oversight on this side. `seenNumbers` records only the first position a
				// value was seen at and never moves it; `seenStrings` is overwritten on every hit
				// so the position advances. See the doc comment above for the measurement.
				seenNumbers := make(map[string]*ast.Node)
				seenStrings := make(map[string]*ast.Node)

				for _, member := range declaration.Members.Nodes {
					enumMember := member.AsEnumMember()
					if enumMember == nil {
						return
					}
					initializer := enumMember.Initializer
					if initializer == nil {
						continue
					}

					switch initializer.Kind {
					case ast.KindNumericLiteral:
						text := initializer.AsNumericLiteral().Text
						if earlier, duplicated := seenNumbers[text]; duplicated {
							ctx.ReportNode(earlier, messageNoDuplicateEnumValues)
							// Deliberately not rewritten. Upstream's number table is a vector it
							// appends to only when the value is new, so a third copy reports
							// against the first rather than the second.
							continue
						}
						seenNumbers[text] = initializer

					case ast.KindStringLiteral:
						text := initializer.AsStringLiteral().Text
						if earlier, duplicated := seenStrings[text]; duplicated {
							ctx.ReportNode(earlier, messageNoDuplicateEnumValues)
						}
						// Written on every hit, duplicate or not, which is what makes a third copy
						// report against the second. Upstream's `insert` returns the displaced
						// value and stores the new one in the same call.
						seenStrings[text] = initializer
					}
				}
			},
		}
	},
}
