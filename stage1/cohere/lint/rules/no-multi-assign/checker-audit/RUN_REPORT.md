# Wave 12 checker migration result

Branch: `codex/lint-checker-wave-12`, based on area `c4bdc23fa86d55cf7e579989201c11258f4d3a62`. This is a blocker-and-evidence branch. No claimed rule has been registered or reported as landing-ready.

All eighteen claimed rules stopped on the exact shared dependency listed in `no-redeclare/CHECKER_AUDIT.md`; each note includes a raw source reproducer. New claimed-rule upstream matches: zero per rule. New claimed-rule mutants: none, because no complete implementation exists. The registered baseline's `TestRulesAgree` passed over 3,879 captured source/rule/options combinations; that result is not coverage of the blocked claims.

The previous pushed no-multi-assign tip is already in area. The previous wave branch was merged with area, without rebasing, at local merge `e1ca3a2bfc6e5bbe2a1566cc7175de850529205b`. Its old bridge extensions fail compilation against the new TypeScript pin; `evidence/wave12-checker-setup.log.gz` preserves exact files and lines. That merge was not pushed as green. No bridge extension or private checker/helper was carried into this area-based branch.

Only the owned `rules/no-multi-assign/checker-audit/` subtree is committed. Notes are nested there because the registry requires every direct `rules/` directory to contain a complete descriptor. The first notes-only layout failed this validation; the layout was corrected without changing the registry or registering incomplete rules. An initial corrected run was interrupted when the shell execution service disconnected; its incomplete log is retained. The final detached run completed with package failure.

The full lint command supplied every input:

```bash
source /workspace/adamic-tools/env.sh
GOMAXPROCS=4 GOFLAGS=-buildvcs=false \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-12/typescript \
ADAMIC_LINT_BENCH=1 \
ADAMIC_LINT_PROFILE_DIR=/workspace/wave-12/checker-profile \
ADAMIC_LINT_PROFILE_SNAPSHOTS=/workspace/wave-12/checker-profile \
go test -json -count=1 -timeout=60m ./stage1/cohere/lint > /tmp/wave12-checker-lint-complete.jsonl 2>&1
```

Completed test events, including subtests: 19 pass, 1 fail, 1 skip. Completed top-level tests: 18 pass, 1 fail, 1 skip. The package failed after timing out at `TestShardsAgree` after 3,600.294 seconds; outer wall time was 3627.985s. `TestCompilerAndStage1Agree` failed at its existing ten-minute native-command deadline; `CORPUS_FAILURE.md` and the preserved 872-file manifest reproduce it. No deadline, assertion or corpus selection was relaxed.

The sole skip is `TestCheckerBridgeRefusalPending`: awaits `codex/tsgo-errors-as-values`, because `tsgoInspect` must return `TSGoError` from the C error buffer. There were no missing-input skip events. Some tests never started before the package timeout, including `TestOwnedWitnesses`, `TestMutants` and profile checks; `evidence/gate-summary.json` enumerates these and paused tests without results. This branch is not green.

Toolchain setup passed on area in 328.780 seconds, with Go 1.27.1, clang 20.1.8 and Node 24.19.0. `nproc=5`, CPU quota four, memory 17.6 GB. Load before/after the final command: [1.02783203125, 5.5361328125, 6.34619140625] / [1.3984375, 1.2734375, 1.38134765625]. Registry generation, vet and formatting checks passed. The logs are under `evidence/`; benchmark timings are for the registered baseline only, not the blocked rules.

Stop conditions: missing shared identifier/global/alias/shorthand-symbol answers or reference tracker for fifteen claims; missing shared high-level IR/SSA/capture helpers for three React claims; old-wave bridge compilation failure; required native corpus deadline; full-package timeout. No new claims were taken.
