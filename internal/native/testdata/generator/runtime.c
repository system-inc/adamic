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

/* Two strings, each built at its yield and owned by nothing but the step that hands it over, so a
 * result that holds one borrowed, or one retained twice, shows under ASan or in the counts. */
static void naming_resume(adamic_generator *generator, adamic_value *yielded) {
    switch (generator->state) {
        case 0:
            yield(generator, yielded, adamic_string_from_number(1), 1);
            return;
        case 1:
            yield(generator, yielded, adamic_string_from_number(2), 2);
            return;
        case 2:
            adamic_generator_finish(generator);
            return;
    }
    assert(!"unknown state");
}
static const adamic_generator_code naming = {"naming", naming_resume};

/* One number, then done. */
static void numbering_resume(adamic_generator *generator, adamic_value *yielded) {
    switch (generator->state) {
        case 0:
            yielded->number = 10;
            generator->state = 1;
            generator->status = adamic_generator_status_suspended;
            return;
        case 1:
            adamic_generator_finish(generator);
            return;
    }
    assert(!"unknown state");
}
static const adamic_generator_code numbering = {"numbering", numbering_resume};

/* One boolean, then done. */
static void flagging_resume(adamic_generator *generator, adamic_value *yielded) {
    switch (generator->state) {
        case 0:
            yielded->boolean = true;
            generator->state = 1;
            generator->status = adamic_generator_status_suspended;
            return;
        case 1:
            adamic_generator_finish(generator);
            return;
    }
    assert(!"unknown state");
}
static const adamic_generator_code flagging = {"flagging", flagging_resume};

/* A hand-driven next()'s result as the emitter will read it, by its slots after #7nw4kgj: value in
 * slot 0, done in slot 1. A string value is printed borrowed, never released here. */
static bool result_done(const adamic_object *result) {
    return result->slots[1].boolean;
}
static void show(const adamic_object *result) {
    const adamic_string *text = result->slots[0].reference;
    printf("%.*s\n", (int)text->length, text->bytes);
}
/* Object.keys(result).join(','), through the runtime's own Object.keys. */
static void show_keys(const adamic_object *result) {
    adamic_array *keys = adamic_object_keys(result);
    for (size_t index = 0; index < keys->length; index++) {
        const adamic_string *key = keys->elements[index].reference;
        printf("%s%.*s", index == 0 ? "" : ",", (int)key->length, key->bytes);
    }
    printf("\n");
    adamic_release(keys);
}
static const char hand_site[] = "harness.ts:3:9";

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
    } else if (strcmp(mode, "next_strings") == 0) {
        /* P1's twin: checker.ts:21636's loop over a string generator, run to done. Each result owns
         * the value it was handed, so releasing it frees the string once. */
        adamic_generator *generator = adamic_generator_new(0, &naming);
        for (;;) {
            adamic_object *status = adamic_iterator_next(generator, adamic_ir_String, hand_site);
            if (result_done(status)) {
                /* The last status is done, and holds no value. */
                assert(status->slots[0].reference == NULL);
                adamic_release(status);
                break;
            }
            show(status);
            adamic_release(status);
        }
        adamic_object *again = adamic_iterator_next(generator, adamic_ir_String, hand_site);
        assert(result_done(again) && again->slots[0].reference == NULL);
        adamic_release(again);
        adamic_release(generator);
    } else if (strcmp(mode, "next_drop") == 0) {
        /* P2's twin: one step, its value read, then the result and the suspended frame let go. */
        adamic_generator *generator = adamic_generator_new(0, &naming);
        adamic_object *status = adamic_iterator_next(generator, adamic_ir_String, hand_site);
        assert(!result_done(status));
        show(status);
        adamic_release(status);
        adamic_release(generator);
    } else if (strcmp(mode, "next_throw") == 0) {
        /* P3's twin: the body throws on the second step, which builds nothing, and next() after it
         * answers done. */
        adamic_generator *generator = adamic_generator_new(1, &throwing);
        generator->cells[0].references = true;
        adamic_object *status = adamic_iterator_next(generator, adamic_ir_String, hand_site);
        assert(adamic_thrown == NULL && !result_done(status));
        show(status);
        adamic_release(status);
        assert(adamic_iterator_next(generator, adamic_ir_String, hand_site) == NULL);
        assert(adamic_thrown != NULL && adamic_generator_done(generator));
        assert_cleared(generator);
        adamic_object *thrown = adamic_thrown;
        adamic_thrown = NULL;
        adamic_object *after = adamic_iterator_next(generator, adamic_ir_String, hand_site);
        assert(adamic_thrown == NULL && result_done(after) && after->slots[0].reference == NULL);
        adamic_release(after);
        print((adamic_value){.reference = adamic_error_to_string(thrown)});
        adamic_release(thrown);
        adamic_release(generator);
    } else if (strcmp(mode, "next_done") == 0) {
        /* P4: a number generator stepped to done, then once more. Its value reads as number |
         * undefined: present with the number while it yields, absent once done, never the done
         * frame's zero. */
        adamic_generator *generator = adamic_generator_new(0, &numbering);
        adamic_object *status = adamic_iterator_next(generator, adamic_ir_Number, hand_site);
        adamic_maybe_number value = adamic_maybe_number_unpack(status->slots[0].number);
        assert(!result_done(status) && value.present);
        printf("%d %g\n", value.present, value.number);
        adamic_release(status);
        for (int step = 0; step < 2; step++) {
            adamic_object *finished = adamic_iterator_next(generator, adamic_ir_Number, hand_site);
            value = adamic_maybe_number_unpack(finished->slots[0].number);
            printf("%d\n", value.present);
            assert(result_done(finished) && !value.present && adamic_object_field_types(finished)[0] == adamic_ir_MaybeNumber);
            if (step == 1) {
                show_keys(finished);
            }
            adamic_release(finished);
        }
        adamic_release(generator);
    } else if (strcmp(mode, "next_done_boolean") == 0) {
        /* P4's boolean row: a boolean generator stepped to done, then once more, its value read as
         * the emitter reads boolean | undefined, by the slot's maybe-boolean byte, with its
         * representation byte beside it. Present, it is the boolean under Boolean; done, it is
         * undefined under MaybeBoolean, which the builder writes (#ak92fcp), never the absent
         * maybe-number word, whose low byte reads true. */
        adamic_generator *generator = adamic_generator_new(0, &flagging);
        for (int step = 0; step < 3; step++) {
            adamic_object *status = adamic_iterator_next(generator, adamic_ir_Boolean, hand_site);
            adamic_maybe_boolean value = adamic_maybe_boolean_unpack(status->slots[0].maybe_boolean);
            assert(result_done(status) == (step != 0));
            printf("%s %d\n", !value.present ? "undefined" : value.boolean ? "true" : "false", adamic_object_field_types(status)[0]);
            adamic_release(status);
        }
        adamic_release(generator);
    } else if (strcmp(mode, "next_collection") == 0) {
        /* A Map's values() through the same entry: its own next builds each result. */
        adamic_map *map = adamic_map_new(false, false);
        adamic_map_set(map, (adamic_value){.number = 1}, (adamic_value){.number = 7});
        adamic_map_set(map, (adamic_value){.number = 2}, (adamic_value){.number = 8});
        adamic_object *iterator = adamic_collection_iterator(map, 2, adamic_ir_Number, adamic_ir_Number, false);
        adamic_release(map);
        for (;;) {
            adamic_object *status = adamic_iterator_next(iterator, adamic_ir_Number, hand_site);
            adamic_maybe_number value = adamic_maybe_number_unpack(status->slots[0].number);
            bool done = result_done(status);
            adamic_release(status);
            if (done) {
                assert(!value.present);
                break;
            }
            printf("%g\n", value.number);
        }
        adamic_release(iterator);
    } else if (strcmp(mode, "next_other") == 0) {
        /* The proof listed every origin, so a string here is the compiler's bug, said with the site. */
        (void)adamic_iterator_next(adamic_string_from_number(3), adamic_ir_String, hand_site);
        assert(!"a string stepped");
    } else if (strcmp(mode, "next_null") == 0) {
        (void)adamic_iterator_next(NULL, adamic_ir_String, hand_site);
        assert(!"NULL stepped");
    } else if (strcmp(mode, "next_kind") == 0) {
        /* A maybe-boolean yield is off the allowlist: no step runs. */
        (void)adamic_iterator_next(adamic_generator_new(0, &naming), adamic_ir_MaybeBoolean, hand_site);
        assert(!"a maybe-boolean yield stepped");
    } else {
        assert(!"unknown mode");
    }
    return 0;
}
