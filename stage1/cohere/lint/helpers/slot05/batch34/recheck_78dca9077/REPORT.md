# Reserved Compile boundary recheck

Fetched origin again. Main remains 48c05d091f0a43c31cbe051b1d6578d99eeedf19 and area/stage1-lint remains 65017b318da1995237ff3ea2c80f59b055b39ac3. Both are ancestors of the published implementation and landing-validation commit 78dca9077ba5f17dac838b6b045cedf69867813f. No integration code changed, so the prior 95-helper, 34-package and 467-mutant validation still applies to that exact code. No new implementation or claim is made.

Exported GOPROXY=https://proxy.golang.org|direct before bash cloud/setup.sh. Setup succeeded in 25.728s with nproc 5. Cumulative timing lines: Go 0.022s, Node 0.022s, submodules 0.067s, markdown dependencies 0.085s, clang 0.188s, Go build 25.537s, test binaries deferred 25.698s, cache warm 25.700s. Sourced /workspace/adamic-tools/env.sh.

Ran ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch34 -count=1 -v with output written directly to a log. PASS, 0.152s. Actual upstream Go Compile and source Node both print true for pattern a against subject a. Adamic still reports at dynamic_pattern.a:3:27: stage 0 can't lower RegExp with a nonconstant pattern yet. The test asserts that specific typed refusal; its green result does not mean Compile is implemented. Retained logs are gzip-compressed losslessly.

The existing reproducer reads the pattern from programArguments and calls new RegExp(pattern, 'u'). Neither lowered backend can produce an artifact because lowering refuses first. No hand-written matcher, callback substitution for compilation, shared harness change or compiler edit was made. Compile remains blocked for @next/next/no-html-link-for-pages, @typescript-eslint/no-empty-object-type, no-restricted-exports and no-restricted-imports. No new rules are unblocked, and no new semantic mutant exists for this undelivered helper. Stopping on this exact compiler gap as instructed.

The complete repository gate and 17 required external correctness checks were not rerun. No check was skipped, relaxed or removed; previous pushed evidence remains intact.
