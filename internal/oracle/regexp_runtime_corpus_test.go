package oracle

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
	reference "github.com/system-inc/adamic/internal/regexp"
)

// This corpus is test input from f9e6ddc8's cohere-patterns pass, with shared
// subjects deduplicated. Every source-valid inventory pattern retains every
// subject and its provenance; no expected Adamic result is stored.
func TestRuntimeConstructorCorpusBudget(t *testing.T) {
	file, err := os.Open(filepath.Join(repository, runtimeConstructorsDirectory, "corpus.json.gz"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	compressed, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer compressed.Close()
	var corpus struct {
		Cohere   string
		Subjects []struct {
			Units        []uint16
			Source, Kind string
		}
		Patterns []struct {
			PatternUnits []uint16
			Flags        string
			Sources      []string
			Inputs       []int
		}
	}
	if err := json.NewDecoder(compressed).Decode(&corpus); err != nil {
		t.Fatal(err)
	}
	if corpus.Cohere != "7945d102a6c18dd36adf9114a758ce646e8b2359" || len(corpus.Patterns) != 503 || len(corpus.Subjects) != 120279 {
		t.Fatal("corpus pin or size changed")
	}
	inventoryFile, err := os.Open(filepath.Join(repository, runtimeConstructorsDirectory, "inventory.json.gz"))
	if err != nil {
		t.Fatal(err)
	}
	defer inventoryFile.Close()
	inventoryZip, err := gzip.NewReader(inventoryFile)
	if err != nil {
		t.Fatal(err)
	}
	defer inventoryZip.Close()
	var inventory []struct {
		File, Flags, LiteralToken, SourceError string
		Line                                   int
		PatternUnits                           []uint16
	}
	if err := json.NewDecoder(inventoryZip).Decode(&inventory); err != nil {
		t.Fatal(err)
	}
	if len(inventory) != 969 {
		t.Fatal("inventory size changed")
	}
	known := map[string]bool{}
	invalidTokens := 0
	for _, p := range inventory {
		if p.SourceError != "" {
			invalidTokens++
			continue
		}
		known[fmt.Sprint(p.PatternUnits)+"/"+p.Flags] = true
	}
	if invalidTokens != 238 || len(known) != 503 {
		t.Fatal("inventory source-token classification changed")
	}
	for _, p := range corpus.Patterns {
		key := fmt.Sprint(p.PatternUnits) + "/" + p.Flags
		if !known[key] {
			t.Fatal("execution corpus pattern not in inventory")
		}
		delete(known, key)
	}
	if len(known) != 0 {
		t.Fatal("inventory patterns missing execution coverage")
	}
	var wire bytes.Buffer
	u32 := func(n int) {
		if err := binary.Write(&wire, binary.LittleEndian, uint32(n)); err != nil {
			t.Fatal(err)
		}
	}
	units := func(xs []uint16) {
		u32(len(xs))
		if err := binary.Write(&wire, binary.LittleEndian, xs); err != nil {
			t.Fatal(err)
		}
	}
	accepted, syntax, refused, withoutSubjects, total := 0, 0, 0, 0, 0
	var selected []int
	for index, p := range corpus.Patterns {
		if len(p.Sources) == 0 {
			t.Fatal("pattern lacks source")
		}
		program, err := reference.CompileUTF16(p.PatternUnits, p.Flags)
		if err == nil {
			err = program.NativeCompatibility()
		}
		if err == nil {
			_, err = program.NativeDeclarations(fmt.Sprintf("corpus_%d", index))
		}
		if err != nil {
			// Node's independent classification below decides whether this is
			// an actual syntax error or a loud native-only refusal.
			t.Logf("REFUSED compile pattern=%d flags=%s: %v", index, p.Flags, err)
			var divergence *reference.V8DivergenceError
			if !errors.As(err, &divergence) || !strings.Contains(err.Error(), "clamps quantifier bounds") || stringFromUnits(p.PatternUnits) != "a{9223372036854775808,9223372036854775807}" {
				t.Fatalf("unexpected native-only refusal: %v", err)
			}
			refused++
			continue
		}
		accepted++
		selected = append(selected, index)
		if len(p.Inputs) == 0 {
			withoutSubjects++
		}
		units(p.PatternUnits)
		units(asciiUnits(p.Flags))
		count := len(p.Inputs)
		if count == 0 {
			count = 1
		}
		u32(count)
		if len(p.Inputs) == 0 {
			units(nil)
			total++
		}
		for _, id := range p.Inputs {
			if id < 0 || id >= len(corpus.Subjects) {
				t.Fatal("bad input index")
			}
			s := corpus.Subjects[id]
			if s.Source == "" || s.Kind == "" {
				t.Fatal("subject lacks provenance")
			}
			units(s.Units)
			total++
		}
	}
	u32(0xffffffff)
	query, _ := json.Marshal(struct {
		Selected  []int
		Corpus    any
		Inventory any
	}{selected, corpus, inventory})
	command := exec.Command("node", "-e", `
if(process.version!=='v24.19.0')throw Error('requires Node 24.19.0');
const q=JSON.parse(require('fs').readFileSync(0,'utf8')), c=q.Corpus;
const text=units=>{let s='';for(let i=0;i<units.length;i+=8192)s+=String.fromCharCode(...units.slice(i,i+8192));return s};
const vm=require('node:vm');for(const p of q.Inventory){if(!p.LiteralToken)continue;let message='';try{new vm.Script('('+p.LiteralToken+')')}catch(e){if(!(e instanceof SyntaxError))throw e;message=e.message}if(message!==p.SourceError)throw Error('inventory source-token diagnostic changed: '+p.File+':'+p.Line+' '+message)}
const strings=c.Subjects.map(s=>text(s.Units));
for(let i=0;i<c.Patterns.length;i++){try{new RegExp(text(c.Patterns[i].PatternUnits),c.Patterns[i].Flags)}catch(e){if(!(e instanceof SyntaxError))throw e;process.stderr.write('SYNTAX '+i+' '+e.message+'\n')}}
let results=[];for(const i of q.Selected){const p=c.Patterns[i],r=new RegExp(text(p.PatternUnits),p.Flags);for(const id of (p.Inputs.length?p.Inputs:[-1])){r.lastIndex=0;results.push(r.test(strings[id]??'')?1:0)}}process.stdout.write(Buffer.from(results));`)
	command.Stdin = bytes.NewReader(query)
	var expected, nodeErrors bytes.Buffer
	command.Stdout = &expected
	command.Stderr = &nodeErrors
	if err := runChild(command); err != nil {
		t.Fatalf("Node corpus: %v %s", err, nodeErrors.Bytes())
	}
	syntax = strings.Count(nodeErrors.String(), "SYNTAX ")
	if syntax != 0 {
		t.Fatalf("source-valid inventory contains %d Node-invalid patterns: %s", syntax, nodeErrors.Bytes())
	}
	if expected.Len() != total {
		t.Fatal("Node observation count changed")
	}
	input := filepath.Join(t.TempDir(), "corpus.bin")
	if err := os.WriteFile(input, wire.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}
	path, _ := filepath.Abs(filepath.Join(repository, runtimeConstructorsDirectory, "construction_failure.a"))
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	main := strings.Index(source, "int main(")
	if main < 0 {
		t.Fatal("main missing")
	}
	source = source[:main] + runtimeConstructorCorpusHarness
	binaryPath := filepath.Join(t.TempDir(), "corpus")
	if err := native.Build(source, binaryPath, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	command = exec.Command(binaryPath, input)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1")
	var actual, errors bytes.Buffer
	command.Stdout = &actual
	command.Stderr = &errors
	if err := runChild(command); err != nil {
		t.Fatalf("runtime corpus stopped: %v %s", err, errors.Bytes())
	}
	if errors.Len() != 0 {
		t.Fatalf("runtime corpus stderr: %s", errors.Bytes())
	}
	if !bytes.Equal(expected.Bytes(), actual.Bytes()) {
		t.Fatalf("corpus disagrees with Node: expected %d observations, got %d", expected.Len(), actual.Len())
	}
	t.Logf("inventory=969 source-token-refused=238 distinct-source-valid=%d accepted=%d syntax=%d native-only-refused=%d empty-only-probes=%d matches=%d step-limit=100000000 budget-stops=0 disagreements=0", len(corpus.Patterns), accepted, syntax, refused, withoutSubjects, total)
	if refused != 1 {
		t.Fatalf("expected the single ruled V8 divergence, got %d refusals", refused)
	}
}

func asciiUnits(s string) []uint16 {
	r := make([]uint16, len(s))
	for i := range s {
		r[i] = uint16(s[i])
	}
	return r
}

const runtimeConstructorCorpusHarness = `
#include <stdio.h>
#include <stdlib.h>
static uint32_t corpus_number(FILE *in) {
    unsigned char b[4];if(fread(b,1,4,in)!=4)exit(2);
    return (uint32_t)b[0]|((uint32_t)b[1]<<8)|((uint32_t)b[2]<<16)|((uint32_t)b[3]<<24);
}
static adamic_string *corpus_string(FILE *in,uint32_t count) {
    double *units=malloc(((size_t)count+1)*sizeof(*units));if(units==NULL)exit(2);
    for(uint32_t i=0;i<count;i++){unsigned char b[2];if(fread(b,1,2,in)!=2)exit(2);units[i]=(double)((uint16_t)b[0]|((uint16_t)b[1]<<8));}
    adamic_string *s=adamic_string_from_char_codes(count,units);free(units);return s;
}
int main(int argc,char **argv) {
    adamic_start(argc,argv);if(argc!=2)return 2;FILE *in=fopen(argv[1],"rb");if(in==NULL)return 2;
    adamic_regex_set_step_limit(UINT64_C(100000000));
    for(;;){uint32_t n=corpus_number(in);if(n==UINT32_MAX)break;
        adamic_string *pattern=corpus_string(in,n),*flags=corpus_string(in,corpus_number(in));
        adamic_object *r=adamic_regex_compile_new(pattern,flags);if(r==NULL)return 3;
        adamic_release(pattern);adamic_release(flags);uint32_t count=corpus_number(in);
        for(uint32_t i=0;i<count;i++){adamic_string *text=corpus_string(in,corpus_number(in));r->slots[1].number=0;
            if(fputc(adamic_regex_test(r,text)?1:0,stdout)==EOF)return 2;adamic_release(text);}
        adamic_release(r);
    }
    if(fclose(in)!=0)return 2;return 0;
}
`

func stringFromUnits(xs []uint16) string {
	var b strings.Builder
	for _, x := range xs {
		b.WriteRune(rune(x))
	}
	return b.String()
}
