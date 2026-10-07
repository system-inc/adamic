# Slot 03 fourth helper batch

Three independent Collapse helpers, one .a file each. isEscapeTerminator accepts a Go byte represented by a number and recognizes exactly TAB, LF, FF, CR and SPACE. isFollowedByWhitespace keeps Go's input/index/callback signature; input is unused upstream. The callback returns raw Go bytes at Go byte positions, with out-of-range behavior supplied by its caller. It peeks index+1 once, returns immediately for SPACE/LF/TAB, and peeks index+2 only after CR, accepting it only when LF follows. Call order is observable.

isIgnoredThemeKey keeps the exact five-namespace/fourteen-key Go table. A key is ignored only when it equals a configured key or starts with that key followed by a hyphen. Unknown namespaces, case differences and word continuations return false. No table from the browser or a different Tailwind release is substituted.

Run with the setup environment sourced:

```
python3 stage1/cohere/lint/helpers/slot03/batch4/testdata/regenerate.py > /tmp/slot03-batch4-capture.log 2>&1
go test ./stage1/cohere/lint/helpers/slot03 -run '^TestBatch4' -count=1 -v -timeout=20m > /tmp/slot03-batch4-tests.log 2>&1
```

The owned Go oracle overlay calls actual private cohere helpers without changing its source tree. Coverage records all six consumers and 157 captured runtime inputs, including live fixture strings captured before external-engine skips. Every source's raw UTF-8 byte positions feed lookahead controls; raw source strings and extracted custom-property tokens feed theme-key comparisons. Exhaustive byte and byte-pair controls independently cover all possible callback byte values. The suite compares Go output against source Node, emitted JavaScript and native under ASan/UBSan; each helper has a compiling semantic mutant. The upstream Tailwind rule-package gate fails on unavailable external installations/corpora, which is recorded rather than reported as passing.

These remove eighteen dependency edges, not complete rules or all remaining helper blockers. See REPORT.md and readiness.json for exact consumers and limits. Go byte-domain and valid UTF-8 string contracts are explicit; arbitrary out-of-domain numeric inputs, invalid UTF-8 namespace strings, complete CSS parsing and real rule findings/fixes/suggestions are not asserted.
