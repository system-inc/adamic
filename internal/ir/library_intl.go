package ir

// IntlCollatorNew is new Intl.Collator(locale, { usage: "sort", sensitivity: "variant", numeric: true }),
// tsc's one collator (core.ts:2100, #zjndrx5). Locale is a string or undefined, borrowed for the call. The
// result is a host object holding the locale and its compare, a closure the runtime makes, bound to the
// collator by construction as ECMA-402 says; a call of it stops in IntlCollatorCompareStop's words on
// both backends. A defined locale stops at construction, in IntlCollatorLocaleStop's.
type IntlCollatorNew struct{ Locale Expression }

func (IntlCollatorNew) Type() Type { return Object }

// IntlCollatorCompareValue is a new collator's compare read as a value, not called (#9tw2b9z, F19):
// `new Intl.Collator(...).compare`, tsc's comparer at core.ts:2116. Collator is that IntlCollatorNew
// itself, which lowering admits as the only receiver (Runtime's ruling, (b)), so the value is the
// collator's own bound compare by allocation, with no other holder to tell apart. Natively it's the
// #compare closure, owned, the collator borrowed for the read; on JavaScript, the stand-in's getter
// bound to its collator. A call of it stops as compare's does.
type IntlCollatorCompareValue struct{ Collator Expression }

func (IntlCollatorCompareValue) Type() Type { return Closure }

// IntlCollatorCompareStop is the whole run-time NotYet, with no operands, a collator's compare stops in
// (Adamic's ruling on #zjndrx5 at 06:39Z): putting the strings in would put run-time data in an
// allowlist entry, which is matched whole. The JavaScript backend's stand-in says these words, and
// native's library_intl.c the same, which internal/native's library_intl_test.go holds.
const IntlCollatorCompareStop = "Intl.Collator's compare isn't built yet"

// IntlCollatorLocaleStop is the whole run-time NotYet new Intl.Collator stops in where its locale is
// defined (Runtime's must 1 on #zjndrx5): Node throws a RangeError for a malformed tag and resolves a
// well-formed one, and neither is built, so construction never goes on silently.
const IntlCollatorLocaleStop = "Intl.Collator with a defined locale isn't built yet"
