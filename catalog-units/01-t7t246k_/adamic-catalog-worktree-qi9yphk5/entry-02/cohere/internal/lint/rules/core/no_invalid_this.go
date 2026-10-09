package core

import (
	"strings"
	"unicode"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/ecmascript/comments"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoInvalidThisUnexpectedThis = rule.Message{
	Id: "unexpectedThis",
	Description: "Nothing binds `this` here, so it is `undefined` at runtime and every property read " +
		"off it throws. The usual cause is a method that was moved out of a class or an object " +
		"literal and left its `this.` prefixes behind. Take what it needs as a parameter, give the " +
		"function a `this` parameter so TypeScript can check the call sites, or make it a method again.",
}

// NoInvalidThis flags a `this` in a context where `this` is not bound to anything.
//
//	valid:   class A { foo() { this.x; } }
//	valid:   const obj = { foo() { this.x; } };
//	valid:   function Foo() { this.x = 1; }              (capitalized, assumed a constructor)
//	valid:   array.forEach(function () { this.x; }, thisArg);
//	valid:   class A { foo = this.bar; }
//	invalid: function foo() { this.x; }
//	invalid: const foo = function () { this.x; };
//	invalid: class A { [this.key]; }                     (a computed key has no receiver)
//
// # This rule already ships under another name, and porting it anyway was a measured decision
//
// `@typescript-eslint/no-invalid-this` is registered and enabled in this tree, and upstream's
// extension is a WRAPPER: its source opens with `getESLintCoreRule('no-invalid-this')` and every
// finding comes from the core rule underneath. The standard's usual reading of that shape is that
// the core algorithm necessarily already lives inside the namespaced port, so a bare port is a
// second implementation of code already present.
//
// That reading is right about the algorithm and wrong about the VERDICTS, and the difference was
// measured rather than argued. Both installed rules -- ESLint 10.8.1's core and
// @typescript-eslint 8.67.0's extension -- were driven over all 572 cases of upstream's core corpus
// through the ESLint Linter interface. The two agree on 566 and disagree on 6, and the six are one
// class: a `this` inside the COMPUTED KEY of a class field.
//
//	class C { [this.foo]; }                          core reports, the extension is silent
//	class C { static {} [this]; }                    core reports, the extension is silent
//	function foo() { class C { accessor [this.a] = foo; } }   core reports, extension silent
//
// The cause is in the wrapper. It pushes `true` onto its validity stack for `PropertyDefinition`
// and `AccessorProperty`, which in ESTree is the whole member INCLUDING its key, so a `this` in the
// key is exempted along with one in the value. Core has no such arm and judges the key against the
// enclosing function, which is correct: a computed key is evaluated at class-definition time, where
// the instance does not exist yet.
//
// So the extension carries a false-negative class and the core rule is the one that catches it.
// That is why this port is worth having beside the namespaced one rather than being redundant with
// it, and it is the answer to the standard's "expect such a port to add no new findings" -- here it
// adds six shapes, and they are real.
//
// # This rule is ENABLED on the live tree, and `--rules-enabled` says it is not
//
// Read this before you conclude anything about whether this rule runs. It was landed deliberately
// unenabled, `CohereSettings.json` contains no `no-invalid-this` key, and it runs anyway.
//
// `settingFor` (`internal/lint/configuration/resolve.go:96`) resolves a config key by exact match,
// and failing that by trimming each configured key by the rule name and requiring what remains to
// end in `/`. The live config carries `"@typescript-eslint/no-invalid-this": "error"`. Trim that
// by this rule's name and `"@typescript-eslint/"` remains, which ends on the boundary, so the
// namespaced sibling's key resolves THIS rule too:
//
//	config key : @typescript-eslint/no-invalid-this
//	rule name  : no-invalid-this
//	remainder  : "@typescript-eslint/"   ends in "/", so it matches
//
// That suffix rule exists for a good reason and is documented at its own line: it lets one config
// entry configure a rule whichever namespace it is spelled under. The case it did not anticipate is
// an `extendsBaseRule` PAIR, where the two names are a genuine suffix of each other and are two
// different rules with two different verdicts.
//
// # The instrument disagrees with the behaviour, and the instrument is the one people read
//
// `--rules-enabled` lists literal config keys, so it reports this rule as not enabled. The resolver
// reaches it by suffix and runs it. Both are working as written and they answer differently, which
// makes the honest-reporting surface the thing that misleads.
//
// So do not settle this question with `--rules-enabled`, and do not settle it on the ahra tree
// either: ahra is clean for this rule, and a zero there cannot separate "did not run" from "ran and
// found nothing". That was tried and it could not distinguish the two.
//
// What does settle it is a probe tree with a PLANTED violation, whose config names only the
// namespaced key:
//
//	<scratch>/probetree/CohereSettings.json   { "rules": { "@typescript-eslint/no-invalid-this": "error" } }
//	<scratch>/probetree/tsconfig.json         any config including source/**/*.ts -- without one,
//	                                          cohere exits before linting and the error greps as silence
//	<scratch>/probetree/source/Seed.ts        export function seedOne() { return this.value; }
//
//	cd <scratch>/probetree && <your build> --no-fix --lint
//
// Measured: TWO findings at the identical line and column, one tagged
// `[@typescript-eslint/no-invalid-this/unexpectedThis]` and one tagged `[no-invalid-this/...]`.
// That is the standard's documented "a core rule and its extension can both be on, and both will
// fire" doubling, arriving through a spelling suffix rather than through anyone writing a second
// key.
//
// Of the registered rules absent from `--rules-enabled`, this is the ONLY one in this state.
// Checked mechanically rather than assumed: every bare-named one was tested for an enabled
// namespaced rule ending in `/<that name>`, and only this rule matched. The other unenabled rules
// have no enabled namespaced twin, so their bare names resolve to nothing and they really are off.
// That is why the usual "32 rules already ship unenabled" precedent does not cover this one.
//
// # Two decisions this rule does not make
//
// Silencing it would take an explicit `"no-invalid-this": "off"` in `CohereSettings.json`, which is
// a well-established shape there (fifteen keys are already `off`). That is Kirk's call, not a
// porter's, and turning a rule on or off through a spelling difference rather than because someone
// changed their mind is the worst way to make such a change -- it leaves no trace and reads as an
// accident nobody authored.
//
// The deeper question is whether suffix resolution should reach a bare rule at all when a
// namespaced sibling is configured. That is a design decision about the resolver rather than a
// config edit, and it is left open here on purpose. Both are recorded rather than resolved.
//
// # The same defect is present in this tree's namespaced port, and is NOT repaired here
//
// `internal/lint/rules/typescript/no_invalid_this.go` reproduces the wrapper faithfully, including
// this arm, so it is silent on all six shapes. Measured directly against it with a probe rather
// than read off its source. It is additionally silent on a seventh shape where the EXTENSION does
// report:
//
//	class C { [this.a]() {} }        core reports, the extension reports, our port is silent
//
// That one is ours alone: the port pushes `true` for `KindMethodDeclaration` before recursing, and
// `ForEachChild` visits the computed key under that push, so the key is judged under the method's
// binding. Upstream's wrapper has no method arm at all, so it never had the chance to make this
// mistake.
//
// Repairing that is a change to another author's committed and ENABLED rule, which would alter what
// the tree reports, and it is not this port's business to make. It is recorded here, with what
// established it, so the next reader finds it rather than re-deriving it.
//
// The shapes are pinned as EXECUTABLE assertions rather than as prose, in this package's own
// tests, so a later change to either rule fails loudly instead of quietly agreeing with a comment:
//
//	go test -run TestNoInvalidThisComputedKeysAreOutsideTheMember ./internal/lint/rules/core/
//
// The claim about the namespaced port was measured, not read off its source. A throwaway probe
// package ran `typescript.NoInvalidThis` over all seven shapes plus two controls that both rules
// report, and it came back silent on all seven while the controls fired. Anyone re-checking should
// rebuild that probe rather than trust this paragraph -- a comment is its author's earlier claim,
// not evidence.
//
// # Fidelity is to the module-mode verdicts, because that is the only configuration here
//
// Core's rule is built on ESLint's code path analysis and consults `scope.isStrict` on every
// function, because in sloppy mode `this` falls back to the global object and is therefore always
// valid. It also special-cases `sourceType: module` and the `globalReturn` parser feature at the
// program level.
//
// None of that survives contact with what cohere lints. Every file here is a TypeScript module, so
// every function body is strict and top-level `this` is always `undefined`. This is the hazard the
// standard names for `strict`: the corpus and this tree are being asked different questions, and
// the corpus's declared verdicts are not portable as written. Measured -- of the 474 cases in
// upstream's JavaScript matrix, 62 change verdict when the same source is re-run as a module, and
// every one of them is a sloppy-mode case where `this` was the global object.
//
// So the fixtures assert what the installed core rule answers under `sourceType: module` with the
// TypeScript parser, not what the corpus file annotates. The corpus's four conditions -- NORMAL,
// USE_STRICT, IMPLIED_STRICT and MODULES -- collapse to one configuration here, and that collapse
// was checked rather than assumed: grouping the 569 distinct (source, options) pairs by their
// source with the condition markers stripped gives 215 groups, and ZERO of those groups contain
// members that disagree. A group whose members disagreed would have meant a marker was doing real
// work and the collapse was hiding it.
//
// # The code path layer reduces to a stack of what we are inside
//
// With strictness settled, core's `onCodePathStart` / `onCodePathEnd` pair is a stack of function
// nodes, which is what the walk below maintains. Its arms are not interchangeable:
//
//	FunctionDeclaration / FunctionExpression   push the answer to the default-binding question
//	MethodDeclaration / accessors / ctor       push true, the receiver is the instance
//	PropertyDeclaration                        push true for the VALUE only, see below
//	ClassStaticBlockDeclaration                push true, `this` is the class
//	ArrowFunction                              push NOTHING, an arrow inherits the binding
//
// The arrow arm is an absence rather than a case, and it is what makes
// `function foo() { this.x; z(() => this.x); }` report twice rather than once: both `this` reads
// resolve against the same enclosing function. A port that pushed for arrows would report the outer
// one and silently exempt the inner. Fifty-one corpus cases carry that exact shape.
//
// # The computed-key discriminator is a node here, which is what makes this portable
//
// Upstream has to reason about which CHILD of the member it is in, because in ESTree the key hangs
// directly off the member node. Here a computed key is its own `KindComputedPropertyName` node
// sitting between the member and the key expression, measured with an ancestry probe over eight
// class shapes. So the walk gets an arm on that node rather than a test against the member's
// fields, and needs no knowledge of which slot the key occupies.
//
// That arm POPS the member's push for the duration of the key and restores it afterwards, rather
// than forcing a report. The difference is load-bearing and cost a draft:
// `class C { static { class D { [this.x]; } } }` is a PASSING corpus case, because popping D's
// member exposes the static block's binding underneath, which is valid. A version that reported
// whenever it found a computed property name in the ancestry reported that case.
//
// # capIsConstructor is on by default, so the zero-value struct is the wrong configuration
//
// The option surface is one boolean and its default is TRUE, which means a decoder that fills a
// zero-value struct inverts it and silently reports every `function Foo() { this.x = 1; }` in the
// tree. Forty-five corpus cases carry the option explicitly, and the decoder below is hand written
// for exactly this reason.
var NoInvalidThis = rule.Rule{
	Name: "no-invalid-this",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings := noInvalidThisSettingsFrom(options)

		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				// The walk is done here rather than through per-kind listeners because the rule
				// needs to know what it is INSIDE when it meets a `this`, and the listener walk is
				// pre-order with no exit hook to pop a stack on.
				walker := &noInvalidThisWalker{ctx: ctx, settings: settings}
				walker.walk(node)
			},
		}
	},
}

// noInvalidThisSettings is what the rule reads after decoding.
type noInvalidThisSettings struct {
	// CapIsConstructor treats a function whose name starts with an uppercase letter as a
	// constructor, so `this` inside it is bound. Defaults to true, which is why this type exists
	// separately from its wire shape.
	CapIsConstructor bool
}

// DefaultNoInvalidThisSettings is upstream's `defaultOptions`, spelled out because the zero value of
// the struct is not it.
func DefaultNoInvalidThisSettings() noInvalidThisSettings {
	return noInvalidThisSettings{CapIsConstructor: true}
}

// noInvalidThisRawOptions is the wire shape. The pointer distinguishes an absent key from
// `capIsConstructor: false`, which mean opposite things for a default-true option.
type noInvalidThisRawOptions struct {
	CapIsConstructor *bool `json:"capIsConstructor"`
}

// DecodeNoInvalidThisOptions maps the wire key onto the setting the rule reads.
//
// Hand written rather than `rule.DecodeOptionsInto` because the only default is true: a rule
// configured as a bare "error" is handed nil options, the generic decoder errors on empty input, and
// the resulting zero-value struct would say capIsConstructor is false. That is not a weaker
// configuration, it is a different rule, and it reports every capitalized constructor function in
// the tree.
func DecodeNoInvalidThisOptions(raw []byte) (any, error) {
	settings := DefaultNoInvalidThisSettings()

	decoded, err := rule.DecodeOptionsInto[noInvalidThisRawOptions]()(raw)
	if err != nil {
		return settings, err
	}
	wire, _ := decoded.(noInvalidThisRawOptions)

	if wire.CapIsConstructor != nil {
		settings.CapIsConstructor = *wire.CapIsConstructor
	}
	return settings, nil
}

// noInvalidThisSettingsFrom recovers the settings from whatever the config layer handed over,
// falling back to the documented defaults for nil.
func noInvalidThisSettingsFrom(options any) noInvalidThisSettings {
	if settings, ok := rule.OptionsAs[noInvalidThisSettings](options); ok {
		return settings
	}
	return DefaultNoInvalidThisSettings()
}

// noInvalidThisWalker carries the validity stack across the recursive walk.
type noInvalidThisWalker struct {
	ctx      rule.Context
	settings noInvalidThisSettings

	// thisIsValid's top is the binding in force for the `this` being visited, and an empty stack
	// means top level, where a module's `this` is `undefined` and therefore invalid.
	thisIsValid []bool
}

func (w *noInvalidThisWalker) walk(node *ast.Node) {
	if node == nil {
		return
	}

	pushed := false
	// popped records that a computed key temporarily removed its member's push, so the walk can
	// put it back on the way out. Restoring `true` rather than the saved value is correct because
	// only the member arms push here, and every one of them pushes true.
	popped := false
	switch node.Kind {
	case ast.KindThisKeyword:
		if !w.currentlyValid() {
			w.ctx.ReportNode(node, messageNoInvalidThisUnexpectedThis)
		}

	case ast.KindComputedPropertyName:
		// A computed key is evaluated when the class or object literal is DEFINED, before any
		// instance exists, so the member that encloses it binds nothing: `class C { [this.a]; }`
		// throws. The walk has already pushed for that member, so the push has to be UNDONE for
		// the key -- the member's value genuinely is valid and both live under the same node.
		//
		// Undoing is a pop rather than a forced report, and that distinction is load-bearing.
		// `class C { static { class D { [this.x]; } } }` is a PASSING corpus case: popping D's
		// member exposes the static block's binding underneath, which is valid. Forcing a report
		// here instead reported it, and forcing one is what an earlier draft did.
		if len(w.thisIsValid) > 0 {
			w.thisIsValid = w.thisIsValid[:len(w.thisIsValid)-1]
			popped = true
		}

	case ast.KindFunctionDeclaration, ast.KindFunctionExpression:
		w.thisIsValid = append(w.thisIsValid, !thisIsDefaultBoundIn(w.ctx, node, w.settings))
		pushed = true

	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor,
		ast.KindConstructor, ast.KindClassStaticBlockDeclaration:
		// A method's `this` is the receiver and a static block's is the class. Core reaches these
		// through the same default-binding test, which answers false for all of them.
		w.thisIsValid = append(w.thisIsValid, true)
		pushed = true

	case ast.KindPropertyDeclaration:
		// A field initializer is an implicit function whose `this` is the instance. `accessor` is a
		// modifier here rather than a separate node, so upstream's PropertyDefinition and
		// AccessorProperty both arrive as this kind.
		w.thisIsValid = append(w.thisIsValid, true)
		pushed = true

	case ast.KindArrowFunction:
		// Deliberately no push. An arrow has no `this` of its own, so the enclosing binding stays
		// in force and a `this` inside it is judged exactly as one outside it would be.
	}

	node.ForEachChild(func(child *ast.Node) bool {
		w.walk(child)
		return false
	})

	if pushed {
		w.thisIsValid = w.thisIsValid[:len(w.thisIsValid)-1]
	}
	if popped {
		w.thisIsValid = append(w.thisIsValid, true)
	}
}

// currentlyValid reads the top of the stack.
//
// An empty stack is top level. Every file cohere lints is a TypeScript module, where top-level
// `this` is `undefined`, so the answer there is false. Core reaches the same answer through
// `node.sourceType === "module"`.
func (w *noInvalidThisWalker) currentlyValid() bool {
	if len(w.thisIsValid) == 0 {
		return false
	}
	return w.thisIsValid[len(w.thisIsValid)-1]
}

// thisIsDefaultBoundIn is ESLint's `astUtils.isDefaultThisBinding`, which is where the judgment lives.
//
// It answers "would calling this function leave `this` as the default binding", and the rule reports
// exactly when it says yes. The traversal walks OUTWARD from the function looking for something that
// binds a receiver, and every arm below is upstream's, in upstream's order.
//
// The direction is worth stating because it inverts twice on the way to the report site: this returns
// true when `this` is UNBOUND, the stack stores whether `this` is VALID, and the report fires when
// the stack says invalid.
func thisIsDefaultBoundIn(ctx rule.Context, node *ast.Node, settings noInvalidThisSettings) bool {
	// A declared `this` parameter binds `this` by contract, and the checker enforces it at every
	// call site. This lives in CORE, not only in the typescript-eslint extension: ESLint 10.8.1's
	// own `isDefaultThisBinding` opens with it, and four corpus cases in upstream's TypeScript
	// block depend on it, including `z(function (x, this: context) { ... })`. Upstream tests every
	// parameter rather than only the first, and that width is reproduced: the grammar allows `this`
	// only in first position, but the parser recovers from `function f(x, this: T)` and hands back
	// the illegal shape, which is exactly what that corpus case writes.
	if (node.Kind == ast.KindFunctionDeclaration || node.Kind == ast.KindFunctionExpression) &&
		noInvalidThisDeclaresThisParameter(node) {
		return false
	}

	// A capitalized name is the convention for a constructor, and upstream trusts it. This is the
	// single option surface, and it reads the function's own name.
	if settings.CapIsConstructor && noInvalidThisStartsWithUpperCase(noInvalidThisFunctionOwnName(node)) {
		return false
	}
	if noInvalidThisHasJSDocThisTag(ctx, node) {
		return false
	}

	isAnonymous := noInvalidThisFunctionOwnName(node) == ""
	current := node
	for current != nil {
		parent := current.Parent
		if parent == nil {
			return true
		}

		switch parent.Kind {
		// typescript-go keeps parentheses as a node and the parser upstream ports from does not,
		// so every arm below would otherwise be unreachable through a parenthesized form. This is
		// a parser difference rather than a judgment, and it is load-bearing on five corpus cases:
		// `obj.foo = (function () { return function () { this.x; }; })();` and
		// `(function () { this.x; }).call(obj);` both put a parenthesis between the function and
		// the node that binds its receiver, and without this skip both reach the default arm and
		// report. Upstream sees no such node and needs no such arm.
		case ast.KindParenthesizedExpression:
			current = parent
			continue

		// A default parameter value is upstream's AssignmentPattern, whose arm is shared with
		// AssignmentExpression: a capitalized target binds by naming convention and anything else
		// does not. typescript-go models the same source as a Parameter carrying an Initializer, so
		// the shape is reached by a different kind and answered the same way. The corpus case is
		// `function foo(Ctor = function () { this.x; }) {}`, which passes because Ctor is
		// capitalized, and its lowercase twin reports.
		case ast.KindParameter:
			if parent.Initializer() != current {
				return true
			}
			name := parent.Name()
			return !(settings.CapIsConstructor && isAnonymous &&
				name != nil && name.Kind == ast.KindIdentifier &&
				noInvalidThisStartsWithUpperCase(name.Text()))

		// Look through to the destination: `obj.foo = nativeFoo || function foo() {};`
		case ast.KindBinaryExpression:
			if noInvalidThisIsLogicalOperator(parent) {
				current = parent
				continue
			}
			return noInvalidThisAssignmentBindsThis(parent, current, isAnonymous, settings)

		case ast.KindConditionalExpression:
			current = parent
			continue

		// If the enclosing function is immediately invoked, follow its return value out.
		//
		// The step out is to the CALL, not to the function's own parent, and the two differ here
		// where they do not upstream: `(function () { return function () {}; })()` gives the inner
		// function a parenthesized parent, so `enclosing.Parent` is the parenthesis rather than the
		// call. Upstream's `func.parent` IS the call because its parser drops the node.
		case ast.KindReturnStatement:
			enclosing := noInvalidThisEnclosingFunctionOf(parent)
			if enclosing == nil {
				return true
			}
			call := noInvalidThisCallOf(enclosing)
			if call == nil {
				return true
			}
			current = call
			continue

		case ast.KindArrowFunction:
			if current != parent.Body() {
				return true
			}
			call := noInvalidThisCallOf(parent)
			if call == nil {
				return true
			}
			current = call
			continue

		// A method or a property's value has a receiver; anything else in these nodes does not.
		//
		// BOTH of these arms are INERT here, and both were measured rather than argued after their
		// mutants survived the whole 215-case corpus AND thirteen shapes hand-built to attack them.
		// The cause is one parser difference in two costumes: typescript-go puts a computed key in
		// its own ComputedPropertyName node and folds a method's body into the method node, where
		// ESTree hangs both directly off the member.
		//
		//	const obj = { [function () {}]: 1 };   parent is ComputedPropertyName, not PropertyAssignment
		//	class C { [function () {}] = 1; }      parent is ComputedPropertyName, not PropertyDeclaration
		//	class C { [function () {}]() {} }      parent is ComputedPropertyName, not MethodDeclaration
		//	class C { foo() {} }                   no function expression node exists at all
		//
		// So `parent.Initializer() != current` can never see a key -- a function reaching that arm is
		// always the value -- and no function expression ever has a method-kind parent. Swept over
		// fifteen class and object-literal shapes including static, private, accessor and computed
		// forms, with a control asserting the sweep saw a member parent at all:
		//
		//	go test -run TestMemberArmReachability ./internal/<probe>/
		//
		// Both are kept because fidelity is to upstream's decision and because the arms cost nothing,
		// but a reader should know they are unreachable rather than deleting them as dead or, worse,
		// trusting them to make a distinction they cannot make. The distinction they LOOK like they
		// draw is real and is drawn instead by the KindComputedPropertyName arm of the walk above.
		//
		// An earlier revision of this comment asserted the class-field case was "NOT inert and is
		// what separates the key from the value". That was wrong, and the mutation sweep is what
		// caught it.
		case ast.KindPropertyAssignment, ast.KindPropertyDeclaration:
			return parent.Initializer() != current
		case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
			return false

		// `var Foo = function () {};` binds by naming convention only.
		case ast.KindVariableDeclaration:
			return !(settings.CapIsConstructor && isAnonymous &&
				parent.Initializer() == current &&
				parent.Name() != nil && parent.Name().Kind == ast.KindIdentifier &&
				noInvalidThisStartsWithUpperCase(parent.Name().Text()))

		// `.bind(obj)`, `.call(obj)`, `.apply(obj, [])` all supply a receiver.
		//
		// The `Expression() == current` half is upstream's and is inert here, kept for fidelity
		// rather than because it decides anything. Upstream needs it because its MemberExpression
		// carries both the object and the property as expression children; typescript-go puts the
		// property in its own `name` slot, so a function reaching this arm is always the receiver.
		case ast.KindPropertyAccessExpression:
			if parent.Expression() == current && noInvalidThisIsBindCallOrApply(parent) {
				// The call is read back through noInvalidThisCallOf rather than from
				// `parent.Parent`, so the parenthesis skip that decides whether this IS a callee
				// and the lookup of the call's arguments cannot disagree about which node the call
				// is. Reading `parent.Parent` here was wrong for `(fn.bind)(obj)`: the skip found
				// the call and the argument read found the parenthesis.
				call := noInvalidThisCallOf(parent)
				return !(call != nil && noInvalidThisCallHasNonNullishArgument(call, 0))
			}
			return true

		// `Reflect.apply(fn, obj, [])`, `Array.from([], fn, obj)`, `list.forEach(fn, obj)`.
		case ast.KindCallExpression:
			return noInvalidThisCallDoesNotBindThis(parent, current)

		default:
			return true
		}
	}
	return true
}

// noInvalidThisFunctionOwnName returns the name a function carries itself, or empty for an anonymous
// one.
//
// Upstream reads `node.id`, which exists only for a declaration or a named function expression. A
// name inherited from a variable or a property is deliberately NOT this: upstream handles that
// separately in the VariableDeclaration and AssignmentExpression arms, and only for anonymous
// functions, so folding it in here would report differently on `var Foo = function bar() { this.x; };`.
func noInvalidThisFunctionOwnName(node *ast.Node) string {
	name := node.Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return ""
	}
	return name.Text()
}

// noInvalidThisStartsWithUpperCase is upstream's constructor-naming convention test.
//
// Upstream writes `name[0] !== name[0].toLocaleLowerCase()`, which is a "has an uppercase form that
// differs" test rather than an "is uppercase" test, and the two disagree on characters with no case
// at all. Reproduced with unicode.ToLower rather than unicode.IsUpper for that reason: `_foo` and
// `1foo` are lowercase-invariant and answer false either way, but a caseless letter answers false
// here and true under IsUpper.
func noInvalidThisStartsWithUpperCase(name string) bool {
	if name == "" {
		return false
	}
	first := []rune(name)[0]
	return unicode.ToLower(first) != first
}

// noInvalidThisHasJSDocThisTag answers whether a `@this` tag documents this function.
//
// Upstream asks two questions and reports if either says yes, and both are reproduced because both
// were measured to matter. This is the largest single piece of machinery behind a rule whose own
// file is under two hundred lines, and it is a good example of the rule file being a thin caller.
//
// The first question is `getJSDocComment`, which walks OUTWARD from the function to whatever node
// the documentation would sit on, and then requires the token immediately before that node to be a
// real JSDoc block. The walk is why the boundary is not a token gap, measured at ESLint 10.8.1:
//
//	function foo() { /** @this*/ return function () { this.x; }; }        silent
//	function foo() { /** @this*/ var q = function () { this.x; }; }       silent
//	function foo() { /** @this*/ q(function () { this.x; }); }            REPORTS
//
// The third is the one that pins the shape: a function passed as a CALL ARGUMENT does not walk out
// at all, because upstream excludes a call or new parent before the loop starts.
//
// The second question is `getCommentsBefore`, upstream's stated fallback for callbacks, which reads
// every comment token in the gap before the function itself, block or line, and asks the same
// pattern of each. The pattern is `/^[\s*]*@this/mu`: anchored per line, allowing asterisks in the
// prefix, and with NO closing word boundary, so `@thisx` satisfies it and `@nothis` does not. That
// reads as a defect and is reproduced rather than corrected, because fidelity is to the authority
// where upstream gives no reasoning.
func noInvalidThisHasJSDocThisTag(ctx rule.Context, node *ast.Node) bool {
	if ctx.SourceFile == nil {
		return false
	}
	if noInvalidThisCommentTextHasThisTag(noInvalidThisJSDocCommentFor(ctx, node)) {
		return true
	}
	for _, comment := range noInvalidThisCommentsImmediatelyBefore(ctx, node) {
		if noInvalidThisCommentTextHasThisTag(comment.Text) {
			return true
		}
	}
	return false
}

// noInvalidThisJSDocCommentFor is upstream's `getJSDocComment` for the function shapes this rule
// reaches, which are a declaration and an expression. It returns the documenting comment's text, or
// empty.
func noInvalidThisJSDocCommentFor(ctx rule.Context, node *ast.Node) string {
	if node.Kind == ast.KindFunctionDeclaration {
		return noInvalidThisJSDocBlockBefore(ctx, node)
	}

	parent := node.Parent
	if parent == nil {
		return noInvalidThisJSDocBlockBefore(ctx, node)
	}
	// A function handed straight to a call does not walk out. This is what separates
	// `q(function () {...})` from `var q = function () {...}` and it is upstream's own guard.
	if parent.Kind == ast.KindCallExpression || parent.Kind == ast.KindNewExpression {
		return noInvalidThisJSDocBlockBefore(ctx, node)
	}

	for parent != nil {
		if len(noInvalidThisCommentsImmediatelyBefore(ctx, parent)) > 0 ||
			noInvalidThisIsFunctionLike(parent) ||
			parent.Kind == ast.KindMethodDeclaration || parent.Kind == ast.KindPropertyAssignment {
			break
		}
		parent = parent.Parent
	}
	if parent != nil && parent.Kind != ast.KindFunctionDeclaration && parent.Kind != ast.KindSourceFile {
		return noInvalidThisJSDocBlockBefore(ctx, parent)
	}
	return noInvalidThisJSDocBlockBefore(ctx, node)
}

// noInvalidThisJSDocBlockBefore is upstream's `findJSDocComment`: the token immediately before the
// node has to be a block comment whose text opens with an asterisk, and it has to end no more than
// one line above.
func noInvalidThisJSDocBlockBefore(ctx rule.Context, node *ast.Node) string {
	preceding := noInvalidThisCommentsImmediatelyBefore(ctx, node)
	if len(preceding) == 0 {
		return ""
	}
	last := preceding[len(preceding)-1]
	if !last.IsBlock || !strings.HasPrefix(strings.TrimPrefix(last.Text, "/*"), "*") {
		return ""
	}
	nodeLine, _ := scanner.GetLineAndCharacterOfPosition(ctx.SourceFile,
		rule.TokenRange(ctx.SourceFile, node).Pos())
	commentLine, _ := scanner.GetLineAndCharacterOfPosition(ctx.SourceFile, last.Range.End())
	if nodeLine-commentLine > 1 {
		return ""
	}
	return last.Text
}

// noInvalidThisCommentsImmediatelyBefore is upstream's `getCommentsBefore`: every comment sitting in
// the trivia between the previous real token and this node, in source order.
//
// Other comments in the gap do not close it, which is why the span is scanned for comments rather
// than merely tested for whitespace. Any non-comment text does close it, and that is what makes the
// gap belong to `1` rather than to the callback in `z(/* @this */ 1, function () {})`.
func noInvalidThisCommentsImmediatelyBefore(ctx rule.Context, node *ast.Node) []comments.Comment {
	tokenStart := rule.TokenRange(ctx.SourceFile, node).Pos()
	sourceText := ctx.SourceFile.Text()
	if tokenStart > len(sourceText) {
		return nil
	}

	found := []comments.Comment{}
	cursor := tokenStart
	all := comments.ForFile(ctx)
	for index := len(all) - 1; index >= 0; index-- {
		comment := all[index]
		if comment.Range.End() > cursor {
			continue
		}
		if text.TrimWhitespace(sourceText[comment.Range.End():cursor]) != "" {
			break
		}
		found = append([]comments.Comment{comment}, found...)
		cursor = comment.Range.Pos()
	}
	return found
}

// noInvalidThisIsFunctionLike answers whether a node introduces its own `this` or parameter scope,
// which is where upstream's outward walk stops.
func noInvalidThisIsFunctionLike(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
		ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor,
		ast.KindConstructor, ast.KindClassStaticBlockDeclaration:
		return true
	}
	return false
}

// noInvalidThisCommentTextHasThisTag tests upstream's `/^[\s*]*@this/mu` against a comment's text.
//
// Line by line rather than with a regular expression, because the pattern is anchored per line and
// asks one question of each: after any run of whitespace and asterisks, does `@this` begin. The
// delimiters are stripped first, since upstream matches against `comment.value` rather than the raw
// source of the comment.
func noInvalidThisCommentTextHasThisTag(text string) bool {
	if text == "" {
		return false
	}
	value := strings.TrimPrefix(text, "/*")
	value = strings.TrimSuffix(value, "*/")
	value = strings.TrimPrefix(value, "//")

	for _, line := range strings.Split(value, "\n") {
		if strings.HasPrefix(strings.TrimLeft(line, " \t\r*"), "@this") {
			return true
		}
	}
	return false
}

// noInvalidThisIsLogicalOperator answers whether a binary expression is `||`, `&&` or `??`.
//
// These are upstream's LogicalExpression, which it looks THROUGH rather than deciding on, because
// `obj.foo = cached || function () {}` still assigns the function to `obj.foo`. Every other binary
// operator, assignment included, is a decision point.
func noInvalidThisIsLogicalOperator(node *ast.Node) bool {
	switch node.AsBinaryExpression().OperatorToken.Kind {
	case ast.KindBarBarToken, ast.KindAmpersandAmpersandToken, ast.KindQuestionQuestionToken:
		return true
	}
	return false
}

// noInvalidThisAssignmentBindsThis handles upstream's shared AssignmentExpression / AssignmentPattern
// arm.
//
// A member-expression target binds a receiver outright, and a capitalized identifier target binds one
// by naming convention, but only for an anonymous function -- a named one has already been judged on
// its own name.
//
// Upstream tests NEITHER the operator NOR which side the function is on, and both omissions are
// reproduced because both were measured to matter. An earlier draft here required
// `KindEqualsToken` and `Right == current`, which is the intuitive reading and is wrong: the three
// logical assignment operators `&&=`, `||=` and `??=` are AssignmentExpression too, and
// `obj.method &&= function () { this.x; };` is a PASSING corpus case that the operator test turned
// into two false findings. Measured against ESLint 10.8.1, all three report zero.
func noInvalidThisAssignmentBindsThis(parent *ast.Node, current *ast.Node, isAnonymous bool,
	settings noInvalidThisSettings) bool {
	left := parent.AsBinaryExpression().Left
	for left != nil && left.Kind == ast.KindParenthesizedExpression {
		left = left.AsParenthesizedExpression().Expression
	}
	if left == nil {
		return true
	}
	if left.Kind == ast.KindPropertyAccessExpression || left.Kind == ast.KindElementAccessExpression {
		return false
	}
	return !(settings.CapIsConstructor && isAnonymous &&
		left.Kind == ast.KindIdentifier && noInvalidThisStartsWithUpperCase(left.Text()))
}

// noInvalidThisDeclaresThisParameter answers whether a function declares an explicit `this`
// parameter, which TypeScript spells as a first parameter literally named `this`.
func noInvalidThisDeclaresThisParameter(node *ast.Node) bool {
	for _, parameter := range node.Parameters() {
		name := parameter.Name()
		if name != nil && name.Kind == ast.KindIdentifier && name.Text() == "this" {
			return true
		}
	}
	return false
}

// noInvalidThisEnclosingFunctionOf walks up to the function whose body contains this node.
func noInvalidThisEnclosingFunctionOf(node *ast.Node) *ast.Node {
	for current := node.Parent; current != nil; current = current.Parent {
		if noInvalidThisIsFunctionLike(current) {
			return current
		}
	}
	return nil
}

// noInvalidThisIsBindCallOrApply answers whether a property access names `bind`, `call` or `apply`.
func noInvalidThisIsBindCallOrApply(node *ast.Node) bool {
	name := node.Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return false
	}
	switch name.Text() {
	case "bind", "call", "apply":
		return true
	}
	return false
}

// noInvalidThisCallOf answers with the call that calls this node, or nil, looking through any
// parentheses between the node and the call.
//
// Upstream's own `isCallee` reads `node.parent` directly and needs no skip, because its parser drops
// parentheses; where it DOES need to look through something -- a `ChainExpression` for `?.` -- it
// says so explicitly at the MemberExpression arm. typescript-go has no ChainExpression (optional
// chaining is a flag on the access node) and does keep parentheses, so both of upstream's cases
// arrive here as the same shape and one skip answers both.
//
// This answers with the CALL rather than with a boolean, because the caller needs the call's
// arguments too and reading them from `parent.Parent` separately let the two disagree: the skip
// found the call while the argument read found the parenthesis, which is what made
// `(function () { this.x; }.bind)(obj)` report twice.
//
// Measured against ESLint 10.8.1, all of these are silent, and without the skip the first two
// report twice each:
//
//	var foo = (function () { this.x; }?.bind)(obj);
//	var foo = (function () { this.x; }.bind)(obj);
//	var foo = function () { this.x; }.bind(obj);
//
// Only the `?.` form appears in upstream's corpus, so its plain twin is a shape the corpus cannot
// flag -- upstream's parser deleted the distinguishing node before the test was written.
func noInvalidThisCallOf(node *ast.Node) *ast.Node {
	current := node
	for current.Parent != nil && current.Parent.Kind == ast.KindParenthesizedExpression {
		current = current.Parent
	}
	parent := current.Parent
	if parent == nil || parent.Kind != ast.KindCallExpression || parent.Expression() != current {
		return nil
	}
	return parent
}

// noInvalidThisCallHasNonNullishArgument answers whether the argument at this index is present and
// is neither `null` nor `undefined` nor `void 0`.
//
// Upstream's test is `isNullOrUndefined`, and it is what makes `.bind(null)` and `.call(undefined)`
// REPORT while `.bind(obj)` does not: binding to a nullish value leaves `this` undefined in strict
// mode, so the call supplies no receiver. Four corpus cases turn on this.
func noInvalidThisCallHasNonNullishArgument(call *ast.Node, index int) bool {
	if call == nil {
		return false
	}
	arguments := call.Arguments()
	if index >= len(arguments) {
		return false
	}
	return !noInvalidThisIsNullOrUndefined(arguments[index])
}

// noInvalidThisIsNullOrUndefined is upstream's `isNullOrUndefined`: the `null` literal, the
// identifier `undefined`, or a `void` expression of any operand.
func noInvalidThisIsNullOrUndefined(node *ast.Node) bool {
	for node != nil && node.Kind == ast.KindParenthesizedExpression {
		node = node.AsParenthesizedExpression().Expression
	}
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindNullKeyword:
		return true
	case ast.KindIdentifier:
		return node.Text() == "undefined"
	case ast.KindVoidExpression:
		return true
	}
	return false
}

// noInvalidThisCallDoesNotBindThis handles upstream's CallExpression arm: the shapes where a
// function handed to a known callee is given a receiver by a later argument.
//
//	Reflect.apply(fn, thisArg, [])            fn at 0, receiver at 1, exactly 3 arguments
//	Array.from(iterable, fn, thisArg)          fn at 1, receiver at 2, exactly 3 arguments
//	Array.fromAsync(iterable, fn, thisArg)     likewise
//	list.forEach(fn, thisArg)                  fn at 0, receiver at 1, exactly 2 arguments
//
// Anything else supplies no receiver, so the answer there is that `this` stays default-bound.
//
// The argument counts are EXACT upstream (`arguments.length !== 3`), not minimums, and that is
// reproduced. A minimum reads as the safer choice and is a different rule: `Array.from([], fn, obj,
// extra)` is not the two-argument-plus-thisArg form upstream recognises.
func noInvalidThisCallDoesNotBindThis(call *ast.Node, current *ast.Node) bool {
	arguments := call.Arguments()
	callee := call.Expression()
	for callee != nil && callee.Kind == ast.KindParenthesizedExpression {
		callee = callee.AsParenthesizedExpression().Expression
	}
	if callee == nil || callee.Kind != ast.KindPropertyAccessExpression {
		return true
	}

	object := callee.Expression()
	for object != nil && object.Kind == ast.KindParenthesizedExpression {
		object = object.AsParenthesizedExpression().Expression
	}
	method := callee.Name()
	if method == nil || method.Kind != ast.KindIdentifier {
		return true
	}
	methodName := method.Text()

	objectName := ""
	if object != nil && object.Kind == ast.KindIdentifier {
		objectName = object.Text()
	}

	if objectName == "Reflect" && methodName == "apply" {
		return len(arguments) != 3 || arguments[0] != current ||
			noInvalidThisIsNullOrUndefined(arguments[1])
	}

	// `Array.from` is matched through upstream's `arrayOrTypedArrayPattern`, which is `/Array$/`
	// rather than the literal name -- so every typed array (`Int8Array.from`, `Float64Array.from`)
	// takes this arm too. `Array.fromAsync` is matched by exact name, which is upstream's own
	// asymmetry and is reproduced rather than tidied.
	if (strings.HasSuffix(objectName, "Array") && methodName == "from") ||
		(objectName == "Array" && methodName == "fromAsync") {
		return len(arguments) != 3 || arguments[1] != current ||
			noInvalidThisIsNullOrUndefined(arguments[2])
	}

	if noInvalidThisIsArrayMethodTakingThisArg(methodName) {
		return len(arguments) != 2 || arguments[0] != current ||
			noInvalidThisIsNullOrUndefined(arguments[1])
	}

	return true
}

// noInvalidThisArrayMethodsTakingThisArg is upstream's `arrayMethodWithThisArgPattern`: the array
// methods whose second argument becomes the callback's `this`.
//
// Upstream writes it as `/^(?:every|filter|find(?:Last)?(?:Index)?|flatMap|forEach|map|some)$/u`,
// so the `find` group expands to four names rather than two. Enumerated here because the set is
// small and a spelled-out set is greppable where a regular expression is not.
//
// The name is prefixed rather than bare because internal/lint/rules/core is one Go namespace shared
// with every committed rule beside it, and a bare `arrayMethods` would collide with a sibling.
var noInvalidThisArrayMethodsTakingThisArg = map[string]bool{
	"every": true, "filter": true, "find": true, "findIndex": true,
	"findLast": true, "findLastIndex": true, "flatMap": true,
	"forEach": true, "map": true, "some": true,
}

func noInvalidThisIsArrayMethodTakingThisArg(name string) bool {
	return noInvalidThisArrayMethodsTakingThisArg[name]
}
