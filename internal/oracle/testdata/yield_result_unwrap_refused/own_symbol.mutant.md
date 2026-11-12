# Match only the name (#2p41tb1)

In `containerYieldResultPair`, remove the same-symbol and `load.IsLibrary`
conditions, retaining the `IteratorYieldResult` name match (remove the unused
load import). This admits the program's own generic interface to the shortcut.

`go test ./internal/oracle -run '^TestYieldResultUnwrapRefused/own_symbol$' -count=1`
is red: `TestYieldResultUnwrapRefused/own_symbol` receives nil instead of the
exact container-fit NotYet. The ordinary walk stops at its generic depth bound;
the name-only mutant unwraps this interface and lowers. Node prints `1`.
