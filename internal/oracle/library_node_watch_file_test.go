package oracle

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/compilerinputs"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/toolchain"
)

var watchFileFixtures = []string{"missing", "appearance", "mtime", "cancel_one", "cancel_all", "duplicate", "batch", "unref", "listener_throw", "tsc_close"}

func init() {
	for _, name := range watchFileFixtures {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/host_watch_file/" + name + ".a", true, false})
	}
}

const watchFileDeadline = 5 * time.Second

// Both the oracle and both backends get a fresh cwd and a deliberately explicit
// environment, never the box's environment. The same fixtures also run in the
// ordinary native, JS, gob and counted families. WASI refuses their host entry.
func watchFileRun(t *testing.T, name string, args ...string) (run, bool) {
	t.Helper()
	path, err := toolchain.Path(name)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), watchFileDeadline)
	defer cancel()
	command := exec.CommandContext(ctx, path, args...)
	command.Dir = t.TempDir()
	command.Env = []string{"TMPDIR=" + command.Dir, "ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error { return syscall.Kill(-command.Process.Pid, syscall.SIGKILL) }
	stdout, stderr, err := childguard.Capture(ctx, command)
	expired := ctx.Err() == context.DeadlineExceeded
	var exited *exec.ExitError
	if err != nil && !expired && !errors.As(err, &exited) {
		t.Fatal(err)
	}
	if command.ProcessState == nil {
		t.Fatal("watch fixture did not start")
	}
	return run{stdout: stdout, stderr: stderr, exitCode: command.ProcessState.ExitCode()}, expired
}
func watchFilePath(t *testing.T, name string) string {
	t.Helper()
	path, err := filepath.Abs("testdata/host_watch_file/" + name + ".a")
	if err != nil {
		t.Fatal(err)
	}
	return path
}
func watchFileNode(t *testing.T, path string) run {
	t.Helper()
	loader, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	result, expired := watchFileRun(t, "node", "--disable-warning=ExperimentalWarning", loader, path)
	if expired {
		t.Fatal("Node watch fixture exceeded its own deadline")
	}
	return result
}

type watchFileMutation struct {
	name, fixture, file, before, after, catcher string
	deadline                                    bool
}

var watchFileMutations = []watchFileMutation{
	{"size-only", "mtime", "node_watch_file_impl.h", "if (old->value.st_size != current->value.st_size) return true;", "return old->value.st_size != current->value.st_size;", "TestWatchFileSameSizeMtimeMatchesNode", true},
	{"initial-missing-suppressed", "missing", "node_watch_file_impl.h", "status != 0 ? poller->status != status :", "status != 0 ? poller->status != 0 && poller->status != status :", "TestWatchFileInitialMissingMatchesNode", true},
	{"unref-ignored", "unref", "node_watch_file_impl.h", "if (!initial && !poller->persistent)", "if (false)", "TestWatchFileUnrefInitialStatMatchesNode", true},
	{"listener-identity-ignored", "cancel_one", "node_watch_file_impl.h", "if ((*at)->callback == callback) match = at;", "if ((*at)->callback != callback) match = at;", "TestWatchFileCancelOneMatchesNode", false},
	{"throw-ignored", "listener_throw", "timers_impl.h", "if (adamic_thrown != NULL) adamic_uncaught();", "/* mutant: ignore the thrown listener */", "TestWatchFileThrowingListenerMatchesNode", false},
}

func watchFileProduct(t *testing.T, fixture, variant string) string {
	t.Helper()
	flags := append(native.Flags(native.Options{Sanitize: true}), "explicit-child-environment", "watch-file-v1")
	for _, name := range []string{"ADAMIC_PROGRAM_REGION", "ADAMIC_NATIVE_SPLIT", "ADAMIC_NATIVE_JOBS", "CPATH", "C_INCLUDE_PATH", "SDKROOT", "MACOSX_DEPLOYMENT_TARGET"} {
		value, set := os.LookupEnv(name)
		flags = append(flags, fmt.Sprintf("env:%s:set=%t:value=%s", name, set, value))
	}
	inputs := buildcache.Inputs{Name: "oracle-watch-file-" + fixture + "-" + variant, Files: append(compilerinputs.Lowering(), "internal/oracle/library_node_watch_file_test.go", "internal/oracle/testdata/host_watch_file/"+fixture+".a"),
		Flags:     flags,
		Toolchain: []string{runtime.Version(), runtime.GOOS + "/" + runtime.GOARCH, buildcache.Tool("clang", "--version"), buildcache.Tool("ar", "--version")}}
	return buildcache.Product(t, inputs, func(directory string) error {
		program, err := lowered(t, watchFilePath(t, fixture))
		if err != nil {
			return err
		}
		source := native.C(program)
		if err := os.WriteFile(filepath.Join(directory, "program.mjs"), []byte(javascript.JavaScript(program)), 0600); err != nil {
			return err
		}
		options := native.Options{Sanitize: true}
		if variant == "control" {
			return native.Build(source, filepath.Join(directory, "native"), options)
		}
		var mutation *watchFileMutation
		for i := range watchFileMutations {
			if watchFileMutations[i].name == variant {
				mutation = &watchFileMutations[i]
			}
		}
		if mutation == nil {
			return fmt.Errorf("unknown watchFile variant %s", variant)
		}
		runtimeDirectory := filepath.Join(directory, "runtime-mutant")
		if err := os.Mkdir(runtimeDirectory, 0700); err != nil {
			return err
		}
		entries, err := os.ReadDir("../native/runtime")
		if err != nil {
			return err
		}
		matched := false
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			data, err := os.ReadFile(filepath.Join("../native/runtime", entry.Name()))
			if err != nil {
				return err
			}
			if entry.Name() == mutation.file {
				if strings.Count(string(data), mutation.before) != 1 {
					return fmt.Errorf("mutant %s must match once", variant)
				}
				data = []byte(strings.Replace(string(data), mutation.before, mutation.after, 1))
				matched = true
			}
			if err := os.WriteFile(filepath.Join(runtimeDirectory, entry.Name()), data, 0600); err != nil {
				return err
			}
		}
		if !matched {
			return fmt.Errorf("mutant file not found")
		}
		library, err := native.RuntimeLibraryForSource(runtimeDirectory, source, options)
		if err != nil {
			return err
		}
		return native.BuildWithRuntime(source, filepath.Join(directory, "native"), options, library)
	})
}
func watchFileAgreement(t *testing.T, name string) {
	t.Helper()
	product := watchFileProduct(t, name, "control")
	truth := watchFileNode(t, watchFilePath(t, name))
	wantExit := 0
	if name == "listener_throw" {
		wantExit = 70
	}
	if truth.exitCode != wantExit {
		t.Fatalf("Node watch fixture: %+v", truth)
	}
	for _, backend := range []string{"native", "javascript"} {
		var got run
		var expired bool
		if backend == "native" {
			got, expired = watchFileRun(t, filepath.Join(product, "native"))
		} else {
			got = watchFileNode(t, filepath.Join(product, "program.mjs"))
		}
		if expired {
			t.Fatal("watchFile did not finish by its own 5 s deadline")
		}
		if difference := disagreement(truth, got); difference != "" {
			t.Fatalf("%s: %s; stdout=%q stderr=%q exit=%d, Node stdout=%q stderr=%q exit=%d", backend, difference, got.stdout, got.stderr, got.exitCode, truth.stdout, truth.stderr, truth.exitCode)
		}
	}
}
func TestWatchFileInitialMissingMatchesNode(t *testing.T) {
	t.Parallel()
	watchFileAgreement(t, "missing")
}
func TestWatchFileAppearanceAndDeletionMatchNode(t *testing.T) {
	t.Parallel()
	watchFileAgreement(t, "appearance")
}
func TestWatchFileSameSizeMtimeMatchesNode(t *testing.T) {
	t.Parallel()
	watchFileAgreement(t, "mtime")
}
func TestWatchFileCancelOneMatchesNode(t *testing.T) {
	t.Parallel()
	watchFileAgreement(t, "cancel_one")
}
func TestWatchFileCancelAllMatchesNode(t *testing.T) {
	t.Parallel()
	watchFileAgreement(t, "cancel_all")
}
func TestWatchFileDuplicateListenerMatchesNode(t *testing.T) {
	t.Parallel()
	watchFileAgreement(t, "duplicate")
}
func TestWatchFileListenerBatchMatchesNode(t *testing.T) {
	t.Parallel()
	watchFileAgreement(t, "batch")
}
func TestWatchFileUnrefInitialStatMatchesNode(t *testing.T) {
	t.Parallel()
	watchFileAgreement(t, "unref")
}
func TestWatchFileThrowingListenerMatchesNode(t *testing.T) {
	t.Parallel()
	watchFileAgreement(t, "listener_throw")
}

func watchFileMutant(t *testing.T, name string) {
	t.Helper()
	var mutation *watchFileMutation
	for i := range watchFileMutations {
		if watchFileMutations[i].name == name {
			mutation = &watchFileMutations[i]
		}
	}
	if mutation == nil {
		t.Fatal(name)
	}
	binary := filepath.Join(watchFileProduct(t, mutation.fixture, name), "native")
	got, expired := watchFileRun(t, binary)
	truth := watchFileNode(t, watchFilePath(t, mutation.fixture))
	if mutation.deadline {
		if !expired {
			t.Fatalf("%s must miss its own deadline: %+v", name, got)
		}
	} else {
		if expired {
			t.Fatal("mutant must complete, not time out")
		}
		if bytes.Contains(got.stderr, []byte("Sanitizer")) || bytes.Contains(got.stderr, []byte("runtime error:")) {
			t.Fatalf("mutant must be sanitizer clean: %+v", got)
		}
		if difference := disagreement(truth, got); difference == "" {
			t.Fatalf("%s survived", name)
		}
	}
	t.Logf("%s goes red with %s", mutation.catcher, name)
}
func TestWatchFileMutantIgnoresSameSizeMtime(t *testing.T) {
	t.Parallel()
	watchFileMutant(t, "size-only")
}
func TestWatchFileMutantSuppressesInitialMissing(t *testing.T) {
	t.Parallel()
	watchFileMutant(t, "initial-missing-suppressed")
}
func TestWatchFileMutantKeepsUnrefAlive(t *testing.T) {
	t.Parallel()
	watchFileMutant(t, "unref-ignored")
}
func TestWatchFileMutantIgnoresListenerIdentity(t *testing.T) {
	t.Parallel()
	watchFileMutant(t, "listener-identity-ignored")
}
func TestWatchFileMutantIgnoresThrownListener(t *testing.T) {
	t.Parallel()
	watchFileMutant(t, "throw-ignored")
}

func TestProduct_WatchFileInitialMissing(t *testing.T) {
	t.Parallel()
	watchFileProduct(t, "missing", "control")
}
func TestProduct_WatchFileAppearance(t *testing.T) {
	t.Parallel()
	watchFileProduct(t, "appearance", "control")
}
func TestProduct_WatchFileSameSizeMtime(t *testing.T) {
	t.Parallel()
	watchFileProduct(t, "mtime", "control")
}
func TestProduct_WatchFileCancelOne(t *testing.T) {
	t.Parallel()
	watchFileProduct(t, "cancel_one", "control")
}
func TestProduct_WatchFileCancelAll(t *testing.T) {
	t.Parallel()
	watchFileProduct(t, "cancel_all", "control")
}
func TestProduct_WatchFileDuplicateListener(t *testing.T) {
	t.Parallel()
	watchFileProduct(t, "duplicate", "control")
}
func TestProduct_WatchFileListenerBatch(t *testing.T) {
	t.Parallel()
	watchFileProduct(t, "batch", "control")
}
func TestProduct_WatchFileUnref(t *testing.T) { t.Parallel(); watchFileProduct(t, "unref", "control") }
func TestProduct_WatchFileThrowingListener(t *testing.T) {
	t.Parallel()
	watchFileProduct(t, "listener_throw", "control")
}
func TestProduct_WatchFileMutantSizeOnly(t *testing.T) {
	t.Parallel()
	watchFileProduct(t, "mtime", "size-only")
}
func TestProduct_WatchFileMutantInitialMissing(t *testing.T) {
	t.Parallel()
	watchFileProduct(t, "missing", "initial-missing-suppressed")
}
func TestProduct_WatchFileMutantUnref(t *testing.T) {
	t.Parallel()
	watchFileProduct(t, "unref", "unref-ignored")
}
func TestProduct_WatchFileMutantListenerIdentity(t *testing.T) {
	t.Parallel()
	watchFileProduct(t, "cancel_one", "listener-identity-ignored")
}
func TestProduct_WatchFileMutantThrowingListener(t *testing.T) {
	t.Parallel()
	watchFileProduct(t, "listener_throw", "throw-ignored")
}

func TestWatchFileTscCloseShapeMatchesNode(t *testing.T) {
	t.Parallel()
	watchFileAgreement(t, "tsc_close")
}
func TestProduct_WatchFileTscCloseShape(t *testing.T) {
	t.Parallel()
	watchFileProduct(t, "tsc_close", "control")
}

func TestWatchFileProductOrderIsCurrent(t *testing.T) {
	t.Parallel()
	script, err := filepath.Abs(filepath.Join(repository, "cloud/fast-gate/watchfile-product-order.py"))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	output, err := exec.CommandContext(ctx, "python3", script, "--check").CombinedOutput()
	if err != nil {
		t.Fatalf("watchFile product generator: %v\n%s", err, output)
	}
}
