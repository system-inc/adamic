Three rows defended in the bounded production-behavior matrix; keep all three.
Each has its own aimed production mutant, standalone diff, observed failure and passing competitors.
Whole-package baseline exceeded 90 seconds without an ordinary assertion failure; limits remain explicit.

Base: bfe0553300773c0b37db2c10df97adeb909705f8. nproc: 5. Warm env.sh worked, setup skipped. npm ci stage3/api ran before baseline. Prettier 3.9.6 installed separately and pinned TypeScript source 050880ce59e30b356b686bd3144efe24f875ebc8 enabled. Current go test -list lists 250 top-level tests; list.log is authoritative. No target vanished. Family has 65 leaf wrappers and its union. Source and tests were restored; no tests changed.

Coverage and semantic defense
- Prefix: 47 exclusive Go blocks versus Documents, including the prefix-update closure return. P1 changes its returned value to zero while preserving the decrement side effect. It prints 0 instead of 2. The port programs use increments as separate statements, which lower through increment directly.
- Number: no exclusive Go blocks versus Documents. N1 replaces the null-kind constant in the existing empty-number branch with string-literal kind. Number('17') prints 0 instead of 17. Port transports use computed Number operands, so they are unaffected. Shared lines with a different AST-kind input defend this row.
- Expressions: Go -coverpkg measures compiler/native builder, not the TypeScript source. Actual NODE_V8_COVERAGE shows formatExpression called 2644 times by leaf 000 and absent from Documents. E1 changes the StringLiteral discriminator for parenthesized standalone expressions. The generated ((('é😀'))) case loses its parentheses in leaf 008. Documents never calls formatExpression; statement families call formatProgram. All eight independently selected TSC agreement leaves pass E1, so that other expression corpus does not cover this case.

Matrix bounds and exclusions
The package's mandatory whole baseline timed out at 90.064 seconds, with 24 top-level passes and no ordinary assertion failures. It was narrowed to Documents, all standalone production gap tests, representative leaf 000 of the expression and statement families, all eight TSC agreement leaves, and 18 small setup/witness checks. E1 additionally ran clean and mutated expression leaf 008 plus a native sanitized product build. matrix.json lists exact observed passes and failures. Commands have a 90-second binary limit and 120-second outer backstop, each mutant uses its own ADAMIC_BUILD_CACHE_DIR. No narrowed run cooked.

Remaining expression leaves belong to the SAME defended row, so they cannot subsume it. Remaining statement leaves execute the same compiled statementsMain.ts program: E1 changes a function that their entry does not call; P1 and N1 change compile-time lowering branches absent from this port's source. grep of Number call sites is preserved. Product rows only prepare/build artifacts, not compare semantic outputs; all three mutants compile (P1/N1 go vet, fresh native fixture/port builds; E1 native sanitized product). The built-in TestMutants family witnesses deliberately wrong ports and is not evidence of semantic subsumption; it was not replayed. These exclusions are source-based, not invented passing observations. We do not claim a complete 250-test execution or repo-wide uniqueness. Central replay can strengthen these bounded defenses.

Oracle strength and owner findings
Expressions compares exact output against executed Go cohere, npm Prettier 3.9.6 and embedded Prettier with recorded allowance checks. Prefix and Number compare native and JavaScript backend outputs with Node and handwritten literals 2/17, plus leaks. Number's name is broader than its only input: it guards Number('17'), not general ToNumber conversion, NaN, null, undefined, or partial-number rejection. Its prior untrue verdict depended on a mutant that left this input unchanged. No row was left undefended in the bounded matrix.

Unclear instructions and time costs
1. The audit's expression-family verdict rested on member 000 and one mutant; it did not test the parenthesized-string cases owned by other leaves. A passing representative cannot settle every behavior of a family.
2. Go -coverpkg cannot instrument a TypeScript port. V8 coverage is supplied alongside Go profiles rather than attributing host compiler coverage to the port.
3. The package cannot complete within the specified 90-second binary budget. A whole-package uniqueness claim would overstate these runs; bounded rows and exclusions are explicit.
4. My initial ownership calculation used 65 rather than 64 input owners. Shard 064 is the remainder check. This cost passing runs of leaf 037. Those logs are retained, the correction is recorded, and the actual owner 008 was tested clean and mutated.
5. Stage1/compiler mutants require native rebuilds despite warm Go tools. Separate caches were used, and commands.jsonl records end-to-end durations. Build and execution are mixed within test setup, so no unsupported separate build-only timing is claimed.
6. The audit files are README.txt, results.json, plan.json and matrix.json rather than report.md/rows.json; all were read and copied.

No empty-answer probes were needed for this defender task. No deletion, weakening, main push or pull request occurred.

Recorded mutant matrix commands took 359.15 wall seconds total, including native setup; each detailed duration is in commands.jsonl. Clean coverage/baseline and npm installation are logged separately. Total session approximately 27 minutes.
