package native

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/boundedrun"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

const decodedReadProbe = `#include "adamic.h"
#include <assert.h>
#include <stdio.h>
#include <string.h>
int main(int argc, char **argv) {
 assert(argc==2); adamic_start(argc,argv);
 adamic_string *path=adamic_decode_utf8((const unsigned char *)argv[1],strlen(argv[1]));
 adamic_object *result=adamic_read_text_file(path);
 adamic_slot_cache kind_cache={0}, text_cache={0};
 adamic_string *kind=adamic_object_field(result,"kind",&kind_cache)->reference;
 assert(kind->length==2 && memcmp(kind->bytes,"Ok",2)==0);
 adamic_string *text=adamic_object_field(result,"text",&text_cache)->reference;
 // The count must exist before the first public character read, including
 // a read that builds the long string's UTF-16 index and view.
 assert(text->units!=0);
 size_t units=text->units-1;
 printf("%zu",units);
 for(size_t i=0;i<units;i++) printf(" %.0f",adamic_string_char_code_at(text,(double)i));
 puts(""); adamic_release(result); adamic_release(path); return 0;
}
`

func decodedReadProduct(test *testing.T) string {
	test.Helper()
	options := Options{Sanitize: true, Count: true}
	files, flags, compiler, version, runtimeInputsError := runtimeInputs("", options, nil)
	if runtimeInputsError != nil {
		test.Fatal(runtimeInputsError)
	}
	headers, headersReadError := headersRead(files, flags, compiler)
	if headersReadError != nil {
		test.Fatal(headersReadError)
	}
	inputs := buildcache.Inputs{
		Name: "decoded-input-read",
		Files: []string{
			"internal/native/runtime",
			"internal/native/decoded_input_test.go",
			"internal/native/native.go",
			"internal/native/library.go",
			"internal/native/build_runtime.go",
			"internal/native/target.go",
		},
		Flags: portableKeyParts(files, flags, compiler, headers),
		Toolchain: []string{
			buildcache.Tool(compiler, "--version"),
			buildcache.Tool(archiverName(compiler), "--version"),
			buildcache.Tool("node", "--version"),
		},
	}
	return buildcache.Product(test, inputs, func(directory string) error {
		cache, cacheDirectoryError := cacheDirectory("decoded-input-read")
		if cacheDirectoryError != nil {
			return cacheDirectoryError
		}
		library, cachedRuntimeError := cachedRuntime(files, flags, compiler, version, cache)
		if cachedRuntimeError != nil {
			return cachedRuntimeError
		}
		return BuildWithRuntime(decodedReadProbe, filepath.Join(directory, "probe"), options, library)
	})
}
func TestProduct_DecodedInputReads(test *testing.T) { test.Parallel(); decodedReadProduct(test) }

// Keep the emitted native and JavaScript read paths on the same source fixture.
func decodedBackendProduct(test *testing.T) string {
	test.Helper()
	inputs, goInputsError := buildcache.GoInputs("decoded-input-backends", "./cmd/adamic", nil, nil)
	if goInputsError != nil {
		test.Fatal(goInputsError)
	}
	inputs.Name = "decoded-input-backends"
	inputs.Files = append(
		inputs.Files,
		"internal/native/decoded_input_test.go",
		"internal/native/testdata/decoded_input.a",
	)
	inputs.Toolchain = append(
		inputs.Toolchain,
		buildcache.Tool("clang", "--version"),
		buildcache.Tool("node", "--version"),
	)
	return buildcache.Product(test, inputs, func(directory string) error {
		entry, absolutePathError := filepath.Abs("testdata/decoded_input.a")
		if absolutePathError != nil {
			return absolutePathError
		}
		program, loadError := load.Load([]string{entry})
		if loadError != nil {
			return loadError
		}
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
		defer cancel()
		lowered, lowerError := lower.Lower(ctx, program)
		if lowerError != nil {
			return lowerError
		}
		if writeError := os.WriteFile(
			filepath.Join(directory, "program.mjs"),
			[]byte(javascript.JavaScript(lowered)),
			0644,
		); writeError != nil {
			return writeError
		}
		return Build(C(lowered), filepath.Join(directory, "program"), Options{Sanitize: true, Count: true, Jobs: 1})
	})
}
func TestProduct_DecodedInputBackends(test *testing.T) { test.Parallel(); decodedBackendProduct(test) }

func checkDecodedInput(test *testing.T, data []byte) {
	test.Helper()
	path := filepath.Join(test.TempDir(), "source.a")
	if writeError := os.WriteFile(path, data, 0600); writeError != nil {
		test.Fatal(writeError)
	}
	script := `const fs=require('node:fs');const s=fs.readFileSync(process.argv[1],'utf8');console.log([s.length,...Array.from({length:s.length},(_,i)=>s.charCodeAt(i))].join(' '));`
	node, release := boundedrun.Command(10*time.Second, "node", "-e", script, path)
	truth, combinedOutputError := node.CombinedOutput()
	release()
	if combinedOutputError != nil {
		test.Fatalf("Node: %v %s", combinedOutputError, truth)
	}
	command, release := boundedrun.Command(10*time.Second, filepath.Join(decodedReadProduct(test), "probe"), path)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1")
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	runError := command.Run()
	release()
	if runError != nil || !bytes.Equal(stdout.Bytes(), truth) {
		test.Fatalf("Node %q, native %q, error %v: %s", truth, stdout.Bytes(), runError, stderr.String())
	}
	var allocations, frees, retains, releases, peak, regions uint64
	n, scanError := fmt.Sscanf(
		stderr.String(),
		"adamic: counts: allocations %d frees %d retains %d releases %d peak %d regions %d",
		&allocations,
		&frees,
		&retains,
		&releases,
		&peak,
		&regions,
	)
	if scanError != nil || n != 6 || allocations != frees+regions {
		test.Fatalf("unbalanced read: %s", stderr.String())
	}
	directory := decodedBackendProduct(test)
	for _, backend := range []string{"native", "JavaScript"} {
		var commandName string
		var args []string
		if backend == "native" {
			commandName = filepath.Join(directory, "program")
			args = []string{path}
		} else {
			commandName = "node"
			args = []string{
				"--disable-warning=ExperimentalWarning",
				"../../oracle/node.mjs",
				filepath.Join(directory, "program.mjs"),
				path,
			}
		}
		run, release := boundedrun.Command(10*time.Second, commandName, args...)
		run.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1")
		stdout.Reset()
		stderr.Reset()
		run.Stdout = &stdout
		run.Stderr = &stderr
		runError := run.Run()
		release()
		if runError != nil || !bytes.Equal(stdout.Bytes(), truth) {
			test.Fatalf("%s: Node %q, got %q, error %v: %s", backend, truth, stdout.Bytes(), runError, stderr.String())
		}
		if backend == "native" {
			n, scanError := fmt.Sscanf(
				stderr.String(),
				"adamic: counts: allocations %d frees %d retains %d releases %d peak %d regions %d",
				&allocations,
				&frees,
				&retains,
				&releases,
				&peak,
				&regions,
			)
			if scanError != nil || n != 6 || allocations != frees+regions {
				test.Fatalf("unbalanced emitted read: %s", stderr.String())
			}
		}
	}
	test.Logf(
		"Node, native ASan/UBSan and JavaScript agree on %d input bytes; decoded metadata populated before indexing",
		len(data),
	)
}
func TestDecodedInputASCII(test *testing.T) {
	test.Parallel()
	checkDecodedInput(test, []byte(strings.Repeat("const value = 1;\n", 16)))
}
func TestDecodedInputMalformedWTF8(test *testing.T) {
	test.Parallel()
	data := []byte("é" + strings.Repeat("const value = 1;\n", 16) + "😀")
	data = append(data, 0xed, 0xa0, 0x80, 0xc0, 0xaf, 0xf0, 0x9f)
	checkDecodedInput(test, data)
}

func TestDecodedInputTwoByte(test *testing.T) {
	test.Parallel()
	checkDecodedInput(test, []byte(strings.Repeat("é", 160)))
}
func TestDecodedInputThreeByte(test *testing.T) {
	test.Parallel()
	checkDecodedInput(test, []byte(strings.Repeat("漢", 160)))
}
func TestDecodedInputAstral(test *testing.T) {
	test.Parallel()
	checkDecodedInput(test, []byte(strings.Repeat("😀", 160)))
}
func TestDecodedInputEmpty(test *testing.T) { test.Parallel(); checkDecodedInput(test, nil) }
func TestDecodedInputTruncatedSequences(test *testing.T) {
	test.Parallel()
	for _, c := range []struct {
		name string
		tail []byte
	}{{"two", []byte{0xc3}}, {"three", []byte{0xe2, 0x82}}, {"four", []byte{0xf0, 0x9f, 0x98}}} {
		test.Run(c.name, func(test *testing.T) { checkDecodedInput(test, append([]byte("é漢😀"), c.tail...)) })
	}
}
