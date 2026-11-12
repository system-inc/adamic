package native

import (
	"path/filepath"
	"testing"
)

// Empty own-key tables belong to both standard error allocations (#z4a9p9e).
func TestErrorJSONEmptyTables(t *testing.T) {
	t.Parallel()
	const code = `#include "adamic.h"
#include "json_stringify.h"
#include <stdio.h>
static adamic_string message = ADAMIC_STRING("x");
static const adamic_json_schema schema = {adamic_json_union, NULL, 0, NULL};
int main(void) {
 for (size_t i = 0; i < 3; i++) {
  adamic_object *error = i == 2 ? adamic_error_new_kind("TypeError", 9, &message) : adamic_error_new(i == 0 ? NULL : &message);
  const adamic_error_json *entry = adamic_error_json_lookup(error->shape);
  if (entry == NULL || entry->count != 0) return 1;
  adamic_string *text = adamic_json_stringify_objects((adamic_value){.reference = error}, &schema, (adamic_value){.reference = NULL}, NULL, (adamic_value){.reference = NULL}, NULL);
  printf("%.*s\n", (int)text->length, text->bytes);
  adamic_release(text); adamic_release(error);
 }
 return 0;
}
`
	binary := filepath.Join(t.TempDir(), "error-json")
	if err := Build(code, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	if got := runWithInput(t, "", binary); got != "{}\n{}\n{}\n" {
		t.Fatalf("got %q", got)
	}
}
