Step 30: correct the Source representation expectation after constraint-backed TNode storage.
Branch: compiler/hidden-boundaries, based on c4e9fb4a; delivery SHA is reported after push.
Checks: representation, TNode and overload controls pass with -timeout 90s; Node-held sanitized oracle controls pass.
Mutants: all 36 witnesses caught, covering 32 existing overload/storage witnesses, two additional TNode witnesses and two Source storage mutations.
Scope: expectation only; no storage change, new fixture, counts change, census replay or full gate.

The test introduced by 963e7c53 predates hidden-06's 702c3ecb constraint-backed storage. Its old assertion equated no single subtype header with no supported storage. The later generic_constraint_storage.go rule, extracted in dacbdb8c, holds structurally constrained object subtypes as ir.Union so their actual brands survive. Source extends CompilerType admits Object and Closure. Both facts remain asserted: the checker subtype relation and their distinct concrete representations. Source now expects tagged Union rather than refusal or an Object header. A length-only constraint admitting Array and String remains unsupported; its refusal expectation stays intact.

Before the edit, the requested test fails with source: representation = (10, true), want (0, false). After the edit it passes. hidden_boundary_generic_tnode.a and hidden_boundary_generic_tnode_constraints.a agree with source Node and generated JavaScript, release native and sanitized native, with leak checks. The constraints witness preserves object identity, subtype fields and string/array specializations, and reads a field through unmapped constraint storage. Its stdout is true:7:node, 3:2, node on separate lines. Existing overload values and visitors controls also pass uncached.

Commands:
- go test ./internal/lower -run '^TestRepresentationClockSourceCheckedTypes$' -count=1 -timeout 90s -v (red before edit).
- go test ./internal/lower -run '^TestRepresentationClockSourceCheckedTypes$|^TestHiddenTNode|^TestOverload' -count=1 -timeout 90s -v (green).
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^hidden_boundary_generic_tnode(_constraints)?[.]a$|^TestOverloadValues$|^TestOverloadVisitors$' -count=1 -timeout 90s -v (green).
- Existing eight overload mutant scripts, output redirected to scratch; each Go test uses -timeout 90s. Field group includes six constraint-storage witnesses. Two additional TNode overlays return default node and accept a generic function value. Two Source overlays choose Object or remove tagged storage: each fails this updated test at its Source assertion. No compilation failure counts as a caught mutant.

Setup succeeded with GOPROXY=https://proxy.golang.org|direct. Cumulative timing lines: Go 0.021s, Node 0.021s, submodules 0.057s, markdown dependencies 0.072s, clang 0.165s, Go build 34.983s, cache 35.144s, done 35.172s. nproc=5, CPU quota=4. Environment sourced from /workspace/adamic-tools/env.sh. Go overlays use GOTMPDIR=/workspace/scratch/hidden-go-build. Logs and per-mutant commands are preserved beside this report. No cohere code was copied.

All eight existing overload mutant groups returned exit 0; every individual mutant test returned exit 1 with its intended assertion, runtime check or Node disagreement. The legacy trust-visitor-input mutant guards its diagnostic path; the visitor invocation mutants independently prove runtime rejection. witnesses.json records all 36 caught results. The final exact requested representation test passes in 0.040s with -timeout 90s.
