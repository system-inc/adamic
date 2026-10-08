// Native layout evidence; no production C is modified.
#include "../../../../internal/native/runtime/adamic.h"
#include <stdio.h>
int main(void) {
    printf("heap=%zu object_header=%zu value=%zu parse_node_13_slots=%zu array_header=%zu string_header=%zu\n", sizeof(adamic_heap), sizeof(adamic_object), sizeof(adamic_value), sizeof(adamic_object) + 13 * sizeof(adamic_value), sizeof(adamic_array), sizeof(adamic_string));
    return 0;
}
