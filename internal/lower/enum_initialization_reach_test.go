package lower

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestEnumInitializationReach(t *testing.T) {
	for _, name := range []string{"native-enum-map", "native-call-before-enum", "safe-helper", "safe-constructor", "callable-selection", "call-graph-24", "unknown-after", "unknown-before", "unknown-callback", "unknown-property"} {
		t.Run(name, func(t *testing.T) {
			path := "../../stage3/fixtures/enum-init-reach/" + name + ".a"
			source := readEnumInitFixture(t, path)
			if _, err := lowerSource(t, source); err != nil {
				t.Fatal(err)
			}
		})
	}
	for _, name := range []string{"reaching-direct", "reaching-helper", "reaching-cycle", "reaching-constructor", "reaching-derived"} {
		t.Run(name, func(t *testing.T) {
			_, err := lowerSource(t, readEnumInitFixture(t, "../../stage3/fixtures/enum-init-reach/"+name+".a"))
			var gap *NotYet
			if !errors.As(err, &gap) || gap.What != "reading an enum before its runtime initialization; move the call after the enum declaration" || !strings.Contains(gap.Where, ":2:") {
				t.Fatalf("pending enum read on line 2 must stay refused: %v", err)
			}
		})
	}
}

func TestEnumInitializationGraphMemo(t *testing.T) {
	for _, depth := range []int{12, 24} {
		t.Run(fmt.Sprint(depth), func(t *testing.T) {
			source := "function f0():number{return Pending.Zero;}\n"
			for level := 1; level <= depth; level++ {
				source += fmt.Sprintf("function f%d():number {return f%d()+f%d();}\n", level, level-1, level-1)
			}
			source += "namespace Debug {export const ready=true;} enum Pending {Zero}"
			graph, statements := namespaceGraphForTest(t, source)
			reach := graph.reach(statements[depth])
			if len(reach) != 1 {
				t.Fatalf("enum reach lost: %d", len(reach))
			}
			if graph.walks != depth+1 {
				t.Fatalf("nonlinear enum function walks: got %d, want %d", graph.walks, depth+1)
			}
			t.Logf("depth %d: %d function body expansions", depth, graph.walks)
			for _, function := range statements[:depth+1] {
				graph.reach(function)
			}
			if graph.walks != depth+1 {
				t.Fatalf("enum memo revisited bodies: %d", graph.walks)
			}
		})
	}
}

func readEnumInitFixture(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestEnumNamespaceSharedCycle(t *testing.T) {
	graph, statements := namespaceGraphForTest(t, `const key:'Zero'='Zero';
 function one():boolean {two();return Pending[key]===0;}
 function two():boolean {one();return Debug.ready;}
 enum Pending {Zero}
 namespace Debug {export const ready=true;}`)
	for _, function := range statements[1:3] {
		reach := graph.reach(function)
		if len(reach) != 2 || reach[statements[3]] == nil || reach[statements[4]] == nil {
			t.Fatalf("one shared component must reach both enum and namespace: %v", reach)
		}
	}
	if graph.walks != 2 {
		t.Fatalf("shared component expanded %d bodies", graph.walks)
	}
}
