// Removing instructions nothing reads, which is React's `deadCodeElimination`.
//
// (`Optimization/DeadCodeElimination.ts` at `react_conformance.UpstreamSha`.)
//
// # Why the absence of this pass is visible three ways
//
// Upstream runs it at `Pipeline.ts:230`, after `DropManualMemoization` at 168 and 198 lines before
// `propagateScopeDependenciesHIR` at 428. Everything the memo rewrite leaves behind is gone before
// any dependency analysis looks. Without it:
//
//   - the loads that built a written dependency array seed the non-null set, so a path the developer
//     merely declared licenses the walk to infer something deeper. Worked around at the hoistable
//     seed rather than fixed.
//   - the `ArrayExpression` itself is visited, and its bare roots mark the dependency tree's root,
//     which prunes every deeper path beneath it: `useMemo(() => [x.y.z], [x])` infers `x`.
//   - whole function bodies survive that upstream deletes. `prune-nonescaping-useMemo.ts` compiles
//     to `function useFoo(x) {}` there, so nothing remains to validate; here `x` keeps a scope and
//     the rule reports on it.
//
// # What is preserved, and why the list is not a judgement call
//
// `pruneableValue` is transcribed rather than reasoned about. The load-bearing entry is that
// `StartMemoize` and `FinishMemoize` are never pruneable -- upstream's comment is "we can't DCE
// without losing the memoization guarantees" -- while `PropertyLoad`, `ArrayExpression` and
// `LoadGlobal` are. That asymmetry is the whole point: what the developer declared survives in the
// marker, and the instructions that built it do not.
package high_level_intermediate_representation

import "github.com/system-inc/cohere/static_single_assignment"

// DeadCodeEliminationResult reports what the pass removed, for measurement.
type DeadCodeEliminationResult struct {
	// Instructions is how many instructions were swept.
	Instructions int
	// Phis is how many phis were swept.
	Phis int
}

// EliminateDeadCode removes instructions and phis whose results nothing reads.
//
// Two phases, upstream's. Liveness runs to a fixpoint first because a use may be visited after its
// declaration when phis carry data around a loop, so sweeping cannot be interleaved with marking.
func EliminateDeadCode(function *Function) DeadCodeEliminationResult {
	result := DeadCodeEliminationResult{}
	if function == nil {
		return result
	}
	state := findReferencedIdentifiers(function)

	for _, block := range function.Blocks {
		if block == nil {
			continue
		}
		retained := block.Instructions[:0]
		for _, instructionId := range block.Instructions {
			instruction := function.Instructions[instructionId]
			if instruction == nil {
				continue
			}
			if state.usedByIdOrName(function, instruction.LValue.Identifier) {
				retained = append(retained, instructionId)
				continue
			}
			result.Instructions++
		}
		block.Instructions = retained

		kept := block.Phis[:0]
		for _, phi := range block.Phis {
			if state.usedByIdOrName(function, phi.Place.Identifier) {
				kept = append(kept, phi)
				continue
			}
			result.Phis++
		}
		block.Phis = kept
	}
	return result
}

// liveIdentifiers is upstream's `State`: the set of values read, plus the names they carry.
//
// Two sets rather than one because a named binding is read by NAME across its versions. If any
// version of `x` is read, every version must be kept, since sweeping one would leave a later read
// resolving to nothing.
type liveIdentifiers struct {
	identifiers map[static_single_assignment.IdentifierId]bool
	named       map[string]bool
}

// reference marks a value as read.
func (l *liveIdentifiers) reference(function *Function, id static_single_assignment.IdentifierId) {
	l.identifiers[id] = true
	if identifier := function.Identifiers[id]; identifier != nil && identifier.Name != "" {
		l.named[identifier.Name] = true
	}
}

// usedByIdOrName reports whether this value, or any version of its name, is read.
//
// # A mutation making this ignore the name set SURVIVES
//
// Measured: id-only liveness scores identically -- goldens 25, ours 120, exact 20. No corpus fixture
// currently has one version of a named binding read while another is swept, because the sweep runs
// immediately after `DropManualMemoization` and the instructions it removes are freshly lowered
// temporaries with no name at all.
//
// Kept because it is the guard against the one unrecoverable failure this pass has: sweeping a
// version whose name is read later leaves that read resolving to nothing, which is a miscompile
// rather than a lost optimisation. The verdict EXPIRES the moment this runs anywhere later in the
// pipeline, where named reassignments are live across versions.
func (l *liveIdentifiers) usedByIdOrName(function *Function, id static_single_assignment.IdentifierId) bool {
	if l.identifiers[id] {
		return true
	}
	identifier := function.Identifiers[id]
	return identifier != nil && identifier.Name != "" && l.named[identifier.Name]
}

// usedById reports whether this exact value is read, ignoring its name.
func (l *liveIdentifiers) usedById(id static_single_assignment.IdentifierId) bool {
	return l.identifiers[id]
}

// findReferencedIdentifiers marks every value some surviving instruction reads.
//
// Blocks are walked in reverse and instructions within them in reverse, so a use is seen before its
// declaration on straight-line code and the whole thing converges in one pass. A loop needs more,
// which is what the outer fixpoint is for: `hasBackEdge` decides whether to run it at all.
func findReferencedIdentifiers(function *Function) *liveIdentifiers {
	state := &liveIdentifiers{
		identifiers: map[static_single_assignment.IdentifierId]bool{},
		named:       map[string]bool{},
	}
	// A mutation forcing this false SURVIVES: measured identical, because no corpus function
	// reaches the sweep with a loop whose liveness needs a second pass. It decides only how many
	// times the loop below runs, so a wrong `false` costs an answer and a wrong `true` costs an
	// iteration -- which is why it is computed rather than assumed either way.
	loops := functionHasBackEdge(function)

	for {
		size := len(state.identifiers)
		for index := len(function.Blocks) - 1; index >= 0; index-- {
			block := function.Blocks[index]
			if block == nil {
				continue
			}
			EachTerminalPlace(block.Terminal, func(place Place, role PlaceRole) {
				if role != PlaceRoleDefine {
					state.reference(function, place.Identifier)
				}
			})

			for position := len(block.Instructions) - 1; position >= 0; position-- {
				instruction := function.Instructions[block.Instructions[position]]
				if instruction == nil {
					continue
				}
				// # A mutation removing this pessimism SURVIVES
				//
				// Measured identical on every instrument. No fixture in this corpus reaches the
				// sweep with a value block whose final instruction would otherwise be pruned,
				// because the memo rewrite produces ordinary blocks. It is upstream's own guard and
				// protects a structural invariant rather than a score: a value block's last
				// instruction IS the block's value, so removing it leaves a block that produces
				// nothing and a consumer reading an absent operand. The verdict EXPIRES if this
				// pass ever runs after ternaries and logical operators are lowered into value
				// blocks that survive to this point.
				//
				// The last instruction of a value block IS that block's value, so it is never
				// eligible for pruning and its operands are all pessimistically live. Upstream:
				// "Pessimistically consider all operands as used to avoid rewriting the last
				// instruction".
				isBlockValue := block.Kind != BlockKindBlock &&
					position == len(block.Instructions)-1
				if isBlockValue {
					state.reference(function, instruction.LValue.Identifier)
					EachPlace(instruction.Value, func(place Place, role PlaceRole) {
						if role != PlaceRoleDefine {
							state.reference(function, place.Identifier)
						}
					})
					continue
				}
				if !state.usedByIdOrName(function, instruction.LValue.Identifier) &&
					pruneableValue(function, instruction.Value, state) {
					continue
				}
				state.reference(function, instruction.LValue.Identifier)
				if store, isStore := instruction.Value.(*StoreLocal); isStore {
					// A declaration's initializer is live only if the declared value itself is.
					// A reassignment always reads its rvalue.
					//
					// A mutation making both arms live SURVIVES on this corpus: nothing here
					// declares a binding whose initializer is otherwise dead. Upstream's asymmetry
					// is kept because it is the difference between `const x = f()` -- where dropping
					// the call would lose an effect -- and a temporary store, and the arms are not
					// interchangeable even where the corpus cannot tell them apart.
					if store.Kind == InstructionKindReassign ||
						state.usedById(store.LValue.Identifier) {
						state.reference(function, store.Value.Identifier)
					}
					continue
				}
				EachPlace(instruction.Value, func(place Place, role PlaceRole) {
					if role != PlaceRoleDefine {
						state.reference(function, place.Identifier)
					}
				})
			}

			for _, phi := range block.Phis {
				if !state.usedByIdOrName(function, phi.Place.Identifier) {
					continue
				}
				// Read in ascending predecessor order, which is the order `Phi.Operands` keeps, so
				// the same function behaves the same way between runs.
				for _, predecessor := range PhiOperandsInOrder(phi) {
					state.reference(function, phi.Operands.At(predecessor).Identifier)
				}
			}
		}
		if len(state.identifiers) <= size || !loops {
			return state
		}
	}
}

// functionHasBackEdge reports whether any block reaches a predecessor that is not before it.
//
// Upstream's `hasBackEdge`. Only used to decide whether liveness needs more than one pass, so a
// conservative true costs an iteration and never an answer.
func functionHasBackEdge(function *Function) bool {
	position := make(map[static_single_assignment.BlockId]int, len(function.Blocks))
	for index, block := range function.Blocks {
		if block != nil {
			position[block.Id] = index
		}
	}
	for index, block := range function.Blocks {
		if block == nil {
			continue
		}
		for _, predecessor := range block.Predecessors {
			if at, known := position[predecessor]; known && at >= index {
				return true
			}
		}
	}
	return false
}

// pruneableValue reports whether an instruction may be removed when its result is unread.
//
// Transcribed from upstream's switch of the same name. The default is NOT pruneable: an instruction
// kind this does not name is kept, so adding a value kind cannot silently start deleting it.
func pruneableValue(function *Function, value InstructionValue, state *liveIdentifiers) bool {
	switch shape := value.(type) {
	case *DeclareLocal:
		return !state.usedByIdOrName(function, shape.LValue.Identifier)

	case *StoreLocal:
		if shape.Kind == InstructionKindReassign {
			// A reassignment goes only if this particular version is unread; a declaration must
			// consider every version of the name.
			return !state.usedById(shape.LValue.Identifier)
		}
		return !state.usedByIdOrName(function, shape.LValue.Identifier)

	case *StartMemoize, *FinishMemoize:
		// Upstream: "This instruction is used by the @enablePreserveExistingMemoizationGuarantees
		// feature to preserve information about memoization semantics in the original code. We can't
		// DCE without losing the memoization guarantees."
		return false

	case *LoadContext, *DeclareContext, *StoreContext:
		return false

	case *Await, *ComputedDelete, *ComputedStore, *PropertyDelete, *PropertyStore, *StoreGlobal:
		// Side effects, so unread does not mean unnecessary.
		return false

	case *CallExpression, *MethodCall, *NewExpression, *TaggedTemplateExpression:
		return false

	case *GetIterator, *NextPropertyOf, *IteratorNext:
		return false

	case *Debugger:
		return false

	case *RegExpLiteral, *MetaProperty, *LoadGlobal, *ArrayExpression, *BinaryExpression,
		*ComputedLoad, *FunctionExpression, *LoadLocal, *JsxExpression, *JsxFragment, *JsxText,
		*ObjectExpression, *Primitive, *PropertyLoad, *TemplateLiteral, *UnaryExpression,
		*TypeCastExpression:
		// Pure by construction: no observable effect beyond the value produced.
		return true

	default:
		return false
	}
}
