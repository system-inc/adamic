# Regex landing plan

Task: #8s7zm5s

Source: `origin/library/area-fixes` at `b2c7dfab`. First base:
`origin/main` at `c0fd8945`. Each later branch starts on the preceding unit.
Only `library/regex-1` is built in this turn.

1. **library/regex-1: capture participation.** Copy `capture_facts.go`,
   `capture_facts_test.go` and `testdata/capture-tsc.json` from the area tip.
   Prove which literal captures always participate, preserving optional captures,
   alternatives, duplicate names, negative assertions and repeat resets.
   No lowering or runtime integration is needed. Gate the entire Go regexp package
   and kill the optional-capture mutant against Node.
2. **library/regex-2: Go and native matching.** Carry the parser/VM fixes, V8
   compatibility refusals, regular candidate engine, search optimizations and
   matching runtime tests together with their C declaration format. Include the
   V8 divergence document. Preserve current constructor behavior while deferring
   dynamic construction entry points to unit 3. The new `NativeDeclarations`
   initializer and `regexp.h` layout must move together; split mixed runtime
   hunks only at this boundary. Test both native and wasm32 against Node.
3. **library/regex-3: runtime pattern compilation.** Carry the runtime parser,
   set compiler, bytecode compiler, V8 checks, generated Unicode tables and their
   identity/corpus tests. Use the area generators, never hand-edit table data.
   Add only the runtime build/include glue this compiler needs. The Go snapshot
   seam and runtime corpus inventory belong here. Do not expose dynamic lowering
   until unit 4.
4. **library/regex-4: lowering and language agreement.** Carry regex constructors,
   capture/group representation, constant evaluation, protocol/string methods,
   literal-input proofs, regex-only load/IR/native glue and oracle fixtures.
   Trace `expression.go`, `library_string.go`, prelude/library declarations and
   native dispatch rather than copying shared files. Add the stage-1 regex slice
   only where supported by this unit. Regenerate counts with `-update-counts`
   and flow/fresh units with `review/test-split-flow/generate.py` when needed.
5. **library/regex-5: test262 RegExp adapter.** Carry the RegExp runner, syntax and
   constructor-error adapters, fixtures and shards, with only required dispatch
   glue. Keep test-parallelism infrastructure out. Shard any single test exceeding
   60 seconds. Drop `bench/regex` from this landing stack; no performance claim.

Every unit must build on its predecessor, vet touched packages, test each touched
package with a 90-second invocation timeout where possible, and keep every single
unit below 60 seconds. Each added function family needs a semantic mutant caught
by Node agreement. Later boundaries are implementation plans, not claimed gates.

Exclude c2 (`origin/cloud/land-stack-c2-28107975`), host work, unrelated
Object/Array/JSON/Date/Error changes, test-parallelism machinery, module files,
buildcache and toolchain changes. Unit 1 has no c2 symbol dependency. If later
regex code requires a c2-only symbol, stop that piece and report the exact symbol.
The Unicode package changes in the area are child-process guard infrastructure,
not new Unicode tables, and are deferred out of this regex stack.
