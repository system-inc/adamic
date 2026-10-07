# Regex scanner facts

`wave07-regex-structure\n<flags>\n<pattern>` uses the existing ABI-v1 inspect
entry and schema-1 UTF-16 frames. The suffix splits at its first two LF characters,
so pattern text can contain LF. The response contains scanner success, matched
character start/end pairs and escape/class scanner boundary records. All spans
are UTF-16 offsets into the supplied pattern, not file bytes. Invalid boundaries
are marked false, and the native decoder validates every span and frame.

The helper exposes raw regular-expression structure only. It does not return
rewritability, flag validity, quantifier balance, findings, messages or edits.
Adamic makes those decisions and builds the complete suggestion. The generic
scanner copy and its tests are attributed under bridge/tsgo/wave07_regex.

Production registration remains pending. The private validation overlay inserts
this one route into bridge/tsgo/checker/facts.go:

```go
case "wave07-regex-structure": return p.wave07RegexStructure(node, question)
```

No shared dispatcher, existing test harness or generator was edited. The normal
archive refuses this question with exit 70, as the question gate measures. The
rest-parameter and Promise-rejection ports use already registered
node-symbol-details metadata and require no new question.
