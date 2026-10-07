# Slot 04 wave 17

Built IsColor, IsNamedColor and IsLength in separate `.a` files. Claim 05bd8666 was pushed before code. All 47 preceding retained helpers were already complete, green across seventeen helper packages and an uncached oracle on current main b8fb957aa839a9e8cb0b54279dd9864fa317bd30, pushed through 727acb53. Initial and final wildcard fetches checked all twenty origin helper branches and their claim trees. These helpers tie the maximum available named fan-out, four consumers each. No competing claim was found. Main remained unchanged, so no rebase was necessary for this batch.

## Observations

Named-color lookup preserves Go's lowercase behavior through an explicit dependency callback and uses the pinned 169-entry table. One observed distinction from JS lowercase is İndigo: Go accepts it as indigo, while direct JS lowercase retains a combining dot and would reject it. Whitespace is not trimmed. Go's private table is replayed dynamically in lower/upper/space-prefixed forms as well as the generated controls, so a new Go name must be supported rather than silently missed.

Color matching uses static hash-prefix and color-function JS RegExp literals, then the named-color helper. It deliberately accepts invalid CSS such as #zzzz and unfinished rgb(. Length matching delegates the numeric-unit check, uses a spacing-function JS RegExp literal, then delegates math-function detection. Function expressions use plain i, preserving Go's fixed-byte prefix behavior on the tested non-ASCII lookalikes. Ordinary JS i does not introduce the long-s and Kelvin-to-ASCII folds of iu. No handwritten matcher, rule listener, finding-position conversion or shared file was added.

The fetched shared regex table b39979305d880d7d05083704b73d405cd3694d68 contains 107 rows, none for IS_COLOR_FN/is_color.go/--spacing. The function literals were transcribed once from the pinned Go comments' upstream expressions. Go lowercasing, numberWithSuffix and hasMathFunction remain explicit separately owned callback dependencies. The driver obtains actual Go casing and numeric-prefix results; suffix membership uses the list supplied by the new length helper. Arbitrary dependency implementations are not claimed as ported.

The actual Go Tailwind suite passed in 0.405s under temporary overlays and captured 112 unique asserted inputs across all four consumers. Eleven real helper calls produced five distinct records: one IsColor input, one IsNamedColor input and three IsLength inputs. All three helpers were reached. Consumer source texts are additionally supplied directly as helper inputs; this is not native execution of complete lint rules.

`go test -count=1 -v ./stage1/cohere/lint/helpers/slot04_wave17` passed in 30.107s. Go, source Node, ASan/UBSan native and emitted JavaScript matched 2,416 control lines, 112 consumer-source lines and five captured-call lines: 2,533 lines containing all three predicate results, or 7,599 boolean comparisons per Adamic mode. The controls comprise 1,909 generated values and 507 dynamic Go named-table replays. They cover every named color in multiple cases, whitespace and suffix rejection, all thirteen color functions plus spacing, incomplete functions, prefixes and suffixes, line terminators, NUL, Unicode letters and astral lookalikes at each function-letter position, numeric exponents, incomplete fractions, unit case, Q versus q, unknown units, var(), and math-function placement/case. Successful runs require exit zero and empty stderr.

Twelve semantic mutants compiled and ran successfully in each Adamic mode, then differed from actual Go:

- Omit named-color lowercasing.
- Replace black with notblack in the table.
- Trim a named color before lowercasing.
- Require two hash characters instead of one.
- Remove the color-function start anchor.
- Remove color-function case insensitivity.
- Require rgba instead of accepting rgb.
- Omit numeric length acceptance.
- Replace the uppercase Q unit with lowercase q.
- Remove the spacing-function start anchor.
- Remove spacing-function case insensitivity.
- Negate math-function detection.

Four consumer-omission mutants failed the coverage check, one per rule below. The earlier 25.931s suite also passed before dynamic table replay was added. `python3 stage1/cohere/lint/helpers/slot04_wave17/testdata/generate.py` reproduces witnesses.json with identical SHA-256 bytes. All logs are in evidence/.

`go vet ./stage1/cohere/lint/helpers/slot04_wave17` passed with an empty log; `git diff --check` passed. `ADAMIC_GATE_UNCACHED=1 go test -count=1 -v -timeout=10m ./internal/oracle -run '^TestNativeAgreesWithNode/internal/oracle/testdata/regexp_unicode.a$'` passed in 0.990s with three native misses, two Node misses and zero cache hits.

The latest shared harness 41eb6eab2 was fetched; it is not yet in main. These helpers do not use the finding model. Incoming allocator changes were not reverted, and no shared harness, compiler, registry or rule files were edited. Inherited setup passed: Go 0s, clang 1s, Node 1s, submodules 2s, warm 165s, total 165s; nproc printed 5 again. Build shells source /workspace/adamic-tools/env.sh. The full repository gate was not rerun: the preceding seventeen helper packages were already green on this unchanged main, and this new package and a filtered uncached oracle were run.

## Readiness inference and limits

Each helper removes its named prerequisite from each of:

- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

That is twelve prerequisite removals across four distinct rules, assuming the explicit callback dependencies are supplied. None becomes free of every listed blocker from this batch alone. Unintegrated work on other workers' branches is not counted as landed. All fifty retained helpers are complete, with no outstanding reservation.

Not covered: complete rule diagnostics/fixes/suggestions or native registry integration, a new implementation of any supplied callback, malformed UTF-8, every arbitrary string or runtime profile, exhaustive Unicode case tables, and the full repository gate. Observed bounded parity and live reachability do not establish exhaustive rule parity. See README.md for callback contracts.
