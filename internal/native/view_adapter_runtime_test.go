package native

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type viewAdapterRun struct {
	stdout, stderr string
	exitCode       int
}

// Each snapshot builds the real cache/heap/weak/count units. Other heap kinds
// deliberately abort in support.c: no closure lifetime is replaced by a stub.
func viewAdapterFixture(t *testing.T, fixture string, sanitize bool, mutant string, arguments ...string) viewAdapterRun {
	t.Helper()
	directory := t.TempDir()
	headers, err := filepath.Glob(filepath.Join("runtime", "*.h"))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range headers {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, filepath.Base(path)), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	var sources []string
	for _, name := range []string{"closure.c", "heap.c", "weak.c", "count.c"} {
		data, err := os.ReadFile(filepath.Join("runtime", name))
		if err != nil {
			t.Fatal(err)
		}
		source := string(data)
		replace := func(before, after string) {
			if strings.Count(source, before) != 1 {
				t.Fatalf("%s mutation site changed in %s", mutant, name)
			}
			source = strings.Replace(source, before, after, 1)
		}
		if name == "closure.c" {
			switch mutant {
			case "skip-intern":
				replace("if (held->view_underlying == underlying && held->view_key == view_key)", "if (false && held->view_underlying == underlying && held->view_key == view_key)")
			case "dying-retain":
				replace("if (value->heap.references == 0) {\n\t\treturn NULL;\n\t}", "/* mutant: count zero is treated as live */")
			case "skip-recheck":
				replace("held = view_adapter_find_locked(underlying, view_key);\n\tif (held == NULL)", "held = NULL;\n\tif (held == NULL)")
			case "canonical-factory":
				replace("|| fresh->canonical_owner != NULL", "|| false")
			case "strong-leak-root":
				replace("(uintptr_t)value ^ VIEW_ADAPTER_HIDDEN", "(uintptr_t)value")
				replace("(adamic_closure *)(value ^ VIEW_ADAPTER_HIDDEN)", "(adamic_closure *)value")
			}
			if fixture == "order" {
				source += `
bool view_adapter_listed_for_test(adamic_closure *adapter) {
    view_adapter_lock_table();
    bool found = false;
    for (adamic_closure *held = view_adapter_reveal(view_adapter_first); held != NULL; held = view_adapter_reveal(held->view_next)) {
        found = found || held == adapter;
    }
    view_adapter_unlock_table();
    return found;
}
`
			}
		}
		if name == "heap.c" {
			if fixture == "dying" {
				replace("#include \"adamic.h\"", "#include \"adamic.h\"\nextern void view_adapter_dying_probe(adamic_closure *);")
				replace("adamic_view_adapter_forget(closure);", "view_adapter_dying_probe(closure);\n\t\t\tadamic_view_adapter_forget(closure);")
			}
			if fixture == "order" {
				replace("#include \"adamic.h\"", "#include \"adamic.h\"\nextern void view_adapter_order_probe(adamic_closure *);")
				replace("adamic_view_adapter_forget(closure);", "adamic_view_adapter_forget(closure);\n\t\t\tview_adapter_order_probe(closure);")
				if mutant == "release-before-remove" {
					replace("adamic_view_adapter_forget(closure);\n\t\t\tview_adapter_order_probe(closure);\n\t\t\tlet_go(closure->view_underlying);", "let_go(closure->view_underlying);\n\t\t\tview_adapter_order_probe(closure);\n\t\t\tadamic_view_adapter_forget(closure);")
				}
			}
		}
		path := filepath.Join(directory, name)
		if err := os.WriteFile(path, []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
		sources = append(sources, path)
	}
	binary := filepath.Join(directory, "fixture")
	flags := append(Flags(Options{Sanitize: sanitize, Count: true}), "-DADAMIC_CANONICAL_CLOSURES=1", "-pthread", "-I", directory, "-I", filepath.Join("testdata", "view-adapters"), "-o", binary)
	flags = append(flags, sources...)
	flags = append(flags, filepath.Join("testdata", "view-adapters", "support.c"), filepath.Join("testdata", "view-adapters", fixture+".c"))
	buildContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if output, err := exec.CommandContext(buildContext, compilerName(Options{}), flags...).CombinedOutput(); err != nil {
		t.Fatalf("C fixture must compile, including mutant: %v\n%s", err, output)
	}
	runContext, cancelRun := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelRun()
	command := exec.CommandContext(runContext, binary, arguments...)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "LSAN_OPTIONS=exitcode=23")
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err = command.Run()
	result := viewAdapterRun{stdout: stdout.String(), stderr: stderr.String()}
	if err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) {
			t.Fatalf("fixture did not exit: %v\n%s", err, stderr.String())
		}
		result.exitCode = exit.ExitCode()
	}
	if runContext.Err() != nil {
		t.Fatalf("factory/cleanup must finish without deadlock: %v", runContext.Err())
	}
	return result
}

func viewAdapterClean(t *testing.T, fixture, output string) {
	t.Helper()
	for _, sanitize := range []bool{false, true} {
		got := viewAdapterFixture(t, fixture, sanitize, "")
		if got.exitCode != 0 || got.stdout != output+"\n" || strings.Contains(got.stderr, "Sanitizer") {
			t.Fatalf("sanitize=%t: %#v", sanitize, got)
		}
		t.Logf("%s sanitize=%t clean: %s", fixture, sanitize, strings.TrimSpace(got.stderr))
	}
}

func TestViewAdapterCacheIntern(t *testing.T) {
	t.Parallel()
	viewAdapterClean(t, "intern", "interned and removed")
}

func TestViewAdapterCacheBounded(t *testing.T) {
	t.Parallel()
	viewAdapterClean(t, "bounded", "bounded allocations and retains")
	for _, sanitize := range []bool{false, true} {
		got := viewAdapterFixture(t, "bounded", sanitize, "skip-intern")
		if got.exitCode != 1 || !strings.Contains(got.stderr, "bounded allocation exceeded") || strings.Contains(got.stderr, "Sanitizer") {
			t.Fatalf("skip-intern must fail the allocation bound cleanly: %#v", got)
		}
		t.Logf("skip-intern caught sanitize=%t: %s", sanitize, strings.TrimSpace(got.stderr))
	}
}

func TestViewAdapterCacheRecheck(t *testing.T) {
	t.Parallel()
	viewAdapterClean(t, "recheck", "factory outside lock and winner rechecked")
	got := viewAdapterFixture(t, "recheck", true, "skip-recheck")
	if got.exitCode != 1 || !strings.Contains(got.stderr, "factory winner was not rechecked") || strings.Contains(got.stderr, "Sanitizer") {
		t.Fatalf("skip-recheck must fail winner selection cleanly: %#v", got)
	}
	t.Logf("skip-recheck caught: %s", strings.TrimSpace(got.stderr))
}

func TestViewAdapterCacheDying(t *testing.T) {
	t.Parallel()
	viewAdapterClean(t, "dying", "dying adapters not resurrected")
	got := viewAdapterFixture(t, "dying", true, "dying-retain")
	if got.exitCode == 0 || !strings.Contains(got.stderr, "ERROR: AddressSanitizer: heap-use-after-free") {
		t.Fatalf("dying-retain mutant must be caught by ASan: %#v", got)
	}
	t.Logf("dying-retain caught by ASan in the actual count-zero destruction window:\n%s", got.stderr)
}

func TestViewAdapterCacheOrder(t *testing.T) {
	t.Parallel()
	viewAdapterClean(t, "order", "removed before underlying release")
	got := viewAdapterFixture(t, "order", true, "release-before-remove")
	if got.exitCode != 1 || !strings.Contains(got.stderr, "remove/release order violated") || strings.Contains(got.stderr, "Sanitizer") {
		t.Fatalf("release-before-remove must fail the order observation cleanly: %#v", got)
	}
	t.Logf("release-before-remove caught: %s", strings.TrimSpace(got.stderr))
}

func TestViewAdapterCacheCanonical(t *testing.T) {
	t.Parallel()
	viewAdapterClean(t, "canonical", "adapter remains noncanonical")
	got := viewAdapterFixture(t, "canonical", true, "", "invalid-factory")
	if got.exitCode != 70 || !strings.Contains(got.stderr, "factory must return a fresh noncanonical counted closure") {
		t.Fatalf("canonical factory must be refused: %#v", got)
	}
	got = viewAdapterFixture(t, "canonical", true, "canonical-factory", "invalid-factory")
	if got.exitCode != 0 || got.stdout != "canonical result admitted\n" || strings.Contains(got.stderr, "Sanitizer") {
		t.Fatalf("canonical guard omission must admit the invalid result: %#v", got)
	}
	t.Log("canonical-factory omission caught by exit 0 instead of required exit 70")
}

func TestViewAdapterCacheWeakLeakRoot(t *testing.T) {
	t.Parallel()
	got := viewAdapterFixture(t, "leak", true, "")
	if got.exitCode != 23 || !strings.Contains(got.stderr, "ERROR: LeakSanitizer: detected memory leaks") {
		t.Fatalf("weak cache must expose an abandoned adapter to LeakSanitizer: %#v", got)
	}
	t.Logf("intentional abandoned adapter detected:\n%s", got.stderr)
	got = viewAdapterFixture(t, "leak", true, "strong-leak-root")
	if got.exitCode != 0 || strings.Contains(got.stderr, "Sanitizer") {
		t.Fatalf("unhidden cache-root mutant must hide the leak: %#v", got)
	}
	t.Log("strong-leak-root mutant caught: LeakSanitizer lost the required leak report")
}

func TestViewAdapterRuntimeStaticsPending(t *testing.T) {
	t.Parallel()
	// runtime's condition 3: one locked global table, statics doc, ThreadSanitizer re-view.
	// List closure.c:view_adapter_first and view_adapter_lock in docs/runtime-statics.md
	// and runtime_statics_parallel_test when those audit files land.
	t.Skip("awaits runtime's slice with the thread-safe heap, pool and region adoption (area/runtime)")
}

func TestViewAdapterPoolTSanPending(t *testing.T) {
	t.Parallel()
	// runtime's condition 3: one locked global table, statics doc, ThreadSanitizer re-view.
	// Pool workers must re-view one shared callable under ThreadSanitizer.
	t.Skip("awaits runtime's slice with the thread-safe heap, pool and region adoption (area/runtime)")
}

func TestViewAdapterRegionAdoptionPending(t *testing.T) {
	t.Parallel()
	// runtime's condition 4: adapters never region members, runtime's slice 6c #c2zbg7a.
	// Both graph and Program-region adoption must refuse an adapter.
	t.Skip("awaits runtime's slice with the thread-safe heap, pool and region adoption (area/runtime)")
}

func TestViewAdapterProgramUnderlyingPending(t *testing.T) {
	t.Parallel()
	// runtime's condition 4: adapters never region members, runtime's slice 6c #c2zbg7a.
	// Exercise an adapter over a canonical closure in the real Program region.
	t.Skip("awaits runtime's slice with the thread-safe heap, pool and region adoption (area/runtime)")
}
