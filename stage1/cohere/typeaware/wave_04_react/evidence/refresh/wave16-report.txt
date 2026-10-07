# Wave 16, throw literals, backreferences and arrow callbacks

The previous twelve ports and validation were pushed at 363fde32 before this
selection. All origin heads were fetched again. The scan covered 197 ranked
checker-dependent rules and 33 Markdown claim documents, excluding integration
head ports and 126 claimed names. Of 46 remaining candidates the first three
were no-throw-literal, no-useless-backreference and prefer-arrow-callback, all
with zero combined compiler/repository count. Whole stage1/cohere searches on
main and the bridge found only inventory/count mentions. Claim 72eff2c5 was
pushed before code. No additional rules were claimed.

Main was ef3d907ecdc4c771b016f7d9c52372def057a340; bridge was
5afbdb83da2ed7ad9815657cd3f6ececd5294bf6. Cohere remains pinned to
715ba94f3608a6500086b1076ce5cb7e51b836db, its TypeScript shim to
8d550c837c90bd1805b047b7eeccc2baac2d5e7a, and the external compiler corpus to
050880ce59e30b356b686bd3144efe24f875ebc8.

## Ports and isolation

Each rule has its own .a file. Throw-literal reproduces the optimistic syntactic
couldBeError recursion and global undefined discrimination. Parentheses are
not skipped in that recursion because production Go cohere does not skip them.
Backreference owns the enclosing-path scanner, Unicode-set class/escape stepping,
constant binding evaluation, alias/global-object reference tracker, duplicate
named-group reachability and ordered five-problem classification. Arrow-callback
owns function frames, arguments/this/super/new.target ownership, checker-resolved
self references, callback/bind recognition and ordered fixes with every production
withholding condition. Its constructor exposes both option flags.

The isolated helper modules use only existing binding-declarations,
symbol-identities and resolved-name bridge facts. No new checker question or
shared registration edit was needed. These three build with an ordinary bridge
archive without an integration overlay. The shared registration generator,
shared test harness and protected compiler files remain untouched. Earlier
Nexus ports still have the dispatch dependency documented in their reports;
that gap does not block this set.

The independent Go oracle imports no bridge implementation and calls unmodified
production registry rules. Canonical bytes include findings, byte spans, IDs,
messages, every ordered fix and suggestion, preserving duplicates. Arrow fixes
are compared in full, including whitespace, removals, insertions and grouping.
The other two offer no fixes; none offers suggestions, whose zero counts are
also compared. Four callback option combinations use the production settings
struct in the independent oracle, with no shared profile compiler edit.

## Commands and results

The retained environment uses /workspace/adamic-tools/env.sh. Original setup
reported Go 0s, clang 1s, Node 1s, submodules 1s, cache 85s, done 85s; nproc again
printed 5. Every test output was sent directly to a log.

```
ADAMIC_WAVE16_FIFTH_ARTIFACTS=/workspace/wave16-artifacts/fifth-final \
ADAMIC_WAVE16_COMPILER_MANIFEST=/workspace/wave16-artifacts/compiler.manifest \
ADAMIC_WAVE16_REPOSITORY_MANIFEST=/workspace/wave16-artifacts/repository.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave16-corpus/typescript \
go test ./stage1/cohere/typeaware -run '^TestWave16FifthAgreementAndMutants$' \
-count=1 -v -timeout 15m > /workspace/wave16-artifacts/fifth-final2-test.log 2>&1
```

PASS, 62.826s. Both ordinary and ASan/UBSan/LeakSanitizer native runs have empty
stderr and match complete Go bytes:

| Population | Findings | Identical bytes |
| --- | ---: | ---: |
| Controls, 347 subjects plus helper and declaration root | 231 | 87880 |
| Repository, frozen 287 roots | 0 | 18485 |
| TypeScript src/compiler, 77 roots | 0 | 5780 |

The controls include 42 throw-literal, 188 backreference and 101 default callback
rows from the pinned ESLint corpus JSON, plus sixteen additional subjects.
Two backreference rows explicitly marked parse-skipped and ten callback rows
carrying options were not imported into the default control list. All control
sources are preserved, with each rule required to have a positive finding.
Additional cases cover throw operator families, parentheses, shadowed undefined,
self-reference shadows, owned and nested arguments, all fix withholding shapes,
comments, async newlines, TypeScript this parameters, Unicode byte spans, mutable
and effectively constant patterns/flags, alias/default/destructuring/conditional/
cast tracking, modified globals, malformed patterns and unknown flag expressions.

The same controls are compared normally and under sanitizers for all remaining
callback profiles:

| allowNamedFunctions | allowUnboundThis | Findings |
| --- | --- | ---: |
| false | true, default | 231 |
| true | true | 225 |
| false | false | 243 |
| true | false | 235 |

Each rule mutant compiles and exits 0 with empty stderr, retaining all 231
findings. Only the independent Go bytes catch it:

| Mutant | Change | First differing byte |
| --- | --- | ---: |
| throw-range | Add one to throw report end | 116 |
| backreference-range | Add one to regex report end | 9227 |
| arrow-fix | Insert an extra space after the arrow | 6021 |

A count-only check cannot catch any of these. The arrow mutant additionally
keeps finding IDs, messages and spans unchanged; its fix bytes alone fail.
Querying a released program exits 70 with invalid or released checker handle.
Keeping the released program in the registry makes that mutant exit 0 and is
caught by the required refusal.

```
go test ./bridge/tsgo/... -count=1 -v -timeout 15m \
 > /workspace/wave16-artifacts/fifth-bridge.log 2>&1
go vet ./stage1/cohere/typeaware ./bridge/tsgo/... \
 > /workspace/wave16-artifacts/fifth-vet.log 2>&1
gofmt -l stage1/cohere/typeaware/wave16_fifth_test.go \
 stage1/cohere/typeaware/testdata/oracle_wave16_fifth.go \
 > /workspace/wave16-artifacts/fifth-gofmt.log
```

PASS, bridge 62.771s and checker 0.229s. Vet and formatting logs are empty.
Foundation again holds 100 C ABI queries and 162 positions / 3261 exact bytes
under sanitizers. Seven established mutants are caught: input and output lengths
by ASan, retained stale handle by its assertion, wrong source-file position by
Go byte 6, removed linking guard by refusal, missing output free and heap region
allocation by LeakSanitizer. Existing independent raw checker fact tests pass.
No new raw question was introduced.

Initial failure is preserved in fifth-test.log: an absent regex flags argument
was passed through skip(-1), causing a missing-node panic. The sentinel is now
checked before parser access. The corrected 341-control run passed in 64.892s;
the expanded formatted run passed in 79.562s; adding production's exclusion of
declaration-file-only constant bindings was then retested on the final sources
in the 62.826s run above. New .a modules were formatted by pinned cohere's stdin
interface using virtual .ts paths, without writing temporary .ts source files.

## Native versus Go

Three quiet alternating rounds use fifth-benchmark.py. Every round has the same
stream hash for its corpus. No tests or builds ran concurrently with timing.
Median seconds on the final sources:

| Population | Runtime | Load | Run | Process |
| --- | --- | ---: | ---: | ---: |
| Compiler | Native | 0.244632 | 2.026501 | 2.291707 |
| Compiler | Go | 0.235482 | 0.100029 | 0.353815 |
| Repository | Native | 0.061993 | 0.211119 | 0.277790 |
| Repository | Go | 0.061714 | 0.043360 | 0.117193 |

Native process time is 6.48 times Go's on compiler and 2.37 times on repository.
Both corpus populations have zero findings, so these timings measure traversal
and checker cost. Exact controls, source/stream hashes, compressed stdout,
stderr, failure and passing logs, benchmark script and raw rounds are committed
in validation-wave16-fifth.

## Limits

No full repository Go gate, emitted-JavaScript differential gate, runtime execution
of repaired programs or exhaustive upstream configuration matrix was run. The
four option profiles cover controls; corpus runs use production defaults. ASan
instruments native/C code, not the Go heap. Earlier Nexus shared dispatch
integration remains separately pending, with its patch already supplied.
