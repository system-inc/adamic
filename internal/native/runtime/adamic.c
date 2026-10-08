// adamic.c: the Adamic runtime.

// SIGPIPE and isatty are POSIX's, which strict C11 doesn't show without asking.
#define _POSIX_C_SOURCE 200809L

#include "adamic.h"
#include "count.h"
#include "parallel.h"

#include <errno.h>
#include <pthread.h>
#ifndef ADAMIC_TARGET_WASI
#include <signal.h>
#include <fcntl.h>
#endif
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

// output_mode is 0 until the first line, then 1 for a buffer, or 2 for a line at a time (a terminal).
static int output_mode;

// broken is the stream a write to has failed, from then on dropped; 0 while none has.
static bool broken[3];
static pthread_mutex_t output_lock = PTHREAD_MUTEX_INITIALIZER;
static atomic_flag panicking = ATOMIC_FLAG_INIT;

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
	// Registration can happen in a worker, after the pool registered its own exit hook.
	adamic_parallel_shutdown();
	pthread_mutex_lock(&output_lock);
	flush();
	bool failed = broken[adamic_stdout] || broken[adamic_stderr];
	pthread_mutex_unlock(&output_lock);
	if (failed) {
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
	pthread_mutex_lock(&output_lock);
	flush();
	pthread_mutex_unlock(&output_lock);
}

// The handler only forwards a byte to a nonblocking pipe. Descriptors are initialized before
// installing handlers and remain open and unchanged until process death, including during exit.
// The normal signal thread can wait for a writer's lock without deadlocking an interrupted writer.
#ifndef ADAMIC_TARGET_WASI
static int stop_pipe[2];
static pthread_t stop_thread;

static void stopped(int signal_number) {
	int saved_errno = errno;
	unsigned char event = (unsigned char)signal_number;
	ssize_t written;
	do { written = write(stop_pipe[1], &event, 1); } while (written < 0 && errno == EINTR);
	errno = saved_errno;
}

static void *stop_loop(void *unused) {
	(void)unused;
	unsigned char event;
	for (;;) {
		ssize_t received = read(stop_pipe[0], &event, 1);
		if (received < 0 && errno == EINTR) { continue; }
		if (received != 1 || event == 0) { return NULL; }
		adamic_output_flush();
		struct sigaction action;
		memset(&action, 0, sizeof action);
		action.sa_handler = SIG_DFL;
		sigemptyset(&action.sa_mask);
		sigaction(event, &action, NULL);
		sigset_t delivered;
		sigemptyset(&delivered);
		sigaddset(&delivered, event);
		pthread_sigmask(SIG_UNBLOCK, &delivered, NULL);
		raise(event);
	}
}

static void stop_end(void) {
	unsigned char event = 0;
	// The read end stays open: handlers on other threads never write to a reused descriptor.
	while (write(stop_pipe[1], &event, 1) < 0) {
		if (errno != EINTR && errno != EAGAIN) { break; }
	}
	pthread_join(stop_thread, NULL);
}

static void stop_start(void) {
	if (pipe(stop_pipe) != 0 || fcntl(stop_pipe[1], F_SETFL, O_NONBLOCK) != 0) {
		adamic_panic("cannot start signal loop", 24);
	}
	sigset_t blocked, previous;
	sigemptyset(&blocked);
	sigaddset(&blocked, SIGTERM);
	sigaddset(&blocked, SIGINT);
	sigaddset(&blocked, SIGHUP);
	pthread_sigmask(SIG_BLOCK, &blocked, &previous);
	int error = pthread_create(&stop_thread, NULL, stop_loop, NULL);
	pthread_sigmask(SIG_SETMASK, &previous, NULL);
	if (error != 0 || atexit(stop_end) != 0) {
		adamic_panic("cannot start signal loop", 24);
	}
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
	sigemptyset(&handler.sa_mask);
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
    adamic_error_init_prototypes();
    if (atexit(adamic_error_reset_prototypes) != 0) adamic_panic("Error prototype cleanup registration failed", sizeof("Error prototype cleanup registration failed") - 1);
	adamic_arguments_save(count, values);
	adamic_node_process_start(count, values);
	// Node ignores SIGPIPE, and a write to a pipe nobody reads is a failed write, not a killed process.
#ifndef ADAMIC_TARGET_WASI
	signal(SIGPIPE, SIG_IGN);
	stop_start();
	stop_with(SIGTERM);
	stop_with(SIGINT);
	stop_with(SIGHUP);
#endif
}

void adamic_write_line(enum adamic_stream stream, const adamic_string *string) {
	pthread_mutex_lock(&output_lock);
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
	pthread_mutex_unlock(&output_lock);
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
#ifndef ADAMIC_TARGET_WASI
	// Losing panics must not terminate the process before the winner finishes its one message.
	if (atomic_flag_test_and_set_explicit(&panicking, memory_order_relaxed)) {
		for (;;) { pause(); }
	}
#else
	(void)panicking;
#endif
	pthread_mutex_lock(&output_lock);
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
