package lower

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"maps"
	"slices"
)

// Discover concrete identities, then construct allocation flags without copying IR.
type programMembership struct {
	types                     ir.SortedMap[int, bool]
	cells, functions, classes map[int]bool
	identities                []string
	refusal                   error
}

func (lowering *lowering) selectProgramMembership(finder *cycleFinder, modules []*ast.SourceFile) *programMembership {
	// Observe allocation and contextual identities even where all writes passed the
	// fresh proof and the ordinary graph-region pass had no seeds (e.g. Weak trees).
	for _, module := range modules {
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if ast.IsPartOfTypeNode(node) {
				return false
			}
			if ast.IsExpressionNode(node) {
				finder.use(finder.lowering.checker.GetTypeAtLocation(node), node)
				if contextual := finder.lowering.slotType(node); contextual != nil {
					finder.use(contextual, node)
				}
			}
			return node.ForEachChild(visit)
		}
		module.AsNode().ForEachChild(visit)
	}
	for _, closure := range finder.lowering.closureRecords {
		finder.use(closure.proven, closure.node)
	}
	candidates := slices.Collect(maps.Keys(lowering.programCandidates))
	slices.SortFunc(candidates, func(left, right *checker.Type) int { return int(left.Id()) - int(right.Id()) })
	for _, proven := range candidates {
		finder.use(proven, lowering.programCandidates[proven])
	}
	for _, local := range slices.Sorted(maps.Keys(lowering.localTypes)) {
		proven := lowering.localTypes[local]
		if proven != nil {
			finder.use(proven, lowering.localNodes[local])
		}
	}
	selected := finder.programRegionSelection()
	program := finder.lowering.result
	program.ProgramTypes = ir.SortedMap[int, bool]{}
	for node := range selected {
		if node.proven != nil && !finder.weak(node.proven) && !finder.lowering.isLibraryType(node.proven, "Promise") {
			program.ProgramTypes[finder.lowering.cycleTypeId(node.proven)] = true
		}
		if node.cell != 0 {
			program.Locals[node.cell-1].ProgramRegion = true
		}
	}
	for _, closure := range finder.lowering.closureRecords {
		program.Functions[closure.function].ProgramRegion = program.ProgramTypes[int(closure.proven.Id())]
	}
	// Capture environments are indivisible allocations; retaining an extra sibling
	// costs retention, while treating an interior cell as independently freeable is unsafe.
	for changed := true; changed; {
		changed = false
		for index := range program.Functions {
			function := &program.Functions[index]
			for _, local := range function.Environment {
				if program.Locals[local].ProgramRegion && !function.ProgramRegion {
					function.ProgramRegion = true
					changed = true
				}
			}
			if function.ProgramRegion {
				for _, local := range function.Environment {
					if !program.Locals[local].ProgramRegion {
						program.Locals[local].ProgramRegion = true
						changed = true
					}
				}
			}
			frame := false
			for _, local := range function.FrameEnvironment {
				frame = frame || program.Locals[local].ProgramRegion
			}
			if frame {
				for _, local := range function.FrameEnvironment {
					if !program.Locals[local].ProgramRegion {
						program.Locals[local].ProgramRegion = true
						changed = true
					}
				}
			}
		}
	}
	for _, instance := range finder.lowering.instances {
		for _, local := range instance.thisLocals {
			for _, proven := range append(
				[]*checker.Type{finder.lowering.localTypes[local]},
				finder.lowering.localAlso[local]...,
			) {
				if proven != nil && program.ProgramTypes[int(proven.Id())] {
					program.Classes[instance.class-1].ProgramRegion = true
				}
			}
		}
	}
	for symbol, instance := range finder.lowering.statics {
		if program.ProgramTypes[int(finder.lowering.checker.GetTypeOfSymbol(symbol).Id())] {
			program.Classes[instance.class-1].ProgramRegion = true
		}
	}
	// Cached canonical closures cannot be moved after publication. Keep the ordinary
	// graph plan until their entire component can adopt before caching (06c).
	for _, function := range program.Functions {
		if function.ProgramRegion && function.NestedParent > 0 {
			program.ProgramTypes = ir.SortedMap[int, bool]{}
			for index := range program.Locals {
				program.Locals[index].ProgramRegion = false
			}
			for index := range program.Functions {
				program.Functions[index].ProgramRegion = false
			}
			for index := range program.Classes {
				program.Classes[index].ProgramRegion = false
			}
			selected = map[cycleNode]bool{}
			break
		}
	}
	plan := &programMembership{
		types:      program.ProgramTypes,
		cells:      map[int]bool{},
		functions:  map[int]bool{},
		classes:    map[int]bool{},
		identities: programAllocationIdentities(program),
	}
	for node := range selected {
		if node.proven != nil && finder.lowering.isLibraryType(node.proven, "Promise") {
			plan.refusal = &Refused{
				Where: lowering.result.Source,
				What:  "a Promise whose payload reaches a Program member",
				Fix:   "remove the captured strong back-reference or use Weak (adamic/cycle-capable)",
			}
		}
	}
	if len(plan.types) != 0 && plan.refusal == nil {
		plan.refusal = finder.graphAdoptionBoundary()
	}

	for index, x := range program.Locals {
		plan.cells[index] = x.ProgramRegion
	}
	for index, x := range program.Functions {
		plan.functions[index] = x.ProgramRegion
	}
	for index, x := range program.Classes {
		plan.classes[index+1] = x.ProgramRegion
	}
	return plan
}

func programAllocationIdentities(program *ir.Program) []string {
	result := []string{}
	for index, local := range program.Locals {
		result = append(
			result,
			fmt.Sprint("local", index, local.Name, local.Type, local.Function, local.Captured, local.EnvironmentCell),
		)
	}
	for index, function := range program.Functions {
		result = append(result, fmt.Sprint(
			"function",
			index,
			function.Name,
			function.Parameters,
			function.Environment,
			function.FrameEnvironment,
		))
	}
	for index, class := range program.Classes {
		result = append(result, fmt.Sprint("class", index, class.Name, class.Constructor, class.Definition))
	}
	return result
}
func (lowering *lowering) applyProgramMetadata(plan *programMembership) error {
	if plan.refusal != nil {
		return plan.refusal
	}
	if fmt.Sprint(programAllocationIdentities(lowering.result)) != fmt.Sprint(plan.identities) {
		return fmt.Errorf("lower: Program discovery changed allocation identities")
	}
	// Keep the ordinary graph proof as a fallback if an allocation is not a Program member.
	for index := range lowering.result.Locals {
		lowering.result.Locals[index].ProgramRegion = plan.cells[index]
	}
	for index := range lowering.result.Functions {
		lowering.result.Functions[index].ProgramRegion = plan.functions[index]
		if plan.functions[index] && lowering.result.Functions[index].NestedParent > 0 {
			return &Refused{
				Where: lowering.result.Source,
				What:  "Program canonical closure before canonical-cache adoption",
				Fix:   "await runtime's 06c canonical closure adoption; use an anonymous closure or a module function",
			}
		}
	}
	for index := range lowering.result.Classes {
		lowering.result.Classes[index].ProgramRegion = plan.classes[index+1]
	}
	parallel, entryPairs := false, false
	check := func(node any) bool {
		switch node.(type) {
		case ir.ParallelMap:
			parallel = true
		case ir.MapEntries:
			entryPairs = true
		}
		return true
	}
	walk(lowering.result.Main, check)
	for _, function := range lowering.result.Functions {
		walk(function.Body, check)
	}
	if entryPairs && len(plan.types) != 0 {
		return &Refused{
			Where: lowering.result.Source,
			What:  "Map entry pair allocation in a Program-region program",
			Fix:   "iterate keys and get values until entry pairs can adopt before publication",
		}
	}
	if parallel && len(plan.types) != 0 {
		return &Refused{
			Where: lowering.result.Source,
			What:  "parallelMap in a program with Program-region members",
			Fix:   "keep Program members in the one-shot CLI thread",
		}
	}
	return nil
}
func (lowering *lowering) programClass(class int) bool {
	return lowering.programPlan != nil && lowering.programPlan.classes[class]
}
func (lowering *lowering) programAllocation(node *ast.Node, value ir.Expression) (ir.Expression, error) {
	if !lowering.programDiscovery && lowering.programPlan == nil {
		return value, nil
	}
	if lowering.programPlan != nil && len(lowering.programPlan.types) != 0 && node.Kind == ast.KindCallExpression {
		callee := ast.SkipParentheses(node.AsCallExpression().Expression)
		if callee.Kind == ast.KindPropertyAccessExpression && callee.Name().Text() == "entries" {
			receiver := callee.AsPropertyAccessExpression().Expression
			if lowering.isLibraryType(
				lowering.checker.GetTypeAtLocation(receiver),
				"Map",
				"ReadonlyMap",
				"Set",
				"ReadonlySet",
			) {
				return nil, &Refused{
					Where: lowering.result.Source,
					What:  "Map entry pair allocation in a Program-region program",
					Fix:   "iterate keys and get values until entry pairs can adopt before publication",
				}
			}
		}
	}
	member := false
	types := []*checker.Type{lowering.checker.GetTypeAtLocation(node), lowering.slotType(node)}
	if target := lowering.impliedTarget(node); target != nil {
		types = append(types, target)
	}
	for _, proven := range types {
		if proven == nil {
			continue
		}
		proven = lowering.concrete(proven)
		if lowering.programCandidates == nil {
			lowering.programCandidates = map[*checker.Type]*ast.Node{}
		}
		lowering.programCandidates[proven] = node
		if lowering.programPlan != nil && lowering.programPlan.types[int(proven.Id())] {
			member = true
		}
	}
	switch value := value.(type) {
	case ir.ObjectLiteral:
		value.ProgramRegion = member
		return value, nil
	case ir.ArrayLiteral:
		value.ProgramRegion = member
		return value, nil
	case ir.ArraySplice:
		value.ProgramRegion = member
		return value, nil
	case ir.ArrayFill:
		value.ProgramRegion = member
		return value, nil
	case ir.ArrayFrom:
		value.ProgramRegion = member
		return value, nil
	case ir.ArrayConcat:
		value.ProgramRegion = member
		return value, nil
	case ir.ArrayMap:
		value.ProgramRegion = member
		return value, nil
	case ir.ArrayVisit:
		value.ProgramRegion = member
		return value, nil
	case ir.MapEntries:
		value.ProgramRegion = member
		return value, nil
	case ir.ArraySlice:
		value.ProgramRegion = member
		return value, nil
	case ir.MapNew:
		value.ProgramRegion = member
		return value, nil
	case ir.MapKeys:
		value.ProgramRegion = member
		return value, nil
	case ir.MapValues:
		value.ProgramRegion = member
		return value, nil
	case ir.SetNew:
		value.ProgramRegion = member
		return value, nil
	case ir.SetValues:
		value.ProgramRegion = member
		return value, nil
	case ir.MakeClosure:
		value.ProgramRegion = member
		return value, nil
	}
	return value, nil
}
