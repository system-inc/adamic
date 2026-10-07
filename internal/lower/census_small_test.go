package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestCensusOverloadRelation(t *testing.T) {
	probes := []struct{ name, source, diagnostic string }{
		{"parameter contravariance", `interface Animal { readonly name: string; }
interface Dog extends Animal { readonly bark: string; }
interface Handler { handle(animal: Animal): void; }
interface DogHandler { handle(dog: Dog): void; }
function serve(handler: DogHandler): string;
function serve(handler: Handler): string { handler.handle({ name: 'cat' }); return 'written'; }
`, "overload 1 of serve parameter handler"},
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
		t.Run(probe.name, func(t *testing.T) {
			_, err := lowerSource(t, probe.source)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(refused.What, probe.diagnostic) {
				t.Fatalf("expected strict overload refusal naming %q; got %v", probe.diagnostic, err)
			}
		})
	}
}

func TestCensusSmallFiniteKeyBoundary(t *testing.T) {
	_, err := lowerSource(t, `const wildcardMatchers = { files: 'file', directories: 'directory', exclude: 'excluded' };
function pick(usage: 'files' | 'directories' | 'exclude'): string { return wildcardMatchers[usage]; }
console.log(pick('files'));
`)
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.Contains(notYet.What, "an ElementAccessExpression") {
		t.Fatalf("expected the computed finite-key access boundary, got %v", err)
	}
}
