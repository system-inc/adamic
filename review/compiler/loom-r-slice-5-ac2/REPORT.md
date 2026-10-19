Freshness readers follow all 13 relocated refusal controls; no reader was removed.
One commit on dbdf977f5b8b0979d414ebe6f20e7aadd312e309, carrying 320b9bd2 plus this repair.
Baseline 14/14; full freshness 918 passing events; a-check 58/58; reader guard and vet passed.
Stale-reader overlay mutant caught by TestFreshCorpusUnitsCoverEveryProgram (exit 1).
No full gate or ahra run; inherited source validation and limitations remain in the earlier report.

The exact pinned main base passed all 13 old reader names plus the coverage guard; see logs/base-readers.jsonl. There were no conflicts applying the source commit.

The coverage guard is `TestFreshCorpusUnitsCoverEveryProgram`. Each old reader below failed on the source branch because its old path no longer exists. Each has a one-to-one canonical replacement, at the same corpus index. Reader count remains 828. The two narrow fresh-only globs register the moved possible/refuse controls, preserving flow exclusion. The generator and its 13 manifest rows now retain these controls. Assumption: register only these two moved control families, rather than every file in the refused directory.

| Former failing reader | Replacement reader | Seconds |
|---|---|---:|
| `TestFreshWrites____oracle_testdata_non_null_possible_read_a_03dbf385a467` | `TestFreshWrites____oracle_testdata_non_null_refused_non_null_possible_read_a_72e36be4a66c` | 0.03 |
| `TestFreshWrites____oracle_testdata_non_null_possible_typeerror_a_6a25b0740445` | `TestFreshWrites____oracle_testdata_non_null_refused_non_null_possible_typeerror_a_be509029a57a` | 0.04 |
| `TestFreshWrites____oracle_testdata_non_null_refuse_argument_a_5044636ebaa9` | `TestFreshWrites____oracle_testdata_non_null_refused_non_null_refuse_argument_a_f0d850484ee0` | 0.04 |
| `TestFreshWrites____oracle_testdata_non_null_refuse_checked_a_15fb80217f3b` | `TestFreshWrites____oracle_testdata_non_null_refused_non_null_refuse_checked_a_3aa862f44c91` | 0.04 |
| `TestFreshWrites____oracle_testdata_non_null_refuse_comparison_a_af272e60c968` | `TestFreshWrites____oracle_testdata_non_null_refused_non_null_refuse_comparison_a_a84152537348` | 0.03 |
| `TestFreshWrites____oracle_testdata_non_null_refuse_conditional_a_9e4099a84016` | `TestFreshWrites____oracle_testdata_non_null_refused_non_null_refuse_conditional_a_ed59101807c8` | 0.04 |
| `TestFreshWrites____oracle_testdata_non_null_refuse_element_a_642d760771d2` | `TestFreshWrites____oracle_testdata_non_null_refused_non_null_refuse_element_a_409e4a444a92` | 0.05 |
| `TestFreshWrites____oracle_testdata_non_null_refuse_expression_a_fd86ddc65e92` | `TestFreshWrites____oracle_testdata_non_null_refused_non_null_refuse_expression_a_b55b07a0a483` | 0.04 |
| `TestFreshWrites____oracle_testdata_non_null_refuse_logical_a_d8ec7fe16365` | `TestFreshWrites____oracle_testdata_non_null_refused_non_null_refuse_logical_a_0a0566b296fb` | 0.03 |
| `TestFreshWrites____oracle_testdata_non_null_refuse_null_a_bc3f0b62bf50` | `TestFreshWrites____oracle_testdata_non_null_refused_non_null_refuse_null_a_32e4e1708057` | 0.04 |
| `TestFreshWrites____oracle_testdata_non_null_refuse_return_a_835892235f21` | `TestFreshWrites____oracle_testdata_non_null_refused_non_null_refuse_return_a_b81e9f6f2f0a` | 0.03 |
| `TestFreshWrites____oracle_testdata_non_null_refuse_spread_a_3f0140d4d4aa` | `TestFreshWrites____oracle_testdata_non_null_refused_non_null_refuse_spread_a_dea22319f423` | 0.03 |
| `TestFreshWrites____oracle_testdata_non_null_refuse_undefined_a_7b36c538740b` | `TestFreshWrites____oracle_testdata_non_null_refused_non_null_refuse_undefined_a_4a2f04eeea6f` | 0.04 |

Paths are mapped explicitly in reader-migration.json. The generated binding count is unchanged. Every renamed leaf still calls t.Parallel first. Coverage guard itself was unchanged.

Commands (all test output redirected directly into logs; GOMAXPROCS=4, GOFLAGS includes -p=4, ADAMIC_UNIT_BUDGET=1 for full fresh):

- `go test ./internal/fresh -run <13 original names plus coverage guard> -count=1 -json -timeout 90s`: 14 pass on the pinned main base (selector recorded by the original reader names above).
- `go test ./internal/fresh -count=1 -json -timeout 90s`: 918 pass events including subtests, no skips/failures; longest 2.64 seconds. Exact per-test timings in test-seconds.json. Includes the coverage guard and remainder.
- `python3 review/compiler/loom-r-slice-5-ac2/check-headers.py final`: actual branch compiler and unchanged pinned Gate.aCheck, 58/58 pass; 42 checked, 0 NotYet, 16 refused, 0 type errors. Per-file statuses/reasons are in final.json; paths.txt is the exact 58-file inventory.
- `go test ./internal/ir -run ^TestCallTargetReaders$ -count=1 -json -timeout 90s`: pass (logs/reader-guard.jsonl).
- `go vet ./internal/fresh`: exit 0.
- `go vet ./cmd/adamic ./internal/flow ./internal/fresh ./internal/ir ./internal/javascript ./internal/lower ./internal/native ./internal/oracle`: exit 0.
- `go test -overlay review/compiler/loom-r-slice-5-ac2/stale-reader-overlay.json ./internal/fresh -run ^TestFreshCorpusUnitsCoverEveryProgram$ -count=1 -json -timeout 90s`: expected exit 1. The valid-Go overlay restores just the first stale path; guard reports missing or duplicated old path at index 550, not a build error. Source never mutated.

No headers added or changed, all 58 program bytes identical to 320b9bd2. Counts unchanged byte-for-byte; no regeneration. Production compiler and runtime lines identical to 320b9bd2. The earlier header-mutant evidence remains in review/compiler/loom-r-slice-5-ac/.

Setup: GOPROXY=https://proxy.golang.org|direct, cloud/setup.sh succeeded; source /workspace/adamic-tools/env.sh; nproc=5 with 4-CPU quota; clang 20.1.8. Timing lines:
    setup: node ready (0.022s)
    setup: go ready (0.025s)
    setup: submodules ready (0.067s)
    setup: markdown dependencies skipped (validated lock and installed bytes); step-duration=0.009s
    setup: markdown dependencies ready (0.077s)
    setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.156s)
    setup: shared cache ready (0.539s)
    setup: go build ready (42.420s)
    setup: test binaries deferred (use --warm-tests) (42.612s)
    setup: build cache warm (42.614s)
    setup: build-flags commit=dbdf977f5b8b0979d414ebe6f20e7aadd312e309 nproc=5 cpu.max=400000 100000 go=go version go1.27.1 linux/amd64 clang=clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261) node=v24.19.0 cached=yes warm-tests=false gocacheprog=on load-before=0.15 0.03 0.03 1/140 4354 load-after=4.93 1.33 0.47 1/144 5188
    setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (42.639s)
    setup: source /workspace/adamic-tools/env.sh
    setup: logs /tmp/adamic-gate/setup.Us3oS2
