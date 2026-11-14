// Assign copies the slot's interpreting byte, including nullish tags but not the private error sentinel.
#include "json_stringify.h"
#include <stdio.h>
#include <string.h>
static const char *const assign_names[] = {"b"};
static const bool assign_references[] = {true};
static const adamic_field_kind assign_kinds[] = {adamic_field_reference};
static const adamic_shape assign_shape = {1, assign_names, assign_references, NULL, assign_kinds};
static const adamic_json_schema assign_reference = {.kind = adamic_json_union};
static const char *const child_names[] = {"n"};
static const bool child_references[] = {false};
static const adamic_field_kind child_kinds[] = {adamic_field_number};
static const adamic_shape child_shape = {1, child_names, child_references, NULL, child_kinds};
static adamic_string *digit(char value) {
 adamic_string *text = adamic_string_allocate(1);
 ((char *)text->bytes)[0] = value;
 return text;
}
int main(int argc, char **argv) {
 adamic_start(argc, argv);
 if (argc != 2) return 1;
 adamic_object *target = adamic_object_new(&assign_shape);
 adamic_object *source = adamic_object_new(&assign_shape);
 unsigned char expected;
 if (strcmp(argv[1], "null") == 0) {
  adamic_object *child = adamic_object_new(&child_shape);
  child->slots[0].number = 1;
  adamic_object_field_types(child)[0] = adamic_ir_Number;
  target->slots[0].reference = child;
  adamic_object_field_types(target)[0] = adamic_ir_Object;
  expected = adamic_ir_NullRepresentation;
  adamic_object_field_types(source)[0] = expected;
 } else {
  target->slots[0].reference = digit('1');
  adamic_object_field_types(target)[0] = adamic_ir_String;
  if (strcmp(argv[1], "undefined") == 0) {
   expected = adamic_ir_UndefinedRepresentation;
   adamic_object_field_types(source)[0] = expected;
  } else {
   source->slots[0].reference = digit('7');
   if (strcmp(argv[1], "sentinel") == 0) {
    // A hand-built source exercises the helper's normalization independently of error enumeration.
    adamic_object_field_types(source)[0] = adamic_error_unwritten;
    expected = adamic_ir_String;
   } else {
    if (strcmp(argv[1], "zero") != 0) return 2;
    adamic_object_field_types(target)[0] = 0;
    expected = 0;
   }
  }
 }
 probe_adamic_object_assign(target, source);
 if (adamic_object_field_types(target)[0] != expected) {
  adamic_release(source);
  adamic_release(target);
  return 3;
 }
 printf("%u", (unsigned)expected);
 if (expected != 0) {
  adamic_value empty = {.reference = NULL};
  adamic_string *text = probe_adamic_json_stringify_objects((adamic_value){.reference = target}, &assign_reference, empty, NULL, empty, NULL);
  putchar(' ');
  fwrite(text->bytes, 1, text->length, stdout);
  adamic_release(text);
 } else if (target->slots[0].reference != source->slots[0].reference) return 4;
 putchar('\n');
 adamic_release(source);
 adamic_release(target);
 fflush(stdout);
 return 0;
}
