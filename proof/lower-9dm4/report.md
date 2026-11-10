# Adamic lower-9dm4 proof report

All four requested checks passed. Objects and Assertions emitted no FAILs, so there are no first errors to quote. Assertions passed on its first run; no python3 deadline occurred and no rerun was needed.

## Tested checkout and setup

- Date: 2026-11-10 (UTC).
- Repository: `https://github.com/system-inc/adamic.git`.
- Fetched with `git fetch origin compiler/method-field-widening-4`.
- Tested detached HEAD: `e54546a3d12be937edb9f19f79a16bb76d99ece1`.
- Verified parent: `337cf6917f7231422b5e168bc215888752a9f194`.
- Completed `git submodule update --init --recursive`:
  - `cohere`: `5360753e19e4417675a90e92a1823bea66b13918`.
  - `cohere/TypeScript`: `d92d9bfee114c80be2c375d72edae966176e3a4f`.
- Loaded `/workspace/adamic-tools/env.sh` before setup and each proof command. It sets `TMPDIR=/tmp/adamic-gate`; created that missing directory outside the repository before invoking Go.
- Provisioned upstream `microsoft/TypeScript` at exactly `050880ce59e30b356b686bd3144efe24f875ebc8` in `/tmp/lower-9dm4-proof/TypeScript`; exported `ADAMIC_TYPESCRIPT_SOURCE` to that path.
- Copied `stage3/api/package.json` and `stage3/api/package-lock.json` outside the repository to `/tmp/lower-9dm4-proof/api`, verified both match the tested checkout, and installed with `npm ci --prefix /tmp/lower-9dm4-proof/api --ignore-scripts --no-audit --no-fund`.
- Prepended `/tmp/lower-9dm4-proof/api/node_modules/.bin` to `PATH` and exported `NODE_PATH=/tmp/lower-9dm4-proof/api/node_modules` for the tests. `tsc --version` reported `Version 6.0.3`.
- Toolchain: Go `go1.27.1 linux/amd64`, Node `v24.19.0`, Python `3.12.14`, clang `20.1.8`.
- Ran the proof commands serially from `/workspace/adamic`. Full command logs and exit-status files are outside the repository under `/tmp/lower-9dm4-proof`.
- The working tree was clean before testing and remained clean after all checks.

## Results

### 1. Formatting

```sh
gofmt -l internal/lower
```

Exit: **0**. Output: **empty** (zero bytes).

### 2. Objects fixtures

```sh
go test -p 1 -count=1 -timeout 60m ./stage3/fixtures -run 'TestFixturesObjects$' -v
```

Exit: **0**. No FAILs; no first errors.

```text
PASS
ok  	github.com/system-inc/adamic/stage3/fixtures	150.663s
```

### 3. Assertions fixtures

```sh
go test -p 1 -count=1 -timeout 60m ./stage3/fixtures -run 'TestFixturesAssertions$' -v
```

First run exit: **0**. No FAILs; no first errors. No python3 deadline; no rerun.

```text
PASS
ok  	github.com/system-inc/adamic/stage3/fixtures	6.268s
```

### 4. Method-as-writable-field regression

```sh
go test ./internal/lower -run 'TestMethodSeenAsWritableFieldIsRefused' -count=1 -v
```

Exit: **0**. All ten subtests passed.

```text
PASS
ok  	github.com/system-inc/adamic/internal/lower	0.154s
```

## Report scope

The report branch is `proof/lower-9dm4-report`, created from the tested HEAD. Its only committed file is `proof/lower-9dm4/report.md`. No compiler, fixture, package, or submodule changes are included. Only this report branch is designated for push to `origin`.
