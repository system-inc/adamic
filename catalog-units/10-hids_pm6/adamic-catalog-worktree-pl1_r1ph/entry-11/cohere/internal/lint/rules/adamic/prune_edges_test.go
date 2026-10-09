package adamic

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/cohere/internal/lint/checking/flow"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
	"github.com/system-inc/cohere/internal/types/program"
)

// offeredSites is a rule that reports every site the file's walker offers, at the site's value, so a test
// can see which targets the walk is handed at all. The three relation rules judge only what this sees.
var offeredSites = rule.Rule{
	Name:             "adamic/offered-sites",
	NeedsTypeChecker: true,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.TypeChecker == nil || ctx.SourceFile == nil {
			return nil
		}
		return flow.WalkerFor(ctx).Listeners(func(site flow.Site) {
			ctx.ReportNode(site.Node, rule.Message{Id: "offered", Description: ctx.TypeChecker.TypeToString(site.Target)})
		})
	},
}

// offeredTargets is the target of every site offered in source, by the value's text.
func offeredTargets(t *testing.T, source string) map[string]string {
	t.Helper()
	result := runAdamic(t, offeredSites, source)
	text := rule_testing.FixtureText(source)
	targets := map[string]string{}
	for _, diagnostic := range result.Diagnostics {
		targets[text[diagnostic.Range.Pos():diagnostic.Range.End()]] = diagnostic.Message.Description
	}
	return targets
}

// typeErrorCount is how many errors tsc reports in source under Adamic's options.
func typeErrorCount(t *testing.T, source string) int {
	t.Helper()
	directory := t.TempDir()
	writeAdamicSupport(t)(directory)
	if err := os.WriteFile(filepath.Join(directory, "Case.ts"), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	graph, err := program.Build(program.Options{ConfigFileName: filepath.Join(directory, "tsconfig.json")})
	if err != nil {
		t.Fatalf("building: %v", err)
	}
	return len(graph.Program.GetSemanticDiagnostics(context.Background(), nil))
}

// TestTheWalkOffersEveryEdgeTargetWithAnObjectPart pins the prune from #m6tyg79 from the offered side:
// a target with an object part anywhere in it is still handed to the rules and judged. `{}` is an
// object type with no slots, and a union of a primitive and an array has one member to write through.
func TestTheWalkOffersEveryEdgeTargetWithAnObjectPart(t *testing.T) {
	t.Parallel()
	for name, fixture := range map[string]struct {
		source string
		value  string
		target string
	}{
		"the empty object type": {animals + `
const dogs: Dog[] = [rex];
const anything: {} = dogs;`, "dogs", "{}"},
		"a union of a primitive and an array": {animals + `
const dogs: Dog[] = [rex];
const either: string | Animal[] = dogs;`, "dogs", "string | Animal[]"},
		// #53w68gt: offered once the walk reaches them, and fired in invariant-mutable's own fixtures.
		"a type parameter whose constraint is an array": {animals + `
function admit<Pack extends Animal[], Narrow extends Pack>(narrow: Narrow): Pack {
	const pack: Pack = narrow;
	return pack;
}`, "narrow", "Pack"},
		"an intersection with an array": {animals + `
declare const tagged: Dog[] & { tag: string };
const wide: Animal[] & { tag: string } = tagged;`, "tagged", "Animal[] & { tag: string; }"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			targets := offeredTargets(t, fixture.source)
			if got, offered := targets[fixture.value]; !offered || got != fixture.target {
				t.Errorf("%s offered as %q (offered: %v), want %q; every site offered: %v", fixture.value, got,
					offered, fixture.target, targets)
			}
		})
	}

	// And judged: the union's array member is a hole, and the empty object type has no slot to write.
	union := animals + `
const dogs: Dog[] = [rex];
const either: string | Animal[] = dogs;`
	result := runAdamic(t, InvariantMutable, union)
	rule_testing.ExpectFindings(t, result, "mutableWidening")
	expectSpans(t, union, result, "dogs")
	rule_testing.ExpectClean(t, runAdamic(t, InvariantMutable, animals+`
const dogs: Dog[] = [rex];
const anything: {} = dogs;`))
}

// TestTheWalkSkipsTargetsNothingCanBeWrittenThrough pins the prune from the skipped side. An unconstrained
// type parameter, `object` and `unknown` have no object part, so a site whose target is one is never
// offered. That is sound only because no write reaches the original through such a name: each probe below
// tries one and tsc refuses it, so a lie there needs a cast, which is no-unchecked-cast's to see.
func TestTheWalkSkipsTargetsNothingCanBeWrittenThrough(t *testing.T) {
	t.Parallel()
	for name, fixture := range map[string]struct {
		source string
		value  string
		write  string
	}{
		// Inside the generic code, where the type parameter is the target: at a call it is instantiated, and a
		// site whose source is the instantiation is the same type, dropped before the prune is reached.
		"an unconstrained type parameter": {
			`function hold<Item, Other extends Item>(other: Other): Item {
	const held: Item = other;
	return held;
}`,
			"other",
			`function hold<Item, Other extends Item>(other: Other, extra: Item): void {
	const held: Item = other;
	held.push(extra);
}`,
		},
		"object": {
			animals + `const dogs: Dog[] = [rex];
const thing: object = dogs;`,
			"dogs",
			animals + `const dogs: Dog[] = [rex];
const thing: object = dogs;
thing.push({ name: 'Tom' });`,
		},
		"unknown": {
			animals + `const dogs: Dog[] = [rex];
const thing: unknown = dogs;`,
			"dogs",
			animals + `const dogs: Dog[] = [rex];
const thing: unknown = dogs;
thing.push({ name: 'Tom' });`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			targets := offeredTargets(t, fixture.source)
			if target, offered := targets[fixture.value]; offered {
				t.Errorf("%s was offered as %q, so the prune no longer skips this target", fixture.value, target)
			}
			if typeErrorCount(t, fixture.source) != 0 {
				t.Fatalf("the skipped site is not a program tsc accepts, so it proves nothing")
			}
			if typeErrorCount(t, fixture.write) == 0 {
				t.Errorf("tsc accepts a write through the skipped target, so skipping it can hide a hole")
			}
		})
	}
}
