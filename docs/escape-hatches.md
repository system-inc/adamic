# Escape hatches: measurements and decisions

## Decisions

Accepted by @system_adamic on October 6, 2026. These are language decisions; the observations below remain measurements of the recorded main commit, not claims that these changes have landed.

1. **Downcasts:** runtime tag checks on tagged members plus checked field reads; untagged interface downcasts create transitive checked views. The October 7 ruling below supersedes the construction-or-refusal proposal.
2. **Non-null !:** a runtime nullish check that panics loudly and includes the expression's text.
3. **As written:** any, as unknown as, expando additions, Object.defineProperty and Function are refused. Upcasts and satisfies require proof. Type predicates and assertion functions require proof from their bodies or are refused. Bivariant methods require a proven contravariant relation or are refused.
4. **Definite assignment:** field!: and let x!: are refused for now; another branch lands that refusal. The target is proven initialization where flow establishes it, otherwise a loud read-before-assignment check like Adamic's temporal dead zone. This is a temporary refusal.
5. **Index signatures:** NotYet, not a permanent refusal. TypeScript's own source uses string-keyed record objects, and stage 3 must compile it unchanged. A sound map-backed representation needs a separate design. Until then, reject them with a diagnostic pointing at Map.

| Hatch | Today on main | Count / per 1,000 nonblank lines | Proposal | Added cost |
|---|---|---:|---|---|
| any | NotYet | 207 / 1.550 | Decided: refused | 0 |
| as upcast | Sound probe | 184 / 1.378 (relation candidates) | Decided: proven sound relation | 0 |
| as downcast | Tagged: checked; primitive: refused | 2,884 / 21.598 (relation candidates) | Decided: runtime tag check; refuse uncheckable targets | O(1) for one tag |
| as unknown as | Refused | 5 / 0.037 chains | Decided: refused | 0 |
| non-null ! | Refused | 938 / 7.024 | Decided: runtime nullish check; panic includes expression text | two comparisons + branch |
| field!: T | Unsound: NaN vs 1 | 0 / 0 | Decided: refused for now; target proven initialization or loud read-before-assignment check | 0 now; target 0 if proven, O(1) check otherwise |
| let x!: T | Unsound: NaN vs 1 | 11 / 0.082 | Decided: refused for now; target proven initialization or loud read-before-assignment check | 0 now; target 0 if proven, O(1) check otherwise |
| x is T | Refused | 327 / 2.449 | Decided: proven predicate body or refused | 0 added runtime |
| asserts x is T | Refused | 16 / 0.120 | Decided: proven assertion body or refused | 0 added runtime |
| expando assignment | Checker error | 0 confirmed; 10 candidates / 0.075 | Decided: refused shape addition | 0 |
| Object.defineProperty | Refused | 2 / 0.015 | Decided: refused | 0 |
| Object.assign(existing, ...) | Adding field: NotYet; fixed fields: sound | 1 / 0.007 (target candidate) | Proven fixed-shape writes; refuse additions | O(fields), existing work |
| bivariant interface methods | Unsafe relation refused | 1,468 / 10.994 method declarations | Decided: proven contravariant relation or refused | 0 |
| Function type | Call NotYet | 0 / 0 | Decided: refused | 0 |
| @ts-ignore | Unsound: 2 vs true | 0 / 0 | Refused | 0 |
| @ts-expect-error | Unsound: 2 vs true | 0 / 0 | Refused | 0 |
| satisfies | NotYet; bad member checker error | 0 / 0 | Decided: proven ordinary type relation | 0 |
| ?. on non-optional | String: sound; number: NotYet | 107 / 0.801 (checker candidates) | Proven ordinary optional operation | 0 if redundant |
| index read typed without undefined | Checker error; signature refused | 214 / 1.603 reads; 11 / 0.082 signatures | Decided: NotYet; diagnostic points at Map pending map-backed design | 0 now; future representation cost to be designed |

## Scope and status

**Observation.** Measured on main commit `5d4c8012a0877094134e6c6bac367ff68f9313e8`, checked out as `codex/escape-hatches`, October 6, 2026. Only this document, repro programs and their measurement tools/evidence are changed. No proposal is implemented. Counts are source-site counts, not execution frequencies. “Refused” in the table means the command did not emit C; the detailed evidence distinguishes checker errors, policy Refused and capability NotYet. Emitting C and failing in clang is a separate outcome.

**Observation.** Four runnable probes silently disagree with Node: uninitialized class and local definite assignment, and each suppression directive applied to a nonliteral number-to-boolean assignment. All finish with exit 0 and empty stderr. A discriminated-union cast instead intentionally adds a check absent from TypeScript source; its failed check is loud, not a silent miscompile.

**Inference.** Definite-assignment and directive handling need compiler follow-up before they can satisfy the promise that types are true. A clang diagnostic on one ill-typed source is not evidence that directives are safe. These findings remain open in this design-only unit.

## Reproduction and corpus counts

**Observation.** Required setup command:

```sh
git fetch origin && git checkout -b codex/escape-hatches origin/main
bash cloud/setup.sh > /tmp/escape-setup.log 2>&1
source /workspace/adamic-tools/env.sh
nproc
```

The script selected `/workspace/adamic-tools/env.sh`, not `/opt/adamic-tools/env.sh`. Its timing lines were:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (74s)
setup: done in 74s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

`nproc` printed `5`. Go was 1.27.1, clang 20.1.8, Node v24.19.0. Setup exited 0.

**Observation.** The existing stage 1 scanner survey pins upstream TypeScript v6.0.3, commit `050880ce59e30b356b686bd3144efe24f875ebc8` (`stage1/typescript/scanner/scanner_test.go:compilerSource`). The cohere/TypeScript submodule is the Go port and has no original src/compiler directory. Recount commands, from the repository root:

```sh
git clone --quiet --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git /tmp/escape-typescript
npm install --prefix /tmp/escape-survey typescript@6.0.3 > /tmp/escape-npm.log 2>&1
NODE_PATH=/tmp/escape-survey/node_modules node docs/escape-hatches/count.cjs /tmp/escape-typescript/src/compiler > /tmp/escape-counts.json 2> /tmp/escape-counts.log
python3 docs/escape-hatches/observe.py /tmp/escape-observations > /tmp/escape-observe.log 2>&1
```

The AST walk counted all 38 direct .ts files in src/compiler, declarations and comments included in the nonblank-line denominator, with tests outside this directory excluded. There are **133,533 nonblank lines**, **145,474 physical lines**. The probe runner invokes exactly `go run ./cmd/adamic c <file>`, then `go run ./cmd/adamic build <file> -o <binary>` for every emitted program, then runs the binary when build succeeds. It copies source .a files to .ts in scratch so Node runs the original source with type stripping, independently of Adamic's JavaScript backend. All command output is captured in files, never piped.

[Count tool](escape-hatches/count.cjs), [count evidence](escape-hatches/counts.json), [probe runner](escape-hatches/observe.py), and [exact observations](escape-hatches/observations.json) are retained. Successful C is in the scratch observations directory; generated C and binaries are not committed.

### Count definitions and limits

**Observation.** The supplied anchors do not all reproduce:

| Metric | This recount, nonblank denominator | Physical-line denominator | Supplied anchor |
|---|---:|---:|---:|
| casts excluding as const | 3,473 / 26.009 | 23.874 | 23.1 |
| non-null expressions | 938 / 7.024 | 6.448 | 6.4 |
| clean files | 2 of 38, 5.263% | same | 12% |

“Clean” here means no any keyword, non-const assertion, non-null expression, predicate/assertion signature or method signature. The seven syntax families overlap (a double assertion contains two cast nodes). No claim is made that the historical clean-file definition was identical. **Inference.** Physical lines explain the non-null anchor to one decimal; they do not fully explain the cast or clean-file anchors. The prior survey's exact script and file selection were not available in this unit, so the difference is unresolved rather than silently normalized.

**Observation.** any counts explicit AnyKeyword nodes, including type arguments and declarations; it does not count inferred any. as const has 16 separate occurrences (0.120), excluded from 3,473 casts. as unknown as counts outer chain nodes, five total. The syntax cast count also recognizes angle assertions, though this corpus uses as for all counted non-const assertions. The cast-relation buckets sum to 3,473: 184 assignable source-to-target, 2,884 assignable target-to-source, 300 neither, and 105 involving source any/unknown or target any. The five double assertions overlap these buckets.

**Inference.** These are TypeScript relation candidates, not Adamic proofs. A generic, mutable or structural relation accepted by TypeScript can still be unsound. “2,884 downcasts” does not mean 2,884 runtime-checkable discriminant casts. The checker is created with strict mode and noUncheckedIndexedAccess disabled for this survey; it resolves imports from the source tree with standard ESNext library types, without installing upstream Node host typings. Host-dependent relation and optionality counts are therefore candidates; syntactic counts do not depend on those resolutions. Full upstream type checking and a per-cast reifiability audit remain future work.

**Observation.** The method count is 1,468 MethodSignature nodes, including nested type literals. It measures declarations exposed to TypeScript's method-parameter rule, not 1,468 witnessed unsafe assignments. Guards and assertion functions are counted separately, 327 and 16. The zero counts for field !, Function references, directives and satisfies are measured zeros, not omitted searches.

**Observation.** Expando additions cannot be counted exactly from syntax: a property assignment may set an existing property or add one depending on runtime presence and aliases. The tool found ten unproven property-write candidates (missing declared property, including any-typed receivers); their locations are in counts.json. Manual inspection found writes to raw config fields (five), profile callFrame.url (two), ImportTypeNode.isTypeOf, Error.stackTraceLimit and this.links (one each). Several are updates to already valid fields through any or missing host types. **No plain expando addition was confirmed**; 0 to 10 describes only the discovered dot-write candidates, not a global bound on runtime additions or indirect writes. Object.defineProperty has two exact direct-call sites: hidden configFile and a debug getter. Indirect aliases of Object methods are not counted.

**Observation.** Object.assign has one direct call whose first argument is not an object literal: setObjectAllocator's objectAllocator target (utilities.ts:8578). It updates an existing allocator; no added property was established. The count is for use of the API on an existing-target candidate, not evidence of a soundness hole at that site. Candidate discovery does not follow helpers or alias calls.

**Observation.** Optional-on-nonoptional counts 107 question-dot operations whose immediate receiver the checker reports as neither nullish nor any/unknown, out of 971 optional operations. Index counts include 11 signature declarations and 214 element-read sites backed by non-nullish index value types when noUncheckedIndexedAccess is disabled, excluding arrays, tuples, explicitly declared literal keys and simple assignment destinations. Compound reads still count. It does not prove any particular key absent. Declaration counts and read counts are different units.

## Decision rationale and remaining recommendations

The Decisions section records the accepted rulings. This section explains their proof obligations and retains the other recommendations as **proposal/inference**. Neither decisions nor recommendations describe new compiler behavior implemented by this document. “Proven” always means acceptance requires the named proof; failed proof is a compile-time refusal with the obligation named. No proof may rely on the assertion syntax or a diagnostic-suppression directive.

### Failure contract

For checked casts and non-null unwraps, choose **panic**, not a catchable TypeError. They claim an invariant, and a failure must not let execution continue with an invalid native representation. This matches the approved cast and panic contract. The compiler must emit the same check in native and JavaScript backends, evaluate the operand once, flush preceding stdout, print the exact line specified below to stderr and exit **70**. No catch or finally runs after a panic. Successful executions exit 0 unless the program itself fails.

A catchable TypeError would require a deliberately different contract: unwind ownership and let callers recover before producing a narrowed value. This unit does not propose that change. Refused or proven cases have no runtime failure message or failure exit code because no check is inserted.

### any: refused

Refuse explicit and flowing inferred any at the gate, including unsafe calls and member access. An unused any annotation must not become an authorization to erase proof later. There is no finite runtime check that certifies arbitrary future operations on any; dynamic tagging alone would not prove callable signatures or mutable alias invariants.

Stage 1 needs unknown plus typeof/discriminant narrowing at dynamic boundaries, concrete type arguments in generic code, and constructors/builders for the known compiler-node variants. The 207 keyword sites are an audit list, not 207 necessarily unsafe value operations.

### Casts: proven upcasts, checked tagged downcasts, refused unrelated assertions

An upcast needs Adamic's independent sound relation: immutable covariance, mutable invariance, nominal class identity, safe function variance and optional-field presence/type compatibility. TypeScript assignability alone is insufficient. Erase a proven cast, with zero runtime cost.

For a downcast from a represented union to its represented member/subunion, first establish that each runtime discriminator uniquely selects a member whose fields have already been proven at construction and every write. Check the tag against the target tag set before narrowing. One-tag cost is one load/comparison and a branch; k tags cost up to k comparisons unless a shared tag mask is available. A primitive tag check could support number|string in a future representation, but it is outside the accepted set until independently validated. A cast already established by flow narrowing erases.

Keep the existing exact message template:
`adamic: panic: cast failed: this <source type> is not a <target type>\n`.
For the retained mutant it is `adamic: panic: cast failed: this A | B is not a A\n`, exit 70. Type names are compile-time strings; never evaluate user conversion code to construct the failure message.

A structural unknown-to-interface cast, generic predicate, callback-signature cast, mutable widening or unrelated double assertion is refused. Do not use “object has fields” as a proof of a mutable interface: later writes through aliases could invalidate it, and callable behavior cannot be certified by a typeof function check. Stage 1 needs explicit discriminated node types, checked constructors and inline narrowing for refused casts. The 300 neither-way and 105 unchecked relation candidates require individual audits; the 2,884 downcast candidates also need a tag/reifiability audit.

### Non-null !: runtime check

**Decision.** Replace the refusal with a runtime nullish check that includes the expression's text in its panic. Evaluate the operand once; reject **both null and undefined**, not other falsy values (0, false and empty string pass). Return the exact original value after success. If independent flow/alias analysis proves non-nullish at the use, erase the check.

Failure prints `adamic: panic: non-null assertion failed: <expression text> is null or undefined\n`, exit 70. The expression text is the original assertion expression captured at compile time, not a runtime value conversion. For the retained probe, the line is `adamic: panic: non-null assertion failed: m.get('missing')! is null or undefined\n`. Cost is at most two tag comparisons and one branch, no allocation on success. Mutation through calls/captures must invalidate the proof before elision. The current missing-map-key probe is the failure input; a future implementation must additionally test null, each falsy non-nullish input, side-effectful operands and invalidation across calls. This decided check has not been implemented or mutant-tested in this unit.

### Definite assignment: refused for now; target proof or read check

**Decision.** Refuse field!: and let x!: for now. Another branch lands that refusal; this document does not implement it. The refusal is temporary. The target is to erase checks where flow proves initialization and otherwise insert a loud read-before-assignment check, following Adamic's temporal-dead-zone approach.

Ignore ! while proving initialization. For a local, every read must be dominated on every executable control-flow path by a compatible assignment, including loop, closure and exception paths. Captured reads need proof valid at the actual read or retain the runtime initialization check. For a class, account for compatible field assignments, normal constructor returns, inherited construction order and reads through this escapes or callbacks during construction. An unproven read must check initialized state before loading a value as T. The ! syntax itself proves nothing.

A proven read costs zero additional runtime work; an unproven read needs initialized-state tracking and an O(1) test and branch. The exact state representation and failure diagnostic need an implementation design. The accepted target requires a loud read-before-assignment failure, not a silent default value.

Do not initialize missing numbers to zero: the source value is undefined. Stage 1 can use constructor assignments, initialized locals, explicit T|undefined state narrowed before reads, or builders that produce a complete object while the temporary refusal stands. The two assigned controls pass on the measured main; removing their assignments produces the retained silent-miscompile mutants. The interim refusal must reject the ! declarations, and the target must either prove each read safe or stop it with the initialization check.

### Guards and assertion functions: proven, otherwise refused

A predicate verifier must establish both directions: every true return implies x is T, and every false return implies x is outside T, because TypeScript narrows the false branch too. Reuse independently established typeof, literal equality, discriminants and nominal instanceof facts; require reachable return paths, not a syntactic whitelist alone. Compose verified helper summaries to a fixed point. Mutation/alias effects invalidate facts, and recursive/opaque helpers without a converged proof are refused. A guard testing a stricter property, such as “number and positive”, cannot claim x is number with a false-branch exclusion.

An assertion verifier establishes x is T on every normal return; failing paths must not return. Do not trust an asserts annotation as its own postcondition. Summaries must account for mutations and aliases as above. Both verifiers add zero runtime work beyond the body the programmer wrote. There is no proposed runtime validator for arbitrary T.

Stage 1 must inline narrowing or rewrite unverified helpers to boolean/result-returning functions that callers narrow explicitly. assert helpers can use a checked branch plus a reasoned panic; the predicate or assertion signature alone never authorizes the narrowing. The 327 guards and 16 assertion signatures need a body audit, not blanket acceptance.

### Shape mutation: refuse additions and descriptors; prove fixed-shape Object.assign

Refuse a property addition to a published object's shape. Stage 1 must allocate complete shapes, use explicit optional fields for staged construction, or use Map for dynamic keys. Property assignments to declared fields remain ordinary checked writes. TypeScript's ordinary expando source already errors in these probes; bypassing it with any or a directive must not evade the policy.

Refuse Object.defineProperty, including descriptors that appear to merely assign: getters, enumerability and writability are observable semantics. Stage 1 needs explicit configuration/debug fields and methods; nonenumerable metadata needs a separate table or a deliberate API with specified reflection behavior, not silent conversion to an enumerable field.

Allow Object.assign into an existing object only when every source's actual own enumerable fields and descriptor behavior are known, all writes fit the existing target shape and field types, and alias/ownership/cycle checks hold. Preserve left-to-right evaluation and overwrite order. Declared structural interfaces alone do not establish the absence or types of hidden fields. This is a **proven** ordinary bulk store, not a check that adds shape flexibility. Added soundness cost is zero beyond O(fields) copy work and existing ownership operations. Refuse unproven sources or additions; stage 1 can replace the allocator update with explicit known-field writes or a complete new allocator object.

### Bivariant methods: proven contravariant relation

Keep method syntax only if every function relation uses contravariant parameters and covariant results, with invariant mutable captured/container types as applicable. Do not permit TypeScript's bivariance escape. At an incompatible assignment, refuse before lowering. Runtime typeof or a call wrapper cannot prove the callee's accepted input set. Runtime cost is zero.

Stage 1 rewrites interface method signatures to properties holding functions, or widens implementations to accept and explicitly narrow the advertised parameter type. The declaration count measures porting exposure; main already refuses the dangerous assignment in the probe.

### Function type and directives: refused

Refuse the untyped Function abstraction; a function value needs a concrete checked signature. This is distinct from new Function's requirement for a runtime compiler. Stage 1 writes specific function types or unknown with a verified boundary adapter; checking only typeof function is insufficient.

Refuse both directives before TypeScript suppression is applied (comments must be parsed as comments, not strings). No runtime checks are sufficient substitutes for all the diagnostics a directive can hide. Stage 1 fixes the ignored error, represents the actual union, or moves intentional negative compile tests into a dedicated external test harness. Do not remove a directive and pretend success without rerunning the checker.

### satisfies and optional chaining: proven ordinary operations

satisfies is not an assertion hatch. It checks a relation while retaining the expression's own type, with contextual typing still allowed; it cannot force an incompatible value to pass. Accept only with the same sound relations and constructor/write obligations as ordinary assignment, then erase it at zero runtime cost. Main's checker rejects the negative control, but its lowering does not yet support even the sound expression. Stage 1 can temporarily remove a redundant satisfies or use a checked annotation, accounting for the annotation's changed inferred type.

Optional chaining on a receiver proven non-nullish is redundant, not a type lie. Preserve single evaluation, property/call order and method this binding; ordinary optional operations must yield undefined when their receiver is nullish. Elide the nullish branch only with a proof valid at that exact read, including aliases and calls. Zero added cost when redundant; otherwise the ordinary optional-chain branch applies. Main's number result gap is an implementation limitation, not a reason to classify this syntax as unsound. Stage 1 can use an ordinary read only where receiver non-nullishness is independently established.

### Index signatures: NotYet pending a map-backed design

**Decision.** Index signatures are NotYet, not refused forever. TypeScript's own compiler uses string-keyed record objects, and stage 3 must compile its source unchanged. A sound map-backed representation needs its own later design. Until then, reject index signatures with a NotYet diagnostic pointing at Map as the current alternative. The retained policy-refusal text below is a historical observation, not the decided diagnostic classification.

A finite set of existing keys is not a proof that every string key exists. The current checker enables noUncheckedIndexedAccess, so the unsafe read is already rejected; the fallback control reaches the recorded policy refusal. Stage 1 can temporarily use Map and treat get as T|undefined, narrowing or using a reasoned panic. This workaround is not a permanent rewrite requirement for stage 3.

The later design must preserve TypeScript object behavior, including integer-keys-first key ordering where observable; Map insertion order alone is insufficient. It must provide sound missing-key reads even when TypeScript's tsconfig omits noUncheckedIndexedAccess, and preserve aliasing, writes and ownership without a garbage collector. Representation cost and the exact missing-key check contract remain for that design. No index-signature representation is implemented here.

## Current behavior, exact probe evidence

**Observation.** The following blocks quote compiler diagnostics exactly as recorded, including absolute paths. `exit status 1` is go run's wrapper line; the underlying compile command returned failure. Node numbers below are process exit codes. Every successful native run reports stdout and stderr, so an unsound result cannot hide behind a compile-success label. These are isolated minimal witnesses, not a claim that every use of the same syntax follows the same path.

### any

[any.a](escape-hatches/any.a):

```ts
const x: any = 'wrong';
const n: number = x;
console.log(`${n + 1}`);
```

C: refused, exit 1. Exact diagnostic:

```text
adamic: /workspace/adamic/docs/escape-hatches/any.a:1:7: stage 0 can't lower a value of type any yet
exit status 1
```

Node source: exit 0; stdout "wrong1\n"; stderr "".

### as upcast

[cast-up.a](escape-hatches/cast-up.a):

```ts
const x: { readonly n: number; readonly s: string } = { n: 1, s: 'ok' };
const y = x as { readonly n: number };
console.log(`${y.n}`);
```

C and native build: exit 0; **compiled and sound for this input**. Native exit 0; stdout "1\n"; stderr "".

Node source: exit 0; stdout "1\n"; stderr "".

### as primitive downcast

[cast-down.a](escape-hatches/cast-down.a):

```ts
function f(x: number | string): number { return x as number; }
console.log(`${f(2) + 1}`);
```

C: refused, exit 1. Exact diagnostic:

```text
adamic: /workspace/adamic/docs/escape-hatches/cast-down.a:1:49: Adamic 0.1 refuses a cast the runtime can't check; narrow it instead (===, typeof, a discriminant), or cast a discriminated union to its members (adamic/no-unchecked-cast)
exit status 1
```

Node source: exit 0; stdout "3\n"; stderr "".

### as primitive downcast wrong input

[cast-down-mutant.a](escape-hatches/cast-down-mutant.a):

```ts
function f(x: number | string): number { return x as number; }
console.log(`${f('wrong') + 1}`);
```

C: refused, exit 1. Exact diagnostic:

```text
adamic: /workspace/adamic/docs/escape-hatches/cast-down-mutant.a:1:49: Adamic 0.1 refuses a cast the runtime can't check; narrow it instead (===, typeof, a discriminant), or cast a discriminated union to its members (adamic/no-unchecked-cast)
exit status 1
```

Node source: exit 0; stdout "wrong1\n"; stderr "".

### as tagged downcast

[cast-tag.a](escape-hatches/cast-tag.a):

```ts
type A = { readonly kind: 'a'; readonly n: number };
type B = { readonly kind: 'b'; readonly n: number };
function f(x: A | B): A { return x as A; }
console.log(`${f({ kind: 'a', n: 2 }).n}`);
```

C and native build: exit 0; **compiled and sound for this input**. Native exit 0; stdout "2\n"; stderr "".

Node source: exit 0; stdout "2\n"; stderr "".

### as tagged downcast wrong tag

[cast-tag-mutant.a](escape-hatches/cast-tag-mutant.a):

```ts
type A = { readonly kind: 'a'; readonly n: number };
type B = { readonly kind: 'b'; readonly n: number };
function f(x: A | B): A { return x as A; }
console.log(`${f({ kind: 'b', n: 2 }).n}`);
```

C and native build: exit 0; **compiled with a loud runtime check (source cast has no such check)**. Native exit 70; stdout ""; stderr "adamic: panic: cast failed: this A | B is not a A\n".

Node source: exit 0; stdout "2\n"; stderr "".

### as unknown as

[cast-unrelated.a](escape-hatches/cast-unrelated.a):

```ts
const n = 'wrong' as unknown as number;
console.log(`${n + 1}`);
```

C: refused, exit 1. Exact diagnostic:

```text
adamic: /workspace/adamic/docs/escape-hatches/cast-unrelated.a:1:11: Adamic 0.1 refuses a cast the runtime can't check; narrow it instead (===, typeof, a discriminant), or cast a discriminated union to its members (adamic/no-unchecked-cast)
exit status 1
```

Node source: exit 0; stdout "wrong1\n"; stderr "".

### non-null !

[non-null.a](escape-hatches/non-null.a):

```ts
const m = new Map<string, number>();
const n = m.get('missing')!;
console.log(`${n + 1}`);
```

C: refused, exit 1. Exact diagnostic:

```text
adamic: /workspace/adamic/docs/escape-hatches/non-null.a:2:11: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case
exit status 1
```

Node source: exit 0; stdout "NaN\n"; stderr "".

### field definite assignment

[definite-field.a](escape-hatches/definite-field.a):

```ts
class Box { n!: number; }
console.log(`${new Box().n + 1}`);
```

C and native build: exit 0; **compiled and unsound**. Native exit 0; stdout "1\n"; stderr "".

Node source: exit 0; stdout "NaN\n"; stderr "".

### field initialization control

[definite-field-assigned.a](escape-hatches/definite-field-assigned.a):

```ts
class Box { n!: number; constructor() { this.n = 2; } }
console.log(`${new Box().n + 1}`);
```

C and native build: exit 0; **compiled and sound for this input**. Native exit 0; stdout "3\n"; stderr "".

Node source: exit 0; stdout "3\n"; stderr "".

### local definite assignment

[definite-local.a](escape-hatches/definite-local.a):

```ts
let n!: number;
console.log(`${n + 1}`);
```

C and native build: exit 0; **compiled and unsound**. Native exit 0; stdout "1\n"; stderr "".

Node source: exit 0; stdout "NaN\n"; stderr "".

### local initialization control

[definite-local-assigned.a](escape-hatches/definite-local-assigned.a):

```ts
let n!: number;
n = 2;
console.log(`${n + 1}`);
```

C and native build: exit 0; **compiled and sound for this input**. Native exit 0; stdout "3\n"; stderr "".

Node source: exit 0; stdout "3\n"; stderr "".

### type guard

[guard.a](escape-hatches/guard.a):

```ts
function isNumber(x: number | string): x is number { return true; }
function f(x: number | string): void { if (isNumber(x)) { console.log(`${x + 1}`); } }
f('wrong');
```

C: refused, exit 1. Exact diagnostic:

```text
adamic: /workspace/adamic/docs/escape-hatches/guard.a:1:40: Adamic 0.1 refuses a type predicate; narrow where you use it, with ===, typeof or instanceof (adamic/no-type-predicate)
exit status 1
```

Node source: exit 0; stdout "wrong1\n"; stderr "".

### assertion function

[assertion.a](escape-hatches/assertion.a):

```ts
function assertNumber(x: number | string): asserts x is number {}
function f(x: number | string): void { assertNumber(x); console.log(`${x + 1}`); }
f('wrong');
```

C: refused, exit 1. Exact diagnostic:

```text
adamic: /workspace/adamic/docs/escape-hatches/assertion.a:1:44: Adamic 0.1 refuses a type predicate; narrow where you use it, with ===, typeof or instanceof (adamic/no-type-predicate)
exit status 1
```

Node source: exit 0; stdout "wrong1\n"; stderr "".

### expando assignment

[expando.a](escape-hatches/expando.a):

```ts
function f() { return {}; }
const x = f();
x.n = 1;
console.log(`${x.n}`);
```

C: refused, exit 1. Exact diagnostic:

```text
docs/escape-hatches/expando.a:3:3: error TS2339: Property 'n' does not exist on type '{}'.
docs/escape-hatches/expando.a:4:18: error TS2339: Property 'n' does not exist on type '{}'.
exit status 1
```

Node source: exit 0; stdout "1\n"; stderr "".

### direct expando assignment

[expando-inferred.a](escape-hatches/expando-inferred.a):

```ts
const x = {};
x.n = 1;
console.log(`${x.n}`);
```

C: refused, exit 1. Exact diagnostic:

```text
docs/escape-hatches/expando-inferred.a:2:3: error TS2339: Property 'n' does not exist on type '{}'.
docs/escape-hatches/expando-inferred.a:3:18: error TS2339: Property 'n' does not exist on type '{}'.
exit status 1
```

Node source: exit 0; stdout "1\n"; stderr "".

### Object.defineProperty

[define-property.a](escape-hatches/define-property.a):

```ts
const x = {};
Object.defineProperty(x, 'n', { value: 1 });
console.log('done');
```

C: refused, exit 1. Exact diagnostic:

```text
adamic: /workspace/adamic/docs/escape-hatches/define-property.a:2:1: Adamic 0.1 refuses Object.defineProperty; property descriptors can change the presence, type or access behavior of fields; Adamic fields have a fixed shape and are plain loads and stores
exit status 1
```

Node source: exit 0; stdout "done\n"; stderr "".

### Object.assign adding field

[assign-existing.a](escape-hatches/assign-existing.a):

```ts
const x = { n: 1 };
Object.assign(x, { n: 2, s: 'new' });
console.log(`${x.n}`);
```

C: refused, exit 1. Exact diagnostic:

```text
adamic: /workspace/adamic/docs/escape-hatches/assign-existing.a:2:18: stage 0 can't lower Object.assign adding a field to its target's fixed shape yet
exit status 1
```

Node source: exit 0; stdout "2\n"; stderr "".

### Object.assign fixed fields

[assign-fixed.a](escape-hatches/assign-fixed.a):

```ts
const x = { n: 1 };
Object.assign(x, { n: 2 });
console.log(`${x.n}`);
```

C and native build: exit 0; **compiled and sound for this input**. Native exit 0; stdout "2\n"; stderr "".

Node source: exit 0; stdout "2\n"; stderr "".

### bivariant method

[bivariant.a](escape-hatches/bivariant.a):

```ts
interface Animal { readonly n: number; }
interface Dog extends Animal { readonly bark: () => string; }
interface Handler { run(x: Animal): string; }
const dogHandler = { run: (x: Dog): string => x.bark() };
const h: Handler = dogHandler;
console.log(h.run({ n: 1 }));
```

C: refused, exit 1. Exact diagnostic:

```text
adamic: /workspace/adamic/docs/escape-hatches/bivariant.a:5:20: Adamic 0.1 refuses a function taking Dog seen as one taking Animal (tsc relates a method's parameters both ways), so it can be handed what it can't take; write the method as a property holding a function (handle: (animal: Animal) => void), which tsc checks one way, or take the wider type in the method (method-signature-style)
exit status 1
```

Node source: exit 1; stdout ""; stderr contains `TypeError: x.bark is not a function`; full engine stack is in observations.json.

### Function type

[function-type.a](escape-hatches/function-type.a):

```ts
const f: Function = (x: number): number => x + 1;
console.log(`${f('wrong')}`);
```

C: refused, exit 1. Exact diagnostic:

```text
adamic: /workspace/adamic/docs/escape-hatches/function-type.a:2:16: stage 0 can't lower a call to an Identifier yet
exit status 1
```

Node source: exit 0; stdout "wrong1\n"; stderr "".

### @ts-ignore nonliteral

[ts-ignore-call.a](escape-hatches/ts-ignore-call.a):

```ts
function value(): number { return 2; }
// @ts-ignore
const b: boolean = value();
console.log(`${b}`);
```

C and native build: exit 0; **compiled and unsound**. Native exit 0; stdout "true\n"; stderr "".

Node source: exit 0; stdout "2\n"; stderr "".

### @ts-expect-error nonliteral

[ts-expect-error-call.a](escape-hatches/ts-expect-error-call.a):

```ts
function value(): number { return 2; }
// @ts-expect-error
const b: boolean = value();
console.log(`${b}`);
```

C and native build: exit 0; **compiled and unsound**. Native exit 0; stdout "true\n"; stderr "".

Node source: exit 0; stdout "2\n"; stderr "".

### @ts-ignore incompatible representation

[ts-ignore.a](escape-hatches/ts-ignore.a):

```ts
// @ts-ignore
const n: number = 'wrong';
console.log(`${n + 1}`);
```

C: emitted, exit 0. Native build failed, exit 1; this is a leak past the front end, not a clean policy refusal. Exact build output:

```text
adamic: native: clang failed: exit status 1
/tmp/adamic-gate/adamic-build-4219602995/main.c:15:20: error: assigning to 'double' from incompatible type 'adamic_string *' (aka 'struct adamic_string *')
   15 |         adamic_global_0_n = &adamic_string_0;
      |                           ^ ~~~~~~~~~~~~~~~~
1 error generated.

exit status 1
```

Node source: exit 0; stdout "wrong1\n"; stderr "".

### @ts-expect-error incompatible representation

[ts-expect-error.a](escape-hatches/ts-expect-error.a):

```ts
// @ts-expect-error
const n: number = 'wrong';
console.log(`${n + 1}`);
```

C: emitted, exit 0. Native build failed, exit 1; this is a leak past the front end, not a clean policy refusal. Exact build output:

```text
adamic: native: clang failed: exit status 1
/tmp/adamic-gate/adamic-build-3299588801/main.c:15:20: error: assigning to 'double' from incompatible type 'adamic_string *' (aka 'struct adamic_string *')
   15 |         adamic_global_0_n = &adamic_string_0;
      |                           ^ ~~~~~~~~~~~~~~~~
1 error generated.

exit status 1
```

Node source: exit 0; stdout "wrong1\n"; stderr "".

### @ts-ignore literal numeric conversion

[ts-ignore-boolean.a](escape-hatches/ts-ignore-boolean.a):

```ts
// @ts-ignore
const b: boolean = 2;
console.log(`${b}`);
```

C: emitted, exit 0. Native build failed, exit 1; this is a leak past the front end, not a clean policy refusal. Exact build output:

```text
adamic: native: clang failed: exit status 1
/tmp/adamic-gate/adamic-build-3428376133/main.c:13:23: error: implicit conversion from 'double' to 'bool' changes value from 2 to true [-Werror,-Wliteral-conversion]
   13 |         adamic_global_0_b = (0x1p+01);
      |                           ~  ^~~~~~~
1 error generated.

exit status 1
```

Node source: exit 0; stdout "2\n"; stderr "".

### @ts-expect-error literal numeric conversion

[ts-expect-error-boolean.a](escape-hatches/ts-expect-error-boolean.a):

```ts
// @ts-expect-error
const b: boolean = 2;
console.log(`${b}`);
```

C: emitted, exit 0. Native build failed, exit 1; this is a leak past the front end, not a clean policy refusal. Exact build output:

```text
adamic: native: clang failed: exit status 1
/tmp/adamic-gate/adamic-build-1831291471/main.c:13:23: error: implicit conversion from 'double' to 'bool' changes value from 2 to true [-Werror,-Wliteral-conversion]
   13 |         adamic_global_0_b = (0x1p+01);
      |                           ~  ^~~~~~~
1 error generated.

exit status 1
```

Node source: exit 0; stdout "2\n"; stderr "".

### satisfies

[satisfies.a](escape-hatches/satisfies.a):

```ts
const x = { n: 1 } satisfies { readonly n: number };
console.log(`${x.n}`);
```

C: refused, exit 1. Exact diagnostic:

```text
adamic: /workspace/adamic/docs/escape-hatches/satisfies.a:1:11: stage 0 can't lower a SatisfiesExpression yet
exit status 1
```

Node source: exit 0; stdout "1\n"; stderr "".

### satisfies negative control

[satisfies-mutant.a](escape-hatches/satisfies-mutant.a):

```ts
const x = { n: 'wrong' } satisfies { readonly n: number };
console.log(x.n);
```

C: refused, exit 1. Exact diagnostic:

```text
docs/escape-hatches/satisfies-mutant.a:1:13: error TS2322: Type 'string' is not assignable to type 'number'.
exit status 1
```

Node source: exit 0; stdout "wrong\n"; stderr "".

### optional number

[optional-chain.a](escape-hatches/optional-chain.a):

```ts
const x = { n: 1 };
console.log(`${x?.n}`);
```

C: refused, exit 1. Exact diagnostic:

```text
adamic: /workspace/adamic/docs/escape-hatches/optional-chain.a:2:16: stage 0 can't lower ?. to a number, which would be number | undefined yet
exit status 1
```

Node source: exit 0; stdout "1\n"; stderr "".

### optional string

[optional-chain-string.a](escape-hatches/optional-chain-string.a):

```ts
const x = { s: 'ok' };
console.log(x?.s);
```

C and native build: exit 0; **compiled and sound for this input**. Native exit 0; stdout "ok\n"; stderr "".

Node source: exit 0; stdout "ok\n"; stderr "".

### index-signature unchecked demand

[index-signature.a](escape-hatches/index-signature.a):

```ts
const x: { [key: string]: number } = {};
const n: number = x['missing'];
console.log(`${n + 1}`);
```

C: refused, exit 1. Exact diagnostic:

```text
docs/escape-hatches/index-signature.a:2:7: error TS2322: Type 'number | undefined' is not assignable to type 'number'.
  Type 'undefined' is not assignable to type 'number'.
exit status 1
```

Node source: exit 0; stdout "NaN\n"; stderr "".

### index-signature fallback

[index-signature-optional.a](escape-hatches/index-signature-optional.a):

```ts
const x: { [key: string]: number } = {};
console.log(`${x['missing'] ?? -1}`);
```

C: refused, exit 1. Exact diagnostic:

```text
adamic: /workspace/adamic/docs/escape-hatches/index-signature-optional.a:1:12: Adamic 0.1 refuses an index signature; use a Map, which keeps keys in the order they were added
exit status 1
```

Node source: exit 0; stdout "-1\n"; stderr "".


## Verification, mutants and remaining limits

**Observation.** These source mutants were executed:

| Mutant | Main result | What caught it |
|---|---|---|
| tagged cast argument kind a changed to b | builds, native panic/70; Node source prints 2/0 | inserted cast tag check, exact message above |
| primitive cast argument 2 changed to wrong string | refused at cast in both versions | existing unchecked-cast refusal; no runtime proof claimed |
| initialized class field assignment removed | native 1, Node NaN, both 0 | independent stdout comparison exposes missing proof; main does not catch it |
| initialized local assignment removed | native 1, Node NaN, both 0 | independent stdout comparison exposes missing proof; main does not catch it |
| satisfies number member changed to string | checker TS2322 | checker relation, before lowering |
| suppression probe's literal replaced by function-returned number | builds, native true vs Node 2 | stdout comparison; literal clang barrier is bypassed for both directives |
| survey's NonNullExpression counting branch removed in scratch | one-site control expects 1, mutant reports 0 | independent known-occurrence control |

The counter control copies non-null.a to a one-file .ts corpus, runs count.cjs, then removes only its NonNullExpression branch in a scratch copy and reruns it. /tmp/escape-counter-check.log records `control ... observed 1 PASS` and `mutant ... observed 0 FAIL`. It demonstrates that the count check can fail; it does not certify every AST visitor branch.

**Observation.** Regression commands, with output in files:

```sh
go test ./internal/lower ./internal/load -count=1 > /tmp/escape-packages.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout 30m ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^(casts|cast_fails|optional_strings|optional_numbers)[.]a$' > /tmp/escape-oracle.log 2>&1
```

Both commands exited 0. lower passed in 6.838s, load in 0.549s. All four selected oracle fixtures passed in 7.386s total; native cache misses 11, Node misses 8, zero hits. This oracle includes the checked-backend comparison for intentional cast failures and native sanitizers/leak checks. An earlier filter `TestNativeAgreesWithNode/.+(casts|cast_fails|optional_strings|optional_numbers)` selected no tests (0.011s); that run is not counted as verification.

**Observation.** No compiler package was edited, so the full gate, vet and formatting gate were not run. These docs repros are intentionally invalid and live outside the existing tsconfig include list, rather than becoming positive oracle fixtures or weakening repository settings. They are measured witnesses, with an explicit runner, not newly integrated gate assertions. No compiler-mutant edits were made. The known miscompiles were measured on unmodified main, and no lowering fix is claimed.

**Inference / not covered.** The decided non-null check, target initialization checks and body/initialization verifiers have not been implemented, benchmarked or independently mutant-tested. Reported proposed costs are operation counts, not measured timings. The census does not measure dynamic execution, inferred any, aliases of Object APIs, actual unsafe method assignments, all expando additions, or which of the 2,884 downcast candidates have reifiable tags. It uses TypeScript's checker as a classifier, not as a soundness oracle. Future implementation units must run isolated mutants for each accepted proof, every elision condition and every runtime failure, including null inputs, alias invalidation, false-branch guard narrowing and exception paths.

## Checked downcasts implemented

On `codex/checked-downcasts`, based on enum commit `7127080756a1904cc8d65ed766b66bbfaa767771`, object union casts accept single literal discriminants, including numeric and string enum members from ordinary and const enums. Targets may be one member or a sub-union. The proof requires unique runtime tag values, one tag representation, and sound member relations; duplicate enum aliases and payload refinements are refused. Proven upcasts and casts already established by flow erase. Structural hidden optional fields cannot become typed fields merely through a cast.

Nominal class downcasts use the same ancestry and erased generic identity as `instanceof`, through ordinary IR helpers shared by both backends. Generic target arguments must already be fixed by invariant source ancestry: every target parameter must appear directly in the source ancestor's type arguments. A nongeneric base cannot prove an arbitrary `Box<number>`, and incompatible generic views remain refused. Successful casts evaluate their operand once and preserve its identity. Failure flushes stdout, prints `adamic: panic: cast failed: this <source type> is not a <target type>` with a newline, and exits 70. The Node runtime now uses blocking stdio and terminal panic exit, so catch and finally do not execute.

Sixteen new oracle fixtures cover success and failure, both backends, Node source, native release, ASan/UBSan and leaks on successful runs; failing fixtures use the oracle's checked flag. Contract tests independently pin messages, 340,009 bytes of preceding stdout, and absence of catch/finally execution. Sixteen runtime mutants are killed by exit or stdout comparison: skip each of eleven failing checks, use a wrong numeric tag, string tag or class identity, and evaluate each operand form twice. Seven restored compiler/runtime source mutants are killed for duplicate tags, omitted nominal ancestry, omitted generic argument proof, hidden optional fields, unsafe writes through union views, catchable panic and unflushed panic. The duplicate-tag mutant emits valid C and finishes without sanitizer findings, printing native `1` against Node's `wrongwrong1`.

Primitive `number | string | boolean` union assertions remain refused. Their packed union tags are not independently validated for cast extraction and ownership in this unit, so the existing narrowing path remains the repair. Unknown-to-interface, callback signature changes, mutable widening, unrelated assertions and `as unknown as` remain refused with `adamic/no-unchecked-cast` and the admitted forms named. Nested or transformed generic argument recovery is conservatively refused. Validation here is Linux only; no performance claim or macOS execution is made. The accepted design above is copied unchanged from `origin/codex/escape-hatches` because the enum base did not contain this document.

## Cast appendix: base interfaces and construction invariants

Status: historical construction-or-refusal proposal, superseded by the October 7
checked-view ruling below. The corpus observations and ledgers remain evidence. Branch
`codex/interface-downcasts` starts at main `ef3d907`; checked-downcasts
`e816797f07a70952e214be079a1da54db80fbd7c` was read, not merged. Main did
not contain this document, so its existing text is preserved from that branch.

### Evidence and limits

Observation: the available original compiler corpus is
`cohere/TypeScript/tsc/testdata/fixtures/compiler`, through the cohere submodule
`715ba94f3608a6500086b1076ce5cb7e51b836db` and its TypeScript checkout
`8d550c837c90bd1805b047b7eeccc2baac2d5e7a`. This is the Go port's compiler
fixture corpus, not a newly cloned `src/compiler` v6.0.3 checkout. No source
from cohere is copied into Adamic. Locations below refer to that corpus.

Cast counts use the assertions ledger at
`origin/codex/stage3-fixtures-assertions`, commit
`5b173f3920ab2c5b7058f0a9fe4b8e4a91f52523`,
`stage3/fixtures/assertions/ledger-summary.json`. Its TypeScript 6.0.3 source is
`050880ce59e30b356b686bd3144efe24f875ebc8`, distinct from the local corpus above.
This unit does not repeat that classification. The ledger partitions 4,101 sites:
208 upcasts, 16 as const, 136 tagged union downcasts, 1,178 structural interface
downcasts without a tag, 10 as unknown as nodes, and 2,553 other sites. Of other,
1,842 are tagged narrowings outside a partitioned union. The original ledger
alone does not give an exact base-interface count or separately count classes. Its locations and field names supply the census, without claiming
that a TypeScript assignability result proves an Adamic construction invariant.
The old counts in this document describe their own historical corpus only.

The follow-up requested by the lead refines only the 2,553 other sites using the
same pinned stock compiler, source hashes, source offsets and checked types.
`stage3/interface-downcasts/refine-other.cjs`, `other-summary.json` and
`other-locations.tsv` retain the procedure, cross-tab and every location. Declared
ancestry means explicit extends chains, not merely structural assignability.

| Refinement of other | Sites |
| --- | ---: |
| Tagged declared base-interface downcast | 1,758 |
| Tagged interface-union narrowing without a valid partition | 57 |
| Tagged composite, generic or mapped narrowing | 25 |
| Tagged tuple narrowing | 2 |
| Other structural or union narrowing | 255 |
| Type-parameter narrowing | 3 |
| Nullish removal to interface | 6 |
| Non-assignable assertion | 343 |
| Any involved | 104 |
| Declared class downcast | 0 |

The 1,758 include 1,497 single-interface targets and 261 unions of declared
subinterfaces; 1,757 narrow kind and one narrows operator. Node-to-Identifier
alone occurs 27 times here. The 57 union cases still need a construction invariant
for open members: a tag cannot partition payload refinements the source does not
promise. Of the 25 tagged composite/generic/mapped cases, 24 narrow kind and one
isTypeOnly; they include staged Node-to-Mutable<Identifier> construction,
BindableObjectDefinePropertyCall intersections, and Node-to-T generic assertions.
The two tuple cases narrow length to zero. These require their own payload,
mutability or generic proof and are not additional directly declared subtype
pairs. The 255 remaining narrowings have source/target forms recorded individually;
interface targets there can arise from mixed unions or mapped/intersection views,
not a single declared base-interface pair. The 104 any cases erase information
needed to certify a subtype relationship. Zero class downcasts applies to this
other-bucket refinement, not a new classification of the entire census.

Observation from that branch's `logs/kind-soundness.json`: a genuine factory
PlusToken held as Node has its kind changed through Mutable<Node> to Identifier.
Stock TypeScript reports zero diagnostics; Node prints `true` then `undefined`
for the tag comparison and escapedText read. This directly disproves a permanent
kind-to-payload tie even when initial construction uses a genuine factory.

Observation: `types.ts:942` defines `Node.kind: SyntaxKind`.
`types.ts:1701` defines `Identifier.kind: SyntaxKind.Identifier` and requires
`escapedText: __String`. Identifier is not the only interface with that kind:
TransientIdentifier adds resolvedSymbol, and GeneratedIdentifier refines
emitNode. A kind cannot prove either refinement.

Observation: `factory/baseNodeFactory.ts:26` obtains replaceable constructors
from objectAllocator. Its identifier allocator returns Node, with kind already
Identifier. `utilities.ts:8531` initializes common fields, but not escapedText.
`factory/nodeFactory.ts:1304` then assigns escapedText in createBaseIdentifier.
The synthetic factory at the end of nodeFactory wraps the same base allocator.
The parser installs another BaseNodeFactory (`parser.ts:1460`) and calls the
shared node factory with it. This is staged construction, not a complete object
at allocation. `nodeFactory.ts:1209` returns a base allocation asserted as
Mutable<T>; the kind-to-T relation there is itself an unproven assertion.
`createNumericLiteral` at 1233 sets text and numericLiteralFlags afterwards.
`cloneNode` at 6368 allocates by kind and copies properties dynamically before
returning. `utilities.ts:8577` lets setObjectAllocator replace constructors.

Observation: even wrapper completion is not the full declared shape.
createBaseIdentifier sets required Declaration.symbol to `undefined!`, for the
checker to initialize later; base constructors set required Node.parent to
`undefined!`. Interface brand fields typed any are not assigned by these
constructors. Consequently the literal claim that every node factory builds
all required fields with their declared types is false. This unit has not
certified every wrapper, allocator replacement, parser path or clone path.

Inference: most ordinary visitors rely on a publication convention: a node's
syntax payload is complete before the visitor sees it. That convention is
plausible for escapedText, but is not proof of every inherited field. A compiler
must prove the convention with definite initialization and escape analysis,
including callbacks and exceptional paths, rather than trust factory names.
A universal invariant at every allocation is demonstrably too strong for tsc.
A weaker invariant at publication/use requires a separate builder proof.

### Shared construction certificate for casts and predicates

The companion body-shape ledger at `origin/codex/stage3-fixtures-predicates`,
commit `e42eaf9854563617734ff13987cc27e4e09f78b1`, records all 651 predicate
syntax nodes and 11 fixtures in `stage3/fixtures/predicates`. Its mutually exclusive
categories are 345 kind-equality bodies, 136 predicate-call bodies, 24 flags bodies,
8 typeof bodies, 16 asserts bodies, 74 other bodies and 48 bodyless contracts.
These are body descriptions, not proof verdicts. Its 227 direct equalities in
nodeTests and the larger kind-equality group expose the same open-Node obligation
as these casts. A predicate's declared return type is not a construction premise.

Both consumers should request the same certificate: for every object observable
through the source view, the tested tag implies the full requested interface, and
publication and subsequent writes preserve this implication. A verified positive
predicate branch can use that certificate after its body establishes the tag.
A negative branch additionally needs equivalence between the body test and target
membership; a positive implication alone cannot justify excluding the target.
Composition must verify each called predicate and every required nested read.
For example, a void-zero test needs a valid VoidExpression.expression before
reading its operand's kind, and a numeric-literal certificate before reading text.
A signed-numeric test additionally needs the prefix-unary operator and operand
shape. A flags test needs a separate flag-to-payload invariant: a Transient bit
does not alone prove TransientSymbol.links. These obligations must not be replaced
by a whitelist of predicate names or circular trust in asserted return types.

The certificate is necessary, not sufficient, for guard verification. An empty
assertion body proves nothing; a debug assertion that can be disabled proves
nothing on that path. A bodyless predicate needs an independently justified host
contract. Generic assertions and nested refinements remain separate obligations.
This prototype verifies construction for casts only; it does not admit the
ledger's predicate declarations or certify all 651 bodies.

### Proposed language choice: (b), prove construction or refuse

Permit a base-interface-to-sub-interface assertion only when the compiler proves
that every possible object reaching that base view with the requested runtime
discriminant satisfies the entire target, and every later write preserves that
fact. Then emit the existing discriminant check. If that proof is unavailable,
refuse with the unproven construction or write site. The assertion and the target
annotation are never premises of their own proof. This is a closed-program fact,
not a claim attached to the spelling Node, Identifier or SyntaxKind.

The proof includes literals, spreads, classes, constructors, allocator overrides,
clones, module imports, host boundaries and alias writes. Matching uses runtime
literal values, so enum aliases cannot establish uniqueness. Two interfaces with
one kind are allowed only if the relevant constructions prove the requested
shape; a stricter payload refinement needs its own proof. An incomplete object
may exist privately inside a verified builder only when no read, publication,
callback or exceptional escape can expose it before completion. Later writes
must preserve field types and presence, through all structural views. Mutable
containers and callable fields need the existing invariant/variance proofs;
checking field presence alone cannot establish them.

Refused programs include a Node literal with Identifier's kind and no name field;
a matching kind with a number in a string payload; a broad/dynamic kind whose
possible Identifier case lacks its payload; an opaque constructor or host source;
a clone whose copied fields are unproven; a payload refinement unsupported by
construction; and a later kind/payload write that can invalidate the implication.
A program may repair this by constructing the full target and preserving its
fields, or by using an actual discriminated union. Adding an unrelated malformed
allocation can invalidate the conservative global proof even if unreachable.
This is intentional in the first prototype, not required of a future reachability
proof. Unknown-to-interface and unrelated assertions remain refused.

Runtime cost after proof: one existing tag read, one comparison, and one failure
branch for a single numeric/boolean tag; a string tag additionally uses the
existing string equality operation (pointer fast path, otherwise byte comparison
linear in tag length). Current native field access can require a layout-cache
miss scan over S field names, O(S plus compared name bytes); a cache hit loads
the cached slot. There is no new allocation, copy, retain or release for proof
itself. Operand evaluation and ordinary result ownership still cost what they
normally do. No whole-program check runs at runtime. The prototype scans the
program per cast, so compile-time work is O(C times N plus C times A times F)
checker queries for C casts, N AST nodes, A literals and F target fields; checker
relation cost is additional. A production implementation should cache summaries.
No runtime timings are claimed.

### Why not choose (a) or (c) yet

A correct layout check is a useful future fallback. The current adamic_shape
contains names and reference flags, not complete field types: number and boolean
layouts can share those flags, as can string and object layouts. Layout membership
alone does not prove initialization, literal refinements, nested structural types,
callable signatures or safe writes through aliases. It cannot silently be treated
as a validator for arbitrary T.

With richer immutable typed-layout certificates, a fallback could compare a
layout ID with a precomputed compatible-layout set, then check value refinements
and initialization. A bitset costs a layout ID load, indexed word load, bit test
and branch, with ceil(L/word-bits) words per target for L layouts. Without such a
certificate, an F-field name/type scan over S slots costs O(F times S plus name
bytes), plus literal tests; recursive contents require traversal proportional
to reachable values, cycle handling and alias guarantees. Neither is the advertised
single discriminant comparison. Typed certificates and publication-state tracking
need their own representation design and mutants. Option (c) is not a license to
fall back to unsafe presence tests when the proof fails.

### Prototype boundary

The proposed flag is `ADAMIC_INTERFACE_DOWNCASTS=1`, default off. It admits
complete object literals, readonly required scalar target fields, and a single
literal discriminant. Factories return complete shapes; visitors hold them through
the Node base. It checks every imported module's allocations, including unused
functions, and rejects staged builders, spreads, new/opaque construction,
optional/nested/callable fields, generic construction and writes to the checked
fields. String/number/boolean literal tags are covered; enum syntax support stays
on the separately reviewed checked-downcasts branch. No protected compiler file
needs editing. Both backends reuse CheckedCast and the existing failure message,
exit 70, single operand evaluation and identity preservation.

Required evidence: positive factories/visitor compared to original source on Node,
both generated backends and native sanitizers/leaks; wrong-kind runtime failure;
missing-payload and wrong-payload refusal; imported malformed construction; writes
through aliases; and a compiler mutant dropping only the construction obligation.
The missing-payload mutant must get past the discriminant test and be killed by
the construction assertion in the test harness, not clang or a sanitizer. The
prototype does not claim to compile unchanged tsc or verify its staged factories.


## Cast ruling: checked views, decided October 7, 2026

Accepted by @system_adamic: choice (c). The construction proof is an eraser,
not an admission gate. The earlier choice (b) and its default-off prototype are
historical; the implementation steps below replace them. This records the ruling,
not a claim that all implementation steps have finished.

First land tagged interface downcasts, the 1,758 cases identified above. Check
the requested discriminant at the cast, evaluating the operand once and preserving
identity. Every subsequent field read through the narrowed view that lacks a
valid proof checks presence, initialized state and the declared runtime type tag
before loading the payload in its target representation. Initialization alone
never proves type, and a matching kind never proves another field's initialization.
Literal and enum refinements need value checks in addition to a primitive tag.
A safe read must also convert between stored and target representations where
necessary, rather than reinterpret a boxed union as a raw scalar.

Failing reads flush preceding stdout, print one newline-terminated message to
stderr and terminate with exit 70, without running catch or finally:

```
adamic: panic: field read failed: <expression text>.<field> is not initialized; expected <type>, found missing|uninitialized
adamic: panic: field read failed: <expression text>.<field> is not a <type>; expected <type>, found <runtime category>
```

Capture the receiver expression text at compile time. Evaluate the receiver and
load the field once; diagnostics must not invoke user conversion code. The existing
cast-failure contract applies when the tag test fails at the cast.

Staged construction is legal. A Node may carry Identifier's kind before escapedText
is filled; createBaseIdentifier may fill it later; undefined! and null! initializers
record uninitialized state; cloneNode may copy staged fields. Presence, current
initialization and actual type metadata must survive copies, assignments, aliases,
reuse and region allocation. The non-null worker on `codex/non-null-check` owns
field initialization state and will push that representation separately. This unit
merges and uses that helper instead of adding a second bitmap or state convention.
Removing the old admission gate before all three read checks exist is unsafe.
`ADAMIC_INTERFACE_DOWNCASTS` has been removed. The supported scalar subset is default; broader contracts remain NotYet during implementation.

A compatible dominating store or a construction certificate valid at this read
can erase the checks. A failed proof leaves a check, rather than refusing the
program. Neither an asserted type nor the ! syntax is a proof. A tag test proves
only the tested tag unless an independent construction invariant proves more.
Calls, alias writes, uncertain exceptional paths and dynamic copies invalidate
facts unless their effects are independently proved. An unused malformed factory
must no longer make an otherwise checked program a compile-time refusal.

Second land the 1,178 untagged interface downcasts. They create checked views
without testing a discriminant at the cast. A field read uses the same three
checks. An object-valued result is another checked view, including through locals,
parameters, returns, aliases and containers, until an independent proof or a valid
tagged narrowing establishes the needed facts. Erasing the view marker at one
of those boundaries is an unsound implementation, not an optimization.

Refuse an unproven operation whose target contract cannot be certified at runtime,
such as a newly asserted callable member's parameter/result signature. A typeof
function check cannot prove that signature. Retain the existing refusals of any,
unrelated double assertions and unsafe function variance. Missing payloads, staged
fields and a wrong payload type are no longer admission refusals: their unproven
reads terminate according to the field-read contract. Unsupported runtime contracts
must be reported by target/member and counted against the original ledger; existing
backend capability gaps must be distinguished from those language refusals.

Both backends emit the same checks. A passing checked program matches the original
source on Node byte for byte. Node's source execution is unchecked, so a failing
check is validated by an independent assertion of its complete failure contract.
Required mutants drop a field check, drop initialization tracking, erase without
proof, skip a transitive view, and evaluate an operand twice. Each must emit valid
code and be killed by a semantic assertion, not clang or a sanitizer.

Third measure actual remaining checks after erasure on named tsc slices and release
runtime against an explicitly named unchecked benchmark control. Report read-site
and dynamic-read denominators separately, compiler/release flags, repetitions,
outputs and timing spread. Source occurrences alone are not execution heat. The
unchecked control is an experimental artifact, not a production option. If checks
are costly, improve proofs; never reduce the required checks. No benchmark or
checked-read share is claimed by this decision entry.

Tagged (c) lands first. The updated delivery targets are tagged admission by 16:00 UTC October 7 and
untagged checked views by 02:00 UTC October 8. Both were reported at risk before
this partial checkpoint because complete tsc field contracts remain unsupported. If that step cannot finish, report it early with the highest-exposure
untagged locations, clearly distinguishing static counts from measured hot paths.

### Implementation progress, October 7

The flag and construction admission gate have been removed from production.
Tagged casts with required number, boolean or string fields now use the tag check
and checked reads by default. Finite literal and enum contracts additionally compare
values. Missing or malformed scalar payloads lower and fail at their first read;
unused malformed factories no longer prevent admission. Staged fields, destructuring,
boxed scalar reads and supported boxed stores have source fixtures on both backends.
The non-null tip `e2ea9ab2`, including `c680ecf4`, supplies the shared readiness helper
and state; no second initialization representation was added.

The reusable lowering entry point is `(*lowering).view(node, value, target)` in
`internal/lower/interface_cast.go`. It returns the original operand, preserving
identity, and registers fields for conservative program-wide checks across aliases.
This checkpoint is incomplete: optional/nullish, object-valued, mixed scalar and
recursive contracts remain NotYet, as do unsupported accessor/conversion aliases.
Untagged admission, transitive object views and proof erasure are not implemented.
The machinery is not ready for the optional-property widening worker to consume.
Direct callable contracts and callable fields are Refused because a runtime function
tag cannot certify a parameter/result contract. Backend gaps are NotYet, not new
permanent language refusals. Generated helper reads and opaque runtime object
producers still need a complete propagation/metadata audit before integration.

The pinned assertions ledger target audit yields zero eligible complete contracts:
19 of 1,758 tagged targets have callable members, and 1,739 need broader field
contracts; 71 of 1,178 untagged targets have callable members, 1,106 need broader
contracts, and one otherwise scalar target still needs untagged admission. Exact
locations, fields and types are in `stage3/interface-downcasts/default-contracts.json`.
This is an outside-checker rejection upper bound, not an observation of compiling
all tsc source files. It establishes 0/1,758 and 0/1,178 supported complete contracts;
positive lowering coverage would require actual lowerer probes. No checked-read
share or release benchmark is claimed at this checkpoint.

The additional native representation storage is exactly one byte per physical field,
following the worker's existing one-byte initialization state. Heap storage is
`sizeof(adamic_object) + field_count * (sizeof(adamic_value) + 2)` before allocator
rounding; region storage rounds that size up to a multiple of 16. A checked required
read makes one runtime helper call, performs cached own-name lookup, reads initialization
and representation bytes, and validates a referenced value's heap kind before using
it. A cache miss compares field names linearly; a hit does not scan. A successful
native scalar read check allocates nothing. JavaScript finite-literal checks currently
construct the allowed-values array at each read; object construction also records
representation metadata in a WeakMap. Reference reads currently retain and later release
the value. Failure formatting allocates a message before the exit-70 panic. JavaScript
checks own presence, the shared readiness WeakMap, and the value's built-in runtime
kind; diagnostic type names are allocated once for the program. These are primitive
costs, not the requested post-erasure counts or release timings for tsc slices.

The exact 69 cumulative latent refusals from `70456b7` are audited separately in
`stage3/interface-downcasts/latent-targets.json`: 44 object targets without direct
callable members, 18 unions, 3 objects with direct callable members, 2 type parameters,
1 array and 1 scalar/other target. Inherited array/string methods are separated from
callable object contracts. This is the baseline on a checker-rejected program, not
the final list of refused shapes after checked views. The latter remains pending.

## Non-null assertion implementation

`e!` now lowers through the existing nullish coalescing panic path. It evaluates
its operand once, checks the stored maybe pair or null reference (including mixed
unions), and returns the non-null representation. The diagnostic captures the
assertion's original source text, including `!`. Panic flushes stdout, prints
`adamic: panic: non-null assertion failed: <expression text> is null or undefined`
to stderr and exits 70 without running catch or finally.

Numbers and booleans in maybe pairs, optional references, mixed unions containing
undefined, and nullable match results are covered. Present zero, false, empty
string and NaN pass. Assertions on already present types emit no additional check;
reads narrowed from nullable storage keep a check, including after calls or
capture writes. No new flow elision or reference representation was introduced.

The nine non_null oracle fixtures pass uncached, with successes held to source
Node and failures held to the checked JavaScript backend. Mutants removing the
native check, rejecting zero, false or empty string, and evaluating twice fail
those comparisons. A changed message fails the exact-text lowering test; ignoring
nullable storage after a capture write fails the oracle under UBSan.

### Shared field readiness representation

Each native object carries one initialized byte per field after its `adamic_value` slots, indexed by its actual shape. `adamic_object_initialized(const adamic_object *object)` exposes those bytes; `adamic_object_set_initialized(adamic_object *object, const char *name, bool initialized)` updates a named slot. Fresh ordinary fields are initialized; an `ir.Field.Uninitialized` starts clear. Writes set the bit. The bytes share the object's allocation, including region allocations.

Checked reads use `adamic_value *adamic_object_read(const adamic_object *object, const char *name, adamic_slot_cache *cache, const char *expression)` in native code and `adamicReadField(object, name, expression, optional = false)` in JavaScript. `ir.Property.Readiness` supplies the source expression. Missing or uninitialized fields panic with `read before assignment: field '<name>' in <expression>`. The state is independent of the value, so zero, false, empty strings, and assigned undefined do not mean uninitialized. JavaScript keeps state in a WeakMap, preserving own keys.

### Uninitialized assertions and definite assignment

`undefined!` and `null!` in a let, const, class field or parameter-default initializer now reserve an uninitialized slot. `let x!: T` and `field!: T` use the same state. An assignment marks the slot ready independently of its value. Reads before assignment use the existing temporal-dead-zone readiness path and panic with exit 70, naming the variable or field and the source expression. Panic runs no catch or finally. Outside initializer positions, literal assertions still perform the loud nullish check. A shadowed `undefined` is an ordinary operand.

Captured locals keep readiness in their existing cell; iteration clones copy that state. The existing control-flow graph proves dominating assignments across joins, loops and exception edges, including reads in closures after their own writes. An assignment that throws does not initialize its destination. Field facts are per binding and invalidated by calls and binding writes. Cross-function assignment proofs remain conservative. Field reads and spreads share `adamic_object_read` and `adamicReadField`; the JavaScript helper also accepts `allowAbsent = false` after `optional`. Optional absence retains its ordinary undefined behavior. Static inherited reads check their actual owning slot.

Weak assertions use the same expression-bearing diagnostic, including after explicit clear or native lifetime release. Source Node agrees byte for byte on the successful fixtures. Inserted read checks and explicit Weak clearing use the oracle's checked JavaScript reference; native lifetime release has a separate pinned runtime assertion because Node retains the target.

Readiness mutants dropping a check, erasing across a zero-iteration loop, initializing to zero, missing captured and exceptional reads, and treating marker initializers as ordinary nullish checks are caught by runtime output assertions. Weak generic-message and native lifetime-message mutants are caught by exact stderr assertions. The latent rerun in `non-null-readiness-census.json` verifies all 78 recorded source hashes: the original 180 non-null Refused and 25 NonNullExpression NotYet findings are now zero for both reasons. Twelve original locations still encounter other recorded refusals or NotYet reasons. This checker-rejected, per-unit measurement does not establish that tsc compiles.

### Lazy computed assertion initializers

The later scanner ruling extends the initializer rule to `let x = e!`, `const x = e!`, `var x = e!`, fields and defaults. The operand is evaluated once in its stored representation. A present value initializes the slot; a nullish value leaves the shared readiness state clear. The eventual read diagnostic names the slot and the original initializer expression. Assignment still makes it ready, and the same dominance proof erases subsequent reads' checks. Arguments, returns, member receivers and other operands keep eager assertions.

The scanner fixture uses `var text = textInitial!` and a captured `setText`, with both missing and provided initial text. Local, instance-field, static-field and default fixtures cover assignment before reading and checked failure before assignment. Zero, false and empty strings remain present. The eager-initializer mutant exits 70 before the scanner can assign, while source Node prints its result. Other var forms remain refused; repeated var assertion declarations are refused explicitly. Function-local var reads before their declaration remain outside supported hoisting, rather than being guessed.
## Predicate bodies proven in stage 0

**Built, October 7, 2026.** Written `x is T`, `asserts x is T` and boolean
`asserts cond` signatures now require proof from their body. The verifier asks the
pinned checker's flow analysis about each return path, using only direct parameter
checks: `typeof`, literal `===`/`!==`, discriminant comparisons and nominal class
`instanceof`, composed with `!`, `&&` and `||`. It accepts blocks, branches,
explicit returns and assertion fallthrough; throws and runtime panics do not return.
A guard's true narrowings must match its declared type, and every false return must
exclude it. Every normal assertion return must have the checker's exact target type.

A rebound parameter is refused, including a write in a captured closure. Calls,
property reads that may invoke getters, and writes discard previous facts, including
discriminant facts reached through aliases.
An unproven return names its location and obligation and recommends inlining the
check or returning a discriminant comparison. Admitted guards lower as ordinary
boolean functions; assertions lower as ordinary void functions. Call sites use the
checker's narrowing. No runtime validator or extra check is inserted.

**Observation.** `proven_guards.a`, `proven_class_guards.a` and
`proven_assertions.a` agree with source Node and both compiled backends, including
false branches, narrowing `filter` callbacks, thrown failures, native sanitizers and
leak checks. Refusal probes include unconditional true, a different variable, stale
field checks, broader and narrower body types, false-branch lies and empty assertions.
Each compiler mutant below was caught by a refusal probe becoming accepted, before
native compilation; none was counted as killed by a build error.

| Mutant | Refusal probe that caught it |
|---|---|
| Skip nominal matching | Structurally identical classes with different nominal identity |
| Accept any body | Unconditional true and empty assertions |
| Skip the true narrowing comparisons | Wider narrowing: number claimed as literal 1 |
| Skip the aggregate exact match | Only literal 1 returns true, but number is declared |
| Skip the false-return proof | A number returns false from an is-number guard |
| Preserve facts across calls and writes | Stale discriminant calls/getter reads and assertion call invalidation |
| Permit parameter rebinding | Direct assignment and a captured closure write |
| Use assignability for assertion matching | A wider optional shape claimed after checking only the kind |
| Skip assertion postconditions | Empty, early, wider-shape and false-condition assertion returns |

**Limits.** `in` and enums remain refused on the main revision used for this unit.
Loops, switches, try/finally, opaque or recursive helper summaries, computed property
checks, default/rest parameters and `this` predicates stay refused by this verifier.
It caps branching at 256 paths. The separate cohere predicate lint rule is unchanged.
The historical survey and other hatch decisions above are not remeasured here.
## Satisfies and proven upcasts: implementation

**Observation.** Built on `codex/proven-relations`, from main
`ef3d907ecdc4c771b016f7d9c52372def057a340`, following the ruling above at
`0f30b52061374de90ceeac8d19b221cb51ae84f9`. Main did not yet contain this document,
so the preceding survey is preserved from that ruling commit. Its linked survey
artifacts remain on `codex/escape-hatches` until that unit is merged.

`expr satisfies T` now proves the relation and lowers to the operand. The checker
keeps the operand's own type, including contextual literal types and required
fields; lowering does not replace it with T. A candidate `expr as T` upcast must
pass the same proof before erasure. The existing checked-downcast lowering is
unchanged.

The common proof in `internal/lower/proven_relations.go` combines ordinary checker
assignability with the existing assignment walks for mutable invariance, readonly
covariance, contravariant function parameters, covariant results and nominal class
ancestry. It reuses the writable-slot diagnostics, which name the source and target
at the failing slot. A failed nominal proof names the class lacking ancestry.
Fresh literals retain the assignment rules: their contained values are checked at
their own sites.

Optional-field compatibility also has to be proven. A structural type omitting a
field cannot certify that the runtime field is absent or compatible. Such a view
is refused with the field path and `adamic/no-optional-widening`. Object literals
without spreads can prove absence, including nested literals and fresh array
elements. Existing mutable slots require compatibility in both directions, and
class type arguments remain invariant. For example, `Counted[] as Named[]` is
refused when Counted has an optional numeric count that Named hides: a write
through Named[] could replace an element with an object hiding a string count.
Readonly views may safely forget that field.

**Observation.** The new `.a` oracle fixtures are
`internal/oracle/testdata/proven_satisfies.a` and `proven_upcasts.a`. They cover
object and array literals, retained literal and required-field types, discriminated
unions, readonly views, interfaces, base-class dispatch, readonly arrays, compatible
function variance, optional literal absence, identity and single operand evaluation.
Both match source Node, generated JavaScript, sanitized native and release native;
finished programs pass the leak check. Their recorded allocations/frees are 28/28
and 15/15. No previous count row changed.

`TestProvenRelationsRefuse` has fifteen independently checked refusal probes.
`TestProvenRelationsErase` compares generated C and JavaScript byte for byte with
an operand-only control. Both operators add no generated work in that control.

**Observation.** Every mutant below was applied to the compiler, run, and restored.
The first eight runs exit 1 because their refusal test instead observes successful
lowering, `got <nil>`. None depends on clang rejecting a mutant. The final mutant
fails the artifact comparison for both backends.

| Compiler mutant | What catches it |
|---|---|
| Accept every satisfies without proof | satisfies mutable-array refusal |
| Erase an unproven upcast without proof | upcast nominal-identity refusal |
| Skip the writable-slot proof | mutable Dog[] to Animal[] and widened function-parameter refusals |
| Skip nominal ancestry | object literal satisfying a class refusal |
| Skip optional-field compatibility | hidden, nested, mutable and invariant-class optional-field refusals |
| Ignore reverse optional compatibility for mutable array elements | Counted[] to Named[] refusal |
| Ignore reverse optional compatibility for mutable object fields | mutable pet field refusal |
| Ignore optional compatibility in invariant class type arguments | Box<Counted> to Box<Named> refusal |
| Wrap satisfies in an unnecessary Coalesce | C and JavaScript erasure comparisons |

Mutant logs are `/tmp/proven-mutant-<name>.log`, with names `accept-satisfies`,
`erase-unproven-upcast`, `skip-invariance`, `skip-nominal`, `skip-optional`,
`skip-optional-array-write`, `skip-optional-field-write`,
`skip-optional-class-invariance`, and `added-work`. The refusal mutants use early
returns in provenRelation or disable the individually named guard. The added-work
mutant wraps the lowered operand in `ir.Coalesce` with itself as its fallback.

**Observation.** Setup selected `/workspace/adamic-tools/env.sh`. The first run
failed cache warming while the cast call site was edited before its helper was
written: `vet: internal/lower/cast.go:30:15: l.provenRelation undefined`.
The completed-source retry exited 0 and printed:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (119s)
setup: done in 119s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

`nproc` printed 5. Tool versions were Go 1.27.1, clang 20.1.8 and Node v24.19.0.

**Limits.** This unit does not extend runtime-checkable downcasts, non-null checks,
predicate verification or the optional-field proofs for ordinary assignment sites.
An open structural source with an omitted optional field is conservatively refused;
this unit does not recover its exact origin through arbitrary aliases or callbacks.
Existing stage 0 representation limits remain in force. No performance benchmark
is claimed; the zero added work observation comes from emitted-artifact equality.

**Observation.** Final scoped verification, with every test's output captured in a
log file, exited 0:

```sh
source /workspace/adamic-tools/env.sh
gofmt -l cmd internal > /tmp/proven-format-final.log
go vet ./... > /tmp/proven-vet-final.log 2>&1
go test ./internal/lower ./internal/load -count=1 > /tmp/proven-packages-final.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(proven_|casts[.]a|cast_fails[.]a)' -count=1 -timeout 30m -v > /tmp/proven-oracle-complete.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m > /tmp/proven-counts-check.log 2>&1
```

Formatting and vet produced empty logs. Lowering passed in 32.267s and loading in
1.154s. The four-fixture uncached oracle passed in 1.770s, including the existing
successful and failing checked-cast regressions, with zero cache hits. The final
count check passed in 60.490s. These are wall times under concurrent test load,
not compiler performance measurements.

The complete worker gate was also attempted:

```sh
go test -count=1 -timeout 30m ./... > /tmp/proven-full.log 2>&1
```

It was stopped after more than nine minutes while broader bridge/native/stage 1
work was still running, using the unit's permitted scoped fallback. Its Go process
exited 143 on termination, and all ten remaining child processes were stopped.
The full gate is not claimed to have passed. The completed scoped commands above
are the regression evidence for this unit.
