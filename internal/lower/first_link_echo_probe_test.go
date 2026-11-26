package lower

import "testing"

// A scratch probe for #nxb3xh1 head 4, never landed: what each Echo guard's candidate program lowers to.
func TestFirstLinkEchoProbe(t *testing.T) {
	t.Parallel()
	statics := "class Counter {\n    static start(): number {\n        return 1;\n    }\n}\ninterface Point {\n    readonly x: number;\n}\n"
	for _, probe := range []struct{ name, file, source string }{
		{"unlowerable", "a", "function outer(): number {\n\tfunction inner(): number {\n\t\tconst self = (): number => self.length;\n\t\treturn self();\n\t}\n\treturn Math.random() + inner();\n}\nfunction live(): number {\n\tconst me = (): number => me.length;\n\treturn me();\n}\nconsole.log(`${outer() + live()}`);\n"},
		{"unlowerable ts", "ts", "function outer(): number {\n\tfunction inner(): number {\n\t\tconst self = (): number => self.length;\n\t\treturn self();\n\t}\n\treturn Math.random() + inner();\n}\nfunction live(): number {\n\tconst me = (): number => me.length;\n\treturn me();\n}\nconsole.log(`${outer() + live()}`);\n"},
		{"spread", "ts", statics + "function outer(point: Point): number {\n    function copy(): Point {\n        return { ...point };\n    }\n    return copy().x + Math.random();\n}\nconsole.log(String(outer({ x: 1 }) + Counter.start()));\n"},
		{"cycle local", "ts", "function trapped(seed: unknown): number {\n\tlet b: unknown = seed;\n\tconst read = (): unknown => b;\n\tb = read;\n\treturn Math.random();\n}\nconsole.log(`${trapped(0)}`);\n"},
		{"cycle parameter", "ts", "function trapped(b: unknown): number {\n\tconst read = (): unknown => b;\n\tb = read;\n\treturn Math.random();\n}\nconsole.log(`${trapped(0)}`);\n"},
		{"cycle live", "ts", "function live(seed: unknown): number {\n\tlet shared: unknown = seed;\n\tconst keep = (): unknown => shared;\n\tconst spoil = (): number => {\n\t\tshared = keep;\n\t\treturn Math.random();\n\t};\n\tshared = keep;\n\treturn spoil();\n}\nconsole.log(`${live(0)}`);\n"},
		{"cycle live and trapped", "ts", "function trapped(b: unknown): number {\n\tconst read = (): unknown => b;\n\tb = read;\n\treturn Math.random();\n}\nfunction live(c: unknown): number {\n\tconst read = (): unknown => c;\n\tc = read;\n\treturn 1;\n}\nconsole.log(`${trapped(0) + live(0)}`);\n"},
	} {
		lowerer := lowerFirstLink
		if probe.file == "ts" {
			lowerer = lowerFirstLinkTS
		}
		program, err := lowerer(t, probe.source, true)
		if err != nil {
			t.Logf("probe %s: traps: error %v", probe.name, err)
		} else {
			t.Logf("probe %s: traps: rows %+v abandoned %+v", probe.name, firstLinkLocalSites(program), program.TrapJudgmentsAbandoned)
		}
		_, err = lowerer(t, probe.source, false)
		t.Logf("probe %s: no traps: %v", probe.name, err)
	}
}
