package lower

import (
	"strings"
	"testing"
)

func TestJSONParseSupportedBoundaries(t *testing.T) {
	for _, source := range []string{
		`JSON.parse('anything');`,
		`const value:number=JSON.parse('1');console.log(String(value));`,
		`const value=JSON.parse('1') as number;console.log(String(value));`,
		`const value:unknown=JSON.parse('{"a":1}');console.log(typeof value);`,
		`const value=JSON.parse('{"a":1}');console.log(typeof value.a);`,
		`interface Info{version:string};const value:Info=JSON.parse('{"version":"ok","extra":1}');console.log(Object.keys(value).join(','));`,
		`function parse(text:string):number[]{return JSON.parse(text) as number[];}console.log(String(parse('[1]')[0]));`,
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Fatal(source, err)
		}
	}
}
func TestJSONParseRefusals(t *testing.T) {
	for _, one := range []struct{ source, reason string }{
		{`const text='{}';const value:{version:string}=JSON.parse(text);`, "complete key layout"},
		{`const text='{}';const value:unknown=JSON.parse(text);`, "complete key layout"},
		{`const value:{flag?:boolean}=JSON.parse('{}');`, "optional booleans"},
		{`const value=JSON.parse('1', (key,value)=>value);`, "revivers"},
		{`const value:{callback:()=>string}=JSON.parse('{}');`, "declared runtime representation"},
		{`function parse<T>(text:string):T{return JSON.parse(text) as T;}parse<number>('1');`, "per-instantiation contracts"},
		{`const value:{value:string}|{value:number}=JSON.parse('{"value":1}');`, "recursive variant selection"},
		{`const value:unknown=JSON.parse('{"\\u0000":1}');`, "constant document layout"},
		{`const value:unknown=JSON.parse('[1,"two"]');`, "heterogeneous array layouts"},
		{`const value:unknown=JSON.parse('[1]');if(typeof value==='object'&&value!==null&&'0' in value){console.log(typeof value['0']);}`, "ElementAccessExpression"},
		{`interface Info{version:string};const value=JSON.parse('{"version":"ok"}') as Info as unknown as {version:number};`, "cast"},
	} {
		_, err := lowerSource(t, one.source)
		if err == nil || !strings.Contains(err.Error(), one.reason) {
			t.Fatalf("%s: want %q refusal, got %v", one.source, one.reason, err)
		}
	}
}
