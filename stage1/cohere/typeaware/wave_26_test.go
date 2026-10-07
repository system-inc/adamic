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

const wave26Preamble = `declare class GraphQlOperationContext<T> { key:T; }
declare function InjectGraphQlOperationContext(): ParameterDecorator;
declare function GraphQlQuery(x?:unknown): MethodDecorator;
declare function GraphQlMutation(x?:unknown): MethodDecorator;
declare function GraphQlFieldResolver(x?:unknown): MethodDecorator;
declare function SomethingElse(x?:unknown): MethodDecorator;
declare class A {a:string;}
declare class B {b:number;}
declare class Derived extends A {extra:number;}
type Branded<T> = MethodDecorator & {readonly __expectedReturnType?:T};
declare function Strict():Branded<string>;
declare function Plain():MethodDecorator;
declare function Any():Branded<any>;
declare function Unknown():Branded<unknown>;
declare const Namespace:{Strict:typeof Strict};
declare const bare:Branded<string>;
declare const OrmManyToOne:any, OrmOneToMany:any, OrmOneToOne:any, OrmColumn:any;
declare const Orm:{OrmManyToOne:any};
`

func wave26Controls() []string {
	controls := []string{
		"class R { @GraphQlQuery() f(@InjectGraphQlOperationContext() c:GraphQlOperationContext<B>):A{return null as never;} }",
		"class R { @GraphQlQuery() f(@InjectGraphQlOperationContext() c:B):A{return null as never;} }",
		"class R { @GraphQlQuery() f(@InjectGraphQlOperationContext() c:GraphQlOperationContext<Derived>):A{return null as never;} }",
		"class R { @GraphQlMutation() f(@InjectGraphQlOperationContext() c:GraphQlOperationContext<A>):Derived{return null as never;} }",
		"class R { @GraphQlFieldResolver() f(@InjectGraphQlOperationContext() c:GraphQlOperationContext<A>):A|B{return null as never;} }",
		"class R { @GraphQlQuery() f(@InjectGraphQlOperationContext() c:GraphQlOperationContext<A|B>):A|B{return null as never;} }",
		"class R { @SomethingElse() f(@InjectGraphQlOperationContext() c:GraphQlOperationContext<B>):A{return null as never;} }",
		"class R { f(@InjectGraphQlOperationContext() c:GraphQlOperationContext<B>):A{return null as never;} }",
		"class R { @GraphQlQuery() f(@InjectGraphQlOperationContext() c:GraphQlOperationContext<A>):A {class Local {@InjectGraphQlOperationContext() p:GraphQlOperationContext<B>; @InjectGraphQlOperationContext() m():B{return null as never;}} return null as never;} }",
		"class R { @GraphQlQuery() f(@InjectGraphQlOperationContext() c:GraphQlOperationContext):A{return null as never;} }",
		"class R { @Strict() bad():number{return 1;} @Strict() narrow():string{return 'ok';} @Strict() good():string|undefined{return undefined;} @Plain() plain():number{return 1;} @Any() any():void{} @Unknown() unknown():never{throw 1;} }",
		"class R { @Namespace.Strict() bad(){return 1;} @bare badBare():number{return 1;} @Strict() property:()=>number=()=>1; @Strict() get accessor(){return 1;} f(@Strict() p:number){} }",
		"/* 世界 🌍 */\r\nclass R { @Strict() méthoδ():number{return 1;} @OrmManyToOne() rélation:A|null; }\r\n",
	}
	for _, wrapped := range []string{"A", "A[]", "ReadonlyArray<A>", "Promise<A>", "Promise<Readonly<A[]>>", "Readonly<A>", "A|null|undefined"} {
		controls = append(controls, fmt.Sprintf("class R { @GraphQlQuery() f(@InjectGraphQlOperationContext() c:GraphQlOperationContext<%s>):Promise<Readonly<A[]>>{return null as never;} }", wrapped))
	}
	for _, annotation := range []string{"A", "A|null", "A|undefined", "A|null|undefined", "any", "unknown", "void", "never", "A[]"} {
		controls = append(controls, "class R { @OrmManyToOne() p:"+annotation+"; @OrmOneToMany() q:"+annotation+"; @OrmOneToOne() r:"+annotation+"; }")
	}
	controls = append(controls, "class R {@OrmManyToOne() optional?:A; @OrmManyToOne bare:A; @Orm.OrmManyToOne() namespaced:A; @OrmManyToOne() ['computed']:A; @OrmManyToOne() 'literal':A; @OrmManyToOne() #private:A; @OrmColumn() ordinary:A; @OrmColumn() @OrmManyToOne() declare relation:A;}")
	return controls
}

func wave26Mutant(h *harness, stage0, archive, name, file, from, to string) string {
	directory := filepath.Join(h.directory, name+"-source")
	if err := os.MkdirAll(directory, 0755); err != nil {
		h.t.Fatal(err)
	}
	for _, extension := range []string{"*.ts", "*.a"} {
		paths, err := filepath.Glob(filepath.Join(h.repository, "stage1/cohere/typeaware", extension))
		if err != nil {
			h.t.Fatal(err)
		}
		for _, path := range paths {
			data, err := os.ReadFile(path)
			if err != nil {
				h.t.Fatal(err)
			}
			source := string(data)
			if filepath.Base(path) == file {
				if strings.Count(source, from) != 1 {
					h.t.Fatalf("nonunique mutant %s", name)
				}
				source = strings.Replace(source, from, to, 1)
			}
			source = strings.ReplaceAll(source, "../../typescript/", filepath.Join(h.repository, "stage1/typescript")+"/")
			source = strings.ReplaceAll(source, "../lint/", filepath.Join(h.repository, "stage1/cohere/lint")+"/")
			if err := os.WriteFile(filepath.Join(directory, filepath.Base(path)), []byte(source), 0644); err != nil {
				h.t.Fatal(err)
			}
		}
	}
	return h.build(stage0, name, filepath.Join(directory, "wave_26.a"), archive, false)
}

// Not parallel: builds, sanitizers and cost observations share a machine.
func TestWave26AgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE26_ARTIFACTS")
	if directory == "" {
		directory = t.TempDir()
	}
	if err := os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, repository: repository, directory: directory}
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("checker", "", false)
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_26.a")
	binary := h.build(stage0, "wave26", entry, archive, false)
	oracle := volumeOracle(h, "wave26-oracle", "oracle_wave_26.go")
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"experimentalDecorators":true,"target":"ES2022","module":"NodeNext","lib":["ES2022"],"noEmit":true},"files":["control-000.a"]}`)
	var paths []string
	for i, source := range wave26Controls() {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), wave26Preamble+source+"\nexport {};\n"))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"matching-operation-context", "matching-provider-return", "optional-relation"} {
		if !bytes.Contains(truth.stdout, []byte("\tbase/correctness-require-"+name+"\t")) {
			t.Fatalf("no positive control for %s", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave26-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, change := range []struct{ name, file, from, to string }{
		{"operation", "matching_operation_context.a", "checker.compare(parameter, generic, result) && checker.compare(parameter, result, generic)", "checker.compare(parameter, generic, result)"},
		{"provider", "matching_provider_return.a", "result.id, expected.root().id", "expected.root().id, result.id"},
		{"relation", "optional_relation.a", "1 | 2 | 4", "1 | 2 | 4 | 8"},
	} {
		mutant := wave26Mutant(h, stage0, archive, change.name, change.file, change.from, change.to)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("%s mutant survived", change.name)
		}
		t.Logf("%s mutant: exit 0, empty stderr, byte oracle catches byte %d", change.name, firstDifference(got.stdout, truth.stdout))
		os.Remove(mutant)
	}
	overlayArguments := h.overlay("type-arguments-mutant", "bridge/tsgo/checker/type_arguments.go", "arguments = append(arguments, g.add(argument))", "arguments = append(arguments, g.add(argument)); arguments = arguments[:0]")
	argumentsArchive := h.archive("type-arguments-mutant", overlayArguments, false)
	argumentsMutant := h.build(stage0, "arguments-mutant", entry, argumentsArchive, false)
	changed := h.must("arguments-mutant-run", exec.Command(argumentsMutant, config, manifest))
	if len(changed.stderr) != 0 || bytes.Equal(changed.stdout, truth.stdout) {
		t.Fatal("type arguments mutant survived")
	}
	t.Logf("type-arguments mutant: exit 0, empty stderr, byte oracle catches byte %d", firstDifference(changed.stdout, truth.stdout))
	os.Remove(argumentsArchive)
	os.Remove(argumentsMutant)
	for _, corpus := range []struct{ name, config, manifest string }{
		{"repository", filepath.Join(repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE26_REPOSITORY_MANIFEST")},
		{"compiler", filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), os.Getenv("ADAMIC_WAVE26_COMPILER_MANIFEST")},
	} {
		if corpus.manifest == "" {
			continue
		}
		h.compare(corpus.name, oracle, binary, corpus.config, corpus.manifest)
		h.compare(corpus.name+"-asan", oracle, asan, corpus.config, corpus.manifest)
		want := h.must(corpus.name+"-timed-go", exec.Command(oracle, corpus.config, corpus.manifest))
		native := exec.Command(binary, corpus.config, corpus.manifest)
		native.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
		got := h.must(corpus.name+"-timed-native", native)
		if !bytes.Equal(want.stdout, got.stdout) {
			t.Fatal("timed output mismatch")
		}
		t.Logf("%s whole process native %s Go %s; native phases %s; Go phases %s", corpus.name, got.elapsed, want.elapsed, got.stderr, want.stderr)
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);console.log(tsgoInspect(program,file,0,1,'Identifier','raw-shape'));tsgoRelease(program);
console.log(tsgoInspect(program,file,0,1,'Identifier','type-arguments\n1'));`)
	probe := h.write("probe.a", "x;\n")
	stale := h.build(stage0, "released", released, archive, false)
	got := h.run("released-run", exec.Command(stale, config, probe))
	if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
		t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
	}
	t.Log("released handle: panic 70, invalid or released checker handle")
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains the released handle.")
	mutantArchive := h.archive("released-registry", overlay, false)
	mutant := h.build(stage0, "released-mutant", released, mutantArchive, false)
	h.must("released-mutant-run", exec.Command(mutant, config, probe))
	t.Log("released-registry mutant: exit 0, required panic 70 catches it")
}
