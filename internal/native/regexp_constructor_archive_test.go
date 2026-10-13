package native

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/toolchain"
)

const constructorPreamble = `#define ADAMIC_REGEXP_RUNTIME_COMPILER 1
#include "adamic.h"
#include <stdio.h>
#include <string.h>
static adamic_string *text(const char *bytes,size_t length) {
 adamic_string *value=adamic_string_allocate(length);memcpy((char *)value->bytes,bytes,length);return value;
}
static void show(adamic_string *value) {
 if(value==NULL){puts("undefined");return;}
 for(size_t i=0;i<value->length;i++)printf("%02x",(unsigned char)value->bytes[i]);puts("");
}
static void item(adamic_array *array,size_t i) {show(array==NULL?NULL:array->elements[i].reference);}
`

// Expected output is produced by live Node from the same counted patterns.
// Native and wasm exercise the public constructor, not the parser test seam.
func constructorArchiveNode(t *testing.T, options Options, family string) {
	t.Helper()
	var c, js strings.Builder
	if family == "splitSource" {
		// Split linking needs a cross-unit call, independently of the larger
		// ownership witness. Its arguments are counted and released by make.
		c.WriteString(`#define ADAMIC_REGEXP_RUNTIME_COMPILER 1
#include "adamic.h"
#include <stdio.h>
#include <string.h>
static adamic_object *make(void) {
 adamic_string *p=adamic_string_allocate(3),*f=adamic_string_allocate(2);
 memcpy((char *)p->bytes,"a/b",3);memcpy((char *)f->bytes,"ig",2);
 adamic_object *r=adamic_regex_compile_new(p,f);adamic_release(p);adamic_release(f);return r;
}
`)
	} else {
		c.WriteString(constructorPreamble)
	}
	c.WriteString("int main(int argc,char **argv){adamic_start(argc,argv);\n")
	js.WriteString(`const show=s=>console.log(s===undefined?'undefined':Buffer.from(s).toString('hex'));`)
	switch family {
	case "splitSource":
		c.WriteString(`adamic_object *r=make();if(r==NULL)adamic_uncaught();
 adamic_string *source=r->slots[2].reference,*flags=r->slots[3].reference;
 fwrite(source->bytes,1,source->length,stdout);puts("");fwrite(flags->bytes,1,flags->length,stdout);puts("");adamic_release(r);
`)
		js.WriteString(`{const r=new RegExp('a/b','ig');console.log(r.source);console.log(r.flags);}`)
	case "source", "flags":
		for index, row := range [][2]string{{"", ""}, {"a/b\n\r\u2028\u2029", "mi"}, {"[/]", "u"}, {"a\x00b", "u"}, {"(?<x>a)", "ysumigd"}, {"[\\q{ab}]", "v"}} {
			if (family == "source" && index >= 4) || (family == "flags" && index < 4) {
				continue
			}
			fmt.Fprintf(&c, "{adamic_string *p=text(%s,%d),*f=text(%s,%d);adamic_object *r=adamic_regex_compile_new(p,f);adamic_release(p);adamic_release(f);if(r==NULL)adamic_uncaught();show(r->slots[2].reference);show(r->slots[3].reference);adamic_release(r);}\n", cString(row[0]), len(row[0]), cString(row[1]), len(row[1]))
			fmt.Fprintf(&js, "{const r=new RegExp(%s,%s);show(r.source);show(r.flags);}", jsQuote(row[0]), jsQuote(row[1]))
		}
		c.WriteString("{adamic_object *r=adamic_regex_compile_new(NULL,NULL);show(r->slots[2].reference);show(r->slots[3].reference);adamic_release(r);}\n")
		js.WriteString("{const r=new RegExp(undefined,undefined);show(r.source);show(r.flags);}")
	case "syntaxErrors", "syntaxBracket", "syntaxEscape", "syntaxGroup", "flagErrors":
		for index, row := range [][2]string{{"[", "u"}, {"\\", ""}, {"(?<x>", "u"}, {"a", "gg"}, {"a", "uv"}} {
			if (strings.HasPrefix(family, "syntax") && index >= 3) || (family == "flagErrors" && index < 3) {
				continue
			}
			// Each wasm error witness is a separate Loom unit, including its build.
			if (family == "syntaxBracket" && index != 0) || (family == "syntaxEscape" && index != 1) || (family == "syntaxGroup" && index != 2) {
				continue
			}
			fmt.Fprintf(&c, "{adamic_string *p=text(%s,%d),*f=text(%s,%d);adamic_object *r=adamic_regex_compile_new(p,f);adamic_release(p);adamic_release(f);if(r!=NULL){puts(\"accepted\");adamic_release(r);}else{show(adamic_thrown->slots[0].reference);show(adamic_thrown->slots[1].reference);adamic_release(adamic_thrown);adamic_thrown=NULL;}}\n", cString(row[0]), len(row[0]), cString(row[1]), len(row[1]))
			fmt.Fprintf(&js, "try{new RegExp(%s,%s);console.log('accepted');}catch(e){show(e.name);show(e.message);}", jsQuote(row[0]), jsQuote(row[1]))
		}
	case "ownership":
		c.WriteString(`{
 adamic_string *p=text("(?<x>a)",7),*f=text("gdu",3),*input=text("aba",3);
 adamic_object *r=adamic_regex_compile_new(p,f);adamic_release(p);adamic_release(f);if(r==NULL)adamic_uncaught();
 r->slots[1].number=2;
 adamic_object *iterator=adamic_regex_match_all(input,r);adamic_release(r);
 adamic_array *m=adamic_regex_iterator_step(iterator);
 item(m,0);adamic_object *groups=adamic_retain(m->properties->slots[2].reference);
 adamic_array *indices=m->properties->slots[3].reference;
 adamic_object *index_groups=adamic_retain(indices->properties->slots[2].reference);
 adamic_release(m);adamic_release(iterator);
 show(adamic_regex_group_lookup(groups,"x",false));puts(adamic_regex_group_lookup(groups,"#compiler",false)==NULL?"undefined":"visible");
 adamic_object *pair=adamic_regex_group_lookup(index_groups,"x",false);
 printf("%.0f %.0f\n",pair->slots[0].number,pair->slots[1].number);
 adamic_release(groups);adamic_release(index_groups);
 p=text("(a)",3);f=text("u",1);r=adamic_regex_compile_new(p,f);adamic_release(p);adamic_release(f);
 adamic_array *pieces=adamic_regex_split(input,r,4294967295);adamic_release(r);
 for(size_t i=0;i<pieces->length;i++)item(pieces,i);adamic_release(pieces);adamic_release(input);
 }
`)
		js.WriteString(`{const r=new RegExp('(?<x>a)','gdu');r.lastIndex=2;const iterator='aba'.matchAll(r);const m=iterator.next().value;show(m[0]);show(m.groups.x);show(m.groups['#compiler']);console.log(m.indices.groups.x.join(' '));for(const s of 'aba'.split(new RegExp('(a)','u')))show(s);}`)
	default:
		t.Fatal("unknown family")
	}
	c.WriteString("return 0;}\n")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	want, err := exec.CommandContext(ctx, "node", "-e", js.String()).CombinedOutput()
	if err != nil {
		t.Fatalf("Node: %v %s", err, want)
	}
	if directory := os.Getenv("ADAMIC_REGEX_CONSTRUCTOR_ARTIFACTS"); directory != "" {
		if err := os.WriteFile(filepath.Join(directory, family+".c"), []byte(c.String()), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, family+".node"), want, 0600); err != nil {
			t.Fatal(err)
		}
	}
	binary := filepath.Join(t.TempDir(), "constructor")
	if err := Build(c.String(), binary, options); err != nil {
		t.Fatal(err)
	}
	runContext, runCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer runCancel()
	command := exec.CommandContext(runContext, binary)
	if options.Target == "wasm32-wasi" {
		command = exec.CommandContext(runContext, "node", "--disable-warning=ExperimentalWarning", "../../oracle/wasi.mjs", binary)
	}
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1")
	got, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("constructor: %v %s", err, got)
	}
	if !bytes.Equal(want, got) {
		t.Fatalf("Node disagreement (%s):\nNode: %s\nnative: %s", family, want, got)
	}
}

// A JSON string is also a JavaScript string literal. NUL stays length-delimited.
func jsQuote(value string) string { return fmt.Sprintf("%q", value) }

func TestRegExpConstructorSourceArchive(t *testing.T) {
	t.Parallel()
	constructorArchiveNode(t, Options{Sanitize: true}, "source")
}
func TestRegExpConstructorSourceArchiveWASM(t *testing.T) {
	t.Parallel()
	constructorArchiveNode(t, Options{Target: "wasm32-wasi"}, "source")
}
func TestRegExpConstructorSyntaxErrorsArchive(t *testing.T) {
	t.Parallel()
	constructorArchiveNode(t, Options{Sanitize: true}, "syntaxErrors")
}
func TestRegExpConstructorSyntaxBracketArchiveWASM(t *testing.T) {
	t.Parallel()
	constructorArchiveNode(t, Options{Target: "wasm32-wasi"}, "syntaxBracket")
}
func TestRegExpConstructorSyntaxEscapeArchiveWASM(t *testing.T) {
	t.Parallel()
	constructorArchiveNode(t, Options{Target: "wasm32-wasi"}, "syntaxEscape")
}
func TestRegExpConstructorSyntaxGroupArchiveWASM(t *testing.T) {
	t.Parallel()
	constructorArchiveNode(t, Options{Target: "wasm32-wasi"}, "syntaxGroup")
}
func TestRegExpConstructorOwnershipArchive(t *testing.T) {
	t.Parallel()
	constructorArchiveNode(t, Options{Sanitize: true}, "ownership")
}
func TestRegExpConstructorOwnershipArchiveWASM(t *testing.T) {
	t.Parallel()
	constructorArchiveNode(t, Options{Target: "wasm32-wasi"}, "ownership")
}
func TestRegExpConstructorSplitSourceArchive(t *testing.T) {
	t.Parallel()
	constructorArchiveNode(t, Options{Split: true, Jobs: 2, Sanitize: true}, "splitSource")
}
func TestRegExpConstructorSplitSourceArchiveWASM(t *testing.T) {
	t.Parallel()
	constructorArchiveNode(t, Options{Target: "wasm32-wasi"}, "splitSource")
}

func TestRegExpConstructorFlagsArchive(t *testing.T) {
	t.Parallel()
	constructorArchiveNode(t, Options{Sanitize: true}, "flags")
}

func compilerArchiveCache(t *testing.T, options Options) {
	t.Helper()
	source := "#define ADAMIC_REGEXP_RUNTIME_COMPILER 1\n"
	library, err := regexpCompilerLibrary(source, options)
	if err != nil {
		t.Fatal(err)
	}
	reused, err := regexpCompilerLibrary(source, options)
	if err != nil || reused != library {
		t.Fatalf("archive reuse: %s %s %v", library, reused, err)
	}
	compiler, err := toolchain.Path(compilerName(options))
	if err != nil {
		t.Fatal(err)
	}
	output, err := toolchain.Command(archiverName(compiler), "t", library).CombinedOutput()
	if err != nil {
		t.Fatalf("archive members: %v %s", err, output)
	}
	members := strings.Fields(string(output))
	if len(members) != 7 {
		t.Fatalf("compiler archive: %v", members)
	}
	for _, member := range members {
		if !strings.HasPrefix(member, "regexp_compile_") {
			t.Fatalf("ordinary unit entered compiler archive: %s", member)
		}
	}
	t.Log("compiler cache reuses seven compiler-only units")
}

func TestRegExpConstructorUsesOrdinaryRuntime(t *testing.T) {
	t.Parallel()
	source := "#define ADAMIC_REGEXP_RUNTIME_COMPILER 1\n"
	ordinary, err := RuntimeLibraryForSource("", source, Options{Sanitize: true})
	if err != nil {
		t.Fatal(err)
	}
	plain, err := RuntimeLibrary("", Options{Sanitize: true})
	if err != nil || ordinary != plain {
		t.Fatalf("construction needs a second ordinary runtime: %s %s %v", ordinary, plain, err)
	}
	t.Log("construction shares the ordinary runtime")
}

func TestRegExpConstructorFlagsArchiveWASM(t *testing.T) {
	t.Parallel()
	constructorArchiveNode(t, Options{Target: "wasm32-wasi"}, "flags")
}

func TestRegExpCompilerArchiveCache(t *testing.T) {
	t.Parallel()
	compilerArchiveCache(t, Options{Sanitize: true})
}
func TestRegExpCompilerArchiveCacheWASM(t *testing.T) {
	t.Parallel()
	compilerArchiveCache(t, Options{Target: "wasm32-wasi"})
}
func TestRegExpConstructorFlagErrorsArchive(t *testing.T) {
	t.Parallel()
	constructorArchiveNode(t, Options{Sanitize: true}, "flagErrors")
}
func TestRegExpConstructorFlagErrorsArchiveWASM(t *testing.T) {
	t.Parallel()
	constructorArchiveNode(t, Options{Target: "wasm32-wasi"}, "flagErrors")
}
