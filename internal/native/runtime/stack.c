// stack.c: running out of stack is a panic, never a bare segfault (docs/0.1.md).
//
// Node gives up at about ten thousand frames with RangeError: Maximum call stack size exceeded. A
// native stack goes much deeper, so the depth differs, but both sides must stop the same way. Every
// function the compiler emits starts with ADAMIC_CHECK_STACK, which compares its frame against a limit
// set before main: the stack's size less a margin, below where the process began. The compare is the
// whole cost, and it works under the sanitizers, which a guard-page signal handler would fight with.

#define _POSIX_C_SOURCE 200809L
#if defined(__APPLE__)
#define _DARWIN_C_SOURCE
#endif

#include "adamic.h"
#include "library_errors.h"

#include <stdint.h>
#if defined(__APPLE__)
#include <pthread.h>
#elif !defined(ADAMIC_TARGET_WASI)
#include <sys/resource.h>
#endif

uintptr_t adamic_stack_limit;

#ifdef ADAMIC_TARGET_WASI
// wasm-ld reserves a downward-growing linear stack. Keep 16 KB for the panic path.
extern unsigned char __stack_low;
__attribute__((constructor)) static void find_stack_limit(void) {
	adamic_stack_limit = (uintptr_t)&__stack_low + 16384;
}
#else

// The stack assumed when the process's is unlimited, and the room kept below the limit for the frame
// that finds it crossed and for the panic that follows, at most: sanitized frames are several times
// larger. A small stack keeps an eighth of itself instead.
#define ASSUMED_STACK ((uintptr_t)8 << 20)
#define MARGIN ((uintptr_t)256 << 10)

// The least Linux allows arguments and environment, whatever the stack (32 pages, ARG_MAX's floor).
#define ARGUMENTS_FLOOR ((uintptr_t)128 << 10)

__attribute__((constructor)) static void find_stack_limit(void) {
#if defined(__APPLE__)
	// Darwin reports this thread's actual stack, including the smaller iOS main stack.
	// RLIMIT_STACK describes the process limit, not necessarily the stack we run on.
	pthread_t thread = pthread_self();
	uintptr_t top = (uintptr_t)pthread_get_stackaddr_np(thread);
	uintptr_t size = (uintptr_t)pthread_get_stacksize_np(thread);
	uintptr_t margin = size / 8 < MARGIN ? size / 8 : MARGIN;
	adamic_stack_limit = size > margin && top > size ? top - size + margin : 0;
#else
	uintptr_t base = (uintptr_t)__builtin_frame_address(0);
	uintptr_t size = ASSUMED_STACK;
	struct rlimit limit;
	if (getrlimit(RLIMIT_STACK, &limit) == 0 && limit.rlim_cur != RLIM_INFINITY && limit.rlim_cur < size * 128) {
		size = (uintptr_t)limit.rlim_cur;
	}
	// The stack's size counts from its top, and above this first frame sit the program's arguments
	// and environment, as long as they are: three arguments of 120 KB are 360 KB the frame never sees,
	// and a limit counted from it once sat below the stack's real end, so deep recursion crashed
	// before the check fired. The system caps them at a quarter of the stack (Linux's execve allows
	// arguments and environment a quarter of the stack's limit, and macOS's ARG_MAX of 1 MB is less
	// than a quarter of its default 8 MB), so a quarter is kept back for them, and never less than
	// the 128 KiB Linux allows them on a small stack. Measuring where they really end would be closer,
	// but then how deep a program gets would change with its environment, and this way it never does.
	//
	// The margin shrinks with a small stack, so one of 1 MiB or 512 KiB still gets a limit: a fixed
	// 256 KiB, with the quarter, once left nothing at 1 MiB, and recursion there crashed where Node
	// panics. What's left to run in is then 5/8 of the stack, less on the smallest.
	uintptr_t arguments = size / 4 > ARGUMENTS_FLOOR ? size / 4 : ARGUMENTS_FLOOR;
	uintptr_t margin = size / 8 < MARGIN ? size / 8 : MARGIN;
	uintptr_t reserved = arguments + margin;
	// A stack too small to keep even that gets no check rather than one that fires at once.
	adamic_stack_limit = size > reserved && base > size ? base - size + reserved : 0;
#endif
}

#endif

_Noreturn void adamic_stack_overflow(void) {
	static const char message[] = "RangeError: Maximum call stack size exceeded";
	adamic_uncaught_library_error(message, sizeof message - 1);
}
