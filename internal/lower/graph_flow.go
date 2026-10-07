package lower

import (
	"reflect"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// graphFlows works backwards from graph views to the allocations that can supply
// them. Local assignments and function results are joined across all paths; this
// is a may-flow proof, not an execution or a choice of one conditional branch.
// Negative IDs name allocation sites in this IR, independently of fresh checker
// identities. They select the existing allocation emission without runtime work.
func (f *cycleFinder) graphFlows() {
	program := f.l.result
	next := 0
	var sites func(reflect.Value) reflect.Value
	sites = func(value reflect.Value) reflect.Value {
		switch value.Kind() {
		case reflect.Interface:
			if value.IsNil() {
				return value
			}
			mapped := sites(value.Elem())
			result := reflect.New(value.Type()).Elem()
			result.Set(mapped)
			return result
		case reflect.Struct:
			result := reflect.New(value.Type()).Elem()
			for i := 0; i < value.NumField(); i++ {
				result.Field(i).Set(sites(value.Field(i)))
			}
			if field := result.FieldByName("GraphTypes"); field.IsValid() && field.Type() == reflect.TypeOf([]int{}) {
				next--
				field.Set(reflect.ValueOf(append(append([]int{}, field.Interface().([]int)...), next)))
			}
			return result
		case reflect.Slice:
			if value.IsNil() {
				return value
			}
			result := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
			for i := 0; i < value.Len(); i++ {
				result.Index(i).Set(sites(value.Index(i)))
			}
			return result
		default:
			return value
		}
	}
	program.Main = sites(reflect.ValueOf(program.Main)).Interface().([]ir.Statement)
	for i := range program.Functions {
		program.Functions[i].Body = sites(reflect.ValueOf(program.Functions[i].Body)).Interface().([]ir.Statement)
	}

	var graphType func(*checker.Type) bool
	graphType = func(proven *checker.Type) bool {
		if proven == nil || f.weak(proven) {
			return false
		}
		if program.GraphTypes[int(proven.Id())] {
			return true
		}
		if proven.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
			for _, member := range proven.Types() {
				if graphType(member) {
					return true
				}
			}
		}
		return false
	}
	// Positive nodes are locals followed by function results.
	resultNode := func(function int) int { return len(program.Locals) + function + 1 }
	sources := map[int][]ir.Expression{}
	wanted := map[int]bool{}
	queue := []int{}
	demand := func(node int) {
		if !wanted[node] {
			wanted[node] = true
			queue = append(queue, node)
		}
	}
	for local, proven := range f.l.localTypes {
		if graphType(proven) {
			demand(local + 1)
		}
	}
	for local, types := range f.l.localAlso {
		for _, proven := range types {
			if graphType(proven) {
				demand(local + 1)
			}
		}
	}
	pending := []ir.Expression{}
	// Site retains the checker view of the actual slot being written.
	slot := func(site int, name string, argument int) bool {
		if site <= 0 || site > len(f.l.writeSites) {
			return false
		}
		holder := f.l.writeSites[site-1].holder
		if holder == nil {
			return false
		}
		if name != "" {
			for _, field := range f.fields(holder) {
				if f.l.cycleFieldMatches(holder, field.Name, name) {
					return graphType(f.l.checker.GetTypeOfSymbol(field))
				}
			}
			return false
		}
		args := f.l.checker.GetTypeArguments(holder)
		return argument < len(args) && graphType(args[argument])
	}
	// A counted aggregate can have graph-typed fields/elements. Seed those
	// initializer slots individually, rather than promoting its whole payload.
	returns := map[int]*checker.Type{}
	for symbol, index := range f.l.functions {
		signatures := f.l.checker.GetSignaturesOfType(f.l.checker.GetTypeOfSymbol(symbol), checker.SignatureKindCall)
		if len(signatures) > 0 {
			returns[index] = f.l.checker.GetReturnTypeOfSignature(signatures[0])
		}
	}
	for _, closure := range f.l.closureRecords {
		signatures := f.l.checker.GetSignaturesOfType(closure.proven, checker.SignatureKindCall)
		if len(signatures) > 0 {
			returns[closure.function] = f.l.checker.GetReturnTypeOfSignature(signatures[0])
		}
	}
	var members func(ir.Expression, *checker.Type)
	members = func(expression ir.Expression, proven *checker.Type) {
		if expression == nil || proven == nil || f.weak(proven) {
			return
		}
		if graphType(proven) {
			pending = append(pending, expression)
		}
		switch value := expression.(type) {
		case ir.ObjectLiteral:
			for _, field := range f.fields(proven) {
				for _, initializer := range value.Fields {
					if f.l.cycleFieldMatches(proven, field.Name, initializer.Name) {
						members(initializer.Value, f.l.checker.GetTypeOfSymbol(field))
					}
				}
			}
		case ir.ArrayLiteral:
			arguments := f.l.checker.GetTypeArguments(proven)
			if len(arguments) > 0 {
				for _, element := range value.Elements {
					members(element, arguments[0])
				}
			}
		case ir.Conditional:
			members(value.WhenTrue, proven)
			members(value.WhenNot, proven)
		}
	}
	collect := func(body []ir.Statement, function int) {
		walk(body, func(node any) bool {
			switch value := node.(type) {
			case ir.Declare:
				members(value.Value, f.l.localTypes[value.Local])
				sources[value.Local+1] = append(sources[value.Local+1], value.Value)
			case ir.Assign:
				members(value.Value, f.l.localTypes[value.Local])
				sources[value.Local+1] = append(sources[value.Local+1], value.Value)
			case ir.Return:
				members(value.Value, returns[function])
				if function >= 0 {
					sources[resultNode(function)] = append(sources[resultNode(function)], value.Value)
				}
			case ir.Call:
				for _, target := range program.CallTargets(value) {
					if target < 0 || target >= len(program.Functions) {
						continue
					}
					for i, param := range program.Functions[target].Parameters {
						if i < len(value.Arguments) {
							sources[param+1] = append(sources[param+1], value.Arguments[i])
						}
					}
				}
			case ir.CallClosure:
				targets := program.ClosureTargets(value)
				if targets.Unknown {
					// Opaque function values are a documented leak-only frontier.
					break
				}
				for _, target := range targets.Functions {
					for i, param := range program.Functions[target].Parameters {
						if i < len(value.Arguments) {
							sources[param+1] = append(sources[param+1], value.Arguments[i])
						}
					}
				}
			case ir.SetProperty:
				if slot(value.Site, value.Name, 0) {
					pending = append(pending, value.Value)
				}
			case ir.SetIndex:
				if slot(value.Site, "", 0) {
					pending = append(pending, value.Value)
				}
			case ir.ArrayPush:
				if slot(value.Site, "", 0) {
					pending = append(pending, value.Value)
				}
			case ir.MapSet:
				if slot(value.Site, "", 0) {
					pending = append(pending, value.Key)
				}
				if slot(value.Site, "", 1) {
					pending = append(pending, value.Value)
				}
			case ir.SetAdd:
				if slot(value.Site, "", 0) {
					pending = append(pending, value.Value)
				}
			}
			return true
		})
	}
	collect(program.Main, -1)
	for i, function := range program.Functions {
		collect(function.Body, i)
	}
	var follow func(ir.Expression)
	follow = func(expression ir.Expression) {
		if expression == nil {
			return
		}
		value := reflect.ValueOf(expression)
		if field := value.FieldByName("GraphTypes"); field.IsValid() {
			ids := field.Interface().([]int)
			program.GraphTypes[ids[len(ids)-1]] = true
			return
		}
		switch value := expression.(type) {
		case ir.Read:
			demand(value.Local + 1)
		case ir.Call:
			for _, target := range program.CallTargets(value) {
				demand(resultNode(target))
			}
		case ir.CallClosure:
			targets := program.ClosureTargets(value)
			if targets.Unknown {
				return
			}
			for _, target := range targets.Functions {
				demand(resultNode(target))
			}
		case ir.Conditional:
			follow(value.WhenTrue)
			follow(value.WhenNot)
		case ir.Box:
			follow(value.Value)
		case ir.Narrow:
			follow(value.Value)
		case ir.Unwrap:
			follow(value.Value)
		case ir.CheckedCast:
			follow(value.Value)
		}
	}
	for _, expression := range pending {
		follow(expression)
	}
	for len(queue) > 0 {
		node := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		for _, expression := range sources[node] {
			follow(expression)
		}
	}
}
