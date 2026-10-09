package lower

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Discover concrete identities, then construct allocation flags without copying IR.
type programMembership struct {
	types                     map[int]bool
	cells, functions, classes map[int]bool
	identities                []string
}

func (l *lowering) selectProgramMembership(f *cycleFinder, modules []*ast.SourceFile) *programMembership {
	// Observe allocation and contextual identities even where all writes passed the
	// fresh proof and the ordinary graph-region pass had no seeds (e.g. Weak trees).
	for _, module := range modules {
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if ast.IsPartOfTypeNode(node) {
				return false
			}
			if ast.IsExpressionNode(node) {
				f.use(f.l.checker.GetTypeAtLocation(node), node)
				if contextual := f.l.checker.GetContextualType(node, checker.ContextFlagsNone); contextual != nil {
					f.use(contextual, node)
				}
			}
			return node.ForEachChild(visit)
		}
		module.AsNode().ForEachChild(visit)
	}
	for _, closure := range f.l.closureRecords {
		f.use(closure.proven, closure.node)
	}
	for proven, where := range l.programCandidates {
		f.use(proven, where)
	}
	for local, proven := range l.localTypes {
		if proven != nil {
			f.use(proven, l.localNodes[local])
		}
	}
	selected := f.programRegionSelection()
	program := f.l.result
	program.ProgramTypes = map[int]bool{}
	for node := range selected {
		if node.proven != nil && !f.weak(node.proven) && !f.l.isLibraryType(node.proven, "Promise") {
			program.ProgramTypes[int(node.proven.Id())] = true
		}
		if node.cell != 0 {
			program.Locals[node.cell-1].ProgramRegion = true
		}
	}
	for _, closure := range f.l.closureRecords {
		program.Functions[closure.function].ProgramRegion = program.ProgramTypes[int(closure.proven.Id())]
	}
	// Capture environments are indivisible allocations; retaining an extra sibling
	// costs retention, while treating an interior cell as independently freeable is unsafe.
	for changed := true; changed; {
		changed = false
		for i := range program.Functions {
			function := &program.Functions[i]
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
	for _, instance := range f.l.instances {
		for _, local := range instance.thisLocals {
			for _, proven := range append([]*checker.Type{f.l.localTypes[local]}, f.l.localAlso[local]...) {
				if proven != nil && program.ProgramTypes[int(proven.Id())] {
					program.Classes[instance.class-1].ProgramRegion = true
				}
			}
		}
	}
	for symbol, instance := range f.l.statics {
		if program.ProgramTypes[int(f.l.checker.GetTypeOfSymbol(symbol).Id())] {
			program.Classes[instance.class-1].ProgramRegion = true
		}
	}
	plan := &programMembership{types: program.ProgramTypes, cells: map[int]bool{}, functions: map[int]bool{}, classes: map[int]bool{}, identities: programAllocationIdentities(program)}
	for i, x := range program.Locals {
		plan.cells[i] = x.ProgramRegion
	}
	for i, x := range program.Functions {
		plan.functions[i] = x.ProgramRegion
	}
	for i, x := range program.Classes {
		plan.classes[i+1] = x.ProgramRegion
	}
	return plan
}

func programAllocationIdentities(p *ir.Program) []string {
	result := []string{}
	for i, l := range p.Locals {
		result = append(result, fmt.Sprint("local", i, l.Name, l.Type, l.Function, l.Captured, l.EnvironmentCell))
	}
	for i, f := range p.Functions {
		result = append(result, fmt.Sprint("function", i, f.Name, f.Parameters, f.Environment, f.FrameEnvironment))
	}
	for i, c := range p.Classes {
		result = append(result, fmt.Sprint("class", i, c.Name, c.Constructor, c.Definition))
	}
	return result
}
func (l *lowering) applyProgramMetadata(plan *programMembership) error {
	if fmt.Sprint(programAllocationIdentities(l.result)) != fmt.Sprint(plan.identities) {
		return fmt.Errorf("lower: Program discovery changed allocation identities")
	}
	for i := range l.result.Locals {
		l.result.Locals[i].ProgramRegion = plan.cells[i]
	}
	for i := range l.result.Functions {
		l.result.Functions[i].ProgramRegion = plan.functions[i]
		if plan.functions[i] && l.result.Functions[i].NestedParent > 0 {
			return &Refused{Where: l.result.Source, What: "Program canonical closure before canonical-cache adoption", Fix: "await runtime's 06c canonical closure adoption; use an anonymous closure or a module function"}
		}
	}
	for i := range l.result.Classes {
		l.result.Classes[i].ProgramRegion = plan.classes[i+1]
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
	walk(l.result.Main, check)
	for _, f := range l.result.Functions {
		walk(f.Body, check)
	}
	if entryPairs && len(plan.types) != 0 {
		return &Refused{Where: l.result.Source, What: "Map entry pair allocation in a Program-region program", Fix: "iterate keys and get values until entry pairs can adopt before publication"}
	}
	if parallel && len(plan.types) != 0 {
		return &Refused{Where: l.result.Source, What: "parallelMap in a program with Program-region members", Fix: "keep Program members in the one-shot CLI thread"}
	}
	return nil
}
func (l *lowering) programClass(class int) bool {
	return l.programPlan != nil && l.programPlan.classes[class]
}
func (l *lowering) programAllocation(node *ast.Node, value ir.Expression) (ir.Expression, error) {
	if !l.programDiscovery && l.programPlan == nil {
		return value, nil
	}
	if l.programPlan != nil && len(l.programPlan.types) != 0 && node.Kind == ast.KindCallExpression {
		callee := ast.SkipParentheses(node.AsCallExpression().Expression)
		if callee.Kind == ast.KindPropertyAccessExpression && callee.Name().Text() == "entries" {
			receiver := callee.AsPropertyAccessExpression().Expression
			if l.isLibraryType(l.checker.GetTypeAtLocation(receiver), "Map", "ReadonlyMap", "Set", "ReadonlySet") {
				return nil, &Refused{Where: l.result.Source, What: "Map entry pair allocation in a Program-region program", Fix: "iterate keys and get values until entry pairs can adopt before publication"}
			}
		}
	}
	member := false
	types := []*checker.Type{l.checker.GetTypeAtLocation(node), l.checker.GetContextualType(node, checker.ContextFlagsNone)}
	if target := l.impliedTarget(node); target != nil {
		types = append(types, target)
	}
	for _, proven := range types {
		if proven == nil {
			continue
		}
		proven = l.concrete(proven)
		if l.programCandidates == nil {
			l.programCandidates = map[*checker.Type]*ast.Node{}
		}
		l.programCandidates[proven] = node
		if l.programPlan != nil && l.programPlan.types[int(proven.Id())] {
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
