// Dropping manual memoization: recognising `useMemo`/`useCallback` and recording what the developer
// WROTE before the compiler is allowed to infer anything.
//
// This is React's `dropManualMemoization` (`Inference/DropManualMemoization.ts` at
// `react_conformance.UpstreamSha`) run over this graph. oxc transcribes it at
// `react_compiler_optimization/drop_manual_memoization.rs`; where the two disagree React wins, and
// each disagreement is recorded at the line that resolves it.
//
// # This pass exists because one rule needs an input nothing else produces
//
// `react/preserve-manual-memoization` reports when a developer's hand-written dependency array
// disagrees with the dependencies the compiler infers. The inferred half is `dependencies.go`. This
// file is the DECLARED half, and before it existed `StartMemoize` and `FinishMemoize` were declared
// instruction values that nothing in this tree constructed -- measured at 0 markers across 66,357
// instructions in 677 functions while the same sweep found 455 `ArrayExpression`s, so the probe was
// reading instructions correctly and the zero was the lowering rather than the corpus.
//
// That is why the pass produces MARKERS rather than a diagnostic. The comparison happens later,
// against a value this file must not compute: if this pass inferred the dependencies itself, both
// sides of the rule's comparison would come from the same code and could never disagree.
//
// # What it does to the callback, which is the surprising half
//
// The name says "drop", and the memoization really is dropped: the call to `useMemo` is REWRITTEN
// to call the callback directly, and `useCallback` is rewritten to a plain alias of it.
//
//	before:  $1 = LoadGlobal useMemo      after:  $1 = LoadGlobal useMemo   (now dead)
//	         $2 = Function fn0                    $2 = Function fn0
//	         $3 = Array [$5]                      $3 = Array [$5]           (now dead)
//	         $4 = Call $1($2, $3)                 $4 = Call $2()
//
// For `useCallback` the last line becomes `$4 = LoadLocal $2` instead, because a callback is not
// invoked where it is created. Upstream's comments at `getManualMemoizationReplacement` spell both
// rewrites out and this reproduces them exactly.
//
// The interaction with everything built around it is therefore REAL rather than inert, and it runs
// in the safe direction. No instruction is deleted, no block is created or removed, and no terminal
// is touched, so the control-flow graph is bit-identical afterwards and every analysis keyed on
// block structure -- scope terminals, postdominators, reverse postorder -- sees exactly what it saw
// before. What changes is one instruction VALUE per memo call, and it changes toward less
// structure: a two-argument call becomes a zero-argument call or a load. Operand counts fall, which
// means inferred mutable ranges and scope dependencies can only shrink or stay equal, never grow.
// This was measured rather than argued; see `TestDropManualMemoizationDifferential`.
//
// The dead `LoadGlobal` and `ArrayExpression` are deliberately LEFT behind, matching upstream, which
// relies on later dead-code elimination. Removing them here would renumber instructions under the
// markers this same pass just inserted.
//
// # Where it runs, and the precondition that decides the whole design
//
// Upstream calls this at `Pipeline.ts:169`, and what sits below that line is the specification:
// `enterSSA` is at line 189, twenty lines LATER. This pass runs BEFORE single-assignment form.
// Upstream's own header says so and gives the reason -- it must compose with
// `InlineImmediatelyInvokedFunctionExpressions`, which runs immediately after it and would
// otherwise invalidate the form.
//
// The consequence is not stylistic. Pre-SSA, a `useMemo` global cannot be found by type inference,
// because `inferTypes` is at line 200. So upstream tracks globals and property loads by hand
// through a sidemap, and this file does the same rather than reaching for a type. That is why
// `Recognise` is syntactic: it is not a shortcut, it is the only information available at this
// point in the pipeline.
//
// `DropManualMemoization` therefore takes a function fresh from `Lower` and must be called BEFORE
// `Construct`. That ordering is asserted rather than documented, because `Construct` is not
// idempotent and a caller who ran it first would get markers whose operands name pre-renaming
// identifiers -- an error that produces plausible output. See `dropAfterConstructIsRefused`.
//
// # DIVERGENCE FROM React: optionality is READ here, not reconstructed
//
// React spends the last fifty lines of its file on `findOptionalPlaces`, which walks `optional`
// terminals and their branch arms to recover which places came from `?.`. It needs that because its
// `PropertyLoad` does not carry the flag.
//
// Ours does: `PropertyLoad.Optional` is set at lowering (`lower_expression.go:190`), so
// `props.a?.b` arrives already marked and the reconstruction is unnecessary. Transcribing
// `findOptionalPlaces` here would have been worse than redundant -- it would have been silently
// EMPTY, because it keys on the `Optional` TERMINAL, which `lower.go:78` records as declared and
// never constructed in this tree. A faithful transcription would have returned an empty set at
// every call site, marked every `?.` dependency as non-optional, and passed every test written from
// unguarded fixtures.
//
// That is the exact failure the rule exists to catch, in the pass that feeds it: `props.a` and
// `props?.a` differ by one character at path depth one, and upstream's `areEqualPaths` compares the
// flag. `TestDropManualMemoizationOptionalPathsSurvive` pins it.
//
// # DIVERGENCE FROM React: `React.useMemo` arrives as a MethodCall with a Primitive property
//
// Upstream recognises the namespaced form by seeing `PropertyLoad` off a `LoadGlobal React`, then
// matching a `CallExpression` whose callee is that load. Our lowering emits a single `MethodCall`
// whose `Property` is a `Primitive` holding the string, so the upstream shape never appears and a
// direct transcription would have missed every `React.useMemo` call while handling bare `useMemo`
// correctly -- a partial recognition that looks like a working pass. Both spellings are handled
// here and `TestDropManualMemoizationRecognisesBothSpellings` holds them together.
//
// # The distribution that proves these markers are real
//
// A count of markers cannot separate a working pass from one that emits a marker at every call, so
// the headline is a CROSS-TABULATION of call sites recognised against markers emitted and
// dependencies extracted, over the corpus. The control cell is calls that are NOT memo calls: a
// pass that emitted markers indiscriminately would empty it. See
// `TestDropManualMemoizationDistributionIsReal`.
package high_level_intermediate_representation

// ManualMemoKind is which of the two memoization hooks a call site used.
//
// The two differ in what they memoize -- `useMemo` a value, `useCallback` a function -- and that
// difference decides both the rewrite and which place `FinishMemoize` records, so the distinction
// is carried rather than collapsed to a boolean.
type ManualMemoKind uint8

const (
	// ManualMemoKindNone is a call that is not a memoization hook.
	ManualMemoKindNone ManualMemoKind = iota
	// ManualMemoKindUseMemo is `useMemo` or `React.useMemo`.
	ManualMemoKindUseMemo
	// ManualMemoKindUseCallback is `useCallback` or `React.useCallback`.
	ManualMemoKindUseCallback
)

// String renders the kind as its source spelling.
func (k ManualMemoKind) String() string {
	switch k {
	case ManualMemoKindUseMemo:
		return "useMemo"
	case ManualMemoKindUseCallback:
		return "useCallback"
	}
	return "none"
}

// ManualMemoization is what `DropManualMemoization` did to one function.
//
// Returned rather than only mutating, because the counts are the instrument: a caller checking that
// this pass ran can compare `Recognised` against the call sites it expected, and the gap between
// `Recognised` and `Marked` is exactly the set of calls that were rewritten but whose shape the
// markers could not describe.
type ManualMemoization struct {
	// Recognised is how many `useMemo`/`useCallback` calls were found and rewritten.
	Recognised int
	// Marked is how many of those got a marker pair. Never greater than Recognised.
	Marked int
	// Dependencies is how many written dependency entries were extracted across all markers.
	Dependencies int
	// WithoutDepsArray counts calls that passed no dependency array at all.
	//
	// A distinct count from a zero-length array, because the two are different claims about the
	// code and `StartMemoize.Deps` distinguishes them with nil.
	WithoutDepsArray int
	// UnextractableDeps counts dependency array entries that were not a simple access path.
	//
	// Upstream reports each of these as a diagnostic. Nothing here reports, so they are counted:
	// a corpus where this number is large means the extraction is too narrow, and a rule reading
	// these markers must know some entries went missing rather than assume the list is complete.
	UnextractableDeps int
	// NotAnArrayLiteral counts calls whose second argument was not an array literal.
	NotAnArrayLiteral int
	// NotAnInlineFunction counts calls whose first argument was not an inline function
	// expression, which upstream refuses to mark for a documented reason reproduced below.
	NotAnInlineFunction int
}

// manualMemoSidemap is what a single forward walk learns about the temporaries in a function.
//
// Upstream calls this `IdentifierSidemap` and builds it in the SAME loop that does the rewriting,
// relying on a temporary always being defined before its use. That holds pre-SSA in a graph built
// from an expression tree, and it is relied on here for the same reason: a value's defining
// instruction precedes every use of it in evaluation order.
type manualMemoSidemap struct {
	// functions maps a temporary to the function expression stored in it.
	functions map[IdentifierId]bool
	// manualMemos maps a temporary holding a loaded `useMemo`/`useCallback` to which one it is.
	manualMemos map[IdentifierId]ManualMemoKind
	// react holds temporaries holding the `React` namespace object.
	react map[IdentifierId]bool
	// depsLists maps a temporary holding an array literal to its element places.
	depsLists map[IdentifierId][]Place
	// deps maps a temporary to the access path it evaluates to, when it is one.
	deps map[IdentifierId]ManualMemoDependency
	// optionals is upstream's `findOptionalPlaces`: property values guarded by an Optional
	// terminal. Structural optional lowering deliberately leaves PropertyLoad.Optional false.
	optionals map[IdentifierId]bool
	// optionalJoins are the only phi blocks through which a single resolved optional arm may be
	// projected. An enclosing ternary or logical join is a distinct expression.
	optionalJoins map[BlockId]bool
	// anchors maps a temporary to the instruction that defined it.
	//
	// Upstream keeps the whole defining instruction on its `ManualMemoCallee`; only the id is
	// needed here, and keeping it for every temporary rather than only callee loads costs one map
	// write per instruction and removes a special case.
	anchors map[IdentifierId]InstructionId
}

// DropManualMemoization rewrites `useMemo`/`useCallback` calls and records what was written.
//
// Must run on a function fresh from `Lower`, BEFORE `Construct`. See the package note above.
func DropManualMemoization(function *Function) ManualMemoization {
	result := ManualMemoization{}
	sidemap := manualMemoSidemap{
		functions:     map[IdentifierId]bool{},
		manualMemos:   map[IdentifierId]ManualMemoKind{},
		react:         map[IdentifierId]bool{},
		depsLists:     map[IdentifierId][]Place{},
		deps:          map[IdentifierId]ManualMemoDependency{},
		optionals:     findManualMemoOptionalPlaces(function),
		optionalJoins: optionalChainJoinBlocks(function),
		anchors:       map[IdentifierId]InstructionId{},
	}

	// An optional chain in the dependency ARRAY spans blocks rather than being one `PropertyLoad`.
	//
	// `collectManualMemoTemporaries` reads `PropertyLoad.Optional` off the instruction, and the
	// optional lowering leaves that false by construction: optionality lives on the `Optional`
	// terminal, which is upstream's shape. So `props?.items` never lands in `sidemap.deps`,
	// `extractManualMemoArguments` abandons the whole list on its first missing element, and `Deps`
	// stays nil -- which disables the comparison entirely.
	//
	// Measured before this: seven goldens, every one an optional-chain fixture, inferred their
	// dependencies correctly and reported nothing, because there was no source list to compare them
	// against. Upstream needs no equivalent, since its `collectMaybeMemoDependencies` reads the
	// optional form while walking the array before SSA creates join phis.
	if chains := CollectOptionalChainSidemap(function); chains != nil {
		recordOptionalChainJoinPhis(function, chains)
		for id, dependency := range chains.TemporariesReadInOptional {
			sidemap.deps[id] = ManualMemoDependency{
				Root: ManualMemoRoot{Place: Place{
					Identifier: dependency.Identifier,
					Reactive:   dependency.Reactive,
				}},
				Path: dependency.Path,
			}
		}
	}

	// Insertions are QUEUED against the instruction they follow rather than applied during the
	// walk, exactly as upstream does, because a block's instruction list is being iterated by
	// index and splicing into it mid-walk would either skip an instruction or revisit one.
	var queued []queuedMarkerInsert

	for _, block := range function.Blocks {
		collectManualMemoPhiDependencies(block, &sidemap)
		for _, instructionId := range block.Instructions {
			instruction := function.Instructions[instructionId]
			kind, calleeLoad := recogniseManualMemoCall(instruction, &sidemap)
			if kind == ManualMemoKindNone {
				collectManualMemoTemporaries(function, instruction, &sidemap)
				continue
			}

			extracted, ok := extractManualMemoArguments(instruction, kind, &sidemap, &result)
			if !ok {
				// Upstream `continue`s here WITHOUT rewriting, leaving the call intact. That is
				// reproduced: a call whose callback argument is missing is not a memoization this
				// pass understands, and rewriting it would change behaviour on malformed input.
				continue
			}
			result.Recognised++

			// The rewrite. `useMemo(fn, deps)` becomes `fn()`; `useCallback(fn, deps)` becomes a
			// load of `fn`. The `LoadGlobal` and the array are left in place as dead code, which
			// is upstream's behaviour and is what keeps instruction ids stable here.
			if kind == ManualMemoKindUseMemo {
				instruction.Value = &CallExpression{Callee: extracted.callback}
			} else {
				instruction.Value = &LoadLocal{Place: extracted.callback}
			}

			if !extracted.inlineFunction {
				// Upstream refuses to mark `useMemo(opaqueFn, [a, b])`, where the callback is not
				// an inline function expression, and records a diagnostic. Its reason is that the
				// validation downstream assumes source deps closely match inferred deps because
				// `exhaustive-deps` only lints inline callbacks; without that assumption the
				// comparison produces noise. The rewrite still happens, the marking does not.
				result.NotAnInlineFunction++
				continue
			}

			memoId := result.Marked
			result.Marked++

			// `useMemo` memoizes the call's RESULT; `useCallback` memoizes the callback itself.
			// Upstream picks between `instr.lvalue` and the callback place on exactly this basis,
			// and getting it backwards would make every `useCallback` marker name a temporary the
			// rule cannot relate to anything.
			declaration := instruction.LValue
			if kind == ManualMemoKindUseCallback {
				declaration = extracted.callback
			}

			start := newMarkerInstruction(function, instruction, &StartMemoize{
				ManualMemoId: memoId,
				Deps:         extracted.dependencies,
			})
			finish := newMarkerInstruction(function, instruction, &FinishMemoize{
				ManualMemoId: memoId,
				Value:        declaration,
			})

			// `StartMemoize` goes after the CALLEE LOAD, not after the call, so that every
			// temporary produced while lowering the inline callback falls between the two markers.
			// Upstream documents this placement with a worked example and it is the reason the
			// callee load instruction is tracked at all. Placing it just before the call instead
			// would exclude the callback's own captures from the region, and a consumer scanning
			// between the markers would find a region missing precisely the values the memo reads.
			queued = append(queued, queuedMarkerInsert{after: calleeLoad, instruction: start})
			queued = append(queued, queuedMarkerInsert{after: instructionId, instruction: finish})
			result.Dependencies += len(extracted.dependencies)
			if extracted.dependencies == nil {
				result.WithoutDepsArray++
			}
		}
	}

	if len(queued) > 0 {
		insertQueuedMarkers(function, queued)
	}
	return result
}

// queuedMarkerInsert is one marker waiting to be spliced in after a given instruction.
type queuedMarkerInsert struct {
	after       InstructionId
	instruction *Instruction
}

// insertQueuedMarkers splices the marker instructions into their blocks.
//
// Separated from the walk for the reason upstream separates its two phases, and it rebuilds each
// affected block's instruction list rather than inserting in place so the cost is linear in the
// block rather than quadratic in the number of insertions.
func insertQueuedMarkers(function *Function, queued []queuedMarkerInsert) {
	byAnchor := make(map[InstructionId][]*Instruction, len(queued))
	for _, insert := range queued {
		byAnchor[insert.after] = append(byAnchor[insert.after], insert.instruction)
	}

	for _, block := range function.Blocks {
		var rebuilt []InstructionId
		for index, instructionId := range block.Instructions {
			followers, ok := byAnchor[instructionId]
			if !ok {
				if rebuilt != nil {
					rebuilt = append(rebuilt, instructionId)
				}
				continue
			}
			if rebuilt == nil {
				rebuilt = append(rebuilt, block.Instructions[:index]...)
			}
			rebuilt = append(rebuilt, instructionId)
			for _, follower := range followers {
				id := InstructionId(len(function.Instructions))
				follower.Id = id
				function.Instructions = append(function.Instructions, follower)
				rebuilt = append(rebuilt, id)
			}
		}
		if rebuilt != nil {
			block.Instructions = rebuilt
		}
	}

	// Evaluation order is now stale for every instruction after the first insertion. Upstream
	// calls `markInstructionIds` for the same reason; here the numbering lives in `graph.go` and
	// renumbering it is the whole fix. Skipping this leaves the markers at order zero, which
	// `MarkEvaluationOrder` documents as meaning "the finalizer did not reach this block".
	MarkEvaluationOrder(function)
}

// newMarkerInstruction builds one marker instruction with a fresh temporary for its result.
//
// The marker's own lvalue is a temporary nothing reads, which is upstream's `createTemporaryPlace`.
// It exists because every instruction in this IR has an lvalue; reusing the call's lvalue would
// give two instructions the same destination and break single-assignment construction downstream.
func newMarkerInstruction(function *Function, anchor *Instruction, value InstructionValue) *Instruction {
	identifier := function.NewIdentifier("", nil, 0)
	return &Instruction{
		LValue: Place{Identifier: identifier.Id, Range: anchor.Range},
		Value:  value,
		Node:   anchor.Node,
		Range:  anchor.Range,
	}
}

// recogniseManualMemoCall reports whether an instruction calls `useMemo`/`useCallback`.
//
// Returns the kind and the instruction that LOADED the callee, which is where `StartMemoize` is
// anchored. Both call spellings are handled; see the namespaced-form divergence in the package note.
func recogniseManualMemoCall(instruction *Instruction, sidemap *manualMemoSidemap) (ManualMemoKind, InstructionId) {
	switch value := instruction.Value.(type) {
	case *CallExpression:
		return sidemap.manualMemos[value.Callee.Identifier], sidemap.calleeAnchor(value.Callee.Identifier)
	case *MethodCall:
		// `React.useMemo(...)`. The receiver must be the React namespace and the property must be
		// the literal name; the property arrives as a `Primitive` string rather than a
		// `PropertyLoad`, which is the divergence recorded above.
		if !sidemap.react[value.Receiver.Identifier] {
			return ManualMemoKindNone, 0
		}
		return sidemap.manualMemos[value.Property.Identifier], sidemap.calleeAnchor(value.Property.Identifier)
	}
	return ManualMemoKindNone, 0
}

// calleeAnchor is the instruction that defined a temporary, for anchoring `StartMemoize`.
func (s *manualMemoSidemap) calleeAnchor(identifier IdentifierId) InstructionId {
	return s.anchors[identifier]
}

// extractedMemoArguments is what one recognised call site yielded.
type extractedMemoArguments struct {
	callback Place
	// inlineFunction reports that the callback argument was an inline function expression.
	inlineFunction bool
	// dependencies is the written dependency list, nil when no array was passed.
	dependencies []ManualMemoDependency
}

// extractManualMemoArguments reads the callback and the dependency array off a recognised call.
//
// Returns false when there is no callback argument at all, which upstream treats as an error and
// refuses to rewrite.
func extractManualMemoArguments(
	instruction *Instruction,
	kind ManualMemoKind,
	sidemap *manualMemoSidemap,
	result *ManualMemoization,
) (extractedMemoArguments, bool) {
	var arguments []Argument
	switch value := instruction.Value.(type) {
	case *CallExpression:
		arguments = value.Args
	case *MethodCall:
		arguments = value.Args
	}
	if len(arguments) == 0 || arguments[0].Spread {
		return extractedMemoArguments{}, false
	}

	extracted := extractedMemoArguments{
		callback:       arguments[0].Place,
		inlineFunction: sidemap.functions[arguments[0].Place.Identifier],
	}
	if len(arguments) < 2 || arguments[1].Spread {
		// No dependency array. `Deps` stays nil, which is a different fact from an empty array.
		return extracted, true
	}

	elements, ok := sidemap.depsLists[arguments[1].Place.Identifier]
	if !ok {
		// `useMemo(fn, someVariable)`. Upstream reports this and returns null, refusing to rewrite
		// the call at all. Reproduced: the caller treats false as "leave it alone".
		result.NotAnArrayLiteral++
		return extractedMemoArguments{}, false
	}

	// Non-nil even when empty: `useMemo(fn, [])` claims the value never changes, and that claim is
	// exactly what a nil `Deps` would erase.
	dependencies := make([]ManualMemoDependency, 0, len(elements))
	for _, element := range elements {
		dependency, ok := sidemap.deps[element.Identifier]
		if !ok {
			// An entry that is not a simple access path, such as `[a + b]` or `[x[0]]`.
			//
			// Upstream calls `env.recordError` here (`DropManualMemoization.ts:356`), which is a
			// COMPILE ERROR: the function bails and `ValidatePreservedManualMemoization` never runs
			// on it. Counting the entry and continuing turns "this list could not be read" into
			// "this list was empty", and those are different claims -- an empty list means the
			// developer promised no dependencies, so every inferred one is reported as a
			// disagreement they never made.
			//
			// A nil `Deps` is the closest faithful equivalent available here, and this file already
			// documents it as the signal that disables the comparison: "no dependency array. `Deps`
			// stays nil, which is a different fact from an empty array". So the whole list is
			// abandoned rather than the entry silently removed from it.
			//
			// Measured on `useMemo-dep-array-literal-access.ts`, whose source is `[x[0]]` and whose
			// own comment says upstream only recognises "hoistable" values in a deps list: we
			// extracted zero of one entry and handed the validator an empty list, which then
			// reported `props` against nothing. Upstream compiles it with no error at all.
			result.UnextractableDeps++
			return extractedMemoArguments{
				callback:       extracted.callback,
				inlineFunction: extracted.inlineFunction,
			}, true
		}
		dependencies = append(dependencies, dependency)
	}
	extracted.dependencies = dependencies
	return extracted, true
}

// collectManualMemoTemporaries records what a non-call instruction stores into its temporary.
//
// This is upstream's `collectTemporaries` merged with `collectMaybeMemoDependencies`. The two are
// separate functions there because the second is also called from the validation pass; here there
// is one caller and splitting would only add a hop.
func collectManualMemoTemporaries(function *Function, instruction *Instruction, sidemap *manualMemoSidemap) {
	target := instruction.LValue.Identifier
	sidemap.anchors[target] = instruction.Id

	switch value := instruction.Value.(type) {
	case *FunctionExpression:
		sidemap.functions[target] = true

	case *LoadGlobal:
		switch value.Name {
		case "useMemo":
			sidemap.manualMemos[target] = ManualMemoKindUseMemo
		case "useCallback":
			sidemap.manualMemos[target] = ManualMemoKindUseCallback
		case "React":
			sidemap.react[target] = true
		default:
			// A global that is not a hook is still a valid dependency ROOT: `[SOME_CONSTANT]` is
			// a legal dependency array entry, and dropping it would silently shorten the list.
			sidemap.deps[target] = ManualMemoDependency{
				Root: ManualMemoRoot{IsGlobal: true, Name: value.Name},
			}
		}

	case *Primitive:
		// `React.useMemo` lowers the property name to a Primitive string; see the divergence note.
		if name, ok := value.Value.(string); ok {
			switch name {
			case "useMemo":
				sidemap.manualMemos[target] = ManualMemoKindUseMemo
			case "useCallback":
				sidemap.manualMemos[target] = ManualMemoKindUseCallback
			}
		}

	case *ArrayExpression:
		// Only an array of plain elements is a dependency list. A spread or a hole means the
		// contents are not statically the written dependencies, and upstream's `every` check
		// rejects the same shapes.
		places := make([]Place, 0, len(value.Elements))
		for _, element := range value.Elements {
			if element.Spread || element.Hole {
				return
			}
			places = append(places, element.Place)
		}
		sidemap.depsLists[target] = places

	case *PropertyLoad:
		// Extending an existing path by one step. New structural chains record optionality on their
		// terminal, while computed accesses and older call shapes may still carry the instruction
		// flag. Reading both preserves the source spelling across that transition.
		object, ok := sidemap.deps[value.Object.Identifier]
		if !ok {
			return
		}
		path := make([]DependencyPathEntry, 0, len(object.Path)+1)
		path = append(path, object.Path...)
		path = append(path, DependencyPathEntry{
			Property: value.Property,
			Optional: value.Optional || sidemap.optionals[target],
		})
		sidemap.deps[target] = ManualMemoDependency{Root: object.Root, Path: path}

	case *LoadLocal:
		propagateManualMemoDependency(function, sidemap, target, value.Place)

	case *LoadContext:
		propagateManualMemoDependency(function, sidemap, target, value.Place)

	case *StoreLocal:
		// A value block's result arrives through a store. Upstream tracks these so that optional
		// chains, which lower through value blocks there, stay reachable as dependency roots.
		//
		// Only into an UNNAMED target. Upstream's condition is `aliased != null &&
		// lvalue.name?.kind !== 'named'` (`DropManualMemoization.ts:116`), so a store into a named
		// binding ends the chain rather than continuing it: the binding is the root a developer can
		// write, and resolving through it would name something they did not.
		//
		// Measured on `useCallback-alias-property-load-dep.ts`, where `const x = propB.x.y` is
		// written as `[propA.x, x]`. Without the guard the written `x` resolved to `propB.x.y`
		// while the inferred side kept `x`, so the two never matched and the rule fired on a
		// fixture upstream compiles cleanly.
		if source, ok := sidemap.deps[value.Value.Identifier]; ok && !isNamedBinding(function, value.LValue.Identifier) {
			sidemap.deps[value.LValue.Identifier] = source
		}
	}
}

// findManualMemoOptionalPlaces recovers the guarded value of each optional source link.
//
// This is DropManualMemoization.ts's `findOptionalPlaces`, kept separate from
// CollectOptionalChainSidemap deliberately. The latter answers which chains are safe inferred
// dependencies and therefore accepts only local/context roots. Source dependency extraction has a
// different contract: `[GLOBAL?.field]` is a valid written dependency even though a LoadGlobal is
// not a hoistable inferred root.
func findManualMemoOptionalPlaces(function *Function) map[IdentifierId]bool {
	result := map[IdentifierId]bool{}
	if function == nil {
		return result
	}
	for _, block := range function.Blocks {
		optional, ok := block.Terminal.(*Optional)
		if !ok || !optional.Optional {
			continue
		}
		test, found := function.Block(optional.Test)
		seen := map[BlockId]bool{}
		for found && test != nil && !seen[test.Id] {
			seen[test.Id] = true
			var next BlockId
			switch terminal := test.Terminal.(type) {
			case *Branch:
				if terminal.Fallthrough == optional.Fallthrough {
					consequent, exists := function.Block(terminal.Consequent)
					if exists && consequent != nil && len(consequent.Instructions) > 0 {
						last := function.Instructions[consequent.Instructions[len(consequent.Instructions)-1]]
						if last != nil {
							if store, isStore := last.Value.(*StoreLocal); isStore {
								result[store.Value.Identifier] = true
							}
						}
					}
					found = false
					continue
				}
				next = terminal.Fallthrough
			case *Optional:
				next = terminal.Fallthrough
			case *Logical:
				next = terminal.Fallthrough
			case *Ternary:
				next = terminal.Fallthrough
			case *Sequence:
				next = terminal.Fallthrough
			case *MaybeThrow:
				next = terminal.Continuation
			default:
				found = false
				continue
			}
			test, found = function.Block(next)
		}
	}
	return result
}

// collectManualMemoPhiDependencies carries a simple source path across SSA joins.
//
// Upstream drops manual memoization before SSA, where StoreLocal writes the value-block result
// directly. This pipeline runs after construction, so the dependency array names the join phi
// instead. A single resolved arm is sufficient for an optional join because the other arm is the
// undefined short circuit; conflicting resolved arms are a real branch and are not collapsed.
func collectManualMemoPhiDependencies(block *BasicBlock, sidemap *manualMemoSidemap) {
	if block == nil || sidemap == nil || !sidemap.optionalJoins[block.Id] {
		return
	}
	for _, phi := range block.Phis {
		if _, known := sidemap.deps[phi.Place.Identifier]; known {
			continue
		}
		var resolved *ManualMemoDependency
		conflict := false
		for _, predecessor := range PhiOperandsInOrder(phi) {
			dependency, ok := sidemap.deps[phi.Operands[predecessor].Identifier]
			if !ok {
				continue
			}
			if resolved == nil {
				copy := dependency
				resolved = &copy
				continue
			}
			if !manualMemoRootsEqual(resolved.Root, dependency.Root) ||
				!equalPaths(resolved.Path, dependency.Path) {
				conflict = true
				break
			}
		}
		if resolved != nil && !conflict {
			sidemap.deps[phi.Place.Identifier] = *resolved
		}
	}
}

// isNamedBinding reports whether a value is a named source binding rather than a temporary.
func isNamedBinding(function *Function, id IdentifierId) bool {
	if function == nil || int(id) >= len(function.Identifiers) {
		return false
	}
	identifier := function.Identifiers[id]
	return identifier != nil && identifier.Name != ""
}

// propagateManualMemoDependency carries a path through a load, or starts one at a named binding.
//
// A load of a temporary that already has a path continues it. A load of a NAMED binding starts a
// fresh path rooted there, which is what makes `props` in `[props.items]` a root at all. An unnamed
// temporary with no path is neither and is left absent, which is what makes `[a + b]` unextractable
// rather than silently rooted at a meaningless place.
func propagateManualMemoDependency(function *Function, sidemap *manualMemoSidemap, target IdentifierId, source Place) {
	if existing, ok := sidemap.deps[source.Identifier]; ok {
		sidemap.deps[target] = existing
		return
	}
	if int(source.Identifier) >= len(function.Identifiers) {
		return
	}
	if function.Identifiers[source.Identifier].Name == "" {
		return
	}
	sidemap.deps[target] = ManualMemoDependency{Root: ManualMemoRoot{Place: source}}
}
