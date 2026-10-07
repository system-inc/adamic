Built: Collapse isJavaScriptSpace, isBlank and isValueSeparator, one helper per .a file.
Implementation: 9192bd4; reservation 4df6533 was pushed before code.
Checks: new package PASS 22.095s; retained traversal package PASS 16.483s; uncached input oracle PASS 1.746s; vet/format clean; setup 165s, nproc 5.
Mutants: BOM replaced with NEL, empty denied, blank stopped early, colon removed, CR accepted; all compiled and were caught by actual Go comparisons on source Node, emitted JavaScript and sanitized native. Three retained traversal mutants and consumer omissions also caught.
Not covered: full repository gate, live external Tailwind corpora, malformed raw UTF-8 strings, out-of-domain byte/rune inputs and direct production integration; no shared harness or registration edits.

## Readiness

This batch removes 18 dependency edges across six distinct consumers. Each helper removes one dependency from all six rules below; none alone removes a final helper blocker. The original readiness.json is frozen and unchanged. This directory's readiness.json records all residual prerequisites after the eight uniquely owned slot 04 helpers. The original 46 helper-ready rules become 47, solely adding @next/next/no-img-element. Other workers' helpers are not assumed integrated, and ready prerequisites do not mean finished rule implementations.

### rules/tailwind/collapse.isJavaScriptSpace

Six consumer dependencies removed, zero final blockers removed alone.

- `better-tailwindcss/enforce-canonical-classes`
- `better-tailwindcss/enforce-consistent-class-order`
- `better-tailwindcss/enforce-consistent-variant-order`
- `better-tailwindcss/enforce-shorthand-classes`
- `better-tailwindcss/no-conflicting-classes`
- `better-tailwindcss/no-unknown-classes`

### rules/tailwind/collapse.isBlank

Six consumer dependencies removed, zero final blockers removed alone.

- `better-tailwindcss/enforce-canonical-classes`
- `better-tailwindcss/enforce-consistent-class-order`
- `better-tailwindcss/enforce-consistent-variant-order`
- `better-tailwindcss/enforce-shorthand-classes`
- `better-tailwindcss/no-conflicting-classes`
- `better-tailwindcss/no-unknown-classes`

### rules/tailwind/collapse.isValueSeparator

Six consumer dependencies removed, zero final blockers removed alone.

- `better-tailwindcss/enforce-canonical-classes`
- `better-tailwindcss/enforce-consistent-class-order`
- `better-tailwindcss/enforce-consistent-variant-order`
- `better-tailwindcss/enforce-shorthand-classes`
- `better-tailwindcss/no-conflicting-classes`
- `better-tailwindcss/no-unknown-classes`

## Exact behavior and API

`isJavaScriptSpace(value)` takes a Go rune represented as a signed integer. It accepts the exact 25-code-point Go set: ASCII whitespace, NBSP, OGHAM SPACE MARK, U+2000..U+200A, the line/paragraph separators, narrow and medium mathematical spaces, ideographic space and BOM. It rejects NEL, zero-width space and the obsolete Mongolian vowel separator. No Unicode casing or generic trim function supplies this predicate.

`isBlank(value)` is true for empty or entirely matching whitespace strings. It uses the preceding helper. Iterating UTF-16 units is equivalent to Go's rune traversal for this verdict: every accepted rune is in the BMP, and all surrogate units and supplementary code points reject. It checks the whole string, so an initial whitespace rune does not hide later non-whitespace. Its input boundary is decoded Unicode strings; arbitrary invalid UTF-8 Go strings are not promised.

`isValueSeparator(value)` takes one byte represented as an integer in 0..255. Exactly colon, comma, equals, greater-than, less-than, LF, space and tab accept. Slash, CR, vertical tab, form feed and all high-bit bytes reject. This set intentionally differs from the whitespace helper. Fractional numbers or integers outside the Go byte/rune domains are not inputs to the ports.

No helper dependencies beyond the owned whitespace predicate remain inside these three implementations. All new Adamic files use .a; the runner imports the existing options_json.ts fixture decoder. These predicates are ready to connect to the separately owned Collapse ports, not already installed into those production modules.

## What was observed

The actual private Go helpers are exported only through a Go build overlay. Expected outputs call those unchanged functions at cohere pin 715ba94f3608a6500086b1076ce5cb7e51b836db. No cohere worktree, compiler, shared registration generator or existing shared rule harness edits are delivered.

- Exhaustive sweep: every integer 0..U+10FFFF, including surrogate integers, is checked against the rune predicate and its one-character blank verdict. Four signed/outside boundary values exercise false rune results. Every byte 0..255 is checked. Total: 1,114,372 output lines, identical on real Go, source Node, emitted JavaScript and sanitized native.
- String controls: 1,930 deterministic inputs cover all accepted whitespace, nearby rejecting characters, every two-character combination of the boundary alphabet, CRLF, BOM, Unicode, empty and random mixtures. Total: 22,597 output lines in all four modes.
- Consumer inputs: 149 asserted upstream fixture sources from all six rules, plus each parsed string/no-substitution-template literal. Source runes, UTF-8 bytes and blank verdicts yield 21,863 identical lines in all four modes.

The consumer data are a pinned subset of the first batch's complete asserted Tailwind capture. That capture and its external fixture workaround are documented in ../slot04_REPORT.md and ../testdata/slot04/capture.py. It installed pinned tailwindcss@4.3.3 with scripts disabled and excluded unavailable live-repository population gates. We do not claim those full upstream live gates pass.

`testdata/generate.py` reproduces the consumer subset and controls; `testdata/readiness.py` derives the residual ledger. Both use the frozen original readiness ledger. The expected output is always computed by actual Go, not by the fixture generator or Adamic implementation.

## Commands and results

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot04_wave2 ./stage1/cohere/lint/helpers/slot04_wave3_space -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot04_wave3_space/evidence/tests.log 2>&1
go vet ./... > stage1/cohere/lint/helpers/slot04_wave3_space/evidence/vet.log 2>&1
gofmt -l cmd internal stage1/cohere/lint/helpers/slot04_wave2 stage1/cohere/lint/helpers/slot04_wave3_space > stage1/cohere/lint/helpers/slot04_wave3_space/evidence/gofmt.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=10m > stage1/cohere/lint/helpers/slot04_wave3_space/evidence/oracle.log 2>&1
python3 stage1/cohere/lint/helpers/slot04_wave3_space/testdata/generate.py > stage1/cohere/lint/helpers/slot04_wave3_space/evidence/generate.log 2>&1
python3 stage1/cohere/lint/helpers/slot04_wave3_space/testdata/readiness.py > stage1/cohere/lint/helpers/slot04_wave3_space/evidence/readiness.log 2>&1
```

The two touched packages pass in 16.483s and 22.095s. The six filtered input oracle fixtures pass in 1.746s with zero cache hits and six probe misses. Vet and formatting logs are empty and successful. Tests require successful execution with no stderr and compare exact stdout, with native ASan/UBSan enabled. Test output goes to logs, never a pipeline. The first-batch helper code is unchanged and its prior 167.977s gate was not repeated. The full repository test gate was not run.

## Mutants and omission controls

Each new mutation compiled and executed without stderr in all three Adamic execution modes, then disagreed with actual Go. Compile failures and sanitizer crashes do not count as catches.

| Mutation | First catch in source Node, native and emitted JavaScript |
| --- | --- |
| isJavaScriptSpace BOM -> NEL | Line 109: blank false instead of true for BOM. |
| isBlank empty -> false | Line 1: blank false instead of true. |
| isBlank stop after first character | Line 268: blank true instead of false. |
| isValueSeparator drop colon | Line 150: byte 58 false instead of true. |
| isValueSeparator accept CR | Line 16: byte 13 true instead of false. |

The retained traversal package reran its three compiling mutants on source Node and native: leading edge changed (line 71), conditional false branch omitted (229), && left substituted for right (403). Its canonical-class consumer omission is caught. This batch separately removes every enforce-canonical-classes row and proves the six-consumer coverage detector fails.

Earlier in this turn, the now-yielded JSX/React helpers also passed a four-way comparison and six compiling mutants in 31.080s. Their code is not delivered or counted. Historical evidence is preserved only in evidence/yielded-tests.log: remove createReactClass (line 14), bypass React namespace (1370), omit callee-parenthesis traversal (660), continue after first matching nonliteral attribute (3571), omit entity unescape (2991), and collapse empty string to absent (2915). Source Node, native and emitted JavaScript all caught these changes. Removing font-display and did-mount consumers separately also failed their coverage check. The duplicate experiments do not clear any dependency in this report.

## Ownership and authentication corrections

The previous batch's 7904fda and 535bd31 were successfully pushed at the start of this turn after GitHub authentication recovered. The full refreshed claims then exposed slot 03's earlier text-scanner reservation c584145 at 01:03:44 UTC, before our 502d072 at 01:03:48. That duplicate implementation, its fixtures and its test functions are now removed from the final tree. The retained prior batch consists of classValuesUnder and collectClassValues, 22 edges across 11 consumers. Its old evidence remains historical and its report/readiness explicitly mark the withdrawal. The old authentication-blocked patch fallback is no longer needed.

Every helper branch and complete claim was refreshed before each reservation. All named helpers above six consumers were reserved; the generic seven-count strict-option label describes remaining per-rule decoding beyond already delivered helpers, not another named shared Go helper. Our first six-count reservation raced: slot 01 owns StringAttributeValue under 6ed891f at 01:19:38 UTC, and slot 02 owns IsEs5ComponentCall/isCreateClassName under 972f103 at 01:20:12, before our 2053731 at 01:20:22. These duplicates were yielded and excluded. Reservation 4df6533 was then pushed for the final three predicates before any code. Subsequent full wildcard fetches show these three are claimed only by slot 04. No fourth helper is reserved and no PR is opened.

## Setup timing

The existing session setup succeeded before the first batch; no rerun was necessary. nproc was checked again and is 5. Go 1.27.1, clang 20.1.8, Node 24.19.0. The original timing log is ../evidence/04/setup.log:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (2s)
setup: build cache warm (165s)
setup: done in 165s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```
