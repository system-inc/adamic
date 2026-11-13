# Polling enum write removal: baobab

Adaptation commit: `0af92480c726a9aa715fd67c0957d2fd894779e6` on typescript/polling-no-enum-write.

PASS on adamic main 28a8ac1fbe11d397d625d1dcdab52213b80d72a1.
TypeScript 6.0.3: 050880ce59e30b356b686bd3144efe24f875ebc8.
Count read first: c7f2bc3d5d811387f053c2c7a392a424225702b5.
Linux x64, Node v24.19.0. No issue was filed.

## Owner and exact change

No previous adaptation pins the whole setCustomPollingValues owner. The retained
owner-pin-search.txt and earlier-sys-owners.txt cover the named helpers and
adaptations 40, 70 and 71. The only older helper-name matches are timer-closures
witness references. A new 99-correctness-polling-no-enum-write adapter pins the
complete owner before and after with LF text and SHA-256, following the recent
for-each-ancestor form. The one initialization caller ignores its void result.

[Exact before](setCustomPollingValues-before.ts.txt),
[exact after](setCustomPollingValues-after.ts.txt), and [diff](owner.diff).
The enum-mutating setCustomLevels call becomes
`!!getCustomLevels("TSC_WATCH_POLLINGINTERVAL")`. Its boolean is unchanged:
any nonempty interval variable still triggers both table rebuild decisions.
The now-dead setCustomLevels and setLevel are removed. Every other line of the
owner stays unchanged. The adaptation README accounts for every operation.

## Exits and scope

| Acceptance check | Exit and observation |
| --- | --- |
| Full main / changed apply | 0 / 0 |
| Whole-owner contract | 0; drift, missing, duplicate, half-owner, LF/CRLF and foreign bytes checked |
| Adapter replay | 0; files 0, edits 0, bytes and mtime preserved |
| Stock TypeScript 6.0.3, 79 compiler roots | 0 / 0; zero diagnostics before / after |
| Ordinary adamic types, same 79 roots | 0 / 0 |
| Before / after local and test harness builds | All four commands 0 |
| Source and tracked-file scope | 0; 747 source files and 81,500 tracked files compared, only sys.ts changed |
| Emitted typescript.d.ts and API reference | 0; byte-identical before / after, hashes and common bytes retained |
| Full-source enum-write census | 0; 51 readers, 36 class (a), 15 class (b), zero writes |
| Restoring mutant | 1, expected; original levels[level] assignment found at sys.ts:176 |
| Actual-source CommonJS witness, both environments | All four child processes 0; behavior assertions passed |
| Published typescript.js and shipped _tsc.js controls | All four processes 0 |
| Plain-table / return-value witness | 0; 64 combinations of empty, zero, NaN and positive numeric inputs |
| All 398 manifest shards | All child exits 0; 106,366 passing, one expected API failure, zero skips |
| Final artifact and actual-file audit | 0 |
| Raw archive | 0; 2,388 files, per-member checksums retained |

The longest shard took 57.494 seconds. Every shard ran in its own
bounded process, with its actual child exit recorded. No lane, manifest, expected
count, reader or API-sanction file changed. The only failed test is exactly
`unittests:: Public APIs for typescript.d.ts should be acknowledged when they change`;
the only baseline diff is api/typescript.d.ts, accepted by the unchanged guard.

Stock checking uses the same existing pinned Node library inputs as Adamic:
@types/node 25.3.3, undici-types 7.18.2 and the existing host_axioms.json rows.
prepare-node.py records that input preparation. No checker option was relaxed,
no diagnostic was filtered, and neither checker nor upstream source was changed
for the stock check. See ATTEMPTS.md for the initial omitted-library failures;
those original exits and logs remain separate from final-checks.json.

## Node behavior

The count's second reference mode loads the actual adapted source modules with
stock transpileModule to CommonJS, without bundle enum inlining. It uses the
real initialized sys/getNodeSystem and intercepts timer registration.

| Environment / implementation | Scheduled intervals | Default callers 250/500/2000 | Configured callers 17/41/89 |
| --- | --- | --- | --- |
| DynamicPriorityPolling + LOW/MEDIUM/HIGH=17/41/89, before source | 17/41/89 | Throw | Succeed |
| Same, after source | 250/500/2000 | Succeed | Throw |
| Same, published bundle and shipped _tsc.js | 250/500/2000 | Succeed | Throw |

For shipped _tsc.js, only its final CLI dispatch is replaced with exports to
inspect its real initialized system; the watch implementation is unchanged.
The published-package issue reproduction needs no such extraction. With all
TSC_WATCH variables cleared, the before and after source outputs are byte-identical.
Raw outputs for each mode and environment are beside this report. These are
schedule-registration witnesses, not a claim about native execution or later
filesystem events. Table values, allocation/retention decisions, environment
read order and undefined return results are checked separately by tables.cjs.

## Native limitation and upstream draft

Tsc's sys.ts does not lower yet, so a native run of this code is not possible
today. No Node observation is presented as native evidence. PROOF.md in the
adaptation states the same limitation.

UPSTREAM-ISSUE.md is a factual, unfiled microsoft/TypeScript issue draft. Its exact
minimal script was executed against the published package; issue-repro.log retains
the output. The upstream ledger labels this JavaScript-changing and records both
upstream behaviors, the adopted default scheduling and Adamic's road (1) ruling.

## Evidence and reproduction

provenance.json identifies the base, count, upstream pin, census hash and WIP.
The original census.cjs is byte-for-byte c7f2bc3d's logic. The assertion wrapper
forwards a failed child exit, otherwise rejects any enum write. The restoring
mutant restores only the pinned original function in a disposable source copy.

Apply logs, checker outputs, four compiler/harness build logs, declaration hashes,
source hashes, census output, mutant output and runtime outputs are retained.
shard-exits.json and shard-logs contain every process verdict; the compressed raw
archive retains observations, test logs, baseline diffs and reports for all shards.
final-audit.py checks those artifacts and actual source/declaration bytes. See the
retained scripts for exact commands and 1700-second child limits. Their paths name
this workspace; use the same before/after checkout layout or adjust the paths when
reproducing elsewhere. Run the adapter from the adaptation branch and evidence
scripts from the evidence worktree, with the locked stage3/api dependencies installed.

The early WIP was pushed after full apply and both checkers passed, before the
remaining census, scope and lane gates. Its SHA is in wip-sha.txt. No force push,
main push, upstream issue submission or PR creation was performed.
