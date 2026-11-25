package lower

import (
	"context"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

func overloadTrapProbeShape() string {
	return strings.Replace(overloadLiteralShape, "BODY", "true", 1)
}

func TestOverloadTrapProbe(t *testing.T) {
	base := overloadTrapProbeShape()
	for _, row := range []struct{ name, source string }{
		{"lowered_then_stopped", base + "function second(): number {\n\tconst ok = canDecorate(true, declaration);\n\tconsole.log(String(ok));\n\treturn Math.random();\n}\nconsole.log('before');\nconsole.log(`${second()}`);\n"},
		{"refused_unit", base + "function second(n: number): boolean {\n\tconst ok = canDecorate(true, declaration);\n\treturn ok == true;\n}\nconsole.log('before');\nconsole.log(`${second(1)}`);\n"},
		{"refused_statement", base + "console.log('before');\nconst flag = canDecorate(true, declaration) == true;\nconsole.log(`${flag}`);\n"},
		{"control_no_stop", base + "console.log(String(canDecorate(true, declaration)));\n"},
	} {
		func() {
			defer func() {
				if recovered := recover(); recovered != nil {
					t.Logf("PROBE %s: panic: %v", row.name, recovered)
				}
			}()
			program, err := load.Load([]string{writeAgreementSource(t, "main.ts", row.source)})
			if err != nil {
				t.Logf("PROBE %s: load: %v", row.name, err)
				return
			}
			lowered, err := LowerWithOptions(context.Background(), program, Options{FirstLinkTraps: true})
			if err != nil {
				t.Logf("PROBE %s: err=%v", row.name, err)
				return
			}
			rows := []string{}
			for _, trap := range lowered.Traps {
				rows = append(rows, trap.Kind+" "+trap.Phase+" "+trap.Unit+" "+trap.Reason)
			}
			t.Logf("PROBE %s: traps=%d checked=%+v %v", row.name, len(lowered.Traps), lowered.OverloadResultChecks, rows)
		}()
	}
}
