Built: loom-r-chain-followers-ac-m, preserving current main and applying the source tip's own change as one commit.
Commits: source a79a5600420f46e3dba79d5e91ca7097232db102; base e1bd97c5034ddaac7c47938f587c304264443b7d; delivery SHA in final response.
Commands and outputs: final whole touched-package suites, counts regeneration, reader/map guard, parallel guard and vet passed; initial failures and exact argv/exits are retained in checks.json.
Mutants: current-tree lowering and/or a-check header proofs are recorded below; no compiler-build failure counts as a catcher.
Not covered: full repository gate, ahra, Loom submission or opt-in lanes that the package suites skipped (skips.json).

Fetched main and source refs. The newest fetched origin/main at branch creation is the base above. Applied only the source tip's own diff using git cherry-pick --no-commit, and independently captured the requested explicit merge-base merge-tree result. No diff against the source branch's whole history was applied. All main work outside the branch's required changes is retained.

Conflict: internal/lower/invariance.go, impliedTarget's ?? / || / && arm. Main commit 477eb9b7 introduced impliedExpressionTarget and its per-parent cache (including cached nil tuple element targets). The source adds a concrete callable context and a contextual-array destination before using the expression's ordinary type. Kept both context checks and retained main's return l.impliedExpressionTarget(parent) as the fallback. Did not restore the old uncached GetTypeAtLocation fallback. Main's conditional and array-literal cached paths and the entire helper/cache remain intact; the conditional callable-context addition merged cleanly. The other source hunks (never input, contextual signature instantiation and contextualGenericType) remain intact. conflicts.patch, merge-tree.log and invariance-resolution.patch give every exact hunk.

No main infrastructure line was reverted. Original source additions/deletions outside invariance/counts match the source-own delta exactly (source-delta.json). The conflict resolution changes only how both additions compose; source's old direct checker fallback is replaced by main's already-existing cache helper. Object lowering auto-merged without a manual edit and preserves main's newer object lowering.

Known boundary expectation: the source proves Path | undefined's primitive representation, so main's element_access_boundaries_test.go paths case reaches the later branded-array element access. Updated that exact expected NotYet to an ElementAccessExpression, retaining its Node check and all refusal/type checks. This is a required expectation change beyond the source, not a relaxed acceptance or a main lowering reversion. The focused paths-boundary check and full oracle suite hold it. This same issue was discovered and fixed by the first branch's complete oracle run.

Counts: even though the count table auto-merged, discarded that result, took main's full table, and regenerated once with the house Linux clang. Exactly 15 added source rows; no changed/deleted existing main row. Every added row has the same values as the source. Runtime source and recut added/deleted lines are both empty, so no runtime C/H resolution line needs clearance.

A-check rechecks all 21 source-added/modified .a programs with this branch's own compiler and the unchanged README-pinned Gate.aCheck. Four existing refusal headers and every program body are retained; both header mutants exercise the same predicate against real captured diagnostics, matching the README method, then restore the program.

Mutation proof: ported the 12 archived host probes as minimal edits over current source. Archived generic-values and string-concatenation snapshots contained unrelated older async/template code; none of those unrelated reversions is applied here. Generic-values keeps just the specialization stop. String-concatenation uses a type-correct null-spelling regression so Node, rather than a C build error, decides the mismatch. Two recorded arrow faults are now held by the source's IR frame tests before C emission, not the archived undeclared-C-identifier catchers. Added independent main-cache and checker-derived operand-context mutants to hold both sides' intent. Original Node/native fixture and built-in IR-mutant tests also run in the whole oracle suite. Full details and exact patches are in mutants/ported-edits.json and mutant-catchers.json.

Environment recovery: the first combined flow/fresh/IR/JavaScript invocation passed flow, fresh and JavaScript, but IR's TestCallTargetReaders and TestProduct_CallTargetExports could not copy exports/279.a because the root volume had no space left. The log is retained in other-touched-all.log. The root product cache occupied 15 GB on a 32 GB overlay, while /tmp was a separate 8.8 GB volume with 8.7 GB free. Used main's supported ADAMIC_BUILD_OUTPUT_DIR=/tmp/adamic-gate/loom-chain-products to place new products on that volume and reran the complete IR package successfully. This changes only the test environment, not products, keys, source or assertions. Continued oracle, guards, vet and mutants with the same setting. disk-space-before.txt and disk-space-after.txt record disk use.


Whitespace: gofmt and git diff --check for Go files pass. The full internal/docs check reports the source commit's existing blank final line in internal/oracle/testdata/host_phantom_overload_primitive.a:33. Retained the original fixture bytes rather than making an unrelated body edit; diff-check.log records it. This was already reported by the source header audit.


Final scope audit: the source also carries stage3/fixtures/{assertions,cycles,namespaces}/status.json updates. Their added/deleted lines are identical to the source. Ran the complete stage3/fixtures/... package and vet, including Node, complete stage0 diagnostics and native checks for compiling fixtures. All passed. All seven programs behind changed status rows remain byte-identical to main; stage3-record-audit.json records their first lines, hashes and status changes. The 21 changed .a files are the a-check set; no unchanged upstream program was rewritten for this audit.

For the extra package run, set ADAMIC_GOCACHE_DIR=/tmp/adamic-gate/loom-chain-go-cache as well as the output directory to keep new Go-cache writes off the almost-full root disk. prepare-cache-view.py creates links to existing immutable cached files and never modifies/removes old data; the Go cache's atomic writer creates/replaces files only in the temporary cache. cache-view.json records the operation. No source, product recipe or key changed.


Counts: took the complete main table, never hand-merged it, then ran the full TestCountsAreRecorded -update-counts on Linux with house clang 20.1.8 after setup compiled the resolved tree. counts.patch and counts.json list the entire delta. Only this change's rows moved; every other main row is preserved. The precise added/deleted lines are below.

Runtime clearance: runtime.json compares source-own and recut added/deleted lines under internal/native/runtime. Additional resolution lines in runtime *.c or *.h: NONE. No runtime approval is needed for an extra resolution line.

Validation: whole-package go test commands are in checks.json. Each uses GOMAXPROCS=4, -parallel 4, -p 1, -count=1, -timeout 12m and JSON output, with an outer per-command hard deadline. Focused guards use -timeout 90s. ADAMIC_GATE_UNCACHED=1 bypasses oracle observation caches. Every requested test command actually ran; no tests were removed. Output goes directly to evidence logs. All measured leaf tests are below 60 seconds; test-times.json/tsv retains every leaf and grouping parent, including failures intentionally produced by mutants. Native/oracle suites include Node agreement, sanitizers, leaks and runtime witnesses. gofmt and Go-source whitespace checks pass. The full internal/docs whitespace check retains one inherited fixture EOF blank line, documented in the resolution notes. Vet covers the changed compiler packages. Opt-in skips are preserved and explicitly listed, not claimed covered.

Setup: GOPROXY=https://proxy.golang.org|direct, bash cloud/setup.sh under a 600-second hard limit; sourced /workspace/adamic-tools/env.sh. nproc=5 and cpu.max=400000 100000 (four CPUs). House clang 20.1.8, Go 1.27.1 and Node v24.19.0. Timing lines:

```text
setup: go ready (0.020s)
setup: node ready (0.021s)
setup: submodules ready (0.057s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.007s
setup: markdown dependencies ready (0.063s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.136s)
setup: shared cache ready (0.785s)
setup: go build ready (40.250s)
setup: test binaries deferred (use --warm-tests) (40.321s)
setup: build cache warm (40.323s)
setup: build-flags commit=e1bd97c5034ddaac7c47938f587c304264443b7d nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=0.50 2.80 3.13 1/337 172509 load-after=2.36 2.96 3.17 1/333 173203
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (40.348s)
setup: source /workspace/adamic-tools/env.sh
setup: logs /tmp/adamic-gate/setup.jKIR9B
```

Assumptions: use the latest fetched origin/main separately for each branch; retain main's newer infrastructure and apply only the source-own change. The newest request explicitly requires whole touched-package suites, superseding the earlier focused-only worker instruction. No cohere code was copied. No PR, ahra command or Loom submission was made. One commit is pushed only to this new own branch, without force. The commit message preserves the original message and adds the requested re-cut line and Ahra co-author trailer.

Commands and observed outcomes:

- counts-regenerate: exit 0, 62.257s.
- paths-boundary: exit 0, 5.392s.
- lower-all: exit 0, 54.882s.
- native-all: exit 0, 210.193s.
- other-touched-all: exit 1, 122.038s.
- ir-all-fixed: exit 0, 10.853s.
- oracle-all: exit 0, 360.436s.
- readers-maps: exit 0, 5.746s.
- parallel-guard: exit 0, 1.22s.
- vet: exit 0, 5.208s.
- mutant-non-null-check: exit 1, 13.589s.
- mutant-non-null-narrowing: exit 1, 13.729s.
- mutant-nested-empty: exit 1, 9.827s.
- mutant-string-concatenation: exit 1, 16.236s.
- mutant-phantom-primitive: exit 1, 9.767s.
- mutant-proven-predicates: exit 1, 9.599s.
- mutant-generic-empty: exit 1, 13.182s.
- mutant-debugger: exit 1, 13.575s.
- mutant-generic-values: exit 1, 9.507s.
- mutant-generic-namespace-interaction: exit 1, 9.553s.
- mutant-phantom-overload-interaction: exit 1, 9.357s.
- mutant-proven-cleanup: exit 1, 9.711s.
- mutant-constructor-context: exit 1, 9.871s.
- mutant-dropped-capture: exit 1, 9.466s.
- mutant-main-implied-target-cache: exit 1, 9.42s.
- mutant-generic-operand-context: exit 1, 9.648s.
- stage3-fixtures-all: exit 0, 179.156s.
- stage3-vet: exit 0, 2.027s.

Counts added lines:

```text
| internal/oracle/testdata/arrow_this_super.a | 5 | 5 | 8 | 8 | 5 | 0 |
| internal/oracle/testdata/arrow_this_field.a | 7 | 7 | 9 | 11 | 6 | 0 |
| internal/oracle/testdata/arrow_this_stored.a | 7 | 7 | 9 | 11 | 7 | 0 |
| internal/oracle/testdata/native-generic-empty-array.a | 3 | 3 | 1 | 5 | 3 | 0 |
| internal/oracle/testdata/generic_empty_array.a | 32 | 32 | 24 | 60 | 17 | 0 |
| internal/oracle/testdata/generic_function_value_arms.a | 18 | 18 | 72 | 99 | 7 | 0 |
| internal/oracle/testdata/generic_function_value.a | 15 | 15 | 27 | 45 | 7 | 0 |
| internal/oracle/testdata/phantom_brands.a | 15 | 15 | 29 | 42 | 6 | 0 |
| internal/oracle/testdata/phantom_undefined.a | 1 | 0 | 5 | 4 | 1 | 0 |
| internal/oracle/testdata/phantom_catch.a | 8 | 8 | 10 | 18 | 2 | 0 |
| internal/oracle/testdata/host_phantom_overload_primitive.a | 25 | 25 | 13 | 43 | 7 | 0 |
| internal/oracle/testdata/host_phantom_optional_array.a | 5 | 5 | 2 | 7 | 2 | 0 |
| internal/oracle/testdata/scanner_expressions/string-number-concatenation.a | 2 | 2 | 3 | 6 | 2 | 0 |
| internal/oracle/testdata/scanner_expressions/number_spellings.a | 85 | 85 | 2 | 88 | 3 | 0 |
| internal/oracle/testdata/scanner_expressions/primitive_concatenation.a | 33 | 33 | 2 | 38 | 3 | 0 |
```

Counts removed lines:

```text

```

Lowering mutant catchers:

- mutant-constructor-context.log: TestArrowFieldCaptureFrames, 0.03s.
- mutant-debugger.log: TestDebuggerNativeEmitsNothing/.a, 0.04s.
- mutant-debugger.log: TestDebuggerNativeEmitsNothing/.ts, 0.04s.
- mutant-debugger.log: TestDebuggerNativeEmitsNothing, 0s.
- mutant-dropped-capture.log: TestArrowSuperCaptureFrames, 0.03s.
- mutant-generic-empty.log: TestNativeAgreesWithNode_059/internal/oracle/testdata/generic_empty_array.a, 0.03s.
- mutant-generic-empty.log: TestNativeAgreesWithNode_059, 0s.
- mutant-generic-namespace-interaction.log: TestGenericFunctionValueNamespaceDirectCallsKeepIdentity, 0.03s.
- mutant-generic-operand-context.log: TestGenericFunctionValueOperandContexts, 0.03s.
- mutant-generic-values.log: TestGenericFunctionValuesHaveSeparateInstances, 0.03s.
- mutant-main-implied-target-cache.log: TestImpliedTargetOrQueriesOnce, 0.03s.
- mutant-nested-empty.log: TestNestedEmptyArrayElementKinds/strings, 0.09s.
- mutant-nested-empty.log: TestNestedEmptyArrayElementKinds/objects, 0.04s.
- mutant-nested-empty.log: TestNestedEmptyArrayElementKinds, 0s.
- mutant-non-null-check.log: TestCheckedNonNullTypeScript/undefined, 0.03s.
- mutant-non-null-check.log: TestCheckedNonNullTypeScript, 0s.
- mutant-non-null-narrowing.log: TestCheckedNonNullTypeScript/narrowed, 0.03s.
- mutant-non-null-narrowing.log: TestCheckedNonNullTypeScript, 0s.
- mutant-phantom-overload-interaction.log: TestPhantomOverloadOnlyBrandsKeepMainProof, 0.03s.
- mutant-phantom-primitive.log: TestPhantomCastsAreErased, 0.03s.
- mutant-proven-cleanup.log: TestCensusGenericRefusalInsideClosure, 0.03s.
- mutant-proven-predicates.log: TestCensusComputedRelation, 0.03s.
- mutant-string-concatenation.log: TestNativeAgreesWithNode_081/internal/oracle/testdata/scanner_expressions/primitive_concatenation.a, 2.86s.
- mutant-string-concatenation.log: TestNativeAgreesWithNode_081, 0s.

Header mutant results:

```json
[
  {
    "path": "review/compiler/chain-followers/host-fixtures/nonvoid-path.a",
    "mutation": "removed-header",
    "exit": 1,
    "caught_by": "unchanged pinned Gate.aCheck replaying the actual branch compiler diagnostic",
    "failure": "a-check failed for 1 file:\nreview/compiler/chain-followers/host-fixtures/nonvoid-path.a: refused, expected checked (adamic: /workspace/adamic/review/compiler/chain-followers/host-fixtures/nonvoid-path.a:3:20: Adamic 0.1 refuses a primitive brand member __pathBrand whose type is not void; make __pathBrand void (or o)",
    "result": {
      "a_check": {
        "review/compiler/chain-followers/host-fixtures/nonvoid-path.a": {
          "outcome": "refused",
          "expected": "checked",
          "first": "adamic: /workspace/adamic/review/compiler/chain-followers/host-fixtures/nonvoid-path.a:3:20: Adamic 0.1 refuses a primitive brand member __pathBrand whose type is not void; make __pathBrand void (or optional and typed undefined) so the brand is phantom"
        }
      }
    }
  },
  {
    "path": "review/compiler/chain-followers/host-fixtures/nonvoid-path.a",
    "mutation": "wrong-reason",
    "exit": 1,
    "caught_by": "unchanged pinned Gate.aCheck replaying the actual branch compiler diagnostic",
    "failure": "a-check failed for 1 file:\nreview/compiler/chain-followers/host-fixtures/nonvoid-path.a: refused, expected refused deliberately-unmatched-chain-followers-reason (adamic: /workspace/adamic/review/compiler/chain-followers/host-fixtures/nonvoid-path.a:3:20: Adamic 0.1 refuses a primitive brand member __pathBrand whose type is not void; make __pathBrand void (or o)",
    "result": {
      "a_check": {
        "review/compiler/chain-followers/host-fixtures/nonvoid-path.a": {
          "outcome": "refused",
          "expected": "refused deliberately-unmatched-chain-followers-reason",
          "first": "adamic: /workspace/adamic/review/compiler/chain-followers/host-fixtures/nonvoid-path.a:3:20: Adamic 0.1 refuses a primitive brand member __pathBrand whose type is not void; make __pathBrand void (or optional and typed undefined) so the brand is phantom"
        }
      }
    }
  }
]
```

Measured leaf count: 10084; maximum observed leaf seconds: 18.3.
