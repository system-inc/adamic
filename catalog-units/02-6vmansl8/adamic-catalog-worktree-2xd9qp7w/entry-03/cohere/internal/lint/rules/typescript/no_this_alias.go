package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoThisAlias = rule.Message{
	Id: "thisAssignment",
	Description: "This binding holds `this` under another name. The alias exists to carry the " +
		"receiver into a nested `function`, which rebinds `this` to something else, and an arrow " +
		"function closes over `this` directly and needs no carrier. Every later reader now has to " +
		"establish which of the two names is the real receiver and whether they are still the " +
		"same object. Use an arrow function and write `this`.",
}

var messageNoThisDestructure = rule.Message{
	Id: "thisDestructure",
	Description: "This pattern pulls members off `this` into local bindings. Each one is read " +
		"once, at the moment of destructuring, so a later write to the property is invisible " +
		"here, and a method taken this way has lost its receiver and throws when called. Read " +
		"through `this` at the point of use instead.",
}

// NoThisAliasOptions configures which aliases of `this` report.
//
// The authoritative surface is oxc's `NoThisAliasConfig`, which derives `JsonSchema` under
// `serde(rename_all = "camelCase", deny_unknown_fields)` and carries exactly two fields. Measured
// against the release binary rather than read off an inventory column: a config naming a third key
// is refused outright with `unknown field, expected one of allowDestructuring, allowNames,
// allowedNames`, which is also how the third accepted spelling was confirmed to be an alias of the
// second rather than a field of its own.
type NoThisAliasOptions struct {
	// ReportDestructuring is upstream's `allowDestructuring` inverted, so that the Go zero value is
	// upstream's default.
	//
	// This inversion is the single most dangerous line in the port and it is deliberate. Upstream's
	// default is `allowDestructuring: true`, written as a hand-rolled `Default` impl rather than
	// derived, precisely because `false` is not what the rule wants when nobody configured it. A Go
	// field spelled `AllowDestructuring` would default to `false`, and every caller who configured
	// nothing would get the strict behavior: `const { props } = this` reporting in a tree where
	// upstream is silent. The name is flipped so that `NoThisAliasOptions{}` behaves the way the
	// rule is documented to behave. `decodeThisAliasOptions` is what maps the upstream spelling
	// onto it, and the zero-value fixture is what pins it.
	ReportDestructuring bool

	// AllowedNames are variable names permitted to hold `this`.
	//
	// Upstream accepts two spellings for this one field, `allowedNames` and `allowNames`, the
	// second being a typo oxc shipped and then kept as a serde alias for compatibility. They are
	// not two options: supplying both at once is a hard configuration error on the release binary,
	// not a merge, which is what settles that they name one field.
	//
	// The exemption reaches identifier aliases only. Measured with `d` in the allow list and
	// `allowDestructuring` off, `const { d } = this` still reports, because the destructure arm
	// never consults the list at all. That is upstream's structure rather than an oversight worth
	// improving on, and the fixture for it is in the silent-versus-reporting pair below.
	AllowedNames []string
}

// noThisAliasRawOptions is the wire shape, which differs from the shape the rule reasons with.
//
// It exists only so the JSON keys can stay upstream's while the struct the rule reads has the
// inverted default described above. A decoder is the right place for that flip because it is the
// one point where "the config said nothing" and "the config said false" are still distinguishable.
type noThisAliasRawOptions struct {
	AllowDestructuring *bool    `json:"allowDestructuring"`
	AllowedNames       []string `json:"allowedNames"`
	AllowNames         []string `json:"allowNames"`
}

// NoThisAlias flags a local binding that holds `this`.
//
//	valid:   const { props, state } = this;              // destructuring is allowed by default
//	valid:   const self = this;                          // only under allowedNames: ["self"]
//	valid:   this.self = this;                           // a member target is never an alias
//	valid:   const self = (this);                        // upstream does not see through parens here
//	invalid: const self = this;
//	invalid: foo = this;
//	invalid: (foo as any) = this;
//	invalid: const { props, state } = this;              // under allowDestructuring: false
//
// Ported from `typescript/no-this-alias`, which oxc in turn ports from
// `@typescript-eslint/no-this-alias`.
//
// # Two messages, and which one an input gets is decided by the target's shape alone
//
// The rule produces two distinct findings and they are not severities of one another. An
// identifier target aliases the receiver itself and always reports; a destructuring target copies
// members off the receiver and reports only when the option asks for it. Because they are separate
// judgments, the allow list applies to the first and not to the second, and an input can be silent
// under one and reporting under the other with nothing else changed.
//
// # Parentheses fall opposite ways on the two sides of the same assignment
//
// This is the part a reading of the upstream source does not settle, and it was measured on the
// release binary rather than reasoned about.
//
// On the right, upstream's `rhs_is_this_reference` destructures `Expression::ThisExpression`
// directly with no unwrapping, so `const self = (this)` is silent, and so are `this as any` and
// `this!`. Our parser materializes a ParenthesizedExpression here exactly as it does everywhere
// else, so reproducing that silence costs nothing: the initializer's kind is simply not
// `KindThisKeyword` and the arm declines. The trap is the other direction, because skipping parens
// on an initializer reads as a free correctness improvement and would report three inputs upstream
// leaves alone.
//
// On the left, upstream reaches the identifier through `get_expression()` and then
// `get_identifier_reference()`, both of which unwrap. Measured: `(g) = this` reports, `((q)) = this`
// reports, `(h as any) = this` reports, `(r satisfies any) = this` reports, `(<any>r) = this`
// reports, and `(t!) = this` reports. In every case the finding points at the bare identifier and
// the span excludes the parentheses. So this side needs a recursive unwrap that the other side must
// not have, in one rule.
//
// # The assignment arm ignores the operator, and `+=` is the proof
//
// Upstream matches `AssignmentExpression` and never looks at which operator it carries. Measured:
// `d ??= this`, `e ||= this` and `g += this` all report. The last one is not aliasing under any
// reading of the rule's intent, since `+=` on a receiver produces a string, but it is what upstream
// decides and a port that filtered on `KindEqualsToken` would be silently narrower than the tool
// the gate runs. Reproduced deliberately, with a fixture, so the next reader finds the measurement
// rather than the intuition.
//
// # A member target is silent, and that is a real decision rather than a gap
//
// `this.self = this` and `m.n.o = this` are both silent upstream, because the simple-target arm
// asks for an identifier reference and a member expression yields none. That matters more than it
// looks: `this.self = this` is the single most common hand-written form of this exact defect, and
// upstream declines it. Measured on the release binary in both shapes.
var NoThisAlias = rule.Rule{
	Name: "@typescript-eslint/no-this-alias",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		// Declining the file rather than the node, matching upstream's `should_run`, which reads
		// `ctx.source_type().is_typescript()`. Measured: the same `const self = this;` reports in a
		// `.ts` file and is silent in a `.js` one.
		if !isTypeScriptSourceFile(ctx.SourceFile.FileName().AsString()) {
			return nil
		}

		parsed, _ := rule.OptionsAs[NoThisAliasOptions](options)

		allowed := make(map[string]bool, len(parsed.AllowedNames))
		for _, name := range parsed.AllowedNames {
			allowed[name] = true
		}

		reportIdentifierUnlessAllowed := func(identifier *ast.Node) {
			if allowed[identifier.Text()] {
				return
			}
			ctx.ReportNode(identifier, messageNoThisAlias)
		}

		return rule.Listeners{
			ast.KindVariableDeclaration: func(node *ast.Node) {
				declaration := node.AsVariableDeclaration()
				if !isThisKeyword(declaration.Initializer) {
					return
				}

				name := declaration.Name()
				if name == nil {
					return
				}

				if name.Kind == ast.KindIdentifier {
					reportIdentifierUnlessAllowed(name)
					return
				}

				// Every remaining name kind is a binding pattern. Upstream reaches this arm only
				// for `ObjectPattern` and `ArrayPattern`, which are the only two things a
				// `BindingPattern` can be once the identifier case is taken, so the else branch is
				// faithful without enumerating the kinds.
				if !parsed.ReportDestructuring {
					return
				}
				ctx.ReportNode(name, messageNoThisDestructure)
			},

			ast.KindBinaryExpression: func(node *ast.Node) {
				expression := node.AsBinaryExpression()

				// Upstream anchors on `AssignmentExpression`, a node kind of its own. Our parser
				// folds every binary operator into one kind, so the assignment test that upstream
				// gets from its grammar has to be written here. `IsAssignmentOperator` covers `=`
				// and every compound form, which is the whole set upstream's arm receives.
				if !ast.IsAssignmentOperator(expression.OperatorToken.Kind) {
					return
				}
				if !isThisKeyword(expression.Right) {
					return
				}

				switch expression.Left.Kind {
				// A destructuring assignment target parses as a literal rather than as a pattern,
				// because the two are only distinguishable once the `=` is seen. `[i] = this` is an
				// ArrayLiteralExpression on the left and `({j} = this)` an ObjectLiteralExpression,
				// and they are upstream's `ArrayAssignmentTarget` and `ObjectAssignmentTarget`.
				case ast.KindArrayLiteralExpression, ast.KindObjectLiteralExpression:
					if !parsed.ReportDestructuring {
						return
					}
					ctx.ReportNode(expression.Left, messageNoThisDestructure)

				default:
					// Everything else is upstream's identifier arm and its simple-target arm, which
					// reach the same conclusion by different routes: unwrap to an identifier
					// reference, report it if there is one, decline silently if there is not. A
					// member expression falls out here with no identifier and is therefore silent,
					// which is upstream's behavior and is measured in the doc comment above.
					target := unwrapAssignmentTarget(expression.Left)
					if target == nil || target.Kind != ast.KindIdentifier {
						return
					}
					reportIdentifierUnlessAllowed(target)
				}
			},
		}
	},
}

// DecodeNoThisAliasOptions maps upstream's JSON keys onto the inverted struct the rule reads.
//
// It is a hand-written decoder rather than `rule.DecodeOptionsInto` because two of the three keys
// need translating rather than binding: `allowDestructuring` is inverted, and `allowNames` is a
// second spelling of `allowedNames`. Neither can be expressed as a struct tag.
func DecodeNoThisAliasOptions(raw []byte) (any, error) {
	decoded, err := rule.DecodeOptionsInto[noThisAliasRawOptions]()(raw)
	if err != nil {
		return NoThisAliasOptions{}, err
	}

	wire, _ := decoded.(noThisAliasRawOptions)

	options := NoThisAliasOptions{}

	// Absent means upstream's default of true, which is this field being false. Only an explicit
	// `false` on the wire turns reporting on, which is why the wire field is a pointer.
	if wire.AllowDestructuring != nil {
		options.ReportDestructuring = !*wire.AllowDestructuring
	}

	// The two spellings are one field upstream, and supplying both is a configuration error there
	// rather than a merge. There is no error channel for that distinction on this side of the
	// config, so the correctly-spelled key wins when both appear and the typo'd one is read only
	// when it is the only one present.
	options.AllowedNames = wire.AllowedNames
	if options.AllowedNames == nil {
		options.AllowedNames = wire.AllowNames
	}

	return options, nil
}

// isThisKeyword is upstream's `rhs_is_this_reference`, and its narrowness is the point.
//
// Upstream destructures `Expression::ThisExpression` with no unwrapping, so anything wrapping the
// keyword defeats the whole rule. Written as a named function so that the absence of a paren skip
// is a visible decision at one site rather than an omission spread over two call sites, and so the
// three inputs that pin it have somewhere to point: `const self = (this)`, `const self = this as
// any` and `const self = this!` are all silent on the release binary.
func isThisKeyword(expression *ast.Node) bool {
	return expression != nil && expression.Kind == ast.KindThisKeyword
}

// unwrapAssignmentTarget reaches the identifier an assignment target ultimately writes to.
//
// This is upstream's `get_expression()` followed by `get_identifier_reference()`, both of which
// unwrap, against a tree that keeps the wrappers as real nodes. The set of kinds here is the set
// measured to report on the release binary: parentheses recursively, and the three type-level
// wrappers that are erased at runtime and therefore leave a writable target underneath.
//
// It returns whatever it lands on rather than only an identifier, so the caller's kind test is
// where the decline is written and a member expression declines at one visible place.
func unwrapAssignmentTarget(target *ast.Node) *ast.Node {
	for target != nil {
		switch target.Kind {
		case ast.KindParenthesizedExpression,
			ast.KindAsExpression,
			ast.KindSatisfiesExpression,
			ast.KindTypeAssertionExpression,
			ast.KindNonNullExpression:
			target = target.Expression()
		default:
			return target
		}
	}
	return nil
}
