package oracle

import (
	"context"
	"errors"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/native_runtime"
)

// TestGetterFixturesTemp records, for #n7mgy3s, how Node, the JavaScript backend and the sanitized
// native build run each getter-under-ThisType fixture. It logs every outcome, refusals included,
// and fails only when the harness itself can't run. Temporary: it goes once the evidence is read.
func TestGetterFixturesTemp(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"g1.ts", "g1.a", "g2.ts", "g2.a"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "stage3/meter/evidence/this-parameter-getter", name))
			if err != nil {
				t.Fatal(err)
			}
			oracle := onNode(t, path)
			t.Logf("Node: stdout %q stderr %q exit %d", oracle.stdout, oracle.stderr, oracle.exitCode)
			loaded, err := load.Load([]string{path})
			if err != nil {
				var checkError *load.CheckError
				if errors.As(err, &checkError) {
					t.Logf("checker refuses:\n%s", checkError.Error())
					return
				}
				t.Fatalf("Load: %v", err)
			}
			for _, file := range loaded.Files() {
				rememberImports(path, file.Text(), loaded)
			}
			program, err := lower.Lower(context.Background(), loaded)
			if err != nil {
				t.Logf("lowering refuses: %v", err)
				return
			}
			backend := onJavaScriptBackend(t, program)
			t.Logf("JavaScript: stdout %q stderr %q exit %d", backend.stdout, backend.stderr, backend.exitCode)
			binary := filepath.Join(t.TempDir(), "program")
			if err := native.Build(native.C(program), binary, native_runtime.Options{Sanitize: true}); err != nil {
				t.Logf("native build fails: %v", err)
				return
			}
			var environment []string
			if runtime.GOOS == "linux" {
				environment = []string{"ASAN_OPTIONS=detect_leaks=0"}
			}
			compiled := executeWith(t, environment, binary)
			t.Logf("native: stdout %q stderr %q exit %d", compiled.stdout, compiled.stderr, compiled.exitCode)
			t.Logf("Node vs JavaScript: %q; Node vs native: %q", disagreement(oracle, backend), disagreement(oracle, compiled))
		})
	}
}
