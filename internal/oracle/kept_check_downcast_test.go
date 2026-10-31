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
// a generic function. Each program's whole diagnostic is pinned in its .refused file. Mutant:
// keptCheckDowncastRefusal admitting every cast lowers all three.
func TestKeptCheckDowncastsAreRefused(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/kept_check_downcast_refused"
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
	if programs != 3 {
		t.Fatalf("%d programs in %s, want 3: a program added or removed needs this count changed with it", programs, directory)
	}
}
