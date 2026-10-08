/* Only linked when stage 0 is explicitly given the external checker archive. */
#ifdef ADAMIC_TSGO
#define _POSIX_C_SOURCE 200809L
#include "tsgo_runtime.h"
#include "tsgo.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <time.h>
#include <inttypes.h>

static uint64_t loaded_ns, queried_ns, first_query_ns, run_started_ns;
static size_t query_count;
static uint64_t input_ns, call_ns, output_ns, facts_bytes;
static bool profiling(void) { return getenv("ADAMIC_TSGO_PROFILE") != NULL; }
static bool timing(void) { return getenv("ADAMIC_TSGO_TIMING") != NULL; }
static uint64_t now_ns(void) {
 struct timespec now;
 if (clock_gettime(CLOCK_MONOTONIC, &now) != 0) {
  static const char failed[] = "tsgo: clock unavailable";
  adamic_panic(failed, sizeof failed - 1);
 }
 return (uint64_t)now.tv_sec * UINT64_C(1000000000) + (uint64_t)now.tv_nsec;
}

static void checked(int status, tsgo_buffer *error) {
 if (status == TSGO_OK) return;
 adamic_string *message;
 if (error->length != 0) {
  message = adamic_decode_utf8((const unsigned char *)error->data, error->length);
 } else {
  static const char failed[] = "tsgo: invalid argument or out of memory";
  message = adamic_string_allocate(sizeof failed - 1);
  memcpy((char *)message->bytes, failed, sizeof failed - 1);
 }
 tsgo_buffer_free(error);
 adamic_panic(message->bytes, message->length);
}

static uint64_t integer(double value) {
 if (!(value >= 0 && value <= 9007199254740991.0 && trunc(value) == value)) {
  static const char invalid[] = "tsgo: expected a nonnegative safe integer";
  adamic_panic(invalid, sizeof invalid - 1);
 }
 return (uint64_t)value;
}

/* Adamic stores WTF-8. Paths use the same UTF-8 replacement of lone surrogates
 * as file input does. The temporary belongs to this call, never to Go. */
static tsgo_view path_view(const adamic_string *path) {
 char *bytes = adamic_path_bytes(path);
 if (bytes == NULL) {
  static const char invalid[] = "tsgo: path contains NUL";
  adamic_panic(invalid, sizeof invalid - 1);
 }
 return (tsgo_view){bytes, path->length};
}

double adamic_tsgo_program(const adamic_string *config, const adamic_array *files) {
 uint64_t started = timing() ? now_ns() : 0;
 tsgo_view config_view = path_view(config);
 tsgo_view *roots = calloc(files->length == 0 ? 1 : files->length, sizeof *roots);
 if (roots == NULL) {
  static const char failed[] = "out of memory";
  adamic_panic(failed, sizeof failed - 1);
 }
 for (size_t index = 0; index < files->length; index++) roots[index] = path_view(files->elements[index].reference);
 tsgo_handle handle = 0;
 tsgo_buffer error = {0};
 int status = tsgo_create(config_view, roots, files->length, &handle, &error);
 for (size_t index = 0; index < files->length; index++) free((void *)roots[index].data);
 free(roots);
 free((void *)config_view.data);
 checked(status, &error);
 if (timing()) { loaded_ns += now_ns() - started; run_started_ns = now_ns(); }
 return (double)handle;
}

adamic_object *adamic_tsgo_query(double handle, const adamic_string *file, double position) {
 return adamic_tsgo_query_in(NULL, handle, file, position);
}

adamic_object *adamic_tsgo_query_in(adamic_region *region, double handle, const adamic_string *file, double position) {
 uint64_t started = timing() ? now_ns() : 0;
 uint64_t id = integer(handle), offset = integer(position);
 tsgo_view path = path_view(file);
 tsgo_result result = {0};
 tsgo_buffer error = {0};
 int status = tsgo_query(id, path, offset, &result, &error);
 free((void *)path.data);
 if (status != TSGO_OK) tsgo_result_free(&result);
 checked(status, &error);
 static const char *const names[] = {"nodeKind", "symbolName", "type"};
 static const bool references[] = {false, true, true};
 static const adamic_shape shape = {3, names, references, NULL, NULL};
 adamic_object *answer = adamic_object_new_in(region, &shape);
 answer->slots[0].number = (double)result.kind;
 answer->slots[1].reference = adamic_decode_utf8((const unsigned char *)result.symbol.data, result.symbol.length);
 answer->slots[2].reference = adamic_decode_utf8((const unsigned char *)result.type.data, result.type.length);
 tsgo_result_free(&result);
 if (timing()) { uint64_t elapsed = now_ns() - started;
  if (query_count == 0) first_query_ns = elapsed;
  queried_ns += elapsed; query_count++; }
 return answer;
}
adamic_string *adamic_tsgo_type_parts(double handle, const adamic_string *file, double start, double end, const adamic_string *kind) {
 bool detailed = profiling();
 uint64_t started = timing() || detailed ? now_ns() : 0;
 uint64_t id = integer(handle), first = integer(start), last = integer(end);
 tsgo_view path = path_view(file), name = path_view(kind);
 tsgo_buffer parts = {0}, error = {0};
 uint64_t entered = detailed ? now_ns() : 0;
 int status = tsgo_type_parts(id, path, first, last, name, &parts, &error);
 uint64_t returned = detailed ? now_ns() : 0;
 if (detailed) facts_bytes += parts.length;
 free((void *)path.data); free((void *)name.data);
 if (status != TSGO_OK) tsgo_buffer_free(&parts);
 checked(status, &error);
 adamic_string *answer = adamic_decode_utf8((const unsigned char *)parts.data, parts.length);
 tsgo_buffer_free(&parts);
 if (detailed) { input_ns += entered - started; call_ns += returned - entered; output_ns += now_ns() - returned; }
 if (timing()) { uint64_t elapsed = now_ns() - started;
  if (query_count == 0) first_query_ns = elapsed;
  queried_ns += elapsed; query_count++; }
 return answer;
}
adamic_string *adamic_tsgo_inspect(double handle, const adamic_string *file, double start, double end, const adamic_string *kind, const adamic_string *question) {
 bool detailed = profiling();
 uint64_t started = timing() || detailed ? now_ns() : 0;
 uint64_t id = integer(handle), first = integer(start), last = integer(end);
 tsgo_view path = path_view(file), name = path_view(kind), query = path_view(question);
 tsgo_buffer facts = {0}, error = {0};
 uint64_t entered = detailed ? now_ns() : 0;
 int status = tsgo_inspect(id, path, first, last, name, query, &facts, &error);
 uint64_t returned = detailed ? now_ns() : 0;
 if (detailed) facts_bytes += facts.length;
 free((void *)path.data); free((void *)name.data); free((void *)query.data);
 if (status != TSGO_OK) tsgo_buffer_free(&facts);
 checked(status, &error);
 adamic_string *answer = adamic_decode_utf8((const unsigned char *)facts.data, facts.length);
 tsgo_buffer_free(&facts);
 if (detailed) { input_ns += entered - started; call_ns += returned - entered; output_ns += now_ns() - returned; }
 if (timing()) { uint64_t elapsed = now_ns() - started;
  if (query_count == 0) first_query_ns = elapsed;
  queried_ns += elapsed; query_count++; }
 return answer;
}
void adamic_tsgo_release(double handle) {
 uint64_t run_ns = timing() && run_started_ns != 0 ? now_ns() - run_started_ns : 0;
 tsgo_buffer error = {0};
 checked(tsgo_release(integer(handle), &error), &error);
 if (profiling()) {
  fprintf(stderr, "tsgo_c_profile: input_ns=%" PRIu64 " call_ns=%" PRIu64 " output_ns=%" PRIu64 " facts_bytes=%" PRIu64 "\n", input_ns, call_ns, output_ns, facts_bytes);
  input_ns = 0; call_ns = 0; output_ns = 0; facts_bytes = 0;
 }
 if (timing()) {
  adamic_output_flush();
  fprintf(stderr, "tsgo: load_ns=%" PRIu64 " query_ns=%" PRIu64 " queries=%zu first_query_ns=%" PRIu64 " run_ns=%" PRIu64 "\n", loaded_ns, queried_ns, query_count, first_query_ns, run_ns);
  loaded_ns = 0; queried_ns = 0; first_query_ns = 0; query_count = 0; run_started_ns = 0;
 }
}
#else
/* C11 forbids an empty translation unit under -pedantic. */
typedef int adamic_tsgo_not_linked;
#endif
