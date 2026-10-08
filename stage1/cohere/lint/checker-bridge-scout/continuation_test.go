package main

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"testing"
)

func TestCheckerOwnedHandles(t *testing.T) {
	t.Parallel()
	first, second := &checker.Checker{}, &checker.Checker{}
	a, b := &checker.Type{}, &checker.Type{}
	left := ownerTable{Types: map[*checker.Type]uint64{}}
	right := ownerTable{Types: map[*checker.Type]uint64{}}
	x, y := left.intern(first, a), right.intern(second, b)
	if x.ID != y.ID {
		t.Fatal("fixture must create colliding local IDs")
	}
	if value, err := left.read(first, x); err != nil || value != a {
		t.Fatal("local round trip failed")
	}
	if _, err := left.read(first, y); err == nil {
		t.Fatal("foreign ID mutant survived")
	}
	left.Symbols = map[*ast.Symbol]uint64{}
	right.Symbols = map[*ast.Symbol]uint64{}
	sx := left.internSymbol(first, &ast.Symbol{})
	sy := right.internSymbol(second, &ast.Symbol{})
	if sx.ID != sy.ID {
		t.Fatal("symbol IDs must collide")
	}
	if _, e := left.readSymbol(first, sy); e == nil {
		t.Fatal("foreign symbol mutant survived")
	}
	x.ID = 99
	if _, err := left.read(first, x); err == nil {
		t.Fatal("stale ID mutant survived")
	}
}
