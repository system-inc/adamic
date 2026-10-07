Rebased the 12 owned rules onto main f8013f0baac41ddc340d76f83bddde38536a8f07 and reran their standalone oracles.
Rule code before this evidence commit: 76f8bb02226aa4355e35411b0a8d53c5ac941975; helper branch pushed at d7434e4368af304492ebb23ceee7c639d001bda4.
472 supported fixtures and 4,176 file/rule pairs agree byte for byte with Go on Node, emitted JavaScript and sanitized native.
All 13 rule mutants and the numeric-kind declaration mutant compile and finish successfully, then fail only their independent output comparison on all three backends.
Shared integration, numeric rule.json kinds and handed-node callbacks remain blocked; no new helper claims.

After source /workspace/adamic-tools/env.sh, all four commands completed successfully:

- python3 stage1/cohere/lint/rules/typescript-no-unnecessary-type-constraint/validate.py --scratch /tmp/wave15-f801-six --typescript /tmp/lint-wave1-15-typescript --mutants
- python3 stage1/cohere/lint/rules/no-multi-str/validate.py --scratch /tmp/wave15-f801-literals --typescript /tmp/lint-wave1-15-typescript --mutants
- python3 stage1/cohere/lint/rules/structure-tailwind-no-physical-direction/validate-standalone.py --scratch /tmp/wave15-f801-three --typescript /tmp/lint-wave1-15-typescript --mutants
- python3 stage1/cohere/lint/rules/structure-tailwind-no-physical-direction/verify-syntax-kinds.py --scratch /tmp/wave15-f801-kinds

The six-rule suite compares 206 fixtures (83,165 bytes) and 2,088 corpus pairs (79,645,656 bytes). The literal suite compares 74 fixtures (26,425 bytes) and 1,044 pairs (39,663,897 bytes). The final three compare 192 fixtures (51,192 bytes) and 1,044 pairs (39,671,239 bytes). Each corpus contains 77 TypeScript compiler files plus 271 stage1 files. Original asserted Go rule tests pass. Raw validator logs and output SHA256 values are retained here; complete generated outputs remain in the named scratch directories.

The filtered oracle command ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v passes in 0.290s, with one native miss and one Node miss, zero hits.

Mutants rerun: unnecessary constraint Any/Unknown to Any/Never; prefer-as-const inverted literal equality; enum initializer position plus two; incorrect NodeFileSystem alias; invalidation method name changed; query binding minus one; multiline U+2029 removed; decimal escape previousNull forced true; octal maximum digit nine to seven; module name changed; default parameter QuestionToken changed; physical direction listener disabled; physical direction replacement uses interpolation-sensitive replaceAll; numeric no-octal declaration eight to nine. The individual backend catches are in the retained logs.

Ten Go-valid parser inputs remain excluded and explicitly refused on all three Adamic paths: nine literal JSX/recovered-syntax inputs and one physical-direction JSX input. This is observed incomplete coverage, not parity for those inputs.

Shared blockers: registry.go still requires rule.ts, defines Kinds as []string and dispatches via context.node(index).kind; generated callbacks pass indexes rather than the node. ParseNode.kind remains a string on current main. All owned descriptors already declare listeners; numeric exports match the pinned parser, but these exports do not accelerate the current driver. The shared finding bridge cannot express all suggestions and multiple independent edits, so standalone validation certifies those data while integration refuses unsupported shapes. Batch 8 Diagnostic is still on its separate branch and no integration SHA has been named. No shared harness files were edited. Recreated registration merge conflicts reuse the published efeb3f66 integration resolution.

cloud/setup.sh exits one during warmup because profile_test.go:32 ranges over the portFiles function without calling it. Go, clang, Node and submodules timing lines are each 0s; no successful warmup or total timing line was printed. nproc is 5. Toolchain: Go 1.27.1, clang 20.1.8, Node 24.19.0. Standalone owned builders are the workaround. Full repository gate and default lint integration are not green.

Performance was not remeasured during this landing rerun. Prior findings-per-second observations and their measurement boundary remain in ../standalone-evidence/REPORT.md; these are not new measurements on f8013f0b.

Parked under Ahra's shared-harness exception. Both owned branches are rebased on current main and green against their independent oracles. The shared registration, node callback and finding bridge blockers above are the remaining integration obstacles; no shared files are edited to bypass them. Resume harness integration when its landing SHA is named.
