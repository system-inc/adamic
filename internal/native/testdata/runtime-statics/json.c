#include "harness.h"
#include "json_stringify.h"
static adamic_string name = ADAMIC_STRING("value");
static const char *const names[] = {"value"};
static const bool references[] = {false};
static const adamic_field_kind shape_kinds_0[] = {adamic_field_number};
static const adamic_shape shape = {1, names, references, NULL, shape_kinds_0};
static const adamic_json_schema number = {.kind = adamic_json_number};
static const adamic_json_field fields[] = {{&name, 0, &number}};
static const adamic_json_schema schema = {.kind = adamic_json_object, .count = 1, .fields = fields};
static void prepare(void) {}
static void cleanup(void) {}
static double exercise(size_t index) {
 (void)index;
 adamic_object *object = adamic_object_new(&shape);
 object->slots[0].number = 1.25;
 adamic_string *json = adamic_json_stringify((adamic_value){.reference = object}, &schema, (adamic_value){0}, NULL, (adamic_value){.number = 2}, &number);
 const char expected[] = "{\n  \"value\": 1.25\n}";
 if (json == NULL || json->length != sizeof expected - 1 || memcmp(json->bytes, expected, sizeof expected - 1) != 0) { abort(); }
 adamic_release(json); adamic_release(object);
 return 1;
}
