# Regex remaining claim

Branch: codex/regex-remaining. This claim follows the separately requested surrogate-split fix (866f466, 310c50e); no remaining-library implementation precedes it.

Baseline: frozen ca96d50 plus the const decoded-input merge correction, stock TypeScript 6.0.3, Node v24.19.0 Linux x64, --adapt --jobs 1 --timeout 60s. Independent process shards avoid sharing the runner TypeScript stream. All 2,189 paths are unique; zero failures or crashes. The earlier live-checkout run is discarded.

| Filter | Pass | Refused | Not TypeScript | Skipped | Total |
|---|---:|---:|---:|---:|---:|
| built-ins/RegExp | 914 | 226 | 223 | 516 | 1879 |
| language/literals/regexp | 16 | 8 | 4 | 212 | 240 |
| annexB/built-ins/RegExp | 1 | 13 | 23 | 25 | 62 |
| annexB/language/literals/regexp | 0 | 0 | 8 | 0 | 8 |

The additional read-only TypeScript audit finds five tsc-invalid programs among the 247 NotYet refusals. They remain in the runner’s measured refused count, are listed below, and are excluded from implementation. No verdict/classify policy changes.

## Work, largest library family first

1. Exact exception-constructor harness (94). The prelude already checks constructor identity; the runner retains a stale refusal. Keep original upstream Node validation and prove a wrong-constructor mutant fails.
2. Regex protocol, toString and replacement callback effects (21). Extend the closed operation whitelist; callbacks retain ordinary call escape/clobber effects. No generic future-method allowance.
3. Runtime patterns / flags (15). Inventory their real inputs. An exact native runtime ECMAScript compiler is a separate substantial dependency; retain explicit refusal rather than freeze dynamic values or run a Go/Node helper.
4. Annex B RegExp.compile (8). Implement constant compilation with same-object mutation, state reset and rollback on SyntaxError; cloning after mutation must not use stale source proofs.
5. Intrinsic regex method metadata (13). Investigate pristine intrinsic observations separately from detached methods, Function.call, for-in and prototype language dependencies. Admit only proven observations if sound.

## Full valid refusal inventory

### 94: not yet: constructor-identity assertion in RegExp harness

Classification: library.

- built-ins/RegExp/15.10.2.15-6-1.js
- built-ins/RegExp/15.10.2.5-3-1.js
- built-ins/RegExp/15.10.4.1-2.js
- built-ins/RegExp/15.10.4.1-3.js
- built-ins/RegExp/duplicate-flags.js
- built-ins/RegExp/duplicate-named-capturing-groups-syntax.js
- built-ins/RegExp/early-err-modifiers-code-point-repeat-i-1.js
- built-ins/RegExp/early-err-modifiers-code-point-repeat-i-2.js
- built-ins/RegExp/early-err-modifiers-other-code-point-arbitrary.js
- built-ins/RegExp/early-err-modifiers-other-code-point-combining-i.js
- built-ins/RegExp/early-err-modifiers-other-code-point-combining-m.js
- built-ins/RegExp/early-err-modifiers-other-code-point-combining-s.js
- built-ins/RegExp/early-err-modifiers-other-code-point-d.js
- built-ins/RegExp/early-err-modifiers-other-code-point-g.js
- built-ins/RegExp/early-err-modifiers-other-code-point-non-display-1.js
- built-ins/RegExp/early-err-modifiers-other-code-point-non-display-2.js
- built-ins/RegExp/early-err-modifiers-other-code-point-non-flag.js
- built-ins/RegExp/early-err-modifiers-other-code-point-u.js
- built-ins/RegExp/early-err-modifiers-other-code-point-uppercase-I.js
- built-ins/RegExp/early-err-modifiers-other-code-point-y.js
- built-ins/RegExp/early-err-modifiers-other-code-point-zwj.js
- built-ins/RegExp/early-err-modifiers-other-code-point-zwnbsp.js
- built-ins/RegExp/early-err-modifiers-other-code-point-zwnj.js
- built-ins/RegExp/early-err-modifiers-should-not-case-fold-i.js
- built-ins/RegExp/early-err-modifiers-should-not-case-fold-m.js
- built-ins/RegExp/early-err-modifiers-should-not-case-fold-s.js
- built-ins/RegExp/early-err-modifiers-should-not-unicode-case-fold-i.js
- built-ins/RegExp/early-err-modifiers-should-not-unicode-case-fold-s.js
- built-ins/RegExp/named-groups/non-unicode-property-names-invalid.js
- built-ins/RegExp/named-groups/unicode-property-names-invalid.js
- built-ins/RegExp/prototype/unicodeSets/uv-flags-constructor.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-add-remove-i.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-add-remove-m.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-add-remove-multi-duplicate.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-add-remove-s-escape.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-add-remove-s.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-both-empty.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-code-point-repeat-i-1.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-code-point-repeat-i-2.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-other-code-point-arbitrary.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-other-code-point-combining-i.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-other-code-point-combining-m.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-other-code-point-combining-s.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-other-code-point-d.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-other-code-point-g.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-other-code-point-non-display-1.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-other-code-point-non-display-2.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-other-code-point-non-flag.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-other-code-point-u.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-other-code-point-uppercase-I.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-other-code-point-y.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-other-code-point-zwj.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-other-code-point-zwnbsp.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-other-code-point-zwnj.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-reverse-add-remove-i.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-reverse-add-remove-m.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-reverse-add-remove-multi-duplicate.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-reverse-add-remove-s-escape.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-reverse-add-remove-s.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-reverse-code-point-repeat-i-1.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-reverse-code-point-repeat-i-2.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-reverse-other-code-point-arbitrary.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-reverse-other-code-point-combining-i.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-reverse-other-code-point-combining-m.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-reverse-other-code-point-combining-s.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-reverse-other-code-point-d.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-reverse-other-code-point-g.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-reverse-other-code-point-non-display-1.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-reverse-other-code-point-non-display-2.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-reverse-other-code-point-non-flag.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-reverse-other-code-point-u.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-reverse-other-code-point-uppercase-I.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-reverse-other-code-point-y.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-reverse-other-code-point-zwj.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-reverse-other-code-point-zwnbsp.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-reverse-other-code-point-zwnj.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-reverse-should-not-case-fold-i.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-reverse-should-not-case-fold-m.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-reverse-should-not-case-fold-s.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-reverse-should-not-unicode-case-fold-i.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-reverse-should-not-unicode-case-fold-s.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-should-not-case-fold-i.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-should-not-case-fold-m.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-should-not-case-fold-s.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-should-not-unicode-case-fold-i.js
- built-ins/RegExp/syntax-err-arithmetic-modifiers-should-not-unicode-case-fold-s.js
- built-ins/RegExp/unicode_restricted_brackets.js
- built-ins/RegExp/unicode_restricted_character_class_escape.js
- built-ins/RegExp/unicode_restricted_identity_escape_u.js
- built-ins/RegExp/unicode_restricted_identity_escape_x.js
- built-ins/RegExp/unicode_restricted_incomplete_quantifier.js
- built-ins/RegExp/unicode_restricted_octal_escape.js
- built-ins/RegExp/unicode_restricted_quantifiable_assertion.js
- built-ins/RegExp/unicode_restricted_quantifier_without_atom.js

### 21: refuses Function.caller, a mutable field of type Function, which can reach back to the Function holding it: a cycle reference counting can't free, and a write in the top level may close one (ir.RegExpCall is a node the cycle finder doesn't know)

Classification: library.

- built-ins/RegExp/S15.10.1_A1_T1.js
- built-ins/RegExp/S15.10.1_A1_T10.js
- built-ins/RegExp/S15.10.1_A1_T11.js
- built-ins/RegExp/S15.10.1_A1_T12.js
- built-ins/RegExp/S15.10.1_A1_T13.js
- built-ins/RegExp/S15.10.1_A1_T14.js
- built-ins/RegExp/S15.10.1_A1_T15.js
- built-ins/RegExp/S15.10.1_A1_T16.js
- built-ins/RegExp/S15.10.1_A1_T2.js
- built-ins/RegExp/S15.10.1_A1_T3.js
- built-ins/RegExp/S15.10.1_A1_T4.js
- built-ins/RegExp/S15.10.1_A1_T5.js
- built-ins/RegExp/S15.10.1_A1_T6.js
- built-ins/RegExp/S15.10.1_A1_T7.js
- built-ins/RegExp/S15.10.1_A1_T8.js
- built-ins/RegExp/S15.10.1_A1_T9.js
- built-ins/RegExp/S15.10.3.1_A2_T1.js
- built-ins/RegExp/S15.10.4.1_A5_T4.js
- built-ins/RegExp/S15.10.4.1_A9_T1.js
- built-ins/RegExp/S15.10.4.1_A9_T2.js
- built-ins/RegExp/S15.10.4.1_A9_T3.js

### 11: not yet: RegExp with nonconstant flags: the native runtime has no ECMAScript pattern compiler

Classification: library (runtime compiler dependency).

- built-ins/RegExp/S15.10.4.1_A1_T2.js
- built-ins/RegExp/S15.10.4.1_A1_T5.js
- built-ins/RegExp/S15.10.4.1_A4_T5.js
- built-ins/RegExp/named-groups/functional-replace-non-global.js
- built-ins/RegExp/named-groups/string-replace-escaped.js
- built-ins/RegExp/named-groups/string-replace-get.js
- built-ins/RegExp/named-groups/string-replace-missing.js
- built-ins/RegExp/named-groups/string-replace-nocaptures.js
- built-ins/RegExp/named-groups/string-replace-numbered.js
- built-ins/RegExp/named-groups/string-replace-unclosed.js
- built-ins/RegExp/named-groups/string-replace-undefined.js

### 8: not yet: RegExp.compile

Classification: library.

- annexB/built-ins/RegExp/prototype/compile/duplicate-named-capturing-groups-syntax.js
- annexB/built-ins/RegExp/prototype/compile/flags-string-invalid.js
- annexB/built-ins/RegExp/prototype/compile/flags-to-string.js
- annexB/built-ins/RegExp/prototype/compile/pattern-string-invalid-u.js
- annexB/built-ins/RegExp/prototype/compile/pattern-string-invalid.js
- annexB/built-ins/RegExp/prototype/compile/pattern-string-u.js
- annexB/built-ins/RegExp/prototype/compile/pattern-string.js
- annexB/built-ins/RegExp/prototype/flags/order-after-compile.js

### 5: refuses a method read as a value (exec would lose its object, and this with it)

Classification: library metadata / language detached methods.

Language handoff (detached receiver): `const fn = /a/.exec; fn.call(/a/, "a");`

- built-ins/RegExp/prototype/exec/S15.10.6.2_A11.js
- built-ins/RegExp/prototype/exec/S15.10.6.2_A2_T10.js
- built-ins/RegExp/prototype/exec/S15.10.6.2_A2_T3.js
- built-ins/RegExp/prototype/exec/S15.10.6.2_A6.js
- built-ins/RegExp/prototype/exec/S15.10.6.2_A8.js

### 5: refuses a method read as a value (test would lose its object, and this with it)

Classification: library metadata / language detached methods.

Language handoff (detached receiver): `const fn = /a/.test; fn.call(/a/, "a");`

- built-ins/RegExp/prototype/test/S15.10.6.3_A11.js
- built-ins/RegExp/prototype/test/S15.10.6.3_A2_T10.js
- built-ins/RegExp/prototype/test/S15.10.6.3_A2_T3.js
- built-ins/RegExp/prototype/test/S15.10.6.3_A6.js
- built-ins/RegExp/prototype/test/S15.10.6.3_A8.js

### 4: not yet: RegExp with a nonconstant pattern: the native runtime has no ECMAScript pattern compiler

Classification: library (runtime compiler dependency).

- built-ins/RegExp/S15.10.2.10_A2.1_T1.js
- built-ins/RegExp/S15.10.2.10_A2.1_T2.js
- built-ins/RegExp/S15.10.4.1_A8_T2.js
- built-ins/RegExp/quantifier-integer-limit.js

### 3: refuses a method read as a value (toString would lose its object, and this with it)

Classification: library metadata / language detached methods.

Language handoff (detached receiver): `const fn = /a/.toString; fn.call(/a/);`

- built-ins/RegExp/prototype/15.10.6.js
- built-ins/RegExp/prototype/toString/S15.10.6.4_A11.js
- built-ins/RegExp/prototype/toString/S15.10.6.4_A8.js

### 41: not yet: a BinaryExpression with a string and a value

Classification: language.

Language handoff reproducer: `const m = /a/.exec('a'); console.log('match=' + (m && m[0]));`

- built-ins/RegExp/S15.10.2.15_A1_T1.js
- built-ins/RegExp/S15.10.2.15_A1_T10.js
- built-ins/RegExp/S15.10.2.15_A1_T11.js
- built-ins/RegExp/S15.10.2.15_A1_T12.js
- built-ins/RegExp/S15.10.2.15_A1_T13.js
- built-ins/RegExp/S15.10.2.15_A1_T14.js
- built-ins/RegExp/S15.10.2.15_A1_T15.js
- built-ins/RegExp/S15.10.2.15_A1_T16.js
- built-ins/RegExp/S15.10.2.15_A1_T17.js
- built-ins/RegExp/S15.10.2.15_A1_T18.js
- built-ins/RegExp/S15.10.2.15_A1_T19.js
- built-ins/RegExp/S15.10.2.15_A1_T2.js
- built-ins/RegExp/S15.10.2.15_A1_T20.js
- built-ins/RegExp/S15.10.2.15_A1_T21.js
- built-ins/RegExp/S15.10.2.15_A1_T22.js
- built-ins/RegExp/S15.10.2.15_A1_T23.js
- built-ins/RegExp/S15.10.2.15_A1_T24.js
- built-ins/RegExp/S15.10.2.15_A1_T25.js
- built-ins/RegExp/S15.10.2.15_A1_T26.js
- built-ins/RegExp/S15.10.2.15_A1_T27.js
- built-ins/RegExp/S15.10.2.15_A1_T28.js
- built-ins/RegExp/S15.10.2.15_A1_T29.js
- built-ins/RegExp/S15.10.2.15_A1_T3.js
- built-ins/RegExp/S15.10.2.15_A1_T30.js
- built-ins/RegExp/S15.10.2.15_A1_T31.js
- built-ins/RegExp/S15.10.2.15_A1_T32.js
- built-ins/RegExp/S15.10.2.15_A1_T33.js
- built-ins/RegExp/S15.10.2.15_A1_T34.js
- built-ins/RegExp/S15.10.2.15_A1_T35.js
- built-ins/RegExp/S15.10.2.15_A1_T36.js
- built-ins/RegExp/S15.10.2.15_A1_T37.js
- built-ins/RegExp/S15.10.2.15_A1_T38.js
- built-ins/RegExp/S15.10.2.15_A1_T39.js
- built-ins/RegExp/S15.10.2.15_A1_T4.js
- built-ins/RegExp/S15.10.2.15_A1_T40.js
- built-ins/RegExp/S15.10.2.15_A1_T41.js
- built-ins/RegExp/S15.10.2.15_A1_T5.js
- built-ins/RegExp/S15.10.2.15_A1_T6.js
- built-ins/RegExp/S15.10.2.15_A1_T7.js
- built-ins/RegExp/S15.10.2.15_A1_T8.js
- built-ins/RegExp/S15.10.2.15_A1_T9.js

### 8: refuses var

Classification: language.

Language handoff reproducer: `var r = /a/; console.log(r.test('a'));`

- built-ins/RegExp/S15.10.2.10_A1.1_T1.js
- built-ins/RegExp/S15.10.2.10_A1.2_T1.js
- built-ins/RegExp/S15.10.2.10_A1.3_T1.js
- built-ins/RegExp/S15.10.2.10_A1.4_T1.js
- built-ins/RegExp/S15.10.2.10_A1.5_T1.js
- built-ins/RegExp/S15.10.2.10_A3.1_T1.js
- built-ins/RegExp/S15.10.2.10_A4.1_T1.js
- built-ins/RegExp/S15.10.2.11_A1_T1.js

### 6: not yet: a BinaryExpression with a string and a boolean

Classification: language.

Language handoff reproducer: `console.log('match=' + /a/.test('a'));`

- language/literals/regexp/S7.8.5_A3.1_T1.js
- language/literals/regexp/S7.8.5_A3.1_T2.js
- language/literals/regexp/S7.8.5_A3.1_T3.js
- language/literals/regexp/S7.8.5_A3.1_T4.js
- language/literals/regexp/S7.8.5_A3.1_T5.js
- language/literals/regexp/S7.8.5_A3.1_T6.js

### 6: not yet: a BinaryExpression with a value and a string

Classification: language.

Language handoff reproducer: `const m = /a/.exec('a'); console.log((m && m[0]) + '!');`

- built-ins/RegExp/S15.10.2.11_A1_T4.js
- built-ins/RegExp/S15.10.2.11_A1_T5.js
- built-ins/RegExp/S15.10.2.11_A1_T6.js
- built-ins/RegExp/S15.10.2.11_A1_T7.js
- built-ins/RegExp/S15.10.2.11_A1_T8.js
- built-ins/RegExp/S15.10.2.11_A1_T9.js

### 5: not yet: a computed class base; name the base class directly

Classification: language.

Language handoff reproducer: `class Derived extends (class extends RegExp {}) {}`

- annexB/built-ins/RegExp/legacy-accessors/input/this-subclass-constructor.js
- annexB/built-ins/RegExp/legacy-accessors/lastMatch/this-subclass-constructor.js
- annexB/built-ins/RegExp/legacy-accessors/lastParen/this-subclass-constructor.js
- annexB/built-ins/RegExp/legacy-accessors/leftContext/this-subclass-constructor.js
- annexB/built-ins/RegExp/legacy-accessors/rightContext/this-subclass-constructor.js

### 4: not yet: an array of never

Classification: language.

Language handoff reproducer: `const errors = []; errors.push('error');`

- built-ins/RegExp/prototype/exec/S15.10.6.2_A3_T1.js
- built-ins/RegExp/prototype/exec/S15.10.6.2_A3_T5.js
- built-ins/RegExp/prototype/exec/S15.10.6.2_A3_T6.js
- built-ins/RegExp/prototype/exec/S15.10.6.2_A3_T7.js

### 4: refuses inherited library member prototype read as an own field

Classification: language.

Language handoff reproducer: `console.log(RegExp.prototype.source);`

- built-ins/RegExp/prototype/global/S15.10.7.2_A8.js
- built-ins/RegExp/prototype/ignoreCase/S15.10.7.3_A8.js
- built-ins/RegExp/prototype/multiline/S15.10.7.4_A8.js
- built-ins/RegExp/prototype/no-regexp-matcher.js

### 3: not yet: instanceof against a value that isn't a declared class

Classification: language.

Language handoff reproducer: `console.log(/a/.exec('a') instanceof Array);`

- built-ins/RegExp/prototype/exec/S15.10.6.2_A1_T1.js
- built-ins/RegExp/prototype/exec/S15.10.6.2_A1_T6.js
- language/literals/regexp/S7.8.5_A4.1.js

### 3: refuses a number as a condition

Classification: language.

Language handoff reproducer: `const m = /a/.exec('a'); if (m !== null && m.index) console.log('nonzero');`

- built-ins/RegExp/prototype/exec/S15.10.6.2_A3_T2.js
- built-ins/RegExp/prototype/exec/S15.10.6.2_A3_T3.js
- built-ins/RegExp/prototype/exec/S15.10.6.2_A3_T4.js

### 2: not yet: a value of type any

Classification: language.

Language handoff reproducer: `const value: any = /a/; console.log(value);`

- built-ins/RegExp/character-class-escape-non-whitespace.js
- language/literals/regexp/y-assertion-start.js

### 2: not yet: hasOwnProperty on a function (its prototype property depends on whether it was declared or made as an arrow)

Classification: language.

Language handoff reproducer: `console.log(RegExp.hasOwnProperty('prototype'));`

- built-ins/RegExp/prototype/S15.10.5.1_A1.js
- built-ins/RegExp/prototype/S15.10.5.1_A2.js

### 2: refuses Object.defineProperty

Classification: language.

Language handoff reproducer: `Object.defineProperty(/a/, 'lastIndex', {writable: false});`

- built-ins/RegExp/prototype/exec/y-fail-lastindex-no-write.js
- built-ins/RegExp/prototype/test/y-fail-lastindex-no-write.js

### 2: refuses isPrototypeOf

Classification: language.

Language handoff reproducer: `console.log(RegExp.prototype.isPrototypeOf(/a/));`

- built-ins/RegExp/S15.10.5_A2_T1.js
- built-ins/RegExp/prototype/S15.10.6_A1_T1.js

### 1: not yet: a PrefixUnaryExpression on a value

Classification: language.

Language handoff reproducer: `console.log(!/a/);`

- built-ins/RegExp/prototype/exec/S15.10.6.2_A5_T2.js

### 1: not yet: reading RegExp

Classification: language.

Language handoff reproducer: `const Constructor = RegExp; console.log(new Constructor('a').source);`

- built-ins/RegExp/S15.10.5_A1.js

### 1: refuses the void operator

Classification: language.

Language handoff reproducer: `console.log(void /a/);`

- built-ins/RegExp/S15.10.4.1_A1_T3.js

## Refused programs rejected by stock TypeScript

- built-ins/RegExp/S15.10.4.1_A3_T2.js: TS2769
- built-ins/RegExp/S15.10.4.1_A3_T3.js: TS2769
- built-ins/RegExp/S15.10.4.1_A5_T1.js: TS2769
- built-ins/RegExp/S15.10.4.1_A6_T1.js: TS2554
- built-ins/RegExp/S15.10.4.1_A7_T2.js: TS2554

Skipped tests and not-TypeScript programs are not implementation work in this unit. Every newly admitted official test must run native, adapted Node, and untouched upstream Node; method-family mutants must fail solely through semantic comparison. New fixtures use .a. No language work, runner verdict/classify changes, constantPattern changes, or cohere fixture edits are claimed.
