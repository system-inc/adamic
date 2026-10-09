package lower

import (
	"strings"
	"testing"
)

func TestLibraryFindLastMissingScalarBoundaries(t *testing.T) {
	for _, source := range []string{
		`const a=[1,2]; a.findLast((n:number):boolean=>{a.pop();return false;});`,
		`const a=[1,2]; a.findLastIndex((n:number):boolean=>{a.length=0;return false;});`,
		`const a=[1,2]; const other={valueOf:():number=>{a.pop();return 1;}}; a.findLast((n:number):boolean=>{const coerced=+other;return coerced===n;});`,
		`const a=[1,2]; const other={push:():number=>{a.pop();return 1;}}; a.findLastIndex((n:number):boolean=>{other.push();return false;});`,
		`const a=[1,2]; class Shrink { get n():number {a.pop();return 1;} } const other=new Shrink(); a.findLast((n:number):boolean=>n===other.n);`,
		`const a=[1,2]; function callback(n:number):boolean {a.pop();return false;} a.findLast(callback);`,
	} {
		_, err := lowerSource(t, source)
		if err == nil || !strings.Contains(err.Error(), "scalar element excludes undefined") {
			t.Fatalf("want scalar completion representation refusal, got %v", err)
		}
	}
	for _, source := range []string{
		`const a=[1,2]; a.findLast((n:number):boolean=>n===2);`,
		`const a=[1,2]; a.findLastIndex((n:number,i:number):boolean=>{a[0]=4;a.push(5);console.log("visited");return n===4;});`,
		`const a=[1,2]; a.findLast((n:number|undefined):boolean=>{a.splice(0,2);return n===undefined;});`,
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Fatal(err)
		}
	}
	_, ignoredErr := lowerSource(t, `const a=[true,false]; a.findLastIndex(():boolean=>false);`)
	if ignoredErr == nil || !strings.Contains(ignoredErr.Error(), "compiler closure ABI support") {
		t.Fatalf("want ignored optional boolean ABI refusal, got %v", ignoredErr)
	}
	_, err := lowerSource(t, `const a=[true,false]; a.findLast((n:boolean|undefined):boolean=>{a.pop();return n===undefined;});`)
	if err == nil || !strings.Contains(err.Error(), "function value taking boolean | undefined") {
		t.Fatalf("want compiler optional boolean ABI refusal, got %v", err)
	}
}

func TestLibraryIteratorTagBoundaries(t *testing.T) {
	for _, source := range []string{
		`const a=[1].values(); Object.defineProperty(a,Symbol.toStringTag,{value:"custom"});`,
		`const a=[1].values(); const copy={...a};`,
		`const a=[1].values(); console.log(Object.getOwnPropertyNames(a).join(","));`,
		`function tag(value:{}):string {return value.toString();} console.log(tag([1].values()));`,
	} {
		if _, err := lowerSource(t, source); err == nil {
			t.Fatalf("unrepresented iterator observation compiled: %s", source)
		}
	}
}

func TestLibraryFindLastReferenceTagBoundaries(t *testing.T) {
	for _, source := range []string{
		`const a:(string|null)[]=["a",null]; a.findLastIndex((n:string|null):boolean=>{a.splice(0,2);console.log(n ?? "null");return false;});`,
		`const a=[{n:1}]; a.findLastIndex(({n}:{n:number}):boolean=>{a.splice(0,1);return n===1;});`,
		`const a=["a"]; function callback(n:string):boolean {a.splice(0,1);return false;} a.findLastIndex(callback);`,
	} {
		_, err := lowerSource(t, source)
		if err == nil || !strings.Contains(err.Error(), "compiler") {
			t.Fatalf("want compiler missing-payload refusal, got %v", err)
		}
	}
}
