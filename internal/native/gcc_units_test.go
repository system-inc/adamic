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
	"count.c",
	"directory.c",
	"dtoa.c",
	"exceptions.c",
	"from_codes.c",
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
	"normalize.c",
	"number.c",
	"object.c",
	"parallel.c",
	"parse.c",
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

func TestProduct_GCCRuntimeAdamic(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "adamic.c") }
func TestGCCRuntimeAdamic(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "adamic.c") }

func TestProduct_GCCRuntimeArray(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "array.c") }
func TestGCCRuntimeArray(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "array.c") }

func TestProduct_GCCRuntimeArrayFrom(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "array_from.c") }
func TestGCCRuntimeArrayFrom(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "array_from.c") }

func TestProduct_GCCRuntimeAsync(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "async.c") }
func TestGCCRuntimeAsync(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "async.c") }

func TestProduct_GCCRuntimeBitwise(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "bitwise.c") }
func TestGCCRuntimeBitwise(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "bitwise.c") }

func TestProduct_GCCRuntimeCase(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "case.c") }
func TestGCCRuntimeCase(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "case.c") }

func TestProduct_GCCRuntimeCensusSmall(t *testing.T) {
	t.Parallel()
	gccRuntimeUnit(t, "census_small.c")
}
func TestGCCRuntimeCensusSmall(t *testing.T) { t.Parallel(); checkGCCRuntimeUnit(t, "census_small.c") }

func TestProduct_GCCRuntimeClassFeatures(t *testing.T) {
	t.Parallel()
	gccRuntimeUnit(t, "class_features.c")
}
func TestGCCRuntimeClassFeatures(t *testing.T) {
	t.Parallel()
	checkGCCRuntimeUnit(t, "class_features.c")
}

func TestProduct_GCCRuntimeClassInheritance(t *testing.T) {
	t.Parallel()
	gccRuntimeUnit(t, "class_inheritance.c")
}
func TestGCCRuntimeClassInheritance(t *testing.T) {
	t.Parallel()
	checkGCCRuntimeUnit(t, "class_inheritance.c")
}

func TestProduct_GCCRuntimeClassStatic(t *testing.T) {
	t.Parallel()
	gccRuntimeUnit(t, "class_static.c")
}
func TestGCCRuntimeClassStatic(t *testing.T) { t.Parallel(); checkGCCRuntimeUnit(t, "class_static.c") }

func TestProduct_GCCRuntimeClosure(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "closure.c") }
func TestGCCRuntimeClosure(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "closure.c") }

func TestProduct_GCCRuntimeCount(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "count.c") }
func TestGCCRuntimeCount(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "count.c") }

func TestProduct_GCCRuntimeDirectory(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "directory.c") }
func TestGCCRuntimeDirectory(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "directory.c") }

func TestProduct_GCCRuntimeDtoa(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "dtoa.c") }
func TestGCCRuntimeDtoa(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "dtoa.c") }

func TestProduct_GCCRuntimeExceptions(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "exceptions.c") }
func TestGCCRuntimeExceptions(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "exceptions.c") }

func TestProduct_GCCRuntimeFromCodes(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "from_codes.c") }
func TestGCCRuntimeFromCodes(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "from_codes.c") }

func TestProduct_GCCRuntimeGraphRegions(t *testing.T) {
	t.Parallel()
	gccRuntimeUnit(t, "graph_regions.c")
}
func TestGCCRuntimeGraphRegions(t *testing.T) {
	t.Parallel()
	checkGCCRuntimeUnit(t, "graph_regions.c")
}

func TestProduct_GCCRuntimeHeap(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "heap.c") }
func TestGCCRuntimeHeap(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "heap.c") }

func TestProduct_GCCRuntimeHostRuntime(t *testing.T) {
	t.Parallel()
	gccRuntimeUnit(t, "host_runtime.c")
}
func TestGCCRuntimeHostRuntime(t *testing.T) { t.Parallel(); checkGCCRuntimeUnit(t, "host_runtime.c") }

func TestProduct_GCCRuntimeHypot(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "hypot.c") }
func TestGCCRuntimeHypot(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "hypot.c") }

func TestProduct_GCCRuntimeIeee754(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "ieee754.c") }
func TestGCCRuntimeIeee754(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "ieee754.c") }

func TestProduct_GCCRuntimeInput(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "input.c") }
func TestGCCRuntimeInput(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "input.c") }

func TestProduct_GCCRuntimeJsonDecode(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "json_decode.c") }
func TestGCCRuntimeJsonDecode(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "json_decode.c") }

func TestProduct_GCCRuntimeJsonEncode(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "json_encode.c") }
func TestGCCRuntimeJsonEncode(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "json_encode.c") }

func TestProduct_GCCRuntimeJsonStringify(t *testing.T) {
	t.Parallel()
	gccRuntimeUnit(t, "json_stringify.c")
}
func TestGCCRuntimeJsonStringify(t *testing.T) {
	t.Parallel()
	checkGCCRuntimeUnit(t, "json_stringify.c")
}

func TestProduct_GCCRuntimeLibraryArray(t *testing.T) {
	t.Parallel()
	gccRuntimeUnit(t, "library_array.c")
}
func TestGCCRuntimeLibraryArray(t *testing.T) {
	t.Parallel()
	checkGCCRuntimeUnit(t, "library_array.c")
}

func TestProduct_GCCRuntimeLibraryLanguage(t *testing.T) {
	t.Parallel()
	gccRuntimeUnit(t, "library_language.c")
}
func TestGCCRuntimeLibraryLanguage(t *testing.T) {
	t.Parallel()
	checkGCCRuntimeUnit(t, "library_language.c")
}

func TestProduct_GCCRuntimeLibraryMathNumber(t *testing.T) {
	t.Parallel()
	gccRuntimeUnit(t, "library_math_number.c")
}
func TestGCCRuntimeLibraryMathNumber(t *testing.T) {
	t.Parallel()
	checkGCCRuntimeUnit(t, "library_math_number.c")
}

func TestProduct_GCCRuntimeLibraryObject(t *testing.T) {
	t.Parallel()
	gccRuntimeUnit(t, "library_object.c")
}
func TestGCCRuntimeLibraryObject(t *testing.T) {
	t.Parallel()
	checkGCCRuntimeUnit(t, "library_object.c")
}

func TestProduct_GCCRuntimeMap(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "map.c") }
func TestGCCRuntimeMap(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "map.c") }

func TestProduct_GCCRuntimeMapSet(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "map_set.c") }
func TestGCCRuntimeMapSet(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "map_set.c") }

func TestProduct_GCCRuntimeMath(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "math.c") }
func TestGCCRuntimeMath(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "math.c") }

func TestProduct_GCCRuntimeMaybe(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "maybe.c") }
func TestGCCRuntimeMaybe(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "maybe.c") }

func TestProduct_GCCRuntimeNodeBuffer(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "node_buffer.c") }
func TestGCCRuntimeNodeBuffer(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "node_buffer.c") }

func TestProduct_GCCRuntimeNodeCrypto(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "node_crypto.c") }
func TestGCCRuntimeNodeCrypto(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "node_crypto.c") }

func TestProduct_GCCRuntimeNodeFsFile(t *testing.T) {
	t.Parallel()
	gccRuntimeUnit(t, "node_fs_file.c")
}
func TestGCCRuntimeNodeFsFile(t *testing.T) { t.Parallel(); checkGCCRuntimeUnit(t, "node_fs_file.c") }

func TestProduct_GCCRuntimeNodeHost(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "node_host.c") }
func TestGCCRuntimeNodeHost(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "node_host.c") }

func TestProduct_GCCRuntimeNormalize(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "normalize.c") }
func TestGCCRuntimeNormalize(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "normalize.c") }

func TestProduct_GCCRuntimeNumber(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "number.c") }
func TestGCCRuntimeNumber(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "number.c") }

func TestProduct_GCCRuntimeObject(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "object.c") }
func TestGCCRuntimeObject(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "object.c") }

func TestProduct_GCCRuntimeParallel(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "parallel.c") }
func TestGCCRuntimeParallel(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "parallel.c") }

func TestProduct_GCCRuntimeParse(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "parse.c") }
func TestGCCRuntimeParse(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "parse.c") }

func TestProduct_GCCRuntimeProgramRegion(t *testing.T) {
	t.Parallel()
	gccRuntimeUnit(t, "program_region.c")
}
func TestGCCRuntimeProgramRegion(t *testing.T) {
	t.Parallel()
	checkGCCRuntimeUnit(t, "program_region.c")
}

func TestProduct_GCCRuntimeRadix(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "radix.c") }
func TestGCCRuntimeRadix(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "radix.c") }

func TestProduct_GCCRuntimeRecord(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "record.c") }
func TestGCCRuntimeRecord(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "record.c") }

func TestProduct_GCCRuntimeRegexp(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "regexp.c") }
func TestGCCRuntimeRegexp(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "regexp.c") }

func TestProduct_GCCRuntimeRegexpCompileBytecode(t *testing.T) {
	t.Parallel()
	gccRuntimeUnit(t, "regexp_compile_bytecode.c")
}
func TestGCCRuntimeRegexpCompileBytecode(t *testing.T) {
	t.Parallel()
	checkGCCRuntimeUnit(t, "regexp_compile_bytecode.c")
}

func TestProduct_GCCRuntimeRegexpCompileOwnership(t *testing.T) {
	t.Parallel()
	gccRuntimeUnit(t, "regexp_compile_ownership.c")
}
func TestGCCRuntimeRegexpCompileOwnership(t *testing.T) {
	t.Parallel()
	checkGCCRuntimeUnit(t, "regexp_compile_ownership.c")
}

func TestProduct_GCCRuntimeRegexpCompileParser(t *testing.T) {
	t.Parallel()
	gccRuntimeUnit(t, "regexp_compile_parser.c")
}
func TestGCCRuntimeRegexpCompileParser(t *testing.T) {
	t.Parallel()
	checkGCCRuntimeUnit(t, "regexp_compile_parser.c")
}

func TestProduct_GCCRuntimeRegexpCompileProperties(t *testing.T) {
	t.Parallel()
	gccRuntimeUnit(t, "regexp_compile_properties.c")
}
func TestGCCRuntimeRegexpCompileProperties(t *testing.T) {
	t.Parallel()
	checkGCCRuntimeUnit(t, "regexp_compile_properties.c")
}

func TestProduct_GCCRuntimeRegexpCompileRuntime(t *testing.T) {
	t.Parallel()
	gccRuntimeUnit(t, "regexp_compile_runtime.c")
}
func TestGCCRuntimeRegexpCompileRuntime(t *testing.T) {
	t.Parallel()
	checkGCCRuntimeUnit(t, "regexp_compile_runtime.c")
}

func TestProduct_GCCRuntimeRegexpCompileSets(t *testing.T) {
	t.Parallel()
	gccRuntimeUnit(t, "regexp_compile_sets.c")
}
func TestGCCRuntimeRegexpCompileSets(t *testing.T) {
	t.Parallel()
	checkGCCRuntimeUnit(t, "regexp_compile_sets.c")
}

func TestProduct_GCCRuntimeRegexpCompileV8(t *testing.T) {
	t.Parallel()
	gccRuntimeUnit(t, "regexp_compile_v8.c")
}
func TestGCCRuntimeRegexpCompileV8(t *testing.T) {
	t.Parallel()
	checkGCCRuntimeUnit(t, "regexp_compile_v8.c")
}

func TestProduct_GCCRuntimeRegexpRegular(t *testing.T) {
	t.Parallel()
	gccRuntimeUnit(t, "regexp_regular.c")
}
func TestGCCRuntimeRegexpRegular(t *testing.T) {
	t.Parallel()
	checkGCCRuntimeUnit(t, "regexp_regular.c")
}

func TestProduct_GCCRuntimeRegexpReplace(t *testing.T) {
	t.Parallel()
	gccRuntimeUnit(t, "regexp_replace.c")
}
func TestGCCRuntimeRegexpReplace(t *testing.T) {
	t.Parallel()
	checkGCCRuntimeUnit(t, "regexp_replace.c")
}

func TestProduct_GCCRuntimeRegion(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "region.c") }
func TestGCCRuntimeRegion(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "region.c") }

func TestProduct_GCCRuntimeSet(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "set.c") }
func TestGCCRuntimeSet(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "set.c") }

func TestProduct_GCCRuntimeShare(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "share.c") }
func TestGCCRuntimeShare(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "share.c") }

func TestProduct_GCCRuntimeSlabQuarantine(t *testing.T) {
	t.Parallel()
	gccRuntimeUnit(t, "slab_quarantine.c")
}
func TestGCCRuntimeSlabQuarantine(t *testing.T) {
	t.Parallel()
	checkGCCRuntimeUnit(t, "slab_quarantine.c")
}

func TestProduct_GCCRuntimeSort(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "sort.c") }
func TestGCCRuntimeSort(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "sort.c") }

func TestProduct_GCCRuntimeSortUndefined(t *testing.T) {
	t.Parallel()
	gccRuntimeUnit(t, "sort_undefined.c")
}
func TestGCCRuntimeSortUndefined(t *testing.T) {
	t.Parallel()
	checkGCCRuntimeUnit(t, "sort_undefined.c")
}

func TestProduct_GCCRuntimeSpread(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "spread.c") }
func TestGCCRuntimeSpread(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "spread.c") }

func TestProduct_GCCRuntimeStack(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "stack.c") }
func TestGCCRuntimeStack(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "stack.c") }

func TestProduct_GCCRuntimeString(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "string.c") }
func TestGCCRuntimeString(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "string.c") }

func TestProduct_GCCRuntimeStringAppend(t *testing.T) {
	t.Parallel()
	gccRuntimeUnit(t, "string_append.c")
}
func TestGCCRuntimeStringAppend(t *testing.T) {
	t.Parallel()
	checkGCCRuntimeUnit(t, "string_append.c")
}

func TestProduct_GCCRuntimeStringFrom(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "string_from.c") }
func TestGCCRuntimeStringFrom(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "string_from.c") }

func TestProduct_GCCRuntimeStringIndex(t *testing.T) {
	t.Parallel()
	gccRuntimeUnit(t, "string_index.c")
}
func TestGCCRuntimeStringIndex(t *testing.T) { t.Parallel(); checkGCCRuntimeUnit(t, "string_index.c") }

func TestProduct_GCCRuntimeStringShare(t *testing.T) {
	t.Parallel()
	gccRuntimeUnit(t, "string_share.c")
}
func TestGCCRuntimeStringShare(t *testing.T) { t.Parallel(); checkGCCRuntimeUnit(t, "string_share.c") }

func TestProduct_GCCRuntimeTsgo(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "tsgo.c") }
func TestGCCRuntimeTsgo(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "tsgo.c") }

func TestProduct_GCCRuntimeTypedArray(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "typed_array.c") }
func TestGCCRuntimeTypedArray(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "typed_array.c") }

func TestProduct_GCCRuntimeTypedArraySort(t *testing.T) {
	t.Parallel()
	gccRuntimeUnit(t, "typed_array_sort.c")
}
func TestGCCRuntimeTypedArraySort(t *testing.T) {
	t.Parallel()
	checkGCCRuntimeUnit(t, "typed_array_sort.c")
}

func TestProduct_GCCRuntimeUnion(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "union.c") }
func TestGCCRuntimeUnion(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "union.c") }

func TestProduct_GCCRuntimeUtf8(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "utf8.c") }
func TestGCCRuntimeUtf8(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "utf8.c") }

func TestProduct_GCCRuntimeViewUnionsMixed(t *testing.T) {
	t.Parallel()
	gccRuntimeUnit(t, "view_unions_mixed.c")
}
func TestGCCRuntimeViewUnionsMixed(t *testing.T) {
	t.Parallel()
	checkGCCRuntimeUnit(t, "view_unions_mixed.c")
}

func TestProduct_GCCRuntimeViewUnionsUntagged(t *testing.T) {
	t.Parallel()
	gccRuntimeUnit(t, "view_unions_untagged.c")
}
func TestGCCRuntimeViewUnionsUntagged(t *testing.T) {
	t.Parallel()
	checkGCCRuntimeUnit(t, "view_unions_untagged.c")
}

func TestProduct_GCCRuntimeWeak(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "weak.c") }
func TestGCCRuntimeWeak(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "weak.c") }

func TestProduct_GCCRuntimeNodeCwdWasi(t *testing.T) {
	t.Parallel()
	gccRuntimeUnit(t, "node_cwd_wasi.c")
}
func TestGCCRuntimeNodeCwdWasi(t *testing.T) { t.Parallel(); checkGCCRuntimeUnit(t, "node_cwd_wasi.c") }

func TestProduct_GCCRuntimeNodeFSDirectory(t *testing.T) {
	t.Parallel()
	gccRuntimeUnit(t, "node_fs_directory.c")
}
func TestGCCRuntimeNodeFSDirectory(t *testing.T) {
	t.Parallel()
	checkGCCRuntimeUnit(t, "node_fs_directory.c")
}

func TestProduct_GCCRuntimeNodePath(t *testing.T) { t.Parallel(); gccRuntimeUnit(t, "node_path.c") }
func TestGCCRuntimeNodePath(t *testing.T)         { t.Parallel(); checkGCCRuntimeUnit(t, "node_path.c") }
