Built CFG Build and rebased all 66 retained helpers; two duplicate claims withdrawn.
Implementation tip b9965e36; rebased onto area d3a37422 containing main b6b1538b.
All 24 helper packages, shared harness, vet, uncached Node oracle and native runtime checks PASS.
Eight new compiling Build mutants and four omissions caught; all retained mutant checks passed again.
Full CFG and whole rule findings are outside Build's callback contract; full repository gate and 17 broader comparisons not selected.

# Landing evidence

75 commits rebased without conflicts onto origin/area/stage1-lint. Incoming typeof-null and lookup-presence behavior, shared finding model, registry, option guard and allocator changes are retained. Protected compiler files and lint.ts have no diff from the integration base. No new replacement helper is claimed while landing this unit. Only codex/lint-helpers-04 is pushed.

Environment /workspace/adamic-tools/env.sh was sourced. Setup timing: Go 0s, clang 1s, Node 1s, submodules 1s, warm cache 79s, total 79s; nproc 5. Every command wrote directly to its raw log.

- go run ./cmd/lint-registry: PASS, 40 descriptors, registry.log.
- go test -p=2 -count=1 -v -timeout=20m ./stage1/cohere/lint/helpers/...: PASS all 24 packages, helpers.log. Build package 94.156s. Package timings are in verification.json.
- go test -count=1 -v -timeout=20m ./stage1/cohere/lint -run '^(TestRulesAgree|TestDotARename|TestCompleteSuggestionSerialization|TestEmittedJavaScriptMismatch)$': PASS 364.674s, harness.log. Integrated comparison agrees on 13053452 bytes. Expected child mutant FAIL is caught by the passing parent. .a rename and complete suggestion serialization pass.
- go vet ./stage1/cohere/lint/helpers/...: PASS empty vet.log.
- ADAMIC_GATE_UNCACHED=1 go test -count=1 -v -timeout=10m ./internal/oracle -run '^TestRuntimeLastIndexOfMatchesNode$': PASS 28.285s; native misses 3, Node misses 2, cache hits 0.
- go test -count=1 -v -timeout=10m ./internal/native -run '^(TestRuntimeReleasePaths|TestRuntimeStringEquality)$': PASS 17.094s.

MUTANTS.md records every named mutant subtest and logged consumer omission from this fresh run. Build's eight mutants change incoming/reachable entry flags, final marking, signature type traversal, block-body selection, and source/static/property dispatch. All run successfully in source Node, emitted JavaScript and sanitized native and are caught through ordinary Go comparison. Removing each of its four consumer corpora fails the coverage check. Withdrawn expr/patternBind preliminary checks receive no credit.

Build removes a prerequisite for array-callback-return, consistent-return, no-unreachable-loop and react-hooks/rules-of-hooks. Other blockers remain. See ../slot04_wave23/REPORT.md and readiness.json for callback seams, 2119 actual upstream asserted inputs and 11 controls. This does not claim full graph construction or native rule findings. Existing helper limits remain unchanged. No selected test skipped. Full repository gate and the 17 broader TypeScript/postcss/graphql/parser comparisons were not selected or bypassed. No shared registration, harness or compiler edits were made.
