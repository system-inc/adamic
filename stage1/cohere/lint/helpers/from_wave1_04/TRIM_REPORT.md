Built: trimLeadingJavaScriptSpace in one .a file, removing a dependency from four Tailwind rules; eleven rules are rebased, green under their frozen contract and parked on #zmh9v36.
Commits: rule parking ebdc5a5c7a7f51206c6f49e3e907f15a0c798db5; helper claim a01eb0a03; main f8013f0baac41ddc340d76f83bddde38536a8f07; implementation follows this evidence.
Commands and outputs: complete owned helper gate PASS 26.695s, expanded trim gate PASS 9.351s, original 38 Go tests PASS 0.186s, vet PASS, six uncached input probes PASS 1.174s; setup PASS 38s, nproc 5.
Mutants: omit leading BOM and strip one extra suffix character compiled and completed cleanly; actual Go comparison catches both on source Node, emitted JavaScript and ASan/UBSan native, alongside five prior helper mutants.
Not covered: whole native Tailwind rule findings/design-system implementation, arbitrary malformed UTF-8 bytes or exhaustive astral strings; whitespace is a separately owned explicit predicate dependency, zero final rule blockers removed alone.

# Helper contract and consumers

trimLeadingJavaScriptSpace(value, isJavaScriptSpace) removes only the maximal
leading run accepted by the rune predicate and returns the untouched remainder.
Empty and entirely whitespace strings produce the empty result. The .a file
handles UTF-16 pairs before invoking the separately owned predicate; it does not
copy that worker's implementation. Tests obtain its exact 25-member set by
executing actual Go isJavaScriptSpace over all Unicode scalar values.

The helper removes four edges from the frozen readiness ledger:

- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

All six-consumer symbols and all larger remaining helpers were reserved when
selected. Every one of 550 origin refs and all 19 distinct helper claim blobs
was inspected, including the older shared comment claim in HELPERS.md. The
claim was pushed before writing code. Both owned branches contain current main.
Cumulative delivery is three helpers, sixteen dependency edges, six distinct
consumers and zero complete additional helper-ready rules. readiness.json retains
all other prerequisites conservatively, without treating others' claims as ports.

# Evidence

trim_capture.py selects every test function in all four consumer files, overlays
only Go's recording calls and the existing external CSS fixture root, and captures
every Run. Production Go rule and helper functions remain unchanged. The original
38 tests pass. Deduplication yields 35 class-order, 21 shorthand, 21 conflicting
and 26 unknown sources, 103 total, preserved in trim_fixture_cases.json with their
original filenames. Relative names are normalized only for Go AST construction.

The real Go parser extracts literal/template texts and their whitespace fields.
There are 148 / 69 / 64 / 65 such inputs for the four respective consumers, 346
total. These are the consumers' source literal domains, not a claim to record every
runtime CSS engine invocation. Add every 63,488 BMP scalar leading character,
empty/all-whitespace/trailing-whitespace controls, BOM and astral examples: 63,845
cases. Exact helper output is 883,340 bytes, matching actual Go on all three Adamic
execution paths. Unpaired Go byte strings are outside this Unicode input population.
The private exported wrappers call the real helper, not a guessed implementation.

The two semantic mutants change a real prefix decision and suffix position. Both
compile, return successfully, and have empty stderr and no sanitizer/leak finding.
They are credited only for differing from Go. Source Node is not used as its own
expected oracle. Existing projectRootOf and loadDesignSystemForProgram comparisons
and all five prior compiling semantic mutants also pass in the final package gate.

Commands, with source /workspace/adamic-tools/env.sh:

- python3 stage1/cohere/lint/helpers/from_wave1_04/trim_capture.py:
  original Go tests PASS; trim_capture.json records source coverage and pinned
  cohere 715ba94f3608a6500086b1076ce5cb7e51b836db.
- go test ./stage1/cohere/lint/helpers/from_wave1_04 -run '^TestTrimLeading'
  -count=1 -v -timeout=10m: PASS 9.351s, expanded corpus and both mutants.
- go test ./stage1/cohere/lint/helpers/from_wave1_04 -count=1 -v -timeout=10m:
  PASS 26.695s, all three helpers and seven compiling semantic mutants.
- go vet ./stage1/cohere/lint/helpers/from_wave1_04: PASS, empty log.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle
  -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=10m:
  PASS 1.174s, six uncached probe misses.

All outputs are retained as files under evidence/. Shared finding.ts, context.ts,
main.ts, registry generation, shared oracle and lint_test comparison are untouched.
Rule parking evidence and its exact legacy-contract boundaries live on
codex/lint-wave1-04 under the optional-chain rule's PARKING.md. No shared harness
handoff SHA has been named; that handoff requires rebasing parked work before
claiming anything further.
