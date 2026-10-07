# GraphQL printer gate input

ADAMIC_GRAPHQL_PRETTIER is opt-in under --gate-inputs and shares the existing css-printer prefix with CSS, TypeScript, YAML and ESTree. That one npm ci install now includes graphql 17.0.2 and prettier 3.9.6 together. Their exact versions and SHA512 archive integrity are held in the checked-in package-lock.json. The existing five-component key (lock, manifest, bootstrap, helper, Node), installed-byte/mode validation and ADAMIC_GATE_UNCACHED bypass are reused. Ordinary Linux setup unsets the new export; the portable macOS lane uses the same catalog.

Read printer_test.go, gaps_test.go, README.md and testdata/prettier.mjs at area/stage1-format a48ae4c3. The loader uses createRequire(prefix/package.json) and asserts Prettier 3.9.6. Actual module resolution, GraphQL parsing and Prettier formatting pass (resolution.log). The shared directory is still prepared once, verified by the existing deduplication test.

A detached scratch worktree at a48ae4c3 initializes cohere at 715ba94f3608a6500086b1076ce5cb7e51b836db. Its nested TypeScript commit 8d550c837c90bd1805b047b7eeccc2baac2d5e7a is the identical existing checkout, linked into the scratch tree. No stage1-format source is merged or committed.

```sh
source /tmp/adamic-gate/node-pin-tools/env.sh
export ADAMIC_GRAPHQL_PRETTIER=/tmp/adamic-gate/graphql-printer-shared
export GOCACHE=/home/agent/.cache/go-build
export GOMODCACHE=/tmp/adamic-gate/setup-modules-proof/with
export GONOPROXY=github.com/klauspost/compress
export XDG_CACHE_HOME=/tmp/adamic-gate/yaml-input-user-cache
ADAMIC_GATE_UNCACHED=1 go -C /tmp/adamic-gate/graphql-printer-proof test ./stage1/cohere/graphql/printer -run '^TestPrinterUpstreamPreflight$' -count=1 -timeout=5m -v > preflight.log 2>&1
ADAMIC_GATE_UNCACHED=1 go -C /tmp/adamic-gate/graphql-printer-proof test ./stage1/cohere/graphql/printer -run '^TestPrinterWhitespaceGap$' -count=1 -timeout=5m -v > whitespace.log 2>&1
unset ADAMIC_GRAPHQL_PRETTIER
go -C /tmp/adamic-gate/graphql-printer-proof test ./stage1/cohere/graphql/printer -run '^TestPrinter(UpstreamPreflight|WhitespaceGap)$' -count=1 -timeout=1m -v > without-input.log 2>&1
```

Both tests run and pass with the input; both skip with the variable unset. Preflight checks 3,564 texts in each of defaults, narrow, tight and tabs, against both npm and the embedded fork: 2,513 formatted answers, 1,046 shared refusals and five explicitly asserted whitespace differences, zero unexpected differences. The separate whitespace proof checks native/Node/Go refusals and npm/embedded empty answers. Initial preflight failed solely because Go's cache exhausted the workspace disk (preflight-disk-failure.log); recently generated cache bytes were retained in scratch, and the retry passed.

cache-proof.py demonstrates a real warm skip, byte-identical uncached tree and stamp, lockfile newline forcing reinstall, and restoration. No new cache implementation was added. Five drop-key assertion mutants, a missing GraphQL export mutant and graphql 17.0.1 pin mutant are caught (mutants.py and logs). Python setup suite passes 35 tests; five existing optional integrations remain opt-in.

ADAMIC_MARKDOWNWIDTH_DEPS remains always exported. Both manifest/lockfile and the actual installed prefix match emoji-regex 10.6.0, get-east-asian-width 1.6.0 and narrow-emojis 0.0.3 (width-confirmation.log). ADAMIC_GRAPHQL_PRINTER_BENCH and ADAMIC_MARKDOWNBLOCKS_CENSUS remain measurement opt-ins and are absent from the gate-input catalog, held by an assertion.

The full repository gate, full GraphQL printer package, full --gate-inputs setup and benchmarks were not rerun. The existing real npm preparation helper, env generation, deduplication and both consuming oracle tests cover this addition. These are functional proofs, not performance benchmarks. Node pin commit f27c09c5ce809e847536a61f8e120102a981d24e precedes this separate gate-input commit.
