package core

import (
	"sort"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// extraBooleanCastFile is where the fixtures pretend to live.
const extraBooleanCastFile = "/repository/source/ExtraBooleanCast.ts"

// The corpus is oxc's, copied rather than rewritten.
//
// Every case below is verbatim from `oxc/crates/oxc_linter/src/rules/eslint/no_extra_boolean_cast.rs`:
// 52 pass and 441 fail from the single Tester block, plus 382 entries from its `expect_fix` vector.
//
// # The counts do not line up and that is the point
//
// The snapshot records 507 diagnostics from those 441 fail inputs, so 66 inputs report twice and a
// fixture asserting one finding per input would be wrong on every one of them. The per-input counts
// below were recovered by aligning each snapshot diagnostic against the source line it prints,
// rather than by walking the two lists in order: 104 fail sources appear more than once under
// different options, and an in-order walk cannot tell those copies apart.
//
// # The option spellings
//
// Upstream runs 104 cases under `enforceForLogicalOperands` and 118 under
// `enforceForInnerExpressions`. Both are carried here under the single field they decode to, which
// is what oxc does, and the two blocks are kept separate below so the divergence stays visible: the
// 14 extra inner cases are exactly the shapes ESLint's legacy option would decline.

func TestNoExtraBooleanCastFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		options    any
		messageIds []string
	}{
		// --- defaults (219 cases) ---
		{"if (!!foo) {}", nil, []string{"redundantDoubleNegation"}},
		{"do {} while (!!foo)", nil, []string{"redundantDoubleNegation"}},
		{"while (!!foo) {}", nil, []string{"redundantDoubleNegation"}},
		{"!!foo ? bar : baz", nil, []string{"redundantDoubleNegation"}},
		{"for (; !!foo;) {}", nil, []string{"redundantDoubleNegation"}},
		{"!!!foo", nil, []string{"redundantDoubleNegation"}},
		{"Boolean(!!foo)", nil, []string{"redundantDoubleNegation"}},
		{"new Boolean(!!foo)", nil, []string{"redundantDoubleNegation"}},
		{"if (Boolean(foo)) {}", nil, []string{"redundantBooleanCall"}},
		{"do {} while (Boolean(foo))", nil, []string{"redundantBooleanCall"}},
		{"while (Boolean(foo)) {}", nil, []string{"redundantBooleanCall"}},
		{"Boolean(foo) ? bar : baz", nil, []string{"redundantBooleanCall"}},
		{"for (; Boolean(foo);) {}", nil, []string{"redundantBooleanCall"}},
		{"!Boolean(foo)", nil, []string{"redundantBooleanCall"}},
		{"!Boolean(foo && bar)", nil, []string{"redundantBooleanCall"}},
		{"!Boolean(foo + bar)", nil, []string{"redundantBooleanCall"}},
		{"!Boolean(+foo)", nil, []string{"redundantBooleanCall"}},
		{"!Boolean(foo())", nil, []string{"redundantBooleanCall"}},
		{"!Boolean(foo = bar)", nil, []string{"redundantBooleanCall"}},
		{"!Boolean(...foo);", nil, []string{"redundantBooleanCall"}},
		{"!Boolean(foo, bar());", nil, []string{"redundantBooleanCall"}},
		{"!Boolean((foo, bar()));", nil, []string{"redundantBooleanCall"}},
		{"!Boolean();", nil, []string{"redundantBooleanCall"}},
		{"!(Boolean());", nil, []string{"redundantBooleanCall"}},
		{"if (!Boolean()) { foo() }", nil, []string{"redundantBooleanCall"}},
		{"while (!Boolean()) { foo() }", nil, []string{"redundantBooleanCall"}},
		{"var foo = Boolean() ? bar() : baz()", nil, []string{"redundantBooleanCall"}},
		{"if (Boolean()) { foo() }", nil, []string{"redundantBooleanCall"}},
		{"while (Boolean()) { foo() }", nil, []string{"redundantBooleanCall"}},
		{"Boolean(Boolean(foo))", nil, []string{"redundantBooleanCall"}},
		{"Boolean(!!foo, bar)", nil, []string{"redundantDoubleNegation"}},
		{"function *foo() { yield!!a ? b : c }", nil, []string{"redundantDoubleNegation"}},
		{"function *foo() { yield!! a ? b : c }", nil, []string{"redundantDoubleNegation"}},
		{"function *foo() { yield! !a ? b : c }", nil, []string{"redundantDoubleNegation"}},
		{"function *foo() { yield !!a ? b : c }", nil, []string{"redundantDoubleNegation"}},
		{"function *foo() { yield(!!a) ? b : c }", nil, []string{"redundantDoubleNegation"}},
		{"function *foo() { yield/**/!!a ? b : c }", nil, []string{"redundantDoubleNegation"}},
		{"x=!!a ? b : c ", nil, []string{"redundantDoubleNegation"}},
		{"void!Boolean()", nil, []string{"redundantBooleanCall"}},
		{"void! Boolean()", nil, []string{"redundantBooleanCall"}},
		{"typeof!Boolean()", nil, []string{"redundantBooleanCall"}},
		{"(!Boolean())", nil, []string{"redundantBooleanCall"}},
		{"+!Boolean()", nil, []string{"redundantBooleanCall"}},
		{"void !Boolean()", nil, []string{"redundantBooleanCall"}},
		{"void(!Boolean())", nil, []string{"redundantBooleanCall"}},
		{"void/**/!Boolean()", nil, []string{"redundantBooleanCall"}},
		{"!/**/!!foo", nil, []string{"redundantDoubleNegation"}},
		{"!!/**/!foo", nil, []string{"redundantDoubleNegation"}},
		{"!!!/**/foo", nil, []string{"redundantDoubleNegation"}},
		{"!!!foo/**/", nil, []string{"redundantDoubleNegation"}},
		{"if(!/**/!foo);", nil, []string{"redundantDoubleNegation"}},
		{"(!!/**/foo ? 1 : 2)", nil, []string{"redundantDoubleNegation"}},
		{"!/**/Boolean(foo)", nil, []string{"redundantBooleanCall"}},
		{"!Boolean/**/(foo)", nil, []string{"redundantBooleanCall"}},
		{"!Boolean(/**/foo)", nil, []string{"redundantBooleanCall"}},
		{"!Boolean(foo/**/)", nil, []string{"redundantBooleanCall"}},
		{"!Boolean(foo)/**/", nil, []string{"redundantBooleanCall"}},
		{"if(Boolean/**/(foo));", nil, []string{"redundantBooleanCall"}},
		{"(Boolean(foo/**/) ? 1 : 2)", nil, []string{"redundantBooleanCall"}},
		{"/**/!Boolean()", nil, []string{"redundantBooleanCall"}},
		{"!/**/Boolean()", nil, []string{"redundantBooleanCall"}},
		{"!Boolean/**/()", nil, []string{"redundantBooleanCall"}},
		{"!Boolean(/**/)", nil, []string{"redundantBooleanCall"}},
		{"!Boolean()/**/", nil, []string{"redundantBooleanCall"}},
		{"if(!/**/Boolean());", nil, []string{"redundantBooleanCall"}},
		{"(!Boolean(/**/) ? 1 : 2)", nil, []string{"redundantBooleanCall"}},
		{"if(/**/Boolean());", nil, []string{"redundantBooleanCall"}},
		{"if(Boolean/**/());", nil, []string{"redundantBooleanCall"}},
		{"if(Boolean(/**/));", nil, []string{"redundantBooleanCall"}},
		{"if(Boolean()/**/);", nil, []string{"redundantBooleanCall"}},
		{"(Boolean/**/() ? 1 : 2)", nil, []string{"redundantBooleanCall"}},
		{"Boolean(!!(a, b))", nil, []string{"redundantDoubleNegation"}},
		{"Boolean(Boolean((a, b)))", nil, []string{"redundantBooleanCall"}},
		{"Boolean((!!(a, b)))", nil, []string{"redundantDoubleNegation"}},
		{"Boolean((Boolean((a, b))))", nil, []string{"redundantBooleanCall"}},
		{"Boolean(!(!(a, b)))", nil, []string{"redundantDoubleNegation"}},
		{"Boolean((!(!(a, b))))", nil, []string{"redundantDoubleNegation"}},
		{"Boolean(!!(a = b))", nil, []string{"redundantDoubleNegation"}},
		{"Boolean((!!(a = b)))", nil, []string{"redundantDoubleNegation"}},
		{"Boolean(Boolean(a = b))", nil, []string{"redundantBooleanCall"}},
		{"Boolean(Boolean((a += b)))", nil, []string{"redundantBooleanCall"}},
		{"Boolean(!!(a === b))", nil, []string{"redundantDoubleNegation"}},
		{"Boolean(!!((a !== b)))", nil, []string{"redundantDoubleNegation"}},
		{"Boolean(!!a.b)", nil, []string{"redundantDoubleNegation"}},
		{"Boolean(Boolean((a)))", nil, []string{"redundantBooleanCall"}},
		{"Boolean((!!(a)))", nil, []string{"redundantDoubleNegation"}},
		{"new Boolean(!!(a, b))", nil, []string{"redundantDoubleNegation"}},
		{"new Boolean(Boolean((a, b)))", nil, []string{"redundantBooleanCall"}},
		{"new Boolean((!!(a, b)))", nil, []string{"redundantDoubleNegation"}},
		{"new Boolean((Boolean((a, b))))", nil, []string{"redundantBooleanCall"}},
		{"new Boolean(!(!(a, b)))", nil, []string{"redundantDoubleNegation"}},
		{"new Boolean((!(!(a, b))))", nil, []string{"redundantDoubleNegation"}},
		{"new Boolean(!!(a = b))", nil, []string{"redundantDoubleNegation"}},
		{"new Boolean((!!(a = b)))", nil, []string{"redundantDoubleNegation"}},
		{"new Boolean(Boolean(a = b))", nil, []string{"redundantBooleanCall"}},
		{"new Boolean(Boolean((a += b)))", nil, []string{"redundantBooleanCall"}},
		{"new Boolean(!!(a === b))", nil, []string{"redundantDoubleNegation"}},
		{"new Boolean(!!((a !== b)))", nil, []string{"redundantDoubleNegation"}},
		{"new Boolean(!!a.b)", nil, []string{"redundantDoubleNegation"}},
		{"new Boolean(Boolean((a)))", nil, []string{"redundantBooleanCall"}},
		{"new Boolean((!!(a)))", nil, []string{"redundantDoubleNegation"}},
		{"if (!!(a, b));", nil, []string{"redundantDoubleNegation"}},
		{"if (Boolean((a, b)));", nil, []string{"redundantBooleanCall"}},
		{"if (!(!(a, b)));", nil, []string{"redundantDoubleNegation"}},
		{"if (!!(a = b));", nil, []string{"redundantDoubleNegation"}},
		{"if (Boolean(a = b));", nil, []string{"redundantBooleanCall"}},
		{"if (!!(a > b));", nil, []string{"redundantDoubleNegation"}},
		{"if (Boolean(a === b));", nil, []string{"redundantBooleanCall"}},
		{"if (!!f(a));", nil, []string{"redundantDoubleNegation"}},
		{"if (Boolean(f(a)));", nil, []string{"redundantBooleanCall"}},
		{"if (!!(f(a)));", nil, []string{"redundantDoubleNegation"}},
		{"if ((!!f(a)));", nil, []string{"redundantDoubleNegation"}},
		{"if ((Boolean(f(a))));", nil, []string{"redundantBooleanCall"}},
		{"if (!!a);", nil, []string{"redundantDoubleNegation"}},
		{"if (Boolean(a));", nil, []string{"redundantBooleanCall"}},
		{"while (!!(a, b));", nil, []string{"redundantDoubleNegation"}},
		{"while (Boolean((a, b)));", nil, []string{"redundantBooleanCall"}},
		{"while (!(!(a, b)));", nil, []string{"redundantDoubleNegation"}},
		{"while (!!(a = b));", nil, []string{"redundantDoubleNegation"}},
		{"while (Boolean(a = b));", nil, []string{"redundantBooleanCall"}},
		{"while (!!(a > b));", nil, []string{"redundantDoubleNegation"}},
		{"while (Boolean(a === b));", nil, []string{"redundantBooleanCall"}},
		{"while (!!f(a));", nil, []string{"redundantDoubleNegation"}},
		{"while (Boolean(f(a)));", nil, []string{"redundantBooleanCall"}},
		{"while (!!(f(a)));", nil, []string{"redundantDoubleNegation"}},
		{"while ((!!f(a)));", nil, []string{"redundantDoubleNegation"}},
		{"while ((Boolean(f(a))));", nil, []string{"redundantBooleanCall"}},
		{"while (!!a);", nil, []string{"redundantDoubleNegation"}},
		{"while (Boolean(a));", nil, []string{"redundantBooleanCall"}},
		{"do {} while (!!(a, b));", nil, []string{"redundantDoubleNegation"}},
		{"do {} while (Boolean((a, b)));", nil, []string{"redundantBooleanCall"}},
		{"do {} while (!(!(a, b)));", nil, []string{"redundantDoubleNegation"}},
		{"do {} while (!!(a = b));", nil, []string{"redundantDoubleNegation"}},
		{"do {} while (Boolean(a = b));", nil, []string{"redundantBooleanCall"}},
		{"do {} while (!!(a > b));", nil, []string{"redundantDoubleNegation"}},
		{"do {} while (Boolean(a === b));", nil, []string{"redundantBooleanCall"}},
		{"do {} while (!!f(a));", nil, []string{"redundantDoubleNegation"}},
		{"do {} while (Boolean(f(a)));", nil, []string{"redundantBooleanCall"}},
		{"do {} while (!!(f(a)));", nil, []string{"redundantDoubleNegation"}},
		{"do {} while ((!!f(a)));", nil, []string{"redundantDoubleNegation"}},
		{"do {} while ((Boolean(f(a))));", nil, []string{"redundantBooleanCall"}},
		{"do {} while (!!a);", nil, []string{"redundantDoubleNegation"}},
		{"do {} while (Boolean(a));", nil, []string{"redundantBooleanCall"}},
		{"for (; !!(a, b););", nil, []string{"redundantDoubleNegation"}},
		{"for (; Boolean((a, b)););", nil, []string{"redundantBooleanCall"}},
		{"for (; !(!(a, b)););", nil, []string{"redundantDoubleNegation"}},
		{"for (; !!(a = b););", nil, []string{"redundantDoubleNegation"}},
		{"for (; Boolean(a = b););", nil, []string{"redundantBooleanCall"}},
		{"for (; !!(a > b););", nil, []string{"redundantDoubleNegation"}},
		{"for (; Boolean(a === b););", nil, []string{"redundantBooleanCall"}},
		{"for (; !!f(a););", nil, []string{"redundantDoubleNegation"}},
		{"for (; Boolean(f(a)););", nil, []string{"redundantBooleanCall"}},
		{"for (; !!(f(a)););", nil, []string{"redundantDoubleNegation"}},
		{"for (; (!!f(a)););", nil, []string{"redundantDoubleNegation"}},
		{"for (; (Boolean(f(a))););", nil, []string{"redundantBooleanCall"}},
		{"for (; !!a;);", nil, []string{"redundantDoubleNegation"}},
		{"for (; Boolean(a););", nil, []string{"redundantBooleanCall"}},
		{"!!(a, b) ? c : d", nil, []string{"redundantDoubleNegation"}},
		{"(!!(a, b)) ? c : d", nil, []string{"redundantDoubleNegation"}},
		{"Boolean((a, b)) ? c : d", nil, []string{"redundantBooleanCall"}},
		{"!!(a = b) ? c : d", nil, []string{"redundantDoubleNegation"}},
		{"Boolean(a -= b) ? c : d", nil, []string{"redundantBooleanCall"}},
		{"(Boolean((a *= b))) ? c : d", nil, []string{"redundantBooleanCall"}},
		{"!!(a ? b : c) ? d : e", nil, []string{"redundantDoubleNegation"}},
		{"Boolean(a ? b : c) ? d : e", nil, []string{"redundantBooleanCall"}},
		{"!!(a || b) ? c : d", nil, []string{"redundantDoubleNegation"}},
		{"Boolean(a && b) ? c : d", nil, []string{"redundantBooleanCall"}},
		{"!!(a === b) ? c : d", nil, []string{"redundantDoubleNegation"}},
		{"Boolean(a < b) ? c : d", nil, []string{"redundantBooleanCall"}},
		{"!!((a !== b)) ? c : d", nil, []string{"redundantDoubleNegation"}},
		{"Boolean((a >= b)) ? c : d", nil, []string{"redundantBooleanCall"}},
		{"!!+a ? b : c", nil, []string{"redundantDoubleNegation"}},
		{"!!+(a) ? b : c", nil, []string{"redundantDoubleNegation"}},
		{"Boolean(!a) ? b : c", nil, []string{"redundantBooleanCall"}},
		{"!!f(a) ? b : c", nil, []string{"redundantDoubleNegation"}},
		{"(!!f(a)) ? b : c", nil, []string{"redundantDoubleNegation"}},
		{"Boolean(a.b) ? c : d", nil, []string{"redundantBooleanCall"}},
		{"!!a ? b : c", nil, []string{"redundantDoubleNegation"}},
		{"Boolean(a) ? b : c", nil, []string{"redundantBooleanCall"}},
		{"!!!(a, b)", nil, []string{"redundantDoubleNegation"}},
		{"!Boolean((a, b))", nil, []string{"redundantBooleanCall"}},
		{"!!!(a = b)", nil, []string{"redundantDoubleNegation"}},
		{"!!(!(a += b))", nil, []string{"redundantDoubleNegation"}},
		{"!(!!(a += b))", nil, []string{"redundantDoubleNegation"}},
		{"!Boolean(a -= b)", nil, []string{"redundantBooleanCall"}},
		{"!Boolean((a -= b))", nil, []string{"redundantBooleanCall"}},
		{"!(Boolean(a -= b))", nil, []string{"redundantBooleanCall"}},
		{"!!!(a || b)", nil, []string{"redundantDoubleNegation"}},
		{"!Boolean(a || b)", nil, []string{"redundantBooleanCall"}},
		{"!!!(a && b)", nil, []string{"redundantDoubleNegation"}},
		{"!Boolean(a && b)", nil, []string{"redundantBooleanCall"}},
		{"!!!(a != b)", nil, []string{"redundantDoubleNegation"}},
		{"!!!(a === b)", nil, []string{"redundantDoubleNegation"}},
		{"var x = !Boolean(a > b)", nil, []string{"redundantBooleanCall"}},
		{"!!!(a - b)", nil, []string{"redundantDoubleNegation"}},
		{"!!!(a ** b)", nil, []string{"redundantDoubleNegation"}},
		{"!Boolean(a ** b)", nil, []string{"redundantBooleanCall"}},
		{"async function f() { !!!(await a) }", nil, []string{"redundantDoubleNegation"}},
		{"async function f() { !Boolean(await a) }", nil, []string{"redundantBooleanCall"}},
		{"!!!!a", nil, []string{"redundantDoubleNegation", "redundantDoubleNegation"}},
		{"!!(!(!a))", nil, []string{"redundantDoubleNegation", "redundantDoubleNegation"}},
		{"!Boolean(!a)", nil, []string{"redundantBooleanCall"}},
		{"!Boolean((!a))", nil, []string{"redundantBooleanCall"}},
		{"!Boolean(!(a))", nil, []string{"redundantBooleanCall"}},
		{"!(Boolean(!a))", nil, []string{"redundantBooleanCall"}},
		{"!!!+a", nil, []string{"redundantDoubleNegation"}},
		{"!!!(+a)", nil, []string{"redundantDoubleNegation"}},
		{"!!(!+a)", nil, []string{"redundantDoubleNegation"}},
		{"!(!!+a)", nil, []string{"redundantDoubleNegation"}},
		{"!Boolean((-a))", nil, []string{"redundantBooleanCall"}},
		{"!Boolean(-(a))", nil, []string{"redundantBooleanCall"}},
		{"!!!(--a)", nil, []string{"redundantDoubleNegation"}},
		{"!Boolean(a++)", nil, []string{"redundantBooleanCall"}},
		{"!!!f(a)", nil, []string{"redundantDoubleNegation"}},
		{"!!!(f(a))", nil, []string{"redundantDoubleNegation"}},
		{"!!!a", nil, []string{"redundantDoubleNegation"}},
		{"!Boolean(a)", nil, []string{"redundantBooleanCall"}},
		{"if (Boolean?.(foo)) {};", nil, []string{"redundantBooleanCall"}},
		{"if (!Boolean(a as any)) { }", nil, []string{"redundantBooleanCall"}},

		// --- the deprecated enforceForLogicalOperands spelling (104 cases) ---
		{"if (!!foo || bar) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"if (!!foo && bar) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"if ((!!foo || bar) && bat) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"if (foo && !!bar) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"do {} while (!!foo || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"while (!!foo || bar) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"!!foo && bat ? bar : baz", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"for (; !!foo || bar;) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"!!!foo || bar", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"Boolean(!!foo || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"new Boolean(!!foo || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"if (Boolean(foo) || bar) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"do {} while (Boolean(foo) || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"while (Boolean(foo) || bar) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"Boolean(foo) || bat ? bar : baz", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"for (; Boolean(foo) || bar;) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"!Boolean(foo) || bar", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"!Boolean(foo && bar) || bat", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"!Boolean(foo + bar) || bat", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"!Boolean(+foo)  || bar", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"!Boolean(foo()) || bar", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"!Boolean(foo() || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"!Boolean(foo = bar) || bat", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"!Boolean(...foo) || bar;", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"!Boolean(foo, bar()) || bar;", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"!Boolean((foo, bar()) || bat);", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"!Boolean() || bar;", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"!(Boolean()) || bar;", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"if (!Boolean() || bar) { foo() }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"while (!Boolean() || bar) { foo() }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"var foo = Boolean() || bar ? bar() : baz()", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"if (Boolean() || bar) { foo() }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"while (Boolean() || bar) { foo() }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"function *foo() { yield(!!a || d) ? b : c }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"function *foo() { yield(!! a || d) ? b : c }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"function *foo() { yield(! !a || d) ? b : c }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"function *foo() { yield (!!a || d) ? b : c }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"function *foo() { yield/**/(!!a || d) ? b : c }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"x=!!a || d ? b : c ", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"void(!Boolean() || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"void(! Boolean() || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"typeof(!Boolean() || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"(!Boolean() || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"void/**/(!Boolean() || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"!/**/(!!foo || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"!!/**/!foo || bar", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"!!!/**/foo || bar", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"!(!!foo || bar)/**/", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"if(!/**/!foo || bar);", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"(!!/**/foo || bar ? 1 : 2)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"!/**/(Boolean(foo) || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"!Boolean/**/(foo) || bar", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"!Boolean(/**/foo) || bar", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"!Boolean(foo/**/) || bar", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"!(Boolean(foo)|| bar)/**/", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"if(Boolean/**/(foo) || bar);", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"(Boolean(foo/**/)|| bar ? 1 : 2)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"/**/!Boolean()|| bar", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"!/**/Boolean()|| bar", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"!Boolean/**/()|| bar", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"!Boolean(/**/)|| bar", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"(!Boolean()|| bar)/**/", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"if(!/**/Boolean()|| bar);", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"(!Boolean(/**/) || bar ? 1 : 2)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"if(/**/Boolean()|| bar);", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"if(Boolean/**/()|| bar);", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"if(Boolean(/**/)|| bar);", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"if(Boolean()|| bar/**/);", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"(Boolean/**/()|| bar ? 1 : 2)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"if (a && !!(b ? c : d)){}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"function *foo() { yield!!a || d ? b : c }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"if (!!(a, b) || !!(c, d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation", "redundantDoubleNegation"}},
		{"if (Boolean((a, b)) || Boolean((c, d))) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall", "redundantBooleanCall"}},
		{"if ((!!((a, b))) || (!!((c, d)))) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation", "redundantDoubleNegation"}},
		{"if (!!(a, b) && !!(c, d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation", "redundantDoubleNegation"}},
		{"if (Boolean((a, b)) && Boolean((c, d))) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall", "redundantBooleanCall"}},
		{"if ((!!((a, b))) && (!!((c, d)))) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation", "redundantDoubleNegation"}},
		{"if (!!(a = b) || !!(c = d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation", "redundantDoubleNegation"}},
		{"if (Boolean(a /= b) || Boolean(c /= d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall", "redundantBooleanCall"}},
		{"if (!!(a >>= b) && !!(c >>= d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation", "redundantDoubleNegation"}},
		{"if (Boolean(a **= b) && Boolean(c **= d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall", "redundantBooleanCall"}},
		{"if (!!(a ? b : c) || !!(d ? e : f)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation", "redundantDoubleNegation"}},
		{"if (Boolean(a ? b : c) || Boolean(d ? e : f)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall", "redundantBooleanCall"}},
		{"if (!!(a ? b : c) && !!(d ? e : f)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation", "redundantDoubleNegation"}},
		{"if (Boolean(a ? b : c) && Boolean(d ? e : f)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall", "redundantBooleanCall"}},
		{"if (!!(a || b) || !!(c || d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation", "redundantDoubleNegation"}},
		{"if (Boolean(a || b) || Boolean(c || d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall", "redundantBooleanCall"}},
		{"if (!!(a || b) && !!(c || d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation", "redundantDoubleNegation"}},
		{"if (Boolean(a || b) && Boolean(c || d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall", "redundantBooleanCall"}},
		{"if (!!(a && b) || !!(c && d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation", "redundantDoubleNegation"}},
		{"if (Boolean(a && b) || Boolean(c && d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall", "redundantBooleanCall"}},
		{"if (!!(a && b) && !!(c && d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation", "redundantDoubleNegation"}},
		{"if (Boolean(a && b) && Boolean(c && d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall", "redundantBooleanCall"}},
		{"if (!!(a !== b) || !!(c !== d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation", "redundantDoubleNegation"}},
		{"if (Boolean(a != b) || Boolean(c != d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall", "redundantBooleanCall"}},
		{"if (!!(a === b) && !!(c === d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation", "redundantDoubleNegation"}},
		{"if (!!(a > b) || !!(c < d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation", "redundantDoubleNegation"}},
		{"if (Boolean(!a) || Boolean(+b)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall", "redundantBooleanCall"}},
		{"if (!!f(a) && !!b.c) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation", "redundantDoubleNegation"}},
		{"if (Boolean(a) || !!b) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall", "redundantDoubleNegation"}},
		{"if (!!a && Boolean(b)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation", "redundantBooleanCall"}},
		{"if ((!!a) || (Boolean(b))) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation", "redundantBooleanCall"}},
		{"if (Boolean(a ?? b) || c) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"if (Boolean?.(a ?? b) || c) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},

		// --- enforceForInnerExpressions (118 cases) ---
		{"if (!!foo || bar) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"if (!!foo && bar) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"if ((!!foo || bar) && bat) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"if (foo && !!bar) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"do {} while (!!foo || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"while (!!foo || bar) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"!!foo && bat ? bar : baz", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"for (; !!foo || bar;) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"!!!foo || bar", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"Boolean(!!foo || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"new Boolean(!!foo || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"if (Boolean(foo) || bar) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"do {} while (Boolean(foo) || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"while (Boolean(foo) || bar) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"Boolean(foo) || bat ? bar : baz", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"for (; Boolean(foo) || bar;) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"!Boolean(foo) || bar", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"!Boolean(foo && bar) || bat", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"!Boolean(foo + bar) || bat", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"!Boolean(+foo)  || bar", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"!Boolean(foo()) || bar", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"!Boolean(foo() || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"!Boolean(foo = bar) || bat", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"!Boolean(...foo) || bar;", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"!Boolean(foo, bar()) || bar;", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"!Boolean((foo, bar()) || bat);", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"!Boolean() || bar;", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"!(Boolean()) || bar;", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"if (!Boolean() || bar) { foo() }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"while (!Boolean() || bar) { foo() }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"var foo = Boolean() || bar ? bar() : baz()", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"if (Boolean() || bar) { foo() }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"while (Boolean() || bar) { foo() }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"function *foo() { yield(!!a || d) ? b : c }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"function *foo() { yield(!! a || d) ? b : c }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"function *foo() { yield(! !a || d) ? b : c }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"function *foo() { yield (!!a || d) ? b : c }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"function *foo() { yield/**/(!!a || d) ? b : c }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"x=!!a || d ? b : c ", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"void(!Boolean() || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"void(! Boolean() || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"typeof(!Boolean() || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"(!Boolean() || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"void/**/(!Boolean() || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"!/**/(!!foo || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"!!/**/!foo || bar", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"!!!/**/foo || bar", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"!(!!foo || bar)/**/", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"if(!/**/!foo || bar);", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"(!!/**/foo || bar ? 1 : 2)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"!/**/(Boolean(foo) || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"!Boolean/**/(foo) || bar", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"!Boolean(/**/foo) || bar", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"!Boolean(foo/**/) || bar", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"!(Boolean(foo)|| bar)/**/", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"if(Boolean/**/(foo) || bar);", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"(Boolean(foo/**/)|| bar ? 1 : 2)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"/**/!Boolean()|| bar", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"!/**/Boolean()|| bar", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"!Boolean/**/()|| bar", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"!Boolean(/**/)|| bar", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"(!Boolean()|| bar)/**/", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"if(!/**/Boolean()|| bar);", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"(!Boolean(/**/) || bar ? 1 : 2)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"if(/**/Boolean()|| bar);", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"if(Boolean/**/()|| bar);", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"if(Boolean(/**/)|| bar);", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"if(Boolean()|| bar/**/);", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"(Boolean/**/()|| bar ? 1 : 2)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"if (a && !!(b ? c : d)){}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"function *foo() { yield!!a || d ? b : c }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"if ((1, 2, Boolean(3))) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"if (a ?? Boolean(b)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"if (a ?? Boolean(b || c)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"if ((a, b, c ?? (d, e, f ?? Boolean(g)))) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"if (!!(a, b) || !!(c, d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation", "redundantDoubleNegation"}},
		{"if (Boolean((a, b)) || Boolean((c, d))) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall", "redundantBooleanCall"}},
		{"if ((!!((a, b))) || (!!((c, d)))) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation", "redundantDoubleNegation"}},
		{"if (!!(a, b) && !!(c, d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation", "redundantDoubleNegation"}},
		{"if (Boolean((a, b)) && Boolean((c, d))) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall", "redundantBooleanCall"}},
		{"if ((!!((a, b))) && (!!((c, d)))) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation", "redundantDoubleNegation"}},
		{"if (!!(a = b) || !!(c = d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation", "redundantDoubleNegation"}},
		{"if (Boolean(a /= b) || Boolean(c /= d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall", "redundantBooleanCall"}},
		{"if (!!(a >>= b) && !!(c >>= d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation", "redundantDoubleNegation"}},
		{"if (Boolean(a **= b) && Boolean(c **= d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall", "redundantBooleanCall"}},
		{"if (!!(a ? b : c) || !!(d ? e : f)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation", "redundantDoubleNegation"}},
		{"if (Boolean(a ? b : c) || Boolean(d ? e : f)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall", "redundantBooleanCall"}},
		{"if (!!(a ? b : c) && !!(d ? e : f)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation", "redundantDoubleNegation"}},
		{"if (Boolean(a ? b : c) && Boolean(d ? e : f)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall", "redundantBooleanCall"}},
		{"if (!!(a || b) || !!(c || d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation", "redundantDoubleNegation"}},
		{"if (Boolean(a || b) || Boolean(c || d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall", "redundantBooleanCall"}},
		{"if (!!(a || b) && !!(c || d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation", "redundantDoubleNegation"}},
		{"if (Boolean(a || b) && Boolean(c || d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall", "redundantBooleanCall"}},
		{"if (!!(a && b) || !!(c && d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation", "redundantDoubleNegation"}},
		{"if (Boolean(a && b) || Boolean(c && d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall", "redundantBooleanCall"}},
		{"if (!!(a && b) && !!(c && d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation", "redundantDoubleNegation"}},
		{"if (Boolean(a && b) && Boolean(c && d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall", "redundantBooleanCall"}},
		{"if (!!(a !== b) || !!(c !== d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation", "redundantDoubleNegation"}},
		{"if (Boolean(a != b) || Boolean(c != d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall", "redundantBooleanCall"}},
		{"if (!!(a === b) && !!(c === d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation", "redundantDoubleNegation"}},
		{"if (!!(a > b) || !!(c < d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation", "redundantDoubleNegation"}},
		{"if (Boolean(!a) || Boolean(+b)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall", "redundantBooleanCall"}},
		{"if (!!f(a) && !!b.c) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation", "redundantDoubleNegation"}},
		{"if (Boolean(a) || !!b) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall", "redundantDoubleNegation"}},
		{"if (!!a && Boolean(b)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation", "redundantBooleanCall"}},
		{"if ((!!a) || (Boolean(b))) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation", "redundantBooleanCall"}},
		{"if (Boolean(a ?? b) || c) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"if (Boolean?.(a ?? b) || c) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"if (a ? Boolean(b) : c) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"if (a ? b : Boolean(c)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"if (a ? b : Boolean(c ? d : e)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"const ternary = Boolean(bar ? !!baz : bat);", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"const commaOperator = Boolean((bar, baz, !!bat));", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantDoubleNegation"}},
		{"for (let i = 0; (console.log(i), Boolean(i < 10)); i++) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"const nullishCoalescingOperator = Boolean(bar ?? Boolean(baz));", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
		{"if (a ? Boolean(b = c) : Boolean(d = e));", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall", "redundantBooleanCall"}},
		{"if (a ? Boolean((b, c)) : Boolean((d, e)));", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall", "redundantBooleanCall"}},
		{"\n\tfunction * generator() {\n\t    if (a ? Boolean(yield y) : x) {\n\t        return a;\n\t    };\n\t}\n\t", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}, []string{"redundantBooleanCall"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunWithOptions(t, NoExtraBooleanCast, extraBooleanCastFile,
					testCase.sourceText, testCase.options), testCase.messageIds...)
		})
	}
}

// The clean cases, which are where the discrimination actually lives.
//
// `if (new Boolean(foo)) {}` is the sharpest: a construction is a coercing *context* for whatever it
// wraps, and is never itself a redundant cast, so a rule that treated the two symmetrically reports
// it. `if ((Boolean(1), 2)) {}` is next: the call is in a sequence but not its last element, so its
// value is discarded rather than tested. And every `var foo = bar || !!baz` case pins the recursion
// terminating: the `||` is not itself in a coercing position, so nothing inside it is either.
func TestNoExtraBooleanCastStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		options    any
	}{
		{"Boolean(bar, !!baz);", nil},
		{"var foo = !!bar;", nil},
		{"function foo() { return !!bar; }", nil},
		{"var foo = bar() ? !!baz : !!bat", nil},
		{"for(!!foo;;) {}", nil},
		{"for(;; !!foo) {}", nil},
		{"var foo = Boolean(bar);", nil},
		{"function foo() { return Boolean(bar); }", nil},
		{"var foo = bar() ? Boolean(baz) : Boolean(bat)", nil},
		{"for(Boolean(foo);;) {}", nil},
		{"for(;; Boolean(foo)) {}", nil},
		{"if (new Boolean(foo)) {}", nil},
		{"if ((Boolean(1), 2)) {}", nil},
		{"var foo = bar || !!baz", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"var foo = bar && !!baz", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"var foo = bar || (baz && !!bat)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"function foo() { return (!!bar || baz); }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"var foo = bar() ? (!!baz && bat) : (!!bat && qux)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"for(!!(foo && bar);;) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"for(;; !!(foo || bar)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"var foo = Boolean(bar) || baz;", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"var foo = bar || Boolean(baz);", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"var foo = Boolean(bar) || Boolean(baz);", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"function foo() { return (Boolean(bar) || baz); }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"var foo = bar() ? Boolean(baz) || bat : Boolean(bat)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"for(Boolean(foo) || bar;;) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"for(;; Boolean(foo) || bar) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (new Boolean(foo) || bar) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (!!foo || bar) {}", nil},
		{"if (!!foo || bar) {}", NoExtraBooleanCastOptions{}},
		{"if (!!foo || bar) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: false}},
		{"if ((!!foo || bar) === baz) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (!!foo ?? bar) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"var foo = bar || !!baz", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"var foo = bar && !!baz", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"var foo = bar || (baz && !!bat)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"function foo() { return (!!bar || baz); }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"var foo = bar() ? (!!baz && bat) : (!!bat && qux)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"for(!!(foo && bar);;) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"for(;; !!(foo || bar)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"var foo = Boolean(bar) || baz;", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"var foo = bar || Boolean(baz);", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"var foo = Boolean(bar) || Boolean(baz);", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"function foo() { return (Boolean(bar) || baz); }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"var foo = bar() ? Boolean(baz) || bat : Boolean(bat)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"for(Boolean(foo) || bar;;) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"for(;; Boolean(foo) || bar) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (new Boolean(foo) || bar) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (!!foo || bar) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: false}},
		{"if ((!!foo || bar) === baz) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (!!foo ?? bar) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if ((1, Boolean(2), 3)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t,
				rule_testing.RunWithOptions(t, NoExtraBooleanCast, extraBooleanCastFile,
					testCase.sourceText, testCase.options))
		})
	}
}

// The Boolean-call repairs, asserted through the shared harness.
//
// These 199 entries are the ones whose only finding is a redundant `Boolean(...)`, so every repair
// is a Fix and `rule_testing.ExpectFixedSource` can apply them. It is the shelf's assertion and it
// refuses overlapping fixes rather than guessing which wins, which is stricter than anything written
// here would be.
//
// The expected output is upstream's own, from the `expect_fix` vector its Tester runs. That is the
// assertion the message-id cases structurally cannot make: a rule reporting the right finding at the
// wrong span, or writing the right text over the wrong bytes, passes all 441 of them and is broken.
func TestNoExtraBooleanCastFixesBooleanCalls(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantSource string
		options    any
	}{
		{"if (Boolean(foo)) {}", "if (foo) {}", nil},
		{"do {} while (Boolean(foo))", "do {} while (foo)", nil},
		{"while (Boolean(foo)) {}", "while (foo) {}", nil},
		{"Boolean(foo) ? bar : baz", "foo ? bar : baz", nil},
		{"for (; Boolean(foo);) {}", "for (; foo;) {}", nil},
		{"!Boolean(foo)", "!foo", nil},
		{"!Boolean(foo && bar)", "!(foo && bar)", nil},
		{"!Boolean(foo + bar)", "!(foo + bar)", nil},
		{"!Boolean(+foo)", "!+foo", nil},
		{"!Boolean(foo())", "!foo()", nil},
		{"!Boolean(foo = bar)", "!(foo = bar)", nil},
		{"!Boolean((foo, bar()));", "!(foo, bar());", nil},
		{"!Boolean();", "true;", nil},
		{"!(Boolean());", "true;", nil},
		{"if (!Boolean()) { foo() }", "if (true) { foo() }", nil},
		{"while (!Boolean()) { foo() }", "while (true) { foo() }", nil},
		{"var foo = Boolean() ? bar() : baz()", "var foo = false ? bar() : baz()", nil},
		{"if (Boolean()) { foo() }", "if (false) { foo() }", nil},
		{"while (Boolean()) { foo() }", "while (false) { foo() }", nil},
		{"Boolean(Boolean(foo))", "Boolean(foo)", nil},
		{"void!Boolean()", "void true", nil},
		{"void! Boolean()", "void true", nil},
		{"typeof!Boolean()", "typeof true", nil},
		{"(!Boolean())", "(true)", nil},
		{"+!Boolean()", "+true", nil},
		{"void !Boolean()", "void true", nil},
		{"void(!Boolean())", "void(true)", nil},
		{"void/**/!Boolean()", "void/**/true", nil},
		{"!/**/Boolean(foo)", "!/**/foo", nil},
		{"!Boolean(foo)/**/", "!foo/**/", nil},
		{"/**/!Boolean()", "/**/true", nil},
		{"!Boolean()/**/", "true/**/", nil},
		{"if(/**/Boolean());", "if(/**/false);", nil},
		{"if(Boolean()/**/);", "if(false/**/);", nil},
		{"if (Boolean(foo) || bar) {}", "if (foo || bar) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"do {} while (Boolean(foo) || bar)", "do {} while (foo || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"while (Boolean(foo) || bar) {}", "while (foo || bar) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"Boolean(foo) || bat ? bar : baz", "foo || bat ? bar : baz", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"for (; Boolean(foo) || bar;) {}", "for (; foo || bar;) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"!Boolean(foo) || bar", "!foo || bar", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"!Boolean(foo && bar) || bat", "!(foo && bar) || bat", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"!Boolean(foo + bar) || bat", "!(foo + bar) || bat", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"!Boolean(+foo)  || bar", "!+foo  || bar", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"!Boolean(foo()) || bar", "!foo() || bar", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"!Boolean(foo() || bar)", "!(foo() || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"!Boolean(foo = bar) || bat", "!(foo = bar) || bat", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"!Boolean((foo, bar()) || bat);", "!((foo, bar()) || bat);", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"!Boolean() || bar;", "true || bar;", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"!(Boolean()) || bar;", "true || bar;", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (!Boolean() || bar) { foo() }", "if (true || bar) { foo() }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"while (!Boolean() || bar) { foo() }", "while (true || bar) { foo() }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"var foo = Boolean() || bar ? bar() : baz()", "var foo = false || bar ? bar() : baz()", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (Boolean() || bar) { foo() }", "if (false || bar) { foo() }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"while (Boolean() || bar) { foo() }", "while (false || bar) { foo() }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"void(!Boolean() || bar)", "void(true || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"void(! Boolean() || bar)", "void(true || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"typeof(!Boolean() || bar)", "typeof(true || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"(!Boolean() || bar)", "(true || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"void/**/(!Boolean() || bar)", "void/**/(true || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"!/**/(Boolean(foo) || bar)", "!/**/(foo || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"!(Boolean(foo)|| bar)/**/", "!(foo|| bar)/**/", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"/**/!Boolean()|| bar", "/**/true|| bar", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"(!Boolean()|| bar)/**/", "(true|| bar)/**/", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if(/**/Boolean()|| bar);", "if(/**/false|| bar);", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if(Boolean()|| bar/**/);", "if(false|| bar/**/);", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (Boolean(foo) || bar) {}", "if (foo || bar) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"do {} while (Boolean(foo) || bar)", "do {} while (foo || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"while (Boolean(foo) || bar) {}", "while (foo || bar) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"Boolean(foo) || bat ? bar : baz", "foo || bat ? bar : baz", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"for (; Boolean(foo) || bar;) {}", "for (; foo || bar;) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"!Boolean(foo) || bar", "!foo || bar", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"!Boolean(foo && bar) || bat", "!(foo && bar) || bat", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"!Boolean(foo + bar) || bat", "!(foo + bar) || bat", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"!Boolean(+foo)  || bar", "!+foo  || bar", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"!Boolean(foo()) || bar", "!foo() || bar", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"!Boolean(foo() || bar)", "!(foo() || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"!Boolean(foo = bar) || bat", "!(foo = bar) || bat", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"!Boolean((foo, bar()) || bat);", "!((foo, bar()) || bat);", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"!Boolean() || bar;", "true || bar;", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"!(Boolean()) || bar;", "true || bar;", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (!Boolean() || bar) { foo() }", "if (true || bar) { foo() }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"while (!Boolean() || bar) { foo() }", "while (true || bar) { foo() }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"var foo = Boolean() || bar ? bar() : baz()", "var foo = false || bar ? bar() : baz()", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (Boolean() || bar) { foo() }", "if (false || bar) { foo() }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"while (Boolean() || bar) { foo() }", "while (false || bar) { foo() }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"void(!Boolean() || bar)", "void(true || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"void(! Boolean() || bar)", "void(true || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"typeof(!Boolean() || bar)", "typeof(true || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"(!Boolean() || bar)", "(true || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"void/**/(!Boolean() || bar)", "void/**/(true || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"!/**/(Boolean(foo) || bar)", "!/**/(foo || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"!(Boolean(foo)|| bar)/**/", "!(foo|| bar)/**/", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"/**/!Boolean()|| bar", "/**/true|| bar", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"(!Boolean()|| bar)/**/", "(true|| bar)/**/", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if(/**/Boolean()|| bar);", "if(/**/false|| bar);", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if(Boolean()|| bar/**/);", "if(false|| bar/**/);", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"Boolean(Boolean((a, b)))", "Boolean((a, b))", nil},
		{"Boolean((Boolean((a, b))))", "Boolean((a, b))", nil},
		{"Boolean(Boolean(a = b))", "Boolean(a = b)", nil},
		{"Boolean(Boolean((a += b)))", "Boolean(a += b)", nil},
		{"Boolean(Boolean((a)))", "Boolean(a)", nil},
		{"new Boolean(Boolean((a, b)))", "new Boolean((a, b))", nil},
		{"new Boolean((Boolean((a, b))))", "new Boolean((a, b))", nil},
		{"new Boolean(Boolean(a = b))", "new Boolean(a = b)", nil},
		{"new Boolean(Boolean((a += b)))", "new Boolean(a += b)", nil},
		{"new Boolean(Boolean((a)))", "new Boolean(a)", nil},
		{"if (Boolean((a, b)));", "if (a, b);", nil},
		{"if (Boolean(a = b));", "if (a = b);", nil},
		{"if (Boolean(a === b));", "if (a === b);", nil},
		{"if (Boolean(f(a)));", "if (f(a));", nil},
		{"if ((Boolean(f(a))));", "if ((f(a)));", nil},
		{"if (Boolean(a));", "if (a);", nil},
		{"while (Boolean((a, b)));", "while (a, b);", nil},
		{"while (Boolean(a = b));", "while (a = b);", nil},
		{"while (Boolean(a === b));", "while (a === b);", nil},
		{"while (Boolean(f(a)));", "while (f(a));", nil},
		{"while ((Boolean(f(a))));", "while ((f(a)));", nil},
		{"while (Boolean(a));", "while (a);", nil},
		{"do {} while (Boolean((a, b)));", "do {} while (a, b);", nil},
		{"do {} while (Boolean(a = b));", "do {} while (a = b);", nil},
		{"do {} while (Boolean(a === b));", "do {} while (a === b);", nil},
		{"do {} while (Boolean(f(a)));", "do {} while (f(a));", nil},
		{"do {} while ((Boolean(f(a))));", "do {} while ((f(a)));", nil},
		{"do {} while (Boolean(a));", "do {} while (a);", nil},
		{"for (; Boolean((a, b)););", "for (; a, b;);", nil},
		{"for (; Boolean(a = b););", "for (; a = b;);", nil},
		{"for (; Boolean(a === b););", "for (; a === b;);", nil},
		{"for (; Boolean(f(a)););", "for (; f(a););", nil},
		{"for (; (Boolean(f(a))););", "for (; (f(a)););", nil},
		{"for (; Boolean(a););", "for (; a;);", nil},
		{"Boolean((a, b)) ? c : d", "(a, b) ? c : d", nil},
		{"Boolean(a -= b) ? c : d", "(a -= b) ? c : d", nil},
		{"(Boolean((a *= b))) ? c : d", "(a *= b) ? c : d", nil},
		{"Boolean(a ? b : c) ? d : e", "(a ? b : c) ? d : e", nil},
		{"Boolean(a && b) ? c : d", "a && b ? c : d", nil},
		{"Boolean(a < b) ? c : d", "a < b ? c : d", nil},
		{"Boolean((a >= b)) ? c : d", "a >= b ? c : d", nil},
		{"Boolean(!a) ? b : c", "!a ? b : c", nil},
		{"Boolean(a.b) ? c : d", "a.b ? c : d", nil},
		{"Boolean(a) ? b : c", "a ? b : c", nil},
		{"!Boolean((a, b))", "!(a, b)", nil},
		{"!Boolean(a -= b)", "!(a -= b)", nil},
		{"!Boolean((a -= b))", "!(a -= b)", nil},
		{"!(Boolean(a -= b))", "!(a -= b)", nil},
		{"!Boolean(a || b)", "!(a || b)", nil},
		{"!Boolean(a && b)", "!(a && b)", nil},
		{"var x = !Boolean(a > b)", "var x = !(a > b)", nil},
		{"!Boolean(a ** b)", "!(a ** b)", nil},
		{"async function f() { !Boolean(await a) }", "async function f() { !await a }", nil},
		{"!Boolean(!a)", "!!a", nil},
		{"!Boolean((!a))", "!!a", nil},
		{"!Boolean(!(a))", "!!(a)", nil},
		{"!(Boolean(!a))", "!(!a)", nil},
		{"!Boolean((-a))", "!-a", nil},
		{"!Boolean(-(a))", "!-(a)", nil},
		{"!Boolean(a++)", "!a++", nil},
		{"!Boolean(a)", "!a", nil},
		{"if (Boolean((a, b)) || Boolean((c, d))) {}", "if ((a, b) || (c, d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (Boolean((a, b)) && Boolean((c, d))) {}", "if ((a, b) && (c, d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (Boolean(a /= b) || Boolean(c /= d)) {}", "if ((a /= b) || (c /= d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (Boolean(a **= b) && Boolean(c **= d)) {}", "if ((a **= b) && (c **= d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (Boolean(a ? b : c) || Boolean(d ? e : f)) {}", "if ((a ? b : c) || (d ? e : f)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (Boolean(a ? b : c) && Boolean(d ? e : f)) {}", "if ((a ? b : c) && (d ? e : f)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (Boolean(a || b) || Boolean(c || d)) {}", "if (a || b || (c || d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (Boolean(a || b) && Boolean(c || d)) {}", "if ((a || b) && (c || d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (Boolean(a && b) || Boolean(c && d)) {}", "if (a && b || c && d) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (Boolean(a && b) && Boolean(c && d)) {}", "if (a && b && (c && d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (Boolean(a != b) || Boolean(c != d)) {}", "if (a != b || c != d) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (Boolean(!a) || Boolean(+b)) {}", "if (!a || +b) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (Boolean(a ?? b) || c) {}", "if ((a ?? b) || c) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (Boolean?.(foo)) {};", "if (foo) {};", nil},
		{"if (Boolean?.(a ?? b) || c) {}", "if ((a ?? b) || c) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (!Boolean(a as any)) { }", "if (!(a as any)) { }", nil},
		{"if ((1, 2, Boolean(3))) {}", "if ((1, 2, 3)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (a ?? Boolean(b)) {}", "if (a ?? b) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (a ?? Boolean(b || c)) {}", "if (a ?? (b || c)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if ((a, b, c ?? (d, e, f ?? Boolean(g)))) {}", "if ((a, b, c ?? (d, e, f ?? g))) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (Boolean((a, b)) || Boolean((c, d))) {}", "if ((a, b) || (c, d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (Boolean((a, b)) && Boolean((c, d))) {}", "if ((a, b) && (c, d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (Boolean(a /= b) || Boolean(c /= d)) {}", "if ((a /= b) || (c /= d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (Boolean(a **= b) && Boolean(c **= d)) {}", "if ((a **= b) && (c **= d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (Boolean(a ? b : c) || Boolean(d ? e : f)) {}", "if ((a ? b : c) || (d ? e : f)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (Boolean(a ? b : c) && Boolean(d ? e : f)) {}", "if ((a ? b : c) && (d ? e : f)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (Boolean(a || b) || Boolean(c || d)) {}", "if (a || b || (c || d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (Boolean(a || b) && Boolean(c || d)) {}", "if ((a || b) && (c || d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (Boolean(a && b) || Boolean(c && d)) {}", "if (a && b || c && d) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (Boolean(a && b) && Boolean(c && d)) {}", "if (a && b && (c && d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (Boolean(a != b) || Boolean(c != d)) {}", "if (a != b || c != d) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (Boolean(!a) || Boolean(+b)) {}", "if (!a || +b) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (Boolean(a ?? b) || c) {}", "if ((a ?? b) || c) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (Boolean?.(a ?? b) || c) {}", "if ((a ?? b) || c) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (a ? Boolean(b) : c) {}", "if (a ? b : c) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (a ? b : Boolean(c)) {}", "if (a ? b : c) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (a ? b : Boolean(c ? d : e)) {}", "if (a ? b : c ? d : e) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"for (let i = 0; (console.log(i), Boolean(i < 10)); i++) {}", "for (let i = 0; (console.log(i), i < 10); i++) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"const nullishCoalescingOperator = Boolean(bar ?? Boolean(baz));", "const nullishCoalescingOperator = Boolean(bar ?? baz);", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (a ? Boolean(b = c) : Boolean(d = e));", "if (a ? b = c : d = e);", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (a ? Boolean((b, c)) : Boolean((d, e)));", "if (a ? (b, c) : (d, e));", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"\n\tfunction * generator() {\n\t    if (a ? Boolean(yield y) : x) {\n\t        return a;\n\t    };\n\t}\n\t", "\n\tfunction * generator() {\n\t    if (a ? yield y : x) {\n\t        return a;\n\t    };\n\t}\n\t", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFixedSource(t,
				rule_testing.RunWithOptions(t, NoExtraBooleanCast, extraBooleanCastFile,
					testCase.sourceText, testCase.options), testCase.wantSource)
		})
	}
}

// The repairs that involve a double negation, which the shared harness cannot apply.
//
// `rule_testing.ExpectFixedSource` reads Fixes only, and correctly so: a Suggestion is not something the
// engine applies, and a harness that silently applied one would be asserting a rewrite that never
// happens in production. These 183 entries each carry at least one `!!` finding, whose repair is a
// Suggestion, so they are applied here instead.
//
// Upstream's expected output covers them because its own tester applies suggestions alongside fixes.
// Comparing against it is still the right assertion: it says what the offered repair would produce
// if a human took it, which is exactly what a suggestion promises.
func TestNoExtraBooleanCastSuggestsTheRightRepairs(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantSource string
		options    any
	}{
		{"if (!!foo) {}", "if (foo) {}", nil},
		{"do {} while (!!foo)", "do {} while (foo)", nil},
		{"while (!!foo) {}", "while (foo) {}", nil},
		{"!!foo ? bar : baz", "foo ? bar : baz", nil},
		{"for (; !!foo;) {}", "for (; foo;) {}", nil},
		{"!!!foo", "!foo", nil},
		{"Boolean(!!foo)", "Boolean(foo)", nil},
		{"new Boolean(!!foo)", "new Boolean(foo)", nil},
		{"Boolean(!!foo, bar)", "Boolean(foo, bar)", nil},
		{"function *foo() { yield!!a ? b : c }", "function *foo() { yield a ? b : c }", nil},
		{"function *foo() { yield!! a ? b : c }", "function *foo() { yield a ? b : c }", nil},
		{"function *foo() { yield! !a ? b : c }", "function *foo() { yield a ? b : c }", nil},
		{"function *foo() { yield !!a ? b : c }", "function *foo() { yield a ? b : c }", nil},
		{"function *foo() { yield(!!a) ? b : c }", "function *foo() { yield(a) ? b : c }", nil},
		{"function *foo() { yield/**/!!a ? b : c }", "function *foo() { yield/**/a ? b : c }", nil},
		{"x=!!a ? b : c ", "x=a ? b : c ", nil},
		{"!/**/!!foo", "!/**/foo", nil},
		{"!!!foo/**/", "!foo/**/", nil},
		{"if (!!foo || bar) {}", "if (foo || bar) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (!!foo && bar) {}", "if (foo && bar) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if ((!!foo || bar) && bat) {}", "if ((foo || bar) && bat) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (foo && !!bar) {}", "if (foo && bar) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"do {} while (!!foo || bar)", "do {} while (foo || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"while (!!foo || bar) {}", "while (foo || bar) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"!!foo && bat ? bar : baz", "foo && bat ? bar : baz", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"for (; !!foo || bar;) {}", "for (; foo || bar;) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"!!!foo || bar", "!foo || bar", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"Boolean(!!foo || bar)", "Boolean(foo || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"new Boolean(!!foo || bar)", "new Boolean(foo || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"function *foo() { yield(!!a || d) ? b : c }", "function *foo() { yield(a || d) ? b : c }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"function *foo() { yield(!! a || d) ? b : c }", "function *foo() { yield(a || d) ? b : c }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"function *foo() { yield(! !a || d) ? b : c }", "function *foo() { yield(a || d) ? b : c }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"function *foo() { yield (!!a || d) ? b : c }", "function *foo() { yield (a || d) ? b : c }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"function *foo() { yield/**/(!!a || d) ? b : c }", "function *foo() { yield/**/(a || d) ? b : c }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"x=!!a || d ? b : c ", "x=a || d ? b : c ", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"!/**/(!!foo || bar)", "!/**/(foo || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"!(!!foo || bar)/**/", "!(foo || bar)/**/", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (a && !!(b ? c : d)){}", "if (a && (b ? c : d)){}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"function *foo() { yield!!a || d ? b : c }", "function *foo() { yield a || d ? b : c }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (!!foo || bar) {}", "if (foo || bar) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (!!foo && bar) {}", "if (foo && bar) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if ((!!foo || bar) && bat) {}", "if ((foo || bar) && bat) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (foo && !!bar) {}", "if (foo && bar) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"do {} while (!!foo || bar)", "do {} while (foo || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"while (!!foo || bar) {}", "while (foo || bar) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"!!foo && bat ? bar : baz", "foo && bat ? bar : baz", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"for (; !!foo || bar;) {}", "for (; foo || bar;) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"!!!foo || bar", "!foo || bar", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"Boolean(!!foo || bar)", "Boolean(foo || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"new Boolean(!!foo || bar)", "new Boolean(foo || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"function *foo() { yield(!!a || d) ? b : c }", "function *foo() { yield(a || d) ? b : c }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"function *foo() { yield(!! a || d) ? b : c }", "function *foo() { yield(a || d) ? b : c }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"function *foo() { yield(! !a || d) ? b : c }", "function *foo() { yield(a || d) ? b : c }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"function *foo() { yield (!!a || d) ? b : c }", "function *foo() { yield (a || d) ? b : c }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"function *foo() { yield/**/(!!a || d) ? b : c }", "function *foo() { yield/**/(a || d) ? b : c }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"x=!!a || d ? b : c ", "x=a || d ? b : c ", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"!/**/(!!foo || bar)", "!/**/(foo || bar)", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"!(!!foo || bar)/**/", "!(foo || bar)/**/", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (a && !!(b ? c : d)){}", "if (a && (b ? c : d)){}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"function *foo() { yield!!a || d ? b : c }", "function *foo() { yield a || d ? b : c }", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"Boolean(!!(a, b))", "Boolean((a, b))", nil},
		{"Boolean((!!(a, b)))", "Boolean((a, b))", nil},
		{"Boolean(!(!(a, b)))", "Boolean((a, b))", nil},
		{"Boolean((!(!(a, b))))", "Boolean((a, b))", nil},
		{"Boolean(!!(a = b))", "Boolean(a = b)", nil},
		{"Boolean((!!(a = b)))", "Boolean((a = b))", nil},
		{"Boolean(!!(a === b))", "Boolean(a === b)", nil},
		{"Boolean(!!((a !== b)))", "Boolean(a !== b)", nil},
		{"Boolean(!!a.b)", "Boolean(a.b)", nil},
		{"Boolean((!!(a)))", "Boolean((a))", nil},
		{"new Boolean(!!(a, b))", "new Boolean((a, b))", nil},
		{"new Boolean((!!(a, b)))", "new Boolean((a, b))", nil},
		{"new Boolean(!(!(a, b)))", "new Boolean((a, b))", nil},
		{"new Boolean((!(!(a, b))))", "new Boolean((a, b))", nil},
		{"new Boolean(!!(a = b))", "new Boolean(a = b)", nil},
		{"new Boolean((!!(a = b)))", "new Boolean((a = b))", nil},
		{"new Boolean(!!(a === b))", "new Boolean(a === b)", nil},
		{"new Boolean(!!((a !== b)))", "new Boolean(a !== b)", nil},
		{"new Boolean(!!a.b)", "new Boolean(a.b)", nil},
		{"new Boolean((!!(a)))", "new Boolean((a))", nil},
		{"if (!!(a, b));", "if (a, b);", nil},
		{"if (!(!(a, b)));", "if (a, b);", nil},
		{"if (!!(a = b));", "if (a = b);", nil},
		{"if (!!(a > b));", "if (a > b);", nil},
		{"if (!!f(a));", "if (f(a));", nil},
		{"if (!!(f(a)));", "if (f(a));", nil},
		{"if ((!!f(a)));", "if ((f(a)));", nil},
		{"if (!!a);", "if (a);", nil},
		{"while (!!(a, b));", "while (a, b);", nil},
		{"while (!(!(a, b)));", "while (a, b);", nil},
		{"while (!!(a = b));", "while (a = b);", nil},
		{"while (!!(a > b));", "while (a > b);", nil},
		{"while (!!f(a));", "while (f(a));", nil},
		{"while (!!(f(a)));", "while (f(a));", nil},
		{"while ((!!f(a)));", "while ((f(a)));", nil},
		{"while (!!a);", "while (a);", nil},
		{"do {} while (!!(a, b));", "do {} while (a, b);", nil},
		{"do {} while (!(!(a, b)));", "do {} while (a, b);", nil},
		{"do {} while (!!(a = b));", "do {} while (a = b);", nil},
		{"do {} while (!!(a > b));", "do {} while (a > b);", nil},
		{"do {} while (!!f(a));", "do {} while (f(a));", nil},
		{"do {} while (!!(f(a)));", "do {} while (f(a));", nil},
		{"do {} while ((!!f(a)));", "do {} while ((f(a)));", nil},
		{"do {} while (!!a);", "do {} while (a);", nil},
		{"for (; !!(a, b););", "for (; a, b;);", nil},
		{"for (; !(!(a, b)););", "for (; a, b;);", nil},
		{"for (; !!(a = b););", "for (; a = b;);", nil},
		{"for (; !!(a > b););", "for (; a > b;);", nil},
		{"for (; !!f(a););", "for (; f(a););", nil},
		{"for (; !!(f(a)););", "for (; f(a););", nil},
		{"for (; (!!f(a)););", "for (; (f(a)););", nil},
		{"for (; !!a;);", "for (; a;);", nil},
		{"!!(a, b) ? c : d", "(a, b) ? c : d", nil},
		{"(!!(a, b)) ? c : d", "(a, b) ? c : d", nil},
		{"!!(a = b) ? c : d", "(a = b) ? c : d", nil},
		{"!!(a ? b : c) ? d : e", "(a ? b : c) ? d : e", nil},
		{"!!(a || b) ? c : d", "a || b ? c : d", nil},
		{"!!(a === b) ? c : d", "a === b ? c : d", nil},
		{"!!((a !== b)) ? c : d", "a !== b ? c : d", nil},
		{"!!+a ? b : c", "+a ? b : c", nil},
		{"!!+(a) ? b : c", "+(a) ? b : c", nil},
		{"!!f(a) ? b : c", "f(a) ? b : c", nil},
		{"(!!f(a)) ? b : c", "(f(a)) ? b : c", nil},
		{"!!a ? b : c", "a ? b : c", nil},
		{"!!!(a, b)", "!(a, b)", nil},
		{"!!!(a = b)", "!(a = b)", nil},
		{"!!(!(a += b))", "!(a += b)", nil},
		{"!(!!(a += b))", "!(a += b)", nil},
		{"!!!(a || b)", "!(a || b)", nil},
		{"!!!(a && b)", "!(a && b)", nil},
		{"!!!(a != b)", "!(a != b)", nil},
		{"!!!(a === b)", "!(a === b)", nil},
		{"!!!(a - b)", "!(a - b)", nil},
		{"!!!(a ** b)", "!(a ** b)", nil},
		{"async function f() { !!!(await a) }", "async function f() { !await a }", nil},
		{"!!!!a", "!!a", nil},
		{"!!(!(!a))", "!!a", nil},
		{"!!!+a", "!+a", nil},
		{"!!(!+a)", "!+a", nil},
		{"!(!!+a)", "!(+a)", nil},
		{"!!!f(a)", "!f(a)", nil},
		{"!!!(f(a))", "!f(a)", nil},
		{"!!!a", "!a", nil},
		{"if (!!(a, b) || !!(c, d)) {}", "if ((a, b) || (c, d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if ((!!((a, b))) || (!!((c, d)))) {}", "if ((a, b) || (c, d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (!!(a, b) && !!(c, d)) {}", "if ((a, b) && (c, d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if ((!!((a, b))) && (!!((c, d)))) {}", "if ((a, b) && (c, d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (!!(a = b) || !!(c = d)) {}", "if ((a = b) || (c = d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (!!(a >>= b) && !!(c >>= d)) {}", "if ((a >>= b) && (c >>= d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (!!(a ? b : c) || !!(d ? e : f)) {}", "if ((a ? b : c) || (d ? e : f)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (!!(a ? b : c) && !!(d ? e : f)) {}", "if ((a ? b : c) && (d ? e : f)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (!!(a || b) || !!(c || d)) {}", "if (a || b || (c || d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (!!(a || b) && !!(c || d)) {}", "if ((a || b) && (c || d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (!!(a && b) || !!(c && d)) {}", "if (a && b || c && d) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (!!(a && b) && !!(c && d)) {}", "if (a && b && (c && d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (!!(a !== b) || !!(c !== d)) {}", "if (a !== b || c !== d) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (!!(a === b) && !!(c === d)) {}", "if (a === b && c === d) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (!!(a > b) || !!(c < d)) {}", "if (a > b || c < d) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (!!f(a) && !!b.c) {}", "if (f(a) && b.c) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (Boolean(a) || !!b) {}", "if (a || b) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (!!a && Boolean(b)) {}", "if (a && b) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if ((!!a) || (Boolean(b))) {}", "if ((a) || (b)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (!!(a, b) || !!(c, d)) {}", "if ((a, b) || (c, d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if ((!!((a, b))) || (!!((c, d)))) {}", "if ((a, b) || (c, d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (!!(a, b) && !!(c, d)) {}", "if ((a, b) && (c, d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if ((!!((a, b))) && (!!((c, d)))) {}", "if ((a, b) && (c, d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (!!(a = b) || !!(c = d)) {}", "if ((a = b) || (c = d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (!!(a >>= b) && !!(c >>= d)) {}", "if ((a >>= b) && (c >>= d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (!!(a ? b : c) || !!(d ? e : f)) {}", "if ((a ? b : c) || (d ? e : f)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (!!(a ? b : c) && !!(d ? e : f)) {}", "if ((a ? b : c) && (d ? e : f)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (!!(a || b) || !!(c || d)) {}", "if (a || b || (c || d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (!!(a || b) && !!(c || d)) {}", "if ((a || b) && (c || d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (!!(a && b) || !!(c && d)) {}", "if (a && b || c && d) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (!!(a && b) && !!(c && d)) {}", "if (a && b && (c && d)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (!!(a !== b) || !!(c !== d)) {}", "if (a !== b || c !== d) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (!!(a === b) && !!(c === d)) {}", "if (a === b && c === d) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (!!(a > b) || !!(c < d)) {}", "if (a > b || c < d) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (!!f(a) && !!b.c) {}", "if (f(a) && b.c) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (Boolean(a) || !!b) {}", "if (a || b) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if (!!a && Boolean(b)) {}", "if (a && b) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"if ((!!a) || (Boolean(b))) {}", "if ((a) || (b)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"const ternary = Boolean(bar ? !!baz : bat);", "const ternary = Boolean(bar ? baz : bat);", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
		{"const commaOperator = Boolean((bar, baz, !!bat));", "const commaOperator = Boolean((bar, baz, bat));", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true}},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoExtraBooleanCast, extraBooleanCastFile,
				testCase.sourceText, testCase.options)
			if got := applyEveryRepair(testCase.sourceText, result); got != testCase.wantSource {
				t.Fatalf("applying the repairs gave %q, wanted %q", got, testCase.wantSource)
			}
		})
	}
}

// applyEveryRepair rewrites source with every repair a run proposed, back to front.
//
// Both kinds, because upstream's expected-output vector was produced by a tester that applies
// suggestions alongside fixes, so comparing against it with fixes only would fail on the 177 entries
// whose only repair is the `!!` removal.
//
// Back to front so no surviving edit is measured against text that moved under it, which is the same
// ordering `internal/fix` uses. Overlapping repairs are skipped rather than applied: a nested cast
// such as `Boolean(Boolean(a))` reports twice with the outer repair spanning the inner one, and
// upstream's fixer lands one pass too.
func applyEveryRepair(sourceText string, result rule_testing.Result) string {
	type edit struct {
		start int
		end   int
		text  string
	}

	edits := []edit{}
	for _, diagnostic := range result.Diagnostics {
		for _, repair := range diagnostic.Fixes {
			edits = append(edits, edit{repair.Range.Pos(), repair.Range.End(), repair.Text})
		}
		for _, suggestion := range diagnostic.Suggestions {
			for _, repair := range suggestion.Fixes {
				edits = append(edits, edit{repair.Range.Pos(), repair.Range.End(), repair.Text})
			}
		}
	}
	sort.Slice(edits, func(first int, second int) bool {
		if edits[first].start != edits[second].start {
			return edits[first].start < edits[second].start
		}
		return edits[first].end > edits[second].end
	})

	// Keep the first of any overlapping group, which after the sort above is the outermost.
	kept := []edit{}
	lastEnd := -1
	for _, candidate := range edits {
		if candidate.start < lastEnd {
			continue
		}
		kept = append(kept, candidate)
		lastEnd = candidate.end
	}

	rewritten := sourceText
	for index := len(kept) - 1; index >= 0; index-- {
		rewritten = rewritten[:kept[index].start] + kept[index].text + rewritten[kept[index].end:]
	}
	return rewritten
}

// The span of every finding, which ExpectFindings cannot see.
//
// A rule that reports the right message at the wrong offsets passes the whole corpus above. The
// expected text here is sliced out of upstream's own snapshot, which prints a line, a column and an
// underline for every one of its 507 diagnostics, so these are measurements rather than guesses.
//
// The `!!` cases are the ones that would catch an off-by-one: the finding covers the outer `!` and
// everything under it, so reporting the inner negation instead is a span one byte to the right that
// still lands inside the same expression.
func TestNoExtraBooleanCastReportsTheRightSpan(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		options    any
		wantSpans  []string
	}{
		{"if (!!foo) {}", nil, []string{"!!foo"}},
		{"do {} while (!!foo)", nil, []string{"!!foo"}},
		{"!!!foo", nil, []string{"!!foo"}},
		{"if (Boolean(foo)) {}", nil, []string{"Boolean(foo)"}},
		// Only the inner negation. The outer call is not itself in a coercing position, so it is a
		// context rather than a finding, and a rule reporting both would be reporting the outer one
		// for no reason. Upstream's snapshot records exactly one diagnostic here, at column 9.
		{"Boolean(!!foo)", nil, []string{"!!foo"}},
		{"new Boolean(!!foo)", nil, []string{"!!foo"}},
		{"!!!!a", nil, []string{"!!!a", "!!a"}},
		{"!Boolean(foo)", nil, []string{"Boolean(foo)"}},
		// The same asymmetry for the call form: the inner call is the first argument of the outer
		// one, which coerces it, while the outer call sits in a statement position that does not.
		{"Boolean(Boolean(foo))", nil, []string{"Boolean(foo)"}},
		{"if (Boolean?.(foo)) {};", nil, []string{"Boolean?.(foo)"}},
		{"if ((Boolean(f(a))));", nil, []string{"Boolean(f(a))"}},
		{"if (!!a && Boolean(b)) {}", NoExtraBooleanCastOptions{EnforceForInnerExpressions: true},
			[]string{"!!a", "Boolean(b)"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, NoExtraBooleanCast, extraBooleanCastFile,
				testCase.sourceText, testCase.options)
			if len(result.Diagnostics) != len(testCase.wantSpans) {
				t.Fatalf("wanted %d findings, got %d", len(testCase.wantSpans),
					len(result.Diagnostics))
			}
			for index, diagnostic := range result.Diagnostics {
				reported := testCase.sourceText[diagnostic.Range.Pos():diagnostic.Range.End()]
				if reported != testCase.wantSpans[index] {
					t.Fatalf("finding %d pointed at %q, wanted %q", index, reported,
						testCase.wantSpans[index])
				}
			}
		})
	}
}

// The two findings carry different repair kinds, and collapsing them is the defect this pins.
//
// `Boolean(x)` unwrapping is a Fix the engine applies unattended. `!!x` removal is a Suggestion a
// human chooses. Upstream declares exactly this with `conditional_fix_or_conditional_suggestion`,
// which is the only occurrence of that declaration in its 946 rules, so there is no second rule to
// copy the shape from and nothing else in this package would notice if the two were swapped.
func TestNoExtraBooleanCastSplitsFixFromSuggestion(t *testing.T) {
	t.Parallel()

	booleanCall := rule_testing.Run(t, NoExtraBooleanCast, extraBooleanCastFile, "if (Boolean(foo)) {}")
	if len(booleanCall.Diagnostics) != 1 {
		t.Fatalf("wanted one finding for the Boolean call, got %d", len(booleanCall.Diagnostics))
	}
	if len(booleanCall.Diagnostics[0].Fixes) != 1 {
		t.Fatalf("the Boolean call should carry exactly one Fix, got %d",
			len(booleanCall.Diagnostics[0].Fixes))
	}
	if len(booleanCall.Diagnostics[0].Suggestions) != 0 {
		t.Fatalf("the Boolean call should carry no Suggestion, got %d",
			len(booleanCall.Diagnostics[0].Suggestions))
	}

	doubleNegation := rule_testing.Run(t, NoExtraBooleanCast, extraBooleanCastFile, "if (!!foo) {}")
	if len(doubleNegation.Diagnostics) != 1 {
		t.Fatalf("wanted one finding for the double negation, got %d",
			len(doubleNegation.Diagnostics))
	}
	if len(doubleNegation.Diagnostics[0].Fixes) != 0 {
		t.Fatalf("the double negation should carry no Fix, got %d",
			len(doubleNegation.Diagnostics[0].Fixes))
	}
	if len(doubleNegation.Diagnostics[0].Suggestions) != 1 {
		t.Fatalf("the double negation should carry exactly one Suggestion, got %d",
			len(doubleNegation.Diagnostics[0].Suggestions))
	}
}

// The findings upstream emits with no repair at all, which the fix vector cannot show.
//
// Upstream's `expect_fix` vector is an allowlist rather than a census, so a case missing from it
// proves nothing on its own. These two are read from the rule body instead: more than one argument,
// or a spread, hits `fixer.noop()`, because there is no single expression `Boolean(...foo)` reduces
// to without knowing what `foo` holds.
func TestNoExtraBooleanCastDeclinesToRepairMultipleArguments(t *testing.T) {
	t.Parallel()

	cases := []string{"!Boolean(...foo);", "!Boolean(foo, bar());"}

	for _, sourceText := range cases {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoExtraBooleanCast, extraBooleanCastFile, sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
			}
			if len(result.Diagnostics[0].Fixes) != 0 ||
				len(result.Diagnostics[0].Suggestions) != 0 {
				t.Fatalf("wanted a finding with no repair, got %d fixes and %d suggestions",
					len(result.Diagnostics[0].Fixes), len(result.Diagnostics[0].Suggestions))
			}
		})
	}
}

// The deprecated option spelling reaches the same field, which no verbatim case can show.
//
// Upstream's corpus runs both spellings but the harness above decodes neither: the cases arrive as
// an already-built options struct. This is the only place the JSON path is exercised, and without it
// a port that simply dropped `enforceForLogicalOperands` from the decoder passes all 441 fail cases.
func TestNoExtraBooleanCastAcceptsTheDeprecatedOptionSpelling(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		rawJSON string
		want    bool
	}{
		{"the current spelling", `{"enforceForInnerExpressions": true}`, true},
		{"the deprecated spelling", `{"enforceForLogicalOperands": true}`, true},
		{"neither spelling", `{}`, false},
		{"the current spelling off", `{"enforceForInnerExpressions": false}`, false},
		{"the deprecated spelling off", `{"enforceForLogicalOperands": false}`, false},
		// serde resolves an alias only when the canonical name is absent, so the canonical name
		// wins rather than whichever was written last.
		{"both spellings, canonical wins",
			`{"enforceForInnerExpressions": false, "enforceForLogicalOperands": true}`, false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			decoded, err := rule.DecodeOptionsInto[NoExtraBooleanCastOptions]()([]byte(testCase.rawJSON))
			if err != nil {
				t.Fatalf("decoding %s: %v", testCase.rawJSON, err)
			}
			options, ok := decoded.(NoExtraBooleanCastOptions)
			if !ok {
				t.Fatalf("decoded to %T, wanted NoExtraBooleanCastOptions", decoded)
			}
			if options.EnforceForInnerExpressions != testCase.want {
				t.Fatalf("EnforceForInnerExpressions was %v, wanted %v",
					options.EnforceForInnerExpressions, testCase.want)
			}
		})
	}
}

// A case written because a mutation survived the imported corpus, not because upstream has it.
//
// `??` cannot sit beside `&&` or `||` without parentheses. That is a grammar restriction rather than
// a precedence one, and in this tree it cannot be expressed as a precedence comparison at all:
// `ast.GetBinaryOperatorPrecedence` gives `??` and `||` the same value, 5, so the ordinary
// comparisons see them as equal and let both through. oxc's own precedence table separates them,
// which is why its version of this check reads as one more comparison and this one has to name the
// operators.
//
// Upstream's corpus reaches the guard only through `if (a ?? Boolean(b || c)) {}`, where the
// replacement is on the right at equal precedence, so the left-associativity branch above already
// returns true and the guard is never the reason. Deleting the guard therefore changed nothing any
// of the 382 repair comparisons could see.
//
// `&&` distinguishes them: it is precedence 6 against `??`'s 5, so neither earlier branch fires, and
// without the guard the repair emits `a ?? b && c`. That does not merely mean something different,
// it does not parse, which is the one failure an unattended fixer structurally cannot refuse.
func TestNoExtraBooleanCastParenthesisesLogicalOperandsUnderCoalescing(t *testing.T) {
	t.Parallel()

	const sourceText = "if (a ?? Boolean(b && c)) {}"

	result := rule_testing.RunWithOptions(t, NoExtraBooleanCast, extraBooleanCastFile, sourceText,
		NoExtraBooleanCastOptions{EnforceForInnerExpressions: true})
	rule_testing.ExpectFixedSource(t, result, "if (a ?? (b && c)) {}")
}

// A second case written for a survivor, covering the other half of the token-boundary padding.
//
// Upstream's corpus exercises the *leading* half thoroughly, through the `yield!!a`, `void!Boolean()`
// and `typeof!Boolean()` family, where the cast being removed was the only thing separating a
// keyword from its operand. It never exercises the trailing half, so deleting that branch changed
// none of the 382 repair comparisons.
//
// `!Boolean()in x` reaches it. The whole `!Boolean()` collapses to the literal `true`, and the next
// character begins the `in` operator with no space in front of it, so without the trailing pad the
// repair writes `truein x`. That parses, as a reference to an identifier named `truein`, and then
// fails to compile against a name that does not exist. A fix producing source that parses and does
// not compile is the one failure the edit engine structurally cannot refuse, which is what makes
// this worth a fixture rather than a comment.
func TestNoExtraBooleanCastPadsTheTrailingTokenBoundary(t *testing.T) {
	t.Parallel()

	const sourceText = "!Boolean()in x"

	result := rule_testing.Run(t, NoExtraBooleanCast, extraBooleanCastFile, sourceText)
	rule_testing.ExpectFixedSource(t, result, "true in x")
}
