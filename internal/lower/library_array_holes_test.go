package lower

import (
	"strings"
	"testing"
)

func TestArrayHolesAliasRefusals(t *testing.T) {
	for _, source := range []string{
		`const holes=new Array<number>(5); const alias=holes; console.log(alias.slice().join(","));`,
		`const holes=new Array<number>(5); const box={values:holes}; console.log(box.values.reverse().join(","));`,
		`function make():number[]{return new Array<number>(5);} function use(values:number[]):string{return values.slice().join(",");} console.log(use(make()));`,
		`const holes=new Array<number>(5); console.log(String(holes.sort((a,b)=>a-b).length));`,
		`const holes=new Array(5); const words:string[]=holes; words[0]="word";`,
	} {
		_, err := lowerSource(t, source)
		if err == nil || (!strings.Contains(err.Error(), "holey") && !strings.Contains(err.Error(), "untyped")) {
			t.Fatalf("expected alias-safe refusal, got %v", err)
		}
	}
}

func TestArrayHolesNullableSearchRefusal(t *testing.T) {
	for _, source := range []string{
		`const cells=new Array<string>(3); const widened:(string|null|undefined)[]=cells; console.log(String(widened.includes(null)));`,
		`const cells=new Array<string|null|undefined>(3); console.log(String(cells.includes(null)));`,
	} {
		_, err := lowerSource(t, source)
		if err == nil || !strings.Contains(err.Error(), "null") {
			t.Fatalf("want nullable slot refusal, got %v", err)
		}
	}
}
