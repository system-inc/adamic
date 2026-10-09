package high_level_intermediate_representation

import (
	"github.com/system-inc/cohere/mutation_aliasing"
	"strings"
	"testing"
)

func TestManualMemoizationPreservesPrimitivePropertyCapture(t *testing.T) {
	t.Parallel()
	const source = `import {useCallback} from 'react'; function Component(props) {
const theme=mergeTheme(props.theme);
const coefficient=theme.configuration.coefficient;
const callback=useCallback(value=>[value*coefficient],[coefficient]);
mutate(theme.colors);
return <div onClick={callback}/>;
}`
	for _, testCase := range []struct {
		name, source string
		findings     int
	}{
		{"arithmetic", source, 0},
		{"nested capture", strings.Replace(source, "value=>[value*coefficient]", "value=>()=>[value*coefficient]", 1), 0},
		{"no later mutation", strings.Replace(source, "mutate(theme.colors);", "", 1), 0},
		{"equality", strings.Replace(source, "value*coefficient", "value===coefficient", 1), 2},
		{"identity", strings.Replace(source, "value*coefficient", "value,coefficient", 1), 2},
		{"missing dependency", strings.Replace(source, "[coefficient]);", "[]);", 1), 1},
		{"constant inputs", strings.Replace(strings.Replace(source, "Component(props)", "Component()", 1), "mergeTheme(props.theme)", "mergeTheme()", 1), 1},
		{"previously called", strings.Replace(source, "const callback=", "coefficient(); const callback=", 1), 2},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			findings, lowered := findingsForSource(t, testCase.source)
			if !lowered || len(findings) != testCase.findings {
				t.Fatalf("lowered=%t findings=%v, want %d findings", lowered, findings, testCase.findings)
			}
		})
	}
}

func TestPrimitivePropertyConstraintsFollowOperators(t *testing.T) {
	t.Parallel()
	for _, operator := range []string{"+", "-", "/", "%", "*", "**", "&", "|", ">>", "<<", "^", ">", "<", ">=", "<=", ">>>", "===", "!==", "==", "!="} {
		t.Run(operator, func(t *testing.T) {
			t.Parallel()
			function, _ := rangesFor(t, "function ordinary() { const model=getModel(); const scalar=model.value; return scalar "+operator+" 2; }")
			want := operator != ">>>" && operator != "===" && operator != "!==" && operator != "==" && operator != "!="
			assertPrimitivePropertyRead(t, function, "value", want)
		})
	}
}

func TestPrimitivePropertyConstraintsRespectIdentity(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name, source, property string
		primitive              bool
	}{
		{"local aliases", "function ordinary() { const model=getModel(); const scalar=model.value; const alias=scalar; return alias*2; }", "value", true},
		{"constant capture", "function ordinary() { const model=getModel(); const scalar=model.value; return ()=>scalar*2; }", "value", true},
		{"two capture boundaries", "function ordinary() { const model=getModel(); const scalar=model.value; return ()=>()=>scalar*2; }", "value", true},
		{"mutable capture", "function ordinary() { const model=getModel(); let scalar=model.value; const callback=()=>scalar*2; scalar=other; return callback; }", "value", false},
		{"shadow parameter", "function ordinary() { const model=getModel(); const scalar=model.value; return [scalar, (scalar)=>scalar*2]; }", "value", false},
		{"unrelated nested value", "function ordinary() { const model=getModel(); const scalar=model.value; return [scalar, ()=>{const nested=getModel(); return nested.other*2;}]; }", "value", false},
		{"known function", "function ordinary() { const model=getModel(); const scalar=model.value; scalar(); return scalar*2; }", "value", false},
		{"known object", "function ordinary() { const model={value:{}}; return model.value*2; }", "value", false},
		{"component props", "function Component(props) { return props.value*2; }", "value", false},
		{"ref current", "function ordinary() { const model=getModel(); return model.current*2; }", "current", false},
		{"phi barrier", "function ordinary(flag) { const model=getModel(); const scalar=flag?model.value:other; return scalar*2; }", "value", false},
		{"template does not constrain", "function ordinary() { const model=getModel(); return `${model.value}`; }", "value", false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			function, _ := rangesFor(t, testCase.source)
			assertPrimitivePropertyRead(t, function, testCase.property, testCase.primitive)
		})
	}
	if got := inferPrimitivePropertyReads(nil); len(got) != 0 {
		t.Fatalf("nil function inferred properties: %v", got)
	}
}

func assertPrimitivePropertyRead(t *testing.T, function *Function, property string, want bool) {
	t.Helper()
	properties := inferPrimitivePropertyReads(function)
	effects := InferAliasingEffects(function)
	matched := 0
	for _, instruction := range function.Instructions {
		if instruction == nil {
			continue
		}
		load, ok := instruction.Value.(*PropertyLoad)
		if !ok || load.Property != property {
			continue
		}
		matched++
		if got := properties[instruction.LValue.Identifier]; got != want {
			t.Errorf("property %s primitive=%t, want %t", property, got, want)
		}
		list := effects.Get(instruction.Id)
		primitive := len(list) == 1 && list[0].Kind == mutation_aliasing.AliasingEffectCreate && list[0].Value == mutation_aliasing.EffectValuePrimitive
		if primitive != want {
			t.Errorf("property %s effects=%v, want primitive=%t", property, list, want)
		}
	}
	if matched != 1 {
		t.Fatalf("found %d property reads, want one", matched)
	}
}
