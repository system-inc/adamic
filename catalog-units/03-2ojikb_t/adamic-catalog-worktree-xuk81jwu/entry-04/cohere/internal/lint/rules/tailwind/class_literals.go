// Package tailwind holds the rules that read Tailwind class names out of source.
//
// These are the eight rules from `eslint-plugin-better-tailwindcss` plus our own, and they are the
// one namespace in the catalog that nothing external covers: rslint ports the React and core rules,
// and zero of these. They are hand-ports.
//
// They are also the rules with no compiler behind them. A wrong prop is a type error and a wrong
// import is a type error, but a wrong class name is a string that renders slightly wrong on one
// breakpoint and goes unnoticed for a month. Lint is the only verification layer the entire visual
// surface has.
package tailwind

import (
	"strings"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	esregexp "github.com/system-inc/cohere/internal/lint/ecmascript/regexp"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// ClassLiteral is one string in the source that holds Tailwind class names.
//
// Range covers the string's contents without its quotes, so a fix can rewrite the classes without
// having to reason about which quote character the author used.
type ClassLiteral struct {
	Node  *ast.Node
	Text  string
	Range core.TextRange
	// Origin says which reading rule found this literal, so a report can explain itself and a test
	// can assert that all three surfaces are covered rather than only the obvious one.
	Origin ClassLiteralOrigin
	// Edges says which of the literal's ends will touch class text once it is substituted into a
	// template's hole, so that the whitespace at that end is all that separates two classes. See
	// holeEdges.
	Edges classValueEdges
}

// ClassLiteralOrigin is where a class string was written.
type ClassLiteralOrigin string

const (
	// ClassLiteralOriginAttribute is `className="..."` or `class="..."`.
	ClassLiteralOriginAttribute ClassLiteralOrigin = "Attribute"
	// ClassLiteralOriginCallee is a string argument to a known class-combining function, such as
	// `mergeClassNames('...')`.
	ClassLiteralOriginCallee ClassLiteralOrigin = "Callee"
	// ClassLiteralOriginVariable is a string assigned to a variable whose name says it holds
	// classes, such as `const buttonClassName = '...'`.
	ClassLiteralOriginVariable ClassLiteralOrigin = "Variable"
)

// ClassLiteralSettings names the three surfaces where class strings are written, each as upstream's
// name patterns.
//
// All three matter, and that is not obvious. On the ahra tree the attribute surface holds 7,773
// literals and the other two hold 2,192, so a rule that reads only JSX attributes silently covers
// 78% of the class surface while reporting a clean tree for the rest. Every rule in this package
// reads through this type for that reason.
//
// Every pattern is a JavaScript regular expression that must match the whole name, upstream's
// matchesName (utils/utils.js at 4.7.0): its first match, and that match is the name. An attribute
// pattern and name are both lowercased first, as getLiteralsByJSXAttribute does.
type ClassLiteralSettings struct {
	// AttributePatterns match the JSX attributes that carry classes, `class` and `className` by default.
	AttributePatterns []string
	// CalleeNamePatterns match a call's name, the identifier or the property a member call ends in,
	// and CalleePathPatterns match its dotted path, `twc.div` for `twc.div(...)`. A callee matching
	// either is read, which is upstream's selector `name` and `path`.
	CalleeNamePatterns []string
	CalleePathPatterns []string
	// VariablePatterns match variable names whose string initializers are classes.
	VariablePatterns []string
}

// DefaultClassLiteralSettings is upstream 4.7.0's default selectors (options/default-options.js),
// the half of them that read strings: the attribute, callee and variable selectors whose matchers
// read strings, or that have none.
//
// The other half is not ported yet and is named here so no one reads this as complete (#gj5nm6e,
// unit 6): object keys (`cn({ 'p-2': on })`, `classList`, `objstr`), object values at a path (cva and
// tv `variants`, `compoundVariants`, `slots`, `base`, and clb), the strings a function returns
// (twc and twx), and tagged templates (twc`...`). Nor is upstream's string matcher's walk ported
// exactly: these rules read the strings in a value position of an argument, and upstream reads every
// string nested under it.
//
// Defaults rather than a required option, because a rule that declines every file when unconfigured
// is indistinguishable from a rule with nothing to report. A project writing its own `attributes`,
// `callees` or `variables`, in a rule's options or in settings["better-tailwindcss"], replaces that
// kind's defaults whole, as upstream's legacy selectors do. Our own repositories name
// `mergeClassNames` and `createVariantClassNames` there.
func DefaultClassLiteralSettings() ClassLiteralSettings {
	return ClassLiteralSettings{
		AttributePatterns: []string{
			`^class(?:Name)?$`,
			`^class:.*$`,
			`(?:^\[class\]$)|(?:^\[ngClass\]$)`,
			`(?:^\[class\..*\]$)`,
			`^v-bind:class$`,
			`^class:list$`,
		},
		CalleeNamePatterns: []string{
			`^cc$`, `^clsx$`, `^cn$`, `^cnb$`, `^ctl$`, `^cva$`, `^cx$`, `^dcnb$`, `^tv$`, `^twJoin$`, `^twMerge$`,
		},
		CalleePathPatterns: []string{`^twc\.\w+`, `^twx\.\w+`},
		VariablePatterns:   []string{`^classNames?$`, `^classes$`, `^styles?$`},
	}
}

// TailwindClassLiteralOptions are upstream's legacy selector options, which every better-tailwindcss
// rule accepts: `attributes`, `callees` and `variables`, each a list of name patterns.
//
// A kind that is written replaces that kind's defaults whole, an empty list included, which then reads
// nothing of that kind; a kind not written keeps its defaults. That is upstream's createRule: a legacy
// kind drops every default selector of the kind (`hasCalleeOverride` and its siblings). A legacy callee
// is upstream's `{name, path}` with both set to the pattern, so it matches a call by either.
//
// Not ported yet: a legacy entry's matcher form, `[name, [{match}]]`, and `tags`, both unit 6 of
// #gj5nm6e, refused until then.
type TailwindClassLiteralOptions struct {
	Attributes []string `json:"attributes"`
	Callees    []string `json:"callees"`
	Variables  []string `json:"variables"`

	// surfaces is these options compiled, set once by the decoder, so a rule asking on every file reads
	// a pointer rather than rebuilding the settings and their key. Nil on options a test builds by hand,
	// which compile on each ask instead.
	surfaces *ClassLiteralSurfaces
}

// ClassLiteralSurfaces is these options compiled: the one shared by every decode that merged to the
// same settings.
func (options TailwindClassLiteralOptions) ClassLiteralSurfaces() *ClassLiteralSurfaces {
	if options.surfaces != nil {
		return options.surfaces
	}
	return ClassLiteralSurfacesFor(options.ClassLiteralSettings())
}

// compileClassLiterals compiles the options once, after the decoder has merged the settings into them.
// Promoted onto every options struct that embeds TailwindClassLiteralOptions, so the decoder can call it.
func (options *TailwindClassLiteralOptions) compileClassLiterals() {
	options.surfaces = ClassLiteralSurfacesFor(options.ClassLiteralSettings())
}

// ClassLiteralSettings is the defaults with every kind these options write put in place of its own.
func (options TailwindClassLiteralOptions) ClassLiteralSettings() ClassLiteralSettings {
	settings := DefaultClassLiteralSettings()
	if options.Attributes != nil {
		settings.AttributePatterns = options.Attributes
	}
	if options.Callees != nil {
		settings.CalleeNamePatterns = options.Callees
		settings.CalleePathPatterns = options.Callees
	}
	if options.Variables != nil {
		settings.VariablePatterns = options.Variables
	}
	return settings
}

// ClassLiteralReader finds class-carrying strings in a file.
//
// Rules get one from ClassLiteralSurfaces.ReaderFor, whose surfaces compile each distinct set of
// settings once per run and which hands every rule reading a file with those settings the same reader,
// so a node is read once however many rules listen to it. NewClassLiteralReader builds an unshared one, for a harness.
type ClassLiteralReader struct {
	attributes  *namePatterns
	calleeNames *namePatterns
	calleePaths *namePatterns
	variables   *namePatterns

	// values holds what each node read as, for the one file this reader serves. Nil on a reader that
	// is not bound to a file, which then reads every node afresh.
	values map[*ast.Node]classValues
}

// namePatterns is one kind's patterns, compiled, with each name's answer remembered.
//
// The answers are a memo because names repeat and the patterns run on a backtracking engine: ahra's
// attributes are almost all `className`, asked about on every JSX attribute in 3,978 files. The answer
// is a pure function of the name, so the memo is shared by every worker reading with these settings.
type namePatterns struct {
	compiled  []*esregexp.RegExp
	lowercase bool
	answers   sync.Map
}

// newNamePatterns compiles patterns as JavaScript does. One that does not compile is skipped rather
// than fatal: a rule that refuses to run because one pattern in a config file is malformed reports a
// clean tree, which is the failure mode this whole tool exists to remove.
func newNamePatterns(patterns []string, lowercase bool) *namePatterns {
	compiledPatterns := &namePatterns{lowercase: lowercase}
	for _, pattern := range patterns {
		if lowercase {
			pattern = strings.ToLower(pattern)
		}
		compiled, err := esregexp.Compile(pattern, "")
		if err != nil {
			continue
		}
		compiledPatterns.compiled = append(compiledPatterns.compiled, compiled)
	}
	return compiledPatterns
}

// matches reports whether any pattern matches the whole name, upstream's matchesName.
func (patterns *namePatterns) matches(name string) bool {
	if name == "" || len(patterns.compiled) == 0 {
		return false
	}
	if patterns.lowercase {
		name = strings.ToLower(name)
	}
	if answer, isAnswered := patterns.answers.Load(name); isAnswered {
		return answer.(bool)
	}
	answer := false
	for _, pattern := range patterns.compiled {
		// The first match, and only if it is the name: `exec(name)[0] === name`. A pattern whose
		// first match is shorter does not match, though a longer match exists, exactly as upstream.
		match, err := pattern.Unwrap().FindStringMatch(name)
		if err == nil && match != nil && match.RuneIndex == 0 && match.String() == name {
			answer = true
			break
		}
	}
	patterns.answers.Store(name, answer)
	return answer
}

// ClassLiteralSurfaces is one set of settings compiled for the run: the unbound reader each file binds,
// and the file-cache key it binds under.
//
// The settings depend on the configuration and never on the file, yet every rule used to build them
// on every file it read, the defaults and then their key: on an ahra cold run 17 MB in 193K objects
// for the defaults and 24 MB in 263K for the keys (#y2nj5ex). Now a rule's decoder compiles them once
// per configuration selection, and a rule reads a pointer.
type ClassLiteralSurfaces struct {
	compiled *ClassLiteralReader
	cacheKey string
}

// compiledClassLiteralSurfaces holds the surfaces for each distinct settings key, for the life of the
// process. Keyed by the settings' value, never by the options that asked, so every selection that
// merges to the same settings shares one entry and one name memo. Each is never written after it is
// stored, so the walk's workers share them freely, and the key is the whole input, so an entry cannot
// go stale.
var compiledClassLiteralSurfaces sync.Map

// ClassLiteralSurfacesFor returns the compiled surfaces for these settings, compiling them on the first
// ask for their value.
func ClassLiteralSurfacesFor(settings ClassLiteralSettings) *ClassLiteralSurfaces {
	key := settings.key()
	if existing, isCompiled := compiledClassLiteralSurfaces.Load(key); isCompiled {
		return existing.(*ClassLiteralSurfaces)
	}
	compiled, _ := compiledClassLiteralSurfaces.LoadOrStore(key, &ClassLiteralSurfaces{
		compiled: NewClassLiteralReader(settings),
		cacheKey: "tailwind.classValues:" + key,
	})
	return compiled.(*ClassLiteralSurfaces)
}

// DefaultClassLiteralSurfaces is DefaultClassLiteralSettings compiled, once for the process, for a rule
// handed no options.
var DefaultClassLiteralSurfaces = sync.OnceValue(func() *ClassLiteralSurfaces {
	return ClassLiteralSurfacesFor(DefaultClassLiteralSettings())
})

// ReaderFor returns the reader for these surfaces in this file, shared through the file's cache by
// every rule that reads with the same settings.
//
// Thirteen rules listen on the same three kinds, so before this each class surface was read thirteen
// times per file, and each rule compiled its own copy of the variable patterns on every file:
// 13 × 3,978 files × 2 patterns in ahra, about 0.08s of CPU spent recompiling two regular
// expressions. Reading is a pure function of the node and the settings, so one reading serves them
// all. A rule configured with different surfaces gets its own reader, keyed apart.
//
// Under --timing the first rule to reach a node pays for reading it and the rest read the memo, so
// the family's total is the honest number and one rule's row carries the shared reading.
//
// A nil cache yields a reader bound to nothing shared, which still memoizes within the one rule.
func (surfaces *ClassLiteralSurfaces) ReaderFor(cache *rule.FileCache) *ClassLiteralReader {
	return rule.Cached(cache, surfaces.cacheKey, surfaces.bind)
}

// bind is a copy of the compiled reader with a memo of its own, for one file.
func (surfaces *ClassLiteralSurfaces) bind() *ClassLiteralReader {
	bound := *surfaces.compiled
	bound.values = map[*ast.Node]classValues{}
	return &bound
}

// key is the settings as one string, every list kept in order and every name kept apart, so two
// settings share a key only when they would read every node the same way.
func (s ClassLiteralSettings) key() string {
	var builder strings.Builder
	for index, names := range [][]string{s.AttributePatterns, s.CalleeNamePatterns, s.CalleePathPatterns, s.VariablePatterns} {
		if index > 0 {
			builder.WriteByte(1)
		}
		for _, name := range names {
			builder.WriteString(name)
			builder.WriteByte(0)
		}
	}
	return builder.String()
}

// NewClassLiteralReader compiles the settings into a reader.
func NewClassLiteralReader(settings ClassLiteralSettings) *ClassLiteralReader {
	return &ClassLiteralReader{
		attributes:  newNamePatterns(settings.AttributePatterns, true),
		calleeNames: newNamePatterns(settings.CalleeNamePatterns, false),
		calleePaths: newNamePatterns(settings.CalleePathPatterns, false),
		variables:   newNamePatterns(settings.VariablePatterns, false),
	}
}

// ListenerKinds are the node kinds a rule must subscribe to in order to see every class literal.
//
// Returned as a list so a rule registers the same set rather than each one remembering three kinds,
// and so adding a fourth surface later reaches every rule at once.
func ListenerKinds() []ast.Kind {
	return []ast.Kind{
		ast.KindJsxAttribute,
		ast.KindCallExpression,
		ast.KindVariableDeclaration,
	}
}

// ClassLiteralsIn returns the class strings a node carries, or nothing.
//
// Every string that can become the class value is returned, each as its own literal: the surface's
// own string, and the strings in the branches of a conditional, the operands of `&&`, `||` and `??`
// that can be the result, the elements of an array, and the expressions inside a template's holes.
// Prettier's Tailwind plugin sorted every one of these, and until #vf1hd6j measured it this reader
// returned only the first: 364 nested literals in ahra and 299 in www-phi-health that no rule here
// had ever read.
//
// Value positions only, which is narrower than the plugin. Its `sortInside` visits every string
// under the braces, so in `size === 'sm' ? 'p-2' : 'p-4'` it also reads `'sm'`. Sorting a single
// word is a no-op, so the plugin loses nothing by it; reporting `'primary-large'` from a comparison
// as an unknown class would be a false positive. A condition is not a class list.
//
// Each literal stands alone, so a rule comparing classes within one literal never compares the two
// branches of a conditional, which are never applied together. The static text of a template with
// holes is not a literal here: a class assembled at runtime is `no-concatenated-classes`'s finding,
// and `ClassSegmentsIn` and `ClassTemplateSegmentsIn` read that text for the rules that edit it.
func (r *ClassLiteralReader) ClassLiteralsIn(node *ast.Node) []ClassLiteral {
	return r.classValuesIn(node).literals
}

// classValues is everything one surface carries in value positions: its strings, and its templates
// with holes, each template with the position it was found in.
type classValues struct {
	literals  []ClassLiteral
	templates []classTemplateValue
}

// classTemplateValue is one template with holes found in a value position.
type classTemplateValue struct {
	node   *ast.Node
	origin ClassLiteralOrigin
	// edges says which of the template's outer ends touch class text, when it is itself inside
	// another template's hole.
	edges classValueEdges
}

// classValuesIn returns what a node carries, read once per file when the reader is bound to one.
//
// The slices in a remembered reading are shared by every rule that asks, so callers range over them
// and never write into them.
func (r *ClassLiteralReader) classValuesIn(node *ast.Node) classValues {
	if r.values == nil {
		return r.readClassValues(node)
	}
	if remembered, isRead := r.values[node]; isRead {
		return remembered
	}
	values := r.readClassValues(node)
	r.values[node] = values
	return values
}

// readClassValues dispatches on the three class surfaces.
func (r *ClassLiteralReader) readClassValues(node *ast.Node) classValues {
	if node == nil {
		return classValues{}
	}

	switch node.Kind {
	case ast.KindJsxAttribute:
		return r.attributeValues(node)

	case ast.KindCallExpression:
		return r.calleeValues(node)

	case ast.KindVariableDeclaration:
		return r.variableValues(node)
	}

	return classValues{}
}

func (r *ClassLiteralReader) attributeValues(node *ast.Node) classValues {
	attribute := node.AsJsxAttribute()
	if attribute == nil {
		return classValues{}
	}

	// A namespaced attribute reads as `namespace:name`, which is what upstream's getAttributeName
	// builds and what Text gives a JsxNamespacedName.
	name := attribute.Name()
	if name == nil || !r.attributes.matches(name.Text()) {
		return classValues{}
	}

	return classValuesUnder(attribute.Initializer, ClassLiteralOriginAttribute)
}

func (r *ClassLiteralReader) calleeValues(node *ast.Node) classValues {
	call := node.AsCallExpression()
	if call == nil || call.Expression == nil {
		return classValues{}
	}

	// Upstream's getESCalleeName: a call is named by its identifier or the property a member call
	// ends in, so `cn(...)` and `utils.cn(...)` are both `cn`, and its path is the dotted chain,
	// `twc.div`. A callee with neither is never read.
	if call.Arguments == nil {
		return classValues{}
	}
	if !r.readsCallee(call.Expression) {
		return classValues{}
	}

	values := classValues{}
	for _, argument := range call.Arguments.Nodes {
		collectClassValues(argument, ClassLiteralOriginCallee, classValueEdges{}, &values)
	}
	return values
}

func (r *ClassLiteralReader) variableValues(node *ast.Node) classValues {
	declaration := node.AsVariableDeclaration()
	if declaration == nil || declaration.Initializer == nil {
		return classValues{}
	}

	name := declaration.Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return classValues{}
	}

	if !r.variables.matches(name.Text()) {
		return classValues{}
	}

	return classValuesUnder(declaration.Initializer, ClassLiteralOriginVariable)
}

// readsCallee reports whether a call's callee is one of these settings', by name or by path. The path
// is built only when the name did not match and a path pattern could, since it is a new string for every
// member call in the file.
func (r *ClassLiteralReader) readsCallee(callee *ast.Node) bool {
	if r.calleeNames.matches(calleeName(callee)) {
		return true
	}
	if len(r.calleePaths.compiled) == 0 {
		return false
	}
	_, path := calleeNameAndPath(callee)
	return r.calleePaths.matches(path)
}

// calleeName is calleeNameAndPath's name alone, which allocates nothing.
func calleeName(callee *ast.Node) string {
	switch callee.Kind {
	case ast.KindIdentifier:
		return callee.Text()
	case ast.KindPropertyAccessExpression:
		access := callee.AsPropertyAccessExpression()
		if access.Expression == nil || access.Expression.Kind == ast.KindSuperKeyword {
			return ""
		}
		if access.Name() != nil && access.Name().Kind == ast.KindIdentifier {
			return access.Name().Text()
		}
	case ast.KindElementAccessExpression:
		access := callee.AsElementAccessExpression()
		if access.Expression == nil || access.Expression.Kind == ast.KindSuperKeyword {
			return ""
		}
		if access.ArgumentExpression != nil && access.ArgumentExpression.Kind == ast.KindStringLiteral {
			return access.ArgumentExpression.Text()
		}
	}
	return ""
}

// calleeNameAndPath is upstream's getESCalleeName (parsers/es.js at 4.7.0) for both of its readings:
// the name is an identifier's text or the last property of a member access, and the path is the whole
// chain joined by dots, present only when every link in it has a name. `this.cn` has the name `cn` and
// no path, and `super.cn` has neither. A computed access counts when its key is a plain string.
func calleeNameAndPath(callee *ast.Node) (string, string) {
	switch callee.Kind {
	case ast.KindIdentifier:
		return callee.Text(), callee.Text()

	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		var object *ast.Node
		property := ""
		if callee.Kind == ast.KindPropertyAccessExpression {
			access := callee.AsPropertyAccessExpression()
			object = access.Expression
			if access.Name() != nil && access.Name().Kind == ast.KindIdentifier {
				property = access.Name().Text()
			}
		} else {
			access := callee.AsElementAccessExpression()
			object = access.Expression
			if access.ArgumentExpression != nil && access.ArgumentExpression.Kind == ast.KindStringLiteral {
				property = access.ArgumentExpression.Text()
			}
		}
		if object == nil || object.Kind == ast.KindSuperKeyword || property == "" {
			return "", ""
		}
		_, objectPath := calleeNameAndPath(object)
		if objectPath == "" {
			return property, ""
		}
		return property, objectPath + "." + property
	}
	return "", ""
}

// classValuesUnder collects every string and template with holes in a value position under an
// expression.
func classValuesUnder(node *ast.Node, origin ClassLiteralOrigin) classValues {
	values := classValues{}
	collectClassValues(node, origin, classValueEdges{}, &values)
	return values
}

// collectClassValues walks the shapes a class value can arrive through, keeping the strings and the
// templates with holes.
//
// `&&` contributes only its right side, because its left side is the result only when it is falsy,
// and a class string is never falsy unless it is empty. `||` and `??` contribute both, since either
// side can be the result. Anything else, a call, a member access, a comparison, is not a class list
// and is not entered: in `cn(getSize('sm px-2'))` the string is an argument to a function nobody
// named as a class callee.
func collectClassValues(node *ast.Node, origin ClassLiteralOrigin, edges classValueEdges, values *classValues) {
	if node == nil {
		return
	}

	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		literal := classLiteralFrom(node, origin)
		literal.Edges = edges
		values.literals = append(values.literals, literal)

	case ast.KindJsxExpression:
		if expression := node.AsJsxExpression(); expression != nil {
			collectClassValues(expression.Expression, origin, edges, values)
		}

	// `className={('flex flex')}` is legal and means exactly what the unparenthesized form means.
	//
	// typescript-go keeps parentheses as real nodes rather than discarding them, so a reader that
	// unwraps only by the kinds it expects walks straight past them and finds nothing. That is a
	// silent under-report on valid code: the rule stays green, the tree looks clean, and the
	// finding never existed. All three rules in this package had it until a probe went looking.
	//
	// Recursive rather than a single unwrap, because `(('a'))` nests and the depth is the author's
	// choice. Stripping is correct here precisely because these rules read a string's contents and
	// parentheses cannot change them; it is not correct everywhere, and a rule whose verdict
	// depends on the parse shape must not copy this.
	case ast.KindParenthesizedExpression:
		if parenthesized := node.AsParenthesizedExpression(); parenthesized != nil {
			collectClassValues(parenthesized.Expression, origin, edges, values)
		}

	case ast.KindAsExpression:
		collectClassValues(node.AsAsExpression().Expression, origin, edges, values)

	case ast.KindSatisfiesExpression:
		collectClassValues(node.AsSatisfiesExpression().Expression, origin, edges, values)

	case ast.KindConditionalExpression:
		conditional := node.AsConditionalExpression()
		collectClassValues(conditional.WhenTrue, origin, edges, values)
		collectClassValues(conditional.WhenFalse, origin, edges, values)

	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary.OperatorToken == nil {
			return
		}
		switch binary.OperatorToken.Kind {
		case ast.KindAmpersandAmpersandToken:
			collectClassValues(binary.Right, origin, edges, values)
		case ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			collectClassValues(binary.Left, origin, edges, values)
			collectClassValues(binary.Right, origin, edges, values)
		}

	case ast.KindArrayLiteralExpression:
		if elements := node.AsArrayLiteralExpression().Elements; elements != nil {
			for _, element := range elements.Nodes {
				collectClassValues(element, origin, edges, values)
			}
		}

	case ast.KindTemplateExpression:
		values.templates = append(values.templates, classTemplateValue{node: node, origin: origin, edges: edges})
		template := node.AsTemplateExpression()
		if template.Head == nil || template.TemplateSpans == nil {
			return
		}
		spans := template.TemplateSpans.Nodes
		before := template.Head.Text()
		for index, spanNode := range spans {
			span := spanNode.AsTemplateSpan()
			if span == nil || span.Literal == nil {
				continue
			}
			after := span.Literal.Text()
			collectClassValues(span.Expression, origin, holeEdges(before, after, index == 0, index == len(spans)-1, edges), values)
			before = after
		}
	}
}

// classValueEdges says which ends of a value touch class text once it is substituted, so that the
// whitespace at that end is the only thing between two classes and must not be trimmed away.
//
// A value on a surface touches nothing at either end. A value inside a template's hole touches the
// text on each side of the hole unless that text supplies whitespace of its own.
type classValueEdges struct {
	Leading  bool
	Trailing bool
}

// holeEdges is which ends of a hole's value touch class text, from the template text either side.
//
// In `flex ${open ? ' hidden ' : ”}` the text before the hole ends with a space, so the value's
// leading space separates nothing and can go, and nothing follows the hole, so its trailing space
// can go too: the result is `'hidden'`, which is what Prettier's Tailwind plugin writes. In
// `flex${open ? ' hidden' : ”}` the leading space is the only thing keeping `hidden` from fusing
// with `flex`, so it stays.
//
// The plugin trims there too, and that is a defect in it rather than a convention to copy: its
// `canCollapseWhitespaceIn` tests whether the quasi before a hole starts with whitespace where the
// question is whether it ends with it, so it rewrites `flex${c ? '  block  ' : ”}` to
// `flex${c ? 'block' : ”}` and the two classes become `flexblock`. Measured 2026-10-02 against
// 0.8.1. This matches it everywhere the trim is safe, which is every hole string the gate found, and
// declines only the rewrites that change what renders.
//
// An empty run between two holes touches the neighbouring hole's value, which nothing here can
// read, so it counts as touching. An empty run at a template's own end inherits that template's
// edges, which is what makes a template nested inside another template's hole come out right.
func holeEdges(before string, after string, firstHole bool, lastHole bool, template classValueEdges) classValueEdges {
	edges := classValueEdges{}
	switch {
	case before == "" && firstHole:
		edges.Leading = template.Leading
	case before == "":
		edges.Leading = true
	default:
		edges.Leading = !isSpace(rune(before[len(before)-1]))
	}
	switch {
	case after == "" && lastHole:
		edges.Trailing = template.Trailing
	case after == "":
		edges.Trailing = true
	default:
		edges.Trailing = !isSpace(rune(after[0]))
	}
	return edges
}

// classLiteralFrom records the literal and the span of its contents.
//
// The range excludes the surrounding quotes so that a fix replaces class text and nothing else. Any
// literal whose source form is shorter than two characters cannot have quotes to strip, so it is
// reported with its own range rather than a negative one.
//
// The range starts at the literal's token, not at its Loc. Loc includes the whitespace before the
// literal, so for one on its own line or after a comma the quote strip cut the opening quote, the
// fixed file stopped parsing, and the edit engine refused that file's whole fix batch, every other
// rule's fixes included (#vf1hd6j, measured at 174 files in ahra).
func classLiteralFrom(literal *ast.Node, origin ClassLiteralOrigin) ClassLiteral {
	contentRange := rule.TokenRange(ast.GetSourceFileOfNode(literal), literal)
	if contentRange.End()-contentRange.Pos() >= 2 {
		contentRange = core.NewTextRange(contentRange.Pos()+1, contentRange.End()-1)
	}

	return ClassLiteral{
		Node:   literal,
		Text:   literal.Text(),
		Range:  contentRange,
		Origin: origin,
	}
}

// SplitClasses breaks a class string the way Tailwind reads it.
//
// Fragments containing `${` are dropped rather than treated as classes. A literal that survived to
// here with a hole in it is only partly known at lint time, and reporting on the visible half would
// produce findings the author cannot act on.
func SplitClasses(classString string) []string {
	var classes []string
	for _, field := range text.WhitespaceFields(classString) {
		if strings.Contains(field, "${") {
			continue
		}
		classes = append(classes, field)
	}
	return classes
}
