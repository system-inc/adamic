# Lower may-hold-null proof

## Checkout and environment

- Tested HEAD: `5f54e9295b67b0df89f662d9366737621d27dca2`, checked out detached after `git fetch origin compiler/may-hold-null-shared`.
- Parent: `a4a7f4ab7b015e1f37b31564a52e885e025e9504` (one commit on main).
- `git submodule update --init --recursive` completed. Submodules: `cohere` at `5360753e19e4417675a90e92a1823bea66b13918`; `cohere/TypeScript` at `d92d9bfee114c80be2c375d72edae966176e3a4f`.
- Loaded `/workspace/adamic-tools/env.sh` first for setup and Go commands. Created its missing `TMPDIR`, `/tmp/adamic-gate`, outside the repository.
- Provisioned upstream `https://github.com/microsoft/TypeScript.git` outside the repository at `/tmp/adamic-proof-typescript`, checked out commit `050880ce59e30b356b686bd3144efe24f875ebc8`, and exported `ADAMIC_TYPESCRIPT_SOURCE=/tmp/adamic-proof-typescript` for every proof run.
- Copied `stage3/api/package.json` and `package-lock.json` to `/tmp/adamic-proof-api` and ran `npm ci --prefix /tmp/adamic-proof-api`. Verified the lockfile matches the tested checkout. Packages: TypeScript `6.0.3`, `@types/node` `25.3.3`. Prepended `/tmp/adamic-proof-api/node_modules/.bin` to PATH; `tsc --version` printed `Version 6.0.3`.
- Tool versions: Go `go1.27.1 linux/amd64`, Node `v24.19.0`, Clang `20.1.8`.
- The initial working tree was clean. Dependencies and logs were kept outside the repository. No implementation changes are included in this report commit.

## 1. Formatting and vet

`gofmt -l internal/lower`: PASS; printed nothing.

`go vet ./internal/lower`: PASS, exit 0; printed nothing after recursive submodule setup completed.

An earlier vet attempt during submodule download exited 1 with `go: cannot load module cohere/TypeScript/tsc listed in go.work file: open cohere/TypeScript/tsc/go.mod: no such file or directory`. It was rerun successfully after setup; this was a setup dependency failure.

## 2. Selected uncached oracle tests

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestParameterDefaultNullAgrees|TestDestructuringDefaultsAgree|TestDestructureUnionAgrees' -count=1 -v -timeout 60m
```

PASS, exit 0. Package time: `276.137s`. Every reported test/subtest result follows, including parent results. There were no failures, hence no first errors to report. Four optional WASI subtests skipped with `set ADAMIC_ORACLE_WASI=1 to run the WASI oracle`; these are not claimed as passes.

| Test/subtest | Result |
| --- | --- |
| `TestDestructuringDefaultsAgree` | PASS |
| `TestDestructuringDefaultsAgree/parameter_pattern.a` | PASS |
| `TestDestructuringDefaultsAgree/let_pattern.a` | PASS |
| `TestDestructuringDefaultsAgree/default_effect.a` | PASS |
| `TestDestructuringDefaultsAgree/version_with.a` | PASS |
| `TestDestructuringDefaultsAgree/earlier_binding.a` | PASS |
| `TestParameterDefaultNullAgrees` | PASS |
| `TestParameterDefaultNullAgrees/effect_on_undefined_only.a` | PASS |
| `TestParameterDefaultNullAgrees/string_null.a` | PASS |
| `TestParameterDefaultNullAgrees/unknown_param.a` | PASS |
| `TestParameterDefaultNullAgrees/number_null.a` | PASS |
| `TestParameterDefaultNullAgrees/maybe_control.a` | PASS |
| `TestDestructureUnionAgrees` | PASS |
| `TestDestructureUnionAgrees/object` | PASS |
| `TestDestructureUnionAgrees/object/wasi` | SKIP |
| `TestDestructureUnionAgrees/object/javascript` | PASS |
| `TestDestructureUnionAgrees/object/native` | PASS |
| `TestDestructureUnionAgrees/object/sanitized` | PASS |
| `TestDestructureUnionAgrees/semver` | PASS |
| `TestDestructureUnionAgrees/semver/wasi` | SKIP |
| `TestDestructureUnionAgrees/semver/javascript` | PASS |
| `TestDestructureUnionAgrees/semver/native` | PASS |
| `TestDestructureUnionAgrees/semver/sanitized` | PASS |
| `TestDestructureUnionAgrees/null` | PASS |
| `TestDestructureUnionAgrees/null/wasi` | SKIP |
| `TestDestructureUnionAgrees/null/javascript` | PASS |
| `TestDestructureUnionAgrees/null/native` | PASS |
| `TestDestructureUnionAgrees/null/sanitized` | PASS |
| `TestDestructureUnionAgrees/former_stops` | PASS |
| `TestDestructureUnionAgrees/former_stops/native` | PASS |
| `TestDestructureUnionAgrees/former_stops/wasi` | SKIP |
| `TestDestructureUnionAgrees/former_stops/javascript` | PASS |
| `TestDestructureUnionAgrees/former_stops/sanitized` | PASS |

## 3. Mutant (b)

Asserted that this exact old text occurred **once** in `internal/lower/expression.go` before replacing it:

```go
return l.includesNull(concrete) || concrete.Flags()&(checker.TypeFlagsUnknown|checker.TypeFlagsAny|checker.TypeFlagsInstantiable) != 0
```

Replacement:

```go
return l.includesNull(concrete)
```

Ran:

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestParameterDefaultNullAgrees/unknown_param.a' -count=1 -v
```

FAIL, exit 1. Red subtest: `TestParameterDefaultNullAgrees/unknown_param.a`; parent `TestParameterDefaultNullAgrees` also reported FAIL. The fixture lowered and reached the comparison, so this mutant has a witness. All three backend comparisons went red. Complete error lines, with the first error first:

```text
parameter_default_null_test.go:32: native: stdout differs, stderr ""
parameter_default_null_test.go:32: sanitized: stdout differs, stderr ""
parameter_default_null_test.go:32: javascript: stdout differs, stderr ""
```

The subtest took `34.05s`; package time was `34.129s`. Restored `internal/lower/expression.go` from the tested HEAD, then `git diff --exit-code` exited 0 with no output before the full lower suite.

## 4. Full lower suite

```sh
go test ./internal/lower -count=1 -timeout 90m > /tmp/lower.log 2>&1
```

PASS, exit **0**. Wall time, measured around the complete command with a monotonic clock: **1482.792 seconds (24m 42.792s)**. Every failing test and its first error line: **none**.

Complete `/tmp/lower.log`:

```text
ok  	github.com/system-inc/adamic/internal/lower	1459.727s
```

## Delivery

Created `proof/lower-mhn-report` from the tested HEAD after all checks completed. This report is the only file to be committed and pushed, and the only push target is `origin` branch `proof/lower-mhn-report`.
