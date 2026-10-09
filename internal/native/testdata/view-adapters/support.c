// The cache tests link the actual heap, weak, closure and count runtime units.
// Their fixtures allocate only closures, cells and weak handles. Reaching an
// unrelated destructor is a test error, rather than a substitute implementation.
#include "adamic.h"
#include <stdio.h>
#include <stdlib.h>

_Noreturn void adamic_panic(const char *message, size_t length) {
    fprintf(stderr, "adamic: panic: %.*s\n", (int)length, message);
    _Exit(70);
}

void adamic_string_free_index(adamic_string *string) { (void)string; abort(); }
void adamic_object_free_children(adamic_object *object, void (*release)(void *)) { (void)object; (void)release; abort(); }
void adamic_map_free_children(adamic_map *map, void (*release)(void *)) { (void)map; (void)release; abort(); }
