package lower

import (
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestNeverFitsWithoutConversion(t *testing.T) {
	for _, to := range []ir.Type{ir.Number, ir.Boolean, ir.String, ir.Object, ir.Array, ir.Map, ir.Closure, ir.Union, ir.Weak, ir.MaybeNumber, ir.MaybeBoolean, ir.Uint8Array, ir.Int32Array, ir.Float64Array} {
		value := ir.Effects{Body: []ir.Statement{ir.Evaluate{Value: ir.Call{Function: 17}}}, Result: ir.Unreachable{}}
		fitted, ok := fit(value, to).(ir.Effects)
		if !ok {
			t.Fatalf("%v: converted a never value", to)
		}
		marker, ok := fitted.Result.(ir.Unreachable)
		if !ok || marker.Type() != to {
			t.Errorf("%v: want destination marker, got %#v", to, fitted.Result)
		}
		if len(fitted.Body) != 1 || !reflect.DeepEqual(fitted.Body[0], value.Body[0]) {
			t.Errorf("%v: lost call effects", to)
		}
	}
}
