package oracle

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// Probe only, never for landing (#g4arkr4 step 1, round 3): why a map callback with no rest gets a rest
// tail packed for it. For each ArrayMap in the program it asks the emitter's own question, the layout
// of ir.CallClosure{Closure: Callback, FunctionType: CallbackType} (callbackCall, native/arguments_length.go),
// and beside it the two sets ClosureArgumentLayout chooses between: FunctionTypeTargets for the type,
// read first, and ClosureTargets for the callback value, read only when the type has no set.
func TestG4aProbe3(t *testing.T) {
	directory := "internal/oracle/testdata/g4a_probe3"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".ts" {
			continue
		}
		t.Run(entry.Name(), func(t *testing.T) {
			name := strings.TrimSuffix(entry.Name(), ".ts")
			absolute, err := filepath.Abs(filepath.Join(repository, directory, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, absolute)
			if err != nil {
				t.Logf("PROBE %s lower stops: %s", name, cut(strings.ReplaceAll(err.Error(), "\n", " / "), 400))
				return
			}
			for index, function := range program.Functions {
				t.Logf("PROBE %s function %d %q file=%s at %d:%d closure=%v receiver=%v parameters=%v rest=%d returns=%d", name, index, function.Name, filepath.Base(function.Position.File), function.Position.Line, function.Position.Column, function.Closure, function.Receiver, function.Parameters, function.RestElement, function.Returns)
			}
			t.Logf("PROBE %s fixedSlots=%d restSlots=%v", name, program.FixedArgumentSlots, program.RestArgumentSlots)
			var maps []ir.ArrayMap
			for _, function := range program.Functions {
				g4aArrayMaps(reflect.ValueOf(function.Body), &maps)
			}
			g4aArrayMaps(reflect.ValueOf(program.Main), &maps)
			if len(maps) == 0 {
				t.Logf("PROBE %s no ArrayMap in the IR", name)
			}
			for index, arrayMap := range maps {
				call := ir.CallClosure{Closure: arrayMap.Callback, FunctionType: arrayMap.CallbackType}
				layout := program.ClosureArgumentLayout(call)
				targets, resolved := program.FunctionTypeTargets[arrayMap.CallbackType]
				bounded := program.ClosureTargets(call)
				t.Logf("PROBE %s map %d callbackType=%d callback=%s elementCheck=%v", name, index, arrayMap.CallbackType, cut(irText(reflect.ValueOf(arrayMap.Callback)), 200), arrayMap.ElementCheck != nil)
				t.Logf("PROBE %s map %d layout fixed=%v rest=%v count=%v packedCount=%v", name, index, layout.Fixed, layout.Rest, layout.Count, program.PackedCountNeededFromCall(call))
				t.Logf("PROBE %s map %d FunctionTypeTargets[%d] resolved=%v targets=%v", name, index, arrayMap.CallbackType, resolved, targets)
				t.Logf("PROBE %s map %d ClosureTargets unknown=%v functions=%v", name, index, bounded.Unknown, bounded.Functions)
			}
		})
	}
}

// g4aArrayMaps collects every ir.ArrayMap reachable from value, an ArrayMap inside another's callback
// or array among them.
func g4aArrayMaps(value reflect.Value, maps *[]ir.ArrayMap) {
	switch value.Kind() {
	case reflect.Interface, reflect.Pointer:
		if !value.IsNil() {
			g4aArrayMaps(value.Elem(), maps)
		}
	case reflect.Struct:
		if value.CanInterface() {
			if arrayMap, ok := value.Interface().(ir.ArrayMap); ok {
				*maps = append(*maps, arrayMap)
			}
		}
		for index := range value.NumField() {
			if value.Type().Field(index).IsExported() {
				g4aArrayMaps(value.Field(index), maps)
			}
		}
	case reflect.Slice, reflect.Array:
		for index := range value.Len() {
			g4aArrayMaps(value.Index(index), maps)
		}
	}
}
