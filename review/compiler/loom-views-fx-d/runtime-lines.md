Runtime C changes against bd90a433. Every added/replacement line follows; there are no other runtime C changes.

`internal/native/runtime/object.c`, lines 233-237:

```c
	if ((actual == adamic_rep_object || actual == adamic_rep_union || actual == adamic_rep_weak) && slot->reference != NULL && ((const adamic_heap *)slot->reference)->kind == adamic_kind_object) {
		const adamic_object *value = slot->reference;
		if (value->tuple) { found = "array"; }
		else if (value->class != NULL && value->class->is_static) { found = "function"; }
	}
```

`internal/native/runtime/view_unions_mixed.c`, lines 69-73:

```c
    if (value->kind == adamic_view_union_object && value->payload.reference != NULL) {
        const adamic_object *object = value->payload.reference;
        if (object->tuple) { found = "array"; }
        else if (object->class != NULL && object->class->is_static) { found = "function"; }
    }
```

`internal/native/runtime/view_unions_untagged.c`, lines 94-94:

```c
    ADAMIC_CHECK_STACK();
```

The untagged line replaces `if (depth > 128) { return false; }`. Object and mixed-union lines are insertions only. Runtime owners must clear these three files before submission.
