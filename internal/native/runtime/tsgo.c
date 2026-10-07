/* Only linked when stage 0 is explicitly given the external checker archive. */
#if defined(ADAMIC_TSGO) && !defined(ADAMIC_TARGET_WASI)
#define _POSIX_C_SOURCE 200809L
#define _DARWIN_C_SOURCE
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
 if (clock_gettime(CLOCK_MONOTONIC, &now) != 0) return 0;
 return (uint64_t)now.tv_sec * UINT64_C(1000000000) + (uint64_t)now.tv_nsec;
}

static adamic_string ok_kind = ADAMIC_STRING("Ok");
static adamic_string error_kind = ADAMIC_STRING("Error");
static const char *const error_names[] = {"kind", "message"};
static const bool error_refs[] = {true, true};
static const adamic_shape error_shape = {2, error_names, error_refs, NULL};
static adamic_object *failure(adamic_region *region, adamic_string *message) {
 adamic_object *result = adamic_object_new_in(region, &error_shape);
 result->slots[0].reference = &error_kind;
 result->slots[1].reference = message;
 return result;
}
static adamic_object *invalid(adamic_region *region, const char *message) {
 return failure(region, adamic_decode_utf8((const unsigned char *)message, strlen(message)));
}
static adamic_object *checked(adamic_region *region, int status, tsgo_buffer *error) {
 if (status == TSGO_OK) { tsgo_buffer_free(error); return NULL; }
 adamic_string *message = error->length != 0
  ? adamic_decode_utf8((const unsigned char *)error->data, error->length)
  : adamic_decode_utf8((const unsigned char *)"tsgo: invalid argument or out of memory", sizeof "tsgo: invalid argument or out of memory" - 1);
 tsgo_buffer_free(error);
 return failure(region, message);
}
static bool integer(double value) {
 return value >= 0 && value <= 9007199254740991.0 && trunc(value) == value;
}
/* WTF-8 paths become UTF-8 with lone surrogates replaced, as input doors do.
 * Input validation happens before allocating any temporary views. */
static bool path_valid(const adamic_string *path) {
 return memchr(path->bytes, 0, path->length) == NULL;
}
static tsgo_view path_view(const adamic_string *path) {
 return (tsgo_view){adamic_path_bytes(path), path->length};
}
static adamic_object *success(adamic_region *region, adamic_value value, bool reference) {
 static const char *const success_names[] = {"kind", "value"};
 static const bool number_refs[] = {true, false}, value_refs[] = {true, true};
 static const adamic_shape number_shape = {2, success_names, number_refs, NULL};
 static const adamic_shape value_shape = {2, success_names, value_refs, NULL};
 adamic_object *result = adamic_object_new_in(region, reference ? &value_shape : &number_shape);
 result->slots[0].reference = &ok_kind;
 result->slots[1] = value;
 return result;
}

adamic_object *adamic_tsgo_program_in(adamic_region *region, const adamic_string *config, const adamic_array *files) {
 uint64_t started = timing() ? now_ns() : 0;
 if (!path_valid(config)) return invalid(region, "tsgo: path contains NUL");
 for (size_t index = 0; index < files->length; index++)
  if (!path_valid(files->elements[index].reference)) return invalid(region, "tsgo: path contains NUL");
 tsgo_view config_view = path_view(config);
 tsgo_view *roots = calloc(files->length == 0 ? 1 : files->length, sizeof *roots);
 if (roots == NULL) {
  free((void *)config_view.data);
  return invalid(region, "tsgo: invalid argument or out of memory");
 }
 for (size_t index = 0; index < files->length; index++) roots[index] = path_view(files->elements[index].reference);
 tsgo_handle handle = 0;
 tsgo_buffer error = {0};
 int status = tsgo_create(config_view, roots, files->length, &handle, &error);
 for (size_t index = 0; index < files->length; index++) free((void *)roots[index].data);
 free(roots);
 free((void *)config_view.data);
 adamic_object *failed = checked(region, status, &error);
 if (failed != NULL) return failed;
 if (timing()) { loaded_ns += now_ns() - started; run_started_ns = now_ns(); }
 return success(region, (adamic_value){.number = (double)handle}, false);
}

adamic_object *adamic_tsgo_query(double handle, const adamic_string *file, double position) {
 return adamic_tsgo_query_in(NULL, handle, file, position);
}

adamic_object *adamic_tsgo_query_in(adamic_region *region, double handle, const adamic_string *file, double position) {
 uint64_t started = timing() ? now_ns() : 0;
 if (!integer(handle) || !integer(position)) return invalid(region, "tsgo: expected a nonnegative safe integer");
 if (!path_valid(file)) return invalid(region, "tsgo: path contains NUL");
 uint64_t id = (uint64_t)handle, offset = (uint64_t)position;
 tsgo_view path = path_view(file);
 tsgo_result result = {0};
 tsgo_buffer error = {0};
 int status = tsgo_query(id, path, offset, &result, &error);
 free((void *)path.data);
 if (status != TSGO_OK) tsgo_result_free(&result);
 adamic_object *failed = checked(region, status, &error);
 if (failed != NULL) return failed;
 static const char *const query_names[] = {"nodeKind", "symbolName", "type"};
 static const bool references[] = {false, true, true};
 static const adamic_shape shape = {3, query_names, references, NULL};
 adamic_object *answer = adamic_object_new_in(region, &shape);
 answer->slots[0].number = (double)result.kind;
 answer->slots[1].reference = adamic_decode_utf8((const unsigned char *)result.symbol.data, result.symbol.length);
 answer->slots[2].reference = adamic_decode_utf8((const unsigned char *)result.type.data, result.type.length);
 tsgo_result_free(&result);
 if (timing()) { uint64_t elapsed = now_ns() - started;
  if (query_count == 0) first_query_ns = elapsed;
  queried_ns += elapsed; query_count++; }
 return success(region, (adamic_value){.reference = answer}, true);
}
adamic_object *adamic_tsgo_type_parts_in(adamic_region *region, double handle, const adamic_string *file, double start, double end, const adamic_string *kind) {
 bool detailed = profiling();
 uint64_t started = timing() || detailed ? now_ns() : 0;
 if (!integer(handle) || !integer(start) || !integer(end)) return invalid(region, "tsgo: expected a nonnegative safe integer");
 if (!path_valid(file) || !path_valid(kind)) return invalid(region, "tsgo: path contains NUL");
 uint64_t id = (uint64_t)handle, first = (uint64_t)start, last = (uint64_t)end;
 tsgo_view path = path_view(file), name = path_view(kind);
 tsgo_buffer parts = {0}, error = {0};
 uint64_t entered = detailed ? now_ns() : 0;
 int status = tsgo_type_parts(id, path, first, last, name, &parts, &error);
 uint64_t returned = detailed ? now_ns() : 0;
 if (detailed) facts_bytes += parts.length;
 free((void *)path.data); free((void *)name.data);
 if (status != TSGO_OK) tsgo_buffer_free(&parts);
 adamic_object *failed = checked(region, status, &error);
 if (failed != NULL) return failed;
 adamic_string *answer = adamic_decode_utf8((const unsigned char *)parts.data, parts.length);
 tsgo_buffer_free(&parts);
 if (detailed) { input_ns += entered - started; call_ns += returned - entered; output_ns += now_ns() - returned; }
 if (timing()) { uint64_t elapsed = now_ns() - started;
  if (query_count == 0) first_query_ns = elapsed;
  queried_ns += elapsed; query_count++; }
 return success(region, (adamic_value){.reference = answer}, true);
}
adamic_object *adamic_tsgo_inspect_in(adamic_region *region, double handle, const adamic_string *file, double start, double end, const adamic_string *kind, const adamic_string *question) {
 bool detailed = profiling();
 uint64_t started = timing() || detailed ? now_ns() : 0;
 if (!integer(handle) || !integer(start) || !integer(end)) return invalid(region, "tsgo: expected a nonnegative safe integer");
 if (!path_valid(file) || !path_valid(kind)) return invalid(region, "tsgo: path contains NUL");
 uint64_t id = (uint64_t)handle, first = (uint64_t)start, last = (uint64_t)end;
 if (!path_valid(question)) return invalid(region, "tsgo: path contains NUL");
 tsgo_view path = path_view(file), name = path_view(kind), query = path_view(question);
 tsgo_buffer facts = {0}, error = {0};
 uint64_t entered = detailed ? now_ns() : 0;
 int status = tsgo_inspect(id, path, first, last, name, query, &facts, &error);
 uint64_t returned = detailed ? now_ns() : 0;
 if (detailed) facts_bytes += facts.length;
 free((void *)path.data); free((void *)name.data); free((void *)query.data);
 if (status != TSGO_OK) tsgo_buffer_free(&facts);
 adamic_object *failed = checked(region, status, &error);
 if (failed != NULL) return failed;
 adamic_string *answer = adamic_decode_utf8((const unsigned char *)facts.data, facts.length);
 tsgo_buffer_free(&facts);
 if (detailed) { input_ns += entered - started; call_ns += returned - entered; output_ns += now_ns() - returned; }
 if (timing()) { uint64_t elapsed = now_ns() - started;
  if (query_count == 0) first_query_ns = elapsed;
  queried_ns += elapsed; query_count++; }
 return success(region, (adamic_value){.reference = answer}, true);
}
adamic_object *adamic_tsgo_release_in(adamic_region *region, double handle) {
 if (!integer(handle)) return invalid(region, "tsgo: expected a nonnegative safe integer");
 uint64_t run_ns = timing() && run_started_ns != 0 ? now_ns() - run_started_ns : 0;
 tsgo_buffer error = {0};
 adamic_object *failed = checked(region, tsgo_release((uint64_t)handle, &error), &error);
 if (failed != NULL) return failed;
 if (profiling()) {
  fprintf(stderr, "tsgo_c_profile: input_ns=%" PRIu64 " call_ns=%" PRIu64 " output_ns=%" PRIu64 " facts_bytes=%" PRIu64 "\n", input_ns, call_ns, output_ns, facts_bytes);
  input_ns = 0; call_ns = 0; output_ns = 0; facts_bytes = 0;
 }
 if (timing()) {
  adamic_output_flush();
  fprintf(stderr, "tsgo: load_ns=%" PRIu64 " query_ns=%" PRIu64 " queries=%zu first_query_ns=%" PRIu64 " run_ns=%" PRIu64 "\n", loaded_ns, queried_ns, query_count, first_query_ns, run_ns);
  loaded_ns = 0; queried_ns = 0; first_query_ns = 0; query_count = 0; run_started_ns = 0;
 }
 static const char *const release_names[] = {"kind"};
 static const bool references[] = {true};
 static const adamic_shape shape = {1, release_names, references, NULL};
 adamic_object *answer = adamic_object_new_in(region, &shape);
 answer->slots[0].reference = &ok_kind;
 return answer;
}
adamic_object *adamic_tsgo_program(const adamic_string *config, const adamic_array *files) {
 return adamic_tsgo_program_in(NULL, config, files);
}
adamic_object *adamic_tsgo_type_parts(double handle, const adamic_string *file, double start, double end, const adamic_string *kind) {
 return adamic_tsgo_type_parts_in(NULL, handle, file, start, end, kind);
}
adamic_object *adamic_tsgo_inspect(double handle, const adamic_string *file, double start, double end, const adamic_string *kind, const adamic_string *question) {
 return adamic_tsgo_inspect_in(NULL, handle, file, start, end, kind, question);
}
adamic_object *adamic_tsgo_release(double handle) {
 return adamic_tsgo_release_in(NULL, handle);
}
#else
/* C11 forbids an empty translation unit under -pedantic. */
typedef int adamic_tsgo_not_linked;
#endif
