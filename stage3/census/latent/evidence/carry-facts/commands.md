# Commands: first-error and cross-file proofs

Source pin: `7e896ce4890d74568585b55973a3c982e7d2629a`, fetched main when this work began. Remote main later advanced to `50503bc6abc22da12e9f9a0caad28b7f4bb1b20d`; after the composition blocker, this stopped result was not rebased onto an unmeasured tip. Branch: `typescript/census-carry-facts-3`. The reviewed prior change was applied as working changes, not retained as ancestors:

```sh
git fetch origin main > /workspace/scratch/carry3-fetch.log 2>&1
git diff --binary 7e896ce4890d74568585b55973a3c982e7d2629a 9bb35aa48f4c28dbfb73ede04a5c2a1e53d39c70 -- stage3/census/latent > /workspace/scratch/carry3-prior.patch
git switch -c typescript/census-carry-facts-3 origin/main > /workspace/scratch/carry3-branch.log 2>&1
git apply --3way /workspace/scratch/carry3-prior.patch > /workspace/scratch/carry3-apply.log 2>&1
```

## Build the unchanged baseline and changed census

The existing detached worktree `/workspace/scratch/carry2-main` is exactly the same pin, with no carry changes; it shares the initialized Cohere checkout by symlink. The baseline uses its own main templates and generator. The first-error baseline has no measurement modifications beyond main's normal census overlay.

```sh
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
python3 /workspace/scratch/carry2-main/stage3/census/latent/make_overlay.py /workspace/scratch/carry2-main /workspace/scratch/carry3-baseline-overlay > /workspace/scratch/carry3-baseline-overlay.log 2>&1
gofmt -w /workspace/scratch/carry3-baseline-overlay/*.go
(cd /workspace/scratch/carry2-main && go build -buildvcs=false -overlay=/workspace/scratch/carry3-baseline-overlay/overlay.json -o /workspace/scratch/carry3-baseline-census ./stage3/census/latent/tool) > /workspace/scratch/carry3-baseline-build.log 2>&1
python3 stage3/census/latent/make_overlay.py "$PWD" /workspace/scratch/carry3-carried-overlay > /workspace/scratch/carry3-carried-overlay.log 2>&1
gofmt -w /workspace/scratch/carry3-carried-overlay/*.go
go build -buildvcs=false -overlay=/workspace/scratch/carry3-carried-overlay/overlay.json -o /workspace/scratch/carry3-carried-census ./stage3/census/latent/tool > /workspace/scratch/carry3-carried-build.log 2>&1
```

## Selection-only shim for full-mode baseline

Main accepts only one `LATENT_ONLY_FILE` name. To compare the same three-file attempted set in one process, its scratch overlay is copied to `/workspace/scratch/carry3-list-baseline-overlay`; overlay destinations are remapped to the copy. In the copied `internal_lower_latent_units.go`, add `strings` and replace only the exact-file skip with:

```go
if selected := os.Getenv("LATENT_ONLY_FILE"); selected != "" {
    matched := false
    for _, candidate := range strings.Split(selected, "\n") {
        if candidate == program.FileName(file) { matched = true; break }
    }
    if !matched { continue }
}
```

No memo, state initialization, registration, first-ask effects, or lowering logic is changed in that baseline. It is not used for the first-error proof or baseline single-file reruns.

```sh
gofmt -w /workspace/scratch/carry3-list-baseline-overlay/*.go
(cd /workspace/scratch/carry2-main && go build -buildvcs=false -overlay=/workspace/scratch/carry3-list-baseline-overlay/overlay.json -o /workspace/scratch/carry3-list-baseline-census ./stage3/census/latent/tool) > /workspace/scratch/carry3-list-baseline-build.log 2>&1
```

## Actual ten-run proof sequence

```sh
python3 /workspace/scratch/carry3-measure.py > /workspace/scratch/carry3-measure.log 2>&1
```

[measure-proofs.py](measure-proofs.py) is the exact runner. It verifies the 79 input hashes, clears inherited `LATENT_*` settings, and executes all census children sequentially, recording `wait4` peak RSS and per-unit timing JSONL. `/proc/PID/status` is sampled every 0.2 s; a run over 14 GiB is killed and its last completed unit is retained. No run crossed that threshold.

The first-error baseline and changed commands each run once, with **no file filter**:

```sh
LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TIMES=/workspace/scratch/carry3-runs/first-error-all79/baseline/latent-times.jsonl /workspace/scratch/carry3-baseline-census /workspace/scratch/tree/src/compiler /workspace/scratch/carry3-runs/first-error-all79/baseline/latent.jsonl > /workspace/scratch/carry3-runs/first-error-all79/baseline/run.log 2>&1
LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TIMES=/workspace/scratch/carry3-runs/first-error-all79/carried/latent-times.jsonl /workspace/scratch/carry3-carried-census /workspace/scratch/tree/src/compiler /workspace/scratch/carry3-runs/first-error-all79/carried/latent.jsonl > /workspace/scratch/carry3-runs/first-error-all79/carried/run.log 2>&1
```

For full mode the wrapper sets `LATENT_FULL=1` and `LATENT_ONLY_FILE` to three exact absolute names joined by newline, in one environment variable. Newline occurs in none of the measured paths. The baseline executable is `carry3-list-baseline-census`; the changed executable is `carry3-carried-census`. Outputs are `/workspace/scratch/carry3-runs/full-multi/{baseline,carried}/latent.jsonl`. It then runs binder, emitter and utilities separately, once per side, with the original plain baseline binary and changed binary. These outputs are `full-single-FILENAME/{baseline,carried}/latent.jsonl`.

Each baseline/changed pair is compared byte-for-byte immediately. All five pairs passed. The runner then compares the raw file records from the singles to the one-process record. It exited **3**, deliberately, on the unchanged baseline utilities record; this is the documented stopping condition, not a passed suite. Diagnostic assembly later retained raw singles-combined files without normalizing metadata. Both sides show the same nine census-wide generic-witness differences; [proofs.json](proofs.json) records exactly which comparisons passed and failed.

## Guard restoration, full tests and vet after the stopping point

```sh
python3 /workspace/scratch/carry3-run-mutants.py > /workspace/scratch/carry3-mutant-summary.log 2>&1
go test ./stage3/census/latent/statecopy ./stage3/census/latent/overlaytext -run '^TestLatentCarryFacts' -count=1 > /workspace/scratch/carry3-restored-green.log 2>&1
go test -count=1 ./stage3/census/... > /workspace/scratch/carry3-tests.log 2>&1
go vet ./stage3/census/latent/statecopy ./stage3/census/latent/overlaytext > /workspace/scratch/carry3-vet.log 2>&1
go vet -overlay=/workspace/scratch/carry3-carried-overlay/overlay.json ./stage3/census/latent/tool > /workspace/scratch/carry3-tool-vet.log 2>&1
```

The table mutant removes only the `entriesRecords` carrier and restores it in `finally`. Type and body mutants redirect only a scratch test-source overlay's read path to scratch copies of lower.go / entries_records.go. Production `internal/` files are never changed. For non-test overlay lowering vet, `go list -overlay=/workspace/scratch/carry3-carried-overlay/overlay.json -json ./internal/lower` supplies `Dir` and `GoFiles`; [overlay-source-vet-command.json](overlay-source-vet-command.json) retains the exact successful expanded argv. Production tests referencing the intentionally removed ordinary lowering entrypoint are not included in that overlay vet.

All earlier profile/answer data and the full-only proof are historical and explicitly labelled in the report. This work proves all-79-files identity in **first-error mode**, and three-file cross-file identity in **full mode**; it does not prove full mode over all 79 files or raw file-record independence from the selected attempted-unit set.
