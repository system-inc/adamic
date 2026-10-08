package load

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestProductionProjectSitesRetainUnconvertedErrors(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t,
		[2]string{"tsconfig.json", `{"compilerOptions":{"strict":true,"useUnknownInCatchVariables":false,"lib":["es2020"],"noEmit":true},"files":["main.ts"]}`},
		[2]string{"main.ts", `const items: number[] = [];
const first: number = items[0];
const point: { x?: number } = { x: undefined };
try { throw new Error('boom'); } catch (error) { const message = error.message; }
const ordinary: number = 'wrong';
`})
	_, err := Load(paths[1:])
	var rejected *CheckError
	if !errors.As(err, &rejected) || len(rejected.OptionSites) != 3 {
		t.Fatalf("unconverted sites must remain errors and be recorded: %v", err)
	}
	if !strings.Contains(err.Error(), "main.ts:5:") || !strings.Contains(err.Error(), "main.ts:4:") {
		t.Fatalf("ordinary or unconverted error disappeared: %v", err)
	}
	if strings.Contains(err.Error(), "main.ts:2:") || strings.Contains(err.Error(), "main.ts:3:") {
		t.Fatalf("indexed row was not deferred to its lowering check: %v", err)
	}
}

func TestProductionProjectPreservesCatchTypeAndLib(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t,
		[2]string{"tsconfig.json", `{"compilerOptions":{"strict":true,"useUnknownInCatchVariables":false,"lib":["es2020"],"noEmit":true},"files":["main.ts"]}`},
		[2]string{"main.ts", `const value = BigInt(1);
try { throw new Error('boom'); } catch (error) { const held = error; }
`})
	program, err := Load(paths[1:])
	if err != nil {
		t.Fatal(err)
	}
	foundCatch := false
	for _, declaration := range program.Declarations(context.Background()) {
		if declaration.Name == "held" {
			foundCatch = true
			if declaration.Type != "any" {
				t.Fatalf("project catch type replaced: %+v", declaration)
			}
		}
	}
	if !foundCatch {
		t.Fatal("catch witness was not checked")
	}
	oldLib := writeProgram(t,
		[2]string{"tsconfig.json", `{"compilerOptions":{"strict":true,"lib":["es2015"],"noEmit":true},"files":["main.ts"]}`},
		[2]string{"main.ts", `const value = BigInt(1);`})
	if _, err := Load(oldLib[1:]); err == nil || !strings.Contains(err.Error(), "BigInt") {
		t.Fatalf("project's old lib was overwritten: %v", err)
	}
}

func TestProductionAdamicKeepsStrictOptions(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t,
		[2]string{"tsconfig.json", `{"compilerOptions":{"strict":false},"files":["main.a"]}`},
		[2]string{"main.a", `const values: number[] = []; const value: number = values[0];`})
	if _, err := Load(paths[1:]); err == nil || !strings.Contains(err.Error(), "undefined") {
		t.Fatalf(".a lost its proof obligation: %v", err)
	}
}

func TestProductionMixedOwnershipRefused(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t,
		[2]string{"tsconfig.json", `{"compilerOptions":{"strict":true,"lib":["es2020"],"noEmit":true},"files":["main.ts"]}`},
		[2]string{"main.ts", `import './proof.a';`},
		[2]string{"proof.a", `export const values: number[] = []; const value: number = values[0];`})
	if _, err := Load(paths[1:2]); err == nil || !strings.Contains(err.Error(), "mixed .a") {
		t.Fatalf("mixed import silently received relaxed options: %v", err)
	}
}

func TestProductionCompositeKeepsProjectRoots(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t,
		[2]string{"tsconfig.json", `{"compilerOptions":{"strict":true,"lib":["es2020"],"composite":true,"module":"esnext","noEmit":true},"files":["main.ts","other.ts"]}`},
		[2]string{"main.ts", `import { value } from './other'; export const held = value;`},
		[2]string{"other.ts", `export const value = 1;`})
	if _, err := Load(paths[1:2]); err != nil {
		t.Fatalf("composite project root list truncated: %v", err)
	}
}

func TestProductionProjectOverlayAttributed(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t,
		[2]string{"tsconfig.json", `{"compilerOptions":{"strict":true,"lib":["es2020"],"noEmit":true},"files":["main.ts"]}`},
		[2]string{"main.ts", `export const value = 1;`})
	loaded, err := LoadOverlay(paths[1:], map[string]string{paths[1]: `const items: number[] = []; const first: number = items[0];`})
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.OptionSites()) != 1 || loaded.OptionSites()[0].Options[0] != "noUncheckedIndexedAccess" {
		t.Fatalf("overlay read was not attributed: %v", loaded.OptionSites())
	}
}

func TestProductionExplicitProjectKeepsTypeScriptOptions(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t,
		[2]string{"tsconfig.json", `{"compilerOptions":{"strict":true,"exactOptionalPropertyTypes":false,"lib":["es2020"],"noEmit":true},"files":["main.ts"]}`},
		[2]string{"main.ts", `const source: { x: number | undefined } = { x: undefined }; const point: { x?: number } = source;`})
	_, err := LoadProject(paths[0])
	var rejected *CheckError
	if !errors.As(err, &rejected) || len(rejected.OptionSites) != 1 || !strings.Contains(rejected.OptionSites[0].Message, "exactOptionalPropertyTypes") {
		t.Fatalf("explicit .ts project must retain its option site: %v", err)
	}
}

func TestProductionExtraRootIsAudited(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t,
		[2]string{"tsconfig.json", `{"compilerOptions":{"strict":true,"lib":["es2020"],"noEmit":true},"files":["configured.ts"]}`},
		[2]string{"configured.ts", `export const configured = 1;`},
		[2]string{"extra.ts", `const source: { x: number | undefined } = { x: undefined }; const point: { x?: number } = source;`})
	_, err := Load(paths[2:])
	var rejected *CheckError
	if !errors.As(err, &rejected) || len(rejected.OptionSites) != 1 || rejected.OptionSites[0].File != paths[2] {
		t.Fatalf("extra root lost its stricter-option audit: %v", err)
	}
}

func TestProductionLiteralContractIsPending(t *testing.T) {
	paths := writeProgram(t,
		[2]string{"tsconfig.json", `{"compilerOptions":{"strict":true,"exactOptionalPropertyTypes":false,"lib":["es2020"],"noEmit":true},"files":["main.ts"]}`},
		[2]string{"main.ts", `const point: { x?: number } = { x: undefined };`})
	program, err := Load(paths[1:])
	if err != nil {
		t.Fatal(err)
	}
	if len(program.optionalLiterals) != 1 || len(program.ExplainedOptionalChecks()) != 0 {
		t.Fatal("a pending literal contract was lost or counted as emitted")
	}
	for _, sites := range program.optionalLiterals {
		if len(sites) != 1 || sites[0].Code != 2375 {
			t.Fatalf("wrong pending site: %v", sites)
		}
	}
}
