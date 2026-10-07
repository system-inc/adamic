Landing only: took current sharded lint area onto all 74 completed helpers; no new reservation.
Previous publication afd70b14; tested owned merge 484c9533; publication SHA follows in final response.
Latest trio PASS 99.994s, link guard/shard parity/merge tests and seven Node probes PASS; vet/format clean; setup 44.166s, nproc 5.
All thirteen rerun compiling semantic mutants caught by original Go output comparisons; exact witnesses in evidence/landing27-proof.json.
Not covered: full repository gate and seventeen required external correctness checks as a complete set; no skipped or failed check credited green.

## Landing and unchanged helper inputs

Main remains 71d7e491b3c9724f7a0e2ee754592149e7f9790b. Area advanced to 7076b4ebe16a296129d17f04fe0e14029d793e92, adding a shard command/package and shared driver tests. It merged cleanly onto the owned branch. Both latest fetched bases are ancestors; no shared source was edited manually. Only codex/lint-helpers-05 is published, never main/area, with no PR.

All 2172 tracked input entries covering cmd/adamic, internal compiler code, oracle, cohere, dependencies, helpers and configuration/library paths match afd70b14 exactly. cmd/lint-shards is newly upstream and is tested through the shared harness; it is not used by the helper or compiler build path. Exact object-list identity is recorded in landing27-proof.json. Earlier complete helper proofs remain applicable to unchanged inputs. This unit freshly reran batch26's 97729 same-source comparisons and thirteen variants; it does not claim a complete all-package rerun or new mutant variants. Each mutant compiles, exits zero and has empty stderr under sanitizers before a semantic Go-output mismatch earns credit.

## Required external input failure and recovery

The initial selected shared gate ran TestNodeTableIsLinkOnly and TestShardsAgree. The link guard passed in 137.05s, with 2154 rows and identical 13085662 bytes. TestShardsAgree failed with `set ADAMIC_TYPESCRIPT_SOURCE to the pinned TypeScript checkout: TestShardsAgree needs its compiler files`. This was retained as a failed check, not skipped, bypassed or credited green. The separately selected shards package had no matching tests; it was then run without a filter and all three actual merge tests passed, package 0.057s.

The required corpus was absent. Fetched https://github.com/microsoft/TypeScript.git at tag v6.0.3 into /tmp/lint05-typescript-6.0.3, verified exact tag and commit 050880ce59e30b356b686bd3144efe24f875ebc8, and reran the unchanged shard check with ADAMIC_TYPESCRIPT_SOURCE set. Final PASS 44.612s: 2192 rows, identical 27549087 bytes and count output at one, two and five shards. Initial failure and final pass are both retained. No check was weakened. No new mutant proof is claimed for the upstream shared tests; the owned thirteen helper witnesses remain separately established.

Seven uncached Node input probes PASS 1.105s, seven misses and zero hits. Vet/format logs empty. Setup 44.166s, Go 1.27.1, clang 20.1.8, Node 24.19.0, nproc 5 with quota four cores. Every setup timing line is retained in landing27-setup.log. GOPROXY='https://proxy.golang.org|direct' was exported before setup and build shells. No regex-runtime blocker or unfinished reservation.

## Commands

Build shells source /workspace/adamic-tools/env.sh. Output goes directly to logs. Commands:

```
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/lint05-landing27-setup.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch26 -count=1 -v -timeout=20m > /tmp/lint05-landing27-helpers.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint ./stage1/cohere/lint/shards -run '^(TestNodeTableIsLinkOnly|TestShard.*|Test.*Shard.*)$' -count=1 -v -timeout=20m > /tmp/lint05-landing27-shared.log 2>&1
go test ./stage1/cohere/lint/shards -count=1 -v > /tmp/lint05-landing27-shard-merge.log 2>&1
git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git /tmp/lint05-typescript-6.0.3 > /tmp/lint05-landing27-typescript-fetch.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint05-typescript-6.0.3 ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint -run '^TestShardsAgree$' -count=1 -v -timeout=20m > /tmp/lint05-landing27-shards-final.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v > /tmp/lint05-landing27-node.log 2>&1
go vet ./... > /tmp/lint05-landing27-vet.log 2>&1
gofmt -l cmd internal stage1/cohere/lint/helpers/slot05 > /tmp/lint05-landing27-format.log
```

The full repository gate and its complete seventeen required stage 1 external-input check set were not run. The newly selected shard test received its required input and passed unchanged. This bounded landing gate does not claim whole-rule findings parity, regex engine coverage or new helper readiness. Readiness remains 74 helpers, 383 prerequisite occurrences across 76 consumers, and 51 helper-ready rules under frozen adapter assumptions. All claims complete; no new claim.
