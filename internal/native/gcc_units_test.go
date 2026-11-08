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

func TestProduct_GccRuntimeAdamic(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "adamic.c") }
func TestGccRuntimeAdamic(test *testing.T)         { test.Parallel(); checkGccRuntimeUnit(test, "adamic.c") }

func TestProduct_GccRuntimeArray(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "array.c") }
func TestGccRuntimeArray(test *testing.T)         { test.Parallel(); checkGccRuntimeUnit(test, "array.c") }

func TestProduct_GccRuntimeArrayFrom(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "array_from.c")
}
func TestGccRuntimeArrayFrom(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "array_from.c")
}

func TestProduct_GccRuntimeAsync(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "async.c") }
func TestGccRuntimeAsync(test *testing.T)         { test.Parallel(); checkGccRuntimeUnit(test, "async.c") }

func TestProduct_GccRuntimeBitwise(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "bitwise.c")
}
func TestGccRuntimeBitwise(test *testing.T) { test.Parallel(); checkGccRuntimeUnit(test, "bitwise.c") }

func TestProduct_GccRuntimeCase(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "case.c") }
func TestGccRuntimeCase(test *testing.T)         { test.Parallel(); checkGccRuntimeUnit(test, "case.c") }

func TestProduct_GccRuntimeCensusSmall(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "census_small.c")
}
func TestGccRuntimeCensusSmall(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "census_small.c")
}

func TestProduct_GccRuntimeClassFeatures(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "class_features.c")
}
func TestGccRuntimeClassFeatures(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "class_features.c")
}

func TestProduct_GccRuntimeClassInheritance(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "class_inheritance.c")
}
func TestGccRuntimeClassInheritance(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "class_inheritance.c")
}

func TestProduct_GccRuntimeClassStatic(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "class_static.c")
}
func TestGccRuntimeClassStatic(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "class_static.c")
}

func TestProduct_GccRuntimeClosure(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "closure.c")
}
func TestGccRuntimeClosure(test *testing.T) { test.Parallel(); checkGccRuntimeUnit(test, "closure.c") }

func TestProduct_GccRuntimeConstruction(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "construction.c")
}
func TestGccRuntimeConstruction(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "construction.c")
}

func TestProduct_GccRuntimeCount(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "count.c") }
func TestGccRuntimeCount(test *testing.T)         { test.Parallel(); checkGccRuntimeUnit(test, "count.c") }

func TestProduct_GccRuntimeDirectory(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "directory.c")
}
func TestGccRuntimeDirectory(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "directory.c")
}

func TestProduct_GccRuntimeDtoa(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "dtoa.c") }
func TestGccRuntimeDtoa(test *testing.T)         { test.Parallel(); checkGccRuntimeUnit(test, "dtoa.c") }

func TestProduct_GccRuntimeExceptions(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "exceptions.c")
}
func TestGccRuntimeExceptions(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "exceptions.c")
}

func TestProduct_GccRuntimeFromCodes(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "from_codes.c")
}
func TestGccRuntimeFromCodes(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "from_codes.c")
}

func TestProduct_GccRuntimeGenerator(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "generator.c")
}
func TestGccRuntimeGenerator(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "generator.c")
}

func TestProduct_GccRuntimeGraphRegions(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "graph_regions.c")
}
func TestGccRuntimeGraphRegions(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "graph_regions.c")
}

func TestProduct_GccRuntimeHeap(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "heap.c") }
func TestGccRuntimeHeap(test *testing.T)         { test.Parallel(); checkGccRuntimeUnit(test, "heap.c") }

func TestProduct_GccRuntimeHostRuntime(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "host_runtime.c")
}
func TestGccRuntimeHostRuntime(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "host_runtime.c")
}

func TestProduct_GccRuntimeHypot(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "hypot.c") }
func TestGccRuntimeHypot(test *testing.T)         { test.Parallel(); checkGccRuntimeUnit(test, "hypot.c") }

func TestProduct_GccRuntimeIeee754(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "ieee754.c")
}
func TestGccRuntimeIeee754(test *testing.T) { test.Parallel(); checkGccRuntimeUnit(test, "ieee754.c") }

func TestProduct_GccRuntimeInput(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "input.c") }
func TestGccRuntimeInput(test *testing.T)         { test.Parallel(); checkGccRuntimeUnit(test, "input.c") }

func TestProduct_GccRuntimeJsonDecode(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "json_decode.c")
}
func TestGccRuntimeJsonDecode(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "json_decode.c")
}

func TestProduct_GccRuntimeJsonEncode(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "json_encode.c")
}
func TestGccRuntimeJsonEncode(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "json_encode.c")
}

func TestProduct_GccRuntimeJsonStringify(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "json_stringify.c")
}
func TestGccRuntimeJsonStringify(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "json_stringify.c")
}

func TestProduct_GccRuntimeLibraryArray(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "library_array.c")
}
func TestGccRuntimeLibraryArray(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "library_array.c")
}

func TestProduct_GccRuntimeLibraryLanguage(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "library_language.c")
}
func TestGccRuntimeLibraryLanguage(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "library_language.c")
}

func TestProduct_GccRuntimeLibraryMathNumber(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "library_math_number.c")
}
func TestGccRuntimeLibraryMathNumber(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "library_math_number.c")
}

func TestProduct_GccRuntimeLibraryObject(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "library_object.c")
}
func TestGccRuntimeLibraryObject(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "library_object.c")
}

func TestProduct_GccRuntimeMap(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "map.c") }
func TestGccRuntimeMap(test *testing.T)         { test.Parallel(); checkGccRuntimeUnit(test, "map.c") }

func TestProduct_GccRuntimeMapSet(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "map_set.c")
}
func TestGccRuntimeMapSet(test *testing.T) { test.Parallel(); checkGccRuntimeUnit(test, "map_set.c") }

func TestProduct_GccRuntimeMath(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "math.c") }
func TestGccRuntimeMath(test *testing.T)         { test.Parallel(); checkGccRuntimeUnit(test, "math.c") }

func TestProduct_GccRuntimeMaybe(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "maybe.c") }
func TestGccRuntimeMaybe(test *testing.T)         { test.Parallel(); checkGccRuntimeUnit(test, "maybe.c") }

func TestProduct_GccRuntimeNodeBuffer(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "node_buffer.c")
}
func TestGccRuntimeNodeBuffer(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "node_buffer.c")
}

func TestProduct_GccRuntimeNodeCrypto(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "node_crypto.c")
}
func TestGccRuntimeNodeCrypto(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "node_crypto.c")
}

func TestProduct_GccRuntimeNodeFileSystemFile(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "node_fs_file.c")
}
func TestGccRuntimeNodeFileSystemFile(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "node_fs_file.c")
}

func TestProduct_GccRuntimeNodeHost(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "node_host.c")
}
func TestGccRuntimeNodeHost(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "node_host.c")
}

func TestProduct_GccRuntimeNormalize(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "normalize.c")
}
func TestGccRuntimeNormalize(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "normalize.c")
}

func TestProduct_GccRuntimeNumber(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "number.c") }
func TestGccRuntimeNumber(test *testing.T)         { test.Parallel(); checkGccRuntimeUnit(test, "number.c") }

func TestProduct_GccRuntimeObject(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "object.c") }
func TestGccRuntimeObject(test *testing.T)         { test.Parallel(); checkGccRuntimeUnit(test, "object.c") }

func TestProduct_GccRuntimeParallel(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "parallel.c")
}
func TestGccRuntimeParallel(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "parallel.c")
}

func TestProduct_GccRuntimeParse(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "parse.c") }
func TestGccRuntimeParse(test *testing.T)         { test.Parallel(); checkGccRuntimeUnit(test, "parse.c") }

func TestProduct_GccRuntimeProgramRegion(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "program_region.c")
}
func TestGccRuntimeProgramRegion(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "program_region.c")
}

func TestProduct_GccRuntimeRadix(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "radix.c") }
func TestGccRuntimeRadix(test *testing.T)         { test.Parallel(); checkGccRuntimeUnit(test, "radix.c") }

func TestProduct_GccRuntimeRecord(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "record.c") }
func TestGccRuntimeRecord(test *testing.T)         { test.Parallel(); checkGccRuntimeUnit(test, "record.c") }

func TestProduct_GccRuntimeRegexp(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "regexp.c") }
func TestGccRuntimeRegexp(test *testing.T)         { test.Parallel(); checkGccRuntimeUnit(test, "regexp.c") }

func TestProduct_GccRuntimeRegexpCompileBytecode(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "regexp_compile_bytecode.c")
}
func TestGccRuntimeRegexpCompileBytecode(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "regexp_compile_bytecode.c")
}

func TestProduct_GccRuntimeRegexpCompileOwnership(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "regexp_compile_ownership.c")
}
func TestGccRuntimeRegexpCompileOwnership(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "regexp_compile_ownership.c")
}

func TestProduct_GccRuntimeRegexpCompileParser(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "regexp_compile_parser.c")
}
func TestGccRuntimeRegexpCompileParser(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "regexp_compile_parser.c")
}

func TestProduct_GccRuntimeRegexpCompileProperties(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "regexp_compile_properties.c")
}
func TestGccRuntimeRegexpCompileProperties(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "regexp_compile_properties.c")
}

func TestProduct_GccRuntimeRegexpCompileRuntime(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "regexp_compile_runtime.c")
}
func TestGccRuntimeRegexpCompileRuntime(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "regexp_compile_runtime.c")
}

func TestProduct_GccRuntimeRegexpCompileSets(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "regexp_compile_sets.c")
}
func TestGccRuntimeRegexpCompileSets(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "regexp_compile_sets.c")
}

func TestProduct_GccRuntimeRegexpCompileV8(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "regexp_compile_v8.c")
}
func TestGccRuntimeRegexpCompileV8(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "regexp_compile_v8.c")
}

func TestProduct_GccRuntimeRegexpRegular(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "regexp_regular.c")
}
func TestGccRuntimeRegexpRegular(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "regexp_regular.c")
}

func TestProduct_GccRuntimeRegexpReplace(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "regexp_replace.c")
}
func TestGccRuntimeRegexpReplace(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "regexp_replace.c")
}

func TestProduct_GccRuntimeRegion(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "region.c") }
func TestGccRuntimeRegion(test *testing.T)         { test.Parallel(); checkGccRuntimeUnit(test, "region.c") }

func TestProduct_GccRuntimeSet(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "set.c") }
func TestGccRuntimeSet(test *testing.T)         { test.Parallel(); checkGccRuntimeUnit(test, "set.c") }

func TestProduct_GccRuntimeShare(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "share.c") }
func TestGccRuntimeShare(test *testing.T)         { test.Parallel(); checkGccRuntimeUnit(test, "share.c") }

func TestProduct_GccRuntimeSlabQuarantine(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "slab_quarantine.c")
}
func TestGccRuntimeSlabQuarantine(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "slab_quarantine.c")
}

func TestProduct_GccRuntimeSort(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "sort.c") }
func TestGccRuntimeSort(test *testing.T)         { test.Parallel(); checkGccRuntimeUnit(test, "sort.c") }

func TestProduct_GccRuntimeSortUndefined(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "sort_undefined.c")
}
func TestGccRuntimeSortUndefined(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "sort_undefined.c")
}

func TestProduct_GccRuntimeSpread(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "spread.c") }
func TestGccRuntimeSpread(test *testing.T)         { test.Parallel(); checkGccRuntimeUnit(test, "spread.c") }

func TestProduct_GccRuntimeStack(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "stack.c") }
func TestGccRuntimeStack(test *testing.T)         { test.Parallel(); checkGccRuntimeUnit(test, "stack.c") }

func TestProduct_GccRuntimeString(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "string.c") }
func TestGccRuntimeString(test *testing.T)         { test.Parallel(); checkGccRuntimeUnit(test, "string.c") }

func TestProduct_GccRuntimeStringAppend(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "string_append.c")
}
func TestGccRuntimeStringAppend(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "string_append.c")
}

func TestProduct_GccRuntimeStringFrom(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "string_from.c")
}
func TestGccRuntimeStringFrom(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "string_from.c")
}

func TestProduct_GccRuntimeStringIndex(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "string_index.c")
}
func TestGccRuntimeStringIndex(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "string_index.c")
}

func TestProduct_GccRuntimeStringShare(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "string_share.c")
}
func TestGccRuntimeStringShare(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "string_share.c")
}

func TestProduct_GccRuntimeTsgo(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "tsgo.c") }
func TestGccRuntimeTsgo(test *testing.T)         { test.Parallel(); checkGccRuntimeUnit(test, "tsgo.c") }

func TestProduct_GccRuntimeTypedArray(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "typed_array.c")
}
func TestGccRuntimeTypedArray(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "typed_array.c")
}

func TestProduct_GccRuntimeTypedArraySort(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "typed_array_sort.c")
}
func TestGccRuntimeTypedArraySort(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "typed_array_sort.c")
}

func TestProduct_GccRuntimeUnion(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "union.c") }
func TestGccRuntimeUnion(test *testing.T)         { test.Parallel(); checkGccRuntimeUnit(test, "union.c") }

func TestProduct_GccRuntimeUtf8(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "utf8.c") }
func TestGccRuntimeUtf8(test *testing.T)         { test.Parallel(); checkGccRuntimeUnit(test, "utf8.c") }

func TestProduct_GccRuntimeViewUnionsMixed(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "view_unions_mixed.c")
}
func TestGccRuntimeViewUnionsMixed(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "view_unions_mixed.c")
}

func TestProduct_GccRuntimeViewUnionsUntagged(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "view_unions_untagged.c")
}
func TestGccRuntimeViewUnionsUntagged(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "view_unions_untagged.c")
}

func TestProduct_GccRuntimeWeak(test *testing.T) { test.Parallel(); gccRuntimeUnit(test, "weak.c") }
func TestGccRuntimeWeak(test *testing.T)         { test.Parallel(); checkGccRuntimeUnit(test, "weak.c") }

func TestProduct_GccRuntimeNodeCwdWasi(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "node_cwd_wasi.c")
}
func TestGccRuntimeNodeCwdWasi(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "node_cwd_wasi.c")
}

func TestProduct_GccRuntimeNodeFileSystemDirectory(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "node_fs_directory.c")
}
func TestGccRuntimeNodeFileSystemDirectory(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "node_fs_directory.c")
}

func TestProduct_GccRuntimeNodePath(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "node_path.c")
}
func TestGccRuntimeNodePath(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "node_path.c")
}

func TestProduct_GccRuntimeNodeProcess(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "node_process.c")
}
func TestGccRuntimeNodeProcess(test *testing.T) {
	test.Parallel()
	checkGccRuntimeUnit(test, "node_process.c")
}

func TestProduct_GccRuntimeProcess(test *testing.T) {
	test.Parallel()
	gccRuntimeUnit(test, "process.c")
}
func TestGccRuntimeProcess(test *testing.T) { test.Parallel(); checkGccRuntimeUnit(test, "process.c") }
