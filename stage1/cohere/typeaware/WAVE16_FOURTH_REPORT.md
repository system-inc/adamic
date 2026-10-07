# Wave 16, native constructor rules

The prior nine ports and their validation were pushed at b5c30a08 before this
selection. All origin heads were fetched again. Of 197 ranked checker-dependent
rules, 55 remained after excluding ports on main/bridge and 117 names claimed
in 33 Markdown documents across every origin head. The next three were
no-new-func, no-new-native-nonconstructor and no-new-wrappers, all with zero
compiler/repository count. Whole stage1/cohere searches on the integration heads
found only inventory/count mentions. The claim was pushed in 031c25bd before code.
Main was ef3d907ecdc4c771b016f7d9c52372def057a340 and bridge was
5afbdb83da2ed7ad9815657cd3f6ececd5294bf6. No further rules were claimed.

Each rule has its own .a file. The isolated constructor_global.a decoder uses
existing resolved-name facts and faithfully tests the first declaration's
IsDeclarationFile flag, matching production resolvesToAGlobal. It does not use
an all-declarations predicate. No new checker question or registration edit is
needed. These three work with an ordinary bridge archive, without an integration
overlay. The shared generator, shared test harness and protected compiler files
remain untouched. The earlier six Nexus ports retain the dispatch dependency
recorded in WAVE16_FOLLOWUP_REPORT.md and WAVE16_THIRD_REPORT.md.

The independent Go oracle calls the unmodified production registry and does not
import bridge code. Complete canonical bytes include rule IDs, message IDs,
messages, byte spans, every fix and every suggestion, with duplicate findings
preserved. These three production rules have no fixes or suggestions; their zero
counts are included in every compared diagnostic.

## Commands and observations

The existing environment was retained: original cloud/setup.sh timing was Go 0s,
clang 1s, Node 1s, submodules 1s, cache 85s, done 85s. nproc again printed 5.
Commands source /workspace/adamic-tools/env.sh. Every test writes directly to logs.

```
ADAMIC_WAVE16_FOURTH_ARTIFACTS=/workspace/wave16-artifacts/fourth-final \
ADAMIC_WAVE16_COMPILER_MANIFEST=/workspace/wave16-artifacts/compiler.manifest \
ADAMIC_WAVE16_REPOSITORY_MANIFEST=/workspace/wave16-artifacts/repository.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave16-corpus/typescript \
go test ./stage1/cohere/typeaware -run '^TestWave16FourthAgreementAndMutants$' \
-count=1 -v -timeout 15m > /workspace/wave16-artifacts/fourth-final-test.log 2>&1
```

PASS, 56.684s. Both normal and ASan/UBSan/LeakSanitizer native comparisons
have empty stderr and exactly match Go:

| Population | Findings | Identical bytes |
| --- | ---: | ---: |
| Controls, 13 subjects plus helper and declaration root | 33 | 12914 |
| Repository, frozen 287 roots | 0 | 18485 |
| TypeScript src/compiler, 77 roots | 0 | 5780 |

Controls distinguish direct/indirect Function invocation, bind's inner call span,
static string/template/computed method names, optional chains, parenthesized
callees and receivers, no-argument new, imported/parameter/class/function/named
function-expression/hoisted/nested shadows, qualified and comma receivers,
nonconstructing methods, conversion calls, primitive wrappers and Unicode offsets.
Parenthesized Symbol/BigInt intentionally remain clean, unlike wrappers, because
that is what production Go cohere does. Each rule must have a positive control.

Every rule mutant compiled, exited 0 with empty stderr, and retained all 33
findings. Only complete Go byte comparison catches these changes:

| Mutant | Change | First different byte |
| --- | --- | ---: |
| function-range | Add one to Function report end | 118 |
| native-range | Report the new expression rather than its callee | 4577 |
| wrapper-range | Add one to wrapper report end | 5155 |

Released-handle query exits 70 with invalid or released checker handle. A mutant
retaining the released program exits 0 and is caught by that required refusal.

```
go test ./bridge/tsgo/... -count=1 -v -timeout 15m \
 > /workspace/wave16-artifacts/fourth-bridge.log 2>&1
go vet ./stage1/cohere/typeaware ./bridge/tsgo/... \
 > /workspace/wave16-artifacts/fourth-vet.log 2>&1
gofmt -l stage1/cohere/typeaware/wave16_fourth_test.go \
 stage1/cohere/typeaware/testdata/oracle_wave16_fourth.go \
 > /workspace/wave16-artifacts/fourth-gofmt.log
```

PASS: bridge 58.604s, checker 0.232s. Vet and formatting logs are empty.
Foundation again holds 100 C ABI queries and 162 positions / 3261 exact bytes
under sanitizers. Its seven mutants are caught: input/output length by ASan,
retained released handle by stale-handle assertion, wrong source-file position
by Go byte 6, removed linking guard by refusal test, missing output free and
heap region allocation by LeakSanitizer. Existing independent checker question
tests also pass. No new raw question was introduced in this set.
New .a modules were formatted through pinned cohere's stdin interface with a
virtual .ts path, then fully retested. No temporary .ts Adamic source was written.

## Native versus Go

Three quiet alternating rounds run by fourth-benchmark.py preserve identical
stream hashes in every round. No test/build ran concurrently with these timings.
Median seconds:

| Population | Runtime | Load | Run | Process |
| --- | --- | ---: | ---: | ---: |
| Compiler | Native | 0.244409 | 1.152390 | 1.412982 |
| Compiler | Go | 0.218934 | 0.044679 | 0.290774 |
| Repository | Native | 0.060827 | 0.135870 | 0.201183 |
| Repository | Go | 0.058995 | 0.042563 | 0.113164 |

Native process time is 4.86 times Go's for compiler and 1.78 times for repository.
Both corpora have zero findings; these measure traversal/checker cost. Pins and
frozen corpus populations remain those recorded in the earlier wave reports.

Exact logs, compressed diagnostic streams, controls, source/stream hashes and
raw timing rounds are committed under validation-wave16-fourth. The initial
unformatted 26-finding run also passed and its test log is retained; final source
hashes and streams correspond to the expanded, formatted 33-finding run.

## Scope limits

No full repository Go gate, emitted-JavaScript differential gate or exhaustive
upstream fixture matrix was run. Validation covers the pinned compiler corpus,
frozen repository population and the controls above. ASan instruments native/C
code, not the Go heap. Earlier Nexus dispatch integration remains a separately
documented shared gap; it does not block this constructor-rule set.
