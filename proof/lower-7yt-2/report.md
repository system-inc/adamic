# Compiler proof — #7yt238q

Date: 2026-11-11 (UTC). Repository: Adamic, TypeScript-to-C compiler.

HEAD: `bc5630c02ff3c1529d029f4f2f54c333d664640a` (`compiler/void-undefined-2`). Base/main: `efe887763f26e96653546fae904d57539139f7ae`. HEAD is one commit above base.

## Decisive results

The proof is **not fully green**. All twelve void-union behavior pins pass, and the documented mutant makes exactly the eight intended behavior subtests fail. Two new refusal pins (`void_field.a`, `void_field.ts`) expect H2 but receive an earlier `adamic/no-void-widening` refusal. That same refusal reproduces against main’s unchanged lowerer when supplied HEAD’s new tests and fixtures.

Formatting, vet, the complete lower suite, and the complete IR suite pass. The full oracle suite fails on those two refusal pins plus environment failures and one transient TSan timeout. The TSan test passes alone at both revisions. The completed full admission delta passes with no admissions, new refusals, stderr differences, or timeouts.

| Step | Result | Wall seconds |
| --- | --- | ---: |

| 1: vet | exit 0 | 349.58 |

| 2: uncached pins | exit 1 | 318.25 |

| 3: mutant pins | exit 1 | 83.84 |

| 4: full lower | exit 0 | 1717.27 |

| 4b: full IR | exit 0 | 147.91 |

| 5: full oracle | exit 1 | 4272.56 |

| 6: completed admission delta | exit 0 | 1453.28 |

## Environment and provisioning

Loaded `/workspace/adamic-tools/env.sh` before repository work and sourced it in each proof shell. Its `TMPDIR=/tmp/adamic-gate` directory did not exist, so it was created outside the repository. All scripts, raw logs, timing records, and dependency installations for this proof live outside the repository under `/tmp/lower-7yt-2`, apart from the two requested full-suite logs `/tmp/lower.log` and `/tmp/oracle.log`.

Provisioned the upstream Microsoft TypeScript Git checkout outside the repository at `/tmp/lower-7yt-2/TypeScript`, detached at `050880ce59e30b356b686bd3144efe24f875ebc8`, and exported `ADAMIC_TYPESCRIPT_SOURCE` to that path. Copied `stage3/api/package.json` and `package-lock.json` to `/tmp/lower-7yt-2/api` and ran `npm ci --ignore-scripts` there. Installed pins: TypeScript 6.0.3 and `@types/node` 25.3.3. Put the external `node_modules/.bin` on PATH; `tsc --version` printed `Version 6.0.3`.

Exported the following for npm and the admission invocation and its inherited children:


```sh
export NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt
export npm_config_cafile=/etc/ssl/certs/ca-certificates.crt
export npm_config_strict_ssl=true
```

The gate’s `toolchain.Env()` filters npm’s environment, so an external PATH wrapper for npm re-exported all three values before executing the real npm. TLS verification stayed enabled.

**The wasm32 leg did not run:** no wasm32-wasi sysroot was found in the installed tools or standard SDK/sysroot locations (including a directory search through `/workspace/adamic-tools`, `/usr/share`, `/opt`, and the filesystem’s first five levels). `WASI_SYSROOT` and `ADAMIC_ORACLE_WASI` were therefore not set. Each of the twelve behavior subtests explicitly logged `wasi leg not run ... WASI_SYSROOT is unset`. The eight refusal subtests stop at lowering and contain no wasm32 backend execution. This applies to both steps 2 and 3.

One heavy command ran at a time throughout. In particular, admission ran alone; only lightweight status/evidence reads accompanied it. `/usr/bin/time` was absent, so wall times were measured by an external Python wrapper using `time.monotonic()` around each full command. An initial vet launch failed before invoking vet because `/usr/bin/time` was missing; the recorded vet run below is the completed replacement.

## 1. Checkout, formatting, vet


```sh
git fetch origin compiler/void-undefined-2
git checkout --detach bc5630c02ff3c1529d029f4f2f54c333d664640a
git submodule update --init --recursive
gofmt -l internal/lower internal/oracle
go vet ./internal/lower ./internal/oracle
```

Fetch, detached checkout, and recursive submodules all exited 0. Submodule pins: `cohere` at `5360753e19e4417675a90e92a1823bea66b13918`, its `TypeScript` at `d92d9bfee114c80be2c375d72edae966176e3a4f`. No AGENTS.md files were found.

`gofmt -l` printed nothing (zero-byte log). Vet: exit 0; wall 349.58 seconds; no diagnostics.

## 2. Uncached pins

Aggregate parents: `TestVoidMemberCountsAsUndefined` PASS; `TestVoidResultReadStillRefused` FAIL.


```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestVoidMemberCountsAsUndefined|TestVoidResultReadStillRefused' -count=1 -v -timeout 40m
```

Result: exit 1; wall 318.25 seconds. No subtests skipped. The behavior pins compare pinned Node output with release native, sanitized native, and JavaScript backends, and perform the leak check.

| Subtest | Result | wasm32 leg |
| --- | --- | --- |

| `TestVoidMemberCountsAsUndefined/alias.a` | PASS | No — no sysroot |

| `TestVoidMemberCountsAsUndefined/alias.ts` | PASS | No — no sysroot |

| `TestVoidMemberCountsAsUndefined/control.a` | PASS | No — no sysroot |

| `TestVoidMemberCountsAsUndefined/control.ts` | PASS | No — no sysroot |

| `TestVoidMemberCountsAsUndefined/field.a` | PASS | No — no sysroot |

| `TestVoidMemberCountsAsUndefined/field.ts` | PASS | No — no sysroot |

| `TestVoidMemberCountsAsUndefined/field_alias.a` | PASS | No — no sysroot |

| `TestVoidMemberCountsAsUndefined/field_alias.ts` | PASS | No — no sysroot |

| `TestVoidMemberCountsAsUndefined/nullish.a` | PASS | No — no sysroot |

| `TestVoidMemberCountsAsUndefined/nullish.ts` | PASS | No — no sysroot |

| `TestVoidMemberCountsAsUndefined/witness.a` | PASS | No — no sysroot |

| `TestVoidMemberCountsAsUndefined/witness.ts` | PASS | No — no sysroot |

| `TestVoidResultReadStillRefused/void_argument.a` | PASS | No — lowering refusal test; no backend leg |

| `TestVoidResultReadStillRefused/void_argument.ts` | PASS | No — lowering refusal test; no backend leg |

| `TestVoidResultReadStillRefused/void_binding.a` | PASS | No — lowering refusal test; no backend leg |

| `TestVoidResultReadStillRefused/void_binding.ts` | PASS | No — lowering refusal test; no backend leg |

| `TestVoidResultReadStillRefused/void_field.a` | FAIL | No — lowering refusal test; no backend leg |

| `TestVoidResultReadStillRefused/void_field.ts` | FAIL | No — lowering refusal test; no backend leg |

| `TestVoidResultReadStillRefused/void_read.a` | PASS | No — lowering refusal test; no backend leg |

| `TestVoidResultReadStillRefused/void_read.ts` | PASS | No — lowering refusal test; no backend leg |

### Every failure: full got and want

For each of these two subtests the complete assertion requires `errors.As(err, *lower.Refused)` to succeed, `Refused.What == "a read of a void value, as a value"`, and `Refused.Fix` to start with `"call it as a statement, or give the function a return type"`. The full observed diagnostic and the test’s printed want follow:

#### `TestVoidResultReadStillRefused/void_field.a`


```text
void_undefined_test.go:97: want H2's void read refused as a value, got /workspace/adamic/internal/oracle/testdata/void_undefined/void_field.a:4:24: Adamic 0.1 refuses a void value at the initializer site through .value seen as void | B | null; declare the function's real return type (adamic/no-void-widening)
```

#### `TestVoidResultReadStillRefused/void_field.ts`


```text
void_undefined_test.go:97: want H2's void read refused as a value, got /workspace/adamic/internal/oracle/testdata/void_undefined/void_field.ts:4:24: Adamic 0.1 refuses a void value at the initializer site through .value seen as void | B | null; declare the function's real return type (adamic/no-void-widening)
```

## 3. Documented mutant

Aggregate parents under the mutant: `TestVoidMemberCountsAsUndefined` FAIL; `TestVoidResultReadStillRefused` FAIL.

`internal/lower/void-undefined.mutants.md` contains one mutant, applied to `internal/lower/expression.go`. Asserted that the following old text occurred **exactly once** before replacing it:


```go
		if member.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 || l.phantomUndefined(member) {
```

Replacement:


```go
		if member.Flags()&checker.TypeFlagsUndefined != 0 || l.phantomUndefined(member) {
```

Reran the exact step 2 command, including `ADAMIC_GATE_UNCACHED=1` and `-timeout 40m`. Result: exit 1; wall 83.84 seconds. Eight behavior subtests turned red: witness, alias, field, and field_alias in both formats. Nullish and control remained green. The two original void_field failures remained red; the other six refusal subtests stayed green. Original source bytes were restored in a `finally` block, and `git diff --exit-code` exited 0.

| Subtest | Mutant result | Change from unmutated | wasm32 leg |
| --- | --- | --- | --- |

| `TestVoidMemberCountsAsUndefined/alias.a` | FAIL | Turned red | No — no sysroot |

| `TestVoidMemberCountsAsUndefined/alias.ts` | FAIL | Turned red | No — no sysroot |

| `TestVoidMemberCountsAsUndefined/control.a` | PASS | Unchanged | No — no sysroot |

| `TestVoidMemberCountsAsUndefined/control.ts` | PASS | Unchanged | No — no sysroot |

| `TestVoidMemberCountsAsUndefined/field.a` | FAIL | Turned red | No — no sysroot |

| `TestVoidMemberCountsAsUndefined/field.ts` | FAIL | Turned red | No — no sysroot |

| `TestVoidMemberCountsAsUndefined/field_alias.a` | FAIL | Turned red | No — no sysroot |

| `TestVoidMemberCountsAsUndefined/field_alias.ts` | FAIL | Turned red | No — no sysroot |

| `TestVoidMemberCountsAsUndefined/nullish.a` | PASS | Unchanged | No — no sysroot |

| `TestVoidMemberCountsAsUndefined/nullish.ts` | PASS | Unchanged | No — no sysroot |

| `TestVoidMemberCountsAsUndefined/witness.a` | FAIL | Turned red | No — no sysroot |

| `TestVoidMemberCountsAsUndefined/witness.ts` | FAIL | Turned red | No — no sysroot |

| `TestVoidResultReadStillRefused/void_argument.a` | PASS | Unchanged | No — lowering refusal test |

| `TestVoidResultReadStillRefused/void_argument.ts` | PASS | Unchanged | No — lowering refusal test |

| `TestVoidResultReadStillRefused/void_binding.a` | PASS | Unchanged | No — lowering refusal test |

| `TestVoidResultReadStillRefused/void_binding.ts` | PASS | Unchanged | No — lowering refusal test |

| `TestVoidResultReadStillRefused/void_field.a` | FAIL | Unchanged | No — lowering refusal test |

| `TestVoidResultReadStillRefused/void_field.ts` | FAIL | Unchanged | No — lowering refusal test |

| `TestVoidResultReadStillRefused/void_read.a` | PASS | Unchanged | No — lowering refusal test |

| `TestVoidResultReadStillRefused/void_read.ts` | PASS | Unchanged | No — lowering refusal test |

### Every red subtest: full observed got

The quoted `Node` stdout in each backend diagnostic is also the pinned want. Each backend observation includes its exit code, complete stdout, and complete stderr. The two refusal failures retain the same full assertion contract described in step 2.

#### `TestVoidMemberCountsAsUndefined/alias.a`


```text
void_undefined_test.go:58: release: stdout differs: exit 0, stdout "true false object\ntrue false object\nfalse false object\ntrue false object\ntrue false object\nfalse false object\n", stderr ""; Node "true false object\nfalse true undefined\nfalse false object\ntrue false object\nfalse true undefined\nfalse false object\n"
void_undefined_test.go:58: sanitized: stdout differs: exit 0, stdout "true false object\ntrue false object\nfalse false object\ntrue false object\ntrue false object\nfalse false object\n", stderr ""; Node "true false object\nfalse true undefined\nfalse false object\ntrue false object\nfalse true undefined\nfalse false object\n"
void_undefined_test.go:58: JavaScript: stdout differs: exit 0, stdout "true false object\nfalse false undefined\nfalse false object\ntrue false object\nfalse false undefined\nfalse false object\n", stderr ""; Node "true false object\nfalse true undefined\nfalse false object\ntrue false object\nfalse true undefined\nfalse false object\n"
```

#### `TestVoidMemberCountsAsUndefined/alias.ts`


```text
void_undefined_test.go:58: release: stdout differs: exit 0, stdout "true false object\ntrue false object\nfalse false object\ntrue false object\ntrue false object\nfalse false object\n", stderr ""; Node "true false object\nfalse true undefined\nfalse false object\ntrue false object\nfalse true undefined\nfalse false object\n"
void_undefined_test.go:58: sanitized: stdout differs: exit 0, stdout "true false object\ntrue false object\nfalse false object\ntrue false object\ntrue false object\nfalse false object\n", stderr ""; Node "true false object\nfalse true undefined\nfalse false object\ntrue false object\nfalse true undefined\nfalse false object\n"
void_undefined_test.go:58: JavaScript: stdout differs: exit 0, stdout "true false object\nfalse false undefined\nfalse false object\ntrue false object\nfalse false undefined\nfalse false object\n", stderr ""; Node "true false object\nfalse true undefined\nfalse false object\ntrue false object\nfalse true undefined\nfalse false object\n"
```

#### `TestVoidMemberCountsAsUndefined/field.a`


```text
void_undefined_test.go:58: sanitized: stdout differs: exit 0, stdout "true false object none\ntrue false object\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\ntrue false object none\ntrue false object\ntrue false object none\ntrue false object\n", stderr ""; Node "false true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\nfalse true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\n"
void_undefined_test.go:58: JavaScript: stdout differs: exit 0, stdout "false false undefined none\nfalse false undefined\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\nfalse false undefined none\nfalse false undefined\ntrue false object none\ntrue false object\n", stderr ""; Node "false true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\nfalse true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\n"
void_undefined_test.go:58: release: stdout differs: exit 0, stdout "true false object none\ntrue false object\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\ntrue false object none\ntrue false object\ntrue false object none\ntrue false object\n", stderr ""; Node "false true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\nfalse true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\n"
```

#### `TestVoidMemberCountsAsUndefined/field.ts`


```text
void_undefined_test.go:58: release: stdout differs: exit 0, stdout "true false object none\ntrue false object\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\ntrue false object none\ntrue false object\ntrue false object none\ntrue false object\n", stderr ""; Node "false true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\nfalse true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\n"
void_undefined_test.go:58: sanitized: stdout differs: exit 0, stdout "true false object none\ntrue false object\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\ntrue false object none\ntrue false object\ntrue false object none\ntrue false object\n", stderr ""; Node "false true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\nfalse true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\n"
void_undefined_test.go:58: JavaScript: stdout differs: exit 0, stdout "false false undefined none\nfalse false undefined\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\nfalse false undefined none\nfalse false undefined\ntrue false object none\ntrue false object\n", stderr ""; Node "false true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\nfalse true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\n"
```

#### `TestVoidMemberCountsAsUndefined/field_alias.a`


```text
void_undefined_test.go:58: release: stdout differs: exit 0, stdout "true false object none\ntrue false object\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\ntrue false object none\ntrue false object\ntrue false object none\ntrue false object\n", stderr ""; Node "false true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\nfalse true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\n"
void_undefined_test.go:58: sanitized: stdout differs: exit 0, stdout "true false object none\ntrue false object\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\ntrue false object none\ntrue false object\ntrue false object none\ntrue false object\n", stderr ""; Node "false true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\nfalse true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\n"
void_undefined_test.go:58: JavaScript: stdout differs: exit 0, stdout "false false undefined none\nfalse false undefined\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\nfalse false undefined none\nfalse false undefined\ntrue false object none\ntrue false object\n", stderr ""; Node "false true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\nfalse true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\n"
```

#### `TestVoidMemberCountsAsUndefined/field_alias.ts`


```text
void_undefined_test.go:58: JavaScript: stdout differs: exit 0, stdout "false false undefined none\nfalse false undefined\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\nfalse false undefined none\nfalse false undefined\ntrue false object none\ntrue false object\n", stderr ""; Node "false true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\nfalse true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\n"
void_undefined_test.go:58: release: stdout differs: exit 0, stdout "true false object none\ntrue false object\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\ntrue false object none\ntrue false object\ntrue false object none\ntrue false object\n", stderr ""; Node "false true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\nfalse true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\n"
void_undefined_test.go:58: sanitized: stdout differs: exit 0, stdout "true false object none\ntrue false object\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\ntrue false object none\ntrue false object\ntrue false object none\ntrue false object\n", stderr ""; Node "false true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\nfalse false object 2\nfalse false object\nfalse true undefined none\nfalse true undefined\ntrue false object none\ntrue false object\n"
```

#### `TestVoidMemberCountsAsUndefined/witness.a`


```text
void_undefined_test.go:58: sanitized: stdout differs: exit 0, stdout "true false true object\ntrue false true object\nfalse false false object\n", stderr ""; Node "true false true object\nfalse true true undefined\nfalse false false object\n"
void_undefined_test.go:58: JavaScript: stdout differs: exit 0, stdout "true false true object\nfalse false true undefined\nfalse false false object\n", stderr ""; Node "true false true object\nfalse true true undefined\nfalse false false object\n"
void_undefined_test.go:58: release: stdout differs: exit 0, stdout "true false true object\ntrue false true object\nfalse false false object\n", stderr ""; Node "true false true object\nfalse true true undefined\nfalse false false object\n"
```

#### `TestVoidMemberCountsAsUndefined/witness.ts`


```text
void_undefined_test.go:58: release: stdout differs: exit 0, stdout "true false true object\ntrue false true object\nfalse false false object\n", stderr ""; Node "true false true object\nfalse true true undefined\nfalse false false object\n"
void_undefined_test.go:58: sanitized: stdout differs: exit 0, stdout "true false true object\ntrue false true object\nfalse false false object\n", stderr ""; Node "true false true object\nfalse true true undefined\nfalse false false object\n"
void_undefined_test.go:58: JavaScript: stdout differs: exit 0, stdout "true false true object\nfalse false true undefined\nfalse false false object\n", stderr ""; Node "true false true object\nfalse true true undefined\nfalse false false object\n"
```

#### `TestVoidResultReadStillRefused/void_field.a`


```text
void_undefined_test.go:97: want H2's void read refused as a value, got /workspace/adamic/internal/oracle/testdata/void_undefined/void_field.a:4:24: Adamic 0.1 refuses a void value at the initializer site through .value seen as void | B | null; declare the function's real return type (adamic/no-void-widening)
```

#### `TestVoidResultReadStillRefused/void_field.ts`


```text
void_undefined_test.go:97: want H2's void read refused as a value, got /workspace/adamic/internal/oracle/testdata/void_undefined/void_field.ts:4:24: Adamic 0.1 refuses a void value at the initializer site through .value seen as void | B | null; declare the function's real return type (adamic/no-void-widening)
```

## 4. Complete lower suite


```sh
go test ./internal/lower -count=1 -timeout 90m > /tmp/lower.log 2>&1
```

Result: exit 0; wall 1717.27 seconds. No failing tests, environment or otherwise. No baseline reruns required.


```text
ok  	github.com/system-inc/adamic/internal/lower	1686.989s
```

## 4b. Complete IR suite and source guards


```sh
go test ./internal/ir -count=1 -timeout 30m
```

Result: exit 0; wall 147.91 seconds. No failures. This complete suite includes `TestCallTargetReadersLower`, the call-target export/archive guards, and IR field/read guards including `TestEveryCalleeFieldReadGoesThroughMethodCallee`.


```text
ok  	github.com/system-inc/adamic/internal/ir	116.082s
```

## 5. Complete oracle suite


```sh
go test ./internal/oracle -count=1 -timeout 120m > /tmp/oracle.log 2>&1
```

Result: exit 1; wall 4272.56 seconds. The test package itself reported 4265.390 seconds. The list below includes every printed failing test and subtest, including aggregate parents. Cache/build informational lines are omitted when identifying the first error line.

### `TestParallelBenchmark_tsan_One_1`

TRANSIENT TIMEOUT / ENVIRONMENT — initial native run exceeded its 60-second deadline. Passed alone at both main and HEAD; see reruns below. Parent entries aggregate the same leaf failure.

First error line:


```text
parallel_test.go:119: exit codes differ: Node exit 0 stdout "4096 files, 20282626 UTF-16 units, 2621440 tokens, 556992 distinct per file, ordered digest 823772603\n" stderr ""; native exit -1 stdout "" stderr ""
```

### `TestParallelBenchmark_tsan_One_1/tsan`

TRANSIENT TIMEOUT / ENVIRONMENT — initial native run exceeded its 60-second deadline. Passed alone at both main and HEAD; see reruns below. Parent entries aggregate the same leaf failure.

First error line:


```text
parallel_test.go:119: exit codes differ: Node exit 0 stdout "4096 files, 20282626 UTF-16 units, 2621440 tokens, 556992 distinct per file, ordered digest 823772603\n" stderr ""; native exit -1 stdout "" stderr ""
```

### `TestParallelBenchmark_tsan_One_1/tsan/1`

TRANSIENT TIMEOUT / ENVIRONMENT — initial native run exceeded its 60-second deadline. Passed alone at both main and HEAD; see reruns below. Parent entries aggregate the same leaf failure.

First error line:


```text
parallel_test.go:119: exit codes differ: Node exit 0 stdout "4096 files, 20282626 UTF-16 units, 2621440 tokens, 556992 distinct per file, ordered digest 823772603\n" stderr ""; native exit -1 stdout "" stderr ""
```

### `TestStdoutColumnsRefusesWASI`

ENVIRONMENT — wasm32-wasi requires the unavailable sysroot; no main rerun required.

First error line:


```text
library_node_stdout_columns_test.go:246: want typed, named WASI refusal, got native: wasm32-wasi requires WASI_SYSROOT
```

### `TestAssignmentProofWiderTargetMutant`

ENVIRONMENT — missing `typescript` module at the test’s hardcoded repository-local NODE_PATH; no main rerun required.

First error line:


```text
assignment_proofs_test.go:202: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
```

### `TestAssignmentProofWiderTarget`

ENVIRONMENT — missing `typescript` module at the test’s hardcoded repository-local NODE_PATH; no main rerun required.

First error line:


```text
assignment_proofs_test.go:197: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
```

### `TestAssignmentProofCompoundMutant`

ENVIRONMENT — missing `typescript` module at the test’s hardcoded repository-local NODE_PATH; no main rerun required.

First error line:


```text
assignment_proofs_test.go:192: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
```

### `TestAssignmentProofCompound`

ENVIRONMENT — missing `typescript` module at the test’s hardcoded repository-local NODE_PATH; no main rerun required.

First error line:


```text
assignment_proofs_test.go:187: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
```

### `TestAssignmentProofChained`

ENVIRONMENT — missing `typescript` module at the test’s hardcoded repository-local NODE_PATH; no main rerun required.

First error line:


```text
assignment_proofs_test.go:177: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
```

### `TestAssignmentProofChainedMutant`

ENVIRONMENT — missing `typescript` module at the test’s hardcoded repository-local NODE_PATH; no main rerun required.

First error line:


```text
assignment_proofs_test.go:182: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
```

### `TestAssignmentProofReturnDefinedMutant`

ENVIRONMENT — missing `typescript` module at the test’s hardcoded repository-local NODE_PATH; no main rerun required.

First error line:


```text
assignment_proofs_test.go:172: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
```

### `TestAssignmentProofReturnDefined`

ENVIRONMENT — missing `typescript` module at the test’s hardcoded repository-local NODE_PATH; no main rerun required.

First error line:


```text
assignment_proofs_test.go:167: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
```

### `TestAssignmentProofIfNarrowingMutant`

ENVIRONMENT — missing `typescript` module at the test’s hardcoded repository-local NODE_PATH; no main rerun required.

First error line:


```text
assignment_proofs_test.go:162: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
```

### `TestAssignmentProofIfNarrowing`

ENVIRONMENT — missing `typescript` module at the test’s hardcoded repository-local NODE_PATH; no main rerun required.

First error line:


```text
assignment_proofs_test.go:157: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
```

### `TestAssignmentProofScannerKeywordMutant`

ENVIRONMENT — missing `typescript` module at the test’s hardcoded repository-local NODE_PATH; no main rerun required.

First error line:


```text
assignment_proofs_test.go:152: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
```

### `TestAssignmentProofScannerKeyword`

ENVIRONMENT — missing `typescript` module at the test’s hardcoded repository-local NODE_PATH; no main rerun required.

First error line:


```text
assignment_proofs_test.go:147: stock TypeScript: exit status 1, node:internal/modules/cjs/loader:1520
```

### `TestVoidResultReadStillRefused`

NON-ENVIRONMENT — refusal assertion mismatch; reproduces against main’s unchanged lowerer with the new test supplied. Parent entry aggregates the two leaf failures.

First error line:


```text
void_undefined_test.go:97: want H2's void read refused as a value, got /workspace/adamic/internal/oracle/testdata/void_undefined/void_field.ts:4:24: Adamic 0.1 refuses a void value at the initializer site through .value seen as void | B | null; declare the function's real return type (adamic/no-void-widening)
```

### `TestVoidResultReadStillRefused/void_field.ts`

NON-ENVIRONMENT — refusal assertion mismatch; reproduces against main’s unchanged lowerer with the new test supplied. Parent entry aggregates the two leaf failures.

First error line:


```text
void_undefined_test.go:97: want H2's void read refused as a value, got /workspace/adamic/internal/oracle/testdata/void_undefined/void_field.ts:4:24: Adamic 0.1 refuses a void value at the initializer site through .value seen as void | B | null; declare the function's real return type (adamic/no-void-widening)
```

### `TestVoidResultReadStillRefused/void_field.a`

NON-ENVIRONMENT — refusal assertion mismatch; reproduces against main’s unchanged lowerer with the new test supplied. Parent entry aggregates the two leaf failures.

First error line:


```text
void_undefined_test.go:97: want H2's void read refused as a value, got /workspace/adamic/internal/oracle/testdata/void_undefined/void_field.a:4:24: Adamic 0.1 refuses a void value at the initializer site through .value seen as void | B | null; declare the function's real return type (adamic/no-void-widening)
```

The twelve assignment-proof errors are all `Error: Cannot find module 'typescript'`, required by `stage3/fixtures/assignment-proofs/check-types.cjs`. `assignmentProofTypecheck` explicitly replaces `NODE_PATH` with `<repository>/stage3/api/node_modules`, so the external pinned installation and PATH do not satisfy that hardcoded module lookup. No repository-local package installation or compiler fix was made to hide these failures.

The TSan leaf failure’s entire native observation was exit -1, stdout `""`, stderr `""` after about 60.05 seconds, versus Node exit 0 and the full expected benchmark stdout shown above. It did not reproduce alone.

### Main comparisons and isolated timeout reruns

Created `/tmp/lower-7yt-2/main` as a detached worktree at the exact base SHA. Its two submodule checkouts use local shared object stores and the same pinned SHAs; no baseline compiler source was changed.

TSan, alone at main: exit 0; wall 256.88 seconds. Alone at HEAD: exit 0; wall 28.24 seconds. Both passed, with native attempts 10.630295501 seconds and 10.827125188 seconds respectively.


```sh
go test ./internal/oracle -run ^TestParallelBenchmark_tsan_One_1$ -count=1 -v -timeout 120m
```


```text
=== RUN   TestParallelBenchmark_tsan_One_1
=== PAUSE TestParallelBenchmark_tsan_One_1
=== CONT  TestParallelBenchmark_tsan_One_1
    parallel_test.go:354: gate cache node miss
    parallel_test.go:354: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 2565158e7096 miss 27.41
=== RUN   TestParallelBenchmark_tsan_One_1/tsan
=== RUN   TestParallelBenchmark_tsan_One_1/tsan/1
    parallel_test.go:126: race variant runs=1 elapsed=10.630295501s
--- PASS: TestParallelBenchmark_tsan_One_1 (49.73s)
    --- PASS: TestParallelBenchmark_tsan_One_1/tsan (13.03s)
        --- PASS: TestParallelBenchmark_tsan_One_1/tsan/1 (10.63s)
PASS
gate cache: native hits=0 misses=0
gate cache: node hits=0 misses=2
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	49.814s
```


```sh
go test ./internal/oracle -run ^TestParallelBenchmark_tsan_One_1$ -count=1 -v -timeout 120m
```


```text
=== RUN   TestParallelBenchmark_tsan_One_1
=== PAUSE TestParallelBenchmark_tsan_One_1
=== CONT  TestParallelBenchmark_tsan_One_1
    parallel_test.go:354: gate cache node miss
    parallel_test.go:354: gate cache node miss
    oracle_test.go:638: build go_build_./oracle/json_types.go_json-types 440b9f1ed3d6 hit 0.54
=== RUN   TestParallelBenchmark_tsan_One_1/tsan
=== RUN   TestParallelBenchmark_tsan_One_1/tsan/1
    parallel_test.go:126: race variant runs=1 elapsed=10.827125188s
--- PASS: TestParallelBenchmark_tsan_One_1 (21.12s)
    --- PASS: TestParallelBenchmark_tsan_One_1/tsan (12.82s)
        --- PASS: TestParallelBenchmark_tsan_One_1/tsan/1 (10.83s)
PASS
gate cache: native hits=0 misses=0
gate cache: node hits=0 misses=2
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	21.144s
```

The new `TestVoidResultReadStillRefused` and its fixtures do not exist in main. To execute the actual failing assertions against main, copied only HEAD’s new `internal/oracle/void_undefined_test.go` and `internal/oracle/testdata/void_undefined/` into that external worktree. Main’s lowerer and other compiler source stayed at the base SHA. Ran:


```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestVoidResultReadStillRefused$/^void_field[.](a|ts)$' -count=1 -v -timeout 40m
```

Result: exit 1; wall 27.62 seconds. Both subtests fail there too, with the same diagnostic except for checkout path. Full baseline output:


```text
=== RUN   TestVoidResultReadStillRefused
=== PAUSE TestVoidResultReadStillRefused
=== CONT  TestVoidResultReadStillRefused
=== RUN   TestVoidResultReadStillRefused/void_field.a
=== PAUSE TestVoidResultReadStillRefused/void_field.a
=== RUN   TestVoidResultReadStillRefused/void_field.ts
=== PAUSE TestVoidResultReadStillRefused/void_field.ts
=== CONT  TestVoidResultReadStillRefused/void_field.a
=== CONT  TestVoidResultReadStillRefused/void_field.ts
=== NAME  TestVoidResultReadStillRefused/void_field.a
    void_undefined_test.go:97: want H2's void read refused as a value, got /tmp/lower-7yt-2/main/internal/oracle/testdata/void_undefined/void_field.a:4:24: Adamic 0.1 refuses a void value at the initializer site through .value seen as void | B | null; declare the function's real return type (adamic/no-void-widening)
=== NAME  TestVoidResultReadStillRefused/void_field.ts
    void_undefined_test.go:97: want H2's void read refused as a value, got /tmp/lower-7yt-2/main/internal/oracle/testdata/void_undefined/void_field.ts:4:24: Adamic 0.1 refuses a void value at the initializer site through .value seen as void | B | null; declare the function's real return type (adamic/no-void-widening)
--- FAIL: TestVoidResultReadStillRefused (0.00s)
    --- FAIL: TestVoidResultReadStillRefused/void_field.a (6.47s)
    --- FAIL: TestVoidResultReadStillRefused/void_field.ts (6.47s)
FAIL
gate cache: native hits=0 misses=0
gate cache: node hits=0 misses=2
gate cache: probe hits=0 misses=0
FAIL	github.com/system-inc/adamic/internal/oracle	6.486s
FAIL
```

## 6. Admission delta

The first admission invocation was interrupted during an environment reconnect: the environment was marked starting again, its process subsequently appeared defunct, its tool session was gone, and its JSON result was empty. No exit status or verdict was captured for that incomplete attempt. Preserved its partial log separately and reran the exact command alone under an external supervisor. The following results are from the completed rerun, not from the interrupted attempt.


```sh
python3 cloud/admission-corpus/manifest.py --sha bc5630c02ff3c1529d029f4f2f54c333d664640a > /tmp/lower-7yt-2/manifest.json
go run ./cmd/adamic-admission-delta --base efe887763f26e96653546fae904d57539139f7ae --head bc5630c02ff3c1529d029f4f2f54c333d664640a --manifest /tmp/lower-7yt-2/manifest.json --manifest-generator cloud/admission-corpus/manifest.py --workers 2 --json
```

Result: exit 0; wall 1453.28 seconds. Started 2026-11-11T14:58:45.057373+00:00. Full head manifest, no corpus filter, no shard, no omitted programs, no runtime budget; workers 2. The manifest contains 5253 unique programs: witnesses 2, fixtures 4305, gaps 105, review 20, fuzz 0, stage3 821. The diff corpus identifies 20 changed programs; all changed `.ts` programs are matched (`diff_ts_unmatched=[]`).

| Measure | Result |
| --- | --- |

| Exit | 0 |

| Verdict | pass |

| Void | False |

| Void reasons | None — neither global nor per-program |

| diff_count | 20 |

| Admitted | 0 |

| Newly refused | 0 |

| Accepted by both | 3294 |

| Refused by both | 1959 |

| Programs with classification stderr differences | 0 |

| Classification timeouts | 0 |

| Runtime timeout runs | 0 |

| Retried runs | 0 |

| Still timed out | 0 |

| Negative witnesses accepted | 50/50 |

| Ruled runtime stops | 52 |

| Node agreements | 13 |

**Every classification stderr difference:** none. All 5253 programs have byte-identical `base_compile.stderr` and `compile.stderr`, so there are no differing text pairs to list.

**Timeout reruns:** none required. Checked all classification and runtime observation error fields plus per-program `timed_out` records: no timeouts at either SHA. The separate full-suite TSan timeout was rerun alone at both SHAs as documented in step 5.

Gate stderr:


```text
negative witnesses: 50
ruled run-time stops: 52
ruled backend divergences: 0
runtime timeouts: 0 runs timed out, 0 retried alone, 0 timed out again
changed programs: 20
```

Phase timing and manifest identity:


```json
{
  "phase_seconds": {
    "base_build": 0.003011652,
    "checkout": 15.688807743,
    "checkout_and_provision": 16.635178375,
    "classification": 789.753859272,
    "head_build": 0.00033517,
    "manifest_and_verification": 25.041767241,
    "provision": 0.946342112,
    "runtime": 620.616519247,
    "total": 1452.0616677
  },
  "generator_revision": "bc5630c02ff3c1529d029f4f2f54c333d664640a",
  "generator_blob": "09d061a03b1e016387e975d25ee0cd5b75ac6c1f",
  "manifest_blob": "1b0830ed8963fb6a43e5b48846577df002ae2342",
  "complete_corpus": true,
  "workers": 2,
  "classification": "lowering-only",
  "diff_count": 20,
  "diff_ts_unmatched": [],
  "admitted": 0,
  "verdict": "pass",
  "void": false,
  "timed_out_runs": 0,
  "retried_runs": 0,
  "still_timed_out_runs": 0
}
```

## Evidence and publication scope

Raw evidence remains outside the repository. SHA-256 digests:

| Evidence | SHA-256 |
| --- | --- |

| `/tmp/lower-7yt-2/pins.log` | `44361e2530b6bffb07c5c7dad65290056731a6851c124224d006b65b6d0c3a3b` |

| `/tmp/lower-7yt-2/mutant.log` | `627d29f192516141d2603c562a3de50b33fbb8651d6670e2d01330e35399152b` |

| `/tmp/lower.log` | `0a7bfb67b782dc5860345959725a2386189a6981199b79b5df9e04376c2ab270` |

| `/tmp/lower-7yt-2/ir.log` | `42c1e1e85fcd83bc3c34a1eb1952bfcadccc383d58a939cf4aa41be3118d6590` |

| `/tmp/oracle.log` | `5f120d94dfbc065f2eccd0dd54405d62dde44f5c3f9ad07e86cc51ed541b49eb` |

| `/tmp/lower-7yt-2/main-tsan.log` | `90763ab37ffe5b2d56804937bf9d276a0bee7a34ce34e439eda4b72bd1729eb6` |

| `/tmp/lower-7yt-2/head-tsan-alone.log` | `a5d451cb2f6af248aa1fd5f17d213799f093d2914f07d589a3c5cfe473a8a93b` |

| `/tmp/lower-7yt-2/main-void-field.log` | `6867b5753e1c4deb6d06d2febda0d959f725e5ecc8e5989156c705387c5a475c` |

| `/tmp/lower-7yt-2/admission.json` | `cb5e64a66aee9e29d6af3e9d6cbdcb50d7f14fd66b7406a7499d0aa68d408c9e` |

| `/tmp/lower-7yt-2/admission.log` | `ca843cc0add975bab734c776cafa4d943c7eefbf88d3edf537f51f93ef268fc8` |

The original proof checkout was clean at the exact HEAD after restoring the mutant and again after all test/admission work. This report is the only repository file written for publication. The requested branch `proof/lower-7yt-2-report` is created from the tested HEAD; only `proof/lower-7yt-2/report.md` is staged and committed. Only that report branch is pushed to origin. No compiler fix, dependency installation, logs, baseline test copies, or other file is committed.
