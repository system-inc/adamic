# Accepted require-result amendment

Implementation: `ad8ca5a3801d979ca8486b73f3b0ea3cbec0b7d8`, one commit on
`8e3a784aa5a601f57342c07b68f145fe56252c21`.
Branch: `typescript/require-result-1610-1` (pushed without force).
Task: #baobab item 9. No Loom submission.

| Proof | Result |
| --- | --- |
| Full single-file sys.ts census, CommonJSBundle | 185 -> 184; new 0, changed 0, gone 1 |
| Restoring mutant | Exact original bytes; all 185 main findings return |
| Stock tsc 6.0.3, 79 roots | Main 0 diagnostics; candidate 0; new 0 |
| adamic types, 79 roots | Exit 0 |
| Final whole apply | Exit 0 |
| Owner contract | Pass: LF, CRLF, drift/half-state/duplicate/mixed-ending refusals, no-write replay |
| Existing crypto source and JavaScript scope oracles | Pass, 747-file source inventory |
| Every emitted source JavaScript file | 604 files, 1208 comparisons, byte-identical |
| Upstream npm ci + hereby dts, each side | Pass |
| typescript.d.ts and tsserverlibrary.d.ts | Byte-identical; both diff files empty |
| Whole unfiltered stage3/lane | PASS: 106366 passing, 1 sanctioned API failure, 0 pending |
| Final source identity | Only sys.ts differs among 747 files; exact measured candidate bytes |
| Actual final-tree adapter replay | 0 writes, identical bytes and mtime |

## Complete census delta

New findings: none. Changed findings: none.

The sole gone finding is retained below in full, including its phase and attempt
owner. Raw completed streams and all original fields are in raw/ and delta.json.

```json
{
  "root_kind": "",
  "depth": 0,
  "measurement": "measured on a checker-rejected program",
  "unit": "/workspace/scratch/require-result-1610/main-adapted/src/compiler/sys.ts:1473:5",
  "phase": "refusal_scan",
  "kind": "Refused",
  "where": "/workspace/scratch/require-result-1610/main-adapted/src/compiler/sys.ts:1610:22",
  "reason": "a value of type (baseDir: string, moduleName: string) => { module: any; modulePath: string; error: undefined; } | { module: undefined; modulePath: undefined; error: unknown; } seen as (baseDir: string, moduleName: string) => ModuleImportResult, which can write string | undefined where string is read",
  "text": "/workspace/scratch/require-result-1610/main-adapted/src/compiler/sys.ts:1610:22: Adamic 0.1 refuses a value of type (baseDir: string, moduleName: string) => { module: any; modulePath: string; error: undefined; } | { module: undefined; modulePath: undefined; error: unknown; } seen as (baseDir: string, moduleName: string) => ModuleImportResult, which can write string | undefined where string is read; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)"
}
```

The annotated return's `module: any` produces no additional finding. The existing
host model for a runtime-computed module path models require as never (see
`docs/axioms/host-any.md:49`, `internal/lower/host_modelled.go` and
`hostModelledNeverMembers` in `internal/lower/any_relations.go`). No model, axiom,
compiler rule or library declaration changes in this amendment. The existing
refusal_scan finding at sys.ts:1612:77 (System as ModuleResolutionHost) remains
unchanged; this change only resolves item 9's arrow-result refusal.

The final lane's only failed test is
`unittests:: Public APIs for typescript.d.ts should be acknowledged when they change`,
with baseline_diffs exactly `["api/typescript.d.ts"]`. Its unchanged sanction
accepts 273 declarations. The oracle/test exit 1 is expected; the lane exits 0.

See commands.json, lane/report.json, lane/execution.json, lane/oracle/, the
empty declaration diffs, declarations.json, contract.log, crypto-scope.json,
final-source.json and javascript.json for receipts. The initial stock probe
without Node type resolution is explicitly invalid and superseded by
stock-valid.json (zero diagnostics on each side). The initial lane was stopped
before tests to rebase the historical crypto proof oracle; only lane/ is the
accepted, fresh whole run at the implementation commit.
