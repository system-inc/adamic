# Reproduction on the recorded box

Only this directory is committed. All .go.txt files are measurement or evidence
sources, not production compiler inputs. Sources are pinned in REPORT.md.
The helper scripts use the recorded /workspace layout.

1. Source /workspace/adamic-tools/env.sh and prepare the tools, pinned API packages
   and upstream input as in the task setup. Prepare detached /workspace/v3-main
   and /workspace/v3-head at the recorded compiler commits. Their matching cohere
   dependency is 06af2e399c6c1fb93fd570b774027942747400b4.
2. From the report checkout, use its meter at the recorded report base:

```sh
python3 stage3/census/latent/make_overlay.py /workspace/v3-main /workspace/scratch/v3-main-meter
python3 stage3/census/views-v3-price/add-overlay.py /workspace/scratch/v3-main-meter /workspace/scratch/v3-main-count
python3 stage3/meter/entry_overlay.py /workspace/scratch/v3-main-meter /workspace/scratch/v3-main-entry-meter
python3 stage3/meter/entry_overlay.py /workspace/scratch/v3-main-count /workspace/scratch/v3-main-entry-count
```

Repeat for head. From each pinned compiler worktree, build its two binaries:

```sh
go build -buildvcs=false -overlay=/workspace/scratch/v3-main-entry-meter/overlay.json -o /workspace/scratch/v3-main-meter-bin ./stage3/census/latent/tool
go build -buildvcs=false -overlay=/workspace/scratch/v3-main-entry-count/overlay.json -o /workspace/scratch/v3-main-count-bin ./stage3/census/latent/tool
```

Repeat for head with its names. Direct all command output to log files.
Run prove-ir.py and prove-pointer.py from this directory to reproduce the known
lowered counts, both skip-kind mutants, shared-pointer proof and fixture meter
byte identity. fixture-test.py proves the source enumeration and its mutant.

3. Run each pinned tree's own apply.sh --build-product into a fresh directory,
   /workspace/v3-main-adapted-complete or /workspace/v3-head-adapted-complete,
   with ADAMIC_TYPESCRIPT_SOURCE pointing at the recorded upstream checkout.
   The source inventory is:

```sh
node stage3/census/views-v3-price/count.cjs /workspace/v3-main-adapted-complete /workspace/scratch/v3-source-bound
python3 stage3/census/views-v3-price/prepare-plan.py
```

prepare-plan.py requires both completed products, compares all src bytes and
freezes the resolved tsc closure. Source records include the read form, receiver
text, exact static type, location and readonly-subset flag. The seven-form
subtotal and the broader named-road inventory are distinct summary fields.
Neither reproduces checked-view admission.

4. run-piece.py accepts one or several exact relative source paths. Each binary
   run is capped at 23 minutes. leg-checkpoint.py commits and plainly pushes its
   raw output immediately. Once all four runs are complete, the piece runner
   proves raw meter identity on both trees and uses stage3/census/pair on the
   ordinary fields with adapted root paths canonicalized. The paired archive
   retains every raw stream and each leg's timing, replacing redundant leg
   archives in the current tree. Their individual commits remain in history.

run-batch.py selects a bounded group and resumes completed observations. It
must first be adjusted to use finer scheduling for the stopped binder group;
blindly rerunning that group repeats an unmodified meter baseline that failed
to fit the budget. Unit scheduling must be proven against the meter before
claiming it supplies missing counts. No unobserved unit or file is zero-filled.

5. update-report.py renders completed paired observations and the source
   inventory. audit-artifacts.py independently recounts raw nodes, unit status,
   both source inventories and all per-file/per-form tables, verifies meter byte
   identity, verifies every archive's inner and outer hashes, and checks the
   output file count and naming rules. The audit does not certify an unfinished
   closure total.
