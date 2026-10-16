package native

import (
	"fmt"
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

func TestShapeKindsDistinguishScalarLayouts(t *testing.T) {
	t.Parallel()
	e := &emitter{program: &ir.Program{}}
	number := e.shapeOf([]string{"value"}, []ir.Type{ir.Number})
	boolean := e.shapeOf([]string{"value"}, []ir.Type{ir.Boolean})
	reference := e.shapeOf([]string{"value"}, []ir.Type{ir.String})
	if number == boolean || number == reference || boolean == reference {
		t.Fatal("different storage kinds share a shape")
	}
	if e.shapeOf([]string{"value"}, []ir.Type{ir.Number}) != number {
		t.Fatal("identical layouts do not share a shape")
	}
	if e.shapeOf([]string{"value"}, []ir.Type{ir.MaybeBoolean}) != number {
		t.Fatal("packed optional scalar must use number storage")
	}
}

// The probes allocate from the producers' actual descriptor declarations. Copying
// one descriptor to change its kinds leaves its ownership bitmap intact. Neither
// a second metadata table nor a mutant rejected by the C compiler proves this guard.
var shapeProducers = []struct{ name, path string }{
	{"Directory", "runtime/directory.c"}, {"Input", "runtime/input.c"},
	{"Exceptions", "runtime/exceptions.c"}, {"AsyncHost", "runtime/async_host_impl.h"},
	{"Record", "runtime/record.c"}, {"RegExp", "runtime/regexp.c"},
	{"TSGo", "runtime/tsgo.c"}, {"NodeHost", "runtime/node_host.c"},
	{"NodeFS", "runtime/node_fs_file.c"}, {"NodeCrypto", "runtime/node_crypto.c"},
	{"Reflection", "runtime/library_object.c"}, {"Iterators", "runtime/map_set.c"},
	{"CaptureGroups", ""}, {"Emitted", ""},
	{"Graph", "graph_regions_test.go"}, {"GraphAdoption", "graph_concurrency_test.go"},
	{"Region", "region_end_test.go"}, {"Program", "testdata/program_region/core.c"},
	{"FieldCache", "testdata/runtime-statics/field_cache.c"},
	{"ShapeRegistry", "testdata/runtime-statics/shape_types.c"},
	{"JSON", "testdata/runtime-statics/json.c"}, {"EmptyWeak", "testdata/runtime-statics/heap_weak.c"},
	{"ParallelMemory", "testdata/parallel/memory.c"}, {"ParallelLifecycle", "testdata/parallel/lifecycle.c"},
	{"ParallelCache", "testdata/parallel/cache_pairs.c"}, {"ParallelMap", "testdata/parallel/map.c"},
	{"EmptyMapHash", "testdata/map_hash.c"}, {"RegExpResult", "regexp_test.go"},
	{"CheckedView", "../oracle/checked_views_v2_migration_test.go"},
}

type shapeDescriptor struct {
	code  string
	count int
}

func producerSource(t *testing.T, name, path string) string {
	t.Helper()
	if name == "CaptureGroups" {
		program, err := regex.Compile("(?<word>a)(?<tail>b)", "d")
		if err != nil {
			t.Fatal(err)
		}
		code, err := program.NativeDeclarations("capture")
		if err != nil {
			t.Fatal(err)
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
		t.Fatal(err)
	}
	return string(code)
}

func producerDescriptors(t *testing.T, source string) []shapeDescriptor {
	t.Helper()
	// Shape arrays are expanded into their individual initializers. Local statics
	// use the nearest preceding declaration, preserving their original C scope.
	pattern := regexp.MustCompile(`(?s)static const adamic_shape \w+(?:\[[^]]*\])?(?: __attribute__\(\(unused\)\))?\s*=\s*\{([^;]*)\};`)
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
			t.Fatalf("shape must have five fields: %s", input.fields)
		}
		count, err := strconv.Atoi(strings.TrimSpace(fields[0]))
		if err != nil {
			t.Fatalf("unrecognized count: %s", input.fields)
		}
		var code strings.Builder
		for _, position := range []int{1, 2, 4} {
			value := strings.TrimSpace(fields[position])
			if value == "NULL" {
				continue
			}
			root := regexp.MustCompile(`^\w+`).FindString(value)
			declaration := regexp.MustCompile(`(?s)static const (?:char \*const|bool|adamic_field_kind) ` + regexp.QuoteMeta(root) + `\[[^;=]*\]\s*=\s*\{[^;]*\};`)
			matches := declaration.FindAllString(source[:input.at], -1)
			if len(matches) == 0 {
				t.Fatalf("missing %s for %s", value, input.fields)
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
		t.Fatal("producer has no recognized descriptors")
	}
	return result
}

func shapeProbeProduct(t *testing.T) string {
	t.Helper()
	var code strings.Builder
	code.WriteString("#include \"adamic.h\"\n#include <string.h>\n_Static_assert(sizeof(adamic_field_kind)==1, \"one byte per slot\");\nint main(int argc,char **argv) {\nadamic_start(argc,argv);\nif(argc!=4) return 2;\nint index=atoi(argv[2]); int mutation=atoi(argv[3]);\n")
	inputs := buildcache.Inputs{Name: "shape-kind-producers", Files: []string{"internal/native/runtime", "internal/native/shape_kinds_test.go", "internal/native/emit_objects.go", "internal/regexp/native.go"}}
	for _, producer := range shapeProducers {
		if producer.path != "" {
			inputs.Files = append(inputs.Files, filepath.ToSlash(filepath.Join("internal/native", producer.path)))
		}
		for index, descriptor := range producerDescriptors(t, producerSource(t, producer.name, producer.path)) {
			fmt.Fprintf(&code, "if(strcmp(argv[1],%s)==0 && index==%d) {\n%s", strconv.Quote(producer.name), index, descriptor.code)
			if descriptor.count > 0 {
				code.WriteString("adamic_field_kind changed[shape.count];\nmemcpy(changed,shape.kinds,sizeof changed);\nif(mutation==1 || mutation==5 || mutation==7) { changed[0]=shape.references[0]?adamic_field_number:adamic_field_reference; shape.kinds=changed; }\nif(mutation==2) shape.kinds=NULL;\nif(mutation==3) { changed[0]=255; shape.kinds=changed; }\n")
			} else {
				code.WriteString("if(mutation) return 3;\n")
			}
			code.WriteString("if(mutation>=4) { adamic_region region={0}; adamic_object *object=mutation<6 ? adamic_object_new_in(&region,&shape) : adamic_object_new_filled_in(&region,&shape); memset(object->slots,0,shape.count*sizeof(adamic_value)); adamic_region_end(&region); } else { adamic_object *object=adamic_object_new(&shape); adamic_release(object); } return 0;\n}\n")
		}
	}
	code.WriteString("return 4;\n}\n")
	source := code.String()
	request, err := RuntimeRequestFor(Options{Count: true})
	if err != nil {
		t.Fatal(err)
	}
	inputs.Flags = append([]string{source}, request.Portable...)
	inputs.Files = append(inputs.Files, "internal/native/testdata/parallel/scaling.c")
	inputs.Toolchain = []string{buildcache.Tool(request.Compiler, "--version"), buildcache.Tool(request.Archiver, "--version")}
	inputs.Files = append(inputs.Files, "internal/native/native.go", "internal/native/library.go", "internal/native/target.go")
	return buildcache.Product(t, inputs, func(directory string) error {
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
		changed := strings.Replace(string(scaling), before, "kinds[0]=adamic_field_reference;\n"+before, 1)
		if err := Build(changed, filepath.Join(directory, "scaling-mutant"), Options{Count: true}); err != nil {
			return err
		}
		for _, sanitize := range []bool{false, true} {
			name := "string"
			if sanitize {
				name += "-sanitized"
			}
			if err := Build(stringShapeWitness, filepath.Join(directory, name), Options{Count: true, Sanitize: sanitize}); err != nil {
				return err
			}
		}
		return nil
	})
}

func TestProduct_ShapeKindProducers(t *testing.T) { t.Parallel(); shapeProbeProduct(t) }

func checkShapeProducer(t *testing.T, name string) {
	t.Helper()
	binary := filepath.Join(shapeProbeProduct(t), "probe")
	var descriptors []shapeDescriptor
	for _, producer := range shapeProducers {
		if producer.name == name {
			descriptors = producerDescriptors(t, producerSource(t, name, producer.path))
		}
	}
	if len(descriptors) == 0 {
		t.Fatal("unknown producer", name)
	}
	for index, descriptor := range descriptors {
		control := inlineArrayRun(t, binary, name, strconv.Itoa(index), "0")
		if control.exit != 0 || len(control.stdout) != 0 || !strings.HasPrefix(string(control.stderr), "adamic: counts:") {
			t.Fatalf("%s descriptor %d control: %+v", name, index, control)
		}
		if descriptor.count == 0 {
			continue
		} // An empty layout has no ownership bit to contradict.
		for _, mutation := range []string{"1", "2", "3"} {
			mutant := inlineArrayRun(t, binary, name, strconv.Itoa(index), mutation)
			if mutant.exit == 0 || !strings.Contains(string(mutant.stderr), "adamic: inconsistent shape kind for field") {
				t.Fatalf("%s descriptor %d mutation %s survived: %+v", name, index, mutation, mutant)
			}
		}
	}
	t.Logf("%s: %d actual descriptors checked, each nonempty bitmap mutant caught", name, len(descriptors))
}

func TestShapeKindsDirectory(t *testing.T) { t.Parallel(); checkShapeProducer(t, "Directory") }

func TestShapeKindsInput(t *testing.T) { t.Parallel(); checkShapeProducer(t, "Input") }

func TestShapeKindsExceptions(t *testing.T) { t.Parallel(); checkShapeProducer(t, "Exceptions") }

func TestShapeKindsAsyncHost(t *testing.T) { t.Parallel(); checkShapeProducer(t, "AsyncHost") }

func TestShapeKindsRecord(t *testing.T) { t.Parallel(); checkShapeProducer(t, "Record") }

func TestShapeKindsRegExp(t *testing.T) { t.Parallel(); checkShapeProducer(t, "RegExp") }

func TestShapeKindsTSGo(t *testing.T) { t.Parallel(); checkShapeProducer(t, "TSGo") }

func TestShapeKindsNodeHost(t *testing.T) { t.Parallel(); checkShapeProducer(t, "NodeHost") }

func TestShapeKindsNodeFS(t *testing.T) { t.Parallel(); checkShapeProducer(t, "NodeFS") }

func TestShapeKindsNodeCrypto(t *testing.T) { t.Parallel(); checkShapeProducer(t, "NodeCrypto") }

func TestShapeKindsReflection(t *testing.T) { t.Parallel(); checkShapeProducer(t, "Reflection") }

func TestShapeKindsIterators(t *testing.T) { t.Parallel(); checkShapeProducer(t, "Iterators") }

func TestShapeKindsCaptureGroups(t *testing.T) { t.Parallel(); checkShapeProducer(t, "CaptureGroups") }

func TestShapeKindsEmitted(t *testing.T) { t.Parallel(); checkShapeProducer(t, "Emitted") }

func TestShapeKindsGraph(t *testing.T) { t.Parallel(); checkShapeProducer(t, "Graph") }

func TestShapeKindsGraphAdoption(t *testing.T) { t.Parallel(); checkShapeProducer(t, "GraphAdoption") }

func TestShapeKindsRegion(t *testing.T) { t.Parallel(); checkShapeProducer(t, "Region") }

func TestShapeKindsProgram(t *testing.T) { t.Parallel(); checkShapeProducer(t, "Program") }

func TestShapeKindsFieldCache(t *testing.T) { t.Parallel(); checkShapeProducer(t, "FieldCache") }

func TestShapeKindsShapeRegistry(t *testing.T) { t.Parallel(); checkShapeProducer(t, "ShapeRegistry") }

func TestShapeKindsJSON(t *testing.T) { t.Parallel(); checkShapeProducer(t, "JSON") }

func TestShapeKindsEmptyWeak(t *testing.T) { t.Parallel(); checkShapeProducer(t, "EmptyWeak") }

func TestShapeKindsParallelMemory(t *testing.T) {
	t.Parallel()
	checkShapeProducer(t, "ParallelMemory")
}

func TestShapeKindsParallelLifecycle(t *testing.T) {
	t.Parallel()
	checkShapeProducer(t, "ParallelLifecycle")
}

func TestShapeKindsParallelCache(t *testing.T) { t.Parallel(); checkShapeProducer(t, "ParallelCache") }

func TestShapeKindsParallelMap(t *testing.T) { t.Parallel(); checkShapeProducer(t, "ParallelMap") }

func TestShapeKindsEmptyMapHash(t *testing.T) { t.Parallel(); checkShapeProducer(t, "EmptyMapHash") }

func TestShapeKindsRegExpResult(t *testing.T) { t.Parallel(); checkShapeProducer(t, "RegExpResult") }

func TestShapeKindsCheckedView(t *testing.T) { t.Parallel(); checkShapeProducer(t, "CheckedView") }

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
  assert(result->shape->kinds[0]==adamic_field_boolean && result->shape->kinds[1]==adamic_field_reference);
  if(result->slots[0].boolean) puts("done");
  else adamic_write_line(adamic_stdout,result->slots[1].reference);
  adamic_release(result);
 }
 adamic_release(iterator);
 return 0;
}
`

func TestShapeKindsParallelScaling(t *testing.T) {
	t.Parallel()
	directory := shapeProbeProduct(t)
	control := inlineArrayRun(t, filepath.Join(directory, "scaling"))
	if control.exit != 0 || string(control.stdout) != "scaling clean\n" {
		t.Fatalf("control: %+v", control)
	}
	mutant := inlineArrayRun(t, filepath.Join(directory, "scaling-mutant"))
	if mutant.exit == 0 || !strings.Contains(string(mutant.stderr), "inconsistent shape kind for field padding") {
		t.Fatalf("dynamic bitmap mutant survived: %+v", mutant)
	}
}

func TestStringIteratorKindsAgreeWithNode(t *testing.T) {
	t.Parallel()
	node := inlineArrayRun(t, "node", "--eval", `for(const value of "a😀b") console.log(value); console.log("done")`)
	if node.exit != 0 || len(node.stderr) != 0 {
		t.Fatalf("Node: %+v", node)
	}
	directory := shapeProbeProduct(t)
	for _, name := range []string{"string", "string-sanitized"} {
		got := inlineArrayRun(t, filepath.Join(directory, name))
		if got.exit != 0 || string(got.stdout) != string(node.stdout) || !strings.HasPrefix(string(got.stderr), "adamic: counts:") {
			t.Fatalf("%s: %+v; Node %q", name, got, node.stdout)
		}
	}
}

func TestShapeProducerInventory(t *testing.T) {
	t.Parallel()
	known := map[string]bool{}
	for _, producer := range shapeProducers {
		known[producer.path] = true
	}
	entries, err := runtime.ReadDir("runtime")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		path := "runtime/" + entry.Name()
		data, err := runtime.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(data), "static const adamic_shape ") && !known[path] {
			t.Errorf("shape producer needs its own control and mutant: %s", path)
		}
	}
}

func TestShapeKindsRegionAllocators(t *testing.T) {
	t.Parallel()
	binary := filepath.Join(shapeProbeProduct(t), "probe")
	for _, mode := range []string{"4", "6"} {
		control := inlineArrayRun(t, binary, "Directory", "0", mode)
		if control.exit != 0 || !strings.HasPrefix(string(control.stderr), "adamic: counts:") {
			t.Fatalf("region control %s: %+v", mode, control)
		}
	}
	for _, mode := range []string{"5", "7"} {
		mutant := inlineArrayRun(t, binary, "Directory", "0", mode)
		if mutant.exit == 0 || !strings.Contains(string(mutant.stderr), "inconsistent shape kind for field") {
			t.Fatalf("region bitmap mutant %s survived: %+v", mode, mutant)
		}
	}
}
