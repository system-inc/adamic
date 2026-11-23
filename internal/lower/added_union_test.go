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

// A class the helper would stop on keeps combine's NotYet at compile time (Language's must on #e0cnaa9, p02
// and p17): a valueOf getter, a valueOf field holding an arrow, a toString parameter property, a valueOf
// taking a parameter, and an object literal's valueOf or toString getter, which lowers as a class of its own
// (J-6y49hh0). The helper dispatches on the object's own class, so any such class in the program
// keeps every union road, an honest class's included: that's the price, and it fails closed. Mutant:
// programClassesConvert answering yes, and each lowers to the helper's run-time stop.
func TestAddedUnionStoppingClassesStayNotYet(t *testing.T) {
	t.Parallel()
	const tail = "function pick(flag: boolean): Late | string { return flag ? new Late() : 's'; }\nconsole.log(pick(true) + '!');\n"
	for _, probe := range []struct{ name, source string }{
		{"valueOf getter", "class Late { get valueOf(): () => number { return (): number => 7; } }\n" + tail},
		{"valueOf field", "class Late { readonly valueOf = (): number => 9; }\n" + tail},
		{"toString parameter property", "class Late { constructor(readonly toString: () => string = (): string => 'p') {} }\n" + tail},
		{"valueOf taking a parameter", "class Late { valueOf(base?: number): number { return (base ?? 0) + 1; } }\n" + tail},
		{"literal valueOf getter", "interface Late { readonly valueOf: () => number }\nfunction make(): Late { return { get valueOf(): () => number { return (): number => 5; } }; }\nfunction pick(flag: boolean): Late | string { return flag ? make() : 's'; }\nconsole.log(pick(true) + '!');\n"},
		{"literal toString getter", "interface Late { readonly toString: () => string }\nfunction make(): Late { return { get toString(): () => string { return (): string => 'g'; } }; }\nfunction pick(flag: boolean): Late | string { return flag ? make() : 's'; }\nconsole.log(pick(true) + '!');\n"},
		{"an honest class beside one", "class Late { valueOf(): number { return 5; } }\nclass Other { get toString(): () => string { return (): string => 'o'; } }\nconsole.log(String(new Other() instanceof Other));\n" + tail},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, ".ts")
			if err == nil || !strings.Contains(err.Error(), "a BinaryExpression with a union of differently held members and a string") {
				t.Fatalf("want combine's NotYet, got %v", err)
			}
		})
	}
}

// s += u doesn't take +'s union road: combine matches only a BinaryExpression whose token is +, so it keeps its
// NotYet (Lower's should). A later widening has to name it here.
func TestAddedUnionPlusAssignStaysNotYet(t *testing.T) {
	t.Parallel()
	const source = "class Money { valueOf(): number { return 5; } }\nfunction pick(flag: boolean): Money | string { return flag ? new Money() : 's'; }\nlet s = '';\ns += pick(true);\nconsole.log(s);\n"
	_, err := lowerSourceAs(t, source, ".ts")
	if err == nil || !strings.Contains(err.Error(), "a BinaryExpression with a string and a union of differently held members") {
		t.Fatalf("want combine's NotYet, got %v", err)
	}
}
