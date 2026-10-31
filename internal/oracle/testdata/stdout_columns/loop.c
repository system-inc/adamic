/* A pending host request is the runtime's event-loop integration seam. The
 * language's Promise.resolve fixtures only drain microtasks, not a host turn. */
#include "async.h"
#include "host_runtime.h"
#include <assert.h>
#include <pthread.h>
#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <unistd.h>
static void print_width(FILE *destination, const char *label) {
    adamic_maybe_number width = adamic_host_stdout_columns();
    if (width.present) fprintf(destination, "%s %.0f\n", label, width.number);
    else fprintf(destination, "%s undefined\n", label);
    fflush(destination);
}
static void *complete(void *request) {
    char byte;
    assert(read(0, &byte, 1) == 1);
    assert(adamic_host_resolve(request, NULL, 0, 0));
    return NULL;
}
int main(int argc, char **argv) {
    adamic_start(argc, argv);
    FILE *ready = stderr;
    if (argc > 1 && strcmp(argv[1], "failure") == 0) {
        ready = fdopen(atoi(getenv("COLUMNS_READY_FD")), "w");
        assert(ready != NULL);
    }
    print_width(ready, "initial");
    if (ready != stderr) fclose(ready);
    if (argc > 1 && strcmp(argv[1], "sync") == 0) {
        char byte;
        assert(read(0, &byte, 1) == 1);
    } else {
        adamic_promise *promise;
        adamic_host_request *request = adamic_host_promise_new(&promise);
        pthread_t worker;
        assert(pthread_create(&worker, NULL, complete, request) == 0);
        adamic_async_run();
        assert(pthread_join(worker, NULL) == 0);
        adamic_release(promise);
    }
    print_width(stderr, "later");
    return 0;
}
