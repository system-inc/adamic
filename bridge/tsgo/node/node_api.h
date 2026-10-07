/* Node-API ABI declarations copied from Node v24.19.0 src/js_native_api*.h.
 * Copyright Node.js contributors. MIT license: see NODE_API_LICENSE.
 * Only the stable C declarations used by node.c are included. */
#ifndef ADAMIC_NODE_API_H
#define ADAMIC_NODE_API_H
#include <stddef.h>
#include <stdint.h>
typedef struct napi_env__* napi_env;
typedef struct napi_value__* napi_value;
typedef struct napi_callback_info__* napi_callback_info;
typedef enum {
  napi_ok,
  napi_invalid_arg,
  napi_object_expected,
  napi_string_expected,
  napi_name_expected,
  napi_function_expected,
  napi_number_expected,
  napi_boolean_expected,
  napi_array_expected,
  napi_generic_failure,
  napi_pending_exception,
  napi_cancelled,
  napi_escape_called_twice,
  napi_handle_scope_mismatch,
  napi_callback_scope_mismatch,
  napi_queue_full,
  napi_closing,
  napi_bigint_expected,
  napi_date_expected,
  napi_arraybuffer_expected,
  napi_detachable_arraybuffer_expected,
  napi_would_deadlock,  // unused
  napi_no_external_buffers_allowed,
  napi_cannot_run_js,
} napi_status;
typedef napi_value (*napi_callback)(napi_env env, napi_callback_info info);
napi_status napi_create_object(napi_env env,
                                                      napi_value* result);
napi_status napi_create_double(napi_env env,
                                                      double value,
                                                      napi_value* result);
napi_status napi_create_string_utf8(napi_env env,
                                                           const char* str,
                                                           size_t length,
                                                           napi_value* result);
napi_status napi_create_function(napi_env env,
                                                        const char* utf8name,
                                                        size_t length,
                                                        napi_callback cb,
                                                        void* data,
                                                        napi_value* result);
napi_status napi_get_value_double(napi_env env,
                                                         napi_value value,
                                                         double* result);
napi_status napi_get_value_string_utf8(
    napi_env env, napi_value value, char* buf, size_t bufsize, size_t* result);
napi_status napi_set_named_property(napi_env env,
                                                           napi_value object,
                                                           const char* utf8name,
                                                           napi_value value);
napi_status napi_get_element(napi_env env,
                                                    napi_value object,
                                                    uint32_t index,
                                                    napi_value* result);
napi_status napi_get_array_length(napi_env env,
                                                         napi_value value,
                                                         uint32_t* result);
napi_status napi_get_cb_info(
    napi_env env,               // [in] Node-API environment handle
    napi_callback_info cbinfo,  // [in] Opaque callback-info handle
    size_t* argc,      // [in-out] Specifies the size of the provided argv array
                       // and receives the actual count of args.
    napi_value* argv,  // [out] Array of values
    napi_value* this_arg,  // [out] Receives the JS 'this' arg for the call
    void** data);
#endif
