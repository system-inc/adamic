// Each translation unit owns one bounded product and consumer.
package native

import "testing"

var gccRuntimeUnits = []string{
	"adamic.c",
	"array.c",
	"array_from.c",
	"async.c",
	"bitwise.c",
	"case.c",
	"census_small.c",
	"class_features.c",
	"class_inheritance.c",
	"class_static.c",
	"closure.c",
	"construction.c",
	"count.c",
	"directory.c",
	"dtoa.c",
	"exceptions.c",
	"from_codes.c",
	"generator.c",
	"graph_regions.c",
	"heap.c",
	"host_runtime.c",
	"hypot.c",
	"ieee754.c",
	"input.c",
	"json_decode.c",
	"json_encode.c",
	"json_stringify.c",
	"library_array.c",
	"library_language.c",
	"library_math_number.c",
	"library_object.c",
	"map.c",
	"map_set.c",
	"math.c",
	"maybe.c",
	"node_buffer.c",
	"node_crypto.c",
	"node_cwd_wasi.c",
	"node_fs_directory.c",
	"node_fs_file.c",
	"node_host.c",
	"node_path.c",
	"node_process.c",
	"normalize.c",
	"number.c",
	"object.c",
	"parallel.c",
	"parse.c",
	"process.c",
	"program_region.c",
	"radix.c",
	"record.c",
	"regexp.c",
	"regexp_compile_bytecode.c",
	"regexp_compile_ownership.c",
	"regexp_compile_parser.c",
	"regexp_compile_properties.c",
	"regexp_compile_runtime.c",
	"regexp_compile_sets.c",
	"regexp_compile_v8.c",
	"regexp_regular.c",
	"regexp_replace.c",
	"region.c",
	"set.c",
	"share.c",
	"slab_quarantine.c",
	"sort.c",
	"sort_undefined.c",
	"spread.c",
	"stack.c",
	"string.c",
	"string_append.c",
	"string_from.c",
	"string_index.c",
	"string_share.c",
	"tsgo.c",
	"typed_array.c",
	"typed_array_sort.c",
	"union.c",
	"utf8.c",
	"view_unions_mixed.c",
	"view_unions_untagged.c",
	"weak.c",
}

func TestProduct_GCCRuntimeAdamic(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "adamic.c") }
func TestGCCRuntimeAdamic(test *testing.T)         { test.Parallel(); checkGCCRuntimeUnit(test, "adamic.c") }

func TestProduct_GCCRuntimeArray(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "array.c") }
func TestGCCRuntimeArray(test *testing.T)         { test.Parallel(); checkGCCRuntimeUnit(test, "array.c") }

func TestProduct_GCCRuntimeArrayFrom(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "array_from.c")
}
func TestGCCRuntimeArrayFrom(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "array_from.c")
}

func TestProduct_GCCRuntimeAsync(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "async.c") }
func TestGCCRuntimeAsync(test *testing.T)         { test.Parallel(); checkGCCRuntimeUnit(test, "async.c") }

func TestProduct_GCCRuntimeBitwise(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "bitwise.c")
}
func TestGCCRuntimeBitwise(test *testing.T) { test.Parallel(); checkGCCRuntimeUnit(test, "bitwise.c") }

func TestProduct_GCCRuntimeCase(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "case.c") }
func TestGCCRuntimeCase(test *testing.T)         { test.Parallel(); checkGCCRuntimeUnit(test, "case.c") }

func TestProduct_GCCRuntimeCensusSmall(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "census_small.c")
}
func TestGCCRuntimeCensusSmall(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "census_small.c")
}

func TestProduct_GCCRuntimeClassFeatures(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "class_features.c")
}
func TestGCCRuntimeClassFeatures(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "class_features.c")
}

func TestProduct_GCCRuntimeClassInheritance(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "class_inheritance.c")
}
func TestGCCRuntimeClassInheritance(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "class_inheritance.c")
}

func TestProduct_GCCRuntimeClassStatic(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "class_static.c")
}
func TestGCCRuntimeClassStatic(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "class_static.c")
}

func TestProduct_GCCRuntimeClosure(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "closure.c")
}
func TestGCCRuntimeClosure(test *testing.T) { test.Parallel(); checkGCCRuntimeUnit(test, "closure.c") }

func TestProduct_GCCRuntimeConstruction(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "construction.c")
}
func TestGCCRuntimeConstruction(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "construction.c")
}

func TestProduct_GCCRuntimeCount(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "count.c") }
func TestGCCRuntimeCount(test *testing.T)         { test.Parallel(); checkGCCRuntimeUnit(test, "count.c") }

func TestProduct_GCCRuntimeDirectory(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "directory.c")
}
func TestGCCRuntimeDirectory(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "directory.c")
}

func TestProduct_GCCRuntimeDtoa(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "dtoa.c") }
func TestGCCRuntimeDtoa(test *testing.T)         { test.Parallel(); checkGCCRuntimeUnit(test, "dtoa.c") }

func TestProduct_GCCRuntimeExceptions(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "exceptions.c")
}
func TestGCCRuntimeExceptions(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "exceptions.c")
}

func TestProduct_GCCRuntimeFromCodes(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "from_codes.c")
}
func TestGCCRuntimeFromCodes(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "from_codes.c")
}

func TestProduct_GCCRuntimeGenerator(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "generator.c")
}
func TestGCCRuntimeGenerator(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "generator.c")
}

func TestProduct_GCCRuntimeGraphRegions(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "graph_regions.c")
}
func TestGCCRuntimeGraphRegions(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "graph_regions.c")
}

func TestProduct_GCCRuntimeHeap(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "heap.c") }
func TestGCCRuntimeHeap(test *testing.T)         { test.Parallel(); checkGCCRuntimeUnit(test, "heap.c") }

func TestProduct_GCCRuntimeHostRuntime(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "host_runtime.c")
}
func TestGCCRuntimeHostRuntime(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "host_runtime.c")
}

func TestProduct_GCCRuntimeHypot(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "hypot.c") }
func TestGCCRuntimeHypot(test *testing.T)         { test.Parallel(); checkGCCRuntimeUnit(test, "hypot.c") }

func TestProduct_GCCRuntimeIeee754(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "ieee754.c")
}
func TestGCCRuntimeIeee754(test *testing.T) { test.Parallel(); checkGCCRuntimeUnit(test, "ieee754.c") }

func TestProduct_GCCRuntimeInput(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "input.c") }
func TestGCCRuntimeInput(test *testing.T)         { test.Parallel(); checkGCCRuntimeUnit(test, "input.c") }

func TestProduct_GCCRuntimeJsonDecode(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "json_decode.c")
}
func TestGCCRuntimeJsonDecode(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "json_decode.c")
}

func TestProduct_GCCRuntimeJsonEncode(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "json_encode.c")
}
func TestGCCRuntimeJsonEncode(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "json_encode.c")
}

func TestProduct_GCCRuntimeJsonStringify(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "json_stringify.c")
}
func TestGCCRuntimeJsonStringify(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "json_stringify.c")
}

func TestProduct_GCCRuntimeLibraryArray(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "library_array.c")
}
func TestGCCRuntimeLibraryArray(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "library_array.c")
}

func TestProduct_GCCRuntimeLibraryLanguage(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "library_language.c")
}
func TestGCCRuntimeLibraryLanguage(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "library_language.c")
}

func TestProduct_GCCRuntimeLibraryMathNumber(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "library_math_number.c")
}
func TestGCCRuntimeLibraryMathNumber(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "library_math_number.c")
}

func TestProduct_GCCRuntimeLibraryObject(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "library_object.c")
}
func TestGCCRuntimeLibraryObject(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "library_object.c")
}

func TestProduct_GCCRuntimeMap(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "map.c") }
func TestGCCRuntimeMap(test *testing.T)         { test.Parallel(); checkGCCRuntimeUnit(test, "map.c") }

func TestProduct_GCCRuntimeMapSet(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "map_set.c")
}
func TestGCCRuntimeMapSet(test *testing.T) { test.Parallel(); checkGCCRuntimeUnit(test, "map_set.c") }

func TestProduct_GCCRuntimeMath(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "math.c") }
func TestGCCRuntimeMath(test *testing.T)         { test.Parallel(); checkGCCRuntimeUnit(test, "math.c") }

func TestProduct_GCCRuntimeMaybe(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "maybe.c") }
func TestGCCRuntimeMaybe(test *testing.T)         { test.Parallel(); checkGCCRuntimeUnit(test, "maybe.c") }

func TestProduct_GCCRuntimeNodeBuffer(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "node_buffer.c")
}
func TestGCCRuntimeNodeBuffer(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "node_buffer.c")
}

func TestProduct_GCCRuntimeNodeCrypto(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "node_crypto.c")
}
func TestGCCRuntimeNodeCrypto(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "node_crypto.c")
}

func TestProduct_GCCRuntimeNodeFsFile(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "node_fs_file.c")
}
func TestGCCRuntimeNodeFsFile(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "node_fs_file.c")
}

func TestProduct_GCCRuntimeNodeHost(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "node_host.c")
}
func TestGCCRuntimeNodeHost(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "node_host.c")
}

func TestProduct_GCCRuntimeNormalize(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "normalize.c")
}
func TestGCCRuntimeNormalize(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "normalize.c")
}

func TestProduct_GCCRuntimeNumber(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "number.c") }
func TestGCCRuntimeNumber(test *testing.T)         { test.Parallel(); checkGCCRuntimeUnit(test, "number.c") }

func TestProduct_GCCRuntimeObject(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "object.c") }
func TestGCCRuntimeObject(test *testing.T)         { test.Parallel(); checkGCCRuntimeUnit(test, "object.c") }

func TestProduct_GCCRuntimeParallel(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "parallel.c")
}
func TestGCCRuntimeParallel(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "parallel.c")
}

func TestProduct_GCCRuntimeParse(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "parse.c") }
func TestGCCRuntimeParse(test *testing.T)         { test.Parallel(); checkGCCRuntimeUnit(test, "parse.c") }

func TestProduct_GCCRuntimeProgramRegion(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "program_region.c")
}
func TestGCCRuntimeProgramRegion(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "program_region.c")
}

func TestProduct_GCCRuntimeRadix(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "radix.c") }
func TestGCCRuntimeRadix(test *testing.T)         { test.Parallel(); checkGCCRuntimeUnit(test, "radix.c") }

func TestProduct_GCCRuntimeRecord(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "record.c") }
func TestGCCRuntimeRecord(test *testing.T)         { test.Parallel(); checkGCCRuntimeUnit(test, "record.c") }

func TestProduct_GCCRuntimeRegexp(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "regexp.c") }
func TestGCCRuntimeRegexp(test *testing.T)         { test.Parallel(); checkGCCRuntimeUnit(test, "regexp.c") }

func TestProduct_GCCRuntimeRegexpCompileBytecode(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "regexp_compile_bytecode.c")
}
func TestGCCRuntimeRegexpCompileBytecode(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "regexp_compile_bytecode.c")
}

func TestProduct_GCCRuntimeRegexpCompileOwnership(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "regexp_compile_ownership.c")
}
func TestGCCRuntimeRegexpCompileOwnership(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "regexp_compile_ownership.c")
}

func TestProduct_GCCRuntimeRegexpCompileParser(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "regexp_compile_parser.c")
}
func TestGCCRuntimeRegexpCompileParser(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "regexp_compile_parser.c")
}

func TestProduct_GCCRuntimeRegexpCompileProperties(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "regexp_compile_properties.c")
}
func TestGCCRuntimeRegexpCompileProperties(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "regexp_compile_properties.c")
}

func TestProduct_GCCRuntimeRegexpCompileRuntime(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "regexp_compile_runtime.c")
}
func TestGCCRuntimeRegexpCompileRuntime(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "regexp_compile_runtime.c")
}

func TestProduct_GCCRuntimeRegexpCompileSets(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "regexp_compile_sets.c")
}
func TestGCCRuntimeRegexpCompileSets(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "regexp_compile_sets.c")
}

func TestProduct_GCCRuntimeRegexpCompileV8(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "regexp_compile_v8.c")
}
func TestGCCRuntimeRegexpCompileV8(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "regexp_compile_v8.c")
}

func TestProduct_GCCRuntimeRegexpRegular(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "regexp_regular.c")
}
func TestGCCRuntimeRegexpRegular(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "regexp_regular.c")
}

func TestProduct_GCCRuntimeRegexpReplace(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "regexp_replace.c")
}
func TestGCCRuntimeRegexpReplace(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "regexp_replace.c")
}

func TestProduct_GCCRuntimeRegion(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "region.c") }
func TestGCCRuntimeRegion(test *testing.T)         { test.Parallel(); checkGCCRuntimeUnit(test, "region.c") }

func TestProduct_GCCRuntimeSet(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "set.c") }
func TestGCCRuntimeSet(test *testing.T)         { test.Parallel(); checkGCCRuntimeUnit(test, "set.c") }

func TestProduct_GCCRuntimeShare(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "share.c") }
func TestGCCRuntimeShare(test *testing.T)         { test.Parallel(); checkGCCRuntimeUnit(test, "share.c") }

func TestProduct_GCCRuntimeSlabQuarantine(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "slab_quarantine.c")
}
func TestGCCRuntimeSlabQuarantine(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "slab_quarantine.c")
}

func TestProduct_GCCRuntimeSort(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "sort.c") }
func TestGCCRuntimeSort(test *testing.T)         { test.Parallel(); checkGCCRuntimeUnit(test, "sort.c") }

func TestProduct_GCCRuntimeSortUndefined(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "sort_undefined.c")
}
func TestGCCRuntimeSortUndefined(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "sort_undefined.c")
}

func TestProduct_GCCRuntimeSpread(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "spread.c") }
func TestGCCRuntimeSpread(test *testing.T)         { test.Parallel(); checkGCCRuntimeUnit(test, "spread.c") }

func TestProduct_GCCRuntimeStack(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "stack.c") }
func TestGCCRuntimeStack(test *testing.T)         { test.Parallel(); checkGCCRuntimeUnit(test, "stack.c") }

func TestProduct_GCCRuntimeString(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "string.c") }
func TestGCCRuntimeString(test *testing.T)         { test.Parallel(); checkGCCRuntimeUnit(test, "string.c") }

func TestProduct_GCCRuntimeStringAppend(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "string_append.c")
}
func TestGCCRuntimeStringAppend(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "string_append.c")
}

func TestProduct_GCCRuntimeStringFrom(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "string_from.c")
}
func TestGCCRuntimeStringFrom(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "string_from.c")
}

func TestProduct_GCCRuntimeStringIndex(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "string_index.c")
}
func TestGCCRuntimeStringIndex(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "string_index.c")
}

func TestProduct_GCCRuntimeStringShare(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "string_share.c")
}
func TestGCCRuntimeStringShare(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "string_share.c")
}

func TestProduct_GCCRuntimeTsgo(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "tsgo.c") }
func TestGCCRuntimeTsgo(test *testing.T)         { test.Parallel(); checkGCCRuntimeUnit(test, "tsgo.c") }

func TestProduct_GCCRuntimeTypedArray(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "typed_array.c")
}
func TestGCCRuntimeTypedArray(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "typed_array.c")
}

func TestProduct_GCCRuntimeTypedArraySort(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "typed_array_sort.c")
}
func TestGCCRuntimeTypedArraySort(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "typed_array_sort.c")
}

func TestProduct_GCCRuntimeUnion(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "union.c") }
func TestGCCRuntimeUnion(test *testing.T)         { test.Parallel(); checkGCCRuntimeUnit(test, "union.c") }

func TestProduct_GCCRuntimeUtf8(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "utf8.c") }
func TestGCCRuntimeUtf8(test *testing.T)         { test.Parallel(); checkGCCRuntimeUnit(test, "utf8.c") }

func TestProduct_GCCRuntimeViewUnionsMixed(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "view_unions_mixed.c")
}
func TestGCCRuntimeViewUnionsMixed(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "view_unions_mixed.c")
}

func TestProduct_GCCRuntimeViewUnionsUntagged(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "view_unions_untagged.c")
}
func TestGCCRuntimeViewUnionsUntagged(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "view_unions_untagged.c")
}

func TestProduct_GCCRuntimeWeak(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "weak.c") }
func TestGCCRuntimeWeak(test *testing.T)         { test.Parallel(); checkGCCRuntimeUnit(test, "weak.c") }

func TestProduct_GCCRuntimeNodeCwdWasi(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "node_cwd_wasi.c")
}
func TestGCCRuntimeNodeCwdWasi(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "node_cwd_wasi.c")
}

func TestProduct_GCCRuntimeNodeFSDirectory(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "node_fs_directory.c")
}
func TestGCCRuntimeNodeFSDirectory(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "node_fs_directory.c")
}

func TestProduct_GCCRuntimeNodePath(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "node_path.c")
}
func TestGCCRuntimeNodePath(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "node_path.c")
}

func TestProduct_GCCRuntimeNodeProcess(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "node_process.c")
}
func TestGCCRuntimeNodeProcess(test *testing.T) {
	test.Parallel()
	checkGCCRuntimeUnit(test, "node_process.c")
}

func TestProduct_GCCRuntimeProcess(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "process.c")
}
func TestGCCRuntimeProcess(test *testing.T) { test.Parallel(); checkGCCRuntimeUnit(test, "process.c") }
