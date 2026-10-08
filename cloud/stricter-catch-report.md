Built preserved caught values, truthful Error identity, Node member reads and named checks at definite typed uses.
Base 390af985caf2a81edb239cdf5e3238add50a7fcd; step 1 a77b4c46bbd6e671cd3a3bb932f04a733c2defc5; step 2 is the commit containing this report.
Final sequential affected gate and complete uncached Node oracle pass; Linux counts keep 509 existing rows unchanged and add five; vet passes.
Six mutant policies are caught: every Error, definite read, erased typed-use guard, MayThrow under-approximation, missing null tag, and unsupported member admission.
Limits: unclassified message/code data reads are admitted; other names, possible accessors/methods/static constructor fields and optional access need a narrowed receiver; whole stage 3 compilation is not claimed.

The binding ruling is recorded in docs/caught-values.md, including the refused .a source and exact diagnostic. A .ts project's declared catch type remains TypeScript's input to checking. Adamic carries the caught value as tagged unknown rather than trusting project any. A member read preserves its actual value and absence. Only a typed use checks the promised primitive representation; a definite string parameter, field or return stops with exit 70 and `adamic: panic: adamic/catch-type: caught value does not match its typed use`. Nullable receivers instead raise an actual catchable TypeError, propagate MayThrow and unwind through cleanup and finally. No check is inserted to claim that an unclassified member read is already a string. The inserted typed-use guards and their source locations appear in `ir.InsertedChecks` as `caught-type`.

The ledger is a1a16427a46149435e25f847c45ad136f1f1e55c. Its five sites are distributed across three files, rather than all being in sys.ts's require. `caught_sites/main.ts` mirrors their use patterns and is held to source Node in both backends:

| Ledger site | Original location | Behavior under the ruling |
|---|---|---|
| D108 | commandLineParser.ts:2301:91 | Read e.message, then check at createCompilerDiagnostic's definite string argument. |
| D128 | program.ts:406:25 | Read e.message, then check at onError's definite string argument. |
| D132 | program.ts:441:25 | The second onError message argument has the same checked use. |
| D152 | sys.ts:1281:43 | e.code === 'ENOSPC' preserves an absent field as undefined; no definite-string check belongs here. |
| D153 | sys.ts:1617:13 | Returning { error: e } preserves the actual tag even through the declared Error-like record view. |

These are pattern probes for the five sites, not a claim that the entire TypeScript compiler lowers. Three sites require primitive guards; the comparison and returned record require truthful representation rather than an unconditional read guard.

The pinned base has MakeError and the runtime Error layout, rather than error_classes.go. Error's two fields, name and message, are unchanged. Step 1 changes the exception carrier from adamic_object * to adamic_heap *, with a separate pending flag: thrown undefined is NULL as a value, but still has an exception to propagate. Catch and finally transfer or restore the owned carrier and pending flag together. Step 2 adds an immortal null tag, IR Box.Null/Field.Null metadata and checked Narrow metadata. Native shape identities distinguish actual scalar representations and nullable layouts, so a boolean or number field cannot be read through a reference slot. None of those changes alters Error's runtime field layout.

Catch-origin preparation follows aliases and returned records before lowering bodies. Only catch-derived generic fields change their slot representation. Definite primitive fields check their initializer and retain their declared layout, so a later ordinary string write still works. Ordinary unknown contextual fields retain their existing initializer layout. This last restriction restores String.raw's borrowing and all existing counts after an intermediate implementation introduced extra retains/releases. Step 2's final counts leave all 509 existing rows exactly equal to step 1; the five added probes are in stricter-catch-counts.md.

`catch_values.a` covers Error, built string, number, object and undefined throws, truthful identity, returning each caught value and undefined through finally. `caught_members/main.ts` covers the five required message probes, number and null throws, null aliases, numeric/boolean/null fields and a nullable class data field, exact nullish TypeError messages, aliases and shorthand returned records. `caught_sites/main.ts` covers the ledger patterns, optional string returns, cross-function TypeError propagation through finally and a checked string record followed by an ordinary write. The parameter, field and return failure probes independently pin exit 70 and the exact named stderr in both backends. `TestAdamicCaughtMessageRequiresProof` pins `main.a:2:40: error TS18046: 'e' is of type 'unknown'.`

An unclassified function name, primitive length, Error stack or prototype method cannot be implemented by the native own-field reader. Those names are refused before emission; message/code names that could target a source method, accessor or static constructor layout are also refused, including imported targets. Narrowing to the declared receiver permits the ordinary typed lowering. Opaque built-in values already have a separate named refusal outside equality/typeof; this unit does not admit them through catch storage.

The GraphQL gate discovered that its gap 2 expectation was obsolete: Error made in another function is now throw-able. The gap is marked closed and tested against Node under sanitizers. The port's existing message-returning helper API remains compatible and is unchanged. The first step's explicit-root loader fix is retained: production roots excluded by tsconfig are audited rather than rejected before the audit; its test still proves an ordinary wrong type is refused.

Mutants, all actually run and restored before the final checks:

| Mutant | What caught it |
|---|---|
| Every caught value is Error | Source Node stdout differs for strings, numbers and ordinary objects in both backends, which otherwise exit 0. |
| A member read is a definite string | Both backends stop early with exit 70 for a missing message, while source Node exits 0 and prints undefined. |
| Remove the definite-use guard | JavaScript continues with `[undefined]` and exit 0; native UBSan catches the null string access. Both fail the pinned exit-70 contract. |
| Clear maybeMessage.MayThrow | Native exits 0 but loses the outer caught TypeError line after finally; source Node stdout catches the missing unwind. |
| Drop the null tag | The sanitized mutant exits 0 without stderr, but typeof/null equality and null-field observations differ from Node. |
| Admit every unclassified property | The source mutant makes length/name/stack/toString and possible local/imported message-method and static constructor field probes compile. All named-refusal expectations fail. |

The first five mutants are executable tests in internal/oracle/catch_values_test.go. The last is a source mutation of supportedCaughtMember, restored immediately, caught by internal/lower/catch_values_test.go. Logs: /tmp/stricter-catch-step2-last-probes.log and /tmp/stricter-catch-property-policy-mutant-final.log.

Commands and observations, with /workspace/adamic-tools/env.sh sourced and output saved to logs:

- Setup: bash cloud/setup.sh, 56.993s total; Go 0.049s, Node 0.048s, submodules 0.110s, markdown 0.132s, clang 0.339s, Go build 56.778s, cache warm 56.956s. nproc 5, quota 4 CPUs. /tmp/stricter-catch-setup.log.
- Step 1 complete affected packages and uncached oracle passed after the explicit-root fix; oracle 193.498s, counts 95.751s. /tmp/stricter-catch-step1-oracle-final.log and /tmp/stricter-catch-step1-counts-final.log.
- Step 2's first all-repository run caught overly broad unknown-context scalar boxing, unsound field-slot fitting and stale gap expectations. Those implementations were corrected; the final test262 runner, fuzz, flow, fresh, native and uncached oracle then passed. The JSON.stringify present-value test independently passed all five cases in 6.396s after restricting boxing.
- Final affected gate: `ADAMIC_GATE_UNCACHED=1 go test -p 1 ./internal/load ./internal/lower ./internal/ir ./internal/javascript ./internal/flow ./internal/fresh ./internal/native ./internal/oracle ./stage1/cohere/graphql ./stage3/stricter-options ./cmd/adamic-test262 -count=1 -timeout 30m`, exit 0. Load 9.251s, lower 26.048s, IR 18.498s, flow 106.716s, fresh 22.788s, native 175.940s, complete uncached oracle 218.370s, GraphQL 34.998s, stricter-options 9.043s, test262 runner 47.097s. JavaScript has no package tests; the oracle compares that backend to source Node. /tmp/stricter-catch-step2-sequential-final.log. Final Linux count regeneration: /tmp/stricter-catch-step2-counts-policy-final.log, pass in 62.552s; 509 unchanged rows, five additions. Final vet: the same affected internal packages, exit 0, /tmp/stricter-catch-step2-vet-sequential-final.log. git diff --check passes.
- An intermediate concurrent gate had stale new-probe counts and a killed Node process during the million-code-point comparison. The cgroup recorded two OOM kills; the same frozen compiler passed that native suite in the final sequential run. Updating the probe counts before the final gate removes the table mismatch. /tmp/stricter-catch-step2-affected-final.log records that unsuccessful intermediate run.
- The additional all-repository attempt is /tmp/stricter-catch-step2-full-gate-final.log. Its obsolete GraphQL gap expectation is fixed and rerun separately; its mixed .a/.ts checker-ownership refusal is outside the catch unit and reproduces on the isolated step-1 baseline with its tracked root tsconfig present: /tmp/stricter-catch-jsx-baseline-owned.log, same named refusal, exit 1. The first sparse baseline lacked that config and was not equivalent. The loader ownership guard itself is unchanged from base 390af985. The broad run was interrupted after confirming this blocker and memory pressure; it is not presented as completed. This report does not claim a green all-repository run.

Only codex/stricter-catch-variables is pushed. No main or area branch was pushed or merged into. The full test262 corpus and whole stage 3 compiler lowering are not covered; the test262 runner's tests are covered. Stage 3 status records and Node observations are unchanged.
