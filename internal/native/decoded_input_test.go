package native

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/boundedrun"
	"github.com/system-inc/adamic/internal/buildcache"
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
	t.Logf("Node and native agree on %d input bytes; allocations %d frees %d", len(data), allocations, frees)
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
