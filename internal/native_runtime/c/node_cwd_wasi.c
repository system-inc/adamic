#define _POSIX_C_SOURCE 200809L
#include "adamic.h"
#include "node_cwd_wasi.h"
#ifdef ADAMIC_TARGET_WASI
#include <stdlib.h>
#include <unistd.h>

// Preview 1 supplies no initial cwd. This function is reachable only through
// cwd-dependent adapters, so file-free modules retain their existing imports.
static bool initialized;
void adamic_node_wasi_cwd(void) {
    if (initialized) { return; }
    const char *directory = getenv("ADAMIC_WASI_CWD");
    if (directory != NULL && chdir(directory) != 0) {
        adamic_panic("wasm32-wasi: cannot enter supplied working directory", sizeof "wasm32-wasi: cannot enter supplied working directory" - 1);
    }
    initialized = true;
}
#endif
