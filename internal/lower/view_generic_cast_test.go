package lower

import (
	"strings"
	"testing"
)

func TestGenericCastInstantiationsRequireConcreteProof(t *testing.T) {
	program, err := lowerSource(t, `interface Node {readonly position:number}; interface Identifier extends Node {readonly text:string}; function select<T extends Node|undefined>(node:Node):T{return node as T}; const raw={position:0,text:"Ken"};const value=select<Identifier>(raw);console.log(value.text);`)
	if err != nil {
		t.Fatal(err)
	}
	if len(program.ViewOrigins) == 0 || !program.CheckedFields["text"] {
		t.Fatal("instantiated assertion lost its checked view")
	}
	_, err = lowerSource(t, `interface Node {readonly position:number}; function select<T>(node:Node):T{return node as T}; const value=select<string>({position:0});console.log(value);`)
	if err == nil || !strings.Contains(err.Error(), "cast") {
		t.Fatalf("an unrelated type argument must re-refuse: %v", err)
	}
}

func TestScalarIndexSnapshotRejectsUnprovedSources(t *testing.T) {
	for _, source := range []string{
		`function select(args:readonly unknown[]):number{return args[0] as number;}`,
		`function select(args:readonly any[]):number{return args[0] as number;}`,
		`interface Nested { readonly value: { readonly value: any } }; function select(args: readonly(number|Nested)[]):number{return args[0] as number;}`,
	} {
		_, err := lowerSource(t, source)
		if err == nil {
			t.Fatalf("unproved indexed source was admitted: %s", source)
		}
	}
}

func TestNullableWritableArrayConsumersStayRefused(t *testing.T) {
	for _, consumer := range []string{"const size=result.length;", "for (const value of result) { const copy=value; }", "const copy=result.slice();", "const [first]=result;"} {
		_, err := lowerSource(t, `interface Chain {readonly next:(number|string)[]|undefined}; function next<T>(chain:Chain):T[]{return chain.next as T[]}; const result=next<number>({next:undefined});`+consumer)
		if err == nil || !strings.Contains(err.Error(), "checked receiver") {
			t.Fatalf("consumer must refuse: %s: %v", consumer, err)
		}
	}
}
