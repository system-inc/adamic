Built: rebased the owned branch onto lint area d3a37422c and current main b6b1538b0; no new claims or source algorithms.
Commits: rebased implementation/report tip 32e7acad07904d692c41112554e6bbe85b6d80dd; this evidence commit is published only to codex/typeaware-wave-12.
Checks: all twelve default standalone ports pass 1,143 controls/895 findings, frozen corpora, sanitizer and released-handle checks; 53 partial TSX source cases/30 findings pass four backends; shared lint has zero skips.
Mutants: thirteen rule/regex and five raw-question mutations caught only by Go bytes; four registry retention mutations caught by released-handle contracts; source/report/quote/metadata/refusal inventory below.
Uncovered: three JSX claims remain partial; three HIR React claims remain parked; shared wave receipt is blocked; full repository gate and fresh nondefault standalone parity are not claimed.

Fresh all-head fetch advanced main to b6b1538b0cebc4ba6741ac34f1aedb60293c1d06 and area to d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898. Rebase of the owned branch succeeded without conflicts. The area contains current main. Shared lint sources have no diff from b46914832. No protected compiler, shared harness, registry generator, option guard or submodule pin was edited. Existing shared allocator checks were retained.

Toolchain: bash cloud/setup.sh passed. Go ready 0s, clang 1s, Node 1s, submodules 1s, cache warm 354s, total 354s; nproc 5, quota four CPUs, memory 17.6 GB. Setup and oracle jobs shared host resources; elapsed times are observations, not isolated benchmarks.

Commands source /workspace/adamic-tools/env.sh. All output was written directly to logs. Four Go suites: go test -v -count=1 -timeout 30m ./stage1/cohere/typeaware -run '^TestWave12AgreementAndMutants$'; the corresponding wave12_next, wave12_third and wave12_fourth packages use '^TestAgreement$'. ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-12/corpus; ADAMIC_WAVE12[_NEXT|_THIRD|_FOURTH]_COMPILER_MANIFEST=/workspace/wave-12/compiler.manifest, _REPOSITORY_MANIFEST=/workspace/wave-12/repository.manifest, _ARTIFACTS=/workspace/wave-12/d3-landing/<batch>. PASS root 516.446s, next 563.955s, third 533.009s, fourth 553.001s. Control bytes root 64,090, next 87,043, third 26,027, fourth 331,957. Complete records include fixes and suggestions. Frozen corpora retain 77 compiler and 287 repository roots; no coverage expansion. Normal and ASan/UBSan/LSan runs match; seventeen JSX controls remain included.

PYTHONDONTWRITEBYTECODE=1 python3 stage1/cohere/typeaware/wave12_sixth/check_source_selection.py /workspace/wave-12/d3-source passed 53 TSX cases/30 findings/11,228 complete bytes on native, sanitized native, source Node and emitted JavaScript. Different scratch paths change canonical byte totals. This is the same constrained source subset described in B469_REPORT.md, not full source rule parity. Four mutants compile, exit 0 with empty stderr, then differ only in Go bytes: undef-ascii first difference 42; fragment-member 4140; fragment-initializer 6468; context-inline-array 8252. Unresolved alias and unsupported provider-context inputs panic 70; each bypass exits 0 and is caught by its refusal contract.

PYTHONDONTWRITEBYTECODE=1 python3 stage1/cohere/typeaware/wave12_sixth/validate.py /workspace/wave-12/d3-sixth passed eighteen prepared findings/7,354 bytes, native JSX parsing, 6,202 Unicode quotation inputs across four backends, three report-ID mutations, printable-character-to-X and four quote-width mutations (all caught only by Go bytes), invalid construction and surrogate refusal bypasses, forty-eight named-kind/node metadata mutations and a checker-capability bypass. Prepared reporting uses Go-supplied ranges and message-family inputs; it proves rendering only. Full source-selection witnesses and streams are compressed in evidence/d3/d3-source; prepared evidence is in evidence/d3/d3-sixth.

ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-12/typescript go test -json -v -count=1 -timeout 30m ./stage1/cohere/lint -run '^Test(CompilerAndStage1Agree|RulesAgree|DecodedOptionsAndMutant)$' passed 570.554s. RulesAgree 323.42s; CompilerAndStage1Agree 190.38s on 450 files; DecodedOptionsAndMutant 56.40s. Zero skips/failures. The ignored-options mutation is caught on source Node and native; the options guard was neither relaxed nor bypassed. Scoped go vet ./stage1/cohere/typeaware/... ./stage1/cohere/lint passed. The other eight required-input packages were not rerun this turn; earlier evidence remains in LATEST_LANDING_REPORT.md. Full repository gate not run.

python3 cloud/lint-wave-check.py still exits 1 before parity tests: missing committed stage1/cohere/lint/claims/codex/typeaware-wave-12.json. Existing reservations are in the older typeaware claims format and standalone suites are not registry callbacks. No receipt or overall landing-ready status is fabricated.

Remaining shared blockers: RuleContext has no checker/file-program handle; buildPort still uses native.Build(native.C(...)) rather than the C-checker profile; oracle/adamic.mjs exports no tsgoProgram; emitted JavaScript refuses an unlinked typescript-go primitive. checker_gap.a is the reproducer, with source Node panic 70 and emitted compilation exit 1. Its successful bypass fails the capability contract. Closing these gaps requires shared files outside authorized rule directories. Unimplemented source algorithms remain fragment binding/import matching, JSX name declarations/global resolution, component decisions, recursive construction discovery and memo/binding stability. These three claims stay ACTIVE and INCOMPLETE. Three HIR claims stay PARKED. Under landing-first and the shared-file restriction, no further claims were taken.

Fresh timing samples run each rebased native binary and Go oracle with the same config and frozen manifest; every sample compares complete stdout bytes. One sample per implementation/corpus, on a shared host, not medians and not causal performance evidence. No prepared-helper versus full-rule ratio is claimed.

root compiler: native 5.114s, Go 1.287s, ratio 3.97.
root repository: native 0.667s, Go 0.211s, ratio 3.17.
next compiler: native 2.848s, Go 0.417s, ratio 6.83.
next repository: native 0.443s, Go 0.139s, ratio 3.20.
third compiler: native 3.404s, Go 0.492s, ratio 6.92.
third repository: native 0.644s, Go 0.154s, ratio 4.19.
fourth compiler: native 5.273s, Go 0.516s, ratio 10.22.
fourth repository: native 0.480s, Go 0.135s, ratio 3.54.

Every standalone mutation and lifetime result, directly from fresh logs:

root: wave_12_test.go:127: no_redeclare.a mutant exits 0, empty stderr, byte oracle catches byte 54
root: wave_12_test.go:127: no_test_on_global_regex.a mutant exits 0, empty stderr, byte oracle catches byte 5556
root: wave_12_test.go:127: no_write_only_collection.a mutant exits 0, empty stderr, byte oracle catches byte 9276
root: wave_12_test.go:155: released handle panics 70; retaining registry mutant exits 0 and is caught
next: agreement_test.go:185: no_process_exit_after_output.a mutant exits 0, empty stderr, byte comparison catches byte 3952
next: agreement_test.go:185: no_uncleared_race_timeout.a mutant exits 0, empty stderr, byte comparison catches byte 4619
next: agreement_test.go:185: require_blocking_standard_streams.a mutant exits 0, empty stderr, byte comparison catches byte 720
next: agreement_test.go:200: ancestry-declaration raw question mutant exits 0, empty stderr, byte comparison catches byte 70
next: agreement_test.go:200: signature-body raw question mutant exits 0, empty stderr, byte comparison catches byte 3952
next: agreement_test.go:200: import-target raw question mutant exits 0, empty stderr, byte comparison catches byte 13315
next: agreement_test.go:219: all three new questions reject released handles with panic 70; registry retention mutant exits 0 and is caught
third: agreement_test.go:128: no_new_func.a mutant exits 0, empty stderr, byte comparison catches byte 1413
third: agreement_test.go:128: no_new_native_nonconstructor.a mutant exits 0, empty stderr, byte comparison catches byte 63
third: agreement_test.go:128: no_new_wrappers.a mutant exits 0, empty stderr, byte comparison catches byte 579
third: agreement_test.go:141: declaration-file raw question mutant exits 0, empty stderr, byte comparison catches byte 63
third: agreement_test.go:160: new question rejects released handles with panic 70; registry retention mutant exits 0 and is caught
fourth: agreement_test.go:135: prefer_regex_literals.a mutant exits 0, empty stderr, byte comparison catches byte 523
fourth: agreement_test.go:135: prefer_rest_params.a mutant exits 0, empty stderr, byte comparison catches byte 5627
fourth: agreement_test.go:135: exhaustive_deps.a mutant exits 0, empty stderr, byte comparison catches byte 7009
fourth: agreement_test.go:135: effect_hook_name.a mutant exits 0, empty stderr, byte comparison catches byte 8393
fourth: agreement_test.go:148: binding-presence raw question mutant exits 0, empty stderr, byte comparison catches byte 64
fourth: agreement_test.go:167: new question rejects released handles with panic 70; registry retention mutant exits 0 and is caught

Compressed logs, complete streams, exact command arguments and exit codes are retained in evidence/d3; hashes.json hashes their uncompressed contents. Implementation and test-source changes were limited to rebasing; this continuation adds evidence only.
