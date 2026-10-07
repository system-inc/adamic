// adamic.c: the Adamic runtime.

// SIGPIPE and isatty are POSIX's, which strict C11 doesn't show without asking.
#define _POSIX_C_SOURCE 200809L

#include "adamic.h"
#include "count.h"

#include <errno.h>
#include <signal.h>
#include <stdatomic.h>
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
// is written to stderr (which may be the same file), before a file is read or written (which may be
// stdout itself, or stdin waiting on what was just printed), before a panic's message, at exit, when
// it's full, after every line when stdout is a terminal, where a person is watching, and when SIGTERM,
// SIGINT or SIGHUP stops the program. What a reader of a pipe sees is the same bytes, in larger pieces.
// SIGKILL can't be caught, and what's in the buffer then is lost, as nothing Node holds would be.
//
// When a write fails (a pipe whose reader is gone, say), the program goes on as it does on Node,
// where the stream's error arrives only once the program's own code has run: what it writes there
// after is dropped, stderr is still written, and the exit is 70 rather than 0, as the oracle's
// runtime makes it.
static char output[1 << 16];
static size_t output_used;

// output_whole is where the buffer's last whole line ends: what a signal's handler writes out, so it
// writes whole lines, as Node would have. It's set after the line's bytes are in, with a fence between,
// so the handler never sees it ahead of them.
static volatile size_t output_whole;

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
		output_whole = 0;
		atomic_signal_fence(memory_order_seq_cst);
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

void adamic_output_flush(void) {
	flush();
}

// stopped is the handler for SIGTERM, SIGINT and SIGHUP: the buffer's whole lines go out, and the
// signal is raised again with its default action, so the program ends the way Node's does, killed by
// it. write is async-signal-safe. A signal arriving inside flush finds the buffer already emptied, and
// what that write hadn't finished is lost.
static void stopped(int signal_number) {
	size_t whole = output_whole;
	if (whole > 0 && !broken[adamic_stdout]) {
		(void)write_all(adamic_stdout, output, whole);
	}
	signal(signal_number, SIG_DFL);
	raise(signal_number);
}

// stop_with installs stopped for a signal, unless whoever started the program ignored it (a
// background job's SIGINT), which stays ignored, as it does for Node.
static void stop_with(int signal_number) {
	struct sigaction current;
	if (sigaction(signal_number, NULL, &current) != 0 || current.sa_handler == SIG_IGN) {
		return;
	}
	struct sigaction handler;
	memset(&handler, 0, sizeof handler);
	handler.sa_handler = stopped;
	sigemptyset(&handler.sa_mask);
	sigaction(signal_number, &handler, NULL);
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
	adamic_node_process_start(count, values);
	// Node ignores SIGPIPE, and a write to a pipe nobody reads is a failed write, not a killed process.
	signal(SIGPIPE, SIG_IGN);
	stop_with(SIGTERM);
	stop_with(SIGINT);
	stop_with(SIGHUP);
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
	if (stream == adamic_stdout) {
		atomic_signal_fence(memory_order_seq_cst);
		output_whole = output_used;
	}
	if (stream == adamic_stdout && output_mode == 2) {
		flush();
	}
}

// sys.write does not add a newline. Share console's encoding, failure and output-order rules.
bool adamic_write_raw(enum adamic_stream stream, const adamic_string *string) {
    if (output_mode == 0) {
        output_mode = isatty(adamic_stdout) ? 2 : 1;
        atexit(finish);
    }
    flush();
    write_text(stream, string->bytes, string->length);
    flush();
    return !broken[stream];
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

// An explicit exit does not unwind JavaScript frames or run their finally clauses.
_Noreturn void adamic_process_exit_now(int code) {
	flush();
	ADAMIC_COUNT_REPORT();
	_exit(code);
}
