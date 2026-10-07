Built: completed six previous claims, then claimed and ported no-lone-blocks, no-lonely-if and no-loss-of-precision.
Commits: six ports c0612bbb, ae160b6c, 0659fc7d, d60b3d7d, 0378e63b, 454e5303; evidence 110b24f6; next claim 38b0f50e.
Checks: all three new original Go suites, 263 fixture files and 236 compiler/stage1 sources match on Node, emitted JavaScript and sanitized native.
Mutants: each new semantic mutant and the additional eleven-pass fix mutant compile, exit zero with clean stderr, and are caught only by Go byte comparison on all three runtimes.
Limits: shared .a registration and independent repair integration remain pending; no shared files were edited, no full repository gate is claimed.

[Previous six complete reports and rates](../typescript-eslint-prefer-as-const/COMPLETE_REPORT.md), [selection snapshot](evidence/selection.json), [complete comparison log](evidence/complete.log), [edge log](../no-loss-of-precision/evidence/edges.log).

| Rule | Fixture files/findings | Corpus findings | Native findings/s | Node findings/s | Go findings/s |
| --- | --- | --- | --- | --- | --- |
| no-lone-blocks | 78 / 47 | 0 | 850.64 | 1193.74 | 4886.87 |
| no-lonely-if | 33 / 24 | 44 | 37.22 | 53.53 | 230.30 |
| no-loss-of-precision | 152 / 57 | 0 | 402.15 | 541.73 | 2335.23 |

Rates use the best of five interleaved full-program count runs including startup/read/parse. Zero-natural-finding rules use 500 witness copies, producing 1,000 block findings and 500 numeric findings respectively. The corpus pins TypeScript 050880ce59e30b356b686bd3144efe24f875ebc8, 77 compiler files and all 159 stage1 sources present at validation. Every original fixture for these three rules is included. Repairs include complete converged source, not just finding counts.

Additional edge comparisons cover Unicode repair trivia, nested overlapping repairs, convergence budget, integer/exponent boundaries and 1,500 deterministic numeric lexemes. The budget mutant changes ten passes to eleven and is caught on all three runtimes solely through external Go comparison.

Toolchain setup exits 1 during warm-up because profile_test.go:32 ranges over portFiles without calling it. Tool installations report 0s each; no successful final setup timing is emitted. nproc is 5, CPU quota is 4; Go 1.27.1, Node 24.19.0, clang 20.1.8. Owned builders bypass that shared package failure. go vet ./... has the same pre-existing error. Filtered oracle TestTheOracleCatchesOneByte and TestCountsAreRecorded passes in 13.498s; gofmt -l cmd internal is empty. See the preceding six-rule report for their exact commands and retained logs.

The cohere CLI was built successfully, but --no-fix --no-cache on the seven owned .a sources exits 1: none is a TypeScript or JavaScript file in its program. [Exact CLI output](evidence/cohere.log). No self-lint pass is claimed.

Read-only inspection of origin/codex/lint-harness-dot-a at 2650ad595b82220c368631ea13139fad4b306ed6 confirms new .a, emitted-JavaScript, profile and suggestion support. That shared branch was not merged here. Independent fix ranges and multiple fix edits still need integration: its fixer uses diagnostic ranges and Go serializer refuses multiple fixes. New no-lone-blocks and no-loss-of-precision already report through context; no-lonely-if refuses enabled shared execution until its independent block fix can be represented safely. The previous six retain the same explicit refusal and complete owned comparisons.

Automatic approval review rejected running the shared registry generator because it writes outside the owned rule directories. It was not rerun; read-only inspection and owned standalone builders provided the permitted alternative. This is not a shared registry or full-gate pass.
