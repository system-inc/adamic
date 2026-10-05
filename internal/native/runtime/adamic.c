// adamic.c: the Adamic runtime.

#include "adamic.h"

#include <errno.h>
#include <stdlib.h>
#include <unistd.h>

// write_all writes every byte, through partial writes and interrupted calls. It reports failure
// rather than retrying forever: a closed stdout is not something a program can wait out.
static int write_all(int descriptor, const char *bytes, size_t length) {
	while (length > 0) {
		ssize_t written = write(descriptor, bytes, length);
		if (written < 0) {
			if (errno == EINTR) {
				continue;
			}
			return -1;
		}
		bytes += written;
		length -= (size_t)written;
	}
	return 0;
}

// write_text writes a string's bytes, each lone surrogate (WTF-8's ED A0 80 to ED BF BF) as U+FFFD,
// which is what Node writes for one.
static int write_text(int descriptor, const char *bytes, size_t length) {
	size_t start = 0;
	for (size_t at = 0; at + 3 <= length; at++) {
		if ((unsigned char)bytes[at] == 0xed && (unsigned char)bytes[at + 1] >= 0xa0) {
			if (write_all(descriptor, bytes + start, at - start) != 0 || write_all(descriptor, "\xef\xbf\xbd", 3) != 0) {
				return -1;
			}
			at += 2;
			start = at + 1;
		}
	}
	return write_all(descriptor, bytes + start, length - start);
}

void adamic_write_line(enum adamic_stream stream, const adamic_string *string) {
	if (write_text((int)stream, string->bytes, string->length) != 0 || write_all((int)stream, "\n", 1) != 0) {
		static const char message[] = "writing output failed";
		adamic_panic(message, sizeof message - 1);
	}
}

_Noreturn void adamic_panic(const char *message, size_t length) {
	static const char prefix[] = "adamic: panic: ";
	// Best effort: stderr may be what failed.
	(void)write_all(adamic_stderr, prefix, sizeof prefix - 1);
	(void)write_all(adamic_stderr, message, length);
	(void)write_all(adamic_stderr, "\n", 1);
	_exit(70);
}

_Noreturn void adamic_unreachable(void) {
	static const char message[] = "compiler bug: a function ended without returning";
	adamic_panic(message, sizeof message - 1);
}
