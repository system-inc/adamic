package lower

import (
	"errors"
	"os"
	"strings"
	"testing"
)

// Each row's overload is refused at a direct call that resolves to it, run's, since a declaration alone
// no longer stops (#4czdn4q). The first three places are measured (J-7bkh7xq). Result covariance's brand is
// phantom (void): with true, run's Path parameter met the brand rule at 1:13 first (J-2e48bb1), which the
// declaration's stop used to hide.
func TestCensusOverloadRelation(t *testing.T) {
	t.Parallel()
	probes := []struct{ name, source, diagnostic, where string }{
		{"parameter contravariance", `interface Animal { readonly name: string; }
interface Dog extends Animal { readonly bark: string; }
interface Handler { handle(animal: Animal): void; }
interface DogHandler { handle(dog: Dog): void; }
function serve(handler: DogHandler): string;
function serve(handler: Handler): string { handler.handle({ name: 'cat' }); return 'written'; }
function run(handler: DogHandler): string { return serve(handler); }
`, "overload 1 of serve parameter handler", "main.a:7:52"},
		{"generic parameter constraints", `interface Animal { readonly name: string; }
interface Dog extends Animal { readonly bark: string; }
interface Handler { handle(animal: Animal): void; }
interface DogHandler { handle(dog: Dog): void; }
function serve<T extends DogHandler>(handler: T): string;
function serve<U extends Handler>(handler: U): string { handler.handle({ name: 'cat' }); return 'written'; }
function run(handler: DogHandler): string { return serve(handler); }
`, "overload 1 of serve type parameter T", "main.a:7:52"},
		{"mutable parameters", `interface Animal { readonly name: string; }
interface Dog extends Animal { readonly bark: string; }
function serve(dogs: Dog[]): string;
function serve(animals: Animal[]): string { animals.push({ name: 'cat' }); return 'written'; }
function run(dogs: Dog[]): string { return serve(dogs); }
`, "overload 1 of serve parameter dogs", "main.a:5:44"},
		{"result covariance", `type Path = string & { readonly __pathBrand: void };
function ensureTrailingDirectorySeparator(path: Path): Path;
function ensureTrailingDirectorySeparator(path: string): string;
function ensureTrailingDirectorySeparator(path: string) { return path + '/'; }
function run(path: Path): Path { return ensureTrailingDirectorySeparator(path); }
`, "overload 1 of ensureTrailingDirectorySeparator result", "main.a:5:41"},
	}
	for _, probe := range probes {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(refused.What, probe.diagnostic) || !strings.HasSuffix(refused.Where, probe.where) {
				t.Fatalf("expected strict overload refusal naming %q at %s; got %v", probe.diagnostic, probe.where, err)
			}
		})
	}
}

// The ruling's own proof (#4czdn4q): result covariance's set, declared with no call, lowers. Its overload 1
// is unserved, and nothing resolves to it, so nothing stops. Mutant: censusOverloads returning the set's
// first error at the declaration again, and this stops at line 2.
func TestCensusOverloadUncalledLowers(t *testing.T) {
	t.Parallel()
	if _, err := lowerSource(t, `type Path = string & { readonly __pathBrand: void };
function ensureTrailingDirectorySeparator(path: Path): Path;
function ensureTrailingDirectorySeparator(path: string): string;
function ensureTrailingDirectorySeparator(path: string) { return path + '/'; }
`); err != nil {
		t.Fatalf("an uncalled set with an unserved overload stopped: %v", err)
	}
}

func TestCensusSmallFiniteKeyRead(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, `const wildcardMatchers = { files: 'file', directories: 'directory', exclude: 'excluded' };
function pick(usage: 'files' | 'directories' | 'exclude'): string { return wildcardMatchers[usage]; }
console.log(pick('files'));
`)
}

func TestCensusRestMutableElements(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `interface Animal { readonly name: string; }
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
	if !errors.As(err, &refused) {
		t.Fatalf("a rest copy still shares mutable elements; expected refusal, got %v", err)
	}
	t.Logf("rest element relation: %v", refused)
}

// Flow-narrowed never is not an executable arm that fails to return.
func TestCensusBooleanDeadBranch(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("../oracle/testdata/census_boolean_dead_branch.a")
	if err != nil {
		t.Fatal(err)
	}
	lowersAndAgreesWithNode(t, string(source))
}
