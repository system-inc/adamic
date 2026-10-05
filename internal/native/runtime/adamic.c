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

void adamic_write_line(enum adamic_stream stream, const char *bytes, size_t length) {
	if (write_all((int)stream, bytes, length) != 0 || write_all((int)stream, "\n", 1) != 0) {
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
