Rebased wave 10 onto current main; all six claimed rules now pass native byte-oracle validation.
Base main: e8ba3d5d; rebased production tip: 31427043; original pushed tip was 5d491682.
Native controls, both frozen corpora, sanitizers, touched-package checks, vet and the filtered uncached Node oracle pass.
All six rule mutants exit cleanly and fail byte comparison; released-registry and Node one-byte mutants are caught too.
No full repository gate, complete upstream fixture matrix or shared emitted-JavaScript lint comparison is claimed.

The branch remains codex/typeaware-wave-10. Rebase completed without conflicts.
No main or area branch is pushed. The earlier internal/fresh compilation blocker
is closed on this main: process and blocking native builds both finish. No compiler
fix was made in this unit, and the exact upstream change responsible is not proven.
Historical source-only evidence remains valid for its original commit; this report
supersedes its incomplete status. Claims remain reserved through integration.

Setup printed Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s,
build cache warm 151s and done 151s. nproc is 5, cpu.max is 400000 100000,
with 17.6 GB. Every build shell sources /workspace/adamic-tools/env.sh.

The original three rule gate passed in 196.702s: 28 controls, 61 findings;
compiler77, 9 findings; repository287, 1 finding. Native and sanitized output
agree with Go. Loop, redundant constituent and includes mutants differ at
bytes 554, 4602 and 8498 respectively, with exit 0 and empty stderr.
The released member-parameters query exits 70; the registry-retention mutant
exits 0 and fails that required panic check.

Timeout passed in 54.129s: 37 controls, 15 findings; both frozen corpora have
zero findings. Normal and sanitized comparisons, the native decision mutant
and the released query/retaining-registry mutant checks pass.

The initial process gate passed in 132.338s over its 64 controls, normal and
sanitized. TestLandingNativeRulesAndMutants then passed in 172.161s over all
92 controls and both corpora. Process reports 77 findings (45,901 bytes) and
blocking reports 30 (24,200 bytes). Every finding span, message, fix and
suggestion field agrees byte for byte; neither supplies a fix or suggestion.
Each rule's compiler77 output is 5,318 bytes and repository287 output is 18,485
bytes, with zero findings, normal and sanitized. Sanitizer stderr is empty.
The native process mutant drops newly recorded writes, and the blocking mutant
disables ordered blocking. Each exits 0 with empty stderr; comparison catches
bytes 114 and 13,904. Source execution also reran successfully in 76.922s with
both source mutants. The new raw syntax/signature native probe prints Block
and 262144; its released syntax query exits 70 with the exact owned panic.

Touched bridge/runtime-fact package checks and vet pass. The filtered uncached
Node oracle passed in 49.951s, with 19 Node and 28 native cache misses and its
one-byte mutant. It covers selected maps, sorting, string index, Unicode,
functions and closure fixtures matched by the filter.

Native time against Go, three alternating complete-process observations per
row, including loading, parsing, checker work and output. Every timed stdout
is compared. These are slowdowns, not speedup claims; the shared worker load
was not controlled. No performance change was made in this landing unit.

| Rule and dataset | Native median seconds | Go median seconds | Native / Go |
| --- | ---: | ---: | ---: |
| process-controls | 3.823015 | 0.082875 | 46.130x |
| process-compiler | 1.762614 | 0.310595 | 5.675x |
| process-repository | 0.276697 | 0.130671 | 2.118x |
| blocking-controls | 5.875436 | 0.082978 | 70.807x |
| blocking-compiler | 2.200932 | 0.294137 | 7.483x |
| blocking-repository | 0.417093 | 0.156808 | 2.660x |

Commands (all stdout/stderr go to files):

```sh
bash cloud/setup.sh > /tmp/wave-10-landing-setup.log 2>&1
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE10_ARTIFACTS=/workspace/wave-10-landing-original \
ADAMIC_WAVE10_REPOSITORY_MANIFEST=/workspace/wave-10-repository.manifest \
ADAMIC_WAVE10_COMPILER_MANIFEST=/workspace/wave-10-compiler.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-10-typescript \
go test ./stage1/cohere/typeaware -run '^TestWave10AgreementAndMutants$' \
  -count=1 -timeout=10m -v > /tmp/wave-10-landing-original.log 2>&1
ADAMIC_WAVE10_NEXT_VALIDATE=1 \
ADAMIC_WAVE10_NEXT_ARTIFACTS=/workspace/wave-10-landing-timeout \
ADAMIC_WAVE10_REPOSITORY_MANIFEST=/workspace/wave-10-repository.manifest \
ADAMIC_WAVE10_COMPILER_MANIFEST=/workspace/wave-10-compiler.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-10-typescript \
go test ./stage1/cohere/typeaware/wave_10_next \
  -run '^TestTimeoutAgreementAndMutants$' -count=1 -timeout=10m -v \
  > /tmp/wave-10-landing-timeout.log 2>&1
ADAMIC_WAVE10_PROCESS_VALIDATE=1 \
ADAMIC_WAVE10_PROCESS_ARTIFACTS=/workspace/wave-10-landing-process \
go test ./stage1/cohere/typeaware/wave_10_next \
  -run '^TestProcessNativeAgreement$' -count=1 -timeout=3m -v \
  > /tmp/wave-10-landing-process.log 2>&1
ADAMIC_WAVE10_LANDING_ARTIFACTS=/workspace/wave-10-landing-process \
ADAMIC_WAVE10_LANDING_FIXTURES=/workspace/wave-10-process-validation \
ADAMIC_WAVE10_REPOSITORY_MANIFEST=/workspace/wave-10-repository.manifest \
ADAMIC_WAVE10_COMPILER_MANIFEST=/workspace/wave-10-compiler.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-10-typescript \
go test ./stage1/cohere/typeaware/wave_10_next \
  -run '^TestLandingNativeRulesAndMutants$' -count=1 -timeout=10m -v \
  > /tmp/wave-10-landing-native-rules.log 2>&1
```

Prepare the 92-control fixture directory using the process/fixture commands
in README.md. evidence/landing holds raw subprocess output compressed without
altering its bytes, complete test logs, benchmark samples and log hashes.
No shared harness, registration generator, protected compiler file or submodule
pin was edited. The only new executable source is the owned landing Go harness;
all native rule modules remain .a.
