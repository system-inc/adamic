Built: ordered multi-root programs, --project tsconfig input, required-option refusals, and host-aware sound prelude coexistence.
Implementation commit: cd8592dc5747ddc6a171a21bc509f716aeb64425, on codex/multi-root from ef3d907ecdc4c771b016f7d9c52372def057a340.
Commands/results: touched packages passed, vet and formatting passed, uncached counts and two existing module oracles passed; setup 97s, nproc 5.
Mutants: all 16 caught by the root oracle, option refusals, glob/alias, JSON, console, and declaration regression tests; detailed below.
Not covered: the complete uncached gate, cycle integration, project-reference builds, native Node console/APIs, native tsc, or symlink-root identity.

# Multi-root and project input report

## What changed

`load.Load` already checked multiple roots. `lower.Lower` now accepts multiple
executable roots, retaining their input order, and runs each module once across
the complete program. `rootOrder` combines the existing single-entry DFS walks
without modifying `modules.go:moduleOrder`. The native binary and JavaScript
backend receive the same ordered main body. Declaration-only roots have no
executable body; a program with none is refused. Direct `.d.ts` inputs still
report declarations through the existing type-query API.

`adamic build --project <tsconfig.json> -o <out>` uses typescript-go's own
`tsoptions.GetParsedCommandLineOfConfigFile` shim for config inheritance,
files/include/exclude, and compiler options. Explicit CLI roots can also be
passed to build, c, and js. Required options are inherited from Adamic when
omitted; an explicit weakening after config inheritance is refused with the
option's name and "Adamic requires". Strict sub-option overrides and checker
bypass options are covered, and additional project strictness remains effective.
The `.a` aliases are exposed to upstream glob matching; only top-level file
specifications are adapted using the upstream JSON syntax tree. No cohere code
was copied. Project references are explicitly not yet supported.

The prelude remains present with its runtime module, Weak brand, sound JSON
signature, Set operations and iterator contracts. A project's actual global
console declarations are discovered in its loaded library/declaration files;
only the restricted Adamic console is omitted when the host supplies one.
The final program uses a fresh cache for its prelude choice, and all its
checker diagnostics are still collected. The regex soundness adapter remains
active. A scratch native project using the census's pinned Node declarations
built and ran under sanitizers with this policy.

Every direct-loader option and its project rule is listed in
[docs/projects.md](../../docs/projects.md), including why removing the census's
whole prelude is a diagnostic experiment, not a sound production policy.

## Entry and integration boundary

I read the census report and both overlays at
`origin/codex/tsc-census`, commit `429c1177f0130f785c19cf590d1860513b2ddbfc`.
The starting main was `ef3d907ecdc4c771b016f7d9c52372def057a340`.

At the pinned upstream TypeScript source commit
`050880ce59e30b356b686bd3144efe24f875ebc8`, `src/tsc/tsc.ts` calls
`executeCommandLine`. `Herebyfile.mjs` supplies that file as the single esbuild
entry for the tsc task; its separate project builder uses `tsc -b`. Thus the
compiler project's 77 config roots are checking/emission inputs, not all runtime
entries for the CLI. Config inputs retain the upstream parser's literal-files
then glob order, deduplicated at their first normalized path. Building every
compiler library root does not reproduce the CLI entry.

`git ls-remote --heads origin codex/import-cycles` returned no branch at both
checks. I did not alter `internal/lower/modules.go`, its `moduleOrder` function,
or its cycle and type-only-edge policy. This unit owns the new `rootOrder`
wrapper and the cardinality/orchestration change in Lower. If the cycle worker
changes the single-entry walk's signature or makes it stateful across calls,
that wrapper is the integration boundary. Cycles are still refused on this
branch. No native emitter files or `internal/oracle/oracle_test.go` were edited.
Only the counts gate was extended to record the new project fixture.

## Commands and observations

Every test command wrote its output to a log before that log was read.
The final shells sourced `/workspace/adamic-tools/env.sh`.

`bash cloud/setup.sh > /tmp/adamic-multi-root-setup.log 2>&1` succeeded:

```
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (97s)
setup: done in 97s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Go was 1.27.1, clang 20.1.8, Node v24.19.0; `nproc` printed `5`.

`gofmt -l cmd internal`, `git diff --check`, and `go vet ./...` all produced no
findings. Their logs are `/tmp/adamic-multi-root-format.log` and
`/tmp/adamic-multi-root-vet.log`.

Final touched packages:

```
go test ./internal/load ./internal/lower ./cmd/adamic -count=1 -timeout 30m
ok github.com/system-inc/adamic/internal/load 0.857s
ok github.com/system-inc/adamic/internal/lower 8.582s
ok github.com/system-inc/adamic/cmd/adamic 0.677s
```

Log: `/tmp/adamic-multi-root-packages.log`. This includes the three-root Node
oracle and sanitized/leak-checked native run, generated JavaScript, both CLI
build forms, 29 option-refusal cases, inherited config/glob order, extra
strictness, host/prelude contracts, ambiguous aliases, JavaScript dependency
checking, and declaration-root behavior.

For roots `third.a`, `first.a`, `second.a`, Node importing the original `.a`
sources through the existing oracle loader and both Adamic backends produced:

```
leaf
third leaf
shared leaf
first shared
second shared
```

The excluded file did not run, and shared modules ran once.

```
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m
ok github.com/system-inc/adamic/internal/oracle 18.399s
```

Log: `/tmp/adamic-multi-root-counts-final.log`. Recording with
`-args -update-counts` had changed only the new row:

| Fixture | Allocations | Frees | Retains | Releases | Peak | Regions |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| internal/lower/testdata/multi_root/tsconfig.json | 4 | 4 | 2 | 8 | 1 | 0 |

Filtered existing module oracle:

```
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/modules|^TestNativeAgreesWithNode$/internal/load/testdata/0.1/compile/07_modules' -count=1 -timeout 30m -v
PASS internal/oracle/testdata/modules/main.a
PASS internal/load/testdata/0.1/compile/07_modules/main.ts
native hits=0 misses=6
node hits=0 misses=4
ok github.com/system-inc/adamic/internal/oracle 0.468s
```

Log: `/tmp/adamic-multi-root-module-oracle.log`. An earlier overly narrow filter
selected no fixtures; the command above corrected that and actually ran both.

Pinned Node declaration smoke check:

```
npm install --prefix /tmp/adamic-multi-root-node --ignore-scripts --no-audit --no-fund @types/node@25.3.3
go run ./cmd/adamic build --project /tmp/adamic-multi-root-node/tsconfig.json -o /tmp/adamic-multi-root-node/program --sanitize
/tmp/adamic-multi-root-node/program
```

Each command exited 0; the native build and run had no output. The `.a` source
imported Adamic's panic and handled `JSON.stringify(undefined)` with a fallback;
the project selected `types: ["node"]` and `lib: ["es2024"]`.

I attempted `ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./...`, logged at
`/tmp/adamic-multi-root-full-gate.log`. After observing it still running beyond
seven minutes, including bridge archive builds and unrelated stage-1 port
suites, I stopped that test process tree and used the unit's scoped-gate
exception. The attempted full run exited 143 and is **not** a full-gate pass.
The scoped commands above ran against the final implementation.

## Mutants

The committed [mutants.py](mutants.py) creates Go overlays in scratch space and
never installs a mutant in the working tree. Every mutant test exited 1 and
printed a test failure; none was counted as caught by a Go build failure or
clang warning. The option mutants remove the explicit refusal but still
normalize the required flags: an innocent source is wrongly accepted with nil
error, so the dedicated policy test alone catches the accepted weak config.

| Mutant | What caught it |
| --- | --- |
| wrong-root-order | Native stdout differed from Node in both explicit-root and project cases. |
| repeated-shared-modules | Native printed leaf/shared modules repeatedly; Node printed each once. |
| weakened-required-options | All 18 required-true options were accepted; their refusal subtests failed. |
| enabled-checker-bypass | All 6 required-false guards were accepted; their refusal subtests failed. |
| wrong-module | The commonjs config was accepted; the module refusal subtest failed. |
| wrong-module-detection | The legacy config was accepted; the moduleDetection refusal subtest failed. |
| wrong-module-resolution | The node16 config was accepted; the moduleResolution refusal subtest failed. |
| old-target | The es2020 config was accepted; the target refusal subtest failed. |
| wrong-class-fields | False useDefineForClassFields was accepted; its refusal subtest failed. |
| unchecked-javascript | allowJs with false checkJs was accepted; the dependency guard test failed. |
| unsound-project-json | A required string accepted JSON.stringify(undefined); the sound-prelude test failed. |
| ambiguous-source-alias | The .a root silently selected its .a.ts sibling; the alias-refusal test failed. |
| missing-fallback-console | The project lost console entirely; the project oracle failed with checker diagnostics. |
| missing-adamic-glob-aliases | Include globs dropped a.a and b.a; the root-order test failed. |
| drop-direct-declarations | Direct type queries lost the declaration root; the declaration regression test failed. |
| declaration-only-guard | Lower panicked indexing the missing executable entry; the declaration-only test failed. |

Commands used for the complete mutant set:

```
python3 stage3/multi-root/mutants.py /tmp/adamic-multi-root-mutants
python3 stage3/multi-root/mutants.py /tmp/adamic-multi-root-mutants missing-adamic-glob-aliases
python3 stage3/multi-root/mutants.py /tmp/adamic-multi-root-mutants drop-direct-declarations declaration-only-guard
```

The first run contained the original 13 mutants; the later commands ran the
three added cases after their regression tests were added. The current script
runs all 16 when no selection is supplied. Summary logs are
`/tmp/adamic-multi-root-mutants.log`, `/tmp/adamic-multi-root-glob-mutant.log`,
and `/tmp/adamic-multi-root-declaration-mutants.log`; each overlay has a separate
named test log in the scratch directory.

## Limits and inferences

The observations establish the selected root/program/option/prelude contracts
and their ability to fail under the listed mutants. They do not establish a
native tsc or an implementation of arbitrary Node declarations. The original
upstream config is deliberately refused because it explicitly weakens strict
sub-options and uses incompatible module/target/checker settings.

The new wrapper should compose with a pure single-entry cycle walk that keeps
its signature. That is an inference from the unchanged boundary, not an
observed merge with the unpushed cycle branch. Project-reference build graphs,
cycle initialization, type-only import changes, host console lowering, complete
upstream compilation, and symlink-root identity remain outside this unit.
