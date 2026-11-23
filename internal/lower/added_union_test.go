package lower

import (
	"strings"
	"testing"
)

// A union operand of string addition is written as + writes it: ToPrimitive with addition's default hint,
// then ToString (#e0cnaa9, rulings.md:200), each arm by #09dvwny's closed-world dispatch. Each row agrees
// with Node on the JavaScript backend and natively under the sanitizers. The witness is a class with both
// methods: + asks valueOf first and prints 5, where the string hint would ask toString and print both.
// Mutant: the helper's hint read as the string hint, and the witness prints both, and a class with valueOf
// alone prints [object Object] (J-kd0h6pv).
func TestAddedUnionAgrees(t *testing.T) {
	t.Parallel()
	const pickText = "function pick(flag: boolean): number | string { return flag ? 7 : 'x'; }\n"
	for _, probe := range []struct{ name, source, stdout string }{
		{"number or string", pickText + "console.log('' + pick(true));\nconsole.log('' + pick(false));\n", "7\nx\n"},
		{"string on the right", pickText + "console.log(pick(true) + '!');\nconsole.log(pick(false) + '!');\n", "7!\nx!\n"},
		{"boolean or number", "function pick(flag: boolean): boolean | number { return flag ? true : 3; }\nconsole.log('v=' + pick(true) + ',' + pick(false));\n", "v=true,3\n"},
		{"class with valueOf", "class Money { constructor(readonly cents: number) {} valueOf(): number { return this.cents; } }\nfunction pick(flag: boolean): Money | number { return flag ? new Money(5) : 3; }\nconsole.log('' + pick(true));\nconsole.log('' + pick(false));\n", "5\n3\n"},
		{"class with toString only", "class Tag { toString(): string { return 'tag'; } }\nfunction pick(flag: boolean): Tag | string { return flag ? new Tag() : 's'; }\nconsole.log('' + pick(true));\nconsole.log('' + pick(false));\n", "tag\ns\n"},
		{"plain object", "interface Plain { readonly n: number }\nfunction pick(flag: boolean): Plain | number { return flag ? { n: 1 } : 3; }\nconsole.log('' + pick(true));\n", "[object Object]\n"},
		{"default hint witness", "class Both { valueOf(): number { return 5; } toString(): string { return 'both'; } }\nfunction pick(flag: boolean): Both | string { return flag ? new Both() : 's'; }\nconst x = pick(true);\nconsole.log('' + x);\n", "5\n"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				program := lowersAndAgreesWithNodeAs(t, probe.source, extension)
				if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != probe.stdout {
					t.Fatalf("native: exit %d, stdout %q, stderr %q; want 0 and %q", code, stdout, stderr, probe.stdout)
				}
			})
		}
	}
}

// A union holding undefined or null, beside an object, keeps combine's NotYet: such a union may not tell
// its null from its undefined (logical_and_reference_null.a), as String's own road holds.
func TestAddedUnionWithANullableObjectStaysNotYet(t *testing.T) {
	t.Parallel()
	const source = "class Money { valueOf(): number { return 5; } }\nfunction pick(flag: boolean): Money | undefined { return flag ? new Money() : undefined; }\nconsole.log('' + pick(true));\n"
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, source, extension)
			if err == nil || !strings.Contains(err.Error(), "a BinaryExpression with a string and") {
				t.Fatalf("want combine's NotYet, got %v", err)
			}
		})
	}
}

// An arm with no conversion the helper knows keeps combine's NotYet at compile time, never a stop at run
// time: a Map, where Node prints [object Map], and a Date, whose own default hint is the string hint. The
// helper would stop on both at run time (J-kd0h6pv), but their types say so first. Node's Buffer, the same
// way, is TestNodeBufferRefusals' (library_node_buffer_test.go).
func TestAddedUnionUnknownArmsStayNotYet(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"map arm", "function pick(flag: boolean): Map<string, number> | string { return flag ? new Map<string, number>() : 's'; }\nconsole.log('' + pick(true));\n"},
		{"date arm", "function pick(flag: boolean): Date | string { return flag ? new Date(0) : 's'; }\nconsole.log(String(('' + pick(true)).length > 10));\n"},
		{"function arm", "function pick(flag: boolean): (() => number) | string { return flag ? () => 1 : 's'; }\nconsole.log('' + pick(true));\n"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSourceAs(t, probe.source, extension)
				if err == nil || !strings.Contains(err.Error(), "a BinaryExpression with a string and") {
					t.Fatalf("want combine's NotYet, got %v", err)
				}
			})
		}
	}
}
