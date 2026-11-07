package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// Every road from a plain number into a member-literal enum slot is refused (#w7w2asn, Language's
// attack on #drwm5pj). Six forms slipped past the slot refusal on main: an object literal written
// where a union goes (plain, nested, in an array, after a spread, as a conditional), judged against
// the union's property, the whole open enum, instead of the member it's built as; and a compound
// write after narrowing. Each made native panic "compiler bug: a field the checker proved is there is
// missing" where Node prints NaN. increment.ts is compound's ++ sibling, and the other nine pin the
// forms that already refused. Each program's whole diagnostic is pinned in its .refused file.
// Mutants: memberLiteralSlotWidened returning no widening lowers the five object-literal forms
// (plain.ts among them, Language's n09); literalUpdateSlot returning nil lowers compound and
// increment.
func TestEnumLiteralSlotsAreRefused(test *testing.T) {
	test.Parallel()
	directory := "internal/oracle/testdata/enum_literal_slot_refused"
	entries, readDirectoryError := os.ReadDir(filepath.Join(repository, directory))
	if readDirectoryError != nil {
		test.Fatal(readDirectoryError)
	}
	programs := 0
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".ts") {
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
			expected, readError := os.ReadFile(strings.TrimSuffix(absolute, ".ts") + ".refused")
			if readError != nil {
				test.Fatal(readError)
			}
			got := strings.ReplaceAll(refused.Error(), strings.TrimSuffix(absolute, path), "")
			if got != strings.TrimSpace(string(expected)) {
				test.Fatalf("diagnostic: got %q, want %q", got, strings.TrimSpace(string(expected)))
			}
		})
	}
	if programs != 16 {
		test.Fatalf(
			"%d programs in %s, want 16: a program added or removed needs this count changed with it",
			programs,
			directory,
		)
	}
}
