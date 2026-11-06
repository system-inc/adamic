// dynamic_keyed_names.h: a run-time key read's stops and names (adamic_dynamic_property_keyed, in
// union.c), written from internal/ir/dynamic_keyed_stops.go, the table the JavaScript backend emits
// from too. Don't edit it; run go test ./internal/native -run '^TestDynamicKeyedNamesHeaderIsCurrent$'
// -args -update-dynamic-keyed-names. Only union.c includes it.

#ifndef ADAMIC_DYNAMIC_KEYED_NAMES_H
#define ADAMIC_DYNAMIC_KEYED_NAMES_H

enum adamic_dynamic_keyed_stop {
	adamic_dynamic_keyed_stop_nullish,
	adamic_dynamic_keyed_stop_unknown_class,
	adamic_dynamic_keyed_stop_primitive,
	adamic_dynamic_keyed_stop_function,
	adamic_dynamic_keyed_stop_collection,
	adamic_dynamic_keyed_stop_typed_array,
	adamic_dynamic_keyed_stop_static_class,
	adamic_dynamic_keyed_stop_runtime_value,
	adamic_dynamic_keyed_stop_private,
	adamic_dynamic_keyed_stop_proto,
	adamic_dynamic_keyed_stop_error_property,
	adamic_dynamic_keyed_stop_object_literal_getter,
	adamic_dynamic_keyed_stop_setter_only,
	adamic_dynamic_keyed_stop_method,
	adamic_dynamic_keyed_stop_class_constructor,
};

static const char *const adamic_dynamic_keyed_stop_messages[] = {
	"compiler bug: a run-time key read reached null or undefined",
	"compiler bug: a run-time key read reached a class with no row in its table",
	"a primitive receiver's properties aren't read yet",
	"a function's properties aren't read by a run-time key yet",
	"a Map's or a Set's properties aren't read by a run-time key yet",
	"a typed array's properties aren't read by a run-time key yet",
	"a class's static properties aren't read by a run-time key yet",
	"this runtime value's properties aren't read by a run-time key yet",
	"a name beginning with # isn't read by a run-time key yet",
	"__proto__ isn't read by a run-time key yet",
	"an Error property Node sets and the native Error lacks isn't read yet",
	"an object literal's getter isn't read by a run-time key yet",
	"a setter without a getter isn't read by a run-time key yet",
	"a method isn't read by a run-time key yet",
	"a class instance's constructor isn't read by a run-time key yet",
};

static const char *const adamic_dynamic_keyed_prototype_method_names[] = {
	"toString",
	"toLocaleString",
	"valueOf",
	"hasOwnProperty",
	"isPrototypeOf",
	"propertyIsEnumerable",
	"__defineGetter__",
	"__defineSetter__",
	"__lookupGetter__",
	"__lookupSetter__",
};

static const char *const adamic_dynamic_keyed_array_prototype_names[] = {
	"constructor",
	"at",
	"concat",
	"copyWithin",
	"fill",
	"find",
	"findIndex",
	"findLast",
	"findLastIndex",
	"lastIndexOf",
	"pop",
	"push",
	"reverse",
	"shift",
	"unshift",
	"slice",
	"sort",
	"splice",
	"includes",
	"indexOf",
	"join",
	"keys",
	"entries",
	"values",
	"forEach",
	"filter",
	"flat",
	"flatMap",
	"map",
	"every",
	"some",
	"reduce",
	"reduceRight",
	"toReversed",
	"toSorted",
	"toSpliced",
	"with",
	"toLocaleString",
	"toString",
};

static const char *const adamic_dynamic_keyed_regexp_string_iterator_names[] = {
	"next",
	"constructor",
	"reduce",
	"toArray",
	"forEach",
	"some",
	"every",
	"find",
	"map",
	"filter",
	"take",
	"drop",
	"flatMap",
};

static const char *const adamic_dynamic_keyed_error_names[] = {
	"stack",
	"cause",
	"errno",
	"syscall",
	"path",
	"dest",
	"toString",
	"constructor",
};

// Each of adamic_dynamic_keyed_array_prototype_names by its index, with Node's length for it.
#define ADAMIC_DYNAMIC_KEYED_ARRAY_PROTOTYPE_FUNCTIONS(function) \
	function(0, 1) \
	function(1, 1) \
	function(2, 1) \
	function(3, 2) \
	function(4, 1) \
	function(5, 1) \
	function(6, 1) \
	function(7, 1) \
	function(8, 1) \
	function(9, 1) \
	function(10, 0) \
	function(11, 1) \
	function(12, 0) \
	function(13, 0) \
	function(14, 1) \
	function(15, 2) \
	function(16, 1) \
	function(17, 2) \
	function(18, 1) \
	function(19, 1) \
	function(20, 1) \
	function(21, 0) \
	function(22, 0) \
	function(23, 0) \
	function(24, 1) \
	function(25, 1) \
	function(26, 0) \
	function(27, 1) \
	function(28, 1) \
	function(29, 1) \
	function(30, 1) \
	function(31, 1) \
	function(32, 1) \
	function(33, 0) \
	function(34, 1) \
	function(35, 2) \
	function(36, 2) \
	function(37, 0) \
	function(38, 0)

// Each of adamic_dynamic_keyed_regexp_string_iterator_names by its index, with Node's length for it.
#define ADAMIC_DYNAMIC_KEYED_REGEXP_STRING_ITERATOR_FUNCTIONS(function) \
	function(0, 0) \
	function(1, 0) \
	function(2, 1) \
	function(3, 0) \
	function(4, 1) \
	function(5, 1) \
	function(6, 1) \
	function(7, 1) \
	function(8, 1) \
	function(9, 1) \
	function(10, 1) \
	function(11, 1) \
	function(12, 1)

#endif
