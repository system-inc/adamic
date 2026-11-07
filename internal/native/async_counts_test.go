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

func asyncCountProtocolRun(test *testing.T, command string, arguments ...string) asyncCountProtocolResult {
	test.Helper()
	process, cancel := boundedrun.Command(10*time.Second, command, arguments...)
	defer cancel()
	leaks := "0"
	if goruntime.GOOS == "linux" {
		leaks = "1"
	}
	process.Env = append(
		os.Environ(),
		"ASAN_OPTIONS=detect_leaks="+leaks+":halt_on_error=1",
		"UBSAN_OPTIONS=halt_on_error=1",
	)
	var stdout, stderr bytes.Buffer
	process.Stdout, process.Stderr = &stdout, &stderr
	runError := process.Run()
	exit := 0
	if runError != nil {
		if process.ProcessState == nil {
			test.Fatal(runError)
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

func asyncCountProtocolProduct(test *testing.T) string {
	test.Helper()
	options := Options{Sanitize: true}
	files, flags, compiler, version, runtimeInputsError := runtimeInputs("", options, nil)
	if runtimeInputsError != nil {
		test.Fatal(runtimeInputsError)
	}
	headers, headersReadError := headersRead(files, flags, compiler)
	if headersReadError != nil {
		test.Fatal(headersReadError)
	}
	inputs := buildcache.Inputs{
		Name: "async-count-protocol",
		Files: []string{
			"internal/native/async_counts_test.go",
			"internal/native/runtime",
			"internal/native/native.go",
			"internal/native/library.go",
			"internal/native/build_runtime.go",
			"internal/native/target.go",
		},
		Flags: portableKeyParts(files, flags, compiler, headers),
		Toolchain: []string{
			buildcache.Tool(compiler, "--version"),
			buildcache.Tool(archiverName(compiler), "--version"),
		},
	}
	return buildcache.Product(test, inputs, func(directory string) error {
		cache, cacheDirectoryError := cacheDirectory("async-count-protocol")
		if cacheDirectoryError != nil {
			return cacheDirectoryError
		}
		baseline, cachedRuntimeError := cachedRuntime(files, flags, compiler, version, cache)
		if cachedRuntimeError != nil {
			return cachedRuntimeError
		}
		if buildWithRuntimeError := BuildWithRuntime(
			asyncCountProtocol,
			filepath.Join(directory, "control"),
			options,
			baseline,
		); buildWithRuntimeError != nil {
			return buildWithRuntimeError
		}
		source, readFileError := runtime.ReadFile("runtime/async.c")
		if readFileError != nil {
			return readFileError
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
			scratch := test.TempDir()
			for _, file := range files {
				if strings.HasSuffix(file.name, ".h") {
					if writeError := os.WriteFile(
						filepath.Join(scratch, file.name),
						file.contents,
						0600,
					); writeError != nil {
						return writeError
					}
				}
			}
			if writeError := os.WriteFile(
				filepath.Join(scratch, "async.c"),
				[]byte(strings.ReplaceAll(text, mutant.before, mutant.after)),
				0600,
			); writeError != nil {
				return writeError
			}
			object := filepath.Join(scratch, "async.o")
			arguments := append(
				append(append([]string{}, flags...), prefixMaps(scratch, "/adamic-runtime")...),
				"-I",
				".",
				"-c",
				"async.c",
				"-o",
				"async.o",
			)
			command, release := boundedrun.Command(45*time.Second, compiler, arguments...)
			command.Dir = scratch
			output, compileError := command.CombinedOutput()
			release()
			if compileError != nil {
				return fmt.Errorf("mutant must compile: %w\n%s", compileError, output)
			}
			archive, readError := os.ReadFile(baseline)
			if readError != nil {
				return readError
			}
			library := filepath.Join(scratch, "runtime.a")
			if writeError := os.WriteFile(library, archive, 0600); writeError != nil {
				return writeError
			}
			staticsCompile(test, archiverName(compiler), "r", library, object)
			if buildWithRuntimeError := BuildWithRuntime(
				asyncCountProtocol,
				filepath.Join(directory, mutant.name),
				options,
				library,
			); buildWithRuntimeError != nil {
				return buildWithRuntimeError
			}
		}
		return nil
	})
}
func TestProduct_AsyncCountProtocol(test *testing.T) {
	test.Parallel()
	asyncCountProtocolProduct(test)
}
func checkAsyncCountProtocol(test *testing.T, mode, mutant string) {
	test.Helper()
	result := asyncCountProtocolRun(test, filepath.Join(asyncCountProtocolProduct(test), mutant), mode)
	if mutant == "control" {
		if result.exit != 0 || string(result.stdout) != "12345\n12345\n" || len(result.stderr) != 0 {
			test.Fatalf("protocol control: %+v", result)
		}
	} else if result.exit == 0 || !strings.Contains(string(result.stderr), "AddressSanitizer: heap-use-after-free") {
		test.Fatalf(
			"required %s owner was not caught by ASan: exit %d stdout %q stderr %s",
			mutant,
			result.exit,
			result.stdout,
			result.stderr,
		)
	}
	test.Logf("%s %s: exit %d", mode, mutant, result.exit)
}
func TestAsyncCountSettledProtocol(test *testing.T) {
	test.Parallel()
	checkAsyncCountProtocol(test, "settled", "control")
}
func TestAsyncCountPendingProtocol(test *testing.T) {
	test.Parallel()
	checkAsyncCountProtocol(test, "pending", "control")
}
func TestAsyncCountQueuedPromiseMutant(test *testing.T) {
	test.Parallel()
	checkAsyncCountProtocol(test, "settled", "promise")
}
func TestAsyncCountSettlementPromiseMutant(test *testing.T) {
	test.Parallel()
	checkAsyncCountProtocol(test, "pending", "promise")
}
func TestAsyncCountReactionFrameMutant(test *testing.T) {
	test.Parallel()
	checkAsyncCountProtocol(test, "pending", "frame")
}
func TestAsyncCountPayloadOwnerRuntimeMutant(test *testing.T) {
	test.Parallel()
	checkAsyncCountProtocol(test, "settled", "payload")
}
func TestAsyncCountPendingOwnerMutant(test *testing.T) {
	test.Parallel()
	checkAsyncCountProtocol(test, "pending", "waiting")
}
