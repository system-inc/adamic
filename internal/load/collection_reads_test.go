package load

import (
	"strings"
	"testing"
)

func TestCollectionReadPayloadErrorsRemain(t *testing.T) {
	for _, source := range []string{
		"function f(map:Map<string,string>):number {return map.get('x');}",
		"function f(map:Map<string,number>):string {return map.get('x');}",
		"function f(map:Map<string,number|null>):number|null {return map.get('x');}",
		"function f(object:{get:(k:string)=>number|undefined}):number {return object.get('x');}",
		"function f(map:Map<string,number>):number {let value=map.get('x');return value;}",
	} {
		t.Run(source, func(t *testing.T) {
			diagnostics := checkErrors(t, writeProgram(t, [2]string{"main.a", source}))
			if !strings.Contains(strings.Join(diagnostics, "\n"), "TS2322") {
				t.Fatalf("wrong refusal: %v", diagnostics)
			}
		})
	}
}
func TestCollectionReadObligations(t *testing.T) {
	for _, source := range []string{
		"function f(array:readonly number[],i:number):number {return array[i];}",
		"function f(map:Map<string,number>):number {return map.get('x');}",
		"function f(map:Map<string,number>):number {const value=map.get('x');return value;}",
		"const map=new Map<string,number>();const value:number=map.get('x');",
		"function f(n:number):void{};const map=new Map<string,number>();f(map.get('x'));",
	} {
		t.Run(source, func(t *testing.T) {
			program, err := Load(writeProgram(t, [2]string{"main.a", source}))
			if err != nil {
				t.Fatal(err)
			}
			if len(program.collectionReads) != 1 {
				t.Fatalf("want one read obligation, got %d", len(program.collectionReads))
			}
		})
	}
}
