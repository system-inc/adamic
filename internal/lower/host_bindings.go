package lower

import (
	"path"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// A .ts file's ambient value declaration is a host binding where internal/load/host_bindings.json names
// it, and lowers by its row (#00qn5az, Adamic's ruling; docs/rulings.md's entry "A `.ts` file's ambient
// value declaration is a host binding"). A declarator matches a row only by its name, its file's
// basename, its keyword and its declared type's exact text, in a top-level `declare` statement of a .ts
// file that isn't exported; anything else stays refused as an ambient value declaration (ambient.go),
// and so does every .a file's. A read is matched to the row by the symbol it resolves to, never by its
// name, so a local that shadows the declared name is judged as the local.
//
// One IR feeds every target, so a row lowers only when its three columns are equal and none is pending
// its owner's clearance; any other row is not yet, named (hostBindingStopIn), at its declaration and at
// every use. The cleared rows lower together; onProfilerEvent and require are Absent on every target.
//
// Absent: typeof answers the constant "undefined", which every backend compares as it compares any
// string, so a guard's test is false at run time with no folding. The name itself is never read. A read
// is admitted only under one of three dominating shapes whose guard is `typeof X === "t"` with t not
// "undefined", or `typeof X !== "undefined"` (hostBindingGuard): the then-branch of if, the right of &&
// and the true arm of ?:. Everything else is not yet (hostBindingRefusal): ||, an early return, a negated
// guard, a guard bound to a local, a guard in a called function. The guarded branch, arm or operand is
// never lowered: it becomes the unreachable placeholder (hostBindingDeadRead), a loud compiler-bug panic,
// with neverValue's typed placeholder after it where a value is owed, never a fabricated initialized
// local. The placeholder rather than the else branch, because the test still runs as written, so the
// placeholder is reached only if the test itself was wrong, and then it stops loudly rather than taking
// a branch on a guess. Nothing in the program may define an Absent name: a write to it, and a
// write naming any row through the global object, `globalThis.<name>` or `globalThis[name]` by any name
// for it, is refused.
//
// Provided: every read routes through the row's model path to the host value nodeProcessPath already
// lowers (hostBindingRoute), as Node's own global's reads do, typeof and member reads included. A row
// whose model no path lowers stays not yet. The declaration is judged against the model export it
// stands for (hostBindingModelRefusal): its type may name only members the model's registration names,
// each typed as the model's is, none writable where the model's is read only. After the declaration
// every read is typed by the program's own interface, so nodeHostObject judges a later view of that
// interface as the model's host object (hostBindingViews). process and performance are Provided on
// every target. Only named optional absent members whose type admits undefined are admitted beyond
// the model; their constant reads are held exact by hostBindingAbsentMemberRefusal (#zqz01vh).

// hostBindingRows is the table lowering reads. Only tests replace it, to lower a row the real table
// doesn't hold yet; production never assigns it.
var hostBindingRows = load.HostBindings

// hostBindingNamed is the row for a name declared in a file, by the file's basename, or nil. A row is its
// name and its file together (internal/load/host_bindings.go): process has one in core.ts and one in
// tracing.ts, each with its own type text, and a declarator is judged only by its own file's row. A table
// that doesn't parse names nothing here, and ambientRefusal reports its error at the first ambient value
// declaration.
func hostBindingNamed(name, file string) *load.HostBinding {
	rows, err := hostBindingRows()
	if err != nil {
		return nil
	}
	for index := range rows {
		if rows[index].Name == name && rows[index].Declaration.File == file {
			return &rows[index]
		}
	}
	return nil
}

// hostBindingNameKnown reports whether a row in any file names a name. It answers for what has no
// declaring file of its own: an identifier before its symbol is resolved to the declaration
// (hostBindingOf, which then takes the row by the declaration's file), and a key written through the
// global object, where every row's name is the host's whichever file declares it.
func hostBindingNameKnown(name string) bool {
	rows, err := hostBindingRows()
	if err != nil {
		return false
	}
	for index := range rows {
		if rows[index].Name == name {
			return true
		}
	}
	return false
}

// hostBindingRow is the row a declarator is the declaration of, or nil.
func (l *lowering) hostBindingRow(declaration *ast.Node) *load.HostBinding {
	if declaration.Kind != ast.KindVariableDeclaration || !ast.IsIdentifier(declaration.Name()) || declaration.Initializer() != nil {
		return nil
	}
	list := declaration.Parent
	if list == nil || list.Kind != ast.KindVariableDeclarationList || list.Parent == nil || list.Parent.Kind != ast.KindVariableStatement {
		return nil
	}
	statement := list.Parent
	if statement.Parent == nil || statement.Parent.Kind != ast.KindSourceFile || !ast.HasSyntacticModifier(statement, ast.ModifierFlagsAmbient) || ast.HasSyntacticModifier(statement, ast.ModifierFlagsExport) {
		return nil
	}
	file := ast.GetSourceFileOfNode(declaration)
	name := file.FileName().AsString()
	if path.Ext(name) != ".ts" || strings.HasSuffix(name, ".d.ts") {
		return nil
	}
	row := hostBindingNamed(declaration.Name().Text(), path.Base(name))
	if row == nil {
		return nil
	}
	keyword := ""
	switch list.Flags & ast.NodeFlagsBlockScoped {
	case ast.NodeFlagsConst:
		keyword = "const"
	case ast.NodeFlagsLet:
		keyword = "let"
	}
	written := declaration.AsVariableDeclaration().Type
	if keyword != row.Declaration.Keyword || written == nil || file.Text()[scanner.GetTokenPosOfNode(written, file, false):written.End()] != row.Declaration.Type {
		return nil
	}
	return row
}

// hostBindingStatement reports whether every declarator of an ambient variable statement is a host
// binding's declaration, and if so, the stop for the first whose row doesn't lower, or nil.
func (l *lowering) hostBindingStatement(statement *ast.Node) (bool, error) {
	if _, err := hostBindingRows(); err != nil {
		return true, err
	}
	var stop error
	for _, declaration := range statement.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes {
		row := l.hostBindingRow(declaration)
		if row == nil {
			return false, nil
		}
		if stop != nil {
			continue
		}
		if reason := l.hostBindingStopIn(*row); reason != "" {
			stop = l.notYet(declaration, reason)
		} else if row.Native.Answer == load.HostBindingProvided {
			stop = l.hostBindingModelRefusal(declaration, *row)
		}
	}
	return true, stop
}

// hostBindingRoute is the path nodeProcessPath already lowers a Provided row's model as, the host
// value every read of the row goes to, or "" for a model no path lowers yet.
func hostBindingRoute(row load.HostBinding) string {
	switch row.Model.Path {
	case "node:perf_hooks":
		return "performance"
	case "node:process":
		return "process"
	}
	return ""
}

// hostBindingRouted reports whether a row lowers as Provided in this lowering's build, its reads routed
// (hostBindingRoute). It judges the row through hostBindingStopIn, so a build declared one CommonJS bundle
// never routes a row measured in ES module mode (#00qn5az).
func (l *lowering) hostBindingRouted(row load.HostBinding) bool {
	return l.hostBindingStopIn(row) == "" && row.Native.Answer == load.HostBindingProvided
}

// hostBindingModel is the type of a Provided row's model export: the value named as the row in the
// ambient module of its model path, which a Node library file declares; nil when there's none.
func (l *lowering) hostBindingModel(row load.HostBinding) *checker.Type {
	module := l.checker.TryFindAmbientModule(row.Model.Path)
	if exported := l.nodeModuleExport(module, row.Name); exported != nil {
		return l.checker.GetTypeOfSymbol(exported)
	}
	// The row's model only, never nodeModuleExport's other readers: a namespace import's member read
	// must not resolve to the export= value.
	if nodeLibraryModule(module) {
		return l.hostBindingExportEqualsModel(module, row.Name)
	}
	return nil
}

// nodeLibraryModule reports whether a Node library file declares the ambient module.
func nodeLibraryModule(module *ast.Symbol) bool {
	if module == nil {
		return false
	}
	for _, declaration := range module.Declarations {
		if declaration.Kind == ast.KindModuleDeclaration && load.IsNodeLibrary(ast.GetSourceFileOfNode(declaration)) {
			return true
		}
	}
	return false
}

// nodeModuleExport is the value an ambient module exports by name, resolved, where a Node library file
// declares the module; nil for another module or no such value. A namespace import of the module reads
// the same export (hostOriginExport).
func (l *lowering) nodeModuleExport(module *ast.Symbol, name string) *ast.Symbol {
	if !nodeLibraryModule(module) {
		return nil
	}
	for _, exported := range l.checker.GetExportsOfModule(module) {
		if exported.Name != name {
			continue
		}
		if resolved := l.resolvedExport(exported); resolved != nil && resolved.Flags&ast.SymbolFlagsValue != 0 {
			return resolved
		}
	}
	return nil
}

// hostBindingExportEqualsModel admits only the export= value's own name, never another row's.
// node:process exports its process value this way (#zqz01vh, Adamic and Lower, Nov 8 2026).
func (l *lowering) hostBindingExportEqualsModel(module *ast.Symbol, name string) *checker.Type {
	exported := module.Exports[ast.InternalSymbolNameExportEquals]
	if exported == nil {
		return nil
	}
	if resolved := l.resolvedExport(exported); resolved != nil && resolved.Name == name && resolved.Flags&ast.SymbolFlagsValue != 0 {
		return l.checker.GetTypeOfSymbol(resolved)
	}
	return nil
}

// hostBindingAbsentMembers names members Node's own model value lacks, including its prototypes.
// Each entry is held to Node in ES modules and CommonJS by TestHostBindingAbsentMembersHeldToNode.
var hostBindingAbsentMembers = map[string]map[string]bool{
	// browser: Adamic and Lower on #zqz01vh, Nov 8 2026. recordreplay, which tsc's sys.ts reads and the
	// prelude declares: Adamic on #b5w3ycg, Nov 7 2026, built as #b7phwab. Each reads as undefined on every
	// target; TestHostBindingAbsentMembersHeldToNode is the witness.
	"node:process.process": {"browser": true, "recordreplay": true},
}

// hostBindingModelRefusal judges a Provided row's declaration against its model export, since every
// read of the name routes to the model while the checker types it by the declaration (Compiler's
// condition on #yewmv46): the declared type may reach only members the model's registration names
// (nodeHostViewReach), each typed so the model's is assignable to it, and none writable where the
// model's is read only (widened). The declared type is taken without undefined: the host defines the
// name on every target, as Node does. A model export it can't find fails closed.
func (l *lowering) hostBindingModelRefusal(declaration *ast.Node, row load.HostBinding) error {
	export := row.Model.Path + "." + row.Name
	model := l.hostBindingModel(row)
	if model == nil {
		return l.notYet(declaration, "the host binding "+row.Name+"'s model export "+export+", which lowering can't find,")
	}
	view := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(declaration.Name()))
	subject := "the host binding " + row.Name + ", declared as " + l.checker.TypeToString(view) + ","
	absent := map[string]bool{}
	for _, property := range l.checker.GetPropertiesOfType(view) {
		if l.checker.GetPropertyOfType(model, property.Name) != nil {
			continue
		}
		if !hostBindingAbsentMembers[export][property.Name] || property.Flags&ast.SymbolFlagsOptional == 0 || !l.checker.IsTypeAssignableTo(l.checker.GetUndefinedType(), l.checker.GetTypeOfSymbol(property)) {
			return l.notYet(declaration, subject+" whose "+property.Name+" its model export "+export+" doesn't have,")
		}
		absent[property.Name] = true
	}
	if reached := l.nodeHostViewReach(model, view, 0, map[nodeHostViewPair]bool{}, false, absent); reached != "" {
		switch {
		case strings.HasPrefix(reached, nodeHostViewOpaque):
			return l.notYet(declaration, subject+" which hides its model export "+export+" behind any, unknown or an unbound type parameter,")
		case strings.HasPrefix(reached, nodeHostViewTooDeep):
			return l.notYet(declaration, subject+" which goes "+reached+" in its model export "+export+",")
		}
		return l.notYet(declaration, subject+" which reaches "+reached+" of its model export "+export+", a member no host registration names,")
	}
	// tsc never related the model to the declaration, so each member's type is related here, and the
	// first that the model's doesn't fit is named.
	for _, property := range l.checker.GetPropertiesOfType(view) {
		if absent[property.Name] {
			continue
		}
		member := l.checker.GetPropertyOfType(model, property.Name)
		if member == nil {
			return l.notYet(declaration, subject+" whose "+property.Name+" its model export "+export+" doesn't have,")
		}
		held, declared := l.checker.GetTypeOfSymbol(member), l.checker.GetTypeOfSymbol(property)
		if !l.checker.IsTypeAssignableTo(held, declared) {
			return l.notYet(declaration, subject+" whose "+property.Name+" is typed "+l.checker.TypeToString(declared)+" where its model export "+export+"'s is "+l.checker.TypeToString(held)+",")
		}
	}
	if !l.checker.IsTypeAssignableTo(model, view) {
		return l.notYet(declaration, subject+" which its model export "+export+", of type "+l.checker.TypeToString(model)+", isn't assignable to,")
	}
	if found := l.widened(model, view, map[[2]*checker.Type]bool{}); found != nil {
		switch {
		case found.readonlyField != "":
			return l.notYet(declaration, subject+" whose field "+found.readonlyField+" is writable where its model export "+export+"'s is read only,")
		case found.parameter:
			return l.notYet(declaration, subject+" which hands its model export "+export+" "+l.checker.TypeToString(found.target)+" where it takes "+l.checker.TypeToString(found.source)+",")
		}
		return l.notYet(declaration, subject+" which can write "+l.checker.TypeToString(found.target)+" where its model export "+export+" holds "+l.checker.TypeToString(found.source)+",")
	}
	return nil
}

// hostBindingViews is, for each Provided row the program declares, its declared type's symbol (each
// member's, for a union) by the type of the model export its reads route through, found once for the
// lowering. A row whose declaration hostBindingModelRefusal stops is held too: that stop ends the
// lowering at the declaration either way.
func (l *lowering) hostBindingViews() map[*ast.Symbol]*checker.Type {
	if l.hostBindingViewsKnown {
		return l.hostBindingViewTypes
	}
	l.hostBindingViewsKnown = true
	if l.program == nil {
		return nil
	}
	for _, file := range l.program.Files() {
		for _, statement := range file.Statements.Nodes {
			if statement.Kind != ast.KindVariableStatement || !ambientDeclaration(statement) {
				continue
			}
			for _, declaration := range statement.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes {
				row := l.hostBindingRow(declaration)
				if row == nil || !l.hostBindingRouted(*row) {
					continue
				}
				model := l.hostBindingModel(*row)
				if model == nil {
					continue
				}
				declared := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(declaration.Name()))
				members := []*checker.Type{declared}
				if declared.Flags()&checker.TypeFlagsUnion != 0 {
					members = declared.Types()
				}
				for _, member := range members {
					if symbol := member.Symbol(); symbol != nil {
						if l.hostBindingViewTypes == nil {
							l.hostBindingViewTypes = map[*ast.Symbol]*checker.Type{}
						}
						l.hostBindingViewTypes[symbol] = model
					}
				}
			}
		}
	}
	return l.hostBindingViewTypes
}

// hostBindingDeclared reports whether a statement declares only host bindings that lower: it declares
// nothing, since the host's names are the host's.
func (l *lowering) hostBindingDeclared(statement *ast.Node) bool {
	if statement.Kind != ast.KindVariableStatement || !ambientDeclaration(statement) {
		return false
	}
	matched, err := l.hostBindingStatement(statement)
	return matched && err == nil
}

// hostBindingRowStop is why a row doesn't lower in a build of the row's own measured module mode, said
// as what stage 0 can't lower yet, or "" when it does. Only hostBindingStopIn calls it: every judgment of
// a row goes through the build's module format first, so no caller can reach this ES-only answer by
// accident (Compiler's read of #00qn5az's require head).
func hostBindingRowStop(row load.HostBinding) string {
	var waits []string
	for _, target := range row.Targets() {
		if target.Column.Answer == load.HostBindingPending {
			waits = append(waits, possessive(target.Column.Owner)+" clearance of its "+target.Target+" column")
		}
	}
	if row.Model.Path == "" && row.Model.Task != "" {
		waits = append(waits, possessive(row.Model.Task)+" host model")
	}
	if len(waits) > 0 {
		said := waits[len(waits)-1]
		if len(waits) > 1 {
			said = strings.Join(waits[:len(waits)-1], ", ") + " and " + said
		}
		return "the host binding " + row.Name + ", pending " + said + ","
	}
	if row.Native.Answer != row.JavaScript.Answer || row.Native.Answer != row.WASM32.Answer {
		return "the host binding " + row.Name + ", whose targets answer differently (native " + string(row.Native.Answer) + ", javascript " + string(row.JavaScript.Answer) + ", wasm32 " + string(row.WASM32.Answer) + ") where one IR feeds every target, until an IR presence node is ruled,"
	}
	if row.Native.Answer == load.HostBindingProvided && hostBindingRoute(row) == "" {
		return "the host binding " + row.Name + "'s reads through its host model, " + row.Model.Path + ","
	}
	return ""
}

// hostBindingStopIn is why a row doesn't lower in this lowering's build, or "" when it does. A row's ES
// answer is measured as the oracle runs every program, an ES module, so it says nothing of a build that
// declares one CommonJS bundle (ModuleFormatCommonJSBundle), where Node may answer otherwise: typeof
// require is "function" there, and folding core.ts's require Absent would make isNodeLikeSystem silently
// false. Such a build judges a row by its CommonJS answer (#93w9rtz): a row with none recorded stops,
// measured in ES module mode only (#redxes2; WebAssembly's must on #00qn5az). So does a row whose two
// answers differ, even both defined, since its model was built and held to the ES answer (Platforms'
// must on #93w9rtz), and so does a row whose columns don't give its CommonJS answer, each at its
// declaration and at every use, with nothing folded.
func (l *lowering) hostBindingStopIn(row load.HostBinding) string {
	if l.moduleFormat == ModuleFormatCommonJSBundle {
		if row.Node.CommonJS == "" {
			return "the host binding " + row.Name + " in a build declared one CommonJS bundle, its row measured in ES module mode only (#redxes2),"
		}
		if row.Node.CommonJS != row.Node.TypeOf {
			return "the host binding " + row.Name + " in a build declared one CommonJS bundle, where Node's typeof is \"" + row.Node.CommonJS + "\" as a CommonJS script and \"" + row.Node.TypeOf + "\" as an ES module (#93w9rtz),"
		}
		if !hostBindingColumnsGive(row, row.Node.CommonJS) {
			return "the host binding " + row.Name + " in a build declared one CommonJS bundle, where Node's typeof is \"" + row.Node.CommonJS + "\" and its columns don't give that (#93w9rtz),"
		}
	}
	return hostBindingRowStop(row)
}

// hostBindingColumnsGive reports whether every cleared column of a row gives Node's typeof answer: Absent
// where it's "undefined", Provided where it's anything else. A pending column gives nothing yet, and
// hostBindingRowStop names it.
func hostBindingColumnsGive(row load.HostBinding, typeOf string) bool {
	for _, target := range row.Targets() {
		switch target.Column.Answer {
		case load.HostBindingAbsent:
			if typeOf != "undefined" {
				return false
			}
		case load.HostBindingProvided:
			if typeOf == "undefined" {
				return false
			}
		}
	}
	return true
}

// possessive is an owner's name with its possessive ending: Runtime's, Platforms'.
func possessive(owner string) string {
	if strings.HasSuffix(owner, "s") {
		return owner + "'"
	}
	return owner + "'s"
}

// hostBindingOf is the row an identifier reads, by the symbol it resolves to, and that symbol; nil when
// it reads anything else.
func (l *lowering) hostBindingOf(node *ast.Node) (*load.HostBinding, *ast.Symbol) {
	node = ast.SkipParentheses(node)
	if !ast.IsIdentifier(node) {
		return nil, nil
	}
	if !hostBindingNameKnown(node.Text()) {
		return nil, nil
	}
	var symbol *ast.Symbol
	switch parent := node.Parent; {
	case parent != nil && parent.Kind == ast.KindShorthandPropertyAssignment && parent.Name() == node:
		symbol = l.checker.GetShorthandAssignmentValueSymbol(parent)
	case parent != nil && parent.Kind == ast.KindExportSpecifier:
		symbol = l.checker.GetExportSpecifierLocalTargetSymbol(parent)
	default:
		symbol = l.checker.GetSymbolAtLocation(node)
	}
	if symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
		symbol = l.checker.GetAliasedSymbol(symbol)
	}
	if symbol == nil || len(symbol.Declarations) != 1 {
		return nil, nil
	}
	row := l.hostBindingRow(symbol.Declarations[0])
	if row == nil {
		return nil, nil
	}
	return row, symbol
}

// hostBindingAbsent reports whether a row lowers as Absent on every target in this lowering's build.
func (l *lowering) hostBindingAbsent(row load.HostBinding) bool {
	return l.hostBindingStopIn(row) == "" && row.Native.Answer == load.HostBindingAbsent
}

// hostBindingGuard is the symbol of the Absent host binding a condition's typeof test proves missing:
// `typeof X === "t"` with t not "undefined", or `typeof X !== "undefined"`, through parentheses. Any
// other condition, a negation, a local holding the test or a call that makes it, proves nothing.
func (l *lowering) hostBindingGuard(condition *ast.Node) *ast.Symbol {
	condition = ast.SkipParentheses(condition)
	if condition.Kind != ast.KindBinaryExpression {
		return nil
	}
	binary := condition.AsBinaryExpression()
	test, answer := ast.SkipParentheses(binary.Left), ast.SkipParentheses(binary.Right)
	if test.Kind != ast.KindTypeOfExpression || answer.Kind != ast.KindStringLiteral && answer.Kind != ast.KindNoSubstitutionTemplateLiteral {
		return nil
	}
	switch binary.OperatorToken.Kind {
	case ast.KindEqualsEqualsEqualsToken:
		if answer.Text() == "undefined" {
			return nil
		}
	case ast.KindExclamationEqualsEqualsToken:
		if answer.Text() != "undefined" {
			return nil
		}
	default:
		return nil
	}
	row, symbol := l.hostBindingOf(test.AsTypeOfExpression().Expression)
	if row == nil || !l.hostBindingAbsent(*row) {
		return nil
	}
	return symbol
}

// hostBindingDominated reports whether a read sits in the then-branch of an if, the right of an && or
// the true arm of a ?: whose guard proves its binding missing, within the read's own function.
func (l *lowering) hostBindingDominated(read *ast.Node, symbol *ast.Symbol) bool {
	for child, parent := read, read.Parent; parent != nil && !ast.IsFunctionLike(parent) && parent.Kind != ast.KindClassStaticBlockDeclaration; child, parent = parent, parent.Parent {
		var guard *ast.Node
		switch parent.Kind {
		case ast.KindIfStatement:
			if parent.AsIfStatement().ThenStatement == child {
				guard = parent.AsIfStatement().Expression
			}
		case ast.KindBinaryExpression:
			if parent.AsBinaryExpression().OperatorToken.Kind == ast.KindAmpersandAmpersandToken && parent.AsBinaryExpression().Right == child {
				guard = parent.AsBinaryExpression().Left
			}
		case ast.KindConditionalExpression:
			if parent.AsConditionalExpression().WhenTrue == child {
				guard = parent.AsConditionalExpression().Condition
			}
		}
		if guard != nil && l.hostBindingGuard(guard) == symbol {
			return true
		}
	}
	return false
}

// hostBindingRefusal judges every use of a host binding before lowering: a write is refused, a row that
// doesn't lower is not yet at each use, a Provided name's read is admitted, and an Absent name's read is
// admitted only at typeof or where a guard proves it dead (hostBindingDominated).
func (l *lowering) hostBindingRefusal(node *ast.Node) error {
	if !ast.IsIdentifier(node) {
		return nil
	}
	row, symbol := l.hostBindingOf(node)
	if row == nil || symbol.Declarations[0].Name() == node || ast.IsPartOfTypeNode(node) {
		return nil
	}
	if ast.IsAssignmentTarget(node) {
		return &Refused{Where: l.program.Where(node), What: "a write to the host binding " + row.Name, Fix: hostBindingWriteFix}
	}
	if stop := l.hostBindingStopIn(*row); stop != "" {
		return l.notYet(node, stop)
	}
	if row.Native.Answer == load.HostBindingProvided {
		return nil // its reads route through its model (nodeProcessPath), as Node's own global's do
	}
	operand := node
	for operand.Parent != nil && operand.Parent.Kind == ast.KindParenthesizedExpression {
		operand = operand.Parent
	}
	if operand.Parent != nil && operand.Parent.Kind == ast.KindTypeOfExpression {
		return nil
	}
	if !l.hostBindingDominated(node, symbol) {
		return l.notYet(node, "a read of "+row.Name+", a host binding absent on every target, outside the then-branch of if, the right of && or the true arm of ?: under a typeof guard that proves it dead,")
	}
	return nil
}

// hostBindingWriteFix is what a refused write to a host binding is told.
const hostBindingWriteFix = "a host binding is the host's to define, and nothing in the program may; keep the value in a variable of your own"

// hostBindingGlobalThisRefusal refuses, at the write, an assignment, update or delete through the global
// object whose key names a host binding or may: globalThis.onProfilerEvent = f would define an Absent
// name behind its folded typeof. The receiver is judged by its type, typeof globalThis, through whatever
// name reads it: globalThis, an alias of it, Node's global, globalThis.globalThis, and through a union or
// intersection holding it, a type parameter constrained to it (globalObject) or an operand that is it
// where the checker's type for the whole forgot it (globalObjectOperand). It runs before every
// other rule, so the stop is this ruling's, not a rule's that happens to meet the receiver or delete
// first. A receiver widened to another type (object, a record view) is no longer typeof globalThis; the
// stop that judges those, the global object as a value, is global_object.go's.
func (l *lowering) hostBindingGlobalThisRefusal(node *ast.Node) error {
	var target *ast.Node
	switch node.Kind {
	case ast.KindBinaryExpression:
		if ast.IsAssignmentOperator(node.AsBinaryExpression().OperatorToken.Kind) {
			target = node.AsBinaryExpression().Left
		}
	case ast.KindPrefixUnaryExpression:
		if operator := node.AsPrefixUnaryExpression().Operator; operator == ast.KindPlusPlusToken || operator == ast.KindMinusMinusToken {
			target = node.AsPrefixUnaryExpression().Operand
		}
	case ast.KindPostfixUnaryExpression:
		target = node.AsPostfixUnaryExpression().Operand
	case ast.KindDeleteExpression:
		target = node.AsDeleteExpression().Expression
	}
	if target == nil {
		return nil
	}
	target = ast.SkipParentheses(target)
	var receiver *ast.Node
	name, known := "", false
	switch target.Kind {
	case ast.KindPropertyAccessExpression:
		receiver, name, known = target.AsPropertyAccessExpression().Expression, target.AsPropertyAccessExpression().Name().Text(), true
	case ast.KindElementAccessExpression:
		receiver = target.AsElementAccessExpression().Expression
		if key := l.checker.GetTypeAtLocation(target.AsElementAccessExpression().ArgumentExpression); key.Flags()&checker.TypeFlagsStringLiteral != 0 {
			name, known = key.AsLiteralType().Value().(string)
		}
	default:
		return nil
	}
	receiver = ast.SkipParentheses(receiver)
	if !l.globalObject(receiver) && !l.globalObjectOperand(receiver) {
		return nil
	}
	if known && !hostBindingNameKnown(name) {
		return nil
	}
	what := "a write through globalThis to the host binding " + name
	if !known {
		what = "a write through globalThis whose key may name a host binding"
	}
	return &Refused{Where: l.program.Where(target), What: what, Fix: hostBindingWriteFix}
}

func init() {
	refusalRules.register(50, "hostBindingGlobalThisRefusal", (*lowering).hostBindingGlobalThisRefusal)
	refusalRules.register(2250, "hostBindingRefusal", (*lowering).hostBindingRefusal)
}

// hostBindingDeadRead is the unreachable placeholder for what a guard proves dead: the guard's test
// compares typeof's constant "undefined" at run time, so this runs only if that was wrong.
func (l *lowering) hostBindingDeadRead(symbol *ast.Symbol) ir.Statement {
	return ir.Panic{Message: ir.StringConstant{Index: l.constant("compiler bug: the typeof guard of " + symbol.Name + ", a host binding absent on every target, was taken")}}
}

// hostBindingTypeOf lowers typeof of a host binding: "undefined" for an Absent one. A Provided one's is
// left to the value its read routes to (nodeProcessPath), as typeof of Node's own global is.
func (l *lowering) hostBindingTypeOf(node *ast.Node) (ir.Expression, bool, error) {
	row, _ := l.hostBindingOf(node.AsTypeOfExpression().Expression)
	if row == nil {
		return nil, false, nil
	}
	if stop := l.hostBindingStopIn(*row); stop != "" {
		return nil, true, l.notYet(node, stop)
	}
	if row.Native.Answer == load.HostBindingProvided {
		return nil, false, nil
	}
	return ir.StringConstant{Index: l.constant("undefined")}, true, nil
}

// hostBindingAnd lowers guard && right, whose right a guard proves dead, as a value: the guard, and the
// placeholder where right would be. Only a boolean result is held; another stops as not yet.
func (l *lowering) hostBindingAnd(node *ast.Node) (ir.Expression, bool, error) {
	binary := node.AsBinaryExpression()
	if binary.OperatorToken.Kind != ast.KindAmpersandAmpersandToken {
		return nil, false, nil
	}
	symbol := l.hostBindingGuard(binary.Left)
	if symbol == nil {
		return nil, false, nil
	}
	left, err := l.expression(binary.Left)
	if err != nil {
		return nil, true, err
	}
	of, err := l.typeOf(node)
	if err != nil {
		return nil, true, err
	}
	if of != ir.Boolean {
		return nil, true, l.notYet(node, "the value of an && whose right a host binding's typeof guard proves dead, held as a "+typeName(of))
	}
	return ir.Binary{Operator: ir.And, Left: left, Right: ir.Effects{Body: []ir.Statement{l.hostBindingDeadRead(symbol)}, Result: unreachablePlaceholder(of)}}, true, nil
}

// hostBindingAndStatement lowers guard && right as a statement, as if (guard) right.
func (l *lowering) hostBindingAndStatement(node *ast.Node) ([]ir.Statement, bool, error) {
	binary := node.AsBinaryExpression()
	if binary.OperatorToken.Kind != ast.KindAmpersandAmpersandToken {
		return nil, false, nil
	}
	symbol := l.hostBindingGuard(binary.Left)
	if symbol == nil {
		return nil, false, nil
	}
	condition, err := l.condition(binary.Left)
	if err != nil {
		return nil, true, err
	}
	return []ir.Statement{ir.If{Condition: condition, Then: []ir.Statement{l.hostBindingDeadRead(symbol)}}}, true, nil
}

// hostBindingConditional lowers guard ? whenTrue : whenFalse, whose true arm the guard proves dead: the
// placeholder there, and the false arm as it is.
func (l *lowering) hostBindingConditional(node *ast.Node, condition ir.Expression, symbol *ast.Symbol) (ir.Expression, error) {
	of, err := l.typeOf(node)
	if err != nil {
		return nil, err
	}
	whenNot, err := l.expression(node.AsConditionalExpression().WhenFalse)
	if err != nil {
		return nil, err
	}
	whenTrue := ir.Effects{Body: []ir.Statement{l.hostBindingDeadRead(symbol)}, Result: unreachablePlaceholder(of)}
	return ir.Conditional{Condition: condition, WhenTrue: whenTrue, WhenNot: fit(whenNot, of), Of: of}, nil
}
