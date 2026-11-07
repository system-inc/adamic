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
func remoteScanProduct(test *testing.T, sanitized bool) string {
	test.Helper()
	files, flags, compiler, version, runtimeInputsError := runtimeInputs(
		"",
		Options{Sanitize: sanitized, slabs: true},
		[]string{"-DADAMIC_SLAB_QUARANTINE_OFF"},
	)
	if runtimeInputsError != nil {
		test.Fatal(runtimeInputsError)
	}
	headers, headersReadError := headersRead(files, flags, compiler)
	if headersReadError != nil {
		test.Fatal(headersReadError)
	}
	inputs := buildcache.Inputs{
		Name: fmt.Sprintf("heap-remote-scan-%t", sanitized),
		Files: []string{
			"internal/native/runtime",
			"internal/native/heap_remote_scan_test.go",
			"internal/native/slab_quarantine_test.go",
			"internal/native/runtime_statics_parallel_test.go",
			"internal/native/product_compile_test.go",
			"internal/native/testdata/heap_remote_scan.c",
			"internal/native/native.go",
			"internal/native/library.go",
			"internal/native/target.go",
		},
		Flags: portableKeyParts(files, flags, compiler, headers),
		Toolchain: []string{
			buildcache.Tool(compiler, "--version"),
			buildcache.Tool(archiverName(compiler), "--version"),
		},
	}
	return buildcache.Product(test, inputs, func(directory string) error {
		source, readError := os.ReadFile("testdata/heap_remote_scan.c")
		if readError != nil {
			return readError
		}
		cache, cacheDirectoryError := cacheDirectory("heap-remote-scan")
		if cacheDirectoryError != nil {
			return cacheDirectoryError
		}
		var baseline string
		for _, variant := range []struct{ name, guard string }{
			{"control", ""},
			{"always", "true"},
			{"never", "false"},
		} {
			snapshot := append([]runtimeFile(nil), files...)
			for i := range snapshot {
				if snapshot[i].name != "heap.c" {
					continue
				}
				text := string(snapshot[i].contents)
				for _, edit := range [][2]string{
					{
						"static void drain_remote(chunk *each);",
						"static size_t test_remote_drains;\nsize_t adamic_test_remote_drains(void) { return test_remote_drains; }\nstatic void drain_remote(chunk *each);",
					},
					{
						"static void drain_remote(chunk *each) {",
						"static void drain_remote(chunk *each) {\n test_remote_drains++;",
					},
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
				var cachedRuntimeError error
				library, cachedRuntimeError = cachedRuntime(snapshot, flags, compiler, version, cache)
				if cachedRuntimeError != nil {
					return cachedRuntimeError
				}
				baseline = library
			} else {
				// Recompile only the mutated translation unit against the same headers.
				scratch := test.TempDir()
				for _, file := range snapshot {
					if file.name == "heap.c" {
						if writeError := os.WriteFile(
							filepath.Join(scratch, "heap.c"),
							file.contents,
							0600,
						); writeError != nil {
							return writeError
						}
					}
				}
				object := filepath.Join(scratch, "heap.o")
				arguments := append(
					append([]string{}, flags...),
					"-I",
					filepath.Dir(baseline),
					"-c",
					filepath.Join(scratch, "heap.c"),
					"-o",
					object,
				)
				staticsCompile(test, compiler, arguments...)
				data, readError := os.ReadFile(baseline)
				if readError != nil {
					return readError
				}
				library = filepath.Join(scratch, "runtime.a")
				if writeError := os.WriteFile(library, data, 0600); writeError != nil {
					return writeError
				}
				staticsCompile(test, archiverName(compiler), "r", library, object)
			}
			binary := buildHarnessWithLibrary(
				test,
				string(source),
				append(append([]string{}, flags...), "-I", filepath.Dir(baseline)),
				compiler,
				library,
			)
			data, readError := os.ReadFile(binary)
			if readError != nil {
				return readError
			}
			if writeError := os.WriteFile(filepath.Join(directory, variant.name), data, 0755); writeError != nil {
				return writeError
			}
		}
		return nil
	})
}
func TestProduct_RemoteScanRelease(test *testing.T)   { test.Parallel(); remoteScanProduct(test, false) }
func TestProduct_RemoteScanSanitized(test *testing.T) { test.Parallel(); remoteScanProduct(test, true) }
func checkRemoteScan(test *testing.T, sanitized bool, variant string) {
	test.Helper()
	binary := filepath.Join(remoteScanProduct(test, sanitized), variant)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	runError := command.Run()
	if ctx.Err() != nil || command.ProcessState == nil {
		test.Fatalf("%s did not finish: %v", variant, runError)
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
		test.Fatalf("%s: exit %d stdout %q stderr %s", variant, got, stdout.String(), stderr.String())
	}
	test.Logf("%s: exit %d, %q", variant, got, stdout.String())
}
func TestSerialSlabsSkipRemoteScans(test *testing.T) {
	test.Parallel()
	checkRemoteScan(test, false, "control")
}
func TestSerialSlabsSkipRemoteScansSanitized(test *testing.T) {
	test.Parallel()
	checkRemoteScan(test, true, "control")
}
func TestRemoteScanAlwaysMutant(test *testing.T) {
	test.Parallel()
	checkRemoteScan(test, false, "always")
}
func TestRemoteScanNeverMutant(test *testing.T) {
	test.Parallel()
	checkRemoteScan(test, false, "never")
}
func TestRemoteScanAlwaysMutantSanitized(test *testing.T) {
	test.Parallel()
	checkRemoteScan(test, true, "always")
}
func TestRemoteScanNeverMutantSanitized(test *testing.T) {
	test.Parallel()
	checkRemoteScan(test, true, "never")
}
