// The external checker is outside this unit. A stateless ABI witness isolates
// the C wrapper's real allocation and profiling paths from Go's runtime/TSan.
#include "tsgo.h"
#include <stdlib.h>
int tsgo_create(tsgo_view config, const tsgo_view *files, size_t count, tsgo_handle *handle, tsgo_buffer *error) {
 (void)config; (void)files; (void)count; (void)error; *handle = 1; return TSGO_OK;
}
int tsgo_query(tsgo_handle handle, tsgo_view file, uint64_t position, tsgo_result *result, tsgo_buffer *error) {
 (void)handle; (void)file; (void)position; (void)result; (void)error; return TSGO_OK;
}
int tsgo_type_parts(tsgo_handle handle, tsgo_view file, uint64_t start, uint64_t end, tsgo_view kind, tsgo_buffer *parts, tsgo_buffer *error) {
 (void)handle; (void)file; (void)start; (void)end; (void)kind; (void)parts; (void)error; return TSGO_OK;
}
int tsgo_inspect(tsgo_handle handle, tsgo_view file, uint64_t start, uint64_t end, tsgo_view kind, tsgo_view question, tsgo_buffer *facts, tsgo_buffer *error) {
 (void)handle; (void)file; (void)start; (void)end; (void)kind; (void)question; (void)facts; (void)error; return TSGO_OK;
}
int tsgo_release(tsgo_handle handle, tsgo_buffer *error) { (void)handle; (void)error; return TSGO_OK; }
void tsgo_buffer_free(tsgo_buffer *buffer) { free(buffer->data); *buffer = (tsgo_buffer){0}; }
void tsgo_result_free(tsgo_result *result) { tsgo_buffer_free(&result->symbol); tsgo_buffer_free(&result->type); }
