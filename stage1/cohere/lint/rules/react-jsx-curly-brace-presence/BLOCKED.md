# Blocked: variadic Array.push lowering

The draft is not registered or certified. `rule.a.txt` and `rule.json.txt` preserve
work without adding an uncompilable rule to shared discovery.

Adamic rejects `Array.push` with more than one value. The compiler reports
`rule.a:33:58: stage 0 can't lower push with other than one value yet`.
The implementation rejection is `internal/lower/object.go:1133`.
The affected upstream symbol is `jsxCurlyBracePresenceFindCharacterReferences`,
`cohere/internal/lint/rules/react/jsx_curly_brace_presence.go:958` (the reference
range append). The draft represents each range as consecutive start/end numbers
and appends both in one call. This is a limitation encountered in this chosen
representation, not a claim that the upstream algorithm requires variadic push.
The user explicitly required stopping on unsupported language features without
rewriting around them, so no workaround was attempted.

Registry generation succeeded. Selected tests are in `evidence/selected.log`:
TestRulesAgree failed compilation; TestMutants/empty-string-is-whitespace failed
before compilation because the oracle's strict witness decoder rejected the
inherited all-rule options key `allowLoop`. That adapter issue also remains open.
No upstream cases or witnesses are certified, and no mutant disagreement is
claimed. The whole lint package was not run after the language-gap stop.
Selected package wall time: 6.752s; pass 0, fail 3 (including the mutant subtest),
skip 0. Host nproc: 5; load at stop: 2.39 3.75 1.95.

Five draft witnesses are in `testdata/`; options sidecars are preserved.
The Go rule body and checkout were unchanged; `oracle.go` calls upstream.
Go's string/template quote asymmetry and element-valued prop oscillation are
preserved in the draft judgment; neither is independently verified yet.
