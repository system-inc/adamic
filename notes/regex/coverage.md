# Regex coverage inventory

Base: origin/codex/regex-v8-divergences at d91e8f9afdce2d428584a62227c9a2fcac44e419.
Also inspected origin/codex/regex-matcher at 50a1dc8217f6d715fc27038e6ce12ab429416a30.
After refreshing origin/main to 50045bd797650a34aa40b55ad751b6a667a6ab31,
the requested comparison contains 38 files and only commit d91e8f9. Its
production changes are v8.go, matcher.go, native.go and runtime/regexp.c.
The initial stale main reference included earlier integration work; the
refreshed diff confirms the source-case inventory below. The inventory
below distinguishes existing source programs from Go/C corpus tests. Those
corpus tests do not count as existing Adamic oracle programs.

Paths in the Existing column are relative to internal/oracle/testdata.
New executable programs have the regex_coverage_ prefix in that directory.
Each refusal is a separate .a file in regex_coverage_refused, registered with
lowers=false. TestRegexCoverageRefusals additionally checks its complete V8
message and NotYet type. Refusal programs do not have allocation rows because
they do not produce binaries.

| Code case or call shape | Existing oracle program | Added evidence |
|---|---|---|
| Constant RegExp strings, const aliases | regexp.a | constructors, exec, controls, replace, flags, methods |
| Constant concatenation, constant template flags, RegExp call as well as new | No combined program | constructors |
| Fresh objects at repeated constructor evaluation | No loop | constructors |
| Fresh literal objects at repeated evaluation in a loop | No loop on this base; matcher tip has regexp_literal_freshness.a | constructors |
| g/y successful and failed repeated exec, empty matches retain lastIndex | regexp_exec.a | exec repeats each call three times |
| g/y repeated test in loops | sweeps/regexp_methods.a resets before each test, regexp.a calls sticky test twice | exec |
| Unicode lastIndex initially inside a pair: rewind | regexp_v8_surrogates.a tests dot uy | exec includes u/v, g/y/gy, anchors, named captures |
| Unicode failed search retries assertion at low half | regexp_v8_surrogates.a tests B and negative W, once | exec and methods, including repeated test |
| Sticky retry after rewound failure, then stop at requested position | regexp_v8_surrogates.a | exec and methods gy u/v |
| Consuming instruction rejects pair interior, both directions | regexp_v8_surrogates.a forward dot | exec adds lookbehind negative dot and consuming negative-W/dot |
| BMP, ASCII, supplementary pair, lone high/low surrogate, empty input | regexp_unicode.a, regexp.a, sweep | exec |
| lastIndex negative, NaN, infinity, past end | sweep | exec adds fractional index and repeated calls |
| Nonstateful exec ignores lastIndex | sweep | controls, flags; replacement single preserves lastIndex |
| matchAll clone begins at lastIndex and does not mutate original | regexp.a | methods starts inside pair and mutates original after iterator creation |
| matchAll for-of loops, empty advancement, g and gy, u and v | regexp.a, regexp_match.a, sweep for g | methods |
| replace/replaceAll, global and sticky, string tokens | regexp_replace.a, sweep | replace uses constructors and literals with g/gy |
| $1, $<name>, missing/unmatched captures | regexp_replace.a, regexp.a | replace and methods with surrogate assertions |
| $FEATURES literal token and $$FEATURES escaped dollar | No program | replace; methods uses $FEATURES alongside interior matches |
| Replacement callbacks | No compiling program | notes/regex/unsupported_replace.a and unsupported_replaceAll.a, both refused |
| split captures, unmatched capture, limits, empty input | regexp_split.a, sweep | flags uses constructed named captures; split note finds disagreement |
| i/u/v/s/m combinations, flags metadata and named group/optional capture | regexp.a tests ms and iu separately, sweep has gms/giu/giv | flags exercises dgimsu, dgimsv, dimsyu and dimsyv via literals and constructors |
| Named captures, lookaround, greediness, backreference, indices | regexp.a, regexp_match.a and sweep | exec/controls/flags add named captures on branch-specific paths |
| v class string multi-character folding | regexp_v8_folding.a | controls, constructor version and competing strings |
| v singleton string actual folding mismatch | regexp_v8_refused/folding.a only Ss-or-x | singleton_literal/constructor/intersection/subtraction/negated |
| v singleton masked by ordinary operand, intersection removed, subtraction removed | No program | controls: a plus q(a), q(a) intersect digits, q(a) subtract a, negated mask |
| v class strings without folding, numeric strings, empty-only/singleton+empty/multi+empty | sweep covers nonempty multi/singleton alternatives | controls |
| Mixed empty/singleton/multi class shape is refused, including dead quantifier | No program | mixed_literal, mixed_constructor, mixed_dead, mixed_union |
| Subtraction removes empty and permits mixed alternatives | No program | controls |
| Scoped i enable/disable tracked in source order | No program | modifier_enable, modifier_disable |
| Source alternatives, quantifiers, lookbehind traversal preserve stale parser flags | No program | modifier_alternative, modifier_quantified, modifier_lookbehind |
| Later u word/nonword escape, word class and negated class affected by stale flags | No program | modifier_word, modifier_nonword, modifier_word_class, modifier_negated_class |
| Later v property escape affected by stale flags | No program | modifier_property |
| Literal/range/digit operands outside group agree in u or unaffected v | No program | controls: (?i:a)b, (?i:a)[b]/u, (?i:a)[0-9]/v |
| Entering ordinary group resets stale flags, nested modifier reset | No program | controls: (?i:a)(?:[b]), nested enable/disable/reset |
| u word-class union masks leaked or dropped i | No program | controls: modified [wW] in u and iu |
| Multi-character v string comparison under stale parser i | No program | modifier_class_string enables i before uppercase q{AB}; controls: (?-i:^)[q{AB}]/iv masks dropped i |
| v property comparison agrees for case-invariant Any | No program | controls: (?i:a)p{Any}/v; modifier_ascii_property refuses ASCII because Unicode i admits Kelvin sign and long s |
| Dropped parser i before a bare u word escape still agrees after matcher folding | No program | controls: (?-i:a)w/iu |
| Negated u word-class union masks stale flags | No program | controls: (?i:^)[^wW]/u |
| Legacy word escape after scoped i does not enter Unicode refusal | No program | controls |
| Scoped s/m enable and disable | No program | controls |
| Refusal before NativeDeclarations and before either stage-0 backend | Folding refusal fixture, lower Go tests | 21 isolated programs, exact-message oracle test, CLI build and js for each |

## Conditions without an Adamic source witness

The Go PropertyProvider failure, reused storage/snapshot cases, and injected
instruction budget cannot be expressed through the typed built-in RegExp API.
The default provider is fixed, and programs cannot set regex_step_limit.
The existing Go/native package tests exercise those seams. Error propagation
from an invalid class does not reach the V8 visit: parsing/initial compilation
has already rejected it. Likewise the visitor's hypothetical unknown AST node
cannot be constructed by an Adamic program. Generated Unicode tables are data,
not new source-call cases; this change does not claim an exhaustive character
survey. This inventory records feature/source cases, not measured Go or C
branch coverage.

Scoped modifier literals are rejected by the current TypeScript target:
TS18062 requires es2025 or later. Constant new RegExp strings exercise the
same supported/refused shapes. The current library also rejects new RegExp()
with TS2554, Expected 1-2 arguments, but got 0. Both attempts are retained in
notes/regex as .a files. Empty-string construction is covered and compiles.

Callbacks for both replacement methods lower to NotYet: regex replacement
other than a string. Their Node output is [a]b[a]; no native output exists.
The known mixed-empty replacement hang is checked as a compile-time refusal,
so no native replacement execution can be written for that shape.
