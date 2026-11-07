package javascript

import "github.com/system-inc/adamic/internal/ir"

// Intl.Collator's stand-in (#zjndrx5 head 1) is Node's collator in the shape native's library_intl.c
// builds: the locale, and compare, whose call stops in ir.IntlCollatorCompareStop's words, so the two
// backends stop alike, never this one answering with Node's own collator where native can't. A defined
// locale stops at construction, in ir.IntlCollatorLocaleStop's. As on Node, the collator has no own
// property (Runtime's ruling at 16:11Z, the allocation form): the locale and compare are private
// fields, and compare is a getter on the class's prototype, so a write to it throws a TypeError in the
// module's strict mode, 'compare' in a collator is true and Object.keys of one is []. The value is an
// AdamicIntlCollator, never a plain object, so a run-time key read, an in test and Object.keys stop on
// it (dynamic_keyed.go) as native's do on its host shape. compare(...) through a collator reaches the
// getter by adamicCallee's prototype road, which calls what it finds with the collator first, so the
// getter hands back a function of that form.
func intlCollatorRuntime() string {
	return "class AdamicIntlCollator {\n" +
		" #compare;\n" +
		" #locale;\n" +
		" constructor(locale) {\n" +
		"  if (locale !== undefined) panic(" + quote(ir.IntlCollatorLocaleStop) + ");\n" +
		"  this.#compare = (collator, x, y) => panic(" + quote(ir.IntlCollatorCompareStop) + ");\n" +
		"  this.#locale = locale;\n" +
		" }\n" +
		" get compare() { return this.#compare; }\n" +
		"}\n" +
		"const adamicIntlCollator = locale => new AdamicIntlCollator(locale);\n" +
		// A new collator's compare read as a value (#9tw2b9z): the getter's function, which takes its
		// collator first as adamicCallee's road calls it, bound to this collator, so it's called with the
		// two strings alone and its length is 2, as Node's bound compare's is.
		"const adamicIntlCollatorCompare = collator => { const compare = collator.compare; return (x, y) => compare(collator, x, y); };\n"
}
