#include "view_unions_untagged.h"

static bool tag_matches(const adamic_view_union_value *slot, const adamic_view_union_member *allowed) {
    if (!allowed->literal || slot->kind != allowed->kind) { return false; }
    switch (slot->kind) {
    case adamic_view_union_number: return slot->payload.number == allowed->value.number;
    case adamic_view_union_boolean: return slot->payload.boolean == allowed->value.boolean;
    case adamic_view_union_string:
        return slot->payload.reference != NULL && allowed->value.reference != NULL && adamic_string_equal(slot->payload.reference, allowed->value.reference);
    default: return false;
    }
}

size_t adamic_view_untagged_union_select(const adamic_view_union_value *value, const adamic_view_untagged_member *members, size_t count, adamic_view_untagged_probe probe, adamic_view_union_match match, void *context, const char *expression, const char *declared) {
    if (value->kind == adamic_view_union_object && value->payload.reference != NULL && match != NULL) {
        for (size_t index = 0; index < count; index++) {
            const adamic_view_untagged_member *member = &members[index];
            if (member->contract == 0) { continue; }
            bool candidate = true;
            for (size_t tag = 0; tag < member->count; tag++) {
                const adamic_view_untagged_tag *constraint = &member->tags[tag];
                adamic_view_union_value slot = {0};
                if (probe == NULL || !probe(context, value, constraint->field, &slot)) { candidate = false; break; }
                bool accepted = false;
                for (size_t literal = 0; literal < constraint->count; literal++) {
                    if (tag_matches(&slot, &constraint->allowed[literal])) { accepted = true; break; }
                }
                if (!accepted) { candidate = false; break; }
            }
            if (!candidate) { continue; }
            adamic_view_union_member descriptor = { .kind = adamic_view_union_object, .contract = member->contract };
            if (match(context, &descriptor, value)) { return member->contract; }
        }
    }
    // Reuse the common field/expected/found failure text and exit-70 path.
    return adamic_view_mixed_union_select(value, NULL, 0, NULL, NULL, expression, declared);
}
