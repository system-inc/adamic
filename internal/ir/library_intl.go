package ir

// IntlCollatorNew is new Intl.Collator(locale, { usage: "sort", sensitivity: "variant", numeric: true }),
// tsc's one collator (core.ts:2100, #zjndrx5). Locale is a string or undefined, borrowed for the call. The
// result is a host object holding the locale and its compare, a closure the runtime makes, bound to the
// collator by construction as ECMA-402 says; a call of it stops in IntlCollatorCompareStop's words on
// both backends. A defined locale stops at construction, in IntlCollatorLocaleStop's.
type IntlCollatorNew struct{ Locale Expression }

func (IntlCollatorNew) Type() Type { return Object }

// IntlCollatorCompareStop is the whole run-time NotYet, with no operands, a collator's compare stops in
// (Adamic's ruling on #zjndrx5 at 06:39Z): putting the strings in would put run-time data in an
// allowlist entry, which is matched whole. The JavaScript backend's stand-in says these words, and
// native's library_intl.c the same, which internal/native's library_intl_test.go holds.
const IntlCollatorCompareStop = "Intl.Collator's compare isn't built yet"

// IntlCollatorLocaleStop is the whole run-time NotYet new Intl.Collator stops in where its locale is
// defined (Runtime's must 1 on #zjndrx5): Node throws a RangeError for a malformed tag and resolves a
// well-formed one, and neither is built, so construction never goes on silently.
const IntlCollatorLocaleStop = "Intl.Collator with a defined locale isn't built yet"
