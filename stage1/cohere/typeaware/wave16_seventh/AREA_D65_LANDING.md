Rebased wave-16 onto the latest integrated native runtime; no new rules claimed.
Commits: area d65a8f931, main 39638d9e2, rebased parent c2f87510e; evidence in this commit.
Checks: fifteen-port oracle 414.449s, new 60-command gate PASS, filtered runtime oracles PASS.
Mutants: all 18 native rule mutants caught by Go bytes; metadata and released-handle checks pass.
Not covered: cross-file context integration, emitted JavaScript for rules and the complete repository gate.

The requested rebase onto origin/area/stage1-lint completed cleanly at d65a8f931.
This contains current origin/main 39638d9e2 and the shared harness 41eb6eab2.
Both remote baselines remained unchanged during validation. Only the worker
branch codex/typeaware-wave-16 is pushed. No shared file was edited or reverted.
The inherited runtime heap/string improvements were accepted and verified.
Rule implementation sources are unchanged from the preceding landing.

The two JSX name rules remain complete on their owned oracle. The context-value
rule remains partial: all implemented local construction/memo/return/escape
paths pass the Go comparison, but foreign helper bodies explicitly refuse.
The current full-arena integration probe still fails compilation at
stage1/typescript/parser/parser.ts:41:20:

    Adamic 0.1 refuses this escaping a constructor before every field is set
    (stored, passed, or a method called on it, which could read a field that
    holds undefined while its type says otherwise); assign every field first,
    then use this

The standalone foreign arena still builds and runs normally. An imported-helper
case reports normally in Go and reaches native NotYet (exit 70), excluded from
agreement. The cause of the full-graph refusal is not established. The previous
shared-parser-only workaround probe is retained in AREA_LANDING.md; its parser
and compiler sources did not change in this runtime integration. This is not a
shared serializer gap or the HIR/capture analysis eligible for parking. The
three previously parked React analysis claims remain parked. No new claims
were taken while the context rule remains incomplete.

Commands, all output written directly to logs:

    source /workspace/adamic-tools/env.sh
    git fetch origin '+refs/heads/*:refs/remotes/origin/*'
    git rebase origin/area/stage1-lint
    python3 stage1/cohere/typeaware/wave16_seventh/validate.py /workspace/wave16-artifacts/seventh-gate
    ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./stage1/cohere/typeaware -run '^TestWave16(Agreement|FollowupAgreement|ThirdAgreement|FourthAgreement|FifthAgreement)AndMutants$' -v
    ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/oracle -run '^TestRuntimeLastIndexOfMatchesNode$' -v
    go test -count=1 -timeout 30m ./internal/native -run '^TestRuntime(ReleasePaths|StringEquality)$' -v
    go vet ./stage1/cohere/typeaware ./bridge/tsgo/checker
    git diff --check

The prior-port command sets its five owned artifact directories, both corpus
manifests and ADAMIC_TYPESCRIPT_SOURCE, as recorded in earlier landing reports.
All five suites pass their controls/options, full canonical corpus findings,
fixes/suggestions, sanitizer checks, fifteen mutants and five retained-registry
counterexamples. The new suite passes 293 controls in four profiles, 77 compiler
files and 287 repository files, normal and sanitizer comparisons, three native
mutants and another released-handle counterexample. The three rules have no
fixes or suggestions; empty arrays are compared too. The Go Kind-name oracle
checks all 18 manifests and rejects numeric, unknown and duplicate kinds. The
three new Go JSX rules contain no Go regular expressions. Vet and diff-check
logs are empty. The full bridge package gate was not rerun; its previous result
and evidence remain in AREA_LANDING.md. Current owned bridge/handle tests pass.

The filtered runtime oracle is uncached and matches Node, emitted JavaScript,
release native and sanitizers on 758 output bytes (9.107s). Release-path and
string-equality tests pass (7.467s package time). This emitted-JavaScript check
covers the inherited runtime fixture, not the lint rules.

Each rule mutant compiles and exits 0 with empty native stderr; only the
independent Go canonical byte comparison catches it. First differing byte:

| Mutant | Byte |
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

The extra signature-fact mutant from the original implementation was not
rerun; its historical evidence is retained separately. Six retained-registry
counterexamples across the owned suites change required panic 70 to exit 0.

Three alternating serial count-mode runs after all concurrent gates ended,
including checker load and native parsing:

| Corpus | Native median | Go median | Native / Go |
| --- | ---: | ---: | ---: |
| TypeScript compiler | 2.305222s | 0.286632s | 8.04x |
| Repository | 0.331355s | 0.128431s | 2.58x |

Setup was reused: Go 0s, clang 1s, Node 1s, submodules 1s, cache 85s, done 85s.
nproc remains 5 with four effective cores. To keep builds from exhausting disk,
38 obsolete ELF/archive files (1,574,858,533 bytes) were removed only from the
old /workspace/wave16-artifacts/rebase-main scratch run; sources and logs remain.
Current evidence, source hashes, commands and serial samples are in evidence/area-d65.
