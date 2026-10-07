# Shared YAML gate input proof

ADAMIC_YAML_LIBRARY shares the css-printer prefix and lockfile with ADAMIC_CSS_PRINTER_LIBRARY and ADAMIC_TS_PRETTIER. All three exports point to the same directory under --gate-inputs; ordinary setup unsets them. npm preparation deduplicates this directory, so it runs once. The shared lock pins yaml 2.9.0, prettier 3.9.6 and yaml-unist-parser 3.2.0.

The YAML branch's loaders resolve createRequire(prefix + '/package.json'); its unist loader imports prefix/node_modules/yaml-unist-parser/dist/parse.mjs directly. Both layouts work. YAML formatting, TypeScript/CSS/document formatting smoke checks, all version assertions, and the unist module's direct import/parse passed (resolution.log and unist-resolution.log).

Lock generation used the existing SHA512-verified npm 11.9.0 bootstrap. Installation uses npm ci --ignore-scripts --bin-links=false and lockfile integrity, with the existing five-component warm key and installed-byte validation. No new cache implementation was introduced. Real cold installation, warm skip, lockfile newline forcing reinstall, restoring the original lock, and byte-identical uncached installation all passed (cache-proof.log). Individually dropping lock, manifest, bootstrap, helper or Node from the key fails an assertion. Missing YAML export and yaml 2.9.1 manifest mutants also fail their intended assertions. All seven mutants and their actual failed assertions are recorded beside this report.

## Scratch YAML test

Fetched stage1-format/yaml with --no-recurse-submodules and created a detached scratch worktree at exactly 4a70daa0349b3ec20948f3995f0d57eb1dd70724 under /tmp/adamic-gate/yaml-input-proof. No YAML branch source was merged or committed.

That older tree lacks bridge/tsgo/archive. Full gate setup was therefore completed on devtools/setup-fast, and its exported inputs were used by the unchanged YAML worktree. The workspace/temporary volumes filled during early attempts at independent TypeScript checkout/archive compilation. Disposable cache artifacts from these proofs were reclaimed or moved to /tmp; the completed setup uses /tmp/adamic-gate/yaml-input-tools. Go downloads use the previously documented GONOPROXY workaround for klauspost/compress with normal checksum verification.

```sh
source /workspace/adamic-tools/env.sh
export ADAMIC_TOOLS=/tmp/adamic-gate/yaml-input-tools
export GOCACHE=/home/agent/.cache/go-build
export GOMODCACHE=/tmp/adamic-gate/setup-modules-proof/with
export GONOPROXY=github.com/klauspost/compress
bash /workspace/adamic/cloud/setup.sh --gate-inputs > /tmp/yaml-input-setup-complete.log 2>&1
source /tmp/adamic-gate/yaml-input-tools/env.sh
export XDG_CACHE_HOME=/tmp/adamic-gate/yaml-input-user-cache
ADAMIC_GATE_UNCACHED=1 go -C /tmp/adamic-gate/yaml-input-proof test ./stage1/cohere/yaml \
  -run '^TestLexerMatchesGo$' -count=1 -timeout=30m -v > /tmp/yaml-input-test-final.log 2>&1
```

Full setup exited 0. Its build-flags line is in setup.log: base HEAD 814e46a597b28b94adb7846e0bcb78d6930abc89 plus this working-tree change, nproc 5, cpu.max 400000 100000, Go 1.27.1, clang 20.1.8, Node 24.19.0, loads and cache mode. XDG_CACHE_HOME redirected the test's native runtime cache off the full workspace filesystem.

TestLexerMatchesGo ran and passed without skipping: 36 repository files, 8,732 complete/chunked cases, 4,181,796 identical answer bytes across Go, native ASan/UBSan/LSan, Node source, emitted JavaScript and yaml 2.9.0. See yaml-test.log. TestUnistMatchesGo was not run; its pinned module and exact direct import were verified.

Focused cloud setup, npm, stage3, module and gate-input suites passed; unrelated opt-in integrations remained opt-in. Full YAML suite/full gate were not rerun. Generated root/cohere workspace checksum changes from setup were restored after proof. Only cloud changes are committed.
