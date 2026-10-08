package native_test

import (
	"github.com/system-inc/adamic/internal/native"
	"os/exec"
	"path/filepath"
	"testing"
)

// Graph adoption moves an allocation. Interior pointers must move with it,
// while separately allocated entry buffers must keep their own address.
func TestShapeGraphMapAdoption(t *testing.T) {
	const harness = `#include "adamic.h"
#include "graph_regions.h"
#include <stdio.h>
static int check(size_t count) {
 adamic_map *map = adamic_map_new(false, false);
 for (size_t i = 0; i < count; i++) {
  adamic_map_set(map, (adamic_value){.number = (double)i}, (adamic_value){.number = (double)i + 10});
 }
 adamic_map_entry *external = count > 4 ? map->entries : NULL;
 map = adamic_graph_adopt(map, sizeof *map);
 if (count <= 4 && map->entries != map->small) { return 2; }
 if (count > 4 && map->entries != external) { return 3; }
 for (size_t i = 0; i < count; i++) {
  adamic_value *value = adamic_map_get(map, (adamic_value){.number = (double)i});
  if (value == NULL || value->number != (double)i + 10) { return 4; }
 }
 adamic_release(map);
 return 0;
}
int main(void) {
 static const char *const names[] = {"value"};
 static const bool references[] = {false};
 static const adamic_shape shape = {1, names, references, NULL};
 adamic_object *object = adamic_object_new(&shape);
 object->slots[0].number = 42;
 adamic_object_initialized(object)[0] = 0;
 adamic_object_field_types(object)[0] = 1;
 adamic_object_contracts(object)[0] = 44;
 object = adamic_graph_adopt(object, sizeof *object + sizeof(adamic_value));
 if (object->slots[0].number != 42 || adamic_object_initialized(object)[0] != 0 || adamic_object_field_types(object)[0] != 1 || adamic_object_contracts(object)[0] != 44) { return 5; }
 adamic_release(object);
 for (size_t count = 0; count <= 8; count++) { int result = check(count); if (result != 0) { return result; } }
 puts("map adoption preserved entries");
 return 0;
}
`
	for _, sanitize := range []bool{false, true} {
		binary := filepath.Join(t.TempDir(), "adoption")
		if err := native.Build(harness, binary, native.Options{Sanitize: sanitize}); err != nil {
			t.Fatal(err)
		}
		output, err := exec.Command(binary).CombinedOutput()
		if err != nil || string(output) != "map adoption preserved entries\n" {
			t.Fatalf("sanitize=%t: %v: %s", sanitize, err, output)
		}
	}
}
