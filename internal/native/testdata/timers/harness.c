#define _POSIX_C_SOURCE 200809L
#include "timers.h"
/* timers_test.go defines TIMERS_GRAPH first when the runtime has graph regions (slice 6). */
#ifdef TIMERS_GRAPH
#include "graph_regions.h"
#endif
#include <assert.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

static const char *mode;
static unsigned fired;
static adamic_timer_handle interval;
static adamic_host_request *request;
static void empty_children(adamic_async_frame *frame, void (*drop)(void *)) { (void)frame; (void)drop; }
static void resume(adamic_async_frame *frame, adamic_value value, bool rejected) {
    (void)value; assert(!rejected);
    puts(frame->state == 0 ? "job" : "first-job");
}
static void job(unsigned state) {
    adamic_async_promise *promise = adamic_async_new();
    adamic_async_settle(promise, (adamic_value){.number = 0}, false, false);
    adamic_async_frame *frame = adamic_allocate(sizeof *frame, adamic_kind_async_frame);
    *frame = (adamic_async_frame){.heap = frame->heap, .state = state,
        .children = empty_children, .resume = resume};
    adamic_async_await(frame, promise);
    adamic_release(frame); adamic_release(promise);
}
static adamic_value callback(adamic_closure *self, adamic_value *args) {
    fired++;
    adamic_string *label = strcmp(mode, "graph") == 0 ?
        ((adamic_object *)self->cells[0]->value.reference)->slots[1].reference : self->cells[0]->value.reference;
    if (strcmp(mode, "order") == 0) {
        puts(fired == 1 ? "first" : "second");
        if (fired == 1) job(1);
    } else if (strcmp(mode, "interval") == 0 || strcmp(mode, "graph") == 0) {
        adamic_string *text = args[0].reference;
        printf("%.*s %u %.*s %.0f\n", (int)label->length, label->bytes, fired,
            (int)text->length, text->bytes, args[1].number);
        if (fired == 3) assert(adamic_timer_cancel(interval));
    } else if (strcmp(mode, "cancel-self") == 0) {
        puts("tick");
        assert(adamic_timer_cancel(interval));
        /* The invocation continues using its captures and arguments after cancel. */
        adamic_string *text = args[0].reference;
        printf("after %.*s %.*s\n", (int)label->length, label->bytes, (int)text->length, text->bytes);
    } else if (strcmp(mode, "unref-live") == 0 || strcmp(mode, "deadlines") == 0) {
        puts(args[1].number == 1 ? "early" : "late");
    } else {
        if (strcmp(mode, "host") == 0) assert(adamic_host_resolve(request, "", 0, 0));
        printf("fire %.*s\n", (int)label->length, label->bytes);
    }
    return (adamic_value){.number = 0};
}
/* The library supplies a signature-checked thunk, not an unchecked code cast. */
static void invoke(adamic_closure *closure, size_t argc, const adamic_value *args) {
    assert(argc == 2); /* Required two actual arguments for this fixture signature. */
    adamic_value arguments[] = {args[0], args[1]};
    (void)closure->code(closure, arguments);
}
static adamic_timer_handle start(double delay, bool repeat, double number, bool graph) {
    adamic_string *label = adamic_string_from_number(12345);
    void *capture = label;
#ifndef TIMERS_GRAPH
    /* Graph regions arrive with runtime's slice 6; the test skips this mode until then. */
    assert(!graph);
#else
    if (graph) {
        static const char *const names[] = {"next", "label"};
        static const bool references[] = {true, true};
        static const adamic_shape shape = {2, names, references, NULL};
        adamic_object *node = adamic_object_new(&shape);
        node = adamic_graph_adopt(node, sizeof *node + 2 * sizeof(adamic_value));
        node->slots[0].reference = adamic_graph_hold(node, node);
        node->slots[1].reference = label;
        capture = node;
    }
#endif
    adamic_cell *cell = adamic_cell_new((adamic_value){.reference = capture}, true);
    adamic_closure *closure = adamic_closure_new(callback, 1);
#ifdef TIMERS_GRAPH
    if (graph) {
        /* A graph closure/cell component; the registration is its last outside owner. */
        closure = adamic_graph_adopt(closure, sizeof *closure + sizeof closure->cells[0]);
        cell = adamic_graph_adopt_owned(cell, sizeof *cell);
        closure->cells[0] = adamic_graph_hold(closure, cell);
        adamic_release(cell);
    } else
#endif
    closure->cells[0] = cell;
    adamic_string *text = adamic_string_from_number(67890);
    adamic_value arguments[] = {{.reference = text}, {.number = number}};
    bool references[] = {true, false};
    adamic_timer_handle handle = adamic_timer_start(adamic_timers_provider(), closure, invoke,
        delay, repeat, 2, arguments, references);
    adamic_release(closure); adamic_release(text);
    /* Stack arrays and their original owners disappear before delivery. */
    memset(arguments, 0, sizeof arguments);
    return handle;
}
static void invoke_optional(adamic_closure *closure, size_t argc, const adamic_value *args) {
    /* Source signature (n = 42, ...rest: number[]): optional missing n gets its
     * default and the checked thunk consumes the actual rest count. */
    assert(argc == 0 || argc == 1 || argc == 3);
    adamic_string *label = closure->cells[0]->value.reference;
    double rest = 0;
    for (size_t index = 1; index < argc; index++) rest += args[index].number;
    printf("%.*s %.0f %zu %.0f\n", (int)label->length, label->bytes,
        argc == 0 ? 42 : args[0].number, argc == 0 ? 0 : argc - 1, rest);
}
static void arity(void) {
    adamic_closure *closure = adamic_closure_new(callback, 1);
    closure->cells[0] = adamic_cell_new((adamic_value){.reference = adamic_string_from_number(12345)}, true);
    adamic_value arguments[] = {{.number = 7}, {.number = 8}, {.number = 9}};
    bool references[] = {false, false, false};
    (void)adamic_timer_start(adamic_timers_provider(), closure, invoke_optional, 1, false, 0, NULL, NULL);
    (void)adamic_timer_start(adamic_timers_provider(), closure, invoke_optional, 1, false, 1, arguments, references);
    (void)adamic_timer_start(adamic_timers_provider(), closure, invoke_optional, 1, false, 3, arguments, references);
    adamic_release(closure);
}
int main(int argc, char **argv) {
    assert(argc == 2); mode = argv[1];
    puts("prefix");
    if (strcmp(mode, "arity") == 0) {
        arity();
    } else if (strcmp(mode, "host") == 0) {
        adamic_promise *promise;
        request = adamic_host_promise_new(&promise);
        adamic_release(promise);
        interval = start(1, false, 7, false);
        assert(adamic_timer_unref(interval));
    } else if (strcmp(mode, "order") == 0) {
        job(0); (void)start(1, false, 7, false); (void)start(1, false, 7, false);
    } else if (strcmp(mode, "interval") == 0 || strcmp(mode, "graph") == 0 || strcmp(mode, "cancel-self") == 0) {
        interval = start(1, true, 7, strcmp(mode, "graph") == 0);
    } else if (strcmp(mode, "cancel") == 0) {
        interval = start(1, false, 7, false);
        assert(adamic_timer_cancel(interval));
        assert(!adamic_timer_cancel(interval));
        assert(!adamic_timer_unref(interval));
    } else if (strcmp(mode, "foreign") == 0) {
        interval = start(1, false, 7, false);
        adamic_timer_handle foreign = {NULL, interval.identity};
        assert(!adamic_timer_cancel(foreign));
        assert(!adamic_timer_unref(foreign));
        assert(!adamic_timer_cancel((adamic_timer_handle){adamic_timers_provider(), UINTPTR_MAX}));
    } else if (strcmp(mode, "unref") == 0) {
        interval = start(1, false, 7, false);
        assert(adamic_timer_unref(interval));
    } else if (strcmp(mode, "unref-live") == 0) {
        interval = start(1, false, 1, false);
        assert(adamic_timer_unref(interval));
        (void)start(20, false, 2, false);
    } else if (strcmp(mode, "deadlines") == 0) {
        (void)start(20, false, 2, false);
        (void)start(0.9, false, 1, false);
    } else abort();
    adamic_async_run();
    assert(!adamic_timer_cancel(interval)); /* Also stale after the final interval fire. */
    return 0;
}
