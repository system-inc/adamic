Built: Canonicalize, CaseClass and writeClass, one .a file per helper; twelve prerequisites removed across four rules.
Commits: retained implementation fd1572584223be52a741d57e218c4ad6ee16ea36; final implementation b340a79f8f9bcf0b98bc0fc735a66cfcae6f5619; claims d05d4f25 and 7eccd64a pushed before their code.
Commands and outputs: full helper oracle PASS 82.262s, 4,476,864 queries; vet/types/format PASS; filtered uncached oracle PASS 0.897s; setup 32s, nproc 5.
Mutants: all twelve compiling semantic variants caught by actual Go output or dependency traces; every exact witness is in evidence/mutants.json and the table below.
Not covered: dependency implementations, arbitrary callbacks or table shapes, full rule diagnostics, full repository gate, shared integration and inputs outside documented adapters.

Landing and claims

All thirty-five prior retained helpers were complete, tested and pushed at 63d32093 on current origin/main c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06. Main was fetched before selection and at final verification, and remains the ancestor. No additional rebase was needed. The user-named harness ab70f38d4 and announced shared leak-check changes are not on this main; these helpers add no finding-model dependency. Only codex/lint-helpers-05 is pushed; integration owns main and area branches.

All twenty origin codex/lint-helpers* branches were fetched and their claims inspected. Comments remain reserved by shared HELPERS.md, and all larger concrete counts are claimed. The retained trio ties the highest available concrete count at four consumers each. Initial claim d05d4f25 was pushed before code. An existing simpleFold claim was missed in the truncated shortlist: wave 1 slot 10's 005c910c at 05:33:46 UTC precedes ours at 05:40:07. That claim was withdrawn before any duplicate code; no simpleFold implementation is delivered or counted. This was a selection error, not a concurrent publication race.

Canonicalize and CaseClass were completed, tested and pushed at fd157258 before choosing the replacement. Their initial four-way gate passed 53.149s over 4,456,896 queries and caught seven mutants. The replacement writeClass claim 7eccd64a was pushed before code. Post-publication and final twenty-branch scans find all three retained symbols only here; evidence/claims.json records the full final scan and earlier ownership. No fourth retained helper is claimed.

Consumers and readiness

Each helper serves @next/next/no-html-link-for-pages, @typescript-eslint/no-empty-object-type, no-restricted-exports and no-restricted-imports. Twelve prerequisite occurrences are removed, with no final listed blocker removed. CONSUMERS.md and readiness.json list every consumer and complete residual dependencies. Cumulative slot 05 delivery is thirty-eight retained helpers and 247 removed prerequisite occurrences across 66 consumers; frozen conditional helper readiness remains fifty. This is not whole-rule parity.

Observed validation

The Go overlay calls the actual private/public functions at cohere 715ba94f3608a6500086b1076ce5cb7e51b836db. It changes only dependency call names inside the tested functions, recording actual call arguments and order while preserving original algorithms and returned values. The production cohere worktree is untouched. Every baseline matches Go against source Node, emitted JavaScript and ASan/UBSan native with exit zero and empty stderr.

Canonicalize passed 2,228,448 queries in 26.01s, preserving Unicode folding, supplementary and full-uppercase-expansion guards, and non-ASCII-to-ASCII refusal. Actual Go supplies sparse standard uppercase and private simpleFold results plus expansion entries; the injected dependencies are explicit and their implementations are not ported or guessed.

CaseClass passed 2,228,448 queries in 26.39s, preserving decline flags, ordered escaped members, exact brackets and dependency calls. Actual Go supplies memberships and each EscapeClassRune string. Both helpers test every integer zero through 0x10ffff in both modes, including surrogate integers, plus 112 signed-rune controls covering int32 extremes, invalid scalars and special case pairs.

writeClass passed 19,968 queries in 29.83s, preserving empty positive/negative special forms, atom order, conditional caseExtras, unchanged Unicode and shared-array delegation, and positive/negative wrappers. Inputs include all 256 uint8 atom kinds, all captured runes, every captured string as set text, empty and mixed lists, and all sixteen rewrite-flag combinations on kind controls. Real Go supplies atom formatting and caseExtras bytes; the test proves their delegation rather than their implementations. Total final queries: 4,476,864. Full gate: PASS 82.262s.

All consuming Go test files supply every nonempty literal, including sources, options and expected strings. There are 904 distinct strings. Per-consumer occurrence counts are 383, 65, 480 and 1,266 in the consumer order above. These captures are helper inputs, not full diagnostic replays. Go's outer case-table order can vary across processes; each corpus preserves its actual runtime order, so archived hashes document observations rather than promise byte-for-byte regeneration.

Mutants

All credited variants compile and run normally before differing from actual Go values or traces. Compiler errors, panic, sanitizer failure and stderr do not count. Every variant is built from temporary copies. Output-line witness indices below are not query indices:

| Helper | Mutation | Witness | Difference |
| --- | --- | ---: | --- |
| Canonicalize | invert Unicode branch | 1 | fold call replaces expand then upper calls |
| Canonicalize | remove supplementary guard | 107 | extra expand/upper calls for 66560 |
| Canonicalize | remove expansion guard | 1 | required expand call absent |
| Canonicalize | remove ASCII-crossing guard | 96 | 83 instead of 383 for long s |
| CaseClass | widen empty decline | 1 | true instead of false |
| CaseClass | omit opening bracket | 38 | Aa] instead of [Aa] |
| CaseClass | omit first member | 38 | [a] instead of [Aa], first escape call missing |
| writeClass | empty negated class matches nothing | 3 | [] instead of [\s\S] |
| writeClass | alter empty positive form | 1 | [] instead of (?!) |
| writeClass | omit caseExtras | 73 | [A] instead of [Aa] |
| writeClass | omit negated prefix | 67 | [A] instead of [^A] |
| writeClass | omit first atom | 65 | [] instead of [A] |

The first replacement run found an unsupported postfix index expression in the test callback. It was rewritten as separate statements within this unit; the final full run above supersedes that failed attempt. No compiler or shared harness was changed.

Commands

All test output went directly to logs. With /workspace/adamic-tools/env.sh sourced:

```
bash cloud/setup.sh > /tmp/lint05-batch13-setup.log 2>&1
ADAMIC_SLOT05_BATCH13_EVIDENCE="$PWD/stage1/cohere/lint/helpers/slot05/batch13/evidence" go test ./stage1/cohere/lint/helpers/slot05/batch13 -count=1 -v -timeout=20m > /tmp/lint05-batch13-helpers.log 2>&1
go vet ./... > /tmp/lint05-batch13-vet.log 2>&1
go run ./cmd/adamic types stage1/cohere/lint/helpers/slot05/batch13/main.a > /tmp/lint05-batch13-types.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v > /tmp/lint05-batch13-oracle.log 2>&1
go vet ./stage1/cohere/lint/helpers/slot05/batch13 > /tmp/lint05-batch13-vet-final.log 2>&1
go run ./cmd/adamic types stage1/cohere/lint/helpers/slot05/batch13/main.a > /tmp/lint05-batch13-types-final.log 2>&1
gofmt -l stage1/cohere/lint/helpers/slot05/batch13 > /tmp/lint05-batch13-format.log 2>&1
```

All final commands exit zero. Vet/format logs are empty; types printed inferred declarations successfully. The filtered uncached oracle passed all six fixtures with zero probe cache hits and six misses. Setup timing: Go ready 0s; clang, Node and submodules ready 0s; cache warm 32s; done in 32s on five processors, cpu.max 400000 100000, 17.6 GB. Toolchain: Go 1.27.1, clang 20.1.8, Node 24.19.0. Retained-only gate and types/vet evidence are also archived.

No matcher, rule listener, rule.json kinds, finding-position conversion, shared registration, finding model, harness or protected compiler edit was added. Not covered: dependency implementations, arbitrary manually supplied callbacks/tables, noninteger or out-of-int32 rune arguments, out-of-uint8 atom kinds, invalid UTF-8, full repository gate or complete findings/fixes/suggestions and integration. Earlier withdrawn overlaps stay withdrawn.
