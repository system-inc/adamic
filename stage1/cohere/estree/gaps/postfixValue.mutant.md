Mutant evidence for TestPostfixValueLowers (stage1/cohere/estree/gaps_test.go), which replaces TestPostfixValueGap now that the gap is closed (#f3n0gyp).

Mutant: updateValue (internal/lower/postfix_value.go) returns its first NotYet unconditionally, the postfix value stopping again.
Command: go test -count=1 -v -run '^TestPostfixValueLowers$' ./stage1/cohere/estree
Expected with the mutant: red, "the postfix value gap opened again". Restored, it passes. The run is named in the commit.
