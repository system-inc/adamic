# The darwin key's mutants

`internal/acheck`'s `darwin` lets a program outside a package go unchecked on a box without Apple's SDK only when both
keys agree: `adamic c` stopped with `platform.MissingExitCode`, and the program's own import graph, parsed by
`internal/appleimports`, reaches an `apple/` module (Adamic's ruling on #w16eq99, applying the admission delta's
rules 5 and 6 from #an8zcp1). `TestADarwinProgramIsKnownByBothKeys` goes red under each mutant below, each planted
alone at `darwin` or `appleimports.Graph.Reaches`:

1. **The import key dropped:** `darwin` returns `code == platform.MissingExitCode` without asking the graph. Red on
   "the platform stop without an apple/ import".
2. **The exit key dropped:** `darwin` asks the graph whatever the exit. Red on "an apple/ import and another failure"
   and "an apple/ import and a clean check".
3. **The walk cut to the program's own imports:** `Reaches` stops at the first file. Red on "apple/ reached through a
   relative import".
4. **A skip with no darwin verdict:** the record's side is the gate's, not this test's, since the record is keyed by
   the tip that holds this file and lives on the tools line. `cloud/fast-gate/run.py`'s a-check phase reads every
   darwin `.a` from the promoted record and fails a path the record carries no `adamic c` run for ("the darwin leg
   <record> recorded no adamic c for this path"), or one whose recorded run is itself a platform stop. Its test is
   `run_test.py`'s `aCheck` cases on the tools line.

A program that imports `apple/` only inside a comment or a string is never darwin's, because the parse reads import
syntax, not text. It's planted as `examples/plain.a` in the same test.
