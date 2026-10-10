Built the nullable contract doc, census candidate map, source-site evidence and three reduced lowering fixtures toward Outcome 53.
Delivery commits are reported in the final response; base is origin/main 79f2067b, inspection base 59edda91.
Focused lowering tests pass in 0.202s; IR reader guard passes in 34.373s; map and Node checks pass.
Five mutants fail: changed census count, null replaced by undefined, two generic function-value input mutations, and erased string brand.
Not covered: exhaustive causal replay, exhaustive per-type inventory, generic field red tests, native operation implementation or runtime oracle/count refresh.

The map reports candidates, not guaranteed closures: (a) 88 full / 69 latent, (b) 30 / 9, (c) 44 / 4, (d) 621 / 142. Full and latent counts are separate. The census has U-return 12 full / 11 latent. `source-sites.json` retains every mapped occurrence and containing unit from the pinned full and latent streams. The map excludes nullable-looking soundness refusals and scalar optional work from these totals.

Current main already lowers the two direct generic reductions. Both remain positive controls instead of false historical NotYet pins. The branded string fixture pins exactly `testdata/nullable_scout/branded-string.a:3:10: stage 0 can't lower a function returning __String | undefined yet`. Each fixture begins with the pinned upstream source file and line. No code was copied from cohere. The reductions were written from inspected upstream TypeScript contracts, without copying their implementations.

No production lowering or backend file changed. Inspected paths include the general optional reference union, boxed null/string and mixed-nullish unions, RegExp result arrays and properties, typeof slot-presence metadata, the narrow generic object-intersection proof, Weak handles and JSON's missing string result. Date on this main only has the fs host numeric timestamp/getTime/valueOf path; the proposed Date string-or-null special case was not found. The map deliberately labels its inventory as inspected paths rather than claiming completeness.

Commands run, with output in this directory:

- `export GOPROXY='https://proxy.golang.org|direct'; timeout 600 bash cloud/setup.sh`: exit 124 during build-cache warming, with no setup error text. Printed timing lines: Go 0.098s, Node 0.098s, clang 0.747s, Markdown dependencies 2.048s, submodules 37.871s, shared cache 50.648s. No build-cache-ready or done timing was printed. `nproc` is 5; cpu.max is 400000 100000.
- `source /workspace/adamic-tools/env.sh`: the setup-written environment file was available. Two compiler builds with the shared-cache hook reached hard limits of 90s and 300s. `unset GOCACHEPROG; timeout 300 go build -o /tmp/nullable-adamic ./cmd/adamic` then exited 0. Empty build logs are retained and the timeout outcomes are recorded here.
- `timeout 120 go test -count=1 -v -timeout 90s ./internal/lower -run '^TestNullableScout'`: restored final run exits 0, package 0.202s. Leaves: generic value 0.14s, branded string 0.16s, generic return 0.17s. The initial focused run also passed, package 0.353s.
- `timeout 180 go test -count=1 -timeout 90s ./internal/ir -run '^TestCallTargetReaders$'`: exit 0, 34.373s.
- `timeout 30 python3 review/compiler/nullable-scout/verify-map.py`: exit 0, all selected full/latent reason counts match the pinned census. Changing one input count exits 1 at the independent census comparison.
- `timeout 30 node review/compiler/nullable-scout/node-empty.cjs`: exit 0, `empty cases match Node`. The input mutant replaces the null case with undefined and exits 1 at the typeof assertion. This observes Node, not generated Adamic output.
- Each source mutant runs only its named top-level `TestNullableScout...` with `-count=1 -v -timeout 90s`, then restores the original fixture. All exit 1 at their test assertions, without reaching clang. Generic return 0.37s and generic value 0.12s reject a generic function value; branded string 0.19s rejects the expected NotYet because lowering now succeeds. Exact patches and logs are retained.
- `git diff --check`: exit 0.

No oracle fixture was registered. These compile-time probes therefore have no allocation-count rows; running the complete counts refresh would measure unrelated runtime fixtures and cannot record a NotYet program. No whole-package test run or full gate was run. Integration lane output is recorded separately after the commit.

Recommendation: keep generic returns before generic values/fields, but split concrete binder recovery from representation. Then prove branded strings without accepting arbitrary primitive/object intersections. Hold shared empty-case conversions to Node before replacing the RegExp-specific paths. Date requires its own library/host scope clarification on this main. Keep scalar Maybe work, unknown/object slot tags, record-field descriptors and overload soundness in separate pieces.

Conservative assumptions: keyword-matching census reasons are a review queue, not a causal proof; unconstrained T needs per-instantiation evidence. The unresolved reviewer choice is whether to retain the two now-positive reductions or require different red witnesses from the larger tsc sites. This delivery retains honest positive controls and reports that the requested three NotYet pins were not achieved.
