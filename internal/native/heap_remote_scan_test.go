package native

import (
	"bytes"
	"context"
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Each profile builds its shipping runtime once. Only heap.c differs in the two
// mutants. A drain counter observes work directly rather than timing empty scans.
func remoteScanProduct(t *testing.T, sanitized bool) string {
	t.Helper()
	files, flags, compiler, version, err := runtimeInputs("", Options{Sanitize: sanitized, slabs: true}, []string{"-DADAMIC_SLAB_QUARANTINE_OFF"})
	if err != nil {
		t.Fatal(err)
	}
	headers, err := headersRead(files, flags, compiler)
	if err != nil {
		t.Fatal(err)
	}
	inputs := buildcache.Inputs{Name: fmt.Sprintf("heap-remote-scan-%t", sanitized), Files: []string{"internal/native/runtime", "internal/native/heap_remote_scan_test.go", "internal/native/slab_quarantine_test.go", "internal/native/runtime_statics_parallel_test.go", "internal/native/product_compile_test.go", "internal/native/testdata/heap_remote_scan.c", "internal/native/native.go", "internal/native/library.go", "internal/native/target.go"}, Flags: portableKeyParts(files, flags, compiler, headers), Toolchain: []string{buildcache.Tool(compiler, "--version"), buildcache.Tool(archiverName(compiler), "--version")}}
	return buildcache.Product(t, inputs, func(directory string) error {
		source, err := os.ReadFile("testdata/heap_remote_scan.c")
		if err != nil {
			return err
		}
		cache, err := cacheDirectory("heap-remote-scan")
		if err != nil {
			return err
		}
		var baseline string
		for _, variant := range []struct{ name, guard string }{{"control", ""}, {"always", "true"}, {"never", "false"}} {
			snapshot := append([]runtimeFile(nil), files...)
			for i := range snapshot {
				if snapshot[i].name != "heap.c" {
					continue
				}
				text := string(snapshot[i].contents)
				for _, edit := range [][2]string{
					{"static void drain_remote(chunk *each);", "static size_t test_remote_drains;\nsize_t adamic_test_remote_drains(void) { return test_remote_drains; }\nstatic void drain_remote(chunk *each);"},
					{"static void drain_remote(chunk *each) {", "static void drain_remote(chunk *each) {\n test_remote_drains++;"},
				} {
					if strings.Count(text, edit[0]) != 1 {
						return fmt.Errorf("missing drain seam %q", edit[0])
					}
					text = strings.Replace(text, edit[0], edit[1], 1)
				}
				if variant.guard != "" {
					old := "atomic_load_explicit(&thread_count, memory_order_relaxed) > 1"
					if strings.Count(text, old) != 1 {
						return fmt.Errorf("missing gate")
					}
					text = strings.Replace(text, old, variant.guard, 1)
				}
				snapshot[i].contents = []byte(text)
			}
			library := baseline
			if variant.guard == "" {
				library, err = cachedRuntime(snapshot, flags, compiler, version, cache)
				if err != nil {
					return err
				}
				baseline = library
			} else {
				// Recompile only the mutated translation unit against the same headers.
				scratch := t.TempDir()
				for _, file := range snapshot {
					if file.name == "heap.c" {
						if err := os.WriteFile(filepath.Join(scratch, "heap.c"), file.contents, 0600); err != nil {
							return err
						}
					}
				}
				object := filepath.Join(scratch, "heap.o")
				arguments := append(append([]string{}, flags...), "-I", filepath.Dir(baseline), "-c", filepath.Join(scratch, "heap.c"), "-o", object)
				staticsCompile(t, compiler, arguments...)
				data, err := os.ReadFile(baseline)
				if err != nil {
					return err
				}
				library = filepath.Join(scratch, "runtime.a")
				if err := os.WriteFile(library, data, 0600); err != nil {
					return err
				}
				staticsCompile(t, archiverName(compiler), "r", library, object)
			}
			binary := buildHarnessWithLibrary(t, string(source), append(append([]string{}, flags...), "-I", filepath.Dir(baseline)), compiler, library)
			data, err := os.ReadFile(binary)
			if err != nil {
				return err
			}
			if err = os.WriteFile(filepath.Join(directory, variant.name), data, 0755); err != nil {
				return err
			}
		}
		return nil
	})
}
func TestProduct_RemoteScanRelease(t *testing.T)   { t.Parallel(); remoteScanProduct(t, false) }
func TestProduct_RemoteScanSanitized(t *testing.T) { t.Parallel(); remoteScanProduct(t, true) }
func checkRemoteScan(t *testing.T, sanitized bool, variant string) {
	t.Helper()
	binary := filepath.Join(remoteScanProduct(t, sanitized), variant)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	if ctx.Err() != nil || command.ProcessState == nil {
		t.Fatalf("%s did not finish: %v", variant, err)
	}
	got := command.ProcessState.ExitCode()
	want, exit := "serial skip; remote reuse\n", 0
	if variant == "always" {
		want, exit = "serial scan\n", 1
	}
	if variant == "never" {
		want, exit = "remote slots lost\n", 4
	}
	if got != exit || stdout.String() != want || stderr.Len() != 0 {
		t.Fatalf("%s: exit %d stdout %q stderr %s", variant, got, stdout.String(), stderr.String())
	}
	t.Logf("%s: exit %d, %q", variant, got, stdout.String())
}
func TestSerialSlabsSkipRemoteScans(t *testing.T) { t.Parallel(); checkRemoteScan(t, false, "control") }
func TestSerialSlabsSkipRemoteScansSanitized(t *testing.T) {
	t.Parallel()
	checkRemoteScan(t, true, "control")
}
func TestRemoteScanAlwaysMutant(t *testing.T) { t.Parallel(); checkRemoteScan(t, false, "always") }
func TestRemoteScanNeverMutant(t *testing.T)  { t.Parallel(); checkRemoteScan(t, false, "never") }
func TestRemoteScanAlwaysMutantSanitized(t *testing.T) {
	t.Parallel()
	checkRemoteScan(t, true, "always")
}
func TestRemoteScanNeverMutantSanitized(t *testing.T) {
	t.Parallel()
	checkRemoteScan(t, true, "never")
}
