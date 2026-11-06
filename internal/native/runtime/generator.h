#ifndef ADAMIC_GENERATOR_H
#define ADAMIC_GENERATOR_H
#include "adamic.h"

/* Cells carry their owner only under canonical closures, and every program
 * with a suspending function turns them on. */
#ifdef ADAMIC_CANONICAL_CLOSURES
/* A generator runs on one counted frame, resumed by next(). Every local and
 * parameter lives in a cell. A cell with references true owns its value or
 * holds NULL, never a stale pointer, so the drop walk needs no state. The
 * emitted code sets each cell's references after _new; the runtime only
 * reads it. In this first slice no slot outside the consuming loop holds a
 * generator, so no reference cycle through a generator's cells can be built
 * yet; once generators become values, cycles.go must see its cells as slots.
 * A nested generator also owns self, the closure it was called through, whose
 * cells are its parent's: the frame points out at them, never they at it. */
typedef struct adamic_generator adamic_generator;
typedef struct adamic_generator_code {
    const char *name; /* for compiler-bug panics */
    /* Runs from state to the next way out and sets status on it. */
    void (*resume)(adamic_generator *, adamic_value *yielded);
} adamic_generator_code;
enum adamic_generator_status {
    adamic_generator_status_start,
    adamic_generator_status_suspended,
    adamic_generator_status_running,
    adamic_generator_status_done,
};
struct adamic_generator {
    adamic_heap heap;
    const adamic_generator_code *code;
    unsigned state; /* where resume picks up; 0 is the entry */
    unsigned char status;
    adamic_object *error; /* a throw on its way out; NULL at every suspension */
    /* A nested generator's carrier, retained by its entry and let go of at
     * finish or the drop walk; NULL for a top-level one. */
    adamic_closure *self;
    size_t count;
    adamic_cell cells[];
};
/* Never runs the body: the first next() does. */
adamic_generator *adamic_generator_new(size_t count, const adamic_generator_code *code);
/* The yielded value is owned by the caller. Completion and a throw yield
 * zero, a throw with adamic_thrown set; a done generator yields zero. */
adamic_value adamic_generator_next(adamic_generator *);
bool adamic_generator_done(const adamic_generator *);
/* An iterable view's dispatch (ir.IsGenerator): false for NULL, as adamic_collection_iterator_is
 * is, so a NULL the proof missed reaches the dispatch's loud default, which names the site, never
 * a fault with none (Memory's should on #qk8rztp S2c-1 commit 3). */
bool adamic_generator_is(const void *);
/* At a yield, for a cell the resume never reads again. */
void adamic_generator_kill(adamic_generator *, size_t index);
/* Completion and throw: every counted cell, error and self released and NULLed. */
void adamic_generator_finish(adamic_generator *);
void adamic_generator_return(adamic_generator *);
/* An iterable view's for-of left (ir.Block's Closes): a generator it held is returned, so a later
 * pass finds it done; any other iterator is left as it is. */
void adamic_iterator_close(void *);
void adamic_generator_free_children(void *, void (*)(void *));
#endif
#endif
