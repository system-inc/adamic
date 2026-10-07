Built: isValidNamedValue, isValidArbitrary and parseThemeOptions in .a; 18 dependency entries removed across six rules.\
Commits: claims 150bf52, 3eb4a65, e6f0999; implementations 482eec9, 11c3c3d, c093291, all pushed; breakpoint duplicate withdrawn.\
Commands and outputs: targeted PASS 15.125s; full helper package PASS 258.542s; vet exit 0; filtered uncached oracle PASS 1.105s; earlier setup 75s, nproc 5.\
Mutants: seven new compiling semantic mutants caught against actual Go; all 31 prior/inherited checks rerun and caught (30 semantic, one byte-domain refusal).\
Not covered: whole-rule parity, dynamic fixture generation, malformed raw UTF-8, cross-slot callback integration or full repository test gate; zero additional final blockers removed.

# Slot 01 fifth batch report

## Delivered behavior and evidence

All previous twelve slot helpers were already tested and pushed before this batch. This helper unit has no outstanding rule claims. New Adamic files are .a, one production helper per file. Existing shared registration, rule harness, compiler ownership files and rule directories remain unchanged. Every private Go helper is called through a slot-owned build overlay, without modifying cohere.

The named-value predicate preserves nonempty ASCII letters/digits plus underscore, dot, percent and hyphen. The arbitrary-value scanner preserves actual Go behavior, including its permissive final return for unfinished bracket stacks and quoted strings. Opening braces do not push, mismatched closers on a nonempty stack do not pop, empty-stack closers and top-level semicolons reject, and escapes/quoted content are skipped. Although the Go doc comment describes rejecting unbalanced brackets, observed behavior of the actual implementation accepts unfinished openings. No stronger validation is silently substituted.

The theme option parser preserves bits reference=2, inline=1, default=4 and static=8, OR accumulation, ignored unknown words and last-prefix-wins, including clearing with an empty prefix. The original params reach an explicit trimming/segmentation callback once. Actual Go TrimSpace and segment provide callback observations in isolation; the actual Go parseThemeOptions separately decides expected output. These callbacks are integration dependencies rather than approximated JavaScript whitespace behavior. Prefix output is observed in UTF-16 code units, including astral Unicode and control characters.

All three corpora include all six consuming rule files: 625 complete fixture string expressions, plus character and grammar controls, yielding 1,706 helper queries each in the final corpus. Every corpus matches actual Go on source Node via oracle/node.mjs, sanitized native and emitted JavaScript. Character controls cover U+0000 through U+00FF in standalone, surrounding named-value and function forms. Grammar controls include unfinished openings, unmatched closers, nested/mismatched brackets, opening braces, top-level/inner semicolons, escaped punctuation, unterminated/escaped quotes, astral text, all theme bits, repeated options, repeated/empty/Unicode prefixes, tabs, NBSP and quoted/nested prefix arguments.

These corpora statically extract complete Go string expressions, including concatenations, labels and messages. They do not execute whole consuming-rule harnesses or dynamically generated inputs. Actual Go is the external behavior oracle; Node source and emitted JavaScript are additional execution comparisons. Every semantic mutant must compile, exit 0 with empty stderr and differ from actual Go stdout. Compiler errors, panic and sanitizer findings do not count as semantic catches.

## Claims and race resolution

Fetched all six origin codex/lint-helpers* branches and read every complete claims file before each reservation. The 24-consumer comment bundle remains reserved by the base branch's shared HELPERS.md outside claims/. All larger concrete helpers are reserved; each selected helper ties the highest unclaimed concrete count at six. The generic strict-option entry is a rule-local schema gap rather than a named shared Go helper. No further helper was claimed after these three.

Named-value claim 150bf52 preceded code; implementation 482eec9 passed and was pushed before the next claim. Breakpoint bucket claim 716e80b collided with slot 04 a33e02d: slot 04 at 01:42:42 UTC precedes slot 01 at 01:42:55. The passing duplicate 8bc257c was removed by 3eb4a65 and is not delivered or credited. Its log bucket.log is explicitly withdrawn historical evidence. The same commit publishes the replacement isValidArbitrary claim before code. Its implementation 11c3c3d passed and was pushed before the final claim e6f0999; theme implementation c093291 passed and is pushed.

Fresh fetches also excluded newly reserved isValidThemePrefix, namespaceForVariantRoot, joinSegments, Comment and Declaration; no code or claims for those symbols were delivered here. Final fetch observes only slot 01 claims for all three retained symbols. Complete fetched snapshots remain in evidence/slot01-wave5/claims-*.json. Normal commits/pushes only; no PR or history rewrite.

## Commands and outputs

All test stdout/stderr was redirected directly to logs, never piped. Evidence paths below are relative to stage1/cohere/lint/helpers; commands ran from repository root with that full path prefix.

- `source /workspace/adamic-tools/env.sh` in each build shell.
- `go test ./stage1/cohere/lint/helpers -run '^TestSlot01Wave5' -count=1 -v -timeout=20m > evidence/slot01-wave5/targeted.log 2>&1`: PASS 15.125s, all seven new tests.
- `go test ./stage1/cohere/lint/helpers -count=1 -v -timeout=20m > evidence/slot01-wave5/final.log 2>&1`: PASS 258.542s, 35 top-level tests.
- `go vet ./... > evidence/slot01-wave5/vet.log 2>&1`: exit 0, empty log.
- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m > evidence/slot01-wave5/oracle.log 2>&1`: PASS 1.105s, six fixture/probe misses.
- `gofmt -l` over slot-owned Go files and `git diff --check`: empty output.

Initial named and arbitrary runs passed in 3.598s and 10.811s; retained named.log and arbitrary.log. Initial oracle build failed because copied unused TypeScript shim imports remained; they were removed from the slot-owned file, and all accepted runs pass. That compiler failure is not counted as a mutant catch. No setup rerun was needed: this unit's setup already completed in 75s, nproc 5, with timing lines go ready 0s, clang ready 1s, node ready 1s, submodules ready 1s, cache warm 75s, done 75s, cgroup cpu.max 400000 100000 and 17.6 GB. Initial moving-tree setup failure and successful retry remain in SLOT01_REPORT.md and its evidence.

## Every new compiling semantic mutant

| Helper | Mutation | Actual Go witness |
|---|---|---|
| isValidNamedValue | exclude percent | line 637: mutant false, Go true |
| isValidNamedValue | accept empty string | line 1: mutant true, Go false |
| isValidArbitrary | omit top-level semicolon rejection | line 4: mutant true, Go false |
| isValidArbitrary | require empty final bracket stack | line 648: mutant false, Go true |
| isValidArbitrary | omit escape skipping | line 1021: mutant true, Go false |
| parseThemeOptions | reference sets inline bit | line 1693: mutant mask 13, Go mask 15 |
| parseThemeOptions | retain first nonempty prefix | line 1695: mutant first, Go second |

All seven compile and execute successfully; Go output differences catch them. Full regression additionally reruns all 31 prior/inherited mutant checks detailed in SLOT01_REPORT.md, SLOT01_WAVE2_REPORT.md, SLOT01_WAVE3_REPORT.md and SLOT01_WAVE4_REPORT.md; final.log retains every witness. These cover JSON control/oneOf/interpolation/unknown-field validation; filename normalization and class expression detection; default entries and three array freshness checks; factory order/invalid patterns/false attributes; cache bypass/unbound/empty entries; quote stripping/short ranges; attribute membership/false entries; entity overflow/surrogates/named lookup; component ASCII restriction/range stride; JSX first nonstring/nil duplicate handling and decoding; uppercase F, container context and invalid byte-domain refusal. The last is an explicit Go-parameter domain check rather than an invented Go boolean for 256. Total retained checks: 37 compiling semantic mutants plus one byte-domain refusal mutant.

## Every affected rule

All three helpers remove one dependency each from each of these six rules:

- `better-tailwindcss/enforce-canonical-classes`
- `better-tailwindcss/enforce-consistent-class-order`
- `better-tailwindcss/enforce-consistent-variant-order`
- `better-tailwindcss/enforce-shorthand-classes`
- `better-tailwindcss/no-conflicting-classes`
- `better-tailwindcss/no-unknown-classes`

The increment is 18 dependency entries across six distinct consumers, with zero additional fully helper-ready rules. Cumulative slot-only final helper blockers removed remain the same three structure rules from the first batch. `slot01_wave5_readiness.json` lists each helper's complete consumers and residual dependencies after this batch and after all fifteen retained slot 01 helpers. Concurrent workers' unmerged helpers are not credited, and no rule implementation status is changed.

Not covered: whole-rule findings/spans/fixes, suggestion serialization, shared profile compilation, arbitrary cross-slot AST adapters, integration of the trimming/segmentation callback, dynamic fixture generation, malformed raw UTF-8 byte strings or isolated UTF-16 surrogates. Native/source/emitted-JavaScript comparisons cover valid Unicode text. The full repository go test ./... gate was not run; bounded verification runs the complete touched helper package, repository-wide vet and a filtered uncached external oracle.
