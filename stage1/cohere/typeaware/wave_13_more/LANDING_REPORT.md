Built: rebased the existing nine rule implementations onto main e8ba3d5d; eight are green, while no-object-constructor retains three shared-parser refusals. No new rules claimed.
Commits: rebased code 4e7d3d11, 99046c15 and 585bb91a; pre-evidence tip 81cecbed, replacing pushed df0f8d80. Claims remain earlier than their code.
Commands and outputs: original trio gate PASS 418.694s; bridge PASS 318.896s; expanded Node oracle PASS 17.193s; vet PASS; Nexus 129/129 and core 430/433 controls agree in normal and sanitized runs.
Mutants: all nine rule and seven raw-fact semantic mutants compile and exit 0, caught only by byte comparison; seven question guards, seven foundation mutants and the released-registry mutant are caught by their dedicated checks.
Not covered: three shared-parser inputs, full repository Go gate, emitted-JavaScript parity and whole-source cohere CLI lint of .a. This branch is rebased and revalidated for supported inputs, but not fully green or landing-ready for all claimed controls.

The work-in-progress cap took precedence over new selection. This is the only
branch pushed in this session. A clean rebase first targeted e011f8f6; main then
advanced with devirtualization/compiler call-target changes, so it was rebased
again onto e8ba3d5d81de4d3773c723914fccd4c76248b965 and the native gates repeated.
Final remote-head checks retain that main SHA and the old wave-13 lease. The
requested rebase is published using --force-with-lease pinned to the old remote
SHA, preserving concurrent changes. No PR or additional claim was made.

Only this rule directory's report and evidence changed in this landing unit.
The compiler changes are inherited from main; no protected compiler file,
shared harness, registration generator or parser was edited. Rule and bridge
sources did not need changes. Go production rules and submodule pins are
unchanged, so the independent Go oracle binaries rebuilt in the first landing
pass are valid in the final pass too. All native suites and mutants were rebuilt
with the final-main compiler. Question-guard mutations ran before the second
rebase; their entire bridge sources are unchanged by the intervening main merge.

## Exact outputs and blockers

| Suite | Controls | Findings | Equal bytes | Normal / ASan |
| --- | ---: | ---: | ---: | --- |
| Original trio | 51 roots | 32 | 19416 | equal / equal |
| Nexus trio | 129 programs, 254 roots | 166 | 120897 | equal / equal |
| Core trio supported | 430 programs | 289, 171 suggestions | 162621 | equal / equal |
| Core trio refused | 3 programs | Go 3 | Go 2009 | native 70 / 70 |

The original control byte count changed with the scratch directory prefix; both
sides use the same new paths. Finding content and counts remain unchanged.
Complete findings, fixes, suggestions, spans, ordering and replacement text were
compared; original and Nexus payload details remain in their earlier reports.
Final normal and sanitized per-control result JSONs are identical.

All suites compare the same frozen compiler (77 roots) and repository (287 roots)
manifests. The original trio emits 4 findings / 7087 bytes on the compiler and
1 finding / 19144 bytes on the repository. Both later trios emit zero findings,
4933 and 18485 bytes respectively. Every corpus agrees in normal and sanitized
runs. Positive controls and mutants prevent zero findings from masking empty
implementations. The bridge also compares 1600 positions across four compiler
files: 54982 bytes agree under ASan/UBSan/LSan.

The core validation command still exits 1 with `cases 433 failures 3`. Go exits
0 for each; native exits 70 before any rule runs, with zero stdout:

| Case | Shape | Refusal |
| --- | --- | --- |
| 53d1c0a9ffc2fefa | `<foo />` then Object() in .ts | GreaterThanToken expected, SlashToken at 5 |
| aff6ced2ca61fe89 | `<foo></foo>` then Object() in .ts | GreaterThanToken expected, Identifier at 7 |
| defcb4c4ce6a2921 | yield label and break yield | semicolon expected at 37 |

These gaps live in the shared TypeScript parser. Ahra explicitly prohibited
shared-file repairs for this unit. No source rewriting or relaxed oracle was
introduced. No additional rule claim is permitted while this claim is partial.
Even the three refusal paths have no sanitizer report. BLOCKED.md gives the
prior branch investigation; final raw refusals are retained in this evidence.

## Every mutant and catcher

Every semantic mutant below compiles, exits 0 and has empty stderr. Only a byte
comparison with the unchanged production Go oracle catches it:

| Mutant | First different byte |
| --- | ---: |
| unassigned | 1372 |
| caught | 4226 |
| exports | 15301 |
| export-chain-flags | 15019 |
| export-module-lookup | 17519 |
| process-state | 134 |
| race-handle | 118 |
| blocking-order | 1690 |
| platform-declaration | 134 |
| resolved-return | 120 |
| graph-reachability | 134 |
| module-resolution | 1412 |
| namespace-alias | 72 |
| literal-parentheses | 587 |
| executor-span | 101 |
| read-shorthand | 72 |

platform_symbol, call_declaration, program_modules, output_flow and source_context
suffix-guard mutants each fail TestOutputCheckerQuestions. The read-symbol suffix
and wrong-kind mutants each fail TestReadSymbolQuestion. Each exits 1 with the
named test failure; retained logs document the accepted-invalid-input assertion.
The released-registry mutant exits 0 instead of the required 70 and is caught by
the lifetime assertion. Unmutated original, source-context and read-symbol
released-handle probes exit 70 with `invalid or released checker handle` and
no sanitizer report.

Seven foundation mutants were repeated on the final base: input length and
output length off by one are caught by ASan heap-buffer-overflow; retaining a
released handle fails the stale-handle assertion; requesting a type at a source
file position differs at byte 6; removing link opt-in fails the refusal check;
omitting the C output free and allocating a region result on the heap are each
caught by LeakSanitizer. Unmutated C ABI tests cover 100 queries, Unicode output,
results surviving release, stale and zero handles, and distinct second handles.

## Commands, toolchain and timing

Fresh `bash cloud/setup.sh` reports Go 0s, clang 0s, Node 0s, submodules 0s,
cache warm 200s, total 200s. `nproc` is 5; cpu.max is 400000/100000, memory 17.6 GB.
Sourced /workspace/adamic-tools/env.sh. All command output went directly to logs.

Final gate commands:

```
ADAMIC_WAVE13_STAGE0=/workspace/wave13-landing-final-adamic ADAMIC_WAVE13_ARCHIVE=/workspace/wave13-landing-checker.a ADAMIC_WAVE13_ARTIFACTS=/workspace/wave13-landing-final-original ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave13-corpus ADAMIC_WAVE13_COMPILER_MANIFEST=/workspace/wave-13-compiler.manifest ADAMIC_WAVE13_REPOSITORY_MANIFEST=/workspace/wave-13-repository.manifest go test ./stage1/cohere/typeaware -run '^TestWave13AgreementAndMutants$' -count=1 -v -timeout=30m
ADAMIC_TSGO_CORPUS=/workspace/wave13-corpus go test ./bridge/tsgo/... -count=1 -timeout=15m -v
go test ./internal/oracle -run 'TestTheOracleCatchesOneByte|TestNativeAgreesWithNode/internal/oracle/testdata/(closures|method_closures|generic_functions|regions|regions_throw|devirtualize|call_targets_closure|call_targets_element|call_targets_region|call_targets_reuse|call_targets_sort)\.a$' -count=1 -timeout=10m -v
go vet ./...
```

Built stage0 with `go build ./cmd/adamic`; normal and Go ASan C archives with
`go build [-asan] -buildmode=c-archive ./bridge/tsgo/archive`. Built both native
continuation suite.a files with `adamic build --tsgo`, adding `--sanitize` for
sanitized runs. Each directory's validate.py/corpora.py, mutants.py,
question_mutants.py and benchmark.py were run with final binaries and retained
controls. Existing reports give their exact positional arguments; landing log
names and stream archives retain every result and production-oracle hash.
Benchmarks use three alternating full-output rounds after all builds and tests.

| Suite / population | Native seconds | Go seconds | Native / Go |
| --- | ---: | ---: | ---: |
| original / compiler | 8.503342 | 2.363765 | 3.60x |
| original / repository | 0.420009 | 0.132343 | 3.17x |
| next / compiler | 3.034379 | 0.547875 | 5.54x |
| next / repository | 0.472465 | 0.210629 | 2.24x |
| more / compiler | 3.478860 | 0.421988 | 8.24x |
| more / repository | 0.477146 | 0.122749 | 3.89x |

Native is slower on every measured population. These are isolated medians,
including checker loading; concurrent gate timings are not speed claims.

Gofmt and source-only `git diff --check` pass. Whole-history whitespace checking
reports spaces at the ends of preserved `void ` suggestion text in oracle stdout.
Those bytes are required evidence and were retained, with both check logs.
The full repository Go test gate and emitted-JavaScript parity were not run;
this unit used its touched packages and the filtered external Node oracle.
Configured-globals surfaces and cohere CLI discovery of .a remain outside the
recorded controls. No whole-source CLI lint pass is claimed.

validation/landing contains summaries, complete normal/sanitized per-control
results, corpus hashes, mutant results, toolchain/build/gate logs, isolated
measurements and a gzip archive of raw streams and mutant evidence. sha256.json
checks every retained top-level artifact. Earlier reports/evidence are historical;
this report supersedes their main-base and timing claims for this landing attempt.
