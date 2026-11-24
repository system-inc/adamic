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

// An array of strings is an arm + converts as Node does, by Array.prototype.toString's join through
// ToString (family 86, tsc's formatStringFromArgs over `string | number | boolean | readonly string[] |
// SourceFile`, under row 205 and Adamic's ruling on + over a union). Each row agrees with Node on the
// JavaScript backend and natively under the sanitizers, an empty array and one holding commas and an empty
// string included, and the union is read once per +. Mutant: addedObjectsConvert's array branch answering
// false, as before, and every row is combine's NotYet.
func TestAddedUnionStringArrayAgrees(t *testing.T) {
	t.Parallel()
	const argument = "function argument(list: boolean): string | readonly string[] {\n    if (list) return [\"a\", \"b\"];\n    return \"value\";\n}\n"
	const holder = "interface Holder { readonly name: string }\ntype Argument = string | number | boolean | readonly string[] | Holder;\n" +
		"function checkDefined<T>(value: T | undefined): T { if (value === undefined) throw new Error('undefined'); return value; }\n" +
		"function format(text: string, args: readonly Argument[]): string {\n    return text.replace(/\\{(\\d+)\\}/g, (_match: string, index: string) => \"\" + checkDefined(args[+index]));\n}\n"
	for _, probe := range []struct{ name, source, stdout string }{
		{"diagnostic plus", argument + "console.log(\"\" + argument(true));\nconsole.log(\"\" + argument(false));\n", "a,b\nvalue\n"},
		{"diagnostic arguments", holder + "const holder: Holder = { name: 'h' };\nconsole.log(format('{0}|{1}|{2}|{3}|{4}', ['s', 7, true, ['x', 'y'], holder]));\n", "s|7|true|x,y|[object Object]\n"},
		{"empty and awkward", "function pick(which: number): string | string[] { return which === 0 ? [] : which === 1 ? ['a,b', '', 'c'] : 'plain'; }\nconsole.log('[' + pick(0) + ']');\nconsole.log('[' + pick(1) + ']');\nconsole.log('[' + pick(2) + ']');\n", "[]\n[a,b,,c]\n[plain]\n"},
		{"each side and a chain", argument + "const list = argument(true);\nconsole.log(list + '');\nconsole.log('' + list);\nconsole.log('x' + list + 'y');\n", "a,b\na,b\nxa,by\n"},
		{"read once per plus", "function read(): string | Array<string> { console.log('read'); return ['p', 'q']; }\nconsole.log('<' + read() + '>');\nconsole.log(read() + '!' + read());\n", "read\n<p,q>\nread\nread\np,q!p,q\n"},
		{"logger field", "interface Entry { readonly message: string | string[] }\nfunction line(entry: Entry): string { return 'log: ' + entry.message; }\nconsole.log(line({ message: 'one' }));\nconsole.log(line({ message: ['two', 'three'] }));\n", "log: one\nlog: two,three\n"},
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

// Every other array keeps combine's NotYet at compile time, since the helper's join stops at run time on an
// array of only numbers or booleans: a number[], a (string | number)[], a string[][], a tuple, and a string
// array beside undefined, which addedUnion's nullable rule holds. += keeps its own NotYet
// (TestAddedUnionPlusAssignStaysNotYet). Mutant: the branch taking any array, and the number row lowers.
func TestAddedUnionOtherArraysStayNotYet(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, arm, value string }{
		{"numbers", "number[]", "[1, 2]"},
		{"mixed", "(string | number)[]", "['a', 1]"},
		{"nested", "readonly string[][]", "[['a']]"},
		{"tuple", "[string, string]", "['a', 'b']"},
		{"beside undefined", "readonly string[] | undefined", "['a']"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			source := "function pick(flag: boolean): string | " + probe.arm + " { return flag ? " + probe.value + " : 's'; }\nconsole.log('' + pick(true));\n"
			_, err := lowerSourceAs(t, source, ".ts")
			if err == nil || !strings.Contains(err.Error(), "a BinaryExpression with a string and") {
				t.Fatalf("want combine's NotYet, got %v", err)
			}
		})
	}
	t.Run("plus assign", func(t *testing.T) {
		t.Parallel()
		const source = "function pick(flag: boolean): string | string[] { return flag ? ['a'] : 's'; }\nlet s = '';\ns += pick(true);\nconsole.log(s);\n"
		_, err := lowerSourceAs(t, source, ".ts")
		if err == nil || !strings.Contains(err.Error(), "a BinaryExpression with a string and a union of differently held members") {
			t.Fatalf("want combine's NotYet, got %v", err)
		}
	})
}
