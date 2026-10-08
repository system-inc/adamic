package lower

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"

	"github.com/system-inc/adamic/internal/ir"
)

// ReachingAllocations queries the shared graph without selecting graph ownership.
// Unknown frontiers retain checks. Joined assignments, returns and arguments use
// the same edges as the graph-regions analysis.
func (graph *allocationFlowGraph) ReachingAllocations(expression ir.Expression) ir.AllocationSet {
	graph.prepareShapeCallbacks()
	return graph.reachingAllocations(expression, true, 0)
}

func (graph *allocationFlowGraph) reachingAllocations(expression ir.Expression, project bool, depth int) (result ir.AllocationSet) {
	if read, ok := expression.(ir.Read); ok && project {
		if graph.projectionReads == nil {
			graph.projectionReads = map[int]ir.AllocationSet{}
			graph.projectionBusy = map[int]bool{}
		}
		if cached, ok := graph.projectionReads[read.Local]; ok {
			return cached
		}
		if graph.projectionBusy[read.Local] {
			return ir.AllocationSet{Unknown: true, Reasons: []string{"recursive projection receiver not certified"}}
		}
		graph.projectionBusy[read.Local] = true
		defer func() { delete(graph.projectionBusy, read.Local); graph.projectionReads[read.Local] = result }()
	}
	result = ir.AllocationSet{}
	sites, visited, reasons := map[int]bool{}, map[int]bool{}, map[string]bool{}
	queue := []int{}
	demand := func(node int) {
		if !visited[node] {
			visited[node] = true
			queue = append(queue, node)
		}
	}
	unknown := func(reason string) { result.Unknown = true; reasons[reason] = true }
	var follow func(ir.Expression, int)
	follow = func(value ir.Expression, level int) {
		if project {
			sources, reasons, recognized := graph.projectedSources(value, level)
			if recognized {
				for _, reason := range reasons {
					unknown(reason)
				}
				for _, source := range sources {
					follow(source, level+1)
				}
				return
			}
			switch wrapper := value.(type) {
			case ir.Conditional:
				follow(wrapper.WhenTrue, level)
				follow(wrapper.WhenNot, level)
				return
			case ir.Box:
				follow(wrapper.Value, level)
				return
			case ir.Narrow:
				follow(wrapper.Value, level)
				return
			case ir.Unwrap:
				follow(wrapper.Value, level)
				return
			case ir.CheckedCast:
				follow(wrapper.Value, level)
				return
			}
		}
		graph.shapeFollow(value, func(site int) { sites[site] = true }, demand, unknown)
	}
	follow(expression, depth)
	for len(queue) != 0 {
		node := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		if graph.callbacks.unknownParameters[node] {
			unknown("parameter may receive untracked or omitted arguments")
		}
		sources := graph.callbacks.sources[node]
		if len(sources) == 0 {
			unknown("local or result has no tracked producer")
		}
		for _, source := range sources {
			if source == nil {
				if node <= len(graph.program.Locals) {
					unknown(fmt.Sprintf("local %d has a producer without a value", node-1))
				} else {
					unknown(fmt.Sprintf("function %d returns without a value", node-len(graph.program.Locals)-1))
				}
				continue
			}
			follow(source, depth)
		}
	}
	if len(sites) == 0 {
		unknown("no reaching allocation established")
	}
	for site := range sites {
		result.Sites = append(result.Sites, site)
	}
	for reason := range reasons {
		result.Reasons = append(result.Reasons, reason)
	}
	sort.Ints(result.Sites)
	sort.Strings(result.Reasons)
	return result
}

// rewriteShapeStatements preserves the IR tree's containers and rewrites only
// interface-held nodes, so metadata never becomes an executable expression.
func rewriteShapeStatements(body []ir.Statement, rewrite func(any) any) []ir.Statement {
	var transform func(reflect.Value) reflect.Value
	transform = func(value reflect.Value) reflect.Value {
		switch value.Kind() {
		case reflect.Interface:
			if value.IsNil() {
				return value
			}
			mapped := transform(value.Elem())
			result := reflect.New(value.Type()).Elem()
			result.Set(reflect.ValueOf(rewrite(mapped.Interface())))
			return result
		case reflect.Struct:
			result := reflect.New(value.Type()).Elem()
			for i := 0; i < value.NumField(); i++ {
				result.Field(i).Set(transform(value.Field(i)))
			}
			return result
		case reflect.Slice:
			if value.IsNil() {
				return value
			}
			result := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
			for i := 0; i < value.Len(); i++ {
				result.Index(i).Set(transform(value.Index(i)))
			}
			return result
		default:
			return value
		}
	}
	return transform(reflect.ValueOf(body)).Interface().([]ir.Statement)
}

// Projection is queried only by shape proofs. Graph-regions traversal stays unchanged.
type allocationProjection struct {
	records            map[int]ir.ObjectLiteral
	arrays             map[int]ir.ArrayLiteral
	stores             map[int]map[string][]ir.Expression
	opaque             map[string]bool
	opaqueSources      map[string][]ir.Expression
	arrayMutation      bool
	opaqueCalls        bool
	dynamicStores      []ir.Expression
	dynamicCache       map[[3]int]shapeDynamicProjectionResult
	dynamicCacheWeight int
	recordSlots        map[int]map[string][]ir.Expression
}

func (graph *allocationFlowGraph) projectionIndex() *allocationProjection {
	if graph.projection != nil {
		return graph.projection
	}
	index := &allocationProjection{records: map[int]ir.ObjectLiteral{}, arrays: map[int]ir.ArrayLiteral{}, stores: map[int]map[string][]ir.Expression{}, opaque: map[string]bool{}, opaqueSources: map[string][]ir.Expression{}}
	graph.projection = index
	stores := []ir.SetProperty{}
	collect := func(node any) bool {
		switch value := node.(type) {
		case ir.ObjectLiteral:
			if len(value.GraphTypes) > 0 {
				index.records[value.GraphTypes[len(value.GraphTypes)-1]] = value
			}
		case ir.ArrayLiteral:
			if len(value.GraphTypes) > 0 {
				index.arrays[value.GraphTypes[len(value.GraphTypes)-1]] = value
			}
		case shapeDynamicMutation:
			index.dynamicStores = append(index.dynamicStores, value.Value)
		case ir.SetProperty:
			stores = append(stores, value)
		case ir.SetIndex, ir.ArrayPush, ir.ArraySplice, ir.ArrayPop:
			index.arrayMutation = true
		case ir.CallClosure:
			targets := graph.shapeClosureTargets(value)
			index.opaqueCalls = index.opaqueCalls || targets.Unknown || len(targets.Functions) == 0
		case ir.ObjectCall:
			index.opaqueCalls = true
		}
		return true
	}
	walk(graph.program.Main, collect)
	for _, function := range graph.program.Functions {
		walk(function.Body, collect)
	}
	for _, store := range stores {
		receiver := graph.reachingAllocations(store.Object, false, 0)
		if receiver.Unknown || store.Uninitialized {
			index.opaque[store.Name] = true
			index.opaqueSources[store.Name] = append(index.opaqueSources[store.Name], store.Object, store.Value)
		}
		for _, site := range receiver.Sites {
			if index.stores[site] == nil {
				index.stores[site] = map[string][]ir.Expression{}
			}
			index.stores[site][store.Name] = append(index.stores[site][store.Name], store.Value)
		}
	}
	return index
}

// projectedSources joins initial fields and all visible stores across all paths.
// An opaque receiver can alias any allocation; its field name stays unknown.
func (graph *allocationFlowGraph) projectedSources(expression ir.Expression, depth int) ([]ir.Expression, []string, bool) {
	if value, ok := expression.(shapeDynamicProjection); ok {
		return graph.dynamicProjectionSources(value.Receiver, value.Key, depth)
	}
	if value, ok := expression.(ir.ArrayIndex); ok && !value.Relative {
		if _, constant := value.Index.(ir.NumberConstant); !constant {
			return graph.dynamicProjectionSources(value.Array, value.Index, depth)
		}
	}
	var receiver ir.Expression
	name := ""
	array := false
	switch value := expression.(type) {
	case ir.Property:
		receiver, name = value.Object, value.Name
	case ir.ArrayIndex:
		receiver = value.Array
		array = true
		key, ok := value.Index.(ir.NumberConstant)
		if !ok || key.Value < 0 || key.Value != float64(int(key.Value)) || value.Relative {
			return nil, []string{"dynamic or relative element key not certified"}, true
		}
		name = strconv.Itoa(int(key.Value))
	default:
		return nil, nil, false
	}
	if depth >= 32 {
		return nil, []string{"recursive projection depth not certified"}, true
	}
	index := graph.projectionIndex()
	sources := []ir.Expression{}
	reasons := []string{}
	if len(index.dynamicStores) != 0 {
		reasons = append(reasons, "indexed store effects not certified")
	}
	if index.opaqueCalls {
		reasons = append(reasons, "opaque call may mutate projected fields")
	}
	if index.opaque[name] {
		reasons = append(reasons, "field store has opaque receiver or clears readiness: "+name)
	}
	if array && index.arrayMutation {
		reasons = append(reasons, "array mutation prevents element certificate")
	}
	bases := graph.reachingAllocations(receiver, true, depth+1)
	reasons = append(reasons, bases.Reasons...)
	for _, site := range bases.Sites {
		found := false
		if literal, ok := index.records[site]; ok {
			if literal.Spread != nil {
				reasons = append(reasons, "spread field projection not certified")
			}
			for _, field := range literal.Fields {
				if field.Name == name {
					sources = append(sources, field.Value)
					found = true
				}
			}
		}
		if literal, ok := index.arrays[site]; ok && array {
			key, _ := strconv.Atoi(name)
			if len(literal.Spread) > 0 {
				reasons = append(reasons, "spread element projection not certified")
			}
			if key < len(literal.Elements) {
				sources = append(sources, literal.Elements[key])
				found = true
			}
		}
		if !found {
			reasons = append(reasons, "projected field absent or allocation not modeled: "+name)
		}
		sources = append(sources, index.stores[site][name]...)
	}
	return sources, reasons, true
}

// shapeCallbackFlow adds closed function-value call edges only for proof queries.
// Graph-region ownership keeps its original sources and callback frontiers.
type shapeCallbackFlow struct {
	sources           map[int][]ir.Expression
	unknownParameters map[int]bool
	facts             map[int]ir.FunctionTargets
	frontiers         map[int][]ir.Expression
}

func (graph *allocationFlowGraph) prepareShapeCallbacks() *shapeCallbackFlow {
	if graph.callbacks != nil {
		return graph.callbacks
	}
	flow := &shapeCallbackFlow{sources: map[int][]ir.Expression{}, unknownParameters: map[int]bool{}, facts: map[int]ir.FunctionTargets{}, frontiers: map[int][]ir.Expression{}}
	graph.callbacks = flow
	for node, sources := range graph.sources {
		flow.sources[node] = append([]ir.Expression{}, sources...)
	}
	calls := []ir.CallClosure{}
	escapes := []ir.Expression{}
	// Unknown operations may retain a function value or invoke it outside this graph.
	escapeChildren := func(node any) {
		walk(node, func(child any) bool {
			if value, ok := child.(ir.Expression); ok {
				escapes = append(escapes, value)
			}
			return true
		})
	}
	collect := func(node any) bool {
		switch value := node.(type) {
		case ir.CallClosure:
			calls = append(calls, value)
		case ir.Call:
			for _, target := range graph.program.CallTargets(value) {
				for i, param := range graph.program.Functions[target].Parameters {
					if i >= len(value.Arguments) {
						flow.unknownParameters[param+1] = true
					}
				}
				for _, argument := range value.Arguments[min(len(value.Arguments), len(graph.program.Functions[target].Parameters)):] {
					escapes = append(escapes, argument)
				}
			}
		case ir.Evaluate:
			switch value.Value.(type) {
			case ir.Call, ir.CallClosure:
			default:
				escapes = append(escapes, value.Value)
			}
		case ir.SetProperty:
			escapes = append(escapes, value.Value)
		case ir.SetIndex:
			escapes = append(escapes, value.Value)
		default:
			if expression, ok := node.(ir.Expression); ok {
				switch expression.(type) {
				case ir.Read, ir.MakeClosure, ir.Call, ir.CallClosure, ir.Conditional, ir.Box, ir.Narrow, ir.Unwrap, ir.CheckedCast:
				default:
					escapeChildren(expression)
				}
			}
		}
		return true
	}
	walk(graph.program.Main, collect)
	for _, function := range graph.program.Functions {
		walk(function.Body, collect)
	}
	merge := func(node int, incoming ir.FunctionTargets) bool {
		current := flow.facts[node]
		seen := map[int]bool{}
		for _, function := range current.Functions {
			seen[function] = true
		}
		changed := incoming.Unknown && !current.Unknown
		current.Unknown = current.Unknown || incoming.Unknown
		for _, function := range incoming.Functions {
			if !seen[function] {
				seen[function] = true
				current.Functions = append(current.Functions, function)
				changed = true
			}
		}
		sort.Ints(current.Functions)
		flow.facts[node] = current
		return changed
	}
	edges := map[[3]int]bool{}
	// Target identities and argument edges grow monotonically. Empty recursive
	// cycles become Unknown after saturation; they never certify an empty target set.
	solve := func() bool {
		changed := false
		for node, sources := range flow.sources {
			for _, source := range sources {
				changed = merge(node, graph.shapeFunctionTargets(source)) || changed
			}
		}
		for i, call := range calls {
			targets := graph.shapeClosureTargets(call)
			for _, target := range targets.Functions {
				for j, param := range graph.program.Functions[target].Parameters {
					if j >= len(call.Arguments) {
						flow.unknownParameters[param+1] = true
						changed = merge(param+1, ir.FunctionTargets{Unknown: true}) || changed
						continue
					}
					key := [3]int{i, target, j}
					if !edges[key] {
						edges[key] = true
						flow.sources[param+1] = append(flow.sources[param+1], call.Arguments[j])
						changed = true
					}
				}
			}
		}
		return changed
	}
	exhausted := true
	for step := 0; step < 1024; step++ {
		if !solve() {
			exhausted = false
			break
		}
	}
	// Empty identities and omitted inputs are real unknown frontiers. Propagate
	// them before testing escapes, including an opaque arm joined to a literal.
	for node := 1; node <= len(graph.program.Locals)+len(graph.program.Functions); node++ {
		if len(flow.facts[node].Functions) == 0 || flow.unknownParameters[node] {
			merge(node, ir.FunctionTargets{Unknown: true})
		}
	}
	escaped := map[int]bool{}
	escapingCalls := map[int]bool{}
	for step := 0; step < 1024; step++ {
		changed := solve()
		for i, call := range calls {
			targets := graph.shapeClosureTargets(call)
			if !escapingCalls[i] && (targets.Unknown || len(targets.Functions) == 0) {
				escapingCalls[i] = true
				escapes = append(escapes, call.Arguments...)
				pending := []int{}
				for _, argument := range call.Arguments {
					pending = append(pending, graph.shapeFunctionTargets(argument).Functions...)
				}
				noted := map[int]bool{}
				for len(pending) > 0 {
					target := pending[len(pending)-1]
					pending = pending[:len(pending)-1]
					if noted[target] {
						continue
					}
					noted[target] = true
					for _, param := range graph.program.Functions[target].Parameters {
						flow.frontiers[param+1] = append(flow.frontiers[param+1], ir.ClosureValue(call))
					}
					pending = append(pending, flow.facts[graph.resultNode(target)].Functions...)
				}

			}
		}
		pending := []int{}
		for _, value := range escapes {
			pending = append(pending, graph.shapeFunctionTargets(value).Functions...)
		}
		for len(pending) > 0 {
			target := pending[len(pending)-1]
			pending = pending[:len(pending)-1]
			if !escaped[target] {
				escaped[target] = true
				for _, param := range graph.program.Functions[target].Parameters {
					if !flow.unknownParameters[param+1] {
						flow.unknownParameters[param+1] = true
						changed = true
					}
					changed = merge(param+1, ir.FunctionTargets{Unknown: true}) || changed
				}
			}
			// Escaped factories can return another callable to an opaque caller.
			for _, returned := range flow.facts[graph.resultNode(target)].Functions {
				if !escaped[returned] {
					pending = append(pending, returned)
				}
			}
			// A factory returning itself must also terminate.
		}
		if !changed {
			break
		}
		if step == 1023 {
			exhausted = true
		}
	}
	if exhausted {
		for _, function := range graph.program.Functions {
			for _, param := range function.Parameters {
				flow.unknownParameters[param+1] = true
			}
		}
		for node, fact := range flow.facts {
			fact.Unknown = true
			flow.facts[node] = fact
		}
	}
	return flow
}

// shapeFunctionTargets resolves identities, never signatures asserted by a view.
// Only closure literals, collected aliases/parameters/returns and joins supply IDs.
func (graph *allocationFlowGraph) shapeFunctionTargets(expression ir.Expression) ir.FunctionTargets {
	flow := graph.callbacks
	union := func(values ...ir.FunctionTargets) ir.FunctionTargets {
		result := ir.FunctionTargets{}
		seen := map[int]bool{}
		for _, value := range values {
			result.Unknown = result.Unknown || value.Unknown
			for _, function := range value.Functions {
				if !seen[function] {
					seen[function] = true
					result.Functions = append(result.Functions, function)
				}
			}
		}
		sort.Ints(result.Functions)
		return result
	}
	switch value := expression.(type) {
	case ir.MakeClosure:
		if value.Function < 0 || value.Function >= len(graph.program.Functions) {
			return ir.FunctionTargets{Unknown: true}
		}
		return ir.FunctionTargets{Functions: []int{value.Function}}
	case ir.Read:
		if value.Local < 0 || value.Local >= len(graph.program.Locals) {
			return ir.FunctionTargets{Unknown: true}
		}
		return flow.facts[value.Local+1]
	case ir.Call:
		result := ir.FunctionTargets{}
		for _, target := range graph.program.CallTargets(value) {
			result = union(result, flow.facts[graph.resultNode(target)])
		}
		return result
	case ir.CallClosure:
		targets := graph.shapeClosureTargets(value)
		result := ir.FunctionTargets{Unknown: targets.Unknown}
		for _, target := range targets.Functions {
			result = union(result, flow.facts[graph.resultNode(target)])
		}
		return result
	case ir.Conditional:
		return union(graph.shapeFunctionTargets(value.WhenTrue), graph.shapeFunctionTargets(value.WhenNot))
	case ir.Box:
		return graph.shapeFunctionTargets(value.Value)
	case ir.Narrow:
		return graph.shapeFunctionTargets(value.Value)
	case ir.Unwrap:
		return graph.shapeFunctionTargets(value.Value)
	case ir.CheckedCast:
		return graph.shapeFunctionTargets(value.Value)
	default:
		return ir.FunctionTargets{Unknown: true}
	}
}

func (graph *allocationFlowGraph) shapeFollow(expression ir.Expression, allocation func(int), demand func(int), unknown func(string)) {
	if call, ok := expression.(ir.CallClosure); ok {
		targets := graph.shapeClosureTargets(call)
		if targets.Unknown || len(targets.Functions) == 0 {
			unknown("callback target identity is unknown")
		}
		for _, target := range targets.Functions {
			demand(graph.resultNode(target))
		}
		return
	}
	graph.follow(expression, allocation, demand, unknown)
}

func (graph *allocationFlowGraph) shapeClosureTargets(call ir.CallClosure) ir.FunctionTargets {
	return graph.program.ClosureTargetsWithFlow(call, graph.shapeFunctionTargets)
}

// shapeDynamicProjection is a query-only source adapter node. Executable IR
// already supplies ArrayIndex; this node also preserves record key provenance.
type shapeDynamicProjection struct{ Receiver, Key ir.Expression }

func (shapeDynamicProjection) Type() ir.Type { return ir.Object }

// projectionKeys uses value producers, never the asserted or declared key type.
// An open producer keeps Unknown even when other producers supply finite keys.
func (graph *allocationFlowGraph) projectionKeys(expression ir.Expression) ([]string, bool) {
	keys := map[string]bool{}
	seen := map[int]bool{}
	pending := []ir.Expression{expression}
	unknown := false
	for len(pending) > 0 {
		value := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		switch value := value.(type) {
		case ir.NumberConstant:
			if value.Value < 0 || value.Value != float64(int(value.Value)) {
				unknown = true
			} else {
				keys[strconv.Itoa(int(value.Value))] = true
			}
		case ir.StringConstant:
			if value.Index < 0 || value.Index >= len(graph.program.Strings) {
				unknown = true
			} else {
				keys[graph.program.Strings[value.Index]] = true
			}
		case ir.Read:
			node := value.Local + 1
			if seen[node] {
				continue
			}
			seen[node] = true
			if graph.callbacks.unknownParameters[node] || len(graph.callbacks.sources[node]) == 0 {
				unknown = true
			}
			pending = append(pending, graph.callbacks.sources[node]...)
		case ir.Conditional:
			pending = append(pending, value.WhenTrue, value.WhenNot)
		default:
			unknown = true
		}
	}
	result := []string{}
	for key := range keys {
		result = append(result, key)
	}
	sort.Strings(result)
	return result, unknown || len(result) == 0
}

// dynamicProjectionSources joins every possible selected slot. An unbounded key
// joins all known slots and remains Unknown because membership is not proven.
type shapeDynamicProjectionResult struct {
	sources []ir.Expression
	reasons []string
}

func (graph *allocationFlowGraph) dynamicProjectionSources(receiver, key ir.Expression, depth int) (sources []ir.Expression, reasons []string, recognized bool) {
	if depth >= 32 {
		return nil, []string{"recursive dynamic projection depth not certified"}, true
	}
	index := graph.projectionIndex()
	// The immutable proof graph reuses projections across cast queries. Include
	// depth so memoization never bypasses the recursion obligation; cap retained
	// entries by weight, without changing any result when caching is skipped.
	receiverRead, receiverOK := receiver.(ir.Read)
	keyRead, keyOK := key.(ir.Read)
	if receiverOK && keyOK {
		cacheKey := [3]int{receiverRead.Local, keyRead.Local, depth}
		if cached, ok := index.dynamicCache[cacheKey]; ok {
			return cached.sources, cached.reasons, true
		}
		defer func() {
			weight := len(sources) + len(reasons) + 1
			if index.dynamicCacheWeight+weight > 65536 {
				return
			}
			if index.dynamicCache == nil {
				index.dynamicCache = map[[3]int]shapeDynamicProjectionResult{}
			}
			index.dynamicCache[cacheKey] = shapeDynamicProjectionResult{sources, reasons}
			index.dynamicCacheWeight += weight
		}()
	}
	names, open := graph.projectionKeys(key)
	bases := graph.reachingAllocations(receiver, true, depth+1)
	sources = []ir.Expression{}
	reasons = append([]string{}, bases.Reasons...)
	if open {
		reasons = append(reasons, "dynamic key membership not proven")
	}
	if len(index.dynamicStores) != 0 {
		reasons = append(reasons, "indexed store effects not certified")
	}
	if index.opaqueCalls {
		reasons = append(reasons, "opaque call may mutate projected fields")
	}
	for _, site := range bases.Sites {
		selected := append([]string{}, names...)
		if open {
			if record, ok := index.records[site]; ok {
				for _, field := range record.Fields {
					selected = append(selected, field.Name)
				}
			}
			if array, ok := index.arrays[site]; ok {
				for i := range array.Elements {
					selected = append(selected, strconv.Itoa(i))
				}
			}
			for name := range index.stores[site] {
				selected = append(selected, name)
			}
		}
		if len(selected) == 0 {
			reasons = append(reasons, "dynamic projection allocation has no modeled slots")
		}
		seenNames := map[string]bool{}
		for _, name := range selected {
			if seenNames[name] {
				continue
			}
			seenNames[name] = true
			found := false
			if record, ok := index.records[site]; ok {
				if record.Spread != nil {
					reasons = append(reasons, "spread field projection not certified")
				}
				if index.recordSlots == nil {
					index.recordSlots = map[int]map[string][]ir.Expression{}
				}
				slots, present := index.recordSlots[site]
				if !present {
					slots = map[string][]ir.Expression{}
					for _, field := range record.Fields {
						slots[field.Name] = append(slots[field.Name], field.Value)
					}
					index.recordSlots[site] = slots
				}
				if values, present := slots[name]; present {
					sources = append(sources, values...)
					found = true
				}
			}
			if array, ok := index.arrays[site]; ok {
				if index.arrayMutation {
					reasons = append(reasons, "array mutation prevents element certificate")
				}
				if len(array.Spread) > 0 {
					reasons = append(reasons, "spread element projection not certified")
				}
				i, err := strconv.Atoi(name)
				if err == nil && i >= 0 && i < len(array.Elements) && strconv.Itoa(i) == name {
					sources = append(sources, array.Elements[i])
					found = true
				}
			}
			if !found {
				reasons = append(reasons, "dynamic projected slot absent or allocation not modeled: "+name)
			}
			sources = append(sources, index.stores[site][name]...)
			sources = append(sources, index.opaqueSources[name]...)
			if index.opaque[name] {
				reasons = append(reasons, "field store has opaque receiver or clears readiness: "+name)
			}
		}
	}
	return sources, reasons, true
}

// Query-only marker for indexed stores whose alias effects are not modeled.
type shapeDynamicMutation struct{ Value ir.Expression }

func (shapeDynamicMutation) Type() ir.Type { return ir.Object }
