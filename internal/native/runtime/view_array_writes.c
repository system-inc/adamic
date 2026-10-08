#include "adamic.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

void adamic_view_array_object_certificate(adamic_object *object, unsigned int contract) { object->array_write_contract = contract; }
void adamic_view_array_element_certificate(adamic_array *array, unsigned int contract) { array->element_contract = contract; }

static _Noreturn void array_reference_failure(const char *expected, const char *found) {
 size_t size = strlen(expected) + strlen(found) + 80;
 char *message = malloc(size);
 if (message == NULL) { static const char oom[] = "out of memory"; adamic_panic(oom, sizeof oom - 1); }
 int length = snprintf(message,size,"element write failed: <array write> expected %s, found %s",expected,found);
 adamic_panic(message,(size_t)length);
}

void adamic_view_array_reference_write(const adamic_array *array, const adamic_object *value, const adamic_array_write_pair *pairs, size_t count, const char *const *names, size_t name_count) {
 if (array->element_kind != 10 || !array->references) { adamic_view_array_storage_check(array,4,"<array write>"); }
 unsigned int target = array->element_contract;
 if (target == 0 || target >= name_count) array_reference_failure("object","uncertified source element contract");
 if (value == NULL || value->heap.kind != adamic_kind_object) array_reference_failure(names[target],value == NULL ? "undefined" : "non-object");
 unsigned int source = value->array_write_contract;
 if (source == 0 || source >= name_count) array_reference_failure(names[target],"uncertified incoming record contract");
 bool compatible = false;
 for (size_t index = 0; index < count; index++) compatible = compatible || (pairs[index].source == source && pairs[index].target == target);
 if (!compatible) array_reference_failure(names[target],names[source]);
}
