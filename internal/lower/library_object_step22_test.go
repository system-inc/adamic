package lower

import (
	"errors"
	"testing"
)

func TestObjectStep22UnrepresentedCasesStayRefused(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`function tag(value: {readonly x:number}): string { return Object.prototype.toString.call(value); }`,
		`Object.prototype.toString.call(new Map<string, number>());`,
		`Object.prototype.hasOwnProperty.call(null, 'x');`,
		`Object.prototype.propertyIsEnumerable.call(undefined, 'x');`,
		`Object.prototype.hasOwnProperty.call({x:1}, Symbol.iterator);`,
		`const method = Object.prototype.toString;`,
		`Object.prototype.toString = ():string => 'changed'; console.log(Object.prototype.toString.call({}));`,
		`Number({valueOf: (): number|{x:number} => true ? 1 : {x:1}, toString: ():string => '2'});`,
		`Number({valueOf: (): {x:number} => ({x:1}), toString: (): {x:number} => ({x:2})});`,
		`Number({valueOf():number {return 1;}});`,
		`Number({valueOf: ():{}=>42, toString: ():string=>'wrong'});`,
		`Number({valueOf: (unused?:number):number => 1});`,
		`const hidden={x:1,toString:():string=>'hidden'}; const view:{readonly x:number}=hidden; Number(view);`,
	} {
		t.Run(source, func(t *testing.T) {
			_, err := lowerSource(t, source)
			if err == nil {
				t.Fatal("unrepresented behavior compiled")
			}
			var notYet *NotYet
			var refused *Refused
			if !errors.As(err, &notYet) && !errors.As(err, &refused) {
				t.Fatalf("want explicit refusal, got %v", err)
			}
		})
	}
}

func TestObjectStep22ProvenCasesLower(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`console.log(Object.prototype.toString.call(null));`,
		`console.log(Object.prototype.toString.call(undefined));`,
		`console.log(typeof Object.prototype.valueOf);`,
		"console.log(`${Object.prototype.toString.length}`);",
		`console.log(Object.prototype.toString.name);`,
		"console.log(`${Object.prototype.toString.prototype === undefined}`);",
		"console.log(`${Object.prototype.toString.hasOwnProperty('prototype')}`);",
		"console.log(`${Object.hasOwn({x:1}, 'missing')}`);",
		`console.log(Object.keys({10:10,2:2,b:3}).join('|'));`,
		`Number({valueOf: ():number=>1});`,
		`String({toString: ():string=>'text'});`,
	} {
		t.Run(source, func(t *testing.T) {
			if _, err := lowerSource(t, source); err != nil {
				t.Fatal(err)
			}
		})
	}
}
