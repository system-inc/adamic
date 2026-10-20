package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// A refinement by a tag that can't prove the layout is refused. An open numeric enum tag is
// refused wherever the refinement lands: on an intersection member, on a smaller union of members,
// and when the union's other member is an intersection. Tags from two enums that share a value
// are refused too. Each program prints text1 on Node, and main's native build printed 1, reading
// the string payload as the number its refined type promised. Each program's whole diagnostic is
// pinned in its .refused file. Mutants: enumRefusal counting and refining only plain object types,
// as it did, lowers the first three; enumTagsShareValue returning false lowers cross_enum.
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
	if programs != 4 {
		t.Fatalf("%d programs in %s, want 4: a program added or removed needs this count changed with it", programs, directory)
	}
}
