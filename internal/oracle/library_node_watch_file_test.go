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

var watchFileFixtures = []string{
	"missing",
	"appearance",
	"mtime",
	"cancel_one",
	"cancel_all",
	"duplicate",
	"batch",
	"unref",
	"listener_throw",
	"tsc_close",
}

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
func watchFileRun(test *testing.T, name string, args ...string) (run, bool) {
	test.Helper()
	path, err := toolchain.Path(name)
	if err != nil {
		test.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), watchFileDeadline)
	defer cancel()
	command := exec.CommandContext(ctx, path, args...)
	command.Dir = test.TempDir()
	command.Env = []string{
		"TMPDIR=" + command.Dir,
		"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1",
		"UBSAN_OPTIONS=halt_on_error=1",
	}
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error { return syscall.Kill(-command.Process.Pid, syscall.SIGKILL) }
	stdout, stderr, err := childguard.Capture(ctx, command)
	expired := ctx.Err() == context.DeadlineExceeded
	var exited *exec.ExitError
	if err != nil && !expired && !errors.As(err, &exited) {
		test.Fatal(err)
	}
	if command.ProcessState == nil {
		test.Fatal("watch fixture did not start")
	}
	return run{stdout: stdout, stderr: stderr, exitCode: command.ProcessState.ExitCode()}, expired
}
func watchFilePath(test *testing.T, name string) string {
	test.Helper()
	path, err := filepath.Abs("testdata/host_watch_file/" + name + ".a")
	if err != nil {
		test.Fatal(err)
	}
	return path
}
func watchFileNode(test *testing.T, path string) run {
	test.Helper()
	loader, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		test.Fatal(err)
	}
	result, expired := watchFileRun(test, "node", "--disable-warning=ExperimentalWarning", loader, path)
	if expired {
		test.Fatal("Node watch fixture exceeded its own deadline")
	}
	return result
}

type watchFileMutation struct {
	name, fixture, file, before, after, catcher string
	deadline                                    bool
}

var watchFileMutations = []watchFileMutation{
	{
		"size-only",
		"mtime",
		"node_watch_file_impl.h",
		"if (old->value.st_size != current->value.st_size) return true;",
		"return old->value.st_size != current->value.st_size;",
		"TestWatchFileSameSizeMtimeMatchesNode",
		true,
	},
	{
		"initial-missing-suppressed",
		"missing",
		"node_watch_file_impl.h",
		"status != 0 ? poller->status != status :",
		"status != 0 ? poller->status != 0 && poller->status != status :",
		"TestWatchFileInitialMissingMatchesNode",
		true,
	},
	{
		"unref-ignored",
		"unref",
		"node_watch_file_impl.h",
		"if (!initial && !poller->persistent)",
		"if (false)",
		"TestWatchFileUnrefInitialStatMatchesNode",
		true,
	},
	{
		"listener-identity-ignored",
		"cancel_one",
		"node_watch_file_impl.h",
		"if ((*at)->callback == callback) match = at;",
		"if ((*at)->callback != callback) match = at;",
		"TestWatchFileCancelOneMatchesNode",
		false,
	},
	{
		"throw-ignored",
		"listener_throw",
		"timers_impl.h",
		"if (adamic_thrown != NULL) adamic_uncaught();",
		"/* mutant: ignore the thrown listener */",
		"TestWatchFileThrowingListenerMatchesNode",
		false,
	},
}

func watchFileProduct(test *testing.T, fixture, variant string) string {
	test.Helper()
	flags := append(native.Flags(native.Options{Sanitize: true}), "explicit-child-environment", "watch-file-v1")
	for _, name := range []string{
		"ADAMIC_PROGRAM_REGION",
		"ADAMIC_NATIVE_SPLIT",
		"ADAMIC_NATIVE_JOBS",
		"CPATH",
		"C_INCLUDE_PATH",
		"SDKROOT",
		"MACOSX_DEPLOYMENT_TARGET",
	} {
		value, set := os.LookupEnv(name)
		flags = append(flags, fmt.Sprintf("env:%s:set=%t:value=%s", name, set, value))
	}
	inputs := buildcache.Inputs{Name: "oracle-watch-file-" + fixture + "-" + variant, Files: append(
		compilerinputs.Lowering(),
		"internal/oracle/library_node_watch_file_test.go",
		"internal/oracle/testdata/host_watch_file/"+fixture+".a",
	),
		Flags: flags,
		Toolchain: []string{
			runtime.Version(),
			runtime.GOOS + "/" + runtime.GOARCH,
			buildcache.Tool("clang", "--version"),
			buildcache.Tool("ar", "--version"),
		}}
	return buildcache.Product(test, inputs, func(directory string) error {
		program, err := lowered(test, watchFilePath(test, fixture))
		if err != nil {
			return err
		}
		source := native.C(program)
		if err := os.WriteFile(
			filepath.Join(directory, "program.mjs"),
			[]byte(javascript.JavaScript(program)),
			0600,
		); err != nil {
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
func watchFileAgreement(test *testing.T, name string) {
	test.Helper()
	product := watchFileProduct(test, name, "control")
	truth := watchFileNode(test, watchFilePath(test, name))
	wantExit := 0
	if name == "listener_throw" {
		wantExit = 70
	}
	if truth.exitCode != wantExit {
		test.Fatalf("Node watch fixture: %+v", truth)
	}
	for _, backend := range []string{"native", "javascript"} {
		var got run
		var expired bool
		if backend == "native" {
			got, expired = watchFileRun(test, filepath.Join(product, "native"))
		} else {
			got = watchFileNode(test, filepath.Join(product, "program.mjs"))
		}
		if expired {
			test.Fatal("watchFile did not finish by its own 5 s deadline")
		}
		if difference := disagreement(truth, got); difference != "" {
			test.Fatalf(
				"%s: %s; stdout=%q stderr=%q exit=%d, Node stdout=%q stderr=%q exit=%d",
				backend,
				difference,
				got.stdout,
				got.stderr,
				got.exitCode,
				truth.stdout,
				truth.stderr,
				truth.exitCode,
			)
		}
	}
}
func TestWatchFileInitialMissingMatchesNode(test *testing.T) {
	test.Parallel()
	watchFileAgreement(test, "missing")
}
func TestWatchFileAppearanceAndDeletionMatchNode(test *testing.T) {
	test.Parallel()
	watchFileAgreement(test, "appearance")
}
func TestWatchFileSameSizeMtimeMatchesNode(test *testing.T) {
	test.Parallel()
	watchFileAgreement(test, "mtime")
}
func TestWatchFileCancelOneMatchesNode(test *testing.T) {
	test.Parallel()
	watchFileAgreement(test, "cancel_one")
}
func TestWatchFileCancelAllMatchesNode(test *testing.T) {
	test.Parallel()
	watchFileAgreement(test, "cancel_all")
}
func TestWatchFileDuplicateListenerMatchesNode(test *testing.T) {
	test.Parallel()
	watchFileAgreement(test, "duplicate")
}
func TestWatchFileListenerBatchMatchesNode(test *testing.T) {
	test.Parallel()
	watchFileAgreement(test, "batch")
}
func TestWatchFileUnrefInitialStatMatchesNode(test *testing.T) {
	test.Parallel()
	watchFileAgreement(test, "unref")
}
func TestWatchFileThrowingListenerMatchesNode(test *testing.T) {
	test.Parallel()
	watchFileAgreement(test, "listener_throw")
}

func watchFileMutant(test *testing.T, name string) {
	test.Helper()
	var mutation *watchFileMutation
	for i := range watchFileMutations {
		if watchFileMutations[i].name == name {
			mutation = &watchFileMutations[i]
		}
	}
	if mutation == nil {
		test.Fatal(name)
	}
	binary := filepath.Join(watchFileProduct(test, mutation.fixture, name), "native")
	got, expired := watchFileRun(test, binary)
	truth := watchFileNode(test, watchFilePath(test, mutation.fixture))
	if mutation.deadline {
		if !expired {
			test.Fatalf("%s must miss its own deadline: %+v", name, got)
		}
	} else {
		if expired {
			test.Fatal("mutant must complete, not time out")
		}
		if bytes.Contains(got.stderr, []byte("Sanitizer")) || bytes.Contains(got.stderr, []byte("runtime error:")) {
			test.Fatalf("mutant must be sanitizer clean: %+v", got)
		}
		if difference := disagreement(truth, got); difference == "" {
			test.Fatalf("%s survived", name)
		}
	}
	test.Logf("%s goes red with %s", mutation.catcher, name)
}
func TestWatchFileMutantIgnoresSameSizeMtime(test *testing.T) {
	test.Parallel()
	watchFileMutant(test, "size-only")
}
func TestWatchFileMutantSuppressesInitialMissing(test *testing.T) {
	test.Parallel()
	watchFileMutant(test, "initial-missing-suppressed")
}
func TestWatchFileMutantKeepsUnrefAlive(test *testing.T) {
	test.Parallel()
	watchFileMutant(test, "unref-ignored")
}
func TestWatchFileMutantIgnoresListenerIdentity(test *testing.T) {
	test.Parallel()
	watchFileMutant(test, "listener-identity-ignored")
}
func TestWatchFileMutantIgnoresThrownListener(test *testing.T) {
	test.Parallel()
	watchFileMutant(test, "throw-ignored")
}

func TestProduct_WatchFileInitialMissing(test *testing.T) {
	test.Parallel()
	watchFileProduct(test, "missing", "control")
}
func TestProduct_WatchFileAppearance(test *testing.T) {
	test.Parallel()
	watchFileProduct(test, "appearance", "control")
}
func TestProduct_WatchFileSameSizeMtime(test *testing.T) {
	test.Parallel()
	watchFileProduct(test, "mtime", "control")
}
func TestProduct_WatchFileCancelOne(test *testing.T) {
	test.Parallel()
	watchFileProduct(test, "cancel_one", "control")
}
func TestProduct_WatchFileCancelAll(test *testing.T) {
	test.Parallel()
	watchFileProduct(test, "cancel_all", "control")
}
func TestProduct_WatchFileDuplicateListener(test *testing.T) {
	test.Parallel()
	watchFileProduct(test, "duplicate", "control")
}
func TestProduct_WatchFileListenerBatch(test *testing.T) {
	test.Parallel()
	watchFileProduct(test, "batch", "control")
}
func TestProduct_WatchFileUnref(test *testing.T) {
	test.Parallel()
	watchFileProduct(test, "unref", "control")
}
func TestProduct_WatchFileThrowingListener(test *testing.T) {
	test.Parallel()
	watchFileProduct(test, "listener_throw", "control")
}
func TestProduct_WatchFileMutantSizeOnly(test *testing.T) {
	test.Parallel()
	watchFileProduct(test, "mtime", "size-only")
}
func TestProduct_WatchFileMutantInitialMissing(test *testing.T) {
	test.Parallel()
	watchFileProduct(test, "missing", "initial-missing-suppressed")
}
func TestProduct_WatchFileMutantUnref(test *testing.T) {
	test.Parallel()
	watchFileProduct(test, "unref", "unref-ignored")
}
func TestProduct_WatchFileMutantListenerIdentity(test *testing.T) {
	test.Parallel()
	watchFileProduct(test, "cancel_one", "listener-identity-ignored")
}
func TestProduct_WatchFileMutantThrowingListener(test *testing.T) {
	test.Parallel()
	watchFileProduct(test, "listener_throw", "throw-ignored")
}

func TestWatchFileTscCloseShapeMatchesNode(test *testing.T) {
	test.Parallel()
	watchFileAgreement(test, "tsc_close")
}
func TestProduct_WatchFileTscCloseShape(test *testing.T) {
	test.Parallel()
	watchFileProduct(test, "tsc_close", "control")
}

func TestWatchFileProductOrderIsCurrent(test *testing.T) {
	test.Parallel()
	script, err := filepath.Abs(filepath.Join(repository, "cloud/fast-gate/watchfile-product-order.py"))
	if err != nil {
		test.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	output, err := exec.CommandContext(ctx, "python3", script, "--check").CombinedOutput()
	if err != nil {
		test.Fatalf("watchFile product generator: %v\n%s", err, output)
	}
}
