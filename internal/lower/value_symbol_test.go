package lower

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// A shorthand property's name is the property to GetSymbolAtLocation, and so to l.symbol, so a walk that
// asks whether an identifier reads or writes x by comparing l.symbol(identifier) with x's symbol misses
// { x } and ({ x } = other) (#3sdynvn). Five walks met it one at a time; l.valueSymbol answers the
// variable the name reads or writes. This holds the rule for the next walk: every function of this
// package that compares a symbol it took with symbol or GetSymbolAtLocation against another symbol, by
// == or !=, as a switch's tag, or as a map's key, looked up or recorded (a set recorded under the
// property's symbol, as assignedSymbol's was, misses the write as surely), is in symbolComparers with why
// it may and how many comparisons it makes. The count holds a listed function to the comparisons its why
// was read against, so one added to it later fails until it's read and counted too. It reads
// syntax only, as TestRefusalRulesReadTheirSlotThroughSlotType does. A symbol is taken by such a call, or
// by a local any assignment binds to one, matched by its name. A comparison with nil asks nothing, and
// one whose other side is valueSymbol's or GetShorthandAssignmentValueSymbol's answer has resolved the
// shorthand. The lowering's own tables, loweringTables, are keyed by declarations and used while
// lowering a name, which handles a shorthand itself (objectDestructuringAssignment, functionValue), so a
// key there isn't a walk asking whether a name reads x.
func TestReferenceWalksResolveShorthandNames(t *testing.T) {
	t.Parallel()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	files := token.NewFileSet()
	// compares is each function's comparisons, as file:line in function.
	compares := map[string][]string{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(files, entry.Name(), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, declaration := range file.Decls {
			function, isFunction := declaration.(*ast.FuncDecl)
			if !isFunction || function.Body == nil || function.Name.Name == "symbol" || function.Name.Name == "valueSymbol" {
				continue
			}
			name := symbolComparerName(function)
			for _, at := range symbolComparisons(function.Body) {
				position := files.Position(at)
				compares[name] = append(compares[name], position.Filename+":"+strconv.Itoa(position.Line)+" in "+name)
			}
		}
	}
	var unlisted, recounted []string
	count := 0
	for name, found := range compares {
		count += len(found)
		listed, isListed := symbolComparers[name]
		if !isListed {
			unlisted = append(unlisted, found...)
			continue
		}
		if listed.count != len(found) {
			sort.Strings(found)
			recounted = append(recounted, name+" is listed with "+strconv.Itoa(listed.count)+" and makes "+strconv.Itoa(len(found))+":\n"+strings.Join(found, "\n"))
		}
	}
	sort.Strings(unlisted)
	sort.Strings(recounted)
	if len(unlisted) > 0 {
		t.Errorf("a function compares a symbol it took with l.symbol or GetSymbolAtLocation, which answers a shorthand name's property rather than the variable it reads or writes (#3sdynvn); use l.valueSymbol, or list it in symbolComparers with why its identifier can't be a shorthand name:\n%s", strings.Join(unlisted, "\n"))
	}
	for _, recount := range recounted {
		t.Errorf("symbolComparers' count for a function differs from the comparisons it makes; read each one, and if its why holds them all, list the count it makes, or else use l.valueSymbol: %s", recount)
	}
	var stale []string
	for name := range symbolComparers {
		if len(compares[name]) == 0 {
			stale = append(stale, name)
		}
	}
	sort.Strings(stale)
	for _, name := range stale {
		t.Errorf("symbolComparers names %s, which no longer compares a symbol it took with l.symbol or GetSymbolAtLocation; take it out", name)
	}
	counted := 0
	for _, comparer := range symbolComparers {
		counted += comparer.count
	}
	t.Logf("%d comparisons in %d functions; symbolComparers lists %d in %d", count, len(compares), counted, len(symbolComparers))
}

// symbolComparerName is how symbolComparers names function: by its name, and a method of a type other
// than lowering by its type's name too, as parallelProof.reassigned.
func symbolComparerName(function *ast.FuncDecl) string {
	if function.Recv == nil || len(function.Recv.List) != 1 {
		return function.Name.Name
	}
	receiver := function.Recv.List[0].Type
	if pointer, isPointer := receiver.(*ast.StarExpr); isPointer {
		receiver = pointer.X
	}
	if named, isNamed := receiver.(*ast.Ident); isNamed && named.Name != "lowering" {
		return named.Name + "." + function.Name.Name
	}
	return function.Name.Name
}

// symbolCall is "Taken" for a call of symbol or GetSymbolAtLocation, "Resolved" for one of valueSymbol or
// GetShorthandAssignmentValueSymbol, and "" for anything else.
func symbolCall(expression ast.Expr) string {
	call, isCall := ast.Unparen(expression).(*ast.CallExpr)
	if !isCall {
		return ""
	}
	selector, isSelector := call.Fun.(*ast.SelectorExpr)
	if !isSelector {
		return ""
	}
	switch selector.Sel.Name {
	case "symbol", "GetSymbolAtLocation":
		return "Taken"
	case "valueSymbol", "GetShorthandAssignmentValueSymbol":
		return "Resolved"
	}
	return ""
}

// symbolComparisons is where body compares a symbol it took with symbol or GetSymbolAtLocation against
// another, as TestReferenceWalksResolveShorthandNames reads it.
func symbolComparisons(body *ast.BlockStmt) []token.Pos {
	// taken and resolved are the locals body binds to a symbolCall's answer, by name. One bound to both is
	// taken: it holds the property's symbol on some road.
	taken, resolved := map[string]bool{}, map[string]bool{}
	bind := func(name, value ast.Expr) {
		local, isLocal := name.(*ast.Ident)
		if !isLocal {
			return
		}
		switch symbolCall(value) {
		case "Taken":
			taken[local.Name] = true
		case "Resolved":
			resolved[local.Name] = true
		}
	}
	ast.Inspect(body, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.AssignStmt:
			if len(node.Lhs) == len(node.Rhs) {
				for position := range node.Lhs {
					bind(node.Lhs[position], node.Rhs[position])
				}
			}
		case *ast.ValueSpec:
			if len(node.Names) == len(node.Values) {
				for position := range node.Names {
					bind(node.Names[position], node.Values[position])
				}
			}
		}
		return true
	})
	isTaken := func(operand ast.Expr) bool {
		if local, isLocal := ast.Unparen(operand).(*ast.Ident); isLocal {
			return taken[local.Name]
		}
		return symbolCall(operand) == "Taken"
	}
	// settled is an operand that can't be a shorthand name's property: nil, or valueSymbol's answer.
	settled := func(operand ast.Expr) bool {
		if local, isLocal := ast.Unparen(operand).(*ast.Ident); isLocal {
			return local.Name == "nil" || resolved[local.Name] && !taken[local.Name]
		}
		return symbolCall(operand) == "Resolved"
	}
	// loweringTable is one of loweringTables on the lowering, l.locals or p.l.locals.
	loweringTable := func(table ast.Expr) bool {
		field, isField := ast.Unparen(table).(*ast.SelectorExpr)
		if !isField || !loweringTables[field.Sel.Name] {
			return false
		}
		switch owner := ast.Unparen(field.X).(type) {
		case *ast.Ident:
			return owner.Name == "l"
		case *ast.SelectorExpr:
			return owner.Sel.Name == "l"
		}
		return false
	}
	var found []token.Pos
	ast.Inspect(body, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.BinaryExpr:
			if (node.Op == token.EQL || node.Op == token.NEQ) && (isTaken(node.X) || isTaken(node.Y)) && !settled(node.X) && !settled(node.Y) {
				found = append(found, node.Pos())
			}
		case *ast.IndexExpr:
			if !loweringTable(node.X) && isTaken(node.Index) {
				found = append(found, node.Pos())
			}
		case *ast.SwitchStmt:
			if node.Tag != nil && isTaken(node.Tag) {
				found = append(found, node.Pos())
			}
		}
		return true
	})
	return found
}

// loweringTables are the lowering's own tables a symbol taken with l.symbol keys: each holds a local, a
// class, a function or a fact about one by its declaration's symbol, filled when the declaration is
// lowered and read when a name is, so TestReferenceWalksResolveShorthandNames doesn't count their keys.
// A table a walk fills for itself, as assignedSymbols, isn't one of them.
var loweringTables = map[string]bool{
	"alwaysUndefined":     true,
	"caught":              true,
	"classes":             true,
	"constEnumReach":      true,
	"derivedAncestors":    true,
	"functions":           true,
	"generics":            true,
	"locals":              true,
	"namespaceContainers": true,
	"namespaceValues":     true,
	"staticGlobals":       true,
	"statics":             true,
}

// symbolComparers are the functions that compare a symbol they took with l.symbol or GetSymbolAtLocation
// against another, each with how many comparisons it makes and why its identifier can't be a shorthand
// name, or how it resolves one itself (TestReferenceWalksResolveShorthandNames). A cycle guard's key is
// the symbol it follows, not a match.
var symbolComparers = map[string]symbolComparer{
	"assignedJustBefore":                {1, "a bare identifier on the left of =; a pattern write isn't recognized, so it refuses"},
	"bindReceiverRule":                  {1, "a bind call's receiver and its first argument, both expressions"},
	"boundKeyedLength":                  {3, "Reflect.get's and Object's call arguments, against the helper's own parameter declarations"},
	"boundLengthMethodEntry":            {2, "a cycle guard keyed by the symbol it follows through a const's initializer, never a reference match; the identifier it starts from is an expression, never a shorthand name; a rebinding is assignedSymbol's, through valueSymbol"},
	"boundLengthOriginal":               {2, "a cycle guard keyed by the symbol it follows through a const's initializer, never a reference match; the identifier it starts from is an expression, never a shorthand name; a rebinding is assignedSymbol's, through valueSymbol"},
	"boundSource":                       {2, "a cycle guard keyed by the symbol it follows through a const's initializer, never a reference match; the identifier it starts from is an expression, never a shorthand name"},
	"brandedChainCall":                  {1, "every identifier spelled as the binding is held at :501, so a shorthand { p } fails closed by its text"},
	"calledOnlyFromTopLevel":            {1, "every identifier spelled as the function must be a call's callee, so a shorthand { f } fails closed by its text"},
	"callsParameter":                    {1, "a call's callee"},
	"censusReturnProof":                 {3, "open: :90's scan for a write to a parameter is the audit's fails-open row 1, left to aaafbfbb (#ffkhsm0), which moves it to valueSymbol; :125 and :130 compare a return statement's expression"},
	"checkMove":                         {1, "proof.symbol, parallelProof's (parallel.go:545), resolves a shorthand name itself"},
	"checkMoveWork":                     {4, "proof.symbol, parallelProof's (parallel.go:545), resolves a shorthand name itself"},
	"checkedPredicateClosureCallee":     {1, "a call's callee, against the const holding the closure"},
	"checkedPredicateInferredPart":      {1, "a call's argument at the claim's index"},
	"checkedPredicateInferredReference": {2, "resolves a shorthand name itself at :383"},
	"classView":                         {1, "a type's symbol against a class declaration's own name"},
	"detachedOwnObjectParameter":        {1, "a shorthand storing the parameter is refused at :218, before :221 compares the rest"},
	"directBindMethodRead":              {1, "a bind call's receiver and its first argument, both expressions"},
	"directPredicateCallee":             {1, "a call's callee, against the predicate declaration's own name"},
	"enumMemberOrigin":                  {1, "an enum member's enum declaration name, against the enum it must belong to"},
	"enumTagWalk.exactKindCall":         {1, "a property access's receiver, parameter.kind"},
	"evolvingObject":                    {1, "a binary expression's left operand at :32; a shorthand write is :49's, through valueSymbol"},
	"factoryAnalysis.alias":             {1, "a property access's receiver, an assignment's right side, an initializer or a returned value; { obj } and { obj: obj } take the same road (the audit's note)"},
	"factoryAnalysis.statement":         {3, "records an alias at :252 and :266 under a variable declaration's own name; false flag at :307, which intersects two branches' alias sets by a range variable that shares :246's name"},
	"functionBind":                      {1, "a bind call's receiver and its first argument, both expressions"},
	"hostMemberValueBinding":            {3, "counts a shorthand reading the binding at :323; judge.seen is a cycle guard"},
	"hostMemberValueHeld":               {2, "a cycle guard keyed by the symbol it follows through a const's initializer, never a reference match; the identifier it starts from is an expression, never a shorthand name"},
	"hostOriginReads":                   {1, "hostOriginSymbol (method_presence.go:331) resolves a shorthand name; the GetSymbolAtLocation arm only adds matches"},
	"inspectorDoomedBinding":            {3, "resolves a shorthand write at :108; seen is a cycle guard"},
	"libraryMethod":                     {2, "a cycle guard keyed by the symbol it follows through a const's initializer, never a reference match; the identifier it starts from is an expression, never a shorthand name"},
	"namespaceCallable":                 {2, "a cycle guard keyed by the symbol it follows through a const's initializer, never a reference match; the identifier it starts from is an expression, never a shorthand name"},
	"nestedDeclarations":                {1, "two function declarations' own names, an overload and its implementation"},
	"neverRestParameter":                {2, "a cycle guard keyed by a parameter declaration's own symbol"},
	"nodeHostExactOptions":              {1, "resolves a shorthand name itself at :124"},
	"nodeRealpathReferenceRefusal":      {1, "resolves a shorthand name itself at :168"},
	"nodeWatchListener":                 {2, "a cycle guard keyed by the symbol it follows through a const's initializer, never a reference match; the identifier it starts from is an expression, never a shorthand name"},
	"onlyCalled":                        {1, "keeps its own identity, and takes valueSymbol for a shorthand name at :109 (#3sdynvn)"},
	"overloadCallbackArgument":          {1, "a call's argument, against the implementation's parameter declarations; a write is overloadCallbackAssigned's, through valueSymbol"},
	"overloadCallbackElement":           {1, "a for-of's iterated expression, against the implementation's parameter declarations; its other reads are overloadCallbackOnlyIterated's"},
	"overloadCallbackOnlyIterated":      {1, "a shorthand spelled as the parameter is held to be it at :388, by its text"},
	"overloadCallbackPresenceTest":      {1, "an equality's other operand, against the undefined global"},
	"overloadCallbackUses":              {1, "a shorthand spelled as the callback is held to be it at :117, by its text"},
	"overloadDefaultAbsent":             {1, "a ?? operand, against the implementation's parameter declarations; a write is predicateOverloadAssigned's, through valueSymbol"},
	"parallelMap":                       {2, "dedups the task's globals by proof.symbol, parallelProof's, which resolves a shorthand name (parallel.go:545)"},
	"parallelProof.functionValue":       {2, "a cycle guard by p.symbol, which resolves a shorthand name; the identifier is a call's argument or an initializer"},
	"phantomOverloadOnly":               {1, "a type alias's references; a shorthand name never names a type"},
	"plainEnumerableObject":             {3, "a binary expression's left operand at :153, and visiting is a cycle guard; a pattern write is :159's, through valueSymbol"},
	"predicateArguments":                {1, "a call's callee, against the predicate declaration's own name"},
	"predicateCallbackCallers":          {3, "walk-only: { f } escapes this walk, which skips declaration names, but a lying predicate value is checked at its own arrow and a view widening the predicate parameter is refused by invariance, shorthand and plain alike (probe J-bfqm1ws, #3sdynvn)"},
	"predicateFlowProof.literal":        {1, "an identifier spelled undefined, against the undefined global"},
	"predicateFlowProof.reference":      {1, "a test's narrowing operand: the test itself, a typeof's operand, a property access's receiver, an array predicate's argument, an instanceof's left side"},
	"predicateOverloadCallTests":        {2, "a call's callee and its claimed argument"},
	"predicateVerifier.expression":      {1, "an identifier read as a return path's test, against the path's known locals"},
	"predicateVerifier.sameParameter":   {1, "a test's operand: a bare truthiness test, a typeof's operand, a property access's receiver, a call's argument"},
	"predicateVerifier.statements":      {1, "records a path's local under a variable declaration's own name"},
	"reachedConstEnums":                 {1, "tsc refuses a const enum used as a value (TS2475), and a shorthand { E } is one, so no shorthand name reaches the set it records"},
	"recordInterveningMutation":         {2, "an element access's receiver, in a delete or an assignment's left side"},
	"recordOwnCheck":                    {1, "Object.hasOwn's arguments, against an element access's receiver and key"},
	"reflectGet":                        {1, "Reflect.get's receiver argument, against Object's argument"},
	"savedFactoryOrigin":                {2, "a cycle guard keyed by the symbol it follows through a const's initializer, never a reference match; the identifier it starts from is an expression, never a shorthand name"},
	"staticDeclaration":                 {1, "an extends clause's expression, against the base class declaration's own name"},
	"staticInitializationReads":         {2, "a property access's receiver at :379 and :470; its reference walk, :420, takes valueSymbol"},
	"storedGeneratorRead":               {1, "resolves a shorthand name itself at :124"},
	"throwStatement":                    {2, "a cycle guard keyed by the symbol it follows through a const's initializer, never a reference match; the identifier it starts from is an expression, never a shorthand name"},
}

// symbolComparer is a function's listing in symbolComparers: how many comparisons it makes, and why each
// may compare l.symbol's answer.
type symbolComparer struct {
	count int
	why   string
}
