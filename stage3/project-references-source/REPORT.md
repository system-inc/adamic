This is the source-roots delivery record. The subsequent `types` ruling and current tsc-entry stop are recorded in [TYPES.md](TYPES.md).

Built transitive project-reference source roots, compatibility refusals, and matching option audits.
Base: codex/stage3-project-loader 16c2226828eec5f56e9a8bc8aec389cf4b881546; delivery SHA is in the final report.
Focused loader tests, Node solution builds, both executable backends, sanitizer oracle, counts, and scoped vet pass.
Seven real production mutants are caught, including lost transitive traversal and declaration-only references.
The real tsc entry now refuses the entry/compiler types conflict; host DOM console lowering remains outside this unit.

`internal/load/project_references.go:projectSourceRoots` parses the reference graph
and adds all configured sources, including unimported ones, exactly once by the
filesystem's case-sensitive or case-insensitive path key. Each edge compares
options before the existing loader strengthens null and function checking.
Cycles are explicitly refused. `loadInput` and the option audit both build one
checker without declaration-output redirects. Explicit execution entries stay
separate; checking a configured source does not execute it.

The conservative assumption is one checking and emission contract across the
entry and every reference. Effective strict defaults, target, module, resolution,
class-field semantics, and the four audited soundness flags must agree. Other
option differences fail closed by default, including `lib`, `types`, and resolver
paths. Where no semantic default getter is available, all three option states
are retained: an unspecified flag cannot be assumed false. Output directories, rootDir, composite/incremental bookkeeping, maps, and
reporting options can differ. The source program uses an explicit common rootDir
and drops composite's separate file-list boundary while retaining the entry's
declaration validation. This is a native whole-program policy, shared by the
JavaScript backend; it is not a TypeScript solution/declaration build.

The pinned refusal has this shape, with actual absolute config names:

```
load: project reference options conflict: <app>/tsconfig.json references <dependency>/tsconfig.json: compiler option strictNullChecks differs; separate checker ownership is required
```

Observations

Stock TypeScript 6.0.3 on Node 24.19.0 accepts both the original two-project
fixture and the three-project chain with `tsc --build`, then prints `7`.
Fresh Adamic loads accept each before any declarations exist and retain one
execution entry. An unimported configured source with `number = "bad"` is
rejected by both checkers at `unused.ts:1:14`, TS2322, with the same message.
An existing declaration output cannot replace the implementation source.
The diamond test inspects the actual command-line roots, including its leaf.
Sixteen pinned refusal cases cover strict flags, target, module, and types;
compatible composite, output-directory, and effective-default differences pass.
Node solution builds also accept and run the differing strictNullChecks profiles;
Adamic deliberately refuses those graphs because it has one checker ownership
contract, with both project names and strictNullChecks pinned in the message.

The original DOM `console.log` reaches an existing inherited-library-member
refusal in both backends. The executable profiles use ``console.log(`${value}`)``,
`lib: ["es2020"]`, and the existing Adamic prelude as the console contract on
both checkers. The loader reuses its embedded prelude instead of declaring it
again from an identical physical copy. Stock emitted JavaScript, native under
ASan/UBSan, and Adamic JavaScript all print exactly `7\n`, exit 0, and have empty
stderr. Declaration output directories are deleted before native compilation.
The ordinary oracle also runs the authored `.a` executable profiles against
source Node, release native, sanitized native, backend Node, and leak checks.

`counts.md` adds only the two executable profiles, both 2 allocations, 2 frees,
0 retains, 2 releases, peak 2, and 0 regions. No existing row changes.

Mutants, each independently applied through a Go overlay

| Mutant | Catcher |
|---|---|
| Skip references of dependencies | `TestProjectReferencesTransitive`: unimported leaf source is absent |
| Restore declaration-output references | `TestProjectLoaderReferenceSources`: TS6305 returns |
| Bypass option compatibility | All sixteen `TestProjectReferencesConflictingOptions` cases fail their pinned refusals |
| Drop referenced option audit findings | `TestProjectReferencesAudit`: unchecked indexed read is admitted |
| Drop final root deduplication | `TestProjectReferencesDiamond`: duplicate checking root |
| Retain a physical prelude copy | `TestProjectReferencesSharedPrelude`: console no longer belongs to the embedded declaration lowering recognizes |
| Permit a reference cycle | `TestProjectReferencesCycle`: expected circular refusal is missing |

The physical-prelude mutant initially survived a load-only check: host-console
discovery suppressed the embedded console. The strengthened test checks the
declaration identity required by lowering and catches that mutant.

All seven runs exit 1 through test assertions, with no build failure substituted
for a test failure. The unmutated differential verifier exits 0. Raw mutant
logs and output observations are in `evidence/`; its verification-directory
prefix alone is replaced with `<verification>`.

Real tsc-entry next stop

`STAGE3_CACHE=/tmp/project-references-stage3-cache bash stage3/apply.sh
/tmp/project-references-tsc-adapted` creates a fresh tree at the pinned TypeScript
commit 050880ce59e30b356b686bd3144efe24f875ebc8 with this base's adaptations.
Building `src/tsc/tsc.ts` with `ADAMIC_NATIVE_SPLIT=0` and `1` now exits 1 with
byte-identical stdout and stderr:

```
adamic: load: project reference options conflict: <adapted>/src/tsc/tsconfig.json references <adapted>/src/compiler/tsconfig.json: compiler option types differs; separate checker ownership is required
```

The entry inherits `types: []`; the compiler writes `types: ["node"]`. No union
of ambient types was inferred or inserted. This is the next observed stop, not
a claim that tsc lowers or runs natively.

Commands actually run, with every test output redirected to a log

```
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/project-references-setup-step32.log 2>&1
source /workspace/adamic-tools/env.sh
go test ./internal/load -run 'TestProjectLoader|TestProjectReferences' -count=1 -timeout 10m > /tmp/project-references-focus.log 2>&1
go test ./internal/load -run 'TestProjectLoader|TestProjectReferences|TestAdamic|TestOverlay|TestATypeError|TestEveryFile|TestThePrelude|TestHouseStyle|TestLoadRefuses|TestTypeScriptIsThePinnedCommit' -count=1 -timeout 10m > /tmp/project-references-regression.log 2>&1
PROJECT_LOADER_TSC=/tmp/project-references-stock/package/lib/tsc.js python3 stage3/project-references-source/verify.py > /tmp/project-references-verify.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode/stage3/project-references-source/' -count=1 -timeout 30m > /tmp/project-references-oracle.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/project-references-counts.log 2>&1
go vet ./internal/load ./internal/oracle > /tmp/project-references-vet.log 2>&1
git diff --check
```

All exit 0. Loader focus 1.940s, loader regression 2.127s, filtered oracle
7.853s, counts 37.867s. Scoped vet and diff checks have no output. The verifier
also builds the probe and CLI, invokes stock tsc, and runs all seven overlays;
`verify.py` records their exact subprocess arguments. Gofmt on all touched Go
files reports no outstanding changes.

Setup succeeded: Go ready 0.034s, Node 0.027s, submodules 0.073s, markdown
0.084s, clang 0.218s, build 50.678s, cache warm 50.792s, done 50.827s.
`nproc` is 5, cgroup quota four CPUs. The printed environment file is
`/workspace/adamic-tools/env.sh`.

Not covered: a native tsc binary, per-project checker/linker ownership,
symlink-alias graphs, every equivalent spelling of every frontend option,
large reference graphs, and host-library console lowering. Unrecognized
option differences are conservatively refused. No whole Go package suite or
full repository gate was run. No files were copied from cohere or changed in
the protected lowering/native emitters or the central oracle test file.
