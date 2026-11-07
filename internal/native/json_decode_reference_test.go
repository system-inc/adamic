package native

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestJSONDecodeReferenceMatchesIR(test *testing.T) {
	test.Parallel()
	decoder, readError := os.ReadFile("runtime/json_decode.c")
	if readError != nil {
		test.Fatal(readError)
	}
	// Compile the actual decoder in the probe's translation unit to call its
	// static reference(). Rename its public entry point to avoid the archive's.
	var source strings.Builder
	source.WriteString("#define adamic_json_decode probe_json_decode\n")
	source.Write(decoder)
	source.WriteString("\nint main(void) {\n")
	tags := ir.TypeTags()
	names := slices.Sorted(maps.Keys(tags))
	for _, name := range names {
		fmt.Fprintf(&source, "printf(\"%%d\\n\", reference(adamic_ir_%s));\n", name)
	}
	source.WriteString("return 0;\n}\n")
	binary := filepath.Join(test.TempDir(), "reference")
	if buildError := Build(source.String(), binary, Options{}); buildError != nil {
		test.Fatal(buildError)
	}
	stdout, stderr, recordRunError := recordRun(binary)
	if recordRunError != nil || stderr != "" {
		test.Fatalf("reference probe: %v\n%s", recordRunError, stderr)
	}
	answers := strings.Fields(stdout)
	if len(answers) != len(names) {
		test.Fatalf("got %d answers for %d IR representations: %q", len(answers), len(names), stdout)
	}
	for i, name := range names {
		want := "0"
		if tags[name].IsReference() {
			want = "1"
		}
		if answers[i] != want {
			test.Errorf("%s: reference()=%s, IsReference()=%s", name, answers[i], want)
		}
	}
}
