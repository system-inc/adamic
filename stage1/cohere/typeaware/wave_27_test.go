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

func wave27Controls() []string {
	preamble := ""
	for _, name := range []string{"OrmColumn", "TimestampColumn", "OrmJoinColumn", "OrmManyToOne", "OrmOneToMany", "OrmOneToOne", "SerializableField", "VerifyIsString", "VerifyIsNumber", "VerifyIsOptional", "VerifyBy", "VerifyIsArray", "VerifyIsNotEmpty", "VerifyArrayMinimumSize", "VerifySomethingCustom", "Other"} {
		preamble += "declare function " + name + "(...args:any[]):any;\n"
	}
	var controls []string
	add := func(body string) { controls = append(controls, preamble+body+"\nexport {};\n") }
	for _, decorator := range []string{"OrmColumn", "TimestampColumn", "OrmJoinColumn"} {
		for _, typ := range []string{"string", "string|null", "string|undefined", "any", "unknown", "void", "never"} {
			add("class A { @" + decorator + "({nullable:true}) x!: " + typ + "; }")
		}
	}
	for _, body := range []string{
		"@OrmColumn({nullable:false}) x!:string;", "@OrmColumn() x!:string;", "@OrmColumn({nullable:'items'}) x!:string;", "@OrmColumn({nullable:true}) x?:string;",
		"@OrmColumn({nullable:true}) @OrmManyToOne() x!:string;", "@OrmColumn({nullable:true}) @OrmOneToMany() x!:string;", "@OrmColumn({nullable:true}) @OrmOneToOne() x!:string;",
		"@OrmColumn({nullable:true}) ['x']!:string;", "@OrmColumn({nullable:true}) 'x'!:string;", "@OrmColumn({nullable:true}) #x!:string;", "@OrmColumn({nullable:true}) accessor x:string='a';",
		"@OrmColumn({nullable:true}) m():string {return 'a';}", "@OrmColumn({nullable:true},{nullable:false}) x!:string;", "@OrmColumn({'nullable':true}) x!:string;",
		"@OrmColumn({nullable:false,nullable:true}) x!:string;", "@OrmColumn({nullable:true}) @OrmColumn({nullable:true}) x!:string;",
	} {
		add("class A {" + body + "}")
	}
	for _, option := range []string{"", "{optional:true}", "{optional:false}", "{optional:'items'}", "{optional:true,defaultValue:undefined}", "{optional:false,defaultValue:false}"} {
		for _, typ := range []string{"string", "string|null", "any", "unknown", "void", "never"} {
			add("class A { @SerializableField(" + option + ") x!:" + typ + "; }")
		}
	}
	for _, typ := range []string{"string[]", "readonly string[]", "[string,number]", "readonly [string]", "Array<string>", "ReadonlyArray<string>", "string[]|null", "string[]|undefined", "any", "unknown", "Set<string>", "string", "never"} {
		add("class A { @VerifyIsString() x!:" + typ + "; }")
	}
	for _, decorator := range []string{"@VerifyIsArray()", "@VerifyArrayMinimumSize(1)", "@VerifyIsNotEmpty()", "@VerifySomethingCustom()", "@VerifyIsOptional()", "@VerifyBy(()=>true)", "@Other()"} {
		add("class A { " + decorator + " @VerifyIsString() x!:string[]; }")
	}
	for _, body := range []string{
		"@VerifyIsString x!:string[];", "@VerifyIsOptional() x!:string[];", "@Other() x!:string[];", "@VerifyIsString() ['x']!:string[];",
		"constructor(@VerifyIsString() public x:string[]) {}", "constructor(@VerifyIsString() readonly x:string[]) {}", "constructor(@VerifyIsString() private x:string[]) {}", "constructor(@VerifyIsString() protected x:string[]) {}", "constructor(@VerifyIsString() x:string[]) {}",
	} {
		add("class A {" + body + "}")
	}
	add("type Alias=readonly string[]|null; class A {@VerifyIsString() x!:Alias;}")
	add("namespace N {export function OrmColumn(...a:any[]):any {return null;} export function SerializableField(...a:any[]):any{return null;} export function VerifyIsString(...a:any[]):any{return null;}} class A {@N.OrmColumn({nullable:true}) x!:string; @N.SerializableField() y?:string; @N.VerifyIsString() z!:string[];}")
	add("/* 世界 🌍 */\r\nclass A {@OrmColumn({nullable:true}) é!:string; @SerializableField() 漢?:string; @VerifyIsString() 名!:string[];}\r\n")
	return controls
}

// Not parallel: builds, sanitizer subprocesses and timings share the machine.
func TestWave27AgreementAndMutants(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE27_ARTIFACTS")
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
	entry := filepath.Join(repository, "stage1/cohere/typeaware/wave_27_suite.a")
	binary := h.build(stage0, "wave27", entry, archive, false)
	oracle := volumeOracle(h, "wave27-oracle", "oracle_wave_27.go")
	h.write("control.d.ts", "export {};\n")
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ES2022","module":"NodeNext","lib":["ES2022"],"experimentalDecorators":true,"noEmit":true},"files":["control.d.ts"]}`)
	var paths []string
	for i, source := range wave27Controls() {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.a", i), source))
	}
	manifest := h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
	truth := h.compare("controls", oracle, binary, config, manifest)
	for _, name := range []string{"orm-column-nullable-parity", "serializable-nullable-parity", "verify-array-parity"} {
		if !bytes.Contains(truth.stdout, []byte("\tbase/correctness-require-"+name+"\t")) {
			t.Fatalf("no positive control for %s", name)
		}
	}
	sanitized := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "wave27-asan", entry, sanitized, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, change := range []struct{ name, file, from, to string }{
		{"orm", "correctness_require_orm_column_nullable_parity.a", "!facts.present || decoratorFacts.nullable(facts)", "!facts.present || !decoratorFacts.nullable(facts)"},
		{"serializable", "correctness_require_serializable_nullable_parity.a", "option === 2 && !nullable", "option === 2 && nullable"},
		{"array", "correctness_require_verify_array_parity.a", "!knownValue || suppressed", "!knownValue || !suppressed"},
	} {
		scratch := filepath.Join(directory, change.name+"-source")
		if err := os.MkdirAll(scratch, 0755); err != nil {
			t.Fatal(err)
		}
		files, err := filepath.Glob(filepath.Join(repository, "stage1/cohere/typeaware/*.a"))
		if err != nil {
			t.Fatal(err)
		}
		for _, path := range files {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			source := string(data)
			if filepath.Base(path) == change.file {
				if strings.Count(source, change.from) != 1 {
					t.Fatalf("nonunique mutant %s", change.name)
				}
				source = strings.Replace(source, change.from, change.to, 1)
			}
			// Keep existing dependencies in their original directory. New .a modules
			// stay together so exactly one rule decision changes.
			for _, dependency := range []string{"rules.ts", "facts.ts", "types.ts", "diagnostic.ts", "unary_minus.ts"} {
				source = strings.ReplaceAll(source, "'./"+dependency+"'", "'"+filepath.Join(repository, "stage1/cohere/typeaware", dependency)+"'")
			}
			source = strings.ReplaceAll(source, "'../../typescript/", "'"+filepath.Join(repository, "stage1/typescript")+"/")
			hpath := filepath.Join(scratch, filepath.Base(path))
			if err := os.WriteFile(hpath, []byte(source), 0644); err != nil {
				t.Fatal(err)
			}
		}
		mutant := h.build(stage0, change.name+"-mutant", filepath.Join(scratch, "wave_27_suite.a"), archive, false)
		got := h.must(change.name+"-run", exec.Command(mutant, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth.stdout) {
			t.Fatalf("%s mutant survived", change.name)
		}
		t.Logf("%s mutant: exit 0, empty stderr, Go byte oracle catches byte %d", change.name, firstDifference(got.stdout, truth.stdout))
	}
	for _, population := range []struct{ name, config, manifest string }{
		{"repository", filepath.Join(repository, "tsconfig.json"), os.Getenv("ADAMIC_WAVE27_REPOSITORY_MANIFEST")},
		{"compiler", filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), os.Getenv("ADAMIC_WAVE27_COMPILER_MANIFEST")},
	} {
		if population.manifest == "" {
			t.Logf("%s corpus not supplied", population.name)
			continue
		}
		h.compare(population.name, oracle, binary, population.config, population.manifest)
		h.compare(population.name+"-asan", oracle, asan, population.config, population.manifest)
		goRun := h.must(population.name+"-timed-go", exec.Command(oracle, population.config, population.manifest))
		nativeCmd := exec.Command(binary, population.config, population.manifest)
		nativeCmd.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
		nativeRun := h.must(population.name+"-timed-native", nativeCmd)
		if !bytes.Equal(goRun.stdout, nativeRun.stdout) {
			t.Fatal("timed bytes differ")
		}
		t.Logf("%s process native=%s Go=%s; native %s Go %s", population.name, nativeRun.elapsed, goRun.elapsed, nativeRun.stderr, goRun.stderr)
	}
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);
console.log(tsgoInspect(program,file,0,1,'Identifier','raw-type'));
`)
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
	t.Log("released-registry mutant exits 0, caught by required panic 70")
}
