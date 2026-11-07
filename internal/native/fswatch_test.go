package native_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/compilerinputs"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

// fs.watch's programs are held to Node in internal/oracle (library_node_watch_test.go). This harness
// pins what no program reaches: a rename, against Node running the same operations, and the error
// path, which Node takes only when libuv fails reading its descriptor and so has no Node answer: a
// pipe whose writer is gone stands in for the inotify descriptor, and the next wait sees POLLHUP.
func fsWatchHarness(t *testing.T) string {
	t.Helper()
	source, err := os.ReadFile("testdata/fswatch/harness.c")
	if err != nil {
		t.Fatal(err)
	}
	return string(source)
}

const (
	fsWatchErrorLine    = "Error: watch: the inotify descriptor hung up or became invalid\n"
	fsWatchUncaught     = "adamic: panic: Error: watch: the inotify descriptor hung up or became invalid\n"
	fsWatchCustomHooks  = "adamic: panic: NotYet: readiness sources with custom host loop hooks\n"
	fsWatchHarnessLimit = 3 * time.Second
	// shared.c returns this after verifying EMFILE at watch and releasing its values.
	fsWatchCapFailureStatus     = 17
	fsWatchRefnessFailureStatus = 18
)

var fsWatchSanitizers = []string{"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}

// fsWatchVariants are the builds a finishing mode runs in: release, then under the sanitizers with
// its value counts balanced.
var fsWatchVariants = []struct {
	name    string
	options native.Options
	env     []string
}{{"release", native.Options{Release: true}, nil}, {"sanitized", native.Options{Sanitize: true, Count: true}, fsWatchSanitizers}}

// fsWatchRun runs mode in a fresh TMPDIR, the harness's own deadline its failure.
func fsWatchRun(t *testing.T, environment []string, binary, mode string) (leakcheck.Run, bool) {
	t.Helper()
	return timerRunWithin(fsWatchHarnessLimit, append([]string{"TMPDIR=" + t.TempDir()}, environment...), binary, mode)
}

// The rename row of Node's measured table: a file renamed in a watched directory is a rename of
// its old name, then of its new one.
func TestFSWatchRenameMatchesNode(t *testing.T) {
	t.Parallel()
	node, err := exec.LookPath("node")
	if err != nil {
		t.Fatal(err)
	}
	script, err := filepath.Abs("testdata/fswatch/oracle.a")
	if err != nil {
		t.Fatal(err)
	}
	truth, expired := timerRunWithin(fsWatchHarnessLimit, []string{"TMPDIR=" + t.TempDir()}, node, "--input-type=commonjs", "-e", `require('node:vm').runInThisContext(require('node:fs').readFileSync(process.argv[1],'utf8'))`, script)
	if expired || truth.ExitCode != 0 || string(truth.Stdout) != "rename b.txt\nrename c.txt\nrename end\ndone\n" {
		t.Fatalf("Node: %+v", truth)
	}
	for _, variant := range fsWatchVariants {
		got, expired := fsWatchRun(t, variant.env, fsWatchProduct(t, variant.name, variant.options), "rename")
		if expired || got.ExitCode != 0 || string(got.Stdout) != string(truth.Stdout) {
			t.Fatalf("%s: %+v", variant.name, got)
		}
		if variant.name == "sanitized" {
			if report := leakcheck.Unbalanced(got); report != "" {
				t.Fatal(report)
			}
		}
	}
}

// The error bit calls the 'error' listeners with a real Error, after closing the watcher, so the
// loop ends rather than dispatching a hung-up descriptor every turn (Runtime's should on #j71fn59).
func TestFSWatchErrorClosesTheWatcher(t *testing.T) {
	t.Parallel()
	for _, variant := range fsWatchVariants {
		got, expired := fsWatchRun(t, variant.env, fsWatchProduct(t, variant.name, variant.options), "error")
		if expired || got.ExitCode != 0 || string(got.Stdout) != fsWatchErrorLine+"done\n" {
			t.Fatalf("%s: expired %t, %+v", variant.name, expired, got)
		}
		if variant.name == "sanitized" {
			if report := leakcheck.Unbalanced(got); report != "" {
				t.Fatal(report)
			}
		}
	}
}

// An 'error' nobody listens for is thrown, as EventEmitter's emit throws it, and nothing catches it.
func TestFSWatchErrorWithoutListenerIsUncaught(t *testing.T) {
	t.Parallel()
	got, expired := fsWatchRun(t, nil, fsWatchProduct(t, "release", native.Options{Release: true}), "error-uncaught")
	if expired || got.ExitCode != 70 || len(got.Stdout) != 0 || string(got.Stderr) != fsWatchUncaught {
		t.Fatalf("uncaught error: expired %t, %+v", expired, got)
	}
}

// Under custom host loop hooks the readiness source's own refusal applies.
func TestFSWatchCustomHooksRefused(t *testing.T) {
	t.Parallel()
	got, _ := fsWatchRun(t, nil, fsWatchProduct(t, "release", native.Options{Release: true}), "custom-hooks")
	if got.ExitCode != 70 || string(got.Stderr) != fsWatchCustomHooks {
		t.Fatalf("custom hooks must refuse in the source's words: %+v", got)
	}
}

// The error path that calls its listeners but leaves the watcher open: the error bit comes back every
// turn, and the loop spins past the pin's deadline.
func TestFSWatchMutantErrorKeepsSource(t *testing.T) {
	t.Parallel()
	got, expired := fsWatchRun(t, nil, fsWatchProduct(t, "mutant-error-keeps-source", native.Options{Release: true}), "error")
	if !expired {
		t.Fatalf("the mutant must spin past the deadline: %+v", got)
	}
	if !strings.HasPrefix(string(got.Stdout), fsWatchErrorLine+fsWatchErrorLine) {
		t.Fatalf("the mutant must call the listener every turn: %q", got.Stdout)
	}
	t.Log("TestFSWatchErrorClosesTheWatcher goes red with the error path leaving its watcher open")
}

// fs.watch('adamic-fswatch-missing', {persistent: true}, ...) on Node v24.14.1 on Linux throws an Error
// whose own keys after message are errno, syscall, code, path and filename, in that order, with these
// values. The first line is the shape's keys in its order, name and message first as every native
// error's; the rest are each key read by name.
const fsWatchAddErrorKeys = "keys name message errno syscall code path filename\n" +
	"name Error\n" +
	"message ENOENT: no such file or directory, watch 'adamic-fswatch-missing'\n" +
	"errno -2\n" +
	"syscall watch\n" +
	"code ENOENT\n" +
	"path adamic-fswatch-missing\n" +
	"filename adamic-fswatch-missing\n"

// A watch that can't start throws Node's keys in Node's order, each read by its name.
func TestFSWatchAddErrorCarriesNodeKeys(t *testing.T) {
	t.Parallel()
	for _, variant := range fsWatchVariants {
		got, expired := fsWatchRun(t, variant.env, fsWatchProduct(t, variant.name, variant.options), "add-error")
		if expired || got.ExitCode != 0 || string(got.Stdout) != fsWatchAddErrorKeys {
			t.Fatalf("%s: expired %t, %+v", variant.name, expired, got)
		}
		if variant.name == "sanitized" {
			if report := leakcheck.Unbalanced(got); report != "" {
				t.Fatal(report)
			}
		}
	}
}

// The shape with syscall dropped: the keys line loses it, and its read by name stops the run as a
// field the checker proved is there and isn't.
func TestFSWatchMutantAddErrorDropsSyscall(t *testing.T) {
	t.Parallel()
	got, expired := fsWatchRun(t, nil, fsWatchProduct(t, "mutant-add-error-drops-syscall", native.Options{Release: true}), "add-error")
	const stdout = "keys name message errno code path filename\n" +
		"name Error\n" +
		"message ENOENT: no such file or directory, watch 'adamic-fswatch-missing'\n" +
		"errno -2\n"
	const stderr = "adamic: panic: compiler bug: a field the checker proved is there is missing\n"
	if expired || got.ExitCode != 70 || string(got.Stdout) != stdout || string(got.Stderr) != stderr {
		t.Fatalf("the mutant's error must lack syscall: expired %t, %+v", expired, got)
	}
	t.Log("TestFSWatchAddErrorCarriesNodeKeys goes red with syscall dropped from the add-time error's shape")
}

const (
	fsWatchErrorClose  = "    adamic_object *object = adamic_retain(watcher->object);\n    fs_watcher_close(watcher);\n"
	fsWatchErrorMutant = "    adamic_object *object = adamic_retain(watcher->object);\n    /* mutant: the watcher stays open */\n"
	// watch_error_shape's four tables, then the same with syscall's name, reference and kind dropped and
	// the count one less, so the mutant compiles.
	fsWatchAddErrorShape = "    static const char *const watch_error_names[] = {\"name\", \"message\", \"errno\", \"syscall\", \"code\", \"path\", \"filename\"};\n" +
		"    static const bool watch_error_references[] = {true, true, false, true, true, true, true};\n" +
		"    static const adamic_field_kind watch_error_kinds[] = {adamic_field_reference, adamic_field_reference, adamic_field_number, adamic_field_reference, adamic_field_reference, adamic_field_reference, adamic_field_reference};\n" +
		"    static const adamic_shape watch_error_shape = {7, watch_error_names, watch_error_references, &adamic_error_methods, watch_error_kinds};\n"
	fsWatchAddErrorMutant = "    static const char *const watch_error_names[] = {\"name\", \"message\", \"errno\", \"code\", \"path\", \"filename\"};\n" +
		"    static const bool watch_error_references[] = {true, true, false, true, true, true};\n" +
		"    static const adamic_field_kind watch_error_kinds[] = {adamic_field_reference, adamic_field_reference, adamic_field_number, adamic_field_reference, adamic_field_reference, adamic_field_reference};\n" +
		"    static const adamic_shape watch_error_shape = {6, watch_error_names, watch_error_references, &adamic_error_methods, watch_error_kinds};\n"
)

// Products hold compiled probes only; every run and comparison happens in the tests above. A
// mutant's runtime is a private copy, linked as the timers' mutants are (timerMutant).
func fsWatchProduct(t *testing.T, variant string, options native.Options) string {
	t.Helper()
	flags := append([]string{fmt.Sprintf("options=%+v", options)}, native.Flags(options)...)
	for _, variable := range []string{"ADAMIC_PROGRAM_REGION", "ADAMIC_NATIVE_SPLIT", "ADAMIC_NATIVE_JOBS", "CPATH", "C_INCLUDE_PATH", "SDKROOT", "MACOSX_DEPLOYMENT_TARGET"} {
		value, set := os.LookupEnv(variable)
		flags = append(flags, fmt.Sprintf("env:%s:set=%t:value=%s", variable, set, value))
	}
	inputs := buildcache.Inputs{Name: "native-fs-watch-" + variant,
		Files:     append(compilerinputs.Lowering(), "internal/native/fswatch_test.go", "internal/native/testdata/fswatch/harness.c", "internal/native/testdata/fswatch/shared.c", "internal/native/testdata/fswatch/shared-node.cjs", "internal/native/testdata/fswatch/head2-mutant.h"),
		Flags:     flags,
		Toolchain: []string{goruntime.Version(), goruntime.GOOS + "/" + goruntime.GOARCH, buildcache.Tool("clang", "--version"), buildcache.Tool("ar", "--version")}}
	directory := buildcache.Product(t, inputs, func(directory string) error {
		code := fsWatchHarness(t)
		output := filepath.Join(directory, "probe")
		switch variant {
		case "release", "sanitized":
			return native.Build(code, output, options)
		case "mutant-error-keeps-source":
			changed := timerMutant(t, "node_fs_watch_impl.h", fsWatchErrorClose, fsWatchErrorMutant)
			header := filepath.Join(changed.directory, "node_fs_watch_impl.h")
			data, err := os.ReadFile(header)
			if err != nil {
				return err
			}
			data = []byte(strings.Replace(string(data), "    fs_watch_stop();\n    size_t count = 0;", "    /* mutant: keep the failed descriptor */\n    size_t count = 0;", 1))
			if err := os.WriteFile(header, data, 0600); err != nil {
				return err
			}
			library, err := native.RuntimeLibraryForSource(changed.directory, code, options)
			if err != nil {
				return err
			}
			return native.BuildWithRuntime(code, output, options, library)
		}
		if strings.HasPrefix(variant, "shared") {
			data, err := os.ReadFile("testdata/fswatch/shared.c")
			if err != nil {
				return err
			}
			code = string(data)
			changed := timerMutant(t, "adamic.h", "#define ADAMIC_H\n", "#define ADAMIC_H\n#define ADAMIC_FS_WATCH_TEST 1\n")
			header := filepath.Join(changed.directory, "node_fs_watch_impl.h")
			data, err = os.ReadFile(header)
			if err != nil {
				return err
			}
			switch variant {
			case "shared-mutant-instances":
				data, err = os.ReadFile("testdata/fswatch/head2-mutant.h")
				if err != nil {
					return err
				}
				code += "\nvoid adamic_fs_watch_test_fail_read(void) {}\nuintptr_t adamic_fs_watch_test_source_identity(void) { return 0; }\n"
			case "shared-mutant-refness":
				before := "    } else if (referenced) {\n        (void)adamic_source_ref(fs_watch_source);\n    } else {\n        (void)adamic_source_unref(fs_watch_source);\n    }"
				after := "    } else {\n        (void)adamic_source_close(fs_watch_source);\n        fs_watch_source = adamic_source_register(fs_watch_descriptor, fs_watcher_dispatch,\n            (void *)fs_watch_generation, referenced);\n    }"
				if strings.Count(string(data), before) != 1 {
					return fmt.Errorf("refness mutant needs exactly one flip body")
				}
				data = []byte(strings.Replace(string(data), before, after, 1))
			case "shared-mutant-remove":
				data = []byte(strings.Replace(string(data), "entry != NULL && entry->watchers == NULL", "entry != NULL", 1))
			case "shared-mutant-identity":
				text := string(data)
				text = strings.Replace(text, "uintptr_t *batch = malloc((count == 0 ? 1 : count) * sizeof *batch);", "uintptr_t *batch = malloc((count == 0 ? 1 : count) * sizeof *batch);\n    fs_watcher **stale = malloc((count == 0 ? 1 : count) * sizeof *stale);\n    if (stale == NULL) abort();", 1)
				text = strings.Replace(text, "batch[--index] = watcher->identity;", "batch[--index] = watcher->identity; stale[index] = watcher;", 1)
				text = strings.Replace(text, "fs_watcher *watcher = fs_watcher_open(batch[index]);", "fs_watcher *watcher = stale[index];", 1)
				text = strings.Replace(text, "    free(batch);\n}", "    free(stale); free(batch);\n}", 1)
				data = []byte(text)
			}
			if err := os.WriteFile(header, data, 0600); err != nil {
				return err
			}
			library, err := native.RuntimeLibraryForSource(changed.directory, code, options)
			if err != nil {
				return err
			}
			return native.BuildWithRuntime(code, output, options, library)
		case "mutant-add-error-drops-syscall":
			changed := timerMutant(t, "node_fs_watch_impl.h", fsWatchAddErrorShape, fsWatchAddErrorMutant)
			library, err := native.RuntimeLibraryForSource(changed.directory, code, options)
			if err != nil {
				return err
			}
			return native.BuildWithRuntime(code, output, options, library)
		}
		return fmt.Errorf("unknown fs.watch product %s", variant)
	})
	buildcache.RequireArtifacts(t, directory, buildcache.Artifact{Name: "probe"})
	return filepath.Join(directory, "probe")
}

func TestProduct_FSWatchHarnessRelease(t *testing.T) {
	t.Parallel()
	fsWatchProduct(t, "release", native.Options{Release: true})
}
func TestProduct_FSWatchHarnessSanitized(t *testing.T) {
	t.Parallel()
	fsWatchProduct(t, "sanitized", native.Options{Sanitize: true, Count: true})
}
func TestProduct_FSWatchHarnessMutantErrorKeepsSource(t *testing.T) {
	t.Parallel()
	fsWatchProduct(t, "mutant-error-keeps-source", native.Options{Release: true})
}
func TestProduct_FSWatchHarnessMutantAddErrorDropsSyscall(t *testing.T) {
	t.Parallel()
	fsWatchProduct(t, "mutant-add-error-drops-syscall", native.Options{Release: true})
}

// The tree builds each product above ahead by its row in cloud/fast-gate/product-order.json, which
// fswatch-product-order.py writes. A product added without rerunning it reds this test.
func TestFSWatchHarnessProductOrderIsCurrent(t *testing.T) {
	t.Parallel()
	script, err := filepath.Abs("../../cloud/fast-gate/fswatch-product-order.py")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	output, err := exec.CommandContext(ctx, "python3", script, "--check").CombinedOutput()
	if err != nil {
		t.Fatalf("fs.watch product generator: %v\n%s", err, output)
	}
}

// The child starts below a 64-descriptor limit; no machine-wide inotify settings change.
func fsWatchSharedNode(t *testing.T, mode string) leakcheck.Run {
	t.Helper()
	script, err := filepath.Abs("testdata/fswatch/shared-node.cjs")
	if err != nil {
		t.Fatal(err)
	}
	node, err := exec.LookPath("node")
	if err != nil {
		t.Fatal(err)
	}
	arguments := []string{node, script, mode}
	if mode == "cap" || mode == "cap-failure" {
		arguments = append([]string{"prlimit", "--nofile=64:64", "--"}, arguments...)
	}
	got, expired := timerRunWithin(fsWatchHarnessLimit, []string{"TMPDIR=" + t.TempDir()}, arguments[0], arguments[1:]...)
	if expired || got.ExitCode != 0 || len(got.Stderr) != 0 {
		t.Fatalf("Node %s: expired=%t %+v", mode, expired, got)
	}
	return got
}
func fsWatchSharedAgreement(t *testing.T, mode string) {
	t.Helper()
	truth := fsWatchSharedNode(t, mode)
	got, expired := fsWatchRun(t, fsWatchSanitizers, fsWatchProduct(t, "shared", native.Options{Sanitize: true, Count: true}), mode)
	if expired || got.ExitCode != truth.ExitCode || string(got.Stdout) != string(truth.Stdout) {
		t.Fatalf("%s: expired=%t native=%+v Node=%+v", mode, expired, got, truth)
	}
	if report := leakcheck.Unbalanced(got); report != "" {
		t.Fatal(report)
	}
}
func TestFSWatchSharedDescriptorCapMatchesNode(t *testing.T) {
	t.Parallel()
	fsWatchSharedAgreement(t, "cap")
}
func TestFSWatchSamePathSurvivesCloseMatchesNode(t *testing.T) {
	t.Parallel()
	fsWatchSharedAgreement(t, "same")
}
func TestFSWatchSharedSourceUnrefMatchesNode(t *testing.T) {
	t.Parallel()
	fsWatchSharedAgreement(t, "unref")
}
func TestFSWatchFailureListenerReplacesWatcher(t *testing.T) {
	t.Parallel()
	got, expired := fsWatchRun(t, fsWatchSanitizers, fsWatchProduct(t, "shared", native.Options{Sanitize: true, Count: true}), "failure")
	if expired || got.ExitCode != 0 || string(got.Stdout) != "first error\nthird change\n" {
		t.Fatalf("failure listener: expired=%t %+v", expired, got)
	}
	if report := leakcheck.Unbalanced(got); report != "" {
		t.Fatal(report)
	}
}
func TestFSWatchMutantPerWatcherDescriptorHitsCap(t *testing.T) {
	t.Parallel()
	truth := fsWatchSharedNode(t, "cap-failure")
	if string(truth.Stdout) != "watch failure code=EMFILE message=EMFILE: too many open files, watch '<fixture-path>'\n" {
		t.Fatalf("Node watch failure: %+v", truth)
	}
	got, expired := fsWatchRun(t, fsWatchSanitizers, fsWatchProduct(t, "shared-mutant-instances", native.Options{Sanitize: true, Count: true}), "cap")
	if expired || got.ExitCode != fsWatchCapFailureStatus || string(got.Stdout) != string(truth.Stdout) {
		t.Fatalf("per-watcher cap mutant: expired=%t %+v", expired, got)
	}
	if string(leakcheck.CountsLine.Find(got.Stderr)) != string(got.Stderr) {
		t.Fatalf("cap fixture must report only its final counts: %q", got.Stderr)
	}
	// The exact chosen status and EMFILE line were checked above. Unbalanced
	// expects a zero status; normalize only its copy so it still checks every count.
	counted := got
	counted.ExitCode = 0
	if report := leakcheck.Unbalanced(counted); report != "" {
		t.Fatal(report)
	}
	t.Logf("TestFSWatchSharedDescriptorCapMatchesNode goes red with head 2's per-watcher descriptors: %s", strings.TrimSpace(string(got.Stdout)))
}
func TestFSWatchMutantRemovesSharedWatchTooSoon(t *testing.T) {
	t.Parallel()
	got, expired := fsWatchRun(t, fsWatchSanitizers, fsWatchProduct(t, "shared-mutant-remove", native.Options{Sanitize: true, Count: true}), "same")
	if !expired {
		t.Fatalf("shared remove mutant must miss its own deadline: %+v", got)
	}
	t.Logf("TestFSWatchSamePathSurvivesCloseMatchesNode goes red when every close removes the kernel watch: deadline exceeded; stdout=%q", got.Stdout)
}
func TestFSWatchMutantFailureUsesFreedRecord(t *testing.T) {
	t.Parallel()
	got, expired := fsWatchRun(t, fsWatchSanitizers, fsWatchProduct(t, "shared-mutant-identity", native.Options{Sanitize: true, Count: true}), "failure")
	if expired || !strings.Contains(string(got.Stderr), "AddressSanitizer: heap-use-after-free") {
		t.Fatalf("failure identity mutant: expired=%t %+v", expired, got)
	}
	t.Log("TestFSWatchFailureListenerReplacesWatcher goes red without identity re-lookup: AddressSanitizer: heap-use-after-free")
}
func TestProduct_FSWatchSharedHarness(t *testing.T) {
	t.Parallel()
	fsWatchProduct(t, "shared", native.Options{Sanitize: true, Count: true})
}
func TestProduct_FSWatchSharedMutantInstances(t *testing.T) {
	t.Parallel()
	fsWatchProduct(t, "shared-mutant-instances", native.Options{Sanitize: true, Count: true})
}
func TestProduct_FSWatchSharedMutantRemove(t *testing.T) {
	t.Parallel()
	fsWatchProduct(t, "shared-mutant-remove", native.Options{Sanitize: true, Count: true})
}
func TestProduct_FSWatchSharedMutantIdentity(t *testing.T) {
	t.Parallel()
	fsWatchProduct(t, "shared-mutant-identity", native.Options{Sanitize: true, Count: true})
}

// This pin observes a refness flip before any logical watcher or source closes.
func TestFSWatchRefnessFlipKeepsSourceAndNoPendingClose(t *testing.T) {
	t.Parallel()
	got, expired := fsWatchRun(t, fsWatchSanitizers, fsWatchProduct(t, "shared", native.Options{Sanitize: true, Count: true}), "refness")
	if expired || got.ExitCode != 0 || string(got.Stdout) != "refness source unchanged pending=0\n" {
		t.Fatalf("refness flip: expired=%t %+v", expired, got)
	}
	if report := leakcheck.Unbalanced(got); report != "" {
		t.Fatal(report)
	}
}
func TestFSWatchMutantRefnessClosesAndReregisters(t *testing.T) {
	t.Parallel()
	got, expired := fsWatchRun(t, fsWatchSanitizers, fsWatchProduct(t, "shared-mutant-refness", native.Options{Sanitize: true, Count: true}), "refness")
	if expired || got.ExitCode != fsWatchRefnessFailureStatus || string(got.Stdout) != "refness source changed pending=1\n" {
		t.Fatalf("refness mutant: expired=%t %+v", expired, got)
	}
	if string(leakcheck.CountsLine.Find(got.Stderr)) != string(got.Stderr) {
		t.Fatalf("refness mutant must report only final counts: %q", got.Stderr)
	}
	counted := got
	counted.ExitCode = 0
	if report := leakcheck.Unbalanced(counted); report != "" {
		t.Fatal(report)
	}
	t.Log("TestFSWatchRefnessFlipKeepsSourceAndNoPendingClose goes red with close/re-register: refness source changed pending=1")
}
func TestProduct_FSWatchSharedMutantRefness(t *testing.T) {
	t.Parallel()
	fsWatchProduct(t, "shared-mutant-refness", native.Options{Sanitize: true, Count: true})
}
