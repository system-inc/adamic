package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// The .ts twins of kept_check_downcast_refused (#m8dvvmf): each cast stays a checked cast in .ts, its
// test passes, and each lowers and agrees with Node natively and through the JavaScript backend.
func init() {
	for _, path := range []string{
		"internal/oracle/testdata/kept_check_downcast/union_member.ts",
		"internal/oracle/testdata/kept_check_downcast/interface.ts",
		"internal/oracle/testdata/kept_check_downcast/generic.ts",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}

// In .a a downcast whose checks no proof erases is refused, naming the narrowing that says the same thing
// (item 8 stage 2, #m8dvvmf): union_member casts a discriminated union to a member, interface casts a
// Node to an Identifier whose tag a Node can fail, and generic casts a generic union to its member inside
// a generic function. distinct_tag_cast and later_tag are stage 1's admissions (#q97a1rr), reversed on
// purpose: later_tag's message names kind, the tag .a tests after passing over the vacuous version.
// Language's attacks on the rule: generic_constraint casts a type parameter's value inside a generic, alias
// casts through an alias of Extract, and satisfies casts inside satisfies, each refused as a member cast;
// class_through_interface and class_union_interface cast to a class from a structural interface, which has
// no nominal ancestry to test, so each is a cast the runtime can't check. Each program's whole diagnostic
// is pinned in its .refused file. Mutants: keptCheckDowncastRefusal admitting every cast lowers the eight
// member and view casts; withinLiterals calling every tag vacuous gives distinct_tag_cast and later_tag
// stage 1's refusal where stage 2's is pinned.
func TestKeptCheckDowncastsAreRefused(test *testing.T) {
	test.Parallel()
	directory := "internal/oracle/testdata/kept_check_downcast_refused"
	entries, readDirectoryError := os.ReadDir(filepath.Join(repository, directory))
	if readDirectoryError != nil {
		test.Fatal(readDirectoryError)
	}
	programs := 0
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".a") {
			continue
		}
		programs++
		path := directory + "/" + name
		test.Run(name, func(test *testing.T) {
			test.Parallel()
			absolute, absolutePathError := filepath.Abs(filepath.Join(repository, path))
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			_, loweredError := lowered(test, absolute)
			var refused *lower.Refused
			if !errors.As(loweredError, &refused) {
				test.Fatalf("want a refusal, got %v", loweredError)
			}
			expected, readError := os.ReadFile(strings.TrimSuffix(absolute, ".a") + ".refused")
			if readError != nil {
				test.Fatal(readError)
			}
			got := strings.ReplaceAll(refused.Error(), strings.TrimSuffix(absolute, path), "")
			if got != strings.TrimSpace(string(expected)) {
				test.Fatalf("diagnostic: got %q, want %q", got, strings.TrimSpace(string(expected)))
			}
		})
	}
	if programs != 10 {
		test.Fatalf(
			"%d programs in %s, want 10: a program added or removed needs this count changed with it",
			programs,
			directory,
		)
	}
}
