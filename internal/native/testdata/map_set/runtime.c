#include "adamic.h"
#include <assert.h>
#include <stdio.h>
#include <string.h>

/* A one-entry map's values(), stepped by hand as an emitted next() call steps it, with the value's
 * representation named by the mode. Values are built at run time, since a literal is immortal and
 * would hide a missing or double release. */
int main(int argc, char **argv) {
    assert(argc == 2);
    const char *mode = argv[1];
    if (strcmp(mode, "string") == 0) {
        /* The control: a counted element steps once, retained by its result, then the iterator is done. */
        adamic_map *map = adamic_map_new(false, true);
        adamic_map_set(map, (adamic_value){.number = 1}, (adamic_value){.reference = adamic_string_from_number(7)});
        adamic_object *iterator = adamic_collection_iterator(map, 2, adamic_ir_Number, adamic_ir_String, false);
        adamic_closure *next = iterator->slots[0].reference;
        adamic_object *step = adamic_closure_call(next, NULL, 0).reference;
        assert(!step->slots[1].boolean);
        assert(adamic_object_field_types(step)[0] == adamic_ir_String && adamic_object_field_types(step)[1] == adamic_ir_Boolean);
        adamic_string *text = step->slots[0].reference;
        printf("%.*s\n", (int)text->length, text->bytes);
        adamic_object *done = adamic_closure_call(next, NULL, 0).reference;
        assert(done->slots[1].boolean && done->slots[0].reference == NULL);
        puts("done");
        adamic_release(step);
        adamic_release(done);
        adamic_release(iterator);
        adamic_release(map);
    } else if (strcmp(mode, "maybe_boolean") == 0) {
        /* MaybeBoolean is no collection element, so the step stops before it reads one. Its true is
         * 1, which retained as a pointer would be address 1. */
        adamic_map *map = adamic_map_new(false, false);
        adamic_value flag = {.number = 0};
        flag.maybe_boolean = 1;
        adamic_map_set(map, (adamic_value){.number = 1}, flag);
        adamic_object *iterator = adamic_collection_iterator(map, 2, adamic_ir_Number, adamic_ir_MaybeBoolean, false);
        adamic_closure *next = iterator->slots[0].reference;
        (void)adamic_closure_call(next, NULL, 0);
        puts("a MaybeBoolean element stepped");
        return 1;
    } else {
        assert(!"unknown mode");
    }
    return 0;
}
