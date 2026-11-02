# Commands and pins

The compiler source is pinned to `9c607b11c6d7a6afad29853db5d86b488baaa08e`, the main tip fetched for the original brief. This continuation starts from evidence commit `69b4b3bd7205c51dc5b94a1790c3b02d91b9a2b5`. It will be pushed as a new commit on that same branch, without rewriting history. The original setup and adapted-tree commands are preserved verbatim in [full-baseline-blocker.md](full-baseline-blocker.md). No later main changes were mixed into the measured compiler.

The TypeScript input pin is `050880ce59e30b356b686bd3144efe24f875ebc8`; [adapted-source-shas.csv](adapted-source-shas.csv) records the exact 79 adapted files. Cohere is `06af2e399c6c1fb93fd570b774027942747400b4`, with TypeScript submodule `d92d9bfee114c80be2c375d72edae966176e3a4f`. Go is 1.27.1, Node 24.19.0, clang 20.1.8. The measured binaries' SHA-256 values are in [binaries.sha256](binaries.sha256).

## Actual subset invocations

From `/workspace/adamic`, using the original unmodified overlay generator for the baseline:

```sh
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
python3 stage3/census/latent/make_overlay.py "$PWD" /workspace/scratch/baseline-overlay > /workspace/scratch/baseline-overlay.log 2>&1
gofmt -w /workspace/scratch/baseline-overlay/*.go
go build -buildvcs=false -overlay=/workspace/scratch/baseline-overlay/overlay.json -o /workspace/scratch/baseline-census ./stage3/census/latent/tool > /workspace/scratch/baseline-build.log 2>&1
python3 /workspace/scratch/run-subset.py baseline /workspace/scratch/baseline-census > /workspace/scratch/subset-baseline.log 2>&1
```

The Python wrapper sequentially invokes the following command for `utilities.ts`, `binder.ts`, then `emitter.ts`, setting `LATENT_TIMES` to a separate file for each invocation. It samples `/proc/PID/status` VmHWM/VmRSS every 0.5 s, kills the child if observed memory exceeds 12 GiB, and records monotonic wall time, peak observed RSS, return code, and threshold flag in each run's `resources.json`. No run crossed that threshold. Its directory layout is `/workspace/scratch/subset/MODE/FILENAME/`.

```sh
LATENT_ONLY_FILE=/workspace/scratch/tree/src/compiler/utilities.ts \
LATENT_TIMES=/workspace/scratch/subset/baseline/utilities.ts/latent-times.jsonl \
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 \
/workspace/scratch/baseline-census /workspace/scratch/tree/src/compiler \
/workspace/scratch/subset/baseline/utilities.ts/latent.jsonl \
> /workspace/scratch/subset/baseline/utilities.ts/run.log 2>&1
```

The profiling overlay was a scratch copy of the baseline overlay, with a CPU-profiler driver and direct whole-program walker timers. Scratch-only scripts `/workspace/scratch/instrument-profile.py`, `instrument-answers.py`, and `instrument-literals.py` added the call timers and per-unit answer probes. The fixed-point timer was also added in scratch. No profiling code is committed.

```sh
gofmt -w /workspace/scratch/profile-overlay/*.go
go build -buildvcs=false -overlay=/workspace/scratch/profile-overlay/overlay.json -o /workspace/scratch/profile-census ./stage3/census/latent/tool > /workspace/scratch/profile-build.log 2>&1
python3 /workspace/scratch/run-subset.py profile /workspace/scratch/profile-census > /workspace/scratch/subset-profile.log 2>&1
go tool pprof -top -cum -nodecount=35 /workspace/scratch/profile-census /workspace/scratch/subset/profile/utilities.ts/cpu.pprof > /workspace/scratch/utilities-profile-top.txt 2>&1
go tool pprof -peek 'enumNeverValue|entriesProgramRecordsWalk|entriesRecordLiteralWalk' -nodefraction=0 -edgefraction=0 /workspace/scratch/profile-census /workspace/scratch/subset/profile/utilities.ts/cpu.pprof > /workspace/scratch/utilities-profile-walk-chain.txt 2>&1
go tool pprof -traces -focus entriesProgramRecordsWalk /workspace/scratch/profile-census /workspace/scratch/subset/profile/utilities.ts/cpu.pprof > /workspace/scratch/utilities-record-traces.txt 2>&1

gofmt -w /workspace/scratch/literal-proof-overlay/*.go
go build -buildvcs=false -overlay=/workspace/scratch/literal-proof-overlay/overlay.json -o /workspace/scratch/literal-proof-census ./stage3/census/latent/tool > /workspace/scratch/literal-proof-build.log 2>&1
python3 /workspace/scratch/run-subset.py literal-proof /workspace/scratch/literal-proof-census > /workspace/scratch/subset-literal-proof.log 2>&1
```

Each profile produces its own `cpu.pprof`, `walks.json`, and `answers.jsonl`; the literal-proof invocations additionally produce `literals.jsonl`. First asks are grouped by exact unit/phase/memo (or unit/phase/literal); changed literal verdicts would be emitted again and rejected by the equality reduction. Results from separate processes are compared by stable source location for literal Booleans, never by pointer address. The Boolean-only intermediate overlay/run is in `/workspace/scratch/carry-overlay` and `subset/changed`; it also passed per-file byte identity, but is not the final performance comparison. The final source is built/run with:

```sh
python3 stage3/census/latent/make_overlay.py "$PWD" /workspace/scratch/final-carry-overlay > /workspace/scratch/final-carry-overlay.log 2>&1
gofmt -w /workspace/scratch/final-carry-overlay/*.go
go build -buildvcs=false -overlay=/workspace/scratch/final-carry-overlay/overlay.json -o /workspace/scratch/final-carry-census ./stage3/census/latent/tool > /workspace/scratch/final-carry-build.log 2>&1
python3 /workspace/scratch/run-subset.py carried /workspace/scratch/final-carry-census > /workspace/scratch/subset-carried.log 2>&1
```

Python compared every per-file JSONL byte sequence, stopping at the first differing line if any. There were no differing lines. It assembled each combined `latent.jsonl` by asserting the three checker headers were identical, retaining one header, then appending binder/emitter/utilities file records. SHA-256 values are in [latent.sha256](latent.sha256).

## Guard and checks

```sh
guard=TestLatentCarryFactsRequiresEveryLoweringFieldPolicy
git grep -n "func $guard"
go test ./stage3/census/latent/statecopy -run '^TestLatentCarryFactsRequiresEveryLoweringFieldPolicy$' -count=1 > /workspace/scratch/carry-final-guard-green.log 2>&1
# Remove only the entriesRecords policy from the table, run the same test to
# carry-final-guard-mutant.log, assert nonzero status and the named field,
# then restore the original file in finally. See carry-facts-mutants.md.
go test ./stage3/census/... > /workspace/scratch/carry-facts-tests.log 2>&1
go vet ./stage3/census/latent/statecopy > /workspace/scratch/carry-facts-vet.log 2>&1
go vet -overlay=/workspace/scratch/final-carry-overlay/overlay.json ./stage3/census/latent/tool > /workspace/scratch/carry-facts-tool-vet.log 2>&1
```

An extra `go vet -overlay=... ./internal/lower ./stage3/census/latent/tool` included incompatible production tests referencing an intentionally disabled entrypoint; its log is retained. To vet the actual overlay source, `go list -overlay=/workspace/scratch/final-carry-overlay/overlay.json -json ./internal/lower` supplied `Dir` and `GoFiles`; `go vet` was then called on those non-test files with the same overlay, successfully. The exact expanded argv is [overlay-source-vet-command.json](overlay-source-vet-command.json).

## Unfiltered proof on a house box with more memory

**Not run here.** On a house box, start in the checkout of the pushed `typescript/census-carry-facts` branch, with `cohere` initialized at the pins above, the same Go tools, and the identical adapted tree at `/workspace/scratch/tree`. This builds both compilers from their actual templates, uses the same absolute input paths, removes the file filter, preserves both complete ledgers, and compares all 79 files. The baseline worktree shares the initialized Cohere checkout only; it has its own baseline templates and source.

```sh
set -euo pipefail
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
repo=$PWD
house_run=$(mktemp -d /workspace/scratch/carry-facts-house.XXXXXX)
ADAPTED=/workspace/scratch/tree
python3 - "$ADAPTED" "$repo/stage3/census/latent/evidence/carry-facts/adapted-source-shas.csv" > "$house_run/input-check.log" 2>&1 <<'PY'
import csv, hashlib, pathlib, sys
root = pathlib.Path(sys.argv[1])
rows = list(csv.DictReader(open(sys.argv[2])))
assert len(rows) == 79
for row in rows:
    assert hashlib.sha256((root / row['path']).read_bytes()).hexdigest() == row['sha256'], row['path']
actual = {str(p.relative_to(root)) for p in (root/'src/compiler').rglob('*') if p.is_file() and p.suffix in ('.ts', '.a')}
assert actual == {row['path'] for row in rows}
PY
git worktree add --detach "$house_run/base" 9c607b11c6d7a6afad29853db5d86b488baaa08e > "$house_run/worktree.log" 2>&1
rmdir "$house_run/base/cohere"
ln -s "$repo/cohere" "$house_run/base/cohere"
python3 "$house_run/base/stage3/census/latent/make_overlay.py" "$house_run/base" "$house_run/baseline-overlay" > "$house_run/baseline-overlay.log" 2>&1
gofmt -w "$house_run/baseline-overlay/"*.go
(cd "$house_run/base" && go build -buildvcs=false -overlay="$house_run/baseline-overlay/overlay.json" -o "$house_run/baseline-census" ./stage3/census/latent/tool) > "$house_run/baseline-build.log" 2>&1
python3 stage3/census/latent/make_overlay.py "$repo" "$house_run/carried-overlay" > "$house_run/carried-overlay.log" 2>&1
gofmt -w "$house_run/carried-overlay/"*.go
go build -buildvcs=false -overlay="$house_run/carried-overlay/overlay.json" -o "$house_run/carried-census" ./stage3/census/latent/tool > "$house_run/carried-build.log" 2>&1
env -u LATENT_ONLY_FILE LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TIMES="$house_run/baseline-times.jsonl" "$house_run/baseline-census" "$ADAPTED/src/compiler" "$house_run/baseline-latent.jsonl" > "$house_run/baseline-run.log" 2>&1
env -u LATENT_ONLY_FILE LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TIMES="$house_run/carried-times.jsonl" "$house_run/carried-census" "$ADAPTED/src/compiler" "$house_run/carried-latent.jsonl" > "$house_run/carried-run.log" 2>&1
cmp "$house_run/baseline-latent.jsonl" "$house_run/carried-latent.jsonl" > "$house_run/compare.log" 2>&1
sha256sum "$house_run/baseline-latent.jsonl" "$house_run/carried-latent.jsonl" > "$house_run/latent.sha256"
printf '%s\n' "$house_run"
```

A nonzero comparison is a failed all-files proof: retain both files and report the first differing line. It must not be described as byte-identical based on the subset result.
