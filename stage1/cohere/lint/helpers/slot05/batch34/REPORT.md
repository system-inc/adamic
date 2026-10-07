# Batch 34 blocked at dynamic RegExp construction

No new helper implementation is delivered. The 95 previous helpers remain complete and pushed. This batch reserves `github.com/system-inc/cohere/internal/lint/ecmascript/regexp.Compile` and stops at a reproduced compiler gap, following the instruction to report shared blockers rather than edit shared files.

## Ownership and ranking

Fetched main, area/stage1-lint and all twenty origin codex/lint-helpers* branches. Read every claim and the inherited shared HELPERS.md claim. The ownership snapshot is retained in evidence/claim-snapshot.json.gz. The comment collection bundle is delivered and excluded. `regexp.Compile` is mentioned in slot 01's factory claim as an external dependency, not reserved there. Earlier substring-based ranking mistakenly treated that mention as a claim. Corrected ownership gives Compile four blocked consumers, above the remaining concrete unclaimed helpers with three.

Claim commit `73aabd5e23fed9a1b77b0527946a7c88c55f0ed5` was successfully pushed before any batch34 source was written. Main `71d7e491b3c9724f7a0e2ee754592149e7f9790b` and lint area `2fbfe42155387f67f297a41c89d318e3b6cde310` are unchanged ancestors of this branch. Prior completed implementation: `c2c40aa70e3758d67a0209ea02dcc03b731c5c72`. No history was rewritten and only the own branch is pushed.

## Exact reproducer

`dynamic_pattern.a` reads a pattern from the command line, builds `new RegExp(pattern, 'u')`, and tests subject `a`. Input pattern `a` is sufficient; no unusual regex feature is needed.

With the setup environment sourced, from the repository root:

```
node --disable-warning=ExperimentalWarning oracle/node.mjs stage1/cohere/lint/helpers/slot05/batch34/dynamic_pattern.a a

go run ./cmd/adamic build stage1/cohere/lint/helpers/slot05/batch34/dynamic_pattern.a -o /tmp/lint05-batch34-probe --sanitize

go run ./cmd/adamic js stage1/cohere/lint/helpers/slot05/batch34/dynamic_pattern.a
```

Each command's output is retained in its own evidence log. Node prints `true` and exits 0. The Go overlay invokes the unchanged pinned cohere Compile with source `a` and flags `u`, then tests `a`: it also prints `true`, exits 0. Both Adamic commands exit 1, reporting:

```
stage 0 can't lower RegExp with a nonconstant pattern yet
```

The lowerer refuses before either backend produces an executable artifact. This is an observed compiler capability gap, not a regex semantic mutant or an oracle mismatch. The owned regression test asserts Go/Node answers and the typed NotYet refusal. An initial shell invocation used an unavailable `adamic` command, then an incorrect build invocation and a boolean console argument; these were corrected to the repository's Go command, required output argument, and string output before recording final evidence. Those preliminary failures are not evidence of the reported gap.

Compile accepts arbitrary option patterns at runtime and returns a compiled engine or an error. Constants cannot implement that contract. The separately owned matching/rewrite dependencies do not provide dynamic native RegExp construction. Supplying Go compilation answers to a callback would test wrapper orchestration but would leave this missing compilation operation external; it would not unblock the consumers. No handwritten matcher, shared compiler edit or reduced acceptance contract is supplied.

Go's wrapper also intentionally differs from strict JavaScript syntax validation under `u`, as documented in regexp.go. That compatibility boundary has been read but not independently tested in this batch; it must be addressed when dynamic construction becomes available.

## Readiness and limits

The four remaining blocked consumers of Compile are:

- @next/next/no-html-link-for-pages
- @typescript-eslint/no-empty-object-type
- no-restricted-exports
- no-restricted-imports

Zero new rules lose a helper blocker. The cumulative delivered helper count remains 95. No other new helper is claimed while this highest-ranked helper is blocked. Its claim remains explicitly blocked pending compiler support. No consumer fixture parity, emitted JavaScript execution or native execution is claimed for Compile, and no compiling semantic mutant exists for an undelivered implementation.

The complete repository gate and its 17 required external correctness comparisons were not run. No selected check was skipped, relaxed or removed. Prior pushed evidence is preserved. No shared harness, registry, rule or compiler file was edited.

## Setup

`export GOPROXY='https://proxy.golang.org|direct'` preceded `bash cloud/setup.sh`; setup succeeded, total 24.913s. `source /workspace/adamic-tools/env.sh` preceded builds. nproc: 5; cgroup cpu.max: 400000 100000; memory: 17.6 GB. Timing observations: Go ready 0.025s, Node ready 0.032s, submodules ready 0.063s, markdown dependencies ready 0.082s, clang ready 0.196s, Go build ready 24.754s, test binaries deferred 24.883s, cache warm 24.885s. The complete setup log is preserved losslessly.

## Validation

The owned boundary test passed in 0.157s. It compares unchanged Go Compile and source Node and asserts the specific shared lowerer refusal. The latest completed batch33 helper suite, all its semantic mutants, the filtered uncached Node input oracle, repository-wide vet and formatting were rerun. Their final observed results are recorded below after completion.

- `ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch34 -count=1 -v`: PASS, 0.157s.
- `ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch33 -count=1 -v -timeout=20m`: PASS, 61.916s; 8,539 Go/source/emitted/native queries and all 21 compiling semantic mutants caught. Exact replacements and first differing output lines are retained in evidence/prior-helpers.log; contracts and complete mutant descriptions remain in ../batch33/REPORT.md. This does not claim all 95 helpers or 471 historical mutants were rerun.
- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v`: PASS, 0.930s; seven probe misses, zero hits.
- `go vet ./...`: exit 0, empty log.
- `gofmt -l cmd internal stage1/cohere/lint/helpers/slot05`: exit 0, empty log.

All test command output went directly to the retained log files. No test run was piped.
