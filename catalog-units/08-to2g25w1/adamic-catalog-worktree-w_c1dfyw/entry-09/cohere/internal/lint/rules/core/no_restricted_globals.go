package core

import (
	"encoding/json"
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/property"
	"github.com/system-inc/cohere/internal/lint/ecmascript/reference"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoRestrictedGlobalsSettings is the decoded option surface.
//
// Upstream's schema is an `anyOf` over two shapes, and both have to be accepted because the corpus
// tests every case twice, once in each spelling:
//
//	["event", {"name": "fdescribe", "message": "Use describe instead."}]
//	[{"globals": ["event"], "checkGlobalObject": true, "globalObjects": ["myGlobal"]}]
//
// The bare-array form has no room for `checkGlobalObject`, so an option written that way always
// leaves it false. That is upstream's own reading: `isGlobalsObject` decides which shape it has and
// the two flags are read only from the object form.
type NoRestrictedGlobalsSettings struct {
	// Globals maps a restricted name to its custom message, or to the empty string for none.
	Globals map[string]string
	// GlobalsOrder preserves the configured order, so a file naming two restricted globals reports
	// them in a stable sequence rather than in Go's randomised map order.
	GlobalsOrder []string
	// CheckGlobalObject extends the rule to `window.event` as well as bare `event`.
	CheckGlobalObject bool
	// GlobalObjects are the names treated as the global object, on top of the three built in.
	GlobalObjects []string
}

// DefaultNoRestrictedGlobalsSettings is an empty list, which upstream treats as the rule doing
// nothing at all: `restrictedGlobals.length === 0` returns an empty visitor.
func DefaultNoRestrictedGlobalsSettings() NoRestrictedGlobalsSettings {
	return NoRestrictedGlobalsSettings{}
}

// noRestrictedGlobalsEntry is one item of the globals list in either of its two spellings.
type noRestrictedGlobalsEntry struct {
	Name    string
	Message string
}

// UnmarshalJSON accepts both a bare string and an object, which is upstream's `oneOf`.
func (entry *noRestrictedGlobalsEntry) UnmarshalJSON(raw []byte) error {
	var name string
	if err := json.Unmarshal(raw, &name); err == nil {
		entry.Name = name
		entry.Message = ""
		return nil
	}
	var object struct {
		Name    string `json:"name"`
		Message string `json:"message"`
	}
	// Strict, which is upstream's `additionalProperties: false`: a misspelled `message` key would
	// otherwise be accepted and the custom message silently never shown.
	if err := rule.UnmarshalOptions(raw, &object); err != nil {
		return err
	}
	if object.Name == "" {
		return fmt.Errorf("no-restricted-globals entry needs a name")
	}
	entry.Name = object.Name
	entry.Message = object.Message
	return nil
}

// noRestrictedGlobalsObjectWire is the second option shape.
type noRestrictedGlobalsObjectWire struct {
	Globals           []noRestrictedGlobalsEntry `json:"globals"`
	CheckGlobalObject *bool                      `json:"checkGlobalObject"`
	GlobalObjects     []string                   `json:"globalObjects"`
}

// DecodeNoRestrictedGlobalsOptions reads either of upstream's option shapes off the config.
//
// Hand rolled rather than routed through the generic helper because upstream's schema is an `anyOf`
// of two LISTS: every element a restricted global (a string or a `{name, message}` object), as in
// `["error", "event", {"name": "fdescribe", "message": "..."}]`, or exactly one element that is an
// object carrying that list plus two flags, `["error", {"globals": [...], "checkGlobalObject":
// true}]`. The rule registers with `DecodeOptionList` and is handed that list whole.
//
// The object form is recognised the way upstream recognises it, by the first element being an
// object with a `globals` key. That form takes no second element, so one is refused rather than
// dropped. Its keys are read strictly, which is upstream's `additionalProperties: false`: a
// misspelled `checkGlobalObject` would otherwise leave the flag silently off.
func DecodeNoRestrictedGlobalsOptions(list []byte) (any, error) {
	settings := DefaultNoRestrictedGlobalsSettings()
	if len(list) == 0 {
		return settings, nil
	}

	var elements []json.RawMessage
	if err := json.Unmarshal(list, &elements); err != nil {
		return settings, fmt.Errorf("no-restricted-globals: the option list is not a JSON array: %w", err)
	}
	if len(elements) == 0 {
		return settings, nil
	}

	var probe map[string]json.RawMessage
	if err := json.Unmarshal(elements[0], &probe); err == nil {
		if _, isGlobalsObject := probe["globals"]; isGlobalsObject {
			if len(elements) > 1 {
				return settings, fmt.Errorf(
					"no-restricted-globals: the {globals} object takes no second element, so %s "+
						"would never be read", elements[1])
			}
			var object noRestrictedGlobalsObjectWire
			if err := rule.UnmarshalOptions(elements[0], &object); err != nil {
				return settings, fmt.Errorf("no-restricted-globals element 1: %w", err)
			}
			checkGlobalObject := false
			if object.CheckGlobalObject != nil {
				checkGlobalObject = *object.CheckGlobalObject
			}
			return buildNoRestrictedGlobalsSettings(object.Globals, checkGlobalObject, object.GlobalObjects), nil
		}
	}

	entries := make([]noRestrictedGlobalsEntry, len(elements))
	for index, element := range elements {
		if err := json.Unmarshal(element, &entries[index]); err != nil {
			return settings, fmt.Errorf("no-restricted-globals element %d: %w", index+1, err)
		}
	}
	return buildNoRestrictedGlobalsSettings(entries, false, nil), nil
}

// buildNoRestrictedGlobalsSettings turns the decoded entries into the lookup the rule uses.
func buildNoRestrictedGlobalsSettings(entries []noRestrictedGlobalsEntry, checkGlobalObject bool,
	globalObjects []string) NoRestrictedGlobalsSettings {

	settings := NoRestrictedGlobalsSettings{
		Globals:           make(map[string]string, len(entries)),
		CheckGlobalObject: checkGlobalObject,
		GlobalObjects:     globalObjects,
	}
	for _, entry := range entries {
		if entry.Name == "" {
			continue
		}
		if _, seen := settings.Globals[entry.Name]; !seen {
			settings.GlobalsOrder = append(settings.GlobalsOrder, entry.Name)
		}
		// A later entry wins, matching upstream's `reduce` into one object.
		settings.Globals[entry.Name] = entry.Message
	}
	return settings
}

var messageNoRestrictedGlobalsDefault = rule.Message{
	Id: "defaultMessage",
	Description: "This name was restricted by configuration. It resolves to a global rather than " +
		"to anything declared here, so what it holds depends on the environment the code runs in.",
}

var messageNoRestrictedGlobalsCustom = rule.Message{
	Id:          "customMessage",
	Description: "This name was restricted by configuration.",
}

// NoRestrictedGlobals reports a reference to a global named in the configuration.
//
//	valid:   foo                                    (with no option, or restricting something else)
//	valid:   var foo = 1; foo                       (a local shadow is not the global)
//	valid:   let value: Test                        (a type position is not a value reference)
//	invalid: foo                                    (restricting "foo")
//	invalid: window.foo()                           (restricting "foo", checkGlobalObject on)
//
// # This rule does nothing without configuration, and that is upstream's design rather than a gap
//
// `create` returns an empty visitor when the list is empty, so an unconfigured rule registers no
// listener and reports nothing. Unlike `no-restricted-syntax`, which is in the same family, the
// configuration here is a list of plain NAMES rather than a selector language, so it is portable:
// nothing about the option depends on a node vocabulary our tree does not have.
//
// # The predicate is "does this name resolve to a source declaration", and it is the inverse of the
// obvious one
//
// Upstream reads two collections from its scope table: `scope.variables` filtered to those with no
// `defs` (names the environment declares, such as browser globals), and `scope.through` (names that
// resolve to nothing at all). We have no scope table, and the checker answers the union of both
// questions directly, but only if the predicate is written the right way round.
//
// `resolvesToAGlobal`, the shipped helper for this shape, answers FALSE when the symbol carries
// zero declarations, and that is precisely one of the cases here: measured, `globalThis` resolves
// to a symbol with no declarations, and `window` and `self` resolve to NO SYMBOL AT ALL under the
// fixture harness's `lib: ["ES2022"]`, identically to a name nobody declared anywhere. Using that
// helper would make this rule silent on most of what it must report.
//
// `identifierIsShadowed` is the correct predicate and it already exists in this package. It asks
// whether any declaration of the symbol lives in a real source file, so:
//
//	nil symbol                  not shadowed  -> report   (undeclared, upstream's scope.through)
//	zero declarations           not shadowed  -> report   (globalThis)
//	declared only in a .d.ts    not shadowed  -> report   (a browser global via lib or @types)
//	declared in source          SHADOWED      -> silent   (upstream's `variable.defs.length`)
//
// Reused rather than reimplemented, and the four rows above were each measured against the checker
// before this rule was written.
//
// # Type positions are exempt, and one corpus case pins the whole exclusion
//
// Upstream carries a `TYPE_NODES` set of five ESTree kinds and declines any reference whose parent
// is one of them. `const x: Promise<any> = Promise.resolve();` restricting `Promise` reports
// exactly ONCE: the annotation is exempt and the value reference is not. A port ignoring type
// positions reports twice on that single case, and a port exempting too much reports zero.
//
// Measured, upstream's five ESTree kinds collapse to three of ours. `TSTypeReference`,
// `TSInterfaceHeritage` and `TSClassImplements` are all `KindTypeReference` here, because our
// parser puts a heritage clause's element through the same node; `TSTypeQuery` is `KindTypeQuery`;
// and `TSQualifiedName` is `KindQualifiedName`.
//
// # No fix
//
// Upstream offers none, and none is possible: the repair is to use a different name, and nothing
// here knows which.
var NoRestrictedGlobals = rule.Rule{
	Name: "no-restricted-globals",

	// Telling a global from a local shadow is name resolution. See the predicate section above for
	// the four resolution outcomes and how each was measured.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, ok := rule.OptionsAs[NoRestrictedGlobalsSettings](options)
		if !ok || len(settings.Globals) == 0 {
			// Upstream returns an empty visitor when nothing is restricted. Registering no listener
			// is the same decision, and it is also what a rule configured as a bare severity gets,
			// since it is handed nil options and there is no sensible default list.
			return rule.Listeners{}
		}

		return rule.Listeners{
			ast.KindIdentifier: func(node *ast.Node) {
				checkNoRestrictedGlobalReference(ctx, node, settings)
			},
		}
	},
}

// checkNoRestrictedGlobalReference judges one identifier.
func checkNoRestrictedGlobalReference(ctx rule.Context, node *ast.Node,
	settings NoRestrictedGlobalsSettings) {

	if ctx.TypeChecker == nil {
		return
	}

	// The bare-reference half, which is upstream's `Program` listener.
	if _, restricted := settings.Globals[node.Text()]; restricted &&
		reference.IsValueReference(node) &&
		!rule.IsDeclaredInASourceFile(reference.ReadSymbol(ctx.TypeChecker, node)) {
		ctx.ReportNode(node, buildNoRestrictedGlobalsMessage(node.Text(), settings))
		return
	}

	// The `window.event` half, which is upstream's `Program:exit` listener under
	// `checkGlobalObject`. Anchored on the global object's identifier rather than on the member
	// access, because that is the node whose resolution decides whether this is the global at all.
	if !settings.CheckGlobalObject {
		return
	}
	if !isNoRestrictedGlobalsObjectName(node.Text(), settings) {
		return
	}
	// The global object has to RESOLVE, and not merely be unshadowed. Upstream reads
	// `getVariableByName(globalScope, name)` and returns early when it finds nothing, so a name
	// nobody declared is not the global object even under this option.
	//
	// The corpus states this as a three-way split that no single test could show:
	//
	//	window.foo()      checkGlobalObject, browser globals declared    REPORTS
	//	window.foo()      checkGlobalObject, no globals declared         clean
	//	globalThis.foo()  checkGlobalObject, no globals declared         REPORTS
	//
	// `globalThis` differs because eslint always knows it. Measured, our checker draws the same
	// line by itself: `globalThis` resolves to a symbol carrying zero declarations, while `window`
	// and `self` resolve to NO SYMBOL under a program whose lib does not declare them, and to a
	// declaration-file symbol under one that does. So "the symbol exists" reproduces upstream's
	// "the scope knows this name" without a scope table, and it reproduces the environment
	// sensitivity too: the same file answers differently under a DOM lib and a bare one, which is
	// what upstream's `languageOptions.globals` was expressing.
	//
	// A local binding is still not the global object either, which is upstream's four
	// `let window; window.foo()` cases, and `identifierIsShadowed` answers that half.
	if !noRestrictedGlobalsObjectResolves(ctx, node) || identifierIsShadowed(ctx, node) {
		return
	}
	reportNoRestrictedGlobalThroughObject(ctx, node, settings)
}

// reportNoRestrictedGlobalThroughObject walks out from the global object to the property it reads.
func reportNoRestrictedGlobalThroughObject(ctx rule.Context, node *ast.Node,
	settings NoRestrictedGlobalsSettings) {

	access := node.Parent
	if access == nil {
		return
	}
	// The identifier has to be the OBJECT of the access rather than the property. Without this,
	// `foo.window.bar()` reports, and upstream's four `foo.<globalObject>.bar()` cases are clean.
	if !isNoRestrictedGlobalsAccessObject(access, node) {
		return
	}

	// `window.window.foo()` names the same object twice, so the walk climbs while the property
	// being read is itself a global object name. Upstream spells this as a `while` over
	// `isSpecificMemberAccess(parent, null, globalObjectName)`, and its `window.window.foo()`,
	// `self.self.foo()`, `globalThis.globalThis.foo()` and `myGlobal.myGlobal.foo()` cases assert
	// all four spellings.
	for {
		name, known := property.AccessedName(access, property.Static)
		if !known || !isNoRestrictedGlobalsObjectName(name, settings) {
			break
		}
		next := access.Parent
		if next == nil || !isNoRestrictedGlobalsAccessObject(next, access) {
			break
		}
		access = next
	}

	// `property.Static` reads the dotted form and a string subscript alike, which is what makes
	// `window["foo"]` report alongside `window.foo`, and answers nothing for a computed access
	// through a variable, which is what keeps `window[foo]` clean.
	name, known := property.AccessedName(access, property.Static)
	if !known {
		return
	}
	if _, restricted := settings.Globals[name]; !restricted {
		return
	}

	// Upstream reports `parent.property`, the property node rather than the whole access, so the
	// finding points at `foo` in `window.foo` rather than at `window.foo`.
	target := noRestrictedGlobalsAccessedNameNode(access)
	if target == nil {
		return
	}
	ctx.ReportNode(target, buildNoRestrictedGlobalsMessage(name, settings))
}

// isNoRestrictedGlobalsAccessObject says whether `inner` is the object half of the access `outer`.
//
// Both a property access and an element access are accepted, since upstream reaches
// `window["foo"]` through the same path, and an optional access is the same node kind here with a
// question-dot token, so `window?.foo()` needs no separate arm.
func isNoRestrictedGlobalsAccessObject(outer *ast.Node, inner *ast.Node) bool {
	switch outer.Kind {
	case ast.KindPropertyAccessExpression:
		return outer.AsPropertyAccessExpression().Expression == inner
	case ast.KindElementAccessExpression:
		return outer.AsElementAccessExpression().Expression == inner
	}
	return false
}

// noRestrictedGlobalsAccessedNameNode returns the node upstream reports for a member access.
func noRestrictedGlobalsAccessedNameNode(access *ast.Node) *ast.Node {
	switch access.Kind {
	case ast.KindPropertyAccessExpression:
		return access.AsPropertyAccessExpression().Name()
	case ast.KindElementAccessExpression:
		return access.AsElementAccessExpression().ArgumentExpression
	}
	return nil
}

// noRestrictedGlobalsObjectResolves says whether the checker knows this name at all.
//
// This is upstream's `getVariableByName(globalScope, name)` returning something rather than
// undefined. A nil symbol means nobody declared the name, in the standard library or anywhere else,
// so it is not the global object and `checkGlobalObject` has nothing to walk.
//
// Deliberately NOT `identifierIsShadowed` inverted. That helper answers false for a nil symbol,
// which is the right answer to "is this shadowed" and the wrong one to "does this exist", and the
// two questions are both asked at this site for different reasons.
func noRestrictedGlobalsObjectResolves(ctx rule.Context, node *ast.Node) bool {
	if ctx.TypeChecker == nil {
		return false
	}
	return ctx.TypeChecker.GetSymbolAtLocation(node) != nil
}

// isNoRestrictedGlobalsObjectName says whether a name denotes the global object.
//
// The three built in are upstream's `GLOBAL_OBJECTS`, and the configured list is added to them
// rather than replacing them, which its `new Set([...GLOBAL_OBJECTS, ...userGlobalObjects])`
// states and its `myGlobal.foo()` cases assert alongside the `window.foo()` ones.
func isNoRestrictedGlobalsObjectName(name string, settings NoRestrictedGlobalsSettings) bool {
	if name == "globalThis" || name == "self" || name == "window" {
		return true
	}
	for _, configured := range settings.GlobalObjects {
		if name == configured {
			return true
		}
	}
	return false
}

// buildNoRestrictedGlobalsMessage renders whichever of the two messages the entry calls for.
func buildNoRestrictedGlobalsMessage(name string,
	settings NoRestrictedGlobalsSettings) rule.Message {

	custom := settings.Globals[name]
	if custom == "" {
		return rule.Message{
			Id: messageNoRestrictedGlobalsDefault.Id,
			Description: fmt.Sprintf("Unexpected use of '%s'. %s",
				name, messageNoRestrictedGlobalsDefault.Description),
		}
	}
	return rule.Message{
		Id: messageNoRestrictedGlobalsCustom.Id,
		Description: fmt.Sprintf("Unexpected use of '%s'. %s %s",
			name, custom, messageNoRestrictedGlobalsCustom.Description),
	}
}
