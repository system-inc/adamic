# S3c-2 M1 proof — #qk8rztp

Baseline passed; M1 was killed by `second_throws.a`; all three requested lower tests passed. The mutant was restored before the lower tests and report commit.

## Revision and setup

- Repository HEAD: `7b20dbed6152b6e69351182b90ab7310e2aa3f27`, fetched from `origin emit/generators-s3c2-1` and checked out detached.
- Loaded `/workspace/adamic-tools/env.sh` first; created `/tmp/adamic-gate` outside the repository; exported `GOFLAGS=-buildvcs=false`.
- `git submodule update --init --recursive`: exit 0. Submodule commits: cohere `5360753e19e4417675a90e92a1823bea66b13918`, nested TypeScript `d92d9bfee114c80be2c375d72edae966176e3a4f`.
- `gofmt -l internal`: exit 0, printed nothing.
- Upstream `https://github.com/microsoft/TypeScript.git` checked out outside the repository at `/tmp/adamic-gate/TypeScript`, exact commit `050880ce59e30b356b686bd3144efe24f875ebc8`; exported `ADAMIC_TYPESCRIPT_SOURCE` to this path.
- Copied stage3/api's package.json and package-lock.json to `/tmp/adamic-gate/api` and ran `npm ci --prefix /tmp/adamic-gate/api` (exit 0). Put its node_modules/.bin on PATH; `tsc --version` printed `Version 6.0.3`.
- For npm, exported `NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt`, `npm_config_cafile=/etc/ssl/certs/ca-certificates.crt`, and `npm_config_strict_ssl=true`.

## Baseline

Command (exit 0):

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestGeneratorsSeveralAgreeWithNode$' -count=1 -v -timeout 40m > /tmp/pin.log 2>&1
```

Exact PASS/FAIL lines from `/tmp/pin.log`:

```text
--- PASS: TestGeneratorsSeveralAgreeWithNode (0.00s)
    --- PASS: TestGeneratorsSeveralAgreeWithNode/held_across_yield.a (289.59s)
    --- PASS: TestGeneratorsSeveralAgreeWithNode/second_throws.a (289.60s)
    --- PASS: TestGeneratorsSeveralAgreeWithNode/elaborate_four.a (289.61s)
PASS
```

No failing baseline subtests; no got/want mismatch to report.

## Mutant M1

Asserted that `for _, function := range next.Functions {` occurred exactly once in `internal/ir/generator.go`, then replaced it with `for _, function := range next.Functions[:1] {`.

`go build ./internal/ir`: exit 0.

Reran the exact baseline command with output redirected to `/tmp/m1.log`: exit 1.

Exact PASS/FAIL lines from `/tmp/m1.log`:

```text
--- FAIL: TestGeneratorsSeveralAgreeWithNode (0.00s)
    --- FAIL: TestGeneratorsSeveralAgreeWithNode/second_throws.a (37.09s)
    --- PASS: TestGeneratorsSeveralAgreeWithNode/held_across_yield.a (37.10s)
    --- PASS: TestGeneratorsSeveralAgreeWithNode/elaborate_four.a (37.10s)
FAIL
FAIL	github.com/system-inc/adamic/internal/oracle	37.126s
FAIL
```

The only red subtest was `TestGeneratorsSeveralAgreeWithNode/second_throws.a`. Both native and sanitized stdout differed; no JavaScript mismatch was reported. Exact diagnostics from `/tmp/m1.log`:

```text
    generators_several_test.go:36: native: stdout differs, stdout "0, 1, finished\n3, finishedcaught loud stopped after 3\nfinished\n2; 3; finished again\n4; caught again loud stopped after 4\n4; 6; finished again\n", stderr ""
    generators_several_test.go:36: sanitized: stdout differs, stdout "0, 1, finished\n3, finishedcaught loud stopped after 3\nfinished\n2; 3; finished again\n4; caught again loud stopped after 4\n4; 6; finished again\n", stderr ""
```

Whole got stdout (identical for native and sanitized; stderr empty):

```text
0, 1, finished
3, finishedcaught loud stopped after 3
finished
2; 3; finished again
4; caught again loud stopped after 4
4; 6; finished again
```

Whole want stdout, captured directly with the oracle's Node runner (exit 0; stderr empty):

```text
0, 1, finished
3, caught loud stopped after 3
finished
2; 3; finished again
4; caught again loud stopped after 4
4; 6; finished again

```

The sole differing native output line is `3, finishedcaught loud stopped after 3`; Node's line is `3, caught loud stopped after 3`. Thus M1 incorrectly executes the `finished` append before catching the pending throw. The observed output retains `caught` after `finished`, more precisely than the anticipated shorthand “finished where Node prints caught.”

Node capture command:

```sh
node --disable-warning=ExperimentalWarning oracle/node.mjs /workspace/adamic/internal/oracle/testdata/generators_several/second_throws.a
```

Restored `internal/ir/generator.go` with `git restore -- internal/ir/generator.go`; `git diff --exit-code`: exit 0, no output. The working tree was clean before report creation.

## Lower checks after restoration

Command (exit 0):

```sh
go test ./internal/lower -run '^(TestGeneratorIntoAnIterableParameterLowers|TestGeneratorIntoAnIterableParameterStopsOnEveryOtherRoad|TestNestedIterableCalleeLowers)$' -count=1 -timeout 30m
```

Output:

```text
ok  	github.com/system-inc/adamic/internal/lower	1.337s

```

Failures: none.

## Publication scope

This report is the only file committed on new branch `proof/s3c2-m1-report`, based directly on the tested HEAD. Only that branch is to be pushed to origin; no compiler change or dependency installation is included.
