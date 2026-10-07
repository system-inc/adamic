Built: addWhitespaceAroundMathOperators in .a, preserving actual Go UTF-8-byte-to-rune expansion on the scanner path.
Commits: claim 89afa076 pushed before implementation; delivery SHA is in the final response.
Checks: owned differential and mutant package PASS 7.253s; 49,162 inputs, six consumers, 873,758 identical bytes on each of Node, emitted JavaScript and ASan/UBSan native; vet clean; setup 76s, nproc 5.
Mutant: comma_spacing_disabled compiles and exits successfully with empty stderr on all three runtimes, then fails only output comparison against real Go.
Limits: valid Unicode text, bounded controls and all six consumers' test-file string literals; no end-to-end Adamic rule findings, arbitrary invalid UTF-8, or full repository gate.

The frozen readiness list assigns this helper to:

- better-tailwindcss/enforce-canonical-classes
- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-consistent-variant-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

This removes six dependency entries. Each consumer retains other blockers, so zero rules become fully helper-ready from this helper alone. The shared readiness.json is left untouched.

Selection inspected all 397 fetched origin refs and twelve distinct helper claims, including short names in claim prose. The already delivered comments bundle is reserved separately in HELPERS.md. All higher concrete-symbol fan-outs were reserved; this helper ties the highest unclaimed count at six.

The corpus generator uses Go's parser and strconv.Unquote to extract 959 actual string literals from all six consuming test files. Original literals and calc-wrapped variants are compared, plus Cartesian controls across every math function, unknown/uppercase function names, operators, units, signs, exponents, nested functions, non-ASCII spaces and supplementary Unicode. These are source-derived helper inputs, not a measurement of runtime helper calls or a replay of final rule findings. The Go oracle imports the unchanged private implementation through an added export in a scratch overlay; no Go implementation is copied or rewritten.

The first differential run failed: Go's string(byte) expands non-ASCII bytes rather than copying their UTF-8 encoding. The fixed Adamic helper reads utf8At and reconstructs the same byte-valued rune stream only after Go's early math-name bail. For example a scanner-path emoji expands to four Latin-1 code points. This behavior is observed in Go and is intentionally preserved. The failed run and final successful run are retained separately; mutant results before a passing baseline are not credited.

Commands, each with output sent directly to the linked evidence logs:

```
ADAMIC_TOOLS=/workspace/adamic-tools bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/lint/helpers/from-wave1-10/testdata/generate.py
go test -count=1 -v -timeout 15m ./stage1/cohere/lint/helpers/from-wave1-10
go vet ./stage1/cohere/lint/helpers/from-wave1-10
```

See [final comparison](evidence/math-tests-final.log.txt), [first mismatch](evidence/math-tests-first.log.txt), [setup](evidence/setup.log.txt), [generation](evidence/generate.log.txt) and [vet](evidence/vet.log.txt).
