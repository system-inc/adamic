Rebased twelve owned rule ports onto current main e8ba3d5d81de4d3773c723914fccd4c76248b965; no new claims.
Commits: source rebase 111957675c00710ce539a9dabef2d1b2f5907640; green helper branch b28c05998fd28afb85da82464ac9e695d6139853; this commit adds the evidence.
Checks: 280 fixtures and 3,114 file/rule corpus comparisons match Go on source Node, emitted JavaScript and sanitized native; independent sentinel passes.
Mutants: all nine semantic mutants compile and exit successfully with empty stderr; only byte comparisons catch each on all three backends.
Not covered: three older ports on the changed shared harness, default integration, full gate and new throughput measurements; shared blockers retain the landing cap.

Commands, after source /workspace/adamic-tools/env.sh:

    python3 stage1/cohere/lint/rules/typescript-no-unnecessary-type-constraint/validate.py --scratch /tmp/wave15-landing-current-owned --typescript /tmp/lint-wave1-15-typescript --mutants
    python3 stage1/cohere/lint/rules/no-multi-str/validate.py --scratch /tmp/wave15-landing-current-literals --typescript /tmp/lint-wave1-15-typescript --mutants
    ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v

All three exit zero. Test output is captured in logs. TypeScript is pinned at v6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8. Corpus: 77 compiler files and 269 stage1 .ts/.a files, 346 files total. The six-rule comparison covers 206 fixtures and 2,076 corpus pairs, with 83,165 fixture bytes and 79,598,310 corpus bytes. The literal comparison covers 74 fixtures and 1,038 corpus pairs, with 26,425 fixture bytes and 39,640,224 corpus bytes. Unchanged Go tests for all nine rules pass. Complete findings, immutable fix/suggestion proposals and fixed source are compared. Corpus output omits repeated individually applied suggestion source. hashes.json records identical hashes and sizes for all four sides; gzip Go outputs preserve the oracle observations. The external sentinel passes in 0.706s, with one native and one Node miss and no hits.

Every mutant below was compiled anew and caught only by differing output on source Node, emitted JavaScript and sanitized native. Raw mutant observations are retained in the two evidence subdirectories.

- nexus-import-require-node-namespace: "'NodeFileSystem'" -> "'NodeWrongFileSystem'".
- structure-network-no-invalidate-cache-literal-key: ".text !== 'invalidate'" -> ".text !== 'invalidateNever'".
- structure-network-no-string-literal-query: "this.binding(unwrapped, node.text)" -> "-1".
- typescript-no-unnecessary-type-constraint: "kind !== 'AnyKeyword' && kind !== 'UnknownKeyword'" -> "kind !== 'AnyKeyword' && kind !== 'NeverKeyword'".
- typescript-prefer-as-const: "node.text === this.context.node(initializer).text" -> "node.text !== this.context.node(initializer).text".
- typescript-prefer-enum-initializers: "`${position + 1}`" -> "`${position + 2}`".
- no-multi-str: "raw.includes('\\u2029')" -> "raw.includes('NEVER')".
- no-nonoctal-decimal-escape: "previousNull = digit === '0'" -> "previousNull = true".
- no-octal: "digit <= '9'" -> "digit <= '7'".

Shared blockers reproduced on the final main revision, without editing shared sources:

- go run ./cmd/lint-registry exits 1: it requires stage1/cohere/lint/rules/next-no-assign-module-variable/rule.ts despite the port's rule.a.
- go test ./stage1/cohere/lint -run '^TestOwnedWitnesses$' -count=1 -v exits 1 while compiling profile_test.go:32, which ranges over the portFiles function instead of its result.
- The owned structure-tailwind-no-physical-direction/validate.py --prepare-only adapter exits 1: two compatibility-patch hunks no longer apply to the changed shared lint_test.go. Rejects are preserved. Consequently @next/next/no-assign-module-variable, @typescript-eslint/default-param-last and structure/tailwind-no-physical-direction have not been re-greened on this harness; their prior implementations and evidence survive the rebase.

The literal comparison separately reproduces nine parser exclusions (six JSX cases and 01.5, 0777.5, 0755n): Go accepts them; each Adamic backend explicitly refuses. All three decimal-escape bridge probes refuse with NotYet: shared lint repair serialization; multiline and octal bridge probes succeed. The six-rule standalone comparison does not certify the shared bridge, whose suggestion and multi-edit limitations remain recorded in its completion report. No default formatter, all-rule ordering or converging fixer is certified.

Setup and host: nproc=5. The first bash cloud/setup.sh attempt printed Go ready 0s, clang ready 1s and Node ready 1s, then failed because the isolated TypeScript submodule checkout was not complete. After initializing the pinned recursive submodules, setup printed Go ready 0s, clang ready 1s, Node ready 1s and submodules ready 1s. Its test warm-up failed on the same profile_test.go:32 compile error, so no successful warm/done timing was printed. Toolchain: Go 1.27.1, clang 20.1.8 and Node 24.19.0. Cached installed tools ran the successful standalone checks. Full gate and previous throughput benchmarks were not rerun.

Both owned branches were rebased again when main advanced from e011f8f6 to e8ba3d5d. The final fetch still showed e8ba3d5d, which is an ancestor of both. Registration merge conflicts reuse published integration efeb3f66, preserving main's existing scanner and volume coverage; no manual shared generator, harness or compiler changes were made. The helper branch is re-green and pushed; the rule branch remains blocked as described, so no new helper or rule claim was made.
