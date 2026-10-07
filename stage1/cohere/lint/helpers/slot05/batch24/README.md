# Slot 05 batch 24

Three separate Adamic files implement Go cohere's `regexp.rewrite`, `regexsyntax.IsHexDigit` and `regexsyntax.AllHexDigits`. They do not register rules or replace the shared harness.

`isHexDigit(byte)` accepts a Go byte projected to a number in 0..255. `allHexDigits(bytes)` inspects an immutable UTF-8 byte array, rejects empty input, and checks every byte through the first helper. Passing JavaScript UTF-16 code units would violate this boundary.

`rewrite(source, options, dependency)` translates ECMAScript regex syntax to the syntax used by upstream regexp2. Source and result text are reversible lowercase hex encodings of Go byte strings. This preserves byte slices, invalid UTF-8 and exact output bytes without treating Go offsets as UTF-16 positions. Option bits are Unicode 1, ignoreCase 2, multiline 4 and dotAll 8. The result preserves rewritten bytes, the exactness bit and the error; an error returns empty output and false exactness. Error text is also byte encoded. This is a syntax translation helper, not a matcher.

The dependency callback supplies UTF-8 rune decoding/encoding, capture counting, named references/openers, specialized escapes, case widening, literal escaping, assertion quantifiers and boundaries, class scanning/construction/writing, modifier parsing and group checks. Atoms travel as an opaque dependency-owned byte string. `Reply` carries the relevant dependency results; unused fields take their zero values. Escape kind values are rune 0, set 1, assertion 2 and backreference 3. Set values are other 0, word 1, nonword 2 and property 3. Capture count and named-group presence are encoded as `groups * 2 + named`. This helper owns ordered dispatch, error propagation, exactness, output accumulation and the modifier option stack. It preserves upstream fallback exactness rather than promising an exact matching engine for every syntax.

The private Go overlay executes the original rewrite body with only dependency call names replaced. Dependencies execute their actual original Go implementations. The comparator checks their arguments and call order beside the result. The private adapter fingerprints byte arguments with two modular hashes and length; the Go generator rejects a collision between distinct observed inputs. A mutant-only unknown request receives a neutral reply with positive rune progress while retaining its wrong trace. Compilation failure, stderr, panic, sanitizer failure or unsuccessful execution never counts as a caught semantic mutant.

Corpus generation reads every string literal, including empty strings, from every frozen consumer's upstream Go test files. These strings include programs, options and prose; they are helper inputs, not a replay of complete rule findings. Rewrite exercises all sixteen option combinations for every source. Boundary additions cover every single byte, every byte between `a` and `F`, invalid UTF-8, named references, modifier nesting, assertions and classes. Both hex consumers' complete captured strings are retained; direct classification additionally covers all 256 bytes. Identical oracle dependency records are pooled, with every row retained. Go's case-extra iteration can change class-member order across fresh processes; every baseline and mutant is compared to the exact result from its own generation.

With `source /workspace/adamic-tools/env.sh`:

```
ADAMIC_SLOT05_BATCH24_EVIDENCE="$PWD/stage1/cohere/lint/helpers/slot05/batch24/evidence" ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch24 -count=1 -v -timeout=20m > /tmp/lint05-batch24-complete.log 2>&1
```

The suite requires Go cohere pin `715ba94f3608a6500086b1076ce5cb7e51b836db`, source Node, emitted JavaScript on Node, sanitized native and successfully executing semantic mutants. See REPORT.md and evidence for the observed results. Compilation/matching, finding positions, rule integration, complete fixes/suggestions, arbitrary inputs and the full repository gate remain outside this bounded comparison.
