# Parked rule branch

Branch: codex/lint-wave1-14. Current-main base: f8013f0baac41ddc340d76f83bddde38536a8f07.
Only owned rule, claim and evidence commits were replayed. The conflicting shared registration and helper foundation commits were excluded.

## Integration blocker

Main lacks the shared RuleContext, directory registration and helper dependencies required by these ports. Its shared finding/comparison contract cannot represent independent edit ranges, fix arrays and suggestion arrays. Those shared files are owned by the harness unification on #zmh9v36. No live shared harness file was edited. Seven repair ports explicitly refuse incompatible registered execution rather than dropping repairs. Integration must supply the shared context, registration, helpers and Diagnostic bridge. The owned complete runners remain the comparison entry points meanwhile.

This is a parked supported-domain oracle result under Ahra's parking instruction. It is not a pass of main's default shared lint driver. Rebase and re-green again when the named harness SHA arrives, before claiming more work.

## Current-main recheck

The current compiler sources build the ports in an isolated checkout with frozen lint dependencies from 661b9e04. parking-foundation.tar.gz is an immutable test snapshot, excluding unrelated inventory/performance artifacts; it is not installed into the branch's shared sources. Actual corpus inputs come from this rebased branch: 77 compiler files at 050880ce59e30b356b686bd3144efe24f875ebc8 plus 254 stage1 .a/.ts files, 331 total. Go cohere is pinned at 715ba94f3608a6500086b1076ce5cb7e51b836db.

All thirteen rule comparisons pass against Go on source Node, emitted JavaScript and ASan/UBSan native, including complete supported fixes and suggestions. Enum corpus output is identical at 1,413,265,526 bytes and 680 findings. Every rule's compiling semantic mutant is caught only by output comparison on all three runtimes. Additional duplicate-options, shared-contract, UTF-8 boundary and eleven-pass convergence mutants are caught. Numeric edge checks include 1,500 deterministic numeric lexemes. Original Go assertions run separately and pass.

Commands: source /workspace/adamic-tools/env.sh; python3 stage1/cohere/lint/rules/nexus-import-require-module-alias/validate_parking.py. This run first completed the three original validators, then resumed with ADAMIC_PARKING_START=typescript-eslint-prefer-as-const/validate_complete.py. Two owned-validator path failures were corrected: canonical cohere paths for Go internal imports, and the boundary builder's subprocess cwd. The affected boundary validator was rerun independently and passed. Its final success log supersedes the failed launch record. No comparison failure was suppressed.

Actual rebased branch: go vet ./... PASS; ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestTheOracleCatchesOneByte|TestCountsAreRecorded' -count=1 -timeout 30m PASS, 25.704s. Full repository gate and default shared lint integration were not run. Setup: tools ready 0s each, submodules 16s, build/cache warm 318s, total reported 318s, nproc 5, CPU quota 4 cores.

## Throughput

Best of five complete count-only runs, findings per second. Rows with 332 files supplement the 331-file natural corpus with 500 witness copies because their natural finding count is zero. These synthetic rows do not claim natural compiler throughput. Timing was serial.

| Rule | Native | Node | Go | Findings | Files |
|---|---:|---:|---:|---:|---:|
| nexus/import-require-module-alias | 1425.21 | 2242.97 | 8206.58 | 2000 | 332 |
| @typescript-eslint/no-duplicate-enum-values | 2.76 | 4.26 | 17.10 | 4 | 331 |
| @typescript-eslint/no-dynamic-delete | 1.39 | 2.21 | 8.62 | 2 | 331 |
| @typescript-eslint/no-misused-new | 1072.75 | 1604.35 | 6378.73 | 1500 | 332 |
| no-unnecessary-type-constraint | 378.66 | 542.85 | 2131.81 | 500 | 332 |
| prefer-as-const | 388.74 | 568.36 | 2224.82 | 500 | 332 |
| prefer-enum-initializers | 514.43 | 778.83 | 3103.33 | 680 | 331 |
| no-extra-non-null-assertion | 771.43 | 1099.40 | 4341.45 | 1000 | 332 |
| no-confusing-non-null-assertion | 386.38 | 546.03 | 2205.85 | 500 | 332 |
| no-unnecessary-parameter-property-assignment | 382.99 | 541.25 | 2259.70 | 500 | 332 |
| no-lone-blocks | 783.60 | 1138.98 | 4375.58 | 1000 | 332 |
| no-lonely-if | 34.81 | 51.00 | 197.57 | 44 | 331 |
| no-loss-of-precision | 364.55 | 545.81 | 2238.04 | 500 | 332 |

## Coverage boundaries

The upstream module-alias JSX silence fixture is explicitly excluded from port comparison because the shared parser has no JSX support. The exact modified-destructured constructor parameter in the parameter-property gaps directory is excluded because the shared parser refuses it. A parameter-property suggestion whose Go byte edit splits a UTF-8 character is explicitly refused on all runtimes; removing that guard compiles and is caught by Go byte comparison. These are existing named supported-domain limits, not full arbitrary-input parity claims. The shared all-rules smoke case was removed from the owned misused-new suite because unrelated repair factories intentionally refuse the incompatible shared contract; rule-selected fixture/corpus parity remains green.

Raw logs are under evidence/parking-*.log. Each rule's mutant.json names its actual semantic mutation; the comparison logs record its clean execution and mismatching bytes. No new helper has been claimed during this parking recheck.
