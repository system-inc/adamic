# getAllKeys proof checkpoint

Base: origin/main e8fc6350f0378eeadb866c4a41ef8eaf9638f8e9. Pinned TypeScript 6.0.3: 050880ce59e30b356b686bd3144efe24f875ebc8.

Setup follows 99-correctness-axolotl-w/PROOF.md: source /workspace/adamic-tools/env.sh, bash cloud/setup.sh, npm ci --prefix stage3/api, and export ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source. Setup and dependency installs exited zero. All 26 owner contract cases pass.

Both fresh full applies succeeded with child exit 0, including packaging: 94 baseline adapters and 95 adapted adapters. applies.json retains their commands and timings. Builds, paired replay, declarations, JavaScript, full lane pieces, and census/mutant checks remain pending at this checkpoint.

The first baseline attempt completed adaptation but failed product packaging because the build check started before tar finished. Its actual child exit was 1. It is excluded and retained separately. The successful full applies were rerun without overlapping builds; verify.py requires both full-apply exit records to be zero before starting checks.
