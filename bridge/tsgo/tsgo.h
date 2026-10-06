/* The typescript-go bridge ABI, version 1. No Go pointers cross this boundary.
 * Inputs are borrowed UTF-8 bytes for the duration of a call, never terminated.
 * data may be NULL only when length is zero. C copies inputs before entering Go.
 * Paths must contain no NUL. Explicit file names are relative to the config's
 * directory; an empty list uses the config's roots. Query paths are relative to
 * the process working directory. Positions count UTF-8 bytes, starting at zero.
 * Queries use half-open AST ranges, including leading trivia; EOF is an error.
 *
 * Each successful create returns an owned, nonzero handle. Release it exactly
 * once. Handles are monotonic IDs, never pointers and never reused. Calls are
 * serialized; release waits for a query in progress. A stale handle is an error.
 * Every output buffer (including errors) is C malloc memory owned by the caller.
 * Free it with tsgo_buffer_free; free each query result with tsgo_result_free.
 * Output allocations contain exactly length bytes, with NO extra terminator.
 * Outputs must be zero-initialized and freed before reuse. Free clears outputs.
 * The Go runtime/collector belongs to the external checker library. Releasing a
 * handle drops all its Go roots; it does not shut down the process's Go runtime.
 * ASan checks the C allocations and copies, not the checker's Go heap.
 */
#ifndef ADAMIC_TSGO_H
#define ADAMIC_TSGO_H
#include <stddef.h>
#include <stdint.h>
#ifdef __cplusplus
extern "C" {
#endif
#define TSGO_ABI_VERSION 1
typedef uint64_t tsgo_handle;
typedef struct { const char *data; size_t length; } tsgo_view;
typedef struct { char *data; size_t length; } tsgo_buffer;
typedef struct { uint32_t kind; tsgo_buffer symbol; tsgo_buffer type; } tsgo_result;
enum tsgo_status { TSGO_OK = 0, TSGO_ARGUMENT = 1, TSGO_HANDLE = 2, TSGO_CHECKER = 3, TSGO_MEMORY = 4 };
int tsgo_create(tsgo_view config, const tsgo_view *files, size_t count, tsgo_handle *handle, tsgo_buffer *error);
int tsgo_query(tsgo_handle handle, tsgo_view file, uint64_t position, tsgo_result *result, tsgo_buffer *error);
int tsgo_release(tsgo_handle handle, tsgo_buffer *error);
void tsgo_buffer_free(tsgo_buffer *buffer);
void tsgo_result_free(tsgo_result *result);
#ifdef __cplusplus
}
#endif
#endif
