Rebased wave-16 onto current main through the integrated lint area; no new rules claimed.
Commits: area b84a9d931, main c7991b900, rebased parent 171d72ab2; evidence in this commit.
Checks: fifteen-port oracle 431.349s, new 60-command gate PASS, filtered compiler oracle 12.147s.
Mutants: all 18 native rule mutants caught by Go bytes; metadata and handle checks pass.
Not covered: cross-file context integration, emitted JavaScript for rules and the full 17-check repository gate.

The requested rebase onto origin/area/stage1-lint completed cleanly at b84a9d931.
It includes current main c7991b900 and the integrated shared harness. Both
baselines remained unchanged through validation. Only codex/typeaware-wave-16
is pushed. No shared file or skip condition was edited, reverted or relaxed.
Rule implementation sources are unchanged; the newly inherited compiler
lowering/refusal and record-runtime changes were accepted and verified.

The two JSX name rules remain complete on their owned comparison. The context
rule's local construction, memo dependency, return and escape paths agree with
Go. Its cross-file integration remains incomplete: the full native arena graph
still refuses at stage1/typescript/parser/parser.ts:41:20:

    Adamic 0.1 refuses this escaping a constructor before every field is set
    (stored, passed, or a method called on it, which could read a field that
    holds undefined while its type says otherwise); assign every field first,
    then use this

The standalone foreign adapter builds and runs. The imported-helper control
reports normally in Go but reaches explicit native NotYet (exit 70), excluded
from agreement. The cause of the full-graph refusal is unestablished. Earlier
owned parser-name and shared-parser-only probes remain recorded in previous
reports; neither is claimed as rerun here. This is not a shared serializer gap
or the HIR/capture analysis eligible for parking. The three previously parked
React analysis claims remain parked. No new rule was claimed while the context
rule remains blocked.

Commands, all output written directly to logs:

    source /workspace/adamic-tools/env.sh
    git fetch origin '+refs/heads/*:refs/remotes/origin/*'
    git rebase origin/area/stage1-lint
    python3 stage1/cohere/typeaware/wave16_seventh/validate.py /workspace/wave16-artifacts/seventh-gate
    ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./stage1/cohere/typeaware -run '^TestWave16(Agreement|FollowupAgreement|ThirdAgreement|FourthAgreement|FifthAgreement)AndMutants$' -v
    ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/proven_(guards|class_guards|assertions|satisfies|upcasts)\.a$' -v
    go vet ./stage1/cohere/typeaware ./bridge/tsgo/checker
    git diff --check

The prior-port command also sets all five owned artifact directories, both
corpus manifests and ADAMIC_TYPESCRIPT_SOURCE, as in earlier reports. All five
suites pass their controls/options and complete canonical corpus findings,
fixes/suggestions, sanitizers, fifteen rule mutants and five retained-registry
counterexamples. The new suite passes 293 controls in four profiles, 77 compiler
files and 287 repository files, normal and sanitizer, three rule mutants and an
additional released-handle counterexample. Its empty fix/suggestion arrays are
compared too. The Go Kind-name oracle checks all 18 listener manifests and
rejects numeric, unknown and duplicate names. The three new Go JSX rules contain
no Go regular expressions. Vet and diff-check logs are empty.

The filtered compiler oracle runs five new proven guard/assertion/satisfies/upcast
fixtures against Node, emitted JavaScript, native and sanitizers, uncached:
12.147s, native misses 15, Node misses 10. It and the selected lint suites have
zero reported skips; the required TypeScript corpus was provided. The full
repository gate, including all 17 required TypeScript/postcss/graphql input
checks, was not run. No repository-wide green or certification of those checks
is claimed. The complete bridge package gate was not rerun; historical results
remain in earlier reports, while current owned bridge/handle checks pass.

Each rule mutant compiles and exits 0 with empty native stderr. Only independent
Go canonical bytes catch it. First differing byte:

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

The extra original signature-fact mutant was not rerun; its historical evidence
is retained. Six current retained-registry counterexamples change the required
released-handle panic 70 to exit 0 and are caught. Metadata counterexamples
reject numeric, unknown and duplicate-kind declarations.

Three alternating serial count-mode runs after all concurrent gates completed,
including checker load and native parsing:

| Corpus | Native median | Go median | Native / Go |
| --- | ---: | ---: | ---: |
| TypeScript compiler | 2.453031s | 0.301581s | 8.13x |
| Repository | 0.351249s | 0.132512s | 2.65x |

Setup was reused: Go 0s, clang 1s, Node 1s, submodules 1s, cache 85s, done 85s.
nproc remains 5 with four effective cores. Current source hashes, logs, exact
commands, mutant bytes and serial samples are in evidence/area-b84.
