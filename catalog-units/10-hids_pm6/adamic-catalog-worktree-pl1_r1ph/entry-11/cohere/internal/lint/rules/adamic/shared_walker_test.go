package adamic

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/types/program"
)

// sharedWalkerFixture has a hole for each of the three rules that share a file's flow.Walker, several of
// them at one node: the call's four arguments are three rules' sites at once.
const sharedWalkerFixture = `interface Animal { name: string }
interface Dog extends Animal { bark(): string }
interface Options { x: number; y?: number }
class Shelter { admit(animal: Animal): string { return animal.name; } }
class DogShelter { admit(dog: Dog): string { return dog.bark(); } }
class Box<T> { item: T; constructor(item: T) { this.item = item; } }

declare const dogs: Dog[];
declare const kennel: { pet: Dog };
declare const dogBox: Box<Dog>;
declare const narrow: { x: number };
declare const dogShelter: DogShelter;

function take(animals: Animal[], pen: { pet: Animal }, box: Box<Animal>, options: Options): void {
	console.log(String(animals.length + pen.pet.name.length) + String(box.item.name) + String(options.x));
}

export function run(): Shelter {
	take(dogs, kennel, dogBox, narrow);
	const animals: Animal[] = dogs;
	const options: Options = narrow;
	const box: Box<Animal> = dogBox;
	console.log(String(animals.length + options.x) + box.item.name);
	return dogShelter;
}
`

// TestTheRulesSharingAWalkerReportWhatEachReportsAlone: the three rules share one walker per file, which
// finds a node's sites once for all of them and remembers the checker's answers (#m6tyg79). Run together
// through the walk a real run uses, one file cache and one dispatch, they report exactly what each reports
// walked alone, so the sharing changes what is asked, never what is found.
func TestTheRulesSharingAWalkerReportWhatEachReportsAlone(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	writeAdamicSupport(t)(directory)
	if err := os.WriteFile(filepath.Join(directory, "Case.ts"), []byte(sharedWalkerFixture), 0o644); err != nil {
		t.Fatal(err)
	}
	walk := func(rules ...rule.Rule) []string {
		t.Helper()
		graph, err := program.Build(program.Options{ConfigFileName: filepath.Join(directory, "tsconfig.json")})
		if err != nil {
			t.Fatalf("building: %v", err)
		}
		if diagnostics := graph.Program.GetSemanticDiagnostics(context.Background(), nil); len(diagnostics) > 0 {
			t.Fatalf("the fixture has %d type errors, so it is not a program tsc accepts", len(diagnostics))
		}
		result, err := graph.Walk(context.Background(), graph.ProjectFiles(), rules)
		if err != nil {
			t.Fatalf("walking: %v", err)
		}
		findings := []string{}
		for _, diagnostic := range result.Diagnostics {
			findings = append(findings, fmt.Sprintf("%s %d-%d %s: %s", filepath.Base(diagnostic.SourceFile.FileName().AsString()),
				diagnostic.Range.Pos(), diagnostic.Range.End(), diagnostic.RuleName, diagnostic.Message.Description))
		}
		sort.Strings(findings)
		return findings
	}

	alone := []string{}
	for _, subject := range []rule.Rule{InvariantMutable, NominalClass, NoOptionalWidening} {
		found := walk(subject)
		if len(found) == 0 {
			t.Fatalf("%s found nothing walked alone, so sharing is untested for it", subject.Name)
		}
		alone = append(alone, found...)
	}
	sort.Strings(alone)
	together := walk(InvariantMutable, NominalClass, NoOptionalWidening)
	if !reflect.DeepEqual(together, alone) {
		t.Errorf("walked together the rules report\n  %q\nand walked alone\n  %q", together, alone)
	}
}
