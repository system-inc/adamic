# Lower proof: 9dmp2

Result: PASS. Formatting and vet passed; both requested tests and all 12 subtests passed.

## Checkout and environment

- Proof date: 2026-11-11 (UTC).
- Repository: `https://github.com/system-inc/adamic.git`.
- Fetched with `git fetch origin compiler/method-field-host-pin-2`, then checked out detached at `9c58d3c8eed7af166e55f7766c40a8e765add8d5`.
- `git submodule update --init --recursive` succeeded. Pins: `cohere` at `5360753e19e4417675a90e92a1823bea66b13918`; `cohere/TypeScript` at `d92d9bfee114c80be2c375d72edae966176e3a4f`.
- Loaded `/workspace/adamic-tools/env.sh` before the checkout and Go commands. Go: `go1.27.1 linux/amd64`.
- Created `/tmp/adamic-gate` outside the repository for the configured `TMPDIR`.
- Provisioned upstream `https://github.com/microsoft/TypeScript.git` outside the repository at `/workspace/adamic-typescript-source`, detached at `050880ce59e30b356b686bd3144efe24f875ebc8`, and exported `ADAMIC_TYPESCRIPT_SOURCE=/workspace/adamic-typescript-source` for the checks.

## Checks

1. `gofmt -l internal/lower`: PASS, exit 0, no output.
2. `go vet ./internal/lower`: PASS, exit 0, no diagnostics.
3. `go test ./internal/lower -run 'TestHostMethodSeenAsWritableFieldStops|TestMethodSeenAsWritableFieldIsRefused' -count=1 -v`: PASS, exit 0.

Every test/subtest result from the verbose run:

```text
--- PASS: TestMethodSeenAsWritableFieldIsRefused (0.00s)
--- PASS: TestMethodSeenAsWritableFieldIsRefused/.a/an_object_literal's_method (0.19s)
--- PASS: TestMethodSeenAsWritableFieldIsRefused/.ts/an_object_literal's_method (0.20s)
--- PASS: TestMethodSeenAsWritableFieldIsRefused/.a/the_same_type (0.27s)
--- PASS: TestMethodSeenAsWritableFieldIsRefused/.a/a_readonly_field (0.38s)
--- PASS: TestMethodSeenAsWritableFieldIsRefused/.a/a_method_seen_as_a_method (0.41s)
--- PASS: TestMethodSeenAsWritableFieldIsRefused/.ts/a_method_seen_as_a_method (0.54s)
--- PASS: TestMethodSeenAsWritableFieldIsRefused/.a/a_class's_method (0.22s)
--- PASS: TestMethodSeenAsWritableFieldIsRefused/.ts/a_class's_method (0.08s)
--- PASS: TestMethodSeenAsWritableFieldIsRefused/.ts/a_readonly_field (0.16s)
--- PASS: TestMethodSeenAsWritableFieldIsRefused/.ts/the_same_type (0.10s)
--- PASS: TestHostMethodSeenAsWritableFieldStops (0.00s)
--- PASS: TestHostMethodSeenAsWritableFieldStops/.ts (1.13s)
--- PASS: TestHostMethodSeenAsWritableFieldStops/.a (1.16s)
PASS
ok  github.com/system-inc/adamic/internal/lower  1.212s
```

## Exact host-method errors

These are the complete error strings printed by a temporary `t.Logf("exact host error: %v", err)` immediately after `lowerSourceAs` in `TestHostMethodSeenAsWritableFieldStops`:

`TestHostMethodSeenAsWritableFieldStops/.a`: PASS

```text
/tmp/adamic-gate/TestHostMethodSeenAsWritableFieldStops.a16936144/001/main.a:3:55: stage 0 can't lower Buffer or Hash viewed as another object type (native host internal slots) yet
```

`TestHostMethodSeenAsWritableFieldStops/.ts`: PASS

```text
/tmp/adamic-gate/TestHostMethodSeenAsWritableFieldStops.ts70021551/001/main.ts:3:55: stage 0 can't lower Buffer or Hash viewed as another object type (native host internal slots) yet
```

The host tests therefore stop under the existing native-host internal-slot rule when a Hash is viewed as another object type. This proof covers only the requested tests and lower-package formatting/vet checks.

## Cleanup and publication

Removed the temporary log. Before writing this report, `git diff --exit-code` and `git diff --cached --exit-code` both passed with no output; `git status --short` was empty, and a repeated `gofmt -l internal/lower` printed nothing. No compiler or test-source changes remain.

The publication branch is `proof/lower-9dmp2-report`, created from the exact checked-out proof HEAD above. The only file committed for this proof is `proof/lower-9dmp2/report.md`; the push targets only `origin` branch `proof/lower-9dmp2-report`.
