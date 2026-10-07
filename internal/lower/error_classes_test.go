package lower

import (
	"strings"
	"testing"
)

func TestErrorBoundariesAreExplicit(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, reason string }{
		{"derived error static own", "class Derived extends Error {} const error=new Derived('x'); Object.hasOwn(error,'name');", "on Error"},
		{"derived error hasOwn", "class Derived extends Error {} const error=new Derived('x'); error.hasOwnProperty('name');", "on Error"},
		{"derived error enumerable", "class Derived extends Error {} const error=new Derived('x'); error.propertyIsEnumerable('message');", "on Error"},
		{"derived error locale", "class Derived extends Error {} const error=new Derived('x'); error.toLocaleString();", "on Error"},
		{"normalization expansion", "function normalized(text:string):string {return text.normalize();} try {console.log(normalized('x'));} catch(e) {console.log('caught');}", "native runtime failure cannot unwind"},
		{"inherited constructor value", "const ctor=new Error('x').constructor;", "adamic/no-prototype-reflection"},
		{"inherited method value", "const method=new Error('x').valueOf;", "method read as a value"},
		{"stack read", "console.log(new Error('x').stack ?? '');", "Error.stack"},
		{"stack write", "const e=new Error('x'); e.stack='made up';", "writing Error.stack"},
		{"mutable cause", "const e=new Error('x'); e.cause=e;", "assigning Error.cause after construction"},
		{"erased cause", "const first=new Error('x',{cause:1}); const cause=first.cause; const second=new Error('y',{cause});", "an erased unknown cause"},
		{"nonliteral options", "const options:ErrorOptions={cause:1}; const e=new Error('x',options);", "ErrorOptions not written as a literal"},
		{"spread", "const e=new Error('x'); const copy={...e};", "spreading an Error"},
		{"structural view", "const e=new Error('x'); const view:{message:string}=e;", "structural object view"},
		{"nested structural view", "const e={error:new Error('x')}; const view:{readonly error:{readonly message:string}}=e;", "without nominal ancestry"},
		{"fake error", "const fake={name:'Error',message:'fake',toString:()=>''}; const e:Error=fake; throw e;", "without nominal ancestry"},
		{"override", "class Custom extends Error { toString():string {return 'custom';} } const e=new Custom('x');", "an Error.toString override"},
		{"nullable generic receiver", "function text(value:{name:string}|undefined):string {return Error.prototype.toString.call(value);}", "a possibly undefined object"},
		{"optional generic field", "function text(value:{name?:string}):string {return Error.prototype.toString.call(value);}", "optional or erased field"},
		{"cause closes a cycle", "class Holder { error:Error|undefined=undefined; } function close(h:Holder):void {h.error=new Error('x',{cause:h});} close(new Holder());", "cycle reference counting can't free"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			if err == nil || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("want refusal containing %q, got %v", probe.reason, err)
			}
		})
	}
}
