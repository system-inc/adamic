Built: one approved full-corpus measurement and an audited before/after join of all 2,936 sites; conformance remains 0% to 0%.
Commits: measured 81e1118e829101dd65ff4a4bec3add403c1173ea; four groups 6e2b7fee, 95935265, c4700f6b and 81e1118e; parent 81c20396.
Commands and outputs: one census exit 0; lossless roundtrip and both independent per-site audits PASS; 260 diagnostic texts unchanged.
Mutants: drop-corpus-site, duplicate-corpus-site and invent-aggregate-free caught by the independent audit; conflate-true-one and drop-array-item caught by canonical JSON hashes.
Not covered: no newly erased sites, no individual-rule attribution of Unknown transitions, no checker repairs or production IR, and no shared fast gate run.

The conforming share is **0/2,936 before and 0/2,936 after (0% to 0%)**. Free, readiness-only and conforms-if are each zero in both measurements. The tagged denominator is 1,758 and the untagged denominator is 1,178. The comparison derives counts from the audited per-site observations, joining exact original file/start/end locations and checking text and kind.

| Outcome | Before tagged | Before untagged | Before total | After tagged | After untagged | After total |
|---|---:|---:|---:|---:|---:|---:|
| Free | 0 | 0 | 0 | 0 | 0 | 0 |
| Readiness-only | 0 | 0 | 0 | 0 | 0 | 0 |
| Conforms-if | 0 | 0 | 0 | 0 | 0 | 0 |
| Unknown: unsupported flow | 192 | 170 | 362 | 65 | 148 | 213 |
| Unknown: diagnosed body or dependency | 1,566 | 1,005 | 2,571 | 1,693 | 1,027 | 2,720 |
| Unknown: host | 0 | 3 | 3 | 0 | 3 | 3 |
| Total | 1,758 | 1,178 | 2,936 | 1,758 | 1,178 | 2,936 |

Exactly 149 sites (127 tagged, 22 untagged) move from unsupported flow to diagnosed body/dependency. Every other primary bucket is unchanged. This is increased visibility of a blocking diagnostic, not proof of conformance. All 260 diagnostic texts and all 2,563 allocation schemas remain unchanged. Detail-message sets change at 2,634 sites and diagnostic-frontier counts at 2,487 sites; these changes do not erase a cast.

**There are no newly erased sites, so no newly erased site used any of the four rules.** [newly-erased-sites.json](step09/newly-erased-sites.json) is the empty list. The two conservative cause rules cannot themselves erase a cast. The combined measurement does not establish which individual rule caused each Unknown transition; no ablation census was run.

| Group | Rule | Newly erased sites |
|---|---|---:|
| 6e2b7fee | Named uninitialized declaration cause | 0 |
| 95935265 | Simple object binding property edge | 0 |
| c4700f6b | Named iteration producer frontier with collection dependencies | 0 |
| 81e1118e | Non-null callable operand reference classification | 0 |

[per-site-table.json](step09/per-site-table.json) contains every original location, expression, kind, before/after outcome and cause, detail-set hash, diagnostic-frontier count and erasure attribution. [changed-buckets.json](step09/changed-buckets.json) records all 149 transitions. [summary.json](step09/summary.json) records the independently recounted totals. [after-interned.json.gz](step09/after-interned.json.gz) preserves the complete new observations losslessly, including detailed provenance.

The saved before snapshot is 65535459b993b61bd7ce6464d2f4d768f78f9256, using overnight/missing-results-interned.json.gz. Its measured loader, graph, adapter and dependency inputs are unchanged through the four-group parent 81c20396903e9be350487ccec48f0bd0db2f5855; the checked path list and empty diff are recorded in [manifest.json](step09/manifest.json). Both measurements use the same 79 source hashes, same 2,936-site map, same cohere and TypeScript pins, and the fixed adaptation of area/stage3 234ab1aa5f728a5221fb6075c35b94f88a2c6437. The after binary and adapter hashes were verified against the pushed 81e1118e manifest before invocation. No new flow rule was introduced for this measurement.

The raw output SHA256 is 71df7417aa350b710a574b79521567090c539a6ec18d96fb128b942c0b6f60ce. The committed manifest gives sizes and SHA256 hashes for the packed output, tables, scripts and logs. The raw output remains at /workspace/shape-overnight-measurements/shape-step09-81e1118e.json.

Commands run, with output saved directly to the linked logs:

- `GOMAXPROCS=2 /tmp/shape-nonnull-callables-binary /tmp/shape-stage3-234ab1aa-adapted /tmp/shape-views-map.json /workspace/shape-overnight-measurements/shape-step09-81e1118e.json`: invoked once, exit 0; [census.log](step09/census.log).
- `python3 stage3/shape-conformance/dynamic-keys/artifacts.py pack RAW MAP ROOT stage3/shape-conformance/step09/after-interned.json.gz`: independent audit and complete observation roundtrip PASS; [pack.log](step09/pack.log).
- `python3 stage3/shape-conformance/step09/compare.py BEFORE AFTER MAP ROOT stage3/shape-conformance/step09`: both snapshot audits and exact per-site join PASS; [compare.log](step09/compare.log). BEFORE is the saved missing-results snapshot, AFTER the new packed snapshot; MAP and ROOT are the same census inputs above.
- `python3 stage3/shape-conformance/step09/audit-mutants.py AFTER MAP ROOT`: all three actual-corpus metadata mutants caught by the exact intended audit assertion; [audit-mutants.log](step09/audit-mutants.log). Dropping a site fails inventory cardinality, duplicating one fails location uniqueness, and inventing a Free aggregate fails independent recount.
- `python3 stage3/shape-conformance/dynamic-keys/artifacts.py --self-test`: both storage mutants caught and lossless scalar/Unicode/repeated-object roundtrip PASS; [storage-mutants.log](step09/storage-mutants.log).

The prior four groups' fixtures, 127 independent controls and 42 valid built flow/query mutant runs are recorded in [OVERNIGHT-REPORT.md](OVERNIGHT-REPORT.md). This unit adds measurement and audit evidence. The input remains checker-rejected, so these analysis outcomes do not certify production emission. The diagnosed bucket is left for the checker ledger. No shared fast gate or integration watcher was invoked or polled.
