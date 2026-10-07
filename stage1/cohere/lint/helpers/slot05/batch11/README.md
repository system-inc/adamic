# Eleventh batch

classAtomWrite returns a set atom's text unchanged, formats a range as literalRune(lo) + '-' + literalRune(hi), and formats every other kind from lo alone. The injected literalRune dependency belongs to another worker. All 256 Go uint8 kinds are compared with actual Go, as are signed rune boundaries, every rune captured from the four consumers and every captured string as set text. The integer kind adapter expects 0 through 255 and rune arguments in signed int32 range.

wordCharacters calls wordClassAtoms once with the unchanged options, then writeClass once with the exact returned array, negation false and all four rewrite flags false. Both dependencies remain externally owned. Actual Go supplies the dependency results. The Go overlay traces the real calls and checks input flags, output flags, negation and array identity across all sixteen flag combinations. The Adamic test proxies return Go's already computed values; this gate proves delegation, not the implementations of those dependencies.

expandsOnUppercase recognizes exactly the three Greek inclusive ranges and three singleton rune values in Go. Every integer from zero through 0x10ffff, including surrogates, plus signed int32 extremes, boundary neighbors and all consumer runes is compared. No regex matcher, listener, dispatch or finding-position code is added.

Each baseline runs against the private Go functions, source Node, emitted JavaScript and sanitized native. The overlay instruments calls without changing Go algorithms or returned bytes; cohere itself remains untouched. Each helper has compiling semantic mutants compared with Go output. Corpora are generated from all nonempty string literals in each consuming Go test file, including sources, options and expected messages. This is not a replay of whole rule diagnostics.

```
source /workspace/adamic-tools/env.sh
ADAMIC_SLOT05_BATCH11_EVIDENCE="$PWD/stage1/cohere/lint/helpers/slot05/batch11/evidence" go test ./stage1/cohere/lint/helpers/slot05/batch11 -count=1 -v -timeout=20m > /tmp/lint05-batch11-helpers.log 2>&1
```

Shared registration and harness files and protected compiler files are untouched. Invalid UTF-8 input, noninteger kind/rune arguments, whole-rule findings and integration are outside this gate. CONSUMERS.md and readiness.json identify every consumer and remaining blocker.
