#ifndef ADAMIC_VIEW_UNIONS_UNTAGGED_H
#define ADAMIC_VIEW_UNIONS_UNTAGGED_H
#include "view_unions_mixed.h"
#include "view_callables_contract.h"

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

// The plain slot probe returns false for missing or uninitialized slots and
// unsupported receivers. Present unsupported storage returns true with Unknown
// kind. It neither panics nor evaluates getters or user code. No new
// readiness state is owned here. Returned normalized snapshots are borrowed.
typedef bool (*adamic_view_untagged_probe)(void *context, const adamic_view_union_value *object, const char *field, adamic_view_union_value *slot);

// Complete structural membership is supplied by the shared contract adapter.
// A tag narrows candidates only. Return the selected contract, not its position,
// so dispatch can preserve it on aliases and every subsequent transitive read.
size_t adamic_view_untagged_union_select(const adamic_view_union_value *value, const adamic_view_untagged_member *members, size_t count, adamic_view_untagged_probe probe, adamic_view_union_match match, void *context, const char *expression, const char *declared);

/* Own-data adapters reuse the shared initialization and representation
 * bytes, including initialized class instance fields. Accessor storage is
 * not certified by this adapter. A class constructor's own data statics are,
 * only where its contract sets statics_admitted (view_unions_untagged.c). */
typedef struct adamic_view_untagged_field {
    const char *name;
    size_t contract;
    bool optional;
} adamic_view_untagged_field;
typedef struct adamic_view_untagged_callable_producer {
    adamic_code code;
    const adamic_callable_signature *signature;
#ifdef ADAMIC_CLOSURE_CONVENTION
    adamic_counted_code counted_code;
    bool counted;
#endif
} adamic_view_untagged_callable_producer;
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
    size_t element;
    const adamic_callable_signature *callable;
    const adamic_view_untagged_callable_producer *producers;
    size_t producer_count;
    bool fixed_tuple;
    size_t tuple_count;
    // ir.ViewContract's StaticsAdmitted: a class constructor may stand for this object.
    bool statics_admitted;
} adamic_view_untagged_contract;
// The probe reads plain objects only; a class constructor's slots it refuses.
bool adamic_view_untagged_plain_slot(void *context, const adamic_view_union_value *object, const char *field, adamic_view_union_value *slot);
bool adamic_view_untagged_plain_matches(const adamic_view_untagged_contract *contracts, size_t count, size_t id, const adamic_view_union_value *value);
size_t adamic_view_untagged_plain_select(const adamic_object *object, const adamic_view_untagged_contract *contracts, size_t count, size_t id, const char *expression, const char *declared);

// A write through a view a class constructor may stand for (ir.Program's StaticsAdmittedReceiver)
// stops when object is one: the view reads a class's statics, never writes them.
void adamic_view_static_write(const adamic_object *object, const char *name);

adamic_view_union_value adamic_view_element(const adamic_view_untagged_contract *contracts, size_t count, size_t id, const adamic_array *array, double index, bool relative, unsigned int wanted, const char *where, const char *view_type);

#endif
