package flow

import (
	"github.com/system-inc/adamic/internal/ir"
	"reflect"
	"slices"
)

// NormalizeAsync exposes awaits and yields after ordinary lowering. All operands before a suspension
// are snapshots, including mutable reads. Conditional operands remain conditional. A generator's body
// is normalized as an async body is, so every local lives in a frame cell (rule 3: anything live
// across a yield is a frame local, never a C local).
func NormalizeAsync(program *ir.Program) error {
	if containsAwait(reflect.ValueOf(program.Main)) {
		index := len(program.Functions)
		// Catch bindings and other lexical module locals move into the wrapper
		// with its body; they must live in the suspension frame, not C locals.
		for local, declared := range program.Locals {
			if !declared.Global && declared.Function < 0 {
				program.Locals[local].Function = index
			}
		}
		program.Functions = append(program.Functions, ir.Function{
			Name:      "module",
			Async:     true,
			AsyncSite: asyncSourceSite(program.Main),
			Returns:   ir.Promise,
			Body:      program.Main,
		})
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
	for index := range program.Functions {
		function := &program.Functions[index]
		if !function.Suspends() {
			continue
		}
		if function.NestedFrame || function.FrameIdentity != 0 {
			return &AsyncError{
				Site: function.AsyncSite,
				What: suspending(*function, "nested declarations need a canonical suspension environment"),
			}
		}
		normalizer := asyncNormalizer{program: program, function: index}
		completion := asyncCompletion{n: &normalizer, loops: map[int]bool{}, names: map[int]string{}}
		completed := completion.statements(function.Body)
		body, statementsError := normalizer.statements(append(completion.locals, completed...))
		if statementsError != nil {
			return statementsError
		}
		var cells []int
		for local, declared := range program.Locals {
			if !declared.Global && declared.Function == index {
				if declared.Type == ir.MaybeBoolean {
					return &AsyncError{
						Site: function.AsyncSite,
						What: suspending(*function, "slots holding boolean or undefined are not yet proven"),
					}
				}
				if declared.Captured && !declared.IterationCell {
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
var yieldType = reflect.TypeOf(ir.Yield{})

// suspending names what a stop is about: an async function's, or a generator's.
func suspending(function ir.Function, what string) string {
	if function.Generator {
		return "generator " + what
	}
	return "async " + what
}

// containsAwait finds a suspension: an await, or a generator's yield.
func containsAwait(value reflect.Value) bool {
	if !value.IsValid() {
		return false
	}
	if value.Type() == awaitType || value.Type() == yieldType {
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

func (normalizer *asyncNormalizer) temporary(of ir.Type) int {
	local := len(normalizer.program.Locals)
	normalizer.program.Locals = append(
		normalizer.program.Locals,
		ir.Local{Name: "suspension", Type: of, Function: normalizer.function},
	)
	return local
}
func (normalizer *asyncNormalizer) snapshot(value ir.Expression, out *[]ir.Statement) ir.Expression {
	if value.Type() == 0 {
		*out = append(*out, ir.Evaluate{Value: value})
		return ir.Undefined{}
	}
	local := normalizer.temporary(value.Type())
	*out = append(*out, ir.Declare{Local: local, Value: value})
	return ir.Read{Local: local, Of: value.Type()}
}
func (normalizer *asyncNormalizer) expression(value ir.Expression, out *[]ir.Statement, snapshot bool) ir.Expression {
	if value == nil {
		return nil
	}
	suspends := containsAwait(reflect.ValueOf(value))
	if !suspends {
		if snapshot {
			return normalizer.snapshot(value, out)
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
		normalizer.failure = normalizer.errorAt(value, "await among structural method operands is not yet proven")
	}
	switch value := value.(type) {
	case ir.Effects:
		body, statementsError := normalizer.statements(value.Body)
		if statementsError != nil {
			normalizer.failure = statementsError
			return value
		}
		*out = append(*out, body...)
		return normalizer.expression(value.Result, out, snapshot)
	case ir.Logical, ir.LogicalAssignment:
		normalizer.failure = normalizer.errorAt(
			value,
			"await inside value-preserving logical operands is not yet proven",
		)
		return value
	case ir.ArrayLiteral:
		for index, spread := range value.Spread {
			if spread {
				value.Elements[index] = ir.ArraySlice{Array: value.Elements[index], Element: value.Element}
			}
		}
		changed := normalizer.operands(reflect.ValueOf(value), out).Interface().(ir.Expression)
		if snapshot {
			return normalizer.snapshot(changed, out)
		}
		return changed
	case ir.ObjectLiteral:
		if value.Spread != nil {
			value.Spread = ir.ObjectLiteral{
				Spread:               value.Spread,
				NoReuse:              true,
				SpreadMaybeUndefined: value.SpreadMaybeUndefined,
				Empty:                value.Empty,
			}
		}
		changed := normalizer.operands(reflect.ValueOf(value), out).Interface().(ir.Expression)
		if snapshot {
			return normalizer.snapshot(changed, out)
		}
		return changed
	case ir.Await:
		operand := normalizer.expression(value.Value, out, true)
		return normalizer.snapshot(ir.Await{Value: operand, Of: value.Of, Site: value.Site}, out)
	case ir.Conditional:
		condition := normalizer.expression(value.Condition, out, true)
		local := normalizer.temporary(value.Type())
		*out = append(*out, ir.Declare{Local: local})
		var consequent, alternate []ir.Statement
		yes := normalizer.expression(value.WhenTrue, &consequent, false)
		no := normalizer.expression(value.WhenNot, &alternate, false)
		consequent = append(consequent, ir.Assign{Local: local, Value: yes})
		alternate = append(alternate, ir.Assign{Local: local, Value: no})
		*out = append(*out, ir.If{Condition: condition, Then: consequent, Else: alternate})
		return ir.Read{Local: local, Of: value.Type()}
	case ir.Coalesce:
		held := normalizer.expression(value.Value, out, true)
		local := normalizer.temporary(value.Of)
		result := ir.Read{Local: local, Of: value.Of}
		*out = append(*out, ir.Declare{Local: local})
		var missing []ir.Statement
		fallback := normalizer.expression(value.Fallback, &missing, false)
		if value.Panic != nil {
			message := normalizer.expression(value.Panic, &missing, false)
			missing = append(missing, ir.Panic{Message: message})
		} else {
			missing = append(missing, ir.Assign{Local: local, Value: fallback})
		}
		present := []ir.Statement{
			ir.Assign{Local: local, Value: ir.Coalesce{Value: held, Fallback: result, Of: value.Of}},
		}
		*out = append(*out, ir.If{Condition: ir.IsUndefined{Value: held}, Then: missing, Else: present})
		return result
	case ir.Binary:
		if value.Operator == ir.And || value.Operator == ir.Or {
			if value.Left.Type() == ir.MaybeBoolean || value.Right.Type() == ir.MaybeBoolean {
				normalizer.failure = normalizer.errorAt(
					value,
					"await inside maybe-boolean logical operands is not yet proven",
				)
				return value
			}
			condition := normalizer.expression(value.Left, out, false)
			local := normalizer.temporary(ir.Boolean)
			*out = append(*out, ir.Declare{Local: local, Value: condition})
			var body []ir.Statement
			right := normalizer.expression(value.Right, &body, false)
			body = append(body, ir.Assign{Local: local, Value: right})
			condition = ir.Read{Local: local, Of: ir.Boolean}
			if value.Operator == ir.Or {
				condition = ir.Unary{Operator: ir.Not, Operand: condition}
			}
			*out = append(*out, ir.If{Condition: condition, Then: body})
			return ir.Read{Local: local, Of: ir.Boolean}
		}
	}
	changed := normalizer.operands(reflect.ValueOf(value), out).Interface().(ir.Expression)
	if snapshot {
		return normalizer.snapshot(changed, out)
	}
	return changed
}

// operands walks the IR's evaluation-ordered fields. It copies IR, never lowers source.
func (normalizer *asyncNormalizer) operands(value reflect.Value, out *[]ir.Statement) reflect.Value {
	if value.Kind() == reflect.Interface {
		if value.IsNil() {
			return value
		}
		if expression, ok := value.Interface().(ir.Expression); ok {
			changed := normalizer.expression(expression, out, true)
			result := reflect.New(value.Type()).Elem()
			if changed != nil {
				result.Set(reflect.ValueOf(changed))
			}
			return result
		}
		result := reflect.New(value.Type()).Elem()
		result.Set(normalizer.operands(value.Elem(), out))
		return result
	}
	switch value.Kind() {
	case reflect.Struct:
		result := reflect.New(value.Type()).Elem()
		for index := 0; index < value.NumField(); index++ {
			result.Field(index).Set(normalizer.operands(value.Field(index), out))
		}
		return result
	case reflect.Slice:
		if value.IsNil() {
			return value
		}
		result := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
		for index := 0; index < value.Len(); index++ {
			result.Index(index).Set(normalizer.operands(value.Index(index), out))
		}
		return result
	case reflect.Array:
		result := reflect.New(value.Type()).Elem()
		for index := 0; index < value.Len(); index++ {
			result.Index(index).Set(normalizer.operands(value.Index(index), out))
		}
		return result
	}
	return value
}
func (normalizer *asyncNormalizer) statements(statements []ir.Statement) ([]ir.Statement, error) {
	var out []ir.Statement
	for _, statement := range statements {
		var before []ir.Statement
		var statementsError error
		switch value := statement.(type) {
		case ir.Return:
			value.Value = normalizer.expression(value.Value, &before, true)
			statement = value
		case ir.If:
			value.Condition = normalizer.expression(value.Condition, &before, true)
			value.Then, statementsError = normalizer.statements(value.Then)
			if statementsError == nil {
				value.Else, statementsError = normalizer.statements(value.Else)
			}
			statement = value
		case ir.Loop:
			// Body declarations (including nested blocks and catch bindings) are
			// new bindings on each execution, just like for-header bindings.
			mark := func(local int) {
				if normalizer.program.Locals[local].Captured {
					normalizer.program.Locals[local].IterationCell = true
					normalizer.program.Locals[local].EnvironmentCell = false
				}
			}
			inspectAsyncIR(reflect.ValueOf([][]ir.Statement{value.Test, value.Body, value.Update}), func(node any) {
				switch binding := node.(type) {
				case ir.Declare:
					mark(binding.Local)
				case ir.Try:
					if binding.HasCatch && binding.CatchLocal >= 0 {
						mark(binding.CatchLocal)
					}
				}
			})
			var copies []ir.Statement
			for _, local := range value.PerIteration {
				mark(local)
				if normalizer.program.Locals[local].Captured {
					copies = append(copies, ir.Declare{
						Local:    local,
						CopyCell: true,
						Value:    ir.Read{Local: local, Of: normalizer.program.Locals[local].Type},
					})
				}
			}
			value.Update = append(copies, value.Update...)
			value.PerIteration = nil
			value.Condition = normalizer.expression(value.Condition, &value.Test, true)
			value.Body, statementsError = normalizer.statements(value.Body)
			if statementsError == nil {
				value.Update, statementsError = normalizer.statements(value.Update)
			}
			statement = value
		case ir.Block:
			value.Body, statementsError = normalizer.statements(value.Body)
			statement = value
		case ir.Try:
			if value.HasFinally {
				return nil, normalizer.errorAt(value, "async finally escaped completion normalization")
			}
			value.Body, statementsError = normalizer.statements(value.Body)
			if statementsError == nil {
				value.Catch, statementsError = normalizer.statements(value.Catch)
			}
			statement = value
		case ir.Labeled:
			value.Body, statementsError = normalizer.statements(value.Body)
			statement = value
		case ir.Evaluate:
			if yield, ok := value.Value.(ir.Yield); ok {
				// The yielded value is a snapshot, so the yield's own instruction is a read that can't
				// throw (flow.Build's yield) and the value is held by a cell until it is handed over.
				yield.Value = normalizer.expression(yield.Value, &before, true)
				statement = ir.Evaluate{Value: yield}
				break
			}
			if containsAwait(reflect.ValueOf(statement)) {
				statement = normalizer.operands(reflect.ValueOf(statement), &before).Interface().(ir.Statement)
			}
		case ir.ForOf:
			if value.RegexIterator || (value.Iterable.Type() != ir.Array &&
				value.Iterable.Type() != ir.String &&
				value.Iterable.Type() != ir.Map) || (value.Pattern != nil && value.MapPart == "") {
				return nil, normalizer.errorAt(value, suspending(
					normalizer.program.Functions[normalizer.function],
					"for-of over this iterable or pattern is not yet proven",
				))
			}
			if value.Iterable.Type() != ir.Array {
				iterable := normalizer.expression(value.Iterable, &before, true)
				var iterator ir.Expression = ir.StringIterator{Value: iterable}
				element := value.Element
				if value.MapPart != "" {
					iterator = ir.CollectionIterator{
						Collection: iterable,
						Part:       value.MapPart,
						Key:        value.Key,
						Value:      value.Value,
					}
					switch value.MapPart {
					case "keys":
						element = value.Key
					case "values":
						element = value.Value
					default:
						element = ir.Object
					}
				}
				held := normalizer.snapshot(iterator, &before)
				stepLocal := normalizer.temporary(ir.Object)
				step := ir.Read{Local: stepLocal, Of: ir.Object}
				body := []ir.Statement{
					ir.Declare{Local: stepLocal, Value: ir.CallClosure{
						Closure: ir.Property{Object: held, Name: "next", Of: ir.Closure},
						Returns: ir.Object,
					}},
					ir.If{
						Condition: ir.Property{Object: step, Name: "done", Of: ir.Boolean},
						Then:      []ir.Statement{ir.Break{}},
					},
				}
				var item ir.Expression = ir.Property{Object: step, Name: "value", Of: element}
				if element == ir.Number {
					item = ir.Unwrap{Value: ir.Property{Object: step, Name: "value", Of: ir.MaybeNumber}}
				}
				if value.Pattern == nil {
					body = append(body, ir.Declare{Local: value.Local, Value: item})
				} else {
					for _, binding := range value.Pattern {
						body = append(body, ir.Declare{Local: binding.Local, Value: ir.Property{
							Object: item,
							Name:   binding.Field,
							Of:     normalizer.program.Locals[binding.Local].Type,
						}})
					}
				}
				body = append(body, value.Body...)
				normalized, failure := normalizer.statements([]ir.Statement{
					ir.Loop{Labels: value.Labels, Condition: ir.BooleanConstant{Value: true}, Body: body},
				})
				if failure != nil {
					return nil, failure
				}
				out = append(out, before...)
				out = append(out, normalized...)
				continue
			}
			// Hold the array itself, not a copy or its current length. The next pass
			// observes mutations made while suspended, like an array iterator on Node.
			array := normalizer.expression(value.Iterable, &before, true)
			index := normalizer.temporary(ir.Number)
			position := ir.Read{Local: index, Of: ir.Number}
			before = append(before, ir.Declare{Local: index, Value: ir.NumberConstant{Value: 0}})
			var element ir.Expression = ir.ArrayIndex{Array: array, Index: position, Element: value.Element}
			if element.Type() != value.Element {
				// Only numbers and booleans use presence pairs. References already
				// have their element representation, and maybe elements stay maybe.
				element = ir.Unwrap{Value: element}
			}
			loop := ir.Loop{
				Labels:    value.Labels,
				Condition: ir.Binary{Operator: ir.Less, Left: position, Right: ir.Length{Array: array}},
				Body:      append([]ir.Statement{ir.Declare{Local: value.Local, Value: element}}, value.Body...),
				Update: []ir.Statement{ir.Assign{
					Local: index,
					Value: ir.Binary{Operator: ir.Add, Left: position, Right: ir.NumberConstant{Value: 1}},
				}},
			}
			// Loop normalization marks the element declaration as a fresh iteration
			// cell when captured, so closures keep their binding across suspension.
			var normalized []ir.Statement
			normalized, statementsError = normalizer.statements([]ir.Statement{loop})
			if statementsError == nil {
				statement = normalized[0]
			}
		case ir.Switch:
			value.Value = normalizer.expression(value.Value, &before, true)
			for index := range value.Cases {
				if containsAwait(reflect.ValueOf(value.Cases[index].Tests)) {
					return nil, normalizer.errorAt(value, "await in switch case tests is not yet proven")
				}
				value.Cases[index].Body, statementsError = normalizer.statements(value.Cases[index].Body)
				if statementsError != nil {
					break
				}
			}
			if statementsError == nil {
				value.Default, statementsError = normalizer.statements(value.Default)
			}
			statement = value
		default:
			if containsAwait(reflect.ValueOf(statement)) {
				statement = normalizer.operands(reflect.ValueOf(statement), &before).Interface().(ir.Statement)
			}
		}
		if statementsError != nil {
			return nil, statementsError
		}
		if normalizer.failure != nil {
			return nil, normalizer.failure
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

// AsyncError preserves the construct's source site across the IR-only pass.
type AsyncError struct {
	Site int
	What string
}

func (asyncError *AsyncError) Error() string { return asyncError.What }
func asyncSourceSite(value any) int {
	site := 0
	inspectAsyncIR(reflect.ValueOf(value), func(node any) {
		if awaited, ok := node.(ir.Await); ok && site == 0 {
			site = awaited.Site
		}
		if yielded, ok := node.(ir.Yield); ok && site == 0 {
			site = yielded.Site
		}
	})
	return site
}
func (normalizer *asyncNormalizer) errorAt(value any, what string) error {
	site := asyncSourceSite(value)
	if site == 0 {
		site = normalizer.program.Functions[normalizer.function].AsyncSite
	}
	return &AsyncError{Site: site, What: what}
}
