#define ADAMIC_CANONICAL_CLOSURES 1
#include "adamic.h"
#include <assert.h>
#include <stdio.h>
#include <string.h>

static const char *const names[] = {"next", "text", "counted", "empty"};
static const bool references[] = {true, true, true, true};
static const adamic_field_kind shape_kinds[] = {adamic_field_reference, adamic_field_reference, adamic_field_reference, adamic_field_reference};
static const adamic_shape shape = {4, names, references, NULL, shape_kinds};
static const char *const leaf_names[] = {"text"};
static const bool leaf_references[] = {true};
static const adamic_field_kind leaf_shape_kinds[] = {adamic_field_reference};
static const adamic_shape leaf_shape = {1, leaf_names, leaf_references, NULL, leaf_shape_kinds};

static adamic_object *member(void) {
    adamic_object *object = adamic_object_new(&shape);
    assert(object->heap.references == 1);
    ((unsigned char *)object)[adamic_object_size(shape.count) - 1] = 7;
    adamic_object_initialized(object)[3] = 0;
    adamic_object_field_types(object)[1] = 3;
    object = adamic_program_adopt_owned(object, adamic_object_size(shape.count));
    /* Pin every tail, including the last byte of the physical allocation. */
    assert(((unsigned char *)object)[adamic_object_size(shape.count) - 1] == 7);
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
/* Deliberately perturb allocation-site membership, independently of lowering.
 * Extra members live until the Program ends; omitted ones keep ordinary owners. */
static int membership(const char *mode) {
    bool over = strcmp(mode, "overinclude") == 0;
    adamic_object *root;
    adamic_object *child = adamic_object_new(&shape);
    if (over) {
        child = adamic_program_adopt_owned(child, adamic_object_size(shape.count));
        root = child;
    } else {
        if (strcmp(mode, "underinclude-graph") == 0) {
            child = adamic_graph_adopt_owned(child, adamic_object_size(shape.count));
            child->slots[0].reference = adamic_graph_hold(child, child);
        }
        root = member();
        child->slots[2].reference = adamic_retain(root);
        root->slots[0].reference = adamic_graph_hold(root, child);
    }
    adamic_string *text = adamic_string_from_number(12345);
    child->slots[1].reference = text;
    adamic_weak *weak = adamic_weak_of(child);
    adamic_release(child);
    adamic_release(root);
    /* Both faults must preserve this read, including the counted back edge. */
    adamic_string *held = child->slots[1].reference;
    assert(held->length == 5 && memcmp(held->bytes, "12345", 5) == 0);
    assert(adamic_weak_target(weak) == child);
    puts("membership 12345");
    adamic_program_region_end();
    assert(adamic_weak_target(weak) == NULL);
    adamic_release(weak);
    assert(adamic_counted.allocations == adamic_counted.frees + adamic_counted.regions);
    return 0;
}
static adamic_object *adopted(bool graph) {
    return adamic_program_adopt_or_fallback(adamic_object_new(&shape), adamic_object_size(shape.count), graph);
}
/* Two Programs through the switch (#64pg02g), each site adopting as lowering emits it. Only the
 * first Program's sites become members. The second's cycle is one graph region, freed when its
 * last owner lets go, and its acyclic site is counted. A second-generation region a member points
 * into is retained until exit by design, so stderr names the second Program's frees and the Go
 * test holds them to exactly the three values nothing first-generation points at. */
static int generations(void) {
    adamic_program_region_begin();
    adamic_object *program = adopted(true);
    adamic_object *file = adopted(true);
    program->slots[0].reference = adamic_graph_hold(program, file);
    file->slots[0].reference = adamic_graph_hold(file, program);
    file->slots[1].reference = adamic_string_from_number(12345);

    adamic_program_region_begin();
    size_t before = (size_t)adamic_counted.frees;
    adamic_object *next = adopted(true);
    adamic_object *next_file = adopted(true);
    next->slots[0].reference = adamic_graph_hold(next, next_file);
    next_file->slots[0].reference = adamic_graph_hold(next_file, next);
    /* An edge into a first-generation member needs no count: the file is carried across. */
    next->slots[2].reference = adamic_graph_hold(next, file);
    adamic_object *counted = adopted(false);
    counted->slots[1].reference = adamic_string_from_number(678);
    /* The first Program's lazily built part, asked only after the second begin, and the
     * boundary inside one strongly connected group: a member and a second-generation graph
     * value pointing at each other. The member's edge is a real retain until exit. */
    adamic_object *lazy = adopted(true);
    lazy->slots[0].reference = adamic_graph_hold(lazy, program);
    program->slots[3].reference = adamic_graph_hold(program, lazy);
    adamic_release(lazy);

    adamic_string *carried = ((adamic_object *)next->slots[2].reference)->slots[1].reference;
    printf("generations %.*s\n", (int)carried->length, carried->bytes);
    adamic_release(next);
    adamic_release(next_file);
    adamic_release(counted);
    fprintf(stderr, "adamic: second Program frees %zu\n", (size_t)adamic_counted.frees - before);
    adamic_string *text = ((adamic_object *)program->slots[0].reference)->slots[1].reference;
    adamic_object *asked = program->slots[3].reference;
    printf("first %.*s %s\n", (int)text->length, text->bytes, asked->slots[0].reference == program ? "lazy" : "lost");
    adamic_release(program);
    adamic_release(file);
    return 0; /* heap's atexit ends the region and frees what the members retained. */
}
int main(int argc, char **argv) {
    assert(argc == 2);
    if (strcmp(argv[1], "overinclude") == 0 || strncmp(argv[1], "underinclude-", 13) == 0) { return membership(argv[1]); }
    if (strcmp(argv[1], "generations") == 0) { return generations(); }
    if (strcmp(argv[1], "aliased-adoption") == 0) {
        adamic_object *owner = adamic_object_new(&shape);
        adamic_object *alias = adamic_retain(owner);
        owner = adamic_program_adopt_owned(owner, adamic_object_size(shape.count));
        assert(alias->shape->count == shape.count);
        adamic_release(alias); adamic_release(owner);
        return 0;
    }
    if (strcmp(argv[1], "cached-adoption") == 0) {
        adamic_environment *owner = adamic_environment_new(1);
        (void)adamic_closure_canonical(&owner->cells[0], code, 0, NULL);
        (void)adamic_program_adopt_owned(owner, sizeof *owner + sizeof owner->cells[0]);
        return 0;
    }
    if (strcmp(argv[1], "unmovable-adoption") == 0) {
        adamic_heap *box = adamic_box_number(7);
        (void)adamic_program_adopt_owned(box, sizeof(adamic_number_box));
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
    assert(environment->cells[0].owner == &environment->heap);
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
