package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestUnknownReflectionRefusals(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ source, reason string }{
		{`function reflect(value: unknown): string { return typeof value; } function test(value: {code:string}|null): string { return reflect(value); }`, "nullable reference"},
		{`function has(key: string, value: object): boolean { return key in value; }`, "dynamic key coercion"},
		{`console.log(String('a\0b' in {}));`, "NUL key"},
		{`function has(pair: readonly [number]): boolean { return 'length' in pair; }`, "tuple"},
		{`function read(value: unknown): unknown { return typeof value === 'object' && value !== null && '__proto__' in value ? value.__proto__ : undefined; }`, "dynamic prototype"},
		{`function has(value: unknown): boolean { return typeof value === 'object' && value !== null && 'stack' in value; }`, "descriptor metadata"},
		{`function has(value: unknown): boolean { return typeof value === 'object' && value !== null && 'errno' in value; }`, "descriptor metadata"},
		{`function box(value: unknown): unknown { return value; } box(new Map<string, number>());`, "opaque host"},
		{`function box(value: unknown): unknown { return value; } box({code:null});`, "nullable field"},
		{`class Code { get code(): string { return 'x'; } } function read(value: unknown): unknown { return typeof value === 'object' && value !== null && 'code' in value ? value.code : undefined; } read(new Code());`, "dynamic getter"},
		{`class Code { code(): string { return 'x'; } } function read(value: unknown): unknown { return typeof value === 'object' && value !== null && 'code' in value ? value.code : undefined; } read(new Code());`, "dynamic getter or method"},
		{`function text(value: unknown): string { return String(value); }`, "ToPrimitive"},
		{`function read(value: unknown): unknown { return typeof value === 'object' && value !== null && 'code' in value ? value.code : undefined; } const nested = {child:{code:null}}; read(nested);`, "nullable field"},
		{`function box(value: unknown): unknown { return value; } box(() => 1);`, "function descriptors"},
		{`function box(value: unknown): unknown { return value; } box(JSON);`, "JSON as a value"},
		{`class Code { static code = 'x'; } console.log(String('code' in Code));`, "class constructor"},
		{`class Code { inspect(value: unknown): void {} } console.log(String('inspect' in new Code()));`, "dispatch table"},
		{`function has(value: {code:number}|undefined): boolean { return 'code' in {...value}; }`, "presence descriptors"},
		{`class Code { declare code: string; } function has(value: unknown): boolean { return typeof value === 'object' && value !== null && 'code' in value; } has(new Code());`, "ambient class field"},
	} {
		t.Run(probe.reason+probe.source, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("got %v, want NotYet explaining %q", err, probe.reason)
			}
		})
	}
}
