package lower

import (
	"github.com/system-inc/adamic/internal/ir"
	"reflect"
)

func generatorHasYield(value any) bool {
	found := false
	var visit func(reflect.Value)
	visit = func(v reflect.Value) {
		if !v.IsValid() {
			return
		}
		if v.CanInterface() {
			if _, yes := v.Interface().(ir.GeneratorYield); yes {
				found = true
				return
			}
		}
		switch v.Kind() {
		case reflect.Interface, reflect.Pointer:
			if !v.IsNil() {
				visit(v.Elem())
			}
		case reflect.Struct:
			for i := 0; i < v.NumField(); i++ {
				visit(v.Field(i))
			}
		case reflect.Slice:
			for i := 0; i < v.Len(); i++ {
				visit(v.Index(i))
			}
		}
	}
	visit(reflect.ValueOf(value))
	return found
}
func (g *generatorMachine) expression(value ir.Expression, ctx generatorContext, use func(ir.Expression) int) int {
	if yield, yes := value.(ir.GeneratorYield); yes {
		temporary := g.local(value.Type())
		continuation := use(g.slot(temporary))
		return g.yield(yield, continuation, ctx, &temporary)
	}
	if !generatorHasYield(value) {
		return use(g.rewrite(value).(ir.Expression))
	}
	if c, yes := value.(ir.Conditional); yes {
		local := g.local(c.Type())
		next := use(g.slot(local))
		store := func(v ir.Expression) int {
			return g.add(append([]ir.Statement{g.store(generatorSlot(local), v)}, g.jump(next)...), ctx)
		}
		yes, no := g.expression(c.WhenTrue, ctx, store), g.expression(c.WhenNot, ctx, store)
		return g.expression(c.Condition, ctx, func(v ir.Expression) int {
			return g.add([]ir.Statement{ir.If{Condition: v, Then: g.jump(yes), Else: g.jump(no)}}, ctx)
		})
	}
	if b, yes := value.(ir.Binary); yes && (b.Operator == ir.And || b.Operator == ir.Or) {
		g.reject("short-circuit yield")
		return use(value)
	}
	return g.operands(value, ctx, func(v any) int { return use(v.(ir.Expression)) })
}
func (g *generatorMachine) statementExpressions(value ir.Statement, ctx generatorContext, use func(ir.Statement) int) int {
	if !generatorHasYield(value) {
		return use(g.rewrite(value).(ir.Statement))
	}
	return g.operands(value, ctx, func(v any) int { return use(v.(ir.Statement)) })
}

// A containing operation resumes only after its operands have been evaluated in
// source order. Snapshots are owned frame slots, not native stack temporaries.
func (g *generatorMachine) operands(value any, ctx generatorContext, use func(any) int) int {
	copy := reflect.New(reflect.TypeOf(value)).Elem()
	copy.Set(reflect.ValueOf(value))
	type operand struct {
		at         reflect.Value
		expression ir.Expression
		local      int
	}
	operands := []operand{}
	var scan func(reflect.Value)
	scan = func(v reflect.Value) {
		if v.CanInterface() {
			if e, yes := v.Interface().(ir.Expression); yes && v.Kind() == reflect.Interface {
				local := g.local(e.Type())
				operands = append(operands, operand{v, e, local})
				v.Set(reflect.ValueOf(g.slot(local)))
				return
			}
		}
		switch v.Kind() {
		case reflect.Struct:
			for i := 0; i < v.NumField(); i++ {
				scan(v.Field(i))
			}
		case reflect.Slice:
			duplicate := reflect.MakeSlice(v.Type(), v.Len(), v.Len())
			reflect.Copy(duplicate, v)
			v.Set(duplicate)
			for i := 0; i < v.Len(); i++ {
				scan(v.Index(i))
			}
		}
	}
	for i := 0; i < copy.NumField(); i++ {
		scan(copy.Field(i))
	}
	next := use(copy.Interface())
	for i := len(operands) - 1; i >= 0; i-- {
		entry := operands[i]
		target := next
		next = g.expression(entry.expression, ctx, func(v ir.Expression) int {
			return g.add(append([]ir.Statement{g.store(generatorSlot(entry.local), v)}, g.jump(target)...), ctx)
		})
	}
	return next
}
func (g *generatorMachine) yield(y ir.GeneratorYield, next int, ctx generatorContext, local *int) int {
	if y.Delegate {
		return g.delegate(y, next, ctx, local)
	}
	resumed := next
	if local != nil {
		if !g.types.NextAllowsUndefined {
			if g.failure == nil {
				g.failure = &NotYet{Where: y.Where, What: "a generator yield result whose next argument is not proved present; use a next type admitting undefined"}
			}
		}
		resumed = g.add(append([]ir.Statement{g.store(generatorSlot(*local), ir.Narrow{Value: g.field("input", ir.Union), To: y.Next})}, g.jump(next)...), ctx)
	}
	returning := g.returned(g.field("input", ir.Union), ctx, 0)
	throwing := g.add([]ir.Statement{ir.Throw{Value: ir.Narrow{Value: g.field("input", ir.Union), To: ir.Object}}}, ctx)
	resume := g.add([]ir.Statement{ir.If{Condition: g.equals("mode", 1), Then: g.jump(returning)}, ir.If{Condition: g.equals("mode", 2), Then: g.jump(throwing)}}, ctx)
	g.states[resume].body = append(g.states[resume].body, g.jump(resumed)...)
	held := g.local(y.Value.Type())
	return g.expression(y.Value, ctx, func(v ir.Expression) int {
		return g.add([]ir.Statement{g.store(generatorSlot(held), v), g.store("pc", ir.NumberConstant{Value: float64(resume)}), g.store("status", ir.NumberConstant{Value: 2}), g.store("input", ir.Undefined{Of: ir.Union}), ir.Return{Value: g.packet(g.slot(held), false)}}, ctx)
	})
}
