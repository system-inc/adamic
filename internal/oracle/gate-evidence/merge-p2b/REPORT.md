# Linux gate fixes on library/merge-p2b

Baseline `047e857207bee9aa62dbd22204445a0ab59d4319`, cohere
`7945d102a6c18dd36adf9114a758ce646e8b2359`. Linux, Node v24.19.0,
Go 1.27.1, clang 20.1.8, WASI SDK 27. `nproc`: 5; CPU quota: 4.

## Four groups

Host status records use the scratch typed collector documented in
stage3/fixtures/host/RERECORD.md: load.Load, lower.Lower, typed error
classification and repository-relative diagnostic text. Only stage0 JSON spans
are replaced. Masking those spans proves every other byte is identical,
including every Node field. No fixture source changes. Twenty values change:
six Checker, fifteen NotYet, three Refused, one Compiles. 05_writeFile now
refuses adamic/no-unchecked-cast at 37:30. The new compiling fixture,
11_setModifiedTime, passes source Node versus native, sanitizers and leaks in
the shared runner. All stage3 packages pass (fixtures: 301.361s).

process_bad_code is an accepted source program whose invalid exit-code setter
throws a catchable error. It does not belong among runtime-only, non-lowering
fixtures. Lowering already classifies process exit/status validation and
cwd/chdir/measure failures as throwable; flow omitted ProcessCall from CanThrow.
The graph now carries the same exception edges, including catch and finally.
The fixture remains covered. Full flow passes (770.929s).

Node observations on immediate exit accept precisely the bytes written or a
prefix, with exit status and the other stream still strict. Tests record byte
counts and whether delivery was full. Adamic's native and JavaScript backends
must deliver every byte. The source runner imports adamic.mjs, which explicitly
makes its standard streams blocking; requiring this reference to deliver exactly
4096 bytes was incorrect. Independent direct Node probes
observed 204800/204800 bytes in blocking mode and prefixes of 5120/204800 stdout
and 4096/204800 stderr in nonblocking mode, all exit 37.

The configuration port previously tracked cohere 715ba94f. Commit 3e66e9f8
renames cohere:house to cohere:style with no alias. All eight embedded preset
texts now match the pin byte for byte; the name inventory remains sorted.
The pin also adds the Next import-meta-path rule and override, includes
*.test.a in system-inc/base, and supports top-level sourceExtensions in its
TypeScript reader. The port honors the extension list, inheritance, empty-list
opt-out, validation and separate extension groups. As Go ReadProjectConfig
does, option-only diagnostics permit enumeration of valid values; mixed
structural diagnostics still refuse. Direct style and obsolete-house cases,
source-extension inheritance, same-stem .a/.ts, opt-out and invalid-value cases
are compared against Go itself. The full configuration package passes (272.916s), including all five loader mutants and the existing discovery mutants.

## Proofs

- Remove ProcessCall throwable handling through a Go overlay: the existing Node trace fails with the five original mid-block call-end diagnostics. Compilation succeeds; only the graph path check catches it. `/tmp/p2b-flow-mutant.log`.
- Alter the recorded 05_writeFile diagnostic in a scratch status file: the exact Node comparison passes and only the stage0 comparison fails. `/tmp/p2b-status-mutant.log`.
- Drop JavaScript output without changing exit 37: clean execution produces zero bytes; only the full-output comparison catches it. Native buffer-flush and stream-order mutants remain in the whole oracle.
- Restore the obsolete preset origin: native, source Node and JavaScript output must agree with one another, execute cleanly, and disagree only with the independent Go result.
- Disable source-extension opt-in: the Go file-list comparison catches the missing .a sources. Existing strict-JSON, inherited-option, tsconfig-exclusion and discovery mutants remain checked.

## Setup and gates

`GOPROXY='https://proxy.golang.org|direct' bash cloud/setup.sh --wasi-sdk`
completed in 39.226s: Node ready 0.036, Go 0.048, markdown 0.158,
submodules 0.189, clang 0.363, SDK 0.452, Go build 38.976,
cache warm 39.179. Environment: `/workspace/adamic-tools/env.sh`.
`npm ci --prefix stage3/api` installed three packages in 454ms.

Commands write output directly to logs:

```sh
go test ./internal/flow ./stage1/cohere/config -count=1 -v -timeout 30m
go test ./stage3/... -count=1 -v -timeout 30m
ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -count=1 -v -timeout 30m -args -update-counts
go test ./stage1/cohere/config -count=1 -v -timeout 30m
go test ./internal/oracle -count=1 -v -timeout 30m -args -update-counts
go vet ./...
gofmt -l cmd internal stage1/cohere/config
git diff --check
```

The first configuration run's four Go reference builds exceeded their ten-minute
bounds under concurrent cold preparation. Explicit module download and one
`go test ./command/cohere -run '^$'` prewarm succeeded; configuration was then
rerun. Intermediate Go comparisons exposed and caught the additional
source-extension and option-diagnostic drift described above.

All seven WASI test groups pass: host input, file effects, runtime refusals,
ordinary fixtures, oracle mutants, runner mutants and emission. The WASI-enabled
package run was interrupted by the execution environment restart during its
remaining native/count checks. A fresh whole native oracle run passed in 546.997s, including counts verification
(270.08s), and regenerated counts on Linux with no table changes. Flow passed in
770.929s and stage3 fixtures passed in 301.361s. The earlier long-backtrack test
hit the acknowledged box deadline (222.78s); it passed in the final whole run.
No timing thresholds are changed. Final vet, formatting and diff checks passed.

Limits: no full ./... repository run or Unicode package test is claimed. The
requested oracle, flow, config and stage3 packages are the gates for this unit.
Known timing-only failures are reported, not repaired. WASI's existing explicit
host refusals and unsupported sanitizer flags remain unchanged.
