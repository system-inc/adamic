Rebased wave-16 onto current main and changed 18 listener manifests to ast.Kind names.
Commits: rebased parent 7fca09931; main b8fb957aa; metadata and evidence in this commit.
Checks: fifteen-port oracle 394.755s, bridge 66.583s, new suite and metadata oracle PASS.
Mutants: all 18 rule mutants caught by Go bytes; released-handle and metadata counterexamples caught.
Not covered: cross-file context-value integration, emitted JavaScript and the complete repository gate.

No new claim was added. jsx-fragments and jsx-no-undef remain complete on the
owned oracle. jsx-no-constructed-context-values remains partial: its local
construction, memo dependency, return and escape paths agree with Go, while
foreign resolved helper bodies explicitly refuse. It is not marked parked.

The branch was rebased cleanly onto origin/main b8fb957aa839a9e8cb0b54279dd9864fa317bd30.
Only codex/typeaware-wave-16 is pushed. Main and area branches are untouched.
The advertised shared harness ab70f38d4 is still on origin/lint-rules/harness
and is not an ancestor of this main. No shared harness, generator, parser or
protected compiler file was edited. Previous historical reports retain their
original baselines; evidence/landing-b8fb contains this continuation's logs.

All 18 owned rule.json declarations now contain ast.Kind names. The three new
listeners consume supplied cached nodes; their private driver's numeric switch
and the single legacy-parser compatibility conversion are unchanged. A new
owned kind_names.go reads typescript-go's enum directly. The validator checks
all declarations against that independent list and rejects numeric, unknown
and duplicate-kind mutants. The check ran separately this time because the
full suite was already running when the metadata check was added. The three
new Go JSX rules contain no Go regular expressions to translate.

Commands (all output redirected to logs):

    source /workspace/adamic-tools/env.sh
    python3 stage1/cohere/typeaware/wave16_seventh/validate.py /workspace/wave16-artifacts/seventh-gate
    ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./stage1/cohere/typeaware -run '^TestWave16(Agreement|FollowupAgreement|ThirdAgreement|FourthAgreement|FifthAgreement)AndMutants$' -v
    go test -count=1 -timeout 30m ./bridge/tsgo/...
    go run stage1/cohere/typeaware/wave16_seventh/kind_names.go
    go vet ./stage1/cohere/typeaware ./bridge/tsgo/checker
    python3 -m py_compile stage1/cohere/typeaware/wave16_seventh/validate.py
    git diff --check

The old-port command additionally sets the five owned artifact directories,
repository/compiler manifests and ADAMIC_TYPESCRIPT_SOURCE as recorded in
previous reports. It passes all five suites, canonical findings/fixes/suggestions,
normal and sanitizer corpora, and released-handle checks. The new suite passes
293 controls in four profiles, 77 compiler files and 287 repository files,
normal and sanitizer comparisons. Its three rules have no fixes or suggestions;
their empty arrays are still compared. Six retained-registry counterexamples
across the owned suites change the required released-handle panic 70 to exit 0.
Bridge checks pass too. Vet, Python syntax and diff-check logs are empty.

Every rule mutant compiles and exits 0; only independent Go diagnostic bytes
catch it. First differing byte:

| Mutant | Byte |
| --- | ---: |
| import-write | 66 |
| exponent-base | 15678 |
| nan-comma | 12079 |
| listener-general | 723 |
| render-number | 7165 |
| mock-range | 3576 |
| shell-heredoc | 8299 |
| sql-string | 11784 |
| alert-range | 130 |
| function-range | 132 |
| native-range | 4598 |
| wrapper-range | 5183 |
| throw-range | 130 |
| backreference-range | 9283 |
| arrow-fix | 6056 |
| fragment-name | 464 |
| undef-intrinsic | 62 |
| context-object | 43450 |

The extra signature-fact mutant from the previous report was not rerun in this
continuation. Its original evidence is retained, not presented as a new run.

The blocker was reproduced on this main. The standalone foreign adapter builds
and runs (exit 0). Integrating that same adapter into the full rule suite rejects
compilation at stage1/typescript/parser/parser.ts:34:20:

    Adamic 0.1 refuses this escaping a constructor before every field is set
    (stored, passed, or a method called on it, which could read a field that
    holds undefined while its type says otherwise); assign every field first,
    then use this

The cause is not established. The imported-helper control reports normally in
Go and reaches native NotYet (exit 70); that case is not counted as agreement.
This is a compiler integration blocker, not a missing suggestion serializer or
an assertion that the standalone parser is unavailable. Further claims remain
blocked by the unfinished rule; no shared-file workaround was attempted.

Three alternating serial count-mode process runs after all concurrent gates
finished include checker load and native parsing:

| Corpus | Native median | Go median | Native / Go |
| --- | ---: | ---: | ---: |
| TypeScript compiler | 2.478738s | 0.284783s | 8.70x |
| Repository | 0.347000s | 0.122481s | 2.83x |

Toolchain setup was reused: Go 0s, clang 1s, Node 1s, submodules 1s, cache 85s,
done 85s. nproc reports 5 (four effective cores). No new setup run is claimed.

Latest landing check: fetched all origin heads again. Main remains b8fb957aa,
and pushed commit 06d773130 is its descendant with a clean worktree. The new
shared harness tip is 41eb6eab2b6de45ede0a40250765be295ee25fbd, still not on
main. No compiler or rule source changed, so the green oracle and mutant
results above were not redundantly rerun.

An isolated scratch probe renamed the owned JSX parser class to Wave16JsxParser
while retaining the complete foreign-arena rule graph. After correcting the
probe's special adamic import, compilation still exits 1 with the identical
shared parser.ts:34:20 constructor-escape refusal. This rules out that specific
rename as a workaround; it does not establish the underlying cause. The final
probe stdout/stderr and fetch log are retained under evidence/landing-b8fb.
The context-value rule remains incomplete, and no new rule was claimed.
