package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func TestNodeFSDirectoryUnionMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/node_fs_directory_union.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	// Reproduce choosing Dirent's layout for every receiver from the union's
	// first method declaration, including a Stats value at runtime.
	changed := strings.ReplaceAll(source, "adamic_node_fs_dirent_is(", "mutant_static_dirent_is(")
	if changed == source {
		t.Fatal("static dispatch mutant changed nothing")
	}
	helper := `
static bool mutant_static_dirent_is(const adamic_object *value, const char *method) {
    static adamic_slot_cache cache;
    double wanted = strcmp(method, "isFile") == 0 ? 1 : strcmp(method, "isDirectory") == 0 ? 2 : 3;
    return adamic_object_field(value, "type", &cache)->number == wanted;
}
`
	changed = strings.Replace(changed, `#include "adamic.h"`, "#include \"adamic.h\"\n#include <string.h>\n"+helper, 1)
	binary := filepath.Join(sharedDirectory(t), "static-dispatch-mutant")
	if err := native.Build(changed, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	how := inputRun{directory: filepath.Dir(path)}
	truth := onNodeWith(t, how, path)
	actual := executeInput(t, how, nil, binary)
	if truth.exitCode != 0 || actual.exitCode != 70 || !strings.Contains(string(actual.stderr), "a field the checker proved is there is missing") {
		t.Fatalf("mutant must reproduce layout panic, Node=%+v mutant=%+v", truth, actual)
	}
	if disagreement(truth, actual) == "" {
		t.Fatal("Node failed to catch static dispatch")
	}
	t.Log("static first-member dispatch builds and runs, then Node comparison catches its native layout panic")
}
