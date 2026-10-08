// adamic.c: the Adamic runtime.

// SIGPIPE and isatty are POSIX's, which strict C11 doesn't show without asking.
#define _POSIX_C_SOURCE 200809L

#include "adamic.h"
#include "count.h"
#include "parallel.h"
#include "host_runtime.h"
#include <stdio.h>

#include <errno.h>
#include <pthread.h>
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

// output_mode is 0 until the first line, then 1 for a pipe buffer, or 2 for a file or terminal.
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
static struct sigaction default_action;

// Catch external stop signals; preserve fault/abort handlers installed by sanitizers.
static const int stop_signals[] = {
	SIGTERM, SIGINT, SIGHUP, SIGQUIT, SIGUSR2,
	SIGALRM, SIGXCPU, SIGVTALRM, SIGPROF,
#ifdef SIGIO
	SIGIO,
#endif
#ifdef SIGPWR
	SIGPWR,
#endif
};

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
		// A writer holds output_lock until its newline is buffered. Waiting for that lock
		// flushes complete lines without racing or interrupting a writer midway through one.
		pthread_mutex_lock(&output_lock);
		flush();
		// Keep writers out until the default action terminates the process.
		sigaction(event, &default_action, NULL);
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
	for (size_t index = 0; index < sizeof stop_signals / sizeof stop_signals[0]; index++) {
		sigaddset(&blocked, stop_signals[index]);
	}
#if defined(SIGRTMIN) && defined(SIGRTMAX)
	for (int number = SIGRTMIN; number <= SIGRTMAX; number++) { sigaddset(&blocked, number); }
#endif
	pthread_sigmask(SIG_BLOCK, &blocked, &previous);
	int error = pthread_create(&stop_thread, NULL, stop_loop, NULL);
	pthread_sigmask(SIG_SETMASK, &previous, NULL);
	if (error != 0 || atexit(stop_end) != 0) {
		adamic_panic("cannot start signal loop", 24);
	}
}

// stop_with installs stopped for a signal, unless the program itself ignores it (SIGPIPE, SIGXFSZ
// and SIGUSR1, as Node does). What a parent ignored was reset at startup (adamic_start).
static void stop_with(int signal_number) {
	struct sigaction current;
	if (sigaction(signal_number, NULL, &current) != 0 || current.sa_handler == SIG_IGN) {
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
	adamic_host_start(count, values);
	adamic_node_process_start(count, values);
#ifndef ADAMIC_TARGET_WASI
	default_action.sa_handler = SIG_DFL;
	sigemptyset(&default_action.sa_mask);
	// Node resets every signal its parent ignored (under nohup, or as a background job) to its
	// default at startup, so the program still stops on them; it then sets its own below.
#if defined(SIGRTMAX)
	int last = SIGRTMAX;
#else
	int last = 31;
#endif
	for (int signal_number = 1; signal_number <= last; signal_number++) {
		struct sigaction current;
		if (signal_number != SIGKILL && signal_number != SIGSTOP && sigaction(signal_number, NULL, &current) == 0 && current.sa_handler == SIG_IGN) {
			(void)sigaction(signal_number, &default_action, NULL);
		}
	}
	// Node ignores SIGPIPE, and a write to a pipe nobody reads is a failed write, not a killed process.
	signal(SIGPIPE, SIG_IGN);
	// Node also ignores file-size-limit signals: writes report EFBIG instead.
	signal(SIGXFSZ, SIG_IGN);
	// SIGUSR1 starts Node's inspector, and the program goes on. There's no inspector here, but the
	// program goes on too.
	signal(SIGUSR1, SIG_IGN);
	stop_start();
	for (size_t index = 0; index < sizeof stop_signals / sizeof stop_signals[0]; index++) {
		stop_with(stop_signals[index]);
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
	pthread_mutex_lock(&output_lock);
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
	if (stream == adamic_stdout && output_mode == 2) {
		flush();
	}
	pthread_mutex_unlock(&output_lock);
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
	write_text(adamic_stderr, message, length);
	(void)write_all(adamic_stderr, "\n", 1);
	ADAMIC_COUNT_REPORT();
	_exit(70);
}

_Noreturn void adamic_unreachable(void) {
	static const char message[] = "compiler bug: a function ended without returning";
	adamic_panic(message, sizeof message - 1);
}

// An explicit process exit bypasses all atexit hooks, worker joins, finally blocks and releases.
_Noreturn void adamic_process_exit_now(int code) {
	adamic_output_flush();
	(void)fflush(stdout);
	(void)fflush(stderr);
#ifdef ADAMIC_COUNT
	bool report = true;
#else
	bool report = getenv("ADAMIC_LEAK_CHECK") != NULL;
#endif
	if (report) {
		char line[80];
		int length = snprintf(line, sizeof line, "\nadamic: intentional exit: status %d\n", (unsigned char)code);
		if (length > 0 && (size_t)length < sizeof line) { (void)write_all(2, line, (size_t)length); }
	}
	ADAMIC_COUNT_REPORT();
	_exit(code);
}

// Checked stream binding: preserve stdout/stderr order and complete the write synchronously.
bool adamic_write_raw(enum adamic_stream stream, const adamic_string *text) {
	pthread_mutex_lock(&output_lock);
	flush();
	write_text(stream, text->bytes, text->length);
	flush();
	bool ok = !broken[stream];
	pthread_mutex_unlock(&output_lock);
	return ok;
}
