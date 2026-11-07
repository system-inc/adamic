// library_intl.c: Intl.Collator as far as tsc builds one (#zjndrx5 head 1, Adamic's ruling at 06:39Z).
//
// A collator is a host object: on Node its prototype is Intl.Collator.prototype, where compare is a
// getter, so it has no own property at all. Its shape is marked host (adamic.h's adamic_host_methods),
// so an in test, a run-time key read and Object.keys stop on it rather than answer from its slots
// (Runtime's must 2), and host_shapes_test.go holds that to a stop fixture. It holds compare, a closure
// bound to it by construction, and its locale. The closure captures the locale, the one thing a
// compare would read, and never the collator itself, which would be a cycle no count frees. Nothing
// here compares: a call of compare stops in one fixed run-time NotYet, never an ordinal answer, since
// faking a locale's order would be a silent miscompile (docs/rulings.md, localeCompare). Real
// collation is #s82g5nx.

#include "adamic.h"

static const char *const collator_names[] = {"compare", "#locale"};
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
	collator->slots[0].reference = compare;
	collator->slots[1].reference = NULL;
	return collator;
}
