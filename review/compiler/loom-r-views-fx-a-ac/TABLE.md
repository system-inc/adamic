# Per-file a-check status

All 54 files were compiled in place by this branch. Checked means `adamic c` exited zero; NotYet would be reported separately. Compiler panics are failed outcomes, not refusals or NotYet. Program bodies have identical SHA-256 values before and after; full diagnostics and hashes are in statuses.json.

| File | Status | Header or failure reason |
|---|---|---|
| internal/lower/testdata/loom_views_a/callable.a | refused | // a-check: refused checked view read of field count with unsupported representation conversion contract |
| internal/lower/testdata/loom_views_a/key.a | refused | // a-check: refused view member read by element access is not yet checked |
| internal/lower/testdata/loom_views_a/native_p21.a | refused | // a-check: refused checked view read of field count with unsupported representation conversion contract |
| internal/lower/testdata/loom_views_a/native_p37.a | refused | // a-check: refused checked view read of field count with unsupported representation conversion contract |
| internal/lower/testdata/loom_views_a/union_key.a | refused | // a-check: refused view member read by element access is not yet checked |
| internal/lower/testdata/loom_views_a/unsupported.a | refused | // a-check: refused checked view read of field label with unsupported representation conversion contract |
| internal/lower/testdata/loom_views_a/views_p01.a | refused | // a-check: refused view member read by element access is not yet checked |
| internal/lower/testdata/loom_views_a/views_p05.a | refused | // a-check: refused view member read by element access is not yet checked |
| internal/lower/testdata/loom_views_a/views_p06.a | refused | // a-check: refused view member read by element access is not yet checked |
| internal/lower/testdata/loom_views_a/views_p19.a | refused | // a-check: refused view member read by element access is not yet checked |
| internal/lower/testdata/loom_views_a/views_p72.a | refused | // a-check: refused view member read by element access is not yet checked |
| review/compiler/loom-r-views-fx-a/controls/key_fix.a | checked |  |
| review/compiler/loom-r-views-fx-a/controls/unsupported_unread.a | checked |  |
| review/compiler/loom-views-fx-a/sources/00584bccadf7/p19-native-call-gap.a | refused | // a-check: refused view member read by element access is not yet checked |
| review/compiler/loom-views-fx-a/sources/0ade804a7ad8/misfit_boolean.a | failed | panic: compiler bug: narrowed scalar union view requires a checked conversion |
| review/compiler/loom-views-fx-a/sources/16353fe155c5/p70_copy.a | checked |  |
| review/compiler/loom-views-fx-a/sources/166413c08ae4/p69.a | checked |  |
| review/compiler/loom-views-fx-a/sources/19560c042dd4/depth-2000.a | checked |  |
| review/compiler/loom-views-fx-a/sources/1dd37b5d0d65/values.a | checked |  |
| review/compiler/loom-views-fx-a/sources/23b93dc07994/p135_tagged_unread.a | checked |  |
| review/compiler/loom-views-fx-a/sources/27fe1106b197/wide.a | refused | // a-check: refused checked view read of field value with unsupported representation conversion contract |
| review/compiler/loom-views-fx-a/sources/2ca18403dfcd/snapshot-maybe-number.a | checked |  |
| review/compiler/loom-views-fx-a/sources/2db1d7d06600/depth-200.a | checked |  |
| review/compiler/loom-views-fx-a/sources/2dcbcfe5ab49/spread.a | checked |  |
| review/compiler/loom-views-fx-a/sources/39035180ca1c/oct9_views_p64.a | refused | // a-check: refused view member read by element access is not yet checked |
| review/compiler/loom-views-fx-a/sources/3a666aeb3afd/control.a | checked |  |
| review/compiler/loom-views-fx-a/sources/4123e0eab5d3/p135_object_result.a | checked |  |
| review/compiler/loom-views-fx-a/sources/4311ed865723/callable_union_result_both.a | checked |  |
| review/compiler/loom-views-fx-a/sources/4959ae25761c/callable_union_result_number.a | checked |  |
| review/compiler/loom-views-fx-a/sources/50f491e5d7c1/entries.a | checked |  |
| review/compiler/loom-views-fx-a/sources/51e01ecde063/oct9_views_p17.a | refused | // a-check: refused view member read by element access is not yet checked |
| review/compiler/loom-views-fx-a/sources/55aa222e57ec/p19_read.a | refused | // a-check: refused view member read by element access is not yet checked |
| review/compiler/loom-views-fx-a/sources/5907cb79c2f5/callable_union_result_zero.a | checked |  |
| review/compiler/loom-views-fx-a/sources/591c2a2453e4/p49.a | checked |  |
| review/compiler/loom-views-fx-a/sources/6c1f0555d211/in.a | checked |  |
| review/compiler/loom-views-fx-a/sources/6f7c7da4c8e2/callable_union_result_string.a | checked |  |
| review/compiler/loom-views-fx-a/sources/702776fe1a62/scalar-tuple.a | checked |  |
| review/compiler/loom-views-fx-a/sources/7417f0789932/p54_s.a | checked |  |
| review/compiler/loom-views-fx-a/sources/78a920f10d1e/depth-65.a | checked |  |
| review/compiler/loom-views-fx-a/sources/792dd026d5e3/destructured_union.a | checked |  |
| review/compiler/loom-views-fx-a/sources/93acc4186926/p37.a | refused | // a-check: refused checked view read of field label with unsupported representation conversion contract |
| review/compiler/loom-views-fx-a/sources/9889d0b80ce2/p62_adapted.a | checked |  |
| review/compiler/loom-views-fx-a/sources/9e6bff2a3644/keys_alias.a | checked |  |
| review/compiler/loom-views-fx-a/sources/9e84065b734b/keys.a | checked |  |
| review/compiler/loom-views-fx-a/sources/a705d6d0d5df/misfit_number.a | failed | panic: compiler bug: narrowed scalar union view requires a checked conversion |
| review/compiler/loom-views-fx-a/sources/ab0bf403ae83/p22_adapted.a | checked |  |
| review/compiler/loom-views-fx-a/sources/b057160c0f66/p51.a | failed | panic: compiler bug: incomplete checked field contract |
| review/compiler/loom-views-fx-a/sources/b1a2a62ba1f4/bad_callable.a | refused | // a-check: refused checked view read of field value with unsupported representation conversion contract |
| review/compiler/loom-views-fx-a/sources/c76e229a6927/p07.a | refused | // a-check: refused view member read by element access is not yet checked |
| review/compiler/loom-views-fx-a/sources/cfb0fcf84cfa/destructured.a | checked |  |
| review/compiler/loom-views-fx-a/sources/d44ab71300a7/snapshot-maybe-boolean.a | checked |  |
| review/compiler/loom-views-fx-a/sources/d8a921ff2453/oct9_views_p48.a | refused | // a-check: refused checked view read of field values with unsupported collection contract |
| review/compiler/loom-views-fx-a/sources/db935ac26f25/literal_key.a | refused | // a-check: refused view member read by element access is not yet checked |
| review/compiler/loom-views-fx-a/sources/fb59d3aecf4e/p59_adapted.a | checked |  |
