# lower-9dmp proof report

Overall result: **FAIL**. Formatting and vet pass; all ten writable-field subtests pass, but both host subtests fail during loading with TS2322, before lowering can return `NotYet` or `Refused`.

## Setup

- Date: 2026-11-11 (UTC).
- Fetched `origin compiler/method-field-host-pin` and checked out detached HEAD `3ef254d57d40f37a7a2f727bcfdc61859a8b1a02`.
- Ran `git submodule update --init --recursive`: `cohere` at `5360753e19e4417675a90e92a1823bea66b13918`; `cohere/TypeScript` at `d92d9bfee114c80be2c375d72edae966176e3a4f`.
- Loaded `/workspace/adamic-tools/env.sh` before setup and Go commands. Created its missing `TMPDIR`, `/tmp/adamic-gate`, outside the repository.
- Provisioned upstream `https://github.com/microsoft/TypeScript.git` outside the repository at `/workspace/typescript-proof-050880ce`, checked out detached commit `050880ce59e30b356b686bd3144efe24f875ebc8`, and exported `ADAMIC_TYPESCRIPT_SOURCE` to that path.
- Toolchain: `go version go1.27.1 linux/amd64`.

## Checks

- `gofmt -l internal/lower`: PASS, exit 0, no output.
- `go vet ./internal/lower`: PASS, exit 0, no output. An initial attempt before recursive submodule setup finished could not load `cohere/TypeScript/tsc/go.mod`; reruns after setup passed.
- Focused test command below: FAIL, exit 1 (`1.080s` on the final run).

```sh
go test ./internal/lower -run 'TestHostMethodSeenAsWritableFieldStops|TestMethodSeenAsWritableFieldIsRefused' -count=1 -v
```

| Subtest | Result |
| --- | --- |
| `TestMethodSeenAsWritableFieldIsRefused/.a/an_object_literal's_method` | PASS |
| `TestMethodSeenAsWritableFieldIsRefused/.a/a_readonly_field` | PASS |
| `TestMethodSeenAsWritableFieldIsRefused/.ts/an_object_literal's_method` | PASS |
| `TestMethodSeenAsWritableFieldIsRefused/.ts/a_method_seen_as_a_method` | PASS |
| `TestMethodSeenAsWritableFieldIsRefused/.ts/the_same_type` | PASS |
| `TestMethodSeenAsWritableFieldIsRefused/.ts/a_readonly_field` | PASS |
| `TestMethodSeenAsWritableFieldIsRefused/.a/a_method_seen_as_a_method` | PASS |
| `TestMethodSeenAsWritableFieldIsRefused/.ts/a_class's_method` | PASS |
| `TestMethodSeenAsWritableFieldIsRefused/.a/the_same_type` | PASS |
| `TestMethodSeenAsWritableFieldIsRefused/.a/a_class's_method` | PASS |
| `TestHostMethodSeenAsWritableFieldStops/.ts` | FAIL |
| `TestHostMethodSeenAsWritableFieldStops/.a` | FAIL |

## Exact host errors from the final run

`lowerSourceAs` calls `t.Fatalf("Load: %v", err)` on a load error. Both subtests abort there; no lowering error reaches the host test's `errors.As` checks. A temporary `t.Logf` after `lowerSourceAs` was therefore not reached. It was removed, and the command above was rerun on unchanged source. The exact diagnostics, including their temporary paths, follow.

### TestHostMethodSeenAsWritableFieldStops/.ts

```text
    method_field_widening_test.go:56: Load: /tmp/adamic-gate/TestHostMethodSeenAsWritableFieldStops.ts1927136146/001/main.ts:3:7: error TS2322: Type 'Hash' is not assignable to type '{ digest: () => string; }'.
          The types returned by 'digest()' are incompatible between these types.
            Type 'NonSharedBuffer' is not assignable to type 'string'.
```

### TestHostMethodSeenAsWritableFieldStops/.a

```text
    method_field_widening_test.go:56: Load: /tmp/adamic-gate/TestHostMethodSeenAsWritableFieldStops.a3688217920/001/main.a:3:7: error TS2322: Type 'Hash' is not assignable to type '{ digest: () => string; }'.
          The types returned by 'digest()' are incompatible between these types.
            Type 'NonSharedBuffer' is not assignable to type 'string'.
```

## Cleanup and delivery

Temporary logging was removed. `git diff --exit-code` passed with no output before and after the final test run, before creating this report. No compiler or test changes remain.

The report is the only file to be committed on `proof/lower-9dmp-report`, a new branch from the specified HEAD. Only that branch is to be pushed to origin.
