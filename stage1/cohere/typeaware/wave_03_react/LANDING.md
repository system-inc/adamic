Rebased wave-03 onto origin/main e8ba3d5d; no new rules claimed.
Validated implementation tip 02239d7aacdf86197ea081c92ed9eca2bc24f65e; this report commit follows it.
Four owned oracle groups PASS, 626.166s; checker package PASS, 1.076s; touched-package vet PASS.
All 12 rule mutants caught by byte comparison; alias, regex/metadata guards and four released-registry mutants caught.
Arbitrary configured regex and numeric listener conversion remain unfinished; emitted-JavaScript integration and full repository gate not covered.

The only branch pushed by this unit is codex/typeaware-wave-03. It was rebased
onto e8ba3d5d81de4d3773c723914fccd4c76248b965 without conflicts. A second fetch
after validation returned the same main tip. Main and area branches were not pushed.

Validation used the unchanged frozen 77 compiler and 287 repository manifests.
The complete owned rule suites compared finding, fix and suggestion bytes against
production Go cohere, including ASan/UBSan corpus and control runs. The complete
output is retained beside this report as landing-oracle.log.gz. Bridge and vet
logs are also retained. The two interrupted, overlapping preliminary runs are
excluded from evidence; the final run used fresh landing-final artifact paths.

Command (after sourcing /workspace/adamic-tools/env.sh):

```sh
ADAMIC_WAVE03_ARTIFACTS=/workspace/wave-03/landing-final-original \
ADAMIC_WAVE03_NEXT_ARTIFACTS=/workspace/wave-03/landing-final-next \
ADAMIC_WAVE03_MORE_ARTIFACTS=/workspace/wave-03/landing-final-more \
ADAMIC_WAVE03_REACT_ARTIFACTS=/workspace/wave-03/landing-final-react \
ADAMIC_WAVE03_COMPILER_MANIFEST=/workspace/wave-03/compiler.manifest \
ADAMIC_WAVE03_REPOSITORY_MANIFEST=/workspace/wave-03/repository.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-03/typescript-corpus \
go test ./stage1/cohere/typeaware -run '^TestWave03(AgreementAndMutants|ContinuationAgreementAndMutants|MoreAgreementAndMutants|ReactAgreementAndMutants)$' -timeout 30m -count=1 -v
go test ./bridge/tsgo/checker -count=1
go vet ./stage1/cohere/typeaware ./bridge/tsgo/checker
```

The rule mutants, one per rule, all compiled, exited successfully, and were
caught by differing output bytes: unsafe-call, imports, template, race, output,
blocking, throw, backreference, callback, unsupported, memo and boolean. The
React message mutants preserved finding counts. Memo alias filtering was also
mutated and caught by bytes. Regex and metadata guard mutants escaped their
required panic 70; the checks caught them. All four released-registry mutants
exited 0 and failed the required panic 70 checks. Normal released queries
produced the required panic.

Observed wall time, seconds (single observations, not a speedup claim):

| Group | Repository native / Go | Compiler native / Go |
|---|---:|---:|
| Original | 0.865 / 0.301 | 5.907 / 1.394 |
| Nexus | 0.625 / 0.223 | 3.843 / 0.581 |
| Core | 0.592 / 0.242 | 4.157 / 0.553 |
| React | 1.194 / 0.214 | 5.657 / 0.529 |

The runtime regex probe on current main still refuses compilation at
wave_03_react/gaps/general_regex.a:3:24 with
`stage 0 can't lower RegExp with a nonconstant pattern yet`. Boolean-prop-naming
therefore remains partial for arbitrary configured patterns. Shared compiler
files were not edited to work around this gap. The new numeric SyntaxKind,
handed-node listener requirement has not been implemented in these legacy
standalone drivers; they still use string kinds. The native parser currently
exposes ParseNode.kind as string. This report does not claim compliance with
the new speed requirement or emitted-JavaScript integration.

nproc reported 5. The existing setup observation remains: Go 0s, clang 1s,
Node 1s, submodules 1s, cache warm 174s, total 174s. No new setup timing is claimed.
