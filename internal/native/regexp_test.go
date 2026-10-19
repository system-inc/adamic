package native

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	regex "github.com/system-inc/adamic/internal/regexp"
	"hash/fnv"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
	"unicode/utf16"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/toolchain"
)

type regexCase struct {
	Pattern      string   `json:"pattern"`
	PatternUnits []uint16 `json:"patternUnits"`
	Flags        string   `json:"flags"`
	Input        []uint16 `json:"input"`
	LastIndex    uint64   `json:"lastIndex"`
	Expected     struct {
		Captures  [][]int          `json:"captures"`
		LastIndex uint64           `json:"lastIndex"`
		Groups    map[string][]int `json:"groups"`
	} `json:"expected"`
}

// The observed test262 calls run through the C interpreter, including captures, named groups,
// indices and lastIndex, under both sanitizers. Building every call into one executable is more than
// one test should carry alone, and nearly all of them are one pattern's, so they're cut by each
// call's content, not by pattern: a call belongs to bucket b of n when the FNV-1a hash of its
// pattern, flags, input and lastIndex is b modulo n. Each bucket is its own executable and its own
// top-level test below, and TestRegExpTest262BucketsEachHaveOneTest holds the buckets to their tests.
var regExpTest262Buckets = map[string]int{
	"bucket 0 of 32":  0,
	"bucket 1 of 32":  1,
	"bucket 2 of 32":  2,
	"bucket 3 of 32":  3,
	"bucket 4 of 32":  4,
	"bucket 5 of 32":  5,
	"bucket 6 of 32":  6,
	"bucket 7 of 32":  7,
	"bucket 8 of 32":  8,
	"bucket 9 of 32":  9,
	"bucket 10 of 32": 10,
	"bucket 11 of 32": 11,
	"bucket 12 of 32": 12,
	"bucket 13 of 32": 13,
	"bucket 14 of 32": 14,
	"bucket 15 of 32": 15,
	"bucket 16 of 32": 16,
	"bucket 17 of 32": 17,
	"bucket 18 of 32": 18,
	"bucket 19 of 32": 19,
	"bucket 20 of 32": 20,
	"bucket 21 of 32": 21,
	"bucket 22 of 32": 22,
	"bucket 23 of 32": 23,
	"bucket 24 of 32": 24,
	"bucket 25 of 32": 25,
	"bucket 26 of 32": 26,
	"bucket 27 of 32": 27,
	"bucket 28 of 32": 28,
	"bucket 29 of 32": 29,
	"bucket 30 of 32": 30,
	"bucket 31 of 32": 31,
}

func TestRegExpBytecodeTest262Bucket0Of32(t *testing.T) {
	t.Parallel()
	runRegExpTest262Bucket(t, "bucket 0 of 32")
}

func TestRegExpBytecodeTest262Bucket1Of32(t *testing.T) {
	t.Parallel()
	runRegExpTest262Bucket(t, "bucket 1 of 32")
}

func TestRegExpBytecodeTest262Bucket2Of32(t *testing.T) {
	t.Parallel()
	runRegExpTest262Bucket(t, "bucket 2 of 32")
}

func TestRegExpBytecodeTest262Bucket3Of32(t *testing.T) {
	t.Parallel()
	runRegExpTest262Bucket(t, "bucket 3 of 32")
}

func TestRegExpBytecodeTest262Bucket4Of32(t *testing.T) {
	t.Parallel()
	runRegExpTest262Bucket(t, "bucket 4 of 32")
}

func TestRegExpBytecodeTest262Bucket5Of32(t *testing.T) {
	t.Parallel()
	runRegExpTest262Bucket(t, "bucket 5 of 32")
}

func TestRegExpBytecodeTest262Bucket6Of32(t *testing.T) {
	t.Parallel()
	runRegExpTest262Bucket(t, "bucket 6 of 32")
}

func TestRegExpBytecodeTest262Bucket7Of32(t *testing.T) {
	t.Parallel()
	runRegExpTest262Bucket(t, "bucket 7 of 32")
}

func TestRegExpBytecodeTest262Bucket8Of32(t *testing.T) {
	t.Parallel()
	runRegExpTest262Bucket(t, "bucket 8 of 32")
}

func TestRegExpBytecodeTest262Bucket9Of32(t *testing.T) {
	t.Parallel()
	runRegExpTest262Bucket(t, "bucket 9 of 32")
}

func TestRegExpBytecodeTest262Bucket10Of32(t *testing.T) {
	t.Parallel()
	runRegExpTest262Bucket(t, "bucket 10 of 32")
}

func TestRegExpBytecodeTest262Bucket11Of32(t *testing.T) {
	t.Parallel()
	runRegExpTest262Bucket(t, "bucket 11 of 32")
}

func TestRegExpBytecodeTest262Bucket12Of32(t *testing.T) {
	t.Parallel()
	runRegExpTest262Bucket(t, "bucket 12 of 32")
}

func TestRegExpBytecodeTest262Bucket13Of32(t *testing.T) {
	t.Parallel()
	runRegExpTest262Bucket(t, "bucket 13 of 32")
}

func TestRegExpBytecodeTest262Bucket14Of32(t *testing.T) {
	t.Parallel()
	runRegExpTest262Bucket(t, "bucket 14 of 32")
}

func TestRegExpBytecodeTest262Bucket15Of32(t *testing.T) {
	t.Parallel()
	runRegExpTest262Bucket(t, "bucket 15 of 32")
}

func TestRegExpBytecodeTest262Bucket16Of32(t *testing.T) {
	t.Parallel()
	runRegExpTest262Bucket(t, "bucket 16 of 32")
}

func TestRegExpBytecodeTest262Bucket17Of32(t *testing.T) {
	t.Parallel()
	runRegExpTest262Bucket(t, "bucket 17 of 32")
}

func TestRegExpBytecodeTest262Bucket18Of32(t *testing.T) {
	t.Parallel()
	runRegExpTest262Bucket(t, "bucket 18 of 32")
}

func TestRegExpBytecodeTest262Bucket19Of32(t *testing.T) {
	t.Parallel()
	runRegExpTest262Bucket(t, "bucket 19 of 32")
}

func TestRegExpBytecodeTest262Bucket20Of32(t *testing.T) {
	t.Parallel()
	runRegExpTest262Bucket(t, "bucket 20 of 32")
}

func TestRegExpBytecodeTest262Bucket21Of32(t *testing.T) {
	t.Parallel()
	runRegExpTest262Bucket(t, "bucket 21 of 32")
}

func TestRegExpBytecodeTest262Bucket22Of32(t *testing.T) {
	t.Parallel()
	runRegExpTest262Bucket(t, "bucket 22 of 32")
}

func TestRegExpBytecodeTest262Bucket23Of32(t *testing.T) {
	t.Parallel()
	runRegExpTest262Bucket(t, "bucket 23 of 32")
}

func TestRegExpBytecodeTest262Bucket24Of32(t *testing.T) {
	t.Parallel()
	runRegExpTest262Bucket(t, "bucket 24 of 32")
}

func TestRegExpBytecodeTest262Bucket25Of32(t *testing.T) {
	t.Parallel()
	runRegExpTest262Bucket(t, "bucket 25 of 32")
}

func TestRegExpBytecodeTest262Bucket26Of32(t *testing.T) {
	t.Parallel()
	runRegExpTest262Bucket(t, "bucket 26 of 32")
}

func TestRegExpBytecodeTest262Bucket27Of32(t *testing.T) {
	t.Parallel()
	runRegExpTest262Bucket(t, "bucket 27 of 32")
}

func TestRegExpBytecodeTest262Bucket28Of32(t *testing.T) {
	t.Parallel()
	runRegExpTest262Bucket(t, "bucket 28 of 32")
}

func TestRegExpBytecodeTest262Bucket29Of32(t *testing.T) {
	t.Parallel()
	runRegExpTest262Bucket(t, "bucket 29 of 32")
}

func TestRegExpBytecodeTest262Bucket30Of32(t *testing.T) {
	t.Parallel()
	runRegExpTest262Bucket(t, "bucket 30 of 32")
}

func TestRegExpBytecodeTest262Bucket31Of32(t *testing.T) {
	t.Parallel()
	runRegExpTest262Bucket(t, "bucket 31 of 32")
}

// Every call lands in exactly one bucket when the buckets are 0 to n-1 of the same n, so a bucket
// missing, a bucket named for another n, a bucket without its test, or a test naming a bucket that
// isn't there fails here without building anything.
func TestRegExpTest262BucketsEachHaveOneTest(t *testing.T) {
	t.Parallel()
	seen := make([]bool, len(regExpTest262Buckets))
	for name, bucket := range regExpTest262Buckets {
		if want := fmt.Sprintf("bucket %d of %d", bucket, len(regExpTest262Buckets)); name != want || bucket < 0 || bucket >= len(seen) || seen[bucket] {
			t.Errorf("bucket %q must be named %q, one of 0 to %d, and the only one", name, want, len(seen)-1)
			continue
		}
		seen[bucket] = true
	}
	checkNamedWrappers(t, "regexp_test.go", "TestRegExpBytecodeTest262", "runRegExpTest262Bucket", regExpTest262Buckets)
}

func runRegExpTest262Bucket(t *testing.T, name string) {
	t.Helper()
	bucket, ok := regExpTest262Buckets[name]
	if !ok {
		t.Fatalf("no test262 bucket %q", name)
	}
	file, err := os.Open("../regexp/testdata/matches.json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	compressed, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer compressed.Close()
	var cases []regexCase
	if err = json.NewDecoder(compressed).Decode(&cases); err != nil {
		t.Fatal(err)
	}
	var owned []regexCase
	for _, c := range cases {
		content := fnv.New32a()
		fmt.Fprintf(content, "%q %v %q %v %d", c.Pattern, c.PatternUnits, c.Flags, c.Input, c.LastIndex)
		if int(content.Sum32()%uint32(len(regExpTest262Buckets))) == bucket {
			owned = append(owned, c)
		}
	}
	if len(owned) == 0 {
		t.Fatalf("%s holds none of the %d calls", name, len(cases))
	}
	t.Logf("%s holds %d of the %d calls", name, len(owned), len(cases))
	runRegexCases(t, owned)
}

// The C witnesses use regex, strings and counted containers. Their archive
// keeps those dependencies, rather than compiling unrelated runtime units for
// every cold Loom unit. Missing dependencies fail at the link, never silently.
// These members are the transitive symbol closure of adamic, regexp and
// from_codes in the ordinary archive (nm -g); their C sources are unchanged.
var regexWitnessRuntimeSources = map[string]bool{
	"adamic.c": true, "array.c": true, "async.c": true, "bitwise.c": true,
	"class_features.c": true, "class_inheritance.c": true, "class_static.c": true,
	"closure.c": true, "dtoa.c": true, "exceptions.c": true, "from_codes.c": true,
	"heap.c": true, "host_runtime.c": true, "input.c": true, "map.c": true, "map_set.c": true,
	"maybe.c": true, "node_process.c": true, "number.c": true, "object.c": true, "parallel.c": true, "record.c": true,
	"regexp.c": true, "regexp_regular.c": true, "region.c": true, "share.c": true, "sort.c": true, "stack.c": true, "string.c": true,
	"string_index.c": true, "string_share.c": true, "union.c": true,
	"view_unions_mixed.c": true, "weak.c": true,
}

// regexWitnessRuntime is the shared recipe used by the product units and
// consumers. Portable holds the source, header, platform and option inputs;
// tool reports key the compiler and archiver without installation paths.
func regexWitnessRuntime(t testing.TB, request RuntimeRequest, part int) string {
	t.Helper()
	files := []string{"internal/native/regexp_test.go", "internal/native/library.go", "internal/native/native.go", "internal/native/target.go"}
	for _, file := range request.files {
		files = append(files, "internal/native/runtime/"+file.name)
	}
	return buildcache.Product(t, buildcache.Inputs{
		Name: fmt.Sprintf("regexp-bytecode-witness-runtime-%d", part), Files: files, Flags: request.Portable,
		Toolchain: []string{buildcache.Tool(request.Compiler, "--version"), buildcache.Tool(request.Archiver, "--version")},
	}, func(directory string) error { return request.Build(directory, 4, nil) })
}

func regexWitnessRequest(t testing.TB, options Options, part int) RuntimeRequest {
	t.Helper()
	files, flags, compiler, version, err := runtimeInputs("", options, nil)
	if err != nil {
		t.Fatal(err)
	}
	// Both products snapshot all headers, but their C members are disjoint.
	// Splitting dependency discovery too keeps its traced cost in each unit.
	kept := files[:0]
	index := 0
	for _, file := range files {
		if strings.HasSuffix(file.name, ".h") {
			kept = append(kept, file)
			continue
		}
		if regexWitnessRuntimeSources[file.name] {
			if index%2 == part {
				kept = append(kept, file)
			}
			index++
		}
	}
	request, err := runtimeRequest(kept, flags, compiler, version, options)
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func TestProduct_RegExpBytecodeNativePart0(t *testing.T) {
	t.Parallel()
	regexWitnessRuntime(t, regexWitnessRequest(t, Options{Sanitize: true}, 0), 0)
}

func TestProduct_RegExpBytecodeNativePart1(t *testing.T) {
	t.Parallel()
	regexWitnessRuntime(t, regexWitnessRequest(t, Options{Sanitize: true}, 1), 1)
}

func buildRegexWitness(t *testing.T, source, binary string, options Options) {
	t.Helper()
	if features := featureFlags(source); len(features) != 0 {
		t.Fatalf("regex witness has undeclared runtime features: %v", features)
	}
	request := regexWitnessRequest(t, options, 0)
	flags, compiler := request.flags, request.Compiler
	runtime0 := regexWitnessRuntime(t, request, 0)
	runtime1 := regexWitnessRuntime(t, regexWitnessRequest(t, options, 1), 1)
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "main.c"), []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(runtime0, filepath.Join(directory, "runtime")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(runtime1, filepath.Join(directory, "runtime-part1")); err != nil {
		t.Fatal(err)
	}
	arguments := append(flags, prefixMaps(directory, "/adamic-build")...)
	arguments = append(arguments, "-I", "runtime", "main.c", "-o", binary)
	archive := filepath.Join("runtime", "runtime.a")
	arguments = append(arguments, RuntimeLinkFlags(archive)...)
	arguments = append(arguments, RuntimeLinkFlags(filepath.Join("runtime-part1", "runtime.a"))...)
	arguments = append(arguments, "-lm")
	output, err := toolchain.CommandIn(directory, compiler, arguments...).CombinedOutput()
	if err != nil {
		t.Fatalf("regex witness: %v\n%s", err, output)
	}
}

func runRegexCases(t *testing.T, cases []regexCase) {
	t.Helper()
	runRegexCasesForTarget(t, cases, false)
}
func runRegexCasesForTarget(t *testing.T, cases []regexCase, wasmOnly bool) {
	t.Helper()
	cases = regexCompatibleCases(t, cases)
	var source, rows, units, spans strings.Builder
	source.WriteString("#include \"adamic.h\"\n#include <stdio.h>\n#include <stdlib.h>\n#include <string.h>\n")
	programs := map[string]int{}
	inputOffset, spanOffset := 0, 0
	for index, c := range cases {
		flags := c.Flags
		if !strings.Contains(flags, "d") {
			flags += "d"
		}
		key := "pattern:" + c.Pattern + "/" + flags
		if c.PatternUnits != nil {
			key = "units:" + fmt.Sprint(c.PatternUnits) + "/" + flags
		}
		id, ok := programs[key]
		if !ok {
			id = len(programs)
			programs[key] = id
			var p *regex.Program
			var err error
			if c.PatternUnits != nil {
				p, err = regex.CompileUTF16(c.PatternUnits, flags)
			} else {
				p, err = regex.Compile(c.Pattern, flags)
			}
			if err != nil {
				t.Fatalf("case %d compile: %v", index, err)
			}
			text, err := p.NativeDeclarations(fmt.Sprintf("regex_%d", id))
			if err != nil {
				t.Fatalf("case %d encode: %v", index, err)
			}
			source.WriteString(text)
		}
		fmt.Fprintf(&rows, "{&regex_%d,%d,%d,UINT64_C(%d),%d,%d,UINT64_C(%d)},\n", id, inputOffset, len(c.Input), c.LastIndex, spanOffset, len(c.Expected.Captures), c.Expected.LastIndex)
		for _, unit := range c.Input {
			fmt.Fprintf(&units, "%d,", unit)
		}
		inputOffset += len(c.Input)
		for _, capture := range c.Expected.Captures {
			if capture == nil {
				spans.WriteString("-1,-1,")
			} else {
				fmt.Fprintf(&spans, "%d,%d,", capture[0], capture[1])
			}
			spanOffset += 2
		}
		names := make([]string, 0, len(c.Expected.Groups))
		for name := range c.Expected.Groups {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			capture := c.Expected.Groups[name]
			if capture == nil {
				spans.WriteString("-1,-1,")
			} else {
				fmt.Fprintf(&spans, "%d,%d,", capture[0], capture[1])
			}
			spanOffset += 2
		}
	}
	fmt.Fprintf(&source, "static const uint16_t input_units[] = {%s0};\nstatic const ptrdiff_t spans[] = {%s0};\n", units.String(), spans.String())
	source.WriteString("typedef struct {const adamic_regex_program *program;size_t input,length;uint64_t last;size_t expected,count;uint64_t final;} probe;\nstatic const probe probes[]={\n")
	source.WriteString(rows.String())
	source.WriteString("};\n")
	source.WriteString(`
static bool pair_equal(adamic_object *pair,const ptrdiff_t *expected) {
 return pair==NULL?expected[0]<0:expected[0]>=0 && pair->slots[0].number==(double)expected[0] && pair->slots[1].number==(double)expected[1];
}
int main(int argc,char **argv) {
 adamic_start(argc,argv);adamic_regex_set_step_limit(argc>1?0:10000000);adamic_regex_set_regular_mode(argc==2?2:argc>2?0:1);size_t disagreements=0;
 for(size_t index=0;index<sizeof probes/sizeof probes[0];index++) {
  const probe *p=&probes[index];double *codes=malloc((p->length+1)*sizeof *codes);if(codes==NULL) abort();
  for(size_t k=0;k<p->length;k++) codes[k]=input_units[p->input+k];
  adamic_string *input=adamic_string_from_char_codes(p->length,codes);free(codes);
  adamic_object *regex=adamic_regex_new(p->program,&adamic_string_empty,&adamic_string_empty);regex->slots[1].number=(double)p->last;
  bool tested=adamic_regex_test(regex,input);bool same=tested==(p->count!=0) && regex->slots[1].number==(double)p->final;regex->slots[1].number=(double)p->last;
  adamic_array *match=adamic_regex_exec(regex,input);same=same && (match==NULL)==(p->count==0) && regex->slots[1].number==(double)p->final;
  if(match!=NULL) {
   same=same && match->length==p->count;
   adamic_array *indices=match->properties->slots[3].reference;
   for(size_t k=0;k<indices->length && k<p->count;k++) {
    same=same && pair_equal(indices->elements[k].reference,&spans[p->expected+2*k]);
    if(spans[p->expected+2*k]<0) same=same && match->elements[k].reference==NULL;
    else {
     adamic_string *want=adamic_string_slice(input,(double)spans[p->expected+2*k],(double)spans[p->expected+2*k+1],true);
     same=same && adamic_string_equal(want,match->elements[k].reference);adamic_release(want);
    }
   }
   adamic_object *groups=indices->properties->slots[2].reference;
   same=same && (groups!=NULL)==(p->program->group_count!=0);
   if(groups!=NULL) for(size_t k=0;k<p->program->group_count;k++) same=same && pair_equal(groups->slots[k].reference,&spans[p->expected+2*p->count+2*k]);
  }
  if(!same) {printf("DISAGREEMENT case=%zu test=%d expected=%d lastIndex=%.0f expected=%llu captures=%zu expected=%zu\n",index,tested,p->count!=0,regex->slots[1].number,(unsigned long long)p->final,match==NULL?0:match->length,p->count);disagreements++;}
  adamic_release(match);adamic_release(regex);adamic_release(input);
 }
 printf("native execution totals: %zu cases, %zu disagreements\n",sizeof probes/sizeof probes[0],disagreements);return disagreements==0?0:1;
}
`)
	if !wasmOnly {
		binary := filepath.Join(t.TempDir(), "regex-probes")
		t.Logf("compiling %d patterns, %d executions, %d C bytes", len(programs), len(cases), source.Len())
		buildRegexWitness(t, source.String(), binary, Options{Sanitize: true})
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		for _, arguments := range [][]string{nil, {"unlimited"}, {"unlimited", "vm"}} {
			command := exec.CommandContext(ctx, binary, arguments...)
			command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1")
			output, err := command.CombinedOutput()
			if err != nil {
				t.Fatalf("native regex oracle (%v): %v\n%s", arguments, err, output)
			}
			if !bytes.Contains(output, []byte("0 disagreements")) {
				t.Fatalf("native regex oracle missing totals: %s", output)
			}
			t.Logf("mode=%v %s", arguments, output)
		}
	}
	if wasmOnly {
		wasm := filepath.Join(t.TempDir(), "regex-probes.wasm")
		if err := Build(source.String(), wasm, Options{Target: "wasm32-wasi"}); err != nil {
			t.Fatal(err)
		}
		for _, arguments := range [][]string{nil, {"unlimited"}, {"unlimited", "vm"}} {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			args := append([]string{"--disable-warning=ExperimentalWarning", "../../oracle/wasi.mjs", wasm}, arguments...)
			output, err := exec.CommandContext(ctx, "node", args...).CombinedOutput()
			cancel()
			if err != nil || !bytes.Contains(output, []byte("0 disagreements")) {
				t.Fatalf("wasm32 regex oracle (%v): %v\n%s", arguments, err, output)
			}
			t.Logf("wasm32 mode=%v %s", arguments, output)
		}
	}
}

func runRegExpBytecodeRandomNodeShard(t *testing.T, unit int) {
	t.Helper()
	random := rand.New(rand.NewSource(0x8_7_5))
	atoms := []string{"a", "b", ".", "[a-z]", "[^]", "(a)", "(?:a)", "(?=a)", "(?<=a)", "\\d", "\\p{Letter}"}
	extra := []string{"(a|(b))", "(a?b)", "(?!b)", "(?<!b)", "[ab]", "[^a]", "\\w", "\\b", "^", "$", "k", "s", "(a)\\1", "\\1(a)", "(?<x>a)\\k<x>", "(?<=([ab]+)([bc]+))", "[\\q{ab|a|}]", "[[a-z]&&[^aeiou]]"}
	quantifiers := []string{"", "*", "+", "?", "{0}", "{1,}", "{2,4}", "*?", "+?", "??"}
	flags := []string{"", "u", "v", "i", "iu", "iv", "gimsy", "gyu", "ms"}
	inputs := []rune("abc 12\n\réKſ🌍")
	var cases []regexCase
	for len(cases) < 10000 {
		flag := flags[random.Intn(len(flags))]
		pattern := ""
		for j := 0; j < 1+random.Intn(4); j++ {
			atom := atoms[random.Intn(len(atoms))]
			if random.Intn(3) == 0 {
				atom = extra[random.Intn(len(extra))]
				if strings.Contains(atom, "\\q") || strings.Contains(atom, "&&") {
					if !strings.Contains(flag, "v") {
						atom = "[ab]"
					}
				}
			}
			if j > 0 && random.Intn(5) == 0 {
				pattern += "|"
			}
			pattern += atom + quantifiers[random.Intn(len(quantifiers))]
		}
		if _, err := regex.Parse(pattern, flag); err != nil {
			continue
		}
		var input []rune
		for j := 0; j < random.Intn(10); j++ {
			input = append(input, inputs[random.Intn(len(inputs))])
		}
		cases = append(cases, regexCase{Pattern: pattern, Flags: flag, Input: utf16.Encode(input), LastIndex: uint64(random.Intn(5))})
	}
	data, err := json.Marshal(cases)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	command := programCommand(ctx, "", "node", "-e", `
const cases=JSON.parse(require('fs').readFileSync(0,'utf8'));
if(!process.version.startsWith('v24.')) throw Error('Node 24 required');
for(const c of cases) {
 const r=new RegExp(c.pattern,c.flags.includes('d')?c.flags:c.flags+'d');
 r.lastIndex=c.lastIndex;
 const m=r.exec(String.fromCharCode(...c.input));
 c.expected={captures:m?Array.from(m.indices,x=>x??null):null,lastIndex:r.lastIndex,groups:m?.indices.groups?Object.fromEntries(Object.entries(m.indices.groups).map(([k,v])=>[k,v??null])):null};
}
process.stdout.write(JSON.stringify(cases));`)
	command.Stdin = bytes.NewReader(data)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("Node oracle: %v\\n%s", err, output)
	}
	if err = json.Unmarshal(output, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != randomRegexCases {
		t.Fatalf("random regex corpus: got %d, want %d", len(cases), randomRegexCases)
	}
	runRegexCases(t, cases[unit*randomRegexUnitSize:min((unit+1)*randomRegexUnitSize, len(cases))])
}

// Jump chains plus match have exact instruction counts, independent of scanning,
// backtracking and the compiler's regexp optimizations.
func TestRegExpNativeStepLimitBoundary(t *testing.T) {
	t.Parallel()
	const source = `#include "adamic.h"
#include <stdio.h>
#include <stdlib.h>
static const adamic_regex_instruction two[] = {{.op=3,.x=1},{.op=0}};
static const adamic_regex_instruction four[] = {{.op=3,.x=1},{.op=3,.x=2},{.op=3,.x=3},{.op=0}};
static const adamic_regex_program programs[] = {{.code=two},{.code=four}};
int main(int argc, char **argv) {
 adamic_start(argc,argv);
 adamic_regex_set_step_limit(strtoull(argv[2],NULL,10));
 adamic_object *regex=adamic_regex_new(&programs[atoi(argv[1])],&adamic_string_empty,&adamic_string_empty);
 bool matched=adamic_regex_test(regex,&adamic_string_empty);
 adamic_release(regex);
 printf("matched %d\n",matched);
 return 0;
}
`
	binary := filepath.Join(t.TempDir(), "boundary")
	if err := Build(source, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	for index, steps := range []int{2, 4} {
		for _, limit := range []int{steps, steps - 1} {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			command := exec.CommandContext(ctx, binary, fmt.Sprint(index), fmt.Sprint(limit))
			// Panic terminates before cleanup; successful cases retain leak detection.
			if limit < steps {
				command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0")
			}
			output, err := command.CombinedOutput()
			cancel()
			if limit == steps {
				if err != nil || string(output) != "matched 1\n" {
					t.Errorf("%d instructions at limit %d: want matched 1, got exit=%v output=%q", steps, limit, err, output)
				}
			} else {
				var exit *exec.ExitError
				if !errors.As(err, &exit) || exit.ExitCode() != 70 || string(output) != "adamic: panic: regexp: instruction step limit exceeded\n" {
					t.Errorf("%d instructions at limit %d: want exit 70 and step-limit diagnostic, got exit=%v output=%q", steps, limit, err, output)
				}
			}
		}
	}
}

// Library IteratorYieldResult types can also describe user objects. Read by
// field name, and preserve a missing optional done instead of assuming slot zero.
func TestRegExpIteratorResultShape(t *testing.T) {
	t.Parallel()
	source := `#include "adamic.h"
#include <stdio.h>
static const char *const names[] = {"value", "done"};
static const bool references[] = {false, false};
static const adamic_shape missing = {1,names,references,NULL};
static const adamic_shape reordered = {2,names,references,NULL};
int main(int argc,char **argv) {
 adamic_start(argc,argv);
 adamic_object *object=adamic_object_new(&missing);object->slots[0].number=42;
 adamic_maybe_boolean done=adamic_regex_done(object);adamic_release(object);
 if(done.present) {puts("unexpected iterator done for an absent field");return 1;}
 object=adamic_object_new(&reordered);object->slots[0].number=42;object->slots[1].boolean=true;
 done=adamic_regex_done(object);adamic_release(object);
 if(!done.present || !done.boolean) {puts("unexpected iterator done for reordered fields");return 1;}
 return 0;
}
`
	binary := filepath.Join(t.TempDir(), "iterator-shape")
	if err := Build(source, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	output, err := exec.Command(binary).CombinedOutput()
	if err != nil {
		t.Fatalf("iterator result shape: %v\n%s", err, output)
	}
}

func TestRegExpBytecodePatternUnits(t *testing.T) {
	t.Parallel()
	var cases []regexCase
	for _, unit := range []uint16{0xd800, 0xd801} {
		// JSON decoders can replace lone surrogates in pattern strings with the same
		// replacement rune. The preserved UTF-16 pattern must determine identity.
		c := regexCase{Pattern: "�", PatternUnits: []uint16{unit}, Input: []uint16{unit}}
		c.Expected.Captures = [][]int{{0, 1}}
		cases = append(cases, c)
	}
	runRegexCases(t, cases)
}

func TestRegExpBytecodeRandomNodeShard00(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeShard(t, 0)
}

func TestRegExpBytecodeRandomNodeShard01(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeShard(t, 1)
}

func TestRegExpBytecodeRandomNodeShard02(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeShard(t, 2)
}

func TestRegExpBytecodeRandomNodeShard03(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeShard(t, 3)
}

func TestRegExpBytecodeRandomNodeShard04(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeShard(t, 4)
}

func TestRegExpBytecodeRandomNodeShard05(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeShard(t, 5)
}

func TestRegExpBytecodeRandomNodeShard06(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeShard(t, 6)
}

func TestRegExpBytecodeRandomNodeShard07(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeShard(t, 7)
}

func TestRegExpBytecodeRandomNodeShard08(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeShard(t, 8)
}

func TestRegExpBytecodeRandomNodeShard09(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeShard(t, 9)
}

func TestRegExpBytecodeRandomNodeShard10(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeShard(t, 10)
}

func TestRegExpBytecodeRandomNodeShard11(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeShard(t, 11)
}

func TestRegExpBytecodeRandomNodeShard12(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeShard(t, 12)
}

func TestRegExpBytecodeRandomNodeShard13(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeShard(t, 13)
}

func TestRegExpBytecodeRandomNodeShard14(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeShard(t, 14)
}

func TestRegExpBytecodeRandomNodeShard15(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeShard(t, 15)
}

func TestRegExpBytecodeRandomNodeShard16(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeShard(t, 16)
}

func TestRegExpBytecodeRandomNodeShard17(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeShard(t, 17)
}

func TestRegExpBytecodeRandomNodeShard18(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeShard(t, 18)
}

func TestRegExpBytecodeRandomNodeShard19(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeShard(t, 19)
}

func TestRegExpBytecodeRandomNodeShard20(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeShard(t, 20)
}

func TestRegExpBytecodeRandomNodeShard21(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeShard(t, 21)
}

func TestRegExpBytecodeRandomNodeShard22(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeShard(t, 22)
}

func TestRegExpBytecodeRandomNodeShard23(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeShard(t, 23)
}

func TestRegExpBytecodeRandomNodeShard24(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeShard(t, 24)
}

func TestRegExpBytecodeRandomNodeShard25(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeShard(t, 25)
}

func TestRegExpBytecodeRandomNodeShard26(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeShard(t, 26)
}

func TestRegExpBytecodeRandomNodeShard27(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeShard(t, 27)
}

func TestRegExpBytecodeRandomNodeShard28(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeShard(t, 28)
}

func TestRegExpBytecodeRandomNodeShard29(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeShard(t, 29)
}

func TestRegExpBytecodeRandomNodeShard30(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeShard(t, 30)
}

func TestRegExpBytecodeRandomNodeShard31(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeShard(t, 31)
}

func TestRegExpBytecodeRandomNodeShard32(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeShard(t, 32)
}

func TestRegExpBytecodeRandomNodeShard33(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeShard(t, 33)
}

func TestRegExpBytecodeRandomNodeShard34(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeShard(t, 34)
}

func TestRegExpBytecodeRandomNodeShard35(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeShard(t, 35)
}

func TestRegExpBytecodeRandomNodeShard36(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeShard(t, 36)
}

func TestRegExpBytecodeRandomNodeShard37(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeShard(t, 37)
}

func TestRegExpBytecodeRandomNodeShard38(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeShard(t, 38)
}

func TestRegExpBytecodeRandomNodeShard39(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeShard(t, 39)
}
