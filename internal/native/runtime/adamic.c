// adamic.c: the Adamic runtime.

// SIGPIPE and isatty are POSIX's, which strict C11 doesn't show without asking.
#define _POSIX_C_SOURCE 200809L

#include "adamic.h"
#include "count.h"

#include <errno.h>
#ifndef ADAMIC_TARGET_WASI
#include <fcntl.h>
#include <poll.h>
#include <signal.h>
#endif
#include <stdatomic.h>
#include <stdlib.h>
#include <string.h>
#ifndef ADAMIC_TARGET_WASI
#include <sys/stat.h>
#endif
#include <unistd.h>

// write_all writes every byte, through partial writes and interrupted calls. It reports failure
// after waiting through EAGAIN on a non-blocking descriptor. A closed reader is a failure.
static int write_all(int descriptor, const char *bytes, size_t length) {
	while (length > 0) {
		ssize_t written = write(descriptor, bytes, length);
		if (written < 0) {
			if (errno == EINTR) {
				continue;
			}
#ifndef ADAMIC_TARGET_WASI
			if (errno == EAGAIN || errno == EWOULDBLOCK) {
				struct pollfd ready = {descriptor, POLLOUT, 0};
				int result;
				do {
					result = poll(&ready, 1, -1);
				} while (result < 0 && errno == EINTR);
				if (result > 0 && (ready.revents & POLLOUT)) {
					continue;
				}
			}
#endif
			return -1;
		}
		if (written == 0) {
			return -1;
		}
		bytes += written;
		length -= (size_t)written;
	}
	return 0;
}

// Node writes each line synchronously to regular files and terminals on POSIX. Match those
// destinations so a live log reader sees each line. Pipes retain the 64 KiB buffer, flushed before
// stderr, file operations, panic, exit and catchable external stop signals. Readers get the same bytes
// in larger pieces. SIGKILL cannot be caught, so a pipe's buffered output is lost then.
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

// output_mode is 0 until the first line, then 1 for a pipe buffer, or 2 for a file or terminal.
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

// stopped handles external stop signals: the buffer's whole lines go out, and the
// signal is raised again with its default action, so the program ends the way Node's does, killed by
// it. write, poll, sigaction and raise are async-signal-safe POSIX calls.
// A signal arriving inside flush finds the buffer already emptied, and
// what that write hadn't finished is lost.
#ifndef ADAMIC_TARGET_WASI
static struct sigaction default_action;

static void stopped(int signal_number) {
	size_t whole = output_whole;
	if (whole > 0 && !broken[adamic_stdout]) {
		(void)write_all(adamic_stdout, output, whole);
	}
	(void)sigaction(signal_number, &default_action, NULL);
	(void)raise(signal_number);
}

// Node resets inherited ignored SIGINT, SIGHUP and SIGTERM at startup. Other ignored signals
// stay ignored; they would not terminate the process with their inherited disposition.
static void stop_with(int signal_number) {
	struct sigaction current;
	if (sigaction(signal_number, NULL, &current) != 0 || (current.sa_handler == SIG_IGN && signal_number != SIGINT && signal_number != SIGHUP && signal_number != SIGTERM)) {
		return;
	}
	struct sigaction handler;
	memset(&handler, 0, sizeof handler);
	handler.sa_handler = stopped;
	sigfillset(&handler.sa_mask);
	sigaction(signal_number, &handler, NULL);
}

#endif

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
#ifndef ADAMIC_TARGET_WASI
	// Node opens /dev/null for any closed standard descriptor, in descriptor order. Do this
	// before opening anything else, so a closed stdout cannot become a program's input file.
	for (int descriptor = 0; descriptor < 3; descriptor++) {
		if (fcntl(descriptor, F_GETFD) < 0 && errno == EBADF) {
			int opened = open("/dev/null", O_RDWR);
			if (opened < 0 || (opened != descriptor && dup2(opened, descriptor) < 0)) {
				_exit(70);
			}
			if (opened != descriptor) {
				close(opened);
			}
		}
	}
#endif
	adamic_arguments_save(count, values);
	// Node ignores SIGPIPE, and a write to a pipe nobody reads is a failed write, not a killed process.
#ifndef ADAMIC_TARGET_WASI
	signal(SIGPIPE, SIG_IGN);
	// Node also ignores file-size-limit signals: writes report EFBIG instead.
	signal(SIGXFSZ, SIG_IGN);
	// SIGUSR1 starts Node's inspector, and the program goes on. There's no inspector here, but the
	// program goes on too.
	signal(SIGUSR1, SIG_IGN);
	default_action.sa_handler = SIG_DFL;
	sigemptyset(&default_action.sa_mask);
	stop_with(SIGTERM);
	stop_with(SIGINT);
	stop_with(SIGHUP);
	// Signals used to stop a process from outside. Leave fault and abort dispositions alone,
	// especially the sanitizer handlers that report runtime bugs with a stack. SIGKILL cannot
	// be caught; SIGPIPE, SIGXFSZ and SIGUSR1 are ignored above. Optional names stay guarded.
	const int fatal[] = {
		SIGQUIT, SIGUSR2,
		SIGALRM, SIGXCPU, SIGVTALRM, SIGPROF,
#ifdef SIGIO
		SIGIO,
#endif
#ifdef SIGPWR
		SIGPWR,
#endif
	};
	for (size_t index = 0; index < sizeof fatal / sizeof fatal[0]; index++) {
		stop_with(fatal[index]);
	}
#if defined(SIGRTMIN) && defined(SIGRTMAX)
	// Realtime signals also terminate by default. The C library excludes its reserved signals.
	for (int signal_number = SIGRTMIN; signal_number <= SIGRTMAX; signal_number++) {
		stop_with(signal_number);
	}
#endif
#endif
}

void adamic_write_line(enum adamic_stream stream, const adamic_string *string) {
	if (output_mode == 0) {
#ifndef ADAMIC_TARGET_WASI
		struct stat destination;
		bool regular_file = fstat(adamic_stdout, &destination) == 0 && S_ISREG(destination.st_mode);
		output_mode = isatty(adamic_stdout) || regular_file ? 2 : 1;
#else
		output_mode = isatty(adamic_stdout) ? 2 : 1;
#endif
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

_Noreturn void adamic_panic(const char *message, size_t length) {
	static const char prefix[] = "adamic: panic: ";
	// Everything the program printed comes first, as on Node.
	flush();
	// Best effort: stderr may be what failed.
	(void)write_all(adamic_stderr, prefix, sizeof prefix - 1);
	write_text(adamic_stderr, message, length);
	(void)write_all(adamic_stderr, "\n", 1);
	ADAMIC_COUNT_REPORT();
	_exit(70);
}

_Noreturn void adamic_unreachable(void) {
	static const char message[] = "compiler bug: a function ended without returning";
	adamic_panic(message, sizeof message - 1);
}
