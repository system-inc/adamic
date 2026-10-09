Published the project-aware loader alone, with project references preserved and option findings refused.
Base: 3ffb1a835184713998a34874e86326cd21db971f; provenance: 29933a3e and 2b490213.
Loader and CLI tests, two-project Node fixture, repository vet and filtered native/Node oracle pass.
Five mutants caught: dropped references, dropped composite roots, audit bypass, noCheck and duplicate console.
No compiler feature merges, runtime-check scheduling, project builds or source-root flattening included.

The loader port

`internal/load/project_loader.go:projectOptionsForRoots` selects the nearest tsconfig for `.ts` roots, parses inherited options and retains strict null/variance checking. `.a` and unconfigured `.ts` roots retain main's defaults. Mixed explicit roots and unsupported cross-project source ownership are refused. `noCheck` is explicitly refused because this publication has no backend for unchecked source.

`internal/load/project_options.go:auditProjectOptions` compares the project's ordinary checker diagnostics with four stricter options, then ablates each option to attribute additional diagnostics. Explicit entry roots join this audit. Findings remain errors: this branch cannot admit them without the separate compiler checks. `AuditProjectOptions` exposes the report for the next compiler unit.

`internal/load/load.go:loadInput` carries the owning project's options, declaration roots and `projectConfig.ProjectReferences()` into `tsoptions.NewParsedCommandLine` and `compiler.NewProgram`. Composite projects check their configured sources plus explicit roots. `Files()` still returns only explicitly selected entries. The loader discovers the host's global console before choosing the fallback prelude and includes declaration diagnostics when declaration emission applies.

Provenance is the pre-reference-flattening `load.go.before` from the uncommitted 29933a3e scratch described by 2b490213. `project_loader.go` and `project_options.go` originated on c09c22b8. This is a source port, not a merge of those feature branches. The runtime-check scheduler, optional/indexed lowering, JSON contract hooks, Node autodiscovery, separate Set prelude, project-entry CLI and source-root experiment are excluded. The changed production files are exclusively `internal/load/`.

Observed fixture behavior

Authored sources are `.a`; verification copies them into temporary `.ts` files and changes the import extension to `.js`. Both projects are composite. Before outputs exist, this loader reports:

```
app/main.ts:1:23: error TS6305: Output file 'out/dependency/value.d.ts' has not been built from source file 'dependency/value.ts'.
```

Node 24.19.0 with TypeScript 6.0.3 accepts `tsc --build app --verbose`, builds dependency before app and produces a program printing `7`. After this build, Adamic's loader accepts the entry and reports exactly one execution root. The fixture does not invoke native lowering.

The premise that current main already reports TS6305 was not reproduced. An isolated overlay restoring main's exact `load.go` reports TS2345 at `app/main.ts:2:13`, because main ignores the tsconfig and its fallback console only accepts strings. The new behavior matches the original project-aware scratch, including its TS6305, while the existing standalone/.a loader tests and native oracle pass. Project-owned `.ts` intentionally gains the project's options and libraries. This is not a claim that every project `.ts` diagnostic is unchanged from the pre-port main loader.

Verification

```
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
go test ./internal/load -count=1
go test ./cmd/adamic -count=1 -timeout 30m
PROJECT_LOADER_TSC=/path/to/typescript/lib/tsc.js python3 stage3/project-loader/verify.py
go vet ./...
go test ./internal/oracle -count=1 -timeout 30m -run '^TestNativeAgreesWithNode/dedication/dedication.a$'
```

Setup timing: Node 0.041s, Go 0.050s, submodules 0.094s, markdown 0.109s, clang 0.285s, build 40.957s, cache 41.124s, total 41.208s. `nproc`: 5; CPU quota: four cores. All test output went to log files. The complete repository test gate was not run; the touched package, CLI package, fixture, five mutants, repository vet and named filtered oracle were run. Normalized output is in `evidence/`.

The five Go overlays mutate production code without modifying the working tree. Dropping configured composite sources loses an error in the non-entry file; removing `ProjectReferences()` loses TS6305; discarding audit sites admits the indexed read; bypassing the noCheck refusal admits unchecked source; keeping the fallback console alongside DOM's console produces duplicate declarations. Each corresponding test fails, with no build failure standing in for a test failure.

Compiler's next unit can add source roots for referenced projects here. This branch deliberately preserves the existing checker reference-output decision; it neither builds dependencies nor redirects their sources.
