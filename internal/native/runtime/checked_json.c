#include "checked_json.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static const char *actual(const adamic_heap *value) {
 if (value == NULL) return "undefined";
 if (value == &adamic_null) return "null";
 switch (value->kind) {
 case adamic_kind_number: return "number";
 case adamic_kind_boolean: return "boolean";
 case adamic_kind_string: return "string";
 case adamic_kind_array: return "array";
 case adamic_kind_object: return "object";
 case adamic_kind_closure: return "function";
 default: return "non-JSON value";
 }
}

adamic_heap *adamic_json_array_get(adamic_array *array, double index) {
 const adamic_value *slot = adamic_array_at(array, index);
 if (slot == NULL) return NULL;
 if (array->references) return adamic_retain(slot->reference);
 switch (array->element_type) {
 case 1: return adamic_box_number(slot->number);
 case 2: return slot->boolean ? &adamic_box_true.heap : &adamic_box_false.heap;
 case 7: { adamic_maybe_number n = adamic_maybe_number_unpack(slot->number); return n.present ? adamic_box_number(n.number) : NULL; }
 default: adamic_panic("checked JSON array has no element representation metadata", sizeof "checked JSON array has no element representation metadata" - 1);
 }
}

static char *joined(const char *path, const char *suffix) {
 size_t a = strlen(path), b = strlen(suffix);
 char *result = malloc(a + b + 1);
 if (result == NULL) adamic_panic("out of memory", sizeof "out of memory" - 1);
 memcpy(result, path, a); memcpy(result + a, suffix, b + 1);
 return result;
}
static _Noreturn void mismatch(const char *path, const char *needed, const char *found) {
 const char *prefix = "checked any: ";
 size_t size = strlen(prefix) + strlen(path) + strlen(needed) + strlen(found) + 16;
 char *message = malloc(size);
 if (message == NULL) adamic_panic("out of memory", sizeof "out of memory" - 1);
 int length = snprintf(message, size, "%s%s needs %s, found %s", prefix, path, needed, found);
 adamic_panic(message, (size_t)length);
}

// A dry pass selects union alternatives without changing source diagnostics.
static bool inspect(adamic_heap *value, const adamic_json_contract *contract, const char *path, size_t depth, bool terminal) {
 const char *kind = actual(value);
 bool domain = strcmp(contract->kind, "json") == 0;
 bool literal = strstr(contract->kind, "_literal") != NULL;
 bool same = literal ? strncmp(contract->kind, kind, strlen(kind)) == 0 : domain ? (strcmp(kind,"non-JSON value") != 0 && strcmp(kind,"function") != 0) : strcmp(kind, contract->kind) == 0;
 if (strcmp(contract->kind, "union") == 0) {
  for (size_t i = 0; i < contract->count; i++) {
   if (inspect(value, contract->alternatives[i], path, depth, false)) return true;
  }
  if (terminal) mismatch(path, contract->name, kind);
  return false;
 }
 if (!same || depth > 64) {
  if (terminal) mismatch(path, contract->name, depth > 64 ? "non-JSON recursion" : kind);
  return false;
 }
 if(literal) {
  bool matches=false;
  if(strcmp(kind,"number")==0) matches=((adamic_number_box *)value)->number==contract->literal_number;
  if(strcmp(kind,"boolean")==0) matches=((adamic_boolean_box *)value)->boolean==contract->literal_boolean;
  if(strcmp(kind,"string")==0) {adamic_string *text=(adamic_string *)value;matches=text->length==contract->literal_length&&memcmp(text->bytes,contract->literal_text,text->length)==0;}
  if(!matches) {if(terminal)mismatch(path,contract->name,kind);return false;}
 }
 if (strcmp(kind, "number") == 0 && !isfinite(((adamic_number_box *)value)->number)) {
  if (terminal) mismatch(path, contract->name, "non-JSON number");
  return false;
 }
 if (strcmp(kind, "object") == 0) {
  const adamic_object *object = (const adamic_object *)value;
  if (object->class != NULL) {
   if (terminal) mismatch(path, contract->name, "class object");
   return false;
  }
  if (domain) {
   for (size_t i = 0; i < object->shape->count; i++) {
    const char *name = object->shape->names[i];
    char *dot=joined(".", name), *child_path=joined(path,dot); free(dot);
    adamic_heap *child = adamic_object_initialized(object)[i] ? adamic_dynamic_property(value,name) : NULL;
    bool valid=inspect(child, contract, child_path, depth+1, terminal);
    adamic_release(child);free(child_path);if(!valid)return false;
   }
  }
  for (size_t i = 0; i < contract->count; i++) {
   const adamic_json_contract_field *field = &contract->fields[i];
   char *dot = joined(".", field->name), *child_path = joined(path, dot); free(dot);
   adamic_heap *child = adamic_dynamic_property(value, field->name);
   bool valid = (field->optional && child == NULL) || inspect(child, field->contract, child_path, depth + 1, terminal);
   adamic_release(child); free(child_path);
   if (!valid) return false;
  }
 }
 if (strcmp(kind, "array") == 0 && (contract->element != NULL || domain)) {
  adamic_array *array = (adamic_array *)value;
  for (size_t i = 0; i < array->length; i++) {
   char suffix[64]; (void)snprintf(suffix, sizeof suffix, "[%zu]", i);
   char *child_path = joined(path, suffix);
   adamic_heap *child = adamic_json_array_get(array, (double)i);
   bool valid = inspect(child, domain ? contract : contract->element, child_path, depth + 1, terminal);
   adamic_release(child); free(child_path);
   if (!valid) return false;
  }
 }
 return true;
}
void adamic_check_json(adamic_heap *value, const adamic_json_contract *contract, const char *path) {
 static const adamic_json_contract domain={"json", "JSON value | undefined", NULL, 0, NULL, NULL, NULL, 0, 0, false};
 (void)inspect(value, &domain, path, 0, true);
 (void)inspect(value, contract, path, 0, true);
}
