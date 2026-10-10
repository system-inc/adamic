package lower

import (
	"strings"
	"testing"
)

func TestLibraryLanguageBoundaries(t *testing.T) {
	t.Parallel()
	for _, test := range []struct{ name, source, reason string }{
		{"array", `for(const key in [1,2]) { console.log(key); }`, "for...in over an array"},
		{"hidden array", `const value: {} = [1,2]; for(const key in value) { console.log(key); }`, "plain-object origin"},
		{"parameter", `function keys(value: {}): void { for(const key in value) { console.log(key); } } keys({a:1});`, "plain-object origin"},
		{"prototype", `const value = {__proto__: {inherited: 1}, own: 2}; for(const key in value) { console.log(key); }`, "refuses __proto__ in an object literal"},
		{"optional object", `const value: {x?: number} = {}; for(const key in value) { console.log(key); }`, "plain-object origin"},
		{"optional spread", `const source: {x?: number} = {}; const value = {...source, x:1}; for(const key in value) { console.log(key); }`, "plain-object origin"},
		{"destructured write", `let value: {} = {}; for(const key in value) { console.log(key); } [value] = [[1,2]];`, "plain-object origin"},
		{"uncheckable dynamic receiver", `interface Receiver {readonly next:Receiver} const value = function(this:Receiver):number {return 1;};`, "this parameter without a runtime-checkable domain"},
		{"constructor alias", `const value = Object; console.log(typeof value);`, "outside equality or typeof"},
		{"constructor passed", `function inspect(value: ObjectConstructor): void { console.log(typeof value); } inspect(Object);`, "outside equality or typeof"},
		{"inherited field", `const value = {}; const inherited = value.hasOwnProperty;`, "unbound-method"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, test.source)
			if err == nil || !strings.Contains(err.Error(), test.reason) {
				t.Fatalf("want %q refusal, got %v", test.reason, err)
			}
		})
	}
}
