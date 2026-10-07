# Slot 05 third batch

Three shared Go judgments, one public helper per .a file:

- matchExactly(candidate, wanted) compares exact strings. It does not fold case, normalize Unicode, trim or search substrings.
- normalizedFileName(files, index) returns empty for a missing SourceFile and replaces every backslash in the supplied FileName with a forward slash. It does not clean dot segments, repeated slashes, case or roots. SourceFileName is an immutable projection; -1 is nil and all other valid indexes address one file record.
- isHookName(name) requires the exact ASCII prefix use and a following Unicode uppercase rune. Go accepts non-ASCII capitals and supplementary capitals, rejects digits and titlecase-only characters, and does not require the rest of the name to be alphanumeric. The private sorted/strided Unicode 17.0.0 table is generated from pinned Go unicode.Upper inside the same helper file. Binary search preserves the Go range/stride classification.

The Go pin is cohere 715ba94f3608a6500086b1076ce5cb7e51b836db. Tests refuse pin drift, missing consumer evidence and uppercase-table regeneration drift. All implementation, test and evidence files stay in this owned directory. Shared registration, shared harness, compiler files and the Go worktree are unchanged.

The isolated oracle scans all nonempty Go string literals in every inventory-listed test file for every dependent rule. Exact matching and hook naming also parse those strings as TSX and collect decoded Identifier, StringLiteral and NoSubstitutionTemplateLiteral text. Description and option strings are included; these counts are not full lint-fixture counts. Dynamic source expressions and external corpora are not reconstructed.

Exact-matching cases pair every captured text with itself, a suffixed string, upper/lower versions, empty and a NUL suffix. Filename cases use all consumer literals, test-file paths, nil and path/Unicode/NUL controls. An oracle-only constructor sets the actual private SourceFile field read by Go FileName: the normal parser forbids unnormalized or empty filenames and would otherwise prevent testing precisely the Windows behavior this helper promises. This overlay changes neither FileName nor NormalizedFileName and does not write the submodule.

Hook cases include all captured source/decoded texts, discriminating controls and all 1,112,064 Unicode scalar values after use. Source regeneration must match the checked-in helper byte for byte. The scalar sweep has a compact JSON instruction, avoiding a million redundant JSON objects.

Every successful baseline and mutant runs on Node source through oracle/node.mjs and sanitized native. A semantic mutant must compile, exit 0 with no stderr, and differ from actual Go output. A compiler or sanitizer failure is never credited. The helper package tests are independently owned here and import only existing compiler APIs; they do not edit the shared rule harness.

Run from the repository root after sourcing /workspace/adamic-tools/env.sh:

    ADAMIC_SLOT05_BATCH3_EVIDENCE=/workspace/adamic/stage1/cohere/lint/helpers/slot05/batch3/evidence ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch3 -count=1 -v -timeout=10m > /tmp/lint05-batch3-final.log 2>&1

Regenerate the Unicode-bearing helper with the pinned setup toolchain:

    go run stage1/cohere/lint/helpers/slot05/batch3/testdata/generate_hook.go > stage1/cohere/lint/helpers/slot05/batch3/react_is_hook_name.a

[CONSUMERS.md](CONSUMERS.md) lists each helper's seven rules. [readiness.json](readiness.json) separates dependency removals from final helper blockers and includes conservative cumulative accounting for all eight slot 05 helpers. [REPORT.md](REPORT.md) records commands, mutants and limits.

Limits: no full rule diagnostics/fixes/suggestions, emitted-JavaScript comparison, production parser/linter integration, dynamic/external fixture reconstruction, raw invalid-UTF-8 strings or lone-surrogate cross-language representation checks, arbitrary malformed arenas, or full repository gate.
