# compiler_test.py's probe repository and cohere's TypeScript path (#g7nnryf)

twice-daily.sh sources `cloud/cohere-typescript-path.sh` and, in per-ref mode, reads cohere's TypeScript submodule
path from the repository's `cohere/.gitmodules`. `CompilerSelectionTests.setUp` lays both in its probe repository.
Each mutant below, planted in setUp, turns `TestCompilerPython` red:

1. **The reader not copied** (drop the `cloud/cohere-typescript-path.sh` write): every measure fails before its run
   is made, `twice-daily.sh: line 15: …/cloud/cohere-typescript-path.sh: No such file or directory`, exit 1 where
   `measure_red` wants 3. This is the red chg_tcm1nnb7 met at c558cf63 (Eve, 14:21Z Nov 25).
2. **No `cohere/.gitmodules`** (drop its write): the per-ref measures fail with `cannot read cohere's TypeScript
   submodule path from …/cohere/.gitmodules`, since the reader never falls back to a default path.
