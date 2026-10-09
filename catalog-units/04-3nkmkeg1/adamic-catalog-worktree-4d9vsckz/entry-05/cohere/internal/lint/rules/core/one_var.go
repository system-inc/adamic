package core

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// OneVarMode is one of upstream's three settings for a declaration group.
//
// PascalCase literals, per this repository's convention for string-literal unions. The wire
// spellings upstream uses are lowercase and are mapped in the decoder, so the enum a reader sees at
// a call site and the JSON a config carries stay separate concerns.
type OneVarMode string

const (
	// OneVarAlways requires one declaration statement per scope for the group.
	OneVarAlways OneVarMode = "Always"
	// OneVarNever requires one declarator per statement.
	OneVarNever OneVarMode = "Never"
	// OneVarConsecutive allows several statements but merges ones that are adjacent.
	OneVarConsecutive OneVarMode = "Consecutive"
)

// oneVarGroup is upstream's per-declaration-kind pair of settings.
//
// Both fields are POINTERS, and that is load bearing rather than defensive. Upstream builds
// `options.var = { uninitialized: mode.var, initialized: mode.var }`, so an unconfigured kind gets
// `undefined` in both slots, and `if (!options[key]) return;` then skips that kind entirely. A
// value type cannot tell "not configured" from a mode, so `{ "var": "never" }` would silently
// govern `let` and `const` as well. Nil means absent, exactly as `undefined` does upstream.
type oneVarGroup struct {
	Uninitialized *OneVarMode
	Initialized   *OneVarMode
}

// configured answers upstream's `if (!options[key])`.
//
// Upstream tests the GROUP object for truthiness, and for the object form it always assigns one, so
// the group is falsy only when the key is not one of the five it knows. Here a group with neither
// half set is the same condition and is what the decoder produces for a kind the config never
// named.
func (group oneVarGroup) configured() bool {
	return group.Uninitialized != nil || group.Initialized != nil
}

// is answers whether a half is set to a particular mode, treating absent as "no".
//
// Every upstream comparison is `options[key].initialized === MODE_ALWAYS` against a value that may
// be `undefined`, and `undefined === "always"` is false. Folding that into one method keeps the
// three-way absent/equal/different logic in one place rather than at each of the fourteen sites
// that ask.
func modeIs(mode *OneVarMode, wanted OneVarMode) bool {
	return mode != nil && *mode == wanted
}

// OneVarSettings is the decoded option surface.
//
// Five declaration kinds because upstream carries five: `var`, `let`, `const`, `using` and
// `await using`. The last is spelled `awaitUsing` in the config, matching upstream's schema.
type OneVarSettings struct {
	Var        oneVarGroup
	Let        oneVarGroup
	Const      oneVarGroup
	Using      oneVarGroup
	AwaitUsing oneVarGroup

	// SeparateRequires reproduces upstream's `!!mode.separateRequires`, which only has meaning
	// alongside an `initialized: "always"` group.
	SeparateRequires bool
}

// oneVarObjectOptions is the wire shape of the object form.
//
// Every field is a POINTER for the same reason `oneVarGroup`'s are: upstream distinguishes an
// absent key from a present one with `Object.hasOwn`, and the two take different paths. A plain
// string field would decode an absent `initialized` to `""`, which matches no mode and reads as
// "configured to something invalid" rather than as "not configured".
//
// The two object arms of upstream's schema are decoded into ONE struct rather than two, because
// they are disjoint by `additionalProperties: false` and a config carrying keys from both is
// refused below. Decoding into one struct and then refusing the mixture keeps the refusal in a
// place that can explain itself.
type oneVarObjectOptions struct {
	SeparateRequires *bool   `json:"separateRequires"`
	Var              *string `json:"var"`
	Let              *string `json:"let"`
	Const            *string `json:"const"`
	Using            *string `json:"using"`
	AwaitUsing       *string `json:"awaitUsing"`
	Initialized      *string `json:"initialized"`
	Uninitialized    *string `json:"uninitialized"`
}

// oneVarModeFromWire maps upstream's lowercase spelling onto the enum.
//
// A value outside the three is an ERROR rather than a silent fallback. Upstream gets that refusal
// from `meta.schema`'s enum before the rule ever runs; there is no schema layer here, so a bad
// value would otherwise decode to a mode matching nothing and produce a rule that registers, runs,
// and enforces nothing. That is the inert-but-configured shape section 7c of the porting standard
// exists to refuse.
func oneVarModeFromWire(wire string) (OneVarMode, error) {
	switch wire {
	case "always":
		return OneVarAlways, nil
	case "never":
		return OneVarNever, nil
	case "consecutive":
		return OneVarConsecutive, nil
	}
	return "", fmt.Errorf("one-var takes \"always\", \"never\" or \"consecutive\", got %q", wire)
}

// DefaultOneVarSettings is what a bare `"error"` gets.
//
// Upstream's `defaultOptions: ["always"]`, which is the string form applied to all five kinds. A
// rule configured without options is handed nil by the options registry, and returning the default
// here rather than a zero value is what keeps such a rule from registering on every file while
// enforcing nothing.
func DefaultOneVarSettings() OneVarSettings {
	always := OneVarAlways
	group := oneVarGroup{Uninitialized: &always, Initialized: &always}
	return OneVarSettings{Var: group, Let: group, Const: group, Using: group, AwaitUsing: group}
}

// DecodeOneVarOptions reads the string or object form off the config.
//
// Hand rolled rather than routed through `rule.DecodeOptionsInto` because the wire shape is a
// UNION: upstream's single schema element is `oneOf` a string enum and two object shapes, and no
// single Go struct decodes all three.
//
// # The variadic hazard does not apply here, and it is worth saying why
//
// `internal/lint/configuration/configuration.go` keeps only `tuple[1]`, so a rule whose upstream
// option surface is VARIADIC loses everything after the first value -- the defect measured on
// `id-match`, whose flag object vanished while the rule went on looking configured. `one-var`'s
// schema declares exactly ONE element, so the single value arrives intact and there is no second
// spelling to accept. Checked against `meta.schema` in the clone rather than assumed.
//
// # Absent is not a mode, at three separate layers
//
// The object form's keys are optional and an absent one means "leave this kind unconfigured", which
// upstream expresses as `undefined` flowing into `options[key]` and being caught by `if
// (!options[key]) return;`. So absence survives as nil through the wire struct, through
// `oneVarGroup`, and into the comparisons. Collapsing any of the three to a zero value makes
// `{ "var": "never" }` govern `let` and `const` too, and no fixture written from the string form
// could see it.
func DecodeOneVarOptions(raw []byte) (any, error) {
	if len(raw) == 0 {
		return DefaultOneVarSettings(), nil
	}

	// The string form first, because it is the common one and because a JSON string cannot also
	// decode as an object, so trying it costs nothing and cannot misfire.
	var wire string
	if err := json.Unmarshal(raw, &wire); err == nil {
		mode, modeErr := oneVarModeFromWire(wire)
		if modeErr != nil {
			return DefaultOneVarSettings(), modeErr
		}
		group := oneVarGroup{Uninitialized: &mode, Initialized: &mode}
		return OneVarSettings{Var: group, Let: group, Const: group, Using: group, AwaitUsing: group}, nil
	}

	var object oneVarObjectOptions
	if err := rule.UnmarshalOptions(raw, &object); err != nil {
		return DefaultOneVarSettings(),
			fmt.Errorf("one-var takes a string or an object, %q: %w", string(raw), err)
	}

	// Upstream's two object arms are `additionalProperties: false`, so a config naming a
	// declaration kind AND an initialization state is rejected by the schema before the rule runs.
	// Reproduced as a refusal rather than by picking one arm, because either choice silently
	// enforces something the author did not write.
	namesKind := object.Var != nil || object.Let != nil || object.Const != nil ||
		object.Using != nil || object.AwaitUsing != nil
	namesState := object.Initialized != nil || object.Uninitialized != nil
	if namesKind && namesState {
		return DefaultOneVarSettings(), fmt.Errorf(
			"one-var takes either declaration kinds (var/let/const/using/awaitUsing) or "+
				"initialization states (initialized/uninitialized), not both: %q", string(raw))
	}

	settings := OneVarSettings{}
	if object.SeparateRequires != nil {
		settings.SeparateRequires = *object.SeparateRequires
	}

	// Each named kind sets BOTH halves of its group, which is upstream's
	// `options.var = { uninitialized: mode.var, initialized: mode.var }`.
	assignKind := func(wire *string, group *oneVarGroup) error {
		if wire == nil {
			return nil
		}
		mode, err := oneVarModeFromWire(*wire)
		if err != nil {
			return err
		}
		group.Uninitialized = &mode
		group.Initialized = &mode
		return nil
	}
	for _, pair := range []struct {
		wire  *string
		group *oneVarGroup
	}{
		{object.Var, &settings.Var},
		{object.Let, &settings.Let},
		{object.Const, &settings.Const},
		{object.Using, &settings.Using},
		{object.AwaitUsing, &settings.AwaitUsing},
	} {
		if err := assignKind(pair.wire, pair.group); err != nil {
			return DefaultOneVarSettings(), err
		}
	}

	// The state form sets one HALF across all five kinds, which is upstream's pair of
	// `Object.hasOwn` blocks. Note it overwrites rather than merges, matching the order upstream
	// applies them in -- though the mixture is refused above, so the two never meet in practice.
	if object.Uninitialized != nil {
		mode, err := oneVarModeFromWire(*object.Uninitialized)
		if err != nil {
			return DefaultOneVarSettings(), err
		}
		settings.Var.Uninitialized = &mode
		settings.Let.Uninitialized = &mode
		settings.Const.Uninitialized = &mode
		settings.Using.Uninitialized = &mode
		settings.AwaitUsing.Uninitialized = &mode
	}
	if object.Initialized != nil {
		mode, err := oneVarModeFromWire(*object.Initialized)
		if err != nil {
			return DefaultOneVarSettings(), err
		}
		settings.Var.Initialized = &mode
		settings.Let.Initialized = &mode
		settings.Const.Initialized = &mode
		settings.Using.Initialized = &mode
		settings.AwaitUsing.Initialized = &mode
	}

	return settings, nil
}

// oneVarUpstreamSentence maps a message id to the sentence upstream renders for it.
//
// Upstream's seven messages are four distinct sentences plus three that differ only in the
// interpolated type, and the four specific ones name the initialization state explicitly. Keeping
// the templates in one table rather than at the report site is what makes them checkable against
// upstream's `meta.messages` by eye.
//
// `%s` is upstream's `{{type}}`, which is the declaration spelling.
var oneVarUpstreamSentence = map[string]string{
	"combineUninitialized": "Combine this with the previous '%s' statement with uninitialized variables.",
	"combineInitialized":   "Combine this with the previous '%s' statement with initialized variables.",
	"splitUninitialized":   "Split uninitialized '%s' declarations into multiple statements.",
	"splitInitialized":     "Split initialized '%s' declarations into multiple statements.",
	"combine":              "Combine this with the previous '%s' statement.",
	"split":                "Split '%s' declarations into multiple statements.",
	"splitRequires":        "Split requires to be separated into a single block.",
}

// oneVarRender builds the message a finding carries.
//
// The upstream sentence comes first and this tree's explanation follows it, so the text a reader
// sees opens with exactly what ESLint says and the fixtures can assert that prefix.
//
// `splitRequires` interpolates nothing, which is why the table's value for it carries no verb.
func oneVarRender(message rule.Message, kind oneVarKind) rule.Message {
	template, found := oneVarUpstreamSentence[message.Id]
	if !found {
		return message
	}
	sentence := template
	if strings.Contains(template, "%s") {
		sentence = fmt.Sprintf(template, kind.text())
	}
	return rule.Message{Id: message.Id, Description: sentence + " " + message.Description}
}

var messageOneVarCombineUninitialized = rule.Message{
	Id: "combineUninitialized",
	Description: "This declares uninitialized variables in a second statement where the scope " +
		"already has one. Reading the declarations of a scope means finding every statement that " +
		"declares, and one list per scope makes that a single place to look.",
}

var messageOneVarCombineInitialized = rule.Message{
	Id: "combineInitialized",
	Description: "This declares initialized variables in a second statement where the scope " +
		"already has one. Reading the declarations of a scope means finding every statement that " +
		"declares, and one list per scope makes that a single place to look.",
}

var messageOneVarSplitUninitialized = rule.Message{
	Id: "splitUninitialized",
	Description: "This declares several uninitialized variables in one statement. One binding per " +
		"statement makes each declaration its own line to move, comment or delete, and keeps a " +
		"diff that touches one variable from touching the others.",
}

var messageOneVarSplitInitialized = rule.Message{
	Id: "splitInitialized",
	Description: "This declares several initialized variables in one statement. One binding per " +
		"statement makes each declaration its own line to move, comment or delete, and keeps a " +
		"diff that touches one variable from touching the others.",
}

var messageOneVarSplitRequires = rule.Message{
	Id: "splitRequires",
	Description: "This mixes require calls with other initializers in one declaration. Separating " +
		"the requires keeps the module's dependencies readable as a block rather than interleaved " +
		"with ordinary locals.",
}

var messageOneVarCombine = rule.Message{
	Id: "combine",
	Description: "This declares variables in a second statement where the scope already has one. " +
		"Reading the declarations of a scope means finding every statement that declares, and one " +
		"list per scope makes that a single place to look.",
}

var messageOneVarSplit = rule.Message{
	Id: "split",
	Description: "This declares several variables in one statement. One binding per statement " +
		"makes each declaration its own line to move, comment or delete, and keeps a diff that " +
		"touches one variable from touching the others.",
}

// oneVarKind names the five declaration spellings upstream tracks separately.
type oneVarKind int

const (
	oneVarKindVar oneVarKind = iota
	oneVarKindLet
	oneVarKindConst
	oneVarKindUsing
	oneVarKindAwaitUsing
)

// text is the spelling the message interpolates, which is upstream's `node.kind`.
func (kind oneVarKind) text() string {
	switch kind {
	case oneVarKindLet:
		return "let"
	case oneVarKindConst:
		return "const"
	case oneVarKindUsing:
		return "using"
	case oneVarKindAwaitUsing:
		return "await using"
	}
	return "var"
}

// oneVarKindOf classifies a VariableDeclarationList by its flags.
//
// # The composite flag, which is the trap this function exists to avoid
//
// `NodeFlagsAwaitUsing` is NOT a distinct bit. `ast/nodeflags.go` defines it as
// `NodeFlagsConst | NodeFlagsUsing`, because those two never co-occur on a single spelling and the
// pair is free to mean the third thing. So `flags&NodeFlagsAwaitUsing != 0` is true for a plain
// `const` and for a plain `using` as well, and a test written that way classifies every `const` in
// the tree as `await using`. Measured directly with a probe rather than read off the constant's
// name:
//
//	var           Let=false Const=false Using=false
//	let           Let=true  Const=false Using=false
//	const         Let=false Const=true  Using=false   <- AwaitUsing mask is TRUE here
//	using         Let=false Const=false Using=true    <- and TRUE here
//	await using   Let=false Const=true  Using=true
//
// So the discriminator is both bits set together, tested before either alone. A mutant using the
// composite mask is caught by 89 assertions.
func oneVarKindOf(list *ast.Node) oneVarKind {
	flags := list.Flags
	isConst := flags&ast.NodeFlagsConst != 0
	isUsing := flags&ast.NodeFlagsUsing != 0
	switch {
	case isConst && isUsing:
		return oneVarKindAwaitUsing
	case isUsing:
		return oneVarKindUsing
	case isConst:
		return oneVarKindConst
	case flags&ast.NodeFlagsLet != 0:
		return oneVarKindLet
	}
	return oneVarKindVar
}

// oneVarCounts is upstream's `countDeclarations` result.
type oneVarCounts struct {
	Uninitialized int
	Initialized   int
}

// oneVarStateFlags is one scope's record of what it has already seen.
//
// Upstream keeps `{ initialized, uninitialized }` on a function frame and the same pair per
// declaration kind on a block frame, plus a `required` flag that only `separateRequires` sets.
type oneVarStateFlags struct {
	Initialized   bool
	Uninitialized bool
	Required      bool
}

// oneVarBlockFrame is upstream's `blockStack` entry.
//
// `var` is deliberately absent: it lives on the function frame, which is what makes a `var` in a
// nested block collide with one in the enclosing function while a `let` does not.
type oneVarBlockFrame struct {
	Let        oneVarStateFlags
	Const      oneVarStateFlags
	Using      oneVarStateFlags
	AwaitUsing oneVarStateFlags
}

// oneVarScopes carries the two stacks upstream maintains while walking.
type oneVarScopes struct {
	functions []oneVarStateFlags
	blocks    []oneVarBlockFrame
}

func (scopes *oneVarScopes) startBlock() {
	scopes.blocks = append(scopes.blocks, oneVarBlockFrame{})
}

func (scopes *oneVarScopes) endBlock() {
	if len(scopes.blocks) > 0 {
		scopes.blocks = scopes.blocks[:len(scopes.blocks)-1]
	}
}

func (scopes *oneVarScopes) startFunction() {
	scopes.functions = append(scopes.functions, oneVarStateFlags{})
	scopes.startBlock()
}

func (scopes *oneVarScopes) endFunction() {
	if len(scopes.functions) > 0 {
		scopes.functions = scopes.functions[:len(scopes.functions)-1]
	}
	scopes.endBlock()
}

// current returns a pointer to the state a declaration kind reads and writes.
//
// A POINTER, because upstream mutates the frame in place through an aliased reference and the
// recorded flags have to outlive the call. Returning a copy makes `recordTypes` write to nothing,
// which reads as a rule that never remembers a previous declaration -- every `always` case would
// then pass and only the corpus could tell. A mutant dropping the `var` arm is caught by 91 lines.
func (scopes *oneVarScopes) current(kind oneVarKind) *oneVarStateFlags {
	if kind == oneVarKindVar {
		if len(scopes.functions) == 0 {
			return nil
		}
		return &scopes.functions[len(scopes.functions)-1]
	}
	if len(scopes.blocks) == 0 {
		return nil
	}
	frame := &scopes.blocks[len(scopes.blocks)-1]
	switch kind {
	case oneVarKindLet:
		return &frame.Let
	case oneVarKindConst:
		return &frame.Const
	case oneVarKindUsing:
		return &frame.Using
	case oneVarKindAwaitUsing:
		return &frame.AwaitUsing
	}
	return nil
}

// oneVarIsFunctionScope answers which nodes upstream gives a `startFunction` listener.
//
// Upstream registers Program, FunctionDeclaration, FunctionExpression, ArrowFunctionExpression and
// StaticBlock. ESTree wraps every method, constructor and accessor body in a FunctionExpression, so
// that one entry covers all of them. Our parser has no such wrapper -- a method IS
// KindMethodDeclaration and is both the member and the function -- so the four member kinds are
// named here explicitly. Measured with a probe: `class C { m() {} }` produces
// KindMethodDeclaration directly under KindClassDeclaration, with no function node between.
//
// Omitting them puts a `var` inside a method on the enclosing frame, so two methods each declaring
// `var x` report the second as a duplicate. The whole 296-case corpus is blind to that -- it writes
// zero class methods, constructors, getters or setters -- which is why
// `TestOneVarStaysSilentOnTypeScriptShapesUpstreamCannotWrite` exists.
//
// A namespace body is deliberately NOT here. Measured: two sibling namespaces each declaring a
// `const` DO report upstream, because a TSModuleBlock is not a function scope.
func oneVarIsFunctionScope(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindSourceFile,
		ast.KindFunctionDeclaration,
		ast.KindFunctionExpression,
		ast.KindArrowFunction,
		ast.KindMethodDeclaration,
		ast.KindConstructor,
		ast.KindGetAccessor,
		ast.KindSetAccessor,
		ast.KindClassStaticBlockDeclaration:
		return true
	}
	return false
}

// oneVarIsBlockScope answers which nodes upstream gives a `startBlock` listener.
//
// Upstream's list is BlockStatement, ForStatement, ForInStatement, ForOfStatement and
// SwitchStatement.
//
// # The static block's inner Block is deliberately excluded
//
// Upstream's `StaticBlock` node is both the function scope AND the statement list, so it pushes one
// function frame and (via startFunction) one block frame. Our parser puts a KindBlock inside
// KindClassStaticBlockDeclaration, which would push a SECOND block frame for the same statements.
//
// Removing this guard is an EQUIVALENT mutant, and that is established structurally rather than
// inferred from its survival. A probe printing the children of a KindClassStaticBlockDeclaration
// shows it has exactly ONE child, its Block, and that every statement lives inside that Block. So
// there is no position where a statement sits in the outer frame but not the inner one: the two
// frames span identical statement sets and no verdict can differ between them.
//
// Kept regardless, because it is upstream's stack depth rather than ours, and a divergence chosen
// for being unobservable today is the kind that stops being unobservable quietly. Not kept as a
// claim that any fixture discriminates it -- none does, and the sweep says so.
func oneVarIsBlockScope(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindBlock:
		if node.Parent != nil && node.Parent.Kind == ast.KindClassStaticBlockDeclaration {
			return false
		}
		return true
	case ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement,
		ast.KindSwitchStatement:
		return true
	}
	return false
}

// oneVarStatementOf returns the node that plays the part of ESTree's VariableDeclaration.
//
// ESTree's VariableDeclaration is the STATEMENT and holds every declarator. Ours splits that into a
// VariableStatement wrapping a VariableDeclarationList, except in a for-head where the list stands
// alone. So the statement-shaped node is the list's parent when that parent is a VariableStatement,
// and the list itself otherwise.
//
// This is the node upstream reports on and the node it indexes within `parent.body`, so getting it
// wrong moves every span and breaks every consecutive verdict at once.
func oneVarStatementOf(list *ast.Node) *ast.Node {
	if list.Parent != nil && list.Parent.Kind == ast.KindVariableStatement {
		return list.Parent
	}
	return list
}

// oneVarStatementList returns the siblings upstream's `parent.body` would hold, or nil.
//
// Upstream's consecutive check reads
//
//	(parent.body && parent.body.length > 0 && parent.body.indexOf(node)) || 0
//
// so this answers "what does `parent.body` contain", which is narrower than "is this a statement
// list".
//
// # A switch clause is deliberately ABSENT, and it is the one container where the two disagree
//
// An ESTree SwitchCase keeps its statements in `.consequent`, not `.body`. So `parent.body` is
// undefined there, the whole index expression is 0, and the consecutive arm never fires inside a
// switch case at all. Measured against ESLint 10.10.0: two adjacent `let`s report `combine` in a
// block and report NOTHING in a `case` or a `default`, while the same clause still reports under
// `always` and `never`, both of which reach the scope stack rather than the sibling index.
//
// Ours calls that node a statement list, because it is one, so returning its statements here made
// the port report where upstream is silent. The corpus writes three switch cases and none of them
// is under `consecutive`, so nothing in it could see the divergence.
//
// `oneVarIsInStatementList` DOES include the clause, and that is not an inconsistency: it answers
// upstream's `isInStatementList`, which tests `astUtils.STATEMENT_LIST_PARENTS` -- a set that names
// SwitchCase explicitly. Two different questions, two different answers, one node.
//
// # Each kind gets its own arm, and that is not stylistic
//
// `As*()` accessors are interface conversions rather than casts, so calling `AsBlock()` on a
// KindSourceFile PANICS instead of answering nil. A shared `case KindA, KindB:` arm reaching for one
// accessor is exactly the shape that crashes, and the walk recovers per FILE rather than per rule,
// so one such panic costs every rule its verdict on that file.
func oneVarStatementList(container *ast.Node) []*ast.Node {
	if container == nil {
		return nil
	}
	switch container.Kind {
	case ast.KindSourceFile:
		return container.AsSourceFile().Statements.Nodes
	case ast.KindBlock:
		return container.AsBlock().Statements.Nodes
	case ast.KindModuleBlock:
		return container.AsModuleBlock().Statements.Nodes
	}
	return nil
}

// oneVarIsInStatementList answers upstream's `isInStatementList`, which gates the split fixer.
//
// Upstream tests whether the node's PARENT is one of Program, BlockStatement, StaticBlock or
// SwitchCase, which is exactly "this statement is a member of a statement list". `if (foo) var x, y;`
// fails it, and upstream declines to autofix there because splitting would move the second
// declaration outside the `if`.
//
// Upstream passes the ExportNamedDeclaration when there is one, because in ESTree that wrapper is
// the list member. Our parser has no such wrapper -- `export var a;` is a VariableStatement carrying
// an export modifier, sitting directly in the statement list -- so the statement node is already the
// right thing to test and the branch collapses.
//
// The switch clauses ARE here, unlike in `oneVarStatementList`. See that function's note.
func oneVarIsInStatementList(statement *ast.Node) bool {
	if statement == nil || statement.Parent == nil {
		return false
	}
	switch statement.Parent.Kind {
	case ast.KindSourceFile, ast.KindBlock, ast.KindModuleBlock,
		ast.KindCaseClause, ast.KindDefaultClause:
		return true
	}
	return false
}

// oneVarDeclarators returns the list's declarators.
func oneVarDeclarators(list *ast.Node) []*ast.Node {
	return list.AsVariableDeclarationList().Declarations.Nodes
}

// oneVarCountDeclarations is upstream's `countDeclarations`.
func oneVarCountDeclarations(declarators []*ast.Node) oneVarCounts {
	counts := oneVarCounts{}
	for _, declarator := range declarators {
		if declarator.AsVariableDeclaration().Initializer == nil {
			counts.Uninitialized++
		} else {
			counts.Initialized++
		}
	}
	return counts
}

// oneVarIsRequire is upstream's `isRequire`.
//
// Upstream reads `decl.init.callee.name === "require"`, which is a NAME test rather than a
// resolution: a locally declared `function require()` satisfies it, and so does a shadowed one.
// Reproduced as written, because the rule's judgment is the spelling. Measured against the
// installed rule, which reports on `function require(x) { return x; } var foo = require('foo'), bar;`.
//
// `callee.name` is `undefined` for any callee that is not an identifier -- a member access, a
// parenthesized expression -- and `undefined === "require"` is false, so those are not requires.
// Testing the kind first is what reproduces that without dereferencing something that has no name:
// measured, `a.require('foo')` is NOT a require upstream.
//
// Every one of the corpus's ten `separateRequires` cases calls `require` and nothing else, so a
// mutant answering true for every call survives all 296. See
// `TestOneVarRequireIsASpellingNotAResolution`.
func oneVarIsRequire(declarator *ast.Node) bool {
	initializer := declarator.AsVariableDeclaration().Initializer
	if initializer == nil || initializer.Kind != ast.KindCallExpression {
		return false
	}
	callee := initializer.AsCallExpression().Expression
	if callee == nil || callee.Kind != ast.KindIdentifier {
		return false
	}
	return callee.Text() == "require"
}

// oneVarAnyRequire and oneVarAllRequire are upstream's `.some(isRequire)` and `.every(isRequire)`.
//
// Note `every` on an EMPTY slice is true in JavaScript, and Go's loop form gives the same answer,
// so a list with no declarators is "all requires" in both. That only ever combines with `some`
// being false, so the pair reads as "no requires" either way.
func oneVarAnyRequire(declarators []*ast.Node) bool {
	for _, declarator := range declarators {
		if oneVarIsRequire(declarator) {
			return true
		}
	}
	return false
}

func oneVarAllRequire(declarators []*ast.Node) bool {
	for _, declarator := range declarators {
		if !oneVarIsRequire(declarator) {
			return false
		}
	}
	return true
}

// oneVarIndexOf finds a statement among its siblings by node identity.
//
// Upstream uses `parent.body.indexOf(node)`, which is identity in JavaScript too. Position would
// also work here and is the wrong instrument: two nodes can share a Pos when one is synthesized,
// and identity is what upstream actually asks.
func oneVarIndexOf(siblings []*ast.Node, statement *ast.Node) int {
	for index, sibling := range siblings {
		if sibling == statement {
			return index
		}
	}
	return -1
}

// oneVarDeclarationListOfStatement returns the declaration list a statement holds, or nil.
//
// Upstream's test is `previousNode.type === "VariableDeclaration"`, which in ESTree is the
// statement itself. Ours has to step through the VariableStatement to reach the list where the kind
// flags live.
func oneVarDeclarationListOfStatement(statement *ast.Node) *ast.Node {
	if statement == nil || statement.Kind != ast.KindVariableStatement {
		return nil
	}
	// An EXPORTED previous sibling is an ExportNamedDeclaration upstream, whose `kind` is
	// `undefined` rather than "const", so `previousNode.kind === type` is false and both the
	// consecutive check and the join repair decline against it. Ours has no wrapper, so the export
	// modifier is what stands in for that node's identity.
	//
	// Measured: `export const a = 1; const b = 2;` under "always" reports and returns `output:
	// null` upstream, while without this the port repaired it to `export const a = 1,  b = 2;` --
	// which moves a binding INTO an export and changes the module's public surface. The corpus
	// cannot reach it: its exported cases are all under "never".
	if oneVarIsExported(statement) {
		return nil
	}
	list := statement.AsVariableStatement().DeclarationList
	if list == nil || list.Kind != ast.KindVariableDeclarationList {
		return nil
	}
	return list
}

// oneVarIsForStatementInitializer answers upstream's `parent.type !== "ForStatement" || parent.init !== node`.
//
// Only the classic three-part `for` head is excluded. A `for...in` or `for...of` head is a different
// parent kind and is NOT excluded here, which is upstream's behaviour and is why the `always` arm
// carries its own separate for-in/for-of guard.
func oneVarIsForStatementInitializer(list *ast.Node) bool {
	parent := list.Parent
	if parent == nil || parent.Kind != ast.KindForStatement {
		return false
	}
	return parent.AsForStatement().Initializer == list
}

// oneVarIsForInOrOfLeft answers upstream's `node.parent.left === node && (ForIn || ForOf)`.
//
// This guard sits inside the `always`/uninitialized arm only, and it suppresses the report entirely
// rather than just the fix. `for (var x in y) {}` declares an uninitialized binding that cannot be
// merged with anything, so reporting it would be unactionable.
func oneVarIsForInOrOfLeft(list *ast.Node) bool {
	parent := list.Parent
	if parent == nil {
		return false
	}
	// Each kind gets its own arm: `As*()` is an interface conversion and an accessor named for one
	// kind panics on a node of another. These two happen to share `AsForInOrOfStatement`, which is
	// why they may safely resolve to the same call -- an accessor named for either one alone would
	// panic on the other, which is the shape that crashes a whole file.
	switch parent.Kind {
	case ast.KindForInStatement:
		return parent.AsForInOrOfStatement().Initializer == list
	case ast.KindForOfStatement:
		return parent.AsForInOrOfStatement().Initializer == list
	}
	return false
}

// oneVarIsExported answers whether a statement carries an `export` modifier.
//
// ESTree wraps an exported declaration in an ExportNamedDeclaration and upstream tests
// `declaration.parent.type`. Our parser has no wrapper: the modifier hangs off the statement, which
// is why this reads modifiers rather than the parent kind.
//
// # The wrapper changes what the rule DOES, not only how it is spelled
//
// Upstream reaches an exported declaration's siblings through `declaration.parent.parent.body` in
// the fixer and `parent.body` in the consecutive check. For an export, `declaration.parent` is the
// ExportNamedDeclaration, which has no `body` array, so both expressions yield nothing:
// `Array.isArray(...) ? ... : []` gives `[]` in the fixer, and `(parent.body && ...) || 0` gives 0
// in the check.
//
// So upstream neither reports a consecutive finding on an exported declaration nor repairs a
// combine on one. Ours puts the exported statement directly in the enclosing statement list, where
// it has real siblings and a real index, so without the collapse this rule reports where upstream
// is silent and -- far worse -- REPAIRS where upstream declines.
//
// Measured before it was fixed, driving both: upstream returns `output: null` for all four export
// shapes while this port produced
//
//	export const a = 1; export const b = 2;   ->   export const a = 1; export,  b = 2;
//
// which is not valid syntax. The corpus cannot see any of it: every one of its seven exported cases
// is under `never`, which reaches the SPLIT fixer and never the join. The split fixer very much
// does repair an export, which is why the collapse is scoped to the join.
func oneVarIsExported(statement *ast.Node) bool {
	modifiers := statement.Modifiers()
	if modifiers == nil {
		return false
	}
	for _, modifier := range modifiers.Nodes {
		if modifier.Kind == ast.KindExportKeyword {
			return true
		}
	}
	return false
}

// oneVarIsAmbient answers whether a statement carries a `declare` modifier.
//
// Read off the statement for the same reason `oneVarIsExported` is: our parser hangs the modifier
// there, beside `export`, rather than on the declaration list.
func oneVarIsAmbient(statement *ast.Node) bool {
	modifiers := statement.Modifiers()
	if modifiers == nil {
		return false
	}
	for _, modifier := range modifiers.Nodes {
		if modifier.Kind == ast.KindDeclareKeyword {
			return true
		}
	}
	return false
}

// oneVarChecker holds what one file's walk needs.
type oneVarChecker struct {
	context  rule.Context
	settings OneVarSettings
	scopes   oneVarScopes
}

// groupFor returns the settings governing a declaration kind.
func (checker *oneVarChecker) groupFor(kind oneVarKind) oneVarGroup {
	switch kind {
	case oneVarKindLet:
		return checker.settings.Let
	case oneVarKindConst:
		return checker.settings.Const
	case oneVarKindUsing:
		return checker.settings.Using
	case oneVarKindAwaitUsing:
		return checker.settings.AwaitUsing
	}
	return checker.settings.Var
}

// recordTypes is upstream's `recordTypes`, marking what this scope has now seen.
//
// Only an `always` half records. A `never` or `consecutive` half leaves the scope flags untouched,
// which is why a `consecutive` rule's verdict depends on the immediately preceding statement rather
// than on anything the scope accumulated.
func (checker *oneVarChecker) recordTypes(kind oneVarKind, declarators []*ast.Node,
	scope *oneVarStateFlags) {
	group := checker.groupFor(kind)
	for _, declarator := range declarators {
		if declarator.AsVariableDeclaration().Initializer == nil {
			if modeIs(group.Uninitialized, OneVarAlways) {
				scope.Uninitialized = true
			}
			continue
		}
		if !modeIs(group.Initialized, OneVarAlways) {
			continue
		}
		if checker.settings.SeparateRequires && oneVarIsRequire(declarator) {
			scope.Required = true
		} else {
			scope.Initialized = true
		}
	}
}

// hasOnlyOneStatement is upstream's function of the same name.
//
// It answers "may this statement declare, given what the scope has already seen", and it RECORDS on
// the way out when the answer is yes. The recording is the reason this cannot be split into a pure
// predicate: the third `var` in a scope is judged against a flag the second one set.
func (checker *oneVarChecker) hasOnlyOneStatement(kind oneVarKind, declarators []*ast.Node) bool {
	counts := oneVarCountDeclarations(declarators)
	group := checker.groupFor(kind)
	scope := checker.scopes.current(kind)
	if scope == nil {
		return true
	}
	hasRequires := oneVarAnyRequire(declarators)

	if modeIs(group.Uninitialized, OneVarAlways) && modeIs(group.Initialized, OneVarAlways) {
		if scope.Uninitialized || scope.Initialized {
			if !hasRequires {
				return false
			}
		}
	}
	if counts.Uninitialized > 0 {
		if modeIs(group.Uninitialized, OneVarAlways) && scope.Uninitialized {
			return false
		}
	}
	if counts.Initialized > 0 {
		if modeIs(group.Initialized, OneVarAlways) && scope.Initialized {
			if !hasRequires {
				return false
			}
		}
	}
	if scope.Required && hasRequires {
		return false
	}
	checker.recordTypes(kind, declarators, scope)
	return true
}

// checkDeclarationList is upstream's `checkVariableDeclaration`.
//
// The three blocks run in upstream's order and are not exclusive: one statement can report a
// `splitRequires` and a `combine` and a `split`, which is why 23 corpus cases expect more than one
// finding on a single input.
func (checker *oneVarChecker) checkDeclarationList(list *ast.Node) {
	kind := oneVarKindOf(list)
	group := checker.groupFor(kind)

	// Upstream's `if (!options[key]) return;`, which skips a declaration kind the config never
	// named. Removing this is an EQUIVALENT mutant today and the sweep says so, for a reason
	// established by enumeration rather than by the survival: every read of a group's two halves
	// goes through `modeIs`, which answers false for nil, so an unconfigured group already
	// declines at every one of the fourteen comparison sites.
	//
	// Kept because that redundancy is a property of `modeIs` rather than of this rule, and a
	// future change there would make this guard load bearing without anything saying so.
	// `TestOneVarSkipsDeclarationKindsTheConfigDidNotName` pins the behaviour directly, so the
	// contract survives even though no corpus row can discriminate the guard itself.
	if !group.configured() {
		return
	}

	statement := oneVarStatementOf(list)
	container := statement.Parent
	declarators := oneVarDeclarators(list)
	counts := oneVarCountDeclarations(declarators)
	mixedRequires := oneVarAnyRequire(declarators) && !oneVarAllRequire(declarators)

	if modeIs(group.Initialized, OneVarAlways) &&
		checker.settings.SeparateRequires && mixedRequires {
		checker.context.ReportNode(statement,
			oneVarRender(messageOneVarSplitRequires, kind))
	}

	// consecutive
	//
	// Upstream's index expression is `(parent.body && parent.body.length > 0 &&
	// parent.body.indexOf(node)) || 0`, whose `|| 0` makes a node that is FIRST in its list (index
	// 0) and a node with no list at all indistinguishable. Both mean "no previous sibling to merge
	// with", so the guard below is `> 0` and reproduces it exactly.
	siblings := oneVarStatementList(container)
	nodeIndex := oneVarIndexOf(siblings, statement)
	if oneVarIsExported(statement) {
		// Upstream's `parent` here is the ExportNamedDeclaration wrapping the declaration, and
		// that node carries no `body` array at all, so `parent.body && ...` is falsy and the whole
		// index expression collapses to 0. Both the consecutive check and the join repair are
		// therefore dead for an exported declaration upstream. See `oneVarIsExported`.
		siblings, nodeIndex = nil, 0
	}
	if nodeIndex > 0 {
		previous := siblings[nodeIndex-1]
		previousList := oneVarDeclarationListOfStatement(previous)
		if previousList != nil && oneVarKindOf(previousList) == kind {
			previousDeclarators := oneVarDeclarators(previousList)
			combined := append(append([]*ast.Node{}, declarators...), previousDeclarators...)
			if !(oneVarAnyRequire(combined) && !oneVarAllRequire(combined)) {
				previousCounts := oneVarCountDeclarations(previousDeclarators)
				checker.reportConsecutive(statement, list, kind, group, counts, previousCounts,
					siblings, nodeIndex)
			}
		}
	}

	// always
	if !checker.hasOnlyOneStatement(kind, declarators) {
		checker.reportAlways(statement, list, kind, group, counts, siblings, nodeIndex)
	}

	// never
	//
	// Upstream skips a `for` statement's own initializer, because splitting `for (var i = 0, j = 0;;)`
	// would move `j` outside the head and change what the loop means. `parent.init !== node` is the
	// test, and here the list IS the init when its parent is the for statement.
	if !oneVarIsForStatementInitializer(list) {
		if counts.Uninitialized+counts.Initialized > 1 {
			checker.reportNever(statement, list, kind, group, counts)
		}
	}
}

// reportConsecutive emits the combine family for the `consecutive` modes.
func (checker *oneVarChecker) reportConsecutive(statement *ast.Node, list *ast.Node,
	kind oneVarKind, group oneVarGroup, counts oneVarCounts, previousCounts oneVarCounts,
	siblings []*ast.Node, nodeIndex int) {
	switch {
	case modeIs(group.Initialized, OneVarConsecutive) &&
		modeIs(group.Uninitialized, OneVarConsecutive):
		checker.reportJoin(statement, list, messageOneVarCombine, kind, siblings, nodeIndex)
	case modeIs(group.Initialized, OneVarConsecutive) &&
		counts.Initialized > 0 && previousCounts.Initialized > 0:
		checker.reportJoin(statement, list, messageOneVarCombineInitialized, kind, siblings, nodeIndex)
	case modeIs(group.Uninitialized, OneVarConsecutive) &&
		counts.Uninitialized > 0 && previousCounts.Uninitialized > 0:
		checker.reportJoin(statement, list, messageOneVarCombineUninitialized, kind, siblings, nodeIndex)
	}
}

// reportAlways emits the combine family for the `always` modes.
func (checker *oneVarChecker) reportAlways(statement *ast.Node, list *ast.Node,
	kind oneVarKind, group oneVarGroup, counts oneVarCounts,
	siblings []*ast.Node, nodeIndex int) {
	if modeIs(group.Initialized, OneVarAlways) && modeIs(group.Uninitialized, OneVarAlways) {
		checker.reportJoin(statement, list, messageOneVarCombine, kind, siblings, nodeIndex)
		return
	}
	if modeIs(group.Initialized, OneVarAlways) && counts.Initialized > 0 {
		checker.reportJoin(statement, list, messageOneVarCombineInitialized, kind, siblings, nodeIndex)
	}
	if modeIs(group.Uninitialized, OneVarAlways) && counts.Uninitialized > 0 {
		// A for-in/for-of head is skipped entirely, report and all. See the helper's comment.
		if oneVarIsForInOrOfLeft(list) {
			return
		}
		checker.reportJoin(statement, list, messageOneVarCombineUninitialized, kind, siblings, nodeIndex)
	}
}

// reportNever emits the split family.
func (checker *oneVarChecker) reportNever(statement *ast.Node, list *ast.Node,
	kind oneVarKind, group oneVarGroup, counts oneVarCounts) {
	switch {
	case modeIs(group.Initialized, OneVarNever) && modeIs(group.Uninitialized, OneVarNever):
		checker.reportSplit(statement, list, messageOneVarSplit, kind)
	case modeIs(group.Initialized, OneVarNever) && counts.Initialized > 0:
		checker.reportSplit(statement, list, messageOneVarSplitInitialized, kind)
	case modeIs(group.Uninitialized, OneVarNever) && counts.Uninitialized > 0:
		checker.reportSplit(statement, list, messageOneVarSplitUninitialized, kind)
	}
}

// reportJoin reports a combine finding, with upstream's `joinDeclarations` repair when it applies.
//
// # Two edits that only mean something together
//
// Upstream's fixer is a GENERATOR yielding two or three edits: turn the previous statement's
// terminator into a comma, drop the `using` token when the kind is `await using`, and delete this
// statement's declaration keyword. They are disjoint spans, and `internal/edit`'s ProposalsFrom
// flattens a diagnostic's fixes into INDEPENDENT proposals, so the engine judges each separately.
//
// Measured against the real engine rather than assumed: both halves survive overlap resolution, and
// a rival fix spanning the gap between them LOSES to this pair rather than splitting it, because
// these sort first and claim their bytes.
//
// What that measurement does not buy is safety against the one-byte gap between the two spans. A
// rival fix landing strictly inside it would take one half and leave the other, and applying only
// the keyword deletion yields `var bar = true; baz = false;` -- which PARSES, and silently declares
// an implicit global. The engine's parse-failure guard cannot see that, so it is the one shape
// worth naming here. It has no live caller today: this is the first shipped rule to report more
// than one fix on one diagnostic, and `ProposalsFrom`'s own comment says grouping should be built
// when a rule genuinely needs it rather than warned about. That is a finding for the shelf and it
// is in the report, not something this rule can fix from here.
func (checker *oneVarChecker) reportJoin(statement *ast.Node, list *ast.Node,
	message rule.Message, kind oneVarKind, siblings []*ast.Node, nodeIndex int) {
	rendered := oneVarRender(message, kind)

	fixes := checker.joinFixes(statement, list, kind, siblings, nodeIndex)
	if len(fixes) == 0 {
		checker.context.ReportNode(statement, rendered)
		return
	}
	checker.context.ReportNodeWithFixes(statement, rendered, fixes...)
}

// joinFixes builds upstream's `joinDeclarationsFixer`, or nothing when it declines.
//
// Upstream's generator yields NOTHING unless the previous sibling is a declaration of the same
// kind. The consecutive arm has already established that, but the `always` arm calls the same fixer
// WITHOUT checking, so a combine reported on a statement whose predecessor is something else
// carries no repair at all. That is a report without a fix rather than a bug, and 26 corpus cases
// record it as `output: null`.
//
// The same-kind check here is therefore not redundant with the caller's. The isolating shape is
// `function f() { var a = 1; let b = 2; var c = 3; }`, where the third statement reports because
// the FUNCTION scope already holds a `var` while its immediate predecessor is a `let`. Measured:
// upstream returns `output: null`. No corpus case writes it, so a mutant dropping this check
// survives all 296; see `TestOneVarJoinDeclinesAgainstADifferentKindPredecessor`.
func (checker *oneVarChecker) joinFixes(statement *ast.Node, list *ast.Node,
	kind oneVarKind, siblings []*ast.Node, nodeIndex int) []rule.Fix {
	if nodeIndex <= 0 || nodeIndex-1 >= len(siblings) {
		return nil
	}
	previous := siblings[nodeIndex-1]
	previousList := oneVarDeclarationListOfStatement(previous)
	if previousList == nil || oneVarKindOf(previousList) != kind {
		return nil
	}

	// An AMBIENT declaration is declined, which is this port's own decision and a measured one.
	// `declare let a: number; declare let b: number;` was joined to `declare let a: number;
	// declare,  b: number;`, because the token before this statement's keyword is its own `declare`
	// rather than the previous statement's semicolon. That does not parse. Upstream's installed
	// build writes `declare let a: number,\n let b: number;`, which does not parse either.
	//
	// Repairing it properly would mean deleting the second `declare` too, and is only meaning-
	// preserving when BOTH statements are ambient: joining an ambient binding into a real one, or the
	// reverse, changes whether a variable exists at runtime. Reporting without a repair is the subset
	// that can be shown correct.
	if oneVarIsAmbient(statement) || oneVarIsAmbient(previous) {
		return nil
	}

	sourceFile := checker.context.SourceFile
	if sourceFile == nil {
		return nil
	}
	text := sourceFile.Text()

	// `type` is upstream's first token of the declaration -- the `var`/`let`/`const`/`using`
	// keyword, or `await` for an `await using`.
	//
	// Taken off the LIST rather than the statement. The two positions differ only for an exported
	// declaration, where the statement's first token is `export` and the list's is the keyword,
	// and the caller has already returned for those -- so a mutation swapping them SURVIVES, and
	// the sweep says so. The list is what upstream reads (its node begins at the keyword, with the
	// export wrapper above it), and reading the statement here would make this function's
	// correctness depend on a guard in another function rather than on anything local.
	keyword := scanner.GetRangeOfTokenAtPosition(sourceFile, list.Pos())

	// `beforeType` is the token before it, which is the previous statement's terminator when there
	// is one. Found by scanning back over trivia from the keyword rather than by taking the previous
	// statement's End(), because a statement's End() sits past its own terminator and a list's does
	// not, so the two disagree by exactly the byte this fix has to replace.
	beforeEnd := oneVarPreviousTokenEnd(text, keyword.Pos(), previous.Pos())
	if beforeEnd < 0 {
		return nil
	}

	fixes := []rule.Fix{}
	if beforeEnd > 0 && text[beforeEnd-1] == ';' {
		fixes = append(fixes, rule.ReplaceRange(core.NewTextRange(beforeEnd-1, beforeEnd), ","))
	} else {
		// Upstream's `insertTextAfter`, for a previous statement with no semicolon (ASI).
		fixes = append(fixes, rule.ReplaceRange(core.NewTextRange(beforeEnd, beforeEnd), ","))
	}

	// `await using` is two tokens, and upstream removes the second one separately before blanking
	// the first. Without this the repair leaves a stray `using` behind.
	if kind == oneVarKindAwaitUsing {
		usingToken := scanner.GetRangeOfTokenAtPosition(sourceFile, keyword.End())
		fixes = append(fixes, rule.RemoveRange(core.NewTextRange(usingToken.Pos(), usingToken.End())))
	}

	fixes = append(fixes, rule.RemoveRange(core.NewTextRange(keyword.Pos(), keyword.End())))
	return fixes
}

// oneVarPreviousTokenEnd scans back from a token start over trivia to the end of the token before it.
//
// Returns -1 when nothing but trivia lies between here and the floor, which means there is no
// previous token to attach a comma to.
//
// Scanning BACKWARD over a line comment is the case that makes this more than a whitespace skip: a
// `//` comment runs to the end of its line, so a backward scan that stopped at the first
// non-whitespace byte would stop on the comment's last character and treat it as a token. The loop
// therefore walks whole lines and re-examines each one for a comment start.
func oneVarPreviousTokenEnd(text string, from int, floor int) int {
	position := from
	for position > floor {
		// Skip whitespace directly before the cursor.
		for position > floor && (text[position-1] == ' ' || text[position-1] == '\t' ||
			text[position-1] == '\r' || text[position-1] == '\n') {
			position--
		}
		if position <= floor {
			return -1
		}
		// A block comment ending here.
		if position-2 >= floor && text[position-2] == '*' && text[position-1] == '/' {
			closing := position - 2
			opening := closing
			for opening-1 > floor && !(text[opening-1] == '*' && opening-2 >= floor &&
				text[opening-2] == '/') {
				opening--
			}
			if opening-2 >= floor && text[opening-2] == '/' {
				position = opening - 2
				continue
			}
			return position
		}
		// A line comment covering the cursor's line.
		lineStart := position
		for lineStart > floor && text[lineStart-1] != '\n' {
			lineStart--
		}
		commentAt := -1
		for scan := lineStart; scan+1 < position; scan++ {
			if text[scan] == '/' && text[scan+1] == '/' {
				commentAt = scan
				break
			}
		}
		if commentAt >= 0 {
			position = commentAt
			continue
		}
		return position
	}
	return -1
}

// reportSplit reports a split finding, with upstream's `splitDeclarations` repair when it applies.
func (checker *oneVarChecker) reportSplit(statement *ast.Node, list *ast.Node,
	message rule.Message, kind oneVarKind) {
	rendered := oneVarRender(message, kind)

	fixes := checker.splitFixes(statement, list, kind)
	if len(fixes) == 0 {
		checker.context.ReportNode(statement, rendered)
		return
	}
	checker.context.ReportNodeWithFixes(statement, rendered, fixes...)
}

// splitFixes builds upstream's `splitDeclarations`.
//
// Upstream returns null -- declining the repair entirely -- when the statement is not a member of a
// statement list, because splitting `if (foo) var x, y;` would move `y` outside the `if` and change
// what the code means. That is a decline of the FIX, not of the report.
//
// Each comma becomes `; <kind> `, with an `export ` prefix when the statement is exported, and the
// three shapes upstream distinguishes are about where the comment and the newline go.
//
// A `declare ` prefix is carried too, which upstream does not do. Measured on the installed build:
// `declare let a: number, b: number;` splits to `declare let a: number; let b: number;`, turning an
// ambient binding into one that exists at runtime, and `export declare const a: number, b: number;`
// splits to `export declare const a: number; export const b: number;`, which does not compile
// because a non-ambient `const` needs an initializer. Upstream's estree hangs `declare` on the
// declaration as a flag it never consults here; ours is a modifier on the statement, beside
// `export`, and both have to be repeated on every statement the split creates.
func (checker *oneVarChecker) splitFixes(statement *ast.Node, list *ast.Node,
	kind oneVarKind) []rule.Fix {
	if !oneVarIsInStatementList(statement) {
		return nil
	}
	sourceFile := checker.context.SourceFile
	if sourceFile == nil {
		return nil
	}
	text := sourceFile.Text()

	modifierPrefix := ""
	if oneVarIsExported(statement) {
		modifierPrefix = "export "
	}
	if oneVarIsAmbient(statement) {
		modifierPrefix += "declare "
	}

	fixes := []rule.Fix{}
	for _, declarator := range oneVarDeclarators(list) {
		comma := scanner.GetRangeOfTokenAtPosition(sourceFile, declarator.End())
		if comma.Pos() >= len(text) || text[comma.Pos()] != ',' {
			// Upstream returns null for a declarator with no token after it, and skips one whose
			// following token is not a comma -- the last declarator, whose next token is the
			// terminator.
			continue
		}

		// `afterComma` is the next token INCLUDING comments, which is what decides between the
		// three shapes below.
		afterCommaPos := oneVarNextTokenOrCommentStart(text, comma.End())
		if afterCommaPos < 0 {
			continue
		}

		// `var x,y` -- nothing at all between the comma and the next declarator.
		if afterCommaPos == comma.End() {
			fixes = append(fixes, rule.ReplaceRange(
				core.NewTextRange(comma.Pos(), comma.End()),
				"; "+modifierPrefix+kind.text()+" "))
			continue
		}

		// A newline, or a comment, between the comma and what follows. Upstream keeps everything
		// from the comma's end up to the last comment's start verbatim and puts the new statement
		// head after it, so `var f, /* c */ l` becomes `var f; /* c */ var l`.
		if oneVarCrossesLine(text, comma.End(), afterCommaPos) ||
			oneVarStartsComment(text, afterCommaPos) {
			lastCommentStart := oneVarLastCommentRunStart(text, comma.End())
			fixes = append(fixes, rule.ReplaceRange(
				core.NewTextRange(comma.Pos(), lastCommentStart),
				";"+text[comma.End():lastCommentStart]+modifierPrefix+kind.text()+" "))
			continue
		}

		// Ordinary whitespace on one line: `var x, y` -- the space after the comma is kept by the
		// replacement text ending without one.
		fixes = append(fixes, rule.ReplaceRange(
			core.NewTextRange(comma.Pos(), comma.End()),
			"; "+modifierPrefix+kind.text()))
	}
	return fixes
}

// oneVarNextTokenOrCommentStart returns where the next token OR comment begins, or -1.
func oneVarNextTokenOrCommentStart(text string, from int) int {
	position := from
	for position < len(text) {
		switch text[position] {
		case ' ', '\t', '\r', '\n':
			position++
			continue
		}
		return position
	}
	return -1
}

// oneVarCrossesLine answers whether a newline sits in a span.
func oneVarCrossesLine(text string, from int, to int) bool {
	for position := from; position < to && position < len(text); position++ {
		if text[position] == '\n' {
			return true
		}
	}
	return false
}

// oneVarStartsComment answers whether a position begins a line or block comment.
func oneVarStartsComment(text string, position int) bool {
	return position+1 < len(text) && text[position] == '/' &&
		(text[position+1] == '/' || text[position+1] == '*')
}

// oneVarLastCommentRunStart walks the run of comments after a comma and returns where the token
// following them begins.
//
// Upstream's loop advances `lastComment` while it is a Line or Block comment, so it lands on the
// first non-comment token; the replacement then runs up to that token's start. Reproduced by
// skipping whitespace and comments alternately.
func oneVarLastCommentRunStart(text string, from int) int {
	position := from
	for position < len(text) {
		for position < len(text) && (text[position] == ' ' || text[position] == '\t' ||
			text[position] == '\r' || text[position] == '\n') {
			position++
		}
		if position+1 < len(text) && text[position] == '/' && text[position+1] == '/' {
			for position < len(text) && text[position] != '\n' {
				position++
			}
			continue
		}
		if position+1 < len(text) && text[position] == '/' && text[position+1] == '*' {
			position += 2
			for position+1 < len(text) && !(text[position] == '*' && text[position+1] == '/') {
				position++
			}
			position += 2
			continue
		}
		return position
	}
	return position
}

// OneVar enforces that variables are declared together or separately, per configuration.
//
//	valid:   function foo() { var bar, baz; }
//	valid:   var foo = 1, bar = 2;
//	invalid: function foo() { var bar; var baz; }      // "always"
//	invalid: var foo = 1, bar = 2;                     // "never"
//	invalid: var a; var b;                             // "consecutive"
//
// Ported from `one-var` in ESLint 10.10.0, read from the clone at `lib/rules/one-var.js`. One
// option (a string enum or one of two object shapes), seven messages, and `meta.fixable` is
// `"code"`.
//
// All 296 of upstream's corpus cases were extracted from its own tester and replayed through the
// ESLint Linter API before any code was written -- twice, once under upstream's own parser settings
// and once under the TypeScript parser with `sourceType: module` that cohere always has. The two
// agree on every case, so nothing here rests on a JavaScript-only reading.
//
// # The walk is manual, because the rule is stateful and there is no exit listener
//
// Upstream registers ten `:exit` handlers to pop two stacks. This tree's listeners are pre-order
// only, so the whole traversal runs inside one `KindSourceFile` listener that pushes and pops the
// same frames around the same constructs. That is the shipped shape for a gather-then-judge rule.
//
// # Five places our AST differs from ESTree, and four of them move a verdict
//
// ESTree's `VariableDeclaration` is the statement AND holds the declarators. Ours splits it into a
// `VariableStatement` wrapping a `VariableDeclarationList`, and in a for-head the list stands alone
// with no statement above it. `oneVarStatementOf` picks whichever plays the statement's part, and
// getting it wrong moves every span and breaks every consecutive verdict at once.
//
// ESTree wraps an exported declaration in `ExportNamedDeclaration`, and that wrapper has no `body`
// array, so upstream neither reports a consecutive finding on an export nor repairs a combine on
// one. Ours has no wrapper at all. See `oneVarIsExported`, which carries the measurement and the
// invalid source the port produced before the collapse was added.
//
// An ESTree `SwitchCase` keeps its statements in `.consequent`, not `.body`, so upstream's
// consecutive arm never fires inside one. Ours calls that node a statement list, because it is one.
// See the note above `oneVarIsInStatementList`.
//
// ESTree wraps a method body in a `FunctionExpression`, so upstream's one listener covers methods,
// constructors and accessors. Ours has no wrapper -- a method IS `KindMethodDeclaration` -- so those
// four kinds are named explicitly as function scopes.
//
// The fifth is inert and recorded anyway: our `ClassStaticBlockDeclaration` holds a `Block` where
// upstream's `StaticBlock` is itself the statement list. See `oneVarIsBlockScope`.
//
// # `await using` is two flags, never the constant named for it
//
// `NodeFlagsAwaitUsing` is defined as `NodeFlagsConst | NodeFlagsUsing`, so a mask against it is
// true for every plain `const`. See `oneVarKindOf`, which tests both bits together instead and
// carries the measured flag table.
var OneVar = rule.Rule{
	Name: "one-var",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, ok := rule.OptionsAs[OneVarSettings](options)
		if !ok {
			// A rule configured as a bare `"error"` is handed nil, and `options.(T)` on nil yields
			// the zero value -- every group unconfigured, so the rule would register on every file
			// and enforce nothing. Upstream's `defaultOptions: ["always"]` is the answer.
			settings = DefaultOneVarSettings()
		}

		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				if ctx.SourceFile == nil {
					return
				}
				checker := &oneVarChecker{context: ctx, settings: settings}

				var walk func(*ast.Node)
				walk = func(current *ast.Node) {
					if current == nil {
						return
					}
					isFunction := oneVarIsFunctionScope(current)
					isBlock := !isFunction && oneVarIsBlockScope(current)
					if isFunction {
						checker.scopes.startFunction()
					} else if isBlock {
						checker.scopes.startBlock()
					}

					if current.Kind == ast.KindVariableDeclarationList {
						checker.checkDeclarationList(current)
					}

					current.ForEachChild(func(child *ast.Node) bool {
						walk(child)
						return false
					})

					if isFunction {
						checker.scopes.endFunction()
					} else if isBlock {
						checker.scopes.endBlock()
					}
				}
				walk(node)
			},
		}
	},
}
