#include "adamic.h"
#include <pthread.h>
#include <stdio.h>
#include <stdlib.h>

static adamic_value method_code(adamic_object *self, adamic_value *arguments) {
 (void)self; (void)arguments;
 return (adamic_value){.number = 11};
}
static adamic_value closure_code(adamic_closure *self, adamic_value *arguments) {
 (void)self; (void)arguments;
 return (adamic_value){.number = 22};
}
static const char *const a_names[] = {"value", "padding"};
static const bool a_references[] = {false, false};
static const char *const method_names[] = {"invoke"};
static const adamic_method method_functions[] = {method_code};
static const adamic_methods methods = {1, method_names, method_functions};
static const adamic_field_kind a_shape_kinds_0[] = {adamic_field_number, adamic_field_number};
static const adamic_shape a_shape = {2, a_names, a_references, &methods, a_shape_kinds_0, NULL, NULL};
static const char *const b_names[] = {"invoke", "value"};
static const bool b_references[] = {true, false};
static const adamic_field_kind b_shape_kinds_0[] = {adamic_field_reference, adamic_field_number};
static const adamic_shape b_shape = {2, b_names, b_references, NULL, b_shape_kinds_0, NULL, NULL};
static adamic_object *objects[2];
static adamic_slot_cache field_cache, method_cache;
static pthread_mutex_t gate = PTHREAD_MUTEX_INITIALIZER;
static pthread_cond_t ready = PTHREAD_COND_INITIALIZER;
static unsigned arrived;

// All readers enter the field-only phase together, before method-cache atomics
// can order later field accesses. The gate adds no ordering within that phase.
static void readers_ready(void) {
 if (pthread_mutex_lock(&gate) != 0) { abort(); }
 arrived++;
 if (pthread_cond_broadcast(&ready) != 0) { abort(); }
 while (arrived != 4) { if (pthread_cond_wait(&ready, &gate) != 0) { abort(); } }
 if (pthread_mutex_unlock(&gate) != 0) { abort(); }
}

static void *read_fields(void *given) {
 size_t worker = (size_t)given;
 readers_ready();
 // Exercise the field cache before method-cache atomics can accidentally order
 // field accesses. TSan may model relaxed atomics more strongly than C does.
 for (size_t i = 0; i < 10000; i++) {
  size_t kind = (i + worker) % 2;
  if (adamic_object_field(objects[kind], "value", &field_cache)->number != (kind == 0 ? 10 : 20)) { abort(); }
 }
 for (size_t i = 0; i < 100000; i++) {
  size_t kind = (i + worker) % 2;
  adamic_object *object = objects[kind];
  if (adamic_object_field(object, "value", &field_cache)->number != (kind == 0 ? 10 : 20)) { abort(); }
  adamic_method method = NULL;
  adamic_closure *closure = adamic_object_callee(object, "invoke", &method_cache, &method);
  if ((closure == NULL) != (kind == 0)) { abort(); }
  adamic_value result = closure != NULL ? closure->code(closure, NULL) : method(object, NULL);
  if (result.number != (kind == 0 ? 11 : 22)) { abort(); }
 }
 return NULL;
}
int main(void) {
 objects[0] = adamic_object_new(&a_shape);
 objects[0]->slots[0].number = 10;
 objects[0]->slots[1].number = 300;
 objects[1] = adamic_object_new(&b_shape);
 objects[1]->slots[0].reference = adamic_closure_new(closure_code, 0);
 objects[1]->slots[1].number = 20;
 adamic_share(objects[0]); adamic_share(objects[1]);
 pthread_t readers[4];
 for (size_t i = 0; i < 4; i++) { if (pthread_create(&readers[i], NULL, read_fields, (void *)i) != 0) { abort(); } }
 for (size_t i = 0; i < 4; i++) { pthread_join(readers[i], NULL); }
 adamic_release(objects[0]); adamic_release(objects[1]);
 puts("cache pairs clean");
 return 0;
}
