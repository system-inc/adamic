Fixed the loader topic's four standalone fixture a-check failures.
Correction builds on 932d3bf88bebe9767bff24eb53af36edf092c488; the delivery SHA is in the final response.
All thirteen authored files check; Node and both backends still print 7, with duplicate diagnostics matching TS2451.
All eleven loader and types-union mutants remain caught by assertions.
These files are absent on 784b577a; no area-next-fixtures merge or full gate was needed.

The first failure in origin's
`gate-logs/932d3bf88beb/20261008T170845Z/fast` is:

```
stage3/project-loader/fixture/app/main.a:2:13: error TS2345: Argument of type 'number' is not assignable to parameter of type 'string'.
```

The same a-check stage reports three more failures: the numeric console call
in the chain's main, then unknown `EntryNumber` and `DependencyNumber` in the
types fixtures. All four files are added relative to 784b577a. The first arrived
with the loader prerequisite 16c22268, the chain was added by this source-root
unit, and the ambient-alias fixtures by its types-union change. They are topic
fixture mistakes, rather than area-next's inherited fixture reds. The same check
cannot run those paths on 784b577a because the files do not exist there.
`origin/compiler/area-next-fixtures` was fetched but not merged.

The gate's pinned tools commit is b082770021ed4092178463b4df7c9e0ba8938c4d.
Its `cloud/fast-gate/run.py:aCheck` invokes `adamic c <file>` independently for
each authored `.a`. A successful exit or a named not-yet lowering stop passes;
a type error fails unless explicitly expected. Replaying that exact outcome rule
on all thirteen files from the failed run now yields thirteen successful exits,
not type-error headers or exemptions. The verifier now repeats standalone
`adamic c` checks before staging its project oracle.

Both numeric console calls now interpolate their values into strings, matching
the existing Adamic prelude's signature and preserving stdout `7\n`. The authored
types fixtures import number aliases as types from the package fixture modules,
so their `.a` forms are independently true. When staging the tsconfig-owned
oracle, the verifier converts the module aliases to ambient package declarations
and removes those type-only imports. Therefore the scratch project still relies
exclusively on each project's `types` selection: selecting only the entry side
still produces TS2304. The duplicate-declaration and declaration-skipping
mutants still produce their required assertion failures. No loader behavior or
option compatibility policy changed in this correction.

Commands run, with outputs sent to logs:

```
go build -o /tmp/project-types-red-adamic ./cmd/adamic > /tmp/project-types-red-build.log 2>&1
PROJECT_LOADER_TSC=/tmp/project-references-stock/package/lib/tsc.js python3 stage3/project-references-source/verify_types.py > /tmp/project-types-red-types.log 2>&1
PROJECT_LOADER_TSC=/tmp/project-references-stock/package/lib/tsc.js python3 stage3/project-references-source/verify.py > /tmp/project-types-red-references.log 2>&1
go test ./internal/load -run 'TestProjectLoader|TestProjectReferences' -count=1 -timeout 10m > /tmp/project-types-red-loader.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/project-types-red-counts.log 2>&1
git diff 784b577a 932d3bf8 --name-status -- stage3/project-loader/fixture/app/main.a stage3/project-references-source/fixture/chain/app/main.a stage3/project-references-source/fixture/types/app/main.a stage3/project-references-source/fixture/types/dependency/value.a
```

The first five commands exit 0. The last records `A` for all four paths. Counts
refresh changes no rows. [Evidence](evidence/a-check) retains the failed gate's
failure summary, base-path comparison, standalone check results, verifier
results, test logs, Node/backend output and all eleven final mutant logs. Scratch
prefixes are normalized.

No whole package test suite or full gate was run. The loader is unchanged since
the previous tsc-entry observation, TS2345 at compiler/builder.ts:1246:69.
