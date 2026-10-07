# Batch 33 handoff

Built IsLikelyReactComponent, typeReferenceName and isNetworkServiceHookCall, one Adamic .a helper per file. Nine frozen dependency occurrences across seven distinct consumers; none newly helper-ready. Cumulative slot05: 95 helpers, 446 occurrences, 81 consumers, 51 helper-ready candidates including the common 46. These are dependency calculations, not completed lint rule implementations.

- IsLikelyReactComponent: structure/consistency-require-matching-file-name, structure/react-component-no-display-name, structure/react-component-require-named-export.
- typeReferenceName: structure/network-require-hook-options-parameter, structure/network-require-hook-variables-type, structure/react-component-require-properties-type-suffix.
- isNetworkServiceHookCall: structure/network-require-hook-options-parameter, structure/network-require-hook-request-suffix, structure/network-require-hook-variables-type.

Claim f485a99df was pushed successfully before source writes. Selection read every claim across all 20 origin codex/lint-helpers* branches and shared HELPERS.md. Each chosen helper tied the highest remaining unclaimed concrete fan-out of three. Current origin/main 71d7e491b3c9724f7a0e2ee754592149e7f9790b and origin/area/stage1-lint 2fbfe42155387f67f297a41c89d318e3b6cde310 remain unchanged ancestors. Prior 92 helpers, compiler/runtime and shared harness inputs are unchanged, so their retained proof remains applicable. Only the worker branch was pushed; no shared files, rule registry or dispatch was authored.

## Actual Go comparison

The owned suite asserts pinned Go cohere 715ba94f3608a6500086b1076ce5cb7e51b836db. An overlay adds thin wrappers around unchanged original functions. Every string literal in every inventory consumer test file is captured and parsed by pinned typescript-go. Component and type modes query every actual node and nil; network mode queries every actual call expression, respecting the original helper's call-only precondition. Factual boolean classifications and direct child indices contain no predicted answer. The component port composes the actual batch32 returnArgumentLooksLikeJsx implementation against a parallel same-index JSX projection.

| Helper | Captured strings | Sources with controls | Queries | Positive/nonempty Go answers |
|---|---:|---:|---:|---:|
| component | 323 | 331 | 5253 | 93 |
| type name | 200 | 208 | 3226 | 47 |
| network call | 182 | 190 | 60 | 50 |
| Total helper queries | overlapping sources | overlapping sources | 8539 | |

All 8539 baseline answers match byte-for-byte on unchanged Go, source Node, emitted JavaScript on Node and sanitized native, requiring exit zero and empty stderr. Controls cover first/later/destructured parameters, methods/accessors and body absence, block/expression/direct/nested/non-return JSX, immediate and nested ternaries, assertions, plain/qualified/other types, callee and receiver parentheses through depth 64, all four method names, optional calls, element access, case differences and missing/false map entries. Go's mutable method map is an explicit dependency, exported from the oracle process; one synthetic false entry exercises value versus membership. No source file or original helper body is changed by that control.

## Commands and observed output

All test output went directly to named log files.

- ADAMIC_SLOT05_BATCH33_EVIDENCE=<owned evidence directory> ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch33 -count=1 -v -timeout=20m: final PASS, 64.699s, all 8539 queries and 21 native semantic mutants caught. Initial suite PASS, 63.853s, 8523 queries and 20 mutants. A focused non-return JSX expression statement and its exact return-kind mutant justified rerunning the complete suite. Both corpora, verdicts, coverage and logs are retained losslessly.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v: PASS, 1.006s, seven actual input probes, zero cache hits and seven probe misses.
- Final go vet ./...: exit zero, empty output. Final gofmt -l cmd internal stage1/cohere/lint/helpers/slot05: exit zero, empty output.
- export GOPROXY='https://proxy.golang.org|direct'; bash cloud/setup.sh: PASS, 25.570s; nproc=5, quota 400000 100000, 17.6 GB. Node ready 0.025s, Go 0.030s, markdown 0.077s, submodules 0.090s, clang 0.195s, Go build 25.389s, test binaries deferred 25.540s, cache warm 25.542s. Go 1.27.1, clang 20.1.8, Node 24.19.0. source /workspace/adamic-tools/env.sh before builds.

## Every semantic mutant

All 21 compiled and ran native with exit zero and empty stderr before their wrong output was compared to unchanged Go. No compiler error, panic, sanitizer finding or stderr is credited as a semantic kill. Baseline source/emitted comparisons pass; mutants run native only. Exact replacements and first independent witnesses are in evidence/mutants.json and helpers.log.gz.

- Component, ten: accept nil; accept an unsupported function kind; omit properties; omit props; inspect the second parameter; accept missing body; omit direct block return recognition; accept an empty/non-JSX block; omit expression-body recognition; drop the return-statement kind guard. Go first-parameter, body and direct-return controls catch all ten. The final mutant is caught by a function whose body contains a JSX expression statement but no return.
- Type name, four: invent a name for nil; invert the TypeReference gate; invert the Identifier gate; corrupt returned text. Actual named references and nil catch all four.
- Network call, seven: omit callee-parenthesis handling; omit receiver-parenthesis handling; accept a non-property callee; match other instead of networkService; use map membership instead of its boolean value; always accept a method; always refuse a method. Actual Go calls and the false/missing map entries catch all seven.

## Limits

The full repository gate and its 17 required external correctness checks, full shared lint suite, whole-rule diagnostics/fixes/suggestions, independent Adamic AST integration, arbitrary malformed/cyclic projections, invalid non-call network inputs and performance were not run. No skipped check is claimed green. The mutable method table and return predicate are explicit dependencies, with the proved batch32 predicate composed in this driver. Whole-file network collection and component-rule integration remain separate helpers and rule work. No regex is required.

Evidence retains every generated factual case, actual Go answer, consumer file/count coverage, kind counts, all mutant witnesses, initial/final proof, identities and hashes. No prior published evidence was deleted.
