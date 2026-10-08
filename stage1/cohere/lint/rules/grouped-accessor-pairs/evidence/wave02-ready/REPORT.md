# Wave 02 ready rules

Branch: `lint-rules/wave1-02-ready`, started at `origin/lint-batch/wave2-02` 95968dd93ad0876245f181af63931c3134e14b3d and merged `origin/area/stage1-lint` 334509eea8a49b8085187e206c495cc6aa24c5c4. Merge commit fc040a7698d80dddbf8c3a45d4443338fecb3ca8. No rebase. All new unit changes are inside the three owned rule directories. Shared harness and compiler changes are taken from the authorized foundations.

Ready ports:

| Rule | Captured upstream combinations | Matched bytes, including owned selected/all witnesses |
| --- | ---: | ---: |
| @typescript-eslint/no-restricted-types | 52 | 12773 |
| base/consistency-no-bare-throw | 51 | 16617 |
| grouped-accessor-pairs | 160 | 30075 |

Every captured upstream source/rule/options combination and owned witness compares byte for byte against unchanged Go cohere on source Node, emitted JavaScript and ASan/UBSan native. Entries are now `.a`, with named kinds in rule.json and no numeric listener exports. Restricted types uses the shared Suggestion and SuggestionEdit types and reportRange. Bare throw uses the landed imports.normalizedFileName helper and honors upstream's `.test.a` exemption.

Certification command:

`go test -overlay=/tmp/wave02-certification-overlay.json ./stage1/cohere/lint -timeout 30m -run '^TestWave02Ready(Parity|Mutants)$' -count=1 -v`

The overlay adds only this directory's wave02-certification.go.txt as a supplemental test file; it does not replace any shared test, comparison, oracle or registry. It calls the existing shared capture, comparison and native mutant build functions. Parity PASS 113.93s; three-runtime semantic mutants PASS 173.04s. Mutants: allowed restricted type ban reported, AggregateError container reported, ignore duplicate getter exemption. Each compiles and runs cleanly, and output comparison alone catches it on all three runtimes. The shared semantic mutant gate now uses Node/emitted JavaScript with a native canary; the supplemental test supplies the per-rule sanitized native certificate requested for this unit.

The first restricted-types adaptation inferred never[] for an empty conditional branch and failed lowering; explicitly typed edits fixed it. Complete upstream parity then found the new upstream .test.a exemption missing in bare throw; the owned rule was repaired. Those failures are retained. The first complete package run was superseded and its own process tree terminated while fixing the known owned mismatch; interrupted-package.jsonl is not a completed gate and earns no credit.

Deferred ports, with unchanged upstream Go tests passing:

| Rule | Upstream combinations | Exact dependency |
| --- | ---: | --- |
| array-callback-return | 288 | control_flow_graph.Build(node, control_flow_graph.Hooks[struct{}]{}) followed by graph.EndReachable |
| dot-notation | 67 | dotnotation.CompileAllowPattern(resolved.AllowPattern), which calls regexp.Compile(source) |
| id-length | 202 | regexp.Compile(pattern) for each wire.ExceptionPatterns entry |

No CFG approximation or runtime regex private copy is added. No new helper claims. Prior losing rule copies remain assigned to the DEDUP_LEDGER winners, including Google font display to wave1-01; the unit does not register duplicates. The original no-explicit-any and no-inferrable-types choices were already skipped as ported, as their claim file records.

Setup used GOPROXY=https://proxy.golang.org|direct before cloud/setup.sh, which completed in 307.418s. Go 1.27.1, clang 20.1.8, Node 24.19.0, nproc=5, CPU quota=4, 17.6 GB. Every changed Go adapter and supplemental Go test source passed gofmt -l with no output. TypeScript checkout is v6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8.

Complete package command, with all inputs set:

`ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 ADAMIC_LINT_BENCH=1 ADAMIC_LINT_PROFILE_DIR=/workspace/scratch/wave02-final-profile ADAMIC_LINT_PROFILE_SNAPSHOTS=/workspace/scratch/wave02-final-profile go test ./stage1/cohere/lint -timeout 60m -count=1 -json`

Exit 0, PASS 1139.491s. 149 passing tests/subtests, zero failures, one existing skip: TestCheckerBridgeRefusalPending, whose unchanged guard awaits TSGoError in internal/load/prelude.d.ts before it can probe tsgoInspect's C error buffer. No input-dependent test skipped. This is not a certificate for that missing bridge-error behavior, nor a full repository gate. Complete raw events are in package.jsonl, including the intentional negative subprocess failures that the canary tests assert.

After the complete package finished, only unused archived files were removed: three numeric listener modules not imported by the active entries, old private probe scripts, and the unregistered historical font-projection candidate. Active rule code, descriptors, Go adapters, shared modules and registered witnesses are unchanged from the passing run. Registry validation is rerun after that cleanup. Google font display's unchanged Go suite also passes all 16 captured combinations; its DEDUP_LEDGER winner remains wave1-01, so no duplicate port is landed here.
