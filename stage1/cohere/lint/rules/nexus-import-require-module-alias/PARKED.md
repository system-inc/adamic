# Parked rule branch

Branch: codex/lint-wave1-14. Current-main base: b8fb957aa839a9e8cb0b54279dd9864fa317bd30.
Only owned rule, claim and evidence commits were replayed. The conflicting shared registration and helper foundation commits were excluded.

## Integration blocker

Main lacks the shared RuleContext, directory registration and helper dependencies required by these ports. Its shared finding/comparison contract cannot represent independent edit ranges, fix arrays and suggestion arrays. Those shared files are owned by the harness unification on #zmh9v36. No live shared harness file was edited. Seven repair ports explicitly refuse incompatible registered execution rather than dropping repairs. Integration must supply the shared context, registration, helpers and Diagnostic bridge. The owned complete runners remain the comparison entry points meanwhile.

This is a parked supported-domain oracle result under Ahra's parking instruction. It is not a pass of main's default shared lint driver. Rebase and re-green again when the named harness SHA arrives, before claiming more work.

## Current-main recheck

The current compiler sources build the ports in an isolated checkout with frozen lint dependencies from 661b9e04. parking-foundation.tar.gz is an immutable test snapshot, excluding unrelated inventory/performance artifacts; it is not installed into the branch's shared sources. Actual corpus inputs come from this rebased branch: 77 compiler files at 050880ce59e30b356b686bd3144efe24f875ebc8 plus 254 stage1 .a/.ts files, 331 total. Go cohere is pinned at 715ba94f3608a6500086b1076ce5cb7e51b836db.

All thirteen rule comparisons pass against Go on source Node, emitted JavaScript and ASan/UBSan native, including complete supported fixes and suggestions. Enum corpus output is identical at 1,413,265,526 bytes and 680 findings. Every rule's compiling semantic mutant is caught only by output comparison on all three runtimes. Additional duplicate-options, shared-contract, UTF-8 boundary and eleven-pass convergence mutants are caught. Numeric edge checks include 1,500 deterministic numeric lexemes. Original Go assertions run separately and pass.

Commands: source /workspace/adamic-tools/env.sh; python3 stage1/cohere/lint/rules/nexus-import-require-module-alias/validate_parking.py. This current compiler rerun completed in one invocation, rebuilding and replaying every supported fixture, corpus, mutant and boundary check. No corpus result was reused after the compiler changed. Historical canonical-path/cwd corrections remain in the owned validators.

Actual rebased branch: go vet ./... PASS. Filtered TestNativeAgreesWithNode/internal/oracle/testdata/inherited_static_field_read.a PASS across Node, emitted JavaScript, sanitized and release native; TestTheOracleCatchesOneByte PASS. Full repository gate and default shared lint integration were not run. Setup: tools ready 0s each, submodules 16s, build/cache warm 318s, total reported 318s, nproc 5, CPU quota 4 cores.

## Throughput

Best of five complete count-only runs, findings per second. Rows with 332 files supplement the 331-file natural corpus with 500 witness copies because their natural finding count is zero. These synthetic rows do not claim natural compiler throughput. Timing was serial.

| Rule | Native | Node | Go | Findings | Files |
|---|---:|---:|---:|---:|---:|
| nexus/import-require-module-alias | 1420.26 | 2310.48 | 8555.91 | 2000 | 332 |
| @typescript-eslint/no-duplicate-enum-values | 2.83 | 4.47 | 17.16 | 4 | 331 |
| @typescript-eslint/no-dynamic-delete | 1.43 | 2.16 | 8.67 | 2 | 331 |
| @typescript-eslint/no-misused-new | 1051.50 | 1691.50 | 6098.47 | 1500 | 332 |
| no-unnecessary-type-constraint | 376.31 | 533.33 | 2143.45 | 500 | 332 |
| prefer-as-const | 377.92 | 521.34 | 2212.53 | 500 | 332 |
| prefer-enum-initializers | 510.75 | 723.35 | 2953.46 | 680 | 331 |
| no-extra-non-null-assertion | 737.02 | 1055.31 | 4325.04 | 1000 | 332 |
| no-confusing-non-null-assertion | 374.56 | 529.38 | 2150.63 | 500 | 332 |
| no-unnecessary-parameter-property-assignment | 369.65 | 528.12 | 2185.39 | 500 | 332 |
| no-lone-blocks | 762.37 | 1106.29 | 4472.86 | 1000 | 332 |
| no-lonely-if | 33.57 | 49.02 | 191.19 | 44 | 331 |
| no-loss-of-precision | 355.63 | 527.05 | 2215.71 | 500 | 332 |

## Coverage boundaries

The upstream module-alias JSX silence fixture is explicitly excluded from port comparison because the shared parser has no JSX support. The exact modified-destructured constructor parameter in the parameter-property gaps directory is excluded because the shared parser refuses it. A parameter-property suggestion whose Go byte edit splits a UTF-8 character is explicitly refused on all runtimes; removing that guard compiles and is caught by Go byte comparison. These are existing named supported-domain limits, not full arbitrary-input parity claims. The shared all-rules smoke case was removed from the owned misused-new suite because unrelated repair factories intentionally refuse the incompatible shared contract; rule-selected fixture/corpus parity remains green.

Raw logs are under evidence/parking-*.log. Each rule's mutant.json names its actual semantic mutation; the comparison logs record its clean execution and mismatching bytes. No new helper has been claimed during this parking recheck.

## Recheck after the stage3 landing

Both owned branches were rebased onto c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06. Compared with f8013f0b, the only changed compiler/dependency/runner/stage1 path is internal/oracle/stage3_hook_test.go, a new test. All compiled program inputs, compiler implementation, tools and the 331-file corpus are identical. The complete corpus result is reused as an unchanged-input cache. Fresh four-way fixture comparisons and all thirteen compiling rule mutants pass again; the actual rebased filtered uncached compiler oracle also passes. The original four use ADAMIC_RULE_TEST_FILTER to select fixture/mutant checks; validate_rebase.py checks the other nine against the retained immutable artifacts. Full validate_parking.py remains the independent rebuild/replay entry point. New parking-stage3 logs distinguish this recheck from the earlier full corpus run.

## Recheck after inherited static-field emission

Rebased onto b8fb957aa839a9e8cb0b54279dd9864fa317bd30. The compiler implementation changed, so the entire supported fixture/corpus/mutant/boundary suite was rebuilt and replayed rather than reusing the stage3 identity cache. All thirteen rules pass. Every wrapper command exits zero, all original Go assertions pass, and all semantic and refusal mutants are caught. The throughput table above records this fresh run. The integration blocker and named parser/UTF-8 boundaries remain.
