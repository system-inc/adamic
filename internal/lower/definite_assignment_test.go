package lower

import (
	"github.com/system-inc/adamic/internal/ir"
	"testing"
)

func TestDefiniteAssignmentUsesReadiness(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		source  string
		checked int
	}{
		{"class Box { n!: number; }\nconsole.log(`${new Box().n + 1}`);\n", 1},
		{"let n!: number;\nconsole.log(`${n + 1}`);\n", 1},
		{"class Box { n!: number; constructor() { this.n = 2; } }\nconsole.log(`${new Box().n + 1}`);\n", 1},
		{"let n!: number;\nn = 2;\nconsole.log(`${n + 1}`);\n", 0},
	} {
		program, err := lowerSource(t, probe.source)
		if err != nil {
			t.Fatal(err)
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
			t.Fatalf("%d checked reads, want %d", checked, probe.checked)
		}
	}
}

func TestDefiniteAssignmentSoundNeighbors(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"class Box { n: number = 2; }\nconsole.log(`${new Box().n + 1}`);\n",
		"let n: number;\nn = 2;\nconsole.log(`${n + 1}`);\n",
		"class Box { n: number; constructor() { this.n = 2; } }\nconsole.log(`${new Box().n + 1}`);\n",
		"// This mentions ts-ignore in prose.\nconsole.log('@ts-ignore');\n",
	} {
		_, err := lowerSource(t, source)
		if err != nil {
			t.Errorf("sound neighbor: %v", err)
		}
	}
}
