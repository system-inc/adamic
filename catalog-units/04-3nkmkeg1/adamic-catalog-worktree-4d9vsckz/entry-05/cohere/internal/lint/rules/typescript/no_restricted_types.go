package typescript

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoRestrictedTypes flags types a project has decided not to use, named in configuration.
//
//	valid:   let value: string;                       with no types configured
//	valid:   let value: Foo<Bar>;                      with only `Foo` banned, since the name differs
//	invalid: let value: string;                        with `{"string": true}`
//	invalid: let value: {};                            with `{"{}": "Use object instead."}`
//	invalid: let value: Foo;                           with `{"Foo": {"fixWith": "Bar"}}`
//
// The rule enforces nothing on its own. It exists so a project can write down "we do not use this
// type here, and this is what to use instead", and have the decision enforced rather than
// remembered. Every finding names the banned type and appends whatever the configuration said about
// it.
//
// # The name is the configured key with whitespace removed, and so is the type
//
// Both sides are stripped, so a key written `Foo<Bar>` matches source written `Foo<  Bar  >`. The
// message names the STRIPPED form while the finding spans the raw text: measured on the installed
// 8.67.0 build, `let value: Foo<  Bar  >;` with `Foo<Bar>` banned underlines the whole raw
// `Foo<  Bar  >` and says "Don't use `Foo<Bar>`".
//
// # Where the trees differ, and why this rule has fewer handlers than upstream
//
// Upstream needs three separate handlers for a type reference, a class `implements` entry, and an
// interface `extends` entry, because estree gives those three different node types. Probed here:
// all three parse as `KindTypeReference`, so ONE handler covers them, and the heritage handlers
// upstream writes would be duplicates rather than additions.
//
// That equivalence is not an assumption. Driven through the installed build on all four heritage
// shapes: `implements Bar` and `extends Bar` each report once, on the name; `implements Bar<Baz>`
// and `extends Bar<Baz>` each report TWICE, once on the name and once on the whole generic, in that
// order. This port produces the same pairs in the same order from the single handler, because
// upstream's own type-reference handler already does both halves.
//
// # Two other shapes, each with a size test
//
// An empty tuple `[]` and an empty type literal `{}` are reported under those spellings, and only
// when they are empty. Upstream tests the element and member counts for exactly that reason: a
// non-empty `{ a: 1 }` is not the banned `{}`.
//
// # A fix and a suggestion are different offers, and the configuration chooses
//
// `fixWith` produces a FIX, which the edit engine applies unattended. `suggest` produces
// suggestions, which a human picks from. Reproducing that split is part of the port rather than a
// detail: shipping a suggestion as a fix means the edit engine rewrites code upstream would only
// have offered to rewrite.
//
// The two are NOT exclusive and an empty `fixWith` is not a fix, both measured rather than read.
// Upstream's corpus asserts thirteen fixer outputs and ZERO suggestions, so nothing imported
// exercises that half at all, and the first version of this rule got both details wrong while every
// imported case stayed green. Driven through the installed 8.67.0 build: an entry carrying both
// produces one finding with a fix AND the suggestions, and `fixWith: ""` produces a finding with no
// fix rather than one deleting the type.
//
// # Cost
//
// Every listener is registered only when the configuration bans something it could match, so a
// project banning nothing registers nothing and the rule costs a map lookup per file.
var NoRestrictedTypes = rule.Rule{
	Name: "@typescript-eslint/no-restricted-types",

	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, ok := rule.OptionsAs[NoRestrictedTypesOptions](options)
		if !ok || len(settings.Types) == 0 {
			// Configured nothing, so there is nothing to enforce. Upstream defaults `types` to an
			// empty object and reaches the same silence one lookup later.
			return rule.Listeners{}
		}

		// checkBannedType is upstream's `checkBannedTypes`. `name` is what the configuration is
		// looked up by and what the message says; the node is what the finding underlines.
		checkBannedType := func(node *ast.Node, name string) {
			banned, isBanned := settings.Types[name]
			if !isBanned || banned.Allowed {
				return
			}
			message := rule.Message{
				Id:          "bannedTypeMessage",
				Description: "Don't use `" + name + "` as a type." + banned.customMessage(),
			}

			nodeRange := rule.TokenRange(ctx.SourceFile, node)

			// A fix and suggestions are NOT exclusive, and the first version of this rule returned
			// after the fix and dropped them. Measured on the installed 8.67.0 build:
			// `{fixWith: 'Bar', suggest: ['Baz']}` produces a finding carrying BOTH, so a reader who
			// declines the automatic repair still gets the offer.
			//
			// An empty `fixWith` produces NO fix. Upstream writes `const fixWith = bannedType.fixWith`
			// and then `fix: fixWith ? ... : null`, so the empty string is falsy and there is no
			// repair. Measured: `{fixWith: ''}` reports with no fix rather than deleting the type.
			// This is the reason the field can be a plain string rather than a pointer, and it is the
			// opposite of what an absent-versus-empty pointer would express.
			fixes := []rule.Fix{}
			if banned.FixWith != "" {
				fixes = append(fixes, rule.ReplaceRange(nodeRange, banned.FixWith))
			}

			suggestions := make([]rule.Suggestion, 0, len(banned.Suggest))
			for _, replacement := range banned.Suggest {
				suggestions = append(suggestions, rule.Suggestion{
					Message: rule.Message{
						Id:          "bannedTypeReplacement",
						Description: "Replace `" + name + "` with `" + replacement + "`.",
					},
					Fixes: []rule.Fix{rule.ReplaceRange(nodeRange, replacement)},
				})
			}

			switch {
			case len(fixes) > 0 && len(suggestions) > 0:
				ctx.Report(rule.Diagnostic{
					Range:       nodeRange,
					Message:     message,
					SourceFile:  ctx.SourceFile,
					Fixes:       fixes,
					Suggestions: suggestions,
				})
			case len(fixes) > 0:
				ctx.ReportNodeWithFixes(node, message, fixes...)
			case len(suggestions) > 0:
				ctx.ReportNodeWithSuggestions(node, message, suggestions...)
			default:
				ctx.ReportNode(node, message)
			}
		}

		// checkBannedNode is the same thing with the name read off the source text, which is what
		// upstream's default parameter does.
		checkBannedNode := func(node *ast.Node) {
			checkBannedType(node, noRestrictedTypesNameOf(ctx, node))
		}

		listeners := rule.Listeners{}

		// A keyword listener is registered only if that keyword is actually banned, which is
		// upstream's `keywordSelectors` reduction. The keyword's name is passed rather than read
		// from source, because upstream passes the literal too.
		//
		// This registration test is a COST decision and not a correctness one, measured rather than
		// assumed: a mutant registering every keyword listener unconditionally survives the whole
		// suite, and it survives because it is equivalent. The lookup inside `checkBannedType`
		// declines any keyword the configuration did not name, so the extra listeners produce the
		// same findings while walking more nodes. It is kept because upstream builds the same table
		// and because eleven listeners firing on every keyword in every file is real work for a
		// project that banned none of them. The fixture asserting the lookup half is load-bearing is
		// in the silent test, at the row banning `number` in a file writing `string`.
		for keyword, kind := range noRestrictedTypesKeywordKinds {
			if _, isBanned := settings.Types[keyword]; !isBanned {
				continue
			}
			bannedKeyword := keyword
			listeners[kind] = func(node *ast.Node) {
				checkBannedType(node, bannedKeyword)
			}
		}

		listeners[ast.KindTypeReference] = func(node *ast.Node) {
			reference := node.AsTypeReferenceNode()
			if reference.TypeName != nil {
				checkBannedNode(reference.TypeName)
			}
			// A generic reference is two candidates: the bare name and the whole thing. Upstream
			// checks both in this order and so does this, which is why `Bar<Baz>` with both banned
			// reports twice with the name first.
			if reference.TypeArguments != nil {
				checkBannedNode(node)
			}
		}

		listeners[ast.KindTupleType] = func(node *ast.Node) {
			elements := node.AsTupleTypeNode().Elements
			if elements == nil || len(elements.Nodes) == 0 {
				checkBannedNode(node)
			}
		}

		listeners[ast.KindTypeLiteral] = func(node *ast.Node) {
			members := node.AsTypeLiteralNode().Members
			if members == nil || len(members.Nodes) == 0 {
				checkBannedNode(node)
			}
		}

		return listeners
	},
}

// noRestrictedTypesKeywordKinds is upstream's `TYPE_KEYWORDS`, mapping a configurable spelling onto
// the node kind that spelling parses to.
//
// These are separate from the type-reference handler because a keyword is not a reference: `string`
// parses as its own kind and has no name node to read.
var noRestrictedTypesKeywordKinds = map[string]ast.Kind{
	"bigint":    ast.KindBigIntKeyword,
	"boolean":   ast.KindBooleanKeyword,
	"never":     ast.KindNeverKeyword,
	"null":      ast.KindNullKeyword,
	"number":    ast.KindNumberKeyword,
	"object":    ast.KindObjectKeyword,
	"string":    ast.KindStringKeyword,
	"symbol":    ast.KindSymbolKeyword,
	"undefined": ast.KindUndefinedKeyword,
	"unknown":   ast.KindUnknownKeyword,
	"void":      ast.KindVoidKeyword,
}

// noRestrictedTypesNameOf reads a node's source text with every whitespace character removed.
//
// Upstream's `removeSpaces` uses the regular expression `\s`, which is Unicode-aware in JavaScript,
// so `text.IsWhitespace`, JavaScript's set, is the matching predicate rather than a test against the
// five ASCII ones or Go's `unicode.IsSpace`.
func noRestrictedTypesNameOf(ctx rule.Context, node *ast.Node) string {
	nodeRange := rule.TokenRange(ctx.SourceFile, node)
	return noRestrictedTypesRemoveSpaces(ctx.SourceFile.Text()[nodeRange.Pos():nodeRange.End()])
}

// noRestrictedTypesRemoveSpaces is upstream's `removeSpaces`, applied to both the configured key and
// the source text so the two are compared in the same shape.
func noRestrictedTypesRemoveSpaces(typeSpelling string) string {
	return strings.Map(func(r rune) rune {
		if text.IsWhitespace(r) {
			return -1
		}
		return r
	}, typeSpelling)
}

// NoRestrictedTypesOptions is the rule's option surface.
//
// Upstream's `types` is an object whose values carry three different shapes, and the decoder below
// flattens them into one struct rather than making the rule ask which shape it got.
type NoRestrictedTypesOptions struct {
	// Types maps a whitespace-stripped type spelling onto what to say and offer about it.
	Types map[string]NoRestrictedTypesBan
}

// NoRestrictedTypesBan is one entry's resolved meaning.
type NoRestrictedTypesBan struct {
	// Allowed records an entry the configuration wrote as `false` or `null`, meaning "not banned".
	//
	// Upstream's RULE tests `bannedType == null || bannedType === false` and declines, so the
	// behavior is real in its source. Its SCHEMA rejects both spellings before the rule can see
	// them: measured against the installed 8.67.0 build, `{"Foo": null}` and `{"Foo": false}` both
	// fail configuration validation and the linter refuses to start. So this branch reproduces
	// upstream's code and is unreachable through upstream's own configuration surface. It is kept
	// because cohere has no schema layer rejecting those spellings, which means a config here CAN
	// contain one, and silently banning a type somebody wrote `false` against would be the worst
	// available reading.
	Allowed bool

	// Message is the text appended after the standard sentence, from either the string form or the
	// object's `message` key.
	Message string

	// FixWith produces a FIX the edit engine applies unattended.
	//
	// A plain string rather than a pointer, deliberately, because upstream cannot distinguish an
	// absent key from an empty one either: it tests the value's truthiness, so `fixWith: ""` means
	// no repair rather than "replace with nothing". Measured on the installed 8.67.0 build, which
	// reports that case with no fix attached. A pointer here would express a distinction upstream
	// does not have and would delete the type.
	FixWith string

	// Suggest produces suggestions a human chooses between.
	Suggest []string
}

// customMessage renders the trailing half of the finding.
//
// Upstream returns the empty string for a bare `true` and otherwise a SPACE followed by the message,
// so the standard sentence and the custom one are separated by exactly one space and a bare ban ends
// at the period.
func (b NoRestrictedTypesBan) customMessage() string {
	if b.Message == "" {
		return ""
	}
	return " " + b.Message
}

// DefaultNoRestrictedTypesOptions is upstream's configured-nothing behavior: ban nothing.
func DefaultNoRestrictedTypesOptions() NoRestrictedTypesOptions {
	return NoRestrictedTypesOptions{Types: map[string]NoRestrictedTypesBan{}}
}

// noRestrictedTypesRawOptions is the wire shape.
//
// `Types` is deliberately `map[string]json.RawMessage` rather than a typed value, because upstream's
// schema is a `oneOf` over three shapes and Go has no sum type to decode into. Each entry is
// re-decoded below against whichever shape it turns out to be.
type noRestrictedTypesRawOptions struct {
	Types map[string]json.RawMessage `json:"types"`
}

// DecodeNoRestrictedTypesOptions maps the wire shape onto what the rule reads.
//
// Hand-written rather than `rule.DecodeOptionsInto` because a value here is a boolean, a string, or
// an object, and a generic decode into any single Go type discards two of the three. It also has to
// distinguish three falsy-looking spellings that mean different things: `true` bans with no message,
// `false` and `null` do not ban at all, and an object with no `message` bans with no message.
func DecodeNoRestrictedTypesOptions(raw []byte) (any, error) {
	settings := DefaultNoRestrictedTypesOptions()
	if len(raw) == 0 {
		return settings, nil
	}

	var wire noRestrictedTypesRawOptions
	if err := rule.UnmarshalOptions(raw, &wire); err != nil {
		return nil, err
	}

	for spelling, entry := range wire.Types {
		ban, include, err := decodeNoRestrictedTypesBan(entry)
		if err != nil {
			return nil, fmt.Errorf("types[%q]: %w", spelling, err)
		}
		if !include {
			continue
		}
		// The key is stripped the same way the source text is, so `"Foo < Bar >"` and `"Foo<Bar>"`
		// are one entry rather than two that never match.
		settings.Types[noRestrictedTypesRemoveSpaces(spelling)] = ban
	}
	return settings, nil
}

// decodeNoRestrictedTypesBan resolves one entry's three possible shapes.
//
// The second return says whether to record the entry at all. A malformed value is dropped rather
// than failing the whole decode, which matches upstream: its schema rejects such a value before the
// rule ever runs, so the rule itself has no behavior for one.
//
// An object is the exception, and it fails the decode. The object's schema is closed
// (`additionalProperties: false`), and strict decoding means a misspelled `mesage` refuses the
// object; dropping the entry for that would silently stop banning the type, where the lenient decode
// before it dropped only the key (#4a4yse4).
func decodeNoRestrictedTypesBan(entry json.RawMessage) (NoRestrictedTypesBan, bool, error) {
	// Go whitespace: raw JSON bytes of a rule's options, whose whitespace is the same in both sets.
	trimmed := strings.TrimSpace(string(entry))
	if trimmed == "null" {
		// Upstream's `bannedType == null` test treats a null entry as not banned.
		return NoRestrictedTypesBan{Allowed: true}, true, nil
	}

	var asBoolean bool
	if err := json.Unmarshal(entry, &asBoolean); err == nil {
		// `true` bans with the standard message; `false` explicitly does not ban.
		return NoRestrictedTypesBan{Allowed: !asBoolean}, true, nil
	}

	var asString string
	if err := json.Unmarshal(entry, &asString); err == nil {
		return NoRestrictedTypesBan{Message: asString}, true, nil
	}

	var asObject struct {
		FixWith string   `json:"fixWith"`
		Message string   `json:"message"`
		Suggest []string `json:"suggest"`
	}
	if strings.HasPrefix(trimmed, "{") {
		if err := rule.UnmarshalOptions(entry, &asObject); err != nil {
			return NoRestrictedTypesBan{}, false, err
		}
		return NoRestrictedTypesBan{
			Message: asObject.Message,
			FixWith: asObject.FixWith,
			Suggest: asObject.Suggest,
		}, true, nil
	}
	return NoRestrictedTypesBan{}, false, nil
}
