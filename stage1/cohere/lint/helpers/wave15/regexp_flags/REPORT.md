Ported Go regexp.parseFlags in parse_flags.a, preserving all five flag fields and exact error bytes, including partial state on failure.
Claim was pushed at e4526d934 before code; rule branch is parked at 1f7f2dc4, helper code is based on current main f8013f0b.
1,197,583 queries produce 173,833,604 identical bytes on Go, source Node, emitted JavaScript and ASan/UBSan native; upstream tests for all four consumers and regexp pass.
The duplicate-flag mutant compiles and finishes successfully with empty stderr on all three Adamic paths; only the independent output comparison catches it.
This supplies four prerequisite edges, zero final blockers; the selection scan initially missed a withdrawn six-consumer cache claim, corrected below. Full regexp execution and complete rule ports are not certified.

Consumers in frozen readiness.json:

- @next/next/no-html-link-for-pages
- @typescript-eslint/no-empty-object-type
- no-restricted-exports
- no-restricted-imports

Production API: parseFlags(flags: string): RegexpFlags. Flags i, m, s, u and y populate fields; g and d are accepted and ignored but participate in duplicate detection. v and other runes produce the actual Go error text. Earlier fields survive every error. Each result is fresh; no state persists between calls. UTF-16 lone surrogates cross the Go string boundary as replacement runes. Raw malformed UTF-8 byte strings are not an additional input interface.

The quoting table is generated from the actual Go toolchain's strconv.IsPrint and validated for drift on every run. It is data in printable_ranges.a, not another helper implementation. Controls use Go's named escapes, byte escapes only below SPACE or for DEL, and four/eight-digit Unicode escapes otherwise. Every Unicode scalar is compared through the real private Go parseFlags, rather than an independently rewritten oracle.

Coverage: all 1,112,064 Unicode scalars, all 2,048 lone UTF-16 surrogate units, every sequence of length zero through five from imsuygdvq, every ordered subset of the seven supported flags, and controls. The generator also extracts every Go string literal from every original fixture file for the four consumers: next 284 unique strings, TypeScript 186, restricted exports 178 and restricted imports 463, each tested directly and after i/imsuy prefixes. This includes source/configuration strings, not just handpicked flags. These bounded sequence families are not an exhaustive proof over arbitrary-length flag strings; the loop's first-error behavior is preserved structurally.

Command, after source /workspace/adamic-tools/env.sh:

python3 stage1/cohere/lint/helpers/wave15/regexp_flags/validate.py > /tmp/wave15-flags-validation.log 2>&1

It builds the Go oracle using an overlay inside the cohere module, generates cases and the printable table, builds both Adamic backends with the owned build.go.txt overlay, compares all outputs, builds and runs the mutant, then runs:

- go test ./internal/lint/rules/next -count=1 -run '^TestNoHtmlLinkForPages' -timeout=10m: PASS, 0.035s
- go test ./internal/lint/rules/typescript -count=1 -run '^TestNoEmptyObjectType' -timeout=10m: PASS, 0.042s
- go test ./internal/lint/rules/core -count=1 -run '^TestNoRestricted(Exports|Imports)' -timeout=10m: PASS, 0.120s
- go test ./internal/lint/ecmascript/regexp -count=1 -run . -timeout=10m: PASS, 0.204s

Mutant: duplicate test (seen & bit) !== 0 becomes (seen & bit) < 0. Witnesses ii, imm, gg, dd, imsuy and iim compile and exit normally. Ignored g/d expose why flag fields alone are insufficient: their mask stays zero, while the error comparison catches the mutation. The stdout hashes of all four full runs are 8f167b33a6c6e8d46cf6fb6ff12d19b2ff06770d581ffd8614a19ddafe81a653. Raw outputs and generated cases remain in /tmp/wave15-regexp-flags; small logs and hashes are retained here.

Observed corrections: the initial comparison caught U+0080 quoted with a byte escape instead of Go's Unicode escape. It was corrected before the final four-way pass. The initial oracle build from the parent module could not resolve cohere's internal package; running the same overlay build in the cohere module fixed it, without changing dependencies or upstream files.

Setup: bash cloud/setup.sh succeeds. Timing lines: Go 0s, clang 0s, Node 0s, submodules 0s, build-cache warm 99s, total 99s. nproc is 5; cgroup cpu.max is 400000 100000. Go 1.27.1, clang 20.1.8, Node 24.19.0; environment /workspace/adamic-tools/env.sh. Setup's compile-only warmup is not a full test gate. The current main landing already reran the two retained CSS helper suites and helper packages; this new helper has its own full comparator, rather than changing that shared harness.

Selection correction: all 530 refs and 17 claim files were checked, but the initial presence-based scan treated slot 05's withdrawn DesignSystemForProgram as still reserved. Its explicit release makes it a higher-count unclaimed helper at six consumers. No other active owner is found. parseFlags is useful and independently validated, but the assertion that four was the highest available count was wrong. The continuation reclaims the six-consumer helper before probing its prerequisites. No shared finding.ts, context.ts, main.ts, registration generator, lint comparison or compiler implementation is edited.
