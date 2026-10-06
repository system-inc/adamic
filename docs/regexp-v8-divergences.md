# RegExp divergences: draft V8 reports

These are drafts for Kirk to review. Nothing has been filed. Observations below
use Node **24.19.0**, V8 **13.6.233.17-node.51**, Linux x86-64. Expected behavior
comes from ECMA-262, not Adamic's output. Section links use the published ES2025
specification; the current draft was also checked for class-string ordering.

## Pinned Node 24.19.0 recheck

All five drafts **reproduce on v24.19.0** at the lines identified below. No
existing refusal was dropped. Node is now pinned at 24.19.0.

| Draft | Reproduces on 24.19.0 | Decision / reproduction line |
| --- | --- | --- |
| 1. Singleton class folding | Yes | Keep refusal; report 1, line 1: false instead of true |
| 2. Scoped modifiers | Yes | Keep refusal; report 2, lines 1, 2, 3 and 5; both integration lines below |
| 3. Mixed empty class strings | Yes | Keep refusal; report 3's replacement call hangs; integration line 6 returns empty instead of a |
| 4. Unicode non-word-boundary | Yes | Keep Node-compatible runtime; report 4 prints offset 2 instead of null |
| 5. Negative word lookahead | Yes | Keep Node-compatible runtime; report 5 prints offset 1 instead of 2 |

Integration's exact program was run locally without changes. `node --version`
prints `v24.19.0`, and the program prints:

```text
2
1
false
true
false
|
```

These are Linux x64 observations, V8 13.6.233.17-node.51. Integration reports
the identical six lines from official Node v24.19.0 darwin-arm64 with the same
V8 version, and earlier from v24.14.1 on macOS arm64. Those macOS observations
were supplied by integration, not locally measured. None of the six outputs
differs across the reported platforms or versions.

The spec answers to those six lines are `4`, `2`, `true`, `false`, `true`, `a|`.
The first code-point non-boundary in `1🌍aac` is between its two `a` characters,
at UTF-16 offset 4. A checked-in test pins these witnesses and every original
minimal refusal against explicit spec answers. The Go reference verifies the
visible-pattern spec witnesses; surrogate matching deliberately follows Node.

Integration's additional modifier witnesses are:

```js
console.log(`${new RegExp('(?i:x|[^a-z])').test('B')}`); // true, spec false
console.log(`${new RegExp('(?i:a)|\\P{Ll}', 'v').test('Σ')}`); // false, spec true
```

The first required a new refusal for scoped `i` lost in a later alternative's
negated legacy class. Case-closed classes remain controls. The second is covered
by the existing stale-parser-flags refusal for Unicode property escapes.

The uppercase empty witness is:

```js
const empty = /[\q{ab|a|}]/iv.exec('A');
console.log(empty === null ? 'null' : `${empty[0]}|`); // |, spec a|
```

This confirms a wrong empty result on both platforms, so the refusal stays.
It does **not isolate alternative priority**: the independently confirmed
singleton-folding bug can make `a` fail on `A`, then let the correctly ordered
empty alternative succeed. The lowercase non-i control agrees with the spec.
Only reproducing drafts, with these qualifications, are candidates for Kirk.
Nothing was filed.

Adamic refuses incompatible pattern shapes before either backend emits code.
The input-dependent surrogate cases instead reproduce Node in both the Go
reference VM and the native C VM. These are deliberate compatibility choices,
not claims that V8's results follow the specification.

## 1. Singleton Unicode-set class strings are not case-insensitive

**24.19.0 status:** Reproduces on reproduction lines 1 and 3; refusal kept.

**Draft title:** RegExp `/[\\q{a}]/iv` fails to match `A`.

**Version:** Node 24.19.0 / V8 13.6.233.17-node.51.

```js
console.log(/[\q{a}]/iv.test('A')); // false
console.log(/[\q{AB}]/iv.test('ab')); // true, control
console.log(/[\q{Ss|x}]/iv.test('X')); // false
```

**Expected:** `true true true`. **Actual:** `false true false`.

**Spec:** [22.2.2.9 CompileToCharSet](https://tc39.es/ecma262/2025/multipage/text-processing.html#sec-compiletocharset)
and [22.2.2.10 CompileClassSetString](https://tc39.es/ecma262/2025/multipage/text-processing.html#sec-compileclasssetstring).
Canonicalization applies to the class string's characters, including strings
of length one. Ordinary class operands supply equivalence classes; V8's parser
places a folded singleton class string directly into its character ranges.

**Adamic:** Compare the resolved accepted sets, accounting for intersections,
subtractions, and masking by ordinary operands. Refuse an actual mismatch, not
every `iv` class string. `[a\q{a}]` and multi-character-only strings remain
accepted. The requested integration fixture prints `true true true` in Node,
but is a refusal fixture because its third pattern, `[\q{Ss|x}]`, also exhibits
the bug on input `X`. An accepted companion uses `[\q{Ss}]`; changing
`copyS[j] = canonicalize(c, f)` to `copyS[j] = c` makes that oracle fail.

## 2. A scoped i modifier leaks into or drops from later Unicode operands

**24.19.0 status:** Reproduces on reproduction lines 1, 2, 3 and 5, and both
integration modifier lines above; refusals kept.

**Draft title:** RegExp modifier groups leave stale parser flags for subsequent
Unicode-set classes and Unicode word escapes.

**Version:** Node 24.19.0 / V8 13.6.233.17-node.51.

```js
console.log(new RegExp('(?i:a)[b]', 'v').test('aB')); // true
console.log(new RegExp('(?-i:a)[b]', 'iv').test('aB')); // false
console.log(new RegExp('(?i:a)\\w', 'u').test('aK')); // true
console.log(new RegExp('(?i:a)(?:[b])', 'v').test('aB')); // false, control
console.log(new RegExp('(?-i:^)\\W', 'iu').test('K')); // true
```

**Expected:** `false true false false false`. **Actual:** `true false true false true`.

**Spec:** [22.2.2.7 CompileAtom](https://tc39.es/ecma262/2025/multipage/text-processing.html#sec-compileatom)
and [22.2.2.7.4 UpdateModifiers](https://tc39.es/ecma262/2025/multipage/text-processing.html#sec-updatemodifiers).
The modified RegExp Record belongs to the group's subpattern; subsequent
operands use the enclosing Record. V8's parser flags change at group entry but
are not restored at exit, although its builders have correctly scoped flags.
Entering a new group resets the stale parser flags, as the control shows.

**Adamic:** Track parser flag state in source order and compare the affected
resolved Unicode sets. Preserve agreeing literals, reset groups, digit sets,
and class strings whose correctly scoped string matcher masks the stale state.
Unicode non-set word escapes with leaked `i`, and `\W` with dropped `i`, are
also refused. The current
stage-0 TypeScript target rejects modifier *literals* with TS18062; constant
`new RegExp` controls exercise Adamic's own refusal without changing that target.

## 3. Mixed empty class strings can hang the replacement slow path

**24.19.0 status:** Reproduces on the final replacement line (bounded hang)
and integration's uppercase empty-result line above; refusal kept. An
independent empty-first priority defect remains unproved.

**Draft title:** RegExp Unicode-set empty match loops forever in deoptimized
String replacement near a surrogate pair.

**Version:** Node 24.19.0 / V8 13.6.233.17-node.51.

```js
const r = /[\q{ab|a|}]/gv;
const [zero] = [0, NaN];
r.lastIndex = zero;
console.log('🌍a🌍'.replace(r, ''));
```

**Expected:** Complete and print `🌍🌍`. **Actual:** No output; killed after a
two-second timeout by the regression test (also observed with a three-second
external timeout). Replacing the assignment with `r.lastIndex = 0` completes
and prints the expected value on the optimized path.

**Spec:** [22.2.6.11 RegExp.prototype [ Symbol.replace ]](https://tc39.es/ecma262/2025/multipage/text-processing.html#sec-regexp.prototype-%symbol.replace%)
uses [22.2.7.3 AdvanceStringIndex](https://tc39.es/ecma262/2025/multipage/text-processing.html#sec-advancestringindex)
to advance after empty matches; repeated execution must make progress.

**Ordering limitation:** The requested direct empty-alternative ordering
divergence was **not reproduced**. `/[\q{ab|a|}]/v.exec('a')[0]` is `a`, which
agrees with published ES2025 and the current draft: longer strings first,
single-character matching next, empty matching last. This draft reports the
observed hang, not an unproved ordering bug.

**Adamic:** Following the ruling, refuse a resolved class containing an empty
string, a singleton, and a multi-character string. This is conservative:
ordinary `exec` can agree on the refused shape, and a dead `{0}` occurrence can
also be refused. Removing empty by subtraction, `[\q{ab|}]`, and `[\q{a|}]`
are outside this rule. The refusal's diagnostic names the replacement behavior
and the spec sections rather than asserting an observed ordering failure.

## 4. Unicode non-word-boundary assertions can match inside a surrogate pair

**24.19.0 status:** Reproduces on the final console line; Node-compatible
runtime kept, no compile-time refusal.

**Draft title:** Unicode RegExp `\B` accepts an interior UTF-16 position.

**Version:** Node 24.19.0 / V8 13.6.233.17-node.51.

```js
const r = /\B/dug;
const m = r.exec('a🌍b');
console.log(m?.indices[0], r.lastIndex); // [2, 2] 2
```

**Expected:** No match: every code-point boundary separates a word character
from a non-word character or the end. **Actual:** An empty match inside the
surrogate pair at UTF-16 offset 2. Sticky execution starting at offset 2 also
accepts that interior position.

**Spec:** [22.2.2.2 CompilePattern](https://tc39.es/ecma262/2025/multipage/text-processing.html#sec-compilepattern)
uses code points in Unicode mode; [22.2.2.4 CompileAssertion](https://tc39.es/ecma262/2025/multipage/text-processing.html#sec-compileassertion)
evaluates boundary assertions over that input representation. An interior
UTF-16 position is not a code-point boundary.

**Adamic:** Match Node deliberately. Retry assertion-only matching at UTF-16
positions during failed search; consuming Unicode instructions reject the
interior of a pair. Preserve V8's initial lastIndex rewind and sticky retry.
Comments in both interpreters identify the spec departure.

## 5. Unicode negative word lookahead succeeds inside a surrogate pair

**24.19.0 status:** Reproduces on the final console line; Node-compatible
runtime kept, no compile-time refusal.

**Draft title:** Unicode RegExp `(?!\W)` succeeds inside a surrogate pair after
the consuming assertion operand rejects the interior position.

**Version:** Node 24.19.0 / V8 13.6.233.17-node.51.

```js
const r = /(?!\W)/dug;
const m = r.exec('🌍');
console.log(m?.indices[0], r.lastIndex); // [1, 1] 1
```

**Expected:** First match at `[2, 2]`, the end of the non-word code point.
**Actual:** First match at `[1, 1]`, inside its surrogate pair. Sticky execution
at offset 1 exhibits the same behavior.

**Spec:** [22.2.2.2 CompilePattern](https://tc39.es/ecma262/2025/multipage/text-processing.html#sec-compilepattern),
[22.2.2.4 CompileAssertion](https://tc39.es/ecma262/2025/multipage/text-processing.html#sec-compileassertion),
and [22.2.7.2 RegExpBuiltinExec](https://tc39.es/ecma262/2025/multipage/text-processing.html#sec-regexpbuiltinexec).
The Unicode input representation and search positions must not expose an
interior surrogate position as a new candidate assertion position.

**Adamic:** Reproduce Node's consuming failure and resulting negative assertion
success. Controls check that consuming atoms cannot consume half the pair and
that an initial `/./uy` execution at offset 1 still rewinds to offset 0.

## Refusal cost and evidence

The checked-in test262 execution corpus has 127,369 executions and 2,791
pattern/flag pairs; the parser corpus has 5,746 extracted patterns. These
refusals affect **zero** executions, patterns, or attributed test262 files in
those corpora. The fixed-seed random execution corpus loses **89 of 10,000**
native cases (83 unique pattern/flag pairs) to the mixed-empty rule; all 89 agree with Node for their sampled
`exec` input. Go still compares all 10,000. Every accepted native case runs both
metered and unlimited C configurations with zero disagreements. This branch's
main base predates the DFA; there is no DFA configuration to claim here.

An additional visible-shape sweep has 4,650 input cases across 310 pattern/flag
pairs: 244 accepted pairs (3,660 cases) have zero disagreements. Refusals cover
66 pairs; their sampled inputs include 830 agreeing cases and 160 differing
cases. This illustrates why a pattern-level refusal also rejects inputs on
which an unsafe pattern happens to agree. It is separate from test262 cost.

See [the unit report](../cloud/reports/regex-v8-divergences/README.md) for exact
commands, before/after totals, refusal lists, additional generated-shape
collateral, and ten real mutants. The fixture and corpus counts describe the
checked-in extraction, not a new full 1,879-test runner survey.
