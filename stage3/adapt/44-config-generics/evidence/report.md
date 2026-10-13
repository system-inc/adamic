# Adaptation 44 measurements

Built twelve type-only edits, one file, sixteen lines added and removed.
Measured against main 45487a809f89885a3fc651cd590e7dabf31362dc.
Both landing lanes PASS: 106366 passing, one identical sanctioned failure, zero pending.
Eighteen distinct mutants caught; ten JS artifacts and two public API artifacts identical.
Raw-field/mutation contracts remain unfinished; no native proof or rebased-43 integration claimed.

## Observations

| Command | Result |
|---|---|
| Main stage3/lane/run.sh /tmp/adapt44-before | exit 0, 633.108 seconds; apply exit 0; oracle exit 1 |
| This branch stage3/lane/run.sh /tmp/adapt44-final | exit 0, 613.303 seconds; apply exit 0; oracle exit 1 |
| check-types.cjs /tmp/adapt44-final/adapted-tree /tmp/adapt44-physical-types.json | exit 0; consumer diagnostics 0; wrong-return mutant 4 diagnostics; subtype witness 1 diagnostic; declined raw-field generic 9 diagnostics |
| proof.cjs /tmp/adapt44-before /tmp/adapt44-final /tmp/adapt44-proof | exit 0; 12 reviewed edits; 13 guard mutants exit 1; all output comparisons identical |
| git diff --check | exit 0 |

The lane invokes stage3/apply.sh and the full stage3/oracle, not a filtered
confirmation run. The adapted oracle's npm ci and npm run build both exit 0;
full npm test exits 1 after 450.739 seconds. Main and adapted each have exactly
106366 passing, 1 failing and 0 pending, with the same failure:
`unittests:: Public APIs for typescript.d.ts should be acknowledged when they change`.
Their only baseline difference is `api/typescript.d.ts`, and their baseline.diff
bytes are identical. The existing lane sanctions that acknowledgment and exits 0.
No new API difference was accepted or blessed.

`proof.json` records SHA256 for all ten actual built JavaScript files and both
public declarations, `built/local/typescript.d.ts` and
`built/local/typescript/typescript.d.ts`. All hashes match main. The whole edited
source's stock transpilation is also byte-identical. All 711 src TypeScript files
are checked: the file set is exact, and only the reviewed converter edits differ.
The predicate's entire source text is exact. Reapplication changes zero sites;
LF and CRLF reconstruction both pass.

The runtime tests call both real built APIs, covering raw falsy values and
primitives, invalid elements, compiler defaults, type acquisition, watch options,
path normalization, inherited keys, a one-read getter, a cyclic object, functions
as invalid compileOnSave input, and the raw compileOnSave boolean mutation.
The private converter's actual extracted function executes too: falsy raw input
returns undefined even when defaults exist, and normalization mutates strict
true to false. Its execution is separate from API entrypoint observations.
`runtime.json.gz` holds the exact compared observations; `types.json` holds
checker witnesses. No array-element predicate proof is inferred from these runs.

## Mutants and the check that caught each

- Twelve reviewed source mutations: parseConfig; S04; S07; S10; S11; S12;
  S13; S14; S15; convertJsonOption; convertJsonOptionOfListType; list-element-caller.
  Each parameter any is changed to an unreviewed unknown, or the list is copied
  before mapping. Each actual adapter process exits 1 at that site's exact source
  guard, before writing the input. These are source-contract guard proofs, not
  independent native soundness proofs. Each has its own log.
- S02 predicate input any changed to unknown: the protected predicate text guard
  rejects it, exit 1. The real predicate remains unchanged.
- Actual compiler-options worker return changed from CompilerOptions to string:
  stock checking produces four diagnostics. No adapter guard is involved.
- Actual private converter falsy branch changed to return defaults: execution
  succeeds, and the expected undefined comparison throws AssertionError.
- One byte appended to a real emitted _tsc.js artifact and, separately, to a real
  typescript.d.ts artifact: each SHA256 comparison throws AssertionError.
- Real measured lane counts copied with passing decremented by one: exact
  main/adapted count comparison throws AssertionError.

There are 18 distinct mutants: 13 guards, 1 checker, 1 runtime, 2 artifact bytes,
1 lane count. All were run and caught by their named checks.

## Toolchain and limits

GOPROXY was set to `https://proxy.golang.org|direct` before setup. Initial restored
workspace setup completed in 31 seconds. Setup was rerun on this main-based
checkout after restoring its exact submodule pins; it exits 0. Its printed
cumulative readiness lines are:

```text
setup: node ready (0.037s)
setup: go ready (0.044s)
setup: submodules ready (0.115s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0.261s)
setup: markdown dependencies ready (1.220s)
setup: go build ready (221.919s)
setup: test binaries deferred (use --warm-tests) (222.048s)
setup: build cache warm (222.050s)
setup: done on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB (222.089s)
```

The printed environment file was sourced. nproc is 5; Node v24.19.0,
Go 1.27.1, clang 20.1.8, stock TypeScript 6.0.3. Full suites ran serially with
NODE_OPTIONS=--max-old-space-size=1400 and eight workers. Setup's required build
cache warming was run; no whole Go test gate or package confirmation tests ran.

A discarded draft lane stopped before its oracle, after the initial generic
array inference failure exposed two TS2345 diagnostics. It is not a measurement.
The final signatures and only accepted lane are recorded above. Referenced 43's
unchanged adapter was also tried in scratch and rejects current main's changed
convertConfigFileToObject diagnostic sink; its log is retained. This is not a
failure of 44's main-plus-44 pipeline and not a claim of 43 integration.

See ../README.md for every remaining raw/public/callback/normalization site and
why it is unfinished. No new cast hides those sites. No .a fixture was added;
there is no counts.md change. The checked-in patch-set.md change is only the 44
row; adapted-patch-set.md records the generated total, 79 files, 5263 additions,
5229 removals. All test output was sent directly to logs, never piped.
