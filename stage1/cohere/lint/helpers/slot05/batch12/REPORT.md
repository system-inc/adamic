Built: caseTables, CaseEquivalents and CaseEquivalenceGroups, one .a file per helper; twelve prerequisites removed across four rules.
Commits: claim 3d6a92883434f1d0c6569fd61aa081a9fe634776 pushed before code; implementation e811907f2d260fd4d56000e1bda600b1237aabdf.
Commands and outputs: helper oracle PASS 26.288s, 2,228,464 queries; vet/types/format PASS; filtered uncached oracle PASS 0.924s; setup 34s, nproc 5.
Mutants: all nine compiling semantic variants caught by actual Go output, constructor traces or alias identity; every witness is in evidence/mutants.json and below.
Not covered: table-constructor implementations, cold concurrent initialization, full rule diagnostics, full repository gate, shared integration and inputs outside the documented adapters.

Landing and ownership

All thirty-two prior retained helpers were landing-ready and pushed at 396fa729 on current origin/main c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06. Their complete eleven-package landing oracle caught eighty-nine mutants in the previous unit. Current main was fetched before selection and again at final verification; it remains the branch's ancestor. No further rebase was needed. The user-named harness commit ab70f38d47de1d4974082b38f84a56af2368b7af is available from lint-rules/harness but is not yet on main. These helpers add no rule or finding-model dependency. The announced shared leak-check change is not on this main.

All twenty origin codex/lint-helpers* branches were fetched and their claims inspected. The five higher-count comments helpers remain reserved in shared HELPERS.md. Every larger concrete count is claimed; these three tie the highest unclaimed count at four consumers. Claim 3d6a9288 was pushed before any code. Post-publication and final scans find these symbols only here; evidence/claims.json records the final scan. No fourth helper is claimed. Only codex/lint-helpers-05 is pushed; integration owns main and area branches.

Consumers and readiness

Each helper serves @next/next/no-html-link-for-pages, @typescript-eslint/no-empty-object-type, no-restricted-exports and no-restricted-imports. The batch removes twelve prerequisite occurrences, with zero final listed blockers. CONSUMERS.md and readiness.json identify every consumer and complete residual dependencies. Cumulative slot 05 delivery is thirty-five retained helpers and 235 removed dependency occurrences across 66 consumers; the frozen conditional helper-readiness count remains fifty. This does not claim completed rule diagnostics.

Observed behavior

The private Go functions at cohere 715ba94f3608a6500086b1076ce5cb7e51b836db supply all expected results. The overlay replaces only cached-constructor call names with wrappers recording their calls. Original algorithms, constructor caches and returned values remain intact; the production cohere worktree is unchanged. Every baseline compares actual Go with source Node, emitted JavaScript and ASan/UBSan native, requiring exit zero and empty stderr.

caseTables chooses exactly one cached constructor and preserves its member map and groups identities. Cache construction and lifetime remain explicit externally owned dependencies. CaseEquivalents delegates once with the unchanged Unicode flag and returns the shared member group; absent members project Go nil to undefined. CaseEquivalenceGroups delegates once and returns the exact shared groups, preserving their values, runtime order and aliases. Results are immutable by contract. No regex matcher, pattern translation, listener, dispatcher or finding-offset conversion is added.

Go supplied 1,144 uppercase groups and 1,482 Unicode fold groups. Go map iteration chooses outer group order, which is nondeterministic across processes. The adapter preserves each observed runtime order instead of sorting it. Corpus hashes document the observed run; they are not a byte-for-byte regeneration promise across processes. Every membership points to the actual shared group. The test-only Go map alias probe writes a sentinel and restores the old map state; slice alias probes compare backing identities. Adamic reconstructs the supplied arrays and member aliases, then compares exact return values and identities against those observations.

All four consumers' Go test files are parsed for nonempty string literals, including sources, options and expected text. There are 904 distinct strings and 112 signed-rune controls. Per-consumer occurrence counts are 383, 65, 480 and 1,266 in the consumer order above. Controls include signed int32 extremes, absent keys, surrogate integers, supplementary case pairs, Kelvin sign, long s and the three fold-only pairs seeded by Go. Captures are helper inputs, not whole diagnostic replays.

caseTables passed eight repeated mode-switch queries; CaseEquivalenceGroups passed eight complete group-array queries. CaseEquivalents passed 2,228,448 queries, covering every integer zero through 0x10ffff, including surrogates, in both modes plus the controls. Total queries are 2,228,464. Full helper gate PASS 26.288s; member lookup PASS 19.75s and group projection PASS 3.31s.

Mutants

All nine credited variants compiled and ran normally, then differed from actual Go. A panic, sanitizer finding, compiler refusal or stderr does not count. The original copying mutant used optional-call syntax that stage 0 cannot lower; it was not credited. Its replacement uses an explicit undefined guard followed by slice, compiles and produces the intended alias mismatch. The final full batch rerun supersedes that earlier attempt.

| Helper | Mutation | Output-line witness | Difference |
| --- | --- | ---: | --- |
| caseTables | invert Unicode branch | 1 | fold:true:fold; versus upper:true:upper; |
| caseTables | use fold constructor for non-Unicode | 1 | fold:true:fold; versus upper:true:upper; |
| caseTables | eagerly call uppercase constructor | 1 | upper:true:upper;upper; versus upper:true:upper; |
| CaseEquivalents | invert delegated flag | 1 | fold trace instead of upper trace |
| CaseEquivalents | look up next rune | 37 | 65,97:false:true:upper; versus nil:true:false:upper; |
| CaseEquivalents | copy present group | 38 | 65,97:false:false:upper; versus 65,97:true:true:upper; |
| CaseEquivalenceGroups | invert delegated flag | 1 | wrong complete groups and fold trace |
| CaseEquivalenceGroups | copy outer groups array | 1 | same values but shared identity false instead of true |
| CaseEquivalenceGroups | drop first group | 1 | first group missing and shared identity false |

The witness log includes the first differing output byte and a bounded surrounding fragment for large group arrays; full baseline output is generated during each run. evidence/mutants.json records every exact logged mismatch and its log line.

Commands and environment

All test output went directly to log files. With /workspace/adamic-tools/env.sh sourced:

```
bash cloud/setup.sh > /tmp/lint05-batch12-setup.log 2>&1
ADAMIC_SLOT05_BATCH12_EVIDENCE="$PWD/stage1/cohere/lint/helpers/slot05/batch12/evidence" go test ./stage1/cohere/lint/helpers/slot05/batch12 -count=1 -v -timeout=20m > /tmp/lint05-batch12-helpers.log 2>&1
go vet ./... > /tmp/lint05-batch12-vet.log 2>&1
go run ./cmd/adamic types stage1/cohere/lint/helpers/slot05/batch12/main.a > /tmp/lint05-batch12-types.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v > /tmp/lint05-batch12-oracle.log 2>&1
go vet ./stage1/cohere/lint/helpers/slot05/batch12 > /tmp/lint05-batch12-vet-final.log 2>&1
go run ./cmd/adamic types stage1/cohere/lint/helpers/slot05/batch12/main.a > /tmp/lint05-batch12-types-final.log 2>&1
gofmt -l stage1/cohere/lint/helpers/slot05/batch12 > /tmp/lint05-batch12-format.log 2>&1
```

All final commands exited zero. Vet and format logs are empty; types printed inferred declarations successfully. Filtered oracle passed all six fixtures with zero probe cache hits and six misses. Setup timing: Go ready 0s; clang, Node and submodules ready 1s; cache warm 34s; done in 34s on five processors, cpu.max 400000 100000, 17.6 GB. Go 1.27.1, clang 20.1.8, Node 24.19.0.

No shared harness, registration generator, finding model or protected compiler file was edited. Not covered: implementations or concurrent cold behavior of cached constructors; arbitrary manually supplied table shapes; invalid UTF-8, noninteger or out-of-int32 rune adapter arguments; full repository test gate, whole-rule findings/fixes/suggestions and integration. Earlier withdrawn overlaps remain withdrawn.
