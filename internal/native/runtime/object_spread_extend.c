// Layouts of added-field spreads. Only immutable layout metadata is interned;
// no object or field value is held by the cache. Shapes must outlive every copy.
#include "object_spread_extend.h"
#include <stdlib.h>
#include <string.h>

typedef struct adamic_spread_layout {
 adamic_shape shape;
 const char **names;
 bool *references;
 const adamic_shape **origins;
 struct adamic_spread_layout *next;
} adamic_spread_layout;

// A finite program has finitely many field names and static field representations.
// Canonicalize the resulting ordered layout, not the source/operation pair: repeated
// spreads of the same fields must not build an ever-growing chain of cached shapes.
static adamic_spread_layout *adamic_spread_layouts;

const adamic_shape *adamic_object_spread_field_shape(const adamic_shape *shape, const char *name) {
 for (const adamic_spread_layout *layout = adamic_spread_layouts; layout != NULL; layout = layout->next) {
  if (&layout->shape != shape) continue;
  for (size_t index = 0; index < shape->count; index++) {
   if (strcmp(shape->names[index], name) == 0) return layout->origins[index];
  }
 }
 return shape;
}

static void adamic_spread_layout_free(adamic_spread_layout *layout) {
 free(layout->names); free(layout->references); free(layout->origins); free(layout);
}

static adamic_spread_layout *adamic_spread_layout_intern(const adamic_shape *source, const adamic_shape *overrides) {
 if (source->count > SIZE_MAX - overrides->count) abort();
 size_t maximum = source->count + overrides->count;
 adamic_spread_layout *layout = calloc(1, sizeof *layout);
 if (layout == NULL) abort();
 layout->names = calloc(maximum, sizeof *layout->names);
 layout->references = calloc(maximum, sizeof *layout->references);
 layout->origins = calloc(maximum, sizeof *layout->origins);
 if (maximum != 0 && (layout->names == NULL || layout->references == NULL || layout->origins == NULL)) abort();
 layout->shape = (adamic_shape){source->count, layout->names, layout->references, NULL};
 for (size_t index = 0; index < source->count; index++) {
  layout->names[index] = source->names[index];
  layout->references[index] = source->references[index];
  layout->origins[index] = adamic_object_spread_field_shape(source, source->names[index]);
 }
 for (size_t own = 0; own < overrides->count; own++) {
  size_t index = 0;
  while (index < layout->shape.count && strcmp(layout->names[index], overrides->names[own]) != 0) index++;
  if (index == layout->shape.count) layout->shape.count++;
  layout->names[index] = overrides->names[own];
  layout->references[index] = overrides->references[own];
  layout->origins[index] = overrides;
 }
 for (adamic_spread_layout *held = adamic_spread_layouts; held != NULL; held = held->next) {
  if (held->shape.count != layout->shape.count) continue;
  bool same = true;
  for (size_t index = 0; index < layout->shape.count; index++) {
   if (strcmp(held->names[index], layout->names[index]) != 0 || held->references[index] != layout->references[index] || held->origins[index] != layout->origins[index]) { same = false; break; }
  }
  if (same) { adamic_spread_layout_free(layout); return held; }
 }
 layout->next = adamic_spread_layouts;
 adamic_spread_layouts = layout;
 return layout;
}

adamic_object *adamic_object_spread_extend(const adamic_object *snapshot, const adamic_shape *overrides) {
 const adamic_shape *shape = &adamic_spread_layout_intern(snapshot->shape, overrides)->shape;
 adamic_object *object = adamic_object_new(shape);
 for (size_t index = 0; index < snapshot->shape->count; index++) {
  bool replaced = false;
  for (size_t own = 0; own < overrides->count; own++) {
   if (strcmp(snapshot->shape->names[index], overrides->names[own]) == 0) { replaced = true; break; }
  }
  if (replaced) continue;
  object->slots[index] = snapshot->slots[index];
  if (shape->references[index]) adamic_retain(object->slots[index].reference);
 }
 return object;
}
