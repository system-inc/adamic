package oracle

import (
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
	// Model classifying a Stats receiver as an unrelated Dirent kind.
	// Returning false keeps execution clean, so only Node output can catch it.
	changed := strings.ReplaceAll(source, "adamic_node_fs_dirent_is(", "mutant_static_dirent_is(")
	if changed == source {
		t.Fatal("static dispatch mutant changed nothing")
	}
	helper := `
static bool mutant_static_dirent_is(const adamic_object *value, const char *method) {
    if (adamic_fs_file_is_stats(value)) return false;
    return adamic_node_fs_dirent_is(value, method);
}
`
	changed = strings.Replace(changed, `#include "adamic.h"`, "#include \"adamic.h\"\n#include <string.h>\n"+helper, 1)
	binary := filepath.Join(sharedDirectory(t), "static-dispatch-mutant")
	if err := native.Build(changed, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	how := inputRun{directory: filepath.Dir(path)}
	truth := onNodeWith(t, how, path)
	actual := executeInput(t, how, nodeFSDirectorySanitizerEnvironment(), binary)
	if truth.exitCode != 0 || len(truth.stderr) != 0 || actual.exitCode != 0 || len(actual.stderr) != 0 {
		t.Fatalf("mutant must finish cleanly, Node=%+v mutant=%+v", truth, actual)
	}
	if disagreement(truth, actual) != "stdout differs" {
		t.Fatal("Node failed to catch static dispatch")
	}
	t.Log("Stats dispatched as false; caught only by Node stdout, sanitizers and leaks clean")
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
	// Gate scratch paths can exceed sockaddr_un's limit. Keep that case in
	// every run; bind the socket by its relative name from a child directory.
	root := filepath.Join(shared, strings.Repeat("long-scratch-", 10), "kinds")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatal(err)
	}
	if err := syscall.Mkfifo(filepath.Join(root, "fifo"), 0600); err != nil {
		t.Fatal(err)
	}
	socket := executeInput(t, inputRun{directory: root}, nil, "node", "--input-type=module", "-e",
		`import net from 'node:net'; const server = net.createServer(); server.listen('socket', () => process.exit(0));`)
	if socket.exitCode != 0 || len(socket.stdout) != 0 || len(socket.stderr) != 0 {
		t.Fatalf("create socket fixture: %+v", socket)
	}
	information, err := os.Lstat(filepath.Join(root, "socket"))
	if err != nil || information.Mode()&os.ModeSocket == 0 {
		t.Fatalf("want a real socket fixture: %v, %v", information, err)
	}
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
		if err := native.Build(native.C(program), binary, native.Options{Target: "wasm32-wasi"}); err != nil {
			t.Fatal(err)
		}
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
		if err := native.Build(source, mutant, native.Options{Target: "wasm32-wasi"}); err != nil {
			t.Fatal(err)
		}
		actual := executeInput(t, how, nil, "node", "--disable-warning=ExperimentalWarning", runner, mutant, root)
		if actual.exitCode != 0 || len(actual.stderr) != 0 || disagreement(truth, actual) != "stdout differs" {
			t.Fatalf("lossy-kind mutant must be caught only by Node stdout: exit %d stderr %q", actual.exitCode, actual.stderr)
		}
		t.Log("missing FIFO/socket refusal mutant caught only by Node stdout comparison")
	}
}
