Built: Collapse topOfStack, peekByte and sortedKeys, one helper per .a file.
Implementation: 076e912; reservations 06a6dab and 906d0ec were pushed before code.
Checks: touched package PASS 25.170s; uncached input oracle PASS 1.560s; vet/format clean; setup 165s, nproc 5.
Mutants: first stack byte, first input byte, nonzero out-of-range, omitted false keys, reversed byte order, UTF-16 comparator; all compiled and were caught on source Node, emitted JavaScript and sanitized native. Consumer omission caught.
Limits: no full repository gate or production integration; actual sortedKeys fixture calls supplied only empty maps; invalid UTF-8 map keys outside the decoded-string boundary. Byte views support arbitrary bytes.

## Readiness and exact consumers

18 dependency edges are removed across six distinct consumers. Each helper removes one prerequisite from all six rules below, but none removes a final blocker alone. Frozen original readiness.json is unchanged. This directory's ledger removes the eleven uniquely owned slot 04 helpers only; 46 original helper-ready rules become 47, solely adding @next/next/no-img-element. Other workers' ports are not assumed integrated. These counts describe prerequisites, not finished rule implementations.

### rules/tailwind/collapse.topOfStack

Six dependency removals, zero final blockers removed alone.

- `better-tailwindcss/enforce-canonical-classes`
- `better-tailwindcss/enforce-consistent-class-order`
- `better-tailwindcss/enforce-consistent-variant-order`
- `better-tailwindcss/enforce-shorthand-classes`
- `better-tailwindcss/no-conflicting-classes`
- `better-tailwindcss/no-unknown-classes`

### rules/tailwind/collapse.peekByte

Six dependency removals, zero final blockers removed alone.

- `better-tailwindcss/enforce-canonical-classes`
- `better-tailwindcss/enforce-consistent-class-order`
- `better-tailwindcss/enforce-consistent-variant-order`
- `better-tailwindcss/enforce-shorthand-classes`
- `better-tailwindcss/no-conflicting-classes`
- `better-tailwindcss/no-unknown-classes`

### rules/tailwind/collapse.sortedKeys

Six dependency removals, zero final blockers removed alone.

- `better-tailwindcss/enforce-canonical-classes`
- `better-tailwindcss/enforce-consistent-class-order`
- `better-tailwindcss/enforce-consistent-variant-order`
- `better-tailwindcss/enforce-shorthand-classes`
- `better-tailwindcss/no-conflicting-classes`
- `better-tailwindcss/no-unknown-classes`

## API and behavior

`topOfStack(stack)` accepts a readonly byte vector, each element an integer in 0..255. It returns the final byte, not the first, and returns zero for an empty vector. Go nil and empty stacks have the same observable result. It reads the current last value after caller mutation and never writes the stack.

`peekByte(input, index)` accepts a readonly byte view of Go string storage and an integer index. It returns that exact byte, or zero at either out-of-range end. A byte view preserves Go's raw indexing, including invalid UTF-8 and NUL bytes; indexing a decoded JavaScript string would change the contract. The consumer adapter obtains the bytes from actual Go source/literal strings. Byte arrays and indices must satisfy their integer domains. The helper is ready to connect to production byte adapters, not already integrated into the separately owned CSS parser.

`sortedKeys(set)` returns a fresh list of every map key, ignoring the boolean value, in Go UTF-8 byte order. It leaves the map untouched. Its explicit comparator distinguishes U+10000 from U+E000 and preserves prefix, NUL, empty-key and normalization-distinct ordering. The input boundary is decoded Unicode configuration/source strings; malformed raw UTF-8 Go map keys are not represented by JavaScript strings and are not promised. Unlike the byte helpers, this helper's key representation is text.

All new Adamic files are .a. Existing options_json.ts is used only to decode fixture input. No shared harness, registration generator, production cohere source or protected compiler files are edited. All implementation/tests/data/docs live in slot04_wave4/, with the required claims/04.md update.

## Observations and fixture limitations

Actual Go private functions at cohere 715ba94f3608a6500086b1076ce5cb7e51b836db decide expected behavior through oracle-only exported wrappers. The byte oracle uses real Go strings and []byte stacks, independently of Adamic's byte-view interface. Expected outputs and actual observations must exit successfully with no stderr and match exact stdout.

- All 256 byte values, all 65,536 ordered two-byte pairs, empty input, every singleton, prefix-stack mutations and extreme indices match Go on source Node, emitted JavaScript and ASan/UBSan native: 723,206 output lines.
- 775 bounded byte-vector controls include long/random stacks, NUL, raw invalid UTF-8 bytes, all bytes in one vector and nested-closer sequences: 12,116 identical lines.
- 149 asserted consumer inputs cover every one of the six rules. Their source strings and each parsed string/no-substitution-template literal are observed as raw byte views: 23,608 identical lines.
- Runtime capture at the actual sortedKeys entry point recorded 137 calls during upstream Tailwind fixture tests. They all supplied empty maps, so there is one distinct captured map and three comparison lines after fresh-list and map-update checks. This is an observation, not evidence that production maps are empty.
- 253 independent map controls cover nonempty maps, true/false values, overwrite/insertion order, empty/NUL keys, Unicode and random mixtures. Each is checked before and after returned-list mutation, then after adding a false-valued map entry: 759 identical map-state lines. These controls supply the nonempty coverage absent from the runtime fixture capture.

The source corpus is a pinned subset of the original full asserted Tailwind capture. The new sorted-map capture reproducer is testdata/capture.py; it adds a recording call to sortedKeys only through a temporary Go overlay and leaves the Go result untouched. It also captures all six source consumer groups. Pinned tailwindcss@4.3.3 is installed with scripts disabled into a temporary stylesheet fixture to work around the upstream hardcoded developer path. The documented selector excludes external live-repository population gates. Neither this batch nor earlier batches assert those live gates pass.

The oracle supports --cases to export byte-vector fixtures and --full for exhaustive bytes/pairs. Nil stack and empty byte view are intentionally represented by the same empty vector. testdata/readiness.py regenerates the conservative residual ledger from the original frozen file.

## Commands and output

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/lint/helpers/slot04_wave4/testdata/capture.py > stage1/cohere/lint/helpers/slot04_wave4/evidence/capture.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot04_wave4 -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot04_wave4/evidence/tests.log 2>&1
go vet ./... > stage1/cohere/lint/helpers/slot04_wave4/evidence/vet.log 2>&1
gofmt -l cmd internal stage1/cohere/lint/helpers/slot04_wave4 > stage1/cohere/lint/helpers/slot04_wave4/evidence/gofmt.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=10m > stage1/cohere/lint/helpers/slot04_wave4/evidence/oracle.log 2>&1
python3 stage1/cohere/lint/helpers/slot04_wave4/testdata/readiness.py > stage1/cohere/lint/helpers/slot04_wave4/evidence/readiness.log 2>&1
```

Touched package passes in 25.170s. Six uncached input oracle fixtures pass in 1.560s with zero cache hits and six probe misses. Vet and formatting output are empty and successful. Prior helper code is unchanged and its previously passing package gates were not repeated. The full repository gate was not run. Test output is written to logs, never piped.

## Mutants

Every retained semantic mutation compiled and executed without stderr in all three modes, then disagreed with actual Go. A compile error or sanitizer crash would not count as a catch.

| Mutation | First catch on source Node, native and emitted JavaScript |
| --- | --- |
| topOfStack returns first byte | Line 16: top 0 instead of top 1. |
| peekByte returns first byte | Line 15: peek 1 is 0 instead of 1. |
| peekByte returns 1 outside range | Line 2: negative extreme gives 1 instead of 0. |
| sortedKeys omits false-valued entries | Line 3: added false entry missing. |
| sortedKeys reverses byte comparator | Line 7: ordered keys differ. |
| sortedKeys uses UTF-16 comparator | Line 7: supplementary/BMP-private order differs. |

Removing all enforce-canonical-classes source rows makes the six-consumer coverage detector fail. This metadata omission is distinct from the compiling semantic mutants.

Before ownership refresh, the now-yielded escape predicate also passed its actual-Go four-way comparison. Dropping form feed and accepting vertical tab both compiled and were caught; the initial package passed in 24.739s. No escape implementation, fixture dependency or mutant is counted in the final delivery. The final tests.log covers only the retained three helpers.

## Reservations

All eight previously retained helpers were ported, tested and pushed through e42f5d0 before this continuation. Every origin codex/lint-helpers* branch and current claim was fetched and inspected before reservation. Every greater-than-six named helper was reserved; the generic strict-option ledger label is remaining rule-local decoding beyond the existing shared option helpers. The three initial byte leaves tied at six consumers. Their reservation 06a6dab was pushed before code.

A later refresh exposed an earlier isEscapeTerminator claim: slot 03 685fc6d at 01:31:08 UTC precedes slot 05 ea3c330 at 01:31:27 and this slot 06a6dab at 01:31:48. This slot yielded and removed that duplicate. topOfStack and peekByte retained unique ownership. Another complete refresh found sortedKeys unclaimed at the highest available count, six; replacement reservation 906d0ec was pushed before code. Subsequent full refreshes showed these three are claimed only by slot 04. No fourth helper is claimed and no pull request is opened.

## Toolchain timing

The session setup succeeded before the first batch; no setup rerun was necessary. nproc was checked again and is 5. Environment: /workspace/adamic-tools/env.sh. Go 1.27.1, clang 20.1.8, Node 24.19.0. Full setup log remains ../evidence/04/setup.log:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (2s)
setup: build cache warm (165s)
setup: done in 165s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```
