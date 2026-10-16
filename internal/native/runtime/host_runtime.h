// host_runtime.h: runtime primitives for checked Node library bindings; no import lowering here.
#ifndef ADAMIC_HOST_RUNTIME_H
#define ADAMIC_HOST_RUNTIME_H
#include "adamic.h"

// Borrowed text, encoded as Node UTF-8; blocking whole write, no added newline; true on success.
bool adamic_write_raw(enum adamic_stream stream, const adamic_string *text);

// Startup runs once before user code. argv storage is borrowed for the process lifetime.
void adamic_host_start(int count, char **values);
// Optional validated Int32 status; false restores undefined. Setting it never exits.
void adamic_host_set_exit_code(bool present, int32_t code);
// Read optional exitCode; the output is authoritative only when true.
bool adamic_host_exit_code(int32_t *code);
// Natural return status (0 when undefined); the OS truncates status at termination.
int adamic_host_exit_status(void);
// Flush synchronous output, report intentional teardown, terminate without hooks or releases.
_Noreturn void adamic_process_exit_now(int code);
// Fractional monotonic milliseconds since runtime bootstrap; never the wall clock.
double adamic_host_performance_now(void);
// Fixed epoch-millisecond timestamp corresponding to the monotonic origin.
double adamic_host_time_origin(void);
// Current epoch milliseconds, truncated to whole milliseconds as Date.now.
double adamic_host_date_now(void);
// Owned decoded text; missing env is NULL, empty env is an owned empty string. Reads are live.
adamic_string *adamic_host_environment(const adamic_string *name);
// Owned current cwd (not memoized); NULL with errno on OS failure.
adamic_string *adamic_host_cwd(void);
// Owned absolute startup executable path, independent of argv[0] and later renames; NULL with errno if unavailable.
adamic_string *adamic_host_exec_path(void);
// Native bundled tsc identity equals execPath; default lib.*.d.ts files reside beside it.
adamic_string *adamic_host_executing_file_path(void);
// Owned arrays: [execPath, execPath, ...user args]; native has no VM arguments (execArgv=[]).
adamic_array *adamic_host_argv(void);
// Owned empty array: native images have no Node VM flags.
adamic_array *adamic_host_exec_argv(void);
// Platform EOL, immortal borrowed string.
adamic_string *adamic_host_eol(void);
// Native bytes, not V8: current RSS; allocator capacity/used; tracked numeric backing bytes.
typedef struct adamic_host_memory {
	double rss, heapTotal, heapUsed, external, arrayBuffers;
} adamic_host_memory;
// false with errno means unavailable; library must not fabricate supported metrics.
bool adamic_host_memory_usage(adamic_host_memory *result);
// Internal payload accounting: one add/drop per numeric backing allocation, never per view.
void adamic_host_backing_add(size_t bytes);
// Subtract storage only when the owner and its last view have released it.
void adamic_host_backing_drop(size_t bytes);
#endif
