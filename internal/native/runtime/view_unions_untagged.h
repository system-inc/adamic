#ifndef ADAMIC_VIEW_UNIONS_UNTAGGED_H
#define ADAMIC_VIEW_UNIONS_UNTAGGED_H
#include "view_unions_mixed.h"

typedef struct adamic_view_untagged_tag {
    const char *field;
    const adamic_view_union_member *allowed;
    size_t count;
} adamic_view_untagged_tag;

typedef struct adamic_view_untagged_member {
    size_t contract;
    const adamic_view_untagged_tag *tags;
    size_t count;
} adamic_view_untagged_member;

// The owning slot probe returns false for missing, uninitialized or unsupported
// storage. It must neither panic nor evaluate getters or user code. No new
// readiness state is owned here. Returned normalized snapshots are borrowed.
typedef bool (*adamic_view_untagged_probe)(void *context, const adamic_view_union_value *object, const char *field, adamic_view_union_value *slot);

// Complete structural membership is supplied by the shared contract adapter.
// A tag narrows candidates only. Return the selected contract, not its position,
// so dispatch can preserve it on aliases and every subsequent transitive read.
size_t adamic_view_untagged_union_select(const adamic_view_union_value *value, const adamic_view_untagged_member *members, size_t count, adamic_view_untagged_probe probe, adamic_view_union_match match, void *context, const char *expression, const char *declared);

/* Plain-object adapters reuse the shared initialization and representation
 * bytes. Class/static/accessor storage is not certified by this adapter. */
typedef struct adamic_view_untagged_field {
    const char *name;
    size_t contract;
    bool optional;
} adamic_view_untagged_field;
typedef struct adamic_view_untagged_contract {
    unsigned int kind;
    unsigned int of;
    bool undefined;
    const adamic_view_union_member *allowed;
    size_t allowed_count;
    const adamic_view_untagged_field *fields;
    size_t field_count;
    const size_t *members;
    size_t member_count;
} adamic_view_untagged_contract;
bool adamic_view_untagged_plain_slot(void *context, const adamic_view_union_value *object, const char *field, adamic_view_union_value *slot);
bool adamic_view_untagged_plain_matches(const adamic_view_untagged_contract *contracts, size_t count, size_t id, const adamic_view_union_value *value);
size_t adamic_view_untagged_plain_select(const adamic_object *object, const adamic_view_untagged_contract *contracts, size_t count, size_t id, const char *expression, const char *declared);

#endif
