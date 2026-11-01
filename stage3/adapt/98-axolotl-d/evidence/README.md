# 98d evidence

Made on main `470e16f8` with 98's adapter at `d26001bf`, from the repository root, in this order.
`before` is `stage3/apply.sh` without 98 (main as it was); `after` is `stage3/apply.sh` with it.

```sh
bash cloud/setup.sh --wasi-sdk && source /opt/adamic-tools/env.sh
git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git /tmp/ax/upstream
npm ci --prefix stage3/api --no-audit --no-fund
ADAMIC_TYPESCRIPT_SOURCE=/tmp/ax/upstream STAGE3_CACHE=/tmp/ax/cache bash stage3/apply.sh /tmp/ax/before   # logs/apply-before.log.gz, exit 0
ADAMIC_TYPESCRIPT_SOURCE=/tmp/ax/upstream STAGE3_CACHE=/tmp/ax/cache bash stage3/apply.sh /tmp/ax/after    # logs/apply-after.log.gz, exit 0
go build -o /tmp/ax/census ./stage3/census/tool
/tmp/ax/census <tree>/src/compiler <run>/census.jsonl     # census/*.jsonl.gz and *.log
(cd <tree> && npm ci && npm run build)                     # logs/build-*.log, both exit 0
```

The first apply with 98 refused the tree: an earlier adaptation marks the worker's tryResolve value members
readonly, a shape the adapter hadn't inventoried. The adapter now checks those members structurally, and the
apply above is the rerun.

**Census inputs.** Each census ran on a copy of the tree's `src` alone, which is what apply writes before
any npm install, so node_modules can't change the program:

- ms-before: before
- ms-after: after
- m-mut1: after with row 1's `!` dropped
- m-mut2: after with `ResolutionWithFailedLookupLocations.files` unwidened

An earlier pair of runs on the full trees overlapped each tree's `npm ci`. Their only difference was one
external `node_modules/@types/source-map-support` line, which had appeared mid-run, so they were replaced and
are not used.

## The files

- **own-file.json:** own-file counts for every one of the 79 source files under src/compiler, by own-file.py, which calls
  stage3/meter/report.py's `report()` unchanged. Only three files change: moduleNameResolver 3 to 0,
  resolutionCache 1 to 0, sourcemap 1 to 0. The total goes from 114 to 109.
- **diagnostic-lines.json:** every distinct diagnostic line across all census records. 114 before, 109 after,
  0 new, and the 5 gone are the five rows.
- **mutants.json:** each mutant against after. Row 1's `!` dropped: moduleNameResolver 0 to 1. files
  unwidened: resolutionCache 0 to 1. Nothing else changes in either.
- **typescript.d.ts.diff:** the built typescript.d.ts from both stock builds. Exactly the two packageId lines.
- **built-js.txt, typescript.js.diff, _tsc.js.diff, run.js.diff:** of the ten built/local JavaScript files,
  three differ, each only at appendSourceMap.
- **transpile.json:** by transpile.cjs. ts.transpileModule of the five edited files: types, program and
  resolutionCache are byte-identical, sourcemap differs at its site, and moduleNameResolver differs at one line.
  That line is 286's shorthand spelled out as `isExternalLibraryImport: isExternalLibraryImport`, the same object.
- **witness.json:** witness.cjs on both built typescript.js files. With the plain array, both give
  `["present"]`. With the getter, without 98 is 3 reads and `[null]`, and with 98 is 2 reads and `["present"]`.
- **idempotence-*:** the adapter's second run on the applied tree. 0 edits, and the SHA-256 of all 711
  source files is the same before and after it.
