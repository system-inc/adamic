# Agreement check for #fj0vqtf

Date: 2026-11-12 (UTC).
Repository: Adamic, TypeScript-to-C compiler.
Checked HEAD: `68f42e36d31ee81c853adefb9bf37e615d460835`, detached after fetching `origin compiler/host-member-value-0`.
Parent: `d106f8290c0f11f372107f26dfe089c1fb8ce4fc`.
Report branch: `proof/lower-fj0w2-1-report`, created from checked HEAD.

## Results

- `git fetch origin compiler/host-member-value-0`: exit 0.
- Detached checkout of the requested HEAD: exit 0.
- `git submodule update --init --recursive`: exit 0. `cohere` is at `5360753e19e4417675a90e92a1823bea66b13918`; `cohere/TypeScript` is at `d92d9bfee114c80be2c375d72edae966176e3a4f`.
- `gofmt -l internal/oracle`: exit 0; stdout empty.
- `go vet ./internal/oracle`: exit 0; stdout and stderr empty.
- Uncached agreement test: exit 1; exactly 22 of 22 subtests failed against the intentional `unmeasured` expectations. Every actual stop is reproduced in full below.
- Node oracle: 11 of 11 programs exited 0; all stderr streams empty. Some programs catch exceptions and print them to stdout.
- TypeScript: the 11 original invocations from the repository root returned TS5112. The full programs were then checked using the same flags and absolute file paths from the external scratch directory: 11 of 11 clean.

## Environment and execution

Loaded `/workspace/adamic-tools/env.sh` before work and in the setup and check execution shells. Created `/tmp/adamic-gate` outside the repository for its configured `TMPDIR`. Heavy commands ran sequentially, with no overlapping heavy commands.

Provisioned upstream from `https://github.com/microsoft/TypeScript.git` into `/tmp/fj0vqtf-typescript`, detached at the verified pin `050880ce59e30b356b686bd3144efe24f875ebc8`, and exported `ADAMIC_TYPESCRIPT_SOURCE=/tmp/fj0vqtf-typescript`.

Copied `stage3/api/package.json` and `stage3/api/package-lock.json` into `/tmp/fj0vqtf-api` and ran `npm ci --prefix /tmp/fj0vqtf-api` (exit 0). Added `/tmp/fj0vqtf-api/node_modules/.bin` to PATH. Versions: `tsc` reports `Version 6.0.3`; Node reports `v24.19.0`; Go reports `go version go1.27.1 linux/amd64`.

Exported these variables for npm, the agreement test, and all its child processes, as well as the direct Node and tsc runs:

```sh
NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt
npm_config_cafile=/etc/ssl/certs/ca-certificates.crt
npm_config_strict_ssl=true
```

## Agreement test: all 22 failures verbatim

Command (working directory `/workspace/adamic`):

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestHostMemberValueAgreesOrStops$' -count=1 -v -timeout 40m
```

Exit code: `1`. Stderr: empty (0 bytes). The complete failure line for each subtest follows, including the original indentation and the entire quoted `got` text.

### TestHostMemberValueAgreesOrStops/h2_plain_call.a

```text
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/h2_plain_call.a:13:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: () => host.read() (unbound-method)", want "unmeasured"
```

### TestHostMemberValueAgreesOrStops/h2_plain_call.ts

```text
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/h2_plain_call.ts:13:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: () => host.read() (unbound-method)", want "unmeasured"
```

### TestHostMemberValueAgreesOrStops/h3_reordered_layout.a

```text
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/h3_reordered_layout.a:19:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
```

### TestHostMemberValueAgreesOrStops/h3_reordered_layout.ts

```text
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/h3_reordered_layout.ts:19:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
```

### TestHostMemberValueAgreesOrStops/p1_arrow_plain_call.a

```text
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/p1_arrow_plain_call.a:10:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (suffix) => host.read(suffix) (unbound-method)", want "unmeasured"
```

### TestHostMemberValueAgreesOrStops/p1_arrow_plain_call.ts

```text
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/p1_arrow_plain_call.ts:10:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (suffix) => host.read(suffix) (unbound-method)", want "unmeasured"
```

### TestHostMemberValueAgreesOrStops/r1_holder_run.a

```text
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/r1_holder_run.a:16:31: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: () => host.read() (unbound-method)", want "unmeasured"
```

### TestHostMemberValueAgreesOrStops/r1_holder_run.ts

```text
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/r1_holder_run.ts:16:31: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: () => host.read() (unbound-method)", want "unmeasured"
```

### TestHostMemberValueAgreesOrStops/r2_alias_field.a

```text
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/r2_alias_field.a:13:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: () => host.read() (unbound-method)", want "unmeasured"
```

### TestHostMemberValueAgreesOrStops/r2_alias_field.ts

```text
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/r2_alias_field.ts:13:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: () => host.read() (unbound-method)", want "unmeasured"
```

### TestHostMemberValueAgreesOrStops/r3_apply.a

```text
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/r3_apply.a:15:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: () => host.read() (unbound-method)", want "unmeasured"
```

### TestHostMemberValueAgreesOrStops/r3_apply.ts

```text
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/r3_apply.ts:15:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: () => host.read() (unbound-method)", want "unmeasured"
```

### TestHostMemberValueAgreesOrStops/s1_class_instance.a

```text
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/s1_class_instance.a:11:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
```

### TestHostMemberValueAgreesOrStops/s1_class_instance.ts

```text
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/s1_class_instance.ts:11:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
```

### TestHostMemberValueAgreesOrStops/s2_unknown_reach.a

```text
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/s2_unknown_reach.a:8:19: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
```

### TestHostMemberValueAgreesOrStops/s2_unknown_reach.ts

```text
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/s2_unknown_reach.ts:8:19: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
```

### TestHostMemberValueAgreesOrStops/v1_tsc_shape.a

```text
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/v1_tsc_shape.a:13:22: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
```

### TestHostMemberValueAgreesOrStops/v1_tsc_shape.ts

```text
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/v1_tsc_shape.ts:13:22: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
```

### TestHostMemberValueAgreesOrStops/v2_same_layout_this.a

```text
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/v2_same_layout_this.a:19:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
```

### TestHostMemberValueAgreesOrStops/v2_same_layout_this.ts

```text
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/v2_same_layout_this.ts:19:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
```

### TestHostMemberValueAgreesOrStops/v2b_function_expression_this.a

```text
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/v2b_function_expression_this.a:19:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
```

### TestHostMemberValueAgreesOrStops/v2b_function_expression_this.ts

```text
    host_member_value_test.go:61: stop: got "internal/oracle/testdata/host_member_value/v2b_function_expression_this.ts:19:18: Adamic 0.1 refuses a method read as a value (read would lose its object, and this with it); call it in an arrow that keeps the object: (name) => host.read(name) (unbound-method)", want "unmeasured"
```

## Node oracle: stdout, stderr, exit code

Working directory: `/workspace/adamic`. Each stdout block preserves the captured text, including its final newline. Empty streams are explicitly marked as 0 bytes.

### h2_plain_call.ts

```sh
node --disable-warning=ExperimentalWarning oracle/node.mjs internal/oracle/testdata/host_member_value/h2_plain_call.ts
```

Exit code: `0`.

Stdout (63 bytes):

```text
TypeError
Cannot read properties of undefined (reading 'name')
```

Stderr: empty (0 bytes).

### h3_reordered_layout.ts

```sh
node --disable-warning=ExperimentalWarning oracle/node.mjs internal/oracle/testdata/host_member_value/h3_reordered_layout.ts
```

Exit code: `0`.

Stdout (8 bytes):

```text
other a
```

Stderr: empty (0 bytes).

### p1_arrow_plain_call.ts

```sh
node --disable-warning=ExperimentalWarning oracle/node.mjs internal/oracle/testdata/host_member_value/p1_arrow_plain_call.ts
```

Exit code: `0`.

Stdout (24 bytes):

```text
arrow a
patched arrow b
```

Stderr: empty (0 bytes).

### r1_holder_run.ts

```sh
node --disable-warning=ExperimentalWarning oracle/node.mjs internal/oracle/testdata/host_member_value/r1_holder_run.ts
```

Exit code: `0`.

Stdout (15 bytes):

```text
read undefined
```

Stderr: empty (0 bytes).

### r2_alias_field.ts

```sh
node --disable-warning=ExperimentalWarning oracle/node.mjs internal/oracle/testdata/host_member_value/r2_alias_field.ts
```

Exit code: `0`.

Stdout (15 bytes):

```text
read undefined
```

Stderr: empty (0 bytes).

### r3_apply.ts

```sh
node --disable-warning=ExperimentalWarning oracle/node.mjs internal/oracle/testdata/host_member_value/r3_apply.ts
```

Exit code: `0`.

Stdout (10 bytes):

```text
TypeError
```

Stderr: empty (0 bytes).

### s1_class_instance.ts

```sh
node --disable-warning=ExperimentalWarning oracle/node.mjs internal/oracle/testdata/host_member_value/s1_class_instance.ts
```

Exit code: `0`.

Stdout (7 bytes):

```text
file a
```

Stderr: empty (0 bytes).

### s2_unknown_reach.ts

```sh
node --disable-warning=ExperimentalWarning oracle/node.mjs internal/oracle/testdata/host_member_value/s2_unknown_reach.ts
```

Exit code: `0`.

Stdout (7 bytes):

```text
host a
```

Stderr: empty (0 bytes).

### v1_tsc_shape.ts

```sh
node --disable-warning=ExperimentalWarning oracle/node.mjs internal/oracle/testdata/host_member_value/v1_tsc_shape.ts
```

Exit code: `0`.

Stdout (32 bytes):

```text
patched host read a
host read b
```

Stderr: empty (0 bytes).

### v2_same_layout_this.ts

```sh
node --disable-warning=ExperimentalWarning oracle/node.mjs internal/oracle/testdata/host_member_value/v2_same_layout_this.ts
```

Exit code: `0`.

Stdout (15 bytes):

```text
other a
host b
```

Stderr: empty (0 bytes).

### v2b_function_expression_this.ts

```sh
node --disable-warning=ExperimentalWarning oracle/node.mjs internal/oracle/testdata/host_member_value/v2b_function_expression_this.ts
```

Exit code: `0`.

Stdout (8 bytes):

```text
other a
```

Stderr: empty (0 bytes).

## TypeScript diagnostics

### Requested commands from the repository root

Working directory: `/workspace/adamic`. Each command returned exit code 1, with the following stdout diagnostic verbatim and empty stderr (0 bytes). TS5112 prevents these invocations from checking the program.

#### h2_plain_call.ts

```sh
tsc --noEmit --strict --target es2022 --module nodenext internal/oracle/testdata/host_member_value/h2_plain_call.ts
```

Exit code: `1`. Stdout:

```text
error TS5112: tsconfig.json is present but will not be loaded if files are specified on commandline. Use '--ignoreConfig' to skip this error.
```

Stderr: empty (0 bytes).

#### h3_reordered_layout.ts

```sh
tsc --noEmit --strict --target es2022 --module nodenext internal/oracle/testdata/host_member_value/h3_reordered_layout.ts
```

Exit code: `1`. Stdout:

```text
error TS5112: tsconfig.json is present but will not be loaded if files are specified on commandline. Use '--ignoreConfig' to skip this error.
```

Stderr: empty (0 bytes).

#### p1_arrow_plain_call.ts

```sh
tsc --noEmit --strict --target es2022 --module nodenext internal/oracle/testdata/host_member_value/p1_arrow_plain_call.ts
```

Exit code: `1`. Stdout:

```text
error TS5112: tsconfig.json is present but will not be loaded if files are specified on commandline. Use '--ignoreConfig' to skip this error.
```

Stderr: empty (0 bytes).

#### r1_holder_run.ts

```sh
tsc --noEmit --strict --target es2022 --module nodenext internal/oracle/testdata/host_member_value/r1_holder_run.ts
```

Exit code: `1`. Stdout:

```text
error TS5112: tsconfig.json is present but will not be loaded if files are specified on commandline. Use '--ignoreConfig' to skip this error.
```

Stderr: empty (0 bytes).

#### r2_alias_field.ts

```sh
tsc --noEmit --strict --target es2022 --module nodenext internal/oracle/testdata/host_member_value/r2_alias_field.ts
```

Exit code: `1`. Stdout:

```text
error TS5112: tsconfig.json is present but will not be loaded if files are specified on commandline. Use '--ignoreConfig' to skip this error.
```

Stderr: empty (0 bytes).

#### r3_apply.ts

```sh
tsc --noEmit --strict --target es2022 --module nodenext internal/oracle/testdata/host_member_value/r3_apply.ts
```

Exit code: `1`. Stdout:

```text
error TS5112: tsconfig.json is present but will not be loaded if files are specified on commandline. Use '--ignoreConfig' to skip this error.
```

Stderr: empty (0 bytes).

#### s1_class_instance.ts

```sh
tsc --noEmit --strict --target es2022 --module nodenext internal/oracle/testdata/host_member_value/s1_class_instance.ts
```

Exit code: `1`. Stdout:

```text
error TS5112: tsconfig.json is present but will not be loaded if files are specified on commandline. Use '--ignoreConfig' to skip this error.
```

Stderr: empty (0 bytes).

#### s2_unknown_reach.ts

```sh
tsc --noEmit --strict --target es2022 --module nodenext internal/oracle/testdata/host_member_value/s2_unknown_reach.ts
```

Exit code: `1`. Stdout:

```text
error TS5112: tsconfig.json is present but will not be loaded if files are specified on commandline. Use '--ignoreConfig' to skip this error.
```

Stderr: empty (0 bytes).

#### v1_tsc_shape.ts

```sh
tsc --noEmit --strict --target es2022 --module nodenext internal/oracle/testdata/host_member_value/v1_tsc_shape.ts
```

Exit code: `1`. Stdout:

```text
error TS5112: tsconfig.json is present but will not be loaded if files are specified on commandline. Use '--ignoreConfig' to skip this error.
```

Stderr: empty (0 bytes).

#### v2_same_layout_this.ts

```sh
tsc --noEmit --strict --target es2022 --module nodenext internal/oracle/testdata/host_member_value/v2_same_layout_this.ts
```

Exit code: `1`. Stdout:

```text
error TS5112: tsconfig.json is present but will not be loaded if files are specified on commandline. Use '--ignoreConfig' to skip this error.
```

Stderr: empty (0 bytes).

#### v2b_function_expression_this.ts

```sh
tsc --noEmit --strict --target es2022 --module nodenext internal/oracle/testdata/host_member_value/v2b_function_expression_this.ts
```

Exit code: `1`. Stdout:

```text
error TS5112: tsconfig.json is present but will not be loaded if files are specified on commandline. Use '--ignoreConfig' to skip this error.
```

Stderr: empty (0 bytes).

### Completed full-program checks from outside the repository

Working directory: `/tmp/fj0vqtf-evidence` (no `tsconfig.json`). These invocations retain all requested flags and use the original full source files by absolute path; no source was copied or edited and no compiler flag was added.

#### h2_plain_call.ts

```sh
tsc --noEmit --strict --target es2022 --module nodenext /workspace/adamic/internal/oracle/testdata/host_member_value/h2_plain_call.ts
```

clean. Exit code: `0`. Stdout: empty (0 bytes). Stderr: empty (0 bytes).

#### h3_reordered_layout.ts

```sh
tsc --noEmit --strict --target es2022 --module nodenext /workspace/adamic/internal/oracle/testdata/host_member_value/h3_reordered_layout.ts
```

clean. Exit code: `0`. Stdout: empty (0 bytes). Stderr: empty (0 bytes).

#### p1_arrow_plain_call.ts

```sh
tsc --noEmit --strict --target es2022 --module nodenext /workspace/adamic/internal/oracle/testdata/host_member_value/p1_arrow_plain_call.ts
```

clean. Exit code: `0`. Stdout: empty (0 bytes). Stderr: empty (0 bytes).

#### r1_holder_run.ts

```sh
tsc --noEmit --strict --target es2022 --module nodenext /workspace/adamic/internal/oracle/testdata/host_member_value/r1_holder_run.ts
```

clean. Exit code: `0`. Stdout: empty (0 bytes). Stderr: empty (0 bytes).

#### r2_alias_field.ts

```sh
tsc --noEmit --strict --target es2022 --module nodenext /workspace/adamic/internal/oracle/testdata/host_member_value/r2_alias_field.ts
```

clean. Exit code: `0`. Stdout: empty (0 bytes). Stderr: empty (0 bytes).

#### r3_apply.ts

```sh
tsc --noEmit --strict --target es2022 --module nodenext /workspace/adamic/internal/oracle/testdata/host_member_value/r3_apply.ts
```

clean. Exit code: `0`. Stdout: empty (0 bytes). Stderr: empty (0 bytes).

#### s1_class_instance.ts

```sh
tsc --noEmit --strict --target es2022 --module nodenext /workspace/adamic/internal/oracle/testdata/host_member_value/s1_class_instance.ts
```

clean. Exit code: `0`. Stdout: empty (0 bytes). Stderr: empty (0 bytes).

#### s2_unknown_reach.ts

```sh
tsc --noEmit --strict --target es2022 --module nodenext /workspace/adamic/internal/oracle/testdata/host_member_value/s2_unknown_reach.ts
```

clean. Exit code: `0`. Stdout: empty (0 bytes). Stderr: empty (0 bytes).

#### v1_tsc_shape.ts

```sh
tsc --noEmit --strict --target es2022 --module nodenext /workspace/adamic/internal/oracle/testdata/host_member_value/v1_tsc_shape.ts
```

clean. Exit code: `0`. Stdout: empty (0 bytes). Stderr: empty (0 bytes).

#### v2_same_layout_this.ts

```sh
tsc --noEmit --strict --target es2022 --module nodenext /workspace/adamic/internal/oracle/testdata/host_member_value/v2_same_layout_this.ts
```

clean. Exit code: `0`. Stdout: empty (0 bytes). Stderr: empty (0 bytes).

#### v2b_function_expression_this.ts

```sh
tsc --noEmit --strict --target es2022 --module nodenext /workspace/adamic/internal/oracle/testdata/host_member_value/v2b_function_expression_this.ts
```

clean. Exit code: `0`. Stdout: empty (0 bytes). Stderr: empty (0 bytes).

## Scope of repository changes

Only `proof/lower-fj0w2-1/report.md` is included in the report commit. The branch starts at the checked HEAD. Compiler code, fixtures, and intentional `.stop` expectations were not edited. The sole push destination for this task is `origin` branch `proof/lower-fj0w2-1-report`.
