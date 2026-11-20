# Scoped tail recording mutants

TestLatentScopedTailEntryFindingLoom03dg pins how the scoped tail records a stop, using the stop the census itself
injects (`LATENT_TAIL_TEST_STOP=forwarderValues@probe`, scoped_tail.go.txt), so no production refusal can move it.
Each mutant below is applied to the census overlay's source and turns that test red, at the line named.

| Mutant | Edit | Red at |
| --- | --- | --- |
| The injection never fires | `latentTailTestStop` returns false | "probe ledger lacks the forwarderValues stop" |
| The stop is not published | `latentFinishUnit` drops `latentFindings = append(latentFindings, *finding)` | "no scoped tail finding for probe" |
| Steps after a stop still run | `latentScopedTail` no longer returns at the first stop (`return ledger, &value` removed) | "a step after the stop ran" |
