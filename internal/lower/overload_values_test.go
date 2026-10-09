package lower

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

func checkOverloadValueStops(t *testing.T, selected string) {
	t.Helper()
	base := `interface Result<T>{readonly value:T;}
function make(value:string|number) {
 function evaluate(input:string):Result<string>;
 function evaluate(input:string|number):Result<string|number>;
 function evaluate(input:string|number):Result<string|number>{return {value};}
 return evaluate;
}
`
	for name, tail := range map[string]string{
		"mutable target":  `let alias=make('word'); alias=make(1); console.log(alias('text').value);`,
		"opaque storage":  `const box={evaluate:make('word')}; console.log(box.evaluate('text').value);`,
		"mixed parameter": `function invoke(visitor:(input:string)=>Result<string>){return visitor('text').value;} function other(input:string):Result<string>{return {value:input};} console.log(invoke(make('word'))); console.log(invoke(other));`,
	} {
		func() {
			if name != selected {
				return
			}
			directory := t.TempDir()
			path := filepath.Join(directory, "main.ts")
			if err := os.WriteFile(path, []byte(base+tail), 0600); err != nil {
				t.Fatal(err)
			}
			program, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			_, err = Lower(context.Background(), program)
			if err == nil || !strings.Contains(err.Error(), "indirect value of an overload") {
				t.Fatalf("expected a closed-target stop: %v", err)
			}
		}()
	}
}

func TestOverloadValueStopsMutableTarget(t *testing.T) {
	t.Parallel()
	checkOverloadValueStops(t, "mutable target")
}

func TestOverloadValueStopsOpaqueStorage(t *testing.T) {
	t.Parallel()
	checkOverloadValueStops(t, "opaque storage")
}

func TestOverloadValueStopsMixedParameter(t *testing.T) {
	t.Parallel()
	checkOverloadValueStops(t, "mixed parameter")
}
