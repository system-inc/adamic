# Mutant evidence for namespaces/06_binary_expression_state.a's row moving from Refused to NotYet under rule G (#7fd51gk)

Run on Codex at f5cec762 (compiler/rule-g-4: rule G re-cut onto main c1e9379e, plus this row), go1.27.1 linux/amd64.

## Fixed (expected pass)
go test -count=1 ./stage3/fixtures/        exit 0, no FAIL lines.
Main c1e9379e, the same, exit 0. The isolated reruns of TestProduct_FixtureOracleHook and
TestHostConfigurationFixtureReaders pass at both shas (their whole-package timeouts were load).

## Mutant: rule G undone, internal/lower/invariance.go from main c1e9379e (expected red)
go test -count=1 -run 'TestFixturesNamespaces' -v ./stage3/fixtures/        exit 1
--- FAIL: TestFixturesNamespaces/namespaces/06_binary_expression_state.a/stage0
fixtures_test.go:493: gap changed: update status.json and check the native output; got {Outcome:Refused What:stage3/fixtures/namespaces/06_binary_expression_state.a:13:27: Adamic 0.1 refuses a function taking TState[] seen as one taking TState[] (tsc relates a method's parameters both ways), so it can be handed what it can't take; write the method as a property holding a function (handle: (animal: Animal) => void), which tsc checks one way, or take the wider type in the method (method-signature-style) Stdout:}; recorded {Outcome:NotYet What:stage3/fixtures/namespaces/06_binary_expression_state.a:8:34: stage 0 can't lower a value of type T yet Stdout:}

So the row holds rule G's move and nothing else: without rule G the identical-pair refusal comes back.
