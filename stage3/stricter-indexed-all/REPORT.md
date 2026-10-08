Integrated all four indexed slices on the newest stricter-options base; reconciled sparse-array witnesses and attempted the exact whole compiler tree.
Commits: merge chain 49812121, 9aa2035f, d19185d6, a5ff080b, c6ba3399, 899cae18, 44ac98bb, f2d8c2d4, 976cf0f5; integration evidence b250367c and retained logs 19339ca6, on codex/stricter-indexed-all.
Commands: all six merged test packages and vet passed; 79 dated source hashes matched; strict native pipeline stopped at 171 checker diagnostics, production pipeline at 72.
Mutants: every supported slice panic erasure was caught by its pinned observation assertion; two report-classifier mutants were rejected; detailed outcomes are in evidence/mutants.log and evidence/classifier-mutants.log.
Not covered: no whole-program native artifact; all 99 original indexed native outcomes remain unmeasured; nine minimal rows refuse, and D069's absent optional-field context remains outside its proof.

## Whole-program result

The original whole program did **not** reach lowering or clang. This is the count that constrains the date, rather than extrapolating from minimal witnesses.

| Profile | Compiler roots | Checker diagnostics | Indexed checker errors | Original indexed native outcomes |
|---|---:|---:|---:|---|
| Literal project strict options enabled | 79 | 171 | 99 | all 99 unmeasured |
| Production project options plus Adamic option conversion | 79 | 72 | 0 | all 99 unmeasured |

The literal project profile enables noUncheckedIndexedAccess, exactOptionalPropertyTypes, useUnknownInCatchVariables and strictBindCallApply in an in-memory overlay of the owning compiler tsconfig. Its 171 diagnostics are exactly the ledger's 99 indexed, 67 optional and five catch rows. No source expression was rewritten. Every indexed ID and its measured diagnostic appears in [indexed-status.csv](evidence/indexed-status.csv), with full metadata in [indexed-status.json](evidence/indexed-status.json).

The production profile preserves the project's own settings. Adamic audits the stricter options and defers the 99 indexed diagnostics to runtime guards. It still rejects the 67 optional and five catch diagnostics, so none of those indexed sites can be measured in native output. Both raw results retain the full diagnostics and option attribution: [literal strict](evidence/whole-strict.json) and [production](evidence/whole-production.json). Two JSON option sites are also recorded by production attribution; these are not checker diagnostics.

Consequently this run establishes **99 indexed errors under literal strict settings**, and **99 whole-program indexed outcomes blocked by 72 earlier errors under production conversion**. It establishes neither 99 native successes nor a per-site native refusal count. Errors were not suppressed to manufacture a lower-stage result. The optional/catch branches were outside this integration's authorized branch list and were not merged.

## Exact tree and reproduction

The named checker-259 directory has LEDGER.md, not a README. Read the dated ledger at a1a16427a46149435e25f847c45ad136f1f1e55c. Its adaptation input is 3b25512566206bc603b93264e8d072c55e075d64, rather than this integration's current stage3 patches. Apply that revision's stage3/apply.sh and run npm ci. All 79 implementation source hashes match the dated ledger ([identity](evidence/source-identity.json), [expected hashes](evidence/ledger-source-hashes.json)). The options inventory's eightieth root is its virtual prelude declaration; production Load injects its own prelude, so the probe passes the 79 implementation roots, not that virtual filename.

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/stricter-indexed-all-setup.log 2>&1
source /workspace/adamic-tools/env.sh
git worktree add --detach /tmp/stricter-indexed-all-ledger-base 3b25512566206bc603b93264e8d072c55e075d64
bash /tmp/stricter-indexed-all-ledger-base/stage3/apply.sh /tmp/stricter-indexed-all-typescript > /tmp/stricter-indexed-all-apply.log 2>&1
npm ci --prefix /tmp/stricter-indexed-all-typescript --ignore-scripts --no-audit --no-fund > /tmp/stricter-indexed-all-npm.log 2>&1
go build -o /tmp/stricter-indexed-all-probe ./stage3/stricter-indexed-all > /tmp/stricter-indexed-all-probe-build.log 2>&1
/tmp/stricter-indexed-all-probe /tmp/stricter-indexed-all-typescript stage3/stricter-indexed-all/evidence/ledger-options.json project-strict /tmp/stricter-indexed-all-strict-native > /tmp/stricter-indexed-all-whole-strict.json 2> /tmp/stricter-indexed-all-whole-strict.log
/tmp/stricter-indexed-all-probe /tmp/stricter-indexed-all-typescript stage3/stricter-indexed-all/evidence/ledger-options.json production /tmp/stricter-indexed-all-production-native > /tmp/stricter-indexed-all-whole-production.json 2> /tmp/stricter-indexed-all-whole-production.log
python3 stage3/stricter-indexed-all/classify.py --tree /tmp/stricter-indexed-all-typescript --strict stage3/stricter-indexed-all/evidence/whole-strict.json --production stage3/stricter-indexed-all/evidence/whole-production.json > /tmp/stricter-indexed-all-classify.log 2>&1
```

Both probe commands exit 1, as expected from their recorded checker rejection. The probe calls native.Build only after successful loading and lowering. Initial harness construction mistakenly counted the virtual prelude among implementation roots and was corrected before these recorded runs. Setup passed: Node 0.017s, Go 0.019s, submodules 0.051s, markdown 0.056s, clang 0.132s, Go build 37.223s, warm 37.331s, done 37.354s; nproc=5, CPU quota four, Go 1.27.1, clang 20.1.8, Node 24.19.0. [Setup log](evidence/setup.log).

## Merged witness coverage

Base dfb82dabbf7b548de84a334ae5e72935bd08ce90 already implements bounded sparse-array presence. Git merges were clean; semantic conflicts were reconciled by inspecting the affected hunks: c D155/D156 now initialize one slot for the present run and leave a hole for the absent run, and d's hole variants follow the same runtime/explain/mutant assertions as its dense variants. No compiler implementation file was edited.

```sh
ADAMIC_GATE_UNCACHED=1 go test ./stage3/stricter-indexed-a ./stage3/stricter-indexed-b ./stage3/stricter-indexed-c ./stage3/stricter-indexed-d ./stage3/stricter-options ./stage3/stricter-indexed-all -count=1 -timeout 10m -v > /tmp/stricter-indexed-all-final-suites.log 2>&1
go vet ./stage3/stricter-indexed-all ./stage3/stricter-indexed-c ./stage3/stricter-indexed-d > /tmp/stricter-indexed-all-vet.log 2>&1
```

All pass: a 56.460s, b 59.668s, c 44.942s, d 72.540s, base options 18.317s, integration routing 0.694s; vet exits zero. [Complete merged log](evidence/merged-suites.log). Coverage is 90 minimal ledger rows proven, nine refused, zero missing manifests. D060/D061 share a program and other downstream rows can share an original read; 90 rows is not 90 original whole-program guard sites.

Each supported witness verifies Node present/undefined observations, backend JavaScript and release/sanitized native observations, exact named stderr and exit 70, and explain coverage. Each erase-panic mutant must compile and fail the pinned contract; successful compiler errors are never counted as mutant kills. [Every logged mutant outcome](evidence/mutants.log) records the supported rows and sparse variants. Existing slice a logs include its individual erasures; the merged test reruns them. The classifier additionally rejects a missing D220 diagnostic and a fabricated built-stage result ([log](evidence/classifier-mutants.log)).

| Minimal refused row | Observed reason |
|---|---|
| D037 | Type 'SourceFile \| undefined' is not assignable to type 'SourceFile' |
| D071 | destructuring anything but a tuple into [names] |
| D072 | destructuring anything but a tuple into [names] |
| D073 | destructuring anything but a tuple into [names] |
| D119 | Adamic 0.1 refuses an index signature; use a Map |
| D129 | Adamic 0.1 refuses an index signature; use a Map |
| D130 | Adamic 0.1 refuses an index signature; use a Map |
| D131 | Adamic 0.1 refuses an index signature; use a Map |
| D151 | stage 0 can't lower a value of type string \| null yet |

D037's optional return is an attribution issue, not an established indexed guard. D069 preinitializes its receiving optional field; slice a retains the separate field-write counterexample in gaps/optional-field.json. Minimal proofs preserve read shapes, not every enclosing compiler operation.

## D212 and D220 routing

Both are in src/compiler/transformers/generators.ts:

| Row | Function | Original receiver syntax |
|---|---|---|
| D212 | transformGenerators.hasImmediateContainingLabeledBlock | line 2445: blockStack![j]; downstream diagnostic line 2446 column 48 |
| D220 | transformGenerators.tryEnterOrLeaveBlock | line 2983: blockOffsets![blockIndex]; diagnostic column 57 |

With the receiver made a guaranteed array, each standalone exact-`!` reduction still refuses in lowering:

`Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case`

[Routing test and observations](evidence/routing.log) establish the redundant-assertion refusal, not a claim that the checker-blocked whole program reached it. Slice b's runtime witnesses erase the redundant receiver assertion to test the indexed guard itself. Route assertion support separately from indexed presence.

## Integration tips and remaining work

Merged and pushed after each merge: b 3ccc3268, c f5a2212c, d c019ca20 then d22099f0 and 9191d209, a b7214299 then 90342391, a17a300f and f5a63e24. Integration evidence b250367c follows f2d8c2d4; retained logs are committed as 19339ca6. The final documentation-only D151 investigation was merged as 976cf0f5; no source or witness changed, so the completed gate remains applicable. Origin was checked each batch and after the final suite. No codex/stricter-records branch was published at the final check. After validation, stricter-options-checks advanced to 7d70c572 via a large origin/main merge and enum-length sparse-array change; this integration retains its start-time newest base dfb82dab and the user-requested later a/c/d tips. The new base was observed, not merged or tested here. Newer optional-write tips were observed but not merged. Only codex/stricter-indexed-all was pushed.

Remaining: remove the 72 whole-program optional/catch checker blockers in their owning units, rerun this exact native pipeline, then classify original indexed sites from actual lowering/native results. Records, array destructuring, nullable string representation and redundant assertion support remain separately routed. No full repository gate was run; the exact six-package gate and three-package vet above were run. Minimal witness completion cannot establish the whole-program completion date.
