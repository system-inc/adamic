Rebased the restored wave-16 workspace onto the current lint-area registry migration.
Commits: area b46914832, main c7991b900, rebased parent f985b32af; evidence in this commit.
Checks: fifteen-port oracle 405.976s; new 60-command gate, sanitizer and metadata checks PASS.
Mutants: all 18 native rule mutants caught by Go bytes; six handle counterexamples caught.
Not covered: cross-file context integration, emitted JavaScript for rules and the full 17-check gate.

The requested rebase onto origin/area/stage1-lint completed cleanly at b46914832.
This contains current main c7991b900 and the integrated harness. The inherited
legacy syntax-rule registration migration was accepted. No shared file or
skip condition was changed, reverted, relaxed or deleted. Only the worker
branch codex/typeaware-wave-16 is pushed. Both baselines remained unchanged
through validation. Rule implementation sources are unchanged.

The restored workspace retained its tools and corpus inputs. The cloud runtime
skill was read for resumed work; current environment observations report running,
restricted HTTP policy enforced, and GOTOOLCHAIN/ADAMIC_TOOLS ready. GitHub access
uses the enforced proxy policy. No credential contents were inspected. Setup was
reused: Go 0s, clang 1s, Node 1s, submodules 1s, cache 85s, done 85s. nproc is 5.

Commands, with output written directly to logs:

    source /workspace/adamic-tools/env.sh
    git fetch origin '+refs/heads/*:refs/remotes/origin/*'
    git rebase origin/area/stage1-lint
    python3 stage1/cohere/typeaware/wave16_seventh/validate.py /workspace/wave16-artifacts/seventh-gate
    ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./stage1/cohere/typeaware -run '^TestWave16(Agreement|FollowupAgreement|ThirdAgreement|FourthAgreement|FifthAgreement)AndMutants$' -v
    go vet ./stage1/cohere/typeaware ./bridge/tsgo/checker
    git diff --check

The prior-port command also sets its five artifact directories, both manifests
and ADAMIC_TYPESCRIPT_SOURCE as in previous reports. It additionally inherited
an unused ADAD=1 assignment, recorded in provenance. All five suites pass controls,
options, full canonical corpus findings/fixes/suggestions, sanitizers, fifteen
mutants and five retained-registry counterexamples. The new suite passes 293
controls in four profiles, 77 compiler files and 287 repository files, normal
and sanitizer, three mutants and one additional released-handle counterexample.
All 287 pinned repository corpus paths still exist; the manifest was not reduced
or changed. Newly added files outside that pinned manifest are not claimed as
covered. The required TypeScript corpus is provided; selected tests report no
skips. Vet and diff-check logs are empty.

The two JSX name rules remain complete on the owned oracle. The context rule's
local construction/memo/return/escape paths match Go. Its full cross-file native
arena integration still fails compilation at shared parser.ts:41:20:

    Adamic 0.1 refuses this escaping a constructor before every field is set
    (stored, passed, or a method called on it, which could read a field that
    holds undefined while its type says otherwise); assign every field first,
    then use this

The standalone arena adapter still builds and runs. An imported-helper control
reports normally in Go and reaches explicit native NotYet (exit 70), excluded
from agreement. Its cause remains unestablished. The earlier parser-name and
shared-parser-only attempts are historical evidence, not rerun here. The context
rule remains incomplete and is not parked under the HIR/capture exception. The
three previously parked React analysis claims remain parked. No new claims.

Each of these mutants compiles, exits 0 with empty native stderr, and is caught
only by independent Go canonical bytes. First differing byte:

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

The Go Kind-name oracle checks all 18 manifests and rejects numeric, unknown and
duplicate-kind declarations. Six retained-registry counterexamples change the
required released-handle panic 70 to exit 0. The original extra signature-fact
mutant was not rerun; its earlier evidence is retained. Empty fix/suggestion
arrays are compared for the new three rules; their Go sources have no regexes.

The full repository gate, including the 17 required real-TypeScript/postcss/
graphql input checks, was not run. Complete bridge and filtered compiler oracles
were not redundantly rerun because their sources and dependencies are unchanged;
previous results remain recorded in AREA_B84_LANDING.md. Current owned bridge
handle and corpus comparisons pass. No repository-wide green is claimed.

Three alternating serial count-mode process runs after concurrent gates finished,
including checker load and native parsing:

| Corpus | Native median | Go median | Native / Go |
| --- | ---: | ---: | ---: |
| TypeScript compiler | 2.213259s | 0.268821s | 8.23x |
| Repository | 0.343844s | 0.124615s | 2.76x |

Current source hashes, logs, commands, mutant results and serial samples are
under evidence/area-b469. Historical reports retain their original baselines.
