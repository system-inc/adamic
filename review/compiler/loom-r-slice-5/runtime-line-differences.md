Every difference in the source and re-cut runtime edit lines follows. Minus/plus inside each fenced diff describes the edit-sequence comparison; the second minus/plus is the actual runtime deletion/addition. Full patches preserve context.

### internal/native/runtime/adamic.h

```diff
--- source own edits
+++ re-cut own edits
@@ -6,8 +6,9 @@
 +size_t *adamic_object_ordered(const adamic_object *object);
 +adamic_object *adamic_object_construct(const adamic_shape *shape);
 +
--	if (cache->index == object->shape->count) {
-+	if (cache->index == object->shape->count || !adamic_object_present(object, cache->index)) {
++adamic_value *adamic_object_field_owned(const adamic_object *object, const char *name, adamic_slot_cache *cache, const adamic_object **owner);
+-	if (slot == object->shape->count) { return NULL; }
++	if (slot == object->shape->count || !adamic_object_present(object, slot)) { return NULL; }
 +	// An interior fixed layout, owned by this allocation, never a wrapper.
 +	adamic_object *metadata; // Never retained, never escapes as a value; retaining then releasing its zero-count header frees interior array storage.
 +adamic_array *adamic_node_array_new(size_t capacity, bool references, const adamic_shape *extras);
```

### internal/native/runtime/class_static.c

```diff
--- source own edits
+++ re-cut own edits
@@ -1 +1 @@
-+        adamic_object_publish(object, cache->index);
++        adamic_object_publish(object, adamic_slot_index(object, slot));
```

### internal/native/runtime/library_object.c

```diff
--- source own edits
+++ re-cut own edits
@@ -21,4 +21,4 @@
 - size_t *indices = ordered(source);
 + size_t *indices = adamic_object_ordered(source);
 +  if (!adamic_object_present(source, index)) continue;
-+  adamic_object_publish(target, cache.index);
++  adamic_object_publish(target, adamic_slot_index(target, slot));
```

### internal/native/runtime/object.c

```diff
--- source own edits
+++ re-cut own edits
@@ -1,13 +1,24 @@
 +	object->write_order = NULL;
+-static adamic_value *field_owned(const adamic_object *object, const char *name, adamic_slot_cache *cache, const adamic_object **owner) {
++adamic_value *adamic_object_field_owned(const adamic_object *object, const char *name, adamic_slot_cache *cache, const adamic_object **owner) {
 -	adamic_object *object = adamic_object_new(shape);
 +	adamic_object *object = source->write_order == NULL ? adamic_object_new(shape) : adamic_object_construct(shape);
 +	if (source->write_order != NULL) memcpy(object->write_order, source->write_order, shape->count * sizeof(size_t));
 +		if (!adamic_object_present(source, index)) continue;
+-			adamic_value *found = expression == NULL ? field_owned(source, shape->names[index], &cache, &owner) : adamic_object_read_mode(source, shape->names[index], &cache, expression, NULL, &owner);
++			adamic_value *found = expression == NULL ? adamic_object_field_owned(source, shape->names[index], &cache, &owner) : adamic_object_read_mode(source, shape->names[index], &cache, expression, NULL, &owner);
 -			return true;
 +			return adamic_object_present(object, index);
--	if (cache->index == object->shape->count) {
-+	if (cache->index == object->shape->count || !adamic_object_present(object, cache->index)) {
-+	adamic_object_publish(object, cache.index);
+-	if (slot == object->shape->count) { return NULL; }
++	if (slot == object->shape->count || !adamic_object_present(object, slot)) { return NULL; }
++	adamic_object_publish(object, adamic_slot_index(object, slot));
 -	adamic_value *slot = adamic_object_optional_field(object, name, cache);
 +	adamic_value *slot = adamic_object_optional_find(object, name, cache);
-+	if (slot == NULL && cache->index < object->shape->count) { slot = &object->slots[cache->index]; }
++	// An absent reserved field still has writable storage. Find it without rereading a shared cache.
++	if (slot == NULL) {
++		for (size_t index = 0; index < object->shape->count; index++) {
++			if (strcmp(object->shape->names[index], name) == 0) { slot = &object->slots[index]; break; }
++		}
++	}
+-	adamic_value *slot = field_owned(object, name, cache, &owner);
++	adamic_value *slot = adamic_object_field_owned(object, name, cache, &owner);
```

### internal/native/runtime/share.c

```diff
--- source own edits
+++ re-cut own edits
@@ -0,0 +1,7 @@
++			// Metadata is interior storage; publish its children, never its zero-count header.
++			if (array->metadata != NULL) {
++				adamic_object *metadata = array->metadata;
++				for (size_t i = 0; i < metadata->shape->count; i++) {
++					if (metadata->shape->references[i]) { append(&pending, metadata->slots[i].reference); }
++				}
++			}
```
