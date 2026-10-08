package oracle

func init() {
	for _, name := range []string{
		"assignment_non_null_field_string_accessor.a",
		"assignment_non_null_field_number_accessor.a",
		"assignment_non_null_uint8array.a", "assignment_non_null_int32array.a", "assignment_non_null_float64array.a", "assignment_non_null_field_string.a",
		"assignment_non_null_array_number.a", "assignment_non_null_array_string.a", "assignment_non_null_field_number.a",
		"assignment_object.a",
		"assignment_value_undefined.a",
		"assignment_value_accessor.a",
		"assignment_value_weak.a",
		"assignment_value_typed_index_float64array.a",
		"assignment_value_typed_index_int32array.a",
		"assignment_value_typed_index_uint8array.a",
		"assignment_value_order.a",
		"assignment_value_local_number.a",
		"assignment_value_local_boolean.a",
		"assignment_value_local_maybe_number.a",
		"assignment_value_local_maybe_boolean.a",
		"assignment_value_local_string.a",
		"assignment_value_local_object.a",
		"assignment_value_local_array.a",
		"assignment_value_local_map.a",
		"assignment_value_local_closure.a",
		"assignment_value_local_union.a",
		"assignment_value_local_uint8.a",
		"assignment_value_local_int32.a",
		"assignment_value_local_float64.a",
		"assignment_value_field_number.a",
		"assignment_value_field_boolean.a",
		"assignment_value_field_maybe_number.a",
		"assignment_value_field_maybe_boolean.a",
		"assignment_value_field_string.a",
		"assignment_value_field_object.a",
		"assignment_value_field_array.a",
		"assignment_value_field_map.a",
		"assignment_value_field_closure.a",
		"assignment_value_field_uint8.a",
		"assignment_value_field_int32.a",
		"assignment_value_field_float64.a",
		"assignment_value_element_number.a",
		"assignment_value_element_boolean.a",
		"assignment_value_element_maybe_number.a",
		"assignment_value_element_string.a",
		"assignment_value_element_object.a",
		"assignment_value_element_array.a",
		"assignment_value_element_map.a",
		"assignment_value_element_closure.a",
		"assignment_value_element_uint8.a",
		"assignment_value_element_int32.a",
		"assignment_value_element_float64.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/" + name, true, false})
	}
}
