# JSON objects main-based landing blocked

The landing attempt starts at origin/main 5e33a17b186a8a2218d27b69b21e2de5acc5b750 on runtime/json-objects-main. No compiler or runtime changes were imported. The language lead explicitly authorized stopping and listing unlanded prerequisites rather than importing area/runtime.

The requested first-parent patch is `git diff 1ba69df3^1 1ba69df3`. `git apply --check` exits 1 on main. The complete output is in json-objects-main-apply-check.txt beside this report. Context failures alone do not prove a semantic prerequisite; the following symbol and history checks do.

* 17b5a053a19b40cad6d08edf2f2aadb6564e865c, runtime: record scalar field kinds in object shapes. The new json_metadata.c validator reads shape->kinds and uses adamic_field_kind, adamic_field_boolean and adamic_field_number. Main's adamic_shape contains count, names, references and methods only, with no field-kind enum. The patch changes emitter initialization from an existing five-field shape to a seven-field shape, also assuming this predecessor. `git merge-base --is-ancestor 17b5a053 origin/main` exits 1. This is the first confirmed blocker to the object metadata portion.
* 2724dabd, Carry unproven cyclic ownership through graph regions. First-parent patch context in emit_expressions.go uses adoptGraph and GraphTypes; reuse.go uses graphUnique. `git log -S` identifies this commit as introducing adoptGraph and graphUnique. Those paths cannot be copied from the area as part of this unit. A smaller port would have to reconstruct descriptor propagation on main's existing allocation paths.
* 23928322, Add ParallelMap IR and sequential JavaScript and native ABI lowering. The requested patch modifies internal/native/parallel.go and runtime/parallel.c, neither of which exists on main. `git log -S` identifies this commit as introducing parallelMap. Those changes must be dropped for a smaller main port. This is a dependency of the submitted parallel portion, not a claim that ordinary JSON requires parallel support.

This list records confirmed dependencies of the submitted patch, not a minimal prerequisite closure for importing the runtime area. No typed-array dependency is claimed proven. No prerequisite was imported, and library efe905da/cf2cb58b and joining 69ae4344 were not applied after this stop.

No new JSON object, array or union cases landed on main. Every existing refusal and its message remains byte-identical because no lowering file changed. In particular object references retain `JSON.stringify object references (structural types can hide fields and toJSON; runtime shapes need complete value metadata)` and container unions retain `JSON.stringify a union containing containers without runtime element metadata`.

The previous runtime/json-objects-joined branch remains pushed, containing 69ae4344 and the subsequent 0ab45ee0. Its evidence is separate from a main landing. Native JSON tests and lower JSON tests passed there; the last oracle JSON run failed on a contract fixture's possibly undefined console.log argument and a void hook returning a value. It is not a complete green proof. Its four requested new mutants were prepared but not run before the base ruling. No main-based build, vet, JSON tests, Node agreement, four-mutant proof or stage-3 roots are claimed for this stopped port.

Toolchain setup on the earlier branch succeeded: go .016s, node .023s, submodules .061s, markdown .066s, clang .193s, go build 62.887s, build cache 63.010s, done 63.039s. Node was v24.19.0 and nproc was 5. These are earlier setup observations, not main-based validation.

Choice: take the explicit stop option at the confirmed missing field-kind prerequisite. No question was asked and no unlanded runtime infrastructure was imported.
