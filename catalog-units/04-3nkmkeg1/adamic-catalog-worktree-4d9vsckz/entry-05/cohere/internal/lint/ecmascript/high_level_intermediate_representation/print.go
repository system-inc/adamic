// Printing a function back out as text.
//
// This exists to make the IR reviewable by a human and testable by a diff. Every lowering test in
// this package asserts against printed output rather than against a structure literal, because a
// structure literal test asserts what the author believed lowering does and a printed one shows
// what it did. The difference matters most exactly where lowering is subtle, which is where the
// tests are needed.
//
// The format is this package's own, not upstream's. Matching upstream's printer would be a second
// fidelity obligation with no oracle behind it: no fixture compares printed HIR, so the only reader
// is us.
package high_level_intermediate_representation

import (
	"fmt"
	"sort"
	"strings"
)

// Print renders a function as text.
func Print(function *Function) string {
	var out strings.Builder
	printFunction(&out, function, 0)
	return out.String()
}

func printFunction(out *strings.Builder, function *Function, depth int) {
	indent := strings.Repeat("  ", depth)

	name := function.Name
	if name == "" {
		name = "<anonymous>"
	}
	fmt.Fprintf(out, "%sfunction %s(", indent, name)
	for index, param := range function.Params {
		if index > 0 {
			out.WriteString(", ")
		}
		out.WriteString(function.PlaceString(param))
	}
	out.WriteString(")")
	if function.Kind != FunctionKindOther {
		fmt.Fprintf(out, " [%s]", function.Kind)
	}
	out.WriteString("\n")

	for _, block := range function.Blocks {
		fmt.Fprintf(out, "%sbb%d (%s)", indent, block.Id, block.Kind)
		if len(block.Predecessors) > 0 {
			predecessors := make([]string, 0, len(block.Predecessors))
			for _, predecessor := range block.Predecessors {
				predecessors = append(predecessors, fmt.Sprintf("bb%d", predecessor))
			}
			sort.Strings(predecessors)
			fmt.Fprintf(out, " preds=[%s]", strings.Join(predecessors, " "))
		}
		out.WriteString(":\n")

		// Operands are read through PhiOperandsInOrder, ascending predecessor order, so the same
		// function renders the same way between runs, which TestLowerIsDeterministic exists to
		// catch.
		for _, phi := range block.Phis {
			operands := make([]string, 0, len(phi.Operands))
			for _, predecessorId := range PhiOperandsInOrder(phi) {
				operands = append(operands, fmt.Sprintf("bb%d: %s",
					predecessorId, function.PlaceString(phi.Operands.At(predecessorId))))
			}
			fmt.Fprintf(out, "%s  %s = phi(%s)\n",
				indent, function.PlaceString(phi.Place), strings.Join(operands, ", "))
		}
		for _, instructionId := range block.Instructions {
			instruction := function.Instructions[instructionId]
			fmt.Fprintf(
				out,
				"%s  %s = %s\n",
				indent,
				function.PlaceString(instruction.LValue),
				printValue(function, instruction.Value),
			)
		}
		fmt.Fprintf(out, "%s  %s\n", indent, printTerminal(function, block.Terminal))
	}

	for _, nested := range function.Functions {
		out.WriteString("\n")
		printFunction(out, nested, depth+1)
	}
}

func printValue(function *Function, value InstructionValue) string {
	place := function.PlaceString
	switch v := value.(type) {
	case *LoadLocal:
		return fmt.Sprintf("LoadLocal %s", place(v.Place))
	case *LoadContext:
		return fmt.Sprintf("LoadContext %s", place(v.Place))
	case *DeclareLocal:
		return fmt.Sprintf("DeclareLocal %s %s", v.Kind, place(v.LValue))
	case *DeclareContext:
		return fmt.Sprintf("DeclareContext %s %s", v.Kind, place(v.LValue))
	case *StoreLocal:
		return fmt.Sprintf("StoreLocal %s %s = %s", v.Kind, place(v.LValue), place(v.Value))
	case *StoreContext:
		return fmt.Sprintf("StoreContext %s %s = %s", v.Kind, place(v.LValue), place(v.Value))
	case *Destructure:
		return fmt.Sprintf("Destructure %s %s = %s", v.Kind, printPattern(function, v.LValue), place(v.Value))
	case *LoadGlobal:
		if v.Source != "" {
			return fmt.Sprintf("LoadGlobal %s from %q", v.Name, v.Source)
		}
		return fmt.Sprintf("LoadGlobal %s", v.Name)
	case *StoreGlobal:
		return fmt.Sprintf("StoreGlobal %s = %s", v.Name, place(v.Value))
	case *PropertyLoad:
		return fmt.Sprintf("PropertyLoad %s%s.%s", place(v.Object), optionalMark(v.Optional), v.Property)
	case *PropertyStore:
		return fmt.Sprintf("PropertyStore %s.%s = %s", place(v.Object), v.Property, place(v.Value))
	case *PropertyDelete:
		return fmt.Sprintf("PropertyDelete %s.%s", place(v.Object), v.Property)
	case *ComputedLoad:
		return fmt.Sprintf("ComputedLoad %s%s[%s]", place(v.Object), optionalMark(v.Optional), place(v.Property))
	case *ComputedStore:
		return fmt.Sprintf("ComputedStore %s[%s] = %s", place(v.Object), place(v.Property), place(v.Value))
	case *ComputedDelete:
		return fmt.Sprintf("ComputedDelete %s[%s]", place(v.Object), place(v.Property))
	case *CallExpression:
		return fmt.Sprintf("Call %s%s(%s)", place(v.Callee), optionalMark(v.Optional), printArgs(function, v.Args))
	case *MethodCall:
		return fmt.Sprintf("MethodCall %s.%s(%s)", place(v.Receiver), place(v.Property), printArgs(function, v.Args))
	case *NewExpression:
		return fmt.Sprintf("New %s(%s)", place(v.Callee), printArgs(function, v.Args))
	case *BinaryExpression:
		return fmt.Sprintf("Binary %s %s %s", place(v.Left), v.Operator, place(v.Right))
	case *UnaryExpression:
		return fmt.Sprintf("Unary %s%s", v.Operator, place(v.Value))
	case *PrefixUpdate:
		return fmt.Sprintf("PrefixUpdate %s%s", v.Operation, place(v.Value))
	case *PostfixUpdate:
		return fmt.Sprintf("PostfixUpdate %s%s", place(v.Value), v.Operation)
	case *Primitive:
		return fmt.Sprintf("Primitive %v", v.Value)
	case *RegExpLiteral:
		return fmt.Sprintf("RegExp /%s/%s", v.Pattern, v.Flags)
	case *TemplateLiteral:
		return fmt.Sprintf("Template %d quasis %d exprs", len(v.Quasis), len(v.Subexprs))
	case *TaggedTemplateExpression:
		return fmt.Sprintf("TaggedTemplate %s", place(v.Tag))
	case *TypeCastExpression:
		return fmt.Sprintf("TypeCast %s", place(v.Value))
	case *MetaProperty:
		return fmt.Sprintf("MetaProperty %s.%s", v.Meta, v.Property)
	case *ObjectExpression:
		parts := make([]string, 0, len(v.Properties))
		for _, property := range v.Properties {
			switch {
			case property.Spread:
				parts = append(parts, "..."+place(property.Value))
			case property.ComputedKey != nil:
				parts = append(parts, fmt.Sprintf("[%s]: %s", place(*property.ComputedKey), place(property.Value)))
			default:
				parts = append(parts, fmt.Sprintf("%s: %s", property.Key, place(property.Value)))
			}
		}
		return fmt.Sprintf("Object {%s}", strings.Join(parts, ", "))
	case *ObjectMethod:
		return fmt.Sprintf("ObjectMethod %s fn%d", v.Key, v.Function)
	case *ArrayExpression:
		parts := make([]string, 0, len(v.Elements))
		for _, element := range v.Elements {
			switch {
			case element.Hole:
				parts = append(parts, "<hole>")
			case element.Spread:
				parts = append(parts, "..."+place(element.Place))
			default:
				parts = append(parts, place(element.Place))
			}
		}
		return fmt.Sprintf("Array [%s]", strings.Join(parts, ", "))
	case *FunctionExpression:
		captures := make([]string, 0, len(v.Captures))
		for _, capture := range v.Captures {
			captures = append(captures, place(capture))
		}
		if len(captures) > 0 {
			return fmt.Sprintf("Function fn%d captures=[%s]", v.Function, strings.Join(captures, " "))
		}
		return fmt.Sprintf("Function fn%d", v.Function)
	case *Await:
		return fmt.Sprintf("Await %s", place(v.Value))
	case *GetIterator:
		return fmt.Sprintf("GetIterator %s", place(v.Value))
	case *IteratorNext:
		return fmt.Sprintf("IteratorNext %s", place(v.Iterator))
	case *NextPropertyOf:
		return fmt.Sprintf("NextPropertyOf %s", place(v.Value))
	case *JsxExpression:
		tag := v.Tag.Name
		if v.Tag.Place != nil {
			tag = place(*v.Tag.Place)
		}
		return fmt.Sprintf("Jsx <%s> props=%d children=%d", tag, len(v.Props), len(v.Children))
	case *JsxFragment:
		return fmt.Sprintf("JsxFragment children=%d", len(v.Children))
	case *JsxText:
		return fmt.Sprintf("JsxText %q", v.Value)
	case *StartMemoize:
		// `deps=none` and `deps=0` are different facts (no dependency array at all versus an empty
		// one) and the printer distinguishes them, because a golden that spelled both `deps=0`
		// would pass while the pass conflated them.
		if v.Deps == nil {
			return fmt.Sprintf("StartMemoize#%d deps=none", v.ManualMemoId)
		}
		printed := make([]string, 0, len(v.Deps))
		for _, dep := range v.Deps {
			printed = append(printed, printManualMemoDependency(dep, place))
		}
		return fmt.Sprintf("StartMemoize#%d deps=[%s]", v.ManualMemoId, strings.Join(printed, ", "))
	case *FinishMemoize:
		if v.Pruned {
			return fmt.Sprintf("FinishMemoize#%d %s pruned", v.ManualMemoId, place(v.Value))
		}
		return fmt.Sprintf("FinishMemoize#%d %s", v.ManualMemoId, place(v.Value))
	case *Debugger:
		return "Debugger"
	case *UnsupportedNode:
		return fmt.Sprintf("Unsupported %s", v.Reason)
	}
	return "<unknown>"
}

func printPattern(function *Function, pattern Pattern) string {
	switch p := pattern.(type) {
	case *PlacePattern:
		return function.PlaceString(p.Place)
	case *ObjectPattern:
		parts := make([]string, 0, len(p.Properties))
		for _, property := range p.Properties {
			key := property.Key
			if property.ComputedKey != nil {
				key = "[" + function.PlaceString(*property.ComputedKey) + "]"
			}
			parts = append(parts, fmt.Sprintf("%s: %s", key, printPattern(function, property.Value)))
		}
		if p.Rest != nil {
			parts = append(parts, "..."+function.PlaceString(*p.Rest))
		}
		return "{" + strings.Join(parts, ", ") + "}"
	case *ArrayPattern:
		parts := make([]string, 0, len(p.Elements))
		for _, element := range p.Elements {
			if element.Value == nil {
				parts = append(parts, "<hole>")
				continue
			}
			parts = append(parts, printPattern(function, element.Value))
		}
		if p.Rest != nil {
			parts = append(parts, "..."+function.PlaceString(*p.Rest))
		}
		return "[" + strings.Join(parts, ", ") + "]"
	}
	return "<pattern>"
}

func printArgs(function *Function, args []Argument) string {
	parts := make([]string, 0, len(args))
	for _, arg := range args {
		if arg.Spread {
			parts = append(parts, "..."+function.PlaceString(arg.Place))
			continue
		}
		parts = append(parts, function.PlaceString(arg.Place))
	}
	return strings.Join(parts, ", ")
}

func optionalMark(optional bool) string {
	if optional {
		return "?"
	}
	return ""
}

func printTerminal(function *Function, terminal Terminal) string {
	place := function.PlaceString
	switch t := terminal.(type) {
	case *Return:
		return fmt.Sprintf("Return %s", place(t.Value))
	case *Throw:
		return fmt.Sprintf("Throw %s", place(t.Value))
	case *Unreachable:
		return "Unreachable"
	case *Unsupported:
		return "Unsupported"
	case *Goto:
		return fmt.Sprintf("Goto bb%d%s", t.Block, gotoVariantMark(t.Variant))
	case *If:
		return fmt.Sprintf("If %s then bb%d else bb%d fallthrough=bb%d",
			place(t.Test), t.Consequent, t.Alternate, t.Fallthrough)
	case *Branch:
		return fmt.Sprintf("Branch %s then bb%d else bb%d fallthrough=bb%d",
			place(t.Test), t.Consequent, t.Alternate, t.Fallthrough)
	case *Switch:
		parts := make([]string, 0, len(t.Cases))
		for _, kase := range t.Cases {
			if kase.Test == nil {
				parts = append(parts, fmt.Sprintf("default -> bb%d", kase.Block))
				continue
			}
			parts = append(parts, fmt.Sprintf("%s -> bb%d", place(*kase.Test), kase.Block))
		}
		return fmt.Sprintf("Switch %s [%s] fallthrough=bb%d", place(t.Test), strings.Join(parts, ", "), t.Fallthrough)
	case *While:
		return fmt.Sprintf("While test=bb%d loop=bb%d fallthrough=bb%d", t.Test, t.Loop, t.Fallthrough)
	case *DoWhile:
		return fmt.Sprintf("DoWhile loop=bb%d test=bb%d fallthrough=bb%d", t.Loop, t.Test, t.Fallthrough)
	case *For:
		update := "none"
		if HasBlock(t.Update) {
			update = fmt.Sprintf("bb%d", t.Update)
		}
		return fmt.Sprintf("For init=bb%d test=bb%d update=%s loop=bb%d fallthrough=bb%d",
			t.Init, t.Test, update, t.Loop, t.Fallthrough)
	case *ForOf:
		return fmt.Sprintf("ForOf init=bb%d test=bb%d loop=bb%d fallthrough=bb%d",
			t.Init, t.Test, t.Loop, t.Fallthrough)
	case *ForIn:
		return fmt.Sprintf("ForIn init=bb%d loop=bb%d fallthrough=bb%d", t.Init, t.Loop, t.Fallthrough)
	case *Logical:
		return fmt.Sprintf("Logical %s test=bb%d fallthrough=bb%d", t.Operator, t.Test, t.Fallthrough)
	case *Ternary:
		return fmt.Sprintf("Ternary test=bb%d fallthrough=bb%d", t.Test, t.Fallthrough)
	case *Optional:
		return fmt.Sprintf("Optional optional=%t test=bb%d fallthrough=bb%d", t.Optional, t.Test, t.Fallthrough)
	case *Sequence:
		return fmt.Sprintf("Sequence block=bb%d fallthrough=bb%d", t.Block, t.Fallthrough)
	case *Label:
		return fmt.Sprintf("Label block=bb%d fallthrough=bb%d", t.Block, t.Fallthrough)
	case *Try:
		binding := "none"
		if t.HandlerBinding != nil {
			binding = place(*t.HandlerBinding)
		}
		return fmt.Sprintf("Try block=bb%d handler=bb%d binding=%s fallthrough=bb%d",
			t.Block, t.Handler, binding, t.Fallthrough)
	case *MaybeThrow:
		handler := "none"
		if HasBlock(t.Handler) {
			handler = fmt.Sprintf("bb%d", t.Handler)
		}
		return fmt.Sprintf("MaybeThrow continuation=bb%d handler=%s", t.Continuation, handler)
	case *Scope:
		return fmt.Sprintf("Scope @%d block=bb%d fallthrough=bb%d", t.Scope, t.Block, t.Fallthrough)
	}
	return "<unknown terminal>"
}

func gotoVariantMark(variant GotoVariant) string {
	switch variant {
	case GotoVariantBreak:
		return " (break)"
	case GotoVariantContinue:
		return " (continue)"
	case GotoVariantTry:
		return " (try)"
	}
	return ""
}

// printManualMemoDependency renders one written dependency as its source path.
//
// The `?.` is printed for an optional entry because that character is the entire distinction the
// `preserve-manual-memoization` rule exists to report, and a printer that dropped it would make the
// two sides of that comparison look identical in every golden.
func printManualMemoDependency(dep ManualMemoDependency, place func(Place) string) string {
	var builder strings.Builder
	if dep.Root.IsGlobal {
		builder.WriteString(dep.Root.Name)
	} else {
		builder.WriteString(place(dep.Root.Place))
	}
	for _, entry := range dep.Path {
		if entry.Optional {
			builder.WriteString("?.")
		} else {
			builder.WriteString(".")
		}
		builder.WriteString(entry.Property)
	}
	return builder.String()
}
