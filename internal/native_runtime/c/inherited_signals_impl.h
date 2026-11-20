// inherited_signals_impl.h: the reset of every signal a parent left ignored, included once by adamic.c.
//
// Node resets every signal its parent ignored (under nohup, or as a background job) to its default at
// startup, all but SIGKILL and SIGSTOP, so the program still stops on them. Asking each signal's
// disposition costs about sixty sigaction calls on Linux, most of a hello world's startup against C's
// (#7z36qn6). The kernel already says which signals are ignored, in /proc/self/status's SigIgn mask, so on
// Linux one read names them and only those take a sigaction. Wherever the mask can't be read whole (no
// /proc in a sandbox or a chroot, a line that isn't there, a mask that isn't hex), and on every other
// system, each signal is asked, as before. A mask that can't be read never means nothing was ignored.

#include <errno.h>
#include <fcntl.h>
#include <signal.h>
#include <stdbool.h>
#include <stdint.h>
#include <string.h>
#include <unistd.h>

// inherited_ignores_sweep asks every signal's disposition, and sets each ignored one to its default.
static void inherited_ignores_sweep(const struct sigaction *to_default) {
#if defined(SIGRTMAX)
	int last = SIGRTMAX;
#else
	int last = 31;
#endif
	for (int signal_number = 1; signal_number <= last; signal_number++) {
		struct sigaction current;
		if (signal_number != SIGKILL && signal_number != SIGSTOP && sigaction(signal_number, NULL, &current) == 0 && current.sa_handler == SIG_IGN) {
			(void)sigaction(signal_number, to_default, NULL);
		}
	}
}

#ifdef __linux__
// inherited_ignores_mask reads a status file's SigIgn line into mask, bit n - 1 for signal n, as the kernel
// prints it: one to sixteen hex digits ending the line. It answers false for a file it can't open or read,
// a line that isn't there, and any mask that isn't that, so the caller asks each signal instead.
static bool inherited_ignores_mask(const char *path, uint64_t *mask) {
	int descriptor;
	do {
		descriptor = open(path, O_RDONLY | O_CLOEXEC);
	} while (descriptor < 0 && errno == EINTR);
	if (descriptor < 0) {
		return false;
	}
	// The whole file is about 1.5 KB, and SigIgn is a third of the way in. A file cut short by the
	// buffer ends in the terminating zero, which is no hex digit and no newline, so it's refused.
	char text[4096];
	size_t used = 0;
	while (used < sizeof text - 1) {
		ssize_t received = read(descriptor, text + used, sizeof text - 1 - used);
		if (received < 0 && errno == EINTR) {
			continue;
		}
		if (received < 0) {
			close(descriptor);
			return false;
		}
		if (received == 0) {
			break;
		}
		used += (size_t)received;
	}
	close(descriptor);
	text[used] = '\0';
	// The file opens with Name:, so the line always follows a newline.
	const char *label = "\nSigIgn:\t";
	const char *line = strstr(text, label);
	if (line == NULL) {
		return false;
	}
	uint64_t value = 0;
	int digits = 0;
	for (const char *at = line + strlen(label); *at != '\n'; at++) {
		int nibble = *at >= '0' && *at <= '9' ? *at - '0' : *at >= 'a' && *at <= 'f' ? *at - 'a' + 10 : *at >= 'A' && *at <= 'F' ? *at - 'A' + 10 : -1;
		if (nibble < 0 || digits == 16) {
			return false;
		}
		value = value << 4 | (uint64_t)nibble;
		digits++;
	}
	if (digits == 0) {
		return false;
	}
	*mask = value;
	return true;
}
#endif

// inherited_ignores_reset sets every signal the process ignores to its default, but SIGKILL and SIGSTOP,
// reading which are ignored from the status file at path on Linux (adamic.c passes /proc/self/status).
static void inherited_ignores_reset(const char *path, const struct sigaction *to_default) {
#ifdef __linux__
	uint64_t mask;
	if (inherited_ignores_mask(path, &mask)) {
		for (int signal_number = 1; signal_number <= 64; signal_number++) {
			if ((mask >> (signal_number - 1) & 1) != 0 && signal_number != SIGKILL && signal_number != SIGSTOP) {
				(void)sigaction(signal_number, to_default, NULL);
			}
		}
		return;
	}
#else
	(void)path;
#endif
	inherited_ignores_sweep(to_default);
}
