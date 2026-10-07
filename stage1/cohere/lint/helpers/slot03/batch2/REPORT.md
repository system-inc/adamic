Built intrinsic JSX matching, template hole edges and class-value dispatch, one `.a` file per helper.  
Commits: implementation 48be7b0, ce7f300, 2bd369a; claims b902cd6, 48be7b0, ce7f300 were pushed before code.  
Checks: go test slot03 PASS 128.997s, 1,323,721 Go/Node/native lines; vet and six uncached oracle fixtures PASS.  
Mutants: intrinsic wrong-kind acceptance, discarded inherited leading edge, and call dispatch to the attribute reader.  
Not covered: complete rule implementations, external Tailwind installations/corpora, full repository gate and surface-reader implementations.  

The branch was already fully pushed before this continuation. Ahra's later correction stops further claiming; this report finishes only the three continuation helpers already claimed and pushed. Every selection fetched all six origin codex/lint-helpers* branches and read all claims. All dependencies with more than 11 consumers were reserved. Each selected helper had 11 consumers, tied for the highest unclaimed count. The shared comments bundle was also excluded under its existing HELPERS.md claim.

The continuation removes 33 dependency edges across 22 distinct rules. No rule loses its final blocker from these helpers alone. readiness.json conservatively removes only this slot's six delivered helpers; it does not assume the other branches have merged.

## Contracts and evidence

See README.md for API details. Nil nodes use a negative arena index; parser kinds and decoded text come from the common AST adapter. The dispatcher's three surface readers are explicit function dependencies. Their string handles remain owned by the surface adapter. The Go oracle instruments the actual dispatcher to record delegate calls without replacing its decisions, and serializes actual Go surface observations, including text, origin, range and edge flags. The port forwards those results unchanged.

Captures include all 22 consumers and 717 actual runtime source inputs, with fixture hooks before external-engine skips. capture.log remains a failing upstream Tailwind rule-package gate because its external installations/corpora are unavailable; regenerate.py accepts only the eight previously identified external/corpus failures and still requires every selected consumer. It is not reported as a passing rule-package gate. coverage checks reject missing/extra consumers, count drift and pin drift. Cohere pin: 715ba94f3608a6500086b1076ce5cb7e51b836db.

Synthetic controls include nil, wrong node kinds, exact case/text matching, member and namespace tags, empty inherited boundaries, all sixteen position/parent-edge flag combinations, the six ASCII whitespace characters, other control characters and Unicode boundaries. UTF-16 and Go UTF-8 boundary tests agree here because the only accepted whitespace bytes are ASCII.

Each mutant compiles, runs to exit 0 with no stderr and changes a semantic output line against Go. The continuation witnesses are output line 24 (wrong-kind intrinsic acceptance), 179230 (discarded inherited leading edge), and 185959 (call routed to attribute instead of callee). Compilation or sanitizer failure is not credited. Temporary copies hold all mutants; production sources are unchanged.

The dispatcher collision was resolved by claim timestamps: slot 03 ce7f300 at 00:48:49 UTC precedes slot 05 dd9e8c3 at 00:48:53 UTC. Slot 03 retains ownership and recorded this in its pushed claim file. No shared registration generator or shared test harness was edited.

## Rules whose dependency is removed

Intrinsic matching (11):

- @next/next/google-font-display
- @next/next/google-font-preconnect
- @next/next/next-script-for-ga
- @next/next/no-css-tags
- @next/next/no-head-element
- @next/next/no-html-link-for-pages
- @next/next/no-page-custom-font
- @next/next/no-styled-jsx-in-document
- @next/next/no-sync-scripts
- @next/next/no-unwanted-polyfillio
- react/jsx-no-target-blank

Both holeEdges and readClassValues independently remove a dependency from the same 11 rules:

- better-tailwindcss/enforce-canonical-classes
- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-consistent-important-position
- better-tailwindcss/enforce-consistent-variable-syntax
- better-tailwindcss/enforce-consistent-variant-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-deprecated-classes
- better-tailwindcss/no-duplicate-classes
- better-tailwindcss/no-unknown-classes
- better-tailwindcss/no-unnecessary-whitespace

## Commands

All commands source /workspace/adamic-tools/env.sh; outputs go directly to evidence logs. Setup from the first batch remains valid: go ready 0s, clang ready 1s, node ready 1s, submodules ready 1s, cache warm 106s, done 106s; nproc printed 5 again in this continuation. No compiler files were edited.

- python3 stage1/cohere/lint/helpers/slot03/batch2/testdata/regenerate.py: evidence/regenerate.log and capture.log.
- Repeated capture and SHA-256 comparison: evidence/reproducibility.log.
- go test ./stage1/cohere/lint/helpers/slot03 -count=1 -v -timeout=20m: evidence/final.log, includes the original three helpers and continuation.
- go vet ./...: evidence/vet.log.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m: evidence/oracle.log.
- gofmt -l stage1/cohere/lint/helpers/slot03: evidence/gofmt.log.

## Final observed results

- Complete touched package: PASS, 128.997s. Continuation: 208,636 Go/Node/sanitized-native output lines; original helpers: 1,115,085 lines. Seven compiling semantic mutants and the consumer-coverage mutant were caught.
- Continuation mutants: wrong intrinsic kind at line 24, inherited leading edge dropped at 179230, call incorrectly routed to attribute at 185959.
- Original mutants rerun: PureComponent removed at 159, vertical tab removed at 847, VariableDeclaration replaced at 1114948, shared mutable listener list at 1114949. Each native mutant ran successfully and disagreed semantically with Go.
- Missing-consumer mutant: dropping better-tailwindcss/no-unknown-classes is rejected by coverageVerdict.
- go vet ./...: exit 0, empty log. Both gofmt checks: exit 0, empty logs.
- Filtered uncached input oracle: PASS, 1.329s; six fixtures passed, zero probe cache hits and six misses.
- Repeated runtime capture: PASS, byte-identical sources.jsonl.gz and coverage.json; hashes recorded in reproducibility.log.

All useful work and evidence are committed and pushed to codex/lint-helpers-03. No PR is opened and no further helpers are claimed after Ahra's correction.
