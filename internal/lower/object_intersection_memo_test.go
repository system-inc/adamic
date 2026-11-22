package lower

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/load"
)

// objectIntersectionChain is a program whose types are a chain of plain shapes, each level's two fields
// holding the next level intersected with one of two plain shapes. The intersections a walk meets number
// two a level, but the paths to them double at each one, so a walk that settles nothing proves 2^levels
// fields (#vp9f4yd, tsc's jsDocCache census, which stalled at 1,860 of 4,064 units).
func objectIntersectionChain(levels int) string {
	var source strings.Builder
	source.WriteString("interface P { readonly p: number; }\ninterface Q { readonly q: string; }\n")
	for level := 0; level < levels; level++ {
		fmt.Fprintf(&source, "interface L%d { readonly a?: L%d & P; readonly b?: L%d & Q; readonly n: number; }\n", level, level+1, level+1)
	}
	fmt.Fprintf(&source, "interface L%d { readonly n: number; }\n", levels)
	source.WriteString("function count(x: L0 & P): number { return x.n + x.p; }\n")
	source.WriteString("console.log(String(count({ n: 1, p: 2 })));\n")
	return source.String()
}

// An intersection reached by many paths is proven once: objectIntersection keeps each settled answer, so a
// chain whose paths double at every level lowers in time (#vp9f4yd). Mutant: the kept answer never read,
// and the 40-level row runs past its bound.
func TestObjectIntersectionSettlesEachIntersectionOnce(t *testing.T) {
	t.Parallel()
	for _, levels := range []int{4, 40} {
		t.Run(fmt.Sprintf("levels_%d", levels), func(t *testing.T) {
			t.Parallel()
			program, err := load.Load([]string{writeAgreementSource(t, "main.ts", objectIntersectionChain(levels))})
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			done := make(chan error, 1)
			go func() {
				_, err := Lower(context.Background(), program)
				done <- err
			}()
			select {
			case err := <-done:
				t.Logf("%d levels: %v", levels, err)
			case <-time.After(60 * time.Second):
				t.Fatalf("%d levels: lowering still running after 60 s", levels)
			}
		})
	}
}
