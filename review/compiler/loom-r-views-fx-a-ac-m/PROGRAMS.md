All paths are under internal/oracle/testdata/review. Each program moved from agree/ to refused/. Its body is unchanged below its new first-line a-check header.

| Program | Exact refusal reason | Fix |
|---|---|---|
| fxspptb_oct9_views_p07_element_access_union_wrong.a | view member read by element access is not yet checked | use view.value |
| fxspptb_oct9_native_p17_callable_call_first_half.a | checked view read of field value with unsupported representation conversion contract | prove or implement the representation conversion contract before reading this field |
| fxspptb_oct9_native_p18_cast_kind_no_field_types.a | checked view read of field value with unsupported representation conversion contract | prove or implement the representation conversion contract before reading this field |
| fxspptb_oct9_native_p20_callable_member_call_good.a | checked view read of field count with unsupported representation conversion contract | prove or implement the representation conversion contract before reading this field |
| fxspptb_oct9_native_p21_callable_view_raw_union_slot.a | checked view read of field count with unsupported representation conversion contract | prove or implement the representation conversion contract before reading this field |
| fxspptb_oct9_native_p26_callable_or_undefined.a | checked view read of field value with unsupported representation conversion contract | prove or implement the representation conversion contract before reading this field |
| fxspptb_oct9_native_p33_map_in_union.a | checked view read of field value with unsupported representation conversion contract | prove or implement the representation conversion contract before reading this field |
| fxspptb_oct9_native_p34_typed_array_in_union.a | checked view read of field value with unsupported representation conversion contract | prove or implement the representation conversion contract before reading this field |
| fxspptb_oct9_native_p35_class_in_union.a | checked view read of field value with unsupported representation conversion contract | prove or implement the representation conversion contract before reading this field |
| fxspptb_oct9_native_p37_unsupported_view_raw_zero.a | checked view read of field count with unsupported representation conversion contract | prove or implement the representation conversion contract before reading this field |
| fxspptb_oct9_views_p01_element_access_wrong.a | view member read by element access is not yet checked | use view.text |
| fxspptb_oct9_views_p05_element_access_number_wrong.a | view member read by element access is not yet checked | use view.count |
| fxspptb_oct9_views_p06_element_access_boolean_wrong.a | view member read by element access is not yet checked | use view.count |
| fxspptb_oct9_views_p19_callable_union_call_element.a | view member read by element access is not yet checked | use view.value |
| fxspptb_oct9_views_p72_mixed_element_access_wrong.a | view member read by element access is not yet checked | use view.value |

Additional header only, not moved: internal/oracle/testdata/review/refused/fxspptb_oct9_views_p64_array_arm_element.a now starts with `// a-check: refused view member read by element access is not yet checked`. Its former main outcome was NotYet and its body is unchanged.
