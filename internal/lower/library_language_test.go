package lower

import (
	"strings"
	"testing"
)

func TestLibraryLanguageBoundaries(t *testing.T) {
	for _, test := range []struct{ name, source, reason string }{
		{"prototype", `const value = {__proto__: {inherited: 1}, own: 2}; for(const key in value) { console.log(key); }`, "refuses __proto__ in an object literal"},
		{"dynamic receiver", `const value = function(this: {x:number}): number { return this.x; };`, "this parameter"},
		{"constructor alias", `const value = Object; console.log(typeof value);`, "outside equality or typeof"},
		{"constructor passed", `function inspect(value: ObjectConstructor): void { console.log(typeof value); } inspect(Object);`, "outside equality or typeof"},
		{"inherited field", `const value = {}; const inherited = value.hasOwnProperty;`, "unbound-method"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := lowerSource(t, test.source)
			if err == nil || !strings.Contains(err.Error(), test.reason) {
				t.Fatalf("want %q refusal, got %v", test.reason, err)
			}
		})
	}
}

func TestForInRuntimeBoundaries(t *testing.T) {
	for _, source := range []string{
		`for(const key in [1,2]) { console.log(key); }`,
		`const value: {} = [1,2]; for(const key in value) { console.log(key); }`,
		`function keys(value: {}): void { for(const key in value) { console.log(key); } } keys({a:1});`,
		`const value: {x?: number} = {}; for(const key in value) { console.log(key); }`,
		`const source: {x?: number} = {}; const value = {...source, x:1}; for(const key in value) { console.log(key); }`,
		`let value: {} = {}; for(const key in value) { console.log(key); } value = [1,2];`,
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Fatal(err)
		}
	}
}
