package lower

import "testing"

func TestRecordDivergenceSupportedNeighbors(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`const r:Record<string,number>={value:1};function unchanged():void{function unused():void{delete r["value"];}}if(r["value"]!==undefined){unchanged();console.log(String(r["value"]+1));}`,
		`const r:Record<string,number>={x:1}; function read(k:string):void{console.log(String(r[k]));} read("x");`,
		`const r:Record<string,number>={}; function has(k:string):void{console.log(String(k in r));} has("x");`,
		`const r:Record<string,number>={}; function set(k:string):void{r[k]=1;} set("x");`,
		`const r:Record<string,number>={}; const k:string="toString"; if(Object.hasOwn(r,k)){console.log(String(r[k]));}`,
		`const k="__proto__"; const r:Record<string,number>={[k]:1}; r[k]=2; console.log(String(r[k]));`,
		`const r:Record<string,number>={value:1};function drop():void{delete r["other"];}if(r["value"]!==undefined){drop();console.log(String(r["value"]+1));}`,
		`const r:Record<string,number>={value:1};function drop():void{delete r["value"];}const value=r["value"];if(value!==undefined){drop();console.log(String(value+1));}`,
		`const r:Record<string,number>={value:1};function drop():void{delete r["value"];}drop();const value=r["value"];if(typeof value==="number"){console.log(String(value+1));}`,
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Fatalf("supported neighbor refused: %s: %v", source, err)
		}
	}
}
