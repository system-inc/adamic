package react

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// forbiddenEntity is one character the rule refuses in JSX text, and what may replace it.
//
// Upstream's `forbid` array is heterogeneous: an entry is either a bare string, which names the
// character and offers nothing, or an object carrying `char` and `alternatives`. The two spellings
// produce different messages and different repairs, so the distinction survives into this struct
// rather than being normalized away. `hasAlternatives` is what separates `"&"` from
// `{char: "&", alternatives: []}`, which upstream renders differently: the first is
// `unescapedEntity` with no suggestions and the second is `unescapedEntityAlts` with an empty
// alternatives list, rendering as "`&` can be escaped with ." Measured against the installed rule,
// both spellings on the same input.
type forbiddenEntity struct {
	character       string
	alternatives    []string
	hasAlternatives bool
}

// NoUnescapedEntitiesOptions is the rule's option surface.
//
// `Forbid` has to be distinguishable as absent, which is why this holds a pointer rather than a
// slice and why the decoder below is hand-written. Absent means upstream's four defaults; an empty
// array means nothing is forbidden and the rule is silent. A plain slice collapses those two into
// one nil, and a rule reading `len(forbid) == 0` as "use the defaults" would report on a
// configuration that explicitly asked for silence. Measured on the installed rule:
// `{forbid: []}` is clean on source the defaults report twice.
type NoUnescapedEntitiesOptions struct {
	Forbid *[]forbiddenEntity
}

// defaultForbiddenEntities is upstream's `DEFAULTS`, copied character for character.
//
// The order matters and is upstream's, because it decides the order findings are reported in when
// one line carries more than one kind. Upstream loops entities in the outer position and text
// positions in the inner one, so every `>` in a line reports before the first `"` does, even when
// the quote comes first in the text. That is preserved below; see reportEntities.
//
// **Order within this list is not observable and that is measured, not assumed.** A mutant swapping
// the first two entries survived the whole fixture set. The loops below are source-major, so the
// entity list is only consulted per byte, and no byte can equal two different single characters, so
// at most one entry ever matches a given position. A mutant DUPLICATING an entry is caught in 10
// lines, which is what proves the inner loop is live rather than the order argument being a cover
// for a dead branch. The order is upstream's anyway, because matching the source costs nothing.
//
// `<` and `{` are absent here and their absence is upstream's own note: both are syntax errors
// when written accidentally in JSX text under the parsers upstream targets, so a rule catching
// them would only ever fire on source that does not parse. Our parser is more permissive than
// that, but adding them would be a divergence nothing asked for, so they stay out.
var defaultForbiddenEntities = []forbiddenEntity{
	{character: ">", alternatives: []string{"&gt;"}, hasAlternatives: true},
	{character: `"`, alternatives: []string{"&quot;", "&ldquo;", "&#34;", "&rdquo;"}, hasAlternatives: true},
	{character: "'", alternatives: []string{"&apos;", "&lsquo;", "&#39;", "&rsquo;"}, hasAlternatives: true},
	{character: "}", alternatives: []string{"&#125;"}, hasAlternatives: true},
}

var messageUnescapedEntity = rule.Message{
	Id: "unescapedEntity",
	Description: "This character is written raw in JSX text where it should be an HTML entity. " +
		"Write the entity instead.",
}

var messageUnescapedEntityAlts = rule.Message{
	Id: "unescapedEntityAlts",
	Description: "This character is written raw in JSX text where it should be an HTML entity. " +
		"Write one of the escapes instead.",
}

var messageReplaceWithAlt = rule.Message{
	Id:          "replaceWithAlt",
	Description: "Replace the character with its HTML entity.",
}

// NoUnescapedEntities flags a raw `>`, `"`, `'` or `}` written in JSX text.
//
//	valid:   <div>Here is some text!</div>
//	valid:   <div>I&rsquo;ve escaped some entities: &gt; &lt; &amp;</div>
//	valid:   <div>{">" + "<" + "&" + '"'}</div>
//	invalid: <div>> default parser</div>
//	invalid: <div>'</div>
//	invalid: <div>{"Unbalanced braces"}}</div>
//
// Ported from `eslint-plugin-react`'s `no-unescaped-entities.js`, which is the authority. Every
// divergence below was measured by driving the installed build at
// `~/Projects/ahra/node_modules/eslint-plugin-react`, version 7.37.5, through the ESLint Linter
// API rather than reasoned from the source.
//
// # The listener, and why only JsxText
//
// Upstream's selector is `'Literal, JSXText'` guarded by `jsxUtil.isJSX(node.parent)`, which
// answers true only for `JSXElement` and `JSXFragment`. That reads as two node kinds and is one:
// **a `Literal` can never be a direct child of a JSX element or fragment.** Measured by walking
// every node whose parent is one of the two over `<div title="q">t{"x"}<b/>{/*c*/}</div>` and over
// a fragment: the child kinds are exactly JSXOpeningElement, JSXText, JSXExpressionContainer,
// JSXElement and JSXClosingElement, with no Literal among them. A string in an attribute has
// parent `JSXAttribute` and a string in braces has parent `JSXExpressionContainer`, and both fail
// `isJSX`. Confirmed against behaviour too: `<div title="a>b" />` and `<div>{"has > and quote"}</div>`
// are both clean upstream, and the second is one of upstream's own passing cases.
//
// So the `Literal` half of upstream's selector is dead, and this rule listens on JsxText alone.
// The parent check upstream performs is likewise unnecessary here rather than dropped as an
// optimization: our parser produces a JsxText node only as a child of a JsxElement or a JsxFragment,
// which the same probe measured. Adding an unreachable guard would read to the next person as a case
// somebody handled.
//
// # Our parser accepts source the corpus calls unparseable, so four more cases are live
//
// The corpus gates four invalid cases behind `allowsInvalidJSX`, computed from the installed
// acorn-jsx version, because espree rejects a bare `>` or `}` in JSX text with a parse error rather
// than reporting. Measured: `var a = <div>> hi</div>;` through the installed rule returns
// "Parsing error: Unexpected token `>`" and no finding at all.
//
// typescript-go is permissive about both. Measured on our own parser: `<div>> default parser</div>`
// produces a JsxText holding `"> default parser"`, and `<div>{"x"}}</div>` produces a JsxText
// holding `"}"`. So all four gated cases parse here, reach the rule, and are fixtured below. They
// are the only fixtures in this file that could not be run against the reference implementation as
// installed, and each was verified against the reference's own asserted output in the corpus rather
// than against a run.
//
// # Positions, and why a byte scan is upstream's character scan
//
// Upstream reads `sourceCode.lines`, slices each line by the node's start and end columns, and
// walks the slice by index. This walks the node's own source range directly, which is the same
// text: measured on our parser, the raw slice `text[node.Pos():node.End()]` equals the JsxText's own
// `.Text` field exactly, entities uncooked, for `&gt;&amp;` as well as for plain text. There is no
// line reconstruction to do, and the per-line slicing upstream performs is bookkeeping it needs
// because ESLint hands it lines rather than offsets.
//
// The scan is byte-wise. Every character this rule can forbid by default is one ASCII byte, and a
// byte scan over UTF-8 cannot produce a false hit inside a multi-byte sequence because continuation
// bytes are all above 0x7F. A configured `forbid` entry longer than one character is silent, which
// is upstream's behaviour too and not a limitation of the scan: upstream compares
// `rawLine[index] === entities[j]`, a single character against a string, so `{forbid: ["oo"]}` can
// never match. Measured clean on `<div>foo bar</div>`.
//
// # Where the finding points, and the one place this deliberately differs
//
// Upstream reports with `loc: {line, column}` and no end, so ESLint renders a **zero-width point**
// at the offending character. Measured: every diagnostic comes back with `line` and `column` set
// and `endLine`/`endColumn` absent. Our Diagnostic carries a range and there is no way to express
// a point, so this reports the one character itself, a range of length one starting where upstream's
// point starts. That is the smallest honest span and it is a divergence stated rather than hidden:
// the start position is identical, the end is one byte later.
//
// # The repair is a suggestion, and it rewrites the whole node upstream-style
//
// `meta.hasSuggestions` is true and `meta.fixable` is unset, so upstream offers repairs a human
// chooses and never applies one unattended. That distinction is ported: escaping a `'` inside JSX
// text changes the rendered output when the surrounding text was meant literally, and picking which
// of four quote entities the author wanted is a choice rather than a mechanical equivalence.
//
// Upstream's fixer replaces the ENTIRE JsxText node with the node's raw text carrying one character
// substituted. Replacing just the character produces byte-identical output and would be the obvious
// simplification. It is not taken, because it is not the same edit: the fix engine reasons about
// overlaps between proposed ranges, and two findings in one JsxText that upstream would have made
// mutually exclusive (both replacing the whole node) would become independently applicable here.
// Reproducing the decision means reproducing the range it covers, so the range is the node.
//
// # An upstream crash, reproduced as a decline
//
// `{forbid: [{char: "&"}]}` passes upstream's schema, which marks neither `char` nor `alternatives`
// as required, and then **crashes the linter**: `entities[j].alternatives.map` throws
// "Cannot read properties of undefined (reading 'map')". Measured, with the stack landing at
// `no-unescaped-entities.js:121`. That is a defect rather than a judgment, so it is not reproduced.
// An object entry with no `alternatives` key is dropped here, which makes it silent rather than
// fatal. `{char: "&", alternatives: []}` is a different configuration and DOES report, with an
// empty alternatives list, and that one is reproduced exactly.
var NoUnescapedEntities = rule.Rule{
	Name: "react/no-unescaped-entities",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		entities := configuredEntities(options)
		// An empty forbid list makes every input clean, and returning no listeners is how this rule
		// says that. **This branch is a cost optimization and nothing else**, measured rather than
		// assumed: a mutant neutralizing it survived the whole fixture set, because the inner loop
		// below runs over the same empty slice and reports nothing either way. No input can
		// distinguish the two versions. The inverse mutation, forcing the return always, fails 26
		// lines, so the statement is reached and the fixtures see through it.
		//
		// Kept because the saving is real, the walk skips 18,604 JsxText nodes on the live tree, and
		// documented as unmeasurable so the next reader does not write a fixture that asserts nothing.
		if len(entities) == 0 {
			return nil
		}

		sourceText := ctx.SourceFile.Text()

		return rule.Listeners{
			ast.KindJsxText: func(node *ast.Node) {
				reportEntities(ctx, sourceText, node, entities)
			},
		}
	},
}

// reportEntities scans one JsxText node and reports every forbidden character it holds.
//
// # The loop nesting is deliberately NOT upstream's, and this is the one place that matters
//
// Upstream puts the entity list outside and the text positions inside, so it EMITS a line holding
// `'` before `>` as `>` first. That order is invisible in ESLint, which sorts diagnostics by
// position before anyone sees them: measured by reversing the `forbid` array on the same input and
// watching the reported order not move, twice.
//
// So upstream's emission order is an artifact rather than a decision, and reproducing it here would
// be reproducing a workaround for a constraint we do not have. `rule_testing.ExpectFindings` asserts
// emission order directly, and the fix engine reasons over proposals in the order they arrive, so
// out-of-order emission is a real difference here where it was none there. Upstream's corpus agrees
// with source order too: its three-finding case `Multiple errors: '>> default parser` lists its
// errors as `'`, `>`, `>`, which is source order and is what the loops below produce.
//
// The first draft of this function had the loops the other way with a doc comment asserting that
// upstream's order was observable. It is not. The claim was reasoned from the source and was wrong,
// and it took reversing the option array against the installed rule to find out.
func reportEntities(ctx rule.Context, sourceText string, node *ast.Node, entities []forbiddenEntity) {
	nodeRange := core.NewTextRange(node.Pos(), node.End())
	raw := sourceText[nodeRange.Pos():nodeRange.End()]

	for index := 0; index < len(raw); index++ {
		for _, entity := range entities {
			// A multi-character entry can never match, because upstream compares one character
			// against the whole string. Skipping is what makes that explicit; without it the byte
			// comparison below would silently match the first byte and report, which is MORE than
			// upstream does.
			if len(entity.character) != 1 || raw[index] != entity.character[0] {
				continue
			}

			characterRange := core.NewTextRange(nodeRange.Pos()+index, nodeRange.Pos()+index+1)

			if !entity.hasAlternatives {
				ctx.ReportRange(characterRange, rule.Message{
					Id: messageUnescapedEntity.Id,
					Description: fmt.Sprintf(
						"HTML entity `%s` is written raw in JSX text and must be escaped.",
						entity.character),
				})
				continue
			}

			ctx.ReportRangeWithSuggestions(
				characterRange,
				rule.Message{
					Id: messageUnescapedEntityAlts.Id,
					Description: fmt.Sprintf(
						"`%s` can be escaped with %s.",
						entity.character, quotedAlternatives(entity.alternatives)),
				},
				alternativeSuggestions(nodeRange, raw, index, entity.alternatives)...,
			)
		}
	}
}

// alternativeSuggestions builds one suggestion per alternative escape.
//
// Each replaces the whole JsxText node rather than the single character, which is upstream's own
// edit shape and is preserved for the overlap reason spelled out on the rule. The text written is
// the node's raw source with the one byte at `index` swapped for the escape, which is exactly what
// upstream's `line.slice(0, index) + alt + line.slice(index + 1)` produces once the per-line
// bookkeeping is unwound: upstream slices `node.raw` on the same line the finding landed on, at the
// same index, and the index it uses is already relative to the sliced line whose first line is
// sliced at the same start column the raw text begins at.
func alternativeSuggestions(nodeRange core.TextRange, raw string, index int, alternatives []string) []rule.Suggestion {
	suggestions := make([]rule.Suggestion, 0, len(alternatives))
	for _, alternative := range alternatives {
		suggestions = append(suggestions, rule.Suggestion{
			Message: rule.Message{
				Id:          messageReplaceWithAlt.Id,
				Description: fmt.Sprintf("Replace with `%s`.", alternative),
			},
			Fixes: []rule.Fix{
				rule.ReplaceRange(nodeRange, raw[:index]+alternative+raw[index+1:]),
			},
		})
	}
	return suggestions
}

// quotedAlternatives renders the alternatives the way upstream's `alts` placeholder does.
//
// Backticked and comma-joined: `&apos;`, `&lsquo;`, `&#39;`, `&rsquo;`. An empty list renders as
// the empty string, which is why upstream's `{char: "&", alternatives: []}` message reads
// "`&` can be escaped with ." with the space before the period. Reproduced rather than tidied,
// because the message text is part of what a port is fidelity to and a reader comparing the two
// tools on the same input would otherwise see a difference this rule chose.
func quotedAlternatives(alternatives []string) string {
	quoted := make([]string, 0, len(alternatives))
	for _, alternative := range alternatives {
		quoted = append(quoted, "`"+alternative+"`")
	}
	return strings.Join(quoted, ", ")
}

// configuredEntities resolves the option value into the list the scan walks.
//
// Nil options mean the rule was configured as a bare `"error"`, which the config layer turns into
// nil for a rule whose options are not required. That path has to land on the defaults, and it is
// the path the real tree takes for every file, so it is handled first and pinned by its own fixture.
// A rule reaching the scan with an empty list here would register on every file and find nothing,
// which is the failure that looks exactly like success.
func configuredEntities(options any) []forbiddenEntity {
	decoded, isOurs := rule.OptionsAs[NoUnescapedEntitiesOptions](options)
	if !isOurs || decoded.Forbid == nil {
		return defaultForbiddenEntities
	}
	return *decoded.Forbid
}

// noUnescapedEntitiesWireEntry is one `forbid` array element as it arrives over the wire.
//
// Upstream's schema is an `anyOf` over a string and an object, so an element is one or the other
// and the JSON decoder cannot be told that with a struct. It is decoded as `json.RawMessage` and
// dispatched on its first byte below.
type noUnescapedEntitiesWireEntry struct {
	Char         *string   `json:"char"`
	Alternatives *[]string `json:"alternatives"`
}

type noUnescapedEntitiesWireOptions struct {
	Forbid *[]json.RawMessage `json:"forbid"`
}

// DecodeNoUnescapedEntitiesOptions maps the wire options onto the struct the rule reads.
//
// Hand-written rather than `rule.DecodeOptionsInto` for two reasons that both come from the
// heterogeneous array. A `forbid` element is a string OR an object, which no single Go type
// unmarshals, and `forbid` itself has to be distinguishable as absent, because absent means the
// four defaults while `[]` means silence.
//
// An object entry with no `char` is dropped, matching upstream, which compares
// `c === entities[j].char` against undefined and never matches. Measured clean on
// `{forbid: [{alternatives: ["x"]}]}`.
//
// An object entry with no `alternatives` is ALSO dropped, and that one is a deliberate divergence:
// upstream crashes on it. See the rule doc.
func DecodeNoUnescapedEntitiesOptions(raw []byte) (any, error) {
	if len(raw) == 0 {
		return NoUnescapedEntitiesOptions{}, nil
	}

	var wire noUnescapedEntitiesWireOptions
	if err := rule.UnmarshalOptions(raw, &wire); err != nil {
		return NoUnescapedEntitiesOptions{}, err
	}
	if wire.Forbid == nil {
		return NoUnescapedEntitiesOptions{}, nil
	}

	entities := make([]forbiddenEntity, 0, len(*wire.Forbid))
	for _, element := range *wire.Forbid {
		entity, usable := decodeForbiddenEntity(element)
		if !usable {
			continue
		}
		entities = append(entities, entity)
	}

	return NoUnescapedEntitiesOptions{Forbid: &entities}, nil
}

// decodeForbiddenEntity reads one `forbid` array element in either of its two spellings.
func decodeForbiddenEntity(element json.RawMessage) (forbiddenEntity, bool) {
	var asString string
	if err := json.Unmarshal(element, &asString); err == nil {
		// The bare-string spelling: the character is forbidden and nothing is offered. Upstream
		// reports `unescapedEntity` with no suggestions for it, which is a different message from
		// the object spelling even when the object carries no alternatives.
		return forbiddenEntity{character: asString}, true
	}

	// Lenient on purpose: eslint-plugin-react 7.37.5's schema leaves a forbid entry object open (no
	// `additionalProperties`), so ESLint loads `{"char": ">", "alternatives": [], "extra": 1}` and so
	// must this.
	var asObject noUnescapedEntitiesWireEntry
	if err := json.Unmarshal(element, &asObject); err != nil {
		return forbiddenEntity{}, false
	}
	if asObject.Char == nil || asObject.Alternatives == nil {
		return forbiddenEntity{}, false
	}

	return forbiddenEntity{
		character:       *asObject.Char,
		alternatives:    *asObject.Alternatives,
		hasAlternatives: true,
	}, true
}
