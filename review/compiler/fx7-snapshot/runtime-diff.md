# Runtime clearance for item 143

The snapshot starts with a full-width zero payload. Number and boolean storage copy only their owned member; maybe storage copies the unpacked member only when present. Null, undefined, and optional absence keep zero. Heap normalization is unchanged. Storage 0 is the documented exception, preserving the raw slot for readers that decide its meaning.

```diff
diff --git a/internal/native/runtime/object.c b/internal/native/runtime/object.c
index 13653e1b..c4a9ee0a 100644
--- a/internal/native/runtime/object.c
+++ b/internal/native/runtime/object.c
@@ -154,23 +154,25 @@ static adamic_value *adamic_object_read_mode(const adamic_object *object, const
 adamic_view_union_value adamic_object_view_union_snapshot(const adamic_object *object, const char *name, adamic_slot_cache *cache, const char *expression, const char *declared, bool absent) {
     adamic_value *slot = object == NULL ? NULL : adamic_object_optional_field(object, name, cache);
     if (object != NULL && slot == NULL && absent) {
-        return (adamic_view_union_value){adamic_view_union_undefined, {.reference = NULL}};
+        return (adamic_view_union_value){adamic_view_union_undefined, {0}};
     }
     const adamic_object *owner = NULL;
     slot = adamic_object_read_mode(object, name, cache, expression, declared, &owner);
     size_t index = adamic_slot_index(owner, slot);
     unsigned char storage = adamic_object_field_types(owner)[index];
-    adamic_view_union_value value = {adamic_view_union_unknown, *slot};
-    if (storage == adamic_rep_number) { value.kind = adamic_view_union_number; }
-    else if (storage == adamic_rep_boolean) { value.kind = adamic_view_union_boolean; }
+    adamic_view_union_value value = {adamic_view_union_unknown, {0}};
+    // Unknown storage is the one exception: its readers decide from the raw slot.
+    if (storage == 0) { value.payload = *slot; }
+    if (storage == adamic_rep_number) { value.kind = adamic_view_union_number; value.payload.number = slot->number; }
+    else if (storage == adamic_rep_boolean) { value.kind = adamic_view_union_boolean; value.payload.boolean = slot->boolean; }
     else if (storage == adamic_rep_maybe_number) {
         adamic_maybe_number number = adamic_maybe_number_unpack(slot->number);
         value.kind = number.present ? adamic_view_union_number : adamic_view_union_undefined;
-        value.payload.number = number.number;
+        if (number.present) { value.payload.number = number.number; }
     } else if (storage == adamic_rep_maybe_boolean) {
         adamic_maybe_boolean boolean = adamic_maybe_boolean_unpack(slot->maybe_boolean);
         value.kind = boolean.present ? adamic_view_union_boolean : adamic_view_union_undefined;
-        value.payload.boolean = boolean.boolean;
+        if (boolean.present) { value.payload.boolean = boolean.boolean; }
     } else if (storage == adamic_rep_null) { value.kind = adamic_view_union_null; }
     else if (storage == adamic_rep_undefined) { value.kind = adamic_view_union_undefined; }
     else if ((storage >= adamic_rep_string && storage <= adamic_rep_map) || storage == adamic_rep_closure || storage == adamic_rep_union || storage == adamic_rep_weak) {
```

Independent evidence: runtime-alone.log restores the old compiler retain and agrees with Node. runtime-mutant.log restores the raw slot only for maybe-boolean and restores the old compiler retain; release native crashes and UBSan diagnoses a misaligned adamic_heap access at heap.c:226, address 0x4045000000000002. The final direct sanitized payload test also covers scalar, maybe, nullish, optional absence and raw unknown slots (native-views.log).

## Counted-reference clearance condition

Record, uint8, int32 and float64 storage preserve `slot->reference` in the snapshot while keeping kind unknown. They have no named mixed-union kinds; naming them here would change admission and failure output. This takes only the member their storage tag owns, and does not restore raw stale bytes for scalar or nullish storage.

```diff
diff --git a/internal/native/runtime/object.c b/internal/native/runtime/object.c
index c4a9ee0a..1c0b5307 100644
--- a/internal/native/runtime/object.c
+++ b/internal/native/runtime/object.c
@@ -180,6 +180,8 @@ adamic_view_union_value adamic_object_view_union_snapshot(const adamic_object *o
         adamic_view_union_kind expected = storage == adamic_rep_string ? adamic_view_union_string : storage == adamic_rep_object || storage == adamic_rep_weak ? adamic_view_union_object : storage == adamic_rep_array ? adamic_view_union_array : storage == adamic_rep_map ? adamic_view_union_map : storage == adamic_rep_closure ? adamic_view_union_function : value.kind;
         if (value.kind != adamic_view_union_undefined && value.kind != expected) { value.kind = adamic_view_union_unknown; }
     }
+    // These counted references have no named mixed-union kind yet.
+    if (storage >= adamic_rep_record) { value.payload.reference = slot->reference; }
     return value;
 }
 
```
