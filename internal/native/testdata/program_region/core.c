#define ADAMIC_CANONICAL_CLOSURES 1
#include "adamic.h"
#include <assert.h>
#include <stdio.h>
#include <string.h>

static const char *const names[] = {"next", "text", "counted", "empty"};
static const bool references[] = {true, true, true, true};
static const adamic_shape shape = {4, names, references, NULL};
static const char *const leaf_names[] = {"text"};
static const bool leaf_references[] = {true};
static const adamic_shape leaf_shape = {1, leaf_names, leaf_references, NULL};

static adamic_object *member(void) {
    adamic_object *object = adamic_object_new(&shape);
    assert(object->heap.references == 1);
    adamic_object_orders(object)[3] = 0;
    adamic_object_initialized(object)[3] = 0;
    adamic_object_field_types(object)[1] = 3;
    object = adamic_program_adopt_owned(object, adamic_object_size(shape.count));
    /* Pin every tail, including the last byte of the physical allocation. */
    assert(adamic_object_orders(object)[0] == 1);
    assert(adamic_object_orders(object)[3] == 0);
    assert(adamic_object_initialized(object)[0] == 1);
    assert(adamic_object_initialized(object)[3] == 0);
    assert(adamic_object_field_types(object)[1] == 3);
    adamic_object_field_types(object)[3] = 4;
    return object;
}
static void no_count(void *value) {
    adamic_heap *heap = value;
    assert(adamic_program_is(value) && heap->references == 0);
    assert(adamic_retain(value) == value);
    assert(adamic_retain_slow(value) == value);
    adamic_release(value);
    adamic_release_slow(value);
    adamic_release(value);
    assert(heap->references == 0);
}
static adamic_value code(adamic_closure *closure, adamic_value *arguments) {
    (void)arguments;
    adamic_string *text = closure->cells[0]->value.reference;
    return (adamic_value){.number = (double)text->length};
}
int main(int argc, char **argv) {
    assert(argc == 2);
    if (strcmp(argv[1], "cached-adoption") == 0) {
        adamic_environment *owner = adamic_environment_new(1);
        (void)adamic_closure_canonical(&owner->cells[0], code, 0, NULL);
        (void)adamic_program_adopt_owned(owner, sizeof *owner + sizeof owner->cells[0]);
        return 0;
    }
    if (strcmp(argv[1], "weak-before-adoption") == 0) {
        adamic_object *owner = adamic_object_new(&shape);
        (void)adamic_weak_of(owner);
        (void)adamic_program_adopt_owned(owner, adamic_object_size(shape.count));
        return 0;
    }
    if (strcmp(argv[1], "weak-cell-before-adoption") == 0) {
        adamic_environment *owner = adamic_environment_new(1);
        (void)adamic_weak_of(&owner->cells[0]);
        (void)adamic_program_adopt_owned(owner, sizeof *owner + sizeof owner->cells[0]);
        return 0;
    }
    adamic_object *first = member();
    if (strcmp(argv[1], "share") == 0) { adamic_share(first); return 0; }
    if (strcmp(argv[1], "share-cell") == 0) {
        adamic_environment *environment = adamic_environment_new(1);
        environment = adamic_program_adopt_owned(environment, sizeof *environment + sizeof environment->cells[0]);
        adamic_share(&environment->cells[0]); return 0;
    }
    adamic_object *second = member();
    first->slots[0].reference = adamic_retain(second);
    second->slots[0].reference = adamic_retain(first);

    adamic_string *text = adamic_string_from_number(12345);
    adamic_object *leaf = adamic_object_new(&leaf_shape);
    leaf->slots[0].reference = adamic_retain(text);
    first->slots[1].reference = adamic_retain(text);
    first->slots[2].reference = leaf; /* Move its only counted owner into the member. */
    second->slots[1].reference = adamic_retain(text);

    adamic_environment *environment = adamic_environment_new(2);
    environment->cells[0].ready = true;
    environment->cells[0].references = true;
    environment->cells[0].value.reference = adamic_retain(text);
    environment->cells[1].ready = true;
    environment->cells[1].value.number = 42;
    environment = adamic_program_adopt_owned(environment,
        sizeof *environment + environment->count * sizeof environment->cells[0]);
    assert(environment->cells[0].owner == environment);
    assert(environment->cells[0].ready && environment->cells[1].value.number == 42);

    adamic_map *map = adamic_map_new(false, true);
    adamic_map_set(map, (adamic_value){.number = 7}, (adamic_value){.reference = adamic_retain(text)});
    map = adamic_program_adopt_owned(map, sizeof *map);
    adamic_map_set(map, (adamic_value){.number = 8}, (adamic_value){.reference = adamic_retain(first)});

    adamic_weak *weak = adamic_weak_of(first);
    adamic_weak *cell_weak = adamic_weak_of(&environment->cells[0]);
    adamic_weak *text_weak = adamic_weak_of(text);
    adamic_weak *leaf_weak = adamic_weak_of(leaf);
    assert(adamic_weak_target(weak) == first);
    assert(adamic_weak_target(cell_weak) == &environment->cells[0]);

    /* A canonical counted closure can cache on the already-adopted environment,
     * then unlink before the Program ends. No canonical address is moved. */
    adamic_cell *cells[] = {&environment->cells[0]};
    adamic_closure *closure = adamic_closure_canonical(&environment->cells[1], code, 1, cells);
    assert(adamic_closure_call(closure, NULL, 0).number == 5);
    adamic_release(closure);
    assert(environment->functions == NULL);

    no_count(first); no_count(second); no_count(environment); no_count(&environment->cells[0]); no_count(map);
    assert(adamic_reference_count(&leaf->heap) == 1);
    assert(adamic_reference_count(&text->heap) == 6);
    adamic_release(text); /* Five outside-child edges remain until Program end. */
    assert(((adamic_string *)adamic_map_get(map, (adamic_value){.number = 7})->reference)->length == 5);
    assert(adamic_weak_target(text_weak) != NULL && adamic_weak_target(leaf_weak) == leaf);
    puts("objects 12345 12345");
    puts("environment 12345 42");
    puts("map 12345 cycle true");
    if (strcmp(argv[1], "exit") != 0) {
        adamic_program_region_end();
        assert(adamic_weak_target(weak) == NULL);
        assert(adamic_weak_target(cell_weak) == NULL);
        assert(adamic_weak_target(text_weak) == NULL);
        assert(adamic_weak_target(leaf_weak) == NULL);
        adamic_program_region_end(); /* No double release or second accounting. */
    }
    adamic_release(weak); adamic_release(cell_weak); adamic_release(text_weak); adamic_release(leaf_weak);
    return 0; /* heap's existing atexit owns final region and allocator cleanup. */
}
