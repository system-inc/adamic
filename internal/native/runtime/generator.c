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
            /* The step holds the generator, so a resume that lets go of the last other holder
             * doesn't free the frame it is running on (Memory, q3 on #qk8rztp). */
            adamic_retain(generator);
            generator->status = adamic_generator_status_running;
            generator->code->resume(generator, &yielded);
            adamic_release(generator);
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

bool adamic_generator_is(const void *value) {
    return value != NULL && ((const adamic_heap *)value)->kind == adamic_kind_generator;
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
            /* As a step does: finish's releases can't free the frame it is finishing. */
            adamic_retain(generator);
            adamic_generator_finish(generator);
            adamic_release(generator);
            break;
        case adamic_generator_status_running:
            running(generator, "returned");
        case adamic_generator_status_done:
            break;
    }
}

void adamic_iterator_close(void *value) {
    if (value != NULL && ((const adamic_heap *)value)->kind == adamic_kind_generator) {
        adamic_generator_return(value);
    }
}

/* Only emitted code reaches a hand-driven next(), whose proof listed every origin, so this is its bug. */
static _Noreturn void hand_next_bug(const char *site, const char *what) {
    char message[256];
    int length = snprintf(message, sizeof message, "compiler bug: the hand-driven next() at %s %s", site, what);
    adamic_panic(message, length < 0 ? 0 : (size_t)length < sizeof message ? (size_t)length : sizeof message - 1);
}

/* A step that left no throw, as a fresh result built by the one builder collection_next shares
 * (map_set.c), so a result's shape, slots and representation bytes never differ by origin. The
 * yielded value is the caller's (generator.h) and is moved in, with no retain. The element goes in
 * as collection_next passes its own: a number packed as number | undefined, and at done the absent
 * word for a scalar or NULL for a reference, never the done frame's zero, which is no value. */
static adamic_object *generator_result(const adamic_generator *generator, adamic_field_kind kind, enum adamic_ir_type element_type, adamic_value yielded) {
    bool present = !adamic_generator_done(generator);
    unsigned char tag = kind == adamic_field_reference ? (unsigned char)element_type : kind == adamic_field_boolean && present ? adamic_ir_Boolean : adamic_ir_MaybeNumber;
    if (!present && kind == adamic_field_reference) {
        yielded.reference = NULL;
    } else if (!present) {
        yielded.number = adamic_maybe_number_pack((adamic_maybe_number){false, 0});
    } else if (kind == adamic_field_number) {
        yielded.number = adamic_maybe_number_pack((adamic_maybe_number){true, yielded.number});
    }
    return adamic_iterator_result(kind, present, yielded, tag);
}

adamic_object *adamic_iterator_next(void *iterator, enum adamic_ir_type element_type, const char *site) {
    if (adamic_generator_is(iterator)) {
        /* The yields a result holds, as an allowlist (Adamic and Language on #qk8rztp S2c-2b): any
         * other kind, a maybe-boolean whose byte would be read as a pointer among them, never steps. */
        adamic_field_kind kind = adamic_field_reference;
        if (element_type == adamic_ir_Number) {
            kind = adamic_field_number;
        } else if (element_type == adamic_ir_Boolean) {
            kind = adamic_field_boolean;
        } else if (!adamic_ir_type_reference(element_type)) {
            hand_next_bug(site, "steps a generator whose yields no result holds");
        }
        /* The step retains the frame around its resume (q3), as every step does. */
        adamic_generator *generator = iterator;
        adamic_value yielded = adamic_generator_next(generator);
        if (adamic_thrown != NULL) {
            /* A throw builds nothing: the body has finished, finish released its cells, and the
             * caller's ladder takes the throw out (Memory's rule 3 on #qk8rztp S2c-2b). */
            return NULL;
        }
        return generator_result(generator, kind, element_type, yielded);
    }
    if (adamic_collection_iterator_is(iterator)) {
        /* Its own next, called as a call through the iterator's next field is: the closure held
         * across the call, no arguments, and the result it builds handed back owned. */
        adamic_closure *next = adamic_retain(((adamic_object *)iterator)->slots[0].reference);
        adamic_value result = adamic_closure_call(next, NULL, 0);
        adamic_release(next);
        return result.reference;
    }
    hand_next_bug(site, "steps a value that is neither a generator nor a collection's iterator");
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
