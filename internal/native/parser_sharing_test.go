package native

import (
	"os/exec"
	"path/filepath"
	"testing"
)

// A graph-publication invariant: NodeArray metadata is interior storage, but its strings are shared values.
func TestParserNodeArrayMetadataSharing(test *testing.T) {
	test.Parallel()
	const source = `#include "adamic.h"
#include <stdio.h>
static const char *const names[] = {"cache"};
static const bool references[] = {true};
static const adamic_field_kind kinds[] = {adamic_field_reference};
static const adamic_shape shape = {1, names, references, NULL, kinds};
int main(void) {
 adamic_array *array = adamic_node_array_make(0, adamic_ir_Number, &shape);
 adamic_string *text = adamic_string_from_number(12345);
 array->metadata->slots[0].reference = text;
 adamic_object_initialized(array->metadata)[0] = 1;
 adamic_object_publish(array->metadata, 0);
 adamic_share(array);
 if (!adamic_is_shared(&text->heap)) { fputs("metadata child not shared\n", stderr); return 1; }
 if (adamic_reference_count(&array->metadata->heap) != 0) { fputs("interior metadata retained\n", stderr); return 2; }
 adamic_release(array);
 puts("metadata graph shared");
 return 0;
}
`
	binary := filepath.Join(test.TempDir(), "metadata-sharing")
	if buildError := Build(source, binary, Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	output, combinedOutputError := exec.Command(binary).CombinedOutput()
	if combinedOutputError != nil || string(output) != "metadata graph shared\n" {
		test.Fatalf("metadata graph: %v\n%s", combinedOutputError, output)
	}
}
