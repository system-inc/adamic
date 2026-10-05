// stack.c: running out of stack is a panic, never a bare segfault (docs/0.1.md).
//
// Node gives up at about ten thousand frames with RangeError: Maximum call stack size exceeded. A
// native stack goes much deeper, so the depth differs, but both sides must stop the same way. Every
// function the compiler emits starts with ADAMIC_CHECK_STACK, which compares its frame against a limit
// set before main: the stack's size less a margin, below where the process began. The compare is the
// whole cost, and it works under the sanitizers, which a guard-page signal handler would fight with.

#define _POSIX_C_SOURCE 200809L

#include "adamic.h"

#include <stdint.h>
#include <sys/resource.h>

uintptr_t adamic_stack_limit;

// The stack assumed when the process's is unlimited, and the room kept below the limit for the frame
// that finds it crossed and for the panic that follows. Sanitized frames are several times larger.
#define ASSUMED_STACK ((uintptr_t)8 << 20)
#define MARGIN ((uintptr_t)256 << 10)

__attribute__((constructor)) static void find_stack_limit(void) {
	uintptr_t base = (uintptr_t)__builtin_frame_address(0);
	uintptr_t size = ASSUMED_STACK;
	struct rlimit limit;
	if (getrlimit(RLIMIT_STACK, &limit) == 0 && limit.rlim_cur != RLIM_INFINITY && limit.rlim_cur < size * 128) {
		size = (uintptr_t)limit.rlim_cur;
	}
	// A stack too small to keep the margin gets no check rather than one that fires at once.
	adamic_stack_limit = size > 2 * MARGIN && base > size ? base - size + MARGIN : 0;
}

_Noreturn void adamic_stack_overflow(void) {
	static const char message[] = "RangeError: Maximum call stack size exceeded";
	adamic_panic(message, sizeof message - 1);
}
