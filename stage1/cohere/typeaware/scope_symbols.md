# Raw scope symbols

`scope-symbols` accepts an exact `LabeledStatement` and no question suffix.
The Go implementation is `bridge/tsgo/checker/scope_symbols.go`; the owned
native decoder is `scope_symbols.a`. Its only shared registration is a new
switch case in checker/facts.go. No generator or shared harness was changed.

All fields use the existing decimal UTF-16-length framing:

1. Version `1`.
2. Mode `scope-symbols`.
3. Number of symbols returned by `GetSymbolsInScope(node, SymbolFlagsValue)`.
4. For each symbol, its UTF-8-safe display name and unsigned flags.

The order is the checker's order and is not promised stable. Native validates
every field and chooses whether the label name occurs in the returned names.
Go supplies no label-clash judgment. The question leases the file's checker,
returns an owned C result through the existing ABI, and retains no symbol pointer
in native code. Its identities have no lifetime beyond the program because it
exposes no identities at all.

`TestScopeSymbolsAtTheExactLabel` compares names and flags to independent direct
checker calls, including a Unicode binding and an inner block. It refuses wrong
kinds and question suffixes. A value-to-variable meaning mutant compiles and
exits 0 with empty stderr, but loses function/class label findings and fails the
production Go byte comparison. The released-program probe and registry mutant
exercise this question through native Adamic.
