package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestDetachedOwnRefusals(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"let own = Object.prototype.hasOwnProperty; own.call({}, 'x');",
		"const own = Object.prototype.hasOwnProperty; own('x');",
		"const own = Object.prototype.hasOwnProperty; own.bind({});",
		"const own = Object.prototype.hasOwnProperty; console.log(typeof own);",
		"const own = Object.prototype.hasOwnProperty; const saved = {own};",
		"export const own = Object.prototype.hasOwnProperty;",
		"const own = Object.prototype.hasOwnProperty; const call = own.call;",
		"Object.prototype.hasOwnProperty('x');",
		"Object.prototype.hasOwnProperty?.call({}, 'x');",
		"Object.prototype?.hasOwnProperty.call({}, 'x');",
		"const own=Object.prototype.hasOwnProperty; own?.call({}, 'x');",
	} {
		t.Run(source, func(t *testing.T) {
			_, err := lowerSource(t, source)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), "hasOwnProperty") {
				t.Fatalf("want detached-method refusal, got %v", err)
			}
		})
	}
}

func TestDetachedOwnRepresentation(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"function own(map: object, key: string): boolean {return Object.prototype.hasOwnProperty.call(map,key);} console.log(String(own([], 'length')));",
		"function own(map: object, key: string): boolean {return Object.prototype.hasOwnProperty.call(map,key);} const r:Record<string,number>={}; console.log(String(own(r,'x')));",
		"function own(map: object, key: string): boolean {return Object.prototype.hasOwnProperty.call(map,key);} console.log(String(own(()=>1,'name')));",
		"const own=Object.prototype.hasOwnProperty; own.call(null,'x');",
		"function own(map: object, key: string): boolean {const view={map}; return Object.prototype.hasOwnProperty.call(map,key);}",
		"function own(map: object, key: string): boolean {const other=map; return Object.prototype.hasOwnProperty.call(map,key);}",
	} {
		t.Run(source, func(t *testing.T) {
			if _, err := lowerSource(t, source); err == nil {
				t.Fatal("an erased or unsupported storage representation must not compile")
			}
		})
	}
}

func TestDetachedOwnAreaAdapters(t *testing.T) {
	for _, source := range []string{
		"const own=Object.prototype.hasOwnProperty; console.log(`${own.apply({x:1},['x'])}`);",
		`const own=Object.prototype.hasOwnProperty; const alias=own; const r:Record<string,number>={}; console.log(String(alias.call(r,'x')));`,
		"const own=Object.prototype.hasOwnProperty; console.log(`${own.call([1],'length')}`);",
		`const values=Object.values; const fixed={x:1}; console.log(values(fixed).join(','));`,
	} {
		t.Run(source, func(t *testing.T) {
			if _, err := lowerSource(t, source); err != nil {
				t.Fatal(err)
			}
		})
	}
}
