// class_features.c: public enumeration excludes private storage.
#include "adamic.h"
#include <string.h>

// ECMAScript enumerates canonical array-index names numerically before other string keys.
static bool adamic_index_name(const char *name, uint64_t *value) {
    if (*name == '\0' || (name[0] == '0' && name[1] != '\0')) { return false; }
    uint64_t number = 0;
    for (const char *digit = name; *digit != '\0'; digit++) {
        if (*digit < '0' || *digit > '9') { return false; }
        number = number * 10 + (unsigned)(*digit - '0');
        if (number >= UINT64_C(4294967295)) { return false; }
    }
    *value = number;
    return true;
}

size_t adamic_public_index(const adamic_shape *shape, size_t position) {
    size_t numeric = 0;
    for (size_t i = 0; i < shape->count; i++) {
        uint64_t value;
        if (!adamic_index_name(shape->names[i], &value)) { continue; }
        numeric++;
        size_t rank = 0;
        for (size_t j = 0; j < shape->count; j++) {
            uint64_t other;
            if (adamic_index_name(shape->names[j], &other) && other < value) { rank++; }
        }
        if (rank == position) { return i; }
    }
    size_t rank = numeric;
    for (size_t i = 0; i < shape->count; i++) {
        uint64_t value;
        if (!adamic_index_name(shape->names[i], &value)) {
            if (rank == position) { return i; }
            rank++;
        }
    }
    return position;
}

adamic_array *adamic_class_object_keys(const adamic_object *object) {
    if (adamic_record_is(object)) { return adamic_record_keys(object); }
    const adamic_shape *shape = object->class == NULL ? object->shape : object->class->public_shape;
    if (object->class != NULL && object->class->is_static) {
        adamic_array *keys = adamic_array_new(shape->count, true);
        for (size_t order = 1; order <= object->shape->count; order++) {
            for (size_t i = 0; i < object->shape->count; i++) {
                size_t flag = object->class->static_flags[i];
                if (flag == 0 || object->slots[flag - 1].number != (double)order) { continue; }
                const char *name = object->shape->names[i];
                if (object->has_captured_stack && strcmp(name, "stack") == 0) { continue; }
                adamic_string *key = adamic_string_allocate(strlen(name));
                memcpy((char *)key->bytes, name, key->length);
                adamic_array_push(keys, (adamic_value){.reference = key});
            }
        }
        return keys;
    }
    adamic_array *keys = adamic_array_new(shape->count, true);
    for (size_t index = 0; index < shape->count; index++) {
        const char *name = shape->names[adamic_public_index(shape, index)];
        if (object->has_captured_stack && strcmp(name, "stack") == 0) { continue; }
        if (object->class != NULL && object->class->is_static) {
            adamic_slot_cache cache = {0};
            adamic_value *slot = adamic_object_find(object, name, &cache);
            (void)slot;
            size_t flag = object->class->static_flags[(size_t)(slot - object->slots)];
            if (flag == 0 || object->slots[flag - 1].number == 0) { continue; }
        }
        adamic_string *key = adamic_string_allocate(strlen(name));
        memcpy((char *)key->bytes, name, key->length);
        adamic_value value = {.reference = key};
        adamic_array_push(keys, value);
    }
    return keys;
}

const adamic_accessor *adamic_accessor_find(const adamic_object *object, const char *name) {
    for (const adamic_class *class = object == NULL ? NULL : object->class; class != NULL; class = class->base) {
        for (size_t i = 0; i < class->accessor_count; i++) {
            if (strcmp(class->accessors[i].name, name) == 0) { return &class->accessors[i]; }
        }
    }
    return NULL;
}

adamic_value adamic_accessor_get(adamic_object *object, const char *name) {
    const adamic_accessor *accessor = adamic_accessor_find(object, name);
    if (accessor == NULL || accessor->get == NULL) {
        static const char message[] = "compiler bug: accessor read without a getter";
        adamic_panic(message, sizeof message - 1);
    }
    return accessor->get(object);
}

void adamic_accessor_set(adamic_object *object, const char *name, adamic_value value, int type) {
    const adamic_accessor *accessor = adamic_accessor_find(object, name);
    if (accessor == NULL || accessor->set == NULL) {
        static const char message[] = "compiler bug: accessor write without a setter";
        adamic_panic(message, sizeof message - 1);
    }
    accessor->set(object, value, type);
}
