// Copyright 2015 the V8 project authors. All rights reserved.
// Use of this source code is governed by a BSD-style license in THIRD_PARTY_NOTICES.md.
//
// V8 13.6.233 src/builtins/builtins-object.cc GetOwnPropertyKeys: ToObject, own keys,
// SKIP_SYMBOLS, ConvertToString. Lowering admits only primitives and complete plain data shapes.
// A boxed string has enumerable UTF-16 indices and a non-enumerable length; boxed numeric and
// boolean primitives have no own keys. Plain shapes reuse the existing canonical index ordering.
#include "adamic.h"

#include <stdio.h>
#include <string.h>

static adamic_string *name_string(const char *bytes, size_t length) {
	adamic_string *name = adamic_string_allocate(length);
	memcpy((char *)name->bytes, bytes, length);
	return name;
}

adamic_array *adamic_object_names(const adamic_heap *value, bool all) {
	if (value->kind == adamic_kind_object) {
		const adamic_object *object = (const adamic_object *)value;
        if (object->error_kind != 0) return adamic_error_names(object, all);
        return adamic_object_keys(object);
	}
	size_t length = value->kind == adamic_kind_string ? adamic_string_length((const adamic_string *)value) : 0;
	bool has_length = all && value->kind == adamic_kind_string;
	adamic_array *names = adamic_array_new(length + (has_length ? 1 : 0), true);
	for (size_t index = 0; index < length; index++) {
		char bytes[32];
		int count = snprintf(bytes, sizeof bytes, "%zu", index);
		adamic_array_push(names, (adamic_value){.reference = name_string(bytes, (size_t)count)});
	}
	if (has_length) {
		adamic_array_push(names, (adamic_value){.reference = name_string("length", 6)});
	}
	return names;
}
