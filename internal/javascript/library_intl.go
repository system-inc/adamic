package javascript

import "github.com/system-inc/adamic/internal/ir"

// Intl.Collator's stand-in (#zjndrx5 head 1) is what native's library_intl.c builds: the locale, and
// compare, a closure in the closure calling convention whose call stops in ir.IntlCollatorCompareStop's
// words, so the two backends stop alike, never this one answering with Node's own collator where
// native can't. A defined locale stops at construction, in ir.IntlCollatorLocaleStop's. The value is an
// AdamicIntlCollator, never a plain object, so a run-time key read, an in test and Object.keys stop on it
// (dynamic_keyed.go) as native's do on its host shape, where Node's collator has no own property at
// all. compare's length is Node's, 2, as native's closure carries it.
func intlCollatorRuntime() string {
	return "class AdamicIntlCollator {\n" +
		" constructor(locale) {\n" +
		"  if (locale !== undefined) panic(" + quote(ir.IntlCollatorLocaleStop) + ");\n" +
		"  this.compare = new AdamicClosure(() => panic(" + quote(ir.IntlCollatorCompareStop) + "), [], false, 2);\n" +
		"  this['#locale'] = locale;\n" +
		" }\n" +
		"}\n" +
		"const adamicIntlCollator = locale => new AdamicIntlCollator(locale);\n"
}
