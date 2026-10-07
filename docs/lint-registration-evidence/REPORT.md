# Lint directory registration verification

Implemented on `codex/lint-registration`, based on origin/main `5d4c8012a0877094134e6c6bac367ff68f9313e8`. Implementation frozen at `3383c50`. All five rules on that main moved: eqeqeq, no-debugger, no-duplicate-case, no-empty, no-var. [Before](rules-before.txt) and [after](rules-after.txt) lists have an empty [diff](rules.diff).

## Implementation

Each rule owns its descriptor, concrete factory/listener, messages, upstream Go adapter, witness and mutant. Discovery generates ignored static TS imports/dispatch and Go oracle selection. Corpus filters, module copying, options capture and mutants derive from the directories too. New rules omit legacy order numbers and need no shared import, switch, options schema or test-list edits. Raw JSON options support rule-local decoding. Factories run after full ancestry is initialized; optional prepare/finish hooks support stateful listeners.

See [the contract](../lint-registration.md) and updated CLAUDE.md. Generated files are atomic, deterministic and only rewritten when bytes change. Go adapters invoke the unchanged pinned cohere implementation. Compiler/native implementation files and the cohere gitlink are unchanged.

Implementation commits, oldest first: `2917544`, `b39cb0b`, `4450eea`, `3b150fb`, `48741f9`, `e93a291`, `3bb5983`, `fe4daf3`, `3383c50`. Subsequent evidence-only commit contains this report.

## Environment and checks

`bash cloud/setup.sh > /tmp/lint-setup.log 2>&1` succeeded: Go ready 0s; clang ready 1s; Node ready 1s; submodules ready 1s; build cache warm 39s; total 39s. `nproc` = 5; cgroup CPU quota = 4 CPUs, memory 17.6 GB. [Exact setup output](setup.txt). Commands source `/workspace/adamic-tools/env.sh`.

Pinned external compiler corpus: Microsoft/TypeScript v6.0.3, commit `050880ce59e30b356b686bd3144efe24f875ebc8`, at `/workspace/scratch/typescript-6.0.3`; 77 compiler source files. Cohere executable built with `go build -o /workspace/scratch/cohere ./command/cohere` inside the pinned submodule.

- `go vet ./... > /tmp/lint-vet-final.log 2>&1`: exit 0, empty output. gofmt output empty.
- Pinned cohere check of 15 owned TS modules: exit 0, all 15 Adamic-ready; [output](cohere.txt).
- `ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 go test ./stage1/cohere/lint -count=1 -v -timeout 30m > /tmp/lint-frozen-tests.log 2>&1`: exit 0, 328.957s. 217 unique upstream combinations plus inherited generated cases; Go, Node and sanitized native output identical, 128446 bytes. Compiler/stage1 corpus: 194 files, 12470202 identical bytes. Owned witnesses: 5438 identical bytes. [Evidence](lint-tests.txt).
- `go test ./stage1/cohere/lint/registry -count=1 -v > /tmp/lint-registry-frozen.log 2>&1`: exit 0, 0.020s. Deterministic bytes/mtime and descriptor rejection mutants pass. [Evidence](registry-tests.txt).
- `ADAMIC_GATE_UNCACHED=1 ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 go test -p 2 -count=1 -timeout 30m ./... > /tmp/lint-full-frozen-uncached.log 2>&1`: **exit 0, every package passed**. Native 232.685s; oracle 201.524s (recorded counts unchanged); Unicode 705.404s; lint 414.890s; registry 0.023s; parser 385.085s. Package concurrency was limited to 2. [Complete output](full-gate.txt). No source edits occurred during this run.

## Mutants actually run

All five owned behavior mutants compiled and ran, then differed from the upstream Go oracle on **both Node and sanitized native**: debugger fix suppressed; empty function body incorrectly reported; equality suggestion applied as fix; var declaration suppressed; duplicate case suppressed. Additional executable mutants: valid but wrong descriptor node kind, omitted finish hook, ignored decoded AllowEmptyCatch option. Each was caught on both backends. A nested module-copy test proved the healthy nested import works; restoring the old root-only rewrite failed Node module resolution and native module loading. Positive factory test checks initialized ancestry, structured options, and factory/prepare/visit/finish sequencing.

Generator rejection mutants: duplicate public name, unknown JSON field, missing visit export, invalid AST kind, missing Go adapter export, no listener kinds, missing factory, missing class, missing finish export, unsafe public name, and duplicate oracle adapter. All rejected with logged errors. Healthy regeneration preserves identical bytes and modification times.

## Actual conflict-free trial merge

Both trial branches forked the same frozen implementation `3383c50`:

- `codex/lint-registration-trial-v2-continue`: `078db09`, only six files in `rules/no-continue/`.
- `codex/lint-registration-trial-v2-with`: `2e1ef26`, only six files in `rules/no-with/`.
- On `codex/lint-registration-trial-v2-merge`, `git merge --no-ff codex/lint-registration-trial-v2-continue` produced `3052b69`, then `git merge --no-ff codex/lint-registration-trial-v2-with` produced `b1b7e43`.

Both merges printed `Merge made by the 'ort' strategy.`, returned 0, and `git ls-files -u` was empty. [Literal merge output](trial-merge.txt). No shared files changed on either rule branch; generated artifacts remained ignored. Trial worktree reused the dependency through an uncommitted symlink, never a committed gitlink change.

Reproduction without local trial refs: branch twice from `3383c50`, apply and commit [continue patch](trial-continue.patch) on one and [with patch](trial-with.patch) on the other, then merge both into a third branch with `--no-ff`. Run `go run ./cmd/lint-registry`; [seven discovered names](trial-rules.txt).

On the merged trial, `go test ./stage1/cohere/lint -run 'TestRulesAgree|TestOwnedWitnesses|TestMutants' -count=1 -v -timeout 30m > /tmp/lint-trial-v2-tests.log 2>&1` returned 0 in 201.485s: 239 unique upstream combinations; 134747 identical bytes across Go/Node/sanitized native; witnesses 7800 identical bytes. **All seven** owned mutants caught on both backends, including the new continue diagnostic-ID and with-span mutations. [Evidence](trial-tests.txt). Trial additions are not part of the production branch; patches preserve the proof.

## Development failures and limits

Initial trial exposed a global mutant-anchor collision (`unexpected` appeared in three files). Mutants now target their owning file; final seven-rule trial passes. Cohere formatting invalidated an old no-empty anchor; stable replacement is verified. An initial anchored upstream test filter captured zero suffixed test names; prefix selection now captures 217. Unsupported array spread was replaced by supported forEach. Decoding settings inside constructors hit the documented constructor-proof gap; driver decoding and factories outside aggregate constructors pass without compiler changes. An initial cohere build at its root failed because that directory has no Go files; the command package build succeeded.

An early full uncached attempt overlapped harness edits and failed with a Go import-config error for regexp; its Unicode Node subprocess also exceeded its four-minute deadline under concurrent load. That attempt did pass internal/oracle in 477.783s without changing counts and native in 794.598s. Final verification uses frozen source and package concurrency 2. Earlier failing anchor output is retained in lint-tests-before-anchor-fix.txt; its focused replacement passed in empty-mutant-after-fix.txt.

Remote main advanced during verification to `50045bd797650a34aa40b55ad751b6a667a6ab31` (six compiler fixes and counts regeneration). Its diff from the starting base contains no lint changes. This branch retains its starting base; the final gate tests that base plus this unit, not those later integration changes.

This migrates main's five rules, not the thirty-rule scanner worker branch. No broader parser recovery, type-checker, new fixer behavior or throughput claim is made (throughput test skipped). The separate fixture/counts migration remains on codex/no-shared-lists and is not included here. Trial branches are local; the requested production branch is pushed without a PR.
