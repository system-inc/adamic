# Host compiler content audit

Base: `031a1259bc7973934792dc6cb1bd4074fc2204b9`. All observations below used a binary built from this unchanged main. Each source fixture was extracted byte for byte from its source tip. Native builds use `--sanitize`, `ASAN_OPTIONS=detect_leaks=1:halt_on_error=1` and `UBSAN_OPTIONS=halt_on_error=1`; JavaScript runs through `oracle/node.mjs`.

| Topic | Source tip / containing branch | Main content | Fixture on main | Both backend observations |
| --- | --- | --- | --- | --- |
| non-null | `a02613ef` / origin/codex/host-proof-combined, origin/codex/non-null-check, origin/codex/non-null-narrowed-number | partial | `non_null_deinitialize_initialized.a` | /workspace/adamic/.host-commits-proof/non_null_deinitialize_initialized.a:2:8: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| non-null-narrowed | `3a6c8231` / origin/codex/host-proof-combined, origin/codex/non-null-narrowed-number | partial | `non_null_storage_local.a` | /workspace/adamic/.host-commits-proof/non_null_storage_local.a:1:21: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| nested-empty | `b3578751` / origin/codex/array-literal-never-element, origin/codex/host-proof-combined | complete | `array_literal_empty_first.a` | Compiles, agrees with fresh Node |
| concat | `998fb3eb` / origin/codex/host-proof-combined, origin/codex/scanner-expressions | partial | `scanner_expressions/primitive_concatenation.a` | /workspace/adamic/.host-commits-proof/scanner_expressions/primitive_concatenation.a:3:13: stage 0 can't lower a BinaryExpression with a string and a value yet |
| phantom | `d90994da` / origin/codex/host-proof-combined, origin/codex/phantom-brands | partial | `phantom_overload_results.a` | /workspace/adamic/.host-commits-proof/phantom_overload_results.a:8:14: Adamic 0.1 refuses a cast the runtime can't check; use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast) |
| memoize | `3550dbc4` / origin/codex/host-proof-combined, origin/codex/memoize-regions | none | `memoize_regions/host14.a` | /workspace/adamic/.host-commits-proof/memoize_regions/host14.a:3:21: Adamic 0.1 refuses 'callback', a variable a function value captures and can be reached from what it holds, so the function holds the variable and the variable holds the function: a cycle reference counting can't free; remove the captured strong back-reference, use a module function declaration that captures nothing, or declare the variable Weak<...> and keep the function somewhere strong (adamic/cycle-capable) |
| predicates | `a58ba402` / origin/codex/host-proof-combined, origin/codex/proven-predicates | partial | `proven_guards.a` | Compiles, agrees with fresh Node |
| generic-empty | `4e022aae` / origin/codex/host-proof-combined, origin/codex/parser-generic-empty-array | partial | `generic_empty_array.a` | /workspace/adamic/.host-commits-proof/generic_empty_array.a:33:22: Adamic 0.1 refuses a value of type never seen as T, a type parameter whose constraint { text: string; } can be written, so it can write what never can't hold; take it as never, or constrain T to something readonly, which can't write (adamic/invariant-mutable) |
| debugger | `7d2cbc89` / origin/codex/debugger-statement, origin/codex/host-proof-combined | complete | `debugger_fail.a` | Compiles, agrees with fresh Node |
| generic-value | `5d1b45e18` / origin/codex/generic-function-value | none | `generic_function_value.a` | /workspace/adamic/.host-commits-proof/generic_function_value.a:4:47: stage 0 can't lower a generic function as a value yet |

Main contains contextual nested-empty literal lowering and debugger IR/native no-op/JavaScript preservation. Concatenation already spells numbers and booleans but omits null, undefined and optional scalars. Generic empty-literal specialization exists, but bottom-to-readonly fallback proofs are missing. Proven guards and overload marker predicates exist, but the source topic additionally fixes computed-name and failed-generic cleanup panics. Phantom-brand overload admission must retain main's stronger overload checks, generic mapping and argument-count ABI. Generic function values are still explicitly NotYet.

## Dependencies and scope

The incoming non-null `.a` forms conflict with main's explicit refusal policy and its eager checked `.ts` semantics. Lazy placeholders/deinitialization require the absent placeholder-nonnull landing; they remain out. Narrowing/scalar checks already exist on main for checked TypeScript input. No syntax policy is weakened in this batch.

Memoize requires runtime graph-region capture-cell/environment ownership (runtime `d70bc6b1`, carried by source `3550dbc4`). Main has step-04 closure frames, not GraphCell/GraphClosure or synchronous cycle graph ownership. That runtime dependency stays out.

Views beyond main slice 1, optional-presence landings and generic nullable-array views carried by the generic-value topic remain out. Only contextually specialized generic function values and their identity guards are candidates. No old topic base is merged.

## Combined host input on main

The 25 sources and recorded Node observations are taken unchanged from `d71dfdfa`. `observe-disagreements.py --all --compiler /tmp/host-commits/main-adamic --logs /tmp/host-commits/host-main --report /tmp/host-commits/host-main.json` completed, exit 1 for the recorded stops. Fresh Node matches every recorded observation. Both backends compile and agree on exactly 05_writeFile, 11_setModifiedTime, 13_createDirectory and 24_useCaseSensitiveFileNames (4/25). The library proof's 24/25 is not a main observation. This initial audit changed no status record. The seven permitted final stage0 updates are listed below; all Node observations remain byte-identical.

## Landing decisions

No old topic base was merged. The independent net changes were rebuilt on main, one implementation commit per topic. There were no Git merge conflicts; the reconciliations below concern behavior and newer main implementations.

| Topic | Source | Delivery | Meaning preserved |
| --- | --- | --- | --- |
| non-null | a02613ef | missing parts excluded | Main eager checked TypeScript and explicit Adamic assertion refusal retained; old lazy placeholder/deinitialization semantics need placeholder-nonnull. |
| narrowed non-null | 3a6c8231 | existing scalar/narrowing checks retained; dependent Adamic forms excluded | Checked TypeScript narrowing and union/scalar representation are tested, including both original main mutants. |
| nested empty arrays | b3578751 | already complete | Existing contextual element identity and source fixture agree with Node. |
| string concatenation | 998fb3eb | b89243521 | Literal null/undefined and optional scalar spelling added; main number/boolean/optional-string spelling and operator dispatch retained. |
| phantom brands | d90994da | 5b7d82fc3, primitive subset only | Reuse independent primitive implementation from d2d3c77e3; main overload, mutability, argument-count ABI and view checks retained. Array-brand presence hooks stay out. |
| memoize | 3550dbc4 | excluded | Depends on graph capture-cell/environment runtime d70bc6b1. No GraphCell/GraphClosure dependency brought in. |
| proven predicates | a58ba402 | 768fa194b | Main guard admission retained; add computed-name guard and state restoration on failed instantiation. |
| generic empty arrays | 4e022aae | 3a0f823c2 | Slotless empty literal identity and readonly fallback added; mutable reverse views remain checked. Scalar truthiness stays main; array OR coalesces only undefined. |
| debugger | 7d2cbc89 | already complete | Existing IR, flow no-op, native omission and JavaScript debugger statement retained. |
| generic function values | 5d1b45e18 | bc9fc1ccf | Concrete callable contextual specialization and identity guards added; nested closure owners, type identity keys and main viewSite gating retained. |

The two interaction commits preserve existing main proofs. c0e78ac24 corrects the generic-value scan to recognize qualified generic namespace calls as direct calls, rather than detached values. 5c429fb28, the final phantom/overload interaction, exempts only aliases with at least one reference and every reference in a bodyless overload declaration. It preserves overload_ancestor_directory.a, while nonvoid brands used in runtime signatures and unused invalid brands still refuse. Main's strict overload-result refusal pin remains unchanged.

The first counts sweep exposed three new lowering stops: params_namespaces_dotted.a and namespaces.a were incorrectly reported as specialized-function identity observations; overload_ancestor_directory.a was incorrectly rejected as a nonvoid runtime brand. Both interactions were fixed. All three unchanged fixtures subsequently passed fresh Node comparisons in both backends, including native release, ASan/UBSan and allocation leak checks. The final Linux counts table adds exactly 12 rows and changes no existing row.

Additional unchanged-main audit evidence: the exact a58ba402 computed_relation.a panics in both lowering backends with `Unhandled case in Node.Text: *ast.ComputedPropertyName`; generic_closure.a panics with `slice bounds out of range [:-1]`. qualified.a lowers on main. The landed guards turn the two panic reductions into named optional-field and generic-mutation refusals; they cannot execute in either backend. These observations distinguish the missing cleanup from already passing proven_guards.a.

## Final host observations

The host input sources and status.json are byte-identical to d71dfdfa, verified against Git blobs in [host-inputs.json](host-commits-evidence/host-inputs.json). The source report's observe-disagreements.py is used with only the input directory redirected to extracted sources and `--sanitize` appended after the output argument. ASAN_OPTIONS is detect_leaks=1:halt_on_error=1 and UBSAN_OPTIONS is halt_on_error=1. Both backend observations and exact first stops are in [host-final.json](host-commits-evidence/host-final.json).

Fresh Node matches every one of the 25 recorded observations. Exactly 05_writeFile.a, 11_setModifiedTime.a, 13_createDirectory.a and 24_useCaseSensitiveFileNames.a compile and agree in both backends (4/25). The other 21 retain named Refused or NotYet stops; the host observation command exits 1 to report those stops. The library proof's 24/25 also requires library/runtime landings outside these ten topics. This batch does not claim that result on main.

| Host fixture | Both backends | First stop |
| --- | --- | --- |
| 01_readFile_utf8.a | Refused | 01_readFile_utf8.a:20:21: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| 02_readFile_utf16le.a | Refused | 02_readFile_utf16le.a:20:21: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| 03_readFile_utf16be.a | Refused | 03_readFile_utf16be.a:20:21: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| 04_readFile_missing.a | Refused | 04_readFile_missing.a:20:21: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| 05_writeFile.a | Compiles | Agrees with fresh Node |
| 06_fileExists.a | NotYet | 06_fileExists.a:53:5: stage 0 can't lower node:fs.symlinkSync yet |
| 07_directoryExists.a | NotYet | 07_directoryExists.a:53:5: stage 0 can't lower node:fs.symlinkSync yet |
| 08_getDirectories.a | NotYet | 08_getDirectories.a:186:25: stage 0 can't lower node:fs.readdirSync yet |
| 09_realpath.a | NotYet | 09_realpath.a:16:17: stage 0 can't lower node:process.Process.platform yet |
| 10_getModifiedTime.a | NotYet | 10_getModifiedTime.a:36:5: stage 0 can't lower node:fs.symlinkSync yet |
| 11_setModifiedTime.a | Compiles | Agrees with fresh Node |
| 12_deleteFile.a | NotYet | 12_deleteFile.a:22:5: stage 0 can't lower node:fs.symlinkSync yet |
| 13_createDirectory.a | Compiles | Agrees with fresh Node |
| 14_getCurrentDirectory.a | Refused | 14_getCurrentDirectory.a:12:28: Adamic 0.1 refuses 'callback', a variable a function value captures and can be reached from what it holds, so the function holds the variable and the variable holds the function: a cycle reference counting can't free; remove the captured strong back-reference, use a module function declaration that captures nothing, or declare the variable Weak<...> and keep the function somewhere strong (adamic/cycle-capable) |
| 15_getExecutingFilePath.a | NotYet | 15_getExecutingFilePath.a:21:52: stage 0 can't lower node:path.dirname yet |
| 16_getEnvironmentVariable.a | NotYet | 16_getEnvironmentVariable.a:12:12: stage 0 can't lower node:process.Process.env yet |
| 17_write.a | NotYet | 17_write.a:12:5: stage 0 can't lower node:net.Socket.write yet |
| 18_exit_0.a | NotYet | 18_exit_0.a:24:30: stage 0 can't lower node:process.Process.exit yet |
| 19_exit_1.a | NotYet | 19_exit_1.a:24:30: stage 0 can't lower node:process.Process.exit yet |
| 20_exit_2.a | NotYet | 20_exit_2.a:24:30: stage 0 can't lower node:process.Process.exit yet |
| 21_createHash.a | Refused | 21_createHash.a:16:18: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |
| 22_createHash_fallback.a | NotYet | 22_createHash_fallback.a:22:1: stage 0 can't lower a function without a body yet |
| 23_newLine.a | NotYet | 23_newLine.a:11:31: stage 0 can't lower node:os.EOL yet |
| 24_useCaseSensitiveFileNames.a | Compiles | Agrees with fresh Node |
| 25_readDirectory.a | Refused | 25_readDirectory.a:522:27: Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case |

## Mutants

Temporary Go overlays revert only a topic's core change. All nine exits are 1 with a named failed test and no Go build failure. The excluded memoize graph implementation has no landed core to revert; its source host14.a remains a named cycle refusal, and no successful execution or mutant is claimed for that dependency.

| Topic | Reversion | What caught it |
| --- | --- | --- |
| non-null | bypass checked nonNullValue | undefined TypeScript fixture loses its required checked-site count |
| narrowed non-null | look through scalar check into union argument | narrowed fixture loses proven-site count; original IR mutant also catches union return before either backend |
| nested empty arrays | disable empty literal target proof | TestNestedEmptyArrayElementKinds fails |
| concatenation | revert to number/boolean-only conversion | primitive_concatenation.a fails lowering |
| phantom primitives | disable phantomBase | TestPhantomCastsAreErased fails lowering |
| predicates | remove computed-name guard | TestCensusComputedRelation panics instead of named refusal |
| generic empty arrays | remove bottom-array readonly exemption | generic_empty_array.a fails lowering |
| debugger | replace debugger with refusal | TestDebuggerNativeEmitsNothing fails |
| generic function values | restore generic-function-as-value NotYet | TestGenericFunctionValuesHaveSeparateInstances fails |

Three extra overlays revert generic-state cleanup, the qualified namespace-call repair, and the overload-only brand exception. Each fails its own regression test. [mutants.json](host-commits-evidence/mutants.json) and the log archive record commands and diagnostics.

Runtime IR mutants also run: number spelling +1, inverted boolean spelling, wrong null spelling, wrong undefined spelling, empty array treated as falsy, fresh arrays shared, and generic function returning zero. Each wrong result exits 0, stays sanitizer/leak clean, and differs from Node stdout in both backends. The existing non-null no-check mutant likewise exits 0 incorrectly and is caught by its check/output proof; the existing union look-through mutant fails representation proof before backend execution. These are deliberate mutants, not unmodified compiler output. No unmodified wrong output with exit 0 was observed.

## Stage 3 records

No Compiles fixture becomes Refused or NotYet. Only these seven stage0 values change. Every byte outside stage0, including every recorded Node observation, remains identical to main. [status-changes.json](host-commits-evidence/status-changes.json) records both diagnostics.

| Fixture | Change | Cause |
| --- | --- | --- |
| namespaces/02_jsx_names.a | Refused to Compiles | primitive phantom brands, 5b7d82fc3 |
| namespaces/03_react_names.a | Refused to Compiles | primitive phantom brands, 5b7d82fc3 |
| cycles/04_directory_callback/main.a | NotYet to Compiles | generic function values, bc9fc1ccf |
| cycles/05_safe_load_read/main.a | NotYet to Compiles | generic empty arrays, 3a0f823c2 |
| cycles/06_import_order_mutant/main.a | NotYet to Compiles | generic empty arrays, 3a0f823c2 |
| assertions/16_unknown_chain.a | stronger Refused: nonvoid __watchFileKind declaration, rather than later unchecked cast | primitive phantom brands, 5b7d82fc3 |
| assertions/17_brand_upcast.a | stronger Refused: nonvoid __escapedIdentifier declaration, rather than later unchecked cast | primitive phantom brands, 5b7d82fc3 |

The five positive records use the fixture runner's -update option after fresh Node and native agreement. The two stronger refusals use fresh compiler diagnostics and exact stage0 JSON-value replacement. A raw-byte verifier removes only each stage0 value and compares the rest of the files to origin/main. The whole stage3 fixture sweep then passes.

## Verification and limits

[commands.txt](host-commits-evidence/commands.txt) contains the exact final selections; [logs.tar.gz](host-commits-evidence/logs.tar.gz) preserves setup, audit, host, topic, sweep and mutant output. Tests write directly to log files. No full oracle or whole-package gate is claimed. The primary topic oracle selection executes 18 positive fixtures; a separate selection executes all four nested-empty fixtures, and the qualified-cast reduction is also compared with Node under sanitizers. These selections execute the cases, not just their parent test; both backends, native sanitizers, native release and allocation leaks are exercised. The separate census and generic/brand refusal tests check first stops where execution is impossible.

Setup succeeded: nproc is 5; Node 24.19.0, Go 1.27.1, clang 20.1.8. Timing lines: Node 0.027s; Go 0.029s; markdown skip 0.009s, ready 0.077s; submodules 0.087s; clang 0.178s; build 43.666s; tests deferred 43.821s; cache warm 43.822s; complete 43.854s. GOPROXY is https://proxy.golang.org|direct; environment file is /workspace/adamic-tools/env.sh. The broad remote-ref fetch initially rejected unrelated non-fast-forward tracking refs; the explicit authorized-topic fetch and final main fetch succeeded, without force updates. Final origin/main remains 031a1259bc7973934792dc6cb1bd4074fc2204b9.

Build, internal vet, changed-file a-check, the filtered stage1 Gap/Gaps/Probes sweep, stage3 fixtures and Linux TestCountsAreRecorded are the requested sweep. A-check scans only 18 .a files added or changed against main: 14 checked and four expected refusals. Negative reductions have explicit expected-refusal headers; no negative is silently treated as a passing executable program.

Out: lazy placeholder/deinitializing non-null semantics, graph-region memoize ownership, array phantom-brand optional presence and views beyond main slice 1. Library/runtime topics underlying the combined host proof are not imported. The original full phantom overload fixture's array tail is excluded; its primitive prefix is reduced separately and held to Node. No WASI/macOS, pristine pre-adaptation host control set, or full gate is covered. No protected branch or unrelated worker branch is merged or pushed. No cohere source is copied.

Automatic approval review rejected broader optional-presence and contextual-view imports because they crossed the unit's excluded dependencies. The independent primitive and concrete callable subsets passed review; the rejected integrations remain out. No additional permission was needed to finish those authorized subsets.

This batch lands five independent compiler changes toward #4my3dqg and supports host-fixture task #7fsgdcn, while recording the remaining integration dependencies explicitly.
