#include "generator.h"
#include <assert.h>
#include <stdio.h>
#include <string.h>

/* Each resume below is what the emitter will write: one switch on state, a
 * cell per local, and every way out setting status. Values are built at run
 * time, since a literal is immortal and would hide a missing or double free. */
static void yield(adamic_generator *generator, adamic_value *yielded, void *value, unsigned state) {
    yielded->reference = value;
    generator->state = state;
    generator->status = adamic_generator_status_suspended;
}
static void print(adamic_value value) {
    adamic_string *text = value.reference;
    printf("%.*s\n", (int)text->length, text->bytes);
    adamic_release(text);
}
static void assert_cleared(const adamic_generator *generator) {
    for (size_t index = 0; index < generator->count; index++) {
        assert(generator->cells[index].value.reference == NULL);
    }
}

/* cells[0] is a parameter, cells[1] a local built at the entry and read only
 * after two more yields. */
static void counting_resume(adamic_generator *generator, adamic_value *yielded) {
    switch (generator->state) {
        case 0:
            generator->cells[1].value.reference = adamic_string_from_number(2);
            yield(generator, yielded, adamic_retain(generator->cells[0].value.reference), 1);
            return;
        case 1:
            yield(generator, yielded, adamic_string_from_number(3), 2);
            return;
        case 2:
            yield(generator, yielded, adamic_retain(generator->cells[1].value.reference), 3);
            return;
        case 3:
            adamic_generator_finish(generator);
            return;
    }
    assert(!"unknown state");
}
static const adamic_generator_code counting = {"counting", counting_resume};
static adamic_generator *counting_new(void) {
    adamic_generator *generator = adamic_generator_new(2, &counting);
    for (size_t index = 0; index < 2; index++) {
        assert(generator->cells[index].owner == &generator->heap);
        assert(!generator->cells[index].references && generator->cells[index].value.reference == NULL);
    }
    assert(generator->state == 0 && generator->status == adamic_generator_status_start && generator->error == NULL && generator->self == NULL);
    generator->cells[0].references = generator->cells[1].references = true;
    generator->cells[0].value.reference = adamic_string_from_number(1);
    return generator;
}

/* A nested generator's carrier, what the emitted entry retains into self:
 * a closure whose code nothing here calls. Its maker lets go of it at once,
 * as a temporary carrier's statement does before the first step, so the
 * frame's is the only reference. */
static adamic_value carrier_code(adamic_closure *self, adamic_value *arguments) {
    (void)self;
    (void)arguments;
    return (adamic_value){.number = 0};
}
static adamic_generator *carried_new(void) {
    adamic_generator *generator = counting_new();
    adamic_closure *carrier = adamic_closure_new(carrier_code, 0);
    generator->self = adamic_retain(carrier);
    adamic_release(carrier);
    return generator;
}

/* The parameter is dead after the first yield, so that yield kills it. */
static void killing_resume(adamic_generator *generator, adamic_value *yielded) {
    switch (generator->state) {
        case 0:
            generator->cells[1].value.reference = adamic_string_from_number(2);
            yield(generator, yielded, adamic_retain(generator->cells[0].value.reference), 1);
            adamic_generator_kill(generator, 0);
            return;
        case 1:
            yield(generator, yielded, adamic_retain(generator->cells[1].value.reference), 2);
            return;
        case 2:
            adamic_generator_finish(generator);
            return;
    }
    assert(!"unknown state");
}
static const adamic_generator_code killing = {"killing", killing_resume};

/* cells[1] borrows what cells[0] owns. */
static void aliasing_resume(adamic_generator *generator, adamic_value *yielded) {
    switch (generator->state) {
        case 0:
            generator->cells[1].value.reference = generator->cells[0].value.reference;
            yield(generator, yielded, adamic_retain(generator->cells[1].value.reference), 1);
            return;
        case 1:
            adamic_generator_finish(generator);
            return;
    }
    assert(!"unknown state");
}
static const adamic_generator_code aliasing = {"aliasing", aliasing_resume};

/* A throw routed through error, then moved out the way the emitter's throw
 * path will: to adamic_thrown, then finish. */
static void throwing_resume(adamic_generator *generator, adamic_value *yielded) {
    switch (generator->state) {
        case 0:
            generator->cells[0].value.reference = adamic_string_from_number(5);
            yield(generator, yielded, adamic_retain(generator->cells[0].value.reference), 1);
            return;
        case 1: {
            adamic_string *message = adamic_string_from_number(6);
            generator->error = adamic_error_new(message);
            adamic_release(message);
            adamic_thrown = generator->error;
            generator->error = NULL;
            adamic_generator_finish(generator);
            return;
        }
    }
    assert(!"unknown state");
}
static const adamic_generator_code throwing = {"throwing", throwing_resume};

/* A throw that leaves error set when finish runs: finish still owns it. */
static void abandoning_resume(adamic_generator *generator, adamic_value *yielded) {
    (void)yielded;
    adamic_string *message = adamic_string_from_number(8);
    generator->error = adamic_error_new(message);
    adamic_release(message);
    adamic_generator_finish(generator);
}
static const adamic_generator_code abandoning = {"abandoning", abandoning_resume};

static void reentering_resume(adamic_generator *generator, adamic_value *yielded) {
    (void)yielded;
    (void)adamic_generator_next(generator);
    assert(!"a running generator resumed");
}
static const adamic_generator_code reentering = {"reentering", reentering_resume};

/* return() called on a generator while its own body runs: the panic names the verb called. */
static void returning_resume(adamic_generator *generator, adamic_value *yielded) {
    (void)yielded;
    adamic_generator_return(generator);
    assert(!"a running generator returned");
}
static const adamic_generator_code returning = {"returning", returning_resume};

int main(int argc, char **argv) {
    assert(argc == 2);
    const char *mode = argv[1];
    if (strcmp(mode, "complete") == 0) {
        adamic_generator *generator = counting_new();
        for (;;) {
            adamic_value value = adamic_generator_next(generator);
            if (adamic_generator_done(generator)) {
                assert(value.reference == NULL);
                break;
            }
            print(value);
        }
        /* Cells left uncleared are left for the drop to release twice, under ASan. */
        assert(adamic_generator_next(generator).reference == NULL);
        adamic_release(generator);
    } else if (strcmp(mode, "drop_start") == 0) {
        adamic_release(counting_new());
    } else if (strcmp(mode, "self_finish") == 0) {
        adamic_generator *generator = carried_new();
        for (;;) {
            adamic_value value = adamic_generator_next(generator);
            if (adamic_generator_done(generator)) {
                break;
            }
            print(value);
        }
        /* A done generator no longer pins its parent's cells. */
        assert(generator->self == NULL);
        adamic_release(generator);
    } else if (strcmp(mode, "self_drop_start") == 0) {
        adamic_release(carried_new());
    } else if (strcmp(mode, "drop_suspended") == 0) {
        adamic_generator *generator = counting_new();
        print(adamic_generator_next(generator));
        assert(generator->status == adamic_generator_status_suspended && generator->cells[1].value.reference != NULL);
        adamic_release(generator);
    } else if (strcmp(mode, "return") == 0) {
        adamic_generator *generator = counting_new();
        print(adamic_generator_next(generator));
        adamic_generator_return(generator);
        assert(adamic_generator_done(generator));
        assert_cleared(generator);
        adamic_generator_return(generator);
        assert(adamic_generator_next(generator).reference == NULL);
        adamic_release(generator);
    } else if (strcmp(mode, "kill") == 0) {
        adamic_generator *generator = adamic_generator_new(2, &killing);
        generator->cells[0].references = generator->cells[1].references = true;
        generator->cells[0].value.reference = adamic_string_from_number(1);
        /* The drop walk meets the killed cell, which must read NULL. */
        print(adamic_generator_next(generator));
        adamic_release(generator);
    } else if (strcmp(mode, "borrowed") == 0) {
        adamic_string *shared = adamic_string_from_number(7);
        adamic_generator *generator = adamic_generator_new(2, &aliasing);
        generator->cells[0].references = true;
        generator->cells[0].value.reference = adamic_retain(shared);
        print(adamic_generator_next(generator));
        adamic_release(generator);
        print((adamic_value){.reference = shared});
    } else if (strcmp(mode, "throw") == 0) {
        adamic_generator *generator = adamic_generator_new(1, &throwing);
        generator->cells[0].references = true;
        print(adamic_generator_next(generator));
        assert(adamic_thrown == NULL);
        adamic_value value = adamic_generator_next(generator);
        assert(value.reference == NULL && adamic_generator_done(generator));
        assert(adamic_thrown != NULL && generator->error == NULL);
        assert_cleared(generator);
        adamic_object *thrown = adamic_thrown;
        adamic_thrown = NULL;
        adamic_string *message = adamic_error_to_string(thrown);
        print((adamic_value){.reference = message});
        adamic_release(thrown);
        adamic_release(generator);
    } else if (strcmp(mode, "finish_error") == 0) {
        adamic_generator *generator = adamic_generator_new(0, &abandoning);
        adamic_value value = adamic_generator_next(generator);
        assert(value.reference == NULL && adamic_generator_done(generator) && generator->error == NULL);
        adamic_release(generator);
    } else if (strcmp(mode, "drop_error") == 0) {
        /* error is NULL at every suspension the emitter makes; the walk still drops one it finds. */
        adamic_generator *generator = counting_new();
        print(adamic_generator_next(generator));
        adamic_string *message = adamic_string_from_number(9);
        generator->error = adamic_error_new(message);
        adamic_release(message);
        adamic_release(generator);
    } else if (strcmp(mode, "reentrant") == 0) {
        adamic_generator *generator = adamic_generator_new(0, &reentering);
        (void)adamic_generator_next(generator);
        assert(!"reentrant next returned");
    } else if (strcmp(mode, "returning") == 0) {
        adamic_generator *generator = adamic_generator_new(0, &returning);
        (void)adamic_generator_next(generator);
        assert(!"returning next returned");
    } else if (strcmp(mode, "is") == 0) {
        /* A view's dispatch asks this first; NULL, a string and a generator each get their answer. */
        adamic_generator *generator = counting_new();
        adamic_string *text = adamic_string_from_number(4);
        printf("%d %d %d\n", adamic_generator_is(NULL), adamic_generator_is(text), adamic_generator_is(generator));
        adamic_release(text);
        adamic_release(generator);
    } else {
        assert(!"unknown mode");
    }
    return 0;
}
