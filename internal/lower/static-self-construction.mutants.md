# Mutants for a class constructing itself in its static initializer (#1451npc)

Each is one edit to class_static.go, run on a house box, then put back.

| Mutant | Old text | New text | Red |
|---|---|---|---|
| Construction always admitted | `failure = l.staticConstructionReads(construction, declaration, available, seen)` | `failure = nil` | TestStaticSelfConstructionStops, every row but "a class with a base" lowers where it must stop; each later-static row then reads an uninitialized static where Node reads undefined |
| No instance accessor reach | `if !current {` (the instance accessor walk) | `if false {` | TestStaticSelfConstructionStops, "a later static read by an instance getter the constructor reads" lowers |
| Parameter defaults skipped | `for _, parameter := range implementation.Parameters() {` in staticConstructionReads | `for _, parameter := range []*ast.Node{} {` | TestStaticSelfConstructionStops, "a later static read by a parameter default" lowers |
| A base's reach assumed | `construction != nil && !extendsClass(declaration)` | `construction != nil` | TestStaticSelfConstructionStops, "a class with a base" lowers |
| No symbol check on the library member | staticLibraryCallsNothing's body after its first `callee = ...` line | `return callee.Kind == ast.KindPropertyAccessExpression && callee.Name().Text() == "captureStackTrace"` | TestStaticWalkReachAdmissions, "a program object's captureStackTrace" lowers, running program code the walk never followed |
| No per-edge check | the first line of staticParameterCalls' body | `return nil` | TestStaticWalkReachAdmissions, "assert reached with a static method reading a later static" lowers |
| Parameter calls never known | `if judged && target.Kind == ast.KindParameter` | `if false && target.Kind == ast.KindParameter` | TestStaticWalkReachAdmissions, "assert reached with no verboseDebugInfo" stops at the indirect call; TestTscVersionZeroPassesTheStaticWalk stops at 67:111, the indirect call, not 73:2 |
