#ifndef ADAMIC_VIEW_REPRESENTATIONS_H
#define ADAMIC_VIEW_REPRESENTATIONS_H
/* Must match internal/ir/ir.go. Semantic null/undefined tags do not change
   the counted-reference slot layout. Typed-array storage remains a reference. */
enum adamic_view_representation {
 adamic_rep_number=1, adamic_rep_boolean=2, adamic_rep_string=3,
 adamic_rep_object=4, adamic_rep_array=5, adamic_rep_map=6,
 adamic_rep_maybe_number=7, adamic_rep_closure=8, adamic_rep_maybe_boolean=9,
 adamic_rep_union=10, adamic_rep_weak=11, adamic_rep_null=12,
 adamic_rep_undefined=13, adamic_rep_record=14,
 adamic_rep_uint8_array=15, adamic_rep_int32_array=16, adamic_rep_float64_array=17
};
#endif
