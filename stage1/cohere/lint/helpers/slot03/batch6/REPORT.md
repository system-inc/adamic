Built AtRule, StyleRule and VariantRegistry.nextOrder in separate .a files: eighteen dependency edges across six consumers.
Commits: d8c56c5 claims all three before code; implementation and evidence follow the final gate.
Commands: eighteen-helper gate PASS 193.568s, 2,564,624 comparison lines; vet/format clean; six uncached input-oracle fixtures PASS 1.647s; capture regeneration identical.
Mutants: wrong constructor kinds, copied backing storage, shared slice headers, reused node records, ignored group order, wrong increment and mutating reads.
Not covered: full repository test gate, complete rule integration, external Tailwind/corpora, arbitrary invalid UTF-8 and full Go int64 overflow/unsafe numeric states.

## Selection and ownership

All fifteen prior helpers were tested and pushed through e3c2fbe before selection. Every origin codex/lint-helpers* branch was fetched, every claims file inspected and dependencies ranked from readiness.json. Higher named counts are reserved; comments remain reserved by shared HELPERS.md and generic strict-option configuration is rule-local. These three tie at the highest available count, six. VariantRegistry.Has was already reserved by slot 04 during pre-publication refresh, so it was never claimed or implemented here. Claim d8c56c5 was pushed before source was written. Subsequent full claim scans confirm unique ownership of all three final symbols.

## Representation and oracle

The Go public constructors and real private nextOrder method decide output; an owned temporary overlay exports the private method without editing cohere. Pointer identities become stable child indices and nil pointers become -1. All Go constructor fields are observed, including zero unrelated fields, nil context, child order/identity and nil-versus-empty lists. Strings are encoded as UTF-16 units to retain supplementary Unicode in the comparison.

The child list adapter copies a Go slice header into a fresh view while sharing its storage. Tests explicitly mutate an element, then truncate the returned list to [:0:0], observing the caller's original length, capacity, nil state and contents as well as a second fresh constructor result. Controls include nil, nonnil empty, empty spare capacity, singleton, duplicate and nil children, and nonempty spare capacity. The valid slice-view contract is explicit; performing append/reslice is outside these constructors and remains the AST adapter's operation.

Registry controls use actual private state fields in an oracle export, call nextOrder, snapshot state and call it again. Orders include safe-integer endpoints, zero and negative groups, absent groups, pinned framework orders and states derived from every captured text. Generated registry construction and full int64 wraparound are not reimplemented or claimed. A present zero group remains distinct from an absent group, and no read advances lastOrder.

Capture covers all six consumers and 157 actual runtime source inputs. Complete sources, extracted lexical tokens, code points U+0000..U+00FF in two positions, NUL, whitespace, raw names/params/selectors and supplementary Unicode feed constructor controls. The final adapter matches 34,566 lines across Go, source Node, emitted JavaScript and sanitized native.

The initial fixed-list adapter matched its smaller corpus but did not prove independent slice headers. It was replaced before delivery. Its superseded full gate was stopped, and its partial log is not counted as a pass. The final representation has independent headers, additional capacity controls and two additional semantic mutants. The runner also dispatches exactly one constructor per observation, matching Go's call sequence. Only the final representation's focused/full results are credited.

## Mutants

Every mutant compiles and exits 0 without stderr before its changed output is compared with Go; ASan/UBSan remains enabled. Temporary variants do not alter production source.

- AtRule wrong kind: line 1; copied storage: line 14 (original child remains 0 versus Go nil -1); shared header: line 14 (caller length/capacity become 0/0 versus Go 1/1); reused record: line 3 (changed name/selector leaks into a second result).
- StyleRule wrong kind: line 4; copied storage: line 17; shared header: line 17; reused record: line 6.
- nextOrder ignores group: line 29488 (-9007199254740989 versus Go -9007199254740991); wrong +2 increment: line 29485 (-9007199254740988 versus Go -9007199254740989); mutation during read: line 29486 (lastOrder changes to -9007199254740989 versus Go unchanged -9007199254740990).

## Rules and readiness

Each of the three helpers removes one dependency from each of:

- better-tailwindcss/enforce-canonical-classes
- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-consistent-variant-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

Eighteen edges are removed across six distinct rules; zero final blockers are removed alone. Remaining dependencies subtract only this slot's eighteen retained helpers and do not assume other branch work merged. Readiness is a dependency calculation, not complete findings/fixes/suggestions parity.

The separate upstream Tailwind capture gate exits 1 because configured /Users/kirkouimet/Projects/ahra/app/_theme/styles installations and external corpora are unavailable. Actually-ran and live-placement guards fail. Capture runs before these skips and still records every consumer; the failed rule-package gate is not credited as helper or full-rule success.

## Environment and commands

All test output is written directly to logs, never piped. Existing setup remains valid: go ready 0s, clang ready 1s, node ready 1s, submodules ready 1s, cache warm 106s, done 106s. nproc=5, cgroup cpu.max=400000 100000. Go 1.27.1, clang 20.1.8 and Node 24.19.0; source /workspace/adamic-tools/env.sh. Cohere pin 715ba94f3608a6500086b1076ce5cb7e51b836db.

- python3 stage1/cohere/lint/helpers/slot03/batch6/testdata/regenerate.py: evidence/regenerate.log and capture.log.
- go test ./stage1/cohere/lint/helpers/slot03 -run '^TestBatch6' -count=1 -v -timeout=20m: evidence/slice-headers.log, PASS 28.658s.
- go test ./stage1/cohere/lint/helpers/slot03 -count=1 -v -timeout=20m: evidence/final.log.
- go vet ./...: evidence/vet.log, rerun after the final representation change.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m: evidence/oracle.log, PASS 1.647s, all six fixtures uncached.
- gofmt -l cmd internal stage1/cohere/lint/helpers/slot03: evidence/gofmt.log, rerun after the final representation change.
- Repeated source/coverage capture and SHA-256 equality: evidence/reproducibility.log, byte-identical fixtures/metadata.

## Final gate observations

Complete slot package PASS 193.568s. The six retained batch baselines contain 1,115,085, 208,636, 1,117,150, 77,616, 11,571 and 34,566 lines: 2,564,624 total. All compare actual Go, source Node and sanitized native; the last four also compare emitted JavaScript. All thirty-one compiling native semantic mutants are caught, plus the missing-consumer coverage mutant. The earlier stopped fixed-list run is excluded.

Every mutant's final witness: component name 159; whitespace 847; listener kind 1114948; listener freshness 1114949; intrinsic kind 24; hole inheritance 179230; dispatcher route 185959; hex uppercase F 43; entity substitution 1114212; parameter backing storage 1115385; absent parameter list 1115375; CSS escape FF 13; CRLF/lazy peeks 3585; theme boundary 75243; key guard 283; empty segments 286; split freshness 288; join delimiter 290; registration exact name 9639; registration first match 9641; AtRule kind 1, backing storage 14, independent header 14 and freshness 3; StyleRule kind 4, backing storage 17, independent header 17 and freshness 6; nextOrder group precedence 29488, increment 29485 and unchanged state 29486. evidence/final.log records the exact got/Go observations for every witness. Compile failures, runtime failures, stderr and sanitizer errors are not credited as semantic catches.

Repository-wide vet and the recorded formatting scan exit 0 with empty logs after the final adapter change. The filtered uncached input oracle passes all six fixtures in 1.647s, with zero cache hits and six probe misses. Repeated capture reproduces source and coverage bytes exactly. Final wildcard fetch and complete claims inspection confirm unique ownership of all three helpers. No shared harness, registration generator, compiler source or other worker directory was edited. Full repository testing, complete consuming-rule integration, arbitrary invalid UTF-8, unsafe numeric/full int64 wraparound behavior and external Tailwind/corpus rule-package gates remain outside the passing observations.
