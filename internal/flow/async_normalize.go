package flow

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"reflect"
	"slices"
)

// NormalizeAsync exposes awaits after ordinary lowering. All operands before a suspension
// are snapshots, including mutable reads. Conditional operands remain conditional.
func NormalizeAsync(program *ir.Program) error {
	if containsAwait(reflect.ValueOf(program.Main)) {
		index := len(program.Functions)
		program.Functions = append(program.Functions, ir.Function{Name: "module", Async: true, Returns: ir.Promise, Body: program.Main})
		program.AsyncEntry = index + 1
		program.Main = nil
		// The top level's own locals (a catch binding, a block's let) belong to the module
		// function now, so they get frame cells and the frame releases them. Globals stay global.
		for local := range program.Locals {
			if program.Locals[local].Function == -1 && !program.Locals[local].Global {
				program.Locals[local].Function = index
			}
		}
	}
	if err := checkAsyncTaskEnvironments(program); err != nil {
		return err
	}
	for index := range program.Functions {
		function := &program.Functions[index]
		if !function.Async {
			continue
		}
		normalizer := asyncNormalizer{program: program, function: index}
		body, err := normalizer.statements(function.Body)
		if err != nil {
			return err
		}
		var cells []int
		for local, declared := range program.Locals {
			if !declared.Global && declared.Function == index {
				if declared.Type == ir.MaybeBoolean {
					return fmt.Errorf("async slots holding boolean or undefined are not yet proven")
				}
				if declared.Captured {
					program.Locals[local].Preallocated = true
					program.Locals[local].EnvironmentCell = true
				}
				cells = append(cells, local)
			}
		}
		slices.Sort(cells)
		function.FrameEnvironment = cells
		// Replace the lexical capture site with its suspension-capable environment layout.
		if len(body) > 0 {
			if _, ok := body[0].(ir.AllocateEnvironment); ok {
				body = body[1:]
			}
		}
		function.Body = append([]ir.Statement{ir.AllocateEnvironment{Cells: cells}}, body...)
	}
	if program.HasAsync() {
		program.Generated = ir.AsyncGeneratedTypes()
	}
	return nil
}

type asyncNormalizer struct {
	program  *ir.Program
	function int
	failure  error
}

var awaitType = reflect.TypeOf(ir.Await{})

func containsAwait(value reflect.Value) bool {
	if !value.IsValid() {
		return false
	}
	if value.Type() == awaitType {
		return true
	}
	switch value.Kind() {
	case reflect.Interface, reflect.Pointer:
		return !value.IsNil() && containsAwait(value.Elem())
	case reflect.Struct:
		for index := 0; index < value.NumField(); index++ {
			if containsAwait(value.Field(index)) {
				return true
			}
		}
	case reflect.Slice, reflect.Array:
		for index := 0; index < value.Len(); index++ {
			if containsAwait(value.Index(index)) {
				return true
			}
		}
	}
	return false
}

func (n *asyncNormalizer) temporary(of ir.Type) int {
	local := len(n.program.Locals)
	n.program.Locals = append(n.program.Locals, ir.Local{Name: "suspension", Type: of, Function: n.function})
	return local
}
func (n *asyncNormalizer) snapshot(value ir.Expression, out *[]ir.Statement) ir.Expression {
	if value.Type() == 0 {
		*out = append(*out, ir.Evaluate{Value: value})
		return nil
	}
	local := n.temporary(value.Type())
	*out = append(*out, ir.Declare{Local: local, Value: value})
	return ir.Read{Local: local, Of: value.Type()}
}
func (n *asyncNormalizer) expression(value ir.Expression, out *[]ir.Statement, snapshot bool) ir.Expression {
	if value == nil {
		return nil
	}
	suspends := containsAwait(reflect.ValueOf(value))
	if !suspends {
		if snapshot {
			return n.snapshot(value, out)
		}
		return value
	}
	// A structural method needs a bound receiver and callee snapshot together.
	// Do not silently turn that operation into a plain property read.
	var methodOperand bool
	inspectAsyncIR(reflect.ValueOf(value), func(node any) {
		if property, ok := node.(ir.Property); ok && property.Method {
			methodOperand = true
		}
	})
	if methodOperand {
		n.failure = fmt.Errorf("await among structural method operands is not yet proven")
	}
	switch value := value.(type) {
	case ir.ArrayLiteral:
		for index, spread := range value.Spread {
			if spread {
				value.Elements[index] = ir.ArraySlice{Array: value.Elements[index]}
			}
		}
		changed := n.operands(reflect.ValueOf(value), out).Interface().(ir.Expression)
		if snapshot {
			return n.snapshot(changed, out)
		}
		return changed
	case ir.ObjectLiteral:
		if value.Spread != nil {
			value.Spread = ir.ObjectLiteral{Spread: value.Spread, NoReuse: true, SpreadMaybeUndefined: value.SpreadMaybeUndefined, Empty: value.Empty}
		}
		changed := n.operands(reflect.ValueOf(value), out).Interface().(ir.Expression)
		if snapshot {
			return n.snapshot(changed, out)
		}
		return changed
	case ir.Await:
		operand := n.expression(value.Value, out, true)
		return n.snapshot(ir.Await{Value: operand, Of: value.Of}, out)
	case ir.Conditional:
		condition := n.expression(value.Condition, out, false)
		local := n.temporary(value.Type())
		*out = append(*out, ir.Declare{Local: local})
		var consequent, alternate []ir.Statement
		yes := n.expression(value.WhenTrue, &consequent, false)
		no := n.expression(value.WhenNot, &alternate, false)
		consequent = append(consequent, ir.Assign{Local: local, Value: yes})
		alternate = append(alternate, ir.Assign{Local: local, Value: no})
		*out = append(*out, ir.If{Condition: condition, Then: consequent, Else: alternate})
		return ir.Read{Local: local, Of: value.Type()}
	case ir.Coalesce:
		held := n.expression(value.Value, out, true)
		local := n.temporary(value.Of)
		result := ir.Read{Local: local, Of: value.Of}
		*out = append(*out, ir.Declare{Local: local})
		var missing []ir.Statement
		fallback := n.expression(value.Fallback, &missing, false)
		if value.Panic != nil {
			message := n.expression(value.Panic, &missing, false)
			missing = append(missing, ir.Panic{Message: message})
		} else {
			missing = append(missing, ir.Assign{Local: local, Value: fallback})
		}
		present := []ir.Statement{ir.Assign{Local: local, Value: ir.Coalesce{Value: held, Fallback: result, Of: value.Of}}}
		*out = append(*out, ir.If{Condition: ir.IsUndefined{Value: held}, Then: missing, Else: present})
		return result
	case ir.Binary:
		if value.Operator == ir.And || value.Operator == ir.Or {
			condition := n.expression(value.Left, out, false)
			local := n.temporary(ir.Boolean)
			*out = append(*out, ir.Declare{Local: local, Value: condition})
			var body []ir.Statement
			right := n.expression(value.Right, &body, false)
			body = append(body, ir.Assign{Local: local, Value: right})
			condition = ir.Read{Local: local, Of: ir.Boolean}
			if value.Operator == ir.Or {
				condition = ir.Unary{Operator: ir.Not, Operand: condition}
			}
			*out = append(*out, ir.If{Condition: condition, Then: body})
			return ir.Read{Local: local, Of: ir.Boolean}
		}
	}
	changed := n.operands(reflect.ValueOf(value), out).Interface().(ir.Expression)
	if snapshot {
		return n.snapshot(changed, out)
	}
	return changed
}

// operands walks the IR's evaluation-ordered fields. It copies IR, never lowers source.
func (n *asyncNormalizer) operands(value reflect.Value, out *[]ir.Statement) reflect.Value {
	if value.Kind() == reflect.Interface {
		if value.IsNil() {
			return value
		}
		if expression, ok := value.Interface().(ir.Expression); ok {
			changed := n.expression(expression, out, true)
			result := reflect.New(value.Type()).Elem()
			if changed != nil {
				result.Set(reflect.ValueOf(changed))
			}
			return result
		}
		result := reflect.New(value.Type()).Elem()
		result.Set(n.operands(value.Elem(), out))
		return result
	}
	switch value.Kind() {
	case reflect.Struct:
		result := reflect.New(value.Type()).Elem()
		for index := 0; index < value.NumField(); index++ {
			result.Field(index).Set(n.operands(value.Field(index), out))
		}
		return result
	case reflect.Slice:
		if value.IsNil() {
			return value
		}
		result := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
		for index := 0; index < value.Len(); index++ {
			result.Index(index).Set(n.operands(value.Index(index), out))
		}
		return result
	case reflect.Array:
		result := reflect.New(value.Type()).Elem()
		for index := 0; index < value.Len(); index++ {
			result.Index(index).Set(n.operands(value.Index(index), out))
		}
		return result
	}
	return value
}
func (n *asyncNormalizer) statements(statements []ir.Statement) ([]ir.Statement, error) {
	var out []ir.Statement
	for _, statement := range statements {
		var before []ir.Statement
		var err error
		switch value := statement.(type) {
		case ir.Return:
			value.Value = n.expression(value.Value, &before, true)
			statement = value
		case ir.If:
			value.Condition = n.expression(value.Condition, &before, true)
			value.Then, err = n.statements(value.Then)
			if err == nil {
				value.Else, err = n.statements(value.Else)
			}
			statement = value
		case ir.Loop:
			// Body declarations (including nested blocks and catch bindings) are
			// new bindings on each execution, just like for-header bindings.
			var captured bool
			inspectAsyncIR(reflect.ValueOf([][]ir.Statement{value.Test, value.Body, value.Update}), func(node any) {
				switch binding := node.(type) {
				case ir.Declare:
					captured = captured || n.program.Locals[binding.Local].Captured
				case ir.Try:
					if binding.HasCatch && binding.CatchLocal >= 0 {
						captured = captured || n.program.Locals[binding.CatchLocal].Captured
					}
				}
			})
			if captured {
				return nil, fmt.Errorf("async per-iteration captured cells in a repeatedly executed body are not yet represented")
			}
			for _, local := range value.PerIteration {
				if n.program.Locals[local].Captured {
					return nil, fmt.Errorf("async per-iteration captured cells are not yet represented")
				}
			}
			value.Condition = n.expression(value.Condition, &value.Test, true)
			value.Body, err = n.statements(value.Body)
			if err == nil {
				value.Update, err = n.statements(value.Update)
			}
			statement = value
		case ir.Block:
			value.Body, err = n.statements(value.Body)
			statement = value
		case ir.Try:
			if containsAwait(reflect.ValueOf(value.Catch)) || containsAwait(reflect.ValueOf(value.Finally)) {
				return nil, fmt.Errorf("await in catch or finally is not yet proven")
			}
			if value.HasFinally {
				return nil, fmt.Errorf("async finally completion routing is not yet proven")
			}
			value.Body, err = n.statements(value.Body)
			if err == nil {
				value.Catch, err = n.statements(value.Catch)
			}
			statement = value
		case ir.ForOf:
			if value.RegexIterator || (value.Iterable.Type() != ir.Array && value.Iterable.Type() != ir.String && value.Iterable.Type() != ir.Map) || (value.Pattern != nil && value.MapPart == "") {
				return nil, fmt.Errorf("async for-of over this iterable or pattern is not yet proven")
			}
			if value.Pattern == nil && n.program.Locals[value.Local].Captured {
				return nil, fmt.Errorf("async per-iteration captured cells are not yet represented")
			}
			for _, binding := range value.Pattern {
				if n.program.Locals[binding.Local].Captured {
					return nil, fmt.Errorf("async per-iteration captured cells are not yet represented")
				}
			}
			var outerBreak bool
			inspectAsyncIR(reflect.ValueOf(value.Body), func(node any) {
				if jump, ok := node.(ir.Break); ok && jump.Depth > 0 {
					outerBreak = true
				}
			})
			if outerBreak && containsAwait(reflect.ValueOf(value.Body)) {
				return nil, fmt.Errorf("async for-of labeled outer break is not yet proven")
			}
			if value.Iterable.Type() != ir.Array {
				iterable := n.expression(value.Iterable, &before, true)
				var iterator ir.Expression = ir.StringIterator{Value: iterable}
				element := value.Element
				if value.MapPart != "" {
					iterator = ir.CollectionIterator{Collection: iterable, Part: value.MapPart, Key: value.Key, Value: value.Value}
					switch value.MapPart {
					case "keys":
						element = value.Key
					case "values":
						element = value.Value
					default:
						element = ir.Object
					}
				}
				held := n.snapshot(iterator, &before)
				stepLocal := n.temporary(ir.Object)
				step := ir.Read{Local: stepLocal, Of: ir.Object}
				body := []ir.Statement{
					ir.Declare{Local: stepLocal, Value: ir.CallClosure{Closure: ir.Property{Object: held, Name: "next", Of: ir.Closure}, Returns: ir.Object}},
					ir.If{Condition: ir.Property{Object: step, Name: "done", Of: ir.Boolean}, Then: []ir.Statement{ir.Break{}}},
				}
				var item ir.Expression = ir.Property{Object: step, Name: "value", Of: element}
				if element == ir.Number {
					item = ir.Unwrap{Value: ir.Property{Object: step, Name: "value", Of: ir.MaybeNumber}}
				}
				if value.Pattern == nil {
					body = append(body, ir.Declare{Local: value.Local, Value: item})
				} else {
					for _, binding := range value.Pattern {
						body = append(body, ir.Declare{Local: binding.Local, Value: ir.Property{Object: item, Name: binding.Field, Of: n.program.Locals[binding.Local].Type}})
					}
				}
				body = append(body, value.Body...)
				normalized, failure := n.statements([]ir.Statement{ir.Loop{Condition: ir.BooleanConstant{Value: true}, Body: body}})
				if failure != nil {
					return nil, failure
				}
				out = append(out, before...)
				out = append(out, normalized...)
				continue
			}
			// Hold the array itself, not a copy or its current length. The next pass
			// observes mutations made while suspended, like an array iterator on Node.
			array := n.expression(value.Iterable, &before, true)
			index := n.temporary(ir.Number)
			position := ir.Read{Local: index, Of: ir.Number}
			before = append(before, ir.Declare{Local: index, Value: ir.NumberConstant{Value: 0}})
			var element ir.Expression = ir.ArrayIndex{Array: array, Index: position, Element: value.Element}
			if element.Type() != value.Element {
				// Only numbers and booleans use presence pairs. References already
				// have their element representation, and maybe elements stay maybe.
				element = ir.Unwrap{Value: element}
			}
			loop := ir.Loop{
				Condition: ir.Binary{Operator: ir.Less, Left: position, Right: ir.Length{Array: array}},
				Body:      append([]ir.Statement{ir.Declare{Local: value.Local, Value: element}}, value.Body...),
				Update:    []ir.Statement{ir.Assign{Local: index, Value: ir.Binary{Operator: ir.Add, Left: position, Right: ir.NumberConstant{Value: 1}}}},
			}
			var normalized []ir.Statement
			normalized, err = n.statements([]ir.Statement{loop})
			if err == nil {
				statement = normalized[0]
			}
		case ir.Switch:
			value.Value = n.expression(value.Value, &before, true)
			for index := range value.Cases {
				if containsAwait(reflect.ValueOf(value.Cases[index].Tests)) {
					return nil, fmt.Errorf("await in switch case tests is not yet proven")
				}
				value.Cases[index].Body, err = n.statements(value.Cases[index].Body)
				if err != nil {
					break
				}
			}
			if err == nil {
				value.Default, err = n.statements(value.Default)
			}
			statement = value
		default:
			if containsAwait(reflect.ValueOf(statement)) {
				statement = n.operands(reflect.ValueOf(statement), &before).Interface().(ir.Statement)
			}
		}
		if err != nil {
			return nil, err
		}
		if n.failure != nil {
			return nil, n.failure
		}
		out = append(out, before...)
		// A void await used for its effects was already emitted by snapshot.
		if value, ok := statement.(ir.Evaluate); ok && value.Value == nil {
			continue
		}
		out = append(out, statement)
	}
	return out, nil
}

func inspectAsyncIR(value reflect.Value, visit func(any)) {
	if !value.IsValid() {
		return
	}
	switch value.Kind() {
	case reflect.Interface, reflect.Pointer:
		if !value.IsNil() {
			inspectAsyncIR(value.Elem(), visit)
		}
	case reflect.Struct:
		visit(value.Interface())
		for index := 0; index < value.NumField(); index++ {
			inspectAsyncIR(value.Field(index), visit)
		}
	case reflect.Slice, reflect.Array:
		for index := 0; index < value.Len(); index++ {
			inspectAsyncIR(value.Index(index), visit)
		}
	}
}

// A pool cannot share loop-owned protocol roots hidden behind captured cells.
func checkAsyncTaskEnvironments(program *ir.Program) error {
	if !program.HasAsync() {
		return nil
	}
	var failure error
	inspect := func(node any) {
		task, ok := node.(ir.ParallelMap)
		if !ok || failure != nil {
			return
		}
		targets := program.ClosureTargets(ir.CallClosure{Closure: task.Work})
		possible := targets.Functions
		if targets.Unknown {
			possible = nil
			for index := range program.Functions {
				possible = append(possible, index)
			}
		}
		for _, target := range possible {
			for _, captured := range program.Functions[target].Environment {
				declared := program.Locals[captured]
				if !declared.Global && declared.Function >= 0 && program.Functions[declared.Function].Async {
					failure = fmt.Errorf("pool tasks capturing an async environment are not yet proven")
					return
				}
			}
		}
	}
	inspectAsyncIR(reflect.ValueOf(program.Main), inspect)
	for _, function := range program.Functions {
		inspectAsyncIR(reflect.ValueOf(function.Body), inspect)
	}
	return failure
}
