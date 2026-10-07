Built: classValuesUnder, collectClassValues and UnescapeStringLiteralText, one helper per .a file.
Implementation: 7904fda; claims 0582ef8 and replacement 502d072 were pushed before code.
Checks: isolated helper package PASS 23.272s; uncached input oracle PASS 1.430s; vet/format clean; prior setup 165s, nproc 5.
Mutants: false outer edge changed, false branch dropped, && side swapped, decoder failure ignored, scan advanced too far; all compiled and were caught on Node and sanitized native; two consumer omissions caught.
Limits: push blocked by GitHub credentials; no shared harness/registration edits, full repository gate, emitted-JavaScript comparison, direct AST integration or external dependency ports.

## Readiness and consumers

This continuation removes 31 dependency edges across 20 distinct consumers: 11 each for the two Tailwind helpers and nine for text unescaping. Zero rules lose their final helper blocker from this continuation alone. Across both slot 04 batches, the frozen original baseline remains 46 -> 47 ready, solely adding @next/next/no-img-element. Existing ready entries in the frozen ledger are excluded from the new-ready count. Other workers' deliveries are deliberately not assumed integrated here. The original readiness.json is unchanged; this directory's readiness.json records every residual dependency.

### rules/tailwind.classValuesUnder

11 dependency removals; zero final blockers removed alone.

- `better-tailwindcss/enforce-canonical-classes`
- `better-tailwindcss/enforce-consistent-class-order`
- `better-tailwindcss/enforce-consistent-important-position`
- `better-tailwindcss/enforce-consistent-variable-syntax`
- `better-tailwindcss/enforce-consistent-variant-order`
- `better-tailwindcss/enforce-shorthand-classes`
- `better-tailwindcss/no-conflicting-classes`
- `better-tailwindcss/no-deprecated-classes`
- `better-tailwindcss/no-duplicate-classes`
- `better-tailwindcss/no-unknown-classes`
- `better-tailwindcss/no-unnecessary-whitespace`

### rules/tailwind.collectClassValues

11 dependency removals; zero final blockers removed alone.

- `better-tailwindcss/enforce-canonical-classes`
- `better-tailwindcss/enforce-consistent-class-order`
- `better-tailwindcss/enforce-consistent-important-position`
- `better-tailwindcss/enforce-consistent-variable-syntax`
- `better-tailwindcss/enforce-consistent-variant-order`
- `better-tailwindcss/enforce-shorthand-classes`
- `better-tailwindcss/no-conflicting-classes`
- `better-tailwindcss/no-deprecated-classes`
- `better-tailwindcss/no-duplicate-classes`
- `better-tailwindcss/no-unknown-classes`
- `better-tailwindcss/no-unnecessary-whitespace`

### ecmascript/text.UnescapeStringLiteralText

9 dependency removals; zero final blockers removed alone.

- `@next/next/google-font-display`
- `@next/next/google-font-preconnect`
- `@next/next/next-script-for-ga`
- `@next/next/no-before-interactive-script-outside-document`
- `@next/next/no-css-tags`
- `@next/next/no-html-link-for-pages`
- `@next/next/no-page-custom-font`
- `@next/next/no-unwanted-polyfillio`
- `structure/boundary-no-project-theme-value`

## API and preserved behavior

`collectClassValues` takes a stable-ID arena, node ID (-1 is nil), origin, edges, a mutable accumulator and dependency callbacks. It appends without clearing prior results. Literals retain node identity, decoded text, content range, origin and edges. Templates retain identity, origin and edges. Parenthesized/JSX/as/satisfies wrappers recurse. Conditionals visit true then false value branches, never the condition. && reads only the right side; || and ?? read both in order. Arrays preserve element order. Other expressions are not entered. Template records are appended before holes, with each valid hole's edges supplied by the independently owned holeEdges dependency. The adapter encodes missing template heads/spans as no traversable spans while retaining the template record.

`classValuesUnder` allocates fresh literal/template arrays and delegates once with false outer edges. The checks mutate a returned literal array and compare a second fresh reading.

`unescapeStringLiteralText` scans the first semicolon after each ampersand, preserves too-short/incomplete or rejected references, and replaces successful references in a single pass. A replacement ampersand is never decoded again. decodeEntity is an explicit callback owned by another worker; the real Go callback supplies oracle adapter answers. Unknown names, numeric overflow, surrogate replacement and the 253 named entities therefore follow Go rather than a browser's different table. The supported input boundary is valid decoded Unicode text, as used by parser strings and fixture sources; malformed raw UTF-8 Go strings are not covered.

All code and new tests live under slot04_wave2/, apart from the explicitly required claims/04.md update. Existing shared harness, generator and protected compiler files remain untouched. The existing options_json.ts parser is imported; no new .ts file is delivered. All new Adamic source files use .a.

## Observations

Go cohere pin: 715ba94f3608a6500086b1076ce5cb7e51b836db. Oracle-only Go overlay exports call the actual private helpers without replacing their behavior. The parser adapter reads raw named AST fields independently of traversal. classLiteralFrom content payload and holeEdges answers are supplied by actual Go dependency calls, keeping those separately owned ports outside this delivery.

- Traversal consumers: 278 captured Tailwind inputs cover every one of the 11 consumers, producing 22,374 comparison lines. Thirteen shape controls produce 1,386 lines. Each node, plus nil, is exercised with all four edge combinations, rotating all three origins and seeding a preexisting literal.
- Text consumers: 248 captured source inputs cover all nine consumers. Each whole source plus each parsed StringLiteral and JsxText field is checked, giving 783 text values. The oracle uses AsJsxText().Text, since Node.Text() refuses JsxText.
- Text controls: 683 inputs include all 253 upstream named entities, numeric, unknown, rejected, nested/adjacent references, Unicode, missing semicolons and 400 deterministic random mixtures.
- Exact stdout agrees among actual Go, Node source and native built with ASan/UBSan. Every run must exit successfully without stderr. All output goes to files.

The source corpora are pinned copies/subsets of the first batch's actual asserted fixture capture, documented by ../slot04_REPORT.md and ../testdata/slot04/capture.py. The text subset is selected by frozen readiness consumers. The prior Tailwind capture used a pinned temporary tailwindcss@4.3.3 fixture and excluded external live-repository population tests. Neither this continuation nor the first batch asserts those live-repository gates pass.

## Commands and evidence

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot04_wave2 -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot04_wave2/evidence/tests.log 2>&1
go vet ./... > stage1/cohere/lint/helpers/slot04_wave2/evidence/vet.log 2>&1
gofmt -l cmd internal stage1/cohere/lint/helpers/slot04_wave2 > stage1/cohere/lint/helpers/slot04_wave2/evidence/gofmt.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=10m > stage1/cohere/lint/helpers/slot04_wave2/evidence/oracle.log 2>&1
```

Complete touched package: PASS 23.272s. The six input oracle fixtures pass in 1.430s with six probe misses and zero cache hits. Vet and gofmt produce empty successful logs. The prior first-batch helper gate passed in 167.977s; it was not rerun because no first-batch code changed. The whole repository gate and emitted-JavaScript backend were not run in this continuation.

## Compiling mutants

| Mutation | First observed catch on both Node and sanitized native |
| --- | --- |
| classValuesUnder leading edge false -> true | Line 71, literal edge true instead of false. |
| collectClassValues omit conditional false branch | Line 229, node 16 literal absent. |
| collectClassValues && right -> left | Line 403, node 7 instead of node 9. |
| UnescapeStringLiteralText ignore decoder failure | Line 5, empty output instead of bytes 38,38,59 for &&;. |
| UnescapeStringLiteralText semicolon + 1 -> + 2 | Line 8, &t; instead of &lt;. |

These mutations compile and execute successfully; sanitizer/compiler failures do not count as catches. Removing every canonical-class consumer row makes the Tailwind coverage detector fail. Removing every google-font-display consumer row makes the text coverage detector fail. Both omission controls are separate from compiling semantic mutants.

## Claims and delivery failure

Every origin codex/lint-helpers* branch was wildcard-fetched and its complete claim read before reservation. The first continuation reservation 0582ef8 covered three tied 11-consumer helpers. A later refresh exposed an earlier attributeValues reservation: slot 01 ad07122 at 00:59:17 UTC precedes our 0582ef8 at 00:59:43 UTC. The local duplicate was removed from delivery, and no attribute helper or mutant is counted. classValuesUnder and collectClassValues remain uniquely reserved here. All 10-consumer helpers are reserved by slot 02; decodeEntity and hexValue are reserved by slots 01/03, leaving UnescapeStringLiteralText as the highest unclaimed nine-consumer helper. Its replacement reservation 502d072 was successfully pushed before code.

Final refresh and delivery push became blocked by GitHub authentication. Normal Git reports `fatal: could not read Username for 'https://github.com': No such device or address`. Retrying with the configured GitHub CLI credential helper reports `remote: Invalid username or token. Password authentication is not supported for Git operations.` and `fatal: Authentication failed for 'https://github.com/system-inc/adamic.git/'`. No credential values are printed. Cloud runtime status shows no configured secrets or outbound identities. This is an authentication failure, not an automatic approval rejection. The claim pushes succeeded earlier, but implementation/report commits remain local. The requested fallback is /workspace/lint-helpers-04-wave2.patch, generated from 502d072..HEAD. Apply it on top of the already-pushed claims; no further helper is claimed and no pull request is opened. The final all-branch ownership refresh could not be completed after credentials failed.

## Toolchain

The existing session setup succeeded before the first batch; no setup rerun was necessary. nproc is 5. Go 1.27.1, clang 20.1.8, Node 24.19.0, environment /workspace/adamic-tools/env.sh. Full original setup output is preserved in ../evidence/04/setup.log:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (2s)
setup: build cache warm (165s)
setup: done in 165s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```
