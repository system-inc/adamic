Final runtime differences against 2c410041. All object.c differences are inherited from dbdf977f5b; mixed/untagged differences are the exact runtime-cleared correction. Each hunk below lists every removed/added line; complete contextual patches are saved alongside this report.

object.c
```diff
--- 2c410041/internal/native/runtime/object.c
+++ final/internal/native/runtime/object.c
@@ -21,0 +22,19 @@
+// field_owned is adamic_object_field that also says which object holds the slot: a constructor's
+// static field can live in a parent constructor until an own write shadows it (class_static.c).
+// Per-slot metadata is indexed from that object's slots, never from the shared packed cache.
+static adamic_value *field_owned(const adamic_object *object, const char *name, adamic_slot_cache *cache, const adamic_object **owner) {
+	while (object->class != NULL && object->class->is_static) {
+		adamic_value *slot = adamic_object_find(object, name, cache);
+		size_t flag = object->class->static_flags[adamic_slot_index(object, slot)];
+		if (flag == 0 || object->slots[flag - 1].number != 0 || object->class->static_parent == 0) {
+			*owner = object;
+			return slot;
+		}
+		object = object->slots[object->class->static_parent - 1].reference;
+	}
+	*owner = object;
+	return adamic_object_data_field(object, name, cache);
+}
+
+static adamic_value *adamic_object_read_mode(const adamic_object *object, const char *name, adamic_slot_cache *cache, const char *expression, const char *expected, const adamic_object **owner);
+
@@ -28 +47 @@
-		adamic_slot_cache cache = {NULL, 0};
+		adamic_slot_cache cache = {0};
@@ -30,5 +49 @@
-		object->slots[index] = accessor == NULL ? *(expression == NULL ? adamic_object_field(source, shape->names[index], &cache) : adamic_object_read(source, shape->names[index], &cache, expression)) : adamic_accessor_get((adamic_object *)source, shape->names[index]);
-		if (accessor != NULL && adamic_thrown != NULL) {
-			adamic_release(object);
-			return NULL;
-		}
+
@@ -36,2 +51,13 @@
-			adamic_object_initialized(object)[index] = adamic_object_initialized(source)[cache.index];
-			adamic_object_field_types(object)[index] = adamic_object_field_types(source)[cache.index];
+			// The slot found says where the field is, not the cache: another thread may store a
+			// different shape's slot in a shared cache (the packed cache, adamic.h).
+			const adamic_object *owner = source;
+			adamic_value *found = expression == NULL ? field_owned(source, shape->names[index], &cache, &owner) : adamic_object_read_mode(source, shape->names[index], &cache, expression, NULL, &owner);
+			object->slots[index] = *found;
+			adamic_object_initialized(object)[index] = adamic_object_initialized(owner)[adamic_slot_index(owner, found)];
+			adamic_object_field_types(object)[index] = adamic_object_field_types(owner)[adamic_slot_index(owner, found)];
+		} else {
+			object->slots[index] = adamic_accessor_get((adamic_object *)source, shape->names[index]);
+			if (adamic_thrown != NULL) {
+				adamic_release(object);
+				return NULL;
+			}
@@ -61,0 +88,10 @@
+void adamic_slot_cache_store(adamic_slot_cache *cache, const adamic_shape *shape, size_t index) {
+	uintptr_t pointer = (uintptr_t)shape;
+	if ((pointer & ~ADAMIC_SLOT_SHAPE_MASK) != 0) {
+		adamic_panic("shape address exceeds 48 bits", sizeof "shape address exceeds 48 bits" - 1);
+	}
+	uint64_t packed = index <= UINT16_MAX ? pointer | ((uint64_t)index << 48) : 0;
+	__atomic_store_n(&cache->packed, packed, __ATOMIC_RELAXED);
+	ADAMIC_TSAN_PAUSE(adamic_tsan_cache_publication);
+}
+
@@ -66,2 +102 @@
-			cache->shape = object->shape;
-			cache->index = index;
+			adamic_slot_cache_store(cache, object->shape, index);
@@ -78 +113,3 @@
-	if (cache->shape != shape) {
+	uint64_t packed = __atomic_load_n(&cache->packed, __ATOMIC_RELAXED);
+	size_t slot = packed >> 48;
+	if ((packed & ADAMIC_SLOT_SHAPE_MASK) != (uintptr_t)shape) {
@@ -82 +119 @@
-				cache->index = index;
+				slot = index;
@@ -88 +125 @@
-				cache->index = shape->count + index;
+				slot = shape->count + index;
@@ -96,6 +133,6 @@
-		cache->shape = shape;
-	}
-	if (cache->index < shape->count) {
-		return object->slots[cache->index].reference;
-	}
-	*method = shape->methods->code[cache->index - shape->count];
+		adamic_slot_cache_store(cache, shape, slot);
+	}
+	if (slot < shape->count) {
+		return object->slots[slot].reference;
+	}
+	*method = shape->methods->code[slot - shape->count];
@@ -109 +146,2 @@
-	if (object->shape->references[cache->index]) {
+	adamic_value *own = object->class != NULL && object->class->is_static ? adamic_object_find(object, name, cache) : slot;
+	if (object->shape->references[(size_t)(own - object->slots)]) {
@@ -121,12 +159,14 @@
-	cache->shape = object->shape;
-	cache->index = object->shape->count;
-	for (size_t index = 0; index < object->shape->count; index++) {
-		if (strcmp(object->shape->names[index], name) == 0) {
-			cache->index = index;
-			break;
-		}
-	}
-	if (cache->index == object->shape->count) {
-		return NULL;
-	}
-	return &((adamic_object *)object)->slots[cache->index];
+	uint64_t packed = __atomic_load_n(&cache->packed, __ATOMIC_RELAXED);
+	size_t slot = packed >> 48;
+	if ((packed & ADAMIC_SLOT_SHAPE_MASK) != (uintptr_t)object->shape) {
+		slot = object->shape->count;
+		for (size_t index = 0; index < object->shape->count; index++) {
+			if (strcmp(object->shape->names[index], name) == 0) {
+				slot = index;
+				break;
+			}
+		}
+		adamic_slot_cache_store(cache, object->shape, slot);
+	}
+	if (slot == object->shape->count) { return NULL; }
+	return &((adamic_object *)object)->slots[slot];
@@ -193,3 +233,4 @@
-	adamic_slot_cache cache = {NULL, 0};
-	(void)adamic_object_field(object, name, &cache);
-	adamic_object_initialized(object)[cache.index] = initialized;
+	adamic_slot_cache cache = {0};
+	// The object's own field: readiness belongs to the object being initialized, never a parent's.
+	adamic_value *slot = adamic_object_find(object, name, &cache);
+	adamic_object_initialized(object)[adamic_slot_index(object, slot)] = initialized;
@@ -204 +245 @@
-	unsigned char actual = adamic_object_field_types(owner)[cache->index];
+	unsigned char actual = adamic_object_field_types(owner)[adamic_slot_index(owner, slot)];
@@ -272,2 +313,3 @@
-	adamic_value *slot = adamic_object_field(object, name, cache);
-	if (object->shape->references[cache->index]) {
+	const adamic_object *owner = object;
+	adamic_value *slot = field_owned(object, name, cache, &owner);
+	if (owner->shape->references[adamic_slot_index(owner, slot)]) {
```

view_unions_mixed.c
```diff
--- 2c410041/internal/native/runtime/view_unions_mixed.c
+++ final/internal/native/runtime/view_unions_mixed.c
@@ -69,5 +68,0 @@
-    if (value->kind == adamic_view_union_object && value->payload.reference != NULL) {
-        const adamic_object *object = value->payload.reference;
-        if (object->tuple) { found = "array"; }
-        else if (object->class != NULL && object->class->is_static) { found = "function"; }
-    }
```

view_unions_untagged.c
```diff
--- 2c410041/internal/native/runtime/view_unions_untagged.c
+++ final/internal/native/runtime/view_unions_untagged.c
@@ -182,0 +183,6 @@
+    // Only this own-data adapter has proved the heap layout. Generic mixed
+    // selection must reject unavailable adapters without inspecting a payload.
+    if (value.kind == adamic_view_union_object && object != NULL) {
+        if (object->tuple) { value.kind = adamic_view_union_array; }
+        else if (object->class != NULL && object->class->is_static) { value.kind = adamic_view_union_function; }
+    }
```
