Built: load pinned Node declarations for an unresolved NodeJS annotation, preserving local namespaces and named refusals.
Commits: checked base 3ce9a33b; measured fix 7d2611f744902fa1e98d3dac3722177dbda79e36; report commit follows.
Checks: targeted loader tests PASS (1.692s), Node/native fixtures PASS (16.500s), counts refresh PASS (61.508s).
Mutants: all four failed their intended tests; no build failure counted as a kill.
Limits: this proves declaration resolution and erased annotation compilation; live Timeout handles and timer ABI are outside this change.

The scout input is adaptation 1e920a31, read through a detached worktree without merging its branch. All 79 compiler source hashes named by scout 4e6121d3 match. The stock manifest covers all 82 compiler files. The delivery branch was advanced to the requested compiler/any-returns-next 3ce9a33b, which already contains the previous generic-return fix.

The reduction is one annotation: `type TimerHandle = NodeJS.Timeout | undefined;`. Before lowering, the base checker reports TS2503, “Cannot find namespace NodeJS”. No pinned Node declaration files are loaded. The reference prints as `NodeJS.Timeout` but has TypeFlagsAny; the union also has TypeFlagsAny. The display name alone conceals the error type.

The cause is in the loader. `compilerOptions` disables automatic @types discovery with an empty Types list. The old `usesNodeModules` in internal/load/node_library.go only requests the pinned package after a top-level node: import or export. An ambient namespace annotation does not satisfy that test. At internal/load/load.go:130, `usesNodeTypes` now also finds qualified NodeJS type references and requests the existing pinned package when the namespace is unresolved. At internal/load/node_library.go:94 it checks both a nil symbol and CheckFlagsUnresolved: the checker can return a synthetic unresolved alias. A resolved local NodeJS namespace keeps its own meaning.

The exact checker path is in the existing cohere submodule, with no code copied or changed: checker.go `resolveQualifiedName` at 16214 fails to resolve NodeJS; `resolveTypeReferenceName` at 23611 falls through to `getUnresolvedSymbolForEntityName` at 23626; that function stores the unresolved intrinsic type on a synthetic type alias at 23656. `getTypeFromTypeAliasReference` at 24104 handles CheckFlagsUnresolved by creating a TypeFlagsAny error type at 24111. This happens before concrete lowering and mapper substitution. The generic-return mapper is not involved.

After the fix, the same annotation has no checker diagnostic. The reference is the actual Timeout class declared in @types/node 25.3.3 timers.d.ts, with no any flag. The union has exactly that checker type object and undefined. Eight loader tests cover the declaration identity, alias references, a local namespace, a local missing member, and unresolved namespaces/members. The latter continue to fail with TS2503 or TS2694 and the missing name. No fabricated type or permissive fallback was introduced.

The `.a` fixture is held to Node in both Adamic backends by the existing oracle, including sanitized and release native execution. It prints `timer type resolved`. Its alias erases, so this fixture alone does not prove runtime representation of a Timeout; the loader identity and union assertions prove the annotation, and the oracle proves its accepted program behavior.

Commands run, with output redirected to logs:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/timeout-setup.log 2>&1
source /workspace/adamic-tools/env.sh
npm ci --prefix stage3/api --ignore-scripts --no-audit --no-fund > /tmp/timeout-node-install.log 2>&1
go test ./internal/load -run 'TestNodeLibrary|TestNodeNamespace' -count=1 -v > /tmp/timeout-load-tests.log 2>&1
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(node_namespace_timeout|any_returns_concrete|node_)' -count=1 -v > /tmp/timeout-oracle-tests.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/timeout-counts.log 2>&1
python3 -B stage3/census/any-returns-concrete/step05/mutants.py > /tmp/timeout-mutants.log 2>&1
```

Setup succeeded in 44.205s: node 0.026s, Go 0.030s, submodules 0.078s, markdown 0.085s, clang 0.175s, build 44.036s, deferred test binaries 44.171s, cache warm 44.173s. nproc is 5, CPU quota is 4. Node 24.19.0, Go 1.27.1, clang 20.1.8. The environment restarted during an earlier measurement; its incomplete ledger was discarded and the complete before run restarted. No partial result is reported.

The required counts refresh records zero allocations for the erased annotation. It also moves an unchanged logical_and_reference_maybe row and removes a stale taste/17_binder_flow row already excluded by the requested base. No existing fixture's recorded numbers change. No whole package test or full gate was run.

| Mutant | Check that failed |
| --- | --- |
| Ignore namespace requests | TestNodeNamespaceAnnotationUsesPinnedTimeout, named TS2503 load failure |
| Ignore unresolved alias flags | TestNodeNamespaceAnnotationUsesPinnedTimeout, same missing namespace |
| Replace a local namespace with the Node package | TestNodeNamespaceLocalDeclarationIsNotReplaced |
| Accept unresolved types despite diagnostics | TestNodeNamespaceUnresolvedTypesStayRejected |

The mutant runner restores each file and requires the test's failure witness, rejecting build failures. The original one-annotation test fails on the requested base; that failure and all final mutant logs are preserved in evidence.

Both complete ledgers have 79 measured source files, checked against an unchanged 82-file stock manifest. Both binaries were built with vcs.modified=false; the before binary identifies 3ce9a33b and the after binary identifies 7d2611f7. The latent tool uses LATENT_FULL=1 and LATENT_ASSERT_NO_OUTPUT=1; it measures refused boundaries on checker-rejected programs and produces no program output. Hidden bytes use the unchanged outermost-cause attribution, so a removed any diagnostic can expose another refusal instead of making those bytes compile.


| Measurement | Base 3ce9a33b | Fix 7d2611f7 | Removed |
| --- | ---: | ---: | ---: |
| Type-any boundaries | 116 | 103 | 13 |
| Credited hidden bytes for type-any reasons | 28,321 | 28,154 | 167 |
| Type-any diagnostic sites | 135 | 123 | 12 |
| clearTimeout annotation boundaries | 1 | 0 | 1 |
| clearTimeout annotation credited bytes | 74 | 0 | 74 |
| Function-return-any boundaries | 8 | 3 | 5 |
| Function-return-any credited bytes | 1,010 | 1,010 | 0 |
| All hidden bytes, across reasons | 3,652,681 | 3,642,530 | 10,151 |

The type-any bucket consists of NotYet/Refused reason text containing the type word `any`. It excludes the English quantifier in “in that narrows a union of object types (any object with the key passes)”: those eight unchanged boundaries have zero credited bytes. A purely textual match would report 124 to 111 boundaries, with identical byte totals and deltas. Boundaries are deduplicated by file and span; one boundary can have multiple reasons, so reason-row counts do not sum to the total. The original partition assigns byte credit to one outermost representative. This explains why five return-any boundaries disappear without changing that reason's credited bytes.

The clearTimeout declaration at sys.ts:52:31 used to fail with “a value of type any”. It now gets past the signature and refuses with “a function without a body” at sys.ts:52:1. It still spans 74 bytes. The resolved namespace also lets the package supply the existing Node globals and module declarations. Twelve any diagnostic sites disappear: performanceCore.ts:64; sys.ts:52,1630,1673,1747,1773,1888,1927,1940,1950; tracing.ts:41,222. No new any diagnostic site appears. Checker diagnostic sites fall from 326 to 260; both missing-NodeJS sites disappear. These observations do not assert that every remaining compiler checker error is fixed.

Remaining type-any reasons:

| Reason | Boundaries | Credited bytes | Why it remains |
| --- | ---: | ---: | --- |
| a value of type any | 65 | 25,881 | The current lowering still observes any at these value sites; this loader change does not supply missing types for other names or remove explicit any. |
| a function returning any | 3 | 1,010 | convertToObject and tryParseJson explicitly return any; objectAllocator.getNodeConstructor returns Node as any. |
| a generic function whose type argument R is any or unresolved | 1 | 430 | maybeBind on the existing System.setTimeout signature at watch.ts:679; the checker still supplies an any or unresolved R. |
| a call returning any | 3 | 336 | Two emitter parenthesizer calls and System.setTimeout in sys.ts:378 still have an any return in their declared interfaces. |
| Array.isArray on a tuple or an erased object/any/unknown view | 2 | 204 | The current input view still falls outside the supported narrowing proof. |
| a generic function whose type argument T is any or unresolved | 4 | 190 | Three fill(Array(...), ...) calls in debug.ts inherit the untyped array element; forEach on exports in moduleSpecifiers.ts:1105 still has an any or unresolved T. |
| storing any in a field | 1 | 55 | The stored value remains any; resolving NodeJS cannot make that store sound. |
| an array of any | 24 | 48 | These arrays still have any elements; no element proof is introduced by loading the Node declarations. |

All remaining diagnostic locations and reasons are in RESULT.json. Unresolved or actual any types remain refused; the fix does not cast them to Timeout or hide their reasons.

Measurement commands:

```sh
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 /tmp/timeout-before-census /tmp/timeout-scout-tree/src/compiler /tmp/timeout-before-retry.jsonl > /tmp/timeout-before-retry-run.log 2>&1
python3 stage3/census/latent/make_overlay.py /workspace/adamic /tmp/timeout-delivery-overlay > /tmp/timeout-delivery-overlay.log 2>&1
go build -overlay /tmp/timeout-delivery-overlay/overlay.json -o /tmp/timeout-delivery-census ./stage3/census/latent/tool > /tmp/timeout-delivery-build.log 2>&1
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 /tmp/timeout-delivery-census /tmp/timeout-scout-tree/src/compiler /tmp/timeout-after.jsonl > /tmp/timeout-after-run.log 2>&1
python3 -B stage3/census/any-returns-concrete/step05/summarize.py /workspace/adamic /tmp/timeout-scout-tree/src/compiler /tmp/timeout-stock.json /tmp/timeout-before-retry.jsonl /tmp/timeout-after.jsonl stage3/census/any-returns-concrete/step05/RESULT.json 7d2611f744902fa1e98d3dac3722177dbda79e36 > /tmp/timeout-summary.log 2>&1
```

Both full runs exited zero. The before executable was built on the requested base before any production edits; the after executable was built on the clean committed fix. Build records, full compressed ledgers, the stock manifest, hashes, trace driver, trace outputs, test outputs and mutant outputs are checked in beside this report. The summary uses the original census attribution scripts by reading revision 6c4fc1af; it does not merge another worker's branch. No protected lowering, emitter, or oracle implementation was changed.
