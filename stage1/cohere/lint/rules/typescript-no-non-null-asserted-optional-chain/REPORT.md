Built: optional-chain assertions and unnecessary type constraints in .a; five JSX decision candidates remain blocked on extraction.
Commits: claim 9ff521a9; implementations 1fe40525, a93ba8ef; JSX decisions 6d7aded4; alias e41b30da was already pushed.
Checks: 67 original cases, 183 compiler/stage1 sources and ten filename/corner cases match Go on source Node, emitted JavaScript and sanitized native.
Mutants: both suggestion-range mutants and five JSX decision mutants compile, exit successfully, and are caught only by output comparison on all three Adamic execution paths.
Not covered: shared production integration, full JSX extraction/ranges, default all-rule mutant corpus and the full repository gate.

The exact published harness is 2650ad595b82220c368631ea13139fad4b306ed6. Its new suggestion classes and serializer close the fourth-batch suggestion-range blocker. A direct cherry-pick conflicts in lint.ts, lint_test.go, main.ts and testdata/oracle.go; it was aborted. Validation uses an isolated checkout of this unchanged commit plus copies of the owned rules and an owned virtual test file. No shared production generator or harness, compiler, parser or cohere source was edited.

Toolchain: Go 1.27.1, clang 20.1.8, Node 24.19.0. Setup on the conflicted checkout initially failed at a merge marker. Setup in the isolated checkout initially refused a symlinked submodule; a real pinned cohere worktree fixed that. Nested submodule setup changed shared Git core.worktree metadata, which was restored to /workspace/adamic/cohere. Final setup succeeded: Go 0s, clang 3s, Node 3s, submodules 22s, warm cache 427s, done 427s, nproc 5, cpu.max 400000 100000, 17.6 GB. Environment sourced from /workspace/adamic-tools/env.sh.

## Observed comparisons

Every test output went directly to its log. Raw logs are in evidence/. The compiler is TypeScript v6.0.3, 050880ce59e30b356b686bd3144efe24f875ebc8, all 77 src/compiler files. The other 106 files are this branch's current stage1 .ts and .a sources, excluding generated registries and intentional gaps. Corpus comparisons select each of the two completed rule names per file, preserving whole formatted findings, ids, messages, ranges, suggestion edit ids/ranges/text and automatic fixed source. Suggestions remain suggestions. Native correctness uses ASan/UBSan with leak checking.

Commands run from /workspace/scratch/wave12-harness, after sourcing env.sh:

```sh
go test ./stage1/cohere/lint -run '^TestCompletedUpstream$' -count=1 -v -timeout=10m > /tmp/wave12-completed-upstream2.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 go test ./stage1/cohere/lint -run '^(TestCompletedUpstream|TestCompletedCorpus|TestCompletedThroughput)$' -count=1 -v -timeout=20m > /tmp/wave12-completed-final.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 ADAMIC_STAGE1_SOURCE=/workspace/adamic/stage1 go test ./stage1/cohere/lint -run '^(TestCompletedCorpus|TestCompletedCorners)$' -count=1 -v -timeout=10m > /tmp/wave12-completed-corpus2.log 2>&1
go test ./stage1/cohere/lint -run '^(TestCompletedCorners|TestCompletedMutants)$' -count=1 -v -timeout=10m > /tmp/wave12-completed-corners-mutants.log 2>&1
go test ./stage1/cohere/lint -run '^TestCompletedCorners$' -count=1 -v -timeout=10m > /tmp/wave12-completed-corners.log 2>&1
go test ./stage1/cohere/lint/registry -count=1 -v > /tmp/wave12-registry.log 2>&1
go vet ./stage1/cohere/lint ./stage1/cohere/lint/registry > /tmp/wave12-vet.log 2>&1
```

Original fixture comparison: 67 unique cases, 28,022 identical bytes; cold run PASS 272.411s, warm run PASS 31.20s. The first combined final command fails only in the owned corpus builder because it supplied relative paths to Go's normalized-absolute path contract. It is not counted as a complete green run. The corrected corpus command passes: 183 files, 366 selected cases, 24,691,297 identical bytes, PASS 75.336s. The corner filter was initially absent from that built test snapshot; its separate actual run passes: 31,361 identical bytes, PASS 25.595s. Corners include UTF-8 names, doubled assertions, comments before bang, const/in/out generic modifiers, defaults and commented commas across .ts/.tsx/.mts/.cts/.a filenames. Registry PASS 0.056s; vet empty log, exit 0.

The shared TestMutants filter failed before executing variants: its generated all-rule rows attach unrelated allowLoop options to no-this-alias, whose real Go decoder correctly rejects them. No shared corpus is edited. The owned TestCompletedMutants instead uses the original selected-rule fixture rows with their actual options. PASS 47.417s. Both variants compile and exit 0 with no stderr before comparison sees the wrong suggestion range; removing an operand byte and deleting the type parameter name are caught on all three backends. No compiler or sanitizer error is credited as a catch.

From /workspace/adamic:

```sh
go test ./stage1/cohere/lint/rules/next-google-font-preconnect -count=1 -v -timeout=10m > /tmp/wave12-jsx-decisions.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v -timeout=10m > /tmp/wave12-filtered-oracle.log 2>&1
```

Decision candidates: PASS 10.678s, ten manually extracted cases, 1,963 identical verdict/message bytes against the actual unchanged Go JSX rules. Five decision mutants are caught only by output comparison on all three backends. This does not test extracting a JSX source into those facts or finding ranges. The shared parser still refuses JSX and the Go harness still forces ScriptKindTS. Earlier positive blocker witnesses remain valid evidence, not per-rule parity. No descriptors for incomplete candidates are installed. Filtered uncached external oracle PASS 0.408s, native/node misses 1 and hits 0.

Pinned CLI ignores direct .a arguments. Formatting was actually performed on fourteen temporary .ts mirrors then copied back to the owned .a files; 14 cohered, no types/lint check. Self-lint of .ts mirrors in the published harness found a scanner property alias, now removed from the owned constraint rule. Remaining unsafe-call findings stem from that pinned CLI being unable to resolve imported .a suggestion classes. No complete self-lint green result is claimed.

## Findings per second

Best of five interleaved count-only runs: startup, reading, parsing and visitors included; compilation, findings formatting and fixing excluded. Rate native is unsanitized release. Correctness and mutants use sanitizers. The machine also ran bounded checks during some samples, so these are observed workload rates, not a performance guarantee.

| Rule | Files/findings | Native findings/s | Node findings/s | Go findings/s |
| --- | --- | ---: | ---: | ---: |
| no-non-null-asserted-optional-chain | 78 / 1,007 | 547.46 | 880.69 | 4,197.86 |
| no-unnecessary-type-constraint | 78 / 1,000 | 536.99 | 888.15 | 4,067.75 |

77 compiler files alone have seven optional-chain assertion findings and zero constraint findings. Each timing corpus adds 1,000 positive declarations. Best elapsed seconds: optional native 1.839402, Node 1.143419, Go 0.239884; constraint native 1.862239, Node 1.125932, Go 0.245836. Prior alias measurements remain in the fourth-batch report.

## Reproduction and limits

Source env.sh then run validate.py --scratch <fresh-directory> --typescript /workspace/scratch/typescript-6.0.3. It archives the exact published harness, links unchanged pinned cohere, copies only owned rules and the owned virtual tests, and writes separate logs. It does not run setup against the symlink. It uses the original branch's stage1 files for corpus input.

Default worker-branch registration still needs the published harness integrated by its owner. No shared conflict resolution is attempted. JSX rules remain unregistered extracted-decision candidates; binding extraction, decoded entities, ranges and full original-fixture corpus are pending. No full repository gate, config/suppression frontend, arbitrary malformed options, invalid UTF-8 or general recovery claim. No PR opened.
