package lower

import (
	"fmt"
	"testing"
)

const overloadTrapPick = "function pick(value: number): number;\nfunction pick(value: string): string;\nfunction pick(value: number | string): number | string {\n\treturn value;\n}\n"

func TestOverloadTrapProbe(t *testing.T) {
	for _, row := range []struct{ name, source string }{
		{"lowered_then_stopped", overloadTrapPick + "function second(): number {\n\tconst n = pick(1);\n\treturn n + Math.random();\n}\nconsole.log('before');\nconsole.log(`${second()}`);\n"},
		{"refused_unit", overloadTrapPick + "function second(): boolean {\n\tconst n = pick(1);\n\treturn n == 1;\n}\nconsole.log('before');\nconsole.log(`${second()}`);\n"},
		{"refused_statement", overloadTrapPick + "console.log('before');\nconst flag = pick(1) == 1;\nconsole.log(`${flag}`);\n"},
		{"control_no_stop", overloadTrapPick + "function second(): number {\n\treturn pick(1);\n}\nconsole.log(`${second()}`);\n"},
	} {
		func() {
			defer func() {
				if recovered := recover(); recovered != nil {
					t.Logf("PROBE %s: panic: %v", row.name, recovered)
				}
			}()
			program, err := lowerFirstLink(t, row.source, true)
			traps := 0
			if program != nil {
				traps = len(program.Traps)
			}
			t.Logf("PROBE %s: err=%v traps=%d %s", row.name, err, traps, fmt.Sprint(firstLinkLocalSitesOrNil(program)))
		}()
	}
}

func firstLinkLocalSitesOrNil(program any) string {
	return fmt.Sprintf("%v", program != nil)
}
