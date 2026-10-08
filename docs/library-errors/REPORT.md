# Library Error objects, step 21

Branch: `codex/library-errors`. Base: fetched `origin/area/library`
`71f916872667150037677b4b975772f165e47898`. No AGENTS.md exists in the
workspace or base. Node oracle: **24.19.0**. Linux execution uses Go 1.27.1,
clang 20.1.8 and wasm32-wasi. Raw build/test output stays under
`/tmp/library-errors-*`, outside this checkout.

## Implemented surface and explicit limits

Error, TypeError, RangeError, SyntaxError, ReferenceError, EvalError, URIError
and AggregateError construct native counted objects under their Node names.
Both `new Error(...)` and direct constructor calls are admitted. Omitted and
literal undefined messages have no own message; explicit strings, including
empty strings, do. Default name/message reads inherit from the corresponding
builtin prototype. String assignments to name/message establish own properties
with Node's enumerability and insertion order. toString implements V8's
name/message algorithm, including both empty-string cases. Native identity,
Error/Object ancestry, builtin prototype links, isPrototypeOf traversal and
literal constructor ownership probes are represented directly.

A plain options literal can install cause, including shorthand, undefined,
null, primitives and represented object identities. Cause stays non-enumerable.
AggregateError snapshots represented arrays into its own non-enumerable errors
array, boxes scalar elements and preserves reference identity; numeric element
reads preserve undefined outside the array. Mutation of the input does not
change the snapshot. Strong children participate in counting, region disposal,
sharing and the compiler's cycle analysis.

These are supported members with represented inputs, not a claim of the whole
ECMAScript object model. Unsupported forms are rejected during checking/lowering:

| Refused form | Reason / prerequisite |
| --- | --- |
| Detached constructors or methods; `.call`/`.apply`; prototype constructor values and generic Function ancestry | Compiler callable identity, receiver binding and generic function prototype representation |
| A message that may dynamically be undefined | Compiler conditional own-message presence lowering; literal undefined is supported |
| Object/symbol message coercion; non-string name/message writes | Compiler ToPrimitive/ToString dispatch and catchable coercion failures |
| Nonliteral options, spreads, getters, inherited options cause | Compiler general HasProperty/Get and evaluation-order proof; no cause stub |
| Non-array AggregateError iterables or unrepresented elements | Compiler iterable protocols / heterogeneous boxed representations |
| cause/errors writes, Error bracket members other than the represented stack read, bracket writes, Error destructuring/spread/Object.assign targets | Compiler property operations must use Error metadata rather than fixed plain slots |
| Own reflection or stack reads on an unproven/host Error origin | Compiler object-origin proof and library host Error descriptor/stack metadata; legacy host shapes are not advertised as constructor shapes |
| Error prototype/unproven-origin mutation in a program importing host modules or using process/Buffer | Library host Error ancestry metadata; no inherited-name approximation |
| Optional Error access, extra constructor arguments, getPrototypeOf on a potentially null ancestor | Compiler short-circuit/evaluation-order/catchable TypeError lowering |
| Frozen/sealed Error mutations or capture targets; generic property descriptors/prototypes | Compiler descriptor failure participation in exception cleanup and generic descriptor representation |
| Structural Error views with a compatible non-Error runtime shape | Compiler object-origin proof; plain objects are never reinterpreted as Error internal slots |

Prototype mutations are counted even when a program leaves a dynamically built
prototype string live. Cleanup is registered at runtime start, after the heap's
exit hook and before workers; it runs after workers join, including programs
that print nothing. Prototype strings are shared before publication; parallel
reads pass the sequential/pool sanitizer and race variants. Runtime storage is audited in `docs/runtime-statics.md`.
V8 13.6.233.17's Error construction, InstallErrorCause, toString and AggregateError
array construction algorithms are credited in `THIRD_PARTY_NOTICES.md`.

## Stack ruling

The first line matches Node exactly, including empty name/message cases. An own
stack string is installed on construction and capture; its header is formatted
lazily and cached, matching the tested name/message mutation behavior. The
represented stackTraceLimit is a readable/writable number. Limit zero produces
only the header. Positive limits currently have one available Adamic creation
or capture frame, with the emitted native function name and source location
where known, on native and wasm32-wasi. The format is `Adamic frame: ...`, never
a fabricated V8 `at` line. Text after the first line carries no parsing contract.
The JavaScript backend uses its actual emitted program's V8 frames.

**The oracle remains full stdout, stderr and exit-code comparison.** The
`library_error_stack_frames.a` header names `stack frames` as a known difference.
Its complete stdout disagrees on every backend, and that disagreement is logged
with both full outputs and exits. The first-line check is an additional contract
check, not a shortened oracle. Unmarked fixtures receive no exemption.
`library_error_stack_header.a` prints complete header-only stacks at limit zero;
its entire output agrees byte for byte with Node.

## Scoped test262 counts, before / after

Pinned test262: `2e0a56762801e275a9fdf96dc49d90ba0cddcf63`.
Stock TypeScript parser/checker: 6.0.3. These are the existing runner's documented
adaptation/harness classifications. A skip or refusal is never a pass.

| Suite | Pass before / after | Refused before / after | Not TypeScript | Skipped | Total |
| --- | --- | --- | --- | --- | --- |
| Error | 0 / 11 | 19 / 8 | 12 | 62 | 93 |
| NativeErrors | 0 / 6 | 18 / 12 | 19 | 57 | 94 |
| AggregateError | 0 / 1 | 2 / 1 | 4 | 19 | 25 |
| Total | 0 / 18 | 39 / 21 | 35 | 138 | 212 |

Zero disagreements and zero crashes. Every one of the **18 passing adapted
sources** was additionally run on native, JavaScript and WASI against Node's
complete stdout/stderr/exit, through TestErrorTest262Backends. The adaptation's
assertThrows prelude checks Error membership rather than exact native constructor
identity; independent family fixtures verify native constructor identity.

Remaining attempted refusal groups:

| Suite | Reason | Count |
| --- | --- | --- |
| Error | Detached toString method value | 5 |
| Error | Constructor alias / enumeration of constructor value | 1 |
| Error | Inherited prototype constructor value | 1 |
| Error | Generic Function.prototype isPrototypeOf | 1 |
| NativeErrors | Detached toString method value | 12 |
| AggregateError | Detached hasOwnProperty method value | 1 |

The checker refusals remain grouped by TS error code in the scoped runner
report. Skips include propertyHelper/nativeErrors/isConstructor harness files,
getters, Symbols, proxies, cross-realm and other unsupported language/harness
features. Descriptor/harness coverage is therefore not implied by these counts.

## Fixtures, mutants and Linux counts

Seven new family fixtures join the four existing capture/limit fixtures:
constructors, cause, identity/prototypes, format, aggregate, stack_header and
stack_frames. The ten ordinary fixtures agree on all backends; the eleventh
records the ruled named frame difference. Both sanitized allocation modes,
release builds and WASI execute the native fixtures; leak checks remain enabled.
Linux `internal/oracle/counts.md` has **650 -> 657 fixture rows**. New Error rows
and affected existing constructor rows are refreshed by TestErrorCountsAreRecorded;
unrelated measurements remain present in their original fixture order.

Eleven mutants run cleanly (exit zero, empty stderr, no sanitizer findings) and
are caught only by comparison with Node stdout, on all three backends:

- TypeError construction becomes Error.
- Cause is omitted.
- instanceof always answers false.
- isPrototypeOf always answers false.
- Constructor ownership probe always answers false.
- TypeError.prototype skips Error.prototype.
- toString returns only the message.
- AggregateError discards its elements.
- Stack header ignores the constructor message.
- captureStackTrace has no body (existing mutant).
- stackTraceLimit assignment is ignored (existing mutant).

Changed library/support packages are checked; native borrow/static-storage and
signal/exit regression checks pass after fixing the issues found by the package
run. The fast gate remains responsible for the rest of the repository.

## tsc source inventory

The inventory parses all 81 vendored compiler fixture files with TypeScript's
AST, rather than matching comments. These files are pinned by the cohere
TypeScript submodule at `d92d9bfee114c80be2c375d72edae966176e3a4f`.
It reports **10 constructions, 26 catch clauses, one capture call**. No direct
TypeError, RangeError, SyntaxError, ReferenceError, EvalError, URIError or
AggregateError construction occurs in this source tree. Host-generated errors
caught by tsc are not counted as source constructions.

| Construction site | Exact expression |
| --- | --- |
| src/compiler/core.ts:1583 | `new Error("Queue is empty")` |
| src/compiler/core.ts:1887 | `new Error("Not implemented")` |
| src/compiler/debug.ts:199 | `new Error(message ? &#96;Debug Failure. ${message}&#96; : "Debug Failure.")` |
| src/compiler/debug.ts:1093 | `new Error()` |
| src/compiler/moduleNameResolver.ts:1681 | `new Error(&#96;Could not resolve JS module '${moduleName}' starting at '${initialDir}'. Looked in: ${failedLookupLocations?.join(", ")}&#96;)` |
| src/compiler/program.ts:1526 | `new Error(&#96;${option.name} is a string value; tsconfig JSON must be parsed with parseJsonSourceFileConfigFileContent or getParsedCommandLineOfConfigFile before passing to createProgram&#96;)` |
| src/compiler/tracing.ts:66 | `new Error(&#96;tracing requires having fs\n(original error: ${e.message &#124;&#124; e})&#96;)` |
| src/compiler/utilitiesPublic.ts:448 | `new Error("start < 0")` |
| src/compiler/utilitiesPublic.ts:451 | `new Error("length < 0")` |
| src/compiler/utilitiesPublic.ts:471 | `new Error("newLength < 0")` |

| Catch file | Lines |
| --- | --- |
| src/compiler/commandLineParser.ts | 2300 |
| src/compiler/moduleSpecifiers.ts | 156 |
| src/compiler/performanceCore.ts | 43 |
| src/compiler/program.ts | 404, 439, 2847 |
| src/compiler/sourcemap.ts | 429 |
| src/compiler/sys.ts | 1276, 1478, 1552, 1599, 1621, 1634, 1704, 1792, 1877, 1921, 1934, 1943 |
| src/compiler/tracing.ts | 65, 240 |
| src/compiler/utilities.ts | 6742, 7813 |
| src/compiler/utilitiesPublic.ts | 741, 751 |
| src/compiler/watchUtilities.ts | 202 |

Capture: `src/compiler/debug.ts:201`,
`(Error as any).captureStackTrace(e, stackCrawlMark || fail)`.

The constructor fixture uses the exact expressions at core.ts:1583/1887,
debug.ts:1093, utilitiesPublic.ts:448/451/471 and program.ts:1526. Other native
families and AggregateError have no direct tsc construction witness and use
ECMA/test262 witnesses. Stack capture fixtures follow debug.ts:201's API shape;
its original `stackCrawlMark || fail` still needs compiler truthiness lowering.

Verified compiler blockers are kept intact rather than rewritten:

- debug.ts:199: the original optional-string truthy conditional is refused;
  the library fixture's chosen Debug Failure message is not claimed as a pass
  for that full source shape.
- debug.ts:203: throwing the stored `e` is separately refused by current throw
  lowering. Try/catch/finally and thrown-value lowering remain compiler-owned.
- moduleNameResolver.ts:1681: `failedLookupLocations?.join(", ")` is refused as
  an optional call.
- tracing.ts:66: the stock embedded checker rejects `e.message` on the original
  unknown catch variable (TS18046) before library lowering. Its `e.message || e`
  coercion is not approximated.

The inventory is reproducible with `docs/library-errors/inventory.cjs`, passing
the TypeScript module, compiler source directory and a JSON output path outside
the checkout. It parses the vendored compiler AST; it does not count comments.

Reproduction (source `/workspace/adamic-tools/env.sh`; keep stdout/stderr under
`/tmp/library-errors-*`; use `GOMAXPROCS=2` and `go ... -p 2`):

1. Run the scoped adamic-test262 runner with `-adapt -json -jobs 2`, the pinned
   external checkout and the three Error filters above.
2. Set ADAMIC_ERROR_TEST262 to that checkout and ADAMIC_ERROR_TEST262_REPORT to
   its JSON report; run `go test ./cmd/adamic-test262 -run TestErrorTest262Backends`.
3. Set ADAMIC_ORACLE_WASI=1; run the Error-only native/WASI oracle subtests and
   TestErrorObjectsMutants / TestErrorCaptureMutants.
4. On Linux run `go test ./internal/oracle -run TestErrorCountsAreRecorded -args
   -update-counts`, then verify without the update flag.
