Rebased wave-16 onto the integrated stage1-lint harness; no new rules claimed.
Commits: area 7481e0324, main 39638d9e2, rebased parent 030908d9c; evidence in this commit.
Checks: fifteen-port oracle 429.309s, bridge 74.941s, new suite 60 commands PASS.
Mutants: all 18 native rule mutants caught by Go bytes; handle and metadata counterexamples caught.
Not covered: cross-file context integration, emitted JavaScript and the complete repository gate.

The requested rebase onto origin/area/stage1-lint completed cleanly. This tested
baseline includes origin/main 39638d9e2, harness 41eb6eab2 and its integration
at 50a5f105. Both main and the area baseline are ancestors of the worker branch.
Only codex/typeaware-wave-16 is pushed. Main remained unchanged during the gate;
the area branch advanced to d65a8f931 while testing. That later area tip is not
claimed as tested. No shared-file diff was reverted or edited.

The two JSX name rules remain complete on the owned comparison. The context
rule's local construction, memo dependency, return and escape paths agree with
Go. Its foreign resolved helper path remains incomplete and explicitly refuses
rather than silently returning incomplete diagnostics. This is not a missing
shared suggestion serializer, nor the HIR analysis that qualifies React rules
for parking. The three previously parked React analysis claims remain parked.
No additional rules were claimed.

Commands, with output written directly to logs:

    source /workspace/adamic-tools/env.sh
    git fetch origin '+refs/heads/*:refs/remotes/origin/*'
    git rebase origin/area/stage1-lint
    python3 stage1/cohere/typeaware/wave16_seventh/validate.py /workspace/wave16-artifacts/seventh-gate
    ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./stage1/cohere/typeaware -run '^TestWave16(Agreement|FollowupAgreement|ThirdAgreement|FourthAgreement|FifthAgreement)AndMutants$' -v
    go test -count=1 -timeout 30m ./bridge/tsgo/...
    go vet ./stage1/cohere/typeaware ./bridge/tsgo/checker
    git diff --check

The prior-port command additionally sets all five owned artifact directories,
compiler/repository manifests and ADAMIC_TYPESCRIPT_SOURCE, as in LANDING_B8FB.md.
All five suites pass canonical findings, fixes and suggestions over their
controls, options and two corpora, including sanitizer comparisons and five
released-handle counterexamples. The new gate passes 293 controls in four
profiles, 77 compiler files and 287 repository files, normal and sanitizer.
It now also runs the owned Go Kind-name oracle: all 18 manifests use valid
ast.Kind names, and numeric, unknown and duplicate-kind mutants are rejected.
Its released-handle check requires panic 70; retaining the handle changes it to
exit 0 and is caught. The new Go JSX rules have no fixes, suggestions or Go
regular expressions; their empty edit/suggestion arrays are compared too.
Vet and diff-check logs are empty. The full repository gate was not run.

Each of these rule mutants compiles, exits 0, and is caught only by independent
Go canonical bytes. First differing byte:

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
rerun in this continuation; its historical evidence is retained separately.

The original full foreign-arena integration probe still fails compilation at
stage1/typescript/parser/parser.ts:41:20 with:

    Adamic 0.1 refuses this escaping a constructor before every field is set
    (stored, passed, or a method called on it, which could read a field that
    holds undefined while its type says otherwise); assign every field first,
    then use this

The standalone arena adapter still builds and runs. An additional scratch
probe uses the newly integrated shared JSX parser for both root and imported
source arenas, removing the duplicate private parser classes from that graph.
This complete rule graph still exits 1 with the identical constructor refusal
at line 41. Its source and stdout/stderr are preserved. No shared parser or
compiler workaround was attempted. The cause remains unestablished. A separate
imported-helper case reports normally in Go and reaches explicit native NotYet
(exit 70); it is excluded from agreement. The context rule remains blocked.

Three alternating serial count-mode runs after all concurrent gates finished
include checker load and native parsing:

| Corpus | Native median | Go median | Native / Go |
| --- | ---: | ---: | ---: |
| TypeScript compiler | 2.558115s | 0.290903s | 8.79x |
| Repository | 0.353082s | 0.129927s | 2.72x |

Toolchain setup was reused: Go 0s, clang 1s, Node 1s, submodules 1s, cache 85s,
done 85s. nproc remains 5, with four effective cores. Current logs, source hashes,
probe sources, exact commands and serial samples are in evidence/area-7481.
