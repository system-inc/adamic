# Literal builtin require

Partial implementation on `codex/require-builtins`, started from main
`ef3d907ecdc4c771b016f7d9c52372def057a340`.

Current main had no Node builtin declarations or lowering. Inspected the pushed
`codex/host-fs-file`, `codex/host-fs-directory`, `codex/host-process`, and
`codex/host-buffer-crypto` branches. Fast-forwarded the existing fs file host
commit `080789ff46f07af44fafc8fb64c2b0ed1e24e129` as the dependency.

The directory tip `f5702191f4aff7028464df3ab2b334389b7ca496` includes the path
runtime, but its corrected report explicitly says the shared @types/node
25.3.3 loader hook is pending. Its lowering recognizes declarations from
node/fs.d.ts and node/path.d.ts, which the current fs-file tip does not provide.
Refreshing both branches confirmed these tips. The directory branch was not
merged because its corrected fixtures are intentionally blocked on that shared
hook. No second path runtime or declaration loader was implemented here.

The fs host serves selected @types/node 25.3.3 declarations in the embedded
prelude and recognizes calls by their declaration symbols. Literal `fs` and
`node:fs` require bindings use `typeof import('node:fs')`, so the checker gives
exactly the namespace import type and the same member symbols. Plain const
namespace bindings need no runtime local, just like namespace imports. Calls
reuse the existing fs host intrinsics. Source is never rewritten, and a local
function called require or a local module object remains ordinary user code.

The NodeRequire any result is refined to a module type for supported literals.
The fallback is unknown and is never admitted to lowering. The full @types/node
package is not loaded; the existing host's selected declarations remain the
source of its declared module type.

Non-literal calls (including a const whose inferred type is the literal fs),
non-builtin packages, relative paths, invented node: names, require.resolve,
require.cache, require read as a value, and module.exports are refused with an
import fix. Indexed module access is refused too. A checker error can precede
this refusal when a rejected require result is used with an invalid type; this
unit does not reorder loader diagnostics.

Node 24.19.0 observes `true true` and `true` from bare, prefixed and function-local
fs.existsSync calls. Native and JavaScript match stdout, stderr and exit status;
ASan, UBSan and the leak check pass. The fs require fixture adds only its count
row: 2 allocations, 2 frees, 0 retains, 2 releases, peak 1, no regions. Existing
rows did not move. An exact IR comparison also passes against namespace imports.

Pending the directory host's shared declaration hook, both path fixtures are
run from source on Node: join and dirname print `a/b`
and `a`. Lowering reports the missing node:path host as NotYet. These are gap
fixtures, not claims that path is implemented.

Mutants actually run and restored:

- Accept a non-literal argument as node:fs. TestCommonJSRefusals fails, including
  a const namespace binding which the mutant really accepts (`got <nil>`).
  The test exits 1; the compiler builds, so no compiler error or clang warning
  catches this mutant.
- Replace the fs module result with a structurally similar existsSync function
  object. TestRequireBindingHasTheImportedModuleType fails for both bindings:
  the checker type differs and the member declaration identity is lost. Exit 1.

Setup: Go ready at 0s; clang ready at 1s; Node ready at 1s; submodules ready at
1s; build cache warm at 95s; done at 95s. nproc is 5, with a four-core cgroup
quota. Tools were sourced from /workspace/adamic-tools/env.sh. Setup succeeded.

Workspace logs are /tmp/require-builtins-{setup,packages,focused,oracle,counts,
mutant,type-mutant,final-tests,format,vet,gate}.log. Logs are not committed,
as .gitignore requires. Commands actually run:

- bash cloud/setup.sh: exit 0, timing lines above.
- source /workspace/adamic-tools/env.sh: used in every test shell.
- go test -count=1 ./internal/load ./internal/lower: both packages print ok.
- go test -count=1 -v ./internal/oracle -run TestRequire: prints PASS; Node
  source observations and the named path gaps are logged.
- go test -count=1 -timeout 30m ./internal/oracle -run TestCountsAreRecorded
  -args -update-counts: prints ok, 19.427s, only the new row changes.
- go test -count=1 -v ./internal/lower -run TestCommonJSRefusals, with the
  non-literal mutant: exit 1, want CommonJS refusal but got <nil>.
- go test -count=1 -v ./internal/load -run
  TestRequireBindingHasTheImportedModuleType, with the structural type mutant:
  exit 1, module types differ and member declaration identity is lost.
- gofmt -l cmd internal: no output.
- go vet ./...: no output, exit 0.

Final checks after the prefix-only builtin correction and gap-fixture relocation:

- go test -count=1 -timeout 30m ./internal/load ./internal/lower: exit 0;
  load 1.752s, lower 33.898s.
- go test -count=1 -timeout 30m ./internal/flow: exit 0, 125.520s. The path gap
  fixtures live under testdata/require_gaps so the compilable-program glob does
  not treat them as supported programs.
- go test -count=1 -v -timeout 30m ./internal/oracle -run
  'TestRequire|TestNodeFSFileAgreesWithNode|TestCountsAreRecorded': exit 0,
  52.678s. Both path source observations, fs and its dependency's twelve host
  fixtures, sanitizers, leaks and counts pass.
- ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./...: exit 1. The initial
  run caught the misplaced path gaps; that is fixed and the full flow package
  subsequently passed. It also caught an inherited fs-host integration gap:
  internal/fresh reports ir.NodeFSFile as unknown, including the dependency's
  pre-existing import fixtures. The fs-file dependency adds the IR node but
  does not update internal/fresh. That directory is outside this unit's
  territory, so the check was not weakened or bypassed.

The full gate was interrupted after more than twelve minutes while unrelated
stage-1 port checks and a four-billion-comparison Unicode probe were still
running. It is not a complete green gate. Before interruption, load, lower,
native and the complete oracle package passed; the external bridge passed in
546.097s. The gate log retains the failures and interrupt. Final format, vet
and git diff --check have no output.

Remaining work: path and other host modules; mutable, destructured, returned or
inline namespace require values; runtime namespace identity comparisons; and
integration against tsc's complete sys.ts with full @types/node. These shapes
are rejected or report NotYet instead of silently receiving a fabricated
runtime namespace. All currently supported Node hosts are ambient declarations,
so they have no source module body to schedule; general graph evaluation for a
future builtin implemented in Adamic is not established by these fixtures.

Feature branch update after integration-boundary correction

No push or merge was made into main or an area/ branch. Only
codex/require-builtins was pushed. origin/main at e8ba3d5 was merged into
this feature branch without conflicts in merge commit
1ea8d98747a539cc98a44f2629ffbd457432d4b6.

Post-merge validation (output retained under /tmp):
- go test -count=1 -timeout=30m ./internal/load ./internal/lower
  ./internal/flow ./internal/fresh: exit 1. Load, lower and flow pass;
  fresh still fails TestEveryWriteIsRecordedAndKnown because ir.NodeFSFile
  is unknown, across dependency import fixtures and require_fs.a.
  Log: /tmp/require-builtins-merged-packages.log.
- go test -count=1 -timeout=30m ./internal/oracle -run
  'TestRequire|TestNodeFSFileAgreesWithNode|TestCountsAreRecorded|TestDevirtualize':
  exit 0, 36.446s. Log: /tmp/require-builtins-merged-oracle.log.
- go vet ./...: exit 0; formatting and git diff --check have no output.

The branch is not fully green. The inherited freshness failure and the
previously documented host and typing limitations remain; none was bypassed.
