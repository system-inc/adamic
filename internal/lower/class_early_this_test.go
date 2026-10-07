package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestConstructorsRefuseEarlyMethodCalls(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, receiver, fields, kind string }{
		{"super_early_silent.a", "super", "readonly leafTag = `leaf${2}`; readonly count = 5;", "`kind:${this.leafTag}:${this.count + 1}`"},
		{"super_early.a", "super", "readonly leafTag = `leaf${2}`;", "`kind:${this.leafTag.toUpperCase()}`"},
		{"super_early_this.a", "this", "readonly leafTag = `leaf${2}`;", "`kind:${this.leafTag.toUpperCase()}`"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			source := "abstract class Root { readonly rootTag = `root${1}`; abstract kind(): string; describe(): string { return `${this.rootTag}/${this.kind()}`; } }\n" +
				"class Middle extends Root { readonly seen: string; constructor() { super(); this.seen = " + probe.receiver + ".describe(); } override kind(): string { return 'middle'; } }\n" +
				"class Leaf extends Middle { " + probe.fields + " override kind(): string { return " + probe.kind + "; } }\n" +
				"console.log(new Middle().seen); console.log(new Leaf().seen);"
			_, err := lowerSource(t, source)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), probe.receiver+".describe(...)") || !strings.Contains(err.Error(), "before derived fields are initialized") || !strings.Contains(err.Error(), "after construction") {
				t.Fatalf("want refusal naming %s.describe with construction fix, got %v", probe.receiver, err)
			}
		})
	}
}

func TestConstructorsAllowInitializedSuperCalls(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`class Root { describe(): string { return 'ready'; } }
class Leaf extends Root { readonly tag: string; constructor() { super(); this.tag = 'set'; console.log(super.describe()); } }
const leaf = new Leaf();`,
		`class Root { describe(): string { return 'ready'; } }
class Middle extends Root { describeAgain(): string { return super.describe(); } }
class Leaf extends Middle { readonly tag = 'set'; }
console.log(new Leaf().describeAgain());`,
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Fatal(err)
		}
	}
}

func TestConstructorsRefuseEarlyThisClosures(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, closure string }{
		{"d7054e9_this_arrow_number.a", "class Counter { readonly start: number; readonly count: number; constructor() { const next = (): number => this.count + 1; this.start = next(); this.count = 5; } }\nconsole.log(`${new Counter().start}`); const scaled = [1, 2].map((value) => value * new Counter().start); console.log(scaled.join(','));", "next"},
		{"d7054e9_this_arrow.a", "class Box { readonly first: string; readonly later: string; readonly count: number; constructor(label: string) { const read = (): string => this.later; const counted = (): number => this.count + 1; this.first = `${read()}:${counted()}`; this.later = label; this.count = 2; } }\nconst box = new Box('set'); console.log(`${box.first}|${box.later}`);", "read"},
		{"nested arrow", "class Counter { readonly count: number; constructor() { const outer = () => () => this.count; this.count = 5; console.log(`${outer()()}`); } } const counter = new Counter();", "outer"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(err.Error(), "closure "+probe.closure) || !strings.Contains(err.Error(), "before every field is set") || !strings.Contains(err.Error(), "create the closure after the fields are set, or pass the value in") {
				t.Fatalf("want early closure refusal naming %s with fix, got %v", probe.closure, err)
			}
		})
	}
}

func TestConstructorsAllowInitializedThisReads(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"class Counter { readonly count: number; readonly start: number; constructor() { this.count = 5; this.start = this.count + 1; } } console.log(`${new Counter().start}`);",
		"class Counter { readonly count: number; constructor() { this.count = 5; const next = (): number => this.count + 1; console.log(`${next()}`); } } const counter = new Counter();",
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Fatal(err)
		}
	}
}
