// Functional replacement for the intrinsic RegExp execution path,
// ECMA-262 22.2.6.11.
#include "adamic.h"
#include <string.h>
#include <stdlib.h>

#ifdef ADAMIC_REGEXP_REPLACE_CALLBACK

// Each argument has its actual kind before conversion to the proven closure
// representation.
static adamic_value regex_replacement_argument(adamic_value value, unsigned actual,
											   unsigned expected) {
	unsigned kind = expected & 15;
	bool union_argument = (expected & 64) != 0;
	bool accepted = union_argument ? ((expected & (1u << actual)) != 0)
								   : (actual == kind || (actual == 0 && (expected & 16) != 0));
	if (!accepted) {
		const char message[] =
			"RegExp replacement callback argument does not fit its declared type";
		adamic_panic(message, sizeof message - 1);
	}
	if (union_argument) {
		if (actual == 2)
			return (adamic_value){.reference = adamic_box_number(value.number)};
		return (adamic_value){.reference = adamic_retain(value.reference)};
	}
	if (actual == 2)
		return value;
	return (adamic_value){.reference = adamic_retain(value.reference)};
}

adamic_string *adamic_regex_replace_callback(adamic_string *input, adamic_object *regex,
											 adamic_closure *callback, bool require_global,
											 const unsigned char *parameters,
											 size_t parameter_count, unsigned rest_kind,
											 unsigned return_kind,
											 const adamic_regex_replacement_group *group_checks,
											 size_t group_check_count, size_t rest_slot) {
	bool global = regex->slots[4].boolean;
	if (require_global && !global) {
		const char message[] = "TypeError: String.prototype.replaceAll called with "
							   "a non-global RegExp argument";
		adamic_panic(message, sizeof message - 1);
	}
	if (global)
		regex->slots[1].number = 0;
	adamic_array *matches = adamic_array_new(0, true);
	for (;;) {
		adamic_array *match = adamic_regex_exec(regex, input);
		if (match == NULL)
			break;
		adamic_array_push(matches, (adamic_value){.reference = match});
		if (!global)
			break;
		if (adamic_string_length(match->elements[0].reference) == 0) {
			double at = regex->slots[1].number;
			double first = adamic_string_char_code_at(input, at);
			double second = adamic_string_char_code_at(input, at + 1);
			bool pair = (regex->slots[7].boolean || regex->slots[10].boolean) && first >= 0xd800 &&
						first <= 0xdbff && second >= 0xdc00 && second <= 0xdfff;
			regex->slots[1].number = at + (pair ? 2 : 1);
		}
	}
	adamic_array *pieces = adamic_array_new(0, true);
	size_t previous = 0;
	size_t packed_count = rest_kind == 0 ? parameter_count : rest_slot + 1;
	adamic_value *packed = calloc(packed_count == 0 ? 1 : packed_count, sizeof *packed);
	if (packed == NULL) {
		const char message[] = "out of memory";
		adamic_panic(message, sizeof message - 1);
	}
	for (size_t k = 0; k < matches->length; k++) {
		adamic_array *match = matches->elements[k].reference;
		size_t offset = (size_t)match->properties->slots[0].number;
		adamic_object *groups = match->properties->slots[2].reference;
		size_t argument_count = match->length + 2 + (groups != NULL);
		adamic_array *rest = rest_kind == 0 ? NULL : adamic_array_new(0, rest_kind != 2);
		for (size_t j = 0; j < argument_count || j < parameter_count; j++) {
			adamic_value value = {.reference = NULL};
			unsigned actual = 0;
			if (j < match->length) {
				value = match->elements[j];
				actual = value.reference == NULL ? 0 : 1;
			} else if (j == match->length) {
				value.number = (double)offset;
				actual = 2;
			} else if (j == match->length + 1) {
				value.reference = input;
				actual = 1;
			} else if (j == match->length + 2 && groups != NULL) {
				value.reference = groups;
				actual = 3;
			}
			if (actual == 3) {
				for (size_t check = 0; check < group_check_count; check++) {
					const adamic_regex_replacement_group *rule = &group_checks[check];
					if (rule->rest ? j < rule->argument : j != rule->argument)
						continue;
					void *field = adamic_regex_group_lookup(value.reference, rule->name, true);
					if (field == NULL && !rule->optional) {
						const char message[] =
							"RegExp replacement callback argument does not fit its declared type";
						adamic_panic(message, sizeof message - 1);
					}
				}
			}
			if (j < parameter_count)
				packed[j] = regex_replacement_argument(value, actual, parameters[j]);
			else if (rest != NULL && j < argument_count)
				adamic_array_push(rest, regex_replacement_argument(value, actual, rest_kind));
		}
		if (rest != NULL)
			packed[rest_slot].reference = rest;
		adamic_value returned = adamic_closure_call(callback, packed, argument_count);
		for (size_t j = 0; j < parameter_count; j++)
			if (parameters[j] != 2)
				adamic_release(packed[j].reference);
		adamic_release(rest);
		if (adamic_exception_pending)
			break;
		adamic_string *replacement = NULL;
		switch (return_kind) {
		case 0:
			replacement = &adamic_string_undefined;
			break;
		case 1:
			replacement =
				returned.reference == NULL ? &adamic_string_undefined : returned.reference;
			break;
		case 3:
			replacement = adamic_string_allocate(4);
 memcpy((char *)replacement->bytes, "null", 4);
			break;
		case 2:
			replacement = adamic_string_from_number(returned.number);
			break;
		case 5:
			replacement = returned.boolean ? &adamic_string_true : &adamic_string_false;
			break;
		case 4:
			replacement = adamic_union_to_string(returned.reference);
			adamic_release(returned.reference);
			break;
		}
		adamic_array_push(pieces,
						  (adamic_value){.reference = adamic_string_slice(input, (double)previous,
																		  (double)offset, true)});
		adamic_array_push(pieces, (adamic_value){.reference = replacement});
		previous = offset + (size_t)adamic_string_length(match->elements[0].reference);
	}
	free(packed);
	adamic_release(matches);
	adamic_string *result = NULL;
	if (!adamic_exception_pending) {
		adamic_array_push(pieces, (adamic_value){.reference = adamic_string_slice(
													 input, (double)previous,
													 adamic_string_length(input), true)});
		result = adamic_array_join(pieces, &adamic_string_empty, adamic_join_strings);
	}
	adamic_release(pieces);
	return result;
}

#endif
