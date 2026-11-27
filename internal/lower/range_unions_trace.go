package lower

import (
	"fmt"
	"os"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

// Probe only, never for main (#nxb3xh1): a trace of what held reached when checkRangeUnionReads stops inside a
// helper in a first-link lowering, printed to stderr a line each, RANGETRACE first, at most rangeTraceCap lines.

const rangeTraceCap = 50

// rangeTraceLines counts the lines printed so far, across every rerun of the pass.
var rangeTraceLines int

func rangeTracef(format string, arguments ...any) {
	if rangeTraceLines >= rangeTraceCap {
		return
	}
	rangeTraceLines++
	line := fmt.Sprintf(format, arguments...)
	if len(line) > 700 {
		line = line[:700] + "..."
	}
	fmt.Fprintln(os.Stderr, "RANGETRACE "+line)
}

// rangeTrace traces a stop at operand inside helper: the stop, why held answered as it did from the operand,
// and every call of the helper whose argument held answers may be a range value for, with its caller, the
// caller's declaration or Main statement, the argument's IR and its checker type where known, and why. A caller
// that is a helper itself is followed to its own calls, a few deep.
func (l *lowering) rangeTrace(r *rangeOperands, helper int, operand ir.Expression, road string, site stopSite) {
	rangeTracef("stop in helper %d %s: %s, operand %s, stopWhere %s", helper, l.result.Functions[helper].Name, road, l.rangeTraceWhat(operand), l.stopWhere(site))
	for _, why := range l.rangeTraceWhy(r, operand, 6) {
		rangeTracef("  operand answers: %s", why)
	}
	l.rangeTraceCallers(r, helper, map[int]bool{}, 0)
}

// rangeTraceCallers prints each call of target whose arguments held answers true for.
func (l *lowering) rangeTraceCallers(r *rangeOperands, target int, seen map[int]bool, depth int) {
	if seen[target] || depth > 3 {
		return
	}
	seen[target] = true
	calls, answered := 0, 0
	visit := func(caller func() string, callerIndex int) func(node any) bool {
		return func(node any) bool {
			var arguments []ir.Expression
			calling := false
			switch call := node.(type) {
			case ir.Call:
				for _, each := range r.l.result.CallTargets(call) {
					if each == target {
						arguments, calling = call.Arguments, true
					}
				}
			case ir.CallClosure:
				targets := r.l.result.ClosureTargets(call)
				for _, each := range targets.Functions {
					if each == target {
						arguments, calling = call.Arguments, true
					}
				}
			}
			if !calling {
				return true
			}
			calls++
			for index, argument := range arguments {
				if argument == nil || !(&rangeOperands{l: l, graph: r.flow()}).held(argument) {
					continue
				}
				answered++
				rangeTracef("%scall of %s from %s, argument %d %s, held true", strings.Repeat("  ", depth+1), l.result.Functions[target].Name, caller(), index, l.rangeTraceWhat(argument))
				for _, why := range l.rangeTraceWhy(r, argument, 3) {
					rangeTracef("%s  because %s", strings.Repeat("  ", depth+1), why)
				}
				if callerIndex >= 0 && callerIndex < len(l.functionNodes) && l.functionNodes[callerIndex] == nil {
					l.rangeTraceCallers(r, callerIndex, seen, depth+1)
				}
			}
			return true
		}
	}
	for index, statement := range l.result.Main {
		walk(statement, visit(func() string { return l.rangeTraceMainWhere(index) }, -1))
	}
	for index, function := range l.result.Functions {
		walk(function.Body, visit(func() string {
			where := "helper"
			if index < len(l.functionNodes) && l.functionNodes[index] != nil && l.program != nil {
				where = l.program.Where(l.functionNodes[index])
			}
			stub := ""
			if firstLinkStubBody(function.Body) {
				stub = " (stub)"
			}
			return fmt.Sprintf("function %d %s at %s%s", index, function.Name, where, stub)
		}, index))
	}
	rangeTracef("%s%s: %d calls, %d arguments held true", strings.Repeat("  ", depth+1), l.result.Functions[target].Name, calls, answered)
}

// rangeTraceMainWhere names Main's statement index by the top-level statement it lowered from, where one is placed.
func (l *lowering) rangeTraceMainWhere(index int) string {
	for unit := range l.firstLinkMain {
		if start, length, placed := l.firstLinkMainAt(unit); placed && index >= start && index < start+length && l.program != nil {
			return fmt.Sprintf("Main statement %d at %s", index, l.program.Where(unit))
		}
	}
	return fmt.Sprintf("Main statement %d", index)
}

// rangeTraceWhat names an IR value: its Go type, and for a Read its local's name, function and checker type.
func (l *lowering) rangeTraceWhat(value ir.Expression) string {
	switch value := value.(type) {
	case nil:
		return "nil"
	case ir.Read:
		name, function := "?", -1
		if value.Local >= 0 && value.Local < len(l.result.Locals) {
			name, function = l.result.Locals[value.Local].Name, l.result.Locals[value.Local].Function
		}
		proven := "no checker type"
		if typed, ok := l.localTypes[value.Local]; ok && typed != nil {
			proven = "checker type " + l.checker.TypeToString(typed)
		}
		return fmt.Sprintf("Read local %d %q of function %d, %s", value.Local, name, function, proven)
	case ir.Call:
		names := []string{}
		for _, target := range l.result.CallTargets(value) {
			names = append(names, l.result.Functions[target].Name)
		}
		return fmt.Sprintf("Call of %v", names)
	case ir.Property:
		return "Property " + value.Name
	case ir.DynamicProperty:
		return "DynamicProperty " + value.Name
	case ir.Box:
		return "Box of " + l.rangeTraceWhat(value.Value)
	}
	return fmt.Sprintf("%T", value)
}

// rangeTraceWhy is every answer held gives true for from value, a line each, held's own walk with nothing cut short,
// at most limit.
func (l *lowering) rangeTraceWhy(r *rangeOperands, value ir.Expression, limit int) []string {
	whys := []string{}
	queue := []ir.Expression{value}
	visited := map[int]bool{}
	add := func(format string, arguments ...any) {
		if len(whys) < limit {
			whys = append(whys, fmt.Sprintf(format, arguments...))
		}
	}
	sources := func(node int, what string) {
		if visited[node] {
			return
		}
		visited[node] = true
		graph := r.flow()
		if why, unknown := graph.unknownAt(node); unknown {
			add("%s: the graph can't place it: %s", what, why)
			return
		}
		if len(graph.sources[node]) == 0 {
			add("%s: no source", what)
			return
		}
		queue = append(queue, graph.sources[node]...)
	}
	for steps := 0; len(queue) != 0 && len(whys) < limit && steps < 2000; steps++ {
		value := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		switch value := value.(type) {
		case nil:
			add("a nil value")
		case ir.Read:
			if proven, typed := l.localTypes[value.Local]; typed {
				if l.holdsRangeUnion(proven) {
					add("%s holds a range type", l.rangeTraceWhat(value))
				}
				for _, also := range l.localAlso[value.Local] {
					if l.holdsRangeUnion(also) {
						add("%s is also %s, a range type", l.rangeTraceWhat(value), l.checker.TypeToString(also))
					}
				}
				continue
			}
			sources(value.Local+1, l.rangeTraceWhat(value))
		case ir.Call:
			targets := l.result.CallTargets(value)
			if len(targets) == 0 {
				add("%s: no target", l.rangeTraceWhat(value))
			}
			for _, target := range targets {
				if target < len(l.functionNodes) && l.functionNodes[target] != nil {
					if returned := l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(l.functionNodes[target])); l.holdsRangeUnion(returned) {
						add("a call of %s, which returns %s, a range type", l.result.Functions[target].Name, l.checker.TypeToString(returned))
					}
					continue
				}
				sources(r.flow().resultNode(target), "the result of helper "+l.result.Functions[target].Name)
			}
		case ir.CallClosure:
			targets := l.result.ClosureTargets(value)
			if targets.Unknown || len(targets.Functions) == 0 {
				add("a closure call with targets unknown")
			}
			for _, target := range targets.Functions {
				sources(r.flow().resultNode(target), "the result of closure "+l.result.Functions[target].Name)
			}
		case ir.Box:
			if value.Value == nil || value.Value.Type() == ir.Array {
				add("a Box of an array, which may be a NodeArray")
				continue
			}
			queue = append(queue, value.Value)
		case ir.Conditional:
			queue = append(queue, value.WhenTrue, value.WhenNot)
		case ir.Coalesce:
			queue = append(queue, value.Value)
			if value.Panic == nil {
				queue = append(queue, value.Fallback)
			}
		case ir.Logical:
			queue = append(queue, value.Left, value.Right)
		case ir.Comma:
			queue = append(queue, value.Right)
		case ir.Effects:
			queue = append(queue, value.Result)
		case ir.Narrow:
			queue = append(queue, value.Value)
		case ir.Unwrap:
			queue = append(queue, value.Value)
		case ir.Defined:
			queue = append(queue, value.Value)
		case ir.CheckedCast:
			queue = append(queue, value.Value)
		case ir.Property:
			if l.rangeFieldNamed(value.Name) {
				add("a read of field %s, a range field name", value.Name)
			}
		case ir.DynamicProperty:
			if l.rangeFieldNamed(value.Name) {
				add("a dynamic read of field %s, a range field name", value.Name)
			}
		case ir.Undefined, ir.Null, ir.NumberConstant, ir.StringConstant, ir.BooleanConstant, ir.ObjectLiteral, ir.ArrayLiteral:
		default:
			add("a %T, storage the IR keeps no type for", value)
		}
	}
	return whys
}
