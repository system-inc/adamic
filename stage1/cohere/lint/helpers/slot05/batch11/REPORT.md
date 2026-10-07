Built: classAtom.write, wordCharacters and expandsOnUppercase, one .a file per helper; twelve dependency entries removed across four rules.
Commits: claim 3d6341d2d223b80b962422d77d5010fa851f5a29 pushed before code; implementation a7356ae44862be46e6c57e3dbb8989418b402108.
Commands and outputs: helper oracle PASS 39.883s, 1,148,449 queries; vet/types/format PASS; filtered uncached oracle PASS 1.034s; setup 29s, nproc 5.
Mutants: eight compiling semantic mutants caught by actual Go result and callback-trace comparison on sanitized native; every witness is below and in evidence/helpers.log.
Not covered: full rule diagnostics, full repository gate, dependency implementations, shared integration, invalid UTF-8, noninteger adapter values and integers outside declared kind/rune ranges.

Landing and ownership

Current origin/main f8013f0baac41ddc340d76f83bddde38536a8f07 remains an ancestor of this branch. All twenty-nine prior retained helpers were complete and pushed at 45136a1; the prior landing gate and batch10 gate ran on this same main. No rebase was needed because main did not advance. The expected shared leak-check changes have not landed on this main. No shared or protected compiler file was edited. Only codex/lint-helpers-05 is pushed; integration owns main and area branches.

Fetched every origin codex/lint-helpers* branch and checked all claims before selection, after claim publication and at final verification. All twenty branches were scanned. The five higher-count comments symbols remain reserved by shared HELPERS.md. Every larger concrete count is claimed. These three tie the highest unclaimed concrete count at four consumers. Both post-publication scans find them only here; evidence/claims.json records the final scan. No fourth helper is claimed.

Consumers and readiness

Each helper serves @next/next/no-html-link-for-pages, @typescript-eslint/no-empty-object-type, no-restricted-exports and no-restricted-imports. Twelve prerequisite occurrences are removed; none is a final listed blocker. CONSUMERS.md and readiness.json list all consumers and residual dependencies. Cumulative slot 05 delivery is thirty-two helpers and 223 removed dependency occurrences across 66 consumers; the frozen ledger remains at fifty helper-ready rules. This is dependency readiness, not complete rule parity.

Observed comparisons

The oracle calls the real private Go functions at cohere 715ba94f3608a6500086b1076ce5cb7e51b836db. A temporary overlay replaces only dependency call names with tracing wrappers, preserving the original algorithms and all returned bytes. The production cohere worktree is unchanged. Every baseline matches Go against unchanged source Node, emitted JavaScript and ASan/UBSan native with exit zero and empty stderr.

classAtom.write passed 34,191 queries in 32.22s. Inputs include all 256 uint8 kinds, every captured rune, signed int32 extremes, invalid scalar and surrogate values, boundary neighbors, and every captured string as set text. Go supplies literalRune results; the trace proves when and in what order this helper calls that dependency. This gate does not port literalRune.

wordCharacters passed all sixteen rewrite-flag combinations in 2.58s. The Go trace proves one wordClassAtoms call with unchanged flags, followed by one writeClass call with the same array, false negation and all false flags. Adamic proxies return the actual Go dependency values. This proves delegation rather than the implementations of wordClassAtoms and writeClass.

expandsOnUppercase passed 1,114,242 queries in 5.07s, including every integer zero through 0x10ffff, surrogate integers, signed int32 extremes, Greek boundary neighbors and all consumer runes. No regex matcher, rule dispatcher, listener or finding-position conversion is added.

Every consumer Go test file is parsed and all nonempty string literals captured, including source, options and expected text. The corpus contains 911 distinct captured/control strings. Per-consumer occurrence counts are 383, 65, 480 and 1,266 in the order of consumers above. Hashes and coverage are in evidence. Captures are not full diagnostic replays. For wordCharacters, flags rather than the source string determine the helper inputs.

Mutants

All eight variants compiled, ran normally and differed from real Go output. A compiler error, panic, sanitizer failure or stderr is not credited as a semantic mutant caught. Witness indices are output lines:

| Helper | Mutation | Witness | Difference |
| --- | --- | --- | --- |
| classAtom.write | set kind 2 becomes 3 | 5 | mutant formats lo; Go returns empty set text |
| classAtom.write | range hi becomes lo | 3 | mutant repeats -2147483648; Go upper endpoint is -1 |
| classAtom.write | range dash becomes colon | 3 | mutant range uses colon; Go uses dash |
| wordCharacters | clear input flags | 4 | mutant dotAll false; Go dotAll true in atoms call |
| wordCharacters | negate writeClass | 2 | mutant negation true; Go false |
| wordCharacters | set output Unicode flag | 2 | mutant Unicode true; Go false in write call |
| expandsOnUppercase | exclude first range upper endpoint | 101 | mutant false; Go true |
| expandsOnUppercase | move final singleton to adjacent rune | 122 | mutant false; Go true |

Commands

All test output was written directly to logs, never piped. With /workspace/adamic-tools/env.sh sourced:

```
bash cloud/setup.sh > /tmp/lint05-continuation-setup.log 2>&1
ADAMIC_SLOT05_BATCH11_EVIDENCE="$PWD/stage1/cohere/lint/helpers/slot05/batch11/evidence" go test ./stage1/cohere/lint/helpers/slot05/batch11 -count=1 -v -timeout=20m > /tmp/lint05-batch11-helpers.log 2>&1
go vet ./... > /tmp/lint05-batch11-vet.log 2>&1
go run ./cmd/adamic types stage1/cohere/lint/helpers/slot05/batch11/main.a > /tmp/lint05-batch11-types.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v > /tmp/lint05-batch11-oracle.log 2>&1
go vet ./stage1/cohere/lint/helpers/slot05/batch11 > /tmp/lint05-batch11-vet-final.log 2>&1
gofmt -l stage1/cohere/lint/helpers/slot05/batch11 > /tmp/lint05-batch11-format.log 2>&1
```

Setup timing: Go, clang, Node and submodules ready at 0s; build cache warm at 29s; setup done in 29s on five processors, cpu.max 400000 100000, 17.6 GB. Go 1.27.1, clang 20.1.8, Node 24.19.0. Repository-wide vet and final package vet have empty successful logs. Types printed the inferred declarations and succeeded; formatting printed no paths. The filtered oracle passed all six input fixtures, with zero cache hits and six probe misses.

Earlier local attempts failed on non-erased type imports, boolean console arguments and an overly broad nested dependency trace. Those were corrected within this unit. The final full package rerun above supersedes them; no shared harness was changed.
