/* The typescript-go bridge ABI, version 1. No Go pointers cross this boundary.
 * Inputs are borrowed UTF-8 bytes for the duration of a call, never terminated.
 * data may be NULL only when length is zero. C copies inputs before entering Go.
 * Paths must contain no NUL. Explicit file names are relative to the config's
 * directory; an empty list uses the config's roots. Configured .d.ts roots are
 * retained with explicit roots. Explicit .a roots are read as TypeScript.
 * Query paths are relative to the process working directory. Positions count
 * UTF-8 bytes, starting at zero.
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
/* Exact node lookup: kind is the TypeScript kind name without "Kind"; start
 * includes leading trivia. A nonexistent kind/span is a checker error.
 * Return the constrained type's union parts (or one nonunion part), in checker
 * order. Each frame is decimal TypeFlags LF decimal UTF-16 units LF type text.
 * Frame text may contain newlines; consume its stated units before the next
 * frame. The enclosing buffer is UTF-8 with an explicit BYTE length and the
 * same ownership as every other output. Flags are pinned to this checker.
 * No rule predicate is applied by the bridge. */
int tsgo_type_parts(tsgo_handle handle, tsgo_view file, uint64_t start, uint64_t end, tsgo_view kind, tsgo_buffer *parts, tsgo_buffer *error);
/* Native type-aware rule facts, schema version 1, documented in bridge/tsgo/facts.md.
 * Exact node selection and buffer ownership are identical to tsgo_type_parts.
 * question is borrowed UTF-8 with an explicit byte length. Supported questions:
 * raw-type, type (constrained), base-type (constrained and literal-widened),
 * signature (resolved call), declarations (named class/interface symbols),
 * options (SourceFile strictNullChecks), raw-shape/type-shape/signature-shape
 * (the corresponding graph with empty names), name LF type-ID (TypeToString for
 * a previously returned type ID in this live program), and
 * assignable LF target-start LF target-end LF target-kind in this same file.
 * Additional compiler facts: strict-this (resolved noImplicitThis),
 * widened-shape, enum-types, scope-locals,
 * contextual-shape, call-returns, symbol-origin; assignable-types LF source-ID
 * LF target-ID; name/type-symbol/type-origin/call-count/apparent-shape/
 * base-shapes/call-parameters LF type-ID; property-shape/property-info LF type-ID
 * LF property-name. These return types, signatures, binder tables, declaration
 * shapes or declaration origins, never a lint decision. See facts.md for fields.
 * Unsupported questions and malformed/inexact targets are refused.
 * Facts contain no Go pointers. Type IDs are opaque, program-scoped, borrowed
 * identities, usable for equality and name queries while that program is live. Release
 * invalidates them all. Do not pass them as program handles. */
int tsgo_inspect(tsgo_handle handle, tsgo_view file, uint64_t start, uint64_t end, tsgo_view kind, tsgo_view question, tsgo_buffer *facts, tsgo_buffer *error);
int tsgo_release(tsgo_handle handle, tsgo_buffer *error);
void tsgo_buffer_free(tsgo_buffer *buffer);
void tsgo_result_free(tsgo_result *result);
#ifdef __cplusplus
}
#endif
#endif
