package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestGenericPropertyHasConcreteBodies(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `function make() { return { call: <T>(callback: () => T): T => callback() }; }
const scanner = make(); scanner.call<number>(() => 0); scanner.call<string>(() => 'x'); scanner.call<{readonly text:string}>(() => ({text:'x'}));`)
	if err != nil {
		t.Fatal(err)
	}
	returns := map[ir.Type]int{}
	for _, function := range program.Functions {
		if strings.HasPrefix(function.Name, "generic_closure_") {
			returns[function.Returns]++
		}
	}
	if returns[ir.Number] != 1 || returns[ir.String] != 1 || returns[ir.Object] != 1 || len(returns) != 3 {
		t.Fatalf("want number/string/object bodies, got %v", returns)
	}
}

func TestGenericPropertyUnsupportedUsesStayDiagnostic(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`const value = <T>(x:T):T => x; value(1); const concrete: (x:number)=>number = value; concrete(1);`,
		`const value = <T>(x:T):T => x; value(1); console.log(value.name);`,
		`const value = <T>(x:T):T => x; value(1); Object.keys(value);`,
		`const value = <T>(x:T):T => x; value(1); function forward<U>(x:U):void { value(x); } forward('x');`,
		`const value = <T>(x:T):T => x;`,
		`const call = <T>(x:T):T => x; call(1); call.hasOwnProperty('length');`,
		`const call = <T>(x:T):T => x; call(1); const object = call as {}; console.log(typeof object);`,
		`const call = <T>(x:T):T => x; call(1); const spread = {...call}; Object.keys(spread);`,
		`const call = <T>(x:T):T => x; call(1); const mixed: (typeof call) | string = call; console.log(typeof mixed);`,
		`function make() { return {call:<T>(callback:()=>T):T => callback()}; } make().call<string>(()=>'x'); function missing():ReturnType<typeof make>|undefined { return undefined; } missing()?.call<string>(()=>'x');`,
		`function make<U>(label:U):string { const call = <T>(callback:()=>T):T => callback(); return call<string>(()=>'x'); } make<number>(0);`,
	} {
		_, err := lowerSource(t, source)
		var notYet *NotYet
		var refused *Refused
		if !errors.As(err, &notYet) && !errors.As(err, &refused) {
			t.Fatalf("want a diagnostic, got %v for %s", err, source)
		}
	}
}
