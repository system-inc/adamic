# Runtime changes for @system_adamic_runtime

Base: origin/compiler/after-chain at 2391c655. Clearance requested for these net runtime hunks. The full patch is runtime.patch.

The chain already implements presence queries, SIZE_MAX absence, publication, key sorting/filtering, static writes, and region sizing. Those implementations are retained. class_static.c has no net hunk. construction.c also initializes the new metadata because it creates objects and interior metadata; leaving it uninitialized would make descriptor reads unsafe.

- `internal/native/runtime/adamic.h`, hunk 1, `@@ -290,21 +290,29 @@ typedef struct adamic_object {`: Retain the chain byte-tail and SIZE_MAX insertion-order layout. Add owned dynamic descriptor metadata, align reserved order storage, and declare optional deletion/reserving-copy helpers. Dynamic descriptors bypass pointer-identity caches because their addresses can be reused.

- `internal/native/runtime/adamic.h`, hunk 2, `@@ -394,6 +402,7 @@ typedef struct adamic_shape_types {`: Retain the chain byte-tail and SIZE_MAX insertion-order layout. Add owned dynamic descriptor metadata, align reserved order storage, and declare optional deletion/reserving-copy helpers. Dynamic descriptors bypass pointer-identity caches because their addresses can be reused.

- `internal/native/runtime/adamic.h`, hunk 3, `@@ -411,7 +420,7 @@ adamic_maybe_boolean adamic_object_maybe_boolean(const adamic_object *object, co`: Retain the chain byte-tail and SIZE_MAX insertion-order layout. Add owned dynamic descriptor metadata, align reserved order storage, and declare optional deletion/reserving-copy helpers. Dynamic descriptors bypass pointer-identity caches because their addresses can be reused.

- `internal/native/runtime/adamic.h`, hunk 4, `@@ -423,7 +432,7 @@ static inline adamic_value *adamic_object_optional_field(const adamic_object *ob`: Retain the chain byte-tail and SIZE_MAX insertion-order layout. Add owned dynamic descriptor metadata, align reserved order storage, and declare optional deletion/reserving-copy helpers. Dynamic descriptors bypass pointer-identity caches because their addresses can be reused.

- `internal/native/runtime/class_features.c`, hunk 1, `@@ -42,6 +42,7 @@ size_t adamic_public_index(const adamic_shape *shape, size_t position) {`: Plain-object enumeration uses runtime keys even without an order tail. The existing chain order-tail dispatch remains.

- `internal/native/runtime/construction.c`, hunk 1, `@@ -19,6 +19,8 @@ static void construction_start(adamic_object *object, const adamic_shape *shape)`: Initialize new descriptor metadata for constructed objects and interior NodeArray metadata. Keep all existing construction and publishing semantics.

- `internal/native/runtime/library_object.c`, hunk 1, `@@ -149,7 +149,7 @@ void adamic_object_assign(adamic_object *target, const adamic_object *source) {`: Assignment uses the existing runtime write path to publish readiness as well as presence. Existing sorting, presence filtering, and publication remain.

- `internal/native/runtime/object.c`, hunk 1, `@@ -14,6 +14,8 @@ adamic_object *adamic_object_new(const adamic_shape *shape) {`: Initialize descriptor metadata; share slot-state copying across fixed and owned dynamic shapes; preserve readiness, physical tags, and order including absent slots; bypass stale method caches for dynamic descriptors; implement deletion with release and reservation with aligned owned descriptors. No second presence representation is introduced.

- `internal/native/runtime/object.c`, hunk 2, `@@ -21,25 +23,38 @@ adamic_object *adamic_object_new(const adamic_shape *shape) {`: Initialize descriptor metadata; share slot-state copying across fixed and owned dynamic shapes; preserve readiness, physical tags, and order including absent slots; bypass stale method caches for dynamic descriptors; implement deletion with release and reservation with aligned owned descriptors. No second presence representation is introduced.

- `internal/native/runtime/object.c`, hunk 3, `@@ -82,7 +97,7 @@ adamic_closure *adamic_object_callee(const adamic_object *object, const char *na`: Initialize descriptor metadata; share slot-state copying across fixed and owned dynamic shapes; preserve readiness, physical tags, and order including absent slots; bypass stale method caches for dynamic descriptors; implement deletion with release and reservation with aligned owned descriptors. No second presence representation is introduced.

- `internal/native/runtime/object.c`, hunk 4, `@@ -204,6 +219,100 @@ void adamic_object_set_initialized(adamic_object *object, const char *name, bool`: Initialize descriptor metadata; share slot-state copying across fixed and owned dynamic shapes; preserve readiness, physical tags, and order including absent slots; bypass stale method caches for dynamic descriptors; implement deletion with release and reservation with aligned owned descriptors. No second presence representation is introduced.

- `internal/native/runtime/region.c`, hunk 1, `@@ -64,6 +64,8 @@ adamic_object *adamic_object_new_in(adamic_region *region, const adamic_shape *s`: Initialize descriptor metadata in region objects while retaining the chain shared size helper and byte tails.

- `internal/native/runtime/union.c`, hunk 1, `@@ -97,6 +97,13 @@ void adamic_register_shape_types(adamic_shape_types *metadata) {`: Expose registered scalar shape types to reserving copies and consult owned descriptor metadata for dynamic scalar reads. Retain the chain presence/readiness and physical-number/boolean checks.

- `internal/native/runtime/union.c`, hunk 2, `@@ -169,10 +176,9 @@ static adamic_heap *dynamic_slot(const adamic_object *object, size_t index) {`: Expose registered scalar shape types to reserving copies and consult owned descriptor metadata for dynamic scalar reads. Retain the chain presence/readiness and physical-number/boolean checks.


## Runtime clearance amendment

The object header no longer holds dynamic_types. dynamic_shape remains in padding.
The reserving-copy allocation holds struct adamic_dynamic_shape, with the shape
first and its type pointer next. Its name/type/reference arrays begin after the
complete descriptor. Dynamic copy and unknown reads recover types through shape.
Ordinary, construction and region allocators remove the obsolete pointer reset;
region.c retains both readiness and representation bitmap initializers.
Ordinary object header size falls from 56 to 48 bytes on this target; reserving
copies move the same eight bytes into their existing descriptor allocation.
All seven affected recorded optional-field count rows are unchanged, including
peak live. Counts measure live values, not object bytes.
The descriptor mutant reads dynamic types from static registration instead of the
owned descriptor. A second reserving copy observes wrong types and exits 71 after
releasing its allocations; both release and ASan/UBSan probes catch it.

## Checked optional writes on V1

These additional runtime hunks await runtime review.

- adamic.h declares the checked optional read and write entry points.
- object.c reads optional presence and readiness through the shared snapshot, normalizes scalar representations, and reports incompatible present values.
- object.c names the destination storage type for write diagnostics.
- object.c validates the incoming value before storing, preserves the physical slot representation, packs optional scalars or boxes union values, retains incoming references before releasing old ones, and publishes readiness and presence after a successful store.

Exact hunks: step2-runtime.patch.

## Runtime review ownership amendment

- adamic.h documents borrowed reference results of adamic_object_view and borrowed strings from adamic_object_optional_view, including NULL for undefined.
- adamic.h states that value is given, owned by adamic_object_view_store; its caller must supply keptValue before the store releases the old slot.
- object.c accepts incoming undefined (13) in string storage (3). NULL is already the optional string representation, and the existing store still publishes property presence independently. Both-backend Node fixtures and a restoring-rejection mutant cover this hunk.

See ownership-runtime.patch for exact hunks. The frozen-write objection was withdrawn: emit_statements.go already calls adamic_object_check_data_write for every SetProperty.
