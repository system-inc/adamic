package lower

import (
	"errors"
	"testing"
)

func TestPredicateContractRefusals(t *testing.T) {
	const shapes = `type A = {readonly kind: 'a'; readonly label: string}; type B = {readonly kind: 'b'; readonly label: string}; type N = A | B;`
	for name, source := range map[string]string{
		"lying helper":                  `function guard(n:N):n is A {return true;} function wrapper(n:N):n is A {return guard(n);}`,
		"recursive helper":              `function guard(n:N):n is A {return guard(n);}`,
		"lying callback":                `function guard(n:N):n is A {return true;} function cast(n:N,test:(n:N)=>n is A):A {if(test(n)){return n;} throw new Error('bad');} const n:N={kind:'a',label:'a'}; cast(n,guard);`,
		"reassigned callback parameter": `function guard(n:N):n is A {return n.kind==='a';} function cast(n:N,test:(n:N)=>n is A):A {test=guard;if(test(n)){return n;} throw new Error('bad');} const n:N={kind:'a',label:'a'}; cast(n,guard);`,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := lowerSource(t, shapes+source)
			var refusal *Refused
			if !errors.As(err, &refusal) {
				t.Fatalf("want predicate refusal, got %v", err)
			}
		})
	}
}
