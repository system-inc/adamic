package lower

import (
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestScannerStringConstructorView(t *testing.T) {
	program, err := lowerSource(t, `const worker: (point: number) => string = (String as any).fromCodePoint ? point => (String as any).fromCodePoint(point) : point => String.fromCharCode(point); console.log(worker(65));`)
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	inspect := func(node any) bool {
		if property, ok := node.(ir.Property); ok && property.Name == "fromCodePoint" && property.View != "" && property.ViewContract != 0 {
			contract := program.ViewContracts[property.ViewContract-1]
			if !contract.ProducerCertified || len(contract.Functions) != 1 || len(contract.Parameters) != 1 || program.ViewContracts[contract.Result-1].Of != ir.String {
				t.Fatal("lost intrinsic callable certificate")
			}
			checked++
		}
		return true
	}
	walk(program.Main, inspect)
	for _, function := range program.Functions {
		walk(function.Body, inspect)
	}
	if checked != 2 {
		t.Fatalf("want two checked member reads, found %d", checked)
	}
}

func TestScannerStringConstructorViewScope(t *testing.T) {
	for _, source := range []string{
		`const encoder = String as any;`,
		`const encode = (String as any).fromCodePoint;`,
		`console.log((String as any).fromCodePoint("65"));`,
		`console.log((String as any).fromCodePoint(65, 66));`,
		`const values = [65]; console.log((String as any).fromCodePoint(...values));`,
		`const own = { fromCodePoint: (point: number) => "own" }; console.log((own as any).fromCodePoint(65));`,
		`function make(): string { const convert = String; return (String as any).fromCodePoint(65); } console.log(make());`,
		`function make(): string { String.fromCodePoint = (point: number) => "changed"; return (String as any).fromCodePoint(65); } console.log(make());`,
		`function make(): string { const host = String; return (String as any).fromCodePoint(65); } console.log(make());`,
		`const value = (String as any).fromCodePoint === (String as any).fromCodePoint;`,
		`const unrelated = { anyResult: (point: number): any => point }; const worker: (point: number) => string = point => unrelated.anyResult(point);`,
	} {
		_, err := lowerSource(t, source)
		if err == nil {
			t.Fatalf("out-of-scope constructor/any use admitted: %s", source)
		}
		if !strings.Contains(err.Error(), "cast") && !strings.Contains(err.Error(), "String") && !strings.Contains(err.Error(), "function returning any") {
			t.Fatalf("unexpected frontier: %v", err)
		}
	}
}
