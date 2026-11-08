// library_intl.c: Intl.Collator as far as tsc builds one (#zjndrx5 head 1, Adamic's ruling at 06:39Z).
//
// A collator is a host object: on Node its prototype is Intl.Collator.prototype, where compare is a
// getter, so it has no own property at all. Its shape is marked host (adamic.h's adamic_host_methods),
// so an in test, a run-time key read and Object.keys stop on it rather than answer from its slots
// (Runtime's must 2), and host_shapes_test.go holds that to a stop fixture. It holds compare, a closure
// bound to it by construction, and its locale, each in a private slot, #compare and #locale: Node's
// collator has no own field named compare, so neither has native's, and a view naming compare finds
// no field of that name to read or write (Runtime's ruling at 16:11Z, the allocation form). A call
// x.compare(...) reaches #compare through adamic_object_callee's arm for this shape, which knows it by
// its address (object.c), never by a field's name. The closure captures the locale, the one thing a
// compare would read, and never the collator itself, which would be a cycle no count frees. Nothing
// here compares: a call of compare stops in one fixed run-time NotYet, never an ordinal answer, since
// faking a locale's order would be a silent miscompile (docs/rulings.md, localeCompare). Real
// collation is #s82g5nx.

#include "adamic.h"
#include <string.h>

// collator_compare_slot is #compare's place in collator_shape, the one place new writes it and the
// value read reads it, so the two can't drift; adamic_object_callee's arm reads slot 0 by the same
// rule (object.c), and collator_names[collator_compare_slot] is checked to be "#compare" where it's read.
enum { collator_compare_slot = 0 };
static const char *const collator_names[] = {"#compare", "#locale"};
static const bool collator_references[] = {true, true};
static const adamic_field_kind collator_kinds[] = {adamic_field_reference, adamic_field_reference};
static const adamic_shape collator_shape = {2, collator_names, collator_references, &adamic_host_methods, collator_kinds};

// collator_compare is a collator's compare: ir.IntlCollatorCompareStop's words, with neither string,
// whatever it's handed.
static adamic_value collator_compare(adamic_closure *self, adamic_value *arguments) {
	(void)self;
	(void)arguments;
	static const char message[] = "Intl.Collator's compare isn't built yet";
	adamic_panic(message, sizeof message - 1);
}

adamic_object *adamic_intl_collator_new(const adamic_string *locale) {
	// Node resolves a well-formed tag and throws a RangeError for a malformed one, and neither is built
	// (Runtime's must 1), so a defined locale stops here, in ir.IntlCollatorLocaleStop's words.
	if (locale != NULL) {
		static const char message[] = "Intl.Collator with a defined locale isn't built yet";
		adamic_panic(message, sizeof message - 1);
	}
	adamic_closure *compare = adamic_closure_new(collator_compare, 1);
	// Node's bound compare takes two strings, so its length is 2.
	compare->source_length = 2;
	compare->cells[0] = adamic_cell_new((adamic_value){.reference = NULL}, true);
	adamic_object *collator = adamic_object_new(&collator_shape);
	adamic_intl_collator_shape_publish(&collator_shape);
	collator->slots[collator_compare_slot].reference = compare;
	collator->slots[1].reference = NULL;
	return collator;
}

// adamic_intl_collator_compare is a collator's compare read as a value, not called (#9tw2b9z): its
// #compare closure, retained for the caller, the collator borrowed. Lowering admits the read only off
// the new expression itself, so the collator is one adamic_intl_collator_new just made, and anything
// else reaching here is a compiler bug, stopped by name rather than retaining another shape's slot.
adamic_closure *adamic_intl_collator_compare(const adamic_object *collator) {
	if (collator == NULL || collator->shape != &collator_shape || strcmp(collator_names[collator_compare_slot], "#compare") != 0) { adamic_invariant("compare read off something other than a new collator"); }
	return adamic_retain(collator->slots[collator_compare_slot].reference);
}
