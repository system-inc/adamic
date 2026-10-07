Built: decodeControlEscape, namedBackreference and namedGroupOpener, one .a file per helper; 12 prerequisites removed across four rules.
Commits: control 6d12e3fad68aaaae5fb75fb7e08f8357e94ebe24; final implementation cdc87e1731f919f04a0426d054ef0dc64d319a5f; claims 94e0c37 and 72538c1 pushed before code.
Commands and outputs: helper oracle PASS 124.034s, 2,381,110 queries; vet/types/format PASS; filtered oracle PASS 0.851s; setup 28s, nproc 5.
Mutants: all twelve compiling semantic mutants caught by actual Go output comparison on sanitized native; each witness is below and in evidence/helpers.log.
Not covered: full rule diagnostics, full repository gate, invalid UTF-8 Go strings, arbitrary noninteger arguments, exact panic wording and integration into shared registration.

All previous twenty-six retained helpers were rebased onto current main f8013f0 and re-greened before this batch, then pushed at 8965e62. Main was fetched again after the final implementation commit and remains f8013f0baac41ddc340d76f83bddde38536a8f07, an ancestor of this branch. Only codex/lint-helpers-05 is pushed; integration owns main and area branches.

Each helper serves @next/next/no-html-link-for-pages, @typescript-eslint/no-empty-object-type, no-restricted-exports and no-restricted-imports. Removing these twelve dependency occurrences removes no final listed blocker. CONSUMERS.md and readiness.json list the complete frozen residual dependencies. Cumulative slot 05 delivery is twenty-nine helpers, 211 removed dependency occurrences across 66 consumers and fifty helper-ready rules in that ledger. Helper readiness is not complete rule parity.

The test-only Go overlay calls the real cohere private functions at 715ba94f3608a6500086b1076ce5cb7e51b836db. Every baseline compares actual Go with source Node, emitted JavaScript and ASan/UBSan native. Consumer coverage includes every nonempty string literal from each consuming Go test file, including sources, configurations and expected strings; these captures are not full diagnostic replays. The control corpus has 123,160 queries over 1,072 strings and controls. Each named reader has 1,128,975 queries: 16,911 source suffixes and controls plus every 1,112,064 Unicode scalar as a name. Corpora hashes and per-consumer coverage are archived in evidence.

The control decoder preserves modulo-32 letters, Annex B class-only digits/underscore, Unicode unsupported-syntax errors, all result fields and zero-width backslash fallback. Prefix readers preserve exact anchors, first delimiter, empty names and UTF-8 byte widths; group openers reject both lookbehind prefixes. Nonempty control errors must propagate as invalid patterns. See README.md for contracts and representation limits.

All mutants compile, run normally with exit zero and empty stderr, then differ from real Go. Witness numbers below are output-line indices, not query counts:

| Helper | Mutation | Witness | Mutant / Go |
| --- | --- | --- | --- |
| control | modulo 31 | 225 | 0:0:4:2:false / 0:0:2:2:false |
| control | widen outside classes | 4057 | 0:0:31:2:false / 0:0:92:0:false |
| control | omit Unicode refusal | 5 | 0:0:92:0:false / 0:0:0:0:false |
| control | fallback width one | 1 | 0:0:92:1:false / 0:0:92:0:false |
| backreference | reject empty name | 8907 | false:0 / true:4 |
| backreference | UTF-16 width | 8920 | true:5 / true:6 |
| backreference | last delimiter | 8909 | true:7 / true:5 |
| backreference | omit prefix guard | 133 | true:41 / false:0 |
| group opener | reject empty name | 5277 | false:0 / true:4 |
| group opener | UTF-16 width | 5292 | true:5 / true:6 |
| group opener | last delimiter | 5281 | true:7 / true:5 |
| group opener | omit negative lookbehind exclusion | 5273 | true:6 / false:0 |

Commands ran with /workspace/adamic-tools/env.sh sourced; all test output went directly to log files:

```
bash cloud/setup.sh > /tmp/lint05-batch10-setup.log 2>&1
go test ./stage1/cohere/lint/helpers/slot05/batch10 -count=1 -v -timeout=20m > /tmp/lint05-batch10-helpers.log 2>&1
go vet ./... > /tmp/lint05-batch10-vet.log 2>&1
go run ./cmd/adamic types stage1/cohere/lint/helpers/slot05/batch10/main.a > /tmp/lint05-batch10-types.log 2>&1
go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v > /tmp/lint05-batch10-oracle.log 2>&1
gofmt -l stage1/cohere/lint/helpers/slot05/batch10 > /tmp/lint05-batch10-format.log 2>&1
```

Setup timing: Go ready 0s; clang ready 1s; Node ready 1s; submodules ready 1s; build cache warm 28s; done in 28s on five processors, cpu.max 400000 100000, 17.6 GB. Filtered oracle ran six fixtures with six probe misses and zero hits. Vet, types and formatting logs are empty successful outputs.

Claims were checked across all eighteen origin codex/lint-helpers* branches. Slot 01 claimed boundedQuantifierWidth and groupKindOf eight seconds earlier (9867cb6 at 04:37:46 UTC versus our 94e0c37 at 04:37:54); both were withdrawn before any duplicate code. Retained control was completed and pushed before replacement claims. Replacement claim 72538c1 was pushed before named-reader code. Slot 02's control claim e6aaf5c2 at 04:39:31 was ninety-seven seconds later than ours, and explicitly withdrawn in 0a261996. The final scan's historical text match is documented in evidence/claims.json; active ownership remains here. No fourth retained helper is claimed.
