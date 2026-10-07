# Slot 01 second batch

Three more helpers, one per `.a` file, against Go cohere `715ba94f3608a6500086b1076ce5cb7e51b836db`. Claim commits were pushed before code; implementation is `c986bcc`. Earlier delivered helpers remain documented in SLOT01_README.md.

| File | API | Caller contract |
|---|---|---|
| tailwind_new_class_literal_reader.a | tailwindNewClassLiteralReader(settings, compile) | Supply a Go regexp.Compile-equivalent compiler returning `{valid, pattern}`. Invalid patterns have valid=false and are skipped individually. Unsupported provider features must be refused explicitly. |
| tailwind_class_values_in.a | tailwindClassValuesIn(node, bound, cache, readClassValues) | Node indices identify one file arena, -1 is nil. Share one cache per bound reader/file/settings; supply the separately owned readClassValues dispatcher. |
| tailwind_class_literal_from.a | tailwindClassLiteralFrom(node, text, origin, tokenRange) | Supply decoded literal text, the same node identity, and actual rule.TokenRange in the parser's offset units. Loc includes trivia and cannot replace TokenRange. |

The factory snapshots attribute/callee names in independent deduplicated maps and preserves ordered successful compiled patterns, including duplicates. Each call creates fresh maps/lists. Native stage 0 cannot lower undefined/null values here, so `valuesBound=false` represents Go's nil unbound values map. The callback returns a tagged result instead of a nullable generic pattern. Regex syntax and execution belong to the supplied provider; this helper does not implement a regex engine or substitute JavaScript regex semantics for Go RE2 semantics.

Memoization reads afresh when unbound. When bound, it distinguishes absent entries from cached empty values using Map.has, caches nil nodes too, and returns the same lists for repeated identities. Returned lists are shared as in Go and callers must not mutate production results. The oracle deliberately mutates one list to expose sharing. A cache must never cross files or reader settings. TailwindClassValues is generic over literal and template payloads; both lists are preserved without copying.

Literal construction strips one unit at each end only when the token range is at least two units long. It retains node/text/origin and creates zero leading/trailing edges. It neither parses literals nor guesses offsets from decoded text. It accepts any origin string, as Go's string-backed origin type does.

Run from the repository root after sourcing /workspace/adamic-tools/env.sh:

```sh
go test ./stage1/cohere/lint/helpers -count=1 -v -timeout=20m > /tmp/lint-helpers-01-wave2.log 2>&1
```

The external Go overlay calls the real factory, memoizer and literal constructor. It reads all consumer test files listed in readiness.json and extracts statically evaluable complete string expressions; labels/messages are also included. Factory tests make each extracted source a configured name and pattern, then compare actual map state, compiled order and matches with Go-provided compiler results. Memoizer tests embed each extracted string in two distinct JSX attribute nodes, query repeated and nil identities in both binding modes, and observe cache size and result sharing. Literal tests parse the actual extracted source, visit every string/no-substitution-template node, and query five origins. Extra controls cover trivia, Unicode, escaped and unterminated literals and ranges shorter than two. Dynamically generated fixture cases and whole rule findings are not run.

The same .a source runs on Node and ASan/UBSan native. Eight semantic mutants must compile and exit successfully with empty stderr, then differ from Go. Compiler rejection, panic or sanitizer failure cannot count as a killed mutant. Reports, exact consumers/residual dependencies and logs are in SLOT01_WAVE2_REPORT.md, slot01_wave2_readiness.json and evidence/slot01-wave2/.
