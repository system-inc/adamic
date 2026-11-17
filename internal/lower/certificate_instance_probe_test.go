package lower

import "testing"

// Scratch probe, never committed: certificates A and B called where the element type isn't the mapped
// type, so the implementation's instance stores T | U while the caller reads U (#5fws6q8).
func TestCertificateInstanceProbe(t *testing.T) {
	show := "function show(entries: Map<string, string>): void {\n    for (const [key, value] of entries) {\n        console.log(`${key}=${value.toUpperCase()}`);\n    }\n}\nconst numbers: number[] = [1, 2, 3];\n"
	for _, row := range []struct{ name, source string }{
		{"a_number_to_string", overloadCertificateArrayFrom + "const counts = new Map<string, number>([[\"a\", 3], [\"b\", 1]]);\nconst named = arrayFrom(counts.values(), x => `v${x}`);\nconsole.log(named.map(s => s.toUpperCase()).join(\",\"));\nconst first = named[0];\nconsole.log(`${first === undefined ? -1 : first.length}`);\n"},
		{"a_string_to_number", overloadCertificateArrayFrom + "const names = new Map<string, string>([[\"a\", \"xyz\"], [\"b\", \"q\"]]);\nconst sizes = arrayFrom(names.values(), s => s.length);\nconsole.log(sizes.map(n => (n + 1).toFixed(1)).join(\",\"));\n"},
		{"b2_number_to_string", overloadCertificateArrayToMap + show + "show(arrayToMap(numbers, n => `k${n}`, n => `v${n * 2}`));\n"},
		{"b2_string_to_number", overloadCertificateArrayToMap + "const items: string[] = [\"a\", \"bb\"];\nconst lengths = arrayToMap(items, s => s, s => s.length);\nfor (const [key, value] of lengths) console.log(`${key}=${(value + 1).toFixed(1)}`);\n"},
		{"b4_number_to_string", overloadCertificateArrayToMap + show + "show(arrayToMap<number, string>(numbers, n => `k${n}`, n => `v${n * 2}`));\n"},
	} {
		t.Run(row.name, func(t *testing.T) {
			program, err := lowerSourceAs(t, row.source, ".ts")
			if err != nil {
				t.Logf("ROW %s stops: %v", row.name, err)
				return
			}
			code, stdout, stderr := brandedNative(t, program)
			t.Logf("ROW %s LOWERS: native exit %d stdout %q stderr %.300q", row.name, code, stdout, stderr)
			lowersAndAgreesWithNodeAs(t, row.source, ".ts")
			t.Logf("ROW %s agrees with Node on the JavaScript backend", row.name)
		})
	}
}
