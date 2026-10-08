package javascript

import "testing"

func TestViewMixedUnionUnknownAndUnavailable(t *testing.T) {
	t.Parallel()
	for _, sample := range []struct{ name, source, declared, found string }{
		{"unknown", `adamicViewMixedUnionSelect({kind:'unknown',value:42},[{kind:'number'}],undefined,'view.value','number | string');`, "number | string", "unsupported representation"},
		{"missing adapter", `adamicViewMixedUnionSelect({kind:'object',value:{text:'ok'}},[{kind:'object',contract:1}],undefined,'view.value','Left | Right');`, "Left | Right", "object"},
		{"missing contract", `adamicViewMixedUnionSelect({kind:'object',value:{text:'ok'}},[{kind:'object'}],()=>true,'view.value','Left | Right');`, "Left | Right", "object"},
		{"null is not undefined", `adamicViewMixedUnionSelect({kind:'null',value:null},[{kind:'undefined'}],undefined,'view.value','string | undefined');`, "string | undefined", "null"},
		{"literal false", `adamicViewMixedUnionSelect({kind:'boolean',value:true},[{kind:'boolean',literal:true,value:false},{kind:'string'}],undefined,'view.value','false | string');`, "false | string", "boolean"},
	} {
		t.Run(sample.name, func(t *testing.T) {
			runViewNode(t, viewTestRuntime+viewMixedUnionsRuntime+sample.source, "", "adamic: panic: cast failed: field read failed: view.value matches no member of "+sample.declared+"; expected "+sample.declared+", found "+sample.found+"\n", 70)
		})
	}
}
