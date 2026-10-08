package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Unsupported contracts are descriptors, never certificates. Their users must
// refuse at a demanded read. Cast admission does not inspect their payloads.
func (l *lowering) viewContract(node *ast.Node, target *checker.Type) (ir.ViewContractID, error) {
	if l.result.ViewContractTypes == nil {
		l.result.ViewContractTypes = map[int]ir.ViewContractID{}
	}
	if id := l.result.ViewContractTypes[int(target.Id())]; id != 0 {
		return id, nil
	}
	family := l.unsupportedViewFamily(target)
	if l.callableViewContract(target) {
		family = "callable"
	}
	if l.checker.IsArrayType(target) {
		family = "array"
	}
	if of, known := l.representation(target); target.Flags()&checker.TypeFlagsUnion != 0 && (!interfaceScalar(target) || !known || of == ir.Union) {
		family = "union"
	}
	var hook viewContractHook
	switch family {
	case "array", "tuple":
		hook = viewArrayContractHook
	case "callable":
		hook = viewCallableContractHook
	case "union":
		hook = viewUnionContractHook
	case "dictionary":
		hook = viewDictionaryContractHook
	case "intersection":
		hook = viewIntersectionContractHook
	}
	if hook != nil {
		id, err := hook(l, node, target, func(child *checker.Type) (ir.ViewContractID, error) { return l.viewContract(node, child) })
		if err == nil && id != 0 {
			return id, nil
		}
	}
	if family == "" {
		id, err := l.strictViewContract(node, target)
		if err == nil {
			return id, nil
		}
		family = "representation conversion"
	}
	// A failing family adapter may have reserved its recursive id already.
	id := l.result.ViewContractTypes[int(target.Id())]
	descriptor := ir.ViewContract{Kind: ir.ViewUnknown, Name: l.checker.TypeToString(target), Unsupported: family}
	if id == 0 {
		id = ir.ViewContractID(len(l.result.ViewContracts) + 1)
		l.result.ViewContracts = append(l.result.ViewContracts, descriptor)
		l.result.ViewContractTypes[int(target.Id())] = id
	} else {
		l.result.ViewContracts[id-1] = descriptor
	}
	return id, nil
}

func (l *lowering) unsupportedViewFamily(target *checker.Type) string {
	flags := target.Flags()
	switch {
	case flags&checker.TypeFlagsAny != 0:
		return "any"
	case flags&checker.TypeFlagsUnknown != 0:
		return "unknown"
	case flags&checker.TypeFlagsNever != 0:
		return "never"
	case flags&checker.TypeFlagsTypeParameter != 0:
		return "generic"
	case flags&checker.TypeFlagsIntersection != 0:
		return "intersection"
	case flags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined) != 0:
		return "nullish"
	case isClassInstance(target):
		return "nominal class"
	case l.isLibraryType(target, "Map", "ReadonlyMap", "Set", "ReadonlySet"):
		return "collection"
	case checker.IsTupleType(target):
		return "tuple"
	case flags&checker.TypeFlagsObject != 0 && len(l.checker.GetIndexInfosOfType(target)) != 0 && !l.checker.IsArrayType(target):
		return "dictionary"
	}

	return ""
}

// Demand uses the same allocations, joined arguments/results and projected
// stores as shape certification. Unknown is never an empty proof of safety.
func (l *lowering) checkLazyViewReads() error {
	program := l.result
	if len(program.ViewOrigins) == 0 {
		return nil
	}
	assignAllocationSites(program)
	graph := newAllocationFlowGraph(program)
	viewed := map[int]bool{}
	unknown := false
	queue := []int{}
	add := func(set ir.AllocationSet) {
		unknown = unknown || set.Unknown
		for _, site := range set.Sites {
			if !viewed[site] {
				viewed[site] = true
				queue = append(queue, site)
			}
		}
	}
	for _, origin := range program.ViewOrigins {
		add(graph.ReachingAllocations(origin))
	}
	index := graph.projectionIndex()
	for len(queue) != 0 {
		site := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		if literal, ok := index.records[site]; ok {
			for _, field := range literal.Fields {
				if viewAggregate(field.Value) {
					add(graph.ReachingAllocations(field.Value))
				}
			}
			if literal.Spread != nil {
				add(graph.ReachingAllocations(literal.Spread))
			}
		}
		if literal, ok := index.arrays[site]; ok {
			for _, element := range literal.Elements {
				if viewAggregate(element) {
					add(graph.ReachingAllocations(element))
				}
			}
		}
		for _, values := range index.stores[site] {
			for _, value := range values {
				if viewAggregate(value) {
					add(graph.ReachingAllocations(value))
				}
			}
		}
	}
	unsupportedFields := map[string]string{}
	for _, descriptor := range program.ViewContracts {
		for _, field := range descriptor.Fields {
			if family := program.ViewContracts[field.Contract-1].Unsupported; family != "" {
				unsupportedFields[field.Name] = family
			}
		}
	}
	var refused error
	disjoint := map[string]bool{}
	inspect := func(node any) bool {
		if refused != nil {
			return false
		}
		read, ok := node.(ir.Property)
		if !ok || !program.CheckedFields[read.Name] {
			return true
		}
		contract := program.ViewContractTypes[read.ViewTypeID]
		family := ""
		if contract != 0 {
			family = program.ViewContracts[contract-1].Unsupported
		}
		if family == "" {
			family = unsupportedFields[read.Name]
		}
		if family == "" {
			return true
		}
		reaches := graph.ReachingAllocations(read.Object)
		demanded := unknown || reaches.Unknown
		for _, site := range reaches.Sites {
			demanded = demanded || viewed[site]
		}
		if !demanded {
			disjoint[read.ViewWhere+"\x00"+read.Name] = true
		}
		if demanded {
			refused = &Refused{Where: read.ViewWhere, What: "checked view read of field " + read.Name + " with unsupported " + family + " contract", Fix: "prove or implement the " + family + " contract before reading this field"}
		}
		return true
	}
	walk(program.Main, inspect)
	for _, function := range program.Functions {
		walk(function.Body, inspect)
	}
	if refused != nil {
		return refused
	}
	// No view check is erased: this unsupported read belongs solely to ordinary
	// allocations. Unknown receivers never establish disjointness.
	rewrite := func(node any) any {
		if read, ok := node.(ir.Property); ok && disjoint[read.ViewWhere+"\x00"+read.Name] {
			read.ViewOrdinary = true
			return read
		}
		return node
	}
	program.Main = rewriteShapeStatements(program.Main, rewrite)
	for i := range program.Functions {
		program.Functions[i].Body = rewriteShapeStatements(program.Functions[i].Body, rewrite)
	}
	return nil
}

func viewAggregate(value ir.Expression) bool {
	if value == nil {
		return false
	}
	switch value.Type() {
	case ir.Object, ir.Array, ir.Map, ir.Union, ir.Closure:
		return true
	}
	return false
}
