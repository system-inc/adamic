# Reproduction and executed commands

All work used private scratch. `/tmp/env-presence-1965` was later relocated to disk-backed `/var/tmp/env-presence-1965`, with the old path retained as a symlink. The shared checkout was read only. Node was v24.19.0; Go was installed as go1.27.1 in scratch. Set TMPDIR, GOTMPDIR, GOCACHE, GOMODCACHE, GOPATH, npm_config_cache, ADAMIC_BUILD_CACHE_DIR and STAGE3_CACHE to the corresponding scratch directories; never reuse HOME. No Loom command was invoked.

Setup used a private shared clone of /workspace/adamic, fetched 8e3a784aa5a601f57342c07b68f145fe56252c21 from https://github.com/system-inc/adamic.git, then `git worktree add --detach` at that exact base. `git submodule update --init --recursive --depth 1` completed. Upstream was cloned with `--depth 1 --branch v6.0.3` from https://github.com/microsoft/TypeScript.git; rev-parse HEAD confirmed 050880ce59e30b356b686bd3144efe24f875ebc8. `npm ci` completed in each worktree's stage3/api. ADAMIC_TYPESCRIPT_SOURCE names that upstream clone.

The ownership search was `rg -n 'getEnvironmentVariable|sys.ts' stage3/adapt --glob owners.json --glob targets.json --glob adapt.cjs`; ownership.json records the files inspected. No top-level target statement is pinned by an existing adapter.

From the evidence worktree, with NODE_PATH pointing to its stage3/api/node_modules:

```sh
node stage3/census/env-presence-1965/witness.cjs
ENV_PRESENCE_COUNTS=/tmp/env-presence-1965/site-control.jsonl node --require ./stage3/census/env-presence-1965/counter.cjs stage3/census/env-presence-1965/control.cjs /tmp/env-presence-1965/TypeScript/src/compiler/sys.ts
stage3/apply.sh --build-product /tmp/env-presence-1965/baseline
```

run-count.sh records the first count exactly, copying baseline to counting, inserting only the counting call, retaining the original condition, and running the full oracle. Its worker was OOM-killed because /tmp was RAM-backed; that result is explicitly invalid. After relocation, the successful full retry was:

```sh
ENV_PRESENCE_COUNTS=/tmp/env-presence-1965/counts-retry.jsonl NODE_OPTIONS=--require=/tmp/env-presence-1965/work/stage3/census/env-presence-1965/counter.cjs stage3/oracle/run.sh /tmp/env-presence-1965/counting /tmp/env-presence-1965/count-oracle-retry
```

Only after its complete 106,366/1/0 result and zero counterexamples was the implementation worktree and adapter constructed. The implementation worktree was separately added with `git worktree add --detach` at the same base, then switched to the requested new branch. Candidate was copied from baseline; the new adapt.cjs was run against it. Restoring was copied from candidate; mutant.cjs was run against restoring and `cmp` proved its sys.ts byte-identical to baseline.

The census binary used the existing source unchanged:

```sh
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/env-presence-1965/overlay
python3 stage3/meter/entry_overlay.py /tmp/env-presence-1965/overlay /tmp/env-presence-1965/entry-overlay
go build -buildvcs=false -overlay=/tmp/env-presence-1965/entry-overlay/overlay.json -o /tmp/env-presence-1965/census ./stage3/census/latent/tool
```

Each of baseline, candidate and restoring was measured with the same binary and exact flags (substitute TREE and OUTPUT):

```sh
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 CENSUS_MODULE_FORMAT=CommonJSBundle LATENT_ONLY_FILE=TREE/src/compiler/sys.ts /tmp/env-presence-1965/census TREE/src/tsc/tsc.ts OUTPUT.jsonl
```

`delta.py main-sys.jsonl candidate-sys.jsonl delta.json` records every new/gone/changed finding in full with phase. A second delta against restoring is empty. span-audit.json additionally checks all 184 retained byte-span fields against precisely the 30-byte condition deletion; all 182 retained findings match.

`contract.cjs ADAPT_CJS BASELINE_SYS CONTRACT_TREE` runs LF/CRLF replay, exact restoring bytes, drift, half-adapted, duplicate, mixed-state and mixed-line-ending controls, counting every source write. `check-javascript.cjs BASELINE CANDIDATE javascript.json` checks all 604 source files in both comment modes. `stock.cjs TREE OUTPUT.json` checks the entire 79-root compiler input with stock 6.0.3 and its project options; both outputs have zero diagnostics.

The unmodified CLI was built with `go build -buildvcs=false -o /tmp/env-presence-1965/adamic ./cmd/adamic`; `adamic types` received all 79 sorted candidate compiler roots, listed in types-roots.json.gz, and exited zero. types.log.gz is its full output.

In baseline and candidate independently, `npm ci` followed by `node_modules/.bin/hereby dts` completed. Both built/local/typescript.d.ts and built/local/tsserverlibrary.d.ts were compared byte for byte; declarations.diff is the actual empty unified diff, and declarations.json has sizes and SHA-256 values.

Final integration, from the implementation worktree, with the scratch environment above and PYTHONDONTWRITEBYTECODE=1:

```sh
stage3/lane/run.sh /tmp/env-presence-1965/final-lane > /tmp/env-presence-1965/final-lane.log 2>&1
```

This runs the unchanged apply.sh and whole unfiltered oracle, eight workers, without a filter or altered expectation. The implementation is committed as one commit only after PASS, then pushed without force. All evidence pushes are fast forwards to the separately authorized evidence branch.
