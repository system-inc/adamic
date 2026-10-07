package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	checking "github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

type Wave01Expand struct {
	Mode                  int
	Inputs, Expected      []int
	Unions, Substitutions map[int][]int
}

func Wave01AwaitExpandCapture(ctx rule.Context, n *ast.Node) []Wave01Expand {
	pool := map[string]*checker.Type{}
	for _, statement := range ctx.SourceFile.Statements.Nodes {
		if statement.Kind == ast.KindTypeAliasDeclaration {
			pool[statement.Name().Text()] = ctx.TypeChecker.GetTypeAtLocation(statement.Type())
		}
	}
	first, second, parameter := pool["First"], pool["Second"], pool["Parameter"]
	if first == nil || second == nil || parameter == nil {
		panic("missing type pool")
	}
	parts := checking.UnionTypeParts(first)
	if len(parts) != 2 {
		panic("union must have two parts")
	}
	inputs := []*checker.Type{first, parameter, first}
	ids := func(types []*checker.Type) []int {
		out := make([]int, 0, len(types))
		for _, t := range types {
			out = append(out, int(t.Id()))
		}
		return out
	}
	unions := map[int][]int{}
	var record func(*checker.Type)
	record = func(t *checker.Type) {
		if _, seen := unions[int(t.Id())]; seen {
			return
		}
		members := checking.UnionTypeParts(t)
		unions[int(t.Id())] = ids(members)
		for _, member := range members {
			if member != t {
				record(member)
			}
		}
	}
	record(first)
	record(second)
	record(parameter)
	var captures []Wave01Expand
	for mode := 0; mode < 16; mode++ {
		substitutions := map[*checker.Type][]*checker.Type{}
		if mode&1 != 0 {
			substitutions[parameter] = []*checker.Type{second}
		}
		if mode&2 != 0 {
			substitutions[parts[0]] = []*checker.Type{parameter, first}
		}
		if mode&4 != 0 {
			substitutions[parameter] = []*checker.Type{}
		}
		if mode&8 != 0 {
			substitutions[parts[1]] = []*checker.Type{}
		}
		raw := map[int][]int{}
		for t, standIns := range substitutions {
			raw[int(t.Id())] = ids(standIns)
		}
		captures = append(captures, Wave01Expand{Mode: mode, Inputs: ids(inputs), Expected: ids(requireAwaitExpandTypes(inputs, substitutions)), Unions: unions, Substitutions: raw})
	}
	return captures
}
