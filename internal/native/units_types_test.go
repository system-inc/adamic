package native

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/boundedrun"
)

func TestSplitGeneratedTypesSharedAcrossUnits(t *testing.T) {
	t.Parallel()
	declarations := []string{
		"typedef struct adamic_generated_link adamic_generated_link;",
		"struct adamic_generated_link { int value; };",
		"typedef struct { adamic_generated_link *link; int values[2]; } adamic_generated_frame_0;",
		"typedef adamic_generated_frame_0 adamic_generated_frame_alias;",
	}
	source := "#include \"adamic.h\"\n#include <stdio.h>\n" + strings.Join(declarations, "\n") + `
static int adamic_read(adamic_generated_frame_0 *frame) { return frame->link->value + frame->values[0] + frame->values[1]; }
int main(void) { adamic_generated_link link = {30}; adamic_generated_frame_alias frame = {&link, {5, 7}}; int result = adamic_read(&frame); printf("%d\n", result); return result == 42 ? 0 : 1; }
`
	header, units, err := splitC(source)
	if err != nil {
		t.Fatal(err)
	}
	previous := -1
	for _, declaration := range declarations {
		position := strings.Index(header, declaration)
		if strings.Count(header, declaration) != 1 || position <= previous {
			t.Fatalf("missing, repeated, or reordered type %q:\n%s", declaration, header)
		}
		previous = position
		for _, unit := range units {
			if strings.Contains(unit.source, declaration) {
				t.Fatalf("type has private ownership in %s", unit.name)
			}
		}
	}
	if len(units) < 2 {
		t.Fatalf("expected function and main units, got %d", len(units))
	}
	binary := filepath.Join(t.TempDir(), "types")
	if err := Build(source, binary, Options{Split: true, Sanitize: true, Jobs: 2}); err != nil {
		t.Fatal(err)
	}
	command, release := boundedrun.Command(5*time.Second, binary)
	defer release()
	if output, err := command.CombinedOutput(); err != nil || string(output) != "42\n" {
		t.Fatalf("shared types: %v\n%s", err, output)
	}
}

func TestSplitTypesDoNotAdmitUnownedStorage(t *testing.T) {
	t.Parallel()
	for _, declaration := range []string{
		"struct adamic_generated_link { int value; } adamic_storage;",
		"struct adamic_generated_link *adamic_storage;",
		"int adamic_storage;",
	} {
		if _, _, err := splitC(declaration + "\nint main(void) { return 0; }\n"); err == nil {
			t.Fatalf("unowned storage admitted: %s", declaration)
		}
	}
}
