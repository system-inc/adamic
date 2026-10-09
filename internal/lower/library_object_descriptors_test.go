package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestObjectDescriptorProofRefusals(t *testing.T) {
	t.Parallel()
	for _, test := range []struct{ source, reason string }{
		{`const full={n:1,hidden:'kept'};const view:{n:number}=full;Object.getOwnPropertyDescriptors(view);`, "complete plain data-property shape"},
		{`function read(value:{n:number}):void {Object.getOwnPropertyDescriptor(value,'n');}`, "complete plain data-property shape"},
		{`Object.defineProperty({n:1},'extra',{value:2});`, "change the fixed shape"},
		{`Object.defineProperty({n:1},'n',{value:'wrong'});`, "preserve the field"},
		{`Object.defineProperty({n:1},'n',{writable:false});`, "individual attribute changes"},
		{`Object.defineProperty({n:1},'n',{get:()=>1});`, "accessor descriptors"},
		{`const d={writable:true};d.writable=false;Object.defineProperty({n:1},'n',d);`, "invariant-mutable"},
		{`const all=Object.getOwnPropertyDescriptors({n:1});const missing=all.missing;`, "proven actual shape"},
		{`const all=Object.getOwnPropertyDescriptors({n:1});const missing=all['missing'];`, "proven actual shape"},
		{`const value={n:1};Error.captureStackTrace(value);Object.getOwnPropertyDescriptors(value);`, "captured stack"},
		{`const value={n:1};Error.captureStackTrace(value);Object.getOwnPropertyDescriptor(value,'stack');`, "accessor descriptor"},
		{`const all=Object.getOwnPropertyDescriptors({n:1});all.n.value=2;`, "tagged descriptor stores"},
		{`enum E {A,B};Object.getOwnPropertyDescriptors(E);`, "complete plain data-property shape"},
	} {
		t.Run(test.reason+test.source, func(t *testing.T) {
			_, err := lowerSource(t, test.source)
			var refusal *Refused
			if !errors.As(err, &refusal) || !strings.Contains(err.Error(), test.reason) {
				t.Fatalf("want refusal containing %q, got %v", test.reason, err)
			}
		})
	}
}
