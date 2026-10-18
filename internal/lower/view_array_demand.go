package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"reflect"
)

// Union fields also demand the owning array adapter, including nullable arrays.
func arrayViewContract(contracts []ir.ViewContract, id ir.ViewContractID, seen map[ir.ViewContractID]bool) (bool, bool) {
	if id <= 0 || int(id) > len(contracts) || seen[id] {
		return false, false
	}
	seen[id] = true
	c := contracts[id-1]
	if c.Kind == ir.ViewArray {
		return true, !c.ArrayReadonly
	}
	found, mutable := false, false
	for _, member := range c.Members {
		f, m := arrayViewContract(contracts, member, seen)
		found = found || f
		mutable = mutable || m
	}
	return found, mutable
}

// An unread array member must not change an otherwise admitted program.
func (l *lowering) activateViewArrayReads(graph *allocationFlowGraph, viewed map[int]bool, unknown bool) error {
	p := l.result
	each := func(visit func(any) bool) {
		walk(p.Main, visit)
		for _, f := range p.Functions {
			walk(f.Body, visit)
		}
	}
	each(func(n any) bool {
		read, ok := n.(ir.Property)
		if !ok || !p.CheckedFields[read.Name] {
			return true
		}
		id := p.ViewContractTypes[read.ViewTypeID]
		array, mutable := arrayViewContract(p.ViewContracts, id, map[ir.ViewContractID]bool{})
		if !array {
			return true
		}
		reaching := graph.ReachingAllocations(read.Object)
		demand := unknown || reaching.Unknown
		for _, site := range reaching.Sites {
			demand = demand || viewed[site]
		}
		p.ArrayViewEnabled = p.ArrayViewEnabled || demand
		if demand && mutable {
			p.ArrayViewNeedsSourceCertificate = true
		}
		return true
	})
	if !p.ArrayViewEnabled {
		if p.ArrayViewTemplateWhere != "" {
			return &NotYet{Where: p.ArrayViewTemplateWhere, What: "a template interpolating an object, an array, a map, a function or undefined"}
		}
		return nil
	}
	seen := map[ir.ViewContractID]bool{}
	var fields func(ir.ViewContractID)
	fields = func(id ir.ViewContractID) {
		if id == 0 || seen[id] {
			return
		}
		seen[id] = true
		c := p.ViewContracts[id-1]
		for _, f := range c.Fields {
			p.CheckedFields[f.Name] = true
			fields(f.Contract)
		}
		fields(c.Element)
		for _, member := range c.Members {
			fields(member)
		}
	}
	var refused error
	each(func(n any) bool {
		if refused != nil {
			return false
		}
		var read ir.ArrayViewRead
		switch n := n.(type) {
		case ir.ArrayIndex:
			read = ir.ArrayViewRead{View: n.View, ViewTypeID: n.ViewTypeID, Element: n.Element}
		case ir.ForOf:
			read = n.ViewRead
		case ir.ArrayMap:
			read = n.ViewRead
		case ir.ArrayVisit:
			read = n.ViewRead
		case ir.ArrayReduce:
			read = n.ViewRead
		case ir.ArrayPop:
			read = n.ViewRead
		case ir.ArrayJoin:
			read = n.ViewRead
			if n.Element != ir.Number && n.Element != ir.Boolean && n.Element != ir.String && n.Element != ir.MaybeNumber {
				refused = &NotYet{What: "checked array string conversion with non-primitive elements"}
			}
		case ir.ArraySearch, ir.ArraySort, ir.ArrayConcat, ir.ArraySplice, ir.ArrayFill, ir.ArrayReverse, ir.ArrayFrom, ir.JSONStringify:
			refused = &NotYet{What: "checked array consumer requiring its source storage certificate"}
		}
		if read.View != "" {
			id := p.ViewContractTypes[read.ViewTypeID]
			if id != 0 {
				c := p.ViewContracts[id-1]
				if c.Unsupported != "" {
					refused = &Refused{What: "checked array element read with unsupported " + c.Unsupported + " contract", Fix: "prove or implement the " + c.Unsupported + " contract before reading this element"}
				}
				fields(id)
			}
		}
		if p.ArrayViewNeedsSourceCertificate {
			v := reflect.ValueOf(n)
			if v.Kind() == reflect.Struct {
				f := v.FieldByName("Site")
				if f.IsValid() && f.Kind() == reflect.Int && f.Int() != 0 {
					refused = &NotYet{Where: p.Source, What: "writing through an array cast without its source-slot type certificate"}
				}
			}
		}
		return true
	})
	if refused != nil {
		return refused
	}
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return err
	}
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if n.Kind == ast.KindGetAccessor && n.Name() != nil && p.CheckedFields[n.Name().Text()] {
			refused = l.notYet(n, "a getter in a checked field contract")
			return true
		}
		return n.ForEachChild(visit)
	}
	for _, m := range modules {
		m.AsNode().ForEachChild(visit)
	}
	return refused
}
