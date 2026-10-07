# Regex remaining results

Code: 31862c1. Claim: 394a34f. Surrogate fixes: 866f466 and 310c50e. Branch codex/regex-remaining, all pushed. No pull request.

Node v24.19.0, V8 13.6.233.17-node.51, Linux x64. Stock TypeScript 6.0.3. Test262 7ab7fafa0003f73fc85c1b95d88094d33f7eb8bd. nproc: 5.

| Filter | Before pass | After pass | Before refused | After refused | Not TypeScript | Skipped | Total |
|---|---:|---:|---:|---:|---:|---:|---:|
| built-ins/RegExp | 914 | 1035 | 226 | 105 | 223 | 516 | 1879 |
| language/literals/regexp | 16 | 16 | 8 | 8 | 4 | 212 | 240 |
| annexB/built-ins/RegExp | 1 | 9 | 13 | 5 | 23 | 25 | 62 |
| annexB/language/literals/regexp | 0 | 0 | 0 | 0 | 8 | 0 | 8 |
| Total | 931 | 1060 | 247 | 118 | 258 | 753 | 2189 |

Zero disagreements, crashes, missing paths or lost passes. Every newly passing official test agrees with native, adapted Node and untouched upstream Node. Gains: 94 exact-constructor harness admissions, 22 regex effect admissions, 8 compile tests, 5 direct intrinsic metadata observations.

Of the 129 new runner passes, 128 pass the independent stock-TypeScript audit. S15.10.4.1_A5_T1.js is the exception: new RegExp(undefined, 'ii') is TS2769 under the stock overloads. Existing constructor behavior agrees with Node; this unit's general toString effect summary incidentally unblocks it. No feature was built specifically for that invalid program. It is excluded from the coverage claim and recorded for the validity worker; verdict/classify policy is untouched. Baseline passes were not independently stock-audited, so this report makes no total stock-valid-pass claim.

The 118 remaining refusals contain 114 stock-valid programs: 15 library runtime-compiler dependencies and 99 language handoffs. The four remaining stock-invalid refusals are S15.10.4.1_A3_T2.js and A3_T3.js (TS2769), A6_T1.js and A7_T2.js (TS2554). Together with A5_T1 these are the five baseline-invalid refusals. They are not implementation work.

## Surrogate assertions

a8cc903 did not fix split: native disagreed on five of the eight supplied pattern/input pairs, and JavaScript lowering on two. The added .a fixture includes the exact supplied loop and exec, test, match, matchAll, replace, replaceAll, search and split for both assertions under u and v. All paths now agree with Node.

V8 split also depends on the hidden representation of a numeric limit: for a sensitive u pattern, an immediate small integer and an equal boxed number can produce different results. Lowering preserves omitted/undefined limits, proves safe literal or insensitive cases, and explicitly refuses unproved numeric limits for sensitive patterns. Bytecode sensitivity includes nested negative assertions. Unicode non-word-boundary candidates use the VM fallback instead of skipping surrogate interiors in the regular finder.

## Validation and mutations

[Full commands and outputs](../cloud/reports/regex-remaining/library/REPORT.txt) and [surrogate evidence](../cloud/reports/regex-remaining/surrogates/) record the checks. Go: 127,369 corpus executions and 10,000 seeded differential cases in budgeted VM, regular plus capture recovery, and unlimited VM. Native: 127,369 corpus executions, 9,911 accepted seeded cases plus 89 explicit existing V8-divergence refusals, and 2,659 V8 edge cases in all three configurations. Zero disagreements. All regex oracle fixtures pass sanitized, release and JavaScript paths, including timing bounds and Linux leak probes. Lower, fresh, flow and runner package checks pass; vet passes. The full uncached go test ./... gate was not run.

Seven mutants exit successfully with empty sanitizer stderr and fail solely on stdout comparison with Node: SyntaxError changed to TypeError; Symbol.search index incremented; toString flags omitted; replacement callback result discarded; compile lastIndex reset omitted; intrinsic toString.length changed; split candidate advancement changed to code points. Paired normal/Node/mutant outputs and the reproducible mutation script are in the library evidence directory.

Setup first failed while merge integration lacked ir.ErrorIs and ir.BuiltinError. After merge resolution the retry completed: Go 0s, clang 0s, Node 0s, submodules 0s, build-cache warm 180s, total 180s on 5 processors (cgroup 4 CPUs). Exact timing lines and initial failure are retained.

## Measurement provenance

The runner includes origin/codex/test262-ts-validity at af128990588aab7ee52ea3ab8527d638009c4979. The merged parents and claim provenance are recorded in the claim commit. Commands use --adapt --json --jobs 1 --timeout 60s, independently sharded processes rather than concurrent use of a shared TypeScript IPC stream. survey.py reproduces the survey. Before: frozen ca96d50 plus decoded-input merge correction. After: frozen claim plus production changes committed as 31862c1; subsequent comments, inventory and fixture changes do not change official test behavior. An environment restart interrupted the first after run; 129 completed records were retained and only the other 2060 paths rerun. The final before/after JSONL files each contain all 2189 unique paths. Interrupted checks are not counted as passing.

The read-only TestRegExpRemainingTypeScriptAudit command uses ADAMIC_REGEX_REMAINING_RESULTS, ADAMIC_REGEX_REMAINING_CORPUS and ADAMIC_REGEX_REMAINING_AUDIT. Setting ADAMIC_REGEX_REMAINING_BEFORE selects newly admitted passes. Remaining audit: PASS 12.166s, 114/118 stock-valid. Admissions audit: PASS 16.468s, 128/129 stock-valid. Both logs and summary JSON are retained beside before.jsonl and after.jsonl.

## Remaining inventory and language handoff

No language feature below was implemented. Lists are stock-TypeScript-valid programs, largest reason first. Runtime syntax/flags require a native ECMAScript compiler; there is no helper-process substitute or freezing of runtime inputs.

### 41: not yet: a BinaryExpression with a string and a value

Language handoff.

One-line reproducer: `const m = /a/.exec('a'); console.log('match=' + (m && m[0]));`

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

### 11: not yet: RegExp with nonconstant flags: the native runtime has no ECMAScript pattern compiler

Library dependency.

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

### 8: refuses var

Language handoff.

One-line reproducer: `var r = /a/; console.log(r.test('a'));`

- built-ins/RegExp/S15.10.2.10_A1.1_T1.js
- built-ins/RegExp/S15.10.2.10_A1.2_T1.js
- built-ins/RegExp/S15.10.2.10_A1.3_T1.js
- built-ins/RegExp/S15.10.2.10_A1.4_T1.js
- built-ins/RegExp/S15.10.2.10_A1.5_T1.js
- built-ins/RegExp/S15.10.2.10_A3.1_T1.js
- built-ins/RegExp/S15.10.2.10_A4.1_T1.js
- built-ins/RegExp/S15.10.2.11_A1_T1.js

### 6: not yet: a BinaryExpression with a string and a boolean

Language handoff.

One-line reproducer: `console.log('match=' + /a/.test('a'));`

- language/literals/regexp/S7.8.5_A3.1_T1.js
- language/literals/regexp/S7.8.5_A3.1_T2.js
- language/literals/regexp/S7.8.5_A3.1_T3.js
- language/literals/regexp/S7.8.5_A3.1_T4.js
- language/literals/regexp/S7.8.5_A3.1_T5.js
- language/literals/regexp/S7.8.5_A3.1_T6.js

### 6: not yet: a BinaryExpression with a value and a string

Language handoff.

One-line reproducer: `const m = /a/.exec('a'); console.log((m && m[0]) + '!');`

- built-ins/RegExp/S15.10.2.11_A1_T4.js
- built-ins/RegExp/S15.10.2.11_A1_T5.js
- built-ins/RegExp/S15.10.2.11_A1_T6.js
- built-ins/RegExp/S15.10.2.11_A1_T7.js
- built-ins/RegExp/S15.10.2.11_A1_T8.js
- built-ins/RegExp/S15.10.2.11_A1_T9.js

### 5: not yet: a computed class base; name the base class directly

Language handoff.

One-line reproducer: `class Derived extends (class extends RegExp {}) {}`

- annexB/built-ins/RegExp/legacy-accessors/input/this-subclass-constructor.js
- annexB/built-ins/RegExp/legacy-accessors/lastMatch/this-subclass-constructor.js
- annexB/built-ins/RegExp/legacy-accessors/lastParen/this-subclass-constructor.js
- annexB/built-ins/RegExp/legacy-accessors/leftContext/this-subclass-constructor.js
- annexB/built-ins/RegExp/legacy-accessors/rightContext/this-subclass-constructor.js

### 4: not yet: RegExp with a nonconstant pattern: the native runtime has no ECMAScript pattern compiler

Library dependency.

- built-ins/RegExp/S15.10.2.10_A2.1_T1.js
- built-ins/RegExp/S15.10.2.10_A2.1_T2.js
- built-ins/RegExp/S15.10.4.1_A8_T2.js
- built-ins/RegExp/quantifier-integer-limit.js

### 4: not yet: an array of never

Language handoff.

One-line reproducer: `const errors = []; errors.push('error');`

- built-ins/RegExp/prototype/exec/S15.10.6.2_A3_T1.js
- built-ins/RegExp/prototype/exec/S15.10.6.2_A3_T5.js
- built-ins/RegExp/prototype/exec/S15.10.6.2_A3_T6.js
- built-ins/RegExp/prototype/exec/S15.10.6.2_A3_T7.js

### 4: refuses inherited library member prototype read as an own field

Language handoff.

One-line reproducer: `console.log(RegExp.prototype.source);`

- built-ins/RegExp/prototype/global/S15.10.7.2_A8.js
- built-ins/RegExp/prototype/ignoreCase/S15.10.7.3_A8.js
- built-ins/RegExp/prototype/multiline/S15.10.7.4_A8.js
- built-ins/RegExp/prototype/no-regexp-matcher.js

### 3: not yet: instanceof against a value that isn't a declared class

Language handoff.

One-line reproducer: `console.log(/a/.exec('a') instanceof Array);`

- built-ins/RegExp/prototype/exec/S15.10.6.2_A1_T1.js
- built-ins/RegExp/prototype/exec/S15.10.6.2_A1_T6.js
- language/literals/regexp/S7.8.5_A4.1.js

### 3: refuses a method read as a value (exec would lose its object, and this with it)

Language handoff.

One-line reproducer: `const fn = /a/.exec; fn.call(/a/, "a");`

- built-ins/RegExp/prototype/exec/S15.10.6.2_A2_T10.js
- built-ins/RegExp/prototype/exec/S15.10.6.2_A2_T3.js
- built-ins/RegExp/prototype/exec/S15.10.6.2_A8.js

### 3: refuses a method read as a value (test would lose its object, and this with it)

Language handoff.

One-line reproducer: `const fn = /a/.test; fn.call(/a/, "a");`

- built-ins/RegExp/prototype/test/S15.10.6.3_A2_T10.js
- built-ins/RegExp/prototype/test/S15.10.6.3_A2_T3.js
- built-ins/RegExp/prototype/test/S15.10.6.3_A8.js

### 3: refuses a number as a condition

Language handoff.

One-line reproducer: `const m = /a/.exec('a'); if (m !== null && m.index) console.log('nonzero');`

- built-ins/RegExp/prototype/exec/S15.10.6.2_A3_T2.js
- built-ins/RegExp/prototype/exec/S15.10.6.2_A3_T3.js
- built-ins/RegExp/prototype/exec/S15.10.6.2_A3_T4.js

### 2: not yet: a value of type any

Language handoff.

One-line reproducer: `const value: any = /a/; console.log(value);`

- built-ins/RegExp/character-class-escape-non-whitespace.js
- language/literals/regexp/y-assertion-start.js

### 2: not yet: hasOwnProperty on a function (its prototype property depends on whether it was declared or made as an arrow)

Language handoff.

One-line reproducer: `console.log(RegExp.hasOwnProperty('prototype'));`

- built-ins/RegExp/prototype/S15.10.5.1_A1.js
- built-ins/RegExp/prototype/S15.10.5.1_A2.js

### 2: refuses Object.defineProperty

Language handoff.

One-line reproducer: `Object.defineProperty(/a/, 'lastIndex', {writable: false});`

- built-ins/RegExp/prototype/exec/y-fail-lastindex-no-write.js
- built-ins/RegExp/prototype/test/y-fail-lastindex-no-write.js

### 2: refuses a method read as a value (toString would lose its object, and this with it)

Language handoff.

One-line reproducer: `const fn = /a/.toString; fn.call(/a/);`

- built-ins/RegExp/prototype/15.10.6.js
- built-ins/RegExp/prototype/toString/S15.10.6.4_A8.js

### 2: refuses isPrototypeOf

Language handoff.

One-line reproducer: `console.log(RegExp.prototype.isPrototypeOf(/a/));`

- built-ins/RegExp/S15.10.5_A2_T1.js
- built-ins/RegExp/prototype/S15.10.6_A1_T1.js

### 1: not yet: a PrefixUnaryExpression on a value

Language handoff.

One-line reproducer: `console.log(!/a/);`

- built-ins/RegExp/prototype/exec/S15.10.6.2_A5_T2.js

### 1: not yet: reading RegExp

Language handoff.

One-line reproducer: `const Constructor = RegExp; console.log(new Constructor('a').source);`

- built-ins/RegExp/S15.10.5_A1.js

### 1: refuses the void operator

Language handoff.

One-line reproducer: `console.log(void /a/);`

- built-ins/RegExp/S15.10.4.1_A1_T3.js

Detached-method diagnostics also cover function expandos and for-in. Additional stock-TypeScript-valid language reproducers: `const fn = /a/.test; for (const key in fn) console.log(key);` (function property enumeration); `function f(text: string): boolean { return true; } f.test = /a/.test;` (function expando); `console.log(Object.prototype.toString.call(RegExp.prototype));` (prototype object and reflective call). These remain language work, not general metadata admission.
