// Copying a nested function's body into its parent's id space.
//
// This is the half of `inlineImmediatelyInvokedFunctionExpressions` that upstream does not have to
// write. React mints block ids, identifier ids and instruction ids from one environment counter, so
// its inliner drops a nested function's blocks straight into the parent's map:
//
//	for (const [id, block] of body.loweredFunc.func.body.blocks) {
//	  block.preds.clear();
//	  fn.body.blocks.set(id, block);
//	}
//
// This tree cannot. Ids here are per-function and overlap between a function and its children, which
// `effects.go:573-577` records as a deliberate property: "InstructionIds are per-function and overlap
// between a function and its children ... the two spaces are kept apart here by construction." A
// nested body's tables all start at zero, so the same move would alias the parent's own values.
//
// So the copy renames as it goes, and this file is only the rename. The control-flow surgery that
// consumes it is upstream's and transcribes directly once ids are safe.
package high_level_intermediate_representation

import "reflect"

// InlineRemap is the correspondence between a nested function's ids and the parent's.
//
// Returned rather than applied, so a caller can rewrite the terminals it owns -- the call site's
// continuation, the goto that replaces a return -- against the same mapping the body was copied
// under. A caller that guessed at these would produce a graph that builds and points at the wrong
// values.
type InlineRemap struct {
	// Identifiers maps a nested identifier id to the parent's.
	Identifiers map[IdentifierId]IdentifierId
	// Blocks maps a nested block id to the parent's.
	Blocks map[BlockId]BlockId
	// Instructions maps a nested instruction id to the parent's.
	Instructions map[InstructionId]InstructionId
	// Entry is the parent-space id of the nested function's entry block.
	Entry BlockId
}

// CopyNestedBodyInto copies a nested function's blocks, instructions and identifiers into parent.
//
// The nested function is left untouched: every structure is copied rather than moved, because
// `Function.Functions` still holds it and a later pass reading it must not see a half-renamed body.
// The cost is one allocation per instruction and per block, bounded by the body being inlined.
//
// # What is renamed, and the enumeration is the point
//
// Four id spaces cross the boundary and missing any one produces a graph that builds:
//
//   - `IdentifierId`, on every `Place` in every instruction value, lvalue, phi and terminal.
//   - `BlockId`, on every terminal reference and every phi operand key.
//   - `InstructionId`, on every block's instruction list.
//   - `FunctionId`, on a `FunctionExpression` or `ObjectMethod` nested inside the body being copied.
//
// The fourth is the one a reader does not expect. A closure inside the inlined closure indexes
// `Function.Functions` of the nested function, and that slice is not the parent's.
//
// # Captures become the mapping's seed rather than new values
//
// A nested function's `Context` names the values it closed over, and `FunctionExpression.Captures`
// names the same values in the parent. Index i of the two is one binding seen from both sides, so
// the mapping starts by pairing them: a capture must resolve to the parent's existing value, not to
// a fresh copy of it, or the inlined body reads a different variable than the code around it.
func CopyNestedBodyInto(parent *Function, nested *Function, captures []Place) (*InlineRemap, bool) {
	if parent == nil || nested == nil {
		return nil, false
	}
	if len(nested.Context) != len(captures) {
		// Upstream pairs these by index and asserts the lengths match. Declining is the honest
		// answer here: a mismatch means the capture list and the context disagree about what was
		// closed over, and inlining under that disagreement rewires reads silently.
		return nil, false
	}

	remap := &InlineRemap{
		Identifiers:  map[IdentifierId]IdentifierId{},
		Blocks:       map[BlockId]BlockId{},
		Instructions: map[InstructionId]InstructionId{},
	}

	// Seed: a captured value keeps the parent's identifier. Seed the declaration map at the same
	// time. Construction has already put the nested body in SSA form, so other nested identifiers
	// may be distinct values of this same source binding; those need fresh parent identifiers but
	// must remain in the capture's declaration equivalence class.
	declarations := map[DeclarationId]DeclarationId{}
	for index, contextValue := range nested.Context {
		capture := captures[index]
		remap.Identifiers[contextValue.Identifier] = capture.Identifier
		contextIdentifier := nested.IdentifierOf(contextValue)
		captureIdentifier := parent.IdentifierOf(capture)
		if contextIdentifier == nil || captureIdentifier == nil {
			return nil, false
		}
		if mapped, exists := declarations[contextIdentifier.Declaration]; exists &&
			mapped != captureIdentifier.Declaration {
			return nil, false
		}
		declarations[contextIdentifier.Declaration] = captureIdentifier.Declaration
	}

	// Identifiers. Nested declaration ids cannot be copied verbatim because the two functions mint
	// them independently, but their equivalence classes are semantic: every SSA version of one
	// source binding must still share one DeclarationId after the copy. Remap each class to one fresh
	// parent declaration (or to the seeded capture declaration) while keeping IdentifierIds distinct.
	for id, identifier := range nested.Identifiers {
		if identifier == nil {
			continue
		}
		if _, seeded := remap.Identifiers[IdentifierId(id)]; seeded {
			continue
		}
		declaration, mapped := declarations[identifier.Declaration]
		var copied *Identifier
		if !mapped || identifier.Declaration == 0 {
			copied = parent.NewIdentifier(identifier.Name, identifier.Node, 0)
			if identifier.Declaration != 0 {
				declarations[identifier.Declaration] = copied.Declaration
			}
		} else {
			copied = parent.NewIdentifier(identifier.Name, identifier.Node, declaration)
		}
		remap.Identifiers[IdentifierId(id)] = copied.Id
	}

	// Blocks, allocated before instructions so a terminal can name a block copied later.
	for _, block := range nested.Blocks {
		if block == nil {
			continue
		}
		remap.Blocks[block.Id] = parent.NewBlock(block.Kind).Id
	}
	remap.Entry = remap.Blocks[nested.Entry]

	// Nested functions of the nested function, appended to the parent's table.
	functionRemap := map[FunctionId]FunctionId{}
	for id, inner := range nested.Functions {
		if inner == nil {
			continue
		}
		parent.Functions = append(parent.Functions, inner)
		functionRemap[FunctionId(id)] = FunctionId(len(parent.Functions) - 1)
	}

	for _, block := range nested.Blocks {
		if block == nil {
			continue
		}
		target, found := parent.Block(remap.Blocks[block.Id])
		if !found || target == nil {
			return nil, false
		}
		for _, instructionId := range block.Instructions {
			source := nested.Instructions[instructionId]
			if source == nil {
				continue
			}
			copied := &Instruction{
				Order:  source.Order,
				LValue: source.LValue,
				Value:  copyInstructionValue(source.Value),
				Node:   source.Node,
				Range:  source.Range,
			}
			// Every place, the lvalue included, is mapped exactly once from its nested id. Mapping the
			// lvalue here as well mapped it twice wherever a fresh parent id was also a nested id
			// that the remap holds: `$26` became `$51` and then `$76`, a definition nothing reads and
			// an id defined twice, so a memo callback whose ids overlapped the parent's fresh range
			// lost its dependency and reported a memoization the compiler keeps (#p67vev4).
			EachInstructionPlacePointer(copied, func(place *Place, role PlaceRole) {
				if mapped, ok := remap.Identifiers[place.Identifier]; ok {
					place.Identifier = mapped
				}
			})
			remapNestedFunctionIds(copied.Value, functionRemap)
			remap.Instructions[instructionId] = parent.AddInstruction(target, copied)
		}

		target.Terminal = copyTerminal(block.Terminal)
		EachTerminalPlacePointer(target.Terminal, func(place *Place, role PlaceRole) {
			if mapped, ok := remap.Identifiers[place.Identifier]; ok {
				place.Identifier = mapped
			}
		})
		EachBlockReferencePointer(target.Terminal, func(reference *BlockId) {
			if mapped, ok := remap.Blocks[*reference]; ok {
				*reference = mapped
			}
		})

		for _, phi := range block.Phis {
			operands := map[BlockId]Place{}
			for predecessor, operand := range phi.Operands {
				mappedBlock, blockOk := remap.Blocks[predecessor]
				if !blockOk {
					mappedBlock = predecessor
				}
				if mapped, ok := remap.Identifiers[operand.Identifier]; ok {
					operand.Identifier = mapped
				}
				operands[mappedBlock] = operand
			}
			place := phi.Place
			if mapped, ok := remap.Identifiers[place.Identifier]; ok {
				place.Identifier = mapped
			}
			target.Phis = append(target.Phis, &Phi{Place: place, Operands: operands})
		}
	}

	return remap, true
}

// remapNestedFunctionIds repoints a value that names a function in the nested table.
func remapNestedFunctionIds(value InstructionValue, functionRemap map[FunctionId]FunctionId) {
	switch shape := value.(type) {
	case *FunctionExpression:
		if mapped, ok := functionRemap[shape.Function]; ok {
			shape.Function = mapped
		}
	case *ObjectMethod:
		if mapped, ok := functionRemap[shape.Function]; ok {
			shape.Function = mapped
		}
	}
}

// copyTerminal returns a terminal holding the same field values as its argument, in storage the
// argument does not share.
//
// The copy exists because remapping rewrites through pointers. `EachTerminalPlacePointer` and
// `EachBlockReferencePointer` both hand out pointers into the terminal they are given, and the
// caller writes through them; handing them the nested function's own terminal would rename the
// nested body rather than the copy.
//
// # Why one level is not enough, which is the correction
//
// This was written as a one-level copy on the reasoning that a terminal's fields are "places, block
// ids and evaluation orders, all values". Two are not. `Switch.Cases` is a `[]SwitchCase`, and a
// struct copy duplicates the slice HEADER while both headers keep pointing at one backing array --
// so a rewrite through a copied case writes into the original. Each `SwitchCase.Test` is a `*Place`,
// and a copied pointer is the same place.
//
// The measured symptom was in the sibling instruction path, which had the same shallow shape: after
// a copy, ten of nineteen places in the nested body had been renamed to the parent's ids, and
// printing the nested function panicked on an identifier past the end of its table. Nothing in the
// remap's own mutation sweep saw it, because that sweep asserted only over block REFERENCES and
// this damage is to places.
//
// So the copy follows pointers and slices to the bottom. `deepCopyValue` is the one implementation
// and both terminals and instruction values go through it.
func copyTerminal(terminal Terminal) Terminal {
	if terminal == nil {
		return nil
	}
	clone, ok := deepCopyAny(terminal).(Terminal)
	if !ok {
		return terminal
	}
	return clone
}

// copyInstructionValue returns an instruction value sharing no mutable storage with its argument.
//
// Same reason as `copyTerminal`, and the same measurement: `EachInstructionPlacePointer` hands out
// pointers into the value, so a copied instruction that still holds the original's `*CallExpression`
// has its operands renamed in the nested function rather than in the copy.
func copyInstructionValue(value InstructionValue) InstructionValue {
	if value == nil {
		return nil
	}
	clone, ok := deepCopyAny(value).(InstructionValue)
	if !ok {
		return value
	}
	return clone
}

// deepCopyAny duplicates a value and everything reachable from it through pointers, slices, maps
// and interfaces.
//
// Written by reflection rather than as one arm per type because the arms would say nothing. A
// hand-written `case *If: return &If{...}` restates the struct definition, and the restatement is
// what rots: a field added to a terminal is silently dropped by a copier that predates it, and
// nothing fails until a rule reads the field through an inlined body. The failure this replaced was
// exactly that shape one level down -- a `[]SwitchCase` the copier's own comment claimed could not
// exist.
//
// `*ast.Node` and the checker types reachable from it are deliberately NOT followed: they are the
// syntax tree, shared by every consumer in the process and never rewritten by a remap. Following
// them would copy the program. The stop condition is therefore structural rather than a type list:
// a pointer to a struct declared outside this package is kept as-is.
func deepCopyAny(value any) any {
	if value == nil {
		return nil
	}
	source := reflect.ValueOf(value)
	return deepCopyValue(source).Interface()
}

func deepCopyValue(source reflect.Value) reflect.Value {
	switch source.Kind() {
	case reflect.Pointer:
		if source.IsNil() {
			return source
		}
		// The syntax tree and anything else this package does not declare is shared, not copied.
		// A remap never writes through those, and copying them would duplicate the program.
		if source.Type().Elem().PkgPath() != hirPackagePath {
			return source
		}
		copied := reflect.New(source.Type().Elem())
		copied.Elem().Set(deepCopyValue(source.Elem()))
		return copied
	case reflect.Interface:
		if source.IsNil() {
			return source
		}
		copied := reflect.New(source.Type()).Elem()
		copied.Set(deepCopyValue(source.Elem()))
		return copied
	case reflect.Slice:
		if source.IsNil() {
			return source
		}
		copied := reflect.MakeSlice(source.Type(), source.Len(), source.Len())
		for index := 0; index < source.Len(); index++ {
			copied.Index(index).Set(deepCopyValue(source.Index(index)))
		}
		return copied
	case reflect.Map:
		if source.IsNil() {
			return source
		}
		copied := reflect.MakeMapWithSize(source.Type(), source.Len())
		iterator := source.MapRange()
		for iterator.Next() {
			copied.SetMapIndex(deepCopyValue(iterator.Key()),
				deepCopyValue(iterator.Value()))
		}
		return copied
	case reflect.Struct:
		copied := reflect.New(source.Type()).Elem()
		copied.Set(source)
		if source.Type().PkgPath() != hirPackagePath {
			// A struct from another package is copied whole and not descended into, for the same
			// reason its pointers are not followed.
			return copied
		}
		for index := 0; index < source.NumField(); index++ {
			field := copied.Field(index)
			if !field.CanSet() {
				// Unexported. The struct copy above already carried its bits across, which is the
				// most a copier outside the declaring type can do, and nothing a remap rewrites is
				// unexported.
				continue
			}
			field.Set(deepCopyValue(source.Field(index)))
		}
		return copied
	default:
		return source
	}
}

// hirPackagePath is this package, used to decide what deepCopyValue descends into.
//
// Taken from a declared type rather than written as a string so a package move cannot silently turn
// the copy shallow again.
var hirPackagePath = reflect.TypeOf(Place{}).PkgPath()
