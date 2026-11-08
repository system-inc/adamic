# Generic method view verification

Primary compiler head: `550acd97678cf5433a29649ded1110b6a92b4c08`; base: `ff41896cee93f832a4a8fa68c4c1822ddef5885e`.
Additional focused/mutant head: `3fb35fb6d3dd0aef2181d26bf924f0d9f0271168` (test-only o2 row).
Evidence branch is based on the additional head. No compiler implementation changes are included.
Author: kirkouimet <kirk@kirkouimet.com>.

| Step | Exit | Result |
|---|---:|---|
| Submodules | 0 | Recursive shallow initialization succeeded; also succeeded in separate base worktree. |
| 1 gofmt | 0 | Empty output. |
| 1 vet | 0 | Passed after creating missing `/tmp/adamic-gate`; initial exit 1 retained. |
| 2 focused, 550acd976 | 0 | Both tests; all 16 subtests passed. Initial environment failure exit 1 retained. |
| 3 mutant, 550acd976 | 1 | All 12 rejection subtests failed; four identical-type control subtests passed. |
| 3 restoration | 0 | Byte-for-byte comparison and Git diff both clean. |
| 4 whole lower | 1 | Four failing tests and default 10-minute timeout; full trace in lower.log. |
| 5 compiler builds | 0 / 0 | Head `/tmp/adamic`, base `/tmp/adamic-base`; separate base worktree. |
| 5 probe runner | 0 | All 14 sources, both revisions, native sanitized and JS. Individual exits in probes.md and probes/. |
| 6 manifest | 0 | 4,957 entries. |
| 6 admission initial / env retry | 1 / 1 | npm SELF_SIGNED_CERT_IN_CHAIN before classification. |
| 6 admission final | 0 | PASS, complete corpus, 4,957 programs, no omissions, zero newly admitted, three newly refused. |
| 7 oracle | not run | Optional; time used for admission and added o2 proof. |
| Added step 2, 3fb35fb6 | 0 | Both tests; all 18 subtests passed, including o2.a and o2.ts. |
| Added step 3 mutant | 1 | All 14 rejection subtests failed, including o2.a and o2.ts; four controls passed. |
| Added restoration | 0 | Byte-for-byte and Git diff clean. |

## Mutant red subtests

Under `TestAGenericMethodSeenThroughAnotherViewIsNotYet`, each of these fails for **both .a and .ts**:

- g1, T unused on both sides
- g4, T only in a callback's type
- g5, generic on the source side only
- g6, generic on the target side only
- g7, an interface seen as another
- one class's generic method under two type arguments
- o2, an overloaded generic view (added head only)

The g1/g4/g5/g6/g7/o2 mutants lower the bad view. The Box row instead gets the nominal-ancestry refusal, failing its required generic-method NotYet diagnostic. Exact names and messages are in mutant*.log and the failure indexes.

## Lower failures

Full messages: [lower-failures.log](lower-failures.log). Full run and timeout goroutines: [lower.log](lower.log).

- TestNodeWatchInheritedOnMutant: second host type stops at generic `on`, contrary to mutant expectation.
- TestNodeWatchInheritedOnStaysNotYet/view: generic `on` stop precedes expected node:events registration stop.
- TestNodeWatchTscShape: generic `on` view stops.
- TestProgramRegionCensusMembership: missing ADAMIC_TYPESCRIPT_SOURCE.
- Then `panic: test timed out after 10m0s`; no whole-package pass is claimed.

## Probes

[Full table](probes.md), [structured results](probe-results.json), and exact [sources](probe-sources/).

Both extensions have the same compiler outcomes. Head refuses g1/g4/g5/g6/g7 with generic `handle` NotYet; g2 prints `a` natively and in JS, matching source Node. **g3 differs from the requested expectation:** head now stops at generic `handle`, while base stops earlier at `a value of type T`. Base native g1/g4/g5/g6/g7 prints `woof`; base JS compiles but throws, exit 70 through oracle/node.mjs. Source .ts Node exits 1 for those bad calls; source .a uses oracle/node.mjs and exits 70. All raw stderr is retained.

## Admission

[New refusals with full messages and requested flags](new-refusals.md). All three are host_watch programs at the generic `on` view. None is an override refusal or a constrained-string handle view in this corpus.

The exact requested command first failed provisioning. An environment-variable CA retry also failed because toolchain.Env filters npm_config_cafile. The final retry used a temporary ~/.npmrc `cafile=/etc/ssl/certs/ca-certificates.crt`, with certificate verification enabled, and explicit original head SHA (the checkout had moved for o2). The prior npm configuration was restored and verified. No compiler/tool code was changed.

Final output: admission-delta-final.json and admission-delta-final.log. Initial output names are preserved. Final report: zero admitted; 13 Node agreements, 50 accepted negative witnesses, 37 ruled runtime stops; verdict pass; void false; no omitted programs.

All numbered run outputs and exit codes are saved under this directory. A `.exit` file holds each exit code; extraction records both archive/tar pipeline exits. The setup-state log precedes the requested checkout; checkout.log records the switch to 550acd976. The environment required `/workspace/adamic-tools/env.sh` to select Go and clang.
