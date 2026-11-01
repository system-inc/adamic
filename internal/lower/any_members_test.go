package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

// A member the program declares any is refused where it's declared, with its fix (#tmdhz1d, step 2,
// #bw3xg7c): written on an interface's property or a type literal's, through an alias of any, on a class
// field, or inferred from a class field's initializer. Language's v2, the program that read text as a
// string natively, now stops at Loose's declaration. Mutant: refuseAnyMember unregistered, under which
// every program here reaches lowering, and v2 stops only at step 1's relation NotYet.
func TestAnyMemberIsRefusedWhereDeclared(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, want string }{
		{"v2", `interface Loose { readonly kind: number; readonly pos: number; readonly text: any; }
interface HasText { readonly text: string; }
function show(n: Loose): void { const h: HasText = n; console.log(h.text); }
const w = { kind: 9, pos: 1, text: 5 }; show(w);`, "a member typed any (text, declared)"},
		{"a type literal's property", `function show(n: { readonly text: any }): void { console.log(typeof n.text); }
show({ text: 5 });`, "a member typed any (text, declared)"},
		{"through an alias of any", `type Anything = any;
interface Bag { readonly item: Anything; }
const b: Bag = { item: 1 }; console.log(typeof b.item);`, "a member typed any (item, declared)"},
		{"a class field", `class Holder { value: any = 1; }
console.log(typeof new Holder().value);`, "a member typed any (value, declared)"},
		{"a class field inferred", `class Holder { value = JSON.parse("1"); }
console.log(typeof new Holder().value);`, "a member typed any (value, inferred from its initializer)"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerTypeScriptAssertionSource(t, probe.source)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(refused.What, probe.want) || !strings.Contains(refused.Fix, "unknown") {
				t.Fatalf("want Refused with %q and the unknown fix, got %v", probe.want, err)
			}
		})
	}
}

// A member typed unknown, or one whose any is nested (items: any[]), isn't this rule's: the first lowers
// and agrees with Node, and the second is step 1's relation to judge where it's read as a typed member.
func TestAnyMemberLeavesUnknownAndNested(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNodeAs(t, `interface Opaque { readonly text: unknown; }
const o: Opaque = { text: 5 };
console.log(typeof o.text);`, ".ts")
	_, err := lowerTypeScriptAssertionSource(t, `interface Bag { readonly items: any[]; }
interface Names { readonly items: readonly string[]; }
function show(b: Bag): void { const n: Names = b; console.log(String(n.items[0])); }
show({ items: [5] });`)
	var pending *NotYet
	if !errors.As(err, &pending) || !strings.Contains(err.Error(), "a member typed any (items[]) related to string") {
		t.Fatalf("a nested any: want step 1's relation NotYet, got %v", err)
	}
}

// tsc's four members typed any whose type its public JSON API fixes are NotYet, each naming its file and
// member and that step 3's JSON design types it (#3f4rr76), and every other member typed any is Refused
// (Adamic's ruling on #tmdhz1d, Nov 1). A fifth member typed any in the same file is Refused, and so is
// one of the four's names in a file of another name. Mutant: a site dropped from anyMemberAllowance, under
// which its pin is Refused.
func TestAnyMemberAllowanceIsTheFourJSONSites(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, file, source, want string }{
		{"ParsedCommandLine.raw", "types.ts", "export interface ParsedCommandLine { raw?: any; }\n", "NotYet: a member typed any (types.ts ParsedCommandLine.raw) that step 3's JSON design types (#3f4rr76)"},
		{"readConfigFile.config", "commandLineParser.ts", "export function readConfigFile(fileName: string): { config?: any } { return { config: fileName }; }\n", "NotYet: a member typed any (commandLineParser.ts readConfigFile.config) that step 3's JSON design types (#3f4rr76)"},
		{"parseConfigFileTextToJson.config", "commandLineParser.ts", "export function parseConfigFileTextToJson(jsonText: string): { config?: any } { return { config: jsonText }; }\n", "NotYet: a member typed any (commandLineParser.ts parseConfigFileTextToJson.config) that step 3's JSON design types (#3f4rr76)"},
		{"ParsedTsconfig.raw", "commandLineParser.ts", "export interface ParsedTsconfig { raw: any; }\n", "NotYet: a member typed any (commandLineParser.ts ParsedTsconfig.raw) that step 3's JSON design types (#3f4rr76)"},
		{"a fifth member", "types.ts", "export interface ParsedCommandLine { raw?: any; other?: any; }\n", "Refused: a member typed any (other, declared)"},
		{"another file", "main.ts", "export interface ParsedCommandLine { raw?: any; }\n", "Refused: a member typed any (raw, declared)"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			file := filepath.Join(t.TempDir(), probe.file)
			if err := os.WriteFile(file, []byte(probe.source+"console.log('ran');\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			program, err := load.Load([]string{file})
			if err != nil {
				t.Fatal(err)
			}
			_, err = Lower(context.Background(), program)
			if err == nil || !strings.HasPrefix(stoppedFor(err), probe.want) {
				t.Fatalf("want %q, got %v", probe.want, err)
			}
		})
	}
}
