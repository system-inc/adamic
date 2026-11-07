// Package split proves which closed-program functions can cross a Wasm boundary.
package split

import (
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

// Function is one decision. Index identifies the original IR function.
type Function struct {
	Index    int
	Name     string
	Position ir.SourcePosition
	Eligible bool
	Reason   string
	Entry    bool
}

type roots map[int]bool

func add(into, from roots) bool {
	changed := false
	for root := range from {
		if !into[root] {
			into[root] = true
			changed = true
		}
	}
	return changed
}
func reference(t ir.Type) bool { return t.IsReference() && t != ir.String }

type analysis struct {
	program         *ir.Program
	aliases         []roots
	contents        []roots
	returns, writes []roots
	edges           [][]int
	direct          []string
	globals         roots
	changed         bool
	current         int
}

func (analysis *analysis) merge(into, from roots) {
	if add(into, from) {
		analysis.changed = true
	}
}

// summarize works out every local's aliases and contents and every function's returns and writes,
// the may sets Analyze judges each function by.
func summarize(program *ir.Program) *analysis {
	n := len(program.Functions)
	a := &analysis{
		program:  program,
		aliases:  make([]roots, len(program.Locals)),
		contents: make([]roots, len(program.Locals)),
		returns:  make([]roots, n),
		writes:   make([]roots, n),
		edges:    make([][]int, n),
		direct:   make([]string, n),
		globals:  roots{},
	}
	for i, local := range program.Locals {
		a.aliases[i] = roots{}
		a.contents[i] = roots{}
		if reference(local.Type) {
			a.aliases[i][i] = true
			if local.Global {
				a.contents[i][i] = true
			}
		}
	}
	for i, function := range program.Functions {
		a.returns[i], a.writes[i] = roots{}, roots{}
		for _, parameter := range function.Parameters {
			if reference(program.Locals[parameter].Type) {
				a.aliases[parameter][parameter] = true
				a.contents[parameter][parameter] = true
			}
		}
	}
	// Alias and mutation summaries are may sets. Revisit recursive calls, joins,
	// captures and globals until no origin or write can be added.
	for {
		a.changed = false
		a.current = -1
		a.walk(program.Main)
		for i, function := range program.Functions {
			a.current = i
			a.walk(function.Body)
		}
		if !a.changed {
			break
		}
	}
	return a
}

// Analyze returns every function in declaration position order. Ties retain IR order.
func Analyze(program *ir.Program) []Function {
	n := len(program.Functions)
	a := summarize(program)
	pure := make([]bool, n)
	reasons := make([]string, n)
	work := make([]bool, n)
	for i, function := range program.Functions {
		reasons[i] = a.direct[i]
		a.current = i
		walk(function.Body, func(node any) {
			if read, ok := node.(ir.Read); ok &&
				program.Locals[read.Local].Global &&
				a.globals[read.Local] &&
				reasons[i] == "" {
				reasons[i] = "impure: mutable global " + program.Locals[read.Local].Name
			}
			switch node.(type) {
			case ir.Loop, ir.ForOf:
				work[i] = true
			}
		})
		pure[i] = reasons[i] == ""
		// A path back to this function proves recursion, including mutual recursion.
		seen := map[int]bool{}
		var reaches func(int) bool
		reaches = func(j int) bool {
			for _, target := range a.edges[j] {
				if target == i {
					return true
				}
				if !seen[target] {
					seen[target] = true
					if reaches(target) {
						return true
					}
				}
			}
			return false
		}
		work[i] = work[i] || reaches(i)
	}
	for changed := true; changed; {
		changed = false
		for i := range program.Functions {
			for _, target := range a.edges[i] {
				if pure[i] && !pure[target] {
					pure[i] = false
					reasons[i] = "impure: calls " + program.Functions[target].Name
					changed = true
				}
				if !work[i] && work[target] {
					work[i] = true
					changed = true
				}
			}
		}
	}
	constructors := map[int]bool{}
	for _, class := range program.Classes {
		if class.Constructor >= 0 {
			constructors[class.Constructor] = true
		}
	}
	result := make([]Function, n)
	for i, function := range program.Functions {
		reason := reasons[i]
		if constructors[i] && reason == "" {
			reason = "signature: constructor not crossable"
		}
		if pure[i] && reason == "" {
			for j, parameter := range function.Parameters {
				if reason == "" && program.Locals[parameter].Type.IsTypedArray() {
					if a.writes[i][parameter] {
						reason = fmt.Sprintf("writes parameter %d", j+1)
					} else {
						reason = fmt.Sprintf("signature: typed-array parameter %d not crossable", j+1)
					}
				}
			}
			if reason == "" && function.Returns.IsTypedArray() {
				reason = "signature: typed-array return not crossable"
			}
			if reason == "" && len(function.Boundary.Parameters) != len(function.Parameters) {
				reason = "signature: missing boundary schema"
			}
			for j, schema := range function.Boundary.Parameters {
				if reason == "" && !crossable(schema.Schema) {
					reason = fmt.Sprintf("signature: parameter %d not crossable", j+1)
				}
			}
			if reason == "" && !crossable(function.Boundary.Return) {
				reason = "signature: return not crossable"
			}
			for j, parameter := range function.Parameters {
				if reason == "" && a.writes[i][parameter] {
					reason = fmt.Sprintf("writes parameter %d", j+1)
				}
			}
			if reason == "" && !work[i] {
				reason = "too small to cross"
			}
		}
		result[i] = Function{
			Index:    i,
			Name:     function.Name,
			Position: function.Position,
			Eligible: reason == "",
			Reason:   reason,
		}
	}
	for caller, targets := range a.edges {
		if !result[caller].Eligible {
			for _, target := range targets {
				if result[target].Eligible {
					result[target].Entry = true
				}
			}
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		left, right := result[i].Position, result[j].Position
		if (left.File == "") != (right.File == "") {
			return left.File != ""
		}
		if left.File != right.File {
			return left.File < right.File
		}
		if left.Line != right.Line {
			return left.Line < right.Line
		}
		return left.Column < right.Column
	})
	return result
}

// Print uses a stable one-line decision format.
func Print(output io.Writer, functions []Function) error {
	for _, function := range functions {
		state := "javascript"
		if function.Eligible {
			state = "wasm"
		}
		entry := ""
		if function.Entry {
			entry = " entry"
		}
		reason := ""
		if function.Reason != "" {
			reason = " " + function.Reason
		}
		if _, fprintfError := fmt.Fprintf(
			output,
			"%s %s%s%s\n",
			function.Name,
			state,
			entry,
			reason,
		); fprintfError != nil {
			return fprintfError
		}
	}
	return nil
}

func crossable(schema *ir.JSONDecodeSchema) bool {
	if schema == nil {
		return false
	}
	active := map[int]bool{}
	var visit func(int, int, bool) bool
	visit = func(index, depth int, field bool) bool {
		if index < 0 || index >= len(schema.Nodes) || active[index] {
			return false
		}
		node := schema.Nodes[index]
		switch node.Kind {
		case "number", "boolean", "string":
			return true
		case "literal":
			return node.Of == ir.Number || node.Of == ir.Boolean || node.Of == ir.String
		case "array":
			if len(node.Children) != 1 {
				return false
			}
			child := schema.Nodes[node.Children[0]]
			return child.Kind == "number" ||
				child.Kind == "string" ||
				child.Kind == "boolean" ||
				child.Kind == "literal" && (child.Of == ir.Number || child.Of == ir.Boolean || child.Of == ir.String)
		case "object":
			if depth >= 2 {
				return false
			}
			active[index] = true
			defer delete(active, index)
			for _, property := range node.Fields {
				if !visit(property.Node, depth+1, true) {
					return false
				}
			}
			return true
		case "union":
			if !field || len(node.Children) != 2 {
				return false
			}
			first, second := node.Children[0], node.Children[1]
			if schema.Nodes[first].Kind == "undefined" {
				return visit(second, depth, true)
			}
			if schema.Nodes[second].Kind == "undefined" {
				return visit(first, depth, true)
			}
		}
		return false
	}
	return visit(schema.Root, 0, false)
}

// walk follows all embedded IR expressions and statements, including newly nested
// operands. Nodes are visited before their children in deterministic field order.
func walk(value any, accept func(any)) {
	var visit func(reflect.Value)
	visit = func(v reflect.Value) {
		if !v.IsValid() {
			return
		}
		switch v.Kind() {
		case reflect.Interface, reflect.Pointer:
			if !v.IsNil() {
				visit(v.Elem())
			}
		case reflect.Struct:
			accept(v.Interface())
			for i := 0; i < v.NumField(); i++ {
				visit(v.Field(i))
			}
		case reflect.Slice, reflect.Array:
			for i := 0; i < v.Len(); i++ {
				visit(v.Index(i))
			}
		}
	}
	visit(reflect.ValueOf(value))
}

func (analysis *analysis) fail(reason string) {
	if analysis.current >= 0 && analysis.direct[analysis.current] == "" {
		analysis.direct[analysis.current] = reason
	}
}
func (analysis *analysis) mutate(value roots) {
	if analysis.current >= 0 {
		analysis.merge(analysis.writes[analysis.current], value)
	}
	for root := range value {
		if analysis.program.Locals[root].Global {
			analysis.merge(analysis.globals, roots{root: true})
			analysis.fail("impure: writes global " + analysis.program.Locals[root].Name)
		}
	}
}

func (analysis *analysis) walk(body []ir.Statement) {
	walk(body, func(node any) {
		if expression, ok := node.(ir.Expression); ok && !knownExpression(expression) {
			analysis.fail("impure: unknown operation " + reflect.TypeOf(expression).Name())
		}
		if statement, ok := node.(ir.Statement); ok && !knownStatements[reflect.TypeOf(statement).Name()] {
			analysis.fail("impure: unknown statement " + reflect.TypeOf(statement).Name())
		}
		switch node := node.(type) {
		case ir.Declare:
			analysis.merge(analysis.aliases[node.Local], analysis.origins(node.Value))
			analysis.merge(analysis.contents[node.Local], analysis.contained(node.Value))
		case ir.Assign:
			analysis.merge(analysis.aliases[node.Local], analysis.origins(node.Value))
			analysis.merge(analysis.contents[node.Local], analysis.contained(node.Value))
			if analysis.program.Locals[node.Local].Global {
				analysis.merge(analysis.globals, roots{node.Local: true})
				analysis.fail("impure: writes global " + analysis.program.Locals[node.Local].Name)
			}
		case ir.ForOf:
			value := analysis.reachable(analysis.origins(node.Iterable))
			if len(node.Pattern) == 0 && reference(analysis.program.Locals[node.Local].Type) {
				analysis.merge(analysis.aliases[node.Local], value)
			}
			for _, binding := range node.Pattern {
				if reference(analysis.program.Locals[binding.Local].Type) {
					analysis.merge(analysis.aliases[binding.Local], value)
				}
			}
		case ir.SetIndex:
			analysis.mutate(analysis.origins(node.Array))
			analysis.store(node.Array, node.Value)
		case ir.SetProperty:
			analysis.mutate(analysis.origins(node.Object))
			analysis.store(node.Object, node.Value)
		case ir.Return:
			if analysis.current >= 0 {
				analysis.merge(analysis.returns[analysis.current], analysis.contained(node.Value))
			}
		case ir.Yield:
			// What a generator yields is what a step of its call is, as a return is a call's.
			if analysis.current >= 0 {
				analysis.merge(analysis.returns[analysis.current], analysis.contained(node.Value))
			}
		case ir.WriteLine:
			analysis.fail("impure: console")
		case ir.Panic, ir.Coalesce:
			// Panic is Wasm-safe: it aborts this call in either backend. The
			// bridge turns AdamicPanic or native exit 70 into the same 500,
			// preserving the message. Operand calls still undergo purity checks.
		case ir.ReadTextFile, ir.WriteTextFile, ir.ReadDirectory, ir.FileStatus, ir.ProgramArguments:
			analysis.fail("impure: adamic I/O")
		case ir.Read:
			local := analysis.program.Locals[node.Local]
			if analysis.current >= 0 && !local.Global && local.Function != analysis.current {
				analysis.fail("impure: captured state " + local.Name)
			}
		case ir.LibraryGlobal:
			analysis.fail("impure: external global " + node.Name)
		case ir.WeakTarget:
			analysis.fail("impure: Weak")
		case ir.MathCall:
			if node.Function == "random" {
				analysis.fail("impure: Math.random")
			}
		case ir.ArrayPush:
			analysis.mutate(analysis.origins(node.Array))
			analysis.store(node.Array, node.Value)
		case ir.ArrayPop:
			analysis.mutate(analysis.origins(node.Array))
		case ir.ArraySplice:
			analysis.mutate(analysis.origins(node.Array))
			for _, item := range node.Items {
				analysis.store(node.Array, item)
			}
		case ir.ArrayFill:
			analysis.mutate(analysis.origins(node.Array))
			analysis.store(node.Array, node.Value)
		case ir.ArrayReverse:
			analysis.mutate(analysis.origins(node.Array))
		case ir.MapSet:
			analysis.mutate(analysis.origins(node.Map))
			analysis.store(node.Map, node.Value)
		case ir.ObjectCall:
			if (node.Method == "assign" || node.Method == "freeze") && len(node.Arguments) > 0 {
				analysis.mutate(analysis.origins(node.Arguments[0]))
			}
		case ir.RegExpCall:
			analysis.mutate(analysis.origins(node.Value))
			for _, arg := range node.Arguments {
				analysis.mutate(analysis.origins(arg))
			}
		case ir.MapDelete:
			analysis.mutate(analysis.origins(node.Map))
		case ir.MapClear:
			analysis.mutate(analysis.origins(node.Map))
		case ir.SetAdd:
			analysis.mutate(analysis.origins(node.Set))
			analysis.store(node.Set, node.Value)
		case ir.Call:
			analysis.call(analysis.program.CallTargets(node), node.Arguments)
		case ir.CallClosure:
			analysis.closure(node, node.Arguments)
		case ir.ArrayMap:
			analysis.closure(node, []ir.Expression{node.Array, node.Array, node.Array})
		case ir.ArrayVisit:
			analysis.closure(node, []ir.Expression{node.Array, node.Array, node.Array})
		case ir.ArrayReduce:
			analysis.closure(node, []ir.Expression{node.Initial, node.Array, node.Array, node.Array})
		case ir.ArrayFrom:
			if node.Callback != nil {
				analysis.closure(node, []ir.Expression{nil, node.Length})
			}
		case ir.ArraySort:
			analysis.mutate(analysis.origins(node.Array))
			analysis.closure(node, []ir.Expression{node.Array, node.Array})
		case ir.MapForEach:
			analysis.closure(node, []ir.Expression{node.Map, node.Map, node.Map})
		}
	})
}
func (analysis *analysis) store(container, value ir.Expression) {
	held := analysis.contained(value)
	for root := range analysis.origins(container) {
		analysis.merge(analysis.contents[root], held)
	}
}
func (analysis *analysis) closure(node ir.Expression, args []ir.Expression) {
	targets := analysis.closureTargets(node)
	if targets.Unknown {
		analysis.fail("impure: unknown callee")
		for _, arg := range args {
			analysis.mutate(analysis.origins(arg))
		}
		// An unresolved closure can reach all reference globals.
		for i, local := range analysis.program.Locals {
			if local.Global && reference(local.Type) {
				analysis.mutate(roots{i: true})
			}
		}
		return
	}
	analysis.call(targets.Functions, args)
}
func (analysis *analysis) substitute(target int, value roots, args []ir.Expression) roots {
	result := roots{}
	function := analysis.program.Functions[target]
	for root := range value {
		mapped := false
		for j, parameter := range function.Parameters {
			if root == parameter {
				if j < len(args) {
					add(result, analysis.contained(args[j]))
				}
				mapped = true
			}
		}
		if !mapped {
			add(result, analysis.aliases[root])
			result[root] = true
		}
	}
	return result
}
func (analysis *analysis) call(targets []int, args []ir.Expression) {
	if len(targets) == 0 {
		analysis.fail("impure: unknown callee")
	}
	for _, target := range targets {
		if analysis.current >= 0 {
			found := false
			for _, previous := range analysis.edges[analysis.current] {
				if previous == target {
					found = true
				}
			}
			if !found {
				analysis.edges[analysis.current] = append(analysis.edges[analysis.current], target)
			}
		}
		analysis.mutate(analysis.substitute(target, analysis.writes[target], args))
	}
}
func (analysis *analysis) origins(value ir.Expression) roots {
	result := roots{}
	if value == nil || !reference(value.Type()) {
		return result
	}
	switch node := value.(type) {
	case ir.Read:
		add(result, analysis.aliases[node.Local])
	case ir.Call:
		for _, target := range analysis.program.CallTargets(node) {
			add(result, analysis.substitute(target, analysis.returns[target], node.Arguments))
		}
	case ir.CallClosure:
		targets := analysis.closureTargets(node)
		for _, target := range targets.Functions {
			add(result, analysis.substitute(target, analysis.returns[target], node.Arguments))
		}
		if targets.Unknown {
			for _, arg := range node.Arguments {
				add(result, analysis.origins(arg))
			}
		}
	case ir.MakeClosure:
		for _, local := range analysis.program.Functions[node.Function].Environment {
			add(result, analysis.aliases[local])
		}
	case ir.GeneratorNext:
		// A step is one of the values the generator's call stands for: its yields, substituted at
		// the call that made it.
		add(result, analysis.origins(node.Generator))
	case ir.Property:
		add(result, analysis.elements(node.Object))
	case ir.ArrayIndex:
		add(result, analysis.elements(node.Array))
	case ir.MapGet:
		add(result, analysis.elements(node.Map))
	case ir.ArrayPop:
		add(result, analysis.elements(node.Array))
	case ir.ArrayReduce:
		add(result, analysis.origins(node.Initial))
		for _, target := range analysis.closureTargets(node).Functions {
			add(result, analysis.substitute(
				target,
				analysis.returns[target],
				[]ir.Expression{node.Initial, node.Array, node.Array, node.Array},
			))
		}
	case ir.ArrayReverse:
		add(result, analysis.origins(node.Array))
	case ir.ArraySort:
		add(result, analysis.origins(node.Array))
	case ir.ArrayFill:
		add(result, analysis.origins(node.Array))
	case ir.ArrayVisit:
		if node.Method != "filter" && reference(node.Element) {
			add(result, analysis.elements(node.Array))
		}
	case ir.ObjectCall:
		if (node.Method == "assign" || node.Method == "freeze") && len(node.Arguments) > 0 {
			add(result, analysis.origins(node.Arguments[0]))
		}
	case ir.TypedArrayNew, ir.ArrayLiteral, ir.ObjectLiteral, ir.MapNew, ir.SetNew, ir.ArraySlice, ir.ArrayConcat, ir.ArrayMap, ir.ArrayFrom, ir.ArraySplice, ir.JSONDecode, ir.RegExpNew, ir.ObjectKeys:
		// Allocated containers have their own identity at the binding. Their shared
		// fields and elements are tracked separately by contained.

	default:
		// Propagate reference origins through fields, narrowing, joins and containers.
		v := reflect.ValueOf(value)
		for i := 0; i < v.NumField(); i++ {
			walk(v.Field(i).Interface(), func(child any) {
				if expression, ok := child.(ir.Expression); ok {
					add(result, analysis.origins(expression))
				}
			})
		}
	}
	return result
}

// closureTargets resolves bindings over every definition in the closed program.
// Any parameter, unknown value, or cyclic unresolved alias leaves the call unknown.
func (analysis *analysis) closureTargets(call ir.Expression) targets {
	var value ir.Expression
	switch node := call.(type) {
	case ir.CallClosure:
		value = node.Closure
	case ir.ArrayMap:
		value = node.Callback
	case ir.ArrayVisit:
		value = node.Callback
	case ir.ArrayReduce:
		value = node.Callback
	case ir.ArrayFrom:
		value = node.Callback
	case ir.MapForEach:
		value = node.Callback
	case ir.ArraySort:
		if node.Callback == nil {
			return targets{Functions: []int{node.Comparator}}
		}
		value = node.Callback
	}
	seen := map[int]bool{}
	result := targets{}
	var resolve func(ir.Expression)
	resolve = func(value ir.Expression) {
		switch node := value.(type) {
		case ir.ClosureSelf:
			if analysis.current >= 0 && analysis.program.Functions[analysis.current].Closure {
				resolve(ir.MakeClosure{Function: analysis.current})
			} else {
				result.Unknown = true
			}
		case ir.MakeClosure:
			for _, previous := range result.Functions {
				if previous == node.Function {
					return
				}
			}
			result.Functions = append(result.Functions, node.Function)
		case ir.Read:
			if seen[node.Local] {
				result.Unknown = true
				return
			}
			seen[node.Local] = true
			defer delete(seen, node.Local)
			found := false
			definitions := func(value any) {
				switch binding := value.(type) {
				case ir.Declare:
					if binding.Local == node.Local {
						found = true
						resolve(binding.Value)
					}
				case ir.Assign:
					if binding.Local == node.Local {
						found = true
						resolve(binding.Value)
					}
				}
			}
			walk(analysis.program.Main, definitions)
			for _, function := range analysis.program.Functions {
				for _, parameter := range function.Parameters {
					if parameter == node.Local {
						result.Unknown = true
					}
				}
				walk(function.Body, definitions)
			}
			if !found {
				result.Unknown = true
			}
		case ir.Conditional:
			resolve(node.WhenTrue)
			resolve(node.WhenNot)
		default:
			result.Unknown = true
		}
	}
	resolve(value)
	return result
}

type targets struct {
	Functions []int
	Unknown   bool
}

func (analysis *analysis) scalarArray(value ir.Expression, seen map[int]bool) bool {
	schema, index := analysis.schema(value)
	if schema != nil && schema.Nodes[index].Kind == "array" {
		child := schema.Nodes[schema.Nodes[index].Children[0]]
		return !reference(child.Of)
	}
	switch node := value.(type) {
	case ir.ArrayLiteral:
		return !reference(node.Element)
	case ir.ArrayFill:
		return !reference(node.Element)
	case ir.ArrayMap:
		return !reference(node.Result)
	case ir.ArrayVisit:
		return !reference(node.Element)
	case ir.ArrayFrom:
		return !reference(node.Element)
	case ir.ArraySlice:
		return analysis.scalarArray(node.Array, seen)
	case ir.Read:
		if seen[node.Local] {
			return false
		}
		seen[node.Local] = true
		defer delete(seen, node.Local)
		found, scalar := false, true
		check := func(value any) {
			switch binding := value.(type) {
			case ir.Declare:
				if binding.Local == node.Local {
					found = true
					scalar = scalar && analysis.scalarArray(binding.Value, seen)
				}
			case ir.Assign:
				if binding.Local == node.Local {
					found = true
					scalar = scalar && analysis.scalarArray(binding.Value, seen)
				}
			}
		}
		walk(analysis.program.Main, check)
		for _, function := range analysis.program.Functions {
			walk(function.Body, check)
		}
		return found && scalar
	}
	return false
}
func (analysis *analysis) schema(value ir.Expression) (*ir.JSONDecodeSchema, int) {
	switch node := value.(type) {
	case ir.Read:
		for _, function := range analysis.program.Functions {
			for j, parameter := range function.Parameters {
				if parameter == node.Local && j < len(function.Boundary.Parameters) {
					schema := function.Boundary.Parameters[j].Schema
					if schema != nil {
						return schema, schema.Root
					}
				}
			}
		}
	case ir.Property:
		schema, index := analysis.schema(node.Object)
		if schema != nil {
			for _, field := range schema.Nodes[index].Fields {
				if field.Name == node.Name {
					return schema, field.Node
				}
			}
		}
	}
	return nil, 0
}

// New runtime operations require a purity review. Reflection traverses operands,
// but an unrecognized operation itself cannot silently become pure.
func knownExpression(expression ir.Expression) bool {
	return knownExpressions[reflect.TypeOf(expression).Name()]
}

// JSON decode and encode perform no I/O; their results depend only on arguments.
// Only TypedArrayNew is reviewed; Fill, Sort, Set and sharing Subarray remain unknown.
var knownExpressions = func() map[string]bool {
	result := map[string]bool{}
	for _, name := range strings.Fields(
		`ArrayConcat ArrayFill ArrayFrom ArrayIndex ArrayJoin ArrayLiteral ArrayMap ArrayPop ArrayPush ArrayReduce ArrayReverse ArraySearch ArraySlice ArraySort ArraySplice ArrayVisit Binary BooleanConstant BooleanToString Box Call CallClosure CharCodeAt CheckedCast ClosureSelf Coalesce CodePoints CollectionIterator Concat Conditional Defined FileStatus HasAccessor HasOwn InstanceOf IsNull IsUndefined JSONDecode JSONEncode JSONNull JSONStringify Length LibraryGlobal MakeClosure MakeError MapClear MapDelete MapEntries MapForEach MapGet MapHas MapKeys MapNew MapSet MapSize MapValues MathCall MaybeOf MaybeToString Narrow Null NumberCall NumberConstant NumberFormat NumberToString ObjectCall ObjectKeys ObjectLiteral ProgramArguments Property Read ReadDirectory ReadTextFile RegExpCall RegExpGroup RegExpNew RegExpProperty SetAdd SetNew SetValues StringCall StringConstant StringFromCodes StringIndex StringLength ToFixed Trim TypeOf TypedArrayNew Unary Undefined UnionToString Unwrap Utf8At Utf8Length WeakOf WeakTarget WriteTextFile`,
	) {
		result[name] = true
	}
	return result
}()

var knownStatements = func() map[string]bool {
	result := map[string]bool{}
	for _, name := range strings.Fields(
		"Assign Block Break Continue Declare Evaluate ForOf If Loop Panic Return SetIndex SetProperty Switch Throw Try WriteLine",
	) {
		result[name] = true
	}
	return result
}()

func (analysis *analysis) reachable(value roots) roots {
	result := roots{}
	queue := []int{}
	for root := range value {
		queue = append(queue, root)
	}
	for len(queue) > 0 {
		root := queue[0]
		queue = queue[1:]
		for child := range analysis.contents[root] {
			if !result[child] {
				result[child] = true
				queue = append(queue, child)
			}
		}
	}
	return result
}

// contained follows mutable fields and elements, independently of the outer
// container's identity. Scalar array copies share no mutable storage.
func (analysis *analysis) contained(value ir.Expression) roots {
	result := analysis.origins(value)
	if value == nil || !reference(value.Type()) {
		return result
	}
	add(result, analysis.reachable(result))
	switch node := value.(type) {
	case ir.Read, ir.Call, ir.CallClosure, ir.Property, ir.ArrayIndex, ir.MapGet, ir.ArrayPop:
		return result
	case ir.ArraySlice:
		if !analysis.scalarArray(node.Array, map[int]bool{}) {
			add(result, analysis.elements(node.Array))
		}
		return result
	case ir.ArrayConcat:
		for _, array := range append([]ir.Expression{node.Array}, node.Others...) {
			if !analysis.scalarArray(array, map[int]bool{}) {
				add(result, analysis.elements(array))
			}
		}
		return result
	case ir.ArraySplice:
		if reference(node.Element) {
			add(result, analysis.elements(node.Array))
		}
		return result
	case ir.ArrayMap:
		for _, target := range analysis.closureTargets(node).Functions {
			add(result, analysis.substitute(
				target,
				analysis.returns[target],
				[]ir.Expression{node.Array, node.Array, node.Array},
			))
		}
		return result
	case ir.ArrayFrom:
		for _, target := range analysis.closureTargets(node).Functions {
			add(result, analysis.substitute(target, analysis.returns[target], []ir.Expression{nil, node.Length}))
		}
		return result
	case ir.ArrayVisit:
		if reference(node.Element) {
			add(result, analysis.elements(node.Array))
		}
		return result
	case ir.TypedArrayNew:
		// Length allocation and FromArray scalar copies share no source storage.
		return result
	case ir.JSONDecode, ir.ObjectKeys, ir.RegExpNew:
		return result
	}
	v := reflect.ValueOf(value)
	for i := 0; i < v.NumField(); i++ {
		walk(v.Field(i).Interface(), func(child any) {
			if expression, ok := child.(ir.Expression); ok {
				add(result, analysis.contained(expression))
			}
		})
	}
	return result
}

func (analysis *analysis) elements(value ir.Expression) roots {
	identity := analysis.origins(value)
	if len(identity) == 0 {
		return analysis.contained(value)
	}
	return analysis.reachable(identity)
}
