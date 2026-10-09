package oracle

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

var nodeHostCoverageFixtures = []string{"fs_numeric_error", "fs_error_preview", "fs_path_preview", "fs_zero_read", "fs_fd_ownership", "host_directory", "fs_large_roundtrip"}

func nodeHostCoveragePath(name string) string {
	return "internal/oracle/testdata/oct8_nodehost_" + name + ".a"
}

func init() {
	for _, name := range nodeHostCoverageFixtures {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{nodeHostCoveragePath(name), true, false})
	}
}

// The native preload interposes successful short I/O only. Keep Node and the
// JavaScript backend as unchanged oracles; compare raw streams and exit codes.
// LD_PRELOAD/dlsym(RTLD_NEXT) make this an explicitly Linux-only check. The large
// roundtrip fixture also runs ordinarily on every host, without interposition.
func nodeHostShortIO(t *testing.T) string {
	t.Helper()
	if runtime.GOOS != "linux" {
		t.Skip("short I/O interposition requires Linux LD_PRELOAD and RTLD_NEXT")
	}
	output := filepath.Join(t.TempDir(), "short-io.so")
	args := []string{"-shared", "-fPIC", "-std=c11", "-Wall", "-Wextra", "-Werror", "-pedantic", filepath.Join(repository, "internal/oracle/testdata/nodehost_short_io.c"), "-ldl", "-o", output}
	if data, err := exec.Command("clang", args...).CombinedOutput(); err != nil {
		t.Fatalf("short I/O shim: %v\n%s", err, data)
	}
	return output
}

func TestNodeHostShortIO(t *testing.T) {
	t.Parallel()
	shim := nodeHostShortIO(t)
	path, binary, script := sanitized(t, nodeHostCoveragePath("fs_large_roundtrip"))
	truth := onNode(t, path)
	backend := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), script)
	p, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	release := filepath.Join(t.TempDir(), "release")
	if err := native.Build(native.C(p), release, native.Options{}); err != nil {
		t.Fatal(err)
	}
	env := []string{"LD_PRELOAD=" + shim, "ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}
	for name, got := range map[string]run{"javascript": backend, "sanitized": executeWith(t, env, binary), "release": executeWith(t, env, release)} {
		if d := disagreement(truth, got); d != "" {
			t.Fatalf("%s: %s\nNode %q %q\ngot %q %q", name, d, truth.stdout, truth.stderr, got.stdout, got.stderr)
		}
	}
}

// Runtime copies isolate each mutation. No test edits the checkout or changes
// the generated program, so its unchanged source on Node is the only oracle.
func nodeHostRuntimeMutant(t *testing.T, source, file, scope, before, after string) string {
	t.Helper()
	directory := t.TempDir()
	entries, err := os.ReadDir(filepath.Join(repository, "internal/native/runtime"))
	if err != nil {
		t.Fatal(err)
	}
	changed := false
	for _, entry := range entries {
		if entry.IsDir() || (!strings.HasSuffix(entry.Name(), ".c") && !strings.HasSuffix(entry.Name(), ".h")) {
			continue
		}
		data, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime", entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if entry.Name() == file {
			contents := string(data)
			start := 0
			if scope != "" {
				start = strings.Index(contents, scope)
				if start < 0 {
					t.Fatal("missing mutant scope", scope)
				}
			}
			at := strings.Index(contents[start:], before)
			if at < 0 {
				t.Fatal("missing mutant anchor", before)
			}
			at += start
			data = []byte(contents[:at] + after + contents[at+len(before):])
			changed = true
		}
		if err := os.WriteFile(filepath.Join(directory, entry.Name()), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	if !changed {
		t.Fatal("mutant changed nothing")
	}
	options := native.Options{Sanitize: true}
	library, err := native.RuntimeLibraryForSource(directory, source, options)
	if err != nil {
		t.Fatal(err)
	}
	main := filepath.Join(directory, "program.c")
	binary := filepath.Join(directory, "mutant")
	if err := os.WriteFile(main, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	args := append(native.Flags(options), "-I", filepath.Dir(library), "-o", binary, main)
	args = append(args, native.RuntimeLinkFlags(library)...)
	args = append(args, "-lm")
	if data, err := exec.Command("clang", args...).CombinedOutput(); err != nil {
		t.Fatalf("mutant must compile: %v\n%s", err, data)
	}
	return binary
}

func TestNodeHostCoverageMutants(t *testing.T) {
	t.Parallel()
	for _, one := range []struct{ name, fixture, file, scope, before, after string }{
		{"number-grouping", "fs_numeric_error", "node_fs_file.c", "static adamic_string *range_number", "fabs(value) <= 4294967296.0", "true"},
		{"string-surrogate", "fs_error_preview", "node_fs_file.c", "static size_t inspect_string_part", `"\\u%04x"`, `"\\u%04X"`},
		{"c1-controls", "fs_error_preview", "node_fs_file.c", "static size_t inspect_string_part", "(c >= 127 && c < 160)", "false"},
		{"vertical-tab", "fs_error_preview", "node_fs_file.c", "static size_t inspect_string_part", `for (size_t j = 0; j < 4; j++) { shown[at++] = (unsigned char)escape[j]; }`, `if (c == 11) { shown[at++] = '\\'; shown[at++] = 'v'; } else { for (size_t j = 0; j < 4; j++) { shown[at++] = (unsigned char)escape[j]; } }`},
		{"zero-read", "fs_zero_read", "node_fs_file.c", "adamic_fs_file_read_sync", "if (length == 0) { return 0; }", ""},
		{"owned-close", "fs_fd_ownership", "node_fs_file.c", "static adamic_array *read_buffer", "if (owned) { close(descriptor); }", "(void)owned;"},
		{"cwd-cache", "host_directory", "node_process.c", "adamic_node_chdir", "if (current_directory != NULL) { adamic_release(current_directory); current_directory = NULL; }", ""},
		{"short-write", "fs_large_roundtrip", "node_fs_file.c", "static double write_data", "used += (size_t)count;", "used += (size_t)count; break;"},
	} {
		t.Run(one.name, func(t *testing.T) {
			t.Parallel()
			env := []string{"UBSAN_OPTIONS=halt_on_error=1"}
			if runtime.GOOS == "linux" {
				env = append(env, "ASAN_OPTIONS=detect_leaks=1:halt_on_error=1")
			}
			if one.name == "short-write" {
				env = append(env, "LD_PRELOAD="+nodeHostShortIO(t))
			}
			path, err := filepath.Abs(filepath.Join(repository, nodeHostCoveragePath(one.fixture)))
			if err != nil {
				t.Fatal(err)
			}
			p, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			binary := nodeHostRuntimeMutant(t, native.C(p), one.file, one.scope, one.before, one.after)
			got := executeWith(t, env, binary)
			if truth.exitCode != 0 || len(truth.stderr) != 0 || got.exitCode != 0 || len(got.stderr) != 0 || disagreement(truth, got) != "stdout differs" {
				t.Fatalf("mutant must be caught only by Node stdout: Node %d %q %q; mutant %d %q %q", truth.exitCode, truth.stdout, truth.stderr, got.exitCode, got.stdout, got.stderr)
			}
			t.Log("caught only by Node stdout; exit zero, empty stderr, clean sanitizers and LeakSanitizer")
		})
	}
}
