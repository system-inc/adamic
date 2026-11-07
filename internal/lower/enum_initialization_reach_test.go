package lower

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestEnumInitializationReach(test *testing.T) {
	test.Parallel()
	for _, name := range []string{
		"native-enum-map",
		"native-call-before-enum",
		"safe-helper",
		"safe-constructor",
		"callable-selection",
		"call-graph-24",
		"unknown-after",
		"unknown-before",
		"unknown-callback",
		"unknown-property",
	} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			path := "../../stage3/fixtures/enum-init-reach/" + name + ".a"
			source := readEnumInitFixture(test, path)
			if _, lowerSourceError := lowerSource(test, source); lowerSourceError != nil {
				test.Fatal(lowerSourceError)
			}
		})
	}
	for _, name := range []string{
		"reaching-direct",
		"reaching-helper",
		"reaching-cycle",
		"reaching-constructor",
		"reaching-derived",
	} {
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			_, lowerSourceError := lowerSource(
				test,
				readEnumInitFixture(test, "../../stage3/fixtures/enum-init-reach/"+name+".a"),
			)
			var gap *NotYet
			if !errors.As(lowerSourceError, &gap) ||
				gap.What != "reading an enum before its runtime initialization; move the call after the enum declaration" ||
				!strings.Contains(gap.Where, ":2:") {
				test.Fatalf("pending enum read on line 2 must stay refused: %v", lowerSourceError)
			}
		})
	}
}

func TestEnumInitializationGraphMemo(test *testing.T) {
	test.Parallel()
	for _, depth := range []int{12, 24} {
		test.Run(fmt.Sprint(depth), func(test *testing.T) {
			test.Parallel()
			source := "function f0():number{return Pending.Zero;}\n"
			for level := 1; level <= depth; level++ {
				source += fmt.Sprintf("function f%d():number {return f%d()+f%d();}\n", level, level-1, level-1)
			}
			source += "namespace Debug {export const ready=true;} enum Pending {Zero}"
			graph, statements := namespaceGraphForTest(test, source)
			reach := graph.reach(statements[depth])
			if len(reach) != 1 {
				test.Fatalf("enum reach lost: %d", len(reach))
			}
			if graph.walks != depth+1 {
				test.Fatalf("nonlinear enum function walks: got %d, want %d", graph.walks, depth+1)
			}
			test.Logf("depth %d: %d function body expansions", depth, graph.walks)
			for _, function := range statements[:depth+1] {
				graph.reach(function)
			}
			if graph.walks != depth+1 {
				test.Fatalf("enum memo revisited bodies: %d", graph.walks)
			}
		})
	}
}

func readEnumInitFixture(test *testing.T, path string) string {
	test.Helper()
	data, readError := os.ReadFile(path)
	if readError != nil {
		test.Fatal(readError)
	}
	return string(data)
}

func TestEnumNamespaceSharedCycle(test *testing.T) {
	test.Parallel()
	graph, statements := namespaceGraphForTest(test, `const key:'Zero'='Zero';
 function one():boolean {two();return Pending[key]===0;}
 function two():boolean {one();return Debug.ready;}
 enum Pending {Zero}
 namespace Debug {export const ready=true;}`)
	for _, function := range statements[1:3] {
		reach := graph.reach(function)
		if len(reach) != 2 || reach[statements[3]] == nil || reach[statements[4]] == nil {
			test.Fatalf("one shared component must reach both enum and namespace: %v", reach)
		}
	}
	if graph.walks != 2 {
		test.Fatalf("shared component expanded %d bodies", graph.walks)
	}
}
