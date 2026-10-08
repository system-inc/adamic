package javascript

import (
	"github.com/system-inc/adamic/internal/ir"
	"testing"
)

func TestViewCallableBoxingUnknownProducer(t *testing.T) {
	e := &emitter{program: &ir.Program{}}
	dispatch := e.viewCallableBoxedDispatch(ir.CallClosure{Returns: ir.Union})
	source := viewTestRuntime + "const foreign = new AdamicClosure((self, arguments_) => 7); console.log((" + dispatch + ")(foreign, []));"
	runViewNode(t, source, "", "adamic: panic: callable ABI adapter: unknown producer signature\n", 70)
	runViewNode(t, "const foreign = () => 7; console.log(foreign());", "7\n", "", 0)
}
