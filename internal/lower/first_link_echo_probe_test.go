package lower

import "testing"

// A scratch probe for #nxb3xh1 head 4, never landed: what each Echo guard's candidate program lowers to.
func TestFirstLinkEchoProbe(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, file, source string }{
		{"live alone", "ts", "function live(c: unknown): number {\n\tconst read = (): unknown => c;\n\treturn read() === undefined ? 0 : 1;\n}\nconsole.log(`${live(0)}`);\n"},
		{"live and a trapped caller", "ts", "function live(c: unknown): number {\n\tconst read = (): unknown => c;\n\treturn read() === undefined ? 0 : 1;\n}\nfunction trapped(b: unknown): number {\n\tconst read = (): unknown => b;\n\treturn live(read) + Math.random();\n}\nconsole.log(`${live(0) + trapped(0)}`);\n"},
		{"two live and a trapped caller", "ts", "function live(c: unknown): number {\n\tconst read = (): unknown => c;\n\treturn read() === undefined ? 0 : 1;\n}\nfunction relay(d: unknown): number {\n\treturn live(d);\n}\nfunction trapped(b: unknown): number {\n\treturn relay(b) + Math.random();\n}\nconsole.log(`${relay(0) + trapped(0)}`);\n"},
		{"live written", "ts", "function live(c: unknown): number {\n\tconst read = (): unknown => c;\n\tc = 1;\n\treturn read() === undefined ? 0 : 1;\n}\nconsole.log(`${live(0)}`);\n"},
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
