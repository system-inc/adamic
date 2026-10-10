#ifndef ADAMIC_VIEW_UNIONS_MIXED_H
#define ADAMIC_VIEW_UNIONS_MIXED_H

#include "adamic.h"
#include "view_representations.h"

// Logical kinds, not object storage bytes or heap-header kinds. The common
// readiness-aware slot probe must normalize a value before calling selection.
typedef enum adamic_view_union_kind {
    adamic_view_union_unknown,
    adamic_view_union_number,
    adamic_view_union_boolean,
    adamic_view_union_string,
    adamic_view_union_object,
    adamic_view_union_array,
    adamic_view_union_map,
    adamic_view_union_function,
    adamic_view_union_null,
    adamic_view_union_undefined
} adamic_view_union_kind;

typedef struct adamic_view_union_value {
    adamic_view_union_kind kind;
    adamic_value payload;
} adamic_view_union_value;

typedef struct adamic_view_union_member {
    adamic_view_union_kind kind;
    bool literal;
    adamic_value value;
    size_t contract;
} adamic_view_union_member;

// Reference members need a complete non-panicking contract test supplied by
// their owning adapter. Kind alone cannot certify an object, array or callable.
// Failure returns false so a later alternative can be tried without trapping.
// Adapters use shared metadata/probes, never getters or user callable bodies.
// Snapshots are borrowed; selection transfers no ownership or readiness proof.
typedef bool (*adamic_view_union_match)(void *context, const adamic_view_union_member *member, const adamic_view_union_value *value);

// Normalize a borrowed, source-certified heap value; NULL represents undefined.
// Null storage must be classified separately by its owning slot probe.
adamic_view_union_value adamic_view_union_heap(const adamic_heap *value);

// Read through the shared readiness and static-owner resolver exactly once.
// Optional absence does not permit an uninitialized present slot.
adamic_view_union_value adamic_object_view_union_snapshot(const adamic_object *object, const char *name, adamic_slot_cache *cache, const char *expression, const char *declared, bool absent);

const char *adamic_view_union_kind_name(adamic_view_union_kind kind);
size_t adamic_view_mixed_union_select(const adamic_view_union_value *value, const adamic_view_union_member *members, size_t count, adamic_view_union_match match, void *context, const char *expression, const char *declared);

#endif
