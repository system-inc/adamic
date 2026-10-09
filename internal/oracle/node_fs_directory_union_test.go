package oracle

import (
	"net"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func TestNodeFSDirectoryUnionMutant(t *testing.T) {
	t.Parallel()
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

// Real special files exercise true results for the shared predicates too.
func TestNodeFSDirectoryUnionSpecialKinds(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/node_fs_directory_union.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	shared := sharedDirectory(t)
	root := filepath.Join(shared, "kinds")
	if err := os.Mkdir(root, 0755); err != nil {
		t.Fatal(err)
	}
	if err := syscall.Mkfifo(filepath.Join(root, "fifo"), 0600); err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("unix", filepath.Join(root, "socket"))
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	if err := os.Symlink("/dev/null", filepath.Join(root, "character")); err != nil {
		t.Fatal(err)
	}
	how := inputRun{directory: filepath.Dir(path), arguments: []string{root}}
	truth := onNodeWith(t, how, path)
	got, binary := inputNatively(t, how, program, shared)
	backend := inputBackend(t, how, program, shared)
	for name, result := range map[string]run{"native": got, "JavaScript": backend} {
		if difference := disagreement(truth, result); difference != "" {
			t.Fatalf("%s: %s\nNode=%+v got=%+v", name, difference, truth, result)
		}
	}
	if leaked := inputLeaks(t, func() inputRun { return how }, program, binary); leaked != "" {
		t.Fatal(leaked)
	}
	if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
		binary := filepath.Join(shared, "kinds.wasm")
		buildWASI(t, native.C(program), binary)
		runner, err := filepath.Abs(filepath.Join(repository, "oracle/wasi.mjs"))
		if err != nil {
			t.Fatal(err)
		}
		result := executeInput(t, how, nil, "node", "--disable-warning=ExperimentalWarning", runner, binary, root)
		if result.exitCode != 70 || string(result.stderr) != "adamic: panic: wasm32-wasi: fs.isFIFO/isSocket cannot distinguish FIFOs from sockets\n" {
			t.Fatalf("WASI special-kind refusal: exit %d stderr %q", result.exitCode, result.stderr)
		}
		t.Log("WASI explicitly refuses the ambiguous FIFO/socket kind")
		// Remove only that new guard while preserving runtime shape dispatch.
		// The mutant returns the host's lossy file kind, silently calling a FIFO
		// a socket. Node must catch it by stdout, with a clean successful run.
		source := strings.ReplaceAll(native.C(program), "adamic_node_fs_dirent_is(", "mutant_lossy_kind(")
		helper := `
#include <string.h>
#include <sys/stat.h>
static bool mutant_lossy_kind(const adamic_object *value, const char *method) {
    if (strcmp(method, "isFIFO") != 0 && strcmp(method, "isSocket") != 0) {
        return adamic_node_fs_dirent_is(value, method);
    }
    static adamic_slot_cache mode_cache, type_cache;
    if (adamic_fs_file_is_stats(value)) {
        mode_t mode = (mode_t)adamic_object_field(value, "_fsFileMode", &mode_cache)->number;
        return strcmp(method, "isFIFO") == 0 ? S_ISFIFO(mode) : S_ISSOCK(mode);
    }
    return adamic_object_field(value, "type", &type_cache)->number == (strcmp(method, "isFIFO") == 0 ? 7 : 8);
}
`
		source = strings.Replace(source, `#include "adamic.h"`, "#include \"adamic.h\"\n"+helper, 1)
		mutant := filepath.Join(shared, "lossy-kind.wasm")
		buildWASI(t, source, mutant)
		actual := executeInput(t, how, nil, "node", "--disable-warning=ExperimentalWarning", runner, mutant, root)
		if actual.exitCode != 0 || len(actual.stderr) != 0 || disagreement(truth, actual) != "stdout differs" {
			t.Fatalf("lossy-kind mutant must be caught only by Node stdout: exit %d stderr %q", actual.exitCode, actual.stderr)
		}
		t.Log("missing FIFO/socket refusal mutant caught only by Node stdout comparison")
	}
}
