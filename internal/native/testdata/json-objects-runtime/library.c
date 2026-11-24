// json_objects.c's stop for a runtime library value, read from the shape's methods table alone: one object holding
// the number 2024 under year, its shape plain, a host's (adamic_host_methods) or a RegExp match's groups
// (adamic_groups_methods), as the argument names. A plain object prints; a host's and a groups object stop.
#include "json_stringify.h"
#include <stdio.h>
#include <string.h>
static const char *const library_names[] = {"year"};
static const bool library_references[] = {false};
static const adamic_field_kind library_kinds[] = {adamic_field_number};
static const adamic_shape plain_shape = {1, library_names, library_references, NULL, library_kinds};
static const adamic_shape host_shape = {1, library_names, library_references, &adamic_host_methods, library_kinds};
static const adamic_shape groups_shape = {1, library_names, library_references, &adamic_groups_methods, library_kinds};
static const adamic_json_schema library_reference = {.kind = adamic_json_union};
int main(int argc, char **argv) {
 adamic_start(argc, argv);
 const adamic_shape *shape = &plain_shape;
 if (argc > 1 && strcmp(argv[1], "host") == 0) shape = &host_shape;
 if (argc > 1 && strcmp(argv[1], "groups") == 0) shape = &groups_shape;
 adamic_object *object = adamic_object_new(shape);
 object->slots[0].number = 2024;
 adamic_object_field_types(object)[0] = adamic_ir_Number;
 adamic_value empty = {.reference = NULL};
 adamic_string *text = probe_adamic_json_stringify_objects((adamic_value){.reference = object}, &library_reference, empty, NULL, empty, NULL);
 fwrite(text->bytes, 1, text->length, stdout);
 putchar('\n');
 adamic_release(text);
 adamic_release(object);
 /* Keep stdout observable when LeakSanitizer exits before stdio flushes it. */
 fflush(stdout);
 return 0;
}
