# Readonly writeFile capture — #b8ffarm

Fresh base 8e3a784aa5a601f57342c07b68f145fe56252c21; fresh upstream v6.0.3 clone 050880ce59e30b356b686bd3144efe24f875ebc8; recursive shallow submodules and stage3/api npm ci complete. The main apply passed. Ownership audit requires a new whole-owner 99-correctness-write-file-view adaptation, with no public rows.

| Proof | Measured result |
|---|---|
| Contract | PASS: exact pins, drift/duplicate/overload/missing/moved/half/parse/mixed/bare-CR refusals; LF/CRLF; zero-write replay/restoring replay with byte/mtime equality |
| Witness | PASS: real pinned owner bodies; receiver, getter, later assignment, identity and own .call preserved |
| JavaScript | 604 files; 1,208 pairs / 2,416 emits; only owner alias/read redirect differs; all other files byte-identical |
| Full CommonJSBundle sys.ts census | 59 → 58; zero new, one gone, zero changed; complete phased delta in CENSUS.md |
| Restoring mutant | 59, exactly main again; original target returns |
| Stock tsc 6.0.3 | Actual CLI, 79 roots: main/candidate/mutant all exit 0 |
| adamic types | Actual production CLI, 79 roots: main/candidate/mutant all exit 0 |
| Upstream npm ci + hereby dts | Both builds pass; both public declaration files byte-identical; exact diffs empty |
| Carried lane count | 2,200 captures and 4,627 calls both sides; source SHA-256 matches the previously counted source exactly |
| Fresh candidate whole apply and lane | Next: full unfiltered stage3/lane/run.sh; no result claimed yet |

Only patchWriteFileEnsuringDirectory's local capture changes. No compiler, public API, sanction, oracle, worker count or expected count is changed. The adapter/README/PROOF/ledger bytes are frozen before the whole lane. The final landing branch will contain exactly one commit on the requested base, without these evidence products.
