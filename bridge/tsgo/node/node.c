/* Node-API adapter to the same explicitly linked checker ABI as native Adamic.
 * No checker behavior is reimplemented here. Outputs/errors retain byte lengths. */
#ifndef ADAMIC_TARGET_WASI
#include "node_api.h"
#include "tsgo.h"
#include <math.h>
#include <stdbool.h>
#include <stdlib.h>
#include <string.h>

static napi_value string(napi_env env, const char *bytes, size_t length) {
 napi_value result = NULL;
 if (napi_create_string_utf8(env, bytes == NULL ? "" : bytes, length, &result) != napi_ok) return NULL;
 return result;
}
static napi_value number(napi_env env, double value) {
 napi_value result = NULL;
 if (napi_create_double(env, value, &result) != napi_ok) return NULL;
 return result;
}
static bool property(napi_env env, napi_value object, const char *name, napi_value value) {
 return value != NULL && napi_set_named_property(env, object, name, value) == napi_ok;
}
static napi_value result(napi_env env, bool ok, napi_value value) {
 napi_value object = NULL;
 if (napi_create_object(env, &object) != napi_ok) return NULL;
 if (!property(env, object, "kind", string(env, ok ? "Ok" : "Error", ok ? 2 : 5))) return NULL;
 if (value != NULL && !property(env, object, ok ? "value" : "message", value)) return NULL;
 return object;
}
static napi_value invalid(napi_env env, const char *message) {
 napi_value value = string(env, message, strlen(message));
 return value == NULL ? NULL : result(env, false, value);
}
static napi_value failed(napi_env env, tsgo_buffer *error) {
 napi_value message = error->length != 0 ? string(env, error->data, error->length)
  : string(env, "tsgo: invalid argument or out of memory", sizeof "tsgo: invalid argument or out of memory" - 1);
 tsgo_buffer_free(error);
 return message == NULL ? NULL : result(env, false, message);
}
static bool integer(napi_env env, napi_value value, uint64_t *output) {
 double number;
 if (napi_get_value_double(env, value, &number) != napi_ok ||
     !(number >= 0 && number <= 9007199254740991.0 && trunc(number) == number)) return false;
 *output = (uint64_t)number;
 return true;
}
/* UTF-8 conversion replaces lone surrogates, as native's path_view does. */
static const char *view(napi_env env, napi_value value, tsgo_view *output) {
 size_t length;
 if (napi_get_value_string_utf8(env, value, NULL, 0, &length) != napi_ok)
  return "tsgo: invalid argument or out of memory";
 char *bytes = malloc(length + 1);
 if (bytes == NULL) return "tsgo: invalid argument or out of memory";
 if (napi_get_value_string_utf8(env, value, bytes, length + 1, &length) != napi_ok) {
  free(bytes); return "tsgo: invalid argument or out of memory";
 }
 if (memchr(bytes, 0, length) != NULL) { free(bytes); return "tsgo: path contains NUL"; }
 *output = (tsgo_view){bytes, length};
 return NULL;
}
static int operations[] = {0, 1, 2, 3, 4};
static napi_value call(napi_env env, napi_callback_info info) {
 napi_value args[6]; size_t count = 6; void *data = NULL;
 if (napi_get_cb_info(env, info, &count, args, NULL, &data) != napi_ok || data == NULL) return NULL;
 int operation = *(int *)data;
 size_t expected = operation == 0 ? 2 : operation == 1 ? 3 : operation == 2 ? 6 : operation == 3 ? 5 : 1;
 if (count != expected) return invalid(env, "tsgo: invalid argument or out of memory");
 uint64_t handle = 0, start = 0, end = 0;
 if (operation != 0 && !integer(env, args[0], &handle)) return invalid(env, "tsgo: expected a nonnegative safe integer");
 if (operation >= 1 && operation <= 3 && !integer(env, args[2], &start)) return invalid(env, "tsgo: expected a nonnegative safe integer");
 if ((operation == 2 || operation == 3) && !integer(env, args[3], &end)) return invalid(env, "tsgo: expected a nonnegative safe integer");
 tsgo_view path = {0}, kind = {0}, question = {0};
 tsgo_view *roots = NULL; uint32_t root_count = 0;
 const char *error_text = NULL;
 if (operation != 4) error_text = view(env, args[operation == 0 ? 0 : 1], &path);
 if (error_text == NULL && (operation == 2 || operation == 3)) error_text = view(env, args[4], &kind);
 if (error_text == NULL && operation == 2) error_text = view(env, args[5], &question);
 if (error_text == NULL && operation == 0) {
  if (napi_get_array_length(env, args[1], &root_count) != napi_ok) error_text = "tsgo: invalid argument or out of memory";
  else {
   roots = calloc(root_count == 0 ? 1 : root_count, sizeof *roots);
   if (roots == NULL) error_text = "tsgo: invalid argument or out of memory";
   else for (uint32_t index = 0; index < root_count; index++) {
    napi_value file;
    if (napi_get_element(env, args[1], index, &file) != napi_ok) error_text = "tsgo: invalid argument or out of memory";
    else error_text = view(env, file, &roots[index]);
    if (error_text != NULL) break;
   }
  }
 }
 napi_value answer = NULL;
 if (error_text != NULL) answer = invalid(env, error_text);
 else {
  tsgo_buffer error = {0}, output = {0}; tsgo_result query = {0};
  int status;
  switch (operation) {
   case 0: status = tsgo_create(path, roots, root_count, &handle, &error); break;
   case 1: status = tsgo_query(handle, path, start, &query, &error); break;
   case 2: status = tsgo_inspect(handle, path, start, end, kind, question, &output, &error); break;
   case 3: status = tsgo_type_parts(handle, path, start, end, kind, &output, &error); break;
   default: status = tsgo_release(handle, &error); break;
  }
  if (status != TSGO_OK) answer = failed(env, &error);
  else {
   napi_value value = NULL;
   if (operation == 0) value = number(env, (double)handle);
   else if (operation == 1) {
    if (napi_create_object(env, &value) != napi_ok ||
        !property(env, value, "nodeKind", number(env, query.kind)) ||
        !property(env, value, "symbolName", string(env, query.symbol.data, query.symbol.length)) ||
        !property(env, value, "type", string(env, query.type.data, query.type.length))) value = NULL;
   } else if (operation != 4) value = string(env, output.data, output.length);
   answer = operation != 4 && value == NULL ? NULL : result(env, true, value);
   tsgo_buffer_free(&error);
  }
  tsgo_buffer_free(&output); tsgo_result_free(&query);
 }
 free((void *)path.data); free((void *)kind.data); free((void *)question.data);
 if (roots != NULL) for (uint32_t index = 0; index < root_count; index++) free((void *)roots[index].data);
 free(roots);
 return answer;
}
/* Node-API's stable symbol, discovered by Node's addon loader. */
napi_value napi_register_module_v1(napi_env env, napi_value exports) {
 static const char *const names[] = {"tsgoProgram", "tsgoQuery", "tsgoInspect", "tsgoTypeParts", "tsgoRelease"};
 for (size_t index = 0; index < 5; index++) {
  napi_value function = NULL;
  if (napi_create_function(env, names[index], strlen(names[index]), call, &operations[index], &function) != napi_ok ||
      !property(env, exports, names[index], function)) return NULL;
 }
 return exports;
}
#else
/* No Node or checker bridge in WASI. */
typedef int adamic_tsgo_node_not_supported;
#endif
