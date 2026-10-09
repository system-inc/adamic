package load

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMixedAdamicAndTypeScriptWithoutReferences(t *testing.T) {
	t.Parallel()
	dir := projectFixture(t, map[string]string{
		"main.ts":       `import { value } from "./value.a"; console.log(value.toString());`,
		"value.a":       `export const value: number = 42;`,
		"tsconfig.json": `{"compilerOptions":{"strict":false,"target":"ES2022","module":"NodeNext","noEmit":true},"files":["main.ts"],"references":[]}`,
	})
	for _, roots := range [][]string{{filepath.Join(dir, "main.ts")}, {filepath.Join(dir, "main.ts"), filepath.Join(dir, "value.a")}} {
		if _, err := Load(roots); err != nil {
			t.Fatalf("mixed file build: %v", err)
		}
	}
	// Ignoring ordinary project options must not relax an Adamic file's checks.
	if err := os.WriteFile(filepath.Join(dir, "value.a"), []byte(`export const value: number = "bad";`), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load([]string{filepath.Join(dir, "main.ts")}); err == nil {
		t.Fatal("mixed file build dropped the Adamic type check")
	}
}
