package lower

import (
	"github.com/system-inc/adamic/internal/ir"
	"testing"
)

func TestMapLikeRecordStores(t *testing.T) {
	program, err := lowerSource(t, `interface Options {named?:string;[key:string]:string} const o:Options={index:'dynamic'}; const pure:Record<string,string>={index:'dictionary'};`)
	if err != nil {
		t.Fatal(err)
	}
	var layouts [][]string
	for _, statement := range program.Main {
		if declaration, ok := statement.(ir.Declare); ok {
			if literal, ok := declaration.Value.(ir.RecordLiteral); ok {
				layouts = append(layouts, literal.Fixed)
			}
		}
	}
	if len(layouts) != 2 || len(layouts[0]) != 1 || layouts[0][0] != "named" || len(layouts[1]) != 0 {
		t.Fatalf("index keys must not enter declared shapes: %v", layouts)
	}
}

func TestMapLikeRecordViewBoundaries(t *testing.T) {
	for _, source := range []string{
		`const r:Record<string,string>={}; const view:{key?:string}=r; console.log(String(view.key));`,
		`function view(r:Record<string,string>):{key?:string} {return r;} view({});`,
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Fatal(err)
		}
	}
	for _, source := range []string{
		`const r:Record<string,number>={}; const view:{key?:number}=r;`,
		`const r:Record<string,string|number>={}; const view:{key?:string}=r;`,
		`interface R {readonly key?:string;[name:string]:string} const r:R={}; const view:{key?:string}=r;`,
		`interface R {key?:'narrow';[name:string]:string} const r:R={}; const view:{key?:string}=r;`,
		`type Named={key?:string}; const records:Record<string,string>[]=[{}]; const views:Named[]=records;`,
		`type Named={key?:string}; const r:Record<string,string>={}; const holder={r}; const view:{r:Named}=holder;`,
		`type Named={key?:string}; const f=(r:Record<string,string>):void=>{console.log(String(r['key']));}; const g:(n:Named)=>void=f;`,
	} {
		t.Run(source, func(t *testing.T) {
			if _, err := lowerSource(t, source); err == nil {
				t.Fatal("unproven storage or write relation accepted")
			}
		})
	}
}
