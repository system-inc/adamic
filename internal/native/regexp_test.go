package native

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
		output, err := runRegExpChild(t, binary, arguments, environment, 5*time.Minute, 2*time.Minute)
		if err != nil {
			t.Fatalf("native regex oracle (%v): %v\n%s", arguments, err, output)
		}
		if !bytes.Contains(output, []byte("0 disagreements")) {
			t.Fatalf("native regex oracle missing totals: %s", output)
		}
		t.Logf("mode=%v %s", arguments, output)
	}

}

func TestRegExpBytecodeRandomNode(t *testing.T) {
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "node", "-e", `
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
	runRegexCases(t, cases)
}
func TestRegExpNativeStepLimit(t *testing.T) {
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
	output, err := runRegExpChild(t, binary, nil, "ASAN_OPTIONS=detect_leaks=0", 5*time.Minute, 10*time.Second)
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 70 || !bytes.Contains(output, []byte("regexp: instruction step limit exceeded")) {
		t.Fatalf("native catastrophic backtracking: exit=%v output=%s", err, output)
	}
}

// Library IteratorYieldResult types can also describe user objects. Read by
// field name, and preserve a missing optional done instead of assuming slot zero.
func TestRegExpIteratorResultShape(t *testing.T) {
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
