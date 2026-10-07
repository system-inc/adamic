Rebased this unit onto fetched origin/main e8ba3d5d and fixed owned JSX compatibility.
Tested implementation commit: 69db2dc85c745b3de48ca4629965e6bcf73d6660.
Five oracle suites PASS, 421.255s; full bridge PASS, 69.592s; vet/format clean.
All fifteen rule mutants caught by independent Go bytes; release and foundation checks pass.
Three React claims remain incomplete; shared Nexus dispatch integration remains pending.

# Wave 16 landing validation

The only branch this unit pushed is codex/typeaware-wave-16. Its previous remote
tip was b03c2289910b0e477dd6fc01f38123285a11686d. The first rebase onto e011f8f6
completed without conflicts. It exposed a nominal Scanner mismatch in the owned
JSX parser and follow-up runner. That failure is retained in
validation-wave16-landing/earlier-base-failure.log. The other four suites passed;
the repaired follow-up passed separately in 94.064s on that earlier base.

A pre-push fetch then found main had advanced with devirtualization and runtime
memory analysis changes. The branch was rebased again, without conflicts, onto
e8ba3d5d81de4d3773c723914fccd4c76248b965. All five suites and the full bridge gate
were rerun together against the compiler on that base. The source commit above
contains the compatibility fix; the following evidence commit changes no code.
The push uses an exact lease on the previous remote tip because the user
explicitly requested rebasing and pushing the published branch.

## Owned compatibility change

The isolated JSX Scanner and the shared Scanner are different nominal classes.
Shared Statements required the latter, while the JSX parser supplied the former.
The owned wave16_jsx_statements.a copies statement descent from main e011f8f6,
changing its scanner import to the owned JSX class. That source is unchanged by
the subsequent devirtualization merge. The owned JSX parser imports this module.
The follow-up runner copies the parsed node arena and expression roots into a
real shared Parser instance, and uses a real shared Scanner for shared analyses.
No cast asserts structural ancestry. The comparisons and sanitizers below cover
the adapter and JSX parsing. No shared generator, harness or protected compiler
file was edited; main's compiler changes arrived through the rebase.

## Commands and results

The retained environment was sourced from /workspace/adamic-tools/env.sh.
Original setup timing remains Go 0s, clang 1s, Node 1s, submodules 1s, cache 85s,
done 85s. nproc printed 5. Every test output went directly to a log.

```
ADAMIC_WAVE16_ARTIFACTS=/workspace/wave16-artifacts/rebase-current/first \
ADAMIC_WAVE16_FOLLOWUP_ARTIFACTS=/workspace/wave16-artifacts/rebase-current/followup \
ADAMIC_WAVE16_THIRD_ARTIFACTS=/workspace/wave16-artifacts/rebase-current/third \
ADAMIC_WAVE16_FOURTH_ARTIFACTS=/workspace/wave16-artifacts/rebase-current/fourth \
ADAMIC_WAVE16_FIFTH_ARTIFACTS=/workspace/wave16-artifacts/rebase-current/fifth \
ADAMIC_WAVE16_COMPILER_MANIFEST=/workspace/wave16-artifacts/compiler.manifest \
ADAMIC_WAVE16_REPOSITORY_MANIFEST=/workspace/wave16-artifacts/repository.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave16-corpus/typescript \
go test ./stage1/cohere/typeaware -run '^TestWave16.*AgreementAndMutants$' \
 -count=1 -v -timeout 20m > /workspace/wave16-artifacts/rebase-current/oracles.log 2>&1
go test ./bridge/tsgo/... -count=1 -v -timeout 15m \
 > /workspace/wave16-artifacts/rebase-current/bridge.log 2>&1
go vet ./stage1/cohere/typeaware ./bridge/tsgo/... \
 > /workspace/wave16-artifacts/rebase-current/vet.log 2>&1
gofmt -l stage1/cohere/typeaware/wave16*test.go \
 stage1/cohere/typeaware/testdata/oracle_wave16*.go \
 > /workspace/wave16-artifacts/rebase-current/gofmt.log
```

The oracle package exits zero, PASS, 421.255s. Individual durations:

| Suite | Seconds |
| --- | ---: |
| TestWave16FifthAgreementAndMutants | 102.26 |
| TestWave16FollowupAgreementAndMutants | 99.07 |
| TestWave16FourthAgreementAndMutants | 62.17 |
| TestWave16AgreementAndMutants | 76.67 |
| TestWave16ThirdAgreementAndMutants | 81.08 |

Normal and ASan/UBSan/LeakSanitizer runs match full production Go findings,
IDs, messages, spans, ordered fixes and suggestions. The frozen repository
population has 287 roots and the compiler population 77 roots.

| Set | Control findings | Repository findings | Compiler findings |
| --- | ---: | ---: | ---: |
| Original | 65 | 2 | 1 |
| Follow-up | 22 | 0 | 0 |
| Third | 50 | 0 | 0 |
| Fourth | 33 | 0 | 0 |
| Fifth | 231 | 0 | 0 |

The fifth set also matches both native modes under its other three option
profiles: 225, 243 and 235 findings. Stream sizes differ from older reports
because canonical file headers include the artifact directory's absolute path.
Each current native stream is compared with Go for exactly the same paths.

All fifteen rule mutants compile and exit zero with empty stderr; the byte
comparison alone catches them:

| Mutant | First differing byte | Mutant findings |
| --- | ---: | ---: |
| throw-range | 134 | 231 |
| backreference-range | 9299 | 231 |
| arrow-fix | 6066 | 231 |
| listener-general | 727 | 23 |
| render-number | 7193 | 14 |
| mock-range | 3594 | 22 |
| function-range | 136 | 33 |
| native-range | 4604 | 33 |
| wrapper-range | 5191 | 33 |
| import-write | 68 | 59 |
| exponent-base | 15696 | 65 |
| nan-comma | 12091 | 63 |
| shell-heredoc | 8313 | 53 |
| sql-string | 11806 | 31 |
| alert-range | 134 | 50 |

Every set also requires a released checker handle to exit 70; each retained
registry mutant instead exits zero and is caught. The full bridge gate passes
100 C ABI queries and 162 positions / 3261 identical bytes under sanitizers.
Its seven foundation mutants are caught: input/output lengths by ASan, retained
stale handle by assertion, source-file position by Go byte 6, removed linking
guard by refusal, and missing output free / heap region allocation by LSan.
The independent raw checker question tests pass. Vet and formatting logs are
empty. Native output, stderr, controls, configs and source/stream hashes are
preserved in validation-wave16-landing; binary archives are excluded.

## Remaining work

The Nexus follow-up and third sets still test the supplied shared dispatch
registration patch through a Go build overlay. Ordinary shared dispatch
integration is pending; these results do not assert it has landed. The new
React claims, set-state-in-effect, set-state-in-render and static-components,
remain blocked before implementation by missing native React HIR lowering,
SSA, capture correspondence and graph analyses. Their dependency report is
WAVE16_SIXTH_REPORT.md. None is represented as a completed rule or a passing
native test. No further rules were claimed.

No full repository Go gate, emitted-JavaScript rule comparison, runtime execution
of repaired programs or fresh rule runtime benchmark was run in this landing
pass. Prior benchmark results remain historical measurements in the five rule
reports. Sanitizers instrument native/C code, not the Go heap.

## Landing cap and dependency recheck

A subsequent all-head fetch still found main at e8ba3d5d. This branch's
published tip was b50cca14, already descended from that main. The only branch
this unit has pushed is codex/typeaware-wave-16. Comparing implementation
Go, .a and .ts sources against the tested 69db2dc8 found no changes outside
stored validation inputs. The passing 421.255s oracle and 69.592s bridge runs
therefore remain applicable; no unchanged test suite was rerun.

The refreshed prerequisite audit covers 465 origin heads and finds zero
native declarations of the five known entry points. Its JSON is preserved in
validation-wave16-landing/dependency-refresh.json. This remains a name-based
inventory, not an exhaustive proof about differently named implementations.
Wave 04's a8c569a2 adds refs lattice joins/destructuring kernels, but its current
report explicitly still lacks source HIR/SSA, capture transfer and sweep.
Wave 08's 242946b4 records its own rebase and gates, without providing the
missing frontend. The bridge branch's shared facts.go still has no call to
this unit's additionalAnswer dispatch hook. These partial kernels and harness
changes do not unblock the three claimed React source rules.

No new rule is claimed, no shared source is edited, and no push targets main
or an area branch. The exact remaining blockers and uncovered checks above
are unchanged. This evidence-only continuation stops under Ahra's instruction
to report prerequisites outside the owned rule directories rather than edit
shared files.

## React reference ambiguity continuation

Current main remains e8ba3d5d. Wave 21 now supplies prepared-HIR native cores
for the same three retained names, while its source adapter remains absent.
This unit independently reproduces a production static-components message
ambiguity: 64 unchanged Go runs emit createA 59 times and createB 5 times,
while a same-creator control is byte-identical across all 64 runs. The sole
canonical difference is byte 203. No output is normalized. See
WAVE16_REACT_REFERENCE_REPORT.md and validation-wave16-react-reference for
all outputs, exact sources, the independent runner and the positive-control
mutation. The completed fifteen native implementations remain unchanged;
no further claim or shared source edit was made.
