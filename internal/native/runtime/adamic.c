// adamic.c: the Adamic runtime.

// SIGPIPE and isatty are POSIX's, which strict C11 doesn't show without asking.
#define _POSIX_C_SOURCE 200809L

#include "adamic.h"
#include "count.h"

#include <errno.h>
#include <signal.h>
#include <stdlib.h>
#include <string.h>
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

// Output is written as Node writes it, so nothing a program prints can be told apart by where it
// lands or what else lands there. Node writes each console.log at once, to a file, a terminal or (on
// Linux) a pipe. Stdout here is held in a buffer instead, since a call to write per line costs more
// than the line, and the buffer goes out at every point the difference could be seen: before anything
// is written to stderr (which may be the same file), before a panic's message, at exit, when it's
// full, and after every line when stdout is a terminal, where a person is watching. What a reader of
// a pipe sees is the same bytes, in larger pieces.
//
// When a write fails (a pipe whose reader is gone, say), the program goes on as it does on Node,
// where the stream's error arrives only once the program's own code has run: what it writes there
// after is dropped, stderr is still written, and the exit is 70 rather than 0, as the oracle's
// runtime makes it.
static char output[1 << 16];
static size_t output_used;

// output_mode is 0 until the first line, then 1 for a buffer, or 2 for a line at a time (a terminal).
static int output_mode;

// broken is the stream a write to has failed, from then on dropped; 0 while none has.
static bool broken[3];

static void output_failed(enum adamic_stream stream) {
	broken[stream] = true;
}

static void flush(void) {
	if (output_used > 0) {
		size_t used = output_used;
		output_used = 0;
		if (!broken[adamic_stdout] && write_all(adamic_stdout, output, used) != 0) {
			output_failed(adamic_stdout);
		}
	}
}

// finish is the end of the program's output, at exit.
static void finish(void) {
	flush();
	if (broken[adamic_stdout] || broken[adamic_stderr]) {
		_exit(70);
	}
}

// buffer adds bytes to stdout's buffer, writing it out as it fills; more than the buffer holds goes
// straight out after it.
static void buffer(const char *bytes, size_t length) {
	if (length > sizeof output - output_used) {
		flush();
		if (length > sizeof output) {
			if (!broken[adamic_stdout] && write_all(adamic_stdout, bytes, length) != 0) {
				output_failed(adamic_stdout);
			}
			return;
		}
	}
	memcpy(output + output_used, bytes, length);
	output_used += length;
}

// put writes bytes to stdout's buffer, or straight to stderr.
static void put(enum adamic_stream stream, const char *bytes, size_t length) {
	if (stream == adamic_stdout) {
		buffer(bytes, length);
	} else if (!broken[stream] && write_all(stream, bytes, length) != 0) {
		output_failed(stream);
	}
}

// write_text writes a string's bytes, each lone surrogate (WTF-8's ED A0 80 to ED BF BF) as U+FFFD,
// which is what Node writes for one.
static void write_text(enum adamic_stream stream, const char *bytes, size_t length) {
	size_t start = 0;
	for (size_t at = 0; at + 3 <= length; at++) {
		if ((unsigned char)bytes[at] == 0xed && (unsigned char)bytes[at + 1] >= 0xa0) {
			put(stream, bytes + start, at - start);
			put(stream, "\xef\xbf\xbd", 3);
			at += 2;
			start = at + 1;
		}
	}
	put(stream, bytes + start, length - start);
}

void adamic_start(int count, char **values) {
	adamic_arguments_save(count, values);
	// Node ignores SIGPIPE, and a write to a pipe nobody reads is a failed write, not a killed process.
	signal(SIGPIPE, SIG_IGN);
}

void adamic_write_line(enum adamic_stream stream, const adamic_string *string) {
	if (output_mode == 0) {
		output_mode = isatty(adamic_stdout) ? 2 : 1;
		atexit(finish);
	}
	if (stream == adamic_stderr) {
		// Whatever stdout holds was written first, and stderr may be the same file.
		flush();
	}
	write_text(stream, string->bytes, string->length);
	put(stream, "\n", 1);
	if (stream == adamic_stdout && output_mode == 2) {
		flush();
	}
}

_Noreturn void adamic_panic(const char *message, size_t length) {
	static const char prefix[] = "adamic: panic: ";
	// Everything the program printed comes first, as on Node.
	flush();
	// Best effort: stderr may be what failed.
	(void)write_all(adamic_stderr, prefix, sizeof prefix - 1);
	(void)write_all(adamic_stderr, message, length);
	(void)write_all(adamic_stderr, "\n", 1);
	ADAMIC_COUNT_REPORT();
	_exit(70);
}

_Noreturn void adamic_unreachable(void) {
	static const char message[] = "compiler bug: a function ended without returning";
	adamic_panic(message, sizeof message - 1);
}
