// process.c: checked Node process bindings over host_runtime.h.
#define _POSIX_C_SOURCE 200809L
#include "adamic.h"
#include "host_runtime.h"

adamic_string *adamic_process_environment(const adamic_string *name) {
    return adamic_host_environment(name);
}
