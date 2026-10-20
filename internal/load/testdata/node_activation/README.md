# Node library activation (#2nznnt3)

The fixtures exercise the production checker: a lazy `require("fs")`, import
types without static imports, and a program with no Node use. The tests also
check that user package names, relative names, nonliteral arguments, property
calls, unknown `node:` names and `undici-types` do not rebuild with the Node root.

`TestProduct_NodeLibraryRequireMutant` declares the keyed `GoTestMutated`
product with the actual hook's require branch disabled. The mutant test runs
that product's `TestNodeLibraryLazyRequire` witness and requires it to fail
with TS2591. It runs the built binary without invoking Go or creating an overlay.
