Replay of 611f6e09343d541c3e03ca5b63f74f6bf4726297 onto checked views, task #ht2nwj5.
Commit: 077830b7cdbcff1c9193886bd5eae279eb6faab5, followed by focused proof and counts.
Checks: five supported fixtures match Node in both backends; the computed fixture remains NotYet.
Mutants: optional receiver, required read, throw propagation and exception edge all caught.
Limits: no full gate; table-wide count refresh is blocked; area-views-next is unpublished.

Base: 7302410ecfe4d54c6321fde785f67bf39fc5dcc4, origin/compiler/area-views-wip3.
Only the original fix's one own commit was replayed. No merge commits were replayed.
Only codex/optional-chain-after-call-views is a delivery target.
Repeated `git ls-remote --heads origin compiler/area-views-next` returned no branch.
The requested rebase remains dependent on that branch being published.

The lowering keeps acceptsUndefined, contextual receiving slots, array-read absence,
parenthesized observations and computed required-read messages from the base. It
adds optionalReceiver for explicit optional property, element and call receivers.
Both backends throw catchable TypeErrors on required reference reads. The flow
and lowering exception analyses preserve their host, array-hole and phantom-member
guards while adding Defined's catchable failure. Native reuse keeps its checked-view
field contracts, initialization state, graph gates and callback conventions.
No protected emitter, lowering orchestration or oracle harness file was edited.
No cohere source was copied or modified.

Source Node observations, all exit 0 with empty stderr:

| Fixture | stdout |
|---|---|
| optional_after_call_79.a | empty |
| optional_after_call_193.a | empty |
| optional_after_call_variants.a | absent, absent, true, true, each on its own line |
| optional_after_call_required.a | TypeError: Cannot read properties of undefined (reading 'next') |
| optional_after_call_propagation.a | TypeError: Cannot read properties of undefined (reading 'label'); finished; TypeError kept 3, each on its own line |
| optional_after_call_gaps/computed.a | -1 |

The five supported fixtures run source Node, backend Node, release native,
ASan/UBSan native and the separate LeakSanitizer check. The computed probe is
pinned to NotYet at 6:16, `?.[] on a value`. This replay does not build optional
array indexing. It cannot claim that probe matches in the backends.

Validation commands, each redirected to the named evidence log:

```
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/oracle -run 'TestOptionalWriteErasure|TestViewPreflightPreservesOtherRefusals|TestCheckedViewDictionaryOptionalShortCircuitMutant|TestNativeAgreesWithNode/internal/oracle/testdata/optional_after_call' -count=1 -v -timeout 30m
ADAMIC_GATE_UNCACHED=1 go test ./internal/flow ./internal/oracle -run 'TestDefinedExceptionEdges|TestArrayHolesExceptionEdges|TestNarrowedFieldUsesSharedReadiness|TestOptionalAfterCallCounts|TestNativeAgreesWithNode/internal/oracle/testdata/(optional_after_call|047cb0d_n_|e4eec87_f1_)' -count=1 -v -timeout 30m
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(optional_after_call|047cb0d_n_|e4eec87_f1_|narrowed_|reuse_narrowed|regexp_null_narrowed)' -count=1 -v -timeout 30m
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/load/testdata/0.1/compile/09_tree' -count=1 -v -timeout 30m
python3 internal/oracle/optional_after_call_mutants.py
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts
go test ./internal/oracle -run TestOptionalAfterCallCounts -count=1 -v -timeout 30m -args -update-counts
go test ./internal/oracle -run TestOptionalAfterCallCounts -count=1 -timeout 30m
git diff --check
```

All focused unmutated commands pass. The final fixture selector runs 34 cases:
33 supported comparisons and the computed NotYet, followed by the tree comparison.
The regression command passes flow in 0.014s and oracle in 5.929s; the final
fixture comparison passes in 4.444s, the tree in 0.587s and the final count check
in 0.931s. A separate combined final selector accidentally selected no source
fixture cases because it grouped across subtest path separators. Its package
controls passed; the explicit fixture and tree commands above supply the proof.
No whole package test or full gate was run.

The table-wide refresh fails with 97 fixture failures. Its first failure is
library_method_values.a:26:41, detached Object.prototype.hasOwnProperty. Other
failures include slotless Map callbacks, overloaded argument counts and graph
cycle refusals. The base's own merge-judgment document records its counts and
validation as red or incomplete. These mechanisms are untouched by this replay.
The scoped count test reuses counted(), measures all 30 relevant rows and changes
only this unit and its existing narrowing controls. It preserves the spelling of
legacy rows without graph counters when both measured graph counters are zero.
It does not suppress the table-wide failure.

Every changed count row follows. Values are allocations/frees/retains/releases/
peak/in-regions, with graph regions/merges on new rows. The ten changed existing
rows are nine failed required reads, each adding one Error allocation, three
retains and two releases, plus the tree. Error values remain live when an uncaught
read stops, which is counted termination, not a completed-run leak. The tree's
exception successors inhibit unsafe overwrite reuse: +14 allocations and frees,
+38 retains, unchanged releases and peak. Five new rows measure the new fixtures.
The other fifteen measured existing rows keep every count. No unrelated row is
refreshed; predicate direction counts are unchanged.

| Fixture | Base | Measured |
|---|---|---|
| `Fixture` | Call sites/Proven/Checked/Unobservable | Allocations/Frees/Retains/Releases/Peak live/In regions |
| `internal/oracle/testdata/string_views_lifetime.a` | 19/19/21/34/6/0/0/0 | 19/19/21/34/6/0 |
| `internal/oracle/testdata/string_views_holders.a` | 45/45/72/79/13/0/0/0 | 45/45/72/79/13/0 |
| `internal/oracle/testdata/string_views_throw.a` | 55/55/62/76/14/0/0/0 | 55/55/62/76/14/0 |
| `internal/oracle/testdata/string_views_loops.a` | 159/159/45/167/21/0/0/0 | 159/159/45/167/21/0 |
| `internal/oracle/testdata/string_views_policy.a` | 344/344/73/371/11/0/0/0 | 344/344/73/371/11/0 |
| `internal/oracle/testdata/string_views_methods.a` | 5253/5253/2022/6546/15/0/0/0 | 5253/5253/2022/6546/15/0 |
| `internal/oracle/testdata/string_views_characters.a` | 281/281/1542/1825/8/0/0/0 | 281/281/1542/1825/8/0 |
| `internal/oracle/testdata/string_views_calls.a` | 85/85/61/116/11/0/0/0 | 85/85/61/116/11/0 |
| `internal/oracle/testdata/string_views_surrogates.a` | 188/188/83/255/13/0/0/0 | 188/188/83/255/13/0 |
| `internal/load/testdata/0.1/compile/09_tree.ts` | 19/19/28/69/16/0 | 33/33/66/69/16/0 |
| `internal/oracle/testdata/narrowed_reads.a` | 5/5/3/7/4/0 | 6/5/6/9/4/0 |
| `internal/oracle/testdata/narrowed_methods.a` | 3/3/3/6/3/0 | 4/3/6/8/3/0 |
| `internal/oracle/testdata/narrowed_fields.a` | 4/3/4/6/4/0 | 5/3/7/8/4/0 |
| `internal/oracle/testdata/reuse_narrowed.a` | 7/5/6/8/5/0 | 8/5/9/10/5/0 |
| `internal/oracle/testdata/regexp_null_narrowed.a` | 4/4/8/5/3/0 | 5/4/11/7/3/0 |
| `internal/oracle/testdata/census_never_rest_marker.a` | 10/10/14/34/8/0 | 10/10/12/31/8/0 |
| `internal/oracle/testdata/census_small_rest.a` | 20/20/35/42/8/0 | 16/16/31/38/6/0 |
| `internal/oracle/testdata/census_unary_numeric.a` | 187/187/178/349/14/0 | 190/190/175/349/14/0 |
| `internal/oracle/testdata/047cb0d_n_element_plain.a` | 4/3/5/8/4/0 | 5/3/8/10/4/0 |
| `internal/oracle/testdata/047cb0d_n_element.a` | 4/3/6/8/4/0 | 5/3/9/10/4/0 |
| `internal/oracle/testdata/047cb0d_n_element_method.a` | 4/3/6/9/4/0 | 5/3/9/11/4/0 |
| `internal/oracle/testdata/047cb0d_n_arrayindex.a` | 1/1/1/2/1/0 | 2/1/4/4/1/0 |
| `internal/oracle/testdata/nested_minimal.a` | 2/2/0/2/1/0/0/0 | 2/2/0/2/1/0 |
| `internal/oracle/testdata/nested_captures.a` | 22/22/8/22/7/0/0/0 | 22/22/8/22/7/0 |
| `internal/oracle/testdata/nested_hoisting.a` | 2/2/0/2/1/0/0/0 | 2/2/0/2/1/0 |
| `internal/oracle/testdata/nested_mutual.a` | 10/10/46/52/5/0/0/0 | 10/10/46/52/5/0 |
| `internal/oracle/testdata/nested_returned.a` | 21/21/10/29/11/0/0/0 | 21/21/10/29/11/0 |
| `internal/oracle/testdata/nested_array.a` | 17/17/17/27/9/0/0/0 | 17/17/17/27/9/0 |
| `internal/oracle/testdata/nested_three_levels.a` | 8/8/7/13/7/0/0/0 | 8/8/7/13/7/0 |
| `internal/oracle/testdata/nested_tdz.a` | 2/0/1/0/2/0/0/0 | 2/0/1/0/2/0 |
| `internal/oracle/testdata/nested_tdz_write.a` | 2/0/1/0/2/0/0/0 | 2/0/1/0/2/0 |
| `internal/oracle/testdata/nested_weak.a` | 4/4/5/10/4/0/0/0 | 4/4/5/10/4/0 |
| `internal/oracle/testdata/nested_destructured.a` | 13/13/7/17/6/0/0/0 | 13/13/7/17/6/0 |
| `internal/oracle/testdata/nested_destructured_tdz.a` | 2/0/1/0/2/0/0/0 | 2/0/1/0/2/0 |
| `internal/oracle/testdata/nested_mixed.a` | 7/7/9/12/5/0/0/0 | 7/7/9/12/5/0 |
| `internal/oracle/testdata/nested_pattern_parameter.a` | 6/6/5/10/5/0/0/0 | 6/6/5/10/5/0 |
| `internal/oracle/testdata/node_fs_file_close.a` | 27/27/11/27/7/0 | 27/27/13/30/7/0 |
| `internal/oracle/testdata/node_fs_file_write_file.a` | 30/30/7/31/8/0 | 31/31/11/36/9/0 |
| `internal/oracle/testdata/node_fs_file_mkdir.a` | 59/59/19/72/8/0 | 61/61/24/81/10/0 |
| `internal/oracle/testdata/node_fs_file_write_buffer.a` | 31/31/8/37/9/0 | 34/34/11/45/12/0 |
| `internal/oracle/testdata/optional_after_call_79.a` | new | 4/4/6/10/4/0/0/0 |
| `internal/oracle/testdata/optional_after_call_193.a` | new | 3/3/6/9/3/0/0/0 |
| `internal/oracle/testdata/optional_after_call_variants.a` | new | 11/11/10/21/3/0/0/0 |
| `internal/oracle/testdata/optional_after_call_required.a` | new | 3/3/5/7/2/0/0/0 |
| `internal/oracle/testdata/optional_after_call_propagation.a` | new | 18/18/13/20/7/0/0/0 |

Mutant observations:

- throwing-optional-receiver restores the old comparedWithUndefined-only bypass,
  removing both equivalent optional exemptions now present on the views base.
  Seed 79: source Node exit 0, native and backend exit 70, caught by Node comparison.
- required-read-panics makes Defined.Throws return false. The required fixture's
  source catches the TypeError and exits 0; both backends panic with exit 70.
- missing-throw-propagation drops Defined from the lowering throw summary. The
  propagation fixture disagrees with Node and reports a leaked Error under ASan.
- missing-exception-edge drops Defined's exceptional flow successor.
  TestDefinedExceptionEdges fails: catch successor false, expected true.
  Initially this mutant survived the propagation source fixture. The added flow
  test pins the liveness contract; this is an assertion catcher, not a sanitizer
  claim. Required and optional phantom cases also pin the base guard in that test.

All mutants compile and run. No warning or compiler error is credited as a catcher.
The runner restores each exact source file in finally. Its full logs are in evidence/.

The initial setup cache warm failed because it compiled unresolved cherry-pick
markers in flow/build.go. The repeat succeeded: Go ready 0.021s, Node ready
0.023s, markdown dependencies ready 0.071s, submodules ready 0.072s, clang ready
0.184s, build cache warm 46.806s, done 46.868s. nproc is 5; cgroup quota is four
cores. Go 1.27.1, Node 24.19.0 and clang 20.1.8 were used.

The final post-mutant delivery run repeats the explicit optional fixture selector,
all selected base guards, flow edge checks and all 30 scoped count checks. It passes
lower, flow and oracle; oracle reports 6.308s. See optional-views-delivery.log.gz.
