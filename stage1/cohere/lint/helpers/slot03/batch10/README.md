# URL and size predicates

Three independent `.a` files port actual private Go helpers. No dependencies, AST dispatch, shared registration or finding-model changes.

- `isURL(value)` requires exact lowercase `url(`, a final `)`, and at least five code units. Its interior excludes LF, CR, U+2028 and U+2029. Empty bodies, nested punctuation, Unicode, NUL, tabs, vertical tabs, form feeds, NEL, NBSP and BOM are accepted when the envelope is valid. No trimming or URL validation is added. Go rejects a line terminator after the final `)`, regardless of a JavaScript regexp's end-anchor behavior.
- `isAbsoluteSize(value)` matches exactly `xx-small`, `x-small`, `small`, `medium`, `large`, `x-large`, `xx-large`, `xxx-large`.
- `isRelativeSize(value)` matches exactly `larger` and `smaller`.

All inputs must be valid UTF-8-derived Unicode strings. Raw invalid UTF-8 and unpaired UTF-16 surrogates are outside this adapter. No predicate folds case, trims whitespace or normalizes Unicode.

The oracle calls the pinned Go functions through an owned temporary export overlay. Capture records all four consuming rules and 118 runtime fixtures before external installed-engine/corpus skips. Full sources, source-derived tokens, URL envelope/line-terminator controls and every accepted size keyword with case/whitespace/NUL/affix mutations yield 357 strings and 1,071 outputs. Real Go, source Node, emitted JavaScript and sanitized native must match byte for byte.

Five semantic mutants compile and exit 0 without stderr before comparison: URL case folding, admitting U+2028, dropping xxx-large, size case folding and dropping smaller. Compiler refusal, crashes and sanitizer failures are never credited as catches. The earlier digit/prefix implementations lost a one-second claim race and are not delivered or counted.

Run `source /workspace/adamic-tools/env.sh`, then `go test ./stage1/cohere/lint/helpers/slot03 -count=1 -v -timeout=20m`, redirecting output directly to a log. Regeneration is `python3 stage1/cohere/lint/helpers/slot03/batch10/testdata/regenerate.py`. The known upstream live Tailwind/corpus capture gate remains failed. Helper parity does not imply whole-rule findings, fixes or suggestions parity.
