Built: tailwind.projectRootOf in project_root.a, preserving configured-file preference, POSIX filepath.Dir and lazy fallback reads.
Commits: claim 398cedef was pushed before implementation; implementation commit follows this report.
Commands and outputs: owned four-way comparison and three compiling mutants PASS in 12.757s; 2,935 cases and 85,445 identical bytes.
Mutants: accept empty configured path, omit parent cleaning, and clean fallback cwd all finish cleanly and differ from actual Go on all three execution paths.
Not covered: Windows filepath semantics, concurrent program mutation, full program/CSS engine integration, or a full repository gate.

This is one helper per .a file. The options result is a structural interface: the initial class-returning callback was explicitly refused by stage 0's nominal ancestry check. The owned model was changed to the recommended interface; no compiler or shared harness was edited. The superseded compile refusal is in evidence/first-nominal-refusal.log and is not credited as a mutant catch.

The implementation calls options once, uses the config filename's directory only when options are present and its filename is nonempty, and otherwise calls currentDirectory once and returns its exact bytes without normalization. Config directories use the current POSIX Go filepath semantics: repeated separators, dot/parent components, rooted parents, empty/bare filenames, Unicode and literal backslashes. A non-null typed RootProgram is required, like the actual Go callers. The interface makes program reads explicit; it does not implement a program engine or read compiler options from disk.

Capture used scratch overlays of Go's testing layer, without changing production rules. The selected original Go tests PASS in 0.280s. Actual compiler-options and current-directory values are captured before subject.Run from all six consuming rules: 35 canonical, 50 class-order, 16 variant-order, 2 shorthand, 22 conflicting and 26 unknown cases, 151 total. Each contains its original filename and source. The shared oracle executes the real private Go projectRootOf through an oracle-only exported wrapper. An embedded Program with only Options and GetCurrentDirectory overridden records exact read order; an unexpected new program read would fail. The captured actual program values are replayed independently of the rule's findings. This is helper parity, not a native whole-rule parity claim.

The 2,784 additional controls cross 174 config paths, eight cwd strings, and present/absent options. They preserve unusual cwd values to prove the fallback is not normalized. Go source, source Node, emitted JavaScript and ASan/UBSan native compare result text and callback trace byte for byte. All credited mutants compile, exit zero and have empty stderr, including no sanitizer/leak report. None is credited for compilation failure or panic.

Six dependency removals, zero newly complete rules:

- better-tailwindcss/enforce-canonical-classes
- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-consistent-variant-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

Every listed rule loses project-root selection as a prerequisite; each still needs other recorded helpers. This is a conservative frozen-ledger calculation, not a claim that these six rules are fully helper-ready. The original 198-rule readiness.json is not edited. The new private readiness.json lists the exact residual dependencies.

Commands, with source /workspace/adamic-tools/env.sh:

* bash cloud/setup.sh > /tmp/wave104-helper-setup.log 2>&1: PASS. Go, clang, Node and submodules each reported ready at 0s; cache warm 76s; done 76s, nproc 5, cgroup quota 400000 100000. Go 1.27.1, clang 20.1.8, Node 24.19.0.
* python3 stage1/cohere/lint/helpers/from_wave1_04/capture.py > .../evidence/generation.log 2>&1: original Go suite exit 0, captured 151 cases from six actual-program consumers. The machine-specific Tailwind package lookup is redirected to the existing tailwindcss@4.1.18 installation at /tmp/wave104-tailwind. No production Go helper changes. Reproduction requires that installation.
* ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/from_wave1_04 -count=1 -v -timeout=10m > .../evidence/tests.log 2>&1: PASS 12.757s, 2,935 cases, 85,445 bytes, all three semantic mutants caught on source Node, emitted JavaScript and sanitized native.
* go vet ./stage1/cohere/lint/helpers/from_wave1_04 > .../evidence/vet.log 2>&1: exit zero, empty log.
* ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=10m > .../evidence/oracle.log 2>&1: result recorded in log.

The new branch is based on origin/codex/lint-helpers at 95100eb440b47f3f18e960c6f5b49cadf0dc1d9d, pinned Go cohere 715ba94f3608a6500086b1076ce5cb7e51b836db. Before claiming, 402 origin refs and 15 distinct helper claim blobs were checked. Comment helpers are already delivered on the base; FindEntryPoint was claimed by slot 11 during the refresh and was left untouched. No previous rule code is carried onto this helper branch.
