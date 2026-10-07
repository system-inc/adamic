Built parseCSSString in one .a helper file, preserving raw bytes, ranges and error messages.
Claim 32d1d9d1 was pushed before code on codex/lint-helpers-from-codex/lint-wave1-15; the earlier VariantKind duplicate was withdrawn.
validate.py: 198,584 cases, 5,648,824 identical output bytes on Go, source Node, emitted JavaScript and ASan/UBSan native; original consumer tests passed.
Escape-skipping mutant compiled and ran with exit zero and empty stderr, then differed from Go on all three backends.
Not covered: whole CSS parser or lint-driver integration, invalid byte-view/start/quote API arguments, arbitrary unbounded configuration and full repository gate.

Consumers and readiness

The six rules are better-tailwindcss/enforce-canonical-classes,
better-tailwindcss/enforce-consistent-class-order,
better-tailwindcss/enforce-consistent-variant-order,
better-tailwindcss/enforce-shorthand-classes,
better-tailwindcss/no-conflicting-classes and better-tailwindcss/no-unknown-classes.
Six prerequisite entries are supplied; zero final helper blockers are removed.
The entire CSS parser and other dependencies still have separate owners. This
is helper parity, not six completed rule ports. No shared readiness file changes.

Contract and observed behavior

The input is a raw-byte view, not UTF-16 string indexing: valid and malformed
UTF-8, NUL, line endings and Unicode are preserved without replacement. start
is an integer in [0, byte length], quote is an integer byte. Values in the view
must be bytes; callers construct the view from source storage. The result carries
closing byte index, failure state, original byte offset and raw error bytes.
The error renderer retains Go fmt %c's UTF-8 encoding of a non-ASCII quote byte.
No result message is decoded through a lossy UTF-8 conversion.

Backslash skips exactly one byte. The matching quote wins before the newline
error tests. A semicolon before LF or CRLF is included in the error message;
a bare LF or CRLF is excluded. A lone CR is ordinary content. EOF before a
closing quote is success at the original start index, however surprising that
is. The unchanged private Go function, not a rewrite of it, decides all answers.
The owned overlay adds only an exported oracle seam and standalone entry.

Evidence and reproduction

    source /workspace/adamic-tools/env.sh
    python3 stage1/cohere/lint/helpers/wave15/css_string/validate.py > /tmp/wave15-helper-css-string-validation.log 2>&1
    ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v > /tmp/wave15-helper-input-oracle.log 2>&1
    bash cloud/setup.sh > /tmp/wave15-helper-setup.log 2>&1

The Go AST collector unquotes every distinct string literal from all six
consumer test families: 131, 196, 48, 159, 130 and 139 literals. Quoted wrappers
and every quote position form helper probes. They are fixture-derived controls,
not a claim of observed helper calls from each full rule run. Original Go tests
for all six consumers run independently with their own assertions and pass.

Controls cover all 256 raw byte values crossed with all 256 quote byte values,
escaped bytes and nonzero offsets; EOF/start bounds, CR, CRLF, semicolon,
multibyte prefixes and continuation cases are also pinned. All 198,584 successful
Go/Node/JavaScript/native outputs match exactly; empty stderr is required.
The semantic mutant disables escaping (92 to 999), rebuilds emitted JavaScript
and sanitized native, exits normally on every case, and differs only in output.
No compilation refusal or sanitizer failure is credited as a mutant catch.

The external uncached input oracle passed in 6.521s, with six probe misses and
zero probe hits. Setup succeeded: Go, clang, Node and submodules ready in 0s,
build cache warm 74s, done 74s, nproc 5. Environment Go 1.27.1, clang 20.1.8,
Node 24.19.0. The full repository gate was not run.

Raw generated cases and backend observations larger than 1 MB are compressed
without changing bytes. output-sha256.json records every uncompressed size and
hash; validation, ownership, build, stderr, original consumer and filtered oracle
logs are retained. No compiler, shared helper, harness or cohere source was
edited. One .a file supplies the helper and its result data class; main.a is the
owned test driver. The withdrawn VariantKind evidence is historical, not an
additional delivered helper. The final 416-ref audit confirms only this slot
claims parseCSSString; all other higher-count named helpers remain reserved.
