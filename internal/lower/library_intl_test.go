package lower

import (
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// intlCollatorConstruct is tsc's construction (core.ts:2100), its locale a string | undefined.
const intlCollatorConstruct = "function make(locale: string | undefined): Intl.Collator {\n\treturn new Intl.Collator(locale, { usage: \"sort\", sensitivity: \"variant\", numeric: true });\n}\n"

// tsc's collator lowers (#zjndrx5 head 1): new Intl.Collator from its locale and its options literal,
// its options exactly as tsc writes them, and compare called on the collator, directly, through
// a variable and inside the arrow tsc's comparer is. The oracle holds what they run to
// (internal/oracle's library_intl_test.go); this holds that they lower, to one IntlCollatorNew each.
func TestIntlCollatorLowers(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, source string
		made         int
	}{
		{"tsc's construction", intlCollatorConstruct + "console.log(typeof make(undefined));\n", 1},
		{"an undefined locale written in place", "const collator = new Intl.Collator(undefined, { usage: \"sort\", sensitivity: \"variant\", numeric: true });\nconsole.log(typeof collator);\n", 1},
		{"compare through a variable", intlCollatorConstruct + "const collator = make(undefined);\nconsole.log(String(collator.compare(\"a\", \"b\")));\n", 1},
		{"compare in tsc's comparer", intlCollatorConstruct + "function comparer(locale: string | undefined): (a: string, b: string) => number {\n\tconst collator = make(locale);\n\treturn (a, b) => collator.compare(a, b);\n}\nconsole.log(String(comparer(undefined)(\"a\", \"b\")));\n", 1},
		{"compare on the new collator", "console.log(String(new Intl.Collator(undefined, { usage: \"sort\", sensitivity: \"variant\", numeric: true }).compare(\"a\", \"b\")));\n", 1},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSource(t, probe.source)
			if err != nil {
				t.Fatalf("want it to lower, got %v", err)
			}
			if made := intlCollatorsMade(program); made != probe.made {
				t.Errorf("got %d IntlCollatorNew, want %d", made, probe.made)
			}
		})
	}
}

// intlCollatorsMade counts the IntlCollatorNew expressions a program holds, wherever they are.
func intlCollatorsMade(program *ir.Program) int {
	made := 0
	var walk func(reflect.Value)
	walk = func(value reflect.Value) {
		if !value.IsValid() {
			return
		}
		// Counted as the struct itself, so an interface holding one, which the walk passes through to
		// reach it, isn't counted again.
		if value.Kind() == reflect.Struct && value.Type() == reflect.TypeFor[ir.IntlCollatorNew]() {
			made++
		}
		switch value.Kind() {
		case reflect.Interface, reflect.Pointer:
			if !value.IsNil() {
				walk(value.Elem())
			}
		case reflect.Struct:
			for index := range value.NumField() {
				if value.Type().Field(index).IsExported() {
					walk(value.Field(index))
				}
			}
		case reflect.Slice, reflect.Array:
			for index := range value.Len() {
				walk(value.Index(index))
			}
		}
	}
	walk(reflect.ValueOf(program.Main))
	for _, function := range program.Functions {
		walk(reflect.ValueOf(function.Body))
	}
	return made
}

// Every other road to a collator is a compile-time NotYet that says what it is, at it (Adamic's three
// conditions at 06:39Z, the note after them at 06:40Z and Language's roads at 06:41Z): an option
// changed, added, cast or missing, options in another order, quoted, spread, shorthand, held in a const
// or handed back by a call, another locale type, no options or no arguments, resolvedOptions,
// supportedLocalesOf, Intl.Collator called without new and a class extending it. Of the enumerations,
// Object.hasOwn would answer from native's own compare slot, so it's a NotYet by name, as a write by
// key and Object.assign onto a collator are (internal/oracle's
// TestIntlCollatorWritesAndCopiesStopByName holds the rest of Runtime's roads on head 1's C); for...in
// and JSON.stringify stop in their own words; Object.keys, an in test and a run-time key read stop at
// run time (internal/oracle's own_keys.a, in_compare.a and reflect_get's stop_host_collator.a), so
// none lists compare (Runtime's must 2). compare read as a value, held, handed to sort, or through call
// or bind, stops where it stops today, the method value refusal, until Lower's #9tw2b9z reads the slot.
// Mutants: the value test dropped, so "numeric: false" lowers; an unknown option let by, so "caseFirst
// added" lowers; the resolvedOptions and without-new stops dropped, so those rows meet other words.
func TestIntlCollatorOtherRoadsAreNotYet(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, want string }{
		{"numeric: false", "const collator = new Intl.Collator(undefined, { usage: \"sort\", sensitivity: \"variant\", numeric: false });\nconsole.log(typeof collator);\n",
			"main.a:1:88: stage 0 can't lower new Intl.Collator with the option numeric: false (only tsc's { usage: \"sort\", sensitivity: \"variant\", numeric: true } is built) yet"},
		{"caseFirst added", "const collator = new Intl.Collator(undefined, { usage: \"sort\", sensitivity: \"variant\", numeric: true, caseFirst: \"upper\" });\nconsole.log(typeof collator);\n",
			"main.a:1:103: stage 0 can't lower new Intl.Collator with the option caseFirst (only tsc's { usage: \"sort\", sensitivity: \"variant\", numeric: true } is built) yet"},
		{"sensitivity base", "const collator = new Intl.Collator(undefined, { usage: \"sort\", sensitivity: \"base\", numeric: true });\nconsole.log(typeof collator);\n",
			"main.a:1:64: stage 0 can't lower new Intl.Collator with the option sensitivity: \"base\" (only tsc's { usage: \"sort\", sensitivity: \"variant\", numeric: true } is built) yet"},
		{"numeric cast to boolean", "const collator = new Intl.Collator(undefined, { usage: \"sort\", sensitivity: \"variant\", numeric: true as boolean });\nconsole.log(typeof collator);\n",
			"main.a:1:88: stage 0 can't lower new Intl.Collator with the option numeric: true as boolean (only tsc's { usage: \"sort\", sensitivity: \"variant\", numeric: true } is built) yet"},
		{"options in another order", "const collator = new Intl.Collator(undefined, { numeric: true, usage: \"sort\", sensitivity: \"variant\" });\nconsole.log(typeof collator);\n",
			"main.a:1:49: stage 0 can't lower new Intl.Collator with its options in another order than tsc's (only tsc's { usage: \"sort\", sensitivity: \"variant\", numeric: true } is built) yet"},
		{"quoted option names", "const collator = new Intl.Collator(undefined, { \"usage\": \"sort\", \"sensitivity\": \"variant\", \"numeric\": true });\nconsole.log(typeof collator);\n",
			"main.a:1:49: stage 0 can't lower new Intl.Collator with an option that isn't a plain field (only tsc's { usage: \"sort\", sensitivity: \"variant\", numeric: true } is built) yet"},
		{"a spread of options", "const base: Intl.CollatorOptions = { usage: \"sort\" };\nconst collator = new Intl.Collator(undefined, { ...base, sensitivity: \"variant\", numeric: true });\nconsole.log(typeof collator);\n",
			"main.a:2:49: stage 0 can't lower new Intl.Collator with an option that isn't a plain field (only tsc's { usage: \"sort\", sensitivity: \"variant\", numeric: true } is built) yet"},
		{"options held in a const", "const options: Intl.CollatorOptions = { usage: \"sort\", sensitivity: \"variant\", numeric: true };\nconst collator = new Intl.Collator(undefined, options);\nconsole.log(typeof collator);\n",
			"main.a:2:47: stage 0 can't lower new Intl.Collator with options that aren't an object literal written in place (only tsc's { usage: \"sort\", sensitivity: \"variant\", numeric: true } is built) yet"},
		{"options from a call", "function options(): Intl.CollatorOptions {\n\treturn { usage: \"sort\", sensitivity: \"variant\", numeric: true };\n}\nconst collator = new Intl.Collator(undefined, options());\nconsole.log(typeof collator);\n",
			"main.a:4:47: stage 0 can't lower new Intl.Collator with options that aren't an object literal written in place (only tsc's { usage: \"sort\", sensitivity: \"variant\", numeric: true } is built) yet"},
		{"an option missing", "const collator = new Intl.Collator(undefined, { usage: \"sort\", sensitivity: \"variant\" });\nconsole.log(typeof collator);\n",
			"main.a:1:47: stage 0 can't lower new Intl.Collator without the option numeric (only tsc's { usage: \"sort\", sensitivity: \"variant\", numeric: true } is built) yet"},
		{"a shorthand option", "const numeric = true;\nconst collator = new Intl.Collator(undefined, { usage: \"sort\", sensitivity: \"variant\", numeric });\nconsole.log(typeof collator);\n",
			"main.a:2:88: stage 0 can't lower new Intl.Collator with an option that isn't a plain field (only tsc's { usage: \"sort\", sensitivity: \"variant\", numeric: true } is built) yet"},
		{"no options", "const collator = new Intl.Collator(undefined);\nconsole.log(typeof collator);\n",
			"main.a:1:18: stage 0 can't lower new Intl.Collator without exactly a locale and options (only tsc's { usage: \"sort\", sensitivity: \"variant\", numeric: true } is built) yet"},
		{"no arguments", "const collator = new Intl.Collator();\nconsole.log(typeof collator);\n",
			"main.a:1:18: stage 0 can't lower new Intl.Collator without exactly a locale and options (only tsc's { usage: \"sort\", sensitivity: \"variant\", numeric: true } is built) yet"},
		{"a list of locales", "const collator = new Intl.Collator([\"en\"], { usage: \"sort\", sensitivity: \"variant\", numeric: true });\nconsole.log(typeof collator);\n",
			"main.a:1:36: stage 0 can't lower new Intl.Collator with a locale that isn't a string or undefined (a list of locales isn't built yet) yet"},
		{"resolvedOptions", "const collator = new Intl.Collator(undefined, { usage: \"sort\", sensitivity: \"variant\", numeric: true });\nconst options = collator.resolvedOptions();\nconsole.log(typeof options);\n",
			"main.a:2:17: stage 0 can't lower Intl.Collator's resolvedOptions yet"},
		{"supportedLocalesOf", "const locales = Intl.Collator.supportedLocalesOf(\"en\");\nconsole.log(String(locales.length));\n",
			"main.a:1:17: stage 0 can't lower Intl.Collator.supportedLocalesOf yet"},
		{"called without new", "const collator = Intl.Collator(undefined, { usage: \"sort\", sensitivity: \"variant\", numeric: true });\nconsole.log(typeof collator);\n",
			"main.a:1:18: stage 0 can't lower Intl.Collator called without new yet"},
		{"a class extending it", "class Sorter extends Intl.Collator {}\nconsole.log(typeof new Sorter());\n",
			"main.a:1:22: stage 0 can't lower a class extending Intl.Collator yet"},
		{"Object.hasOwn of compare", "const collator = new Intl.Collator(undefined, { usage: \"sort\", sensitivity: \"variant\", numeric: true });\nconsole.log(String(Object.hasOwn(collator, \"compare\")));\n",
			"main.a:2:20: stage 0 can't lower Object.hasOwn on an Intl.Collator, which has no own property on Node yet"},
		{"a write to compare by key", "const collator = new Intl.Collator(undefined, { usage: \"sort\", sensitivity: \"variant\", numeric: true });\ncollator[\"compare\"] = (a: string, b: string): number => a.length - b.length;\nconsole.log(\"wrote\");\n",
			"main.a:2:1: stage 0 can't lower a write to an Intl.Collator's compare, which Node's collator holds only on Intl.Collator.prototype (compare there is a getter, and writing it throws a TypeError) yet"},
		{"Object.assign onto one", "const collator = new Intl.Collator(undefined, { usage: \"sort\", sensitivity: \"variant\", numeric: true });\nObject.assign(collator, {});\nconsole.log(\"assigned\");\n",
			"main.a:2:1: stage 0 can't lower Object.assign on an Intl.Collator, which has no own property on Node yet"},
		{"for...in over one", "const collator = new Intl.Collator(undefined, { usage: \"sort\", sensitivity: \"variant\", numeric: true });\nfor (const key in collator) {\n\tconsole.log(key);\n}\n",
			"stage 0 can't lower for...in without a proven fixed plain-object origin"},
		{"JSON.stringify of one", "const collator = new Intl.Collator(undefined, { usage: \"sort\", sensitivity: \"variant\", numeric: true });\nconsole.log(JSON.stringify(collator) ?? \"undefined\");\n",
			"stage 0 can't lower JSON.stringify object references"},
		{"compare held in a const", "const collator = new Intl.Collator(undefined, { usage: \"sort\", sensitivity: \"variant\", numeric: true });\nconst compare = collator.compare;\nconsole.log(String(compare(\"a\", \"b\")));\n",
			"main.a:2:17: Adamic 0.1 refuses a method read as a value (compare would lose its object, and this with it)"},
		{"compare handed to sort", "const collator = new Intl.Collator(undefined, { usage: \"sort\", sensitivity: \"variant\", numeric: true });\nconsole.log([\"b\", \"a\"].sort(collator.compare).join());\n",
			"main.a:2:29: Adamic 0.1 refuses a method read as a value (compare would lose its object, and this with it)"},
		{"compare through call", "const collator = new Intl.Collator(undefined, { usage: \"sort\", sensitivity: \"variant\", numeric: true });\nconsole.log(String(collator.compare.call(collator, \"a\", \"b\")));\n",
			"main.a:2:20: Adamic 0.1 refuses a method read as a value (compare would lose its object, and this with it)"},
		{"compare through bind", "const collator = new Intl.Collator(undefined, { usage: \"sort\", sensitivity: \"variant\", numeric: true });\nconst bound = collator.compare.bind(collator);\nconsole.log(String(bound(\"a\", \"b\")));\n",
			"main.a:2:15: Adamic 0.1 refuses a method read as a value (compare would lose its object, and this with it)"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			if err == nil || !strings.Contains(err.Error(), probe.want) {
				t.Errorf("got %v, want a stop containing %q", err, probe.want)
			}
		})
	}
}

// delete collator.compare is true on Node, which has no own compare to delete, and compare still
// reads. TypeScript refuses the program itself, since compare isn't optional (TS2790), so it never
// reaches lowering, and that's the pin (Runtime's must on #zjndrx5 head 1's C).
func TestIntlCollatorDeleteOfCompareIsTypeScriptsError(t *testing.T) {
	t.Parallel()
	source := "const collator = new Intl.Collator(undefined, { usage: \"sort\", sensitivity: \"variant\", numeric: true });\nconsole.log(String(delete collator.compare));\nconsole.log(typeof collator.compare);\n"
	_, err := load.Load([]string{writeAgreementSource(t, "main.a", source)})
	if err == nil || !strings.Contains(err.Error(), "TS2790") {
		t.Fatalf("got %v, want TypeScript's TS2790, the operand of delete must be optional", err)
	}
}
