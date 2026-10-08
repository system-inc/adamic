Built: rebased eighteen existing ports onto lint integration b46914832; owned sources unchanged; no unclaimed rule remains.
Commits: rebased/compiler tree 2559be4d7; main c7991b900; published 0b1552c2d ancestry retained for a normal push.
Commands: all owned byte oracles, sanitizer corpora, handles, registry descriptor/mutation tests and vet pass; older gate 739.135s; setup 81s, nproc 5.
Mutants: eighteen rule mutations, scope export, three regression guards and retained handles detected; registry rejection/mutation proofs pass.
Uncovered: full required-input repository gate, dynamic native RegExp, whole-rule JavaScript/shared checker context and three parked React analysis rules.

The integration update moves the remaining legacy syntax rules onto the shared
registry. The rebase completed without conflicts. Compiler, native runtime,
checker, TypeScript parser and owned type-aware source trees are unchanged from
published 0b1552c2d. The compiler was rebuilt; its metadata records revision
2559be4d725f51fd4024cf3fb22a5b645918cfe3 and vcs.modified=false. No shared file
was manually edited. The new registry regenerated successfully; all descriptor,
module discovery and mutation tests passed in 0.332s. Vet passed.

All eighteen owned rule ports were re-green against unmodified production Go
cohere, including full finding/fix/suggestion bytes, clean comparison-only rule
mutants, and normal/ASan/UBSan runs over all 287 frozen repository and 77 pinned
compiler roots. The older twelve-rule gate passed in 739.135s. The recent batches
accepted 176 and 272 controls; default streams match at 52,437 and 73,462 bytes.
Their options and both corpora agree under sanitizers. Corpus silence remains
held by positive controls and mutants, rather than counts alone.

| Mutation | First differing byte from Go |
| --- | ---: |
| numeric-prefix | 14792 |
| has-own-fix-span | 18785 |
| spread-parens | 314 |
| promise-condition | 16849 |
| spread-await-edit | 22057 |
| lost-write-span | 65053 |
| nullish-suggestion | 361 |
| qualifier-fix | 22460 |
| private-read | 4152 |
| scope-export-mutant | 29967 |
| global-provenance | 5940 |
| global-declaration-span | 85544 |
| timer-string | 63292 |
| fragment-id | 735 |
| undef-id | 358 |
| adjacent-judgment | 32497 |
| default-id | 772 |
| nested-id | 18310 |
| sort-id | 53364 |
| computed-trivia | 66117 |
| jsx-attribute-flag | 18244 |
| class-fallback-flag | 933 |

Every listed mutant compiled and ran with exit 0 and empty stderr; Go's byte
oracle caught it. The older handle-retention mutations were caught by required
refusal. Both owned question probes, symbol-declaration-syntax and
unicode-node-text, reject released programs with exit 70 and the exact invalid-
or-released-checker-handle panic under sanitizers.

| Batch / corpus | Native seconds | Go seconds |
| --- | ---: | ---: |
| sixth / repository | 0.829 | 0.303 |
| sixth / compiler | 3.488 | 0.500 |
| seventh / repository | 5.363 | 0.300 |
| seventh / compiler | 37.151 | 0.975 |

These are single-process observations during concurrent landing checks, not an
isolated benchmark. Native remains slower. Fresh streams, exact commands and
timings, compiled source hashes and logs are in
[landing-registry-b46914832](landing-registry-b46914832/summary.json).

The commands ran with output redirected to named wave-22-registry log files:

```sh
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-22-typescript-pinned go test ./stage1/cohere/typeaware -run '^TestWave22(AgreementAndMutants|NextAgreement|ThirdAgreement|FourthAgreement)$' -count=1 -v -timeout=30m
go build -o /workspace/wave-22-seventh-adamic ./cmd/adamic
go build -buildmode=c-archive -o /workspace/wave-22-seventh-checker.a ./bridge/tsgo/archive
python3 stage1/cohere/typeaware/rules/wave-22-sixth/validate.py
python3 stage1/cohere/typeaware/rules/wave-22-seventh/validate.py
python3 stage1/cohere/typeaware/rules/wave-22-seventh/prove_flags.py
go run ./cmd/lint-registry
go test ./stage1/cohere/lint/registry -count=1 -v
go vet ./bridge/tsgo/... ./stage1/cohere/typeaware
```

The full repository/required-input correctness gate was not run. Selected
correctness checks supplied their pinned TypeScript inputs and did not skip.
Generic bridge, Unicode question mutation and Node/runtime suites were not
repeated on this registry-only update; their core source trees are unchanged
and their preceding green runs are documented in LANDING_MAIN_C7991B900.md.
Fresh probes confirm that dynamic RegExp still fails at options.a:12:46 and
whole-rule JavaScript still refuses the native checker binding. Shared
RuleContext still exposes no checker-program binding. The three prior React
HIR/SSA/capture claims remain parked under #dnv6f2c; no integration coverage for
those gaps is claimed.

The refreshed snapshot has 640 origin refs, 33 distinct Markdown claim
blobs and all 197 ranked rules. Zero rules remain unported and unclaimed. No new
reservation was made. Only codex/typeaware-wave-22 is pushed; integration merges.
