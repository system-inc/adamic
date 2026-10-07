Rebased wave-16 onto the current lint area and revalidated the owned ports.
Commits: area d3a37422c, main b6b1538b0, rebased parent 8aa8904af; evidence in this commit.
Checks: fifteen-port oracle PASS 431.585s; new 60-command gate PASS; inherited typeof oracles PASS.
Mutants: 18 rule mutants caught by Go bytes, eight inherited typeof mutants caught by Node stdout.
Not covered: cross-file context integration, emitted JavaScript for rules and the full 17-check gate.

The clean rebase contains current main and origin/area/stage1-lint, including
the inherited typeof null/slot/constructor fixes. No shared implementation,
harness or skip condition was edited. Only codex/typeaware-wave-16 is pushed.
The baselines were confirmed unchanged after checks. All 287 pinned repository
paths and 77 compiler paths remain present; no corpus reduction or owned skips.

The five older suites pass byte comparison on controls and both pinned corpora,
including sanitizer executions. The new suite passes 293 controls in four typed
option profiles (388/368/381/361 findings), both corpora, sanitizer checks and
released-handle checks. Fixes and suggestions are compared; the three new Go
rules produce empty fix/suggestion arrays. All 18 kind manifests validate, with
numeric, unknown and duplicate declarations rejected. Six retained-registry
counterexamples are caught by required panic 70. The historical extra raw
signature-fact mutant was not rerun.

Inherited compiler checks: seven typeof fixtures pass Node, emitted JavaScript,
native, sanitizer and leak checks in 11.670s. Eight typeof mutants pass their
negative checks in 11.655s: five null fixtures, slot presence, constructor and
string literal. Successful sanitized executions differ from Node stdout alone.
No inherited compiler source was changed by this worker.

The context rule remains incomplete, not parked. Its full foreign-arena graph
refuses at stage1/typescript/parser/parser.ts:41:20: escaping this before every
field is set. The standalone foreign-arena adapter builds and runs successfully.
The cause is unestablished. Imported-helper Go findings versus native explicit
NotYet (70) reproduce the uncovered boundary and are excluded from equality.
No new rules are claimed while this blocker remains. Earlier three React-hook
claims remain parked for their documented HIR/SSA/capture requirements.

Commands and individual command outputs are archived in evidence/area-d3.
Owned gate: go test -count=1 -timeout 30m ./stage1/cohere/typeaware with the five
Wave16 agreement/mutant tests, ADAMIC_GATE_UNCACHED=1 and pinned corpus manifests;
python3 stage1/cohere/typeaware/wave16_seventh/validate.py with seventh-gate output.
Inherited oracle filters: TestNativeAgreesWithNode typeof_null/dispatch/string_literal;
TestTypeOfNullMutant, TestTypeOfNullSlotPresenceMutant, TestTypeOfConstructorMutant
and TestTypeOfStringLiteralMutant. go vet ./stage1/cohere/typeaware
./bridge/tsgo/checker and git diff --check pass. Full repository correctness gate,
including the 17 required external-input checks, was not run; no broad green claim.

Toolchain setup was reused: Go 0s, clang 1s, Node 1s, submodules 1s, cache 85s,
done 85s; nproc 5 (effective 4). To retain working space, only 35 obsolete generated
ELF/archive artifacts were removed, freeing 1,455,160,769 bytes; sources and logs
were retained. No active or current evidence artifacts were removed.

Rule mutants (successful native build/run, Go byte comparison catches):

| Mutant | First differing byte |
| --- | ---: |
| throw-range | 130 |
| backreference-range | 9283 |
| arrow-fix | 6056 |
| listener-general | 723 |
| render-number | 7165 |
| mock-range | 3576 |
| function-range | 132 |
| native-range | 4598 |
| wrapper-range | 5183 |
| import-write | 66 |
| exponent-base | 15678 |
| nan-comma | 12079 |
| shell-heredoc | 8299 |
| sql-string | 11784 |
| alert-range | 130 |
| fragment-name | 464 |
| undef-intrinsic | 62 |
| context-object | 43450 |

Three alternating serial process runs after all gates finished, including checker
load and parsing. Go timing stderr is expected and retained.

| Corpus | Native median | Go median | Native / Go |
| --- | ---: | ---: | ---: |
| compiler | 2.534698s | 0.344670s | 7.35x |
| repository | 0.340676s | 0.148455s | 2.29x |
