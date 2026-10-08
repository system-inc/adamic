# Validation commands

All commands ran from `/workspace/adamic`, except the explicitly named cohere
upstream capture. Toolchain setup and timing output are in `setup.log`.

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh
bash cloud/setup.sh --wasi-sdk
source /workspace/adamic-tools/env.sh
git fetch origin
git fetch --no-recurse-submodules origin '+refs/heads/*:refs/remotes/origin/*'
git grep -l '"name": "react/no-will-update-set-state"' $(git for-each-ref --format='%(refname)' refs/remotes/origin) -- 'stage1/cohere/lint/rules/*/rule.json'
git switch -c lint-rules/react-no-will-update-set-state origin/area/stage1-lint
git -C cohere/TypeScript fetch --no-recurse-submodules origin 050880ce59e30b356b686bd3144efe24f875ebc8
git -C cohere/TypeScript worktree add --detach /tmp/s13-typescript-clean 050880ce59e30b356b686bd3144efe24f875ebc8
git -C /tmp/s13-typescript-clean status --porcelain
git -C /tmp/s13-typescript-clean rev-parse HEAD
go run ./cmd/lint-registry
export ADAMIC_TYPESCRIPT_SOURCE=/tmp/s13-typescript-clean
export ADAMIC_LINT_PROFILE_DIR=/tmp/s13-lint-profile.g3Mbv1
export ADAMIC_LINT_PROFILE_SNAPSHOTS=/tmp/s13-lint-profile.g3Mbv1
export ADAMIC_LINT_BENCH=1
# WASI_SYSROOT is exported by env.sh.
go test ./stage1/cohere/lint -run 'TestRulesAgree|TestOwnedWitnesses|TestMutants/react-no-will-update-set-state' -count=1 -timeout=40m -v
go test ./stage1/cohere/lint -count=1 -timeout=60m -json
nproc
cat /proc/loadavg
```

The duplicate check returned no matches. The detached TypeScript checkout was
clean at the exact pin. Both profile variables refer to one fresh directory.
`selected.log` is the final selected run. An earlier attempt hit stderr from Go
module downloads; a superseded pre-guard run was stopped before final validation.

For the independent count and source archive, an external Go overlay inserted
`RecordAssertedCase(t, result)` at the upstream harness's return and ran:

```sh
cd /workspace/adamic/cohere
COHERE_DOCS_CAPTURE=/tmp/s13-react-no-will-update-set-state-capture/records go test -overlay=/tmp/s13-react-no-will-update-set-state-capture/overlay.json ./internal/lint/rules/react -run '^TestNoWillUpdateSetState' -count=1 -v
```

The overlay changes the capture hook only. It does not change the upstream rule.
The resulting 67 unique file/source/options cases are in `upstream-cases.json`.
The package harness independently captures all these cases and compares the Go
oracle with Node, emitted JavaScript and sanitized native. It compiles and runs
the semantic mutant on Node and emitted JavaScript; both catches are preserved
verbatim in `mutant.log`.
