Finished the existing SkipPatternEscape, ClassEnd and isReactComponentBaseName reservations in separate .a files; no additional claim.
Reservation 0041ec9b1dddd02c4f43219a5e52ef6d230e31b4 was pushed before any implementation; publication SHA follows in final response.
Batch26 PASS 107.654s, seven uncached Node probes PASS 7.155s, vet/format clean; setup 238.972s, nproc 5.
All thirteen compiling native semantic mutants caught by actual Go output mismatches; exact changes and first witnesses in evidence/mutants.json.
Not covered: full repository gate, seventeen required stage 1 external-input checks, whole-rule findings, UTF-8 decoder implementation or negative offsets.

## Recovery and landing

The container root filesystem was 100% full. Inspection used df -h /workspace /tmp and du -sh /workspace/* /tmp/* 2>/dev/null | sort -h | tail -20. Go reported its build cache at /home/agent/.cache/go-build. As explicitly authorized, go clean -cache -testcache reclaimed the rebuildable cache, recovering 30 GB and reducing root usage to 2%. No scratch tree, toolchain, pushed branch or evidence was deleted. After rebuilding and validation, root usage was 11%, with 27 GB available. The earlier directory-creation ENOSPC occurred before source and is not treated as a code or regex-runtime failure.

Fetched main 71d7e491b3c9724f7a0e2ee754592149e7f9790b and area e667e3e1dbdfd1b9125c3256961bfbc8ec31946b. Both remain ancestors of the owned branch; taking area reported already up to date. Final publication rechecks both bases. Only codex/lint-helpers-05 is pushed; no main/area push, history rewrite or PR.

## Contracts and oracle

skip_pattern_escape.a ports the actual Go escape width dispatch: x with two hex digits, fixed u with four hex digits, u brace form under u/v, c with optional following byte, p/P/q braces under u/v, and generic UTF-8 rune width. Missing brace closers recover with two bytes. A trailing escape fails with zero width. class_end.a validates its opening bracket, handles leading caret, delegates escape stepping and balances nested brackets only under v. Successful end is exclusive; failures preserve original start. is_react_component_base_name.a accepts exactly Component and PureComponent, byte-for-byte and case-sensitive. API domain and dependency assumptions are in README.md.

The isolated Go overlay adds an executable and a one-line exported wrapper around the private React name function; the original Go helper bodies and source worktree are unchanged. Pinned cohere 715ba94f3608a6500086b1076ce5cb7e51b836db supplies every verdict. Every inventory consumer's Go test string literals are captured, including consumers outside the frozen blocked cohort. Each mode is generated from its exact dependency list. Captured sources are helper projections, including prose and full snippets, rather than whole-rule finding executions.

Escape mode: 2206 distinct captured strings from eight consumers; 48476 baseline calls. Class mode: 2063 captured strings from seven consumers; 46592 calls. React mode: 1609 captured strings from nine consumers; 2661 calls. Total 97729 baseline calls. Per-mode coverage JSON names every consumer and source file. Controls include all 256 bytes alone and inside escape/class/name shapes, escaped closing brackets, nested classes, q/u/p/P brace escapes, malformed hex, Unicode and invalid UTF-8. Scanner calls cover every backslash and bracket in each captured string, zero, EOF and past EOF, with all four flag-field combinations. The oracle provides actual utf8.DecodeRuneInString widths at every source byte, including continuation-byte offsets. It does not replace or reimplement that dependency in the helper source.

Each same-source baseline agrees byte-for-byte with actual Go on source Node, emitted-JavaScript Node and ASan/UBSan native. Every variant must compile, exit zero and produce empty stderr before an output difference earns semantic credit. Compile errors, panics and sanitizer-only failures do not count. Six escape variants change x width, Unicode gating, fixed-u width, c width, q recognition and generic rune width. Four class variants enable nesting without v, disable v nesting, discard nesting depth and make the returned end inclusive. Three React variants remove Component, remove PureComponent or replace OR with AND. All thirteen are caught; exact witnesses and replacements are retained in the JSON and lossless raw log gzip.

## Readiness

Each regex helper removes a prerequisite for no-control-regex, no-regex-spaces and no-useless-escape. The React helper removes one for react/no-access-state-in-setstate, react/no-set-state and react/no-unused-state. Nine dependency occurrences across six unique blocked consumers; no final blocker is removed by this trio alone. Cumulative owned readiness is 74 helpers, 383 prerequisite occurrences, 76 unique consumers and 51 helper-ready rules under frozen common AST adapter assumptions. Combined with the previous 71 helpers, this trio removes the final listed helper blocker for no-useless-escape, adding one helper-ready candidate. This is a frozen dependency calculation, not observed rule findings parity. No rule is declared ported. All existing reservations are complete after this batch.

## Validation commands and setup timing

Every build shell sources /workspace/adamic-tools/env.sh and exports GOPROXY='https://proxy.golang.org|direct'. Test output is written directly to logs, never piped.

```
go clean -cache -testcache
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/lint05-batch26-recovery-setup.log 2>&1
ADAMIC_SLOT05_BATCH26_EVIDENCE=/workspace/adamic/stage1/cohere/lint/helpers/slot05/batch26/evidence ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch26 -count=1 -v -timeout=20m > /tmp/lint05-batch26-helpers.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v > /tmp/lint05-batch26-node.log 2>&1
go vet ./... > /tmp/lint05-batch26-vet.log 2>&1
gofmt -l cmd internal stage1/cohere/lint/helpers/slot05 > /tmp/lint05-batch26-format.log
```

Go 1.27.1, clang 20.1.8, Node 24.19.0. Setup lines: Node ready 0.030s, Go ready 0.031s, submodules 0.136s, markdown validated 0.148s, clang 0.345s, Go build ready 238.591s, test-binary warming deferred 238.871s, cache warm 238.874s, done 238.972s. nproc 5, cgroup quota four cores, 17.6 GB. Deferred setup warming is not a skipped correctness check. Exact lines and environment build flags are in evidence/setup.log. Seven Node probes ran uncached, seven misses and zero hits. Vet and format logs are empty. All selected checks passed and none skipped.

Full repository gate and seventeen required external-input correctness checks were not run, relaxed or credited. This is the authorized bounded worker gate over the completed package plus an external Node oracle. Prior 71 helpers retain their unchanged source and prior proofs; no new complete all-package rerun is claimed. No regex engine, AST adapter, rule listener or finding integration was built. No regex-runtime compilation blocker was encountered because these helpers introduce no regex patterns.

## Sharded lint-area landing

See LANDING27.md for area 7076b4eb, unchanged helper/compiler inputs, thirteen fresh semantic mutant witnesses and required TypeScript input recovery followed by passing shard parity. No new reservation.
