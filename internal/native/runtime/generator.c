#include "generator.h"
#include <stdio.h>

#ifdef ADAMIC_CANONICAL_CLOSURES
/* Only emitted code reaches a running generator, so this is its bug. */
static _Noreturn void running(const adamic_generator *generator, const char *verb) {
    char message[256];
    int length = snprintf(message, sizeof message, "compiler bug: %s %s while running", generator->code->name, verb);
    adamic_panic(message, length < 0 ? 0 : (size_t)length < sizeof message ? (size_t)length : sizeof message - 1);
}

adamic_generator *adamic_generator_new(size_t count, const adamic_generator_code *code) {
    adamic_generator *generator = adamic_allocate(sizeof *generator + count * sizeof(adamic_cell), adamic_kind_generator);
    generator->code = code;
    generator->state = 0;
    generator->status = adamic_generator_status_start;
    generator->error = NULL;
    generator->count = count;
    /* Every cell starts empty and owned by the frame; the initializer writes each one whole. */
    adamic_environment_initialize_cells(&generator->heap, generator->cells, count);
    return generator;
}

adamic_value adamic_generator_next(adamic_generator *generator) {
    adamic_value yielded = {.number = 0};
    switch (generator->status) {
        case adamic_generator_status_start:
        case adamic_generator_status_suspended:
            generator->status = adamic_generator_status_running;
            generator->code->resume(generator, &yielded);
            return yielded;
        case adamic_generator_status_running:
            running(generator, "resumed");
        case adamic_generator_status_done:
            break;
    }
    return yielded;
}

bool adamic_generator_done(const adamic_generator *generator) {
    return generator->status == adamic_generator_status_done;
}

void adamic_generator_kill(adamic_generator *generator, size_t index) {
    adamic_cell *cell = &generator->cells[index];
    if (cell->references && cell->value.reference != NULL) {
        void *held = cell->value.reference;
        cell->value.reference = NULL;
        adamic_release(held);
    }
}

void adamic_generator_finish(adamic_generator *generator) {
    for (size_t index = 0; index < generator->count; index++) {
        adamic_cell *cell = &generator->cells[index];
        if (cell->references && cell->value.reference != NULL) {
            void *owned = cell->value.reference;
            cell->value.reference = NULL;
            adamic_release(owned);
        }
    }
    if (generator->error != NULL) {
        adamic_object *error = generator->error;
        generator->error = NULL;
        adamic_release(error);
    }
    generator->status = adamic_generator_status_done;
}

void adamic_generator_return(adamic_generator *generator) {
    switch (generator->status) {
        case adamic_generator_status_start:
        case adamic_generator_status_suspended:
            adamic_generator_finish(generator);
            break;
        case adamic_generator_status_running:
            running(generator, "returned");
        case adamic_generator_status_done:
            break;
    }
}

/* One walk at every state: a dead cell is already NULL, a borrowed one is
 * skipped, and after finish nothing is left to drop. */
void adamic_generator_free_children(void *value, void (*drop)(void *)) {
    adamic_generator *generator = value;
    if (generator->status == adamic_generator_status_running) { running(generator, "dropped"); }
    for (size_t index = 0; index < generator->count; index++) {
        if (generator->cells[index].references && generator->cells[index].value.reference != NULL) {
            drop(generator->cells[index].value.reference);
        }
    }
    if (generator->error != NULL) { drop(generator->error); }
}
#else
/* C11 forbids an empty translation unit under -pedantic. */
typedef int adamic_generator_not_linked;
#endif
