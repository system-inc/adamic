package lower

import (
	"math"
	"reflect"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Exceptions (docs/memory.md, "Exceptions, designed into counting"): throw new Error(message), a
// caught error thrown again, and try with catch, finally or both. What's thrown is only ever an
// Error, made by new Error (or a built-in subclass, new TypeError among them) or caught, since a catch
// binds unknown and 0.2 has no value of every kind. Any other thrown value is NotYet, never refused
// (docs/0.1.md's exceptions row, ruled October 19).

// errorClasses are the built-in Error classes new makes, each an Error whose name slot holds the
// class's name. Each one's message and name, caught and uncaught, agree with Node natively and on the
// JavaScript backend (internal/oracle/throw_type_error_test.go), and every check that keeps an Error off
// a plain object's paths (its name and message reads, its prototype methods, Object.keys) takes them all.
var errorClasses = []string{"Error", "TypeError", "RangeError", "ReferenceError", "SyntaxError"}

// errorClass names the built-in Error class a new expression makes, or "" for any other node.
func (l *lowering) errorClass(node *ast.Node) string {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindNewExpression {
		return ""
	}
	for _, class := range errorClasses {
		if l.isLibraryGlobal(node.AsNewExpression().Expression, class) {
			return class
		}
	}
	return ""
}

// tryRecord is a try statement lowered, for the checks made once every function is: where it is,
// and its body.
type tryRecord struct {
	node *ast.Node
	body []ir.Statement
}

// throwStatement lowers throw.
func (l *lowering) throwStatement(node *ast.Node) ([]ir.Statement, error) {
	thrown := ast.SkipParentheses(node.AsThrowStatement().Expression)
	isNewError := l.errorClass(thrown) != ""
	isCaught := ast.IsIdentifier(thrown) && l.caught[l.symbol(thrown)]
	// Error is structurally typed: { name, message, stack } can satisfy it without
	// being an Error on Node. Follow only immutable bindings to a real allocation.
	seen := map[*ast.Symbol]bool{}
	var madeError func(*ast.Node) bool
	madeError = func(value *ast.Node) bool {
		if value == nil {
			return false
		}
		value = ast.SkipParentheses(value)
		if value.Kind == ast.KindNewExpression {
			return l.errorClass(value) != ""
		}
		if !ast.IsIdentifier(value) {
			return false
		}
		symbol := l.symbol(value)
		if symbol == nil || seen[symbol] || len(symbol.Declarations) != 1 {
			return false
		}
		seen[symbol] = true
		declaration := symbol.Declarations[0]
		if declaration.Kind != ast.KindVariableDeclaration || declaration.Parent == nil || declaration.Parent.Kind != ast.KindVariableDeclarationList || declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
			return false
		}
		return madeError(declaration.AsVariableDeclaration().Initializer)
	}
	isStoredError := !isNewError && !isCaught && l.isLibraryType(l.checker.GetTypeAtLocation(thrown), errorClasses...) && madeError(thrown)
	if !isNewError && !isCaught && !isStoredError {
		if l.isLibraryType(l.checker.GetTypeAtLocation(thrown), errorClasses...) {
			return nil, l.notYet(thrown, "throwing an Error that isn't made where it's thrown or caught by the catch around it")
		}
		// Any value may be thrown (docs/0.1.md, ruled October 19): until the runtime carries one that
		// isn't an Error, it isn't built yet, never refused.
		return nil, l.notYet(thrown, "throwing "+l.thrownCategory(l.checker.GetTypeAtLocation(thrown)))
	}
	value, err := l.expression(thrown)
	if err != nil {
		return nil, err
	}
	return []ir.Statement{ir.Throw{Value: value}}, nil
}

// thrownCategory names a thrown value that isn't an Error by what JavaScript's typeof would call it, never
// by its type's text: a literal type prints the literal ("a 3") and a structural one its whole shape, and
// the program's own text is already at the location. A union across categories is "a value".
func (l *lowering) thrownCategory(thrown *checker.Type) string {
	var category func(*checker.Type) string
	category = func(member *checker.Type) string {
		flags := member.Flags()
		switch {
		case flags&checker.TypeFlagsStringLike != 0:
			return "a string"
		case flags&checker.TypeFlagsNumberLike != 0:
			return "a number"
		case flags&checker.TypeFlagsBooleanLike != 0:
			return "a boolean"
		case flags&checker.TypeFlagsBigIntLike != 0:
			return "a bigint"
		case flags&checker.TypeFlagsESSymbolLike != 0:
			return "a symbol"
		case flags&checker.TypeFlagsNull != 0:
			return "null"
		case flags&(checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0:
			return "undefined"
		case flags&checker.TypeFlagsUnion != 0:
			named := ""
			for _, each := range member.Types() {
				if one := category(each); named == "" {
					named = one
				} else if one != named {
					return "a value"
				}
			}
			return named
		case flags&(checker.TypeFlagsObject|checker.TypeFlagsNonPrimitive) != 0:
			if len(l.checker.GetSignaturesOfType(member, checker.SignatureKindCall)) > 0 {
				return "a function"
			}
			return "an object"
		}
		return "a value"
	}
	named := category(thrown)
	if named == "null" || named == "undefined" {
		return named + ", which isn't an Error,"
	}
	return named + " that isn't an Error"
}

// newError lowers new Error(message), and new Error(), or one of the other errorClasses, whose name
// goes in the error's name slot.
func (l *lowering) newError(node *ast.Node) (ir.Expression, error) {
	created := node.AsNewExpression()
	class := l.errorClass(node)
	message := ir.Expression(ir.Undefined{Of: ir.String})
	if created.Arguments != nil && len(created.Arguments.Nodes) > 0 {
		if len(created.Arguments.Nodes) > 1 {
			return nil, l.notYet(node, "new "+class+" with options")
		}
		var err error
		if message, err = l.expression(created.Arguments.Nodes[0]); err != nil {
			return nil, err
		}
		message = fit(message, ir.String)
		if message.Type() != ir.String {
			return nil, l.notYet(node, "new "+class+" with a message that isn't a string")
		}
	}
	if class == "Error" {
		return ir.MakeError{Message: message}, nil
	}
	return ir.MakeError{Message: message, Name: ir.StringConstant{Index: l.constant(class)}}, nil
}

// tryStatement lowers try, with catch, finally or both.
func (l *lowering) tryStatement(node *ast.Node) ([]ir.Statement, error) {
	statement := node.AsTryStatement()
	lowered := ir.Try{CatchLocal: -1}
	var err error
	if lowered.Body, err = l.statements(statement.TryBlock.AsBlock().Statements.Nodes); err != nil {
		return nil, err
	}
	l.tries = append(l.tries, tryRecord{node: node, body: lowered.Body})
	if clause := statement.CatchClause; clause != nil {
		lowered.HasCatch = true
		if declaration := clause.AsCatchClause().VariableDeclaration; declaration != nil {
			name := declaration.Name()
			if !ast.IsIdentifier(name) {
				return nil, l.notYet(name, "a catch that destructures what it caught")
			}
			if l.caught == nil {
				l.caught = map[*ast.Symbol]bool{}
			}
			l.caught[l.symbol(name)] = true
			if lowered.CatchLocal, err = l.declareLocal(name); err != nil {
				return nil, err
			}
		}
		if lowered.Catch, err = l.statements(clause.AsCatchClause().Block.AsBlock().Statements.Nodes); err != nil {
			return nil, err
		}
	}
	if block := statement.FinallyBlock; block != nil {
		lowered.HasFinally = true
		// A return, break or continue out of the finally overrides whatever the try was doing, a
		// throw or a return included: the emitter lets go of the pending error or the held result on
		// the way out, as the scopes they wait in are left.
		if lowered.Finally, err = l.statements(block.AsBlock().Statements.Nodes); err != nil {
			return nil, err
		}
	}
	return []ir.Statement{lowered}, nil
}

// caughtInstanceOfError lowers error instanceof Error on what a catch took, which is always an Error.
func (l *lowering) caughtInstanceOfError(node *ast.Node) (ir.Expression, bool) {
	binary := node.AsBinaryExpression()
	left := ast.SkipParentheses(binary.Left)
	if binary.OperatorToken.Kind != ast.KindInstanceOfKeyword || !ast.IsIdentifier(left) || !l.caught[l.symbol(left)] || !l.isLibraryGlobal(binary.Right, "Error") {
		return nil, false
	}
	l.local(left)
	return ir.BooleanConstant{Value: true}, true
}

// exceptions works out which functions a throw can leave, once every function is lowered, and
// refuses what can't be done yet: a try that can reach a library call whose failure is a panic
// natively but a throw on Node.
func (l *lowering) exceptions() error {
	functions := l.result.Functions
	// A write to a frozen object throws, and which objects a freeze reaches isn't known, so a program
	// that freezes any has every object write counted as one that can (ir.Program's FrozenWriteMayThrow).
	l.result.ObjectsMayFreeze = l.objectCanFreeze()
	// A class's methods are reached through function values too: a call through an interface the
	// class implements calls one where it would call the object's own function value (ir.Property's
	// Method), so a method that can throw makes every such call one that can.
	dispatched := map[int]bool{}
	for _, lowered := range l.instances {
		for _, method := range lowered.methodList() {
			dispatched[method.Function] = true
		}
	}
	// Whether any function value can throw, and so every call through one, grows with what can, so
	// the two are worked out together until neither changes.
	for changed := true; changed; {
		changed = false
		for index := range functions {
			if (functions[index].Closure || dispatched[index]) && functions[index].MayThrow && !l.result.ClosuresMayThrow {
				l.result.ClosuresMayThrow = true
				changed = true
			}
			if !functions[index].Async && !functions[index].MayThrow && l.throwsOut(functions[index].Body) {
				functions[index].MayThrow = true
				changed = true
			}
		}
	}
	for _, record := range l.tries {
		if failing := l.libraryFailure(record.body, map[int]bool{}); failing != "" {
			return l.notYet(record.node, "a try around "+failing+", whose failure is a panic natively but a throw a catch can take on Node (docs/memory.md)")
		}
	}
	return nil
}

// throwsOut reports whether a throw can leave statements: a throw, or a call to a function that can
// throw (through a function value too, a sort's comparator among them), that isn't inside a try's
// body with a catch.
func (l *lowering) throwsOut(statements []ir.Statement) bool {
	found := false
	walk(statements, func(node any) bool {
		switch node := node.(type) {
		case ir.Try:
			if node.HasCatch {
				// The body's throws are caught; the catch's and the finally's aren't.
				found = found || l.throwsOut(node.Catch) || l.throwsOut(node.Finally)
				return false
			}
		case ir.ProcessCall:
			found = found || node.Operation == "exit" || node.Operation == "setExitCode" || node.Operation == "cwd" || node.Operation == "chdir" || node.Operation == "measure"
		case ir.NodeHostCall:
			found = found || node.Throws
		case ir.NodeFSFile:
			found = found || node.MayThrow()
		case ir.ToFixed:
			found = found || node.MayThrow()
		case ir.NumberFormat:
			found = found || node.MayThrow()
		case ir.ArrayFill:
			found = found || node.MayThrow()
		case ir.ObjectLiteral:
			found = found || l.result.ObjectSpreadMayThrow(node)
		case ir.ObjectCall:
			found = found || l.result.FrozenWriteMayThrow(node)
		case ir.PhantomMember:
			found = found || !node.Optional
		case ir.Throw:
			found = true
		case ir.RecordCall, ir.RecordCoalesce:
			found = found || l.result.RecordWriteMayThrow(node)
		case ir.SetProperty:
			found = found || l.result.RecordWriteMayThrow(node) || l.result.FrozenWriteMayThrow(node)
		case ir.Defined:
			found = found || node.Throws()
		case ir.Call:
			if l.result.CallMayThrow(node) {
				found = true
			}
		case ir.RegExpCall:
			if (node.Replacement != nil && l.result.ClosuresMayThrow) || l.result.GlobalFlagMayThrow(node) {
				found = true
			}
		case ir.CallClosure, ir.ArrayMap, ir.ArrayVisit, ir.ArrayReduce, ir.MapForEach:
			if l.result.ClosureMayThrow(node.(ir.Expression)) {
				found = true
			}
		case ir.ArrayFrom:
			// Its length past 2^32 - 1, or the callback it calls.
			if node.MayThrow() || l.result.ClosureMayThrow(node) {
				found = true
			}
		case ir.ParallelMap:
			// A call through a function value, written out or made by the runtime's loop.
			if l.result.ClosuresMayThrow {
				found = true
			}
		case ir.ArraySort:
			if (node.Callback != nil && l.result.ClosuresMayThrow) || (node.Callback == nil && l.result.Functions[node.Comparator].MayThrow) {
				found = true
			}
		}
		return !found
	})
	return found
}

// libraryFailure names the first library call statements can reach, directly or through the
// functions they call, that throws in JavaScript for some argument and isn't given a constant that
// can't fail, or returns "".
func (l *lowering) libraryFailure(statements []ir.Statement, visited map[int]bool) string {
	failing := ""
	callsClosures := false
	walk(statements, func(node any) bool {
		switch node := node.(type) {
		case ir.Call:
			for _, target := range l.result.CallTargets(node) {
				if !visited[target] {
					visited[target] = true
					failing = l.libraryFailure(l.result.Functions[target].Body, visited)
					if failing != "" {
						break
					}
				}
			}
		case ir.CallClosure, ir.ParallelMap, ir.ArrayMap, ir.ArrayVisit, ir.ArrayReduce, ir.MapForEach:
			callsClosures = true
		case ir.ArraySort:
			if node.Callback != nil {
				callsClosures = true
			} else if !visited[node.Comparator] {
				visited[node.Comparator] = true
				failing = l.libraryFailure(l.result.Functions[node.Comparator].Body, visited)
			}
		case ir.NodeBufferCall:
			if node.Function == "hash_update" || node.Function == "hash_digest" {
				failing = "Hash finalization, whose catchable .code contract is not supported yet"
			}
		case ir.RegExpCall:
			if node.Replacement != nil {
				callsClosures = true
			}
		case ir.StringCall:
			switch {
			case node.Method == "repeat" && !ir.ConstantWithin(node.Arguments[0], 0, math.MaxFloat64):
				failing = "repeat"
			case node.Method == "normalize" && len(node.Arguments) > 0 && !isNormalizationForm(node.Arguments[0], l.result.Strings):
				failing = "normalize"
			}
		case ir.TypedArrayNew:
			if !node.FromArray && !ir.ConstantWithin(node.Source, 0, 9007199254740991) {
				failing = "typed array length conversion"
			}
		case ir.TypedArraySet:
			failing = "typed array set range validation"
		case ir.ArrayFrom:
			callsClosures = true
		}
		return failing == ""
	})
	if failing == "" && callsClosures {
		// What a function value does is any function value's: each of them is reached.
		for _, closure := range l.closureRecords {
			if !visited[closure.function] {
				visited[closure.function] = true
				if failing = l.libraryFailure(l.result.Functions[closure.function].Body, visited); failing != "" {
					break
				}
			}
		}
	}
	return failing
}

func isNormalizationForm(value ir.Expression, strings []string) bool {
	constant, isConstant := value.(ir.StringConstant)
	if !isConstant {
		return false
	}
	switch strings[constant.Index] {
	case "NFC", "NFD", "NFKC", "NFKD":
		return true
	}
	return false
}

// walk visits every IR statement and expression in a value, depth first, while visit says to go on
// into each. It doesn't go into a function's body: a call is visited as a call.
func walk(value any, visit func(node any) bool) {
	walkValue(reflect.ValueOf(value), visit)
}

var (
	expressionType = reflect.TypeOf((*ir.Expression)(nil)).Elem()
	statementType  = reflect.TypeOf((*ir.Statement)(nil)).Elem()
)

func walkValue(value reflect.Value, visit func(node any) bool) {
	switch value.Kind() {
	case reflect.Interface:
		if !value.IsNil() {
			walkValue(value.Elem(), visit)
		}
	case reflect.Slice:
		for index := 0; index < value.Len(); index++ {
			walkValue(value.Index(index), visit)
		}
	case reflect.Array:
		for index := 0; index < value.Len(); index++ {
			walkValue(value.Index(index), visit)
		}
	case reflect.Struct:
		if value.Type().Implements(expressionType) || value.Type().Implements(statementType) {
			if !visit(value.Interface()) {
				return
			}
		}
		for index := 0; index < value.NumField(); index++ {
			walkValue(value.Field(index), visit)
		}
	}
}
