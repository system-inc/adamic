package lower

import (
	"errors"
	"github.com/system-inc/adamic/internal/ir"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScoutMapKeyOutcomes(t *testing.T) {
	for _, probe := range []struct {
		name, reason string
		refused      bool
	}{
		{"branded_map", "a value of type Path", false},
		{"branded_set", "a value of type Path", false},
		{"mixed_map", "a Map whose keys aren't", false},
		{"nullish_set", "a Set of null | undefined", false},
		{"maplike", "an index signature", true},
		{"reused_pair_mutation", "assigning an element of a value", false},
	} {
		t.Run(probe.name, func(t *testing.T) {
			source, err := os.ReadFile(filepath.Join("..", "..", "stage3", "map-keys", probe.name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowerSource(t, string(source))
			var refused *Refused
			var notYet *NotYet
			if (probe.refused && !errors.As(err, &refused)) || (!probe.refused && !errors.As(err, &notYet)) || err == nil || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("want %q with refused=%v, got %v", probe.reason, probe.refused, err)
			}
			t.Log(err)
		})
	}
}

func TestScoutMapIteratorPairsLower(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("..", "..", "stage3", "map-keys", "iterator_pairs.a"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = lowerSource(t, string(source)); err != nil {
		t.Fatal(err)
	}
}

func TestScoutMapIteratorPairRepresentationGap(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("..", "..", "stage3", "map-keys", "iterator_pairs.a"))
	if err != nil {
		t.Fatal(err)
	}
	widened := strings.Replace(string(source), "new Map(pairs())", "new Map<string, string | number>(pairs())", 1)
	_, err = lowerSource(t, widened)
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "new Map from pairs held otherwise than the Map's keys and values") {
		t.Fatalf("want representation-conversion gap, got %v", err)
	}
}

func TestScoutMapBrandBoundary(t *testing.T) {
	for _, input := range []string{"42", "undefined", "false", "null"} {
		t.Run(input, func(t *testing.T) {
			source := "type __String = (string & { __escapedIdentifier:void }) | (void & { __escapedIdentifier:void }) | '__call'; function f(x: unknown):__String { return x as __String; } console.log(typeof f(" + input + "));"
			program, err := lowerSource(t, source)
			if err != nil {
				t.Fatal(err)
			}
			checks := 0
			for _, function := range program.Functions {
				if function.Name == "checked_collection_brand" {
					for _, statement := range function.Body {
						if _, ok := statement.(ir.If); ok {
							checks++
						}
					}
				}
			}
			if checks != 1 {
				t.Fatalf("want one string boundary check, got %d", checks)
			}
		})
	}
}

func TestScoutMapBrandRefinementRefused(t *testing.T) {
	source := "type Path = string & { __pathBrand:void }; function f(text:string):Path & 'fixed' { return text as Path & 'fixed'; }"
	_, err := lowerSource(t, source)
	var refused *Refused
	if !errors.As(err, &refused) {
		t.Fatalf("want literal-refinement refusal, got %v", err)
	}
}

func TestScoutMapPresenceRules(t *testing.T) {
	for _, test := range []struct {
		name, body string
		checked    int
	}{
		{"early missing return", "if(!map.has(key)){return -1;}return map.get(key);", 0},
		{"early return then alias write", "const alias=map;if(!map.has(key))return -1;alias.clear();return map.get(key);", 1},
		{"same map/key", "if(map.has(key)){return map.get(key);}return -1;", 0},
		{"different key", "if(map.has(key)){return map.get('missing');}return -1;", 1},
		{"different map", "const other=new Map<string,number>();if(map.has(key)){return other.get(key);}return -1;", 1},
		{"alias write", "const alias=map;if(map.has(key)){alias.delete(key);return map.get(key);}return -1;", 1},
		{"unknown call", "function clear():void{map.clear();}if(map.has(key)){clear();return map.get(key);}return -1;", 1},
		{"key update", "if(map.has(key)){key='missing';return map.get(key);}return -1;", 1},
		{"optional stored undefined", "const optional=new Map<string,number|undefined>();if(optional.has(key)){return optional.get(key);}return -1;", 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			program, err := lowerSource(t, "function read(map:Map<string,number>,key:string):number{"+test.body+"}")
			if err != nil {
				t.Fatal(err)
			}
			checks := 0
			for _, function := range program.Functions {
				walk(function.Body, func(node any) bool {
					if c, ok := node.(ir.Coalesce); ok && c.Panic != nil {
						if text, ok := c.Panic.(ir.StringConstant); ok && strings.HasPrefix(program.Strings[text.Index], "collection lookup failed:") {
							checks++
						}
					}
					return true
				})
			}
			if checks != test.checked {
				t.Fatalf("want %d lookup checks, got %d", test.checked, checks)
			}
		})
	}
}

func TestScoutArrayRangeRules(t *testing.T) {
	for _, test := range []struct {
		name, body string
		checked    int
	}{
		{"range", "for(let i=0;i<array.length;i++){const value=array[i];take(value);}", 0},
		{"alias mutation before read", "const alias=array;for(let i=0;i<array.length;i++){alias.pop();const value=array[i];take(value);}", 1},
		{"index mutation after read", "for(let i=0;i<array.length;i++){const value=array[i];take(value);i=-1;}", 1},
		{"hoisted index", "for(var i=0;i<array.length;i++){const value=array[i];take(value);}", 1},
		{"negative start", "for(let i=-1;i<array.length;i++){const value=array[i];take(value);}", 1},
		{"different array", "const other:number[]=[];for(let i=0;i<array.length;i++){const value=other[i];take(value);}", 1},
		{"undefined payload", "const other:(number|undefined)[]=[undefined];for(let i=0;i<other.length;i++){const value=other[i];take(value);}", 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			source := "function take(value:number):void{} function run(array:number[]):void{" + test.body + "}"
			var program *ir.Program
			var err error
			if test.name == "hoisted index" {
				program, err = lowerTypeScriptSource(t, source)
			} else {
				program, err = lowerSource(t, source)
			}
			if err != nil {
				if test.name == "hoisted index" && strings.Contains(err.Error(), "refuses var") {
					return
				}
				t.Fatal(err)
			}
			checks := 0
			for _, function := range program.Functions {
				walk(function.Body, func(node any) bool {
					if value, ok := node.(ir.Coalesce); ok && value.Panic != nil {
						if text, ok := value.Panic.(ir.StringConstant); ok && strings.HasPrefix(program.Strings[text.Index], "collection lookup failed:") {
							checks++
						}
					}
					return true
				})
			}
			if checks != test.checked {
				t.Fatalf("want %d checked reads, got %d", test.checked, checks)
			}
		})
	}
}

func TestScoutMultiMapFactoryNativeGap(t *testing.T) {
	replacement, err := os.ReadFile(filepath.Join("..", "..", "stage3", "adapt", "76-multimap-composition", "replacement.a"))
	if err != nil {
		t.Fatal(err)
	}
	source := `interface MultiMap<K,V> extends Map<K,V[]> {add(key:K,value:V):V[];remove(key:K,value:V):void;}
 function unorderedRemoveItem<T>(array:T[],value:T):boolean {const index=array.indexOf(value);if(index<0)return false;const last:T=array[array.length-1];array[index]=last;array.pop();return true;}
 ` + string(replacement) + `const map=createMultiMap<string,number>();map.add('key',1);`
	_, err = lowerSource(t, source)
	var refused *Refused
	var gap *NotYet
	if !errors.As(err, &refused) && !errors.As(err, &gap) {
		t.Fatalf("want explicit factory dependency, got %v", err)
	}
	t.Log(err)
}

func TestScoutMultiMapExpandoStaysRefused(t *testing.T) {
	_, err := lowerSource(t, `interface MultiMap extends Map<string,number[]> {add(key:string,value:number):number[];} const map=new Map<string,number[]>() as MultiMap;map.add=(key,value)=>[value];`)
	var refused *Refused
	var gap *NotYet
	if !errors.As(err, &refused) && !errors.As(err, &gap) {
		t.Fatalf("want a loud expando refusal, got %v", err)
	}
	t.Log(err)
}

func TestScoutOriginalCustomSetOutcome(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("..", "..", "stage3", "map-keys", "custom_set.a"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowerSource(t, string(source))
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "generator") {
		t.Fatalf("want the current generator dependency, got %v", err)
	}
	t.Log(err)
}
