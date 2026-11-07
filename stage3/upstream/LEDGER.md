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

## 3. Excess JavaScript type argument crashes getTypeArgumentConstraint

Status: fixed in our fork by JavaScript-changing adaptation 95f under the
November 1 Correctness ruling; draft issue, not filed.
Upstream: TypeScript 6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8.
Location: src/compiler/checker.ts:42849, getTypeArgumentConstraint calls
getConstraintOfTypeParameter with typeParameters[typeArgumentPosition] even
when the argument position exceeds the generic arity.

Node counterexample: a one-parameter JavaScript generic C and JSDoc
`@type {C<string, number, boolean>}` or `@extends {C<string, number, boolean>}`.
Querying the second or third argument through TypeChecker.getTypeArgumentConstraint
throws TypeError reading resolvedBaseConstraint. JavaScript's class-reference
resolution deliberately continues after invalid arity, so the error-type early
return used for TypeScript does not establish parameter presence here.

Our patch: stage3/adapt/95-checker-axolotl-f/adapt.cjs inserts only:

```diff
 const relevantTypeParameter = typeParameters[typeArgumentPosition];
+if (relevantTypeParameter === undefined) return undefined;
 const constraint = getConstraintOfTypeParameter(relevantTypeParameter);
```

The API already returns Type | undefined; internal and service callers already
handle no constraint. This is a crash fix, not a required-value assertion.
Official typescript-go 89d5d5b2849a0db0957065889ca58536fa6d2e4a and TypeScript
main 6161772c1370bdd4a82f808e6823bd755eb1870d both return nil past arity in
the equivalent Go checker function. Evidence is on branch
typescript/axolotl-checker-f-evidence under stage3/evidence/95-checker-axolotl-f/.

### Draft GitHub issue

Title: getTypeArgumentConstraint crashes on excess JSDoc type arguments in TypeScript 6.0.3

Version: TypeScript 6.0.3 on Node 24.19.0. Current Go main already bounds-checks
this position; this draft documents the 6.0.3 defect and fork repair.

Save the following as repro.cjs beside an installed typescript@6.0.3 and run
`node repro.cjs`:

```js
const ts = require('typescript');
const file = '/repro.js';
const text = '/** @template T */ class C {}\n/** @type {C<string, number, boolean>} */ let value;';
const host = ts.createCompilerHost({allowJs: true, checkJs: true, noEmit: true});
const original = host.getSourceFile;
host.getSourceFile = (name, ...args) => name === file
    ? ts.createSourceFile(name, text, ts.ScriptTarget.Latest, true)
    : original.call(host, name, ...args);
const program = ts.createProgram([file], {allowJs: true, checkJs: true, noEmit: true}, host);
const checker = program.getTypeChecker();
const source = program.getSourceFile(file);
const argument = source.statements[1].jsDoc[0].tags[0].typeExpression.type.typeArguments[1];
console.log(checker.getTypeArgumentConstraint(argument));
```

Expected: undefined (the excess argument has no corresponding parameter and
therefore no constraint). Actual: TypeError: Cannot read properties of undefined
(reading 'resolvedBaseConstraint'). Suggested change: handle the absent local
before calling getConstraintOfTypeParameter, as the Go checker already does.

This entry and the independent G entry both append to 95a's ledger end. Resolve
the append conflict when merging the second branch; preserve both issue drafts.

## 4. Missing index declaration crashes getSymbolAtLocation

Status: JavaScript-changing fork fix in 95g, scoped to the previously throwing
parent fallback. Draft issue, not filed. No public API line is changed.
Upstream: TypeScript 6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8.
Location: src/compiler/checker.ts:50153, getApplicableIndexSymbol reads
copy.declarations[0].parent after mapDefined may remove every declaration.

An augmented CommonJS export of a class with a broad computed static index
reaches one applicable index with no declaration and neither aliasSymbol nor
symbol. getSymbolAtLocation(v.xyz) throws reading parent with zero diagnostics,
at both noImplicitAny settings. The fix keeps the ordered aliasSymbol / symbol /
fallback selection. Only inside the third arm, a missing first declaration
returns undefined; every successful arm retains its original result.

This follows the behavior at the throwing arm in
[typescript-go 89d5d5b2849a0db0957065889ca58536fa6d2e4a, checker.go:32209](https://github.com/microsoft/typescript-go/blob/89d5d5b2849a0db0957065889ca58536fa6d2e4a/internal/checker/checker.go#L32209)
and [TypeScript main 6161772c1370bdd4a82f808e6823bd755eb1870d, checker.go:32895](https://github.com/microsoft/TypeScript/blob/6161772c1370bdd4a82f808e6823bd755eb1870d/tsc/internal/checker/checker.go#L32895).
Both Go implementations guard the empty filtered list more broadly. They also
return no symbol where 6.0.3 successfully returns a zero-declaration __index
symbol with a known type symbol. 95g deliberately preserves that successful
6.0.3 path, including Computed.xyz and the unchanged super.xyz implementation
query (undefined, zero diagnostics).

On an imported derived-class super.xyz witness, getImplementationAtPosition
already threw in the checker. It now throws reading valueDeclaration in
services/findAllReferences.ts:580. This is a moved existing crash, not a new
crash on a successful input; no service edit is included. The FourSlash
verifySymbolAtLocation assertion likewise now reports its missing-symbol error
instead of the checker exception. Actual built-compiler service probes preserve
all 29 successful rows and introduce zero new crashes; 27 formerly throwing
rows succeed and one moves. Full evidence is on typescript/axolotl-checker-g-evidence
under stage3/evidence/95-checker-axolotl-g/scoped-final/.

### Draft GitHub issue

Title: getSymbolAtLocation crashes on a declaration-free computed static index in TypeScript 6.0.3

Version: TypeScript 6.0.3 on Node 24.19.0. Current Go implementations already
return no symbol on the throwing input (references above).

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
the second statement's PropertyAccessExpression. Both noImplicitAny settings
produce zero diagnostics. Actual: TypeError: Cannot read properties of undefined
(reading 'parent') in getApplicableIndexSymbol. Expected: undefined when neither
type symbol supplies a parent and the fallback has no declaration, following
the Go behavior on this input. Preserve successful symbol-supplied parent paths.

Suggested fix: guard the first declaration only in the third parent-selection
arm. Follow-up: the service super-property implementation reader assumes a
symbol, so the derived-class variant moves the existing exception there rather
than completing. Audit that reader independently; it is not changed by 95g.

This entry and F's entry independently append to 95a's ledger end. Resolve the
append conflict when merging second; preserve both issue drafts and leave
renumbering to the wave cutter.

## 5. Empty repeated extends clause crashes declaration emit

Status: fixed in our fork by adaptation 99e; draft issue, not filed.
Upstream: TypeScript 6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8.
Location: src/compiler/transformers/declarations.ts:1678, the heritage callback
inside transformTopLevelDeclaration's expression-base rewrite.

Node counterexample input: `export class C extends f() extends {}`.
A further reduced standalone witness is `class C extends 0 extends {}`; it
also crashes both CLI declaration emit and transpileDeclaration, with TS2507
for the numeric base and TS1172 for the repeated clause.
The parser reports no diagnostics and creates heritage type lengths `[1, 0]`.
Semantic checking reports TS1172, `'extends' clause already seen.`, for the
second clause. `tsc --declaration`, `ts.transpileDeclaration`, and program.emit
throw `TypeError: Cannot read properties of undefined (reading 'kind')` in
isVariableDeclaration, called by createGetSymbolAccessibilityDiagnosticForNode.
The outer getEffectiveBaseTypeNode guard proves only the first clause's type.

Our patch: stage3/adapt/99-correctness-axolotl-e/adapt.cjs:

```diff
- const heritageClauses = factory.createNodeArray(map(input.heritageClauses, clause => {
+ const heritageClauses = factory.createNodeArray(mapDefined(input.heritageClauses, clause => {
      if (clause.token === SyntaxKind.ExtendsKeyword) {
+         const firstType = clause.types[0];
+         if (!firstType) return undefined;
          const oldDiag = getSymbolAccessibilityDiagnostic;
-         getSymbolAccessibilityDiagnostic = createGetSymbolAccessibilityDiagnosticForNode(clause.types[0]);
+         getSymbolAccessibilityDiagnostic = createGetSymbolAccessibilityDiagnosticForNode(firstType);
```

The empty clause is omitted, matching the ordinary heritage transformation's
empty-clause filtering. Nonempty clauses retain their existing treatment.
The declaration becomes `declare const C_base: any;` followed by
`export declare class C extends C_base {}` and `export {};`.
Node witnesses, full stacks, control equality, diagnostic counts and baseline
results are recorded in the adaptation's README and proofs.

The pinned native compiler at d92d9bfee114c80be2c375d72edae966176e3a4f
and TypeScript main at 6161772c1370bdd4a82f808e6823bd755eb1870d use a
first-effective-base rewrite and elide empty clauses. The legacy release-6.0
branch still has the unsafe JavaScript transform; no JavaScript fix commit
is claimed. Both native and fixed JavaScript emit the same declaration for
the original f() witness. On the reduced numeric witness, native emits a
literal base type 0 while TypeScript 6 emits number, matching its unchanged
single-clause control. This patch does not alter type inference.

### Draft GitHub issue

Title: Repeated empty extends clause crashes declaration emit in TypeScript 6.0.3

Version: TypeScript 6.0.3, reproduced on Node 24.19.0.

Minimal reproduction, in a fresh directory without tsconfig.json:

```sh
npm install --no-audit --no-fund typescript@6.0.3
printf '%s\n' 'export class C extends f() extends {}' > input.ts
node node_modules/typescript/lib/tsc.js --declaration input.ts
```

API reproduction:

```js
const ts = require("typescript");
ts.transpileDeclaration("export class C extends f() extends {}", { fileName: "input.ts" });
```

Expected: report applicable diagnostics and finish declaration emit without an
exception. The empty repeated clause should be omitted, consistently with the
ordinary heritage transformer and the native compiler.

Actual: both commands throw a TypeError reading kind on undefined, from
isVariableDeclaration through createGetSymbolAccessibilityDiagnosticForNode,
map's heritage callback, and transformTopLevelDeclaration. The source parses
without errors; semantic checking reports TS1172 for the second extends clause
and TS2304 for the undeclared f. Declaration emit crashes even though
noEmitOnError is false, its default.

Suggested upstream change for the legacy implementation: use mapDefined for
the expression-base heritage mapping, return undefined for an empty extends
clause, and pass the checked first type to the diagnostic helper. Native main
already handles this case; this draft concerns the TypeScript 6 JavaScript
compiler. No claim is made that erroneous input must yield a valid program.

## 6. Init emits an undefined token for an explicitly cleared option

Status: fixed in our fork by adaptation 99g; draft issue, not filed.
Upstream: TypeScript 6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8.
Locations: src/compiler/commandLineParser.ts parseOptionValue:2111 stores
undefined for a null token; generateTSConfig:2838 retains the own key;
emitOption:2933 selects options[setting] ?? defaultValue; formatSingleValue:2960
returns JSON.stringify's undefined result. These are stock source coordinates.

Node counterexample: `tsc --init --outFile null` exits successfully but writes
`"outFile": undefined,` in tsconfig.json. The token undefined is not valid
JSON config syntax, so a subsequent invocation cannot read that generated file
cleanly. The user explicitly cleared this option; retaining it as null both
preserves that instruction and produces a valid config value.

The pinned Go compiler at system-inc/TypeScript
d92d9bfee114c80be2c375d72edae966176e3a4f writes `"outFile": null,` for the same
command. It also writes null for the fixed strict setting. Its Go parser
stores nil and its config generator serializes that value as JSON null. The
pinned Go compiler instead panics for a cleared target, because it attempts
an enum lookup for nil; our fork follows the ruled null-output behavior for
that setting as well, rather than copying the panic.

Our patch: stage3/adapt/99-correctness-axolotl-g/adapt.cjs changes only
emitOption's formatting. Read options[setting] once; if it is an own undefined
value, format it as the literal null. Otherwise use exactly the existing
nullish default selection and formatting. The active/commented form, option
ordering and absent-option defaults are preserved. This intentionally changes
JavaScript behavior only for own undefined option values. No assertion is
added, no public API is changed, and no test baseline is rewritten.

The executable before/after CLI witnesses, unchanged-output controls, Go
comparisons and original-function mutant are in
[99g's witness](../adapt/99-correctness-axolotl-g/witness.cjs). The adapter README
records the full proof and oracle results. The earlier upstream snapshot at
Microsoft/TypeScript main 6161772c1370bdd4a82f808e6823bd755eb1870d has an active
Go init.go identical to the pinned version and retains a TypeScript fixture
with generateTSConfig unchanged from 6.0.3.

### Draft GitHub issue

Title: tsc --init --outFile null writes an undefined value into tsconfig.json

Version: TypeScript 6.0.3, Node 24.19.0, Linux x64.

Minimal reproduction in a fresh directory, with TypeScript 6.0.3 installed:

```sh
printf 'export {};\n' > index.ts
npx tsc --init --outFile null
npx tsc --showConfig
```

Expected: init writes `"outFile": null,`, preserving the explicit clearing,
and the generated file reads back successfully. This matches the native Go
compiler for outFile and the fixed strict setting.

Actual: init exits 0 but emits `"outFile": undefined,`. The next command exits
1 and reports:

```text
tsconfig.json(44,16): error TS5024: Compiler option 'outFile' requires a value of type string.
```

Suggested upstream change: in generateTSConfig's emitOption, distinguish an
own undefined value from an absent option. Emit the literal null for the
former in the same active or commented form already selected for that option.
Preserve the existing nullish fallback and formatting for every other value.
This must also apply to preset options such as strict; only changing the
remaining-options loop would leave those cleared settings using their defaults.

## 7. tsc passes a caught value's message, undefined or a non-string, to a callback declared `(message: string)`

Status: fixed in our fork by JavaScript-changing adaptation 97e, which converts
the message to text under Adamic's Nov 1 ruling (#axochk); draft issue, not
filed.
Upstream: TypeScript 6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8.
Location: src/compiler/program.ts createGetSourceFile and
createWriteFileMeasuringIO, each `catch (e) { if (onError) { onError(e.message); } }`;
src/compiler/types.ts's two declarations of that callback,
WriteFileCallback's `onError?: (message: string) => void` (types.ts:4584) and
CompilerHost.getSourceFile's (types.ts:8142). Lines are in the adapted tree at
adaptation 94's tip.

Both functions wrap a supplied callback: readFile for getSourceFile, the
writer, createDirectory and directoryExists for writeFile. Nothing constrains
what that callback throws, and the catch hands `e.message` to onError as it
is. The declared `string` holds only when the thrown value is an object with a
string message. tsc's own consumers then pass that message to
formatStringFromArgs, which fills each slot through `Debug.checkDefined`
(utilities.ts:8584). An undefined message fails there with `Debug Failure.`
instead of reporting `Cannot read file` or `Could not write file`.

Node 24.14.1 counterexample: a compiler host whose readFile throws
`{ message: 42 }`, then `{}`.
Before: onError receives the number 42, then undefined. getSourceFile returns
the empty source file both times. A thrown string or number also hands onError
undefined. A thrown null or undefined throws TypeError reading message out of
getSourceFile instead of calling onError. createWriteFileMeasuringIO's catch
gives the same results when the writer throws them. Adaptation 47's README
records these as the counterexamples that declined both sites.
After: onError receives the string "42", then the string "undefined", and so
does every other thrown value that reaches it. The consumer's
`Cannot read file 'a.ts': undefined.` replaces the Debug Failure. A thrown
null or undefined still throws the same TypeError.

Our patch: stage3/adapt/97-cli-axolotl-e/adapt.cjs, under Adamic's ruling of
Nov 1 (#axochk), "convert, don't cast". Widening onError to unknown was
measured and breaks five places in the tree, and a cast to the primitive is
refused, so the declaration stays `string` and the value is converted:

```diff
+import { propertyOf } from "./hostErrors.js";
-                onError(e.message);
+                onError(String(propertyOf(e, "message")));
```

at both sites. propertyOf is adaptation 94's helper. It returns exactly what
the plain read returns, and it throws V8's own TypeError on null and undefined.
The global String then turns what it returns into text. JavaScript changes at
exactly those two lines. A string message, an Error's for instance, passes
through unchanged, and a null or undefined thrown value throws the same
TypeError, with propertyOf's frame on top. The conversion keeps the meaning,
the message as text, and turns the crash into the diagnostic tsc meant. The
witnesses, the mutant and the transpile diff are in the adaptation's README.

### Draft GitHub issue

Title: getSourceFile and writeFile pass a non-string or undefined message to their `onError(message: string)` callback

Version: TypeScript 6.0.3, reproduced on Node 24.14.1.

Minimal reproduction:

```sh
mkdir tsc-onerror-message-repro
cd tsc-onerror-message-repro
npm install --no-audit --no-fund typescript@6.0.3
```

Save this as repro.cjs and run `node repro.cjs`:

```js
const ts = require("typescript");
const host = ts.createCompilerHost({});
for (const thrown of [{ message: 42 }, {}]) {
    host.readFile = () => { throw thrown; };
    host.getSourceFile("a.ts", ts.ScriptTarget.Latest, message => console.log(typeof message, message));
}
```

Expected: onError is declared `(message: string) => void` in CompilerHost
and WriteFileCallback, so either it receives a string or the declaration says
it may not.

Actual:

```text
number 42
undefined undefined
```

The compiler host's writeFile has the same catch: replacing `ts.sys.writeFile`
with a function throwing the same values prints the same two lines. A readFile
that throws null or undefined makes getSourceFile throw
`TypeError: Cannot read properties of null (reading 'message')` from inside
its catch.

Suggested upstream change: either convert the caught value to a string before
calling onError, for example with a message read that falls back to
`String(e)`, or declare the callback's parameter as what it receives. Either
way, the declaration and the call agree. This draft makes no claim about
which one upstream prefers. Our fork converts the message read with String,
so tsc's own consumers report `Cannot read file 'a.ts': undefined.` where an
undefined message used to fail Debug.checkDefined in formatStringFromArgs.

## 8. Raising the assertion level writes a cache entry over the assertion it cached

Status: removed in our fork by JavaScript-changing adaptation
99-cli-debug-assertion-level, under Adamic's Nov 5 ruling (#0kpp2zc,
#b1y988t); draft issue, not filed.
Upstream: TypeScript 6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8.
Location: src/compiler/debug.ts:170, in setAssertionLevel's restore loop,
`(Debug as any)[key] = cachedFunc;`. The cache is written by
shouldAssertFunction at debug.ts:189 and 190. These are stock source
coordinates.

Below its level, shouldAssertFunction stores `{ level, assertion: Debug[name] }`
in assertionCache and replaces `Debug[name]` with noop. When setAssertionLevel
raises the level, its loop means to put the assertion back, and its own
condition compares `Debug[key]` with `cachedFunc.assertion`. It assigns
`cachedFunc`, the entry, instead. After that, `Debug.assertNode` (or any of the
six assertion functions) is a plain object, and every call site throws.

tsc itself sets the level once, from `NODE_ENV`, before any assertion runs
(sys.ts:1973), so the command line never raises it after a cached call. An API
consumer that calls `ts.Debug.setAssertionLevel` after the compiler has run at
None does. Debug is internal and absent from typescript.d.ts, but it is on the
published module object.

Node 24.14.1 counterexample, with `NODE_ENV` unset: call
`Debug.assertNode(undefined, isIdentifier)`, then
`Debug.setAssertionLevel(AssertionLevel.Normal)`, then the same call again.
Before: `Debug.assertNode` is the object with keys `level` and `assertion`, and
the second call throws `TypeError: Debug.assertNode is not a function`. The
same holds for assertEachNode, assertNotNode, assertOptionalNode,
assertOptionalToken and assertMissingNode. After: the second call runs the
assertion and throws `Debug Failure. False expression: Unexpected node.`, as
it does when the level was Normal from the start.

Our patch: stage3/adapt/99-cli-debug-assertion-level/adapt.cjs. Adamic refuses
both namespace writes and ruled adapt, don't lower, so the fork removes the
patching instead of correcting line 170:

```diff
-    const assertionCache: Partial<Record<AssertionKeys, { level: AssertionLevel; assertion: AnyFunction; } | undefined>> = {};
     export function setAssertionLevel(level: AssertionLevel): void {
-        const prevAssertionLevel = currentAssertionLevel;
         currentAssertionLevel = level;
-        ...the restore loop...
     }
     function shouldAssertFunction<K extends AssertionKeys>(level: AssertionLevel, name: K): boolean {
-        if (!shouldAssert(level)) {
-            assertionCache[name] = { level, assertion: Debug[name] };
-            (Debug as any)[name] = noop;
-            return false;
-        }
-        return true;
+        return shouldAssert(level);
     }
```

The imports of `noop` and `getOwnKeys`, which only those lines named, go too,
and shouldAssertFunction's JSDoc no longer describes a cache. Each disabled
assertion now costs a call and a comparison on every call, where upstream paid
them once. No declaration changes. The witnesses, the mutant and the
transpile diff are in the adaptation's README.

Oracle counter (Correctness, #0kpp2zc condition 1; typescript/debug-assertion-level-census
aa84f129, stage3/census/debug-assertion-level): across the whole lane oracle,
setAssertionLevel was called 9 times, every one from sys.ts's module
initialization, each from level 0 to 1, and all 9 with an empty
assertionCache. So no call found a cached assertion to restore, and no noop
swap ran. The instrumented oracle kept 106,366 passing plus the one sanctioned
failure, with identities exactly main's: the adaptation is unobservable across
tsc's baselines.

### Draft GitHub issue

Title: Debug.setAssertionLevel replaces a cached assertion with its cache entry, so the next call throws "is not a function"

Version: TypeScript 6.0.3, reproduced on Node 24.14.1.

Minimal reproduction:

```sh
mkdir tsc-assertion-level-repro
cd tsc-assertion-level-repro
npm install --no-audit --no-fund typescript@6.0.3
```

Save this as repro.cjs and run `node repro.cjs` with `NODE_ENV` unset:

```js
const ts = require("typescript");
// With NODE_ENV unset, loading typescript set the assertion level to None.
ts.Debug.assertNode(undefined, ts.isIdentifier);
ts.Debug.setAssertionLevel(ts.AssertionLevel.Normal);
console.log(typeof ts.Debug.assertNode, JSON.stringify(Object.keys(ts.Debug.assertNode)));
ts.Debug.assertNode(undefined, ts.isIdentifier);
```

Expected: raising the level restores the assertion. The log prints
`function []`, and the last call fails the assertion with
`Debug Failure. False expression: Unexpected node.`, as it does when
setAssertionLevel runs before the first call.

Actual:

```text
object ["level","assertion"]
TypeError: ts.Debug.assertNode is not a function
```

Cause: setAssertionLevel's restore loop (src/compiler/debug.ts:170) assigns
the cache entry, `(Debug as any)[key] = cachedFunc;`, where its condition on
the line above compares against `cachedFunc.assertion`.

Suggested upstream change: assign `cachedFunc.assertion`. Alternatively, have
shouldAssertFunction return `shouldAssert(level)` without replacing anything,
which removes the cache and the restore loop. That is what our fork does. This
draft makes no claim about which one upstream prefers.

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

- takeWhile (99-correctness-core-takewhile): Adamic, task b5w3ycg, Nov 7 03:20Z: established-shape adaptations, Compiler accepts. JavaScript-changing prefix push loop admitted on the complete 220-call lane count: 0 accepted-range-hole calls, 0 result differences and 0 predicate-sequence differences. An accepted sparse hole is preserved by slice but becomes an own undefined slot with push; this witnessed class occurs 0 times in tsc inputs.

- arrayOf, 99-correctness-core-arrayof: JavaScript-changing; Adamic, task b5w3ycg, Nov 7 03:20Z: established-shape adaptations, Compiler accepts. Whole-lane count 860 non-negative integer calls, 0 counterexamples; negative, fractional, NaN and infinite calls each 0. Node -1, 1.5 and NaN differ outside the counted domain; see the adapter README.

- createQueue boxed slots, 99s amendment, F15/F16/F17: Adamic, task b5w3ycg, Nov 7 03:20Z: established-shape adaptations, Compiler accepts. Whole lane and all 398 bounded units each compare 10041 enqueues, 10123 dequeues and 13660 public isEmpty calls, with 0 differences; sparse initial arrays 0, own iterator overrides 0. Sparse and undefined-payload witnesses match stock. The deliberately changed iterator detector seed is outside the observed own-iterator class (0 calls). [Proof](../adapt/99-correctness-axolotl-s/QUEUE-BOX-PROOF.md).

- levenshteinWithMax filled number rows and six checked indexed reads: Adamic, task b5w3ycg, Nov 7 03:20Z: established-shape adaptations, Compiler accepts. Whole-lane differential count: 798965 calls, 0 counterexamples. Six assertions use adaptation 30's checked-indexed-read rule class, authorized by the user. See stage3/adapt/99-correctness-core-levenshtein/COUNT.json.

compareComparableValues, 99-correctness-core-compare: Adamic, task b5w3ycg, Nov 7 03:20Z: established-shape adaptations, Compiler accepts. Compiler amendment on b5w3ycg: the mixed arm is Debug.fail("compareComparableValues: a string compared with a number"), reached on 0 of the 144,697,378 calls. The retained whole-lane count has 0 mixed calls and 0 result differences; it was not rerun. Mixed Node witnesses now throw against the original coerced answer.

Adamic, task b5w3ycg, Nov 7 03:20Z: established-shape adaptations, Compiler accepts. MultiMap composition split between 94-correctness-multimap, the 95h createProgram amendment and 99-correctness-core-multimap; 4,355,837 original/free helper calls over all 398 lane units, 0 differences, 0 own-key enumerations, 0 sparse or empty buckets and 0 missing-key/absent-value removes (edge witnesses pass); both full applies and semantic checkpoints pass, 79 roots at 0 diagnostics, full oracle passes its 106,366/1/0 expectation, public declarations identical, exactly 3 core sites removed with none added and mutant restores all 3. No upstream issue filed.
Adamic's ruling (b), task b5w3ycg, Nov 7 11:06Z: getAllKeys moved to its only user, the fourslash harness, retaining Object.getPrototypeOf, which Adamic refuses (rulings.md:83). The harness runs on Node and is never lowered; if stage 3 ever needs it native, it goes back to Adamic. Measured: 6,566 calls, all from fourslashImpl.ts:461, 0 namespace reads.
99z amendment k220r1g: Compiler accepts the established-shape elementAt function declaration under Adamic's delegation (task b5w3ycg, Nov 7 03:20Z); 11 call sites, 1,936 calls, 0 differences, and 0 each of early reads, new, prototype, name or length reads, value uses, reassignments, and ts.elementAt reads.

Adamic, task b5w3ycg, Nov 7 19:16Z, road (b): 99-correctness-core-debug-messages is JavaScript-changing at four lazy debug.ts assertion failure messages, omitting only the test function name. Baseline whole-lane callback counts are 0/0/0/0 over all 398 units; four forced-failure controls fire before and after, same Error/conditions/throw point. Both full applies/builds, 79-root type checks, paired replay and public declaration byte checks pass; exactly four census sites disappear, none are added, and the restoring mutant brings one back. getFunctionName is unchanged; its standalone census unit remains for Adamic. No upstream issue filed.

- 99-correctness-core-cast-message: JavaScript-changing. Adamic, task b5w3ycg, Nov 7 03:20Z: established-shape adaptations, Compiler accepts; specific approval Nov 7 22:06:52Z. cast retains its condition, Debug.fail call and ${value}; only the predicate name disappears from its failure text. The full-lane caller count is 0, with a firing control, in fcb70da5e2d899df060d6f2521e35824442c9951. No real-call counterexample. Stopped draft, not landable: both full applies were stopped and the required census restoring-mutant proof cannot be demonstrated because cast is deferred with no recorded site before, after or restored.

- 99-correctness-core-getfunctionname-move: STOPPED draft before implementation because the move makes earlier adapter 45 refuse replay and changes 41's refusal. Adamic, task b5w3ycg, Nov 7 03:20Z: established-shape adaptations, Compiler accepts; specific approval Nov 7 22:06:52Z. By Adamic's word: deprecatedCompat ships in typescript.js, not tsc, so when typescript.js comes home after 1.0.0, getFunctionName's reflection (Function.prototype.toString) meets us there. The authorized road (b) guard relocation is pending; no helper has moved and no guard is relaxed.
