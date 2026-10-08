// Copyright 2015 the V8 project authors. All rights reserved.
// Use of this source code is governed by a BSD-style license in THIRD_PARTY_NOTICES.md.
//
// Port of ObjectIsSealed/ObjectSeal in V8 13.6.233 src/builtins/builtins-object.cc and
// GenericTestIntegrityLevel/SetIntegrityLevel in src/objects/js-objects.cc. The lowering proof
// restricts mutations to complete plain shapes: all public slots are configurable, writable data
// properties initially. No proxy or accessor reaches here; scalar descriptor metadata is tested separately.
#include "adamic.h"
#include <stdlib.h>

typedef struct collection_integrity {
 const adamic_heap *value;
 bool sealed;
 struct collection_integrity *next;
} collection_integrity;
static collection_integrity *collections;
static collection_integrity *collection_level(const adamic_heap *value) {
 for (collection_integrity *entry = collections; entry != NULL; entry = entry->next) if (entry->value == value) return entry;
 return NULL;
}
adamic_heap *adamic_object_set_collection_integrity(adamic_heap *value, bool sealed) {
 collection_integrity *entry = collection_level(value);
 if (entry == NULL) {
  entry = malloc(sizeof *entry);
  if (entry == NULL) adamic_panic("out of memory", sizeof "out of memory" - 1);
  *entry = (collection_integrity){value, false, collections};
  collections = entry;
 }
 entry->sealed = entry->sealed || sealed;
 return adamic_retain(value);
}
void adamic_object_collection_integrity_forget(const adamic_heap *value) {
 collection_integrity **at = &collections;
 while (*at != NULL) {
  collection_integrity *entry = *at;
  if (entry->value != value) { at = &entry->next; continue; }
  *at = entry->next; free(entry); return;
 }
}

static bool primitive(const adamic_heap *value) {
	return value == NULL || value == &adamic_null || value->kind == adamic_kind_string ||
		value->kind == adamic_kind_number || value->kind == adamic_kind_boolean;
}

bool adamic_object_is_extensible(const adamic_heap *value) {
	if (primitive(value)) return false;
	if (value->kind == adamic_kind_map) return collection_level(value) == NULL;
	if (value->kind != adamic_kind_object) return true;
	const adamic_object *object = (const adamic_object *)value;
	return !object->frozen && !object->nonextensible;
}

bool adamic_object_test_integrity(const adamic_heap *value, bool frozen) {
	// V8's builtin returns true for non-receivers, without boxing them.
	if (primitive(value)) return true;
	// A supported collection has no own properties, so preventing extensions
	// already satisfies both integrity levels, regardless of its internal entries.
	if (value->kind == adamic_kind_map) return collection_level(value) != NULL;
	// Integrity mutation of arrays and closures is refused by lowering.
	if (value->kind != adamic_kind_object) return false;
	const adamic_object *object = (const adamic_object *)value;
	return adamic_object_descriptor_integrity(object, frozen);
}

adamic_object *adamic_object_set_integrity(adamic_object *object, bool sealed) {
	// SetIntegrityLevel prevents extensions before making every own key non-configurable.
	object->nonextensible = true;
	if (sealed) object->sealed = true;
	return adamic_retain(object);
}
