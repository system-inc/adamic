# No this alias candidate

Public name @typescript-eslint/no-this-alias. Independent oracle:
cohere/internal/lint/rules/typescript.NoThisAlias, pinned cohere commit
715ba94f3608a6500086b1076ce5cb7e51b836db.

Only .ts, .tsx, .mts and .cts filenames enable the rule, exactly as Go. .js and
.a sources are excluded by Go's extension predicate and by this candidate.
Variable declarations and all assignment operators require a bare ThisKeyword
on the right; parentheses and type wrappers there are not skipped. Identifier
assignment targets unwrap parentheses, as, satisfies, angle-bracket assertions
and non-null assertions; member targets are silent. Direct object/array targets
report under the destructuring option. Parenthesized object/array targets go
through Go's non-identifier path and remain silent. Findings point at the bare
identifier or the whole direct pattern. Exact messages, ranges, no fixes and
no suggestions match Go.

The local JSON decoder preserves default allowed destructuring, the inverted
ReportDestructuring captured Go field, allowDestructuring, both allow-list
spellings, correct-name precedence and duplicate/null handling. OptionsJson
provides JSON and string decoding. Go captures marshal the already-decoded Go
struct; the oracle adapter distinguishes that shape from public wire options.
Malformed configuration is explicitly rejected; cross-runtime configuration
error text/exit parity is not claimed.

Original cohere fixture capture preserves File and Options, then reconstructs
the original filename suffix per fixture. This avoids the shared capture's loss
of JavaScript exclusions. All forty distinct source/file/options cases compare
with Go on source Node, emitted JavaScript and ASan/UBSan native. A six-extension
option matrix also tests public and captured options, duplicate/null behavior,
identifier wrappers and direct versus wrapped destructuring. Raw TypeScript
witnesses live in testdata/*.ts.txt; these modules use .ts temporarily under Ahra's October 7 correction; integration owns the codemod to .a.

The semantic mutant inverts allowed-name membership. It compiles and exits
successfully with empty stderr before comparison detects missing findings on
all three Adamic execution paths. No compiler failure or sanitizer report is
credited as catching it.

Default integration is still blocked by the shared .a registration rejection
and profile API mismatch. The proposed compatibility.patch in the boolean
outcome directory remains unapplied. validate.py reconstructs it into scratch
and maps the owned validation_test.go.txt into a virtual Go test file. Source
/workspace/adamic-tools/env.sh, supply --scratch and --typescript (pinned
TypeScript v6.0.3), and run the bounded four-way corpus, original fixture,
options, throughput, mutant, registry, vet and filtered-oracle checks. This is
a candidate validated through that overlay, not an unconditional merge claim.

The other two rules in this batch remain blocked by suggestion subranges. Exact
commands, measurements and limits are in the
[batch report](../../claims/wave1-12-batch4-evidence/REPORT.md).
