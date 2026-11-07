package lower

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestCensusOverloadRelation(test *testing.T) {
	test.Parallel()
	probes := []struct{ name, source, diagnostic string }{
		{"parameter contravariance", `interface Animal { readonly name: string; }
interface Dog extends Animal { readonly bark: string; }
interface Handler { handle(animal: Animal): void; }
interface DogHandler { handle(dog: Dog): void; }
function serve(handler: DogHandler): string;
function serve(handler: Handler): string { handler.handle({ name: 'cat' }); return 'written'; }
`, "overload 1 of serve parameter handler"},
		{"generic parameter constraints", `interface Animal { readonly name: string; }
interface Dog extends Animal { readonly bark: string; }
interface Handler { handle(animal: Animal): void; }
interface DogHandler { handle(dog: Dog): void; }
function serve<T extends DogHandler>(handler: T): string;
function serve<U extends Handler>(handler: U): string { handler.handle({ name: 'cat' }); return 'written'; }
`, "overload 1 of serve type parameter T"},
		{"mutable parameters", `interface Animal { readonly name: string; }
interface Dog extends Animal { readonly bark: string; }
function serve(dogs: Dog[]): string;
function serve(animals: Animal[]): string { animals.push({ name: 'cat' }); return 'written'; }
`, "overload 1 of serve parameter dogs"},
		{"result covariance", `type Path = string & { readonly __pathBrand: true };
function ensureTrailingDirectorySeparator(path: Path): Path;
function ensureTrailingDirectorySeparator(path: string): string;
function ensureTrailingDirectorySeparator(path: string) { return path + '/'; }
`, "overload 1 of ensureTrailingDirectorySeparator result"},
	}
	for _, probe := range probes {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			_, lowerSourceError := lowerSource(test, probe.source)
			var refused *Refused
			if !errors.As(lowerSourceError, &refused) || !strings.Contains(refused.What, probe.diagnostic) {
				test.Fatalf("expected strict overload refusal naming %q; got %v", probe.diagnostic, lowerSourceError)
			}
		})
	}
}

func TestCensusSmallFiniteKeyRead(test *testing.T) {
	test.Parallel()
	lowersAndAgreesWithNode(
		test,
		`const wildcardMatchers = { files: 'file', directories: 'directory', exclude: 'excluded' };
function pick(usage: 'files' | 'directories' | 'exclude'): string { return wildcardMatchers[usage]; }
console.log(pick('files'));
`,
	)
}

func TestCensusRestMutableElements(test *testing.T) {
	test.Parallel()
	_, lowerSourceError := lowerSource(test, `interface Animal { readonly name: string; }
interface Dog extends Animal { readonly bark: string; }
interface House { pet: Animal; }
interface DogHouse { pet: Dog; }
function replace(...houses: House[]): void { for (const house of houses) { house.pet = { name: 'cat' }; } }
const houses: DogHouse[] = [{ pet: { name: 'dog', bark: 'woof' } }];
const first = houses[0] ?? { pet: { name: 'fallback', bark: 'fallback' } };
replace(...houses);
console.log(first.pet.bark);
`)
	var refused *Refused
	if !errors.As(lowerSourceError, &refused) {
		test.Fatalf("a rest copy still shares mutable elements; expected refusal, got %v", lowerSourceError)
	}
	test.Logf("rest element relation: %v", refused)
}

// Flow-narrowed never is not an executable arm that fails to return.
func TestCensusBooleanDeadBranch(test *testing.T) {
	test.Parallel()
	source, readError := os.ReadFile("../oracle/testdata/census_boolean_dead_branch.a")
	if readError != nil {
		test.Fatal(readError)
	}
	lowersAndAgreesWithNode(test, string(source))
}
