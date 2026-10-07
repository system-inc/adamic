# Batch 32 handoff

Built returnArgumentLooksLikeJsx, jsxAfterParentheses and isJsxValue, one Adamic .a helper per file. Each removes a recorded prerequisite for structure/consistency-require-matching-file-name, structure/react-component-no-display-name and structure/react-component-require-named-export. Nine occurrences across three distinct rules; none newly helper-ready. Cumulative slot05: 92 helpers, 437 occurrences, 77 consumers, 51 helper-ready candidates including the common 46. These are frozen dependency calculations, not completed rule ports.

Claim 59c6b8d82 was successfully pushed before source writes. Selection read every claim on all 20 origin codex/lint-helpers* branches plus shared HELPERS.md. Each selected helper tied the highest remaining unclaimed concrete fan-out, three. Current origin/main 71d7e491b3c9724f7a0e2ee754592149e7f9790b and origin/area/stage1-lint db2ecc00447f9ebe8adecb190f71ac222e5db860 were unchanged ancestors of the published branch. Prior 89 helpers and compiler/runtime/harness inputs remain unchanged with retained proof. No shared files, rule dispatch or registry was authored; only the worker branch was pushed.

## Observations

Pinned Go cohere 715ba94f3608a6500086b1076ce5cb7e51b836db is asserted by the owned harness. An overlay adds thin private helper exports without modifying any original Go helper. The oracle captures every string literal in every inventory consumer test file: 323 distinct strings plus eight controls, 331 sources. Pinned typescript-go parses them, and every resulting AST node plus nil is queried for each helper. There are 5191 queries per helper, 15573 total. Positive Go verdicts: isJsxValue 100, jsxAfterParentheses 242, returnArgumentLooksLikeJsx 250.

All baseline answers agree byte-for-byte with actual Go on source Node, emitted JavaScript on Node and sanitized native, requiring exit zero and empty stderr. The factual flat projection contains boolean kind classifications and only direct parenthesis/conditional child identities; it does not contain a predicted result. Controls require all JSX kinds, parentheses through depth 128, conditional true/false/condition/nested-conditional cases, assertions and logical/container expressions. All original fixture strings are retained, including parser recovery trees. Nil is represented as -1.

The original helper unwraps parentheses alone. As/satisfies/non-null/type assertions stay in place. Its ternary test asks whether either immediate arm becomes JSX after parentheses; it does not recursively inspect a nested ternary or find JSX in the condition. Those boundaries are preserved. No regex is used.

## Commands and outputs

All test output was written directly to named log files.

- ADAMIC_SLOT05_BATCH32_EVIDENCE=<owned evidence directory> ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch32 -count=1 -v -timeout=20m: final PASS, 31.904s, 15573 baseline queries and all 16 native semantic mutants caught.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v: PASS, 0.908s, seven actual input probes, zero cache hits and seven probe misses.
- Final go vet ./...: exit zero, empty output. Final gofmt -l cmd internal stage1/cohere/lint/helpers/slot05: exit zero, empty output.
- export GOPROXY='https://proxy.golang.org|direct'; bash cloud/setup.sh: PASS, 25.024s; nproc=5, quota 400000 100000, 17.6 GB. Node ready 0.026s, Go 0.029s, markdown 0.082s, submodules 0.097s, clang 0.215s, Go build 24.874s, test binaries deferred 24.995s, cache warm 24.996s. Go 1.27.1, clang 20.1.8, Node 24.19.0. source /workspace/adamic-tools/env.sh before builds.

Initial suite FAIL, 11.575s: Adamic explicitly refused non-null assertions in owned source. Those were replaced locally with ?? panic('missing AST node') for invalid indices. No shared compiler or harness edits. This refusal is retained in initial-refusal.log.gz and receives no mutant credit. The final whole owned suite was rerun after the source correction.

## All mutants

Every final mutant compiles and runs native with exit zero and empty stderr before wrong output is compared against unchanged Go. No compiler refusal, panic, sanitizer finding or stderr is credited as a semantic kill. Baselines cover source and emitted JavaScript; mutants run native only. Exact replacements and independent first witnesses are in evidence/mutants.json and helpers.log.gz.

- isJsxValue, four: omit element, omit fragment, omit self-closing, accept nil. Actual JSX nodes and nil independently catch them.
- jsxAfterParentheses, four: omit unwrapping, unwrap only one layer, test the original node instead of the unwrapped one, accept nil. Parenthesized JSX through multiple layers and nil catch them.
- returnArgumentLooksLikeJsx, eight: accept nil, omit parentheses, omit direct JSX, omit conditional handling, omit the true arm, omit the false arm, require both arms instead of either, accept an unrelated non-JSX node. Go's true-only and false-only ternaries discriminate the arm mutants, and direct/parenthesized JSX and negative node queries catch the rest.

## Limits

The full repository gate, its 17 required external correctness checks, the full shared lint suite, whole-rule diagnostics/fixes/suggestions, Adamic's independent AST parser integration, arbitrary invalid/cyclic projections and performance were not run. No skipped check is claimed green. The post-unwrapping nil guard is retained from Go; malformed parenthesized nil children are outside the valid parser projection contract, and no semantic kill of that redundant guard is claimed. These are helper comparisons on actual Go parser geometry, not proof of whole-rule operation on Adamic's parser.

Evidence retains lossless logs, all generated factual cases and actual Go answers, consumer file/count coverage, kind counts, all mutant witnesses, identities and SHA-256 hashes. No prior published evidence was deleted.
