Seven original added lines are adapted in object.c. The two zero-cache initializers are listed separately by function. Main concurrency infrastructure is retained.

adamic_object_maybe_boolean:
```c
// Source own added line
if (adamic_object_field_types(object)[cache->index] == 2) { return (adamic_maybe_boolean){true, slot->boolean}; }
// Rebased added line
if (adamic_object_field_types(owner)[adamic_slot_index(owner, slot)] == 2) { return (adamic_maybe_boolean){true, slot->boolean}; }
```

adamic_contract_object cache:
```c
// Source own added line
adamic_slot_cache cache = {NULL, 0};
// Rebased added line
adamic_slot_cache cache = {0};
```

adamic_contract_object readiness:
```c
// Source own added line
if (!adamic_object_initialized(value)[cache.index] || adamic_accessor_find(value, contract->field_names[index]) != NULL) { return false; }
// Rebased added line
if (!adamic_object_initialized(value)[adamic_slot_index(value, slot)] || adamic_accessor_find(value, contract->field_names[index]) != NULL) { return false; }
```

adamic_contract_object runtime field:
```c
// Source own added line
if (!adamic_contract_runtime_field(expected, value, cache.index)) { return false; }
// Rebased added line
if (!adamic_contract_runtime_field(expected, value, adamic_slot_index(value, slot))) { return false; }
```

adamic_contract_object field contract:
```c
// Source own added line
const adamic_field_contract *actual = &value->shape->contracts[cache.index];
// Rebased added line
const adamic_field_contract *actual = &value->shape->contracts[adamic_slot_index(value, slot)];
```

adamic_object_check_contract cache:
```c
// Source own added line
adamic_slot_cache cache = {NULL, 0};
// Rebased added line
adamic_slot_cache cache = {0};
```

adamic_object_check_contract field contract:
```c
// Source own added line
const adamic_field_contract *contract = slot == NULL || object->shape->contracts == NULL ? NULL : &object->shape->contracts[cache.index];
// Rebased added line
const adamic_field_contract *contract = slot == NULL || object->shape->contracts == NULL ? NULL : &object->shape->contracts[adamic_slot_index(object, slot)];
```

One additional main runtime line is adapted in async_host_impl.h, host_settle:
```c
// Main
static const adamic_shape shape = {2, names, references, NULL};
// Rebased
static const adamic_shape shape = {2, names, references, NULL, NULL};
```

All other source own runtime +/- lines are identical. Full signed comparison: runtime-delta-differences.patch.
