// Package submodules finds where cohere keeps its TypeScript submodule and its TypeScript shims at run time, from the
// files that declare them, so a reader holds before and after cohere moves TypeScript/ to libraries/typescript/ and
// TypeScript-shim/ to libraries/typescript_shim/. A path spelled into a reader breaks on the day of the move; a path
// read from the declaration moves with it.
//
// The submodule is found in .gitmodules, by its name, walking into each submodule's own .gitmodules, since TypeScript
// is cohere's submodule and cohere is Adamic's. It is read as text, never through git: a Loom runner's source is
// unpacked with no .git (#cyasrr4), and git submodule there answers nothing. The name is what holds across the move:
// cohere's section stays [submodule "TypeScript"] while its path changes.
//
// The shims are a plain directory in cohere, not a submodule, so .gitmodules never names them. Adamic's go.mod does:
// each shim module is replaced by its directory, and the replace directive is where the build itself looks.
//
// It is a leaf, importing the standard library alone, so any reader in the tree can take it without taking anything
// else. internal/buildcache carries its own unexported go.mod reader (goModReplacements in settle.go), which this
// package's matches token for token; folding that one into this is a later change.
package submodules
