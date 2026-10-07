// Copyright 2015 the V8 project authors. All rights reserved.
// Use of this source code is governed by a BSD-style license in THIRD_PARTY_NOTICES.md.
//
// Port of ObjectIsSealed/ObjectSeal in V8 13.6.233 src/builtins/builtins-object.cc and
// GenericTestIntegrityLevel/SetIntegrityLevel in src/objects/js-objects.cc. The lowering proof
// restricts mutations to complete plain shapes: all public slots are configurable, writable data
// properties initially. No proxy, accessor or individually redefined descriptor reaches here.
#include "adamic.h"

static bool primitive(const adamic_heap *value) {
	return value == NULL || value->kind == adamic_kind_string ||
		value->kind == adamic_kind_number || value->kind == adamic_kind_boolean;
}

bool adamic_object_is_extensible(const adamic_heap *value) {
	if (primitive(value)) return false;
	if (value->kind != adamic_kind_object) return true;
	const adamic_object *object = (const adamic_object *)value;
	return !object->frozen && !object->nonextensible;
}

bool adamic_object_test_integrity(const adamic_heap *value, bool frozen) {
	// V8's builtin returns true for non-receivers, without boxing them.
	if (primitive(value)) return true;
	// Integrity mutation of arrays, maps and closures is refused by lowering.
	if (value->kind != adamic_kind_object) return false;
	const adamic_object *object = (const adamic_object *)value;
	// GenericTestIntegrityLevel first rejects an extensible receiver, even an empty one.
	if (adamic_object_is_extensible(value)) return false;
	for (size_t index = 0; index < object->shape->count; index++) {
		// Private fields are internal slots, not OwnPropertyKeys.
		if (object->shape->names[index][0] == '#') continue;
		if (!object->sealed && !object->frozen) return false;
		if (frozen && !object->frozen) return false;
	}
	return true;
}

adamic_object *adamic_object_set_integrity(adamic_object *object, bool sealed) {
	// SetIntegrityLevel prevents extensions before making every own key non-configurable.
	object->nonextensible = true;
	if (sealed) object->sealed = true;
	return adamic_retain(object);
}
