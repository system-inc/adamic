package lower

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

func TestUnclassifiedCaughtPrototypeMemberIsNotYet(t *testing.T) {
	for _, name := range []string{"length", "name", "stack", "toString", "message", "static-message", "imported-message"} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "main.ts")
			prefix := ""
			if name == "message" {
				prefix = "class Tagged { message(): string { return 'm'; } }\n"
			}
			if name == "static-message" {
				prefix = "class Tagged { static message = 3; }\n"
				name = "message"
			}
			dependency := "export {};"
			if name == "imported-message" {
				prefix = "import { Tagged } from './tagged';\n"
				dependency = "export class Tagged { message(): string { return 'm'; } }"
				name = "message"
			}
			source := prefix + "try { throw 'x'; } catch (e) { console.log(`${typeof e." + name + "}`); }\n"
			for file, content := range map[string]string{
				"main.ts":       source,
				"tagged.ts":     dependency,
				"tsconfig.json": `{"compilerOptions":{"strict":true,"useUnknownInCatchVariables":false,"lib":["es2020"],"types":[],"noEmit":true},"files":["main.ts"]}`,
			} {
				if err := os.WriteFile(filepath.Join(dir, file), []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
			}
			program, err := load.Load([]string{path})
			if err != nil {
				t.Fatal(err)
			}
			_, err = Lower(context.Background(), program)
			want := "stage 0 can't lower an unclassified caught property outside ordinary message or code data fields; narrow to the receiver's declared type before reading it yet"
			if err == nil || !strings.HasSuffix(err.Error(), want) {
				t.Fatalf("got %v, want %q", err, want)
			}
		})
	}
}

func TestProjectCatchNullableViewsKeepUnsupportedTagsRefused(t *testing.T) {
	for _, source := range []string{
		"function inspect(value: unknown): void { try { throw value; } catch (e) { console.log(typeof e.message); } } const value: {message: string | null | undefined} = {message: null}; inspect(value);",
		"function inspect(value: unknown): void { try { throw value; } catch (e) { console.log(typeof e.message); } } inspect({payload: null});",
	} {
		directory := t.TempDir()
		for name, content := range map[string]string{
			"main.ts":       source,
			"tsconfig.json": `{"compilerOptions":{"strict":true,"useUnknownInCatchVariables":false,"lib":["es2020"],"types":[],"noEmit":true},"files":["main.ts"]}`,
		} {
			if err := os.WriteFile(filepath.Join(directory, name), []byte(content), 0644); err != nil {
				t.Fatal(err)
			}
		}
		program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
		if err != nil {
			t.Fatal(err)
		}
		_, err = Lower(context.Background(), program)
		if err == nil || !strings.Contains(err.Error(), "nullable field viewed as unknown or object") {
			t.Fatalf("unsupported null tags admitted: %v", err)
		}
	}
}
