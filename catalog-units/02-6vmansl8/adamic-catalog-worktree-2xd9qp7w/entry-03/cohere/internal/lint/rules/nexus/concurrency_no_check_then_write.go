package nexus

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/ecmascript/property"
	"github.com/system-inc/cohere/internal/lint/ecmascript/reference"
	"github.com/system-inc/cohere/internal/lint/ecmascript/scope"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

const concurrencyNoCheckThenWriteId = "checkThenWrite"

// concurrencyNoCheckThenWriteText is the rule's message, whose wording lives in
// `policy/messages/concurrency-no-check-then-write.json`.
var concurrencyNoCheckThenWriteText = policy.MessageOf("nexus/concurrency-no-check-then-write", concurrencyNoCheckThenWriteId)

// concurrencyNoCheckThenWriteMessage names the write that claims the name, so a reader of the finding
// on the check can find the second half of the race without reading the whole function.
func concurrencyNoCheckThenWriteMessage(write string) rule.Message {
	return rule.Message{
		Id:          concurrencyNoCheckThenWriteId,
		Description: concurrencyNoCheckThenWriteText.Render(map[string]string{"write": write}),
	}
}

// ConcurrencyNoCheckThenWrite reports a loop that searches for a free file name by checking whether
// each candidate exists, followed by a write that creates the chosen name without refusing a file that
// is already there. The check and the write are two steps, so two writers can both find the same
// name free, and the second write replaces the first's file.
//
//	invalid: while(true) { try { await access(join(directory, `${index}.png`)); index++; } catch { break; } }
//	         await writeFile(join(directory, `${index}.png`), bytes);
//	invalid: do { name = `${base}-${number}.png`; number++; } while(existsSync(join(directory, name)));
//	         copyFileSync(source, join(directory, name));
//	valid:   await writeFile(join(directory, `${index}.png`), bytes, { flag: 'wx' });   claims and writes at once
//	valid:   while(existsSync(lockPath)) { await delay(100); } writeFileSync(lockPath, '');   no counter
//
// # Where it came from
//
// `modules/intelligence/IntelligenceMediaGenerationApi.ts` in ahra, `persistArtifact` (fixed
// 2026-10-01): it found the next free `<index>.<ext>` with `access()` in a loop and then wrote that
// name with a plain `writeFile`, so two generations saving at once picked the same index and one image
// was lost. The fix, now at `:121`, writes with `{ flag: 'wx' }` and moves to the next index on
// `EEXIST`. Eight live copies of the before-shape were found by the new-rules sweep (`#tevhg3f`) and
// filed as `#9erqsas`: `modules/kling/KlingApi.ts:253` and `:383`, `modules/magnific/MagnificApi.ts:406`,
// `modules/openai/OpenAiApi.ts:309`, `modules/openai/OpenAiCommandLineInterface.ts:144`,
// `modules/art/ArtApi.ts:231`, `modules/art/ArtLibrary.ts:65` and
// `modules/phi/social/PhiSocialLibrary.ts:210`. Built as `#2xqtnzj`.
//
// # Every condition, each one required
//
//   - **A loop that stops on a free name.** Two shapes, nothing else:
//     the loop's condition is exactly `existsSync(path)` (a `while`, a `do ... while` or a `for`), so
//     it ends when the path does not exist; or the loop body holds, as one of its own statements, a
//     `try` with no `finally` whose `catch` is exactly `break;`, and whose `try` block opens with the
//     check (`await access(path)`, `await stat(path)`, `await lstat(path)` from the promise API, or
//     `accessSync`, `statSync`, `lstatSync`) followed by nothing but `++`, `--`, `+=` or `-=` on a
//     variable, so it ends exactly when the check throws. The check takes one argument: `statSync(path,
//     { throwIfNoEntry: false })` never throws, and a mode passed to `access` asks a different question.
//   - **The check is Node's.** Every callee resolves through the checker, aliases followed, to a
//     function declared in the module `fs`, `node:fs`, `fs/promises` or `node:fs/promises`, whether it
//     was reached through a namespace import, a named import or `fs.promises`. A project's own
//     `exists()` or `access()` never matches, whatever its name.
//   - **The checked path is built from a counter.** Some variable the path reads is changed inside the
//     loop by `++`, `--`, `+=` or `-=`, directly or through a variable assigned (or declared) in the loop
//     from one that is. That is what makes the loop a search for a free name rather than a wait (a
//     `while(existsSync(lockPath))` that sleeps until a lock file goes away is a different bug class,
//     and is left alone).
//   - **A write after the loop creates that same path without exclusive create.** The write is in a
//     statement that runs after the loop in the same function, not inside a nested function and not
//     past an enclosing loop (where the next round would run the search again). It resolves to Node's
//     `writeFile`, `writeFileSync`, `copyFile`, `copyFileSync`, `rename`, `renameSync` or
//     `createWriteStream`, or to sharp's `toFile` (below). It is not exclusive only when the source
//     says so: no options, an encoding string, a callback, or an object literal whose `flag` (`flags`
//     for a stream) is absent or a string without `x`. A `copyFile` mode is not exclusive only when
//     absent. Anything the rule cannot read, a variable holding the options, a spread, a computed key,
//     leaves the write alone.
//   - **Same path, proven.** The write's path and the checked path are the same expression once
//     `const` locals are read through their initializers: identifiers resolving to the same binding,
//     string, template and number literals with the same text, `+`, and calls to Node's `path.join`
//     (resolved like the `fs` functions, pure, so the same arguments give the same path). Anything else
//     in the path, a property read, another call, makes the two unequal. A `const` is read through only
//     when it was declared after the loop, or in the loop body before the `try`; one declared earlier
//     is compared as itself, because its value was fixed before the counter moved.
//   - **The path did not move in between.** Every `let`, `var` or parameter the path reads is written
//     only by the function holding the loop (never from a nested function, which could run during an
//     `await`), and never between the end of the loop and the write, nor, when a `const` from the loop
//     body was read through, between that declaration and the check.
//
// # What it deliberately does not see
//
// Every gap is a missed finding, never a false one.
//
//   - A check followed by a write without a loop (`if(!existsSync(path)) writeFileSync(path, ...)`).
//     Measured by the research probe at 27 such sites on ahra, nearly all create-if-missing and cache
//     files, where the second writer would write the same bytes. The loop with a counter is what says
//     the name is being claimed.
//   - A write inside the loop, or in a callback after it, or a check that is not the loop's exit
//     (`if(!existsSync(path)) break;`, a `catch` that tests the error code before breaking).
//   - A path read through a property (`input.outputDirectory`), because anything else holding the
//     object can change it during the `await` between the check and the write, so the two paths cannot
//     be proven equal. That is how the incident's own before-file read its directory and extension,
//     so `persistArtifact` as it stood before the fix is a missed finding (a fixture says so); with
//     the fields read into locals it fires. Also a call other than `path.join` (`path.resolve` reads
//     the working directory, which can change), and a `const` declared before the loop and spelled
//     differently on the two sides.
//   - sharp's `toFile` on an instance in a file that does not import `sharp` itself. sharp's `toFile`
//     is matched only when the method resolves to a declaration in the very file the compiler resolved
//     this file's own `sharp` import to, a package the resolver identifies as `sharp`. That is identity
//     through the resolver and the checker, not a name, and it always overwrites: sharp has no
//     exclusive-create option.
//   - `open` with a write flag, `cp`, `symlink`, `link`, and every third-party writer other than sharp.
//
// # No fix
//
// The fix changes the shape of the loop, claim and write in one step and retry on `EEXIST`, and the
// write is often far from the check, past a network download (`KlingApi.ts:253`) or an encode
// (`PhiSocialLibrary.ts`). Moving it is the author's job. `#9erqsas` proposes one library helper for
// all eight sites rather than eight patched loops.
var ConcurrencyNoCheckThenWrite = rule.Rule{
	Name:             "nexus/concurrency-no-check-then-write",
	NeedsTypeChecker: true,
	// Where this file's `sharp` import resolves.
	ProgramReads: rule.ReadsModuleResolution,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.TypeChecker == nil {
			// Every callee is identified through the checker. Without it the rule could only match
			// names, so it declines the file.
			return nil
		}
		check := func(node *ast.Node) {
			concurrencyNoCheckThenWriteCheckLoop(ctx, node)
		}
		return rule.Listeners{
			ast.KindWhileStatement: check,
			ast.KindDoStatement:    check,
			ast.KindForStatement:   check,
		}
	},
}

// concurrencyNoCheckThenWriteNodeModules are the module names Node's file system API is declared
// under, by `@types/node` from version 20 (`fs`, re-exported as `node:fs`) through 26 (`node:fs`,
// re-exported as `fs`), mapped to whether the module is the promise API.
var concurrencyNoCheckThenWriteNodeModules = map[string]bool{
	"fs":               false,
	"node:fs":          false,
	"fs/promises":      true,
	"node:fs/promises": true,
}

// concurrencyNoCheckThenWritePathModules are the module names Node's `path` is declared under.
var concurrencyNoCheckThenWritePathModules = map[string]bool{
	"path":      true,
	"node:path": true,
}

// concurrencyNoCheckThenWriteThrowingChecks are the checks that throw when the path does not exist,
// keyed by whether they come from the promise API (which must be awaited to throw inside the `try`).
var concurrencyNoCheckThenWriteThrowingChecks = map[bool]map[string]bool{
	false: {"accessSync": true, "statSync": true, "lstatSync": true},
	true:  {"access": true, "stat": true, "lstat": true},
}

// concurrencyNoCheckThenWriteWrite is one write the rule can read: where its destination path sits
// among the arguments, where its options sit, and which option key holds the open flag.
type concurrencyNoCheckThenWriteWrite struct {
	destinationIndex int
	optionsIndex     int
	flagKey          string
}

// concurrencyNoCheckThenWriteWrites are Node's writes that create or replace a file, keyed by name.
// `rename` takes no flag and always replaces. A `copyFile` mode is a number, read below as never
// provably non-exclusive unless it is absent, so it has no flag key.
var concurrencyNoCheckThenWriteWrites = map[string]concurrencyNoCheckThenWriteWrite{
	"writeFile":         {destinationIndex: 0, optionsIndex: 2, flagKey: "flag"},
	"writeFileSync":     {destinationIndex: 0, optionsIndex: 2, flagKey: "flag"},
	"copyFile":          {destinationIndex: 1, optionsIndex: 2},
	"copyFileSync":      {destinationIndex: 1, optionsIndex: 2},
	"rename":            {destinationIndex: 1, optionsIndex: -1},
	"renameSync":        {destinationIndex: 1, optionsIndex: -1},
	"createWriteStream": {destinationIndex: 0, optionsIndex: 1, flagKey: "flags"},
}

// concurrencyNoCheckThenWriteLoop is the analysis of one loop.
type concurrencyNoCheckThenWriteLoop struct {
	ctx  rule.Context
	loop *ast.Node
	// function is the function-like node holding the loop, or nil at module scope.
	function *ast.Node
	// check is the existence check the loop exits on, and checkStatement the `try` holding it, or nil
	// when the check is the loop's condition.
	check          *ast.Node
	checkStatement *ast.Node
	// bodyDeclarations are the `const` statements in the loop body before the `try`, the only ones
	// inside the loop that may be read through.
	bodyDeclarations map[*ast.Node]bool
	varying          map[*ast.Symbol]bool
	writes           map[*ast.Symbol][]*ast.Node
}

// concurrencyNoCheckThenWriteComparison is one attempt to prove a write's path equal to the checked
// path. It records the earliest loop-body `const` read through and every non-constant binding the two
// paths read, whose stability is checked once the shapes match.
type concurrencyNoCheckThenWriteComparison struct {
	bodyDeclarationStart int
	bindings             map[*ast.Symbol]bool
}

// concurrencyNoCheckThenWriteCheckLoop reports a loop that finds a free name and a later write that
// creates it without exclusive create.
func concurrencyNoCheckThenWriteCheckLoop(ctx rule.Context, loop *ast.Node) {
	analysis := &concurrencyNoCheckThenWriteLoop{ctx: ctx, loop: loop, function: scope.EnclosingFunctionLike(loop)}
	if !analysis.findCheck() {
		return
	}
	checkedPath := analysis.check.AsCallExpression().Arguments.Nodes[0]
	analysis.collectVarying()
	if !analysis.readsVarying(checkedPath) {
		return
	}
	for _, statement := range analysis.followingStatements() {
		write, destination := analysis.findWrite(statement, checkedPath)
		if write != nil {
			ctx.ReportNode(analysis.check, concurrencyNoCheckThenWriteMessage(destination))
			return
		}
	}
}

// findCheck finds the existence check the loop exits on, in either of the two shapes the doc comment
// names.
func (analysis *concurrencyNoCheckThenWriteLoop) findCheck() bool {
	var condition *ast.Node
	var body *ast.Node
	switch analysis.loop.Kind {
	case ast.KindWhileStatement:
		condition = analysis.loop.AsWhileStatement().Expression
		body = analysis.loop.AsWhileStatement().Statement
	case ast.KindDoStatement:
		condition = analysis.loop.AsDoStatement().Expression
		body = analysis.loop.AsDoStatement().Statement
	case ast.KindForStatement:
		condition = analysis.loop.AsForStatement().Condition
		body = analysis.loop.AsForStatement().Statement
	}
	if condition != nil {
		call := ast.SkipParentheses(condition)
		if name, promise, matched := analysis.nodeFileSystemFunction(call); matched && !promise && name == "existsSync" &&
			concurrencyNoCheckThenWriteHasOneArgument(call) {
			analysis.check = call
			return true
		}
	}
	if body == nil || body.Kind != ast.KindBlock {
		return false
	}
	statements := body.AsBlock().Statements.Nodes
	for index, statement := range statements {
		if check := analysis.throwingCheckOf(statement); check != nil {
			analysis.check = check
			analysis.checkStatement = statement
			analysis.bodyDeclarations = map[*ast.Node]bool{}
			for _, earlier := range statements[:index] {
				if earlier.Kind == ast.KindVariableStatement {
					analysis.bodyDeclarations[earlier] = true
				}
			}
			return true
		}
	}
	return false
}

// throwingCheckOf returns the check in `try { check; counter++; } catch { break; }`, or nil when the
// statement is not exactly that shape.
func (analysis *concurrencyNoCheckThenWriteLoop) throwingCheckOf(statement *ast.Node) *ast.Node {
	if statement.Kind != ast.KindTryStatement {
		return nil
	}
	tryStatement := statement.AsTryStatement()
	if tryStatement.FinallyBlock != nil || tryStatement.CatchClause == nil {
		return nil
	}
	catchStatements := tryStatement.CatchClause.AsCatchClause().Block.AsBlock().Statements.Nodes
	if len(catchStatements) != 1 || catchStatements[0].Kind != ast.KindBreakStatement ||
		catchStatements[0].AsBreakStatement().Label != nil {
		return nil
	}
	tryStatements := tryStatement.TryBlock.AsBlock().Statements.Nodes
	if len(tryStatements) == 0 || tryStatements[0].Kind != ast.KindExpressionStatement {
		return nil
	}
	for _, following := range tryStatements[1:] {
		if following.Kind != ast.KindExpressionStatement ||
			concurrencyNoCheckThenWriteCounterOf(following.AsExpressionStatement().Expression) == nil {
			return nil
		}
	}
	call := ast.SkipParentheses(tryStatements[0].AsExpressionStatement().Expression)
	awaited := call.Kind == ast.KindAwaitExpression
	if awaited {
		call = ast.SkipParentheses(call.AsAwaitExpression().Expression)
	}
	name, promise, matched := analysis.nodeFileSystemFunction(call)
	if !matched || promise != awaited || !concurrencyNoCheckThenWriteThrowingChecks[promise][name] ||
		!concurrencyNoCheckThenWriteHasOneArgument(call) {
		return nil
	}
	return call
}

// concurrencyNoCheckThenWriteCounterOf returns the identifier an expression counts with (`x++`, `--x`,
// `x += n`, `x -= n`), or nil.
func concurrencyNoCheckThenWriteCounterOf(expression *ast.Node) *ast.Node {
	expression = ast.SkipParentheses(expression)
	var operand *ast.Node
	switch expression.Kind {
	case ast.KindPrefixUnaryExpression:
		if reference.IsUpdateOperator(expression.AsPrefixUnaryExpression().Operator) {
			operand = expression.AsPrefixUnaryExpression().Operand
		}
	case ast.KindPostfixUnaryExpression:
		if reference.IsUpdateOperator(expression.AsPostfixUnaryExpression().Operator) {
			operand = expression.AsPostfixUnaryExpression().Operand
		}
	case ast.KindBinaryExpression:
		switch expression.AsBinaryExpression().OperatorToken.Kind {
		case ast.KindPlusEqualsToken, ast.KindMinusEqualsToken:
			operand = expression.AsBinaryExpression().Left
		}
	}
	if operand == nil {
		return nil
	}
	operand = ast.SkipParentheses(operand)
	if operand.Kind != ast.KindIdentifier {
		return nil
	}
	return operand
}

// concurrencyNoCheckThenWriteHasOneArgument says whether a call passes exactly one argument, not
// spread.
func concurrencyNoCheckThenWriteHasOneArgument(call *ast.Node) bool {
	arguments := call.AsCallExpression().Arguments.Nodes
	return len(arguments) == 1 && arguments[0].Kind != ast.KindSpreadElement
}

// resolve returns the symbol an identifier names, aliases followed.
func (analysis *concurrencyNoCheckThenWriteLoop) resolve(node *ast.Node) *ast.Symbol {
	symbol := analysis.ctx.TypeChecker.GetSymbolAtLocation(node)
	if symbol == nil {
		return nil
	}
	return checker.SkipAlias(symbol, analysis.ctx.TypeChecker)
}

// calleeSymbol returns the symbol a call's callee resolves to, for a callee written as a name or a
// non-optional property access, and nil for anything else.
func (analysis *concurrencyNoCheckThenWriteLoop) calleeSymbol(call *ast.Node) *ast.Symbol {
	if call == nil || call.Kind != ast.KindCallExpression || call.AsCallExpression().QuestionDotToken != nil {
		return nil
	}
	callee := ast.SkipParentheses(call.AsCallExpression().Expression)
	switch callee.Kind {
	case ast.KindIdentifier:
		return analysis.resolve(callee)
	case ast.KindPropertyAccessExpression:
		if callee.AsPropertyAccessExpression().QuestionDotToken != nil {
			return nil
		}
		return analysis.resolve(callee.AsPropertyAccessExpression().Name())
	}
	return nil
}

// concurrencyNoCheckThenWriteDeclaringModule returns the name of the ambient module (`declare module
// 'node:fs'`) a declaration sits in, looking out through any namespace inside it, or "" when it is not
// in one.
func concurrencyNoCheckThenWriteDeclaringModule(declaration *ast.Node) string {
	for current := declaration.Parent; current != nil; current = current.Parent {
		if current.Kind == ast.KindModuleDeclaration {
			name := current.AsModuleDeclaration().Name()
			if name != nil && name.Kind == ast.KindStringLiteral {
				return name.Text()
			}
		}
	}
	return ""
}

// concurrencyNoCheckThenWriteDeclaredIn says whether a symbol is a function every declaration of which
// sits in one of the named ambient modules, and returns that module.
func concurrencyNoCheckThenWriteDeclaredIn(symbol *ast.Symbol, modules map[string]bool) (string, bool) {
	if symbol == nil || len(symbol.Declarations) == 0 {
		return "", false
	}
	module := ""
	for _, declaration := range symbol.Declarations {
		if declaration.Kind != ast.KindFunctionDeclaration && declaration.Kind != ast.KindMethodSignature {
			return "", false
		}
		declared := concurrencyNoCheckThenWriteDeclaringModule(declaration)
		if _, known := modules[declared]; !known || (module != "" && declared != module) {
			return "", false
		}
		module = declared
	}
	return module, true
}

// nodeFileSystemFunction returns the name of the Node file system function a call resolves to and
// whether it is the promise API.
func (analysis *concurrencyNoCheckThenWriteLoop) nodeFileSystemFunction(call *ast.Node) (string, bool, bool) {
	symbol := analysis.calleeSymbol(call)
	module, matched := concurrencyNoCheckThenWriteDeclaredIn(symbol, concurrencyNoCheckThenWriteNodeModules)
	if !matched {
		return "", false, false
	}
	return symbol.Name, concurrencyNoCheckThenWriteNodeModules[module], true
}

// isPathJoin says whether a call is Node's `path.join`, which is pure: the same arguments give the
// same path.
func (analysis *concurrencyNoCheckThenWriteLoop) isPathJoin(call *ast.Node) (*ast.Symbol, bool) {
	symbol := analysis.calleeSymbol(call)
	if _, matched := concurrencyNoCheckThenWriteDeclaredIn(symbol, concurrencyNoCheckThenWritePathModules); !matched || symbol.Name != "join" {
		return nil, false
	}
	return symbol, true
}

// isSharpToFile says whether a call is sharp's `toFile`: the method resolves to a declaration in the
// file the compiler resolved this file's own import of the package `sharp` to.
func (analysis *concurrencyNoCheckThenWriteLoop) isSharpToFile(call *ast.Node) bool {
	symbol := analysis.calleeSymbol(call)
	if symbol == nil || symbol.Name != "toFile" || len(symbol.Declarations) == 0 {
		return false
	}
	sharpFiles := rule.Cached(analysis.ctx.FileCache, "concurrencyNoCheckThenWriteSharpFiles", func() map[string]bool {
		files := map[string]bool{}
		if analysis.ctx.Program == nil {
			return files
		}
		for _, specifier := range analysis.ctx.SourceFile.Imports() {
			resolved := analysis.ctx.Program.ResolveModule(analysis.ctx.SourceFile, specifier)
			if resolved.IsResolved() && resolved.IsExternalLibraryImport && resolved.PackageName == "sharp" {
				files[resolved.ResolvedFileName] = true
			}
		}
		return files
	})
	for _, declaration := range symbol.Declarations {
		if declaration.Kind != ast.KindMethodSignature || !sharpFiles[ast.GetSourceFileOfNode(declaration).FileName().AsString()] {
			return false
		}
	}
	return true
}

// forEachOwnNode visits a subtree in source order without entering nested functions, whose code runs
// on its own schedule.
func concurrencyNoCheckThenWriteForEachOwnNode(root *ast.Node, visitor func(node *ast.Node)) {
	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		if ast.IsFunctionLike(node) || node.Kind == ast.KindClassStaticBlockDeclaration {
			return false
		}
		visitor(node)
		node.ForEachChild(visit)
		return false
	}
	visit(root)
}

// readsAny says whether an expression reads, by identifier, any binding in a
// set.
func (analysis *concurrencyNoCheckThenWriteLoop) readsAny(expression *ast.Node, symbols map[*ast.Symbol]bool) bool {
	found := false
	concurrencyNoCheckThenWriteForEachOwnNode(expression, func(node *ast.Node) {
		if !found && node.Kind == ast.KindIdentifier {
			if symbol := analysis.resolve(node); symbol != nil && symbols[symbol] {
				found = true
			}
		}
	})
	return found
}

// collectVarying finds the loop's counters, the variables it changes with `++`, `--`, `+=` or `-=`,
// and every variable assigned or declared in the loop from one of them.
func (analysis *concurrencyNoCheckThenWriteLoop) collectVarying() {
	analysis.varying = map[*ast.Symbol]bool{}
	type assignment struct {
		target *ast.Symbol
		value  *ast.Node
	}
	var assignments []assignment
	concurrencyNoCheckThenWriteForEachOwnNode(analysis.loop, func(node *ast.Node) {
		if counter := concurrencyNoCheckThenWriteCounterOf(node); counter != nil {
			if symbol := analysis.resolve(counter); symbol != nil {
				analysis.varying[symbol] = true
			}
			return
		}
		switch node.Kind {
		case ast.KindBinaryExpression:
			binary := node.AsBinaryExpression()
			left := ast.SkipParentheses(binary.Left)
			if binary.OperatorToken.Kind == ast.KindEqualsToken && left.Kind == ast.KindIdentifier {
				if symbol := analysis.resolve(left); symbol != nil {
					assignments = append(assignments, assignment{target: symbol, value: binary.Right})
				}
			}
		case ast.KindVariableDeclaration:
			declaration := node.AsVariableDeclaration()
			if declaration.Name().Kind == ast.KindIdentifier && declaration.Initializer != nil {
				if symbol := analysis.resolve(declaration.Name()); symbol != nil {
					assignments = append(assignments, assignment{target: symbol, value: declaration.Initializer})
				}
			}
		}
	})
	for changed := true; changed; {
		changed = false
		for _, candidate := range assignments {
			if !analysis.varying[candidate.target] && analysis.readsAny(candidate.value, analysis.varying) {
				analysis.varying[candidate.target] = true
				changed = true
			}
		}
	}
}

// readsVarying says whether the checked path reads a counter, through any `const` the comparison is
// allowed to read through.
func (analysis *concurrencyNoCheckThenWriteLoop) readsVarying(path *ast.Node) bool {
	found := false
	visited := map[*ast.Symbol]bool{}
	var visit func(node *ast.Node)
	visit = func(node *ast.Node) {
		concurrencyNoCheckThenWriteForEachOwnNode(node, func(inner *ast.Node) {
			if found || inner.Kind != ast.KindIdentifier {
				return
			}
			symbol := analysis.resolve(inner)
			if symbol == nil || visited[symbol] {
				return
			}
			visited[symbol] = true
			if analysis.varying[symbol] {
				found = true
				return
			}
			if initializer, _ := analysis.readThrough(symbol); initializer != nil {
				visit(initializer)
			}
		})
	}
	visit(path)
	return found
}

// followingStatements lists the statements that run after the loop in the same function, climbing out
// through blocks, branches and `try` but stopping at an enclosing loop or function.
func (analysis *concurrencyNoCheckThenWriteLoop) followingStatements() []*ast.Node {
	var following []*ast.Node
	current := analysis.loop
	for current.Parent != nil {
		parent := current.Parent
		if ast.IsFunctionLike(parent) || ast.IsIterationStatement(parent, false) ||
			parent.Kind == ast.KindClassStaticBlockDeclaration || parent.Kind == ast.KindModuleBlock {
			break
		}
		var statements []*ast.Node
		switch parent.Kind {
		case ast.KindBlock:
			statements = parent.AsBlock().Statements.Nodes
		case ast.KindCaseClause, ast.KindDefaultClause:
			statements = parent.AsCaseOrDefaultClause().Statements.Nodes
		case ast.KindSourceFile:
			statements = parent.AsSourceFile().Statements.Nodes
		}
		for index, statement := range statements {
			if statement == current {
				following = append(following, statements[index+1:]...)
				break
			}
		}
		current = parent
	}
	return following
}

// findWrite returns the first write in a statement that creates the checked path without exclusive
// create, with the name to show for it.
func (analysis *concurrencyNoCheckThenWriteLoop) findWrite(statement *ast.Node, checkedPath *ast.Node) (*ast.Node, string) {
	var found *ast.Node
	var name string
	concurrencyNoCheckThenWriteForEachOwnNode(statement, func(node *ast.Node) {
		if found != nil || node.Kind != ast.KindCallExpression {
			return
		}
		destination, writeName := analysis.nonExclusiveDestination(node)
		if destination == nil {
			return
		}
		if analysis.samePathAt(checkedPath, destination, node) {
			found = node
			name = writeName
		}
	})
	return found, name
}

// nonExclusiveDestination returns the path a call creates or replaces when it is a write the rule
// knows and its source proves it is not exclusive, with the name to show for it.
func (analysis *concurrencyNoCheckThenWriteLoop) nonExclusiveDestination(call *ast.Node) (*ast.Node, string) {
	arguments := call.AsCallExpression().Arguments.Nodes
	for _, argument := range arguments {
		if argument.Kind == ast.KindSpreadElement {
			return nil, ""
		}
	}
	if analysis.isSharpToFile(call) {
		if len(arguments) == 0 {
			return nil, ""
		}
		return arguments[0], "toFile"
	}
	name, _, matched := analysis.nodeFileSystemFunction(call)
	if !matched {
		return nil, ""
	}
	write, known := concurrencyNoCheckThenWriteWrites[name]
	if !known || len(arguments) <= write.destinationIndex {
		return nil, ""
	}
	if write.optionsIndex >= 0 && len(arguments) > write.optionsIndex &&
		!concurrencyNoCheckThenWriteOptionsNotExclusive(arguments[write.optionsIndex], write.flagKey) {
		return nil, ""
	}
	return arguments[write.destinationIndex], name
}

// concurrencyNoCheckThenWriteOptionsNotExclusive says whether an options argument provably leaves the
// write non-exclusive. A callback in the options position (`writeFile(path, data, done)`) passes no
// options. Anything the source does not settle answers false, which leaves the write alone.
func concurrencyNoCheckThenWriteOptionsNotExclusive(options *ast.Node, flagKey string) bool {
	options = ast.SkipParentheses(options)
	switch options.Kind {
	case ast.KindArrowFunction, ast.KindFunctionExpression:
		return true
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		// An encoding. `copyFile` and `rename` never take a string here.
		return flagKey != ""
	case ast.KindObjectLiteralExpression:
		if flagKey == "" {
			return false
		}
		for _, member := range options.AsObjectLiteralExpression().Properties.Nodes {
			if member.Kind != ast.KindPropertyAssignment {
				// A spread, a shorthand or a method could carry the flag in a way the source does
				// not show.
				return false
			}
			key, static := property.Name(member.Name(), property.Static)
			if !static {
				return false
			}
			if key != flagKey {
				continue
			}
			value := ast.SkipParentheses(member.AsPropertyAssignment().Initializer)
			if value.Kind != ast.KindStringLiteral && value.Kind != ast.KindNoSubstitutionTemplateLiteral {
				return false
			}
			if strings.Contains(value.Text(), "x") {
				return false
			}
		}
		return true
	}
	return false
}

// readThrough returns the initializer of a `const` the comparison may read through, and, for one
// declared in the loop body, the position of its statement.
func (analysis *concurrencyNoCheckThenWriteLoop) readThrough(symbol *ast.Symbol) (*ast.Node, int) {
	declaration := symbol.ValueDeclaration
	if declaration == nil || declaration.Kind != ast.KindVariableDeclaration || !ast.IsVarConst(declaration) {
		return nil, -1
	}
	variable := declaration.AsVariableDeclaration()
	if variable.Name().Kind != ast.KindIdentifier || variable.Initializer == nil ||
		ast.GetSourceFileOfNode(declaration) != analysis.ctx.SourceFile ||
		scope.EnclosingFunctionLike(declaration) != analysis.function {
		return nil, -1
	}
	if declaration.Pos() >= analysis.loop.End() {
		return variable.Initializer, -1
	}
	statement := declaration.Parent.Parent
	if analysis.bodyDeclarations[statement] {
		return variable.Initializer, statement.Pos()
	}
	return nil, -1
}

// samePathAt says whether a write's destination is proven to be the checked path when the write runs.
func (analysis *concurrencyNoCheckThenWriteLoop) samePathAt(checkedPath *ast.Node, destination *ast.Node, write *ast.Node) bool {
	comparison := &concurrencyNoCheckThenWriteComparison{bodyDeclarationStart: -1, bindings: map[*ast.Symbol]bool{}}
	if !analysis.same(comparison, checkedPath, destination) {
		return false
	}
	for symbol := range comparison.bindings {
		if !analysis.holdsStill(symbol, comparison.bodyDeclarationStart, write) {
			return false
		}
	}
	return true
}

// expand reads an identifier through a `const` the comparison may read through, recording a loop-body
// declaration's position, and returns the node to compare in its place.
func (analysis *concurrencyNoCheckThenWriteLoop) expand(comparison *concurrencyNoCheckThenWriteComparison, node *ast.Node) *ast.Node {
	for depth := 0; depth < 16; depth++ {
		node = ast.SkipParentheses(node)
		if node.Kind != ast.KindIdentifier {
			return node
		}
		symbol := analysis.resolve(node)
		if symbol == nil {
			return node
		}
		initializer, bodyPosition := analysis.readThrough(symbol)
		if initializer == nil {
			return node
		}
		if bodyPosition >= 0 && (comparison.bodyDeclarationStart < 0 || bodyPosition < comparison.bodyDeclarationStart) {
			comparison.bodyDeclarationStart = bodyPosition
		}
		node = initializer
	}
	return node
}

// same compares two path expressions as the doc comment's "Same path, proven" describes.
func (analysis *concurrencyNoCheckThenWriteLoop) same(comparison *concurrencyNoCheckThenWriteComparison, left *ast.Node, right *ast.Node) bool {
	left = analysis.expand(comparison, left)
	right = analysis.expand(comparison, right)
	if left.Kind != right.Kind {
		return false
	}
	switch left.Kind {
	case ast.KindIdentifier:
		leftSymbol := analysis.resolve(left)
		if leftSymbol == nil || leftSymbol != analysis.resolve(right) {
			return false
		}
		if !concurrencyNoCheckThenWriteIsConstant(leftSymbol) {
			comparison.bindings[leftSymbol] = true
		}
		return true
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral, ast.KindNumericLiteral:
		return left.Text() == right.Text()
	case ast.KindTemplateExpression:
		leftTemplate := left.AsTemplateExpression()
		rightTemplate := right.AsTemplateExpression()
		if leftTemplate.Head.Text() != rightTemplate.Head.Text() ||
			len(leftTemplate.TemplateSpans.Nodes) != len(rightTemplate.TemplateSpans.Nodes) {
			return false
		}
		for index, leftSpan := range leftTemplate.TemplateSpans.Nodes {
			rightSpan := rightTemplate.TemplateSpans.Nodes[index]
			if leftSpan.AsTemplateSpan().Literal.Text() != rightSpan.AsTemplateSpan().Literal.Text() ||
				!analysis.same(comparison, leftSpan.AsTemplateSpan().Expression, rightSpan.AsTemplateSpan().Expression) {
				return false
			}
		}
		return true
	case ast.KindBinaryExpression:
		leftBinary := left.AsBinaryExpression()
		rightBinary := right.AsBinaryExpression()
		return leftBinary.OperatorToken.Kind == ast.KindPlusToken && rightBinary.OperatorToken.Kind == ast.KindPlusToken &&
			analysis.same(comparison, leftBinary.Left, rightBinary.Left) &&
			analysis.same(comparison, leftBinary.Right, rightBinary.Right)
	case ast.KindCallExpression:
		leftJoin, leftIsJoin := analysis.isPathJoin(left)
		rightJoin, rightIsJoin := analysis.isPathJoin(right)
		if !leftIsJoin || !rightIsJoin || leftJoin != rightJoin {
			return false
		}
		leftArguments := left.AsCallExpression().Arguments.Nodes
		rightArguments := right.AsCallExpression().Arguments.Nodes
		if len(leftArguments) != len(rightArguments) {
			return false
		}
		for index := range leftArguments {
			if leftArguments[index].Kind == ast.KindSpreadElement || rightArguments[index].Kind == ast.KindSpreadElement ||
				!analysis.same(comparison, leftArguments[index], rightArguments[index]) {
				return false
			}
		}
		return true
	}
	return false
}

// concurrencyNoCheckThenWriteIsConstant says whether a binding can never be reassigned: a `const`, an
// import, a function, a class or an enum. Everything else is a `let`, a `var` or a parameter, whose
// writes are checked.
func concurrencyNoCheckThenWriteIsConstant(symbol *ast.Symbol) bool {
	declaration := symbol.ValueDeclaration
	if declaration == nil {
		return false
	}
	switch declaration.Kind {
	case ast.KindVariableDeclaration:
		return ast.IsVarConst(declaration)
	case ast.KindFunctionDeclaration, ast.KindClassDeclaration, ast.KindEnumDeclaration,
		ast.KindImportSpecifier, ast.KindImportClause, ast.KindNamespaceImport:
		return true
	}
	return false
}

// holdsStill says whether a `let`, `var` or parameter keeps the value the check saw until the write
// runs: it is declared in this file, every write to it is in the loop's own function, none from a nested one, and none between the
// end of the loop and the write, nor between a loop-body `const` read through and the check.
func (analysis *concurrencyNoCheckThenWriteLoop) holdsStill(symbol *ast.Symbol, bodyDeclarationStart int, write *ast.Node) bool {
	declaration := symbol.ValueDeclaration
	if declaration == nil || ast.GetSourceFileOfNode(declaration) != analysis.ctx.SourceFile {
		// Another module's `let` can change while this function waits, and its writes are not in
		// this file to read.
		return false
	}
	if analysis.writes == nil {
		analysis.writes = map[*ast.Symbol][]*ast.Node{}
	}
	writes, cached := analysis.writes[symbol]
	if !cached {
		root := scope.EnclosingFunctionLike(declaration)
		if root == nil {
			root = analysis.ctx.SourceFile.AsNode()
		}
		var visit func(node *ast.Node) bool
		visit = func(node *ast.Node) bool {
			if node.Kind == ast.KindIdentifier && !ast.IsDeclarationName(node) &&
				reference.WritesToBinding(node) && analysis.resolve(node) == symbol {
				writes = append(writes, node)
			}
			node.ForEachChild(visit)
			return false
		}
		visit(root)
		analysis.writes[symbol] = writes
	}
	for _, written := range writes {
		if scope.EnclosingFunctionLike(written) != analysis.function {
			return false
		}
		if written.Pos() >= analysis.loop.End() && written.Pos() < write.Pos() {
			return false
		}
		if bodyDeclarationStart >= 0 && written.Pos() >= bodyDeclarationStart && written.Pos() < analysis.checkStatement.Pos() {
			return false
		}
	}
	return true
}
