#include "async.h"
#include <assert.h>
#include <string.h>

typedef struct {
    adamic_async_frame base;
    adamic_cell cells[2];
} frame;
static unsigned cleaned, dropped;
static void children(adamic_async_frame *base, void (*drop)(void *)) {
    adamic_environment_drop_cells(((frame *)base)->cells, 2, drop);
}
static void resume(adamic_async_frame *base, adamic_value value, bool rejected) {
    (void)base; (void)value; (void)rejected;
    assert(!"abandoned frame resumed");
}
static void cleanup(adamic_async_frame *base) {
    frame *owned = (frame *)base;
    void *held = owned->cells[0].value.reference;
    owned->cells[0].value.reference = NULL;
    adamic_release(held);
    cleaned++;
}
static adamic_value code(adamic_closure *self, adamic_value *arguments) {
    (void)self; (void)arguments;
    return (adamic_value){.number = 0};
}
static frame *make_frame(void) {
    frame *owned = adamic_allocate(sizeof *owned, adamic_kind_async_frame);
    owned->base.state = 0;
    owned->base.output = owned->base.waiting = NULL;
    owned->base.resume = resume;
    owned->base.children = children;
    memset(owned->cells, 0xff, sizeof owned->cells);
    adamic_environment_initialize_cells(&owned->base.heap, owned->cells, 2);
    for (size_t i = 0; i < 2; i++) {
        assert(owned->cells[i].owner == &owned->base.heap);
        assert(!owned->cells[i].ready && !owned->cells[i].references);
        assert(owned->cells[i].value.reference == NULL);
    }
    return owned;
}
static void ignore_drop(void *value) { (void)value; }
static void count_drop(void *value) {
    assert(value != NULL);
    dropped++;
}
int main(int argc, char **argv) {
    assert(argc == 2);
    if (strcmp(argv[1], "environment") == 0) {
        frame *owned = make_frame();
        adamic_cell *cell = adamic_retain(&owned->cells[1]);
        assert(adamic_reference_count(&owned->base.heap) == 2);
        adamic_release(owned);
        assert(adamic_reference_count(cell->owner) == 1);
        cell->references = true;
        cell->value.reference = adamic_string_from_number(12345);
        adamic_release(cell);
        adamic_environment *environment = adamic_environment_new(2);
        environment->cells[0].references = true;
        environment->cells[0].value.reference = adamic_string_from_number(67890);
        environment->cells[1].value.number = 3;
        adamic_environment_drop_cells(environment->cells, 2, count_drop);
        assert(dropped == 1);
        adamic_share(&environment->cells[0]);
        assert(adamic_is_shared(&environment->heap));
        assert(adamic_is_shared(environment->cells[0].value.reference));
        adamic_release(environment);
    } else if (strcmp(argv[1], "share") == 0) {
        frame *owned = make_frame();
        adamic_share(&owned->cells[0]);
        assert(!"shared an async frame through an interior cell");
    } else {
        frame *owned = make_frame();
        adamic_async_promise *promise = adamic_async_new();
        bool free_children = strcmp(argv[1], "free") == 0;
        bool forget = free_children || strcmp(argv[1], "forget") == 0;
        if (!forget) {
            adamic_closure *closure = adamic_closure_new(code, 1);
            closure->cells[0] = adamic_retain(&owned->cells[1]);
            owned->cells[0].references = true;
            owned->cells[0].value.reference = closure;
        }
        owned->cells[1].references = true;
        owned->cells[1].value.reference = adamic_string_from_number(24680);
        adamic_async_register_cleanup(&owned->base, cleanup);
        if (free_children) {
            adamic_async_free_children(&owned->base, ignore_drop);
        } else if (forget) {
            adamic_async_forget_cleanup(&owned->base);
        }
        adamic_async_await(&owned->base, promise);
        adamic_release(owned);
        if (strcmp(argv[1], "exit") == 0) {
            adamic_release(promise);
            adamic_async_teardown();
        } else {
            adamic_async_cancel(promise);
            adamic_release(promise);
            adamic_async_teardown();
        }
        assert(cleaned == (forget ? 0u : 1u));
    }
    return 0;
}
