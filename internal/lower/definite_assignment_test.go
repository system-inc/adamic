package lower

import (
	"github.com/system-inc/adamic/internal/ir"
	"strings"
	"testing"
)

func TestDefiniteAssignmentUsesReadiness(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct {
		source  string
		checked int
	}{
		{"class Box { n!: number; }\nconsole.log(`${new Box().n + 1}`);\n", 1},
		{"let n!: number;\nconsole.log(`${n + 1}`);\n", 1},
		{"class Box { n!: number; constructor() { this.n = 2; } }\nconsole.log(`${new Box().n + 1}`);\n", 1},
		{"let n!: number;\nn = 2;\nconsole.log(`${n + 1}`);\n", 0},
		// A call may rebind the box: the field written before swap() is unknown again after it.
		// Without that, the read skips its check and prints 0 where Node prints undefined.
		{
			"class Box { n!: number; }\nlet box = new Box();\nfunction swap(): void { box = new Box(); }\nbox.n = 3;\nswap();\nconsole.log(`${box.n}`);\n",
			1,
		},
	} {
		// Behavior cannot observe redundant readiness metadata after initialization.
		// Uninitialized rows deliberately panic where unchecked source produces NaN.
		if strings.Contains(probe.source, "= 2") {
			lowersAndAgreesWithNode(test, probe.source)
		}
		program, lowerSourceError := lowerSource(test, probe.source)
		if lowerSourceError != nil {
			test.Fatal(lowerSourceError)
		}
		checked := 0
		visit := func(node any) bool {
			switch read := node.(type) {
			case ir.Read:
				if read.Readiness != "" {
					checked++
				}
			case ir.Property:
				if read.Readiness != "" {
					checked++
				}
			}
			return true
		}
		walk(program.Main, visit)
		for _, function := range program.Functions {
			walk(function.Body, visit)
		}
		if checked != probe.checked {
			test.Fatalf("%d checked reads, want %d", checked, probe.checked)
		}
	}
}
