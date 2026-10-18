#include "harness.h"
#include "host_runtime.h"
static adamic_string line = ADAMIC_STRING("host worker\n");
// memoryUsage reads glibc's mallinfo2, which ThreadSanitizer doesn't intercept: its first call
// sets up glibc's own arena state, and doing that from a worker under TSan can fail glibc's
// attach count at thread exit (arena.c __malloc_arena_thread_freeres). Make the first call here,
// on the main thread; the workers still call it in exercise.
static void prepare(void) {
 adamic_host_memory memory;
 if (!adamic_host_memory_usage(&memory)) { abort(); }
}
static void cleanup(void) {}
static double exercise(size_t index) {
 adamic_host_set_exit_code(true, (int32_t)index);
 int32_t code;
 if (!adamic_host_exit_code(&code) || code < 0 || code >= 256) { abort(); }
 (void)adamic_host_exit_status();
 adamic_string *path = adamic_host_exec_path();
 adamic_array *args = adamic_host_argv();
 if (path == NULL || args == NULL || path->length == 0) { abort(); }
 adamic_release(path); adamic_release(args);
 if (adamic_host_time_origin() <= 0 || adamic_host_performance_now() < 0) { abort(); }
 adamic_string *eol = adamic_host_eol();
 if (adamic_string_length(eol) != 1) { abort(); }
 adamic_release(adamic_retain(eol));
 adamic_host_backing_add(8);
 adamic_host_memory memory;
 if (!adamic_host_memory_usage(&memory)) { abort(); }
 adamic_host_backing_drop(8);
 if (!adamic_write_raw(adamic_stdout, &line)) { abort(); }
 return 1;
}
