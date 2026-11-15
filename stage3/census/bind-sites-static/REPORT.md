**Partial static projection: (a) 15 candidate new count stops; (b) 0 demonstrated known-reach overflows; (c) 19 source-pool count-clean sites. Another 21 have an older rest-bearing candidate and 1 is the unused helper body.**

The asterisk on `a*` and `c*` below is essential: these are conditional source-level classifications, **not a certified census of the head’s completed IR `closureRecords`**. The exact rule, root closure, all site identities, member types and source parameter counts are measured. The source implementation/flow sets are explicitly incomplete/conservative; generated and specialized IR records have not been exhaustively reconstructed. No claim that tsc reaches this pass, newly compiles, or actually stops at 15 sites is made. This is a measurement/report only; compiler and adapters are unchanged.

**Scope correction:** the requested pins produce **56 `.bind` calls and zero `maybeBind` calls**, not 12 and 31. Pristine v6.0.3 has 10 `.bind` calls (including the helper) and 46 `maybeBind` calls. Main’s 32, 95h and 99x adapters rewrite all 46 calls. There are zero such sites in `binder.ts`. All 56 actual sites are retained below, with their pristine origins. No requested site was silently omitted to fit the supplied counts.

## Exact head rule

`M` in this rule is the length of the generated bound invocation’s `Arguments`, not JavaScript `.length`, not the number of required parameters, and not the number of arguments at the bind call. The wrapper sends the fixed bind prefix plus the target signature’s remaining declared slots. Thus optional/defaulted slots count. If the fixed prefix itself exceeds the declared slots, its size controls the invocation count. Every actual non-helper site here binds only the receiver, so its invocation count equals the target signature parameter count tabulated as M.

`internal/lower/function_bind.go:113–128` at `ba619e28774130e24331c3026a938cd6f3cd6a34`:

```go
113: 	arguments := append([]ir.Expression{}, reads[2:]...)
114: 	// The library models bind's remaining parameters as a generic rest tuple.
115: 	// Use the declared target's slots after the fixed prefix, rather than treating
116: 	// that tuple as a single object argument.
117: 	for _, parameter := range parameters[min(len(values)-2, len(parameters)):] {
118: 		proven := l.checker.GetTypeOfSymbol(parameter)
119: 		of, known := l.bindParameterType(parameter)
120: 		if !known {
121: 			return nil, true, l.notYet(node, "a bound function parameter without a value representation")
122: 		}
123: 		local := len(l.result.Locals)
124: 		l.result.Locals = append(l.result.Locals, ir.Local{Name: "argument", Type: of, Function: wrapper})
125: 		l.noteLocal(local, proven, node)
126: 		function.Parameters = append(function.Parameters, local)
127: 		arguments = append(arguments, ir.Read{Local: local, Of: of})
128: 	}
```

`internal/lower/function_bind.go:363–381` at `ba619e28774130e24331c3026a938cd6f3cd6a34`:

```go
363: // declaresBeyond is whether candidate declares more parameters than a call passing passed arguments
364: // fills: a value held as a narrower type, `(shape: Shape, index?: number) => boolean` held as
365: // `(shape: Shape) => boolean`, is called with the type's arguments alone, while JavaScript hands it
366: // every argument its caller passed (#04f5s0s, #jqf1drp). Parameters holds the declared parameters in
367: // order (a destructured or defaulted one arrives as one incoming local), behind a method's leading
368: // receiver, which Receiver marks and no argument fills, so it's left out of the count. The environment
369: // is Environment, never Parameters. A lone this parameter is counted, which can only stop more.
370: func declaresBeyond(candidate ir.Function, passed int) bool {
371: 	return declaredParameters(candidate) > passed
372: }
373:
374: // declaredParameters is how many parameters candidate declares, its leading receiver left out
375: // (declaresBeyond).
376: func declaredParameters(candidate ir.Function) int {
377: 	if candidate.Receiver {
378: 		return len(candidate.Parameters) - 1
379: 	}
380: 	return len(candidate.Parameters)
381: }
```

The count excludes an IR method’s inserted leading receiver; captured environment cells are not parameters. A written lone `this` parameter is conservatively counted. Source N below counts written parameter declarations, including optional/defaulted/destructured/rest declarations; class insertion is kept separate.

A known reach requires the allocation graph to report no unknown frontier and every reached site to be one registered by this query as a `MakeClosure`, `BoundClosure`, undefined, or null. Only the first two contribute functions; null/undefined contribute none. Any other origin, untracked producer/argument/result, graph-owned object/array allocation or unknown projection makes the reach unknown. There is no “use the known subset” fallback.

`internal/lower/function_bind.go:405–445` at `ba619e28774130e24331c3026a938cd6f3cd6a34`:

```go
405: func (l *lowering) boundValueReach() func(ir.Expression) ([]int, bool) {
406: 	assignAllocationSites(l.result)
407: 	graph := l.flowGraph(l.result)
408: 	const nothing, other = -1, -2
409: 	named := []int{}
410: 	site := func(function int) int {
411: 		named = append(named, function)
412: 		return len(named)
413: 	}
414: 	graph.entry = func(value ir.Expression) int {
415: 		switch value := value.(type) {
416: 		case ir.MakeClosure:
417: 			return site(value.Function)
418: 		case ir.BoundClosure:
419: 			return site(value.Function)
420: 		case ir.Undefined, ir.Null:
421: 			return site(nothing)
422: 		case ir.Read, ir.Call, ir.CallClosure, ir.Conditional, ir.Box, ir.Narrow, ir.Unwrap, ir.CheckedCast, ir.Defined, ir.Coalesce, ir.Logical, ir.Property, ir.ArrayIndex, ir.ObjectLiteral, ir.ArrayLiteral:
423: 			return 0
424: 		}
425: 		return site(other)
426: 	}
427: 	return func(value ir.Expression) ([]int, bool) {
428: 		reaches := graph.ReachingAllocations(value)
429: 		if reaches.Unknown {
430: 			return nil, false
431: 		}
432: 		functions := []int{}
433: 		for _, reached := range reaches.Sites {
434: 			if reached <= 0 || reached > len(named) || named[reached-1] == other {
435: 				// An allocation the graph names itself, an object or an array, or a value made
436: 				// another way.
437: 				return nil, false
438: 			}
439: 			if named[reached-1] != nothing {
440: 				functions = append(functions, named[reached-1])
441: 			}
442: 		}
443: 		return functions, true
444: 	}
445: }
```

The unknown fallback is **all recorded lowered function implementations whose recorded checker type is assignable to the bound checker type**, irrespective of member name, host identity, assignment into that member, execution liveness or binding at this site. It is neither “every function with this member name” nor “all functions of arity M”. It includes recorded source/module functions, closures, bound wrappers, specializations and generated callable adapters when their records are present. Class method-table functions under `BoundName` are added separately and are never narrowed by the reach result.

`internal/lower/arguments_length.go:139–177` at `ba619e28774130e24331c3026a938cd6f3cd6a34`:

```go
139: // resolveCountTypes uses the same assignability relation and closed-world function
140: // records as the cycle finder. A parameter or mutable function value is bounded
141: // by every implementation that can be seen through its function type.
142: func (l *lowering) resolveCountTypes(modules []*ast.SourceFile) {
143: 	types := map[int]*checker.Type{}
144: 	for _, module := range modules {
145: 		var visit ast.Visitor
146: 		visit = func(node *ast.Node) bool {
147: 			if node.Kind == ast.KindCallExpression {
148: 				proven := l.concrete(l.checker.GetTypeAtLocation(node.AsCallExpression().Expression))
149: 				types[int(proven.Id())] = proven
150: 				for _, argument := range node.AsCallExpression().Arguments.Nodes {
151: 					callback := l.concrete(l.checker.GetTypeAtLocation(argument))
152: 					if len(l.checker.GetSignaturesOfType(callback, checker.SignatureKindCall)) > 0 {
153: 						types[int(callback.Id())] = callback
154: 					}
155: 				}
156: 			}
157: 			return node.ForEachChild(visit)
158: 		}
159: 		module.AsNode().ForEachChild(visit)
160: 	}
161: 	for _, proven := range l.localTypes {
162: 		if proven != nil {
163: 			proven = l.concrete(proven)
164: 			types[int(proven.Id())] = proven
165: 		}
166: 	}
167: 	l.resolveStructuralMethodThunks(modules)
168: 	l.result.FunctionTypeTargets = map[int][]int{}
169: 	for id, proven := range types {
170: 		l.cycleTypeID(proven) // Retain transient count-target types for differential tests.
171: 		l.result.FunctionTypeTargets[id] = []int{}
172: 		for _, record := range l.closureRecords {
173: 			if l.checker.IsTypeAssignableTo(record.proven, proven) {
174: 				l.result.FunctionTypeTargets[id] = append(l.result.FunctionTypeTargets[id], record.function)
175: 			}
176: 		}
177: 	}
```

`internal/ir/call_targets.go:174–203` at `ba619e28774130e24331c3026a938cd6f3cd6a34`:

```go
174: // MethodsNamed is every function a method table holds under the name, in function order: the
175: // methods of the objects class constructors make (ObjectLiteral's Methods), a class's static side's
176: // among them, which is all a call through an interface finds a method in.
177: func (p *Program) MethodsNamed(name string) []int {
178: 	if tables := p.methodTables; tables != nil && tables.functions == len(p.Functions) {
179: 		return tables.named[name]
180: 	}
181: 	tables := &methodTables{functions: len(p.Functions), named: map[string][]int{}}
182: 	held := map[Method]bool{}
183: 	visit := func(node any) {
184: 		literal, isLiteral := node.(ObjectLiteral)
185: 		if !isLiteral {
186: 			return
187: 		}
188: 		for _, method := range literal.Methods {
189: 			if !held[method] {
190: 				held[method] = true
191: 				tables.named[method.Name] = append(tables.named[method.Name], method.Function)
192: 			}
193: 		}
194: 	}
195: 	spreadStorageWalk(reflect.ValueOf(p.Main), visit)
196: 	for _, function := range p.Functions {
197: 		spreadStorageWalk(reflect.ValueOf(function.Body), visit)
198: 	}
199: 	for _, methods := range tables.named {
200: 		slices.Sort(methods)
201: 	}
202: 	p.methodTables = tables
203: 	return tables.named[name]
```

Existing argument-count/rest refusal takes precedence across the entire type/name candidate set. Only if it did not find a refusal does the extra-parameter check run. It asks reach only when some type-wide candidate has N > M; known reach replaces the typed count set, unknown reach retains the whole typed set. Named methods remain in either case. Any counted N > M stops. A known reach containing such a function therefore also stops; categories (a) and the question’s literal (b) overlap. This report uses stop-first counting and reserves (b) for a demonstrated overflow escaping the new guard. None is established.

`internal/lower/function_bind.go:459–504` at `ba619e28774130e24331c3026a938cd6f3cd6a34`:

```go
459: 			if call, ok := node.(ir.CallClosure); ok && call.Bound {
460: 				typed := l.result.FunctionTypeTargets[call.FunctionType]
461: 				candidates := append([]int{}, typed...)
462: 				methods := []int{}
463: 				if call.BoundName != "" {
464: 					methods = l.result.MethodsNamed(call.BoundName)
465: 					candidates = append(candidates, methods...)
466: 				}
467: 				for _, target := range candidates {
468: 					if candidate := l.result.Functions[target]; candidate.ReadsArguments || candidate.RestElement != 0 {
469: 						found = l.notYet(record.node, "bind of a function that observes its argument count or collects rest arguments")
470: 					}
471: 				}
472: 				if found != nil {
473: 					// The count's stop gives way to the arguments and rest one, whose words stay as pinned.
474: 					return true
475: 				}
476: 				// The count asks only what can be the bound value: every function of its type counts
477: 				// functions the program never binds here, such as one taking an optional parameter held as
478: 				// () => number elsewhere, which stopped function_bind.a's honest absentProperty.f?.bind.
479: 				// A method bind keeps every method of the name.
480: 				counted := append([]int{}, methods...)
481: 				if slices.ContainsFunc(typed, func(target int) bool { return declaresBeyond(l.result.Functions[target], len(call.Arguments)) }) {
482: 					if reach == nil {
483: 						reach = l.boundValueReach()
484: 					}
485: 					reached, known := reach(call.Closure)
486: 					if observe := boundReachObserved.Load(); observe != nil {
487: 						names := []string{}
488: 						for _, function := range reached {
489: 							names = append(names, l.result.Functions[function].Name)
490: 						}
491: 						(*observe)(l.program.Where(record.node), names, known)
492: 					}
493: 					if known {
494: 						counted = append(counted, reached...)
495: 					} else {
496: 						counted = append(counted, typed...)
497: 					}
498: 				}
499: 				for _, target := range counted {
500: 					if candidate := l.result.Functions[target]; declaresBeyond(candidate, len(call.Arguments)) {
501: 						found = l.notYet(record.node, "bind of a function that declares "+parameterCount(declaredParameters(candidate))+" where the type it's bound as names "+strconv.Itoa(len(call.Arguments)))
502: 						break
503: 					}
504: 				}
```

Projection certificates are globally conservative. Any `CallClosure`, `ObjectCall`, generator/iterator step or yield anywhere in Main or any function body sets `opaqueCalls`; every property/element projection then records an unknown reason, even if its literal field can be identified. Opaque field stores, mutation, spread, absent/unmodelled fields, recursive projection and dynamic element keys add other unknown reasons.

`internal/lower/shape_flow.go:267–279` at `ba619e28774130e24331c3026a938cd6f3cd6a34`:

```go
267: 		case ir.SetProperty:
268: 			stores = append(stores, value)
269: 		case ir.SetIndex, ir.ArrayPush, ir.ArraySplice, ir.ArrayPop:
270: 			index.arrayMutation = true
271: 		case ir.CallClosure, ir.ObjectCall, ir.GeneratorNext, ir.IteratorNext, ir.Yield:
272: 			index.opaqueCalls = true
273: 		}
274: 		return true
275: 	}
276: 	walk(graph.program.Main, collect)
277: 	for _, function := range graph.program.Functions {
278: 		walk(function.Body, collect)
279: 	}
```

`internal/lower/shape_flow.go:319–335` at `ba619e28774130e24331c3026a938cd6f3cd6a34`:

```go
319: 	if depth >= 32 {
320: 		return nil, []string{"recursive projection depth not certified"}, true
321: 	}
322: 	index := graph.projectionIndex()
323: 	sources := []ir.Expression{}
324: 	reasons := []string{}
325: 	if index.opaqueCalls {
326: 		reasons = append(reasons, "opaque call may mutate projected fields")
327: 	}
328: 	if index.opaque[name] {
329: 		reasons = append(reasons, "field store has opaque receiver or clears readiness: "+name)
330: 	}
331: 	if array && index.arrayMutation {
332: 		reasons = append(reasons, "array mutation prevents element certificate")
333: 	}
334: 	bases := graph.reachingAllocations(receiver, true, depth+1)
335: 	reasons = append(reasons, bases.Reasons...)
```

`internal/lower/shape_flow.go:359–364` at `ba619e28774130e24331c3026a938cd6f3cd6a34`:

```go
359: 		if !found {
360: 			reasons = append(reasons, "projected field absent or allocation not modeled: "+name)
361: 		}
362: 		sources = append(sources, index.stores[site][name]...)
363: 	}
364: 	return sources, reasons, true
```

**All 55 non-helper sites use the direct same-identifier method-bind path. If their count reach is queried, it is unknown in this head.** The method path builds a `BoundMethod` whose factory captures the runtime lookup, rather than an ordinary IR `Call` with the placeholder target argument. The shared graph’s call/parameter collection has no `BoundMethod` argument edge. Its factory’s first `bound_value` parameter therefore has no tracked producer. `reachingAllocations` reports that as unknown. Even the three resolutionCache members whose source object literals are statically identifiable do not get a known head certificate. Opaque projection calls provide an additional barrier when source properties are followed; do not equate the supplementary source-flow result with head reach.

`internal/lower/function_bind.go:62–74` at `ba619e28774130e24331c3026a938cd6f3cd6a34`:

```go
62: 	var value ir.Expression
63: 	var err error
64: 	if method {
65: 		value = ir.Undefined{Of: ir.Closure}
66: 	} else {
67: 		value, err = l.expression(target)
68: 	}
69: 	if err != nil {
70: 		return nil, true, err
71: 	}
72: 	values := []ir.Expression{value}
73: 	types := []*checker.Type{proven}
74: 	nodes := []*ast.Node{target}
```

`internal/lower/function_bind.go:137–156` at `ba619e28774130e24331c3026a938cd6f3cd6a34`:

```go
137: 	invoke := ir.CallClosure{Bound: true, Closure: reads[0], Arguments: arguments, FunctionType: int(proven.Id()), Returns: function.Returns}
138: 	if method {
139: 		invoke.BoundName = target.Name().Text()
140: 	}
141: 	if function.Returns == 0 {
142: 		function.Body = []ir.Statement{ir.Evaluate{Value: invoke}, ir.Return{}}
143: 	} else {
144: 		function.Body = []ir.Statement{ir.Return{Value: invoke}}
145: 	}
146: 	l.result.Functions[wrapper] = function
147: 	l.closureRecords = append(l.closureRecords, closureRecord{proven: boundType, function: wrapper, node: node})
148: 	made, err := l.programAllocation(node, l.graphAllocation(ir.BoundClosure{Function: wrapper, Fixed: len(values) - 2}, node))
149: 	if err != nil {
150: 		return nil, true, err
151: 	}
152: 	l.result.Functions[maker].Parameters = environment
153: 	l.result.Functions[maker].Body = []ir.Statement{ir.Return{Value: made}}
154: 	if method {
155: 		l.wholeObjectCopies = append(l.wholeObjectCopies, wholeObjectCopy{node: target, source: values[1], copied: l.checker.GetTypeAtLocation(target.AsPropertyAccessExpression().Expression), reader: "bind lookup"})
156: 		return ir.BoundMethod{Object: values[1], Name: target.Name().Text(), Optional: access.QuestionDotToken != nil, Fixed: values[2:], Factory: maker}, true, nil
```

`internal/lower/view_flow.go:123–143` at `ba619e28774130e24331c3026a938cd6f3cd6a34`:

```go
123: 	pass := func(target int, arguments []ir.Expression, spread []bool) {
124: 		function := &program.Functions[target]
125: 		for i, param := range function.Parameters {
126: 			switch {
127: 			case i < len(arguments):
128: 				graph.sources[param+1] = append(graph.sources[param+1], arguments[i])
129: 			case function.OptionalParameters[param] && !slices.Contains(spread, true):
130: 				graph.sources[param+1] = append(graph.sources[param+1], ir.Undefined{Of: program.Locals[param].Type})
131: 			default:
132: 				graph.unknownParameters[param+1] = true
133: 			}
134: 		}
135: 	}
136: 	collect := func(body []ir.Statement, function int) {
137: 		walk(body, func(node any) bool {
138: 			graph.noteWrites(reflect.ValueOf(node))
139: 			switch value := node.(type) {
140: 			case ir.BoundClosure:
141: 				for _, parameter := range program.Functions[value.Function].Parameters {
142: 					graph.unknownParameters[parameter+1] = true
143: 				}
```

`internal/lower/view_flow.go:153–185` at `ba619e28774130e24331c3026a938cd6f3cd6a34`:

```go
153: 			case ir.Declare:
154: 				// Road 7: a cell made before its declaration runs, so hoisted functions can capture it
155: 				// (nested_functions.go's preallocateNestedBindings, locals.go's definite-assignment let),
156: 				// holds no value until an Assign makes it ready, and it's no source. A read before that
157: 				// stops (readiness.go), and a local only ever declared this way still has no producer.
158: 				if value.Uninitialized && value.Value == nil {
159: 					break
160: 				}
161: 				graph.sources[value.Local+1] = append(graph.sources[value.Local+1], value.Value)
162: 			case ir.Assign:
163: 				graph.sources[value.Local+1] = append(graph.sources[value.Local+1], value.Value)
164: 			case ir.Return:
165: 				if function >= 0 {
166: 					graph.sources[graph.resultNode(function)] = append(graph.sources[graph.resultNode(function)], value.Value)
167: 				}
168: 			case ir.Call:
169: 				for _, target := range program.CallTargets(value) {
170: 					if target < 0 || target >= len(program.Functions) {
171: 						continue
172: 					}
173: 					pass(target, value.Arguments, value.Spread)
174: 				}
175: 			case ir.CallClosure:
176: 				targets := program.ClosureTargets(value)
177: 				if targets.Unknown {
178: 					// Opaque function values are a documented leak-only frontier.
179: 					break
180: 				}
181: 				for _, target := range targets.Functions {
182: 					pass(target, value.Arguments, value.Spread)
183: 				}
184:
185: 			}
```

`internal/lower/shape_flow.go:123–139` at `ba619e28774130e24331c3026a938cd6f3cd6a34`:

```go
123: 	for len(queue) != 0 {
124: 		node := queue[len(queue)-1]
125: 		queue = queue[:len(queue)-1]
126: 		if why, unknownSource := graph.unknownAt(node); unknownSource {
127: 			unknown(why)
128: 		}
129: 		sources := graph.sources[node]
130: 		if len(sources) == 0 {
131: 			unknown("local or result has no tracked producer")
132: 		}
133: 		for _, source := range sources {
134: 			follow(source, depth)
135: 		}
136: 	}
137: 	if len(sites) == 0 {
138: 		unknown("no reaching allocation established")
139: 	}
```

The pass returns the first recorded error (`function_bind.go:508–512`), and runs only after ordinary lowering, cycle discovery and count-target construction (`lower.go:215–220`). Static site rows are not an observed error sequence.

## maybeBind and the adapted program

`internal/lower/function_bind.go:28–35` at `ba619e28774130e24331c3026a938cd6f3cd6a34`:

```go
28: 	if (node.Flags&ast.NodeFlagsOptionalChain != 0 || access.QuestionDotToken != nil) && target.Kind != ast.KindPropertyAccessExpression {
29: 		return nil, true, l.notYet(node, "bind through an optional chain")
30: 	}
31: 	proven := l.concrete(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(target)))
32: 	signatures := l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)
33: 	if l.isLibraryType(proven, "Function") || len(signatures) != 1 {
34: 		return nil, true, l.notYet(node, "bind without one declared function signature")
35: 	}
```

The retained helper is `src/compiler/core.ts:1519–1521`, pristine `:1522–1524`:

```typescript
export function maybeBind<T, A extends any[], R>(obj: T, fn: ((this: T, ...args: A) => R) | undefined): ((...args: A) => R) | undefined {
    return fn?.bind(obj);
}
```

Adamic monomorphizes generic direct calls: `generic.go:24–49` infers each call’s concrete parameter/result mapping, `:80–114` keys/reuses instances, `:141–168` lowers with the mapper, and `instantiate.go:50–56` substitutes types through `concrete`. It does not judge the helper once as a monomorphic source function with its written generic constraint. At this exact adapted pin **there are no core-helper callers and hence no caller instantiations to enumerate**. If a caller were retained, `fn?.bind(obj)` has an identifier target and is rejected by the quoted optional-chain rule before parameter counting, regardless of the caller’s instantiation. The rewritten `host.member?.bind(host)` sites have a property target and pass that particular syntactic check; they are judged independently with the member’s concrete type. The helper row is `H`, not a fabricated per-caller count verdict.

The API program uses all 79 recursively enumerated `.ts` roots under `src/compiler`, the pinned `@types/node@25.3.3` ambient index and Adamic’s strict ES2024/bundler type options. It reports **zero TypeScript diagnostics**. All 79 loaded non-declaration sources are within `src/compiler`; **no services, server or harness source is loaded**. System’s source implementation in `sys.ts` is loaded and inventoried. Ambient Node declarations are not source implementations and have no N/body to invent.

## One row per actual adapted site

`U` = the head reach would be unknown if queried; `—` = earlier rest refusal means the reach query is not made, or no typed overflow requires one. `U/—` states both the structural result and that query gating. `a*` = source-pool count-stop candidate; `c*` = source-pool count-clean; `P` = an existing rest-bearing source candidate takes precedence if recorded; `H` = unused helper body/earlier optional-chain blocker. **P and H are not called clean.** The API source-pool relation is an upper-bound audit, not the completed IR table. Every `a*` has an explicit non-generic-declaration witness (`Debug.fail` N=2, or the overloaded `assertOptionalToken` implementation N=4 for setTimeout); absence of additional generated earlier stops is not certified.

Implementation IDs in column I point to the conservative same-field source audit below. “Source flows” lists values found by the supplementary allocation pass; `?` retains an unknown ingress and **does not mean only those values flow**. Each implementation/wrapper ID gives N and a source citation in the catalogue. The three source-known cache rows still have head reach U/—.

| # | Adapted site and quoted call (pristine origin) | Member type / M | I | Enumerated source flows | Head reach | Verdict |
|---|---|---|---|---|---|---|
| B01 | `src/compiler/core.ts:1520:12` `fn?.bind(obj)`; pristine `src/compiler/core.ts:1523 (bind)` | T01, M=1 | I01 | none enumerated; ? | not instantiated | H |
| B02 | `src/compiler/checker.ts:54284:26` `host.getSymlinkCache?.bind(host)`; pristine `src/compiler/checker.ts:54264 (maybeBind)` | T02, M=0 | I02 | F020; ? | U (queried) | a* |
| B03 | `src/compiler/checker.ts:54295:40` `host.getGlobalTypingsCacheLocation?.bind(host)`; pristine `src/compiler/checker.ts:54275 (maybeBind)` | T03, M=0 | I03 | W010; ? | U (queried) | a* |
| B04 | `src/compiler/program.ts:495:21` `system.createHash?.bind(system)`; pristine `src/compiler/program.ts:494 (maybeBind)` | T04, M=1 | I04 | F061; ? | U/— | P |
| B05 | `src/compiler/program.ts:1605:42` `host.resolveModuleNameLiterals.bind(host)`; pristine `src/compiler/program.ts:1604 (bind)` | T05, M=6 | I05 | W063, F069, W100, W102; ? | U/— | c* |
| B06 | `src/compiler/program.ts:1651:58` `host.resolveTypeReferenceDirectiveReferences.bind(host)`; pristine `src/compiler/program.ts:1650 (bind)` | T06, M=6 | I06 | W064, F071, W103, W105; ? | U/— | c* |
| B07 | `src/compiler/program.ts:1687:32` `host.resolveLibrary.bind(host)`; pristine `src/compiler/program.ts:1686 (bind)` | T07, M=4 | I07 | W065, F072, W106, W107; ? | U/— | c* |
| B08 | `src/compiler/program.ts:1735:22` `host.readFile.bind(host)`; pristine `src/compiler/program.ts:1734 (bind)` | T08, M=1 | I08 | F021, F037, F079; ? | U/— | P |
| B09 | `src/compiler/program.ts:1951:19` `host.realpath?.bind(host)`; pristine `src/compiler/program.ts:1950 (bind)` | T09, M=1 | I09 | F019, F029, F060, W083; ? | U/— | P |
| B10 | `src/compiler/program.ts:1957:40` `host.getGlobalTypingsCacheLocation?.bind(host)`; pristine `src/compiler/program.ts:1956 (maybeBind)` | T03, M=0 | I03 | none enumerated; ? | U (queried) | a* |
| B11 | `src/compiler/program.ts:2627:23` `host.realpath?.bind(host)`; pristine `src/compiler/program.ts:2626 (maybeBind)` | T09, M=1 | I09 | F019, F029, F060, W083; ? | U/— | P |
| B12 | `src/compiler/program.ts:2633:25` `host.createHash?.bind(host)`; pristine `src/compiler/program.ts:2632 (maybeBind)` | T04, M=1 | I04 | W030, W086, W098; ? | U/— | P |
| B13 | `src/compiler/program.ts:2635:20` `host.trace?.bind(host)`; pristine `src/compiler/program.ts:2634 (maybeBind)` | T10, M=1 | I10 | F022, W080; ? | U/— | P |
| B14 | `src/compiler/program.ts:5115:26` `directoryStructureHost.directoryExists?.bind(directoryStructureHost)`; pristine `src/compiler/program.ts:5114 (maybeBind)` | T11, M=1 | I11 | F023, F024, F039, W081, F089, F112; ? | U (queried) | a* |
| B15 | `src/compiler/program.ts:5116:25` `directoryStructureHost.getDirectories?.bind(directoryStructureHost)`; pristine `src/compiler/program.ts:5115 (maybeBind)` | T12, M=1 | I12 | F026, F028, W082, F090, F113; ? | U (queried) | a* |
| B16 | `src/compiler/program.ts:5117:19` `directoryStructureHost.realpath?.bind(directoryStructureHost)`; pristine `src/compiler/program.ts:5116 (maybeBind)` | T09, M=1 | I09 | F019, F029, F060, W083, W092, F115; ? | U/— | P |
| B17 | `src/compiler/builder.ts:1750:30` `host.writeFile?.bind(host)`; pristine `src/compiler/builder.ts:1727 (maybeBind)` | T13, M=6 | I13 | F017, F038; ? | U/— | P |
| B18 | `src/compiler/builder.ts:1838:34` `host.writeFile?.bind(host)`; pristine `src/compiler/builder.ts:1815 (maybeBind)` | T13, M=6 | I13 | F017, F038; ? | U/— | P |
| B19 | `src/compiler/builder.ts:1937:78` `host.writeFile?.bind(host)`; pristine `src/compiler/builder.ts:1914 (maybeBind)` | T13, M=6 | I13 | F017, F038; ? | U/— | P |
| B20 | `src/compiler/watch.ts:676:20` `system.watchFile?.bind(system)`; pristine `src/compiler/watch.ts:677 (maybeBind)` | T14, M=4 | I14 | F046; ? | U/— | c* |
| B21 | `src/compiler/watch.ts:677:25` `system.watchDirectory?.bind(system)`; pristine `src/compiler/watch.ts:678 (maybeBind)` | T15, M=4 | I15 | F047; ? | U/— | c* |
| B22 | `src/compiler/watch.ts:678:21` `system.setTimeout?.bind(system)`; pristine `src/compiler/watch.ts:679 (maybeBind)` | T16, M=3 | I16 | F053; ? | U (queried) | a* |
| B23 | `src/compiler/watch.ts:679:23` `system.clearTimeout?.bind(system)`; pristine `src/compiler/watch.ts:680 (maybeBind)` | T17, M=1 | I17 | F054; ? | U/— | P |
| B24 | `src/compiler/watch.ts:760:32` `host.getDefaultLibLocation?.bind(host)`; pristine `src/compiler/watch.ts:761 (maybeBind)` | T18, M=0 | I18 | F005; ? | U (queried) | a* |
| B25 | `src/compiler/watch.ts:773:16` `host.trace?.bind(host)`; pristine `src/compiler/watch.ts:774 (maybeBind)` | T10, M=1 | I10 | F094; ? | U/— | P |
| B26 | `src/compiler/watch.ts:774:26` `directoryStructureHost.directoryExists?.bind(directoryStructureHost)`; pristine `src/compiler/watch.ts:775 (maybeBind)` | T11, M=1 | I11 | F039, F089, F112; ? | U (queried) | a* |
| B27 | `src/compiler/watch.ts:775:25` `directoryStructureHost.getDirectories?.bind(directoryStructureHost)`; pristine `src/compiler/watch.ts:776 (maybeBind)` | T12, M=1 | I12 | F090, F113; ? | U (queried) | a* |
| B28 | `src/compiler/watch.ts:776:19` `host.realpath?.bind(host)`; pristine `src/compiler/watch.ts:777 (maybeBind)` | T09, M=1 | I09 | W092; ? | U/— | P |
| B29 | `src/compiler/watch.ts:777:33` `host.getEnvironmentVariable?.bind(host)`; pristine `src/compiler/watch.ts:778 (maybeBind)` | T19, M=1 | I19 | W093; ? | U/— | P |
| B30 | `src/compiler/watch.ts:778:21` `host.createHash?.bind(host)`; pristine `src/compiler/watch.ts:779 (maybeBind)` | T04, M=1 | I04 | W096; ? | U/— | P |
| B31 | `src/compiler/watch.ts:779:24` `host.readDirectory?.bind(host)`; pristine `src/compiler/watch.ts:780 (maybeBind)` | T20, M=5 | I20 | F091; ? | U/— | c* |
| B32 | `src/compiler/watch.ts:855:19` `system.realpath?.bind(system)`; pristine `src/compiler/watch.ts:856 (maybeBind)` | T09, M=1 | I09 | F060; ? | U/— | P |
| B33 | `src/compiler/watch.ts:856:33` `system.getEnvironmentVariable?.bind(system)`; pristine `src/compiler/watch.ts:857 (maybeBind)` | T21, M=1 | I19 | F052; ? | U/— | P |
| B34 | `src/compiler/watch.ts:860:21` `system.createHash?.bind(system)`; pristine `src/compiler/watch.ts:861 (maybeBind)` | T04, M=1 | I04 | F061; ? | U/— | P |
| B35 | `src/compiler/watch.ts:863:14` `system.now?.bind(system)`; pristine `src/compiler/watch.ts:864 (maybeBind)` | T22, M=0 | I21 | none enumerated; ? | U (queried) | a* |
| B36 | `src/compiler/watchPublic.ts:124:23` `system.createHash?.bind(system)`; pristine `src/compiler/watchPublic.ts:125 (maybeBind)` | T04, M=1 | I04 | F061; ? | U/— | P |
| B37 | `src/compiler/watchPublic.ts:505:56` `host.useSourceOfProjectReferenceRedirect?.bind(host)`; pristine `src/compiler/watchPublic.ts:506 (maybeBind)` | T23, M=0 | I22 | none enumerated; ? | U/— | P |
| B38 | `src/compiler/watchPublic.ts:528:46` `host.resolveModuleNameLiterals?.bind(host)`; pristine `src/compiler/watchPublic.ts:529 (maybeBind)` | T05, M=6 | I05 | none enumerated; ? | U/— | c* |
| B39 | `src/compiler/watchPublic.ts:529:39` `host.resolveModuleNames?.bind(host)`; pristine `src/compiler/watchPublic.ts:530 (maybeBind)` | T24, M=6 | I23 | none enumerated; ? | U/— | c* |
| B40 | `src/compiler/watchPublic.ts:531:50` `resolutionCache.resolveModuleNameLiterals.bind(resolutionCache)`; pristine `src/compiler/watchPublic.ts:532 (bind)` | T05, M=6 | I05 | F040 | U/— | c* |
| B41 | `src/compiler/watchPublic.ts:533:60` `host.resolveTypeReferenceDirectiveReferences?.bind(host)`; pristine `src/compiler/watchPublic.ts:534 (maybeBind)` | T06, M=6 | I06 | none enumerated; ? | U/— | c* |
| B42 | `src/compiler/watchPublic.ts:534:51` `host.resolveTypeReferenceDirectives?.bind(host)`; pristine `src/compiler/watchPublic.ts:535 (maybeBind)` | T25, M=5 | I24 | none enumerated; ? | U/— | c* |
| B43 | `src/compiler/watchPublic.ts:536:64` `resolutionCache.resolveTypeReferenceDirectiveReferences.bind(resolutionCache)`; pristine `src/compiler/watchPublic.ts:537 (bind)` | T06, M=6 | I06 | F045 | U/— | c* |
| B44 | `src/compiler/watchPublic.ts:539:9` `resolutionCache.resolveLibrary.bind(resolutionCache)`; pristine `src/compiler/watchPublic.ts:540 (bind)` | T07, M=4 | I07 | F041 | U/— | c* |
| B45 | `src/compiler/watchPublic.ts:540:9` `host.resolveLibrary.bind(host)`; pristine `src/compiler/watchPublic.ts:541 (bind)` | T07, M=4 | I07 | none enumerated; ? | U/— | c* |
| B46 | `src/compiler/watchPublic.ts:542:9` `host.getModuleResolutionCache?.bind(host)`; pristine `src/compiler/watchPublic.ts:543 (maybeBind)` | T26, M=0 | I25 | none enumerated; ? | U (queried) | a* |
| B47 | `src/compiler/watchPublic.ts:548:9` `host.hasInvalidatedResolutions?.bind(host)`; pristine `src/compiler/watchPublic.ts:549 (maybeBind)` | T27, M=1 | I26 | none enumerated; ? | U (queried) | a* |
| B48 | `src/compiler/watchPublic.ts:551:9` `host.hasInvalidatedLibResolutions?.bind(host)`; pristine `src/compiler/watchPublic.ts:552 (maybeBind)` | T28, M=1 | I27 | none enumerated; ? | U (queried) | a* |
| B49 | `src/compiler/tsbuildPublic.ts:301:16` `system.now?.bind(system)`; pristine `src/compiler/tsbuildPublic.ts:302 (maybeBind)` | T22, M=0 | I21 | none enumerated; ? | U (queried) | a* |
| B50 | `src/compiler/tsbuildPublic.ts:437:46` `host.resolveModuleNameLiterals?.bind(host)`; pristine `src/compiler/tsbuildPublic.ts:438 (maybeBind)` | T05, M=6 | I05 | none enumerated; ? | U/— | c* |
| B51 | `src/compiler/tsbuildPublic.ts:438:60` `host.resolveTypeReferenceDirectiveReferences?.bind(host)`; pristine `src/compiler/tsbuildPublic.ts:439 (maybeBind)` | T06, M=6 | I06 | none enumerated; ? | U/— | c* |
| B52 | `src/compiler/tsbuildPublic.ts:439:35` `host.resolveLibrary?.bind(host)`; pristine `src/compiler/tsbuildPublic.ts:440 (maybeBind)` | T07, M=4 | I07 | none enumerated; ? | U/— | c* |
| B53 | `src/compiler/tsbuildPublic.ts:440:39` `host.resolveModuleNames?.bind(host)`; pristine `src/compiler/tsbuildPublic.ts:441 (maybeBind)` | T24, M=6 | I23 | none enumerated; ? | U/— | c* |
| B54 | `src/compiler/tsbuildPublic.ts:441:51` `host.resolveTypeReferenceDirectives?.bind(host)`; pristine `src/compiler/tsbuildPublic.ts:442 (maybeBind)` | T25, M=5 | I24 | none enumerated; ? | U/— | c* |
| B55 | `src/compiler/tsbuildPublic.ts:442:45` `host.getModuleResolutionCache?.bind(host)`; pristine `src/compiler/tsbuildPublic.ts:443 (maybeBind)` | T26, M=0 | I25 | none enumerated; ? | U (queried) | a* |
| B56 | `src/compiler/tsbuildPublic.ts:499:16` `host.trace?.bind(host)`; pristine `src/compiler/tsbuildPublic.ts:500 (maybeBind)` | T10, M=1 | I10 | F094; ? | U/— | P |

## Member types

Types below are `getNonNullableType(getTypeAtLocation(member))` and call signatures from the pinned checker. The declaration citations are the original member contracts, not implementations. Optional membership is removed only for reading the callable signature. Full declaration text and resolved bind signatures are retained in `evidence/inventory.json.gz`.

- **T01**, M=1: `(this: T, ...args: A) => R`. Contract: `src/compiler/core.ts:1519:58`.
- **T02**, M=0: `() => SymlinkCache`. Contract: `src/compiler/types.ts:10010:5`.
- **T03**, M=0: `() => string \| undefined`. Contract: `src/compiler/types.ts:10013:5`.
- **T04**, M=1: `(data: string) => string`. Contract: `src/compiler/sys.ts:1432:5`.
- **T05**, M=6: `(moduleLiterals: readonly StringLiteralLike[], containingFile: string, redirectedReference: ResolvedProjectReference \| undefined, options: CompilerOptions, containingSourceFile: SourceFile, reusedNames: readonly StringLiteralLike[] \| undefined) => readonly ResolvedModuleWithFailedLookupLocations[]`. Contract: `src/compiler/types.ts:8179:5`.
- **T06**, M=6: `<T extends FileReference \| string>(typeDirectiveReferences: readonly T[], containingFile: string, redirectedReference: ResolvedProjectReference \| undefined, options: CompilerOptions, containingSourceFile: SourceFile \| undefined, reusedNames: readonly T[] \| undefined) => readonly ResolvedTypeReferenceDirectiveWithFailedLookupLocations[]`. Contract: `src/compiler/types.ts:8187:5`.
- **T07**, M=4: `(libraryName: string, resolveFrom: string, options: CompilerOptions, libFileName: string) => ResolvedModuleWithFailedLookupLocations`. Contract: `src/compiler/types.ts:8196:5`.
- **T08**, M=1: `(fileName: string) => string \| undefined`. Contract: `src/compiler/types.ts:8004:5`.
- **T09**, M=1: `(path: string) => string`. Contract: `src/compiler/types.ts:8011:5`.
- **T10**, M=1: `(s: string) => void`. Contract: `src/compiler/types.ts:8005:5`.
- **T11**, M=1: `(path: string) => boolean`. Contract: `src/compiler/watchUtilities.ts:73:5`.
- **T12**, M=1: `(path: string) => string[]`. Contract: `src/compiler/watchUtilities.ts:74:5`.
- **T13**, M=6: `WriteFileCallback`. Contract: `src/compiler/builderPublic.ts:31:5`.
- **T14**, M=4: `(path: string, callback: FileWatcherCallback, pollingInterval?: number \| undefined, options?: WatchOptions \| undefined) => FileWatcher`. Contract: `src/compiler/sys.ts:1415:5`.
- **T15**, M=4: `(path: string, callback: DirectoryWatcherCallback, recursive?: boolean \| undefined, options?: WatchOptions \| undefined) => FileWatcher`. Contract: `src/compiler/sys.ts:1416:5`.
- **T16**, M=3: `(callback: (label?: unknown) => void, ms: number, label?: unknown) => unknown`. Contract: `src/compiler/sys.ts:1445:5`.
- **T17**, M=1: `(timeoutId: unknown) => void`. Contract: `src/compiler/sys.ts:1446:5`.
- **T18**, M=0: `() => string`. Contract: `src/compiler/watchPublic.ts:184:5`.
- **T19**, M=1: `(name: string) => string \| undefined`. Contract: `src/compiler/watchPublic.ts:210:5`.
- **T20**, M=5: `(path: string, extensions?: readonly string[] \| undefined, exclude?: readonly string[] \| undefined, include?: readonly string[] \| undefined, depth?: number \| undefined) => string[]`. Contract: `src/compiler/watchPublic.ts:203:5`.
- **T21**, M=1: `(name: string) => string`. Contract: `src/compiler/sys.ts:1441:22`.
- **T22**, M=0: `() => Date`. Contract: `src/compiler/sys.ts:1454:22`.
- **T23**, M=0: `() => boolean`. Contract: `src/compiler/watchPublic.ts:278:5`, `src/compiler/watchPublic.ts:278:5`, `src/compiler/watchPublic.ts:278:5`, `src/compiler/watchPublic.ts:278:5`.
- **T24**, M=6: `(moduleNames: string[], containingFile: string, reusedNames: string[] \| undefined, redirectedReference: ResolvedProjectReference \| undefined, options: CompilerOptions, containingSourceFile?: SourceFile \| undefined) => (ResolvedModule \| undefined)[]`. Contract: `src/compiler/watchPublic.ts:217:5`, `src/compiler/watchPublic.ts:217:5`, `src/compiler/watchPublic.ts:217:5`, `src/compiler/watchPublic.ts:217:5`.
- **T25**, M=5: `(typeReferenceDirectiveNames: string[] \| readonly FileReference[], containingFile: string, redirectedReference: ResolvedProjectReference \| undefined, options: CompilerOptions, containingFileMode?: ResolutionMode) => (ResolvedTypeReferenceDirective \| undefined)[]`. Contract: `src/compiler/watchPublic.ts:223:5`, `src/compiler/watchPublic.ts:223:5`, `src/compiler/watchPublic.ts:223:5`, `src/compiler/watchPublic.ts:223:5`.
- **T26**, M=0: `() => ModuleResolutionCache \| undefined`. Contract: `src/compiler/watchPublic.ts:258:5`, `src/compiler/watchPublic.ts:258:5`, `src/compiler/watchPublic.ts:258:5`, `src/compiler/watchPublic.ts:258:5`.
- **T27**, M=1: `(filePath: Path) => boolean`. Contract: `src/compiler/watchPublic.ts:254:5`, `src/compiler/watchPublic.ts:254:5`, `src/compiler/watchPublic.ts:254:5`, `src/compiler/watchPublic.ts:254:5`.
- **T28**, M=1: `(libFileName: string) => boolean`. Contract: `src/compiler/watchPublic.ts:251:5`, `src/compiler/watchPublic.ts:251:5`, `src/compiler/watchPublic.ts:251:5`, `src/compiler/watchPublic.ts:251:5`.

## Conservative implementation catalogue (incomplete flow coverage)

This catalogue joins definitions/writes of the same property name across receivers, then follows source symbols, initializers, known call returns and argument edges. It deliberately includes unrelated same-name implementations, e.g. six-parameter `WatchFactory` adapters do not thereby become System’s four-parameter watcher. It also retains unresolved/dynamic/frontier cases. **It is not an exhaustive set of values proven to flow into each particular bound member.** That requested part remains unfinished. The narrower allocation pass is an aid, not a replacement completeness proof: it is context/flow insensitive and filters object allocations by host-bearing properties within depth two. Its fixed point took 228219 steps, with 192132 slots and 45237 call edges. All non-helper host inputs remain open except the three local resolutionCache object members.

No loaded class method has any of the bound member names (`candidates.json.gz`, class-only method-table audit). All enumerated host implementations are source functions, object-literal closures or wrappers. No harness/class-host body is silently imported from TypeScript’s tests. For properties with only wrapper cycles or no in-program definition, outside callers may supply a function; the source program does not identify its N.

- **I01 `fn`**: no source implementation enumerated.
  Audit frontiers: `parameter frontier src/compiler/core.ts:1519:58`.
- **I02 `getSymlinkCache`**: W001 (N=0), F020 (N=0).
- **I03 `getGlobalTypingsCacheLocation`**: W003 (N=0), W010 (N=0).
- **I04 `createHash`**: W013 (N=1), W030 (N=1), F061 (N=1), W086 (N=1), W096 (N=1), W098 (N=1).
- **I05 `resolveModuleNameLiterals`**: F040 (N=6), W063 (N=6), F069 (N=5), W100 (N=6), W102 (N=6).
- **I06 `resolveTypeReferenceDirectiveReferences`**: F045 (N=6), W064 (N=6), F071 (N=5), W103 (N=6), W105 (N=6).
- **I07 `resolveLibrary`**: F041 (N=4), W065 (N=4), F072 (N=3), W106 (N=4), W107 (N=4).
- **I08 `readFile`**: F002 (N=1), W007 (N=1), F011 (N=1), F021 (N=1), F032 (N=1), F037 (N=1), F055 (N=2), F079 (N=1), F088 (N=2), F110 (N=2).
- **I09 `realpath`**: W009 (N=1), W012 (N=1), F019 (N=1), F029 (N=1), W035 (N=1), F060 (N=1), W083 (N=1), W092 (N=1), F115 (N=1).
- **I10 `trace`**: W015 (N=1), F022 (N=1), F036 (N=1), W080 (N=1), F094 (N=1).
- **I11 `directoryExists`**: F023 (N=1), F024 (N=1), W033 (N=1), F039 (N=1), F058 (N=1), W081 (N=1), F089 (N=1), F112 (N=1).
  Audit frontiers: `call return frontier src/compiler/utilities.ts:7819:37`; `call return frontier src/compiler/watchUtilities.ts:329:30`; `unhandled PrefixUnaryExpression src/compiler/utilities.ts:7819:12`.
- **I12 `getDirectories`**: F026 (N=1), F028 (N=1), W034 (N=1), F059 (N=1), W082 (N=1), F090 (N=1), F113 (N=1).
- **I13 `writeFile`**: F004 (N=1), F006 (N=3), F016 (N=6), F017 (N=4), F038 (N=3), F049 (N=3), F056 (N=3), F095 (N=3), F111 (N=3).
  Audit frontiers: `parameter frontier src/compiler/program.ts:2599:26`; `parameter frontier src/compiler/program.ts:2651:28`; `parameter frontier src/compiler/program.ts:2691:9`; `parameter frontier src/compiler/program.ts:2722:9`.
- **I14 `watchFile`**: F046 (N=4), F073 (N=0), W074 (N=4), F116 (N=4), F118 (N=6), F119 (N=6), F121 (N=6).
  Audit frontiers: `no in-program definition for DynamicPriority`; `no in-program definition for DynamicPriorityPolling`; `no in-program definition for PriorityInterval`; `no in-program definition for PriorityPollingInterval`; `no in-program definition for UseFsEvents`; `no in-program definition for UseFsEventsOnParentDirectory`; `parameter frontier src/compiler/sys.ts:1092:9`; `parameter frontier src/compiler/sys.ts:1093:9`.
- **I15 `watchDirectory`**: F047 (N=4), F048 (N=4), F073 (N=0), W075 (N=4), F117 (N=4), F118 (N=6), F120 (N=6), F121 (N=6).
  Audit frontiers: `no in-program definition for DynamicPriorityPolling`; `no in-program definition for FixedPollingInterval`; `no in-program definition for UseFsEvents`.
- **I16 `setTimeout`**: F004 (N=1), F050 (N=3), F053 (N=3), W076 (N=3).
- **I17 `clearTimeout`**: F004 (N=1), F051 (N=1), F054 (N=1), W077 (N=1).
- **I18 `getDefaultLibLocation`**: F005 (N=0), F018 (N=0), W078 (N=0).
- **I19 `getEnvironmentVariable`**: F025 (N=1), F052 (N=1), W084 (N=1), F085 (N=0), W093 (N=1).
- **I20 `readDirectory`**: F027 (N=5), F031 (N=5), F057 (N=5), W087 (N=5), F091 (N=5), F114 (N=5).
- **I21 `now`**: W062 (N=0), W097 (N=0).
- **I22 `useSourceOfProjectReferenceRedirect`**: W099 (N=0).
  Audit frontiers: `unhandled PrefixUnaryExpression src/compiler/program.ts:1724:49`; `unhandled PrefixUnaryExpression src/compiler/program.ts:1725:9`.
- **I23 `resolveModuleNames`**: W066 (N=6), W101 (N=6).
- **I24 `resolveTypeReferenceDirectives`**: W067 (N=5), W104 (N=5).
- **I25 `getModuleResolutionCache`**: F008 (N=0), F014 (N=0), F042 (N=0), W068 (N=0), F070 (N=0), W108 (N=0), F109 (N=0).
- **I26 `hasInvalidatedResolutions`**: F043 (N=1).
  Audit frontiers: `parameter frontier src/compiler/watchPublic.ts:681:31`.
- **I27 `hasInvalidatedLibResolutions`**: F044 (N=1).
  Audit frontiers: `parameter frontier src/compiler/watchPublic.ts:681:85`.

### Implementation and wrapper source declarations

- **W001**, N=0: `src/compiler/checker.ts:54284:26`, generated by the source bind of `host.getSymlinkCache`. This is the adapter’s declared slot count, not the original target’s N.
- **F002**, N=1: `src/compiler/checker.ts:54292:36` `fileName =>`.
- **W003**, N=0: `src/compiler/checker.ts:54295:40`, generated by the source bind of `host.getGlobalTypingsCacheLocation`. This is the adapter’s declared slot count, not the original target’s N.
- **F004**, N=1: `src/compiler/core.ts:1797:1` `export function noop(_?: unknown): void`.
- **F005**, N=0: `src/compiler/core.ts:1896:12` `() =>`.
- **F006**, N=3: `src/compiler/emitter.ts:1374:5` `function writeFile(sourceFile: SourceFile, output: EmitTextWriter, sourceMapGenerator: SourceMapGenerator \| undefined)`.
- **W007**, N=1: `src/compiler/program.ts:1735:22`, generated by the source bind of `host.readFile`. This is the adapter’s declared slot count, not the original target’s N.
- **F008**, N=0: `src/compiler/program.ts:1883:35` `() =>`.
- **W009**, N=1: `src/compiler/program.ts:1951:19`, generated by the source bind of `host.realpath`. This is the adapter’s declared slot count, not the original target’s N.
- **W010**, N=0: `src/compiler/program.ts:1957:40`, generated by the source bind of `host.getGlobalTypingsCacheLocation`. This is the adapter’s declared slot count, not the original target’s N.
- **F011**, N=1: `src/compiler/program.ts:2618:23` `f =>`.
- **W012**, N=1: `src/compiler/program.ts:2627:23`, generated by the source bind of `host.realpath`. This is the adapter’s declared slot count, not the original target’s N.
- **W013**, N=1: `src/compiler/program.ts:2633:25`, generated by the source bind of `host.createHash`. This is the adapter’s declared slot count, not the original target’s N.
- **F014**, N=0: `src/compiler/program.ts:2634:39` `() =>`.
- **W015**, N=1: `src/compiler/program.ts:2635:20`, generated by the source bind of `host.trace`. This is the adapter’s declared slot count, not the original target’s N.
- **F016**, N=6: `src/compiler/program.ts:2640:5` `function writeFile(         fileName: string,         text: string,         writeByteOrderMark: boolean,         onError?: (message: string) => void,         sourceFiles?: readonly SourceFile[],         data?: WriteFileCallbackData,     )`.
- **F017**, N=4: `src/compiler/program.ts:421:12` `(fileName, data, writeByteOrderMark, onError) =>`.
- **F018**, N=0: `src/compiler/program.ts:467:5` `function getDefaultLibLocation(): string`.
- **F019**, N=1: `src/compiler/program.ts:472:42` `(path: string) =>`.
- **F020**, N=0: `src/compiler/program.ts:4798:5` `function getSymlinkCache(): SymlinkCache`.
- **F021**, N=1: `src/compiler/program.ts:487:19` `fileName =>`.
- **F022**, N=1: `src/compiler/program.ts:488:16` `(s: string) =>`.
- **F023**, N=1: `src/compiler/program.ts:489:26` `directoryName =>`.
- **F024**, N=1: `src/compiler/program.ts:4908:63` `path =>`.
- **F025**, N=1: `src/compiler/program.ts:490:33` `name =>`.
- **F026**, N=1: `src/compiler/program.ts:491:25` `(path: string) =>`.
- **F027**, N=5: `src/compiler/program.ts:493:24` `(path, extensions, include, exclude, depth) =>`.
- **F028**, N=1: `src/compiler/program.ts:4940:44` `path =>`.
- **F029**, N=1: `src/compiler/program.ts:4948:38` `s =>`.
- **W030**, N=1: `src/compiler/program.ts:495:21`, generated by the source bind of `system.createHash`. This is the adapter’s declared slot count, not the original target’s N.
- **F031**, N=5: `src/compiler/program.ts:5110:9` `readDirectory(root, extensions, excludes, includes, depth)`.
- **F032**, N=1: `src/compiler/program.ts:5114:19` `f =>`.
- **W033**, N=1: `src/compiler/program.ts:5115:26`, generated by the source bind of `directoryStructureHost.directoryExists`. This is the adapter’s declared slot count, not the original target’s N.
- **W034**, N=1: `src/compiler/program.ts:5116:25`, generated by the source bind of `directoryStructureHost.getDirectories`. This is the adapter’s declared slot count, not the original target’s N.
- **W035**, N=1: `src/compiler/program.ts:5117:19`, generated by the source bind of `directoryStructureHost.realpath`. This is the adapter’s declared slot count, not the original target’s N.
- **F036**, N=1: `src/compiler/program.ts:5121:29` `s =>`.
- **F037**, N=1: `src/compiler/program.ts:547:21` `fileName =>`.
- **F038**, N=3: `src/compiler/program.ts:583:26` `(fileName, data, ...rest) =>`; rest parameter present.
- **F039**, N=1: `src/compiler/program.ts:606:32` `directory =>`.
- **F040**, N=6: `src/compiler/resolutionCache.ts:1013:5` `function resolveModuleNameLiterals(         moduleLiterals: readonly StringLiteralLike[],         containingFile: string,         redirectedReference: ResolvedProjectReference \| undefined,         options: CompilerOptions,         containingSourceFile: SourceFile,         reusedNames: readonly StringLiteralLike[] \| undefined,     ): readonly ResolvedModuleWithFailedLookupLocations[]`.
- **F041**, N=4: `src/compiler/resolutionCache.ts:1043:5` `function resolveLibrary(         libraryName: string,         resolveFrom: string,         options: CompilerOptions,         libFileName: string,     )`.
- **F042**, N=0: `src/compiler/resolutionCache.ts:657:35` `() =>`.
- **F043**, N=1: `src/compiler/resolutionCache.ts:748:40` `path =>`.
- **F044**, N=1: `src/compiler/resolutionCache.ts:753:43` `libFileName =>`.
- **F045**, N=6: `src/compiler/resolutionCache.ts:984:5` `function resolveTypeReferenceDirectiveReferences<T extends FileReference \| string>(         typeDirectiveReferences: readonly T[],         containingFile: string,         redirectedReference: ResolvedProjectReference \| undefined,         options: CompilerOptions,         containingSourceFile: SourceFile \| undefined,         reusedNames: readonly T[] \| undefined,     ): readonly ResolvedTypeReferenceDirectiveWithFailedLookupLocations[]`.
- **F046**, N=4: `src/compiler/sys.ts:1025:5` `function watchFile(fileName: string, callback: FileWatcherCallback, pollingInterval: PollingInterval, options: WatchOptions \| undefined): FileWatcher`.
- **F047**, N=4: `src/compiler/sys.ts:1105:5` `function watchDirectory(directoryName: string, callback: DirectoryWatcherCallback, recursive: boolean, options: WatchOptions \| undefined): FileWatcher`.
- **F048**, N=4: `src/compiler/sys.ts:1132:5` `function nonRecursiveWatchDirectory(directoryName: string, callback: DirectoryWatcherCallback, recursive: boolean, options: WatchOptions \| undefined): FileWatcher`.
- **F049**, N=3: `src/compiler/sys.ts:1386:21` `(path, data, writeBom) =>`.
- **F050**, N=3: `src/compiler/sys.ts:1499:25` `(callback: (label?: unknown) => void, ms: number, label?: unknown): unknown =>`.
- **F051**, N=1: `src/compiler/sys.ts:1500:27` `(timeoutId: Parameters<typeof clearTimeout>[0]): void =>`.
- **F052**, N=1: `src/compiler/sys.ts:1559:13` `getEnvironmentVariable(name: string)`.
- **F053**, N=3: `src/compiler/sys.ts:1597:25` `(callback: (label?: unknown) => void, ms: number, label?: unknown): unknown =>`.
- **F054**, N=1: `src/compiler/sys.ts:1598:27` `(timeoutId: Parameters<typeof clearTimeout>[0]): void =>`.
- **F055**, N=2: `src/compiler/sys.ts:1781:9` `function readFile(fileName: string, _encoding?: string): string \| undefined`.
- **F056**, N=3: `src/compiler/sys.ts:1813:9` `function writeFile(fileName: string, data: string, writeByteOrderMark?: boolean): void`.
- **F057**, N=5: `src/compiler/sys.ts:1876:9` `function readDirectory(path: string, extensions?: readonly string[], excludes?: readonly string[], includes?: readonly string[], depth?: number): string[]`.
- **F058**, N=1: `src/compiler/sys.ts:1899:9` `function directoryExists(path: string): boolean`.
- **F059**, N=1: `src/compiler/sys.ts:1903:9` `function getDirectories(path: string): string[]`.
- **F060**, N=1: `src/compiler/sys.ts:1911:9` `function realpath(path: string): string`.
- **F061**, N=1: `src/compiler/sys.ts:1942:9` `function createSHA256Hash(data: string): string`.
- **W062**, N=0: `src/compiler/tsbuildPublic.ts:301:16`, generated by the source bind of `system.now`. This is the adapter’s declared slot count, not the original target’s N.
- **W063**, N=6: `src/compiler/tsbuildPublic.ts:437:46`, generated by the source bind of `host.resolveModuleNameLiterals`. This is the adapter’s declared slot count, not the original target’s N.
- **W064**, N=6: `src/compiler/tsbuildPublic.ts:438:60`, generated by the source bind of `host.resolveTypeReferenceDirectiveReferences`. This is the adapter’s declared slot count, not the original target’s N.
- **W065**, N=4: `src/compiler/tsbuildPublic.ts:439:35`, generated by the source bind of `host.resolveLibrary`. This is the adapter’s declared slot count, not the original target’s N.
- **W066**, N=6: `src/compiler/tsbuildPublic.ts:440:39`, generated by the source bind of `host.resolveModuleNames`. This is the adapter’s declared slot count, not the original target’s N.
- **W067**, N=5: `src/compiler/tsbuildPublic.ts:441:51`, generated by the source bind of `host.resolveTypeReferenceDirectives`. This is the adapter’s declared slot count, not the original target’s N.
- **W068**, N=0: `src/compiler/tsbuildPublic.ts:442:45`, generated by the source bind of `host.getModuleResolutionCache`. This is the adapter’s declared slot count, not the original target’s N.
- **F069**, N=5: `src/compiler/tsbuildPublic.ts:446:50` `(moduleNames, containingFile, redirectedReference, options, containingSourceFile) =>`.
- **F070**, N=0: `src/compiler/tsbuildPublic.ts:457:49` `() =>`.
- **F071**, N=5: `src/compiler/tsbuildPublic.ts:467:64` `(typeDirectiveNames, containingFile, redirectedReference, options, containingSourceFile) =>`.
- **F072**, N=3: `src/compiler/tsbuildPublic.ts:482:39` `(libraryName, resolveFrom, options) =>`.
- **F073**, N=0: `src/compiler/watch.ts:669:38` `(): FileWatcher =>`.
- **W074**, N=4: `src/compiler/watch.ts:676:20`, generated by the source bind of `system.watchFile`. This is the adapter’s declared slot count, not the original target’s N.
- **W075**, N=4: `src/compiler/watch.ts:677:25`, generated by the source bind of `system.watchDirectory`. This is the adapter’s declared slot count, not the original target’s N.
- **W076**, N=3: `src/compiler/watch.ts:678:21`, generated by the source bind of `system.setTimeout`. This is the adapter’s declared slot count, not the original target’s N.
- **W077**, N=1: `src/compiler/watch.ts:679:23`, generated by the source bind of `system.clearTimeout`. This is the adapter’s declared slot count, not the original target’s N.
- **W078**, N=0: `src/compiler/watch.ts:760:32`, generated by the source bind of `host.getDefaultLibLocation`. This is the adapter’s declared slot count, not the original target’s N.
- **F079**, N=1: `src/compiler/watch.ts:772:19` `f =>`.
- **W080**, N=1: `src/compiler/watch.ts:773:16`, generated by the source bind of `host.trace`. This is the adapter’s declared slot count, not the original target’s N.
- **W081**, N=1: `src/compiler/watch.ts:774:26`, generated by the source bind of `directoryStructureHost.directoryExists`. This is the adapter’s declared slot count, not the original target’s N.
- **W082**, N=1: `src/compiler/watch.ts:775:25`, generated by the source bind of `directoryStructureHost.getDirectories`. This is the adapter’s declared slot count, not the original target’s N.
- **W083**, N=1: `src/compiler/watch.ts:776:19`, generated by the source bind of `host.realpath`. This is the adapter’s declared slot count, not the original target’s N.
- **W084**, N=1: `src/compiler/watch.ts:777:33`, generated by the source bind of `host.getEnvironmentVariable`. This is the adapter’s declared slot count, not the original target’s N.
- **F085**, N=0: `src/compiler/watch.ts:777:77` `() =>`.
- **W086**, N=1: `src/compiler/watch.ts:778:21`, generated by the source bind of `host.createHash`. This is the adapter’s declared slot count, not the original target’s N.
- **W087**, N=5: `src/compiler/watch.ts:779:24`, generated by the source bind of `host.readDirectory`. This is the adapter’s declared slot count, not the original target’s N.
- **F088**, N=2: `src/compiler/watch.ts:851:19` `(path, encoding) =>`.
- **F089**, N=1: `src/compiler/watch.ts:852:26` `path =>`.
- **F090**, N=1: `src/compiler/watch.ts:853:25` `path =>`.
- **F091**, N=5: `src/compiler/watch.ts:854:24` `(path, extensions, exclude, include, depth) =>`.
- **W092**, N=1: `src/compiler/watch.ts:855:19`, generated by the source bind of `system.realpath`. This is the adapter’s declared slot count, not the original target’s N.
- **W093**, N=1: `src/compiler/watch.ts:856:33`, generated by the source bind of `system.getEnvironmentVariable`. This is the adapter’s declared slot count, not the original target’s N.
- **F094**, N=1: `src/compiler/watch.ts:857:16` `s =>`.
- **F095**, N=3: `src/compiler/watch.ts:859:20` `(path, data, writeByteOrderMark) =>`.
- **W096**, N=1: `src/compiler/watch.ts:860:21`, generated by the source bind of `system.createHash`. This is the adapter’s declared slot count, not the original target’s N.
- **W097**, N=0: `src/compiler/watch.ts:863:14`, generated by the source bind of `system.now`. This is the adapter’s declared slot count, not the original target’s N.
- **W098**, N=1: `src/compiler/watchPublic.ts:124:23`, generated by the source bind of `system.createHash`. This is the adapter’s declared slot count, not the original target’s N.
- **W099**, N=0: `src/compiler/watchPublic.ts:505:56`, generated by the source bind of `host.useSourceOfProjectReferenceRedirect`. This is the adapter’s declared slot count, not the original target’s N.
- **W100**, N=6: `src/compiler/watchPublic.ts:528:46`, generated by the source bind of `host.resolveModuleNameLiterals`. This is the adapter’s declared slot count, not the original target’s N.
- **W101**, N=6: `src/compiler/watchPublic.ts:529:39`, generated by the source bind of `host.resolveModuleNames`. This is the adapter’s declared slot count, not the original target’s N.
- **W102**, N=6: `src/compiler/watchPublic.ts:531:50`, generated by the source bind of `resolutionCache.resolveModuleNameLiterals`. This is the adapter’s declared slot count, not the original target’s N.
- **W103**, N=6: `src/compiler/watchPublic.ts:533:60`, generated by the source bind of `host.resolveTypeReferenceDirectiveReferences`. This is the adapter’s declared slot count, not the original target’s N.
- **W104**, N=5: `src/compiler/watchPublic.ts:534:51`, generated by the source bind of `host.resolveTypeReferenceDirectives`. This is the adapter’s declared slot count, not the original target’s N.
- **W105**, N=6: `src/compiler/watchPublic.ts:536:64`, generated by the source bind of `resolutionCache.resolveTypeReferenceDirectiveReferences`. This is the adapter’s declared slot count, not the original target’s N.
- **W106**, N=4: `src/compiler/watchPublic.ts:539:9`, generated by the source bind of `resolutionCache.resolveLibrary`. This is the adapter’s declared slot count, not the original target’s N.
- **W107**, N=4: `src/compiler/watchPublic.ts:540:9`, generated by the source bind of `host.resolveLibrary`. This is the adapter’s declared slot count, not the original target’s N.
- **W108**, N=0: `src/compiler/watchPublic.ts:542:9`, generated by the source bind of `host.getModuleResolutionCache`. This is the adapter’s declared slot count, not the original target’s N.
- **F109**, N=0: `src/compiler/watchPublic.ts:543:10` `() =>`.
- **F110**, N=2: `src/compiler/watchUtilities.ts:128:19` `(path, encoding) =>`.
- **F111**, N=3: `src/compiler/watchUtilities.ts:215:5` `function writeFile(fileName: string, data: string, writeByteOrderMark?: boolean): void`.
- **F112**, N=1: `src/compiler/watchUtilities.ts:231:5` `function directoryExists(dirPath: string): boolean`.
- **F113**, N=1: `src/compiler/watchUtilities.ts:251:5` `function getDirectories(rootDir: string): string[]`.
- **F114**, N=5: `src/compiler/watchUtilities.ts:260:5` `function readDirectory(rootDir: string, extensions?: readonly string[], excludes?: readonly string[], includes?: readonly string[], depth?: number): string[]`.
- **F115**, N=1: `src/compiler/watchUtilities.ts:291:5` `function realpath(s: string)`.
- **F116**, N=4: `src/compiler/watchUtilities.ts:703:20` `(file, callback, pollingInterval, options) =>`.
- **F117**, N=4: `src/compiler/watchUtilities.ts:704:25` `(directory, callback, flags, options) =>`.
- **F118**, N=6: `src/compiler/watchUtilities.ts:728:16` `(             file: string,             cb: FileWatcherCallback \| DirectoryWatcherCallback,             flags: PollingInterval \| WatchDirectoryFlags,             options: WatchOptions \| undefined,             detailInfo1: X,             detailInfo2?: Y,         ) =>`.
- **F119**, N=6: `src/compiler/watchUtilities.ts:759:5` `function createFileWatcherWithLogging(         file: string,         cb: FileWatcherCallback,         flags: PollingInterval,         options: WatchOptions \| undefined,         detailInfo1: X,         detailInfo2?: Y,     ): FileWatcher`.
- **F120**, N=6: `src/compiler/watchUtilities.ts:777:5` `function createDirectoryWatcherWithLogging(         file: string,         cb: DirectoryWatcherCallback,         flags: WatchDirectoryFlags,         options: WatchOptions \| undefined,         detailInfo1: X,         detailInfo2?: Y,     ): FileWatcher`.
- **F121**, N=6: `src/compiler/watchUtilities.ts:804:16` `(             file: string,             cb: FileWatcherCallback \| DirectoryWatcherCallback,             flags: PollingInterval \| WatchDirectoryFlags,             options: WatchOptions \| undefined,             detailInfo1: X,             detailInfo2?: Y,         ) =>`.

## Source-pool stop evidence and limits

For every site, `candidates.json.gz` retains **all** source callable-body declarations assignable under the stock checker to its target type, all N > M candidates, argument readers/rest flags and class-name candidates. Constructors/accessors are excluded from this callable pool. This pool is not renamed as the head’s `FunctionTypeTargets`: it can include uninstantiated generic/nested bodies, and does not enumerate every generated/specialized IR record. The positive witnesses below establish why the new fail-closed count predicate matters; they do not certify the entire IR candidate universe or the absence of an earlier generated stop.

| Site | M | Explicit source witness | N |
|---|---:|---|---:|
| B02 | 0 | `src/compiler/debug.ts:174:5` `fail` | 2 |
| B03 | 0 | `src/compiler/debug.ts:174:5` `fail` | 2 |
| B10 | 0 | `src/compiler/debug.ts:174:5` `fail` | 2 |
| B14 | 1 | `src/compiler/debug.ts:174:5` `fail` | 2 |
| B15 | 1 | `src/compiler/debug.ts:174:5` `fail` | 2 |
| B22 | 3 | `src/compiler/debug.ts:324:5` `assertOptionalToken` | 4 |
| B24 | 0 | `src/compiler/debug.ts:174:5` `fail` | 2 |
| B26 | 1 | `src/compiler/debug.ts:174:5` `fail` | 2 |
| B27 | 1 | `src/compiler/debug.ts:174:5` `fail` | 2 |
| B35 | 0 | `src/compiler/debug.ts:174:5` `fail` | 2 |
| B46 | 0 | `src/compiler/debug.ts:174:5` `fail` | 2 |
| B47 | 1 | `src/compiler/debug.ts:174:5` `fail` | 2 |
| B48 | 1 | `src/compiler/debug.ts:174:5` `fail` | 2 |
| B49 | 0 | `src/compiler/debug.ts:174:5` `fail` | 2 |
| B55 | 0 | `src/compiler/debug.ts:174:5` `fail` | 2 |

`Debug.fail` is declared at adapted `src/compiler/debug.ts:174` as `fail(message?: string, stackCrawlMark?: AnyFunction): never` (N=2); TypeScript permits it through the zero-/one-slot target signatures above. `assertOptionalToken` at `debug.ts:321–324` has overloads and a four-parameter implementation; the checker relates its overload type to the three-slot setTimeout target. Neither witness need ever be assigned to the bound member: that is exactly what the unknown-reach fallback counts.

For string-returning and void targets, unrelated `combinePaths(path: string, ...paths: (string | undefined)[]): string` (`path.ts:581`) and `resolvePath` (`:608`) are rest-bearing assignable candidates. The cached writeFile closure (`program.ts:583`) and generic predicate closure (`core.ts:2489`) account for the other recorded source rest candidates. A rest flag inferred from an uninstantiated generic body is not a certified `RestElement` in the missing completed IR. P is retained as an earlier-stop bucket rather than inflated into “new count stops”.

**Stopping reason / remaining coverage:** exact per-site, exhaustive implementation-flow sets and a complete reconstruction of specialized/generated IR closure records are not established by these API analyses. In particular, a field-name join is not a receiver-flow proof and a stock-checker generic signature is not a record of every concrete Adamic specialization. Publishing the 15/19 projection as an unconditional 56-site a/b/c verdict would hide that gap. There are no observed native tsc wrong answers or new whole-program stops in this report. The rule and all 56 sites/types are complete; exhaustive flow certification is partial.

## Downstream uses of the fifteen a* bound values

**6 of the 15 a* sites have a downstream callback, forwarding store/copy, rebinding or indirect call:** **B10** `program.ts:1957`, **B22** `watch.ts:678`, **B26** `watch.ts:774`, **B27** `watch.ts:775`, **B47** `watchPublic.ts:548`, and **B48** `watchPublic.ts:551`. **None of the identified caller sites passes more function arguments than that origin’s M.** In particular, no selected bound value is used as an array-method callback, invoked through `.apply`, or invoked through `Reflect` in this audit.

Counting non-invoking presence/comparison guards as well gives a **conservative 11/15**: `B03` (`checker.ts:54295`), `B10` (`program.ts:1957`), `B14` (`program.ts:5115`), `B22` (`watch.ts:678`), `B24` (`watch.ts:760`), `B26` (`watch.ts:774`), `B27` (`watch.ts:775`), `B35` (`watch.ts:863`), `B47` (`watchPublic.ts:548`), `B48` (`watchPublic.ts:551`), `B49` (`tsbuildPublic.ts:301`). The six-site headline excludes those guards and excludes each bind’s initial storage; otherwise initial storage alone would count every site. Guards are listed separately below. The guard-inclusive count is a may-flow count: the shared `ProgramHost.now`/solution-host path includes the later overwrite at B49, so it is not proof that the B35 wrapper itself executes at `host.now()`.

This follows the bound result and its receiver-only rebindings, not every unrelated function with the same member name. It follows field reads, local aliases, shorthand/destructured cache records, source-call arguments/returns and the source `copyProperties`/`assign`/`clone`/`extend` helpers. A fresh closure which calls the bound function is inspected at that call; its own identity is not silently equated with the captured bound function. Returned host containers are followed to their local consumers; their initial return is not counted as another function-valued store. The scope is the same 79 source roots, with no harness/server/services code and no invented outside caller bodies.

### All fifteen origins

| Origin | Initial storage of the bound result | M | Downstream non-direct uses | Ordinary typed calls found (argument counts) | More than M at an identified caller? |
|---|---|---:|---|---|---|
| B02 `checker.ts:54284` | `ModuleSpecifierResolutionHost.getSymlinkCache` | 0 | none found | `moduleSpecifiers.ts:736` (0), `moduleSpecifiers.ts:811` (0) | No identified case |
| B03 `checker.ts:54295` | `ModuleSpecifierResolutionHost.getGlobalTypingsCacheLocation` | 0 | none found | `moduleNameResolver.ts:3056` (0), `moduleSpecifiers.ts:1235` (0) | No identified case |
| B10 `program.ts:1957` | `Program.getGlobalTypingsCacheLocation` | 0 | U001, U020 | `moduleNameResolver.ts:3056` (0), `moduleSpecifiers.ts:1235` (0) | No identified case |
| B14 `program.ts:5115` | `ParseConfigFileHost.directoryExists` | 1 | none found | `utilities.ts:7819` (1) | No identified case |
| B15 `program.ts:5116` | `ParseConfigFileHost.getDirectories` | 1 | none found | none found | No identified case |
| B22 `watch.ts:678` | `WatchHost.setTimeout` | 3 | U037, U042 | `tsbuildPublic.ts:2074` (3), `watchPublic.ts:872` (3), `watchPublic.ts:900` (3) | No identified case |
| B24 `watch.ts:760` | `CompilerHost.getDefaultLibLocation` | 0 | none found | `program.ts:1582` (0) | No identified case |
| B26 `watch.ts:774` | `CompilerHost.directoryExists` | 1 | U012, U014, U015, U021, U024, U027, U029, U031, U045, U046 | `moduleNameResolver.ts:824` (1), `program.ts:460` (1), `utilities.ts:7819` (1) | No identified case |
| B27 `watch.ts:775` | `CompilerHost.getDirectories` | 1 | U022, U028, U030, U032 | `moduleNameResolver.ts:825` (1) | No identified case |
| B35 `watch.ts:863` | `ProgramHost<T>.now` | 0 | none found | `tsbuildPublic.ts:205` (0) | No identified case |
| B46 `watchPublic.ts:542` | `compilerHost.getModuleResolutionCache` | 0 | none found | `program.ts:1606` (0), `program.ts:1625` (0) | No identified case |
| B47 `watchPublic.ts:548` | `customHasInvalidatedResolutions` | 1 | U043 | `resolutionCache.ts:749` (1) | No identified case |
| B48 `watchPublic.ts:551` | `customHasInvalidLibResolutions` | 1 | U044 | `resolutionCache.ts:754` (1) | No identified case |
| B49 `tsbuildPublic.ts:301` | `host.now` | 0 | none found | `tsbuildPublic.ts:205` (0) | No identified case |
| B55 `tsbuildPublic.ts:442` | `compilerHost.getModuleResolutionCache` | 0 | none found | `program.ts:1606` (0), `program.ts:1625` (0) | No identified case |

The storage contracts and declaration locations are recorded in `evidence/a-star-uses.json.gz`. For B46/B55 the destination is `compilerHost.getModuleResolutionCache`; for B49 it is the solution-builder host’s `now`. B47/B48 first enter the two local `customHasInvalidated...` variables rather than a field.

### Every identified downstream callback/store/copy/rebind/indirect-call use

Each row is an API-discovered read of a bound identity or its receiver-only rebound descendant. `.call`’s first argument is the receiver and is excluded from the function-argument count. Rebinding here supplies no fixed function arguments. Copies/stores do not invoke the value; the caller column names the downstream invocation evidence. The source alias/call fields are retained for all origins, including the nine with no such use.

| Use | Origin(s) | Source quote | Form | Can this use introduce more arguments than M? |
|---|---|---|---|---|
| U001 | B10 | `src/compiler/checker.ts:54295:40` `host.getGlobalTypingsCacheLocation?.bind(host)` | bind | No invocation here; 0 fixed function arguments. Descendant calls remain at the original M. |
| U012 | B26 | `src/compiler/program.ts:528:11` `originalDirectoryExists = host.directoryExists` | alias-store | No invocation: alias/field movement. Its identified indirect/direct callers pass one path argument. |
| U014 | B26 | `src/compiler/program.ts:610:30` `originalDirectoryExists.call(host, directory)` | call | No: exactly 1 function argument(s); no spread. |
| U015 | B26 | `src/compiler/program.ts:627:9` `originalDirectoryExists` | field-store | No invocation: alias/field movement. Its identified indirect/direct callers pass one path argument. |
| U020 | B10 | `src/compiler/program.ts:2636:13` `getGlobalTypingsCacheLocation: program.getGlobalTypingsCacheLocation` | field-store | No invocation: alias/field movement. The getter remains a zero-argument getter. |
| U021 | B26 | `src/compiler/program.ts:4895:11` `originalDirectoryExists = host.compilerHost.directoryExists` | alias-store | No invocation: alias/field movement. Its identified indirect/direct callers pass one path argument. |
| U022 | B27 | `src/compiler/program.ts:4896:11` `originalGetDirectories = host.compilerHost.getDirectories` | alias-store | No invocation: alias/field movement. Its identified indirect/direct callers pass one path argument. |
| U024 | B26 | `src/compiler/program.ts:4909:17` `originalDirectoryExists.call(host.compilerHost, path)` | call | No: exactly 1 function argument(s); no spread. |
| U027 | B26 | `src/compiler/program.ts:4941:81` `originalDirectoryExists.call(host.compilerHost, path)` | call | No: exactly 1 function argument(s); no spread. |
| U028 | B27 | `src/compiler/program.ts:4942:17` `originalGetDirectories.call(host.compilerHost, path)` | call | No: exactly 1 function argument(s); no spread. |
| U029 | B26 | `src/compiler/program.ts:4957:9` `host.compilerHost.directoryExists = originalDirectoryExists` | store | No invocation: alias/field movement. Its identified indirect/direct callers pass one path argument. |
| U030 | B27 | `src/compiler/program.ts:4958:9` `host.compilerHost.getDirectories = originalGetDirectories` | store | No invocation: alias/field movement. Its identified indirect/direct callers pass one path argument. |
| U031 | B26 | `src/compiler/program.ts:5115:26` `directoryStructureHost.directoryExists?.bind(directoryStructureHost)` | bind | No invocation here; 0 fixed function arguments. Descendant calls remain at the original M. |
| U032 | B27 | `src/compiler/program.ts:5116:25` `directoryStructureHost.getDirectories?.bind(directoryStructureHost)` | bind | No invocation here; 0 fixed function arguments. Descendant calls remain at the original M. |
| U037 | B22 | `src/compiler/tsbuildPublic.ts:326:5` `copyProperties(host, watchHost)` | object-copy | No invocation: own-property copy; the copied timer is called with 3 arguments (M=3). See core.ts:1513 below. |
| U042 | B22 | `src/compiler/watch.ts:873:5` `copyProperties(result, createWatchHost(system, reportWatchStatus))` | object-copy | No invocation: own-property copy; the copied timer is called with 3 arguments (M=3). See core.ts:1513 below. |
| U043 | B47 | `src/compiler/watchPublic.ts:628:77` `resolutionCache.createHasInvalidatedResolutions(customHasInvalidatedResolutions, customHasInvalidLibResolutions)` | callback-or-argument | No: the source callback consumer invokes it with one argument: `resolutionCache.ts:749` `customHasInvalidatedResolutions(path)` (1 argument). |
| U044 | B48 | `src/compiler/watchPublic.ts:628:77` `resolutionCache.createHasInvalidatedResolutions(customHasInvalidatedResolutions, customHasInvalidLibResolutions)` | callback-or-argument | No: the source callback consumer invokes it with one argument: `resolutionCache.ts:754` `customHasInvalidatedLibResolutions(libFileName)` (1 argument). |
| U045 | B26 | `src/compiler/watchPublic.ts:632:13` `originalDirectoryExists` | destructure-store | No invocation: alias/field movement. Its identified indirect/direct callers pass one path argument. |
| U046 | B26 | `src/compiler/watchPublic.ts:660:9` `compilerHost.directoryExists = originalDirectoryExists` | store | No invocation: alias/field movement. Its identified indirect/direct callers pass one path argument. |

`copyProperties` performs the actual dynamic read and store at **`src/compiler/core.ts:1513`**, quoted exactly:

```typescript
first[id] = second[id];
```

For B22, the two call sites above use `second` equal to the watch host, and `id = "setTimeout"` copies that bound value. The receiver/label timer calls are `tsbuildPublic.ts:2074`, `watchPublic.ts:872`, and `watchPublic.ts:900`; each passes exactly `(callback, timeoutDelay, t1)`, three function arguments, matching M=3. Being the *first/target* argument of `copyProperties` does not read an existing `now` or directory method; the key set comes from `second`. This prevents counting unrelated existing target members as spread uses.

For B26, the cache forwarding is included, not stopped at the first alias: `changeCompilerHostLikeToUseCache` captures `originalDirectoryExists`, returns it as a shorthand field, the watch caller destructures that record, and later restores `compilerHost.directoryExists`. Its `originalDirectoryExists.call(host, directory)` passes one argument. The originals record is modeled **per helper call**: the solution builder’s separate call on its `ProgramHost` does not inherit the watch compiler host’s bound method merely because both calls use the same helper body. Any B26 attribution to that separate solution-host restoration is excluded.

B47/B48 are passed as callbacks to `resolutionCache.createHasInvalidatedResolutions` at `watchPublic.ts:628`. The actual implementation is `resolutionCache.ts:739`; it captures them in new one-parameter predicates and invokes them at `:749` and `:754` as `customHasInvalidatedResolutions(path)` and `customHasInvalidatedLibResolutions(libFileName)`. It supplies no array index, array object, rest tail or spread. The later predicate-field aliases in `program.ts:1603` and `:1684` refer to those fresh wrapper predicates, not directly to the original two bound-function identities; their calls into the original identities are the two explicit one-argument calls just listed.

### Presence/comparison reads (no function invocation)

| Read | Origin(s) in the may-flow audit | Quoted value read | Can add function arguments? |
|---|---|---|---|
| U002 `src/compiler/moduleNameResolver.ts:820:9` | B26 | `host.directoryExists` | No: presence/comparison only |
| U003 `src/compiler/moduleNameResolver.ts:820:33` | B27 | `host.getDirectories` | No: presence/comparison only |
| U009 `src/compiler/moduleSpecifiers.ts:1235:40` | B03, B10 | `host.getGlobalTypingsCacheLocation` | No: presence/comparison only |
| U013 `src/compiler/program.ts:605:9` | B26 | `originalDirectoryExists` | No: presence/comparison only |
| U016 `src/compiler/program.ts:1582:32` | B24 | `host.getDefaultLibLocation` | No: presence/comparison only |
| U023 `src/compiler/program.ts:4904:9` | B26 | `originalDirectoryExists` | No: presence/comparison only |
| U025 `src/compiler/program.ts:4937:9` | B27 | `originalGetDirectories` | No: presence/comparison only |
| U026 `src/compiler/program.ts:4941:54` | B26 | `originalDirectoryExists` | No: presence/comparison only |
| U035 `src/compiler/tsbuildPublic.ts:205:12` | B35, B49 | `host.now` | No: presence/comparison only |
| U038 `src/compiler/tsbuildPublic.ts:2061:10` | B22 | `hostWithWatch.setTimeout` | No: presence/comparison only |
| U040 `src/compiler/utilities.ts:7819:13` | B14, B26 | `host.directoryExists` | No: presence/comparison only |
| U047 `src/compiler/watchPublic.ts:862:14` | B22 | `host.setTimeout` | No: presence/comparison only |
| U049 `src/compiler/watchPublic.ts:887:14` | B22 | `host.setTimeout` | No: presence/comparison only |

### Cross-checks and coverage

The API pass scanned **257367** member/alias read expressions, **52426** source calls and **51** object-spread syntax nodes. It retained 50 attributed read/use records, 38 non-direct callable member-name reads for the independent surface cross-check, and 546 conservative host-carrier argument records. The latter are an over-approximation used to inspect copy/index/Reflect paths, not additional function invocation sites. Callable identities are constrained by checker types; generic callback contracts stay open. Object carriers stay conservative. The originals cache record is modeled separately at each call, including its shorthand return fields and destructured aliases; the audit reports any other function-valued literal/array carrier it cannot model. There are zero such unresolved carriers in this result.

The independent non-direct member-read cross-check leaves only three non-guard/non-bind same-name cases outside the attributed identity set: native `Date.now` at `performanceCore.ts:103` (a different symbol and number-valued contract), and the two fresh invalidation-wrapper aliases at `program.ts:1603`/`:1684` explained above. Source index/Reflect carrier reads were inspected; the relevant dynamic function-valued read is `second[id]` in the explicit copy helper. No attributed `.apply`, `Reflect` invocation, array-method callback or untyped direct call remains in the records. All recorded direct calls have fixed argument lists of size ≤ M, and both callback consumers have source bodies with one-argument invocation lists.

**Limit:** this is a compiler-API static may-use audit, not execution tracing or a theorem of exact path feasibility. It is flow/context insensitive except for the cache originals record, retains source field/call contracts and explicit unknown host boundaries, and cannot constrain code outside the 79 roots. The six positive sites and their quoted uses are reviewable; “none found” is scoped to this audit. In particular, the result does not assert that outside JavaScript callers can never supply extra arguments. No native tsc run or production lowering change was used.

Reproduce the additional audit with `node stage3/census/bind-sites-static/uses-run.cjs ADAPTED_TREE NEW_OUTPUT_DIRECTORY`, then render with `python3 stage3/census/bind-sites-static/write-uses.py NEW_OUTPUT_DIRECTORY`. The renderer checks the 15 origins, six affected sites, fixed direct-call vectors, callback body vectors and absence of an attributed apply/Reflect/array-callback use before writing this section. `write-report.py` incorporates the saved section without changing the original measurements.

## Reproduction and provenance

Pins: rule `ba619e28774130e24331c3026a938cd6f3cd6a34`; adapters/base `55c375d9d245b580c010c3e0b5659f2849f24fb6`; TypeScript v6.0.3 `050880ce59e30b356b686bd3144efe24f875ebc8`; cohere `5360753e19e4417675a90e92a1823bea66b13918`; cohere/TypeScript `d92d9bfee114c80be2c375d72edae966176e3a4f`. Both Adamic worktrees were fresh detached worktrees; the report branch was subsequently created on the main worktree. Recursive depth-one submodule initialization and `npm ci` in stage3/api succeeded.

```sh
git fetch origin ba619e28774130e24331c3026a938cd6f3cd6a34 55c375d9d245b580c010c3e0b5659f2849f24fb6
git worktree add --detach "$SCRATCH/rule" ba619e28774130e24331c3026a938cd6f3cd6a34
git worktree add --detach "$SCRATCH/main" 55c375d9d245b580c010c3e0b5659f2849f24fb6
git -C "$SCRATCH/rule" submodule update --init --recursive --depth 1
git -C "$SCRATCH/main" submodule update --init --recursive --depth 1
git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git "$SCRATCH/TypeScript"
git -C "$SCRATCH/TypeScript" rev-parse HEAD
export ADAMIC_TYPESCRIPT_SOURCE="$SCRATCH/TypeScript"
(cd "$SCRATCH/main/stage3/api" && npm ci)
(cd "$SCRATCH/main" && stage3/apply.sh "$SCRATCH/adapted")
# Run the scripts from this report commit, whose API dependencies are installed.
node stage3/census/bind-sites-static/run.cjs "$SCRATCH/adapted" "$SCRATCH/evidence"
```

For the pristine-origin map, run `measure.cjs "$SCRATCH/TypeScript" "$SCRATCH/pristine-evidence"`; the matching uses API-discovered file/member identities and occurrence order, with all 56 origins required to match. It does not use regex to discover sites, types or flow. Python only renders/compresses already-recorded API results and quotes the rule source. No census run or production compiler modification is used to bypass earlier failures.

Validation: stock TypeScript 6.0.3; 79/79 roots present; 79/79 loaded implementation files within compiler; 0 diagnostics; 56/56 sites and pristine origins matched; 0 core-helper callers; 0 bound-name class method declarations; fixed-point supplemental flow terminated. `evidence/SHA256SUMS` hashes the committed compressed API evidence, original-site inventory, apply log and patch-set receipt. The per-source SHA-256s are recorded in the inventory. The applied patch-set total is 118 files, 7,640 added and 6,796 removed lines.
