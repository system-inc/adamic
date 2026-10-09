The subsequent gate fixture correction is recorded in [A-CHECK.md](A-CHECK.md).

Built a transitive, deduplicated union of referenced projects' ambient `types` selections.
Base is this delivery branch at c00c629b51e6099d262e680c3186cf944ebee8cf; the new delivery SHA is in the final response.
Focused loader tests, both Node build profiles, native and JavaScript output, counts refresh, and scoped vet pass.
Four union mutants and all seven existing source-root mutants are caught by assertions.
The actual tsc entry reaches ordinary TS2345 at compiler/builder.ts:1246:69; a native tsc binary remains outside this unit.

The entry's and every transitive reference's `types` names are collected once in
first-seen reference order. The production checker and every option-audit
program use that same union. Package resolution and declaration compatibility
remain the frontend's responsibility. A nonempty union enables declaration
checking in both production and audit programs, even if the separate projects
set `skipLibCheck`. This prevents the merge from silently hiding declaration
conflicts. Differences in the projects' original `skipLibCheck` options are still
refused by the unchanged compatibility comparison. `types` is the sole new exception to the
existing option comparison: strict flags, target, module, `lib`, `paths`, and
other checking/emit conflicts still name both projects and the conflicting
option. Output-directory and composite differences remain compatible.

The clean fixture selects `entry` in one project and `dependency` in the other.
Each source uses its own package's ambient number alias. The dependency exports
`number`, keeping its local ambient alias out of its declaration API. Stock
TypeScript initially exposed that fixture mistake, so the fixture was corrected
before the final oracle run. Both stock `tsc --build` and a single stock program
with the explicit union accept the final fixture and emit programs printing
`7\n`. Native under ASan/UBSan and Adamic JavaScript print the same output,
exit 0, and have empty stderr. A Go fixture also checks a three-project types
union and repeated package-name deduplication.

The duplicate fixture adds `declare const sharedAmbient: number` to both
selected packages. With declaration checking enabled (`skipLibCheck: false`),
the separate stock solution build accepts the independently scoped projects.
The stock shared program and Adamic both reject the union with exactly:

```
@types/dependency/index.d.ts:2:15: TS2451: Cannot redeclare block-scoped variable 'sharedAmbient'.
@types/entry/index.d.ts:2:15: TS2451: Cannot redeclare block-scoped variable 'sharedAmbient'.
```

The same duplicates are refused when both projects set `skipLibCheck: true`;
the shared program still validates the declarations. The option audit classifies
these as ordinary errors rather than stricter-option sites. No union diagnostic is filtered or replaced with a loader refusal.

Mutants run independently through Go overlays:

| Mutant | Assertion that catches it |
|---|---|
| Return only the entry's types | Clean union fixture fails with missing `DependencyNumber`, TS2304 |
| Use only entry types in the audit | Clean audit reports missing `DependencyNumber`, TS2304 |
| Skip union declaration checking in production | Duplicate fixture with project declaration skipping is silently admitted |
| Skip union declaration checking in the audit | Duplicate fixture no longer has ordinary audit errors |
| Drop transitive references | Configured, unimported leaf disappears |
| Restore declaration-output references | Two-project load reports TS6305 |
| Bypass option comparison | Seventeen pinned conflicts fail, including `lib` and `paths` |
| Drop referenced option-audit findings | Unchecked indexed read is admitted |
| Duplicate checking roots | Diamond root uniqueness fails |
| Keep a duplicate physical prelude | Embedded-console identity assertion fails |
| Permit reference cycles | Expected circular refusal disappears |

All eleven final mutants exit 1 through assertions, with no build failure counted
as a kill. An initial audit mutant had an unused variable and was corrected;
only the subsequent assertion failure is evidence. Both verifier runs exit 0.
Raw logs and result JSON are in [evidence/types](evidence/types); scratch
prefixes and stock verbose-build trailing padding are normalized.

The tsc-entry rerun uses the existing fresh `stage3/apply.sh` tree at pinned
TypeScript 050880ce59e30b356b686bd3144efe24f875ebc8. Its package lock selects
`@types/node` 25.3.3. Dependencies were installed with:

```
npm install --prefix /tmp/project-references-tsc-adapted --ignore-scripts --no-save --package-lock=false @types/node@25.3.3 > /tmp/project-types-entry-dependencies.log 2>&1
```

The union now supplies `node` despite the entry's inherited empty types list.
Both `ADAMIC_NATIVE_SPLIT=0` and `1` exit 1 with byte-identical diagnostics. The
first reported diagnostic, in the loader's string-sorted order, is:

```
<adapted>/src/compiler/builder.ts:1246:69: error TS2345: Argument of type 'Path | undefined' is not assignable to parameter of type 'string'.
  Type 'undefined' is not assignable to type 'string'.
```

This comes from ordinary checking under the existing Adamic prelude, which
changes collection iterator result declarations. Replaying the internal audit
with that same embedded prelude records 90 ordinary errors and 171 additional
stricter-option sites; this first diagnostic belongs to the ordinary errors.
The audit without that prelude records zero ordinary errors and the same
171 stricter sites. These are observations, not authorization to suppress either
set of diagnostics. No TS6305 or `types` conflict remains in either entry run.

Commands actually run, with test output sent to logs:

```
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/project-types-setup.log 2>&1
source /workspace/adamic-tools/env.sh
go test ./internal/load -run 'TestProjectLoader|TestProjectReferences' -count=1 -timeout 10m > /tmp/project-types-focus.log 2>&1
go test ./internal/load -run 'TestProjectLoader|TestProjectReferences|TestAdamic|TestOverlay|TestATypeError|TestEveryFile|TestThePrelude|TestHouseStyle|TestLoadRefuses|TestTypeScriptIsThePinnedCommit' -count=1 -timeout 10m > /tmp/project-types-regression.log 2>&1
PROJECT_LOADER_TSC=/tmp/project-references-stock/package/lib/tsc.js python3 stage3/project-references-source/verify_types.py > /tmp/project-types-verify.log 2>&1
PROJECT_LOADER_TSC=/tmp/project-references-stock/package/lib/tsc.js python3 stage3/project-references-source/verify.py > /tmp/project-types-reference-verify.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode/stage3/project-references-source/' -count=1 -timeout 30m > /tmp/project-types-oracle.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/project-types-counts.log 2>&1
go vet ./internal/load ./internal/oracle > /tmp/project-types-vet.log 2>&1
go build -o /tmp/project-types-adamic ./cmd/adamic > /tmp/project-types-build.log 2>&1
ADAMIC_NATIVE_SPLIT=0 /tmp/project-types-adamic build /tmp/project-references-tsc-adapted/src/tsc/tsc.ts -o /tmp/project-types-tsc-0 > /tmp/project-types-entry-0.stdout 2> /tmp/project-types-entry-0.stderr
ADAMIC_NATIVE_SPLIT=1 /tmp/project-types-adamic build /tmp/project-references-tsc-adapted/src/tsc/tsc.ts -o /tmp/project-types-tsc-1 > /tmp/project-types-entry-1.stdout 2> /tmp/project-types-entry-1.stderr
```

Tests and verifiers exit 0. Loader focus 13.668s, final loader regression 10.664s,
filtered backend oracle 0.505s, counts 27.636s. Counts refresh changes no rows:
new ambient-package fixtures synthesize tsconfig-owned scratch sources through
the dedicated verifier rather than entering the standalone `.a` oracle roster.
The existing executable source-root fixtures retain their recorded counts.
Scoped vet, gofmt, and `git diff --check` are clean. Diagnostic attribution was
measured by a temporary in-package witness calling `auditProjectOptions` with
and without `preludePath`, run as `go test ./internal/load -run
'^TestProjectTypesTraceScratch$' -count=1 -timeout 10m >
/tmp/project-types-entry-production-audit.log 2>&1`. The witness was removed
after the run; its report JSON and test log are retained. One concurrent mutant
verification hit a compile-time file-read race during that removal. It was
rerun after removal; that build failure is not counted as a mutant kill.

Setup: Go ready 0.029s, Node 0.029s, submodules 0.079s, markdown 0.084s,
clang 0.212s, build 84.380s, cache warm 84.557s, done 84.593s. `nproc` is 5,
with a four-CPU quota. Setup deferred test binaries as its current default.

Not covered: native tsc execution, separate checker ownership, symlink aliases,
locally divergent installations of the same ambient package name, or every
frontend option spelling. No whole package suite or full gate was run. No
protected lowering/native files or central oracle dispatch were changed.
