# Proven writable-view counts

Compiler head: `781766edccfd21ed4f2e9f9106dd488d28751e96`.

All 98 original receiver classifications remain unchanged on the pinned source. This is stock TypeScript audit evidence. Individual original-site Adamic compile/check counts are unmeasured.

Two compiling witnesses represent 86 sites; both agree with source Node in sanitized native and emitted JavaScript, with zero writable-view checks. Eight witnesses representing 12 sites remain blocked.

| Witness | Sites represented | In-place result | Writable-view checks observed |
|---|---:|---|---:|
| 01_truthiness.a | 82 | Compiles | 0 |
| 02_control.a | 4 | Compiles | 0 |
| 03_kind_reader.a | 3 | NotYet | unmeasured |
| 04_range_reader.a | 2 | Refused | unmeasured |
| 05_diagnostic_related.a | 2 | Refused | unmeasured |
| 06_modifiers_initialize.a | 1 | Refused | unmeasured |
| 07_assert_clause.a | 1 | Refused | unmeasured |
| 08_property_initializer.a | 1 | Refused | unmeasured |
| 09_static_modifiers.a | 1 | Refused | unmeasured |
| 10_module_parent.a | 1 | Refused | unmeasured |

Finished native witnesses pass ASan, UBSan and LeakSanitizer. No allocation-count measurement was made for these witnesses. Required checks remain zero at each of the 98 original sites.
