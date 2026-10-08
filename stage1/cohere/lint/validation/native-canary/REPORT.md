SHA base: 3668d3ae932dfdc280233e47b77a086adad3c0c2; additive commits on lint-rules/native-canary only.
Canary: react/jsx-no-comment-textnodes; JSX parsing, text spans, linked visitation and finding serialization; sanitized native equals 527,103 mutated-Node bytes.
Choices: all 79 semantic mutants use Node and emitted JavaScript; one rule is also the native canary. No memory/ownership mutants in this package. Native baseline checks stay.
Catches: before 70 rules on Node/JS/native and nine extra semantic mutations; after identical 70+9 identities caught on Node/JS, canary equality passes; valid planted survivor and both new-check negatives fail as intended.
Times: package wall 1763.344s vs supplied 2,034s; TestMutants group 177.627s vs supplied 748s. Go parent own running time 0.04s. Full package PASS.

# Native canary policy

The before catch list is historical validation evidence from fetched origin branches, including compressed logs. It is not a fresh re-run of the package-time baseline. The 2,034s package and 748s mutant baselines are the measurements supplied by system_adamic. The exact before backend/source paths and after catch names are in catches-before-after.json. Rule mutant definitions, port source, oracle adapters and witness inputs are unchanged from 3668d3ae9. There are no missing rule names, and all nine additional semantic mutations remain caught. Five extra tests gain emitted-JavaScript coverage where the base had only Node/native.

Node and emitted JavaScript must each differ byte-for-byte from actual Go for every semantic mutation. The canary additionally requires sanitized native to equal the mutated Node output, not merely differ from Go. All generated all-rule rows and the owned JSX witness remain in its manifest. TestMutants fails if the named canary is absent from the registry. Every canary/semantic process must compile and exit cleanly before comparison, as execute already requires.

| Test | Mutation | Backend choice |
|---|---|---|
| TestMutants | 70 rule logic/range/fix/option mutations | Node + emitted JavaScript vs Go; one JSX sanitized canary equals mutated Node |
| TestCountGuardMutant | Count arithmetic | Node + emitted JavaScript vs Go; ordinary output must remain equal |
| TestDecorationOptionMutant | Range membership | Node + emitted JavaScript vs Go |
| TestPositionIndexMutant | First source anchor | Node + emitted JavaScript vs Go |
| TestCommentFoldMutant | Unicode folding arithmetic | Node + emitted JavaScript vs Go |
| TestRegistrationMutant | Listener subscription | Node + emitted JavaScript vs Go |
| TestDecodedOptionsAndMutant | Decoded option name | Node + emitted JavaScript vs Go; baseline native comparison unchanged |
| TestLegacyMutants | Overlapping fix winner text | Node + emitted JavaScript vs Go |
| TestCompleteSuggestionSerialization | Second suggestion edit range | Node + emitted JavaScript vs Go; baseline native comparison unchanged |
| TestFactoryHooks | Finish hook omission | Node + emitted JavaScript vs expected hook sequence; positive native hook check retained |

There are no memory or ownership mutations in these root-package tests: none changes allocation, references, reuse, frees, arenas or object lifetime. The helper/parser/native package tests are untouched. TestNestedOutsideModuleCopy retains its native module-loading refusal check: a deliberately broken copied import is rejected by the loader, and the positive copied program still runs on native. TestEmittedJavaScriptMismatch retains its unmodified native build and ordinary comparison; only emitted JavaScript receives its planted mismatch. These are compiler-loading/comparison proofs, not extra semantic-rule native canaries.

The unmodified-port comparisons are exactly as before. Sixteen baseline functions are byte-identical to the base, including compare/compareWithJavaScript, TestRulesAgree, TestCompilerAndStage1Agree, TestOwnedWitnesses, TestNodeTableIsLinkOnly, TestShardsAgree, snapshot/profile, rename and module-copy checks. The existing positive synthetic serialization and factory-hook native checks also remain; only their incorrect semantic variants drop native. shared_test.go is unchanged, preserving the owner's once-per-run unmodified port, oracle and capture.

Negative proofs:

- A valid no-debugger mutant replaces removable with (removable), leaving findings unchanged. Descriptor validation succeeds and the process runs; the test fails with "debugger fix suppressed mutant survived on Node". The original mutant is restored. A preliminary identical From/To attempt was rejected by descriptor validation and earns no comparison credit.
- In a temporary test edit, the canary builds the unmodified port instead of the mutated copy. It compiles and runs sanitized, then fails only "native canary differs from mutated Node". This proves the new equality assertion detects the wrong edited-copy artifact.
- A temporary canary constant naming react/missing-canary fails "native canary rule react/missing-canary is missing". All test source bytes are restored in finally blocks before final vet/commit.

Top five tests by their own Go-reported running time (not sums of descendant times):

| Test | Seconds |
|---|---:|
| TestCompilerAndStage1Agree | 469.40 |
| TestProfileSnapshotsAgree | 242.32 |
| TestShardsAgree | 196.66 |
| TestJsxLintTrees | 137.82 |
| TestDotARename | 98.64 |

TestMutants runs parallel subtests; its Go parent Elapsed is 0.04s and excludes those subtests. The 177.627s group duration above is the run-to-pass event span and is the relevant comparison to the supplied 748s. Package wall includes go-command overhead; package JSON's own elapsed is separately retained. The observed package wall is 13.3% lower and the mutant group 76.3% lower than the supplied baselines. These are single-run observations, not controlled repeated benchmark claims.

Inputs: TypeScript 6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8; git status --porcelain --ignored is empty before and after. One fresh empty directory is used for both profile variables. cohere 715ba94f3608a6500086b1076ce5cb7e51b836db, nested TypeScript 8d550c837c90bd1805b047b7eeccc2baac2d5e7a. Go 1.27.1, Node 24.19.0, clang 20.1.8, nproc 5 (four-core quota). GOPROXY fallback is set. No WASI SDK was requested by these tests. Two optional throughput tests skip because ADAMIC_LINT_BENCH is unset; no correctness/input check is skipped. Full repository gate is outside this package-only task.

Commands, after sourcing /workspace/adamic-tools/env.sh:

```sh
export GOPROXY='https://proxy.golang.org|direct'
export GOCACHE=/workspace/scratch/wave12-required-go-cache
export TMPDIR=/workspace/scratch/wave12-b469-gate
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3
export ADAMIC_LINT_PROFILE_DIR=$(mktemp -d /workspace/scratch/native-canary-profile.XXXXXX)
export ADAMIC_LINT_PROFILE_SNAPSHOTS=$ADAMIC_LINT_PROFILE_DIR
go vet ./stage1/cohere/lint
go test -p 1 ./stage1/cohere/lint -count=1 -v -json -timeout=60m
```

The JSON test output is redirected straight to its log file, never piped. A Python subprocess timer records wall time because /usr/bin/time is absent. The 60-minute package timeout accommodates the supplied 2,034s baseline. Separate filtered -count=1 -v runs hold the smoke and three intended negative failures. The final code is exactly the fully gated source; final vet and diff --check pass. No main, area or package-time push is made.
