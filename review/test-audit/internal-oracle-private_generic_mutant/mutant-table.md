| ID | Origin/main file:line | Change | Failed rows |
|---|---|---|---|
| M01 | internal/native/runtime/string_search_impl.h:115 | return (double)adamic_string_units(string); → return 0; | TestRuntimeLastIndexOfMatchesNode |
| M02 | internal/native/runtime/string_search_impl.h:123 | string->length - search->length + 1 → string->length - search->length | TestRuntimeLastIndexOfMatchesNode |
| M03 | internal/native/emit_expressions.go:288 | if identity > 0 { → if identity < 0 { | TestScannerNestedReferences |
| M04 | internal/javascript/javascript.go:51 | if (!values.has(code)) → if (values.has(code)) | TestScannerNestedReferences |
