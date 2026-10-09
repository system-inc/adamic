package lower

import (
	"math"
	"reflect"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// Exceptions (docs/memory.md, "Exceptions, designed into counting"): throw new Error(message), a
// caught value thrown again, and try with catch, finally or both. Payloads use the owned
// tagged union representation; catch bindings stay unknown until a real runtime test.

// tryRecord is a try statement lowered, for the checks made once every function is: where it is,
// and its body.
type tryRecord struct {
	node *ast.Node
	body []ir.Statement
}

// throwStatement lowers throw.
func (l *lowering) throwStatement(node *ast.Node) ([]ir.Statement, error) {
	thrown := ast.SkipParentheses(node.AsThrowStatement().Expression)
	value, err := l.expression(thrown)
	if err != nil {
		return nil, err
	}
	// Nullable references use NULL in their typed slot. Convert that sentinel to the
	// explicit null tag before entering unknown storage, evaluating the operand once.
	if _, literal := value.(ir.Null); !literal && value.Type() != ir.Union && l.includesNull(l.checker.GetTypeAtLocation(thrown)) {
		b := l.libraryArrayBuilder([]ir.Expression{value})
		held := b.read(b.parameters[0])
		value = b.finish("throw_nullable", ir.Conditional{Condition: ir.IsUndefined{Value: held}, WhenTrue: fit(ir.Null{}, ir.Union), WhenNot: fit(held, ir.Union), Of: ir.Union})
	}
	return []ir.Statement{ir.Throw{Value: fit(value, ir.Union)}}, nil
}

// newError lowers new Error(message), and new Error().
func (l *lowering) newError(node *ast.Node) (ir.Expression, error) {
	created := node.AsNewExpression()
	message := ir.Expression(ir.Undefined{Of: ir.String})
	if created.Arguments != nil && len(created.Arguments.Nodes) > 0 {
		if len(created.Arguments.Nodes) > 1 {
			return nil, l.notYet(node, "new Error with options")
		}
		var err error
		if message, err = l.expression(created.Arguments.Nodes[0]); err != nil {
			return nil, err
		}
		message = fit(message, ir.String)
		if message.Type() != ir.String {
			return nil, l.notYet(node, "new Error with a message that isn't a string")
		}
	}
	constructor := ast.SkipParentheses(created.Expression).Text()
	return ir.MakeError{Message: message, Constructor: constructor}, nil
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

// caughtInstanceOfError recognizes the intrinsic Error constructor. Class -1 is
// its nominal runtime identity, shared by native Error objects and host errors.
func (l *lowering) caughtInstanceOfError(node *ast.Node) (ir.Expression, bool) {
	binary := node.AsBinaryExpression()
	identity := 0
	for index, name := range []string{"Error", "RangeError", "TypeError", "ReferenceError"} {
		if l.isLibraryGlobal(binary.Right, name) {
			identity = -index - 1
		}
	}
	if binary.OperatorToken.Kind != ast.KindInstanceOfKeyword || identity == 0 {
		return nil, false
	}
	value, err := l.expression(binary.Left)
	if err != nil {
		l.unlowerable = err
		return nil, false
	}
	return ir.InstanceOf{Value: value, Class: identity}, true
}

// exceptions works out which functions a throw can leave, once every function is lowered, and
// refuses what can't be done yet: a try that can reach a library call whose failure is a panic
// natively but a throw on Node.
func (l *lowering) exceptions() error {
	l.libraryExceptions()
	functions := l.result.Functions
	// A class's methods are reached through function values too: a call through an interface the
	// class implements calls one where it would call the object's own function value (ir.Property's
	// Method), so a method that can throw makes every such call one that can.
	dispatched := map[int]bool{}
	for _, lowered := range l.instances {
		for _, method := range lowered.methodList() {
			dispatched[method.Function] = true
		}
	}
	for _, class := range l.result.Classes {
		for _, accessor := range class.Accessors {
			if accessor.Getter >= 0 {
				dispatched[accessor.Getter] = true
			}
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
			if !functions[index].MayThrow && l.throwsOut(functions[index].Body) {
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
	return l.throwsOutReadiness(statements, false)
}

// Readiness checks are counted after readiness has proved and removed safe reads.
func (l *lowering) throwsOutReadiness(statements []ir.Statement, readinessChecks bool) bool {
	found := false
	walk(statements, func(node any) bool {
		switch node := node.(type) {
		case ir.Try:
			if node.HasCatch {
				// The body's throws are caught; the catch's and the finally's aren't.
				found = found || l.throwsOutReadiness(node.Catch, readinessChecks) || l.throwsOutReadiness(node.Finally, readinessChecks)
				return false
			}
		case ir.NodeFSFile:
			found = found || node.MayThrow()
		case ir.NodeBufferCall:
			found = found || node.MayThrow()
		case ir.Throw:
			found = true
		case ir.Assign:
			found = found || readinessChecks && node.Throws(l.result)
		case ir.Read:
			found = found || readinessChecks && node.Throws(l.result)
		case ir.Defined:
			found = found || node.Throws()
		case ir.Call:
			if l.result.CallMayThrow(node) {
				found = true
			}
		case ir.RegExpCall:
			if node.Replacement != nil && l.result.ClosuresMayThrow {
				found = true
			}
		case ir.IteratorMethod, ir.IteratorField, ir.CallClosure, ir.ArrayMap, ir.ArrayVisit, ir.ArrayReduce, ir.ArrayFrom, ir.MapForEach:
			// A call through a function value, written out or made by the runtime's loop.
			if l.result.ClosuresMayThrow {
				found = true
			}
		case ir.ArraySort:
			if l.result.ClosureMayThrow(node) {
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
				if !visited[target] && !l.result.Functions[target].CheckedLibrary {
					visited[target] = true
					failing = l.libraryFailure(l.result.Functions[target].Body, visited)
					if failing != "" {
						break
					}
				}
			}
		case ir.IteratorMethod, ir.IteratorField, ir.CallClosure, ir.ArrayMap, ir.ArrayVisit, ir.ArrayReduce, ir.MapForEach:
			callsClosures = true
		case ir.ArraySort:
			if node.Callback != nil {
				callsClosures = true
			} else if !visited[node.Comparator] {
				visited[node.Comparator] = true
				failing = l.libraryFailure(l.result.Functions[node.Comparator].Body, visited)
			}
		case ir.SetProperty:
			if l.objectCanFreeze() {
				failing = "a write to a potentially frozen object"
			}
		case ir.ObjectCall:
			if node.Method == "assign" && l.objectCanFreeze() {
				failing = "Object.assign into a potentially frozen object"
			}
		case ir.RegExpCall:
			if node.Replacement != nil {
				callsClosures = true
			}
			if node.Method == "replaceAll" || node.Method == "matchAll" {
				failing = "RegExp global-flag validation"
			}
		case ir.StringCall:
			switch {
			case node.Method == "repeat" && !constantWithin(node.Arguments[0], 0, math.MaxFloat64):
				failing = "repeat"
			case node.Method == "normalize" && len(node.Arguments) > 0 && !isNormalizationForm(node.Arguments[0], l.result.Strings):
				failing = "normalize"
			}
		case ir.ToFixed:
			if !constantWithin(node.Digits, 0, 100) {
				failing = "toFixed"
			}
		case ir.NumberFormat:
			if bounds := formatArguments[node.Method]; node.Argument != nil && !constantWithin(node.Argument, bounds[0], bounds[1]) {
				failing = node.Method
			}
		case ir.TypedArrayNew:
			if !node.FromArray && !constantWithin(node.Source, 0, 9007199254740991) {
				failing = "typed array length conversion"
			}
		case ir.TypedArraySet:
			failing = "typed array set range validation"
		case ir.ArrayFill:
			if node.Array == nil && !constantWithin(node.Length, 0, 4294967295) {
				failing = "new Array(length)"
			}
		case ir.ArrayFrom:
			callsClosures = true
			if !constantWithin(node.Length, math.Inf(-1), 4294967295) {
				failing = "Array.from({ length })"
			}
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

// formatArguments are the arguments each number format takes without throwing.
var formatArguments = map[string][2]float64{"toExponential": {0, 100}, "toPrecision": {1, 100}, "toString": {2, 36}}

// constantWithin reports whether a value is a constant integer from low to high, which a call taking
// it can't fail on.
func constantWithin(value ir.Expression, low float64, high float64) bool {
	constant, isConstant := value.(ir.NumberConstant)
	return isConstant && constant.Value >= low && constant.Value <= high && (high == math.MaxFloat64 || constant.Value == math.Trunc(constant.Value)) && !math.IsInf(constant.Value, 0)
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
