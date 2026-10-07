Built splitThemeKey, joinSegments and breakpointGroupOrder in separate .a files: eighteen dependency edges across six consumers.
Commits: 2390671 claims breakpoint lookup; 4e513f3 yields earlier-owned duplicates and claims split/join before code; implementation follows final validation.
Commands: fifteen-helper gate PASS 172.931s, 2,530,058 comparison lines; vet/format clean; six uncached input-oracle fixtures PASS 1.461s; capture regeneration identical.
Mutants: one-hyphen guard, dropped empty segments, cached mutable list, omitted join separator, wrong registration name and reversed first-match order.
Not covered: full repository test gate, complete consuming-rule integration, external Tailwind/corpora, invalid UTF-8 keys and non-safe-integer registration orders.

## Selection and ownership

All twelve prior helpers were tested and pushed through c79d039 before this batch. Refreshed every origin codex/lint-helpers* branch and checked every complete claims file against readiness.json. All higher named counts are reserved; comments remain reserved by shared HELPERS.md and generic strict-option decoding is per-rule configuration work. These three tie at the highest available named count, six.

Initial claim 2390671 at 01:41:11 UTC collided on isValidThemePrefix and namespaceForVariantRoot with slot 02 c332c14 at 01:41:08 UTC. Yield both; their passing local implementations were removed and are not delivered or counted. Retain breakpointGroupOrder. During replacement selection the fresh claim check also found isValidNamedValue already owned by slot 01, so it was never claimed or implemented here. Replacement claim 4e513f3 reserves splitThemeKey and joinSegments before code. Subsequent all-branch fetch confirms only slot 03 claims each of the final three symbols.

## Oracle and observations

The actual Go private helpers decide expected output. A temporary overlay exports the helpers without changing cohere's worktree. For group lookup, the isolated oracle process temporarily replaces FrameworkVariantRegistrations and restores it after each call; no shared process/table or repository file is mutated. The unmodified generated table and the same table with sm removed are included. Cases include empty lists, zero-order present, absent name, case/whitespace/NUL differences, negative orders, duplicate sm records with distinct orders and every captured text in a candidate registration. The Adamic input retains precisely the fields the Go helper reads and makes the generated table an explicit caller dependency.

Capture includes all six consumers and 157 actual runtime source inputs. Key and join controls use each complete source, extracted lexical tokens, actual framework names, U+0000..U+00FF in multiple positions, repeated/interior/trailing hyphens, NUL, newlines, non-ASCII and supplementary Unicode. Go and Adamic segment observations encode UTF-16 units, segment count and boundaries. Every split result is mutated before a second call so cached mutable backing storage changes an independently observed result. Ordered arbitrary segment arrays also test joining separately from splitting.

11,571 lines match actual Go, source Node, emitted JavaScript through oracle/node.mjs and sanitized native. The initial duplicate-only suite passed but is excluded from all delivered counts. The retained five-mutant focused run passes in 10.860s; the additional fresh-list mutant passes in 1.962s. All mutants compile and exit 0 without stderr before their changed semantic output is compared. Native ASan/UBSan remains enabled.

Mutant witnesses: allowing one leading hyphen differs at line 283 (one empty segment versus no segments); dropping empty segments differs at 286 (no segments versus one empty); caching a mutable split list differs at 288 (changed versus empty first segment); joining without a separator differs at 290 (NUL versus NUL plus hyphen); choosing lg differs at 9639 (2/true versus 0/false); reverse scanning differs at 9641 (23/true versus -17/true).

The upstream Tailwind capture rule-package gate exits 1: configured /Users/kirkouimet/Projects/ahra/app/_theme/styles installations and external corpora are unavailable, so actually-ran and live-placement guards fail. Capture occurs before external skips and still covers every consumer. This remains a failed rule-package gate, not a passing whole-rule result.

## Rules and readiness

Each helper removes one dependency from each of these six rules:

- better-tailwindcss/enforce-canonical-classes
- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-consistent-variant-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

Eighteen edges are removed across six distinct rules, with zero final blockers removed alone. Residual readiness subtracts only this slot's fifteen retained helpers and does not assume another branch has merged. This is dependency readiness, not completed findings/fixes/suggestions parity.

## Commands and environment

All test output is written directly to logs, never piped. Existing setup remains valid: go ready 0s, clang ready 1s, node ready 1s, submodules ready 1s, cache warm 106s, done 106s. nproc=5, cgroup cpu.max=400000 100000. Go 1.27.1, clang 20.1.8, Node 24.19.0; source /workspace/adamic-tools/env.sh. Cohere pin 715ba94f3608a6500086b1076ce5cb7e51b836db.

- python3 stage1/cohere/lint/helpers/slot03/batch5/testdata/regenerate.py: evidence/regenerate.log and capture.log.
- go test ./stage1/cohere/lint/helpers/slot03 -run '^TestBatch5' -count=1 -v -timeout=20m: evidence/replacements.log, PASS 10.860s.
- go test ./stage1/cohere/lint/helpers/slot03 -run '^TestBatch5Mutants/split_theme_key[.]a#02$' -count=1 -v -timeout=10m: evidence/freshness.log, PASS 1.962s.
- go test ./stage1/cohere/lint/helpers/slot03 -count=1 -v -timeout=20m: evidence/final.log.
- go vet ./...: evidence/vet.log.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m: evidence/oracle.log.
- gofmt -l cmd internal stage1/cohere/lint/helpers/slot03: evidence/gofmt.log.
- Repeated source/coverage capture and SHA-256 equality: evidence/reproducibility.log.

## Final gate observations

Complete slot package PASS 172.931s. Original, second, third, fourth and fifth batch baseline outputs contain 1,115,085, 208,636, 1,117,150, 77,616 and 11,571 lines, respectively: 2,530,058 total. Every batch compares actual Go, Node source and sanitized native; the last three also compare emitted JavaScript. All twenty compiling native semantic mutants are caught, plus the missing-consumer coverage mutant.

Final witnesses for every semantic mutant: component name 159; whitespace 847; listener kind 1114948; listener freshness 1114949; intrinsic kind 24; hole inheritance 179230; dispatcher route 185959; hex uppercase F 43; entity substitution 1114212; parameter backing storage 1115385; absent parameter list 1115375; CSS escape FF 13; CRLF pairing/lazy peeks 3585; theme hyphen boundary 75243; key two-hyphen guard 283; empty segments 286; mutable split caching 288; join delimiter 290; registration exact name 9639; registration first match 9641. evidence/final.log records exact got/Go output for every witness. Compile failures, runtime failures or sanitizer errors are not credited as mutant catches.

Repository-wide vet and the recorded formatting scan exit 0 with empty logs. The filtered uncached input oracle passes all six fixtures in 1.461s with zero cache hits and six probe misses. Repeated capture produces byte-identical sources.jsonl.gz and coverage.json. Final wildcard fetch and full claim scan confirm unique ownership of all three delivered helpers. No shared harness, registration generator, compiler source or other worker directory was edited. The full repository test gate, complete consuming-rule integration and unavailable external Tailwind/corpus gates remain uncovered as described above.
