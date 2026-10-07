# Fixed hexadecimal regexp escapes

Retained helper: github.com/system-inc/cohere/internal/lint/ecmascript/regexp.decodeFixedHex.
Claim 7d7a1dcc7 was pushed before implementation after inspecting all 580
origin heads, 20 distinct helper claim contents and HELPERS.md reservations.
Four consumers ties the maximum remaining unclaimed concrete-symbol fan-out.

The .a implementation takes raw bytes, byte index, introducer width, digit
count and Unicode mode. Bytes must be integers in [0,255], index an integer
in [0,length], and width/digits nonnegative integers at most 2^31-1. This keeps arithmetic
exact and outside Go int overflow. Actual escape-scanner callers use width 1
and digit count 2 or 4. It preserves all decoded
fields, signed rune conversion after uint32 parsing, exact width, absence of
error, errors.Is(ErrUnsupportedSyntax), and raw error bytes. Invalid UTF-8
fallback and Go fmt's one-rune precision retain their different semantics.
Panicking Go inputs outside this domain are not covered.

Reproduce with the setup environment sourced:
python3 stage1/cohere/lint/helpers/slot14/validate_hex.py > /tmp/fixed-hex.log 2>&1

The oracle calls the pinned Go cohere private helper itself, through test-only
Go overlays. No shared repository harness, parser, compiler or Go submodule
source is edited. Source Node, emitted JavaScript and ASAN/UBSAN native each
compare every output byte over 68,844 unique inputs (2,149,768 output bytes).
Controls include every two-byte hex payload, all fallback bytes, malformed
UTF-8, byte offsets after multibyte characters, signed uint32 boundaries,
prefix/sign/underscore rejection, zero/large digit counts, and seeded cases.

All four original consumer Go suites pass their assertions. None invokes this
leaf on its original fixtures; this fact is recorded rather than inferred as
live coverage. Additional real decoder/rule tests record two private calls for
no-restricted-exports, two for no-restricted-imports, one for
@typescript-eslint/no-empty-object-type, and two for Next.js route compilation. Those captured inputs join the same
three-runtime Go comparison. The Next.js control drives its real compileRouteSources path and checks both
matching and nonmatching hrefs. These checks compare the leaf
result; they do not run an Adamic regexp engine or compare complete consumer
findings/fixes on native. The surrounding regexp/parser adapters still need
integration, outside this helper's owned file.

Mutation: omit the introducer from successful width. It compiles, exits zero
with clean stderr and is caught only by byte comparison on all three runtimes.
Two initial targeted-fixture launches failed: a variadic assertion was called
with an unexpanded slice, then an expected import finding ID was singular
rather than Go's observed "patterns". Both owned test mistakes were corrected
before the final run; neither is counted as a helper mutant.

Dependency entries made available: four, for @next/next/no-html-link-for-pages,
@typescript-eslint/no-empty-object-type, no-restricted-exports and
no-restricted-imports. Complete rule blocker sets removed by this leaf alone:
zero. The complete-regexp and complete-consumer-parity gaps are explicit;
this bounded result must not be advertised as four completed rule ports.

Rebased onto current origin/main b8fb957aa839a9e8cb0b54279dd9864fa317bd30.
The retained helper and width mutant were rebuilt and compared again after
the compiler change. Archived withdrawn witnesses were also re-green: 4,761
integer cases and 1,012 declaration cases; they remain excluded from delivered
helper counts. The actual inherited_static_field_read.a oracle passes across
source Node, emitted JavaScript, sanitized native and release native, and
TestTheOracleCatchesOneByte passes. An initial abbreviated fixture filter ran
only the parent test; the full slash-separated path was corrected and its
executed subtest is visible in the retained log. Full repository gate not run.

Latest landing rebase: 39638d9e278d38bb5aeae887f46d55a70e47aaad. Changes since b8fb957aa are outside compiler/dependency/runner/stage1 paths. The retained helper, all four original consumer suites, captured targeted paths and all three width-mutant executions pass again: 68,844 inputs / 2,149,768 bytes. The correctly selected inherited-field oracle and one-byte failure control pass again (cached artifacts with unchanged compiled inputs). No additional helper was claimed while the landing cap was being resolved.
