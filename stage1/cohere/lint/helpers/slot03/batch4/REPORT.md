Built three .a Collapse helpers: escape terminators, lazy whitespace lookahead and ignored-theme-key boundaries, each serving six rules.
Commits: 685fc6d publishes all claims before code; implementation and evidence are committed after the final gate.
Commands: twelve-helper gate PASS 156.584s, 2,518,487 comparison lines; vet/format clean; six uncached input-oracle fixtures PASS 1.542s; capture regeneration identical.
Mutants: dropped FF, accepted lone CR and allowed word-prefix matches; all compile, exit successfully and are caught by semantic comparison.
Not covered: full repository test gate, full rule integration, external Tailwind installation/corpora and out-of-contract numeric/invalid UTF-8 inputs.

## Ownership and selection

All nine previous slot helpers were tested and pushed through 35d1274 before this batch began. Every origin codex/lint-helpers* branch was fetched, and each complete claims file checked against the base readiness.json. All larger named symbols are reserved; comments remain owned by the shared HELPERS.md bundle, and strict option decoding is rule-local custom configuration beyond the original helpers. These three tie at the highest unclaimed named-symbol count of six.

Claim 685fc6d at 01:31:08 UTC precedes slot 05's duplicate escape reservation ea3c330 at 01:31:27 UTC and slot 04's 06a6dab at 01:31:48 UTC. Slot 03 retains the earliest claim. Final refresh confirms both slots 04 and 05 have withdrawn their duplicates and selected replacements. All three claims were pushed before source was written. No further helper is claimed. Shared harness, generator, compiler and other slot directories are unchanged.

## Observations

The real Go private helpers decide the expected output, via oracle-only exports in a temporary Go overlay. The baseline matches 77,616 lines across source Node, emitted JavaScript through the existing oracle/node.mjs runtime loader, and sanitized native. Exhaustive controls cover all 256 byte values and 65,536 ordered byte pairs, with callback position/order included in every lookahead output. Negative, boundary and positive indices are included. All 157 captured source strings contribute actual raw-byte positions and all source/custom-property token key comparisons. Theme controls cover every actual Go ignored namespace/key, equality, hyphen boundary, word continuation, unknown namespaces, case, Unicode and supplementary suffixes.

The helper oracle does not require the external live Tailwind engine. The separate upstream rule-package capture gate exits 1 because its configured /Users/kirkouimet/Projects/ahra/app/_theme/styles installation and external corpora are unavailable; explicit actually-ran/live-placement guards fail. This is a failed rule-package gate, not helper parity evidence. Capture nevertheless observes all six consumers before external skips; coverage checks exact consumer membership and pinned cohere commit.

The first capture attempt used an incorrect ledger suffix and selected no packages; the suffix was corrected. The first runner attempt used incorrect runtime/JSON method names and failed before baseline comparisons; it was corrected to programArguments, the readTextFile result and OptionsJson.node/field. These early failures are not counted as passing checks or mutant kills.

Compiling semantic mutants run in temporary copies, each exiting 0 with no stderr before comparison: removing FF differs at line 13 (false versus Go true); accepting CR without LF differs at line 3585 (true:10 versus Go false:10,11), proving both pairing and the second callback call; replacing equality-or-hyphen matching with bare prefix differs at line 75243 (true versus Go false). Native sanitizers remain enabled. No production source is left mutated.

## Rules and readiness

Each of all three helpers removes one dependency from each of these six rules:

- better-tailwindcss/enforce-canonical-classes
- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-consistent-variant-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

The batch removes eighteen edges across six distinct consumers, with zero final blockers removed alone. The residual ledger subtracts only this slot's twelve delivered symbols and does not assume another branch's work has merged. Helper readiness is a dependency calculation; no complete findings/fixes/suggestions parity claim is made.

## Environment and commands

The existing setup remains valid in this workspace: go ready 0s, clang ready 1s, node ready 1s, submodules ready 1s, build cache warm 106s, done 106s. nproc prints 5, cgroup cpu.max is 400000 100000. Go 1.27.1, clang 20.1.8 and Node 24.19.0; source /workspace/adamic-tools/env.sh before commands. Cohere pin 715ba94f3608a6500086b1076ce5cb7e51b836db.

All test output goes directly to logs, never piped.

- python3 stage1/cohere/lint/helpers/slot03/batch4/testdata/regenerate.py: evidence/regenerate.log and evidence/capture.log.
- go test ./stage1/cohere/lint/helpers/slot03 -run '^TestBatch4' -count=1 -v -timeout=20m: evidence/helpers-fixed.log, PASS 8.853s.
- go test ./stage1/cohere/lint/helpers/slot03 -count=1 -v -timeout=20m: evidence/final.log.
- go vet ./...: evidence/vet.log.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m: evidence/oracle.log.
- gofmt -l cmd internal stage1/cohere/lint/helpers/slot03: evidence/gofmt.log.
- Repeated capture and SHA-256 equality: evidence/reproducibility.log.

## Final gate observations

Complete slot package PASS 156.584s: 1,115,085 original, 208,636 second-batch, 1,117,150 third-batch and 77,616 fourth-batch output lines, totaling 2,518,487. All compare real Go, Node source and sanitized native; third/fourth batches also compare emitted JavaScript. All fourteen compiling semantic mutants are caught, plus the missing-consumer coverage mutant.

Full-gate mutant witnesses: component name line 159; whitespace 847; listener kind 1114948; listener freshness 1114949; intrinsic kind 24; hole-boundary inheritance 179230; dispatcher route 185959; hexadecimal uppercase F 43; entity substitution 1114212; parameter shared backing storage 1115385; absent parameter list 1115375; CSS escape FF 13; CRLF pairing/lazy peeks 3585; theme hyphen boundary 75243. Each compiles and runs successfully before its changed semantic output is compared. The full log records exact got/Go output for every witness.

Whole-repository vet and the recorded formatting scan exit 0 with empty logs. The filtered uncached input oracle passes all six fixtures in 1.542s, with zero cache hits and six probe misses. Repeated capture regenerates sources.jsonl.gz and coverage.json byte for byte. No full-repository test gate or complete consuming-rule implementation is claimed; external Tailwind rule-package guards remain failed as described above.
