# Morning rerun evidence, 2026-10-08

## Pins and source identity

[The table](TABLE.md) records the result. [pins.json](pins.json) freezes the fetched refs and their full SHAs at `2026-10-08T10:40:28Z`, 04:40:28 MDT. The mandatory runs completed at output mtimes 10:50:01 UTC and 10:52:13 UTC (04:50:01 and 04:52:13 MDT), on the morning's actual date. These are output modification times, not invented process timing measurements.

The report branch started at `57b9777c8eb4ee28b1f50220e8c8fb51a2dfadf7`. Its provenance tooling supplies the same measurement overlay to both compiler pins. No production compiler file is edited in the report branch. `source-manifest.json` was copied byte-for-byte from the existing 81-file table manifest and every byte count and SHA-256 was checked against `/tmp/stage3-notyet-adapted` before and after both runs. The source tree was not re-adapted.

```sh
git fetch origin '+refs/heads/area/compiler:refs/remotes/origin/area/compiler' '+refs/heads/codex/notyet-*:refs/remotes/origin/codex/notyet-*' '+refs/heads/codex/non-null-checked-area:refs/remotes/origin/codex/non-null-checked-area' '+refs/heads/codex/stage3-notyet-table:refs/remotes/origin/codex/stage3-notyet-table' > /tmp/stage3-0730-fetch.log 2>&1
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/stage3-0730-setup.log 2>&1
source /workspace/adamic-tools/env.sh
nproc
```

Fetch and setup exited 0; `nproc` = 5. [Fetch log](fetch.log.txt), [setup log](setup.log.txt). Setup's cumulative timing lines:

```text
setup: node ready (0.031s)
setup: go ready (0.032s)
setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.007s
setup: markdown dependencies ready (0.088s)
setup: submodules ready (0.093s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.199s)
setup: go build ready (40.651s)
setup: test binaries deferred (use --warm-tests) (40.774s)
setup: build cache warm (40.775s)
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (40.810s)
```

Go 1.27.1, clang 20.1.8, Node v24.19.0. The restarted managed cloud runtime was inspected using the cloud-environment-runtime skill; it was ready with HTTP access available through its configured policy. No credentials were printed or changed.

## Mandatory scratch integration

```sh
git worktree add -b scratch/stage3-0730-before /tmp/stage3-0730-before b68b2fe1af2dd6ad369ad8dfbfedf66be5a42861
git worktree add -b scratch/stage3-0730-after /tmp/stage3-0730-after b68b2fe1af2dd6ad369ad8dfbfedf66be5a42861
git -C /tmp/stage3-0730-after merge --no-edit --no-ff 1db2b32429c7edd8e4b6cfcd36ac98cea8df4a9a > /tmp/stage3-0730-after-statics-merge.log 2>&1
git -C /tmp/stage3-0730-after merge --no-edit --no-ff a8ad5bf0778fe8fd9f30b70bbe3eb807864e6b2a > /tmp/stage3-0730-after-binary-merge.log 2>&1
git -C /tmp/stage3-0730-after merge --no-edit --no-ff c41c0e062e99da37820f822968d4df1b48cdaee7 > /tmp/stage3-0730-after-non-null-merge.log 2>&1
```

All three exited 0. Statics merged cleanly to `015e5e8c280a2b9c2150c78e6680032678846829`; binary merged cleanly to `0e5661e4243af95ae3247d066de14ccb2b582a94`. Non-null printed `Already up to date.` It is already in the after ancestry, and is not in the before ancestry. [Statics merge](after-merge-statics.log.txt), [binary merge](after-merge-binary.log.txt), [non-null merge](after-merge-non-null.log.txt).

Before used the initialized cohere checkout by scratch symlink. After used shared detached cohere and nested TypeScript git worktrees at the exact gitlink pins, with no source copied from cohere. The first after build failed because the new cohere worktree had an uninitialized nested TypeScript directory. Adding the nested worktree at `d92d9bfee114c80be2c375d72edae966176e3a4f` fixed the missing `tsc/go.mod`; the retried build and census exited 0. The after build log preserves the original errors. A redundant attempt to add a worktree at TypeScript/tsc was rejected because that path is an ordinary tree already present; it changed nothing.

## Census commands and outputs

The following before command was run in `/tmp/stage3-0730-before`, after sourcing the toolchain and exporting GOPROXY:

```sh
python3 /workspace/adamic/stage3/census/latent/make_overlay.py /tmp/stage3-0730-before /tmp/stage3-0730-before-overlay > /tmp/stage3-0730-before-build.log 2>&1
python3 /workspace/adamic/stage3/meter/entry_overlay.py /tmp/stage3-0730-before-overlay /tmp/stage3-0730-before-entry >> /tmp/stage3-0730-before-build.log 2>&1
go build -buildvcs=false -overlay=/tmp/stage3-0730-before-entry/overlay.json -o /tmp/stage3-0730-before-census ./stage3/census/latent/tool >> /tmp/stage3-0730-before-build.log 2>&1
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 /tmp/stage3-0730-before-census /tmp/stage3-notyet-adapted/src/tsc/tsc.ts /tmp/stage3-0730-before.jsonl > /tmp/stage3-0730-before.log 2>&1
```

The same commands, with every `0730-before` replaced by `0730-after`, were run in `/tmp/stage3-0730-after`; the Go build was repeated after initializing its nested submodule. The frozen measurement tooling, source directory, entry file and flags were identical. Both final builds and censuses exited 0. Both outputs contain one metadata header and all 81 source records, with the exact same checker diagnostics, roots, source reach and measurement status. [Before build](before/build.log.txt), [after build](after/build.log.txt), [before run](before/census.log.txt), [after run](after/census.log.txt). `run.json` in each directory records production and submodule pins, overlay hashes, raw-output SHA-256, byte count and output mtime.

Raw output SHA-256:

```text
before eb3cea79d5c1e797705fc93aed8252dbbdf874fbe3ea3460f174657fdc8c2063 47863452 bytes
after  8c251bab65ddd5f44d6af90d44e690f2e58c528b5feec7e63cdf5fff672af650 44200986 bytes
```

## Accounting, independent audit and mutants

```sh
python3 stage3/notyet-table/rerun-0730/compare.py --manifest stage3/notyet-table/rerun-0730/source-manifest.json --source /tmp/stage3-notyet-adapted --before /tmp/stage3-0730-before.jsonl --after /tmp/stage3-0730-after.jsonl --output stage3/notyet-table/rerun-0730 > stage3/notyet-table/rerun-0730/comparison.log.txt 2>&1
python3 stage3/notyet-table/rerun-0730/audit.py /tmp/stage3-notyet-adapted > stage3/notyet-table/rerun-0730/audit.log.txt 2>&1
python3 stage3/notyet-table/rerun-0730/compare.py --manifest stage3/notyet-table/rerun-0730/source-manifest.json --source /tmp/stage3-notyet-adapted --before stage3/notyet-table/rerun-0730/before/full.jsonl.gz --after stage3/notyet-table/rerun-0730/after/full.jsonl.gz --output /tmp/stage3-0730-reproduction > stage3/notyet-table/rerun-0730/reproduction.log.txt 2>&1
```

All exited 0. The independently grouped SQL roots match every CSV signature and each reason total. Every regenerated CSV, root summary, unit-change ledger and comparison JSON matches the saved report byte-for-byte. Gzip decompression reproduces the raw hashes and byte counts. The report compares observed roots only; no worker report or credited-site figure is an input.

[Audit log](audit.log.txt), [reproduction log](reproduction.log.txt). Ten deliberate mutants were executed and caught:

| Mutant | Assertion that caught it |
| --- | --- |
| Swap retired and newly exposed | SQL membership flags |
| Drop one surviving signature | Complete union of observed SQL roots |
| Count a proven rollback echo as a root | SQL membership flags |
| Add one to net retirement | Net difference |
| Change one frozen source hash | Source hash |
| Drop one source result | Exact frozen source set |
| Remove a referenced governing root | Actual root exists in the same attempt |
| Remove a tagged read's declaration symbol | Declaration symbol provenance |
| Change checker diagnostics | Identical checker metadata |
| Change unit eligibility | Unchanged unit eligibility |

The observation-transition CSV additionally partitions the 581 after-only roots into 359 with no earlier NotYet at that location, 207 with a different earlier root, and 15 with a different earlier echo. None is the same signature merely promoted from echo-only to root, and no retired root becomes an exact echo-only signature. These subdivisions are observations, not inferred semantic completion.

## Poison provenance control on both builds

```sh
python3 stage3/census/latent/provenance_audit.py /tmp/stage3-0730-before-census stage3/notyet-table/rerun-0730/before-provenance > stage3/notyet-table/rerun-0730/before-provenance.log.txt 2>&1
python3 stage3/census/latent/provenance_audit.py /tmp/stage3-0730-after-census stage3/notyet-table/rerun-0730/after-provenance > stage3/notyet-table/rerun-0730/after-provenance.log.txt 2>&1
```

Both exited 0. On each binary, the positive witness preserves the failed initializer's actual `a function without a body` root across three ordinary/shorthand poison reads, and flattens the secondary binding to that root. Shadowed names and an independent parameter remain usable. Running `LATENT_MUTANT_DROP_BLOCKED_BY=1` makes the witness fail at `poison echo must carry blocked_by pointing at its actual root`. The audit requires that exact failure, not just any nonzero exit. [Before evidence](before-provenance.log.txt), [after evidence](after-provenance.log.txt). Witness source and complete positive/mutant records are saved beside the logs.

## No-output guards

The initial existing `audit_output_guards.py` harness passed a source directory. The entry driver requires an individual file; the harness failed on that input contract before reaching a guard. [Preserved first-attempt log](output-guard-first-attempt.log.txt). It did not affect either completed census.

`entry_guard_audit.py` retains the exact entry overlay and changes the witness input to a `.a` file. Its baseline witness measurement must first pass; then it builds two scratch-only mutants and requires each expected guard panic:

```sh
python3 stage3/notyet-table/rerun-0730/entry_guard_audit.py /tmp/stage3-0730-before /tmp/stage3-0730-before-guard-baseline /tmp/stage3-0730-before-census /tmp/stage3-0730-entry-guard-mutants > stage3/notyet-table/rerun-0730/output-guard-mutants.log.txt 2>&1
```

The baseline overlay directory consolidates the before compiler and entry replacement files using the exact before-entry `overlay.json` mapping, with only scratch paths relocated. Exit 0. [Guard audit log](output-guard-mutants.log.txt). Making production Lower return non-nil IR is caught by `measurement returned usable IR`. Making ordinary Load expose LatentLoad's rejected program is caught by `measurement loader exposed an output program`. Both mutant build and panic logs are preserved in `output-guards/`. Production compiler sources remain untouched.

## Optional integration and limits

The third run is not measured. All 26 pushed notyet refs were frozen, including the two mandatory tips. Sequential scratch integration reached predicates after 16 other tips; earlier conflicts and exact resolved files are listed in [all-merge-attempt.json](all-merge-attempt.json). Its predicates merge conflicted in 24 files across overlapping view-contract, closure, field-layout and backend changes. The unresolved merge was aborted and [optional-status.json](optional-status.json) records the conflicting files and partial head. No partial branch was measured or pushed. Full semantic integration of those implementations is outside this measurement unit; no third-run result is claimed.

Checker-diagnosed bodies, children behind failed compound boundaries, final ownership, module order and backend execution remain outside the census. Untagged reads are kept conservatively. No registered oracle fixture was added; therefore no fixture counts refresh, whole-package test or full gate was run. Verification consisted of the two census builds/runs, poison controls and drop-tag mutants, independent SQL accounting with ten mutants, byte-for-byte regeneration, output guard controls/mutants, manifest checks and git diff whitespace checks. Only the report branch is pushed; scratch integration branches are not pushed.
