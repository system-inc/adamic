#include "adamic.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

typedef struct intersection_seen {
 const adamic_object *object;
 size_t id;
 const struct intersection_seen *previous;
} intersection_seen;

static bool literal_matches(adamic_value value, unsigned char type, const adamic_intersection_literal *literal) {
 if (type == 1) return value.number == literal->number;
 if (type == 2) return value.boolean == literal->boolean;
 if (type == 3) {
  const adamic_string *string = value.reference;
  size_t length = literal->length;
  return string != NULL && string->length == length && memcmp(string->bytes, literal->string, length) == 0;
 }
 return false;
}

static void require_object(const adamic_object *object, const adamic_intersection_contract *contracts, size_t id, const char *path, const intersection_seen *previous) {
 if (object == NULL) return;
 for (const intersection_seen *seen = previous; seen != NULL; seen = seen->previous) {
  if (seen->object == object && seen->id == id) return;
 }
 intersection_seen seen = {object, id, previous};
 const adamic_intersection_contract *contract = &contracts[id];
 for (size_t i = 0; i < contract->count; i++) {
  const adamic_intersection_field *field = &contract->fields[i];
  size_t capacity = strlen(path) + strlen(field->name) + 2;
  char *next = malloc(capacity);
  if (next == NULL) { static const char message[] = "out of memory"; adamic_panic(message, sizeof message - 1); }
  (void)snprintf(next, capacity, "%s.%s", path, field->name);
  adamic_slot_cache cache = {0};
  adamic_value value = adamic_object_optional_view_undefined(object, field->name, &cache, field->type, field->expected, next, field->optional, false, field->undefined_member);
  unsigned char type = field->type;
  bool present = true;
  if (type == 7) { adamic_maybe_number number = adamic_maybe_number_unpack(value.number); present = number.present; value.number = number.number; type = 1; }
  if (type == 9) { present = value.reference != NULL; value.boolean = present && ((const adamic_boolean_box *)value.reference)->boolean; type = 2; }
  if (type == 3 || type == 4) present = value.reference != NULL;
  if (present && field->allowed_count != 0) {
   bool match = false;
   for (size_t j = 0; j < field->allowed_count; j++) match = match || literal_matches(value, type, &field->allowed[j]);
   if (!match) adamic_view_literal_failure(next, field->expected, type, value);
  }
  if (present && field->child != 0) require_object(value.reference, contracts, field->child, next, &seen);
  free(next);
 }
}

void adamic_intersection_recursive_require(const adamic_object *object, const adamic_intersection_contract *contracts, size_t id, const char *path) {
 require_object(object, contracts, id, path, NULL);
}

// The bounded walk descends into each descriptor once per path, as the inline
// field walk does: a descriptor already active is checked for its kind by the
// field read and leaves its fields to their own checked reads.
typedef struct bounded_active {
 size_t id;
 const struct bounded_active *previous;
} bounded_active;

static char *bounded_path(const char *path, const char *name) {
 size_t capacity = strlen(path) + strlen(name) + 2;
 char *next = malloc(capacity);
 if (next == NULL) { static const char message[] = "out of memory"; adamic_panic(message, sizeof message - 1); }
 (void)snprintf(next, capacity, "%s.%s", path, name);
 return next;
}

static void require_bounded(const adamic_object *object, const adamic_intersection_contract *contracts, size_t id, const char *path, const bounded_active *previous) {
 if (object == NULL) return;
 for (const bounded_active *active = previous; active != NULL; active = active->previous) {
  if (active->id == id) return;
 }
 bounded_active here = {id, previous};
 const adamic_intersection_contract *contract = &contracts[id];
 if (contract->tag != NULL) {
  char *next = bounded_path(path, contract->tag);
  adamic_slot_cache cache = {0};
  adamic_value value = adamic_object_view(object, contract->tag, &cache, contract->tag_type, contract->tag_expected, next);
  for (size_t i = 0; i < contract->arm_count; i++) {
   const adamic_intersection_arm *arm = &contract->arms[i];
   bool match = false;
   for (size_t j = 0; j < arm->allowed_count; j++) match = match || literal_matches(value, contract->tag_type, &arm->allowed[j]);
   if (!match) continue;
   free(next);
   if (arm->contract != 0) require_bounded(object, contracts, arm->contract, path, &here);
   return;
  }
  adamic_view_literal_failure(next, contract->tag_expected, contract->tag_type, value);
 }
 for (size_t i = 0; i < contract->count; i++) {
  const adamic_intersection_field *field = &contract->fields[i];
  char *next = bounded_path(path, field->name);
  adamic_slot_cache cache = {0};
  adamic_value value = adamic_object_optional_view_undefined(object, field->name, &cache, field->type, field->expected, next, field->optional, false, field->undefined_member);
  unsigned char type = field->type;
  bool present = true;
  if (type == 7) { adamic_maybe_number number = adamic_maybe_number_unpack(value.number); present = number.present; value.number = number.number; type = 1; }
  if (type == 9) { present = value.reference != NULL; value.boolean = present && ((const adamic_boolean_box *)value.reference)->boolean; type = 2; }
  if (type >= 3 && type <= 6) present = value.reference != NULL;
  if (type == 8) present = value.reference != NULL;
  if (present && field->allowed_count != 0) {
   bool match = false;
   for (size_t j = 0; j < field->allowed_count; j++) match = match || literal_matches(value, type, &field->allowed[j]);
   if (!match) adamic_view_literal_failure(next, field->expected, type, value);
  }
  if (present && field->child != 0) require_bounded(value.reference, contracts, field->child, next, &here);
  free(next);
 }
}

void adamic_intersection_bounded_require(const adamic_object *object, const adamic_intersection_contract *contracts, size_t id, const char *path) {
 require_bounded(object, contracts, id, path, NULL);
}
