// process.c: checked Node process bindings over host_runtime.h.
#define _POSIX_C_SOURCE 200809L
#include "adamic.h"
#include "host_runtime.h"
#include <unistd.h>

adamic_maybe_boolean adamic_process_is_tty(enum adamic_stream stream) {
	bool terminal = isatty(stream) != 0;
	return (adamic_maybe_boolean){.present = terminal, .boolean = terminal};
}

adamic_string *adamic_process_environment(const adamic_string *name) {
    return adamic_host_environment(name);
}
