# Runtime clearance

These are the exact signed runtime lines in this re-cut that are beyond e47a602d81's own added/removed lines. Minus lines are current main lines extended by the integration; plus lines are final code. Common source-patch lines are excluded, with repeated lines counted by occurrence. The complete ordered patches are json_stringify.c-resolution.patch and json_stringify.h-resolution.patch.

Main's typed providers, toJSON, throw handling, cycle ancestors, nullable metadata and borrowed object slot views are retained. The record enum is appended after toJSON, preserving main's existing enum values. The record path prepares each field exactly once before testing omission or writing its key, releases transferred reference results, releases its key snapshot on throws, and participates in main's cycle guard. No main behavior is reverted. Node, balanced counts, ASan/UBSan/LeakSanitizer and four independent runtime mutants cover the new interactions.

### internal/native/runtime/json_stringify.c

```diff
-    return schema->kind == adamic_json_string || schema->kind == adamic_json_array || schema->kind == adamic_json_object || schema->kind == adamic_json_toJSON || schema->kind == adamic_json_union || schema->kind == adamic_json_maybe_boolean || schema->kind == adamic_json_map || schema->kind == adamic_json_function;
+    return schema->kind == adamic_json_string || schema->kind == adamic_json_array || schema->kind == adamic_json_object || schema->kind == adamic_json_toJSON || schema->kind == adamic_json_union || schema->kind == adamic_json_maybe_boolean || schema->kind == adamic_json_map || schema->kind == adamic_json_function || schema->kind == adamic_json_record;
-    bool container = s.kind == adamic_json_array || s.kind == adamic_json_tuple || s.kind == adamic_json_object || s.kind == adamic_json_toJSON;
+    bool container = s.kind == adamic_json_array || s.kind == adamic_json_tuple || s.kind == adamic_json_object || s.kind == adamic_json_toJSON || s.kind == adamic_json_record;
+		const adamic_record *record = s.value.reference;
+			bool owned;
+			adamic_json_result result = prepare(w, *slot, schema->element, key, &owned);
+			if (adamic_thrown != NULL) {
+				if (owned && reference_result(result.schema)) { adamic_release(result.value.reference); }
+				adamic_release(ordered); return false;
+			}
+			json_scalar item = scalar(result.value, result.schema, w->runtime);
+			if (item.kind != adamic_json_undefined && item.kind != adamic_json_function) {
+				if (written++ != 0) { ascii(w, ","); }
+				indent(w, depth + 1); quote(w, key); ascii(w, w->pretty ? ": " : ":");
+				(void)write_prepared(w, result.value, result.schema, depth + 1);
+			}
+			if (owned && reference_result(result.schema)) { adamic_release(result.value.reference); }
+			if (adamic_thrown != NULL) { adamic_release(ordered); return false; }
```

### internal/native/runtime/json_stringify.h

```diff
-	adamic_json_tuple, adamic_json_object, adamic_json_toJSON
+	adamic_json_tuple, adamic_json_object, adamic_json_toJSON, adamic_json_record
```
