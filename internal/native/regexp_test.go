package native

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/system-inc/adamic/internal/childguard"
	regex "github.com/system-inc/adamic/internal/regexp"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"sort"
	"strings"
	"testing"
	"time"
	"unicode/utf16"
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

// One executable runs all the observed test262 calls through the C interpreter,
// including captures, named groups, indices and lastIndex, under both sanitizers.
func TestRegExpBytecodeTest262(t *testing.T) {
	t.Parallel()
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
	runRegexCases(t, cases)
}
func runRegexCases(t *testing.T, cases []regexCase) {
	t.Helper()
	cases = regexCompatibleCases(t, cases)
	var source, rows, units, spans strings.Builder
	source.WriteString("#include \"adamic.h\"\n#include <stdio.h>\n#include <stdlib.h>\n#include <string.h>\n")
	programs := map[string]int{}
	regularPrograms := map[int]bool{}
	regularCases := 0
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
			regularPrograms[id] = strings.Contains(text, "_regular_code[]")
		}
		if regularPrograms[id] {
			regularCases++
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
  if((index+1)%128==0 || index+1==sizeof probes/sizeof probes[0]) {printf("native regex progress: %zu/%zu\n",index+1,sizeof probes/sizeof probes[0]);fflush(stdout);}
 }
 printf("native execution totals: %zu cases, %zu disagreements\n",sizeof probes/sizeof probes[0],disagreements);return disagreements==0?0:1;
}
`)
	binary := filepath.Join(t.TempDir(), "regex-probes")
	t.Logf("compiling %d patterns, %d executions, %d C bytes; regular eligible: %d patterns, %d executions", len(programs), len(cases), source.Len(), len(regularProgramsWithEngine(regularPrograms)), regularCases)
	if err := Build(source.String(), binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	for _, arguments := range [][]string{nil, {"unlimited"}, {"unlimited", "vm"}} {
		environment := ""
		// LeakSanitizer is Linux's; macOS AddressSanitizer rejects this option.
		if goruntime.GOOS == "linux" {
			environment = "ASAN_OPTIONS=detect_leaks=1"
		}
		output, err := runRegExpChild(t, binary, arguments, environment, childguard.Options{}, 2*time.Minute)
		if err != nil {
			t.Fatalf("native regex oracle (%v): %v\n%s", arguments, err, output)
		}
		if !bytes.Contains(output, []byte("0 disagreements")) {
			t.Fatalf("native regex oracle missing totals: %s", output)
		}
		t.Logf("mode=%v %s", arguments, output)
	}

}

func runRegExpBytecodeRandomNodeUnit(t *testing.T, unit int) {
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
	command := exec.Command("node", "-e", `
const cases=JSON.parse(require('fs').readFileSync(0,'utf8'));
if(!process.version.startsWith('v24.')) throw Error('Node 24 required');
const fs=require('fs');fs.writeSync(2,'regex progress: start\n');
let completed=0;
for(const c of cases) {
 const r=new RegExp(c.pattern,c.flags.includes('d')?c.flags:c.flags+'d');
 r.lastIndex=c.lastIndex;
 const m=r.exec(String.fromCharCode(...c.input));
 c.expected={captures:m?Array.from(m.indices,x=>x??null):null,lastIndex:r.lastIndex,groups:m?.indices.groups?Object.fromEntries(Object.entries(m.indices.groups).map(([k,v])=>[k,v??null])):null};
 if(++completed%128===0) fs.writeSync(2,'regex progress: '+completed+'/'+cases.length+'\n');
}
process.stdout.write(JSON.stringify(cases));`)
	command.Stdin = bytes.NewReader(data)
	var output, progress bytes.Buffer
	command.Stdout, command.Stderr = &output, &progress
	if err := childguard.Run(command, childguard.Options{}); err != nil {
		t.Fatalf("Node oracle: %v\n%s\n%s", err, progress.Bytes(), output.Bytes())
	}
	if err = json.Unmarshal(output.Bytes(), &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != randomRegexCases {
		t.Fatalf("random regex corpus: got %d, want %d", len(cases), randomRegexCases)
	}
	runRegexCases(t, cases[unit*randomRegexUnitSize:min((unit+1)*randomRegexUnitSize, len(cases))])
}

func TestRegExpNativeStepLimit(t *testing.T) {
	t.Parallel()
	program, err := regex.Compile("(a+)+$", "")
	if err != nil {
		t.Fatal(err)
	}
	declarations, err := program.NativeDeclarations("probe")
	if err != nil {
		t.Fatal(err)
	}
	source := "#include \"adamic.h\"\n" + declarations + `
static adamic_string input = ADAMIC_STRING("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaab");
int main(int argc,char **argv) {
 adamic_start(argc,argv);adamic_regex_set_step_limit(1000);
 adamic_object *regex=adamic_regex_new(&probe,&adamic_string_empty,&adamic_string_empty);
 bool result=adamic_regex_test(regex,&input);adamic_release(regex);return result?1:0;
}
`
	binary := filepath.Join(t.TempDir(), "budget")
	if err := Build(source, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	output, err := runRegExpChild(t, binary, nil, "ASAN_OPTIONS=detect_leaks=0", childguard.Options{}, 10*time.Second)
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 70 || !bytes.Contains(output, []byte("regexp: instruction step limit exceeded")) {
		t.Fatalf("native catastrophic backtracking: exit=%v output=%s", err, output)
	}
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

func regularProgramsWithEngine(programs map[int]bool) []int {
	var ids []int
	for id, eligible := range programs {
		if eligible {
			ids = append(ids, id)
		}
	}
	return ids
}

func TestRegExpBytecodeRandomNodeUnit00(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeUnit(t, 0)
}

func TestRegExpBytecodeRandomNodeUnit01(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeUnit(t, 1)
}

func TestRegExpBytecodeRandomNodeUnit02(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeUnit(t, 2)
}

func TestRegExpBytecodeRandomNodeUnit03(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeUnit(t, 3)
}

func TestRegExpBytecodeRandomNodeUnit04(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeUnit(t, 4)
}

func TestRegExpBytecodeRandomNodeUnit05(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeUnit(t, 5)
}

func TestRegExpBytecodeRandomNodeUnit06(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeUnit(t, 6)
}

func TestRegExpBytecodeRandomNodeUnit07(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeUnit(t, 7)
}

func TestRegExpBytecodeRandomNodeUnit08(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeUnit(t, 8)
}

func TestRegExpBytecodeRandomNodeUnit09(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeUnit(t, 9)
}

func TestRegExpBytecodeRandomNodeUnit10(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeUnit(t, 10)
}

func TestRegExpBytecodeRandomNodeUnit11(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeUnit(t, 11)
}

func TestRegExpBytecodeRandomNodeUnit12(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeUnit(t, 12)
}

func TestRegExpBytecodeRandomNodeUnit13(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeUnit(t, 13)
}

func TestRegExpBytecodeRandomNodeUnit14(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeUnit(t, 14)
}

func TestRegExpBytecodeRandomNodeUnit15(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeUnit(t, 15)
}

func TestRegExpBytecodeRandomNodeUnit16(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeUnit(t, 16)
}

func TestRegExpBytecodeRandomNodeUnit17(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeUnit(t, 17)
}

func TestRegExpBytecodeRandomNodeUnit18(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeUnit(t, 18)
}

func TestRegExpBytecodeRandomNodeUnit19(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeUnit(t, 19)
}

func TestRegExpBytecodeRandomNodeUnit20(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeUnit(t, 20)
}

func TestRegExpBytecodeRandomNodeUnit21(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeUnit(t, 21)
}

func TestRegExpBytecodeRandomNodeUnit22(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeUnit(t, 22)
}

func TestRegExpBytecodeRandomNodeUnit23(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeUnit(t, 23)
}

func TestRegExpBytecodeRandomNodeUnit24(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeUnit(t, 24)
}

func TestRegExpBytecodeRandomNodeUnit25(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeUnit(t, 25)
}

func TestRegExpBytecodeRandomNodeUnit26(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeUnit(t, 26)
}

func TestRegExpBytecodeRandomNodeUnit27(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeUnit(t, 27)
}

func TestRegExpBytecodeRandomNodeUnit28(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeUnit(t, 28)
}

func TestRegExpBytecodeRandomNodeUnit29(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeUnit(t, 29)
}

func TestRegExpBytecodeRandomNodeUnit30(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeUnit(t, 30)
}

func TestRegExpBytecodeRandomNodeUnit31(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeUnit(t, 31)
}

func TestRegExpBytecodeRandomNodeUnit32(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeUnit(t, 32)
}

func TestRegExpBytecodeRandomNodeUnit33(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeUnit(t, 33)
}

func TestRegExpBytecodeRandomNodeUnit34(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeUnit(t, 34)
}

func TestRegExpBytecodeRandomNodeUnit35(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeUnit(t, 35)
}

func TestRegExpBytecodeRandomNodeUnit36(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeUnit(t, 36)
}

func TestRegExpBytecodeRandomNodeUnit37(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeUnit(t, 37)
}

func TestRegExpBytecodeRandomNodeUnit38(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeUnit(t, 38)
}

func TestRegExpBytecodeRandomNodeUnit39(t *testing.T) {
	t.Parallel()
	runRegExpBytecodeRandomNodeUnit(t, 39)
}
