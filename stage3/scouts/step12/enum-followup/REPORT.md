Pinned the scanner stop to upstream `return token = keyword;` at scanner.ts:1823.
Branch: codex/step12-scout; follow-up to ebaf1bc0, with no compiler edits.
Four witnesses, four Node source mutants, three compiler revisions tested.
Exact expression is NotYet on all three; split statements are native-equal.
Historical full-closure refusal is not reproduced; language questions remain open.

# Enum slot follow-up

## Exact source and counts

Pinned TypeScript 6.0.3 source commit
050880ce59e30b356b686bd3144efe24f875ebc8 contains:

```typescript
// src/compiler/scanner.ts:1815
function getIdentifierToken(): SyntaxKind.Identifier | KeywordSyntaxKind {
    const len = tokenValue.length;
    if (len >= 2 && len <= 12) {
        const ch = tokenValue.charCodeAt(0);
        if (ch >= CharacterCodes.a && ch <= CharacterCodes.z) {
            const keyword = textToKeyword.get(tokenValue); // 1821
            if (keyword !== undefined) {                  // 1822
                return token = keyword;                   // 1823
            }
        }
    }
    return token = SyntaxKind.Identifier;                 // 1827
}
```

The gathered closure maps the first return to scanner.ts:1269:21; column 28
is its `token` left operand. This is the keyword branch, not the fallback
Identifier constant. The fallback maps to 1273:9. There is **one** such keyword
return-assignment and **one** fallback in this function. Source audit records
UTF-16 offsets, original file SHA-256, both exact return texts and coordinates.
It also verifies the complete enum and keyword alias copied into the larger
control (only export modifiers and CRLF are removed).

`token: SyntaxKind` is declared at scanner.ts:1048:9. Its destination is the
whole enum. The function result is the closed union `Identifier | KeywordSyntaxKind`.
`KeywordSyntaxKind` at types.ts:582 has 84 members. `textToKeywordObj` at
scanner.ts:135 has 84 literal entries, each an explicit keyword member; its
annotation is `MapLike<KeywordSyntaxKind>`. At :222 the Map is constructed with
`Object.entries(textToKeywordObj)`. No switch or keyword-number comparison
establishes membership at :1823. `keyword !== undefined` proves presence of
an already typed lookup result. Length and lowercase checks are text filters,
not membership proofs.

This chain is contingent on sound construction and invariant writes to the Map.
The unresolved Object.entries shape ruling still applies upstream. The reduced
witness replaces that separate stop with an explicitly typed two-entry Map literal;
it does not validate Object.entries or the full scanner closure.

## Node and specification

The reduced real-code driver scans `abstract`, `break`, `name`, `Abstract`, `x`.
Node (v24.19.0, stock TypeScript 6.0.3 transpilation) prints both return and
captured token: `128 128`, `83 83`, then three `80 80` lines. Enum values agree
with stock TypeScript's runtime constants. Misses return undefined from Map.get;
the guard takes the fallback. The const-enum syntax itself adds no JS check.

ECMAScript [simple assignment evaluation](https://tc39.es/ecma262/#sec-assignment-operators-runtime-semantics-evaluation)
evaluates the left reference, evaluates/GetValues the right expression, performs
PutValue, then returns **that right value**. Return evaluates/GetValues its
expression. Therefore this assignment result is `keyword`, not a fresh read of
an arbitrarily valued `token`. The implementation must evaluate RHS once, store
once, and return that same value; it must preserve failure and effect order.
[Map.prototype.get](https://tc39.es/ecma262/#sec-map.prototype.get) returns the
stored value or undefined. JS provides no numeric enum-membership guarantee.

At pinned typescript-go source cohere 7945d102a6c18dd36adf9114a758ce646e8b2359,
`internal/scanner/scanner.go:44` builds a typed `map[string]ast.Kind` literal.
`GetIdentifierToken` at :2214-2222 looks up the map, compares to `ast.KindUnknown`
(the missing-map zero sentinel), and returns the keyword or `ast.KindIdentifier`.
It has a concrete integer return type, not the TypeScript member-union promise;
its return does not assign scanner state. This is source inspection, not a Go
execution or evidence of native scanner equality.

## Origin branch inspection and measured builds

`git ls-remote --heads origin '*enum-tag-narrowing*'` and explicit fetch found:

| Remote branch | Tested tip |
| --- | --- |
| codex/enum-tag-narrowing | 41231d514cf0090d1a7ec103326858b7eb34b6e3 |
| codex/enum-tag-narrowing-2 | 649522529a22d44023a2a3c13ed82f2d03cbbfab |
| claude/enum-tag-narrowing-nil-deref-g2aflr | same 64952252 |

Both distinct revisions were built in detached scratch worktrees, using their
pinned cohere (the same 7945d102). Main control is the scout base 45487a80.
No delivery-branch merge or compiler edits were made.

| Witness | Main 45487a80 | 41231d51 | 64952252 |
| --- | --- | --- | --- |
| identifier-return.a, two-member reduction | NotYet BinaryExpression | same | same |
| identifier-full-enum.a, complete upstream const enum and 84-member alias | NotYet BinaryExpression | same | same |
| identifier-split.a, assign then return RHS | native equals Node | native equals Node | native equals Node |
| unproven-member.a, Math.trunc(99) into Identifier | adamic/enum-literal refusal | same | same |
| Derived proven-member control, explicit Identifier constant | native 80 | native 80 | native 80 |

The historical full-closure diagnostic naming Identifier does not reappear in
the reductions. Neither feature tip lowers the exact return-assignment. A green
split witness demonstrates the lookup's type/presence proof is admitted in that
reduction; it does not prove the reported full-closure refusal has been repaired.

Source explanation: `internal/lower/expression.go` handles BinaryExpression by
lowering operands and calling combine; it has no simple-assignment expression
case. Assignment **statements** have a separate path in assignments.go.
`enum_tag_views.go` handles object/tag storage promises, not this expression's
value semantics. The newer tip's 64952252 fix checks every member of a split
object union; e17ef990 adds nested overload implementation binding. Neither
adds assignment-expression lowering. `enum_never.go` retains the distinction
between whole numeric enum storage and member-origin proofs. The negative
witness confirms the member-slot check is active and not simply bypassed.

## Smallest candidate rule, not a ruling

For simple `slot = rhs` used as a value, retain the proof of **rhs** for the
expression result. Prove the write admissible into slot separately. Materialize
a temporary if needed so RHS is evaluated once and the value returned after the
write is exactly the stored RHS, without widening its proof to the slot's whole
numeric enum type. With this typed Map and definedness guard, rhs's finite
keyword-member union is a subset of the declared return union; fallback is an
explicit Identifier constant. No new keyword-number test is needed once that
Map provenance is established.

An adaptation could use `token = keyword; return keyword;` and
`token = SyntaxKind.Identifier; return SyntaxKind.Identifier;` as measured here.
That preserves this local variable's behavior but is not applied to upstream.
It also avoids claiming a flow-narrowed reread of mutable captured state.
General property assignment requires preserving reference evaluation, setters,
exceptions and effects. Compound assignment and aliased mutable member slots
need separate proofs; this scout does not admit them.

## Tests, mutants and limits

`run.cjs` ran four stock-Node goldens and four actual source mutants on Node:

| Witness | Mutation | Catcher |
| --- | --- | --- |
| identifier-return | first driver `abstract` becomes `break` | stdout first row 128/128 becomes 83/83 |
| identifier-full-enum | same driver mutation | same Node stdout difference |
| identifier-split | keyword branch stores Identifier instead | stdout 128/128 becomes 128/80; each of three native runs also differs from golden and equals mutated Node |
| unproven-member | Math.trunc(99) becomes Math.trunc(100) | Node stdout 99 becomes 100 |

A fifth, audit-only mutant changes the copied full enum Identifier declaration
to value 79. The source audit fails with `full type copy mismatch` (exit 1);
the unmutated audit passes. This is distinct from the four Node source mutants.

For the refusal boundary, substituting the proven enum constant yields a green
native control on all three revisions, printing 80. The invalid number cannot
silently enter the member slot; the explicit constant is admitted.
15 control build expectations plus three mutant builds, six native controls,
three native mutants passed. Raw
stdout/stderr, result matrix and source audit are retained in evidence. Failed
exploratory reductions are documented: ordinary/full const-enum exact returns
both hit NotYet; an initial whole-enum-to-member variable attempt hit TS2322
before lowering. The final negative fixture isolates the actual enum-literal
check with a number-valued RHS.

Commands: `go build -buildvcs=false -o SCRATCH_BINARY ./cmd/adamic` in each of
two feature worktrees; `SCOUT_TYPESCRIPT=API/lib/typescript.js node
stage3/scouts/step12/enum-followup/audit.cjs TREE RAW_CLOSURE`; and
`SCOUT_TYPESCRIPT=API/lib/typescript.js node stage3/scouts/step12/enum-followup/run.cjs
NEW_OUTPUT COMPILER_PATHS.json`, each redirected to a log file. No whole package
tests, full gate or scanner lane ran. This is a reduced-expression measurement.

Required warm setup succeeded with GOPROXY=https://proxy.golang.org|direct.
Cumulative timing lines: Node 0.019s, Go 0.021s, submodules 0.057s, markdown
0.067s (skipped step 0.006s), clang 0.143s, Go build 31.766s, deferred tests
31.869s, warm cache 31.870s, done 31.895s. nproc=5, cpu.max=400000/100000.

## Questions for @system_adamic, undecided

1. Should a simple assignment expression carry its RHS member-union proof into
   the result while the write independently checks the destination slot, or
   should scanner adaptation split these statements first?
2. Which proof/check ruling for the upstream Object.entries initialization will
   establish that all runtime Map values inhabit KeywordSyntaxKind, including
   aliases and possible extra enumerable entries? The literal reduction does
   not settle that ruling.
3. Can the original adapted snapshot/compiler SHA that produced the enum-slot
   diagnostic at 1269:28 be supplied to distinguish a union/contextual-type
   refusal from a previously fixed guard? These reductions do not reproduce it.
