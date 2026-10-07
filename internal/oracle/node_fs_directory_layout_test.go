package oracle

import (
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func TestNodeFSDirectoryRuntimeLayouts(t *testing.T) {
	t.Parallel()
	fixture, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/node_fs_directory_layout.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, fixture)
	if err != nil {
		t.Fatal(err)
	}
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	how := inputRun{directory: root}
	truth := onNodeWith(t, how, fixture)
	shared := sharedDirectory(t)
	backend := inputBackend(t, how, program, shared)
	got, binary := inputNatively(t, how, program, shared)
	for name, result := range map[string]run{"native": got, "JavaScript": backend} {
		if difference := disagreement(truth, result); difference != "" {
			t.Errorf("%s: %s", name, difference)
		}
	}
	if leaked := inputLeaks(t, func() inputRun { return how }, program, binary); leaked != "" {
		t.Fatal(leaked)
	}
	// Omitting the runtime's (kind, path) shape falsely makes path uniform at
	// slot 0, the position of the program's own field. Reproduce that old C.
	generated := native.C(program)
	pattern := regexp.MustCompile(`adamic_object_field\((adamic_[A-Za-z0-9_]+), "path", &adamic_cache_[0-9]+\)`)
	changed := pattern.ReplaceAllString(generated, "(&${1}->slots[0])")
	if changed == generated || strings.Contains(changed, `"path", &adamic_cache_`) {
		t.Fatal("layout mutant changed nothing or missed a path read")
	}
	mutant := filepath.Join(shared, "mutant")
	if err := native.Build(changed, mutant, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	// LeakSanitizer is Linux's: macOS's AddressSanitizer aborts when asked for it.
	var environment []string
	if runtime.GOOS == "linux" {
		environment = []string{"ASAN_OPTIONS=detect_leaks=1"}
	}
	result := executeInput(t, how, environment, mutant)
	if result.exitCode != 0 || len(result.stderr) != 0 {
		t.Fatalf("mutant failed outside comparison: exit %d stderr %s", result.exitCode, result.stderr)
	}
	if difference := disagreement(truth, result); difference != "stdout differs" {
		t.Fatalf("mutant caught by %q", difference)
	}
	t.Log("realPath layout agrees on both backends; slot-0 mutant caught only by Node stdout comparison")
}
