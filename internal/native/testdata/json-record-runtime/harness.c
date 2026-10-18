#include "json_stringify.h"
#include <stdio.h>
#include <string.h>
static const adamic_json_schema merge_number = {.kind = adamic_json_number};
static const adamic_json_schema merge_absent = {.kind = adamic_json_undefined};
static const adamic_json_schema merge_hook = {.kind = adamic_json_toJSON};
static const adamic_json_schema merge_numbers = {.kind = adamic_json_record, .element = &merge_number};
static const adamic_json_schema merge_hooks = {.kind = adamic_json_record, .element = &merge_hook};
static const adamic_json_schema merge_cycle = {.kind = adamic_json_record, .element = &merge_cycle};
static const adamic_shape merge_shape = {0, NULL, NULL, NULL};
static adamic_record *merge_value;
static bool merge_throwing;
static adamic_string *merge_text(const char *text) {
 size_t length = strlen(text);
 adamic_string *result = adamic_string_allocate(length);
 memcpy((char *)result->bytes, text, length);
 return result;
}
static adamic_json_result merge_to_json(adamic_object *receiver, const adamic_string *key) {
 printf("key:%.*s\n", (int)key->length, key->bytes);
 if (key->length == 4 && memcmp(key->bytes, "omit", 4) == 0) return (adamic_json_result){{.reference = NULL}, &merge_absent};
 if (merge_throwing) {
  adamic_thrown = adamic_retain(receiver);
  return (adamic_json_result){{.reference = NULL}, &merge_absent};
 }
 return (adamic_json_result){{.reference = adamic_retain(merge_value)}, &merge_numbers};
}
static const adamic_json_runtime merge_provider = {NULL, NULL, merge_to_json};
static void merge_print(adamic_record *record, const adamic_json_schema *schema) {
 adamic_value empty = {.reference = NULL};
 adamic_string *text = adamic_json_stringify_runtime((adamic_value){.reference = record}, schema, empty, NULL, empty, NULL, &merge_provider);
 if (adamic_thrown != NULL) {
  adamic_release(adamic_thrown); adamic_thrown = NULL; puts("caught");
 } else if (text == NULL) puts("undefined");
 else { fwrite(text->bytes, 1, text->length, stdout); putchar('\n'); }
 adamic_release(text);
}
int main(int argc, char **argv) {
 adamic_start(argc, argv);
 if (argc > 1) {
  adamic_record *cycle = adamic_record_new(true);
  adamic_record_define(cycle, merge_text("self"), (adamic_value){.reference = adamic_retain(cycle)});
  merge_print(cycle, &merge_cycle);
  return 0;
 }
 merge_value = adamic_record_new(false);
 adamic_record_define(merge_value, merge_text("x"), (adamic_value){.number = 7});
 adamic_object *receiver = adamic_object_new(&merge_shape);
 adamic_record *record = adamic_record_new(true);
 const char *names[] = {"omit", "nested", "again"};
 for (size_t index = 0; index < 3; index++) adamic_record_define(record, merge_text(names[index]), (adamic_value){.reference = adamic_retain(receiver)});
 merge_print(record, &merge_hooks);
 merge_throwing = true;
 merge_print(record, &merge_hooks);
 merge_print(NULL, &merge_numbers);
 adamic_release(record); adamic_release(receiver); adamic_release(merge_value); merge_value = NULL;
 /* Keep stdout observable when LeakSanitizer exits before stdio flushes it. */
 fflush(stdout);
 return 0;
}
