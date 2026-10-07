// process.c: the prelude's numeric exit status and read-only external observations.
#define _POSIX_C_SOURCE 200809L
#include "adamic.h"
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

static adamic_maybe_number exit_code;

adamic_maybe_number adamic_process_exit_code(void) { return exit_code; }

int adamic_process_status(void) { return exit_code.present ? (int)exit_code.number : 0; }

// Node validates before changing its status, and keeps ToInt32(code), not the original number.
static bool valid_code(adamic_maybe_number code) {
	if (!code.present || (isfinite(code.number) && trunc(code.number) == code.number)) { return true; }
	static adamic_string prefix = ADAMIC_STRING("The value of \"code\" is out of range. It must be an integer. Received ");
	static adamic_string range_name = ADAMIC_STRING("RangeError");
	adamic_string *number = adamic_string_from_number(code.number);
	adamic_string *message = adamic_string_concat(2, (adamic_string *const[]){&prefix, number});
	static adamic_string range_code = ADAMIC_STRING("ERR_OUT_OF_RANGE");
	adamic_node_error(&range_name, message, &range_code);
	adamic_release(number);
	adamic_release(message);
	return false;
}

void adamic_process_set_exit_code(adamic_maybe_number code) {
	if (!valid_code(code)) { return; }
	exit_code = code;
	if (code.present) { exit_code.number = adamic_bitwise_or(code.number, 0); }
}

void adamic_process_exit(adamic_maybe_number code) {
	if (!valid_code(code)) { return; }
	// Lowering supplies exitCode for an omitted argument. Node 24's explicit undefined means 0.
	adamic_process_set_exit_code(code);
	adamic_process_exit_now(adamic_process_status());
}

adamic_maybe_boolean adamic_process_is_tty(enum adamic_stream stream) {
	bool terminal = isatty(stream) != 0;
	return (adamic_maybe_boolean){.present = terminal, .boolean = terminal};
}

adamic_string *adamic_process_environment(const adamic_string *name) {
	char *bytes = malloc(name->length + 1);
	if (bytes == NULL) { static const char message[] = "out of memory"; adamic_panic(message, sizeof message - 1); }
	memcpy(bytes, name->bytes, name->length);
	// Node encodes lone surrogates as U+FFFD, and passes names through a NUL-terminated API.
	for (size_t at = 0; at + 3 <= name->length; at++) {
		if ((unsigned char)bytes[at] == 0xed && (unsigned char)bytes[at + 1] >= 0xa0) { memcpy(bytes + at, "\xef\xbf\xbd", 3); at += 2; }
	}
	bytes[name->length] = 0;
	const char *value = getenv(bytes);
	free(bytes);
	return value == NULL ? NULL : adamic_decode_utf8((const unsigned char *)value, strlen(value));
}
