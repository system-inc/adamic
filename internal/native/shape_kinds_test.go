package native

import (
	"context"
	"fmt"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/ir"
	regex "github.com/system-inc/adamic/internal/regexp"
)

func TestShapeKindsDistinguishScalarLayouts(test *testing.T) {
	test.Parallel()
	e := &emitter{program: &ir.Program{}}
	number := e.shapeOf([]string{"value"}, []ir.Type{ir.Number})
	boolean := e.shapeOf([]string{"value"}, []ir.Type{ir.Boolean})
	reference := e.shapeOf([]string{"value"}, []ir.Type{ir.String})
	if number == boolean || number == reference || boolean == reference {
		test.Fatal("different storage kinds share a shape")
	}
	if e.shapeOf([]string{"value"}, []ir.Type{ir.Number}) != number {
		test.Fatal("identical layouts do not share a shape")
	}
	if e.shapeOf([]string{"value"}, []ir.Type{ir.MaybeBoolean}) != number {
		test.Fatal("packed optional scalar must use number storage")
	}
}

// The probes allocate from the producers' actual descriptor declarations. Copying
// one descriptor to change its kinds leaves its ownership bitmap intact. Neither
// a second metadata table nor a mutant rejected by the C compiler proves this guard.
var shapeProducers = []struct{ name, path string }{
	{"CompiledRegExpOwner", "runtime/regexp_compile_ownership.c"},
	{"JSONDecoder", "runtime/json_decode.c"},
	{"PrototypeProbe", "prototype_contract_probe_test.go"},
	{"ConcurrentShapeRegistry", "runtime_statics_parallel_test.go"},
	{"JSONRecordProbe", "testdata/json-record-runtime/harness.c"},
	{"PrimitiveProbe", "testdata/to-primitive/runtime.c"},
	{"TupleViewProbe", "fx7_view_runtime_test.go"},
	{"JSONEncoderProbe", "../oracle/testdata/json_encoder_contract/runtime.c"},
	{"AsyncFrameObjects", ""},
	{"WASIRecords", ""},
	{"NodeFSDirectory", "runtime/node_fs_directory.c"},
	{"Directory", "runtime/directory.c"},
	{"Input", "runtime/input.c"},
	{"Exceptions", "runtime/exceptions.c"},
	{"AsyncHost", "runtime/async_host_impl.h"},
	{"Record", "runtime/record.c"},
	{"RegExp", "runtime/regexp.c"},
	{"TSGo", "runtime/tsgo.c"},
	{"NodeProcess", "runtime/node_process.c"},
	{"ProcessStatus", "runtime/process.c"},
	{"NodeFS", "runtime/node_fs_file.c"},
	{"NodeCrypto", "runtime/node_crypto.c"},
	{"Reflection", "runtime/library_object.c"},
	{"Iterators", "runtime/map_set.c"},
	{"CaptureGroups", ""},
	{"Emitted", ""},
	{"Graph", "graph_regions_test.go"},
	{"GraphAdoption", "graph_concurrency_test.go"},
	{"Region", "region_end_test.go"},
	{"EnvironmentMixed", "environment_placement_runtime_test.go"},
	{"Program", "testdata/program_region/core.c"},
	{"FieldCache", "testdata/runtime-statics/field_cache.c"},
	{"ShapeRegistry", "testdata/runtime-statics/shape_types.c"},
	{"JSON", "testdata/runtime-statics/json.c"},
	{"EmptyWeak", "testdata/runtime-statics/heap_weak.c"},
	{"ParallelMemory", "testdata/parallel/memory.c"},
	{"ParallelLifecycle", "testdata/parallel/lifecycle.c"},
	{"ParallelCache", "testdata/parallel/cache_pairs.c"},
	{"ParallelMap", "testdata/parallel/map.c"},
	{"EmptyMapHash", "testdata/map_hash.c"},
	{"RegExpResult", "regexp_test.go"},
	{"CheckedView", "../oracle/checked_views_v2_migration_test.go"},
}

type shapeDescriptor struct {
	code  string
	count int
}

func producerSource(test *testing.T, name, path string) string {
	test.Helper()
	if name == "AsyncFrameObjects" || name == "WASIRecords" {
		text := "async function inspect(): Promise<void> { const state = {value:'held',flag:true}; await Promise.resolve(); console.log(`${state.value}:${state.flag}`); } await inspect();"
		if name == "WASIRecords" {
			text = "export function echo(value: {n:number,b:boolean,s:string}): {n:number,b:boolean,s:string} {return value;}"
		}
		path := filepath.Join(test.TempDir(), "shape.a")
		if err := os.WriteFile(path, []byte(text), 0644); err != nil {
			test.Fatal(err)
		}
		source, err := load.Load([]string{path})
		if err != nil {
			test.Fatal(err)
		}
		program, err := lower.LowerWithOptions(context.Background(), source, lower.Options{})
		if err != nil {
			test.Fatal(err)
		}
		if name == "AsyncFrameObjects" {
			code := C(program)
			if !strings.Contains(code, "adamic_generated_frame_") ||
				!strings.Contains(code, "adamic_async_frame base;") {
				test.Fatal("missing ordinary async frame")
			}
			// Frames use cells and a child walker, not an object descriptor. Their held objects still use shapes.
			return code
		}
		record := ABIType{Kind: "record", Fields: []ABIField{
			{Name: "n", Type: ABIType{Kind: "number"}},
			{Name: "b", Type: ABIType{Kind: "boolean"}},
			{Name: "s", Type: ABIType{Kind: "string"}},
		}}
		index := -1
		for i, f := range program.Functions {
			if f.Name == "echo" {
				index = i
			}
		}
		code, err := WASIExports(program, []ABIExport{
			{Name: "echo", Function: index, Parameters: []ABIField{{Name: "value", Type: record}}, Returns: record},
		})
		if err != nil {
			test.Fatal(err)
		}
		return code
	}
	if name == "CaptureGroups" {
		program, err := regex.Compile("(?<word>a)(?<tail>b)", "d")
		if err != nil {
			test.Fatal(err)
		}
		code, err := program.NativeDeclarations("capture")
		if err != nil {
			test.Fatal(err)
		}
		return code
	}
	if name == "Emitted" {
		e := &emitter{program: &ir.Program{}}
		for _, kind := range []ir.Type{ir.Number, ir.Boolean, ir.String, ir.MaybeNumber, ir.MaybeBoolean} {
			e.shapeOf([]string{"value"}, []ir.Type{kind})
		}
		e.shapeOf(nil, nil)
		return strings.Join(e.declarations, "\n")
	}
	code, err := os.ReadFile(path)
	if err != nil {
		test.Fatal(err)
	}
	return string(code)
}

func producerDescriptors(test *testing.T, source string) []shapeDescriptor {
	test.Helper()
	// Shape arrays are expanded into their individual initializers. Local statics
	// use the nearest preceding declaration, preserving their original C scope.
	pattern := regexp.MustCompile(
		`(?s)static const adamic_shape \w+(?:\[[^]]*\])?(?: __attribute__\(\(unused\)\))?\s*=\s*\{([^;]*)\};`,
	)
	assigned := regexp.MustCompile(`\(adamic_shape\)\s*\{([^{}]*)\}`)
	type initializer struct {
		fields string
		at     int
	}
	var inputs []initializer
	for _, match := range pattern.FindAllStringSubmatchIndex(source, -1) {
		fields := source[match[2]:match[3]]
		if strings.Contains(fields, "{") {
			for _, inner := range regexp.MustCompile(`\{([^{}]*)\}`).FindAllStringSubmatch(fields, -1) {
				inputs = append(inputs, initializer{inner[1], match[0]})
			}
		} else {
			inputs = append(inputs, initializer{fields, match[0]})
		}
	}
	for _, match := range assigned.FindAllStringSubmatchIndex(source, -1) {
		inputs = append(inputs, initializer{source[match[2]:match[3]], match[0]})
	}
	var result []shapeDescriptor
	for _, input := range inputs {
		fields := strings.Split(input.fields, ",")
		if len(fields) != 5 {
			test.Fatalf("shape must have five fields: %s", input.fields)
		}
		count, err := strconv.Atoi(strings.TrimSpace(fields[0]))
		if err != nil {
			test.Fatalf("unrecognized count: %s", input.fields)
		}
		var code strings.Builder
		for _, position := range []int{1, 2, 4} {
			value := strings.TrimSpace(fields[position])
			if value == "NULL" {
				if position == 4 && count != 0 {
					test.Fatalf("missing kinds for %d fields: %s", count, input.fields)
				}
				continue
			}
			root := regexp.MustCompile(`^\w+`).FindString(value)
			declaration := regexp.MustCompile(`(?s)static const (?:char \*const|bool|adamic_field_kind) ` +
				regexp.QuoteMeta(root) +
				`\[[^;=]*\]\s*=\s*\{[^;]*\};`)
			matches := declaration.FindAllString(source[:input.at], -1)
			if len(matches) == 0 {
				test.Fatalf("missing %s for %s", value, input.fields)
			}
			if position == 4 {
				// Check explicit entries too: C zero-fills a missing numeric kind.
				rows := regexp.MustCompile(`\{([^{}]*)\}`).FindAllStringSubmatch(matches[len(matches)-1], -1)
				row := 0
				if index := regexp.MustCompile(`\[(\d+)\]$`).FindStringSubmatch(value); index != nil {
					row, _ = strconv.Atoi(index[1])
				}
				if row >= len(rows) {
					test.Fatalf("missing kind row %s", value)
				}
				entries := strings.Trim(strings.TrimSpace(rows[row][1]), ",")
				number := 0
				if entries != "" {
					number = len(strings.Split(entries, ","))
				}
				if number < count {
					test.Fatalf("%s lists %d explicit kinds for %d fields", value, number, count)
				}
			}
			code.WriteString(matches[len(matches)-1])
			code.WriteByte('\n')
		}
		// Method dispatch is independent of storage and ownership metadata.
		fields[3] = "NULL"
		fmt.Fprintf(&code, "adamic_shape shape = {%s};\n", strings.Join(fields, ","))
		result = append(result, shapeDescriptor{code.String(), count})
	}
	if len(result) == 0 {
		test.Fatal("producer has no recognized descriptors")
	}
	return result
}

func shapeProbeProduct(test *testing.T) string {
	test.Helper()
	var code strings.Builder
	code.WriteString(
		"#include \"adamic.h\"\n#include <string.h>\n_Static_assert(sizeof(adamic_field_kind)==1, \"one byte per slot\");\nint main(int argc,char **argv) {\nadamic_start(argc,argv);\nif(argc!=4) return 2;\nint index=atoi(argv[2]); int mutation=atoi(argv[3]);\n",
	)
	inputs := buildcache.Inputs{Name: "shape-kind-producers", Files: []string{
		"internal/native/runtime",
		"internal/native/shape_kinds_test.go",
		"internal/native/emit_objects.go",
		"internal/regexp/native.go",
	}}
	for _, producer := range shapeProducers {
		if producer.path != "" {
			inputs.Files = append(inputs.Files, filepath.ToSlash(filepath.Join("internal/native", producer.path)))
		}
		for index, descriptor := range producerDescriptors(test, producerSource(test, producer.name, producer.path)) {
			fmt.Fprintf(
				&code,
				"if(strcmp(argv[1],%s)==0 && index==%d) {\n%s",
				strconv.Quote(producer.name),
				index,
				descriptor.code,
			)
			if descriptor.count > 0 {
				code.WriteString(
					"adamic_field_kind changed[shape.count];\nmemcpy(changed,shape.kinds,sizeof changed);\nif(mutation==1 || mutation==5 || mutation==7) { changed[0]=shape.references[0]?adamic_field_number:adamic_field_reference; shape.kinds=changed; }\nif(mutation==2) shape.kinds=NULL;\nif(mutation==3) { changed[0]=255; shape.kinds=changed; }\nif(mutation>=100) { size_t field=(size_t)(mutation-100); if(field>=shape.count) return 5; changed[field]=shape.references[field]?adamic_field_number:adamic_field_reference; shape.kinds=changed; }\n",
				)
			} else {
				code.WriteString("if(mutation) return 3;\n")
			}
			code.WriteString(
				"if(mutation>=4 && mutation<=7) { adamic_region region={0}; adamic_object *object=mutation<6 ? adamic_object_new_in(&region,&shape) : adamic_object_new_filled_in(&region,&shape); memset(object->slots,0,shape.count*sizeof(adamic_value)); adamic_region_end(&region); } else { adamic_object *object=adamic_object_new(&shape); adamic_release(object); } return 0;\n}\n",
			)
		}
	}
	code.WriteString("return 4;\n}\n")
	source := code.String()
	request, err := RuntimeRequestFor(Options{Count: true})
	if err != nil {
		test.Fatal(err)
	}
	inputs.Flags = append([]string{source}, request.Portable...)
	inputs.Files = append(inputs.Files, "internal/native/testdata/parallel/scaling.c")
	inputs.Toolchain = []string{
		buildcache.Tool(request.Compiler, "--version"),
		buildcache.Tool(request.Archiver, "--version"),
	}
	inputs.Files = append(
		inputs.Files,
		"internal/native/native.go",
		"internal/native/library.go",
		"internal/native/target.go",
	)
	return buildcache.Product(test, inputs, func(directory string) error {
		if err := Build(source, filepath.Join(directory, "probe"), Options{Count: true}); err != nil {
			return err
		}
		scaling, err := os.ReadFile("testdata/parallel/scaling.c")
		if err != nil {
			return err
		}
		if err := Build(string(scaling), filepath.Join(directory, "scaling"), Options{Count: true}); err != nil {
			return err
		}
		before := "adamic_object *object = adamic_object_new(&shape);"
		if strings.Count(string(scaling), before) != 1 {
			return fmt.Errorf("scaling mutation is stale")
		}
		changed := strings.Replace(string(scaling), before, "kinds[count-1]=adamic_field_reference;\n"+before, 1)
		if err := Build(changed, filepath.Join(directory, "scaling-mutant"), Options{Count: true}); err != nil {
			return err
		}
		for _, sanitize := range []bool{false, true} {
			name := "string"
			if sanitize {
				name += "-sanitized"
			}
			if err := Build(
				stringShapeWitness,
				filepath.Join(directory, name),
				Options{Count: true, Sanitize: sanitize},
			); err != nil {
				return err
			}
		}
		return nil
	})
}

func TestProduct_ShapeKindProducers(test *testing.T) { test.Parallel(); shapeProbeProduct(test) }

func checkShapeProducer(test *testing.T, name string) {
	test.Helper()
	binary := filepath.Join(shapeProbeProduct(test), "probe")
	var descriptors []shapeDescriptor
	for _, producer := range shapeProducers {
		if producer.name == name {
			descriptors = producerDescriptors(test, producerSource(test, name, producer.path))
		}
	}
	if len(descriptors) == 0 {
		test.Fatal("unknown producer", name)
	}
	for index, descriptor := range descriptors {
		control := inlineArrayRun(test, binary, name, strconv.Itoa(index), "0")
		if control.exit != 0 ||
			len(control.stdout) != 0 ||
			!strings.HasPrefix(string(control.stderr), "adamic: counts:") {
			test.Fatalf("%s descriptor %d control: %+v", name, index, control)
		}
		if descriptor.count == 0 {
			continue
		} // An empty layout has no ownership bit to contradict.
		mutations := []string{"1", "2", "3"}
		for field := 1; field < descriptor.count; field++ {
			mutations = append(mutations, strconv.Itoa(100+field))
		}
		for _, mutation := range mutations {
			mutant := inlineArrayRun(test, binary, name, strconv.Itoa(index), mutation)
			if mutant.exit == 0 ||
				!strings.Contains(string(mutant.stderr), "adamic: inconsistent shape kind for field") {
				test.Fatalf("%s descriptor %d mutation %s survived: %+v", name, index, mutation, mutant)
			}
		}
	}
	test.Logf("%s: %d actual descriptors checked, each nonempty bitmap mutant caught", name, len(descriptors))
}

func TestShapeKindsDirectory(test *testing.T) { test.Parallel(); checkShapeProducer(test, "Directory") }

func TestShapeKindsInput(test *testing.T) { test.Parallel(); checkShapeProducer(test, "Input") }

func TestShapeKindsExceptions(test *testing.T) {
	test.Parallel()
	checkShapeProducer(test, "Exceptions")
}

func TestShapeKindsAsyncHost(test *testing.T) { test.Parallel(); checkShapeProducer(test, "AsyncHost") }

func TestShapeKindsRecord(test *testing.T) { test.Parallel(); checkShapeProducer(test, "Record") }

func TestShapeKindsRegExp(test *testing.T) { test.Parallel(); checkShapeProducer(test, "RegExp") }

func TestShapeKindsTSGo(test *testing.T) { test.Parallel(); checkShapeProducer(test, "TSGo") }

func TestShapeKindsNodeProcess(test *testing.T) {
	test.Parallel()
	checkShapeProducer(test, "NodeProcess")
}

func TestShapeKindsNodeFS(test *testing.T) { test.Parallel(); checkShapeProducer(test, "NodeFS") }

func TestShapeKindsNodeCrypto(test *testing.T) {
	test.Parallel()
	checkShapeProducer(test, "NodeCrypto")
}

func TestShapeKindsReflection(test *testing.T) {
	test.Parallel()
	checkShapeProducer(test, "Reflection")
}

func TestShapeKindsIterators(test *testing.T) { test.Parallel(); checkShapeProducer(test, "Iterators") }

func TestShapeKindsCaptureGroups(test *testing.T) {
	test.Parallel()
	checkShapeProducer(test, "CaptureGroups")
}

func TestShapeKindsEmitted(test *testing.T) { test.Parallel(); checkShapeProducer(test, "Emitted") }

func TestShapeKindsGraph(test *testing.T) { test.Parallel(); checkShapeProducer(test, "Graph") }

func TestShapeKindsGraphAdoption(test *testing.T) {
	test.Parallel()
	checkShapeProducer(test, "GraphAdoption")
}

func TestShapeKindsRegion(test *testing.T) { test.Parallel(); checkShapeProducer(test, "Region") }

func TestShapeKindsProgram(test *testing.T) { test.Parallel(); checkShapeProducer(test, "Program") }

func TestShapeKindsFieldCache(test *testing.T) {
	test.Parallel()
	checkShapeProducer(test, "FieldCache")
}

func TestShapeKindsShapeRegistry(test *testing.T) {
	test.Parallel()
	checkShapeProducer(test, "ShapeRegistry")
}

func TestShapeKindsJSON(test *testing.T) { test.Parallel(); checkShapeProducer(test, "JSON") }

func TestShapeKindsEmptyWeak(test *testing.T) { test.Parallel(); checkShapeProducer(test, "EmptyWeak") }

func TestShapeKindsParallelMemory(test *testing.T) {
	test.Parallel()
	checkShapeProducer(test, "ParallelMemory")
}

func TestShapeKindsParallelLifecycle(test *testing.T) {
	test.Parallel()
	checkShapeProducer(test, "ParallelLifecycle")
}

func TestShapeKindsParallelCache(test *testing.T) {
	test.Parallel()
	checkShapeProducer(test, "ParallelCache")
}

func TestShapeKindsParallelMap(test *testing.T) {
	test.Parallel()
	checkShapeProducer(test, "ParallelMap")
}

func TestShapeKindsEmptyMapHash(test *testing.T) {
	test.Parallel()
	checkShapeProducer(test, "EmptyMapHash")
}

func TestShapeKindsRegExpResult(test *testing.T) {
	test.Parallel()
	checkShapeProducer(test, "RegExpResult")
}

func TestShapeKindsCheckedView(test *testing.T) {
	test.Parallel()
	checkShapeProducer(test, "CheckedView")
}

const stringShapeWitness = `#include "adamic.h"
#include <assert.h>
int main(int argc,char **argv) {
 adamic_start(argc,argv);
 static adamic_string text=ADAMIC_STRING("a😀b");
 adamic_object *iterator=adamic_string_iterator(&text);
 adamic_closure *next=iterator->slots[0].reference;
 adamic_object *state=next->cells[0]->value.reference;
 assert(state->shape->kinds[0]==adamic_field_reference && state->shape->kinds[1]==adamic_field_number);
 for(int i=0;i<4;i++) {
  adamic_object *result=adamic_closure_call(next,NULL,0).reference;
  assert(result->shape->kinds[0]==adamic_field_reference && result->shape->kinds[1]==adamic_field_boolean);
  if(result->slots[1].boolean) puts("done");
  else adamic_write_line(adamic_stdout,result->slots[0].reference);
  adamic_release(result);
 }
 adamic_release(iterator);
 return 0;
}
`

func TestShapeKindsParallelScaling(test *testing.T) {
	test.Parallel()
	directory := shapeProbeProduct(test)
	control := inlineArrayRun(test, filepath.Join(directory, "scaling"))
	if control.exit != 0 || string(control.stdout) != "scaling clean\n" {
		test.Fatalf("control: %+v", control)
	}
	mutant := inlineArrayRun(test, filepath.Join(directory, "scaling-mutant"))
	if mutant.exit == 0 || !strings.Contains(string(mutant.stderr), "inconsistent shape kind for field value") {
		test.Fatalf("dynamic bitmap mutant survived: %+v", mutant)
	}
}

func TestStringIteratorKindsAgreeWithNode(test *testing.T) {
	test.Parallel()
	node := inlineArrayRun(test, "node", "--eval", `for(const value of "a😀b") console.log(value); console.log("done")`)
	if node.exit != 0 || len(node.stderr) != 0 {
		test.Fatalf("Node: %+v", node)
	}
	directory := shapeProbeProduct(test)
	for _, name := range []string{"string", "string-sanitized"} {
		got := inlineArrayRun(test, filepath.Join(directory, name))
		if got.exit != 0 ||
			string(got.stdout) != string(node.stdout) ||
			!strings.HasPrefix(string(got.stderr), "adamic: counts:") {
			test.Fatalf("%s: %+v; Node %q", name, got, node.stdout)
		}
	}
}

func TestShapeProducerInventory(test *testing.T) {
	test.Parallel()
	known := map[string]bool{}
	for _, producer := range shapeProducers {
		known[producer.path] = true
	}
	entries, err := runtime.ReadDir("runtime")
	if err != nil {
		test.Fatal(err)
	}
	for _, entry := range entries {
		path := "runtime/" + entry.Name()
		data, err := runtime.ReadFile(path)
		if err != nil {
			test.Fatal(err)
		}
		hasShape := strings.Contains(string(data), "static const adamic_shape ") ||
			strings.Contains(string(data), "(adamic_shape){")
		if path == "runtime/regexp_compile_bytecode.c" || path == "runtime/regexp_compile_runtime.c" {
			// TestShapeKindsRuntimeCompiledGroups checks the actual dynamic descriptors.
			continue
		}
		if hasShape && !known[path] {
			test.Errorf("shape producer needs its own control and mutant: %s", path)
		}
	}
}

func TestShapeKindsRegionAllocators(test *testing.T) {
	test.Parallel()
	binary := filepath.Join(shapeProbeProduct(test), "probe")
	for _, mode := range []string{"4", "6"} {
		control := inlineArrayRun(test, binary, "Directory", "0", mode)
		if control.exit != 0 || !strings.HasPrefix(string(control.stderr), "adamic: counts:") {
			test.Fatalf("region control %s: %+v", mode, control)
		}
	}
	for _, mode := range []string{"5", "7"} {
		mutant := inlineArrayRun(test, binary, "Directory", "0", mode)
		if mutant.exit == 0 || !strings.Contains(string(mutant.stderr), "inconsistent shape kind for field") {
			test.Fatalf("region bitmap mutant %s survived: %+v", mode, mutant)
		}
	}
}

func TestShapeKindsCompiledRegExpOwner(test *testing.T) {
	test.Parallel()
	checkShapeProducer(test, "CompiledRegExpOwner")
}
func TestShapeKindsPrototypeProbe(test *testing.T) {
	test.Parallel()
	checkShapeProducer(test, "PrototypeProbe")
}
func TestShapeKindsConcurrentShapeRegistry(test *testing.T) {
	test.Parallel()
	checkShapeProducer(test, "ConcurrentShapeRegistry")
}
func TestShapeKindsJSONRecordProbe(test *testing.T) {
	test.Parallel()
	checkShapeProducer(test, "JSONRecordProbe")
}
func TestShapeKindsPrimitiveProbe(test *testing.T) {
	test.Parallel()
	checkShapeProducer(test, "PrimitiveProbe")
}
func TestShapeKindsTupleViewProbe(test *testing.T) {
	test.Parallel()
	checkShapeProducer(test, "TupleViewProbe")
}
func TestShapeKindsJSONEncoderProbe(test *testing.T) {
	test.Parallel()
	checkShapeProducer(test, "JSONEncoderProbe")
}
func TestShapeKindsAsyncFrameObjects(test *testing.T) {
	test.Parallel()
	checkShapeProducer(test, "AsyncFrameObjects")
}
func TestShapeKindsWASIRecords(test *testing.T) {
	test.Parallel()
	checkShapeProducer(test, "WASIRecords")
}

func TestShapeKindsRuntimeCompiledGroups(test *testing.T) {
	test.Parallel()
	code := constructorPreamble + `
#include <assert.h>
static void check(const adamic_shape *shape,size_t count) {
 assert(shape!=NULL && shape->count==count && shape->kinds!=NULL);
 for(size_t i=0;i<count;i++) assert(shape->kinds[i]==adamic_field_reference);
 adamic_shape_check(shape);
}
int main(int argc,char **argv) {
 adamic_start(argc,argv);
 const unsigned char pattern[]="(?<first>a)(?<last>b)";
 adamic_string *p=text((const char *)pattern,sizeof pattern-1);
 adamic_object *regex=adamic_regex_compile_new(p,NULL);adamic_release(p);
 assert(regex!=NULL && regex->shape->kinds[12]==adamic_field_reference);
 adamic_string *input=text("ab",2);
 adamic_array *match=adamic_regex_exec(regex,input);assert(match!=NULL);
 adamic_object *groups=adamic_regex_property(match,"groups").reference;
 check(groups->shape,3);
 adamic_release(match);adamic_release(input);adamic_release(regex);
 return 0;
}`
	binary := filepath.Join(test.TempDir(), "groups")
	if err := Build(code+"\n", binary, Options{Count: true}); err != nil {
		test.Fatal(err)
	}
	got := inlineArrayRun(test, binary)
	if got.exit != 0 || len(got.stdout) != 0 || !strings.HasPrefix(string(got.stderr), "adamic: counts:") {
		test.Fatalf("dynamic shape control: %+v", got)
	}
}

func TestShapeKindsEnvironmentMixed(test *testing.T) {
	test.Parallel()
	checkShapeProducer(test, "EnvironmentMixed")
}
func TestShapeKindsJSONDecoder(test *testing.T) {
	test.Parallel()
	checkShapeProducer(test, "JSONDecoder")
}

func TestShapeKindsNodeFSDirectory(test *testing.T) {
	test.Parallel()
	checkShapeProducer(test, "NodeFSDirectory")
}

func TestShapeKindsProcessStatus(test *testing.T) {
	test.Parallel()
	checkShapeProducer(test, "ProcessStatus")
}
