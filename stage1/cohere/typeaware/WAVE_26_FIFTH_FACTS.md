# Preference-rule raw questions

Four isolated questions serve the fifth batch. Each has a separate Go entry
file and a separate `.a` decoder; facts.go has one registration line per question.
They use the unchanged tsgo_inspect ABI and owned UTF-8 frame protocol. A released
program is rejected before any question runs. All symbol identities are opaque,
program-scoped numbers, never addresses. Native decoders require version 1,
matching question names, canonical numbers and exact frame consumption.

## preference-structure

An exact SourceFile request returns module kind, node count, then a preorder raw
syntax projection. Each node supplies kind, byte Pos/End/token-start, parent,
flags, cooked literal/identifier text, immediate children, eight syntax roles,
arguments, parameters, operator kind, spread marker and IsDeclarationName.
The roles are expression, name, type/property name, body, left, right, true branch
and false branch. Tagged templates use expression/name for tag/template;
template spans use them for substitution/literal; template expressions use name
for the head. Binding elements and export specifiers carry property name in role
two. Export declarations carry module specifier in expression. Calls and new
expressions both supply callee and arguments. This is a new schema and does not
change the existing binding-structure question or decoder.

Parse diagnostics and extra question fields are refused. Native verifies every
node range and numeric child identity. Classification as a value reference,
write, pass-through, target or constructor stays in Adamic.

## preference-binding

An exact Identifier request returns two groups: the ordinary symbol and the
value-read symbol. Each group contains opaque symbol identity, declaration count
and the declaration-file flag for every declaration in compiler order. Value-read
uses GetShorthandAssignmentValueSymbol for shorthand names and
GetExportSpecifierLocalTargetSymbol for local exports, otherwise the ordinary
symbol. Nil symbols have identity zero and zero declarations.

The native rejection rule requires a nonzero ordinary symbol with at least one
declaration and every declaration ambient for global Promise. The regex raw tag
uses the first declaration flag. Implicit arguments requires a nonzero symbol
with zero declarations. Undefined requires no symbol or zero declarations.
Reference tracking uses read-symbol equality and actual source declarations.
None of these predicates is returned by Go.

## regex-pattern

An exact CallExpression or NewExpression request carries flags and pattern as
`regex-pattern`, LF, flags, LF, pattern. Native validates flags before requesting
this scan. The pattern may itself contain LF; the split is limited to three parts.
Payload supplies the raw character walk's completion flag, character count and
each character's UTF-16 start/end/written text, followed by escape and class
extent groups. Each extent is UTF-16 start/end plus scan-success flag. UTF-16
positions describe the query pattern, not the source file. Decoders require
character slices to equal the written text and reject invalid extents.

The grammar walker, escape/class parser and hexadecimal helpers are copied from
pinned cohere 715ba94f into the isolated regex_pattern subpackages, with only
package paths and comment punctuation adjusted. They are syntax machinery, not
rule implementations. The copied comment scanner below has the same provenance;
cohere's license is covered by THIRD_PARTY_NOTICES.md.

Go returns no flags-validity, printable-pattern, balance, rewrite, report or
suggestion verdict. Native validates flags, balances groups and quantifiers,
checks printable characters, rewrites selected raw character spellings, chooses
padding, constructs descriptions, and serializes proposed edits. The Go oracle
imports the unchanged production rule and its original helpers independently,
not these copies. The scan-completion mutant remains a valid frame and is caught
only by different proposed suggestion bytes.

## source-comments

An exact SourceFile request returns comment count and raw byte start/end pairs.
The isolated copied comments.All scanner uses parser node/list positions and the
compiler's leading/trailing comment scanners, including empty delimiter lists.
It does not infer comments from arbitrary slashes inside strings or regexes.
Native checks range bounds and decides whether a comment lies inside a proposed
whole-call replacement. Removing comment ranges produces valid frames and a
normal exit, but the independent oracle catches the added suggestion.

The direct contract test holds binding groups to compiler operations, syntax
population to the compiler walk, raw pattern characters to hardcoded extents,
comment positions to source offsets, and malformed/wrong-kind requests to
refusal. Full controls independently check findings and suggestions. All four
question mutants compile, exit normally and are caught only by those bytes.
ASan/UBSan/LeakSanitizer cover C output ownership and native decoding; they do
not instrument the Go heap.
