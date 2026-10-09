package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestObjectRefusalsExplainSoundness(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ source, reason string }{
		{`Object.defineProperty({value:1}, 'value', {value:'wrong'});`, "property descriptors"},
		{`Object.defineProperties({value:1}, {value:{value:'wrong'}});`, "property descriptors"},
		{`function descriptor(value:{value:number}) { Object.getOwnPropertyDescriptor(value, 'value'); }`, "property descriptors"},
		{`function descriptors(value:{value:number}) { Object.getOwnPropertyDescriptors(value); }`, "property descriptors"},
		{`Object.setPrototypeOf({value:1}, {});`, "prototypes"},
		{`Object.fromEntries([['key', 1]]);`, "index-signature"},
		{`Object.assign({value:1}, {value:'wrong'});`, "intersection result"},
		{`Object.assign({value:1}, {value:2}, {value:3}, {value:4}, {value:5});`, "result is any"},
		{`Object.assign({value:1});`, "result is any"},
		{`const source={value:1, hidden:'wrong'}; const view:{readonly value:number}=source; Object.assign({value:1},view);`, "widened source"},
		{`Object.values({number:1, text:'wrong'});`, "homogeneous"},
		{`Object.entries({number:1, text:'wrong'});`, "homogeneous"},
	} {
		t.Run(probe.reason+probe.source, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var refusal *Refused
			if !errors.As(err, &refusal) || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("got %v, want refusal with %q", err, probe.reason)
			}
		})
	}
}

func TestObjectUnprovenShapesStayNotYet(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`Object.assign({value:1}, {extra:2});`,
		`function keys(source:{value:number}|undefined):string[] { return Object.keys({...source}); }`,
		`function own(object:{value?:number}):boolean { return Object.hasOwn(object,'value'); }`,
		`Object.groupBy([1,2], (value:number):string => value===1 ? 'one' : 'other');`,
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, source)
			var notYet *NotYet
			if !errors.As(err, &notYet) {
				t.Fatalf("got %v, want NotYet", err)
			}
		})
	}
}

func TestObjectIntegrityRefusesUnrepresentedDescriptors(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`Object.seal([1]);`,
		`function seal(value: Map<string, number>): void { Object.seal(value); }`,
		`function seal(value: { n?: number }): void { Object.seal(value); }`,
		`function compare(left: number | null | undefined): boolean { return Object.is(left, undefined); }`,
		`const tuple: [number] = [1]; Object.isSealed(tuple);`,
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, source)
			var notYet *NotYet
			if !errors.As(err, &notYet) {
				t.Fatalf("got %v, want NotYet", err)
			}
		})
	}
}

func TestObjectReplacedBindingsUseChecks(t *testing.T) {
	t.Parallel()
	for _, write := range []string{
		`value = hidden;`,
		`[value] = [hidden];`,
		`({ value } = { value: hidden });`,
		`for (value of [hidden]) { }`,
	} {
		source := `let value = { n: 1 }; const hidden = { n: 2, extra: 'wrong' }; ` + write + ` Object.values(value);`
		t.Run(write, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSource(t, source)
			if write == `value = hidden;` || write == `[value] = [hidden];` {
				if err != nil {
					t.Fatal(err)
				}
				checked := false
				walk(program.Main, func(node any) bool {
					if call, ok := node.(ir.ObjectCall); ok && call.Method == "values" {
						checked = call.Checked
					}
					return true
				})
				if !checked {
					t.Fatal("replaced binding lost its enumeration check")
				}
				return
			}
			var notYet *NotYet
			if !errors.As(err, &notYet) {
				t.Fatalf("got %v, want NotYet", err)
			}
		})
	}
}

func TestFrozenPropertyErrorIsCatchable(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `const object={value:1}; Object.freeze(object); try { object.value=2; } catch { console.log('caught'); }`)
	if err != nil {
		t.Fatal(err)
	}
}
