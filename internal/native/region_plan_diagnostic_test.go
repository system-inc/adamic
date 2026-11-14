package native

// A measurement for #ahnc64w, never landed: why bench-shaped graph.ts and copy.ts get no region. It
// prints the region plan's verdict on each function and statement of the programs ADAMIC_PLAN_SOURCES
// names, and the facts each verdict rests on.

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func TestRegionPlanDiagnostic(t *testing.T) {
	sources := os.Getenv("ADAMIC_PLAN_SOURCES")
	if sources == "" {
		t.Skip("ADAMIC_PLAN_SOURCES is unset")
	}
	for _, path := range strings.Fields(sources) {
		source, err := load.Load([]string{path})
		if err != nil {
			fmt.Printf("PLAN %s load: %v\n", path, err)
			continue
		}
		program, err := lower.Lower(context.Background(), source)
		if err != nil {
			fmt.Printf("PLAN %s lower: %v\n", path, err)
			continue
		}
		plan := planRegions(program)
		fmt.Printf("PLAN %s: %d functions, %d region statements, %d graph type ids\n", path, len(program.Functions), len(plan.statements), len(program.GraphTypes))
		for index, function := range program.Functions {
			if function.Name == "" || strings.HasPrefix(function.Name, "adamic") {
				continue
			}
			escapes := []string{}
			for position, parameter := range function.Parameters {
				local := program.Locals[parameter]
				escapes = append(escapes, fmt.Sprintf("p%d type=%d captured=%t assigned=%t escapes=%t", position, local.Type, local.Captured, local.ExpressionAssigned, plan.escapes[index][position]))
			}
			fmt.Printf("  fn %d %s returns=%d closure=%t rest=%d programRegion=%t fresh=%t | %s\n", index, function.Name, function.Returns, function.Closure, function.RestElement, function.ProgramRegion, plan.fresh[index], strings.Join(escapes, "; "))
			var statements func([]ir.Statement)
			statements = func(list []ir.Statement) {
				for _, statement := range list {
					if returned, ok := statement.(ir.Return); ok {
						fmt.Printf("    return %T fresh=%t%s\n", returned.Value, plan.freshValue(returned.Value), describe(program, plan, returned.Value))
					}
					walkStatement(statement, func(ir.Expression) {}, statements)
				}
			}
			statements(function.Body)
			describeStatements(program, plan, "    ", function.Body)
		}
		fmt.Printf("  main\n")
		describeStatements(program, plan, "    ", program.Main)
	}
}

func describe(program *ir.Program, plan *regionPlan, value ir.Expression) string {
	switch value := value.(type) {
	case ir.ObjectLiteral:
		return fmt.Sprintf(" literal programRegion=%t record=%t class=%d spread=%t graphTypes=%v isGraph=%t", value.ProgramRegion, value.Record, value.Class, value.Spread != nil, value.GraphTypes, program.IsGraph(value.GraphTypes))
	case ir.Call:
		return fmt.Sprintf(" call targets=%v virtual=%d expands=%t regionTarget=%d", program.CallTargets(value), value.Virtual, program.CallExpandsArguments(value), plan.regionTarget(value))
	}
	return ""
}

// describeStatements prints every statement holding a call whose argument is itself a call: the
// shape a statement region needs, with each half's verdict.
func describeStatements(program *ir.Program, plan *regionPlan, indent string, list []ir.Statement) {
	for index := range list {
		statement := list[index]
		walk(statement, func(expression ir.Expression) {
			call, ok := expression.(ir.Call)
			if !ok {
				return
			}
			for position, argument := range call.Arguments {
				if inner, isCall := argument.(ir.Call); isCall {
					fmt.Printf("%s%T region=%t: outer%s argument %d escapes=%t; inner%s\n", indent, statement, plan.statements[&list[index]], describe(program, plan, call), position, plan.callEscapes(call, position), describe(program, plan, inner))
				} else {
					fmt.Printf("%s%T outer%s argument %d is %T\n", indent, statement, describe(program, plan, call), position, argument)
				}
			}
		})
	}
}
