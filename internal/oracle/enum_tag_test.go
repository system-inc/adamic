package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// tsc's node shape compiles: kind tells the members apart, and the open flags field every member
// shares the same type of rests no refinement (#drwm5pj). Mutant: sharedByEveryMember returning
// false, which refuses it as an open enum tag.
func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/enum_tag_passenger.a", true, false})
	// A union whose tag can lie, refined only by steps its incoming types prove: a field compared
	// directly, a switch on it, a null test, an unrelated condition, and a comparison after one
	// member is left (#drwm5pj part 1). Mutant: enumTagRefinementProven returning false refuses it.
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/enum_tag_flow_admitted.a", true, false})
}

// A refinement by a tag that can't prove the layout is refused. An open numeric enum tag is
// refused wherever the refinement lands: on an intersection member, on a smaller union of members,
// and when the union's other member is an intersection. Tags from two enums that share a value
// are refused too. Each program prints text1 on Node, and main's native build printed 1, reading
// the string payload as the number its refined type promised. Each program's whole diagnostic is
// pinned in its .refused file. Mutants: enumRefusal counting and refining only plain object types,
// as it did, lowers the first three; enumTagsShareValue returning false lowers both cross_enum
// programs; keeping only the first enum seen for a value lowers cross_enum_union;
// sharedByEveryMember passing over intersection members lowers shared_beside_intersection.
// The other eight are part 1's walk (#drwm5pj): a refinement is admitted only when every step
// that narrowed it is a recognized comparison its incoming type proves. alias, destructured,
// const_key, switch_true, optional and getter_tag are steps it doesn't recognize; join reaches the
// read on a kind path that proves the layout and an alias path that doesn't; sequential narrows by
// kind and then by an open tag. Mutants: a step the walk doesn't recognize passing lowers the six,
// and all but getter_tag then print 1 natively where Node prints text1; a recognized comparison
// passing without judging its incoming type lowers sequential, intersection, subset_union,
// intersection_member, shared_beside_intersection and both cross_enum programs; a label walking
// only its first path lowers join, which then prints 1.
func TestEnumTagRefinementsAreRefused(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/enum_tag_refused"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".a") {
			continue
		}
		programs++
		path := directory + "/" + name
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			absolute, err := filepath.Abs(filepath.Join(repository, path))
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowered(t, absolute)
			var refused *lower.Refused
			if !errors.As(err, &refused) {
				t.Fatalf("want a refusal, got %v", err)
			}
			expected, err := os.ReadFile(strings.TrimSuffix(absolute, ".a") + ".refused")
			if err != nil {
				t.Fatal(err)
			}
			got := strings.ReplaceAll(refused.Error(), strings.TrimSuffix(absolute, path), "")
			if got != strings.TrimSpace(string(expected)) {
				t.Fatalf("diagnostic: got %q, want %q", got, strings.TrimSpace(string(expected)))
			}
		})
	}
	if programs != 14 {
		t.Fatalf("%d programs in %s, want 14: a program added or removed needs this count changed with it", programs, directory)
	}
}
