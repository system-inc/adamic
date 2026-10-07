# Outstanding Next rule completed against its parser dependency

Built `@next/next/no-title-in-document-head` in `no_title_in_document_head.a`.
The original claim is `27d09b55`; this completion is pushed before another claim.
`TestWave18TitleWithJsxSlice` passed in 63.700s; full comparison logs are preserved.
The intrinsic-name mutant exits 0 with empty stderr and differs from Go at byte 57.
Integration still requires the native JSX parser slice; shared parser and harness files are untouched.

## Implementation and dependency

The visitor reproduces the production Go rule's first-named-specifier behavior, aliases,
shadow-aware declaration identity, preorder import visibility, direct JSX children, self-closing
titles, duplicate findings, head-tag range and complete message. It uses the existing
`binding-declarations` question, with no new checker extension and no Go lint verdict.

Native JSX descent already exists on `origin/codex/stage1-jsx-lint`, commit
`a8a62d62ca49db7415e14c3887dd305022b17309`. Validation extracts that branch's
`stage1/typescript` directory into an isolated scratch snapshot. The test builds private copies
of the rule driver and dependencies with imports pointing at that snapshot. This is actual native
parsing, not a Go-provided AST. No parser, shared harness, registration generator, compiler hot
file or submodule pin is changed on this branch. The snapshot source hashes are preserved.

The parser on this branch still lacks JSX grammar and refuses the previously recorded positive
control with `parser slice expected GreaterThanToken, got Identifier`. The rule source and tests
are complete; integrating that already-published parser dependency is the remaining shared gap.
Per the latest instruction, work on the rule is pushed and continuation may proceed. The test
explicitly skips when its dependency environment variable is absent; a skip is not an agreement pass.

## Evidence

Thirty controls preserve the production Next test sources, including its three original upstream
fixtures, first-specifier ordering, default/namespace imports, aliases, shadowing, direct/nested
children, expression children, duplicate titles and tag-name spans, with extra Unicode/CRLF and
fragment cases. The pinned production Next tests pass before byte comparison. Independent Go
uses the unchanged rule through the registry, its own checker loader and AST walk. Complete
findings, fix fields and suggestion fields match; this rule produces no fixes or suggestions.

| Population | Findings | Identical bytes, normal and ASan/UBSan |
| --- | ---: | ---: |
| JSX controls, 30 files | 12 | 6228 |
| Frozen repository, 287 files | 0 | 18485 |
| TypeScript compiler, 77 files | 0 | 5318 |

The mutant reverses the intrinsic `title` comparison. It compiles and exits successfully with
empty stderr; only the complete independent Go byte comparison catches it. The same compiler
and repository manifests, TypeScript v6.0.3 pin and ownership/sanitizer foundation from the original
wave apply. The earlier released-handle and registry mutant checks remain passed and committed.

Observed whole-process median seconds over three rounds: repository native 0.185151 versus
Go 0.113240; compiler native 1.346403 versus Go 0.288483. These rounds use native followed by
Go, not alternating order, and do not establish general speed parity. The corpora have no findings
for this rule. Raw load/query/run timings are in the log.

```sh
# Extract the published parser dependency with git archive into a scratch directory first.
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE18_TITLE_ARTIFACTS=/workspace/wave-18-title-validation \
ADAMIC_WAVE18_JSX_SLICE=/workspace/wave-18-jsx-slice \
ADAMIC_WAVE18_COMPILER_MANIFEST=/tmp/wave-18-compiler.manifest \
ADAMIC_WAVE18_REPOSITORY_MANIFEST=/tmp/wave-18-repository.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-18-typescript \
go test ./stage1/cohere/typeaware -run '^TestWave18TitleWithJsxSlice$' \
  -count=1 -timeout=30m -v > /tmp/wave-18-title-test.log 2>&1
```

Both new Adamic files were formatted with the pinned cohere native printer and checked stable
on a second pass. The new Go files are gofmt-clean. Full repository tests and the shared `.a`-aware
CLI lint gate are not claimed. The two earlier default-option ports and their independent mutants
remain as recorded in WAVE_18_REPORT.md. No PR is opened.
