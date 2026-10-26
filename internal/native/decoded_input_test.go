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

func decodedReadProduct(t *testing.T) string {
	t.Helper()
	options := Options{Sanitize: true, Count: true}
	files, flags, compiler, version, err := runtimeInputs("", options, nil)
	if err != nil {
		t.Fatal(err)
	}
	headers, err := headersRead(files, flags, compiler)
	if err != nil {
		t.Fatal(err)
	}
	inputs := buildcache.Inputs{Name: "decoded-input-read", Files: []string{"internal/native/runtime", "internal/native/decoded_input_test.go", "internal/native/native.go", "internal/native/library.go", "internal/native/build_runtime.go", "internal/native/target.go"}, Flags: portableKeyParts(files, flags, compiler, headers), Toolchain: []string{buildcache.Tool(compiler, "--version"), buildcache.Tool(archiverName(compiler), "--version"), buildcache.Tool("node", "--version")}}
	return buildcache.Product(t, inputs, func(directory string) error {
		cache, err := cacheDirectory("decoded-input-read")
		if err != nil {
			return err
		}
		library, err := cachedRuntime(files, flags, compiler, version, cache)
		if err != nil {
			return err
		}
		return BuildWithRuntime(decodedReadProbe, filepath.Join(directory, "probe"), options, library)
	})
}
func TestProduct_DecodedInputReads(t *testing.T) { t.Parallel(); decodedReadProduct(t) }

// Keep the emitted native and JavaScript read paths on the same source fixture.
func decodedBackendProduct(t *testing.T) string {
	t.Helper()
	inputs, err := buildcache.GoInputs("decoded-input-backends", "./cmd/adamic", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	inputs.Name = "decoded-input-backends"
	inputs.Files = append(inputs.Files, "internal/native/decoded_input_test.go", "internal/native/testdata/decoded_input.a")
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"), buildcache.Tool("node", "--version"))
	return buildcache.Product(t, inputs, func(directory string) error {
		entry, err := filepath.Abs("testdata/decoded_input.a")
		if err != nil {
			return err
		}
		program, err := load.Load([]string{entry})
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
		defer cancel()
		lowered, err := lower.Lower(ctx, program)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(directory, "program.mjs"), []byte(javascript.JavaScript(lowered)), 0644); err != nil {
			return err
		}
		return Build(C(lowered), filepath.Join(directory, "program"), Options{Sanitize: true, Count: true, Jobs: 1})
	})
}
func TestProduct_DecodedInputBackends(t *testing.T) { t.Parallel(); decodedBackendProduct(t) }

func checkDecodedInput(t *testing.T, data []byte) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "source.a")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	script := `const fs=require('node:fs');const s=fs.readFileSync(process.argv[1],'utf8');console.log([s.length,...Array.from({length:s.length},(_,i)=>s.charCodeAt(i))].join(' '));`
	node, release := boundedrun.Command(10*time.Second, "node", "-e", script, path)
	truth, err := node.CombinedOutput()
	release()
	if err != nil {
		t.Fatalf("Node: %v %s", err, truth)
	}
	command, release := boundedrun.Command(10*time.Second, filepath.Join(decodedReadProduct(t), "probe"), path)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1")
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err = command.Run()
	release()
	if err != nil || !bytes.Equal(stdout.Bytes(), truth) {
		t.Fatalf("Node %q, native %q, error %v: %s", truth, stdout.Bytes(), err, stderr.String())
	}
	var allocations, frees, retains, releases, peak, regions uint64
	n, err := fmt.Sscanf(stderr.String(), "adamic: counts: allocations %d frees %d retains %d releases %d peak %d regions %d", &allocations, &frees, &retains, &releases, &peak, &regions)
	if err != nil || n != 6 || allocations != frees+regions {
		t.Fatalf("unbalanced read: %s", stderr.String())
	}
	directory := decodedBackendProduct(t)
	for _, backend := range []string{"native", "JavaScript"} {
		var commandName string
		var args []string
		if backend == "native" {
			commandName = filepath.Join(directory, "program")
			args = []string{path}
		} else {
			commandName = "node"
			args = []string{"--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", filepath.Join(directory, "program.mjs"), path}
		}
		run, release := boundedrun.Command(10*time.Second, commandName, args...)
		run.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1")
		stdout.Reset()
		stderr.Reset()
		run.Stdout = &stdout
		run.Stderr = &stderr
		err := run.Run()
		release()
		if err != nil || !bytes.Equal(stdout.Bytes(), truth) {
			t.Fatalf("%s: Node %q, got %q, error %v: %s", backend, truth, stdout.Bytes(), err, stderr.String())
		}
		if backend == "native" {
			n, err := fmt.Sscanf(stderr.String(), "adamic: counts: allocations %d frees %d retains %d releases %d peak %d regions %d", &allocations, &frees, &retains, &releases, &peak, &regions)
			if err != nil || n != 6 || allocations != frees+regions {
				t.Fatalf("unbalanced emitted read: %s", stderr.String())
			}
		}
	}
	t.Logf("Node, native ASan/UBSan and JavaScript agree on %d input bytes; decoded metadata populated before indexing", len(data))
}
func TestDecodedInputASCII(t *testing.T) {
	t.Parallel()
	checkDecodedInput(t, []byte(strings.Repeat("const value = 1;\n", 16)))
}
func TestDecodedInputMalformedWTF8(t *testing.T) {
	t.Parallel()
	data := []byte("é" + strings.Repeat("const value = 1;\n", 16) + "😀")
	data = append(data, 0xed, 0xa0, 0x80, 0xc0, 0xaf, 0xf0, 0x9f)
	checkDecodedInput(t, data)
}

func TestDecodedInputTwoByte(t *testing.T) {
	t.Parallel()
	checkDecodedInput(t, []byte(strings.Repeat("é", 160)))
}
func TestDecodedInputThreeByte(t *testing.T) {
	t.Parallel()
	checkDecodedInput(t, []byte(strings.Repeat("漢", 160)))
}
func TestDecodedInputAstral(t *testing.T) {
	t.Parallel()
	checkDecodedInput(t, []byte(strings.Repeat("😀", 160)))
}
func TestDecodedInputEmpty(t *testing.T) { t.Parallel(); checkDecodedInput(t, nil) }
func TestDecodedInputTruncatedSequences(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		name string
		tail []byte
	}{{"two", []byte{0xc3}}, {"three", []byte{0xe2, 0x82}}, {"four", []byte{0xf0, 0x9f, 0x98}}} {
		t.Run(c.name, func(t *testing.T) { checkDecodedInput(t, append([]byte("é漢😀"), c.tail...)) })
	}
}
