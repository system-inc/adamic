package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func TestNodeFSDirectoryUnionMutant(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/node_fs_directory_union.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	source := native.C(program)
	// Model classifying a Stats receiver as an unrelated Dirent kind.
	// Returning false keeps execution clean, so only Node output can catch it.
	changed := strings.ReplaceAll(source, "adamic_node_fs_dirent_is(", "mutant_static_dirent_is(")
	if changed == source {
		test.Fatal("static dispatch mutant changed nothing")
	}
	helper := `
static bool mutant_static_dirent_is(const adamic_object *value, const char *method) {
    if (adamic_fs_file_is_stats(value)) return false;
    return adamic_node_fs_dirent_is(value, method);
}
`
	changed = strings.Replace(changed, `#include "adamic.h"`, "#include \"adamic.h\"\n#include <string.h>\n"+helper, 1)
	binary := filepath.Join(sharedDirectory(test), "static-dispatch-mutant")
	if buildError := native.Build(changed, binary, native.Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	how := inputRun{directory: filepath.Dir(path)}
	truth := onNodeWith(test, how, path)
	actual := executeInput(test, how, nodeFSDirectorySanitizerEnvironment(), binary)
	if truth.exitCode != 0 || len(truth.stderr) != 0 || actual.exitCode != 0 || len(actual.stderr) != 0 {
		test.Fatalf("mutant must finish cleanly, Node=%+v mutant=%+v", truth, actual)
	}
	if disagreement(truth, actual) != "stdout differs" {
		test.Fatal("Node failed to catch static dispatch")
	}
	test.Log("Stats dispatched as false; caught only by Node stdout, sanitizers and leaks clean")
}

// Real special files exercise true results for the shared predicates too.
func TestNodeFSDirectoryUnionSpecialKinds(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/node_fs_directory_union.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	shared := sharedDirectory(test)
	// Gate scratch paths can exceed sockaddr_un's limit. Keep that case in
	// every run; bind the socket by its relative name from a child directory.
	root := filepath.Join(shared, strings.Repeat("long-scratch-", 10), "kinds")
	if createDirectoryError := os.MkdirAll(root, 0755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	if mkfifoError := syscall.Mkfifo(filepath.Join(root, "fifo"), 0600); mkfifoError != nil {
		test.Fatal(mkfifoError)
	}
	socket := executeInput(test, inputRun{directory: root}, nil, "node", "--input-type=module", "-e",
		`import net from 'node:net'; const server = net.createServer(); server.listen('socket', () => process.exit(0));`)
	if socket.exitCode != 0 || len(socket.stdout) != 0 || len(socket.stderr) != 0 {
		test.Fatalf("create socket fixture: %+v", socket)
	}
	information, statError := os.Lstat(filepath.Join(root, "socket"))
	if statError != nil || information.Mode()&os.ModeSocket == 0 {
		test.Fatalf("want a real socket fixture: %v, %v", information, statError)
	}
	if symlinkError := os.Symlink("/dev/null", filepath.Join(root, "character")); symlinkError != nil {
		test.Fatal(symlinkError)
	}
	how := inputRun{directory: filepath.Dir(path), arguments: []string{root}}
	truth := onNodeWith(test, how, path)
	got, binary := inputNatively(test, how, program, shared)
	backend := inputBackend(test, how, program, shared)
	for name, result := range map[string]run{"native": got, "JavaScript": backend} {
		if difference := disagreement(truth, result); difference != "" {
			test.Fatalf("%s: %s\nNode=%+v got=%+v", name, difference, truth, result)
		}
	}
	if leaked := inputLeaks(test, func() inputRun { return how }, program, binary); leaked != "" {
		test.Fatal(leaked)
	}
	if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
		binary := filepath.Join(shared, "kinds.wasm")
		if buildError := native.Build(
			native.C(program),
			binary,
			native.Options{Target: "wasm32-wasi"},
		); buildError != nil {
			test.Fatal(buildError)
		}
		runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle/wasi.mjs"))
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		result := executeInput(test, how, nil, "node", "--disable-warning=ExperimentalWarning", runner, binary, root)
		if result.exitCode != 70 || string(
			result.stderr,
		) != "adamic: panic: wasm32-wasi: fs.isFIFO/isSocket cannot distinguish FIFOs from sockets\n" {
			test.Fatalf("WASI special-kind refusal: exit %d stderr %q", result.exitCode, result.stderr)
		}
		test.Log("WASI explicitly refuses the ambiguous FIFO/socket kind")
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
		if buildError := native.Build(source, mutant, native.Options{Target: "wasm32-wasi"}); buildError != nil {
			test.Fatal(buildError)
		}
		actual := executeInput(test, how, nil, "node", "--disable-warning=ExperimentalWarning", runner, mutant, root)
		if actual.exitCode != 0 || len(actual.stderr) != 0 || disagreement(truth, actual) != "stdout differs" {
			test.Fatalf(
				"lossy-kind mutant must be caught only by Node stdout: exit %d stderr %q",
				actual.exitCode,
				actual.stderr,
			)
		}
		test.Log("missing FIFO/socket refusal mutant caught only by Node stdout comparison")
	}
}
