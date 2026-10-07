# Original slot 08 rules

Ported no-useless-computed-key, react-hooks/gating and react/forbid-foreign-prop-types in independent directories with .a modules, options, upstream Go adapters and comparison mutants. No shared production files changed.

Observed comparison: source Node, emitted JavaScript, ASan/UBSan native and upstream Go matched 208 supported upstream cases (120 core, 35 gating, 53 foreign) in 69,704 output bytes. Compiler/stage1 comparison matched 228 files and 13,054,023 bytes. All owned witnesses matched 16,816 bytes. This comparison serializes findings, fix ranges/text and applied source through the owned rich transport.

Mutants computed_replacement_wrong, gating_invalid_suppressed and foreign_read_suppressed compiled and ran successfully with empty stderr on each Adamic backend. Only output comparison killed all three on source Node, emitted JavaScript and sanitized native. The core mutant changes only replacement text, preserving diagnostic ranges.

Best of five rotating process timings, TypeScript compiler's 77 files plus 1,000 positive examples (78 files total), release native. Process startup included; no emitted-JavaScript timing requested:

| Rule | Findings | Native findings/s | Node findings/s | Go findings/s |
| --- | ---: | ---: | ---: | ---: |
| no-useless-computed-key | 1001 | 680.17 | 942.03 | 3783.04 |
| react-hooks/gating | 1000 | 669.53 | 887.14 | 3681.98 |
| react/forbid-foreign-prop-types | 1000 | 603.18 | 805.16 | 3482.35 |

Setup reported tools ready 0s, gate warm 37s, total 37s, nproc=5. go vet with overlay passed; uncached TestTheOracleCatchesOneByte passed. Exact commands and output are in evidence logs and validate.py reproduces the scratch overlay using the prior owned compatibility.patch.

Gaps observed: shared module registration still requires rule.ts; its profile and rich fix serialization also require scratch compatibility. The advertised origin/codex/lint-harness-dot-a branch was absent when fetched. Fourteen upstream JSX cases (seven per React rule) cannot be parsed by the shared stage1 parser. A minimal gating JSX witness fails source Node with exit 70, parser slice expected semicolon at 56. Two core malformed-source recovery cases, ({ ['x' }); and ({ ['x': 0 });, are rejected by the Go comparison driver (']' expected), though upstream rule tests exercise recovery. Full upstream comparison therefore fails explicitly; these cases were not silently rewritten. Default integration and the full repository gate are not claimed. All other behavior was ported and compared, and the blockers remain outside owned directories.
