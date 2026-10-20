// Package repowalk is adamiclint's rule against walking files outside the package without declaring them.
//
// Loom keys each test unit by its package's Go closure, its own testdata/, and the reads lines its planner keeps in
// cloud/fast-gate/executors.txt. A test that walks a directory outside its package, with no reads line, isn't rerun
// when what it walks changes, so a breaking change lands on a stale green and the red surfaces later on someone else's
// change. All four such reds in eight days were walks: on Oct 19 cmd/adamic's runtime_owners_test.go walked another
// package's _test.go files, and internal/dry's TestCohereMirrors walked the whole repository.
//
// A walk is filepath.WalkDir, filepath.Walk, fs.WalkDir, os.ReadDir, filepath.Glob, os.DirFS or parser.ParseDir. Its
// root leaves the package when it is a constant path that cleans to one starting with "..", what filepath.Join, Abs,
// Clean or Dir make of one, or the directory found by walking up from os.Getwd through filepath.Dir to the one that
// holds go.mod. The root is followed within the package through variables, parameters and results, since the real
// cases go through helpers: TestCohereMirrors hands "../.." to checkMirrors, which hands it to goFiles, which walks it.
// While the path is known it is read against the package's own directory, so "../../internal/oracle/testdata" walked
// from internal/oracle stays in the package; joined with something the rule can't read, the path is judged by what
// is known of it so far. A root the rule can't trace, such as t.TempDir() or a product directory from
// internal/buildcache, isn't flagged: the rule never guesses. A root that leaves the package only to open one fixed
// file is no walk, so it isn't flagged.
//
// The finding sits on the source, where the reach outward happens, since the walk itself may be in a helper many
// callers share: on the "../.." argument, or on the call to the function that walks up to the repository root. A
// // Reads: <glob> comment on that line or just above it declares the walk, one comment per glob, written as in the
// reads line. The rule doesn't hold the glob against executors.txt; Loom's planner does that from the comments.
package repowalk

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"path/filepath"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"
)

// Message says what goes wrong and how to declare it, so the author can act on the finding without looking anything up.
const Message = "walking files outside this package that its key doesn't name: a change to them won't rerun this " +
	"test; declare it with // Reads: <glob> and a matching reads line in cloud/fast-gate/executors.txt"

var Analyzer = &analysis.Analyzer{
	Name: "repowalk",
	Doc:  "flags a directory walk rooted outside the package's own directory with no // Reads: <glob> declaring it",
	Run:  run,
}

const escapePrefix = "// Reads:"

// deepestPath is how many elements a known path may grow to before the rest of it counts as unknown.
const deepestPath = 32

// walkSinks holds each walk by package path and function name, with the index of the argument it walks. fs.WalkDir
// walks its file system, whose root os.DirFS carries; its own root argument is a path inside that file system.
var walkSinks = map[string]map[string]int{
	"path/filepath": {"WalkDir": 0, "Walk": 0, "Glob": 0},
	"io/fs":         {"WalkDir": 0},
	"os":            {"ReadDir": 0, "DirFS": 0},
	"go/parser":     {"ParseDir": 1},
}

// pathCarriers are the path/filepath functions whose first result names the same place as their first argument, so a
// root that leaves the package still leaves it after them. filepath.Dir moves the place up, and is read apart.
var pathCarriers = map[string]bool{
	"Abs": true, "Clean": true, "EvalSymlinks": true, "FromSlash": true, "ToSlash": true,
}

// origin is where a followed path comes from. A source that leaves the package has the position its // Reads:
// comment answers; a parameter is resolved to the arguments at each call. A root found by walking up to go.mod is
// moved, at each call of the function that walks up (walksUpIn), to that call, since the call is where a reader
// sees the reach outward.
//
// path is what is known of the path: for a source, the path as the program builds it, relative to the package's
// directory, and for a parameter, what is joined onto it. opaque says something the rule can't read was joined on,
// so path is only the part before it.
type origin struct {
	position  token.Pos
	parameter *types.Var
	walksUpIn *types.Func
	path      string
	opaque    bool
}

type origins map[origin]bool

// add puts every origin of more into into, and says whether into grew.
func (into origins) add(more origins) bool {
	grew := false
	for each := range more {
		if !into[each] {
			into[each] = true
			grew = true
		}
	}
	return grew
}

// assigned is one value given to a variable: an expression, or one result of a call that returns several.
type assigned struct {
	expression ast.Expr
	result     int
}

// function is what the rule knows of one function declared in the package.
type function struct {
	object      *types.Func
	declaration *ast.FuncDecl
	signature   *types.Signature
	// returns are the function's own return statements, a function literal's left out, and calls are every call in
	// its body, a function literal's included, since a closure walking what its function holds is the same walk.
	returns []*ast.ReturnStmt
	calls   []*ast.CallExpr
	// results are the origins of each result, and walked the origins that reach a walk, each with the walk's position.
	results []origins
	walked  map[origin]token.Pos
}

// walker follows roots through one package until nothing more is learned.
type walker struct {
	pass *analysis.Pass
	// directory is the package's own directory, which every known path is read against.
	directory   string
	functions   map[*types.Func]*function
	assignments map[*types.Var][]assigned
	variables   map[*types.Var]origins
	owners      map[*types.Var]*function
	// unreadable holds the variables written some way the rule doesn't read as a value, by a range, an operator
	// assignment, its address or a declaration with no value, and every parameter and result. constants holds the
	// strings a variable is only ever given as constants, so the path joined from it is known.
	unreadable map[*types.Var]bool
	constants  map[*types.Var][]string
}

func run(pass *analysis.Pass) (any, error) {
	if len(pass.Files) == 0 {
		return nil, nil
	}
	walker := &walker{
		pass:        pass,
		directory:   filepath.Dir(pass.Fset.File(pass.Files[0].Pos()).Name()),
		functions:   map[*types.Func]*function{},
		assignments: map[*types.Var][]assigned{},
		variables:   map[*types.Var]origins{},
		owners:      map[*types.Var]*function{},
		unreadable:  map[*types.Var]bool{},
		constants:   map[*types.Var][]string{},
	}
	walker.collect()
	// Every set only grows, and every origin is a position in the package or one of its parameters, with a path
	// joined from the package's own constants, so this ends.
	for walker.settle() {
	}
	walker.report()
	return nil, nil
}

// collect reads the package once: its functions, their parameters, returns and calls, every assignment to a variable
// of the package, and the walk-up loops that find the repository root.
func (walker *walker) collect() {
	for _, file := range walker.pass.Files {
		for _, declaration := range file.Decls {
			declaration, ok := declaration.(*ast.FuncDecl)
			if !ok || declaration.Body == nil {
				continue
			}
			object, ok := walker.pass.TypesInfo.Defs[declaration.Name].(*types.Func)
			if !ok {
				continue
			}
			signature, ok := object.Type().(*types.Signature)
			if !ok {
				continue
			}
			declared := &function{
				object:      object,
				declaration: declaration,
				signature:   signature,
				results:     make([]origins, signature.Results().Len()),
				walked:      map[origin]token.Pos{},
			}
			for index := range declared.results {
				declared.results[index] = origins{}
				walker.unreadable[signature.Results().At(index)] = true
			}
			for index := range signature.Params().Len() {
				parameter := signature.Params().At(index)
				walker.owners[parameter] = declared
				walker.unreadable[parameter] = true
				walker.variables[parameter] = origins{{parameter: parameter}: true}
			}
			walker.functions[object] = declared
			ast.Inspect(declaration.Body, func(node ast.Node) bool {
				switch node := node.(type) {
				case *ast.CallExpr:
					declared.calls = append(declared.calls, node)
				case *ast.ReturnStmt:
					declared.returns = append(declared.returns, node)
				case *ast.FuncLit:
					// A function literal's returns are its own, so they are read apart from the function's.
					ast.Inspect(node.Body, func(inner ast.Node) bool {
						if call, ok := inner.(*ast.CallExpr); ok {
							declared.calls = append(declared.calls, call)
						}
						return true
					})
					return false
				}
				return true
			})
			walker.findWalkUps(declared)
		}
		ast.Inspect(file, func(node ast.Node) bool {
			switch node := node.(type) {
			case *ast.AssignStmt:
				for index, target := range node.Lhs {
					if node.Tok != token.ASSIGN && node.Tok != token.DEFINE {
						walker.markUnreadable(target)
						continue
					}
					walker.assign(target, node.Rhs, len(node.Lhs), index)
				}
			case *ast.ValueSpec:
				for index, name := range node.Names {
					walker.assign(name, node.Values, len(node.Names), index)
				}
			case *ast.RangeStmt:
				walker.markUnreadable(node.Key)
				walker.markUnreadable(node.Value)
			case *ast.UnaryExpr:
				if node.Op == token.AND {
					walker.markUnreadable(node.X)
				}
			}
			return true
		})
	}
	for variable, values := range walker.assignments {
		if walker.unreadable[variable] {
			continue
		}
		var texts []string
		for _, value := range values {
			text, ok := walker.constantString(value.expression)
			if !ok || value.result != -1 {
				texts = nil
				break
			}
			texts = append(texts, text)
		}
		if texts != nil {
			walker.constants[variable] = texts
		}
	}
}

func (walker *walker) markUnreadable(target ast.Expr) {
	if target == nil {
		return
	}
	if variable := walker.variableOf(target); variable != nil {
		walker.unreadable[variable] = true
	}
}

// assign records the value the index'th of count targets is given from values: its own, or one result of the call
// that is the only value.
func (walker *walker) assign(target ast.Expr, values []ast.Expr, count, index int) {
	variable := walker.variableOf(target)
	if variable == nil {
		return
	}
	switch {
	case len(values) == count:
		walker.assignments[variable] = append(walker.assignments[variable], assigned{expression: values[index], result: -1})
	case len(values) == 1:
		walker.assignments[variable] = append(walker.assignments[variable], assigned{expression: values[0], result: index})
	default:
		// A declaration with no value holds the zero value, which is no path the rule reads.
		walker.unreadable[variable] = true
	}
}

// variableOf is the package's own variable an expression names, or nil: a field, or another package's variable,
// is never followed.
func (walker *walker) variableOf(expression ast.Expr) *types.Var {
	identifier, ok := ast.Unparen(expression).(*ast.Ident)
	if !ok {
		return nil
	}
	variable, ok := walker.pass.TypesInfo.ObjectOf(identifier).(*types.Var)
	if !ok || variable.IsField() || variable.Pkg() != walker.pass.Pkg {
		return nil
	}
	return variable
}

// findWalkUps finds the loop that walks up from the working directory to the repository root: a variable on a cycle
// of assignments through filepath.Dir, reached from os.Getwd's result, whose cycle joins a variable with "go.mod".
// Each variable on the cycle is that root from the start, since the loop may return or walk any of them.
func (walker *walker) findWalkUps(declared *function) {
	// sources holds, for each variable, the variables it is given and whether through filepath.Dir.
	type edge struct {
		from             *types.Var
		throughDirectory bool
	}
	sources := map[*types.Var][]edge{}
	workingDirectories := map[*types.Var]token.Pos{}
	joinedWithGoMod := map[*types.Var]bool{}
	record := func(target ast.Expr, value ast.Expr, result int) {
		variable := walker.variableOf(target)
		if variable == nil {
			return
		}
		call, isCall := ast.Unparen(value).(*ast.CallExpr)
		if isCall && result == 0 && walker.isFunction(call, "os", "Getwd") {
			workingDirectories[variable] = call.Pos()
			return
		}
		if result != -1 {
			return
		}
		if isCall && walker.isFunction(call, "path/filepath", "Dir") && len(call.Args) == 1 {
			if from := walker.variableOf(call.Args[0]); from != nil {
				sources[variable] = append(sources[variable], edge{from: from, throughDirectory: true})
			}
			return
		}
		if from := walker.variableOf(value); from != nil {
			sources[variable] = append(sources[variable], edge{from: from})
		}
	}
	ast.Inspect(declared.declaration.Body, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.AssignStmt:
			for index, target := range node.Lhs {
				switch {
				case len(node.Lhs) == len(node.Rhs):
					record(target, node.Rhs[index], -1)
				case len(node.Rhs) == 1:
					record(target, node.Rhs[0], index)
				}
			}
		case *ast.ValueSpec:
			for index, name := range node.Names {
				switch {
				case len(node.Names) == len(node.Values):
					record(name, node.Values[index], -1)
				case len(node.Values) == 1:
					record(name, node.Values[0], index)
				}
			}
		case *ast.CallExpr:
			if !walker.isFunction(node, "path/filepath", "Join") {
				return true
			}
			namesGoMod := false
			for _, argument := range node.Args {
				if value, ok := walker.constantString(argument); ok && filepath.Base(value) == "go.mod" {
					namesGoMod = true
				}
			}
			if !namesGoMod {
				return true
			}
			for _, argument := range node.Args {
				if variable := walker.variableOf(argument); variable != nil {
					joinedWithGoMod[variable] = true
				}
			}
		}
		return true
	})
	// reaches is every variable start is given, directly or through others, with whether some step went through
	// filepath.Dir; a variable reaching itself through filepath.Dir is on the walk-up cycle.
	reaches := func(start *types.Var) map[*types.Var]bool {
		type state struct {
			variable         *types.Var
			throughDirectory bool
		}
		seen := map[state]bool{}
		pending := []state{{variable: start}}
		reached := map[*types.Var]bool{}
		for len(pending) > 0 {
			current := pending[len(pending)-1]
			pending = pending[:len(pending)-1]
			for _, each := range sources[current.variable] {
				next := state{variable: each.from, throughDirectory: current.throughDirectory || each.throughDirectory}
				if seen[next] {
					continue
				}
				seen[next] = true
				reached[next.variable] = reached[next.variable] || next.throughDirectory
				pending = append(pending, next)
			}
		}
		return reached
	}
	for variable := range sources {
		reached := reaches(variable)
		if !reached[variable] {
			continue
		}
		// The cycle is every variable that reaches back to this one; the go.mod test may be on any of them.
		readsGoMod := joinedWithGoMod[variable]
		for other := range reached {
			if _, onCycle := reaches(other)[variable]; onCycle && joinedWithGoMod[other] {
				readsGoMod = true
			}
		}
		// The loop starts from os.Getwd's result, given to this variable or one it reaches; the earliest is the
		// position, so the finding is stable.
		position := token.NoPos
		for other := range reached {
			if found, ok := workingDirectories[other]; ok && (position == token.NoPos || found < position) {
				position = found
			}
		}
		if found, ok := workingDirectories[variable]; ok && (position == token.NoPos || found < position) {
			position = found
		}
		if !readsGoMod || position == token.NoPos {
			continue
		}
		if walker.variables[variable] == nil {
			walker.variables[variable] = origins{}
		}
		// Where the repository root is from the package isn't written anywhere, so the path is unknown from the start.
		walker.variables[variable][origin{position: position, walksUpIn: declared.object, opaque: true}] = true
	}
}

// settle reads every assignment, return and call once more with what is known so far, and says whether anything grew.
func (walker *walker) settle() bool {
	grew := false
	for variable, values := range walker.assignments {
		known := walker.variables[variable]
		if known == nil {
			known = origins{}
			walker.variables[variable] = known
		}
		packageLevel := variable.Parent() == walker.pass.Pkg.Scope()
		for _, value := range values {
			found := walker.evaluate(value.expression, value.result)
			for each := range found {
				// A package variable belongs to no call, so a parameter it is given can't be resolved.
				if packageLevel && each.parameter != nil {
					delete(found, each)
				}
			}
			grew = known.add(found) || grew
		}
	}
	for _, declared := range walker.functions {
		grew = walker.settleReturns(declared) || grew
		for _, call := range declared.calls {
			grew = walker.settleCall(declared, call) || grew
		}
	}
	return grew
}

func (walker *walker) settleReturns(declared *function) bool {
	grew := false
	results := declared.signature.Results()
	for _, statement := range declared.returns {
		switch {
		case len(statement.Results) == 0:
			// A bare return gives the named results as they stand.
			for index := range results.Len() {
				grew = declared.results[index].add(walker.variables[results.At(index)]) || grew
			}
		case len(statement.Results) == results.Len():
			for index, result := range statement.Results {
				grew = declared.results[index].add(walker.evaluate(result, -1)) || grew
			}
		case len(statement.Results) == 1:
			for index := range results.Len() {
				grew = declared.results[index].add(walker.evaluate(statement.Results[0], index)) || grew
			}
		}
	}
	return grew
}

// settleCall records what a call in declared walks: the root of a walk itself, or the arguments a function of the
// package hands to its own walks.
func (walker *walker) settleCall(declared *function, call *ast.CallExpr) bool {
	grew := false
	walked := func(found origins, position token.Pos) {
		for each := range found {
			// Only this function's own parameters are resolved at its calls; anything else isn't followed.
			if each.parameter != nil && walker.owners[each.parameter] != declared {
				continue
			}
			if existing, ok := declared.walked[each]; !ok || position < existing {
				declared.walked[each] = position
				grew = true
			}
		}
	}
	if index, ok := walker.sinkArgument(call); ok {
		walked(walker.evaluate(call.Args[index], -1), call.Pos())
		return grew
	}
	callee := walker.packageFunction(call)
	if callee == nil {
		return grew
	}
	for each, position := range callee.walked {
		if each.parameter == nil {
			continue
		}
		walked(walker.resolve(call, callee, each), position)
	}
	return grew
}

// resolve is a parameter origin of callee as one call gives it: the origins of the arguments, with what callee
// joined onto the parameter joined onto them.
func (walker *walker) resolve(call *ast.CallExpr, callee *function, parameter origin) origins {
	found := origins{}
	for _, argument := range walker.arguments(call, callee, parameter.parameter) {
		found.add(walker.evaluate(argument, -1))
	}
	if parameter.opaque {
		return walker.joined(found, nil)
	}
	return walker.joined(found, []string{parameter.path})
}

// evaluate is where the value of an expression comes from, or of one result of it when result isn't -1.
func (walker *walker) evaluate(expression ast.Expr, result int) origins {
	expression = ast.Unparen(expression)
	if result != -1 {
		call, ok := expression.(*ast.CallExpr)
		if !ok {
			return origins{}
		}
		return walker.callResult(call, result)
	}
	if value, ok := walker.constantString(expression); ok {
		return walker.source(expression.Pos(), value)
	}
	switch expression := expression.(type) {
	case *ast.Ident:
		if variable := walker.variableOf(expression); variable != nil {
			found := origins{}
			found.add(walker.variables[variable])
			return found
		}
	case *ast.BinaryExpr:
		// A path built by concatenation starts where its left side does; a right side that starts a new path element
		// joins it on, and anything else makes the rest of the path unknown.
		if expression.Op != token.ADD {
			return origins{}
		}
		left := walker.evaluate(expression.X, -1)
		found := origins{}
		values := walker.stringValues(expression.Y)
		if values == nil {
			return walker.joined(left, nil)
		}
		for _, value := range values {
			element, ok := strings.CutPrefix(value, "/")
			if !ok {
				found.add(walker.joined(left, nil))
				continue
			}
			found.add(walker.joined(left, []string{element}))
		}
		return found
	case *ast.CallExpr:
		if walker.pass.TypesInfo.Types[expression.Fun].IsType() && len(expression.Args) == 1 {
			return walker.evaluate(expression.Args[0], -1)
		}
		return walker.callResult(expression, 0)
	}
	return origins{}
}

// callResult is where one result of a call comes from: a path function's argument, or a package function's result
// with its parameters resolved to this call's arguments.
func (walker *walker) callResult(call *ast.CallExpr, result int) origins {
	callee := typeutil.StaticCallee(walker.pass.TypesInfo, call)
	if callee == nil {
		return origins{}
	}
	callee = callee.Origin()
	if isPackageFunction(callee, "path/filepath") && result == 0 && len(call.Args) > 0 {
		switch {
		case callee.Name() == "Join":
			return walker.join(call)
		case callee.Name() == "Dir":
			return walker.parent(walker.evaluate(call.Args[0], -1))
		case pathCarriers[callee.Name()]:
			return walker.evaluate(call.Args[0], -1)
		}
	}
	// os.DirFS's file system is rooted where its argument is, so fs.WalkDir over it walks there.
	if isPackageFunction(callee, "os") && callee.Name() == "DirFS" && result == 0 {
		return walker.evaluate(call.Args[0], -1)
	}
	declared := walker.functions[callee]
	if declared == nil || result >= len(declared.results) {
		return origins{}
	}
	found := origins{}
	for each := range declared.results[result] {
		switch {
		case each.parameter != nil:
			found.add(walker.resolve(call, declared, each))
		case each.walksUpIn == callee:
			each.position = call.Pos()
			each.walksUpIn = nil
			found[each] = true
		default:
			found[each] = true
		}
	}
	return found
}

// join is where filepath.Join's result comes from. Its leading constant arguments are joined as the call would join
// them, so filepath.Join("..", "..") leaves the package though neither argument does alone; past them, the path
// starts where its first argument does, and each argument after is joined on.
func (walker *walker) join(call *ast.CallExpr) origins {
	if call.Ellipsis.IsValid() {
		return origins{}
	}
	var leading []string
	for _, argument := range call.Args {
		value, ok := walker.constantString(argument)
		if !ok {
			break
		}
		leading = append(leading, value)
	}
	var found origins
	var rest []ast.Expr
	if len(leading) > 0 {
		found = walker.source(call.Pos(), filepath.Join(leading...))
		rest = call.Args[len(leading):]
	} else {
		found = walker.evaluate(call.Args[0], -1)
		rest = call.Args[1:]
	}
	for _, argument := range rest {
		values := walker.stringValues(argument)
		if values == nil {
			found = walker.joined(found, nil)
			continue
		}
		joined := origins{}
		for _, value := range values {
			joined.add(walker.joined(found, []string{value}))
		}
		found = joined
	}
	return found
}

// source is the origin of a constant path at position, if the path leaves the package.
func (walker *walker) source(position token.Pos, path string) origins {
	if !walker.leavesPackage(path) {
		return origins{}
	}
	return origins{{position: position, path: path}: true}
}

// joined is found with elements joined onto each path, or, for nil elements, with something the rule can't read
// joined on. A source whose known path comes back into the package is dropped; one joined with something unknown is
// judged by its known part, which still leaves the package, since a source that came back was dropped already.
func (walker *walker) joined(found origins, elements []string) origins {
	result := origins{}
	for each := range found {
		switch {
		case each.opaque:
		case elements == nil:
			each.opaque = true
		default:
			each.path = filepath.Join(append([]string{each.path}, elements...)...)
			if each.parameter == nil && !walker.leavesPackage(each.path) {
				continue
			}
			// A function that calls itself with a path joined onto its parameter would join on forever, so a path
			// this deep is judged by what is known of it, as if the rest were unknown.
			if strings.Count(each.path, "/") >= deepestPath {
				each.opaque = true
			}
		}
		result[each] = true
	}
	return result
}

// parent is found after filepath.Dir: the directory holding each path, read as the program reads it.
func (walker *walker) parent(found origins) origins {
	result := origins{}
	for each := range found {
		switch {
		case each.opaque:
		case each.parameter != nil && each.path == "":
			// The parameter's own parent is no path joined onto it.
			each.opaque = true
		default:
			each.path = filepath.Dir(each.path)
			if each.parameter != nil && each.path == "." {
				each.path = ""
			}
			if each.parameter == nil && !walker.leavesPackage(each.path) {
				continue
			}
		}
		result[each] = true
	}
	return result
}

// stringValues are the strings an expression may be: a constant, or a variable only ever given constants, or nil.
func (walker *walker) stringValues(expression ast.Expr) []string {
	if value, ok := walker.constantString(expression); ok {
		return []string{value}
	}
	if variable := walker.variableOf(expression); variable != nil {
		return walker.constants[variable]
	}
	return nil
}

// arguments are the expressions a call gives a parameter: one, or all of a variadic parameter's unless the call
// spreads a slice into it.
func (walker *walker) arguments(call *ast.CallExpr, callee *function, parameter *types.Var) []ast.Expr {
	parameters := callee.signature.Params()
	if selector, ok := ast.Unparen(call.Fun).(*ast.SelectorExpr); ok {
		// A method expression takes its receiver as the first argument, which the rule doesn't follow.
		if selection, ok := walker.pass.TypesInfo.Selections[selector]; ok && selection.Kind() == types.MethodExpr {
			return nil
		}
	}
	for index := range parameters.Len() {
		if parameters.At(index) != parameter {
			continue
		}
		if callee.signature.Variadic() && index == parameters.Len()-1 && !call.Ellipsis.IsValid() {
			if index >= len(call.Args) {
				return nil
			}
			return call.Args[index:]
		}
		if index < len(call.Args) {
			return call.Args[index : index+1]
		}
	}
	return nil
}

// packageFunction is the function of this package a call reaches, or nil.
func (walker *walker) packageFunction(call *ast.CallExpr) *function {
	callee := typeutil.StaticCallee(walker.pass.TypesInfo, call)
	if callee == nil {
		return nil
	}
	return walker.functions[callee.Origin()]
}

// sinkArgument is the index of the argument a walk call walks, matched by the function object, so a function of the
// same name in another package is never mistaken for it.
func (walker *walker) sinkArgument(call *ast.CallExpr) (int, bool) {
	callee := typeutil.StaticCallee(walker.pass.TypesInfo, call)
	if callee == nil || callee.Pkg() == nil {
		return 0, false
	}
	walks, ok := walkSinks[callee.Pkg().Path()]
	if !ok || !isPackageFunction(callee, callee.Pkg().Path()) {
		return 0, false
	}
	index, ok := walks[callee.Name()]
	if !ok || index >= len(call.Args) {
		return 0, false
	}
	return index, true
}

// isFunction says call reaches the package-level function name of the package at packagePath.
func (walker *walker) isFunction(call *ast.CallExpr, packagePath, name string) bool {
	callee := typeutil.StaticCallee(walker.pass.TypesInfo, call)
	return callee != nil && isPackageFunction(callee, packagePath) && callee.Name() == name
}

// isPackageFunction says callee is a function, not a method, of the package at packagePath.
func isPackageFunction(callee *types.Func, packagePath string) bool {
	if callee.Pkg() == nil || callee.Pkg().Path() != packagePath {
		return false
	}
	signature, ok := callee.Type().(*types.Signature)
	return ok && signature.Recv() == nil
}

func (walker *walker) constantString(expression ast.Expr) (string, bool) {
	value := walker.pass.TypesInfo.Types[expression].Value
	if value == nil || value.Kind() != constant.String {
		return "", false
	}
	return constant.StringVal(value), true
}

// leavesPackage says a relative path, read from the package's directory, names a place outside it. A path that climbs
// out and back in, such as "../../internal/oracle/testdata" read from internal/oracle, stays in the package.
func (walker *walker) leavesPackage(path string) bool {
	if filepath.IsAbs(path) {
		return false
	}
	relative, err := filepath.Rel(walker.directory, filepath.Join(walker.directory, path))
	if err != nil {
		return false
	}
	relative = filepath.ToSlash(relative)
	return relative == ".." || strings.HasPrefix(relative, "../")
}

// report flags each source that reaches a walk once, on the source, with the walk it reaches as related.
func (walker *walker) report() {
	declaredLines := map[*token.File]map[int]bool{}
	for _, file := range walker.pass.Files {
		declaredLines[walker.pass.Fset.File(file.Pos())] = readsLines(walker.pass.Fset, file)
	}
	walks := map[token.Pos]token.Pos{}
	for _, declared := range walker.functions {
		for each, walk := range declared.walked {
			if each.parameter != nil {
				continue
			}
			if existing, ok := walks[each.position]; !ok || walk < existing {
				walks[each.position] = walk
			}
		}
	}
	for source, walk := range walks {
		file := walker.pass.Fset.File(source)
		if file == nil || declaredLines[file][file.Line(source)] {
			continue
		}
		walker.pass.Report(analysis.Diagnostic{
			Pos:     source,
			Message: Message,
			Related: []analysis.RelatedInformation{{Pos: walk, Message: "the walk it reaches"}},
		})
	}
}

// readsLines are the lines a // Reads: <glob> comment declares: its own, for a comment at the end of a line, and the
// line just after its comment group, so a walk reading several globs carries a comment for each above the source.
func readsLines(fileSet *token.FileSet, file *ast.File) map[int]bool {
	lines := map[int]bool{}
	for _, group := range file.Comments {
		for _, comment := range group.List {
			rest, ok := strings.CutPrefix(comment.Text, escapePrefix)
			if !ok {
				continue
			}
			glob, spaced := strings.CutPrefix(rest, " ")
			// An empty glob names nothing for the planner to check, and a comment holds one glob, written as in
			// the reads line, so anything else declares nothing.
			if !spaced || glob == "" || strings.ContainsFunc(glob, unicode.IsSpace) {
				continue
			}
			lines[fileSet.Position(comment.Pos()).Line] = true
			lines[fileSet.Position(group.End()).Line+1] = true
		}
	}
	return lines
}
