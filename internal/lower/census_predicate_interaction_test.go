package lower

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// This exact parser boundary probe previously reached union lowering after the
// predicates merge. Node observes the declaration; Adamic must refuse its live
// contract before considering the callback's representation.
func TestCensusPredicateInteractionBoundary(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs("testdata/census_marker_live_boundary.a")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	// This is a refusal row, so acceptance agreement would fail before running Node.
	// Use the shared bounded source runner and retain the independent declaration check.
	observation := runAgreementNode(test, path)
	if observation.code != 0 ||
		string(observation.stdout) != "boundary declaration loaded\n" ||
		len(observation.stderr) != 0 {
		test.Fatalf("Node: stdout %q, stderr %q, exit %d", observation.stdout, observation.stderr, observation.code)
	}
	source, readError := os.ReadFile(path)
	if readError != nil {
		test.Fatal(readError)
	}
	_, lowerSourceError := lowerSource(test, string(source))
	var refused *Refused
	if !errors.As(lowerSourceError, &refused) ||
		refused.What != "a type predicate whose return is not proven (there is no body proving this parameter)" ||
		!strings.Contains(refused.Fix, "adamic/no-type-predicate") {
		test.Fatalf("want marker proof boundary diagnostic, got %v", lowerSourceError)
	}
}

func TestCensusPredicateInteractionEscapes(test *testing.T) {
	test.Parallel()
	for index, source := range []string{
		`function apply(callback: (value: number) => value is number): boolean { return callback(1); }
   function guard(value: number): value is number { return typeof value === 'number'; }
   const alias = apply; alias(guard);`,
		`function apply(callback: (value: number) => value is number): typeof callback { return callback; }
   function guard(value: number): value is number { return typeof value === 'number'; }
   apply(guard);`,
		`export function apply(callback: (value: number) => value is number): boolean { return callback(1); }
   function guard(value: number): value is number { return typeof value === 'number'; }
   apply(guard);`,
	} {
		test.Run([]string{"function alias", "callback escape", "exported function"}[index], func(test *testing.T) {
			test.Parallel()
			_, lowerSourceError := lowerSource(test, source)
			var refused *Refused
			if !errors.As(lowerSourceError, &refused) ||
				!strings.Contains(refused.What, "there is no body proving this parameter") {
				test.Fatalf("want unproven live callback refused, got %v", lowerSourceError)
			}
		})
	}
}
