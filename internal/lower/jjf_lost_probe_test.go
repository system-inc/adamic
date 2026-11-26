package lower

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// Probe for #jjf9ppe, never for main: where -2's not-yet went on the parameter and assignment roads.
func TestJjfLostProbe(t *testing.T) {
	for _, name := range []string{"parameter.ts", "assignment.ts", "const_binding.ts"} {
		source, err := os.ReadFile(filepath.Join("..", "oracle", "testdata", "placeholder_destructure", name))
		if err != nil {
			t.Fatal(err)
		}
		program, err := lowerSourceAs(t, string(source), ".ts")
		uses, unsetViews, unset := 0, 0, 0
		if program != nil {
			count := func(node any) bool {
				switch value := node.(type) {
				case ir.PlaceholderUse:
					uses++
				case ir.Property:
					if value.Unset {
						unset++
						if value.View != "" {
							unsetViews++
						}
					}
				}
				return true
			}
			walk(program.Main, count)
			for _, function := range program.Functions {
				walk(function.Body, count)
			}
		}
		fmt.Printf("PROBE %s err=%v uses=%d unset=%d unsetViews=%d\n", name, err, uses, unset, unsetViews)
		if program != nil {
			for _, check := range program.PlaceholderChecks {
				fmt.Printf("PROBE %s check %+v\n", name, check)
			}
		}
	}
}
