## 9. skipWhile promises that every suffix element fails its predicate

Status: soundness adaptation 99s under Adamic's Nov 5 ruling for #b5w3ycg;
draft issue, not filed.
Upstream: TypeScript 6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8.
Location: src/compiler/core.ts:2572-2585, skipWhile's three declarations and
its returned suffix cast.

The declared `Exclude<T, U>[]` result is false. The loop proves that the first
returned element fails the predicate, but never tests later returned elements.

Node counterexample, using the actual owner transpiled by witness.cjs:

```js
skipWhile([1, "x", 2], x => typeof x === "number"); // ["x", 2]
```

The declared result excludes number, yet the second returned element is 2.
Before and after our patch Node returns the same array, with length 2.
No extra predicate call or runtime check is introduced.

Our patch: stage3/adapt/99-correctness-axolotl-s/adapt.cjs changes all three
return types to `T[]` (retaining `| undefined` in the two optional signatures)
and removes `as Exclude<T, U>[]` from the returned slice. The only adapted-src
caller, injectClassTypeMetadata in transformers/ts.ts, filters the result with
isModifier at its own use and does not read the old exclusion narrowing.

### Draft GitHub issue

Title: skipWhile's type predicate overload incorrectly narrows the entire suffix

Version: TypeScript 6.0.3, reproduced on Node 24.19.0.

Minimal reproduction in an internal compiler test:

```ts
const values: (number | string)[] = [1, "x", 2];
const suffix = skipWhile(values, (x): x is number => typeof x === "number");
// Inferred string[], actual ["x", 2].
for (const value of suffix) value.toUpperCase(); // throws for 2
```

Expected: the suffix has type `(number | string)[]`, since skipWhile stops
checking at the first failed predicate. Actual: the declared result is string[].
Suggested upstream change: return `T[]`, retaining the undefined result for
undefined input, and remove the suffix cast. This correction changes types only.

- createQueue boxed slots, 99s amendment, F15/F16/F17: Adamic, task b5w3ycg, Nov 7 03:20Z: established-shape adaptations, Compiler accepts. Whole lane and all 398 bounded units each compare 10041 enqueues, 10123 dequeues and 13660 public isEmpty calls, with 0 differences; sparse initial arrays 0, own iterator overrides 0. Sparse and undefined-payload witnesses match stock. The deliberately changed iterator detector seed is outside the observed own-iterator class (0 calls). [Proof](../adapt/99-correctness-axolotl-s/QUEUE-BOX-PROOF.md).

- **JavaScript-changing (byte-different, value-identical), 99-correctness-axolotl-s sourcemap amendment (Compiler):** TypeScript 6.0.3 src/compiler/sourcemap.ts:316 clears createSourceMapGenerator.flushMappingBuffer's private mappingCharCodes with `splice(0)` instead of `length = 0`; the ignored splice result changes emitted bytes while preserving mappings and subsequent empty-buffer reads. Evidence: typescript/sourcemap-splice-evidence, stage3/census/sourcemap-splice-evidence.
