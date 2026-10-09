package nexus

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
	"github.com/system-inc/cohere/internal/types/program"
)

/*
 * The rule's findings are TypeScript's TS7030, so TypeScript is the oracle (#yj96emr). Each fixture is
 * built twice: once with noImplicitReturns on, where the type check reports TS7030 and those spans are
 * collected, and once with it off, where the rule runs. The two sets of spans must be the same, and each
 * fixture also states how many there are, so a pair of empty sets cannot pass a firing fixture.
 */

// implicitReturnConfiguration is the fixture tsconfig, with the two flags the rule reads left to the fixture
func implicitReturnConfiguration(noImplicitReturns bool, strict bool) string {
	return fmt.Sprintf(`{
	"compilerOptions": {
		"strict": %t,
		"noImplicitReturns": %t,
		"target": "ES2022",
		"lib": ["ES2022"],
		"moduleDetection": "force",
		"types": []
	},
	"include": ["**/*.ts"]
}`, strict, noImplicitReturns)
}

// typeScriptImplicitReturns is the spans TypeScript reports as TS7030 for the fixture with the flag on
func typeScriptImplicitReturns(t *testing.T, source string, strict bool) []string {
	t.Helper()
	directory := t.TempDir()
	text := strings.TrimSpace(source) + "\n"
	for name, contents := range map[string]string{
		"Case.ts":       text,
		"tsconfig.json": implicitReturnConfiguration(true, strict),
	} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	graph, err := program.Build(program.Options{
		ConfigFileName:   filepath.Join(directory, "tsconfig.json"),
		CurrentDirectory: directory,
		SingleThreaded:   true,
	})
	if err != nil {
		t.Fatalf("building the oracle's program: %v", err)
	}
	spans := []string{}
	for _, sourceFile := range graph.ProjectFiles() {
		if !strings.HasSuffix(sourceFile.FileName().AsString(), "/Case.ts") {
			continue
		}
		for _, diagnostic := range graph.Diagnostics(context.Background(), sourceFile) {
			if diagnostic.Code() == 7030 {
				spans = append(spans, text[diagnostic.Loc().Pos():diagnostic.Loc().End()])
			}
		}
	}
	sort.Strings(spans)
	return spans
}

// ruleImplicitReturns is the spans the rule reports for the fixture with the flag off
func ruleImplicitReturns(t *testing.T, source string, strict bool, noImplicitReturns bool) []string {
	t.Helper()
	result := rule_testing.RunTypedFilesWithSetup(t, CorrectnessNoImplicitReturn,
		map[string]string{"Case.ts": source}, "Case.ts",
		func(directory string) {
			configuration := implicitReturnConfiguration(noImplicitReturns, strict)
			if err := os.WriteFile(filepath.Join(directory, "tsconfig.json"), []byte(configuration), 0o644); err != nil {
				t.Fatal(err)
			}
		})
	text := strings.TrimSpace(source) + "\n"
	spans := []string{}
	for _, diagnostic := range result.Diagnostics {
		spans = append(spans, text[diagnostic.Range.Pos():diagnostic.Range.End()])
	}
	sort.Strings(spans)
	return spans
}

var implicitReturnFixtures = []struct {
	name   string
	source string
	strict bool
	want   int
}{
	// The positive control: a value on one path and the end reachable on another
	{"a return on one path and none at the end", `
export function label(kind: string) {
    if(kind === 'a') {
        return 'A';
    }
}`, true, 1},
	{"a declared type that admits undefined", `
export function first(list: string[]): string | undefined {
    for(const item of list) {
        if(item) {
            return item;
        }
    }
}`, true, 1},
	{"an async function, its promise unwrapped", `
export async function load(flag: boolean) {
    if(flag) {
        return 1;
    }
}`, true, 1},
	{"an arrow function with a block body", `
export const pick = (flag: boolean) => {
    if(flag) {
        return 1;
    }
};`, true, 1},
	{"a class method and an object literal method", `
export class Picker {
    pick(flag: boolean) {
        if(flag) {
            return 1;
        }
    }
}
export const picker = {
    pick(flag: boolean) {
        if(flag) {
            return 1;
        }
    },
};`, true, 2},
	{"a getter", `
export class Holder {
    get value() {
        if(Math.random() > 0.5) {
            return 1;
        }
    }
}`, true, 1},
	// Without strictNullChecks a bare return is reported where the function returns a value elsewhere
	{"a bare return without strictNullChecks", `
export function measure(flag: boolean) {
    if(flag) {
        return;
    }
    return 1;
}`, false, 1},

	// Kirk's ruling of 2026-08-25: an exhaustive switch ends nothing. Base's getDriver.
	{"an exhaustive switch over a union", `
type DatabaseType = { driver: 'd1' } | { driver: 'durable' } | { driver: 'planetscale' } | { driver: 'better-sqlite' };
export function getDriver(databaseType: DatabaseType): string | undefined {
    switch(databaseType.driver) {
        case 'd1':
            return 'd1-http';
        case 'durable':
            return 'durable-sqlite';
        case 'planetscale':
        case 'better-sqlite':
            return undefined;
    }
}`, true, 0},
	// The same ruling: a call returning never ends its path. Base's parseToml.
	{"a catch that ends in a call returning never", `
declare const process: { exit(code: number): never };
declare function parse(text: string): unknown;
export function parseToml(location: string, content: string): unknown {
    try {
        return parse(content);
    }
    catch(error) {
        console.error(location, error);
        process.exit(1);
    }
}`, true, 0},
	{"no return value anywhere", `
export function log(flag: boolean) {
    if(flag) {
        console.log('yes');
        return;
    }
}`, true, 0},
	{"every path returns", `
export function sign(value: number) {
    if(value < 0) {
        return -1;
    }
    return 1;
}`, true, 0},
	// With strictNullChecks a bare return is checked as returning undefined, which noImplicitReturns leaves alone
	{"a bare return beside a value, under strictNullChecks", `
export function measure(flag: boolean) {
    if(flag) {
        return;
    }
    return 1;
}`, true, 0},
	// A declared any needs no return value, whatever the body returns
	{"a declared any return", `
export function loose(flag: boolean): any {
    if(flag) {
        return 1;
    }
}`, true, 0},
	// A declared type that excludes undefined is TS2366, a type error the flag does not decide, not TS7030
	{"a declared type that excludes undefined", `
export function name(flag: boolean): string {
    if(flag) {
        return 'a';
    }
}`, true, 0},
	// A getter takes its type from its setter's annotation, which excludes undefined, so the same holds
	{"a getter typed by its setter", `
export class Holder {
    #stored = 0;
    get value() {
        if(Math.random() > 0.5) {
            return this.#stored;
        }
    }
    set value(next: number) {
        this.#stored = next;
    }
}`, true, 0},
	{"a declared void return", `
export function run(flag: boolean): void {
    if(flag) {
        return;
    }
}`, true, 0},
}

func TestCorrectnessNoImplicitReturnReportsWhatTypeScriptDoes(t *testing.T) {
	t.Parallel()

	for _, fixture := range implicitReturnFixtures {
		t.Run(fixture.name, func(t *testing.T) {
			t.Parallel()
			typeScript := typeScriptImplicitReturns(t, fixture.source, fixture.strict)
			if len(typeScript) != fixture.want {
				t.Fatalf("TypeScript reports %d TS7030 (%q), and the fixture says %d", len(typeScript), typeScript, fixture.want)
			}
			reported := ruleImplicitReturns(t, fixture.source, fixture.strict, false)
			if strings.Join(reported, "\x00") != strings.Join(typeScript, "\x00") {
				t.Errorf("the rule reports %q, and TypeScript reports %q", reported, typeScript)
			}
		})
	}
}

// Where the tsconfig turns the flag on, the type check already reports these, so the rule stands down
func TestCorrectnessNoImplicitReturnStandsDownWhenTheFlagIsOn(t *testing.T) {
	t.Parallel()

	control := implicitReturnFixtures[0]
	if reported := ruleImplicitReturns(t, control.source, true, false); len(reported) != 1 {
		t.Fatalf("the control reports %q with the flag off, so the next assertion would prove nothing", reported)
	}
	if reported := ruleImplicitReturns(t, control.source, true, true); len(reported) != 0 {
		t.Errorf("the rule reports %q with noImplicitReturns on, where the type check already does", reported)
	}
}

// The fixture pair through the default harness, whose tsconfig is strict and leaves the flag off
func TestCorrectnessNoImplicitReturnFixturePair(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFindings(t,
		rule_testing.RunTyped(t, CorrectnessNoImplicitReturn, "Case.ts", implicitReturnFixtures[0].source),
		"implicitReturn")
	for _, fixture := range implicitReturnFixtures {
		if fixture.want == 0 && fixture.strict {
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, CorrectnessNoImplicitReturn, "Case.ts", fixture.source))
		}
	}
}
