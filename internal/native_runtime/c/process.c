// process.c: checked Node process bindings over host_runtime.h.
#define _POSIX_C_SOURCE 200809L
#include "adamic.h"
#include "host_runtime.h"
#include <stdlib.h>

adamic_maybe_number adamic_process_exit_code(void) {
    int32_t value;
    bool present = adamic_host_exit_code(&value);
    return (adamic_maybe_number){.present = present, .number = value};
}

// Node groups an integer's formatted characters from the right, including exponent notation (number.c).
static adamic_string *exit_code_text(double value, bool grouped) {
    char text[ADAMIC_NUMBER_RECEIVED_MAX];
    size_t length = adamic_number_received(value, grouped, text);
    return adamic_decode_utf8((const unsigned char *)text, length);
}

// Node validates before changing its status, and keeps ToInt32(code), not the original number.
static bool valid_code(adamic_maybe_number code) {
    if (!code.present) { return true; }
    bool integer = isfinite(code.number) && trunc(code.number) == code.number;
    bool range = integer && fabs(code.number) > 9007199254740991.0;
    if (integer && !range) { return true; }
	static adamic_string prefix = ADAMIC_STRING("The value of \"code\" is out of range. It must be an integer. Received ");
	static adamic_string range_name = ADAMIC_STRING("RangeError");
    static adamic_string range_prefix = ADAMIC_STRING("The value of \"code\" is out of range. It must be >= -9007199254740991 && <= 9007199254740991. Received ");
    adamic_string *number = exit_code_text(code.number, range);
	adamic_string *message = adamic_string_concat(2, (adamic_string *const[]){range ? &range_prefix : &prefix, number});
	static adamic_string range_code = ADAMIC_STRING("ERR_OUT_OF_RANGE");
	static const char *const names[] = {"name", "message", "code"};
	static const bool references[] = {true, true, true};
	static const adamic_field_kind kinds[] = {adamic_field_reference, adamic_field_reference, adamic_field_reference};
	// Node builds this RangeError with makeNodeErrorWithCode, so its toString brackets the code (exceptions.c).
	static const adamic_shape shape = {3, names, references, &adamic_coded_error_methods, kinds};
	static const char *const shape_json_keys[] = {"code"};
	static adamic_error_json shape_json = {&shape, shape_json_keys, 1, NULL, false};
	adamic_register_error_json(&shape_json);
	adamic_thrown = adamic_object_new(&shape);
	adamic_thrown->slots[0].reference = adamic_retain(&range_name);
	adamic_thrown->slots[1].reference = adamic_retain(message);
	adamic_thrown->slots[2].reference = adamic_retain(&range_code);
	adamic_object_field_types(adamic_thrown)[2] = adamic_ir_String;
	adamic_release(number);
	adamic_release(message);
	return false;
}

void adamic_process_set_exit_code(adamic_maybe_number code) {
	if (!valid_code(code)) { return; }
    int32_t value = code.present ? (int32_t)adamic_bitwise_or(code.number, 0) : 0;
    adamic_host_set_exit_code(code.present, value);
}

void adamic_process_exit(adamic_maybe_number code) {
	if (!valid_code(code)) { return; }
	// Lowering supplies exitCode for an omitted argument. Node 24's explicit undefined means 0.
	adamic_process_set_exit_code(code);
	adamic_process_exit_now(adamic_host_exit_status());
}


adamic_string *adamic_process_environment(const adamic_string *name) {
    return adamic_host_environment(name);
}
