// inherited_ignores.c drives inherited_signals_impl.h's reset by hand (inherited_ignores_test.go). The test
// puts the header's text, as it is or mutated, where its include stands. The harness ignores SIGUSR2,
// SIGALRM, SIGWINCH and a real-time signal, as a parent would, resets them through the status file its one
// argument names, and prints the signals ignored before and after, and the sigaction calls the reset made.
#define _POSIX_C_SOURCE 200809L
// SIGWINCH isn't POSIX: macOS hides it once _POSIX_C_SOURCE is set, unless Darwin's own extensions are
// asked for too, as adamic.c does. glibc ignores this macro.
#define _DARWIN_C_SOURCE
#include <errno.h>
#include <fcntl.h>
#include <signal.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdio.h>
#include <string.h>
#include <unistd.h>

static int calls;

static int counted_sigaction(int signal_number, const struct sigaction *action, struct sigaction *previous) {
	calls++;
	return sigaction(signal_number, action, previous);
}

// Only the call form is renamed, so struct sigaction keeps its name.
#define sigaction(signal_number, action, previous) counted_sigaction(signal_number, action, previous)
#include "inherited_signals_impl.h"
#undef sigaction

static void print_ignored(const char *when) {
#if defined(SIGRTMAX)
	int last = SIGRTMAX;
#else
	int last = 31;
#endif
	printf("ignored %s:", when);
	for (int signal_number = 1; signal_number <= last; signal_number++) {
		struct sigaction current;
		if (signal_number != SIGKILL && signal_number != SIGSTOP && sigaction(signal_number, NULL, &current) == 0 && current.sa_handler == SIG_IGN) {
			printf(" %d", signal_number);
		}
	}
	printf("\n");
}

int main(int count, char **values) {
	if (count != 2) {
		fprintf(stderr, "usage: inherited_ignores <status file>\n");
		return 2;
	}
	int parent[] = {
		SIGUSR2, SIGALRM, SIGWINCH,
#if defined(SIGRTMIN)
		SIGRTMIN + 3,
#endif
	};
	for (size_t index = 0; index < sizeof parent / sizeof parent[0]; index++) {
		if (signal(parent[index], SIG_IGN) == SIG_ERR) {
			fprintf(stderr, "cannot ignore signal %d\n", parent[index]);
			return 2;
		}
	}
	print_ignored("before");
	struct sigaction to_default;
	memset(&to_default, 0, sizeof to_default);
	to_default.sa_handler = SIG_DFL;
	sigemptyset(&to_default.sa_mask);
	calls = 0;
	inherited_ignores_reset(values[1], &to_default);
	int made = calls;
	print_ignored("after");
	printf("calls %d\n", made);
	return 0;
}
