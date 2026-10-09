# u025 audit plan, before mutants
Base: 7b18d0576930caca4e22ce2eef92fcf563af52d0 (origin/main).
Code under test: Adamic Load, loader configuration, filesystem adaptation, regexp declaration rewrites, diagnostic formatting. Oracle: self-written nonempty CheckError requirement, not diagnostic identity.
Pin row code under test: repository submodule construction. Oracle: actual git rev-parse HEAD compared with recorded pin; external-run plus self-recorded expected hash. No production load function reached.
Reached load package functions on regexp path: Load, load, compilerOptions, rootFileName, usesNodeModules; sourceFS.adamicFile, displayName, FileExists, ReadFile, DirectoryExists, Stat, Realpath; regexpLibraryFS.ReadFile; Program.diagnostics, formatDiagnostic, lineAndColumn; starCollision (early return), writeChain (including recursion when chains exist). External checker and inherited vfs methods are dependencies, not mutation targets. No Node imports, overlay writes, declarations, or filesystem mutation methods on these inputs.
Fixed menu: M1 drop exec-array rewrite (regexp_library.go:20); M2 drop match-array rewrite (:21); M3 drop RegExp split rewrite (:22); M4 Strict true to false (load.go:59); M5 diagnostic code to 0 (load.go:253); M6 column +1 to +0 (load.go:278).
P1: Load returns nil,nil at entry (load.go:80). Only regexp row uses this entry; pin vacuity null.
S1: change TypeScript checkout construction to another commit, preserving test and Git oracle. Setup-check evidence, never production kill.
