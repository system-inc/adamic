# Wave 1 slot 05 second continuation

Built three .a TypeScript syntax-rule candidates with complete suggestion edit lists and directory registration descriptors.
Claim 4a310175 pushed before code; rule commits 40df871e, 572e2567 and 6bcc40e1; prior work through 4570161b was pushed first.
Final four-way comparison PASS: 100 upstream cases without exclusions, 222 compiler/stage1 files (14,365,214 bytes), owned witnesses (68,548 bytes).
Parenthesized-chain omission, postfix-assignment suggestion and compound-this-alias omission each compiled, ran cleanly and failed output comparison on all three runtimes.
Default .a discovery, shared profile compilation and general suggestion protocol integration remain blocked; arbitrary configuration-decoder acceptance and the full repository gate were not covered.

## Selection

Fetched all 320 origin refs after pushing prior work. Scanned every Markdown file recursively under stage1/cohere/lint/claims/ on every origin branch: 39 distinct blobs. Checked main's flat modules and rule descriptors. All 46 helper-ready rules were claimed. The first three eligible syntax-only inventory entries were:

1. @typescript-eslint/no-non-null-asserted-optional-chain
2. @typescript-eslint/no-non-null-assertion
3. @typescript-eslint/no-this-alias

The inventory syntax queue includes syntax-ready and syntax-waiting-on-helpers, excluding checker/type-information/binding rules. There were 205 eligible syntax entries at selection. selection.json records all origin-ref hashes used for this claim, origin/main ef3d907ecdc4c771b016f7d9c52372def057a340, inventory 73ac2eb0963e1a4166eaa0fbd160203f11dcdbdf and the pushed claim commit. No new rule is reported merely from inventory status.

## Implementations and evidence

Optional-chain assertions preserve chain flags, outermost/root distinctions, parentheses, computed/argument positions and the independent removal suggestion. Non-null assertions report every expression, preserving assignment-target recursion and eligibility of the optional-chain suggestion; dot access carries two edits, with raw Go dot-search behavior retained even inside a comment. This aliases preserve declaration and assignment distinctions, target wrappers without initializer unwrapping, every assignment operator, filename gates and typed decoded options.

The unchanged Go rule bodies and policy catalog are pinned in cohere at 715ba94f3608a6500086b1076ce5cb7e51b836db. The exact-name upstream capture contains 24 optional-chain cases, 36 non-null cases and 40 this-alias cases. Zero cases were excluded. The final capture included 554 cases across all discovered rules; only these 100 new-rule cases enter TestWave05SecondCorpus. final.log records 67,426 identical bytes for that manifest.

The compiler and stage1 run includes all discovered rules, all 77 TypeScript src/compiler files and stage1 .ts/.a sources: 222 files total. Tests compare rendered messages, diagnostic ids/ranges, suggestion ids/text and every ordered suggestion edit, plus complete fixed output. These three rules offer no automatic fixes, and upstream fixed source remains unchanged. Additional raw witnesses cover supplementary Unicode, non-ASCII identifiers, commented dots, further optional-chain roots, doubled assertions, argument/index assertions and type/parenthesis wrappers. Every witness has an upstream finding. Source Node, emitted JavaScript and ASan/UBSan native must all exit successfully with empty stderr before their output can be credited.

Cached byte offsets avoid scanning source prefixes for every suggestion. The complete comparisons and all three mutants were rerun after this change. final.log passes in 99.267s: compiler/stage1 63.86s, witnesses 14.88s, upstream 20.52s. final-mutants.log passes in 42.091s: 13.55s, 13.28s and 13.03s. Registry tests and vet pass; the filtered uncached TestTheOracleCatchesOneByte control also passes.

## Integration limits and reproduction

No shared source was edited. The existing directory generator requires rule.ts and fails on the earlier .a candidate, recorded in default-registry.log. The helper foundation's profile_test.go still ranges over portFiles as a variable, although directory registration made it a function. cloud/setup.sh fails during cache warming on that compile error. Both are addressed only in the scratch overlay.

The inherited oracle serializer panics when a suggestion edits a range different from the finding or carries more than one edit. The own rule modules store ordered `byte-start:byte-end:replacement` records separated by `|`, with repair `suggestion-edits:<upstream suggestion id>`. validate.py independently extends the scratch Go serializer to emit actual upstream suggestion ids and edit records. The existing main serializer can print those fields without a source change; its automatic fixer ignores suggestions. The replacements here are only empty strings and `?.`, so the delimiter cannot collide. This records every range/text instead of normalizing several edits into one replacement or deleting suggestions. Existing single-edit findings retain their old protocol.

compatibility.patch is the unchanged scratch .a/harness migration from the prior continuation, originally published by origin/codex/lint-wave1-12. validate.py additionally retains captured filename extensions and TSX parsing in the Go driver. It changes no upstream rule body. These remain comparison-validated candidates pending shared infrastructure integration, rather than automatically integrated ports.

From the repository root, source /workspace/adamic-tools/env.sh. Obtain TypeScript v6.0.3 at commit 050880ce59e30b356b686bd3144efe24f875ebc8. Run:

```sh
python3 stage1/cohere/lint/claims/wave1-05-second-evidence/validate.py --scratch /tmp/lint-wave1-05-second-repro
export ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint-wave1-05-typescript
go test -overlay=/tmp/lint-wave1-05-second-repro/overlay.json ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestWave05SecondCorpus|TestCompilerAndStage1Agree)$' -count=1 -v -timeout=15m > /tmp/wave05-second-final.log 2>&1
go test -overlay=/tmp/lint-wave1-05-second-repro/overlay.json ./stage1/cohere/lint -run '^TestMutants$/(parenthesized_chain_missed|postfix_assignment_suggestion|compound_this_alias_missed)$' -count=1 -v -timeout=10m > /tmp/wave05-second-mutants.log 2>&1
go test -overlay=/tmp/lint-wave1-05-second-repro/overlay.json ./stage1/cohere/lint -run '^TestWave05SecondThroughput$' -count=1 -v -timeout=10m > /tmp/wave05-second-throughput.log 2>&1
go test -overlay=/tmp/lint-wave1-05-second-repro/overlay.json ./stage1/cohere/lint/registry -count=1 -v > /tmp/wave05-second-registry.log 2>&1
go vet -overlay=/tmp/lint-wave1-05-second-repro/overlay.json ./stage1/cohere/lint/... > /tmp/wave05-second-vet.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v > /tmp/wave05-second-oracle-control.log 2>&1
```

## Setup

nproc is 5. Continuation cloud/setup.sh reported Go, clang, Node and submodules ready in 0s each, then failed cache warming at profile_test.go:32:23: cannot range over portFiles (function). No cache-warm or completion timing was printed. Setup failure was worked around by sourcing the installed toolchain and using the isolated overlay for focused tests. Initial setup timing is retained in the prior continuation report. No full repository test or repository-wide vet is claimed.

## Throughput

Final measurement is recorded in final-throughput.log. End-to-end driver time includes startup, parsing and counting, excluding compilation. Best of three interleaved Go/native/Node rounds, 77 compiler files plus a 1,000-statement stress file for each rule. Counts match in every round. Native correctness is sanitized; native throughput is unsanitized. This-alias has zero compiler findings, so the stress file gives a meaningful nonzero count. Compiler-only findings are seven optional-chain assertions, 1,123 non-null assertions and zero this aliases. These are observations for the stated corpus, not general throughput promises.

| Rule | Findings | native findings/s | Node findings/s | Go findings/s |
| --- | ---: | ---: | ---: | ---: |
| no-non-null-asserted-optional-chain | 1,007 | 857.46 | 1,142.65 | 5,587.91 |
| no-non-null-assertion | 2,123 | 1,791.50 | 2,289.66 | 12,416.97 |
| no-this-alias | 1,000 | 910.30 | 1,271.13 | 5,791.80 |

Final throughput test passed in 26.180s. Earlier pre-cache measurements were
superseded; no earlier timing is used in this table. All owned changes are under
rule directories and this claim/evidence directory. No pull request was opened.
