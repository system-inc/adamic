# TypeScript upstream bug ledger

Fork fixes stay separate from soundness adaptations. Kirk decides when to file
these drafted issues. Nothing in this ledger has been sent upstream.

## 1. Empty single-quoted triple-slash pragma argument throws

Status: fixed in our fork by adaptation 46; draft issue, not filed.
Upstream: TypeScript 6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8.
Location: src/compiler/parser.ts:10731 selects the argument using
`matchResult[2] || matchResult[3]`; parser.ts:10737 reads `value.length` and throws.
The quote-alternative matcher is constructed at parser.ts:10707.

Node counterexample input: `/// <reference path='' />`.
Before: `createSourceFile` throws
`TypeError: Cannot read properties of undefined (reading 'length')` in
extractPragmas. The successful match has group 2 equal to `""` and group 3
undefined, so truthy selection discards the participating empty capture.
After: parsing succeeds, referencedFiles has one entry whose fileName is `""`
and whose capture span has length 0. Nonempty single/double quotes and empty
double quotes retain their previous results.

Our patch: stage3/adapt/46-fix-pragma-empty-argument/adapt.cjs changes only the
operator in adaptation 45's checked boundary:

```diff
- const value = (matchResult[2] || matchResult[3])!;
+ const value = (matchResult[2] ?? matchResult[3])!;
```

Groups 2 and 3 are alternative quote branches and exactly one participates in
every successful match. Either participating capture may be empty. Nullish
selection keeps the empty string, and the entire result is therefore present.
Neither individual capture is asserted mandatory.

### Draft GitHub issue

Title: Empty single-quoted reference path crashes createSourceFile in TypeScript 6.0.3

Version: TypeScript 6.0.3, reproduced on Node 24.19.0.

Minimal reproduction:

```sh
mkdir tsc-empty-reference-repro
cd tsc-empty-reference-repro
npm install --no-audit --no-fund typescript@6.0.3
```

Save this as repro.cjs and run `node repro.cjs`:

```js
const ts = require("typescript");
const source = ts.createSourceFile("empty.ts", "/// <reference path='' />", ts.ScriptTarget.Latest);
console.log(JSON.stringify(source.referencedFiles.map(file => file.fileName)));
console.log(source.referencedFiles[0].end - source.referencedFiles[0].pos);
```

Expected: the parser completes and represents the empty value consistently
with a double-quoted empty reference path, printing:

```text
[""]
0
```

Actual: the parser throws
`TypeError: Cannot read properties of undefined (reading 'length')` in
extractPragmas. The equivalent `/// <reference path="" />` parses successfully.

Suggested upstream change: use `matchResult[2] ?? matchResult[3]` when choosing
between the single-quoted and double-quoted argument captures. The `!` in our
fork patch belongs to our sounder regex declarations; it is not needed in
upstream's standard string-valued capture declarations.

## 2. Truthiness treats a returned exit code as a watch handle

Status: soundness adaptation 79 under the Oct 19 language-lead ruling for
#se1zcgr; draft issue, not filed.
Upstream: TypeScript 6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8.
Locations: src/compiler/sys.ts System.exit; src/compiler/executeCommandLine.ts;
src/testRunner/unittests/helpers/tscWatch.ts:78 and
tscWatch/incremental.ts:77 (locations in the numbered adapted tree).

System.exit promises void, but test hosts override it with
`sys.exit = exitCode => sys.exitCode = exitCode`. executeCommandLine forwards
that value through its void arm. The harness treats the result as void or a
watch/solution object, and `if (closer) closer.close()` incorrectly excludes
all numbers using the declared void arm. A nonzero exit code is truthy.

Node 24.19.0 counterexample, using the real built TypeScript compiler:

```js
const ts = require("./built/local/typescript.js");
const sys = { ...ts.sys, write() {}, exit(_code) { return 1; } };
const closer = ts.executeCommandLine(sys, () => {}, ["--invalid-adamic-option"]);
console.log(closer); // 1
if (closer) closer.close(); // TypeError: closer.close is not a function
```

Observed: the invalid-option path calls exit with 1, executeCommandLine returns
1, and the harness's exact old close statement throws the TypeError above.
[The runnable proof](../adapt/79-exit-code/proofs/host-witness.cjs) extracts that
statement from the actual consumer, rather than maintaining a second copy.
This is a counterexample to the narrowing, not evidence that the unchanged
upstream suite exercises that invalid-option input in this particular consumer.

Our adaptation declares System.exit as number | undefined, gives the two
noop-as-exit test hosts dedicated stubs with that signature, and narrows the
two consumers using `typeof result === "object"`. Compiler JavaScript stays
byte-identical. The full unchanged upstream suite and its exact passing count
are the condition on accepting the harness changes; no test or baseline input
is added, deleted, filtered, or rewritten by adaptation 79.

## 3. Missing index declaration crashes getSymbolAtLocation

Status: observed 6.0.3 defect; Correctness ruled to follow typescript-go and
TypeScript main and return no symbol for an empty filtered declaration list.
95g stops before building because the caller audit found an unsafe service
reader. Draft issue, not filed. No fork fix or reader-safety proof is claimed.
Upstream: TypeScript 6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8.
Location: src/compiler/checker.ts:50153, getApplicableIndexSymbol reads
copy.declarations[0].parent after mapDefined may remove every declaration.

Node counterexample: an augmented CommonJS export of a class with a broad
computed static index, then getSymbolAtLocation of a consumer's indexed
property access. One applicable index has no declaration; neither aliasSymbol
nor symbol supplies a parent in the fallback. Before: TypeError reading parent.
With both noImplicitAny settings the witness has zero compiler diagnostics.

The resolved behavior is to return undefined before creating the copied symbol.
Official typescript-go 89d5d5b2849a0db0957065889ca58536fa6d2e4a and TypeScript
main 6161772c1370bdd4a82f808e6823bd755eb1870d instead create no symbol for an
empty declaration list and return nil. The behavior now follows these references.
However, services/findAllReferences.ts:580-581 getImplementationReferenceEntries
asserts getSymbolAtLocation is present and reads symbol.valueDeclaration. A
zero-diagnostic super.xyz implementation lookup returns undefined before the
candidate guard and throws reading valueDeclaration with it. The unit requires
stopping at this unsafe reader before building. Evidence and the complete runnable Node witness
are on branch typescript/axolotl-checker-g-evidence under
stage3/evidence/95-checker-axolotl-g/.

### Draft GitHub issue

Title: getSymbolAtLocation crashes on a declaration-free computed static index in TypeScript 6.0.3

Version: TypeScript 6.0.3 on Node 24.19.0. The current Go implementations have
an empty-declaration guard and return no symbol.

Save exported.js:

```js
/** @type {string} */
let key = 'xyz';
class Computed { static [key] = 1; }
module.exports = /** @type {typeof Computed} */ (class {});
module.exports.extra = 2;
```

Save consumer.ts:

```ts
import v = require('./exported');
v.xyz;
```

Create a program containing consumer.ts using allowJs, checkJs, noEmit and
CommonJS module resolution. Obtain its checker and call getSymbolAtLocation on
the second statement's PropertyAccessExpression. The retained witness supplies
these exact files through a compiler host and records every diagnostic and
stack; it runs against an installed or built typescript.js.

Expected: undefined for an empty filtered declaration list, following
typescript-go and TypeScript main as Correctness ruled. Actual: TypeError: Cannot read
properties of undefined (reading 'parent') in getApplicableIndexSymbol.
Suggested upstream direction: handle an empty mapDefined result before creating
the copied index symbol, and audit callers for undefined. In 6.0.3 the
super-property implementation lookup consumer needs an explicit absence guard
before reading valueDeclaration. This prerequisite is not applied by this
stopped head.

This entry and F's entry independently append to 95a's ledger end. Resolve the
append conflict when merging second; preserve both issue drafts.
