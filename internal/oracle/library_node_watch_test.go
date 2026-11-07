package oracle

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/compilerinputs"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

// fsWatchFixtures are held to Node on native under the sanitizers and on the JavaScript backend
// here, and run in the ordinary native, JavaScript, gob and counted families too; WASI refuses their
// host entry by name (wasi_refusals_test.go). Each does its file system work before the loop runs, so
// both runtimes read the same inotify queue, and each uses one watcher, since Node doesn't define the
// order across watchers in one turn. The rename row of Node's measured table is the C harness's
// (internal/native/fswatch_test.go): no program can rename yet.
var fsWatchFixtures = []string{"tsc_shape", "write", "create", "unlink", "mkdir", "self_delete", "self_close", "unref", "missing", "close_twice", "on_returns", "two_events"}

func init() {
	for _, name := range fsWatchFixtures {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/host_watch/" + name + ".a", true, false})
	}
}

func fsWatchPath(t *testing.T, name string) string {
	t.Helper()
	path, err := filepath.Abs("testdata/host_watch/" + name + ".a")
	if err != nil {
		t.Fatal(err)
	}
	return path
}

// A mutation's catch is its own deadline, a disagreement with Node, or, for a fixture that stops
// by name where Node would go on, a run that doesn't stop that way.
type fsWatchMutation struct {
	name, fixture, before, after, catcher, stop string
	deadline                                    bool
}

// fsWatchMutations are real edits to node_fs_watch_impl.h, each turning its named pin red.
var fsWatchMutations = []fsWatchMutation{
	{"change-rename-swapped", "create", `!= 0 ? "rename" : "change";`, `!= 0 ? "change" : "rename";`, "TestFSWatchCreateMatchesNode", "", false},
	{"self-basename-dropped", "write", "event->len != 0 ? event->name : watcher->base", `event->len != 0 ? event->name : ""`, "TestFSWatchWriteMatchesNode", "", false},
	{"close-keeps-source", "self_close", "    if (fs_watchers == NULL) fs_watch_stop(); else fs_watch_reference();\n", "    /* mutant: the last close leaves the shared source registered */\n    if (fs_watchers != NULL) fs_watch_reference();\n", "TestFSWatchSelfCloseExitsMatchesNode", "", true},
	{"event-drain-dropped", "two_events", "                adamic_source_drain_jobs();\n", "                /* mutant: one checkpoint for the whole read */\n", "TestFSWatchTwoEventsMatchesNode", "", false},
	{"view-name-check-dropped", "view_other_event", `    if (event == NULL || event->heap.kind != adamic_kind_string || event->length != 5 || memcmp(event->bytes, "error", 5) != 0) {`, "    if (event == NULL && false) {", "TestFSWatchViewOtherEventStopsByName", fsWatchOtherEventStop, false},
}

const (
	fsWatchOtherEventStop = "adamic: panic: NotYet: FSWatcher.on with an event other than 'error'\n"
	fsWatchAfterCloseStop = "adamic: panic: NotYet: FSWatcher.on through a structural view after close\n"
)

func fsWatchProduct(t *testing.T, fixture, variant string) string {
	t.Helper()
	flags := append(native.Flags(native.Options{Sanitize: true}), "explicit-child-environment", "fs-watch-v1")
	for _, name := range []string{"ADAMIC_PROGRAM_REGION", "ADAMIC_NATIVE_SPLIT", "ADAMIC_NATIVE_JOBS", "CPATH", "C_INCLUDE_PATH", "SDKROOT", "MACOSX_DEPLOYMENT_TARGET"} {
		value, set := os.LookupEnv(name)
		flags = append(flags, fmt.Sprintf("env:%s:set=%t:value=%s", name, set, value))
	}
	inputs := buildcache.Inputs{Name: "oracle-fs-watch-" + fixture + "-" + variant, Files: append(compilerinputs.Lowering(), "internal/oracle/library_node_watch_test.go", "internal/oracle/testdata/host_watch/"+fixture+".a"),
		Flags:     flags,
		Toolchain: []string{runtime.Version(), runtime.GOOS + "/" + runtime.GOARCH, buildcache.Tool("clang", "--version"), buildcache.Tool("ar", "--version")}}
	return buildcache.Product(t, inputs, func(directory string) error {
		program, err := lowered(t, fsWatchPath(t, fixture))
		if err != nil {
			return err
		}
		source := native.C(program)
		if err := os.WriteFile(filepath.Join(directory, "program.mjs"), []byte(javascript.JavaScript(program)), 0600); err != nil {
			return err
		}
		options := native.Options{Sanitize: true}
		if variant == "control" {
			if err := native.Build(source, filepath.Join(directory, "native-release"), native.Options{Release: true}); err != nil {
				return err
			}
			return native.Build(source, filepath.Join(directory, "native"), options)
		}
		var mutation *fsWatchMutation
		for index := range fsWatchMutations {
			if fsWatchMutations[index].name == variant {
				mutation = &fsWatchMutations[index]
			}
		}
		if mutation == nil {
			return fmt.Errorf("unknown fs.watch variant %s", variant)
		}
		runtimeDirectory := filepath.Join(directory, "runtime-mutant")
		if err := os.Mkdir(runtimeDirectory, 0700); err != nil {
			return err
		}
		entries, err := os.ReadDir("../native/runtime")
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			data, err := os.ReadFile(filepath.Join("../native/runtime", entry.Name()))
			if err != nil {
				return err
			}
			if entry.Name() == "node_fs_watch_impl.h" {
				if strings.Count(string(data), mutation.before) != 1 {
					return fmt.Errorf("mutant %s must match once", variant)
				}
				data = []byte(strings.Replace(string(data), mutation.before, mutation.after, 1))
			}
			if err := os.WriteFile(filepath.Join(runtimeDirectory, entry.Name()), data, 0600); err != nil {
				return err
			}
		}
		library, err := native.RuntimeLibraryForSource(runtimeDirectory, source, options)
		if err != nil {
			return err
		}
		return native.BuildWithRuntime(source, filepath.Join(directory, "native"), options, library)
	})
}

func fsWatchAgreement(t *testing.T, name string) {
	t.Helper()
	product := fsWatchProduct(t, name, "control")
	truth := watchFileNode(t, fsWatchPath(t, name))
	if truth.exitCode != 0 {
		t.Fatalf("Node watch fixture: %+v", truth)
	}
	for _, backend := range []string{"native", "native-release", "javascript"} {
		var got run
		var expired bool
		if backend != "javascript" {
			got, expired = watchFileRun(t, filepath.Join(product, backend))
		} else {
			got = watchFileNode(t, filepath.Join(product, "program.mjs"))
		}
		if expired {
			t.Fatal("fs.watch did not finish by its own 5 s deadline")
		}
		if difference := disagreement(truth, got); difference != "" {
			t.Fatalf("%s: %s; stdout=%q stderr=%q exit=%d, Node stdout=%q stderr=%q exit=%d", backend, difference, got.stdout, got.stderr, got.exitCode, truth.stdout, truth.stderr, truth.exitCode)
		}
	}
}

func TestFSWatchTscShapeMatchesNode(t *testing.T) {
	t.Parallel()
	fsWatchAgreement(t, "tsc_shape")
}
func TestFSWatchWriteMatchesNode(t *testing.T) {
	t.Parallel()
	fsWatchAgreement(t, "write")
}
func TestFSWatchCreateMatchesNode(t *testing.T) {
	t.Parallel()
	fsWatchAgreement(t, "create")
}
func TestFSWatchUnlinkMatchesNode(t *testing.T) {
	t.Parallel()
	fsWatchAgreement(t, "unlink")
}
func TestFSWatchMkdirMatchesNode(t *testing.T) {
	t.Parallel()
	fsWatchAgreement(t, "mkdir")
}
func TestFSWatchSelfDeleteMatchesNode(t *testing.T) {
	t.Parallel()
	fsWatchAgreement(t, "self_delete")
}
func TestFSWatchSelfCloseExitsMatchesNode(t *testing.T) {
	t.Parallel()
	fsWatchAgreement(t, "self_close")
}
func TestFSWatchUnrefExitsMatchesNode(t *testing.T) {
	t.Parallel()
	fsWatchAgreement(t, "unref")
}
func TestFSWatchMissingThrowsMatchesNode(t *testing.T) {
	t.Parallel()
	fsWatchAgreement(t, "missing")
}
func TestFSWatchCloseTwiceMatchesNode(t *testing.T) {
	t.Parallel()
	fsWatchAgreement(t, "close_twice")
}
func TestFSWatchOnReturnsWatcherMatchesNode(t *testing.T) {
	t.Parallel()
	fsWatchAgreement(t, "on_returns")
}

// recursive: true stops by name on both backends, where Node would watch: tsc takes that arm only
// on win32 and darwin.
func TestFSWatchRecursiveStopsByName(t *testing.T) {
	t.Parallel()
	fsWatchStops(t, "recursive", "adamic: panic: NotYet: fs.watch with recursive: true\n")
}

// tsc's FsWatchWorkerWatcher view reaches on's slot, where lowering's literal "error" can't look: an
// event other than 'error', and an on after close, stop by name at run time on every leg.
func TestFSWatchViewOtherEventStopsByName(t *testing.T) {
	t.Parallel()
	fsWatchStops(t, "view_other_event", fsWatchOtherEventStop)
}
func TestFSWatchViewAfterCloseStopsByName(t *testing.T) {
	t.Parallel()
	fsWatchStops(t, "view_after_close", fsWatchAfterCloseStop)
}

// fsWatchStops holds a fixture Node would run on to one named stop on native, sanitized and in
// release, and on the JavaScript backend, before it prints anything.
func fsWatchStops(t *testing.T, fixture, want string) {
	t.Helper()
	product := fsWatchProduct(t, fixture, "control")
	for _, backend := range []string{"native", "native-release", "javascript"} {
		var got run
		var expired bool
		if backend != "javascript" {
			got, expired = watchFileRun(t, filepath.Join(product, backend))
		} else {
			got = watchFileNode(t, filepath.Join(product, "program.mjs"))
		}
		if expired || got.exitCode != 70 || string(got.stderr) != want || len(got.stdout) != 0 {
			t.Fatalf("%s: %+v", backend, got)
		}
	}
}

func fsWatchMutant(t *testing.T, name string) {
	t.Helper()
	var mutation *fsWatchMutation
	for index := range fsWatchMutations {
		if fsWatchMutations[index].name == name {
			mutation = &fsWatchMutations[index]
		}
	}
	if mutation == nil {
		t.Fatal(name)
	}
	got, expired := watchFileRun(t, filepath.Join(fsWatchProduct(t, mutation.fixture, name), "native"))
	if mutation.deadline {
		if bytes.Contains(got.stderr, []byte("Sanitizer")) || bytes.Contains(got.stderr, []byte("runtime error:")) {
			t.Fatalf("deadline mutant must be sanitizer clean: %s", got.stderr)
		}
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
		if mutation.stop != "" {
			if got.exitCode == 70 && string(got.stderr) == mutation.stop {
				t.Fatalf("%s survived", name)
			}
		} else if difference := disagreement(watchFileNode(t, fsWatchPath(t, mutation.fixture)), got); difference == "" {
			t.Fatalf("%s survived", name)
		}
	}
	t.Logf("%s goes red with %s", mutation.catcher, name)
}
func TestFSWatchMutantSwapsChangeAndRename(t *testing.T) {
	t.Parallel()
	fsWatchMutant(t, "change-rename-swapped")
}
func TestFSWatchMutantDropsSelfBasename(t *testing.T) {
	t.Parallel()
	fsWatchMutant(t, "self-basename-dropped")
}
func TestFSWatchMutantCloseKeepsSource(t *testing.T) {
	t.Parallel()
	fsWatchMutant(t, "close-keeps-source")
}
func TestFSWatchMutantDrainsOncePerRead(t *testing.T) {
	t.Parallel()
	fsWatchMutant(t, "event-drain-dropped")
}
func TestFSWatchMutantViewSkipsTheNameCheck(t *testing.T) {
	t.Parallel()
	fsWatchMutant(t, "view-name-check-dropped")
}

// Two events in one read, each listener's microtask between them (Runtime's must on #j71fn59).
func TestFSWatchTwoEventsMatchesNode(t *testing.T) {
	t.Parallel()
	fsWatchAgreement(t, "two_events")
}

func TestProduct_FSWatchTscShape(t *testing.T) {
	t.Parallel()
	fsWatchProduct(t, "tsc_shape", "control")
}
func TestProduct_FSWatchWrite(t *testing.T) {
	t.Parallel()
	fsWatchProduct(t, "write", "control")
}
func TestProduct_FSWatchCreate(t *testing.T) {
	t.Parallel()
	fsWatchProduct(t, "create", "control")
}
func TestProduct_FSWatchUnlink(t *testing.T) {
	t.Parallel()
	fsWatchProduct(t, "unlink", "control")
}
func TestProduct_FSWatchMkdir(t *testing.T) {
	t.Parallel()
	fsWatchProduct(t, "mkdir", "control")
}
func TestProduct_FSWatchSelfDelete(t *testing.T) {
	t.Parallel()
	fsWatchProduct(t, "self_delete", "control")
}
func TestProduct_FSWatchSelfClose(t *testing.T) {
	t.Parallel()
	fsWatchProduct(t, "self_close", "control")
}
func TestProduct_FSWatchUnref(t *testing.T) {
	t.Parallel()
	fsWatchProduct(t, "unref", "control")
}
func TestProduct_FSWatchMissing(t *testing.T) {
	t.Parallel()
	fsWatchProduct(t, "missing", "control")
}
func TestProduct_FSWatchCloseTwice(t *testing.T) {
	t.Parallel()
	fsWatchProduct(t, "close_twice", "control")
}
func TestProduct_FSWatchOnReturns(t *testing.T) {
	t.Parallel()
	fsWatchProduct(t, "on_returns", "control")
}
func TestProduct_FSWatchRecursive(t *testing.T) {
	t.Parallel()
	fsWatchProduct(t, "recursive", "control")
}
func TestProduct_FSWatchMutantSwapped(t *testing.T) {
	t.Parallel()
	fsWatchProduct(t, "create", "change-rename-swapped")
}
func TestProduct_FSWatchMutantBasename(t *testing.T) {
	t.Parallel()
	fsWatchProduct(t, "write", "self-basename-dropped")
}
func TestProduct_FSWatchMutantCloseKeepsSource(t *testing.T) {
	t.Parallel()
	fsWatchProduct(t, "self_close", "close-keeps-source")
}
func TestProduct_FSWatchTwoEvents(t *testing.T) {
	t.Parallel()
	fsWatchProduct(t, "two_events", "control")
}
func TestProduct_FSWatchViewOtherEvent(t *testing.T) {
	t.Parallel()
	fsWatchProduct(t, "view_other_event", "control")
}
func TestProduct_FSWatchViewAfterClose(t *testing.T) {
	t.Parallel()
	fsWatchProduct(t, "view_after_close", "control")
}
func TestProduct_FSWatchMutantDrainOncePerRead(t *testing.T) {
	t.Parallel()
	fsWatchProduct(t, "two_events", "event-drain-dropped")
}
func TestProduct_FSWatchMutantViewNameCheck(t *testing.T) {
	t.Parallel()
	fsWatchProduct(t, "view_other_event", "view-name-check-dropped")
}

// The tree builds each product above ahead by its row in cloud/fast-gate/product-order.json, which
// fswatch-product-order.py writes from this file's and internal/native/fswatch_test.go's
// TestProduct_ names. A product added without rerunning it reds this test.
func TestFSWatchProductOrderIsCurrent(t *testing.T) {
	t.Parallel()
	script, err := filepath.Abs(filepath.Join(repository, "cloud/fast-gate/fswatch-product-order.py"))
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
