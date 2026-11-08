# Verification at c9fa1ea

Head: `c9fa1ea7129678df8e3e01875830808ec87c9b4a`.
Base: `ff41896cee93f832a4a8fa68c4c1822ddef5885e`.
Isolated compiler worktree: `/tmp/adamic-b3`, switched from b3a5b76b to the exact new head; evidence worktree stayed on `evidence/compiler/generic-method-view`.

Steps A/B were reported immediately upon completion, before C began. [A/B submission summary](step-B.md).

| Step | Exit | Result |
|---|---:|---|
| A gofmt -l internal/lower/ | 0 | Empty output. |
| A go vet ./internal/lower/ | 0 | Passed. |
| B combined generic / NodeWatch pins | 0 | Every test and subtest passed; all 18 generic subtests (including o2 for both extensions) and all NodeWatch tests. |
| B host-declaration mutant, NodeWatch tests | 1 | The three intended NodeWatch tests went red. |
| B exact restoration / Git diff | 0 / 0 | Restored byte-for-byte; whole source worktree clean. |
| C whole lower, -timeout 40m | 0 | Passed in 653.506 seconds; no failures, no timeout. |
| C manifest | 0 | Generated at exact head. |
| C admission delta | 0 | Complete-corpus PASS; 4,957 programs, zero omissions, zero newly admitted, **zero new refusals**. |

## Commands

All run output and exit codes are retained as `.log` / `.exit` pairs. Go 1.27.1, clang 20.1.8 and Node 24.19.0 are selected with `/workspace/adamic-tools/env.sh`.

```sh
gofmt -l internal/lower/
go vet ./internal/lower/
go test -count=1 -run 'TestAGenericMethod|TestNodeWatch' -v ./internal/lower/
# Remove exactly " || l.hostDeclaredSymbol(inside)":
go test -count=1 -run 'TestNodeWatch' -v ./internal/lower/
# Restore exactly, then:
ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-typescript-source go test -count=1 -timeout 40m -skip '^TestProduct_' ./internal/lower/
python3 cloud/admission-corpus/manifest.py --sha HEAD > /workspace/adamic/evidence/c9fa1ea/manifest.json
go run ./cmd/adamic-admission-delta --base ff41896cee93f832a4a8fa68c4c1822ddef5885e --head c9fa1ea7129678df8e3e01875830808ec87c9b4a --manifest /workspace/adamic/evidence/c9fa1ea/manifest.json --manifest-generator cloud/admission-corpus/manifest.py --json
```

The upstream TypeScript checkout is the exact `050880ce59e30b356b686bd3144efe24f875ebc8` required by stage3/source.json. The preceding evidence verified census passes at both base and b3a5b76b with this pin. This full lower run uses the same pin, and passes.

## Host-declaration mutant

`host-mutant.patch` removes only ` || l.hostDeclaredSymbol(inside)` from invariance.go. The baseline is green; the mutant is red:

- TestNodeWatchInheritedOnMutant
- TestNodeWatchTscShape
- TestNodeWatchInheritedOnStaysNotYet/view (and its parent)

The full messages and every red row are in host-mutant.log and host-mutant-red-rows.log. Byte-for-byte restoration and Git diff checks both exit 0. No permanent compiler or test edits were made.

## Full suite and admission

`lower.log` records the whole package pass; lower-failure-index.log is empty.

[New refusals](new-refusals.md): **none**. The admission JSON records a complete, non-void pass, zero newly admitted programs, 13 Node agreements, 50 accepted negative witnesses and 37 ruled runtime stops. The three host_watch programs newly refused by preceding heads are no longer newly refused.

As in the preceding successful admission runs, npm provisioning used a temporary ~/.npmrc `cafile=/etc/ssl/certs/ca-certificates.crt` with certificate verification enabled. The original configuration was restored and checked afterward. No compiler or tool code was changed except the requested exactly restored mutant.

All committed files for this run are under evidence/c9fa1ea/.
