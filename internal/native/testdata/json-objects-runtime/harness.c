// json_objects.c's null-pointer reading, from the allocation alone: an object whose one reference slot holds a
// null pointer, under each representation byte its last write could leave. Lowering refuses every program whose
// type shows such a field held as a bare pointer (#ckxcnr6), so only this harness reaches the stop.
#include "json_stringify.h"
#include <stdio.h>
#include <string.h>
static const char *const probe_names[] = {"target"};
static const bool probe_references[] = {true};
static const adamic_field_kind probe_kinds[] = {adamic_field_reference};
static const adamic_shape probe_shape = {1, probe_names, probe_references, NULL, probe_kinds};
static const adamic_json_schema probe_reference = {.kind = adamic_json_union};
static void probe_print(unsigned char type) {
 adamic_object *object = adamic_object_new(&probe_shape);
 adamic_object_field_types(object)[0] = type;
 adamic_value empty = {.reference = NULL};
 adamic_string *text = probe_adamic_json_stringify_objects((adamic_value){.reference = object}, &probe_reference, empty, NULL, empty, NULL);
 printf("%u ", (unsigned)type);
 fwrite(text->bytes, 1, text->length, stdout);
 putchar('\n');
 adamic_release(text);
 adamic_release(object);
}
int main(int argc, char **argv) {
 adamic_start(argc, argv);
 if (argc > 1) {
  fflush(stdout);
  probe_print(adamic_ir_Object);
  fflush(stdout);
  return 0;
 }
 const unsigned char types[] = {adamic_ir_NullRepresentation, adamic_ir_UndefinedRepresentation, adamic_ir_String, adamic_ir_Union, adamic_ir_Weak, adamic_ir_Record};
 for (size_t index = 0; index < sizeof types; index++) probe_print(types[index]);
 /* Keep stdout observable when LeakSanitizer exits before stdio flushes it. */
 fflush(stdout);
 return 0;
}
