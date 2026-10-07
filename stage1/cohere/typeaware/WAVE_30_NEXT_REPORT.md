Built the three continuation claims: collection misuse, discarded outcome and discarded pure result, in separate .a files.
Commits: prior work 333b5ccd pushed first; continuation claim e6496e1a pushed before code; implementation a74cf501; evidence in this commit.
Commands: final three-rule test PASS 70.397s; earlier two-rule regression PASS 47.80s; bridge PASS 65.846s; Node PASS 18.711s; final checker PASS 0.119s; vet/gofmt/diff clean.
Mutants: all three rule mutants exited zero and were caught only by independent Go bytes; canonical identity, ancestry suffix and UTF-8 checker mutants failed their direct checks.
Not covered: full repository gate, every upstream fixture/configuration, JSX/JavaScript corpus; standalone pinned cohere CLI .a support is pending; earlier released lost-update is unimplemented.

# Wave 30 continuation

All three rules in the continuation claim are complete and validated at production defaults. There are no rule options in these three upstream implementations. The prior two ports remain present and passed regression. This is a three-rule continuation, not a completion claim for the earlier unimplemented lost-update rule. No more rules were claimed after Ahra's correction.

## Selection and sequencing

All prior work was already committed and `git push origin codex/typeaware-wave-30` reported Everything up-to-date. Then all origin heads were fetched. Selection combined compiler-all.counts and repository-all.counts from VOLUME_REPORT.md, sorted descending with lexical ties. The 197-entry checker ranking includes 25 of the original 26 ports; method-signature-style is not checker-dependent. Every higher-ranked rule not in those ports was mentioned in an origin claim. The first three free rules were:

- `nexus/correctness-no-collection-misuse`
- `nexus/correctness-no-discarded-outcome`
- `nexus/correctness-no-discarded-pure-result`

All three recorded volumes are zero. Current main and c-library source/validation records were checked for these names and contain only inventory, count and skipped-type records, not ports. The claim commit e6496e1a was pushed before any implementation. `selection.json` preserves the fetched origin heads, pre-claim branch tip, ranked exclusions and claim blob identities. The work stayed on codex/typeaware-wave-30; no PR was opened.

The existing cloud setup passed in 78s: Go ready 0s, clang ready 0s, Node ready 0s, submodules 0s, build cache warm 78s, done 78s. `nproc` is 5, quota 4 cores. All commands source `/workspace/adamic-tools/env.sh`; Go 1.27.1, clang 20.1.8, Node 24.19.0. No new toolchain or submodule pin was installed for the continuation.

## Implementation boundaries

Each rule owns one `.a` file and makes its decisions in native Adamic. New checker questions own separate Go files and matching `.a` decoders:

- `type-leaf-facts`: raw flags, array/tuple status, call signatures, literal value, symbol declaration origins and property lookup for a live type identity.
- `type-declaration-ancestry`: raw or awaited type constituents, their declaration ancestry, locations, names, type-child locations and direct child counts.

Their individual wire contracts are in bridge/tsgo/type-leaf-facts.md and bridge/tsgo/type-declaration-ancestry.md. Neither question contains rule names, outcome alias lists, method allowlists, lint verdicts or diagnostics. The only shared implementation edit is one inspect switch case per question, four gofmt-formatted lines in facts.go. Those registrations were added before Ahra's correction. No shared registration generator or test harness was edited. The new wave-specific runner, test file and production Go oracle are separate files. No protected emitter/lowerer/oracle implementation was changed.

The collection rule reproduces array `in` checks, impossible size comparisons, collection bracket misuse and Object-method misuse, with constrained union parts, literal/property tests and default-library identities. It reproduces the upstream canonical-infinity numeric behavior as well as decimal, hexadecimal, negative-zero and mirrored comparisons.

The outcome rule checks a bare or awaited call at expression-statement scope, groups original type-literal arm identities by union declaration and verifies complete arms of a listed top-level Nexus alias. Paths, alias names, parenthesized types, optional results, generic instantiations, narrowed subsets and deterministic alias priority are native decisions.

The pure-result rule checks every method declaration against the library interface and allowlist, and excludes any function-valued, any or unknown argument. Optional member/call syntax, readonly arrays and union receivers are supported. A project lookalike and a method with a callback stay silent.

## Independent byte agreement

The new oracle loads its own program and invokes the unmodified pinned Go cohere production registry rules. It has its own AST walk and program views, and imports no bridge implementation. Canonical bytes include file headers, spans, rule IDs, message IDs/text, every fix and every suggestion. All three rules currently emit findings without fixes or suggestions; the complete serializer remains active.

Final normal and ASan/UBSan/LeakSanitizer results:

| Population | Files | Findings | Identical bytes |
| --- | ---: | ---: | ---: |
| Controls | 37 | 74 | 32282 |
| Frozen repository | 287 | 0 | 18485 |
| TypeScript src/compiler | 77 | 0 | 5318 |

Every normal/sanitizer native comparison exited zero with empty stderr. Control assertions require positives for every rule, all four collection message IDs and all five listed outcome aliases. Controls include property/index boundaries, subclasses and intersections, constrained generic arrays, optional receivers, shadowed globals, callbacks, unknown values, optional calls, generic outcome arms, success-only narrowing, extra union arms, namespace declarations and Unicode/CRLF spans.

Outcome helper contents are `.a` files. Generated `.ts` symlink filenames reproduce the exact declaration identity checked by the production rule; both checker loaders see the same aliases. No `.ts` Adamic implementation was authored, and no shared module loader was changed. The evidence keeps the helper `.a` contents and hashes, and the wave-specific test recreates the aliases.

Compiler checkout: TypeScript v6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8. Cohere and typescript-go remain at the branch's original pinned commits. The corpus uses the existing portable validation-volume/compiler.manifest and validation-coverage/repository.manifest. input-hashes.json records every source byte count and SHA-256; compressed outputs and output-hashes.json preserve the exact streams. Absolute file headers make stream hashes specific to this artifact directory.

## Mutants, refusals and defects caught

All rule mutants compiled, exited zero and had empty stderr. The independent Go finding bytes alone caught:

| Rule | Mutation | First differing byte |
| --- | --- | ---: |
| collection misuse | Change size threshold `value <= 0` to `value < 0` | 409 |
| discarded outcome | Extend finding end by one byte | 27084 |
| discarded pure result | Extend finding end by one byte | 1098 |

The direct question tests compare flags, array/tuple facts, calls, literals, property presence, declaration origins and every ancestry field with compiler operations. Removing canonical identity spelling rejection accepts 01 and fails the identity check; allowing an arbitrary ancestry suffix fails the suffix check. Removing anonymous symbol display-name normalization produces invalid UTF-8 and fails the explicit wire-validity check. These mutations are isolated Go overlays and are not in the production sources.

The first agreement pass caught missing negative-infinity behavior and optional-chain property-name selection. Subsequent controls caught both and the final pass agrees. A strengthened origin test exposed an internal anonymous symbol name with an invalid UTF-8 byte. The final question normalizes that display name in its own file, while property lookup still uses the original name. No shared origin writer was edited. Initial failed comparison logs and the final passing logs are preserved.

Released handles queried with either new question exit 70 and the exact `invalid or released checker handle` panic. The full bridge gate passes its 100-query C ownership checks and its independent 162-position, 3261-byte normal/sanitizer oracle. Its input/output length mutants trigger ASan; stale registry retention fails the stale assertion; wrong position differs from Go at byte 6; missing link opt-in is refused; missing output free and region allocation on heap trigger LSan.

The seven selected Node fixtures pass native/JS/Node comparisons with sanitizer/leak checks, and the one-byte mutant is caught. Earlier two-rule regression: 25 controls, 17 findings, 10132 matched bytes; both corpora still match; both earlier mutants and the stale lineage query pass their expected failure checks. No additional change to those ports was made.

## Commands and observed output

All test stdout/stderr went to files, never through a pipe. Representative exact commands from the repository root:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE_30_NEXT_ARTIFACTS=/workspace/wave-30-next-validation-reviewed ADAMIC_WAVE_30_NEXT_REPOSITORY_MANIFEST=/workspace/wave-30-repository.manifest ADAMIC_WAVE_30_NEXT_COMPILER_MANIFEST=/workspace/wave-30-compiler.manifest ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-30-typescript go test ./stage1/cohere/typeaware -run '^TestWave30NextAgreementAndMutants$' -count=1 -v -timeout=30m > /workspace/wave-30-next-reviewed.log 2>&1
# PASS 70.397s.
go test ./bridge/tsgo/... -count=1 -timeout=15m -v > /workspace/wave-30-next-bridge.log 2>&1
# PASS bridge 65.846s; checker 0.132s before the final UTF-8 display change.
go test ./bridge/tsgo/checker -count=1 -v > /workspace/wave-30-next-checker-final.log 2>&1
# PASS 0.119s, final source and stronger validity/origin assertions.
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|functions|closures)\.a$' -count=1 -timeout=10m -v > /workspace/wave-30-next-node.log 2>&1
# PASS 18.711s; includes method_closures and generic_functions.
go vet ./bridge/tsgo/... ./stage1/cohere/typeaware > /workspace/wave-30-next-vet.log 2>&1
gofmt -l bridge/tsgo/checker stage1/cohere/typeaware > /workspace/wave-30-next-gofmt.log
git diff --check > /workspace/wave-30-next-diff.log 2>&1
# Empty logs, exit zero.
```

The combined new/earlier rule regression used both sets of ADAMIC_WAVE_30[_NEXT] artifact and corpus variables and `-run '^(TestWave30NextAgreementAndMutants|TestWave30AgreementAndMutants)$'`, passing in 116.823s before the final new-question UTF-8 change. The changed question is used only by the new three-rule runner, which was then rerun in full. The guard and normalization overlay test commands, failures and expected exits are recorded in the corresponding mutant logs.

## Quiet timing

Three alternating native/Go rounds after builds and tests finished. Every round compares its output bytes. Medians are whole-process wall time including program load, startup and teardown, for these three rules together.

| Corpus | Native | Go | Native/Go |
| --- | ---: | ---: | ---: |
| compiler | 2489.972 ms | 734.421 ms | 3.39x |
| repository | 312.848 ms | 149.932 ms | 2.09x |

Raw rounds and load/run/query phases are in timings.json and the adjacent stderr logs. The final runner makes 15564 compiler and 4056 repository adapter queries. Both corpora have zero findings, so these timings measure program load, walks and queries, not positive-finding throughput. No speed-parity claim is made.

```sh
python3 bridge/tsgo/profile/volume_bench.py /workspace/wave-30-next-validation-reviewed/wave-30-next /workspace/wave-30-next-validation-reviewed/wave-30-next-oracle /workspace/wave-30-next-bench --corpus compiler /workspace/wave-30-typescript/src/compiler/tsconfig.json /workspace/wave-30-compiler.manifest --corpus repository /workspace/adamic/tsconfig.json /workspace/wave-30-repository.manifest > /workspace/wave-30-next-bench.log 2>&1
```

## Limits

The full repository gate and every upstream per-rule fixture were not run. The fixed runner does not apply edits or suppressions and is not the full cohere CLI. JSX/JavaScript projects and all ambient/default-library augmentation combinations are not claimed. The pinned standalone cohere CLI's inability to check `.a` files, recorded in the previous report, remains pending shared module support; this continuation did not edit that harness or replace its programs with `.ts`. The wave-specific native runner and independent Go oracle both load `.a` controls. The earlier lost-update rule remains released and unimplemented. The three continuation claims are complete; work stops here under Ahra's correction.
