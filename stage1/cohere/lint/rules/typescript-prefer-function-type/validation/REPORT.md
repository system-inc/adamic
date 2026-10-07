Built prefer-function-type, prefer-namespace-keyword and triple-slash-reference as rule-directory candidates; Google-font decisions were completed separately, with JSX execution blocked.
Claim 878da5f8 was pushed before this batch's code; Google partial port 3b58afe4 and all earlier claimed work were already pushed.
validate.py passed 131 upstream cases, 3 owned witnesses and 233 compiler/stage1 files, byte for byte on source Node, emitted JavaScript and sanitized native.
The array-parentheses, export-span and repeated-directive mutants each compiled and ran with exit zero and empty stderr; all three runtime comparisons caught each mutant.
Default integration on this checkout remains blocked by shared profile compilation and fix-range handling; no full repository gate or post-integration default-harness certification is claimed.

# Observed comparison

The public rules are @typescript-eslint/prefer-function-type, @typescript-eslint/prefer-namespace-keyword and @typescript-eslint/triple-slash-reference. Their directory descriptors register without editing any shared list. Sources use Ahra's explicit .ts fallback; the independent driver is driver.a. All new edits are inside these three rule directories, apart from the earlier required claim update.

The unmodified Go rule bodies and tests came from cohere commit 715ba94f3608a6500086b1076ce5cb7e51b836db. TypeScript compiler sources came from v6.0.3 commit 050880ce59e30b356b686bd3144efe24f875ebc8. The comparison has 833 manifest rows: 39 function-type cases, 22 namespace cases, 70 triple-slash cases, 3 owned witnesses, and each rule over all 233 compiler/stage1 files (77 compiler, 156 stage1). No successful upstream Run was excluded. Raw testdata witnesses and generated artifacts are outside the stage1 source graph; each raw witness was separately replayed as source.

The rule-owned Go driver selects the actual upstream rules, sorts diagnostics stably, renders Go findings and serializes diagnostic ranges, fix text, actual fix ranges and final source. The Adamic driver reproduces that protocol independently. All baseline outputs matched 37,198,852 bytes. Sanitized native compilation used native.Options{Sanitize:true}; every successful build and run had empty stderr. output-hashes.json records complete-output SHA256 hashes and lengths. corpus.json records every input hash, decoded options and byte length. The small witness logs retain representative full outputs without committing the 37 MB volume logs. The original complete logs remain in /tmp/w05-third-validation on this workspace.

The Go tests deliberately contain malformed reference directive syntax. Their own rule-testing path still invokes the listener. The independent oracle therefore does the same, without rejecting parser diagnostics before the rule runs, and bypasses the parse-validating fix engine only when no fix proposal exists. Files with fixes still use cohere's original edit.FixText engine, including reparse and convergence checks. The independent Adamic fixer reparses each pass and applies actual ranges. This covers findings and unchanged source for those malformed directives, rather than claiming they are compiler-valid programs.

The Go option-decoder and wrong-type panic tests pass in upstream-tests.log. Cross-runtime replay uses the decoded options captured from successful rule runs. Arbitrary malformed raw configuration, pointer-versus-value option identity and their exception contracts are not represented by this decoded JSON protocol; no cross-runtime claim is made for those configuration errors.

# Reproduction and checks

Run from the repository root with the pinned compiler checkout available:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint-wave1-05-typescript python3 stage1/cohere/lint/rules/typescript-prefer-function-type/validation/validate.py > /tmp/w05-third-validation.log 2>&1
```

The script builds an owned Go overlay, captures successful unchanged upstream tests through a scratch testing overlay, builds emitted JavaScript and sanitized native from driver.a, compares all three runtimes, runs each compiling semantic mutant, then measures release-native, source-Node and Go throughput. It never changes shared tracked files. Its temporary overlays and binaries are under /tmp/w05-third-validation. All test stdout and stderr are redirected into log files.

Focused upstream command: go test -overlay=/tmp/w05-third-validation/capture-overlay.json ./internal/lint/rules/typescript -run '^Test(PreferFunctionType|PreferNamespaceKeyword|TripleSlashReference|DecodeTripleSlashReferenceOptions)' -count=1 -v, from cohere. PASS, 131 distinct successful source/options cases captured. See upstream-tests.log.

Directory generation: go run ./cmd/lint-registry. PASS; registration.log.
Registry package: go test ./stage1/cohere/lint/registry -count=1. PASS; registry-tests.log.
Filtered compiler oracle: go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1. PASS; compiler-one-byte-oracle.log. An initial filter selected no tests and was replaced by this exact uncached test.
Shared package: go test ./stage1/cohere/lint -run 'TestLint' -count=1. BUILD FAIL at profile_test.go:32 because portFiles is a function, not a rangeable list; shared-profile-blocker.log. No full repository gate was run.

# Mutants

| Rule | Semantic mutation | What disagrees |
| --- | --- | --- |
| prefer-function-type | Remove ArrayType from the required parentheses cases | Fix text and resulting valid source type; array of functions becomes a function returning an array |
| prefer-namespace-keyword | Compute diagnostic start after DeclareKeyword instead of ExportKeyword | Exported declaration finding span and rendered column |
| triple-slash-reference | Retain the first repeated types directive instead of the last | Report position; message and finding count remain unchanged |

Every mutant uses the same complete corpus. All nine mutated runtime executions exited zero with empty stderr; only byte comparison rejected them. mutant-differences.json records the first differing line for each runtime, expected Go text, mutated text and clean execution status. Each rule's mutant.json contains the exact single source replacement. comparison.log records all three baseline passes and all nine mutant kills.

# Findings per second

These are whole-process measurements over the 77 compiler files plus an owned 1,000-statement stress source per rule. Every engine reports exactly 1,000 findings. Counts are cross-checked in each of three interleaved Go/native/Node rounds; the best monotonic elapsed time is reported. Startup, parsing, listeners and fix construction are included; serialization and fix application are excluded with --count. Native throughput uses the release build; correctness uses the separate sanitized build. Node means the .a source driver through oracle/node.mjs. These synthetic volume rates describe this machine and corpus, not all projects.

| Rule | Native findings/s | Node findings/s | Go findings/s |
| --- | ---: | ---: | ---: |
| prefer-function-type | 253.96 | 660.03 | 5,259.12 |
| prefer-namespace-keyword | 942.33 | 1,226.93 | 5,500.22 |
| triple-slash-reference | 503.50 | 939.03 | 4,053.28 |

# Toolchain and integration limits

nproc is 5. The already-completed initial cloud/setup.sh timing lines were Go ready 0s, clang ready 1s, Node ready 1s, submodules ready 1s, build cache warm 75s, done in 75s on 5 processors. Continuation setup reported Go, clang, Node and submodules ready in 0s each, then failed cache warming on the same profile_test.go compilation error; no cache-warm or completion timing was printed. Both exact logs are retained. Sourcing the installed /workspace/adamic-tools/env.sh and using the owned driver worked around that shared-package failure. Versions: Go 1.27.1, clang 20.1.8, Node 24.19.0.

The old shared Go serializer requires fix range to equal diagnostic range, and this checkout's Linter.fixed() replaces the diagnostic range. Both fixing rules need a different edit range. They therefore store actual ranges in editStart/editEnd and mark repairs range-fix, which the old default fixer safely skips. The owned comparison driver maps that marker to fix in its serialization and applies the true ranges.

During final validation origin/codex/lint-harness-dot-a became available at 2650ad595b82220c368631ea13139fad4b306ed6. Inspection showed profile compilation, .a discovery, emitted-JavaScript comparison and actual-edit serialization support; its newer base also has a fixer using actual edit ranges. This unit fetched it but did not merge its broader base or modify shared files. Integration should bring in that owner-maintained harness, switch the two owned range-fix markers to fix, and rename Adamic .ts sources with the integration codemod. The existing private protocol is evidence, not a change to the shared protocol. No default-harness pass after that future integration is claimed.

The earlier Google-font partial port remains blocked by shared JSX parsing, recorded in [its owned README](../../next-google-font-display/README.md) and validation. Its 276 URL/entity decisions matched Go and its decision mutant was caught on all three runtimes, but full JSX findings/ranges and full-rule throughput were not certified. These limitations were pushed before claiming this batch.
