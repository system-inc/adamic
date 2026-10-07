# Shared ESTree gate input

--gate-inputs exports ADAMIC_ESTREE_LIBRARY to the existing css-printer npm prefix, alongside ADAMIC_CSS_PRINTER_LIBRARY, ADAMIC_TS_PRETTIER and ADAMIC_YAML_LIBRARY. One lockfile and one npm ci install provide prettier 3.9.6, yaml 2.9.0, yaml-unist-parser 3.2.0, @typescript-eslint/typescript-estree 8.65.0 and typescript 6.0.3. Ordinary setup unsets ESTREE_LIBRARY. ESTREE_CORPUS and ESTREE_BENCHMARK are untouched. TypeScript source remains pinned to 050880ce59e30b356b686bd3144efe24f875ebc8.

The unchanged ESTree library.mjs at cbcc1575 imports typescript-estree/dist/index.js and prettier/index.mjs from that root. Both modes run successfully (raw-resolution.log, prettier-resolution.log). The other consumers' prefix-relative imports, version assertions, YAML unist parse and CSS/TS/YAML Prettier plugin resolution all pass (shared-resolution.log). npm phase deduplicates aliases, verified by test_shared_prettier_installed_once.

Installation uses the existing integrity-verified npm bootstrap and npm ci --ignore-scripts. No new cache implementation: the existing lock/manifest/bootstrap/helper/Node key and installed-byte verification apply. cache-proof.py records real warm skip, uncached byte-identical tree, lockfile-byte change causing reinstall and original-lock restoration. Five drop-key mutants are caught; a missing ESTree export mutant fails the export assertion, and a TypeScript 6.0.2 mutant fails the exact-pin assertion. Setup regression and generic npm cache/integrity suites pass (optional integrations remain opt-in).

Scratch proof: git worktree add --detach /tmp/adamic-gate/estree-input-proof cbcc1575; initialize cohere at 715ba94f3608a6500086b1076ce5cb7e51b836db. Its nested TypeScript pin is exactly the existing checkout 8d550c837c90bd1805b047b7eeccc2baac2d5e7a; the scratch tree links that identical checkout because Git refuses a shallow reference clone. No ESTree branch source is merged or committed.

Focused test command, stdout/stderr to estree-test.log:

```sh
source /workspace/adamic-tools/env.sh
export ADAMIC_ESTREE_LIBRARY=/tmp/adamic-gate/estree-shared
export GOMODCACHE=/tmp/adamic-gate/setup-modules-proof/with
export GONOPROXY=github.com/klauspost/compress
export XDG_CACHE_HOME=/tmp/adamic-gate/yaml-input-user-cache
ADAMIC_GATE_UNCACHED=1 go -C /tmp/adamic-gate/estree-input-proof test ./stage1/cohere/estree -run '^TestOriginalLibraries$' -count=1 -timeout=5m -v
```

Full gate, corpus and benchmarks are not run. Full setup was not rerun; the real existing npm preparation helper, env generation, alias deduplication and setup regression tests exercise this input's integration.

TestOriginalLibraries passed without skipping: raw 78/78 identical; postprocessed 75/78 identical with precisely its three asserted known gaps. Without ADAMIC_ESTREE_LIBRARY it skips with the required-package message (without-input.log). The initial compile exhausted /tmp; deleting only a prior disposable scratch cache backup freed space and the retry passed.
