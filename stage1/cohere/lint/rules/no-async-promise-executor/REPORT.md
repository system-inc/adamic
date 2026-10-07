Built: three more owned .a ports, no-async-promise-executor, no-case-declarations and no-compare-neg-zero; complete suggestion records preserved.
Commits: all prior claims pushed at 6c1993fa; new pre-code claim b9a7116d; each new rule committed separately on codex/lint-wave1-03.
Commands and outputs: 113 comparable cases and 176 compiler/stage1 files per rule match Go on Node, emitted JavaScript and ASan/UBSan native; registry/vet PASS; setup 19s, nproc 5.
Mutants: async_executor_ignored, closing_brace_missing and negative_zero_ignored compile/run, then are caught solely by Go output comparison on all three backends.
Not covered: installed .a registration, installed multi-edit suggestion serialization, one invalid decorated executor recovery case, prior JSX gaps and full repository gate.

## Claim ordering

Every previous claimed rule now has owned implementation code, independent
comparison evidence and a compiling semantic mutant. All of that was pushed
through 6c1993fa before fetching every origin head again. Main remained
ef3d907ecdc4c771b016f7d9c52372def057a340. The original 46 helper-ready names
were all claimed. A search of 52 distinct claim Markdown blobs across origin
heads, plus main's lint implementations, selected the first available names
from the inventory's syntax ready for AST/API adaptation queue:

1. no-async-promise-executor
2. no-case-declarations
3. no-compare-neg-zero

The refreshed search skipped names other workers had claimed during the prior
validation. evidence/selection.json preserves the complete ref snapshot and
available queue. Claim b9a7116d was pushed before creating any of these rule files.
No additional batch was taken after these three.

## Implementation

No-async-promise-executor unwraps parentheses on callee and argument zero, matches
exactly the Promise identifier, handles generic type arguments through the
parser's argument count, and reports the async modifier's own token range. It
checks no other argument and no nested function body. Async generators are
included; calls without new, member callees and differently named callees decline.

No-case-declarations visits direct clause statements only. Ambient declarations,
var, type/interface/enum/module declarations and nested lexical declarations are
exempt; block-scoped let/const/using/await-using and concrete class/function
statements report. Each finding retains the suggestion ID, exact message and
both edits wrapping the entire clause body, not just the reported declaration.
Suggestions are never automatically applied.

No-compare-neg-zero preserves all eight comparison operators, unwraps parentheses
at both the negation and literal levels, matches the canonical numeric value 0,
and reports once against the full comparison. BigInt, assertions, arithmetic,
assignment and Object.is decline. The operator spelling is rendered in the
message exactly as Go does. No fix or suggestion is offered.

## Shared boundaries and parser recovery

Existing records.a in our owned type-constraint rule directory carries complete
repairs. No-case-declarations' registered finish hook explicitly refuses the
unsupported shared suggestion shape. The owned driver reads full records and
compares both edits independently. No finding is silently exposed with a lost
repair. Normal registration still needs .a support; the announced harness branch
was absent from fetched heads. All shared source files remain unchanged.

Go's upstream test API examines recovered invalid syntax. One captured executor
case is new Promise(@dec async () => {}), on which the actual Go rule is silent
because recovery gives the executor a decorator modifier, not an async modifier.
The production Go fixer correctly refuses this input before applying a fix.
Node source, emitted JavaScript and sanitized native each explicitly refuse it
with exit 70 and parser slice unsupported primary AtToken at 12. Raw Go and
three-backend observations are saved as gap-13 logs; evidence/gaps.json lists it.
It is not counted as matching parity and remains a shared-parser recovery gap.

The owned oracle allows recovered parse errors only for captured upstream test
rows carrying an explicit recovery field. If such a row has parse errors, it
asserts there are no automatic fixes and preserves the original source, rather
than falsely claiming the production fixer converged. This also permits the
upstream invalid -0_0 and -00 literal cases to be compared through the actual
recovered Go AST; both match Adamic. Valid test inputs, owned witnesses and every
compiler/stage1 row retain strict parsing and the unchanged converging Go fixer.
Rule bodies, messages and reporting remain the independent Go implementation.

## Commands and evidence

Go cohere pin 715ba94f3608a6500086b1076ce5cb7e51b836db; TypeScript compiler
050880ce59e30b356b686bd3144efe24f875ebc8. Source /workspace/adamic-tools/env.sh.
Go 1.27.1, Node 24.19.0, clang 20.1.8. nproc 5, quota 400000/100000.
Setup used the existing scratch .a compatibility overlay: Go ready 0s, clang 1s,
Node 1s, submodules 1s, cache warm 19s, done 19s on 5 processors, 17.6 GB.
The setup log is retained in the preceding six-rule report's evidence directory.

From the repository root:

```
python3 stage1/cohere/lint/rules/no-async-promise-executor/validate.py \
  --scratch /tmp/lint-wave1-03-next-final --typescript /tmp/lint-wave1-03-typescript --mutants --benchmark
```

PASS: 24 executor, 31 case-declaration and 56 negative-zero upstream cases
captured. The one explicit parser gap leaves 110 comparable upstream cases;
three independent owned witnesses bring the total to 113. All three paths match
33,108 bytes of formatted findings, byte ranges, message IDs, complete suggestion
IDs/messages/edits and fixed source.

Corpus PASS: 176 files per rule, comprising all 77 compiler files and 99 then-
present stage1 .ts/.a programs, excluding generated registries and deliberately
invalid gaps. All three paths match 36,843,665 bytes. Compressed canonical Go
output and each independent output's SHA-256/size are retained in evidence.

The first comparison attempt stopped at Go's strict parse refusal on the
invalid decorator fixture. The next observed the production fixer's explicit
refusal. The final scoped recovery handling above follows the upstream test API
and proves the corresponding Adamic refusal. Neither failure is a semantic mutant.

```
go test -overlay=/tmp/lint-wave1-03-corpora/overlay.json ./stage1/cohere/lint/registry -count=1 -v
go vet -overlay=/tmp/lint-wave1-03-corpora/overlay.json ./...
```

Registry PASS 0.103s, including metadata rejection probes. Vet exits 0 with empty
output. All tests log directly to files, never pipes. Native comparisons and
mutants use ASan/UBSan and normal Linux leak checking. A successful process must
exit 0 with empty stderr; compilation errors or sanitizer failures cannot count
as caught semantic mutants. Shared integration and the full gate are not certified.

## Mutants and observed rates

| Rule | Mutation caught on all three paths | Native/s | Node/s | Go/s | Findings |
| --- | --- | ---: | ---: | ---: | ---: |
| no-async-promise-executor | async_executor_ignored | 884.51 | 1139.25 | 5522.08 | 1000 |
| no-case-declarations | closing_brace_missing | 1116.48 | 1401.63 | 7005.52 | 1192 |
| no-compare-neg-zero | negative_zero_ignored | 960.60 | 1182.54 | 5889.19 | 1000 |

Every mutant compiles and runs successfully before output comparison catches it.
The case-declaration mutant changes only the closing-brace edit in a suggestion;
its findings and message remain identical, proving that repair comparison matters.
Rates are best of three interleaved rounds, startup and parsing included, over
77 compiler files plus 1,000 planted violations. Native is unsanitized for timing
only. All counts agree in every round. These are worker observations, not a claim
native beats Go.
