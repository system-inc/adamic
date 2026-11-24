Mutant evidence for 7_postfix_property.ts's recorded stop in gaps_test.go's representationProbes (#f3n0gyp).

What moved: the postfix value of a numeric local lowers since #f3n0gyp (internal/lower/postfix_value.go's updateValue). A property target like `counter.value++` keeps its stop, now in updateValue's words, "a postfix update value whose target is not a plain name", where it said "a PostfixUnaryExpression". The gap stays open; only its words moved.

Mutant: updateValue's plain-name check is dropped, so the property target reaches the local path.
Command: go test -count=1 -v -run '^TestParserRepresentationProbes$' ./stage1/cohere/markdownblocks
Expected with the mutant: red on 7_postfix_property.ts, "NotYet gap changed". Restored, it passes. The run is named in the commit.
