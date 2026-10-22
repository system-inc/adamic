package native

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/boundedrun"

	"github.com/system-inc/adamic/internal/buildcache"
)

type asyncCountProtocolResult struct {
	stdout, stderr []byte
	exit           int
}

func asyncCountProtocolRun(t *testing.T, command string, arguments ...string) asyncCountProtocolResult {
	t.Helper()
	process, cancel := boundedrun.Command(10*time.Second, command, arguments...)
	defer cancel()
	leaks := "0"
	if goruntime.GOOS == "linux" {
		leaks = "1"
	}
	process.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks="+leaks+":halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1")
	var stdout, stderr bytes.Buffer
	process.Stdout, process.Stderr = &stdout, &stderr
	err := process.Run()
	exit := 0
	if err != nil {
		if process.ProcessState == nil {
			t.Fatal(err)
		}
		exit = process.ProcessState.ExitCode()
	}
	return asyncCountProtocolResult{stdout.Bytes(), stderr.Bytes(), exit}
}

const asyncCountProtocol = `#include "async.h"
#include <assert.h>
#include <string.h>
static void children(adamic_async_frame *frame,void (*drop)(void *)){(void)frame;(void)drop;}
static void resume(adamic_async_frame *frame,adamic_value value,bool rejected){
 assert(!rejected && frame->waiting==NULL);
 adamic_write_line(adamic_stdout,value.reference);
}
static void subscribe(adamic_async_promise *promise){
 adamic_async_frame *frame=adamic_allocate(sizeof *frame,adamic_kind_async_frame);
 memset((char *)frame+sizeof(adamic_heap),0,sizeof *frame-sizeof(adamic_heap));
 frame->resume=resume;frame->children=children;
 adamic_async_await(frame,promise);adamic_release(frame);
}
int main(int argc,char **argv){
 assert(argc==2);adamic_start(argc,argv);
 adamic_async_promise *promise=adamic_async_new();
 adamic_string *text=adamic_string_from_number(12345);
 bool settled=strcmp(argv[1],"settled")==0;
 if(settled)adamic_async_settle(promise,(adamic_value){.reference=text},true,false);
 subscribe(promise);subscribe(promise);
 if(!settled)adamic_async_settle(promise,(adamic_value){.reference=text},true,false);
 adamic_release(text);adamic_release(promise);adamic_async_run();return 0;
}
`

func asyncCountProtocolProduct(t *testing.T) string {
	t.Helper()
	options := Options{Sanitize: true}
	files, flags, compiler, version, err := runtimeInputs("", options, nil)
	if err != nil {
		t.Fatal(err)
	}
	headers, err := headersRead(files, flags, compiler)
	if err != nil {
		t.Fatal(err)
	}
	inputs := buildcache.Inputs{Name: "async-count-protocol", Files: []string{"internal/native/async_counts_test.go", "internal/native/runtime", "internal/native/native.go", "internal/native/library.go", "internal/native/build_runtime.go", "internal/native/target.go"}, Flags: portableKeyParts(files, flags, compiler, headers), Toolchain: []string{buildcache.Tool(compiler, "--version"), buildcache.Tool(archiverName(compiler), "--version")}}
	return buildcache.Product(t, inputs, func(directory string) error {
		cache, err := cacheDirectory("async-count-protocol")
		if err != nil {
			return err
		}
		baseline, err := cachedRuntime(files, flags, compiler, version, cache)
		if err != nil {
			return err
		}
		if err := BuildWithRuntime(asyncCountProtocol, filepath.Join(directory, "control"), options, baseline); err != nil {
			return err
		}
		source, err := runtime.ReadFile("runtime/async.c")
		if err != nil {
			return err
		}
		for _, mutant := range []struct{ name, before, after string }{
			{"promise", "reaction->promise = adamic_retain(promise);", "reaction->promise = promise;"},
			{"frame", "reaction->frame = adamic_retain(frame);", "reaction->frame = frame;"},
			{"payload", "if (references) adamic_retain(value.reference);", "/* payload owner dropped */"},
			{"waiting", "frame->waiting = adamic_retain(promise);", "frame->waiting = promise;"},
		} {
			text := string(source)
			if !strings.Contains(text, mutant.before) {
				return fmt.Errorf("missing mutant seam: %s", mutant.name)
			}
			scratch := t.TempDir()
			for _, file := range files {
				if strings.HasSuffix(file.name, ".h") {
					if err := os.WriteFile(filepath.Join(scratch, file.name), file.contents, 0600); err != nil {
						return err
					}
				}
			}
			if err := os.WriteFile(filepath.Join(scratch, "async.c"), []byte(strings.ReplaceAll(text, mutant.before, mutant.after)), 0600); err != nil {
				return err
			}
			object := filepath.Join(scratch, "async.o")
			arguments := append(append(append([]string{}, flags...), prefixMaps(scratch, "/adamic-runtime")...), "-I", ".", "-c", "async.c", "-o", "async.o")
			command, release := boundedrun.Command(45*time.Second, compiler, arguments...)
			command.Dir = scratch
			output, compileErr := command.CombinedOutput()
			release()
			if compileErr != nil {
				return fmt.Errorf("mutant must compile: %w\n%s", compileErr, output)
			}
			archive, err := os.ReadFile(baseline)
			if err != nil {
				return err
			}
			library := filepath.Join(scratch, "runtime.a")
			if err := os.WriteFile(library, archive, 0600); err != nil {
				return err
			}
			staticsCompile(t, archiverName(compiler), "r", library, object)
			if err := BuildWithRuntime(asyncCountProtocol, filepath.Join(directory, mutant.name), options, library); err != nil {
				return err
			}
		}
		return nil
	})
}
func TestProduct_AsyncCountProtocol(t *testing.T) { t.Parallel(); asyncCountProtocolProduct(t) }
func checkAsyncCountProtocol(t *testing.T, mode, mutant string) {
	t.Helper()
	result := asyncCountProtocolRun(t, filepath.Join(asyncCountProtocolProduct(t), mutant), mode)
	if mutant == "control" {
		if result.exit != 0 || string(result.stdout) != "12345\n12345\n" || len(result.stderr) != 0 {
			t.Fatalf("protocol control: %+v", result)
		}
	} else if result.exit == 0 || !strings.Contains(string(result.stderr), "AddressSanitizer: heap-use-after-free") {
		t.Fatalf("required %s owner was not caught by ASan: exit %d stdout %q stderr %s", mutant, result.exit, result.stdout, result.stderr)
	}
	t.Logf("%s %s: exit %d", mode, mutant, result.exit)
}
func TestAsyncCountSettledProtocol(t *testing.T) {
	t.Parallel()
	checkAsyncCountProtocol(t, "settled", "control")
}
func TestAsyncCountPendingProtocol(t *testing.T) {
	t.Parallel()
	checkAsyncCountProtocol(t, "pending", "control")
}
func TestAsyncCountQueuedPromiseMutant(t *testing.T) {
	t.Parallel()
	checkAsyncCountProtocol(t, "settled", "promise")
}
func TestAsyncCountSettlementPromiseMutant(t *testing.T) {
	t.Parallel()
	checkAsyncCountProtocol(t, "pending", "promise")
}
func TestAsyncCountReactionFrameMutant(t *testing.T) {
	t.Parallel()
	checkAsyncCountProtocol(t, "pending", "frame")
}
func TestAsyncCountPayloadOwnerRuntimeMutant(t *testing.T) {
	t.Parallel()
	checkAsyncCountProtocol(t, "settled", "payload")
}
func TestAsyncCountPendingOwnerMutant(t *testing.T) {
	t.Parallel()
	checkAsyncCountProtocol(t, "pending", "waiting")
}
