// Production Go listener registrations are the independent numeric-kind oracle.
package main

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/rule"
	core "github.com/system-inc/cohere/internal/lint/rules/core"
	"github.com/system-inc/cohere/internal/lint/rules/nexus"
	"github.com/system-inc/cohere/internal/lint/rules/react"
	"sort"
)

func main() {
	deny, err := core.DecodeIdDenylistOptions([]byte(`["bad"]`))
	if err != nil {
		panic(err)
	}
	match, err := core.DecodeIdMatchOptions([]byte(`["^good$"]`))
	if err != nil {
		panic(err)
	}
	globals, err := core.DecodeNoRestrictedGlobalsOptions([]byte(`["status"]`))
	if err != nil {
		panic(err)
	}
	subjects := []struct {
		r       rule.Rule
		options any
	}{
		{core.IdDenylist, deny}, {core.IdMatch, match}, {nexus.ConcurrencyNoCheckThenWrite, nil},
		{core.NoRestrictedGlobals, globals}, {core.NoSetterReturn, nil}, {core.NoShadowRestrictedNames, nil},
		{react.SetStateInEffect, nil}, {react.SetStateInRender, nil}, {react.StaticComponents, nil},
	}
	// No listener is invoked. A nonnil checker admits registrations guarded by its availability.
	ctx := rule.Context{TypeChecker: &checker.Checker{}}
	for _, s := range subjects {
		kinds := []int{}
		for kind := range s.r.Run(ctx, s.options) {
			kinds = append(kinds, int(kind))
		}
		sort.Ints(kinds)
		if len(kinds) == 0 {
			panic("vacuous listener registration")
		}
		fmt.Printf("%s\t", s.r.Name)
		for _, kind := range kinds {
			fmt.Printf("%d,", kind)
		}
		fmt.Println()
	}
}
