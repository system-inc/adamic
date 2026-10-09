package load

import (
	"strings"
	"testing"
)

func TestExplicitProjectRootIsAudited(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t,
		[2]string{"tsconfig.json", `{"compilerOptions":{"strict":true,"useUnknownInCatchVariables":false,"lib":["es2020"]},"files":["included.ts"]}`},
		[2]string{"included.ts", `export {};`},
		[2]string{"extra.ts", `const value: number = 'wrong';`})
	if _, err := Load(paths[2:]); err == nil || !strings.Contains(err.Error(), "Type 'string' is not assignable to type 'number'") {
		t.Fatalf("explicit root escaped project checking: %v", err)
	}
}
