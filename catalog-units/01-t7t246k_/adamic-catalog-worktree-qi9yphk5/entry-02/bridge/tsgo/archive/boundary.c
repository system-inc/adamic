#include "tsgo.h"
#include <stdlib.h>
#include <string.h>
#include <limits.h>

/* Private cgo entry points: callers use only the public functions below. */
extern int tsgo_go_create(tsgo_view, tsgo_view *, size_t, tsgo_handle *, tsgo_buffer *);
extern int tsgo_go_query(tsgo_handle, tsgo_view, uint64_t, tsgo_result *, tsgo_buffer *);
extern int tsgo_go_type_parts(tsgo_handle, tsgo_view, uint64_t, uint64_t, tsgo_view, tsgo_buffer *, tsgo_buffer *);
extern int tsgo_go_inspect(tsgo_handle, tsgo_view, uint64_t, uint64_t, tsgo_view, tsgo_view, tsgo_buffer *, tsgo_buffer *);
extern int tsgo_go_release(tsgo_handle, tsgo_buffer *);

static int copy_view(tsgo_view input, tsgo_view *output) {
 if (input.length > INT_MAX || (input.length != 0 && input.data == NULL)) return TSGO_ARGUMENT;
 char *copy = malloc(input.length == 0 ? 1 : input.length);
 if (copy == NULL) return TSGO_MEMORY;
 if (input.length != 0) memcpy(copy, input.data, input.length);
 output->data = copy;
 output->length = input.length;
 return TSGO_OK;
}

int tsgo_create(tsgo_view config, const tsgo_view *files, size_t count, tsgo_handle *handle, tsgo_buffer *error) {
 if (handle == NULL || error == NULL || (count != 0 && files == NULL) || count > INT_MAX / sizeof(tsgo_view)) return TSGO_ARGUMENT;
 *handle = 0;
 tsgo_view copied = {0};
 int status = copy_view(config, &copied);
 if (status != TSGO_OK) return status;
 tsgo_view *roots = calloc(count == 0 ? 1 : count, sizeof *roots);
 if (roots == NULL) { free((void *)copied.data); return TSGO_MEMORY; }
 size_t index = 0;
 for (; index < count; index++) {
  status = copy_view(files[index], &roots[index]);
  if (status != TSGO_OK) break;
 }
 if (status == TSGO_OK) status = tsgo_go_create(copied, roots, count, handle, error);
 for (size_t copied_index = 0; copied_index < index; copied_index++) free((void *)roots[copied_index].data);
 free(roots);
 free((void *)copied.data);
 return status;
}

int tsgo_query(tsgo_handle handle, tsgo_view file, uint64_t position, tsgo_result *result, tsgo_buffer *error) {
 if (result == NULL || error == NULL) return TSGO_ARGUMENT;
 tsgo_view copied = {0};
 int status = copy_view(file, &copied);
 if (status != TSGO_OK) return status;
 status = tsgo_go_query(handle, copied, position, result, error);
 free((void *)copied.data);
 return status;
}
int tsgo_type_parts(tsgo_handle handle, tsgo_view file, uint64_t start, uint64_t end, tsgo_view kind, tsgo_buffer *parts, tsgo_buffer *error) {
 if (parts == NULL || error == NULL) return TSGO_ARGUMENT;
 tsgo_view path = {0}, name = {0};
 int status = copy_view(file, &path);
 if (status != TSGO_OK) return status;
 status = copy_view(kind, &name);
 if (status == TSGO_OK) status = tsgo_go_type_parts(handle, path, start, end, name, parts, error);
 free((void *)path.data);
 free((void *)name.data);
 return status;
}
int tsgo_inspect(tsgo_handle handle, tsgo_view file, uint64_t start, uint64_t end, tsgo_view kind, tsgo_view question, tsgo_buffer *facts, tsgo_buffer *error) {
 if (facts == NULL || error == NULL) return TSGO_ARGUMENT;
 tsgo_view path = {0}, name = {0}, query = {0};
 int status = copy_view(file, &path);
 if (status != TSGO_OK) return status;
 status = copy_view(kind, &name);
 if (status == TSGO_OK) status = copy_view(question, &query);
 if (status == TSGO_OK) status = tsgo_go_inspect(handle, path, start, end, name, query, facts, error);
 free((void *)path.data); free((void *)name.data); free((void *)query.data);
 return status;
}
int tsgo_release(tsgo_handle handle, tsgo_buffer *error) {
 if (error == NULL) return TSGO_ARGUMENT;
 return tsgo_go_release(handle, error);
}
void tsgo_buffer_free(tsgo_buffer *buffer) {
 if (buffer == NULL) return;
 free(buffer->data);
 buffer->data = NULL;
 buffer->length = 0;
}
void tsgo_result_free(tsgo_result *result) {
 if (result == NULL) return;
 tsgo_buffer_free(&result->symbol);
 tsgo_buffer_free(&result->type);
 result->kind = 0;
}
