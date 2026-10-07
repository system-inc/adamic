Built: retained eight winning/unique rule ports and removed nine losing copies named by the dedup ledger.
Commits: this dedup commit follows pushed rule head 4a97dd6de; both owned branches retain main b8fb957a.
Commands: three fresh raw-witness/mutant suites PASS (eight witness rows); overlay vet PASS; JSX refusal control PASS.
Mutants: all eight retained semantic mutants compiled, exited cleanly and were caught only by Go comparison on source Node, emitted JavaScript and sanitized native.
Not covered: unified harness integration, two Nexus frontend/capture gaps, fresh broad corpus/performance and full repository gate; no new claim.

## Ledger decision

Read the complete DEDUP_LEDGER.md from origin/lint-rules/harness at 41eb6eab2b6de45ede0a40250765be295ee25fbd before deleting any owned copy. The pinned ledger is retained under evidence/parking/dedup-ledger.md. The harness is not an ancestor of current origin/main b8fb957aa839a9e8cb0b54279dd9864fa317bd30. This change removes only this unit's losing rule directories and updates its claim/evidence; it neither edits shared files nor imports the harness ahead of integration.

Retained rules: nexus/consistency-no-hand-rolled-delay, nexus/consistency-no-return-void, default-case-last, for-direction, guard-for-in, no-constructor-return, no-delete-var and no-eq-null. The five contested retained copies are the ledger winners; the other three are unique in its scope.

| Removed rule | Winning branch |
| --- | --- |
| @next/next/no-assign-module-variable | codex/lint-wave1-15 |
| @typescript-eslint/default-param-last | codex/lint-wave1-15 |
| @typescript-eslint/no-unnecessary-type-constraint | codex/lint-wave1-08 |
| @typescript-eslint/prefer-as-const | codex/lint-wave1-08 |
| @typescript-eslint/prefer-enum-initializers | codex/lint-wave1-08 |
| no-multi-str | codex/lint-wave1-15 |
| no-nonoctal-decimal-escape | codex/lint-wave1-15 |
| no-octal | codex/lint-wave1-15 |
| structure/tailwind-no-physical-direction | codex/lint-wave1-05 |

The ledger's 35 batch-only rows identify batch source branches, not a new assignment to slot 13. No batch-only rule is claimed in this change. Historical reports and logs below PARKING_REPORT.md continue to describe the former 17-rule branch; the current active scope is eight rules.

## Verification and limits

The scratch compatibility checkout uses the current main compiler through its existing symlinks. The same nine losing directories were deleted there before testing, so registration and oracle construction now discover only the eight retained owned ports. Their active .a modules, descriptors, mutants and Go adapters are checked byte-identical to this branch. No winning rule implementation changes.

Command for each group initial, loops and returns:

```
source /workspace/adamic-tools/env.sh
go test -overlay=/tmp/wave13-parking-c019-<group>/overlay.json ./stage1/cohere/lint -run '^(TestParkingWitnesses|TestWave13.*Mutants)$' -count=1 -v -timeout=20m > /tmp/wave13-dedup-<group>.log 2>&1
go vet -overlay=/tmp/wave13-parking-c019-initial/overlay.json ./... > /tmp/wave13-dedup-vet.log 2>&1
go test -overlay=/tmp/wave13-parking-c019-initial/overlay.json ./stage1/cohere/lint -run '^TestWave13JsxGap$' -count=1 -v -timeout=10m > /tmp/wave13-dedup-jsx-gap.log 2>&1
```

The JSX test is a refusal control, not byte-for-byte parity with Go. The two retained exclusions are the Go-valid top-level await delay input and the JSX return-void callback that the capture/parser path cannot faithfully replay. Removing other workers' copies withdraws the nine additional exclusions with those ports; it does not repair the Nexus gaps. These still exceed the parking exception's only-harness condition on this tested configuration. No new helper claim follows.

The existing helper branch remains at c88a20178 on the same main. Its unchanged three helpers passed 4,470 direct-helper cases, 820,676 Go bytes and twelve compiling mutants on this main in the prior current-main run. Those checks are narrower than full parity for all six consuming Tailwind rules; they remove eighteen dependency occurrences and zero final blockers. No helper implementation changes in this dedup pass.

## Fresh results

All three suites exited zero: initial 45.093s (two witness rows and two mutants), loops 56.967s (three rows and three mutants), returns 48.356s (three rows and three mutants). Vet exited zero with empty output. JSX refusal control passed in 13.536s while reproducing parser exit 70 on Node and sanitized native; this remains a gap. Full command output is retained as dedup-{initial,loops,returns,vet,jsx-gap}.log under evidence/parking.

Mutants caught: delay-resolve-name-ignored; return-void-braces-omitted; default-case-last-semantic-mutant; for-direction-semantic-mutant; guard-for-in-semantic-mutant; no-constructor-return-semantic-mutant; no-delete-var-semantic-mutant; no-eq-null-semantic-mutant. Each log records its mismatch independently on all three modes.

Checked 40 retained active module/descriptor/mutant/oracle files byte-identical between this branch and the scratch checkout after dedup. No source or rate measurement changed.
