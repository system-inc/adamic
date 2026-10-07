Rebased the owned wave-04 branch onto current main and the newly advanced lint area, then preserved fresh green oracle evidence.
Commits: tested 840cfc75ee5c27c5aeedef3450cdd4a7c9d2044d; area-rebased 71bacde735bc967fc1e03c65b43cf6af203f3373 has the identical Git tree; evidence commit follows.
Commands: six completed native rules and two reserved JSX source visitors pass their Go/corpus/sanitizer/handle oracles; all owned partial validators and adopted main checks pass.
Mutants: eight clean source byte mutants, retained-handle mutants, all partial helper/refusal/metadata mutants and adopted record mutants are caught.
Uncovered: constructed-context source analysis, shared checker/factory integration, JavaScript checker-call execution and the full repository gate; no new claims.

Main moved from 39638d9e2 to c7991b900. `git rebase --rebase-merges origin/main` succeeded. A normal merge of the original lint area retained its ancestry without a source change. During final verification the area advanced to b84a9d9314b65d3d0261ee017e233287b4f071da by merging that same main. `git rebase --rebase-merges origin/area/stage1-lint` then succeeded. Its tree ddfc28f8f3e9e78f90e8f653a453531f594ae78c is exactly the tested tree, so the successful observations apply without rebuilding identical source. Both current main and current area are ancestors. No protected compiler or shared harness file was manually edited.

The fresh compiler was built with `go build -o /workspace/typeaware-wave-04-landing-d65/adamic ./cmd/adamic`. Setup was already completed for this restored environment: tool/submodule readiness 0s each, cache warm 124s, total 124s; current nproc is 5. Scratch artifacts reused owned directories to stay inside the 32GB disk; 1.4GB remains. Every check wrote directly to its own log, retained in evidence/landing-c799/validation.tar.gz alongside raw oracle outputs and result records.

Original command:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE04_ARTIFACTS=/workspace/typeaware-wave-04-landing-d65/original ADAMIC_WAVE04_COMPILER_MANIFEST=/workspace/typeaware-wave-04/compiler.manifest ADAMIC_WAVE04_REPOSITORY_MANIFEST=/workspace/typeaware-wave-04/repository.manifest ADAMIC_TYPESCRIPT_SOURCE=/workspace/typeaware-wave-04/typescript go test ./stage1/cohere/typeaware -run '^TestWave04AgreementAndMutants$' -count=1 -v -timeout=30m > /workspace/wave04-fragment-source/landing-c799-original.log 2>&1
```

PASS 248.925s. Controls 39 findings / 17,123 identical bytes; repository 14 / 23,598; compiler 200 / 77,043, each normal and sanitized. Casing, matching-return and strict-void mutants compile, exit zero with empty stderr; only Go comparisons catch differing bytes 5752, 122 and 1588. Released handle exits 70; a retained registry handle compiles/exits zero and fails the required-panic assertion.

Continuation command:

```sh
python3 stage1/cohere/typeaware/wave_04_next/validate.py /workspace/typeaware-wave-04-landing-d65/next --adamic /workspace/typeaware-wave-04-landing-d65/adamic --compiler-manifest /workspace/typeaware-wave-04/compiler.manifest --compiler-config /workspace/typeaware-wave-04/typescript/src/compiler/tsconfig.json --repository-manifest /workspace/typeaware-wave-04/repository.manifest --blocking-controls /workspace/typeaware-wave-04-next/blocking-controls-final --process-controls /workspace/typeaware-wave-04-next/process-controls-ts > /workspace/wave04-fragment-source/landing-c799-next.log 2>&1
```

PASS all three positive rules, complete findings/fixes/suggestions wire bytes, requested corpora, ASan/UBSan/leaks and released/retained handles. Three span mutants compile and finish cleanly; comparisons catch bytes 80, 79 and 9245. Whole-process native versus Go observations: repository 0.292519s / 0.153563s; compiler 2.183686s / 0.377176s. These overlapped other checks, not isolated repeated benchmarks.

Both JSX source `validate_source.py` commands from SOURCE_RULES.md were rerun against the rebuilt compiler and rebuilt continuation archives. PASS both modes, positive controls, both frozen corpora under sanitizers, span mutants and exit-70 released handles. The entire undefined-name validator now includes the .cjs control with allowJs:true and passes. Fragment normal compiler native/Go 3.925453s/0.847202s and repository 0.661825s/0.450825s; undefined-name compiler 2.091547s/0.544372s and repository 0.446352s/0.362714s, concurrent observations.

All existing owned validators were rerun with `--adamic /workspace/typeaware-wave-04-landing-d65/adamic`, reusing their corresponding owned landing directories:

- JSX reporting: 56 records / 4,941 bytes; Go/native/sanitizers/source Node/emitted JS; three reporting and three removed-refusal mutants.
- JSX references/declarations: 187 actual Go AST snapshots / 1,291 records / 16,619 bytes; five execution backends; twelve successful-exit comparison-only mutants.
- Parked React reporting: 11 findings / 5,948 bytes, six reporting/refusal mutants; native/sanitized/source/emitted JS comparisons.
- Refs joins: 1,297 records / 55,984 bytes, nine comparison-only mutants.
- Refs environment: 451 records / 9,943 bytes, six comparison-only mutants.
- Refs predicates: 96 cases / 180 records / 3,136 bytes, four comparison-only mutants.
- Named listeners: nine rules / 658 bytes, nine compiling comparison-only metadata mutants; source Node passes; emitted-JS checker imports remain unsupported.
- Rule JSON: nine production/module/manifest agreements and nine rejected in-memory metadata mutations.

`go test ./bridge/tsgo/checker ./stage1/cohere/lint/registry -count=1` PASS 0.181s and 0.068s. `go vet` for these packages plus typeaware passed with an empty log.

The new main changes include proof lowering and record runtime code. Supplied `go test ./internal/lower -run 'Test(ProvenRelations|UnprovenPredicate|PredicateBodies)' -count=1` passed 1.482s. Supplied `go test ./internal/native -run '^TestRecord(sAgainstNode|Mutants|ReadMutants)$' -count=1 -v` passed 61.139s; its Node and sanitizer/mutant evidence is preserved. No new core mutation was introduced by this unit.

The first filtered internal/oracle command selected only TestTheOracleCatchesOneByte, which passed 2.999s. A corrected exact subtest filter then ran all five adopted fixtures:

```sh
go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/^proven_(guards|class_guards|assertions|satisfies|upcasts)\.a$' -count=1 -v > /workspace/wave04-fragment-source/landing-c799-proof-node.log 2>&1
```

PASS 0.844s, all five named fixture subtests. Worker oracle caching was enabled, as permitted by CLAUDE.md; this was not an uncached integration gate. The source and generated native/JavaScript comparisons remain the oracle's responsibility. The 17 required-input correctness checks were not invoked because this is a filtered worker validation, not a full gate. No selected check skipped, and none was relaxed or deleted.

Landing-first took this unit. Constructed-context source construction/component/stability/escape analysis remains owned unfinished work, not a HIR blocker. Shared RuleContext still lacks checker/program access needed to integrate the two native JSX visitors as shared factories. Existing claims stay reserved. No additional rule was selected or claimed, and only codex/typeaware-wave-04 will be pushed.
