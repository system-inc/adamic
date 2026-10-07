package typeaware

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The independent production-rule oracle remains the specification. Indexes
// only change lookup cost; controls include repeated sites and duplicate names.
func TestFlowIndexAgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if path := os.Getenv("ADAMIC_SPEED_ARTIFACTS"); path != "" {
		directory, err = filepath.Abs(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
	}
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("checker", "", false)
	oracle := volumeOracle(h, "coverage-oracle", "oracle_coverage.go")
	binary := h.build(stage0, "coverage", filepath.Join(repository, "stage1/cohere/typeaware/coverage_suite.ts"), archive, false)
	config := filepath.Join(repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	controls := append(coverageControls(),
		"interface S{left:{a:number};right:{a:number;b:number}};interface T{left:{a:number;b?:number};right:{a:number;b?:number}};declare const s:S;export const first:T=s;export const second:T=s;",
		"interface A{x:number};interface B extends A{y:number};declare const rows:B[];export const first:A[]=rows;export const second:A[]=rows;")
	paths := []string{}
	for i, source := range controls {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.ts", i), source+"\nexport {};\n"))
	}
	paths = append(paths, h.write("helper.ts", "export const Used=1,Unused=2,AlsoUnused=3;\n"))
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, change := range []struct{ name, file, from, to string }{
		{"wrong-type-slot", "types.ts", "if(record !== undefined && record.id === id) { return record; }", "if(record !== undefined && record.id === id) { return this.ordered[0] ?? panic('missing checker type index'); }"},
		{"wrong-property-slot", "checker_facts.ts", "this.firstName.set(name, at);", "this.firstName.set(name, names.length - 1 - at);"},
		{"retained-pairs", "flow.ts", "this.visited.clear();", "// Mutant carries visited pairs into the next judgment."},
	} {
		mutant := coverageSourceMutant(h, stage0, archive, change.name, change.file, change.from, change.to)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("%s survived byte oracle", change.name)
		}
		t.Logf("%s: exit 0, Go byte oracle catches byte %d; %s", change.name, firstDifference(got.stdout, truth.stdout), summary(got.stdout))
		if err := os.Remove(mutant); err != nil {
			t.Fatal(err)
		}
	}
	// An independent linear-search expectation also covers sparse identities,
	// duplicate property names and the original record order.
	indexProbe := h.write("index-probe.ts", fmt.Sprintf(`import { Types } from %q;
import { TypeFact } from %q;
import { Properties } from %q;
const records=[new TypeFact(47,0,'first',false,0,false,-1,0,0,[],[]),new TypeFact(2,0,'small',false,0,false,-1,0,0,[],[]),new TypeFact(47,0,'second',false,0,false,-1,0,0,[],[]),new TypeFact(9007199254740991,0,'large',false,0,false,-1,0,0,[],[])];
const graph=new Types(true,true,[47],[],records);
const properties=new Properties(['dup','other','dup',''],[1,2,3,4],[0,0,0,0],[false,false,false,false],graph);
console.log(graph.type(47).name+' '+graph.type(2).name+' '+graph.type(9007199254740991).name+' '+(graph.records[0]?.name ?? '')+' '+(properties.firstName.get('dup') ?? -1).toString()+' '+(properties.firstName.get('') ?? -1).toString()+' '+(properties.firstName.has('absent') ? 'true' : 'false'));
`, filepath.Join(repository, "stage1/cohere/typeaware/types.ts"), filepath.Join(repository, "stage1/cohere/typeaware/type_fact.ts"), filepath.Join(repository, "stage1/cohere/typeaware/checker_facts.ts")))
	indexBinary := h.build(stage0, "index-probe", indexProbe, archive, false)
	indexResult := h.must("index-probe-run", exec.Command(indexBinary))
	if string(indexResult.stdout) != "first small large first 0 3 false\n" || len(indexResult.stderr) != 0 {
		t.Fatalf("index changed linear-search semantics: %s %s", indexResult.stdout, indexResult.stderr)
	}
	coverageSourceMutant(h, stage0, archive, "last-property", "checker_facts.ts", "if(!this.firstName.has(name))", "if(true)")
	indexData, err := os.ReadFile(indexProbe)
	if err != nil {
		t.Fatal(err)
	}
	lastSource := strings.Replace(string(indexData), filepath.Join(repository, "stage1/cohere/typeaware/checker_facts.ts"), filepath.Join(directory, "last-property-source/checker_facts.ts"), 1)
	lastBinary := h.build(stage0, "last-property-probe", h.write("last-property-probe.ts", lastSource), archive, false)
	lastResult := h.must("last-property-run", exec.Command(lastBinary))
	if string(lastResult.stdout) != "first small large first 2 3 false\n" || len(lastResult.stderr) != 0 {
		t.Fatalf("last property mutant did not replace first slot: %s %s", lastResult.stdout, lastResult.stderr)
	}
	t.Log("duplicate-property mutant: exits 0, linear-search expectation catches slot 2 instead of first slot 0")
	// Missing IDs must still refuse, including values between stored IDs. This
	// exercises the equality guard independently of graph-link validation.
	probe := h.write("missing-type.ts", fmt.Sprintf(`import { Types } from %q;
import { TypeFact } from %q;
const fact=new TypeFact(11,0,'',false,0,false,-1,0,0,[],[]);
console.log(new Types(true,true,[11],[],[fact]).type(3).id.toString());
`, filepath.Join(repository, "stage1/cohere/typeaware/types.ts"), filepath.Join(repository, "stage1/cohere/typeaware/type_fact.ts")))
	missing := h.build(stage0, "missing-type", probe, archive, false)
	got := h.run("missing-type-run", exec.Command(missing))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: missing checker type identity\n" {
		t.Fatalf("missing identity escaped: %v %s", got.err, got.stderr)
	}
	// The same mutation source helper builds a full runner; the direct probe is
	// rebuilt against its copied types module to prove the guard can fail.
	coverageSourceMutant(h, stage0, archive, "missing-equality", "types.ts", "record !== undefined && record.id === id", "record !== undefined")
	data, err := os.ReadFile(probe)
	if err != nil {
		t.Fatal(err)
	}
	copied := strings.Replace(string(data), filepath.Join(repository, "stage1/cohere/typeaware/types.ts"), filepath.Join(directory, "missing-equality-source/types.ts"), 1)
	mutantProbe := h.build(stage0, "missing-equality-probe", h.write("missing-equality-probe.ts", copied), archive, false)
	escaped := h.must("missing-equality-run", exec.Command(mutantProbe))
	if string(escaped.stdout) != "11\n" || len(escaped.stderr) != 0 {
		t.Fatalf("equality mutant did not escape as expected: %s %s", escaped.stdout, escaped.stderr)
	}
	t.Log("missing identity: panic 70; equality-guard mutant exits 0 and returns wrong identity 11")
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "coverage-asan", filepath.Join(repository, "stage1/cohere/typeaware/coverage_suite.ts"), sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	// Types is shared with the older sixteen rules, so validate both complete
	// finding/fix streams, normally and under ASan/UBSan/LSan.
	oldOracle := volumeOracle(h, "volume-oracle", "oracle_volume.go")
	old := h.build(stage0, "volume", filepath.Join(repository, "stage1/cohere/typeaware/volume_suite.ts"), archive, false)
	oldAsan := h.build(stage0, "volume-asan", filepath.Join(repository, "stage1/cohere/typeaware/volume_suite.ts"), sanitized, true)
	for _, corpus := range []struct{ name, manifest, config string }{
		{"repository", os.Getenv("ADAMIC_SPEED_REPOSITORY_MANIFEST"), filepath.Join(repository, "tsconfig.json")},
		{"compiler", os.Getenv("ADAMIC_SPEED_COMPILER_MANIFEST"), filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json")},
	} {
		if corpus.manifest == "" {
			continue
		}
		h.compare(corpus.name+"-coverage", oracle, binary, corpus.config, corpus.manifest)
		h.compare(corpus.name+"-coverage-asan", oracle, asan, corpus.config, corpus.manifest)
		h.compare(corpus.name+"-volume", oldOracle, old, corpus.config, corpus.manifest)
		h.compare(corpus.name+"-volume-asan", oldOracle, oldAsan, corpus.config, corpus.manifest)
	}
}

func TestFlowTypeIndexBoundsMutant(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	if path := os.Getenv("ADAMIC_SPEED_BOUNDS_ARTIFACTS"); path != "" {
		directory, err = filepath.Abs(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
	}
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("checker", "", false)
	source := fmt.Sprintf(`import { Types } from %q;
import { TypeFact } from %q;
const fact=new TypeFact(11,0,'eleven',false,0,false,-1,0,0,[],[]);
console.log(new Types(true,true,[11],[],[fact]).type(11).name);
`, filepath.Join(repository, "stage1/cohere/typeaware/types.ts"), filepath.Join(repository, "stage1/cohere/typeaware/type_fact.ts"))
	normal := h.build(stage0, "bounds-control", h.write("bounds-control.ts", source), archive, false)
	good := h.must("bounds-control-run", exec.Command(normal))
	if string(good.stdout) != "eleven\n" || len(good.stderr) != 0 {
		t.Fatalf("bounds control failed: %s %s", good.stdout, good.stderr)
	}
	coverageSourceMutant(h, stage0, archive, "past-end", "types.ts", "const middle = Math.floor((first + last) / 2);", "const middle = this.ordered.length;")
	changed := strings.Replace(source, filepath.Join(repository, "stage1/cohere/typeaware/types.ts"), filepath.Join(directory, "past-end-source/types.ts"), 1)
	mutant := h.build(stage0, "past-end-probe", h.write("past-end-probe.ts", changed), archive, false)
	got := h.run("past-end-run", exec.Command(mutant))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: missing checker type index\n" {
		t.Fatalf("past-end index escaped: %v %s", got.err, got.stderr)
	}
	t.Log("past-end midpoint mutant: panic 70, missing checker type index")
}
