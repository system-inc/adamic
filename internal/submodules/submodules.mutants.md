# Mutants for internal/submodules (#g7nnryf)

`internal/submodules` finds cohere's TypeScript submodule by its name in `.gitmodules` and the shim's directory by its
module in go.mod's replaces. Two mutants show its tests can fail.

## The entry chosen by its url, not its name

`Find`, old (this head): an entry matches when `submodule.Name == name`. New: it matches when `submodule.URL` names
it, as a reader that finds TypeScript by where it points would.

Run: `go test -count=1 ./internal/submodules/`, sealed, at the mutant's commit 1978dc70e2 (Loom job J-s0hcrvf).

Red, exit 1, three tests:
- `TestSubmodulePathFindsTypeScriptBeforeTheMove`
- `TestSubmodulePathFindsTypeScriptAfterTheMove`
- `TestSubmodulePathRefusesANameTwiceAtOneDepth`

## A name at two depths answered by the shallowest

`Find`, old (this head): every entry with the name is gathered across all depths, and more than one is refused. New:
the shallowest depth holding the name answers, as the first cut of this package did.

Run: the same, at the mutant's commit 7ad8d8555c (Loom job J-78a1nd0).

Red, exit 1, one test: `TestSubmodulePathRefusesANameAtTwoDepths`, at `gitmodules_test.go:271: a name at two depths
answers "cohere/TypeScript", want a refusal`.

Green at this head: Loom job J-0z7qyqk at e7326b7d6f9ab6f0e7169a3d12886ca8e11cd725, gofmt, go vet and all 11 tests,
then `internal/compilercensus` and `cmd/adamic-gate/parallelcensus`, `PASS`.
