package lower

import (
	"os"
	"testing"
)

func TestZZNeverRestProbe(t *testing.T) {
	for _, probe := range []struct{ name, source string }{
		{"tsc", "interface Node { readonly kind: number; }\ntype AnyFunction = (...args: never[]) => void;\nfunction fail(message: string, mark?: AnyFunction): never {\n\tthrow new Error(message);\n}\nfunction assertIsDefined<T>(value: T, message?: string, mark?: AnyFunction): asserts value is NonNullable<T> {\n\tif (value === undefined || value === null) {\n\t\tfail(message ?? 'undefined', mark || assertIsDefined);\n\t}\n}\nfunction assertEachIsDefined<T extends Node>(value: readonly T[], message?: string, mark?: AnyFunction): asserts value is readonly T[];\nfunction assertEachIsDefined<T>(value: readonly T[], message?: string, mark?: AnyFunction): asserts value is readonly NonNullable<T>[];\nfunction assertEachIsDefined<T>(value: readonly T[], message?: string, mark?: AnyFunction) {\n\tfor (const v of value) {\n\t\tassertIsDefined(v, message, mark || assertEachIsDefined);\n\t}\n}\nconst xs: readonly number[] = [1, 2];\nassertEachIsDefined(xs);\nconsole.log('ok');\n"},
		{"identity", "interface Node { readonly kind: number; }\ntype AnyFunction = (...args: never[]) => void;\nfunction fail(message: string, mark?: AnyFunction): never {\n\tthrow new Error(message);\n}\nfunction assertIsDefined<T>(value: T, message?: string, mark?: AnyFunction): asserts value is NonNullable<T> {\n\tif (value === undefined || value === null) {\n\t\tfail(message ?? 'undefined', mark || assertIsDefined);\n\t}\n}\nfunction assertEachIsDefined<T extends Node>(value: readonly T[], message?: string, mark?: AnyFunction): asserts value is readonly T[];\nfunction assertEachIsDefined<T>(value: readonly T[], message?: string, mark?: AnyFunction): asserts value is readonly NonNullable<T>[];\nfunction assertEachIsDefined<T>(value: readonly T[], message?: string, mark?: AnyFunction) {\n\tfor (const v of value) {\n\t\tassertIsDefined(v, message, mark || assertEachIsDefined);\n\t}\n}\nconst m: AnyFunction = assertEachIsDefined;\nconsole.log(String(m === assertEachIsDefined));\n"},
		{"zero_call", "interface Node { readonly kind: number; }\ntype AnyFunction = (...args: never[]) => void;\nfunction fail(message: string, mark?: AnyFunction): never {\n\tthrow new Error(message);\n}\nfunction assertIsDefined<T>(value: T, message?: string, mark?: AnyFunction): asserts value is NonNullable<T> {\n\tif (value === undefined || value === null) {\n\t\tfail(message ?? 'undefined', mark || assertIsDefined);\n\t}\n}\nfunction assertEachIsDefined<T extends Node>(value: readonly T[], message?: string, mark?: AnyFunction): asserts value is readonly T[];\nfunction assertEachIsDefined<T>(value: readonly T[], message?: string, mark?: AnyFunction): asserts value is readonly NonNullable<T>[];\nfunction assertEachIsDefined<T>(value: readonly T[], message?: string, mark?: AnyFunction) {\n\tfor (const v of value) {\n\t\tassertIsDefined(v, message, mark || assertEachIsDefined);\n\t}\n}\nconst m: AnyFunction = assertEachIsDefined;\nm();\nconsole.log('reached');\n"},
		{"node_slot", "interface Node { readonly kind: number; }\ntype AnyFunction = (...args: never[]) => void;\nfunction fail(message: string, mark?: AnyFunction): never {\n\tthrow new Error(message);\n}\nfunction assertIsDefined<T>(value: T, message?: string, mark?: AnyFunction): asserts value is NonNullable<T> {\n\tif (value === undefined || value === null) {\n\t\tfail(message ?? 'undefined', mark || assertIsDefined);\n\t}\n}\nfunction assertEachIsDefined<T extends Node>(value: readonly T[], message?: string, mark?: AnyFunction): asserts value is readonly T[];\nfunction assertEachIsDefined<T>(value: readonly T[], message?: string, mark?: AnyFunction): asserts value is readonly NonNullable<T>[];\nfunction assertEachIsDefined<T>(value: readonly T[], message?: string, mark?: AnyFunction) {\n\tfor (const v of value) {\n\t\tassertIsDefined(v, message, mark || assertEachIsDefined);\n\t}\n}\nconst f: (value: readonly Node[]) => void = assertEachIsDefined;\nf([{ kind: 1 }]);\nconsole.log('ok');\n"},
		{"one_signature", "interface Node { readonly kind: number; }\ntype AnyFunction = (...args: never[]) => void;\nfunction each<T extends Node>(value: readonly T[]): void {\n\tfor (const v of value) {\n\t\tconsole.log(String(v.kind));\n\t}\n}\nconst m: AnyFunction = each;\nconsole.log(String(m === each));\neach([{ kind: 3 }]);\n"},
	} {
		for _, extension := range []string{".ts", ".a"} {
			_, err := lowerSourceAs(t, probe.source, extension)
			result := "lowered"
			if err != nil {
				result = err.Error()
			}
			os.Stderr.WriteString("probe " + probe.name + extension + ": " + result + "\n")
		}
	}
}
