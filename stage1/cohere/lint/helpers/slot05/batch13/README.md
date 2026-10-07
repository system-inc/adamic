# Thirteenth batch

Canonicalize delegates to simpleFold under Unicode mode. Otherwise it leaves supplementary runes unchanged, preserves the full-uppercase-expansion exceptions, and prevents non-ASCII characters from uppercasing into ASCII. The Unicode uppercase operation, simpleFold and expandsOnUppercase remain explicit function dependencies. Their results come from real Go in the private test driver, using generated sparse numeric maps and expansion entries. No standard-library mapping is guessed or reimplemented here.

CaseClass delegates to CaseEquivalents once with unchanged rune and mode. Nil (represented as undefined) and empty results decline with empty text and widened=false. Otherwise it preserves member order, escapes each member through the externally owned EscapeClassRune, wraps the text in brackets and reports widened=true. Actual Go supplies the membership aliases and every escape string.

writeClass preserves Go's empty positive (?!) and negative [\s\S] forms. Nonempty inputs format atoms in order, append caseExtras only when ignoreCase is true, pass the unchanged Unicode flag and shared atom array, then wrap in positive or negative brackets. Atom formatting and caseExtras remain explicit dependencies. The private driver supplies their actual Go outputs and checks invocation order, values and array identity against instrumented Go.

The overlay instruments only dependency call names inside the actual functions. It changes no production cohere worktree or algorithm. Baselines compare actual Go, source Node, emitted JavaScript and ASan/UBSan native, requiring exit zero and empty stderr. Semantic mutants must compile and run normally before differing from Go.

Every consuming Go test file contributes nonempty literals, including source, options and expected text. The corpus has 904 distinct strings and 112 signed-rune controls. Canonicalize and CaseClass each sweep every integer zero through 0x10ffff in both modes, including surrogate integers and signed int32 extremes. writeClass covers empty and mixed atom lists, all 256 uint8 kinds, consumer runes and set text, positive/negative forms and all rewrite-flag combinations on the kind controls. Go's outer case-table iteration order may vary across processes; corpus hashes document each observed capture rather than promise byte-for-byte regeneration.

```
source /workspace/adamic-tools/env.sh
ADAMIC_SLOT05_BATCH13_EVIDENCE="$PWD/stage1/cohere/lint/helpers/slot05/batch13/evidence" go test ./stage1/cohere/lint/helpers/slot05/batch13 -count=1 -v -timeout=20m > /tmp/lint05-batch13-helpers.log 2>&1
```

No matcher, rule listener, finding-position conversion, shared harness, registration generator or protected compiler edit is added. Not covered: dependency implementations, arbitrary callback values or table shapes, noninteger or out-of-int32 rune arguments, out-of-uint8 atom kinds, invalid UTF-8 and whole-rule diagnostics/integration. Earlier simpleFold ownership was overlooked, then withdrawn before any duplicate code; the two retained helpers were pushed before claiming writeClass.
