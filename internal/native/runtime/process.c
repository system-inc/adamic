// process.c: the prelude's numeric exit status and read-only external observations.
#define _POSIX_C_SOURCE 200809L
#include "adamic.h"
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

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
