Built: isolated blocker reproduction and Go oracle; no third-batch rule port is complete.
Commits: six completed ports through 98f1cd3; third-batch reservation f62601fe; this report follows.
Commands and outputs: Go oracle exits 0 with 1 finding; existing native suite exits 70 parsing valid JSX.
Mutants: none for this unimplemented batch; all six earlier rule mutants passed their comparison checks.
Not covered: the three reserved React rules, their parity, fixes, suggestions, sanitizers, handles and timing.

# Third batch blocked at native JSX parsing

Reserved rules are react-hooks/globals, react-hooks/immutability and
react-hooks/no-deriving-state-in-effects, each zero in both ranked corpora.
The final refresh skipped newly claimed prefer-promise-reject-errors,
prefer-regex-literals, prefer-rest-params and react-hooks/exhaustive-deps.
The reservation was pushed before any implementation.

## Observation

The input in testdata/parser_blocker.a is the moduleLetWrittenInComponent
positive control in cohere/internal/lint/rules/react/globals_test.go.
The reproduction copies it byte for byte to Subject.tsx, because the independent
TypeScript parser must select JSX syntax. The rule source remains untouched.
The local Go overlay selects the three reserved production rules, resolving
registry registration normally. It reports one globals finding at bytes 36..37,
with no fixes or suggestions, and exits 0.

The native executable is the completed previous batch's suite, built from
98f1cd3. That suite invokes the same stage1/typescript/parser/Parser before any
rule runs. On the identical file it exits 70 with empty stdout and:

```
adamic: panic: parser slice expected GreaterThanToken, got SlashToken at 57 in /workspace/wave-01-third-blocker/Subject.tsx
```

This proves a parser blocker independently of any unfinished rule logic.
It is not evidence of a lint verdict from the previous suite.
Commands, exit statuses and exact output are in validation/records.json.
The oracle build succeeds with empty build logs. Go's observed load time is
39.436 ms and run time is 0.227 ms on this one control; there is no meaningful
native-versus-Go rule timing because native never reaches a rule.

## Dependency inspection and scope

stage1/typescript/parser has no JSX implementation on this branch. The three
Go rules' positive controls include JSX and their component gate explicitly
walks JSX nodes. Replacing JSX with a hook call would change the control and
would not establish byte agreement on the original corpus.

Two further dependencies are visible in Go source but have not been run in
native code: immutability requires ForFunction (Lower plus Construct/SSA) and
AsCompilationUnit; no-deriving-state-in-effects additionally requires
ForFunctionWithoutManualMemoization (memo erasure and callback inlining).
The bridge's existing syntax-flow-graph is a different representation: it does
not carry HIR instruction values, SSA phis or nested capture/context identity.
No rule verdict should be moved to Go to conceal these missing native passes.

Ahra's instruction says to keep changes in our own rule directories and stop
on other blockers. The parser is shared and outside those directories, so this
batch stops here. No shared file was changed, no native rule stub was added,
and no parity or mutant result is claimed for the three reservations.
They remain reserved for continuation when JSX parser support is available.
The previous six completed ports remain pushed, with reports in
../WAVE_01_REPORT.md and ../wave_01_next/REPORT.md.

## Reproduction

From the repository, source /workspace/adamic-tools/env.sh. Create a temporary
Subject.tsx by copying testdata/parser_blocker.a and a tsconfig enabling JSX
preserve, then create a manifest containing the absolute Subject.tsx path.
Build testdata/oracle.go using a Go overlay mapping a virtual file directly
under cohere to this oracle, with cohere as the build working directory.
Run that oracle and /workspace/wave-01-next-final/native with the same tsconfig
and manifest. The exact session commands are retained in validation/records.json.

## Origin refresh on continuation

All origin heads were fetched again. No further rules were claimed.

```
origin/main e011f8f60899586d6373a5ccb07335ad82cfbf3c
origin/codex/tsgo-c-library 5afbdb83da2ed7ad9815657cd3f6ececd5294bf6
origin/codex/lint-harness-dot-a f4d98cab50048692781da3599131317dc569d466
origin/codex/stage1-jsx-lint a8a62d62ca49db7415e14c3887dd305022b17309
```

JSX support is now published on origin/codex/stage1-jsx-lint. Its
implementation commit is e715ef4a2f898230af63c40195dea6586a557899, documented
in stage1/typescript/parser/JSX_REPORT.md on that branch. It changes the shared
parser and scanner; it has not reached this branch, main, or tsgo-c-library.
The earlier absence statement refers to this branch, not every origin branch.

The current scope instruction forbids changing shared files and requires stopping
on other blockers. Integrating these parser/scanner changes therefore remains
outside this unit. The three rules remain reserved and unported. No new test,
mutant, sanitizer or performance result is claimed on this continuation.

## Landing readiness

Rebased without conflicts onto origin/main e011f8f60899586d6373a5ccb07335ad82cfbf3c.
The tested rebased code head is 9a4da8944516e6f707e1e9afae724d55ca9de63d.
Original gate: TestWave01AgreementAndMutants PASS 147.095s.
Continuation verify.py: PASS all three claimed rules, including checker and
bridge package tests. Both gates compare controls and compiler/repository
corpora with Go, run sanitizer comparisons, kill all six rule mutants and
exercise seven released-handle questions with registry-retention mutants.
Logs and continuation command records are in validation/landing.
No new React implementation is claimed; JSX integration remains outside scope.

Main advanced during the first gate. Rebased again onto e8ba3d5d81de4d3773c723914fccd4c76248b965
and reran both full gates on code head 76493d0bce37f2d70b4001cf5556a9e461874e2f.
Original gate PASS 139.718s; continuation gate PASS, including package tests.
Final logs and command records are in validation/landing.

## Numeric listener declarations

The six completed rules now expose readonly listenerKinds arrays of numeric
typescript-go SyntaxKind values. The values were read from the independent
Go enum; listener_check.py compares each array with the production Go rule
registration and rejects an in-memory +1 kind mutant for each rule.
The current driver does not consume these declarations.

Both full gates passed again after the declarations: original 140.235s;
continuation PASS including checker and bridge package tests. All six rule
mutants and seven released-handle checks passed, with exact compiler/repository
comparisons under sanitizers. Evidence is in validation/listeners.

This is metadata preparation, not completion of the speed rule. ParseNode.kind
is still a string in the shared parser, and the existing rule routines still
use string-kind helpers and per-rule traversal. A numeric parser field and
shared node delivery are needed to remove those paths. No speed improvement
is claimed; the shared parser/driver remains outside this worker scope.
